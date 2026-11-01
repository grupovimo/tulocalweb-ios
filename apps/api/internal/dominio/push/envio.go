package push

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/roles"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/apns"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/webpush"
)

// Enviador es lo que hace falta de APNs (*apns.Cliente lo cumple; en los tests, uno falso).
type Enviador interface {
	Enviar(context.Context, apns.Envio) error
}

// EnviadorWeb es lo que hace falta de Web Push (*webpush.Cliente lo cumple; en los tests, uno falso).
type EnviadorWeb interface {
	Enviar(context.Context, webpush.Envio) error
}

// Servicio manda los avisos del outbox. Sin APNs ni Web Push (nil), los manejadores no hacen nada.
type Servicio struct {
	Pool Pool
	APNs Enviador
	// Web: el push web del panel instalado como app (§43). nil: apagado.
	Web EnviadorWeb
	Log *slog.Logger
}

// Nuevo arma el servicio. Con cli nil (sin MOSTRADOR_APNS_*) el envío al iPhone queda apagado y lo
// dice el log una sola vez, al arrancar.
func Nuevo(p Pool, cli *apns.Cliente, log *slog.Logger) *Servicio {
	if log == nil {
		log = slog.Default()
	}
	s := &Servicio{Pool: p, Log: log}
	if cli == nil {
		log.Info("push: APNs sin configurar (MOSTRADOR_APNS_*), no se mandan avisos al iPhone")
		return s
	}
	s.APNs = cli
	return s
}

// ConWeb suma el push web. Con cli nil (sin MOSTRADOR_VAPID_*) queda apagado y lo dice el log.
func (s *Servicio) ConWeb(cli *webpush.Cliente) *Servicio {
	if cli == nil {
		s.Log.Info("push: Web Push sin configurar (MOSTRADOR_VAPID_*), no se mandan avisos al panel instalado")
		return s
	}
	s.Web = cli
	return s
}

func (s *Servicio) apnsActivo() bool {
	if s == nil || s.APNs == nil {
		return false
	}
	if c, ok := s.APNs.(*apns.Cliente); ok && c == nil {
		return false
	}
	return true
}

func (s *Servicio) webActivo() bool {
	if s == nil || s.Web == nil {
		return false
	}
	if c, ok := s.Web.(*webpush.Cliente); ok && c == nil {
		return false
	}
	return true
}

func (s *Servicio) activo() bool { return s.apnsActivo() || s.webActivo() }

// Manejadores del outbox.
func (s *Servicio) Manejadores() map[string]outbox.Manejador {
	return map[string]outbox.Manejador{
		"push.venta":          s.venta,
		"push.pregunta_ml":    s.preguntaML,
		"cuenta.ingreso_push": s.ingreso,
		"push.sin_stock":      s.sinStock,
		"cuenta.push_prueba":  s.prueba,
	}
}

// AvisoSinStock: «Te quedaste sin stock» con el producto (y la variante), y la ruta para reponer.
func AvisoSinStock(tienda, producto string, a pedidos.Agotado) Aviso {
	nombre := a.Nombre
	if a.Variante != "" {
		nombre += " (" + a.Variante + ")"
	}
	return Aviso{Titulo: "Te quedaste sin stock", Cuerpo: "Se vendió la última unidad de " + nombre + ". Tocá para reponer.",
		Hilo: HiloTienda(tienda), Ruta: "producto/" + producto, Tienda: tienda}
}

