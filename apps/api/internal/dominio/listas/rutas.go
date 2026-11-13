package listas

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/roles"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

// conPrecios dice si la sesión puede ver y tocar precios y costos (roles.Catalogo). Sin sesión en
// el contexto (tests del paquete con otra forma de dar la tienda), sí.
func conPrecios(r *http.Request) bool {
	d, ok := cuentas.SesionDe(r)
	return !ok || roles.Puede(d.Rol, roles.Catalogo)
}

// sinCostos saca el costo y lo que sale de él (margen, valor del stock) de una lista de productos
// o de inventario, para quien no tiene Catalogo.
func sinCostos(l Lista) Lista {
	for _, r := range l.Filas {
		delete(r, "costo")
		delete(r, "margen")
		delete(r, "lowm")
	}
	delete(l.Resumen, "valor")
	delete(l.Resumen, "valorTotal")
	return l
}

func escribir(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(v)
}

// sinInternos saca los campos que no van al cliente ("_ficha").
func sinInternos(l Lista) Lista {
	for i, r := range l.Filas {
		c := Registro{}
		for k, v := range r {
			if k != "_ficha" {
				c[k] = v
			}
		}
		l.Filas[i] = c
	}
	return l
}

// Rutas registra el motor de listas y los productos.
// fact puede ser nil (sin ARCA en el servidor): la lista de Facturación sale vacía.
func Rutas(mux *http.ServeMux, pool tx, tienda func(*http.Request) string, fact Facturas) {
	con := func(h func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			id := tienda(r)
			if id == "" {
				problema.Escribir(w, problema.SinSesion())
				return
			}
			h(w, r, id)
		}
	}
	mux.HandleFunc("GET /v1/panel/listas/{lista}", con(func(w http.ResponseWriter, r *http.Request, id string) {
		f := FiltrosDe(r.URL.Query())
		var regs []Registro
		var def Def
		var err error
		switch r.PathValue("lista") {
		case "clientes":
			regs, err = Clientes(r.Context(), pool, id, time.Now())
			def = DefClientes
		case "productos":
			regs, err = Productos(r.Context(), pool, id, time.Now())
			def = DefProductos
		case "inventario":
			var prods []Registro
			prods, err = Productos(r.Context(), pool, id, time.Now())
			regs, def = Inventario(prods), DefInventario
		case "envios":
			regs, err = Envios(r.Context(), pool, id, time.Now())
			def = DefEnvios
		case "facturacion":
			regs, err = Comprobantes(r.Context(), pool, id, time.Now())
			def = DefFacturacion
		default:
			problema.Escribir(w, problema.NoEncontrado("Esa lista no existe."))
			return
		}
		if err != nil {
			problema.Escribir(w, problema.Interno())
			return
		}
		l := sinInternos(Aplicar(regs, def, f))
		if lista := r.PathValue("lista"); (lista == "productos" || lista == "inventario") && !conPrecios(r) {
			l = sinCostos(l)
		}
		escribir(w, l)
	}))
	mux.HandleFunc("POST /v1/panel/listas/{lista}/acciones", con(func(w http.ResponseWriter, r *http.Request, id string) {
		var c struct {
			IDs    []string        `json:"ids"`
			Accion string          `json:"accion"`
			Valor  json.RawMessage `json:"valor"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&c); err != nil || len(c.IDs) > 500 {
			problema.Escribir(w, problema.Invalido("Elegí filas y una acción."))
			return
		}
		n, err := 0, error(nil)
		if r.PathValue("lista") == "productos" {
			switch c.Accion {
			case "publicar":
				var on bool
				_ = json.Unmarshal(c.Valor, &on)
				n, err = Publicar(r.Context(), pool, id, c.IDs, on)
			case "pausar":
				n, err = Publicar(r.Context(), pool, id, c.IDs, false)
			case "precio":
				if !conPrecios(r) {
					problema.Escribir(w, problema.SinPermiso(roles.Mensaje(roles.Catalogo)))
					return
				}
				var pct float64
				_ = json.Unmarshal(c.Valor, &pct)
				n, err = Precio(r.Context(), pool, id, c.IDs, pct)
			}
		}
		if r.PathValue("lista") == "inventario" {
			switch c.Accion {
			case "ajustar", "ingreso":
				var q int
				_ = json.Unmarshal(c.Valor, &q)
				n, err = AjustarStock(r.Context(), pool, id, c.IDs, q)
			case "conteo":
				var conteos []struct {
					ID       string `json:"id"`
					Cantidad int    `json:"cantidad"`
				}
				_ = json.Unmarshal(c.Valor, &conteos)
				for _, x := range conteos {
					m, e := FijarStock(r.Context(), pool, id, x.ID, x.Cantidad)
					n, err = n+m, errors.Join(err, e)
				}
			}
		}
		if r.PathValue("lista") == "envios" && (c.Accion == "etiquetas" || c.Accion == "sin-etiqueta") {
			n, err = Etiquetas(r.Context(), pool, id, c.IDs, c.Accion == "etiquetas")
		}
		if r.PathValue("lista") == "facturacion" && fact != nil {
			switch c.Accion {
			case "emitir":
				n, err = fact.EmitirManual(r.Context(), id, c.Valor, time.Now())
			case "reintentar":
				n, err = fact.Reintentar(r.Context(), id, c.IDs, time.Now())
			case "anular":
				n, err = fact.Anular(r.Context(), id, c.IDs, time.Now())
			default:
				problema.Escribir(w, problema.Invalido("Esa acción no está disponible en Facturación."))
				return
			}
			if err != nil && n == 0 {
				problema.Escribir(w, problema.Invalido(err.Error()))
				return
			}
			err = nil
		}
		if r.PathValue("lista") == "clientes" && c.Accion == "crear" {
			n, err = CrearCliente(r.Context(), pool, id, c.Valor)
		}
		// Rectificar (sus datos) y suprimir (el pedido de supresión de quien compró).
		if r.PathValue("lista") == "clientes" && (c.Accion == "editar" || c.Accion == "suprimir") {
			if c.Accion == "editar" && len(c.IDs) == 1 {
				n, err = EditarCliente(r.Context(), pool, id, c.IDs[0], c.Valor)
			} else if c.Accion == "suprimir" {
				n, err = SuprimirClientes(r.Context(), pool, id, c.IDs)
			}
			switch {
			case errors.Is(err, ErrEnCurso):
				problema.Escribir(w, problema.Conflicto("en_curso", "Tiene un pedido, un turno, una suscripción o una devolución en curso: terminalo o cancelalo primero."))
				return
			case errors.Is(err, ErrEmailEnUso):
				problema.Escribir(w, problema.Conflicto("email_en_uso", "Ese email ya es de otro cliente."))
				return
			}
		}
		// Etiquetas de clientes y mensajes llegan con WhatsApp: hasta entonces, 0.
		if err != nil {
			problema.Escribir(w, problema.Invalido(err.Error()))
			return
		}
		escribir(w, map[string]int{"hechos": n})
	}))
	// Los pedidos importados de otra plataforma de un cliente (su ficha).
	mux.HandleFunc("GET /v1/panel/clientes/{id}/historial", con(func(w http.ResponseWriter, r *http.Request, id string) {
		l, err := Historial(r.Context(), pool, id, r.PathValue("id"))
		if err != nil {
			problema.Escribir(w, problema.Interno())
			return
		}
		escribir(w, l)
	}))
	mux.HandleFunc("GET /v1/panel/inventario/movimientos", con(func(w http.ResponseWriter, r *http.Request, id string) {
		m, err := MovimientosHoy(r.Context(), pool, id, time.Now())
		if err != nil {
			problema.Escribir(w, problema.Interno())
			return
		}
		escribir(w, m)
	}))
	guardar := func(w http.ResponseWriter, r *http.Request, tiendaID, id string, estado int) {
		var e Editable
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 128<<10)).Decode(&e); err != nil {
			problema.Escribir(w, problema.Invalido("Mandá el producto."))
			return
		}
		e.SinPrecios = !conPrecios(r)
		p, err := Guardar(r.Context(), pool, tiendaID, id, e)
		switch {
		case errors.Is(err, ErrNoExiste):
			problema.Escribir(w, problema.NoEncontrado("Ese producto no existe."))
		case err != nil:
			problema.Escribir(w, problema.Invalido(err.Error()))
		default:
			w.WriteHeader(estado)
			escribir(w, p)
		}
	}
	mux.HandleFunc("POST /v1/panel/productos", con(func(w http.ResponseWriter, r *http.Request, id string) {
		guardar(w, r, id, "", http.StatusCreated)
	}))
	mux.HandleFunc("PUT /v1/panel/productos/{id}", con(func(w http.ResponseWriter, r *http.Request, id string) {
		guardar(w, r, id, r.PathValue("id"), http.StatusOK)
	}))
	mux.HandleFunc("DELETE /v1/panel/productos/{id}", con(func(w http.ResponseWriter, r *http.Request, id string) {
		switch err := Borrar(r.Context(), pool, id, r.PathValue("id")); {
		case errors.Is(err, ErrNoExiste):
			problema.Escribir(w, problema.NoEncontrado("Ese producto no existe."))
		case errors.Is(err, ErrEnMercadoLibre):
			problema.Escribir(w, problema.Conflicto("en_mercadolibre", "Está publicado en Mercado Libre: sacalo de ahí primero."))
		case err != nil:
			problema.Escribir(w, problema.Interno())
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	mux.HandleFunc("POST /v1/panel/productos/{id}/duplicar", con(func(w http.ResponseWriter, r *http.Request, id string) {
		p, err := Duplicar(r.Context(), pool, id, r.PathValue("id"), time.Now())
		switch {
		case errors.Is(err, ErrNoExiste):
			problema.Escribir(w, problema.NoEncontrado("Ese producto no existe."))
		case err != nil:
			problema.Escribir(w, problema.Invalido(err.Error()))
		default:
			w.WriteHeader(http.StatusCreated)
			escribir(w, p)
		}
	}))
	mux.HandleFunc("GET /v1/panel/productos/{id}", con(func(w http.ResponseWriter, r *http.Request, id string) {
		p, err := Obtener(r.Context(), pool, id, r.PathValue("id"), time.Now())
		switch {
		case errors.Is(err, ErrNoExiste):
			problema.Escribir(w, problema.NoEncontrado("Ese producto no existe."))
		case err != nil:
			problema.Escribir(w, problema.Interno())
		default:
			if !conPrecios(r) {
				p.Costo = 0
			}
			escribir(w, p)
		}
	}))
}
