// Package clerk verifica los tokens de sesión de Clerk (la identidad compartida de Grupo Vimo) y
// lee a la persona con la Backend API.
//
// El token es un JWT RS256 que firma la instancia de Clerk. Se verifica sin ir a la red en cada
// pedido: las claves públicas (JWKS) se bajan una vez con la clave secreta y quedan en memoria; sólo
// se vuelven a pedir si llega un `kid` que no se conoce (como mucho una vez por minuto). Lo que
// importa de la persona (email verificado, nombre, foto) no viaja en el token: se lee de
// GET /v1/users/{id}, así nada de lo que diga el navegador o la app se usa sin pasar por Clerk.
package clerk

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"
)

// ErrToken: el token no es de esta instancia, venció, está mal firmado o viene de un origen que no
// está permitido.
var ErrToken = errors.New("clerk: token inválido")

// Margen para la diferencia de relojes entre Clerk y el VPS.
const margen = 5 * time.Second

// Cliente de una instancia de Clerk.
type Cliente struct {
	// Clave secreta de la instancia (sk_live_… / sk_test_…). Sólo va a la Backend API.
	Secreto string
	// Emisor de los tokens (`iss`): la Frontend API de la instancia, p. ej. https://clerk.grupovimo.com.ar.
	Emisor string
	// Clave publicable (pk_live_… / pk_test_…): es pública. La app de iPhone la pide a
	// /v1/auth/proveedores para hablar con la Frontend API de la misma instancia que verifica Go.
	Publicable string
	// Orígenes que pueden haber pedido el token (`azp`). Un token sin `azp` (la app nativa de
	// iPhone, que no tiene origen) se acepta; uno con `azp` tiene que estar en la lista.
	Origenes []string
	// Backend API (tests: el servidor falso). Vacía: https://api.clerk.com.
	API  string
	HTTP *http.Client
	// Ahora (tests). nil: time.Now.
	Ahora func() time.Time

	mu       sync.Mutex
	claves   map[string]*rsa.PublicKey
	buscadas time.Time
}

// EmisorDe saca la Frontend API de la clave publicable: pk_<entorno>_<base64("host$")>.
func EmisorDe(publicable string) (string, error) {
	partes := strings.SplitN(publicable, "_", 3)
	if len(partes) != 3 || partes[0] != "pk" || (partes[1] != "live" && partes[1] != "test") {
		return "", errors.New("clerk: la clave publicable no tiene la forma pk_live_… o pk_test_…")
	}
	b, err := base64.RawStdEncoding.DecodeString(strings.TrimRight(partes[2], "="))
	host := strings.TrimSuffix(string(b), "$")
	if err != nil || host == "" || strings.ContainsAny(host, "/ ") {
		return "", errors.New("clerk: la clave publicable no trae un dominio válido")
	}
	return "https://" + host, nil
}

// Nuevo arma el cliente con la clave secreta y la publicable (de la que sale el emisor).
func Nuevo(secreto, publicable string, origenes []string) (*Cliente, error) {
	if !strings.HasPrefix(secreto, "sk_") {
		return nil, errors.New("clerk: la clave secreta no tiene la forma sk_live_… o sk_test_…")
	}
	emisor, err := EmisorDe(publicable)
	if err != nil {
		return nil, err
	}
	return &Cliente{Secreto: secreto, Emisor: emisor, Publicable: publicable, Origenes: origenes}, nil
}

func (c *Cliente) ahora() time.Time {
	if c.Ahora != nil {
		return c.Ahora()
	}
	return time.Now()
}

func (c *Cliente) api() string {
	if c.API != "" {
		return strings.TrimRight(c.API, "/")
	}
	return "https://api.clerk.com"
}

func (c *Cliente) cliente() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return &http.Client{Timeout: 10 * time.Second}
}

var reUsuario = regexp.MustCompile(`^user_[A-Za-z0-9]{1,64}$`)

type reclamos struct {
	Sub string  `json:"sub"`
	Iss string  `json:"iss"`
	Azp string  `json:"azp"`
	Exp float64 `json:"exp"`
	Nbf float64 `json:"nbf"`
}

