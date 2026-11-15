-- +goose Up
-- Promociones de la tienda (docs/especificaciones/tienda-promociones.md): ventanas que captan mails,
-- ruleta, cupones únicos y métricas por ventana y día.

-- Cupones únicos que genera una ventana (captación o ruleta): un uso, con vencimiento y su origen,
-- para atribuirle las ventas.
ALTER TABLE cupones ADD COLUMN origen text CHECK (origen IS NULL OR origen ~ '^ventana:[a-z0-9]{6,16}$');
CREATE INDEX cupones_origen ON cupones (tienda_id, origen) WHERE origen IS NOT NULL;

-- El contacto captado entra a la newsletter con su origen («ventana:<id>»), y con su nombre y WhatsApp
-- si la ventana los pidió. La supresión del comprador (app.suprimir_comprador) ya borra la fila.
ALTER TABLE suscriptores DROP CONSTRAINT suscriptores_origen_check;
ALTER TABLE suscriptores ADD CONSTRAINT suscriptores_origen_check
	CHECK (origen IN ('tienda', 'checkout') OR origen ~ '^ventana:[a-z0-9]{6,16}$');
ALTER TABLE suscriptores ADD COLUMN nombre text CHECK (nombre IS NULL OR length(nombre) BETWEEN 1 AND 60),
	ADD COLUMN whatsapp text CHECK (whatsapp IS NULL OR whatsapp ~ '^\+?[0-9 ()-]{6,24}$');
GRANT UPDATE (nombre, whatsapp) ON suscriptores TO mostrador_app;

-- Una captación por ventana y persona: el premio que se le dio (para responder lo mismo si vuelve) y,
-- en la ruleta, la porción. Sin datos personales: el email va como hash (sha256 del email en
-- minúsculas); el contacto vive en suscriptores.
CREATE TABLE promo_captaciones (
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	ventana text NOT NULL CHECK (ventana ~ '^[a-z0-9]{6,16}$'),
	email_hash bytea NOT NULL CHECK (length(email_hash) = 32),
	porcion int CHECK (porcion IS NULL OR porcion BETWEEN 0 AND 7),
	premio jsonb NOT NULL CHECK (jsonb_typeof(premio) = 'object' AND pg_column_size(premio) <= 1024),
	creado_en timestamptz NOT NULL DEFAULT now(),
	PRIMARY KEY (tienda_id, ventana, email_hash)
);
CREATE INDEX promo_captaciones_dia ON promo_captaciones (tienda_id, creado_en);
SELECT app.aplicar_rls('promo_captaciones');
GRANT SELECT, INSERT ON promo_captaciones TO mostrador_app;

-- Lo que manda la tienda en lote (navigator.sendBeacon), agregado por ventana y día de Argentina. Los
-- envíos salen de promo_captaciones y las ventas de los pedidos con el cupón de la ventana.
CREATE TABLE promo_metricas (
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	ventana text NOT NULL CHECK (ventana ~ '^[a-z0-9]{6,16}$'),
	dia date NOT NULL,
	vistas int NOT NULL DEFAULT 0 CHECK (vistas >= 0),
	cierres int NOT NULL DEFAULT 0 CHECK (cierres >= 0),
	clics int NOT NULL DEFAULT 0 CHECK (clics >= 0),
	PRIMARY KEY (tienda_id, ventana, dia)
);
SELECT app.aplicar_rls('promo_metricas');
GRANT SELECT, INSERT, UPDATE (vistas, cierres, clics) ON promo_metricas TO mostrador_app;

-- +goose Down
DROP TABLE promo_metricas;
DROP TABLE promo_captaciones;
ALTER TABLE suscriptores DROP COLUMN whatsapp, DROP COLUMN nombre;
DELETE FROM suscriptores WHERE origen LIKE 'ventana:%';
ALTER TABLE suscriptores DROP CONSTRAINT suscriptores_origen_check;
ALTER TABLE suscriptores ADD CONSTRAINT suscriptores_origen_check CHECK (origen IN ('tienda', 'checkout'));
DROP INDEX cupones_origen;
ALTER TABLE cupones DROP COLUMN origen;
