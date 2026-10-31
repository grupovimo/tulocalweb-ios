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

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/roles"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

// arcaVence manda el aviso de vencimiento del certificado de ARCA (evento "arca.vence", §48) a la
// dueña (al email del negocio o, si no hay, al de su cuenta) y a quien del equipo tiene Finanzas.
// Si en el medio subió un certificado nuevo, no se manda.
func (s *Servicio) arcaVence(ctx context.Context, ev outbox.Evento) error {
	var d struct {
		Dias  int       `json:"dias"`
		Vence time.Time `json:"vence"`
	}
	if json.Unmarshal(ev.Datos, &d) != nil || s.Correo == nil {
		return nil
	}
	var tienda string
	var sigue bool
	var para []string
	err := s.Pool.EnTienda(ctx, ev.TiendaID, func(t pgx.Tx) error {
		var negocio string
		if err := t.QueryRow(ctx, `SELECT t.nombre, coalesce(t.negocio->>'email', ''),
				coalesce(c.certificado_vence = $2, false)
			FROM tiendas t LEFT JOIN arca_config c ON c.tienda_id = t.id WHERE t.id = $1`, ev.TiendaID, d.Vence).
			Scan(&tienda, &negocio, &sigue); err != nil {
			return err
		}
		filas, err := t.Query(ctx, `SELECT m.rol, u.email::text FROM auth.membresias m
			JOIN auth.usuarios u ON u.id = m.usuario_id WHERE m.tienda_id = $1 ORDER BY m.creada_en`, ev.TiendaID)
		if err != nil {
			return err
		}
		defer filas.Close()
		visto := map[string]bool{}
		for filas.Next() {
			var rol, email string
			if err := filas.Scan(&rol, &email); err != nil {
				return err
			}
			if !roles.Puede(rol, roles.Finanzas) {
				continue
			}
			if rol == roles.Duenia && strings.TrimSpace(negocio) != "" {
				email = strings.TrimSpace(negocio)
			}
			if k := strings.ToLower(email); emailValido(email) && !visto[k] {
				visto[k] = true
				para = append(para, email)
			}
		}
		return filas.Err()
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if !sigue {
		return nil
	}
	var fallo error
	enviados := 0
	for _, p := range para {
		m := mailArcaVence(s.Sitio, tienda, d.Dias, d.Vence)
		m.Para = p
		if err := s.Correo.Enviar(ctx, m); err != nil {
			fallo = err
			continue
		}
		enviados++
	}
	if enviados == 0 {
		return fallo
	}
	return nil
}

var horaArgentina = time.FixedZone("ART", -3*60*60)

func argentina() *time.Location { return horaArgentina }

// mailArcaVence arma el aviso con el botón para subir el certificado nuevo.
func mailArcaVence(sitio, tienda string, dias int, vence time.Time) correo.Mensaje {
	fecha := vence.In(argentina()).Format("02/01/2006")
	var titulo string
	switch {
	case dias <= 0:
		titulo = "Tu certificado de ARCA venció"
	case dias == 1:
		titulo = "Tu certificado de ARCA vence mañana"
	default:
		titulo = fmt.Sprintf("Tu certificado de ARCA vence en %d días", dias)
	}
	parrafos := []string{
		fmt.Sprintf("El certificado con el que Tu Local Web factura por vos vence el %s.", fecha),
		"Generá uno nuevo en ARCA y subilo en Integraciones › ARCA: son cinco minutos y las ventas se siguen facturando solas.",
	}
	if dias <= 0 {
		parrafos = []string{
			fmt.Sprintf("El certificado con el que Tu Local Web factura por vos venció el %s.", fecha),
			"Hasta que subas uno nuevo, las ventas quedan sin facturar. Generalo en ARCA y subilo en Integraciones › ARCA: las pendientes se facturan solas.",
		}
	}
	link := strings.TrimRight(sitio, "/") + "/panel/integraciones?arca"
	txt := titulo + "\n\n" + strings.Join(parrafos, "\n\n") + "\n\nSubir el certificado: " + link + "\n\nTu Local Web"

	var b strings.Builder
	fmt.Fprintf(&b, `<!doctype html><html lang="es-AR"><body style="margin:0;background:#FBF7F2;font-family:Helvetica,Arial,sans-serif;color:#2A140C"><div style="max-width:560px;margin:0 auto;padding:32px 24px"><div style="font-size:20px;font-weight:700">%s</div><h1 style="margin:24px 0 0;font-size:26px;line-height:1.15;font-weight:600">%s</h1>`,
		html.EscapeString(tienda), html.EscapeString(titulo))
	for _, x := range parrafos {
		fmt.Fprintf(&b, `<p style="margin:12px 0 0;font-size:15.5px;line-height:1.5;color:#5C4536">%s</p>`, sinViuda(x))
	}
	fmt.Fprintf(&b, `<p style="margin:22px 0 0"><a href="%s" style="display:inline-block;padding:12px 18px;border-radius:999px;background:#2A140C;color:#fff;text-decoration:none;font-weight:600">Subir el certificado</a></p>`, html.EscapeString(link))
	b.WriteString(`<p style="margin:28px 0 0;font-size:12px;line-height:1.5;color:#9C8676">Tu Local Web.</p></div></body></html>`)
	return correo.Mensaje{Asunto: titulo, Texto: txt, HTML: b.String()}
}