// sinStock: push.sin_stock, que encola Go al descontar (uno por producto por día, junto con el mail
// tienda.sin_stock). A quien tiene Productos; no sale si la tienda apagó Sin stock › Celular ni si
// el producto ya no está en 0 o se pausó.
func (s *Servicio) sinStock(ctx context.Context, ev outbox.Evento) error {
	if !s.activo() {
		return nil
	}
	var d pedidos.SinStock
	if json.Unmarshal(ev.Datos, &d) != nil || d.Producto == "" {
		return nil
	}
	var a pedidos.Agotado
	var prendido bool
	err := s.Pool.EnTienda(ctx, ev.TiendaID, func(t pgx.Tx) error {
		var err error
		if a, err = pedidos.LeerSinStock(ctx, t, d); err != nil {
			return err
		}
		return t.QueryRow(ctx, `SELECT coalesce(preferencias->'canales'->'matriz'->'sinstock:celular', 'true') <> 'false'
			FROM tiendas WHERE id = $1`, ev.TiendaID).Scan(&prendido)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if !prendido || !a.Sigue {
		return nil
	}
	dest, err := s.deTienda(ctx, ev.TiendaID)
	if err != nil {
		return err
	}
	aviso := AvisoSinStock(ev.TiendaID, d.Producto, a)
	var envios []envio
	for _, x := range dest {
		if roles.Puede(x.rol, roles.Productos) {
			envios = append(envios, para(x, aviso, "sin-stock-"+d.Producto, 24*time.Hour))
		}
	}
	return s.mandar(ctx, envios)
}

// destino es un iPhone (token de APNs) o un panel instalado (suscripción web), con el rol de su
// persona en la tienda del aviso.
type destino struct {
	usuario, rol   string
	token, entorno string
	web            *webpush.Suscripcion
}

// envio es un aviso ya armado para un destino: por APNs o por Web Push.
type envio struct {
	apns *apns.Envio
	web  *webpush.Envio
}

// para arma el envío de un aviso a un destino. colapso agrupa los reintentos (apns-collapse-id, el
// Topic de Web Push y la etiqueta de la notificación).
func para(d destino, a Aviso, colapso string, vida time.Duration) envio {
	if d.web != nil {
		return envio{web: &webpush.Envio{Suscripcion: *d.web, Payload: PayloadWeb(a, colapso), TTL: vida, Tema: colapso}}
	}
	return envio{apns: &apns.Envio{Token: d.token, Entorno: d.entorno, Payload: Payload(a), Colapso: colapso, Vence: time.Now().Add(vida)}}
}

// deTienda: los iPhone y los paneles instalados del equipo de la tienda con sesión vigente (de los
// canales configurados).
func (s *Servicio) deTienda(ctx context.Context, tienda string) ([]destino, error) {
	var out []destino
	err := s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		if s.apnsActivo() {
			filas, err := t.Query(ctx, `SELECT usuario_id::text, rol, token, entorno FROM app.dispositivos_de_tienda()`)
			if err != nil {
				return err
			}
			ds, err := pgx.CollectRows(filas, func(f pgx.CollectableRow) (destino, error) {
				var d destino
				err := f.Scan(&d.usuario, &d.rol, &d.token, &d.entorno)
				return d, err
			})
			if err != nil {
				return err
			}
			out = append(out, ds...)
		}
		if s.webActivo() {
			filas, err := t.Query(ctx, `SELECT usuario_id::text, rol, endpoint, p256dh, auth FROM app.suscripciones_web_de_tienda()`)
			if err != nil {
				return err
			}
			ds, err := pgx.CollectRows(filas, func(f pgx.CollectableRow) (destino, error) {
				d := destino{web: &webpush.Suscripcion{}}
				err := f.Scan(&d.usuario, &d.rol, &d.web.Endpoint, &d.web.P256dh, &d.web.Auth)
				return d, err
			})
			if err != nil {
				return err
			}
			out = append(out, ds...)
		}
		return nil
	})
	return out, err
}

