package mercadopago_test

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/configuracion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cupones"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/vidriera"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/mercadopago"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/cifrado"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

// Un pedido de la tienda online con Mercado Pago que no se pagó no es una venta (00085): ni
// Telegram, ni mail o push de «Nueva venta», ni la confirmación a quien compra, ni el «Pedido nuevo»
// del panel, ni lo vendido del cupón. Al acreditarse sale todo, una sola vez. Con transferencia,
// como siempre: al crearse. A las 24 h sin pagar se cancela solo y vuelven el stock y el cupón.
func TestMercadoPagoSinPagarNoEsVenta(t *testing.T) {
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
	if err := configuracion.GuardarNegocio(ctx, b.App, luna, configuracion.Negocio{Nombre: "Luna", Alias: "luna.mates"}); err != nil {
		t.Fatal(err)
	}
	if err := configuracion.GuardarEntregas(ctx, b.App, luna, configuracion.Entregas{Retiro: true, RetiroDireccion: "Gorriti 4500"}); err != nil {
		t.Fatal(err)
	}
	// Un cupón de un solo uso: lo reserva el pedido que espera el pago.
	if _, err := cupones.Crear(ctx, b.App, luna, cupones.Nuevo{Codigo: "UNICO10", Tipo: "pct", Valor: 10, Alcance: "todo",
		Quien: "todos", Limite: 1, PorCliente: "1", Vigencia: "sin", Canales: []string{"web"}}, ahora); err != nil {
		t.Fatal(err)
	}
	stock := 5
	mate, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Mate", Precio: 10000, Stock: &stock})
	if err != nil {
		t.Fatal(err)
	}

	admin, err := pgx.Connect(ctx, b.AdminURL)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close(ctx)
	// La tienda tiene Telegram vinculado (si no, el trigger no encola nada y no habría qué mirar).
	if _, err := admin.Exec(ctx, `INSERT INTO telegram_vinculos (tienda_id, chat, nombre) VALUES ($1, 4242, 'Ana')`, luna); err != nil {
		t.Fatal(err)
	}
	// El tiempo real sale por NOTIFY: se escucha desde otra conexión.
	escucha, err := pgx.Connect(ctx, b.AdminURL)
	if err != nil {
		t.Fatal(err)
	}
	defer escucha.Close(ctx)
	if _, err := escucha.Exec(ctx, `LISTEN tiempo_real`); err != nil {
		t.Fatal(err)
	}
	// avisosTR: los tipos que llegaron para ese pedido desde la última vez.
	avisosTR := func(numero int) map[string]int {
		t.Helper()
		out := map[string]int{}
		for {
			c, cancelar := context.WithTimeout(ctx, 400*time.Millisecond)
			n, err := escucha.WaitForNotification(c)
			cancelar()
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) || c.Err() != nil {
					return out
				}
				t.Fatal(err)
			}
			var a struct {
				K  string `json:"k"`
				ID string `json:"id"`
			}
			_ = json.Unmarshal([]byte(n.Payload), &a)
			if a.ID == strconv.Itoa(numero) {
				out[a.K]++
			}
		}
	}
	// cola: cuántos eventos de aviso hay en el outbox para el pedido.
	cola := func(numero int) map[string]int {
		t.Helper()
		out := map[string]int{}
		filas, err := admin.Query(ctx, `SELECT tipo || coalesce(':' || (datos->>'aviso'), '') FROM outbox
			WHERE tipo IN ('telegram.venta', 'tienda.venta', 'push.venta', 'comprador.aviso') AND (datos->>'numero')::int = $1`, numero)
		if err != nil {
			t.Fatal(err)
		}
		defer filas.Close()
		for filas.Next() {
			var k string
			if err := filas.Scan(&k); err != nil {
				t.Fatal(err)
			}
			out[k]++
		}
		return out
	}
	cupon := func() (usos int, vendio, desconto int64) {
		t.Helper()
		if err := admin.QueryRow(ctx, `SELECT usos, vendio, desconto FROM cupones WHERE codigo = 'UNICO10'`).Scan(&usos, &vendio, &desconto); err != nil {
			t.Fatal(err)
		}
		return
	}
	stockDe := func() int {
		t.Helper()
		var n int
		if err := admin.QueryRow(ctx, `SELECT stock FROM productos WHERE id = $1`, mate.ID).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	compra := func(pago, cupon string) vidriera.Resultado {
		t.Helper()
		c := vidriera.Compra{Email: "lucia@ejemplo.com", Nombre: "Lucía Pérez", Telefono: "11 5555-0101", CP: "1414",
			Envio: "retiro", Pago: pago, Cupon: cupon}
		c.Carrito = append(c.Carrito, vidriera.LineaCarrito{ID: mate.ID, Cantidad: 1})
		r, err := vidriera.Comprar(ctx, b.App, s, s.Config.Sitio, "luna", c, ahora)
		if err != nil {
			t.Fatal(err)
		}
		return r
	}

	// 1. Mercado Pago, sin pagar: nada de venta. La lista se refresca, pero no hay «Pedido nuevo».
	r := compra("mercadopago", "UNICO10")
	if c := cola(r.Numero); len(c) != 0 {
		t.Fatalf("sin pagar ya avisó: %v", c)
	}
	if tr := avisosTR(r.Numero); tr["pedido.nuevo"] != 0 || tr["pedido.cambio"] == 0 {
		t.Fatalf("tiempo real sin pagar: %v", tr)
	}
	if usos, vendio, desconto := cupon(); usos != 1 || vendio != 0 || desconto != 0 {
		t.Fatalf("cupón reservado: usos %d vendió %d descontó %d", usos, vendio, desconto)
	}
	// La preferencia vence con el pedido (24 h), así nadie paga uno ya cancelado.
	if pref := f.cuerpos["POST /checkout/preferences"]; pref["expires"] != true || pref["expiration_date_to"] == nil {
		t.Fatalf("la preferencia no vence: %v", pref)
	}
	d, err := pedidos.Obtener(ctx, b.App, luna, r.Numero)
	if err != nil || d.Estado != "Sin pagar" || d.Comprador.Pedidos != 1 {
		t.Fatalf("detalle esperando el pago: %+v %v", d, err)
	}

	// 2. Se acredita (el aviso llega dos veces): todo una sola vez, con la confirmación (no «pago»).
	pref := f.cuerpos["POST /checkout/preferences"]
	wk := &outbox.Worker{Pool: b.Worker, Manejadores: s.Manejadores()}
	webhook := func(id int, req string) {
		t.Helper()
		f.mu.Lock()
		f.pagos[strconv.Itoa(id)] = map[string]any{"id": id, "status": "approved", "external_reference": pref["external_reference"], "transaction_amount": float64(d.Total)}
		f.mu.Unlock()
		rq := httptest.NewRequest(http.MethodPost, "/v1/webhooks/mercadopago?data.id="+strconv.Itoa(id)+"&type=payment",
			strings.NewReader(`{"type":"payment","action":"payment.created","user_id":779,"data":{"id":"`+strconv.Itoa(id)+`"}}`))
		rq.Header.Set("x-request-id", req)
		rq.Header.Set("x-signature", firmar("whsec", strconv.Itoa(id), req, ahora.UnixMilli()))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, rq)
		if w.Code != http.StatusOK {
			t.Fatalf("webhook: %d", w.Code)
		}
		if _, err := wk.Pasada(ctx); err != nil {
			t.Fatal(err)
		}
	}
	webhook(771, "req-1")
	webhook(771, "req-2")
	want := map[string]int{"telegram.venta": 1, "tienda.venta": 1, "push.venta": 1, "comprador.aviso:confirmacion": 1}
	if c := cola(r.Numero); len(c) != len(want) || c["telegram.venta"] != 1 || c["tienda.venta"] != 1 || c["push.venta"] != 1 ||
		c["comprador.aviso:confirmacion"] != 1 {
		t.Fatalf("al acreditarse: %v (quiero %v)", c, want)
	}
	if tr := avisosTR(r.Numero); tr["pedido.nuevo"] != 1 || tr["pago.acreditado"] != 1 {
		t.Fatalf("tiempo real al acreditarse: %v", tr)
	}
	if usos, vendio, desconto := cupon(); usos != 1 || vendio != d.Total || desconto != 1000 {
		t.Fatalf("cupón al confirmarse: usos %d vendió %d descontó %d (total %d)", usos, vendio, desconto, d.Total)
	}

	// 3. Transferencia: como siempre, avisa al crearse (falta el comprobante).
	tr := compra("transferencia", "")
	if c := cola(tr.Numero); c["telegram.venta"] != 1 || c["tienda.venta"] != 1 || c["push.venta"] != 1 || c["comprador.aviso:confirmacion"] != 1 {
		t.Fatalf("transferencia: %v", c)
	}
	if e := avisosTR(tr.Numero); e["pedido.nuevo"] != 1 {
		t.Fatalf("tiempo real de la transferencia: %v", e)
	}

	// 4. Mercado Pago que nunca se paga: a las 24 h se cancela solo, sin avisar, y vuelve el stock.
	antes := stockDe()
	v := compra("mercadopago", "")
	if stockDe() != antes-1 {
		t.Fatal("el pedido sin pagar no reservó el stock")
	}
	if n, err := pedidos.VencerSinPago(ctx, b.App); err != nil || n != 0 {
		t.Fatalf("venció antes de tiempo: %d %v", n, err)
	}
	if _, err := admin.Exec(ctx, `UPDATE pedidos SET creado_en = now() - interval '25 hours' WHERE numero = $1`, v.Numero); err != nil {
		t.Fatal(err)
	}
	avisosTR(v.Numero)
	if n, err := pedidos.VencerSinPago(ctx, b.App); err != nil || n != 1 {
		t.Fatalf("vencer: %d %v", n, err)
	}
	if n, _ := pedidos.VencerSinPago(ctx, b.App); n != 0 {
		t.Fatal("venció dos veces")
	}
	dv, _ := pedidos.Obtener(ctx, b.App, luna, v.Numero)
	if dv.Estado != "Cancelado" || stockDe() != antes {
		t.Fatalf("vencido: %s, stock %d (antes %d)", dv.Estado, stockDe(), antes)
	}
	if c := cola(v.Numero); len(c) != 0 {
		t.Fatalf("el vencido avisó: %v", c)
	}
	if e := avisosTR(v.Numero); e["pedido.nuevo"] != 0 {
		t.Fatalf("el vencido se anunció: %v", e)
	}
}

