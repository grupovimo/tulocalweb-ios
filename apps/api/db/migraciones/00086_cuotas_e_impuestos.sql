-- +goose Up
-- Cuotas sin interés y precio sin impuestos nacionales (docs/especificaciones/tienda-nivel-tiendanube.md).
--
-- Cuotas: prendidas o apagadas viven en automaticas->>'cuotas' (la misma promoción automática de
-- Cupones, como el envío gratis y el mayorista), así hay un solo interruptor; acá, cuántas (3 a 12)
-- y desde qué monto en pesos. Sólo se ofrecen con Mercado Pago conectado (dominio/cobros). Hasta hoy
-- ese interruptor no hacía nada (la vidriera mandaba 0): se apaga en todas, porque ahora limita las
-- cuotas del checkout y promete en la tienda, y eso lo tiene que elegir cada una.
ALTER TABLE tiendas
	ADD COLUMN cuotas_sin_interes smallint NOT NULL DEFAULT 6 CHECK (cuotas_sin_interes IN (3, 6, 9, 12)),
	ADD COLUMN cuotas_desde bigint NOT NULL DEFAULT 30000 CHECK (cuotas_desde BETWEEN 0 AND 100000000),
	-- Mostrar «Precio sin impuestos nacionales» (Ley 27.743). NULL: lo que corresponde a la condición
	-- fiscal (prendido para responsables inscriptos, apagado para monotributo).
	ADD COLUMN impuestos_mostrar boolean;
GRANT UPDATE (cuotas_sin_interes, cuotas_desde, impuestos_mostrar) ON tiendas TO mostrador_app;
UPDATE tiendas SET automaticas = automaticas || '{"cuotas": false}' WHERE (automaticas->>'cuotas')::boolean;

-- La alícuota de IVA de cada producto, para el precio sin impuestos nacionales: 21 por defecto.
ALTER TABLE productos ADD COLUMN iva numeric(3, 1) NOT NULL DEFAULT 21 CHECK (iva IN (0, 10.5, 21, 27));
GRANT UPDATE (iva) ON productos TO mostrador_app;

-- +goose Down
ALTER TABLE productos DROP COLUMN iva;
ALTER TABLE tiendas DROP COLUMN cuotas_sin_interes, DROP COLUMN cuotas_desde, DROP COLUMN impuestos_mostrar;
