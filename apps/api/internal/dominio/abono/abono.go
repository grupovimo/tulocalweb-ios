// Package abono: los precios del abono de Tu Local Web y la cotización de un cambio de plan (qué se
// paga hoy, desde cuándo rige y cuánto se cobra después). Es puro: la pantalla de Plan y facturación,
// la app y el cobro con Mercado Pago salen de la misma cuenta. Su espejo en TypeScript es
// apps/web/src/lib/panel/cotizar.ts, con la misma tabla de casos (testdata/cotizar.json).
package abono

import (
	"errors"
	"math"
	"strconv"
	"time"
)

// Precios por mes, en pesos, IVA incluido (lib/panel/planes.ts). El anual es 25 % menos, cobrado una
// vez por año.
var precios = map[string]int64{"emprende": 19_900, "pro": 49_900, "negocio": 149_900}

// Nombres para mostrar.
var Nombres = map[string]string{"gratis": "Gratis", "emprende": "Emprende", "pro": "Pro", "negocio": "Negocio"}

// DiasDePrueba de la prueba gratis, desde el alta de la tienda.
const DiasDePrueba = 14

// ErrInvalido: plan o ciclo que no existen.
var ErrInvalido = errors.New("abono: plan o ciclo inválido")

// ErrProgramado: hay un cambio programado; primero se deshace (eligiendo el plan vigente).
var ErrProgramado = errors.New("abono: hay un cambio programado")

// Precio de un plan en un ciclo: lo que se cobra cada vez. El gratis vale 0.
func Precio(plan, ciclo string) (int64, bool) {
	if ciclo != "mes" && ciclo != "anio" {
		return 0, false
	}
	if plan == "gratis" {
		return 0, true
	}
	p, ok := precios[plan]
	if !ok {
		return 0, false
	}
	if ciclo == "anio" {
		return p * 12 * 3 / 4, true
	}
	return p, true
}

// Cambio programado: el plan que rige desde una fecha (el próximo cobro).
type Cambio struct {
	Plan  string    `json:"plan"`
	Ciclo string    `json:"ciclo"`
	Desde time.Time `json:"desde"`
}

// Estado del abono de una tienda.
type Estado struct {
	Plan      string    `json:"plan"`  // el vigente
	Ciclo     string    `json:"ciclo"` // mes | anio
	FinPrueba time.Time `json:"finPrueba"`
	// Activo: hay un débito autorizado en Mercado Pago que va a cobrar en Proximo.
	Activo bool `json:"activo"`
	// Proximo: hasta cuándo está pago el plan vigente (el próximo cobro del débito, o el fin de lo
	// pagado si se canceló). nil si nunca pagó.
	Proximo   *time.Time `json:"proximo"`
	Pendiente *Cambio    `json:"pendiente"`
}

// Linea del detalle (con signo: los créditos van en negativo).
type Linea struct {
	Texto   string `json:"texto"`
	Importe int64  `json:"importe"`
}

// Cobro: una fecha y un importe.
type Cobro struct {
	Fecha   time.Time `json:"fecha"`
	Importe int64     `json:"importe"`
}

// Cotizacion de pasar a plan/ciclo.
type Cotizacion struct {
	Plan  string `json:"plan"`
	Ciclo string `json:"ciclo"`
	// Tipo: actual · prueba · alta · sube · baja · ciclo · gratis · programado · mantener.
	Tipo string `json:"tipo"`
	// Hoy: lo que se paga ahora, en pesos.
	Hoy int64 `json:"hoy"`
	// Rige: desde cuándo se tiene el plan elegido.
	Rige time.Time `json:"rige"`
	// Hasta: con un cambio que rige después, el plan vigente sigue hasta acá.
	Hasta *time.Time `json:"hasta"`
	// Cobro: el próximo cobro del débito con el plan elegido (nil si es el gratis).
	Cobro *Cobro `json:"cobro"`
	// PorCiclo: lo que se paga en cada cobro de ahí en adelante.
	PorCiclo int64 `json:"porCiclo"`
	// Lineas: el detalle de Hoy (suman Hoy).
	Lineas []Linea `json:"lineas"`
	// Pago: cómo se aplica. ninguno · suscripcion (débito nuevo en Mercado Pago) · diferencia (un pago
	// único por el prorrateo) · actualizar (cambia el monto del débito, sin pasar por Mercado Pago) ·
	// cancelar (baja del débito; el plan sigue hasta Hasta).
	Pago string `json:"pago"`
	// Inicio: con pago=suscripcion, desde cuándo cobra el débito nuevo (nil: según la prueba o ya).
	Inicio *time.Time `json:"inicio"`
}

