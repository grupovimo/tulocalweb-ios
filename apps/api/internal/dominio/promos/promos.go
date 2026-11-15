// Package promos: lo que las ventanas de la tienda online necesitan del servidor
// (docs/especificaciones/tienda-promociones.md): captar un contacto con su premio, sortear la ruleta,
// juntar las métricas por ventana y día, las compras recientes para la prueba social y los números
// del panel. Las ventanas viven en el diseño publicado (mitienda.Ventana); acá sólo se las lee.
package promos

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"math/big"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cupones"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/mitienda"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/newsletter"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

type tx interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
	EnTx(context.Context, func(pgx.Tx) error) error
}

var (
	ErrNoExiste = errors.New("promos: no existe")
	// ErrInvalido: el email, el nombre o el WhatsApp no sirven, o falta el consentimiento.
	ErrInvalido = errors.New("promos: datos inválidos")
	// ErrTerminada: la ventana no está activa, no está completa, venció o su cupón ya terminó.
	ErrTerminada = errors.New("promos: la promo terminó")

	argentina = func() *time.Location {
		l, err := time.LoadLocation("America/Argentina/Buenos_Aires")
		if err != nil {
			return time.FixedZone("ART", -3*3600)
		}
		return l
	}()
	reSlug     = regexp.MustCompile(`^[a-z0-9-]{2,60}$`)
	reVentana  = regexp.MustCompile(`^[a-z0-9]{6,16}$`)
	reWhatsapp = regexp.MustCompile(`^\+?[0-9 ()-]{6,24}$`)
)

// Pedido es lo que manda la ventana: components/schemas/PromoCaptar.
type Pedido struct {
	Ventana        string `json:"ventana"`
	Email          string `json:"email"`
	Nombre         string `json:"nombre"`
	Whatsapp       string `json:"whatsapp"`
	Consentimiento bool   `json:"consentimiento"`
}

// Premio es lo que se le da: components/schemas/PromoPremio.
type Premio struct {
	// Tipo: cuponFijo, cuponUnico, descarga o nada.
	Tipo string `json:"tipo"`
	// Codigo del cupón (fijo o único); Descripcion: «10 % off», «Envío gratis»…
	Codigo      string `json:"codigo,omitempty"`
	Descripcion string `json:"descripcion,omitempty"`
	// Descarga: el link del archivo y su nombre.
	Descarga      string     `json:"descarga,omitempty"`
	NombreArchivo string     `json:"nombreArchivo,omitempty"`
	Vence         *time.Time `json:"vence,omitempty"`
}

// Resultado de captar o girar: components/schemas/PromoResultado.
type Resultado struct {
	Premio Premio `json:"premio"`
	// Porcion: en la ruleta, la que salió (0 es la primera de la lista) y su texto.
	Porcion *int   `json:"porcion,omitempty"`
	Texto   string `json:"texto,omitempty"`
	// Repetido: ya había participado con ese email (se devuelve lo mismo que la primera vez).
	Repetido bool `json:"repetido"`
	// Demo: la tienda es una vitrina de demostración; el premio se muestra pero no se guarda ni sirve.
	Demo bool `json:"demo,omitempty"`
}

// tienda: lo publicado que necesita una ventana.
type tienda struct {
	id, nombre, slug, plantilla string
	diseno                      mitienda.Diseno
	vitrina                     bool
}

