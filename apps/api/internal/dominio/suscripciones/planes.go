package suscripciones

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/auditoria"
)

const columnasPlan = `p.id::text, p.nombre, p.descripcion, p.foto, p.precio, p.frecuencia, p.productos, p.incluye, p.envio, p.retiro,
	p.cupo, p.activo, p.orden,
	(SELECT count(*) FROM suscripciones s WHERE s.tienda_id = p.tienda_id AND s.plan_id = p.id AND s.estado IN ('activa', 'pausada'))::int`

func escanearPlan(f pgx.Row) (Plan, error) {
	var p Plan
	err := f.Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.Foto, &p.Precio, &p.Frecuencia, &p.Productos, &p.Incluye, &p.Envio, &p.Retiro,
		&p.Cupo, &p.Activo, &p.Orden, &p.Suscriptores)
	if p.Productos == nil {
		p.Productos = []ProductoPlan{}
	}
	return p, err
}

func leerPlanes(ctx context.Context, t pgx.Tx, soloActivos bool) ([]Plan, error) {
	q := `SELECT ` + columnasPlan + ` FROM suscripcion_planes p`
	if soloActivos {
		q += ` WHERE p.activo`
	}
	filas, err := t.Query(ctx, q+` ORDER BY p.orden, p.creado_en`)
	if err != nil {
		return nil, err
	}
	out, err := pgx.CollectRows(filas, func(f pgx.CollectableRow) (Plan, error) { return escanearPlan(f) })
	if out == nil {
		out = []Plan{}
	}
	return out, err
}

const columnasSuscriptor = `s.id::text, s.plan_id::text, s.plan_nombre, s.precio, s.frecuencia, s.nombre, s.email, s.telefono, s.entrega,
	s.direccion, s.estado, s.proximo_cobro, s.cobros, s.ultimo_cobro, s.creada_en, s.cancelada_por`

func escanearSuscriptor(f pgx.Row) (Suscriptor, error) {
	var s Suscriptor
	err := f.Scan(&s.ID, &s.PlanID, &s.Plan, &s.Precio, &s.Frecuencia, &s.Nombre, &s.Email, &s.Telefono, &s.Entrega,
		&s.Direccion, &s.Estado, &s.ProximoCobro, &s.Cobros, &s.UltimoCobro, &s.CreadaEn, &s.CanceladaPor)
	return s, err
}

// PanelDe: los planes (activos y pausados) y los suscriptores, de los más nuevos a los más viejos.
// Las altas que nunca se autorizaron en Mercado Pago no se listan pasado un día: nadie se suscribió.
func PanelDe(ctx context.Context, pool tx, tienda string) (Panel, error) {
	var out Panel
	err := pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		var err error
		if out.MercadoPago, err = conMercadoPago(ctx, t, tienda); err != nil {
			return err
		}
		if out.Planes, err = leerPlanes(ctx, t, false); err != nil {
			return err
		}
		filas, err := t.Query(ctx, `SELECT `+columnasSuscriptor+` FROM suscripciones s
			WHERE s.estado <> 'pendiente' OR s.creada_en > now() - interval '1 day'
			ORDER BY s.creada_en DESC LIMIT 500`)
		if err != nil {
			return err
		}
		out.Suscriptores, err = pgx.CollectRows(filas, func(f pgx.CollectableRow) (Suscriptor, error) { return escanearSuscriptor(f) })
		return err
	})
	if out.Suscriptores == nil {
		out.Suscriptores = []Suscriptor{}
	}
	return out, err
}

