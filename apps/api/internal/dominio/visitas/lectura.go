package visitas

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

// Comparado: el período y el anterior del mismo largo.
type Comparado struct {
	Actual   float64 `json:"actual"`
	Anterior float64 `json:"anterior"`
}

// Pagina: las vistas de un tipo de página en el período.
type Pagina struct {
	Pagina string `json:"pagina"`
	Vistas int    `json:"vistas"`
}

// Producto: un producto con sus vistas en el período.
type Producto struct {
	ProductoID string  `json:"productoId"`
	Nombre     string  `json:"nombre"`
	Foto       *string `json:"foto"`
	Vistas     int     `json:"vistas"`
	Visitantes int     `json:"visitantes"`
}

// Resumen son las visitas del Resumen del panel (Tablero.visitas). Visitas = personas distintas
// de cada día, sumadas: la sal cambia cada día, así que la misma persona dos días son dos visitas.
type Resumen struct {
	// Desde: el primer día con visitas medidas (nil: todavía ninguna).
	Desde *string `json:"desde"`
	// Hoy (contra ayer completo), los últimos 7 y los últimos 30 días (contra los anteriores).
	Hoy    Comparado `json:"hoy"`
	Dias7  Comparado `json:"dias7"`
	Dias30 Comparado `json:"dias30"`
	// Del período del tablero: visitas, páginas vistas, por página y los productos más vistos.
	Visitantes Comparado  `json:"visitantes"`
	Vistas     Comparado  `json:"vistas"`
	Paginas    []Pagina   `json:"paginas"`
	MasVistos  []Producto `json:"masVistos"`
}

// DiasDe: los días del período del tablero (hoy, 7d o 30d).
func DiasDe(periodo string) int {
	switch periodo {
	case "hoy":
		return 1
	case "30d":
		return 30
	}
	return 7
}

// Ventana: el primer y el último día (inclusive) de los `dias` que terminan hoy, y los del período
// anterior del mismo largo.
func Ventana(ahora time.Time, dias int) (desde, hasta, antesDesde, antesHasta time.Time) {
	hasta = Dia(ahora)
	desde = hasta.AddDate(0, 0, -(dias - 1))
	antesHasta = desde.AddDate(0, 0, -1)
	antesDesde = desde.AddDate(0, 0, -dias)
	return
}

// Leer arma el resumen de visitas para el período del tablero (en la transacción de la tienda).
func Leer(ctx context.Context, t pgx.Tx, tiendaID string, periodo string, ahora time.Time) (Resumen, error) {
	r := Resumen{Paginas: []Pagina{}, MasVistos: []Producto{}}
	var desde *time.Time
	if err := t.QueryRow(ctx, `SELECT min(dia) FROM visitas WHERE tienda_id = $1 AND pagina = 'tienda'`, tiendaID).Scan(&desde); err != nil {
		return r, err
	}
	if desde == nil {
		return r, nil
	}
	d := desde.Format("2006-01-02")
	r.Desde = &d

	// Por día de los últimos 60: alcanza para hoy/ayer, 7 y 30 contra sus anteriores.
	hoy := Dia(ahora)
	porDia := map[string][2]int{}
	filas, err := t.Query(ctx, `SELECT dia, visitantes, vistas FROM visitas
		WHERE tienda_id = $1 AND pagina = 'tienda' AND dia > $2`, tiendaID, hoy.AddDate(0, 0, -60))
	if err != nil {
		return r, err
	}
	for filas.Next() {
		var dia time.Time
		var u, v int
		if err := filas.Scan(&dia, &u, &v); err != nil {
			filas.Close()
			return r, err
		}
		porDia[dia.Format("2006-01-02")] = [2]int{u, v}
	}
	filas.Close()
	if err := filas.Err(); err != nil {
		return r, err
	}
	sumar := func(dias int, i int) Comparado {
		a, b, c, e := Ventana(ahora, dias)
		var x Comparado
		for d := a; !d.After(b); d = d.AddDate(0, 0, 1) {
			x.Actual += float64(porDia[d.Format("2006-01-02")][i])
		}
		for d := c; !d.After(e); d = d.AddDate(0, 0, 1) {
			x.Anterior += float64(porDia[d.Format("2006-01-02")][i])
		}
		return x
	}
	r.Hoy, r.Dias7, r.Dias30 = sumar(1, 0), sumar(7, 0), sumar(30, 0)
	dias := DiasDe(periodo)
	r.Visitantes, r.Vistas = sumar(dias, 0), sumar(dias, 1)

	a, b, _, _ := Ventana(ahora, dias)
	filas, err = t.Query(ctx, `SELECT pagina, sum(vistas)::int FROM visitas
		WHERE tienda_id = $1 AND dia BETWEEN $2 AND $3 AND pagina <> 'tienda'
		GROUP BY 1 ORDER BY 2 DESC, 1`, tiendaID, a, b)
	if err != nil {
		return r, err
	}
	for filas.Next() {
		var p Pagina
		if err := filas.Scan(&p.Pagina, &p.Vistas); err != nil {
			filas.Close()
			return r, err
		}
		r.Paginas = append(r.Paginas, p)
	}
	filas.Close()
	if err := filas.Err(); err != nil {
		return r, err
	}
	r.MasVistos, err = MasVistos(ctx, t, tiendaID, a, b, 5)
	return r, err
}

