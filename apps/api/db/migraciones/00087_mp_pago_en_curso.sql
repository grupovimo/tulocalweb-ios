-- +goose Up
-- Un pedido con un pago de Mercado Pago en curso (efectivo en Rapipago/Pago Fácil, o en revisión)
-- no vence a las 24 h: el cupón de pago vale varios días y, si se cancelaba, el pago se acreditaba
-- después sobre un pedido cancelado, con el stock ya devuelto y sin aviso de venta. Se espera hasta
-- 7 días a que Mercado Pago lo resuelva (si vence el cupón, avisa «cancelled» y el pedido vence).
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION app.mp_sin_pagar_vencidos(p_limite int) RETURNS TABLE (tienda_id uuid, numero int)
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
BEGIN
	PERFORM set_config('app.mp_sin_pagar_worker', 'si', true);
	RETURN QUERY SELECT p.tienda_id, p.numero FROM pedidos p
		WHERE p.estado = 'Sin pagar' AND p.canal = 'web' AND p.pago = 'mp' AND p.sena IS NULL AND p.codigo_publico IS NOT NULL
			AND p.estado_pago = 'Pendiente' AND p.creado_en < now() - interval '24 hours'
			AND NOT EXISTS (SELECT 1 FROM mp_cobros c WHERE c.tienda_id = p.tienda_id AND c.tipo = 'pedido' AND c.objetivo = p.numero::text
				AND c.estado IN ('pending', 'in_process', 'authorized') AND c.actualizado_en > now() - interval '7 days')
		ORDER BY p.creado_en LIMIT p_limite;
	PERFORM set_config('app.mp_sin_pagar_worker', '', true);
END
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION app.mp_sin_pagar_vencidos(p_limite int) RETURNS TABLE (tienda_id uuid, numero int)
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
BEGIN
	PERFORM set_config('app.mp_sin_pagar_worker', 'si', true);
	RETURN QUERY SELECT p.tienda_id, p.numero FROM pedidos p
		WHERE p.estado = 'Sin pagar' AND p.canal = 'web' AND p.pago = 'mp' AND p.sena IS NULL AND p.codigo_publico IS NOT NULL
			AND p.estado_pago = 'Pendiente' AND p.creado_en < now() - interval '24 hours'
		ORDER BY p.creado_en LIMIT p_limite;
	PERFORM set_config('app.mp_sin_pagar_worker', '', true);
END
$$;
-- +goose StatementEnd
