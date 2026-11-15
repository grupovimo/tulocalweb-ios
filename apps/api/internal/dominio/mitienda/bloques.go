package mitienda

// El catálogo de bloques y los presets viven en packages/contrato (fuera del módulo Go): acá van
// copias embebidas. bloques_test.go falla si una copia difiere del original, así nadie edita uno solo.
//go:generate cp ../../../../../packages/contrato/bloques.json bloques.json
//go:generate cp ../../../../../packages/contrato/presets.json presets.json

import (
	_ "embed"
	"encoding/json"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	//go:embed bloques.json
	bloquesJSON []byte
	//go:embed presets.json
	presetsJSON []byte
)

// campo es un campo de un bloque (bloques.json › bloques[].campos), de una página o de la barra.
type campo struct {
	ID        string   `json:"id"`
	Tipo      string   `json:"tipo"`
	Max       *int     `json:"max"`
	Min       *int     `json:"min"`
	Opciones  []string `json:"opciones"`
	Requerido bool     `json:"requerido"`
	Item      []campo  `json:"item"`
}

type tipoBloque struct {
	Tipo      string   `json:"tipo"`
	Max       int      `json:"max"`
	Variantes []string `json:"variantes"`
	Campos    []campo  `json:"campos"`
}

type opciones struct {
	Opciones []string `json:"opciones"`
}

type paginaCatalogo struct {
	Variantes []string `json:"variantes"`
	Campos    []campo  `json:"campos"`
	Bloques   bool     `json:"bloques"`
}

// Catalogo es bloques.json: límites, opciones del estilo, páginas y bloques.
var Catalogo = func() (c struct {
	Limites struct {
		BloquesPorPagina int `json:"bloquesPorPagina"`
		ItemsPorLista    int `json:"itemsPorLista"`
		BytesDiseno      int `json:"bytesDiseno"`
		FilasTabla       int `json:"filasTabla"`
		ColumnasTabla    int `json:"columnasTabla"`
		LetrasCelda      int `json:"letrasCelda"`
		Ventanas         int `json:"ventanas"`
	} `json:"limites"`
	Estilo           map[string]opciones       `json:"estilo"`
	Fondos           opciones                  `json:"fondos"`
	FuentesProductos opciones                  `json:"fuentesProductos"`
	Iconos           []string                  `json:"iconos"`
	Paginas          map[string]paginaCatalogo `json:"paginas"`
	Bloques          []tipoBloque              `json:"bloques"`
	BarraAnuncios    paginaCatalogo            `json:"barraAnuncios"`
	Conversion       paginaCatalogo            `json:"conversion"`
	Ventanas         catalogoVentanas          `json:"ventanas"`
}) {
	if err := json.Unmarshal(bloquesJSON, &c); err != nil {
		panic("mitienda: bloques.json: " + err.Error())
	}
	return c
}()

// Preset es una de las 17 plantillas como punto de partida (presets.json).
type Preset struct {
	Direccion   string            `json:"direccion"`
	Estilo      map[string]string `json:"estilo"`
	Orden       []string          `json:"orden"`
	Variantes   map[string]string `json:"variantes"`
	Gastronomia bool              `json:"gastronomia"`
}

// Presets por id de plantilla.
var Presets = func() map[string]Preset {
	var p struct {
		Presets map[string]Preset `json:"presets"`
	}
	if err := json.Unmarshal(presetsJSON, &p); err != nil {
		panic("mitienda: presets.json: " + err.Error())
	}
	return p.Presets
}()

// presetDe: el de la plantilla, o el de ceibo si no existe (no debería pasar: tiendas.Plantillas).
func presetDe(plantilla string) Preset {
	if p, ok := Presets[plantilla]; ok {
		return p
	}
	return Presets["ceibo"]
}

func tipoDe(tipo string) *tipoBloque {
	for i := range Catalogo.Bloques {
		if Catalogo.Bloques[i].Tipo == tipo {
			return &Catalogo.Bloques[i]
		}
	}
	return nil
}

