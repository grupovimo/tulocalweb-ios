package avisos

import (
	"strings"
	"testing"
)

func TestMailsDeSuscripcion(t *testing.T) {
	x := suscripcion{codigo: "abcdefghijklmnopqrstuvwx", plan: "Club Explorador", frecuencia: "mensual", nombre: "Julia Paz",
		email: "julia@mail.com", entrega: "envio", direccion: "Gorriti 4500", precio: 118000, tienda: "Vinoteca Tres", slug: "vinoteca-tres"}
	m := mailAltaComprador("https://tulocalweb.com.ar/", x)
	link := "https://tulocalweb.com.ar/t/vinoteca-tres/suscripcion/abcdefghijklmnopqrstuvwx"
	if !strings.Contains(m.Texto, link) || !strings.Contains(m.HTML, link) || !strings.Contains(m.Texto, "$ 118.000 por mes") ||
		!strings.Contains(m.Texto, "Gorriti 4500") || !strings.Contains(m.Asunto, "Vinoteca Tres") {
		t.Fatalf("alta:\n%s", m.Texto)
	}
	// Sin renglón viudo: las dos últimas palabras de cada párrafo van juntas.
	if !strings.Contains(m.HTML, "cuando&nbsp;quieras") && !strings.Contains(m.HTML, "sólo&nbsp;tuyo.") {
		t.Fatalf("viuda:\n%s", m.HTML)
	}
	x.entrega = "retiro"
	if m := mailAltaComprador("https://tulocalweb.com.ar", x); !strings.Contains(m.Texto, "Lo retirás en el local") {
		t.Fatalf("retiro:\n%s", m.Texto)
	}
	m = mailTiendaSuscripcion("https://tulocalweb.com.ar", x, "Nuevo suscriptor: Club Explorador", "Julia se suscribió.")
	if !strings.Contains(m.Texto, "https://tulocalweb.com.ar/panel/suscripciones") || !strings.Contains(m.Texto, "julia@mail.com") {
		t.Fatalf("tienda:\n%s", m.Texto)
	}
}
