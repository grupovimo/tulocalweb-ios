package ventas

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/marca"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

var (
	ErrCarritoInvalido = errors.New("ventas: carrito inválido")
	ErrNoExisteCarrito = errors.New("ventas: ese carrito no existe")
	ErrRecienRecordado = errors.New("ventas: ya se le recordó hace menos de un día")
)

var emailCarrito = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// LineaCarrito: lo que manda el checkout (los precios salen de la tienda, no del navegador).
type LineaCarrito struct {
	ID         string  `json:"id"`
	VarianteID *string `json:"varianteId,omitempty"`
	Cantidad   int     `json:"cantidad"`
}

func nuevoToken() string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// GuardarCarrito guarda (o actualiza) el carrito abierto de ese email. Sólo productos publicados de
// la tienda; el total se calcula con sus precios.
func GuardarCarrito(ctx context.Context, pool tx, tienda, email, nombre string, lineas []LineaCarrito) error {
	email = strings.ToLower(strings.TrimSpace(email))
	nombre = strings.TrimSpace(nombre)
	if !emailCarrito.MatchString(email) || len(email) > 120 || len([]rune(nombre)) > 80 || len(lineas) == 0 || len(lineas) > 50 {
		return ErrCarritoInvalido
	}
	return pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		var total int64
		var validas []LineaCarrito
		for _, l := range lineas {
			if l.Cantidad < 1 || l.Cantidad > 99 {
				return ErrCarritoInvalido
			}
			var precio int64
			err := t.QueryRow(ctx, `SELECT precio FROM productos WHERE id::text = $1 AND publicado`, l.ID).Scan(&precio)
			if err == nil && l.VarianteID != nil {
				// Con variante, su precio (si tiene uno propio); una que ya no se vende, se ignora.
				var propio *int64
				err = t.QueryRow(ctx, `SELECT precio FROM variantes WHERE producto_id::text = $1 AND id::text = $2
					AND activa`, l.ID, *l.VarianteID).Scan(&propio)
				if propio != nil {
					precio = *propio
				}
			}
			if errors.Is(err, pgx.ErrNoRows) {
				continue // un producto (o variante) que ya no está: se ignora
			}
			if err != nil {
				return err
			}
			total += precio * int64(l.Cantidad)
			validas = append(validas, l)
		}
		if len(validas) == 0 {
			return ErrCarritoInvalido
		}
		items, _ := json.Marshal(validas)
		// Si ese email ya pidió no recibir recordatorios, el carrito nuevo nace con la baja.
		_, err := t.Exec(ctx, `INSERT INTO carritos (tienda_id, email, nombre, items, total, token, baja_en)
			VALUES ($1, $2, $3, $4, $5, $6, (SELECT max(baja_en) FROM carritos WHERE tienda_id = $1 AND lower(email) = $2))
			ON CONFLICT (tienda_id, lower(email)) WHERE recuperado_en IS NULL
			DO UPDATE SET items = excluded.items, total = excluded.total, actualizado = now(),
				nombre = CASE WHEN excluded.nombre <> '' THEN excluded.nombre ELSE carritos.nombre END`,
			tienda, email, nombre, items, total, nuevoToken())
		return err
	})
}

// CarritoRecuperado cierra el carrito de ese email cuando compra (lo llama la vidriera), con el
// pedido y lo que se cobró: si antes le llegó un recordatorio, cuenta como recuperado.
func CarritoRecuperado(ctx context.Context, t pgx.Tx, email string, numero int, total int64) error {
	_, err := t.Exec(ctx, `UPDATE carritos SET recuperado_en = now(), pedido_numero = $2, pedido_total = $3
		WHERE lower(email) = lower($1) AND recuperado_en IS NULL`, email, numero, total)
	return err
}

// CarritoPorToken devuelve las líneas para el link del recordatorio (restaura el carrito).
func CarritoPorToken(ctx context.Context, pool tx, tienda, token string) ([]LineaCarrito, error) {
	var out []LineaCarrito
	err := pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		var b []byte
		if err := t.QueryRow(ctx, `SELECT items FROM carritos WHERE token = $1`, token).Scan(&b); err != nil {
			return err
		}
		return json.Unmarshal(b, &out)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoExisteCarrito
	}
	return out, err
}

