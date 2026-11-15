package mitienda

import (
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"
)

// Ventana: un popup de la tienda (docs/especificaciones/tienda-promociones.md, bloques.json › ventanas).
// Datos son los campos de su contenido. Una ventana válida puede estar incompleta (el borrador se guarda
// a medias): sale en la tienda sólo si Completa.
type Ventana struct {
	ID           string         `json:"id"`
	Nombre       string         `json:"nombre"`
	Activa       bool           `json:"activa"`
	Variante     string         `json:"variante"`
	Contenido    string         `json:"contenido"`
	Datos        map[string]any `json:"datos"`
	Disparo      Disparo        `json:"disparo"`
	Frecuencia   Frecuencia     `json:"frecuencia"`
	Publico      string         `json:"publico,omitempty"`
	Paginas      []string       `json:"paginas"`
	Dispositivos string         `json:"dispositivos,omitempty"`
	Desde        string         `json:"desde,omitempty"`
	Hasta        string         `json:"hasta,omitempty"`
}

type Disparo struct {
	Cuando string `json:"cuando"`
	Valor  *int   `json:"valor,omitempty"`
}

type Frecuencia struct {
	Cada string `json:"cada"`
	Dias *int   `json:"dias,omitempty"`
}

type rango struct {
	Min        int `json:"min"`
	Max        int `json:"max"`
	PorDefecto int `json:"porDefecto"`
}

type contenidoVentana struct {
	Tipo   string  `json:"tipo"`
	Campos []campo `json:"campos"`
}

type catalogoVentanas struct {
	Variantes  []string           `json:"variantes"`
	Campos     []campo            `json:"campos"`
	Contenidos []contenidoVentana `json:"contenidos"`
	Disparos   []struct {
		Cuando string `json:"cuando"`
		Valor  *rango `json:"valor"`
	} `json:"disparos"`
	Frecuencias []struct {
		Cada string `json:"cada"`
		Dias *rango `json:"dias"`
	} `json:"frecuencias"`
	Plantillas []struct {
		ID          string          `json:"id"`
		HastaEnDias int             `json:"hastaEnDias"`
		Ventana     json.RawMessage `json:"ventana"`
	} `json:"plantillas"`
	Publicos     opciones `json:"publicos"`
	Paginas      opciones `json:"paginas"`
	Dispositivos opciones `json:"dispositivos"`
}

var (
	reCupon       = regexp.MustCompile(`^[A-Z0-9-]{3,20}$`)
	tiposPremio   = []string{"cuponFijo", "cuponUnico", "descarga", "nada"}
	descuentos    = []string{"porcentaje", "monto", "envioGratis"}
	clavesPremio  = []string{"tipo", "cupon", "descuento", "valor", "dias", "archivo", "nombreArchivo"}
	uno, noventa  = 1, 90
	cien, cienMil = 100, 100000
	sesenta       = 60
)

func contenidoDe(tipo string) *contenidoVentana {
	for i := range Catalogo.Ventanas.Contenidos {
		if Catalogo.Ventanas.Contenidos[i].Tipo == tipo {
			return &Catalogo.Ventanas.Contenidos[i]
		}
	}
	return nil
}

