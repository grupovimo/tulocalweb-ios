package abono

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/auditoria"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

type tx interface {
	EnTienda(ctx context.Context, tiendaID string, fn func(pgx.Tx) error) error
	EnTx(ctx context.Context, fn func(pgx.Tx) error) error
}

// Leer el estado del abono de una tienda (dentro de su transacción).
func Leer(ctx context.Context, t pgx.Tx, tiendaID string) (Estado, error) {
	var e Estado
	var creada time.Time
	var estadoMP *string
	var progPlan, progCiclo *string
	var progDesde *time.Time
	err := t.QueryRow(ctx, `SELECT plan, plan_ciclo, creada_en, suscripcion_estado, abono_proximo,
			plan_programado, ciclo_programado, programado_desde
		FROM tiendas WHERE id = $1`, tiendaID).
		Scan(&e.Plan, &e.Ciclo, &creada, &estadoMP, &e.Proximo, &progPlan, &progCiclo, &progDesde)
	if err != nil {
		return e, err
	}
	e.FinPrueba = creada.Add(DiasDePrueba * 24 * time.Hour)
	e.Activo = estadoMP != nil && *estadoMP == "authorized"
	if e.Activo && e.Proximo == nil {
		// Débitos de antes de §56, sin el próximo cobro guardado: el que caería contando desde la prueba
		// (el próximo aviso de Mercado Pago lo corrige).
		p := e.FinPrueba
		for !p.After(time.Now()) {
			p = Siguiente(p, e.Ciclo)
		}
		e.Proximo = &p
	}
	if progPlan != nil && progDesde != nil {
		ciclo := e.Ciclo
		if progCiclo != nil {
			ciclo = *progCiclo
		}
		e.Pendiente = &Cambio{Plan: *progPlan, Ciclo: ciclo, Desde: *progDesde}
	}
	return e, nil
}

// Programar el cambio que rige desde una fecha (nil: borra el programado).
func Programar(ctx context.Context, t pgx.Tx, tiendaID string, c *Cambio) error {
	if c == nil {
		_, err := t.Exec(ctx, `UPDATE tiendas SET plan_programado = NULL, ciclo_programado = NULL, programado_desde = NULL
			WHERE id = $1`, tiendaID)
		return err
	}
	_, err := t.Exec(ctx, `UPDATE tiendas SET plan_programado = $2, ciclo_programado = $3, programado_desde = $4 WHERE id = $1`,
		tiendaID, c.Plan, c.Ciclo, c.Desde)
	return err
}

// Aplicar el plan ya: el plan, el ciclo, la auditoría y la campanita (§45) si es un plan pago.
func Aplicar(ctx context.Context, t pgx.Tx, tiendaID, plan, ciclo, texto string) error {
	var antes string
	if err := t.QueryRow(ctx, `SELECT plan FROM tiendas WHERE id = $1 FOR UPDATE`, tiendaID).Scan(&antes); err != nil {
		return err
	}
	if _, err := t.Exec(ctx, `UPDATE tiendas SET plan = $2, plan_ciclo = $3, cancelada_en = NULL WHERE id = $1`,
		tiendaID, plan, ciclo); err != nil {
		return err
	}
	if plan == "gratis" {
		if _, err := t.Exec(ctx, `UPDATE tiendas SET abono_proximo = NULL WHERE id = $1`, tiendaID); err != nil {
			return err
		}
	}
	if err := auditoria.Registrar(ctx, t, tiendaID, auditoria.Registro{Accion: "plan.cambiar", Entidad: "tienda", EntidadID: tiendaID,
		Texto: texto, Antes: map[string]string{"plan": antes}, Despues: map[string]string{"plan": plan, "ciclo": ciclo}}); err != nil {
		return err
	}
	if plan == "gratis" || plan == antes {
		return nil
	}
	return outbox.Publicar(ctx, t, tiendaID, "plan.activado", map[string]string{"plan": plan})
}

// AplicarProgramados: los cambios que llegaron a su fecha (el worker, cada pocos minutos). Sólo ids y
// UPDATE condicionados: dos workers no aplican dos veces.
func AplicarProgramados(ctx context.Context, pool tx) (int, error) {
	var ids []string
	if err := pool.EnTx(ctx, func(t pgx.Tx) error {
		filas, err := t.Query(ctx, `SELECT id::text FROM app.abonos_programados(200) AS id`)
		if err != nil {
			return err
		}
		defer filas.Close()
		for filas.Next() {
			var id string
			if err := filas.Scan(&id); err != nil {
				return err
			}
			ids = append(ids, id)
		}
		return filas.Err()
	}); err != nil {
		return 0, err
	}
	n := 0
	for _, id := range ids {
		err := pool.EnTienda(ctx, id, func(t pgx.Tx) error {
			var plan, ciclo string
			err := t.QueryRow(ctx, `SELECT plan_programado, coalesce(ciclo_programado, plan_ciclo) FROM tiendas
				WHERE id = $1 AND programado_desde <= now() FOR UPDATE`, id).Scan(&plan, &ciclo)
			if err == pgx.ErrNoRows {
				return nil
			}
			if err != nil {
				return err
			}
			if err := Programar(ctx, t, id, nil); err != nil {
				return err
			}
			n++
			return Aplicar(ctx, t, id, plan, ciclo, "empezó el plan "+Nombres[plan]+" (cambio programado)")
		})
		if err != nil {
			return n, err
		}
	}
	return n, nil
}
