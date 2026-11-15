-- +goose Up
-- «Avisame cuando vuelva» (docs/decisiones.md §60): en la ficha de un producto o variante sin stock,
-- quien compra deja su email (y, si quiere, su WhatsApp) con su consentimiento. Cuando el stock pasa
-- de 0 a más de 0, por el camino que sea (panel, apps, importación, cancelación de un pedido), sale un
-- mail con la marca de la tienda y el link al producto, una sola vez por pedido de aviso.
CREATE TABLE avisos_stock (
	id uuid NOT NULL DEFAULT gen_random_uuid(),
	tienda_id uuid NOT NULL,
	producto_id uuid NOT NULL,
	-- NULL: el producto (sin opciones, o cualquiera de sus variantes).
	variante_id uuid,
	email text NOT NULL CHECK (length(email) BETWEEN 3 AND 200 AND email = lower(email)),
	whatsapp text CHECK (whatsapp IS NULL OR whatsapp ~ '^\+?[0-9 ()-]{6,24}$'),
	consentimiento_en timestamptz NOT NULL,
	creado_en timestamptz NOT NULL DEFAULT now(),
	avisado_en timestamptz,
	PRIMARY KEY (tienda_id, id),
	FOREIGN KEY (tienda_id, producto_id) REFERENCES productos (tienda_id, id) ON DELETE CASCADE
);
-- Uno pendiente por persona y producto (o variante): pedirlo de nuevo no suma.
CREATE UNIQUE INDEX avisos_stock_uno ON avisos_stock
	(tienda_id, producto_id, coalesce(variante_id, '00000000-0000-0000-0000-000000000000'::uuid), email)
	WHERE avisado_en IS NULL;
CREATE INDEX avisos_stock_pendientes ON avisos_stock (tienda_id, producto_id, creado_en) WHERE avisado_en IS NULL;
SELECT app.aplicar_rls('avisos_stock');
GRANT SELECT, INSERT, DELETE ON avisos_stock TO mostrador_app;
GRANT UPDATE (whatsapp, avisado_en) ON avisos_stock TO mostrador_app;

-- Volvió el stock: se encola el aviso en la misma transacción que lo repone (si la reposición no se
-- hace, tampoco el aviso). Sólo si hay alguien esperando, así una reposición común no toca el outbox.
-- +goose StatementBegin
CREATE FUNCTION app.stock_repuesto() RETURNS trigger
LANGUAGE plpgsql SET search_path = public, app AS $$
DECLARE
	v uuid;
	p uuid;
BEGIN
	IF TG_TABLE_NAME = 'variantes' THEN
		v := NEW.id; p := NEW.producto_id;
		IF NOT (NEW.activa AND NEW.stock > 0 AND (OLD.stock = 0 OR NOT OLD.activa)) THEN
			RETURN NULL;
		END IF;
		IF NOT EXISTS (SELECT 1 FROM avisos_stock WHERE tienda_id = NEW.tienda_id AND producto_id = p
			AND variante_id = v AND avisado_en IS NULL) THEN
			RETURN NULL;
		END IF;
	ELSE
		p := NEW.id;
		IF NOT (OLD.stock = 0 AND NEW.stock > 0) THEN
			RETURN NULL;
		END IF;
		IF NOT EXISTS (SELECT 1 FROM avisos_stock WHERE tienda_id = NEW.tienda_id AND producto_id = p
			AND variante_id IS NULL AND avisado_en IS NULL) THEN
			RETURN NULL;
		END IF;
	END IF;
	INSERT INTO outbox (tienda_id, tipo, datos)
	VALUES (NEW.tienda_id, 'stock.repuesto', jsonb_build_object('producto', p));
	RETURN NULL;
END
$$;
-- +goose StatementEnd
CREATE TRIGGER productos_stock_repuesto AFTER UPDATE OF stock ON productos
	FOR EACH ROW EXECUTE FUNCTION app.stock_repuesto();
CREATE TRIGGER variantes_stock_repuesto AFTER UPDATE OF stock, activa ON variantes
	FOR EACH ROW EXECUTE FUNCTION app.stock_repuesto();

-- La supresión del comprador (00079) también borra sus avisos de stock.
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION app.suprimir_comprador(p_clave text) RETURNS int
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app, pg_temp AS $$
DECLARE
	t uuid := app.tienda_actual();
	n int := 0;
	k int;
	v_email text;
	v_pedidos int[];
