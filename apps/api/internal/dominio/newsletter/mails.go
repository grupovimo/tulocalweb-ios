package newsletter

import (
	"fmt"
	"html"
	"regexp"
	"strings"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
)

// Producto destacado en una novedad: con foto, precio y link a su página.
type Producto struct {
	ID, Nombre string
	Precio     int64
	Foto       string
}

// Tienda: lo que el mail necesita de la tienda.
type Tienda struct {
	Nombre, Slug string
	// ResponderA: el email del negocio (o el de la dueña): las respuestas le llegan a la tienda.
	ResponderA string
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

var unsplashID = regexp.MustCompile(`^[\w-]+$`)

// fotoAbsoluta: la foto como URL que abre un cliente de correo (en la demo puede ser un id de
// Unsplash; las subidas son rutas de la web).
func fotoAbsoluta(sitio, f string) string {
	switch {
	case f == "":
		return ""
	case strings.HasPrefix(f, "https://"):
		return f
	case strings.HasPrefix(f, "/"):
		return strings.TrimRight(sitio, "/") + f
	case unsplashID.MatchString(f):
		return "https://images.unsplash.com/photo-" + f + "?auto=format&fit=crop&w=480&h=480&q=75"
	}
	return ""
}

func responderA(e string) string {
	if strings.ContainsAny(e, "\r\n<>,; ") || !strings.Contains(e, "@") {
		return ""
	}
	return e
}

const (
	cabecera = `<!doctype html><html lang="es-AR"><body style="margin:0;background:#FBF7F2;font-family:Helvetica,Arial,sans-serif;color:#2A140C"><div style="max-width:560px;margin:0 auto;padding:32px 24px"><div style="font-size:20px;font-weight:700">%s</div><h1 style="margin:24px 0 0;font-size:26px;line-height:1.15;font-weight:600">%s</h1>`
	parrafo  = `<p style="margin:12px 0 0;font-size:15.5px;line-height:1.5;color:#5C4536">%s</p>`
	boton    = `<p style="margin:22px 0 0"><a href="%s" style="display:inline-block;padding:12px 18px;border-radius:999px;background:#2A140C;color:#fff;text-decoration:none;font-weight:600">%s</a></p>`
	pie      = `<p style="margin:28px 0 0;font-size:12px;line-height:1.5;color:#9C8676">%s</p></div></body></html>`
)

// MailConfirmacion: «Confirmá tu suscripción», con el link que la deja activa.
func MailConfirmacion(sitio string, t Tienda, para, token string) correo.Mensaje {
	link := fmt.Sprintf("%s/t/%s/newsletter/confirmar/%s", strings.TrimRight(sitio, "/"), t.Slug, token)
	texto := fmt.Sprintf("Confirmá tu suscripción\n\nPara recibir las novedades de %s, confirmá tu email acá: %s\n\nSi no fuiste vos, ignorá este mail: sin confirmar no te llega nada.\n\n%s",
		t.Nombre, link, t.Nombre)
	var b strings.Builder
	fmt.Fprintf(&b, cabecera, html.EscapeString(t.Nombre), "Confirmá tu suscripción")
	fmt.Fprintf(&b, parrafo, "Para recibir las novedades de "+html.EscapeString(t.Nombre)+", confirmá tu email.")
	fmt.Fprintf(&b, boton, html.EscapeString(link), "Confirmar mi suscripción")
	fmt.Fprintf(&b, pie, "Si no fuiste vos, ignorá este mail: sin confirmar no te llega nada. Tienda creada con Tu Local Web.")
	return correo.Mensaje{Para: para, Asunto: "Confirmá tu suscripción a " + t.Nombre, Texto: texto, HTML: b.String(),
		DeNombre: t.Nombre, ResponderA: responderA(t.ResponderA)}
}

// LinkBaja: la página de baja (y el destino del POST de un clic).
func LinkBaja(sitio, slug, token string) string {
	return fmt.Sprintf("%s/t/%s/newsletter/baja/%s", strings.TrimRight(sitio, "/"), slug, token)
}

// MailNovedad arma la novedad para un destinatario: el texto en párrafos, los productos con foto,
// precio y link, el botón a la tienda y la baja de un clic (link y encabezados).
func MailNovedad(sitio string, t Tienda, n Novedad, ps []Producto, para, token string) correo.Mensaje {
	base := strings.TrimRight(sitio, "/")
	tienda := fmt.Sprintf("%s/t/%s", base, t.Slug)
	baja := LinkBaja(sitio, t.Slug, token)
	var parrafos []string
	for _, p := range strings.Split(n.Texto, "\n\n") {
		if p = strings.TrimSpace(p); p != "" {
			parrafos = append(parrafos, p)
		}
	}
	var tx strings.Builder
	tx.WriteString(n.Asunto + "\n\n" + strings.Join(parrafos, "\n\n"))
	for _, p := range ps {
		fmt.Fprintf(&tx, "\n\n%s — %s\n%s/producto/%s", p.Nombre, pesos(p.Precio), tienda, p.ID)
	}
	fmt.Fprintf(&tx, "\n\nVisitá la tienda: %s\n\n%s\n\nTe llega porque te sumaste a las novedades de %s. Para darte de baja: %s",
		tienda, t.Nombre, t.Nombre, baja)

	var b strings.Builder
	fmt.Fprintf(&b, cabecera, html.EscapeString(t.Nombre), html.EscapeString(n.Asunto))
	for _, p := range parrafos {
		fmt.Fprintf(&b, parrafo, strings.ReplaceAll(html.EscapeString(p), "\n", "<br>"))
	}
	if len(ps) > 0 {
		// Dos por fila en una tabla (los clientes de correo no entienden grid ni flex).
		b.WriteString(`<table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="margin:18px 0 0;border-collapse:separate;border-spacing:0 10px">`)
		for i := 0; i < len(ps); i += 2 {
			b.WriteString("<tr>")
			for j := i; j < i+2; j++ {
				if j >= len(ps) {
					b.WriteString(`<td width="50%"></td>`)
					continue
				}
				p := ps[j]
				link := html.EscapeString(fmt.Sprintf("%s/producto/%s", tienda, p.ID))
				pad := "padding:0 5px 0 0"
				if j%2 == 1 {
					pad = "padding:0 0 0 5px"
				}
				fmt.Fprintf(&b, `<td width="50%%" valign="top" style="%s"><a href="%s" style="display:block;padding:10px;border-radius:14px;background:#fff;color:#2A140C;text-decoration:none">`, pad, link)
				if f := fotoAbsoluta(sitio, p.Foto); f != "" {
					fmt.Fprintf(&b, `<img src="%s" alt="" width="240" style="display:block;width:100%%;height:auto;border-radius:10px">`, html.EscapeString(f))
				}
				fmt.Fprintf(&b, `<div style="margin-top:8px;font-size:14px;font-weight:600;line-height:1.3">%s</div><div style="margin-top:2px;font-size:14px;color:#5C4536">%s</div></a></td>`,
					html.EscapeString(p.Nombre), pesos(p.Precio))
			}
			b.WriteString("</tr>")
		}
		b.WriteString(`</table>`)
	}
	fmt.Fprintf(&b, boton, html.EscapeString(tienda), "Ir a la tienda")
	fmt.Fprintf(&b, pie, "Te llega porque te sumaste a las novedades de "+html.EscapeString(t.Nombre)+
		`. <a href="`+html.EscapeString(baja)+`" style="color:#9C8676">Darme de baja</a>. Tienda creada con Tu Local Web.`)
	return correo.Mensaje{Para: para, Asunto: n.Asunto, Texto: tx.String(), HTML: b.String(),
		DeNombre: t.Nombre, ResponderA: responderA(t.ResponderA),
		Encabezados: map[string]string{
			"List-Unsubscribe":      "<" + baja + ">",
			"List-Unsubscribe-Post": "List-Unsubscribe=One-Click",
		}}
}
