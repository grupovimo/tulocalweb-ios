package vidriera_test

import (
	"context"
	"encoding/json"
	"slices"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/mitienda"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/vidriera"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

// La tienda pública manda el diseño publicado en v2 (un v1 se migra al leer) y, mientras convivan,
// las secciones y el estilo de v1 armados desde el v2.
func TestDisenoV2EnLaTienda(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	id := pruebadb.CrearTienda(t, b.App, "nube")

	// Sin diseño: el inicial del preset (litoral), con los bloques que suma (novedades y ofertas).
	p, err := vidriera.Publica(ctx, b.App, "nube")
	if err != nil || p.Diseno.Version != 2 || len(p.Diseno.Inicio) != len(mitienda.Inicial("litoral").Inicio) ||
		!slices.Equal(p.Secciones, []string{"hero", "cats", "prods", "revs", "ig", "wa"}) || string(p.Estilo) != "{}" {
		t.Fatalf("sin diseño: %d bloques (%d), secciones %v, estilo %s: %v", len(p.Diseno.Inicio), len(mitienda.Inicial("litoral").Inicio), p.Secciones, p.Estilo, err)
	}

	// Un v1 guardado antes: se migra, y lo de v1 se sigue armando igual.
	v1 := `{"secciones":[{"id":"prods","visible":true},{"id":"hero","visible":true},{"id":"cats","visible":false},
		{"id":"revs","visible":true},{"id":"ig","visible":true},{"id":"wa","visible":false}],"paleta":2,
		"estilo":{"anuncio":"Envío gratis","textos":{"prods":"Lo nuevo"}}}`
	if err := b.App.EnTienda(ctx, id, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE tiendas SET diseno_publicado = $2::jsonb WHERE id = $1`, id, v1)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	p, err = vidriera.Publica(ctx, b.App, "nube")
	if err != nil || p.Diseno.Version != 2 || p.Diseno.Inicio[0].Tipo != "productos" || p.Paleta != 2 {
		t.Fatalf("v1: %+v %v", p.Diseno, err)
	}
	if !slices.Equal(p.Secciones, []string{"prods", "hero", "revs", "ig"}) {
		t.Fatalf("secciones de v1: %v", p.Secciones)
	}
	var e struct {
		Anuncio string            `json:"anuncio"`
		Textos  map[string]string `json:"textos"`
		Barra   *mitienda.Barra   `json:"barra"`
	}
	if err := json.Unmarshal(p.Estilo, &e); err != nil || e.Anuncio != "Envío gratis" || e.Textos["prods"] != "Lo nuevo" || e.Barra == nil {
		t.Fatalf("estilo de v1: %s %v", p.Estilo, err)
	}

	// Un v2 con un producto destacado: el producto viene en la respuesta.
	prod, err := listas.Guardar(ctx, b.App, id, "", listas.Editable{Nombre: "Mate imperial", Precio: 25000})
	if err != nil {
		t.Fatal(err)
	}
	d := mitienda.Inicial("litoral")
	d.Inicio = append(d.Inicio, mitienda.Bloque{ID: "destacado1", Tipo: "destacado", Variante: "partido",
		Datos: map[string]any{"producto": prod.ID, "texto": "El de siempre"}})
	if err := mitienda.Guardar(ctx, b.App, id, d); err != nil {
		t.Fatal(err)
	}
	if err := mitienda.Publicar(ctx, b.App, id); err != nil {
		t.Fatal(err)
	}
	p, err = vidriera.Publica(ctx, b.App, "nube")
	if err != nil {
		t.Fatal(err)
	}
	ult := p.Diseno.Inicio[len(p.Diseno.Inicio)-1]
	if ult.Tipo != "destacado" || ult.Datos["producto"] != prod.ID || !slices.ContainsFunc(p.Productos,
		func(x vidriera.ProductoPublico) bool { return x.ID == prod.ID }) {
		t.Fatalf("destacado: %+v (%d productos)", ult, len(p.Productos))
	}
	if !slices.Equal(p.Diseno.ProductosReferidos(), []string{prod.ID}) {
		t.Fatalf("referidos: %v", p.Diseno.ProductosReferidos())
	}
}
