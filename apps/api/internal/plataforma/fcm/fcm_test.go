package fcm

import (
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
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func claveDePrueba(t *testing.T) (*rsa.PrivateKey, string) {
	t.Helper()
	k, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	der, _ := x509.MarshalPKCS8PrivateKey(k)
	return k, string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
}

// falso hace de Google (token) y de FCM (messages:send). responder decide la respuesta del envío.
type falso struct {
	tokens, envios atomic.Int32
	ultimo         map[string]any
	auth           string
	responder      func(n int32) (int, string)
}

func (f *falso) servidor(t *testing.T, clave *rsa.PublicKey) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/token":
			_ = r.ParseForm()
			jwt := r.PostForm.Get("assertion")
			partes := strings.Split(jwt, ".")
			if len(partes) != 3 || r.PostForm.Get("grant_type") != "urn:ietf:params:oauth:grant-type:jwt-bearer" {
				http.Error(w, `{"error":"invalid_grant"}`, 400)
				return
			}
			h := sha256.Sum256([]byte(partes[0] + "." + partes[1]))
			firma, _ := base64.RawURLEncoding.DecodeString(partes[2])
			if rsa.VerifyPKCS1v15(clave, crypto.SHA256, h[:], firma) != nil {
				http.Error(w, `{"error":"invalid_grant"}`, 400)
				return
			}
			var cue map[string]any
			b, _ := base64.RawURLEncoding.DecodeString(partes[1])
			_ = json.Unmarshal(b, &cue)
			if cue["scope"] != Alcance || cue["iss"] != "push@proyecto.iam.gserviceaccount.com" {
				http.Error(w, `{"error":"invalid_scope"}`, 400)
				return
			}
			n := f.tokens.Add(1)
			_, _ = io.WriteString(w, `{"access_token":"acceso-`+string(rune('0'+n))+`","expires_in":3600}`)
		case r.URL.Path == "/v1/projects/proyecto/messages:send":
			n := f.envios.Add(1)
			f.auth = r.Header.Get("authorization")
			b, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(b, &f.ultimo)
			estado, cuerpo := 200, `{"name":"projects/proyecto/messages/1"}`
			if f.responder != nil {
				estado, cuerpo = f.responder(n)
			}
			w.WriteHeader(estado)
			_, _ = io.WriteString(w, cuerpo)
		default:
			http.NotFound(w, r)
		}
	}))
}

func cliente(t *testing.T, f *falso) (*Cliente, *time.Time) {
	t.Helper()
	k, p := claveDePrueba(t)
	srv := f.servidor(t, &k.PublicKey)
	t.Cleanup(srv.Close)
	cuenta, _ := json.Marshal(Cuenta{ProjectID: "proyecto", ClientEmail: "push@proyecto.iam.gserviceaccount.com", PrivateKey: p, TokenURI: srv.URL + "/token"})
	c, err := DeCuenta(cuenta)
	if err != nil {
		t.Fatal(err)
	}
	ahora := time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC)
	c.Host, c.HTTP, c.Ahora = srv.URL, srv.Client(), func() time.Time { return ahora }
	return c, &ahora
}

func TestEnviaYReusaElToken(t *testing.T) {
	f := &falso{}
	c, ahora := cliente(t, f)
	e := Envio{Token: "tok:123", Titulo: "Nueva reserva", Cuerpo: "Ana, corte, mañana 10:30",
		Datos: map[string]string{"ruta": "turnos/1"}, Colapso: "noti-1", Vida: 2 * time.Hour}
	for range 2 {
		if err := c.Enviar(context.Background(), e); err != nil {
			t.Fatal(err)
		}
	}
	if f.tokens.Load() != 1 || f.envios.Load() != 2 {
		t.Fatalf("tokens %d, envíos %d: el token de acceso se tiene que reusar", f.tokens.Load(), f.envios.Load())
	}
	if f.auth != "Bearer acceso-1" {
		t.Fatalf("authorization = %q", f.auth)
	}
	m := f.ultimo["message"].(map[string]any)
	if m["token"] != "tok:123" || m["notification"].(map[string]any)["title"] != "Nueva reserva" ||
		m["data"].(map[string]any)["ruta"] != "turnos/1" {
		t.Fatalf("mensaje: %v", m)
	}
	a := m["android"].(map[string]any)
	if a["ttl"] != "7200s" || a["collapse_key"] != "noti-1" || a["priority"] != "HIGH" {
		t.Fatalf("android: %v", a)
	}
	// Cerca del vencimiento pide otro.
	*ahora = ahora.Add(56 * time.Minute)
	if err := c.Enviar(context.Background(), e); err != nil {
		t.Fatal(err)
	}
	if f.tokens.Load() != 2 {
		t.Fatalf("a 4 minutos de vencer tenía que pedir otro token (pidió %d)", f.tokens.Load())
	}
}

