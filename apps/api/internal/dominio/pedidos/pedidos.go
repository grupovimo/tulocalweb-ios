// Package pedidos: los pedidos de cada tienda (components/schemas/PedidoFila, PedidoDetalle,
// ListaPedidos). La lista filtra en SQL y cuenta cada faceta sin su propio filtro, igual que la demo.
package pedidos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/auditoria"
)

// Circuito es el orden de las etapas (Cancelado va aparte).
var Circuito = []string{"Sin pagar", "Para preparar", "Listo para despachar", "En camino", "Entregado"}

// Estados válidos.
var Estados = append(slices.Clone(Circuito), "Cancelado")

var (
	ErrNoExiste   = errors.New("el pedido no existe")
	ErrTransicion = errors.New("el pedido no puede pasar a ese estado")
)

// CircuitoDe: el retiro en el local no pasa por "En camino" (Listo para retirar → Retirado).
func CircuitoDe(envio string) []string {
	if envio == "retiro" {
		return slices.DeleteFunc(slices.Clone(Circuito), func(e string) bool { return e == "En camino" })
	}
	return Circuito
}

// TransicionValida: avanzar una etapa, volver una (deshacer) o cancelar (y deshacer la cancelación),
// en el circuito del envío del pedido.
func TransicionValida(de, a, envio string) bool {
	if de == a || !slices.Contains(Estados, a) {
		return false
	}
	if a == "Cancelado" {
		return de != "Entregado"
	}
	if de == "Cancelado" {
		return a != "Entregado"
	}
	c := CircuitoDe(envio)
	if !slices.Contains(c, a) {
		return false
	}
	d := slices.Index(c, a) - slices.Index(c, de)
	return d == 1 || d == -1
}

type tx interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
}

// Item es un producto del pedido, tal como se vendió.
type Item struct {
	// El producto de la tienda (para «Volver a vender»); los pedidos de antes no lo guardaban.
	ProductoID *string `json:"productoId"`
	Nombre     string  `json:"nombre"`
	Foto       *string `json:"foto"`
	Cantidad   int     `json:"cantidad"`
	Precio     int64   `json:"precio"`
	// La variante vendida y su texto ("Talle M · Color Negro"), copiado al vender: sobrevive a que
	// la variante se borre o cambie.
	VarianteID *string `json:"varianteId"`
	Variante   *string `json:"variante"`
}

// Comprador es quien compró, copiado en el pedido.
type Comprador struct {
	Nombre   string  `json:"nombre"`
	Email    *string `json:"email"`
	Telefono *string `json:"telefono"`
	Foto     *string `json:"foto"`
	// Factura: quien compra pidió factura A a su CUIT (la tienda online de una responsable inscripta).
	Factura *Factura `json:"factura,omitempty"`
}

// Factura son los datos del receptor de una factura A: el CUIT ya validado y la razón social.
type Factura struct {
	CUIT        string `json:"cuit"`
	RazonSocial string `json:"razonSocial"`
}

// Nuevo es lo necesario para crear un pedido.
type Nuevo struct {
	Comprador Comprador
	Canal     string
	Pago      string
	Envio     string
	Ciudad    string
	Estado    string
	Items     []Item
	Hornero   bool
	Mayorista bool
	Creado    time.Time
}

