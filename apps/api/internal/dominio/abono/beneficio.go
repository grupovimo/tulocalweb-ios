package abono

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/auditoria"
)

// Beneficio para quedarse (§57): un descuento sobre los próximos cobros del débito, no sobre las
// diferencias prorrateadas. Una vez por tienda.
type Beneficio struct {
	Pct    int `json:"pct"`
	Cobros int `json:"cobros"`
}

// Ofertas por motivo de baja (MOTIVOS de la web): «es caro» → 50 % por 3 cobros; «otro» → 50 % el
// próximo. Los demás motivos ofrecen hablar con el equipo o pausar.
var Ofertas = map[int]Beneficio{0: {Pct: 50, Cobros: 3}, 4: {Pct: 50, Cobros: 1}}

// ErrBeneficioUsado: la tienda ya aceptó uno (o está en el plan Gratis).
var ErrBeneficioUsado = errors.New("abono: el beneficio ya se usó")

// ConDescuento: lo que se cobra en el próximo débito con el beneficio vigente.
func ConDescuento(importe int64, b Beneficio) int64 {
	if b.Cobros <= 0 || b.Pct <= 0 {
		return importe
	}
	return (importe*int64(100-b.Pct) + 50) / 100
}

// LeerBeneficio: el vigente (Cobros 0 si no hay) y si ya se usó alguna vez.
func LeerBeneficio(ctx context.Context, t pgx.Tx, tiendaID string) (Beneficio, bool, error) {
	var b Beneficio
	var usado bool
	err := t.QueryRow(ctx, `SELECT beneficio_pct, beneficio_cobros, beneficio_en IS NOT NULL FROM tiendas WHERE id = $1`,
		tiendaID).Scan(&b.Pct, &b.Cobros, &usado)
	return b, usado, err
}

// Otorgar el beneficio: una vez por tienda y con un plan pago.
func Otorgar(ctx context.Context, t pgx.Tx, tiendaID string, b Beneficio) error {
	tag, err := t.Exec(ctx, `UPDATE tiendas SET beneficio_pct = $2, beneficio_cobros = $3, beneficio_en = now()
		WHERE id = $1 AND beneficio_en IS NULL AND plan <> 'gratis'`, tiendaID, b.Pct, b.Cobros)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrBeneficioUsado
	}
	return auditoria.Registrar(ctx, t, tiendaID, auditoria.Registro{Accion: "plan.beneficio", Entidad: "tienda", EntidadID: tiendaID,
		Texto: "se quedó con un descuento en los próximos cobros", Despues: b})
}

// Revocar el beneficio recién otorgado (Mercado Pago no tomó el monto nuevo).
func Revocar(ctx context.Context, t pgx.Tx, tiendaID string) error {
	_, err := t.Exec(ctx, `UPDATE tiendas SET beneficio_pct = 0, beneficio_cobros = 0, beneficio_en = NULL WHERE id = $1`, tiendaID)
	return err
}

// GastarCobro: un débito cobrado con el descuento. true cuando era el último (el débito vuelve al precio).
func GastarCobro(ctx context.Context, t pgx.Tx, tiendaID string) (bool, error) {
	var quedan int
	err := t.QueryRow(ctx, `UPDATE tiendas SET beneficio_cobros = beneficio_cobros - 1
		WHERE id = $1 AND beneficio_cobros > 0 RETURNING beneficio_cobros`, tiendaID).Scan(&quedan)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	return err == nil && quedan == 0, err
}

// Debito: el plan y el ciclo que cobra el próximo débito (el programado si es pago) y su precio con el
// beneficio vigente.
func Debito(ctx context.Context, t pgx.Tx, tiendaID string) (plan, ciclo string, importe int64, err error) {
	if err = t.QueryRow(ctx, `SELECT CASE WHEN coalesce(plan_programado, 'gratis') <> 'gratis' THEN plan_programado ELSE plan END,
			CASE WHEN coalesce(plan_programado, 'gratis') <> 'gratis' THEN coalesce(ciclo_programado, plan_ciclo) ELSE plan_ciclo END
		FROM tiendas WHERE id = $1`, tiendaID).Scan(&plan, &ciclo); err != nil {
		return
	}
	b, _, err := LeerBeneficio(ctx, t, tiendaID)
	if err != nil {
		return
	}
	precio, _ := Precio(plan, ciclo)
	return plan, ciclo, ConDescuento(precio, b), nil
}
