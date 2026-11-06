-- +goose Up
-- Verificación en dos pasos (docs/decisiones.md §58): un código de una app autenticadora (TOTP,
-- RFC 6238) además de la contraseña, Google, Apple, Mercado Pago o el link por WhatsApp. Es de la
-- persona, no de la tienda. El secreto va cifrado (cifrado.Cifrador, contexto «totp:<usuario>»); los
-- códigos de recuperación, sólo su hash.
ALTER TABLE auth.usuarios
	ADD COLUMN totp_secreto bytea,
	ADD COLUMN totp_pendiente bytea,
	ADD COLUMN totp_activo_en timestamptz,
	ADD COLUMN codigos_recuperacion text[] NOT NULL DEFAULT '{}';
GRANT UPDATE (totp_secreto, totp_pendiente, totp_activo_en, codigos_recuperacion) ON auth.usuarios TO mostrador_app;

-- El desafío: lo que queda entre «la contraseña está bien» y «el código está bien». Dura 10 minutos y
-- admite 5 intentos; con el código correcto se abre la sesión con la vida que se había pedido.
CREATE TABLE auth.desafios (
	token_hash text PRIMARY KEY,
	usuario_id uuid NOT NULL REFERENCES auth.usuarios (id) ON DELETE CASCADE,
	vida interval NOT NULL,
	vence_en timestamptz NOT NULL,
	intentos integer NOT NULL DEFAULT 0,
	ip inet,
	agente text CHECK (length(agente) <= 200)
);
CREATE INDEX desafios_vence ON auth.desafios (vence_en);
GRANT SELECT, INSERT, UPDATE (intentos), DELETE ON auth.desafios TO mostrador_app;

-- +goose Down
DROP TABLE auth.desafios;
ALTER TABLE auth.usuarios DROP COLUMN totp_secreto, DROP COLUMN totp_pendiente, DROP COLUMN totp_activo_en,
	DROP COLUMN codigos_recuperacion;
