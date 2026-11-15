package mitienda_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/mitienda"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

// mostradorctl diseno-v2: con --seco lista y no toca; después migra borrador y publicado (cada uno con
// su plantilla) y una segunda pasada no encuentra nada.
func TestMigrarBase(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	vieja := pruebadb.CrearTienda(t, b.App, "vieja")
	nueva := pruebadb.CrearTienda(t, b.App, "nueva")
	pruebadb.CrearTienda(t, b.App, "virgen") // sin diseño: se arma al leer
	poner := func(id, borrador, publicado string) {
		t.Helper()
		if err := b.App.EnTienda(ctx, id, func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `UPDATE tiendas SET diseno_borrador = $2::jsonb, diseno_publicado = $3::jsonb WHERE id = $1`,
				id, borrador, publicado)
			return err
		}); err != nil {
			t.Fatal(err)
		}
	}
	secciones := `[{"id":"hero","visible":true},{"id":"cats","visible":true},{"id":"prods","visible":true},{"id":"revs","visible":true},{"id":"ig","visible":true},{"id":"wa","visible":false}]`
	poner(vieja, `{"plantilla":"miga","paleta":1,"secciones":`+secciones+`}`, `{"paleta":1,"secciones":`+secciones+`,"estilo":{"anuncio":"Hola"}}`)
	v2, _ := json.Marshal(mitienda.Inicial("litoral"))
	poner(nueva, string(v2), string(v2))

	conn, err := pgx.Connect(ctx, b.MigradorURL)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(ctx)

	cambios, err := mitienda.MigrarBase(ctx, conn, true)
	if err != nil || len(cambios) != 1 || cambios[0] != (mitienda.CambioV2{Slug: "vieja", Borrador: true, Publicado: true}) {
		t.Fatalf("en seco: %+v %v", cambios, err)
	}
	leer := func() (borrador, publicado mitienda.Diseno) {
		t.Helper()
		var a, p []byte
		if err := b.App.EnTienda(ctx, vieja, func(tx pgx.Tx) error {
			return tx.QueryRow(ctx, `SELECT diseno_borrador, diseno_publicado FROM tiendas WHERE id = $1`, vieja).Scan(&a, &p)
		}); err != nil {
			t.Fatal(err)
		}
		_ = json.Unmarshal(a, &borrador)
		_ = json.Unmarshal(p, &publicado)
		return
	}
	if a, _ := leer(); a.Version != 0 {
		t.Fatal("en seco cambió la base")
	}

	if cambios, err = mitienda.MigrarBase(ctx, conn, false); err != nil || len(cambios) != 1 {
		t.Fatalf("migrar: %+v %v", cambios, err)
	}
	a, p := leer()
	if a.Version != 2 || a.Plantilla != "miga" || a.Inicio[0].Variante != "partida" || len(a.Secciones) != 0 || !a.Valido() {
		t.Fatalf("borrador: %+v", a)
	}
	// El publicado, con la plantilla de la tienda (litoral: portada collage) y el anuncio en la barra.
	if p.Version != 2 || p.Inicio[0].Variante != "collage" || p.Estilo == nil || p.Estilo.Barra == nil ||
		p.Estilo.WhatsappFlotante == nil || *p.Estilo.WhatsappFlotante || !p.Valido() {
		t.Fatalf("publicado: %+v", p)
	}
	if cambios, err = mitienda.MigrarBase(ctx, conn, false); err != nil || len(cambios) != 0 {
		t.Fatalf("segunda pasada: %+v %v", cambios, err)
	}
	// La app sigue sin ver las otras tiendas (la política nueva es sólo del migrador).
	var n int
	if err := b.App.EnTienda(ctx, nueva, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `SELECT set_config('app.diseno_v2', 'si', true)`); err != nil {
			return err
		}
		return tx.QueryRow(ctx, `SELECT count(*) FROM tiendas`).Scan(&n)
	}); err != nil || n != 1 {
		t.Fatalf("la app ve %d tiendas (%v)", n, err)
	}
}
