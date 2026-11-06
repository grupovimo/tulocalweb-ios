// Package fcm manda notificaciones push a Android por Firebase Cloud Messaging (API HTTP v1) con la
// cuenta de servicio del producto: firma un JWT RS256 con su clave, lo canjea en Google por un token
// de acceso y lo reusa mientras vale (dura una hora; acá se renueva 5 minutos antes). Sólo
// biblioteca estándar, como apns: está pensado para llevarlo tal cual al resto de las apps de Vimo.
//
// Sin configuración (la ruta del JSON de la cuenta), Desde devuelve nil y Enviar sobre un cliente
// nil no hace nada: el envío queda apagado sin romper nada.
package fcm

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Alcance que pide FCM para mandar.
const Alcance = "https://www.googleapis.com/auth/firebase.messaging"

// Host de la API de FCM.
const Host = "https://fcm.googleapis.com"

// Margen: cuánto antes de que venza el token de acceso se pide otro.
const Margen = 5 * time.Minute

// ErrTokenInvalido: FCM dice que ese token de dispositivo ya no sirve (UNREGISTERED, o un token mal
// formado). Hay que borrarlo.
var ErrTokenInvalido = errors.New("fcm: token de dispositivo inválido")

// Error es otra respuesta no exitosa de FCM o de Google (nunca lleva el token ni la clave).
type Error struct {
	Estado int
	Motivo string
}

func (e *Error) Error() string { return fmt.Sprintf("fcm: HTTP %d %s", e.Estado, e.Motivo) }

// Reintentable dice si vale la pena volver a probar más tarde (FCM caído, cuota agotada).
func (e *Error) Reintentable() bool {
	return e.Estado == http.StatusTooManyRequests || e.Estado >= 500
}

// Cuenta es lo que hace falta del JSON de la cuenta de servicio de Google.
type Cuenta struct {
	ProjectID   string `json:"project_id"`
	ClientEmail string `json:"client_email"`
	PrivateKey  string `json:"private_key"`
	TokenURI    string `json:"token_uri"`
}

// Cliente habla con FCM. Usalo por puntero; es seguro entre goroutines.
type Cliente struct {
	Proyecto, Email, URLToken string
	Clave                     *rsa.PrivateKey
	// HTTP: en los tests, el cliente del servidor falso. Vacío: uno con 15 s de espera.
	HTTP *http.Client
	// Host de FCM. Vacío: el de Google.
	Host string
	// Ahora: el reloj (tests). Vacío: time.Now.
	Ahora func() time.Time

	mu     sync.Mutex
	acceso string
	vence  time.Time
}

// Desde arma el cliente con el JSON de la cuenta de servicio. Sin ruta, devuelve nil y ningún error:
// el envío queda apagado. Un JSON que no se puede leer sí es error (está configurado, pero mal).
func Desde(ruta string) (*Cliente, error) {
	if ruta == "" {
		return nil, nil
	}
	b, err := os.ReadFile(ruta)
	if err != nil {
		return nil, fmt.Errorf("fcm: no se pudo leer la cuenta de servicio: %w", err)
	}
	return DeCuenta(b)
}

// DeCuenta arma el cliente con el contenido del JSON de la cuenta de servicio.
func DeCuenta(b []byte) (*Cliente, error) {
	var c Cuenta
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, errors.New("fcm: la cuenta de servicio no es JSON")
	}
	if c.ProjectID == "" || c.ClientEmail == "" || c.PrivateKey == "" {
		return nil, errors.New("fcm: a la cuenta de servicio le falta project_id, client_email o private_key")
	}
	clave, err := LeerClave([]byte(c.PrivateKey))
	if err != nil {
		return nil, err
	}
	if c.TokenURI == "" {
		c.TokenURI = "https://oauth2.googleapis.com/token"
	}
	return &Cliente{Proyecto: c.ProjectID, Email: c.ClientEmail, URLToken: c.TokenURI, Clave: clave}, nil
}

