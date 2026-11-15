-- +goose Up
-- Un pedido de la tienda online con Mercado Pago que todavía no se pagó no es una venta, como en
-- Shopify (docs/decisiones.md §63). Nada avisa «venta nueva» hasta que el pago se acredita (o se
-- paga la seña): ni Telegram, ni el mail de confirmación a quien compra, ni el «Pedido nuevo» del
-- panel en tiempo real, ni lo vendido del cupón. Con transferencia sigue como antes: avisa al crearse.
-- Si no se paga en 24 h, el worker lo cancela, devuelve el stock y libera el cupón (sin avisar nada).
--
-- «Confirmado» es pagado o con la seña paga (§52): ahí empieza a ser una venta.

-- +goose StatementBegin
CREATE FUNCTION app.esperando_pago(p_canal text, p_pago text, p_estado_pago text) RETURNS boolean
LANGUAGE sql IMMUTABLE AS $$
	SELECT p_canal = 'web' AND p_pago = 'mp' AND p_estado_pago NOT IN ('Aprobado', 'Acreditada', 'Cobrado', 'Seña')
$$;
-- +goose StatementEnd

-- El pedido de Mercado Pago de la tienda online pasó a confirmado en este UPDATE (y no está cancelado:
-- un pago que llega tarde lo devuelve la tienda, no es una venta).
-- +goose StatementBegin
CREATE FUNCTION app.mp_confirmado(p_canal text, p_pago text, p_estado text, p_antes text, p_ahora text) RETURNS boolean
LANGUAGE sql IMMUTABLE AS $$
	SELECT p_estado <> 'Cancelado' AND app.esperando_pago(p_canal, p_pago, p_antes)
		AND NOT app.esperando_pago(p_canal, p_pago, p_ahora)
$$;
-- +goose StatementEnd

-- Telegram: al crearse, salvo que espere el pago; si esperaba, cuando se confirma. Una sola vez
-- (pedidos.avisos->'telegram', lo marca el worker).
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION app.avisar_venta_telegram() RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app, pg_temp AS $$
BEGIN
	IF TG_OP = 'INSERT' AND app.esperando_pago(NEW.canal, NEW.pago, NEW.estado_pago) THEN
		RETURN NEW;
	END IF;
	IF TG_OP = 'UPDATE' AND NOT app.mp_confirmado(NEW.canal, NEW.pago, NEW.estado, OLD.estado_pago, NEW.estado_pago) THEN
		RETURN NEW;
	END IF;
	IF EXISTS (SELECT 1 FROM telegram_vinculos WHERE tienda_id = NEW.tienda_id) THEN
		INSERT INTO outbox (tienda_id, tipo, datos)
		VALUES (NEW.tienda_id, 'telegram.venta', jsonb_build_object('numero', NEW.numero));
	END IF;
	RETURN NEW;
END
$$;
-- +goose StatementEnd
DROP TRIGGER avisar_venta_telegram ON pedidos;
CREATE TRIGGER avisar_venta_telegram AFTER INSERT OR UPDATE OF estado_pago ON pedidos
	FOR EACH ROW EXECUTE FUNCTION app.avisar_venta_telegram();

-- «Nueva venta #N» (mail y push) a la dueña: igual que en 00058, y además al pagarse la seña (antes lo
-- encolaba Go, sin el push) y nunca por un pedido cancelado.
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION app.avisar_venta_tienda() RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app, pg_temp AS $$
BEGIN
	IF NEW.canal <> 'web' THEN
		RETURN NEW;
	END IF;
	IF (TG_OP = 'INSERT' AND NOT app.esperando_pago(NEW.canal, NEW.pago, NEW.estado_pago))
		OR (TG_OP = 'UPDATE' AND app.mp_confirmado(NEW.canal, NEW.pago, NEW.estado, OLD.estado_pago, NEW.estado_pago)) THEN
		INSERT INTO outbox (tienda_id, tipo, datos)
		VALUES (NEW.tienda_id, 'tienda.venta', jsonb_build_object('numero', NEW.numero)),
			(NEW.tienda_id, 'push.venta', jsonb_build_object('numero', NEW.numero));
	END IF;
	RETURN NEW;
