// Package marca: la marca de la tienda en los mails que reciben sus compradores (aviso de stock,
// recordatorios del carrito): nombre, logo, acento del diseño y a quién se responde. El cuerpo es la
// plantilla de los mails de la tienda (fondo crema, botón redondo), con el botón en el acento.
package marca

import (
	"context"
	"fmt"
	"html"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/acentos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/mitienda"
)

// Marca de una tienda para sus mails.
type Marca struct {
	Nombre, Slug string
	// Logo: URL absoluta (https) o vacío.
	Logo                string
	Acento, SobreAcento string
	// ResponderA: el email del negocio (o el de la dueña); vacío si no sirve como Reply-To.
	ResponderA string
}

const sqlTienda = `SELECT t.nombre, t.slug, t.plantilla, t.diseno_publicado, coalesce(nullif(t.negocio->>'email', ''), (SELECT u.email::text
		FROM auth.membresias m JOIN auth.usuarios u ON u.id = m.usuario_id WHERE m.tienda_id = t.id AND m.rol = 'duenio'
		ORDER BY m.creada_en LIMIT 1), '') FROM tiendas t WHERE t.id = $1`

// De lee la marca, dentro de una transacción ya fijada en la tienda. sitio arma el logo absoluto.
func De(ctx context.Context, t pgx.Tx, tiendaID, sitio string) (Marca, error) {
	var m Marca
	var plantilla string
	var diseno []byte
	if err := t.QueryRow(ctx, sqlTienda, tiendaID).Scan(&m.Nombre, &m.Slug, &plantilla, &diseno, &m.ResponderA); err != nil {
		return m, err
	}
	d := mitienda.LeerDiseno(diseno, plantilla)
	var acento, sobre string
	if d.Estilo != nil {
		acento, sobre = d.Estilo.Colores["acento"], d.Estilo.Colores["sobreAcento"]
	}
	m.Acento, m.SobreAcento = acentos.DeLaTienda(plantilla, d.Paleta, acento, sobre)
	m.Logo = Absoluta(sitio, d.Logo)
	if strings.ContainsAny(m.ResponderA, "\r\n<>,; ") || !strings.Contains(m.ResponderA, "@") {
		m.ResponderA = ""
	}
	return m, nil
}

var unsplashID = regexp.MustCompile(`^[\w-]+$`)

// Absoluta: una foto como URL que abre un cliente de correo (en la demo puede ser un id de Unsplash;
// las subidas son rutas de la web). Vacío si no hay o no sirve.
func Absoluta(sitio, f string) string {
	switch {
	case f == "":
		return ""
	case strings.HasPrefix(f, "https://"):
		return f
	case strings.HasPrefix(f, "/") && !strings.HasPrefix(f, "//"):
		return strings.TrimRight(sitio, "/") + f
	case unsplashID.MatchString(f):
		return "https://images.unsplash.com/photo-" + f + "?auto=format&fit=crop&w=480&h=480&q=75"
	}
	return ""
}

// Mail arma el HTML y el texto con la marca: cabecera (logo o nombre), título, párrafos, una foto
// opcional, el botón principal y el pie (con el link de baja si hay).
type Mail struct {
	Titulo     string
	Parrafos   []string
	Foto       string
	Detalle    string // debajo de la foto (el producto y su precio)
	Boton      string
	Link       string
	Codigo     string // un cupón destacado (opcional)
	CodigoNota string
	Pie        string
	Baja       string // link para darse de baja (opcional)
}

// HTML y texto del mail.
func (m Marca) Armar(c Mail) (string, string) {
	var b, t strings.Builder
	b.WriteString(`<!doctype html><html lang="es-AR"><body style="margin:0;background:#FBF7F2;font-family:Helvetica,Arial,sans-serif;color:#2A140C"><div style="max-width:560px;margin:0 auto;padding:32px 24px">`)
	if m.Logo != "" {
		fmt.Fprintf(&b, `<img src="%s" alt="%s" height="44" style="display:block;height:44px;width:auto;border:0">`, html.EscapeString(m.Logo), html.EscapeString(m.Nombre))
	} else {
		fmt.Fprintf(&b, `<div style="font-size:20px;font-weight:700">%s</div>`, html.EscapeString(m.Nombre))
	}
	fmt.Fprintf(&b, `<h1 style="margin:24px 0 0;font-size:26px;line-height:1.15;font-weight:600">%s</h1>`, html.EscapeString(c.Titulo))
	fmt.Fprintf(&t, "%s\n\n", c.Titulo)
	for _, p := range c.Parrafos {
		fmt.Fprintf(&b, `<p style="margin:12px 0 0;font-size:15.5px;line-height:1.5;color:#5C4536">%s</p>`, html.EscapeString(p))
		fmt.Fprintf(&t, "%s\n\n", p)
	}
	if c.Foto != "" {
		fmt.Fprintf(&b, `<img src="%s" alt="" width="240" style="display:block;margin:18px 0 0;width:240px;max-width:100%%;border-radius:14px;border:0">`, html.EscapeString(c.Foto))
	}
	if c.Detalle != "" {
		fmt.Fprintf(&b, `<p style="margin:10px 0 0;font-size:15px;font-weight:600">%s</p>`, html.EscapeString(c.Detalle))
		fmt.Fprintf(&t, "%s\n\n", c.Detalle)
	}
	if c.Codigo != "" {
		fmt.Fprintf(&b, `<div style="margin:16px 0 0;padding:16px;border-radius:14px;background:#fff;border:2px dashed %s;font-size:24px;font-weight:700;letter-spacing:2px;text-align:center">%s</div>`,
			m.Acento, html.EscapeString(c.Codigo))
		fmt.Fprintf(&t, "Tu código: %s\n", c.Codigo)
		if c.CodigoNota != "" {
			fmt.Fprintf(&b, `<p style="margin:10px 0 0;font-size:13.5px;color:#9C8676">%s</p>`, html.EscapeString(c.CodigoNota))
			fmt.Fprintf(&t, "%s\n", c.CodigoNota)
		}
		t.WriteString("\n")
	}
	if c.Link != "" {
		fmt.Fprintf(&b, `<p style="margin:22px 0 0"><a href="%s" style="display:inline-block;padding:12px 18px;border-radius:999px;background:%s;color:%s;text-decoration:none;font-weight:600">%s</a></p>`,
			html.EscapeString(c.Link), m.Acento, m.SobreAcento, html.EscapeString(c.Boton))
		fmt.Fprintf(&t, "%s: %s\n\n", c.Boton, c.Link)
	}
	pie := html.EscapeString(c.Pie)
	if c.Baja != "" {
		pie += ` <a href="` + html.EscapeString(c.Baja) + `" style="color:#9C8676">No quiero más estos mails</a>.`
	}
	fmt.Fprintf(&b, `<p style="margin:28px 0 0;font-size:12px;line-height:1.5;color:#9C8676">%s Tienda creada con Tu Local Web.</p></div></body></html>`, pie)
	t.WriteString(c.Pie)
	if c.Baja != "" {
		fmt.Fprintf(&t, "\nNo quiero más estos mails: %s", c.Baja)
	}
	return b.String(), t.String()
}
