// Package cuentas: registro, ingreso con contraseña, sesiones y recuperación por mail.
//
// Las respuestas no delatan quién tiene cuenta: un email inexistente tarda lo mismo y contesta igual
// que una contraseña incorrecta, y "recuperar" responde siempre que salió.
package cuentas

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/tiendas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/clerk"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/geoip"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/limite"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/whatsapp"
)

// Errores de dominio. El handler los traduce a problem+json.
var (
	ErrCredenciales       = errors.New("email o contraseña incorrectos")
	ErrEmailEnUso         = errors.New("ese email ya tiene cuenta")
	ErrDemasiadosIntentos = errors.New("demasiados intentos")
	ErrTokenInvalido      = errors.New("el link venció o ya se usó")
	// ErrOtroEmail: la invitación es para otro email que el de la cuenta que la quiere usar.
	ErrOtroEmail = errors.New("la invitación es para otro email")
	ErrSinSesion = errors.New("sin sesión")
)

// MaxIntentosIngreso: intentos con contraseña por email cada 15 minutos; después se pausa el ingreso.
const MaxIntentosIngreso = 5

// ErrClave es ErrCredenciales con los intentos que quedan antes de pausar el ingreso de ese email.
type ErrClave struct{ Restantes int }

func (e ErrClave) Error() string          { return ErrCredenciales.Error() }
func (e ErrClave) Is(objetivo error) bool { return objetivo == ErrCredenciales }

// ErrInvalido es un dato mal cargado, con el campo.
type ErrInvalido struct{ Campo, Mensaje string }

func (e ErrInvalido) Error() string { return e.Campo + ": " + e.Mensaje }

// Duraciones.
const (
	SesionCorta     = 12 * time.Hour
	SesionLarga     = 365 * 24 * time.Hour
	VidaRestablecer = 30 * time.Minute
)

// Servicio de cuentas.
type Servicio struct {
	Pool   *db.Pool
	Correo correo.Correo
	// URL pública del sitio, para los links de los mails.
	Sitio string
	// Ingreso con Mercado Pago y Mercado Libre (Proveedores); los apagados no tienen ClientID.
	OAuth map[string]*Proveedor
	// Clerk: Google, Apple y los demás ingresos sociales de Grupo Vimo (web y app de iPhone). nil:
	// apagado.
	Clerk *clerk.Cliente
	// WhatsApp de la plataforma: links de ingreso y de contraseña nueva. nil: no se manda.
	WhatsApp whatsapp.Mensajero
	// Cliente HTTP para los proveedores (tests: el del servidor falso).
	HTTP *http.Client
	// SesionIA resuelve los tokens de las IA conectadas (paquete ia). nil: no se aceptan.
	SesionIA func(ctx context.Context, token string) (Datos, error)
	// SesionEquipo resuelve los tokens de "ver la tienda" del backoffice (paquete equipo): una
	// sesión de sólo lectura del panel de esa tienda. nil: no se aceptan.
	SesionEquipo func(ctx context.Context, token string) (Datos, error)
	// Geo ubica la IP de cada sesión (lista de sesiones y mail de ingreso). nil: sin lugar.
	Geo geoip.Ubicador
}

// Cliente es de dónde viene el pedido (para la sesión y los límites).
type Cliente struct{ IP, Agente string }

// Sesion recién creada: el token se le da al BFF una sola vez.
type Sesion struct {
	Token    string    `json:"token"`
	VenceEn  time.Time `json:"venceEn"`
	TiendaID string    `json:"tiendaId"`
}

// Alta de cuenta (registro).
type Alta struct {
	Nombre, Tienda, Email, Clave string
}

func normalizarEmail(e string) (string, error) {
	e = strings.ToLower(strings.TrimSpace(e))
	if e == "" {
		return "", ErrInvalido{"email", "Escribí tu email."}
	}
	a, err := mail.ParseAddress(e)
	if err != nil || a.Address != e || len(e) > 254 || !strings.Contains(e[strings.LastIndex(e, "@"):], ".") {
		return "", ErrInvalido{"email", "Ese email no parece válido."}
	}
	return e, nil
}

