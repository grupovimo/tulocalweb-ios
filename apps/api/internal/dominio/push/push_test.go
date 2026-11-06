package push_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/configuracion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/push"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/vidriera"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/apns"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/fcm"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/servidor"
)

func silencio() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

func token(c string) string { return strings.Repeat(c, 64) }

// --- sin base ---

func TestPayload(t *testing.T) {
	b := push.Payload(push.AvisoVenta("t1", "duenio", push.Venta{Numero: 12, Total: 24020, Productos: 3, Pago: "mp", EstadoPago: "Aprobado"}))
	var p struct {
		Aps struct {
			Alert    struct{ Title, Body string } `json:"alert"`
			Sound    string                       `json:"sound"`
			ThreadID string                       `json:"thread-id"`
			Category string                       `json:"category"`
		} `json:"aps"`
		Ruta, Tienda string
	}
	if err := json.Unmarshal(b, &p); err != nil {
		t.Fatal(err)
	}
	if p.Aps.Alert.Title != "Nueva venta #12" || p.Aps.Alert.Body != "$ 24.020 · 3 productos · Mercado Pago acreditado" {
		t.Errorf("alerta: %+v", p.Aps.Alert)
	}
	// category «venta»: los botones Ver pedido y Marcar como preparado del iPhone.
	if p.Aps.Sound != "default" || p.Aps.ThreadID != "tienda-t1" || p.Aps.Category != "venta" || p.Ruta != "pedido/12" || p.Tienda != "t1" {
		t.Errorf("payload: %s", b)
	}
	if len(b) > 4096 {
		t.Errorf("el payload pasa los 4 KB de APNs: %d", len(b))
	}
}

func TestMontosSoloConFinanzas(t *testing.T) {
	v := push.Venta{Numero: 7, Total: 15215, Productos: 1, Pago: "transf", EstadoPago: "Pendiente"}
	for rol, conMonto := range map[string]bool{"duenio": true, "admin": true, "ventas": false, "local": false} {
		a := push.AvisoVenta("t", rol, v)
		if strings.Contains(a.Cuerpo, "$") != conMonto {
			t.Errorf("%s: %q (con monto=%v)", rol, a.Cuerpo, conMonto)
		}
		if !strings.HasSuffix(a.Cuerpo, "1 producto · Transferencia por confirmar") {
			t.Errorf("%s: %q", rol, a.Cuerpo)
		}
	}
	if a := push.AvisoIngreso("iPhone"); a.Hilo != "cuenta" || a.Ruta != "cuenta/seguridad" || strings.Contains(a.Cuerpo, ".1") {
		t.Errorf("ingreso: %+v", a)
	}
	if a := push.AvisoPreguntaML("t", 99, "Yerba"); a.Ruta != "preguntas/ml/99" || a.Cuerpo != "Te preguntaron por Yerba." {
		t.Errorf("pregunta: %+v", a)
	}
}

func TestNormalizar(t *testing.T) {
	if d, err := push.Normalizar(push.Dispositivo{Token: strings.ToUpper(token("a")), Modelo: "iPhone15,2\n"}); err != nil || d.Entorno != "produccion" || d.Token != token("a") || d.Modelo != "iPhone15,2" {
		t.Fatalf("%+v %v", d, err)
	}
	for _, d := range []push.Dispositivo{{Token: "corto"}, {Token: token("g")}, {Token: token("a"), Entorno: "dev"}} {
		if _, err := push.Normalizar(d); !errors.Is(err, push.ErrInvalido) {
			t.Errorf("%+v: %v", d, err)
		}
	}
}

// Sin APNs configurado no se manda nada y el evento se da por publicado.
func TestSinConfigNoMandaNada(t *testing.T) {
	s := push.Nuevo(nil, nil, silencio())
	for tipo, m := range s.Manejadores() {
		if err := m(context.Background(), outbox.Evento{Tipo: tipo, TiendaID: "x", Datos: []byte(`{"numero":1,"id":1,"usuario":"u"}`)}); err != nil {
			t.Errorf("%s sin config: %v", tipo, err)
		}
	}
}

// --- con base ---

type enviado struct {
	token string
	aviso map[string]any
}

