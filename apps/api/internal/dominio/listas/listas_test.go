package listas_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

var ctx = context.Background()

func TestFiltroGenerico(t *testing.T) {
	regs := []listas.Registro{
		{"id": "1", "nombre": "Mate de calabaza", "sku": "A", "categoria": "Mates", "precio": int64(18500), "vendidos30": 30, "stock": 4, "margen": 40, "canales": []string{"web", "ml"}, "estado": "active", "lowm": false, "star": false},
		{"id": "2", "nombre": "Yerba orgánica", "sku": "B", "categoria": "Yerbas", "precio": int64(8200), "vendidos30": 90, "stock": 0, "margen": 20, "canales": []string{"web"}, "estado": "out", "lowm": true, "star": true},
		{"id": "3", "nombre": "Bombilla", "sku": "C", "categoria": "Mates", "precio": int64(1700), "vendidos30": 10, "stock": 9, "margen": 50, "canales": []string{"ml"}, "estado": "active", "lowm": false, "star": false},
	}
	l := listas.Aplicar(regs, listas.DefProductos, listas.Filtros{Facetas: map[string][]string{"categoria": {"Mates"}}})
	if l.Total != 2 || l.Facetas["categoria"]["Yerbas"] != 1 || l.Segmentos["out"] != 0 {
		t.Fatalf("faceta: %+v", l)
	}
	if l.Filas[0]["id"] != "1" {
		t.Fatalf("orden por vendidos: %v", l.Filas[0]["id"])
	}
	q := listas.Aplicar(regs, listas.DefProductos, listas.Filtros{Q: "organica"})
	if q.Total != 1 {
		t.Fatalf("búsqueda sin tildes: %d", q.Total)
	}
	b := listas.Aplicar(regs, listas.DefProductos, listas.Filtros{Banderas: []string{"star"}, Rango: 5000})
	if b.Total != 1 || b.Filas[0]["id"] != "2" {
		t.Fatalf("bandera y rango: %+v", b.Filas)
	}
	if got := listas.Aplicar(regs, listas.DefProductos, listas.Filtros{Facetas: map[string][]string{"canales": {"ml"}}}).Total; got != 2 {
		t.Fatalf("faceta de lista: %d", got)
	}
}

func TestProductosYClientes(t *testing.T) {
	b := pruebadb.Nueva(t)
	una := pruebadb.CrearTienda(t, b.App, "una")
	otra := pruebadb.CrearTienda(t, b.App, "otra")
	stock := 5
	p, err := listas.Guardar(ctx, b.App, una, "", listas.Editable{Nombre: "Mate de calabaza", Precio: 18500, Stock: &stock, Canales: map[string]bool{"web": true, "ml": true}})
	if err != nil {
		t.Fatal(err)
	}
	if p.Slug != "mate-de-calabaza" || !p.Publicado || p.Stock != 5 {
		t.Fatalf("alta: %+v", p)
	}
	if _, err := listas.Guardar(ctx, b.App, una, "", listas.Editable{Nombre: "x"}); err == nil {
		t.Fatal("aceptó un nombre de una letra")
	}
	if _, err := listas.Obtener(ctx, b.App, otra, p.ID, time.Now()); !errors.Is(err, listas.ErrNoExiste) {
		t.Fatalf("aislamiento: %v", err)
	}
	if n, _ := listas.Precio(ctx, b.App, una, []string{p.ID}, 10); n != 1 {
		t.Fatalf("precio: %d", n)
	}
	if n, _ := listas.Publicar(ctx, b.App, una, []string{p.ID}, false); n != 1 {
		t.Fatalf("pausar: %d", n)
	}
	p, _ = listas.Obtener(ctx, b.App, una, p.ID, time.Now())
	if p.Precio != 20350 || p.Publicado {
		t.Fatalf("después: %+v", p)
	}

	email := "lucia@example.com"
	for range 2 {
		if _, err := pedidos.Crear(ctx, b.App, una, pedidos.Nuevo{
			Comprador: pedidos.Comprador{Nombre: "Lucía Fernández", Email: &email}, Canal: "wa", Pago: "mp", Envio: "andreani",
			Ciudad: "Rosario", Estado: "Entregado", Items: []pedidos.Item{{Nombre: "Mate de calabaza", Cantidad: 2, Precio: 18500}},
		}); err != nil {
			t.Fatal(err)
		}
	}
	regs, err := listas.Clientes(ctx, b.App, una, time.Now())
	if err != nil || len(regs) != 1 || regs[0]["pedidos"] != 2 || regs[0]["gastado"] != int64(74000) {
		t.Fatalf("clientes: %+v %v", regs, err)
	}
	prods, _ := listas.Productos(ctx, b.App, una, time.Now())
	if prods[0]["vendidos30"] != 4 {
		t.Fatalf("vendidos: %v", prods[0]["vendidos30"])
	}
	if otros, _ := listas.Clientes(ctx, b.App, otra, time.Now()); len(otros) != 0 {
		t.Fatalf("la otra tienda ve clientes: %d", len(otros))
	}
}

