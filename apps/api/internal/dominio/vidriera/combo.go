package vidriera

import (
	"math"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/mitienda"
)

// Combo: un bloque «combo» del diseño con descuento (%). Agregados juntos, sus productos quedan
// marcados en el carrito con el id del bloque.
type Combo struct {
	ID        string   `json:"id"`
	Productos []string `json:"productos"`
	Descuento int64    `json:"descuento"`
}

// LineaCombo: lo que mira la regla de cada línea, con el precio que se cobra de la unidad.
type LineaCombo struct {
	ID         string  `json:"id"`
	VarianteID *string `json:"varianteId"`
	Cantidad   int     `json:"cantidad"`
	Combo      *string `json:"combo"`
	Precio     int64   `json:"precio"`
}

// CombosDe: los combos del diseño publicado (portada, ficha y Nosotros) que descuentan algo. Igual
// que lib/tienda/combos.ts › combosDe.
func CombosDe(d mitienda.Diseno) []Combo {
	bloques := append([]mitienda.Bloque(nil), d.Inicio...)
	if d.Paginas != nil {
		if d.Paginas.Producto != nil {
			bloques = append(bloques, d.Paginas.Producto.Bloques...)
		}
		if d.Paginas.Nosotros != nil {
			bloques = append(bloques, d.Paginas.Nosotros.Bloques...)
		}
	}
	var out []Combo
	for _, b := range bloques {
		if b.Tipo != "combo" || b.Oculto {
			continue
		}
		var productos []string
		vistos := map[string]bool{}
		if l, ok := b.Datos["productos"].([]any); ok {
			for _, x := range l {
				if s, ok := x.(string); ok && !vistos[s] {
					vistos[s] = true
					productos = append(productos, s)
				}
			}
		}
		pct, _ := b.Datos["descuento"].(float64)
		if len(productos) < 2 || pct <= 0 || pct > 50 {
			continue
		}
		out = append(out, Combo{ID: b.ID, Productos: productos, Descuento: int64(math.Floor(pct))})
	}
	return out
}

// DescuentoCombos: lo que descuentan los combos completos del carrito. Las líneas marcadas con el
// combo (sin variante) cuentan sólo si están todos sus productos; los juegos son la menor cantidad
// entre ellas; el descuento es el % de la suma de un juego por los juegos, redondeado al peso
// (casos compartidos: packages/contrato/casos/combos.json).
func DescuentoCombos(combos []Combo, lineas []LineaCombo) int64 {
	var total int64
	for _, c := range combos {
		de := map[string]LineaCombo{}
		for _, l := range lineas {
			if l.Combo != nil && *l.Combo == c.ID && (l.VarianteID == nil || *l.VarianteID == "") {
				if _, ya := de[l.ID]; !ya {
					de[l.ID] = l
				}
			}
		}
		juegos := math.MaxInt
		var suma int64
		completo := true
		for _, id := range c.Productos {
			l, ok := de[id]
			if !ok {
				completo = false
				break
			}
			juegos = min(juegos, l.Cantidad)
			suma += l.Precio
		}
		if !completo || juegos < 1 {
			continue
		}
		total += (suma*int64(juegos)*c.Descuento + 50) / 100
	}
	return total
}