// apnsFalso registra lo enviado y responde ErrTokenInvalido a los tokens muertos.
type apnsFalso struct {
	mu      sync.Mutex
	envios  []enviado
	muertos map[string]bool
}

func (f *apnsFalso) Enviar(_ context.Context, e apns.Envio) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	var p map[string]any
	_ = json.Unmarshal(e.Payload, &p)
	f.envios = append(f.envios, enviado{e.Token, p})
	if f.muertos[e.Token] {
		return apns.ErrTokenInvalido
	}
	return nil
}

func (f *apnsFalso) cuerpos() map[string]string {
	out := map[string]string{}
	for _, e := range f.envios {
		out[e.token] = e.aviso["aps"].(map[string]any)["alert"].(map[string]any)["body"].(string)
	}
	return out
}

type escenario struct {
	b      *pruebadb.Bases
	cs     *cuentas.Servicio
	h      http.Handler
	tienda string
	// Tokens de sesión por rol, e ids de persona.
	ses, ids map[string]string
}

func armar(t *testing.T) *escenario {
	t.Helper()
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	cs := &cuentas.Servicio{Pool: b.App, Correo: &correo.Memoria{}, Sitio: "https://tulocalweb.test"}
	e := &escenario{b: b, cs: cs, ses: map[string]string{}, ids: map[string]string{},
		h: servidor.Nuevo(servidor.Dependencias{App: b.App, Cuentas: cs, Log: silencio()})}
	e.tienda = pruebadb.CrearTienda(t, b.App, "luna")
	hash, err := cuentas.HashClave("clave-larga-1")
	if err != nil {
		t.Fatal(err)
	}
	for _, rol := range []string{"duenio", "admin", "ventas", "local"} {
		email := rol + "@luna.com"
		if err := b.App.EnTx(ctx, func(tx pgx.Tx) error {
			var id string
			if err := tx.QueryRow(ctx, `INSERT INTO auth.usuarios (email, nombre, clave_hash, email_verificado_en)
				VALUES ($1, $2, $3, now()) RETURNING id::text`, email, "Persona "+rol, hash).Scan(&id); err != nil {
				return err
			}
			e.ids[rol] = id
			_, err := tx.Exec(ctx, `INSERT INTO auth.membresias (usuario_id, tienda_id, rol) VALUES ($1, $2, $3)`, id, e.tienda, rol)
			return err
		}); err != nil {
			t.Fatal(err)
		}
		s, err := cs.Ingresar(ctx, email, "clave-larga-1", false, cuentas.Cliente{IP: "190.210.32.1", Agente: "TuLocalWeb/1.0 (iPhone; iOS 18.1)"})
		if err != nil {
			t.Fatal(err)
		}
		e.ses[rol] = s.Token
	}
	return e
}

func (e *escenario) pedir(sesion, metodo, cuerpo string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(metodo, "/v1/panel/dispositivos", strings.NewReader(cuerpo))
	if sesion != "" {
		r.Header.Set("Authorization", "Bearer "+sesion)
	}
	r.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.h.ServeHTTP(rec, r)
	return rec
}

func (e *escenario) registrar(t *testing.T, rol, tok string) {
	t.Helper()
	if rec := e.pedir(e.ses[rol], "POST", `{"token":"`+tok+`","entorno":"sandbox","modelo":"iPhone 15","version":"1.0 (1)"}`); rec.Code != http.StatusNoContent {
		t.Fatalf("registrar %s: %d %s", rol, rec.Code, rec.Body)
	}
}

