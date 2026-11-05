package abono

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"
	"time"
)

// La misma tabla que usa el espejo de TypeScript (apps/web/src/lib/panel/cotizar.casos.json).
type caso struct {
	Nombre string          `json:"nombre"`
	Estado Estado          `json:"estado"`
	Ahora  time.Time       `json:"ahora"`
	Plan   string          `json:"plan"`
	Ciclo  string          `json:"ciclo"`
	Espera json.RawMessage `json:"espera"`
	Error  string          `json:"error"`
}

func TestCotizar(t *testing.T) {
	b, err := os.ReadFile("testdata/cotizar.json")
	if err != nil {
		t.Fatal(err)
	}
	var casos []caso
	if err := json.Unmarshal(b, &casos); err != nil {
		t.Fatal(err)
	}
	for _, c := range casos {
		t.Run(c.Nombre, func(t *testing.T) {
			got, err := Cotizar(c.Estado, c.Plan, c.Ciclo, c.Ahora)
			switch c.Error {
			case "invalido":
				if !errors.Is(err, ErrInvalido) {
					t.Fatalf("esperaba ErrInvalido, vino %v", err)
				}
				return
			case "programado":
				if !errors.Is(err, ErrProgramado) {
					t.Fatalf("esperaba ErrProgramado, vino %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			var espera Cotizacion
			if err := json.Unmarshal(c.Espera, &espera); err != nil {
				t.Fatal(err)
			}
			espera.Plan, espera.Ciclo = c.Plan, c.Ciclo
			if !reflect.DeepEqual(normal(got), normal(espera)) {
				g, _ := json.Marshal(got)
				t.Fatalf("\nvino:   %s\nespera: %s", g, c.Espera)
			}
			// Las líneas suman lo de hoy.
			var suma int64
			for _, l := range got.Lineas {
				suma += l.Importe
			}
			if len(got.Lineas) > 0 && suma != got.Hoy {
				t.Fatalf("las líneas suman %d y hoy es %d", suma, got.Hoy)
			}
		})
	}
}

// normal compara por JSON (las fechas en UTC, nil y vacío iguales).
func normal(c Cotizacion) map[string]any {
	b, _ := json.Marshal(c)
	var m map[string]any
	_ = json.Unmarshal(b, &m)
	return m
}

func TestPrecio(t *testing.T) {
	if p, _ := Precio("pro", "anio"); p != 49_900*9 {
		t.Fatalf("anual: %d", p)
	}
	if _, ok := Precio("pro", "semana"); ok {
		t.Fatal("ciclo inválido")
	}
}
