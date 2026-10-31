package lanzamientos_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/lanzamientos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/vidriera"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

var ctx = context.Background()

func ptr[T any](v T) *T { return &v }

// senadorFalso hace de Mercado Pago para la seña (el de verdad se prueba en el paquete mercadopago).
type senadorFalso struct {
	mu      sync.Mutex
	falla   bool
	pedidos []int
}

func (s *senadorFalso) PagarSena(_ context.Context, _ string, numero int, vuelta string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.falla {
		return "", errors.New("mercado pago caído")
	}
	s.pedidos = append(s.pedidos, numero)
	return "https://www.mercadopago.com.ar/checkout/v1/redirect?pref_id=x&vuelta=" + vuelta, nil
}

func conectarMP(t *testing.T, app *db.Pool, tienda string) {
	t.Helper()
	if err := app.EnTienda(ctx, tienda, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO mp_conexiones (tienda_id, mp_usuario, acceso, renovacion, vence_en)
			VALUES ($1, 1000 + (random() * 1e6)::int, '\x00', '\x00', now() + interval '1 day')`, tienda)
		return err
	}); err != nil {
		t.Fatal(err)
	}
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

func publico(t *testing.T, b *pruebadb.Bases, id string) vidriera.ProductoPublico {
	t.Helper()
	tienda, err := vidriera.Publica(ctx, b.App, "luna")
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range tienda.Productos {
		if p.ID == id {
			return p
		}
	}
	t.Fatalf("no está publicado %s", id)
	return vidriera.ProductoPublico{}
}

func TestLanzamientoProximamenteYAvisos(t *testing.T) {
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	buzo, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Buzo Heavy", Precio: 50000, Stock: ptr(5)})
	if err != nil {
		t.Fatal(err)
	}
	ahora := time.Now()
	sale := ahora.Add(2 * time.Hour).Truncate(time.Second)

	// Validaciones: fecha pasada, reserva sin lanzamiento, oferta más cara que el precio.
	casos := []lanzamientos.Config{
		{Lanzamiento: ptr(ahora.Add(-time.Hour))},
		{Reserva: ptr(30)},
		{Lanzamiento: &sale, Reserva: ptr(0)},
		{OfertaPrecio: ptr(int64(60000))},
		{OfertaHasta: ptr(ahora.Add(time.Hour))},
	}
	for i, c := range casos {
		var inv lanzamientos.ErrInvalido
		if _, err := lanzamientos.Guardar(ctx, b.App, luna, buzo.ID, c, ahora); !errors.As(err, &inv) {
			t.Fatalf("caso %d: %v", i, err)
		}
	}
	e, err := lanzamientos.Guardar(ctx, b.App, luna, buzo.ID, lanzamientos.Config{Lanzamiento: &sale, Reserva: ptr(30)}, ahora)
	if err != nil || e.Lanzamiento == nil || !e.Lanzamiento.Equal(sale) || *e.Reserva != 30 {
		t.Fatalf("guardar: %+v %v", e, err)
	}

	// En la tienda: «Próximamente», con la fecha, y sin reserva hasta que haya Mercado Pago.
	p := publico(t, b, buzo.ID)
	if p.Lanzamiento == nil || p.Etiqueta == nil || *p.Etiqueta != "Próximamente" || p.Reserva != nil {
		t.Fatalf("antes del lanzamiento: %+v", p)
	}
	conectarMP(t, b.App, luna)
	if p := publico(t, b, buzo.ID); p.Reserva == nil || *p.Reserva != 30 {
		t.Fatalf("con Mercado Pago, la reserva: %+v", p)
	}
	// No se compra antes de la hora.
	compra := vidriera.Compra{Email: "lu@ejemplo.com", Nombre: "Lucía", Envio: "retiro", Pago: "mercadopago",
		Carrito: []vidriera.LineaCarrito{{ID: buzo.ID, Cantidad: 1}}}
	if _, err := vidriera.Comprar(ctx, b.App, nil, "", "luna", compra, ahora); !errors.Is(err, vidriera.ErrInvalido) && !errors.Is(err, vidriera.ErrProximamente) {
		t.Fatalf("comprar antes: %v", err)
	}

	// Avisos: repetido no suma, mal escrito no entra.
	for _, m := range []string{"Ana@ejemplo.com", "ana@ejemplo.com ", "beto@ejemplo.com"} {
		if err := lanzamientos.Avisar(ctx, b.App, "luna", buzo.ID, m, ahora); err != nil {
			t.Fatal(err)
		}
	}
	var inv lanzamientos.ErrInvalido
	if err := lanzamientos.Avisar(ctx, b.App, "luna", buzo.ID, "no-es-un-mail", ahora); !errors.As(err, &inv) {
		t.Fatalf("email inválido: %v", err)
	}
	for i := range 60 {
		if err := lanzamientos.Avisar(ctx, b.App, "luna", buzo.ID, fmt.Sprintf("p%02d@ejemplo.com", i), ahora); err != nil {
			t.Fatal(err)
		}
	}
	if e, _ := lanzamientos.Leer(ctx, b.App, luna, buzo.ID); e.Avisos != 62 || e.Avisados != 0 {
		t.Fatalf("avisos: %+v", e)
	}

	// Antes de la hora no hay nada que avisar.
	if n, err := lanzamientos.PorLanzar(ctx, b.App); err != nil || n != 0 {
		t.Fatalf("por lanzar antes: %d %v", n, err)
	}
	// Llega la hora: un evento, que manda de a 50 y encola el resto.
	admin(t, b, `UPDATE productos SET lanzamiento = now() - interval '1 minute' WHERE id = $1`, buzo.ID)
	if err := lanzamientos.Avisar(ctx, b.App, "luna", buzo.ID, "tarde@ejemplo.com", time.Now()); !errors.Is(err, lanzamientos.ErrYaSalio) {
		t.Fatalf("aviso después de salir: %v", err)
	}
	if n, err := lanzamientos.PorLanzar(ctx, b.App); err != nil || n != 1 {
		t.Fatalf("por lanzar: %d %v", n, err)
	}
	if n, _ := lanzamientos.PorLanzar(ctx, b.App); n != 0 {
		t.Fatal("avisó dos veces el mismo lanzamiento")
	}
	mails := &correo.Memoria{}
	s := &lanzamientos.Servicio{Pool: b.App, Correo: mails, Sitio: "https://tulocalweb.com.ar"}
	w := &outbox.Worker{Pool: b.Worker, Manejadores: s.Manejadores()}
	if _, err := w.Pasada(ctx); err != nil {
		t.Fatal(err)
	}
	if len(mails.Enviados) != lanzamientos.Tanda {
		t.Fatalf("primera tanda: %d", len(mails.Enviados))
	}
	if _, err := w.Pasada(ctx); err != nil {
		t.Fatal(err)
	}
	enviados := mails.Enviados
	if len(enviados) != 62 || !strings.Contains(enviados[0].Asunto, "Ya salió Buzo Heavy") ||
		!strings.Contains(enviados[0].Texto, "https://tulocalweb.com.ar/t/luna/producto/") {
		t.Fatalf("todas las tandas: %d %+v", len(enviados), enviados[0])
	}
	if e, _ := lanzamientos.Leer(ctx, b.App, luna, buzo.ID); e.Avisados != 62 {
		t.Fatalf("avisados: %+v", e)
	}
	// Ya salió: se compra, sin «Próximamente».
	if p := publico(t, b, buzo.ID); p.Lanzamiento != nil || p.Reserva != nil {
		t.Fatalf("después del lanzamiento: %+v", p)
	}
}

func TestReservaConSenaYCupo(t *testing.T) {
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	conectarMP(t, b.App, luna)
	buzo, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Buzo Heavy", Precio: 50000, Stock: ptr(2)})
	if err != nil {
		t.Fatal(err)
	}
	ahora := time.Now()
	if _, err := lanzamientos.Guardar(ctx, b.App, luna, buzo.ID, lanzamientos.Config{Lanzamiento: ptr(ahora.Add(time.Hour)), Reserva: ptr(25)}, ahora); err != nil {
		t.Fatal(err)
	}
	mp := &senadorFalso{}
	pedir := func(nombre string) (vidriera.ResultadoSena, error) {
		return vidriera.Senar(ctx, b.App, mp, "https://tulocalweb.com.ar", "luna", vidriera.PedidoSena{ProductoID: buzo.ID,
			Nombre: nombre, Email: strings.ToLower(nombre) + "@ejemplo.com", Telefono: "1155550101"}, ahora)
	}
	r, err := pedir("Lucia")
	if err != nil || r.Sena != 12500 || r.Total != 50000 || !strings.Contains(r.URL, "vuelta=https://tulocalweb.com.ar/t/luna/pedido/"+r.Codigo) {
		t.Fatalf("reservar: %+v %v", r, err)
	}
	d, err := pedidos.Obtener(ctx, b.App, luna, r.Numero)
	if err != nil || d.Estado != "Sin pagar" || d.EstadoPago != "Pendiente" || d.Sena == nil || d.Sena.Monto != 12500 ||
		d.Sena.Falta != 37500 || d.Sena.Pagada || d.Sena.Tipo != "reserva" || d.Sena.Vence == nil {
		t.Fatalf("el pedido de la reserva: %+v %+v %v", d, d.Sena, err)
	}
	if e, _ := lanzamientos.Leer(ctx, b.App, luna, buzo.ID); e.Reservas != 1 {
		t.Fatalf("reservas: %+v", e)
	}
	// El cupo es el stock: la tercera no entra.
	if _, err := pedir("Beto"); err != nil {
		t.Fatal(err)
	}
	if _, err := pedir("Caro"); !errors.Is(err, vidriera.ErrSinStock) {
		t.Fatalf("pasó el cupo: %v", err)
	}
	// La seña de Beto no se paga a tiempo: el worker la cancela y el stock vuelve.
	admin(t, b, `UPDATE pedidos SET sena_vence = now() - interval '1 minute' WHERE numero = $1`, r.Numero+1)
	if n, err := lanzamientos.VencerSenas(ctx, b.App); err != nil || n != 1 {
		t.Fatalf("vencer: %d %v", n, err)
	}
	if p := publico(t, b, buzo.ID); p.Stock != 1 {
		t.Fatalf("el stock no volvió: %d", p.Stock)
	}
	// Mercado Pago caído: el pedido no queda tomando el stock.
	mp.falla = true
	if _, err := pedir("Dani"); !errors.Is(err, vidriera.ErrMercadoPago) {
		t.Fatalf("con Mercado Pago caído: %v", err)
	}
	if p := publico(t, b, buzo.ID); p.Stock != 1 {
		t.Fatalf("el stock quedó tomado: %d", p.Stock)
	}
	mp.falla = false

	// La seña pagada (lo que hace el webhook) y «saldado» desde el panel.
	admin(t, b, `UPDATE pedidos SET sena_pagada_en = now(), sena_vence = NULL, estado = 'Para preparar', estado_pago = 'Seña' WHERE numero = $1`, r.Numero)
	if err := pedidos.Saldar(ctx, b.App, luna, r.Numero+1, true); !errors.Is(err, pedidos.ErrSinSena) {
		t.Fatalf("saldar uno cancelado: %v", err)
	}
	if err := pedidos.Saldar(ctx, b.App, luna, r.Numero, true); err != nil {
		t.Fatal(err)
	}
	if d, _ := pedidos.Obtener(ctx, b.App, luna, r.Numero); !d.Sena.Saldado || d.Sena.Falta != 0 || d.EstadoPago != "Cobrado" {
		t.Fatalf("saldado: %+v %+v", d, d.Sena)
	}
	if err := pedidos.Saldar(ctx, b.App, luna, r.Numero, false); err != nil {
		t.Fatal(err)
	}
	if d, _ := pedidos.Obtener(ctx, b.App, luna, r.Numero); d.Sena.Saldado || d.EstadoPago != "Seña" {
		t.Fatalf("deshacer: %+v", d)
	}
	// Ya lanzado y sin encargo, no hay seña.
	admin(t, b, `UPDATE productos SET lanzamiento = now() - interval '1 minute' WHERE id = $1`, buzo.ID)
	if _, err := pedir("Eli"); !errors.Is(err, vidriera.ErrSinSena) {
		t.Fatalf("seña sin encargo: %v", err)
	}
}

func TestEncargoConVarianteYOfertaConFin(t *testing.T) {
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	conectarMP(t, b.App, luna)
	mesa, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Mesa Tala", Precio: 600000,
		Opciones: []listas.Opcion{{Nombre: "Madera", Valores: []string{"Paraíso", "Quebracho"}}}})
	if err != nil {
		t.Fatal(err)
	}
	var paraiso, quebracho string
	for _, v := range mesa.Variantes {
		if v.Valores["Madera"] == "Paraíso" {
			paraiso = v.ID
		} else {
			quebracho = v.ID
		}
	}
	if _, err := listas.Guardar(ctx, b.App, luna, mesa.ID, listas.Editable{Nombre: "Mesa Tala", Precio: 600000,
		Variantes: []listas.VarianteEditable{{Valores: map[string]string{"Madera": "Paraíso"}, Stock: ptr(3)},
			{Valores: map[string]string{"Madera": "Quebracho"}, Stock: ptr(3), Precio: ptr(int64(700000))}}}); err != nil {
		t.Fatal(err)
	}
	ahora := time.Now()
	fin := ahora.Add(48 * time.Hour)
	if _, err := lanzamientos.Guardar(ctx, b.App, luna, mesa.ID, lanzamientos.Config{Encargo: ptr(50), OfertaPrecio: ptr(int64(540000)), OfertaHasta: &fin}, ahora); err != nil {
		t.Fatal(err)
	}
	p := publico(t, b, mesa.ID)
	if p.Precio != 540000 || p.Anterior == nil || *p.Anterior != 600000 || p.OfertaHasta == nil || p.Encargo == nil || *p.Encargo != 50 {
		t.Fatalf("oferta y encargo: %+v", p)
	}
	mp := &senadorFalso{}
	// Con opciones hace falta la variante; la de precio propio no entra en la oferta.
	if _, err := vidriera.Senar(ctx, b.App, mp, "", "luna", vidriera.PedidoSena{ProductoID: mesa.ID, Nombre: "Ana", Email: "ana@ejemplo.com"}, ahora); !errors.Is(err, vidriera.ErrInvalido) {
		t.Fatalf("sin variante: %v", err)
	}
	r, err := vidriera.Senar(ctx, b.App, mp, "", "luna", vidriera.PedidoSena{ProductoID: mesa.ID, VarianteID: paraiso, Nombre: "Ana", Email: "ana@ejemplo.com"}, ahora)
	if err != nil || r.Total != 540000 || r.Sena != 270000 {
		t.Fatalf("encargo con oferta: %+v %v", r, err)
	}
	d, _ := pedidos.Obtener(ctx, b.App, luna, r.Numero)
	if d.Items[0].Precio != 540000 || d.Sena.Tipo != "encargo" {
		t.Fatalf("el pedido guarda el precio de oferta: %+v", d.Items)
	}
	r2, err := vidriera.Senar(ctx, b.App, mp, "", "luna", vidriera.PedidoSena{ProductoID: mesa.ID, VarianteID: quebracho, Cantidad: 2, Nombre: "Ana", Email: "ana@ejemplo.com"}, ahora)
	if err != nil || r2.Total != 1400000 || r2.Sena != 700000 {
		t.Fatalf("variante con precio propio: %+v %v", r2, err)
	}
	// La compra común también cobra la oferta.
	if err := b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE tiendas SET negocio = negocio || '{"alias":"luna.mp"}' WHERE id = $1`, luna)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	c, err := vidriera.Comprar(ctx, b.App, nil, "", "luna", vidriera.Compra{Email: "beto@ejemplo.com", Nombre: "Beto", Envio: "retiro",
		Pago: "transferencia", Carrito: []vidriera.LineaCarrito{{ID: mesa.ID, VarianteID: paraiso, Cantidad: 1}}}, ahora)
	if err != nil || c.Total != 540000 {
		t.Fatalf("compra con oferta: %+v %v", c, err)
	}
	// Terminada la oferta, vuelve el precio de lista.
	admin(t, b, `UPDATE productos SET oferta_hasta = now() - interval '1 minute' WHERE id = $1`, mesa.ID)
	if p := publico(t, b, mesa.ID); p.Precio != 600000 || p.Anterior != nil || p.OfertaHasta != nil {
		t.Fatalf("después de la oferta: %+v", p)
	}
}
