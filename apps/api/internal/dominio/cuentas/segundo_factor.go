package cuentas

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1" //nolint:gosec // TOTP (RFC 6238) usa HMAC-SHA1: es lo que entienden las apps autenticadoras
	"crypto/sha256"
	"encoding/base32"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

// Verificación en dos pasos (docs/decisiones.md §58). Prendida, todo ingreso a una cuenta que ya
// existe (contraseña, Google o Apple por Clerk, Mercado Pago, link por WhatsApp) termina en un
// desafío en vez de una sesión: Sesion.SegundoFactor trae su token y la sesión se abre recién con un
// código de la app autenticadora o uno de recuperación (POST /v1/auth/segundo-factor).

var (
	// ErrCodigo: el código de seis números (o el de recuperación) no es el que corresponde.
	ErrCodigo = errors.New("código incorrecto")
	// ErrDesafio: el desafío no existe, venció o se le acabaron los intentos.
	ErrDesafio = errors.New("desafío vencido")
	// ErrSinCifrado: la API no tiene clave de cifrado (MOSTRADOR_CLAVE_CIFRADO): no se puede prender.
	ErrSinCifrado = errors.New("sin clave de cifrado")
	// ErrYaActivo / ErrNoActivo: prender lo que ya estaba, o apagar lo que no.
	ErrYaActivo = errors.New("la verificación en dos pasos ya está prendida")
	ErrNoActivo = errors.New("la verificación en dos pasos está apagada")
)

const (
	// VidaDesafio: cuánto se espera el código después de la contraseña.
	VidaDesafio = 10 * time.Minute
	// MaxIntentosCodigo por desafío.
	MaxIntentosCodigo = 5
	// CodigosRecuperacion que se entregan al prender (y al pedir nuevos).
	CodigosRecuperacion = 10
	emisorTOTP          = "Tu Local Web"
)

var base32Sin = base32.StdEncoding.WithPadding(base32.NoPadding)

// codigoTOTP es el código de seis números del secreto en ese paso de 30 segundos (RFC 6238).
func codigoTOTP(secreto []byte, paso int64) string {
	var msg [8]byte
	binary.BigEndian.PutUint64(msg[:], uint64(paso))
	m := hmac.New(sha1.New, secreto)
	m.Write(msg[:])
	h := m.Sum(nil)
	o := h[len(h)-1] & 0x0f
	n := binary.BigEndian.Uint32(h[o:o+4]) & 0x7fffffff
	return fmt.Sprintf("%06d", n%1_000_000)
}

// totpValido acepta el código del paso actual y el de al lado (el reloj del teléfono puede estar
// corrido unos segundos).
func totpValido(secreto []byte, codigo string, ahora time.Time) bool {
	codigo = soloNumeros(codigo)
	if len(codigo) != 6 {
		return false
	}
	paso := ahora.Unix() / 30
	for d := int64(-1); d <= 1; d++ {
		if hmac.Equal([]byte(codigoTOTP(secreto, paso+d)), []byte(codigo)) {
			return true
		}
	}
	return false
}

