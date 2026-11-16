package configuracion

import (
	"context"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cobros"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

// Cobros es Configuracion.cobros: lo que eligió la tienda (cobros.Config) y si tiene Mercado Pago
// conectado (sin Mercado Pago la tienda no ofrece cuotas, aunque estén elegidas).
type Cobros struct {
	cobros.Config
	MercadoPago bool `json:"mercadoPago"`
}

func leerCobros(ctx context.Context, t pgx.Tx, tiendaID string) (Cobros, error) {
	c, err := cobros.Leer(ctx, t, tiendaID)
	if err != nil {
		return Cobros{}, err
	}
	out := Cobros{Config: c}
	err = t.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM mp_conexiones WHERE tienda_id = $1 AND NOT reautorizar)`, tiendaID).
		Scan(&out.MercadoPago)
	return out, err
}

// GuardarCobros guarda Configuración › Cobros: cuotas sin interés y precio sin impuestos nacionales.
func GuardarCobros(ctx context.Context, pool tx, tiendaID string, c cobros.Config) error {
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error { return cobros.Guardar(ctx, t, tiendaID, c) })
}

func rutasCobros(mux *http.ServeMux, pool tx, con func(func(http.ResponseWriter, *http.Request, string, Quien)) http.HandlerFunc,
	errorDe func(http.ResponseWriter, error), leer func(http.ResponseWriter, *http.Request, any) bool) {
	mux.HandleFunc("PUT /v1/panel/configuracion/cobros", con(func(w http.ResponseWriter, r *http.Request, id string, _ Quien) {
		var c cobros.Config
		if !leer(w, r, &c) {
			return
		}
		if err := GuardarCobros(r.Context(), pool, id, c); errors.Is(err, cobros.ErrInvalido) {
			problema.Escribir(w, problema.Invalido("Revisá las cuotas (ninguna, 3, 6, 9 o 12) y el monto: hasta $ 100.000.000."))
		} else {
			errorDe(w, err)
		}
	}))
}
