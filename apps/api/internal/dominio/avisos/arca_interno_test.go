package avisos

import (
	"strings"
	"testing"
	"time"
)

func TestMailArcaVence(t *testing.T) {
	vence := time.Date(2026, 11, 7, 15, 0, 0, 0, time.UTC)
	m := mailArcaVence("https://tulocalweb.com.ar", "Luna", 7, vence)
	if m.Asunto != "Tu certificado de ARCA vence en 7 días" || !strings.Contains(m.Texto, "07/11/2026") ||
		!strings.Contains(m.Texto, "https://tulocalweb.com.ar/panel/integraciones?arca") {
		t.Fatalf("a 7 días: %q\n%s", m.Asunto, m.Texto)
	}
	if m := mailArcaVence("https://tulocalweb.com.ar/", "Luna", 1, vence); m.Asunto != "Tu certificado de ARCA vence mañana" {
		t.Fatalf("mañana: %q", m.Asunto)
	}
	m = mailArcaVence("https://tulocalweb.com.ar", "Luna", 0, vence)
	if m.Asunto != "Tu certificado de ARCA venció" || !strings.Contains(m.Texto, "quedan sin facturar") {
		t.Fatalf("vencido: %q\n%s", m.Asunto, m.Texto)
	}
}
