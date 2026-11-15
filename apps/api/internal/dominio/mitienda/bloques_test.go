package mitienda

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/tiendas"
)

// Las copias embebidas son iguales a las de packages/contrato (go generate ./internal/dominio/mitienda).
func TestCopiasDelContrato(t *testing.T) {
	for _, c := range []struct {
		nombre string
		copia  []byte
	}{{"bloques.json", bloquesJSON}, {"presets.json", presetsJSON}} {
		orig, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "..", "packages", "contrato", c.nombre))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(orig, c.copia) {
			t.Errorf("%s: la copia de mitienda difiere de packages/contrato; corré go generate ./internal/dominio/mitienda", c.nombre)
		}
	}
}

// Cada plantilla de tiendas tiene su preset, y los presets usan tipos y variantes del catálogo.
func TestPresetsContraElCatalogo(t *testing.T) {
	for _, p := range tiendas.Plantillas {
		if _, ok := Presets[p]; !ok {
			t.Errorf("la plantilla %s no tiene preset", p)
		}
	}
	for id, p := range Presets {
		if !slices.Contains(Catalogo.Estilo["direccion"].Opciones, p.Direccion) {
			t.Errorf("%s: dirección %q", id, p.Direccion)
		}
		for tipo, v := range p.Variantes {
			if b := tipoDe(tipo); b == nil || !slices.Contains(b.Variantes, v) {
				t.Errorf("%s: %s/%s no está en el catálogo", id, tipo, v)
			}
		}
	}
}

func con(f func(*Diseno)) Diseno {
	d := Inicial("ceibo")
	f(&d)
	return d
}

func bloque(id, tipo, variante, datos string) Bloque {
	b := Bloque{ID: id, Tipo: tipo, Variante: variante, Datos: map[string]any{}}
	if datos != "" {
		if err := json.Unmarshal([]byte(datos), &b.Datos); err != nil {
			panic(err)
		}
	}
	return b
}

