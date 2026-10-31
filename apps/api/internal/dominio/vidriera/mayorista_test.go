package vidriera_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/configuracion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cupones"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/mayorista"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/vidriera"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

func p64(v int64) *int64 { return &v }
func pint(v int) *int    { return &v }

// Precios mayoristas (§49): Go los recalcula al comprar, con el mínimo de la tienda, y marca el pedido.
func TestVentaMayorista(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	if err := configuracion.GuardarNegocio(ctx, b.App, luna, configuracion.Negocio{Nombre: "Luna", Alias: "luna.tienda"}); err != nil {
		t.Fatal(err)
	}
	stock := 200
	mate, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Mate", Precio: 10000, PrecioMayorista: p64(8000), Stock: &stock})
	if err != nil {
		t.Fatal(err)
	}
	yerba, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Yerba", Precio: 5000, Stock: &stock})
	if err != nil {
		t.Fatal(err)
	}
	cien := 100
	buzo, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Buzo", Precio: 24000, PrecioMayorista: p64(20000),
		Opciones: []listas.Opcion{{Nombre: "Talle", Valores: []string{"M", "L"}}},
		Variantes: []listas.VarianteEditable{
			{Valores: map[string]string{"Talle": "M"}, Stock: &cien},
			{Valores: map[string]string{"Talle": "L"}, Stock: &cien, Precio: p64(26000), PrecioMayorista: p64(21000)},
		}})
	if err != nil {
		t.Fatal(err)
	}
	if *mate.PrecioMayorista != 8000 || yerba.PrecioMayorista != nil || *buzo.Variantes[1].PrecioMayorista != 21000 {
		t.Fatalf("fichas: %+v %+v %+v", mate.PrecioMayorista, yerba.PrecioMayorista, buzo.Variantes)
	}
	talle := map[string]string{}
	for _, v := range buzo.Variantes {
		talle[v.Valores["Talle"]] = v.ID
	}

	// Apagada (de entrada): la tienda no publica el mínimo ni los precios mayoristas y cobra minorista.
	tienda, err := vidriera.Publica(ctx, b.App, "luna")
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range tienda.Productos {
		if p.Mayorista != nil {
			t.Fatalf("mayorista publicado con la venta apagada: %+v", p)
		}
	}
	if tienda.Mayorista != nil {
		t.Fatalf("mínimo publicado apagado: %+v", tienda.Mayorista)
	}
	lunes := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)
	comprar := func(lineas ...vidriera.LineaCarrito) (int64, bool) {
		t.Helper()
		c := compra(lineas[0].ID, lineas[0].Cantidad, "transferencia")
		c.Carrito = lineas
		r, err := vidriera.Comprar(ctx, b.App, nil, "", "luna", c, lunes)
		if err != nil {
			t.Fatal(err)
		}
		var marca bool
		var items string
		if err := b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
			return tx.QueryRow(ctx, `SELECT mayorista, items::text FROM pedidos WHERE numero = $1`, r.Numero).Scan(&marca, &items)
		}); err != nil {
			t.Fatal(err)
		}
		return r.Total, marca
	}
	linea := func(id string, n int) vidriera.LineaCarrito { return vidriera.LineaCarrito{ID: id, Cantidad: n} }
	if total, marca := comprar(linea(mate.ID, 10)); total != 100000 || marca {
		t.Fatalf("apagada: %d %v", total, marca)
	}

	// Sin mínimo no se guarda; prendida con 6 unidades.
	if err := configuracion.GuardarMayorista(ctx, b.App, luna, mayorista.Config{Activa: true}); !errors.Is(err, mayorista.ErrSinMinimo) {
		t.Fatalf("sin mínimo: %v", err)
	}
	if err := configuracion.GuardarMayorista(ctx, b.App, luna, mayorista.Config{Activa: true, Unidades: pint(6)}); err != nil {
		t.Fatal(err)
	}
	tienda, _ = vidriera.Publica(ctx, b.App, "luna")
	if tienda.Mayorista == nil || *tienda.Mayorista.Unidades != 6 || tienda.Mayorista.Monto != nil {
		t.Fatalf("mínimo público: %+v", tienda.Mayorista)
	}
	// Cupones › Automáticas muestra el mismo interruptor con el mínimo de verdad.
	cs, err := cupones.Leer(ctx, b.App, luna, lunes)
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range cs["automaticas"].([]map[string]any) {
		if a["id"] == "mayorista" && (a["activa"] != true || !strings.HasPrefix(a["detalle"].(string), "Desde 6 unidades")) {
			t.Fatalf("automática: %+v", a)
		}
	}

	// 5 unidades: minorista. 6 (4 mates y 2 yerbas): los mates a mayorista, la yerba (sin mayorista)
	// igual; el pedido queda marcado. Transferencia: 15 % sobre eso (no está prendida: sin descuento).
	if total, marca := comprar(linea(mate.ID, 3), linea(yerba.ID, 2)); total != 40000 || marca {
		t.Fatalf("5 unidades: %d %v", total, marca)
	}
	if total, marca := comprar(linea(mate.ID, 4), linea(yerba.ID, 2)); total != 4*8000+2*5000 || !marca {
		t.Fatalf("6 unidades: %d %v", total, marca)
	}
	// Al mínimo pero sin ningún producto con precio mayorista: no es un pedido mayorista.
	if total, marca := comprar(linea(yerba.ID, 6)); total != 30000 || marca {
		t.Fatalf("sin precios mayoristas: %d %v", total, marca)
	}
	// Variantes: la suya (L) o la del producto (M).
	l := vidriera.LineaCarrito{ID: buzo.ID, VarianteID: talle["L"], Cantidad: 3}
	m := vidriera.LineaCarrito{ID: buzo.ID, VarianteID: talle["M"], Cantidad: 3}
	if total, marca := comprar(l, m); total != 3*21000+3*20000 || !marca {
		t.Fatalf("variantes: %d %v", total, marca)
	}

	// Por monto (a precio minorista): $ 30.000 son 3 mates.
	if err := configuracion.GuardarMayorista(ctx, b.App, luna, mayorista.Config{Activa: true, Monto: p64(30000)}); err != nil {
		t.Fatal(err)
	}
	if total, marca := comprar(linea(mate.ID, 3)); total != 24000 || !marca {
		t.Fatalf("por monto: %d %v", total, marca)
	}
	if total, _ := comprar(linea(mate.ID, 2)); total != 20000 {
		t.Fatalf("por monto, abajo: %d", total)
	}
	// Con los dos mínimos, tienen que cumplirse los dos.
	if err := configuracion.GuardarMayorista(ctx, b.App, luna, mayorista.Config{Activa: true, Unidades: pint(6), Monto: p64(100000)}); err != nil {
		t.Fatal(err)
	}
	if total, _ := comprar(linea(mate.ID, 6)); total != 60000 {
		t.Fatalf("los dos, sin monto: %d", total)
	}
	if total, _ := comprar(linea(mate.ID, 10)); total != 80000 {
		t.Fatalf("los dos: %d", total)
	}

	// Editar: nil deja el mayorista, 0 lo saca; quien no toca precios (Ventas) lo conserva.
	editar := func(e listas.Editable) listas.Producto {
		t.Helper()
		p, err := listas.Guardar(ctx, b.App, luna, mate.ID, e)
		if err != nil {
			t.Fatal(err)
		}
		return p
	}
	if p := editar(listas.Editable{Nombre: "Mate", Precio: 10000}); p.PrecioMayorista == nil || *p.PrecioMayorista != 8000 {
		t.Fatalf("nil lo borró: %+v", p.PrecioMayorista)
	}
	if p := editar(listas.Editable{Nombre: "Mate", Precio: 1, PrecioMayorista: p64(0), SinPrecios: true}); p.Precio != 10000 ||
		p.PrecioMayorista == nil || *p.PrecioMayorista != 8000 {
		t.Fatalf("Ventas cambió precios: %+v", p)
	}
	if _, err := listas.Guardar(ctx, b.App, luna, mate.ID, listas.Editable{Nombre: "Mate", Precio: 10000, PrecioMayorista: p64(10000)}); err == nil {
		t.Fatal("un mayorista igual al precio")
	}
	if p := editar(listas.Editable{Nombre: "Mate", Precio: 10000, PrecioMayorista: p64(0)}); p.PrecioMayorista != nil {
		t.Fatalf("0 no lo sacó: %+v", p.PrecioMayorista)
	}
	// El cambio de precio por porcentaje lleva el mayorista también (y no inventa uno donde no hay).
	editar(listas.Editable{Nombre: "Mate", Precio: 10000, PrecioMayorista: p64(8000)})
	if _, err := listas.Precio(ctx, b.App, luna, []string{mate.ID, yerba.ID, buzo.ID}, 10); err != nil {
		t.Fatal(err)
	}
	tienda, _ = vidriera.Publica(ctx, b.App, "luna")
	for _, p := range tienda.Productos {
		switch p.ID {
		case mate.ID:
			if p.Precio != 11000 || *p.Mayorista != 8800 {
				t.Fatalf("mate +10 %%: %+v", p)
			}
		case yerba.ID:
			if p.Mayorista != nil {
				t.Fatalf("yerba con mayorista inventado: %+v", p.Mayorista)
			}
		case buzo.ID:
			for _, v := range p.Variantes {
				if v.ID == talle["M"] && (v.Precio != nil || v.Mayorista != nil) {
					t.Fatalf("variante M sin precios propios: %+v", v)
				}
				if v.ID == talle["L"] && (*v.Precio != 28600 || *v.Mayorista != 23100) {
					t.Fatalf("variante L +10 %%: %+v %+v", *v.Precio, *v.Mayorista)
				}
			}
		}
	}

	// Apagada otra vez: vuelve a minorista.
	if err := configuracion.GuardarMayorista(ctx, b.App, luna, mayorista.Config{Unidades: pint(6)}); err != nil {
		t.Fatal(err)
	}
	if total, marca := comprar(linea(yerba.ID, 1), linea(mate.ID, 10)); total != 5500+110000 || marca {
		t.Fatalf("apagada de nuevo: %d %v", total, marca)
	}
}
