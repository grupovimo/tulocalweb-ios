// Package lanzamientos: los lanzamientos («drops») de un producto, la oferta con fecha de fin, el
// encargo con seña y los avisos «Avisame cuando salga» (docs/decisiones.md §52). La tienda online
// los muestra con la vidriera; acá están la configuración del panel, el alta del aviso, el worker
// que avisa a la hora del lanzamiento y el que cancela las señas que no se pagaron.
package lanzamientos

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/auditoria"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

type tx interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
	EnTx(context.Context, func(pgx.Tx) error) error
}

var (
	ErrNoExiste = errors.New("lanzamientos: no existe")
	// ErrYaSalio: el producto ya está a la venta (o no tiene lanzamiento): no hay nada que avisar.
	ErrYaSalio = errors.New("lanzamientos: ya salió")
	// ErrLleno: el producto llegó al tope de avisos.
	ErrLleno = errors.New("lanzamientos: tope de avisos")
)

// ErrInvalido dice qué no se puede guardar.
type ErrInvalido struct{ Motivo string }

func (e ErrInvalido) Error() string { return e.Motivo }

// Zona: los lanzamientos y las ofertas se cargan y se muestran en hora de Argentina.
var Zona = func() *time.Location {
	if z, err := time.LoadLocation("America/Argentina/Buenos_Aires"); err == nil {
		return z
	}
	return time.FixedZone("ART", -3*3600)
}()

// TopeAvisos: más que esto por producto no se guarda (contra el abuso: el envío es de a tandas).
const TopeAvisos = 20000

// Config es lo que la tienda configura en el editor del producto. Reserva y Encargo son el % de
// seña (null: no se ofrece); OfertaPrecio, el precio de oferta (el de lista sigue en el producto).
type Config struct {
	Lanzamiento  *time.Time `json:"lanzamiento"`
	Reserva      *int       `json:"reserva"`
	Encargo      *int       `json:"encargo"`
	OfertaPrecio *int64     `json:"ofertaPrecio"`
	OfertaHasta  *time.Time `json:"ofertaHasta"`
}

// Estado es la configuración más lo que pasó: cuántos pidieron aviso (y a cuántos ya les salió),
// y cuántos pedidos reservaron o encargaron con seña (sin los cancelados).
type Estado struct {
	Config
	Avisos   int `json:"avisos"`
	Avisados int `json:"avisados"`
	Reservas int `json:"reservas"`
	Encargos int `json:"encargos"`
}

// Leer la configuración y los números de un producto.
func Leer(ctx context.Context, pool tx, tiendaID, id string) (Estado, error) {
	var e Estado
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error { return leerEnTx(ctx, t, tiendaID, id, &e) })
	return e, err
}

func leerEnTx(ctx context.Context, t pgx.Tx, tiendaID, id string, e *Estado) error {
	var reserva, encargo *int16
	err := t.QueryRow(ctx, `SELECT lanzamiento, reserva_sena, encargo_sena, oferta_precio, oferta_hasta,
			(SELECT count(*) FROM avisos_lanzamiento a WHERE a.tienda_id = p.tienda_id AND a.producto_id = p.id),
			(SELECT count(*) FROM avisos_lanzamiento a WHERE a.tienda_id = p.tienda_id AND a.producto_id = p.id AND a.avisado_en IS NOT NULL),
			(SELECT count(*) FILTER (WHERE x.sena_tipo = 'reserva') FROM pedidos x WHERE x.tienda_id = p.tienda_id
				AND x.sena IS NOT NULL AND x.estado <> 'Cancelado' AND x.items @> jsonb_build_array(jsonb_build_object('productoId', p.id::text))),
			(SELECT count(*) FILTER (WHERE x.sena_tipo = 'encargo') FROM pedidos x WHERE x.tienda_id = p.tienda_id
				AND x.sena IS NOT NULL AND x.estado <> 'Cancelado' AND x.items @> jsonb_build_array(jsonb_build_object('productoId', p.id::text)))
		FROM productos p WHERE p.tienda_id = $1 AND p.id::text = $2`, tiendaID, id).
		Scan(&e.Lanzamiento, &reserva, &encargo, &e.OfertaPrecio, &e.OfertaHasta, &e.Avisos, &e.Avisados, &e.Reservas, &e.Encargos)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNoExiste
	}
	if reserva != nil {
		n := int(*reserva)
		e.Reserva = &n
	}
	if encargo != nil {
		n := int(*encargo)
		e.Encargo = &n
	}
	return err
}

