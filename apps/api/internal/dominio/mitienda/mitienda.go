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

// Seccion es del diseño v1 (las 6 fijas): se lee para migrar.
type Seccion struct {
	ID      string `json:"id"`
	Visible bool   `json:"visible"`
}

// Diseno es components/schemas/DisenoTienda (v2: docs/especificaciones/tienda-bloques.md). Plantilla
// va sólo mientras el borrador cambia la de la tienda: al publicar pasa a tiendas.plantilla y se saca
// del diseño. Titulo, Lema, Portada (foto) y Logo vacíos: lo de siempre (el nombre de la tienda y la
// foto de un producto). Compartir*: la vista previa al compartir la tienda (vacíos: lo sugerido y la
// imagen armada con la plantilla). Sin Version (o 1) es un diseño v1, con Secciones: Migrar lo pasa.
type Diseno struct {
	Version   int    `json:"version,omitempty"`
	Paleta    int    `json:"paleta"`
	Plantilla string `json:"plantilla,omitempty"`
	Titulo    string `json:"titulo,omitempty"`
	Lema      string `json:"lema,omitempty"`
	Portada   string `json:"portada,omitempty"`
	Logo      string `json:"logo,omitempty"`

	CompartirTitulo      string `json:"compartirTitulo,omitempty"`
	CompartirDescripcion string `json:"compartirDescripcion,omitempty"`
	CompartirFoto        string `json:"compartirFoto,omitempty"`

	// Estilo: colores, tipografías, logos, redes, menú, barra y opciones de la dirección (estilo.go).
	Estilo *Estilo `json:"estilo,omitempty"`
	// Inicio: los bloques de la portada, en orden.
	Inicio  []Bloque `json:"inicio"`
	Paginas *Paginas `json:"paginas,omitempty"`
	// Ventanas: los popups de la tienda (ventanas.go), en orden: si varias califican, gana la primera.
	Ventanas []Ventana `json:"ventanas,omitempty"`

	// Secciones: sólo v1.
	Secciones []Seccion `json:"secciones,omitempty"`
}

// Bloque: uno de bloques.json › bloques. Datos son sus campos; lo que falta toma lo de por defecto o
// lo sugerido (lo resuelve quien dibuja).
type Bloque struct {
	ID       string         `json:"id"`
	Tipo     string         `json:"tipo"`
	Variante string         `json:"variante"`
	Oculto   bool           `json:"oculto,omitempty"`
	Datos    map[string]any `json:"datos"`
}

// Paginas: las opciones de las otras páginas (bloques.json › paginas). Vacío: lo de por defecto.
type Paginas struct {
	Catalogo *PaginaCatalogo `json:"catalogo,omitempty"`
	Producto *PaginaProducto `json:"producto,omitempty"`
	Carrito  *PaginaCarrito  `json:"carrito,omitempty"`
	Nosotros *PaginaNosotros `json:"nosotros,omitempty"`
}

type PaginaCatalogo struct {
	Variante string `json:"variante,omitempty"`
	Filtros  string `json:"filtros,omitempty"`
	Columnas int    `json:"columnas,omitempty"`
}

type PaginaProducto struct {
	Variante      string `json:"variante,omitempty"`
	Relacionados  *bool  `json:"relacionados,omitempty"`
	CalcularEnvio *bool  `json:"calcularEnvio,omitempty"`
	// Combos: mostrar en la ficha el combo del que forma parte el producto («Completá el look»).
	Combos *bool `json:"combos,omitempty"`
	// GuiaTalles: { foto?, tabla? } (bloques.json › paginas.producto › guiaTalles).
	GuiaTalles map[string]any `json:"guiaTalles,omitempty"`
	Bloques    []Bloque       `json:"bloques"`
}

type PaginaCarrito struct {
	Variante         string `json:"variante,omitempty"`
	EnvioGratisDesde *bool  `json:"envioGratisDesde,omitempty"`
}

type PaginaNosotros struct {
	Bloques []Bloque `json:"bloques"`
}

// Inicial: el diseño v2 de partida del preset de la plantilla (sus bloques, sin datos propios).
func Inicial(plantilla string) Diseno {
	p := presetDe(plantilla)
	d := Diseno{Version: 2, Inicio: []Bloque{}}
	for _, tipo := range p.Orden {
		d.Inicio = append(d.Inicio, Bloque{ID: tipo, Tipo: tipo, Variante: p.Variantes[tipo], Datos: map[string]any{}})
	}
	return d
}

// foto: una URL http(s) corta (las fotos subidas a /panel/fotos vuelven con su origen).
func foto(u string) bool {
	return u == "" || (len(u) <= 300 && (strings.HasPrefix(u, "https://") || strings.HasPrefix(u, "http://")) &&
		!strings.ContainsAny(u, " \"'<>"))
}

