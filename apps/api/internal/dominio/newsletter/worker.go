package newsletter

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

// Mails del newsletter (el worker): la confirmación y las tandas de cada novedad.
type Mails struct {
	Pool   tx
	Correo correo.Correo
	Sitio  string
	// Ahora (tests): la hora de la tanda.
	Ahora func() time.Time
}

func (m *Mails) Manejadores() map[string]outbox.Manejador {
	return map[string]outbox.Manejador{"newsletter.confirmar": m.confirmar, "newsletter.tanda": m.tanda}
}

func (m *Mails) ahora() time.Time {
	if m.Ahora != nil {
		return m.Ahora()
	}
	return time.Now()
}

const sqlTienda = `SELECT t.nombre, t.slug, coalesce(nullif(t.negocio->>'email', ''), (SELECT u.email::text FROM auth.membresias m
		JOIN auth.usuarios u ON u.id = m.usuario_id WHERE m.tienda_id = t.id AND m.rol = 'duenio'
		ORDER BY m.creada_en LIMIT 1), '') FROM tiendas t WHERE t.id = $1`

func (m *Mails) confirmar(ctx context.Context, ev outbox.Evento) error {
	var d struct{ ID string }
	if json.Unmarshal(ev.Datos, &d) != nil || m.Correo == nil {
		return nil
	}
	var t Tienda
	var email, token string
	var confirmado bool
	err := m.Pool.EnTienda(ctx, ev.TiendaID, func(x pgx.Tx) error {
		if err := x.QueryRow(ctx, `SELECT email, token, confirmado_en IS NOT NULL FROM suscriptores WHERE id::text = $1`, d.ID).
			Scan(&email, &token, &confirmado); err != nil {
			return err
		}
		return x.QueryRow(ctx, sqlTienda, ev.TiendaID).Scan(&t.Nombre, &t.Slug, &t.ResponderA)
	})
	if errors.Is(err, pgx.ErrNoRows) || confirmado {
		return nil // se borró o ya confirmó mientras tanto
	}
	if err != nil {
		return err
	}
	return m.Correo.Enviar(ctx, MailConfirmacion(m.Sitio, t, email, token))
}

type destinatario struct {
	email, token string
	activo       bool
}