func cargar(ctx context.Context, pool tx, slug string) (tienda, error) {
	var t tienda
	if !reSlug.MatchString(slug) {
		return t, ErrNoExiste
	}
	var id *string
	if err := pool.EnTx(ctx, func(x pgx.Tx) error {
		return x.QueryRow(ctx, `SELECT app.tienda_de_slug($1)::text`, slug).Scan(&id)
	}); err != nil {
		return t, err
	}
	if id == nil {
		return t, ErrNoExiste
	}
	t.id = *id
	err := pool.EnTienda(ctx, t.id, func(x pgx.Tx) error {
		var crudo []byte
		if err := x.QueryRow(ctx, `SELECT nombre, slug, plantilla, diseno_publicado, demo FROM tiendas WHERE id = $1`, t.id).
			Scan(&t.nombre, &t.slug, &t.plantilla, &crudo, &t.vitrina); err != nil {
			return err
		}
		t.diseno = mitienda.LeerDiseno(crudo, t.plantilla)
		return nil
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return t, ErrNoExiste
	}
	return t, err
}

func (t tienda) ventana(id string) (mitienda.Ventana, bool) {
	for _, v := range t.diseno.Ventanas {
		if v.ID == id {
			return v, true
		}
	}
	return mitienda.Ventana{}, false
}

// vigente: activa, completa y dentro de sus fechas.
func vigente(v mitienda.Ventana, ahora time.Time) bool {
	if !v.Activa || !v.Completa() {
		return false
	}
	if d, err := time.Parse(time.RFC3339, v.Desde); err == nil && ahora.Before(d) {
		return false
	}
	if h, err := time.Parse(time.RFC3339, v.Hasta); err == nil && !ahora.Before(h) {
		return false
	}
	return true
}

func hashEmail(e string) []byte {
	h := sha256.Sum256([]byte(strings.ToLower(e)))
	return h[:]
}

// limpiar valida lo que manda la ventana: email siempre; nombre y WhatsApp opcionales.
func limpiar(p Pedido) (email string, nombre, whatsapp *string, err error) {
	email, ok := newsletter.Email(p.Email)
	if !ok || !p.Consentimiento || !reVentana.MatchString(p.Ventana) {
		return "", nil, nil, ErrInvalido
	}
	if n := strings.TrimSpace(p.Nombre); n != "" {
		if utf8.RuneCountInString(n) > 60 || strings.ContainsAny(n, "<>\r\n") {
			return "", nil, nil, ErrInvalido
		}
		nombre = &n
	}
	if w := strings.TrimSpace(p.Whatsapp); w != "" {
		if !reWhatsapp.MatchString(w) {
			return "", nil, nil, ErrInvalido
		}
		whatsapp = &w
	}
	return email, nombre, whatsapp, nil
}

// Sortear elige una porción según sus pesos (con crypto/rand: nadie la predice desde el navegador).
// -1 si ningún peso es mayor a 0.
func Sortear(pesos []int) int {
	total := 0
	for _, p := range pesos {
		if p > 0 {
			total += p
		}
	}
	if total == 0 {
		return -1
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(total)))
	if err != nil {
		return -1
	}
	x := int(n.Int64())
	for i, p := range pesos {
		if p <= 0 {
			continue
		}
		if x < p {
			return i
		}
		x -= p
	}
	return len(pesos) - 1
}

func numero(v any) int {
	f, _ := v.(float64)
	return int(f)
}

func texto(v any) string {
	s, _ := v.(string)
	return s
}

