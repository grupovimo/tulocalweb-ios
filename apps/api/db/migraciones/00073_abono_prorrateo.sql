-- +goose Up
-- El abono con prorrateo (§56). Lo que hasta ahora se calculaba (el próximo cobro, desde el alta) pasa a
-- ser lo que dice Mercado Pago, y los cambios que rigen en el próximo cobro (bajar, pasar a gratis,
-- cambiar de ciclo) quedan programados aparte de la suscripción que espera autorización:
--   plan_pendiente / ciclo_pendiente / espera_desde: un débito creado que todavía no autorizó la dueña
--     (espera_desde: el débito arranca recién ahí, sin cobrar nada antes).
--   plan_programado / ciclo_programado / programado_desde: el plan que rige desde esa fecha (el worker
--     lo aplica cuando llega).
ALTER TABLE tiendas
	ADD COLUMN abono_proximo timestamptz,
	ADD COLUMN espera_desde timestamptz,
	ADD COLUMN plan_programado text CHECK (plan_programado IN ('gratis', 'emprende', 'pro', 'negocio')),
	ADD COLUMN ciclo_programado text CHECK (ciclo_programado IN ('mes', 'anio')),
	ADD COLUMN programado_desde timestamptz,
	ADD CONSTRAINT tiendas_programado_completo CHECK ((plan_programado IS NULL) = (programado_desde IS NULL));
GRANT UPDATE (abono_proximo, espera_desde, plan_programado, ciclo_programado, programado_desde) ON tiendas TO mostrador_app;

-- Los cobros del abono (cada débito y cada diferencia prorrateada): la lista de Facturas de la pantalla.
-- La clave es la idempotencia de los avisos: «cobro:<id>» (débito) o «pago:<id>» (diferencia).
CREATE TABLE abono_cobros (
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	clave text NOT NULL CHECK (length(clave) <= 100),
	detalle text NOT NULL CHECK (length(detalle) <= 200),
	importe bigint NOT NULL CHECK (importe >= 0),
	creado_en timestamptz NOT NULL DEFAULT now(),
	PRIMARY KEY (tienda_id, clave)
);
CREATE INDEX abono_cobros_fecha ON abono_cobros (tienda_id, creado_en DESC);
SELECT app.aplicar_rls('abono_cobros');
GRANT SELECT, INSERT ON abono_cobros TO mostrador_app;

-- Para el worker, que pasa por todas las tiendas (sólo ids, como app.tiendas_con_arca_por_vencer): los
-- cambios programados que ya llegaron a su fecha.
CREATE POLICY abonos_programados ON tiendas FOR SELECT TO mostrador_migrador
	USING (current_setting('app.abonos_programados', true) = 'si');
-- +goose StatementBegin
CREATE FUNCTION app.abonos_programados(p_limite int) RETURNS SETOF uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
BEGIN
	PERFORM set_config('app.abonos_programados', 'si', true);
	RETURN QUERY SELECT t.id FROM tiendas t WHERE t.programado_desde <= now() ORDER BY t.programado_desde LIMIT p_limite;
	PERFORM set_config('app.abonos_programados', '', true);
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.abonos_programados(int) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.abonos_programados(int) TO mostrador_app;
CREATE INDEX tiendas_programado ON tiendas (programado_desde) WHERE programado_desde IS NOT NULL;

-- +goose Down
DROP INDEX tiendas_programado;
DROP FUNCTION app.abonos_programados(int);
DROP POLICY abonos_programados ON tiendas;
DROP TABLE abono_cobros;
ALTER TABLE tiendas DROP CONSTRAINT tiendas_programado_completo, DROP COLUMN abono_proximo, DROP COLUMN espera_desde,
	DROP COLUMN plan_programado, DROP COLUMN ciclo_programado, DROP COLUMN programado_desde;
