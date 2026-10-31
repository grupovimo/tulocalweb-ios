// Package suscripciones: los planes con cobro recurrente de la tienda (el club de vinos, el café
// cada dos semanas) y quienes se suscriben desde la tienda online (docs/decisiones.md §51).
//
//   - La tienda arma los planes en el panel; quien compra se suscribe desde la tienda y autoriza el
//     débito en Mercado Pago (preapproval con la cuenta conectada de la tienda). Sin Mercado Pago
//     conectado, la tienda no ofrece planes.
//   - El precio sale siempre del plan en la base, nunca del navegador; la suscripción lo copia.
//   - Cada cobro aprobado crea un solo pedido de la tienda online (RegistrarCobro, idempotente por
//     el id del pago de Mercado Pago) y entra al circuito normal: preparar, enviar, facturar, avisar.
//   - El paquete mercadopago habla con Mercado Pago (Cobrador) y concilia los avisos con las
//     funciones de cobros.go; este paquete no conoce Mercado Pago.
package suscripciones

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"regexp"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
)

var (
	ErrNoExiste    = errors.New("suscripciones: no existe")
	ErrInvalido    = errors.New("suscripciones: datos inválidos")
	ErrSinCobros   = errors.New("suscripciones: la tienda no conectó Mercado Pago")
	ErrCupo        = errors.New("suscripciones: el plan no tiene más cupo")
	ErrEnUso       = errors.New("suscripciones: el plan tiene suscriptores")
	ErrEstado      = errors.New("suscripciones: no se puede en este estado")
	ErrMercadoPago = errors.New("suscripciones: Mercado Pago no respondió")
)

type tx interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
	EnTx(context.Context, func(pgx.Tx) error) error
}

// Cobrador habla con Mercado Pago (el paquete mercadopago). estado: paused, authorized o cancelled.
type Cobrador interface {
	SuscribirCliente(ctx context.Context, tiendaID, suscripcionID, vuelta string) (string, error)
	CambiarSuscripcionCliente(ctx context.Context, tiendaID, suscripcionID, estado string) error
}

// Frecuencias de cobro: lo que la API de suscripciones de Mercado Pago acepta sin vueltas
// (cada 14 días, cada mes, cada 2 meses).
var Frecuencias = map[string]bool{"quincenal": true, "mensual": true, "bimestral": true}

// Estados de una suscripción.
const (
	Pendiente = "pendiente"
	Activa    = "activa"
	Pausada   = "pausada"
	Cancelada = "cancelada"
)

// ProductoPlan es un producto del catálogo que va en cada entrega.
type ProductoPlan struct {
	ProductoID string `json:"productoId"`
	VarianteID string `json:"varianteId"`
	Cantidad   int    `json:"cantidad"`
}

// Plan es components/schemas/PlanSuscripcion (el panel).
type Plan struct {
	ID          string         `json:"id"`
	Nombre      string         `json:"nombre"`
	Descripcion string         `json:"descripcion"`
	Foto        *string        `json:"foto"`
	Precio      int64          `json:"precio"`
	Frecuencia  string         `json:"frecuencia"`
	Productos   []ProductoPlan `json:"productos"`
	Incluye     string         `json:"incluye"`
	Envio       bool           `json:"envio"`
	Retiro      bool           `json:"retiro"`
	Cupo        *int           `json:"cupo"`
	Activo      bool           `json:"activo"`
	Orden       int            `json:"orden"`
	// Suscriptores vigentes (activas y pausadas).
	Suscriptores int `json:"suscriptores"`
}

// Incluido es un producto del plan como lo ve quien compra.
type Incluido struct {
	ID       string  `json:"id"`
	Nombre   string  `json:"nombre"`
	Foto     *string `json:"foto"`
	Cantidad int     `json:"cantidad"`
	Variante *string `json:"variante"`
}

// PlanPublico es TiendaPublica.suscripciones[]: sólo los activos.
type PlanPublico struct {
	ID          string     `json:"id"`
	Nombre      string     `json:"nombre"`
	Descripcion string     `json:"descripcion"`
	Foto        *string    `json:"foto"`
	Precio      int64      `json:"precio"`
	Frecuencia  string     `json:"frecuencia"`
	Incluye     string     `json:"incluye"`
	Productos   []Incluido `json:"productos"`
	Envio       bool       `json:"envio"`
	Retiro      bool       `json:"retiro"`
	// Agotado: tiene cupo y está lleno.
	Agotado bool `json:"agotado"`
}

