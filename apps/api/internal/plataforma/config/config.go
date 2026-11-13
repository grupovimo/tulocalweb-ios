// Package config lee la configuración de la API del entorno. Fuera de desarrollo, lo que falte
// hace que el proceso no arranque, en vez de arrancar a medias.
package config

import (
	"fmt"
	"os"
	"strings"
)

// Config de la API y el worker.
type Config struct {
	Entorno   string // desarrollo | staging | produccion
	Direccion string
	// URLs de Postgres por rol.
	BaseApp, BaseWorker, BaseMigrador, BaseAdmin string
	// BaseStaff: el rol mostrador_staff (backoffice del equipo). Vacía fuera de desarrollo: no se monta.
	BaseStaff string
	// Sitio público (links de los mails).
	Sitio string
	// Clave que comparte con el BFF (SvelteKit): habilita creerle X-Cliente-IP.
	ClaveBFF string
	// Correo: Resend, o un SMTP con usuario (el servidor propio); en desarrollo, SMTP de Mailpit.
	ResendClave, CorreoDe, SMTP, SMTPUsuario, SMTPClave string
	// Bandeja del equipo: le llegan las consultas de ventas, las postulaciones de partners y los
	// arrepentimientos. Vacía: no se avisa (los pedidos quedan igual en la base).
	CorreoEquipo string
	// Ingreso con proveedores: {mercadopago|mercadolibre|tiendanube: [client id, secreto]}. Sin
	// client id, el proveedor queda apagado.
	OAuth map[string][2]string
	// Clerk (Google, Apple y demás ingresos sociales de Grupo Vimo): clave secreta y publicable de la
	// instancia (de la publicable sale el emisor de los tokens) y los orígenes que pueden haber pedido
	// el token (azp; vacío: el sitio). Sin clave secreta, apagado.
	ClerkSecreto, ClerkPublicable string
	ClerkOrigenes                 []string
	// WhatsApp Cloud de la plataforma (links de ingreso y de contraseña nueva).
	WhatsAppToken, WhatsAppNumeroID string
	// Mercado Pago: token de la cuenta de la plataforma (el abono) y clave de los webhooks. El OAuth
	// de las tiendas usa el mismo client id y secreto que el ingreso.
	MPToken, MPSecretoWebhook string
	// Clave maestra (32 bytes en base64) para cifrar tokens de terceros en la base.
	ClaveCifrado string
	// Token del bot de Telegram (Hornero para la dueña): con él, el worker avisa cada venta al chat
	// vinculado. Vacío: no se avisa. Es el mismo TELEGRAM_BOT_TOKEN que tiene el Worker.
	TelegramToken string
	// Correo Argentino (API MiCorreo): las credenciales de API de Tu Local Web para /token (las da
	// Correo, una por ambiente). Vacías: cada tienda conecta con las suyas. CAPruebas usa apitest.
	CAUsuario, CAClave string
	CAPruebas          bool
	// Base DB-IP City Lite (mmdb) para ubicar la IP de cada sesión. Vacía: las sesiones van sin lugar.
	GeoIP string
	// APNs (push al iPhone, docs/decisiones.md §40): Key ID y Team ID de la cuenta de Apple, la ruta
	// al .p8 (nunca en el repo) y el topic (el bundle id de la app). Sin Key ID, Team ID o clave, el
	// worker no manda pushes.
	APNsKeyID, APNsTeamID, APNsClave, APNsTopic string
	// FCM (push a Android): la ruta del JSON de la cuenta de servicio de Firebase (nunca en el repo).
	// Sin eso, el worker no manda avisos a Android.
	FCMCuenta string
	// Web Push del panel instalado (docs/decisiones.md §43): el par VAPID en base64url (`go run
	// ./cmd/vapid`) y el sujeto (mailto: o https:). Sin las tres, no hay push web.
	VAPIDPublica, VAPIDPrivada, VAPIDSujeto string
	// Tiempo real (docs/decisiones.md §44): orígenes extra que pueden abrir el WebSocket del panel,
	// además de MOSTRADOR_SITIO (separados por coma; admiten comodines como https://*.ejemplo.com).
	TiempoRealOrigenes []string
	// Borrado de las tiendas cerradas hace más de 90 días (MOSTRADOR_PURGA_TIENDAS=1). Apagado por
	// defecto: `mostradorctl purga-tiendas --listar` muestra antes qué borraría.
	PurgaTiendas bool
	// La carpeta de archivos de la web (su DATOS_DIR/archivos, montada en el worker): de ahí se borran
	// las fotos de las tiendas purgadas (fotos/<tienda>/). Vacía: las fotos no se tocan.
	DirArchivos string
}

