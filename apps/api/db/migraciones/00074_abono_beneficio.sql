-- +goose Up
-- El beneficio para quedarse (§57): el descuento que acepta la dueña al cancelar. Se cobra de verdad: el
-- débito de Mercado Pago baja a beneficio_pct % por beneficio_cobros cobros y vuelve al precio solo.
--   beneficio_en: cuándo lo aceptó (una vez por tienda: con fecha, no se ofrece más).
--   contacto_en: cuándo pidió que alguien del equipo le escriba.
ALTER TABLE tiendas
	ADD COLUMN beneficio_pct int NOT NULL DEFAULT 0 CHECK (beneficio_pct BETWEEN 0 AND 90),
	ADD COLUMN beneficio_cobros int NOT NULL DEFAULT 0 CHECK (beneficio_cobros BETWEEN 0 AND 12),
	ADD COLUMN beneficio_en timestamptz,
	ADD COLUMN contacto_en timestamptz;
GRANT UPDATE (beneficio_pct, beneficio_cobros, beneficio_en, contacto_en) ON tiendas TO mostrador_app;

-- +goose Down
ALTER TABLE tiendas
	DROP COLUMN beneficio_pct,
	DROP COLUMN beneficio_cobros,
	DROP COLUMN beneficio_en,
	DROP COLUMN contacto_en;
