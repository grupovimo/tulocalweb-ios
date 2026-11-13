package configuracion_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/configuracion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

// sembrar carga, como superusuario (sin RLS), lo mínimo de una tienda con ventas y facturas.
func sembrar(t *testing.T, admin *pgx.Conn, tienda string) {
	t.Helper()
	ctx := context.Background()
	// Los triggers (outbox, tiempo real) escriben con la tienda fijada.
	if _, err := admin.Exec(ctx, `SELECT set_config('app.tienda_id', $1, false)`, tienda); err != nil {
		t.Fatal(err)
	}
	for _, q := range []string{
		`INSERT INTO productos (tienda_id, nombre, precio) VALUES ($1, 'Mate', 1000)`,
		`INSERT INTO clientes (tienda_id, nombre, email, origen) VALUES ($1, 'Ana Pérez', 'ana@x.com', 'panel')`,
		`INSERT INTO pedidos (tienda_id, numero, comprador, canal, pago, envio, estado, items, subtotal, total)
			VALUES ($1, 1, '{"nombre": "Ana Pérez", "email": "ana@x.com"}', 'web', 'mp', 'retiro', 'Entregado', '[{"nombre": "Mate"}]', 1000, 1000)`,
		`INSERT INTO arca_config (tienda_id, cuit, razon_social, condicion, punto_venta, clave)
			VALUES ($1, '20123456789', 'Luna SRL', 'monotributo', 3, '\x00')`,
		`INSERT INTO comprobantes (tienda_id, pedido_numero, tipo, punto_venta, numero, cae, estado, receptor, concepto, total, neto, iva, emitido_en)
			VALUES ($1, 1, 11, 3, 7, '12345678901234', 'emitido',
				'{"nombre": "Ana Pérez", "docTipo": 96, "docNro": 30111222, "condicion": 5}', 'Pedido #1', 1000, 100000, 0, now())`,
		`INSERT INTO comprobantes (tienda_id, tipo, punto_venta, estado, receptor, total, neto, iva)
			VALUES ($1, 11, 3, 'pendiente', '{"nombre": "Otro"}', 500, 50000, 0)`,
		`INSERT INTO abono_cobros (tienda_id, clave, detalle, importe) VALUES ($1, 'mp-1', 'Plan Pro, noviembre', 1500000)`,
	} {
		if _, err := admin.Exec(ctx, q, tienda); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}
}

