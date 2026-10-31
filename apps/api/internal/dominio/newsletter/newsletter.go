// Package newsletter: el newsletter de cada tienda (docs/decisiones.md §50). Quien compra se suma
// desde la tienda online (el bloque del pie o la casilla del checkout) y queda suscripto recién al
// tocar el link del mail de confirmación. La tienda manda novedades desde el panel (asunto, texto y
// hasta 4 productos); el worker las reparte de a tandas con un tope diario por tienda, cada mail con
// la baja de un clic (List-Unsubscribe y List-Unsubscribe-Post, RFC 8058).
package newsletter

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

const (
	// TopeDiario: mails de novedades por tienda por día de Argentina (cuida la reputación del
	// dominio). Lo que no entra sigue al día siguiente.
	TopeDiario = 500
	// Tanda: cuántos manda el worker por evento antes de encolar el siguiente.
	Tanda = 25
	// MaxProductos destacados por novedad.
	MaxProductos = 4
	// MaxIntentos por destinatario antes de darlo por fallido.
	MaxIntentos = 3
)

var (
	ErrInvalido         = errors.New("newsletter: datos inválidos")
	ErrNoExiste         = errors.New("newsletter: no existe")
	ErrSinDestinatarios = errors.New("newsletter: no hay suscriptos confirmados")
	ErrDemasiados       = errors.New("newsletter: demasiados intentos")
)

// ErrCampo: un dato inválido con el motivo para mostrar.
type ErrCampo struct{ Campo, Mensaje string }

func (e ErrCampo) Error() string { return e.Campo + ": " + e.Mensaje }

var argentina = time.FixedZone("ART", -3*3600)

type tx interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
}

func nuevoToken() string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// Email normaliza y valida un email (minúsculas, sin espacios, hasta 120).
func Email(v string) (string, bool) {
	e := strings.ToLower(strings.TrimSpace(v))
	a, err := mail.ParseAddress(e)
	if err != nil || a.Address != e || len(e) > 120 || !strings.Contains(e[strings.LastIndex(e, "@"):], ".") {
		return "", false
	}
	return e, true
}

// Suscribir suma el email como pendiente y encola el mail de confirmación. Si ya estaba pendiente,
// vuelve a mandar la confirmación (con el mismo link); si ya confirmó, no hace nada.
func Suscribir(ctx context.Context, pool tx, tienda, email, origen string) error {
	return pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		return SuscribirEnTx(ctx, t, tienda, email, origen)
	})
}

// SuscribirEnTx es Suscribir dentro de una transacción con la tienda fijada (el checkout).
func SuscribirEnTx(ctx context.Context, t pgx.Tx, tienda, email, origen string) error {
	e, ok := Email(email)
	if !ok || (origen != "tienda" && origen != "checkout") {
		return ErrInvalido
	}
	var id string
	err := t.QueryRow(ctx, `INSERT INTO suscriptores (tienda_id, email, origen, token) VALUES ($1, $2, $3, $4)
		ON CONFLICT (tienda_id, lower(email)) DO UPDATE SET origen = suscriptores.origen
		WHERE suscriptores.confirmado_en IS NULL
		RETURNING id::text`, tienda, e, origen, nuevoToken()).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil // ya confirmado: nada que mandar
	}
	if err != nil {
		return err
	}
	return outbox.Publicar(ctx, t, tienda, "newsletter.confirmar", map[string]string{"id": id})
}

func tokenValido(token string) bool {
	if len(token) < 20 || len(token) > 64 {
		return false
	}
	for _, r := range token {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_') {
			return false
		}
	}
	return true
}

// Confirmar deja suscripto al dueño del token (tocar dos veces el link no cambia la fecha).
func Confirmar(ctx context.Context, pool tx, tienda, token string) error {
	if !tokenValido(token) {
		return ErrNoExiste
	}
	return pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		tag, err := t.Exec(ctx, `UPDATE suscriptores SET confirmado_en = coalesce(confirmado_en, now()) WHERE token = $1`, token)
		if err == nil && tag.RowsAffected() == 0 {
			return ErrNoExiste
		}
		return err
	})
}

// Baja borra la suscripción del token (el email no queda guardado). Las tandas pendientes la saltean.
func Baja(ctx context.Context, pool tx, tienda, token string) error {
	if !tokenValido(token) {
		return ErrNoExiste
	}
	return pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		tag, err := t.Exec(ctx, `DELETE FROM suscriptores WHERE token = $1`, token)
		if err == nil && tag.RowsAffected() == 0 {
			return ErrNoExiste
		}
		return err
	})
}

