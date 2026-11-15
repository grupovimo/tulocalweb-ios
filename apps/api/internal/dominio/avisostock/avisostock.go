// Package avisostock: «Avisame cuando vuelva» (docs/decisiones.md §60). En la ficha de un producto o
// una variante sin stock, quien compra deja su email (y, si quiere, su WhatsApp) con su
// consentimiento. Cuando el stock pasa de 0 a más de 0, por el camino que sea, el trigger de la
// migración 00083 encola `stock.repuesto` en la misma transacción y el worker manda el mail con la
// marca de la tienda y el link al producto, una sola vez por pedido de aviso.
package avisostock

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/marca"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/limite"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

type tx interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
	EnTx(context.Context, func(pgx.Tx) error) error
}

// Topes: cuántos pendientes por producto y cuántos mails manda cada evento (si quedan, encola otro).
const (
	TopeAvisos = 20000
	Tanda      = 50
)

var (
	ErrNoExiste          = errors.New("avisostock: ese producto no existe")
	ErrHayStock          = errors.New("avisostock: hay stock")
	ErrLleno             = errors.New("avisostock: no entran más avisos")
	emailValido          = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	whatsappValido       = regexp.MustCompile(`^\+?[0-9 ()-]{6,24}$`)
	uuidValido           = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	ErrSinConsentimiento = ErrInvalido{"Marcá que querés que te escribamos."}
)

// ErrInvalido: un dato que no va (el mensaje es para la persona).
type ErrInvalido struct{ Motivo string }

func (e ErrInvalido) Error() string { return e.Motivo }

// Pedido es lo que manda la ficha.
type Pedido struct {
	VarianteID     *string `json:"varianteId"`
	Email          string  `json:"email"`
	Whatsapp       string  `json:"whatsapp"`
	Consentimiento bool    `json:"consentimiento"`
}

// Anotar guarda el pedido de aviso. Repetido no suma (si trae WhatsApp, lo actualiza); con stock, no
// se anota (ErrHayStock: la ficha ya deja comprar).
func Anotar(ctx context.Context, pool tx, slug, productoID string, p Pedido, ahora time.Time) error {
	email := strings.ToLower(strings.TrimSpace(p.Email))
	if !emailValido.MatchString(email) || len(email) > 200 || strings.ContainsAny(email, "\r\n<>,;") {
		return ErrInvalido{"Revisá el email."}
	}
	wa := strings.TrimSpace(p.Whatsapp)
	if wa != "" && !whatsappValido.MatchString(wa) {
		return ErrInvalido{"Revisá el WhatsApp."}
	}
	if !p.Consentimiento {
		return ErrSinConsentimiento
	}
	if !uuidValido.MatchString(productoID) || (p.VarianteID != nil && !uuidValido.MatchString(*p.VarianteID)) {
		return ErrNoExiste
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
	var whatsapp *string
	if wa != "" {
		whatsapp = &wa
	}
	return pool.EnTienda(ctx, *tienda, func(t pgx.Tx) error {
		var stock, cuantos int
		var conOpciones bool
		err := t.QueryRow(ctx, `SELECT p.stock, p.opciones <> '[]'::jsonb, (SELECT count(*) FROM avisos_stock a
				WHERE a.tienda_id = p.tienda_id AND a.producto_id = p.id AND a.avisado_en IS NULL)
			FROM productos p WHERE p.tienda_id = $1 AND p.id = $2::uuid AND p.publicado`, *tienda, productoID).Scan(&stock, &conOpciones, &cuantos)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoExiste
		}
		if err != nil {
			return err
		}
		if p.VarianteID != nil && conOpciones {
			err := t.QueryRow(ctx, `SELECT stock FROM variantes WHERE tienda_id = $1 AND producto_id = $2::uuid AND id = $3::uuid AND activa`,
				*tienda, productoID, *p.VarianteID).Scan(&stock)
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNoExiste
			}
			if err != nil {
				return err
			}
		} else {
			p.VarianteID = nil
		}
		if stock > 0 {
			return ErrHayStock
		}
		if cuantos >= TopeAvisos {
			return ErrLleno
		}
		_, err = t.Exec(ctx, `INSERT INTO avisos_stock (tienda_id, producto_id, variante_id, email, whatsapp, consentimiento_en)
			VALUES ($1, $2::uuid, $3::uuid, $4, $5, $6)
			ON CONFLICT (tienda_id, producto_id, coalesce(variante_id, '00000000-0000-0000-0000-000000000000'::uuid), email)
				WHERE avisado_en IS NULL
			DO UPDATE SET whatsapp = coalesce(excluded.whatsapp, avisos_stock.whatsapp)`,
			*tienda, productoID, p.VarianteID, email, whatsapp, ahora)
		return err
	})
}

