package vidriera

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// CuponTienda es un cupón como lo evalúa la tienda online (lib/tienda/cupon.ts hace lo mismo para
// mostrarlo: los dos pasan testdata/cupones.json).
type CuponTienda struct {
	Codigo     string     `json:"codigo"`
	Tipo       string     `json:"tipo"`
	Valor      int64      `json:"valor"`
	Pausado    bool       `json:"pausado"`
	Usos       int        `json:"usos"`
	Limite     int        `json:"limite"`
	Canales    []string   `json:"canales"`
	Empieza    *time.Time `json:"empieza"`
	Termina    *time.Time `json:"termina"`
	Minimo     int64      `json:"minimo"`
	Quien      string     `json:"quien"`
	PorCliente string     `json:"porCliente"`
	Alcance    struct {
		Tipo      string   `json:"tipo"`
		Productos []string `json:"productos"`
		Categoria string   `json:"categoria"`
	} `json:"alcance"`
	Combinable bool `json:"combinable"`
}

type LineaCupon struct {
	ID        string  `json:"id"`
	Precio    int64   `json:"precio"`
	Cantidad  int     `json:"cantidad"`
	Categoria *string `json:"categoria"`
}

func (c CuponTienda) vale(l LineaCupon) bool {
	switch {
	case c.Alcance.Tipo == "productos":
		return slices.Contains(c.Alcance.Productos, l.ID)
	case c.Alcance.Tipo == "categoria" && c.Alcance.Categoria != "":
		return l.Categoria != nil && *l.Categoria == c.Alcance.Categoria
	}
	return true
}

// Historial de quien compra, por su email.
type Historial struct {
	Pedidos      int        `json:"pedidos"`
	UsosDelCupon int        `json:"usosDelCupon"`
	VIP          bool       `json:"vip"`
	UltimaCompra *time.Time `json:"ultimaCompra"`
}

type ResultadoCupon struct {
	OK          bool   `json:"ok"`
	Descuento   int64  `json:"descuento,omitempty"`
	EnvioGratis bool   `json:"envioGratis,omitempty"`
	Combinable  bool   `json:"combinable,omitempty"`
	Motivo      string `json:"motivo,omitempty"`
}

func no(m string) ResultadoCupon { return ResultadoCupon{Motivo: m} }

// AplicarCupon dice si el cupón vale para estas líneas y cuánto descuenta.
func AplicarCupon(c CuponTienda, lineas []LineaCupon, h Historial, ahora time.Time, canal string) ResultadoCupon {
	switch {
	case c.Pausado:
		return no("Este cupón está pausado.")
	case c.Empieza != nil && ahora.Before(*c.Empieza):
		return no("Este cupón todavía no empezó.")
	case c.Termina != nil && ahora.After(*c.Termina):
		return no("Este cupón ya venció.")
	case c.Limite > 0 && c.Usos >= c.Limite:
		return no("Este cupón ya se usó todas las veces posibles.")
	case !slices.Contains(c.Canales, canal):
		return no("Este cupón no vale en la tienda online.")
	case c.Tipo == "regalo":
		return no("Este cupón se canjea en el local.")
	}
	var subtotal, elegible int64
	for _, l := range lineas {
		subtotal += l.Precio * int64(l.Cantidad)
		if c.vale(l) {
			elegible += l.Precio * int64(l.Cantidad)
		}
	}
	if subtotal < c.Minimo {
		return no("Este cupón es para compras desde $ " + miles(c.Minimo) + ".")
	}
	switch c.Quien {
	case "nuevos":
		if h.Pedidos > 0 {
			return no("Este cupón es para la primera compra.")
		}
	case "vip":
		if !h.VIP {
			return no("Este cupón es para clientes VIP.")
		}
	case "riesgo":
		if h.Pedidos == 0 || h.UltimaCompra == nil || ahora.Sub(*h.UltimaCompra) <= 60*24*time.Hour {
			return no("Este cupón es para quienes hace un tiempo que no compran.")
		}
	}
	tope := math.MaxInt
	switch c.PorCliente {
	case "1":
		tope = 1
	case "3":
		tope = 3
	}
	if h.UsosDelCupon >= tope {
		if c.PorCliente == "1" {
			return no("Ya usaste este cupón.")
		}
		return no("Ya usaste este cupón las veces posibles.")
	}
	if elegible == 0 {
		return no("Este cupón no vale para los productos de tu carrito.")
	}
	if c.Tipo == "envio" {
		return ResultadoCupon{OK: true, EnvioGratis: true, Combinable: c.Combinable}
	}
	v := max(0, c.Valor)
	var d int64
	if c.Tipo == "pct" {
		d = int64(math.Round(float64(elegible) * float64(min(100, v)) / 100))
	} else {
		d = min(elegible, v)
	}
	return ResultadoCupon{OK: true, Descuento: d, Combinable: c.Combinable}
}

