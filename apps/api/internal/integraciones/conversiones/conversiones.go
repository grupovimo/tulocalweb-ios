// Package conversiones manda las compras de la tienda online a la API de conversiones de Meta
// (decisiones §47), con el token que carga cada tienda (cifrado con su clave). Va además del píxel
// del navegador: mismo evento (Purchase) y mismo event_id (el número de pedido), así Meta los junta
// y la compra se cuenta aunque el navegador bloquee el píxel. Sólo si quien compra aceptó las
// cookies de medición de la tienda.
package conversiones

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/cifrado"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

// Version de la Graph API.
const Version = "v23.0"

// TipoCompra es el evento del outbox que manda una compra.
const TipoCompra = "meta.compra"

var (
	ErrInvalido     = errors.New("conversiones: datos inválidos")
	ErrSinPixel     = errors.New("conversiones: falta el píxel")
	ErrSinCifrado   = errors.New("conversiones: sin clave de cifrado")
	ErrTokenNoSirve = errors.New("conversiones: Meta no aceptó el token")
	ErrMetaCaida    = errors.New("conversiones: no se pudo hablar con Meta")
)

var (
	tokenValido  = regexp.MustCompile(`^[A-Za-z0-9_-]{20,500}$`)
	codigoValido = regexp.MustCompile(`^TEST[0-9A-Za-z]{1,20}$`)
)

// Servicio guarda la conexión de cada tienda y manda los eventos.
type Servicio struct {
	Pool  *db.Pool
	Cifra *cifrado.Cifrador
	HTTP  *http.Client
	// Base de la Graph API (en los tests, un servidor falso). Vacía: https://graph.facebook.com.
	Base string
}

func (s *Servicio) url(partes ...string) string {
	base := s.Base
	if base == "" {
		base = "https://graph.facebook.com"
	}
	return strings.TrimRight(base, "/") + "/" + Version + "/" + strings.Join(partes, "/")
}

func (s *Servicio) cli() *http.Client {
	if s.HTTP != nil {
		return s.HTTP
	}
	return &http.Client{Timeout: 15 * time.Second}
}

func contexto(tienda string) string { return "meta:" + tienda }

// Estado es components/schemas/Conversiones: nunca lleva el token.
type Estado struct {
	Conectada    bool       `json:"conectada"`
	CodigoPrueba *string    `json:"codigoPrueba"`
	UltimoEnvio  *time.Time `json:"ultimoEnvio"`
	UltimoError  *string    `json:"ultimoError"`
}

// Leer devuelve cómo está la conexión de la tienda.
func (s *Servicio) Leer(ctx context.Context, tienda string) (Estado, error) {
	var e Estado
	err := s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT codigo_prueba, ultimo_envio, ultimo_error FROM meta_conversiones WHERE tienda_id = $1`, tienda).
			Scan(&e.CodigoPrueba, &e.UltimoEnvio, &e.UltimoError)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Estado{}, nil
	}
	e.Conectada = err == nil
	return e, err
}

// Conexion es lo que carga la tienda. Token vacío con la conexión hecha: sólo cambia el código de
// prueba.
type Conexion struct {
	Token        string `json:"token"`
	CodigoPrueba string `json:"codigoPrueba"`
}

// Guardar comprueba el token con Meta (que pueda leer el píxel de la tienda) y lo guarda cifrado.
func (s *Servicio) Guardar(ctx context.Context, tienda string, c Conexion) error {
	c.Token, c.CodigoPrueba = strings.TrimSpace(c.Token), strings.ToUpper(strings.TrimSpace(c.CodigoPrueba))
	if c.CodigoPrueba != "" && !codigoValido.MatchString(c.CodigoPrueba) {
		return ErrInvalido
	}
	var codigo *string
	if c.CodigoPrueba != "" {
		codigo = &c.CodigoPrueba
	}
	if c.Token == "" {
		return s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
			tag, err := t.Exec(ctx, `UPDATE meta_conversiones SET codigo_prueba = $2, actualizada_en = now() WHERE tienda_id = $1`, tienda, codigo)
			if err == nil && tag.RowsAffected() == 0 {
				return ErrInvalido
			}
			return err
		})
	}
	if !tokenValido.MatchString(c.Token) {
		return ErrInvalido
	}
	if s.Cifra == nil {
		return ErrSinCifrado
	}
	pixel, err := s.pixel(ctx, tienda)
	if err != nil {
		return err
	}
	if err := s.comprobar(ctx, pixel, c.Token); err != nil {
		return err
	}
	cif, err := s.Cifra.Cifrar(contexto(tienda), []byte(c.Token))
	if err != nil {
		return err
	}
	return s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `INSERT INTO meta_conversiones (tienda_id, token, codigo_prueba) VALUES ($1, $2, $3)
			ON CONFLICT (tienda_id) DO UPDATE SET token = $2, codigo_prueba = $3, ultimo_error = NULL, actualizada_en = now()`,
			tienda, cif, codigo)
		return err
	})
}

// Borrar desconecta (el token se borra).
func (s *Servicio) Borrar(ctx context.Context, tienda string) error {
	return s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `DELETE FROM meta_conversiones WHERE tienda_id = $1`, tienda)
		return err
	})
}

func (s *Servicio) pixel(ctx context.Context, tienda string) (string, error) {
	var pixel *string
	err := s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT pixel_meta FROM tiendas WHERE id = $1`, tienda).Scan(&pixel)
	})
	if err != nil {
		return "", err
	}
	if pixel == nil || *pixel == "" {
		return "", ErrSinPixel
	}
	return *pixel, nil
}