func TestDisenoValido(t *testing.T) {
	agregar := func(id, tipo, variante, datos string) Diseno {
		return con(func(d *Diseno) { d.Inicio = append(d.Inicio, bloque(id, tipo, variante, datos)) })
	}
	si := true
	casos := []struct {
		nombre string
		d      Diseno
		ok     bool
	}{
		{"inicial", Inicial("ceibo"), true},
		{"sin version", con(func(d *Diseno) { d.Version = 0 }), false},
		{"con secciones v1", con(func(d *Diseno) { d.Secciones = []Seccion{{"hero", true}} }), false},
		{"con anuncio v1", con(func(d *Diseno) { d.Estilo = &Estilo{Anuncio: "x"} }), false},
		{"sin bloques", con(func(d *Diseno) { d.Inicio = nil }), true},
		{"banner bien", agregar("b7k2m9q1", "banner", "doble",
			`{"titulo":"Hot Sale","texto":"30% off","foto":"https://tulocalweb.com.ar/f/a.webp","boton":{"t":"Ver","a":"cat:Remeras"},"fondo":"acento","segundo":{"titulo":"Otra","boton":{"t":"Ir","a":"prod:3f2a-11"}}}`), true},
		{"tipo inventado", agregar("abcdef", "carrusel3d", "x", ""), false},
		{"variante inventada", agregar("abcdef", "banner", "gigante", ""), false},
		{"id corto", agregar("abc", "banner", "banda", ""), false},
		{"id con mayúsculas", agregar("ABCDEFG", "banner", "banda", ""), false},
		{"id repetido", agregar("portada", "banner", "banda", ""), false},
		{"dos portadas", agregar("portada2", "portada", "foto", ""), false},
		{"campo desconocido", agregar("abcdef", "banner", "banda", `{"color":"rojo"}`), false},
		{"texto largo de más", agregar("abcdef", "banner", "banda", `{"titulo":"`+strings.Repeat("a", 81)+`"}`), false},
		{"texto con html", agregar("abcdef", "banner", "banda", `{"titulo":"<script>"}`), false},
		{"texto en dos líneas", agregar("abcdef", "banner", "banda", `{"titulo":"a\nb"}`), false},
		{"texto largo con saltos", agregar("abcdef", "textoImagen", "derecha", `{"titulo":"Historia","texto":"a\nb"}`), true},
		{"fondo inventado", agregar("abcdef", "banner", "banda", `{"fondo":"violeta"}`), false},
		{"botón a otro lado", agregar("abcdef", "banner", "banda", `{"boton":{"t":"x","a":"javascript:alert(1)"}}`), false},
		{"botón con campo de más", agregar("abcdef", "banner", "banda", `{"boton":{"t":"x","a":"inicio","c":"y"}}`), false},
		{"foto no web", agregar("abcdef", "banner", "banda", `{"foto":"data:image/png;base64,xx"}`), false},
		{"video de youtube", agregar("abcdef", "video", "ancho", `{"url":"https://www.youtube.com/watch?v=abc"}`), true},
		{"video de otro lado", agregar("abcdef", "video", "ancho", `{"url":"https://videos.com/abc"}`), false},
		{"video http", agregar("abcdef", "video", "ancho", `{"url":"http://vimeo.com/1"}`), false},
		{"fecha bien", agregar("abcdef", "oferta", "banda", `{"hasta":"2026-11-20T23:59:00-03:00"}`), true},
		{"fecha mal", agregar("abcdef", "oferta", "banda", `{"hasta":"el viernes"}`), false},
		{"productos bien", agregar("abcdef", "productos", "carrusel", `{"productos":{"fuente":"elegidos","cantidad":3,"ids":["a1","b2"]},"columnas":3,"verTodo":false}`), true},
		{"productos sin fuente", agregar("abcdef", "productos", "grilla", `{"productos":{"cantidad":3}}`), false},
		{"productos de más", agregar("abcdef", "productos", "grilla", `{"productos":{"fuente":"nuevos","cantidad":25}}`), false},
		{"fuente inventada", agregar("abcdef", "productos", "grilla", `{"productos":{"fuente":"caros"}}`), false},
		{"columnas de más", agregar("abcdef", "productos", "grilla", `{"columnas":6}`), false},
		{"columnas con coma", agregar("abcdef", "productos", "grilla", `{"columnas":3.5}`), false},
		{"si que no es booleano", agregar("abcdef", "productos", "grilla", `{"verTodo":"sí"}`), false},
		{"combo con uno", agregar("abcdef", "combo", "tarjeta", `{"productos":["a1"]}`), false},
		{"combo con dos", agregar("abcdef", "combo", "tarjeta", `{"productos":["a1","b2"],"descuento":10}`), true},
		{"combo con descuento de más", agregar("abcdef", "combo", "tarjeta", `{"productos":["a1","b2"],"descuento":60}`), false},
		{"beneficios bien", agregar("abcdef", "beneficios", "fila", `{"items":[{"icono":"envio","titulo":"Envío gratis"}]}`), true},
		{"beneficio sin título", agregar("abcdef", "beneficios", "fila", `{"items":[{"icono":"envio"}]}`), false},
		{"ícono inventado", agregar("abcdef", "beneficios", "fila", `{"items":[{"icono":"cohete","titulo":"x"}]}`), false},
		{"beneficios de más", agregar("abcdef", "beneficios", "fila", `{"items":[`+strings.Repeat(`{"icono":"envio","titulo":"x"},`, 4)+`{"icono":"envio","titulo":"x"}]}`), false},
		{"ítem con campo de más", agregar("abcdef", "preguntas", "acordeon", `{"items":[{"p":"¿?","r":"Sí","x":1}]}`), false},
		{"local con horarios", agregar("ellocal", "local", "mapa", `{"horarios":[{"dias":"Lunes a viernes","horas":"9 a 18"}]}`), true},
		{"galería con link a instagram", agregar("abcdef", "galeria", "fila", `{"link":"https://instagram.com/x"}`), true},
		{"menú con categorías", agregar("abcdef", "menu", "lista", `{"categorias":["Pizzas","Empanadas"]}`), true},
		{"estilo v2 bien", con(func(d *Diseno) {
			d.Estilo = &Estilo{Direccion: "pedido", Forma: "pildora", Densidad: "compacta", Ancho: "amplio", FotoProducto: "apaisada",
				Encabezado: "buscador", Pie: "columnas", WhatsappFlotante: &si,
				Barra: &Barra{Variante: "rotativa", Fondo: "acento", Mensajes: []MensajeBarra{{"Envío gratis", "catalogo"}, {"3 cuotas", ""}}}}
		}), true},
		{"dirección inventada", con(func(d *Diseno) { d.Estilo = &Estilo{Direccion: "galeria"} }), false},
		{"barra con html", con(func(d *Diseno) { d.Estilo = &Estilo{Barra: &Barra{Mensajes: []MensajeBarra{{"<b>x</b>", ""}}}} }), false},
		{"barra con seis", con(func(d *Diseno) {
			d.Estilo = &Estilo{Barra: &Barra{Mensajes: slices.Repeat([]MensajeBarra{{"x", ""}}, 6)}}
		}), false},
		{"barra a otro lado", con(func(d *Diseno) { d.Estilo = &Estilo{Barra: &Barra{Mensajes: []MensajeBarra{{"x", "ftp://x"}}}} }), false},
		{"páginas bien", con(func(d *Diseno) {
			d.Paginas = &Paginas{Catalogo: &PaginaCatalogo{"lista", "arriba", 3}, Carrito: &PaginaCarrito{"panel", &si},
				Producto: &PaginaProducto{Variante: "carrusel", Relacionados: &si,
					Bloques: []Bloque{bloque("faqficha", "preguntas", "acordeon", `{"items":[{"p":"¿Talles?","r":"Del S al XL"}]}`)}},
				Nosotros: &PaginaNosotros{Bloques: []Bloque{bloque("historia", "textoImagen", "izquierda", `{"titulo":"Nuestra historia"}`)}}}
		}), true},
		{"catálogo con variante inventada", con(func(d *Diseno) { d.Paginas = &Paginas{Catalogo: &PaginaCatalogo{Variante: "mosaico"}} }), false},
		{"catálogo con 6 columnas", con(func(d *Diseno) { d.Paginas = &Paginas{Catalogo: &PaginaCatalogo{Columnas: 6}} }), false},
		{"filtros inventados", con(func(d *Diseno) { d.Paginas = &Paginas{Catalogo: &PaginaCatalogo{Filtros: "abajo"}} }), false},
		{"id repetido entre páginas", con(func(d *Diseno) {
			d.Paginas = &Paginas{Nosotros: &PaginaNosotros{Bloques: []Bloque{bloque("portada", "banner", "banda", "")}}}
		}), false},
		{"demasiados bloques", con(func(d *Diseno) {
			for i := range 30 {
				d.Inicio = append(d.Inicio, bloque(fmt.Sprintf("banner%02d", i), "banner", "banda", ""))
			}
		}), false},
		{"título con html", con(func(d *Diseno) { d.Titulo = "<b>Hola</b>" }), false},
	}
	for _, c := range casos {
		if got := c.d.Valido(); got != c.ok {
			t.Errorf("%s: Valido() = %v, quería %v", c.nombre, got, c.ok)
		}
	}
}

func TestLinks(t *testing.T) {
	for l, ok := range map[string]bool{
		"inicio": true, "catalogo": true, "turnos": true, "cat:Remeras": true, "prod:9b1f0c2e-1d2a-4b5c-8d9e-0f1a2b3c4d5e": true,
		"https://instagram.com/x": true, "http://x.com": false, "cat:": false, "prod:": false, "prod:a b": false,
		"javascript:alert(1)": false, "ofertas": false, `cat:a"b`: false,
	} {
		if linkValido(l) != ok {
			t.Errorf("linkValido(%q) != %v", l, ok)
		}
	}
}