func (p *Plan) validar() error {
	p.Nombre, p.Descripcion, p.Incluye = strings.TrimSpace(p.Nombre), strings.TrimSpace(p.Descripcion), strings.TrimSpace(p.Incluye)
	if p.Foto != nil && strings.TrimSpace(*p.Foto) == "" {
		p.Foto = nil
	}
	if p.Frecuencia == "" {
		p.Frecuencia = "mensual"
	}
	if largo(p.Nombre) < 2 || largo(p.Nombre) > 60 || largo(p.Descripcion) > 400 || largo(p.Incluye) > 400 ||
		(p.Foto != nil && len(*p.Foto) > 2048) || p.Precio < 100 || p.Precio > 100_000_000 || !Frecuencias[p.Frecuencia] ||
		(!p.Envio && !p.Retiro) || (p.Cupo != nil && (*p.Cupo < 1 || *p.Cupo > 100_000)) || len(p.Productos) > 30 {
		return ErrInvalido
	}
	if p.Productos == nil {
		p.Productos = []ProductoPlan{}
	}
	// Que incluya algo: productos del catálogo, texto, o los dos.
	if len(p.Productos) == 0 && p.Incluye == "" {
		return ErrInvalido
	}
	for _, x := range p.Productos {
		if !idValido.MatchString(x.ProductoID) || (x.VarianteID != "" && !idValido.MatchString(x.VarianteID)) || x.Cantidad < 1 || x.Cantidad > 99 {
			return ErrInvalido
		}
	}
	return nil
}

// productosDeLaTienda: cada producto del plan es de la tienda (RLS) y, con opciones, trae una
// variante suya.
func productosDeLaTienda(ctx context.Context, t pgx.Tx, ps []ProductoPlan) error {
	for _, x := range ps {
		var conOpciones bool
		err := t.QueryRow(ctx, `SELECT opciones <> '[]'::jsonb FROM productos WHERE id::text = $1`, x.ProductoID).Scan(&conOpciones)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvalido
		}
		if err != nil {
			return err
		}
		if conOpciones != (x.VarianteID != "") {
			return ErrInvalido
		}
		if x.VarianteID != "" {
			var ok bool
			if err := t.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM variantes WHERE id::text = $1 AND producto_id::text = $2)`,
				x.VarianteID, x.ProductoID).Scan(&ok); err != nil {
				return err
			}
			if !ok {
				return ErrInvalido
			}
		}
	}
	return nil
}

// GuardarPlan crea (id vacío) o edita un plan. Editar no cambia lo que pagan quienes ya se
// suscribieron (la suscripción copió el precio que autorizó en Mercado Pago).
func GuardarPlan(ctx context.Context, pool tx, tienda, id string, p Plan) (Plan, error) {
	if err := p.validar(); err != nil {
		return Plan{}, err
	}
	if id != "" && !idValido.MatchString(id) {
		return Plan{}, ErrNoExiste
	}
	var out Plan
	err := pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		if err := productosDeLaTienda(ctx, t, p.Productos); err != nil {
			return err
		}
		productos, _ := json.Marshal(p.Productos)
		var nuevo string
		if id == "" {
			err := t.QueryRow(ctx, `INSERT INTO suscripcion_planes (tienda_id, nombre, descripcion, foto, precio, frecuencia, productos,
					incluye, envio, retiro, cupo, activo, orden)
				SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, coalesce(max(orden) + 1, 0) FROM suscripcion_planes WHERE tienda_id = $1
				RETURNING id::text`, tienda, p.Nombre, p.Descripcion, p.Foto, p.Precio, p.Frecuencia, productos, p.Incluye,
				p.Envio, p.Retiro, p.Cupo, p.Activo).Scan(&nuevo)
			if err != nil {
				return err
			}
			id = nuevo
		} else {
			tag, err := t.Exec(ctx, `UPDATE suscripcion_planes SET nombre = $2, descripcion = $3, foto = $4, precio = $5, frecuencia = $6,
					productos = $7, incluye = $8, envio = $9, retiro = $10, cupo = $11, activo = $12
				WHERE id = $1`, id, p.Nombre, p.Descripcion, p.Foto, p.Precio, p.Frecuencia, productos, p.Incluye,
				p.Envio, p.Retiro, p.Cupo, p.Activo)
			if err != nil {
				return err
			}
			if tag.RowsAffected() == 0 {
				return ErrNoExiste
			}
		}
		var err error
		if out, err = escanearPlan(t.QueryRow(ctx, `SELECT `+columnasPlan+` FROM suscripcion_planes p WHERE p.id = $1`, id)); err != nil {
			return err
		}
		texto := "editó el plan de suscripción «" + p.Nombre + "»"
		accion := "cobros.suscripcion_plan_editar"
		if nuevo != "" {
			texto, accion = "creó el plan de suscripción «"+p.Nombre+"» de "+auditoria.Pesos(p.Precio), "cobros.suscripcion_plan_crear"
		}
		return auditoria.Registrar(ctx, t, tienda, auditoria.Registro{Accion: accion, Entidad: "suscripcion", EntidadID: id, Texto: texto})
	})
	return out, err
}

// BorrarPlan lo saca si nunca tuvo suscriptores; si tuvo, ErrEnUso (se pausa: las suscripciones
// siguen apuntando a él).
func BorrarPlan(ctx context.Context, pool tx, tienda, id string) error {
	if !idValido.MatchString(id) {
		return ErrNoExiste
	}
	return pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		var usado bool
		if err := t.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM suscripciones WHERE plan_id = $1)`, id).Scan(&usado); err != nil {
			return err
		}
		if usado {
			return ErrEnUso
		}
		var nombre string
		err := t.QueryRow(ctx, `DELETE FROM suscripcion_planes WHERE id = $1 RETURNING nombre`, id).Scan(&nombre)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoExiste
		}
		if err != nil {
			return err
		}
		return auditoria.Registrar(ctx, t, tienda, auditoria.Registro{Accion: "cobros.suscripcion_plan_borrar", Entidad: "suscripcion",
			EntidadID: id, Texto: "borró el plan de suscripción «" + nombre + "»"})
	})
}

