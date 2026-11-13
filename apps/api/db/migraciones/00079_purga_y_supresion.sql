-- +goose Up
-- Lo que la política de privacidad promete y faltaba hacer (legales/aaip/pendientes.md §3.2):
--
-- 1. Las tiendas cerradas se borran a los 90 días. El worker (con MOSTRADOR_PURGA_TIENDAS=1) llama
--    a app.purgar_tienda: guarda en archivo_contable lo que la ley obliga a conservar 10 años (los
--    comprobantes emitidos con su CAE y los cobros del abono), sin datos de quien compró, y borra la
--    tienda: todo lo demás cae por ON DELETE CASCADE (productos, pedidos, clientes, turnos, etc.).
-- 2. La supresión de un comprador (art. 16 de la Ley 25.326) y la rectificación de su nombre, email y
--    teléfono, desde la lista de Clientes. Y el borrado de los clientes inactivos (sin compras en 3
--    años), que el panel programa a 7 días y el worker hace.
--
-- Todo con funciones SECURITY DEFINER, como app.abonos_programados: la app no gana permisos sobre las
-- tablas, y las funciones de una tienda usan la que fijó EnTienda (app.tienda_actual()).

-- Lo contable de las tiendas ya borradas. Sin FK: la tienda no existe más. Lo lee staff (soporte y
-- administración, ante un pedido de ARCA); la app no lo ve.
CREATE TABLE archivo_contable (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	tienda_id uuid NOT NULL,
	tipo text NOT NULL CHECK (tipo IN ('comprobante', 'abono')),
	datos jsonb NOT NULL CHECK (jsonb_typeof(datos) = 'object' AND octet_length(datos::text) <= 4000),
	-- La del comprobante o la del cobro: desde acá corren los 10 años.
	fecha timestamptz NOT NULL,
	archivado_en timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX archivo_contable_fecha ON archivo_contable (fecha);
CREATE INDEX archivo_contable_tienda ON archivo_contable (tienda_id);
SELECT app.aplicar_rls('archivo_contable');
CREATE POLICY vencido_lee ON archivo_contable FOR SELECT TO mostrador_migrador
	USING (current_setting('app.archivo_vencido', true) = 'si' AND fecha < now() - interval '10 years');
CREATE POLICY vencido_borra ON archivo_contable FOR DELETE TO mostrador_migrador
	USING (current_setting('app.archivo_vencido', true) = 'si' AND fecha < now() - interval '10 years');

-- Para el worker y `mostradorctl purga-tiendas --listar`: las tiendas cerradas hace más de p_dias.
CREATE POLICY para_purgar ON tiendas FOR SELECT TO mostrador_migrador
	USING (current_setting('app.para_purgar', true) = 'si');
CREATE INDEX tiendas_cerradas ON tiendas (cerrada_en) WHERE cerrada_en IS NOT NULL;
-- +goose StatementBegin
CREATE FUNCTION app.tiendas_para_purgar(p_dias int, p_limite int)
RETURNS TABLE (id uuid, nombre text, cerrada_en timestamptz)
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app, pg_temp AS $$
BEGIN
	PERFORM set_config('app.para_purgar', 'si', true);
	RETURN QUERY SELECT t.id, t.nombre, t.cerrada_en FROM tiendas t
		WHERE t.cerrada_en < now() - make_interval(days => p_dias)
		ORDER BY t.cerrada_en LIMIT p_limite;
	PERFORM set_config('app.para_purgar', '', true);
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.tiendas_para_purgar(int, int) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.tiendas_para_purgar(int, int) TO mostrador_app;

-- Borra una tienda cerrada hace más de p_dias (lo vuelve a mirar acá, con la fila tomada: una tienda
-- que se reabrió en el medio no se borra). Devuelve si la borró.
-- +goose StatementBegin
CREATE FUNCTION app.purgar_tienda(p_tienda uuid, p_dias int) RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app, pg_temp AS $$
DECLARE
	v_cerrada timestamptz;
	v_nombre text;
	v_emisor jsonb;
BEGIN
	PERFORM set_config('app.tienda_id', p_tienda::text, true);
	SELECT t.cerrada_en, t.nombre INTO v_cerrada, v_nombre FROM tiendas t WHERE t.id = p_tienda FOR UPDATE;
	IF v_cerrada IS NULL OR v_cerrada >= now() - make_interval(days => p_dias) THEN
		RETURN false;
	END IF;
	-- Quien emitió (la tienda ante ARCA): sin esto el comprobante no se puede ubicar.
	SELECT jsonb_build_object('cuit', a.cuit, 'razonSocial', a.razon_social, 'condicion', a.condicion)
		INTO v_emisor FROM arca_config a WHERE a.tienda_id = p_tienda;
	-- Los comprobantes emitidos (con CAE): número, tipo, importes y concepto. De quien recibió sólo
	-- queda el tipo de documento y la condición ante el IVA; el original completo está en ARCA.
	INSERT INTO archivo_contable (tienda_id, tipo, datos, fecha)
	SELECT p_tienda, 'comprobante', jsonb_strip_nulls(jsonb_build_object(
			'emisor', v_emisor, 'tipo', c.tipo, 'puntoVenta', c.punto_venta, 'numero', c.numero,
			'cae', c.cae, 'caeVence', c.cae_vence, 'total', c.total, 'neto', c.neto, 'iva', c.iva,
			'concepto', c.concepto, 'pedido', c.pedido_numero,
			'asociado', CASE WHEN a.id IS NOT NULL THEN jsonb_build_object('tipo', a.tipo, 'puntoVenta', a.punto_venta, 'numero', a.numero) END,
			'receptor', jsonb_build_object('docTipo', c.receptor->'docTipo', 'condicion', c.receptor->'condicion'))),
		coalesce(c.emitido_en, c.creado_en)
	FROM comprobantes c LEFT JOIN comprobantes a ON a.id = c.asociado
	WHERE c.tienda_id = p_tienda AND c.estado = 'emitido';
	-- Lo que Tu Local Web le cobró a la tienda (el abono): son los registros contables propios.
	INSERT INTO archivo_contable (tienda_id, tipo, datos, fecha)
	SELECT p_tienda, 'abono', jsonb_build_object('tienda', v_nombre, 'clave', ac.clave, 'detalle', ac.detalle, 'importe', ac.importe),
		ac.creado_en
	FROM abono_cobros ac WHERE ac.tienda_id = p_tienda;
	DELETE FROM tiendas WHERE id = p_tienda;
	RETURN true;
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.purgar_tienda(uuid, int) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.purgar_tienda(uuid, int) TO mostrador_app;

-- Pasados los 10 años, lo archivado se borra. Devuelve cuántos.
-- +goose StatementBegin
CREATE FUNCTION app.purgar_archivo_contable() RETURNS bigint
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app, pg_temp AS $$
DECLARE n bigint;
BEGIN
	PERFORM set_config('app.archivo_vencido', 'si', true);
	DELETE FROM archivo_contable WHERE fecha < now() - interval '10 years';
	GET DIAGNOSTICS n = ROW_COUNT;
	PERFORM set_config('app.archivo_vencido', '', true);
	RETURN n;
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.purgar_archivo_contable() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.purgar_archivo_contable() TO mostrador_app;

-- La clave de un comprador es la de la lista de Clientes (listas.claveCliente): su email en minúsculas
-- o, si compró sin email, su nombre; una ficha sin email se nombra por su id.
-- +goose StatementBegin
CREATE FUNCTION app.clave_comprador(p_email text, p_nombre text) RETURNS text
LANGUAGE sql IMMUTABLE AS $$
	SELECT coalesce(lower(nullif(p_email, '')), nullif(p_nombre, ''))
$$;
-- +goose StatementEnd

-- Suprime a un comprador de la tienda fijada: borra su ficha, sus carritos, su suscripción al
-- newsletter y sus avisos de lanzamiento, y deja sus pedidos, turnos, reseñas, preguntas, devoluciones
-- y suscripciones sin datos personales (las ventas siguen contando para la caja). Las facturas
-- emitidas no se tocan: la ley obliga a guardarlas como salieron. Si tiene algo en curso (un pedido sin
-- entregar, un turno por venir, una suscripción viva o una devolución pendiente), no hace nada y avisa
-- con 'comprador_en_curso'. Devuelve cuántos registros tocó.
-- +goose StatementBegin
CREATE FUNCTION app.suprimir_comprador(p_clave text) RETURNS int
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
REVOKE ALL ON FUNCTION app.suprimir_comprador(text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.suprimir_comprador(text) TO mostrador_app;

-- Rectifica el nombre, el email y el teléfono de un comprador de la tienda fijada: en su ficha, en sus
-- pedidos y en lo importado de otra plataforma (de ahí sale la lista de Clientes). Las facturas
-- emitidas no cambian. Si el email nuevo ya es de otro cliente, avisa con 'email_en_uso'. Devuelve
-- cuántos registros tocó; Go valida los datos antes.
-- +goose StatementBegin
CREATE FUNCTION app.rectificar_comprador(p_clave text, p_nombre text, p_email text, p_telefono text) RETURNS int
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app, pg_temp AS $$
DECLARE
	t uuid := app.tienda_actual();
	n int := 0;
	k int;
	v_email text := lower(nullif(p_email, ''));
	v_tel text := nullif(p_telefono, '');
BEGIN
	IF t IS NULL OR coalesce(p_clave, '') = '' OR coalesce(p_nombre, '') = '' THEN
		RETURN 0;
	END IF;
	IF v_email IS NOT NULL AND v_email <> p_clave AND (
		EXISTS (SELECT 1 FROM pedidos p WHERE p.tienda_id = t AND lower(p.comprador->>'email') = v_email)
		OR EXISTS (SELECT 1 FROM clientes c WHERE c.tienda_id = t AND lower(c.email) = v_email AND c.id::text <> p_clave)
		OR EXISTS (SELECT 1 FROM historial_pedidos h WHERE h.tienda_id = t AND lower(h.cliente_email) = v_email)) THEN
		RAISE EXCEPTION 'email_en_uso' USING ERRCODE = 'P0001';
	END IF;
	UPDATE pedidos p SET comprador = p.comprador || jsonb_build_object('nombre', p_nombre, 'email', v_email, 'telefono', v_tel)
		WHERE p.tienda_id = t AND app.clave_comprador(p.comprador->>'email', p.comprador->>'nombre') = p_clave;
	GET DIAGNOSTICS k = ROW_COUNT; n := n + k;
	UPDATE historial_pedidos h SET cliente_nombre = left(p_nombre, 120), cliente_email = v_email
		WHERE h.tienda_id = t AND app.clave_comprador(h.cliente_email, h.cliente_nombre) = p_clave;
	GET DIAGNOSTICS k = ROW_COUNT; n := n + k;
	UPDATE clientes c SET nombre = p_nombre, email = v_email, telefono = v_tel
		WHERE c.tienda_id = t AND (lower(c.email) = p_clave OR c.id::text = p_clave);
	GET DIAGNOSTICS k = ROW_COUNT; n := n + k;
	RETURN n;
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.rectificar_comprador(text, text, text, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.rectificar_comprador(text, text, text, text) TO mostrador_app;

-- Para el worker: las tiendas cuyo borrado de clientes inactivos ya llegó a su fecha.
CREATE POLICY inactivos_programados ON tiendas FOR SELECT TO mostrador_migrador
	USING (current_setting('app.inactivos_programados', true) = 'si');
-- +goose StatementBegin
CREATE FUNCTION app.inactivos_programados(p_limite int) RETURNS SETOF uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app, pg_temp AS $$
BEGIN
	PERFORM set_config('app.inactivos_programados', 'si', true);
	RETURN QUERY SELECT t.id FROM tiendas t WHERE t.borrado_inactivos <= now() AND t.cerrada_en IS NULL
		ORDER BY t.borrado_inactivos LIMIT p_limite;
	PERFORM set_config('app.inactivos_programados', '', true);
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.inactivos_programados(int) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.inactivos_programados(int) TO mostrador_app;

-- +goose Down
DROP FUNCTION app.inactivos_programados(int);
DROP POLICY inactivos_programados ON tiendas;
DROP FUNCTION app.rectificar_comprador(text, text, text, text);
DROP FUNCTION app.suprimir_comprador(text);
DROP FUNCTION app.clave_comprador(text, text);
DROP FUNCTION app.purgar_archivo_contable();
DROP FUNCTION app.purgar_tienda(uuid, int);
DROP FUNCTION app.tiendas_para_purgar(int, int);
DROP INDEX tiendas_cerradas;
DROP POLICY para_purgar ON tiendas;
DROP TABLE archivo_contable;