// Crear da de alta un pedido con el próximo número de la tienda.
func Crear(ctx context.Context, pool tx, tiendaID string, n Nuevo) (int, error) {
	if len(n.Items) == 0 {
		return 0, errors.New("el pedido no tiene productos")
	}
	if n.Estado == "" {
		n.Estado = "Sin pagar"
	}
	if n.Creado.IsZero() {
		n.Creado = time.Now()
	}
	var sub int64
	for _, i := range n.Items {
		sub += i.Precio * int64(i.Cantidad)
	}
	comprador, _ := json.Marshal(n.Comprador)
	items, _ := json.Marshal(n.Items)
	var numero int
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		// El número es correlativo por tienda: se toma la fila de la tienda para no repetirlo.
		if _, err := t.Exec(ctx, `SELECT 1 FROM tiendas WHERE id = $1 FOR UPDATE`, tiendaID); err != nil {
			return err
		}
		return t.QueryRow(ctx, `
			INSERT INTO pedidos (tienda_id, numero, comprador, canal, pago, envio, ciudad_envio, estado,
				items, subtotal, total, hornero, mayorista, creado_en)
			SELECT $1, coalesce(max(numero), 1000) + 1, $2, $3, $4, $5, $6, $7, $8, $9, $9, $10, $11, $12
			FROM pedidos WHERE tienda_id = $1
			RETURNING numero`,
			tiendaID, comprador, n.Canal, n.Pago, n.Envio, n.Ciudad, n.Estado, items, sub, n.Hornero, n.Mayorista, n.Creado,
		).Scan(&numero)
	})
	return numero, err
}

// Filtros son los de la URL del panel.
type Filtros struct {
	Q                     string
	Estado, Canal, Pago   []string
	Envio                 []string
	Fecha                 string
	Minimo                int64
	VIP, Hornero, Mayoris bool
	Orden                 string
	Desde, Cuantos        int
}

// Producto es la miniatura de la fila.
type Producto struct {
	Nombre string  `json:"nombre"`
	Foto   *string `json:"foto"`
}

// Fila es un pedido en la lista.
type Fila struct {
	Numero            int        `json:"numero"`
	Cliente           string     `json:"cliente"`
	Foto              *string    `json:"foto"`
	Ciudad            string     `json:"ciudad"`
	Canal             string     `json:"canal"`
	Pago              string     `json:"pago"`
	Envio             string     `json:"envio"`
	Estado            string     `json:"estado"`
	Total             int64      `json:"total"`
	Creado            time.Time  `json:"creado"`
	Productos         []Producto `json:"productos"`
	CantidadProductos int        `json:"cantidadProductos"`
	VIP               bool       `json:"vip"`
	Hornero           bool       `json:"hornero"`
	Mayorista         bool       `json:"mayorista"`
	// Telefono: el WhatsApp de quien compró, para escribirle desde la lista.
	Telefono *string `json:"telefono"`
	// Codigo: el público de la confirmación (ahí se deja la reseña); null si se cargó a mano.
	Codigo *string `json:"codigo"`
}

// Lista es la respuesta de la lista.
type Lista struct {
	Total   int            `json:"total"`
	Suma    int64          `json:"suma"`
	Filas   []Fila         `json:"filas"`
	Etapas  map[string]int `json:"etapas"`
	Facetas struct {
		Canal map[string]int `json:"canal"`
		Pago  map[string]int `json:"pago"`
		Envio map[string]int `json:"envio"`
	} `json:"facetas"`
}