// vigentes: lo que ocupa cupo (activas, pausadas y altas de la última hora que todavía se están
// autorizando en Mercado Pago).
const vigentes = `SELECT count(*) FROM suscripciones WHERE plan_id = $1
	AND (estado IN ('activa', 'pausada') OR (estado = 'pendiente' AND creada_en > now() - interval '1 hour'))`

// PlanesDe la tienda pública (TiendaPublica.suscripciones): los activos, sólo si la tienda puede
// cobrarlos (Mercado Pago conectado). cobra=false (sin la integración en el servidor): ninguno.
func PlanesDe(ctx context.Context, t pgx.Tx, tienda string, cobra bool) ([]PlanPublico, error) {
	out := []PlanPublico{}
	if !cobra {
		return out, nil
	}
	mp, err := conMercadoPago(ctx, t, tienda)
	if err != nil || !mp {
		return out, err
	}
	planes, err := leerPlanes(ctx, t, true)
	if err != nil {
		return nil, err
	}
	for _, p := range planes {
		pp := PlanPublico{ID: p.ID, Nombre: p.Nombre, Descripcion: p.Descripcion, Foto: p.Foto, Precio: p.Precio,
			Frecuencia: p.Frecuencia, Incluye: p.Incluye, Envio: p.Envio, Retiro: p.Retiro, Productos: []Incluido{}}
		if p.Cupo != nil {
			var n int
			if err := t.QueryRow(ctx, vigentes, p.ID).Scan(&n); err != nil {
				return nil, err
			}
			pp.Agotado = n >= *p.Cupo
		}
		for _, x := range p.Productos {
			inc, ok, err := incluido(ctx, t, x)
			if err != nil {
				return nil, err
			}
			if ok {
				pp.Productos = append(pp.Productos, inc)
			}
		}
		out = append(out, pp)
	}
	return out, nil
}

// incluido: el producto del plan con su nombre y foto actuales (false si ya no existe).
func incluido(ctx context.Context, t pgx.Tx, x ProductoPlan) (Incluido, bool, error) {
	inc := Incluido{ID: x.ProductoID, Cantidad: x.Cantidad}
	var opciones []listas.Opcion
	err := t.QueryRow(ctx, `SELECT nombre, foto, opciones FROM productos WHERE id::text = $1`, x.ProductoID).Scan(&inc.Nombre, &inc.Foto, &opciones)
	if errors.Is(err, pgx.ErrNoRows) {
		return inc, false, nil
	}
	if err != nil {
		return inc, false, err
	}
	if x.VarianteID != "" {
		var valores map[string]string
		var foto *string
		err := t.QueryRow(ctx, `SELECT valores, foto FROM variantes WHERE id::text = $1 AND producto_id::text = $2`, x.VarianteID, x.ProductoID).
			Scan(&valores, &foto)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return inc, false, err
		}
		if err == nil {
			texto := listas.TextoVariante(opciones, valores)
			inc.Variante = &texto
			if foto != nil {
				inc.Foto = foto
			}
		}
	}
	return inc, true, nil
}