// Suscriptor es una fila de la lista del panel (components/schemas/Newsletter).
type Suscriptor struct {
	ID           string     `json:"id"`
	Email        string     `json:"email"`
	Origen       string     `json:"origen"`
	Confirmado   bool       `json:"confirmado"`
	Creado       time.Time  `json:"creado"`
	ConfirmadoEn *time.Time `json:"confirmadoEn"`
}

// Envio es un «Mandar novedad» del historial.
type Envio struct {
	ID            string    `json:"id"`
	Asunto        string    `json:"asunto"`
	Creado        time.Time `json:"creado"`
	Destinatarios int       `json:"destinatarios"`
	Enviados      int       `json:"enviados"`
	// Estado: enviando, manana (se llegó al tope del día y sigue mañana) o enviado.
	Estado string `json:"estado"`
}

// Panel es lo que muestra Crecer › Newsletter.
type Panel struct {
	Suscriptores []Suscriptor `json:"suscriptores"`
	Confirmados  int          `json:"confirmados"`
	Pendientes   int          `json:"pendientes"`
	Envios       []Envio      `json:"envios"`
	// Hoy: mails de novedades que ya salieron hoy; Tope: el máximo del día.
	Hoy  int `json:"hoy"`
	Tope int `json:"tope"`
	// Productos publicados para destacar en una novedad (por nombre).
	Productos []Elegible `json:"productos"`
}

// Elegible: un producto publicado que se puede destacar.
type Elegible struct {
	ID     string  `json:"id"`
	Nombre string  `json:"nombre"`
	Precio int64   `json:"precio"`
	Foto   *string `json:"foto"`
}

func inicioDelDia(ahora time.Time) time.Time {
	a := ahora.In(argentina)
	return time.Date(a.Year(), a.Month(), a.Day(), 0, 0, 0, 0, argentina)
}

// MaxLista: cuántos suscriptos devuelve el panel (la exportación usa la misma lista).
const MaxLista = 5000

// Leer arma el panel: los suscriptos (los más nuevos primero), los números y los últimos envíos.
func Leer(ctx context.Context, pool tx, tienda string, ahora time.Time) (Panel, error) {
	p := Panel{Suscriptores: []Suscriptor{}, Envios: []Envio{}, Tope: TopeDiario, Productos: []Elegible{}}
	err := pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		filas, err := t.Query(ctx, `SELECT id::text, email, origen, creado, confirmado_en FROM suscriptores
			WHERE tienda_id = $1 ORDER BY creado DESC, id LIMIT $2`, tienda, MaxLista)
		if err != nil {
			return err
		}
		for filas.Next() {
			var s Suscriptor
			if err := filas.Scan(&s.ID, &s.Email, &s.Origen, &s.Creado, &s.ConfirmadoEn); err != nil {
				filas.Close()
				return err
			}
			s.Confirmado = s.ConfirmadoEn != nil
			p.Suscriptores = append(p.Suscriptores, s)
		}
		filas.Close()
		if err := filas.Err(); err != nil {
			return err
		}
		if err := t.QueryRow(ctx, `SELECT count(*) FILTER (WHERE confirmado_en IS NOT NULL), count(*) FILTER (WHERE confirmado_en IS NULL)
			FROM suscriptores WHERE tienda_id = $1`, tienda).Scan(&p.Confirmados, &p.Pendientes); err != nil {
			return err
		}
		if err := t.QueryRow(ctx, `SELECT count(*) FROM newsletter_entregas WHERE tienda_id = $1 AND estado = 'enviado'
			AND enviado_en >= $2`, tienda, inicioDelDia(ahora)).Scan(&p.Hoy); err != nil {
			return err
		}
		pf, err := t.Query(ctx, `SELECT id::text, nombre, precio, foto FROM productos WHERE tienda_id = $1 AND publicado
			ORDER BY nombre LIMIT 300`, tienda)
		if err != nil {
			return err
		}
		for pf.Next() {
			var e Elegible
			if err := pf.Scan(&e.ID, &e.Nombre, &e.Precio, &e.Foto); err != nil {
				pf.Close()
				return err
			}
			p.Productos = append(p.Productos, e)
		}
		pf.Close()
		if err := pf.Err(); err != nil {
			return err
		}
		filas, err = t.Query(ctx, `SELECT e.id::text, e.asunto, e.creado, e.destinatarios, e.terminado IS NOT NULL,
				(SELECT count(*) FROM newsletter_entregas x WHERE x.envio_id = e.id AND x.estado = 'enviado')
			FROM newsletter_envios e WHERE e.tienda_id = $1 ORDER BY e.creado DESC LIMIT 30`, tienda)
		if err != nil {
			return err
		}
		defer filas.Close()
		for filas.Next() {
			var e Envio
			var terminado bool
			if err := filas.Scan(&e.ID, &e.Asunto, &e.Creado, &e.Destinatarios, &terminado, &e.Enviados); err != nil {
				return err
			}
			switch {
			case terminado:
				e.Estado = "enviado"
			case p.Hoy >= TopeDiario:
				e.Estado = "manana"
			default:
				e.Estado = "enviando"
			}
			p.Envios = append(p.Envios, e)
		}
		return filas.Err()
	})
	return p, err
}

