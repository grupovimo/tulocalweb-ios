package demo_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/demo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/mitienda"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/tiendas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db"
)

// El SQL sale siempre igual para la misma fecha (las demos se pueden regenerar sin sorpresas).
func TestSQLEstable(t *testing.T) {
	ahora := time.Date(2026, 11, 5, 18, 0, 0, 0, time.UTC)
	a, err := demo.SQL("x@example.com", ahora)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := demo.SQL("x@example.com", ahora)
	if a != b || len(a) < 100_000 {
		t.Fatalf("el SQL no es estable o salió vacío (%d bytes)", len(a))
	}
}

// Con DEMO_BASE (una base con las demos cargadas, con un usuario del rol de la app: la RLS tiene que
// actuar) y DEMO_ADMIN (para encontrar las tiendas por su dirección), las lee con las mismas
// funciones que el panel: si una forma de JSON no es la que espera la app, falla acá y no en una
// demo en vivo.
func TestDemosSeLeenComoElPanel(t *testing.T) {
	url, adminURL := os.Getenv("DEMO_BASE"), os.Getenv("DEMO_ADMIN")
	if url == "" || adminURL == "" {
		t.Skip("sin DEMO_BASE y DEMO_ADMIN")
	}
	ctx := context.Background()
	pool, err := db.Abrir(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	admin, err := db.Abrir(ctx, adminURL)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	ahora := time.Now()
	for _, d := range demo.Tiendas {
		var id string
		if err := admin.QueryRow(ctx, `SELECT id FROM tiendas WHERE slug = $1`, d.Slug).Scan(&id); err != nil {
			t.Fatalf("%s: %v", d.Slug, err)
		}
		tab, err := tiendas.ObtenerTablero(ctx, pool, id, "30d", ahora)
		if err != nil {
			t.Fatalf("%s tablero: %v", d.Slug, err)
		}
		l, err := pedidos.Listar(ctx, pool, id, pedidos.Filtros{Cuantos: 50}, ahora)
		if err != nil || l.Total == 0 {
			t.Fatalf("%s pedidos: %v (%d)", d.Slug, err, l.Total)
		}
		if _, err := pedidos.Obtener(ctx, pool, id, l.Filas[0].Numero); err != nil {
			t.Fatalf("%s pedido %d: %v", d.Slug, l.Filas[0].Numero, err)
		}
		ps, err := listas.Productos(ctx, pool, id, ahora)
		if err != nil || len(ps) == 0 {
			t.Fatalf("%s productos: %v", d.Slug, err)
		}
		cs, err := listas.Clientes(ctx, pool, id, ahora)
		if err != nil || len(cs) == 0 {
			t.Fatalf("%s clientes: %v", d.Slug, err)
		}
		if _, err := mitienda.Leer(ctx, pool, id, "publicado"); err != nil {
			t.Fatalf("%s diseño: %v", d.Slug, err)
		}
		t.Logf("%s: tablero %+v · %d pedidos · %d productos · %d clientes", d.Slug, tab, l.Total, len(ps), len(cs))
	}
}
