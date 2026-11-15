package mitienda

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// Los casos compartidos con la web (lib/tienda/sistema/migrar.spec.ts).
func TestMigrarCasosCompartidos(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "..", "packages", "contrato", "casos", "diseno-v2.json"))
	if err != nil {
		t.Fatal(err)
	}
	var casos struct {
		Casos []struct {
			Nombre    string          `json:"nombre"`
			Plantilla string          `json:"plantilla"`
			V1        json.RawMessage `json:"v1"`
			V2        any             `json:"v2"`
		} `json:"casos"`
	}
	if err := json.Unmarshal(b, &casos); err != nil {
		t.Fatal(err)
	}
	if len(casos.Casos) == 0 {
		t.Fatal("sin casos")
	}
	for _, c := range casos.Casos {
		var d *Diseno
		if string(c.V1) != "null" {
			d = &Diseno{}
			if err := json.Unmarshal(c.V1, d); err != nil {
				t.Fatalf("%s: %v", c.Nombre, err)
			}
		}
		m := Migrar(d, c.Plantilla)
		var got any
		j, _ := json.Marshal(m)
		_ = json.Unmarshal(j, &got)
		if !reflect.DeepEqual(got, c.V2) {
			w, _ := json.Marshal(c.V2)
			t.Errorf("%s:\n got %s\nwant %s", c.Nombre, j, w)
		}
		// Lo migrado se puede guardar y migrarlo de nuevo no cambia nada.
		if !m.Valido() {
			t.Errorf("%s: lo migrado no es válido: %s", c.Nombre, j)
		}
		otra, _ := json.Marshal(Migrar(&m, c.Plantilla))
		if string(otra) != string(j) {
			t.Errorf("%s: migrar dos veces cambia: %s", c.Nombre, otra)
		}
	}
}

// Todas las plantillas tienen preset y su inicial es válido.
func TestInicialDeCadaPreset(t *testing.T) {
	if len(Presets) != 17 {
		t.Fatalf("presets: %d", len(Presets))
	}
	for id := range Presets {
		if d := Inicial(id); !d.Valido() || len(d.Inicio) == 0 {
			t.Errorf("%s: inicial inválido %+v", id, d)
		}
	}
}