var (
	reIDBloque = regexp.MustCompile(`^[a-z0-9]{6,16}$`)
	reProducto = regexp.MustCompile(`^[0-9A-Za-z-]{1,64}$`)
	hostsVideo = []string{"youtube.com", "www.youtube.com", "youtu.be", "vimeo.com", "www.vimeo.com", "player.vimeo.com"}
)

// linkValido: una página de la tienda, «cat:<categoría>», «prod:<id>» o un link https (las mismas
// reglas que el menú, más prod:).
func linkValido(a string) bool {
	if slices.Contains(paginasMenu, a) {
		return true
	}
	if c, ok := strings.CutPrefix(a, "cat:"); ok {
		return strings.TrimSpace(c) != "" && utf8.RuneCountInString(c) <= 80 && !strings.ContainsAny(c, "<>\"")
	}
	if p, ok := strings.CutPrefix(a, "prod:"); ok {
		return reProducto.MatchString(p)
	}
	return strings.HasPrefix(a, "https://") && foto(a)
}

// categoriaValida: el nombre de una categoría en un bloque (hasta 60, como la web).
func categoriaValida(c string) bool {
	return strings.TrimSpace(c) != "" && utf8.RuneCountInString(c) <= 60 && !strings.ContainsAny(c, "<>\"")
}

// fotoBloque: las fotos de los bloques, sólo https (la web descarta las otras).
func fotoBloque(s string) bool { return strings.HasPrefix(s, "https://") && foto(s) }

func videoValido(u string) bool {
	if !strings.HasPrefix(u, "https://") || !foto(u) {
		return false
	}
	p, err := url.Parse(u)
	return err == nil && len(p.Path) > 1 && slices.Contains(hostsVideo, strings.ToLower(p.Hostname()))
}

func fechaValida(s string) bool {
	_, err := time.Parse(time.RFC3339, s)
	return err == nil
}

// textoValido: hasta max letras, sin HTML; en una línea salvo los largos.
func textoValido(v any, max int, largo bool) bool {
	s, ok := v.(string)
	if !ok || utf8.RuneCountInString(s) > max || strings.ContainsAny(s, "<>") {
		return false
	}
	return largo || !strings.ContainsAny(s, "\n\r")
}

func entero(v any, min, max *int) bool {
	n, ok := v.(float64)
	if !ok || n != float64(int64(n)) {
		return false
	}
	return (min == nil || n >= float64(*min)) && (max == nil || n <= float64(*max))
}

func lista(v any) ([]any, bool) {
	l, ok := v.([]any)
	return l, ok
}

// vacio: lo que no cuenta como cargado (para «requerido»).
func vacio(v any) bool {
	switch x := v.(type) {
	case nil:
		return true
	case string:
		return strings.TrimSpace(x) == ""
	case []any:
		return len(x) == 0
	}
	return false
}

func maxO(c campo, def int) int {
	if c.Max != nil {
		return *c.Max
	}
	return def
}

