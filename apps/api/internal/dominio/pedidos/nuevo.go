package pedidos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/auditoria"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

// ErrSinStock: no alcanza el stock de algún producto.
var ErrSinStock = errors.New("no alcanza el stock")

// ErrProducto: algún producto no existe en la tienda.
var ErrProducto = errors.New("el producto no existe")

// ErrSinVariante: el producto tiene opciones y la línea no dice cuál variante.
var ErrSinVariante = errors.New("elegí una variante")

// ErrVariante: la variante no es de ese producto, no existe o no está a la venta.
var ErrVariante = errors.New("la variante no está a la venta")

// Linea es un producto del pedido nuevo. Con opciones, VarianteID es obligatoria.
type Linea struct {
	ProductoID string `json:"productoId"`
	VarianteID string `json:"varianteId"`
	Cantidad   int    `json:"cantidad"`
}

// PedidoNuevo es lo que manda el punto de venta o Crear pedido (components/schemas/PedidoNuevo).
// Entrega: la dirección y el código postal del envío.
type Entrega struct {
	Direccion string `json:"direccion"`
	CP        string `json:"cp"`
	// Correo Argentino: D (domicilio) o S (sucursal), la provincia (código de MiCorreo), la localidad y
	// la sucursal elegida.
	Tipo           string `json:"tipo,omitempty"`
	Provincia      string `json:"provincia,omitempty"`
	Localidad      string `json:"localidad,omitempty"`
	Sucursal       string `json:"sucursal,omitempty"`
	SucursalNombre string `json:"sucursalNombre,omitempty"`
}

type PedidoNuevo struct {
	Items   []Linea `json:"items"`
	Pago    string  `json:"pago"`
	Canal   string  `json:"canal"`
	Envio   string  `json:"envio"`
	Cliente *string `json:"cliente"`
	// ClienteID: en la API los clientes salen de los compradores, así que alcanza con nombre y teléfono.
	ClienteID *string `json:"clienteId"`
	Telefono  *string `json:"telefono"`
	Nota      *string `json:"nota"`
	// Pagado en false: queda Sin pagar (se cobra con un link).
	Pagado *bool `json:"pagado"`
	// Precios cobrados en otro canal (Mercado Libre), por producto o "producto:variante". Nunca viene
	// del navegador.
	Precios map[string]int64 `json:"-"`
	// La tienda online: envío, descuento por transferencia, email y el código de la confirmación.
	// Los calcula el servidor (vidriera), nunca vienen del navegador.
	CostoEnvio    int64  `json:"-"`
	Descuento     int64  `json:"-"`
	Email         string `json:"-"`
	CiudadEnvio   string `json:"-"`
	CodigoPublico string `json:"-"`
	// Adónde se manda (la tienda online), para la etiqueta de envío.
	Entrega *Entrega `json:"-"`
	// Cupón ya validado por la vidriera y lo que descontó (va dentro de Descuento). El uso se cuenta
	// acá, en la transacción: si se agotó en el medio, ErrCupon.
	Cupon          string `json:"-"`
	DescuentoCupon int64  `json:"-"`
	// Factura A pedida en el checkout (validada por la vidriera): la cola de ARCA la emite a ese CUIT.
	Factura *Factura `json:"-"`
}

// ErrCupon: el cupón se agotó o se pausó entre que se validó y se usó.
var ErrCupon = errors.New("pedidos: cupón agotado")

var (
	pagos   = []string{"mp", "tarjeta", "payway", "modo", "uala", "transf", "efectivo"}
	canales = []string{"web", "wa", "ml", "ig", "pos", "tg"}
	envios  = []string{"andreani", "oca", "correo", "moto", "retiro", "mercadoenvios"}
)

