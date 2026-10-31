package mayorista_test

import (
	"errors"
	"testing"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/mayorista"
)

func ptr[T any](v T) *T { return &v }

// Los mismos casos que apps/web/src/lib/tienda/mayorista.spec.ts.
func TestAlcanza(t *testing.T) {
	casos := []struct {
		nombre   string
		c        mayorista.Config
		unidades int
		monto    int64
		quiero   bool
	}{
		{"apagada", mayorista.Config{Unidades: ptr(6)}, 10, 0, false},
		{"sin mínimos", mayorista.Config{Activa: true}, 10, 100000, false},
		{"unidades justas", mayorista.Config{Activa: true, Unidades: ptr(6)}, 6, 0, true},
		{"falta una unidad", mayorista.Config{Activa: true, Unidades: ptr(6)}, 5, 999999, false},
		{"monto justo", mayorista.Config{Activa: true, Monto: ptr(int64(30000))}, 1, 30000, true},
		{"falta monto", mayorista.Config{Activa: true, Monto: ptr(int64(30000))}, 50, 29999, false},
		{"los dos: cumple los dos", mayorista.Config{Activa: true, Unidades: ptr(6), Monto: ptr(int64(30000))}, 6, 30000, true},
		{"los dos: sólo unidades", mayorista.Config{Activa: true, Unidades: ptr(6), Monto: ptr(int64(30000))}, 6, 29000, false},
		{"los dos: sólo monto", mayorista.Config{Activa: true, Unidades: ptr(6), Monto: ptr(int64(30000))}, 5, 90000, false},
	}
	for _, c := range casos {
		if got := c.c.Alcanza(c.unidades, c.monto); got != c.quiero {
			t.Errorf("%s: %v, quería %v", c.nombre, got, c.quiero)
		}
	}
}

func TestPrecioYDetalle(t *testing.T) {
	if mayorista.Precio(10000, nil) != 10000 || mayorista.Precio(10000, ptr(int64(8000))) != 8000 ||
		mayorista.Precio(10000, ptr(int64(12000))) != 10000 || mayorista.Precio(10000, ptr(int64(0))) != 10000 {
		t.Fatal("precio: el mayorista sólo si es menor")
	}
	if *mayorista.DeVariante(nil, ptr(int64(5)), nil) != 5 || *mayorista.DeVariante(ptr(int64(3)), ptr(int64(5)), ptr(int64(9))) != 3 ||
		mayorista.DeVariante(nil, nil, nil) != nil || mayorista.DeVariante(nil, ptr(int64(5)), ptr(int64(9))) != nil {
		t.Fatal("variante: la suya o la del producto")
	}
	for c, quiero := range map[*mayorista.Config]string{
		{Unidades: ptr(12)}:                           "Desde 12 unidades",
		{Monto: ptr(int64(50000))}:                    "Desde $ 50.000",
		{Unidades: ptr(6), Monto: ptr(int64(150000))}: "Desde 6 unidades y $ 150.000",
	} {
		if got := c.Detalle(); got != quiero {
			t.Errorf("detalle: %q, quería %q", got, quiero)
		}
	}
	if !errors.Is((mayorista.Config{Activa: true}).Validar(), mayorista.ErrSinMinimo) ||
		!errors.Is((mayorista.Config{Unidades: ptr(1)}).Validar(), mayorista.ErrInvalido) ||
		!errors.Is((mayorista.Config{Monto: ptr(int64(0))}).Validar(), mayorista.ErrInvalido) ||
		(mayorista.Config{Unidades: ptr(2), Monto: ptr(int64(1))}).Validar() != nil {
		t.Fatal("validar")
	}
}