// tanda manda la siguiente tanda de una novedad, sin pasarse del tope diario de la tienda, y encola
// la que sigue: ya mismo, a los 10 minutos si el servidor de correo rechazó alguno, o mañana a las
// 9 si se llegó al tope. Cada destinatario se toma con su fila bloqueada y se marca apenas sale:
// una tanda repetida no le manda dos veces a nadie.
func (m *Mails) tanda(ctx context.Context, ev outbox.Evento) error {
	var d struct{ Envio string }
	if json.Unmarshal(ev.Datos, &d) != nil || m.Correo == nil {
		return nil
	}
	ahora := m.ahora()
	var t Tienda
	var n Novedad
	var ids []byte
	var terminado bool
	var hoy int
	var ds []destinatario
	var ps []Producto
	err := m.Pool.EnTienda(ctx, ev.TiendaID, func(x pgx.Tx) error {
		if err := x.QueryRow(ctx, `SELECT asunto, texto, productos, terminado IS NOT NULL FROM newsletter_envios WHERE id::text = $1`, d.Envio).
			Scan(&n.Asunto, &n.Texto, &ids, &terminado); err != nil {
			return err
		}
		if terminado {
			return nil
		}
		if err := x.QueryRow(ctx, sqlTienda, ev.TiendaID).Scan(&t.Nombre, &t.Slug, &t.ResponderA); err != nil {
			return err
		}
		if err := x.QueryRow(ctx, `SELECT count(*) FROM newsletter_entregas WHERE estado = 'enviado' AND enviado_en >= $1`,
			inicioDelDia(ahora)).Scan(&hoy); err != nil {
			return err
		}
		libre := min(Tanda, TopeDiario-hoy)
		if libre <= 0 {
			return nil
		}
		filas, err := x.Query(ctx, `SELECT e.email, coalesce(s.token, ''), s.id IS NOT NULL AND s.confirmado_en IS NOT NULL
			FROM newsletter_entregas e LEFT JOIN suscriptores s ON s.id = e.suscriptor_id
			WHERE e.envio_id::text = $1 AND e.estado = 'pendiente' ORDER BY e.email LIMIT $2`, d.Envio, libre)
		if err != nil {
			return err
		}
		for filas.Next() {
			var x destinatario
			if err := filas.Scan(&x.email, &x.token, &x.activo); err != nil {
				filas.Close()
				return err
			}
			ds = append(ds, x)
		}
		filas.Close()
		if err := filas.Err(); err != nil {
			return err
		}
		var lista []string
		_ = json.Unmarshal(ids, &lista)
		if len(lista) == 0 {
			return nil
		}
		// Los productos como están ahora (si alguno se despublicó, no va).
		pf, err := x.Query(ctx, `SELECT id::text, nombre, precio, coalesce(foto, '') FROM productos WHERE id::text = ANY($1) AND publicado`, lista)
		if err != nil {
			return err
		}
		defer pf.Close()
		porID := map[string]Producto{}
		for pf.Next() {
			var p Producto
			if err := pf.Scan(&p.ID, &p.Nombre, &p.Precio, &p.Foto); err != nil {
				return err
			}
			porID[p.ID] = p
		}
		for _, id := range lista {
			if p, ok := porID[id]; ok {
				ps = append(ps, p)
			}
		}
		return pf.Err()
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil // el envío ya no existe
	}
	if err != nil || terminado {
		return err
	}
	fallas, enviados := 0, 0
	for _, x := range ds {
		ok, err := m.mandarUno(ctx, ev.TiendaID, d.Envio, x, t, n, ps, ahora)
		if err != nil {
			return err
		}
		if ok {
			enviados++
		} else {
			fallas++
		}
	}
	return m.seguir(ctx, ev.TiendaID, d.Envio, ahora, hoy+enviados, fallas > 0)
}

// mandarUno toma la fila del destinatario, le manda la novedad y la marca. false: el correo falló.
func (m *Mails) mandarUno(ctx context.Context, tienda, envio string, x destinatario, t Tienda, n Novedad, ps []Producto, ahora time.Time) (bool, error) {
	ok := true
	err := m.Pool.EnTienda(ctx, tienda, func(tx pgx.Tx) error {
		var intentos int
		err := tx.QueryRow(ctx, `SELECT intentos FROM newsletter_entregas WHERE envio_id::text = $1 AND email = $2 AND estado = 'pendiente'
			FOR UPDATE SKIP LOCKED`, envio, x.email).Scan(&intentos)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil // otra tanda ya la tomó
		}
		if err != nil {
			return err
		}
		if !x.activo {
			_, err := tx.Exec(ctx, `UPDATE newsletter_entregas SET estado = 'omitido' WHERE envio_id::text = $1 AND email = $2`, envio, x.email)
			return err
		}
		if err := m.Correo.Enviar(ctx, MailNovedad(m.Sitio, t, n, ps, x.email, x.token)); err != nil {
			slog.Warn("newsletter: no salió un mail", "envio", envio, "intento", intentos+1, "error", err)
			ok = false
			_, err := tx.Exec(ctx, `UPDATE newsletter_entregas SET intentos = intentos + 1,
				estado = CASE WHEN intentos + 1 >= $3 THEN 'fallido' ELSE 'pendiente' END
				WHERE envio_id::text = $1 AND email = $2`, envio, x.email, MaxIntentos)
			return err
		}
		_, err = tx.Exec(ctx, `UPDATE newsletter_entregas SET estado = 'enviado', enviado_en = $3, intentos = intentos + 1
			WHERE envio_id::text = $1 AND email = $2`, envio, x.email, ahora)
		return err
	})
	return ok, err
}

// seguir encola la tanda siguiente o, si no queda nadie, da el envío por terminado.
func (m *Mails) seguir(ctx context.Context, tienda, envio string, ahora time.Time, hoy int, fallas bool) error {
	return m.Pool.EnTienda(ctx, tienda, func(x pgx.Tx) error {
		var pendientes int
		if err := x.QueryRow(ctx, `SELECT count(*) FROM newsletter_entregas WHERE envio_id::text = $1 AND estado = 'pendiente'`, envio).
			Scan(&pendientes); err != nil {
			return err
		}
		if pendientes == 0 {
			_, err := x.Exec(ctx, `UPDATE newsletter_envios SET terminado = $2 WHERE id::text = $1 AND terminado IS NULL`, envio, ahora)
			return err
		}
		cuando := ahora
		switch {
		case hoy >= TopeDiario:
			cuando = inicioDelDia(ahora).Add(24*time.Hour + 9*time.Hour)
		case fallas:
			cuando = ahora.Add(10 * time.Minute)
		}
		datos, _ := json.Marshal(map[string]string{"envio": envio})
		_, err := x.Exec(ctx, `INSERT INTO outbox (tienda_id, tipo, datos, proximo_intento) VALUES ($1, 'newsletter.tanda', $2, $3)`,
			tienda, datos, cuando)
		return err
	})
}
