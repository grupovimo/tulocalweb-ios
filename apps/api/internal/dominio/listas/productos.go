package listas

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/auditoria"
)

// ErrNoExiste: el producto no es de esta tienda.
var ErrNoExiste = errors.New("el producto no existe")

// Editable es lo que se carga en el editor (components/schemas/ProductoEditable).
type Editable struct {
	Nombre      string  `json:"nombre"`
	Descripcion *string `json:"descripcion"`
	Categoria   *string `json:"categoria"`
	Precio      int64   `json:"precio"`
	// PrecioMayorista (§49): nil deja el que había; 0 lo saca. Menor que el precio.
	PrecioMayorista *int64          `json:"precioMayorista"`
	Costo           *int64          `json:"costo"`
	SKU             *string         `json:"sku"`
	Stock           *int            `json:"stock"`
	Foto            *string         `json:"foto"`
	Canales         map[string]bool `json:"canales"`
	// Fotos es la galería (la primera es la portada). Sin mandarla, manda Foto: reemplaza la portada
	// y deja el resto (o vacía la galería si Foto es null).
	Fotos []string `json:"fotos"`
	// Opciones nil deja las que había; [] las saca (y con ellas las variantes).
	Opciones []Opcion `json:"opciones"`
	// Variantes: sólo las combinaciones que se quieren tocar; las demás se conservan o se generan.
	Variantes []VarianteEditable `json:"variantes"`
	// SinPrecios: quien guarda no puede tocar precios ni costos (rol Ventas, docs/decisiones.md §37):
	// quedan los que había, también los de las variantes. No viene del cliente.
	SinPrecios bool `json:"-"`
}

// Producto es la ficha (components/schemas/Producto).
type Producto struct {
	ID          string  `json:"id"`
	Nombre      string  `json:"nombre"`
	Descripcion string  `json:"descripcion"`
	Categoria   *string `json:"categoria"`
	Precio      int64   `json:"precio"`
	// PrecioMayorista: null si no tiene.
	PrecioMayorista *int64          `json:"precioMayorista"`
	Costo           int64           `json:"costo"`
	SKU             string          `json:"sku"`
	Stock           int             `json:"stock"`
	Foto            *string         `json:"foto"`
	Canales         map[string]bool `json:"canales"`
	Publicado       bool            `json:"publicado"`
	Vendidos30d     int             `json:"vendidos30d"`
	Slug            string          `json:"slug"`
	Fotos           []string        `json:"fotos"`
	Opciones        []Opcion        `json:"opciones"`
	Variantes       []Variante      `json:"variantes"`
}

// Validar revisa lo que se carga.
func (e Editable) Validar() error {
	n := len([]rune(strings.TrimSpace(e.Nombre)))
	if n < 2 || n > 120 {
		return errors.New("el nombre tiene que tener entre 2 y 120 letras")
	}
	if e.Precio < 0 || (e.Costo != nil && *e.Costo < 0) || (e.Stock != nil && *e.Stock < 0) {
		return errors.New("precio, costo y stock no pueden ser negativos")
	}
	if m := e.PrecioMayorista; m != nil && !e.SinPrecios && (*m < 0 || (*m > 0 && *m >= e.Precio)) {
		return errors.New("el precio mayorista tiene que ser menor que el precio")
	}
	return nil
}

func valor[T any](p *T, def T) T {
	if p == nil {
		return def
	}
	return *p
}

// Guardar crea (id vacío) o actualiza un producto y devuelve su ficha. Con opciones, genera o ajusta
// las variantes y el stock del producto pasa a ser la suma de las activas.
func Guardar(ctx context.Context, pool tx, tiendaID, id string, e Editable) (Producto, error) {
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var err error
		id, err = GuardarEnTx(ctx, t, tiendaID, id, e)
		return err
	})
	if err != nil {
		return Producto{}, err
	}
	return Obtener(ctx, pool, tiendaID, id, time.Now())
}

