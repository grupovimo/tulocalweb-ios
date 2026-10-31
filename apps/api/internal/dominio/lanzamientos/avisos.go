package lanzamientos

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

// Tanda: cuántos mails manda cada evento. Si quedan más, encola otro (el outbox los reparte y
// reintenta cada tanda sola).
const Tanda = 50

// Servicio manda los avisos del lanzamiento (worker).
type Servicio struct {
	Pool   tx
	Correo correo.Correo
	// Sitio público, para el link del producto.
	Sitio string
}

// Manejadores del outbox para el worker.
func (s *Servicio) Manejadores() map[string]outbox.Manejador {
	return map[string]outbox.Manejador{"lanzamiento.avisos": s.avisar}
}

type producto struct {
	nombre, slug, tienda, tiendaSlug string
	precio                           int64
	publicado                        bool
	lanzamiento                      *time.Time
}

// avisar manda una tanda de «Ya salió». Cada mail se marca enviado apenas sale: si uno falla, el
// reintento sigue con los que faltan y nadie lo recibe dos veces.
func (s *Servicio) avisar(ctx context.Context, ev outbox.Evento) error {
	var d struct {
		Producto string `json:"producto"`
	}
	if json.Unmarshal(ev.Datos, &d) != nil || d.Producto == "" || s.Correo == nil {
		return nil
	}
	var p producto
	var emails []string
	err := s.Pool.EnTienda(ctx, ev.TiendaID, func(t pgx.Tx) error {
		err := t.QueryRow(ctx, `SELECT p.nombre, p.slug,
				CASE WHEN p.oferta_precio < p.precio AND (p.oferta_hasta IS NULL OR p.oferta_hasta > now()) THEN p.oferta_precio ELSE p.precio END,
				p.publicado, p.lanzamiento, t.nombre, t.slug
			FROM productos p JOIN tiendas t ON t.id = p.tienda_id WHERE p.tienda_id = $1 AND p.id::text = $2`,
			ev.TiendaID, d.Producto).Scan(&p.nombre, &p.slug, &p.precio, &p.publicado, &p.lanzamiento, &p.tienda, &p.tiendaSlug)
		if err != nil {
			return err
		}
		filas, err := t.Query(ctx, `SELECT email FROM avisos_lanzamiento WHERE tienda_id = $1 AND producto_id::text = $2
			AND avisado_en IS NULL ORDER BY creado_en LIMIT $3`, ev.TiendaID, d.Producto, Tanda+1)
		if err != nil {
			return err
		}
		emails, err = pgx.CollectRows(filas, pgx.RowTo[string])
		return err
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	// Lo pausaron o lo reprogramaron: ahora no se avisa (al reprogramarlo se vuelve a encolar).
	if !p.publicado || p.lanzamiento == nil || p.lanzamiento.After(time.Now()) {
		return nil
	}
	quedan := len(emails) > Tanda
	if quedan {
		emails = emails[:Tanda]
	}
	for _, e := range emails {
		if err := s.Correo.Enviar(ctx, mailYaSalio(s.Sitio, p, e)); err != nil {
			return err
		}
		if err := s.Pool.EnTienda(ctx, ev.TiendaID, func(t pgx.Tx) error {
			_, err := t.Exec(ctx, `UPDATE avisos_lanzamiento SET avisado_en = now()
				WHERE tienda_id = $1 AND producto_id::text = $2 AND email = $3`, ev.TiendaID, d.Producto, e)
			return err
		}); err != nil {
			return err
		}
	}
	if !quedan {
		return nil
	}
	return s.Pool.EnTienda(ctx, ev.TiendaID, func(t pgx.Tx) error {
		return outbox.Publicar(ctx, t, ev.TiendaID, "lanzamiento.avisos", d)
	})
}

func pesos(n int64) string {
	s := fmt.Sprint(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "." + s[i:]
	}
	return "$ " + s
}

func mailYaSalio(sitio string, p producto, email string) correo.Mensaje {
	link := ""
	if sitio != "" {
		link = fmt.Sprintf("%s/t/%s/producto/%s", strings.TrimRight(sitio, "/"), p.tiendaSlug, p.slug)
	}
	asunto := fmt.Sprintf("Ya salió %s en %s", p.nombre, p.tienda)
	texto := fmt.Sprintf("Ya salió %s.\n\nTe anotaste para que te avisemos: ya está a la venta en %s (%s).", p.nombre, p.tienda, pesos(p.precio))
	if link != "" {
		texto += "\n\nComprarlo: " + link
	}
	texto += "\n\nTe escribimos sólo por este lanzamiento: no te vamos a mandar más mails por esto."
	var b strings.Builder
	fmt.Fprintf(&b, `<!doctype html><html lang="es-AR"><body style="margin:0;background:#FBF7F2;font-family:Helvetica,Arial,sans-serif;color:#2A140C"><div style="max-width:560px;margin:0 auto;padding:32px 24px"><div style="font-size:20px;font-weight:700">%s</div><h1 style="margin:24px 0 0;font-size:26px;line-height:1.15;font-weight:600">Ya salió %s</h1>`,
		html.EscapeString(p.tienda), html.EscapeString(p.nombre))
	fmt.Fprintf(&b, `<p style="margin:12px 0 0;font-size:15.5px;line-height:1.5;color:#5C4536">Te anotaste para que te avisemos: ya está a la venta en %s, a %s.</p>`,
		html.EscapeString(p.tienda), html.EscapeString(pesos(p.precio)))
	if link != "" {
		fmt.Fprintf(&b, `<p style="margin:22px 0 0"><a href="%s" style="display:inline-block;padding:12px 18px;border-radius:999px;background:#2A140C;color:#fff;text-decoration:none;font-weight:600">Comprarlo</a></p>`, html.EscapeString(link))
	}
	b.WriteString(`<p style="margin:22px 0 0;font-size:13px;color:#7A6252">Te escribimos sólo por este lanzamiento: no te vamos a mandar más mails por esto.</p><p style="margin:28px 0 0;font-size:12px;color:#9C8676">Tienda creada con Tu Local Web.</p></div></body></html>`)
	return correo.Mensaje{Para: email, Asunto: asunto, Texto: texto, HTML: b.String()}
}
