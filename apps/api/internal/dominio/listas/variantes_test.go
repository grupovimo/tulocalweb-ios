package listas_test

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

func ptr[T any](v T) *T { return &v }

func variante(p listas.Producto, valores ...string) listas.Variante {
	for _, v := range p.Variantes {
		ok := true
		for i := 0; i < len(valores); i += 2 {
			ok = ok && v.Valores[valores[i]] == valores[i+1]
		}
		if ok {
			return v
		}
	}
	return listas.Variante{}
}

func TestVariantesAlGuardar(t *testing.T) {
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	talles := listas.Opcion{Nombre: "Talle", Valores: []string{"S", "M"}}
	color := listas.Opcion{Nombre: "Color", Valores: []string{"Negro"}}

	// Al guardar con opciones se generan las combinaciones, con stock 0 y activas; el stock que
	// manda el editor no cuenta (es la suma de las variantes).
	p, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Remera lisa", Precio: 20000, Stock: ptr(50),
		Fotos: []string{"https://f/a.jpg", "https://f/b.jpg"}, Opciones: []listas.Opcion{talles, color}})
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Variantes) != 2 || p.Stock != 0 || *p.Foto != "https://f/a.jpg" || len(p.Fotos) != 2 {
		t.Fatalf("alta: %+v", p)
	}
	if p.Variantes[0].Valores["Talle"] != "S" || !p.Variantes[0].Activa || p.Variantes[0].Precio != nil {
		t.Fatalf("orden o valores por defecto: %+v", p.Variantes)
	}

	// Stock, precio, SKU y foto por variante; el producto suma las activas (trigger).
	p, err = listas.Guardar(ctx, b.App, luna, p.ID, listas.Editable{Nombre: "Remera lisa", Precio: 20000, Variantes: []listas.VarianteEditable{
		{Valores: map[string]string{"Talle": "M", "Color": "Negro"}, Stock: ptr(4), Precio: ptr(int64(21000)), SKU: ptr("REM-M"), Foto: ptr("https://f/b.jpg")},
		{Valores: map[string]string{"Talle": "S", "Color": "Negro"}, Stock: ptr(2)},
	}, Fotos: []string{"https://f/a.jpg", "https://f/b.jpg"}})
	if err != nil {
		t.Fatal(err)
	}
	if p.Stock != 6 {
		t.Fatalf("stock total: %d", p.Stock)
	}

	// Cambiar opciones: M sigue con lo suyo, L entra en 0, S se va.
	talles.Valores = []string{"M", "L"}
	p, err = listas.Guardar(ctx, b.App, luna, p.ID, listas.Editable{Nombre: "Remera lisa", Precio: 20000,
		Fotos: []string{"https://f/a.jpg", "https://f/b.jpg"}, Opciones: []listas.Opcion{talles, color}})
	if err != nil {
		t.Fatal(err)
	}
	m, l := variante(p, "Talle", "M"), variante(p, "Talle", "L")
	if len(p.Variantes) != 2 || m.Stock != 4 || *m.Precio != 21000 || m.SKU != "REM-M" || *m.Foto != "https://f/b.jpg" || l.ID == "" || l.Stock != 0 {
		t.Fatalf("conservar al cambiar opciones: %+v", p.Variantes)
	}
	if p.Stock != 4 {
		t.Fatalf("stock después de sacar S: %d", p.Stock)
	}

	// Sacar la foto de la galería la saca de la variante; desactivar la saca de la suma.
	p, err = listas.Guardar(ctx, b.App, luna, p.ID, listas.Editable{Nombre: "Remera lisa", Precio: 20000, Fotos: []string{"https://f/a.jpg"},
		Variantes: []listas.VarianteEditable{{Valores: m.Valores, Activa: ptr(false), Precio: ptr(int64(21000))}}})
	if err != nil {
		t.Fatal(err)
	}
	if m = variante(p, "Talle", "M"); m.Foto != nil || m.Activa || m.Stock != 4 || p.Stock != 0 {
		t.Fatalf("desactivar: %+v stock %d", m, p.Stock)
	}

	// Un editor viejo (sin fotos): foto reemplaza la portada y deja el resto.
	p, _ = listas.Guardar(ctx, b.App, luna, p.ID, listas.Editable{Nombre: "Remera lisa", Precio: 20000, Fotos: []string{"https://f/a.jpg", "https://f/b.jpg"}})
	p, _ = listas.Guardar(ctx, b.App, luna, p.ID, listas.Editable{Nombre: "Remera lisa", Precio: 20000, Foto: ptr("https://f/c.jpg")})
	if !slices.Equal(p.Fotos, []string{"https://f/c.jpg", "https://f/b.jpg"}) || *p.Foto != "https://f/c.jpg" || len(p.Variantes) != 2 {
		t.Fatalf("compatibilidad con foto: %+v", p)
	}

	// Sin opciones vuelve a ser un producto simple: las variantes se borran y el stock se edita.
	p, err = listas.Guardar(ctx, b.App, luna, p.ID, listas.Editable{Nombre: "Remera lisa", Precio: 20000, Stock: ptr(9), Opciones: []listas.Opcion{}})
	if err != nil || len(p.Variantes) != 0 || p.Stock != 9 {
		t.Fatalf("sacar opciones: %+v %v", p, err)
	}

	// Validaciones.
	malas := []listas.Editable{
		{Nombre: "Remera", Opciones: []listas.Opcion{{"A", []string{"1"}}, {"B", []string{"1"}}, {"C", []string{"1"}}, {"D", []string{"1"}}}},
		{Nombre: "Remera", Opciones: []listas.Opcion{{"Talle", []string{"M", "m"}}}},
		{Nombre: "Remera", Opciones: []listas.Opcion{{"Talle", []string{"M"}}, {"talle", []string{"L"}}}},
		{Nombre: "Remera", Opciones: []listas.Opcion{{"Talle", []string{}}}},
		{Nombre: "Remera", Opciones: []listas.Opcion{{"Talle", []string{strings.Repeat("x", 21)}}}},
		{Nombre: "Remera", Opciones: []listas.Opcion{{"A", seq(20)}, {"B", seq(6)}}},
		{Nombre: "Remera", Fotos: make([]string, 9)},
		{Nombre: "Remera", Variantes: []listas.VarianteEditable{{Valores: map[string]string{"Talle": "M"}}}},
		{Nombre: "Remera", Opciones: []listas.Opcion{{"Talle", []string{"M"}}}, Variantes: []listas.VarianteEditable{{Valores: map[string]string{"Talle": "XL"}}}},
		{Nombre: "Remera", Opciones: []listas.Opcion{{"Talle", []string{"M"}}}, Variantes: []listas.VarianteEditable{{Valores: map[string]string{"Talle": "M"}, Foto: ptr("https://otra")}}},
		{Nombre: "Remera", Opciones: []listas.Opcion{{"Talle", []string{"M"}}}, Variantes: []listas.VarianteEditable{{Valores: map[string]string{"Talle": "M"}, Stock: ptr(-1)}}},
	}
	for i, e := range malas {
		if _, err := listas.Guardar(ctx, b.App, luna, "", e); err == nil {
			t.Errorf("caso %d: aceptó %+v", i, e)
		}
	}
}

