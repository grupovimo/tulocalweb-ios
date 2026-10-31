package conversiones

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

func escribirError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalido):
		problema.Escribir(w, problema.Invalido("Revisá el token y el código de prueba."))
	case errors.Is(err, ErrSinPixel):
		problema.Escribir(w, problema.Conflicto("sin_pixel", "Primero guardá el ID de tu píxel."))
	case errors.Is(err, ErrTokenNoSirve):
		problema.Escribir(w, problema.Conflicto("token", "Meta no aceptó ese token para tu píxel. Generalo de nuevo desde el Administrador de eventos."))
	case errors.Is(err, ErrMetaCaida):
		problema.Escribir(w, problema.Conflicto("meta", "No pudimos hablar con Meta. Probá en un rato."))
	case errors.Is(err, ErrSinCifrado):
		problema.Escribir(w, problema.Conflicto("sin_cifrado", "No podemos guardar el token en este momento. Probá más tarde."))
	default:
		slog.Error("conversiones", "error", err)
		problema.Escribir(w, problema.Interno())
	}
}

// Rutas del panel (Integraciones › Meta Ads). El token nunca vuelve al navegador.
func (s *Servicio) Rutas(mux *http.ServeMux) {
	panel := func(h func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			d, ok := cuentas.SesionDe(r)
			if !ok || d.TiendaID == "" || d.IA != "" {
				problema.Escribir(w, problema.SinSesion())
				return
			}
			h(w, r, d.TiendaID)
		}
	}
	mux.HandleFunc("GET /v1/panel/integraciones/medicion/conversiones", panel(func(w http.ResponseWriter, r *http.Request, tienda string) {
		e, err := s.Leer(r.Context(), tienda)
		if err != nil {
			escribirError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(e)
	}))
	mux.HandleFunc("PUT /v1/panel/integraciones/medicion/conversiones", panel(func(w http.ResponseWriter, r *http.Request, tienda string) {
		var c Conexion
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&c) != nil {
			escribirError(w, ErrInvalido)
			return
		}
		if err := s.Guardar(r.Context(), tienda, c); err != nil {
			escribirError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	mux.HandleFunc("DELETE /v1/panel/integraciones/medicion/conversiones", panel(func(w http.ResponseWriter, r *http.Request, tienda string) {
		if err := s.Borrar(r.Context(), tienda); err != nil {
			escribirError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
}
