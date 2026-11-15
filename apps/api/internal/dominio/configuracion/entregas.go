package configuracion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cupones"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/auditoria"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

// Entregas es components/schemas/Entregas: cómo entrega la tienda (Configuración › Envíos y retiro).
// Correo Argentino no está acá: se ofrece si la cuenta de MiCorreo está conectada.
type Entregas struct {
	Retiro          bool   `json:"retiro"`
	RetiroDireccion string `json:"retiroDireccion"`
	RetiroHorario   string `json:"retiroHorario"`
	Moto            bool   `json:"moto"`
	// MotoCosto: en pesos, fijo para toda la zona.
	MotoCosto  int64  `json:"motoCosto"`
	MotoZona   string `json:"motoZona"`
	MotoDemora string `json:"motoDemora"`
	// El envío gratis (el interruptor es la promoción automática «envio» de Cupones) y desde qué monto,
	// en pesos. Al guardar, sin mandar: queda como estaba (las apps viejas no lo mandan).
	EnvioGratis      *bool  `json:"envioGratis,omitempty"`
	EnvioGratisMonto *int64 `json:"envioGratisMonto,omitempty"`
}

// EstadoEntregas es components/schemas/EstadoEntregas: lo guardado más lo que se ve desde otras
// pantallas (Correo en Integraciones, el envío gratis en Cupones › Automáticas).
type EstadoEntregas struct {
	Entregas
	Correo      bool   `json:"correo"`
	GratisDesde *int64 `json:"gratisDesde"`
}

// ErrSinEntrega: se apagó todo y Correo no está conectado (nadie podría recibir lo que compra).
var ErrSinEntrega = errors.New("configuracion: sin ninguna forma de entrega")

// ErrEntrega: un dato de la entrega no va (el mensaje es para la persona).
type ErrEntrega struct{ Motivo string }

func (e ErrEntrega) Error() string { return e.Motivo }

// Topes (los mismos de la migración 00049 y de la web).
const (
	maxDireccion = 160
	maxHorario   = 80
	maxZona      = 160
	maxDemora    = 60
	maxCostoMoto = 1_000_000
)

// EntregasEnTx lee la configuración de entrega de la tienda dentro de una transacción ya fijada en
// ella. Sin fila: retiro en el local y sin moto. Dirección y horario vacíos: los del negocio.
func EntregasEnTx(ctx context.Context, t pgx.Tx, tiendaID string) (Entregas, error) {
	var e Entregas
	var negocio []byte
	err := t.QueryRow(ctx, `SELECT coalesce(e.retiro, true), coalesce(e.retiro_direccion, ''), coalesce(e.retiro_horario, ''),
			coalesce(e.moto, false), coalesce(e.moto_costo, 0), coalesce(e.moto_zona, ''), coalesce(e.moto_demora, ''), t.negocio
		FROM tiendas t LEFT JOIN entregas e ON e.tienda_id = t.id WHERE t.id = $1`, tiendaID).
		Scan(&e.Retiro, &e.RetiroDireccion, &e.RetiroHorario, &e.Moto, &e.MotoCosto, &e.MotoZona, &e.MotoDemora, &negocio)
	if errors.Is(err, pgx.ErrNoRows) {
		return e, ErrNoExiste
	}
	if err != nil {
		return e, err
	}
	var n struct{ Direccion, Horario string }
	_ = json.Unmarshal(negocio, &n)
	if strings.TrimSpace(e.RetiroDireccion) == "" {
		e.RetiroDireccion = strings.TrimSpace(n.Direccion)
	}
	if strings.TrimSpace(e.RetiroHorario) == "" {
		e.RetiroHorario = strings.TrimSpace(n.Horario)
	}
	return e, nil
}

