package pedidos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/auditoria"
)

// Sena es la seña de un pedido encargado o reservado (docs/decisiones.md §52): cuánto se pagó por
// Mercado Pago, cuánto falta (lo que se cobra al retirar o entregar) y si la tienda ya lo saldó.
type Sena struct {
	Tipo    string `json:"tipo"`
	Monto   int64  `json:"monto"`
	Falta   int64  `json:"falta"`
	Pagada  bool   `json:"pagada"`
	Saldado bool   `json:"saldado"`
	// Vence: hasta cuándo se espera el pago de la seña (después se cancela solo); null si ya se pagó.
	Vence *time.Time `json:"vence"`
}

// ErrSinSena: el pedido no tiene una seña pagada por saldar (o no tiene seña).
var ErrSinSena = errors.New("pedidos: sin seña pagada")

func senaDe(ctx context.Context, t pgx.Tx, tiendaID string, numero int, total int64) (*Sena, error) {
	var monto *int64
	var tipo *string
	var s Sena
	err := t.QueryRow(ctx, `SELECT sena, sena_tipo, sena_pagada_en IS NOT NULL, saldado_en IS NOT NULL, sena_vence
		FROM pedidos WHERE tienda_id = $1 AND numero = $2`, tiendaID, numero).Scan(&monto, &tipo, &s.Pagada, &s.Saldado, &s.Vence)
	if err != nil || monto == nil || tipo == nil {
		return nil, err
	}
	s.Tipo, s.Monto, s.Falta = *tipo, *monto, max(0, total-*monto)
	if s.Saldado {
		s.Falta = 0
	}
	return &s, nil
}

// Saldar marca que la tienda cobró el resto de un pedido con la seña pagada (o lo desmarca, para
// el «Deshacer»). Saldado, el pago queda «Cobrado»: la venta se completa (mail a quien compró y
// factura, como cualquier pedido cobrado).
func Saldar(ctx context.Context, pool tx, tiendaID string, numero int, saldado bool) error {
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var falta int64
		var antes bool
		err := t.QueryRow(ctx, `SELECT total - sena, saldado_en IS NOT NULL FROM pedidos
			WHERE tienda_id = $1 AND numero = $2 AND sena IS NOT NULL AND sena_pagada_en IS NOT NULL AND estado <> 'Cancelado'
			FOR UPDATE`, tiendaID, numero).Scan(&falta, &antes)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrSinSena
		}
		if err != nil || antes == saldado {
			return err
		}
		if _, err := t.Exec(ctx, `UPDATE pedidos SET saldado_en = CASE WHEN $3 THEN now() END,
				estado_pago = CASE WHEN $3 THEN 'Cobrado' ELSE 'Seña' END
			WHERE tienda_id = $1 AND numero = $2`, tiendaID, numero, saldado); err != nil {
			return err
		}
		r := auditoria.Registro{Accion: "pedido.saldado", Entidad: "pedido", EntidadID: fmt.Sprint(numero),
			Texto:   fmt.Sprintf("marcó como saldado el pedido #%d (cobró %s del resto)", numero, auditoria.Pesos(falta)),
			Despues: map[string]bool{"saldado": saldado}}
		if !saldado {
			r.Texto = fmt.Sprintf("volvió a dejar pendiente el saldo del pedido #%d", numero)
		}
		return auditoria.Registrar(ctx, t, tiendaID, r)
	})
}

// CancelarSinSena cancela un pedido con seña que todavía espera el pago y devuelve su stock. Lo usan
// el worker (la seña no se pagó a tiempo) y la vidriera (Mercado Pago no respondió al armar el
// pago). Si ya se pagó o ya no espera, no hace nada y devuelve false.
func CancelarSinSena(ctx context.Context, pool tx, tiendaID string, numero int, soloVencido bool) (bool, error) {
	var hecho bool
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var items []byte
		err := t.QueryRow(ctx, `UPDATE pedidos SET estado = 'Cancelado', sena_vence = NULL,
				estados_en = estados_en || jsonb_build_object('Cancelado', now())
			WHERE tienda_id = $1 AND numero = $2 AND sena IS NOT NULL AND sena_pagada_en IS NULL AND estado = 'Sin pagar'
				AND (NOT $3 OR sena_vence < now())
			RETURNING items`, tiendaID, numero, soloVencido).Scan(&items)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		var ls []Item
		if err := json.Unmarshal(items, &ls); err != nil {
			return err
		}
		for _, it := range ls {
			if it.ProductoID == nil {
				continue
			}
			if it.VarianteID != nil {
				_, err = t.Exec(ctx, `UPDATE variantes SET stock = stock + $4
					WHERE tienda_id = $1 AND producto_id::text = $2 AND id::text = $3`, tiendaID, *it.ProductoID, *it.VarianteID, it.Cantidad)
			} else {
				_, err = t.Exec(ctx, `UPDATE productos SET stock = stock + $3
					WHERE tienda_id = $1 AND id::text = $2 AND opciones = '[]'::jsonb`, tiendaID, *it.ProductoID, it.Cantidad)
			}
			if err != nil {
				return err
			}
		}
		hecho = true
		return nil
	})
	return hecho, err
}
