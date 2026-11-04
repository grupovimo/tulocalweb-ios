package mitienda

import "testing"

func TestEstiloValido(t *testing.T) {
	casos := []struct {
		nombre string
		e      *Estilo
		ok     bool
	}{
		{"sin estilo", nil, true},
		{"colores bien", &Estilo{Colores: map[string]string{"fondo": "#101010", "pieTexto": "#FFFFFF"}}, true},
		{"color vacío", &Estilo{Colores: map[string]string{"acento": ""}}, true},
		{"rol inventado", &Estilo{Colores: map[string]string{"marco": "#101010"}}, false},
		{"color con CSS", &Estilo{Colores: map[string]string{"fondo": "red;background:url(x)"}}, false},
		{"fuente rara", &Estilo{FuenteTitulos: "Inter';x"}, false},
		{"logo alto de más", &Estilo{LogoAlto: 200}, false},
		{"link que no es web", &Estilo{AnuncioLink: "javascript:alert(1)"}, false},
		{"título con html", &Estilo{Textos: map[string]string{"cats": "<b>x</b>"}}, false},
		{"red inventada", &Estilo{Redes: map[string]string{"myspace": "x"}}, false},
		{"menú bien", &Estilo{Menu: []ItemMenu{{"Nosotros", "nosotros"}, {"Remeras", "cat:Remeras"}, {"Blog", "https://blog.tienda.com"}}}, true},
		{"menú a otro lado", &Estilo{Menu: []ItemMenu{{"x", "javascript:alert(1)"}}}, false},
		{"menú sin texto", &Estilo{Menu: []ItemMenu{{" ", "inicio"}}}, false},
		{"menú largo", &Estilo{Menu: make([]ItemMenu, 9)}, false},
		{"redes bien", &Estilo{Redes: map[string]string{"instagram": "tienda.ok", "email": "hola@tienda.com"}}, true},
	}
	for _, c := range casos {
		if got := c.e.valido(); got != c.ok {
			t.Errorf("%s: valido() = %v, quería %v", c.nombre, got, c.ok)
		}
	}
}

func TestEstiloLimpiar(t *testing.T) {
	e := (&Estilo{Colores: map[string]string{"fondo": " #abcdef ", "texto": ""}, Anuncio: "  "}).limpiar()
	if e == nil || e.Colores["fondo"] != "#ABCDEF" || len(e.Colores) != 1 {
		t.Fatalf("limpiar: %+v", e)
	}
	if (&Estilo{Colores: map[string]string{"texto": " "}}).limpiar() != nil {
		t.Fatal("un estilo vacío tiene que quedar nil")
	}
}
