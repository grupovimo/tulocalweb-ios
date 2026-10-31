package lanzamientos

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/limite"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

func escribir(w http.ResponseWriter, estado int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(estado)
	_ = json.NewEncoder(w).Encode(v)
}

func errorDe(w http.ResponseWriter, err error) {
	var inv ErrInvalido
	switch {
	case errors.As(err, &inv):
		problema.Escribir(w, problema.Invalido(inv.Motivo))
	case errors.Is(err, ErrNoExiste):
		problema.Escribir(w, problema.NoEncontrado("Ese producto no existe."))
	case errors.Is(err, ErrYaSalio):
		problema.Escribir(w, problema.Conflicto("ya_salio", "Este producto ya está a la venta."))
	case errors.Is(err, ErrLleno):
		problema.Escribir(w, problema.Conflicto("lleno", "No podemos anotar más avisos para este producto."))
	case errors.Is(err, pedidos.ErrSinSena):
		problema.Escribir(w, problema.Conflicto("sin_sena", "Este pedido no tiene una seña pagada."))
	default:
		problema.Escribir(w, problema.Interno())
	}
}

// Rutas: la configuración del panel, «saldado» de los pedidos con seña y el aviso público. tienda
// resuelve la tienda de la sesión; ip, la de quien pide (para el tope del aviso).
func Rutas(mux *http.ServeMux, pool tx, tienda func(*http.Request) string, ip func(*http.Request) string) {
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
	mux.HandleFunc("GET /v1/panel/productos/{id}/lanzamiento", con(func(w http.ResponseWriter, r *http.Request, id string) {
		e, err := Leer(r.Context(), pool, id, r.PathValue("id"))
		if err != nil {
			errorDe(w, err)
			return
		}
		escribir(w, http.StatusOK, e)
	}))
	mux.HandleFunc("PUT /v1/panel/productos/{id}/lanzamiento", con(func(w http.ResponseWriter, r *http.Request, id string) {
		var c Config
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&c) != nil {
			problema.Escribir(w, problema.Invalido("Revisá el lanzamiento y la oferta."))
			return
		}
		e, err := Guardar(r.Context(), pool, id, r.PathValue("id"), c, time.Now())
		if err != nil {
			errorDe(w, err)
			return
		}
		escribir(w, http.StatusOK, e)
	}))
	saldar := func(saldado bool) http.HandlerFunc {
		return con(func(w http.ResponseWriter, r *http.Request, id string) {
			n, err := strconv.Atoi(r.PathValue("numero"))
			if err != nil || n < 1 {
				problema.Escribir(w, problema.NoEncontrado("Ese pedido no existe."))
				return
			}
			if err := pedidos.Saldar(r.Context(), pool, id, n, saldado); err != nil {
				errorDe(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})
	}
	mux.HandleFunc("POST /v1/panel/pedidos/{numero}/saldado", saldar(true))
	mux.HandleFunc("DELETE /v1/panel/pedidos/{numero}/saldado", saldar(false))
	// «Avisame cuando salga»: diez por hora por IP (como reseñas y preguntas) y un tope por producto.
	mux.HandleFunc("POST /v1/publico/tiendas/{slug}/productos/{id}/aviso", func(w http.ResponseWriter, r *http.Request) {
		if c, ok := pool.(limite.Consultor); ok && ip != nil {
			if ok, err := limite.Permitido(r.Context(), c, "tienda:aviso:"+ip(r), 10, time.Hour); err == nil && !ok {
				problema.Escribir(w, &problema.Problema{Estado: http.StatusTooManyRequests, Codigo: "demasiados",
					Titulo: "Demasiados envíos", Detalle: "Probá de nuevo en un rato."})
				return
			}
		}
		var b struct {
			Email string `json:"email"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 2<<10)).Decode(&b) != nil {
			errorDe(w, ErrInvalido{"Revisá el email."})
			return
		}
		if err := Avisar(r.Context(), pool, r.PathValue("slug"), r.PathValue("id"), b.Email, time.Now()); err != nil {
			errorDe(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}
