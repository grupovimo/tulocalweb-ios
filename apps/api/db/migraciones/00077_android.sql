-- +goose Up
-- La app de Android (docs/decisiones.md §40 y §54): cada teléfono dice su plataforma. El de Android
-- registra su token de FCM (no es hexadecimal) y el worker le manda los avisos y los pedidos por
-- preparar por Firebase, sólo con datos: la app arma la notificación con sus acciones. La
-- notificación en curso usa el mismo token del teléfono como «actividad» y como «inicio».
ALTER TABLE auth.dispositivos
	ADD COLUMN plataforma text NOT NULL DEFAULT 'ios' CHECK (plataforma IN ('ios', 'android')),
	DROP CONSTRAINT dispositivos_token_check,
	DROP CONSTRAINT dispositivos_token_actividad_check,
	DROP CONSTRAINT dispositivos_token_inicio_check;
ALTER TABLE auth.dispositivos
	ADD CONSTRAINT dispositivos_token_check CHECK (
		(plataforma = 'ios' AND token ~ '^[0-9a-f]{64,200}$')
		OR (plataforma = 'android' AND token ~ '^[A-Za-z0-9_:-]+$' AND length(token) BETWEEN 20 AND 4096)),
	ADD CONSTRAINT dispositivos_token_actividad_check CHECK (
		(plataforma = 'ios' AND token_actividad ~ '^[0-9a-f]+$' AND length(token_actividad) BETWEEN 32 AND 400)
		OR (plataforma = 'android' AND token_actividad = token)),
	ADD CONSTRAINT dispositivos_token_inicio_check CHECK (
		(plataforma = 'ios' AND token_inicio ~ '^[0-9a-f]+$' AND length(token_inicio) BETWEEN 32 AND 400)
		OR (plataforma = 'android' AND token_inicio = token));

-- Alta o renovación del teléfono de la sesión, ahora con la plataforma.
DROP FUNCTION app.registrar_dispositivo(text, text, text, text, text);
-- +goose StatementBegin
CREATE FUNCTION app.registrar_dispositivo(p_sesion text, p_token text, p_entorno text, p_modelo text, p_version text, p_plataforma text)
RETURNS uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app, pg_temp AS $$
DECLARE
	u uuid := app.usuario_actual();
	nuevo uuid;
