package mercadopago_test

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/lanzamientos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/vidriera"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/mercadopago"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/cifrado"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

// La seña de un pedido reservado (§52): la preferencia por la seña, sin cuotas y con vencimiento;
// el webhook la aplica (el pedido pasa a «Para preparar» con el pago en «Seña») y avisa una vez.
func TestSenaDePedido(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	var ana string
	if err := b.App.EnTx(ctx, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `INSERT INTO auth.usuarios (email, nombre) VALUES ('ana@luna.com', 'Ana') RETURNING id::text`).Scan(&ana)
	}); err != nil {
		t.Fatal(err)
	}
	f := &mpFalso{cuerpos: map[string]map[string]any{}, pagos: map[string]map[string]any{}, susc: map[string]map[string]any{}, usuario: 779, refreshOK: true}
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
	stock := 4
	buzo, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Buzo Heavy", Precio: 50000, Stock: &stock})
	if err != nil {
		t.Fatal(err)
	}
	sale := ahora.Add(time.Hour)
	treinta := 30
	if _, err := lanzamientos.Guardar(ctx, b.App, luna, buzo.ID, lanzamientos.Config{Lanzamiento: &sale, Reserva: &treinta}, ahora); err != nil {
		t.Fatal(err)
	}
	res, err := vidriera.Senar(ctx, b.App, s, s.Config.Sitio, "luna", vidriera.PedidoSena{ProductoID: buzo.ID, Nombre: "Lucía Pérez",
		Email: "lucia@ejemplo.com", Telefono: "1155550101"}, ahora)
	if err != nil || !strings.Contains(res.URL, "pref_id=pref-1") || res.Sena != 15000 {
		t.Fatalf("reservar: %+v %v", res, err)
	}
	pref := f.cuerpos["POST /checkout/preferences"]
	items := pref["items"].([]any)
	if items[0].(map[string]any)["unit_price"].(float64) != 15000 || pref["expires"] != true ||
		pref["payment_methods"].(map[string]any)["installments"].(float64) != 1 ||
		pref["back_urls"].(map[string]any)["success"] != "https://tulocalweb.com.ar/t/luna/pedido/"+res.Codigo+"?pago=ok" {
		t.Fatalf("preferencia de la seña: %v", pref)
	}
	wk := &outbox.Worker{Pool: b.Worker, Manejadores: s.Manejadores()}
	// El pago de la seña la confirma; el aviso repetido no duplica nada.
	webhook := func(id int, req string, monto float64) {
		t.Helper()
		f.mu.Lock()
		f.pagos[strconv.Itoa(id)] = map[string]any{"id": id, "status": "approved", "external_reference": pref["external_reference"], "transaction_amount": monto}
		f.mu.Unlock()
		r := httptest.NewRequest(http.MethodPost, "/v1/webhooks/mercadopago?data.id="+strconv.Itoa(id)+"&type=payment",
			strings.NewReader(`{"type":"payment","action":"payment.created","user_id":779,"data":{"id":"`+strconv.Itoa(id)+`"}}`))
		r.Header.Set("x-request-id", req)
		r.Header.Set("x-signature", firmar("whsec", strconv.Itoa(id), req, ahora.UnixMilli()))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		if w.Code != http.StatusOK {
			t.Fatalf("webhook: %d", w.Code)
		}
		if _, err := wk.Pasada(ctx); err != nil {
			t.Fatal(err)
		}
	}
	webhook(661, "req-2", 15000)
	webhook(661, "req-3", 15000)
	d, _ := pedidos.Obtener(ctx, b.App, luna, res.Numero)
	if !d.Sena.Pagada || d.Sena.Vence != nil || d.Estado != "Para preparar" || d.EstadoPago != "Seña" || d.Sena.Falta != 35000 {
		t.Fatalf("después de la seña: %+v %+v", d, d.Sena)
	}
	admin, err := pgx.Connect(ctx, b.AdminURL)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close(ctx)
	var ventas, senas int
	if err := admin.QueryRow(ctx, `SELECT count(*) FILTER (WHERE tipo = 'tienda.venta'),
			count(*) FILTER (WHERE tipo = 'comprador.aviso' AND datos->>'aviso' = 'sena') FROM outbox`).Scan(&ventas, &senas); err != nil {
		t.Fatal(err)
	}
	if ventas != 1 || senas != 1 {
		t.Fatalf("avisos: %d ventas, %d señas", ventas, senas)
	}
}
