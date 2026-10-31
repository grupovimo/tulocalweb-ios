package configuracion

import (
	"context"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/mayorista"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

// GuardarMayorista guarda Configuración › Venta mayorista (§49).
func GuardarMayorista(ctx context.Context, pool tx, tiendaID string, c mayorista.Config) error {
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error { return mayorista.Guardar(ctx, t, tiendaID, c) })
}

func rutasMayorista(mux *http.ServeMux, pool tx, con func(func(http.ResponseWriter, *http.Request, string, Quien)) http.HandlerFunc,
	errorDe func(http.ResponseWriter, error), leer func(http.ResponseWriter, *http.Request, any) bool) {
	mux.HandleFunc("PUT /v1/panel/configuracion/mayorista", con(func(w http.ResponseWriter, r *http.Request, id string, _ Quien) {
		var c mayorista.Config
		if !leer(w, r, &c) {
			return
		}
		switch err := GuardarMayorista(r.Context(), pool, id, c); {
		case errors.Is(err, mayorista.ErrSinMinimo):
			problema.Escribir(w, problema.Invalido("Poné un mínimo de unidades, de monto o los dos."))
		case errors.Is(err, mayorista.ErrInvalido):
			problema.Escribir(w, problema.Invalido("Revisá el mínimo: de 2 a 10.000 unidades y hasta $ 100.000.000."))
		default:
			errorDe(w, err)
		}
	}))
}
