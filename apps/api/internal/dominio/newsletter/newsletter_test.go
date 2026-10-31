package newsletter_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/configuracion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/newsletter"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

const sitio = "https://tulocalweb.com.ar"

// tokenDe saca el token del link del mail (confirmar o baja).
func tokenDe(t *testing.T, texto, ruta string) string {
	t.Helper()
	i := strings.Index(texto, ruta)
	if i < 0 {
		t.Fatalf("sin %q en %q", ruta, texto)
	}
	return strings.Fields(texto[i+len(ruta):])[0]
}

func envio(p newsletter.Panel, id string) newsletter.Envio {
	for _, e := range p.Envios {
		if e.ID == id {
			return e
		}
	}
	return newsletter.Envio{}
}

func pasar(t *testing.T, w *outbox.Worker) {
	t.Helper()
	for range 20 {
		n, err := w.Pasada(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if n == 0 {
			return
		}
	}
}

func TestSuscripcionConDobleConfirmacionYBaja(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	sol := pruebadb.CrearTienda(t, b.App, "sol")
	if err := configuracion.GuardarNegocio(ctx, b.App, luna, configuracion.Negocio{Nombre: "Luna", Email: "hola@luna.com.ar"}); err != nil {
		t.Fatal(err)
	}
	mails := &correo.Memoria{}
	w := &outbox.Worker{Pool: b.Worker, Manejadores: (&newsletter.Mails{Pool: b.App, Correo: mails, Sitio: sitio}).Manejadores()}

	if err := newsletter.Suscribir(ctx, b.App, luna, "no-es-un-email", "tienda"); !errors.Is(err, newsletter.ErrInvalido) {
		t.Fatalf("email inválido: %v", err)
	}
	if err := newsletter.Suscribir(ctx, b.App, luna, " Lucia@Ejemplo.com ", "tienda"); err != nil {
		t.Fatal(err)
	}
	pasar(t, w)
	m, ok := mails.Ultimo("lucia@ejemplo.com")
	if !ok || !strings.Contains(m.Asunto, "Confirmá") || m.DeNombre != "Luna" || m.ResponderA != "hola@luna.com.ar" {
		t.Fatalf("confirmación: %+v", m)
	}
	token := tokenDe(t, m.Texto, "/t/luna/newsletter/confirmar/")
	p, err := newsletter.Leer(ctx, b.App, luna, time.Now())
	if err != nil || p.Pendientes != 1 || p.Confirmados != 0 || len(p.Suscriptores) != 1 || p.Suscriptores[0].Origen != "tienda" {
		t.Fatalf("pendiente: %+v %v", p, err)
	}
	// Pedirlo de nuevo sin confirmar vuelve a mandar el mismo link.
	if err := newsletter.Suscribir(ctx, b.App, luna, "lucia@ejemplo.com", "checkout"); err != nil {
		t.Fatal(err)
	}
	pasar(t, w)
	if len(mails.Enviados) != 2 || tokenDe(t, mails.Enviados[1].Texto, "/confirmar/") != token {
		t.Fatalf("reenvío: %d mails", len(mails.Enviados))
	}
	// El token de una tienda no confirma en otra.
	if err := newsletter.Confirmar(ctx, b.App, sol, token); !errors.Is(err, newsletter.ErrNoExiste) {
		t.Fatalf("otra tienda: %v", err)
	}
	if err := newsletter.Confirmar(ctx, b.App, luna, "x"); !errors.Is(err, newsletter.ErrNoExiste) {
		t.Fatalf("token inválido: %v", err)
	}
	if err := newsletter.Confirmar(ctx, b.App, luna, token); err != nil {
		t.Fatal(err)
	}
	// Ya confirmada: suscribirse otra vez no manda nada.
	if err := newsletter.Suscribir(ctx, b.App, luna, "lucia@ejemplo.com", "tienda"); err != nil {
		t.Fatal(err)
	}
	pasar(t, w)
	if len(mails.Enviados) != 2 {
		t.Fatalf("mandó otra confirmación a una confirmada: %d", len(mails.Enviados))
	}
	if p, _ := newsletter.Leer(ctx, b.App, luna, time.Now()); p.Confirmados != 1 || !p.Suscriptores[0].Confirmado {
		t.Fatalf("confirmada: %+v", p)
	}
	if p, _ := newsletter.Leer(ctx, b.App, sol, time.Now()); len(p.Suscriptores) != 0 {
		t.Fatal("otra tienda ve suscriptos ajenos")
	}
	// La baja borra la fila; el link ya no sirve.
	if err := newsletter.Baja(ctx, b.App, luna, token); err != nil {
		t.Fatal(err)
	}
	if err := newsletter.Baja(ctx, b.App, luna, token); !errors.Is(err, newsletter.ErrNoExiste) {
		t.Fatalf("segunda baja: %v", err)
	}
	if p, _ := newsletter.Leer(ctx, b.App, luna, time.Now()); len(p.Suscriptores) != 0 {
		t.Fatalf("después de la baja: %+v", p)
	}
}

func TestMandarNovedadDeATandasConTopeDiario(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	sol := pruebadb.CrearTienda(t, b.App, "sol")
	mails := &correo.Memoria{}
	ahora := time.Date(2026, 10, 30, 15, 0, 0, 0, time.UTC)
	nl := &newsletter.Mails{Pool: b.App, Correo: mails, Sitio: sitio, Ahora: func() time.Time { return ahora }}
	w := &outbox.Worker{Pool: b.Worker, Manejadores: nl.Manejadores()}
	stock := 3
	foto := "https://fotos.ejemplo.com/mate.jpg"
	mate, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Mate imperial", Precio: 25000, Stock: &stock, Foto: &foto})
	if err != nil {
		t.Fatal(err)
	}
	ajeno, err := listas.Guardar(ctx, b.App, sol, "", listas.Editable{Nombre: "Termo", Precio: 1000, Stock: &stock})
	if err != nil {
		t.Fatal(err)
	}
	n := newsletter.Novedad{Asunto: "Llegaron los mates", Texto: "Hola.\n\nYa están los mates nuevos.", Productos: []string{mate.ID}}

	if _, err := newsletter.Mandar(ctx, b.App, luna, n, ahora); !errors.Is(err, newsletter.ErrSinDestinatarios) {
		t.Fatalf("sin suscriptos: %v", err)
	}
	var c newsletter.ErrCampo
	if _, err := newsletter.Mandar(ctx, b.App, luna, newsletter.Novedad{Asunto: "Hola", Texto: "corto"}, ahora); !errors.As(err, &c) || c.Campo != "texto" {
		t.Fatalf("texto corto: %v", err)
	}
	// Tres confirmados y uno pendiente (al pendiente no le llega).
	for _, e := range []string{"ana@ejemplo.com", "bea@ejemplo.com", "cami@ejemplo.com", "dani@ejemplo.com"} {
		if err := newsletter.Suscribir(ctx, b.App, luna, e, "checkout"); err != nil {
			t.Fatal(err)
		}
	}
	pasar(t, w)
	for _, e := range []string{"ana@ejemplo.com", "bea@ejemplo.com", "cami@ejemplo.com"} {
		m, _ := mails.Ultimo(e)
		if err := newsletter.Confirmar(ctx, b.App, luna, tokenDe(t, m.Texto, "/confirmar/")); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := newsletter.Mandar(ctx, b.App, luna, newsletter.Novedad{Asunto: "Hola", Texto: "Un texto largo", Productos: []string{ajeno.ID}}, ahora); !errors.As(err, &c) || c.Campo != "productos" {
		t.Fatalf("producto de otra tienda: %v", err)
	}

	// Hoy ya salieron 499 (de otro envío): entra uno y el resto sigue mañana a las 9.
	err = b.App.EnTienda(ctx, luna, func(x pgx.Tx) error {
		var viejo string
		if err := x.QueryRow(ctx, `INSERT INTO newsletter_envios (tienda_id, asunto, texto, destinatarios, creado, terminado)
			VALUES ($1, 'Anterior', 'Una novedad anterior', 499, $2, $2) RETURNING id::text`, luna, ahora).Scan(&viejo); err != nil {
			return err
		}
		_, err := x.Exec(ctx, `INSERT INTO newsletter_entregas (envio_id, tienda_id, email, estado, enviado_en)
			SELECT $1, $2, 'viejo' || g || '@ejemplo.com', 'enviado', $3 FROM generate_series(1, 499) g`, viejo, luna, ahora)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	antes := len(mails.Enviados)
	e, err := newsletter.Mandar(ctx, b.App, luna, n, ahora)
	if err != nil || e.Destinatarios != 3 {
		t.Fatalf("mandar: %+v %v", e, err)
	}
	pasar(t, w)
	if len(mails.Enviados) != antes+1 {
		t.Fatalf("con el tope: salieron %d", len(mails.Enviados)-antes)
	}
	m := mails.Enviados[antes]
	if m.Para != "ana@ejemplo.com" || m.Asunto != "Llegaron los mates" || m.DeNombre != "Tienda luna" ||
		!strings.Contains(m.Texto, "Mate imperial — $ 25.000") || !strings.Contains(m.HTML, foto) ||
		!strings.Contains(m.Texto, "/t/luna/producto/"+mate.ID) {
		t.Fatalf("novedad: %+v", m)
	}
	baja := tokenDe(t, m.Texto, "/t/luna/newsletter/baja/")
	if m.Encabezados["List-Unsubscribe"] != "<"+sitio+"/t/luna/newsletter/baja/"+baja+">" ||
		m.Encabezados["List-Unsubscribe-Post"] != "List-Unsubscribe=One-Click" {
		t.Fatalf("encabezados: %v", m.Encabezados)
	}
	p, _ := newsletter.Leer(ctx, b.App, luna, ahora)
	if p.Hoy != 500 || len(p.Envios) != 2 || envio(p, e.ID).Estado != "manana" || envio(p, e.ID).Enviados != 1 {
		t.Fatalf("panel con el tope: %+v", p)
	}
	var cuando time.Time
	if err := b.Worker.QueryRow(ctx, `SELECT proximo_intento FROM outbox WHERE tipo = 'newsletter.tanda' AND publicado_en IS NULL`).Scan(&cuando); err != nil {
		t.Fatal(err)
	}
	if want := time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC); !cuando.Equal(want) {
		t.Fatalf("sigue %v, quería %v", cuando, want)
	}

	// Bea se da de baja antes de que le llegue; al otro día sale el resto.
	m2, _ := mails.Ultimo("bea@ejemplo.com")
	if err := newsletter.Baja(ctx, b.App, luna, tokenDe(t, m2.Texto, "/confirmar/")); err != nil {
		t.Fatal(err)
	}
	ahora = ahora.Add(24 * time.Hour)
	if _, err := b.Worker.Exec(ctx, `UPDATE outbox SET proximo_intento = now() WHERE tipo = 'newsletter.tanda' AND publicado_en IS NULL`); err != nil {
		t.Fatal(err)
	}
	pasar(t, w)
	if len(mails.Enviados) != antes+2 || mails.Enviados[antes+1].Para != "cami@ejemplo.com" {
		t.Fatalf("al otro día: %d", len(mails.Enviados)-antes)
	}
	p, _ = newsletter.Leer(ctx, b.App, luna, ahora)
	if x := envio(p, e.ID); x.Estado != "enviado" || x.Enviados != 2 || x.Destinatarios != 3 || p.Hoy != 1 {
		t.Fatalf("terminado: %+v hoy %d", x, p.Hoy)
	}
	// Borrar desde el panel.
	if err := newsletter.Borrar(ctx, b.App, sol, p.Suscriptores[0].ID); !errors.Is(err, newsletter.ErrNoExiste) {
		t.Fatalf("borrar de otra tienda: %v", err)
	}
	if err := newsletter.Borrar(ctx, b.App, luna, p.Suscriptores[0].ID); err != nil {
		t.Fatal(err)
	}
}
