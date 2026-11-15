package promos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/acentos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/mitienda"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

// Mails de las ventanas (el worker): el premio de quien dejó su email, con la marca de la tienda.
type Mails struct {
	Pool   tx
	Correo correo.Correo
	Sitio  string
}

func (m *Mails) Manejadores() map[string]outbox.Manejador {
	return map[string]outbox.Manejador{"promos.premio": m.premio}
}

// DatosMail: lo que lleva el mail del premio.
type DatosMail struct {
	Tienda, Slug, ResponderA string
	Acento, SobreAcento      string
	Para                     string
	// Titulo: el de la ventana («10 % en tu primera compra»).
	Titulo string
	Premio Premio
	// Confirmar: el token para confirmar la newsletter, si todavía no confirmó.
	Confirmar string
}

func fecha(t time.Time) string {
	t = t.In(argentina)
	meses := []string{"enero", "febrero", "marzo", "abril", "mayo", "junio", "julio", "agosto", "septiembre", "octubre", "noviembre", "diciembre"}
	return fmt.Sprintf("%d de %s", t.Day(), meses[t.Month()-1])
}

// MailPremio arma el mail: el código (o la descarga), cuándo vence, el link a la tienda y, si hace
// falta, el botón para confirmar la suscripción.
func MailPremio(sitio string, d DatosMail) correo.Mensaje {
	base := strings.TrimRight(sitio, "/") + "/t/" + d.Slug
	tienda := html.EscapeString(d.Tienda)
	titulo := d.Titulo
	if titulo == "" {
		titulo = "Gracias por sumarte"
	}
	var b, t strings.Builder
	fmt.Fprintf(&b, `<!doctype html><html lang="es-AR"><body style="margin:0;background:#FBF7F2;font-family:Helvetica,Arial,sans-serif;color:#2A140C"><div style="max-width:560px;margin:0 auto;padding:32px 24px"><div style="font-size:20px;font-weight:700">%s</div><h1 style="margin:24px 0 0;font-size:26px;line-height:1.15;font-weight:600">%s</h1>`,
		tienda, html.EscapeString(titulo))
	fmt.Fprintf(&t, "%s\n\n%s\n\n", titulo, d.Tienda)
	boton := func(link, texto string, lleno bool) {
		estilo := fmt.Sprintf("background:%s;color:%s", d.Acento, d.SobreAcento)
		if !lleno {
			estilo = "background:#fff;color:#2A140C;border:1px solid #E5DCD3"
		}
		fmt.Fprintf(&b, `<p style="margin:22px 0 0"><a href="%s" style="display:inline-block;padding:12px 18px;border-radius:999px;%s;text-decoration:none;font-weight:600">%s</a></p>`,
			html.EscapeString(link), estilo, html.EscapeString(texto))
		fmt.Fprintf(&t, "%s: %s\n\n", texto, link)
	}
	p := d.Premio
	switch p.Tipo {
	case "cuponFijo", "cuponUnico":
		intro := "Tu código de descuento"
		if p.Descripcion != "" {
			intro += " (" + p.Descripcion + ")"
		}
		fmt.Fprintf(&b, `<p style="margin:12px 0 0;font-size:15.5px;line-height:1.5;color:#5C4536">%s:</p><div style="margin:12px 0 0;padding:16px;border-radius:14px;background:#fff;border:2px dashed %s;font-size:24px;font-weight:700;letter-spacing:2px;text-align:center">%s</div>`,
			html.EscapeString(intro), d.Acento, html.EscapeString(p.Codigo))
		fmt.Fprintf(&t, "%s: %s\n", intro, p.Codigo)
		if p.Vence != nil {
			v := "Vale hasta el " + fecha(*p.Vence) + ". Es de un solo uso."
			if p.Tipo == "cuponFijo" {
				v = "Vale hasta el " + fecha(*p.Vence) + "."
			}
			fmt.Fprintf(&b, `<p style="margin:10px 0 0;font-size:13.5px;color:#9C8676">%s</p>`, html.EscapeString(v))
			fmt.Fprintf(&t, "%s\n", v)
		}
		t.WriteString("\n")
		boton(base+"/catalogo?cupon="+p.Codigo, "Usarlo en la tienda", true)
	case "descarga":
		nombre := p.NombreArchivo
		if nombre == "" {
			nombre = "el archivo"
		}
		fmt.Fprintf(&b, `<p style="margin:12px 0 0;font-size:15.5px;line-height:1.5;color:#5C4536">Acá tenés %s, como te prometimos.</p>`, html.EscapeString(nombre))
		fmt.Fprintf(&t, "Acá tenés %s, como te prometimos.\n\n", nombre)
		boton(p.Descarga, "Descargar", true)
		boton(base, "Ver la tienda", false)
	default:
		fmt.Fprintf(&b, `<p style="margin:12px 0 0;font-size:15.5px;line-height:1.5;color:#5C4536">Gracias por sumarte a %s. Te vamos a contar las novedades antes que a nadie.</p>`, tienda)
		fmt.Fprintf(&t, "Gracias por sumarte a %s. Te vamos a contar las novedades antes que a nadie.\n\n", d.Tienda)
		boton(base, "Ver la tienda", true)
	}
	if d.Confirmar != "" {
		link := base + "/newsletter/confirmar/" + d.Confirmar
		fmt.Fprintf(&b, `<p style="margin:26px 0 0;font-size:14.5px;line-height:1.5;color:#5C4536">¿Querés recibir las novedades de %s? Confirmá tu email:</p>`, tienda)
		t.WriteString("¿Querés recibir las novedades? Confirmá tu email.\n")
		boton(link, "Confirmar mi suscripción", false)
	}
	pie := "Te escribimos porque dejaste tu email en la tienda de " + d.Tienda + ". Tienda creada con Tu Local Web."
	fmt.Fprintf(&b, `<p style="margin:28px 0 0;font-size:12px;line-height:1.5;color:#9C8676">%s</p></div></body></html>`, html.EscapeString(pie))
	t.WriteString(pie)
	asunto := "Tu premio de " + d.Tienda
	switch p.Tipo {
	case "descarga":
		asunto = "Tu descarga de " + d.Tienda
	case "nada":
		asunto = "Gracias por sumarte a " + d.Tienda
	}
	return correo.Mensaje{Para: d.Para, Asunto: asunto, Texto: t.String(), HTML: b.String(), DeNombre: d.Tienda, ResponderA: d.ResponderA}
}

