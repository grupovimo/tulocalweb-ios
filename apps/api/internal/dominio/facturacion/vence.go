package facturacion

import (
	"context"
	"math"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

// TipoVence es el evento del outbox del aviso de vencimiento del certificado (§48).
const TipoVence = "arca.vence"

// Vence son los datos del evento: los días que faltan (0 o menos: ya venció) y la fecha.
type Vence struct {
	Dias  int       `json:"dias"`
	Vence time.Time `json:"vence"`
}

// umbral: 30, 7 o 0 según lo que falta (-1: todavía no hay que avisar).
func umbral(dias int) int {
	switch {
	case dias <= 0:
		return 0
	case dias <= 7:
		return 7
	case dias <= 30:
		return 30
	}
	return -1
}

// AvisarVencimientos encola un aviso por tienda cuando su certificado cruza los 30 días, los 7 y al
// vencer. Cada umbral se avisa una vez (aviso_vence); un certificado nuevo arranca de cero.
func (s *Servicio) AvisarVencimientos(ctx context.Context, ahora time.Time) error {
	var tiendas []string
	err := s.Pool.EnTx(ctx, func(t pgx.Tx) error {
		filas, err := t.Query(ctx, `SELECT id::text FROM app.tiendas_con_arca_por_vencer(100) id`)
		if err != nil {
			return err
		}
		tiendas, err = pgx.CollectRows(filas, pgx.RowTo[string])
		return err
	})
	if err != nil {
		return err
	}
	for _, tienda := range tiendas {
		if err := s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
			var vence time.Time
			var avisado *int
			if err := t.QueryRow(ctx, `SELECT certificado_vence, aviso_vence FROM arca_config WHERE tienda_id = $1
					AND certificado IS NOT NULL AND certificado_vence IS NOT NULL FOR UPDATE`, tienda).Scan(&vence, &avisado); err != nil {
				return err
			}
			dias := int(math.Ceil(vence.Sub(ahora).Hours() / 24))
			u := umbral(dias)
			if u < 0 || (avisado != nil && *avisado <= u) {
				return nil
			}
			if _, err := t.Exec(ctx, `UPDATE arca_config SET aviso_vence = $2 WHERE tienda_id = $1`, tienda, u); err != nil {
				return err
			}
			return outbox.Publicar(ctx, t, tienda, TipoVence, Vence{Dias: dias, Vence: vence})
		}); err != nil && ctx.Err() == nil {
			return err
		}
	}
	return nil
}