// MasVistos: los productos con más vistas entre dos días (inclusive).
func MasVistos(ctx context.Context, t pgx.Tx, tiendaID string, desde, hasta time.Time, cuantos int) ([]Producto, error) {
	ps := []Producto{}
	filas, err := t.Query(ctx, `SELECT v.producto_id::text, p.nombre, p.foto, sum(v.vistas)::int, sum(v.visitantes)::int
		FROM visitas v JOIN productos p ON p.tienda_id = v.tienda_id AND p.id = v.producto_id
		WHERE v.tienda_id = $1 AND v.pagina = 'producto' AND v.dia BETWEEN $2 AND $3
		GROUP BY 1, 2, 3 ORDER BY 4 DESC, 2 LIMIT $4`, tiendaID, desde, hasta, cuantos)
	if err != nil {
		return ps, err
	}
	defer filas.Close()
	for filas.Next() {
		var p Producto
		if err := filas.Scan(&p.ProductoID, &p.Nombre, &p.Foto, &p.Vistas, &p.Visitantes); err != nil {
			return ps, err
		}
		ps = append(ps, p)
	}
	return ps, filas.Err()
}

// Conversion: pedidos de la tienda online (sin los cancelados) por cada 100 visitas, con un
// decimal, en los días del período y del anterior. Sin visitas, 0.
func Conversion(ctx context.Context, t pgx.Tx, tiendaID string, visitantes Comparado, ahora time.Time, dias int) (Comparado, error) {
	a, b, c, e := Ventana(ahora, dias)
	var actual, anterior float64
	err := t.QueryRow(ctx, `SELECT
			count(*) FILTER (WHERE (creado_en AT TIME ZONE 'America/Argentina/Buenos_Aires')::date BETWEEN $2 AND $3),
			count(*) FILTER (WHERE (creado_en AT TIME ZONE 'America/Argentina/Buenos_Aires')::date BETWEEN $4 AND $5)
		FROM pedidos WHERE tienda_id = $1 AND canal = 'web' AND estado <> 'Cancelado' AND NOT (estado = 'Sin pagar' AND canal = 'web' AND pago = 'mp')
			AND creado_en >= $4::date - interval '1 day'`, tiendaID, a, b, c, e).Scan(&actual, &anterior)
	tasa := func(pedidos, visitas float64) float64 {
		if visitas == 0 {
			return 0
		}
		return float64(int64(pedidos/visitas*1000+0.5)) / 10
	}
	return Comparado{tasa(actual, visitantes.Actual), tasa(anterior, visitantes.Anterior)}, err
}

