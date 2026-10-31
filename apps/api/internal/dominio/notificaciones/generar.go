package notificaciones

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/roles"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/auditoria"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

// Texto es lo que ve una persona en la campanita.
type Texto struct {
	Tipo, Grupo, Titulo, Detalle, Ruta string
}

// Miembro es una persona del equipo de la tienda, con su rol.
type Miembro struct {
	Usuario, Rol string
}

// Hecho es lo que sale de un evento: de qué se trata (Clave, para no duplicar) y qué le toca a cada
// persona (nada si su rol no lo puede ver).
type Hecho struct {
	Clave string
	Para  func(Miembro) (Texto, bool)
}

// generador lee el evento con la tienda fijada. Un Hecho sin clave: no hay nada que avisar.
type generador func(context.Context, pgx.Tx, outbox.Evento) (Hecho, error)

// Generador escribe las notificaciones desde el outbox.
type Generador struct {
	Pool Pool
}

// Generadores: los eventos que dejan notificación. Los que ya tenían manejador (mail, push) lo
// conservan; los demás sólo escriben la notificación.
func (g *Generador) generadores() map[string]generador {
	return map[string]generador{
		"tienda.venta":            venta,
		"tienda.sin_stock":        sinStock,
		"push.pregunta_ml":        preguntaML,
		"tienda.resena":           resena,
		"pago.aprobado":           pagoAprobado,
		"integracion.reautorizar": reautorizar,
		"equipo.aceptada":         aceptada,
		"plan.activado":           planActivado,
		"plan.pausado":            planPausado,
		"arca.vence":              arcaVence,
	}
}

// Tipos: los eventos del outbox que generan notificación.
func (g *Generador) Tipos() []string {
	var ts []string
	for t := range g.generadores() {
		ts = append(ts, t)
	}
	return ts
}

// Envolver suma la notificación a los manejadores del worker: primero se escribe (idempotente por
// clave) y después corre el manejador que ya estaba (mail, push). Si la notificación falla, el evento
// se reintenta sin haber mandado nada; si falla el mail, el reintento no la duplica.
func (g *Generador) Envolver(ms map[string]outbox.Manejador) {
	for tipo, gen := range g.generadores() {
		antes := ms[tipo]
		ms[tipo] = func(ctx context.Context, ev outbox.Evento) error {
			if err := g.Generar(ctx, ev, gen); err != nil {
				return err
			}
			if antes != nil {
				return antes(ctx, ev)
			}
			return nil
		}
	}
}

// Generar escribe las notificaciones de un evento, una por persona que puede verlo.
func (g *Generador) Generar(ctx context.Context, ev outbox.Evento, gen generador) error {
	if ev.TiendaID == "" {
		return nil
	}
	err := g.Pool.EnTienda(ctx, ev.TiendaID, func(t pgx.Tx) error {
		h, err := gen(ctx, t, ev)
		if errors.Is(err, pgx.ErrNoRows) || err == nil && h.Clave == "" {
			return nil
		}
		if err != nil {
			return err
		}
		filas, err := t.Query(ctx, `SELECT usuario_id::text, rol FROM auth.membresias WHERE tienda_id = $1 ORDER BY creada_en`, ev.TiendaID)
		if err != nil {
			return err
		}
		ms, err := pgx.CollectRows(filas, func(f pgx.CollectableRow) (Miembro, error) {
			var m Miembro
			err := f.Scan(&m.Usuario, &m.Rol)
			return m, err
		})
		if err != nil {
			return err
		}
		var us, tipos, grupos, titulos, detalles, rutas []string
		for _, m := range ms {
			tx, ok := h.Para(m)
			if !ok {
				continue
			}
			us, tipos, grupos = append(us, m.Usuario), append(tipos, tx.Tipo), append(grupos, tx.Grupo)
			titulos, detalles, rutas = append(titulos, tx.Titulo), append(detalles, tx.Detalle), append(rutas, tx.Ruta)
		}
		if len(us) == 0 {
			return nil
		}
		_, err = t.Exec(ctx, `SELECT app.notificar($1, $2::uuid[], $3, $4, $5, $6, $7)`,
			h.Clave, us, tipos, grupos, titulos, detalles, rutas)
		return err
	})
	if err != nil {
		return fmt.Errorf("notificaciones: %s: %w", ev.Tipo, err)
	}
	return nil
}

