package push_test

import (
	"context"
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/push"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/servidor"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/webpush"
)

// suscripcionDePrueba: lo que daría PushSubscription.toJSON() en un navegador, con claves de verdad.
func suscripcionDePrueba(t *testing.T, endpoint string) push.SuscripcionWeb {
	t.Helper()
	k, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	a := make([]byte, 16)
	_, _ = rand.Read(a)
	enc := base64.RawURLEncoding
	return push.SuscripcionWeb{Endpoint: endpoint, P256dh: enc.EncodeToString(k.PublicKey().Bytes()), Auth: enc.EncodeToString(a), Aparato: "iPhone · Safari"}
}

var ingresoMac = cuentas.Cliente{IP: "190.210.32.1", Agente: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)"}

func endpointFCM(c string) string {
	return "https://fcm.googleapis.com/fcm/send/" + strings.Repeat(c, 20)
}

// conWeb arma el handler con el push web prendido (clave pública VAPID).
func (e *escenario) conWeb(t *testing.T) {
	t.Helper()
	pub, _, err := webpush.GenerarClaves()
	if err != nil {
		t.Fatal(err)
	}
	e.h = servidor.Nuevo(servidor.Dependencias{App: e.b.App, Cuentas: e.cs, Log: silencio(), VAPIDPublica: pub})
}

func (e *escenario) web(sesion, metodo, ruta, cuerpo string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo))
	if sesion != "" {
		r.Header.Set("Authorization", "Bearer "+sesion)
	}
	r.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.h.ServeHTTP(rec, r)
	return rec
}

func (e *escenario) suscribir(t *testing.T, rol string, s push.SuscripcionWeb) {
	t.Helper()
	b, _ := json.Marshal(s)
	if rec := e.web(e.ses[rol], "POST", "/v1/panel/push-web", string(b)); rec.Code != http.StatusNoContent {
		t.Fatalf("suscribir %s: %d %s", rol, rec.Code, rec.Body)
	}
}

func (e *escenario) suscripciones(t *testing.T, donde string, args ...any) int {
	t.Helper()
	conn, err := pgx.Connect(context.Background(), e.b.AdminURL)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(context.Background())
	var n int
	if err := conn.QueryRow(context.Background(), `SELECT count(*) FROM auth.suscripciones_web WHERE `+donde, args...).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// webFalso registra lo enviado y responde ErrVencida a los endpoints muertos.
type webFalso struct {
	mu      sync.Mutex
	envios  []webpush.Envio
	muertos map[string]bool
}

func (f *webFalso) Enviar(_ context.Context, e webpush.Envio) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.envios = append(f.envios, e)
	if f.muertos[e.Endpoint] {
		return webpush.ErrVencida
	}
	return nil
}

type avisoWeb struct {
	Title, Body, URL, Tienda, Tag string
}

func (f *webFalso) avisos(t *testing.T) map[string]avisoWeb {
	t.Helper()
	out := map[string]avisoWeb{}
	for _, e := range f.envios {
		var a avisoWeb
		if err := json.Unmarshal(e.Payload, &a); err != nil {
			t.Fatal(err)
		}
		out[e.Endpoint] = a
	}
	return out
}

