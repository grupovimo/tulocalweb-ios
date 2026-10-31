package suscripciones

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/auditoria"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

func (a *Alta) validar(p Plan) error {
	a.Nombre, a.Email = strings.Join(strings.Fields(a.Nombre), " "), strings.TrimSpace(a.Email)
	a.Telefono, a.Direccion, a.CP = strings.TrimSpace(a.Telefono), strings.Join(strings.Fields(a.Direccion), " "), strings.TrimSpace(a.CP)
	if largo(a.Nombre) < 2 || largo(a.Nombre) > 80 || !emailValido.MatchString(a.Email) || len(a.Email) > 120 ||
		!telValido.MatchString(a.Telefono) || len(a.CP) > 10 {
		return ErrInvalido
	}
	switch a.Entrega {
	case "envio":
		if !p.Envio || largo(a.Direccion) < 4 || largo(a.Direccion) > 200 {
			return ErrInvalido
		}
	case "retiro":
		if !p.Retiro {
			return ErrInvalido
		}
		a.Direccion, a.CP = "", ""
	default:
		return ErrInvalido
	}
	return nil
}

// Suscribir: el alta desde la tienda. Mira el cupo con el plan bloqueado, guarda la suscripción
// pendiente con el precio del plan y arma el débito en Mercado Pago (devuelve adónde autorizarlo).
// Sin Mercado Pago conectado, ErrSinCobros. Si Mercado Pago no responde, la suscripción queda
// cancelada en el momento (no ocupa cupo) y ErrMercadoPago.
func Suscribir(ctx context.Context, pool tx, mp Cobrador, sitio, slug string, a Alta) (Resultado, error) {
	if mp == nil {
		return Resultado{}, ErrSinCobros
	}
	tienda, err := tiendaDe(ctx, pool, slug)
	if err != nil {
		return Resultado{}, err
	}
	if !idValido.MatchString(a.PlanID) {
		return Resultado{}, ErrNoExiste
	}
	codigo := nuevoCodigo()
	var id string
	err = pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		ok, err := conMercadoPago(ctx, t, tienda)
		if err != nil {
			return err
		}
		if !ok {
			return ErrSinCobros
		}
		p, err := escanearPlan(t.QueryRow(ctx, `SELECT `+columnasPlan+` FROM suscripcion_planes p WHERE p.id = $1 AND p.activo FOR UPDATE`, a.PlanID))
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoExiste
		}
		if err != nil {
			return err
		}
		if err := a.validar(p); err != nil {
			return err
		}
		if p.Cupo != nil {
			var n int
			if err := t.QueryRow(ctx, vigentes, p.ID).Scan(&n); err != nil {
				return err
			}
			if n >= *p.Cupo {
				return ErrCupo
			}
		}
		return t.QueryRow(ctx, `INSERT INTO suscripciones (tienda_id, plan_id, codigo, plan_nombre, precio, frecuencia, email, nombre,
				telefono, entrega, direccion, cp)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id::text`,
			tienda, p.ID, codigo, p.Nombre, p.Precio, p.Frecuencia, a.Email, a.Nombre, a.Telefono, a.Entrega, a.Direccion, a.CP).Scan(&id)
	})
	if err != nil {
		return Resultado{}, err
	}
	url, err := mp.SuscribirCliente(ctx, tienda, id, fmt.Sprintf("%s/t/%s/suscripcion/%s", sitio, slug, codigo))
	if err != nil {
		_ = pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
			_, err := t.Exec(ctx, `UPDATE suscripciones SET estado = 'cancelada', cancelada_por = 'mercadopago', cancelada_en = now(),
				actualizada_en = now() WHERE id = $1 AND estado = 'pendiente'`, id)
			return err
		})
		return Resultado{}, fmt.Errorf("%w: %w", ErrMercadoPago, err)
	}
	return Resultado{Codigo: codigo, URL: url}, nil
}

// Ver la suscripción por su link privado.
func Ver(ctx context.Context, pool tx, slug, codigo string) (Publica, error) {
	if !codigoValido.MatchString(codigo) {
		return Publica{}, ErrNoExiste
	}
	tienda, err := tiendaDe(ctx, pool, slug)
	if err != nil {
		return Publica{}, err
	}
	var out Publica
	err = pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		return ver(ctx, t, codigo, &out)
	})
	return out, err
}

func ver(ctx context.Context, t pgx.Tx, codigo string, out *Publica) error {
	var nombre string
	err := t.QueryRow(ctx, `SELECT s.codigo, s.plan_id::text, s.plan_nombre, p.foto, p.incluye, s.precio, s.frecuencia, s.nombre, s.entrega,
			s.direccion, s.estado, s.proximo_cobro, s.cobros, s.creada_en, ti.nombre, coalesce(ti.negocio->>'whatsapp', '')
		FROM suscripciones s JOIN suscripcion_planes p ON p.id = s.plan_id JOIN tiendas ti ON ti.id = s.tienda_id
		WHERE s.codigo = $1`, codigo).Scan(&out.Codigo, &out.PlanID, &out.Plan, &out.Foto, &out.Incluye, &out.Precio, &out.Frecuencia,
		&nombre, &out.Entrega, &out.Direccion, &out.Estado, &out.ProximoCobro, &out.Cobros, &out.CreadaEn, &out.Tienda, &out.Whatsapp)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNoExiste
	}
	// Sólo el primer nombre: el link puede reenviarse.
	out.Nombre = strings.Split(nombre+" ", " ")[0]
	return err
}

