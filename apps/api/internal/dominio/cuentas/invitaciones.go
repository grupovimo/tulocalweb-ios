package cuentas

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/roles"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
)

// VidaInvitacion: lo que dura el link de una invitación al equipo de una tienda.
const VidaInvitacion = 7 * 24 * time.Hour

// ErrNoEsTuya: se pidió ver una tienda de la que la persona no es parte.
var ErrNoEsTuya = errors.New("no es parte de esa tienda")

// NuevoToken y HashToken: los mismos tokens que las sesiones (32 bytes al azar; se guarda el sha256).
func NuevoToken() string { return nuevoToken() }

// HashToken es el sha256 del token, como se guarda.
func HashToken(t string) string { return hashToken(t) }

// MailInvitacion son los datos del mail de una invitación (evento cuenta.invitacion del outbox).
type MailInvitacion struct {
	Para   string `json:"para"`
	Tienda string `json:"tienda"`
	Invita string `json:"invita"`
	Rol    string `json:"rol"`
	Link   string `json:"link"`
	Vence  string `json:"vence"`
}

// EncolarInvitacion guarda el link de la invitación (un solo uso, vence en 7 días; uno solo vigente
// por invitación) y encola el mail, en la transacción de quien invita. sitio arma el link.
func EncolarInvitacion(ctx context.Context, tx pgx.Tx, sitio, tiendaID, invitacionID string, d MailInvitacion) error {
	token := nuevoToken()
	var vence time.Time
	if _, err := tx.Exec(ctx, `DELETE FROM auth.invitaciones WHERE invitacion_id = $1`, invitacionID); err != nil {
		return err
	}
	if err := tx.QueryRow(ctx, `INSERT INTO auth.invitaciones (token_hash, invitacion_id, tienda_id, vence_en)
		VALUES ($1, $2, $3, now() + $4::interval) RETURNING vence_en`, hashToken(token), invitacionID, tiendaID, VidaInvitacion.String()).
		Scan(&vence); err != nil {
		return err
	}
	d.Link = sitio + "/invitacion?token=" + token
	d.Vence = HoraAR(vence)
	if n, ok := roles.Nombre[d.Rol]; ok {
		d.Rol = n
	}
	b, err := json.Marshal(d)
	if err != nil {
		return err
	}
	// Sin tienda, como los otros mails de cuentas: la tienda no ve el outbox con el link.
	_, err = tx.Exec(ctx, `INSERT INTO outbox (tienda_id, tipo, datos) VALUES (NULL, 'cuenta.invitacion', $1)`, b)
	return err
}

// Invitacion es lo que ve quien abre el link (components/schemas/Invitacion).
type Invitacion struct {
	Tienda      string    `json:"tienda"`
	Rol         string    `json:"rol"`
	Email       string    `json:"email"`
	Invita      string    `json:"invita"`
	Vence       time.Time `json:"vence"`
	TieneCuenta bool      `json:"tieneCuenta"`
}

// invitacionVigente lee la invitación del token y deja la tienda fijada en la transacción.
func invitacionVigente(ctx context.Context, tx pgx.Tx, token string) (tiendaID, invitacionID string, inv Invitacion, err error) {
	if token == "" || len(token) > 200 {
		return "", "", inv, ErrTokenInvalido
	}
	err = tx.QueryRow(ctx, `SELECT tienda_id::text, invitacion_id::text, vence_en FROM auth.invitaciones
		WHERE token_hash = $1 AND vence_en > now()`, hashToken(token)).Scan(&tiendaID, &invitacionID, &inv.Vence)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", inv, ErrTokenInvalido
	}
	if err != nil {
		return "", "", inv, err
	}
	if _, err = tx.Exec(ctx, "SELECT set_config('app.tienda_id', $1, true)", tiendaID); err != nil {
		return "", "", inv, err
	}
	err = tx.QueryRow(ctx, `SELECT t.nombre, i.rol, i.contacto, coalesce(u.nombre, '')
		FROM invitaciones i JOIN tiendas t ON t.id = i.tienda_id LEFT JOIN auth.usuarios u ON u.id = i.invitada_por
		WHERE i.id = $1 AND t.cerrada_en IS NULL`, invitacionID).Scan(&inv.Tienda, &inv.Rol, &inv.Email, &inv.Invita)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", inv, ErrTokenInvalido
	}
	return tiendaID, invitacionID, inv, err
}