// comprobar: el token tiene que poder leer el píxel (un GET que no crea eventos).
func (s *Servicio) comprobar(ctx context.Context, pixel, token string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.url(pixel)+"?fields=id", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	res, err := s.cli().Do(req)
	if err != nil {
		return ErrMetaCaida
	}
	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, 64<<10))
	switch {
	case res.StatusCode == http.StatusOK:
		return nil
	case res.StatusCode >= 500 || res.StatusCode == http.StatusTooManyRequests:
		return ErrMetaCaida
	default:
		return ErrTokenNoSirve
	}
}

// Linea de la compra (el id es el del producto, como en el píxel y el feed).
type Linea struct {
	ID       string `json:"id"`
	Cantidad int    `json:"cantidad"`
	Precio   int64  `json:"precio"`
}

// Navegador son los datos de quien compra que Meta usa para reconocerla (sólo con su consentimiento).
type Navegador struct {
	IP  string `json:"ip"`
	UA  string `json:"ua"`
	FBP string `json:"fbp"`
	FBC string `json:"fbc"`
	URL string `json:"url"`
}

// Compra es el evento que se encola: el email y el teléfono van ya hasheados (SHA-256).
type Compra struct {
	Numero    int       `json:"numero"`
	Total     int64     `json:"total"`
	Hora      int64     `json:"hora"`
	Lineas    []Linea   `json:"lineas"`
	Email     string    `json:"em,omitempty"`
	Telefono  string    `json:"ph,omitempty"`
	Navegador Navegador `json:"navegador"`
}

func hash(s string) string {
	if s == "" {
		return ""
	}
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// HashEmail y HashTelefono normalizan como pide Meta (minúsculas; sólo dígitos con el 54 adelante).
func HashEmail(email string) string { return hash(strings.ToLower(strings.TrimSpace(email))) }

func HashTelefono(tel string) string {
	d := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, tel)
	d = strings.TrimLeft(d, "0")
	if len(d) < 8 {
		return ""
	}
	if !strings.HasPrefix(d, "54") {
		d = "54" + d
	}
	return hash(d)
}

// EnTienda es lo que hace falta del pool para encolar (la vidriera tiene el suyo).
type EnTienda interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
}

