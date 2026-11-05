// Comando api: el servidor HTTP de Tu Local Web para el VPS.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/arrepentimiento"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/equipo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/facturacion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/formularios"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/ia"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/importacion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/conversiones"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/correoargentino"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/mercadolibre"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/mercadopago"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/telegram"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/tiendanube"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/cifrado"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/clerk"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/config"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/geoip"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/servidor"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/tiemporeal"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/webpush"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/whatsapp"
)

func main() {
	if err := correr(); err != nil {
		slog.Error("la api no arrancó", "error", err)
		os.Exit(1)
	}
}

func correr() error {
	cfg, err := config.Leer()
	if err != nil {
		return err
	}
	if err := cfg.Exigir("BaseApp"); err != nil {
		return err
	}
	// Fuera de desarrollo hace falta un correo de verdad (Resend o el servidor propio) y la clave del
	// BFF: si faltan, la API no arranca (mejor que arrancar y no mandar los links de recuperación).
	if !cfg.Desarrollo() {
		if err := cfg.Exigir("ClaveBFF"); err != nil {
			return err
		}
		if err := cfg.ExigirCorreo(); err != nil {
			return err
		}
	}
	mails := correo.Elegir(cfg.ResendClave, cfg.SMTP, cfg.SMTPUsuario, cfg.SMTPClave, cfg.CorreoDe)
	// WhatsApp de la plataforma: con credenciales, Cloud API; en desarrollo, al log; si no, apagado.
	var wa whatsapp.Mensajero
	switch {
	case cfg.WhatsAppToken != "":
		wa = &whatsapp.Cloud{Token: cfg.WhatsAppToken, NumeroID: cfg.WhatsAppNumeroID,
			Plantillas: map[string]string{"ingreso": "mostrador_ingreso", "restablecer": "mostrador_clave"}}
	case cfg.Desarrollo():
		wa = whatsapp.Log{}
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app, err := db.Abrir(ctx, cfg.BaseApp)
	if err != nil {
		return err
	}
	defer app.Close()
	// Backoffice del equipo: con el rol mostrador_staff (lee todas las tiendas, no escribe en ninguna).
	var eq *equipo.Servicio
	if cfg.BaseStaff != "" {
		staff, err := db.Abrir(ctx, cfg.BaseStaff)
		if err != nil {
			return err
		}
		defer staff.Close()
		eq = &equipo.Servicio{Staff: staff, App: app, Sitio: cfg.Sitio, ClaveBFF: cfg.ClaveBFF}
	}

	// Ubicación de las sesiones: la base se abre una sola vez. Sin base o si no abre, las sesiones
	// van sin lugar (nunca uno inventado) y la API arranca igual.
	cs := &cuentas.Servicio{Pool: app, Correo: mails, Sitio: cfg.Sitio,
		OAuth: cuentas.Proveedores(cfg.OAuth), WhatsApp: wa, Clerk: clerkDe(cfg)}
	if geo, err := geoip.Abrir(cfg.GeoIP); err != nil {
		slog.Error("geoip: no abrió la base, las sesiones van sin lugar", "ruta", cfg.GeoIP, "error", err)
	} else if geo != nil {
		defer geo.Cerrar()
		cs.Geo = geo
	}

	// Tiempo real (§44): una conexión dedicada con LISTEN y el hub que reparte por tienda y rol.
	hub := tiemporeal.NuevoHub()
	go hub.Correr(ctx)
	go (&tiemporeal.Escucha{URL: cfg.BaseApp, Hub: hub}).Correr(ctx)

	srv := &http.Server{
		Addr: cfg.Direccion,
		Handler: servidor.Nuevo(servidor.Dependencias{
			App:             app,
			Cuentas:         cs,
			Arrepentimiento: &arrepentimiento.Servicio{Pool: app, Correo: mails, Equipo: cfg.CorreoEquipo},
			Formularios:     &formularios.Servicio{Pool: app, Correo: mails, Sitio: cfg.Sitio, Equipo: cfg.CorreoEquipo},
			ClaveBFF:        cfg.ClaveBFF,
			Correo:          mails,
			CorreoEquipo:    cfg.CorreoEquipo,
			Sitio:           cfg.Sitio,
			MercadoPago: mercadopago.DesdeConfig(app, cfg.OAuth["mercadopago"][0], cfg.OAuth["mercadopago"][1], cfg.MPToken,
				cfg.MPSecretoWebhook, cfg.Sitio, cfg.ClaveCifrado, slog.Default()),
			MercadoLibre: mercadolibre.DesdeConfig(app, cfg.OAuth["mercadolibre"][0], cfg.OAuth["mercadolibre"][1],
				cfg.Sitio, cfg.ClaveCifrado, slog.Default()),
			CorreoArgentino:    correoargentino.DesdeConfig(app, cfg.CAUsuario, cfg.CAClave, cfg.CAPruebas, cfg.ClaveCifrado, slog.Default()),
			IA:                 &ia.Servicio{Pool: app},
			Facturacion:        facturacionDe(app, cfg.ClaveCifrado),
			Telegram:           &telegram.Servicio{Pool: app},
			Equipo:             eq,
			Importacion:        importacionDe(app, cfg.ClaveCifrado, cfg.OAuth["tiendanube"]),
			Conversiones:       conversionesDe(app, cfg.ClaveCifrado),
			VAPIDPublica:       vapidPublica(cfg),
			TiempoReal:         hub,
			TiempoRealOrigenes: cfg.TiempoRealOrigenes,
			Desarrollo:         cfg.Desarrollo(),
		}),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       90 * time.Second,
	}
	// Al apagar, los WebSocket (conexiones tomadas, que Shutdown no espera) se cierran con «la API
	// se reinicia» y los navegadores reconectan solos.
	srv.RegisterOnShutdown(hub.Cerrar)
	errs := make(chan error, 1)
	go func() {
		slog.Info("api escuchando", "direccion", cfg.Direccion, "entorno", cfg.Entorno)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errs <- err
		}
	}()
	select {
	case err := <-errs:
		return err
	case <-ctx.Done():
	}
	apagar, cancelar := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancelar()
	return srv.Shutdown(apagar)
}