// LeerClave interpreta la clave de la cuenta de servicio (PEM «PRIVATE KEY», PKCS#8, RSA).
func LeerClave(pemBytes []byte) (*rsa.PrivateKey, error) {
	b, _ := pem.Decode(pemBytes)
	if b == nil {
		return nil, errors.New("fcm: la clave no es PEM")
	}
	k, err := x509.ParsePKCS8PrivateKey(b.Bytes)
	if err != nil {
		return nil, fmt.Errorf("fcm: la clave no es PKCS#8: %w", err)
	}
	r, ok := k.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("fcm: la clave no es RSA")
	}
	return r, nil
}

// Activo dice si hay cliente (nil: apagado).
func (c *Cliente) Activo() bool { return c != nil }

func (c *Cliente) ahora() time.Time {
	if c.Ahora != nil {
		return c.Ahora()
	}
	return time.Now()
}

func (c *Cliente) cli() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return clientePorDefecto
}

var clientePorDefecto = &http.Client{Timeout: 15 * time.Second}

// FirmarJWT arma la aserción que se canjea por el token de acceso: {"alg":"RS256"} .
// {"iss","scope","aud","iat","exp"} firmado con la clave de la cuenta.
func FirmarJWT(clave *rsa.PrivateKey, email, aud string, ahora time.Time) (string, error) {
	if clave == nil {
		return "", errors.New("fcm: sin clave")
	}
	enc := base64.RawURLEncoding
	cab, _ := json.Marshal(map[string]string{"alg": "RS256", "typ": "JWT"})
	cue, _ := json.Marshal(map[string]any{"iss": email, "scope": Alcance, "aud": aud, "iat": ahora.Unix(), "exp": ahora.Add(time.Hour).Unix()})
	firmable := enc.EncodeToString(cab) + "." + enc.EncodeToString(cue)
	h := sha256.Sum256([]byte(firmable))
	firma, err := rsa.SignPKCS1v15(rand.Reader, clave, crypto.SHA256, h[:])
	if err != nil {
		return "", err
	}
	return firmable + "." + enc.EncodeToString(firma), nil
}

// Token devuelve el token de acceso, pidiendo uno nuevo a Google si no hay o está por vencer.
func (c *Cliente) Token(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	ahora := c.ahora()
	if c.acceso != "" && ahora.Before(c.vence.Add(-Margen)) {
		return c.acceso, nil
	}
	jwt, err := FirmarJWT(c.Clave, c.Email, c.URLToken, ahora)
	if err != nil {
		return "", err
	}
	form := url.Values{"grant_type": {"urn:ietf:params:oauth:grant-type:jwt-bearer"}, "assertion": {jwt}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URLToken, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	res, err := c.cli().Do(req)
	if err != nil {
		return "", errors.New("fcm: no se pudo pedir el token de acceso")
	}
	defer res.Body.Close()
	cuerpo, _ := io.ReadAll(io.LimitReader(res.Body, 8<<10))
	var r struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
		Error       string `json:"error"`
	}
	_ = json.Unmarshal(cuerpo, &r)
	if res.StatusCode != http.StatusOK || r.AccessToken == "" {
		return "", &Error{Estado: res.StatusCode, Motivo: "token: " + r.Error}
	}
	if r.ExpiresIn <= 0 {
		r.ExpiresIn = 3600
	}
	c.acceso, c.vence = r.AccessToken, ahora.Add(time.Duration(r.ExpiresIn)*time.Second)
	return c.acceso, nil
}

// olvidarToken descarta el token de acceso (FCM respondió 401) para que el próximo envío pida otro.
func (c *Cliente) olvidarToken() {
	c.mu.Lock()
	c.acceso = ""
	c.mu.Unlock()
}

// Envio es una notificación para un dispositivo.
type Envio struct {
	Token          string // de registro de FCM del dispositivo
	Titulo, Cuerpo string
	// Datos: lo que lee la app al tocar el aviso (la ruta, el negocio). Sólo texto.
	Datos map[string]string
	// Colapso (collapse_key y tag): dos envíos con el mismo se ven como uno.
	Colapso string
	// Vida: cuánto guarda FCM el aviso si el teléfono está apagado. Cero: una hora.
	Vida time.Duration
	// SoloDatos: sin bloque notification; el título y el cuerpo van en data (titulo, cuerpo) y la app
	// arma el aviso ella misma (con sus acciones). Así llega a onMessageReceived también en segundo plano.
	SoloDatos bool
}

