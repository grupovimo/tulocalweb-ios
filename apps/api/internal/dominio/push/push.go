// Package push: los avisos al iPhone de quienes atienden la tienda (docs/decisiones.md §40).
//
// La app registra el token de APNs de cada iPhone atado a la sesión con que entró
// (POST /v1/panel/dispositivos) y lo da de baja al salir (DELETE). El worker manda, por el outbox:
// «Nueva venta #N» (push.venta, el mismo momento que el mail tienda.venta), «Pregunta nueva en
// Mercado Libre» (push.pregunta_ml), «Te quedaste sin stock» (push.sin_stock, §41) y «Entraron a tu
// cuenta» (cuenta.ingreso_push, sólo a la persona). Cada persona recibe lo que su rol puede ver: los montos, sólo con Finanzas.
package push

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

// Pool: el de la app (rol mostrador_app). *db.Pool lo cumple.
type Pool interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
	EnUsuario(context.Context, string, func(pgx.Tx) error) error
	EnTx(context.Context, func(pgx.Tx) error) error
}

// Dispositivo es lo que manda la app al registrarse (components/schemas/Dispositivo).
type Dispositivo struct {
	Token   string `json:"token"`
	Entorno string `json:"entorno"`
	Modelo  string `json:"modelo"`
	Version string `json:"version"`
}

// ErrInvalido: el token o el entorno no tienen la forma esperada.
var ErrInvalido = errors.New("dispositivo inválido")

var tokenAPNs = regexp.MustCompile(`^[0-9a-f]{64,200}$`)

func limpiar(s string, max int) string {
	s = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, strings.TrimSpace(s))
	for utf8.RuneCountInString(s) > max {
		_, n := utf8.DecodeLastRuneInString(s)
		s = s[:len(s)-n]
	}
	return s
}

// Normalizar valida y limpia lo que mandó la app.
func Normalizar(d Dispositivo) (Dispositivo, error) {
	d.Token = strings.ToLower(strings.TrimSpace(d.Token))
	if !tokenAPNs.MatchString(d.Token) {
		return d, ErrInvalido
	}
	switch d.Entorno {
	case "sandbox", "produccion":
	case "":
		d.Entorno = "produccion"
	default:
		return d, ErrInvalido
	}
	d.Modelo, d.Version = limpiar(d.Modelo, 60), limpiar(d.Version, 30)
	return d, nil
}

// Registrar ata el iPhone a la sesión (sesion es el hash del token de la sesión) y a la persona.
// Si el token ya estaba con otra sesión, se mueve a esta.
func Registrar(ctx context.Context, p Pool, usuario, sesion string, d Dispositivo) error {
	d, err := Normalizar(d)
	if err != nil {
		return err
	}
	return p.EnUsuario(ctx, usuario, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `SELECT app.registrar_dispositivo($1, $2, $3, $4, $5)`, sesion, d.Token, d.Entorno, d.Modelo, d.Version)
		return err
	})
}

// Baja borra el iPhone de esta sesión (la app lo llama al cerrar sesión; cerrar la sesión también
// lo borra).
func Baja(ctx context.Context, p Pool, usuario, sesion string) error {
	return p.EnUsuario(ctx, usuario, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `DELETE FROM auth.dispositivos WHERE sesion_hash = $1`, sesion)
		return err
	})
}

// Rutas del panel: POST y DELETE /v1/panel/dispositivos (roles.Panel: cualquiera del equipo). Sólo
// con la sesión de una persona: ni una IA conectada ni el soporte mirando la tienda.
func Rutas(mux *http.ServeMux, p Pool) {
	con := func(h func(http.ResponseWriter, *http.Request, cuentas.Datos)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			d, ok := cuentas.SesionDe(r)
			if !ok {
				problema.Escribir(w, problema.SinSesion())
				return
			}
			if d.SesionHash == "" || d.IA != "" || d.Mirada {
				problema.Escribir(w, problema.SinPermiso("Los avisos al iPhone son de la sesión de una persona."))
				return
			}
			h(w, r, d)
		}
	}
	mux.HandleFunc("POST /v1/panel/dispositivos", con(func(w http.ResponseWriter, r *http.Request, s cuentas.Datos) {
		var d Dispositivo
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&d); err != nil {
			problema.Escribir(w, problema.Invalido("Revisá los datos."))
			return
		}
		err := Registrar(r.Context(), p, s.Usuario.ID, s.SesionHash, d)
		switch {
		case errors.Is(err, ErrInvalido):
			problema.Escribir(w, problema.Invalido("El token del iPhone o el entorno no son válidos.",
				problema.Campo{Ruta: "token", Mensaje: "Tiene que ser el token de APNs en hexadecimal."}))
		case err != nil:
			slog.Error("push: registrar dispositivo", "error", err)
			problema.Escribir(w, problema.Interno())
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	mux.HandleFunc("DELETE /v1/panel/dispositivos", con(func(w http.ResponseWriter, r *http.Request, s cuentas.Datos) {
		if err := Baja(r.Context(), p, s.Usuario.ID, s.SesionHash); err != nil {
			slog.Error("push: baja de dispositivo", "error", err)
			problema.Escribir(w, problema.Interno())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	rutasActividad(mux, p)
}

// Aviso es una notificación ya decidida, antes de pasar a JSON.
type Aviso struct {
	Titulo, Cuerpo string
	// Hilo (thread-id): agrupa en el iPhone; uno por tienda, o "cuenta" para lo de la persona.
	Hilo string
	// Ruta de la app para abrir la pantalla (pedido/12, preguntas/ml/123, cuenta/seguridad).
	Ruta string
	// Tienda: el id, para que la app cambie de tienda antes de abrir la ruta si hace falta.
	Tienda string
}

// Payload arma el JSON de APNs: aps.alert con título y cuerpo, sonido, thread-id y, fuera de aps,
// la ruta (y la tienda) para el enlace profundo.
func Payload(a Aviso) []byte {
	type alerta struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	type aps struct {
		Alert    alerta `json:"alert"`
		Sound    string `json:"sound"`
		ThreadID string `json:"thread-id,omitempty"`
	}
	b, _ := json.Marshal(struct {
		Aps    aps    `json:"aps"`
		Ruta   string `json:"ruta,omitempty"`
		Tienda string `json:"tienda,omitempty"`
	}{aps{alerta{a.Titulo, a.Cuerpo}, "default", a.Hilo}, a.Ruta, a.Tienda})
	return b
}

func pesos(n int64) string {
	s := fmt.Sprint(n)
	var b strings.Builder
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(r)
	}
	return "$ " + b.String()
}

// HiloTienda es el thread-id de los avisos de una tienda.
func HiloTienda(tienda string) string { return "tienda-" + tienda }