// diasEntre redondea para arriba: un día empezado se cuenta entero.
func diasEntre(a, b time.Time) int64 {
	if !b.After(a) {
		return 0
	}
	return int64(math.Ceil(b.Sub(a).Hours() / 24))
}

// Siguiente cobro desde t: un mes o un año después.
func Siguiente(t time.Time, ciclo string) time.Time {
	if ciclo == "anio" {
		return t.AddDate(1, 0, 0)
	}
	return t.AddDate(0, 1, 0)
}

func anterior(t time.Time, ciclo string) time.Time {
	if ciclo == "anio" {
		return t.AddDate(-1, 0, 0)
	}
	return t.AddDate(0, -1, 0)
}

func nombreCiclo(ciclo string) string {
	if ciclo == "anio" {
		return "anual"
	}
	return "mensual"
}

// Cotizar el pase a plan/ciclo desde el estado e, a la hora ahora.
func Cotizar(e Estado, plan, ciclo string, ahora time.Time) (Cotizacion, error) {
	precio, ok := Precio(plan, ciclo)
	if !ok {
		return Cotizacion{}, ErrInvalido
	}
	c := Cotizacion{Plan: plan, Ciclo: ciclo, Rige: ahora, PorCiclo: precio, Pago: "ninguno", Lineas: []Linea{}}
	enPrueba := ahora.Before(e.FinPrueba)
	pago := e.Proximo != nil && e.Proximo.After(ahora) // tiene días pagos por delante
	actual, _ := Precio(e.Plan, e.Ciclo)
	elegido := plan + "/" + ciclo
	vigente := e.Plan + "/" + e.Ciclo

	// Con un cambio programado, sólo se puede confirmarlo (ya está) o deshacerlo (volver al vigente).
	if e.Pendiente != nil {
		switch elegido {
		case e.Pendiente.Plan + "/" + e.Pendiente.Ciclo:
			c.Tipo, c.Rige = "programado", e.Pendiente.Desde
			c.Hasta = e.Proximo
			if plan != "gratis" {
				c.Cobro = &Cobro{Fecha: e.Pendiente.Desde, Importe: precio}
			}
			return c, nil
		case vigente:
			c.Tipo = "mantener"
			if e.Plan != "gratis" && pago {
				c.Cobro = &Cobro{Fecha: *e.Proximo, Importe: precio}
				// Si el débito se dio de baja (gratis) o se cambió de ciclo, hace falta uno nuevo desde
				// el próximo cobro; si sólo bajó el monto, se vuelve a subir.
				if !e.Activo || e.Pendiente.Ciclo != e.Ciclo {
					c.Pago, c.Inicio = "suscripcion", e.Proximo
				} else {
					c.Pago = "actualizar"
				}
			}
			return c, nil
		default:
			return Cotizacion{}, ErrProgramado
		}
	}

	if elegido == vigente || (plan == "gratis" && e.Plan == "gratis") {
		c.Tipo = "actual"
		if plan != "gratis" && pago {
			c.Cobro = &Cobro{Fecha: *e.Proximo, Importe: precio}
		}
		return c, nil
	}

	if plan == "gratis" {
		c.Tipo = "gratis"
		if e.Activo && pago {
			// Lo pagado no se pierde: sigue con su plan hasta el próximo cobro y el débito se da de baja.
			c.Rige, c.Hasta, c.Pago = *e.Proximo, e.Proximo, "cancelar"
		} else if e.Activo {
			c.Pago = "cancelar"
		}
		return c, nil
	}

	linea := Nombres[plan] + " " + nombreCiclo(ciclo)
	switch {
	case enPrueba:
		// Lo que queda de la prueba no se cobra: el primer cobro es cuando termina.
		c.Tipo = "prueba"
		c.Cobro = &Cobro{Fecha: e.FinPrueba, Importe: precio}
		if e.Activo && ciclo == e.Ciclo {
			c.Pago = "actualizar" // ya hay un débito con la prueba: sólo cambia el monto
		} else {
			c.Pago = "suscripcion"
		}
	case !e.Activo && pago && e.Plan != "gratis":
		// Dio de baja el débito pero le quedan días pagos: el débito nuevo arranca cuando terminan.
		c.Tipo = "ciclo"
		if plan == e.Plan {
			c.Tipo = "mantener"
		}
		c.Rige, c.Hasta = *e.Proximo, e.Proximo
		c.Cobro = &Cobro{Fecha: *e.Proximo, Importe: precio}
		c.Pago, c.Inicio = "suscripcion", e.Proximo
	case !e.Activo || !pago:
		// Sin débito: se paga hoy el primer período.
		c.Tipo = "alta"
		c.Hoy = precio
		c.Cobro = &Cobro{Fecha: Siguiente(ahora, ciclo), Importe: precio}
		c.Lineas = append(c.Lineas, Linea{Texto: linea, Importe: precio})
		c.Pago = "suscripcion"
	case ciclo != e.Ciclo:
		// Cambio de ciclo: el actual ya está pago hasta el próximo cobro; el nuevo arranca ahí.
		c.Tipo = "ciclo"
		c.Rige, c.Hasta = *e.Proximo, e.Proximo
		c.Cobro = &Cobro{Fecha: *e.Proximo, Importe: precio}
		c.Pago, c.Inicio = "suscripcion", e.Proximo
	case precio > actual:
		// Sube: hoy paga la diferencia por los días que quedan; desde el próximo cobro, el precio nuevo.
		c.Tipo = "sube"
		dias := diasEntre(ahora, *e.Proximo)
		periodo := diasEntre(anterior(*e.Proximo, e.Ciclo), *e.Proximo)
		nuevo := int64(math.Round(float64(precio*dias) / float64(periodo)))
		credito := int64(math.Round(float64(actual*dias) / float64(periodo)))
		c.Hoy = nuevo - credito
		c.Cobro = &Cobro{Fecha: *e.Proximo, Importe: precio}
		d, esos := strconv.FormatInt(dias, 10)+" días", "esos días"
		if dias == 1 {
			d, esos = "1 día", "ese día"
		}
		c.Lineas = append(c.Lineas,
			Linea{Texto: Nombres[plan] + " por " + d, Importe: nuevo},
			Linea{Texto: Nombres[e.Plan] + " ya pagado por " + esos, Importe: -credito})
		c.Pago = "diferencia"
		if c.Hoy <= 0 {
			c.Hoy, c.Pago = 0, "actualizar"
		}
	default:
		// Baja: no se cobra nada hoy; sigue con el plan pago hasta el próximo cobro.
		c.Tipo = "baja"
		c.Rige, c.Hasta = *e.Proximo, e.Proximo
		c.Cobro = &Cobro{Fecha: *e.Proximo, Importe: precio}
		c.Pago = "actualizar"
	}
	return c, nil
}

// Todas las cotizaciones, por "plan/ciclo" (las de un cambio no permitido no van).
func Todas(e Estado, ahora time.Time) map[string]Cotizacion {
	r := map[string]Cotizacion{}
	for _, p := range []string{"gratis", "emprende", "pro", "negocio"} {
		for _, ci := range []string{"mes", "anio"} {
			if c, err := Cotizar(e, p, ci, ahora); err == nil {
				r[p+"/"+ci] = c
			}
		}
	}
	return r
}
