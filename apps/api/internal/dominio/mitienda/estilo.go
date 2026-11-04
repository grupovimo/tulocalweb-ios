package mitienda

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// Estilo es components/schemas/EstiloTienda (lib/tienda/estilo/estilo.ts): colores por rol y por
// zona, tipografías, logos, anuncio, textos de las secciones y redes. Todo opcional: vacío es lo de
// la plantilla. Los colores y las fuentes terminan en el CSS de la tienda, por eso se validan acá
// además de en la web.
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
	for _, p := range paginasMenu {
		if i.A == p {
			return true
		}
	}
	if strings.HasPrefix(i.A, "cat:") {
		return len(i.A) > 4 && utf8.RuneCountInString(i.A) <= 84 && !strings.ContainsAny(i.A, "<>\"")
	}
	return strings.HasPrefix(i.A, "https://") && foto(i.A)
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
	return permitidoEn(e.Textos, textosIDs) && permitidoEn(e.Redes, redesIDs)
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
	if e.Colores == nil && e.Textos == nil && e.Redes == nil && len(e.Menu) == 0 && e.FuenteTitulos == "" && e.FuenteTexto == "" &&
		e.LogoOscuro == "" && e.Icono == "" && e.LogoAlto == 0 && e.Anuncio == "" && e.AnuncioLink == "" && e.LeyendaPie == "" {
		return nil
	}
	return e
}
