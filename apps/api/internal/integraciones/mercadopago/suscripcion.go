package mercadopago

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/abono"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

// El abono de Tu Local Web (§56): un débito de Mercado Pago (suscripción sin plan, en estado pending,
// como la guía de Subscriptions) más, al subir de plan a mitad de período, un pago único por la
// diferencia prorrateada. Qué se cobra y cuándo sale de abono.Cotizar, la misma cuenta que muestran la
// pantalla y la app. Tu Local Web nunca ve la tarjeta: la persona la carga en Mercado Pago.

// DiasDePrueba de la prueba gratis.
const DiasDePrueba = abono.DiasDePrueba

// PrecioDe un plan pago en un ciclo: lo que se cobra cada vez (el gratis no se cobra).
func PrecioDe(plan, ciclo string) (int64, bool) {
	if plan == "gratis" {
		return 0, false
	}
	return abono.Precio(plan, ciclo)
}

type preapproval struct {
	ID                string `json:"id"`
	Status            string `json:"status"`
	InitPoint         string `json:"init_point"`
	ExternalReference string `json:"external_reference"`
	Reason            string `json:"reason"`
	NextPaymentDate   string `json:"next_payment_date"`
}

// aFechaMP: las fechas que se mandan a Mercado Pago llevan milisegundos y la zona (-04:00).
func aFechaMP(t time.Time) string { return t.Format("2006-01-02T15:04:05.000-07:00") }

func proximoDe(p preapproval) *time.Time { return fechaMP(p.NextPaymentDate) }

// CambiarAbono aplica el pase a plan/ciclo según su cotización. Devuelve la URL de Mercado Pago cuando
// hay que pasar por ahí (cargar la tarjeta del débito o pagar la diferencia); "" si quedó hecho.
func (s *Servicio) CambiarAbono(ctx context.Context, tiendaID, usuarioID, plan, ciclo string) (string, error) {
	var e abono.Estado
	var b abono.Beneficio
	var sus *string
	if err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var err error
		if e, err = abono.Leer(ctx, t, tiendaID); err != nil {
			return err
		}
		if b, _, err = abono.LeerBeneficio(ctx, t, tiendaID); err != nil {
			return err
		}
		return t.QueryRow(ctx, `SELECT suscripcion_mp FROM tiendas WHERE id = $1`, tiendaID).Scan(&sus)
	}); err != nil {
		return "", err
	}
	c, err := abono.Cotizar(e, plan, ciclo, s.ahora())
	if err != nil {
		return "", err
	}
	// Sin credenciales de la plataforma (desarrollo): el plan cambia sin cobrar.
	if s.Config.TokenPlataforma == "" {
		return "", s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
			if err := abono.Programar(ctx, t, tiendaID, nil); err != nil {
				return err
			}
			return abono.Aplicar(ctx, t, tiendaID, plan, ciclo, "cambió al plan "+abono.Nombres[plan])
		})
	}
	switch c.Pago {
	case "suscripcion":
		return s.suscribir(ctx, tiendaID, usuarioID, plan, ciclo, c.Inicio)
	case "diferencia":
		return s.pagarDiferencia(ctx, tiendaID, c)
	case "actualizar":
		if sus == nil {
			return s.suscribir(ctx, tiendaID, usuarioID, plan, ciclo, c.Inicio)
		}
		if err := s.cambiarMonto(ctx, *sus, abono.ConDescuento(c.PorCiclo, b), plan, ciclo); err != nil {
			return "", err
		}
		return "", s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
			switch c.Tipo {
			case "baja":
				return abono.Programar(ctx, t, tiendaID, &abono.Cambio{Plan: plan, Ciclo: ciclo, Desde: c.Rige})
			case "mantener":
				return abono.Programar(ctx, t, tiendaID, nil)
			}
			return abono.Aplicar(ctx, t, tiendaID, plan, ciclo, "cambió al plan "+abono.Nombres[plan])
		})
	case "cancelar":
		if sus != nil {
			if err := s.cambiarSuscripcion(ctx, *sus, "cancelled"); err != nil {
				return "", err
			}
		}
		return "", s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
			if _, err := t.Exec(ctx, `UPDATE tiendas SET suscripcion_estado = 'cancelled' WHERE id = $1`, tiendaID); err != nil {
				return err
			}
			if c.Hasta != nil {
				return abono.Programar(ctx, t, tiendaID, &abono.Cambio{Plan: "gratis", Ciclo: e.Ciclo, Desde: *c.Hasta})
			}
			return abono.Aplicar(ctx, t, tiendaID, "gratis", e.Ciclo, "pasó al plan Gratis")
		})
	}
	// ninguno: el plan vigente, el cambio ya programado, o pasar a gratis sin débito.
	if c.Tipo == "gratis" {
		return "", s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
			return abono.Aplicar(ctx, t, tiendaID, "gratis", e.Ciclo, "pasó al plan Gratis")
		})
	}
	return "", nil
}