func soloNumeros(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// normalCodigo: los de recuperación se comparan sin guiones ni espacios y en minúsculas.
func normalCodigo(s string) string {
	return strings.ToLower(strings.NewReplacer("-", "", " ", "").Replace(strings.TrimSpace(s)))
}

func hashCodigo(c string) string {
	h := sha256.Sum256([]byte("recuperacion:" + normalCodigo(c)))
	return hex.EncodeToString(h[:])
}

// nuevosCodigos: diez códigos «abcd-efgh» y sus hashes (lo único que se guarda).
func nuevosCodigos() ([]string, []string) {
	const letras = "abcdefghjkmnpqrstuvwxyz23456789" // sin las que se confunden (i, l, o, 0, 1)
	var cs, hs []string
	for range CodigosRecuperacion {
		var b [8]byte
		_, _ = rand.Read(b[:])
		var s strings.Builder
		for i, x := range b {
			if i == 4 {
				s.WriteByte('-')
			}
			s.WriteByte(letras[int(x)%len(letras)])
		}
		cs = append(cs, s.String())
		hs = append(hs, hashCodigo(s.String()))
	}
	return cs, hs
}

func (s *Servicio) contextoTOTP(usuario string) string { return "totp:" + usuario }

// segundoFactorActivo dice si la persona tiene prendida la verificación en dos pasos.
func segundoFactorActivo(ctx context.Context, tx pgx.Tx, usuario string) (bool, error) {
	var activo bool
	err := tx.QueryRow(ctx, `SELECT totp_activo_en IS NOT NULL FROM auth.usuarios WHERE id = $1`, usuario).Scan(&activo)
	return activo, err
}

// crearDesafio guarda lo necesario para abrir la sesión cuando llegue el código.
func crearDesafio(ctx context.Context, tx pgx.Tx, usuario string, vida time.Duration, cl Cliente) (Sesion, error) {
	t := nuevoToken()
	var vence time.Time
	if _, err := tx.Exec(ctx, `DELETE FROM auth.desafios WHERE vence_en < now()`); err != nil {
		return Sesion{}, err
	}
	err := tx.QueryRow(ctx, `INSERT INTO auth.desafios (token_hash, usuario_id, vida, vence_en, ip, agente)
		VALUES ($1, $2, $3::interval, now() + $4::interval, nullif($5, '')::inet, nullif($6, ''))
		RETURNING vence_en`, hashToken(t), usuario, vida.String(), VidaDesafio.String(), cl.IP, cl.Agente).Scan(&vence)
	return Sesion{SegundoFactor: t, VenceEn: vence}, err
}

// verificarCodigo: un código TOTP o uno de recuperación (que se gasta). Corre en la transacción.
func (s *Servicio) verificarCodigo(ctx context.Context, tx pgx.Tx, usuario, codigo string) (bool, error) {
	if s.Cifra == nil {
		return false, ErrSinCifrado
	}
	var cifrado []byte
	var codigos []string
	if err := tx.QueryRow(ctx, `SELECT totp_secreto, codigos_recuperacion FROM auth.usuarios WHERE id = $1 FOR UPDATE`,
		usuario).Scan(&cifrado, &codigos); err != nil {
		return false, err
	}
	if cifrado == nil {
		return false, ErrNoActivo
	}
	secreto, err := s.Cifra.Descifrar(s.contextoTOTP(usuario), cifrado)
	if err != nil {
		return false, err
	}
	if totpValido(secreto, codigo, time.Now()) {
		return true, nil
	}
	if h := hashCodigo(codigo); len(normalCodigo(codigo)) == 8 && slices.Contains(codigos, h) {
		_, err := tx.Exec(ctx, `UPDATE auth.usuarios SET codigos_recuperacion = array_remove(codigos_recuperacion, $2) WHERE id = $1`, usuario, h)
		return err == nil, err
	}
	return false, nil
}

// CompletarSegundoFactor abre la sesión del desafío si el código es correcto.
func (s *Servicio) CompletarSegundoFactor(ctx context.Context, desafio, codigo string, cl Cliente) (Sesion, error) {
	if ok, err := s.permitido(ctx, "segundo:ip:"+cl.IP, 30, 15*time.Minute); err != nil || !ok {
		return Sesion{}, errOr(err, ErrDemasiadosIntentos)
	}
	var ses Sesion
	var mal bool
	err := s.Pool.EnTx(ctx, func(tx pgx.Tx) error {
		var usuario, vida string
		var intentos int
		err := tx.QueryRow(ctx, `SELECT usuario_id, vida::text, intentos FROM auth.desafios
			WHERE token_hash = $1 AND vence_en > now() FOR UPDATE`, hashToken(desafio)).Scan(&usuario, &vida, &intentos)
		if errors.Is(err, pgx.ErrNoRows) || intentos >= MaxIntentosCodigo {
			return ErrDesafio
		}
		if err != nil {
			return err
		}
		ok, err := s.verificarCodigo(ctx, tx, usuario, codigo)
		if err != nil {
			return err
		}
		if !ok {
			// El intento se cuenta aunque la transacción siga: se guarda y se corta acá.
			mal = true
			_, err := tx.Exec(ctx, `UPDATE auth.desafios SET intentos = intentos + 1 WHERE token_hash = $1`, hashToken(desafio))
			return err
		}
		if _, err := tx.Exec(ctx, `DELETE FROM auth.desafios WHERE token_hash = $1`, hashToken(desafio)); err != nil {
			return err
		}
		d, err := time.ParseDuration(intervaloAGo(vida))
		if err != nil {
			d = SesionCorta
		}
		var tienda *string
		if err := tx.QueryRow(ctx, `SELECT tienda_id FROM auth.membresias WHERE usuario_id = $1 ORDER BY creada_en LIMIT 1`,
			usuario).Scan(&tienda); err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		t := ""
		if tienda != nil {
			t = *tienda
		}
		ses, err = crearSesion(ctx, tx, usuario, t, d, cl)
		if err != nil {
			return err
		}
		return s.avisarIngreso(ctx, tx, usuario, hashToken(ses.Token), cl)
	})
	if err == nil && mal {
		return Sesion{}, ErrCodigo
	}
	return ses, err
}

// intervaloAGo pasa un interval de Postgres («8760:00:00» o «12:00:00») a una duración de Go.
func intervaloAGo(v string) string {
	p := strings.Split(v, ":")
	if len(p) != 3 {
		return v
	}
	return p[0] + "h" + p[1] + "m" + strings.SplitN(p[2], ".", 2)[0] + "s"
}

// EstadoSegundoFactor: si está prendida y cuántos códigos de recuperación quedan.
type EstadoSegundoFactor struct {
	Activo    bool       `json:"activo"`
	Desde     *time.Time `json:"desde"`
	Restantes int        `json:"codigosRestantes"`
}

func (s *Servicio) estadoSegundoFactor(ctx context.Context, usuario string) (EstadoSegundoFactor, error) {
	var e EstadoSegundoFactor
	err := s.Pool.QueryRow(ctx, `SELECT totp_activo_en IS NOT NULL, totp_activo_en, cardinality(codigos_recuperacion)
		FROM auth.usuarios WHERE id = $1`, usuario).Scan(&e.Activo, &e.Desde, &e.Restantes)
	return e, err
}

// Alta: el secreto nuevo (pendiente hasta que se confirme con un código) y el link para el QR.
type AltaSegundoFactor struct {
	Secreto string `json:"secreto"`
	URI     string `json:"uri"`
}

func (s *Servicio) prepararSegundoFactor(ctx context.Context, d Datos) (AltaSegundoFactor, error) {
	if s.Cifra == nil {
		return AltaSegundoFactor{}, ErrSinCifrado
	}
	e, err := s.estadoSegundoFactor(ctx, d.Usuario.ID)
	if err != nil {
		return AltaSegundoFactor{}, err
	}
	if e.Activo {
		return AltaSegundoFactor{}, ErrYaActivo
	}
	var secreto [20]byte
	_, _ = rand.Read(secreto[:])
	cifrado, err := s.Cifra.Cifrar(s.contextoTOTP(d.Usuario.ID), secreto[:])
	if err != nil {
		return AltaSegundoFactor{}, err
	}
	if _, err := s.Pool.Exec(ctx, `UPDATE auth.usuarios SET totp_pendiente = $2 WHERE id = $1`, d.Usuario.ID, cifrado); err != nil {
		return AltaSegundoFactor{}, err
	}
	b32 := base32Sin.EncodeToString(secreto[:])
	q := url.Values{"secret": {b32}, "issuer": {emisorTOTP}, "digits": {"6"}, "period": {"30"}}
	uri := "otpauth://totp/" + url.PathEscape(emisorTOTP+":"+d.Usuario.Email) + "?" + q.Encode()
	return AltaSegundoFactor{Secreto: b32, URI: uri}, nil
}

// confirmarSegundoFactor: con el primer código correcto queda prendida y se entregan los códigos de
// recuperación (es la única vez que se ven).
func (s *Servicio) confirmarSegundoFactor(ctx context.Context, d Datos, codigo string) ([]string, error) {
	if s.Cifra == nil {
		return nil, ErrSinCifrado
	}
	var codigos []string
	err := s.Pool.EnTx(ctx, func(tx pgx.Tx) error {
		var pendiente []byte
		var activo bool
		if err := tx.QueryRow(ctx, `SELECT totp_pendiente, totp_activo_en IS NOT NULL FROM auth.usuarios WHERE id = $1 FOR UPDATE`,
			d.Usuario.ID).Scan(&pendiente, &activo); err != nil {
			return err
		}
		if activo {
			return ErrYaActivo
		}
		if pendiente == nil {
			return ErrNoActivo
		}
		secreto, err := s.Cifra.Descifrar(s.contextoTOTP(d.Usuario.ID), pendiente)
		if err != nil {
			return err
		}
		if !totpValido(secreto, codigo, time.Now()) {
			return ErrCodigo
		}
		var hashes []string
		codigos, hashes = nuevosCodigos()
		_, err = tx.Exec(ctx, `UPDATE auth.usuarios SET totp_secreto = totp_pendiente, totp_pendiente = NULL, totp_activo_en = now(),
			codigos_recuperacion = $2 WHERE id = $1`, d.Usuario.ID, hashes)
		return err
	})
	return codigos, err
}

// apagarSegundoFactor pide un código (de la app o de recuperación): con la sesión sola no alcanza.
func (s *Servicio) apagarSegundoFactor(ctx context.Context, d Datos, codigo string) error {
	return s.Pool.EnTx(ctx, func(tx pgx.Tx) error {
		ok, err := s.verificarCodigo(ctx, tx, d.Usuario.ID, codigo)
		if err != nil {
			return err
		}
		if !ok {
			return ErrCodigo
		}
		_, err = tx.Exec(ctx, `UPDATE auth.usuarios SET totp_secreto = NULL, totp_pendiente = NULL, totp_activo_en = NULL,
			codigos_recuperacion = '{}' WHERE id = $1`, d.Usuario.ID)
		return err
	})
}

// nuevosCodigosRecuperacion reemplaza los códigos (los viejos dejan de servir).
func (s *Servicio) nuevosCodigosRecuperacion(ctx context.Context, d Datos, codigo string) ([]string, error) {
	var codigos []string
	err := s.Pool.EnTx(ctx, func(tx pgx.Tx) error {
		ok, err := s.verificarCodigo(ctx, tx, d.Usuario.ID, codigo)
		if err != nil {
			return err
		}
		if !ok {
			return ErrCodigo
		}
		var hashes []string
		codigos, hashes = nuevosCodigos()
		_, err = tx.Exec(ctx, `UPDATE auth.usuarios SET codigos_recuperacion = $2 WHERE id = $1`, d.Usuario.ID, hashes)
		return err
	})
	return codigos, err
}

func errorSegundoFactor(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrCodigo):
		problema.Escribir(w, problema.Invalido("El código no es correcto. Fijate el que muestra tu app ahora.",
			problema.Campo{Ruta: "codigo", Mensaje: "El código no es correcto."}))
	case errors.Is(err, ErrDesafio):
		p := problema.SinSesion()
		p.Codigo, p.Titulo, p.Detalle = "desafio_vencido", "Se venció el ingreso", "Pasaron más de 10 minutos o muchos intentos. Ingresá de nuevo."
		problema.Escribir(w, p)
	case errors.Is(err, ErrSinCifrado):
		problema.Escribir(w, problema.NoProcesable("sin_cifrado", "La verificación en dos pasos todavía no está disponible."))
	case errors.Is(err, ErrYaActivo):
		problema.Escribir(w, problema.Conflicto("ya_activo", "La verificación en dos pasos ya está prendida."))
	case errors.Is(err, ErrNoActivo):
		problema.Escribir(w, problema.Conflicto("no_activo", "Primero prendé la verificación en dos pasos."))
	default:
		escribirError(w, err)
	}
}

