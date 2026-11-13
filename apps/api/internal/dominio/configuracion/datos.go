package configuracion

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/auditoria"
)

var dominioValido = regexp.MustCompile(`^([a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,}$`)

// Resolver permite cambiar la búsqueda de DNS en los tests.
var Resolver interface {
	LookupCNAME(context.Context, string) (string, error)
	LookupTXT(context.Context, string) ([]string, error)
} = net.DefaultResolver

type Chequeo struct {
	Titulo string `json:"titulo"`
	Estado string `json:"estado"`
}

type Dominio struct {
	Nombre    string    `json:"nombre"`
	Detalle   string    `json:"detalle"`
	Estado    string    `json:"estado"`
	Principal bool      `json:"principal"`
	Chequeos  []Chequeo `json:"chequeos,omitempty"`
}

func host(sitio string) string {
	return strings.TrimPrefix(strings.TrimPrefix(sitio, "https://"), "http://")
}

// verificar mira el DNS de verdad: www con CNAME a las tiendas y el TXT con el slug.
func verificar(ctx context.Context, nombre, sitio, slug string) []Chequeo {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	cname, _ := Resolver.LookupCNAME(ctx, "www."+nombre)
	okCname := strings.TrimSuffix(cname, ".") == "tiendas."+host(sitio)
	txts, _ := Resolver.LookupTXT(ctx, "_tulocalweb."+nombre)
	okTxt := false
	for _, t := range txts {
		okTxt = okTxt || t == "tulocalweb-verificacion="+slug
	}
	est := func(ok bool) string {
		if ok {
			return "ok"
		}
		return "esperando"
	}
	ssl := "pendiente"
	if okCname && okTxt {
		ssl = "esperando"
	}
	return []Chequeo{{"Registro CNAME de www", est(okCname)}, {"Verificación TXT", est(okTxt)}, {"Certificado SSL", ssl}}
}

// Dominios lista la dirección gratis y los propios; los que no están verificados se vuelven a mirar.
func Dominios(ctx context.Context, pool tx, tiendaID, sitio string) (map[string]any, error) {
	var res map[string]any
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var slug string
		if err := t.QueryRow(ctx, `SELECT slug FROM tiendas WHERE id = $1`, tiendaID).Scan(&slug); err != nil {
			return err
		}
		gratis := host(sitio) + "/t/" + slug
		filas, err := t.Query(ctx, `SELECT nombre, principal, verificado_en IS NOT NULL FROM dominios WHERE tienda_id = $1 ORDER BY creado_en`, tiendaID)
		if err != nil {
			return err
		}
		type fila struct {
			nombre                string
			principal, verificado bool
		}
		var propios []fila
		for filas.Next() {
			var f fila
			if err := filas.Scan(&f.nombre, &f.principal, &f.verificado); err != nil {
				return err
			}
			propios = append(propios, f)
		}
		if err := filas.Err(); err != nil {
			return err
		}
		hayPrincipal := false
		lista := []Dominio{}
		for _, f := range propios {
			d := Dominio{Nombre: f.nombre, Principal: f.principal, Estado: "activo", Detalle: "Conectado"}
			if !f.verificado {
				d.Chequeos = verificar(ctx, f.nombre, sitio, slug)
				if d.Chequeos[0].Estado == "ok" && d.Chequeos[1].Estado == "ok" {
					// El certificado lo emite Cloudflare for SaaS al registrar el dominio (llega con esa integración).
					if _, err := t.Exec(ctx, `UPDATE dominios SET verificado_en = now() WHERE nombre = $1`, f.nombre); err != nil {
						return err
					}
				} else {
					d.Estado, d.Detalle = "verificando", "Agregado · revisando DNS"
				}
			}
			hayPrincipal = hayPrincipal || f.principal
			lista = append(lista, d)
		}
		res = map[string]any{
			"gratis":   gratis,
			"dominios": append([]Dominio{{Nombre: gratis, Detalle: "La dirección gratis de tu tienda", Estado: "activo", Principal: !hayPrincipal}}, lista...),
			"registros": []map[string]string{
				{"tipo": "CNAME", "nombre": "www", "valor": "tiendas." + host(sitio)},
				{"tipo": "TXT", "nombre": "_tulocalweb", "valor": "tulocalweb-verificacion=" + slug},
			},
		}
		return nil
	})
	return res, err
}

