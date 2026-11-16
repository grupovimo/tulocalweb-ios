package mitienda

import (
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"
)

// Estilo es components/schemas/EstiloTienda (lib/tienda/estilo/estilo.ts): colores por rol y por
// zona, tipografías, logos, redes, menú y, desde el diseño v2 (docs/especificaciones/tienda-bloques.md),
// dirección, forma, densidad, ancho, fotos, encabezado, pie, barra de anuncios y botón de WhatsApp.
// Todo opcional: vacío es lo del preset. Los colores y las fuentes terminan en el CSS de la tienda,
// por eso se validan acá además de en la web. Anuncio, AnuncioLink y Textos son de v1: se leen para
// migrar (Migrar) y un diseño v2 no los trae.
type Estilo struct {
	Colores       map[string]string `json:"colores,omitempty"`
	FuenteTitulos string            `json:"fuenteTitulos,omitempty"`
	FuenteTexto   string            `json:"fuenteTexto,omitempty"`
	LogoOscuro    string            `json:"logoOscuro,omitempty"`
	Icono         string            `json:"icono,omitempty"`
	LogoAlto      int               `json:"logoAlto,omitempty"`
	Anuncio       string            `json:"anuncio,omitempty"`
	AnuncioLink   string            `json:"anuncioLink,omitempty"`
	LeyendaPie    string            `json:"leyendaPie,omitempty"`
	Textos        map[string]string `json:"textos,omitempty"`
	Redes         map[string]string `json:"redes,omitempty"`
	// Menu: las pestañas del menú (lib/tienda/estilo/menu.ts); vacío, las de la plantilla.
	Menu []ItemMenu `json:"menu,omitempty"`

	// v2: opciones de bloques.json › estilo (vacío: lo de la dirección y el preset).
	Direccion    string `json:"direccion,omitempty"`
	Forma        string `json:"forma,omitempty"`
	Densidad     string `json:"densidad,omitempty"`
	Ancho        string `json:"ancho,omitempty"`
	FotoProducto string `json:"fotoProducto,omitempty"`
	Encabezado   string `json:"encabezado,omitempty"`
	Pie          string `json:"pie,omitempty"`
	// Barra: la de anuncios, arriba de todo (bloques.json › barraAnuncios).
	Barra *Barra `json:"barra,omitempty"`
	// WhatsappFlotante: false saca el botón flotante de WhatsApp (nil: se ve, si hay número).
	WhatsappFlotante *bool `json:"whatsappFlotante,omitempty"`
	// Conversion: las herramientas para vender más (bloques.json › conversion); lo que falta, porDefecto.
	Conversion map[string]any `json:"conversion,omitempty"`
}

// Barra de anuncios: fija (un mensaje) o rotativa.
type Barra struct {
	Variante string         `json:"variante,omitempty"`
	Mensajes []MensajeBarra `json:"mensajes"`
	Fondo    string         `json:"fondo,omitempty"`
	// Automaticos: false para no sumar las cuotas, el envío gratis y la transferencia de la
	// configuración (nil o true: se suman). Con false la barra puede quedar sin mensajes.
	Automaticos *bool `json:"automaticos,omitempty"`
}

type MensajeBarra struct {
	Texto string `json:"texto"`
	Link  string `json:"link,omitempty"`
	// Hasta: cuenta regresiva; al vencer el mensaje se esconde (ISO 8601).
	Hasta string `json:"hasta,omitempty"`
}

// ItemMenu: el texto de la pestaña y adónde lleva (una página, «cat:<categoría>» o un link https).
type ItemMenu struct {
	T string `json:"t"`
	A string `json:"a"`
}

var paginasMenu = []string{"inicio", "catalogo", "nosotros", "turnos", "suscripciones"}

func (i ItemMenu) valido() bool {
	t := strings.TrimSpace(i.T)
	if t == "" || utf8.RuneCountInString(t) > 30 || strings.ContainsAny(t, "<>") {
		return false
	}
	return linkValido(i.A)
}

