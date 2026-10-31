package listas

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5"
)

// Límites de opciones, variantes y galería (los mismos que valida la web).
const (
	MaxOpciones  = 3
	MaxValores   = 20
	MaxVariantes = 100
	MaxFotos     = 8
)

// Opcion es una opción del producto: "Talle" con sus valores ("S", "M", "L").
type Opcion struct {
	Nombre  string   `json:"nombre"`
	Valores []string `json:"valores"`
}

// Variante es una combinación de valores, con su stock (components/schemas/Variante).
type Variante struct {
	ID      string            `json:"id"`
	Valores map[string]string `json:"valores"`
	SKU     string            `json:"sku"`
	// Precio nil: usa el del producto.
	Precio *int64 `json:"precio"`
	// PrecioMayorista nil: usa el mayorista del producto (si tiene). §49.
	PrecioMayorista *int64  `json:"precioMayorista"`
	Stock           int     `json:"stock"`
	Foto            *string `json:"foto"`
	Activa          bool    `json:"activa"`
}

// VarianteEditable es lo que manda el editor por cada combinación que quiere tocar. Se identifica
// por sus valores (el id se ignora). Stock y activa en nil dejan lo que había (o 0 y true si es nueva).
type VarianteEditable struct {
	Valores map[string]string `json:"valores"`
	SKU     *string           `json:"sku"`
	Precio  *int64            `json:"precio"`
	// PrecioMayorista pisa como Precio: nil usa el mayorista del producto.
	PrecioMayorista *int64  `json:"precioMayorista"`
	Stock           *int    `json:"stock"`
	Foto            *string `json:"foto"`
	Activa          *bool   `json:"activa"`
}

// TextoVariante es lo que se lee en el pedido: "Talle M · Color Negro" (en el orden de las opciones).
func TextoVariante(opciones []Opcion, valores map[string]string) string {
	partes := make([]string, 0, len(opciones))
	for _, o := range opciones {
		if v, ok := valores[o.Nombre]; ok {
			partes = append(partes, o.Nombre+" "+v)
		}
	}
	return strings.Join(partes, " · ")
}

// clave identifica una combinación según las opciones (vacía si los valores no les corresponden).
func clave(opciones []Opcion, valores map[string]string) string {
	if len(valores) != len(opciones) {
		return ""
	}
	partes := make([]string, 0, len(opciones))
	for _, o := range opciones {
		v, ok := valores[o.Nombre]
		if !ok || !slices.Contains(o.Valores, v) {
			return ""
		}
		partes = append(partes, o.Nombre+"\x00"+v)
	}
	return strings.Join(partes, "\x01")
}

// combinaciones arma todas las combinaciones, la primera opción por fuera.
func combinaciones(opciones []Opcion) []map[string]string {
	if len(opciones) == 0 {
		return nil
	}
	out := []map[string]string{{}}
	for _, o := range opciones {
		sig := make([]map[string]string, 0, len(out)*len(o.Valores))
		for _, c := range out {
			for _, v := range o.Valores {
				n := make(map[string]string, len(c)+1)
				for k, x := range c {
					n[k] = x
				}
				n[o.Nombre] = v
				sig = append(sig, n)
			}
		}
		out = sig
	}
	return out
}

// OrdenarVariantes deja las variantes en el orden de las combinaciones (como las arma el editor).
func OrdenarVariantes(opciones []Opcion, vs []Variante) {
	pos := map[string]int{}
	for i, c := range combinaciones(opciones) {
		pos[clave(opciones, c)] = i
	}
	slices.SortStableFunc(vs, func(a, b Variante) int {
		return pos[clave(opciones, a.Valores)] - pos[clave(opciones, b.Valores)]
	})
}

func largo(s string) int { return len([]rune(s)) }

