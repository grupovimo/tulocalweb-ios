// Package configuracion: la pantalla Configuración (design/PanelOperacion.dc.html). Negocio,
// equipo (membresías e invitaciones), seguridad y avisos, sesiones de la persona y plan.
package configuracion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/mayorista"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/roles"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/auditoria"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/geoip"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/limite"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

type tx interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
}

var (
	ErrNoExiste  = errors.New("no existe")
	ErrRepetido  = errors.New("ya está en el equipo")
	ErrNoSePuede = errors.New("no se puede")
	ErrInvalido  = errors.New("inválido")
	ErrSinEmail  = errors.New("la invitación no tiene email")
	ErrDuenia    = errors.New("a la dueña no se la quita ni se le cambia el rol")
	ErrVos       = errors.New("no podés cambiarte a vos")
	avisosDef    = map[string]bool{"venta": true, "chats": true, "stock": true, "diario": true, "ideas": true, "novedades": false}
	planes       = []string{"gratis", "emprende", "pro", "negocio"}
	emailValido  = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
)

// Negocio es components/schemas/Negocio.
type Negocio struct {
	Nombre      string `json:"nombre"`
	RazonSocial string `json:"razonSocial"`
	Cuit        string `json:"cuit"`
	Condicion   string `json:"condicion"`
	Direccion   string `json:"direccion"`
	Whatsapp    string `json:"whatsapp"`
	Email       string `json:"email"`
	Horario     string `json:"horario"`
	// Datos para cobrar por transferencia (los ve quien compra). Sin alias ni CBU, la tienda online
	// no ofrece transferencia.
	Alias   string `json:"alias"`
	CBU     string `json:"cbu"`
	Titular string `json:"titular"`
	// Política de cambios y devoluciones que se muestra en la tienda (vacía: el texto legal base).
	Politica string `json:"politica"`
}

var aliasValido = regexp.MustCompile(`^[a-zA-Z0-9.-]{6,20}$`)

// CBUValido comprueba los dos dígitos verificadores del CBU/CVU (22 dígitos).
func CBUValido(c string) bool {
	if len(c) != 22 {
		return false
	}
	for _, r := range c {
		if r < '0' || r > '9' {
			return false
		}
	}
	d := func(i int) int { return int(c[i] - '0') }
	verif := func(desde int, pesos []int) int {
		s := 0
		for i, p := range pesos {
			s += d(desde+i) * p
		}
		return (10 - s%10) % 10
	}
	return verif(0, []int{7, 1, 3, 9, 7, 1, 3}) == d(7) &&
		verif(8, []int{3, 9, 7, 1, 3, 9, 7, 1, 3, 9, 7, 1, 3}) == d(21)
}

// Miembro es components/schemas/Miembro.
type Miembro struct {
	ID        string     `json:"id"`
	Nombre    string     `json:"nombre"`
	Contacto  string     `json:"contacto"`
	Rol       string     `json:"rol"`
	Foto      *string    `json:"foto"`
	Ultimo    *time.Time `json:"ultimo"`
	Pendiente bool       `json:"pendiente"`
	// Vence: invitación pendiente, cuándo vence su link (nil si nunca se mandó).
	Vence       *time.Time `json:"vence"`
	Vos         bool       `json:"vos"`
	DobleFactor bool       `json:"dobleFactor"`
}

// Sesion es una sesión vigente de la persona que mira (components/schemas/Configuracion.sesiones).
type Sesion struct {
	ID string `json:"id"`
	// Tipo: compu, celu o tablet (el ícono).
	Tipo string `json:"tipo"`
	// Nombre: navegador y sistema («Chrome en Mac»).
	Nombre string `json:"nombre"`
	// Ubicacion: «Ciudad, Provincia» o el país, según la IP. Vacía si no se sabe.
	Ubicacion string `json:"ubicacion"`
	// IP abreviada (vacía si no hay).
	IP string `json:"ip"`
	// Creada: cuándo se abrió; Cuando: la última vez que se usó (se anota como mucho una vez por hora).
	Creada time.Time `json:"creada"`
	Cuando time.Time `json:"cuando"`
	Actual bool      `json:"actual"`
}