// premioDe arma el premio de la configuración (el de la captación o el de una porción): crea el cupón
// único, busca el fijo o devuelve la descarga. En la vitrina no se guarda nada.
func premioDe(ctx context.Context, x pgx.Tx, t tienda, v mitienda.Ventana, conf map[string]any, ahora time.Time) (Premio, error) {
	p := Premio{Tipo: texto(conf["tipo"])}
	switch p.Tipo {
	case "nada":
	case "descarga":
		p.Descarga, p.NombreArchivo = texto(conf["archivo"]), texto(conf["nombreArchivo"])
	case "cuponFijo":
		p.Codigo = texto(conf["cupon"])
		if t.vitrina {
			return p, nil
		}
		var termina *time.Time
		err := x.QueryRow(ctx, `SELECT descripcion, termina FROM cupones WHERE tienda_id = $1 AND codigo = $2 AND NOT pausado
				AND (termina IS NULL OR termina > now()) AND (limite = 0 OR usos < limite) ORDER BY creado_en DESC LIMIT 1`,
			t.id, p.Codigo).Scan(&p.Descripcion, &termina)
		if errors.Is(err, pgx.ErrNoRows) {
			return p, ErrTerminada
		}
		if err != nil {
			return p, err
		}
		p.Vence = termina
	case "cuponUnico":
		desc, valor, dias := texto(conf["descuento"]), numero(conf["valor"]), numero(conf["dias"])
		if dias < 1 {
			dias = 7
		}
		_, d, ok := cupones.TipoYDescripcion(desc, valor)
		if !ok {
			return p, ErrTerminada
		}
		p.Descripcion = d
		if t.vitrina {
			vence := ahora.Add(time.Duration(dias) * 24 * time.Hour)
			p.Codigo, p.Vence = "DEMO-"+strings.ToUpper(v.ID[:4]), &vence
			return p, nil
		}
		codigo, vence, err := cupones.CrearUnico(ctx, x, t.id, cupones.Unico{Prefijo: prefijo(v), Descuento: desc, Valor: valor,
			Dias: dias, Origen: "ventana:" + v.ID, Nombre: v.Nombre}, ahora)
		if err != nil {
			return p, err
		}
		p.Codigo, p.Vence = codigo, &vence
	default:
		return p, ErrTerminada
	}
	return p, nil
}

