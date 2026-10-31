package mercadopago_test

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/suscripciones"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/vidriera"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/mercadopago"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/cifrado"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

// mpSuscripciones imita la API de suscripciones de Mercado Pago con la cuenta de una tienda:
// preapproval (alta, consulta, cambio de estado) y los cobros autorizados.
type mpSuscripciones struct {
	mu          sync.Mutex
	pedidos     []string
	preapproval map[string]map[string]any
	cobros      map[string]map[string]any
	pagos       map[string]map[string]any
	ultimoAlta  map[string]any
	n           int
}

func (f *mpSuscripciones) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	tok := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	f.pedidos = append(f.pedidos, r.Method+" "+r.URL.Path+" "+tok)
	b, _ := io.ReadAll(r.Body)
	var c map[string]any
	_ = json.Unmarshal(b, &c)
	enviar := func(v any) { _ = json.NewEncoder(w).Encode(v) }
	switch {
	case r.URL.Path == "/oauth/token":
		enviar(map[string]any{"access_token": "APP_USR-vinos", "refresh_token": "TG-r", "expires_in": 15552000, "user_id": 4242})
	case r.URL.Path == "/preapproval" && r.Method == http.MethodPost:
		f.n++
		id := "pre-" + string(rune('0'+f.n))
		f.ultimoAlta = c
		f.preapproval[id] = map[string]any{"id": id, "status": "pending", "external_reference": c["external_reference"]}
		enviar(map[string]any{"id": id, "status": "pending", "init_point": "https://www.mercadopago.com.ar/subscriptions/checkout?preapproval_id=" + id})
	case strings.HasPrefix(r.URL.Path, "/preapproval/"):
		p, ok := f.preapproval[strings.TrimPrefix(r.URL.Path, "/preapproval/")]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Method == http.MethodPut {
			p["status"] = c["status"]
		}
		enviar(p)
	case strings.HasPrefix(r.URL.Path, "/authorized_payments/"):
		p, ok := f.cobros[strings.TrimPrefix(r.URL.Path, "/authorized_payments/")]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		enviar(p)
	case strings.HasPrefix(r.URL.Path, "/v1/payments/"):
		p, ok := f.pagos[strings.TrimPrefix(r.URL.Path, "/v1/payments/")]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		enviar(p)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func TestSuscripcionesDeLaTienda(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	vinos := pruebadb.CrearTienda(t, b.App, "vinos")
	otra := pruebadb.CrearTienda(t, b.App, "otra")
	var ana, malbec string
	if err := b.App.EnTx(ctx, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `INSERT INTO auth.usuarios (email, nombre) VALUES ('ana@vinos.com', 'Ana') RETURNING id::text`).Scan(&ana)
	}); err != nil {
		t.Fatal(err)
	}
	if err := b.App.EnTienda(ctx, vinos, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `INSERT INTO productos (tienda_id, nombre, precio, stock) VALUES ($1, 'Malbec reserva', 20000, 10) RETURNING id::text`, vinos).Scan(&malbec)
	}); err != nil {
		t.Fatal(err)
	}

	f := &mpSuscripciones{preapproval: map[string]map[string]any{}, cobros: map[string]map[string]any{}, pagos: map[string]map[string]any{}}
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
	suscripciones.Rutas(mux, b.App, suscripciones.Config{MP: s, Sitio: "https://tulocalweb.com.ar",
		Tienda: func(r *http.Request) string { return r.Header.Get("x-tienda") }})

	cupo := 2
	plan, err := suscripciones.GuardarPlan(ctx, b.App, vinos, "", suscripciones.Plan{Nombre: "Club Explorador", Precio: 118000,
		Frecuencia: "mensual", Productos: []suscripciones.ProductoPlan{{ProductoID: malbec, Cantidad: 3}}, Incluye: "3 tintos de altura",
		Envio: true, Retiro: true, Cupo: &cupo, Activo: true})
	if err != nil {
		t.Fatal(err)
	}
	// Otra tienda no puede usar un producto ajeno ni ver el plan (RLS).
	if _, err := suscripciones.GuardarPlan(ctx, b.App, otra, "", suscripciones.Plan{Nombre: "Robado", Precio: 1000, Frecuencia: "mensual",
		Productos: []suscripciones.ProductoPlan{{ProductoID: malbec, Cantidad: 1}}, Envio: true, Activo: true}); !errors.Is(err, suscripciones.ErrInvalido) {
		t.Fatalf("usó un producto de otra tienda: %v", err)
	}
	if p, _ := suscripciones.PanelDe(ctx, b.App, otra); len(p.Planes) != 0 {
		t.Fatal("otra tienda ve los planes")
	}

	alta := suscripciones.Alta{PlanID: plan.ID, Email: "juli+club@mail.com", Nombre: "Julia Paz", Telefono: "11 5555 4444",
		Entrega: "envio", Direccion: "Av. Corrientes 1234", CP: "1043"}
	// Sin Mercado Pago conectado, la tienda no ofrece planes ni deja suscribirse.
	if pub, _ := vidriera.Publica(ctx, b.App, "vinos"); len(pub.Suscripciones) != 0 {
		t.Fatal("ofrece planes sin Mercado Pago")
	}
	if _, err := suscripciones.Suscribir(ctx, b.App, s, "https://tulocalweb.com.ar", "vinos", alta); !errors.Is(err, suscripciones.ErrSinCobros) {
		t.Fatalf("se suscribió sin Mercado Pago: %v", err)
	}

	u, err := s.Conectar(ctx, vinos, ana)
	if err != nil {
		t.Fatal(err)
	}
	q, _ := url.Parse(u)
	if err := s.Volver(ctx, vinos, ana, "codigo", q.Query().Get("state")); err != nil {
		t.Fatal(err)
	}
	pub, err := vidriera.Publica(ctx, b.App, "vinos")
	if err != nil || len(pub.Suscripciones) != 1 || pub.Suscripciones[0].Productos[0].Nombre != "Malbec reserva" {
		t.Fatalf("planes públicos: %+v %v", pub.Suscripciones, err)
	}

	// El alta: el precio sale del plan (el navegador no lo manda) y el débito es de la tienda.
	res, err := suscripciones.Suscribir(ctx, b.App, s, "https://tulocalweb.com.ar", "vinos", alta)
	if err != nil || !strings.Contains(res.URL, "preapproval_id=pre-1") || len(res.Codigo) < 16 {
		t.Fatalf("suscribir: %+v %v", res, err)
	}
	rec := f.ultimoAlta["auto_recurring"].(map[string]any)
	if rec["transaction_amount"] != float64(118000) || rec["frequency"] != float64(1) || rec["frequency_type"] != "months" ||
		f.ultimoAlta["payer_email"] != "juli@mail.com" || f.ultimoAlta["back_url"] != "https://tulocalweb.com.ar/t/vinos/suscripcion/"+res.Codigo ||
		!strings.HasPrefix(f.ultimoAlta["external_reference"].(string), "SUS-") {
		t.Fatalf("preapproval: %v", f.ultimoAlta)
	}
	if len(f.pedidos) == 0 || !strings.HasSuffix(f.pedidos[len(f.pedidos)-1], "APP_USR-vinos") {
		t.Fatalf("no usó el token de la tienda: %v", f.pedidos)
	}
	ver, err := suscripciones.Ver(ctx, b.App, "vinos", res.Codigo)
	if err != nil || ver.Estado != "pendiente" || ver.Nombre != "Julia" {
		t.Fatalf("ver: %+v %v", ver, err)
	}
	// Desde otra tienda el código no existe.
	if _, err := suscripciones.Ver(ctx, b.App, "otra", res.Codigo); !errors.Is(err, suscripciones.ErrNoExiste) {
		t.Fatal("el link privado se ve desde otra tienda")
	}

	// El aviso del preapproval (firmado) va al outbox; el worker lo trae de Mercado Pago y la activa.
	f.mu.Lock()
	f.preapproval["pre-1"]["status"] = "authorized"
	f.preapproval["pre-1"]["next_payment_date"] = "2026-12-01T10:00:00.000-03:00"
	f.mu.Unlock()
	avisar := func(tipo, id, cuerpo string) {
		t.Helper()
		r := httptest.NewRequest(http.MethodPost, "/v1/webhooks/mercadopago?data.id="+id+"&type="+tipo, strings.NewReader(cuerpo))
		r.Header.Set("x-request-id", "req-"+id)
		r.Header.Set("x-signature", firmar("whsec", strings.ToLower(id), "req-"+id, ahora.UnixMilli()))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("webhook %s: %d", tipo, w.Code)
		}
	}
	worker := &outbox.Worker{Pool: b.Worker, Manejadores: s.Manejadores()}
	pasar := func() {
		t.Helper()
		if _, err := worker.Pasada(ctx); err != nil {
			t.Fatal(err)
		}
	}
	avisar("subscription_preapproval", "pre-1", `{"type":"subscription_preapproval","data":{"id":"pre-1"}}`)
	pasar()
	if ver, _ = suscripciones.Ver(ctx, b.App, "vinos", res.Codigo); ver.Estado != "activa" || ver.ProximoCobro == nil {
		t.Fatalf("no quedó activa: %+v", ver)
	}
	contar := func(q string, args ...any) int {
		t.Helper()
		var n int
		if err := b.App.EnTienda(ctx, vinos, func(tx pgx.Tx) error { return tx.QueryRow(ctx, q, args...).Scan(&n) }); err != nil {
			t.Fatal(err)
		}
		return n
	}
	if n := contar(`SELECT count(*) FROM outbox WHERE tipo = 'suscripcion.alta'`); n != 1 {
		t.Fatalf("eventos de alta: %d", n)
	}
	// El mismo aviso otra vez no da otra alta.
	avisar("subscription_preapproval", "pre-1", `{"type":"subscription_preapproval","data":{"id":"pre-1"}}`)
	pasar()
	if n := contar(`SELECT count(*) FROM outbox WHERE tipo = 'suscripcion.alta'`); n != 1 {
		t.Fatalf("alta repetida: %d", n)
	}

	// Cada cobro aprobado crea un solo pedido, aunque Mercado Pago repita el aviso (y aunque llegue
	// también como «payment»).
	f.mu.Lock()
	f.cobros["9001"] = map[string]any{"id": 9001, "preapproval_id": "pre-1", "status": "processed", "transaction_amount": 118000,
		"payment": map[string]any{"id": 777001, "status": "approved"}}
	f.pagos["777001"] = map[string]any{"id": 777001, "status": "approved", "external_reference": f.ultimoAlta["external_reference"], "transaction_amount": 118000}
	f.mu.Unlock()
	for range 3 {
		avisar("subscription_authorized_payment", "9001", `{"type":"subscription_authorized_payment","user_id":4242,"data":{"id":"9001"}}`)
	}
	avisar("payment", "777001", `{"type":"payment","user_id":4242,"data":{"id":"777001"}}`)
	pasar()
	if n := contar(`SELECT count(*) FROM pedidos WHERE suscripcion_id IS NOT NULL`); n != 1 {
		t.Fatalf("pedidos de la suscripción: %d", n)
	}
	var canal, pago, estadoPago, estado string
	var total int64
	var stock int
	if err := b.App.EnTienda(ctx, vinos, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `SELECT canal, pago, estado_pago, estado, total FROM pedidos WHERE suscripcion_id IS NOT NULL`).
			Scan(&canal, &pago, &estadoPago, &estado, &total); err != nil {
			return err
		}
		return tx.QueryRow(ctx, `SELECT stock FROM productos WHERE id = $1`, malbec).Scan(&stock)
	}); err != nil {
		t.Fatal(err)
	}
	if canal != "web" || pago != "mp" || estadoPago != "Aprobado" || estado != "Para preparar" || total != 118000 || stock != 7 {
		t.Fatalf("pedido: %s %s %s %s %d, stock %d", canal, pago, estadoPago, estado, total, stock)
	}
	// Entra al circuito normal: «Nueva venta» a la tienda y la confirmación a quien compró.
	if n := contar(`SELECT count(*) FROM outbox WHERE tipo = 'tienda.venta'`); n != 1 {
		t.Fatalf("aviso de venta: %d", n)
	}
	if ver, _ = suscripciones.Ver(ctx, b.App, "vinos", res.Codigo); ver.Cobros != 1 {
		t.Fatalf("cobros: %d", ver.Cobros)
	}

	// Lo cobrado por debajo del precio no arma pedido.
	f.mu.Lock()
	f.cobros["9002"] = map[string]any{"id": 9002, "preapproval_id": "pre-1", "status": "processed", "transaction_amount": 100,
		"payment": map[string]any{"id": 777002, "status": "approved"}}
	// Un cobro rechazado avisa una sola vez.
	f.cobros["9003"] = map[string]any{"id": 9003, "preapproval_id": "pre-1", "status": "recycling", "transaction_amount": 118000,
		"payment": map[string]any{"id": 777003, "status": "rejected"}}
	f.mu.Unlock()
	avisar("subscription_authorized_payment", "9002", `{"type":"subscription_authorized_payment","user_id":4242,"data":{"id":"9002"}}`)
	avisar("subscription_authorized_payment", "9003", `{"type":"subscription_authorized_payment","user_id":4242,"data":{"id":"9003"}}`)
	avisar("subscription_authorized_payment", "9003", `{"type":"subscription_authorized_payment","user_id":4242,"data":{"id":"9003"}}`)
	pasar()
	if n := contar(`SELECT count(*) FROM pedidos WHERE suscripcion_id IS NOT NULL`); n != 1 {
		t.Fatalf("armó pedido por menos plata: %d", n)
	}
	if n := contar(`SELECT count(*) FROM outbox WHERE tipo = 'suscripcion.rechazo'`); n != 1 {
		t.Fatalf("avisos de rechazo: %d", n)
	}

	// Cupo: con 1 vigente y cupo 2 entra una más; la tercera, no.
	alta2 := alta
	alta2.Email = "beto@mail.com"
	if _, err := suscripciones.Suscribir(ctx, b.App, s, "https://tulocalweb.com.ar", "vinos", alta2); err != nil {
		t.Fatal(err)
	}
	alta2.Email = "caro@mail.com"
	if _, err := suscripciones.Suscribir(ctx, b.App, s, "https://tulocalweb.com.ar", "vinos", alta2); !errors.Is(err, suscripciones.ErrCupo) {
		t.Fatalf("pasó el cupo: %v", err)
	}

	// Pausar y cancelar desde el link privado: primero en Mercado Pago.
	if v, err := suscripciones.CambiarPublica(ctx, b.App, s, "vinos", res.Codigo, "pausar"); err != nil || v.Estado != "pausada" {
		t.Fatalf("pausar: %+v %v", v, err)
	}
	if f.preapproval["pre-1"]["status"] != "paused" {
		t.Fatal("no pausó en Mercado Pago")
	}
	if _, err := suscripciones.CambiarPublica(ctx, b.App, s, "vinos", res.Codigo, "pausar"); !errors.Is(err, suscripciones.ErrEstado) {
		t.Fatal("pausó una pausada")
	}
	if v, err := suscripciones.CambiarPublica(ctx, b.App, s, "vinos", res.Codigo, "cancelar"); err != nil || v.Estado != "cancelada" {
		t.Fatalf("cancelar: %+v %v", v, err)
	}
	if f.preapproval["pre-1"]["status"] != "cancelled" {
		t.Fatal("no canceló en Mercado Pago")
	}
	if n := contar(`SELECT count(*) FROM outbox WHERE tipo = 'suscripcion.baja'`); n != 1 {
		t.Fatalf("avisos de baja: %d", n)
	}
	// El aviso de Mercado Pago de la cancelación no da otra baja.
	avisar("subscription_preapproval", "pre-1", `{"type":"subscription_preapproval","data":{"id":"pre-1"}}`)
	pasar()
	if n := contar(`SELECT count(*) FROM outbox WHERE tipo = 'suscripcion.baja'`); n != 1 {
		t.Fatalf("baja repetida: %d", n)
	}

	// La tienda cancela desde el panel (HTTP, con su sesión); otra tienda no puede.
	panel := func(tienda, metodo, ruta string) int {
		r := httptest.NewRequest(metodo, ruta, nil)
		r.Header.Set("x-tienda", tienda)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		return w.Code
	}
	p, _ := suscripciones.PanelDe(ctx, b.App, vinos)
	var beto string
	for _, x := range p.Suscriptores {
		if x.Email == "beto@mail.com" {
			beto = x.ID
		}
	}
	if c := panel(otra, http.MethodPost, "/v1/panel/suscripciones/"+beto+"/cancelar"); c != http.StatusNotFound {
		t.Fatalf("otra tienda canceló: %d", c)
	}
	if c := panel(vinos, http.MethodPost, "/v1/panel/suscripciones/"+beto+"/cancelar"); c != http.StatusOK {
		t.Fatalf("cancelar desde el panel: %d", c)
	}
	// Con suscriptores el plan no se borra (se pausa).
	if c := panel(vinos, http.MethodDelete, "/v1/panel/suscripciones/planes/"+plan.ID); c != http.StatusConflict {
		t.Fatalf("borró un plan con suscriptores: %d", c)
	}
}