// miles formatea como es-AR: 30000 → 30.000.
func miles(n int64) string {
	s := []byte{}
	d := []byte(itoa(n))
	for i := range d {
		if i > 0 && (len(d)-i)%3 == 0 {
			s = append(s, '.')
		}
		s = append(s, d[i])
	}
	return string(s)
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		b = append([]byte{'-'}, b...)
	}
	return string(b)
}

// ErrCupon: el cupón no vale para esta compra (el motivo se le muestra a quien compra).
type ErrCupon struct{ Motivo string }

func (e ErrCupon) Error() string { return "vidriera: cupón: " + e.Motivo }

// precios: el de cada línea como se cobra (preciosCarrito: el de la variante o el mayorista); sin
// ellos (nil), el del producto.
func lineasDe(c Compra, porID map[string]ProductoPublico, precios []int64) []LineaCupon {
	var out []LineaCupon
	for i, l := range c.Carrito {
		if p, ok := porID[l.ID]; ok {
			precio := p.Precio
			if i < len(precios) {
				precio = precios[i]
			}
			out = append(out, LineaCupon{ID: l.ID, Precio: precio, Cantidad: l.Cantidad, Categoria: p.Categoria})
		}
	}
	return out
}

// EvaluarCupon busca el cupón de la tienda y el historial de quien compra (por email) y aplica la
// regla. Un código que no existe es un motivo más, no un error.
func EvaluarCupon(ctx context.Context, pool tx, tienda, codigo, email string, lineas []LineaCupon, ahora time.Time) (ResultadoCupon, error) {
	var c CuponTienda
	var h Historial
	email = strings.ToLower(strings.TrimSpace(email))
	err := pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		var alcance []byte
		err := t.QueryRow(ctx, `SELECT codigo, tipo, valor, pausado, usos, limite, canales, empieza, termina, minimo,
				quien, por_cliente, alcance, coalesce((opciones->>'combinable')::boolean, false)
			FROM cupones WHERE tienda_id = $1 AND codigo = $2`, tienda, codigo).
			Scan(&c.Codigo, &c.Tipo, &c.Valor, &c.Pausado, &c.Usos, &c.Limite, &c.Canales, &c.Empieza, &c.Termina,
				&c.Minimo, &c.Quien, &c.PorCliente, &alcance, &c.Combinable)
		if err != nil {
			return err
		}
		_ = json.Unmarshal(alcance, &c.Alcance)
		if email == "" {
			return nil
		}
		return t.QueryRow(ctx, `SELECT count(*) FILTER (WHERE estado <> 'Cancelado'),
				count(*) FILTER (WHERE estado <> 'Cancelado' AND cupon = $2),
				coalesce(bool_or(vip), false),
				max(creado_en) FILTER (WHERE estado <> 'Cancelado')
			FROM pedidos WHERE tienda_id = $1 AND lower(comprador->>'email') = $3`, tienda, codigo, email).
			Scan(&h.Pedidos, &h.UsosDelCupon, &h.VIP, &h.UltimaCompra)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return no("Ese cupón no existe. Revisá cómo está escrito."), nil
	}
	if err != nil {
		return ResultadoCupon{}, err
	}
	return AplicarCupon(c, lineas, h, ahora, "web"), nil
}
