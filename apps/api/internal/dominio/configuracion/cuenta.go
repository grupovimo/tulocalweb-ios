package configuracion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/abono"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/roles"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/auditoria"
)

func mes(t time.Time) string { return t.In(time.FixedZone("ART", -3*3600)).Format("2006-01") }

// esDuenia: ErrDuenia si id es la dueña de la tienda.
func esDuenia(ctx context.Context, t pgx.Tx, tiendaID, id string) error {
	var es bool
	if err := t.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM auth.membresias WHERE tienda_id = $1 AND usuario_id::text = $2 AND rol = 'duenio')`,
		tiendaID, id).Scan(&es); err != nil {
		return err
	}
	if es {
		return ErrDuenia
	}
	return nil
}

// CambiarRol: a nadie se le da ni se le saca el rol de dueña, y una no se cambia a sí misma. Vale
// para alguien del equipo o para una invitación pendiente.
func CambiarRol(ctx context.Context, pool tx, tiendaID string, q Quien, id, rol string) error {
	if !roles.Invitable(rol) {
		return ErrInvalido
	}
	if id == q.UsuarioID {
		return ErrVos
	}
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		if err := esDuenia(ctx, t, tiendaID, id); err != nil {
			return err
		}
		var quien, antes string
		err := t.QueryRow(ctx, `UPDATE auth.membresias m SET rol = $3 FROM auth.membresias a, auth.usuarios u
			WHERE m.tienda_id = $1 AND m.usuario_id::text = $2 AND m.rol <> 'duenio'
				AND a.tienda_id = m.tienda_id AND a.usuario_id = m.usuario_id AND u.id = m.usuario_id
			RETURNING coalesce(nullif(u.nombre, ''), u.email), a.rol`, tiendaID, id, rol).Scan(&quien, &antes)
		if errors.Is(err, pgx.ErrNoRows) {
			err = t.QueryRow(ctx, `UPDATE invitaciones i SET rol = $3 FROM invitaciones a
				WHERE i.tienda_id = $1 AND i.id::text = $2 AND a.tienda_id = i.tienda_id AND a.id = i.id
				RETURNING 'la invitación de ' || i.contacto, a.rol`, tiendaID, id, rol).Scan(&quien, &antes)
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNoExiste
			}
		}
		if err != nil || antes == rol {
			return err
		}
		return auditoria.Registrar(ctx, t, tiendaID, auditoria.Registro{Accion: "personas.rol", Entidad: "persona", EntidadID: id,
			Texto: fmt.Sprintf("le cambió el rol a %s de %s a %s", quien, roles.Nombre[antes], roles.Nombre[rol]),
			Antes: map[string]string{"rol": antes}, Despues: map[string]string{"rol": rol}})
	})
}

// Quitar saca a alguien del equipo (y le cierra las sesiones que tenía abiertas en esta tienda) o
// cancela la invitación (su link deja de valer). A la dueña no se la quita.
func Quitar(ctx context.Context, pool tx, tiendaID string, q Quien, id string) error {
	if id == q.UsuarioID {
		return ErrVos
	}
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		if err := esDuenia(ctx, t, tiendaID, id); err != nil {
			return err
		}
		var quien, rol string
		err := t.QueryRow(ctx, `DELETE FROM auth.membresias m USING auth.usuarios u
			WHERE m.tienda_id = $1 AND m.usuario_id::text = $2 AND m.rol <> 'duenio' AND u.id = m.usuario_id
			RETURNING coalesce(nullif(u.nombre, ''), u.email), m.rol`, tiendaID, id).Scan(&quien, &rol)
		if err == nil {
			if _, err := t.Exec(ctx, `DELETE FROM auth.sesiones WHERE usuario_id::text = $1 AND tienda_id = $2`, id, tiendaID); err != nil {
				return err
			}
			return auditoria.Registrar(ctx, t, tiendaID, auditoria.Registro{Accion: "personas.quitar", Entidad: "persona", EntidadID: id,
				Texto: fmt.Sprintf("sacó a %s (%s) del equipo", quien, roles.Nombre[rol]), Antes: map[string]string{"rol": rol}})
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		err = t.QueryRow(ctx, `DELETE FROM invitaciones WHERE tienda_id = $1 AND id::text = $2 RETURNING contacto, rol`, tiendaID, id).
			Scan(&quien, &rol)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoExiste
		}
		if err != nil {
			return err
		}
		return auditoria.Registrar(ctx, t, tiendaID, auditoria.Registro{Accion: "personas.cancelar", Entidad: "invitacion", EntidadID: id,
			Texto: fmt.Sprintf("canceló la invitación de %s (%s)", quien, roles.Nombre[rol])})
	})
}

// Plan arma la pantalla Plan y facturación: el abono, lo que cuesta cada cambio (las cotizaciones, la
// misma cuenta con la que se cobra), el uso, los datos de facturación y los cobros.
func Plan(ctx context.Context, pool tx, tiendaID string, ahora time.Time) (map[string]any, error) {
	var res map[string]any
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		e, err := abono.Leer(ctx, t, tiendaID)
		if err != nil {
			return err
		}
		var packs map[string]int
		var negocio Negocio
		var productos, personas int
		if err := t.QueryRow(ctx, `SELECT packs, negocio,
				(SELECT count(*) FROM productos WHERE tienda_id = $1),
				(SELECT count(*) FROM auth.membresias WHERE tienda_id = $1) + (SELECT count(*) FROM invitaciones WHERE tienda_id = $1)
			FROM tiendas WHERE id = $1`, tiendaID).Scan(&packs, &negocio, &productos, &personas); err != nil {
			return err
		}
		ben, usado, err := abono.LeerBeneficio(ctx, t, tiendaID)
		if err != nil {
			return err
		}
		var beneficio any
		if ben.Cobros > 0 {
			beneficio = ben
		}
		enPrueba := ahora.Before(e.FinPrueba)
		// El próximo cobro: el del débito; en la prueba, cuando termina.
		var proximo any
		switch {
		case e.Activo && e.Proximo != nil:
			proximo = *e.Proximo
		case enPrueba:
			proximo = e.FinPrueba
		}
		// Cancelado: la baja (pasar a gratis) programada para cuando termina lo pagado.
		var canc any
		if e.Pendiente != nil && e.Pendiente.Plan == "gratis" {
			canc = map[string]any{"hasta": e.Pendiente.Desde}
		}
		tipo := "Factura B"
		if negocio.Cuit != "" {
			tipo = "Factura A"
		}
		facturas := []map[string]any{}
		filas, err := t.Query(ctx, `SELECT clave, detalle, importe, creado_en FROM abono_cobros WHERE tienda_id = $1
			ORDER BY creado_en DESC LIMIT 24`, tiendaID)
		if err != nil {
			return err
		}
		for filas.Next() {
			var clave, detalle string
			var importe int64
			var cuando time.Time
			if err := filas.Scan(&clave, &detalle, &importe, &cuando); err != nil {
				filas.Close()
				return err
			}
			facturas = append(facturas, map[string]any{"numero": clave, "detalle": detalle, "total": importe, "fecha": cuando})
		}
		filas.Close()
		if err := filas.Err(); err != nil {
			return err
		}
		res = map[string]any{
			"actual": e.Plan, "ciclo": e.Ciclo, "enPrueba": enPrueba, "diasPrueba": max(0, int(math.Ceil(e.FinPrueba.Sub(ahora).Hours()/24))),
			"finPrueba":      e.FinPrueba,
			"proximoCobro":   proximo,
			"abono":          map[string]any{"activo": e.Activo, "proximo": e.Proximo, "programado": e.Pendiente},
			"cotizaciones":   abono.Todas(e, ahora),
			"beneficio":      beneficio,
			"beneficioUsado": usado,
			// Las respuestas de Hornero se cuentan cuando Hornero corra en la API.
			"uso":         map[string]any{"productos": productos, "personas": personas, "hornero": 0, "packs": packs[mes(ahora)], "sinIa": 0},
			"medio":       nil,
			"facturacion": map[string]any{"razonSocial": negocio.RazonSocial, "cuit": negocio.Cuit, "tipo": tipo},
			"facturas":    facturas,
			"cancelado":   canc,
		}
		return nil
	})
	return res, err
}

// Ganchos de la cuenta que viven en otros paquetes.
type Ganchos struct {
	// Beneficio para quedarse (§57) con el débito de Mercado Pago. nil: sólo queda anotado.
	Beneficio func(ctx context.Context, tiendaID string, b abono.Beneficio) error
	// Contacto: avisa al equipo que una dueña pidió que le escriban. nil: no se avisa.
	Contacto func(ctx context.Context, asunto, texto string) error
}

func rutasCuenta(mux *http.ServeMux, pool tx, g Ganchos, con func(func(http.ResponseWriter, *http.Request, string, Quien)) http.HandlerFunc,
	errorDe func(http.ResponseWriter, error), leer func(http.ResponseWriter, *http.Request, any) bool) {
	escribir := func(w http.ResponseWriter, v any, err error) {
		if err != nil {
			errorDe(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(v)
	}
	mux.HandleFunc("GET /v1/panel/cuenta/plan", con(func(w http.ResponseWriter, r *http.Request, id string, _ Quien) {
		v, err := Plan(r.Context(), pool, id, time.Now())
		escribir(w, v, err)
	}))
	mux.HandleFunc("POST /v1/panel/cuenta/pack", con(func(w http.ResponseWriter, r *http.Request, id string, _ Quien) {
		errorDe(w, pool.EnTienda(r.Context(), id, func(t pgx.Tx) error {
			if _, err := t.Exec(r.Context(), `UPDATE tiendas SET packs = packs || jsonb_build_object($2::text, coalesce((packs->>$2)::int, 0) + 1) WHERE id = $1`, id, mes(time.Now())); err != nil {
				return err
			}
			return auditoria.Registrar(r.Context(), t, id, auditoria.Registro{Accion: "plan.pack", Entidad: "tienda", EntidadID: id,
				Texto: "sumó un pack extra este mes"})
		}))
	}))
	mux.HandleFunc("DELETE /v1/panel/cuenta/pack", con(func(w http.ResponseWriter, r *http.Request, id string, _ Quien) {
		errorDe(w, pool.EnTienda(r.Context(), id, func(t pgx.Tx) error {
			tag, err := t.Exec(r.Context(), `UPDATE tiendas SET packs = packs || jsonb_build_object($2::text, (packs->>$2)::int - 1)
				WHERE id = $1 AND coalesce((packs->>$2)::int, 0) > 0`, id, mes(time.Now()))
			if err == nil && tag.RowsAffected() == 0 {
				return ErrNoExiste
			}
			if err != nil {
				return err
			}
			return auditoria.Registrar(r.Context(), t, id, auditoria.Registro{Accion: "plan.pack_quitar", Entidad: "tienda", EntidadID: id,
				Texto: "sacó un pack extra de este mes"})
		}))
	}))
	mux.HandleFunc("POST /v1/panel/cuenta/cancelacion", con(func(w http.ResponseWriter, r *http.Request, id string, q Quien) {
		var b struct {
			Motivo int    `json:"motivo"`
			Accion string `json:"accion"`
		}
		if !leer(w, r, &b) {
			return
		}
		// La oferta para quedarse (§57): un descuento que se cobra de verdad, o que alguien le escriba.
		if b.Accion == "oferta" {
			if ben, ok := abono.Ofertas[b.Motivo]; ok {
				otorgar := g.Beneficio
				if otorgar == nil {
					otorgar = func(ctx context.Context, tiendaID string, b abono.Beneficio) error {
						return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error { return abono.Otorgar(ctx, t, tiendaID, b) })
					}
				}
				escribir(w, map[string]any{"beneficio": ben}, otorgar(r.Context(), id, ben))
				return
			}
			escribir(w, map[string]any{}, pedirContacto(r.Context(), pool, g, id, q.UsuarioID, b.Motivo))
			return
		}
		if b.Accion == "cancelar" {
			if err := pool.EnTienda(r.Context(), id, func(t pgx.Tx) error {
				if _, err := t.Exec(r.Context(), `UPDATE tiendas SET cancelada_en = now() WHERE id = $1`, id); err != nil {
					return err
				}
				return auditoria.Registrar(r.Context(), t, id, auditoria.Registro{Accion: "plan.cancelar", Entidad: "tienda", EntidadID: id,
					Texto: "canceló el plan (sigue hasta el próximo cobro)"})
			}); err != nil {
				errorDe(w, err)
				return
			}
		}
		v, err := Plan(r.Context(), pool, id, time.Now())
		if err != nil {
			errorDe(w, err)
			return
		}
		escribir(w, map[string]any{"hasta": v["proximoCobro"]}, nil)
	}))
	mux.HandleFunc("DELETE /v1/panel/cuenta/cancelacion", con(func(w http.ResponseWriter, r *http.Request, id string, _ Quien) {
		errorDe(w, pool.EnTienda(r.Context(), id, func(t pgx.Tx) error {
			tag, err := t.Exec(r.Context(), `UPDATE tiendas SET cancelada_en = NULL WHERE id = $1 AND cancelada_en IS NOT NULL`, id)
			if err != nil || tag.RowsAffected() == 0 {
				return err
			}
			return auditoria.Registrar(r.Context(), t, id, auditoria.Registro{Accion: "plan.reactivar", Entidad: "tienda", EntidadID: id,
				Texto: "deshizo la cancelación del plan"})

		}))
	}))
	mux.HandleFunc("PUT /v1/panel/equipo/{id}", con(func(w http.ResponseWriter, r *http.Request, id string, q Quien) {
		var b struct{ Rol string }
		if leer(w, r, &b) {
			errorDe(w, CambiarRol(r.Context(), pool, id, q, r.PathValue("id"), b.Rol))
		}
	}))
	mux.HandleFunc("DELETE /v1/panel/equipo/{id}", con(func(w http.ResponseWriter, r *http.Request, id string, q Quien) {
		errorDe(w, Quitar(r.Context(), pool, id, q, r.PathValue("id")))
	}))
}

// pedirContacto: la dueña quiere que alguien del equipo le escriba antes de irse. Un aviso por día.
func pedirContacto(ctx context.Context, pool tx, g Ganchos, tiendaID, usuarioID string, motivo int) error {
	var nombre, slug, email string
	var whatsapp *string
	var reciente bool
	if err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		if err := t.QueryRow(ctx, `SELECT nombre, slug, coalesce(contacto_en > now() - interval '1 day', false) FROM tiendas WHERE id = $1`,
			tiendaID).Scan(&nombre, &slug, &reciente); err != nil {
			return err
		}
		return t.QueryRow(ctx, `SELECT email, telefono FROM auth.usuarios WHERE id = $1`, usuarioID).Scan(&email, &whatsapp)
	}); err != nil || reciente {
		return err
	}
	if g.Contacto != nil {
		motivos := []string{"Es caro para lo que vendo", "Me falta una función", "Voy a pausar el negocio",
			"Me voy a otra plataforma", "Otro motivo"}
		m := "Otro motivo"
		if motivo >= 0 && motivo < len(motivos) {
			m = motivos[motivo]
		}
		tel := "—"
		if whatsapp != nil && *whatsapp != "" {
			tel = *whatsapp
		}
		if err := g.Contacto(ctx, "Quiere irse: "+nombre,
			fmt.Sprintf("%s (%s.tulocalweb.com.ar) estaba por cancelar el plan y pidió que le escribamos.\n\nMotivo: %s\nCorreo: %s\nWhatsApp: %s\n\nLe prometimos escribirle hoy.",
				nombre, slug, m, email, tel)); err != nil {
			return err
		}
	}
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		if _, err := t.Exec(ctx, `UPDATE tiendas SET contacto_en = now() WHERE id = $1`, tiendaID); err != nil {
			return err
		}
		return auditoria.Registrar(ctx, t, tiendaID, auditoria.Registro{Accion: "plan.contacto", Entidad: "tienda", EntidadID: tiendaID,
			Texto: "pidió que alguien del equipo le escriba antes de cancelar"})
	})
}
