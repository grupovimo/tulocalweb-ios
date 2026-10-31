package suscripciones

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

// Config de las rutas.
type Config struct {
	// Tienda de la sesión ("" sin sesión).
	Tienda func(*http.Request) string
	// MP arma y cambia los débitos. nil = sin la integración (no se ofrecen planes).
	MP Cobrador
	// Sitio público, para la vuelta de Mercado Pago.
	Sitio string
	Log   *slog.Logger
	// IP de quien se suscribe (la que manda el BFF con su clave) y el contador: tope de altas por IP
	// y por email. nil = sin tope (tests).
	IP        func(*http.Request) string
	Permitido func(ctx context.Context, clave string, max int, ventana time.Duration) (bool, error)
}

func escribir(w http.ResponseWriter, estado int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(estado)
	_ = json.NewEncoder(w).Encode(v)
}

func leer(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(v); err != nil {
		problema.Escribir(w, problema.Invalido("No entendimos el pedido."))
		return false
	}
	return true
}

// Rutas del panel (/v1/panel/suscripciones) y de la tienda (/v1/publico/tiendas/{slug}/suscripciones).
func Rutas(mux *http.ServeMux, pool tx, cfg Config) {
	log := cfg.Log
	if log == nil {
		log = slog.Default()
	}
	errorDe := func(w http.ResponseWriter, err error) {
		switch {
		case errors.Is(err, ErrNoExiste):
			problema.Escribir(w, problema.NoEncontrado("No lo encontramos."))
		case errors.Is(err, ErrInvalido):
			problema.Escribir(w, problema.Invalido("Revisá los datos."))
		case errors.Is(err, ErrSinCobros):
			problema.Escribir(w, problema.Conflicto("sin_cobros", "La tienda todavía no conectó Mercado Pago."))
		case errors.Is(err, ErrCupo):
			problema.Escribir(w, problema.Conflicto("cupo", "Este plan no tiene más lugar por ahora."))
		case errors.Is(err, ErrEnUso):
			problema.Escribir(w, problema.Conflicto("en_uso", "Este plan tiene suscriptores: pausalo en vez de borrarlo."))
		case errors.Is(err, ErrEstado):
			problema.Escribir(w, problema.Conflicto("estado", "La suscripción ya no está en ese estado. Recargá la página."))
		case errors.Is(err, ErrMercadoPago):
			log.Warn("suscripciones: Mercado Pago no respondió", "error", err)
			problema.Escribir(w, &problema.Problema{Estado: http.StatusBadGateway, Codigo: "mercadopago",
				Titulo: "Mercado Pago no respondió", Detalle: "Probá de nuevo en un rato."})
		default:
			log.Error("suscripciones", "error", err)
			problema.Escribir(w, problema.Interno())
		}
	}
	con := func(h func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			id := ""
			if cfg.Tienda != nil {
				id = cfg.Tienda(r)
			}
			if id == "" {
				problema.Escribir(w, problema.SinSesion())
				return
			}
			h(w, r, id)
		}
	}
	responder := func(w http.ResponseWriter, estado int, v any, err error) {
		switch {
		case err != nil:
			errorDe(w, err)
		case v == nil:
			w.WriteHeader(http.StatusNoContent)
		default:
			escribir(w, estado, v)
		}
	}

	// Panel.
	mux.HandleFunc("GET /v1/panel/suscripciones", con(func(w http.ResponseWriter, r *http.Request, id string) {
		p, err := PanelDe(r.Context(), pool, id)
		if cfg.MP == nil {
			p.MercadoPago = false
		}
		responder(w, http.StatusOK, p, err)
	}))
	guardar := func(nuevo bool) http.HandlerFunc {
		return con(func(w http.ResponseWriter, r *http.Request, id string) {
			p := Plan{Activo: true, Envio: true, Retiro: true}
			if !leer(w, r, &p) {
				return
			}
			estado, pid := http.StatusOK, r.PathValue("id")
			if nuevo {
				estado, pid = http.StatusCreated, ""
			}
			out, err := GuardarPlan(r.Context(), pool, id, pid, p)
			responder(w, estado, out, err)
		})
	}
	mux.HandleFunc("POST /v1/panel/suscripciones/planes", guardar(true))
	mux.HandleFunc("PUT /v1/panel/suscripciones/planes/{id}", guardar(false))
	mux.HandleFunc("DELETE /v1/panel/suscripciones/planes/{id}", con(func(w http.ResponseWriter, r *http.Request, id string) {
		responder(w, 0, nil, BorrarPlan(r.Context(), pool, id, r.PathValue("id")))
	}))
	mux.HandleFunc("POST /v1/panel/suscripciones/{id}/cancelar", con(func(w http.ResponseWriter, r *http.Request, id string) {
		s, err := CancelarDesdePanel(r.Context(), pool, cfg.MP, id, r.PathValue("id"))
		responder(w, http.StatusOK, s, err)
	}))

	// Público (sin sesión: lo pide el BFF de /t/{slug}).
	mux.HandleFunc("POST /v1/publico/tiendas/{slug}/suscripciones", func(w http.ResponseWriter, r *http.Request) {
		var a Alta
		if !leer(w, r, &a) {
			return
		}
		if cfg.Permitido != nil && cfg.IP != nil {
			for _, l := range []struct {
				clave string
				max   int
			}{{"suscripcion:ip:" + cfg.IP(r), 10}, {"suscripcion:email:" + r.PathValue("slug") + ":" + strings.ToLower(strings.TrimSpace(a.Email)), 5}} {
				ok, err := cfg.Permitido(r.Context(), l.clave, l.max, time.Hour)
				if err != nil {
					errorDe(w, err)
					return
				}
				if !ok {
					w.Header().Set("Retry-After", "3600")
					problema.Escribir(w, &problema.Problema{Estado: http.StatusTooManyRequests, Codigo: "demasiados",
						Titulo: "Demasiados intentos", Detalle: "Probá de nuevo en una hora o escribile a la tienda."})
					return
				}
			}
		}
		res, err := Suscribir(r.Context(), pool, cfg.MP, cfg.Sitio, r.PathValue("slug"), a)
		responder(w, http.StatusCreated, res, err)
	})
	mux.HandleFunc("GET /v1/publico/tiendas/{slug}/suscripciones/{codigo}", func(w http.ResponseWriter, r *http.Request) {
		s, err := Ver(r.Context(), pool, r.PathValue("slug"), r.PathValue("codigo"))
		responder(w, http.StatusOK, s, err)
	})
	mux.HandleFunc("POST /v1/publico/tiendas/{slug}/suscripciones/{codigo}/{accion}", func(w http.ResponseWriter, r *http.Request) {
		s, err := CambiarPublica(r.Context(), pool, cfg.MP, r.PathValue("slug"), r.PathValue("codigo"), r.PathValue("accion"))
		responder(w, http.StatusOK, s, err)
	})
}