// solo: el mismo texto para quien tenga alguno de los permisos.
func solo(tx Texto, ps ...roles.Permiso) func(Miembro) (Texto, bool) {
	return func(m Miembro) (Texto, bool) {
		for _, p := range ps {
			if roles.Puede(m.Rol, p) {
				return tx, true
			}
		}
		return Texto{}, false
	}
}

func cantidad(n int) string {
	if n == 1 {
		return "1 producto"
	}
	return fmt.Sprintf("%d productos", n)
}

// venta: tienda.venta (el trigger de la 00052/00058, sólo la tienda online). Por transferencia, al
// crearse; por Mercado Pago, al acreditarse. A quien tiene Pedidos; el monto sólo con Finanzas.
func venta(ctx context.Context, t pgx.Tx, ev outbox.Evento) (Hecho, error) {
	var d struct {
		Numero int `json:"numero"`
	}
	if json.Unmarshal(ev.Datos, &d) != nil || d.Numero <= 0 {
		return Hecho{}, nil
	}
	var total int64
	var productos int
	var pago, estadoPago, estado string
	err := t.QueryRow(ctx, `SELECT p.total, coalesce((SELECT sum((i->>'cantidad')::int) FROM jsonb_array_elements(p.items) i), 0)::int,
			p.pago, p.estado_pago, p.estado
		FROM pedidos p WHERE p.numero = $1`, d.Numero).Scan(&total, &productos, &pago, &estadoPago, &estado)
	if err != nil || estado == "Cancelado" {
		return Hecho{}, err
	}
	return Hecho{Clave: fmt.Sprintf("venta:%d", d.Numero), Para: func(m Miembro) (Texto, bool) {
		if !roles.Puede(m.Rol, roles.Pedidos) {
			return Texto{}, false
		}
		return TextoVenta(m.Rol, d.Numero, total, productos, pago, estadoPago), true
	}}, nil
}

// RutaPedido: la lista de pedidos con la hoja de ese pedido abierta.
func RutaPedido(numero int) string { return fmt.Sprintf("/panel/pedidos?pedido=%d", numero) }

// TextoVenta: «Nueva venta #N» con lo que el rol puede ver (sin monto para Ventas y Local) y nada
// de quien compró.
func TextoVenta(rol string, numero int, total int64, productos int, pago, estadoPago string) Texto {
	tx := Texto{Tipo: "web", Grupo: "ventas", Titulo: fmt.Sprintf("Nueva venta #%d", numero),
		Ruta: RutaPedido(numero)}
	detalle := ""
	if roles.Puede(rol, roles.Finanzas) {
		detalle = auditoria.Pesos(total) + " · "
	}
	if productos > 0 {
		detalle += cantidad(productos) + " · "
	}
	switch {
	case pago == "transf" && estadoPago == "Pendiente":
		tx.Tipo, detalle = "transf", detalle+"Transferencia por confirmar"
	case pago == "transf":
		tx.Tipo, detalle = "transf", detalle+"Transferencia acreditada"
	case pago == "mp":
		tx.Tipo, detalle = "mp", detalle+"Pago acreditado en Mercado Pago"
	default:
		detalle += "Entró por la tienda online"
	}
	tx.Detalle = detalle
	return tx
}

// sinStock: tienda.sin_stock (§41). A quien tiene Productos; no sale si ya repusieron o lo pausaron.
func sinStock(ctx context.Context, t pgx.Tx, ev outbox.Evento) (Hecho, error) {
	var d pedidos.SinStock
	if json.Unmarshal(ev.Datos, &d) != nil || d.Producto == "" {
		return Hecho{}, nil
	}
	a, err := pedidos.LeerSinStock(ctx, t, d)
	if err != nil || !a.Sigue {
		return Hecho{}, err
	}
	nombre := a.Nombre
	if a.Variante != "" {
		nombre += " (" + a.Variante + ")"
	}
	return Hecho{Clave: fmt.Sprintf("sin_stock:%d", ev.ID), Para: solo(Texto{Tipo: "warn", Grupo: "sistema",
		Titulo: "Te quedaste sin stock", Detalle: nombre + " figura «Sin stock» en la tienda online.",
		Ruta: "/panel/productos?editar=" + url.QueryEscape(d.Producto)}, roles.Productos)}, nil
}