// dePersona: los iPhone y paneles instalados de una persona, con sesión vigente y no inactiva, cuya
// sesión cumple cond (sobre d.sesion_hash, con $1 = sesion).
func (s *Servicio) dePersona(ctx context.Context, usuario, cond, sesion string) ([]destino, error) {
	const vigente = ` AND s.vence_en > now()
		AND NOT (coalesce((u.seguridad->>'inactivas')::boolean, false) AND s.visto_en < now() - $2::interval)`
	var out []destino
	err := s.Pool.EnUsuario(ctx, usuario, func(t pgx.Tx) error {
		if s.apnsActivo() {
			filas, err := t.Query(ctx, `SELECT d.token, d.entorno FROM auth.dispositivos d
				JOIN auth.sesiones s ON s.token_hash = d.sesion_hash
				JOIN auth.usuarios u ON u.id = d.usuario_id
				WHERE `+cond+vigente, sesion, cuentas.Inactividad.String())
			if err != nil {
				return err
			}
			ds, err := pgx.CollectRows(filas, func(f pgx.CollectableRow) (destino, error) {
				var d destino
				err := f.Scan(&d.token, &d.entorno)
				return d, err
			})
			if err != nil {
				return err
			}
			out = append(out, ds...)
		}
		if s.webActivo() {
			filas, err := t.Query(ctx, `SELECT d.endpoint, d.p256dh, d.auth FROM auth.suscripciones_web d
				JOIN auth.sesiones s ON s.token_hash = d.sesion_hash
				JOIN auth.usuarios u ON u.id = d.usuario_id
				WHERE `+cond+vigente, sesion, cuentas.Inactividad.String())
			if err != nil {
				return err
			}
			ds, err := pgx.CollectRows(filas, func(f pgx.CollectableRow) (destino, error) {
				d := destino{web: &webpush.Suscripcion{}}
				err := f.Scan(&d.web.Endpoint, &d.web.P256dh, &d.web.Auth)
				return d, err
			})
			if err != nil {
				return err
			}
			out = append(out, ds...)
		}
		return nil
	})
	return out, err
}

// mandar entrega cada envío. Un token que APNs da por muerto o una suscripción que el servicio de
// push da por vencida (404/410) se borra; si algún envío falló de una forma que vale reintentar
// (servicio caído, 429), devuelve error para que el outbox reintente (apns-collapse-id y Topic
// evitan que se vea dos veces). Lo demás se anota y sigue.
func (s *Servicio) mandar(ctx context.Context, envios []envio) error {
	var reintentar error
	for _, e := range envios {
		var err error
		if e.web != nil {
			err = s.mandarWeb(ctx, *e.web)
		} else {
			err = s.mandarAPNs(ctx, *e.apns)
		}
		if err != nil {
			reintentar = err
		}
	}
	return reintentar
}

// mandarAPNs: devuelve error sólo si vale reintentar.
func (s *Servicio) mandarAPNs(ctx context.Context, e apns.Envio) error {
	err := s.APNs.Enviar(ctx, e)
	var ea *apns.Error
	switch {
	case err == nil:
	case errors.Is(err, apns.ErrTokenInvalido):
		return s.Pool.EnTx(ctx, func(t pgx.Tx) error {
			_, err := t.Exec(ctx, `SELECT app.olvidar_dispositivo($1)`, e.Token)
			return err
		})
	case errors.As(err, &ea) && !ea.Reintentable():
		s.Log.Warn("push: APNs rechazó el aviso", "estado", ea.Estado, "motivo", ea.Motivo)
	default:
		return err
	}
	return nil
}

// mandarWeb: devuelve error sólo si vale reintentar.
func (s *Servicio) mandarWeb(ctx context.Context, e webpush.Envio) error {
	err := s.Web.Enviar(ctx, e)
	var ew *webpush.Error
	switch {
	case err == nil:
	case errors.Is(err, webpush.ErrVencida), errors.Is(err, webpush.ErrInvalida):
		return s.Pool.EnTx(ctx, func(t pgx.Tx) error {
			_, err := t.Exec(ctx, `SELECT app.olvidar_suscripcion_web($1)`, e.Endpoint)
			return err
		})
	case errors.As(err, &ew) && !ew.Reintentable():
		s.Log.Warn("push: el servicio de push web rechazó el aviso", "estado", ew.Estado)
	default:
		return err
	}
	return nil
}

// Venta es lo que dice el push de una venta, según quién lo lee.
type Venta struct {
	Numero           int
	Total            int64
	Productos        int
	Pago, EstadoPago string
}

