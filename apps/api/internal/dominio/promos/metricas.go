package promos

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/mitienda"
)

// Numeros de una ventana (o de un día): components/schemas/PromoNumeros.
type Numeros struct {
	Vistas  int `json:"vistas"`
	Cierres int `json:"cierres"`
	Clics   int `json:"clics"`
	// Envios: contactos captados (captación y ruleta).
	Envios int `json:"envios"`
	// Ventas y Monto: pedidos (no cancelados) que usaron el cupón de la ventana, y su total.
	Ventas int   `json:"ventas"`
	Monto  int64 `json:"monto"`
}

func (n *Numeros) sumar(o Numeros) {
	n.Vistas += o.Vistas
	n.Cierres += o.Cierres
	n.Clics += o.Clics
	n.Envios += o.Envios
	n.Ventas += o.Ventas
	n.Monto += o.Monto
}

type Dia struct {
	Dia string `json:"dia"`
	Numeros
}

// MetricaVentana: components/schemas/PromoMetricaVentana.
type MetricaVentana struct {
	ID        string `json:"id"`
	Nombre    string `json:"nombre"`
	Contenido string `json:"contenido"`
	Activa    bool   `json:"activa"`
	// Borrada: tiene números pero ya no está en el diseño.
	Borrada bool `json:"borrada"`
	Numeros
	// TasaEnvio: envíos sobre vistas (0–1); null sin vistas.
	TasaEnvio *float64 `json:"tasaEnvio"`
	Dias      []Dia    `json:"dias"`
}

// Metricas: components/schemas/PromoMetricas.
type Metricas struct {
	Desde    string           `json:"desde"`
	Hasta    string           `json:"hasta"`
	Total    Numeros          `json:"total"`
	Ventanas []MetricaVentana `json:"ventanas"`
}

// Rango: desde y hasta (YYYY-MM-DD, inclusive). Vacíos: los últimos 30 días; hasta un año.
func Rango(desde, hasta string, ahora time.Time) (time.Time, time.Time, bool) {
	hoy := time.Date(ahora.In(argentina).Year(), ahora.In(argentina).Month(), ahora.In(argentina).Day(), 0, 0, 0, 0, argentina)
	h, d := hoy, hoy.AddDate(0, 0, -29)
	var err error
	if hasta != "" {
		if h, err = time.ParseInLocation(time.DateOnly, hasta, argentina); err != nil {
			return d, h, false
		}
	}
	if desde != "" {
		if d, err = time.ParseInLocation(time.DateOnly, desde, argentina); err != nil {
			return d, h, false
		}
	} else if hasta != "" {
		d = h.AddDate(0, 0, -29)
	}
	if h.Before(d) || h.Sub(d) > 366*24*time.Hour {
		return d, h, false
	}
	return d, h, true
}

