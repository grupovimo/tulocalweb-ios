package vidriera_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/configuracion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/vidriera"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

// El newsletter desde la tienda online (§50): el pie, con su tope por IP, y la casilla del checkout.
func TestNewsletterDesdeLaTienda(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	if err := configuracion.GuardarNegocio(ctx, b.App, luna, configuracion.Negocio{Nombre: "Luna", Alias: "luna.tienda"}); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	vidriera.Rutas(mux, b.App, nil, "https://tulocalweb.com.ar", nil, func(*http.Request) string { return "200.1.2.3" }, nil)
	post := func(ruta, cuerpo string) int {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(http.MethodPost, ruta, strings.NewReader(cuerpo)))
		return w.Code
	}
	if c := post("/v1/publico/tiendas/luna/newsletter", `{"email":"no"}`); c != http.StatusBadRequest {
		t.Fatalf("email inválido: %d", c)
	}
	if c := post("/v1/publico/tiendas/no-existe/newsletter", `{"email":"ana@ejemplo.com"}`); c != http.StatusNotFound {
		t.Fatalf("tienda que no existe: %d", c)
	}
	if c := post("/v1/publico/tiendas/luna/newsletter", `{"email":"ana@ejemplo.com"}`); c != http.StatusAccepted {
		t.Fatalf("suscribir: %d", c)
	}
	// Diez por hora por IP y tienda.
	for i := range 9 {
		post("/v1/publico/tiendas/luna/newsletter", `{"email":"x`+string(rune('a'+i))+`@ejemplo.com"}`)
	}
	if c := post("/v1/publico/tiendas/luna/newsletter", `{"email":"otra@ejemplo.com"}`); c != http.StatusTooManyRequests {
		t.Fatalf("tope por IP: %d", c)
	}
	var token string
	_ = b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT token FROM suscriptores WHERE email = 'ana@ejemplo.com'`).Scan(&token)
	})
	if c := post("/v1/publico/tiendas/luna/newsletter/confirmar", `{"token":"`+token+`"}`); c != http.StatusNoContent {
		t.Fatalf("confirmar: %d", c)
	}
	if c := post("/v1/publico/tiendas/luna/newsletter/baja", `{"token":"`+token+`"}`); c != http.StatusNoContent {
		t.Fatalf("baja: %d", c)
	}
	if c := post("/v1/publico/tiendas/luna/newsletter/baja", `{"token":"`+token+`"}`); c != http.StatusNotFound {
		t.Fatalf("baja repetida: %d", c)
	}

	// La casilla del checkout: queda pendiente, con origen checkout.
	stock := 3
	p, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Mate", Precio: 10000, Stock: &stock})
	if err != nil {
		t.Fatal(err)
	}
	c := compra(p.ID, 1, "transferencia")
	c.Email, c.Novedades = "lucia@ejemplo.com", true
	if _, err := vidriera.Comprar(ctx, b.App, nil, "", "luna", c, time.Now()); err != nil {
		t.Fatal(err)
	}
	var origen string
	var confirmado bool
	if err := b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT origen, confirmado_en IS NOT NULL FROM suscriptores WHERE email = 'lucia@ejemplo.com'`).Scan(&origen, &confirmado)
	}); err != nil || origen != "checkout" || confirmado {
		t.Fatalf("checkout: %q %v %v", origen, confirmado, err)
	}
}
