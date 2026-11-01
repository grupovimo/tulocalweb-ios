// Package vidriera: la tienda online pública (/t/{slug}). Lo publicado, sin costos ni datos de
// clientes; comprar arma el pedido con precios y envío calculados acá, y con Mercado Pago devuelve el
// checkout de la cuenta de la tienda.
package vidriera

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/configuracion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cupones"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/mayorista"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/newsletter"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/suscripciones"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/turnos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/ventas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/arca"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/conversiones"
)

var (
	ErrNoExiste = errors.New("vidriera: no existe")
	ErrInvalido = errors.New("vidriera: pedido inválido")
	ErrSinStock = errors.New("vidriera: sin stock")
	// ErrProximamente: un producto del carrito todavía no salió a la venta (§52).
	ErrProximamente = errors.New("vidriera: todavía no salió")
)

type tx interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
	EnTx(context.Context, func(pgx.Tx) error) error
}

// Pagador cobra un pedido con Mercado Pago (el paquete mercadopago).
type Pagador interface {
	PagarPedido(ctx context.Context, tiendaID string, numero int, vuelta string) (string, error)
}

type ProductoPublico struct {
	ID string `json:"id"`
	// Slug: el link de la tienda (/producto/<slug>), del nombre y fijo desde que se creó.
	Slug        string  `json:"slug"`
	Nombre      string  `json:"nombre"`
	Descripcion string  `json:"descripcion"`
	Categoria   *string `json:"categoria"`
	Precio      int64   `json:"precio"`
	// Mayorista: el precio con el carrito al mínimo mayorista (§49); null si no tiene o si la tienda
	// no vende por mayor.
	Mayorista *int64  `json:"mayorista"`
	Anterior  *int64  `json:"anterior"`
	Foto      *string `json:"foto"`
	Etiqueta  *string `json:"etiqueta"`
	// Stock: con opciones, la suma de las variantes activas.
	Stock     int               `json:"stock"`
	Fotos     []string          `json:"fotos"`
	Opciones  []listas.Opcion   `json:"opciones"`
	Variantes []VariantePublica `json:"variantes"`
	// Lanzamiento (§52): la fecha y hora en que sale a la venta, sólo mientras no llegó (antes se ve
	// «Próximamente» y no se compra). Reserva: el % de seña para reservarlo antes, si la tienda lo
	// ofrece y cobra con Mercado Pago. Encargo: el % de seña para encargarlo ya lanzado (el resto
	// al retirar o recibir). OfertaHasta: cuándo termina la oferta (Precio es el de oferta y
	// Anterior el de lista); null si no tiene fin o no hay oferta.
	Lanzamiento *time.Time `json:"lanzamiento"`
	Reserva     *int       `json:"reserva"`
	Encargo     *int       `json:"encargo"`
	OfertaHasta *time.Time `json:"ofertaHasta"`
}

// VariantePublica: sólo las activas, sin SKU. Precio nil = el del producto.
type VariantePublica struct {
	ID      string            `json:"id"`
	Valores map[string]string `json:"valores"`
	Precio  *int64            `json:"precio"`
	// Mayorista nil: el mayorista del producto (si tiene).
	Mayorista *int64  `json:"mayorista"`
	Stock     int     `json:"stock"`
	Foto      *string `json:"foto"`
}

type Categoria struct {
	Nombre   string  `json:"nombre"`
	Cantidad int     `json:"cantidad"`
	Foto     *string `json:"foto"`
}

// Tienda es components/schemas/TiendaPublica (lib/tienda/tipos.ts).
type Tienda struct {
	Slug      string         `json:"slug"`
	Nombre    string         `json:"nombre"`
	Plantilla string         `json:"plantilla"`
	Paleta    int            `json:"paleta"`
	Secciones []string       `json:"secciones"`
	Portada   map[string]any `json:"portada"`
	// Logo subido en Mi tienda (URL); vacío: el nombre en letras.
	Logo string `json:"logo"`
	// Compartir: título, descripción y foto propia de la vista previa al compartir (vacíos: lo sugerido).
	Compartir Compartir `json:"compartir"`
	Ciudad    string    `json:"ciudad"`
	Direccion string    `json:"direccion"`
	Whatsapp  string    `json:"whatsapp"`
	// Política de cambios y devoluciones (Configuración › Negocio); vacía: el texto legal base.
	Politica    string            `json:"politica"`
	Categorias  []Categoria       `json:"categorias"`
	Productos   []ProductoPublico `json:"productos"`
	MasVendidos []string          `json:"masVendidos"`
	Resenas     []ResenaPublica   `json:"resenas"`
	// Preguntas respondidas, por producto.
	Preguntas []PreguntaPublica `json:"preguntas"`
	Promedio  *float64          `json:"promedio"`
	Fotos     []string          `json:"fotos"`
	Cobros    map[string]any    `json:"cobros"`
	Envios    Envios            `json:"envios"`
	Vitrina   bool              `json:"vitrina"`
	// Turnos: servicios, profesionales, seña y horario; null si la tienda no toma turnos.
	Turnos *turnos.Resumen `json:"turnos"`
	// Suscripciones: los planes activos (§51); vacío si no tiene o no conectó Mercado Pago.
	Suscripciones []suscripciones.PlanPublico `json:"suscripciones"`
	Medicion      Medicion                    `json:"medicion"`
	// Dominio: el propio que eligió como principal, ya verificado (null: sólo /t/<slug>).
	Dominio *string `json:"dominio"`
	// Mayorista: el mínimo del carrito para los precios mayoristas (§49); null si no vende por mayor.
	Mayorista *MinimoMayorista `json:"mayorista"`
}

