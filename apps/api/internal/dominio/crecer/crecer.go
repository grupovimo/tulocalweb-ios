// Package crecer: marketing, canales e informes (design/PanelCrecer.dc.html). Los números salen de
// los pedidos de la tienda. Las campañas necesitan WhatsApp Business conectado y los canales
// externos su integración: hasta entonces responden que falta conectarlos.
package crecer

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/visitas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

type tx interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
}

var (
	ErrNoExiste       = errors.New("no existe")
	ErrSinConectar    = errors.New("sin conectar")
	argentina         = time.FixedZone("ART", -3*3600)
	meses             = []string{"ene", "feb", "mar", "abr", "may", "jun", "jul", "ago", "sept", "oct", "nov", "dic"}
	mesesLargos       = []string{"enero", "febrero", "marzo", "abril", "mayo", "junio", "julio", "agosto", "septiembre", "octubre", "noviembre", "diciembre"}
	costoMensaje      = 120
	conectadosSiempre = []string{"web", "pos"}
)

// Canal es un elemento de components/schemas/Canales.
type Canal struct {
	ID        string  `json:"id"`
	Nombre    string  `json:"nombre"`
	Vendido   int64   `json:"vendido"`
	Pedidos   int     `json:"pedidos"`
	Comision  float64 `json:"comision"`
	Nota      string  `json:"nota"`
	Activo    bool    `json:"activo"`
	Conectado bool    `json:"conectado"`
}

var canales = []struct {
	id, nombre, nota string
	comision         float64
}{
	{"web", "Tienda online", "Sin comisión", 0},
	{"pos", "Local", "Punto de venta", 0},
	// La comisión de Mercado Libre depende de la categoría y del tipo de publicación: no se inventa un
	// porcentaje fijo (antes decía 14 % para todo).
	{"ml", "Mercado Libre", "La comisión la cobra Mercado Libre según la categoría", 0},
	{"gg", "Google Shopping", "Gratis · con el feed cargado en Merchant Center", 0},
	{"wa", "WhatsApp", "Todavía no se puede conectar WhatsApp Business", 0},
	{"ig", "Instagram", "Todavía no se puede conectar Instagram", 0},
}

// Canales: lo vendido en los últimos 30 días por canal.
func Canales(ctx context.Context, pool tx, tiendaID string, ahora time.Time) (map[string]any, error) {
	lista := []Canal{}
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var pausados map[string]bool
		var ml, gg bool
		if err := t.QueryRow(ctx, `SELECT canales_pausados, EXISTS (SELECT 1 FROM ml_conexiones WHERE tienda_id = $1),
			'gm' = ANY (feeds_cargados) FROM tiendas WHERE id = $1`, tiendaID).Scan(&pausados, &ml, &gg); err != nil {
			return err
		}
		filas, err := t.Query(ctx, `SELECT canal, coalesce(sum(total), 0), count(*) FROM pedidos
			WHERE tienda_id = $1 AND estado <> 'Cancelado' AND NOT (estado = 'Sin pagar' AND canal = 'web' AND pago = 'mp') AND creado_en > $2 GROUP BY canal`, tiendaID, ahora.Add(-30*24*time.Hour))
		if err != nil {
			return err
		}
		ventas := map[string][2]int64{}
		for filas.Next() {
			var c string
			var v, n int64
			if err := filas.Scan(&c, &v, &n); err != nil {
				return err
			}
			ventas[c] = [2]int64{v, n}
		}
		if err := filas.Err(); err != nil {
			return err
		}
		for _, c := range canales {
			// Conectado de verdad: la tienda y el local siempre; Mercado Libre con su cuenta conectada y
			// Google Shopping con el feed cargado (lo que la tienda marcó en Integraciones).
			conectado := slices.Contains(conectadosSiempre, c.id) || (c.id == "ml" && ml) || (c.id == "gg" && gg)
			lista = append(lista, Canal{ID: c.id, Nombre: c.nombre, Vendido: ventas[c.id][0], Pedidos: int(ventas[c.id][1]),
				Comision: c.comision, Nota: c.nota, Conectado: conectado, Activo: conectado && !pausados[c.id]})
		}
		return nil
	})
	return map[string]any{"mes": mesesLargos[ahora.In(argentina).Month()-1], "canales": lista}, err
}

// FijarCanal pausa o activa un canal conectado.
func FijarCanal(ctx context.Context, pool tx, tiendaID, id string, activo bool) error {
	if !slices.ContainsFunc(canales, func(c struct {
		id, nombre, nota string
		comision         float64
	}) bool {
		return c.id == id
	}) {
		return ErrNoExiste
	}
	if !slices.Contains(conectadosSiempre, id) {
		return ErrSinConectar
	}
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `UPDATE tiendas SET canales_pausados = canales_pausados || jsonb_build_object($2::text, $3::boolean)
			WHERE id = $1`, tiendaID, id, !activo)
		return err
	})
}

