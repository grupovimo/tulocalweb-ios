package pedidos_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

var ctx = context.Background()

func nuevo(nombre, canal, pago, envio, estado string, precio int64, hace time.Duration) pedidos.Nuevo {
	return pedidos.Nuevo{
		Comprador: pedidos.Comprador{Nombre: nombre},
		Canal:     canal, Pago: pago, Envio: envio, Ciudad: "Rosario", Estado: estado,
		Items:  []pedidos.Item{{Nombre: "Mate de calabaza", Cantidad: 1, Precio: precio}},
		Creado: time.Now().Add(-hace),
	}
}

func conTelefono(n pedidos.Nuevo, tel string) pedidos.Nuevo {
	n.Comprador.Telefono = &tel
	return n
}

func TestListarFiltraYCuentaFacetas(t *testing.T) {
	b := pruebadb.Nueva(t)
	una := pruebadb.CrearTienda(t, b.App, "una")
	otra := pruebadb.CrearTienda(t, b.App, "otra")
	for _, n := range []pedidos.Nuevo{
		nuevo("Lucía Fernández", "wa", "mp", "andreani", "Para preparar", 42900, time.Hour),
		conTelefono(nuevo("Martín Gómez", "ml", "mp", "oca", "En camino", 21990, 30*time.Hour), "11 5555 0000"),
		nuevo("Norma Ruiz", "wa", "transf", "retiro", "Sin pagar", 8000, 50*time.Hour),
	} {
		if _, err := pedidos.Crear(ctx, b.App, una, n); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := pedidos.Crear(ctx, b.App, otra, nuevo("Otro", "web", "mp", "moto", "Para preparar", 1, time.Hour)); err != nil {
		t.Fatal(err)
	}

	todo, err := pedidos.Listar(ctx, b.App, una, pedidos.Filtros{}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if todo.Total != 3 || todo.Suma != 42900+21990+8000 || todo.Filas[0].Numero != 1001 {
		t.Fatalf("lista: total %d suma %d primera %d (la más nueva es la de hace 1 h)", todo.Total, todo.Suma, todo.Filas[0].Numero)
	}
	// Con canal WhatsApp: el conteo de canal no se filtra a sí mismo; el de estado sí.
	wa, _ := pedidos.Listar(ctx, b.App, una, pedidos.Filtros{Canal: []string{"wa"}}, time.Now())
	if wa.Total != 2 || wa.Facetas.Canal["ml"] != 1 || wa.Etapas["En camino"] != 0 || wa.Etapas["Sin pagar"] != 1 {
		t.Fatalf("facetas: %+v etapas %+v", wa.Facetas.Canal, wa.Etapas)
	}
	q, _ := pedidos.Listar(ctx, b.App, una, pedidos.Filtros{Q: "gómez"}, time.Now())
	if q.Total != 1 || q.Filas[0].Cliente != "Martín Gómez" || q.Filas[0].Telefono == nil || *q.Filas[0].Telefono != "11 5555 0000" {
		t.Fatalf("búsqueda: %+v", q.Filas)
	}
	hoy, _ := pedidos.Listar(ctx, b.App, una, pedidos.Filtros{Fecha: "hoy", Minimo: 10000}, time.Now())
	if hoy.Total != 1 {
		t.Fatalf("hoy y mínimo: %d", hoy.Total)
	}
	monto, _ := pedidos.Listar(ctx, b.App, una, pedidos.Filtros{Orden: "monto", Cuantos: 1, Desde: 1}, time.Now())
	if len(monto.Filas) != 1 || monto.Filas[0].Total != 21990 {
		t.Fatalf("orden y ventana: %+v", monto.Filas)
	}
	// La otra tienda sólo ve lo suyo, aunque pida el número de la primera.
	if _, err := pedidos.Obtener(ctx, b.App, otra, 1003); !errors.Is(err, pedidos.ErrNoExiste) {
		t.Fatalf("aislamiento: %v", err)
	}
}

func TestEstadosYLote(t *testing.T) {
	b := pruebadb.Nueva(t)
	id := pruebadb.CrearTienda(t, b.App, "luna")
	n1, _ := pedidos.Crear(ctx, b.App, id, nuevo("Ana", "web", "mp", "andreani", "Para preparar", 1000, time.Hour))
	n2, _ := pedidos.Crear(ctx, b.App, id, nuevo("Beto", "web", "mp", "andreani", "Sin pagar", 1000, time.Hour))

	if err := pedidos.CambiarEstado(ctx, b.App, id, n1, "Entregado"); !errors.Is(err, pedidos.ErrTransicion) {
		t.Fatalf("saltear etapas: %v", err)
	}
	if err := pedidos.CambiarEstado(ctx, b.App, id, n1, "Listo para despachar"); err != nil {
		t.Fatal(err)
	}
	d, _ := pedidos.Obtener(ctx, b.App, id, n1)
	if d.Estado != "Listo para despachar" || !d.Historial[2].Hecho || d.Comprador.Pedidos != 1 {
		t.Fatalf("detalle: %s %+v", d.Estado, d.Historial)
	}
	// Cargado a mano: sin código público (no hay link para reseñar) y sin teléfono.
	if d.Codigo != nil || d.Telefono != nil {
		t.Fatalf("código y teléfono: %v %v", d.Codigo, d.Telefono)
	}
	if err := pedidos.CambiarEstado(ctx, b.App, id, n1, "Para preparar"); err != nil {
		t.Fatalf("deshacer: %v", err)
	}
	cambios, err := pedidos.EnLote(ctx, b.App, id, []int{n1, n2}, "preparar")
	if err != nil || len(cambios) != 1 || cambios[0].Numero != n1 || cambios[0].Antes != "Para preparar" {
		t.Fatalf("lote: %+v %v", cambios, err)
	}
}

func TestTransiciones(t *testing.T) {
	casos := []struct {
		de, a, envio string
		ok           bool
	}{
		{"Sin pagar", "Para preparar", "correo", true},
		{"Para preparar", "Sin pagar", "correo", true},
		{"Para preparar", "En camino", "correo", false},
		{"Entregado", "Cancelado", "correo", false},
		{"Cancelado", "Para preparar", "correo", true},
		{"En camino", "Inventado", "correo", false},
		{"Listo para despachar", "En camino", "correo", true},
		// El retiro en el local: Listo para retirar → Retirado, sin "En camino".
		{"Listo para despachar", "En camino", "retiro", false},
		{"Listo para despachar", "Entregado", "retiro", true},
		{"Entregado", "Listo para despachar", "retiro", true},
		{"Listo para despachar", "Entregado", "correo", false},
	}
	for _, c := range casos {
		if pedidos.TransicionValida(c.de, c.a, c.envio) != c.ok {
			t.Errorf("%s → %s (%s): quería %v", c.de, c.a, c.envio, c.ok)
		}
	}
}

// Abrir un pedido cancelado de alguien sin otras compras: min(creado_en) daba NULL y el detalle, 500.
func TestDetalleDeUnCanceladoSolo(t *testing.T) {
	b := pruebadb.Nueva(t)
	una := pruebadb.CrearTienda(t, b.App, "una")
	n, err := pedidos.Crear(ctx, b.App, una, nuevo("Rita Paz", "wa", "transf", "retiro", "Cancelado", 8000, time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	d, err := pedidos.Obtener(ctx, b.App, una, n)
	if err != nil {
		t.Fatal(err)
	}
	if d.Comprador.Pedidos != 0 || !d.Comprador.Desde.Equal(d.Creado) {
		t.Fatalf("comprador: %+v creado %v", d.Comprador, d.Creado)
	}
}