// El cupón único que reservó un pedido sin pagar vuelve a estar disponible al vencer.
func TestMercadoPagoSinPagarLiberaElCupon(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	ahora := time.Now()
	if _, err := cupones.Crear(ctx, b.App, luna, cupones.Nuevo{Codigo: "UNICO10", Tipo: "pct", Valor: 10, Alcance: "todo",
		Quien: "todos", Limite: 1, PorCliente: "1", Vigencia: "sin", Canales: []string{"web"}}, ahora); err != nil {
		t.Fatal(err)
	}
	stock := 5
	mate, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Mate", Precio: 10000, Stock: &stock})
	if err != nil {
		t.Fatal(err)
	}
	cliente, f := "Lucía", false
	nuevo := pedidos.PedidoNuevo{Canal: "web", Pago: "mp", Envio: "retiro", Cliente: &cliente, Email: "lucia@ejemplo.com",
		Pagado: &f, Cupon: "UNICO10", DescuentoCupon: 1000, Descuento: 1000, CodigoPublico: "codigo-publico-de-prueba-1",
		Items: []pedidos.Linea{{ProductoID: mate.ID, Cantidad: 1}}}
	n, _, err := pedidos.CrearDesdeProductos(ctx, b.App, luna, nuevo)
	if err != nil {
		t.Fatal(err)
	}
	// Otro no lo puede usar mientras espera el pago.
	otro := nuevo
	otro.Email, otro.CodigoPublico = "beto@ejemplo.com", "codigo-publico-de-prueba-2"
	if _, _, err := pedidos.CrearDesdeProductos(ctx, b.App, luna, otro); !errors.Is(err, pedidos.ErrCupon) {
		t.Fatalf("el cupón reservado se usó dos veces: %v", err)
	}
	admin, err := pgx.Connect(ctx, b.AdminURL)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close(ctx)
	if _, err := admin.Exec(ctx, `UPDATE pedidos SET creado_en = now() - interval '25 hours' WHERE numero = $1`, n); err != nil {
		t.Fatal(err)
	}
	if v, err := pedidos.VencerSinPago(ctx, b.App); err != nil || v != 1 {
		t.Fatalf("vencer: %d %v", v, err)
	}
	if _, _, err := pedidos.CrearDesdeProductos(ctx, b.App, luna, otro); err != nil {
		t.Fatalf("el cupón no se liberó: %v", err)
	}
}
