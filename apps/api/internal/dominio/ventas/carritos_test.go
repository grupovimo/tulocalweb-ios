package ventas_test

import (
	"context"
	"errors"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/configuracion"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/ventas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/vidriera"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

func TestCarritosAbandonados(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	// Con datos para transferir (sin ellos la tienda no ofrece transferencia).
	if err := configuracion.GuardarNegocio(ctx, b.App, luna, configuracion.Negocio{Nombre: "Luna", Alias: "luna.tienda"}); err != nil {
		t.Fatal(err)
	}
	sol := pruebadb.CrearTienda(t, b.App, "sol")
	mails := &correo.Memoria{}
	w := &outbox.Worker{Pool: b.Worker, Manejadores: (&ventas.MailsCarritos{Pool: b.App, Correo: mails, Sitio: "https://tulocalweb.com.ar"}).Manejadores()}
	stock := 10
	p, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Mate", Precio: 10000, Stock: &stock})
	if err != nil {
		t.Fatal(err)
	}
	linea := []ventas.LineaCarrito{{ID: p.ID, Cantidad: 2}}
	ahora := time.Now()

	if err := ventas.GuardarCarrito(ctx, b.App, luna, "no-es-un-email", "", linea); err != ventas.ErrCarritoInvalido {
		t.Fatalf("email inválido: %v", err)
	}
	// Un producto de otra tienda no cuenta: el carrito queda vacío y no se guarda.
	if err := ventas.GuardarCarrito(ctx, b.App, sol, "ana@ejemplo.com", "Ana", linea); err != ventas.ErrCarritoInvalido {
		t.Fatalf("producto ajeno: %v", err)
	}
	if err := ventas.GuardarCarrito(ctx, b.App, luna, "Lucia@Ejemplo.com", "Lucía", linea); err != nil {
		t.Fatal(err)
	}
	// El mismo email actualiza el mismo carrito (no hay dos abiertos).
	if err := ventas.GuardarCarrito(ctx, b.App, luna, "lucia@ejemplo.com", "", []ventas.LineaCarrito{{ID: p.ID, Cantidad: 3}}); err != nil {
		t.Fatal(err)
	}
	cs, abiertos, monto, _, _, err := ventas.CarritosPanel(ctx, b.App, luna, ahora)
	if err != nil || len(cs) != 1 || abiertos != 1 || monto != 30000 || cs[0].Cantidad != 3 || cs[0].Nombre != "Lucía" {
		t.Fatalf("panel: %+v %d %d %v", cs, abiertos, monto, err)
	}
	if cs2, _, _, _, _, _ := ventas.CarritosPanel(ctx, b.App, sol, ahora); len(cs2) != 0 {
		t.Fatal("otra tienda ve carritos ajenos")
	}

	// Recordar: mail con el link que restaura el carrito; uno por día.
	if err := ventas.RecordarCarrito(ctx, b.App, luna, cs[0].ID, ahora); err != nil {
		t.Fatal(err)
	}
	if err := ventas.RecordarCarrito(ctx, b.App, luna, cs[0].ID, ahora.Add(time.Hour)); err != ventas.ErrRecienRecordado {
		t.Fatalf("dos el mismo día: %v", err)
	}
	if err := ventas.RecordarCarrito(ctx, b.App, sol, cs[0].ID, ahora); err != ventas.ErrNoExisteCarrito {
		t.Fatalf("otra tienda recordó: %v", err)
	}
	if _, err := w.Pasada(ctx); err != nil {
		t.Fatal(err)
	}
	m, ok := mails.Ultimo("lucia@ejemplo.com")
	if !ok || !strings.Contains(m.Texto, "3 × Mate") || !strings.Contains(m.Texto, "/t/luna/carrito/recuperar/") {
		t.Fatalf("mail: %+v", m)
	}
	token := m.Texto[strings.Index(m.Texto, "/recuperar/")+len("/recuperar/"):]
	token = strings.Fields(token)[0]
	if ls, err := ventas.CarritoPorToken(ctx, b.App, luna, token); err != nil || len(ls) != 1 || ls[0].Cantidad != 3 {
		t.Fatalf("restaurar: %+v %v", ls, err)
	}

	// Compra con el mismo email: el carrito queda recuperado y deja de aparecer.
	c := vidriera.Compra{Email: "lucia@ejemplo.com", Nombre: "Lucía Pérez", Telefono: "11 5555-0101", CP: "1414",
		Direccion: "Gorriti 4500", Envio: "retiro", Pago: "transferencia"}
	c.Carrito = append(c.Carrito, vidriera.LineaCarrito{ID: p.ID, Cantidad: 1})
	if _, err := vidriera.Comprar(ctx, b.App, nil, "", "luna", c, ahora); err != nil {
		t.Fatal(err)
	}
	if cs, _, _, _, _, _ := ventas.CarritosPanel(ctx, b.App, luna, ahora); len(cs) != 0 {
		t.Fatalf("sigue abierto después de comprar: %+v", cs)
	}

	// Automático al tercer día, sólo si la tienda lo prendió.
	if err := ventas.GuardarCarrito(ctx, b.App, luna, "ana@ejemplo.com", "Ana", linea); err != nil {
		t.Fatal(err)
	}
	_ = b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE carritos SET actualizado = now() - interval '3 days 12 hours' WHERE email = 'ana@ejemplo.com'`)
		return err
	})
	if n, _ := ventas.RecordatoriosAutomaticos(ctx, b.App, ahora); n != 0 {
		t.Fatalf("recordó con el automático apagado: %d", n)
	}
	if err := ventas.GuardarAutomaticos(ctx, b.App, luna, ventas.Automaticos{Email: true}); err != nil {
		t.Fatal(err)
	}
	if n, err := ventas.RecordatoriosAutomaticos(ctx, b.App, ahora); err != nil || n != 1 {
		t.Fatalf("automático: %d %v", n, err)
	}
}

func TestTresRecordatoriosConCuponYBaja(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	if err := configuracion.GuardarNegocio(ctx, b.App, luna, configuracion.Negocio{Nombre: "Luna", Alias: "luna.tienda"}); err != nil {
		t.Fatal(err)
	}
	mails := &correo.Memoria{}
	w := &outbox.Worker{Pool: b.Worker, Manejadores: (&ventas.MailsCarritos{Pool: b.App, Correo: mails, Sitio: "https://tulocalweb.com.ar"}).Manejadores()}
	stock := 10
	p, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Mate", Precio: 10000, Stock: &stock})
	if err != nil {
		t.Fatal(err)
	}
	linea := []ventas.LineaCarrito{{ID: p.ID, Cantidad: 2}}
	envejecer := func(email string, horas int) {
		t.Helper()
		_ = b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `UPDATE carritos SET actualizado = actualizado - make_interval(hours => $2),
				recordado_en = recordado_en - make_interval(hours => $2) WHERE email = $1`, email, horas)
			return err
		})
	}
	pasada := func() {
		t.Helper()
		if _, err := w.Pasada(ctx); err != nil {
			t.Fatal(err)
		}
	}

	// Por defecto: 1 h, 24 h y 72 h, apagados. Validaciones.
	r, err := ventas.LeerRecordatorios(ctx, b.App, luna)
	if err != nil || r.Activo || r.Pasos[0].Horas != 1 || r.Pasos[1].Horas != 24 || r.Pasos[2].Horas != 72 {
		t.Fatalf("por defecto: %+v %v", r, err)
	}
	var mal ventas.ErrRecordatorios
	malos := []func(*ventas.Recordatorios){
		func(r *ventas.Recordatorios) { r.Pasos[1].Horas = 1 },
		func(r *ventas.Recordatorios) { r.Pasos[2].Horas = 500 },
		func(r *ventas.Recordatorios) {
			r.Pasos[0].Cupon = &ventas.CuponPaso{Descuento: "porcentaje", Valor: 10, Dias: 3}
		},
		func(r *ventas.Recordatorios) {
			r.Pasos[1].Cupon = &ventas.CuponPaso{Descuento: "porcentaje", Valor: 95, Dias: 3}
		},
	}
	for i, f := range malos {
		x := r
		f(&x)
		if err := ventas.GuardarRecordatorios(ctx, b.App, luna, x); !errors.As(err, &mal) {
			t.Fatalf("caso %d: %v", i, err)
		}
	}
	r.Activo = true
	r.Pasos[1].Cupon = &ventas.CuponPaso{Descuento: "porcentaje", Valor: 10, Dias: 3}
	if err := ventas.GuardarRecordatorios(ctx, b.App, luna, r); err != nil {
		t.Fatal(err)
	}
	if a, _ := ventas.LeerAutomaticos(ctx, b.App, luna); !a.Email {
		t.Fatal("las apps viejas no ven el interruptor prendido")
	}

	// Ana deja el carrito: a la hora, el primero; a las 24 h, el segundo con un cupón único.
	if err := ventas.GuardarCarrito(ctx, b.App, luna, "ana@ejemplo.com", "Ana Paz", linea); err != nil {
		t.Fatal(err)
	}
	if n, _ := ventas.RecordatoriosAutomaticos(ctx, b.App, time.Now()); n != 0 {
		t.Fatalf("antes de la hora: %d", n)
	}
	envejecer("ana@ejemplo.com", 2)
	if n, err := ventas.RecordatoriosAutomaticos(ctx, b.App, time.Now()); err != nil || n != 1 {
		t.Fatalf("primero: %d %v", n, err)
	}
	if n, _ := ventas.RecordatoriosAutomaticos(ctx, b.App, time.Now()); n != 0 {
		t.Fatal("mandó dos veces el primero")
	}
	pasada()
	m, _ := mails.Ultimo("ana@ejemplo.com")
	if !strings.Contains(m.Asunto, "¿Te olvidaste algo") || !strings.Contains(m.Texto, "2 × Mate") ||
		m.Encabezados["List-Unsubscribe-Post"] != "List-Unsubscribe=One-Click" || !strings.Contains(m.Texto, "/t/luna/carrito/baja/") {
		t.Fatalf("primero: %+v", m)
	}
	envejecer("ana@ejemplo.com", 23)
	if n, _ := ventas.RecordatoriosAutomaticos(ctx, b.App, time.Now()); n != 1 {
		t.Fatalf("segundo: %d", n)
	}
	pasada()
	m, _ = mails.Ultimo("ana@ejemplo.com")
	var codigo string
	_ = b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT codigo FROM cupones WHERE origen LIKE 'carrito:%'`).Scan(&codigo)
	})
	if codigo == "" || !strings.Contains(m.Texto, "?cupon="+codigo) || !strings.Contains(m.Asunto, "10 % off") {
		t.Fatalf("segundo con cupón %q: %+v", codigo, m)
	}

	// Compra con el cupón: se corta (el tercero no sale) y cuenta como recuperado, con lo cobrado.
	c := vidriera.Compra{Email: "ana@ejemplo.com", Nombre: "Ana Paz", Telefono: "11 5555-0101", CP: "1414",
		Direccion: "Gorriti 4500", Envio: "retiro", Pago: "transferencia", Cupon: codigo,
		Carrito: []vidriera.LineaCarrito{{ID: p.ID, Cantidad: 2}}}
	res, err := vidriera.Comprar(ctx, b.App, nil, "", "luna", c, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	envejecer("ana@ejemplo.com", 72)
	if n, _ := ventas.RecordatoriosAutomaticos(ctx, b.App, time.Now()); n != 0 {
		t.Fatalf("recordó después de comprar: %d", n)
	}
	_, _, _, rec, _, err := ventas.CarritosYRecuperados(ctx, b.App, luna, time.Now())
	if err != nil || rec.Pedidos != 1 || rec.Monto != int64(res.Total) || res.Total >= 20000 {
		t.Fatalf("recuperados: %+v %+v %v", rec, res, err)
	}

	// Beto se da de baja desde el primero: no le llega nada más, tampoco con un carrito nuevo.
	if err := ventas.GuardarCarrito(ctx, b.App, luna, "beto@ejemplo.com", "Beto", linea); err != nil {
		t.Fatal(err)
	}
	envejecer("beto@ejemplo.com", 2)
	if n, _ := ventas.RecordatoriosAutomaticos(ctx, b.App, time.Now()); n != 1 {
		t.Fatalf("beto: %d", n)
	}
	pasada()
	m, _ = mails.Ultimo("beto@ejemplo.com")
	token := strings.Fields(m.Texto[strings.Index(m.Texto, "/baja/")+len("/baja/"):])[0]
	if err := ventas.BajaCarrito(ctx, b.App, luna, token); err != nil {
		t.Fatal(err)
	}
	envejecer("beto@ejemplo.com", 30)
	if n, _ := ventas.RecordatoriosAutomaticos(ctx, b.App, time.Now()); n != 0 {
		t.Fatalf("recordó después de la baja: %d", n)
	}
}
