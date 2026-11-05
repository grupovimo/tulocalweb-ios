package mercadopago

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/abono"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/limite"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

// Problema traduce los errores para el panel y la página de pago. Nunca devuelve lo que respondió
// Mercado Pago.
func (s *Servicio) Problema(w http.ResponseWriter, err error) {
	var emp *ErrorMP
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrApagado), errors.Is(err, ErrSinConexion):
		problema.Escribir(w, problema.Conflicto("sin_cobros", "La tienda todavía no conectó Mercado Pago."))
	case errors.Is(err, ErrReautorizar):
		problema.Escribir(w, problema.Conflicto("reconectar", "Hay que volver a conectar Mercado Pago."))
	case errors.Is(err, ErrYaPagado):
		problema.Escribir(w, problema.Conflicto("pagado", "Esto ya está pagado."))
	case errors.Is(err, ErrSinCaja):
		problema.Escribir(w, problema.Conflicto("sin_caja", "Falta crear la caja para cobrar con QR."))
	case errors.Is(err, ErrNoExiste):
		problema.Escribir(w, problema.NoEncontrado("No lo encontramos."))
	case errors.Is(err, ErrEstado):
		problema.Escribir(w, problema.Invalido("La conexión venció. Probá de nuevo."))
	case errors.Is(err, abono.ErrProgramado):
		problema.Escribir(w, problema.Conflicto("programado", "Ya tenés un cambio de plan programado: deshacelo para elegir otro."))
	case errors.Is(err, ErrInvalido), errors.Is(err, ErrPlanInvalido), errors.Is(err, abono.ErrInvalido):
		problema.Escribir(w, problema.Invalido("Revisá los datos."))
	case errors.As(err, &emp):
		s.log().Warn("mercadopago: la API respondió con error", "estado", emp.Estado, "cuerpo", recortar(emp.Cuerpo, 300))
		problema.Escribir(w, &problema.Problema{Estado: http.StatusBadGateway, Codigo: "mercadopago",
			Titulo: "Mercado Pago no respondió", Detalle: "Probá de nuevo en un rato."})
	default:
		s.log().Error("mercadopago", "error", err)
		problema.Escribir(w, problema.Interno())
	}
}

func escribir(w http.ResponseWriter, estado int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(estado)
	_ = json.NewEncoder(w).Encode(v)
}

func leer(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10)).Decode(v); err != nil {
		problema.Escribir(w, problema.Invalido("No entendimos el pedido."))
		return false
	}
	return true
}