// Quien es la persona que mira (para "vos", sus sesiones y lo que su rol puede ver).
type Quien struct{ UsuarioID, SesionID, Rol string }

func conDefecto(def, v map[string]bool) map[string]bool {
	out := map[string]bool{}
	for k, d := range def {
		out[k] = d
		if x, ok := v[k]; ok {
			out[k] = x
		}
	}
	return out
}

// sesiones: todas las vigentes de la persona, la actual primero. Con «Cerrar sesiones inactivas»
// prendido no aparecen las que ya no valen (cuentas.Sesion las borra al próximo uso).
func sesiones(ctx context.Context, t pgx.Tx, geo geoip.Ubicador, q Quien) ([]Sesion, error) {
	out := []Sesion{}
	filas, err := t.Query(ctx, `SELECT left(s.token_hash, 16), coalesce(s.agente, ''), coalesce(host(s.ip), ''), s.creada_en, s.visto_en
		FROM auth.sesiones s JOIN auth.usuarios u ON u.id = s.usuario_id
		WHERE s.usuario_id::text = $1 AND s.vence_en > now()
			AND NOT (coalesce((u.seguridad->>'inactivas')::boolean, false) AND s.visto_en < now() - $2::interval)
		ORDER BY s.visto_en DESC`, q.UsuarioID, cuentas.Inactividad.String())
	if err != nil {
		return nil, err
	}
	defer filas.Close()
	for filas.Next() {
		var s Sesion
		var agente, ip string
		if err := filas.Scan(&s.ID, &agente, &ip, &s.Creada, &s.Cuando); err != nil {
			return nil, err
		}
		a := cuentas.Dispositivo(agente)
		s.Tipo, s.Nombre, s.IP, s.Actual = a.Tipo, a.Nombre(), cuentas.IPCorta(ip), s.ID == q.SesionID
		if geo != nil {
			s.Ubicacion = geo.Ubicar(ip)
		}
		out = append(out, s)
	}
	if err := filas.Err(); err != nil {
		return nil, err
	}
	slices.SortStableFunc(out, func(a, b Sesion) int {
		switch {
		case a.Actual == b.Actual:
			return 0
		case a.Actual:
			return -1
		}
		return 1
	})
	return out, nil
}