func igual(a, b *time.Time) bool {
	return (a == nil && b == nil) || (a != nil && b != nil && a.Equal(*b))
}

// Guardar la configuración. Una fecha nueva (lanzamiento o fin de oferta) tiene que ser futura; la
// que ya estaba se puede dejar aunque haya pasado. Reprogramar el lanzamiento a una fecha futura
// vuelve a dejar pendiente el aviso.
func Guardar(ctx context.Context, pool tx, tiendaID, id string, c Config, ahora time.Time) (Estado, error) {
	var e Estado
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var antes Estado
		if err := leerEnTx(ctx, t, tiendaID, id, &antes); err != nil {
			return err
		}
		var nombre string
		var precio int64
		if err := t.QueryRow(ctx, `SELECT nombre, precio FROM productos WHERE tienda_id = $1 AND id::text = $2 FOR UPDATE`,
			tiendaID, id).Scan(&nombre, &precio); err != nil {
			return err
		}
		if err := validar(c, antes.Config, precio, ahora); err != nil {
			return err
		}
		if _, err := t.Exec(ctx, `UPDATE productos SET lanzamiento = $3, reserva_sena = $4, encargo_sena = $5,
				oferta_precio = $6, oferta_hasta = $7,
				lanzamiento_avisado = CASE WHEN $3::timestamptz IS DISTINCT FROM lanzamiento AND $3 > now() THEN NULL ELSE lanzamiento_avisado END
			WHERE tienda_id = $1 AND id::text = $2`, tiendaID, id, c.Lanzamiento, c.Reserva, c.Encargo, c.OfertaPrecio, c.OfertaHasta); err != nil {
			return err
		}
		if err := auditoria.Registrar(ctx, t, tiendaID, auditoria.Registro{Accion: "producto.lanzamiento", Entidad: "producto",
			EntidadID: id, Texto: textoAuditoria(nombre, c), Antes: antes.Config, Despues: c}); err != nil {
			return err
		}
		return leerEnTx(ctx, t, tiendaID, id, &e)
	})
	return e, err
}

func validar(c, antes Config, precio int64, ahora time.Time) error {
	pct := func(p *int) bool { return p == nil || (*p >= 1 && *p <= 100) }
	switch {
	case c.Lanzamiento != nil && !igual(c.Lanzamiento, antes.Lanzamiento) && !c.Lanzamiento.After(ahora):
		return ErrInvalido{"La fecha del lanzamiento tiene que ser más adelante que ahora."}
	case c.Lanzamiento != nil && c.Lanzamiento.After(ahora.AddDate(1, 0, 0)):
		return ErrInvalido{"El lanzamiento puede ser hasta dentro de un año."}
	case c.Reserva != nil && c.Lanzamiento == nil:
		return ErrInvalido{"Para reservar con seña, poné la fecha del lanzamiento."}
	case !pct(c.Reserva) || !pct(c.Encargo):
		return ErrInvalido{"La seña va de 1 % a 100 %."}
	case c.OfertaPrecio != nil && (*c.OfertaPrecio <= 0 || *c.OfertaPrecio >= precio):
		return ErrInvalido{"El precio de oferta tiene que ser menor que el precio del producto."}
	case c.OfertaHasta != nil && c.OfertaPrecio == nil:
		return ErrInvalido{"Para ponerle fin a la oferta, cargá el precio de oferta."}
	case c.OfertaHasta != nil && !igual(c.OfertaHasta, antes.OfertaHasta) && !c.OfertaHasta.After(ahora):
		return ErrInvalido{"La oferta tiene que terminar más adelante que ahora."}
	}
	return nil
}

func textoAuditoria(nombre string, c Config) string {
	partes := []string{}
	if c.Lanzamiento != nil {
		partes = append(partes, "lanzamiento el "+c.Lanzamiento.In(Zona).Format("02/01 15:04"))
	}
	if c.Reserva != nil {
		partes = append(partes, fmt.Sprintf("reserva con %d %% de seña", *c.Reserva))
	}
	if c.Encargo != nil {
		partes = append(partes, fmt.Sprintf("encargo con %d %% de seña", *c.Encargo))
	}
	if c.OfertaPrecio != nil {
		o := "oferta a " + auditoria.Pesos(*c.OfertaPrecio)
		if c.OfertaHasta != nil {
			o += " hasta el " + c.OfertaHasta.In(Zona).Format("02/01 15:04")
		}
		partes = append(partes, o)
	}
	if len(partes) == 0 {
		return "sacó el lanzamiento, la seña y la oferta de " + nombre
	}
	return "configuró " + nombre + ": " + strings.Join(partes, ", ")
}

