-- +goose Up
-- La Live Activity de los pedidos por preparar en el iPhone (docs/decisiones.md §54). Cada iPhone
-- guarda, además de su token de avisos, el token de la actividad en curso (para actualizarla o
-- cerrarla por APNs) y el de «push to start» (iOS 17.2+, para arrancarla cuando entra un pedido con
-- la app cerrada). Son de la persona, como el token de avisos.
ALTER TABLE auth.dispositivos
	ADD COLUMN token_actividad text CHECK (token_actividad ~ '^[0-9a-f]+$' AND length(token_actividad) BETWEEN 32 AND 400),
	ADD COLUMN token_inicio text CHECK (token_inicio ~ '^[0-9a-f]+$' AND length(token_inicio) BETWEEN 32 AND 400);

-- Un token de actividad que APNs da por muerto se borra por el token, sin saber de quién era (desde
-- la función de abajo, que fija cuál).
CREATE POLICY por_token_actividad ON auth.dispositivos FOR UPDATE TO mostrador_migrador
	USING (token_actividad = current_setting('app.actividad_token', true)
		OR token_inicio = current_setting('app.actividad_token', true));
CREATE POLICY por_token_actividad_lee ON auth.dispositivos FOR SELECT TO mostrador_migrador
	USING (token_actividad = current_setting('app.actividad_token', true)
		OR token_inicio = current_setting('app.actividad_token', true));

-- La app avisa el token de la actividad (null: terminó) y el de inicio. El iPhone tiene que ser de
-- la sesión vigente de quien llama.
-- +goose StatementBegin
CREATE FUNCTION app.actividad_dispositivo(p_sesion text, p_token text, p_actividad text, p_inicio text)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app, pg_temp AS $$
DECLARE
	u uuid := app.usuario_actual();
	n int;
BEGIN
	IF u IS NULL OR NOT EXISTS (SELECT 1 FROM auth.sesiones s
		WHERE s.token_hash = p_sesion AND s.usuario_id = u AND s.vence_en > now()) THEN
		RAISE EXCEPTION 'sesión ajena o vencida' USING ERRCODE = '42501';
	END IF;
	UPDATE auth.dispositivos
		SET token_actividad = nullif(p_actividad, ''),
			token_inicio = coalesce(nullif(p_inicio, ''), token_inicio),
			actualizado_en = now()
		WHERE token = p_token AND usuario_id = u AND sesion_hash = p_sesion;
	GET DIAGNOSTICS n = ROW_COUNT;
	RETURN n > 0;
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.actividad_dispositivo(text, text, text, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.actividad_dispositivo(text, text, text, text) TO mostrador_app;

-- Los iPhone del equipo de la tienda con actividad o con token de inicio (como dispositivos_de_tienda).
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

-- +goose StatementBegin
CREATE FUNCTION app.olvidar_actividad(p_token text) RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app, pg_temp AS $$
BEGIN
	PERFORM set_config('app.actividad_token', p_token, true);
	UPDATE auth.dispositivos SET token_actividad = NULL WHERE token_actividad = p_token;
	UPDATE auth.dispositivos SET token_inicio = NULL WHERE token_inicio = p_token;
	PERFORM set_config('app.actividad_token', '', true);
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.olvidar_actividad(text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.olvidar_actividad(text) TO mostrador_app;

-- +goose Down
DROP FUNCTION app.olvidar_actividad(text);
DROP FUNCTION app.actividades_de_tienda();
DROP FUNCTION app.actividad_dispositivo(text, text, text, text);
DROP POLICY por_token_actividad_lee ON auth.dispositivos;
DROP POLICY por_token_actividad ON auth.dispositivos;
ALTER TABLE auth.dispositivos DROP COLUMN token_inicio, DROP COLUMN token_actividad;