END
$$;
-- +goose StatementEnd

-- Mails a quien compra. Esperando el pago no sale nada: la confirmación sale cuando se acredita, una
-- sola (con el pago adentro), en vez de «Gracias por tu compra» al crearse y «Recibimos tu pago»
-- después. La seña tiene su propio mail (Go, «Recibimos tu seña»), y el resto pago, el de «pago».
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION app.avisar_comprador() RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app, pg_temp AS $$
DECLARE
	v_aviso text;
	v_avisos text[] := '{}';
	pagado constant text[] := ARRAY['Aprobado', 'Acreditada', 'Cobrado'];
BEGIN
	IF coalesce(NEW.comprador->>'email', '') = '' OR NEW.canal = 'pos' THEN
		RETURN NEW;
	END IF;
	IF TG_OP = 'INSERT' THEN
		IF NOT app.esperando_pago(NEW.canal, NEW.pago, NEW.estado_pago) THEN
			v_avisos := ARRAY['confirmacion'];
		END IF;
	ELSE
		IF NEW.estado_pago = ANY (pagado) AND NOT (OLD.estado_pago = ANY (pagado)) THEN
			IF app.mp_confirmado(NEW.canal, NEW.pago, NEW.estado, OLD.estado_pago, NEW.estado_pago) AND NEW.sena IS NULL AND NOT (NEW.avisos ? 'confirmacion') THEN
				v_avisos := v_avisos || 'confirmacion'::text;
			ELSIF NEW.estado <> 'Cancelado' THEN
				v_avisos := v_avisos || 'pago'::text;
			END IF;
		END IF;
		IF NEW.estado = 'En camino' AND OLD.estado <> 'En camino' AND NEW.envio <> 'retiro' THEN
			v_avisos := v_avisos || 'en_camino'::text;
		END IF;
		IF NEW.estado = 'Listo para despachar' AND OLD.estado <> 'Listo para despachar' AND NEW.envio = 'retiro' THEN
			v_avisos := v_avisos || 'retiro'::text;
		END IF;
	END IF;
	FOREACH v_aviso IN ARRAY v_avisos LOOP
		INSERT INTO outbox (tienda_id, tipo, datos)
		VALUES (NEW.tienda_id, 'comprador.aviso', jsonb_build_object('numero', NEW.numero, 'aviso', v_aviso));
	END LOOP;
	RETURN NEW;
END
$$;
-- +goose StatementEnd

-- El cupón: al crearse el pedido que espera el pago se reserva el uso (Go suma `usos`, así un cupón
-- único no lo usan dos a la vez), pero lo vendido y lo descontado se suman recién al confirmarse.
-- +goose StatementBegin
CREATE FUNCTION app.sumar_cupon_confirmado() RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app, pg_temp AS $$
BEGIN
	IF NEW.cupon IS NULL OR NOT app.mp_confirmado(NEW.canal, NEW.pago, NEW.estado, OLD.estado_pago, NEW.estado_pago) THEN
		RETURN NEW;
	END IF;
	UPDATE cupones SET vendio = vendio + NEW.total,
		desconto = desconto + coalesce((SELECT sum((d->>'monto')::bigint) FROM jsonb_array_elements(NEW.descuentos) d
			WHERE d->>'concepto' = 'Cupón ' || NEW.cupon), 0)
	WHERE tienda_id = NEW.tienda_id AND codigo = NEW.cupon;
	RETURN NEW;
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.sumar_cupon_confirmado() FROM PUBLIC;
CREATE TRIGGER sumar_cupon_confirmado AFTER UPDATE OF estado_pago ON pedidos
	FOR EACH ROW EXECUTE FUNCTION app.sumar_cupon_confirmado();

-- Tiempo real (§44): el pedido que espera el pago entra como «pedido.cambio» (la lista se refresca,
-- el panel no lo anuncia ni suena) y se anuncia como «pedido.nuevo» cuando se confirma.
-- +goose StatementBegin
CREATE FUNCTION app.tr_tiempo_real_pedido_nuevo() RETURNS trigger
LANGUAGE plpgsql SET search_path = public, app, pg_temp AS $$
DECLARE
	r record;
