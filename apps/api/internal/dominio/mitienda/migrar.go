package mitienda

import (
	"maps"
	"slices"
)

// De sección v1 a tipo de bloque v2.
var seccionABloque = map[string]string{"hero": "portada", "cats": "categorias", "prods": "productos", "revs": "resenas", "ig": "galeria", "wa": "whatsapp"}

// ordenV1: el orden de las secciones como venían en las plantillas.
var ordenV1 = []string{"hero", "cats", "prods", "revs", "ig", "wa"}

// Migrar pasa un diseño v1 (o ninguno) a v2; uno ya v2 vuelve igual. Es pura y hace lo mismo que
// lib/tienda/sistema/migrar.ts (los dos corren packages/contrato/casos/diseno-v2.json):
//   - el preset es el de la plantilla del diseño (si el borrador la cambia) o el de la tienda;
//   - un bloque por tipo del orden del preset (id = tipo), en ese orden, salvo que la tienda haya
//     reordenado las secciones: entonces el suyo, filtrado a los del preset (los que faltan, al final);
//   - las secciones ocultas quedan oculto:true; hero → portada con titulo, lema y portada;
//     textos[sección] → datos.titulo; wa oculta → estilo.whatsappFlotante=false;
//   - anuncio y anuncioLink → estilo.barra fija con un mensaje;
//   - se sacan secciones, textos, anuncio y anuncioLink.
func Migrar(d *Diseno, plantilla string) Diseno {
	if d == nil {
		return Inicial(plantilla)
	}
	if d.Version == 2 {
		return *d
	}
	if d.Plantilla != "" {
		plantilla = d.Plantilla
	}
	p := presetDe(plantilla)
	out := *d
	out.Version, out.Secciones = 2, nil

	// Las secciones de la tienda, en su orden.
	var suyas []string
	oculta := map[string]bool{}
	for _, s := range d.Secciones {
		if _, ok := seccionABloque[s.ID]; ok && !slices.Contains(suyas, s.ID) {
			suyas = append(suyas, s.ID)
			oculta[s.ID] = !s.Visible
		}
	}
	orden := p.Orden
	if reordenada(suyas) {
		orden = nil
		for _, s := range suyas {
			if t := seccionABloque[s]; slices.Contains(p.Orden, t) {
				orden = append(orden, t)
			}
		}
		for _, t := range p.Orden {
			if !slices.Contains(orden, t) {
				orden = append(orden, t)
			}
		}
	}
	var textos map[string]string
	if d.Estilo != nil {
		textos = d.Estilo.Textos
	}
	out.Inicio = []Bloque{}
	for _, t := range orden {
		sec := seccionDe(t)
		b := Bloque{ID: t, Tipo: t, Variante: p.Variantes[t], Oculto: oculta[sec], Datos: map[string]any{}}
		if t == "portada" {
			if d.Titulo != "" {
				b.Datos["titulo"] = d.Titulo
			}
			if d.Lema != "" {
				b.Datos["texto"] = d.Lema
			}
			if d.Portada != "" {
				b.Datos["fotos"] = []any{d.Portada}
			}
		}
		if x := textos[sec]; x != "" {
			b.Datos["titulo"] = x
		}
		out.Inicio = append(out.Inicio, b)
	}

	// El estilo: lo mismo, sin lo de v1.
	var e Estilo
	if d.Estilo != nil {
		e = *d.Estilo
		e.Colores, e.Redes = maps.Clone(e.Colores), maps.Clone(e.Redes)
		e.Menu = slices.Clone(e.Menu)
	}
	if e.Anuncio != "" && e.Barra == nil {
		e.Barra = &Barra{Variante: "fija", Mensajes: []MensajeBarra{{Texto: e.Anuncio, Link: e.AnuncioLink}}}
	}
	e.Anuncio, e.AnuncioLink, e.Textos = "", "", nil
	if oculta["wa"] {
		no := false
		e.WhatsappFlotante = &no
	}
	out.Estilo = nil
	if !e.vacio() {
		out.Estilo = &e
	}
	return out
}

// reordenada: las secciones no siguen el orden en que venían en las plantillas.
func reordenada(suyas []string) bool {
	ultimo := -1
	for _, s := range suyas {
		i := slices.Index(ordenV1, s)
		if i < ultimo {
			return true
		}
		ultimo = i
	}
	return false
}

func seccionDe(tipo string) string {
	for s, t := range seccionABloque {
		if t == tipo {
			return s
		}
	}
	return ""
}
