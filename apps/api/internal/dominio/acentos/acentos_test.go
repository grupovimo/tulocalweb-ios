package acentos

import "testing"

// Los mismos casos que apps/web/src/lib/tienda/acentos.spec.ts.

func TestCuatroOpcionesDistintasLaPrimeraEsLaSuya(t *testing.T) {
	for _, p := range ConAcento() {
		o := Opciones(p)
		if len(o) != 4 {
			t.Fatalf("%s: %d opciones", p, len(o))
		}
		vistos := map[string]bool{}
		for _, x := range o {
			vistos[x.Color] = true
		}
		if len(vistos) != 4 {
			t.Fatalf("%s: opciones repetidas %v", p, o)
		}
	}
	if c := Opciones("ceibo")[0].Color; c != "#9C4A2F" {
		t.Fatalf("ceibo: %s", c)
	}
}

func TestAlternativasConContrasteAA(t *testing.T) {
	for _, p := range ConAcento() {
		for _, x := range Opciones(p)[1:] {
			if c := Contraste(x.Color, x.Sobre); c < 4.5 {
				t.Errorf("%s %s: contraste %.2f", p, x.Color, c)
			}
		}
	}
}

func TestLaOpcionCeroNoPisaNada(t *testing.T) {
	if e := Elegido("yerbal", 0); e != "" {
		t.Fatalf("yerbal 0: %q", e)
	}
	if e := Elegido("yerbal", 1); e != "#B0441D" {
		t.Fatalf("yerbal 1: %q", e)
	}
}

func TestDeLaTienda(t *testing.T) {
	casos := []struct {
		plantilla      string
		paleta         int
		acento, sobre  string
		quiere, encima string
	}{
		{"pampa", 0, "", "", "#E6FF4D", "#111111"},
		{"yerbal", 1, "", "", "#B0441D", "#FFFFFF"},
		{"mostaza", 2, "", "", "#D8B23A", "#2A170C"},
		{"no-existe", 0, "", "", "#6B3A22", "#FFFFFF"},
		{"yerbal", 7, "", "", "#2F5D50", "#FFFFFF"},
		// Lo pisado en Estilo manda; si el texto de la plantilla no se lee, va el que contrasta.
		{"yerbal", 1, "#f5e663", "", "#F5E663", "#111111"},
		{"yerbal", 0, "#123456", "", "#123456", "#FFFFFF"},
		{"pampa", 0, "", "#222222", "#E6FF4D", "#222222"},
		{"pampa", 0, "rojo", "", "#E6FF4D", "#111111"},
	}
	for _, c := range casos {
		a, s := DeLaTienda(c.plantilla, c.paleta, c.acento, c.sobre)
		if a != c.quiere || s != c.encima {
			t.Errorf("%v: %s %s", c, a, s)
		}
	}
}
