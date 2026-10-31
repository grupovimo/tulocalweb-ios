package mercadopago

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/suscripciones"
)

// Las suscripciones que vende cada tienda (§51): preapproval sin plan, en estado pending, con el
// token de la cuenta conectada de la tienda (nunca el de la plataforma: eso es el abono de Tu Local
// Web, suscripcion.go). El external_reference es "SUS-<id de la suscripción>": no se confunde con
// el del abono (el id de la tienda). Los avisos se concilian en el worker (mp.suscripcion y
// mp.cobro_suscripcion) trayendo el preapproval o el cobro autorizado de Mercado Pago.

const prefijoSuscripcion = "SUS-"

// frecuenciaMP: cada cuánto cobra Mercado Pago.
var frecuenciaMP = map[string][2]any{"quincenal": {14, "days"}, "mensual": {1, "months"}, "bimestral": {2, "months"}}

type preapprovalTienda struct {
	ID                string `json:"id"`
	Status            string `json:"status"`
	InitPoint         string `json:"init_point"`
	ExternalReference string `json:"external_reference"`
	NextPaymentDate   string `json:"next_payment_date"`
}

// SuscribirCliente arma el débito de una suscripción pendiente con la cuenta de la tienda y devuelve
// el init_point. El importe y la frecuencia salen de la suscripción (copiados del plan).
func (s *Servicio) SuscribirCliente(ctx context.Context, tiendaID, suscripcionID, vuelta string) (string, error) {
	var plan, frecuencia, email, tienda string
	var precio int64
	err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT s.plan_nombre, s.frecuencia, s.email, s.precio, t.nombre FROM suscripciones s
			JOIN tiendas t ON t.id = s.tienda_id WHERE s.id = $1 AND s.estado = 'pendiente'`, suscripcionID).
			Scan(&plan, &frecuencia, &email, &precio, &tienda)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNoExiste
	}
	if err != nil {
		return "", err
	}
	token, _, err := s.tokenDe(ctx, tiendaID)
	if err != nil {
		return "", err
	}
	f, ok := frecuenciaMP[frecuencia]
	if !ok {
		return "", ErrInvalido
	}
	var p preapprovalTienda
	if err := s.cliente().JSON(ctx, http.MethodPost, "/preapproval", token, "suscripcion:"+suscripcionID, map[string]any{
		"reason":             recortar(plan+" · "+tienda, 250),
		"external_reference": prefijoSuscripcion + suscripcionID,
		"payer_email":        sinEtiqueta(email),
		"auto_recurring": map[string]any{"frequency": f[0], "frequency_type": f[1], "transaction_amount": precio,
			"currency_id": "ARS"},
		"back_url": vuelta,
		"status":   "pending",
	}, &p); err != nil {
		return "", err
	}
	if p.ID == "" || p.InitPoint == "" {
		return "", errors.New("mercadopago: la suscripción no trajo init_point")
	}
	err = s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `UPDATE suscripciones SET mp_preapproval = $2, actualizada_en = now() WHERE id = $1`, suscripcionID, p.ID)
		return err
	})
	return p.InitPoint, err
}

// CambiarSuscripcionCliente pausa (paused), reanuda (authorized) o cancela (cancelled) el débito.
func (s *Servicio) CambiarSuscripcionCliente(ctx context.Context, tiendaID, suscripcionID, estado string) error {
	var id *string
	if err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT mp_preapproval FROM suscripciones WHERE id = $1`, suscripcionID).Scan(&id)
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoExiste
		}
		return err
	}
	if id == nil {
		return nil
	}
	token, _, err := s.tokenDe(ctx, tiendaID)
	if err != nil {
		return err
	}
	return s.cliente().JSON(ctx, http.MethodPut, "/preapproval/"+url.PathEscape(*id), token, "",
		map[string]string{"status": estado}, nil)
}

func fechaMP(s string) *time.Time {
	if s == "" {
		return nil
	}
	for _, f := range []string{time.RFC3339Nano, "2006-01-02T15:04:05.000-07:00", "2006-01-02T15:04:05.000Z0700"} {
		if t, err := time.Parse(f, s); err == nil {
			return &t
		}
	}
	return nil
}

