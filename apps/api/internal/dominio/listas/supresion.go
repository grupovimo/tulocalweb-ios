package listas

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/auditoria"
)

// La supresión y la rectificación de los datos de un comprador (arts. 16 y 4 de la Ley 25.326), con
// app.suprimir_comprador y app.rectificar_comprador (migración 00079). La clave es el id de la fila
// de Clientes: el email en minúsculas, el nombre si compró sin email, o el id de su ficha.

var (
	// ErrEnCurso: tiene un pedido sin entregar, un turno por venir, una suscripción viva o una
	// devolución abierta. Primero se termina o se cancela eso.
	ErrEnCurso = errors.New("tiene un pedido, un turno, una suscripción o una devolución en curso: terminalo o cancelalo primero")
	// ErrEmailEnUso: el email nuevo ya es de otro cliente de la tienda.
	ErrEmailEnUso = errors.New("ese email ya es de otro cliente")
	// ErrDatosCliente: el nombre, el email o el teléfono no sirven.
	ErrDatosCliente = errors.New("revisá el nombre (de 2 a 80 letras), el email y el WhatsApp")
)

// AniosInactivo: sin compras en 3 años, el cliente cuenta como inactivo (Mis datos).
const AniosInactivo = 3

// clavesInactivas: los clientes de la lista (pedidos no cancelados, historial importado y fichas)
// cuya última actividad es de hace más de $2 años.
const clavesInactivas = `
	SELECT clave FROM (
		SELECT app.clave_comprador(comprador->>'email', comprador->>'nombre') AS clave, creado_en AS f
		FROM pedidos WHERE tienda_id = $1 AND estado <> 'Cancelado' AND NOT ` + anonimo + `
		UNION ALL
		SELECT app.clave_comprador(cliente_email, cliente_nombre), fecha FROM historial_pedidos WHERE tienda_id = $1
		UNION ALL
		SELECT coalesce(lower(email), id::text), creado_en FROM clientes WHERE tienda_id = $1
	) a WHERE clave IS NOT NULL
	GROUP BY clave HAVING max(f) < now() - make_interval(years => $2)`

// ClientesInactivos devuelve las claves de los clientes sin compras en AniosInactivo años.
func ClientesInactivos(ctx context.Context, t pgx.Tx, tiendaID string) ([]string, error) {
	filas, err := t.Query(ctx, clavesInactivas, tiendaID, AniosInactivo)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(filas, pgx.RowTo[string])
}

// ContarInactivos cuenta los clientes sin compras en AniosInactivo años.
func ContarInactivos(ctx context.Context, t pgx.Tx, tiendaID string) (int, error) {
	var n int
	err := t.QueryRow(ctx, `SELECT count(*) FROM (`+clavesInactivas+`) i`, tiendaID, AniosInactivo).Scan(&n)
	return n, err
}

func errFuncion(err error) error {
	var pg *pgconn.PgError
	if errors.As(err, &pg) && pg.Code == "P0001" {
		switch pg.Message {
		case "comprador_en_curso":
			return ErrEnCurso
		case "email_en_uso":
			return ErrEmailEnUso
		}
	}
	return err
}

// suprimir corre la supresión dentro de una transacción de la tienda, en un savepoint: si el
// comprador tiene algo en curso, la transacción sigue viva para los demás.
func suprimir(ctx context.Context, t pgx.Tx, clave string) (int, error) {
	sp, err := t.Begin(ctx)
	if err != nil {
		return 0, err
	}
	var n int
	if err := sp.QueryRow(ctx, `SELECT app.suprimir_comprador($1)`, clave).Scan(&n); err != nil {
		_ = sp.Rollback(ctx)
		return 0, errFuncion(err)
	}
	return n, sp.Commit(ctx)
}

// SuprimirClientes borra o anonimiza a los compradores de esas claves (el pedido de supresión). Los
// que tienen algo en curso quedan como estaban; si ninguno se pudo, devuelve ErrEnCurso. Devuelve
// cuántos se suprimieron. En la actividad queda que se hizo, sin decir de quién.
func SuprimirClientes(ctx context.Context, pool tx, tiendaID string, claves []string) (int, error) {
	hechos, enCurso := 0, 0
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		for _, c := range claves {
			c = strings.TrimSpace(c)
			if c == "" || len(c) > 200 {
				continue
			}
			n, err := suprimir(ctx, t, c)
			if errors.Is(err, ErrEnCurso) {
				enCurso++
				continue
			}
			if err != nil {
				return err
			}
			if n > 0 {
				hechos++
			}
		}
		if hechos == 0 {
			return nil
		}
		texto := "borró los datos de un cliente"
		if hechos > 1 {
			texto = "borró los datos de varios clientes"
		}
		return auditoria.Registrar(ctx, t, tiendaID, auditoria.Registro{Accion: "cliente.suprimir", Entidad: "cliente", Texto: texto})
	})
	if err == nil && hechos == 0 && enCurso > 0 {
		return 0, ErrEnCurso
	}
	return hechos, err
}

// EditarCliente rectifica el nombre, el email y el teléfono de un comprador (valor: {nombre, email,
// telefono}) en su ficha, sus pedidos y lo importado. Devuelve 1 si lo encontró.
func EditarCliente(ctx context.Context, pool tx, tiendaID, clave string, v json.RawMessage) (int, error) {
	var c struct{ Nombre, Email, Telefono string }
	if json.Unmarshal(v, &c) != nil {
		return 0, ErrDatosCliente
	}
	nombre := strings.TrimSpace(c.Nombre)
	email := strings.ToLower(strings.TrimSpace(c.Email))
	tel := strings.TrimSpace(c.Telefono)
	if n := len([]rune(nombre)); n < 2 || n > 80 || len(email) > 120 || len([]rune(tel)) > 30 ||
		(email != "" && (len(email) < 3 || !strings.Contains(email, "@"))) || strings.TrimSpace(clave) == "" {
		return 0, ErrDatosCliente
	}
	n := 0
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		if err := t.QueryRow(ctx, `SELECT app.rectificar_comprador($1, $2, $3, $4)`, clave, nombre, email, tel).Scan(&n); err != nil {
			return errFuncion(err)
		}
		if n == 0 {
			return nil
		}
		n = 1
		return auditoria.Registrar(ctx, t, tiendaID, auditoria.Registro{Accion: "cliente.editar", Entidad: "cliente",
			Texto: "corrigió los datos de " + nombre})
	})
	return n, err
}