// Desarrollo dice si se está corriendo en local.
func (c Config) Desarrollo() bool { return c.Entorno == "desarrollo" }

// Leer arma la configuración desde las variables MOSTRADOR_*.
func Leer() (Config, error) {
	c := Config{
		Entorno:      valor("MOSTRADOR_ENTORNO", "desarrollo"),
		Direccion:    valor("MOSTRADOR_DIRECCION", ":4000"),
		BaseApp:      os.Getenv("MOSTRADOR_BASE_DE_DATOS"),
		BaseWorker:   os.Getenv("MOSTRADOR_BASE_DE_DATOS_WORKER"),
		BaseMigrador: os.Getenv("MOSTRADOR_BASE_DE_DATOS_MIGRADOR"),
		BaseAdmin:    os.Getenv("MOSTRADOR_BASE_DE_DATOS_ADMIN"),
		BaseStaff:    os.Getenv("MOSTRADOR_BASE_DE_DATOS_STAFF"),
		Sitio:        valor("MOSTRADOR_SITIO", "https://tulocalweb.com.ar"),
		ClaveBFF:     os.Getenv("MOSTRADOR_CLAVE_BFF"),
		ResendClave:  os.Getenv("MOSTRADOR_RESEND_CLAVE"),
		CorreoDe:     valor("MOSTRADOR_CORREO_DE", "Tu Local Web <hola@tulocalweb.com.ar>"),
		SMTP:         os.Getenv("MOSTRADOR_SMTP"),
		SMTPUsuario:  os.Getenv("MOSTRADOR_SMTP_USUARIO"),
		SMTPClave:    os.Getenv("MOSTRADOR_SMTP_CLAVE"),
		CorreoEquipo: os.Getenv("MOSTRADOR_CORREO_EQUIPO"),
		OAuth: map[string][2]string{
			"mercadopago":  {os.Getenv("MOSTRADOR_MP_CLIENT_ID"), os.Getenv("MOSTRADOR_MP_SECRETO")},
			"mercadolibre": {os.Getenv("MOSTRADOR_ML_CLIENT_ID"), os.Getenv("MOSTRADOR_ML_SECRETO")},
			"tiendanube":   {os.Getenv("MOSTRADOR_TN_CLIENT_ID"), os.Getenv("MOSTRADOR_TN_SECRETO")},
		},
		ClerkSecreto:       os.Getenv("MOSTRADOR_CLERK_SECRETO"),
		ClerkPublicable:    os.Getenv("MOSTRADOR_CLERK_PUBLICABLE"),
		ClerkOrigenes:      lista(os.Getenv("MOSTRADOR_CLERK_ORIGENES")),
		WhatsAppToken:      os.Getenv("MOSTRADOR_WHATSAPP_TOKEN"),
		WhatsAppNumeroID:   os.Getenv("MOSTRADOR_WHATSAPP_NUMERO_ID"),
		MPToken:            os.Getenv("MOSTRADOR_MP_TOKEN"),
		MPSecretoWebhook:   os.Getenv("MOSTRADOR_MP_SECRETO_WEBHOOK"),
		ClaveCifrado:       os.Getenv("MOSTRADOR_CLAVE_CIFRADO"),
		TelegramToken:      os.Getenv("MOSTRADOR_TELEGRAM_BOT_TOKEN"),
		CAUsuario:          os.Getenv("MOSTRADOR_CA_USUARIO"),
		CAClave:            os.Getenv("MOSTRADOR_CA_CLAVE"),
		CAPruebas:          os.Getenv("MOSTRADOR_CA_AMBIENTE") == "pruebas",
		GeoIP:              os.Getenv("MOSTRADOR_GEOIP"),
		APNsKeyID:          os.Getenv("MOSTRADOR_APNS_KEY_ID"),
		APNsTeamID:         os.Getenv("MOSTRADOR_APNS_TEAM_ID"),
		APNsClave:          os.Getenv("MOSTRADOR_APNS_CLAVE"),
		APNsTopic:          valor("MOSTRADOR_APNS_TOPIC", "ar.com.tulocalweb.app"),
		FCMCuenta:          os.Getenv("MOSTRADOR_FCM_CUENTA"),
		VAPIDPublica:       os.Getenv("MOSTRADOR_VAPID_PUBLICA"),
		VAPIDPrivada:       os.Getenv("MOSTRADOR_VAPID_PRIVADA"),
		VAPIDSujeto:        os.Getenv("MOSTRADOR_VAPID_SUJETO"),
		TiempoRealOrigenes: lista(os.Getenv("MOSTRADOR_TIEMPO_REAL_ORIGENES")),
		PurgaTiendas:       os.Getenv("MOSTRADOR_PURGA_TIENDAS") == "1",
		DirArchivos:        os.Getenv("MOSTRADOR_DIR_ARCHIVOS"),
	}
	switch c.Entorno {
	case "desarrollo", "staging", "produccion":
	default:
		return c, fmt.Errorf("MOSTRADOR_ENTORNO=%q: tiene que ser desarrollo, staging o produccion", c.Entorno)
	}
	if len(c.ClerkOrigenes) == 0 && !c.Desarrollo() {
		c.ClerkOrigenes = []string{c.Sitio}
	}
	if c.Desarrollo() {
		// Postgres de infra/compose.dev.yml. La base de la app es mostrador_dev (la crea el bootstrap
		// con el migrador como dueño); el admin entra a la base por defecto del contenedor.
		local := "postgres://%s@127.0.0.1:5440/%s?sslmode=disable"
		c.BaseApp = conDefecto(c.BaseApp, fmt.Sprintf(local, "mostrador_app", "mostrador_dev"))
		c.BaseWorker = conDefecto(c.BaseWorker, fmt.Sprintf(local, "mostrador_worker", "mostrador_dev"))
		c.BaseMigrador = conDefecto(c.BaseMigrador, fmt.Sprintf(local, "mostrador_migrador", "mostrador_dev"))
		c.BaseAdmin = conDefecto(c.BaseAdmin, fmt.Sprintf(local, "mostrador", "mostrador"))
		c.BaseStaff = conDefecto(c.BaseStaff, fmt.Sprintf(local, "mostrador_staff", "mostrador_dev"))
		c.SMTP = conDefecto(c.SMTP, "127.0.0.1:1026")
		c.Sitio = conDefecto(os.Getenv("MOSTRADOR_SITIO"), "http://localhost:5173")
		if len(c.ClerkOrigenes) == 0 {
			c.ClerkOrigenes = []string{c.Sitio}
		}
	}
	return c, nil
}