// Suscribir crea el débito pendiente del plan y devuelve el init_point (lo que queda de la prueba no se
// cobra). Se mantiene para la app de antes de §56.
func (s *Servicio) Suscribir(ctx context.Context, tiendaID, usuarioID, plan, ciclo string) (string, error) {
	return s.suscribir(ctx, tiendaID, usuarioID, plan, ciclo, nil)
}

// suscribir: el débito pendiente. Con inicio, arranca a cobrar recién ahí (lo anterior ya está pago);
// sin inicio, lo que queda de la prueba gratis no se cobra. El plan cambia cuando Mercado Pago avisa que
// quedó autorizado.
func (s *Servicio) suscribir(ctx context.Context, tiendaID, usuarioID, plan, ciclo string, inicio *time.Time) (string, error) {
	if s.Config.TokenPlataforma == "" {
		return "", ErrApagado
	}
	importe, ok := PrecioDe(plan, ciclo)
	if !ok {
		return "", ErrPlanInvalido
	}
	var email string
	var creada time.Time
	if err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		if err := t.QueryRow(ctx, `SELECT creada_en FROM tiendas WHERE id = $1`, tiendaID).Scan(&creada); err != nil {
			return err
		}
		b, _, err := abono.LeerBeneficio(ctx, t, tiendaID)
		if err != nil {
			return err
		}
		importe = abono.ConDescuento(importe, b)
		return t.QueryRow(ctx, `SELECT email FROM auth.usuarios WHERE id = $1`, usuarioID).Scan(&email)
	}); err != nil {
		return "", err
	}
	frecuencia := 1
	if ciclo == "anio" {
		frecuencia = 12
	}
	recurrencia := map[string]any{
		"frequency": frecuencia, "frequency_type": "months", "transaction_amount": importe, "currency_id": "ARS",
	}
	if inicio != nil {
		recurrencia["start_date"] = aFechaMP(*inicio)
	} else if dias := DiasDePrueba - int(s.ahora().Sub(creada).Hours()/24); dias > 0 {
		recurrencia["free_trial"] = map[string]any{"frequency": dias, "frequency_type": "days"}
	}
	nombreCiclo := "mensual"
	if ciclo == "anio" {
		nombreCiclo = "anual"
	}
	var p preapproval
	if err := s.cliente().JSON(ctx, http.MethodPost, "/preapproval", s.Config.TokenPlataforma, "", map[string]any{
		"reason":             "Tu Local Web " + abono.Nombres[plan] + " (" + nombreCiclo + ")",
		"external_reference": tiendaID,
		"payer_email":        sinEtiqueta(email),
		"auto_recurring":     recurrencia,
		"back_url":           s.Config.Sitio + "/panel/plan-y-facturacion?suscripcion=vuelta",
		"status":             "pending",
	}, &p); err != nil {
		return "", err
	}
	if p.ID == "" || p.InitPoint == "" {
		return "", errors.New("mercadopago: la suscripción no trajo init_point")
	}
	err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `UPDATE tiendas SET plan_pendiente = $2, ciclo_pendiente = $3, espera_desde = $5,
			suscripcion_estado = CASE WHEN suscripcion_estado = 'authorized' THEN suscripcion_estado ELSE 'pending' END,
			suscripcion_mp = CASE WHEN suscripcion_estado = 'authorized' THEN suscripcion_mp ELSE $4 END
			WHERE id = $1`, tiendaID, plan, ciclo, p.ID, inicio)
		return err
	})
	return p.InitPoint, err
}