func TestInventarioYEnvios(t *testing.T) {
	b := pruebadb.Nueva(t)
	id := pruebadb.CrearTienda(t, b.App, "luna")
	stock := 3
	p, _ := listas.Guardar(ctx, b.App, id, "", listas.Editable{Nombre: "Termo", Precio: 1000, Stock: &stock})
	if n, _ := listas.AjustarStock(ctx, b.App, id, []string{p.ID}, -10); n != 1 {
		t.Fatalf("ajuste: %d", n)
	}
	prods, _ := listas.Productos(ctx, b.App, id, time.Now())
	inv := listas.Inventario(prods)
	if inv[0]["stock"] != 0 || inv[0]["st"] != "out" {
		t.Fatalf("no queda negativo: %+v", inv[0])
	}
	if n, _ := listas.FijarStock(ctx, b.App, id, p.ID, 40); n != 1 {
		t.Fatalf("conteo: %d", n)
	}
	n1, _ := pedidos.Crear(ctx, b.App, id, pedidos.Nuevo{Comprador: pedidos.Comprador{Nombre: "Ana"}, Canal: "web", Pago: "mp", Envio: "andreani", Ciudad: "Salta", Estado: "Listo para despachar", Items: []pedidos.Item{{Nombre: "Termo", Cantidad: 1, Precio: 1000}}})
	envs, _ := listas.Envios(ctx, b.App, id, time.Now())
	if len(envs) != 1 || envs[0]["estado"] != "todo" {
		t.Fatalf("envíos: %+v", envs)
	}
	if n, _ := listas.Etiquetas(ctx, b.App, id, nil, true); n != 1 {
		t.Fatalf("etiquetas: %d", n)
	}
	d, _ := pedidos.Obtener(ctx, b.App, id, n1)
	if d.Estado != "En camino" {
		t.Fatalf("estado: %s", d.Estado)
	}
	movs, _ := listas.MovimientosHoy(ctx, b.App, id, time.Now())
	if len(movs) != 1 || movs[0].Cantidad != -1 {
		t.Fatalf("movimientos: %+v", movs)
	}
}

func TestVentaEnElLocalDescuentaStock(t *testing.T) {
	b := pruebadb.Nueva(t)
	id := pruebadb.CrearTienda(t, b.App, "luna")
	stock := 2
	p, _ := listas.Guardar(ctx, b.App, id, "", listas.Editable{Nombre: "Termo", Precio: 42900, Stock: &stock})
	venta := func(cant int) (int, error) {
		var n pedidos.PedidoNuevo
		n.Pago, n.Canal = "efectivo", "pos"
		n.Items = append(n.Items, pedidos.Linea{ProductoID: p.ID, Cantidad: cant})
		num, _, err := pedidos.CrearDesdeProductos(ctx, b.App, id, n)
		return num, err
	}
	if _, err := venta(3); !errors.Is(err, pedidos.ErrSinStock) {
		t.Fatalf("vendió más de lo que hay: %v", err)
	}
	num, err := venta(2)
	if err != nil {
		t.Fatal(err)
	}
	d, _ := pedidos.Obtener(ctx, b.App, id, num)
	if d.Estado != "Entregado" || d.Total != 85800 || d.Cliente != "Consumidor final" {
		t.Fatalf("pedido: %+v", d.Fila)
	}
	p, _ = listas.Obtener(ctx, b.App, id, p.ID, time.Now())
	if p.Stock != 0 {
		t.Fatalf("stock: %d", p.Stock)
	}
}

// El tipo y sus datos (00071): se guardan, sin mandarlos quedan y se copian al duplicar.
func TestTipoYAtributos(t *testing.T) {
	b := pruebadb.Nueva(t)
	una := pruebadb.CrearTienda(t, b.App, "una")
	hogar := "hogar"
	p, err := listas.Guardar(ctx, b.App, una, "", listas.Editable{Nombre: "Mesa ratona", Precio: 90000, Tipo: &hogar,
		Atributos: map[string]string{"peso": "8500", "alto": "45", "material": "Pino", "marca": ""}})
	if err != nil {
		t.Fatal(err)
	}
	if p.Tipo == nil || *p.Tipo != "hogar" || p.Atributos["peso"] != "8500" || p.Atributos["material"] != "Pino" || len(p.Atributos) != 3 {
		t.Fatalf("alta: %v %v", p.Tipo, p.Atributos)
	}
	p, err = listas.Guardar(ctx, b.App, una, p.ID, listas.Editable{Nombre: "Mesa ratona", Precio: 95000})
	if err != nil || p.Tipo == nil || *p.Tipo != "hogar" || p.Atributos["alto"] != "45" {
		t.Fatalf("sin mandarlos tienen que quedar: %v %v %v", p.Tipo, p.Atributos, err)
	}
	c, err := listas.Duplicar(ctx, b.App, una, p.ID, time.Now())
	if err != nil || c.Tipo == nil || *c.Tipo != "hogar" || c.Atributos["peso"] != "8500" {
		t.Fatalf("copia: %v %v %v", c.Tipo, c.Atributos, err)
	}
	sinTipo, err := listas.Guardar(ctx, b.App, una, "", listas.Editable{Nombre: "Termo", Precio: 1000})
	if err != nil || sinTipo.Tipo != nil || sinTipo.Atributos == nil {
		t.Fatalf("sin tipo: %v %v %v", sinTipo.Tipo, sinTipo.Atributos, err)
	}
	malo := "auto"
	if _, err := listas.Guardar(ctx, b.App, una, "", listas.Editable{Nombre: "Auto", Precio: 1, Tipo: &malo}); err == nil {
		t.Fatal("aceptó un tipo que no existe")
	}
}
