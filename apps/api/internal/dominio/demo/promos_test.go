package demo_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/demo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/promos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

// Las demos cargadas en una base de prueba estrenan sus promociones: la prueba social de Almacén
// Verde con pedidos pagados de la semana y los números de las ventanas en el panel.
func TestPromocionesDeLasDemos(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	ahora := time.Now()
	s, err := demo.SQL("nadie@example.com", ahora)
	if err != nil {
		t.Fatal(err)
	}
	admin, err := pgx.Connect(ctx, b.AdminURL)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close(ctx)
	if _, err := admin.Exec(ctx, s); err != nil {
		t.Fatal(err)
	}
	compras, err := promos.Recientes(ctx, b.App, "almacenverde", ahora)
	if err != nil || len(compras) == 0 {
		t.Fatalf("recientes de Almacén Verde: %v (%d)", err, len(compras))
	}
	if compras[0].Nombre == "" || compras[0].Provincia != "Córdoba" {
		t.Fatalf("compra reciente: %+v", compras[0])
	}
	if c, _ := promos.Recientes(ctx, b.App, "lumbre", ahora); len(c) != 0 {
		t.Fatal("Lumbre no prendió la prueba social")
	}
	for slug, contenido := range map[string]string{"lumbre": "captacion", "cafebruma": "ruleta", "almacenverde": "promo"} {
		var id string
		if err := admin.QueryRow(ctx, `SELECT id FROM tiendas WHERE slug = $1`, slug).Scan(&id); err != nil {
			t.Fatal(err)
		}
		desde, hasta, _ := promos.Rango("", "", ahora)
		m, err := promos.LeerMetricas(ctx, b.App, id, desde, hasta)
		if err != nil || len(m.Ventanas) != 1 || m.Ventanas[0].Contenido != contenido || m.Ventanas[0].Vistas == 0 {
			t.Fatalf("%s: métricas %+v %v", slug, m, err)
		}
		if contenido != "promo" && m.Ventanas[0].Envios == 0 {
			t.Fatalf("%s: sin contactos captados", slug)
		}
	}
	// La vitrina de demostración muestra el premio sin guardar nada.
	r, err := promos.Participar(ctx, b.App, "lumbre", "captacion",
		promos.Pedido{Ventana: "bienvenida", Email: "visita@example.com", Consentimiento: true}, ahora)
	if err != nil || !r.Demo || r.Premio.Codigo == "" {
		t.Fatalf("vitrina: %+v %v", r, err)
	}
}