// limpiarOpciones recorta espacios y valida: hasta 3 opciones de 1 a 20 valores, sin repetidos.
func limpiarOpciones(os []Opcion) ([]Opcion, error) {
	if len(os) > MaxOpciones {
		return nil, fmt.Errorf("un producto puede tener hasta %d opciones", MaxOpciones)
	}
	out := make([]Opcion, 0, len(os))
	nombres := map[string]bool{}
	total := 1
	for _, o := range os {
		n := strings.TrimSpace(o.Nombre)
		if largo(n) < 1 || largo(n) > 20 {
			return nil, errors.New("el nombre de cada opción tiene que tener entre 1 y 20 letras")
		}
		if nombres[normal(n)] {
			return nil, fmt.Errorf("la opción %q está repetida", n)
		}
		nombres[normal(n)] = true
		if len(o.Valores) < 1 || len(o.Valores) > MaxValores {
			return nil, fmt.Errorf("%s tiene que tener entre 1 y %d valores", n, MaxValores)
		}
		vistos := map[string]bool{}
		vals := make([]string, 0, len(o.Valores))
		for _, v := range o.Valores {
			v = strings.TrimSpace(v)
			if largo(v) < 1 || largo(v) > 20 {
				return nil, fmt.Errorf("los valores de %s tienen que tener entre 1 y 20 letras", n)
			}
			if vistos[normal(v)] {
				return nil, fmt.Errorf("%s tiene el valor %q repetido", n, v)
			}
			vistos[normal(v)] = true
			vals = append(vals, v)
		}
		total *= len(vals)
		out = append(out, Opcion{Nombre: n, Valores: vals})
	}
	if total > MaxVariantes {
		return nil, fmt.Errorf("esas opciones arman %d variantes: el máximo es %d", total, MaxVariantes)
	}
	return out, nil
}

// limpiarFotos: hasta 8, sin vacías ni repetidas.
func limpiarFotos(fs []string) ([]string, error) {
	if len(fs) > MaxFotos {
		return nil, fmt.Errorf("la galería admite hasta %d fotos", MaxFotos)
	}
	out := make([]string, 0, len(fs))
	for _, f := range fs {
		f = strings.TrimSpace(f)
		if f == "" || len(f) > 600 {
			return nil, errors.New("hay una foto inválida")
		}
		if !slices.Contains(out, f) {
			out = append(out, f)
		}
	}
	return out, nil
}