// preguntaML: push.pregunta_ml (§40). A quien tiene Clientes, sin el texto de la pregunta.
func preguntaML(ctx context.Context, t pgx.Tx, ev outbox.Evento) (Hecho, error) {
	var d struct {
		ID int64 `json:"id"`
	}
	if json.Unmarshal(ev.Datos, &d) != nil || d.ID <= 0 {
		return Hecho{}, nil
	}
	var producto, estado string
	err := t.QueryRow(ctx, `SELECT coalesce(p.nombre, ''), q.estado FROM ml_preguntas q
		LEFT JOIN ml_publicaciones m ON m.tienda_id = q.tienda_id AND m.item_id = q.item_id
		LEFT JOIN productos p ON p.tienda_id = m.tienda_id AND p.id = m.producto_id
		WHERE q.pregunta_id = $1`, d.ID).Scan(&producto, &estado)
	if err != nil || estado != "UNANSWERED" {
		return Hecho{}, err
	}
	detalle := "Respondela rápido: ayuda a vender."
	if producto != "" {
		detalle = "Te preguntaron por " + producto + "."
	}
	return Hecho{Clave: fmt.Sprintf("pregunta_ml:%d", d.ID), Para: solo(Texto{Tipo: "ml", Grupo: "ventas",
		Titulo: "Pregunta nueva en Mercado Libre", Detalle: detalle, Ruta: "/panel/resenas-y-preguntas"}, roles.Clientes)}, nil
}

// resena: tienda.resena, que encola la tienda online al dejar una reseña. A quien tiene Clientes,
// con las estrellas y el producto; el texto se lee en la pantalla.
func resena(ctx context.Context, t pgx.Tx, ev outbox.Evento) (Hecho, error) {
	var d struct {
		ID string `json:"id"`
	}
	if json.Unmarshal(ev.Datos, &d) != nil || d.ID == "" {
		return Hecho{}, nil
	}
	var estrellas int
	var producto string
	err := t.QueryRow(ctx, `SELECT r.estrellas, coalesce(p.nombre, '') FROM resenas r
		LEFT JOIN productos p ON p.tienda_id = r.tienda_id AND p.id = r.producto_id WHERE r.id::text = $1`, d.ID).
		Scan(&estrellas, &producto)
	if err != nil {
		return Hecho{}, err
	}
	detalle := fmt.Sprintf("%d de 5 estrellas", estrellas)
	if estrellas == 1 {
		detalle = "1 de 5 estrellas"
	}
	if producto != "" {
		detalle += " para " + producto
	}
	return Hecho{Clave: "resena:" + d.ID, Para: solo(Texto{Tipo: "star", Grupo: "ventas", Titulo: "Reseña nueva",
		Detalle: detalle + ".", Ruta: "/panel/resenas-y-preguntas"}, roles.Clientes)}, nil
}

