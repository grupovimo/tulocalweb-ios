// Package apns manda notificaciones push al iPhone por APNs (HTTP/2) con autenticación por token:
// un JWT ES256 firmado con la clave .p8 de la cuenta de Apple, que se reusa mientras vale (Apple
// pide renovarlo entre 20 y 60 minutos; acá, cada 40). Sólo biblioteca estándar.
//
// Sin configuración (Key ID, Team ID, clave o topic), Desde devuelve nil y Enviar sobre un cliente
// nil no hace nada: el envío queda apagado sin romper nada (docs/decisiones.md §40).
package apns

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Hosts de APNs por entorno del token (el de la app de Xcode es sandbox; TestFlight y App Store,
// producción).
const (
	HostProduccion = "https://api.push.apple.com"
	HostSandbox    = "https://api.sandbox.push.apple.com"
)

// VidaJWT: cuánto se reusa el token del proveedor antes de firmar otro.
const VidaJWT = 40 * time.Minute

// ErrTokenInvalido: APNs dice que ese token de dispositivo ya no sirve (410, BadDeviceToken o
// Unregistered). Hay que borrarlo.
var ErrTokenInvalido = errors.New("apns: token de dispositivo inválido")

// Error es otra respuesta no exitosa de APNs (nunca lleva el token ni el JWT).
type Error struct {
	Estado int
	Motivo string
}

func (e *Error) Error() string { return fmt.Sprintf("apns: HTTP %d %s", e.Estado, e.Motivo) }

// Reintentable dice si vale la pena volver a probar más tarde (APNs caído, demasiados pedidos).
func (e *Error) Reintentable() bool {
	return e.Estado == http.StatusTooManyRequests || e.Estado >= 500 || e.Motivo == "ExpiredProviderToken"
}

// Cliente habla con APNs. Usalo por puntero; es seguro entre goroutines.
type Cliente struct {
	KeyID, TeamID, Topic string
	Clave                *ecdsa.PrivateKey
	// HTTP: en los tests, el cliente del servidor falso. Vacío: uno con HTTP/2 y 15 s de espera.
	HTTP *http.Client
	// Hosts por entorno ("sandbox", "produccion"). Vacío: los de Apple.
	Hosts map[string]string
	// Ahora: el reloj (tests). Vacío: time.Now.
	Ahora func() time.Time

	mu      sync.Mutex
	jwt     string
	firmado time.Time
}

// Desde arma el cliente con la configuración. Si falta algo, devuelve nil y ningún error: el envío
// queda apagado. Un .p8 que no se puede leer sí es error (está configurado, pero mal).
func Desde(keyID, teamID, rutaClave, topic string) (*Cliente, error) {
	if keyID == "" || teamID == "" || rutaClave == "" || topic == "" {
		return nil, nil
	}
	pemBytes, err := os.ReadFile(rutaClave)
	if err != nil {
		return nil, fmt.Errorf("apns: no se pudo leer la clave: %w", err)
	}
	clave, err := LeerClave(pemBytes)
	if err != nil {
		return nil, err
	}
	return &Cliente{KeyID: keyID, TeamID: teamID, Topic: topic, Clave: clave}, nil
}

// LeerClave interpreta el .p8 de Apple (PEM «PRIVATE KEY», PKCS#8, P-256).
func LeerClave(pemBytes []byte) (*ecdsa.PrivateKey, error) {
	b, _ := pem.Decode(pemBytes)
	if b == nil {
		return nil, errors.New("apns: la clave no es PEM")
	}
	k, err := x509.ParsePKCS8PrivateKey(b.Bytes)
	if err != nil {
		return nil, fmt.Errorf("apns: la clave no es PKCS#8: %w", err)
	}
	ec, ok := k.(*ecdsa.PrivateKey)
	if !ok || ec.Curve.Params().BitSize != 256 {
		return nil, errors.New("apns: la clave no es ECDSA P-256")
	}
	return ec, nil
}

// Activo dice si hay cliente (nil: apagado).
func (c *Cliente) Activo() bool { return c != nil }

func (c *Cliente) ahora() time.Time {
	if c.Ahora != nil {
		return c.Ahora()
	}
	return time.Now()
}

// Token devuelve el JWT del proveedor, firmando uno nuevo si el guardado tiene más de VidaJWT.
func (c *Cliente) Token() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	ahora := c.ahora()
	if c.jwt != "" && ahora.Sub(c.firmado) < VidaJWT {
		return c.jwt, nil
	}
	t, err := FirmarJWT(c.Clave, c.KeyID, c.TeamID, ahora)
	if err != nil {
		return "", err
	}
	c.jwt, c.firmado = t, ahora
	return t, nil
}

// olvidarToken descarta el JWT (APNs lo dio por vencido) para que el próximo envío firme otro.
func (c *Cliente) olvidarToken() {
	c.mu.Lock()
	c.jwt = ""
	c.mu.Unlock()
}

