package promos

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/limite"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

func escribir(w http.ResponseWriter, estado int, v any) {
	w.Header().Set("Content-Type", "application/json")
	if w.Header().Get("Cache-Control") == "" {
		w.Header().Set("Cache-Control", "no-store")
	}
	w.WriteHeader(estado)
	_ = json.NewEncoder(w).Encode(v)
}

func demasiados(w http.ResponseWriter) {
	w.Header().Set("Retry-After", "3600")
	problema.Escribir(w, &problema.Problema{Estado: http.StatusTooManyRequests, Codigo: "demasiados_envios",
		Titulo: "Esperá un rato", Detalle: "Recibimos varios pedidos seguidos. Probá de nuevo en un rato."})
}

// Rutas: las públicas de la tienda (las pide el BFF de /t/{slug}/promos/…) y los números del panel.
// tienda resuelve la tienda de la sesión; ip, la de quien pide (para los topes).
func Rutas(mux *http.ServeMux, pool tx, tienda func(*http.Request) string, ip func(*http.Request) string) {
	// tope: por IP y tienda, y por tienda (que nadie llene la newsletter ni los cupones de otro).
	tope := func(r *http.Request, que string, porIP, porTienda int) bool {
		c, ok := pool.(limite.Consultor)
		if !ok {
			return true
		}
		slug := r.PathValue("slug")
		claves := []struct {
			clave string
			max   int
		}{{"promos:" + que + ":tienda:" + slug, porTienda}}
		if ip != nil {
			claves = append(claves, struct {
				clave string
				max   int
			}{"promos:" + que + ":ip:" + slug + ":" + ip(r), porIP})
		}
		for _, l := range claves {
			if ok, err := limite.Permitido(r.Context(), c, l.clave, l.max, time.Hour); err == nil && !ok {
				return false
			}
		}
		return true
	}
	participar := func(contenido string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if !tope(r, "participar", 10, 600) {
				demasiados(w)
				return
			}
			var p Pedido
			if json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&p) != nil {
				problema.Escribir(w, problema.Invalido("Revisá el email."))
				return
			}
			res, err := Participar(r.Context(), pool, r.PathValue("slug"), contenido, p, time.Now())
			switch {
			case errors.Is(err, ErrInvalido):
				if !p.Consentimiento {
					problema.Escribir(w, problema.Invalido("Marcá la casilla para aceptar que te escriban.",
						problema.Campo{Ruta: "consentimiento", Mensaje: "Marcá la casilla para seguir."}))
					return
				}
				problema.Escribir(w, problema.Invalido("Revisá el email.", problema.Campo{Ruta: "email", Mensaje: "Revisá el email."}))
			case errors.Is(err, ErrNoExiste):
				problema.Escribir(w, problema.NoEncontrado("Esta promo no existe."))
			case errors.Is(err, ErrTerminada):
				problema.Escribir(w, problema.Conflicto("terminada", "Esta promo ya terminó. ¡Gracias por tu interés!"))
			case err != nil:
				problema.Escribir(w, problema.Interno())
			default:
				escribir(w, http.StatusOK, res)
			}
		}
	}
	mux.HandleFunc("POST /v1/publico/tiendas/{slug}/promos/captar", participar("captacion"))
	mux.HandleFunc("POST /v1/publico/tiendas/{slug}/promos/ruleta", participar("ruleta"))
	// Eventos en lote (navigator.sendBeacon manda text/plain: no se mira el Content-Type).
	mux.HandleFunc("POST /v1/publico/tiendas/{slug}/promos/eventos", func(w http.ResponseWriter, r *http.Request) {
		if !tope(r, "eventos", 120, 20000) {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		var b struct {
			Eventos []Evento `json:"eventos"`
		}
		crudo, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 8<<10))
		if err != nil || json.Unmarshal(crudo, &b) != nil || len(b.Eventos) > MaxEventos {
			problema.Escribir(w, problema.Invalido("Mandá hasta 50 eventos."))
			return
		}
		err = Registrar(r.Context(), pool, r.PathValue("slug"), b.Eventos, time.Now())
		switch {
		case errors.Is(err, ErrNoExiste):
			problema.Escribir(w, problema.NoEncontrado("Esta tienda no existe."))
		case err != nil:
			problema.Escribir(w, problema.Interno())
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	})
	mux.HandleFunc("GET /v1/publico/tiendas/{slug}/promos/recientes", func(w http.ResponseWriter, r *http.Request) {
		c, err := Recientes(r.Context(), pool, r.PathValue("slug"), time.Now())
		switch {
		case errors.Is(err, ErrNoExiste):
			problema.Escribir(w, problema.NoEncontrado("Esta tienda no existe."))
		case err != nil:
			problema.Escribir(w, problema.Interno())
		default:
			w.Header().Set("Cache-Control", "public, max-age=300")
			escribir(w, http.StatusOK, map[string]any{"compras": c})
		}
	})
	mux.HandleFunc("GET /v1/panel/promos/metricas", func(w http.ResponseWriter, r *http.Request) {
		id := tienda(r)
		if id == "" {
			problema.Escribir(w, problema.SinSesion())
			return
		}
		desde, hasta, ok := Rango(r.URL.Query().Get("desde"), r.URL.Query().Get("hasta"), time.Now())
		if !ok {
			problema.Escribir(w, problema.Invalido("Elegí un rango de fechas de hasta un año (AAAA-MM-DD)."))
			return
		}
		m, err := LeerMetricas(r.Context(), pool, id, desde, hasta)
		if err != nil {
			problema.Escribir(w, problema.Interno())
			return
		}
		escribir(w, http.StatusOK, m)
	})
}
