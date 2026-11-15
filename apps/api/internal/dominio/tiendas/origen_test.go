package tiendas_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/tiendas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

func TestOrigenSeFijaUnaSolaVez(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	if _, err := tiendas.FijarOrigen(ctx, b.App, luna, tiendas.Origen{Ref: "<script>"}); err != tiendas.ErrOrigen {
		t.Fatalf("valor raro: %v", err)
	}
	if _, err := tiendas.FijarOrigen(ctx, b.App, luna, tiendas.Origen{Fuente: "facebook", Contenido: "a b"}); err != tiendas.ErrOrigen {
		t.Fatalf("anuncio raro: %v", err)
	}
	// Un origen vacío ({} o null) no se fija: taparía el ref que llegue después.
	if _, err := tiendas.FijarOrigen(ctx, b.App, luna, tiendas.Origen{Llego: "2026-10-23T08:00:00Z"}); err != tiendas.ErrOrigen {
		t.Fatalf("origen vacío: %v", err)
	}
	if ok, err := tiendas.FijarOrigen(ctx, b.App, luna, tiendas.Origen{Fuente: "flyer", Contenido: "1203", Ref: "BRUMA-REF"}); err != nil || !ok {
		t.Fatalf("primera vez: %v %v", ok, err)
	}
	if ok, _ := tiendas.FijarOrigen(ctx, b.App, luna, tiendas.Origen{Ref: "OTRO"}); ok {
		t.Fatal("se pisó el origen")
	}
	var ref, contenido string
	_ = b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT origen->>'ref', origen->>'contenido' FROM tiendas WHERE id = $1`, luna).Scan(&ref, &contenido)
	})
	if ref != "BRUMA-REF" || contenido != "1203" {
		t.Fatalf("ref: %q, contenido: %q", ref, contenido)
	}
}

func TestOrigenDeUnaTiendaNoTocaOtraNiLasViejas(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	sol := pruebadb.CrearTienda(t, b.App, "sol")
	vieja := pruebadb.CrearTienda(t, b.App, "vieja")
	admin, err := pgx.Connect(ctx, b.AdminURL)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close(ctx)
	if _, err := admin.Exec(ctx, `UPDATE tiendas SET creada_en = now() - interval '2 days' WHERE id = $1`, vieja); err != nil {
		t.Fatal(err)
	}

	if ok, err := tiendas.FijarOrigen(ctx, b.App, luna, tiendas.Origen{Ref: "BRUMA-REF"}); err != nil || !ok {
		t.Fatalf("luna: %v %v", ok, err)
	}
	if ok, err := tiendas.FijarOrigen(ctx, b.App, vieja, tiendas.Origen{Ref: "TARDE"}); err != nil || ok {
		t.Fatalf("una tienda de hace dos días no se atribuye: %v %v", ok, err)
	}
	for _, id := range []string{sol, vieja} {
		var nulo bool
		_ = b.App.EnTienda(ctx, id, func(tx pgx.Tx) error {
			return tx.QueryRow(ctx, `SELECT origen IS NULL FROM tiendas WHERE id = $1`, id).Scan(&nulo)
		})
		if !nulo {
			t.Fatalf("la tienda %s quedó con origen", id)
		}
	}
	// Con la sesión de luna no se ve ni se toca sol (RLS).
	var visibles int
	_ = b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT count(*) FROM tiendas WHERE id = $1`, sol).Scan(&visibles)
	})
	if visibles != 0 {
		t.Fatal("luna ve la tienda sol")
	}
}
