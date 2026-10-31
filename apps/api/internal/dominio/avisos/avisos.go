// Package avisos: los mails a quien compra en la tienda online (y «Nueva venta #N» a la dueña, venta.go). El trigger app.avisar_comprador
// encola "comprador.aviso" en el outbox con el pedido y el tipo (confirmacion, pago, en_camino,
// retiro); acá se arma el mail con los datos del pedido y de la tienda y se manda una sola vez
// (pedidos.avisos guarda cuáles salieron: el outbox entrega al menos una vez).
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

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/configuracion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

type tx interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
}

type Servicio struct {
	Pool   tx
	Correo correo.Correo
	// Sitio público, para el link del pedido.
	Sitio string
}

func (s *Servicio) Manejadores() map[string]outbox.Manejador {
	return map[string]outbox.Manejador{"comprador.aviso": s.manejar, "pregunta.respondida": s.preguntaRespondida,
		"tienda.venta": s.ventaNueva, "tienda.sin_stock": s.sinStock, "arca.vence": s.arcaVence}
}

type item struct {
	Nombre   string `json:"nombre"`
	Cantidad int    `json:"cantidad"`
	Precio   int64  `json:"precio"`
	Variante string `json:"variante"`
}

type pedido struct {
	numero                        int
	email, nombre, pago, envio    string
	estadoPago                    string
	codigo, seguimiento           *string
	items                         []item
	subtotal, envioCosto, total   int64
	tienda, slug                  string
	alias, cbu, titular, whatsapp string
	direccion                     string
	// La tienda: a dónde responde quien compró (el email de Configuración › Negocio o el de la cuenta).
	responderA string
	horario    string
	descuentos []pedidos.Descuento
	// sena: lo que se pagó de seña (§52); 0 si se paga entero.
	sena int64
}

func pesos(n int64) string {
	s := fmt.Sprint(n)
	var b strings.Builder
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(r)
	}
	return "$ " + b.String()
}

func (s *Servicio) manejar(ctx context.Context, ev outbox.Evento) error {
	var d struct {
		Numero int    `json:"numero"`
		Aviso  string `json:"aviso"`
	}
	if err := json.Unmarshal(ev.Datos, &d); err != nil {
		return nil // un evento mal formado no se reintenta
	}
	if s.Correo == nil {
		return nil
	}
	var p pedido
	var ya bool
	err := s.Pool.EnTienda(ctx, ev.TiendaID, func(t pgx.Tx) error {
		var items, comprador, negocio []byte
		err := t.QueryRow(ctx, `SELECT p.numero, p.comprador, p.pago, p.envio, p.estado_pago, p.codigo_publico, p.seguimiento,
				p.items, p.subtotal, p.costo_envio, p.total, p.avisos ? $2, t.nombre, t.slug, t.negocio, p.descuentos, coalesce(p.sena, 0),
				coalesce(nullif(t.negocio->>'email', ''), (SELECT u.email::text FROM auth.membresias m
					JOIN auth.usuarios u ON u.id = m.usuario_id WHERE m.tienda_id = t.id AND m.rol = 'duenio'
					ORDER BY m.creada_en LIMIT 1), '')
			FROM pedidos p JOIN tiendas t ON t.id = p.tienda_id WHERE p.numero = $1`, d.Numero, d.Aviso).
			Scan(&p.numero, &comprador, &p.pago, &p.envio, &p.estadoPago, &p.codigo, &p.seguimiento, &items,
				&p.subtotal, &p.envioCosto, &p.total, &ya, &p.tienda, &p.slug, &negocio, &p.descuentos, &p.sena, &p.responderA)
		if err != nil {
			return err
		}
		var c struct{ Nombre, Email string }
		_ = json.Unmarshal(comprador, &c)
		p.nombre, p.email = strings.Split(strings.TrimSpace(c.Nombre)+" ", " ")[0], c.Email
		_ = json.Unmarshal(items, &p.items)
		p.descuentos = pedidos.DescuentosCompletos(p.descuentos, p.subtotal+p.envioCosto-p.total)
		var n struct{ Alias, CBU, Titular, Whatsapp, Direccion string }
		_ = json.Unmarshal(negocio, &n)
		p.alias, p.cbu, p.titular, p.whatsapp, p.direccion = n.Alias, n.CBU, n.Titular, n.Whatsapp, n.Direccion
		// Dónde se retira: lo de Envíos y retiro manda sobre la dirección del negocio.
		if ent, err := configuracion.EntregasEnTx(ctx, t, ev.TiendaID); err == nil && ent.RetiroDireccion != "" {
			p.direccion, p.horario = ent.RetiroDireccion, ent.RetiroHorario
		}
		return nil
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if ya || p.email == "" {
		return nil
	}
	m, ok := s.mail(p, d.Aviso)
	if !ok {
		return nil
	}
	if err := s.Correo.Enviar(ctx, m); err != nil {
		return err // el outbox reintenta con espera creciente
	}
	return s.Pool.EnTienda(ctx, ev.TiendaID, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `UPDATE pedidos SET avisos = avisos || jsonb_build_object($2::text, $3::text) WHERE numero = $1`,
			d.Numero, d.Aviso, time.Now().UTC().Format(time.RFC3339))
		return err
	})
}