// AvisoVenta arma «Nueva venta #N» para un rol: el monto sólo con Finanzas (§37); nada de quien
// compró (se ve en la pantalla bloqueada).
func AvisoVenta(tienda, rol string, v Venta) Aviso {
	cuerpo := ""
	if roles.Puede(rol, roles.Finanzas) {
		cuerpo = pesos(v.Total) + " · "
	}
	switch v.Productos {
	case 0:
	case 1:
		cuerpo += "1 producto · "
	default:
		cuerpo += strconv.Itoa(v.Productos) + " productos · "
	}
	switch {
	case v.Pago == "transf" && v.EstadoPago == "Pendiente":
		cuerpo += "Transferencia por confirmar"
	case v.Pago == "transf":
		cuerpo += "Transferencia acreditada"
	case v.Pago == "mp":
		cuerpo += "Mercado Pago acreditado"
	default:
		cuerpo += "Entró por la tienda online"
	}
	return Aviso{Titulo: fmt.Sprintf("Nueva venta #%d", v.Numero), Cuerpo: cuerpo, Hilo: HiloTienda(tienda),
		Ruta: fmt.Sprintf("pedido/%d", v.Numero), Tienda: tienda}
}

// venta: push.venta, que encola el trigger app.avisar_venta_tienda junto con el mail. Sale una sola
// vez (pedidos.avisos->'push_venta'), a quien tiene Pedidos, y no sale si la tienda apagó
// Notificaciones › Venta nueva › Celular (vale para el iPhone y para el panel instalado).
func (s *Servicio) venta(ctx context.Context, ev outbox.Evento) error {
	if !s.activo() {
		return nil
	}
	var d struct {
		Numero int `json:"numero"`
	}
	if json.Unmarshal(ev.Datos, &d) != nil || d.Numero <= 0 {
		return nil
	}
	var v Venta
	var ya, prendido bool
	var estado string
	err := s.Pool.EnTienda(ctx, ev.TiendaID, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT p.numero, p.total, coalesce((SELECT sum((i->>'cantidad')::int) FROM jsonb_array_elements(p.items) i), 0)::int,
				p.pago, p.estado_pago, p.estado, p.avisos ? 'push_venta',
				coalesce(t.preferencias->'canales'->'matriz'->'venta:celular', 'true') <> 'false'
			FROM pedidos p JOIN tiendas t ON t.id = p.tienda_id WHERE p.numero = $1`, d.Numero).
			Scan(&v.Numero, &v.Total, &v.Productos, &v.Pago, &v.EstadoPago, &estado, &ya, &prendido)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if ya || !prendido || estado == "Cancelado" {
		return nil
	}
	dest, err := s.deTienda(ctx, ev.TiendaID)
	if err != nil {
		return err
	}
	var envios []envio
	for _, x := range dest {
		if !roles.Puede(x.rol, roles.Pedidos) {
			continue
		}
		envios = append(envios, para(x, AvisoVenta(ev.TiendaID, x.rol, v), fmt.Sprintf("venta-%d", v.Numero), 24*time.Hour))
	}
	if len(envios) == 0 {
		return nil
	}
	if err := s.mandar(ctx, envios); err != nil {
		return err
	}
	// La Live Activity de los pedidos por preparar (§54): con la app cerrada se actualiza por acá.
	s.actualizarActividades(ctx, ev.TiendaID)
	return s.Pool.EnTienda(ctx, ev.TiendaID, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `UPDATE pedidos SET avisos = avisos || jsonb_build_object('push_venta', $2::text) WHERE numero = $1`,
			d.Numero, time.Now().UTC().Format(time.RFC3339))
		return err
	})
}

// AvisoPreguntaML: el nombre del producto si se sabe; la pregunta no (puede traer datos de quien
// pregunta y se ve en la pantalla bloqueada).
func AvisoPreguntaML(tienda string, id int64, producto string) Aviso {
	cuerpo := "Respondela rápido: ayuda a vender."
	if producto != "" {
		cuerpo = "Te preguntaron por " + producto + "."
	}
	return Aviso{Titulo: "Pregunta nueva en Mercado Libre", Cuerpo: cuerpo, Hilo: HiloTienda(tienda),
		Ruta: fmt.Sprintf("preguntas/ml/%d", id), Tienda: tienda}
}

// preguntaML: push.pregunta_ml, que encola mercadolibre al guardar una pregunta nueva sin
// responder. A quien tiene Clientes (las preguntas son de ahí).
func (s *Servicio) preguntaML(ctx context.Context, ev outbox.Evento) error {
	if !s.activo() {
		return nil
	}
	var d struct {
		ID int64 `json:"id"`
	}
	if json.Unmarshal(ev.Datos, &d) != nil || d.ID <= 0 {
		return nil
	}
	var producto, estado string
	err := s.Pool.EnTienda(ctx, ev.TiendaID, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT coalesce(p.nombre, ''), q.estado FROM ml_preguntas q
			LEFT JOIN ml_publicaciones m ON m.tienda_id = q.tienda_id AND m.item_id = q.item_id
			LEFT JOIN productos p ON p.tienda_id = m.tienda_id AND p.id = m.producto_id
			WHERE q.pregunta_id = $1`, d.ID).Scan(&producto, &estado)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if estado != "UNANSWERED" {
		return nil
	}
	dest, err := s.deTienda(ctx, ev.TiendaID)
	if err != nil {
		return err
	}
	var envios []envio
	for _, x := range dest {
		if roles.Puede(x.rol, roles.Clientes) {
			envios = append(envios, para(x, AvisoPreguntaML(ev.TiendaID, d.ID, producto), fmt.Sprintf("pregunta-ml-%d", d.ID), 24*time.Hour))
		}
	}
	return s.mandar(ctx, envios)
}