// Verificar valida el token de sesión y devuelve el id de la persona en Clerk (user_…).
func (c *Cliente) Verificar(ctx context.Context, token string) (string, error) {
	partes := strings.Split(token, ".")
	if len(token) > 8<<10 || len(partes) != 3 {
		return "", ErrToken
	}
	var cab struct{ Alg, Kid, Typ string }
	if !decodificar(partes[0], &cab) || cab.Alg != "RS256" || cab.Kid == "" {
		return "", ErrToken
	}
	var r reclamos
	if !decodificar(partes[1], &r) {
		return "", ErrToken
	}
	firma, err := base64.RawURLEncoding.DecodeString(partes[2])
	if err != nil {
		return "", ErrToken
	}
	clave, err := c.clave(ctx, cab.Kid)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256([]byte(partes[0] + "." + partes[1]))
	if rsa.VerifyPKCS1v15(clave, crypto.SHA256, h[:], firma) != nil {
		return "", ErrToken
	}
	ahora := c.ahora()
	switch {
	case r.Iss != c.Emisor,
		!reUsuario.MatchString(r.Sub),
		r.Exp == 0 || ahora.After(time.Unix(int64(r.Exp), 0).Add(margen)),
		r.Nbf != 0 && ahora.Add(margen).Before(time.Unix(int64(r.Nbf), 0)),
		r.Azp != "" && !slices.Contains(c.Origenes, r.Azp):
		return "", ErrToken
	}
	return r.Sub, nil
}

func decodificar(parte string, v any) bool {
	b, err := base64.RawURLEncoding.DecodeString(parte)
	return err == nil && json.Unmarshal(b, v) == nil
}

// clave devuelve la clave pública del kid; si no está, vuelve a bajar el JWKS (una vez por minuto).
func (c *Cliente) clave(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if k, ok := c.claves[kid]; ok {
		return k, nil
	}
	if !c.buscadas.IsZero() && c.ahora().Sub(c.buscadas) < time.Minute {
		return nil, ErrToken
	}
	var jwks struct {
		Keys []struct{ Kid, Kty, N, E string } `json:"keys"`
	}
	if err := c.pedir(ctx, "/v1/jwks", &jwks); err != nil {
		return nil, fmt.Errorf("clerk: jwks: %w", err)
	}
	c.buscadas = c.ahora()
	c.claves = map[string]*rsa.PublicKey{}
	for _, k := range jwks.Keys {
		n, err1 := base64.RawURLEncoding.DecodeString(k.N)
		e, err2 := base64.RawURLEncoding.DecodeString(k.E)
		if k.Kty != "RSA" || err1 != nil || err2 != nil || len(e) > 4 {
			continue
		}
		c.claves[k.Kid] = &rsa.PublicKey{N: new(big.Int).SetBytes(n), E: int(new(big.Int).SetBytes(e).Int64())}
	}
	if k, ok := c.claves[kid]; ok {
		return k, nil
	}
	return nil, ErrToken
}

// Usuario es lo que Tu Local Web toma de la persona en Clerk.
type Usuario struct {
	ID string
	// Email: el principal si está verificado; si no, el primero verificado. Vacío: ninguno lo está.
	Email            string
	Nombre, Apellido string
	// Foto de perfil (la de Google o Apple, o la que subió). Vacía: no tiene.
	Foto string
}

// Usuario lee a la persona con la Backend API.
func (c *Cliente) Usuario(ctx context.Context, id string) (Usuario, error) {
	if !reUsuario.MatchString(id) {
		return Usuario{}, ErrToken
	}
	var r struct {
		ID             string `json:"id"`
		PrimaryEmailID string `json:"primary_email_address_id"`
		FirstName      string `json:"first_name"`
		LastName       string `json:"last_name"`
		ImageURL       string `json:"image_url"`
		HasImage       bool   `json:"has_image"`
		Emails         []struct {
			ID           string `json:"id"`
			Email        string `json:"email_address"`
			Verification *struct {
				Status string `json:"status"`
			} `json:"verification"`
		} `json:"email_addresses"`
	}
	if err := c.pedir(ctx, "/v1/users/"+id, &r); err != nil {
		return Usuario{}, fmt.Errorf("clerk: usuario: %w", err)
	}
	u := Usuario{ID: r.ID, Nombre: strings.TrimSpace(r.FirstName), Apellido: strings.TrimSpace(r.LastName)}
	if r.HasImage && strings.HasPrefix(r.ImageURL, "https://") && len(r.ImageURL) <= 1000 {
		u.Foto = r.ImageURL
	}
	verificado := func(i int) bool {
		v := r.Emails[i].Verification
		return v != nil && v.Status == "verified" && r.Emails[i].Email != ""
	}
	for i := range r.Emails {
		if r.Emails[i].ID == r.PrimaryEmailID && verificado(i) {
			u.Email = r.Emails[i].Email
		}
	}
	for i := range r.Emails {
		if u.Email == "" && verificado(i) {
			u.Email = r.Emails[i].Email
		}
	}
	return u, nil
}

func (c *Cliente) pedir(ctx context.Context, ruta string, v any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.api()+ruta, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.Secreto)
	req.Header.Set("Accept", "application/json")
	res, err := c.cliente().Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(res.Body, 256<<10))
	if res.StatusCode >= 300 {
		return fmt.Errorf("respondió %d", res.StatusCode)
	}
	return json.Unmarshal(b, v)
}