func seq(n int) []string {
	out := []string{}
	for i := range n {
		out = append(out, strings.Repeat("v", i+1))
	}
	return out
}

func TestStockPorVariante(t *testing.T) {
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	sol := pruebadb.CrearTienda(t, b.App, "sol")
	p, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Zapatilla urbana", Precio: 80000,
		Opciones: []listas.Opcion{{Nombre: "Número", Valores: []string{"38", "39", "40"}}}})
	if err != nil {
		t.Fatal(err)
	}
	clave := func(v listas.Variante) string { return p.ID + ":" + v.ID }
	v38, v39 := variante(p, "Número", "38"), variante(p, "Número", "39")

	// Ingreso a dos variantes y al producto entero (que no se toca: tiene opciones).
	if n, err := listas.AjustarStock(ctx, b.App, luna, []string{clave(v38), clave(v39), p.ID}, 5); err != nil || n != 2 {
		t.Fatalf("ajuste: %d %v", n, err)
	}
	if n, _ := listas.FijarStock(ctx, b.App, luna, clave(v39), 2); n != 1 {
		t.Fatalf("conteo: %d", n)
	}
	if n, _ := listas.FijarStock(ctx, b.App, luna, p.ID, 100); n != 0 {
		t.Fatal("fijó el stock de un producto con opciones")
	}
	prods, _ := listas.Productos(ctx, b.App, luna, time.Now())
	inv := listas.Inventario(prods)
	vs := inv[0]["variantes"].([]listas.Registro)
	if inv[0]["stock"] != 7 || len(vs) != 3 || vs[0]["id"] != clave(v38) || vs[0]["stock"] != 5 || vs[0]["texto"] != "Número 38" {
		t.Fatalf("inventario: %+v", inv[0])
	}

	// Aislamiento: otra tienda no ve las variantes ni las puede tocar.
	if n, _ := listas.FijarStock(ctx, b.App, sol, clave(v38), 0); n != 0 {
		t.Fatal("otra tienda fijó el stock")
	}
	if n, _ := listas.AjustarStock(ctx, b.App, sol, []string{clave(v38)}, -5); n != 0 {
		t.Fatal("otra tienda ajustó el stock")
	}
	var cuantas int
	_ = b.App.EnTienda(ctx, sol, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT count(*) FROM variantes`).Scan(&cuantas)
	})
	if cuantas != 0 {
		t.Fatalf("otra tienda ve %d variantes", cuantas)
	}
	// Ni puede colgarle variantes a un producto ajeno.
	err = b.App.EnTienda(ctx, sol, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO variantes (tienda_id, producto_id, valores) VALUES ($1, $2::uuid, '{"Número": "41"}')`, luna, p.ID)
		return err
	})
	if err == nil {
		t.Fatal("insertó una variante en otra tienda")
	}

	// Subir precios toca también los de las variantes con precio propio.
	_, _ = listas.Guardar(ctx, b.App, luna, p.ID, listas.Editable{Nombre: "Zapatilla urbana", Precio: 80000,
		Variantes: []listas.VarianteEditable{{Valores: v38.Valores, Precio: ptr(int64(90000))}}})
	if n, _ := listas.Precio(ctx, b.App, luna, []string{p.ID}, 10); n != 1 {
		t.Fatal("precio")
	}
	p, _ = listas.Obtener(ctx, b.App, luna, p.ID, time.Now())
	if p.Precio != 88000 || *variante(p, "Número", "38").Precio != 99000 || variante(p, "Número", "40").Precio != nil {
		t.Fatalf("aumento: %d %+v", p.Precio, p.Variantes)
	}
}

