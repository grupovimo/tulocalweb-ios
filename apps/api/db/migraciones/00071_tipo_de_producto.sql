-- +goose Up
-- El tipo de producto y sus datos (ropa, calzado, comida y bebida, belleza, hogar y objetos,
-- tecnología u otro). El editor pregunta primero qué es y según eso pide talle y color, contenido,
-- peso y medidas, marca y modelo… NULL: todavía no se eligió (los productos de antes).
-- atributos: clave → texto (peso en gramos y medidas en centímetros van como número en texto).
ALTER TABLE productos
	ADD COLUMN tipo text CHECK (tipo IN ('ropa', 'calzado', 'alimentos', 'belleza', 'hogar', 'tecnologia', 'otro')),
	ADD COLUMN atributos jsonb NOT NULL DEFAULT '{}' CHECK (jsonb_typeof(atributos) = 'object');
GRANT UPDATE (tipo, atributos) ON productos TO mostrador_app;

-- +goose Down
REVOKE UPDATE (tipo, atributos) ON productos FROM mostrador_app;
ALTER TABLE productos DROP COLUMN atributos, DROP COLUMN tipo;
