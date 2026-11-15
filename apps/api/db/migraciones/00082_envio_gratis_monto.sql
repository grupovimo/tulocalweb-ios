-- +goose Up
-- Envío gratis con el monto de cada tienda (antes, fijo: cupones.EnvioGratisDesde = 40.000). El
-- interruptor sigue siendo automaticas->>'envio'; las que lo tenían prendido conservan 40.000.
ALTER TABLE tiendas ADD COLUMN envio_gratis_desde bigint NOT NULL DEFAULT 40000
	CHECK (envio_gratis_desde BETWEEN 1000 AND 100000000);
GRANT UPDATE (envio_gratis_desde) ON tiendas TO mostrador_app;

-- +goose Down
ALTER TABLE tiendas DROP COLUMN envio_gratis_desde;
