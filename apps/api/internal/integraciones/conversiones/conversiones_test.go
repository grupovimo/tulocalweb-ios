package conversiones_test

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/configuracion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/vidriera"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/conversiones"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/cifrado"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

const (
	pixel = "123456789012345"
	token = "EAAGtokenDePruebaDeLaTienda_123"
)

// graphFalso hace de la Graph API: el token bueno lee el píxel; los eventos se guardan.
type graphFalso struct {
	mu       sync.Mutex
	eventos  []map[string]any
	responde int
}

func (g *graphFalso) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.mu.Lock()
	defer g.mu.Unlock()
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/"+conversiones.Version+"/"+pixel:
		if r.Header.Get("Authorization") != "Bearer "+token {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_, _ = io.WriteString(w, `{"id":"`+pixel+`"}`)
	case r.Method == http.MethodPost && r.URL.Path == "/"+conversiones.Version+"/"+pixel+"/events":
		var cuerpo map[string]any
		_ = json.NewDecoder(r.Body).Decode(&cuerpo)
		if cuerpo["access_token"] != token {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, `{"error":{"message":"Invalid OAuth access token."}}`)
			return
		}
		if g.responde != 0 {
			w.WriteHeader(g.responde)
			_, _ = io.WriteString(w, `{"error":{"message":"Se cayó."}}`)
			return
		}
		g.eventos = append(g.eventos, cuerpo)
		_, _ = io.WriteString(w, `{"events_received":1}`)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func cifrador(t *testing.T) *cifrado.Cifrador {
	k := make([]byte, 32)
	_, _ = rand.Read(k)
	c, err := cifrado.Nuevo(base64.StdEncoding.EncodeToString(k))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestComprasALaAPIDeConversiones(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	sol := pruebadb.CrearTienda(t, b.App, "sol")
	g := &graphFalso{}
	srv := httptest.NewServer(g)
	defer srv.Close()
	s := &conversiones.Servicio{Pool: b.App, Cifra: cifrador(t), HTTP: srv.Client(), Base: srv.URL}

	// Sin píxel no se conecta; con un token que Meta no acepta, tampoco.
	if err := s.Guardar(ctx, luna, conversiones.Conexion{Token: token}); !errors.Is(err, conversiones.ErrSinPixel) {
		t.Fatalf("sin píxel: %v", err)
	}
	p := pixel
	if err := configuracion.GuardarMedicion(ctx, b.App, luna, configuracion.Medicion{Pixel: &p}); err != nil {
		t.Fatal(err)
	}
	if err := s.Guardar(ctx, luna, conversiones.Conexion{Token: "EAAGotroTokenQueNoEsDeLaTienda"}); !errors.Is(err, conversiones.ErrTokenNoSirve) {
		t.Fatalf("token ajeno: %v", err)
	}
	if err := s.Guardar(ctx, luna, conversiones.Conexion{Token: token, CodigoPrueba: "no-es-un-codigo"}); !errors.Is(err, conversiones.ErrInvalido) {
		t.Fatalf("código de prueba: %v", err)
	}
	if err := s.Guardar(ctx, luna, conversiones.Conexion{Token: token, CodigoPrueba: "test12345"}); err != nil {
		t.Fatal(err)
	}
	e, err := s.Leer(ctx, luna)
	if err != nil || !e.Conectada || e.CodigoPrueba == nil || *e.CodigoPrueba != "TEST12345" {
		t.Fatalf("estado: %+v %v", e, err)
	}
	// El token queda cifrado (nunca en claro en la base) y otra tienda no ve nada.
	var guardado []byte
	_ = b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT token FROM meta_conversiones`).Scan(&guardado)
	})
	if strings.Contains(string(guardado), token) {
		t.Fatal("el token quedó en claro")
	}
	if e, _ := s.Leer(ctx, sol); e.Conectada {
		t.Fatal("otra tienda ve la conexión")
	}

	// Una compra con consentimiento se encola; sin consentimiento, no.
	stock := 10
	prod, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Mate", Precio: 10000, Stock: &stock})
	if err != nil {
		t.Fatal(err)
	}
	if err := configuracion.GuardarNegocio(ctx, b.App, luna, configuracion.Negocio{Nombre: "Luna", Alias: "luna.tienda"}); err != nil {
		t.Fatal(err)
	}
	comprar := func(medir bool) int {
		t.Helper()
		c := vidriera.Compra{Email: " Lucia@Ejemplo.com", Nombre: "Lucía Pérez", Telefono: "11 5555-0101", Envio: "retiro", Pago: "transferencia"}
		c.Carrito = append(c.Carrito, vidriera.LineaCarrito{ID: prod.ID, Cantidad: 2})
		if medir {
			c.Medicion = &conversiones.Navegador{IP: "181.1.2.3", UA: "Mozilla/5.0", FBP: "fb.1.1700000000000.123", URL: "https://tulocalweb.com.ar/t/luna/checkout"}
		}
		r, err := vidriera.Comprar(ctx, b.App, nil, "", "luna", c, time.Now())
		if err != nil {
			t.Fatal(err)
		}
		return r.Numero
	}
	pendientes := func() []outbox.Evento {
		t.Helper()
		filas, err := b.Worker.Query(ctx, `SELECT id, tienda_id::text, tipo, datos FROM outbox WHERE tipo = $1 ORDER BY id`, conversiones.TipoCompra)
		if err != nil {
			t.Fatal(err)
		}
		defer filas.Close()
		var es []outbox.Evento
		for filas.Next() {
			var e outbox.Evento
			if err := filas.Scan(&e.ID, &e.TiendaID, &e.Tipo, &e.Datos); err != nil {
				t.Fatal(err)
			}
			es = append(es, e)
		}
		return es
	}
	comprar(false)
	if n := len(pendientes()); n != 0 {
		t.Fatalf("sin consentimiento se encolaron %d", n)
	}
	numero := comprar(true)
	es := pendientes()
	if len(es) != 1 {
		t.Fatalf("encolados: %d", len(es))
	}
	if strings.Contains(string(es[0].Datos), "lucia@") || strings.Contains(string(es[0].Datos), "5555") {
		t.Fatalf("email o teléfono sin hashear en el outbox: %s", es[0].Datos)
	}

	// El worker lo manda: Purchase con el número de pedido (deduplica con el píxel), email hasheado,
	// el código de prueba y el valor.
	enviar := s.Manejadores()[conversiones.TipoCompra]
	if err := enviar(ctx, es[0]); err != nil {
		t.Fatal(err)
	}
	if len(g.eventos) != 1 {
		t.Fatalf("eventos en Meta: %d", len(g.eventos))
	}
	b2, _ := json.Marshal(g.eventos[0])
	cuerpo := string(b2)
	for _, quiere := range []string{`"event_name":"Purchase"`, `"event_id":"` + itoa(numero) + `"`, `"test_event_code":"TEST12345"`,
		`"em":["` + conversiones.HashEmail("lucia@ejemplo.com") + `"]`, `"ph":["` + conversiones.HashTelefono("+54 11 5555 0101") + `"]`,
		`"client_ip_address":"181.1.2.3"`, `"fbp":"fb.1.1700000000000.123"`, `"content_ids":["` + prod.ID + `"]`, `"currency":"ARS"`} {
		if !strings.Contains(cuerpo, quiere) {
			t.Fatalf("falta %s en %s", quiere, cuerpo)
		}
	}
	if e, _ := s.Leer(ctx, luna); e.UltimoEnvio == nil || e.UltimoError != nil {
		t.Fatalf("después de mandar: %+v", e)
	}

	// Meta caída: error para que el outbox reintente. Token vencido: no se reintenta y el panel lo ve.
	g.responde = http.StatusServiceUnavailable
	if err := enviar(ctx, es[0]); err == nil {
		t.Fatal("Meta caída y no se reintenta")
	}
	g.responde = 0
	_ = b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		c, _ := cifrador(t).Cifrar("meta:"+luna, []byte("EAAGtokenVencidoDeLaTiendaXX"))
		_, err := tx.Exec(ctx, `UPDATE meta_conversiones SET token = $1`, c)
		return err
	})
	// Un token que no se descifra (otra clave): error, se reintenta y la compra no se pierde.
	if err := enviar(ctx, es[0]); err == nil {
		t.Fatal("token que no se descifra y no se reintenta")
	}
	s2 := &conversiones.Servicio{Pool: b.App, Cifra: s.Cifra, HTTP: srv.Client(), Base: srv.URL}
	if err := s2.Guardar(ctx, luna, conversiones.Conexion{Token: token}); err != nil {
		t.Fatal(err)
	}
	_ = b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		c, _ := s.Cifra.Cifrar("meta:"+luna, []byte("EAAGtokenVencidoDeLaTiendaXX"))
		_, err := tx.Exec(ctx, `UPDATE meta_conversiones SET token = $1`, c)
		return err
	})
	if err := enviar(ctx, es[0]); err != nil {
		t.Fatalf("token vencido se reintenta: %v", err)
	}
	if e, _ := s.Leer(ctx, luna); e.UltimoError == nil || !strings.Contains(*e.UltimoError, "Invalid OAuth") {
		t.Fatalf("el panel no ve el error: %+v", e)
	}

	// Desconectar: las compras que quedaban no salen.
	if err := s.Borrar(ctx, luna); err != nil {
		t.Fatal(err)
	}
	antes := len(g.eventos)
	if err := enviar(ctx, es[0]); err != nil || len(g.eventos) != antes {
		t.Fatalf("desconectada mandó igual: %v", err)
	}
}

func itoa(n int) string { b, _ := json.Marshal(n); return string(b) }