// CarritoPanel es un carrito como lo muestra el panel (components/schemas/Carritos).
type CarritoPanel struct {
	ID        string    `json:"id"`
	Nombre    string    `json:"nombre"`
	Foto      *string   `json:"foto"`
	Desde     string    `json:"desde"`
	Creado    time.Time `json:"creado"`
	Fotos     []*string `json:"fotos"`
	Cantidad  int       `json:"cantidad"`
	Total     int64     `json:"total"`
	Recordado bool      `json:"recordado"`
}

// Recuperados: lo que se vendió este mes a quienes les llegó un recordatorio (pedidos y lo cobrado).
type Recuperados struct {
	Pedidos int   `json:"pedidos"`
	Monto   int64 `json:"monto"`
}

// CarritosPanel: los abiertos de los últimos 30 días y los números del mes.
func CarritosPanel(ctx context.Context, pool tx, tienda string, ahora time.Time) (cs []CarritoPanel, abiertos int, monto, recuperado int64, tasa float64, err error) {
	cs, abiertos, monto, r, tasa, err := CarritosYRecuperados(ctx, pool, tienda, ahora)
	return cs, abiertos, monto, r.Monto, tasa, err
}

// CarritosYRecuperados es CarritosPanel con los pedidos recuperados.
func CarritosYRecuperados(ctx context.Context, pool tx, tienda string, ahora time.Time) (cs []CarritoPanel, abiertos int, monto int64, rec Recuperados, tasa float64, err error) {
	cs = []CarritoPanel{}
	err = pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		filas, err := t.Query(ctx, `SELECT c.id::text, coalesce(nullif(c.nombre, ''), split_part(c.email, '@', 1)), c.creado, c.items, c.total,
				c.recordado_en IS NOT NULL
			FROM carritos c WHERE c.tienda_id = $1 AND c.recuperado_en IS NULL AND c.actualizado > $2
			ORDER BY c.actualizado DESC LIMIT 200`, tienda, ahora.Add(-30*24*time.Hour))
		if err != nil {
			return err
		}
		for filas.Next() {
			var c CarritoPanel
			var items []byte
			if err := filas.Scan(&c.ID, &c.Nombre, &c.Creado, &items, &c.Total, &c.Recordado); err != nil {
				filas.Close()
				return err
			}
			var ls []LineaCarrito
			_ = json.Unmarshal(items, &ls)
			for _, l := range ls {
				c.Cantidad += l.Cantidad
			}
			c.Desde = "Tienda online"
			cs = append(cs, c)
			abiertos++
			monto += c.Total
		}
		filas.Close()
		if err := filas.Err(); err != nil {
			return err
		}
		// Fotos de los productos de cada carrito (hasta 3).
		for i := range cs {
			var items []byte
			if err := t.QueryRow(ctx, `SELECT items FROM carritos WHERE id::text = $1`, cs[i].ID).Scan(&items); err != nil {
				return err
			}
			var ls []LineaCarrito
			_ = json.Unmarshal(items, &ls)
			for _, l := range ls {
				if len(cs[i].Fotos) == 3 {
					break
				}
				var f *string
				_ = t.QueryRow(ctx, `SELECT foto FROM productos WHERE id::text = $1`, l.ID).Scan(&f)
				cs[i].Fotos = append(cs[i].Fotos, f)
			}
			if len(cs[i].Fotos) > 0 {
				cs[i].Foto = cs[i].Fotos[0]
			}
			if cs[i].Fotos == nil {
				cs[i].Fotos = []*string{}
			}
		}
		inicioMes := time.Date(ahora.Year(), ahora.Month(), 1, 0, 0, 0, 0, ahora.Location())
		// Recuperado este mes: las compras de quienes recibieron un recordatorio antes de comprar, con
		// lo que se cobró en el pedido (los carritos de antes de 00084, con lo que tenían).
		var recordados int
		if err := t.QueryRow(ctx, `SELECT count(*) FILTER (WHERE recordado_en IS NOT NULL AND (recuperado_en IS NULL OR recordado_en <= recuperado_en)),
				count(*) FILTER (WHERE recuperado_en >= $2 AND recordado_en <= recuperado_en),
				coalesce(sum(coalesce(pedido_total, total)) FILTER (WHERE recuperado_en >= $2 AND recordado_en <= recuperado_en), 0)
			FROM carritos WHERE tienda_id = $1 AND (creado >= $2 OR recuperado_en >= $2)`, tienda, inicioMes).
			Scan(&recordados, &rec.Pedidos, &rec.Monto); err != nil {
			return err
		}
		if recordados > 0 {
			tasa = float64(rec.Pedidos) * 100 / float64(recordados)
			if tasa > 100 {
				tasa = 100
			}
		}
		return nil
	})
	return
}