// pagoAprobado: pago.aprobado, que encola la conciliación de Mercado Pago. Los pedidos de la tienda
// online ya avisan con «Nueva venta» y el QR lo ve quien cobra en el mostrador: acá van los links de
// pago (a quien tiene Finanzas), las señas de turnos y los pedidos cargados en el panel (a Pedidos).
// El monto, sólo con Finanzas.
func pagoAprobado(ctx context.Context, t pgx.Tx, ev outbox.Evento) (Hecho, error) {
	var d struct {
		Referencia string `json:"referencia"`
		Tipo       string `json:"tipo"`
		Objetivo   string `json:"objetivo"`
		Monto      int64  `json:"monto"`
	}
	if json.Unmarshal(ev.Datos, &d) != nil || d.Referencia == "" {
		return Hecho{}, nil
	}
	clave := "pago:" + d.Referencia
	conMonto := func(rol, detalle string) string {
		if roles.Puede(rol, roles.Finanzas) && d.Monto > 0 {
			return auditoria.Pesos(d.Monto) + " · " + detalle
		}
		return detalle
	}
	switch d.Tipo {
	case "link":
		var concepto string
		if err := t.QueryRow(ctx, `SELECT concepto FROM links WHERE codigo = $1`, d.Objetivo).Scan(&concepto); err != nil {
			return Hecho{}, err
		}
		return Hecho{Clave: clave, Para: func(m Miembro) (Texto, bool) {
			if !roles.Puede(m.Rol, roles.Finanzas) {
				return Texto{}, false
			}
			return Texto{Tipo: "mp", Grupo: "ventas", Titulo: "Te pagaron un link", Detalle: conMonto(m.Rol, concepto),
				Ruta: "/panel/finanzas"}, true
		}}, nil
	case "turno":
		var existe bool
		if err := t.QueryRow(ctx, `SELECT true FROM turnos WHERE id::text = $1`, d.Objetivo).Scan(&existe); err != nil {
			return Hecho{}, err
		}
		return Hecho{Clave: clave, Para: func(m Miembro) (Texto, bool) {
			if !roles.Puede(m.Rol, roles.Pedidos) {
				return Texto{}, false
			}
			return Texto{Tipo: "mp", Grupo: "ventas", Titulo: "Seña acreditada",
				Detalle: conMonto(m.Rol, "Pagaron la seña de un turno por Mercado Pago"), Ruta: "/panel/turnos"}, true
		}}, nil
	case "pedido":
		var numero int
		var canal string
		if err := t.QueryRow(ctx, `SELECT numero, canal FROM pedidos WHERE numero::text = $1`, d.Objetivo).Scan(&numero, &canal); err != nil {
			return Hecho{}, err
		}
		if canal == "web" {
			return Hecho{}, nil
		}
		return Hecho{Clave: clave, Para: func(m Miembro) (Texto, bool) {
			if !roles.Puede(m.Rol, roles.Pedidos) {
				return Texto{}, false
			}
			return Texto{Tipo: "mp", Grupo: "ventas", Titulo: fmt.Sprintf("Pago acreditado del pedido #%d", numero),
				Detalle: conMonto(m.Rol, "Mercado Pago"), Ruta: RutaPedido(numero)}, true
		}}, nil
	}
	return Hecho{}, nil
}

// Nombres de las integraciones que pueden pedir reautorizar.
var integraciones = map[string]string{"mp": "Mercado Pago", "ml": "Mercado Libre"}

// reautorizar: integracion.reautorizar, cuando Mercado Pago o Mercado Libre dejan de aceptar la
// conexión (el refresco del token falló). A quien tiene Tienda o Finanzas, como el tiempo real.
func reautorizar(_ context.Context, _ pgx.Tx, ev outbox.Evento) (Hecho, error) {
	var d struct {
		Canal string `json:"canal"`
	}
	if json.Unmarshal(ev.Datos, &d) != nil || integraciones[d.Canal] == "" {
		return Hecho{}, nil
	}
	nombre := integraciones[d.Canal]
	detalle := "Volvé a conectarlo para seguir cobrando con Mercado Pago."
	if d.Canal == "ml" {
		detalle = "Volvé a conectarlo para seguir sincronizando tus publicaciones."
	}
	return Hecho{Clave: fmt.Sprintf("reautorizar:%d", ev.ID), Para: solo(Texto{Tipo: d.Canal, Grupo: "sistema",
		Titulo: nombre + " se desconectó", Detalle: detalle, Ruta: "/panel/integraciones"}, roles.Tienda, roles.Finanzas)}, nil
}

