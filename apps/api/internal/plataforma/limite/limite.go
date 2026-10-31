// Package limite cuenta intentos por clave en una ventana de tiempo (auth.intentos), con un solo
// upsert atómico, como el ratelimit de Caja. Lo usan los formularios públicos: ingreso,
// recuperación, arrepentimiento.
package limite

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

// Consultor es lo mínimo que hace falta: un pool o una transacción.
type Consultor interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Permitido suma un intento a la clave y dice si sigue dentro del máximo de la ventana.
func Permitido(ctx context.Context, q Consultor, clave string, max int, ventana time.Duration) (bool, error) {
	n, err := Sumar(ctx, q, clave, ventana)
	return n <= max, err
}

// Sumar suma un intento a la clave y devuelve cuántos van en la ventana (con este).
func Sumar(ctx context.Context, q Consultor, clave string, ventana time.Duration) (int, error) {
	var n int
	err := q.QueryRow(ctx, `
		INSERT INTO auth.intentos (clave, cantidad, desde) VALUES ($1, 1, now())
		ON CONFLICT (clave) DO UPDATE SET
			cantidad = CASE WHEN auth.intentos.desde < now() - $2::interval THEN 1 ELSE auth.intentos.cantidad + 1 END,
			desde = CASE WHEN auth.intentos.desde < now() - $2::interval THEN now() ELSE auth.intentos.desde END
		RETURNING cantidad`, clave, ventana.String()).Scan(&n)
	return n, err
}