// MinimoMayorista: unidades y/o monto (a precio minorista); con los dos, tienen que cumplirse los dos.
type MinimoMayorista struct {
	Unidades *int   `json:"unidades"`
	Monto    *int64 `json:"monto"`
}

// Compartir es lo que la tienda escribió en Mi tienda › Al compartir.
type Compartir struct {
	Titulo      string `json:"titulo"`
	Descripcion string `json:"descripcion"`
	Foto        string `json:"foto"`
}

// Medicion son los IDs de GA4 y del píxel de Meta de la tienda (la tienda online los carga sólo con
// el consentimiento de quien compra).
type Medicion struct {
	GA4   *string `json:"ga4"`
	Pixel *string `json:"pixel"`
}

type ResenaPublica struct {
	Texto     string  `json:"texto"`
	Autor     string  `json:"autor"`
	Producto  string  `json:"producto"`
	Estrellas int     `json:"estrellas"`
	ConCompra bool    `json:"conCompra"`
	Respuesta *string `json:"respuesta"`
}

type PreguntaPublica struct {
	ProductoID string `json:"productoId"`
	Pregunta   string `json:"pregunta"`
	Respuesta  string `json:"respuesta"`
}

// Envios: lo que la tienda configuró en Envíos y retiro (configuracion.Entregas) más Correo y el
// envío gratis. Sin retiro, la dirección y el horario van vacíos; sin moto, su costo, zona y demora.
type Envios struct {
	GratisDesde     *int64 `json:"gratisDesde"`
	Retiro          bool   `json:"retiro"`
	RetiroDireccion string `json:"retiroDireccion"`
	RetiroHorario   string `json:"retiroHorario"`
	Moto            bool   `json:"moto"`
	MotoCosto       int64  `json:"motoCosto"`
	MotoZona        string `json:"motoZona"`
	MotoDemora      string `json:"motoDemora"`
	Correo          bool   `json:"correo"`
}

// Cuotas sin interés: ninguna (0), porque las absorbe la tienda y no se le promete al que compra lo
// que la tienda no eligió; las plantillas no las muestran con menos de 2. El descuento por
// transferencia y el envío gratis salen de las promociones automáticas que prendió la tienda
// (cupones.TransferenciaPct, cupones.EnvioGratisDesde): la misma fuente que muestra el panel.
const cuotas = 0

var slugValido = regexp.MustCompile(`^[a-z0-9-]{2,60}$`)

