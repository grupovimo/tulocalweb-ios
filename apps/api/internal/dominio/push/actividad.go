package push

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/roles"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/apns"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/fcm"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

// La Live Activity de los pedidos por preparar (docs/decisiones.md §54). La app la arranca y la
// actualiza cuando está abierta; con la app cerrada, el worker la actualiza (o la arranca, con el
// token de «push to start») al entrar una venta. En Android es una notificación en curso: el
// teléfono pasa su token de FCM como el de la actividad y el de inicio, y el worker le manda el
// estado como datos (tipo «actividad»).

// EstadoActividad es el ContentState de PedidosPorPreparar en la app (Compartido/Actividad.swift):
// los nombres de los campos tienen que coincidir.
type EstadoActividad struct {
	PorPreparar   int     `json:"porPreparar"`
	PreparadosHoy int     `json:"preparadosHoy"`
	UltimoNumero  *int    `json:"ultimoNumero,omitempty"`
	UltimoCliente *string `json:"ultimoCliente,omitempty"`
}

// TokensActividad: lo que la app avisa (PUT /v1/panel/dispositivos/actividad).
type TokensActividad struct {
	Token     string `json:"token"`
	Actividad string `json:"actividad"`
	Inicio    string `json:"inicio"`
}

var tokenActividad = regexp.MustCompile(`^[0-9a-f]{32,400}$`)

// GuardarActividad guarda los tokens de la actividad del teléfono de la sesión. Actividad vacía: terminó.
func GuardarActividad(ctx context.Context, p Pool, usuario, sesion string, t TokensActividad) error {
	if tok := strings.TrimSpace(t.Token); !tokenAPNs.MatchString(strings.ToLower(tok)) && TokenFCM(tok) {
		// Android: el token de FCM, tal cual, también como actividad e inicio.
		t.Token, t.Actividad, t.Inicio = tok, strings.TrimSpace(t.Actividad), strings.TrimSpace(t.Inicio)
		if (t.Actividad != "" && t.Actividad != tok) || (t.Inicio != "" && t.Inicio != tok) {
			return ErrInvalido
		}
		return p.EnUsuario(ctx, usuario, func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `SELECT app.actividad_dispositivo($1, $2, $3, $4)`, sesion, t.Token, t.Actividad, t.Inicio)
			return err
		})
	}
	t.Token = strings.ToLower(strings.TrimSpace(t.Token))
	t.Actividad = strings.ToLower(strings.TrimSpace(t.Actividad))
	t.Inicio = strings.ToLower(strings.TrimSpace(t.Inicio))
	if !tokenAPNs.MatchString(t.Token) || (t.Actividad != "" && !tokenActividad.MatchString(t.Actividad)) ||
		(t.Inicio != "" && !tokenActividad.MatchString(t.Inicio)) {
		return ErrInvalido
	}
	return p.EnUsuario(ctx, usuario, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `SELECT app.actividad_dispositivo($1, $2, $3, $4)`, sesion, t.Token, t.Actividad, t.Inicio)
		return err
	})
}

func rutasActividad(mux *http.ServeMux, p Pool) {
	mux.HandleFunc("PUT /v1/panel/dispositivos/actividad", func(w http.ResponseWriter, r *http.Request) {
		d, ok := cuentas.SesionDe(r)
		if !ok {
			problema.Escribir(w, problema.SinSesion())
			return
		}
		if d.SesionHash == "" || d.IA != "" || d.Mirada {
			problema.Escribir(w, problema.SinPermiso("La actividad del teléfono es de la sesión de una persona."))
			return
		}
		var t TokensActividad
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&t); err != nil {
			problema.Escribir(w, problema.Invalido("Revisá los datos."))
			return
		}
		err := GuardarActividad(r.Context(), p, d.Usuario.ID, d.SesionHash, t)
		switch {
		case errors.Is(err, ErrInvalido):
			problema.Escribir(w, problema.Invalido("Los tokens no son válidos."))
		case err != nil:
			slog.Error("push: guardar actividad", "error", err)
			problema.Escribir(w, problema.Interno())
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	})
}

// PayloadActividad arma el push de la Live Activity: update, end o start (con los atributos y un
// aviso, que Apple pide para arrancarla).
func PayloadActividad(evento string, e EstadoActividad, tienda string, ahora time.Time) []byte {
	aps := map[string]any{
		"timestamp":     ahora.Unix(),
		"event":         evento,
		"content-state": e,
	}
	switch evento {
	case "end":
		aps["dismissal-date"] = ahora.Add(15 * time.Minute).Unix()
	case "start":
		aps["attributes-type"] = "PedidosPorPreparar"
		aps["attributes"] = map[string]string{"tienda": tienda}
		aps["alert"] = map[string]string{"title": "Pedidos por preparar", "body": "Entró un pedido nuevo."}
	}
	b, _ := json.Marshal(map[string]any{"aps": aps})
	return b
}

