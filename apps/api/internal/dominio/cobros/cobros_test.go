package cobros_test

import (
	"errors"
	"testing"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cobros"
)

func TestCondicionYValidar(t *testing.T) {
	for c, quiero := range map[string]bool{"Responsable inscripta": true, "responsable_inscripto": true, "RI": true,
		"Monotributo": false, "monotributo": false, "exento": false, "": false} {
		if cobros.ResponsableInscripto(c) != quiero {
			t.Errorf("%q: quería %v", c, quiero)
		}
	}
	for _, c := range []cobros.Config{{}, {Cuotas: 3, CuotasDesde: 30000}, {Cuotas: 12}} {
		if err := c.Validar(); err != nil {
			t.Errorf("%+v: %v", c, err)
		}
	}
	for _, c := range []cobros.Config{{Cuotas: 1}, {Cuotas: 18}, {Cuotas: 6, CuotasDesde: -1}} {
		if !errors.Is(c.Validar(), cobros.ErrInvalido) {
			t.Errorf("%+v pasó", c)
		}
	}
	if !cobros.IVAValido(10.5) || !cobros.IVAValido(0) || cobros.IVAValido(19) {
		t.Fatal("iva")
	}
	if d := (cobros.Config{CuotasDesde: 30000}).Detalle(6); d != "6 cuotas sin interés desde $ 30.000" {
		t.Fatalf("detalle: %q", d)
	}
}