// LeerEntregas devuelve la configuración con el estado de Correo y del envío gratis.
func LeerEntregas(ctx context.Context, pool tx, tiendaID string) (EstadoEntregas, error) {
	var out EstadoEntregas
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var err error
		if out.Entregas, err = EntregasEnTx(ctx, t, tiendaID); err != nil {
			return err
		}
		if err := t.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM ca_conexiones WHERE tienda_id = $1)`, tiendaID).Scan(&out.Correo); err != nil {
			return err
		}
		activa, desde, err := cupones.EnvioGratis(ctx, t, tiendaID)
		if err != nil {
			return err
		}
		out.EnvioGratis, out.EnvioGratisMonto = &activa, &desde
		if activa {
			out.GratisDesde = &desde
		}
		return nil
	})
	return out, err
}

func texto(s string, max int) (string, bool) {
	s = strings.TrimSpace(s)
	return s, utf8.RuneCountInString(s) <= max
}

// GuardarEntregas valida y guarda. Tiene que quedar al menos una forma de entrega (retiro, moto o
// Correo conectado); el retiro pide una dirección (la propia o la del negocio) y la moto, la zona.
func GuardarEntregas(ctx context.Context, pool tx, tiendaID string, e Entregas) error {
	var ok [4]bool
	e.RetiroDireccion, ok[0] = texto(e.RetiroDireccion, maxDireccion)
	e.RetiroHorario, ok[1] = texto(e.RetiroHorario, maxHorario)
	e.MotoZona, ok[2] = texto(e.MotoZona, maxZona)
	e.MotoDemora, ok[3] = texto(e.MotoDemora, maxDemora)
	if !ok[0] || !ok[1] || !ok[2] || !ok[3] || e.MotoCosto < 0 || e.MotoCosto > maxCostoMoto {
		return ErrInvalido
	}
	if m := e.EnvioGratisMonto; m != nil && (*m < cupones.EnvioGratisMinimo || *m > cupones.EnvioGratisMaximo) {
		return ErrEntrega{Motivo: "El envío gratis va desde $ 1.000 hasta $ 100.000.000."}
	}
	if e.Moto && utf8.RuneCountInString(e.MotoZona) < 3 {
		return ErrEntrega{Motivo: "Poné hasta dónde llega la moto (barrios o localidades)."}
	}
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var correo bool
		var negocio []byte
		if err := t.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM ca_conexiones WHERE tienda_id = $1), negocio
			FROM tiendas WHERE id = $1`, tiendaID).Scan(&correo, &negocio); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNoExiste
			}
			return err
		}
		if !e.Retiro && !e.Moto && !correo {
			return ErrSinEntrega
		}
		var n struct{ Direccion string }
		_ = json.Unmarshal(negocio, &n)
		if e.Retiro && e.RetiroDireccion == "" && strings.TrimSpace(n.Direccion) == "" {
			return ErrEntrega{Motivo: "Poné la dirección donde retiran."}
		}
		antes, err := EntregasEnTx(ctx, t, tiendaID)
		if err != nil {
			return err
		}
		if _, err := t.Exec(ctx, `INSERT INTO entregas (tienda_id, retiro, retiro_direccion, retiro_horario, moto, moto_costo, moto_zona, moto_demora)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (tienda_id) DO UPDATE SET retiro = $2, retiro_direccion = $3, retiro_horario = $4, moto = $5,
				moto_costo = $6, moto_zona = $7, moto_demora = $8, actualizada_en = now()`,
			tiendaID, e.Retiro, e.RetiroDireccion, e.RetiroHorario, e.Moto, e.MotoCosto, e.MotoZona, e.MotoDemora); err != nil {
			return err
		}
		despues, err := EntregasEnTx(ctx, t, tiendaID)
		if err != nil {
			return err
		}
		if err := auditoria.Registrar(ctx, t, tiendaID, registrosEntregas(antes, despues)...); err != nil {
			return err
		}
		return cupones.FijarEnvioGratis(ctx, t, tiendaID, e.EnvioGratis, e.EnvioGratisMonto)
	})
}

// registrosEntregas: lo que cambió de Envíos y retiro, como se lee en la Actividad.
func registrosEntregas(a, d Entregas) []auditoria.Registro {
	var rs []auditoria.Registro
	reg := func(texto string) {
		rs = append(rs, auditoria.Registro{Accion: "envios.configurar", Entidad: "entregas", Texto: texto, Antes: a, Despues: d})
	}
	switch {
	case a.Retiro && !d.Retiro:
		reg("apagó el retiro en el local")
	case !a.Retiro && d.Retiro:
		reg("prendió el retiro en el local")
	case d.Retiro && (a.RetiroDireccion != d.RetiroDireccion || a.RetiroHorario != d.RetiroHorario):
		reg("cambió la dirección o el horario de retiro")
	}
	switch {
	case a.Moto && !d.Moto:
		reg("apagó el envío en moto")
	case !a.Moto && d.Moto:
		reg(fmt.Sprintf("prendió el envío en moto a %s", auditoria.Pesos(d.MotoCosto)))
	case d.Moto && a.MotoCosto != d.MotoCosto:
		reg(fmt.Sprintf("cambió el costo del envío en moto de %s a %s", auditoria.Pesos(a.MotoCosto), auditoria.Pesos(d.MotoCosto)))
	case d.Moto && (a.MotoZona != d.MotoZona || a.MotoDemora != d.MotoDemora):
		reg("cambió la zona o la demora del envío en moto")
	}
	return rs
}

func rutasEntregas(mux *http.ServeMux, pool tx, con func(func(http.ResponseWriter, *http.Request, string, Quien)) http.HandlerFunc,
	errorDe func(http.ResponseWriter, error), leer func(http.ResponseWriter, *http.Request, any) bool) {
	mux.HandleFunc("GET /v1/panel/configuracion/entregas", con(func(w http.ResponseWriter, r *http.Request, id string, _ Quien) {
		e, err := LeerEntregas(r.Context(), pool, id)
		if err != nil {
			errorDe(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(e)
	}))
	mux.HandleFunc("PUT /v1/panel/configuracion/entregas", con(func(w http.ResponseWriter, r *http.Request, id string, _ Quien) {
		var e Entregas
		if !leer(w, r, &e) {
			return
		}
		err := GuardarEntregas(r.Context(), pool, id, e)
		var mal ErrEntrega
		switch {
		case errors.Is(err, ErrSinEntrega):
			problema.Escribir(w, problema.Invalido("Dejá al menos una forma de entrega: retiro, moto o Correo Argentino."))
		case errors.As(err, &mal):
			problema.Escribir(w, problema.Invalido(mal.Motivo))
		default:
			errorDe(w, err)
		}
	}))
}
