package promos_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/configuracion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/mitienda"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/newsletter"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/promos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

func ventana(t *testing.T, js string) mitienda.Ventana {
	t.Helper()
	var v mitienda.Ventana
	if err := json.Unmarshal([]byte(js), &v); err != nil {
		t.Fatal(err)
	}
	return v
}

func publicar(t *testing.T, pool interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
}, tienda string, d mitienda.Diseno) {
	t.Helper()
	ctx := context.Background()
	if err := mitienda.Guardar(ctx, pool, tienda, d); err != nil {
		t.Fatal(err)
	}
	if err := mitienda.Publicar(ctx, pool, tienda); err != nil {
		t.Fatal(err)
	}
}

func TestCaptarRuletaEventosYMetricas(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	if err := configuracion.GuardarNegocio(ctx, b.App, luna, configuracion.Negocio{Nombre: "Luna", Email: "hola@luna.com.ar"}); err != nil {
		t.Fatal(err)
	}
	d := mitienda.Inicial("litoral")
	d.Ventanas = []mitienda.Ventana{
		ventana(t, `{"id":"bienve01","nombre":"Bienvenida","activa":true,"variante":"centrada","contenido":"captacion",
			"datos":{"titulo":"10 % en tu primera compra","pedirNombre":true,"premio":{"tipo":"cuponUnico","descuento":"porcentaje","valor":10,"dias":7}},
			"disparo":{"cuando":"segundos","valor":8},"frecuencia":{"cada":"sesion"},"publico":"todos","paginas":[],"dispositivos":"todos"}`),
		ventana(t, `{"id":"ruleta01","nombre":"Ruleta","activa":true,"variante":"centrada","contenido":"ruleta",
			"datos":{"titulo":"Girá","porciones":[
				{"texto":"Nada","premio":{"tipo":"nada"},"peso":0},
				{"texto":"Envío gratis","premio":{"tipo":"cuponUnico","descuento":"envioGratis"},"peso":10},
				{"texto":"Nada 2","premio":{"tipo":"nada"},"peso":0},
				{"texto":"Nada 3","premio":{"tipo":"nada"},"peso":0}]},
			"disparo":{"cuando":"entrar"},"frecuencia":{"cada":"sesion"},"paginas":[]}`),
		ventana(t, `{"id":"apagada","nombre":"Apagada","activa":false,"variante":"centrada","contenido":"captacion",
			"datos":{"titulo":"x","premio":{"tipo":"nada"}},"disparo":{"cuando":"entrar"},"frecuencia":{"cada":"sesion"},"paginas":[]}`),
	}
	d.Estilo = &mitienda.Estilo{Conversion: map[string]any{"pruebaSocial": map[string]any{"activa": true}}}
	publicar(t, b.App, luna, d)
	ahora := time.Now()

	// Sin consentimiento o con un email que no sirve: no.
	if _, err := promos.Participar(ctx, b.App, "luna", "captacion", promos.Pedido{Ventana: "bienve01", Email: "ana@example.com"}, ahora); !errors.Is(err, promos.ErrInvalido) {
		t.Fatalf("sin consentimiento: %v", err)
	}
	if _, err := promos.Participar(ctx, b.App, "luna", "captacion", promos.Pedido{Ventana: "apagada", Email: "ana@example.com", Consentimiento: true}, ahora); !errors.Is(err, promos.ErrTerminada) {
		t.Fatalf("apagada: %v", err)
	}
	if _, err := promos.Participar(ctx, b.App, "luna", "ruleta", promos.Pedido{Ventana: "bienve01", Email: "ana@example.com", Consentimiento: true}, ahora); !errors.Is(err, promos.ErrNoExiste) {
		t.Fatalf("captación por la ruleta: %v", err)
	}
	p := promos.Pedido{Ventana: "bienve01", Email: " Ana@Example.com ", Nombre: "Ana", Consentimiento: true}
	r1, err := promos.Participar(ctx, b.App, "luna", "captacion", p, ahora)
	if err != nil {
		t.Fatal(err)
	}
	if r1.Premio.Tipo != "cuponUnico" || !strings.HasPrefix(r1.Premio.Codigo, "BIENVE-") || r1.Premio.Descripcion != "10 % off" || r1.Premio.Vence == nil || r1.Repetido {
		t.Fatalf("premio: %+v", r1)
	}
	// Idempotente: el mismo email devuelve el mismo código.
	r2, err := promos.Participar(ctx, b.App, "luna", "captacion", p, ahora)
	if err != nil || r2.Premio.Codigo != r1.Premio.Codigo || !r2.Repetido {
		t.Fatalf("repetido: %+v %v", r2, err)
	}
	// El cupón es de un uso, con su origen, y no aparece en la lista del panel.
	var limite int
	var origen string
	if err := b.App.EnTienda(ctx, luna, func(x pgx.Tx) error {
		return x.QueryRow(ctx, `SELECT limite, origen FROM cupones WHERE codigo = $1`, r1.Premio.Codigo).Scan(&limite, &origen)
	}); err != nil || limite != 1 || origen != "ventana:bienve01" {
		t.Fatalf("cupón: %d %q %v", limite, origen, err)
	}
	// El contacto está en la newsletter, con su origen y su nombre, pendiente.
	panel, err := newsletter.Leer(ctx, b.App, luna, ahora)
	if err != nil || len(panel.Suscriptores) != 1 || panel.Suscriptores[0].Origen != "ventana:bienve01" ||
		panel.Suscriptores[0].Nombre == nil || *panel.Suscriptores[0].Nombre != "Ana" || panel.Suscriptores[0].Confirmado {
		t.Fatalf("newsletter: %+v %v", panel.Suscriptores, err)
	}
	// El mail del premio sale una vez, con el código y el botón para confirmar.
	mails := &correo.Memoria{}
	w := &outbox.Worker{Pool: b.Worker, Manejadores: (&promos.Mails{Pool: b.App, Correo: mails, Sitio: "https://tulocalweb.com.ar"}).Manejadores()}
	for range 5 {
		if n, err := w.Pasada(ctx); err != nil {
			t.Fatal(err)
		} else if n == 0 {
			break
		}
	}
	if len(mails.Enviados) != 1 || !strings.Contains(mails.Enviados[0].Texto, r1.Premio.Codigo) ||
		!strings.Contains(mails.Enviados[0].HTML, "/newsletter/confirmar/") || mails.Enviados[0].ResponderA != "hola@luna.com.ar" {
		t.Fatalf("mails: %+v", mails.Enviados)
	}

	// La ruleta: el servidor elige (sólo una porción tiene peso) y es una por email.
	g, err := promos.Participar(ctx, b.App, "luna", "ruleta", promos.Pedido{Ventana: "ruleta01", Email: "beto@example.com", Consentimiento: true}, ahora)
	if err != nil || g.Porcion == nil || *g.Porcion != 1 || g.Texto != "Envío gratis" || g.Premio.Descripcion != "Envío gratis" {
		t.Fatalf("ruleta: %+v %v", g, err)
	}
	g2, err := promos.Participar(ctx, b.App, "luna", "ruleta", promos.Pedido{Ventana: "ruleta01", Email: "beto@example.com", Consentimiento: true}, ahora)
	if err != nil || !g2.Repetido || g2.Premio.Codigo != g.Premio.Codigo || *g2.Porcion != 1 {
		t.Fatalf("ruleta repetida: %+v %v", g2, err)
	}

	// Eventos: lo que no es de una ventana publicada se ignora.
	if err := promos.Registrar(ctx, b.App, "luna", []promos.Evento{{Ventana: "bienve01", Tipo: "vista", N: 3}, {Ventana: "bienve01", Tipo: "cierre"},
		{Ventana: "otraaaa1", Tipo: "vista"}, {Ventana: "ruleta01", Tipo: "clic"}}, ahora); err != nil {
		t.Fatal(err)
	}
	if err := promos.Registrar(ctx, b.App, "luna", []promos.Evento{{Ventana: "bienve01", Tipo: "vista", N: 500}}, ahora); err != nil {
		t.Fatal(err)
	}

	// Una venta con el cupón de la ventana (pagada) y la prueba social.
	if err := b.App.EnTienda(ctx, luna, func(x pgx.Tx) error {
		_, err := x.Exec(ctx, `INSERT INTO pedidos (tienda_id, numero, comprador, canal, pago, estado_pago, envio, ciudad_envio, estado, items,
				subtotal, total, cupon, creado_en)
			VALUES ($1, 1001, '{"nombre":"ana gómez","email":"ana@example.com"}', 'web', 'mp', 'Aprobado', 'correo', 'X5000', 'Para preparar',
				'[{"nombre":"Mate Imperial","cantidad":1,"precio":20000}]', 20000, 18000, $2, $3)`, luna, r1.Premio.Codigo, ahora.Add(-2*time.Hour))
		return err
	}); err != nil {
		t.Fatal(err)
	}
	compras, err := promos.Recientes(ctx, b.App, "luna", ahora)
	if err != nil || len(compras) != 1 || compras[0].Nombre != "Ana" || compras[0].Provincia != "Córdoba" || compras[0].Producto != "Mate Imperial" {
		t.Fatalf("recientes: %+v %v", compras, err)
	}

	desde, hasta, _ := promos.Rango("", "", ahora)
	m, err := promos.LeerMetricas(ctx, b.App, luna, desde, hasta)
	if err != nil {
		t.Fatal(err)
	}
	var bv, rl promos.MetricaVentana
	for _, v := range m.Ventanas {
		switch v.ID {
		case "bienve01":
			bv = v
		case "ruleta01":
			rl = v
		}
	}
	if bv.Vistas != 23 || bv.Cierres != 1 || bv.Envios != 1 || bv.Ventas != 1 || bv.Monto != 18000 || bv.TasaEnvio == nil || len(bv.Dias) != 1 {
		t.Fatalf("métricas de la bienvenida: %+v", bv)
	}
	if rl.Clics != 1 || rl.Envios != 1 || m.Total.Envios != 2 || len(m.Ventanas) != 3 {
		t.Fatalf("métricas: %+v", m)
	}

	// Por HTTP: el tope por IP de captar.
	mux := http.NewServeMux()
	promos.Rutas(mux, b.App, func(*http.Request) string { return luna }, func(*http.Request) string { return "200.1.2.3" })
	codigo := 0
	for i := range 11 {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/v1/publico/tiendas/luna/promos/captar",
			strings.NewReader(`{"ventana":"bienve01","email":"x`+string(rune('a'+i))+`@example.com","consentimiento":true}`)))
		codigo = w.Code
	}
	if codigo != http.StatusTooManyRequests {
		t.Fatalf("el tope por IP: %d", codigo)
	}
	w2 := httptest.NewRecorder()
	mux.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/v1/panel/promos/metricas?desde=2026-01-01&hasta=2025-01-01", nil))
	if w2.Code != http.StatusBadRequest {
		t.Fatalf("rango al revés: %d", w2.Code)
	}
	w3 := httptest.NewRecorder()
	mux.ServeHTTP(w3, httptest.NewRequest(http.MethodPost, "/v1/publico/tiendas/luna/promos/eventos", strings.NewReader(`{"eventos":[{"ventana":"bienve01","tipo":"vista"}]}`)))
	if w3.Code != http.StatusNoContent {
		t.Fatalf("eventos: %d", w3.Code)
	}
}