func (s *Servicio) link(p pedido) string {
	if p.codigo == nil {
		return ""
	}
	return fmt.Sprintf("%s/t/%s/pedido/%s", strings.TrimRight(s.Sitio, "/"), p.slug, *p.codigo)
}

// mail arma el aviso. Texto plano y HTML con lo mismo; sin imágenes remotas (se ven igual en todos).
func (s *Servicio) mail(p pedido, aviso string) (correo.Mensaje, bool) {
	var asunto, titulo string
	var parrafos []string
	switch aviso {
	case "confirmacion":
		asunto = fmt.Sprintf("Tu pedido #%d en %s", p.numero, p.tienda)
		titulo = "¡Gracias por tu compra!"
		parrafos = append(parrafos, fmt.Sprintf("Hola, %s. Recibimos tu pedido #%d.", p.nombre, p.numero))
		if p.pago == "transf" && p.estadoPago == "Pendiente" {
			datos := []string{}
			if p.alias != "" {
				datos = append(datos, "Alias: "+p.alias)
			}
			if p.cbu != "" {
				datos = append(datos, "CBU/CVU: "+p.cbu)
			}
			if p.titular != "" {
				datos = append(datos, "Titular: "+p.titular)
			}
			parrafos = append(parrafos, fmt.Sprintf("Para confirmarlo, transferí %s. %s.", pesos(p.total), strings.Join(datos, " · ")),
				"Cuando se acredite te avisamos y empezamos a prepararlo.")
			if p.responderA != "" {
				parrafos = append(parrafos, "Mandanos el comprobante respondiendo este mail.")
			}
		} else if p.sena > 0 && p.estadoPago == "Pendiente" {
			parrafos = append(parrafos, fmt.Sprintf("Para confirmarlo, pagá la seña de %s con Mercado Pago. El resto (%s) lo pagás al retirarlo.",
				pesos(p.sena), pesos(p.total-p.sena)))
		} else if p.estadoPago == "Pendiente" {
			parrafos = append(parrafos, "Cuando se acredite el pago te avisamos y empezamos a prepararlo.")
		}
	case "pago":
		asunto = fmt.Sprintf("Recibimos tu pago · pedido #%d", p.numero)
		titulo = "Recibimos tu pago"
		if p.sena > 0 {
			// Con seña, «pagado» es cuando la tienda cobró el resto al entregarlo (§52).
			parrafos = append(parrafos, fmt.Sprintf("Hola, %s. Quedó pago todo tu pedido #%d. ¡Gracias por tu compra!", p.nombre, p.numero))
		} else {
			parrafos = append(parrafos, fmt.Sprintf("Hola, %s. Se acreditó el pago de tu pedido #%d: ya lo estamos preparando.", p.nombre, p.numero))
		}
	case "sena":
		asunto = fmt.Sprintf("Recibimos tu seña · pedido #%d", p.numero)
		titulo = "Recibimos tu seña"
		parrafos = append(parrafos, fmt.Sprintf("Hola, %s. Se acreditó la seña de %s de tu pedido #%d: ya está reservado.", p.nombre, pesos(p.sena), p.numero),
			fmt.Sprintf("Te falta pagar %s cuando lo retires. Te avisamos cuando esté listo.", pesos(p.total-p.sena)))
	case "en_camino":
		asunto = fmt.Sprintf("Tu pedido #%d está en camino", p.numero)
		titulo = "Tu pedido está en camino"
		parrafos = append(parrafos, fmt.Sprintf("Hola, %s. Despachamos tu pedido #%d.", p.nombre, p.numero))
		if p.seguimiento != nil && *p.seguimiento != "" {
			parrafos = append(parrafos, "Número de seguimiento: "+*p.seguimiento+".")
		}
	case "retiro":
		asunto = fmt.Sprintf("Tu pedido #%d está listo para retirar", p.numero)
		titulo = "Ya podés pasar a buscarlo"
		parrafos = append(parrafos, fmt.Sprintf("Hola, %s. Tu pedido #%d está listo.", p.nombre, p.numero))
		if p.direccion != "" && p.horario != "" {
			parrafos = append(parrafos, "Retiralo en "+p.direccion+" ("+strings.TrimSuffix(p.horario, ".")+").")
		} else if p.direccion != "" {
			parrafos = append(parrafos, "Retiralo en "+p.direccion+".")
		}
	default:
		return correo.Mensaje{}, false
	}
	var det []string
	for _, it := range p.items {
		n := it.Nombre
		if it.Variante != "" {
			n += " (" + it.Variante + ")"
		}
		det = append(det, fmt.Sprintf("%d × %s — %s", it.Cantidad, n, pesos(it.Precio*int64(it.Cantidad))))
	}
	if aviso == "confirmacion" {
		if len(p.descuentos) > 0 || p.envioCosto > 0 {
			det = append(det, "Subtotal — "+pesos(p.subtotal))
		}
		if p.envioCosto > 0 {
			det = append(det, "Envío — "+pesos(p.envioCosto))
		}
		for _, d := range p.descuentos {
			det = append(det, d.Concepto+" — −"+pesos(d.Monto))
		}
		det = append(det, "Total — "+pesos(p.total))
	}
	link := s.link(p)
	contacto := ""
	if p.whatsapp != "" {
		contacto = "¿Dudas? Escribile a " + p.tienda + " por WhatsApp: " + p.whatsapp + "."
	}
	texto := titulo + "\n\n" + strings.Join(parrafos, "\n\n")
	if aviso == "confirmacion" {
		texto += "\n\n" + strings.Join(det, "\n")
	}
	if link != "" {
		texto += "\n\nSeguí tu pedido: " + link
	}
	if contacto != "" {
		texto += "\n\n" + contacto
	}
	texto += "\n\n" + p.tienda
	var b strings.Builder
	fmt.Fprintf(&b, `<!doctype html><html lang="es-AR"><body style="margin:0;background:#FBF7F2;font-family:Helvetica,Arial,sans-serif;color:#2A140C"><div style="max-width:560px;margin:0 auto;padding:32px 24px"><div style="font-size:20px;font-weight:700">%s</div><h1 style="margin:24px 0 0;font-size:26px;line-height:1.15;font-weight:600">%s</h1>`,
		html.EscapeString(p.tienda), html.EscapeString(titulo))
	for _, x := range parrafos {
		fmt.Fprintf(&b, `<p style="margin:12px 0 0;font-size:15.5px;line-height:1.5;color:#5C4536">%s</p>`, html.EscapeString(x))
	}
	if aviso == "confirmacion" {
		b.WriteString(`<div style="margin:18px 0 0;padding:14px 16px;border-radius:14px;background:#fff;font-size:14.5px;line-height:1.7">`)
		for _, x := range det {
			fmt.Fprintf(&b, `<div>%s</div>`, html.EscapeString(x))
		}
		b.WriteString(`</div>`)
	}
	if link != "" {
		fmt.Fprintf(&b, `<p style="margin:22px 0 0"><a href="%s" style="display:inline-block;padding:12px 18px;border-radius:999px;background:#2A140C;color:#fff;text-decoration:none;font-weight:600">Ver mi pedido</a></p>`, html.EscapeString(link))
	}
	if contacto != "" {
		fmt.Fprintf(&b, `<p style="margin:18px 0 0;font-size:13.5px;color:#7A6252">%s</p>`, html.EscapeString(contacto))
	}
	b.WriteString(`<p style="margin:28px 0 0;font-size:12px;color:#9C8676">Tienda creada con Tu Local Web.</p></div></body></html>`)
	// La respuesta (el comprobante de la transferencia, una duda) le llega a la tienda, no a Tu Local Web.
	responder := ""
	if !strings.ContainsAny(p.responderA, "\r\n<>,; ") && strings.Contains(p.responderA, "@") {
		responder = p.responderA
	}
	return correo.Mensaje{Para: p.email, Asunto: asunto, Texto: texto, HTML: b.String(), ResponderA: responder}, true
}
