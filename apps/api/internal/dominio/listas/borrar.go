package listas

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/auditoria"
)

var esID = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// ErrEnMercadoLibre: el producto tiene una publicación en Mercado Libre; borrarlo la dejaría viva.
var ErrEnMercadoLibre = errors.New("el producto está publicado en Mercado Libre")

// Borrar saca el producto del catálogo.
func Borrar(ctx context.Context, pool tx, tiendaID, id string) error {
	if !esID.MatchString(id) {
		return ErrNoExiste
	}
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var enMl bool
		if err := t.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM ml_publicaciones WHERE tienda_id = $1 AND producto_id = $2)`,
			tiendaID, id).Scan(&enMl); err != nil {
			return err
		}
		if enMl {
			return ErrEnMercadoLibre
		}
		var nombre string
		var precio int64
		err := t.QueryRow(ctx, `DELETE FROM productos WHERE tienda_id = $1 AND id = $2 RETURNING nombre, precio`, tiendaID, id).
			Scan(&nombre, &precio)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoExiste
		}
		if err != nil {
			return err
		}
		return auditoria.Registrar(ctx, t, tiendaID, auditoria.Registro{Accion: "producto.borrar", Entidad: "producto",
			EntidadID: id, Texto: "borró el producto " + nombre, Antes: map[string]any{"nombre": nombre, "precio": precio}})
	})
}

// Duplicar crea una copia pausada, con las mismas fotos, opciones y precios, sin stock ni SKU (son
// de cada producto).
func Duplicar(ctx context.Context, pool tx, tiendaID, id string, ahora time.Time) (Producto, error) {
	p, err := Obtener(ctx, pool, tiendaID, id, ahora)
	if err != nil {
		return Producto{}, err
	}
	nombre := []rune(p.Nombre + " (copia)")
	if len(nombre) > 120 {
		nombre = nombre[:120]
	}
	cero, vacio := 0, ""
	e := Editable{Nombre: string(nombre), Descripcion: &p.Descripcion, Categoria: p.Categoria, Precio: p.Precio,
		Costo: &p.Costo, SKU: &vacio, Stock: &cero, Foto: p.Foto, Canales: map[string]bool{"web": true}, Fotos: p.Fotos,
		Opciones: p.Opciones, Tipo: p.Tipo, Atributos: p.Atributos}
	for _, v := range p.Variantes {
		e.Variantes = append(e.Variantes, VarianteEditable{Valores: v.Valores, SKU: &vacio, Precio: v.Precio, Stock: &cero,
			Foto: v.Foto, Activa: &v.Activa})
	}
	// La copia no se anota como «creó»: se anota que se duplicó.
	c, err := Guardar(auditoria.Silenciar(ctx), pool, tiendaID, "", e)
	if err != nil {
		return Producto{}, err
	}
	err = pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		if _, err := t.Exec(ctx, `UPDATE productos SET publicado = false WHERE tienda_id = $1 AND id = $2`, tiendaID, c.ID); err != nil {
			return err
		}
		return auditoria.Registrar(ctx, t, tiendaID, auditoria.Registro{Accion: "producto.duplicar", Entidad: "producto",
			EntidadID: c.ID, Texto: "duplicó " + p.Nombre + " (la copia quedó pausada)", Antes: map[string]string{"original": p.ID}})
	})
	c.Publicado = false
	return c, err
}
