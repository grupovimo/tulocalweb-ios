package tiendas

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db"
)

// Origen es lo que trajo a la persona a crearse la cuenta (/empezar o un anuncio, /l/).
type Origen struct {
	Fuente  string `json:"fuente,omitempty"`
	Medio   string `json:"medio,omitempty"`
	Campana string `json:"campana,omitempty"`
	// utm_content: el anuncio puntual (en Meta, su id).
	Contenido string `json:"contenido,omitempty"`
	Ref       string `json:"ref,omitempty"`
	Cupon     string `json:"cupon,omitempty"`
	// Cuándo llegó a /empezar (ISO), lo pone el BFF.
	Llego string `json:"llego,omitempty"`
}

var valorOrigen = regexp.MustCompile(`^[A-Za-z0-9_.-]{1,40}$`)

// ErrOrigen: algún valor no es válido.
var ErrOrigen = errors.New("tiendas: origen inválido")

// valido: al menos un dato de atribución (un origen vacío no se fija: taparía un ref posterior).
func (o Origen) valido() bool {
	alguno := false
	for _, v := range []string{o.Fuente, o.Medio, o.Campana, o.Contenido, o.Ref, o.Cupon} {
		if v != "" && !valorOrigen.MatchString(v) {
			return false
		}
		alguno = alguno || v != ""
	}
	return alguno && len(o.Llego) <= 40
}

// FijarOrigen guarda el origen de la tienda una sola vez (después no se pisa: una comisión no se
// cambia de dueño) y sólo en el alta: una tienda de más de un día ya no se puede atribuir a un
// partner pasando por /empezar. Devuelve false si ya tenía uno o si es vieja.
func FijarOrigen(ctx context.Context, pool *db.Pool, tiendaID string, o Origen) (bool, error) {
	if !o.valido() {
		return false, ErrOrigen
	}
	b, _ := json.Marshal(o)
	var fijado bool
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		tag, err := t.Exec(ctx, `UPDATE tiendas SET origen = $2
			WHERE id = $1 AND origen IS NULL AND creada_en > now() - interval '1 day'`, tiendaID, b)
		fijado = tag.RowsAffected() == 1
		return err
	})
	return fijado, err
}