BEGIN
	FOR r IN
		WITH por AS (SELECT tienda_id, count(*) AS cuantas FROM nuevas GROUP BY tienda_id)
		SELECT n.tienda_id, n.numero::text AS id, app.esperando_pago(n.canal, n.pago, n.estado_pago) AS espera, false AS lote
			FROM nuevas n JOIN por p USING (tienda_id) WHERE p.cuantas <= 20
		UNION ALL
		SELECT tienda_id, NULL, false, true FROM por WHERE cuantas > 20
	LOOP
		IF r.lote THEN
			PERFORM app.avisar_tiempo_real(r.tienda_id, 'pedido.nuevo', NULL, NULL, '{"lote": true}');
		ELSE
			PERFORM app.avisar_tiempo_real(r.tienda_id, CASE WHEN r.espera THEN 'pedido.cambio' ELSE 'pedido.nuevo' END, r.id);
		END IF;
	END LOOP;
	RETURN NULL;
END
$$;
-- +goose StatementEnd
DROP TRIGGER tiempo_real_pedido_ins ON pedidos;
CREATE TRIGGER tiempo_real_pedido_ins AFTER INSERT ON pedidos REFERENCING NEW TABLE AS nuevas
	FOR EACH STATEMENT EXECUTE FUNCTION app.tr_tiempo_real_pedido_nuevo();

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION app.tr_tiempo_real_pago() RETURNS trigger
LANGUAGE plpgsql SET search_path = public, app, pg_temp AS $$
DECLARE
	pagado constant text[] := ARRAY['Aprobado', 'Acreditada', 'Cobrado'];
	r record;
BEGIN
	FOR r IN SELECT n.tienda_id, n.numero,
			n.estado_pago = ANY (pagado) AND NOT (o.estado_pago = ANY (pagado)) AS acreditado,
			app.mp_confirmado(n.canal, n.pago, n.estado, o.estado_pago, n.estado_pago) AS confirmado
		FROM nuevas n JOIN viejas o USING (tienda_id, numero)
	LOOP
		IF r.acreditado THEN
			PERFORM app.avisar_tiempo_real(r.tienda_id, 'pago.acreditado', r.numero::text);
		END IF;
		IF r.confirmado THEN
			PERFORM app.avisar_tiempo_real(r.tienda_id, 'pedido.nuevo', r.numero::text);
		END IF;
	END LOOP;
	RETURN NULL;
END
$$;
-- +goose StatementEnd

-- Vencimiento: las compras de la tienda online (con código público) que esperan el pago hace más de
-- 24 h (sin seña: esas vencen con la suya, §52). El
-- worker ve todas las tiendas sólo dentro de esta función (como app.senas_vencidas) y cancela cada
-- uno con la tienda fijada.
CREATE INDEX pedidos_mp_esperando ON pedidos (creado_en)
	WHERE estado = 'Sin pagar' AND canal = 'web' AND pago = 'mp' AND sena IS NULL AND codigo_publico IS NOT NULL;
CREATE POLICY mp_sin_pagar ON pedidos FOR SELECT TO mostrador_migrador
	USING (current_setting('app.mp_sin_pagar_worker', true) = 'si');
-- +goose StatementBegin
CREATE FUNCTION app.mp_sin_pagar_vencidos(p_limite int) RETURNS TABLE (tienda_id uuid, numero int)
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
REVOKE ALL ON FUNCTION app.mp_sin_pagar_vencidos(int) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.mp_sin_pagar_vencidos(int) TO mostrador_app;

-- +goose Down
DROP FUNCTION app.mp_sin_pagar_vencidos(int);
DROP POLICY mp_sin_pagar ON pedidos;
DROP INDEX pedidos_mp_esperando;

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION app.tr_tiempo_real_pago() RETURNS trigger
LANGUAGE plpgsql SET search_path = public, app, pg_temp AS $$
DECLARE
	pagado constant text[] := ARRAY['Aprobado', 'Acreditada', 'Cobrado'];
	r record;
BEGIN
	FOR r IN SELECT n.tienda_id, n.numero FROM nuevas n JOIN viejas o USING (tienda_id, numero)
		WHERE n.estado_pago = ANY (pagado) AND NOT (o.estado_pago = ANY (pagado))
	LOOP
		PERFORM app.avisar_tiempo_real(r.tienda_id, 'pago.acreditado', r.numero::text);
	END LOOP;
	RETURN NULL;