// FirmarJWT arma el token del proveedor: {"alg":"ES256","kid"} . {"iss":team,"iat"} con la firma
// r‖s de 64 bytes (JWS, no DER).
func FirmarJWT(clave *ecdsa.PrivateKey, keyID, teamID string, ahora time.Time) (string, error) {
	if clave == nil {
		return "", errors.New("apns: sin clave")
	}
	enc := base64.RawURLEncoding
	cab, _ := json.Marshal(map[string]string{"alg": "ES256", "kid": keyID})
	cue, _ := json.Marshal(map[string]any{"iss": teamID, "iat": ahora.Unix()})
	firmable := enc.EncodeToString(cab) + "." + enc.EncodeToString(cue)
	h := sha256.Sum256([]byte(firmable))
	r, s, err := ecdsa.Sign(rand.Reader, clave, h[:])
	if err != nil {
		return "", err
	}
	firma := make([]byte, 64)
	r.FillBytes(firma[:32])
	s.FillBytes(firma[32:])
	return firmable + "." + enc.EncodeToString(firma), nil
}

// Envio es una notificación para un dispositivo.
type Envio struct {
	Token   string // del dispositivo, en hexadecimal
	Entorno string // sandbox | produccion
	Payload []byte
	// Colapso (apns-collapse-id): dos envíos con el mismo se ven como uno (los reintentos del outbox).
	Colapso string
	// Vence: hasta cuándo APNs guarda el aviso si el iPhone está apagado. Cero: una hora.
	Vence time.Time
	// Actividad: es una actualización de Live Activity (apns-push-type liveactivity, al topic
	// <bundle>.push-type.liveactivity) y no un aviso.
	Actividad bool
}

func (c *Cliente) cli() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return clientePorDefecto
}

var clientePorDefecto = &http.Client{
	Timeout:   15 * time.Second,
	Transport: &http.Transport{ForceAttemptHTTP2: true, MaxIdleConnsPerHost: 4, IdleConnTimeout: 5 * time.Minute},
}

func (c *Cliente) host(entorno string) string {
	if entorno != "sandbox" {
		entorno = "produccion"
	}
	if h, ok := c.Hosts[entorno]; ok {
		return h
	}
	if entorno == "sandbox" {
		return HostSandbox
	}
	return HostProduccion
}

// Enviar manda la notificación. Sobre un cliente nil no hace nada. Devuelve ErrTokenInvalido si hay
// que borrar el token, o *Error con lo que respondió APNs. Si APNs da por vencido el JWT, firma otro
// y prueba una vez más.
func (c *Cliente) Enviar(ctx context.Context, e Envio) error {
	if c == nil {
		return nil
	}
	err := c.enviar(ctx, e)
	var ea *Error
	if errors.As(err, &ea) && ea.Motivo == "ExpiredProviderToken" {
		c.olvidarToken()
		err = c.enviar(ctx, e)
	}
	return err
}

func (c *Cliente) enviar(ctx context.Context, e Envio) error {
	jwt, err := c.Token()
	if err != nil {
		return err
	}
	url := strings.TrimRight(c.host(e.Entorno), "/") + "/3/device/" + e.Token
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(e.Payload))
	if err != nil {
		return err
	}
	vence := e.Vence
	if vence.IsZero() {
		vence = c.ahora().Add(time.Hour)
	}
	req.Header.Set("authorization", "bearer "+jwt)
	if e.Actividad {
		req.Header.Set("apns-topic", c.Topic+".push-type.liveactivity")
		req.Header.Set("apns-push-type", "liveactivity")
	} else {
		req.Header.Set("apns-topic", c.Topic)
		req.Header.Set("apns-push-type", "alert")
	}
	req.Header.Set("apns-priority", "10")
	req.Header.Set("apns-expiration", strconv.FormatInt(vence.Unix(), 10))
	if e.Colapso != "" {
		req.Header.Set("apns-collapse-id", e.Colapso)
	}
	req.Header.Set("content-type", "application/json")
	res, err := c.cli().Do(req)
	if err != nil {
		// El error de red trae la URL con el token del dispositivo: no se propaga.
		return errors.New("apns: no se pudo conectar")
	}
	defer res.Body.Close()
	cuerpo, _ := io.ReadAll(io.LimitReader(res.Body, 4<<10))
	if res.StatusCode == http.StatusOK {
		return nil
	}
	var r struct {
		Reason string `json:"reason"`
	}
	_ = json.Unmarshal(cuerpo, &r)
	if res.StatusCode == http.StatusGone || r.Reason == "BadDeviceToken" || r.Reason == "Unregistered" {
		return ErrTokenInvalido
	}
	return &Error{Estado: res.StatusCode, Motivo: r.Reason}
}