// RecordarCarrito encola el mail con el link que restaura el carrito. Uno por día como mucho.
func RecordarCarrito(ctx context.Context, pool tx, tienda, id string, ahora time.Time) error {
	return pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		var recordado *time.Time
		err := t.QueryRow(ctx, `SELECT recordado_en FROM carritos WHERE id::text = $1 AND recuperado_en IS NULL FOR UPDATE`, id).Scan(&recordado)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoExisteCarrito
		}
		if err != nil {
			return err
		}
		if recordado != nil && ahora.Sub(*recordado) < 24*time.Hour {
			return ErrRecienRecordado
		}
		if _, err := t.Exec(ctx, `UPDATE carritos SET recordado_en = $2, recordatorios = recordatorios + 1 WHERE id::text = $1`, id, ahora); err != nil {
			return err
		}
		return outbox.Publicar(ctx, t, tienda, "carrito.recordar", map[string]string{"id": id})
	})
}

// Mails de carritos (el worker): "¿Te olvidaste algo?" con el link que restaura el carrito.
type MailsCarritos struct {
	Pool   tx
	Correo correo.Correo
	Sitio  string
}

func (m *MailsCarritos) Manejadores() map[string]outbox.Manejador {
	return map[string]outbox.Manejador{"carrito.recordar": m.recordar}
}

func (m *MailsCarritos) recordar(ctx context.Context, ev outbox.Evento) error {
	var d struct {
		ID    string `json:"id"`
		Paso  int    `json:"paso"`
		Cupon string `json:"cupon"`
	}
	if json.Unmarshal(ev.Datos, &d) != nil || m.Correo == nil {
		return nil
	}
	var email, nombre, token string
	var items []byte
	var recuperado, baja bool
	var mc marca.Marca
	var nombres []string
	var cupon *cuponMail
	err := m.Pool.EnTienda(ctx, ev.TiendaID, func(t pgx.Tx) error {
		if err := t.QueryRow(ctx, `SELECT c.email, c.nombre, c.token, c.items, c.recuperado_en IS NOT NULL, c.baja_en IS NOT NULL
			FROM carritos c WHERE c.id::text = $1`, d.ID).
			Scan(&email, &nombre, &token, &items, &recuperado, &baja); err != nil {
			return err
		}
		var err error
		if mc, err = marca.De(ctx, t, ev.TiendaID, m.Sitio); err != nil {
			return err
		}
		var ls []LineaCarrito
		_ = json.Unmarshal(items, &ls)
		for _, l := range ls {
			var n string
			if t.QueryRow(ctx, `SELECT nombre FROM productos WHERE id::text = $1`, l.ID).Scan(&n) == nil {
				nombres = append(nombres, fmt.Sprintf("%d × %s", l.Cantidad, n))
			}
		}
		if d.Cupon != "" {
			var c cuponMail
			if t.QueryRow(ctx, `SELECT codigo, descripcion, termina FROM cupones WHERE codigo = $1 AND origen = $2`,
				d.Cupon, "carrito:"+d.ID).Scan(&c.codigo, &c.descripcion, &c.vence) == nil {
				cupon = &c
			}
		}
		return nil
	})
	if errors.Is(err, pgx.ErrNoRows) || recuperado || baja {
		return nil // compró o se dio de baja mientras tanto: no se manda
	}
	if err != nil {
		return err
	}
	return m.Correo.Enviar(ctx, MailRecordatorio(m.Sitio, mc, d.Paso, email, nombre, token, nombres, cupon))
}