// ConciliarSuscripcionCliente trae el preapproval con el token de la tienda (el aviso no es la fuente
// de verdad) y aplica su estado. false si no es una suscripción de una tienda (puede ser el abono).
func (s *Servicio) ConciliarSuscripcionCliente(ctx context.Context, tiendaID, preapprovalID string) (bool, error) {
	token, _, err := s.tokenDe(ctx, tiendaID)
	if err != nil {
		return false, err
	}
	var p preapprovalTienda
	if err := s.cliente().JSON(ctx, http.MethodGet, "/preapproval/"+url.PathEscape(preapprovalID), token, "", nil, &p); err != nil {
		var e *ErrorMP
		if errors.As(err, &e) && e.Estado == http.StatusNotFound {
			return false, nil
		}
		return false, err
	}
	id, ok := strings.CutPrefix(p.ExternalReference, prefijoSuscripcion)
	if !ok || !suscripciones.IDValido(id) {
		return false, nil
	}
	return true, s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		// Sólo si es de esta tienda (RLS) y es el débito que se armó para ella.
		var actual *string
		err := t.QueryRow(ctx, `SELECT mp_preapproval FROM suscripciones WHERE id = $1`, id).Scan(&actual)
		if errors.Is(err, pgx.ErrNoRows) || (err == nil && actual != nil && *actual != p.ID) {
			return nil
		}
		if err != nil {
			return err
		}
		if actual == nil {
			if _, err := t.Exec(ctx, `UPDATE suscripciones SET mp_preapproval = $2 WHERE id = $1`, id, p.ID); err != nil {
				return err
			}
		}
		return suscripciones.AplicarEstado(ctx, t, tiendaID, id, p.Status, fechaMP(p.NextPaymentDate))
	})
}

// cobroAutorizado es lo que interesa de GET /authorized_payments/{id}.
type cobroAutorizado struct {
	ID                int64   `json:"id"`
	PreapprovalID     string  `json:"preapproval_id"`
	Status            string  `json:"status"`
	TransactionAmount float64 `json:"transaction_amount"`
	Payment           struct {
		ID     int64  `json:"id"`
		Status string `json:"status"`
	} `json:"payment"`
}

// ConciliarCobroSuscripcion trae el cobro autorizado (con el token de la tienda) y lo aplica: aprobado,
// el pedido; rechazado, el aviso a la tienda. Idempotente.
func (s *Servicio) ConciliarCobroSuscripcion(ctx context.Context, tiendaID, autorizadoID string) error {
	token, _, err := s.tokenDe(ctx, tiendaID)
	if err != nil {
		return err
	}
	var c cobroAutorizado
	if err := s.cliente().JSON(ctx, http.MethodGet, "/authorized_payments/"+url.PathEscape(autorizadoID), token, "", nil, &c); err != nil {
		var e *ErrorMP
		if errors.As(err, &e) && e.Estado == http.StatusNotFound {
			return nil
		}
		return err
	}
	if c.PreapprovalID == "" {
		return nil
	}
	return s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		id, err := suscripciones.PorPreapproval(ctx, t, c.PreapprovalID)
		if err != nil || id == "" {
			return err
		}
		switch {
		case c.Payment.ID != 0 && (c.Payment.Status == "approved" || c.Payment.Status == "authorized"):
			_, err := suscripciones.RegistrarCobro(ctx, t, tiendaID, id, c.Payment.ID, int64(c.TransactionAmount+0.5))
			return err
		case c.Payment.Status == "rejected" || c.Status == "cancelled":
			return suscripciones.RegistrarRechazo(ctx, t, tiendaID, id, strconv.FormatInt(c.ID, 10), int64(c.TransactionAmount+0.5))
		}
		return nil
	})
}

// cobroDeSuscripcion: un pago (aviso «payment») cuya referencia es la de una suscripción de la tienda.
// Si está aprobado arma el pedido, igual que el cobro autorizado (el mismo pago no crea dos).
func cobroDeSuscripcion(ctx context.Context, t pgx.Tx, tiendaID, ref string, pagoID int64, estado string, pagado int64) error {
	id, ok := strings.CutPrefix(ref, prefijoSuscripcion)
	if !ok || !suscripciones.IDValido(id) || (estado != "approved" && estado != "processed") {
		return nil
	}
	_, err := suscripciones.RegistrarCobro(ctx, t, tiendaID, id, pagoID, pagado)
	return err
}