// cuantos cuenta los dispositivos con el admin (sin RLS).
func (e *escenario) cuantos(t *testing.T, donde string, args ...any) int {
	t.Helper()
	conn, err := pgx.Connect(context.Background(), e.b.AdminURL)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(context.Background())
	var n int
	if err := conn.QueryRow(context.Background(), `SELECT count(*) FROM auth.dispositivos WHERE `+donde, args...).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func (e *escenario) admin(t *testing.T, sql string, args ...any) {
	t.Helper()
	conn, err := pgx.Connect(context.Background(), e.b.AdminURL)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(context.Background())
	if _, err := conn.Exec(context.Background(), sql, args...); err != nil {
		t.Fatal(err)
	}
}

func TestRegistroBajaYSesion(t *testing.T) {
	e := armar(t)
	ctx := context.Background()

	if rec := e.pedir("", "POST", `{"token":"`+token("a")+`"}`); rec.Code != http.StatusUnauthorized {
		t.Fatalf("sin sesión: %d", rec.Code)
	}
	if rec := e.pedir(e.ses["duenio"], "POST", `{"token":"nope"}`); rec.Code != http.StatusBadRequest {
		t.Fatalf("token inválido: %d", rec.Code)
	}
	// Cualquier rol registra su iPhone (roles.Panel).
	for rol, c := range map[string]string{"duenio": "a", "admin": "b", "ventas": "c", "local": "d"} {
		e.registrar(t, rol, token(c))
	}
	if n := e.cuantos(t, "true"); n != 4 {
		t.Fatalf("%d dispositivos, quería 4", n)
	}
	// Registrar otra vez el mismo es idempotente.
	e.registrar(t, "duenio", token("a"))
	if n := e.cuantos(t, "token = $1", token("a")); n != 1 {
		t.Fatalf("el mismo token quedó %d veces", n)
	}

	t.Run("RLS: cada persona ve y borra sólo los suyos", func(t *testing.T) {
		var vistos int
		var borrados int64
		err := e.b.App.EnUsuario(ctx, e.ids["ventas"], func(tx pgx.Tx) error {
			if err := tx.QueryRow(ctx, `SELECT count(*) FROM auth.dispositivos`).Scan(&vistos); err != nil {
				return err
			}
			tag, err := tx.Exec(ctx, `DELETE FROM auth.dispositivos WHERE token = $1`, token("a"))
			borrados = tag.RowsAffected()
			return err
		})
		if err != nil || vistos != 1 || borrados != 0 {
			t.Fatalf("ventas: err=%v ve %d (quería 1), borró %d ajenos", err, vistos, borrados)
		}
		var sinPersona int
		_ = e.b.App.EnTx(ctx, func(tx pgx.Tx) error {
			return tx.QueryRow(ctx, `SELECT count(*) FROM auth.dispositivos`).Scan(&sinPersona)
		})
		if sinPersona != 0 {
			t.Fatalf("sin persona fijada ve %d", sinPersona)
		}
		// Nadie inserta directo (sólo app.registrar_dispositivo, que exige la sesión propia).
		err = e.b.App.EnUsuario(ctx, e.ids["ventas"], func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `INSERT INTO auth.dispositivos (usuario_id, sesion_hash, token, entorno)
				SELECT $1, token_hash, $2, 'sandbox' FROM auth.sesiones WHERE usuario_id = $1 LIMIT 1`, e.ids["ventas"], token("f"))
			return err
		})
		if err == nil {
			t.Fatal("la app insertó un dispositivo sin pasar por la función")
		}
		// Ni registra el suyo con la sesión de otra persona.
		err = e.b.App.EnUsuario(ctx, e.ids["ventas"], func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `SELECT app.registrar_dispositivo((SELECT token_hash FROM auth.sesiones WHERE usuario_id = $1), $2, 'sandbox', '', '', 'ios')`,
				e.ids["duenio"], token("f"))
			return err
		})
		if err == nil {
			t.Fatal("registró un dispositivo con una sesión ajena")
		}
	})

	t.Run("el mismo iPhone con otra cuenta pasa a esa", func(t *testing.T) {
		e.registrar(t, "local", token("a"))
		if n := e.cuantos(t, "token = $1 AND usuario_id = $2", token("a"), e.ids["local"]); n != 1 {
			t.Fatal("el token no pasó a local")
		}
		if n := e.cuantos(t, "usuario_id = $1", e.ids["duenio"]); n != 0 {
			t.Fatalf("a la dueña le quedaron %d", n)
		}
		e.registrar(t, "duenio", token("a"))
	})

	t.Run("DELETE da de baja el de la sesión", func(t *testing.T) {
		if rec := e.pedir(e.ses["admin"], "DELETE", ""); rec.Code != http.StatusNoContent {
			t.Fatalf("baja: %d %s", rec.Code, rec.Body)
		}
		if n := e.cuantos(t, "usuario_id = $1", e.ids["admin"]); n != 0 {
			t.Fatalf("admin conserva %d", n)
		}
	})

	t.Run("cerrar la sesión se lleva el dispositivo", func(t *testing.T) {
		if err := e.cs.Salir(ctx, e.ses["ventas"]); err != nil {
			t.Fatal(err)
		}
		if n := e.cuantos(t, "usuario_id = $1", e.ids["ventas"]); n != 0 {
			t.Fatalf("ventas conserva %d tras salir", n)
		}
	})

	t.Run("una sesión vencida no recibe", func(t *testing.T) {
		e.admin(t, `UPDATE auth.sesiones SET vence_en = now() - interval '1 second' WHERE usuario_id = $1`, e.ids["local"])
		var tokens []string
		err := e.b.App.EnTienda(ctx, e.tienda, func(tx pgx.Tx) error {
			filas, err := tx.Query(ctx, `SELECT token FROM app.dispositivos_de_tienda()`)
			if err != nil {
				return err
			}
			tokens, err = pgx.CollectRows(filas, pgx.RowTo[string])
			return err
		})
		if err != nil || len(tokens) != 1 || tokens[0] != token("a") {
			t.Fatalf("destinos: %v %v, quería sólo el de la dueña", tokens, err)
		}
	})
}

