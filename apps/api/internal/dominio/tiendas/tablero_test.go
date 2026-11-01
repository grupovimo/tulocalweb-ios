package tiendas_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/tiendas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

func TestVentana(t *testing.T) {
	ahora := time.Date(2026, 10, 20, 2, 30, 0, 0, time.UTC) // 23:30 del 19 en Argentina
	desde, hasta, err := tiendas.Ventana("hoy", ahora)
	if err != nil {
		t.Fatal(err)
	}
	if want := time.Date(2026, 10, 19, 3, 0, 0, 0, time.UTC); !desde.Equal(want) || !hasta.Equal(ahora) {
		t.Fatalf("hoy: %s – %s", desde, hasta)
	}
	if d, _, _ := tiendas.Ventana("7d", ahora); !d.Equal(ahora.Add(-7 * 24 * time.Hour)) {
		t.Fatalf("7d: %s", d)
	}
	if _, _, err := tiendas.Ventana("anual", ahora); err == nil {
		t.Fatal("aceptó un período inválido")
	}
}

func TestMetaPorTienda(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	una, _ := tiendas.Crear(ctx, b.App, "una", "Una", "moda")
	otra, _ := tiendas.Crear(ctx, b.App, "otra", "Otra", "moda")
	if err := tiendas.GuardarMeta(ctx, b.App, una, 20_000_000); err != nil {
		t.Fatal(err)
	}
	if err := tiendas.GuardarMeta(ctx, b.App, una, -1); !errors.Is(err, tiendas.ErrMeta) {
		t.Fatalf("meta negativa: %v", err)
	}
	tu, err := tiendas.ObtenerTablero(ctx, b.App, una, "7d", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	to, _ := tiendas.ObtenerTablero(ctx, b.App, otra, "7d", time.Now())
	if tu.Meta.Objetivo != 20_000_000 || to.Meta.Objetivo != 0 {
		t.Fatalf("metas: %d / %d", tu.Meta.Objetivo, to.Meta.Objetivo)
	}
	if !tu.Montos || len(tu.Serie.Actual) != 14 || tu.UltimosPedidos == nil || tu.Avisos.CertificadoDias != nil {
		t.Fatalf("tablero vacío mal armado: %+v", tu)
	}
}

// El Inicio cuenta las ventas de verdad: un pedido pagado entra, uno sin pagar o cancelado no.
func TestTableroConVentas(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	id, _ := tiendas.Crear(ctx, b.App, "almacen", "Almacén", "almacen")
	ahora := time.Now()
	mail := "ana@ejemplo.com"
	pedido := func(estado, canal string, total int64, hace time.Duration) {
		t.Helper()
		if _, err := pedidos.Crear(ctx, b.App, id, pedidos.Nuevo{
			Comprador: pedidos.Comprador{Nombre: "Ana", Email: &mail}, Canal: canal, Pago: "transf", Envio: "retiro",
			Estado: estado, Creado: ahora.Add(-hace),
			Items: []pedidos.Item{{Nombre: "Yerba 1 kg", Cantidad: 2, Precio: total / 2}},
		}); err != nil {
			t.Fatal(err)
		}
	}
	pedido("Entregado", "web", 15_000, time.Hour)
	pedido("Para preparar", "wa", 5_000, 2*time.Hour)
	pedido("Sin pagar", "web", 99_000, 3*time.Hour)
	pedido("Cancelado", "web", 77_000, 4*time.Hour)

	tb, err := tiendas.ObtenerTablero(ctx, b.App, id, "7d", ahora)
	if err != nil {
		t.Fatal(err)
	}
	if tb.Ventas.Actual != 20_000 || tb.Kpis.Pedidos.Actual != 2 || tb.Kpis.Ticket.Actual != 10_000 {
		t.Fatalf("ventas %v, pedidos %v, ticket %v", tb.Ventas.Actual, tb.Kpis.Pedidos.Actual, tb.Kpis.Ticket.Actual)
	}
	if tb.Kpis.ClientesNuevos.Actual != 1 || tb.Serie.Actual[13] != 20_000 {
		t.Fatalf("clientes nuevos %v, serie %v", tb.Kpis.ClientesNuevos.Actual, tb.Serie.Actual)
	}
	if len(tb.Canales) != 2 || len(tb.UltimosPedidos) != 4 || len(tb.MasVendidos) != 1 {
		t.Fatalf("canales %v, últimos %d, más vendidos %v", tb.Canales, len(tb.UltimosPedidos), tb.MasVendidos)
	}
	// La meta es del mes en Argentina: en las primeras horas del mes, las ventas de hace 1 a 3 horas
	// son del mes anterior y no cuentan.
	ar, _ := time.LoadLocation("America/Argentina/Buenos_Aires")
	mismoMes := ahora.Add(-3*time.Hour).In(ar).Month() == ahora.In(ar).Month()
	if tb.ParaHacer.Preparar != 1 || tb.ParaHacer.Transferencias != 1 || *tb.ParaHacer.TransferenciasMonto != 99_000 || (mismoMes && tb.Meta.Actual < 15_000) {
		t.Fatalf("para hacer %+v, meta %+v", tb.ParaHacer, tb.Meta)
	}

	// Sin montos (Ventas, Local): quedan las cantidades y no va ningún monto, ni en cero.
	sin := tb
	sin.SinMontos()
	j, _ := json.Marshal(sin)
	for _, clave := range []string{`"ventas"`, `"serie"`, `"ticket"`, `"meta"`, `"caja"`, `"canales"`, `"total"`,
		`"transferenciasMonto"`, `"carritosMonto"`, `"liquidaciones"`} {
		if strings.Contains(string(j), clave) {
			t.Fatalf("sin montos quedó %s: %s", clave, j)
		}
	}
	if sin.Montos || sin.Kpis.Pedidos.Actual != 2 || sin.ParaHacer.Transferencias != 1 || len(sin.MasVendidos) != 1 {
		t.Fatalf("sin montos perdió cantidades: %+v", sin)
	}

	// Tincho ve lo mismo: ventas por día, clientes y canales del mes.
	c := tiendas.ContextoHornero(ctx, b.App, id)
	dias := c["ventasPorDia"].([]any)
	if len(dias) != 14 || c["clientes"].(map[string]int)["total"] != 1 || len(c["canales30d"].([]any)) != 2 {
		t.Fatalf("contexto de Tincho: %+v", c)
	}
	c = tiendas.ContextoSinMontos(c)
	if j, _ := json.Marshal(c); strings.Contains(string(j), `"ventas"`) || !strings.Contains(string(j), `"canal":"web","pedidos":1`) {
		t.Fatalf("contexto sin montos: %s", j)
	}
}