func TestSuscripcionWeb(t *testing.T) {
	e := armar(t)
	// Sin VAPID: el panel sabe que está apagado y no se puede suscribir.
	if rec := e.web(e.ses["duenio"], "GET", "/v1/panel/push-web", ""); rec.Code != 200 || !strings.Contains(rec.Body.String(), `"clave":""`) {
		t.Fatalf("estado apagado: %d %s", rec.Code, rec.Body)
	}
	b, _ := json.Marshal(suscripcionDePrueba(t, endpointFCM("a")))
	if rec := e.web(e.ses["duenio"], "POST", "/v1/panel/push-web", string(b)); rec.Code != http.StatusConflict {
		t.Fatalf("suscribir apagado: %d", rec.Code)
	}
	e.conWeb(t)
	if rec := e.web("", "POST", "/v1/panel/push-web", string(b)); rec.Code != http.StatusUnauthorized {
		t.Fatalf("sin sesión: %d", rec.Code)
	}
	// Un endpoint que no es de un servicio de push conocido no entra (el worker le haría un POST).
	for _, mala := range []push.SuscripcionWeb{suscripcionDePrueba(t, "https://169.254.169.254/x"), {Endpoint: endpointFCM("a"), P256dh: "x", Auth: "y"}} {
		b, _ := json.Marshal(mala)
		if rec := e.web(e.ses["duenio"], "POST", "/v1/panel/push-web", string(b)); rec.Code != http.StatusBadRequest {
			t.Fatalf("suscripción mala %+v: %d", mala, rec.Code)
		}
	}
	// Prueba sin suscripción: pide activarla primero.
	if rec := e.web(e.ses["duenio"], "POST", "/v1/panel/push-web/prueba", ""); rec.Code != http.StatusConflict {
		t.Fatalf("prueba sin suscripción: %d", rec.Code)
	}
	for rol, c := range map[string]string{"duenio": "a", "admin": "b", "ventas": "c", "local": "d"} {
		e.suscribir(t, rol, suscripcionDePrueba(t, endpointFCM(c)))
	}
	if rec := e.web(e.ses["duenio"], "GET", "/v1/panel/push-web", ""); !strings.Contains(rec.Body.String(), `"suscripta":true`) {
		t.Fatalf("estado: %s", rec.Body)
	}
	// La misma suscripción con otra cuenta en ese navegador pasa a esa.
	e.suscribir(t, "local", suscripcionDePrueba(t, endpointFCM("a")))
	if n := e.suscripciones(t, "usuario_id = $1", e.ids["duenio"]); n != 0 {
		t.Fatalf("a la dueña le quedaron %d", n)
	}
	e.suscribir(t, "duenio", suscripcionDePrueba(t, endpointFCM("a")))
	// RLS: cada persona ve sólo las suyas.
	var vistas int
	if err := e.b.App.EnUsuario(context.Background(), e.ids["ventas"], func(tx pgx.Tx) error {
		return tx.QueryRow(context.Background(), `SELECT count(*) FROM auth.suscripciones_web`).Scan(&vistas)
	}); err != nil || vistas != 1 {
		t.Fatalf("ventas ve %d (%v)", vistas, err)
	}
	// DELETE da de baja la de la sesión; salir se lleva la suya.
	if rec := e.web(e.ses["admin"], "DELETE", "/v1/panel/push-web", ""); rec.Code != http.StatusNoContent {
		t.Fatalf("baja: %d", rec.Code)
	}
	if err := e.cs.Salir(context.Background(), e.ses["ventas"]); err != nil {
		t.Fatal(err)
	}
	if n := e.suscripciones(t, "true"); n != 2 {
		t.Fatalf("quedaron %d suscripciones, quería 2 (dueña y local)", n)
	}
}

func TestVentaYPruebaPorPushWeb(t *testing.T) {
	e := armar(t)
	e.conWeb(t)
	ctx := context.Background()
	for rol, c := range map[string]string{"duenio": "a", "admin": "b", "ventas": "c", "local": "d"} {
		e.suscribir(t, rol, suscripcionDePrueba(t, endpointFCM(c)))
	}
	f := &webFalso{muertos: map[string]bool{endpointFCM("d"): true}}
	s := &push.Servicio{Pool: e.b.App, Web: f, Log: silencio()}
	w := &outbox.Worker{Pool: e.b.Worker, Manejadores: s.Manejadores()}
	// Los ingresos de armar: nadie tenía suscripción todavía.
	if _, err := w.Pasada(ctx); err != nil {
		t.Fatal(err)
	}
	f.envios = nil
	vender(t, e)
	if _, err := w.Pasada(ctx); err != nil {
		t.Fatal(err)
	}
	got := f.avisos(t)
	if len(got) != 4 {
		t.Fatalf("salieron %d avisos, quería 4: %v", len(got), got)
	}
	for c, conMonto := range map[string]bool{"a": true, "b": true, "c": false, "d": false} {
		a := got[endpointFCM(c)]
		if strings.Contains(a.Body, "$ 24.020") != conMonto || strings.Contains(a.Body, "Lucía") || !strings.HasPrefix(a.Title, "Nueva venta #") {
			t.Errorf("%s: %+v", c, a)
		}
		if !strings.HasPrefix(a.URL, "/panel/pedidos/") || a.Tienda != e.tienda || !strings.HasPrefix(a.Tag, "venta-") {
			t.Errorf("%s: %+v", c, a)
		}
	}
	if f.envios[0].Tema == "" || f.envios[0].TTL <= 0 {
		t.Errorf("envío sin Topic o TTL: %+v", f.envios[0])
	}
	// La suscripción vencida (410) se borró sola.
	if n := e.suscripciones(t, "endpoint = $1", endpointFCM("d")); n != 0 {
		t.Fatal("la suscripción vencida sigue")
	}

	// «Mandame un aviso de prueba»: sólo a la suscripción de esa sesión.
	f.envios = nil
	if rec := e.web(e.ses["ventas"], "POST", "/v1/panel/push-web/prueba", ""); rec.Code != http.StatusAccepted {
		t.Fatalf("prueba: %d %s", rec.Code, rec.Body)
	}
	if _, err := w.Pasada(ctx); err != nil {
		t.Fatal(err)
	}
	if len(f.envios) != 1 || f.envios[0].Endpoint != endpointFCM("c") {
		t.Fatalf("prueba: %+v", f.envios)
	}
	if a := f.avisos(t)[endpointFCM("c")]; a.Title != "Aviso de prueba" || a.URL != "/panel/notificaciones" {
		t.Errorf("prueba: %+v", a)
	}

	// «Entraron a tu cuenta»: sólo a la persona, no a la sesión nueva.
	f.envios = nil
	if _, err := e.cs.Ingresar(ctx, "duenio@luna.com", "clave-larga-1", false, ingresoMac); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Pasada(ctx); err != nil {
		t.Fatal(err)
	}
	if len(f.envios) != 1 || f.envios[0].Endpoint != endpointFCM("a") {
		t.Fatalf("ingreso: %+v", f.envios)
	}
	if a := f.avisos(t)[endpointFCM("a")]; a.Title != "Entraron a tu cuenta" || a.URL != "/panel/configuracion?tab=seguridad" || strings.Contains(a.Body, "190.") {
		t.Errorf("ingreso: %+v", a)
	}
}