func (s *Servicio) rutasSegundoFactor(mux *http.ServeMux, claveBFF string) {
	// Completar el ingreso: el desafío que devolvió el ingreso y el código.
	mux.HandleFunc("POST /v1/auth/segundo-factor", func(w http.ResponseWriter, r *http.Request) {
		var a struct{ Desafio, Codigo string }
		if !leer(w, r, &a) {
			return
		}
		ses, err := s.CompletarSegundoFactor(r.Context(), a.Desafio, a.Codigo, ClienteDe(r, claveBFF))
		if err != nil {
			errorSegundoFactor(w, err)
			return
		}
		escribir(w, http.StatusOK, ses)
	})
	// Lo demás es de la persona con su sesión (no una IA ni una mirada del soporte).
	persona := func(w http.ResponseWriter, r *http.Request) (Datos, bool) {
		d, ok := SesionDe(r)
		if !ok || d.SesionHash == "" {
			problema.Escribir(w, problema.SinSesion())
			return d, false
		}
		return d, true
	}
	mux.HandleFunc("GET /v1/auth/segundo-factor", func(w http.ResponseWriter, r *http.Request) {
		d, ok := persona(w, r)
		if !ok {
			return
		}
		e, err := s.estadoSegundoFactor(r.Context(), d.Usuario.ID)
		if err != nil {
			errorSegundoFactor(w, err)
			return
		}
		escribir(w, http.StatusOK, e)
	})
	mux.HandleFunc("POST /v1/auth/segundo-factor/activar", func(w http.ResponseWriter, r *http.Request) {
		d, ok := persona(w, r)
		if !ok {
			return
		}
		a, err := s.prepararSegundoFactor(r.Context(), d)
		if err != nil {
			errorSegundoFactor(w, err)
			return
		}
		escribir(w, http.StatusOK, a)
	})
	conCodigo := func(f func(context.Context, Datos, string) ([]string, error)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			d, ok := persona(w, r)
			if !ok {
				return
			}
			var a struct{ Codigo string }
			if !leer(w, r, &a) {
				return
			}
			if ok, err := s.permitido(r.Context(), "segundo:usuario:"+d.Usuario.ID, 10, 15*time.Minute); err != nil || !ok {
				errorSegundoFactor(w, errOr(err, ErrDemasiadosIntentos))
				return
			}
			cs, err := f(r.Context(), d, a.Codigo)
			if err != nil {
				errorSegundoFactor(w, err)
				return
			}
			if cs == nil {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			escribir(w, http.StatusOK, map[string][]string{"codigos": cs})
		}
	}
	mux.HandleFunc("POST /v1/auth/segundo-factor/confirmar", conCodigo(s.confirmarSegundoFactor))
	mux.HandleFunc("POST /v1/auth/segundo-factor/codigos", conCodigo(s.nuevosCodigosRecuperacion))
	mux.HandleFunc("POST /v1/auth/segundo-factor/desactivar", conCodigo(func(ctx context.Context, d Datos, c string) ([]string, error) {
		return nil, s.apagarSegundoFactor(ctx, d, c)
	}))
}
