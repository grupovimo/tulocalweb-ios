package vidriera

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
)

// lanzamiento es lo que la tienda configuró en el producto para el lanzamiento, la oferta y el
// encargo (§52), tal como está en la base.
type lanzamiento struct {
	fecha            *time.Time
	reserva, encargo *int16
	oferta           *int64
	hasta            *time.Time
}

// aplicar deja en el producto público lo que vale ahora: la oferta vigente (precio de oferta y el de
// lista tachado), «Próximamente» antes del lanzamiento y las señas que se pueden pagar (sólo con
// Mercado Pago conectado, que es con lo que se cobran).
func (l lanzamiento) aplicar(p *ProductoPublico, mp bool, ahora time.Time) {
	if l.oferta != nil && *l.oferta < p.Precio && (l.hasta == nil || l.hasta.After(ahora)) {
		anterior := p.Precio
		p.Anterior, p.Precio, p.OfertaHasta = &anterior, *l.oferta, l.hasta
	}
	pct := func(v *int16) *int {
		if v == nil || !mp {
			return nil
		}
		n := int(*v)
		return &n
	}
	if l.fecha != nil && l.fecha.After(ahora) {
		e := "Próximamente"
		p.Lanzamiento, p.Etiqueta, p.Reserva = l.fecha, &e, pct(l.reserva)
		return
	}
	p.Encargo = pct(l.encargo)
}

// fijarPrecio anota el precio de una línea para el pedido (por variante si tiene).
func fijarPrecio(m map[string]int64, producto, variante string, precio int64) map[string]int64 {
	if m == nil {
		m = map[string]int64{}
	}
	if variante != "" {
		m[producto+":"+variante] = precio
	} else {
		m[producto] = precio
	}
	return m
}

// Senador cobra la seña de un pedido con Mercado Pago (el paquete mercadopago).
type Senador interface {
	PagarSena(ctx context.Context, tiendaID string, numero int, vuelta string) (string, error)
}

// EsperaSena: cuánto se guarda el stock esperando el pago de la seña. La preferencia de Mercado
// Pago vence a la misma hora; después, el worker cancela el pedido y el stock vuelve.
const EsperaSena = 30 * time.Minute

// ErrSinSena: el producto no se reserva ni se encarga con seña, o la tienda no cobra con Mercado Pago.
var ErrSinSena = errors.New("vidriera: sin seña")

// ErrMercadoPago: Mercado Pago no respondió al armar el pago de la seña (el pedido no quedó).
var ErrMercadoPago = errors.New("vidriera: mercado pago no respondió")

// PedidoSena es lo que manda quien reserva o encarga: el producto (y su variante) y sus datos. El
// precio y la seña NO vienen: los calcula el servidor.
type PedidoSena struct {
	ProductoID string `json:"productoId"`
	VarianteID string `json:"varianteId"`
	Cantidad   int    `json:"cantidad"`
	Nombre     string `json:"nombre"`
	Email      string `json:"email"`
	Telefono   string `json:"telefono"`
}

// ResultadoSena: el pedido, lo que se paga ahora y adónde pagarlo.
type ResultadoSena struct {
	Numero int    `json:"numero"`
	Codigo string `json:"codigo"`
	Total  int64  `json:"total"`
	Sena   int64  `json:"sena"`
	URL    string `json:"url"`
}

// SenaDe: el porcentaje del total, redondeado al peso (como la seña de los turnos), nunca menos de $ 1.
func SenaDe(total int64, pct int) int64 {
	return max(1, (total*int64(pct)+50)/100)
}