// Esperando: cuántas personas esperan cada producto (todas sus variantes), con la tienda fijada.
func Esperando(ctx context.Context, t pgx.Tx, tiendaID string) (map[string]int, error) {
	filas, err := t.Query(ctx, `SELECT producto_id::text, count(*) FROM avisos_stock WHERE tienda_id = $1 AND avisado_en IS NULL
		GROUP BY producto_id`, tiendaID)
	if err != nil {
		return nil, err
	}
	defer filas.Close()
	out := map[string]int{}
	for filas.Next() {
		var id string
		var n int
		if err := filas.Scan(&id, &n); err != nil {
			return nil, err
		}
		out[id] = n
	}
	return out, filas.Err()
}

// Servicio manda los avisos (worker).
type Servicio struct {
	Pool   tx
	Correo correo.Correo
	// Sitio público, para el link del producto.
	Sitio string
}

// Manejadores del outbox para el worker.
func (s *Servicio) Manejadores() map[string]outbox.Manejador {
	return map[string]outbox.Manejador{"stock.repuesto": s.avisar}
}

type aviso struct {
	id, email, variante string
}

type producto struct {
	nombre, slug string
	precio       int64
	foto         *string
}

// avisar manda una tanda de «Volvió». Sólo a quien espera algo que hoy tiene stock (el producto o su
// variante): si lo vendieron de nuevo o lo pausaron antes de que salga, espera la próxima. Cada mail
// se marca enviado apenas sale: un reintento sigue con los que faltan y nadie lo recibe dos veces.
func (s *Servicio) avisar(ctx context.Context, ev outbox.Evento) error {
	var d struct {
		Producto string `json:"producto"`
	}
	if json.Unmarshal(ev.Datos, &d) != nil || !uuidValido.MatchString(d.Producto) || s.Correo == nil {
		return nil
	}
	var p producto
	var m marca.Marca
	var avisos []aviso
	err := s.Pool.EnTienda(ctx, ev.TiendaID, func(t pgx.Tx) error {
		err := t.QueryRow(ctx, `SELECT nombre, slug, CASE WHEN oferta_precio < precio AND (oferta_hasta IS NULL OR oferta_hasta > now())
				THEN oferta_precio ELSE precio END, foto
			FROM productos WHERE tienda_id = $1 AND id = $2::uuid AND publicado`, ev.TiendaID, d.Producto).Scan(&p.nombre, &p.slug, &p.precio, &p.foto)
		if err != nil {
			return err
		}
		if m, err = marca.De(ctx, t, ev.TiendaID, s.Sitio); err != nil {
			return err
		}
		filas, err := t.Query(ctx, `SELECT a.id::text, a.email, coalesce(v.valores::text, '')
			FROM avisos_stock a
			JOIN productos p ON p.tienda_id = a.tienda_id AND p.id = a.producto_id
			LEFT JOIN variantes v ON v.tienda_id = a.tienda_id AND v.producto_id = a.producto_id AND v.id = a.variante_id
			WHERE a.tienda_id = $1 AND a.producto_id = $2::uuid AND a.avisado_en IS NULL
				AND ((a.variante_id IS NULL AND p.stock > 0) OR (v.activa AND v.stock > 0))
			ORDER BY a.creado_en LIMIT $3`, ev.TiendaID, d.Producto, Tanda+1)
		if err != nil {
			return err
		}
		defer filas.Close()
		for filas.Next() {
			var a aviso
			if err := filas.Scan(&a.id, &a.email, &a.variante); err != nil {
				return err
			}
			avisos = append(avisos, a)
		}
		return filas.Err()
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil // lo pausaron o lo borraron: no se avisa
	}
	if err != nil {
		return err
	}
	quedan := len(avisos) > Tanda
	if quedan {
		avisos = avisos[:Tanda]
	}
	for _, a := range avisos {
		if err := s.Correo.Enviar(ctx, MailVolvio(s.Sitio, m, p.nombre, p.slug, variante(a.variante), p.precio, p.foto, a.email)); err != nil {
			return err
		}
		if err := s.Pool.EnTienda(ctx, ev.TiendaID, func(t pgx.Tx) error {
			_, err := t.Exec(ctx, `UPDATE avisos_stock SET avisado_en = now() WHERE tienda_id = $1 AND id = $2::uuid`, ev.TiendaID, a.id)
			return err
		}); err != nil {
			return err
		}
	}
	if !quedan {
		return nil
	}
	return s.Pool.EnTienda(ctx, ev.TiendaID, func(t pgx.Tx) error {
		return outbox.Publicar(ctx, t, ev.TiendaID, "stock.repuesto", d)
	})
}

// variante: «Talle M · Negro» a partir de los valores guardados ({"Talle": "M", "Color": "Negro"}).
func variante(valores string) string {
	if valores == "" {
		return ""
	}
	var m map[string]string
	if json.Unmarshal([]byte(valores), &m) != nil {
		return ""
	}
	partes := make([]string, 0, len(m))
	for _, v := range m {
		partes = append(partes, v)
	}
	// El orden de un mapa no es fijo: se ordena para que el mail diga siempre lo mismo.
	sort.Strings(partes)
	return strings.Join(partes, " · ")
}

func pesos(n int64) string {
	s := fmt.Sprint(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "." + s[i:]
	}
	return "$ " + s
}

// MailVolvio: «Volvió X», con la foto, el precio y el botón a la ficha, con la marca de la tienda.
func MailVolvio(sitio string, m marca.Marca, nombre, slug, variante string, precio int64, foto *string, para string) correo.Mensaje {
	link := fmt.Sprintf("%s/t/%s/producto/%s", strings.TrimRight(sitio, "/"), m.Slug, slug)
	que := nombre
	if variante != "" {
		que += " (" + variante + ")"
	}
	f := ""
	if foto != nil {
		f = marca.Absoluta(sitio, *foto)
	}
	h, t := m.Armar(marca.Mail{
		Titulo:   "Volvió " + nombre,
		Parrafos: []string{fmt.Sprintf("Nos pediste que te avisemos: %s ya tiene stock de nuevo en %s.", que, m.Nombre)},
		Foto:     f,
		Detalle:  fmt.Sprintf("%s · %s", que, pesos(precio)),
		Boton:    "Comprarlo",
		Link:     link,
		Pie:      "Te escribimos sólo por este aviso: no te vamos a mandar más mails por esto.",
	})
	return correo.Mensaje{Para: para, Asunto: fmt.Sprintf("Volvió %s a %s", nombre, m.Nombre), Texto: t, HTML: h,
		DeNombre: m.Nombre, ResponderA: m.ResponderA}
}

// Rutas: el pedido de aviso público, con tope de 10 por hora por IP (como el de lanzamientos).
func Rutas(mux *http.ServeMux, pool tx, ip func(*http.Request) string) {
	mux.HandleFunc("POST /v1/publico/tiendas/{slug}/productos/{id}/aviso-stock", func(w http.ResponseWriter, r *http.Request) {
		if c, ok := pool.(limite.Consultor); ok && ip != nil {
			if ok, err := limite.Permitido(r.Context(), c, "tienda:aviso-stock:"+ip(r), 10, time.Hour); err == nil && !ok {
				problema.Escribir(w, &problema.Problema{Estado: http.StatusTooManyRequests, Codigo: "demasiados",
					Titulo: "Demasiados envíos", Detalle: "Probá de nuevo en un rato."})
				return
			}
		}
		var p Pedido
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 2<<10)).Decode(&p) != nil {
			problema.Escribir(w, problema.Invalido("Revisá el email."))
			return
		}
		err := Anotar(r.Context(), pool, r.PathValue("slug"), r.PathValue("id"), p, time.Now())
		var inv ErrInvalido
		switch {
		case err == nil:
			w.WriteHeader(http.StatusNoContent)
		case errors.As(err, &inv):
			problema.Escribir(w, problema.Invalido(inv.Motivo))
		case errors.Is(err, ErrNoExiste):
			problema.Escribir(w, problema.NoEncontrado("Ese producto ya no está publicado."))
		case errors.Is(err, ErrHayStock):
			problema.Escribir(w, problema.Conflicto("hay_stock", "Ya hay stock: lo podés comprar."))
		case errors.Is(err, ErrLleno):
			problema.Escribir(w, problema.Conflicto("lleno", "No podemos anotar más avisos para este producto."))
		default:
			problema.Escribir(w, problema.Interno())
		}
	})
}