// Suscriptor es components/schemas/Suscriptor (el panel).
type Suscriptor struct {
	ID           string     `json:"id"`
	PlanID       string     `json:"planId"`
	Plan         string     `json:"plan"`
	Precio       int64      `json:"precio"`
	Frecuencia   string     `json:"frecuencia"`
	Nombre       string     `json:"nombre"`
	Email        string     `json:"email"`
	Telefono     string     `json:"telefono"`
	Entrega      string     `json:"entrega"`
	Direccion    string     `json:"direccion"`
	Estado       string     `json:"estado"`
	ProximoCobro *time.Time `json:"proximoCobro"`
	Cobros       int        `json:"cobros"`
	UltimoCobro  *time.Time `json:"ultimoCobro"`
	CreadaEn     time.Time  `json:"creadaEn"`
	CanceladaPor *string    `json:"canceladaPor"`
}

// Panel es lo que ve la tienda en Suscripciones.
type Panel struct {
	// MercadoPago: la tienda tiene Mercado Pago conectado (sin eso, no puede ofrecer planes).
	MercadoPago  bool         `json:"mercadoPago"`
	Planes       []Plan       `json:"planes"`
	Suscriptores []Suscriptor `json:"suscriptores"`
}

// Alta es lo que manda el checkout del plan (el precio NO viene: sale del plan).
type Alta struct {
	PlanID    string `json:"planId"`
	Email     string `json:"email"`
	Nombre    string `json:"nombre"`
	Telefono  string `json:"telefono"`
	Entrega   string `json:"entrega"`
	Direccion string `json:"direccion"`
	CP        string `json:"cp"`
}

// Resultado del alta: el link privado y adónde autorizar el débito.
type Resultado struct {
	Codigo string `json:"codigo"`
	URL    string `json:"url,omitempty"`
}

// Publica es lo que ve quien se suscribió en su link privado (components/schemas/SuscripcionPublica).
type Publica struct {
	Codigo       string     `json:"codigo"`
	PlanID       string     `json:"planId"`
	Plan         string     `json:"plan"`
	Foto         *string    `json:"foto"`
	Incluye      string     `json:"incluye"`
	Precio       int64      `json:"precio"`
	Frecuencia   string     `json:"frecuencia"`
	Nombre       string     `json:"nombre"`
	Entrega      string     `json:"entrega"`
	Direccion    string     `json:"direccion"`
	Estado       string     `json:"estado"`
	ProximoCobro *time.Time `json:"proximoCobro"`
	Cobros       int        `json:"cobros"`
	CreadaEn     time.Time  `json:"creadaEn"`
	Tienda       string     `json:"tienda"`
	Whatsapp     string     `json:"whatsapp"`
}

var (
	idValido     = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	emailValido  = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	codigoValido = regexp.MustCompile(`^[A-Za-z0-9_-]{16,64}$`)
	slugValido   = regexp.MustCompile(`^[a-z0-9-]{2,60}$`)
	telValido    = regexp.MustCompile(`^[\d\s+()-]{8,20}$`)
)

// IDValido: un uuid en minúsculas.
func IDValido(s string) bool { return idValido.MatchString(s) }

func largo(s string) int { return utf8.RuneCountInString(s) }

func nuevoCodigo() string {
	b := make([]byte, 18)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func tiendaDe(ctx context.Context, pool tx, slug string) (string, error) {
	if !slugValido.MatchString(slug) {
		return "", ErrNoExiste
	}
	var id *string
	if err := pool.EnTx(ctx, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT app.tienda_de_slug($1)::text`, slug).Scan(&id)
	}); err != nil {
		return "", err
	}
	if id == nil {
		return "", ErrNoExiste
	}
	return *id, nil
}

func conMercadoPago(ctx context.Context, t pgx.Tx, tienda string) (bool, error) {
	var mp bool
	err := t.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM mp_conexiones WHERE tienda_id = $1 AND NOT reautorizar)`, tienda).Scan(&mp)
	return mp, err
}