// venta compra en la tienda online por transferencia y devuelve el número.
func vender(t *testing.T, e *escenario) {
	t.Helper()
	ctx := context.Background()
	stock := 20
	p, err := listas.Guardar(ctx, e.b.App, e.tienda, "", listas.Editable{Nombre: "Mate", Precio: 12010, Stock: &stock})
	if err != nil {
		t.Fatal(err)
	}
	if err := configuracion.GuardarNegocio(ctx, e.b.App, e.tienda, configuracion.Negocio{Nombre: "Luna", Alias: "luna.mates"}); err != nil {
		t.Fatal(err)
	}
	if err := configuracion.GuardarEntregas(ctx, e.b.App, e.tienda, configuracion.Entregas{Retiro: true, RetiroDireccion: "Gorriti 4500"}); err != nil {
		t.Fatal(err)
	}
	c := vidriera.Compra{Email: "lucia@ejemplo.com", Nombre: "Lucía Pérez", Telefono: "11 5555-0101", CP: "1414", Envio: "retiro", Pago: "transferencia"}
	c.Carrito = append(c.Carrito, vidriera.LineaCarrito{ID: p.ID, Cantidad: 2})
	if _, err := vidriera.Comprar(ctx, e.b.App, nil, "", "luna", c, time.Now()); err != nil {
		t.Fatal(err)
	}
}

func TestVentaNuevaPorRol(t *testing.T) {
	e := armar(t)
	ctx := context.Background()
	for rol, c := range map[string]string{"duenio": "a", "admin": "b", "ventas": "c", "local": "d"} {
		e.registrar(t, rol, token(c))
	}
	f := &apnsFalso{muertos: map[string]bool{token("d"): true}}
	s := &push.Servicio{Pool: e.b.App, APNs: f, Log: silencio()}
	w := &outbox.Worker{Pool: e.b.Worker, Manejadores: s.Manejadores()}
	vender(t, e)
	if _, err := w.Pasada(ctx); err != nil {
		t.Fatal(err)
	}
	got := f.cuerpos()
	if len(got) != 4 {
		t.Fatalf("salieron %d avisos, quería 4: %v", len(got), got)
	}
	for c, conMonto := range map[string]bool{"a": true, "b": true, "c": false, "d": false} {
		if strings.Contains(got[token(c)], "$ 24.020") != conMonto || strings.Contains(got[token(c)], "Lucía") {
			t.Errorf("%s: %q", c, got[token(c)])
		}
	}
	if ruta := f.envios[0].aviso["ruta"]; ruta == nil || !strings.HasPrefix(ruta.(string), "pedido/") {
		t.Errorf("ruta: %v", ruta)
	}
	// El token muerto se borró solo.
	if n := e.cuantos(t, "token = $1", token("d")); n != 0 {
		t.Fatal("el token inválido sigue registrado")
	}
	// Una sola vez: reprocesar el evento no manda otro.
	if err := s.Manejadores()["push.venta"](ctx, outbox.Evento{TiendaID: e.tienda, Datos: []byte(`{"numero":1}`)}); err != nil {
		t.Fatal(err)
	}
	if len(f.envios) != 4 {
		t.Fatalf("reprocesar mandó otra vez: %d", len(f.envios))
	}

	// Con Venta nueva › Celular apagado, no sale.
	e.admin(t, `UPDATE tiendas SET preferencias = jsonb_set(coalesce(preferencias, '{}'), '{canales}', '{"matriz":{"venta:celular":false}}') WHERE id = $1`, e.tienda)
	vender(t, e)
	if _, err := w.Pasada(ctx); err != nil {
		t.Fatal(err)
	}
	if len(f.envios) != 4 {
		t.Fatalf("con el push apagado salieron %d más", len(f.envios)-4)
	}
}

