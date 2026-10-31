package vidriera

import "github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/mayorista"

// mayoristaDe la línea: el de la variante si tiene, si no el del producto (nil: no tiene).
func mayoristaDe(p ProductoPublico, varianteID string) *int64 {
	if len(p.Opciones) == 0 {
		return p.Mayorista
	}
	for _, v := range p.Variantes {
		if v.ID == varianteID {
			return mayorista.DeVariante(v.Mayorista, p.Mayorista)
		}
	}
	return nil
}

// preciosCarrito: el precio de cada línea como se cobra. Con el carrito al mínimo mayorista (las
// unidades y el monto a precio minorista), las que tienen precio mayorista pasan a ese (§49).
// ok en false: una línea no es válida (producto, variante o cantidad).
func preciosCarrito(t Tienda, porID map[string]ProductoPublico, carrito []LineaCarrito) (precios []int64, esMayorista, ok bool) {
	precios = make([]int64, len(carrito))
	var monto int64
	var unidades int
	for i, l := range carrito {
		p, existe := porID[l.ID]
		if !existe || l.Cantidad < 1 || l.Cantidad > 99 {
			return nil, false, false
		}
		precio, vale := precioDe(p, l.VarianteID)
		if !vale {
			return nil, false, false
		}
		precios[i] = precio
		monto += precio * int64(l.Cantidad)
		unidades += l.Cantidad
	}
	if t.Mayorista == nil {
		return precios, false, true
	}
	c := mayorista.Config{Activa: true, Unidades: t.Mayorista.Unidades, Monto: t.Mayorista.Monto}
	if !c.Alcanza(unidades, monto) {
		return precios, false, true
	}
	for i, l := range carrito {
		p := porID[l.ID]
		if m := mayorista.Precio(precios[i], mayoristaDe(p, l.VarianteID)); m < precios[i] {
			precios[i], esMayorista = m, true
		}
	}
	return precios, esMayorista, true
}
