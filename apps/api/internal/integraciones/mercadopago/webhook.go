package mercadopago

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

// Aviso es el cuerpo de una notificación de Webhooks.
type Aviso struct {
	Type   string          `json:"type"`
	Action string          `json:"action"`
	UserID json.RawMessage `json:"user_id"`
	Data   struct {
		ID json.RawMessage `json:"id"`
	} `json:"data"`
}

func crudo(r json.RawMessage) string {
	return strings.Trim(strings.TrimSpace(string(r)), `"`)
}

var alfanumerico = regexp.MustCompile(`^[A-Za-z0-9]+$`)

// FirmaValida verifica x-signature (HMAC-SHA256 de "id:{data.id};request-id:{x-request-id};ts:{ts};"
// con la clave secreta de los webhooks), en tiempo constante. dataID es el de la URL; si es
// alfanumérico va en minúsculas. Además rechaza avisos de más de 10 minutos.
func FirmaValida(secreto, firma, requestID, dataID string, ahora time.Time) bool {
	if secreto == "" || firma == "" {
		return false
	}
	var ts, v1 string
	for _, parte := range strings.Split(firma, ",") {
		k, v, _ := strings.Cut(strings.TrimSpace(parte), "=")
		switch strings.TrimSpace(k) {
		case "ts":
			ts = strings.TrimSpace(v)
		case "v1":
			v1 = strings.TrimSpace(v)
		}
	}
	if ts == "" || v1 == "" {
		return false
	}
	if alfanumerico.MatchString(dataID) {
		dataID = strings.ToLower(dataID)
	}
	var b strings.Builder
	if dataID != "" {
		b.WriteString("id:" + dataID + ";")
	}
	if requestID != "" {
		b.WriteString("request-id:" + requestID + ";")
	}
	b.WriteString("ts:" + ts + ";")
	m := hmac.New(sha256.New, []byte(secreto))
	m.Write([]byte(b.String()))
	esperada := hex.EncodeToString(m.Sum(nil))
	if !hmac.Equal([]byte(esperada), []byte(v1)) {
		return false
	}
	// ts viene en milisegundos.
	if n, err := strconv.ParseInt(ts, 10, 64); err == nil {
		cuando := time.UnixMilli(n)
		if n < 1e12 {
			cuando = time.Unix(n, 0)
		}
		if ahora.Sub(cuando).Abs() > 10*time.Minute {
			return false
		}
	}
	return true
}

// Recibir procesa un aviso ya verificado. Los pagos y las órdenes de una tienda van al outbox (el
// worker los concilia con reintentos); las suscripciones del abono se concilian en el momento: si
// algo falla, se devuelve error y Mercado Pago reintenta.
func (s *Servicio) Recibir(ctx context.Context, a Aviso) error {
	id := crudo(a.Data.ID)
	if id == "" || len(id) > 64 {
		return nil
	}
	switch a.Type {
	case "subscription_preapproval":
		// La suscripción que vende una tienda (§51): por el débito ya guardado o por la cuenta del aviso.
		tienda, err := s.tiendaDelAviso(ctx, `SELECT app.tienda_de_preapproval($1)::text`, id)
		if err != nil {
			return err
		}
		if tienda == nil {
			if tienda, err = s.tiendaDeUsuario(ctx, a.UserID); err != nil {
				return err
			}
		}
		if tienda != nil {
			return s.Pool.EnTienda(ctx, *tienda, func(t pgx.Tx) error {
				return outbox.Publicar(ctx, t, *tienda, "mp.suscripcion", map[string]string{"id": id})
			})
		}
		// El abono de Tu Local Web.
		err = s.ConciliarSuscripcion(ctx, id)
		var e *ErrorMP
		if errors.As(err, &e) && e.Estado == http.StatusNotFound {
			return nil // no es de la plataforma: nada que conciliar
		}
		return err
	case "subscription_authorized_payment":
		// Cada cobro de la suscripción de una tienda: el pedido (aprobado) o el aviso (rechazado).
		tienda, err := s.tiendaDeUsuario(ctx, a.UserID)
		if err != nil || tienda == nil {
			return err
		}
		return s.Pool.EnTienda(ctx, *tienda, func(t pgx.Tx) error {
			return outbox.Publicar(ctx, t, *tienda, "mp.cobro_suscripcion", map[string]string{"id": id})
		})
	case "payment", "order":
		tienda, err := s.tiendaDeUsuario(ctx, a.UserID)
		if err != nil {
			return err
		}
		if tienda == nil {
			// Pagos de la cuenta de la plataforma (el abono) o de una cuenta que se desconectó.
			return nil
		}
		tipo := "mp.pago"
		if a.Type == "order" {
			tipo = "mp.orden"
		}
		return s.Pool.EnTienda(ctx, *tienda, func(t pgx.Tx) error {
			return outbox.Publicar(ctx, t, *tienda, tipo, map[string]string{"id": id})
		})
	}
	return nil
}

// tiendaDeUsuario: la tienda conectada con la cuenta de Mercado Pago del aviso (nil si ninguna).
func (s *Servicio) tiendaDeUsuario(ctx context.Context, usuario json.RawMessage) (*string, error) {
	n, err := strconv.ParseInt(crudo(usuario), 10, 64)
	if err != nil {
		return nil, nil
	}
	return s.tiendaDelAviso(ctx, `SELECT app.tienda_de_mp($1)::text`, n)
}

func (s *Servicio) tiendaDelAviso(ctx context.Context, q string, arg any) (*string, error) {
	var tienda *string
	err := s.Pool.EnTx(ctx, func(t pgx.Tx) error { return t.QueryRow(ctx, q, arg).Scan(&tienda) })
	return tienda, err
}

// Manejadores del outbox para el worker.
func (s *Servicio) Manejadores() map[string]outbox.Manejador {
	leer := func(ev outbox.Evento) (string, error) {
		var d struct{ ID string }
		if err := json.Unmarshal(ev.Datos, &d); err != nil || d.ID == "" {
			return "", errors.New("mercadopago: evento sin id")
		}
		return d.ID, nil
	}
	return map[string]outbox.Manejador{
		"mp.pago": func(ctx context.Context, ev outbox.Evento) error {
			id, err := leer(ev)
			if err != nil {
				return err
			}
			n, err := strconv.ParseInt(id, 10, 64)
			if err != nil {
				return nil
			}
			return s.ConciliarPago(ctx, ev.TiendaID, n)
		},
		// La suscripción de una tienda (§51). Si el preapproval no es de una tienda, puede ser el abono
		// de Tu Local Web con una cuenta que además está conectada como tienda.
		"mp.suscripcion": func(ctx context.Context, ev outbox.Evento) error {
			id, err := leer(ev)
			if err != nil {
				return err
			}
			deTienda, err := s.ConciliarSuscripcionCliente(ctx, ev.TiendaID, id)
			if err != nil || deTienda || s.Config.TokenPlataforma == "" {
				return err
			}
			err = s.ConciliarSuscripcion(ctx, id)
			var e *ErrorMP
			if errors.As(err, &e) && e.Estado == http.StatusNotFound {
				return nil
			}
			return err
		},
		"mp.cobro_suscripcion": func(ctx context.Context, ev outbox.Evento) error {
			id, err := leer(ev)
			if err != nil {
				return err
			}
			return s.ConciliarCobroSuscripcion(ctx, ev.TiendaID, id)
		},
		"mp.orden": func(ctx context.Context, ev outbox.Evento) error {
			id, err := leer(ev)
			if err != nil {
				return err
			}
			_, err = s.conciliarOrden(ctx, ev.TiendaID, id)
			return err
		},
	}
}