// ConectarDominio lo agrega; si es de otra tienda, la clave primaria lo impide.
func ConectarDominio(ctx context.Context, pool tx, tiendaID, nombre string) error {
	nombre = strings.ToLower(strings.TrimSpace(nombre))
	if len(nombre) > 253 || !dominioValido.MatchString(nombre) {
		return ErrInvalido
	}
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		if _, err := t.Exec(ctx, `INSERT INTO dominios (nombre, tienda_id) VALUES ($1, $2)`, nombre, tiendaID); err != nil {
			return err
		}
		return auditoria.Registrar(ctx, t, tiendaID, auditoria.Registro{Accion: "dominio.conectar", Entidad: "dominio",
			EntidadID: nombre, Texto: "conectó el dominio " + nombre})
	})
	var pe *pgconn.PgError
	if errors.As(err, &pe) && pe.Code == "23505" {
		return ErrRepetido
	}
	return err
}

// DominioPrincipal: el gratis no está en la tabla, así que elegirlo es dejar ninguno como principal.
func DominioPrincipal(ctx context.Context, pool tx, tiendaID, nombre, sitio string) error {
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var slug string
		if err := t.QueryRow(ctx, `SELECT slug FROM tiendas WHERE id = $1`, tiendaID).Scan(&slug); err != nil {
			return err
		}
		if nombre == host(sitio)+"/t/"+slug {
			if _, err := t.Exec(ctx, `UPDATE dominios SET principal = false WHERE tienda_id = $1`, tiendaID); err != nil {
				return err
			}
			return auditoria.Registrar(ctx, t, tiendaID, auditoria.Registro{Accion: "dominio.principal", Entidad: "dominio",
				EntidadID: nombre, Texto: "dejó como principal la dirección gratis " + nombre})
		}
		var verificado *time.Time
		err := t.QueryRow(ctx, `SELECT verificado_en FROM dominios WHERE tienda_id = $1 AND nombre = $2`, tiendaID, nombre).Scan(&verificado)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoExiste
		}
		if err != nil {
			return err
		}
		if verificado == nil {
			return ErrNoSePuede
		}
		if _, err = t.Exec(ctx, `UPDATE dominios SET principal = (nombre = $2) WHERE tienda_id = $1`, tiendaID, nombre); err != nil {
			return err
		}
		return auditoria.Registrar(ctx, t, tiendaID, auditoria.Registro{Accion: "dominio.principal", Entidad: "dominio",
			EntidadID: nombre, Texto: "eligió " + nombre + " como dominio principal"})
	})
}

func QuitarDominio(ctx context.Context, pool tx, tiendaID, nombre string) error {
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		tag, err := t.Exec(ctx, `DELETE FROM dominios WHERE tienda_id = $1 AND nombre = $2`, tiendaID, nombre)
		if err == nil && tag.RowsAffected() == 0 {
			return ErrNoExiste
		}
		if err != nil {
			return err
		}
		return auditoria.Registrar(ctx, t, tiendaID, auditoria.Registro{Accion: "dominio.quitar", Entidad: "dominio",
			EntidadID: nombre, Texto: "quitó el dominio " + nombre})
	})
}

