package mitienda

import (
	"maps"
	"slices"
)

// Mientras la web vieja (las 17 plantillas) conviva con la de bloques, la tienda pública manda
// también lo de v1 armado desde el diseño v2. Cuando todo migró, esto se borra.

// LeerDiseno: el jsonb de la base como diseño v2 (los v1 y los nulos se migran con la plantilla).
func LeerDiseno(crudo []byte, plantilla string) Diseno { return leerDiseno(crudo, plantilla) }

// SeccionesV1: las secciones visibles como las esperan las plantillas viejas, en el orden de los
// bloques del inicio («wa» al final, salvo que la tienda apagó el botón de WhatsApp).
func (d Diseno) SeccionesV1() []string {
	out := []string{}
	for _, b := range d.Inicio {
		// Una sección v1 por tipo: con novedades y ofertas hay tres bloques de productos, y v1 sólo
		// tiene una grilla de productos (repetida, la plantilla vieja la dibujaba tres veces).
		if s := seccionDe(b.Tipo); s != "" && s != "wa" && !b.Oculto && !slices.Contains(out, s) {
			out = append(out, s)
		}
	}
	if d.Estilo == nil || d.Estilo.WhatsappFlotante == nil || *d.Estilo.WhatsappFlotante {
		out = append(out, "wa")
	}
	return out
}

// EstiloV1: el estilo con el anuncio (el primer mensaje de la barra) y los títulos de las secciones
// (los de los bloques), como lo leen las plantillas viejas. nil si no hay nada.
func (d Diseno) EstiloV1() *Estilo {
	var e Estilo
	if d.Estilo != nil {
		e = *d.Estilo
	}
	if e.Barra != nil && len(e.Barra.Mensajes) > 0 {
		e.Anuncio, e.AnuncioLink = e.Barra.Mensajes[0].Texto, e.Barra.Mensajes[0].Link
	}
	// El título de cada sección sale sólo del primer bloque de su tipo (el que dibuja v1): si no, el de
	// «Recién llegados» (segundo bloque de productos) titulaba la grilla de todos los productos.
	textos := map[string]string{}
	vistas := map[string]bool{}
	for _, b := range d.Inicio {
		if s := seccionDe(b.Tipo); s != "" && s != "hero" && !vistas[s] {
			vistas[s] = true
			if t, ok := b.Datos["titulo"].(string); ok && t != "" {
				textos[s] = t
			}
		}
	}
	if len(textos) > 0 {
		e.Textos = maps.Clone(textos)
	}
	if e.vacio() {
		return nil
	}
	return &e
}

// ProductosReferidos: los ids de producto que nombran los bloques (destacado, combo, productos
// elegidos a mano y links prod:<id>), sin repetir, de todas las páginas.
func (d Diseno) ProductosReferidos() []string {
	vistos := map[string]bool{}
	var out []string
	sumar := func(v any) {
		if s, ok := v.(string); ok && reProducto.MatchString(s) && !vistos[s] {
			vistos[s] = true
			out = append(out, s)
		}
	}
	var recorrer func(v any)
	recorrer = func(v any) {
		switch x := v.(type) {
		case map[string]any:
			for k, y := range x {
				if k == "a" || k == "link" {
					if s, ok := y.(string); ok && len(s) > 5 && s[:5] == "prod:" {
						sumar(s[5:])
					}
				}
				recorrer(y)
			}
		case []any:
			for _, y := range x {
				recorrer(y)
			}
		}
	}
	bloques := append([]Bloque(nil), d.Inicio...)
	if d.Paginas != nil {
		if d.Paginas.Producto != nil {
			bloques = append(bloques, d.Paginas.Producto.Bloques...)
		}
		if d.Paginas.Nosotros != nil {
			bloques = append(bloques, d.Paginas.Nosotros.Bloques...)
		}
	}
	for _, b := range bloques {
		datos, ok := normalizar(b.Datos)
		if !ok {
			continue
		}
		switch b.Tipo {
		case "destacado":
			sumar(datos["producto"])
		case "combo":
			if l, ok := datos["productos"].([]any); ok {
				for _, x := range l {
					sumar(x)
				}
			}
		}
		if p, ok := datos["productos"].(map[string]any); ok {
			if l, ok := p["ids"].([]any); ok {
				for _, x := range l {
					sumar(x)
				}
			}
		}
		recorrer(datos)
	}
	if d.Estilo != nil && d.Estilo.Barra != nil {
		for _, m := range d.Estilo.Barra.Mensajes {
			if len(m.Link) > 5 && m.Link[:5] == "prod:" {
				sumar(m.Link[5:])
			}
		}
	}
	return out
}
