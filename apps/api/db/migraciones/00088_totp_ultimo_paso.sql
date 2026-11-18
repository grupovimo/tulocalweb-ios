-- +goose Up
-- El último paso de 30 s con el que entró cada persona con su app autenticadora: un código ya usado
-- (o uno anterior) no vuelve a valer. Antes, un código interceptado abría otra sesión durante su
-- minuto y medio de validez.
ALTER TABLE auth.usuarios ADD COLUMN totp_ultimo_paso bigint;
GRANT UPDATE (totp_ultimo_paso) ON auth.usuarios TO mostrador_app;

-- +goose Down
REVOKE UPDATE (totp_ultimo_paso) ON auth.usuarios FROM mostrador_app;
ALTER TABLE auth.usuarios DROP COLUMN totp_ultimo_paso;