func validarClave(c string) error {
	if utf8.RuneCountInString(c) < 8 {
		return ErrInvalido{"clave", "La contraseña tiene que tener al menos 8 caracteres."}
	}
	if len(c) > 200 {
		return ErrInvalido{"clave", "La contraseña es demasiado larga."}
	}
	return nil
}

func texto(campo, v, vacio string) (string, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return "", ErrInvalido{campo, vacio}
	}
	if utf8.RuneCountInString(v) > 80 {
		return "", ErrInvalido{campo, "Es demasiado largo."}
	}
	return v, nil
}

// Registrar crea la cuenta, su tienda (rubro y plantilla por defecto: los ajusta el onboarding) y
// la sesión, todo en una transacción.
func (s *Servicio) Registrar(ctx context.Context, a Alta, cl Cliente) (Sesion, error) {
	nombre, err := texto("nombre", a.Nombre, "¿Cómo te llamás?")
	if err != nil {
		return Sesion{}, err
	}
	// El nombre de la tienda se elige en el onboarding: si no viene, uno provisorio como con Google.
	tienda := strings.TrimSpace(a.Tienda)
	if tienda == "" {
		tienda = "La tienda de " + strings.Fields(nombre)[0]
	}
	if utf8.RuneCountInString(tienda) > 80 {
		return Sesion{}, ErrInvalido{"tienda", "Es demasiado largo."}
	}
	email, err := normalizarEmail(a.Email)
	if err != nil {
		return Sesion{}, err
	}
	if err := validarClave(a.Clave); err != nil {
		return Sesion{}, err
	}
	if ok, err := s.permitido(ctx, "registro:ip:"+cl.IP, 10, time.Hour); err != nil || !ok {
		return Sesion{}, errOr(err, ErrDemasiadosIntentos)
	}
	hash, err := HashClave(a.Clave)
	if err != nil {
		return Sesion{}, err
	}
	var ses Sesion
	err = s.Pool.EnTx(ctx, func(tx pgx.Tx) error {
		usuario, tiendaID, err := crearCuenta(ctx, tx, email, nombre, &hash, tienda)
		if err != nil {
			return err
		}
		ses, err = crearSesion(ctx, tx, usuario, tiendaID, SesionLarga, cl)
		return err
	})
	if err == nil {
		s.bienvenida(ctx, email, nombre, tienda)
	}
	return ses, err
}

// bienvenida manda el mail de cuenta nueva. Si el correo falla, la cuenta igual queda creada: se
// anota y se sigue (no vale la pena frenar el registro por eso).
func (s *Servicio) bienvenida(ctx context.Context, email, nombre, tienda string) {
	if s.Correo == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 8*time.Second)
	defer cancel()
	if err := s.Correo.Enviar(ctx, mailBienvenida(email, nombre, tienda, s.Sitio+"/onboarding")); err != nil {
		slog.Warn("cuentas: no salió el mail de bienvenida", "error", err)
	}
}

// VerificarRestablecer dice si el link de contraseña nueva todavía sirve, sin usarlo: la pantalla
// avisa que venció o que ya se usó antes de pedir la contraseña.
func (s *Servicio) VerificarRestablecer(ctx context.Context, token string) error {
	if token == "" {
		return ErrTokenInvalido
	}
	var ok bool
	if err := s.Pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM auth.restablecimientos
		WHERE token_hash = $1 AND usado_en IS NULL AND vence_en > now())`, hashToken(token)).Scan(&ok); err != nil {
		return err
	}
	if !ok {
		return ErrTokenInvalido
	}
	return nil
}

// crearCuenta da de alta a la persona, su tienda (rubro y plantilla por defecto: los ajusta el
// onboarding) y la membresía de dueña. Sin clave (hash nil) es una cuenta que entra con un
// proveedor o con link.
func crearCuenta(ctx context.Context, tx pgx.Tx, email, nombre string, hash *string, tienda string) (string, string, error) {
	var usuario string
	err := tx.QueryRow(ctx, `INSERT INTO auth.usuarios (email, nombre, clave_hash) VALUES ($1, $2, $3) RETURNING id`,
		email, nombre, hash).Scan(&usuario)
	var pg *pgconn.PgError
	if errors.As(err, &pg) && pg.Code == "23505" {
		return "", "", ErrEmailEnUso
	}
	if err != nil {
		return "", "", err
	}
	slug, err := slugLibre(ctx, tx, tiendas.Slug(tienda))
	if err != nil {
		return "", "", err
	}
	r, _ := tiendas.BuscarRubro(tiendas.RubroPorDefecto)
	var tiendaID string
	if err := tx.QueryRow(ctx, "SELECT app.crear_tienda($1, $2, $3, $4)", slug, tienda, r.ID, r.Plantilla).Scan(&tiendaID); err != nil {
		return "", "", err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO auth.membresias (usuario_id, tienda_id, rol) VALUES ($1, $2, 'duenio')`, usuario, tiendaID); err != nil {
		return "", "", err
	}
	return usuario, tiendaID, nil
}