// Con el cliente de Web Push de verdad contra un servicio falso: llega cifrado con VAPID, y un 410
// borra la suscripción.
func TestPushWebContraServicioFalso(t *testing.T) {
	e := armar(t)
	e.conWeb(t)
	ctx := context.Background()
	e.suscribir(t, "duenio", suscripcionDePrueba(t, endpointFCM("a")))
	e.suscribir(t, "admin", suscripcionDePrueba(t, endpointFCM("b")))
	var mu sync.Mutex
	var pedidos []*http.Request
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		pedidos = append(pedidos, r)
		mu.Unlock()
		if strings.HasSuffix(r.URL.Path, "/vencida") {
			w.WriteHeader(http.StatusGone)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()
	// Los endpoints apuntan al servicio falso (el registro sólo acepta los servicios conocidos).
	e.admin(t, `UPDATE auth.suscripciones_web SET endpoint = $1 WHERE endpoint = $2`, srv.URL+"/ok", endpointFCM("a"))
	e.admin(t, `UPDATE auth.suscripciones_web SET endpoint = $1 WHERE endpoint = $2`, srv.URL+"/vencida", endpointFCM("b"))
	pub, priv, _ := webpush.GenerarClaves()
	cli, err := webpush.Desde(pub, priv, "mailto:hola@tulocalweb.com.ar")
	if err != nil {
		t.Fatal(err)
	}
	u, _ := url.Parse(srv.URL)
	cli.HTTP, cli.Hosts = srv.Client(), []string{u.Hostname()}
	s := push.Nuevo(e.b.App, nil, silencio()).ConWeb(cli)
	w := &outbox.Worker{Pool: e.b.Worker, Manejadores: s.Manejadores()}
	vender(t, e)
	if _, err := w.Pasada(ctx); err != nil {
		t.Fatal(err)
	}
	if len(pedidos) != 2 {
		t.Fatalf("pedidos al servicio: %d", len(pedidos))
	}
	for _, r := range pedidos {
		if r.Header.Get("Content-Encoding") != "aes128gcm" || !strings.HasPrefix(r.Header.Get("Authorization"), "vapid t=") ||
			!strings.HasSuffix(r.Header.Get("Authorization"), ", k="+pub) {
			t.Errorf("encabezados: %v", r.Header)
		}
	}
	if n := e.suscripciones(t, "endpoint = $1", srv.URL+"/vencida"); n != 0 {
		t.Fatal("el 410 no borró la suscripción")
	}
	if n := e.suscripciones(t, "endpoint = $1", srv.URL+"/ok"); n != 1 {
		t.Fatal("se borró una suscripción que andaba")
	}
}