// Exportar devuelve las filas (el BFF arma el CSV, el Excel o el ZIP).
func Exportar(ctx context.Context, pool tx, tiendaID, tipo string) ([][]any, error) {
	consultas := map[string]struct {
		cab []any
		sql string
	}{
		"productos": {[]any{"SKU", "Nombre", "Categoría", "Precio", "Costo", "Stock", "Publicado"},
			`SELECT coalesce(sku, ''), nombre, coalesce(categoria, ''), precio, costo, stock, CASE WHEN publicado THEN 'Sí' ELSE 'No' END
			 FROM productos WHERE tienda_id = $1 ORDER BY nombre`},
		"clientes": {[]any{"Nombre", "Email", "Teléfono", "Pedidos", "Gastado"},
			`SELECT comprador->>'nombre', coalesce(comprador->>'email', ''), coalesce(comprador->>'telefono', ''), count(*), sum(total)
			 FROM pedidos WHERE tienda_id = $1 AND estado <> 'Cancelado'
				AND NOT (comprador->>'nombre' = 'Consumidor final' AND coalesce(comprador->>'email', '') = ''
				AND coalesce(comprador->>'telefono', '') = '')
			 GROUP BY 1, 2, 3 ORDER BY 1`},
		"pedidos": {[]any{"Número", "Fecha", "Cliente", "Canal", "Pago", "Envío", "Estado", "Total"},
			`SELECT numero, to_char(creado_en, 'YYYY-MM-DD'), comprador->>'nombre', canal, pago, envio, estado, total
			 FROM pedidos WHERE tienda_id = $1 ORDER BY numero DESC`},
		// Los comprobantes de Facturación (pedidos.factura es texto, no un objeto). Sin ninguno,
		// sale sólo el encabezado.
		"facturas": {[]any{"Fecha", "Comprobante", "Cliente", "Pedido", "Total", "Estado", "CAE"},
			`SELECT to_char(c.creado_en AT TIME ZONE 'America/Argentina/Buenos_Aires', 'YYYY-MM-DD'),
				CASE WHEN c.tipo IN (3, 8, 13) THEN 'NC ' ELSE '' END ||
				CASE WHEN c.tipo IN (1, 3) THEN 'A' WHEN c.tipo IN (6, 8) THEN 'B' ELSE 'C' END ||
				CASE WHEN c.numero IS NULL THEN '' ELSE ' ' || lpad(coalesce(c.punto_venta, 0)::text, 4, '0') || '-' || lpad(c.numero::text, 8, '0') END,
				coalesce(c.receptor->>'nombre', 'Consumidor final'), c.pedido_numero, c.total,
				CASE WHEN EXISTS (SELECT 1 FROM comprobantes n WHERE n.asociado = c.id AND n.estado = 'emitido') THEN 'void' ELSE c.estado END,
				coalesce(c.cae, '')
			 FROM comprobantes c WHERE c.tienda_id = $1 ORDER BY c.creado_en DESC`},
		// Las reseñas de la tienda online (antes salía sólo el encabezado aunque las hubiera).
		"resenas": {[]any{"Fecha", "Cliente", "Producto", "Estrellas", "Reseña", "Respuesta", "Oculta"},
			`SELECT to_char(r.creada AT TIME ZONE 'America/Argentina/Buenos_Aires', 'YYYY-MM-DD'), r.autor,
				coalesce(p.nombre, ''), r.estrellas, r.texto, coalesce(r.respuesta, ''), CASE WHEN r.oculta THEN 'Sí' ELSE 'No' END
			 FROM resenas r LEFT JOIN productos p ON p.tienda_id = r.tienda_id AND p.id = r.producto_id
			 WHERE r.tienda_id = $1 ORDER BY r.creada DESC`},
	}
	q, ok := consultas[tipo]
	if !ok {
		// Las conversaciones llegan con WhatsApp: mientras tanto sale el encabezado solo.
		switch tipo {
		case "conversaciones":
			return [][]any{{"Conversación", "Cliente", "De", "Mensaje", "Fecha"}}, nil
		}
		return nil, ErrNoExiste
	}
	out := [][]any{q.cab}
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		filas, err := t.Query(ctx, q.sql, tiendaID)
		if err != nil {
			return err
		}
		for filas.Next() {
			v, err := filas.Values()
			if err != nil {
				return err
			}
			out = append(out, v)
		}
		return filas.Err()
	})
	return out, err
}

// MisDatos: cuánto hay de cada cosa.
func MisDatos(ctx context.Context, pool tx, tiendaID string) (map[string]any, error) {
	c := map[string]int{}
	var recientes []Exportacion
	var inactivos int
	var programado *time.Time
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var p, cl, pe, f, r int
		if err := t.QueryRow(ctx, `SELECT (SELECT count(*) FROM productos WHERE tienda_id = $1),
				(SELECT count(DISTINCT comprador->>'nombre') FROM pedidos WHERE tienda_id = $1
					AND NOT (comprador->>'nombre' = 'Consumidor final' AND coalesce(comprador->>'email', '') = ''
					AND coalesce(comprador->>'telefono', '') = '')),
				(SELECT count(*) FROM pedidos WHERE tienda_id = $1),
				(SELECT count(*) FROM comprobantes WHERE tienda_id = $1),
				(SELECT count(*) FROM resenas WHERE tienda_id = $1)`, tiendaID).Scan(&p, &cl, &pe, &f, &r); err != nil {
			return err
		}
		c = map[string]int{"productos": p, "clientes": cl, "pedidos": pe, "facturas": f, "resenas": r, "conversaciones": 0}
		var err error
		if inactivos, err = listas.ContarInactivos(ctx, t, tiendaID); err != nil {
			return err
		}
		return t.QueryRow(ctx, `SELECT coalesce(preferencias->'exportaciones', '[]'), borrado_inactivos FROM tiendas WHERE id = $1`,
			tiendaID).Scan(&recientes, &programado)
	})
	if recientes == nil {
		recientes = []Exportacion{}
	}
	var borrado any
	if programado != nil && inactivos > 0 {
		borrado = map[string]any{"clientes": inactivos, "el": programado.UTC()}
	}
	return map[string]any{"conteos": c, "recientes": recientes, "inactivos": inactivos, "borrado": borrado}, err
}