// Rutas del panel, de la página de pago y del webhook. pagarLink la usa el paquete links.
func (s *Servicio) Rutas(mux *http.ServeMux) {
	// sesion: la tienda y la persona salen de la sesión. Conectar cobros, crear la caja y el abono son
	// cosas de la dueña o de quien administra.
	sesion := func(soloAdmin bool, h func(http.ResponseWriter, *http.Request, cuentas.Datos)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			d, ok := cuentas.SesionDe(r)
			if !ok || d.TiendaID == "" {
				problema.Escribir(w, problema.SinSesion())
				return
			}
			if soloAdmin && d.Rol != "duenio" && d.Rol != "admin" {
				problema.Escribir(w, problema.SinPermiso("Sólo la dueña o quien administra puede hacerlo."))
				return
			}
			h(w, r, d)
		}
	}

	mux.HandleFunc("GET /v1/panel/integraciones/mercadopago", sesion(false, func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		e, err := s.EstadoDe(r.Context(), d.TiendaID)
		if err != nil {
			s.Problema(w, err)
			return
		}
		escribir(w, http.StatusOK, e)
	}))
	mux.HandleFunc("POST /v1/panel/integraciones/mercadopago/conectar", sesion(true, func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		u, err := s.Conectar(r.Context(), d.TiendaID, d.Usuario.ID)
		if err != nil {
			s.Problema(w, err)
			return
		}
		escribir(w, http.StatusOK, map[string]string{"url": u})
	}))
	mux.HandleFunc("POST /v1/panel/integraciones/mercadopago/vuelta", sesion(true, func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		var b struct{ Codigo, Estado string }
		if leer(w, r, &b) {
			s.Problema(w, s.Volver(r.Context(), d.TiendaID, d.Usuario.ID, b.Codigo, b.Estado))
		}
	}))
	mux.HandleFunc("DELETE /v1/panel/integraciones/mercadopago", sesion(true, func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		s.Problema(w, s.Desconectar(r.Context(), d.TiendaID))
	}))
	mux.HandleFunc("POST /v1/panel/integraciones/mercadopago/caja", sesion(true, func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		var u Ubicacion
		if leer(w, r, &u) {
			s.Problema(w, s.CrearCaja(r.Context(), d.TiendaID, u))
		}
	}))
	mux.HandleFunc("POST /v1/panel/pedidos/{numero}/qr", sesion(false, func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		n, err := strconv.Atoi(r.PathValue("numero"))
		if err != nil {
			s.Problema(w, ErrNoExiste)
			return
		}
		q, err := s.CobrarQR(r.Context(), d.TiendaID, n)
		if err != nil {
			s.Problema(w, err)
			return
		}
		escribir(w, http.StatusCreated, q)
	}))
	mux.HandleFunc("GET /v1/panel/cobros/qr/{referencia}", sesion(false, func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		e, err := s.EstadoQR(r.Context(), d.TiendaID, r.PathValue("referencia"))
		if err != nil {
			s.Problema(w, err)
			return
		}
		escribir(w, http.StatusOK, map[string]string{"estado": e})
	}))
	mux.HandleFunc("DELETE /v1/panel/cobros/qr/{referencia}", sesion(false, func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		s.Problema(w, s.CancelarQR(r.Context(), d.TiendaID, r.PathValue("referencia")))
	}))
	mux.HandleFunc("POST /v1/panel/pedidos/{numero}/reembolso", sesion(true, func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		n, err := strconv.Atoi(r.PathValue("numero"))
		clave := r.Header.Get("Idempotency-Key")
		if err != nil || len(clave) < 8 || len(clave) > 255 {
			problema.Escribir(w, problema.Invalido("Falta la Idempotency-Key."))
			return
		}
		var b struct{ Monto int64 }
		if r.ContentLength > 0 && !leer(w, r, &b) {
			return
		}
		if b.Monto < 0 {
			s.Problema(w, ErrInvalido)
			return
		}
		s.Problema(w, s.Reembolsar(r.Context(), d.TiendaID, n, b.Monto, d.TiendaID+":"+clave))
	}))
	// El abono (§56): cualquier pase de plan o ciclo, con su cotización. url vacía: quedó hecho; si no,
	// hay que pasar por Mercado Pago (cargar la tarjeta del débito o pagar la diferencia).
	cambiarAbono := sesion(true, func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		var b struct{ Plan, Ciclo string }
		if !leer(w, r, &b) {
			return
		}
		u, err := s.CambiarAbono(r.Context(), d.TiendaID, d.Usuario.ID, b.Plan, b.Ciclo)
		if err != nil {
			s.Problema(w, err)
			return
		}
		escribir(w, http.StatusOK, map[string]string{"url": u})
	})
	mux.HandleFunc("POST /v1/panel/cuenta/abono", cambiarAbono)
	// La app de antes de §56 sólo pedía el débito: ahora pasa por la misma cotización.
	mux.HandleFunc("POST /v1/panel/cuenta/suscripcion", cambiarAbono)
	// La vuelta de Mercado Pago: confirma el pago o el débito sin esperar el aviso.
	mux.HandleFunc("POST /v1/panel/cuenta/abono/vuelta", sesion(true, func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		var b struct{ Pago, Suscripcion string }
		if !leer(w, r, &b) {
			return
		}
		if len(b.Pago) > 64 || len(b.Suscripcion) > 64 {
			s.Problema(w, ErrInvalido)
			return
		}
		s.Problema(w, s.VolverAbono(r.Context(), d.TiendaID, b.Pago, b.Suscripcion))
	}))

	mux.HandleFunc("DELETE /v1/panel/cuenta/suscripcion", sesion(true, func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		s.Problema(w, s.CancelarSuscripcion(r.Context(), d.TiendaID))
	}))

	// Webhook: primero la firma, después todo lo demás. Sin firma válida, 401 y nada se procesa.
	// Quien compra vuelve de Mercado Pago: el pago se confirma sin esperar el webhook.
	mux.HandleFunc("POST /v1/publico/tiendas/{slug}/pagos/{pago}", func(w http.ResponseWriter, r *http.Request) {
		// Tope por tienda: nadie llena la cola de consultas a Mercado Pago con ids inventados.
		if c, ok := s.Pool.(limite.Consultor); ok {
			if ok, err := limite.Permitido(r.Context(), c, "mp:vuelta:"+r.PathValue("slug"), 120, time.Hour); err == nil && !ok {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
		}
		if err := s.AvisarVuelta(r.Context(), r.PathValue("slug"), r.PathValue("pago")); err != nil {
			s.Problema(w, err)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	})
	mux.HandleFunc("POST /v1/webhooks/mercadopago", func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 64<<10))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var a Aviso
		if json.Unmarshal(b, &a) != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		dataID := r.URL.Query().Get("data.id")
		if dataID == "" {
			dataID = crudo(a.Data.ID)
		}
		if !FirmaValida(s.Config.SecretoWebhook, r.Header.Get("x-signature"), r.Header.Get("x-request-id"), dataID, s.ahora()) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if err := s.Recibir(r.Context(), a); err != nil {
			s.log().Error("mercadopago: webhook", "tipo", a.Type, "error", err)
			w.WriteHeader(http.StatusInternalServerError) // Mercado Pago reintenta
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}
