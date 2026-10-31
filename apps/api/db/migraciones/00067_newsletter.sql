-- +goose Up
-- Newsletter de cada tienda (docs/decisiones.md §50). Quien se suma desde la tienda online (el bloque
-- del pie o la casilla del checkout) queda pendiente hasta tocar el link del mail de confirmación
-- (doble opt-in). El token no se adivina y sirve para confirmar y para la baja con un clic; la baja
-- borra la fila (no queda el email guardado).
CREATE TABLE suscriptores (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	email text NOT NULL CHECK (length(email) BETWEEN 3 AND 120),
	origen text NOT NULL CHECK (origen IN ('tienda', 'checkout')),
	token text NOT NULL UNIQUE CHECK (token ~ '^[A-Za-z0-9_-]{20,64}$'),
	creado timestamptz NOT NULL DEFAULT now(),
	confirmado_en timestamptz
);
CREATE UNIQUE INDEX suscriptores_email ON suscriptores (tienda_id, lower(email));
CREATE INDEX suscriptores_recientes ON suscriptores (tienda_id, creado DESC);
SELECT app.aplicar_rls('suscriptores');
GRANT SELECT, INSERT, UPDATE (origen, creado, confirmado_en, token), DELETE ON suscriptores TO mostrador_app;

-- Cada «Mandar novedad»: asunto, texto y hasta 4 productos. La lista de destinatarios se fija al
-- mandar (los confirmados de ese momento) y el worker la recorre de a tandas, con el tope diario.
CREATE TABLE newsletter_envios (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	asunto text NOT NULL CHECK (length(asunto) BETWEEN 3 AND 120),
	texto text NOT NULL CHECK (length(texto) BETWEEN 10 AND 5000),
	productos jsonb NOT NULL DEFAULT '[]' CHECK (jsonb_typeof(productos) = 'array' AND jsonb_array_length(productos) <= 4),
	destinatarios int NOT NULL CHECK (destinatarios >= 0),
	creado timestamptz NOT NULL DEFAULT now(),
	terminado timestamptz
);
CREATE INDEX newsletter_envios_recientes ON newsletter_envios (tienda_id, creado DESC);
SELECT app.aplicar_rls('newsletter_envios');
GRANT SELECT, INSERT, UPDATE (terminado) ON newsletter_envios TO mostrador_app;

-- Una fila por envío y destinatario: pendiente, enviado, omitido (se dio de baja en el medio) o
-- fallido (el servidor de correo lo rechazó tres veces). El email se copia: si la persona se borra,
-- la fila queda omitida sin perder la cuenta del historial.
CREATE TABLE newsletter_entregas (
	envio_id uuid NOT NULL REFERENCES newsletter_envios (id) ON DELETE CASCADE,
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	suscriptor_id uuid REFERENCES suscriptores (id) ON DELETE SET NULL,
	email text NOT NULL,
	estado text NOT NULL DEFAULT 'pendiente' CHECK (estado IN ('pendiente', 'enviado', 'omitido', 'fallido')),
	intentos int NOT NULL DEFAULT 0,
	enviado_en timestamptz,
	PRIMARY KEY (envio_id, email)
);
CREATE INDEX newsletter_entregas_pendientes ON newsletter_entregas (envio_id) WHERE estado = 'pendiente';
-- El tope diario cuenta lo enviado hoy por tienda.
CREATE INDEX newsletter_entregas_hoy ON newsletter_entregas (tienda_id, enviado_en) WHERE estado = 'enviado';
SELECT app.aplicar_rls('newsletter_entregas');
GRANT SELECT, INSERT, UPDATE (estado, intentos, enviado_en, suscriptor_id) ON newsletter_entregas TO mostrador_app;

-- +goose Down
DELETE FROM outbox WHERE tipo IN ('newsletter.confirmar', 'newsletter.tanda');
DROP TABLE newsletter_entregas;
DROP TABLE newsletter_envios;
DROP TABLE suscriptores;