// LeerMetricas: los números por ventana y día del rango. Las ventanas salen del borrador (con su
// nombre de ahora) y del publicado; las que tienen números y ya no están, como borradas.
func LeerMetricas(ctx context.Context, pool tx, tiendaID string, desde, hasta time.Time) (Metricas, error) {
	out := Metricas{Desde: desde.Format(time.DateOnly), Hasta: hasta.Format(time.DateOnly), Ventanas: []MetricaVentana{}}
	fin := hasta.AddDate(0, 0, 1)
	porVentana := map[string]map[string]*Numeros{}
	sumar := func(v, dia string, f func(*Numeros)) {
		if porVentana[v] == nil {
			porVentana[v] = map[string]*Numeros{}
		}
		n := porVentana[v][dia]
		if n == nil {
			n = &Numeros{}
			porVentana[v][dia] = n
		}
		f(n)
	}
	var ventanas []mitienda.Ventana
	err := pool.EnTienda(ctx, tiendaID, func(x pgx.Tx) error {
		var plantilla string
		var borrador, publicado []byte
		if err := x.QueryRow(ctx, `SELECT plantilla, diseno_borrador, diseno_publicado FROM tiendas WHERE id = $1`, tiendaID).
			Scan(&plantilla, &borrador, &publicado); err != nil {
			return err
		}
		vistas := map[string]bool{}
		for _, d := range []mitienda.Diseno{mitienda.LeerDiseno(borrador, plantilla), mitienda.LeerDiseno(publicado, plantilla)} {
			for _, v := range d.Ventanas {
				if !vistas[v.ID] {
					vistas[v.ID] = true
					ventanas = append(ventanas, v)
				}
			}
		}
		filas, err := x.Query(ctx, `SELECT ventana, dia::text, vistas, cierres, clics FROM promo_metricas
			WHERE tienda_id = $1 AND dia BETWEEN $2::date AND $3::date`, tiendaID, out.Desde, out.Hasta)
		if err != nil {
			return err
		}
		for filas.Next() {
			var v, dia string
			var n Numeros
			if err := filas.Scan(&v, &dia, &n.Vistas, &n.Cierres, &n.Clics); err != nil {
				filas.Close()
				return err
			}
			sumar(v, dia, func(x *Numeros) {
				x.Vistas, x.Cierres, x.Clics = x.Vistas+n.Vistas, x.Cierres+n.Cierres, x.Clics+n.Clics
			})
		}
		filas.Close()
		if err := filas.Err(); err != nil {
			return err
		}
		filas, err = x.Query(ctx, `SELECT ventana, (creado_en AT TIME ZONE 'America/Argentina/Buenos_Aires')::date::text, count(*)
			FROM promo_captaciones WHERE tienda_id = $1 AND creado_en >= $2 AND creado_en < $3 GROUP BY 1, 2`, tiendaID, desde, fin)
		if err != nil {
			return err
		}
		for filas.Next() {
			var v, dia string
			var n int
			if err := filas.Scan(&v, &dia, &n); err != nil {
				filas.Close()
				return err
			}
			sumar(v, dia, func(x *Numeros) { x.Envios += n })
		}
		filas.Close()
		if err := filas.Err(); err != nil {
			return err
		}
		// Ventas: el cupón único de la ventana (origen), o el cupón fijo que muestra la ventana (contenido
		// cupón, o el premio cuponFijo de una captación).
		fijos := map[string]string{}
		for _, v := range ventanas {
			if c := codigoFijo(v); c != "" {
				if _, ya := fijos[c]; !ya {
					fijos[c] = v.ID
				}
			}
		}
		codigos := make([]string, 0, len(fijos))
		for c := range fijos {
			codigos = append(codigos, c)
		}
		filas, err = x.Query(ctx, `SELECT p.cupon, coalesce(c.origen, ''), (p.creado_en AT TIME ZONE 'America/Argentina/Buenos_Aires')::date::text, p.total
			FROM pedidos p LEFT JOIN cupones c ON c.tienda_id = p.tienda_id AND c.codigo = p.cupon AND c.origen IS NOT NULL
			WHERE p.tienda_id = $1 AND p.cupon IS NOT NULL AND p.estado <> 'Cancelado' AND p.creado_en >= $2 AND p.creado_en < $3
				AND (c.origen IS NOT NULL OR p.cupon = ANY($4))`, tiendaID, desde, fin, codigos)
		if err != nil {
			return err
		}
		for filas.Next() {
			var cupon, origen, dia string
			var total int64
			if err := filas.Scan(&cupon, &origen, &dia, &total); err != nil {
				filas.Close()
				return err
			}
			v, ok := strings.CutPrefix(origen, "ventana:")
			if !ok {
				v = fijos[cupon]
			}
			sumar(v, dia, func(x *Numeros) { x.Ventas++; x.Monto += total })
		}
		filas.Close()
		return filas.Err()
	})
	if err != nil {
		return out, err
	}
	enDiseno := map[string]bool{}
	agregar := func(m MetricaVentana) {
		dias := porVentana[m.ID]
		m.Dias = []Dia{}
		for dia, n := range dias {
			m.Dias = append(m.Dias, Dia{Dia: dia, Numeros: *n})
			m.Numeros.sumar(*n)
		}
		sort.Slice(m.Dias, func(i, j int) bool { return m.Dias[i].Dia < m.Dias[j].Dia })
		if m.Vistas > 0 {
			t := float64(m.Envios) / float64(m.Vistas)
			m.TasaEnvio = &t
		}
		out.Total.sumar(m.Numeros)
		out.Ventanas = append(out.Ventanas, m)
	}
	for _, v := range ventanas {
		enDiseno[v.ID] = true
		agregar(MetricaVentana{ID: v.ID, Nombre: v.Nombre, Contenido: v.Contenido, Activa: v.Activa})
	}
	var borradas []string
	for id := range porVentana {
		if !enDiseno[id] && id != "" {
			borradas = append(borradas, id)
		}
	}
	sort.Strings(borradas)
	for _, id := range borradas {
		agregar(MetricaVentana{ID: id, Nombre: id, Borrada: true})
	}
	return out, nil
}

// codigoFijo: el cupón existente que muestra la ventana (contenido cupón o premio cuponFijo).
func codigoFijo(v mitienda.Ventana) string {
	switch v.Contenido {
	case "cupon":
		return texto(v.Datos["cupon"])
	case "captacion":
		if p, _ := v.Datos["premio"].(map[string]any); texto(p["tipo"]) == "cuponFijo" {
			return texto(p["cupon"])
		}
	}
	return ""
}
