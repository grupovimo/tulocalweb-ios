// Package mitienda: el editor de Mi tienda (design/PanelOperacion.dc.html). El diseño se guarda
// como borrador y se publica aparte; la tienda online lee el publicado.
package mitienda

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"

	"unicode/utf8"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/tiendas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/auditoria"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

type tx interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
}

var ErrInvalido = errors.New("diseño inválido")

type Seccion struct {
	ID      string `json:"id"`
	Visible bool   `json:"visible"`
}

// Diseno es components/schemas/DisenoTienda. Plantilla va sólo mientras el borrador cambia la de la
// tienda: al publicar pasa a tiendas.plantilla y se saca del diseño. Titulo, Lema, Portada (foto) y
// Logo vacíos: lo de siempre (el nombre de la tienda y la foto de un producto). Compartir*: la vista
// previa al compartir la tienda (vacíos: lo sugerido y la imagen armada con la plantilla).
type Diseno struct {
	Secciones []Seccion `json:"secciones"`
	Paleta    int       `json:"paleta"`
	Plantilla string    `json:"plantilla,omitempty"`
	Titulo    string    `json:"titulo,omitempty"`
	Lema      string    `json:"lema,omitempty"`
	Portada   string    `json:"portada,omitempty"`
	Logo      string    `json:"logo,omitempty"`

	CompartirTitulo      string `json:"compartirTitulo,omitempty"`
	CompartirDescripcion string `json:"compartirDescripcion,omitempty"`
	CompartirFoto        string `json:"compartirFoto,omitempty"`
}

var ids = []string{"hero", "cats", "prods", "revs", "ig", "wa"}

// Inicial: las secciones como vienen en la plantilla.
func Inicial() Diseno {
	d := Diseno{}
	for _, id := range ids {
		d.Secciones = append(d.Secciones, Seccion{ID: id, Visible: true})
	}
	return d
}

// foto: una URL http(s) corta (las fotos subidas a /panel/fotos vuelven con su origen).
func foto(u string) bool {
	return u == "" || (len(u) <= 300 && (strings.HasPrefix(u, "https://") || strings.HasPrefix(u, "http://")) &&
		!strings.ContainsAny(u, " \"'<>"))
}

func (d Diseno) valido() bool {
	if d.Paleta < 0 || d.Paleta > 3 || len(d.Secciones) != len(ids) {
		return false
	}
	if d.Plantilla != "" && !slices.Contains(tiendas.Plantillas, d.Plantilla) {
		return false
	}
	if utf8.RuneCountInString(d.Titulo) > 80 || utf8.RuneCountInString(d.Lema) > 120 || !foto(d.Portada) || !foto(d.Logo) {
		return false
	}
	if utf8.RuneCountInString(d.CompartirTitulo) > 70 || utf8.RuneCountInString(d.CompartirDescripcion) > 160 || !foto(d.CompartirFoto) {
		return false
	}
	vistos := map[string]bool{}
	for _, s := range d.Secciones {
		if !slices.Contains(ids, s.ID) || vistos[s.ID] {
			return false
		}
		vistos[s.ID] = true
	}
	return true
}

type producto struct {
	Nombre string  `json:"nombre"`
	Foto   *string `json:"foto"`
	Precio int64   `json:"precio"`
}

