// Package servidor arma el http.Handler de la API: rutas de salud y, más adelante, los dominios.
package servidor

import (
	"context"
	"encoding/json"
	"errors"
	"html"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/arrepentimiento"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/configuracion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/crecer"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cupones"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/devoluciones"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/equipo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/facturacion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/finanzas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/formularios"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/hoy"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/ia"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/importacion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/lanzamientos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/links"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/mitienda"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/newsletter"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/notificaciones"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/partners"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/precios"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/push"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/roles"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/suscripciones"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/tiendas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/turnos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/ventas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/vidriera"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/visitas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/conversiones"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/correoargentino"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/mercadolibre"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/mercadopago"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/telegram"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/auditoria"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/geoip"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/limite"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/tiemporeal"
)

// Dependencias del servidor. App puede ser nil en tests que no tocan la base.
type Dependencias struct {
	App *db.Pool
	Log *slog.Logger
	// Cuentas y sesiones. Sin esto, el panel responde 401.
	Cuentas  *cuentas.Servicio
	ClaveBFF string
	// Botón de arrepentimiento (público, sin sesión).
	Arrepentimiento *arrepentimiento.Servicio
	// Formularios públicos del sitio: contacto, partners y novedades.
	Formularios *formularios.Servicio
	// Sitio público (para armar la dirección de las tiendas).
	Sitio string
	// Mercado Pago: conexión de las tiendas, cobros, abono y webhook. Sin credenciales, responde 409.
	MercadoPago *mercadopago.Servicio
	// Mercado Libre como canal: conexión, publicaciones, pedidos y preguntas.
	MercadoLibre *mercadolibre.Servicio
	// Correo Argentino (API MiCorreo): conexión, cotización del checkout, sucursales y envíos.
	CorreoArgentino *correoargentino.Servicio
	// IA conectadas por MCP (OAuth 2.1). Sin esto, los tokens de IA no entran.
	IA *ia.Servicio
	// Facturación con ARCA (comprobantes, certificado, factura manual).
	Facturacion *facturacion.Servicio
	// Correo para los avisos de la tienda (arrepentimiento). nil: no se manda.
	Correo correo.Correo
	// Bandeja del equipo: los pedidos de contacto de quien está por cancelar. Vacía: no se avisa.
	CorreoEquipo string
	// Hornero en Telegram: vínculo del chat (panel) y lo que pide el bot por el BFF.
	Telegram *telegram.Servicio
	// Backoffice del equipo (/v1/equipo/*) y "ver la tienda" de sólo lectura. nil: no se monta.
	Equipo *equipo.Servicio
	// Importar la tienda desde otra plataforma (el avance lo hace el worker). nil: no se monta.
	Importacion *importacion.Servicio
	// API de conversiones de Meta (§47): el token de cada tienda. nil: no se monta.
	Conversiones *conversiones.Servicio
	// Clave pública VAPID del push web del panel instalado (vacía: apagado; §43).
	VAPIDPublica string
	// Tiempo real (§44): el hub que reparte los avisos de Postgres. nil: no se monta.
	TiempoReal *tiemporeal.Hub
	// De dónde se acepta el WebSocket del panel (además del Sitio).
	TiempoRealOrigenes []string
	// Desarrollo: el tiempo real acepta las cookies sin __Host- y los orígenes de localhost.
	Desarrollo bool
}

