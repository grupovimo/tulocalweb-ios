package promos

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

var reCPA = regexp.MustCompile(`^[A-Z][0-9]{4}([A-Z]{3})?$`)

// Las 24 jurisdicciones por la letra del CPA (la misma que usa MiCorreo como código de provincia).
var provinciaPorLetra = map[byte]string{
	'A': "Salta", 'B': "Buenos Aires", 'C': "CABA", 'D': "San Luis", 'E': "Entre Ríos", 'F': "La Rioja",
	'G': "Santiago del Estero", 'H': "Chaco", 'J': "San Juan", 'K': "Catamarca", 'L': "La Pampa", 'M': "Mendoza",
	'N': "Misiones", 'P': "Formosa", 'Q': "Neuquén", 'R': "Río Negro", 'S': "Santa Fe", 'T': "Tucumán",
	'U': "Chubut", 'V': "Tierra del Fuego", 'W': "Corrientes", 'X': "Córdoba", 'Y': "Jujuy", 'Z': "Santa Cruz",
}

// Rangos de códigos postales de 4 cifras que caen enteros en una provincia. Los que se reparten entre
// dos (hay muchos en el límite de Buenos Aires y Santa Fe) no están: mejor sin provincia que una
// equivocada.
var rangosCP = []struct {
	desde, hasta int
	provincia    string
}{
	{1000, 1499, "CABA"}, {1600, 1999, "Buenos Aires"}, {2000, 2199, "Santa Fe"}, {3000, 3099, "Santa Fe"},
	{3100, 3299, "Entre Ríos"}, {3300, 3399, "Misiones"}, {3400, 3499, "Corrientes"}, {3500, 3599, "Chaco"},
	{3600, 3699, "Formosa"}, {4000, 4199, "Tucumán"}, {4200, 4399, "Santiago del Estero"}, {4400, 4499, "Salta"},
	{4600, 4699, "Jujuy"}, {4700, 4799, "Catamarca"}, {5000, 5299, "Córdoba"}, {5300, 5399, "La Rioja"},
	{5400, 5499, "San Juan"}, {5500, 5699, "Mendoza"}, {5700, 5799, "San Luis"}, {6300, 6399, "La Pampa"},
	{7000, 7699, "Buenos Aires"}, {8000, 8199, "Buenos Aires"}, {8300, 8399, "Neuquén"}, {8400, 8599, "Río Negro"},
	{9000, 9299, "Chubut"}, {9400, 9409, "Santa Cruz"}, {9410, 9421, "Tierra del Fuego"},
}

// ProvinciaDeCP: la provincia de un código postal (CPA con letra o de 4 cifras); "" si no se sabe.
func ProvinciaDeCP(cp string) string {
	cp = strings.ToUpper(strings.TrimSpace(cp))
	if cp == "" {
		return ""
	}
	if c := cp[0]; c >= 'A' && c <= 'Z' {
		if !reCPA.MatchString(cp) {
			return ""
		}
		return provinciaPorLetra[c]
	}
	digitos := strings.TrimRightFunc(cp, func(r rune) bool { return !unicode.IsDigit(r) })
	n, err := strconv.Atoi(digitos)
	if err != nil || len(digitos) != 4 {
		return ""
	}
	for _, r := range rangosCP {
		if n >= r.desde && n <= r.hasta {
			return r.provincia
		}
	}
	return ""
}

// provinciaDe: la de la ficha del cliente, la que eligió para Correo Argentino o la de su código postal.
func provinciaDe(ficha, codigoCorreo, cp string) string {
	if f := strings.TrimSpace(ficha); f != "" && len(f) <= 40 {
		return f
	}
	if len(codigoCorreo) == 1 {
		if p, ok := provinciaPorLetra[strings.ToUpper(codigoCorreo)[0]]; ok {
			return p
		}
	}
	return ProvinciaDeCP(cp)
}