// Exportacion es una descarga de Mis datos ("Exportaciones recientes").
type Exportacion struct {
	Tipo    string    `json:"tipo"`
	Formato string    `json:"formato"`
	En      time.Time `json:"en"`
}

var tiposExport = map[string]bool{"productos": true, "clientes": true, "pedidos": true, "facturas": true,
	"resenas": true, "conversaciones": true, "todo": true}

// RegistrarExportacion guarda la descarga en las preferencias de la tienda (las últimas 5).
func RegistrarExportacion(ctx context.Context, pool tx, tiendaID, tipo, formato string, ahora time.Time) error {
	if !tiposExport[tipo] || (formato != "csv" && formato != "excel") {
		return ErrInvalido
	}
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var antes []Exportacion
		if err := t.QueryRow(ctx, `SELECT coalesce(preferencias->'exportaciones', '[]') FROM tiendas WHERE id = $1 FOR UPDATE`, tiendaID).Scan(&antes); err != nil {
			return err
		}
		lista := append([]Exportacion{{Tipo: tipo, Formato: formato, En: ahora.UTC()}}, antes...)
		if len(lista) > 5 {
			lista = lista[:5]
		}
		if _, err := t.Exec(ctx, `UPDATE tiendas SET preferencias = jsonb_set(preferencias, '{exportaciones}', $2) WHERE id = $1`, tiendaID, lista); err != nil {
			return err
		}
		que := map[string]string{"productos": "los productos", "clientes": "los clientes", "pedidos": "los pedidos",
			"facturas": "las facturas", "resenas": "las reseñas", "conversaciones": "las conversaciones", "todo": "todos los datos de la tienda"}[tipo]
		en := map[string]string{"csv": "CSV", "excel": "Excel"}[formato]
		if tipo == "todo" {
			en = "un ZIP"
		}
		return auditoria.Registrar(ctx, t, tiendaID, auditoria.Registro{Accion: "datos.exportar", Entidad: "datos", EntidadID: tipo,
			Texto: "bajó " + que + " en " + en})
	})

}

// CerrarCuenta: con el nombre de la tienda como confirmación. Queda offline y se borra a los 90 días.
func CerrarCuenta(ctx context.Context, pool tx, tiendaID, confirmacion string) error {
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var nombre string
		if err := t.QueryRow(ctx, `SELECT nombre FROM tiendas WHERE id = $1`, tiendaID).Scan(&nombre); err != nil {
			return err
		}
		if !strings.EqualFold(strings.TrimSpace(confirmacion), nombre) {
			return ErrInvalido
		}
		_, err := t.Exec(ctx, `UPDATE tiendas SET cerrada_en = now(), cancelada_en = coalesce(cancelada_en, now()) WHERE id = $1`, tiendaID)
		return err
	})
}