// GuardarEnTx es Guardar dentro de una transacción con la tienda fijada, sin armar la ficha (que lee
// todo el catálogo): lo usa la importación para guardar un lote con las mismas reglas que el panel.
// Devuelve el id.
func GuardarEnTx(ctx context.Context, t pgx.Tx, tiendaID, id string, e Editable) (string, error) {
	if err := e.Validar(); err != nil {
		return "", err
	}
	canales, _ := json.Marshal(e.Canales)
	if e.Canales == nil {
		canales = []byte(`{"web": true}`)
	}
	var fotosPedidas []string
	if e.Fotos != nil {
		var err error
		if fotosPedidas, err = limpiarFotos(e.Fotos); err != nil {
			return "", err
		}
	}
	var opcionesPedidas []Opcion
	if e.Opciones != nil {
		var err error
		if opcionesPedidas, err = limpiarOpciones(e.Opciones); err != nil {
			return "", err
		}
	}
	opciones, fotos := []Opcion{}, []string{}
	var antes antesProducto
	if id != "" {
		err := t.QueryRow(ctx, `SELECT opciones, fotos, nombre, precio, costo, stock, publicado
			FROM productos WHERE tienda_id = $1 AND id::text = $2 FOR UPDATE`,
			tiendaID, id).Scan(&opciones, &fotos, &antes.nombre, &antes.precio, &antes.costo, &antes.stock, &antes.publicado)
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNoExiste
		}
		if err != nil {
			return "", err
		}
	}
	switch {
	case e.Fotos != nil:
		fotos = fotosPedidas
	case e.Foto == nil:
		fotos = []string{}
	default:
		resto := []string{}
		for i, f := range fotos {
			if i > 0 && f != *e.Foto {
				resto = append(resto, f)
			}
		}
		fotos = append([]string{*e.Foto}, resto...)
	}
	var foto *string
	if len(fotos) > 0 {
		foto = &fotos[0]
	}
	if e.Opciones != nil {
		opciones = opcionesPedidas
	}
	if len(e.Variantes) > 0 && len(opciones) == 0 {
		return "", errors.New("para cargar variantes, el producto tiene que tener opciones")
	}
	// Con opciones, el stock lo pone la suma de las variantes (no el que manda el editor).
	stock := e.Stock
	if len(opciones) > 0 {
		stock = nil
	}
	if id == "" {
		var mayor *int64
		if e.PrecioMayorista != nil && *e.PrecioMayorista > 0 && !e.SinPrecios {
			mayor = e.PrecioMayorista
		}
		if err := t.QueryRow(ctx, `INSERT INTO productos (tienda_id, nombre, descripcion, categoria, precio, costo, sku, stock, foto, canales, fotos, opciones, precio_mayorista)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) RETURNING id`,
			tiendaID, strings.TrimSpace(e.Nombre), valor(e.Descripcion, ""), e.Categoria, e.Precio, valor(e.Costo, 0),
			valor(e.SKU, ""), valor(stock, 0), foto, canales, fotos, opciones, mayor).Scan(&id); err != nil {
			return "", err
		}
		nombre := strings.TrimSpace(e.Nombre)
		if err := auditoria.Registrar(ctx, t, tiendaID, auditoria.Registro{Accion: "producto.crear", Entidad: "producto",
			EntidadID: id, Texto: fmt.Sprintf("creó el producto %s a %s", nombre, auditoria.Pesos(e.Precio)),
			Despues: map[string]any{"nombre": nombre, "precio": e.Precio}}); err != nil {
			return "", err
		}
	} else if _, err := t.Exec(ctx, `UPDATE productos SET nombre = $3, descripcion = coalesce($4, descripcion),
			categoria = $5, precio = CASE WHEN $14 THEN precio ELSE $6 END,
			costo = CASE WHEN $14 THEN costo ELSE coalesce($7, costo) END, sku = coalesce($8, sku),
			stock = coalesce($9, stock), foto = $10, canales = $11, publicado = true, fotos = $12, opciones = $13,
			precio_mayorista = CASE WHEN $14 OR $15::bigint IS NULL THEN precio_mayorista ELSE nullif($15::bigint, 0) END
		WHERE tienda_id = $1 AND id::text = $2`,
		tiendaID, id, strings.TrimSpace(e.Nombre), e.Descripcion, e.Categoria, e.Precio, e.Costo, e.SKU, stock, foto, canales,
		fotos, opciones, e.SinPrecios, e.PrecioMayorista); err != nil {
		return "", err
	} else if err := registrarEdicion(ctx, t, tiendaID, id, antes, e, len(opciones) == 0); err != nil {
		return "", err
	}
	if err := sincronizarVariantes(ctx, t, tiendaID, id, opciones, fotos, e.Variantes, e.SinPrecios); err != nil {
		return "", err
	}
	if len(opciones) == 0 {
		return id, nil
	}
	// El trigger suma con cada cambio de variante; esto cubre el caso de opciones nuevas sin cambios.
	_, err := t.Exec(ctx, `UPDATE productos SET stock = (SELECT coalesce(sum(stock), 0) FROM variantes
			WHERE tienda_id = $1 AND producto_id::text = $2 AND activa)
		WHERE tienda_id = $1 AND id::text = $2`, tiendaID, id)
	return id, err
}

