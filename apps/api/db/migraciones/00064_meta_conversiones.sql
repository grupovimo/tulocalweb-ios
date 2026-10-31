-- +goose Up
-- API de conversiones de Meta (§47): el token de la tienda, cifrado con su clave, y cómo salió el
-- último envío (el panel lo muestra; el token nunca vuelve).
CREATE TABLE meta_conversiones (
	tienda_id uuid PRIMARY KEY REFERENCES tiendas (id) ON DELETE CASCADE,
	token bytea NOT NULL,
	codigo_prueba text CHECK (codigo_prueba ~ '^TEST[0-9A-Za-z]{1,20}$'),
	ultimo_envio timestamptz,
	ultimo_error text CHECK (length(ultimo_error) <= 300),
	actualizada_en timestamptz NOT NULL DEFAULT now()
);
SELECT app.aplicar_rls('meta_conversiones');
GRANT SELECT, INSERT, UPDATE, DELETE ON meta_conversiones TO mostrador_app;

-- +goose Down
DROP TABLE meta_conversiones;
