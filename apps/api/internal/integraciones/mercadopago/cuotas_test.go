package mercadopago_test

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cobros"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/configuracion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/vidriera"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/mercadopago"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/cifrado"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

// Cuotas sin interés: la preferencia del pedido limita las cuotas al máximo que eligió la tienda en
// Configuración › Cobros; sin elegir, no pone límite propio (el de la cuenta de Mercado Pago).
func TestCuotasDelPedido(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	var ana string
	if err := b.App.EnTx(ctx, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `INSERT INTO auth.usuarios (email, nombre) VALUES ('ana@luna.com', 'Ana') RETURNING id::text`).Scan(&ana)
	}); err != nil {
		t.Fatal(err)
	}
	f := &mpFalso{cuerpos: map[string]map[string]any{}, pagos: map[string]map[string]any{}, susc: map[string]map[string]any{}, usuario: 780, refreshOK: true}
	srv := httptest.NewServer(f)
	defer srv.Close()
	k := make([]byte, 32)
	_, _ = rand.Read(k)
	cif, _ := cifrado.Nuevo(base64.StdEncoding.EncodeToString(k))
	ahora := time.Now()
	s := &mercadopago.Servicio{Pool: b.App, Cifra: cif, Cliente: &mercadopago.Cliente{Base: srv.URL},
		Ahora:  func() time.Time { return ahora },
		Config: mercadopago.Config{AppID: "123", Secreto: "sec", SecretoWebhook: "whsec", Sitio: "https://tulocalweb.com.ar"}}
	u, err := s.Conectar(ctx, luna, ana)
	if err != nil {
		t.Fatal(err)
	}
	q, _ := url.Parse(u)
	if err := s.Volver(ctx, luna, ana, "codigo", q.Query().Get("state")); err != nil {
		t.Fatal(err)
	}
	stock := 10
	mate, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Mate", Precio: 40000, Stock: &stock})
	if err != nil {
		t.Fatal(err)
	}
	comprar := func() map[string]any {
		t.Helper()
		c := vidriera.Compra{Email: "lucia@ejemplo.com", Nombre: "Lucía Pérez", Telefono: "11 5555-0101", CP: "1414",
			Envio: "retiro", Pago: "mercadopago", Carrito: []vidriera.LineaCarrito{{ID: mate.ID, Cantidad: 1}}}
		if _, err := vidriera.Comprar(ctx, b.App, s, s.Config.Sitio, "luna", c, ahora); err != nil {
			t.Fatal(err)
		}
		return f.cuerpos["POST /checkout/preferences"]
	}
	if pm, hay := comprar()["payment_methods"]; hay {
		t.Fatalf("sin cuotas elegidas no se limita nada: %v", pm)
	}
	if err := configuracion.GuardarCobros(ctx, b.App, luna, cobros.Config{Cuotas: 6, CuotasDesde: 30000}); err != nil {
		t.Fatal(err)
	}
	if n := comprar()["payment_methods"].(map[string]any)["installments"]; n != float64(6) {
		t.Fatalf("installments: %v", n)
	}
}
