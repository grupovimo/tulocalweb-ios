package cuentas

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

type claveCtx struct{}

// TiendaDe devuelve la tienda de la sesión del pedido ("" si no hay).
func TiendaDe(r *http.Request) string {
	if d, ok := r.Context().Value(claveCtx{}).(Datos); ok {
		return d.TiendaID
	}
	return ""
}

// SesionDe devuelve los datos de la sesión del pedido.
func SesionDe(r *http.Request) (Datos, bool) {
	d, ok := r.Context().Value(claveCtx{}).(Datos)
	return d, ok
}

func token(r *http.Request) string {
	a := r.Header.Get("Authorization")
	if t, ok := strings.CutPrefix(a, "Bearer "); ok {
		return strings.TrimSpace(t)
	}
	return ""
}

// Resolver: de quién es un token, en orden: una sesión de persona, una IA conectada o una mirada
// del soporte. ErrSinSesion si no es de nadie.
func (s *Servicio) Resolver(ctx context.Context, t string) (Datos, error) {
	d, err := s.Sesion(ctx, t)
	if errors.Is(err, ErrSinSesion) && s.SesionIA != nil {
		d, err = s.SesionIA(ctx, t)
	}
	if errors.Is(err, ErrSinSesion) && s.SesionEquipo != nil {
		d, err = s.SesionEquipo(ctx, t)
	}
	return d, err
}

