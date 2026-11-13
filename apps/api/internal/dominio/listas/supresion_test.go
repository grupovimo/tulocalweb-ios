package listas_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/configuracion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

func TestSuprimirYEditarCliente(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	admin, err := pgx.Connect(ctx, b.AdminURL)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close(ctx)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	sol := pruebadb.CrearTienda(t, b.App, "sol")
	ejecutar := func(tienda, q string, args ...any) {
		t.Helper()
		if _, err := admin.Exec(ctx, `SELECT set_config('app.tienda_id', $1, false)`, tienda); err != nil {
			t.Fatal(err)
		}
		if _, err := admin.Exec(ctx, q, args...); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}
	pedido := func(tienda string, numero int, comprador, estado string) {
		ejecutar(tienda, `INSERT INTO pedidos (tienda_id, numero, comprador, canal, pago, envio, estado, items, subtotal, total, nota, entrega)
			VALUES ($1, $2, $3, 'web', 'mp', 'correo', $4, '[{"nombre": "Mate"}]', 1000, 1000, 'tocar timbre 3B',
				'{"direccion": "Mitre 123", "cp": "7600", "provincia": "B"}')`, tienda, numero, comprador, estado)
	}
	pedido(luna, 1, `{"nombre": "Ana Pérez", "email": "Ana@x.com", "telefono": "11 5555-1234"}`, "Entregado")
	pedido(luna, 2, `{"nombre": "Beto", "email": "beto@x.com"}`, "Para preparar")
	pedido(sol, 1, `{"nombre": "Ana Pérez", "email": "ana@x.com"}`, "Entregado")
	ejecutar(luna, `INSERT INTO clientes (tienda_id, nombre, email, telefono, origen) VALUES ($1, 'Ana Pérez', 'ana@x.com', '11 5555-1234', 'panel')`, luna)
	ejecutar(luna, `INSERT INTO resenas (tienda_id, pedido_numero, autor, estrellas, texto, origen) VALUES ($1, 1, 'Ana P.', 5, 'Hermoso', 'web')`, luna)
	ejecutar(luna, `INSERT INTO suscriptores (tienda_id, email, origen, token) VALUES ($1, 'ana.maria@x.com', 'tienda', 'tok-ana-1234567890123')`, luna)

	// Rectificar: nombre, email y teléfono en sus pedidos y en su ficha.
	if _, err := listas.EditarCliente(ctx, b.App, luna, "ana@x.com", json.RawMessage(`{"nombre": "x"}`)); !errors.Is(err, listas.ErrDatosCliente) {
		t.Fatalf("nombre corto: %v", err)
	}
	if _, err := listas.EditarCliente(ctx, b.App, luna, "ana@x.com", json.RawMessage(`{"nombre": "Ana", "email": "beto@x.com"}`)); !errors.Is(err, listas.ErrEmailEnUso) {
		t.Fatalf("email de otro: %v", err)
	}
	n, err := listas.EditarCliente(ctx, b.App, luna, "ana@x.com", json.RawMessage(`{"nombre": "Ana María Pérez", "email": "ANA.MARIA@x.com", "telefono": "223 555-0000"}`))
	if err != nil || n != 1 {
		t.Fatalf("editar: %d %v", n, err)
	}
	regs, err := listas.Clientes(ctx, b.App, luna, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	var ana listas.Registro
	for _, r := range regs {
		if r["id"] == "ana.maria@x.com" {
			ana = r
		}
	}
	if ana == nil || ana["nombre"] != "Ana María Pérez" || *ana["telefono"].(*string) != "223 555-0000" || len(regs) != 2 {
		t.Fatalf("después de editar: %+v", regs)
	}

	// Suprimir: con un pedido en curso, no; a Ana sí, y sólo en esta tienda.
	if _, err := listas.SuprimirClientes(ctx, b.App, luna, []string{"beto@x.com"}); !errors.Is(err, listas.ErrEnCurso) {
		t.Fatalf("con pedido en curso: %v", err)
	}
	if n, err := listas.SuprimirClientes(ctx, b.App, luna, []string{"ana.maria@x.com"}); err != nil || n != 1 {
		t.Fatalf("suprimir: %d %v", n, err)
	}
	var comprador, entrega string
	var nota *string
	if err := admin.QueryRow(ctx, `SELECT comprador::text, entrega::text, nota FROM pedidos WHERE tienda_id = $1 AND numero = 1`, luna).Scan(&comprador, &entrega, &nota); err != nil {
		t.Fatal(err)
	}
	if comprador != `{"nombre": "Consumidor final"}` || entrega != `{"provincia": "B"}` || nota != nil {
		t.Fatalf("pedido anonimizado: %s %s %v", comprador, entrega, nota)
	}
	var fichas, suscriptores int
	var autor string
	if err := admin.QueryRow(ctx, `SELECT (SELECT count(*) FROM clientes WHERE tienda_id = $1), (SELECT count(*) FROM suscriptores WHERE tienda_id = $1),
			(SELECT autor FROM resenas WHERE tienda_id = $1)`, luna).Scan(&fichas, &suscriptores, &autor); err != nil {
		t.Fatal(err)
	}
	if fichas != 0 || suscriptores != 0 || autor != "Anónimo" {
		t.Fatalf("fichas=%d suscriptores=%d autor=%q", fichas, suscriptores, autor)
	}
	regs, _ = listas.Clientes(ctx, b.App, luna, time.Now())
	if len(regs) != 1 || regs[0]["id"] != "beto@x.com" {
		t.Fatalf("la lista sigue mostrando a Ana: %+v", regs)
	}
	if regs, _ = listas.Clientes(ctx, b.App, sol, time.Now()); len(regs) != 1 {
		t.Fatal("la supresión tocó otra tienda")
	}
	var actividad int
	_ = admin.QueryRow(ctx, `SELECT count(*) FROM auditoria WHERE tienda_id = $1 AND texto LIKE '%Ana%' AND accion = 'cliente.suprimir'`, luna).Scan(&actividad)
	if actividad != 0 {
		t.Fatal("la actividad nombra a quien se suprimió")
	}
}

func TestBorrarInactivos(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	admin, err := pgx.Connect(ctx, b.AdminURL)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close(ctx)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	if _, err := admin.Exec(ctx, `SELECT set_config('app.tienda_id', $1, false)`, luna); err != nil {
		t.Fatal(err)
	}
	for _, q := range []string{
		`INSERT INTO pedidos (tienda_id, numero, comprador, canal, pago, envio, estado, items, subtotal, total, creado_en)
			VALUES ($1, 1, '{"nombre": "Vieja", "email": "vieja@x.com"}', 'web', 'mp', 'retiro', 'Entregado', '[{"nombre": "Mate"}]', 1, 1, now() - interval '4 years')`,
		`INSERT INTO pedidos (tienda_id, numero, comprador, canal, pago, envio, estado, items, subtotal, total)
			VALUES ($1, 2, '{"nombre": "Nueva", "email": "nueva@x.com"}', 'web', 'mp', 'retiro', 'Entregado', '[{"nombre": "Mate"}]', 1, 1)`,
		`INSERT INTO clientes (tienda_id, nombre, origen, creado_en) VALUES ($1, 'Ficha vieja', 'panel', now() - interval '5 years')`,
	} {
		if _, err := admin.Exec(ctx, q, luna); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}
	d, err := configuracion.MisDatos(ctx, b.App, luna)
	if err != nil || d["inactivos"] != 2 || d["borrado"] != nil {
		t.Fatalf("mis datos: %+v %v", d, err)
	}
	if n, err := configuracion.ProgramarBorradoInactivos(ctx, b.App, luna); err != nil || n != 2 {
		t.Fatalf("programar: %d %v", n, err)
	}
	if d, _ = configuracion.MisDatos(ctx, b.App, luna); d["borrado"] == nil {
		t.Fatal("no quedó programado")
	}
	// Antes de los 7 días el worker no hace nada; frenado, tampoco.
	if n, err := configuracion.BorrarInactivosProgramados(ctx, b.App); err != nil || n != 0 {
		t.Fatalf("antes de tiempo: %d %v", n, err)
	}
	if _, err := admin.Exec(ctx, `UPDATE tiendas SET borrado_inactivos = now() - interval '1 minute' WHERE id = $1`, luna); err != nil {
		t.Fatal(err)
	}
	if n, err := configuracion.BorrarInactivosProgramados(ctx, b.App); err != nil || n != 2 {
		t.Fatalf("borrar: %d %v", n, err)
	}
	regs, _ := listas.Clientes(ctx, b.App, luna, time.Now())
	if len(regs) != 1 || regs[0]["id"] != "nueva@x.com" {
		t.Fatalf("después del borrado: %+v", regs)
	}
	if d, _ = configuracion.MisDatos(ctx, b.App, luna); d["inactivos"] != 0 || d["borrado"] != nil {
		t.Fatalf("mis datos después: %+v", d)
	}
}