var (
	rolesColor  = []string{"fondo", "superficie", "texto", "textoSuave", "borde", "acento", "sobreAcento", "encabezado", "encabezadoTexto", "pie", "pieTexto"}
	textosIDs   = map[string]int{"cats": 60, "prods": 60, "revs": 60, "ig": 60, "wa": 60}
	redesIDs    = map[string]int{"instagram": 60, "tiktok": 60, "facebook": 100, "youtube": 100, "email": 120, "telefono": 30}
	reColor     = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)
	reFuente    = regexp.MustCompile(`^[a-z0-9-]{1,40}$`)
	permitidoEn = func(m map[string]string, ids map[string]int) bool {
		for k, v := range m {
			max, ok := ids[k]
			if !ok || utf8.RuneCountInString(v) > max || strings.ContainsAny(v, "<>") {
				return false
			}
		}
		return true
	}
)

func (e *Estilo) valido() bool {
	if e == nil {
		return true
	}
	for k, v := range e.Colores {
		ok := false
		for _, r := range rolesColor {
			ok = ok || r == k
		}
		if !ok || (v != "" && !reColor.MatchString(v)) {
			return false
		}
	}
	for _, f := range []string{e.FuenteTitulos, e.FuenteTexto} {
		if f != "" && !reFuente.MatchString(f) {
			return false
		}
	}
	if !foto(e.LogoOscuro) || !foto(e.Icono) || !foto(e.AnuncioLink) || e.LogoAlto < 0 || e.LogoAlto > 72 {
		return false
	}
	if utf8.RuneCountInString(e.Anuncio) > 120 || utf8.RuneCountInString(e.LeyendaPie) > 160 {
		return false
	}
	if len(e.Menu) > 8 {
		return false
	}
	for _, i := range e.Menu {
		if !i.valido() {
			return false
		}
	}
	for k, v := range map[string]string{"direccion": e.Direccion, "forma": e.Forma, "densidad": e.Densidad, "ancho": e.Ancho,
		"fotoProducto": e.FotoProducto, "encabezado": e.Encabezado, "pie": e.Pie} {
		if v != "" && !slices.Contains(Catalogo.Estilo[k].Opciones, v) {
			return false
		}
	}
	if !e.Barra.valido() || !conversionValida(e.Conversion) {
		return false
	}
	return permitidoEn(e.Textos, textosIDs) && permitidoEn(e.Redes, redesIDs)
}

// v1 dice si trae campos del diseño viejo (que v2 reemplaza por la barra y los títulos de los bloques).
func (e *Estilo) v1() bool {
	return e != nil && (e.Anuncio != "" || e.AnuncioLink != "" || len(e.Textos) > 0)
}

// vacio: nada distinto de lo del preset.
func (e *Estilo) vacio() bool {
	return e.Colores == nil && e.Textos == nil && e.Redes == nil && len(e.Menu) == 0 && e.FuenteTitulos == "" && e.FuenteTexto == "" &&
		e.LogoOscuro == "" && e.Icono == "" && e.LogoAlto == 0 && e.Anuncio == "" && e.AnuncioLink == "" && e.LeyendaPie == "" &&
		e.Direccion == "" && e.Forma == "" && e.Densidad == "" && e.Ancho == "" && e.FotoProducto == "" && e.Encabezado == "" &&
		e.Pie == "" && e.Barra == nil && e.WhatsappFlotante == nil && len(e.Conversion) == 0
}

// limpiar saca los espacios de más y lo vacío, para que el borrador no cambie por nada.
func (e *Estilo) limpiar() *Estilo {
	if e == nil {
		return nil
	}
	poda := func(m map[string]string) map[string]string {
		for k, v := range m {
			if v = strings.TrimSpace(v); v == "" {
				delete(m, k)
			} else {
				m[k] = v
			}
		}
		if len(m) == 0 {
			return nil
		}
		return m
	}
	e.Colores, e.Textos, e.Redes = poda(e.Colores), poda(e.Textos), poda(e.Redes)
	for k, v := range e.Colores {
		e.Colores[k] = strings.ToUpper(v)
	}
	e.Anuncio, e.LeyendaPie = strings.TrimSpace(e.Anuncio), strings.TrimSpace(e.LeyendaPie)
	for k := range e.Menu {
		e.Menu[k].T = strings.TrimSpace(e.Menu[k].T)
	}
	if e.Barra != nil {
		for k := range e.Barra.Mensajes {
			e.Barra.Mensajes[k].Texto = strings.TrimSpace(e.Barra.Mensajes[k].Texto)
		}
		if len(e.Barra.Mensajes) == 0 {
			e.Barra = nil
		}
	}
	if len(e.Conversion) == 0 {
		e.Conversion = nil
	}
	if e.vacio() {
		return nil
	}
	return e
}
