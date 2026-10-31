// Comando worker: consume el outbox (eventos de dominio) con reintentos. Corre aparte de la API.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/avisos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/equipo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/facturacion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/ia"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/importacion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/notificaciones"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/push"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/ventas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/visitas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/conversiones"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/mercadolibre"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/mercadopago"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/telegram"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/apns"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/auditoria"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/cifrado"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/config"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/webpush"
)

func main() {
	if err := correr(); err != nil {
		slog.Error("el worker no arrancó", "error", err)
		os.Exit(1)
	}
}

func correr() error {
	cfg, err := config.Leer()
	if err != nil {
		return err
	}
	if err := cfg.Exigir("BaseWorker", "BaseApp"); err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	pool, err := db.Abrir(ctx, cfg.BaseWorker)
	if err != nil {
		return err
	}
	defer pool.Close()
	// Los manejadores usan el rol de la app (con la RLS de cada tienda); el worker sólo toma eventos.
	app, err := db.Abrir(ctx, cfg.BaseApp)
	if err != nil {
		return err
	}
	defer app.Close()
	mp := mercadopago.DesdeConfig(app, cfg.OAuth["mercadopago"][0], cfg.OAuth["mercadopago"][1], cfg.MPToken,
		cfg.MPSecretoWebhook, cfg.Sitio, cfg.ClaveCifrado, slog.Default())
	w := &outbox.Worker{
		Pool: pool,
		// Los dominios registran acá sus manejadores a medida que se suman.
		Manejadores: map[string]outbox.Manejador{},
	}
	ml := mercadolibre.DesdeConfig(app, cfg.OAuth["mercadolibre"][0], cfg.OAuth["mercadolibre"][1], cfg.Sitio,
		cfg.ClaveCifrado, slog.Default())
	// Mails a quien compra (confirmación, pago, en camino, retiro): el mismo correo que la API, y
	// como en la API, sin un correo de verdad fuera de desarrollo no arranca (no se manda a Mailpit en
	// producción).
	if !cfg.Desarrollo() {
		if err := cfg.ExigirCorreo(); err != nil {
			return err
		}
	}
	mails := correo.Elegir(cfg.ResendClave, cfg.SMTP, cfg.SMTPUsuario, cfg.SMTPClave, cfg.CorreoDe)
	mc := &ventas.MailsCarritos{Pool: app, Correo: mails, Sitio: cfg.Sitio}
	go func() {
		t := time.NewTicker(time.Hour)
		defer t.Stop()
		for {
			if _, err := ventas.RecordatoriosAutomaticos(ctx, app, time.Now()); err != nil && ctx.Err() == nil {
				slog.Error("carritos: recordatorios", "error", err)
			}
			select {
			case <-ctx.Done():
				return
			case <-t.C:
			}
		}
	}()
	av := &avisos.Servicio{Pool: app, Correo: mails, Sitio: cfg.Sitio}
	// Aviso de cada venta al Telegram de la dueña. Sin token, el manejador no manda nada.
	tg := &telegram.Servicio{Pool: app, Bot: &telegram.Bot{Token: cfg.TelegramToken}}
	// Importación desde otra plataforma: cada paso guarda una página y encola el siguiente.
	imp := &importacion.Servicio{Pool: app, Fuentes: importacion.Fuentes(importacion.Dependencias{MercadoLibre: ml})}
	if c, err := cifrado.Nuevo(cfg.ClaveCifrado); err == nil && cfg.ClaveCifrado != "" {
		imp.Cifra = c
	}
	// Push al iPhone por APNs. Sin MOSTRADOR_APNS_*, apagado (lo dice el log una vez); una clave mal
	// configurada no tira el worker: se anota y queda apagado.
	cliAPNs, err := apns.Desde(cfg.APNsKeyID, cfg.APNsTeamID, cfg.APNsClave, cfg.APNsTopic)
	if err != nil {
		slog.Error("push: la clave de APNs no sirve, no se mandan avisos al iPhone", "error", err)
	}
	// Push web al panel instalado como app (§43). Sin MOSTRADOR_VAPID_*, apagado; unas claves que no
	// sirven tampoco tiran el worker.
	cliWeb, err := webpush.Desde(cfg.VAPIDPublica, cfg.VAPIDPrivada, cfg.VAPIDSujeto)
	if err != nil {
		slog.Error("push: las claves VAPID no sirven, no se mandan avisos al panel instalado", "error", err)
	}
	// API de conversiones de Meta (§47): las compras con consentimiento, con el token de cada tienda.
	conv := &conversiones.Servicio{Pool: app}
	if c, err := cifrado.Nuevo(cfg.ClaveCifrado); err == nil && cfg.ClaveCifrado != "" {
		conv.Cifra = c
	}
	ps := push.Nuevo(app, cliAPNs, slog.Default()).ConWeb(cliWeb)
	for _, ms := range []map[string]outbox.Manejador{mp.Manejadores(), ml.Manejadores(), av.Manejadores(), mc.Manejadores(), tg.Manejadores(), equipo.Manejadores(mails), imp.Manejadores(),
		cuentas.Manejadores(mails, cfg.Sitio), ps.Manejadores(), conv.Manejadores()} {
		for tipo, m := range ms {
			w.Manejadores[tipo] = m
		}
	}
	// La campanita del panel (§45): los mismos eventos dejan además una notificación por persona.
	(&notificaciones.Generador{Pool: app}).Envolver(w.Manejadores)
	// Limpieza de las IA conectadas: códigos vencidos y clientes que nadie usa hace 30 días.
	go func() {
		iaS := &ia.Servicio{Pool: app}
		t := time.NewTicker(6 * time.Hour)
		defer t.Stop()
		for {
			if err := iaS.Purgar(ctx); err != nil && ctx.Err() == nil {
				slog.Error("ia: purgar", "error", err)
			}
			select {
			case <-ctx.Done():
				return
			case <-t.C:
			}
		}
	}()
	// Visitas de las tiendas: cada hora se borran los vistos de días anteriores (con la sal del día
	// perdida ya no sirven) y las visitas de hace más de 400 días.
	go func() {
		t := time.NewTicker(time.Hour)
		defer t.Stop()
		for {
			if _, err := visitas.Purgar(ctx, app); err != nil && ctx.Err() == nil {
				slog.Error("visitas: purgar", "error", err)
			}
			select {
			case <-ctx.Done():
				return
			case <-t.C:
			}
		}
	}()
	// Notificaciones del panel: se guardan 90 días; una vez por día se borra lo más viejo.
	go func() {
		t := time.NewTicker(24 * time.Hour)
		defer t.Stop()
		for {
			if _, err := notificaciones.Purgar(ctx, app); err != nil && ctx.Err() == nil {
				slog.Error("notificaciones: purgar", "error", err)
			}
			select {
			case <-ctx.Done():
				return
			case <-t.C:
			}
		}
	}()
	// Actividad de las tiendas: se guarda un año; una vez por día se borra lo más viejo.
	go func() {
		t := time.NewTicker(24 * time.Hour)
		defer t.Stop()
		for {
			if _, err := auditoria.Purgar(ctx, app); err != nil && ctx.Err() == nil {
				slog.Error("actividad: purgar", "error", err)
			}
			select {
			case <-ctx.Done():
				return
			case <-t.C:
			}
		}
	}()
	// Facturación: el drenador pide los CAE de las ventas cobradas (cada 30 s, con espera creciente
	// por comprobante). Corre en el servidor propio.
	go func() {
		f := &facturacion.Servicio{Pool: app}
		if c, err := cifrado.Nuevo(cfg.ClaveCifrado); err == nil && cfg.ClaveCifrado != "" {
			f.Cifra = c
		} else {
			slog.Error("facturacion: sin clave de cifrado válida, las facturas quedan pendientes", "error", err)
		}
		t := time.NewTicker(30 * time.Second)
		defer t.Stop()
		for {
			if err := f.Drenar(ctx, time.Now()); err != nil && ctx.Err() == nil {
				slog.Error("facturacion: drenar", "error", err)
			}
			// El aviso de vencimiento del certificado (§48): barato, va en la misma vuelta.
			if err := f.AvisarVencimientos(ctx, time.Now()); err != nil && ctx.Err() == nil {
				slog.Error("facturacion: avisar vencimientos", "error", err)
			}
			select {
			case <-ctx.Done():
				return
			case <-t.C:
			}
		}
	}()
	slog.Info("worker consumiendo el outbox", "entorno", cfg.Entorno)
	return w.Correr(ctx)
}
