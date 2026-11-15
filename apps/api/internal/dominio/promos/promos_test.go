package promos

import (
	"strings"
	"testing"
	"time"
)

func TestSortearRespetaLosPesos(t *testing.T) {
	if Sortear([]int{0, 0}) != -1 || Sortear(nil) != -1 {
		t.Fatal("sin pesos no sale ninguna")
	}
	cuenta := make([]int, 4)
	for range 4000 {
		i := Sortear([]int{70, 0, 30, 0})
		if i < 0 {
			t.Fatal("salió -1 con pesos")
		}
		cuenta[i]++
	}
	if cuenta[1] != 0 || cuenta[3] != 0 {
		t.Fatalf("salió una porción de peso 0: %v", cuenta)
	}
	if cuenta[0] < 2500 || cuenta[0] > 3100 {
		t.Fatalf("la de 70 salió %d de 4000", cuenta[0])
	}
}

func TestProvincias(t *testing.T) {
	for cp, p := range map[string]string{
		"C1425ABC": "CABA", "x5000": "Córdoba", "1425": "CABA", "5000": "Córdoba", "7600": "Buenos Aires",
		"8400": "Río Negro", "2900": "", "": "", "abc": "", "12345": "", "9410": "Tierra del Fuego",
	} {
		if got := ProvinciaDeCP(cp); got != p {
			t.Errorf("ProvinciaDeCP(%q) = %q, quería %q", cp, got, p)
		}
	}
	if provinciaDe("Mendoza", "X", "1425") != "Mendoza" || provinciaDe("", "X", "1425") != "Córdoba" || provinciaDe("", "", "1425") != "CABA" {
		t.Fatal("provinciaDe: la ficha, después Correo, después el CP")
	}
}

func TestNombrePila(t *testing.T) {
	for n, p := range map[string]string{
		"ana maría gómez": "Ana", "  LUCÍA  ": "Lucía", "Consumidor final": "", "": "", "a@b.com": "",
		"J": "", "R2D2 Robot": "", "María-José Paz": "María-josé",
	} {
		if got := nombrePila(n); got != p {
			t.Errorf("nombrePila(%q) = %q, quería %q", n, got, p)
		}
	}
}

func TestRango(t *testing.T) {
	ahora := time.Date(2026, 11, 15, 2, 0, 0, 0, time.UTC) // 14/11 en Argentina
	d, h, ok := Rango("", "", ahora)
	if !ok || h.Format(time.DateOnly) != "2026-11-14" || d.Format(time.DateOnly) != "2026-10-16" {
		t.Fatalf("por defecto: %v %v %v", d, h, ok)
	}
	if _, _, ok := Rango("2026-11-10", "2026-11-01", ahora); ok {
		t.Fatal("hasta antes que desde")
	}
	if _, _, ok := Rango("2024-01-01", "2026-01-01", ahora); ok {
		t.Fatal("más de un año")
	}
	if _, _, ok := Rango("ayer", "", ahora); ok {
		t.Fatal("fecha mal escrita")
	}
}

func TestMailPremio(t *testing.T) {
	vence := time.Date(2026, 11, 22, 15, 0, 0, 0, time.UTC)
	m := MailPremio("https://tulocalweb.com.ar/", DatosMail{Tienda: "Lumbre <b>", Slug: "lumbre", Acento: "#C2410C", SobreAcento: "#FFFFFF",
		Para: "ana@example.com", Titulo: "10 % en tu primera compra", Confirmar: "tok123",
		Premio: Premio{Tipo: "cuponUnico", Codigo: "BIENVE-ABCD2345", Descripcion: "10 % off", Vence: &vence}})
	for _, s := range []string{"BIENVE-ABCD2345", "22 de noviembre", "/t/lumbre/newsletter/confirmar/tok123", "#C2410C", "Lumbre &lt;b&gt;"} {
		if !strings.Contains(m.HTML, s) {
			t.Errorf("el HTML no tiene %q", s)
		}
	}
	if !strings.Contains(m.Texto, "BIENVE-ABCD2345") || m.DeNombre != "Lumbre <b>" || !strings.HasPrefix(m.Asunto, "Tu premio") {
		t.Fatalf("mail: %+v", m)
	}
	d := MailPremio("https://tulocalweb.com.ar", DatosMail{Tienda: "Verde", Slug: "verde", Acento: "#000", SobreAcento: "#fff",
		Premio: Premio{Tipo: "descarga", Descarga: "https://x.com/guia.pdf", NombreArchivo: "la guía"}})
	if !strings.Contains(d.HTML, "https://x.com/guia.pdf") || strings.Contains(d.HTML, "newsletter/confirmar") || !strings.HasPrefix(d.Asunto, "Tu descarga") {
		t.Fatalf("descarga: %+v", d)
	}
}