// CrearDesdeProductos arma el pedido con los precios de la tienda y descuenta el stock, todo en una
// transacción: si un producto no alcanza, no se descuenta nada.
func CrearDesdeProductos(ctx context.Context, pool tx, tiendaID string, n PedidoNuevo) (int, int64, error) {
	if n.Envio == "" {
		n.Envio = "retiro"
	}
	if len(n.Items) == 0 || len(n.Items) > 100 || !slices.Contains(pagos, n.Pago) || !slices.Contains(canales, n.Canal) || !slices.Contains(envios, n.Envio) {
		return 0, 0, errors.New("pedido inválido")
	}
	nombre := "Consumidor final"
	if n.Cliente != nil && strings.TrimSpace(*n.Cliente) != "" {
		nombre = strings.TrimSpace(*n.Cliente)
	}
	estado := "Para preparar"
	if n.Canal == "pos" && n.Envio == "retiro" {
		estado = "Entregado"
	}
	estadoPago := map[string]string{"efectivo": "Cobrado", "transf": "Acreditada"}[n.Pago]
	if estadoPago == "" {
		estadoPago = "Aprobado"
	}
	if n.Pagado != nil && !*n.Pagado {
		estado, estadoPago = "Sin pagar", "Pendiente"
	}
	var numero int
	var total int64
	// Se descuenta en un orden fijo (producto, variante): dos ventas con los mismos productos en otro
	// orden no se traban entre sí. Los ítems quedan en el orden en que vinieron.
	orden := make([]int, len(n.Items))
	for i := range orden {
		orden[i] = i
	}
	slices.SortStableFunc(orden, func(a, b int) int {
		x, y := n.Items[a], n.Items[b]
		return strings.Compare(x.ProductoID+":"+x.VarianteID, y.ProductoID+":"+y.VarianteID)
	})
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		items := make([]Item, len(n.Items))
		for _, k := range orden {
			i := n.Items[k]
			if i.Cantidad < 1 || i.Cantidad > 10000 {
				return errors.New("cantidad inválida")
			}
			it, quedan, err := descontar(ctx, t, tiendaID, i)
			if err != nil {
				return err
			}
			if quedan == 0 {
				if err := avisarSinStock(ctx, t, tiendaID, i); err != nil {
					return err
				}
			}
			if p, ok := n.Precios[i.ProductoID+":"+i.VarianteID]; ok && i.VarianteID != "" {
				it.Precio = p
			} else if p, ok := n.Precios[i.ProductoID]; ok {
				it.Precio = p
			}
			it.Cantidad = i.Cantidad
			total += it.Precio * int64(it.Cantidad)
			items[k] = it
		}
		var tel *string
		if n.Telefono != nil && strings.TrimSpace(*n.Telefono) != "" {
			t := strings.TrimSpace(*n.Telefono)
			tel = &t
		}
		var email *string
		if n.Email != "" {
			email = &n.Email
		}
		comprador, _ := json.Marshal(Comprador{Nombre: nombre, Telefono: tel, Email: email, Factura: n.Factura})
		if n.CostoEnvio < 0 || n.Descuento < 0 || n.Descuento > total {
			return errors.New("pedido inválido")
		}
		var codigo *string
		if n.CodigoPublico != "" {
			codigo = &n.CodigoPublico
		}
		js, _ := json.Marshal(items)
		desc, _ := json.Marshal(lineasDescuento(n))
		if _, err := t.Exec(ctx, `SELECT 1 FROM tiendas WHERE id = $1 FOR UPDATE`, tiendaID); err != nil {
			return err
		}
		var cupon *string
		if n.Cupon != "" {
			cupon = &n.Cupon
			tag, err := t.Exec(ctx, `UPDATE cupones SET usos = usos + 1, vendio = vendio + $3, desconto = desconto + $4
				WHERE tienda_id = $1 AND codigo = $2 AND NOT pausado AND (limite = 0 OR usos < limite)`,
				tiendaID, n.Cupon, total+n.CostoEnvio-n.Descuento, n.DescuentoCupon)
			if err != nil {
				return err
			}
			if tag.RowsAffected() == 0 {
				return ErrCupon
			}
			// El tope por cliente se vuelve a mirar acá: con la tienda bloqueada, dos compras a la vez
			// del mismo email no pasan las dos.
			if n.Email != "" {
				var pasado bool
				if err := t.QueryRow(ctx, `SELECT count(p.*) >= CASE c.por_cliente WHEN '1' THEN 1 WHEN '3' THEN 3 ELSE 2147483647 END
					FROM cupones c LEFT JOIN pedidos p ON p.tienda_id = c.tienda_id AND p.cupon = c.codigo
						AND p.estado <> 'Cancelado' AND lower(p.comprador->>'email') = lower($3)
					WHERE c.tienda_id = $1 AND c.codigo = $2 GROUP BY c.por_cliente`, tiendaID, n.Cupon, n.Email).Scan(&pasado); err != nil {
					return err
				}
				if pasado {
					return ErrCupon
				}
			}
		}
		if err := t.QueryRow(ctx, `
			INSERT INTO pedidos (tienda_id, numero, comprador, canal, pago, estado_pago, envio, estado, facturado,
				items, subtotal, costo_envio, total, nota, estados_en, ciudad_envio, codigo_publico, cupon, entrega, descuentos)
			SELECT $1, coalesce(max(numero), 1000) + 1, $2, $3, $4, $5, $6, $7, false, $8, $9, $11::bigint, $9::bigint + $11::bigint - $12::bigint, $10,
				jsonb_build_object($7::text, now()), $13, $14, $15, $16, $17
			FROM pedidos WHERE tienda_id = $1
			RETURNING numero`,
			tiendaID, comprador, n.Canal, n.Pago, estadoPago, n.Envio, estado, js, total, n.Nota,
			n.CostoEnvio, n.Descuento, n.CiudadEnvio, codigo, cupon, n.Entrega, desc).Scan(&numero); err != nil {
			return err
		}
		// Lo que carga alguien del equipo (el panel o una IA): la compra en la tienda online no tiene autor.
		final := total + n.CostoEnvio - n.Descuento
		r := auditoria.Registro{Accion: "pedido.crear", Entidad: "pedido", EntidadID: fmt.Sprint(numero),
			Texto:   fmt.Sprintf("cargó a mano el pedido #%d de %s por %s", numero, nombre, auditoria.Pesos(final)),
			Despues: map[string]any{"total": final, "pago": n.Pago, "canal": n.Canal, "estado": estado}}
		if n.Canal == "pos" {
			r.Accion = "venta.cobro"
			r.Texto = fmt.Sprintf("cobró %s en el punto de venta (venta #%d, %s)", auditoria.Pesos(final), numero, nombrePago[n.Pago])
			if estado == "Sin pagar" {
				r.Texto = fmt.Sprintf("cargó una venta del local por %s sin cobrar (venta #%d)", auditoria.Pesos(final), numero)
			}
		}
		return auditoria.Registrar(ctx, t, tiendaID, r)
	})
	return numero, total + n.CostoEnvio - n.Descuento, err
}