// Informes: los últimos 12 meses (de 30 días, como el panel), del más viejo al actual.
func Informes(ctx context.Context, pool tx, tiendaID string, ahora time.Time) (map[string]any, error) {
	ventas := make([]int64, 12)
	unidades := make([]int64, 12)
	clientes := make([]int64, 12)
	pedidos := make([]int64, 12)
	margen := make([]*int, 12)
	var primero *time.Time
	var visitasMes []*int64
	// Meses calendario en hora de Argentina: el 11 es el mes en curso. (Antes eran tramos de 30
	// días desde hoy, y un mes podía repetirse y otro faltar en el eje.)
	local := ahora.In(argentina)
	mesActual := time.Date(local.Year(), local.Month(), 1, 0, 0, 0, 0, argentina)
	inicio := mesActual.AddDate(0, -11, 0)
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		filas, err := t.Query(ctx, `
			WITH p AS (
				SELECT (extract(year FROM creado_en AT TIME ZONE 'America/Argentina/Buenos_Aires')::int * 12
					+ extract(month FROM creado_en AT TIME ZONE 'America/Argentina/Buenos_Aires')::int) - $4::int AS k,
					total, items, comprador->>'nombre' AS quien, creado_en
				FROM pedidos WHERE tienda_id = $1 AND estado <> 'Cancelado' AND NOT (estado = 'Sin pagar' AND canal = 'web' AND pago = 'mp') AND creado_en >= $3 AND creado_en <= $2
			), primeras AS (
				SELECT comprador->>'nombre' AS quien, min(creado_en) AS en FROM pedidos WHERE tienda_id = $1 AND estado <> 'Cancelado' AND NOT (estado = 'Sin pagar' AND canal = 'web' AND pago = 'mp') GROUP BY 1
			)
			SELECT k, sum(total), sum((SELECT coalesce(sum((i->>'cantidad')::int), 0) FROM jsonb_array_elements(items) i)),
				count(DISTINCT quien) FILTER (WHERE (SELECT en FROM primeras f WHERE f.quien = p.quien) = p.creado_en),
				count(*)
			FROM p GROUP BY k`, tiendaID, ahora, inicio, inicio.Year()*12+int(inicio.Month()))
		if err != nil {
			return err
		}
		for filas.Next() {
			var k int
			var v, u, c, n int64
			if err := filas.Scan(&k, &v, &u, &c, &n); err != nil {
				return err
			}
			if k >= 0 && k < 12 {
				ventas[k], unidades[k], clientes[k], pedidos[k] = v, u, c, n
			}
		}
		if err := filas.Err(); err != nil {
			return err
		}
		if visitasMes, err = visitas.PorMes(ctx, t, tiendaID, inicio); err != nil {
			return err
		}
		return t.QueryRow(ctx, `SELECT min(creado_en) FROM pedidos WHERE tienda_id = $1`, tiendaID).Scan(&primero)
	})
	if err != nil {
		return nil, err
	}
	nombres := make([]string, 12)
	for i := range 12 {
		nombres[i] = meses[inicio.AddDate(0, i, 0).Month()-1]
	}
	vs := func(a, b int64) *float64 {
		if b == 0 {
			return nil
		}
		x := float64(int64(float64(a-b)/float64(b)*1000)) / 10
		return &x
	}
	ticket := func(k int) int64 {
		if pedidos[k] == 0 {
			return 0
		}
		return ventas[k] / pedidos[k]
	}
	// Meses calendario con ventas desde el primer pedido (el mes del primero cuenta).
	historia := 0
	if primero != nil {
		p := primero.In(argentina)
		historia = (local.Year()-p.Year())*12 + int(local.Month()-p.Month()) + 1
	}
	series := map[string]any{"ventas": ventas, "productos": unidades, "clientes": clientes, "margen": margen}
	// Las visitas, sólo si ya se miden (null en los meses de antes de empezar a medir).
	if visitasMes != nil {
		series["visitas"] = visitasMes
	}
	return map[string]any{
		"meses":  nombres,
		"series": series,
		"kpis": map[string]any{
			"ventas": ventas[11], "ventasVs": vs(ventas[11], ventas[10]),
			"ticket": ticket(11), "ticketVs": vs(ticket(11), ticket(10)),
			"margen": nil, "margenVs": nil,
			// Contra una plataforma que cobra 2,1 % por venta.
			"ahorro":      ventas[11] * 21 / 1000 / 1000 * 1000,
			"mesAnterior": mesesLargos[mesActual.AddDate(0, -1, 0).Month()-1],
		},
		"historia": historia,
		"insights": []string{},
	}, nil
}