// Senar reserva un lanzamiento o encarga un producto pagando una seña. El pedido se crea con el
// stock descontado (el cupo de la reserva es el stock) y queda «Sin pagar» hasta que llega la seña;
// si no se paga en EsperaSena, se cancela y el stock vuelve. El retiro (o la entrega) se coordina
// con la tienda: el resto se paga ahí.
func Senar(ctx context.Context, pool tx, mp Senador, sitio, slug string, s PedidoSena, ahora time.Time) (ResultadoSena, error) {
	t, err := Publica(ctx, pool, slug)
	if err != nil {
		return ResultadoSena{}, err
	}
	id, _ := tiendaDe(ctx, pool, slug)
	s.Nombre, s.Email, s.Telefono = strings.TrimSpace(s.Nombre), strings.ToLower(strings.TrimSpace(s.Email)), strings.TrimSpace(s.Telefono)
	if s.Cantidad == 0 {
		s.Cantidad = 1
	}
	if s.Cantidad < 1 || s.Cantidad > 10 || len(s.Nombre) < 2 || len(s.Nombre) > 80 || !emailValido.MatchString(s.Email) ||
		len(s.Email) > 200 || len(s.Telefono) > 30 || t.Vitrina {
		return ResultadoSena{}, ErrInvalido
	}
	var p *ProductoPublico
	for i := range t.Productos {
		if t.Productos[i].ID == s.ProductoID {
			p = &t.Productos[i]
		}
	}
	if p == nil {
		return ResultadoSena{}, ErrInvalido
	}
	tipo, pct := "encargo", p.Encargo
	if p.Lanzamiento != nil {
		tipo, pct = "reserva", p.Reserva
	}
	if pct == nil || mp == nil {
		return ResultadoSena{}, ErrSinSena
	}
	precio, ok := precioDe(*p, s.VarianteID)
	if !ok {
		return ResultadoSena{}, ErrInvalido
	}
	total := precio * int64(s.Cantidad)
	sena := SenaDe(total, *pct)
	vence := ahora.Add(EsperaSena)
	nota := "Reserva del lanzamiento con seña"
	if tipo == "encargo" {
		nota = "Encargo con seña"
	}
	nota += fmt.Sprintf(": %s ahora, %s al retirar", pesos(sena), pesos(total-sena))
	if !t.Envios.Retiro {
		nota += " (la entrega se coordina con la tienda)"
	}
	b := make([]byte, 18)
	_, _ = rand.Read(b)
	f := false
	n := pedidos.PedidoNuevo{Canal: "web", Pago: "mp", Envio: "retiro", Pagado: &f, Cliente: &s.Nombre, Telefono: &s.Telefono,
		Email: s.Email, Nota: &nota, CodigoPublico: base64.RawURLEncoding.EncodeToString(b),
		Items:   []pedidos.Linea{{ProductoID: s.ProductoID, VarianteID: s.VarianteID, Cantidad: s.Cantidad}},
		Precios: fijarPrecio(nil, s.ProductoID, s.VarianteID, precio),
		Sena:    sena, SenaTipo: tipo, SenaVence: &vence}
	numero, total, err := pedidos.CrearDesdeProductos(ctx, pool, id, n)
	switch {
	case errors.Is(err, pedidos.ErrSinStock):
		return ResultadoSena{}, ErrSinStock
	case errors.Is(err, pedidos.ErrVariante) || errors.Is(err, pedidos.ErrSinVariante) || errors.Is(err, pedidos.ErrProducto):
		return ResultadoSena{}, ErrInvalido
	case err != nil:
		return ResultadoSena{}, err
	}
	r := ResultadoSena{Numero: numero, Codigo: n.CodigoPublico, Total: total, Sena: sena}
	r.URL, err = mp.PagarSena(ctx, id, numero, fmt.Sprintf("%s/t/%s/pedido/%s", sitio, slug, r.Codigo))
	if err != nil {
		// Sin el pago armado, el pedido no puede quedar tomando el stock.
		if _, e := pedidos.CancelarSinSena(ctx, pool, id, numero, false); e != nil {
			return ResultadoSena{}, e
		}
		return ResultadoSena{}, ErrMercadoPago
	}
	return r, nil
}

func pesos(n int64) string {
	s := fmt.Sprint(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "." + s[i:]
	}
	return "$ " + s
}
