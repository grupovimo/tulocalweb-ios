package configuracion

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/auditoria"
)

// DiasParaPurgar: una tienda cerrada se borra a los 90 días (la política de privacidad y Mis datos
// lo dicen así). Antes, se puede reabrir.
const DiasParaPurgar = 90

// DiasParaFrenar: el borrado de los clientes inactivos espera 7 días, para poder frenarlo.
const DiasParaFrenar = 7

// enTx es lo que necesita lo que pasa por todas las tiendas (el worker y mostradorctl).
type enTx interface {
	tx
	EnTx(context.Context, func(pgx.Tx) error) error
}

// TiendaCerrada es una tienda que ya se puede borrar.
type TiendaCerrada struct {
	ID        string
	Nombre    string
	CerradaEn time.Time
}

// ParaPurgar lista las tiendas cerradas hace más de DiasParaPurgar días (hasta limite).
func ParaPurgar(ctx context.Context, pool enTx, limite int) ([]TiendaCerrada, error) {
	var l []TiendaCerrada
	err := pool.EnTx(ctx, func(t pgx.Tx) error {
		filas, err := t.Query(ctx, `SELECT id::text, nombre, cerrada_en FROM app.tiendas_para_purgar($1, $2)`, DiasParaPurgar, limite)
		if err != nil {
			return err
		}
		l, err = pgx.CollectRows(filas, func(f pgx.CollectableRow) (TiendaCerrada, error) {
			var c TiendaCerrada
			return c, f.Scan(&c.ID, &c.Nombre, &c.CerradaEn)
		})
		return err
	})
	return l, err
}

var idTienda = regexp.MustCompile(`^[0-9a-f-]{36}$`)

// PurgarCerradas borra las tiendas cerradas hace más de DiasParaPurgar días, de a una por
// transacción (app.purgar_tienda: lo contable queda en archivo_contable, sin datos de quien compró),
// y sus fotos de dirFotos/fotos/<tienda> (la carpeta de archivos de la web; vacía, no toca el disco).
// También borra lo archivado hace más de 10 años. Devuelve cuántas tiendas borró.
func PurgarCerradas(ctx context.Context, pool enTx, dirFotos string) (int, error) {
	l, err := ParaPurgar(ctx, pool, 50)
	if err != nil {
		return 0, err
	}
	n := 0
	var errs error
	for _, c := range l {
		var borrada bool
		if err := pool.EnTx(ctx, func(t pgx.Tx) error {
			return t.QueryRow(ctx, `SELECT app.purgar_tienda($1, $2)`, c.ID, DiasParaPurgar).Scan(&borrada)
		}); err != nil {
			errs = errors.Join(errs, err)
			continue
		}
		if !borrada {
			continue
		}
		n++
		if dirFotos != "" && idTienda.MatchString(c.ID) {
			errs = errors.Join(errs, os.RemoveAll(filepath.Join(dirFotos, "fotos", c.ID)))
		}
	}
	errs = errors.Join(errs, pool.EnTx(ctx, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `SELECT app.purgar_archivo_contable()`)
		return err
	}))
	return n, errs
}

// ProgramarBorradoInactivos deja programado, a DiasParaFrenar días, el borrado de los clientes sin
// compras en 3 años. Devuelve cuántos son (0: no programa nada).
func ProgramarBorradoInactivos(ctx context.Context, pool tx, tiendaID string) (int, error) {
	n := 0
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var err error
		if n, err = listas.ContarInactivos(ctx, t, tiendaID); err != nil || n == 0 {
			return err
		}
		_, err = t.Exec(ctx, `UPDATE tiendas SET borrado_inactivos = now() + make_interval(days => $2) WHERE id = $1`, tiendaID, DiasParaFrenar)
		return err
	})
	return n, err
}

// FrenarBorradoInactivos cancela el borrado programado.
func FrenarBorradoInactivos(ctx context.Context, pool tx, tiendaID string) error {
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `UPDATE tiendas SET borrado_inactivos = NULL WHERE id = $1`, tiendaID)
		return err
	})
}

// BorrarInactivosProgramados hace los borrados de clientes inactivos que llegaron a su fecha (el
// worker, una vez por día). Los clientes con algo en curso quedan. Devuelve cuántos borró.
func BorrarInactivosProgramados(ctx context.Context, pool enTx) (int, error) {
	var tiendas []string
	if err := pool.EnTx(ctx, func(t pgx.Tx) error {
		filas, err := t.Query(ctx, `SELECT t::text FROM app.inactivos_programados(50) t`)
		if err != nil {
			return err
		}
		tiendas, err = pgx.CollectRows(filas, pgx.RowTo[string])
		return err
	}); err != nil {
		return 0, err
	}
	total := 0
	var errs error
	for _, id := range tiendas {
		var claves []string
		err := pool.EnTienda(ctx, id, func(t pgx.Tx) error {
			var err error
			claves, err = listas.ClientesInactivos(ctx, t, id)
			return err
		})
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}
		ctxW := auditoria.ConActor(ctx, auditoria.Actor{Tipo: "sistema", Nombre: "Tu Local Web"})
		n, err := listas.SuprimirClientes(ctxW, pool, id, claves)
		if err != nil && !errors.Is(err, listas.ErrEnCurso) {
			errs = errors.Join(errs, err)
			continue
		}
		total += n
		errs = errors.Join(errs, FrenarBorradoInactivos(ctx, pool, id))
	}
	return total, errs
}