// lineasDescuento: cada descuento del pedido por separado (el cupón y el de la transferencia), para
// que el detalle y el mail cierren: subtotal + envío − descuentos = total.
func lineasDescuento(n PedidoNuevo) []Descuento {
	out := []Descuento{}
	if n.Cupon != "" && n.DescuentoCupon > 0 {
		out = append(out, Descuento{Concepto: "Cupón " + n.Cupon, Monto: n.DescuentoCupon})
	}
	if resto := n.Descuento - n.DescuentoCupon; resto > 0 {
		concepto := "Descuento"
		if n.Pago == "transf" {
			concepto = "Descuento por transferencia"
		}
		out = append(out, Descuento{Concepto: concepto, Monto: resto})
	}
	return out
}

// descontar baja el stock de la línea sólo si alcanza (la condición del UPDATE evita vender lo que no
// hay con dos cajas a la vez) y devuelve el ítem con el precio y la foto de la tienda, y cuántas
// unidades quedan (del producto o de la variante).
func descontar(ctx context.Context, t pgx.Tx, tiendaID string, i Linea) (Item, int, error) {
	it := Item{ProductoID: &i.ProductoID}
	var quedan int
	if i.VarianteID == "" {
		err := t.QueryRow(ctx, `UPDATE productos SET stock = stock - $3
			WHERE tienda_id = $1 AND id::text = $2 AND stock >= $3 AND opciones = '[]'::jsonb
			RETURNING nombre, foto, precio, stock`, tiendaID, i.ProductoID, i.Cantidad).Scan(&it.Nombre, &it.Foto, &it.Precio, &quedan)
		if !errors.Is(err, pgx.ErrNoRows) {
			return it, quedan, err
		}
		var conOpciones *bool
		_ = t.QueryRow(ctx, `SELECT opciones <> '[]'::jsonb FROM productos WHERE tienda_id = $1 AND id::text = $2`,
			tiendaID, i.ProductoID).Scan(&conOpciones)
		switch {
		case conOpciones == nil:
			return it, 0, ErrProducto
		case *conOpciones:
			return it, 0, ErrSinVariante
		}
		return it, 0, ErrSinStock
	}
	// La variante: activa, de ese producto y de esta tienda (RLS además lo asegura).
	var opciones []listas.Opcion
	var valores map[string]string
	var id string
	err := t.QueryRow(ctx, `UPDATE variantes v SET stock = v.stock - $4
		FROM productos p
		WHERE v.tienda_id = $1 AND v.producto_id::text = $2 AND v.id::text = $3 AND v.activa AND v.stock >= $4
			AND p.tienda_id = v.tienda_id AND p.id = v.producto_id AND p.opciones <> '[]'::jsonb
		RETURNING p.nombre, coalesce(v.foto, p.foto), coalesce(v.precio, p.precio), v.id::text, v.valores, p.opciones, v.stock`,
		tiendaID, i.ProductoID, i.VarianteID, i.Cantidad).Scan(&it.Nombre, &it.Foto, &it.Precio, &id, &valores, &opciones, &quedan)
	if errors.Is(err, pgx.ErrNoRows) {
		var activa *bool
		_ = t.QueryRow(ctx, `SELECT v.activa FROM variantes v JOIN productos p ON p.tienda_id = v.tienda_id AND p.id = v.producto_id
			WHERE v.tienda_id = $1 AND v.producto_id::text = $2 AND v.id::text = $3 AND p.opciones <> '[]'::jsonb`,
			tiendaID, i.ProductoID, i.VarianteID).Scan(&activa)
		if activa != nil && *activa {
			return it, 0, ErrSinStock
		}
		return it, 0, ErrVariante
	}
	if err != nil {
		return it, 0, err
	}
	texto := listas.TextoVariante(opciones, valores)
	it.VarianteID, it.Variante = &id, &texto
	return it, quedan, nil
}