func TestTokenDeDispositivoInvalido(t *testing.T) {
	for _, caso := range []struct {
		estado int
		cuerpo string
	}{
		{404, `{"error":{"code":404,"status":"NOT_FOUND","details":[{"@type":"type.googleapis.com/google.firebase.fcm.v1.FcmError","errorCode":"UNREGISTERED"}]}}`},
		{400, `{"error":{"code":400,"status":"INVALID_ARGUMENT","details":[{"errorCode":"INVALID_ARGUMENT"}]}}`},
	} {
		f := &falso{responder: func(int32) (int, string) { return caso.estado, caso.cuerpo }}
		c, _ := cliente(t, f)
		if err := c.Enviar(context.Background(), Envio{Token: "x", Titulo: "a"}); !errors.Is(err, ErrTokenInvalido) {
			t.Fatalf("HTTP %d: error = %v, quería ErrTokenInvalido", caso.estado, err)
		}
	}
}

func TestReintentaConTokenNuevoSiFCMDice401(t *testing.T) {
	f := &falso{responder: func(n int32) (int, string) {
		if n == 1 {
			return 401, `{"error":{"code":401,"status":"UNAUTHENTICATED"}}`
		}
		return 200, `{}`
	}}
	c, _ := cliente(t, f)
	if err := c.Enviar(context.Background(), Envio{Token: "x", Titulo: "a"}); err != nil {
		t.Fatal(err)
	}
	if f.tokens.Load() != 2 || f.envios.Load() != 2 {
		t.Fatalf("tokens %d, envíos %d", f.tokens.Load(), f.envios.Load())
	}
}

func TestErrorReintentable(t *testing.T) {
	f := &falso{responder: func(int32) (int, string) {
		return 503, `{"error":{"code":503,"status":"UNAVAILABLE"}}`
	}}
	c, _ := cliente(t, f)
	var e *Error
	if err := c.Enviar(context.Background(), Envio{Token: "x", Titulo: "a"}); !errors.As(err, &e) || !e.Reintentable() {
		t.Fatalf("error = %v", err)
	}
}

func TestSinConfiguracionQuedaApagado(t *testing.T) {
	c, err := Desde("")
	if c != nil || err != nil {
		t.Fatalf("Desde(\"\") = %v, %v", c, err)
	}
	if err := c.Enviar(context.Background(), Envio{Token: "x"}); err != nil || c.Activo() {
		t.Fatalf("un cliente nil no hace nada: %v", err)
	}
	if _, err := Desde(filepath.Join(t.TempDir(), "no-existe.json")); err == nil {
		t.Fatal("una ruta que no existe es error")
	}
	ruta := filepath.Join(t.TempDir(), "mal.json")
	_ = os.WriteFile(ruta, []byte(`{"project_id":"p","client_email":"e","private_key":"no"}`), 0o600)
	if _, err := Desde(ruta); err == nil {
		t.Fatal("una clave que no es PEM es error")
	}
}

func TestElErrorDeRedNoLlevaElToken(t *testing.T) {
	f := &falso{}
	c, _ := cliente(t, f)
	if _, err := c.Token(context.Background()); err != nil {
		t.Fatal(err)
	}
	c.Host = "http://127.0.0.1:1"
	err := c.Enviar(context.Background(), Envio{Token: "secreto-del-telefono"})
	if err == nil || strings.Contains(err.Error(), "secreto") || strings.Contains(err.Error(), url.PathEscape("proyecto")+"/") {
		t.Fatalf("error = %v", err)
	}
}

func TestSoloDatos(t *testing.T) {
	var m map[string]map[string]any
	b := Mensaje(Envio{Token: "tok:123", Titulo: "Nueva reserva", Cuerpo: "Ana", Datos: map[string]string{"ruta": "turnos/1"}, Colapso: "c", SoloDatos: true})
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	x := m["message"]
	if _, hay := x["notification"]; hay {
		t.Fatalf("con SoloDatos no va notification: %s", b)
	}
	if _, hay := x["android"].(map[string]any)["notification"]; hay {
		t.Fatalf("con SoloDatos no va android.notification: %s", b)
	}
	d := x["data"].(map[string]any)
	if d["titulo"] != "Nueva reserva" || d["cuerpo"] != "Ana" || d["ruta"] != "turnos/1" {
		t.Fatalf("data: %v", d)
	}
}