// premioValido: { tipo, … } con las claves de su tipo bien (las de otro tipo se ignoran). Lo que falta
// no lo invalida: eso lo dice premioCompleto.
func premioValido(permitidos []string, v any) bool {
	m, ok := v.(map[string]any)
	if !ok {
		return false
	}
	if len(permitidos) == 0 {
		permitidos = tiposPremio
	}
	tipo, _ := m["tipo"].(string)
	if !slices.Contains(permitidos, tipo) {
		return false
	}
	for k := range m {
		if !slices.Contains(clavesPremio, k) {
			return false
		}
	}
	switch tipo {
	case "cuponFijo":
		if x, ok := m["cupon"]; ok {
			if s, ok := x.(string); !ok || (s != "" && !reCupon.MatchString(s)) {
				return false
			}
		}
	case "cuponUnico":
		d, conDescuento := m["descuento"]
		desc, _ := d.(string)
		if conDescuento && !slices.Contains(descuentos, desc) {
			return false
		}
		if x, ok := m["valor"]; ok {
			switch desc {
			case "porcentaje":
				if !entero(x, &uno, &noventa) {
					return false
				}
			case "monto":
				if !entero(x, &cien, &cienMil) {
					return false
				}
			case "envioGratis":
				return false
			default:
				if !entero(x, &uno, &cienMil) {
					return false
				}
			}
		}
		if x, ok := m["dias"]; ok && !entero(x, &uno, &noventa) {
			return false
		}
	case "descarga":
		if x, ok := m["archivo"]; ok {
			if s, ok := x.(string); !ok || (s != "" && !fotoBloque(s)) {
				return false
			}
		}
		if x, ok := m["nombreArchivo"]; ok && !textoValido(x, sesenta, false) {
			return false
		}
	}
	return true
}

// premioCompleto: el premio trae lo que necesita para entregarse.
func premioCompleto(v any) bool {
	m, ok := v.(map[string]any)
	if !ok {
		return false
	}
	switch m["tipo"] {
	case "cuponFijo":
		return !vacio(m["cupon"])
	case "cuponUnico":
		d, _ := m["descuento"].(string)
		return d == "envioGratis" || (slices.Contains(descuentos, d) && m["valor"] != nil)
	case "descarga":
		return !vacio(m["archivo"])
	case "nada":
		return true
	}
	return false
}

func opcionesValidas(permitidas []string, v any) bool {
	l, ok := lista(v)
	if !ok || len(l) > len(permitidas) {
		return false
	}
	vistas := map[string]bool{}
	for _, x := range l {
		s, ok := x.(string)
		if !ok || !slices.Contains(permitidas, s) || vistas[s] {
			return false
		}
		vistas[s] = true
	}
	return true
}

func paginasValidas(l []string) bool {
	a := make([]any, len(l))
	for i, s := range l {
		a[i] = s
	}
	return opcionesValidas(Catalogo.Ventanas.Paginas.Opciones, a)
}

func enRango(r *rango, v *int) bool {
	if v == nil {
		return true
	}
	return r != nil && *v >= r.Min && *v <= r.Max
}

// valida: la ventana según el catálogo (no dice si está completa).
func (v Ventana) valida() bool {
	c := Catalogo.Ventanas
	n := strings.TrimSpace(v.Nombre)
	if !reIDBloque.MatchString(v.ID) || n == "" || !textoValido(v.Nombre, maxO(c.Campos[0], 40), false) ||
		!slices.Contains(c.Variantes, v.Variante) {
		return false
	}
	cont := contenidoDe(v.Contenido)
	if cont == nil {
		return false
	}
	datos, ok := normalizar(v.Datos)
	if !ok || !camposValidos(cont.Campos, datos, false) {
		return false
	}
	i := slices.IndexFunc(c.Disparos, func(d struct {
		Cuando string `json:"cuando"`
		Valor  *rango `json:"valor"`
	}) bool {
		return d.Cuando == v.Disparo.Cuando
	})
	if i < 0 || !enRango(c.Disparos[i].Valor, v.Disparo.Valor) {
		return false
	}
	j := slices.IndexFunc(c.Frecuencias, func(f struct {
		Cada string `json:"cada"`
		Dias *rango `json:"dias"`
	}) bool {
		return f.Cada == v.Frecuencia.Cada
	})
	if j < 0 || !enRango(c.Frecuencias[j].Dias, v.Frecuencia.Dias) {
		return false
	}
	if (v.Publico != "" && !slices.Contains(c.Publicos.Opciones, v.Publico)) ||
		(v.Dispositivos != "" && !slices.Contains(c.Dispositivos.Opciones, v.Dispositivos)) || !paginasValidas(v.Paginas) {
		return false
	}
	var desde, hasta time.Time
	var err error
	if v.Desde != "" {
		if desde, err = time.Parse(time.RFC3339, v.Desde); err != nil {
			return false
		}
	}
	if v.Hasta != "" {
		if hasta, err = time.Parse(time.RFC3339, v.Hasta); err != nil {
			return false
		}
	}
	return v.Desde == "" || v.Hasta == "" || hasta.After(desde)
}