type cuponMail struct {
	codigo, descripcion string
	vence               *time.Time
}

var argentina = time.FixedZone("ART", -3*3600)

// MailRecordatorio: el recordatorio del carrito con la marca de la tienda. paso 0 es el que se manda
// a mano; 1 a 3, los automáticos (cada uno con su texto). Con cupón, el código destacado y el link
// lo deja aplicado. Siempre con el link de baja y los encabezados de baja con un clic (RFC 8058).
func MailRecordatorio(sitio string, mc marca.Marca, paso int, email, nombre, token string, lineas []string, c *cuponMail) correo.Mensaje {
	base := fmt.Sprintf("%s/t/%s/carrito", strings.TrimRight(sitio, "/"), mc.Slug)
	link := base + "/recuperar/" + token
	baja := base + "/baja/" + token
	hola := "Hola"
	if nombre != "" {
		hola += ", " + strings.Split(nombre, " ")[0]
	}
	titulo, asunto := "¿Te olvidaste algo?", "¿Te olvidaste algo en "+mc.Nombre+"?"
	intro := hola + ". Dejaste esto en tu carrito de " + mc.Nombre + ":"
	switch paso {
	case 2:
		titulo, asunto = "Tu carrito te espera", "Tu carrito de "+mc.Nombre+" te espera"
		intro = hola + ". Te guardamos lo que elegiste en " + mc.Nombre + ":"
	case 3:
		titulo, asunto = "Última vuelta por tu carrito", "Última vuelta por tu carrito de "+mc.Nombre
		intro = hola + ". Lo que dejaste en " + mc.Nombre + " sigue guardado, pero no por mucho más:"
	}
	mail := marca.Mail{Parrafos: []string{intro}, Titulo: titulo, Lista: lineas, Boton: "Volver a mi carrito", Link: link,
		Pie: "Te escribimos porque dejaste tu email en el checkout de " + mc.Nombre + ".", Baja: baja}
	if c != nil {
		asunto = "Te dejamos " + strings.ToLower(strings.SplitN(c.descripcion, " · ", 2)[0]) + " para tu carrito de " + mc.Nombre
		mail.CodigoIntro = "Para que lo termines, te dejamos un cupón de " + strings.ToLower(strings.SplitN(c.descripcion, " · ", 2)[0]) + ":"
		mail.Codigo = c.codigo
		if c.vence != nil {
			v := c.vence.In(argentina)
			meses := []string{"enero", "febrero", "marzo", "abril", "mayo", "junio", "julio", "agosto", "septiembre", "octubre", "noviembre", "diciembre"}
			mail.CodigoNota = fmt.Sprintf("Vale hasta el %d de %s. Es de un solo uso.", v.Day(), meses[v.Month()-1])
		}
		mail.Link = link + "?cupon=" + c.codigo
		mail.Boton = "Usar mi cupón"
	}
	h, t := mc.Armar(mail)
	return correo.Mensaje{Para: email, Asunto: asunto, Texto: t, HTML: h, DeNombre: mc.Nombre, ResponderA: mc.ResponderA,
		Encabezados: map[string]string{"List-Unsubscribe": "<" + baja + ">", "List-Unsubscribe-Post": "List-Unsubscribe=One-Click"}}
}

// RecordarTodos: los abiertos de la semana que no se recordaron en el último día.
func RecordarTodos(ctx context.Context, pool tx, tienda string, ahora time.Time) (int, error) {
	var ids []string
	err := pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		filas, err := t.Query(ctx, `SELECT id::text FROM carritos WHERE recuperado_en IS NULL AND actualizado > $1
			AND (recordado_en IS NULL OR recordado_en < $2) LIMIT 200`, ahora.Add(-7*24*time.Hour), ahora.Add(-24*time.Hour))
		if err != nil {
			return err
		}
		defer filas.Close()
		for filas.Next() {
			var id string
			if err := filas.Scan(&id); err != nil {
				return err
			}
			ids = append(ids, id)
		}
		return filas.Err()
	})
	n := 0
	for _, id := range ids {
		if RecordarCarrito(ctx, pool, tienda, id, ahora) == nil {
			n++
		}
	}
	return n, err
}