// valorValido valida un valor (ya decodificado de JSON) según el tipo de su campo (tiposCampo).
func valorValido(c campo, v any) bool {
	items := Catalogo.Limites.ItemsPorLista
	switch c.Tipo {
	case "texto":
		return textoValido(v, maxO(c, 200), false)
	case "textoLargo":
		return textoValido(v, maxO(c, 1000), true)
	case "foto":
		s, ok := v.(string)
		return ok && fotoBloque(s)
	case "fotos":
		l, ok := lista(v)
		if !ok || len(l) > maxO(c, items) {
			return false
		}
		for _, x := range l {
			if s, ok := x.(string); !ok || !fotoBloque(s) {
				return false
			}
		}
		return true
	case "link":
		s, ok := v.(string)
		return ok && linkValido(s)
	case "boton":
		m, ok := v.(map[string]any)
		if !ok {
			return false
		}
		for k := range m {
			if k != "t" && k != "a" {
				return false
			}
		}
		t, _ := m["t"].(string)
		a, _ := m["a"].(string)
		return ItemMenu{T: t, A: a}.valido()
	case "fondo":
		s, ok := v.(string)
		return ok && slices.Contains(Catalogo.Fondos.Opciones, s)
	case "fecha":
		s, ok := v.(string)
		return ok && fechaValida(s)
	case "numero":
		return entero(v, c.Min, c.Max)
	case "si":
		_, ok := v.(bool)
		return ok
	case "opcion":
		s, ok := v.(string)
		return ok && slices.Contains(c.Opciones, s)
	case "icono":
		s, ok := v.(string)
		return ok && slices.Contains(Catalogo.Iconos, s)
	case "video":
		s, ok := v.(string)
		return ok && videoValido(s)
	case "producto":
		s, ok := v.(string)
		return ok && reProducto.MatchString(s)
	case "elegidos":
		l, ok := lista(v)
		if !ok || len(l) > maxO(c, items) || (c.Min != nil && len(l) < *c.Min) {
			return false
		}
		for _, x := range l {
			if s, ok := x.(string); !ok || !reProducto.MatchString(s) {
				return false
			}
		}
		return true
	case "productos":
		return productosValidos(v)
	case "categoria":
		s, ok := v.(string)
		return ok && categoriaValida(s)
	case "categorias":
		l, ok := lista(v)
		if !ok || len(l) > maxO(c, items) {
			return false
		}
		for _, x := range l {
			if s, ok := x.(string); !ok || !categoriaValida(s) {
				return false
			}
		}
		return true
	case "lista":
		l, ok := lista(v)
		if !ok || len(l) > maxO(c, items) || (c.Min != nil && len(l) < *c.Min) {
			return false
		}
		for _, x := range l {
			m, ok := x.(map[string]any)
			if !ok || !camposValidos(c.Item, m, true) {
				return false
			}
		}
		return true
	case "grupo":
		m, ok := v.(map[string]any)
		return ok && camposValidos(c.Item, m, false)
	case "tabla":
		return tablaValida(v)
	case "cupon":
		s, ok := v.(string)
		return ok && (s == "" || reCupon.MatchString(s))
	case "premio":
		return premioValido(c.Opciones, v)
	case "opciones":
		return opcionesValidas(c.Opciones, v)
	}
	return false
}

// tablaValida: filas × columnas de textos cortos (la primera fila es el encabezado), de 2 a
// limites.filasTabla filas y de 1 a limites.columnasTabla columnas, todas del mismo ancho; cada celda
// hasta limites.letrasCelda letras en una línea, sin HTML (puede quedar vacía).
func tablaValida(v any) bool {
	l := Catalogo.Limites
	filas, ok := lista(v)
	if !ok || len(filas) < 2 || len(filas) > l.FilasTabla {
		return false
	}
	ancho := -1
	for _, f := range filas {
		celdas, ok := lista(f)
		if !ok || len(celdas) < 1 || len(celdas) > l.ColumnasTabla || (ancho >= 0 && len(celdas) != ancho) {
			return false
		}
		ancho = len(celdas)
		for _, c := range celdas {
			if !textoValido(c, l.LetrasCelda, false) {
				return false
			}
		}
	}
	return true
}

// productosValidos: { fuente, cantidad 1–24, ids? (elegidos), categoria? }; fuente y cantidad siempre.
func productosValidos(v any) bool {
	m, ok := v.(map[string]any)
	if !ok {
		return false
	}
	uno, veinticuatro := 1, 24
	for k, x := range m {
		switch k {
		case "fuente":
			s, ok := x.(string)
			if !ok || !slices.Contains(Catalogo.FuentesProductos.Opciones, s) {
				return false
			}
		case "cantidad":
			if !entero(x, &uno, &veinticuatro) {
				return false
			}
		case "ids":
			if !valorValido(campo{Tipo: "elegidos", Max: &veinticuatro}, x) {
				return false
			}
		case "categoria":
			if s, ok := x.(string); !ok || (s != "" && !categoriaValida(s)) {
				return false
			}
		default:
			return false
		}
	}
	_, conFuente := m["fuente"]
	_, conCantidad := m["cantidad"]
	return conFuente && conCantidad
}