// Nuevo devuelve el handler raíz de la API.
func Nuevo(d Dependencias) http.Handler {
	if d.Log == nil {
		d.Log = slog.Default()
	}
	mux := http.NewServeMux()
	// Vivo: el proceso responde. No toca la base, para que un corte de Postgres no reinicie la API.
	mux.HandleFunc("GET /salud/vivo", func(w http.ResponseWriter, _ *http.Request) {
		escribirJSON(w, http.StatusOK, map[string]string{"estado": "vivo"})
	})
	// Listo: puede atender pedidos (la base responde). El balanceador saca la instancia si no.
	mux.HandleFunc("GET /salud/listo", func(w http.ResponseWriter, r *http.Request) {
		if d.App == nil {
			problema.Escribir(w, problema.Interno())
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := d.App.Ping(ctx); err != nil {
			d.Log.Warn("salud: la base no responde", "error", err)
			w.Header().Set("Retry-After", "5")
			problema.Escribir(w, &problema.Problema{Estado: http.StatusServiceUnavailable, Codigo: "sin_base",
				Titulo: "No disponible", Detalle: "La base de datos no responde."})
			return
		}
		escribirJSON(w, http.StatusOK, map[string]string{"estado": "listo"})
	})
	var h http.Handler = mux
	if d.App != nil {
		// La tienda del pedido sale siempre de la sesión: nunca de algo que diga el cliente.
		tienda := func(*http.Request) string { return "" }
		if d.Cuentas != nil {
			d.Cuentas.Rutas(mux, d.ClaveBFF)
			tienda = cuentas.TiendaDe
			var conGuardia http.Handler = mux
			if d.IA != nil {
				d.Cuentas.SesionIA = d.IA.ParaCuentas
				d.IA.Rutas(mux, func(r *http.Request) string { return cuentas.ClienteDe(r, d.ClaveBFF).IP })
				conGuardia = d.IA.Guardia(mux)
			}
			if d.Equipo != nil {
				d.Cuentas.SesionEquipo = d.Equipo.ParaCuentas
				conGuardia = equipo.SoloLectura(conGuardia)
			}
			// El rol de cada persona del equipo, en cada ruta del panel (docs/decisiones.md §37).
			conGuardia = roles.Guardia(func(r *http.Request) (string, bool) {
				s, ok := cuentas.SesionDe(r)
				return s.Rol, ok && s.TiendaID != ""
			}, conGuardia)
			h = d.Cuentas.Autenticar(conActor(d.ClaveBFF, conGuardia))
		}
		tiendas.Rutas(mux, d.App, tienda)
		notificaciones.Rutas(mux, d.App)
		pedidos.Rutas(mux, d.App, tienda)
		var fact listas.Facturas
		if d.Facturacion != nil {
			fact = facturacion.ParaListas{S: d.Facturacion}
			d.Facturacion.Rutas(mux)
		}
		listas.Rutas(mux, d.App, tienda, fact)
		var preguntas ventas.Preguntas
		if d.MercadoLibre != nil {
			preguntas = d.MercadoLibre
		}
		ventas.Rutas(mux, d.App, tienda, preguntas)
		var pagar links.Pagar
		if d.MercadoPago != nil {
			pagar = d.MercadoPago.PagarLink
		}
		links.Rutas(mux, d.App, tienda, pagar, d.MercadoPago.Problema)
		precios.Rutas(mux, d.App, tienda)
		cupones.Rutas(mux, d.App, tienda)
		crecer.Rutas(mux, d.App, tienda)
		newsletter.Rutas(mux, d.App, tienda)
		hoy.Rutas(mux, d.App, tienda)
		finanzas.Rutas(mux, d.App, tienda)
		mitienda.Rutas(mux, d.App, tienda, d.Sitio)
		// Los iPhone de la app (token de APNs atado a la sesión).
		push.Rutas(mux, d.App)
		// El panel instalado como app (suscripción de Web Push atada a la sesión).
		push.RutasWeb(mux, d.App, d.VAPIDPublica)
		partners.Rutas(mux, d.App, d.ClaveBFF, nil)
		var geo geoip.Ubicador
		if d.Cuentas != nil {
			geo = d.Cuentas.Geo
		}
		ganchos := configuracion.Ganchos{}
		if d.MercadoPago != nil {
			ganchos.Beneficio = d.MercadoPago.AplicarBeneficio
		}
		if d.Correo != nil && d.CorreoEquipo != "" {
			ganchos.Contacto = func(ctx context.Context, asunto, texto string) error {
				return d.Correo.Enviar(ctx, correo.Mensaje{Para: d.CorreoEquipo, Asunto: asunto, Texto: texto,
					HTML: "<pre style=\"font:14px/1.5 sans-serif;white-space:pre-wrap\">" + html.EscapeString(texto) + "</pre>"})
			}
		}
		configuracion.Rutas(mux, d.App, geo, d.Sitio, d.MercadoPago != nil && d.MercadoPago.Config.TokenPlataforma != "", ganchos)
		if d.MercadoPago != nil {
			d.MercadoPago.Rutas(mux)
		}
		if d.MercadoLibre != nil {
			d.MercadoLibre.Rutas(mux)
		}
		var pagador vidriera.Pagador
		if d.MercadoPago != nil {
			pagador = d.MercadoPago
		}
		var cobrador turnos.Cobrador
		if d.MercadoPago != nil {
			cobrador = d.MercadoPago
		}
		turnos.Rutas(mux, d.App, turnos.Config{Tienda: tienda, MP: cobrador, Sitio: d.Sitio, Log: d.Log,
			IP: func(r *http.Request) string { return cuentas.ClienteDe(r, d.ClaveBFF).IP },
			Permitido: func(ctx context.Context, clave string, max int, ventana time.Duration) (bool, error) {
				return limite.Permitido(ctx, d.App, clave, max, ventana)
			},
			Admin: func(r *http.Request) bool {
				s, ok := cuentas.SesionDe(r)
				return ok && (s.Rol == "duenio" || s.Rol == "admin")
			}})
		// Suscripciones de la tienda (§51): sin Mercado Pago en el servidor no se ofrecen planes.
		var suscriptor suscripciones.Cobrador
		if d.MercadoPago != nil {
			suscriptor = d.MercadoPago
		}
		suscripciones.Rutas(mux, d.App, suscripciones.Config{Tienda: tienda, MP: suscriptor, Sitio: d.Sitio, Log: d.Log,
			IP: func(r *http.Request) string { return cuentas.ClienteDe(r, d.ClaveBFF).IP },
			Permitido: func(ctx context.Context, clave string, max int, ventana time.Duration) (bool, error) {
				return limite.Permitido(ctx, d.App, clave, max, ventana)
			}})
		if d.Telegram != nil {
			d.Telegram.Rutas(mux, d.ClaveBFF)
		}
		if d.Importacion != nil {
			d.Importacion.Rutas(mux)
		}
		if d.Conversiones != nil {
			d.Conversiones.Rutas(mux)
		}
		ip := func(r *http.Request) string { return cuentas.ClienteDe(r, d.ClaveBFF).IP }
		dev := &devoluciones.Servicio{Pool: d.App, Correo: d.Correo}
		dev.Rutas(mux, tienda)
		var correo vidriera.Correo
		if d.CorreoArgentino != nil {
			correo = d.CorreoArgentino
			d.CorreoArgentino.Rutas(mux, d.App, ip)
		}
		vidriera.Rutas(mux, d.App, pagador, d.Sitio, dev, ip, correo)
		// Lanzamientos, ofertas con fin y señas (§52): el editor, «saldado» y «Avisame cuando salga».
		lanzamientos.Rutas(mux, d.App, tienda, ip)
		// Las visitas de las tiendas las manda la web por lotes, con su clave.
		visitas.Rutas(mux, d.App, d.ClaveBFF, d.Log)
		if d.Arrepentimiento != nil {
			d.Arrepentimiento.Rutas(mux, ip)
		}
		if d.Formularios != nil {
			d.Formularios.Rutas(mux, ip)
		}
		if d.Equipo != nil {
			d.Equipo.Rutas(mux)
		}
		if d.TiempoReal != nil && d.Cuentas != nil {
			rutasTiempoReal(mux, d, ip)
		}
	}
	return recuperar(d.Log, h)
}

// rutasTiempoReal monta el WebSocket del panel y el de la tienda online (docs/decisiones.md §44).
func rutasTiempoReal(mux *http.ServeMux, d Dependencias, ip func(*http.Request) string) {
	origenes := append([]string{strings.TrimRight(d.Sitio, "/")}, d.TiempoRealOrigenes...)
	if d.Desarrollo {
		origenes = append(origenes, "http://localhost:*", "http://127.0.0.1:*")
	}
	tiemporeal.Rutas(mux, tiemporeal.Config{
		Hub:        d.TiempoReal,
		Origenes:   origenes,
		Desarrollo: d.Desarrollo,
		IP:         ip,
		Log:        d.Log,
		Autenticar: func(ctx context.Context, token string) (tiemporeal.Quien, error) {
			s, err := d.Cuentas.Resolver(ctx, token)
			if errors.Is(err, cuentas.ErrSinSesion) {
				return tiemporeal.Quien{}, tiemporeal.ErrNoExiste
			}
			if err != nil {
				return tiemporeal.Quien{}, err
			}
			return tiemporeal.Quien{Usuario: s.Usuario.ID, Sesion: s.SesionID, Tienda: s.TiendaID, Rol: s.Rol,
				VenceEn: s.VenceEn, IA: s.IA != ""}, nil
		},
		Publico: func(ctx context.Context, slug, codigo string) (string, string, error) {
			t, p, err := vidriera.TiendaYPedido(ctx, d.App, slug, codigo)
			if errors.Is(err, vidriera.ErrNoExiste) {
				return "", "", tiemporeal.ErrNoExiste
			}
			return t, p, err
		},
		SlugDeDominio: func(ctx context.Context, dominio string) (string, error) {
			return vidriera.SlugDeDominio(ctx, d.App, dominio)
		},
	})
}

// conActor deja en el contexto quién hace el pedido, para que cada dominio lo anote en la Actividad
// donde hace la acción (auditoria.Registrar). Una persona del equipo de Tu Local Web mirando la
// tienda (sólo lectura) no escribe nada, así que no lleva autor.
func conActor(claveBFF string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sólo el panel: lo que hace la dueña como compradora en la tienda pública no es «del equipo».
		s, ok := cuentas.SesionDe(r)
		if ok && s.TiendaID != "" && !s.Mirada && strings.HasPrefix(r.URL.Path, "/v1/panel/") {
			c := cuentas.ClienteDe(r, claveBFF)
			a := auditoria.Actor{Tipo: "persona", ID: s.Usuario.ID, Nombre: s.Usuario.Nombre, Rol: s.Rol, IP: c.IP, Agente: c.Agente}
			if s.IA != "" {
				a.Tipo, a.IA = "ia", s.IANombre
				if a.IA == "" {
					a.IA = "Una IA"
				}
			}
			r = r.WithContext(auditoria.ConActor(r.Context(), a))
		}
		next.ServeHTTP(w, r)
	})
}

// recuperar convierte un panic en un 500 con problem+json y lo registra, sin tirar el proceso.
func recuperar(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				if v == http.ErrAbortHandler {
					panic(v)
				}
				log.Error("panic en un handler", "ruta", r.URL.Path, "valor", v)
				problema.Escribir(w, problema.Interno())
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func escribirJSON(w http.ResponseWriter, estado int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(estado)
	_ = json.NewEncoder(w).Encode(v)
}
