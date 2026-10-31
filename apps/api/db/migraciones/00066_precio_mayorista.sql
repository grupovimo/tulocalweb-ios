-- +goose Up
-- Precios mayoristas (docs/decisiones.md §49). Cada producto (y cada variante) puede tener un precio
-- mayorista; NULL: no tiene. En la variante, NULL usa el del producto (como el precio). La tienda
-- define el mínimo del carrito para pasar a esos precios: unidades y/o monto (en pesos, contra los
-- precios minoristas). Prendido o apagado vive en automaticas->>'mayorista' (la misma promoción
-- automática de Cupones), así hay un solo interruptor. El pedido ya tenía su marca (pedidos.mayorista).
ALTER TABLE productos ADD COLUMN precio_mayorista bigint CHECK (precio_mayorista > 0);
GRANT UPDATE (precio_mayorista) ON productos TO mostrador_app;
ALTER TABLE variantes ADD COLUMN precio_mayorista bigint CHECK (precio_mayorista > 0);
GRANT UPDATE (precio_mayorista) ON variantes TO mostrador_app;
ALTER TABLE tiendas
	ADD COLUMN mayorista_unidades int DEFAULT 12 CHECK (mayorista_unidades BETWEEN 2 AND 10000),
	ADD COLUMN mayorista_monto bigint CHECK (mayorista_monto BETWEEN 1 AND 100000000),
	ADD CONSTRAINT mayorista_con_minimo CHECK (mayorista_unidades IS NOT NULL OR mayorista_monto IS NOT NULL);
GRANT UPDATE (mayorista_unidades, mayorista_monto) ON tiendas TO mostrador_app;

-- +goose Down
REVOKE UPDATE (mayorista_unidades, mayorista_monto) ON tiendas FROM mostrador_app;
ALTER TABLE tiendas DROP CONSTRAINT mayorista_con_minimo, DROP COLUMN mayorista_monto, DROP COLUMN mayorista_unidades;
REVOKE UPDATE (precio_mayorista) ON variantes FROM mostrador_app;
ALTER TABLE variantes DROP COLUMN precio_mayorista;
REVOKE UPDATE (precio_mayorista) ON productos FROM mostrador_app;
ALTER TABLE productos DROP COLUMN precio_mayorista;