// clerkDe arma el ingreso con Clerk. Sin clave secreta queda apagado; con claves mal formadas también,
// con un aviso (el resto de la API sigue).
func clerkDe(cfg config.Config) *clerk.Cliente {
	if cfg.ClerkSecreto == "" {
		return nil
	}
	c, err := clerk.Nuevo(cfg.ClerkSecreto, cfg.ClerkPublicable, cfg.ClerkOrigenes)
	if err != nil {
		slog.Error("clerk: configuración inválida, ingreso con Clerk apagado", "error", err)
		return nil
	}
	return c
}

// facturacionDe arma la facturación con ARCA. Sin clave de cifrado la clave privada no se puede
// guardar: la facturación responde que falta configurarla, el resto de la API sigue.
func facturacionDe(app *db.Pool, claveCifrado string) *facturacion.Servicio {
	s := &facturacion.Servicio{Pool: app}
	if claveCifrado != "" {
		if c, err := cifrado.Nuevo(claveCifrado); err == nil {
			s.Cifra = c
		} else {
			slog.Error("facturacion: clave de cifrado inválida", "error", err)
		}
	}
	return s
}

// conversionesDe arma la API de conversiones de Meta. Sin clave de cifrado se puede leer el estado
// pero no guardar un token.
func conversionesDe(app *db.Pool, claveCifrado string) *conversiones.Servicio {
	s := &conversiones.Servicio{Pool: app}
	if claveCifrado != "" {
		if c, err := cifrado.Nuevo(claveCifrado); err == nil {
			s.Cifra = c
		}
	}
	return s
}

// importacionDe arma la importación con las fuentes que hay. Sin clave de cifrado se puede importar
// igual desde lo que no pide credencial (una tienda pública, un CSV); Tiendanube, no (su token se
// guarda cifrado). tn son el id y el secreto de la app de socio de Tiendanube.
func importacionDe(app *db.Pool, claveCifrado string, tn [2]string) *importacion.Servicio {
	s := &importacion.Servicio{Pool: app, Fuentes: importacion.Fuentes(importacion.Dependencias{}),
		Tiendanube: tiendanube.App{ID: tn[0], Secreto: tn[1]}}
	if claveCifrado != "" {
		if c, err := cifrado.Nuevo(claveCifrado); err == nil {
			s.Cifra = c
		}
	}
	return s
}

// vapidPublica: la clave pública del push web para el navegador, sólo si el par VAPID está completo
// y es válido (si no, el panel dice que los avisos al celular no están disponibles).
func vapidPublica(cfg config.Config) string {
	c, err := webpush.Desde(cfg.VAPIDPublica, cfg.VAPIDPrivada, cfg.VAPIDSujeto)
	if err != nil {
		slog.Error("push web: las claves VAPID no sirven, queda apagado", "error", err)
		return ""
	}
	return c.ClavePublica()
}