func rutasNuevo(mux *http.ServeMux, pool tx, con func(func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc) {
	mux.HandleFunc("POST /v1/panel/pedidos", con(func(w http.ResponseWriter, r *http.Request, id string) {
		var n PedidoNuevo
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&n); err != nil {
			problema.Escribir(w, problema.Invalido("Mandá los productos del pedido."))
			return
		}
		numero, total, err := CrearDesdeProductos(r.Context(), pool, id, n)
		switch {
		case errors.Is(err, ErrSinStock):
			problema.Escribir(w, problema.Conflicto("sin_stock", "No alcanza el stock de algún producto."))
		case errors.Is(err, ErrProducto):
			problema.Escribir(w, problema.Invalido("Hay un producto que no existe."))
		case errors.Is(err, ErrSinVariante):
			problema.Escribir(w, problema.Invalido("Elegí la variante de los productos con opciones (talle, color…)."))
		case errors.Is(err, ErrVariante):
			problema.Escribir(w, problema.Invalido("Hay una variante que no existe o no está a la venta."))
		case err != nil && strings.Contains(err.Error(), "inválid"):
			problema.Escribir(w, problema.Invalido("Revisá el pedido."))
		case err != nil:
			problema.Escribir(w, problema.Interno())
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{"numero": numero, "total": total})
		}
	}))
}