// «Te quedaste sin stock» al iPhone: a quien tiene Productos (no a Local), con su interruptor.
func TestSinStockPorRol(t *testing.T) {
	e := armar(t)
	ctx := context.Background()
	for rol, c := range map[string]string{"duenio": "a", "admin": "b", "ventas": "c", "local": "d"} {
		e.registrar(t, rol, token(c))
	}
	f := &apnsFalso{}
	s := &push.Servicio{Pool: e.b.App, APNs: f, Log: silencio()}
	w := &outbox.Worker{Pool: e.b.Worker, Manejadores: s.Manejadores()}
	nuevo := func(nombre string) listas.Producto {
		t.Helper()
		uno := 1
		p, err := listas.Guardar(ctx, e.b.App, e.tienda, "", listas.Editable{Nombre: nombre, Precio: 9000, Stock: &uno})
		if err != nil {
			t.Fatal(err)
		}
		return p
	}
	ultima := func(p listas.Producto) {
		t.Helper()
		if _, _, err := pedidos.CrearDesdeProductos(ctx, e.b.App, e.tienda, pedidos.PedidoNuevo{Pago: "efectivo", Canal: "pos",
			Items: []pedidos.Linea{{ProductoID: p.ID, Cantidad: 1}}}); err != nil {
			t.Fatal(err)
		}
		if _, err := w.Pasada(ctx); err != nil {
			t.Fatal(err)
		}
	}
	mate := nuevo("Mate imperial")
	ultima(mate)
	got := f.cuerpos()
	if len(got) != 3 || got[token("d")] != "" {
		t.Fatalf("salieron %d avisos, quería 3 (sin Local): %v", len(got), got)
	}
	if got[token("a")] != "Se vendió la última unidad de Mate imperial. Tocá para reponer." {
		t.Errorf("cuerpo: %q", got[token("a")])
	}
	if ruta := f.envios[0].aviso["ruta"]; ruta != "producto/"+mate.ID {
		t.Errorf("ruta: %v", ruta)
	}
	// Con Sin stock › Celular apagado, no sale.
	e.admin(t, `UPDATE tiendas SET preferencias = jsonb_set(coalesce(preferencias, '{}'), '{canales}', '{"matriz":{"sinstock:celular":false}}') WHERE id = $1`, e.tienda)
	ultima(nuevo("Yerba"))
	if len(f.envios) != 3 {
		t.Fatalf("con el push apagado salieron %d más", len(f.envios)-3)
	}
}

