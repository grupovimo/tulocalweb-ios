package vidriera

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/devoluciones"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/newsletter"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/ventas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/limite"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

func escribir(w http.ResponseWriter, estado int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(estado)
	_ = json.NewEncoder(w).Encode(v)
}

// Rutas públicas de la tienda online (sin sesión: las pide el BFF de /t/{slug}).
func Rutas(mux *http.ServeMux, pool tx, mp Pagador, sitio string, dev *devoluciones.Servicio, ip func(*http.Request) string, correo Correo) {
	error404 := func(w http.ResponseWriter, err error) {
		var ec ErrCupon
		switch {
		case errors.As(err, &ec):
			problema.Escribir(w, problema.Conflicto("cupon", ec.Motivo))
		case errors.Is(err, ErrNoExiste):
			problema.Escribir(w, problema.NoEncontrado("Esta tienda no existe o no está publicada."))
		case errors.Is(err, ErrInvalido):
			problema.Escribir(w, problema.Invalido("Revisá los datos del pedido."))
		case errors.Is(err, ErrCorreo):
			problema.Escribir(w, problema.Conflicto("sin_cotizacion", "No pudimos cotizar el envío con Correo Argentino. Probá de nuevo o elegí otra forma de entrega."))
		case errors.Is(err, ErrSinStock):
			problema.Escribir(w, problema.Conflicto("sin_stock", "Algún producto se quedó sin stock."))
		case errors.Is(err, ErrProximamente):
			problema.Escribir(w, problema.Conflicto("proximamente", "Ese producto todavía no salió a la venta."))
		case errors.Is(err, ErrSinSena):
			problema.Escribir(w, problema.Conflicto("sin_sena", "Este producto no se reserva ni se encarga con seña."))
		case errors.Is(err, ErrMercadoPago):
			problema.Escribir(w, problema.Conflicto("mercadopago", "Mercado Pago no respondió. Probá de nuevo en un rato."))
		default:
			problema.Escribir(w, problema.Interno())
		}
	}
	// La tienda de un dominio propio: la web la pide al recibir mitienda.com.ar.
	mux.HandleFunc("GET /v1/publico/dominios/{dominio}", func(w http.ResponseWriter, r *http.Request) {
		slug, err := SlugDeDominio(r.Context(), pool, r.PathValue("dominio"))
		if err != nil {
			error404(w, err)
			return
		}
		escribir(w, http.StatusOK, map[string]string{"slug": slug})
	})
	// Una dirección vieja de una tienda (la cambió): la web responde 301 a la vigente.
	mux.HandleFunc("GET /v1/publico/slugs/{slug}", func(w http.ResponseWriter, r *http.Request) {
		slug, err := SlugVigente(r.Context(), pool, r.PathValue("slug"))
		if errors.Is(err, ErrNoExiste) {
			problema.Escribir(w, problema.NoEncontrado("Esa dirección no fue de ninguna tienda."))
			return
		}
		if err != nil {
			error404(w, err)
			return
		}
		escribir(w, http.StatusOK, map[string]string{"slug": slug})
	})
	// Para el sitemap de la web: lo público de todas las tiendas.
	mux.HandleFunc("GET /v1/publico/sitemap", func(w http.ResponseWriter, r *http.Request) {
		e, err := Sitemap(r.Context(), pool)
		if err != nil {
			problema.Escribir(w, problema.Interno())
			return
		}
		w.Header().Set("Cache-Control", "public, max-age=3600")
		escribir(w, http.StatusOK, e)
	})
	mux.HandleFunc("GET /v1/publico/tiendas/{slug}", func(w http.ResponseWriter, r *http.Request) {
		t, err := Publica(r.Context(), pool, r.PathValue("slug"))
		if err != nil {
			error404(w, err)
			return
		}
		escribir(w, http.StatusOK, t)
	})
	mux.HandleFunc("POST /v1/publico/tiendas/{slug}/pedidos", func(w http.ResponseWriter, r *http.Request) {
		var c Compra
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 32<<10)).Decode(&c) != nil {
			error404(w, ErrInvalido)
			return
		}
		res, err := ComprarConCorreo(r.Context(), pool, mp, correo, sitio, r.PathValue("slug"), c, time.Now())
		if err != nil && res.Numero == 0 {
			error404(w, err)
			return
		}
		// Si el pedido quedó pero Mercado Pago no respondió, igual se confirma: se puede pagar después.
		escribir(w, http.StatusCreated, res)
	})
	// Carrito abandonado: el checkout lo guarda cuando quien compra deja su email (tope por IP).
	mux.HandleFunc("POST /v1/publico/tiendas/{slug}/carritos", func(w http.ResponseWriter, r *http.Request) {
		if c, ok := pool.(limite.Consultor); ok && ip != nil {
			if ok, err := limite.Permitido(r.Context(), c, "tienda:carrito:"+ip(r), 60, time.Hour); err == nil && !ok {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
		}
		var b struct {
			Email   string                `json:"email"`
			Nombre  string                `json:"nombre"`
			Carrito []ventas.LineaCarrito `json:"carrito"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&b) != nil {
			error404(w, ErrInvalido)
			return
		}
		id, err := tiendaDe(r.Context(), pool, r.PathValue("slug"))
		if err == nil {
			err = ventas.GuardarCarrito(r.Context(), pool, id, b.Email, b.Nombre, b.Carrito)
		}
		if errors.Is(err, ventas.ErrCarritoInvalido) {
			error404(w, ErrInvalido)
			return
		}
		if err != nil {
			error404(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	// Newsletter (§50): sumarse desde el pie de la tienda. Queda pendiente hasta confirmar por mail.
	// Topes contra abuso: por IP y tienda, por tienda y por email (para no llenarle la casilla a nadie).
	mux.HandleFunc("POST /v1/publico/tiendas/{slug}/newsletter", func(w http.ResponseWriter, r *http.Request) {
		var b struct {
			Email string `json:"email"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&b) != nil {
			problema.Escribir(w, problema.Invalido("Revisá el email."))
			return
		}
		email, ok := newsletter.Email(b.Email)
		if !ok {
			problema.Escribir(w, problema.Invalido("Revisá el email.", problema.Campo{Ruta: "email", Mensaje: "Revisá el email."}))
			return
		}
		id, err := tiendaDe(r.Context(), pool, r.PathValue("slug"))
		if err != nil {
			error404(w, err)
			return
		}
		if c, ok := pool.(limite.Consultor); ok {
			type tope struct {
				clave   string
				max     int
				ventana time.Duration
			}
			claves := []tope{{"newsletter:tienda:" + id, 300, time.Hour}, {"newsletter:email:" + id + ":" + email, 3, 24 * time.Hour}}
			if ip != nil {
				claves = append(claves, tope{"newsletter:ip:" + id + ":" + ip(r), 10, time.Hour})
			}
			for _, l := range claves {
				if ok, err := limite.Permitido(r.Context(), c, l.clave, l.max, l.ventana); err == nil && !ok {
					w.Header().Set("Retry-After", "3600")
					problema.Escribir(w, &problema.Problema{Estado: http.StatusTooManyRequests, Codigo: "demasiados_envios",
						Titulo: "Esperá un rato", Detalle: "Recibimos varios pedidos seguidos. Probá de nuevo en un rato."})
					return
				}
			}
		}
		if err := newsletter.Suscribir(r.Context(), pool, id, email, "tienda"); err != nil {
			error404(w, err)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	})
	// El link del mail de confirmación y la baja de un clic (el token es el de cada suscripto).
	for _, accion := range []string{"confirmar", "baja"} {
		mux.HandleFunc("POST /v1/publico/tiendas/{slug}/newsletter/"+accion, func(w http.ResponseWriter, r *http.Request) {
			var b struct {
				Token string `json:"token"`
			}
			if json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<10)).Decode(&b) != nil {
				error404(w, ErrInvalido)
				return
			}
			id, err := tiendaDe(r.Context(), pool, r.PathValue("slug"))
			if err == nil {
				if accion == "confirmar" {
					err = newsletter.Confirmar(r.Context(), pool, id, b.Token)
				} else {
					err = newsletter.Baja(r.Context(), pool, id, b.Token)
				}
			}
			if errors.Is(err, newsletter.ErrNoExiste) {
				problema.Escribir(w, problema.NoEncontrado("Ese link ya no sirve: puede que ya te hayas dado de baja."))
				return
			}
			if err != nil {
				error404(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})
	}
	// El link del recordatorio: las líneas para volver a armar el carrito.
	mux.HandleFunc("GET /v1/publico/tiendas/{slug}/carritos/{token}", func(w http.ResponseWriter, r *http.Request) {
		id, err := tiendaDe(r.Context(), pool, r.PathValue("slug"))
		var ls []ventas.LineaCarrito
		if err == nil {
			ls, err = ventas.CarritoPorToken(r.Context(), pool, id, r.PathValue("token"))
		}
		if errors.Is(err, ventas.ErrNoExisteCarrito) {
			error404(w, ErrNoExiste)
			return
		}
		if err != nil {
			error404(w, err)
			return
		}
		escribir(w, http.StatusOK, map[string]any{"carrito": ls})
	})
	// Vista previa del cupón en el carrito y el checkout (al comprar se vuelve a evaluar).
	mux.HandleFunc("POST /v1/publico/tiendas/{slug}/cupon", func(w http.ResponseWriter, r *http.Request) {
		var c Compra
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 32<<10)).Decode(&c) != nil || len(c.Carrito) > 50 {
			error404(w, ErrInvalido)
			return
		}
		t, err := Publica(r.Context(), pool, r.PathValue("slug"))
		if err != nil {
			error404(w, err)
			return
		}
		id, _ := tiendaDe(r.Context(), pool, r.PathValue("slug"))
		porID := map[string]ProductoPublico{}
		for _, p := range t.Productos {
			porID[p.ID] = p
		}
		// Con los precios que se van a cobrar (variante y mayorista); un carrito con líneas que ya no
		// valen se evalúa con el precio del producto, como antes.
		precios, _, _ := preciosCarrito(t, porID, c.Carrito)
		res, err := EvaluarCupon(r.Context(), pool, id, strings.ToUpper(strings.TrimSpace(c.Cupon)), c.Email, lineasDe(c, porID, precios), time.Now())
		if err != nil {
			error404(w, err)
			return
		}
		escribir(w, http.StatusOK, res)
	})
	// Botón de arrepentimiento de la tienda (Ley 24.240): devuelve el código de trámite.
	mux.HandleFunc("POST /v1/publico/tiendas/{slug}/arrepentimiento", func(w http.ResponseWriter, r *http.Request) {
		if c, ok := pool.(limite.Consultor); ok {
			if ok, err := limite.Permitido(r.Context(), c, "tienda:arrepentimiento:"+ip(r), 10, time.Hour); err == nil && !ok {
				problema.Escribir(w, &problema.Problema{Estado: http.StatusTooManyRequests, Codigo: "demasiados",
					Titulo: "Demasiados pedidos", Detalle: "Probá de nuevo en un rato."})
				return
			}
		}
		var p devoluciones.Pedido
		if dev == nil || json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10)).Decode(&p) != nil {
			error404(w, ErrInvalido)
			return
		}
		t, err := Publica(r.Context(), pool, r.PathValue("slug"))
		if err != nil {
			error404(w, err)
			return
		}
		id, _ := tiendaDe(r.Context(), pool, r.PathValue("slug"))
		codigo, err := dev.Arrepentirse(r.Context(), id, t.Nombre, p, time.Now())
		switch {
		case errors.Is(err, devoluciones.ErrInvalido):
			problema.Escribir(w, problema.Invalido("Revisá tu nombre y tu email."))
		case errors.Is(err, devoluciones.ErrNoEncontrado):
			problema.Escribir(w, problema.NoEncontrado("No encontramos ese pedido. Revisá el número y que el email sea el de la compra."))
		case errors.Is(err, devoluciones.ErrFueraDePlazo):
			problema.Escribir(w, problema.Conflicto("fuera_de_plazo", "Pasaron más de 10 días desde que recibiste la compra. Escribile a la tienda para ver qué se puede hacer."))
		case err != nil:
			error404(w, err)
		default:
			escribir(w, http.StatusCreated, map[string]string{"codigo": codigo})
		}
	})
	// Reseñas y preguntas de quien compra (con tope por IP: no se puede inundar una tienda).
	tope := func(w http.ResponseWriter, r *http.Request, que string) bool {
		c, ok := pool.(limite.Consultor)
		if !ok {
			return true
		}
		if ok, err := limite.Permitido(r.Context(), c, "tienda:"+que+":"+ip(r), 10, time.Hour); err == nil && !ok {
			problema.Escribir(w, &problema.Problema{Estado: http.StatusTooManyRequests, Codigo: "demasiados",
				Titulo: "Demasiados envíos", Detalle: "Probá de nuevo en un rato."})
			return false
		}
		return true
	}
	mux.HandleFunc("POST /v1/publico/tiendas/{slug}/resenas", func(w http.ResponseWriter, r *http.Request) {
		if !tope(w, r, "resena") {
			return
		}
		var n ventas.NuevaResena
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10)).Decode(&n) != nil {
			error404(w, ErrInvalido)
			return
		}
		id, err := tiendaDe(r.Context(), pool, r.PathValue("slug"))
		if err == nil {
			err = ventas.DejarResena(r.Context(), pool, id, n)
		}
		switch {
		case errors.Is(err, ventas.ErrYaReseno):
			problema.Escribir(w, problema.Conflicto("ya_reseno", "Ya dejaste tu reseña de esa compra. ¡Gracias!"))
		case errors.Is(err, ventas.ErrResenaInvalida):
			error404(w, ErrInvalido)
		case err != nil:
			error404(w, err)
		default:
			w.WriteHeader(http.StatusCreated)
		}
	})
	mux.HandleFunc("POST /v1/publico/tiendas/{slug}/preguntas", func(w http.ResponseWriter, r *http.Request) {
		if !tope(w, r, "pregunta") {
			return
		}
		var q struct{ ProductoID, Nombre, Email, Pregunta string }
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10)).Decode(&q) != nil {
			error404(w, ErrInvalido)
			return
		}
		id, err := tiendaDe(r.Context(), pool, r.PathValue("slug"))
		if err == nil {
			err = ventas.Preguntar(r.Context(), pool, id, q.ProductoID, q.Nombre, q.Email, q.Pregunta)
		}
		if errors.Is(err, ventas.ErrResenaInvalida) {
			error404(w, ErrInvalido)
			return
		}
		if err != nil {
			error404(w, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
	})
	// Reservar un lanzamiento o encargar un producto con seña (§52). Tope por IP, como los turnos.
	mux.HandleFunc("POST /v1/publico/tiendas/{slug}/senas", func(w http.ResponseWriter, r *http.Request) {
		if !tope(w, r, "sena") {
			return
		}
		var s PedidoSena
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10)).Decode(&s) != nil {
			error404(w, ErrInvalido)
			return
		}
		senador, _ := mp.(Senador)
		res, err := Senar(r.Context(), pool, senador, sitio, r.PathValue("slug"), s, time.Now())
		if err != nil {
			error404(w, err)
			return
		}
		escribir(w, http.StatusCreated, res)
	})
	mux.HandleFunc("GET /v1/publico/tiendas/{slug}/pedidos/{codigo}", func(w http.ResponseWriter, r *http.Request) {
		c, err := VerPedido(r.Context(), pool, r.PathValue("slug"), r.PathValue("codigo"))
		if err != nil {
			error404(w, err)
			return
		}
		escribir(w, http.StatusOK, c)
	})
	// Un link de la tienda vieja (importada): 301 al path de Tu Local Web, relativo a la tienda. Lo pide
	// el BFF cuando un path de /t/{slug} no existe; el cuerpo repite el destino para no leer el header.
	mux.HandleFunc("GET /v1/publico/tiendas/{slug}/redireccion", func(w http.ResponseWriter, r *http.Request) {
		hacia, err := Redireccion(r.Context(), pool, r.PathValue("slug"), r.URL.Query().Get("path"))
		if errors.Is(err, ErrNoExiste) {
			problema.Escribir(w, problema.NoEncontrado("Esa dirección no existe en la tienda."))
			return
		}
		if err != nil {
			error404(w, err)
			return
		}
		w.Header().Set("Location", hacia)
		escribir(w, http.StatusMovedPermanently, map[string]string{"hacia": hacia})
	})
}