// slugLibre agrega -2, -3… hasta encontrar uno libre.
func slugLibre(ctx context.Context, tx pgx.Tx, base string) (string, error) {
	if base == "" {
		base = "tienda"
	}
	for i := 1; i < 100; i++ {
		cand := base
		if i > 1 {
			suf := fmt.Sprintf("-%d", i)
			cand = strings.TrimRight(base[:min(len(base), 48-len(suf))], "-") + suf
		}
		var libre bool
		if err := tx.QueryRow(ctx, "SELECT app.slug_disponible($1)", cand).Scan(&libre); err != nil {
			return "", err
		}
		if libre {
			return cand, nil
		}
	}
	return "", errors.New("no hay slug libre")
}

// Ingresar con email y contraseña.
func (s *Servicio) Ingresar(ctx context.Context, email, clave string, recordarme bool, cl Cliente) (Sesion, error) {
	email, err := normalizarEmail(email)
	if err != nil {
		return Sesion{}, err
	}
	if clave == "" {
		return Sesion{}, ErrInvalido{"clave", "Escribí tu contraseña."}
	}
	intentos, err := limite.Sumar(ctx, s.Pool, "ingreso:email:"+email, 15*time.Minute)
	if err != nil || intentos > MaxIntentosIngreso {
		return Sesion{}, errOr(err, ErrDemasiadosIntentos)
	}
	if ok, err := s.permitido(ctx, "ingreso:ip:"+cl.IP, 30, 15*time.Minute); err != nil || !ok {
		return Sesion{}, errOr(err, ErrDemasiadosIntentos)
	}
	// Con la contraseña mal se dice cuántos intentos quedan (igual exista o no la cuenta: no
	// delata qué emails tienen cuenta).
	claveMal := ErrClave{Restantes: MaxIntentosIngreso - intentos}
	var usuario string
	var hash *string
	err = s.Pool.QueryRow(ctx, `SELECT id, clave_hash FROM auth.usuarios WHERE email = $1`, email).Scan(&usuario, &hash)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && hash == nil) {
		_, _ = VerificarClave(clave, hashFalso) // mismo tiempo que con una cuenta real
		return Sesion{}, claveMal
	}
	if err != nil {
		return Sesion{}, err
	}
	ok, err := VerificarClave(clave, *hash)
	if err != nil {
		return Sesion{}, err
	}
	if !ok {
		return Sesion{}, claveMal
	}
	_, _ = s.Pool.Exec(ctx, `DELETE FROM auth.intentos WHERE clave = $1`, "ingreso:email:"+email)
	vida := SesionCorta
	if recordarme {
		vida = SesionLarga
	}
	return s.sesionDeUsuario(ctx, usuario, vida, cl)
}