func tiendaDe(ctx context.Context, pool tx, slug string) (string, error) {
	if !slugValido.MatchString(slug) {
		return "", ErrNoExiste
	}
	var id *string
	if err := pool.EnTx(ctx, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT app.tienda_de_slug($1)::text`, slug).Scan(&id)
	}); err != nil {
		return "", err
	}
	if id == nil {
		return "", ErrNoExiste
	}
	return *id, nil
}

// Publica arma la tienda como la ve quien compra.
func Publica(ctx context.Context, pool tx, slug string) (Tienda, error) {
	id, err := tiendaDe(ctx, pool, slug)
	if err != nil {
		return Tienda{}, err
	}
	var miPortada struct{ Titulo, Lema, Portada string }
	out := Tienda{Slug: slug, Secciones: []string{"hero", "cats", "prods", "revs", "ig", "wa"}, Resenas: []ResenaPublica{}, Preguntas: []PreguntaPublica{},
		Fotos: []string{}, MasVendidos: []string{}}
	// Los más vendidos de los últimos 30 días (los ítems de los pedidos guardan el nombre).
	vendidos := map[string]int{}
	err = pool.EnTienda(ctx, id, func(t pgx.Tx) error {
		var diseno, negocio []byte
		var mp bool
		var autos map[string]bool
		if err := t.QueryRow(ctx, `SELECT nombre, plantilla, diseno_publicado, negocio,
				EXISTS (SELECT 1 FROM mp_conexiones WHERE tienda_id = $1 AND NOT reautorizar), ga4, pixel_meta, automaticas
			FROM tiendas WHERE id = $1`, id).Scan(&out.Nombre, &out.Plantilla, &diseno, &negocio, &mp,
			&out.Medicion.GA4, &out.Medicion.Pixel, &autos); err != nil {
			return err
		}
		// Correo Argentino se ofrece sólo con la cuenta de MiCorreo conectada (cotiza la API, nunca una tabla).
		if err := t.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM ca_conexiones WHERE tienda_id = $1)`, id).Scan(&out.Envios.Correo); err != nil {
			return err
		}
		// Retiro y moto: lo que configuró la tienda (sin configurar, retiro en el local con la dirección del negocio).
		ent, err := configuracion.EntregasEnTx(ctx, t, id)
		if err != nil {
			return err
		}
		if ent.Retiro {
			out.Envios.Retiro, out.Envios.RetiroDireccion, out.Envios.RetiroHorario = true, ent.RetiroDireccion, ent.RetiroHorario
		}
		if ent.Moto {
			out.Envios.Moto, out.Envios.MotoCosto, out.Envios.MotoZona, out.Envios.MotoDemora = true, ent.MotoCosto, ent.MotoZona, ent.MotoDemora
		}
		if err := t.QueryRow(ctx, `SELECT nombre FROM dominios WHERE tienda_id = $1 AND principal AND verificado_en IS NOT NULL`,
			id).Scan(&out.Dominio); err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		var d struct {
			Secciones []struct {
				ID      string `json:"id"`
				Visible bool   `json:"visible"`
			} `json:"secciones"`
			Paleta                      int `json:"paleta"`
			Titulo, Lema, Portada, Logo string
			CompartirTitulo             string `json:"compartirTitulo"`
			CompartirDescripcion        string `json:"compartirDescripcion"`
			CompartirFoto               string `json:"compartirFoto"`
		}
		if json.Unmarshal(diseno, &d) == nil {
			miPortada.Titulo, miPortada.Lema, miPortada.Portada = d.Titulo, d.Lema, d.Portada
			out.Logo = d.Logo
			out.Compartir = Compartir{d.CompartirTitulo, d.CompartirDescripcion, d.CompartirFoto}
		}
		if len(d.Secciones) > 0 {
			out.Secciones = out.Secciones[:0]
			for _, s := range d.Secciones {
				if s.Visible {
					out.Secciones = append(out.Secciones, s.ID)
				}
			}
			out.Paleta = d.Paleta
		}
		var n struct{ Direccion, Whatsapp, Alias, CBU, Titular, Politica string }
		_ = json.Unmarshal(negocio, &n)
		out.Direccion, out.Whatsapp, out.Politica = n.Direccion, n.Whatsapp, n.Politica
		out.Cobros = map[string]any{"transferencia": 0, "cuotas": cuotas, "mercadoPago": mp, "datosTransferencia": nil}
		// Sin alias ni CBU cargados no se ofrece transferencia (nunca un alias inventado), ni su descuento.
		if n.Alias != "" || n.CBU != "" {
			out.Cobros["datosTransferencia"] = map[string]string{"alias": n.Alias, "cbu": n.CBU, "titular": n.Titular}
			if autos["transf"] {
				out.Cobros["transferencia"] = cupones.TransferenciaPct
			}
		}
		// Factura A desde el checkout: la tienda es responsable inscripta y conectó ARCA.
		var facturaA bool
		if err := t.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM arca_config WHERE tienda_id = $1
				AND certificado IS NOT NULL AND condicion = 'responsable_inscripto')`, id).Scan(&facturaA); err != nil {
			return err
		}
		out.Cobros["facturaA"] = facturaA
		// El envío gratis sólo si la tienda envía (moto o Correo): con sólo retiro no hay envío que regalar.
		if autos["envio"] && (out.Envios.Moto || out.Envios.Correo) {
			g := int64(cupones.EnvioGratisDesde)
			out.Envios.GratisDesde = &g
		}
		if out.Turnos, err = turnos.ResumenDe(ctx, t, id, mp); err != nil {
			return err
		}
		if out.Suscripciones, err = suscripciones.PlanesDe(ctx, t, id, mp); err != nil {
			return err
		}
		may, err := mayorista.Leer(ctx, t, id)
		if err != nil {
			return err
		}
		if may.Activa {
			out.Mayorista = &MinimoMayorista{Unidades: may.Unidades, Monto: may.Monto}
		}
		// Sin venta mayorista, los precios mayoristas no se publican.
		soloSi := func(p *int64) *int64 {
			if may.Activa {
				return p
			}
			return nil
		}

		variantes := map[string][]listas.Variante{}
		vs, err := t.Query(ctx, `SELECT producto_id::text, id::text, valores, precio, stock, foto, precio_mayorista FROM variantes
			WHERE tienda_id = $1 AND activa`, id)
		if err != nil {
			return err
		}
		for vs.Next() {
			var p string
			v := listas.Variante{Activa: true}
			if err := vs.Scan(&p, &v.ID, &v.Valores, &v.Precio, &v.Stock, &v.Foto, &v.PrecioMayorista); err != nil {
				vs.Close()
				return err
			}
			variantes[p] = append(variantes[p], v)
		}
		if err := vs.Err(); err != nil {
			return err
		}
		filas, err := t.Query(ctx, `SELECT id::text, slug, nombre, descripcion, categoria, precio, foto, stock, fotos, opciones,
				lanzamiento, reserva_sena, encargo_sena, oferta_precio, oferta_hasta, precio_mayorista FROM productos
			WHERE tienda_id = $1 AND publicado AND coalesce((canales->>'web')::boolean, true) ORDER BY creado_en DESC LIMIT 500`, id)
		if err != nil {
			return err
		}
		ahora := time.Now()
		for filas.Next() {
			var p ProductoPublico
			var l lanzamiento
			if err := filas.Scan(&p.ID, &p.Slug, &p.Nombre, &p.Descripcion, &p.Categoria, &p.Precio, &p.Foto, &p.Stock, &p.Fotos, &p.Opciones,
				&l.fecha, &l.reserva, &l.encargo, &l.oferta, &l.hasta, &p.Mayorista); err != nil {
				return err
			}
			p.Mayorista = soloSi(p.Mayorista)
			p.Variantes = []VariantePublica{}
			if len(p.Opciones) > 0 {
				vs := variantes[p.ID]
				listas.OrdenarVariantes(p.Opciones, vs)
				for _, v := range vs {
					p.Variantes = append(p.Variantes, VariantePublica{ID: v.ID, Valores: v.Valores, Precio: v.Precio, Stock: v.Stock, Foto: v.Foto,
						Mayorista: soloSi(v.PrecioMayorista)})
				}
			}
			if p.Stock > 0 && p.Stock <= 3 {
				u := "Últimos"
				p.Etiqueta = &u
			}
			l.aplicar(&p, mp, ahora)
			out.Productos = append(out.Productos, p)
		}
		if err := filas.Err(); err != nil {
			return err
		}
		// Reseñas visibles (las 60 últimas) y su promedio; preguntas respondidas.
		var suma, cant int
		rs, err := t.Query(ctx, `SELECT r.texto, r.autor, coalesce(p.nombre, ''), r.estrellas, r.pedido_numero IS NOT NULL, r.respuesta
			FROM resenas r LEFT JOIN productos p ON p.tienda_id = r.tienda_id AND p.id = r.producto_id
			WHERE r.tienda_id = $1 AND NOT r.oculta ORDER BY r.creada DESC LIMIT 60`, id)
		if err != nil {
			return err
		}
		for rs.Next() {
			var r ResenaPublica
			if err := rs.Scan(&r.Texto, &r.Autor, &r.Producto, &r.Estrellas, &r.ConCompra, &r.Respuesta); err != nil {
				rs.Close()
				return err
			}
			out.Resenas = append(out.Resenas, r)
			suma, cant = suma+r.Estrellas, cant+1
		}
		rs.Close()
		if cant > 0 {
			prom := float64(suma) / float64(cant)
			out.Promedio = &prom
		}
		qs, err := t.Query(ctx, `SELECT producto_id::text, pregunta, respuesta FROM preguntas
			WHERE tienda_id = $1 AND respuesta IS NOT NULL ORDER BY respondida_en DESC LIMIT 200`, id)
		if err != nil {
			return err
		}
		for qs.Next() {
			var q PreguntaPublica
			if err := qs.Scan(&q.ProductoID, &q.Pregunta, &q.Respuesta); err != nil {
				qs.Close()
				return err
			}
			out.Preguntas = append(out.Preguntas, q)
		}
		qs.Close()
		rows, err := t.Query(ctx, `SELECT i->>'nombre', sum((i->>'cantidad')::int) FROM pedidos, jsonb_array_elements(items) i
			WHERE tienda_id = $1 AND estado <> 'Cancelado' AND creado_en > now() - interval '30 days' GROUP BY 1`, id)
		if err != nil {
			return err
		}
		for rows.Next() {
			var nombre string
			var c int
			if err := rows.Scan(&nombre, &c); err != nil {
				return err
			}
			vendidos[nombre] = c
		}
		return rows.Err()
	})
	if err != nil {
		return Tienda{}, err
	}
	if out.Productos == nil {
		out.Productos = []ProductoPublico{}
	}
	orden := append([]ProductoPublico(nil), out.Productos...)
	sort.SliceStable(orden, func(i, j int) bool { return vendidos[orden[i].Nombre] > vendidos[orden[j].Nombre] })
	cats := map[string]*Categoria{}
	var nombres []string
	for _, p := range orden {
		out.MasVendidos = append(out.MasVendidos, p.ID)
		if p.Foto != nil && len(out.Fotos) < 6 && !contiene(out.Fotos, *p.Foto) {
			out.Fotos = append(out.Fotos, *p.Foto)
		}
		if p.Categoria != nil && *p.Categoria != "" {
			c, ok := cats[*p.Categoria]
			if !ok {
				c = &Categoria{Nombre: *p.Categoria}
				cats[*p.Categoria] = c
				nombres = append(nombres, *p.Categoria)
			}
			c.Cantidad++
			if c.Foto == nil {
				c.Foto = p.Foto
			}
		}
	}
	for _, n := range nombres {
		out.Categorias = append(out.Categorias, *cats[n])
	}
	if out.Categorias == nil {
		out.Categorias = []Categoria{}
	}
	var foto *string
	if len(out.Fotos) > 0 {
		foto = &out.Fotos[0]
	}
	if miPortada.Portada != "" {
		foto = &miPortada.Portada
	}
	titulo := out.Nombre
	if miPortada.Titulo != "" {
		titulo = miPortada.Titulo
	}
	out.Portada = map[string]any{"titulo": titulo, "lema": miPortada.Lema, "foto": foto}
	return out, nil
}

func contiene(l []string, s string) bool {
	for _, x := range l {
		if x == s {
			return true
		}
	}
	return false
}

// LineaCarrito: un producto (y, si tiene opciones, su variante) del carrito.
type LineaCarrito struct {
	ID         string `json:"id"`
	VarianteID string `json:"varianteId"`
	Cantidad   int    `json:"cantidad"`
}

// Compra es lo que manda el checkout (precios y envío NO vienen: se calculan acá).
type Compra struct {
	Carrito   []LineaCarrito `json:"carrito"`
	Email     string         `json:"email"`
	Nombre    string         `json:"nombre"`
	Telefono  string         `json:"telefono"`
	DNI       string         `json:"dni"`
	CP        string         `json:"cp"`
	Direccion string         `json:"direccion"`
	Envio     string         `json:"envio"`
	Pago      string         `json:"pago"`
	Cupon     string         `json:"cupon"`
	// Correo Argentino: provincia (código de MiCorreo), localidad (a domicilio) y sucursal elegida.
	Provincia string `json:"provincia"`
	Localidad string `json:"localidad"`
	Sucursal  string `json:"sucursal"`
	// Factura A: sólo si la tienda la ofrece (cobros.facturaA); la razón social, si no viene, es el nombre.
	CUIT        string `json:"cuit"`
	RazonSocial string `json:"razonSocial"`
	// Medicion: sólo si quien compra aceptó las cookies de medición de la tienda (lo arma el BFF con
	// su navegador). Con la API de conversiones conectada, la compra se manda a Meta (§47).
	Medicion *conversiones.Navegador `json:"medicion"`
	// Novedades: marcó «Quiero recibir novedades» (plantillas sin bloque de newsletter, §50). Queda
	// pendiente hasta que confirme por mail, como desde el pie.
	Novedades bool `json:"novedades"`
}

// Resultado de comprar: el pedido, su código público y (con Mercado Pago) adónde pagar.
type Resultado struct {
	Numero int    `json:"numero"`
	Codigo string `json:"codigo"`
	Total  int64  `json:"total"`
	URL    string `json:"url,omitempty"`
}

var emailValido = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// Comprar arma el pedido de la tienda online (sin Correo Argentino).
func Comprar(ctx context.Context, pool tx, mp Pagador, sitio, slug string, c Compra, hoy time.Time) (Resultado, error) {
	return ComprarConCorreo(ctx, pool, mp, nil, sitio, slug, c, hoy)
}

// ComprarConCorreo arma el pedido; con Correo Argentino conectado, el envío por Correo se vuelve a
// cotizar acá con la API.
func ComprarConCorreo(ctx context.Context, pool tx, mp Pagador, correo Correo, sitio, slug string, c Compra, hoy time.Time) (Resultado, error) {
	t, err := Publica(ctx, pool, slug)
	if err != nil {
		return Resultado{}, err
	}
	id, _ := tiendaDe(ctx, pool, slug)
	c.Nombre, c.Email = strings.TrimSpace(c.Nombre), strings.TrimSpace(c.Email)
	if len(c.Carrito) == 0 || len(c.Carrito) > 50 || len(c.Nombre) < 2 || len(c.Nombre) > 80 || !emailValido.MatchString(c.Email) ||
		(c.Pago != "transferencia" && c.Pago != "mercadopago") {
		return Resultado{}, ErrInvalido
	}
	if c.Pago == "mercadopago" && (mp == nil || t.Cobros["mercadoPago"] != true) {
		return Resultado{}, ErrInvalido
	}
	if c.Pago == "transferencia" && t.Cobros["datosTransferencia"] == nil {
		return Resultado{}, ErrInvalido
	}
	factura, err := facturaDe(c, t.Cobros["facturaA"] == true)
	if err != nil {
		return Resultado{}, err
	}
	porID := map[string]ProductoPublico{}
	for _, p := range t.Productos {
		porID[p.ID] = p
	}
	// Los precios salen de acá: con el carrito al mínimo mayorista, los que tienen precio mayorista
	// pasan a ese (§49) y el pedido queda marcado.
	precios, esMayorista, ok := preciosCarrito(t, porID, c.Carrito)
	if !ok {
		return Resultado{}, ErrInvalido
	}
	var subtotal int64
	var unidades int
	var n pedidos.PedidoNuevo
	for i, l := range c.Carrito {
		if porID[l.ID].Lanzamiento != nil {
			return Resultado{}, ErrProximamente
		}
		subtotal += precios[i] * int64(l.Cantidad)
		unidades += l.Cantidad
		n.Items = append(n.Items, pedidos.Linea{ProductoID: l.ID, VarianteID: l.VarianteID, Cantidad: l.Cantidad})
		// El precio de cada línea lo fija la vidriera (con la oferta vigente, §52, y el mayorista,
		// §49), no el de lista.
		n.Precios = fijarPrecio(n.Precios, l.ID, l.VarianteID, precios[i])
	}
	n.Mayorista = esMayorista
	var costoEnvio int64
	var entrega *pedidos.Entrega
	n.Canal, n.Envio = "web", c.Envio
	if esCorreo(c.Envio) {
		if !t.Envios.Correo {
			return Resultado{}, ErrInvalido
		}
		gratis := t.Envios.GratisDesde != nil && subtotal >= *t.Envios.GratisDesde
		if costoEnvio, entrega, err = envioCorreo(ctx, correo, id, c, unidades, gratis); err != nil {
			return Resultado{}, err
		}
		n.Envio = "correo"
	} else {
		var opcion *Opcion
		for _, o := range Cotizar(c.CP, subtotal, t.Envios, hoy) {
			if o.ID == c.Envio {
				o := o
				opcion = &o
			}
		}
		if opcion == nil || (c.Envio != "retiro" && len(strings.TrimSpace(c.Direccion)) < 4) {
			return Resultado{}, ErrInvalido
		}
		costoEnvio = opcion.Precio
		if c.Envio != "retiro" {
			entrega = &pedidos.Entrega{Direccion: strings.TrimSpace(c.Direccion), CP: strings.TrimSpace(c.CP)}
		}
	}
	// Cupón: se vuelve a evaluar acá con los precios y el historial del servidor.
	var descCupon int64
	combinable := true
	if codigo := strings.ToUpper(strings.TrimSpace(c.Cupon)); codigo != "" {
		r, err := EvaluarCupon(ctx, pool, id, codigo, c.Email, lineasDe(c, porID, precios), hoy)
		if err != nil {
			return Resultado{}, err
		}
		if !r.OK {
			return Resultado{}, ErrCupon{Motivo: r.Motivo}
		}
		n.Cupon, descCupon, combinable = codigo, r.Descuento, r.Combinable
		if r.EnvioGratis {
			costoEnvio = 0
		}
	}
	n.Pago = "transf"
	if c.Pago == "mercadopago" {
		n.Pago = "mp"
	} else if pct, _ := t.Cobros["transferencia"].(int); combinable && pct > 0 {
		n.Descuento = (subtotal - descCupon) * int64(pct) / 100
	}
	n.Descuento += descCupon
	n.DescuentoCupon = descCupon
	f := false
	n.Pagado = &f
	n.Cliente, n.Telefono = &c.Nombre, &c.Telefono
	n.Email, n.CostoEnvio, n.CiudadEnvio = c.Email, costoEnvio, strings.TrimSpace(c.CP)
	n.Entrega, n.Factura = entrega, factura
	n.Nota = notaCompra(c, factura)
	b := make([]byte, 18)
	_, _ = rand.Read(b)
	n.CodigoPublico = base64.RawURLEncoding.EncodeToString(b)
	numero, total, err := pedidos.CrearDesdeProductos(ctx, pool, id, n)
	if errors.Is(err, pedidos.ErrSinStock) {
		return Resultado{}, ErrSinStock
	}
	if errors.Is(err, pedidos.ErrCupon) {
		return Resultado{}, ErrCupon{Motivo: "Este cupón se acaba de agotar."}
	}
	if errors.Is(err, pedidos.ErrVariante) || errors.Is(err, pedidos.ErrSinVariante) || errors.Is(err, pedidos.ErrProducto) {
		return Resultado{}, ErrInvalido
	}
	if err != nil {
		return Resultado{}, err
	}
	// Compró: su carrito abandonado (si había) queda recuperado.
	_ = pool.EnTienda(ctx, id, func(t pgx.Tx) error { return ventas.CarritoRecuperado(ctx, t, c.Email) })
	if c.Novedades {
		if err := newsletter.Suscribir(ctx, pool, id, c.Email, "checkout"); err != nil {
			slog.Warn("vidriera: no se sumó al newsletter desde el checkout", "error", err)
		}
	}
	if c.Medicion != nil && t.Medicion.Pixel != nil {
		medirCompra(ctx, pool, id, c, n, porID, numero, total, hoy)
	}
	r := Resultado{Numero: numero, Codigo: n.CodigoPublico, Total: total}
	if c.Pago == "mercadopago" {
		r.URL, err = mp.PagarPedido(ctx, id, numero, fmt.Sprintf("%s/t/%s/pedido/%s", sitio, slug, r.Codigo))
	}
	return r, err
}

// medirCompra encola la compra para la API de conversiones de Meta. Si falla, la compra sigue (sólo
// se pierde esa medición, que el píxel del navegador igual mandó).
func medirCompra(ctx context.Context, pool tx, tienda string, c Compra, n pedidos.PedidoNuevo, porID map[string]ProductoPublico,
	numero int, total int64, hoy time.Time) {
	// Las líneas con el precio minorista, como las manda el píxel (§49: el píxel y el feed no saben
	// del mayorista); el total sí es el cobrado.
	lineas := make([]conversiones.Linea, 0, len(n.Items))
	for _, l := range n.Items {
		precio, _ := precioDe(porID[l.ProductoID], l.VarianteID)
		lineas = append(lineas, conversiones.Linea{ID: l.ProductoID, Cantidad: l.Cantidad, Precio: precio})
	}
	nav := *c.Medicion
	recortar := func(s string, n int) string {
		if len(s) > n {
			return s[:n]
		}
		return s
	}
	nav.IP, nav.UA, nav.FBP, nav.FBC, nav.URL = recortar(nav.IP, 64), recortar(nav.UA, 512), recortar(nav.FBP, 200), recortar(nav.FBC, 300), recortar(nav.URL, 300)
	compra := conversiones.Compra{Numero: numero, Total: total, Hora: hoy.Unix(), Lineas: lineas,
		Email: conversiones.HashEmail(c.Email), Telefono: conversiones.HashTelefono(c.Telefono), Navegador: nav}
	if err := conversiones.Encolar(ctx, pool, tienda, compra); err != nil {
		slog.Error("vidriera: no se encoló la compra para Meta", "error", err)
	}
}

// facturaDe: el CUIT del checkout, sólo si la tienda ofrece factura A. Vacío, sin factura; inválido, error.
func facturaDe(c Compra, ofrece bool) (*pedidos.Factura, error) {
	cuit := strings.NewReplacer("-", "", " ", "", ".", "").Replace(c.CUIT)
	if cuit == "" || !ofrece {
		return nil, nil
	}
	razon := strings.Join(strings.Fields(c.RazonSocial), " ")
	if razon == "" {
		razon = c.Nombre
	}
	if !arca.CUITValido(cuit) || len(razon) < 2 || len(razon) > 120 {
		return nil, ErrInvalido
	}
	return &pedidos.Factura{CUIT: cuit, RazonSocial: razon}, nil
}

// notaCompra: el DNI, la factura A y adónde se entrega, sólo lo que vino (nada de campos vacíos).
// Sin nada, sin nota.
func notaCompra(c Compra, f *pedidos.Factura) *string {
	partes := []string{}
	if d := strings.TrimSpace(c.DNI); d != "" {
		partes = append(partes, "DNI "+d)
	}
	if f != nil {
		partes = append(partes, fmt.Sprintf("Factura A: %s, CUIT %s", f.RazonSocial, f.CUIT))
	}
	if c.Envio != "retiro" {
		dir, cp := strings.TrimSpace(c.Direccion), strings.TrimSpace(c.CP)
		switch {
		case dir != "" && cp != "":
			partes = append(partes, fmt.Sprintf("Entrega: %s (CP %s)", dir, cp))
		case dir != "":
			partes = append(partes, "Entrega: "+dir)
		}
	}
	if len(partes) == 0 {
		return nil
	}
	nota := strings.Join(partes, " · ")
	return &nota
}

// precioDe la línea: el de la variante si tiene, si no el del producto. Con opciones exige una
// variante activa del producto; sin opciones no acepta ninguna.
func precioDe(p ProductoPublico, varianteID string) (int64, bool) {
	if len(p.Opciones) == 0 {
		return p.Precio, varianteID == ""
	}
	for _, v := range p.Variantes {
		if v.ID == varianteID {
			if v.Precio != nil {
				return *v.Precio, true
			}
			return p.Precio, true
		}
	}
	return 0, false
}

// Confirmacion es lo que muestra la página del pedido (y su seguimiento).
type Confirmacion struct {
	Numero      int     `json:"numero"`
	Total       int64   `json:"total"`
	Pago        string  `json:"pago"`
	Nombre      string  `json:"nombre"`
	EstadoPago  string  `json:"estadoPago"`
	Estado      string  `json:"estado"`
	Envio       string  `json:"envio"`
	Seguimiento *string `json:"seguimiento"`
	Alias       *string `json:"alias"`
	CBU         *string `json:"cbu"`
	Titular     *string `json:"titular"`
	// Sena: el pedido se reservó o encargó con seña (§52); null si se paga entero.
	Sena *SenaPublica `json:"sena"`
}

// SenaPublica es la seña de un pedido como la ve quien compró.
type SenaPublica struct {
	Tipo    string `json:"tipo"`
	Monto   int64  `json:"monto"`
	Falta   int64  `json:"falta"`
	Pagada  bool   `json:"pagada"`
	Saldado bool   `json:"saldado"`
}

// VerPedido por su código público.
func VerPedido(ctx context.Context, pool tx, slug, codigo string) (Confirmacion, error) {
	id, err := tiendaDe(ctx, pool, slug)
	if err != nil {
		return Confirmacion{}, err
	}
	var c Confirmacion
	err = pool.EnTienda(ctx, id, func(t pgx.Tx) error {
		var sena *int64
		var tipo *string
		var s SenaPublica
		if err := t.QueryRow(ctx, `SELECT p.numero, p.total, CASE WHEN p.pago = 'mp' THEN 'mercadopago' ELSE 'transferencia' END,
				split_part(p.comprador->>'nombre', ' ', 1), p.estado_pago, p.estado, p.envio, p.seguimiento,
				nullif(t.negocio->>'alias', ''), nullif(t.negocio->>'cbu', ''), nullif(t.negocio->>'titular', ''),
				p.sena, p.sena_tipo, p.sena_pagada_en IS NOT NULL, p.saldado_en IS NOT NULL
			FROM pedidos p JOIN tiendas t ON t.id = p.tienda_id WHERE p.codigo_publico = $1`, codigo).
			Scan(&c.Numero, &c.Total, &c.Pago, &c.Nombre, &c.EstadoPago, &c.Estado, &c.Envio, &c.Seguimiento, &c.Alias, &c.CBU, &c.Titular,
				&sena, &tipo, &s.Pagada, &s.Saldado); err != nil {
			return err
		}
		if sena != nil && tipo != nil {
			s.Tipo, s.Monto, s.Falta = *tipo, *sena, max(0, c.Total-*sena)
			if s.Saldado {
				s.Falta = 0
			}
			c.Sena = &s
		}
		return nil
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return c, ErrNoExiste
	}
	return c, err
}