// cambiarMonto del débito vigente: rige desde el próximo cobro, sin pasar por Mercado Pago.
func (s *Servicio) cambiarMonto(ctx context.Context, id string, importe int64, plan, ciclo string) error {
	nombreCiclo := "mensual"
	if ciclo == "anio" {
		nombreCiclo = "anual"
	}
	return s.cliente().JSON(ctx, http.MethodPut, "/preapproval/"+url.PathEscape(id), s.Config.TokenPlataforma, "", map[string]any{
		"reason":         "Tu Local Web " + abono.Nombres[plan] + " (" + nombreCiclo + ")",
		"auto_recurring": map[string]any{"transaction_amount": importe, "currency_id": "ARS"},
	}, nil)
}

// pagarDiferencia: la preferencia de Checkout Pro (con la cuenta de la plataforma) por la diferencia
// prorrateada. La referencia lleva tienda, plan, ciclo e importe: el aviso del pago aplica el plan sólo
// si se pagó eso.
func (s *Servicio) pagarDiferencia(ctx context.Context, tiendaID string, c abono.Cotizacion) (string, error) {
	vuelta := s.Config.Sitio + "/panel/plan-y-facturacion"
	p := preferencia{
		Items: []item{{Title: "Tu Local Web: pasar a " + abono.Nombres[c.Plan] + " (diferencia prorrateada)", Quantity: 1,
			UnitPrice: c.Hoy, CurrencyID: "ARS"}},
		ExternalReference: fmt.Sprintf("abono:%s:%s:%s:%d", tiendaID, c.Plan, c.Ciclo, c.Hoy),
		BackURLs:          map[string]string{"success": vuelta + "?pago=ok", "pending": vuelta + "?pago=pendiente", "failure": vuelta + "?pago=error"},
		StatementDesc:     "TU LOCAL WEB",
	}
	if s.publico() {
		p.AutoReturn = "approved"
	}
	var r struct {
		InitPoint string `json:"init_point"`
	}
	if err := s.cliente().JSON(ctx, http.MethodPost, "/checkout/preferences", s.Config.TokenPlataforma, "", p, &r); err != nil {
		return "", err
	}
	if r.InitPoint == "" {
		return "", errors.New("mercadopago: la preferencia no trajo init_point")
	}
	return r.InitPoint, nil
}

var esUUID = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// ConciliarSuscripcion trae el débito (con el token de la plataforma) y lo aplica a su tienda.
func (s *Servicio) ConciliarSuscripcion(ctx context.Context, id string) error {
	_, err := s.conciliarSuscripcion(ctx, id, "")
	return err
}