// Leer arma toda la configuración. geo ubica las sesiones (nil: sin lugar).
func Leer(ctx context.Context, pool tx, geo geoip.Ubicador, tiendaID string, q Quien, ahora time.Time) (map[string]any, error) {
	var res map[string]any
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var nombre, plan, ciclo string
		var negocio Negocio
		var prefs struct {
			Avisos map[string]bool `json:"avisos"`
		}
		var creada time.Time
		if err := t.QueryRow(ctx, `SELECT nombre, negocio, preferencias, plan, plan_ciclo, creada_en FROM tiendas WHERE id = $1`, tiendaID).
			Scan(&nombre, &negocio, &prefs, &plan, &ciclo, &creada); err != nil {
			return err
		}
		negocio.Nombre = nombre
		equipo := []Miembro{}
		filas, err := t.Query(ctx, `SELECT u.id::text, u.nombre, u.email, m.rol,
				(SELECT max(visto_en) FROM auth.sesiones s WHERE s.usuario_id = u.id)
			FROM auth.membresias m JOIN auth.usuarios u ON u.id = m.usuario_id
			WHERE m.tienda_id = $1 ORDER BY m.creada_en`, tiendaID)
		if err != nil {
			return err
		}
		for filas.Next() {
			var m Miembro
			if err := filas.Scan(&m.ID, &m.Nombre, &m.Contacto, &m.Rol, &m.Ultimo); err != nil {
				return err
			}
			m.Vos = m.ID == q.UsuarioID
			equipo = append(equipo, m)
		}
		if err := filas.Err(); err != nil {
			return err
		}
		inv, err := t.Query(ctx, `SELECT i.id::text, i.contacto, i.rol, l.vence_en
			FROM invitaciones i LEFT JOIN auth.invitaciones l ON l.invitacion_id = i.id
			WHERE i.tienda_id = $1 ORDER BY i.creada_en`, tiendaID)
		if err != nil {
			return err
		}
		for inv.Next() {
			var m Miembro
			if err := inv.Scan(&m.ID, &m.Contacto, &m.Rol, &m.Vence); err != nil {
				return err
			}
			m.Nombre, m.Pendiente = m.Contacto, true
			equipo = append(equipo, m)
		}
		if err := inv.Err(); err != nil {
			return err
		}
		may, err := mayorista.Leer(ctx, t, tiendaID)
		if err != nil {
			return err
		}
		var seguridad map[string]bool
		if err := t.QueryRow(ctx, `SELECT seguridad FROM auth.usuarios WHERE id::text = $1`, q.UsuarioID).Scan(&seguridad); err != nil &&
			!errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		ses, err := sesiones(ctx, t, geo, q)
		if err != nil {
			return err
		}
		// Lo que el rol no puede ver no sale: los datos del negocio (alias y CBU) sólo con Tienda; el
		// equipo, sólo la dueña (las demás se ven a sí mismas).
		if !roles.Puede(q.Rol, roles.Tienda) {
			negocio = Negocio{Nombre: nombre}
		}
		if !roles.Puede(q.Rol, roles.Cuenta) {
			equipo = slices.DeleteFunc(equipo, func(m Miembro) bool { return !m.Vos })
		}
		prueba := creada.Add(14 * 24 * time.Hour)
		res = map[string]any{
			"negocio":   negocio,
			"equipo":    equipo,
			"seguridad": conDefecto(cuentas.SeguridadDef, seguridad),
			"sesiones":  ses,
			"avisos":    conDefecto(avisosDef, prefs.Avisos),
			"mayorista": may,
			"plan": map[string]any{"actual": plan, "ciclo": ciclo, "enPrueba": ahora.Before(prueba),
				"diasPrueba": max(0, int(prueba.Sub(ahora).Hours()/24))},
		}
		return nil
	})
	return res, err
}

// GuardarNegocio guarda los datos (el nombre va en la columna de la tienda).
func GuardarNegocio(ctx context.Context, pool tx, tiendaID string, n Negocio) error {
	n.Nombre = strings.TrimSpace(n.Nombre)
	n.Alias, n.CBU, n.Titular = strings.TrimSpace(n.Alias), strings.ReplaceAll(strings.TrimSpace(n.CBU), " ", ""), strings.TrimSpace(n.Titular)
	n.Politica = strings.TrimSpace(n.Politica)
	if utf8.RuneCountInString(n.Politica) > 1000 {
		return ErrInvalido
	}
	if n.Nombre == "" || len(n.Nombre) > 80 || (n.Alias != "" && !aliasValido.MatchString(n.Alias)) ||
		(n.CBU != "" && !CBUValido(n.CBU)) || len(n.Titular) > 80 {
		return ErrInvalido
	}
	resto := n
	resto.Nombre = ""
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var antes Negocio
		if err := t.QueryRow(ctx, `SELECT nombre, coalesce(negocio->>'alias', ''), coalesce(negocio->>'cbu', ''),
				coalesce(negocio->>'titular', '') FROM tiendas WHERE id = $1 FOR UPDATE`, tiendaID).
			Scan(&antes.Nombre, &antes.Alias, &antes.CBU, &antes.Titular); err != nil {
			return err
		}
		if _, err := t.Exec(ctx, `UPDATE tiendas SET nombre = $2, negocio = $3 WHERE id = $1`, tiendaID, n.Nombre, resto); err != nil {
			return err
		}
		return auditoria.Registrar(ctx, t, tiendaID, registrosNegocio(tiendaID, antes, n)...)
	})
}