// sesionDeUsuario abre una sesión de la persona en su primera tienda. Es el camino de todo ingreso a
// una cuenta que ya existe (clave, Google y Apple por Clerk, Mercado Pago, link por WhatsApp): en la
// misma transacción encola el mail «Entraron a tu cuenta». El registro no pasa por acá.
func (s *Servicio) sesionDeUsuario(ctx context.Context, usuario string, vida time.Duration, cl Cliente) (Sesion, error) {
	var ses Sesion
	err := s.Pool.EnTx(ctx, func(tx pgx.Tx) error {
		var tienda *string
		err := tx.QueryRow(ctx, `SELECT tienda_id FROM auth.membresias WHERE usuario_id = $1 ORDER BY creada_en LIMIT 1`, usuario).Scan(&tienda)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		t := ""
		if tienda != nil {
			t = *tienda
		}
		ses, err = crearSesion(ctx, tx, usuario, t, vida, cl)
		if err != nil {
			return err
		}
		return s.avisarIngreso(ctx, tx, usuario, hashToken(ses.Token), cl)
	})
	return ses, err
}

// Datos de la sesión en curso.
type Datos struct {
	Usuario struct {
		ID     string `json:"id"`
		Nombre string `json:"nombre"`
		Email  string `json:"email"`
		// Foto de perfil (la de Google o Apple al entrar con Clerk). Vacía: no tiene.
		Foto string `json:"foto,omitempty"`
	} `json:"usuario"`
	TiendaID string    `json:"tiendaId"`
	Rol      string    `json:"rol"`
	VenceEn  time.Time `json:"venceEn"`
	// SesionID identifica la sesión (un prefijo del hash del token) para listarla y cerrarla.
	SesionID string `json:"-"`
	// SesionHash: el hash entero del token, sólo en las sesiones de personas (no en las de una IA
	// ni del equipo). Ata un iPhone a esta sesión (push.Registrar).
	SesionHash string `json:"-"`
	// IA es la conexión de una IA (paquete ia) cuando el token es de un conector MCP; IAAcciones,
	// si esa conexión puede escribir. Las sesiones de personas los dejan vacíos.
	IA         string `json:"-"`
	IAAcciones bool   `json:"-"`
	// IANombre: el nombre de la IA conectada (Claude, ChatGPT…), para la Actividad.
	IANombre string `json:"-"`
	// Mirada: una persona del equipo viendo el panel como la dueña. Sólo lectura (equipo.SoloLectura).
	Mirada bool `json:"-"`
}

// Sesion resuelve el token. Actualiza "visto" como mucho una vez por hora (como en Caja: no escribir
// en cada pedido). Con «Cerrar sesiones inactivas» prendido, una sesión sin usar hace más de
// Inactividad ya no vale y se borra.
func (s *Servicio) Sesion(ctx context.Context, token string) (Datos, error) {
	var d Datos
	if token == "" {
		return d, ErrSinSesion
	}
	var tienda, rol *string
	var visto time.Time
	var cerrarInactivas bool
	err := s.Pool.QueryRow(ctx, `
		SELECT u.id, u.nombre, u.email, coalesce(u.foto, ''), s.tienda_id, m.rol, s.vence_en, s.visto_en,
			coalesce((u.seguridad->>'inactivas')::boolean, false)
		FROM auth.sesiones s
		JOIN auth.usuarios u ON u.id = s.usuario_id
		LEFT JOIN auth.membresias m ON m.usuario_id = s.usuario_id AND m.tienda_id = s.tienda_id
		WHERE s.token_hash = $1 AND s.vence_en > now()`, hashToken(token)).
		Scan(&d.Usuario.ID, &d.Usuario.Nombre, &d.Usuario.Email, &d.Usuario.Foto, &tienda, &rol, &d.VenceEn, &visto, &cerrarInactivas)
	if errors.Is(err, pgx.ErrNoRows) {
		return d, ErrSinSesion
	}
	if err != nil {
		return d, err
	}
	if cerrarInactivas && time.Since(visto) > Inactividad {
		_, _ = s.Pool.Exec(ctx, `DELETE FROM auth.sesiones WHERE token_hash = $1`, hashToken(token))
		return Datos{}, ErrSinSesion
	}
	// Sin membresía (se la quitaron) no hay tienda: la sesión sigue, pero sin acceso a datos.
	if tienda != nil && rol != nil {
		d.TiendaID, d.Rol = *tienda, *rol
	}
	d.SesionHash = hashToken(token)
	d.SesionID = d.SesionHash[:16]
	if time.Since(visto) > time.Hour {
		_, _ = s.Pool.Exec(ctx, `UPDATE auth.sesiones SET visto_en = now() WHERE token_hash = $1`, hashToken(token))
	}
	return d, nil
}