// conciliarSuscripcion: con deTienda, sólo si el débito es de esa tienda (la vuelta desde el panel).
func (s *Servicio) conciliarSuscripcion(ctx context.Context, id, deTienda string) (string, error) {
	if s.Config.TokenPlataforma == "" {
		return "", ErrApagado
	}
	var p preapproval
	if err := s.cliente().JSON(ctx, http.MethodGet, "/preapproval/"+url.PathEscape(id), s.Config.TokenPlataforma, "", nil, &p); err != nil {
		return "", err
	}
	if !esUUID.MatchString(p.ExternalReference) {
		return "", nil // no la creó Tu Local Web
	}
	tiendaID := p.ExternalReference
	if deTienda != "" && deTienda != tiendaID {
		return "", ErrNoExiste
	}
	ahora := s.ahora()
	var otras []string
	err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var actual, estado, pendiente, cicloPend *string
		var espera, proximo *time.Time
		var plan, ciclo string
		var programado *string
		err := t.QueryRow(ctx, `SELECT suscripcion_mp, suscripcion_estado, plan_pendiente, ciclo_pendiente, espera_desde,
				abono_proximo, plan, plan_ciclo, plan_programado
			FROM tiendas WHERE id = $1 FOR UPDATE`, tiendaID).
			Scan(&actual, &estado, &pendiente, &cicloPend, &espera, &proximo, &plan, &ciclo, &programado)
		if err != nil {
			return err
		}
		switch p.Status {
		case "authorized":
			if pendiente == nil || cicloPend == nil {
				// Un aviso del débito vigente (por ejemplo, después de un cobro): el próximo cobro al día.
				if actual != nil && *actual == p.ID {
					if prox := proximoDe(p); prox != nil {
						_, err = t.Exec(ctx, `UPDATE tiendas SET abono_proximo = $2 WHERE id = $1`, tiendaID, *prox)
					}
				}
				return err
			}
			if actual != nil && *actual != p.ID {
				otras = append(otras, *actual)
			}
			prox := proximoDe(p)
			if espera != nil && espera.After(ahora) {
				prox = espera
			}
			if _, err := t.Exec(ctx, `UPDATE tiendas SET suscripcion_mp = $2, suscripcion_estado = 'authorized', plan_pendiente = NULL,
					ciclo_pendiente = NULL, espera_desde = NULL, abono_proximo = coalesce($3, abono_proximo) WHERE id = $1`,
				tiendaID, p.ID, prox); err != nil {
				return err
			}
			if espera != nil && espera.After(ahora) {
				// Arranca a cobrar en espera: hasta ahí sigue lo ya pagado. Si es el mismo plan (volvió atrás
				// una baja), no queda nada programado.
				if *pendiente == plan && *cicloPend == ciclo {
					return abono.Programar(ctx, t, tiendaID, nil)
				}
				return abono.Programar(ctx, t, tiendaID, &abono.Cambio{Plan: *pendiente, Ciclo: *cicloPend, Desde: *espera})
			}
			if err := abono.Programar(ctx, t, tiendaID, nil); err != nil {
				return err
			}
			return abono.Aplicar(ctx, t, tiendaID, *pendiente, *cicloPend, "empezó el plan "+abono.Nombres[*pendiente]+" (débito de Mercado Pago)")
		case "cancelled", "paused":
			// Sólo cuenta si es el vigente: un débito pendiente viejo que se cancela no cambia nada.
			if actual == nil || *actual != p.ID {
				return nil
			}
			if _, err := t.Exec(ctx, `UPDATE tiendas SET suscripcion_estado = $2 WHERE id = $1`, tiendaID, p.Status); err != nil {
				return err
			}
			if p.Status == "paused" {
				if estado != nil && *estado == "paused" {
					return nil
				}
				// Mercado Pago pausó el cobro del abono vigente: la campanita de la dueña (§45).
				return outbox.Publicar(ctx, t, tiendaID, "plan.pausado", map[string]string{})
			}
			// Dado de baja: lo pagado no se pierde. Si ya hay una baja programada, queda esa.
			if programado != nil {
				return nil
			}
			if proximo != nil && proximo.After(ahora) && plan != "gratis" {
				return abono.Programar(ctx, t, tiendaID, &abono.Cambio{Plan: "gratis", Ciclo: ciclo, Desde: *proximo})
			}
			return abono.Aplicar(ctx, t, tiendaID, "gratis", ciclo, "pasó al plan Gratis (se dio de baja el débito)")
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	// El anterior queda cancelado para no cobrar dos abonos.
	for _, o := range otras {
		if err := s.cambiarSuscripcion(ctx, o, "cancelled"); err != nil {
			s.log().Error("mercadopago: no se pudo cancelar la suscripción anterior", "tienda", tiendaID, "error", err)
		}
	}
	return tiendaID, nil
}

// ConciliarCobroAbono: cada cobro del débito del abono (subscription_authorized_payment de la cuenta de
// la plataforma). Lo aprobado va a Facturas y el próximo cobro se pone al día.
func (s *Servicio) ConciliarCobroAbono(ctx context.Context, id string) error {
	if s.Config.TokenPlataforma == "" {
		return ErrApagado
	}
	var a struct {
		ID                json.Number `json:"id"`
		PreapprovalID     string      `json:"preapproval_id"`
		Status            string      `json:"status"`
		TransactionAmount float64     `json:"transaction_amount"`
		Payment           struct {
			Status string `json:"status"`
		} `json:"payment"`
	}
	if err := s.cliente().JSON(ctx, http.MethodGet, "/authorized_payments/"+url.PathEscape(id), s.Config.TokenPlataforma, "", nil, &a); err != nil {
		return err
	}
	if a.PreapprovalID == "" || a.Payment.Status != "approved" {
		return nil
	}
	var p preapproval
	if err := s.cliente().JSON(ctx, http.MethodGet, "/preapproval/"+url.PathEscape(a.PreapprovalID), s.Config.TokenPlataforma, "", nil, &p); err != nil {
		return err
	}
	if !esUUID.MatchString(p.ExternalReference) {
		return nil
	}
	// vuelve: era el último cobro con el beneficio (§57); el débito vuelve al precio del plan.
	var vuelve bool
	err := s.Pool.EnTienda(ctx, p.ExternalReference, func(t pgx.Tx) error {
		detalle := p.Reason
		if detalle == "" {
			detalle = "Abono de Tu Local Web"
		}
		tag, err := t.Exec(ctx, `INSERT INTO abono_cobros (tienda_id, clave, detalle, importe) VALUES ($1, $2, $3, $4)
			ON CONFLICT DO NOTHING`, p.ExternalReference, "cobro:"+id, recortar(detalle, 200), int64(a.TransactionAmount))
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 1 {
			if vuelve, err = abono.GastarCobro(ctx, t, p.ExternalReference); err != nil {
				return err
			}
		}
		if prox := proximoDe(p); prox != nil {
			_, err := t.Exec(ctx, `UPDATE tiendas SET abono_proximo = $2 WHERE id = $1 AND suscripcion_mp = $3`,
				p.ExternalReference, *prox, p.ID)
			return err
		}
		return nil
	})
	if err != nil || !vuelve {
		return err
	}
	return s.montoDelDebito(ctx, p.ExternalReference)
}

// montoDelDebito: el débito vigente al precio que le toca (el del plan, con el beneficio si queda).
func (s *Servicio) montoDelDebito(ctx context.Context, tiendaID string) error {
	var sus, estado *string
	var plan, ciclo string
	var importe int64
	if err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		if err := t.QueryRow(ctx, `SELECT suscripcion_mp, suscripcion_estado FROM tiendas WHERE id = $1`, tiendaID).
			Scan(&sus, &estado); err != nil {
			return err
		}
		var err error
		plan, ciclo, importe, err = abono.Debito(ctx, t, tiendaID)
		return err
	}); err != nil {
		return err
	}
	if sus == nil || estado == nil || *estado != "authorized" || importe <= 0 {
		return nil
	}
	return s.cambiarMonto(ctx, *sus, importe, plan, ciclo)
}