var emailValido = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// Avisar anota un email para avisarle cuando salga el producto. Repetido no suma (ni dice nada:
// así no se puede averiguar quién se anotó).
func Avisar(ctx context.Context, pool tx, slug, productoID, email string, ahora time.Time) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if !emailValido.MatchString(email) || len(email) > 200 || strings.ContainsAny(email, "\r\n<>,;") {
		return ErrInvalido{"Revisá el email."}
	}
	var tienda *string
	if err := pool.EnTx(ctx, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT app.tienda_de_slug($1)::text`, slug).Scan(&tienda)
	}); err != nil {
		return err
	}
	if tienda == nil {
		return ErrNoExiste
	}
	return pool.EnTienda(ctx, *tienda, func(t pgx.Tx) error {
		var lanzamiento *time.Time
		var cuantos int
		err := t.QueryRow(ctx, `SELECT p.lanzamiento, (SELECT count(*) FROM avisos_lanzamiento a WHERE a.tienda_id = p.tienda_id AND a.producto_id = p.id)
			FROM productos p WHERE p.tienda_id = $1 AND p.id::text = $2 AND p.publicado FOR UPDATE`, *tienda, productoID).Scan(&lanzamiento, &cuantos)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoExiste
		}
		if err != nil {
			return err
		}
		if lanzamiento == nil || !lanzamiento.After(ahora) {
			return ErrYaSalio
		}
		if cuantos >= TopeAvisos {
			return ErrLleno
		}
		_, err = t.Exec(ctx, `INSERT INTO avisos_lanzamiento (tienda_id, producto_id, email) VALUES ($1, $2::uuid, $3)
			ON CONFLICT DO NOTHING`, *tienda, productoID, email)
		return err
	})
}

// PorLanzar encola los avisos de los lanzamientos que llegaron (lo corre el worker cada minuto). Un
// producto se marca una sola vez: el UPDATE condicionado decide aunque corran dos workers.
func PorLanzar(ctx context.Context, pool tx) (int, error) {
	type par struct{ tienda, producto string }
	var pares []par
	if err := pool.EnTx(ctx, func(t pgx.Tx) error {
		filas, err := t.Query(ctx, `SELECT tienda_id::text, producto_id::text FROM app.lanzamientos_por_avisar(200)`)
		if err != nil {
			return err
		}
		defer filas.Close()
		for filas.Next() {
			var p par
			if err := filas.Scan(&p.tienda, &p.producto); err != nil {
				return err
			}
			pares = append(pares, p)
		}
		return filas.Err()
	}); err != nil {
		return 0, err
	}
	n := 0
	for _, p := range pares {
		err := pool.EnTienda(ctx, p.tienda, func(t pgx.Tx) error {
			tag, err := t.Exec(ctx, `UPDATE productos SET lanzamiento_avisado = now()
				WHERE tienda_id = $1 AND id::text = $2 AND lanzamiento <= now() AND lanzamiento_avisado IS NULL`, p.tienda, p.producto)
			if err != nil || tag.RowsAffected() == 0 {
				return err
			}
			n++
			return outbox.Publicar(ctx, t, p.tienda, "lanzamiento.avisos", map[string]string{"producto": p.producto})
		})
		if err != nil {
			return n, err
		}
	}
	return n, nil
}

// VencerSenas cancela los pedidos cuya seña no se pagó a tiempo y devuelve su stock.
func VencerSenas(ctx context.Context, pool tx) (int, error) {
	type par struct {
		tienda string
		numero int
	}
	var pares []par
	if err := pool.EnTx(ctx, func(t pgx.Tx) error {
		filas, err := t.Query(ctx, `SELECT tienda_id::text, numero FROM app.senas_vencidas(200)`)
		if err != nil {
			return err
		}
		defer filas.Close()
		for filas.Next() {
			var p par
			if err := filas.Scan(&p.tienda, &p.numero); err != nil {
				return err
			}
			pares = append(pares, p)
		}
		return filas.Err()
	}); err != nil {
		return 0, err
	}
	n := 0
	for _, p := range pares {
		hecho, err := pedidos.CancelarSinSena(ctx, pool, p.tienda, p.numero, true)
		if err != nil {
			return n, err
		}
		if hecho {
			n++
		}
	}
	return n, nil
}
