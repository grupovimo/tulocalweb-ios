package newsletter

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

func escribir(w http.ResponseWriter, estado int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(estado)
	_ = json.NewEncoder(w).Encode(v)
}

// Rutas del panel (Crecer › Newsletter; permiso Tienda en roles.Rutas).
func Rutas(mux *http.ServeMux, pool tx, tienda func(*http.Request) string) {
	con := func(h func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			id := tienda(r)
			if id == "" {
				problema.Escribir(w, problema.SinSesion())
				return
			}
			h(w, r, id)
		}
	}
	mux.HandleFunc("GET /v1/panel/newsletter", con(func(w http.ResponseWriter, r *http.Request, id string) {
		p, err := Leer(r.Context(), pool, id, time.Now())
		if err != nil {
			problema.Escribir(w, problema.Interno())
			return
		}
		escribir(w, http.StatusOK, p)
	}))
	mux.HandleFunc("DELETE /v1/panel/newsletter/suscriptores/{id}", con(func(w http.ResponseWriter, r *http.Request, id string) {
		switch err := Borrar(r.Context(), pool, id, r.PathValue("id")); {
		case errors.Is(err, ErrNoExiste):
			problema.Escribir(w, problema.NoEncontrado("Ese suscripto ya no está en la lista."))
		case err != nil:
			problema.Escribir(w, problema.Interno())
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	mux.HandleFunc("POST /v1/panel/newsletter/envios", con(func(w http.ResponseWriter, r *http.Request, id string) {
		var n Novedad
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 32<<10)).Decode(&n); err != nil {
			problema.Escribir(w, problema.Invalido("El pedido no es un JSON válido."))
			return
		}
		e, err := Mandar(r.Context(), pool, id, n, time.Now())
		var c ErrCampo
		switch {
		case errors.As(err, &c):
			problema.Escribir(w, problema.Invalido(c.Mensaje, problema.Campo{Ruta: c.Campo, Mensaje: c.Mensaje}))
		case errors.Is(err, ErrSinDestinatarios):
			problema.Escribir(w, problema.Conflicto("sin_suscriptos", "Todavía no hay suscriptos confirmados: la novedad no tendría a quién llegar."))
		case err != nil:
			problema.Escribir(w, problema.Interno())
		default:
			escribir(w, http.StatusCreated, e)
		}
	}))
}
