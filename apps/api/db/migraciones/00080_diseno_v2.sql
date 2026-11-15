-- +goose Up
-- Tienda en bloques (docs/especificaciones/tienda-bloques.md): `mostradorctl diseno-v2` pasa los
-- diseños v1 a v2 con mitienda.Migrar (la misma función que migra al leer). Para recorrer todas las
-- tiendas en una transacción, el migrador las lee con app.diseno_v2 = 'si'; cada UPDATE va con la
-- tienda fijada (la política de siempre).
CREATE POLICY diseno_v2 ON tiendas FOR SELECT TO mostrador_migrador
	USING (current_setting('app.diseno_v2', true) = 'si');

-- +goose Down
DROP POLICY diseno_v2 ON tiendas;