const sqlTienda = `SELECT t.nombre, t.slug, t.plantilla, t.diseno_publicado, coalesce(nullif(t.negocio->>'email', ''), (SELECT u.email::text
		FROM auth.membresias m JOIN auth.usuarios u ON u.id = m.usuario_id WHERE m.tienda_id = t.id AND m.rol = 'duenio'
		ORDER BY m.creada_en LIMIT 1), '') FROM tiendas t WHERE t.id = $1`

func (m *Mails) premio(ctx context.Context, ev outbox.Evento) error {
	var e struct{ Suscriptor, Ventana string }
	if json.Unmarshal(ev.Datos, &e) != nil || m.Correo == nil {
		return nil
	}
	var d DatosMail
	err := m.Pool.EnTienda(ctx, ev.TiendaID, func(x pgx.Tx) error {
		var pendiente bool
		var token string
		if err := x.QueryRow(ctx, `SELECT email, token, confirmado_en IS NULL FROM suscriptores WHERE id::text = $1`, e.Suscriptor).
			Scan(&d.Para, &token, &pendiente); err != nil {
			return err
		}
		if pendiente {
			d.Confirmar = token
		}
		var crudo []byte
		if err := x.QueryRow(ctx, `SELECT premio FROM promo_captaciones WHERE tienda_id = $1 AND ventana = $2 AND email_hash = $3`,
			ev.TiendaID, e.Ventana, hashEmail(d.Para)).Scan(&crudo); err != nil {
			return err
		}
		if err := json.Unmarshal(crudo, &d.Premio); err != nil {
			return err
		}
		var plantilla string
		var diseno []byte
		if err := x.QueryRow(ctx, sqlTienda, ev.TiendaID).Scan(&d.Tienda, &d.Slug, &plantilla, &diseno, &d.ResponderA); err != nil {
			return err
		}
		dis := mitienda.LeerDiseno(diseno, plantilla)
		var acento, sobre string
		if dis.Estilo != nil {
			acento, sobre = dis.Estilo.Colores["acento"], dis.Estilo.Colores["sobreAcento"]
		}
		d.Acento, d.SobreAcento = acentos.DeLaTienda(plantilla, dis.Paleta, acento, sobre)
		for _, v := range dis.Ventanas {
			if v.ID == e.Ventana {
				d.Titulo = texto(v.Datos["titulo"])
			}
		}
		return nil
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil // se dio de baja o se borró mientras tanto
	}
	if err != nil {
		return err
	}
	if strings.ContainsAny(d.ResponderA, "\r\n<>,; ") || !strings.Contains(d.ResponderA, "@") {
		d.ResponderA = ""
	}
	return m.Correo.Enviar(ctx, MailPremio(m.Sitio, d))
}
