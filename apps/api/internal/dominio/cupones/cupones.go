// Package cupones: cupones y promociones automáticas (design/PanelCupones.dc.html). Los cupones se
// crean, pausan y duplican de verdad; lo vendido, los segmentos y la oferta relámpago llegan con el
// checkout de la tienda, así que hasta entonces van en cero y vacíos, nunca inventados.
package cupones

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/mayorista"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/auditoria"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

type tx interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
}

var (
	ErrNoExiste  = errors.New("el cupón no existe")
	ErrRepetido  = errors.New("ya hay un cupón con ese código")
	ErrTerminado = errors.New("el cupón ya terminó")
	ErrInvalido  = errors.New("cupón inválido")
	codigoValido = regexp.MustCompile(`^[A-Z0-9-]{3,20}$`)
	argentina    = time.FixedZone("ART", -3*3600)
)

// Cupon es components/schemas/Cupon.
type Cupon struct {
	ID          string     `json:"id"`
	Codigo      string     `json:"codigo"`
	Tipo        string     `json:"tipo"`
	Valor       int        `json:"valor"`
	Descripcion string     `json:"descripcion"`
	Estado      string     `json:"estado"`
	Usos        int        `json:"usos"`
	Limite      int        `json:"limite"`
	Vendio      int64      `json:"vendio"`
	Desconto    int64      `json:"desconto"`
	Canales     []string   `json:"canales"`
	Empieza     *time.Time `json:"empieza"`
	Termina     *time.Time `json:"termina"`
	Minimo      int64      `json:"minimo"`
	Consejo     *string    `json:"consejo"`
	Creado      time.Time  `json:"creado"`
}

// Nuevo es components/schemas/CuponNuevo.
type Nuevo struct {
	Codigo     string   `json:"codigo"`
	Tipo       string   `json:"tipo"`
	Valor      int      `json:"valor"`
	Alcance    string   `json:"alcance"`
	Productos  []string `json:"productos"`
	Quien      string   `json:"quien"`
	Minimo     int64    `json:"minimo"`
	Limite     int      `json:"limite"`
	PorCliente string   `json:"porCliente"`
	Vigencia   string   `json:"vigencia"`
	Canales    []string `json:"canales"`
	Opciones   struct {
		Hornero    bool `json:"hornero"`
		Combinable bool `json:"combinable"`
		Unicos     bool `json:"unicos"`
		Link       bool `json:"link"`
	} `json:"opciones"`
	Programar bool `json:"programar"`
}

const columnas = `id::text, codigo, tipo, valor, descripcion, pausado, usos, limite, vendio, desconto, canales,
	empieza, termina, minimo, opciones->>'hornero' = 'true', creado_en`

func escanear(f pgx.Row, ahora time.Time) (Cupon, error) {
	var c Cupon
	var pausado, hornero bool
	err := f.Scan(&c.ID, &c.Codigo, &c.Tipo, &c.Valor, &c.Descripcion, &pausado, &c.Usos, &c.Limite, &c.Vendio,
		&c.Desconto, &c.Canales, &c.Empieza, &c.Termina, &c.Minimo, &hornero, &c.Creado)
	switch {
	case c.Termina != nil && c.Termina.Before(ahora), c.Limite > 0 && c.Usos >= c.Limite:
		c.Estado = "terminado"
	case pausado:
		c.Estado = "pausado"
	case c.Empieza != nil && c.Empieza.After(ahora):
		c.Estado = "programado"
	default:
		c.Estado = "activo"
	}
	if hornero && c.Estado != "terminado" {
		s := "Tincho ya lo puede ofrecer en el chat."
		c.Consejo = &s
	}
	return c, err
}

func pesos(n int64) string {
	s := fmt.Sprint(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "." + s[i:]
	}
	return "$ " + s
}

// vigencia: cuándo arranca y vence (lib/panel/cupones.ts, vigenciaDe).
func vigencia(v string, ahora time.Time) (*time.Time, *time.Time) {
	ar := ahora.In(argentina)
	p := func(t time.Time) *time.Time { return &t }
	switch v {
	case "7":
		return nil, p(ahora.Add(7 * 24 * time.Hour))
	case "30":
		return nil, p(ahora.Add(30 * 24 * time.Hour))
	case "finde":
		dias := (7 - int(ar.Weekday())) % 7
		return nil, p(time.Date(ar.Year(), ar.Month(), ar.Day()+dias, 23, 59, 0, 0, argentina))
	case "hotsale":
		d := time.Date(ar.Year(), time.May, 11, 0, 0, 0, 0, argentina)
		if !d.After(ahora) {
			d = d.AddDate(1, 0, 0)
		}
		return &d, p(d.Add(3*24*time.Hour - time.Minute))
	}
	return nil, nil
}