func rutasDatos(mux *http.ServeMux, pool tx, sitio string, con func(func(http.ResponseWriter, *http.Request, string, Quien)) http.HandlerFunc,
	errorDe func(http.ResponseWriter, error), leer func(http.ResponseWriter, *http.Request, any) bool) {
	escribir := func(w http.ResponseWriter, v any, err error) {
		if err != nil {
			errorDe(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(v)
	}
	mux.HandleFunc("GET /v1/panel/dominios", con(func(w http.ResponseWriter, r *http.Request, id string, _ Quien) {
		v, err := Dominios(r.Context(), pool, id, sitio)
		escribir(w, v, err)
	}))
	mux.HandleFunc("POST /v1/panel/dominios", con(func(w http.ResponseWriter, r *http.Request, id string, _ Quien) {
		var b struct{ Dominio string }
		if !leer(w, r, &b) {
			return
		}
		if err := ConectarDominio(r.Context(), pool, id, b.Dominio); err != nil {
			errorDe(w, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}))
	mux.HandleFunc("PUT /v1/panel/dominios/{dominio}", con(func(w http.ResponseWriter, r *http.Request, id string, _ Quien) {
		errorDe(w, DominioPrincipal(r.Context(), pool, id, r.PathValue("dominio"), sitio))
	}))
	mux.HandleFunc("DELETE /v1/panel/dominios/{dominio}", con(func(w http.ResponseWriter, r *http.Request, id string, _ Quien) {
		errorDe(w, QuitarDominio(r.Context(), pool, id, r.PathValue("dominio")))
	}))
	mux.HandleFunc("GET /v1/panel/notificaciones/preferencias", con(func(w http.ResponseWriter, r *http.Request, id string, _ Quien) {
		var p json.RawMessage
		err := pool.EnTienda(r.Context(), id, func(t pgx.Tx) error {
			return t.QueryRow(r.Context(), `SELECT coalesce(preferencias->'canales', '{"matriz": {}, "silencio": true, "resumen": "whatsapp"}') FROM tiendas WHERE id = $1`, id).Scan(&p)
		})
		escribir(w, p, err)
	}))
	mux.HandleFunc("PUT /v1/panel/notificaciones/preferencias", con(func(w http.ResponseWriter, r *http.Request, id string, _ Quien) {
		var b struct {
			Matriz   map[string]bool `json:"matriz"`
			Silencio bool            `json:"silencio"`
			Resumen  string          `json:"resumen"`
		}
		if !leer(w, r, &b) {
			return
		}
		if len(b.Matriz) > 60 || (b.Resumen != "whatsapp" && b.Resumen != "email" && b.Resumen != "off") {
			errorDe(w, ErrInvalido)
			return
		}
		errorDe(w, pool.EnTienda(r.Context(), id, func(t pgx.Tx) error {
			_, err := t.Exec(r.Context(), `UPDATE tiendas SET preferencias = jsonb_set(preferencias, '{canales}', $2) WHERE id = $1`, id, b)
			return err
		}))
	}))
	mux.HandleFunc("GET /v1/panel/datos", con(func(w http.ResponseWriter, r *http.Request, id string, _ Quien) {
		v, err := MisDatos(r.Context(), pool, id)
		escribir(w, v, err)
	}))
	mux.HandleFunc("GET /v1/panel/datos/{tipo}", con(func(w http.ResponseWriter, r *http.Request, id string, _ Quien) {
		v, err := Exportar(r.Context(), pool, id, r.PathValue("tipo"))
		escribir(w, map[string]any{"filas": v}, err)
	}))
	mux.HandleFunc("POST /v1/panel/datos/exportaciones", con(func(w http.ResponseWriter, r *http.Request, id string, _ Quien) {
		var b struct{ Tipo, Formato string }
		if !leer(w, r, &b) {
			return
		}
		if err := RegistrarExportacion(r.Context(), pool, id, b.Tipo, b.Formato, time.Now()); err != nil {
			errorDe(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	// Borrar los clientes inactivos: queda programado a 7 días (lo hace el worker); POST lo frena.
	mux.HandleFunc("DELETE /v1/panel/datos/inactivos", con(func(w http.ResponseWriter, r *http.Request, id string, _ Quien) {
		n, err := ProgramarBorradoInactivos(r.Context(), pool, id)
		escribir(w, map[string]int{"clientes": n}, err)
	}))
	mux.HandleFunc("POST /v1/panel/datos/inactivos", con(func(w http.ResponseWriter, r *http.Request, id string, _ Quien) {
		if err := FrenarBorradoInactivos(r.Context(), pool, id); err != nil {
			errorDe(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	mux.HandleFunc("DELETE /v1/panel/cuenta", con(func(w http.ResponseWriter, r *http.Request, id string, _ Quien) {
		var b struct{ Confirmacion string }
		if leer(w, r, &b) {
			errorDe(w, CerrarCuenta(r.Context(), pool, id, b.Confirmacion))
		}
	}))
}
