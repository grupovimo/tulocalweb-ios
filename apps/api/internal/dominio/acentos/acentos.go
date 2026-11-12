// Package acentos es el color de marca de cada tienda: el acento de su plantilla, o la alternativa que
// eligió en Mi tienda › Colores (paleta 1–3), o el que pisó a mano en Mi tienda › Estilo; y el color
// del texto que va encima. Es la tabla de apps/web/src/lib/tienda/acentos.ts portada tal cual: si
// cambia allá, cambia acá (los tests usan los mismos casos que acentos.spec.ts).
package acentos

import (
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	// Oscuros, para llevar texto claro encima.
	conTextoClaro = []string{"#B0441D", "#2F5D50", "#1F4E8C", "#7A3B69"}
	// Claros, para llevar texto oscuro encima (plantillas oscuras: Pampa, Malbec).
	conTextoOscuro = []string{"#7FE7C4", "#FFB36B", "#B9A4FF"}
	// Medios, para llevar texto oscuro encima sobre un fondo claro (Mostaza, Patio, Hora, Lente).
	medios = []string{"#E8875F", "#D8B23A", "#8FBF9F"}
)

type acento struct {
	original, sobre string
	familia         []string
}

// Plantillas con archivo propio y las del motor (Motor.svelte): su acento y el texto encima.
var plantillas = map[string]acento{
	"andes":        {"#2F6BFF", "#FFFFFF", conTextoClaro},
	"ceibo":        {"#9C4A2F", "#F4EFEA", conTextoClaro},
	"malbec":       {"#E8B86A", "#1E0E12", conTextoOscuro},
	"miga":         {"#D9452B", "#F4EADB", conTextoClaro},
	"miga-clasica": {"#B5651D", "#FBF1E6", conTextoClaro},
	"mostaza":      {"#C9821A", "#2A170C", medios},
	"pampa":        {"#E6FF4D", "#111111", conTextoOscuro},
	"quebracho":    {"#4F5B3A", "#FFFFFF", conTextoClaro},
	"rosedal":      {"#B4475B", "#FFFFFF", conTextoClaro},
	"tijera":       {"#5B3FD6", "#FFFFFF", conTextoClaro},
	"yerbal":       {"#2F5D50", "#FFFFFF", conTextoClaro},
	"patio":        {"#F08A24", "#2A1E10", medios},
	"taller":       {"#6B4A2F", "#FFFFFF", conTextoClaro},
	"hora":         {"#B08D57", "#161616", medios},
	"litoral":      {"#6B3A22", "#FFFFFF", conTextoClaro},
	"torno":        {"#A4623D", "#FFFFFF", conTextoClaro},
	"lente":        {"#E0A31B", "#1F1A12", medios},
}

func deLaPlantilla(p string) acento {
	if a, ok := plantillas[p]; ok {
		return a
	}
	return plantillas["litoral"]
}

// Opcion es una de las cuatro de Mi tienda › Colores.
type Opcion struct{ Color, Sobre string }

// Opciones: la de la plantilla y tres alternativas distintas de ella.
func Opciones(plantilla string) []Opcion {
	a := deLaPlantilla(plantilla)
	r := []Opcion{{a.original, a.sobre}}
	for _, c := range a.familia {
		if len(r) == 4 {
			break
		}
		if !strings.EqualFold(c, a.original) {
			r = append(r, Opcion{c, a.sobre})
		}
	}
	return r
}

// Elegido: el acento de la paleta 1–3, o "" si es el de la plantilla (paleta 0 o fuera de rango).
func Elegido(plantilla string, paleta int) string {
	o := Opciones(plantilla)
	if paleta <= 0 || paleta >= len(o) {
		return ""
	}
	return o[paleta].Color
}

var reColor = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

// DeLaTienda: el acento con el que se ve la tienda y el color del texto encima. Manda lo que se
// pisó en Mi tienda › Estilo (colores.acento / colores.sobreAcento), después la paleta y por último
// la plantilla, como en Plantilla.svelte. Si el estilo pisa el acento pero no el texto, el texto es
// el de la plantilla salvo que no llegue a AA: entonces, blanco o casi negro, el que más contraste.
func DeLaTienda(plantilla string, paleta int, estiloAcento, estiloSobre string) (string, string) {
	o := Opciones(plantilla)
	color, sobre := o[0].Color, o[0].Sobre
	if e := Elegido(plantilla, paleta); e != "" {
		color = e
	}
	if reColor.MatchString(estiloAcento) {
		color = strings.ToUpper(estiloAcento)
		if Contraste(color, sobre) < 4.5 {
			sobre = legible(color)
		}
	}
	if reColor.MatchString(estiloSobre) {
		sobre = strings.ToUpper(estiloSobre)
	}
	return color, sobre
}

// legible: blanco o #111111, el que más contraste tenga con el color.
func legible(color string) string {
	if Contraste(color, "#FFFFFF") >= Contraste(color, "#111111") {
		return "#FFFFFF"
	}
	return "#111111"
}

// luminancia relativa (WCAG 2).
func luminancia(hex string) float64 {
	var c [3]float64
	for i := range c {
		v, _ := strconv.ParseUint(hex[1+2*i:3+2*i], 16, 8)
		x := float64(v) / 255
		if x <= 0.03928 {
			c[i] = x / 12.92
		} else {
			c[i] = math.Pow((x+0.055)/1.055, 2.4)
		}
	}
	return 0.2126*c[0] + 0.7152*c[1] + 0.0722*c[2]
}

// Contraste entre dos colores #RRGGBB (1 a 21).
func Contraste(a, b string) float64 {
	l := []float64{luminancia(a), luminancia(b)}
	sort.Sort(sort.Reverse(sort.Float64Slice(l)))
	return (l[0] + 0.05) / (l[1] + 0.05)
}

// ConAcento: las plantillas que tienen acento propio (para los tests).
func ConAcento() []string {
	r := make([]string, 0, len(plantillas))
	for p := range plantillas {
		r = append(r, p)
	}
	sort.Strings(r)
	return r
}