BEGIN
	IF u IS NULL OR NOT EXISTS (SELECT 1 FROM auth.sesiones s
		WHERE s.token_hash = p_sesion AND s.usuario_id = u AND s.vence_en > now()) THEN
		RAISE EXCEPTION 'sesión ajena o vencida' USING ERRCODE = '42501';
	END IF;
	PERFORM set_config('app.dispositivo_token', p_token, true);
	DELETE FROM auth.dispositivos WHERE token = p_token;
	PERFORM set_config('app.dispositivo_token', '', true);
	INSERT INTO auth.dispositivos (usuario_id, sesion_hash, token, entorno, modelo, version_app, plataforma)
	VALUES (u, p_sesion, p_token, p_entorno, coalesce(p_modelo, ''), coalesce(p_version, ''), coalesce(nullif(p_plataforma, ''), 'ios'))
	RETURNING id INTO nuevo;
	RETURN nuevo;
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.registrar_dispositivo(text, text, text, text, text, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.registrar_dispositivo(text, text, text, text, text, text) TO mostrador_app;

-- Los teléfonos del equipo de la tienda, ahora con la plataforma (APNs o FCM).
DROP FUNCTION app.dispositivos_de_tienda();
-- +goose StatementBegin
CREATE FUNCTION app.dispositivos_de_tienda()
RETURNS TABLE (usuario_id uuid, rol text, token text, entorno text, plataforma text)
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app, pg_temp AS $$
DECLARE
	m record;
	antes text := current_setting('app.usuario_id', true);
BEGIN
	IF app.tienda_actual() IS NULL THEN
		RETURN;
	END IF;
	FOR m IN SELECT mm.usuario_id, mm.rol FROM auth.membresias mm WHERE mm.tienda_id = app.tienda_actual() LOOP
		PERFORM set_config('app.usuario_id', m.usuario_id::text, true);
		RETURN QUERY SELECT d.usuario_id, m.rol, d.token, d.entorno, d.plataforma
			FROM auth.dispositivos d
			JOIN auth.sesiones s ON s.token_hash = d.sesion_hash
			JOIN auth.usuarios u ON u.id = d.usuario_id
			WHERE d.usuario_id = m.usuario_id AND s.vence_en > now()
				AND NOT (coalesce((u.seguridad->>'inactivas')::boolean, false) AND s.visto_en < now() - interval '7 days');
	END LOOP;
	PERFORM set_config('app.usuario_id', coalesce(antes, ''), true);
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.dispositivos_de_tienda() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.dispositivos_de_tienda() TO mostrador_app;

-- Los teléfonos con actividad o con token de inicio, con la plataforma.
DROP FUNCTION app.actividades_de_tienda();
-- +goose StatementBegin
CREATE FUNCTION app.actividades_de_tienda()
RETURNS TABLE (rol text, token_actividad text, token_inicio text, entorno text, plataforma text)
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app, pg_temp AS $$
DECLARE
	m record;
	antes text := current_setting('app.usuario_id', true);
BEGIN
	IF app.tienda_actual() IS NULL THEN
		RETURN;
	END IF;
	FOR m IN SELECT mm.usuario_id, mm.rol FROM auth.membresias mm WHERE mm.tienda_id = app.tienda_actual() LOOP
		PERFORM set_config('app.usuario_id', m.usuario_id::text, true);
		RETURN QUERY SELECT m.rol, d.token_actividad, d.token_inicio, d.entorno, d.plataforma
			FROM auth.dispositivos d
			JOIN auth.sesiones s ON s.token_hash = d.sesion_hash
			WHERE d.usuario_id = m.usuario_id AND s.vence_en > now()
				AND (d.token_actividad IS NOT NULL OR d.token_inicio IS NOT NULL);
	END LOOP;
	PERFORM set_config('app.usuario_id', coalesce(antes, ''), true);
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.actividades_de_tienda() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.actividades_de_tienda() TO mostrador_app;

-- +goose Down
DROP FUNCTION app.actividades_de_tienda();
-- +goose StatementBegin
CREATE FUNCTION app.actividades_de_tienda()
RETURNS TABLE (rol text, token_actividad text, token_inicio text, entorno text)
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app, pg_temp AS $$
DECLARE
	m record;
	antes text := current_setting('app.usuario_id', true);
BEGIN
	IF app.tienda_actual() IS NULL THEN
		RETURN;
	END IF;
	FOR m IN SELECT mm.usuario_id, mm.rol FROM auth.membresias mm WHERE mm.tienda_id = app.tienda_actual() LOOP
		PERFORM set_config('app.usuario_id', m.usuario_id::text, true);
		RETURN QUERY SELECT m.rol, d.token_actividad, d.token_inicio, d.entorno
			FROM auth.dispositivos d
			JOIN auth.sesiones s ON s.token_hash = d.sesion_hash
			WHERE d.usuario_id = m.usuario_id AND s.vence_en > now()
				AND (d.token_actividad IS NOT NULL OR d.token_inicio IS NOT NULL);
	END LOOP;
	PERFORM set_config('app.usuario_id', coalesce(antes, ''), true);
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.actividades_de_tienda() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.actividades_de_tienda() TO mostrador_app;
DROP FUNCTION app.dispositivos_de_tienda();
-- +goose StatementBegin
CREATE FUNCTION app.dispositivos_de_tienda()
RETURNS TABLE (usuario_id uuid, rol text, token text, entorno text)
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app, pg_temp AS $$
DECLARE
	m record;
	antes text := current_setting('app.usuario_id', true);
BEGIN
	IF app.tienda_actual() IS NULL THEN
		RETURN;
	END IF;
	FOR m IN SELECT mm.usuario_id, mm.rol FROM auth.membresias mm WHERE mm.tienda_id = app.tienda_actual() LOOP
		PERFORM set_config('app.usuario_id', m.usuario_id::text, true);
		RETURN QUERY SELECT d.usuario_id, m.rol, d.token, d.entorno
			FROM auth.dispositivos d
			JOIN auth.sesiones s ON s.token_hash = d.sesion_hash
			JOIN auth.usuarios u ON u.id = d.usuario_id
			WHERE d.usuario_id = m.usuario_id AND s.vence_en > now()
				AND NOT (coalesce((u.seguridad->>'inactivas')::boolean, false) AND s.visto_en < now() - interval '7 days');
	END LOOP;
	PERFORM set_config('app.usuario_id', coalesce(antes, ''), true);
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.dispositivos_de_tienda() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.dispositivos_de_tienda() TO mostrador_app;
DROP FUNCTION app.registrar_dispositivo(text, text, text, text, text, text);
-- +goose StatementBegin
CREATE FUNCTION app.registrar_dispositivo(p_sesion text, p_token text, p_entorno text, p_modelo text, p_version text)
RETURNS uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app, pg_temp AS $$
DECLARE
	u uuid := app.usuario_actual();
	nuevo uuid;
BEGIN
	IF u IS NULL OR NOT EXISTS (SELECT 1 FROM auth.sesiones s
		WHERE s.token_hash = p_sesion AND s.usuario_id = u AND s.vence_en > now()) THEN
		RAISE EXCEPTION 'sesión ajena o vencida' USING ERRCODE = '42501';
	END IF;
	PERFORM set_config('app.dispositivo_token', p_token, true);
	DELETE FROM auth.dispositivos WHERE token = p_token;
	PERFORM set_config('app.dispositivo_token', '', true);
	INSERT INTO auth.dispositivos (usuario_id, sesion_hash, token, entorno, modelo, version_app)
	VALUES (u, p_sesion, p_token, p_entorno, coalesce(p_modelo, ''), coalesce(p_version, ''))
	RETURNING id INTO nuevo;
	RETURN nuevo;
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.registrar_dispositivo(text, text, text, text, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.registrar_dispositivo(text, text, text, text, text) TO mostrador_app;
DELETE FROM auth.dispositivos WHERE plataforma = 'android';
ALTER TABLE auth.dispositivos
	DROP CONSTRAINT dispositivos_token_check,
	DROP CONSTRAINT dispositivos_token_actividad_check,
	DROP CONSTRAINT dispositivos_token_inicio_check,
	DROP COLUMN plataforma;
ALTER TABLE auth.dispositivos
	ADD CONSTRAINT dispositivos_token_check CHECK (token ~ '^[0-9a-f]{64,200}$'),
	ADD CONSTRAINT dispositivos_token_actividad_check CHECK (token_actividad ~ '^[0-9a-f]+$' AND length(token_actividad) BETWEEN 32 AND 400),
	ADD CONSTRAINT dispositivos_token_inicio_check CHECK (token_inicio ~ '^[0-9a-f]+$' AND length(token_inicio) BETWEEN 32 AND 400);