// Leer devuelve el editor con la vista previa armada con los productos de la tienda.
func Leer(ctx context.Context, pool tx, tiendaID, sitio string) (map[string]any, error) {
	var res map[string]any
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var nombre, slug, plantilla string
		var borrador, publicado *Diseno
		if err := t.QueryRow(ctx, `SELECT nombre, slug, plantilla, diseno_borrador, diseno_publicado FROM tiendas WHERE id = $1`,
			tiendaID).Scan(&nombre, &slug, &plantilla, &borrador, &publicado); err != nil {
			return err
		}
		ini := Inicial()
		if borrador == nil {
			borrador = &ini
		}
		if publicado == nil {
			publicado = &ini
		}
		a, _ := json.Marshal(borrador)
		b, _ := json.Marshal(publicado)
		filas, err := t.Query(ctx, `SELECT nombre, foto, precio FROM productos WHERE tienda_id = $1 AND publicado
			ORDER BY creado_en LIMIT 6`, tiendaID)
		if err != nil {
			return err
		}
		prods, err := pgx.CollectRows(filas, pgx.RowToStructByPos[producto])
		if err != nil {
			return err
		}
		var cats []string
		if err := t.QueryRow(ctx, `SELECT coalesce(array_agg(DISTINCT categoria), '{}') FROM productos
			WHERE tienda_id = $1 AND categoria IS NOT NULL`, tiendaID).Scan(&cats); err != nil {
			return err
		}
		fotos := []string{}
		for _, p := range prods {
			if p.Foto != nil {
				fotos = append(fotos, *p.Foto)
			}
		}
		var portada *string
		if borrador.Portada != "" {
			portada = &borrador.Portada
		} else if len(fotos) > 0 {
			portada = &fotos[0]
		}
		titulo := nombre
		if borrador.Titulo != "" {
			titulo = borrador.Titulo
		}
		if borrador.Plantilla != "" {
			plantilla = borrador.Plantilla
		}
		cuatro := prods
		if len(cuatro) > 4 {
			cuatro = cuatro[:4]
		}
		res = map[string]any{
			"plantilla":   strings.ToUpper(plantilla[:1]) + plantilla[1:],
			"plantillaId": plantilla,
			"direccion":   slug + "." + strings.TrimPrefix(strings.TrimPrefix(sitio, "https://"), "http://"),
			"diseno":      borrador,
			"sinPublicar": string(a) != string(b),
			"vista": map[string]any{
				"tienda": nombre, "categorias": cats, "productos": cuatro, "resena": nil, "fotos": fotos,
				"portada": map[string]any{"foto": portada, "titulo": titulo, "lema": borrador.Lema},
				"logo":    borrador.Logo,
			},
		}
		return nil
	})
	return res, err
}

// Guardar el borrador.
func Guardar(ctx context.Context, pool tx, tiendaID string, d Diseno) error {
	if !d.valido() {
		return ErrInvalido
	}
	d.Titulo, d.Lema = strings.TrimSpace(d.Titulo), strings.TrimSpace(d.Lema)
	d.CompartirTitulo, d.CompartirDescripcion = strings.TrimSpace(d.CompartirTitulo), strings.TrimSpace(d.CompartirDescripcion)
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		// La plantilla va en el borrador sólo si cambia la que tiene la tienda.
		_, err := t.Exec(ctx, `UPDATE tiendas SET diseno_borrador = CASE WHEN $2::jsonb->>'plantilla' = plantilla
			THEN $2::jsonb - 'plantilla' ELSE $2::jsonb END WHERE id = $1`, tiendaID, d)
		return err
	})
}

// Publicar copia el borrador a lo que ve la gente.
func Publicar(ctx context.Context, pool tx, tiendaID string) error {
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		// La plantilla del borrador pasa a la tienda y sale de los dos diseños.
		_, err := t.Exec(ctx, `UPDATE tiendas SET
				plantilla = coalesce(nullif(diseno_borrador->>'plantilla', ''), plantilla),
				diseno_borrador = diseno_borrador - 'plantilla',
				diseno_publicado = coalesce(diseno_borrador - 'plantilla', diseno_publicado)
			WHERE id = $1`, tiendaID)
		if err != nil {
			return err
		}
		return auditoria.Registrar(ctx, t, tiendaID, auditoria.Registro{Accion: "mitienda.publicar", Entidad: "diseno",
			Texto: "publicó el diseño de la tienda (Mi tienda)"})
	})
}

// Rutas del panel.
func Rutas(mux *http.ServeMux, pool tx, tienda func(*http.Request) string, sitio string) {
	con := func(h func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			id := tienda(r)
			if id == "" {
				problema.Escribir(w, problema.SinSesion())
				return
			}
			h(w, r, id)
		}
	}
	mux.HandleFunc("GET /v1/panel/mi-tienda", con(func(w http.ResponseWriter, r *http.Request, id string) {
		v, err := Leer(r.Context(), pool, id, sitio)
		if err != nil {
			problema.Escribir(w, problema.Interno())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(v)
	}))
	mux.HandleFunc("PUT /v1/panel/mi-tienda", con(func(w http.ResponseWriter, r *http.Request, id string) {
		var d Diseno
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&d); err != nil {
			problema.Escribir(w, problema.Invalido("Mandá las secciones y los colores."))
			return
		}
		if err := Guardar(r.Context(), pool, id, d); err != nil {
			problema.Escribir(w, problema.Invalido("Revisá el diseño: las secciones, la plantilla, los textos o las fotos."))
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	mux.HandleFunc("POST /v1/panel/mi-tienda/publicar", con(func(w http.ResponseWriter, r *http.Request, id string) {
		if err := Publicar(r.Context(), pool, id); err != nil {
			problema.Escribir(w, problema.Interno())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
}