// Salir cierra la sesión.
func (s *Servicio) Salir(ctx context.Context, token string) error {
	_, err := s.Pool.Exec(ctx, `DELETE FROM auth.sesiones WHERE token_hash = $1`, hashToken(token))
	return err
}

// Recuperar manda el link para crear una contraseña nueva, si la cuenta existe. Nunca dice si existe.
func (s *Servicio) Recuperar(ctx context.Context, email string, cl Cliente) error {
	email, err := normalizarEmail(email)
	if err != nil {
		return err
	}
	if ok, err := s.permitido(ctx, "recuperar:email:"+email, 3, time.Hour); err != nil || !ok {
		return errOr(err, ErrDemasiadosIntentos)
	}
	if ok, err := s.permitido(ctx, "recuperar:ip:"+cl.IP, 20, time.Hour); err != nil || !ok {
		return errOr(err, ErrDemasiadosIntentos)
	}
	var usuario, nombre string
	err = s.Pool.QueryRow(ctx, `SELECT id, nombre FROM auth.usuarios WHERE email = $1`, email).Scan(&usuario, &nombre)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	token := nuevoToken()
	if _, err := s.Pool.Exec(ctx, `INSERT INTO auth.restablecimientos (token_hash, usuario_id, vence_en) VALUES ($1, $2, now() + $3::interval)`,
		hashToken(token), usuario, VidaRestablecer.String()); err != nil {
		return err
	}
	return s.Correo.Enviar(ctx, mailRecuperar(email, nombre, s.Sitio+"/ingresar?m=restablecer&token="+token))
}

// Restablecer cambia la contraseña con el link y cierra todas las sesiones de la cuenta.
func (s *Servicio) Restablecer(ctx context.Context, token, clave string) error {
	if err := validarClave(clave); err != nil {
		return err
	}
	hash, err := HashClave(clave)
	if err != nil {
		return err
	}
	return s.Pool.EnTx(ctx, func(tx pgx.Tx) error {
		var usuario string
		err := tx.QueryRow(ctx, `
			UPDATE auth.restablecimientos SET usado_en = now()
			WHERE token_hash = $1 AND usado_en IS NULL AND vence_en > now()
			RETURNING usuario_id`, hashToken(token)).Scan(&usuario)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrTokenInvalido
		}
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE auth.usuarios SET clave_hash = $2 WHERE id = $1`, usuario, hash); err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `DELETE FROM auth.sesiones WHERE usuario_id = $1`, usuario)
		return err
	})
}

func crearSesion(ctx context.Context, tx pgx.Tx, usuario, tienda string, vida time.Duration, cl Cliente) (Sesion, error) {
	token := nuevoToken()
	var vence time.Time
	err := tx.QueryRow(ctx, `
		INSERT INTO auth.sesiones (token_hash, usuario_id, tienda_id, vence_en, ip, agente)
		VALUES ($1, $2, nullif($3, '')::uuid, now() + $4::interval, nullif($5, '')::inet, nullif($6, ''))
		RETURNING vence_en`, hashToken(token), usuario, tienda, vida.String(), cl.IP, cl.Agente).Scan(&vence)
	return Sesion{Token: token, VenceEn: vence, TiendaID: tienda}, err
}

// permitido suma un intento a la clave y dice si sigue dentro del máximo de la ventana.
func (s *Servicio) permitido(ctx context.Context, clave string, max int, ventana time.Duration) (bool, error) {
	return limite.Permitido(ctx, s.Pool, clave, max, ventana)
}

func nuevoToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func hashToken(t string) string {
	h := sha256.Sum256([]byte(t))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func errOr(err, alt error) error {
	if err != nil {
		return err
	}
	return alt
}
