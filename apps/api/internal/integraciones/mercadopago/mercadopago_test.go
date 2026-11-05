package mercadopago_test

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/links"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/turnos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/mercadopago"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/cifrado"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

// mpFalso imita lo que usa Tu Local Web de api.mercadopago.com.
type mpFalso struct {
	mu        sync.Mutex
	pedidos   []string // "MÉTODO ruta token"
	cuerpos   map[string]map[string]any
	pagos     map[string]map[string]any
	susc      map[string]map[string]any
	autoriz   map[string]map[string]any // cobros de débitos (authorized_payments)
	usuario   int64
	refreshOK bool
}

func (f *mpFalso) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	tok := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	f.pedidos = append(f.pedidos, r.Method+" "+r.URL.Path+" "+tok)
	b, _ := io.ReadAll(r.Body)
	var c map[string]any
	_ = json.Unmarshal(b, &c)
	f.cuerpos[r.Method+" "+r.URL.Path] = c
	enviar := func(v any) { _ = json.NewEncoder(w).Encode(v) }
	switch {
	case r.URL.Path == "/oauth/token":
		v, _ := url.ParseQuery(string(b))
		if v.Get("grant_type") == "authorization_code" && len(v.Get("code_verifier")) < 43 {
			w.WriteHeader(http.StatusBadRequest) // PKCE obligatorio
			return
		}
		if v.Get("grant_type") == "refresh_token" && !f.refreshOK {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		enviar(map[string]any{"access_token": "APP_USR-vendedor-" + v.Get("grant_type"), "refresh_token": "TG-renovar",
			"expires_in": 15552000, "scope": "offline_access payments", "user_id": f.usuario})
	case r.URL.Path == "/checkout/preferences":
		enviar(map[string]any{"id": "pref-1", "init_point": "https://www.mercadopago.com.ar/checkout/v1/redirect?pref_id=pref-1"})
	case strings.HasPrefix(r.URL.Path, "/v1/payments/") && strings.HasSuffix(r.URL.Path, "/refunds"):
		id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v1/payments/"), "/refunds")
		p, ok := f.pagos[id]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		p["status"] = "refunded"
		enviar(map[string]any{"id": 1, "payment_id": p["id"], "status": "approved"})
	case strings.HasPrefix(r.URL.Path, "/v1/payments/"):
		p, ok := f.pagos[strings.TrimPrefix(r.URL.Path, "/v1/payments/")]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		enviar(p)
	case r.URL.Path == "/preapproval" && r.Method == http.MethodPost:
		enviar(map[string]any{"id": "sus-2", "status": "pending", "init_point": "https://www.mercadopago.com.ar/subscriptions/checkout?preapproval_id=sus-2"})
	case strings.HasPrefix(r.URL.Path, "/preapproval/"):
		id := strings.TrimPrefix(r.URL.Path, "/preapproval/")
		if r.Method == http.MethodPut {
			// Cambia lo que viene: el estado (baja, pausa) o el monto del débito.
			if v, ok := c["status"]; ok {
				f.susc[id]["status"] = v
			}
			if rec, ok := c["auto_recurring"].(map[string]any); ok {
				f.susc[id]["transaction_amount"] = rec["transaction_amount"]
			}
		}
		enviar(f.susc[id])
	case strings.HasPrefix(r.URL.Path, "/authorized_payments/"):
		a, ok := f.autoriz[strings.TrimPrefix(r.URL.Path, "/authorized_payments/")]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		enviar(a)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (f *mpFalso) vio(pref string) []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []string
	for _, p := range f.pedidos {
		if strings.HasPrefix(p, pref) {
			out = append(out, p)
		}
	}
	return out
}

func firmar(secreto, dataID, req string, ts int64) string {
	m := hmac.New(sha256.New, []byte(secreto))
	fmt.Fprintf(m, "id:%s;request-id:%s;ts:%d;", dataID, req, ts)
	return fmt.Sprintf("ts=%d,v1=%s", ts, hex.EncodeToString(m.Sum(nil)))
}

func TestMercadoPagoDePuntaAPunta(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	sol := pruebadb.CrearTienda(t, b.App, "sol")
	var ana string
	if err := b.App.EnTx(ctx, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `INSERT INTO auth.usuarios (email, nombre) VALUES ('ana@luna.com', 'Ana') RETURNING id::text`).Scan(&ana)
	}); err != nil {
		t.Fatal(err)
	}

	f := &mpFalso{cuerpos: map[string]map[string]any{}, pagos: map[string]map[string]any{}, susc: map[string]map[string]any{}, usuario: 777, refreshOK: true}
	srv := httptest.NewServer(f)
	defer srv.Close()
	k := make([]byte, 32)
	_, _ = rand.Read(k)
	cif, _ := cifrado.Nuevo(base64.StdEncoding.EncodeToString(k))
	ahora := time.Now()
	s := &mercadopago.Servicio{Pool: b.App, Cifra: cif, Cliente: &mercadopago.Cliente{Base: srv.URL},
		Ahora:  func() time.Time { return ahora },
		Config: mercadopago.Config{AppID: "123", Secreto: "sec", TokenPlataforma: "APP_USR-plataforma", SecretoWebhook: "whsec", Sitio: "https://tulocalweb.com.ar"}}
	mux := http.NewServeMux()
	s.Rutas(mux)

	// Sin conexión no se cobra (y nunca con el token de la plataforma).
	cod, _ := links.Crear(ctx, b.App, luna, links.Nuevo{Monto: 42900, Concepto: "Termo"})
	if _, err := s.PagarLink(ctx, luna, cod); !errors.Is(err, mercadopago.ErrSinConexion) {
		t.Fatalf("cobró sin conexión: %v", err)
	}

	// OAuth: el state es de un solo uso y de esa tienda.
	u, err := s.Conectar(ctx, luna, ana)
	if err != nil {
		t.Fatal(err)
	}
	q, _ := url.Parse(u)
	estado := q.Query().Get("state")
	if q.Query().Get("code_challenge_method") != "S256" || q.Query().Get("platform_id") != "mp" || q.Query().Get("redirect_uri") != "https://tulocalweb.com.ar/integraciones/mercadopago/callback" {
		t.Fatalf("url de autorización: %s", u)
	}
	if err := s.Volver(ctx, sol, ana, "codigo", estado); !errors.Is(err, mercadopago.ErrEstado) {
		t.Fatal("otra tienda usó el state")
	}
	if err := s.Volver(ctx, luna, ana, "codigo", estado); err != nil {
		t.Fatal(err)
	}
	if err := s.Volver(ctx, luna, ana, "codigo", estado); !errors.Is(err, mercadopago.ErrEstado) {
		t.Fatal("el state se usó dos veces")
	}
	// El token quedó cifrado.
	var crudo []byte
	_ = b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT acceso FROM mp_conexiones WHERE tienda_id = $1`, luna).Scan(&crudo)
	})
	if bytes.Contains(crudo, []byte("APP_USR")) {
		t.Fatal("el token está en claro en la base")
	}

	// Link de pago: preferencia con el token de la tienda, sin comisión, y un doble clic no crea dos.
	u1, err := s.PagarLink(ctx, luna, cod)
	if err != nil || !strings.Contains(u1, "pref_id=pref-1") {
		t.Fatalf("pagar: %q %v", u1, err)
	}
	u2, _ := s.PagarLink(ctx, luna, cod)
	if u2 != u1 || len(f.vio("POST /checkout/preferences APP_USR-vendedor-authorization_code")) != 1 {
		t.Fatalf("preferencias: %v", f.vio("POST /checkout/preferences"))
	}
	pref := f.cuerpos["POST /checkout/preferences"]
	if _, hay := pref["marketplace_fee"]; hay || pref["auto_return"] != "approved" {
		t.Fatalf("preferencia: %v", pref)
	}
	ref := pref["external_reference"].(string)

	// Webhook: sin firma, 401; con firma, va al outbox y el worker concilia.
	f.pagos["555"] = map[string]any{"id": 555, "status": "approved", "external_reference": ref, "transaction_amount": 42900}
	aviso := `{"type":"payment","action":"payment.updated","user_id":777,"data":{"id":"555"}}`
	pedir := func(firma string) int {
		r := httptest.NewRequest(http.MethodPost, "/v1/webhooks/mercadopago?data.id=555&type=payment", strings.NewReader(aviso))
		r.Header.Set("x-request-id", "req-1")
		r.Header.Set("x-signature", firma)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		return w.Code
	}
	if c := pedir("ts=1,v1=00"); c != http.StatusUnauthorized {
		t.Fatalf("aceptó una firma falsa: %d", c)
	}
	if c := pedir(firmar("whsec", "555", "req-1", ahora.Add(-time.Hour).UnixMilli())); c != http.StatusUnauthorized {
		t.Fatalf("aceptó un aviso viejo: %d", c)
	}
	if c := pedir(firmar("whsec", "555", "req-1", ahora.UnixMilli())); c != http.StatusOK {
		t.Fatalf("webhook firmado: %d", c)
	}
	w := &outbox.Worker{Pool: b.Worker, Manejadores: s.Manejadores()}
	if _, err := w.Pasada(ctx); err != nil {
		t.Fatal(err)
	}
	if l, _ := links.Ver(ctx, b.App, cod); !l.Pagado {
		t.Fatal("el link no quedó pagado")
	}
	if _, err := s.PagarLink(ctx, luna, cod); !errors.Is(err, mercadopago.ErrYaPagado) {
		t.Fatal("dejó pagar dos veces")
	}

	// Renovación: cerca del vencimiento se renueva; si falla, queda para reconectar (sin otro token).
	ahora = ahora.Add(178 * 24 * time.Hour)
	f.refreshOK = false
	cod2, _ := links.Crear(ctx, b.App, luna, links.Nuevo{Monto: 1000, Concepto: "Otro"})
	if _, err := s.PagarLink(ctx, luna, cod2); !errors.Is(err, mercadopago.ErrReautorizar) {
		t.Fatalf("con el refresh caído: %v", err)
	}
	if e, _ := s.EstadoDe(ctx, luna); !e.Reautorizar {
		t.Fatal("no marcó reconectar")
	}
	if len(f.vio("POST /checkout/preferences APP_USR-plataforma")) != 0 {
		t.Fatal("usó el token de la plataforma para cobrar")
	}

	// Abono: pending con el catálogo del servidor; el plan cambia recién con el aviso.
	url3, err := s.Suscribir(ctx, luna, ana, "pro", "anio")
	if err != nil || !strings.Contains(url3, "preapproval_id=sus-2") {
		t.Fatalf("suscribir: %q %v", url3, err)
	}
	cuerpo := f.cuerpos["POST /preapproval"]
	rec := cuerpo["auto_recurring"].(map[string]any)
	if cuerpo["status"] != "pending" || rec["transaction_amount"].(float64) != 449100 || rec["frequency"].(float64) != 12 {
		t.Fatalf("preapproval: %v", cuerpo)
	}
	plan := func() string {
		var p string
		_ = b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
			return tx.QueryRow(ctx, `SELECT plan FROM tiendas WHERE id = $1`, luna).Scan(&p)
		})
		return p
	}
	if plan() != "gratis" {
		t.Fatal("cambió el plan antes de pagar")
	}
	f.susc["sus-2"] = map[string]any{"id": "sus-2", "status": "authorized", "external_reference": luna}
	if err := s.Recibir(ctx, mercadopago.Aviso{Type: "subscription_preapproval", Data: struct {
		ID json.RawMessage `json:"id"`
	}{ID: json.RawMessage(`"sus-2"`)}}); err != nil {
		t.Fatal(err)
	}
	if plan() != "pro" {
		t.Fatalf("plan después del aviso: %s", plan())
	}
	// Una suscripción ajena (otra referencia) no toca a la tienda.
	f.susc["sus-x"] = map[string]any{"id": "sus-x", "status": "cancelled", "external_reference": sol}
	_ = s.ConciliarSuscripcion(ctx, "sus-x")
	if plan() != "pro" {
		t.Fatal("una suscripción de otra tienda cambió el plan")
	}
}

func TestFirma(t *testing.T) {
	ahora := time.Now()
	f := firmar("sec", "abc123", "r", ahora.UnixMilli())
	if !mercadopago.FirmaValida("sec", f, "r", "ABC123", ahora) {
		t.Fatal("no aceptó la firma (data.id alfanumérico va en minúsculas)")
	}
	if mercadopago.FirmaValida("otra", f, "r", "abc123", ahora) || mercadopago.FirmaValida("", f, "r", "abc123", ahora) {
		t.Fatal("aceptó con otra clave o sin clave")
	}
	if mercadopago.FirmaValida("sec", f, "r2", "abc123", ahora) {
		t.Fatal("aceptó con otro request id")
	}
}

func TestPrecios(t *testing.T) {
	for _, c := range []struct {
		plan, ciclo string
		precio      int64
	}{{"emprende", "mes", 19900}, {"pro", "mes", 49900}, {"negocio", "anio", 1349100}} {
		if p, ok := mercadopago.PrecioDe(c.plan, c.ciclo); !ok || p != c.precio {
			t.Fatalf("%s %s: %d", c.plan, c.ciclo, p)
		}
	}
	if _, ok := mercadopago.PrecioDe("gratis", "mes"); ok {
		t.Fatal("el plan gratis no se cobra")
	}
}

// La seña de un turno: preferencia por el monto de la seña con el token de la tienda, el webhook
// confirma el turno y devolverla pasa por /refunds.
func TestSenaDeTurno(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	var ana string
	if err := b.App.EnTx(ctx, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `INSERT INTO auth.usuarios (email, nombre) VALUES ('ana@luna.com', 'Ana') RETURNING id::text`).Scan(&ana)
	}); err != nil {
		t.Fatal(err)
	}
	f := &mpFalso{cuerpos: map[string]map[string]any{}, pagos: map[string]map[string]any{}, susc: map[string]map[string]any{}, usuario: 778, refreshOK: true}
	srv := httptest.NewServer(f)
	defer srv.Close()
	k := make([]byte, 32)
	_, _ = rand.Read(k)
	cif, _ := cifrado.Nuevo(base64.StdEncoding.EncodeToString(k))
	ahora := time.Now()
	s := &mercadopago.Servicio{Pool: b.App, Cifra: cif, Cliente: &mercadopago.Cliente{Base: srv.URL},
		Ahora:  func() time.Time { return ahora },
		Config: mercadopago.Config{AppID: "123", Secreto: "sec", SecretoWebhook: "whsec", Sitio: "https://tulocalweb.com.ar"}}
	mux := http.NewServeMux()
	s.Rutas(mux)
	u, err := s.Conectar(ctx, luna, ana)
	if err != nil {
		t.Fatal(err)
	}
	q, _ := url.Parse(u)
	if err := s.Volver(ctx, luna, ana, "codigo", q.Query().Get("state")); err != nil {
		t.Fatal(err)
	}

	// Una barbería que cobra el 30 % de seña.
	// El webhook compara el vencimiento con el reloj de Postgres: el "ahora" del test es el real y
	// el turno, el lunes que viene (a 2 días o más, dentro de los 14 que se ofrecen).
	ya := time.Now().In(turnos.Zona)
	dia := ya.AddDate(0, 0, 2)
	for dia.Weekday() != time.Monday {
		dia = dia.AddDate(0, 0, 1)
	}
	lunes := dia.Format("2006-01-02")
	if _, err := turnos.GuardarAjustes(ctx, b.App, luna, turnos.Ajustes{Activo: true, Sena: 30, AnticipacionMin: 60, DiasVista: 14,
		Horario: []turnos.Franja{{Dia: 1, Desde: "09:00", Hasta: "13:00"}}}); err != nil {
		t.Fatal(err)
	}
	corte, _ := turnos.GuardarServicio(ctx, b.App, luna, "", turnos.Servicio{Nombre: "Corte clásico", Precio: 12000, Duracion: 30, Activo: true}, nil)
	if _, err := turnos.GuardarProfesional(ctx, b.App, luna, "", turnos.Profesional{Nombre: "Beto", Activo: true}, nil); err != nil {
		t.Fatal(err)
	}
	res, err := turnos.Reservar(ctx, b.App, s, s.Config.Sitio, "luna", turnos.Reserva{ServicioID: corte.ID, Inicio: lunes + "T10:00:00-03:00",
		Nombre: "Lucía Pérez", Telefono: "1155550101", Email: "lucia@ejemplo.com"}, ya)
	if err != nil || !strings.Contains(res.URL, "pref_id=pref-1") {
		t.Fatalf("reservar: %+v %v", res, err)
	}
	if len(f.vio("POST /checkout/preferences APP_USR-vendedor-authorization_code")) != 1 {
		t.Fatalf("preferencias: %v", f.vio("POST /checkout/preferences"))
	}
	pref := f.cuerpos["POST /checkout/preferences"]
	items := pref["items"].([]any)
	back := pref["back_urls"].(map[string]any)
	if items[0].(map[string]any)["unit_price"].(float64) != 3600 || pref["expires"] != true ||
		back["success"] != "https://tulocalweb.com.ar/t/luna/turno/"+res.Codigo+"?pago=ok" ||
		pref["payment_methods"].(map[string]any)["installments"].(float64) != 1 {
		t.Fatalf("preferencia de la seña: %v", pref)
	}
	if v, _ := turnos.Ver(ctx, b.App, "luna", res.Codigo, ya); v.Estado != "reservado" || v.SenaPagada {
		t.Fatalf("antes de pagar: %+v", v)
	}

	// El pago aprobado llega por el webhook y el worker confirma el turno.
	f.pagos["556"] = map[string]any{"id": 556, "status": "approved", "external_reference": pref["external_reference"], "transaction_amount": 3600}
	r := httptest.NewRequest(http.MethodPost, "/v1/webhooks/mercadopago?data.id=556&type=payment",
		strings.NewReader(`{"type":"payment","action":"payment.created","user_id":778,"data":{"id":"556"}}`))
	r.Header.Set("x-request-id", "req-2")
	r.Header.Set("x-signature", firmar("whsec", "556", "req-2", ahora.UnixMilli()))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("webhook: %d", w.Code)
	}
	wk := &outbox.Worker{Pool: b.Worker, Manejadores: s.Manejadores()}
	if _, err := wk.Pasada(ctx); err != nil {
		t.Fatal(err)
	}
	v, _ := turnos.Ver(ctx, b.App, "luna", res.Codigo, ya.Add(time.Hour))
	if v.Estado != "confirmado" || !v.SenaPagada || v.VenceEn != nil || v.Vencido {
		t.Fatalf("después del pago: %+v", v)
	}

	// Devolver la seña: refund del pago con el token de la tienda; queda anotada como devuelta.
	l, _ := turnos.Agenda(ctx, b.App, luna, lunes, lunes, "", ya)
	tu, err := turnos.DevolverSena(ctx, b.App, s, luna, l[0].ID, ya)
	if err != nil || !tu.SenaDevuelta || len(f.vio("POST /v1/payments/556/refunds APP_USR-vendedor-authorization_code")) != 1 {
		t.Fatalf("devolver la seña: %+v %v", tu, err)
	}
	if _, err := turnos.DevolverSena(ctx, b.App, s, luna, l[0].ID, ya); !errors.Is(err, turnos.ErrSinSena) {
		t.Fatalf("devolvió dos veces: %v", err)
	}

	// La seña llega tarde: la reserva ya venció y otra persona tomó el horario. El turno vencido no
	// se confirma (sería un doble turno): queda cancelado con la seña pagada, para devolverla.
	tarde, err := turnos.Reservar(ctx, b.App, s, s.Config.Sitio, "luna", turnos.Reserva{ServicioID: corte.ID, Inicio: lunes + "T11:00:00-03:00",
		Nombre: "Tomás", Telefono: "1155550102", Email: "tomas@ejemplo.com"}, ya)
	if err != nil {
		t.Fatal(err)
	}
	refTarde := f.cuerpos["POST /checkout/preferences"]["external_reference"]
	admin, err := pgx.Connect(ctx, b.AdminURL)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close(ctx)
	if _, err := admin.Exec(ctx, `UPDATE turnos SET vence_en = now() - interval '1 minute' WHERE codigo_publico = $1`, tarde.Codigo); err != nil {
		t.Fatal(err)
	}
	otra, err := turnos.Reservar(ctx, b.App, s, s.Config.Sitio, "luna", turnos.Reserva{ServicioID: corte.ID, Inicio: lunes + "T11:00:00-03:00",
		Nombre: "Paula", Telefono: "1155550103", Email: "paula@ejemplo.com"}, ya)
	if err != nil {
		t.Fatalf("el horario vencido no se liberó: %v", err)
	}
	f.pagos["557"] = map[string]any{"id": 557, "status": "approved", "external_reference": refTarde, "transaction_amount": 3600}
	r = httptest.NewRequest(http.MethodPost, "/v1/webhooks/mercadopago?data.id=557&type=payment",
		strings.NewReader(`{"type":"payment","action":"payment.created","user_id":778,"data":{"id":"557"}}`))
	r.Header.Set("x-request-id", "req-3")
	r.Header.Set("x-signature", firmar("whsec", "557", "req-3", ahora.UnixMilli()))
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("webhook tardío: %d", w.Code)
	}
	if _, err := wk.Pasada(ctx); err != nil {
		t.Fatal(err)
	}
	if v, _ := turnos.Ver(ctx, b.App, "luna", tarde.Codigo, ya); v.Estado != "cancelado" || !v.SenaPagada {
		t.Fatalf("seña tardía: %+v", v)
	}
	if v, _ := turnos.Ver(ctx, b.App, "luna", otra.Codigo, ya); v.Estado != "reservado" {
		t.Fatalf("el turno de la otra persona: %+v", v)
	}
}