// AplicarBeneficio (§57): la dueña se queda con el descuento. Sin débito todavía, lo toma el débito
// cuando se cree; con débito, baja el monto desde el próximo cobro. Si Mercado Pago no lo toma, el
// beneficio no queda.
func (s *Servicio) AplicarBeneficio(ctx context.Context, tiendaID string, b abono.Beneficio) error {
	if err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		return abono.Otorgar(ctx, t, tiendaID, b)
	}); err != nil {
		return err
	}
	if s.Config.TokenPlataforma == "" {
		return nil
	}
	if err := s.montoDelDebito(ctx, tiendaID); err != nil {
		_ = s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error { return abono.Revocar(ctx, t, tiendaID) })
		return err
	}
	return nil
}

// ConciliarPagoAbono: el pago de una diferencia prorrateada (payment de la cuenta de la plataforma).
// Aprobado y por el importe cotizado, aplica el plan nuevo ya y sube el monto del débito desde el
// próximo cobro.
func (s *Servicio) ConciliarPagoAbono(ctx context.Context, pagoID string) error {
	_, err := s.conciliarPagoAbono(ctx, pagoID, "")
	return err
}

func (s *Servicio) conciliarPagoAbono(ctx context.Context, pagoID, deTienda string) (string, error) {
	if s.Config.TokenPlataforma == "" {
		return "", ErrApagado
	}
	var pago struct {
		Status            string  `json:"status"`
		ExternalReference string  `json:"external_reference"`
		TransactionAmount float64 `json:"transaction_amount"`
	}
	if err := s.cliente().JSON(ctx, http.MethodGet, "/v1/payments/"+url.PathEscape(pagoID), s.Config.TokenPlataforma, "", nil, &pago); err != nil {
		return "", err
	}
	partes := strings.Split(pago.ExternalReference, ":")
	if len(partes) != 5 || partes[0] != "abono" || !esUUID.MatchString(partes[1]) {
		return "", nil // no es una diferencia del abono
	}
	tiendaID, plan, ciclo := partes[1], partes[2], partes[3]
	if deTienda != "" && deTienda != tiendaID {
		return "", ErrNoExiste
	}
	importe, err := strconv.ParseInt(partes[4], 10, 64)
	precio, ok := PrecioDe(plan, ciclo)
	if err != nil || !ok {
		return "", nil
	}
	if pago.Status != "approved" || int64(pago.TransactionAmount) < importe {
		return tiendaID, nil
	}
	var sus *string
	nuevo := false
	if err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		tag, err := t.Exec(ctx, `INSERT INTO abono_cobros (tienda_id, clave, detalle, importe) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING`,
			tiendaID, "pago:"+pagoID, "Diferencia prorrateada para pasar a "+abono.Nombres[plan], importe)
		if err != nil || tag.RowsAffected() == 0 {
			return err // ya aplicado (el aviso y la vuelta llegan los dos)
		}
		nuevo = true
		if err := t.QueryRow(ctx, `SELECT suscripcion_mp FROM tiendas WHERE id = $1`, tiendaID).Scan(&sus); err != nil {
			return err
		}
		if err := abono.Programar(ctx, t, tiendaID, nil); err != nil {
			return err
		}
		return abono.Aplicar(ctx, t, tiendaID, plan, ciclo, "pasó al plan "+abono.Nombres[plan]+" (pagó la diferencia prorrateada)")
	}); err != nil {
		return "", err
	}
	if nuevo && sus != nil {
		if err := s.cambiarMonto(ctx, *sus, precio, plan, ciclo); err != nil {
			s.log().Error("mercadopago: no se pudo subir el monto del débito", "tienda", tiendaID, "error", err)
		}
	}
	return tiendaID, nil
}