// VerInvitacion dice de qué tienda y con qué rol es la invitación, sin usarla.
func (s *Servicio) VerInvitacion(ctx context.Context, token string) (Invitacion, error) {
	var inv Invitacion
	err := s.Pool.EnTx(ctx, func(tx pgx.Tx) error {
		var err error
		if _, _, inv, err = invitacionVigente(ctx, tx, token); err != nil {
			return err
		}
		return tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM auth.usuarios WHERE email = $1)`, inv.Email).Scan(&inv.TieneCuenta)
	})
	return inv, err
}

// sumar usa la invitación: la borra (un solo uso) y suma a la persona a la tienda con ese rol. Si ya
// era del equipo, queda con el rol que tenía.
func sumar(ctx context.Context, tx pgx.Tx, usuario, token string) (string, error) {
	tiendaID, invitacionID, inv, err := invitacionVigente(ctx, tx, token)
	if err != nil {
		return "", err
	}
	// Sólo la cuenta del email invitado: un link reenviado o abierto con otra sesión no suma a nadie más.
	var propio bool
	if err := tx.QueryRow(ctx, `SELECT lower(email) = lower($2) FROM auth.usuarios WHERE id = $1`, usuario, inv.Email).
		Scan(&propio); err != nil {
		return "", err
	}
	if !propio {
		return "", ErrOtroEmail
	}
	if _, err := tx.Exec(ctx, `DELETE FROM invitaciones WHERE id = $1`, invitacionID); err != nil {
		return "", err
	}
	tag, err := tx.Exec(ctx, `INSERT INTO auth.membresias (usuario_id, tienda_id, rol) VALUES ($1, $2, $3)
		ON CONFLICT (usuario_id, tienda_id) DO NOTHING`, usuario, tiendaID, inv.Rol)
	if err != nil {
		return "", err
	}
	// Se sumó alguien nuevo: la campanita de quien maneja el equipo (§45). Con la tienda todavía fijada.
	if tag.RowsAffected() == 1 {
		if _, err := tx.Exec(ctx, `INSERT INTO outbox (tienda_id, tipo, datos) VALUES ($1, 'equipo.aceptada', jsonb_build_object('usuario', $2::text))`,
			tiendaID, usuario); err != nil {
			return "", err
		}
	}
	_, err = tx.Exec(ctx, "SELECT set_config('app.tienda_id', '', true)")
	return tiendaID, err
}

// AceptarInvitacion suma a la persona de la sesión a la tienda y deja la sesión en esa tienda.
func (s *Servicio) AceptarInvitacion(ctx context.Context, sesionToken, token string) (string, error) {
	var tiendaID string
	err := s.Pool.EnTx(ctx, func(tx pgx.Tx) error {
		var usuario string
		err := tx.QueryRow(ctx, `SELECT usuario_id::text FROM auth.sesiones WHERE token_hash = $1 AND vence_en > now()`,
			hashToken(sesionToken)).Scan(&usuario)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrSinSesion
		}
		if err != nil {
			return err
		}
		if tiendaID, err = sumar(ctx, tx, usuario, token); err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `UPDATE auth.sesiones SET tienda_id = $2 WHERE token_hash = $1`, hashToken(sesionToken), tiendaID)
		return err
	})
	return tiendaID, err
}

// RegistrarConInvitacion crea la cuenta con el email de la invitación (queda verificado: le llegó
// el link) y la suma a la tienda, sin crear una tienda propia.
func (s *Servicio) RegistrarConInvitacion(ctx context.Context, token, nombre, clave string, cl Cliente) (Sesion, error) {
	nombre, err := texto("nombre", nombre, "¿Cómo te llamás?")
	if err != nil {
		return Sesion{}, err
	}
	if err := validarClave(clave); err != nil {
		return Sesion{}, err
	}
	if ok, err := s.permitido(ctx, "registro:ip:"+cl.IP, 10, time.Hour); err != nil || !ok {
		return Sesion{}, errOr(err, ErrDemasiadosIntentos)
	}
	hash, err := HashClave(clave)
	if err != nil {
		return Sesion{}, err
	}
	var ses Sesion
	err = s.Pool.EnTx(ctx, func(tx pgx.Tx) error {
		_, _, inv, err := invitacionVigente(ctx, tx, token)
		if err != nil {
			return err
		}
		email, err := normalizarEmail(inv.Email)
		if err != nil {
			return ErrTokenInvalido // una invitación vieja por WhatsApp: no hay email para la cuenta
		}
		var usuario string
		err = tx.QueryRow(ctx, `INSERT INTO auth.usuarios (email, nombre, clave_hash, email_verificado_en) VALUES ($1, $2, $3, now())
			RETURNING id::text`, email, nombre, hash).Scan(&usuario)
		var pg *pgconn.PgError
		if errors.As(err, &pg) && pg.Code == "23505" {
			return ErrEmailEnUso
		}
		if err != nil {
			return err
		}
		tiendaID, err := sumar(ctx, tx, usuario, token)
		if err != nil {
			return err
		}
		ses, err = crearSesion(ctx, tx, usuario, tiendaID, SesionLarga, cl)
		return err
	})
	return ses, err
}

// TiendaDeLaPersona es una de las tiendas de la persona (components/schemas/Acceso.tiendas).
type TiendaDeLaPersona struct {
	ID     string `json:"id"`
	Nombre string `json:"nombre"`
	Slug   string `json:"slug"`
	Rol    string `json:"rol"`
	// Logo: el que se cargó en Mi tienda (foto subida o URL); null si no tiene.
	Logo *string `json:"logo"`
}

// Acceso: la tienda que se ve, el rol con sus permisos y todas las tiendas de la persona.
type Acceso struct {
	TiendaID string              `json:"tiendaId"`
	Rol      string              `json:"rol"`
	Permisos []roles.Permiso     `json:"permisos"`
	Tiendas  []TiendaDeLaPersona `json:"tiendas"`
}

// AccesoDe arma el acceso de la sesión. Las sesiones de una IA o de alguien del equipo de Tu Local
// Web no listan otras tiendas.
func (s *Servicio) AccesoDe(ctx context.Context, d Datos) (Acceso, error) {
	a := Acceso{TiendaID: d.TiendaID, Rol: d.Rol, Permisos: roles.De(d.Rol), Tiendas: []TiendaDeLaPersona{}}
	if d.IA != "" || d.Mirada {
		return a, nil
	}
	filas, err := s.Pool.Query(ctx, `SELECT id::text, nombre, slug, rol, logo FROM app.tiendas_de_usuario($1)`, d.Usuario.ID)
	if err != nil {
		return a, err
	}
	defer filas.Close()
	for filas.Next() {
		var t TiendaDeLaPersona
		if err := filas.Scan(&t.ID, &t.Nombre, &t.Slug, &t.Rol, &t.Logo); err != nil {
			return a, err
		}
		a.Tiendas = append(a.Tiendas, t)
	}
	return a, filas.Err()
}

// ElegirTienda cambia la tienda de la sesión, si la persona es parte de esa tienda.
func (s *Servicio) ElegirTienda(ctx context.Context, sesionToken, tiendaID string) error {
	tag, err := s.Pool.Exec(ctx, `UPDATE auth.sesiones s SET tienda_id = m.tienda_id
		FROM auth.membresias m
		WHERE s.token_hash = $1 AND s.vence_en > now() AND m.usuario_id = s.usuario_id AND m.tienda_id::text = $2`,
		hashToken(sesionToken), tiendaID)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNoEsTuya
	}
	return err
}

// mailInvitacion: «Te invitaron a <tienda>». Misma plantilla que los otros mails de cuentas.
func mailInvitacion(d MailInvitacion) correo.Mensaje {
	quien := "Te invitaron"
	if d.Invita != "" {
		quien = d.Invita + " te invitó"
	}
	t := html.EscapeString(d.Tienda)
	return correo.Mensaje{
		Para:   d.Para,
		Asunto: fmt.Sprintf("%s a %s en Tu Local Web", quien, d.Tienda),
		Texto: fmt.Sprintf("Hola.\n\n%s a sumarte al equipo de %s en Tu Local Web, con el rol %s.\n\n"+
			"Para aceptar entrá a este link (sirve una sola vez y vence el %s):\n%s\n\n"+
			"Si ya tenés cuenta, entrás con ella; si no, la creás ahí mismo.\n\nSi no esperabas esta invitación, ignorá este mail.\n\nTu Local Web",
			quien, d.Tienda, d.Rol, d.Vence, d.Link),
		HTML: fmt.Sprintf(`<!doctype html><html lang="es-AR"><body style="margin:0;background:#FBF7F2;font-family:'Instrument Sans',Helvetica,Arial,sans-serif;color:#2A140C">
<div style="max-width:560px;margin:0 auto;padding:32px 24px">
<div style="font-size:22px;font-weight:700;letter-spacing:-.03em">Tu Local Web</div>
<h1 style="margin:28px 0 0;font-size:28px;line-height:1.1;letter-spacing:-.03em;font-weight:600">Sumate al equipo de&nbsp;%s</h1>
<p style="margin:12px 0 0;font-size:15.5px;line-height:1.5;color:#5C4536">%s a sumarte al equipo de <b>%s</b> en Tu Local Web, con el rol <b>%s</b>. Si ya tenés cuenta, entrás con ella; si no, la creás ahí&nbsp;mismo.</p>
<p style="margin:24px 0"><a href="%s" style="display:inline-block;padding:14px 22px;border-radius:999px;background:#D2582F;color:#fff;font-weight:600;text-decoration:none">Aceptar la invitación</a></p>
<p style="margin:0;font-size:13.5px;line-height:1.5;color:#7A6252">El link sirve una sola vez y vence el %s. Si no esperabas esta invitación, ignorá este&nbsp;mail.</p>
</div></body></html>`, t, html.EscapeString(quien), t, html.EscapeString(d.Rol), html.EscapeString(d.Link), html.EscapeString(d.Vence)),
	}
}