// AvisoIngreso: «Entraron a tu cuenta», con el aparato y sin IP ni lugar.
func AvisoIngreso(aparato string) Aviso {
	cuerpo := "Alguien entró a tu cuenta de Tu Local Web. Si no fuiste vos, cerrá esa sesión."
	if aparato != "" {
		cuerpo = "Entraron desde " + aparato + ". Si no fuiste vos, cerrá esa sesión."
	}
	return Aviso{Titulo: "Entraron a tu cuenta", Cuerpo: cuerpo, Hilo: "cuenta", Ruta: "cuenta/seguridad"}
}

// ingreso: cuenta.ingreso_push, sólo a los iPhone y paneles instalados de la persona, menos los de
// la sesión nueva.
func (s *Servicio) ingreso(ctx context.Context, ev outbox.Evento) error {
	if !s.activo() {
		return nil
	}
	var d cuentas.IngresoPush
	if json.Unmarshal(ev.Datos, &d) != nil || d.Usuario == "" {
		return nil
	}
	dest, err := s.dePersona(ctx, d.Usuario, `d.sesion_hash <> $1`, d.Sesion)
	if err != nil {
		return err
	}
	a := AvisoIngreso(d.Aparato)
	var envios []envio
	for _, x := range dest {
		envios = append(envios, para(x, a, fmt.Sprintf("ingreso-%d", ev.ID), 6*time.Hour))
	}
	return s.mandar(ctx, envios)
}

// prueba: cuenta.push_prueba, «Mandame un aviso de prueba» de Notificaciones. Sólo al panel
// instalado de esa sesión (el celular donde se tocó el botón).
func (s *Servicio) prueba(ctx context.Context, ev outbox.Evento) error {
	if !s.webActivo() {
		return nil
	}
	var d Prueba
	if json.Unmarshal(ev.Datos, &d) != nil || d.Usuario == "" || d.Sesion == "" {
		return nil
	}
	dest, err := s.dePersona(ctx, d.Usuario, `d.sesion_hash = $1`, d.Sesion)
	if err != nil {
		return err
	}
	var envios []envio
	for _, x := range dest {
		if x.web != nil {
			envios = append(envios, para(x, AvisoPrueba(), fmt.Sprintf("prueba-%d", ev.ID), 10*time.Minute))
		}
	}
	return s.mandar(ctx, envios)
}