BEGIN
	IF t IS NULL OR coalesce(p_clave, '') = '' THEN
		RETURN 0;
	END IF;
	-- El email de quien se suprime (de la clave o de su ficha sin email no hay).
	v_email := CASE WHEN position('@' IN p_clave) > 0 THEN lower(p_clave) END;
	SELECT array_agg(p.numero) INTO v_pedidos FROM pedidos p
		WHERE p.tienda_id = t AND app.clave_comprador(p.comprador->>'email', p.comprador->>'nombre') = p_clave;
	IF EXISTS (SELECT 1 FROM pedidos p WHERE p.tienda_id = t AND p.numero = ANY (v_pedidos)
			AND p.estado NOT IN ('Entregado', 'Cancelado'))
		OR EXISTS (SELECT 1 FROM turnos u WHERE u.tienda_id = t AND u.estado IN ('reservado', 'confirmado') AND u.fin > now()
			AND app.clave_comprador(u.cliente_email, u.cliente_nombre) = p_clave)
		OR EXISTS (SELECT 1 FROM suscripciones s WHERE s.tienda_id = t AND s.estado IN ('pendiente', 'activa', 'pausada')
			AND app.clave_comprador(s.email, s.nombre) = p_clave)
		OR EXISTS (SELECT 1 FROM devoluciones d WHERE d.tienda_id = t AND d.estado <> 'Aprobada'
			AND (d.pedido_numero = ANY (v_pedidos) OR app.clave_comprador(d.email, d.nombre) = p_clave)) THEN
		RAISE EXCEPTION 'comprador_en_curso' USING ERRCODE = 'P0001';
	END IF;

	-- Pedidos: quedan como venta de «Consumidor final» (la lista de Clientes ya no los cuenta como
	-- cliente), sin la dirección ni la nota de la entrega.
	UPDATE pedidos p SET comprador = '{"nombre": "Consumidor final"}', cliente_id = NULL, nota = NULL,
			entrega = CASE WHEN p.entrega IS NULL THEN NULL ELSE p.entrega - 'direccion' - 'cp' END
		WHERE p.tienda_id = t AND p.numero = ANY (v_pedidos);
	GET DIAGNOSTICS k = ROW_COUNT; n := n + k;
	UPDATE resenas r SET autor = 'Anónimo' WHERE r.tienda_id = t AND r.pedido_numero = ANY (v_pedidos);
	GET DIAGNOSTICS k = ROW_COUNT; n := n + k;
	UPDATE devoluciones d SET nombre = 'Consumidor final', email = NULL
		WHERE d.tienda_id = t AND (d.pedido_numero = ANY (v_pedidos) OR app.clave_comprador(d.email, d.nombre) = p_clave);
	GET DIAGNOSTICS k = ROW_COUNT; n := n + k;
	UPDATE historial_pedidos h SET cliente_email = NULL, cliente_nombre = ''
		WHERE h.tienda_id = t AND app.clave_comprador(h.cliente_email, h.cliente_nombre) = p_clave;
	GET DIAGNOSTICS k = ROW_COUNT; n := n + k;
	UPDATE turnos u SET cliente_nombre = 'Consumidor final', cliente_email = '', cliente_telefono = '', nota = NULL
		WHERE u.tienda_id = t AND app.clave_comprador(u.cliente_email, u.cliente_nombre) = p_clave;
	GET DIAGNOSTICS k = ROW_COUNT; n := n + k;
	UPDATE suscripciones s SET email = 'borrado@tulocalweb.invalid', nombre = 'Consumidor final', telefono = '',
			direccion = '', cp = ''
		WHERE s.tienda_id = t AND app.clave_comprador(s.email, s.nombre) = p_clave;
	GET DIAGNOSTICS k = ROW_COUNT; n := n + k;
	DELETE FROM clientes c WHERE c.tienda_id = t AND (lower(c.email) = p_clave OR c.id::text = p_clave);
	GET DIAGNOSTICS k = ROW_COUNT; n := n + k;
	IF v_email IS NOT NULL THEN
		UPDATE preguntas q SET nombre = 'Anónimo', email = NULL WHERE q.tienda_id = t AND lower(q.email) = v_email;
		GET DIAGNOSTICS k = ROW_COUNT; n := n + k;
		DELETE FROM carritos c WHERE c.tienda_id = t AND lower(c.email) = v_email;
		GET DIAGNOSTICS k = ROW_COUNT; n := n + k;
		DELETE FROM newsletter_entregas e WHERE e.tienda_id = t AND lower(e.email) = v_email;
		DELETE FROM suscriptores s WHERE s.tienda_id = t AND lower(s.email) = v_email;
		GET DIAGNOSTICS k = ROW_COUNT; n := n + k;
		DELETE FROM avisos_lanzamiento a WHERE a.tienda_id = t AND lower(a.email) = v_email;
		GET DIAGNOSTICS k = ROW_COUNT; n := n + k;
		DELETE FROM avisos_stock a WHERE a.tienda_id = t AND a.email = v_email;
		GET DIAGNOSTICS k = ROW_COUNT; n := n + k;
	END IF;
	RETURN n;
END
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION app.suprimir_comprador(p_clave text) RETURNS int
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app, pg_temp AS $$
DECLARE
	t uuid := app.tienda_actual();
	n int := 0;
	k int;
	v_email text;
	v_pedidos int[];
