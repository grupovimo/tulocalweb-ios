package suscripciones

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

// Lo que usa el paquete mercadopago para conciliar los avisos. Todo corre en la transacción que ya
// tiene la tienda fijada (RLS) y es idempotente: Mercado Pago repite los avisos.

// EstadoDeMP traduce el estado de un preapproval.
func EstadoDeMP(status string) (string, bool) {
	e, ok := map[string]string{"pending": Pendiente, "authorized": Activa, "paused": Pausada, "cancelled": Cancelada}[status]
	return e, ok
}

// AplicarEstado lleva la suscripción al estado que dice Mercado Pago (status del preapproval) y
// anota el próximo cobro. Una cancelada no vuelve. Pasar a activa por primera vez es el alta
// (suscripcion.alta: mail a quien se suscribió y aviso a la tienda); a cancelada, la baja.
func AplicarEstado(ctx context.Context, t pgx.Tx, tienda, id, status string, proximo *time.Time) error {
	nuevo, ok := EstadoDeMP(status)
	if !ok {
		return nil
	}
	var actual string
	var activada *time.Time
	err := t.QueryRow(ctx, `SELECT estado, activada_en FROM suscripciones WHERE id = $1 FOR UPDATE`, id).Scan(&actual, &activada)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if actual == Cancelada || (nuevo == Pendiente && actual != Pendiente) {
		return nil
	}
	if nuevo != Activa {
		proximo = nil
	}
	if _, err := t.Exec(ctx, `UPDATE suscripciones SET estado = $2, proximo_cobro = $3, actualizada_en = now(),
			activada_en = CASE WHEN $2 = 'activa' THEN coalesce(activada_en, now()) ELSE activada_en END,
			cancelada_en = CASE WHEN $2 = 'cancelada' THEN now() ELSE cancelada_en END,
			cancelada_por = CASE WHEN $2 = 'cancelada' THEN coalesce(cancelada_por, 'mercadopago') ELSE cancelada_por END
		WHERE id = $1`, id, nuevo, proximo); err != nil {
		return err
	}
	switch {
	case nuevo == actual:
		return nil
	case nuevo == Activa && activada == nil:
		return outbox.Publicar(ctx, t, tienda, "suscripcion.alta", map[string]string{"id": id})
	case nuevo == Cancelada && actual != Pendiente:
		return outbox.Publicar(ctx, t, tienda, "suscripcion.baja", map[string]string{"id": id, "por": "mercadopago"})
	}
	return nil
}

// RegistrarCobro: un cobro aprobado de la suscripción crea su pedido (canal web, pagado con Mercado
// Pago, marcado con la suscripción) con el plan y los productos que incluye. Un mismo pago crea un
// solo pedido aunque el aviso llegue varias veces (la clave «pago:<id>»). Si lo cobrado no llega al
// precio acordado, no se arma pedido. Devuelve el número (0 si no hizo nada).
func RegistrarCobro(ctx context.Context, t pgx.Tx, tienda, id string, pago, cobrado int64) (int, error) {
	var s struct {
		plan, planNombre, email, nombre, telefono, entrega, direccion, cp, estado string
		precio                                                                    int64
		activada                                                                  *time.Time
		productos                                                                 []ProductoPlan
	}
	err := t.QueryRow(ctx, `SELECT s.plan_id::text, s.plan_nombre, s.email, s.nombre, s.telefono, s.entrega, s.direccion, s.cp, s.estado,
			s.precio, s.activada_en, p.productos
		FROM suscripciones s JOIN suscripcion_planes p ON p.id = s.plan_id WHERE s.id = $1 FOR UPDATE OF s`, id).
		Scan(&s.plan, &s.planNombre, &s.email, &s.nombre, &s.telefono, &s.entrega, &s.direccion, &s.cp, &s.estado,
			&s.precio, &s.activada, &s.productos)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	if err != nil || cobrado < s.precio {
		return 0, err
	}
	tag, err := t.Exec(ctx, `INSERT INTO suscripcion_cobros (tienda_id, clave, suscripcion_id, estado, monto, mp_pago)
		VALUES ($1, $2, $3, 'aprobado', $4, $5) ON CONFLICT DO NOTHING`, tienda, fmt.Sprintf("pago:%d", pago), id, cobrado, pago)
	if err != nil || tag.RowsAffected() == 0 {
		return 0, err
	}
	numero, err := crearPedido(ctx, t, tienda, id, s.planNombre, s.precio, s.email, s.nombre, s.telefono, s.entrega, s.direccion, s.cp, s.productos)
	if err != nil {
		return 0, err
	}
	if _, err := t.Exec(ctx, `UPDATE suscripcion_cobros SET pedido = $3 WHERE tienda_id = $1 AND clave = $2`,
		tienda, fmt.Sprintf("pago:%d", pago), numero); err != nil {
		return 0, err
	}
	// Mercado Pago puede avisar el cobro antes que la autorización: el cobro también es el alta.
	estado := s.estado
	if estado == Pendiente {
		estado = Activa
	}
	if _, err := t.Exec(ctx, `UPDATE suscripciones SET cobros = cobros + 1, ultimo_cobro = now(), estado = $2, actualizada_en = now(),
			activada_en = coalesce(activada_en, now()) WHERE id = $1`, id, estado); err != nil {
		return 0, err
	}
	if s.activada == nil && s.estado == Pendiente {
		if err := outbox.Publicar(ctx, t, tienda, "suscripcion.alta", map[string]string{"id": id}); err != nil {
			return 0, err
		}
	}
	return numero, nil
}

