package pedidos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/auditoria"
)

// Un pedido de la tienda online con Mercado Pago que todavía no se pagó no es una venta, como en
// Shopify (docs/decisiones.md §63): se ve «Esperando pago», no avisa ni cuenta en las ventas hasta
// que el pago (o la seña) se acredita, y a las 24 h se cancela solo y devuelve el stock y el cupón.

// EsperaPago: el mismo criterio que app.esperando_pago en la base (00085). Las consultas de ventas,
// clientes y métricas dejan afuera lo que está «Sin pagar» con Mercado Pago en la tienda online.
func EsperaPago(canal, pago, estadoPago string) bool {
	switch estadoPago {
	case "Aprobado", "Acreditada", "Cobrado", "Seña":
		return false
	}
	return canal == "web" && pago == "mp"
}

// devolver vuelve a sumar el stock de los ítems y libera el uso del cupón (si tenía), en la
// transacción del pedido que se cancela.
func devolver(ctx context.Context, t pgx.Tx, tiendaID string, items []byte, cupon *string) error {
	var ls []Item
	if err := json.Unmarshal(items, &ls); err != nil {
		return err
	}
	for _, it := range ls {
		if it.ProductoID == nil {
			continue
		}
		var err error
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
	if cupon == nil {
		return nil
	}
	_, err := t.Exec(ctx, `UPDATE cupones SET usos = greatest(usos - 1, 0) WHERE tienda_id = $1 AND codigo = $2`, tiendaID, *cupon)
	return err
}

// CancelarSinPago cancela un pedido de la tienda online que espera el pago de Mercado Pago hace más
// de 24 h: el stock vuelve y el cupón se libera. No avisa a nadie (no fue una venta); queda en la
// Actividad. Si mientras tanto se pagó, no hace nada y devuelve false.
func CancelarSinPago(ctx context.Context, pool tx, tiendaID string, numero int) (bool, error) {
	var hecho bool
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var items []byte
		var cupon *string
		err := t.QueryRow(ctx, `UPDATE pedidos SET estado = 'Cancelado',
				estados_en = estados_en || jsonb_build_object('Cancelado', now())
			WHERE tienda_id = $1 AND numero = $2 AND estado = 'Sin pagar' AND canal = 'web' AND pago = 'mp'
				AND sena IS NULL AND codigo_publico IS NOT NULL AND estado_pago = 'Pendiente' AND creado_en < now() - interval '24 hours'
			RETURNING items, cupon`, tiendaID, numero).Scan(&items, &cupon)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		if err := devolver(ctx, t, tiendaID, items, cupon); err != nil {
			return err
		}
		hecho = true
		_, err = auditoria.Anotar(ctx, t, tiendaID, auditoria.Actor{Tipo: "sistema"}, auditoria.Registro{
			Accion: "pedido.vencido", Entidad: "pedido", EntidadID: fmt.Sprint(numero),
			Texto: fmt.Sprintf("canceló el pedido #%d: no se pagó con Mercado Pago en 24 horas (el stock volvió)", numero),
			Antes: map[string]string{"estado": "Sin pagar"}, Despues: map[string]string{"estado": "Cancelado"}})
		return err
	})
	return hecho, err
}

type txs interface {
	tx
	EnTx(context.Context, func(pgx.Tx) error) error
}

// VencerSinPago cancela, de a tandas, los pedidos que esperan el pago de Mercado Pago hace más de
// 24 h. Lo corre el worker cada minuto; el UPDATE condicionado evita que dos workers cancelen dos veces.
func VencerSinPago(ctx context.Context, pool txs) (int, error) {
	type par struct {
		tienda string
		numero int
	}
	var pares []par
	if err := pool.EnTx(ctx, func(t pgx.Tx) error {
		filas, err := t.Query(ctx, `SELECT tienda_id::text, numero FROM app.mp_sin_pagar_vencidos(200)`)
		if err != nil {
			return err
		}
		defer filas.Close()
		for filas.Next() {
			var p par
			if err := filas.Scan(&p.tienda, &p.numero); err != nil {
				return err
			}
			pares = append(pares, p)
		}
		return filas.Err()
	}); err != nil {
		return 0, err
	}
	n := 0
	for _, p := range pares {
		hecho, err := CancelarSinPago(ctx, pool, p.tienda, p.numero)
		if err != nil {
			return n, err
		}
		if hecho {
			n++
		}
	}
	return n, nil
}
