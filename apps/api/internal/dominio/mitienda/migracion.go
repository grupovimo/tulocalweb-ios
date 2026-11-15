package mitienda

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
)

// CambioV2: una tienda con algún diseño v1 (borrador, publicado o los dos).
type CambioV2 struct {
	Slug      string
	Borrador  bool
	Publicado bool
}

// esV1: un diseño guardado que todavía no es v2 (null o vacío no: se arma al leer).
func esV1(crudo []byte) bool {
	if len(crudo) == 0 || string(crudo) == "null" {
		return false
	}
	var v struct {
		Version int `json:"version"`
	}
	return json.Unmarshal(crudo, &v) != nil || v.Version != 2
}

// MigrarBase pasa a v2 los diseños v1 de todas las tiendas, en una sola transacción (como el
// migrador: db/migraciones/00080_diseno_v2.sql). Es idempotente: lo que ya es v2 no se toca. Con
// seco sólo dice qué cambiaría.
func MigrarBase(ctx context.Context, conn interface {
	Begin(context.Context) (pgx.Tx, error)
}, seco bool) ([]CambioV2, error) {
	var cambios []CambioV2
	err := pgx.BeginFunc(ctx, conn, func(t pgx.Tx) error {
		if _, err := t.Exec(ctx, `SELECT set_config('app.diseno_v2', 'si', true)`); err != nil {
			return err
		}
		filas, err := t.Query(ctx, `SELECT id::text, slug, plantilla, diseno_borrador, diseno_publicado FROM tiendas ORDER BY slug`)
		if err != nil {
			return err
		}
		type fila struct {
			id, slug, plantilla string
			borrador, publicado []byte
		}
		var todas []fila
		for filas.Next() {
			var f fila
			if err := filas.Scan(&f.id, &f.slug, &f.plantilla, &f.borrador, &f.publicado); err != nil {
				filas.Close()
				return err
			}
			todas = append(todas, f)
		}
		if err := filas.Err(); err != nil {
			return err
		}
		for _, f := range todas {
			c := CambioV2{Slug: f.slug, Borrador: esV1(f.borrador), Publicado: esV1(f.publicado)}
			if !c.Borrador && !c.Publicado {
				continue
			}
			cambios = append(cambios, c)
			if seco {
				continue
			}
			// Cada uno con su plantilla: la del borrador, si la cambia, la toma Migrar del diseño.
			b, p := f.borrador, f.publicado
			if c.Borrador {
				b, _ = json.Marshal(leerDiseno(f.borrador, f.plantilla))
			}
			if c.Publicado {
				p, _ = json.Marshal(leerDiseno(f.publicado, f.plantilla))
			}
			if _, err := t.Exec(ctx, `SELECT set_config('app.tienda_id', $1, true)`, f.id); err != nil {
				return err
			}
			if _, err := t.Exec(ctx, `UPDATE tiendas SET diseno_borrador = $2::jsonb, diseno_publicado = $3::jsonb WHERE id = $1`,
				f.id, nulo(b), nulo(p)); err != nil {
				return err
			}
		}
		return nil
	})
	return cambios, err
}

// nulo: un jsonb vacío va como NULL.
func nulo(b []byte) any {
	if len(b) == 0 {
		return nil
	}
	return string(b)
}
