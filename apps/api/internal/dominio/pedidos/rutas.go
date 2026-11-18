package pedidos

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

var slugEstado = map[string]string{
	"sin-pagar": "Sin pagar", "para-preparar": "Para preparar", "listo": "Listo para despachar",
	"en-camino": "En camino", "entregado": "Entregado", "cancelado": "Cancelado",
}

func lista(v string) []string {
	if v == "" {
		return nil
	}
	return strings.Split(v, ",")
}

// FiltrosDe lee los filtros de la query (los mismos de la URL del panel).
func FiltrosDe(q map[string][]string) Filtros {
	g := func(k string) string {
		if v := q[k]; len(v) > 0 {
			return v[0]
		}
		return ""
	}
	f := Filtros{Q: g("q"), Canal: lista(g("canal")), Pago: lista(g("pago")), Envio: lista(g("envio")),
		Fecha: g("fecha"), Orden: g("orden"), VIP: g("vip") == "true" || g("vip") == "1",
		Hornero: g("hornero") == "true" || g("hornero") == "1", Mayoris: g("mayorista") == "true" || g("mayorista") == "1"}
	if len(f.Q) > 100 {
		f.Q = f.Q[:100]
	}
	for _, e := range lista(g("estado")) {
		if s, ok := slugEstado[e]; ok {
			e = s
		}
		f.Estado = append(f.Estado, e)
	}
	f.Minimo, _ = strconv.ParseInt(g("minimo"), 10, 64)
	f.Desde, _ = strconv.Atoi(g("desde"))
	f.Cuantos, _ = strconv.Atoi(g("cuantos"))
	return f
}

func escribir(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(v)
}

// Rutas registra los endpoints de pedidos. tienda resuelve la tienda de la sesión.
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
	numeroDe := func(w http.ResponseWriter, r *http.Request) (int, bool) {
		n, err := strconv.Atoi(r.PathValue("numero"))
		if err != nil || n < 1 {
			problema.Escribir(w, problema.NoEncontrado("Ese pedido no existe."))
			return 0, false
		}
		return n, true
	}

	rutasNuevo(mux, pool, con)
	mux.HandleFunc("GET /v1/panel/pedidos", con(func(w http.ResponseWriter, r *http.Request, id string) {
		l, err := Listar(r.Context(), pool, id, FiltrosDe(r.URL.Query()), time.Now())
		if err != nil {
			problema.Escribir(w, problema.Interno())
			return
		}
		escribir(w, l)
	}))
	mux.HandleFunc("GET /v1/panel/pedidos/{numero}", con(func(w http.ResponseWriter, r *http.Request, id string) {
		n, ok := numeroDe(w, r)
		if !ok {
			return
		}
		d, err := Obtener(r.Context(), pool, id, n)
		switch {
		case errors.Is(err, ErrNoExiste):
			problema.Escribir(w, problema.NoEncontrado("Ese pedido no existe."))
		case err != nil:
			problema.Escribir(w, problema.Interno())
		default:
			escribir(w, d)
		}
	}))
	mux.HandleFunc("POST /v1/panel/pedidos/{numero}/estado", con(func(w http.ResponseWriter, r *http.Request, id string) {
		n, ok := numeroDe(w, r)
		if !ok {
			return
		}
		var c struct {
			Estado string `json:"estado"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<10)).Decode(&c); err != nil {
			problema.Escribir(w, problema.Invalido("Mandá el estado nuevo."))
			return
		}
		switch err := CambiarEstado(r.Context(), pool, id, n, c.Estado); {
		case err == nil:
			w.WriteHeader(http.StatusNoContent)
		case errors.Is(err, ErrNoExiste):
			problema.Escribir(w, problema.NoEncontrado("Ese pedido no existe."))
		case errors.Is(err, ErrTransicion):
			problema.Escribir(w, problema.Conflicto("transicion", "El pedido ya cambió de etapa."))
		case errors.Is(err, ErrSinStock):
			problema.Escribir(w, problema.Conflicto("sin_stock", "Para reabrirlo no alcanza el stock de algún producto."))
		default:
			problema.Escribir(w, problema.Interno())
		}
	}))
	// El recordatorio sale por WhatsApp: hasta conectar WhatsApp Cloud no hay cómo mandarlo.
	mux.HandleFunc("POST /v1/panel/pedidos/{numero}/recordatorio", con(func(w http.ResponseWriter, r *http.Request, _ string) {
		if _, ok := numeroDe(w, r); !ok {
			return
		}
		problema.Escribir(w, problema.Conflicto("sin_whatsapp", "Conectá WhatsApp para mandar recordatorios de pago."))
	}))
	mux.HandleFunc("POST /v1/panel/pedidos/lote", con(func(w http.ResponseWriter, r *http.Request, id string) {
		var c struct {
			Numeros []int  `json:"numeros"`
			Accion  string `json:"accion"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&c); err != nil || len(c.Numeros) == 0 || len(c.Numeros) > 500 {
			problema.Escribir(w, problema.Invalido("Elegí pedidos y una acción."))
			return
		}
		cambios, err := EnLote(r.Context(), pool, id, c.Numeros, c.Accion)
		if err != nil {
			problema.Escribir(w, problema.Interno())
			return
		}
		escribir(w, map[string]any{"hechos": len(cambios), "cambios": cambios})
	}))
}