// camposValidos: sólo los campos del tipo, cada uno con su valor. Los requeridos se exigen dentro
// de las listas (un ítem sin su dato no sirve); arriba, en el bloque, lo que falta toma lo de por
// defecto o lo sugerido, y si no alcanza el bloque no se dibuja (así el borrador se guarda a medias).
func camposValidos(campos []campo, datos map[string]any, exigir bool) bool {
	for k, v := range datos {
		i := slices.IndexFunc(campos, func(c campo) bool { return c.ID == k })
		if i < 0 {
			return false
		}
		if v == nil {
			continue
		}
		if !valorValido(campos[i], v) {
			return false
		}
	}
	if exigir {
		for _, c := range campos {
			if c.Requerido && vacio(datos[c.ID]) {
				return false
			}
		}
	}
	return true
}

// normalizar pasa los datos por JSON, para validar lo mismo que llega por la API y lo que se arma
// en Go (demo, migración).
func normalizar(m map[string]any) (map[string]any, bool) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, false
	}
	var out map[string]any
	if json.Unmarshal(b, &out) != nil {
		return nil, false
	}
	return out, true
}

// bloquesValidos: tipo y variante del catálogo, campos según su tipo, el máximo por tipo (portada:
// 1), ids únicos en todo el diseño y el tope de bloques por página.
func bloquesValidos(l []Bloque, vistos map[string]bool) bool {
	if len(l) > Catalogo.Limites.BloquesPorPagina {
		return false
	}
	porTipo := map[string]int{}
	for _, b := range l {
		t := tipoDe(b.Tipo)
		if t == nil || !slices.Contains(t.Variantes, b.Variante) || !reIDBloque.MatchString(b.ID) || vistos[b.ID] {
			return false
		}
		vistos[b.ID] = true
		porTipo[b.Tipo]++
		if t.Max > 0 && porTipo[b.Tipo] > t.Max {
			return false
		}
		datos, ok := normalizar(b.Datos)
		if !ok || !camposValidos(t.Campos, datos, false) {
			return false
		}
	}
	return true
}

func opcionDe(pagina, id string) campo {
	for _, c := range Catalogo.Paginas[pagina].Campos {
		if c.ID == id {
			return c
		}
	}
	return campo{}
}

func varianteValida(pagina, v string) bool {
	return v == "" || slices.Contains(Catalogo.Paginas[pagina].Variantes, v)
}

// valido: las páginas con sus variantes y opciones, y sus bloques.
func (p *Paginas) valido(vistos map[string]bool) bool {
	if p == nil {
		return true
	}
	if c := p.Catalogo; c != nil {
		if !varianteValida("catalogo", c.Variante) ||
			(c.Filtros != "" && !slices.Contains(opcionDe("catalogo", "filtros").Opciones, c.Filtros)) {
			return false
		}
		col := opcionDe("catalogo", "columnas")
		if c.Columnas != 0 && !entero(float64(c.Columnas), col.Min, col.Max) {
			return false
		}
	}
	if c := p.Producto; c != nil {
		if !varianteValida("producto", c.Variante) || !bloquesValidos(c.Bloques, vistos) {
			return false
		}
		if c.GuiaTalles != nil {
			g, ok := normalizar(c.GuiaTalles)
			if !ok || !valorValido(opcionDe("producto", "guiaTalles"), g) {
				return false
			}
		}
	}
	if c := p.Carrito; c != nil && !varianteValida("carrito", c.Variante) {
		return false
	}
	if c := p.Nosotros; c != nil && !bloquesValidos(c.Bloques, vistos) {
		return false
	}
	return true
}

// valido: la barra de anuncios (bloques.json › barraAnuncios).
func (b *Barra) valido() bool {
	if b == nil {
		return true
	}
	if b.Variante != "" && !slices.Contains(Catalogo.BarraAnuncios.Variantes, b.Variante) {
		return false
	}
	if b.Fondo != "" && !slices.Contains(Catalogo.Fondos.Opciones, b.Fondo) {
		return false
	}
	var mensajes campo
	for _, c := range Catalogo.BarraAnuncios.Campos {
		if c.ID == "mensajes" {
			mensajes = c
		}
	}
	if len(b.Mensajes) > maxO(mensajes, 5) {
		return false
	}
	for _, m := range b.Mensajes {
		if strings.TrimSpace(m.Texto) == "" || !textoValido(m.Texto, 100, false) || (m.Link != "" && !linkValido(m.Link)) ||
			(m.Hasta != "" && !fechaValida(m.Hasta)) {
			return false
		}
	}
	return true
}