END
$$;
-- +goose StatementEnd
DROP TRIGGER tiempo_real_pedido_ins ON pedidos;
DROP FUNCTION app.tr_tiempo_real_pedido_nuevo();
CREATE TRIGGER tiempo_real_pedido_ins AFTER INSERT ON pedidos REFERENCING NEW TABLE AS nuevas
	FOR EACH STATEMENT EXECUTE FUNCTION app.tr_tiempo_real('pedido.nuevo', 'numero', 'tienda_id', '', '', '');

DROP TRIGGER sumar_cupon_confirmado ON pedidos;
DROP FUNCTION app.sumar_cupon_confirmado();

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION app.avisar_comprador() RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app, pg_temp AS $$
DECLARE
	v_aviso text;
	v_avisos text[] := '{}';
	pagado constant text[] := ARRAY['Aprobado', 'Acreditada', 'Cobrado'];
BEGIN
	IF coalesce(NEW.comprador->>'email', '') = '' OR NEW.canal = 'pos' THEN
		RETURN NEW;
	END IF;
	IF TG_OP = 'INSERT' THEN
		v_avisos := ARRAY['confirmacion'];
	ELSE
		IF NEW.estado_pago = ANY (pagado) AND NOT (OLD.estado_pago = ANY (pagado)) THEN
			v_avisos := v_avisos || 'pago'::text;
		END IF;
		IF NEW.estado = 'En camino' AND OLD.estado <> 'En camino' AND NEW.envio <> 'retiro' THEN
			v_avisos := v_avisos || 'en_camino'::text;
		END IF;
		IF NEW.estado = 'Listo para despachar' AND OLD.estado <> 'Listo para despachar' AND NEW.envio = 'retiro' THEN
			v_avisos := v_avisos || 'retiro'::text;
		END IF;
	END IF;
	FOREACH v_aviso IN ARRAY v_avisos LOOP
		INSERT INTO outbox (tienda_id, tipo, datos)
		VALUES (NEW.tienda_id, 'comprador.aviso', jsonb_build_object('numero', NEW.numero, 'aviso', v_aviso));
	END LOOP;
	RETURN NEW;
END
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION app.avisar_venta_tienda() RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app, pg_temp AS $$
DECLARE
	pagado constant text[] := ARRAY['Aprobado', 'Acreditada', 'Cobrado'];
BEGIN
	IF NEW.canal <> 'web' THEN
		RETURN NEW;
	END IF;
	IF (TG_OP = 'INSERT' AND (NEW.pago <> 'mp' OR NEW.estado_pago = ANY (pagado)))
		OR (TG_OP = 'UPDATE' AND NEW.pago = 'mp' AND NEW.estado_pago = ANY (pagado)
			AND NOT (OLD.estado_pago = ANY (pagado))) THEN
		INSERT INTO outbox (tienda_id, tipo, datos)
		VALUES (NEW.tienda_id, 'tienda.venta', jsonb_build_object('numero', NEW.numero)),
			(NEW.tienda_id, 'push.venta', jsonb_build_object('numero', NEW.numero));
	END IF;
	RETURN NEW;
END
$$;
-- +goose StatementEnd

DROP TRIGGER avisar_venta_telegram ON pedidos;
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION app.avisar_venta_telegram() RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app, pg_temp AS $$
BEGIN
	IF EXISTS (SELECT 1 FROM telegram_vinculos WHERE tienda_id = NEW.tienda_id) THEN
		INSERT INTO outbox (tienda_id, tipo, datos)
		VALUES (NEW.tienda_id, 'telegram.venta', jsonb_build_object('numero', NEW.numero));
	END IF;
	RETURN NEW;
END
$$;
-- +goose StatementEnd
CREATE TRIGGER avisar_venta_telegram AFTER INSERT ON pedidos
	FOR EACH ROW EXECUTE FUNCTION app.avisar_venta_telegram();

DROP FUNCTION app.mp_confirmado(text, text, text, text, text);
DROP FUNCTION app.esperando_pago(text, text, text);