// Obtener devuelve la ficha de un producto.
func Obtener(ctx context.Context, pool tx, tiendaID, id string, ahora time.Time) (Producto, error) {
	regs, err := Productos(ctx, pool, tiendaID, ahora)
	if err != nil {
		return Producto{}, err
	}
	for _, r := range regs {
		if r["id"] == id {
			p := r["_ficha"].(Producto)
			return p, nil
		}
	}
	return Producto{}, ErrNoExiste
}

// Productos arma la lista con lo vendido (por nombre, de los pedidos de las últimas 8 semanas).
func Productos(ctx context.Context, pool tx, tiendaID string, ahora time.Time) ([]Registro, error) {
	regs := []Registro{}
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		semanas := map[string][]int{}
		ultimos30 := map[string]int{}
		v, err := t.Query(ctx, `SELECT i->>'nombre', (i->>'cantidad')::int, creado_en
			FROM pedidos, jsonb_array_elements(items) i
			WHERE tienda_id = $1 AND estado <> 'Cancelado' AND creado_en > $2`, tiendaID, ahora.Add(-56*24*time.Hour))
		if err != nil {
			return err
		}
		for v.Next() {
			var n string
			var c int
			var cu time.Time
			if err := v.Scan(&n, &c, &cu); err != nil {
				return err
			}
			if semanas[n] == nil {
				semanas[n] = make([]int, 8)
			}
			if k := 7 - int(ahora.Sub(cu).Hours()/24/7); k >= 0 && k < 8 {
				semanas[n][k] += c
			}
			if ahora.Sub(cu) < 30*24*time.Hour {
				ultimos30[n] += c
			}
		}
		if err := v.Err(); err != nil {
			return err
		}
		porProducto, err := variantesDe(ctx, t, tiendaID)
		if err != nil {
			return err
		}
		filas, err := t.Query(ctx, `SELECT id, nombre, descripcion, categoria, precio, costo, sku, stock, foto, canales, publicado, fotos, opciones, slug,
				precio_mayorista
			FROM productos WHERE tienda_id = $1 ORDER BY creado_en DESC`, tiendaID)
		if err != nil {
			return err
		}
		for filas.Next() {
			var p Producto
			if err := filas.Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.Categoria, &p.Precio, &p.Costo, &p.SKU, &p.Stock, &p.Foto, &p.Canales, &p.Publicado, &p.Fotos, &p.Opciones, &p.Slug,
				&p.PrecioMayorista); err != nil {
				return err
			}
			p.Variantes = porProducto[p.ID]
			if p.Variantes == nil {
				p.Variantes = []Variante{}
			}
			OrdenarVariantes(p.Opciones, p.Variantes)
			resumen := make([]Registro, 0, len(p.Variantes))
			for _, v := range p.Variantes {
				resumen = append(resumen, Registro{"id": v.ID, "texto": TextoVariante(p.Opciones, v.Valores), "sku": v.SKU,
					"precio": v.Precio, "precioMayorista": v.PrecioMayorista, "stock": v.Stock, "activa": v.Activa})
			}
			p.Vendidos30d = ultimos30[p.Nombre]
			ventas := semanas[p.Nombre]
			if ventas == nil {
				ventas = make([]int, 8)
			}
			// Sin costo cargado el margen no se sabe (null), no es 100 %.
			var margen any
			lowm := false
			if p.Precio > 0 && p.Costo > 0 {
				m := int((p.Precio - p.Costo) * 100 / p.Precio)
				margen, lowm = m, m < 25
			}
			canales := []string{}
			for k, on := range p.Canales {
				if on {
					canales = append(canales, k)
				}
			}
			estado := "active"
			switch {
			case !p.Publicado && p.Foto == nil:
				estado = "draft"
			case !p.Publicado:
				estado = "paused"
			case p.Stock == 0:
				estado = "out"
			}
			cat := "Sin categoría"
			if p.Categoria != nil && *p.Categoria != "" {
				cat = *p.Categoria
			}
			regs = append(regs, Registro{
				"id": p.ID, "nombre": p.Nombre, "descripcion": p.Descripcion, "foto": p.Foto, "sku": p.SKU,
				"categoria": cat, "precio": p.Precio, "precioMayorista": p.PrecioMayorista, "costo": p.Costo,
				"margen": margen, "stock": p.Stock,
				"ventas": ventas, "vendidos30": p.Vendidos30d, "canales": canales, "publicado": p.Publicado,
				"estado": estado, "lowm": lowm, "star": p.Vendidos30d > 60, "variantes": resumen, "slug": p.Slug, "_ficha": p,
			})
		}
		return filas.Err()
	})
	return regs, err
}

