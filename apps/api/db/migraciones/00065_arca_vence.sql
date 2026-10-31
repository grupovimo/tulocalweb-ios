-- +goose Up
-- Aviso de vencimiento del certificado de ARCA (§48): a 30 días, a 7 y al vencer. aviso_vence guarda
-- el último umbral avisado (30, 7 o 0); subir un certificado nuevo lo vuelve a NULL.
ALTER TABLE arca_config ADD COLUMN aviso_vence int CHECK (aviso_vence IN (0, 7, 30));
GRANT UPDATE (aviso_vence) ON arca_config TO mostrador_app;

-- Las tiendas con el certificado por vencer y un umbral sin avisar (para el worker, que pasa por
-- todas): sólo ids.
CREATE POLICY por_vencer ON arca_config FOR SELECT TO mostrador_migrador
	USING (current_setting('app.arca_por_vencer', true) = 'si');
-- +goose StatementBegin
CREATE FUNCTION app.tiendas_con_arca_por_vencer(p_limite int) RETURNS SETOF uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
BEGIN
	PERFORM set_config('app.arca_por_vencer', 'si', true);
	RETURN QUERY SELECT c.tienda_id FROM arca_config c
		WHERE c.certificado IS NOT NULL AND c.certificado_vence < now() + interval '30 days'
			AND (c.aviso_vence IS NULL
				OR (c.aviso_vence > 7 AND c.certificado_vence < now() + interval '7 days')
				OR (c.aviso_vence > 0 AND c.certificado_vence <= now()))
		LIMIT p_limite;
	PERFORM set_config('app.arca_por_vencer', '', true);
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.tiendas_con_arca_por_vencer(int) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.tiendas_con_arca_por_vencer(int) TO mostrador_app;

-- +goose Down
DROP FUNCTION app.tiendas_con_arca_por_vencer(int);
DROP POLICY por_vencer ON arca_config;
ALTER TABLE arca_config DROP COLUMN aviso_vence;