// sincronizarVariantes deja en la base una variante por combinación: las que siguen conservan stock,
// precio, SKU y foto; las nuevas arrancan con stock 0 y activas; las que ya no existen se borran
// (los pedidos guardaron el texto). Lo que manda el editor pisa lo que había.
func sincronizarVariantes(ctx context.Context, t pgx.Tx, tiendaID, productoID string, opciones []Opcion, fotos []string,
	editadas []VarianteEditable, sinPrecios bool) error {
	pedidas := map[string]VarianteEditable{}
	for _, e := range editadas {
		k := clave(opciones, e.Valores)
		if k == "" {
			return errors.New("hay una variante que no corresponde a las opciones")
		}
		if _, ok := pedidas[k]; ok {
			return errors.New("hay una variante repetida")
		}
		if (e.Precio != nil && *e.Precio < 0) || (e.PrecioMayorista != nil && *e.PrecioMayorista < 0) || (e.Stock != nil && (*e.Stock < 0 || *e.Stock > 1000000)) {
			return errors.New("precio y stock de las variantes no pueden ser negativos")
		}
		if e.SKU != nil && largo(strings.TrimSpace(*e.SKU)) > 40 {
			return errors.New("el SKU de una variante tiene hasta 40 letras")
		}
		if e.Foto != nil && !slices.Contains(fotos, *e.Foto) {
			return errors.New("la foto de una variante tiene que ser una de las del producto")
		}
		pedidas[k] = e
	}
	existentes := map[string]Variante{}
	filas, err := t.Query(ctx, `SELECT id::text, valores, sku, precio, precio_mayorista, stock, foto, activa FROM variantes
		WHERE tienda_id = $1 AND producto_id::text = $2`, tiendaID, productoID)
	if err != nil {
		return err
	}
	var borrar []string
	for filas.Next() {
		var v Variante
		if err := filas.Scan(&v.ID, &v.Valores, &v.SKU, &v.Precio, &v.PrecioMayorista, &v.Stock, &v.Foto, &v.Activa); err != nil {
			filas.Close()
			return err
		}
		if k := clave(opciones, v.Valores); k != "" {
			existentes[k] = v
		} else {
			borrar = append(borrar, v.ID)
		}
	}
	if err := filas.Err(); err != nil {
		return err
	}
	if len(borrar) > 0 {
		if _, err := t.Exec(ctx, `DELETE FROM variantes WHERE tienda_id = $1 AND producto_id::text = $2 AND id::text = ANY($3)`,
			tiendaID, productoID, borrar); err != nil {
			return err
		}
	}
	for _, c := range combinaciones(opciones) {
		k := clave(opciones, c)
		v, ya := existentes[k]
		if !ya {
			v = Variante{Valores: c, Activa: true}
		}
		antes := v
		if v.Foto != nil && !slices.Contains(fotos, *v.Foto) {
			v.Foto = nil // la foto salió de la galería
		}
		if e, ok := pedidas[k]; ok {
			if e.SKU != nil {
				v.SKU = strings.TrimSpace(*e.SKU)
			}
			if !sinPrecios {
				v.Precio = e.Precio
				v.PrecioMayorista = e.PrecioMayorista
				if v.PrecioMayorista != nil && *v.PrecioMayorista == 0 {
					v.PrecioMayorista = nil
				}
			}
			v.Foto = e.Foto
			if e.Stock != nil {
				v.Stock = *e.Stock
			}
			if e.Activa != nil {
				v.Activa = *e.Activa
			}
		}
		if !ya {
			if _, err := t.Exec(ctx, `INSERT INTO variantes (tienda_id, producto_id, valores, sku, precio, stock, foto, activa, precio_mayorista)
				VALUES ($1, $2::uuid, $3, $4, $5, $6, $7, $8, $9)`, tiendaID, productoID, c, v.SKU, v.Precio, v.Stock, v.Foto, v.Activa,
				v.PrecioMayorista); err != nil {
				return err
			}
			continue
		}
		if igualVariante(antes, v) {
			continue
		}
		if _, err := t.Exec(ctx, `UPDATE variantes SET sku = $4, precio = $5, stock = $6, foto = $7, activa = $8, precio_mayorista = $9
			WHERE tienda_id = $1 AND producto_id::text = $2 AND id::text = $3`,
			tiendaID, productoID, v.ID, v.SKU, v.Precio, v.Stock, v.Foto, v.Activa, v.PrecioMayorista); err != nil {
			return err
		}
	}
	return nil
}

func igualVariante(a, b Variante) bool {
	eq := func(x, y *int64) bool { return (x == nil) == (y == nil) && (x == nil || *x == *y) }
	eqs := func(x, y *string) bool { return (x == nil) == (y == nil) && (x == nil || *x == *y) }
	return a.SKU == b.SKU && eq(a.Precio, b.Precio) && eq(a.PrecioMayorista, b.PrecioMayorista) && a.Stock == b.Stock && eqs(a.Foto, b.Foto) && a.Activa == b.Activa
}

// variantesDe trae las variantes de la tienda agrupadas por producto.
func variantesDe(ctx context.Context, t pgx.Tx, tiendaID string) (map[string][]Variante, error) {
	por := map[string][]Variante{}
	filas, err := t.Query(ctx, `SELECT producto_id::text, id::text, valores, sku, precio, precio_mayorista, stock, foto, activa
		FROM variantes WHERE tienda_id = $1`, tiendaID)
	if err != nil {
		return nil, err
	}
	defer filas.Close()
	for filas.Next() {
		var p string
		var v Variante
		if err := filas.Scan(&p, &v.ID, &v.Valores, &v.SKU, &v.Precio, &v.PrecioMayorista, &v.Stock, &v.Foto, &v.Activa); err != nil {
			return nil, err
		}
		por[p] = append(por[p], v)
	}
	return por, filas.Err()
}

// separarIDs: en Inventario una fila de variante se identifica "producto:variante".
func separarIDs(ids []string) (productos, variantes []string) {
	for _, id := range ids {
		if strings.Contains(id, ":") {
			variantes = append(variantes, id)
		} else {
			productos = append(productos, id)
		}
	}
	return
}