// DefProductos: cómo filtra y ordena Productos.
var DefProductos = Def{
	Buscar: func(r Registro) string {
		return r["nombre"].(string) + " " + r["sku"].(string) + " " + r["categoria"].(string) + skusDe(r)
	},
	Segmento: "estado", Facetas: []string{"categoria", "canales"}, Rango: "precio", Banderas: []string{"lowm", "star"},
	Ordenes: []Orden{
		{"sold", func(a, b Registro) bool { return numero(a["vendidos30"]) > numero(b["vendidos30"]) }},
		{"stock", func(a, b Registro) bool { return numero(a["stock"]) < numero(b["stock"]) }},
		{"price", func(a, b Registro) bool { return numero(a["precio"]) > numero(b["precio"]) }},
		{"margin", func(a, b Registro) bool {
			// Los que no tienen costo (margen sin dato) van al final.
			if a["margen"] == nil || b["margen"] == nil {
				return a["margen"] != nil && b["margen"] == nil
			}
			return numero(a["margen"]) > numero(b["margen"])
		}},
	},
	Resumen: func(todos, l []Registro) map[string]float64 {
		r := map[string]float64{"precioMax": 10000}
		for _, x := range l {
			r["unidades30"] += numero(x["vendidos30"])
		}
		for _, x := range todos {
			r["precioMax"] = max(r["precioMax"], numero(x["precio"]))
			cs, _ := x["canales"].([]string)
			for _, c := range cs {
				if c == "ml" && x["estado"] == "out" {
					r["sinStockEnMl"]++
				}
			}
		}
		return r
	},
}