// Valido: un diseño v2 completo según bloques.json (los v1 no se guardan más: se migran al leer).
func (d Diseno) Valido() bool {
	if d.Version != 2 || len(d.Secciones) > 0 || d.Estilo.v1() || d.Paleta < 0 || d.Paleta > 3 {
		return false
	}
	if d.Plantilla != "" && !slices.Contains(tiendas.Plantillas, d.Plantilla) {
		return false
	}
	if utf8.RuneCountInString(d.Titulo) > 80 || utf8.RuneCountInString(d.Lema) > 120 || strings.ContainsAny(d.Titulo+d.Lema, "<>") ||
		!foto(d.Portada) || !foto(d.Logo) {
		return false
	}
	if !d.Estilo.valido() {
		return false
	}
	if utf8.RuneCountInString(d.CompartirTitulo) > 70 || utf8.RuneCountInString(d.CompartirDescripcion) > 160 || !foto(d.CompartirFoto) {
		return false
	}
	vistos := map[string]bool{}
	return bloquesValidos(d.Inicio, vistos) && d.Paginas.valido(vistos) && ventanasValidas(d.Ventanas)
}

// leerDiseno: el jsonb de la base como diseño v2 (los v1 y los nulos se migran con la plantilla).
func leerDiseno(crudo []byte, plantilla string) Diseno {
	var d *Diseno
	if len(crudo) > 0 && string(crudo) != "null" {
		d = &Diseno{}
		if json.Unmarshal(crudo, d) != nil {
			d = nil
		}
	}
	return Migrar(d, plantilla)
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
		var crudoB, crudoP []byte
		if err := t.QueryRow(ctx, `SELECT nombre, slug, plantilla, diseno_borrador, diseno_publicado FROM tiendas WHERE id = $1`,
			tiendaID).Scan(&nombre, &slug, &plantilla, &crudoB, &crudoP); err != nil {
			return err
		}
		// Un borrador v1 que cambia la plantilla se migra con la suya (Migrar la toma del diseño).
		b0, p0 := leerDiseno(crudoB, plantilla), leerDiseno(crudoP, plantilla)
		borrador, publicado := &b0, &p0
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

// Guardar el borrador (sólo v2 válido).
func Guardar(ctx context.Context, pool tx, tiendaID string, d Diseno) error {
	if !d.Valido() {
		return ErrInvalido
	}
	if d.Inicio == nil {
		d.Inicio = []Bloque{}
	}
	limpiarBloques(d.Inicio)
	if d.Paginas != nil {
		if d.Paginas.Producto != nil {
			limpiarBloques(d.Paginas.Producto.Bloques)
		}
		if d.Paginas.Nosotros != nil {
			limpiarBloques(d.Paginas.Nosotros.Bloques)
		}
	}
	for i := range d.Ventanas {
		if d.Ventanas[i].Datos == nil {
			d.Ventanas[i].Datos = map[string]any{}
		}
		if d.Ventanas[i].Paginas == nil {
			d.Ventanas[i].Paginas = []string{}
		}
		d.Ventanas[i].Nombre = strings.TrimSpace(d.Ventanas[i].Nombre)
	}
	d.Titulo, d.Lema = strings.TrimSpace(d.Titulo), strings.TrimSpace(d.Lema)
	d.CompartirTitulo, d.CompartirDescripcion = strings.TrimSpace(d.CompartirTitulo), strings.TrimSpace(d.CompartirDescripcion)
	d.Estilo = d.Estilo.limpiar()
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		// La plantilla va en el borrador sólo si cambia la que tiene la tienda.
		_, err := t.Exec(ctx, `UPDATE tiendas SET diseno_borrador = CASE WHEN $2::jsonb->>'plantilla' = plantilla
			THEN $2::jsonb - 'plantilla' ELSE $2::jsonb END WHERE id = $1`, tiendaID, d)
		return err
	})
}

// limpiarBloques: datos nunca null (la web los recorre).
func limpiarBloques(l []Bloque) {
	for i := range l {
		if l[i].Datos == nil {
			l[i].Datos = map[string]any{}
		}
	}
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
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, int64(Catalogo.Limites.BytesDiseno))).Decode(&d); err != nil {
			problema.Escribir(w, problema.Invalido("Mandá el diseño completo (hasta 64 KB)."))
			return
		}
		if err := Guardar(r.Context(), pool, id, d); err != nil {
			problema.Escribir(w, problema.Invalido("Revisá el diseño: los bloques, la plantilla, los textos, los links o las fotos."))
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
