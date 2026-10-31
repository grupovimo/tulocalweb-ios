-- +goose Up
-- Lanzamientos con reserva, ofertas con fecha de fin y encargos con seña (docs/decisiones.md §52).
--
-- Producto: lanzamiento (antes de esa hora se ve «Próximamente» y no se compra), la seña de la
-- reserva antes del lanzamiento (% del precio; NULL = sin reserva), la seña del encargo (% del
-- precio; NULL = se compra entero), la oferta (precio de oferta y hasta cuándo; el precio de lista
-- sigue en `precio`) y cuándo se encolaron los avisos del lanzamiento.
ALTER TABLE productos
	ADD COLUMN lanzamiento timestamptz,
	ADD COLUMN reserva_sena smallint CHECK (reserva_sena BETWEEN 1 AND 100),
	ADD COLUMN encargo_sena smallint CHECK (encargo_sena BETWEEN 1 AND 100),
	ADD COLUMN oferta_precio bigint CHECK (oferta_precio > 0),
	ADD COLUMN oferta_hasta timestamptz,
	ADD COLUMN lanzamiento_avisado timestamptz,
	ADD CONSTRAINT productos_reserva_con_lanzamiento CHECK (reserva_sena IS NULL OR lanzamiento IS NOT NULL),
	ADD CONSTRAINT productos_oferta_hasta_con_precio CHECK (oferta_hasta IS NULL OR oferta_precio IS NOT NULL);
GRANT UPDATE (lanzamiento, reserva_sena, encargo_sena, oferta_precio, oferta_hasta, lanzamiento_avisado)
	ON productos TO mostrador_app;

-- «Avisame cuando salga»: un email por producto (sin cuenta). avisado_en = ya le salió el mail.
CREATE TABLE avisos_lanzamiento (
	tienda_id uuid NOT NULL,
	producto_id uuid NOT NULL,
	email text NOT NULL CHECK (length(email) BETWEEN 3 AND 200 AND email = lower(email)),
	creado_en timestamptz NOT NULL DEFAULT now(),
	avisado_en timestamptz,
	PRIMARY KEY (tienda_id, producto_id, email),
	FOREIGN KEY (tienda_id, producto_id) REFERENCES productos (tienda_id, id) ON DELETE CASCADE
);
CREATE INDEX avisos_lanzamiento_pendientes ON avisos_lanzamiento (tienda_id, producto_id, creado_en)
	WHERE avisado_en IS NULL;
SELECT app.aplicar_rls('avisos_lanzamiento');
GRANT SELECT, INSERT, DELETE ON avisos_lanzamiento TO mostrador_app;
GRANT UPDATE (avisado_en) ON avisos_lanzamiento TO mostrador_app;

-- Pedido con seña (reserva de un lanzamiento o encargo): sena es lo que se cobró por Mercado Pago;
-- el resto (total − seña) se paga al retirar o recibir. sena_vence: hasta cuándo espera el pago de
-- la seña (después se cancela y el stock vuelve). saldado_en: la tienda marcó que cobró el resto.
ALTER TABLE pedidos
	ADD COLUMN sena bigint CHECK (sena > 0),
	ADD COLUMN sena_tipo text CHECK (sena_tipo IN ('reserva', 'encargo')),
	ADD COLUMN sena_pagada_en timestamptz,
	ADD COLUMN sena_vence timestamptz,
	ADD COLUMN saldado_en timestamptz,
	ADD CONSTRAINT pedidos_sena_con_tipo CHECK ((sena IS NULL) = (sena_tipo IS NULL));
GRANT UPDATE (sena_pagada_en, sena_vence, saldado_en) ON pedidos TO mostrador_app;
CREATE INDEX pedidos_senas_por_vencer ON pedidos (sena_vence) WHERE sena_vence IS NOT NULL;

-- Para el worker, que pasa por todas las tiendas (sólo ids, como app.tiendas_con_arca_por_vencer):
-- los lanzamientos que ya llegaron y no avisaron, y las señas que no se pagaron a tiempo.
CREATE POLICY lanzamientos_por_avisar ON productos FOR SELECT TO mostrador_migrador
	USING (current_setting('app.lanzamientos_worker', true) = 'si');
CREATE POLICY senas_vencidas ON pedidos FOR SELECT TO mostrador_migrador
	USING (current_setting('app.lanzamientos_worker', true) = 'si');
-- +goose StatementBegin
CREATE FUNCTION app.lanzamientos_por_avisar(p_limite int) RETURNS TABLE (tienda_id uuid, producto_id uuid)
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
BEGIN
	PERFORM set_config('app.lanzamientos_worker', 'si', true);
	RETURN QUERY SELECT p.tienda_id, p.id FROM productos p
		WHERE p.lanzamiento <= now() AND p.lanzamiento > now() - interval '7 days' AND p.lanzamiento_avisado IS NULL
		ORDER BY p.lanzamiento LIMIT p_limite;
	PERFORM set_config('app.lanzamientos_worker', '', true);
END
$$;
CREATE FUNCTION app.senas_vencidas(p_limite int) RETURNS TABLE (tienda_id uuid, numero int)
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
BEGIN
	PERFORM set_config('app.lanzamientos_worker', 'si', true);
	RETURN QUERY SELECT p.tienda_id, p.numero FROM pedidos p
		WHERE p.sena_vence < now() AND p.sena_pagada_en IS NULL AND p.estado = 'Sin pagar'
		ORDER BY p.sena_vence LIMIT p_limite;
	PERFORM set_config('app.lanzamientos_worker', '', true);
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.lanzamientos_por_avisar(int) FROM PUBLIC;
REVOKE ALL ON FUNCTION app.senas_vencidas(int) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.lanzamientos_por_avisar(int) TO mostrador_app;
GRANT EXECUTE ON FUNCTION app.senas_vencidas(int) TO mostrador_app;

-- +goose Down
DELETE FROM outbox WHERE tipo = 'lanzamiento.avisos';
DROP FUNCTION app.senas_vencidas(int);
DROP FUNCTION app.lanzamientos_por_avisar(int);
DROP POLICY senas_vencidas ON pedidos;
DROP POLICY lanzamientos_por_avisar ON productos;
DROP INDEX pedidos_senas_por_vencer;
REVOKE UPDATE (sena_pagada_en, sena_vence, saldado_en) ON pedidos FROM mostrador_app;
ALTER TABLE pedidos DROP CONSTRAINT pedidos_sena_con_tipo, DROP COLUMN saldado_en, DROP COLUMN sena_vence,
	DROP COLUMN sena_pagada_en, DROP COLUMN sena_tipo, DROP COLUMN sena;
DROP TABLE avisos_lanzamiento;
REVOKE UPDATE (lanzamiento, reserva_sena, encargo_sena, oferta_precio, oferta_hasta, lanzamiento_avisado)
	ON productos FROM mostrador_app;
ALTER TABLE productos DROP CONSTRAINT productos_oferta_hasta_con_precio, DROP CONSTRAINT productos_reserva_con_lanzamiento,
	DROP COLUMN lanzamiento_avisado, DROP COLUMN oferta_hasta, DROP COLUMN oferta_precio, DROP COLUMN encargo_sena,
	DROP COLUMN reserva_sena, DROP COLUMN lanzamiento;