// Marketing: audiencias y lo necesario para armar una campaña.
func Marketing(ctx context.Context, pool tx, tiendaID string) (map[string]any, error) {
	res := map[string]any{}
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var nombre string
		var conWa int
		if err := t.QueryRow(ctx, `SELECT nombre, (SELECT count(DISTINCT comprador->>'telefono') FROM pedidos
			WHERE tienda_id = $1 AND comprador->>'telefono' IS NOT NULL) FROM tiendas WHERE id = $1`, tiendaID).Scan(&nombre, &conWa); err != nil {
			return err
		}
		var prod *struct {
			Nombre string  `json:"nombre"`
			Foto   *string `json:"foto"`
		}
		var pn string
		var pf *string
		if err := t.QueryRow(ctx, `SELECT nombre, foto FROM productos WHERE tienda_id = $1 ORDER BY creado_en LIMIT 1`, tiendaID).Scan(&pn, &pf); err == nil {
			prod = &struct {
				Nombre string  `json:"nombre"`
				Foto   *string `json:"foto"`
			}{pn, pf}
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		var categoria *string
		_ = t.QueryRow(ctx, `SELECT categoria FROM productos WHERE tienda_id = $1 AND categoria IS NOT NULL
			GROUP BY categoria ORDER BY count(*) DESC LIMIT 1`, tiendaID).Scan(&categoria)
		cat := "un producto"
		if categoria != nil {
			cat = strings.ToLower(*categoria)
		}
		type cup struct {
			Codigo      string `json:"codigo"`
			Descripcion string `json:"descripcion"`
			Usos        string `json:"usos"`
		}
		filas, err := t.Query(ctx, `SELECT codigo, descripcion, usos::text || ' usos' FROM cupones WHERE tienda_id = $1 AND NOT pausado AND origen IS NULL
			AND (empieza IS NULL OR empieza <= now()) AND (termina IS NULL OR termina > now()) AND (limite = 0 OR usos < limite)
			ORDER BY usos DESC, creado_en DESC LIMIT 3`, tiendaID)
		if err != nil {
			return err
		}
		cupones, err := pgx.CollectRows(filas, pgx.RowToStructByPos[cup])
		if err != nil {
			return err
		}
		var cupon *string
		_ = t.QueryRow(ctx, `SELECT codigo FROM cupones WHERE tienda_id = $1 AND tipo = 'pct' AND NOT pausado AND origen IS NULL
			AND (termina IS NULL OR termina > now()) ORDER BY usos DESC LIMIT 1`, tiendaID).Scan(&cupon)
		que := "lo nuevo"
		var foto *string
		if prod != nil {
			que, foto = prod.Nombre, prod.Foto
		}
		res = map[string]any{
			"tienda": map[string]any{"nombre": nombre, "foto": foto},
			"audiencias": []map[string]any{
				{"id": "todos", "nombre": "Todos", "cantidad": conWa},
				{"id": "vip", "nombre": "VIP", "cantidad": 0},
				{"id": "riesgo", "nombre": "En riesgo", "cantidad": 0},
				{"id": "categoria", "nombre": "Compraron " + cat, "cantidad": 0},
			},
			"costoMensaje": costoMensaje,
			"retorno":      nil,
			"mensajes": []string{
				"¡Hola {nombre}! Este finde " + que + " tiene 10% off solo para clientes. ¿Te lo reservo?",
				"{nombre}, ¡volvió " + que + "! Para vos, 10% off hasta el domingo.",
				"Hola {nombre} 👋 Armamos algo especial en " + nombre + ". Para vos, 10% off este finde.",
			},
			"producto": prod,
			"cupon":    cupon,
			"campanas": []any{},
			"cupones":  cupones,
		}
		return nil
	})
	return res, err
}

// Rutas del panel.
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
	escribir := func(w http.ResponseWriter, v any, err error) {
		if err != nil {
			problema.Escribir(w, problema.Interno())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(v)
	}
	mux.HandleFunc("GET /v1/panel/marketing", con(func(w http.ResponseWriter, r *http.Request, id string) {
		v, err := Marketing(r.Context(), pool, id)
		escribir(w, v, err)
	}))
	mux.HandleFunc("POST /v1/panel/campanas", con(func(w http.ResponseWriter, _ *http.Request, _ string) {
		problema.Escribir(w, problema.Conflicto("sin_whatsapp", "Conectá WhatsApp Business para mandar campañas."))
	}))
	mux.HandleFunc("GET /v1/panel/canales", con(func(w http.ResponseWriter, r *http.Request, id string) {
		v, err := Canales(r.Context(), pool, id, time.Now())
		escribir(w, v, err)
	}))
	mux.HandleFunc("PUT /v1/panel/canales/{id}", con(func(w http.ResponseWriter, r *http.Request, id string) {
		var b struct {
			Activo bool `json:"activo"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<10)).Decode(&b); err != nil {
			problema.Escribir(w, problema.Invalido("Falta si va activo."))
			return
		}
		switch err := FijarCanal(r.Context(), pool, id, r.PathValue("id"), b.Activo); {
		case errors.Is(err, ErrNoExiste):
			problema.Escribir(w, problema.NoEncontrado("Ese canal no existe."))
		case errors.Is(err, ErrSinConectar):
			problema.Escribir(w, problema.Conflicto("sin_conectar", "Primero conectalo desde Integraciones."))
		case err != nil:
			problema.Escribir(w, problema.Interno())
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	mux.HandleFunc("GET /v1/panel/informes", con(func(w http.ResponseWriter, r *http.Request, id string) {
		v, err := Informes(r.Context(), pool, id, time.Now())
		escribir(w, v, err)
	}))
}