// Encolar deja la compra para el worker, sólo si la tienda tiene la API de conversiones conectada.
func Encolar(ctx context.Context, pool EnTienda, tienda string, c Compra) error {
	return pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		var hay bool
		if err := t.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM meta_conversiones WHERE tienda_id = $1)`, tienda).Scan(&hay); err != nil || !hay {
			return err
		}
		return outbox.Publicar(ctx, t, tienda, TipoCompra, c)
	})
}

// Manejadores del outbox.
func (s *Servicio) Manejadores() map[string]outbox.Manejador {
	return map[string]outbox.Manejador{TipoCompra: s.enviarCompra}
}

// evento arma el Purchase de la Graph API.
func evento(c Compra) map[string]any {
	ids := make([]string, 0, len(c.Lineas))
	contenidos := make([]map[string]any, 0, len(c.Lineas))
	unidades := 0
	for _, l := range c.Lineas {
		ids = append(ids, l.ID)
		contenidos = append(contenidos, map[string]any{"id": l.ID, "quantity": l.Cantidad, "item_price": l.Precio})
		unidades += l.Cantidad
	}
	usuario := map[string]any{}
	if c.Email != "" {
		usuario["em"] = []string{c.Email}
	}
	if c.Telefono != "" {
		usuario["ph"] = []string{c.Telefono}
	}
	for k, v := range map[string]string{"client_ip_address": c.Navegador.IP, "client_user_agent": c.Navegador.UA,
		"fbp": c.Navegador.FBP, "fbc": c.Navegador.FBC} {
		if v != "" {
			usuario[k] = v
		}
	}
	ev := map[string]any{
		"event_name": "Purchase", "event_time": c.Hora, "event_id": fmt.Sprint(c.Numero), "action_source": "website",
		"user_data": usuario,
		"custom_data": map[string]any{"currency": "ARS", "value": c.Total, "content_type": "product",
			"content_ids": ids, "contents": contenidos, "num_items": unidades, "order_id": fmt.Sprint(c.Numero)},
	}
	if c.Navegador.URL != "" {
		ev["event_source_url"] = c.Navegador.URL
	}
	return ev
}

func (s *Servicio) enviarCompra(ctx context.Context, e outbox.Evento) error {
	var c Compra
	if err := json.Unmarshal(e.Datos, &c); err != nil {
		return nil // un evento roto no se reintenta
	}
	// Meta no acepta compras de hace más de 7 días: después de un corte largo, se descarta.
	if time.Since(time.Unix(c.Hora, 0)) > 7*24*time.Hour-time.Hour {
		return nil
	}
	var cif []byte
	var codigo, pixel *string
	err := s.Pool.EnTienda(ctx, e.TiendaID, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT m.token, m.codigo_prueba, t.pixel_meta FROM meta_conversiones m JOIN tiendas t ON t.id = m.tienda_id
			WHERE m.tienda_id = $1`, e.TiendaID).Scan(&cif, &codigo, &pixel)
	})
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && (pixel == nil || *pixel == "")) {
		return nil // se desconectó o borró el píxel después de la compra
	}
	if err != nil {
		return err
	}
	if s.Cifra == nil {
		return ErrSinCifrado
	}
	token, err := s.Cifra.Descifrar(contexto(e.TiendaID), cif)
	if err != nil {
		return err
	}
	cuerpo := map[string]any{"data": []any{evento(c)}, "access_token": string(token)}
	if codigo != nil {
		cuerpo["test_event_code"] = *codigo
	}
	b, _ := json.Marshal(cuerpo)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url(*pixel, "events"), bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := s.cli().Do(req)
	if err != nil {
		return ErrMetaCaida // el error de red no se propaga (no tiene que llevar nada del pedido)
	}
	defer res.Body.Close()
	resp, _ := io.ReadAll(io.LimitReader(res.Body, 64<<10))
	if res.StatusCode >= 500 || res.StatusCode == http.StatusTooManyRequests {
		return fmt.Errorf("conversiones: Meta respondió %d", res.StatusCode)
	}
	var motivo *string
	if res.StatusCode != http.StatusOK {
		// Token vencido, píxel ajeno, evento rechazado: no se reintenta; la tienda lo ve en el panel.
		var r struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		_ = json.Unmarshal(resp, &r)
		m := strings.TrimSpace(r.Error.Message)
		if m == "" {
			m = fmt.Sprintf("Meta respondió %d.", res.StatusCode)
		}
		if rs := []rune(m); len(rs) > 280 {
			m = string(rs[:280])
		}
		motivo = &m
	}
	return s.Pool.EnTienda(ctx, e.TiendaID, func(t pgx.Tx) error {
		if motivo != nil {
			_, err := t.Exec(ctx, `UPDATE meta_conversiones SET ultimo_error = $2 WHERE tienda_id = $1`, e.TiendaID, motivo)
			return err
		}
		_, err := t.Exec(ctx, `UPDATE meta_conversiones SET ultimo_envio = now(), ultimo_error = NULL WHERE tienda_id = $1`, e.TiendaID)
		return err
	})
}