// registrosNegocio: lo que se anota en la Actividad al guardar el negocio. El CBU nunca va entero:
// sólo sus últimos 4 números.
func registrosNegocio(tiendaID string, a, n Negocio) []auditoria.Registro {
	var rs []auditoria.Registro
	reg := func(accion, texto string) {
		rs = append(rs, auditoria.Registro{Accion: accion, Entidad: "tienda", EntidadID: tiendaID, Texto: texto})
	}
	if a.Nombre != n.Nombre {
		reg("negocio.nombre", fmt.Sprintf("cambió el nombre de la tienda de «%s» a «%s»", a.Nombre, n.Nombre))
	}
	cambio := func(campo, antes, despues, mostrarAntes, mostrarDespues string) {
		switch {
		case antes == despues:
		case despues == "":
			reg("cobros."+strings.ToLower(campo), "sacó el "+campo+" para cobrar por transferencia (era "+mostrarAntes+")")
		case antes == "":
			reg("cobros."+strings.ToLower(campo), "cargó el "+campo+" para cobrar por transferencia: "+mostrarDespues)
		default:
			reg("cobros."+strings.ToLower(campo), "cambió el "+campo+" para cobrar por transferencia de "+mostrarAntes+" a "+mostrarDespues)
		}
	}
	cambio("alias", a.Alias, n.Alias, a.Alias, n.Alias)
	cambio("CBU", a.CBU, n.CBU, "uno "+auditoria.Final(a.CBU), "uno "+auditoria.Final(n.CBU))
	cambio("titular", a.Titular, n.Titular, a.Titular, n.Titular)
	return rs
}

// FijarOpcion prende o apaga una opción de seguridad (de la persona: cuentas.SeguridadDef) o un
// aviso (de la tienda). Prender «inactivas» cierra en el momento las sesiones que ya no valdrían.
func FijarOpcion(ctx context.Context, pool tx, tiendaID string, q Quien, grupo, clave string, activo bool) error {
	def := map[string]map[string]bool{"seguridad": cuentas.SeguridadDef, "avisos": avisosDef}[grupo]
	if _, ok := def[clave]; !ok {
		return ErrNoExiste
	}
	if grupo == "seguridad" {
		return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
			if _, err := t.Exec(ctx, `UPDATE auth.usuarios SET seguridad = seguridad || jsonb_build_object($2::text, $3::boolean)
				WHERE id::text = $1`, q.UsuarioID, clave, activo); err != nil {
				return err
			}
			if clave == "inactivas" && activo {
				_, err := t.Exec(ctx, `DELETE FROM auth.sesiones WHERE usuario_id::text = $1 AND left(token_hash, 16) <> $2
					AND visto_en < now() - $3::interval`, q.UsuarioID, q.SesionID, cuentas.Inactividad.String())
				return err
			}
			return nil
		})
	}
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `UPDATE tiendas SET preferencias = jsonb_set(preferencias, ARRAY[$2::text],
			coalesce(preferencias->$2, '{}') || jsonb_build_object($3::text, $4::boolean)) WHERE id = $1`, tiendaID, grupo, clave, activo)
		return err
	})
}

// CerrarSesion cierra otra sesión de la misma persona.
func CerrarSesion(ctx context.Context, pool tx, tiendaID string, q Quien, id string) error {
	if id == q.SesionID {
		return ErrNoSePuede
	}
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		tag, err := t.Exec(ctx, `DELETE FROM auth.sesiones WHERE usuario_id::text = $1 AND left(token_hash, 16) = $2`, q.UsuarioID, id)
		if err == nil && tag.RowsAffected() == 0 {
			return ErrNoExiste
		}
		if err != nil {
			return err
		}
		return auditoria.Registrar(ctx, t, tiendaID, auditoria.Registro{Accion: "sesion.cerrar", Entidad: "sesion",
			Texto: "cerró una de sus sesiones abiertas en otro dispositivo"})
	})
}

// CerrarOtras cierra todas las sesiones de la persona menos la actual. Devuelve cuántas cerró.
func CerrarOtras(ctx context.Context, pool tx, tiendaID string, q Quien) (int64, error) {
	if q.SesionID == "" || q.UsuarioID == "" {
		return 0, ErrNoSePuede // sólo desde una sesión de la persona (no una IA ni una mirada del equipo)
	}
	var n int64
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		tag, err := t.Exec(ctx, `DELETE FROM auth.sesiones WHERE usuario_id::text = $1 AND left(token_hash, 16) <> $2`, q.UsuarioID, q.SesionID)
		n = tag.RowsAffected()
		if err != nil || n == 0 {
			return err
		}
		return auditoria.Registrar(ctx, t, tiendaID, auditoria.Registro{Accion: "sesion.cerrar_otras", Entidad: "sesion",
			Texto: fmt.Sprintf("cerró sus otras sesiones abiertas (%d)", n)})
	})
	return n, err
}