// RegistrarRechazo: Mercado Pago no pudo cobrar (clave del cobro autorizado). Avisa una sola vez a
// la tienda (suscripcion.rechazo); Mercado Pago reintenta solo.
func RegistrarRechazo(ctx context.Context, t pgx.Tx, tienda, id, clave string, monto int64) error {
	tag, err := t.Exec(ctx, `INSERT INTO suscripcion_cobros (tienda_id, clave, suscripcion_id, estado, monto)
		SELECT $1, $2, s.id, 'rechazado', $4 FROM suscripciones s WHERE s.id = $3 ON CONFLICT DO NOTHING`,
		tienda, "rechazo:"+clave, id, monto)
	if err != nil || tag.RowsAffected() == 0 {
		return err
	}
	return outbox.Publicar(ctx, t, tienda, "suscripcion.rechazo", map[string]string{"id": id})
}

// PorPreapproval: la suscripción de la tienda con ese débito de Mercado Pago ("" si no hay).
func PorPreapproval(ctx context.Context, t pgx.Tx, preapproval string) (string, error) {
	var id string
	err := t.QueryRow(ctx, `SELECT id::text FROM suscripciones WHERE mp_preapproval = $1`, preapproval).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return id, err
}

// crearPedido arma el pedido del cobro: el plan por el precio cobrado y, abajo, lo que hay que
// preparar (los productos del plan, en $ 0). El stock se descuenta sin frenar: el cobro ya se hizo
// y el pedido tiene que existir aunque falte algo (la tienda lo ve en Para preparar).
func crearPedido(ctx context.Context, t pgx.Tx, tienda, id, plan string, precio int64, email, nombre, telefono, entrega, direccion, cp string,
	productos []ProductoPlan) (int, error) {
	items := []pedidos.Item{{Nombre: "Suscripción: " + plan, Cantidad: 1, Precio: precio}}
	for _, x := range productos {
		it := pedidos.Item{Cantidad: x.Cantidad}
		pid := x.ProductoID
		it.ProductoID = &pid
		var opciones []listas.Opcion
		err := t.QueryRow(ctx, `UPDATE productos SET stock = greatest(stock - $2, 0) WHERE id::text = $1
			RETURNING nombre, foto, opciones`, x.ProductoID, x.Cantidad).Scan(&it.Nombre, &it.Foto, &opciones)
		if errors.Is(err, pgx.ErrNoRows) {
			continue // lo borraron del catálogo: el plan sigue con el resto
		}
		if err != nil {
			return 0, err
		}
		if x.VarianteID != "" {
			var valores map[string]string
			var foto *string
			err := t.QueryRow(ctx, `UPDATE variantes SET stock = greatest(stock - $3, 0) WHERE id::text = $1 AND producto_id::text = $2
				RETURNING valores, foto`, x.VarianteID, x.ProductoID, x.Cantidad).Scan(&valores, &foto)
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return 0, err
			}
			if err == nil {
				vid, texto := x.VarianteID, listas.TextoVariante(opciones, valores)
				it.VarianteID, it.Variante = &vid, &texto
				if foto != nil {
					it.Foto = foto
				}
			}
		}
		items = append(items, it)
	}
	var tel *string
	if telefono != "" {
		tel = &telefono
	}
	comprador, _ := json.Marshal(pedidos.Comprador{Nombre: nombre, Email: &email, Telefono: tel})
	js, _ := json.Marshal(items)
	envio, nota := "retiro", "Suscripción a «"+plan+"»"
	var ent *pedidos.Entrega
	if entrega == "envio" {
		envio, ent = "moto", &pedidos.Entrega{Direccion: direccion, CP: cp}
		var correo bool
		if err := t.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM ca_conexiones WHERE tienda_id = $1)`, tienda).Scan(&correo); err != nil {
			return 0, err
		}
		if correo {
			envio = "correo"
		}
		nota += " · Entrega: " + direccion
		if cp != "" {
			nota += " (CP " + cp + ")"
		}
	}
	codigo := nuevoCodigo()
	if _, err := t.Exec(ctx, `SELECT 1 FROM tiendas WHERE id = $1 FOR UPDATE`, tienda); err != nil {
		return 0, err
	}
	var numero int
	err := t.QueryRow(ctx, `
		INSERT INTO pedidos (tienda_id, numero, comprador, canal, pago, estado_pago, envio, estado, facturado,
			items, subtotal, costo_envio, total, nota, estados_en, ciudad_envio, codigo_publico, entrega, descuentos, suscripcion_id)
		SELECT $1, coalesce(max(numero), 1000) + 1, $2, 'web', 'mp', 'Aprobado', $3, 'Para preparar', false,
			$4, $5::bigint, 0, $5::bigint, $6, jsonb_build_object('Para preparar', now()), $7, $8, $9, '[]'::jsonb, $10
		FROM pedidos WHERE tienda_id = $1
		RETURNING numero`, tienda, comprador, envio, js, precio, nota, cp, codigo, ent, id).Scan(&numero)
	return numero, err
}
