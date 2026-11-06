-- +goose Up
-- Tiendas de demostración (dominio/demo: Tienda Lumbre, Café Bruma, Almacén Verde). Su tienda online
-- se ve como la vitrina de una plantilla: todo a la vista, pero no se compra, no se mide ni se guarda
-- nada de quien la visita. El panel funciona como el de cualquier tienda.
ALTER TABLE tiendas ADD COLUMN demo boolean NOT NULL DEFAULT false;
UPDATE tiendas SET demo = true WHERE slug IN ('lumbre', 'cafebruma', 'almacenverde');

-- +goose Down
ALTER TABLE tiendas DROP COLUMN demo;
