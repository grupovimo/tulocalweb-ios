-- +goose Up
-- Las tiendas de la persona traen también el logo que se cargó en Mi tienda (el del borrador, que es
-- el último que eligió; si no, el publicado): la app y el panel lo muestran en vez de las iniciales.
DROP FUNCTION app.tiendas_de_usuario(uuid);
-- +goose StatementBegin
CREATE FUNCTION app.tiendas_de_usuario(p_usuario uuid) RETURNS TABLE (id uuid, nombre text, slug text, rol text, logo text)
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
DECLARE
	m record;
	antes text := current_setting('app.tienda_id', true);
BEGIN
	FOR m IN SELECT mm.tienda_id, mm.rol FROM auth.membresias mm WHERE mm.usuario_id = p_usuario ORDER BY mm.creada_en LOOP
		PERFORM set_config('app.tienda_id', m.tienda_id::text, true);
		RETURN QUERY SELECT t.id, t.nombre, t.slug, m.rol,
			coalesce(nullif(t.diseno_borrador->>'logo', ''), nullif(t.diseno_publicado->>'logo', ''))
			FROM tiendas t WHERE t.id = m.tienda_id AND t.cerrada_en IS NULL;
	END LOOP;
	PERFORM set_config('app.tienda_id', coalesce(antes, ''), true);
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.tiendas_de_usuario(uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.tiendas_de_usuario(uuid) TO mostrador_app;

-- +goose Down
DROP FUNCTION app.tiendas_de_usuario(uuid);
-- +goose StatementBegin
CREATE FUNCTION app.tiendas_de_usuario(p_usuario uuid) RETURNS TABLE (id uuid, nombre text, slug text, rol text)
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
DECLARE
	m record;
	antes text := current_setting('app.tienda_id', true);
BEGIN
	FOR m IN SELECT mm.tienda_id, mm.rol FROM auth.membresias mm WHERE mm.usuario_id = p_usuario ORDER BY mm.creada_en LOOP
		PERFORM set_config('app.tienda_id', m.tienda_id::text, true);
		RETURN QUERY SELECT t.id, t.nombre, t.slug, m.rol FROM tiendas t WHERE t.id = m.tienda_id AND t.cerrada_en IS NULL;
	END LOOP;
	PERFORM set_config('app.tienda_id', coalesce(antes, ''), true);
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.tiendas_de_usuario(uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.tiendas_de_usuario(uuid) TO mostrador_app;