// Mensaje arma el cuerpo de messages:send.
func Mensaje(e Envio) []byte {
	vida := e.Vida
	if vida <= 0 {
		vida = time.Hour
	}
	type notificacion struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	type notiAndroid struct {
		Tag   string `json:"tag,omitempty"`
		Sound string `json:"sound"`
	}
	type android struct {
		CollapseKey  string       `json:"collapse_key,omitempty"`
		Priority     string       `json:"priority"`
		TTL          string       `json:"ttl"`
		Notification *notiAndroid `json:"notification,omitempty"`
	}
	type mensaje struct {
		Token        string            `json:"token"`
		Notification *notificacion     `json:"notification,omitempty"`
		Data         map[string]string `json:"data,omitempty"`
		Android      android           `json:"android"`
	}
	m := mensaje{Token: e.Token, Data: e.Datos,
		Android: android{CollapseKey: e.Colapso, Priority: "HIGH", TTL: strconv.FormatInt(int64(vida/time.Second), 10) + "s"}}
	if e.SoloDatos {
		m.Data = make(map[string]string, len(e.Datos)+2)
		for k, v := range e.Datos {
			m.Data[k] = v
		}
		m.Data["titulo"], m.Data["cuerpo"] = e.Titulo, e.Cuerpo
	} else {
		m.Notification = &notificacion{e.Titulo, e.Cuerpo}
		m.Android.Notification = &notiAndroid{Tag: e.Colapso, Sound: "default"}
	}
	b, _ := json.Marshal(map[string]mensaje{"message": m})
	return b
}

// Enviar manda la notificación. Sobre un cliente nil no hace nada. Devuelve ErrTokenInvalido si hay
// que borrar el token, o *Error con lo que respondió FCM. Si FCM rechaza el token de acceso, pide
// otro y prueba una vez más.
func (c *Cliente) Enviar(ctx context.Context, e Envio) error {
	if c == nil {
		return nil
	}
	err := c.enviar(ctx, e)
	var ef *Error
	if errors.As(err, &ef) && ef.Estado == http.StatusUnauthorized {
		c.olvidarToken()
		err = c.enviar(ctx, e)
	}
	return err
}

func (c *Cliente) enviar(ctx context.Context, e Envio) error {
	acceso, err := c.Token(ctx)
	if err != nil {
		return err
	}
	host := c.Host
	if host == "" {
		host = Host
	}
	u := strings.TrimRight(host, "/") + "/v1/projects/" + url.PathEscape(c.Proyecto) + "/messages:send"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(Mensaje(e)))
	if err != nil {
		return err
	}
	req.Header.Set("authorization", "Bearer "+acceso)
	req.Header.Set("content-type", "application/json")
	res, err := c.cli().Do(req)
	if err != nil {
		return errors.New("fcm: no se pudo conectar")
	}
	defer res.Body.Close()
	cuerpo, _ := io.ReadAll(io.LimitReader(res.Body, 8<<10))
	if res.StatusCode == http.StatusOK {
		return nil
	}
	// {"error":{"code":404,"status":"NOT_FOUND","details":[{"@type":"…FcmError","errorCode":"UNREGISTERED"}]}}
	var r struct {
		Error struct {
			Status  string `json:"status"`
			Details []struct {
				ErrorCode string `json:"errorCode"`
			} `json:"details"`
		} `json:"error"`
	}
	_ = json.Unmarshal(cuerpo, &r)
	codigo := r.Error.Status
	for _, d := range r.Error.Details {
		if d.ErrorCode != "" {
			codigo = d.ErrorCode
		}
	}
	// UNREGISTERED: la app se desinstaló o el token rotó. INVALID_ARGUMENT con 400 es, en la práctica,
	// un token mal formado (el resto del mensaje lo arma este paquete).
	if codigo == "UNREGISTERED" || res.StatusCode == http.StatusNotFound ||
		(res.StatusCode == http.StatusBadRequest && codigo == "INVALID_ARGUMENT") {
		return ErrTokenInvalido
	}
	return &Error{Estado: res.StatusCode, Motivo: codigo}
}
