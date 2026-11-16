package vidriera_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cobros"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/configuracion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cupones"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/vidriera"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

// Cuotas sin interés y precio sin impuestos nacionales (tienda-nivel-tiendanube.md): la vidriera
// manda lo que eligió la tienda, las cuotas sólo con Mercado Pago, y el IVA de cada producto.
func TestCuotasEImpuestos(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	if err := configuracion.GuardarNegocio(ctx, b.App, luna, configuracion.Negocio{Nombre: "Luna", Condicion: "Monotributo"}); err != nil {
		t.Fatal(err)
	}
	diez, iva := 10, 10.5
	if _, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Yerba", Precio: 5000, Stock: &diez, IVA: &iva}); err != nil {
		t.Fatal(err)
	}
	buzo, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Buzo", Precio: 24000,
		Opciones: []listas.Opcion{{Nombre: "Talle", Valores: []string{"M"}}}})
	if err != nil || buzo.IVA != 21 {
		t.Fatalf("buzo: %v %v", buzo.IVA, err)
	}
	mal := 15.0
	if _, err := listas.Guardar(ctx, b.App, luna, buzo.ID, listas.Editable{Nombre: "Buzo", Precio: 24000, IVA: &mal}); err == nil {
		t.Fatal("un IVA de 15 % no se guarda")
	}
	leer := func() vidriera.Tienda {
		t.Helper()
		tienda, err := vidriera.Publica(ctx, b.App, "luna")
		if err != nil {
			t.Fatal(err)
		}
		return tienda
	}

	// De entrada: sin cuotas, y monotributo no muestra el precio sin impuestos.
	tienda := leer()
	if tienda.Cobros["cuotas"] != 0 || tienda.Cobros["cuotasDesde"] != int64(0) ||
		tienda.Cobros["impuestos"] != (cobros.Impuestos{Mostrar: false}) {
		t.Fatalf("de entrada: %+v", tienda.Cobros)
	}
	for _, p := range tienda.Productos {
		quiero := 21.0
		if p.Nombre == "Yerba" {
			quiero = 10.5
		}
		if p.IVA != quiero || (len(p.Variantes) > 0 && p.Variantes[0].IVA != quiero) {
			t.Fatalf("iva de %s: %+v", p.Nombre, p)
		}
	}

	// Una responsable inscripta lo muestra sin elegir nada.
	if err := configuracion.GuardarNegocio(ctx, b.App, luna, configuracion.Negocio{Nombre: "Luna", Condicion: "Responsable inscripta"}); err != nil {
		t.Fatal(err)
	}
	if leer().Cobros["impuestos"] != (cobros.Impuestos{Mostrar: true}) {
		t.Fatal("responsable inscripta: muestra el precio sin impuestos")
	}

	// Elige 12 cuotas desde $ 50.000 y apaga los impuestos: sin Mercado Pago, la tienda no las ofrece.
	if err := configuracion.GuardarCobros(ctx, b.App, luna, cobros.Config{Cuotas: 12, CuotasDesde: 50000}); err != nil {
		t.Fatal(err)
	}
	tienda = leer()
	if tienda.Cobros["cuotas"] != 0 || tienda.Cobros["impuestos"] != (cobros.Impuestos{Mostrar: false}) {
		t.Fatalf("sin Mercado Pago: %+v", tienda.Cobros)
	}
	if err := b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO mp_conexiones (tienda_id, mp_usuario, acceso, renovacion, vence_en)
			VALUES ($1, 4242, '\x00', '\x00', now() + interval '1 day')`, luna)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	tienda = leer()
	if tienda.Cobros["cuotas"] != 12 || tienda.Cobros["cuotasDesde"] != int64(50000) {
		t.Fatalf("con Mercado Pago: %+v", tienda.Cobros)
	}

	// Es el mismo interruptor que la automática de Cupones: apagada, 0; prendida, vuelven las 12.
	if err := cupones.FijarAutomatica(ctx, b.App, luna, "cuotas", false); err != nil {
		t.Fatal(err)
	}
	if leer().Cobros["cuotas"] != 0 {
		t.Fatal("apagada en Cupones")
	}
	if err := cupones.FijarAutomatica(ctx, b.App, luna, "cuotas", true); err != nil {
		t.Fatal(err)
	}
	if leer().Cobros["cuotas"] != 12 {
		t.Fatal("prendida en Cupones vuelve a las elegidas")
	}

	// Lo que no es una opción no se guarda.
	for _, c := range []cobros.Config{{Cuotas: 4}, {Cuotas: 6, CuotasDesde: -1}, {Cuotas: 3, CuotasDesde: cobros.MaxDesde + 1}} {
		if err := configuracion.GuardarCobros(ctx, b.App, luna, c); !errors.Is(err, cobros.ErrInvalido) {
			t.Errorf("%+v: %v", c, err)
		}
	}
}