// Autenticar resuelve la sesión del Bearer (si viene) y la deja en el contexto. No exige sesión:
// cada ruta decide.
func (s *Servicio) Autenticar(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if t := token(r); t != "" {
			d, err := s.Resolver(r.Context(), t)
			if err == nil {
				r = r.WithContext(context.WithValue(r.Context(), claveCtx{}, d))
			} else if !errors.Is(err, ErrSinSesion) {
				slog.Error("cuentas: leer sesión", "error", err)
				problema.Escribir(w, problema.Interno())
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// EsBFF dice si el pedido trae la clave compartida del BFF (comparada en tiempo constante). Sin
// clave configurada, nunca.
func EsBFF(r *http.Request, claveBFF string) bool {
	return claveBFF != "" && subtle.ConstantTimeCompare([]byte(r.Header.Get("X-Mostrador-BFF")), []byte(claveBFF)) == 1
}

// ClienteDe arma IP y agente. La IP del visitante la manda el BFF en X-Cliente-IP, y sólo se le cree
// si el pedido trae la clave compartida del BFF. Lo que llega directo (la app de iPhone contra
// api.tulocalweb.com.ar) entra por Cloudflare y Caddy: la conexión es la de Caddy (loopback o la red
// de Docker) y la IP de la persona es la de Cf-Connecting-Ip, que Cloudflare pisa siempre (el VPS
// sólo acepta 80/443 desde Cloudflare). Sin eso, vale la IP de la conexión.
func ClienteDe(r *http.Request, claveBFF string) Cliente {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = r.RemoteAddr
	}
	if EsBFF(r, claveBFF) {
		if v := net.ParseIP(r.Header.Get("X-Cliente-IP")); v != nil {
			ip = v.String()
		}
	} else if c := net.ParseIP(ip); c != nil && (c.IsLoopback() || c.IsPrivate()) {
		if v := net.ParseIP(strings.TrimSpace(r.Header.Get("Cf-Connecting-Ip"))); v != nil {
			ip = v.String()
		}
	}
	agente := r.Header.Get("X-Cliente-Agente")
	if agente == "" {
		agente = r.UserAgent()
	}
	if len(agente) > 200 {
		agente = agente[:200]
	}
	return Cliente{IP: ip, Agente: agente}
}

// Rutas del dominio.
func (s *Servicio) Rutas(mux *http.ServeMux, claveBFF string) {
	s.rutasSociales(mux, claveBFF)
	s.rutasSegundoFactor(mux, claveBFF)
	mux.HandleFunc("POST /v1/auth/registro", func(w http.ResponseWriter, r *http.Request) {
		var a struct{ Nombre, Tienda, Email, Clave string }
		if !leer(w, r, &a) {
			return
		}
		ses, err := s.Registrar(r.Context(), Alta(a), ClienteDe(r, claveBFF))
		if err != nil {
			escribirError(w, err)
			return
		}
		escribir(w, http.StatusCreated, ses)
	})
	mux.HandleFunc("POST /v1/auth/ingresar", func(w http.ResponseWriter, r *http.Request) {
		var a struct {
			Email, Clave string
			Recordarme   bool
		}
		if !leer(w, r, &a) {
			return
		}
		ses, err := s.Ingresar(r.Context(), a.Email, a.Clave, a.Recordarme, ClienteDe(r, claveBFF))
		if err != nil {
			escribirError(w, err)
			return
		}
		escribir(w, http.StatusOK, ses)
	})
	mux.HandleFunc("POST /v1/auth/salir", func(w http.ResponseWriter, r *http.Request) {
		if t := token(r); t != "" {
			if err := s.Salir(r.Context(), t); err != nil {
				escribirError(w, err)
				return
			}
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /v1/auth/sesion", func(w http.ResponseWriter, r *http.Request) {
		d, ok := SesionDe(r)
		if !ok {
			problema.Escribir(w, problema.SinSesion())
			return
		}
		escribir(w, http.StatusOK, d)
	})
	s.rutasEquipo(mux, claveBFF)
	mux.HandleFunc("POST /v1/auth/recuperar", func(w http.ResponseWriter, r *http.Request) {
		var a struct{ Email, Telefono string }
		if !leer(w, r, &a) {
			return
		}
		cl := ClienteDe(r, claveBFF)
		var err error
		if a.Telefono != "" {
			err = s.RecuperarPorWhatsApp(r.Context(), a.Telefono, cl)
		} else {
			err = s.Recuperar(r.Context(), a.Email, cl)
		}
		if err != nil {
			escribirError(w, err)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	})
	mux.HandleFunc("POST /v1/auth/restablecer/verificar", func(w http.ResponseWriter, r *http.Request) {
		var a struct{ Token string }
		if !leer(w, r, &a) {
			return
		}
		if err := s.VerificarRestablecer(r.Context(), a.Token); err != nil {
			escribirError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("POST /v1/auth/restablecer", func(w http.ResponseWriter, r *http.Request) {
		var a struct{ Token, Clave string }
		if !leer(w, r, &a) {
			return
		}
		if err := s.Restablecer(r.Context(), a.Token, a.Clave); err != nil {
			escribirError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

// rutasEquipo: las tiendas de la persona y las invitaciones al equipo de una tienda.
func (s *Servicio) rutasEquipo(mux *http.ServeMux, claveBFF string) {
	mux.HandleFunc("GET /v1/auth/tiendas", func(w http.ResponseWriter, r *http.Request) {
		d, ok := SesionDe(r)
		if !ok || d.TiendaID == "" {
			problema.Escribir(w, problema.SinSesion())
			return
		}
		a, err := s.AccesoDe(r.Context(), d)
		if err != nil {
			escribirError(w, err)
			return
		}
		escribir(w, http.StatusOK, a)
	})
	mux.HandleFunc("PUT /v1/auth/tienda", func(w http.ResponseWriter, r *http.Request) {
		var a struct {
			TiendaID string `json:"tiendaId"`
		}
		if _, ok := SesionDe(r); !ok {
			problema.Escribir(w, problema.SinSesion())
			return
		}
		if !leer(w, r, &a) {
			return
		}
		if err := s.ElegirTienda(r.Context(), token(r), a.TiendaID); err != nil {
			escribirError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("POST /v1/auth/invitacion", func(w http.ResponseWriter, r *http.Request) {
		var a struct{ Token string }
		if !leer(w, r, &a) {
			return
		}
		inv, err := s.VerInvitacion(r.Context(), a.Token)
		if err != nil {
			escribirError(w, err)
			return
		}
		escribir(w, http.StatusOK, inv)
	})
	mux.HandleFunc("POST /v1/auth/invitacion/aceptar", func(w http.ResponseWriter, r *http.Request) {
		var a struct{ Token string }
		if _, ok := SesionDe(r); !ok {
			problema.Escribir(w, problema.SinSesion())
			return
		}
		if !leer(w, r, &a) {
			return
		}
		t, err := s.AceptarInvitacion(r.Context(), token(r), a.Token)
		if err != nil {
			escribirError(w, err)
			return
		}
		escribir(w, http.StatusOK, map[string]string{"tiendaId": t})
	})
	mux.HandleFunc("POST /v1/auth/invitacion/registro", func(w http.ResponseWriter, r *http.Request) {
		var a struct{ Token, Nombre, Clave string }
		if !leer(w, r, &a) {
			return
		}
		ses, err := s.RegistrarConInvitacion(r.Context(), a.Token, a.Nombre, a.Clave, ClienteDe(r, claveBFF))
		if err != nil {
			escribirError(w, err)
			return
		}
		escribir(w, http.StatusCreated, ses)
	})
}

// rutasSociales: proveedores OAuth y link por WhatsApp.
func (s *Servicio) rutasSociales(mux *http.ServeMux, claveBFF string) {
	mux.HandleFunc("GET /v1/auth/proveedores", func(w http.ResponseWriter, _ *http.Request) {
		// clerkPublicable: la app de iPhone habla con la Frontend API de esta instancia (sin
		// configurarla aparte); vacía si Clerk está apagado.
		publicable := ""
		if s.Clerk != nil {
			publicable = s.Clerk.Publicable
		}
		escribir(w, http.StatusOK, map[string]any{"oauth": s.Disponibles(), "clerk": s.Clerk != nil, "clerkPublicable": publicable,
			"whatsapp": s.WhatsApp != nil})
	})
	mux.HandleFunc("POST /v1/auth/oauth/{proveedor}/inicio", func(w http.ResponseWriter, r *http.Request) {
		u, err := s.IniciarOAuth(r.Context(), r.PathValue("proveedor"), ClienteDe(r, claveBFF))
		if err != nil {
			escribirError(w, err)
			return
		}
		escribir(w, http.StatusOK, map[string]string{"url": u})
	})
	mux.HandleFunc("POST /v1/auth/oauth/{proveedor}/vuelta", func(w http.ResponseWriter, r *http.Request) {
		var a struct{ Codigo, Estado string }
		if !leer(w, r, &a) {
			return
		}
		res, err := s.VolverOAuth(r.Context(), r.PathValue("proveedor"), a.Codigo, a.Estado, ClienteDe(r, claveBFF))
		if err != nil {
			escribirError(w, err)
			return
		}
		escribir(w, http.StatusOK, res)
	})
	// Token de sesión de Clerk (web o app de iPhone) → sesión de Tu Local Web.
	mux.HandleFunc("POST /v1/auth/clerk", func(w http.ResponseWriter, r *http.Request) {
		var a struct{ Token, Invitacion string }
		if !leer(w, r, &a) {
			return
		}
		res, err := s.IngresarConClerkInvitada(r.Context(), a.Token, a.Invitacion, ClienteDe(r, claveBFF))
		if err != nil {
			escribirError(w, err)
			return
		}
		escribir(w, http.StatusOK, res)
	})
	mux.HandleFunc("POST /v1/auth/enlace", func(w http.ResponseWriter, r *http.Request) {
		var a struct{ Telefono string }
		if !leer(w, r, &a) {
			return
		}
		if err := s.PedirEnlace(r.Context(), a.Telefono, ClienteDe(r, claveBFF)); err != nil {
			escribirError(w, err)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	})
	mux.HandleFunc("POST /v1/auth/enlace/usar", func(w http.ResponseWriter, r *http.Request) {
		var a struct{ Token string }
		if !leer(w, r, &a) {
			return
		}
		ses, err := s.UsarEnlace(r.Context(), a.Token, ClienteDe(r, claveBFF))
		if err != nil {
			escribirError(w, err)
			return
		}
		escribir(w, http.StatusOK, ses)
	})
}

func leer(w http.ResponseWriter, r *http.Request, v any) bool {
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10))
	if err := d.Decode(v); err != nil {
		problema.Escribir(w, problema.Invalido("El pedido no es un JSON válido."))
		return false
	}
	return true
}

func escribir(w http.ResponseWriter, estado int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(estado)
	_ = json.NewEncoder(w).Encode(v)
}

func escribirError(w http.ResponseWriter, err error) {
	var inv ErrInvalido
	switch {
	case errors.As(err, &inv):
		problema.Escribir(w, problema.Invalido(inv.Mensaje, problema.Campo{Ruta: inv.Campo, Mensaje: inv.Mensaje}))
	case errors.Is(err, ErrCredenciales):
		p := problema.SinSesion()
		p.Codigo, p.Titulo, p.Detalle = "credenciales_invalidas", "No pudimos ingresar", "El email o la contraseña no son correctos."
		// Cuántos intentos quedan antes de pausar el ingreso de ese email (la app de iPhone lo dice).
		var c ErrClave
		if errors.As(err, &c) {
			p.Restantes = &c.Restantes
		}
		problema.Escribir(w, p)
	case errors.Is(err, ErrOtroEmail):
		problema.Escribir(w, problema.Conflicto("otro_email", "Esta invitación es para otro email. Entrá con la cuenta de ese email o pedile a quien te invitó que te mande una nueva."))
	case errors.Is(err, ErrEmailEnUso):
		problema.Escribir(w, problema.Conflicto("email_en_uso", "Ese email ya tiene una cuenta. Probá ingresar o recuperar la contraseña."))
	case errors.Is(err, ErrDemasiadosIntentos):
		w.Header().Set("Retry-After", "900")
		problema.Escribir(w, &problema.Problema{Estado: http.StatusTooManyRequests, Codigo: "demasiados_intentos",
			Titulo: "Esperá un rato", Detalle: "Hubo demasiados intentos. Probá de nuevo en 15 minutos."})
	case errors.Is(err, ErrProveedorApagado):
		problema.Escribir(w, problema.NoProcesable("proveedor_apagado", "Ese ingreso todavía no está disponible. Probá con tu email."))
	case errors.Is(err, ErrSinWhatsApp):
		problema.Escribir(w, problema.NoProcesable("whatsapp_apagado", "Entrar por WhatsApp todavía no está disponible. Probá con tu email."))
	case errors.Is(err, ErrEstadoInvalido):
		problema.Escribir(w, problema.NoProcesable("estado_invalido", "El ingreso se cortó o tardó demasiado. Probá de nuevo."))
	case errors.Is(err, ErrSinEmail):
		problema.Escribir(w, problema.NoProcesable("sin_email", "No pudimos saber tu email desde esa cuenta. Probá con tu email."))
	case errors.Is(err, ErrVincular):
		problema.Escribir(w, problema.Conflicto("vincular", "Ya hay una cuenta con ese email. Ingresá con tu contraseña y conectá esa cuenta desde Mis datos."))
	case errors.Is(err, ErrTokenClerk):
		p := problema.SinSesion()
		p.Codigo, p.Titulo, p.Detalle = "clerk_token_invalido", "No pudimos ingresar", "El ingreso venció o no es válido. Probá de nuevo."
		problema.Escribir(w, p)
	case errors.Is(err, ErrSinSesion):
		problema.Escribir(w, problema.SinSesion())
	case errors.Is(err, ErrNoEsTuya):
		problema.Escribir(w, problema.NoEncontrado("No sos parte de esa tienda."))
	case errors.Is(err, ErrTokenInvalido):
		problema.Escribir(w, problema.NoProcesable("link_invalido", "El link venció o ya se usó. Pedí uno nuevo."))
	default:
		slog.Error("cuentas", "error", err)
		problema.Escribir(w, problema.Interno())
	}
}
