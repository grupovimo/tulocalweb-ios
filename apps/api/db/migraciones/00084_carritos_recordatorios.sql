-- +goose Up
-- Carrito abandonado más rápido, como Shopify (docs/decisiones.md §62): hasta tres recordatorios por
-- mail con su demora (por defecto 1 h, 24 h y 72 h), cupón opcional de un uso en el segundo o el
-- tercero, baja con un clic y lo recuperado (pedidos y monto). Reemplaza al mail único del tercer día.
--
-- La configuración: activo (el interruptor de siempre, carritos_automaticos->>'email', que se
-- conserva) y los tres pasos {activo, horas, cupon: null | {descuento, valor, dias}}.
ALTER TABLE tiendas ADD COLUMN carritos_recordatorios jsonb NOT NULL DEFAULT
	'{"activo": false, "pasos": [{"activo": true, "horas": 1}, {"activo": true, "horas": 24}, {"activo": true, "horas": 72}]}'
	CHECK (jsonb_typeof(carritos_recordatorios) = 'object' AND pg_column_size(carritos_recordatorios) <= 2048);
UPDATE tiendas SET carritos_recordatorios = jsonb_set(carritos_recordatorios, '{activo}',
	to_jsonb(coalesce((carritos_automaticos->>'email')::boolean, false)));
GRANT UPDATE (carritos_recordatorios) ON tiendas TO mostrador_app;

-- El carrito: qué paso automático ya salió (0 a 3), si se dio de baja, el cupón que se le dio y el
-- pedido con el que se recuperó (para las métricas: lo que se cobró de verdad, no lo que tenía).
ALTER TABLE carritos
	ADD COLUMN paso smallint NOT NULL DEFAULT 0 CHECK (paso BETWEEN 0 AND 3),
	ADD COLUMN baja_en timestamptz,
	ADD COLUMN cupon text CHECK (cupon IS NULL OR length(cupon) <= 40),
	ADD COLUMN pedido_numero int,
	ADD COLUMN pedido_total bigint CHECK (pedido_total IS NULL OR pedido_total >= 0);
-- Los que ya recibieron el mail del tercer día no vuelven a empezar.
UPDATE carritos SET paso = 3 WHERE recordado_en IS NOT NULL;
GRANT UPDATE (paso, baja_en, cupon, pedido_numero, pedido_total) ON carritos TO mostrador_app;
CREATE INDEX carritos_por_recordar ON carritos (actualizado) WHERE recuperado_en IS NULL AND baja_en IS NULL AND paso < 3;

-- El cupón único de un recordatorio: origen «carrito:<id>», como los de las ventanas.
ALTER TABLE cupones DROP CONSTRAINT cupones_origen_check;
ALTER TABLE cupones ADD CONSTRAINT cupones_origen_check CHECK (origen IS NULL OR origen ~ '^ventana:[a-z0-9]{6,16}$'
	OR origen ~ '^carrito:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$');

-- Para el worker, de todas las tiendas (sólo ids): los carritos abiertos de la última semana, sin
-- baja, de tiendas con los recordatorios prendidos y con algún paso que ya venció. Cuál manda lo
-- decide Go con la configuración.
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION app.carritos_para_recordar(p_limite int) RETURNS TABLE (tienda_id uuid, id uuid)
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
BEGIN
	PERFORM set_config('app.carritos_recordar', 'si', true);
	RETURN QUERY SELECT c.tienda_id, c.id FROM carritos c JOIN tiendas t ON t.id = c.tienda_id
		WHERE c.recuperado_en IS NULL AND c.baja_en IS NULL AND c.paso < 3
			AND c.actualizado > now() - interval '8 days'
			AND coalesce((t.carritos_recordatorios->>'activo')::boolean, false)
			AND EXISTS (SELECT 1 FROM jsonb_array_elements(t.carritos_recordatorios->'pasos') WITH ORDINALITY AS p (x, i)
				WHERE p.i > c.paso AND coalesce((p.x->>'activo')::boolean, false)
					AND c.actualizado <= now() - make_interval(hours => coalesce((p.x->>'horas')::int, 72)))
		ORDER BY c.actualizado LIMIT p_limite;
	PERFORM set_config('app.carritos_recordar', '', true);
END
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION app.carritos_para_recordar(p_limite int) RETURNS TABLE (tienda_id uuid, id uuid)
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
BEGIN
	PERFORM set_config('app.carritos_recordar', 'si', true);
	RETURN QUERY SELECT c.tienda_id, c.id FROM carritos c JOIN tiendas t ON t.id = c.tienda_id
		WHERE c.recuperado_en IS NULL AND c.recordado_en IS NULL
			AND c.actualizado BETWEEN now() - interval '4 days' AND now() - interval '3 days'
			AND coalesce((t.carritos_automaticos->>'email')::boolean, false)
		LIMIT p_limite;
	PERFORM set_config('app.carritos_recordar', '', true);
END
$$;
-- +goose StatementEnd
DELETE FROM cupones WHERE origen LIKE 'carrito:%';
ALTER TABLE cupones DROP CONSTRAINT cupones_origen_check;
ALTER TABLE cupones ADD CONSTRAINT cupones_origen_check CHECK (origen IS NULL OR origen ~ '^ventana:[a-z0-9]{6,16}$');
DROP INDEX carritos_por_recordar;
ALTER TABLE carritos DROP COLUMN pedido_total, DROP COLUMN pedido_numero, DROP COLUMN cupon, DROP COLUMN baja_en, DROP COLUMN paso;
ALTER TABLE tiendas DROP COLUMN carritos_recordatorios;
