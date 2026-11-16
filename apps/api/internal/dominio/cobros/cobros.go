// Package cobros es lo que la tienda elige para cobrar que ve quien compra: las cuotas sin interés
// (cuántas y desde qué monto) y si muestra el precio sin impuestos nacionales (Ley 27.743)
// (docs/especificaciones/tienda-nivel-tiendanube.md). La regla es una sola: la usan la vidriera,
// el checkout de Mercado Pago, Configuración › Cobros y Cupones › Automáticas (el interruptor de
// las cuotas es el mismo, automaticas->>'cuotas'). La demo la replica en el sandbox.
package cobros

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/auditoria"
)

// Cuotas que se pueden ofrecer (0: ninguna) y el «desde» por defecto, en pesos (los de la migración 00086).
var CuotasValidas = []int{0, 3, 6, 9, 12}

const (
	CuotasPorDefecto = 6
	DesdePorDefecto  = 30_000
	MaxDesde         = 100_000_000
)

// IVAs: las alícuotas de un producto (21 por defecto).
var IVAs = []float64{0, 10.5, 21, 27}

const IVAPorDefecto = 21.0

// ErrInvalido: cuotas fuera de la lista o un «desde» fuera de los topes.
var ErrInvalido = errors.New("cobros: inválido")

// Impuestos: si la tienda muestra «Precio sin impuestos nacionales».
type Impuestos struct {
	Mostrar bool `json:"mostrar"`
}

// Config es components/schemas/CobrosTienda: Configuración › Cobros.
type Config struct {
	// Cuotas sin interés que ofrece: 0 (ninguna), 3, 6, 9 o 12. Lo que la tienda eligió, aunque no
	// tenga Mercado Pago (la vidriera, sin Mercado Pago, manda 0).
	Cuotas int `json:"cuotas"`
	// CuotasDesde: monto mínimo del carrito para ofrecerlas, en pesos.
	CuotasDesde int64     `json:"cuotasDesde"`
	Impuestos   Impuestos `json:"impuestos"`
}

// Validar revisa las cuotas y el «desde».
func (c Config) Validar() error {
	if !slices.Contains(CuotasValidas, c.Cuotas) || c.CuotasDesde < 0 || c.CuotasDesde > MaxDesde {
		return ErrInvalido
	}
	return nil
}

// IVAValido dice si es una de las alícuotas de IVA.
func IVAValido(iva float64) bool { return slices.Contains(IVAs, iva) }

// ResponsableInscripto: la condición fiscal de Configuración › Negocio es texto libre
// («Responsable inscripta», «responsable_inscripto», «RI»); todo lo demás (monotributo, exento,
// vacío) no discrimina IVA.
func ResponsableInscripto(condicion string) bool {
	c := strings.ToLower(strings.TrimSpace(condicion))
	return strings.Contains(c, "responsable") || strings.Contains(c, "inscript") || c == "ri"
}

// Leer la configuración de la tienda, dentro de una transacción con la tienda fijada. Sin elegir,
// el precio sin impuestos se muestra si es responsable inscripta (en Negocio o en ARCA).
func Leer(ctx context.Context, t pgx.Tx, tiendaID string) (Config, error) {
	var c Config
	var activa bool
	var cuotas int
	var mostrar *bool
	var condicion string
	var ri bool
	err := t.QueryRow(ctx, `SELECT coalesce((automaticas->>'cuotas')::boolean, false), cuotas_sin_interes, cuotas_desde,
			impuestos_mostrar, coalesce(negocio->>'condicion', ''),
			EXISTS (SELECT 1 FROM arca_config a WHERE a.tienda_id = $1 AND a.condicion = 'responsable_inscripto')
		FROM tiendas WHERE id = $1`, tiendaID).Scan(&activa, &cuotas, &c.CuotasDesde, &mostrar, &condicion, &ri)
	if err != nil {
		return c, err
	}
	if activa {
		c.Cuotas = cuotas
	}
	c.Impuestos.Mostrar = ri || ResponsableInscripto(condicion)
	if mostrar != nil {
		c.Impuestos.Mostrar = *mostrar
	}
	return c, nil
}

// Detalle de las cuotas en palabras, para Cupones › Automáticas: «6 cuotas sin interés desde $ 30.000».
func (c Config) Detalle(cuotas int) string {
	if c.CuotasDesde == 0 {
		return fmt.Sprintf("%d cuotas sin interés en cualquier compra", cuotas)
	}
	return fmt.Sprintf("%d cuotas sin interés desde %s", cuotas, auditoria.Pesos(c.CuotasDesde))
}

// Elegidas: las cuotas guardadas aunque el interruptor esté apagado (las que vuelven al prenderlo).
func Elegidas(ctx context.Context, t pgx.Tx, tiendaID string) (int, error) {
	var n int
	err := t.QueryRow(ctx, `SELECT cuotas_sin_interes FROM tiendas WHERE id = $1`, tiendaID).Scan(&n)
	return n, err
}

// Guardar valida y guarda, con su renglón en la Actividad. Cuotas 0 apaga el interruptor y deja
// las elegidas antes (prenderlo desde Cupones vuelve a ellas).
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
	if _, err := t.Exec(ctx, `UPDATE tiendas SET automaticas = automaticas || jsonb_build_object('cuotas', $2::boolean),
			cuotas_sin_interes = CASE WHEN $2 THEN $3 ELSE cuotas_sin_interes END, cuotas_desde = $4, impuestos_mostrar = $5
		WHERE id = $1`, tiendaID, c.Cuotas > 0, c.Cuotas, c.CuotasDesde, c.Impuestos.Mostrar); err != nil {
		return err
	}
	var regs []auditoria.Registro
	reg := func(accion, texto string) {
		regs = append(regs, auditoria.Registro{Accion: accion, Entidad: "tienda", EntidadID: tiendaID, Texto: texto,
			Antes: antes, Despues: c})
	}
	switch {
	case antes.Cuotas > 0 && c.Cuotas == 0:
		reg("cobros.cuotas", "dejó de ofrecer cuotas sin interés")
	case c.Cuotas > 0 && (antes.Cuotas != c.Cuotas || antes.CuotasDesde != c.CuotasDesde):
		reg("cobros.cuotas", "ofrece "+c.Detalle(c.Cuotas))
	}
	if antes.Impuestos.Mostrar != c.Impuestos.Mostrar {
		if c.Impuestos.Mostrar {
			reg("cobros.impuestos", "muestra el precio sin impuestos nacionales")
		} else {
			reg("cobros.impuestos", "dejó de mostrar el precio sin impuestos nacionales")
		}
	}
	return auditoria.Registrar(ctx, t, tiendaID, regs...)
}