// Borrar saca a un suscripto de la lista (confirmado o no).
func Borrar(ctx context.Context, pool tx, tienda, id string) error {
	return pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		tag, err := t.Exec(ctx, `DELETE FROM suscriptores WHERE id::text = $1`, id)
		if err == nil && tag.RowsAffected() == 0 {
			return ErrNoExiste
		}
		return err
	})
}

// Novedad: lo que se manda desde el panel.
type Novedad struct {
	Asunto    string   `json:"asunto"`
	Texto     string   `json:"texto"`
	Productos []string `json:"productos"`
}

// Validar normaliza la novedad y devuelve el primer problema.
func (n *Novedad) Validar() error {
	n.Asunto = strings.Join(strings.Fields(n.Asunto), " ")
	n.Texto = strings.TrimSpace(strings.ReplaceAll(n.Texto, "\r\n", "\n"))
	switch l := utf8.RuneCountInString(n.Asunto); {
	case l < 3:
		return ErrCampo{"asunto", "Poné un asunto."}
	case l > 120:
		return ErrCampo{"asunto", "El asunto puede tener hasta 120 caracteres."}
	}
	switch l := utf8.RuneCountInString(n.Texto); {
	case l < 10:
		return ErrCampo{"texto", "Escribí la novedad (al menos 10 caracteres)."}
	case l > 5000:
		return ErrCampo{"texto", "El texto puede tener hasta 5000 caracteres."}
	}
	if len(n.Productos) > MaxProductos {
		return ErrCampo{"productos", "Elegí hasta 4 productos."}
	}
	vistos := map[string]bool{}
	ps := n.Productos[:0]
	for _, p := range n.Productos {
		if p != "" && !vistos[p] {
			vistos[p] = true
			ps = append(ps, p)
		}
	}
	n.Productos = ps
	return nil
}

// Mandar fija los destinatarios (los confirmados de ahora) y encola la primera tanda.
func Mandar(ctx context.Context, pool tx, tienda string, n Novedad, ahora time.Time) (Envio, error) {
	if err := n.Validar(); err != nil {
		return Envio{}, err
	}
	e := Envio{Asunto: n.Asunto, Creado: ahora, Estado: "enviando"}
	err := pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		if len(n.Productos) > 0 {
			var validos int
			if err := t.QueryRow(ctx, `SELECT count(*) FROM productos WHERE id::text = ANY($1) AND publicado`, n.Productos).Scan(&validos); err != nil {
				return err
			}
			if validos != len(n.Productos) {
				return ErrCampo{"productos", "Algún producto ya no está publicado. Elegilos de nuevo."}
			}
		}
		productos, _ := json.Marshal(n.Productos)
		if err := t.QueryRow(ctx, `INSERT INTO newsletter_envios (tienda_id, asunto, texto, productos, destinatarios, creado)
			VALUES ($1, $2, $3, $4, (SELECT count(*) FROM suscriptores WHERE confirmado_en IS NOT NULL), $5)
			RETURNING id::text, destinatarios`, tienda, n.Asunto, n.Texto, productos, ahora).Scan(&e.ID, &e.Destinatarios); err != nil {
			return err
		}
		if e.Destinatarios == 0 {
			return ErrSinDestinatarios
		}
		if _, err := t.Exec(ctx, `INSERT INTO newsletter_entregas (envio_id, tienda_id, suscriptor_id, email)
			SELECT $1, tienda_id, id, lower(email) FROM suscriptores WHERE confirmado_en IS NOT NULL`, e.ID); err != nil {
			return err
		}
		return outbox.Publicar(ctx, t, tienda, "newsletter.tanda", map[string]string{"envio": e.ID})
	})
	return e, err
}
