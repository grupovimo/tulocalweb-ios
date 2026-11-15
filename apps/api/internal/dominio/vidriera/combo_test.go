package vidriera

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/mitienda"
)

// Los mismos casos que la web (lib/tienda/combos.spec.ts).
func TestDescuentoCombos(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "..", "packages", "contrato", "casos", "combos.json"))
	if err != nil {
		t.Fatal(err)
	}
	var casos struct {
		Casos []struct {
			Nombre    string       `json:"nombre"`
			Combos    []Combo      `json:"combos"`
			Lineas    []LineaCombo `json:"lineas"`
			Descuento int64        `json:"descuento"`
		} `json:"casos"`
	}
	if err := json.Unmarshal(b, &casos); err != nil {
		t.Fatal(err)
	}
	for _, c := range casos.Casos {
		if got := DescuentoCombos(c.Combos, c.Lineas); got != c.Descuento {
			t.Errorf("%s: descuento %d, se esperaba %d", c.Nombre, got, c.Descuento)
		}
	}
}

func TestCombosDe(t *testing.T) {
	b := func(id string, datos map[string]any, oculto bool) mitienda.Bloque {
		return mitienda.Bloque{ID: id, Tipo: "combo", Variante: "tarjeta", Oculto: oculto, Datos: datos}
	}
	d := mitienda.Diseno{Inicio: []mitienda.Bloque{
		b("a", map[string]any{"productos": []any{"x", "y"}, "descuento": 10.0}, false),
		b("b", map[string]any{"productos": []any{"x", "y"}, "descuento": 0.0}, false),
		b("c", map[string]any{"productos": []any{"x"}, "descuento": 10.0}, false),
		b("d", map[string]any{"productos": []any{"x", "y"}, "descuento": 10.0}, true),
	}}
	got := CombosDe(d)
	if len(got) != 1 || got[0].ID != "a" || got[0].Descuento != 10 || len(got[0].Productos) != 2 {
		t.Fatalf("combos: %+v", got)
	}
}
