package configuracion_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/configuracion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cupones"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

func TestEntregas(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	sol := pruebadb.CrearTienda(t, b.App, "sol")

	// Sin configurar: retiro en el local con la dirección y el horario del negocio, sin moto.
	if err := configuracion.GuardarNegocio(ctx, b.App, luna, configuracion.Negocio{Nombre: "Luna",
		Direccion: "Gorriti 4500, Palermo", Horario: "Lun a vie de 10 a 19"}); err != nil {
		t.Fatal(err)
	}
	e, err := configuracion.LeerEntregas(ctx, b.App, luna)
	if err != nil || !e.Retiro || e.RetiroDireccion != "Gorriti 4500, Palermo" || e.RetiroHorario != "Lun a vie de 10 a 19" ||
		e.Moto || e.Correo || e.GratisDesde != nil {
		t.Fatalf("de entrada: %+v %v", e, err)
	}

	// Se guarda limpio; una dirección propia pisa la del negocio.
	nueva := configuracion.Entregas{Retiro: true, RetiroDireccion: "  Thames 1200 ", RetiroHorario: "Sáb de 10 a 14",
		Moto: true, MotoCosto: 3500, MotoZona: " Palermo, Villa Crespo ", MotoDemora: "En el día"}
	if err := configuracion.GuardarEntregas(ctx, b.App, luna, nueva); err != nil {
		t.Fatal(err)
	}
	if err := cupones.FijarAutomatica(ctx, b.App, luna, "envio", true); err != nil {
		t.Fatal(err)
	}
	e, err = configuracion.LeerEntregas(ctx, b.App, luna)
	if err != nil || e.RetiroDireccion != "Thames 1200" || e.MotoZona != "Palermo, Villa Crespo" || e.MotoCosto != 3500 ||
		e.GratisDesde == nil || *e.GratisDesde != cupones.EnvioGratisPorDefecto {
		t.Fatalf("guardado: %+v %v", e, err)
	}
	// Dirección vacía: vuelve la del negocio.
	nueva.RetiroDireccion = ""
	if err := configuracion.GuardarEntregas(ctx, b.App, luna, nueva); err != nil {
		t.Fatal(err)
	}
	if e, _ = configuracion.LeerEntregas(ctx, b.App, luna); e.RetiroDireccion != "Gorriti 4500, Palermo" {
		t.Fatalf("sin dirección propia: %+v", e)
	}

	// Al menos una forma de entrega; moto sin zona; retiro sin ninguna dirección; topes.
	if err := configuracion.GuardarEntregas(ctx, b.App, luna, configuracion.Entregas{}); !errors.Is(err, configuracion.ErrSinEntrega) {
		t.Fatalf("todo apagado: %v", err)
	}
	var mal configuracion.ErrEntrega
	if err := configuracion.GuardarEntregas(ctx, b.App, luna, configuracion.Entregas{Moto: true, MotoCosto: 100}); !errors.As(err, &mal) {
		t.Fatalf("moto sin zona: %v", err)
	}
	if err := configuracion.GuardarEntregas(ctx, b.App, sol, configuracion.Entregas{Retiro: true}); !errors.As(err, &mal) {
		t.Fatalf("retiro sin dirección: %v", err)
	}
	if err := configuracion.GuardarEntregas(ctx, b.App, luna, configuracion.Entregas{Moto: true, MotoZona: "CABA", MotoCosto: -1}); !errors.Is(err, configuracion.ErrInvalido) {
		t.Fatalf("costo negativo: %v", err)
	}

	// Con Correo conectado se pueden apagar retiro y moto.
	if err := b.App.EnTienda(ctx, sol, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO ca_conexiones (tienda_id, cliente, email, origen, paquete)
			VALUES ($1, '0090000025', 'sol@ejemplo.com', '{"cp":"1414"}', '{"peso":500,"alto":10,"ancho":10,"largo":10}')`, sol)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if err := configuracion.GuardarEntregas(ctx, b.App, sol, configuracion.Entregas{}); err != nil {
		t.Fatalf("sólo Correo: %v", err)
	}
	if e, err := configuracion.LeerEntregas(ctx, b.App, sol); err != nil || e.Retiro || e.Moto || !e.Correo {
		t.Fatalf("sólo Correo: %+v %v", e, err)
	}
	// Cada tienda lo suyo.
	if e, _ := configuracion.LeerEntregas(ctx, b.App, luna); !e.Moto || e.Correo {
		t.Fatalf("se mezcló con otra tienda: %+v", e)
	}

	// El envío gratis con el monto de la tienda: se cambia desde Envíos; sin mandarlo, queda como estaba.
	monto := int64(65000)
	nueva.EnvioGratisMonto = &monto
	if err := configuracion.GuardarEntregas(ctx, b.App, luna, nueva); err != nil {
		t.Fatal(err)
	}
	e, _ = configuracion.LeerEntregas(ctx, b.App, luna)
	if e.GratisDesde == nil || *e.GratisDesde != 65000 || !*e.EnvioGratis || *e.EnvioGratisMonto != 65000 {
		t.Fatalf("monto propio: %+v", e)
	}
	apagado := false
	nueva.EnvioGratis, nueva.EnvioGratisMonto = &apagado, nil
	if err := configuracion.GuardarEntregas(ctx, b.App, luna, nueva); err != nil {
		t.Fatal(err)
	}
	if e, _ = configuracion.LeerEntregas(ctx, b.App, luna); e.GratisDesde != nil || *e.EnvioGratis || *e.EnvioGratisMonto != 65000 {
		t.Fatalf("apagado conserva el monto: %+v", e)
	}
	if e, _ = configuracion.LeerEntregas(ctx, b.App, sol); *e.EnvioGratisMonto != cupones.EnvioGratisPorDefecto || *e.EnvioGratis {
		t.Fatalf("la otra tienda, por defecto: %+v", e)
	}
	chico := int64(10)
	nueva.EnvioGratisMonto = &chico
	if err := configuracion.GuardarEntregas(ctx, b.App, luna, nueva); !errors.As(err, &mal) {
		t.Fatalf("monto fuera de rango: %v", err)
	}
}
