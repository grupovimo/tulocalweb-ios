package avisos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

// Los mails de las suscripciones de la tienda (§51): el alta a quien se suscribió (con su link
// privado para verla, pausarla o darla de baja) y a la tienda; la baja a los dos; el cobro rechazado
// a la tienda. Cada uno sale una sola vez (suscripciones.avisos), aunque el outbox repita el evento.

type suscripcion struct {
	codigo, plan, frecuencia, nombre, email, telefono, entrega, direccion string
	precio                                                                int64
	tienda, slug, para, whatsapp                                          string
}

// Cada cuánto, en palabras.
var cadaCuanto = map[string]string{"quincenal": "cada 2 semanas", "mensual": "por mes", "bimestral": "cada 2 meses"}

func (s *Servicio) leerSuscripcion(ctx context.Context, tienda, id string) (suscripcion, error) {
	var x suscripcion
	err := s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT s.codigo, s.plan_nombre, s.frecuencia, s.nombre, s.email, s.telefono, s.entrega, s.direccion, s.precio,
				t.nombre, t.slug, coalesce(t.negocio->>'whatsapp', ''),
				coalesce(nullif(t.negocio->>'email', ''), (SELECT u.email::text FROM auth.membresias m
					JOIN auth.usuarios u ON u.id = m.usuario_id WHERE m.tienda_id = t.id AND m.rol = 'duenio'
					ORDER BY m.creada_en LIMIT 1), '')
			FROM suscripciones s JOIN tiendas t ON t.id = s.tienda_id WHERE s.id = $1`, id).
			Scan(&x.codigo, &x.plan, &x.frecuencia, &x.nombre, &x.email, &x.telefono, &x.entrega, &x.direccion, &x.precio,
				&x.tienda, &x.slug, &x.whatsapp, &x.para)
	})
	return x, err
}

func (s *Servicio) leerAvisos(ctx context.Context, tienda, id string) (map[string]string, error) {
	out := map[string]string{}
	err := s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		var b []byte
		if err := t.QueryRow(ctx, `SELECT avisos FROM suscripciones WHERE id = $1`, id).Scan(&b); err != nil {
			return err
		}
		return json.Unmarshal(b, &out)
	})
	return out, err
}

// mandarUnaVez manda el mail si esa clave no salió todavía y la anota.
func (s *Servicio) mandarUnaVez(ctx context.Context, tienda, id, clave string, ya map[string]string, m correo.Mensaje) error {
	if _, ok := ya[clave]; ok || !emailValido(m.Para) {
		return nil
	}
	if err := s.Correo.Enviar(ctx, m); err != nil {
		return err
	}
	return s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `UPDATE suscripciones SET avisos = avisos || jsonb_build_object($2::text, $3::text) WHERE id = $1`,
			id, clave, time.Now().UTC().Format(time.RFC3339))
		return err
	})
}

func idDe(ev outbox.Evento) (string, string) {
	var d struct{ ID, Por string }
	_ = json.Unmarshal(ev.Datos, &d)
	return d.ID, d.Por
}

// suscripcionAlta: suscripcion.alta (Mercado Pago autorizó el débito o llegó el primer cobro).
func (s *Servicio) suscripcionAlta(ctx context.Context, ev outbox.Evento) error {
	id, _ := idDe(ev)
	if id == "" || s.Correo == nil {
		return nil
	}
	x, err := s.leerSuscripcion(ctx, ev.TiendaID, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	ya, err := s.leerAvisos(ctx, ev.TiendaID, id)
	if err != nil {
		return err
	}
	m := mailAltaComprador(s.Sitio, x)
	m.Para, m.ResponderA = x.email, x.para
	if err := s.mandarUnaVez(ctx, ev.TiendaID, id, "alta_comprador", ya, m); err != nil {
		return err
	}
	m = mailTiendaSuscripcion(s.Sitio, x, "Nuevo suscriptor: "+x.plan,
		fmt.Sprintf("%s se suscribió a %s (%s %s). Cada cobro aprobado te crea el pedido solo.", x.nombre, x.plan, pesos(x.precio), cadaCuanto[x.frecuencia]))
	m.Para, m.ResponderA = x.para, x.email
	return s.mandarUnaVez(ctx, ev.TiendaID, id, "alta_tienda", ya, m)
}

// suscripcionBaja: suscripcion.baja (la canceló quien se suscribió, la tienda o Mercado Pago).
func (s *Servicio) suscripcionBaja(ctx context.Context, ev outbox.Evento) error {
	id, por := idDe(ev)
	if id == "" || s.Correo == nil {
		return nil
	}
	x, err := s.leerSuscripcion(ctx, ev.TiendaID, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	ya, err := s.leerAvisos(ctx, ev.TiendaID, id)
	if err != nil {
		return err
	}
	var quien, aVos string
	switch por {
	case "comprador":
		quien = x.nombre + " dio de baja su suscripción a " + x.plan + "."
		aVos = "Diste de baja tu suscripción a " + x.plan + ". No se te va a cobrar más."
	case "tienda":
		quien = "Cancelaste la suscripción de " + x.nombre + " a " + x.plan + "."
		aVos = x.tienda + " canceló tu suscripción a " + x.plan + ". No se te va a cobrar más."
	default:
		quien = "Mercado Pago canceló la suscripción de " + x.nombre + " a " + x.plan + " (por ejemplo, porque no pudo cobrarla)."
		aVos = "Mercado Pago canceló tu suscripción a " + x.plan + ". No se te va a cobrar más."
	}
	m := mailComprador(s.Sitio, x, "Tu suscripción quedó cancelada", []string{aVos, "Si querés volver, te suscribís de nuevo desde la tienda cuando quieras."}, false)
	m.Para, m.ResponderA = x.email, x.para
	if err := s.mandarUnaVez(ctx, ev.TiendaID, id, "baja_comprador", ya, m); err != nil {
		return err
	}
	m = mailTiendaSuscripcion(s.Sitio, x, "Se dio de baja una suscripción", quien)
	m.Para, m.ResponderA = x.para, x.email
	return s.mandarUnaVez(ctx, ev.TiendaID, id, "baja_tienda", ya, m)
}

// suscripcionRechazo: suscripcion.rechazo (Mercado Pago no pudo cobrar; reintenta solo).
func (s *Servicio) suscripcionRechazo(ctx context.Context, ev outbox.Evento) error {
	id, _ := idDe(ev)
	if id == "" || s.Correo == nil {
		return nil
	}
	x, err := s.leerSuscripcion(ctx, ev.TiendaID, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	ya, err := s.leerAvisos(ctx, ev.TiendaID, id)
	if err != nil {
		return err
	}
	m := mailTiendaSuscripcion(s.Sitio, x, "Mercado Pago no pudo cobrar una suscripción",
		fmt.Sprintf("No se pudo cobrar la suscripción de %s a %s (%s). Mercado Pago lo vuelve a intentar en los próximos días; si querés, escribile para que revise su medio de pago.",
			x.nombre, x.plan, pesos(x.precio)))
	m.Para, m.ResponderA = x.para, x.email
	// Una vez por evento: cada rechazo es un cobro distinto.
	return s.mandarUnaVez(ctx, ev.TiendaID, id, fmt.Sprintf("rechazo_%d", ev.ID), ya, m)
}

// LinkSuscripcion: el link privado de quien se suscribió.
func LinkSuscripcion(sitio, slug, codigo string) string {
	return strings.TrimRight(sitio, "/") + "/t/" + slug + "/suscripcion/" + codigo
}

func mailAltaComprador(sitio string, x suscripcion) correo.Mensaje {
	parrafos := []string{
		fmt.Sprintf("Te suscribiste a %s de %s: %s %s, con débito automático de Mercado Pago.", x.plan, x.tienda, pesos(x.precio), cadaCuanto[x.frecuencia]),
	}
	if x.entrega == "envio" && x.direccion != "" {
		parrafos = append(parrafos, "Te lo mandamos a "+x.direccion+".")
	} else {
		parrafos = append(parrafos, "Lo retirás en el local; te avisamos cuando esté listo.")
	}
	parrafos = append(parrafos, "Con este link ves tu suscripción y la pausás o la das de baja cuando quieras, sin cuenta. Guardalo: es sólo tuyo.")
	return mailComprador(sitio, x, "¡Listo, ya estás suscripto!", parrafos, true)
}

// mailComprador: el mail a quien se suscribió, con el botón a su link privado.
func mailComprador(sitio string, x suscripcion, titulo string, parrafos []string, boton bool) correo.Mensaje {
	link := LinkSuscripcion(sitio, x.slug, x.codigo)
	txt := titulo + "\n\n" + strings.Join(parrafos, "\n\n") + "\n\nTu suscripción: " + link + "\n\n" + x.tienda
	var b strings.Builder
	fmt.Fprintf(&b, `<!doctype html><html lang="es-AR"><body style="margin:0;background:#FBF7F2;font-family:Helvetica,Arial,sans-serif;color:#2A140C"><div style="max-width:560px;margin:0 auto;padding:32px 24px"><div style="font-size:20px;font-weight:700">%s</div><h1 style="margin:24px 0 0;font-size:26px;line-height:1.15;font-weight:600">%s</h1>`,
		html.EscapeString(x.tienda), sinViuda(titulo))
	for _, p := range parrafos {
		fmt.Fprintf(&b, `<p style="margin:12px 0 0;font-size:15.5px;line-height:1.5;color:#5C4536">%s</p>`, sinViuda(p))
	}
	if boton {
		fmt.Fprintf(&b, `<p style="margin:22px 0 0"><a href="%s" style="display:inline-block;padding:12px 18px;border-radius:999px;background:#2A140C;color:#fff;text-decoration:none;font-weight:600">Ver mi suscripción</a></p>`, html.EscapeString(link))
	}
	if x.whatsapp != "" {
		fmt.Fprintf(&b, `<p style="margin:22px 0 0;font-size:14px;line-height:1.5;color:#5C4536">¿Dudas? Respondé este mail o escribinos por WhatsApp al %s.</p>`, html.EscapeString(x.whatsapp))
	}
	b.WriteString(`<p style="margin:28px 0 0;font-size:12px;line-height:1.5;color:#9C8676">Hecho con Tu Local Web.</p></div></body></html>`)
	return correo.Mensaje{Asunto: titulo + " · " + x.tienda, Texto: txt, HTML: b.String()}
}

// mailTiendaSuscripcion: el aviso a la tienda, con el botón a Suscripciones del panel.
func mailTiendaSuscripcion(sitio string, x suscripcion, titulo, texto string) correo.Mensaje {
	link := strings.TrimRight(sitio, "/") + "/panel/suscripciones"
	datos := []string{x.nombre, x.email}
	if x.telefono != "" {
		datos = append(datos, x.telefono)
	}
	txt := titulo + "\n\n" + texto + "\n\n" + strings.Join(datos, " · ") + "\n\nVer suscripciones: " + link + "\n\nTu Local Web"
	var b strings.Builder
	fmt.Fprintf(&b, `<!doctype html><html lang="es-AR"><body style="margin:0;background:#FBF7F2;font-family:Helvetica,Arial,sans-serif;color:#2A140C"><div style="max-width:560px;margin:0 auto;padding:32px 24px"><div style="font-size:20px;font-weight:700">%s</div><h1 style="margin:24px 0 0;font-size:26px;line-height:1.15;font-weight:600">%s</h1>`,
		html.EscapeString(x.tienda), sinViuda(titulo))
	fmt.Fprintf(&b, `<p style="margin:12px 0 0;font-size:15.5px;line-height:1.5;color:#5C4536">%s</p>`, sinViuda(texto))
	fmt.Fprintf(&b, `<p style="margin:12px 0 0;font-size:14px;line-height:1.5;color:#5C4536">%s</p>`, html.EscapeString(strings.Join(datos, " · ")))
	fmt.Fprintf(&b, `<p style="margin:22px 0 0"><a href="%s" style="display:inline-block;padding:12px 18px;border-radius:999px;background:#2A140C;color:#fff;text-decoration:none;font-weight:600">Ver suscripciones</a></p>`, html.EscapeString(link))
	b.WriteString(`<p style="margin:28px 0 0;font-size:12px;line-height:1.5;color:#9C8676">Tu Local Web.</p></div></body></html>`)
	return correo.Mensaje{Asunto: titulo, Texto: txt, HTML: b.String()}
}