// Las apps mandan cada variante con valores, precio y SKU (stock sólo si se tocó): lo que no viene en
// el JSON se conserva; null y "" sí borran.
func TestVariantesConservanLoQueNoVino(t *testing.T) {
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	fotos := []string{"https://f/a.jpg", "https://f/b.jpg"}
	p, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Remera lisa", Precio: 20000, PrecioMayorista: ptr(int64(15000)),
		Fotos: fotos, Opciones: []listas.Opcion{{Nombre: "Talle", Valores: []string{"M"}}},
		Variantes: []listas.VarianteEditable{{Valores: map[string]string{"Talle": "M"}, Stock: ptr(7), SKU: ptr("REM-M"),
			Foto: ptr("https://f/b.jpg"), PrecioMayorista: ptr(int64(14000))}}})
	if err != nil {
		t.Fatal(err)
	}
	guardar := func(cuerpo string) listas.Producto {
		t.Helper()
		var e listas.Editable
		if err := json.Unmarshal([]byte(cuerpo), &e); err != nil {
			t.Fatal(err)
		}
		p, err := listas.Guardar(ctx, b.App, luna, p.ID, e)
		if err != nil {
			t.Fatal(err)
		}
		return p
	}
	// Una venta entra con el editor abierto: el stock baja a 6 y el editor no lo manda.
	if err := b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE variantes SET stock = 6 WHERE tienda_id = $1 AND producto_id = $2::uuid`, luna, p.ID)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	p = guardar(`{"nombre": "Remera lisa", "precio": 20000, "fotos": ["https://f/a.jpg", "https://f/b.jpg"],
		"opciones": [{"nombre": "Talle", "valores": ["M"]}],
		"variantes": [{"valores": {"Talle": "M"}, "precio": 21000, "sku": "REM-M"}]}`)
	m := variante(p, "Talle", "M")
	if m.Foto == nil || *m.Foto != "https://f/b.jpg" || m.PrecioMayorista == nil || *m.PrecioMayorista != 14000 || m.Stock != 6 ||
		*m.Precio != 21000 || p.Stock != 6 {
		t.Fatalf("pisó lo que no vino: %+v stock %d", m, p.Stock)
	}
	// SKU vacío lo borra; null saca foto y mayorista (como manda la web).
	p = guardar(`{"nombre": "Remera lisa", "precio": 20000, "fotos": ["https://f/a.jpg", "https://f/b.jpg"],
		"variantes": [{"valores": {"Talle": "M"}, "precio": null, "sku": "", "foto": null, "precioMayorista": null}]}`)
	if m = variante(p, "Talle", "M"); m.SKU != "" || m.Foto != nil || m.PrecioMayorista != nil || m.Precio != nil || m.Stock != 6 {
		t.Fatalf("borrar con null y vacío: %+v", m)
	}
}

// Editar un producto pausado no lo vuelve a publicar; sin stock en el cuerpo queda el que hay.
func TestEditarConservaPausaYStock(t *testing.T) {
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	p, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Mate", Precio: 9000, Stock: ptr(5)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := listas.Publicar(ctx, b.App, luna, []string{p.ID}, false); err != nil {
		t.Fatal(err)
	}
	if err := b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE productos SET stock = 4 WHERE tienda_id = $1 AND id = $2::uuid`, luna, p.ID)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	p, err = listas.Guardar(ctx, b.App, luna, p.ID, listas.Editable{Nombre: "Mate imperial", Precio: 9500})
	if err != nil {
		t.Fatal(err)
	}
	if p.Publicado || p.Stock != 4 || p.Nombre != "Mate imperial" {
		t.Fatalf("editar pausado: publicado %v stock %d", p.Publicado, p.Stock)
	}
}