// arcaVence: arca.vence, a 30 días, a 7 y al vencer el certificado (§48). A quien tiene Finanzas.
func arcaVence(_ context.Context, _ pgx.Tx, ev outbox.Evento) (Hecho, error) {
	var d struct {
		Dias int `json:"dias"`
	}
	if json.Unmarshal(ev.Datos, &d) != nil {
		return Hecho{}, nil
	}
	titulo, detalle := TituloArcaVence(d.Dias), "Renovalo para que tus ventas se sigan facturando solas."
	if d.Dias <= 0 {
		detalle = "Tus ventas quedan sin facturar hasta que subas uno nuevo."
	}
	return Hecho{Clave: fmt.Sprintf("arca.vence:%d", ev.ID), Para: solo(Texto{Tipo: "arca", Grupo: "sistema",
		Titulo: titulo, Detalle: detalle, Ruta: "/panel/integraciones?arca"}, roles.Finanzas)}, nil
}

// TituloArcaVence: «Tu certificado de ARCA vence en 7 días» / «vence mañana» / «venció».
func TituloArcaVence(dias int) string {
	switch {
	case dias <= 0:
		return "Tu certificado de ARCA venció"
	case dias == 1:
		return "Tu certificado de ARCA vence mañana"
	}
	return fmt.Sprintf("Tu certificado de ARCA vence en %d días", dias)
}

// aceptada: equipo.aceptada, cuando alguien acepta la invitación. A quien maneja el equipo (Cuenta),
// menos a la persona que se sumó.
func aceptada(ctx context.Context, t pgx.Tx, ev outbox.Evento) (Hecho, error) {
	var d struct {
		Usuario string `json:"usuario"`
	}
	if json.Unmarshal(ev.Datos, &d) != nil || d.Usuario == "" {
		return Hecho{}, nil
	}
	var nombre, rol string
	err := t.QueryRow(ctx, `SELECT u.nombre, m.rol FROM auth.membresias m JOIN auth.usuarios u ON u.id = m.usuario_id
		WHERE m.tienda_id = $1 AND m.usuario_id::text = $2`, ev.TiendaID, d.Usuario).Scan(&nombre, &rol)
	if err != nil {
		return Hecho{}, err
	}
	tx := Texto{Tipo: "ok", Grupo: "sistema", Titulo: nombre + " se sumó al equipo",
		Detalle: "Aceptó la invitación como " + roles.Nombre[rol] + ".", Ruta: "/panel/usuarios-y-permisos"}
	return Hecho{Clave: "aceptada:" + d.Usuario + fmt.Sprintf(":%d", ev.ID), Para: func(m Miembro) (Texto, bool) {
		return tx, m.Usuario != d.Usuario && roles.Puede(m.Rol, roles.Cuenta)
	}}, nil
}

// Nombres de los planes, como los lee una persona.
var planes = map[string]string{"emprende": "Emprende", "pro": "Pro", "negocio": "Negocio"}

// planActivado: plan.activado, cuando Mercado Pago autoriza la suscripción del abono. A Cuenta.
func planActivado(_ context.Context, _ pgx.Tx, ev outbox.Evento) (Hecho, error) {
	var d struct {
		Plan string `json:"plan"`
	}
	if json.Unmarshal(ev.Datos, &d) != nil || planes[d.Plan] == "" {
		return Hecho{}, nil
	}
	return Hecho{Clave: fmt.Sprintf("plan:%d", ev.ID), Para: solo(Texto{Tipo: "ok", Grupo: "sistema",
		Titulo: "Tu plan " + planes[d.Plan] + " está activo", Detalle: "Mercado Pago autorizó el cobro del abono.",
		Ruta: "/panel/plan-y-facturacion"}, roles.Cuenta)}, nil
}

// planPausado: plan.pausado, cuando Mercado Pago pausa la suscripción vigente. A Cuenta.
func planPausado(_ context.Context, _ pgx.Tx, ev outbox.Evento) (Hecho, error) {
	return Hecho{Clave: fmt.Sprintf("plan:%d", ev.ID), Para: solo(Texto{Tipo: "err", Grupo: "sistema",
		Titulo: "Se pausó el cobro de tu plan", Detalle: "Mercado Pago pausó la suscripción. Revisá el medio de pago.",
		Ruta: "/panel/plan-y-facturacion"}, roles.Cuenta)}, nil
}
