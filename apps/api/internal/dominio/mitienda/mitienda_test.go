package mitienda_test

import (
	"context"
	"strings"
	"testing"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/mitienda"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/vidriera"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

func TestBorradorYPublicado(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	id := pruebadb.CrearTienda(t, b.App, "luna")
	v, err := mitienda.Leer(ctx, b.App, id, "https://tulocalweb.com.ar")
	if err != nil || v["sinPublicar"] != false || v["direccion"] != "luna.tulocalweb.com.ar" {
		t.Fatalf("inicial: %+v %v", v, err)
	}
	d := mitienda.Inicial()
	d.Secciones[0], d.Secciones[1] = d.Secciones[1], d.Secciones[0]
	d.Secciones[5].Visible = false
	d.Paleta = 3
	if err := mitienda.Guardar(ctx, b.App, id, d); err != nil {
		t.Fatal(err)
	}
	if v, _ = mitienda.Leer(ctx, b.App, id, ""); v["sinPublicar"] != true {
		t.Fatal("no quedó sin publicar")
	}
	if err := mitienda.Publicar(ctx, b.App, id); err != nil {
		t.Fatal(err)
	}
	if v, _ = mitienda.Leer(ctx, b.App, id, ""); v["sinPublicar"] != false {
		t.Fatal("no se publicó")
	}
	malo := mitienda.Inicial()
	malo.Secciones = malo.Secciones[:3]
	if err := mitienda.Guardar(ctx, b.App, id, malo); err == nil {
		t.Fatal("aceptó un diseño sin todas las secciones")
	}
}

// Plantilla, textos, portada y logo van al borrador y se ven en la tienda recién al publicar.
func TestPlantillaTextosYFotosAlPublicar(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	id := pruebadb.CrearTienda(t, b.App, "sol")
	antes, err := vidriera.Publica(ctx, b.App, "sol")
	if err != nil {
		t.Fatal(err)
	}
	otra := "ceibo"
	if antes.Plantilla == otra {
		otra = "yerbal"
	}
	d := mitienda.Inicial()
	d.Plantilla, d.Titulo, d.Lema = otra, "  Pan de masa madre ", "Horneado cada mañana"
	d.Portada, d.Logo = "https://tulocalweb.com.ar/f/fotos/a.webp", "https://tulocalweb.com.ar/f/fotos/b.webp"
	if err := mitienda.Guardar(ctx, b.App, id, d); err != nil {
		t.Fatal(err)
	}
	v, _ := mitienda.Leer(ctx, b.App, id, "")
	if v["sinPublicar"] != true || v["plantillaId"] != otra {
		t.Fatalf("borrador: %+v", v)
	}
	if p, _ := vidriera.Publica(ctx, b.App, "sol"); p.Plantilla != antes.Plantilla || p.Logo != "" || p.Portada["titulo"] != antes.Nombre {
		t.Fatalf("la tienda cambió antes de publicar: %+v", p)
	}
	if err := mitienda.Publicar(ctx, b.App, id); err != nil {
		t.Fatal(err)
	}
	p, err := vidriera.Publica(ctx, b.App, "sol")
	if err != nil {
		t.Fatal(err)
	}
	if p.Plantilla != otra || p.Logo != d.Logo || p.Portada["titulo"] != "Pan de masa madre" ||
		p.Portada["lema"] != "Horneado cada mañana" || *(p.Portada["foto"].(*string)) != d.Portada {
		t.Fatalf("publicada: %s %s %+v", p.Plantilla, p.Logo, p.Portada)
	}
	if v, _ = mitienda.Leer(ctx, b.App, id, ""); v["sinPublicar"] != false || v["plantillaId"] != otra {
		t.Fatalf("después de publicar: %+v", v)
	}
	for _, malo := range []func(*mitienda.Diseno){
		func(d *mitienda.Diseno) { d.Plantilla = "inventada" },
		func(d *mitienda.Diseno) { d.Logo = "javascript:alert(1)" },
		func(d *mitienda.Diseno) { d.Portada = "https://x.com/a b.webp" },
		func(d *mitienda.Diseno) { d.Titulo = strings.Repeat("a", 81) },
	} {
		x := mitienda.Inicial()
		malo(&x)
		if err := mitienda.Guardar(ctx, b.App, id, x); err == nil {
			t.Fatalf("aceptó %+v", x)
		}
	}
}

// Al compartir: título, descripción y foto propia van al borrador y a la tienda al publicar.
func TestAlCompartir(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	id := pruebadb.CrearTienda(t, b.App, "brisa")
	d := mitienda.Inicial()
	d.CompartirTitulo, d.CompartirDescripcion = " Brisa · Velas de soja ", "Hechas a mano en Mendoza."
	d.CompartirFoto = "https://tulocalweb.com.ar/f/fotos/c.webp"
	if err := mitienda.Guardar(ctx, b.App, id, d); err != nil {
		t.Fatal(err)
	}
	if p, _ := vidriera.Publica(ctx, b.App, "brisa"); p.Compartir != (vidriera.Compartir{}) {
		t.Fatalf("se vio antes de publicar: %+v", p.Compartir)
	}
	if err := mitienda.Publicar(ctx, b.App, id); err != nil {
		t.Fatal(err)
	}
	p, err := vidriera.Publica(ctx, b.App, "brisa")
	if err != nil {
		t.Fatal(err)
	}
	if p.Compartir != (vidriera.Compartir{Titulo: "Brisa · Velas de soja", Descripcion: d.CompartirDescripcion, Foto: d.CompartirFoto}) {
		t.Fatalf("publicada: %+v", p.Compartir)
	}
	for _, malo := range []func(*mitienda.Diseno){
		func(x *mitienda.Diseno) { x.CompartirTitulo = strings.Repeat("a", 71) },
		func(x *mitienda.Diseno) { x.CompartirDescripcion = strings.Repeat("a", 161) },
		func(x *mitienda.Diseno) { x.CompartirFoto = "javascript:alert(1)" },
	} {
		x := mitienda.Inicial()
		malo(&x)
		if err := mitienda.Guardar(ctx, b.App, id, x); err == nil {
			t.Fatalf("aceptó %+v", x)
		}
	}
}