BEGIN
	IF t IS NULL OR coalesce(p_clave, '') = '' THEN
		RETURN 0;
	END IF;
	-- El email de quien se suprime (de la clave o de su ficha sin email no hay).
	v_email := CASE WHEN position('@' IN p_clave) > 0 THEN lower(p_clave) END;
	SELECT array_agg(p.numero) INTO v_pedidos FROM pedidos p
		WHERE p.tienda_id = t AND app.clave_comprador(p.comprador->>'email', p.comprador->>'nombre') = p_clave;
	IF EXISTS (SELECT 1 FROM pedidos p WHERE p.tienda_id = t AND p.numero = ANY (v_pedidos)
			AND p.estado NOT IN ('Entregado', 'Cancelado'))
		OR EXISTS (SELECT 1 FROM turnos u WHERE u.tienda_id = t AND u.estado IN ('reservado', 'confirmado') AND u.fin > now()
			AND app.clave_comprador(u.cliente_email, u.cliente_nombre) = p_clave)
		OR EXISTS (SELECT 1 FROM suscripciones s WHERE s.tienda_id = t AND s.estado IN ('pendiente', 'activa', 'pausada')
			AND app.clave_comprador(s.email, s.nombre) = p_clave)
		OR EXISTS (SELECT 1 FROM devoluciones d WHERE d.tienda_id = t AND d.estado <> 'Aprobada'
			AND (d.pedido_numero = ANY (v_pedidos) OR app.clave_comprador(d.email, d.nombre) = p_clave)) THEN
		RAISE EXCEPTION 'comprador_en_curso' USING ERRCODE = 'P0001';
	END IF;

	-- Pedidos: quedan como venta de «Consumidor final» (la lista de Clientes ya no los cuenta como
	-- cliente), sin la dirección ni la nota de la entrega.
	UPDATE pedidos p SET comprador = '{"nombre": "Consumidor final"}', cliente_id = NULL, nota = NULL,
			entrega = CASE WHEN p.entrega IS NULL THEN NULL ELSE p.entrega - 'direccion' - 'cp' END
		WHERE p.tienda_id = t AND p.numero = ANY (v_pedidos);
	GET DIAGNOSTICS k = ROW_COUNT; n := n + k;
	UPDATE resenas r SET autor = 'Anónimo' WHERE r.tienda_id = t AND r.pedido_numero = ANY (v_pedidos);
	GET DIAGNOSTICS k = ROW_COUNT; n := n + k;
	UPDATE devoluciones d SET nombre = 'Consumidor final', email = NULL
		WHERE d.tienda_id = t AND (d.pedido_numero = ANY (v_pedidos) OR app.clave_comprador(d.email, d.nombre) = p_clave);
	GET DIAGNOSTICS k = ROW_COUNT; n := n + k;
	UPDATE historial_pedidos h SET cliente_email = NULL, cliente_nombre = ''
		WHERE h.tienda_id = t AND app.clave_comprador(h.cliente_email, h.cliente_nombre) = p_clave;
	GET DIAGNOSTICS k = ROW_COUNT; n := n + k;
	UPDATE turnos u SET cliente_nombre = 'Consumidor final', cliente_email = '', cliente_telefono = '', nota = NULL
		WHERE u.tienda_id = t AND app.clave_comprador(u.cliente_email, u.cliente_nombre) = p_clave;
	GET DIAGNOSTICS k = ROW_COUNT; n := n + k;
	UPDATE suscripciones s SET email = 'borrado@tulocalweb.invalid', nombre = 'Consumidor final', telefono = '',
			direccion = '', cp = ''
		WHERE s.tienda_id = t AND app.clave_comprador(s.email, s.nombre) = p_clave;
	GET DIAGNOSTICS k = ROW_COUNT; n := n + k;
	DELETE FROM clientes c WHERE c.tienda_id = t AND (lower(c.email) = p_clave OR c.id::text = p_clave);
	GET DIAGNOSTICS k = ROW_COUNT; n := n + k;
	IF v_email IS NOT NULL THEN
		UPDATE preguntas q SET nombre = 'Anónimo', email = NULL WHERE q.tienda_id = t AND lower(q.email) = v_email;
		GET DIAGNOSTICS k = ROW_COUNT; n := n + k;
		DELETE FROM carritos c WHERE c.tienda_id = t AND lower(c.email) = v_email;
		GET DIAGNOSTICS k = ROW_COUNT; n := n + k;
		DELETE FROM newsletter_entregas e WHERE e.tienda_id = t AND lower(e.email) = v_email;
		DELETE FROM suscriptores s WHERE s.tienda_id = t AND lower(s.email) = v_email;
		GET DIAGNOSTICS k = ROW_COUNT; n := n + k;
		DELETE FROM avisos_lanzamiento a WHERE a.tienda_id = t AND lower(a.email) = v_email;
		GET DIAGNOSTICS k = ROW_COUNT; n := n + k;
	END IF;
	RETURN n;
END
$$;
-- +goose StatementEnd
DROP TRIGGER variantes_stock_repuesto ON variantes;
DROP TRIGGER productos_stock_repuesto ON productos;
DROP FUNCTION app.stock_repuesto();
DROP TABLE avisos_stock;