// PorMes: las visitas de cada mes calendario desde `inicio` (12 meses); nil en los meses
// anteriores al primer día medido.
func PorMes(ctx context.Context, t pgx.Tx, tiendaID string, inicio time.Time) ([]*int64, error) {
	meses := make([]*int64, 12)
	var desde *time.Time
	if err := t.QueryRow(ctx, `SELECT min(dia) FROM visitas WHERE tienda_id = $1 AND pagina = 'tienda'`, tiendaID).Scan(&desde); err != nil || desde == nil {
		return nil, err
	}
	primero := desde.Year()*12 + int(desde.Month()) - (inicio.Year()*12 + int(inicio.Month()))
	for i := max(0, primero); i < 12; i++ {
		meses[i] = new(int64)
	}
	filas, err := t.Query(ctx, `SELECT (extract(year FROM dia)::int * 12 + extract(month FROM dia)::int) - $3::int, sum(visitantes)
		FROM visitas WHERE tienda_id = $1 AND pagina = 'tienda' AND dia >= $2::date GROUP BY 1`,
		tiendaID, inicio.Format("2006-01-02"), inicio.Year()*12+int(inicio.Month()))
	if err != nil {
		return nil, err
	}
	defer filas.Close()
	for filas.Next() {
		var k int
		var n int64
		if err := filas.Scan(&k, &n); err != nil {
			return nil, err
		}
		if k >= 0 && k < 12 && meses[k] != nil {
			*meses[k] = n
		}
	}
	return meses, filas.Err()
}

// ParaHornero: lo que Tincho sabe de las visitas: desde cuándo se miden, las de cada uno de los
// últimos 14 días, las de 30 días contra los 30 anteriores, la conversión y los más vistos.
// nil si todavía no hay ninguna (Tincho dice que no se miden, no inventa).
func ParaHornero(ctx context.Context, t pgx.Tx, tiendaID string, ahora time.Time) (map[string]any, error) {
	r, err := Leer(ctx, t, tiendaID, "30d", ahora)
	if err != nil || r.Desde == nil {
		return nil, err
	}
	conv, err := Conversion(ctx, t, tiendaID, r.Dias30, ahora, 30)
	if err != nil {
		return nil, err
	}
	hoy := Dia(ahora)
	por := map[string]int{}
	filas, err := t.Query(ctx, `SELECT dia, visitantes FROM visitas WHERE tienda_id = $1 AND pagina = 'tienda' AND dia > $2`,
		tiendaID, hoy.AddDate(0, 0, -14))
	if err != nil {
		return nil, err
	}
	for filas.Next() {
		var d time.Time
		var n int
		if err := filas.Scan(&d, &n); err != nil {
			filas.Close()
			return nil, err
		}
		por[d.Format("2006-01-02")] = n
	}
	filas.Close()
	if err := filas.Err(); err != nil {
		return nil, err
	}
	dias := []any{}
	for d := hoy.AddDate(0, 0, -13); !d.After(hoy); d = d.AddDate(0, 0, 1) {
		f := d.Format("2006-01-02")
		dias = append(dias, map[string]any{"fecha": f, "visitas": por[f]})
	}
	vistos := []any{}
	for _, p := range r.MasVistos {
		vistos = append(vistos, map[string]any{"nombre": p.Nombre, "vistas": p.Vistas, "visitas": p.Visitantes})
	}
	return map[string]any{
		"medidasDesde":         *r.Desde,
		"visitasPorDia":        dias,
		"visitas30d":           r.Dias30.Actual,
		"visitas30dAnteriores": r.Dias30.Anterior,
		"paginasVistas30d":     r.Vistas.Actual,
		"conversion30d":        conv.Actual,
		"masVistos30d":         vistos,
	}, nil
}