func TestIngresoSoloALaPersona(t *testing.T) {
	e := armar(t)
	ctx := context.Background()
	e.registrar(t, "duenio", token("a"))
	e.registrar(t, "admin", token("b"))
	f := &apnsFalso{}
	s := &push.Servicio{Pool: e.b.App, APNs: f, Log: silencio()}
	w := &outbox.Worker{Pool: e.b.Worker, Manejadores: s.Manejadores()}
	// Los ingresos de armar ya están encolados: nadie tenía iPhone todavía.
	if _, err := w.Pasada(ctx); err != nil {
		t.Fatal(err)
	}
	if len(f.envios) != 0 {
		t.Fatalf("salieron %d avisos de ingresos sin iPhone", len(f.envios))
	}
	// La dueña entra desde otro lado: le llega a su iPhone, no al de admin.
	nueva, err := e.cs.Ingresar(ctx, "duenio@luna.com", "clave-larga-1", false, cuentas.Cliente{IP: "190.210.32.1", Agente: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Pasada(ctx); err != nil {
		t.Fatal(err)
	}
	if len(f.envios) != 1 || f.envios[0].token != token("a") {
		t.Fatalf("envíos: %+v", f.envios)
	}
	alerta := f.envios[0].aviso["aps"].(map[string]any)["alert"].(map[string]any)
	if alerta["title"] != "Entraron a tu cuenta" || strings.Contains(alerta["body"].(string), "190.") {
		t.Errorf("alerta: %v", alerta)
	}
	// Si después registra ese iPhone con la sesión nueva, el aviso de esa misma sesión no le llega.
	f.envios = nil
	e.ses["duenio"] = nueva.Token
	e.registrar(t, "duenio", token("e"))
	if err := s.Manejadores()["cuenta.ingreso_push"](ctx, outbox.Evento{Datos: mustJSON(cuentas.IngresoPush{Usuario: e.ids["duenio"],
		Sesion: hashDe(t, e, nueva.Token), Aparato: "Mac"})}); err != nil {
		t.Fatal(err)
	}
	if len(f.envios) != 1 || f.envios[0].token != token("a") {
		t.Fatalf("al iPhone de la sesión nueva no le tenía que llegar: %+v", f.envios)
	}
	// Con el aviso de ingreso apagado, tampoco sale el push.
	e.admin(t, `UPDATE auth.usuarios SET seguridad = '{"avisoIngreso": false}' WHERE id = $1`, e.ids["duenio"])
	var antes int
	_ = e.b.Worker.QueryRow(ctx, `SELECT count(*) FROM outbox WHERE tipo = 'cuenta.ingreso_push'`).Scan(&antes)
	if _, err := e.cs.Ingresar(ctx, "duenio@luna.com", "clave-larga-1", false, cuentas.Cliente{IP: "190.210.32.1"}); err != nil {
		t.Fatal(err)
	}
	var despues int
	_ = e.b.Worker.QueryRow(ctx, `SELECT count(*) FROM outbox WHERE tipo = 'cuenta.ingreso_push'`).Scan(&despues)
	if despues != antes {
		t.Fatal("con el aviso de ingreso apagado se encoló el push")
	}
}

func claveDePrueba(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()
	k, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return k
}

func mustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

// hashDe busca el hash de la sesión con el token (la sesión resuelta lo trae).
func hashDe(t *testing.T, e *escenario, tok string) string {
	t.Helper()
	d, err := e.cs.Sesion(context.Background(), tok)
	if err != nil {
		t.Fatal(err)
	}
	return d.SesionHash
}

// Con el cliente de APNs de verdad contra un servidor HTTP/2 falso: un 410 borra el token.
func TestUn410BorraElToken(t *testing.T) {
	e := armar(t)
	ctx := context.Background()
	e.registrar(t, "duenio", token("a"))
	e.registrar(t, "admin", token("b"))
	var mu sync.Mutex
	var rutas []string
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		rutas = append(rutas, r.Proto+" "+r.URL.Path)
		mu.Unlock()
		if strings.HasSuffix(r.URL.Path, token("b")) {
			w.WriteHeader(http.StatusGone)
			_, _ = w.Write([]byte(`{"reason":"Unregistered","timestamp":1790000000000}`))
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	srv.EnableHTTP2 = true
	srv.StartTLS()
	defer srv.Close()
	cli := &apns.Cliente{KeyID: "KEY1234567", TeamID: "TEAM123456", Topic: "ar.com.tulocalweb.app", Clave: claveDePrueba(t),
		HTTP: srv.Client(), Hosts: map[string]string{"sandbox": srv.URL, "produccion": srv.URL}}
	s := push.Nuevo(e.b.App, cli, silencio())
	w := &outbox.Worker{Pool: e.b.Worker, Manejadores: s.Manejadores()}
	vender(t, e)
	if _, err := w.Pasada(ctx); err != nil {
		t.Fatal(err)
	}
	if len(rutas) != 2 || !strings.HasPrefix(rutas[0], "HTTP/2.0 /3/device/") {
		t.Fatalf("pedidos a APNs: %v", rutas)
	}
	if n := e.cuantos(t, "token = $1", token("b")); n != 0 {
		t.Fatal("el 410 no borró el token")
	}
	if n := e.cuantos(t, "token = $1", token("a")); n != 1 {
		t.Fatal("se borró un token que andaba")
	}
}

// --- Android (FCM) ---

// tokenFCM: con mayúsculas y dos puntos, como los de verdad (no se pasan a minúscula).
func tokenFCM(c string) string {
	return "dA" + strings.Repeat(c, 20) + ":APA91b" + strings.Repeat(c, 40)
}

// fcmFalso registra lo enviado y responde ErrTokenInvalido a los tokens muertos.
type fcmFalso struct {
	mu      sync.Mutex
	envios  []fcm.Envio
	muertos map[string]bool
}

func (f *fcmFalso) Enviar(_ context.Context, e fcm.Envio) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.envios = append(f.envios, e)
	if f.muertos[e.Token] {
		return fcm.ErrTokenInvalido
	}
	return nil
}

func (e *escenario) registrarAndroid(t *testing.T, rol, tok string) {
	t.Helper()
	if rec := e.pedir(e.ses[rol], "POST", `{"token":"`+tok+`","entorno":"produccion","modelo":"Pixel 9","version":"1.0","plataforma":"android"}`); rec.Code != http.StatusNoContent {
		t.Fatalf("registrar Android %s: %d %s", rol, rec.Code, rec.Body)
	}
}

func TestNormalizarAndroid(t *testing.T) {
	d, err := push.Normalizar(push.Dispositivo{Token: " " + tokenFCM("x") + " ", Plataforma: "android"})
	if err != nil || d.Token != tokenFCM("x") || d.Plataforma != "android" {
		t.Fatalf("Android: %+v %v", d, err)
	}
	for _, x := range []push.Dispositivo{
		{Token: "corto", Plataforma: "android"},
		{Token: tokenFCM("x") + " ", Plataforma: "windows"},
		{Token: tokenFCM("x")}, // sin plataforma es iPhone: no es hexadecimal
	} {
		if _, err := push.Normalizar(x); !errors.Is(err, push.ErrInvalido) {
			t.Errorf("%+v: %v", x, err)
		}
	}
	if d, _ := push.Normalizar(push.Dispositivo{Token: token("a")}); d.Plataforma != "ios" {
		t.Errorf("sin plataforma: %q", d.Plataforma)
	}
}

// La venta llega al iPhone por APNs y a Android por FCM sólo con datos (la app arma el aviso con
// sus botones); un token de FCM muerto se borra.
func TestVentaNuevaEnAndroid(t *testing.T) {
	e := armar(t)
	ctx := context.Background()
	e.registrar(t, "duenio", token("a"))
	e.registrarAndroid(t, "admin", tokenFCM("b"))
	e.registrarAndroid(t, "ventas", tokenFCM("c"))
	e.registrarAndroid(t, "local", tokenFCM("d"))
	a := &apnsFalso{}
	f := &fcmFalso{muertos: map[string]bool{tokenFCM("d"): true}}
	s := &push.Servicio{Pool: e.b.App, APNs: a, FCM: f, Log: silencio()}
	w := &outbox.Worker{Pool: e.b.Worker, Manejadores: s.Manejadores()}
	vender(t, e)
	if _, err := w.Pasada(ctx); err != nil {
		t.Fatal(err)
	}
	if len(a.envios) != 1 || a.envios[0].token != token("a") {
		t.Fatalf("APNs: %+v", a.envios)
	}
	var avisos []fcm.Envio
	for _, x := range f.envios {
		if x.Datos["tipo"] != "actividad" {
			avisos = append(avisos, x)
		}
	}
	if len(avisos) != 3 {
		t.Fatalf("FCM: salieron %d avisos, quería 3: %+v", len(avisos), avisos)
	}
	for _, x := range avisos {
		if !x.SoloDatos || !strings.HasPrefix(x.Titulo, "Nueva venta #") || x.Datos["categoria"] != "venta" ||
			x.Datos["ruta"] != "pedido/"+strings.TrimPrefix(x.Titulo, "Nueva venta #") || x.Datos["tienda"] != e.tienda {
			t.Errorf("aviso a Android: %+v", x)
		}
		// El monto, sólo con Finanzas (admin sí, ventas no).
		if conMonto := strings.Contains(x.Cuerpo, "$ 24.020"); conMonto != (x.Token == tokenFCM("b")) {
			t.Errorf("%s: %q", x.Token[:4], x.Cuerpo)
		}
	}
	if n := e.cuantos(t, "token = $1", tokenFCM("d")); n != 0 {
		t.Fatal("el token de FCM inválido sigue registrado")
	}
	if n := e.cuantos(t, "token = $1 AND plataforma = 'android'", tokenFCM("b")); n != 1 {
		t.Fatal("el token de Android no quedó con su plataforma (o se pasó a minúscula)")
	}
}

// Sin FCM configurado, los teléfonos Android no se eligen (y el iPhone sigue recibiendo).
func TestSinFCMNoSeEligeAndroid(t *testing.T) {
	e := armar(t)
	ctx := context.Background()
	e.registrar(t, "duenio", token("a"))
	e.registrarAndroid(t, "admin", tokenFCM("b"))
	a := &apnsFalso{}
	s := &push.Servicio{Pool: e.b.App, APNs: a, Log: silencio()}
	w := &outbox.Worker{Pool: e.b.Worker, Manejadores: s.Manejadores()}
	vender(t, e)
	if _, err := w.Pasada(ctx); err != nil {
		t.Fatal(err)
	}
	if len(a.envios) != 1 || a.envios[0].token != token("a") {
		t.Fatalf("APNs: %+v", a.envios)
	}
}

// La notificación en curso de Android: el teléfono pasa su token como el de inicio; con la venta, el
// worker le manda el estado por FCM (tipo «actividad»).
func TestActividadEnAndroid(t *testing.T) {
	e := armar(t)
	ctx := context.Background()
	e.registrarAndroid(t, "duenio", tokenFCM("a"))
	r := httptest.NewRequest("PUT", "/v1/panel/dispositivos/actividad", strings.NewReader(`{"token":"`+tokenFCM("a")+`","actividad":"","inicio":"`+tokenFCM("a")+`"}`))
	r.Header.Set("Authorization", "Bearer "+e.ses["duenio"])
	r.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.h.ServeHTTP(rec, r)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("actividad: %d %s", rec.Code, rec.Body)
	}
	// Otro token como actividad no vale.
	if err := push.GuardarActividad(ctx, e.b.App, e.ids["duenio"], "x", push.TokensActividad{Token: tokenFCM("a"), Actividad: tokenFCM("z")}); !errors.Is(err, push.ErrInvalido) {
		t.Errorf("actividad ajena: %v", err)
	}
	f := &fcmFalso{}
	s := &push.Servicio{Pool: e.b.App, FCM: f, Log: silencio()}
	w := &outbox.Worker{Pool: e.b.Worker, Manejadores: s.Manejadores()}
	vender(t, e)
	// El pedido de prueba puede quedar «Sin pagar» (transferencia): se lo pasa a preparar.
	e.admin(t, `UPDATE pedidos SET estado = 'Para preparar' WHERE tienda_id = $1`, e.tienda)
	if _, err := w.Pasada(ctx); err != nil {
		t.Fatal(err)
	}
	var act *fcm.Envio
	for i, x := range f.envios {
		if x.Datos["tipo"] == "actividad" {
			act = &f.envios[i]
		}
	}
	if act == nil {
		t.Fatalf("no salió la actividad: %+v", f.envios)
	}
	if !act.SoloDatos || act.Token != tokenFCM("a") || act.Datos["porPreparar"] == "" || act.Datos["tienda"] == "" || act.Datos["terminar"] != "" {
		t.Errorf("actividad: %+v", act)
	}
}
