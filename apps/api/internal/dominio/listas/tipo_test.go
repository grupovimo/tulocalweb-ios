package listas

import "testing"

func TestLimpiarTipo(t *testing.T) {
	tipo := " hogar "
	tp, a, err := limpiarTipo(&tipo, map[string]string{"peso": " 1200 ", "alto": "40", "material": "Pino", "marca": ""})
	if err != nil || *tp != "hogar" || a["peso"] != "1200" || a["material"] != "Pino" || len(a) != 3 {
		t.Fatalf("tipo %v, atributos %v, error %v", tp, a, err)
	}
	if _, a, _ := limpiarTipo(nil, nil); a != nil {
		t.Fatal("sin atributos tiene que quedar nil (deja los que había)")
	}
	malo := "auto"
	for _, c := range []struct {
		tipo *string
		a    map[string]string
	}{
		{&malo, nil},
		{nil, map[string]string{"color": "rojo"}},
		{nil, map[string]string{"peso": "1,5 kg"}},
		{nil, map[string]string{"alto": "-3"}},
	} {
		if _, _, err := limpiarTipo(c.tipo, c.a); err == nil {
			t.Errorf("%v %v tenía que dar error", c.tipo, c.a)
		}
	}
}