// estadoMP: lo que se le pide a Mercado Pago para cada acción.
var estadoMP = map[string]string{"pausar": "paused", "reanudar": "authorized", "cancelar": "cancelled"}

// cambiar aplica pausar, reanudar o cancelar a una suscripción (por id). Primero en Mercado Pago
// (si ya hay débito armado) y recién después en la base: si Mercado Pago no responde, nada cambia.
// Cancelar una cancelada no es error.
func cambiar(ctx context.Context, pool tx, mp Cobrador, tienda, id, accion, por string) error {
	destino, ok := estadoMP[accion]
	if !ok {
		return ErrInvalido
	}
	var estado string
	var preapproval *string
	if err := pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT estado, mp_preapproval FROM suscripciones WHERE id = $1`, id).Scan(&estado, &preapproval)
	}); errors.Is(err, pgx.ErrNoRows) {
		return ErrNoExiste
	} else if err != nil {
		return err
	}
	switch {
	case estado == Cancelada && accion == "cancelar":
		return nil
	case accion == "pausar" && estado != Activa, accion == "reanudar" && estado != Pausada, estado == Cancelada:
		return ErrEstado
	}
	if preapproval != nil {
		if mp == nil {
			return ErrSinCobros
		}
		if err := mp.CambiarSuscripcionCliente(ctx, tienda, id, destino); err != nil {
			return fmt.Errorf("%w: %w", ErrMercadoPago, err)
		}
	}
	nuevo := map[string]string{"pausar": Pausada, "reanudar": Activa, "cancelar": Cancelada}[accion]
	return pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		var planNombre string
		err := t.QueryRow(ctx, `UPDATE suscripciones SET estado = $2, actualizada_en = now(),
				cancelada_por = CASE WHEN $2 = 'cancelada' THEN $3 ELSE cancelada_por END,
				cancelada_en = CASE WHEN $2 = 'cancelada' THEN now() ELSE cancelada_en END,
				proximo_cobro = CASE WHEN $2 = 'activa' THEN proximo_cobro ELSE NULL END
			WHERE id = $1 AND estado <> 'cancelada' RETURNING plan_nombre`, id, nuevo, por).Scan(&planNombre)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		if nuevo != Cancelada {
			return nil
		}
		// Una alta que nunca se autorizó no avisa baja: no había suscripción.
		if estado != Pendiente {
			if err := outbox.Publicar(ctx, t, tienda, "suscripcion.baja", map[string]string{"id": id, "por": por}); err != nil {
				return err
			}
		}
		if por == "tienda" {
			return auditoria.Registrar(ctx, t, tienda, auditoria.Registro{Accion: "cobros.suscripcion_cancelar", Entidad: "suscripcion",
				EntidadID: id, Texto: "canceló una suscripción a «" + planNombre + "»"})
		}
		return nil
	})
}

// CambiarPublica: quien se suscribió pausa, reanuda o cancela desde su link privado.
func CambiarPublica(ctx context.Context, pool tx, mp Cobrador, slug, codigo, accion string) (Publica, error) {
	if !codigoValido.MatchString(codigo) {
		return Publica{}, ErrNoExiste
	}
	tienda, err := tiendaDe(ctx, pool, slug)
	if err != nil {
		return Publica{}, err
	}
	var id string
	if err := pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT id::text FROM suscripciones WHERE codigo = $1`, codigo).Scan(&id)
	}); errors.Is(err, pgx.ErrNoRows) {
		return Publica{}, ErrNoExiste
	} else if err != nil {
		return Publica{}, err
	}
	if err := cambiar(ctx, pool, mp, tienda, id, accion, "comprador"); err != nil {
		return Publica{}, err
	}
	return Ver(ctx, pool, slug, codigo)
}

// CancelarDesdePanel: la tienda da de baja una suscripción (deja de cobrarse en Mercado Pago).
func CancelarDesdePanel(ctx context.Context, pool tx, mp Cobrador, tienda, id string) (Suscriptor, error) {
	if !idValido.MatchString(id) {
		return Suscriptor{}, ErrNoExiste
	}
	if err := cambiar(ctx, pool, mp, tienda, id, "cancelar", "tienda"); err != nil {
		return Suscriptor{}, err
	}
	var out Suscriptor
	err := pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		var err error
		out, err = escanearSuscriptor(t.QueryRow(ctx, `SELECT `+columnasSuscriptor+` FROM suscripciones s WHERE s.id = $1`, id))
		return err
	})
	return out, err
}

// Zona: las fechas de negocio son de Argentina.
var Zona = time.FixedZone("ART", -3*60*60)