func lunes9(ahora time.Time) time.Time {
	ar := ahora.In(argentina)
	dias := (8 - int(ar.Weekday())) % 7
	if dias == 0 {
		dias = 7
	}
	return time.Date(ar.Year(), ar.Month(), ar.Day()+dias, 9, 0, 0, 0, argentina)
}

func encabezado(n Nuevo) string {
	switch n.Tipo {
	case "pct":
		return fmt.Sprintf("%d %% off", n.Valor)
	case "fijo":
		return pesos(int64(n.Valor)) + " off"
	case "envio":
		return "Envío gratis"
	}
	return "Llevá 2, pagá 1"
}

// Crear guarda el cupón. El nombre de los productos y la categoría salen de la tienda.
func Crear(ctx context.Context, pool tx, tiendaID string, n Nuevo, ahora time.Time) (Cupon, error) {
	var c Cupon
	canalesOk := len(n.Canales) > 0
	for _, k := range n.Canales {
		canalesOk = canalesOk && slices.Contains([]string{"web", "wa", "ig", "pos"}, k)
	}
	if !codigoValido.MatchString(n.Codigo) || !canalesOk || n.Minimo < 0 || n.Limite < 0 || len(n.Productos) > 50 ||
		!slices.Contains([]string{"pct", "fijo", "envio", "regalo"}, n.Tipo) ||
		!slices.Contains([]string{"todo", "categoria", "productos"}, n.Alcance) ||
		!slices.Contains([]string{"todos", "nuevos", "vip", "riesgo"}, n.Quien) ||
		!slices.Contains([]string{"1", "3", "libre"}, n.PorCliente) ||
		!slices.Contains([]string{"7", "finde", "30", "hotsale", "sin"}, n.Vigencia) ||
		(n.Tipo == "pct" && (n.Valor < 1 || n.Valor > 90)) || (n.Tipo == "fijo" && (n.Valor < 100 || n.Valor > 100000)) {
		return c, ErrInvalido
	}
	if n.Tipo == "envio" || n.Tipo == "regalo" {
		n.Valor = 0
	}
	empieza, termina := vigencia(n.Vigencia, ahora)
	if n.Programar {
		l := lunes9(ahora)
		empieza = &l
		if termina != nil && termina.Before(l.Add(24*time.Hour)) {
			t := l.Add(24 * time.Hour)
			termina = &t
		}
	}
	opciones, _ := json.Marshal(n.Opciones)
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		// Uno por código entre los vigentes: se bloquea la tienda para que dos altas no choquen.
		if _, err := t.Exec(ctx, `SELECT 1 FROM tiendas WHERE id = $1 FOR UPDATE`, tiendaID); err != nil {
			return err
		}
		var repetido bool
		if err := t.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM cupones WHERE tienda_id = $1 AND codigo = $2
			AND (termina IS NULL OR termina > now()) AND (limite = 0 OR usos < limite))`, tiendaID, n.Codigo).Scan(&repetido); err != nil {
			return err
		}
		if repetido {
			return ErrRepetido
		}
		donde := "en toda la tienda"
		// La categoría queda guardada en el alcance: la tienda online la necesita para aplicarlo.
		categoria := ""
		switch n.Alcance {
		case "categoria":
			var cat *string
			_ = t.QueryRow(ctx, `SELECT categoria FROM productos WHERE tienda_id = $1 AND categoria IS NOT NULL
				GROUP BY categoria ORDER BY count(*) DESC LIMIT 1`, tiendaID).Scan(&cat)
			donde = "en una categoría"
			if cat != nil {
				donde = "en " + strings.ToLower(*cat)
				categoria = *cat
			}
		case "productos":
			var nombres []string
			_ = t.QueryRow(ctx, `SELECT coalesce(array_agg(nombre ORDER BY nombre), '{}') FROM productos
				WHERE tienda_id = $1 AND id::text = ANY($2)`, tiendaID, n.Productos).Scan(&nombres)
			donde = "en los productos que elijas"
			if len(nombres) > 0 {
				donde = "en " + strings.Join(nombres, ", ")
			}
		}
		alcance, _ := json.Marshal(map[string]any{"tipo": n.Alcance, "productos": n.Productos, "categoria": categoria})
		desc := encabezado(n) + " " + donde
		if n.Minimo > 0 {
			desc += " · mín. " + pesos(n.Minimo)
		}
		var err error
		c, err = escanear(t.QueryRow(ctx, `INSERT INTO cupones (tienda_id, codigo, tipo, valor, descripcion, limite, canales,
				empieza, termina, minimo, quien, por_cliente, alcance, opciones)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14) RETURNING `+columnas,
			tiendaID, n.Codigo, n.Tipo, n.Valor, desc, n.Limite, n.Canales, empieza, termina, n.Minimo, n.Quien, n.PorCliente,
			alcance, opciones), ahora)
		if err != nil {
			return err
		}
		return auditoria.Registrar(ctx, t, tiendaID, auditoria.Registro{Accion: "promocion.cupon_crear", Entidad: "cupon",
			EntidadID: c.ID, Texto: fmt.Sprintf("creó el cupón %s (%s)", c.Codigo, desc)})
	})
	return c, err
}

// Accion: pausar, activar, duplicar, minimo o deshacer-minimo.
func Accion(ctx context.Context, pool tx, tiendaID, id, accion string, minimo int64, ahora time.Time) (Cupon, error) {
	var c Cupon
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var err error
		c, err = escanear(t.QueryRow(ctx, `SELECT `+columnas+` FROM cupones WHERE tienda_id = $1 AND id::text = $2 FOR UPDATE`, tiendaID, id), ahora)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoExiste
		}
		if err != nil {
			return err
		}
		sinMinimo := regexp.MustCompile(` · mín\. .*$`).ReplaceAllString(c.Descripcion, "")
		original, minimoAntes := c.Codigo, c.Minimo
		var q string
		var args []any
		switch accion {
		case "duplicar":
			codigo := c.Codigo + "-2"
			if len(codigo) > 20 {
				codigo = codigo[:20]
			}
			c, err = escanear(t.QueryRow(ctx, `INSERT INTO cupones (tienda_id, codigo, tipo, valor, descripcion, pausado, limite,
					canales, minimo, quien, por_cliente, alcance, opciones)
				SELECT tienda_id, $3, tipo, valor, descripcion, true, limite, canales, minimo, quien, por_cliente, alcance, opciones
				FROM cupones WHERE tienda_id = $1 AND id::text = $2 RETURNING `+columnas, tiendaID, id, codigo), ahora)
			if err != nil {
				return err
			}
			return auditoria.Registrar(ctx, t, tiendaID, auditoria.Registro{Accion: "promocion.cupon_duplicar", Entidad: "cupon",
				EntidadID: c.ID, Texto: fmt.Sprintf("duplicó el cupón %s como %s (quedó pausado)", original, c.Codigo)})
		case "pausar", "activar":
			q, args = `UPDATE cupones SET pausado = $3`, []any{accion == "pausar"}
		case "minimo":
			if minimo < 1 {
				return ErrInvalido
			}
			q, args = `UPDATE cupones SET minimo = $3, descripcion = $4`, []any{minimo, sinMinimo + " · mín. " + pesos(minimo)}
		case "deshacer-minimo":
			q, args = `UPDATE cupones SET minimo = 0, descripcion = $3`, []any{sinMinimo}
		default:
			return ErrInvalido
		}
		if c.Estado == "terminado" {
			return ErrTerminado
		}
		c, err = escanear(t.QueryRow(ctx, q+` WHERE tienda_id = $1 AND id::text = $2 RETURNING `+columnas,
			append([]any{tiendaID, id}, args...)...), ahora)
		if err != nil {
			return err
		}
		texto := map[string]string{
			"pausar":          "pausó el cupón " + c.Codigo,
			"activar":         "activó el cupón " + c.Codigo,
			"minimo":          fmt.Sprintf("le puso un mínimo de compra de %s al cupón %s", pesos(minimo), c.Codigo),
			"deshacer-minimo": "le sacó el mínimo de compra al cupón " + c.Codigo,
		}[accion]
		return auditoria.Registrar(ctx, t, tiendaID, auditoria.Registro{Accion: "promocion.cupon_" + strings.ReplaceAll(accion, "-", "_"),
			Entidad: "cupon", EntidadID: c.ID, Texto: texto,
			Antes: map[string]int64{"minimo": minimoAntes}, Despues: map[string]int64{"minimo": c.Minimo}})
	})
	return c, err
}

type automatica struct {
	ID, Titulo, Detalle, Ejemplo string
}

// Lo que aplica la tienda online cuando la tienda prende estas promociones (vidriera las lee de acá):
// la pantalla las describe con los mismos números, así no dice una cosa y cobra otra.
const (
	TransferenciaPct = 15
	EnvioGratisDesde = 40000
)

var automaticas = []automatica{
	{"transf", "Descuento por transferencia", fmt.Sprintf("%d %% si pagan con transferencia", TransferenciaPct),
		fmt.Sprintf("Carrito de $ 50.000 → paga %s", pesosTxt(50000*(100-TransferenciaPct)/100))},
	{"envio", "Envío gratis desde un monto", fmt.Sprintf("Desde %s a todo el país", pesosTxt(EnvioGratisDesde)),
		"Le faltan $ 6.000 → “Sumá algo y el envío es gratis”"},
	{"cantidad", "Descuento por cantidad", "3 o más unidades: 15 % off", "3 unidades → 15 % off automático"},
	{"3x2", "3x2 en una categoría", "Llevás 3, pagás 2", "3 unidades → la más barata gratis"},
	{"cuotas", "Cuotas sin interés", "3 y 6 cuotas desde $ 30.000", "6 cuotas sin interés desde $ 30.000"},
	// El detalle del mayorista sale del mínimo de Configuración › Venta mayorista (§49).
	{"mayorista", "Precio mayorista", "", "Carrito al mínimo → precio mayorista automático"},
	{"carrito", "Rescate de carrito con Tincho", "Ofrece hasta 10 % si el carrito se enfría", "Carrito abandonado 2 h → “Te guardo 10 % hasta las 20 h”"},
	{"cumple", "Cumpleaños", "15 % el día del cumple del cliente", "Le llega por WhatsApp a las 9 h"},
}

func pesosTxt(n int64) string {
	s := fmt.Sprint(n)
	var b strings.Builder
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(r)
	}
	return "$ " + b.String()
}

// Leer devuelve todo lo de la pantalla.
func Leer(ctx context.Context, pool tx, tiendaID string, ahora time.Time) (map[string]any, error) {
	res := map[string]any{}
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		filas, err := t.Query(ctx, `SELECT `+columnas+` FROM cupones WHERE tienda_id = $1 ORDER BY creado_en DESC LIMIT 500`, tiendaID)
		if err != nil {
			return err
		}
		lista := []Cupon{}
		for filas.Next() {
			c, err := escanear(filas, ahora)
			if err != nil {
				return err
			}
			lista = append(lista, c)
		}
		if err := filas.Err(); err != nil {
			return err
		}
		var activos, usos int
		var vendido, descontado int64
		for _, c := range lista {
			if c.Estado == "activo" {
				activos++
				usos += c.Usos
			}
			vendido += c.Vendio
			descontado += c.Desconto
		}
		var prendidas map[string]bool
		var nombre string
		if err := t.QueryRow(ctx, `SELECT automaticas, nombre FROM tiendas WHERE id = $1`, tiendaID).Scan(&prendidas, &nombre); err != nil {
			return err
		}
		may, err := mayorista.Leer(ctx, t, tiendaID)
		if err != nil {
			return err
		}
		autos := []map[string]any{}
		for _, a := range automaticas {
			if a.ID == "mayorista" {
				a.Detalle = may.Detalle() + ", con el precio mayorista de cada producto"
			}
			estado := "Apagada"
			if prendidas[a.ID] {
				estado = "Activa"
			}
			autos = append(autos, map[string]any{"id": a.ID, "titulo": a.Titulo, "detalle": a.Detalle, "ejemplo": a.Ejemplo,
				"estado": estado, "aporte": nil, "activa": prendidas[a.ID]})
		}
		type prod struct {
			ID     string  `json:"id"`
			Nombre string  `json:"nombre"`
			Foto   *string `json:"foto"`
		}
		pf, err := t.Query(ctx, `SELECT id::text, nombre, foto FROM productos WHERE tienda_id = $1 ORDER BY creado_en DESC LIMIT 6`, tiendaID)
		if err != nil {
			return err
		}
		productos, err := pgx.CollectRows(pf, pgx.RowToStructByPos[prod])
		if err != nil {
			return err
		}
		var categoria *string
		_ = t.QueryRow(ctx, `SELECT categoria FROM productos WHERE tienda_id = $1 AND categoria IS NOT NULL
			GROUP BY categoria ORDER BY count(*) DESC LIMIT 1`, tiendaID).Scan(&categoria)
		res = map[string]any{
			"cupones":     lista,
			"kpis":        map[string]any{"activos": activos, "usosMes": usos, "vendido": vendido, "descontado": descontado, "vsMesAnterior": nil},
			"consejo":     nil,
			"segmentos":   map[string]int{"vip": 0, "riesgo": 0},
			"automaticas": autos,
			"relampago":   map[string]any{"vivo": nil, "proximas": []any{}},
			"tienda":      nombre,
			"productos":   productos,
			"categoria":   categoria,
		}
		return nil
	})
	return res, err
}

// FijarAutomatica prende o apaga una promoción automática.
func FijarAutomatica(ctx context.Context, pool tx, tiendaID, id string, activa bool) error {
	i := slices.IndexFunc(automaticas, func(a automatica) bool { return a.ID == id })
	if i < 0 {
		return ErrNoExiste
	}
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var antes *bool
		if err := t.QueryRow(ctx, `SELECT (automaticas->>$2)::boolean FROM tiendas WHERE id = $1 FOR UPDATE`, tiendaID, id).Scan(&antes); err != nil {
			return err
		}
		if _, err := t.Exec(ctx, `UPDATE tiendas SET automaticas = automaticas || jsonb_build_object($2::text, $3::boolean) WHERE id = $1`,
			tiendaID, id, activa); err != nil {
			return err
		}
		if antes != nil && *antes == activa {
			return nil
		}
		verbo := "apagó"
		if activa {
			verbo = "prendió"
		}
		return auditoria.Registrar(ctx, t, tiendaID, auditoria.Registro{Accion: "promocion.automatica", Entidad: "promocion",
			EntidadID: id, Texto: fmt.Sprintf("%s la promoción «%s»", verbo, automaticas[i].Titulo),
			Despues: map[string]bool{"activa": activa}})
	})
}

// Rutas del panel.
func Rutas(mux *http.ServeMux, pool tx, tienda func(*http.Request) string) {
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
	escribir := func(w http.ResponseWriter, estado int, v any) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(estado)
		_ = json.NewEncoder(w).Encode(v)
	}
	errorDe := func(w http.ResponseWriter, err error) {
		switch {
		case errors.Is(err, ErrNoExiste):
			problema.Escribir(w, problema.NoEncontrado("Ese cupón no existe."))
		case errors.Is(err, ErrRepetido):
			problema.Escribir(w, problema.Conflicto("repetido", "Ya tenés un cupón con ese código."))
		case errors.Is(err, ErrTerminado):
			problema.Escribir(w, problema.Conflicto("terminado", "Ese cupón ya terminó. Duplicalo para usarlo de nuevo."))
		case errors.Is(err, ErrInvalido):
			problema.Escribir(w, problema.Invalido("Revisá los datos del cupón."))
		default:
			problema.Escribir(w, problema.Interno())
		}
	}
	mux.HandleFunc("GET /v1/panel/cupones", con(func(w http.ResponseWriter, r *http.Request, id string) {
		res, err := Leer(r.Context(), pool, id, time.Now())
		if err != nil {
			errorDe(w, err)
			return
		}
		escribir(w, http.StatusOK, res)
	}))
	mux.HandleFunc("POST /v1/panel/cupones", con(func(w http.ResponseWriter, r *http.Request, id string) {
		var n Nuevo
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&n); err != nil {
			errorDe(w, ErrInvalido)
			return
		}
		c, err := Crear(r.Context(), pool, id, n, time.Now())
		if err != nil {
			errorDe(w, err)
			return
		}
		escribir(w, http.StatusCreated, c)
	}))
	mux.HandleFunc("POST /v1/panel/cupones/{id}", con(func(w http.ResponseWriter, r *http.Request, id string) {
		var b struct {
			Accion string `json:"accion"`
			Minimo int64  `json:"minimo"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&b); err != nil {
			errorDe(w, ErrInvalido)
			return
		}
		c, err := Accion(r.Context(), pool, id, r.PathValue("id"), b.Accion, b.Minimo, time.Now())
		if err != nil {
			errorDe(w, err)
			return
		}
		escribir(w, http.StatusOK, c)
	}))
	// Sin consejos de Hornero todavía: no hay nada que ocultar.
	mux.HandleFunc("DELETE /v1/panel/cupones/consejo", con(func(w http.ResponseWriter, _ *http.Request, _ string) {
		w.WriteHeader(http.StatusNoContent)
	}))
	mux.HandleFunc("PUT /v1/panel/automaticas/{id}", con(func(w http.ResponseWriter, r *http.Request, id string) {
		var b struct {
			Activa bool `json:"activa"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<10)).Decode(&b); err != nil {
			errorDe(w, ErrInvalido)
			return
		}
		if err := FijarAutomatica(r.Context(), pool, id, r.PathValue("id"), b.Activa); err != nil {
			errorDe(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	mux.HandleFunc("POST /v1/panel/relampago", con(func(w http.ResponseWriter, _ *http.Request, _ string) {
		problema.Escribir(w, problema.NoEncontrado("No hay una oferta relámpago."))
	}))
}
