package cupones_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cupones"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

var ctx = context.Background()

func nuevo(codigo string) cupones.Nuevo {
	n := cupones.Nuevo{Codigo: codigo, Tipo: "pct", Valor: 15, Alcance: "todo", Quien: "todos", Limite: 100,
		PorCliente: "1", Vigencia: "30", Canales: []string{"web"}}
	n.Opciones.Hornero = true
	return n
}

func TestCuponesCrearPausarDuplicarYMinimo(t *testing.T) {
	b := pruebadb.Nueva(t)
	id := pruebadb.CrearTienda(t, b.App, "luna")
	otra := pruebadb.CrearTienda(t, b.App, "sol")
	ahora := time.Date(2026, 10, 22, 18, 0, 0, 0, time.UTC)

	c, err := cupones.Crear(ctx, b.App, id, nuevo("FINDE15"), ahora)
	if err != nil || c.Estado != "activo" || c.Descripcion != "15 % off en toda la tienda" || c.Consejo == nil {
		t.Fatalf("crear: %+v %v", c, err)
	}
	if _, err := cupones.Crear(ctx, b.App, id, nuevo("FINDE15"), ahora); !errors.Is(err, cupones.ErrRepetido) {
		t.Fatalf("repetido: %v", err)
	}
	// El mismo código en otra tienda sí.
	if _, err := cupones.Crear(ctx, b.App, otra, nuevo("FINDE15"), ahora); err != nil {
		t.Fatal(err)
	}
	if _, err := cupones.Crear(ctx, b.App, id, cupones.Nuevo{Codigo: "x"}, ahora); !errors.Is(err, cupones.ErrInvalido) {
		t.Fatalf("inválido: %v", err)
	}
	// El porqué llega a la persona: $10 fijo (lo que mandaba la app por defecto) o 95 %.
	fijo := nuevo("DIEZ")
	fijo.Tipo, fijo.Valor = "fijo", 10
	var dato *cupones.ErrDato
	if _, err := cupones.Crear(ctx, b.App, id, fijo, ahora); !errors.As(err, &dato) || !strings.Contains(dato.Mensaje, "$100") || !errors.Is(err, cupones.ErrInvalido) {
		t.Fatalf("fijo chico: %v", err)
	}
	pct := nuevo("MUCHO")
	pct.Tipo, pct.Valor = "pct", 95
	if _, err := cupones.Crear(ctx, b.App, id, pct, ahora); !errors.As(err, &dato) || !strings.Contains(dato.Mensaje, "1 a 90") {
		t.Fatalf("porcentaje alto: %v", err)
	}
	if p, _ := cupones.Accion(ctx, b.App, id, c.ID, "pausar", 0, ahora); p.Estado != "pausado" {
		t.Fatalf("pausar: %+v", p)
	}
	d, err := cupones.Accion(ctx, b.App, id, c.ID, "duplicar", 0, ahora)
	if err != nil || d.Codigo != "FINDE15-2" || d.Estado != "pausado" || d.ID == c.ID {
		t.Fatalf("duplicar: %+v %v", d, err)
	}
	m, _ := cupones.Accion(ctx, b.App, id, c.ID, "minimo", 25000, ahora)
	if m.Minimo != 25000 || m.Descripcion != "15 % off en toda la tienda · mín. $ 25.000" {
		t.Fatalf("mínimo: %+v", m)
	}
	if _, err := cupones.Accion(ctx, b.App, otra, c.ID, "pausar", 0, ahora); !errors.Is(err, cupones.ErrNoExiste) {
		t.Fatalf("otra tienda tocó el cupón: %v", err)
	}

	// Programado: arranca el lunes a las 9; vencido, termina.
	n := nuevo("LUNES10")
	n.Programar = true
	p, _ := cupones.Crear(ctx, b.App, id, n, ahora)
	if p.Estado != "programado" || !p.Empieza.Equal(time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)) {
		t.Fatalf("programado: %+v", p)
	}
	res, _ := cupones.Leer(ctx, b.App, id, ahora.Add(40*24*time.Hour))
	for _, x := range res["cupones"].([]cupones.Cupon) {
		if x.Codigo == "LUNES10" && x.Estado != "terminado" {
			t.Fatalf("no terminó: %+v", x)
		}
	}
	if err := cupones.FijarAutomatica(ctx, b.App, id, "cumple", true); err != nil {
		t.Fatal(err)
	}
	if err := cupones.FijarAutomatica(ctx, b.App, id, "nada", true); !errors.Is(err, cupones.ErrNoExiste) {
		t.Fatal(err)
	}
}
