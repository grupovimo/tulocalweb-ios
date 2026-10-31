package cuentas_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/clerk"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/servidor"
)

func pedir(h http.Handler, metodo, ruta, cuerpo, token string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo))
	r.Header.Set("Content-Type", "application/json")
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, r)
	return rec
}

func TestHTTPDeCuentas(t *testing.T) {
	s, _, b := nuevo(t)
	h := servidor.Nuevo(servidor.Dependencias{App: b.App, Cuentas: s})

	rec := pedir(h, "POST", "/v1/auth/registro", `{"nombre":"Lucía","tienda":"La Esquina","email":"lucia@laesquina.com.ar","clave":"Abcdefghijk1"}`, "")
	if rec.Code != 201 {
		t.Fatalf("registro: %d %s", rec.Code, rec.Body)
	}
	var ses cuentas.Sesion
	_ = json.Unmarshal(rec.Body.Bytes(), &ses)
	if ses.Token == "" || rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("sesión %+v cache %q", ses, rec.Header().Get("Cache-Control"))
	}

	t.Run("la sesión da acceso a su tienda", func(t *testing.T) {
		if rec := pedir(h, "GET", "/v1/auth/sesion", "", ses.Token); rec.Code != 200 || !strings.Contains(rec.Body.String(), `"rol":"duenio"`) {
			t.Fatalf("sesion: %d %s", rec.Code, rec.Body)
		}
		rec := pedir(h, "GET", "/v1/panel/tienda", "", ses.Token)
		if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"slug":"la-esquina"`) {
			t.Fatalf("tienda: %d %s", rec.Code, rec.Body)
		}
		if rec := pedir(h, "GET", "/v1/panel/tienda", "", "inventado"); rec.Code != 401 {
			t.Fatalf("token inventado: %d", rec.Code)
		}
	})

	t.Run("errores como problem+json con código", func(t *testing.T) {
		for _, c := range []struct {
			ruta, cuerpo string
			estado       int
			codigo       string
		}{
			{"/v1/auth/ingresar", `{"email":"lucia@laesquina.com.ar","clave":"mal"}`, 401, "credenciales_invalidas"},
			{"/v1/auth/registro", `{"nombre":"L","tienda":"X","email":"lucia@laesquina.com.ar","clave":"Abcdefghijk1"}`, 409, "email_en_uso"},
			{"/v1/auth/registro", `{"nombre":"L","tienda":"X","email":"x@y.co","clave":"corta"}`, 400, "invalido"},
			{"/v1/auth/restablecer", `{"token":"x","clave":"Abcdefghijk1"}`, 422, "link_invalido"},
			{"/v1/auth/ingresar", `no es json`, 400, "invalido"},
		} {
			rec := pedir(h, "POST", c.ruta, c.cuerpo, "")
			if rec.Code != c.estado || rec.Header().Get("Content-Type") != "application/problem+json" || !strings.Contains(rec.Body.String(), `"code":"`+c.codigo+`"`) {
				t.Errorf("%s: %d %s", c.ruta, rec.Code, rec.Body)
			}
		}
	})

	t.Run("recuperar responde 202 exista o no la cuenta", func(t *testing.T) {
		for _, e := range []string{"lucia@laesquina.com.ar", "nadie@laesquina.com.ar"} {
			if rec := pedir(h, "POST", "/v1/auth/recuperar", `{"email":"`+e+`"}`, ""); rec.Code != 202 {
				t.Fatalf("%s: %d", e, rec.Code)
			}
		}
	})

	t.Run("salir", func(t *testing.T) {
		if rec := pedir(h, "POST", "/v1/auth/salir", "", ses.Token); rec.Code != 204 {
			t.Fatalf("salir: %d", rec.Code)
		}
		if rec := pedir(h, "GET", "/v1/auth/sesion", "", ses.Token); rec.Code != 401 {
			t.Fatalf("después de salir: %d", rec.Code)
		}
	})
}

func TestClienteDeSoloLeCreeAlBFF(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.0.0.1:5555"
	r.Header.Set("X-Cliente-IP", "190.2.3.4")
	if c := cuentas.ClienteDe(r, "secreto"); c.IP != "10.0.0.1" {
		t.Fatalf("sin clave le creyó a X-Cliente-IP: %s", c.IP)
	}
	r.Header.Set("X-Mostrador-BFF", "secreto")
	if c := cuentas.ClienteDe(r, "secreto"); c.IP != "190.2.3.4" {
		t.Fatalf("con clave: %s", c.IP)
	}
	if c := cuentas.ClienteDe(r, ""); c.IP != "10.0.0.1" {
		t.Fatal("sin clave configurada no se le cree a nadie")
	}
}

// La app de iPhone dice cuántos intentos quedan antes de pausar el ingreso (handoff 1b/1c), exista
// o no la cuenta.
func TestIngresoDiceCuantosIntentosQuedan(t *testing.T) {
	s, _, b := nuevo(t)
	h := servidor.Nuevo(servidor.Dependencias{App: b.App, Cuentas: s})
	if rec := pedir(h, "POST", "/v1/auth/registro", `{"nombre":"Lucía","tienda":"La Esquina","email":"lucia@laesquina.com.ar","clave":"Abcdefghijk1"}`, ""); rec.Code != 201 {
		t.Fatalf("registro: %d %s", rec.Code, rec.Body)
	}
	for _, email := range []string{"lucia@laesquina.com.ar", "nadie@laesquina.com.ar"} {
		for quedan := cuentas.MaxIntentosIngreso - 1; quedan >= 0; quedan-- {
			rec := pedir(h, "POST", "/v1/auth/ingresar", `{"email":"`+email+`","clave":"mal"}`, "")
			var p struct {
				Code      string `json:"code"`
				Restantes *int   `json:"restantes"`
			}
			_ = json.Unmarshal(rec.Body.Bytes(), &p)
			if rec.Code != 401 || p.Code != "credenciales_invalidas" || p.Restantes == nil || *p.Restantes != quedan {
				t.Fatalf("%s, quedan %d: %d %s", email, quedan, rec.Code, rec.Body)
			}
		}
		if rec := pedir(h, "POST", "/v1/auth/ingresar", `{"email":"`+email+`","clave":"Abcdefghijk1"}`, ""); rec.Code != 429 || rec.Header().Get("Retry-After") == "" {
			t.Fatalf("%s, después de 5: %d %s", email, rec.Code, rec.Body)
		}
	}
}

// La app de iPhone saca de acá la clave publicable de Clerk: la de la instancia que verifica Go.
func TestProveedoresDaLaClavePublicableDeClerk(t *testing.T) {
	s, _, b := nuevo(t)
	h := servidor.Nuevo(servidor.Dependencias{App: b.App, Cuentas: s})
	leer := func() map[string]any {
		rec := pedir(h, "GET", "/v1/auth/proveedores", "", "")
		var m map[string]any
		if rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &m) != nil {
			t.Fatalf("proveedores: %d %s", rec.Code, rec.Body)
		}
		return m
	}
	if m := leer(); m["clerk"] != false || m["clerkPublicable"] != "" {
		t.Fatalf("sin Clerk: %v", m)
	}
	const pk = "pk_test_Y2xlcmsuZ3J1cG92aW1vLmNvbS5hciQ" // clerk.grupovimo.com.ar$
	c, err := clerk.Nuevo("sk_test_x", pk, nil)
	if err != nil {
		t.Fatal(err)
	}
	s.Clerk = c
	if m := leer(); m["clerk"] != true || m["clerkPublicable"] != pk {
		t.Fatalf("con Clerk: %v", m)
	}
}