// Publicar prende o apaga productos; Precio los sube o baja un porcentaje (redondeo a $ 10).
func Publicar(ctx context.Context, pool tx, tiendaID string, ids []string, on bool) (int, error) {
	var n int64
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		filas, err := t.Query(ctx, `UPDATE productos p SET publicado = $3
			FROM productos v WHERE p.tienda_id = $1 AND p.id::text = ANY($2) AND v.tienda_id = p.tienda_id AND v.id = p.id
			RETURNING p.id::text, p.nombre, v.publicado`, tiendaID, ids, on)
		if err != nil {
			return err
		}
		tocados, err := pgx.CollectRows(filas, pgx.RowToStructByPos[publicacion])
		if err != nil {
			return err
		}
		n = int64(len(tocados))
		accion, verbo := "producto.pausar", "pausó "
		if on {
			accion, verbo = "producto.publicar", "publicó "
		}
		rs := make([]auditoria.Registro, 0, len(tocados))
		for _, c := range tocados {
			if c.Antes != on {
				rs = append(rs, auditoria.Registro{Accion: accion, Entidad: "producto", EntidadID: c.ID, Texto: verbo + c.Nombre,
					Antes: map[string]bool{"publicado": !on}, Despues: map[string]bool{"publicado": on}})
			}
		}
		return auditoria.Registrar(ctx, t, tiendaID, rs...)
	})
	return int(n), err
}

// Precio cambia el precio un porcentaje.
func Precio(ctx context.Context, pool tx, tiendaID string, ids []string, pct float64) (int, error) {
	if pct <= -90 || pct > 200 {
		return 0, errors.New("porcentaje fuera de rango")
	}
	var n int64
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		// El precio mayorista (§49) acompaña: sube o baja lo mismo.
		filas, err := t.Query(ctx, `UPDATE productos p SET precio = greatest(0, round(p.precio * (1 + $3 / 100.0) / 10) * 10),
				precio_mayorista = CASE WHEN p.precio_mayorista IS NOT NULL
					THEN greatest(1, round(p.precio_mayorista * (1 + $3 / 100.0) / 10) * 10) END
			FROM productos v WHERE p.tienda_id = $1 AND p.id::text = ANY($2) AND v.tienda_id = p.tienda_id AND v.id = p.id
			RETURNING p.id::text, p.nombre, v.precio, p.precio`, tiendaID, ids, pct)
		if err != nil {
			return err
		}
		cambios, err := pgx.CollectRows(filas, pgx.RowToStructByPos[cambioPrecio])
		if err != nil {
			return err
		}
		n = int64(len(cambios))
		rs := make([]auditoria.Registro, 0, len(cambios))
		for _, c := range cambios {
			if c.Antes != c.Despues {
				rs = append(rs, auditoria.Registro{Accion: "producto.precio", Entidad: "producto", EntidadID: c.ID,
					Texto: fmt.Sprintf("cambió el precio de %s de %s a %s", c.Nombre, auditoria.Pesos(c.Antes), auditoria.Pesos(c.Despues)),
					Antes: map[string]int64{"precio": c.Antes}, Despues: map[string]int64{"precio": c.Despues}})
			}
		}
		if err := auditoria.Registrar(ctx, t, tiendaID, rs...); err != nil {
			return err
		}
		// Las variantes con precio propio suben (o bajan) igual.
		// (greatest ignora los NULL: sin el CASE, un precio vacío pasaría a 0).
		_, err = t.Exec(ctx, `UPDATE variantes SET
				precio = CASE WHEN precio IS NOT NULL THEN greatest(0, round(precio * (1 + $3 / 100.0) / 10) * 10) END,
				precio_mayorista = CASE WHEN precio_mayorista IS NOT NULL
					THEN greatest(1, round(precio_mayorista * (1 + $3 / 100.0) / 10) * 10) END
			WHERE tienda_id = $1 AND producto_id::text = ANY($2) AND (precio IS NOT NULL OR precio_mayorista IS NOT NULL)`, tiendaID, ids, pct)
		return err
	})
	return int(n), err
}

// skusDe suma los SKU de las variantes a la búsqueda.
func skusDe(r Registro) string {
	vs, _ := r["variantes"].([]Registro)
	var b strings.Builder
	for _, v := range vs {
		if s, _ := v["sku"].(string); s != "" {
			b.WriteString(" " + s)
		}
	}
	return b.String()
}
