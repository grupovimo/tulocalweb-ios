package correo

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestResendMandaLoQueCorresponde(t *testing.T) {
	var auth string
	var cuerpo map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &cuerpo)
		w.WriteHeader(200)
	}))
	defer srv.Close()
	r := &Resend{Clave: "re_x", De: "Tu Local Web <hola@tulocalweb.com.ar>", URL: srv.URL}
	if err := r.Enviar(context.Background(), Mensaje{Para: "a@b.co", Asunto: "Hola", HTML: "<b>h</b>", Texto: "h"}); err != nil {
		t.Fatal(err)
	}
	if auth != "Bearer re_x" || cuerpo["subject"] != "Hola" || cuerpo["to"].([]any)[0] != "a@b.co" {
		t.Fatalf("auth=%q cuerpo=%v", auth, cuerpo)
	}
}

func TestResendDevuelveElError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(422)
		_, _ = w.Write([]byte(`{"message":"dominio sin verificar"}`))
	}))
	defer srv.Close()
	err := (&Resend{Clave: "x", URL: srv.URL}).Enviar(context.Background(), Mensaje{Para: "a@b.co"})
	if err == nil || !strings.Contains(err.Error(), "dominio sin verificar") {
		t.Fatalf("err = %v", err)
	}
}

func TestDireccionDe(t *testing.T) {
	if got := direccionDe("Tu Local Web <hola@x.dev>"); got != "hola@x.dev" {
		t.Fatal(got)
	}
}

func TestElegir(t *testing.T) {
	if _, ok := Elegir("re_x", "mail:465", "u", "c", "De <a@b.c>").(*Resend); !ok {
		t.Fatal("con clave de Resend tiene que usar Resend")
	}
	s, ok := Elegir("", "mail.tulocalweb.com.ar:465", "u", "c", "De <a@b.c>").(*SMTP)
	if !ok || s.Usuario != "u" || s.Clave != "c" {
		t.Fatalf("sin Resend tiene que usar SMTP con usuario: %+v", s)
	}
}

func TestEncabezadosCodificados(t *testing.T) {
	if got := de("Tu Local Web <hola@tulocalweb.com.ar>"); got != "Tu Local Web <hola@tulocalweb.com.ar>" {
		t.Errorf("un nombre en ASCII queda igual: %q", got)
	}
	if got := de("Señora Ñoño <a@b.c>"); !strings.HasPrefix(got, "=?utf-8?q?") || !strings.HasSuffix(got, " <a@b.c>") {
		t.Errorf("un nombre con tildes va codificado: %q", got)
	}
	id := idMensaje("Tu Local Web <hola@tulocalweb.com.ar>")
	if !strings.HasPrefix(id, "<") || !strings.HasSuffix(id, "@tulocalweb.com.ar>") || id == idMensaje("x <hola@tulocalweb.com.ar>") {
		t.Errorf("Message-ID único y con el dominio: %q", id)
	}
}

func TestRemitenteConElNombreDeLaTiendaYEncabezados(t *testing.T) {
	var cuerpo map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &cuerpo)
		w.WriteHeader(200)
	}))
	defer srv.Close()
	r := &Resend{Clave: "x", De: "Tu Local Web <hola@tulocalweb.com.ar>", URL: srv.URL}
	m := Mensaje{Para: "a@b.co", Asunto: "Hola", DeNombre: `Café "Bruma", <Palermo>`,
		Encabezados: map[string]string{"List-Unsubscribe-Post": "List-Unsubscribe=One-Click"}}
	if err := r.Enviar(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	if cuerpo["from"] != "Café Bruma Palermo <hola@tulocalweb.com.ar>" {
		t.Fatalf("from = %v", cuerpo["from"])
	}
	if h, _ := cuerpo["headers"].(map[string]any); h["List-Unsubscribe-Post"] != "List-Unsubscribe=One-Click" {
		t.Fatalf("headers = %v", cuerpo["headers"])
	}
	// Un encabezado con salto de línea no sale por SMTP.
	m.Encabezados = map[string]string{"X-Algo": "a\r\nBcc: x@y.z"}
	if err := (&SMTP{Direccion: "127.0.0.1:1", De: "hola@tulocalweb.com.ar"}).Enviar(context.Background(), m); err == nil ||
		!strings.Contains(err.Error(), "encabezado") {
		t.Fatalf("err = %v", err)
	}
}