// ExigirCorreo falla si no hay por dónde mandar mails de verdad: Resend, o un SMTP con usuario y
// clave (nunca el Mailpit de desarrollo).
func (c Config) ExigirCorreo() error {
	if c.ResendClave != "" || (c.SMTP != "" && c.SMTPUsuario != "" && c.SMTPClave != "") {
		return nil
	}
	return fmt.Errorf("falta configurar el correo: MOSTRADOR_RESEND_CLAVE, o MOSTRADOR_SMTP con MOSTRADOR_SMTP_USUARIO y MOSTRADOR_SMTP_CLAVE")
}

// Exigir falla si falta alguna de las variables pedidas (por nombre de campo).
func (c Config) Exigir(campos ...string) error {
	vals := map[string]string{
		"BaseApp": c.BaseApp, "BaseWorker": c.BaseWorker, "BaseMigrador": c.BaseMigrador, "BaseAdmin": c.BaseAdmin, "BaseStaff": c.BaseStaff,
		"ResendClave": c.ResendClave, "ClaveBFF": c.ClaveBFF,
	}
	var faltan []string
	for _, f := range campos {
		if vals[f] == "" {
			faltan = append(faltan, f)
		}
	}
	if len(faltan) > 0 {
		return fmt.Errorf("falta configurar: %s", strings.Join(faltan, ", "))
	}
	return nil
}

// lista separa por comas y saca los vacíos.
func lista(v string) []string {
	var out []string
	for _, x := range strings.Split(v, ",") {
		if x = strings.TrimSpace(x); x != "" {
			out = append(out, x)
		}
	}
	return out
}

func valor(k, def string) string { return conDefecto(os.Getenv(k), def) }

func conDefecto(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