// Completa: además de válida, sale en la tienda: los requeridos del contenido cargados, el premio con lo
// suyo y, en la ruleta, porciones completas con algún peso mayor a 0.
func (v Ventana) Completa() bool {
	if !v.valida() {
		return false
	}
	cont := contenidoDe(v.Contenido)
	datos, _ := normalizar(v.Datos)
	for _, c := range cont.Campos {
		if c.Requerido && vacio(datos[c.ID]) {
			return false
		}
		if m, ok := datos[c.ID].(map[string]any); ok && c.Requerido && len(m) == 0 {
			return false
		}
		if c.Tipo == "premio" && !premioCompleto(datos[c.ID]) {
			return false
		}
	}
	if v.Contenido == "ruleta" {
		total := 0.0
		porciones, _ := lista(datos["porciones"])
		for _, x := range porciones {
			p, _ := x.(map[string]any)
			if !premioCompleto(p["premio"]) {
				return false
			}
			peso, _ := p["peso"].(float64)
			total += peso
		}
		if total <= 0 {
			return false
		}
	}
	return true
}

// Porciones de una ruleta (ya validada), para sortear.
func (v Ventana) Porciones() []map[string]any {
	datos, _ := normalizar(v.Datos)
	l, _ := lista(datos["porciones"])
	out := make([]map[string]any, 0, len(l))
	for _, x := range l {
		if m, ok := x.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}

// ventanasValidas: hasta limites.ventanas, ids únicos y cada una válida.
func ventanasValidas(l []Ventana) bool {
	if len(l) > Catalogo.Limites.Ventanas {
		return false
	}
	vistos := map[string]bool{}
	for _, v := range l {
		if vistos[v.ID] || !v.valida() {
			return false
		}
		vistos[v.ID] = true
	}
	return true
}

// conversionValida: estilo.conversion, un grupo de los campos de bloques.json › conversion.
func conversionValida(m map[string]any) bool {
	if m == nil {
		return true
	}
	d, ok := normalizar(m)
	return ok && camposValidos(Catalogo.Conversion.Campos, d, false)
}

// DePlantilla: una plantilla del catálogo (bloques.json › ventanas.plantillas) como ventana, con su id,
// el monto de envío gratis en lugar de {envioGratis} y, si la plantilla lo pide, la cuenta regresiva
// desde ahora (lo mismo que hace el editor al crear; las demos la usan así).
func DePlantilla(plantilla, id string, gratisDesde int64, ahora time.Time) (Ventana, bool) {
	for _, p := range Catalogo.Ventanas.Plantillas {
		if p.ID != plantilla {
			continue
		}
		var v Ventana
		if json.Unmarshal(p.Ventana, &v) != nil {
			return v, false
		}
		v.ID = id
		monto := ""
		if gratisDesde > 0 {
			monto = "$ " + miles(gratisDesde)
		}
		for _, k := range []string{"titulo", "texto", "etiqueta"} {
			if s, ok := v.Datos[k].(string); ok {
				v.Datos[k] = strings.ReplaceAll(s, "{envioGratis}", monto)
			}
		}
		if p.HastaEnDias > 0 {
			v.Datos["hasta"] = ahora.Add(time.Duration(p.HastaEnDias) * 24 * time.Hour).UTC().Format(time.RFC3339)
		}
		return v, v.valida()
	}
	return Ventana{}, false
}

func miles(n int64) string {
	s := fmt.Sprint(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "." + s[i:]
	}
	return s
}