// condiciones arma el WHERE; `sin` deja afuera el filtro de esa faceta (para contarla).
func condiciones(f Filtros, ahora time.Time, sin string) (string, []any) {
	w := []string{"true"}
	args := []any{}
	arg := func(v any) string { args = append(args, v); return fmt.Sprintf("$%d", len(args)) }
	switch f.Fecha {
	case "hoy":
		w = append(w, "creado_en >= "+arg(ahora.Add(-24*time.Hour)))
	case "7d":
		w = append(w, "creado_en >= "+arg(ahora.Add(-7*24*time.Hour)))
	case "30d":
		w = append(w, "creado_en >= "+arg(ahora.Add(-30*24*time.Hour)))
	}
	if f.Minimo > 0 {
		w = append(w, "total >= "+arg(f.Minimo))
	}
	if f.VIP {
		w = append(w, "vip")
	}
	if f.Hornero {
		w = append(w, "hornero")
	}
	if f.Mayoris {
		w = append(w, "mayorista")
	}
	if q := strings.TrimSpace(f.Q); q != "" {
		p := arg("%" + strings.NewReplacer(`\`, `\\`, "%", `\%`, "_", `\_`).Replace(strings.TrimPrefix(q, "#")) + "%")
		w = append(w, fmt.Sprintf(`(numero::text || ' ' || (comprador->>'nombre') || ' ' || ciudad_envio || ' ' ||
			coalesce((SELECT string_agg(i->>'nombre', ' ') FROM jsonb_array_elements(items) i), '')) ILIKE %s`, p))
	}
	for _, c := range []struct {
		col  string
		vals []string
	}{{"estado", f.Estado}, {"canal", f.Canal}, {"pago", f.Pago}, {"envio", f.Envio}} {
		if c.col != sin && len(c.vals) > 0 {
			w = append(w, c.col+" = ANY("+arg(c.vals)+")")
		}
	}
	return strings.Join(w, " AND "), args
}

const columnasFila = `numero, comprador->>'nombre', comprador->>'foto', ciudad_envio, canal, pago, envio, estado,
	total, creado_en, items, vip, hornero, mayorista, nullif(comprador->>'telefono', ''), codigo_publico`

func escanearFila(r pgx.Row) (Fila, error) {
	var f Fila
	var items []Item
	err := r.Scan(&f.Numero, &f.Cliente, &f.Foto, &f.Ciudad, &f.Canal, &f.Pago, &f.Envio, &f.Estado,
		&f.Total, &f.Creado, &items, &f.VIP, &f.Hornero, &f.Mayorista, &f.Telefono, &f.Codigo)
	f.CantidadProductos = len(items)
	f.Productos = []Producto{}
	for i, it := range items {
		if i == 3 {
			break
		}
		f.Productos = append(f.Productos, Producto{Nombre: it.Nombre, Foto: it.Foto})
	}
	return f, err
}

// Listar devuelve la ventana pedida con los conteos.
func Listar(ctx context.Context, pool tx, tiendaID string, f Filtros, ahora time.Time) (Lista, error) {
	l := Lista{Filas: []Fila{}, Etapas: map[string]int{}}
	l.Facetas.Canal, l.Facetas.Pago, l.Facetas.Envio = map[string]int{}, map[string]int{}, map[string]int{}
	if f.Cuantos <= 0 || f.Cuantos > 200 {
		f.Cuantos = 60
	}
	orden := map[string]string{
		"monto":  "total DESC, numero DESC",
		"viejos": "creado_en, numero",
		"urgentes": `CASE WHEN estado IN ('Para preparar', 'Listo para despachar') THEN
				CASE WHEN creado_en > now() - interval '20 hours' THEN 3 ELSE 2 END
			WHEN estado = 'Sin pagar' AND creado_en < now() - interval '24 hours' THEN 1 ELSE 0 END DESC, creado_en DESC`,
	}[f.Orden]
	if orden == "" {
		orden = "creado_en DESC, numero DESC"
	}
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		w, args := condiciones(f, ahora, "")
		if err := t.QueryRow(ctx, `SELECT count(*), coalesce(sum(total), 0) FROM pedidos WHERE `+w, args...).Scan(&l.Total, &l.Suma); err != nil {
			return err
		}
		filas, err := t.Query(ctx, fmt.Sprintf(`SELECT %s FROM pedidos WHERE %s ORDER BY %s LIMIT %d OFFSET %d`,
			columnasFila, w, orden, f.Cuantos, max(0, f.Desde)), args...)
		if err != nil {
			return err
		}
		for filas.Next() {
			fi, err := escanearFila(filas)
			if err != nil {
				return err
			}
			l.Filas = append(l.Filas, fi)
		}
		if err := filas.Err(); err != nil {
			return err
		}
		for _, c := range []struct {
			col string
			m   map[string]int
		}{{"estado", l.Etapas}, {"canal", l.Facetas.Canal}, {"pago", l.Facetas.Pago}, {"envio", l.Facetas.Envio}} {
			w, args := condiciones(f, ahora, c.col)
			r, err := t.Query(ctx, fmt.Sprintf(`SELECT %s, count(*) FROM pedidos WHERE %s GROUP BY 1`, c.col, w), args...)
			if err != nil {
				return err
			}
			for r.Next() {
				var k string
				var n int
				if err := r.Scan(&k, &n); err != nil {
					return err
				}
				c.m[k] = n
			}
			if err := r.Err(); err != nil {
				return err
			}
		}
		return nil
	})
	return l, err
}

// Paso del historial.
type Paso struct {
	Texto  string     `json:"texto"`
	Cuando *time.Time `json:"cuando"`
	Hecho  bool       `json:"hecho"`
}

// Descuento es una línea del detalle: "Cupón VERANO10" o "Descuento por transferencia" y lo que restó.
type Descuento struct {
	Concepto string `json:"concepto"`
	Monto    int64  `json:"monto"`
}

// Detalle es el pedido completo (la hoja).
type Detalle struct {
	Fila
	Items      []Item `json:"items"`
	Subtotal   int64  `json:"subtotal"`
	CostoEnvio int64  `json:"costoEnvio"`
	// Descuentos: cada uno por separado; si suman menos que subtotal + envío − total (pedidos de
	// antes de guardarlos), la diferencia va como "Descuento".
	Descuentos []Descuento `json:"descuentos"`
	Comprador  struct {
		Comprador
		Desde   time.Time `json:"desde"`
		Pedidos int       `json:"pedidos"`
		Gastado int64     `json:"gastado"`
	} `json:"comprador"`
	EstadoPago  string   `json:"estadoPago"`
	Factura     *string  `json:"factura"`
	Seguimiento *string  `json:"seguimiento"`
	Llega       *string  `json:"llega"`
	Nota        *string  `json:"nota"`
	Entrega     *Entrega `json:"entrega"`
	Historial   []Paso   `json:"historial"`
}

// Obtener devuelve el detalle de un pedido.
func Obtener(ctx context.Context, pool tx, tiendaID string, numero int) (Detalle, error) {
	var d Detalle
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var comprador Comprador
		var estadosEn map[string]time.Time
		err := t.QueryRow(ctx, `SELECT `+columnasFila+`, comprador, subtotal, costo_envio, estado_pago, factura,
				seguimiento, nota, estados_en, entrega, descuentos
			FROM pedidos WHERE tienda_id = $1 AND numero = $2`, tiendaID, numero).Scan(
			&d.Numero, &d.Cliente, &d.Foto, &d.Ciudad, &d.Canal, &d.Pago, &d.Envio, &d.Estado, &d.Total, &d.Creado,
			&d.Items, &d.VIP, &d.Hornero, &d.Mayorista, &d.Telefono, &d.Codigo, &comprador, &d.Subtotal, &d.CostoEnvio, &d.EstadoPago,
			&d.Factura, &d.Seguimiento, &d.Nota, &estadosEn, &d.Entrega, &d.Descuentos)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoExiste
		}
		if err != nil {
			return err
		}
		d.Comprador.Comprador = comprador
		d.Descuentos = DescuentosCompletos(d.Descuentos, d.Subtotal+d.CostoEnvio-d.Total)
		d.CantidadProductos = len(d.Items)
		d.Productos = []Producto{}
		for i, it := range d.Items {
			if i < 3 {
				d.Productos = append(d.Productos, Producto{Nombre: it.Nombre, Foto: it.Foto})
			}
		}
		// Historia del comprador en la tienda (por email o, si no hay, por nombre).
		if err := t.QueryRow(ctx, `SELECT min(creado_en), count(*), coalesce(sum(total), 0) FROM pedidos
			WHERE tienda_id = $1 AND estado <> 'Cancelado' AND
				CASE WHEN $2::text IS NOT NULL THEN comprador->>'email' = $2 ELSE comprador->>'nombre' = $3 END`,
			tiendaID, comprador.Email, comprador.Nombre).Scan(&d.Comprador.Desde, &d.Comprador.Pedidos, &d.Comprador.Gastado); err != nil {
			return err
		}
		d.Historial = historial(d.Fila, d.Canal, d.Pago, d.Envio, d.Factura != nil, estadosEn)
		return nil
	})
	return d, err
}

// DescuentosCompletos suma, si falta, la línea que hace cerrar la cuenta (pedidos viejos).
func DescuentosCompletos(ds []Descuento, total int64) []Descuento {
	out := slices.Clone(ds)
	if out == nil {
		out = []Descuento{}
	}
	var suma int64
	for _, x := range out {
		suma += x.Monto
	}
	if total-suma > 0 {
		out = append(out, Descuento{Concepto: "Descuento", Monto: total - suma})
	}
	return out
}

func historial(f Fila, canal, pago, envio string, facturado bool, en map[string]time.Time) []Paso {
	cuando := func(e string) *time.Time {
		if t, ok := en[e]; ok {
			return &t
		}
		return nil
	}
	creado := f.Creado
	h := []Paso{{Texto: "Pedido creado por " + nombreCanal[canal], Cuando: &creado, Hecho: true}}
	if f.Estado == "Cancelado" {
		return append(h, Paso{Texto: "Pedido cancelado", Cuando: cuando("Cancelado"), Hecho: true})
	}
	paso := slices.Index(Circuito, f.Estado)
	if envio == "retiro" && paso >= 3 {
		paso = 4 // el retiro no pasa por "En camino"
	}
	if paso >= 1 {
		h = append(h, Paso{Texto: "Pago aprobado · " + nombrePago[pago], Cuando: cuando("Para preparar"), Hecho: true})
	} else {
		h = append(h, Paso{Texto: "Esperando el pago"})
	}
	if facturado {
		h = append(h, Paso{Texto: "Factura B emitida", Cuando: cuando("Para preparar"), Hecho: true})
	}
	if paso >= 2 {
		texto := "Preparado"
		if envio == "retiro" {
			texto = "Listo para retirar"
		}
		h = append(h, Paso{Texto: texto, Cuando: cuando("Listo para despachar"), Hecho: true})
	} else if envio == "retiro" {
		h = append(h, Paso{Texto: "Preparar para retirar"})
	} else {
		h = append(h, Paso{Texto: "Preparar y despachar"})
	}
	if envio != "retiro" {
		h = append(h, Paso{Texto: "Despachado con " + nombreEnvio[envio], Cuando: cuando("En camino"), Hecho: paso >= 3})
	}
	texto := "Entregado"
	if envio == "retiro" {
		texto = "Retirado"
	}
	return append(h, Paso{Texto: texto, Cuando: cuando("Entregado"), Hecho: paso >= 4})
}

var (
	nombreCanal = map[string]string{"web": "la tienda online", "wa": "WhatsApp", "ml": "Mercado Libre", "ig": "Instagram", "pos": "el local", "tg": "Telegram"}
	nombrePago  = map[string]string{"mp": "Mercado Pago", "tarjeta": "tarjeta", "payway": "Payway", "modo": "MODO", "uala": "Ualá", "transf": "transferencia", "efectivo": "efectivo"}
	nombreEnvio = map[string]string{"andreani": "Andreani", "oca": "OCA", "correo": "Correo Argentino", "moto": "Moto propia", "retiro": "Retiro en el local"}
)

// CambiarEstado avanza, vuelve atrás o cancela un pedido.
func CambiarEstado(ctx context.Context, pool tx, tiendaID string, numero int, estado string) error {
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var de, envio string
		err := t.QueryRow(ctx, `SELECT estado, envio FROM pedidos WHERE tienda_id = $1 AND numero = $2 FOR UPDATE`, tiendaID, numero).Scan(&de, &envio)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoExiste
		}
		if err != nil {
			return err
		}
		if !TransicionValida(de, estado, envio) {
			return ErrTransicion
		}
		// «Marcar como pagado» (Sin pagar → Para preparar) es la tienda diciendo que cobró: la
		// transferencia queda acreditada y lo demás cobrado (Mercado Pago lo confirma su aviso). Antes sólo
		// cambiaba la etapa y el pago seguía «Pendiente»: no salía el mail de pago ni se podía reseñar.
		// Deshacerlo (Para preparar → Sin pagar) lo vuelve a pendiente.
		_, err = t.Exec(ctx, `UPDATE pedidos SET estado = $3,
				estados_en = estados_en || jsonb_build_object($3::text, now()),
				estado_pago = CASE
					WHEN pago = 'mp' THEN estado_pago
					WHEN $4 = 'Sin pagar' AND $3 = 'Para preparar' AND estado_pago = 'Pendiente'
						THEN CASE WHEN pago = 'transf' THEN 'Acreditada' ELSE 'Cobrado' END
					WHEN $4 = 'Para preparar' AND $3 = 'Sin pagar' AND estado_pago IN ('Acreditada', 'Cobrado')
						THEN 'Pendiente'
					ELSE estado_pago END
			WHERE tienda_id = $1 AND numero = $2`, tiendaID, numero, estado, de)
		if err != nil {
			return err
		}
		return auditoria.Registrar(ctx, t, tiendaID, registroEstado(numero, de, estado))
	})
}

// registroEstado: cómo se lee en la Actividad un cambio de etapa.
func registroEstado(numero int, de, a string) auditoria.Registro {
	r := auditoria.Registro{Accion: "pedido.estado", Entidad: "pedido", EntidadID: fmt.Sprint(numero),
		Texto: fmt.Sprintf("pasó el pedido #%d de %s a %s", numero, strings.ToLower(de), strings.ToLower(a)),
		Antes: map[string]string{"estado": de}, Despues: map[string]string{"estado": a}}
	switch {
	case a == "Cancelado":
		r.Accion, r.Texto = "pedido.cancelar", fmt.Sprintf("canceló el pedido #%d", numero)
	case de == "Sin pagar" && a == "Para preparar":
		r.Accion, r.Texto = "pedido.pagado", fmt.Sprintf("marcó como pagado el pedido #%d", numero)
	case de == "Para preparar" && a == "Sin pagar":
		r.Accion, r.Texto = "pedido.pagado_deshecho", fmt.Sprintf("deshizo el pago del pedido #%d (volvió a sin pagar)", numero)
	}
	return r
}

// Cambio es lo que deja una acción en lote (para deshacer).
type Cambio struct {
	Numero int    `json:"numero"`
	Antes  string `json:"antes"`
}

// EnLote aplica una acción a varios pedidos; los que no están en la etapa justa se saltean.
// Facturar y avisar necesitan ARCA y WhatsApp conectados: hasta entonces no hacen nada.
func EnLote(ctx context.Context, pool tx, tiendaID string, numeros []int, accion string) ([]Cambio, error) {
	paso := map[string][2]string{
		"preparar":  {"Para preparar", "Listo para despachar"},
		"etiquetas": {"Listo para despachar", "En camino"},
	}[accion]
	de, a := paso[0], paso[1]
	cambios := []Cambio{}
	if de == "" {
		return cambios, nil
	}
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		// Los de retiro no se despachan: "etiquetas" no los pasa a En camino.
		r, err := t.Query(ctx, `UPDATE pedidos SET estado = $4, estados_en = estados_en || jsonb_build_object($4::text, now())
			WHERE tienda_id = $1 AND numero = ANY($2) AND estado = $3 AND ($4 <> 'En camino' OR envio <> 'retiro')
			RETURNING numero`, tiendaID, numeros, de, a)
		if err != nil {
			return err
		}
		for r.Next() {
			var n int
			if err := r.Scan(&n); err != nil {
				return err
			}
			cambios = append(cambios, Cambio{Numero: n, Antes: de})
		}
		if err := r.Err(); err != nil {
			return err
		}
		rs := make([]auditoria.Registro, 0, len(cambios))
		for _, c := range cambios {
			rs = append(rs, registroEstado(c.Numero, de, a))
		}
		return auditoria.Registrar(ctx, t, tiendaID, rs...)
	})
	return cambios, err
}
