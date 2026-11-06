package cuentas_test

import (
	"crypto/hmac"
	"crypto/sha1" //nolint:gosec // el de TOTP
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/cifrado"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/servidor"
)

// totp: el código de ahora, calculado aparte (como lo haría la app autenticadora).
func totp(t *testing.T, b32 string) string {
	s, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(b32)
	if err != nil {
		t.Fatal(err)
	}
	var msg [8]byte
	binary.BigEndian.PutUint64(msg[:], uint64(time.Now().Unix()/30))
	m := hmac.New(sha1.New, s)
	m.Write(msg[:])
	h := m.Sum(nil)
	o := h[len(h)-1] & 0x0f
	return fmt.Sprintf("%06d", (binary.BigEndian.Uint32(h[o:o+4])&0x7fffffff)%1_000_000)
}

// La verificación en dos pasos de punta a punta: prenderla con un QR y un código, que el ingreso
// devuelva un desafío en vez de la sesión, completarlo (mal, bien y con un código de recuperación
// que se gasta) y apagarla.
func TestSegundoFactor(t *testing.T) {
	s, _, b := nuevo(t)
	c, err := cifrado.Nuevo(base64.StdEncoding.EncodeToString([]byte("01234567890123456789012345678901")))
	if err != nil {
		t.Fatal(err)
	}
	s.Cifra = c
	h := servidor.Nuevo(servidor.Dependencias{App: b.App, Cuentas: s})
	rec := pedir(h, "POST", "/v1/auth/registro", `{"nombre":"Lucía","tienda":"La Esquina","email":"lucia@laesquina.com.ar","clave":"Abcdefghijk1"}`, "")
	var ses cuentas.Sesion
	_ = json.Unmarshal(rec.Body.Bytes(), &ses)
	ingresar := func() cuentas.Sesion {
		rec := pedir(h, "POST", "/v1/auth/ingresar", `{"email":"lucia@laesquina.com.ar","clave":"Abcdefghijk1"}`, "")
		if rec.Code != 200 {
			t.Fatalf("ingresar: %d %s", rec.Code, rec.Body)
		}
		var x cuentas.Sesion
		_ = json.Unmarshal(rec.Body.Bytes(), &x)
		return x
	}

	if x := ingresar(); x.Token == "" || x.SegundoFactor != "" {
		t.Fatalf("apagada entra directo: %+v", x)
	}
	rec = pedir(h, "POST", "/v1/auth/segundo-factor/activar", ``, ses.Token)
	var alta cuentas.AltaSegundoFactor
	_ = json.Unmarshal(rec.Body.Bytes(), &alta)
	if rec.Code != 200 || !strings.HasPrefix(alta.URI, "otpauth://totp/Tu%20Local%20Web:lucia@laesquina.com.ar?") {
		t.Fatalf("activar: %d %s", rec.Code, rec.Body)
	}
	if rec := pedir(h, "POST", "/v1/auth/segundo-factor/confirmar", `{"codigo":"000000"}`, ses.Token); rec.Code != 400 {
		t.Fatalf("confirmar con un código mal: %d", rec.Code)
	}
	rec = pedir(h, "POST", "/v1/auth/segundo-factor/confirmar", `{"codigo":"`+totp(t, alta.Secreto)+`"}`, ses.Token)
	var r struct{ Codigos []string }
	_ = json.Unmarshal(rec.Body.Bytes(), &r)
	if rec.Code != 200 || len(r.Codigos) != cuentas.CodigosRecuperacion {
		t.Fatalf("confirmar: %d %s", rec.Code, rec.Body)
	}
	if rec := pedir(h, "GET", "/v1/auth/segundo-factor", ``, ses.Token); !strings.Contains(rec.Body.String(), `"activo":true`) {
		t.Fatalf("estado: %s", rec.Body)
	}

	x := ingresar()
	if x.Token != "" || x.SegundoFactor == "" {
		t.Fatalf("prendida tiene que pedir el código: %+v", x)
	}
	completar := func(desafio, codigo string) (int, cuentas.Sesion) {
		rec := pedir(h, "POST", "/v1/auth/segundo-factor", `{"desafio":"`+desafio+`","codigo":"`+codigo+`"}`, "")
		var y cuentas.Sesion
		_ = json.Unmarshal(rec.Body.Bytes(), &y)
		return rec.Code, y
	}
	if code, _ := completar(x.SegundoFactor, "123456"); code != 400 {
		t.Fatalf("código mal: %d", code)
	}
	if code, y := completar(x.SegundoFactor, totp(t, alta.Secreto)); code != 200 || y.Token == "" {
		t.Fatalf("código bien: %d %+v", code, y)
	}
	if code, _ := completar(x.SegundoFactor, totp(t, alta.Secreto)); code != 401 {
		t.Fatalf("el desafío se usa una vez: %d", code)
	}
	// Un código de recuperación sirve una sola vez.
	x = ingresar()
	if code, y := completar(x.SegundoFactor, strings.ToUpper(r.Codigos[0])); code != 200 || y.Token == "" {
		t.Fatalf("recuperación: %d", code)
	}
	x = ingresar()
	if code, _ := completar(x.SegundoFactor, r.Codigos[0]); code != 400 {
		t.Fatalf("la recuperación ya usada: %d", code)
	}
	// Cinco intentos mal y el desafío no sirve más.
	x = ingresar()
	for range cuentas.MaxIntentosCodigo {
		completar(x.SegundoFactor, "111111")
	}
	if code, _ := completar(x.SegundoFactor, totp(t, alta.Secreto)); code != 401 {
		t.Fatalf("después de cinco intentos: %d", code)
	}
	// Apagarla pide un código.
	if rec := pedir(h, "POST", "/v1/auth/segundo-factor/desactivar", `{"codigo":"000000"}`, ses.Token); rec.Code != 400 {
		t.Fatalf("desactivar sin código bien: %d", rec.Code)
	}
	if rec := pedir(h, "POST", "/v1/auth/segundo-factor/desactivar", `{"codigo":"`+r.Codigos[1]+`"}`, ses.Token); rec.Code != 204 {
		t.Fatalf("desactivar: %d %s", rec.Code, rec.Body)
	}
	if x := ingresar(); x.Token == "" {
		t.Fatal("apagada vuelve a entrar directo")
	}
}
