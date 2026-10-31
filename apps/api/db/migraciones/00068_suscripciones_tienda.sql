-- +goose Up
-- Suscripciones de la tienda (docs/decisiones.md §51): la tienda arma planes con cobro recurrente
-- (el club de vinos de Malbec, el café de Mostaza) y quien compra se suscribe desde la tienda con
-- la suscripción de Mercado Pago (preapproval) de la cuenta conectada de la tienda. Cada cobro
-- aprobado crea un pedido de la tienda online marcado como de suscripción.

CREATE TABLE suscripcion_planes (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	nombre text NOT NULL CHECK (length(nombre) BETWEEN 2 AND 60),
	descripcion text NOT NULL DEFAULT '' CHECK (length(descripcion) <= 400),
	foto text CHECK (length(foto) <= 2048),
	-- Lo que se cobra cada período, en pesos (envío incluido si el plan se envía).
	precio bigint NOT NULL CHECK (precio BETWEEN 100 AND 100000000),
	frecuencia text NOT NULL DEFAULT 'mensual' CHECK (frecuencia IN ('quincenal', 'mensual', 'bimestral')),
	-- [{productoId, varianteId, cantidad}] del catálogo: lo que se prepara con cada cobro.
	productos jsonb NOT NULL DEFAULT '[]' CHECK (jsonb_typeof(productos) = 'array' AND pg_column_size(productos) <= 8192),
	-- Texto libre de qué incluye ("6 botellas elegidas por la sommelier").
	incluye text NOT NULL DEFAULT '' CHECK (length(incluye) <= 400),
	envio boolean NOT NULL DEFAULT true,
	retiro boolean NOT NULL DEFAULT true,
	CHECK (envio OR retiro),
	-- Cupo de suscripciones vigentes (null = sin tope).
	cupo int CHECK (cupo IS NULL OR cupo BETWEEN 1 AND 100000),
	activo boolean NOT NULL DEFAULT true,
	orden int NOT NULL DEFAULT 0,
	creado_en timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX suscripcion_planes_tienda ON suscripcion_planes (tienda_id, orden);
SELECT app.aplicar_rls('suscripcion_planes');
GRANT SELECT, INSERT, DELETE ON suscripcion_planes TO mostrador_app;
-- UPDATE también hace falta para el SELECT … FOR UPDATE con que se mira el cupo al suscribirse.
GRANT UPDATE (nombre, descripcion, foto, precio, frecuencia, productos, incluye, envio, retiro, cupo, activo, orden)
	ON suscripcion_planes TO mostrador_app;

-- Cada suscripción copia nombre, precio y frecuencia del plan: es lo que se acordó con Mercado Pago,
-- aunque la tienda después edite el plan.
CREATE TABLE suscripciones (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	plan_id uuid NOT NULL REFERENCES suscripcion_planes (id),
	-- El link privado de quien se suscribió (/t/{slug}/suscripcion/{codigo}): ver, pausar, cancelar.
	codigo text NOT NULL UNIQUE CHECK (length(codigo) BETWEEN 16 AND 64),
	plan_nombre text NOT NULL,
	precio bigint NOT NULL CHECK (precio > 0),
	frecuencia text NOT NULL CHECK (frecuencia IN ('quincenal', 'mensual', 'bimestral')),
	email text NOT NULL CHECK (length(email) BETWEEN 3 AND 120),
	nombre text NOT NULL CHECK (length(nombre) BETWEEN 2 AND 80),
	telefono text NOT NULL DEFAULT '' CHECK (length(telefono) <= 20),
	entrega text NOT NULL CHECK (entrega IN ('envio', 'retiro')),
	direccion text NOT NULL DEFAULT '' CHECK (length(direccion) <= 200),
	cp text NOT NULL DEFAULT '' CHECK (length(cp) <= 10),
	-- pendiente (falta autorizar en Mercado Pago), activa, pausada o cancelada.
	estado text NOT NULL DEFAULT 'pendiente' CHECK (estado IN ('pendiente', 'activa', 'pausada', 'cancelada')),
	mp_preapproval text CHECK (length(mp_preapproval) <= 64),
	proximo_cobro timestamptz,
	cobros int NOT NULL DEFAULT 0,
	ultimo_cobro timestamptz,
	-- Quién la dio de baja: comprador (su link), tienda (el panel) o mercadopago.
	cancelada_por text CHECK (cancelada_por IN ('comprador', 'tienda', 'mercadopago')),
	-- Los mails que ya salieron (alta, baja…): el outbox entrega al menos una vez, el mail sale una.
	avisos jsonb NOT NULL DEFAULT '{}',
	creada_en timestamptz NOT NULL DEFAULT now(),
	activada_en timestamptz,
	cancelada_en timestamptz,
	actualizada_en timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX suscripciones_tienda ON suscripciones (tienda_id, creada_en DESC);
CREATE INDEX suscripciones_plan ON suscripciones (tienda_id, plan_id, estado);
CREATE UNIQUE INDEX suscripciones_preapproval ON suscripciones (tienda_id, mp_preapproval) WHERE mp_preapproval IS NOT NULL;
SELECT app.aplicar_rls('suscripciones');
GRANT SELECT, INSERT ON suscripciones TO mostrador_app;
GRANT UPDATE (estado, mp_preapproval, proximo_cobro, cobros, ultimo_cobro, cancelada_por, activada_en, cancelada_en, actualizada_en, avisos)
	ON suscripciones TO mostrador_app;

-- El aviso de una suscripción trae el id de Mercado Pago, no la tienda: esto dice de qué tienda es
-- (como app.tienda_de_mp; la RLS forzada también alcanza al dueño de la función).
CREATE POLICY por_preapproval ON suscripciones FOR SELECT TO mostrador_migrador
	USING (mp_preapproval = current_setting('app.mp_preapproval', true));
-- +goose StatementBegin
CREATE FUNCTION app.tienda_de_preapproval(p_id text) RETURNS uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
DECLARE
	t uuid;
BEGIN
	PERFORM set_config('app.mp_preapproval', p_id, true);
	SELECT tienda_id INTO t FROM suscripciones WHERE mp_preapproval = p_id LIMIT 1;
	PERFORM set_config('app.mp_preapproval', '', true);
	RETURN t;
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.tienda_de_preapproval(text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.tienda_de_preapproval(text) TO mostrador_app;

-- Un renglón por cobro (aprobado o rechazado). La clave es la idempotencia de los avisos de Mercado
-- Pago: «pago:<id>» crea un solo pedido aunque el aviso llegue mil veces; «rechazo:<id>» avisa una vez.
CREATE TABLE suscripcion_cobros (
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	clave text NOT NULL CHECK (length(clave) <= 100),
	suscripcion_id uuid NOT NULL REFERENCES suscripciones (id),
	estado text NOT NULL CHECK (estado IN ('aprobado', 'rechazado')),
	monto bigint NOT NULL DEFAULT 0,
	mp_pago bigint,
	pedido int,
	creado_en timestamptz NOT NULL DEFAULT now(),
	PRIMARY KEY (tienda_id, clave)
);
CREATE INDEX suscripcion_cobros_suscripcion ON suscripcion_cobros (tienda_id, suscripcion_id, creado_en DESC);
SELECT app.aplicar_rls('suscripcion_cobros');
GRANT SELECT, INSERT ON suscripcion_cobros TO mostrador_app;
GRANT UPDATE (pedido) ON suscripcion_cobros TO mostrador_app;

-- El pedido que salió de un cobro de suscripción.
ALTER TABLE pedidos ADD COLUMN suscripcion_id uuid;

-- +goose Down
DROP FUNCTION app.tienda_de_preapproval(text);
ALTER TABLE pedidos DROP COLUMN suscripcion_id;
DROP TABLE suscripcion_cobros;
DROP TABLE suscripciones;
DROP TABLE suscripcion_planes;
DELETE FROM outbox WHERE tipo IN ('suscripcion.alta', 'suscripcion.baja', 'suscripcion.rechazo', 'mp.suscripcion', 'mp.cobro_suscripcion');
