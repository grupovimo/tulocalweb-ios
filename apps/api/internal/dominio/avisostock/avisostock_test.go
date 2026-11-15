package avisostock_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/avisostock"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

var ctx = context.Background()

func ptr[T any](v T) *T { return &v }

func esperan(t *testing.T, b *pruebadb.Bases, tienda, id string) int {
	t.Helper()
	regs, err := listas.Productos(ctx, b.App, tienda, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range regs {
		if r["id"] == id {
			return r["esperan"].(int)
		}
	}
	t.Fatalf("no está %s", id)
	return 0
}

func admin(t *testing.T, b *pruebadb.Bases, sql string, args ...any) {
	t.Helper()
	conn, err := pgx.Connect(ctx, b.AdminURL)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(ctx)
	if _, err := conn.Exec(ctx, sql, args...); err != nil {
		t.Fatal(err)
	}
}

func TestAvisameCuandoVuelva(t *testing.T) {
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	pruebadb.CrearTienda(t, b.App, "sol")
	mate, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Mate Imperial", Precio: 25000, Stock: ptr(0)})
	if err != nil {
		t.Fatal(err)
	}
	lleno, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Bombilla", Precio: 5000, Stock: ptr(4)})
	if err != nil {
		t.Fatal(err)
	}
	ahora := time.Now()
	pedido := func(email string) avisostock.Pedido {
		return avisostock.Pedido{Email: email, Consentimiento: true}
	}

	// Validaciones: email, WhatsApp, consentimiento, con stock, de otra tienda.
	var inv avisostock.ErrInvalido
	if err := avisostock.Anotar(ctx, b.App, "luna", mate.ID, pedido("no-es-mail"), ahora); !errors.As(err, &inv) {
		t.Fatalf("email: %v", err)
	}
	if err := avisostock.Anotar(ctx, b.App, "luna", mate.ID, avisostock.Pedido{Email: "a@b.com", Whatsapp: "hola", Consentimiento: true}, ahora); !errors.As(err, &inv) {
		t.Fatalf("whatsapp: %v", err)
	}
	if err := avisostock.Anotar(ctx, b.App, "luna", mate.ID, avisostock.Pedido{Email: "a@b.com"}, ahora); !errors.As(err, &inv) {
		t.Fatalf("sin consentimiento: %v", err)
	}
	if err := avisostock.Anotar(ctx, b.App, "luna", lleno.ID, pedido("a@b.com"), ahora); !errors.Is(err, avisostock.ErrHayStock) {
		t.Fatalf("con stock: %v", err)
	}
	if err := avisostock.Anotar(ctx, b.App, "sol", mate.ID, pedido("a@b.com"), ahora); !errors.Is(err, avisostock.ErrNoExiste) {
		t.Fatalf("de otra tienda: %v", err)
	}

	// Repetido no suma (y el WhatsApp se guarda si llega después).
	for _, p := range []avisostock.Pedido{pedido("Ana@ejemplo.com"), {Email: "ana@ejemplo.com ", Whatsapp: "+54 11 5555-1234", Consentimiento: true}, pedido("beto@ejemplo.com")} {
		if err := avisostock.Anotar(ctx, b.App, "luna", mate.ID, p, ahora); err != nil {
			t.Fatal(err)
		}
	}
	if n := esperan(t, b, luna, mate.ID); n != 2 {
		t.Fatalf("esperan: %d", n)
	}
	var wa *string
	_ = b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT whatsapp FROM avisos_stock WHERE email = 'ana@ejemplo.com'`).Scan(&wa)
	})
	if wa == nil || *wa != "+54 11 5555-1234" {
		t.Fatalf("whatsapp: %v", wa)
	}

	mails := &correo.Memoria{}
	s := &avisostock.Servicio{Pool: b.App, Correo: mails, Sitio: "https://tulocalweb.com.ar"}
	w := &outbox.Worker{Pool: b.Worker, Manejadores: s.Manejadores()}
	pasada := func() {
		t.Helper()
		if _, err := w.Pasada(ctx); err != nil {
			t.Fatal(err)
		}
	}

	// Otra reposición (la bombilla, que nadie espera) no encola nada.
	if _, err := listas.Guardar(ctx, b.App, luna, lleno.ID, listas.Editable{Nombre: "Bombilla", Precio: 5000, Stock: ptr(9)}); err != nil {
		t.Fatal(err)
	}
	pasada()
	if len(mails.Enviados) != 0 {
		t.Fatalf("sin nadie esperando: %d", len(mails.Enviados))
	}

	// Vuelve el stock desde el editor: un mail a cada uno, con la marca y el link.
	if _, err := listas.Guardar(ctx, b.App, luna, mate.ID, listas.Editable{Nombre: "Mate Imperial", Precio: 25000, Stock: ptr(3)}); err != nil {
		t.Fatal(err)
	}
	pasada()
	if len(mails.Enviados) != 2 {
		t.Fatalf("avisos: %d", len(mails.Enviados))
	}
	m := mails.Enviados[0]
	if !strings.Contains(m.Asunto, "Volvió Mate Imperial") || !strings.Contains(m.Texto, "https://tulocalweb.com.ar/t/luna/producto/") ||
		!strings.Contains(m.Texto, "$ 25.000") {
		t.Fatalf("mail: %+v", m)
	}
	if n := esperan(t, b, luna, mate.ID); n != 0 {
		t.Fatalf("esperan después: %d", n)
	}

	// Una sola vez por pedido: se agota y vuelve (por cualquier camino, acá SQL directo) y no repite.
	admin(t, b, `UPDATE productos SET stock = 0 WHERE id = $1`, mate.ID)
	admin(t, b, `UPDATE productos SET stock = 5 WHERE id = $1`, mate.ID)
	pasada()
	if len(mails.Enviados) != 2 {
		t.Fatalf("repitió: %d", len(mails.Enviados))
	}

	// Variantes: el aviso de una variante sale cuando vuelve esa, no otra.
	remera, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Remera", Precio: 15000,
		Opciones: []listas.Opcion{{Nombre: "Talle", Valores: []string{"S", "M"}}}})
	if err != nil {
		t.Fatal(err)
	}
	var s1, m1 string
	_ = b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `SELECT id::text FROM variantes WHERE producto_id = $1 AND valores->>'Talle' = 'S'`, remera.ID).Scan(&s1); err != nil {
			return err
		}
		return tx.QueryRow(ctx, `SELECT id::text FROM variantes WHERE producto_id = $1 AND valores->>'Talle' = 'M'`, remera.ID).Scan(&m1)
	})
	if err := avisostock.Anotar(ctx, b.App, "luna", remera.ID, avisostock.Pedido{VarianteID: &m1, Email: "caro@ejemplo.com", Consentimiento: true}, ahora); err != nil {
		t.Fatal(err)
	}
	admin(t, b, `UPDATE variantes SET stock = 2 WHERE id = $1`, s1)
	pasada()
	if len(mails.Enviados) != 2 {
		t.Fatalf("avisó por otra variante: %d", len(mails.Enviados))
	}
	admin(t, b, `UPDATE variantes SET stock = 1 WHERE id = $1`, m1)
	pasada()
	if len(mails.Enviados) != 3 || !strings.Contains(mails.Enviados[2].Texto, "Remera (M)") {
		t.Fatalf("variante: %d %+v", len(mails.Enviados), mails.Enviados[len(mails.Enviados)-1])
	}
}
