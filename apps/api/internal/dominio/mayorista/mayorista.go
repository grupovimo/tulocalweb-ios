// Package mayorista es la venta por mayor de la tienda online (docs/decisiones.md §49): cada producto
// (y cada variante) puede tener un precio mayorista, y cuando el carrito llega al mínimo que puso la
// tienda (unidades y/o monto, contra los precios minoristas), todos los que lo tienen pasan a ese
// precio. La regla es una sola: la usan la vidriera (que la recalcula al comprar), Configuración y
// Cupones › Automáticas (el interruptor es el mismo, automaticas->>'mayorista'). La web la replica en
// lib/tienda/mayorista.ts, con los mismos casos.
package mayorista

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/auditoria"
)

// Config es components/schemas/VentaMayorista: Configuración › Venta mayorista.
type Config struct {
	Activa bool `json:"activa"`
	// Unidades: mínimo de unidades en el carrito (nil: sin mínimo de unidades).
	Unidades *int `json:"unidades"`
	// Monto: mínimo del carrito en pesos, a precio minorista (nil: sin mínimo de monto).
	Monto *int64 `json:"monto"`
}

// Topes (los mismos de la migración 00066 y de la web).
const (
	MinUnidades = 2
	MaxUnidades = 10_000
	MaxMonto    = 100_000_000
)

// ErrSinMinimo: no cargó ni unidades ni monto.
var ErrSinMinimo = errors.New("mayorista: poné un mínimo de unidades o de monto")

// ErrInvalido: un mínimo fuera de los topes.
var ErrInvalido = errors.New("mayorista: mínimo inválido")

// Validar revisa los mínimos (al menos uno, dentro de los topes).
func (c Config) Validar() error {
	if c.Unidades == nil && c.Monto == nil {
		return ErrSinMinimo
	}
	if (c.Unidades != nil && (*c.Unidades < MinUnidades || *c.Unidades > MaxUnidades)) ||
		(c.Monto != nil && (*c.Monto < 1 || *c.Monto > MaxMonto)) {
		return ErrInvalido
	}
	return nil
}

// Alcanza dice si un carrito (sus unidades y su monto a precio minorista) llega al mínimo. Con los
// dos mínimos cargados, tienen que cumplirse los dos.
func (c Config) Alcanza(unidades int, monto int64) bool {
	if !c.Activa || (c.Unidades == nil && c.Monto == nil) {
		return false
	}
	return (c.Unidades == nil || unidades >= *c.Unidades) && (c.Monto == nil || monto >= *c.Monto)
}

// Precio de una línea con el carrito ya mayorista: el mayorista si tiene y es menor; si no, el de
// siempre (un precio mayorista nunca sube lo que se cobra).
func Precio(minorista int64, mayor *int64) int64 {
	if mayor != nil && *mayor > 0 && *mayor < minorista {
		return *mayor
	}
	return minorista
}

// DeVariante: el precio mayorista de una variante es el suyo o, si no tiene, el del producto, pero
// sólo si la variante se vende al precio del producto (una con precio propio, 1 kg contra 250 g, no
// puede heredar el mayorista de otra presentación).
func DeVariante(variante, producto, precioVariante *int64) *int64 {
	if variante != nil {
		return variante
	}
	if precioVariante != nil {
		return nil
	}
	return producto
}

// Detalle es el mínimo en palabras: «Desde 12 unidades», «Desde $ 50.000» o los dos.
func (c Config) Detalle() string {
	switch {
	case c.Unidades != nil && c.Monto != nil:
		return fmt.Sprintf("Desde %d unidades y %s", *c.Unidades, auditoria.Pesos(*c.Monto))
	case c.Unidades != nil:
		return fmt.Sprintf("Desde %d unidades", *c.Unidades)
	case c.Monto != nil:
		return "Desde " + auditoria.Pesos(*c.Monto)
	}
	return ""
}

// Leer la configuración de la tienda, dentro de una transacción con la tienda fijada.
func Leer(ctx context.Context, t pgx.Tx, tiendaID string) (Config, error) {
	var c Config
	err := t.QueryRow(ctx, `SELECT coalesce((automaticas->>'mayorista')::boolean, false), mayorista_unidades, mayorista_monto
		FROM tiendas WHERE id = $1`, tiendaID).Scan(&c.Activa, &c.Unidades, &c.Monto)
	return c, err
}

// Guardar valida y guarda (prendida o apagada y los mínimos), con su renglón en la Actividad.
func Guardar(ctx context.Context, t pgx.Tx, tiendaID string, c Config) error {
	if err := c.Validar(); err != nil {
		return err
	}
	if _, err := t.Exec(ctx, `SELECT 1 FROM tiendas WHERE id = $1 FOR UPDATE`, tiendaID); err != nil {
		return err
	}
	antes, err := Leer(ctx, t, tiendaID)
	if err != nil {
		return err
	}
	if _, err := t.Exec(ctx, `UPDATE tiendas SET automaticas = automaticas || jsonb_build_object('mayorista', $2::boolean),
			mayorista_unidades = $3, mayorista_monto = $4 WHERE id = $1`, tiendaID, c.Activa, c.Unidades, c.Monto); err != nil {
		return err
	}
	var texto string
	switch {
	case antes.Activa && !c.Activa:
		texto = "apagó la venta mayorista"
	case !antes.Activa && c.Activa:
		texto = "prendió la venta mayorista (" + minuscula(c.Detalle()) + ")"
	case antes.Detalle() != c.Detalle():
		texto = "cambió el mínimo mayorista: " + minuscula(c.Detalle())
	default:
		return nil
	}
	return auditoria.Registrar(ctx, t, tiendaID, auditoria.Registro{Accion: "mayorista.configurar", Entidad: "tienda",
		Texto: texto, Antes: antes, Despues: c})
}

func minuscula(s string) string {
	if s == "" {
		return s
	}
	return "d" + s[1:]
}
