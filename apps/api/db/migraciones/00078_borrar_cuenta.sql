-- +goose Up
-- Borrar la cuenta desde la app (Google Play y App Store lo exigen; decisión del 12/11/2026): la cuenta
-- se desactiva al instante (no entra más, se cierran sus sesiones y las tiendas donde era la única
-- dueña) y a los 30 días el worker borra sus datos personales. Las ventas y los registros del negocio
-- quedan, sin datos de la persona. En esos 30 días soporte puede reactivarla.
ALTER TABLE auth.usuarios ADD COLUMN desactivada_en timestamptz;
GRANT UPDATE (desactivada_en, email) ON auth.usuarios TO mostrador_app;

-- +goose Down
REVOKE UPDATE (desactivada_en, email) ON auth.usuarios FROM mostrador_app;
ALTER TABLE auth.usuarios DROP COLUMN desactivada_en;