// prefijo del código: las primeras letras del nombre de la ventana («Bienvenida» → BIENVE).
func prefijo(v mitienda.Ventana) string {
	var b strings.Builder
	for _, r := range strings.ToUpper(sinTildes(v.Nombre)) {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func sinTildes(s string) string {
	return strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ñ", "n", "Á", "A", "É", "E", "Í", "I",
		"Ó", "O", "Ú", "U", "Ñ", "N", "ü", "u", "Ü", "U").Replace(s)
}

// Participar es captar (contenido captacion) o girar (ruleta): una vez por email y ventana. La segunda
// vez devuelve lo mismo que la primera. Suma el contacto a la newsletter con origen «ventana:<id>» y
// encola el mail con el premio.
func Participar(ctx context.Context, pool tx, slug, contenido string, p Pedido, ahora time.Time) (Resultado, error) {
	email, nombre, whatsapp, err := limpiar(p)
	if err != nil {
		return Resultado{}, err
	}
	t, err := cargar(ctx, pool, slug)
	if err != nil {
		return Resultado{}, err
	}
	v, ok := t.ventana(p.Ventana)
	if !ok || v.Contenido != contenido {
		return Resultado{}, ErrNoExiste
	}
	if !vigente(v, ahora) {
		return Resultado{}, ErrTerminada
	}
	datos := v.Datos
	var res Resultado
	elegir := func() (map[string]any, error) {
		if contenido != "ruleta" {
			conf, _ := datos["premio"].(map[string]any)
			return conf, nil
		}
		porciones := v.Porciones()
		pesos := make([]int, len(porciones))
		for i, x := range porciones {
			pesos[i] = numero(x["peso"])
		}
		i := Sortear(pesos)
		if i < 0 {
			return nil, ErrTerminada
		}
		res.Porcion, res.Texto = &i, texto(porciones[i]["texto"])
		conf, _ := porciones[i]["premio"].(map[string]any)
		return conf, nil
	}
	if t.vitrina {
		conf, err := elegir()
		if err != nil {
			return Resultado{}, err
		}
		res.Demo = true
		res.Premio, err = premioDe(ctx, nil, t, v, conf, ahora)
		return res, err
	}
	hash := hashEmail(email)
	err = pool.EnTienda(ctx, t.id, func(x pgx.Tx) error {
		// Uno a la vez por persona y ventana: dos toques seguidos no dan dos premios.
		if _, err := x.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, t.id+v.ID+email); err != nil {
			return err
		}
		var crudo []byte
		var porcion *int
		err := x.QueryRow(ctx, `SELECT premio, porcion FROM promo_captaciones WHERE tienda_id = $1 AND ventana = $2 AND email_hash = $3`,
			t.id, v.ID, hash).Scan(&crudo, &porcion)
		if err == nil {
			res.Repetido, res.Porcion = true, porcion
			if porcion != nil {
				if ps := v.Porciones(); *porcion < len(ps) {
					res.Texto = texto(ps[*porcion]["texto"])
				}
			}
			return json.Unmarshal(crudo, &res.Premio)
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		conf, err := elegir()
		if err != nil {
			return err
		}
		if res.Premio, err = premioDe(ctx, x, t, v, conf, ahora); err != nil {
			return err
		}
		guardado, _ := json.Marshal(res.Premio)
		if _, err := x.Exec(ctx, `INSERT INTO promo_captaciones (tienda_id, ventana, email_hash, porcion, premio, creado_en)
			VALUES ($1, $2, $3, $4, $5, $6)`, t.id, v.ID, hash, res.Porcion, guardado, ahora); err != nil {
			return err
		}
		s, err := newsletter.SuscribirDesdeVentana(ctx, x, t.id, email, "ventana:"+v.ID, nombre, whatsapp)
		if err != nil {
			return err
		}
		return outbox.Publicar(ctx, x, t.id, "promos.premio", map[string]string{"suscriptor": s.ID, "ventana": v.ID})
	})
	return res, err
}

// Evento de la tienda: vista, cierre o clic de una ventana (n veces, hasta 20 por evento).
type Evento struct {
	Ventana string `json:"ventana"`
	Tipo    string `json:"tipo"`
	N       int    `json:"n"`
}

// MaxEventos por lote.
const MaxEventos = 50

// Registrar suma un lote de eventos al día de hoy. Lo que no es de una ventana publicada se ignora;
// en la vitrina de demostración no se guarda nada.
func Registrar(ctx context.Context, pool tx, slug string, eventos []Evento, ahora time.Time) error {
	if len(eventos) > MaxEventos {
		return ErrInvalido
	}
	t, err := cargar(ctx, pool, slug)
	if err != nil || t.vitrina {
		return err
	}
	type suma struct{ vistas, cierres, clics int }
	sumas := map[string]*suma{}
	for _, e := range eventos {
		if _, ok := t.ventana(e.Ventana); !ok {
			continue
		}
		n := e.N
		if n < 1 {
			n = 1
		}
		if n > 20 {
			n = 20
		}
		s := sumas[e.Ventana]
		if s == nil {
			s = &suma{}
			sumas[e.Ventana] = s
		}
		switch e.Tipo {
		case "vista":
			s.vistas += n
		case "cierre":
			s.cierres += n
		case "clic":
			s.clics += n
		}
	}
	if len(sumas) == 0 {
		return nil
	}
	dia := ahora.In(argentina).Format(time.DateOnly)
	return pool.EnTienda(ctx, t.id, func(x pgx.Tx) error {
		for id, s := range sumas {
			if _, err := x.Exec(ctx, `INSERT INTO promo_metricas (tienda_id, ventana, dia, vistas, cierres, clics) VALUES ($1, $2, $3, $4, $5, $6)
				ON CONFLICT (tienda_id, ventana, dia) DO UPDATE SET vistas = promo_metricas.vistas + EXCLUDED.vistas,
					cierres = promo_metricas.cierres + EXCLUDED.cierres, clics = promo_metricas.clics + EXCLUDED.clics`,
				t.id, id, dia, s.vistas, s.cierres, s.clics); err != nil {
				return err
			}
		}
		return nil
	})
}

// Compra reciente para la prueba social: components/schemas/PromoReciente.
type Compra struct {
	// Nombre de pila, nada más.
	Nombre string `json:"nombre"`
	// Provincia: "" si no se sabe (no se inventa).
	Provincia  string    `json:"provincia"`
	Producto   string    `json:"producto"`
	ProductoID *string   `json:"productoId"`
	Foto       *string   `json:"foto"`
	Cuando     time.Time `json:"cuando"`
}

type enCache struct {
	hasta   time.Time
	compras []Compra
}

var (
	cacheMu sync.Mutex
	cache   = map[string]enCache{}
)

// CacheRecientes: cuánto se guardan las compras recientes de cada tienda.
const CacheRecientes = 5 * time.Minute

// nombrePila: la primera palabra, si parece un nombre («Consumidor final» y los vacíos no salen).
func nombrePila(n string) string {
	n = strings.TrimSpace(n)
	if n == "" || strings.EqualFold(n, "consumidor final") || strings.Contains(n, "@") {
		return ""
	}
	p := strings.Fields(n)[0]
	r := []rune(p)
	if len(r) < 2 || len(r) > 20 {
		return ""
	}
	for _, c := range r {
		if !unicode.IsLetter(c) && c != '\'' && c != '-' {
			return ""
		}
	}
	return strings.ToUpper(string(r[:1])) + strings.ToLower(string(r[1:]))
}

// Recientes: los pedidos pagados de los últimos 7 días (hasta 20), con nombre de pila, provincia,
// producto y cuándo. Sólo si la tienda prendió la prueba social en su diseño publicado: si no, vacío.
func Recientes(ctx context.Context, pool tx, slug string, ahora time.Time) ([]Compra, error) {
	cacheMu.Lock()
	if c, ok := cache[slug]; ok && ahora.Before(c.hasta) {
		cacheMu.Unlock()
		return c.compras, nil
	}
	cacheMu.Unlock()
	t, err := cargar(ctx, pool, slug)
	if err != nil {
		return nil, err
	}
	compras := []Compra{}
	if PruebaSocialActiva(t.diseno) {
		err = pool.EnTienda(ctx, t.id, func(x pgx.Tx) error {
			filas, err := x.Query(ctx, `SELECT p.comprador->>'nombre', coalesce(c.provincia, ''), coalesce(p.entrega->>'provincia', ''),
					coalesce(nullif(p.entrega->>'cp', ''), p.ciudad_envio), coalesce(p.items->0->>'nombre', ''), p.items->0->>'productoId',
					p.items->0->>'foto', p.creado_en
				FROM pedidos p LEFT JOIN clientes c ON c.tienda_id = p.tienda_id AND c.id = p.cliente_id
				WHERE p.tienda_id = $1 AND p.estado <> 'Cancelado' AND p.estado_pago IN ('Aprobado', 'Acreditada', 'Cobrado')
					AND p.creado_en > $2 AND p.creado_en <= $3
				ORDER BY p.creado_en DESC LIMIT 40`, t.id, ahora.Add(-7*24*time.Hour), ahora)
			if err != nil {
				return err
			}
			defer filas.Close()
			for filas.Next() {
				var nombre, ficha, correo, cp string
				var c Compra
				if err := filas.Scan(&nombre, &ficha, &correo, &cp, &c.Producto, &c.ProductoID, &c.Foto, &c.Cuando); err != nil {
					return err
				}
				if c.Nombre = nombrePila(nombre); c.Nombre == "" || c.Producto == "" || len(compras) >= 20 {
					continue
				}
				c.Provincia = provinciaDe(ficha, correo, cp)
				compras = append(compras, c)
			}
			return filas.Err()
		})
		if err != nil {
			return nil, err
		}
	}
	cacheMu.Lock()
	cache[slug] = enCache{hasta: ahora.Add(CacheRecientes), compras: compras}
	cacheMu.Unlock()
	return compras, nil
}

// PruebaSocialActiva: estilo.conversion.pruebaSocial.activa (por defecto, apagada).
func PruebaSocialActiva(d mitienda.Diseno) bool {
	if d.Estilo == nil {
		return false
	}
	ps, _ := d.Estilo.Conversion["pruebaSocial"].(map[string]any)
	a, _ := ps["activa"].(bool)
	return a
}