// VolverAbono: la persona vuelve de Mercado Pago (pago de la diferencia o débito autorizado) y el
// panel lo confirma sin esperar el aviso. Sólo vale para su propia tienda.
func (s *Servicio) VolverAbono(ctx context.Context, tiendaID, pago, suscripcion string) error {
	if pago != "" {
		if _, err := s.conciliarPagoAbono(ctx, pago, tiendaID); err != nil {
			return err
		}
	}
	if suscripcion != "" {
		if _, err := s.conciliarSuscripcion(ctx, suscripcion, tiendaID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Servicio) cambiarSuscripcion(ctx context.Context, id, estado string) error {
	return s.cliente().JSON(ctx, http.MethodPut, "/preapproval/"+url.PathEscape(id), s.Config.TokenPlataforma, "",
		map[string]string{"status": estado}, nil)
}

// CancelarSuscripcion: la baja del débito desde el panel (el plan sigue hasta lo pagado).
func (s *Servicio) CancelarSuscripcion(ctx context.Context, tiendaID string) error {
	if s.Config.TokenPlataforma == "" {
		return nil
	}
	var id *string
	if err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT suscripcion_mp FROM tiendas WHERE id = $1`, tiendaID).Scan(&id)
	}); err != nil || id == nil {
		return err
	}
	return s.cambiarSuscripcion(ctx, *id, "cancelled")
}

// sinEtiqueta saca la parte "+algo" del mail ("ana+tienda@x.com" → "ana@x.com"): Mercado Pago
// rechaza esos mails con "User bad request", y es la misma casilla.
func sinEtiqueta(email string) string {
	local, dominio, ok := strings.Cut(email, "@")
	if !ok {
		return email
	}
	if base, _, ok := strings.Cut(local, "+"); ok && base != "" {
		return base + "@" + dominio
	}
	return email
}