// Invitar suma una invitación por email y encola el mail con el link (un solo uso, 7 días). Quien
// invita queda anotada (sale en el mail).
func Invitar(ctx context.Context, pool tx, sitio, tiendaID string, q Quien, contacto, rol string) (Miembro, error) {
	contacto = strings.ToLower(strings.TrimSpace(contacto))
	m := Miembro{Nombre: contacto, Contacto: contacto, Rol: rol, Pendiente: true}
	if len(contacto) < 6 || len(contacto) > 120 || !emailValido.MatchString(contacto) || !roles.Invitable(rol) {
		return m, ErrInvalido
	}
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var ya bool
		if err := t.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM auth.membresias mm JOIN auth.usuarios u ON u.id = mm.usuario_id
			WHERE mm.tienda_id = $1 AND u.email = $2)`, tiendaID, contacto).Scan(&ya); err != nil {
			return err
		}
		if ya {
			return ErrRepetido
		}
		err := t.QueryRow(ctx, `INSERT INTO invitaciones (tienda_id, contacto, rol, invitada_por, enviada_en)
			VALUES ($1, $2, $3, nullif($4, '')::uuid, now())
			ON CONFLICT (tienda_id, contacto) DO NOTHING RETURNING id::text`, tiendaID, contacto, rol, q.UsuarioID).Scan(&m.ID)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrRepetido
		}
		if err != nil {
			return err
		}
		if m.Vence, err = mandar(ctx, t, sitio, tiendaID, m.ID, contacto, rol, q); err != nil {
			return err
		}
		return auditoria.Registrar(ctx, t, tiendaID, auditoria.Registro{Accion: "personas.invitar", Entidad: "invitacion",
			EntidadID: m.ID, Texto: fmt.Sprintf("invitó a %s al equipo como %s", contacto, roles.Nombre[rol]),
			Despues: map[string]string{"rol": rol}})
	})
	return m, err
}

// mandar arma el link nuevo de la invitación y encola el mail. Devuelve cuándo vence.
func mandar(ctx context.Context, t pgx.Tx, sitio, tiendaID, id, para, rol string, q Quien) (*time.Time, error) {
	var tienda, invita string
	if err := t.QueryRow(ctx, `SELECT nombre, coalesce((SELECT nombre FROM auth.usuarios WHERE id::text = $2), '')
		FROM tiendas WHERE id = $1`, tiendaID, q.UsuarioID).Scan(&tienda, &invita); err != nil {
		return nil, err
	}
	if err := cuentas.EncolarInvitacion(ctx, t, sitio, tiendaID, id,
		cuentas.MailInvitacion{Para: para, Tienda: tienda, Invita: invita, Rol: rol}); err != nil {
		return nil, err
	}
	var v time.Time
	err := t.QueryRow(ctx, `SELECT vence_en FROM auth.invitaciones WHERE invitacion_id = $1`, id).Scan(&v)
	return &v, err
}

// Reenviar manda otra vez la invitación con un link nuevo (el anterior deja de valer).
func Reenviar(ctx context.Context, pool tx, sitio, tiendaID string, q Quien, id string) error {
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var contacto, rol string
		err := t.QueryRow(ctx, `UPDATE invitaciones SET enviada_en = now() WHERE tienda_id = $1 AND id::text = $2
			RETURNING contacto, rol`, tiendaID, id).Scan(&contacto, &rol)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoExiste
		}
		if err != nil {
			return err
		}
		if !emailValido.MatchString(contacto) {
			return ErrSinEmail // una invitación vieja a un WhatsApp: no hay a dónde mandarla
		}
		if _, err = mandar(ctx, t, sitio, tiendaID, id, strings.ToLower(contacto), rol, q); err != nil {
			return err
		}
		return auditoria.Registrar(ctx, t, tiendaID, auditoria.Registro{Accion: "personas.reenviar", Entidad: "invitacion",
			EntidadID: id, Texto: "le volvió a mandar la invitación a " + strings.ToLower(contacto)})
	})
}

// permitidoInvitar: hasta 30 mails de invitación por tienda por hora (cada uno es un mail que sale).
func permitidoInvitar(ctx context.Context, pool tx, tiendaID string) error {
	c, ok := pool.(limite.Consultor)
	if !ok {
		return nil
	}
	ok, err := limite.Permitido(ctx, c, "invitar:tienda:"+tiendaID, 30, time.Hour)
	if err != nil {
		return err
	}
	if !ok {
		return cuentas.ErrDemasiadosIntentos
	}
	return nil
}

// CambiarPlan guarda el plan elegido. El cobro llega con Mercado Pago Suscripciones.
func CambiarPlan(ctx context.Context, pool tx, tiendaID, plan, ciclo string) error {
	if !slices.Contains(planes, plan) || (ciclo != "" && ciclo != "mes" && ciclo != "anio") {
		return ErrInvalido
	}
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var antes string
		if err := t.QueryRow(ctx, `SELECT plan FROM tiendas WHERE id = $1 FOR UPDATE`, tiendaID).Scan(&antes); err != nil {
			return err
		}
		if _, err := t.Exec(ctx, `UPDATE tiendas SET plan = $2, plan_ciclo = coalesce(nullif($3, ''), plan_ciclo) WHERE id = $1`, tiendaID, plan, ciclo); err != nil {
			return err
		}
		if antes == plan {
			return nil
		}
		return auditoria.Registrar(ctx, t, tiendaID, auditoria.Registro{Accion: "plan.cambiar", Entidad: "tienda", EntidadID: tiendaID,
			Texto: fmt.Sprintf("cambió el plan de %s a %s", nombrePlan(antes), nombrePlan(plan)),
			Antes: map[string]string{"plan": antes}, Despues: map[string]string{"plan": plan}})
	})
}

func nombrePlan(p string) string {
	if p == "" {
		return p
	}
	return strings.ToUpper(p[:1]) + p[1:]
}

// Rutas del panel. cobraAbono: con Mercado Pago activo, un plan pago sólo se activa por la
// suscripción (POST /v1/panel/cuenta/suscripcion); acá sólo se puede bajar al gratis.
func Rutas(mux *http.ServeMux, pool tx, geo geoip.Ubicador, sitio string, cobraAbono bool) {
	con := func(h func(http.ResponseWriter, *http.Request, string, Quien)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			d, ok := cuentas.SesionDe(r)
			if !ok || d.TiendaID == "" {
				problema.Escribir(w, problema.SinSesion())
				return
			}
			h(w, r, d.TiendaID, Quien{UsuarioID: d.Usuario.ID, SesionID: d.SesionID, Rol: d.Rol})
		}
	}
	errorDe := func(w http.ResponseWriter, err error) {
		switch {
		case err == nil:
			w.WriteHeader(http.StatusNoContent)
		case errors.Is(err, ErrNoExiste):
			problema.Escribir(w, problema.NoEncontrado("No existe."))
		case errors.Is(err, ErrRepetido):
			problema.Escribir(w, problema.Conflicto("repetido", "Esa persona ya está en el equipo."))
		case errors.Is(err, ErrDuenia):
			problema.Escribir(w, problema.Conflicto("duenia", "A la dueña no se la quita ni se le cambia el rol."))
		case errors.Is(err, ErrVos):
			problema.Escribir(w, problema.Conflicto("vos", "No podés cambiar tu propio rol ni quitarte del equipo."))
		case errors.Is(err, ErrSinEmail):
			problema.Escribir(w, problema.NoProcesable("sin_email", "Esa invitación era por WhatsApp y no se puede mandar: cancelala e invitá por email."))
		case errors.Is(err, cuentas.ErrDemasiadosIntentos):
			w.Header().Set("Retry-After", "3600")
			problema.Escribir(w, &problema.Problema{Estado: http.StatusTooManyRequests, Codigo: "demasiados_intentos",
				Titulo: "Esperá un rato", Detalle: "Mandaste muchas invitaciones seguidas. Probá de nuevo en una hora."})
		case errors.Is(err, ErrNoSePuede):
			problema.Escribir(w, problema.Conflicto("actual", "Para cerrar esta, salí de la cuenta."))
		case errors.Is(err, ErrInvalido):
			problema.Escribir(w, problema.Invalido("Revisá los datos."))
		default:
			problema.Escribir(w, problema.Interno())
		}
	}
	leer := func(w http.ResponseWriter, r *http.Request, v any) bool {
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10)).Decode(v); err != nil {
			problema.Escribir(w, problema.Invalido("Revisá los datos."))
			return false
		}
		return true
	}
	rutasCuenta(mux, pool, con, errorDe, leer)
	rutasDatos(mux, pool, sitio, con, errorDe, leer)
	rutasMedicion(mux, pool, con, errorDe, leer)
	rutasEntregas(mux, pool, con, errorDe, leer)
	rutasMayorista(mux, pool, con, errorDe, leer)
	mux.HandleFunc("GET /v1/panel/configuracion", con(func(w http.ResponseWriter, r *http.Request, id string, q Quien) {
		v, err := Leer(r.Context(), pool, geo, id, q, time.Now())
		if err != nil {
			errorDe(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(v)
	}))
	mux.HandleFunc("PUT /v1/panel/configuracion/negocio", con(func(w http.ResponseWriter, r *http.Request, id string, _ Quien) {
		var n Negocio
		if leer(w, r, &n) {
			errorDe(w, GuardarNegocio(r.Context(), pool, id, n))
		}
	}))
	mux.HandleFunc("PUT /v1/panel/configuracion/{grupo}/{clave}", con(func(w http.ResponseWriter, r *http.Request, id string, q Quien) {
		var b struct {
			Activo bool `json:"activo"`
		}
		if leer(w, r, &b) {
			errorDe(w, FijarOpcion(r.Context(), pool, id, q, r.PathValue("grupo"), r.PathValue("clave"), b.Activo))
		}
	}))
	mux.HandleFunc("DELETE /v1/panel/sesiones/{id}", con(func(w http.ResponseWriter, r *http.Request, id string, q Quien) {
		errorDe(w, CerrarSesion(r.Context(), pool, id, q, r.PathValue("id")))
	}))
	mux.HandleFunc("DELETE /v1/panel/sesiones", con(func(w http.ResponseWriter, r *http.Request, id string, q Quien) {
		n, err := CerrarOtras(r.Context(), pool, id, q)
		if err != nil {
			errorDe(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]int64{"cerradas": n})
	}))
	mux.HandleFunc("POST /v1/panel/equipo", con(func(w http.ResponseWriter, r *http.Request, id string, q Quien) {
		var b struct{ Contacto, Rol string }
		if !leer(w, r, &b) {
			return
		}
		if err := permitidoInvitar(r.Context(), pool, id); err != nil {
			errorDe(w, err)
			return
		}
		m, err := Invitar(r.Context(), pool, sitio, id, q, b.Contacto, b.Rol)
		if err != nil {
			errorDe(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(m)
	}))
	mux.HandleFunc("POST /v1/panel/equipo/{id}/reenviar", con(func(w http.ResponseWriter, r *http.Request, id string, q Quien) {
		if err := permitidoInvitar(r.Context(), pool, id); err != nil {
			errorDe(w, err)
			return
		}
		errorDe(w, Reenviar(r.Context(), pool, sitio, id, q, r.PathValue("id")))
	}))
	mux.HandleFunc("PUT /v1/panel/plan", con(func(w http.ResponseWriter, r *http.Request, id string, _ Quien) {
		var b struct{ Plan, Ciclo string }
		if !leer(w, r, &b) {
			return
		}
		if cobraAbono && b.Plan != "gratis" {
			problema.Escribir(w, problema.Conflicto("pagar", "Los planes pagos se activan con la suscripción de Mercado Pago."))
			return
		}
		errorDe(w, CambiarPlan(r.Context(), pool, id, b.Plan, b.Ciclo))
	}))
}
