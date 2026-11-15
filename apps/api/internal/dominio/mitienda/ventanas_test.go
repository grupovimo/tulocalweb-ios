package mitienda

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// packages/contrato/casos/ventanas.json: los mismos casos que corre la web.
func TestCasosVentanas(t *testing.T) {
	crudo, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "..", "packages", "contrato", "casos", "ventanas.json"))
	if err != nil {
		t.Fatal(err)
	}
	var casos struct {
		Ventanas []struct {
			Nombre   string          `json:"nombre"`
			Ventana  json.RawMessage `json:"ventana"`
			Valida   bool            `json:"valida"`
			Completa bool            `json:"completa"`
		} `json:"ventanas"`
		Listas []struct {
			Nombre   string          `json:"nombre"`
			Ventanas json.RawMessage `json:"ventanas"`
			Valida   bool            `json:"valida"`
		} `json:"listas"`
		Conversion []struct {
			Nombre     string         `json:"nombre"`
			Conversion map[string]any `json:"conversion"`
			Valida     bool           `json:"valida"`
		} `json:"conversion"`
		Barra []struct {
			Nombre string `json:"nombre"`
			Barra  Barra  `json:"barra"`
			Valida bool   `json:"valida"`
		} `json:"barra"`
	}
	if err := json.Unmarshal(crudo, &casos); err != nil {
		t.Fatal(err)
	}
	// Estricto: una clave desconocida o de otro tipo es inválida (la API la descartaría).
	leer := func(b []byte, v any) error {
		dec := json.NewDecoder(bytes.NewReader(b))
		dec.DisallowUnknownFields()
		return dec.Decode(v)
	}
	for _, c := range casos.Ventanas {
		var v Ventana
		valida := leer(c.Ventana, &v) == nil && v.valida()
		if valida != c.Valida {
			t.Errorf("%s: válida = %v, quería %v", c.Nombre, valida, c.Valida)
			continue
		}
		if valida && v.Completa() != c.Completa {
			t.Errorf("%s: completa = %v, quería %v", c.Nombre, v.Completa(), c.Completa)
		}
		if valida {
			d := Inicial("ceibo")
			d.Ventanas = []Ventana{v}
			if !d.Valido() {
				t.Errorf("%s: el diseño con la ventana no es válido", c.Nombre)
			}
		}
	}
	for _, c := range casos.Listas {
		var l []Ventana
		d := Inicial("ceibo")
		valida := leer(c.Ventanas, &l) == nil
		d.Ventanas = l
		if valida = valida && d.Valido(); valida != c.Valida {
			t.Errorf("lista %s: válida = %v, quería %v", c.Nombre, valida, c.Valida)
		}
	}
	for _, c := range casos.Conversion {
		e := &Estilo{Conversion: c.Conversion}
		if got := e.valido(); got != c.Valida {
			t.Errorf("conversión %s: válida = %v, quería %v", c.Nombre, got, c.Valida)
		}
	}
	for _, c := range casos.Barra {
		if got := c.Barra.valido(); got != c.Valida {
			t.Errorf("barra %s: válida = %v, quería %v", c.Nombre, got, c.Valida)
		}
	}
}

// Las plantillas del catálogo son ventanas válidas (completas, una vez que el editor llena lo suyo:
// eso lo prueban los casos).
func TestPlantillasDeVentanas(t *testing.T) {
	var cat struct {
		Ventanas struct {
			Plantillas []struct {
				ID      string          `json:"id"`
				Ventana json.RawMessage `json:"ventana"`
			} `json:"plantillas"`
		} `json:"ventanas"`
	}
	if err := json.Unmarshal(bloquesJSON, &cat); err != nil {
		t.Fatal(err)
	}
	if len(cat.Ventanas.Plantillas) != 7 {
		t.Errorf("plantillas: %d, quería 7", len(cat.Ventanas.Plantillas))
	}
	for _, p := range cat.Ventanas.Plantillas {
		var v Ventana
		if err := json.Unmarshal(p.Ventana, &v); err != nil {
			t.Fatal(err)
		}
		v.ID = "plantilla1"
		if !v.valida() {
			t.Errorf("plantilla %s: no es válida", p.ID)
		}
	}
}