func TestPurgarCerradas(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	admin, err := pgx.Connect(ctx, b.AdminURL)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close(ctx)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	sol := pruebadb.CrearTienda(t, b.App, "sol")
	mar := pruebadb.CrearTienda(t, b.App, "mar")
	sembrar(t, admin, luna)
	sembrar(t, admin, sol)
	// Luna se cerró hace 91 días; Sol hace 10 (todavía puede volver); Mar sigue abierta.
	if _, err := admin.Exec(ctx, `UPDATE tiendas SET cerrada_en = now() - interval '91 days' WHERE id = $1`, luna); err != nil {
		t.Fatal(err)
	}
	if _, err := admin.Exec(ctx, `UPDATE tiendas SET cerrada_en = now() - interval '10 days' WHERE id = $1`, sol); err != nil {
		t.Fatal(err)
	}

	l, err := configuracion.ParaPurgar(ctx, b.App, 100)
	if err != nil || len(l) != 1 || l[0].ID != luna || l[0].Nombre != "Tienda luna" || l[0].CerradaEn.IsZero() {
		t.Fatalf("para purgar: %+v %v", l, err)
	}

	dir := t.TempDir()
	for _, id := range []string{luna, sol} {
		if err := os.MkdirAll(filepath.Join(dir, "fotos", id), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "fotos", id, "a.webp"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	n, err := configuracion.PurgarCerradas(ctx, b.App, dir)
	if err != nil || n != 1 {
		t.Fatalf("purgar: %d %v", n, err)
	}

	cuenta := func(q string, args ...any) int {
		t.Helper()
		var c int
		if err := admin.QueryRow(ctx, q, args...).Scan(&c); err != nil {
			t.Fatal(err)
		}
		return c
	}
	for _, tabla := range []string{"productos", "clientes", "pedidos", "comprobantes", "abono_cobros", "arca_config", "auth.membresias", "auditoria"} {
		if c := cuenta(`SELECT count(*) FROM `+tabla+` WHERE tienda_id = $1`, luna); c != 0 {
			t.Errorf("%s: quedaron %d filas de la tienda borrada", tabla, c)
		}
	}
	if cuenta(`SELECT count(*) FROM tiendas WHERE id = $1`, luna) != 0 {
		t.Fatal("la tienda sigue")
	}
	if cuenta(`SELECT count(*) FROM pedidos WHERE tienda_id = $1`, sol) != 1 || cuenta(`SELECT count(*) FROM tiendas WHERE id IN ($1, $2)`, sol, mar) != 2 {
		t.Fatal("se llevó otra tienda")
	}
	if _, err := os.Stat(filepath.Join(dir, "fotos", luna)); !os.IsNotExist(err) {
		t.Fatalf("las fotos de la tienda borrada siguen: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "fotos", sol, "a.webp")); err != nil {
		t.Fatalf("borró fotos de otra tienda: %v", err)
	}

	// Lo contable: el comprobante emitido (no el pendiente) y el cobro del abono, sin datos de quien
	// compró.
	filas, err := admin.Query(ctx, `SELECT tipo, datos FROM archivo_contable WHERE tienda_id = $1 ORDER BY tipo`, luna)
	if err != nil {
		t.Fatal(err)
	}
	type archivado struct {
		Tipo  string
		Datos map[string]any
	}
	arch, err := pgx.CollectRows(filas, func(f pgx.CollectableRow) (archivado, error) {
		var a archivado
		var d []byte
		if err := f.Scan(&a.Tipo, &d); err != nil {
			return a, err
		}
		return a, json.Unmarshal(d, &a.Datos)
	})
	if err != nil || len(arch) != 2 || arch[0].Tipo != "abono" || arch[1].Tipo != "comprobante" {
		t.Fatalf("archivo contable: %+v %v", arch, err)
	}
	c := arch[1].Datos
	rec, _ := c["receptor"].(map[string]any)
	emisor, _ := c["emisor"].(map[string]any)
	if c["cae"] != "12345678901234" || c["numero"] != float64(7) || emisor["cuit"] != "20123456789" ||
		rec["nombre"] != nil || rec["docNro"] != nil || rec["docTipo"] != float64(96) {
		t.Fatalf("comprobante archivado: %+v", c)
	}
	if arch[0].Datos["importe"] != float64(1500000) {
		t.Fatalf("abono archivado: %+v", arch[0].Datos)
	}

	// Una segunda vuelta no hace nada; y lo archivado con más de 10 años se borra.
	if n, err := configuracion.PurgarCerradas(ctx, b.App, dir); err != nil || n != 0 {
		t.Fatalf("segunda vuelta: %d %v", n, err)
	}
	if _, err := admin.Exec(ctx, `UPDATE archivo_contable SET fecha = now() - interval '10 years 1 day' WHERE tipo = 'abono'`); err != nil {
		t.Fatal(err)
	}
	if _, err := configuracion.PurgarCerradas(ctx, b.App, ""); err != nil {
		t.Fatal(err)
	}
	if cuenta(`SELECT count(*) FROM archivo_contable`) != 1 {
		t.Fatal("no borró lo archivado hace más de 10 años, o borró de más")
	}
	// La app no ve el archivo.
	var visto int
	if err := b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT count(*) FROM archivo_contable`).Scan(&visto)
	}); err == nil {
		t.Fatalf("la app leyó el archivo contable (%d filas)", visto)
	}
}