// DatosActividad: lo que lee la notificación en curso de Android (ActividadEnApp.desdePush), como
// texto. terminar «1»: cerrarla.
func DatosActividad(e EstadoActividad, tienda string, terminar bool) map[string]string {
	d := map[string]string{"tipo": "actividad", "tienda": tienda,
		"porPreparar": strconv.Itoa(e.PorPreparar), "preparadosHoy": strconv.Itoa(e.PreparadosHoy)}
	if e.UltimoNumero != nil {
		d["ultimoNumero"] = strconv.Itoa(*e.UltimoNumero)
	}
	if e.UltimoCliente != nil {
		d["ultimoCliente"] = *e.UltimoCliente
	}
	if terminar {
		d["terminar"] = "1"
	}
	return d
}

// estadoActividad: cuántos pedidos esperan preparación, cuántos se prepararon hoy y el más nuevo.
func (s *Servicio) estadoActividad(ctx context.Context, tienda string) (EstadoActividad, string, error) {
	var e EstadoActividad
	var nombre string
	err := s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		if err := t.QueryRow(ctx, `SELECT nombre FROM tiendas WHERE id = app.tienda_actual()`).Scan(&nombre); err != nil {
			return err
		}
		if err := t.QueryRow(ctx, `SELECT count(*) FILTER (WHERE estado = 'Para preparar'),
				count(*) FILTER (WHERE estado IN ('Listo para despachar', 'En camino', 'Entregado')
					AND creado_en >= date_trunc('day', now() AT TIME ZONE 'America/Argentina/Buenos_Aires') AT TIME ZONE 'America/Argentina/Buenos_Aires')
			FROM pedidos`).Scan(&e.PorPreparar, &e.PreparadosHoy); err != nil {
			return err
		}
		var n int
		var c string
		err := t.QueryRow(ctx, `SELECT numero, coalesce(comprador->>'nombre', '') FROM pedidos
			WHERE estado = 'Para preparar' ORDER BY creado_en DESC LIMIT 1`).Scan(&n, &c)
		if err == nil {
			e.UltimoNumero, e.UltimoCliente = &n, &c
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		return nil
	})
	return e, nombre, err
}

// actualizarActividades manda el estado nuevo a los teléfonos del equipo con Pedidos: actualiza (o
// cierra, si no queda ninguno) la actividad en curso, o la arranca con el token de inicio. No frena
// el aviso de venta: lo que falle se anota y listo.
func (s *Servicio) actualizarActividades(ctx context.Context, tienda string) {
	if !s.telefonos() {
		return
	}
	e, nombre, err := s.estadoActividad(ctx, tienda)
	if err != nil {
		s.Log.Warn("push: estado de la actividad", "error", err)
		return
	}
	type fila struct{ rol, actividad, inicio, entorno, plataforma string }
	var filas []fila
	err = s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		rs, err := t.Query(ctx, `SELECT rol, coalesce(token_actividad, ''), coalesce(token_inicio, ''), entorno, plataforma FROM app.actividades_de_tienda()`)
		if err != nil {
			return err
		}
		filas, err = pgx.CollectRows(rs, func(r pgx.CollectableRow) (fila, error) {
			var f fila
			err := r.Scan(&f.rol, &f.actividad, &f.inicio, &f.entorno, &f.plataforma)
			return f, err
		})
		return err
	})
	if err != nil {
		s.Log.Warn("push: teléfonos con actividad", "error", err)
		return
	}
	ahora := time.Now()
	for _, f := range filas {
		if !roles.Puede(f.rol, roles.Pedidos) || !s.sale(f.plataforma) {
			continue
		}
		var token, evento string
		switch {
		case f.actividad != "":
			token, evento = f.actividad, "update"
			if e.PorPreparar == 0 {
				evento = "end"
			}
		case f.inicio != "" && e.PorPreparar > 0:
			token, evento = f.inicio, "start"
		default:
			continue
		}
		var err error
		if f.plataforma == "android" {
			err = s.FCM.Enviar(ctx, fcm.Envio{Token: token, Datos: DatosActividad(e, nombre, evento == "end"),
				Colapso: "actividad-" + tienda, Vida: 4 * time.Hour, SoloDatos: true})
		} else {
			err = s.APNs.Enviar(ctx, apns.Envio{Token: token, Entorno: f.entorno, Payload: PayloadActividad(evento, e, nombre, ahora),
				Vence: ahora.Add(4 * time.Hour), Actividad: true})
		}
		if errors.Is(err, apns.ErrTokenInvalido) || errors.Is(err, fcm.ErrTokenInvalido) {
			_ = s.Pool.EnTx(ctx, func(t pgx.Tx) error {
				_, err := t.Exec(ctx, `SELECT app.olvidar_actividad($1)`, token)
				return err
			})
		} else if err != nil {
			s.Log.Warn("push: actividad", "error", err)
		}
	}
}
