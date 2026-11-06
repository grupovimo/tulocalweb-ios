// Package demo arma las tiendas de demostración de Tu Local Web: tres negocios inventados pero
// creíbles (ropa, café y almacén), con catálogo, clientes, cuatro meses de pedidos, reseñas,
// preguntas, visitas y cupones, para mostrar el producto en vivo. Las usa `mostradorctl demo-tiendas`.
//
// La tienda queda con demo = true: su tienda online es una vitrina (se ve todo y no se compra).
//
// Sale SQL para correr como superusuario con los disparadores apagados (session_replication_role =
// replica): con ellos prendidos, cada pedido avisaría a la dueña, a Telegram y al comprador, y
// encolaría una factura en ARCA. Nada de eso tiene que pasar con ventas que no existieron.
//
// Todo es inventado: los nombres de las tiendas son los de las plantillas, los clientes no existen y
// sus mails son de example.com (reservado para ejemplos), así nunca se le escribe a nadie real.
package demo

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/mitienda"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
)

//go:embed catalogos.json
var catalogosJSON []byte

// ProductoEjemplo y Catalogo son los de web/src/lib/catalogos.ts (sacados del diseño).
type ProductoEjemplo struct {
	ID        string  `json:"id"`
	Nombre    string  `json:"nombre"`
	Categoria *string `json:"categoria"`
	Precio    int64   `json:"precio"`
	Anterior  *int64  `json:"anterior"`
	Foto      string  `json:"foto"`
	Etiqueta  *string `json:"etiqueta"`
	Detalle   *string `json:"detalle"`
}

type Catalogo struct {
	Plantilla  string            `json:"plantilla"`
	Tienda     string            `json:"tienda"`
	Rubro      string            `json:"rubro"`
	Ciudad     string            `json:"ciudad"`
	Categorias []string          `json:"categorias"`
	Productos  []ProductoEjemplo `json:"productos"`
}

// Tienda es cada demo: su dirección, su rubro y cómo vende.
type Tienda struct {
	Slug, Nombre, Rubro, Plantilla string
	Lema, Portada                  string
	Ciudad, Provincia, Direccion   string
	Horario, Whatsapp              string
	// Pedidos por día (mínimo y máximo) y cuánto más vende el fin de semana.
	PorDia      [2]int
	FinDeSemana float64
	Canales     map[string]int
	Pagos       map[string]int
	Envios      map[string]int
	MetaMes     int64
	Clientes    int
}

// Tiendas son las tres demos.
var Tiendas = []Tienda{
	{
		Slug: "lumbre", Nombre: "Tienda Lumbre", Rubro: "moda", Plantilla: "ceibo",
		Lema: "Ropa de autor, hecha para durar", Portada: "https://images.unsplash.com/photo-1573612664822-d7d347da7b80?auto=format&fit=crop&w=1600&q=72",
		Ciudad: "CABA", Provincia: "CABA", Direccion: "Gurruchaga 1840, Palermo", Horario: "Lunes a sábado de 11 a 20",
		Whatsapp: "+54 9 11 5555-0101", PorDia: [2]int{1, 4}, FinDeSemana: 1.3,
		Canales: map[string]int{"web": 45, "ig": 25, "wa": 20, "pos": 10},
		Pagos:   map[string]int{"mp": 45, "transf": 30, "tarjeta": 15, "modo": 10},
		Envios:  map[string]int{"andreani": 35, "correo": 20, "moto": 25, "retiro": 20},
		MetaMes: 14_000_000, Clientes: 70,
	},
	{
		Slug: "cafebruma", Nombre: "Café Bruma", Rubro: "gastronomia", Plantilla: "mostaza",
		Lema: "Café de especialidad tostado en Palermo", Portada: "https://images.unsplash.com/photo-1495474472287-4d71bcdd2085?auto=format&fit=crop&w=1600&q=72",
		Ciudad: "CABA", Provincia: "CABA", Direccion: "Thames 1622, Palermo", Horario: "Todos los días de 8 a 20",
		Whatsapp: "+54 9 11 5555-0102", PorDia: [2]int{4, 11}, FinDeSemana: 1.6,
		Canales: map[string]int{"pos": 65, "wa": 20, "web": 15},
		Pagos:   map[string]int{"efectivo": 30, "mp": 35, "modo": 15, "tarjeta": 20},
		Envios:  map[string]int{"retiro": 75, "moto": 25},
		MetaMes: 4_200_000, Clientes: 90,
	},
	{
		Slug: "almacenverde", Nombre: "Almacén Verde", Rubro: "almacen", Plantilla: "yerbal",
		Lema: "Natural, a granel y de productores locales", Portada: "https://images.unsplash.com/photo-1542838132-92c53300491e?auto=format&fit=crop&w=1600&q=72",
		Ciudad: "Córdoba", Provincia: "Córdoba", Direccion: "Bv. San Juan 640, Nueva Córdoba", Horario: "Lunes a sábado de 9 a 19",
		Whatsapp: "+54 9 351 555-0103", PorDia: [2]int{2, 6}, FinDeSemana: 1.2,
		Canales: map[string]int{"web": 40, "wa": 40, "pos": 20},
		Pagos:   map[string]int{"transf": 40, "mp": 35, "efectivo": 15, "uala": 10},
		Envios:  map[string]int{"moto": 45, "correo": 20, "retiro": 35},
		MetaMes: 3_000_000, Clientes: 80,
	},
}

// Lo que se arma por producto: su fila, sus variantes y cómo se lo nombra en un pedido.
type producto struct {
	id, nombre, foto string
	precio           int64
	variantes        []variante
}

type variante struct {
	id, texto string
	precio    int64
}

type cliente struct {
	id, nombre, telefono, ciudad string
	email                        *string
}

// SQL devuelve el script que crea las tres tiendas para el dueño con ese email. Se puede volver a
// correr: primero borra las demos que hubiera (borrar una tienda borra todo lo suyo en cascada).
func SQL(dueno string, ahora time.Time) (string, error) {
	var cats []Catalogo
	if err := json.Unmarshal(catalogosJSON, &cats); err != nil {
		return "", err
	}
	porPlantilla := map[string]Catalogo{}
	for _, c := range cats {
		porPlantilla[c.Plantilla] = c
	}
	var b strings.Builder
	// Primero se borran las demos que hubiera, con los disparadores prendidos: el borrado en cascada
	// de las claves foráneas también es un disparador, y con replica dejaría filas huérfanas.
	b.WriteString("BEGIN;\n")
	b.WriteString(Borrar())
	b.WriteString("SET LOCAL session_replication_role = replica;\n")
	for i, t := range Tiendas {
		c, ok := porPlantilla[t.Plantilla]
		if !ok {
			return "", fmt.Errorf("falta el catálogo de %s", t.Plantilla)
		}
		r := rand.New(rand.NewPCG(uint64(i+1)*7919, 2026))
		if err := tienda(&b, t, c, dueno, ahora, r); err != nil {
			return "", err
		}
	}
	b.WriteString("COMMIT;\n")
	return b.String(), nil
}

// Borrar devuelve el SQL que saca las demos (y sus direcciones).
func Borrar() string {
	var s []string
	for _, t := range Tiendas {
		s = append(s, q(t.Slug))
	}
	l := strings.Join(s, ", ")
	return fmt.Sprintf("DELETE FROM tiendas WHERE slug IN (%s);\nDELETE FROM app.slugs WHERE slug IN (%s);\n", l, l)
}

func q(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }

func j(v any) string {
	b, _ := json.Marshal(v)
	return q(string(b)) + "::jsonb"
}

func ts(t time.Time) string { return q(t.UTC().Format(time.RFC3339)) + "::timestamptz" }

func uuid(r *rand.Rand) string {
	var x [16]byte
	for i := range x {
		x[i] = byte(r.IntN(256))
	}
	x[6] = x[6]&0x0f | 0x40
	x[8] = x[8]&0x3f | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", x[0:4], x[4:6], x[6:8], x[8:10], x[10:16])
}

func elegir(r *rand.Rand, pesos map[string]int) string {
	total := 0
	claves := make([]string, 0, len(pesos))
	for k, p := range pesos {
		claves = append(claves, k)
		total += p
	}
	// Orden fijo, así el resultado no depende del orden del mapa.
	ordenar(claves)
	n := r.IntN(total)
	for _, k := range claves {
		if n < pesos[k] {
			return k
		}
		n -= pesos[k]
	}
	return claves[0]
}

func ordenar(s []string) {
	for i := 1; i < len(s); i++ {
		for k := i; k > 0 && s[k] < s[k-1]; k-- {
			s[k], s[k-1] = s[k-1], s[k]
		}
	}
}

// opcionesDe: las variantes de cada producto según el rubro (talles, colores, molienda, peso).
func opcionesDe(t Tienda, p ProductoEjemplo) ([]map[string]any, [][]map[string]string, []float64) {
	cat := ""
	if p.Categoria != nil {
		cat = *p.Categoria
	}
	switch {
	case t.Rubro == "moda" && strings.Contains(p.Nombre, "Stiletto"):
		return []map[string]any{{"nombre": "Talle", "valores": []string{"36", "37", "38", "39", "40"}}}, combinar("Talle", []string{"36", "37", "38", "39", "40"}), nil
	case t.Rubro == "moda" && strings.HasPrefix(p.Nombre, "Remera"):
		talles, colores := []string{"S", "M", "L", "XL"}, []string{"Negro", "Blanco", "Arena"}
		if strings.Contains(p.Nombre, "Negra") {
			colores = []string{"Negro"}
		}
		var vs [][]map[string]string
		for _, c := range colores {
			for _, tl := range talles {
				vs = append(vs, []map[string]string{{"Talle": tl, "Color": c}})
			}
		}
		return []map[string]any{{"nombre": "Talle", "valores": talles}, {"nombre": "Color", "valores": colores}}, vs, nil
	case t.Rubro == "moda" && cat != "Accesorios":
		talles := []string{"S", "M", "L", "XL"}
		return []map[string]any{{"nombre": "Talle", "valores": talles}}, combinar("Talle", talles), nil
	case t.Rubro == "gastronomia" && cat == "Café en grano":
		pesos, molienda := []string{"250 g", "1 kg"}, []string{"Grano", "Espresso", "Filtro"}
		var vs [][]map[string]string
		var mult []float64
		for _, pe := range pesos {
			for _, m := range molienda {
				vs = append(vs, []map[string]string{{"Peso": pe, "Molienda": m}})
				if pe == "1 kg" {
					mult = append(mult, 3.6)
				} else {
					mult = append(mult, 1)
				}
			}
		}
		return []map[string]any{{"nombre": "Peso", "valores": pesos}, {"nombre": "Molienda", "valores": molienda}}, vs, mult
	case t.Rubro == "almacen" && cat == "Frutos secos":
		pesos := []string{"250 g", "500 g", "1 kg"}
		return []map[string]any{{"nombre": "Peso", "valores": pesos}}, combinar("Peso", pesos), []float64{0.3, 0.55, 1}
	}
	return nil, nil, nil
}

func combinar(opcion string, valores []string) [][]map[string]string {
	var out [][]map[string]string
	for _, v := range valores {
		out = append(out, []map[string]string{{opcion: v}})
	}
	return out
}

func redondear(c int64) int64 { return (c + 50) / 100 * 100 } // a la centena de pesos

func tienda(b *strings.Builder, t Tienda, c Catalogo, dueno string, ahora time.Time, r *rand.Rand) error {
	tid := uuid(r)
	creada := ahora.AddDate(0, -4, -12)
	fmt.Fprintf(b, "\n-- %s\nINSERT INTO app.slugs (slug, creado_en) VALUES (%s, %s);\n", t.Nombre, q(t.Slug), ts(creada))

	d := mitienda.Inicial()
	d.Plantilla, d.Titulo, d.Lema, d.Portada = t.Plantilla, t.Nombre, t.Lema, t.Portada
	d.CompartirTitulo, d.CompartirDescripcion = t.Nombre, t.Lema
	negocio := map[string]string{
		"nombre": t.Nombre, "razonSocial": "", "cuit": "", "condicion": "monotributo", "direccion": t.Direccion + ", " + t.Ciudad,
		"whatsapp": t.Whatsapp, "email": "", "horario": t.Horario, "alias": "", "cbu": "", "titular": "", "politica": "",
	}
	fmt.Fprintf(b, `INSERT INTO tiendas (id, slug, nombre, rubro, plantilla, plan, creada_en, onboarding_terminado_en, meta_mensual,
	diseno_borrador, diseno_publicado, negocio, prueba_hasta, demo)
VALUES (%s, %s, %s, %s, %s, 'pro', %s, %s, %d, %s, %s, %s, %s, true);
`, q(tid), q(t.Slug), q(t.Nombre), q(t.Rubro), q(t.Plantilla), ts(creada), ts(creada.Add(2*time.Hour)), t.MetaMes,
		j(d), j(d), j(negocio), ts(ahora.AddDate(0, 2, 0)))
	fmt.Fprintf(b, "INSERT INTO auth.membresias (usuario_id, tienda_id, rol) SELECT id, %s, 'duenio' FROM auth.usuarios WHERE email = %s;\n",
		q(tid), q(dueno))

	// Catálogo.
	var prods []producto
	for i, p := range c.Productos {
		pid := uuid(r)
		precio := p.Precio
		var oferta, ofertaHasta string = "NULL", "NULL"
		if p.Anterior != nil {
			precio = *p.Anterior
			oferta, ofertaHasta = fmt.Sprint(p.Precio), ts(ahora.AddDate(0, 0, 9))
		}
		desc := ""
		if p.Detalle != nil {
			desc = *p.Detalle
		}
		desc = descripcion(t, p, desc)
		opciones, vars, mult := opcionesDe(t, p)
		if opciones == nil {
			opciones = []map[string]any{}
		}
		stock := 8 + r.IntN(40)
		if t.Rubro == "gastronomia" && (p.Categoria == nil || *p.Categoria != "Café en grano") {
			stock = 60 + r.IntN(80)
		}
		if p.Etiqueta != nil && *p.Etiqueta == "Últimos" {
			stock = 2
		}
		costo := precio * int64(42+r.IntN(14)) / 100
		cat := "NULL"
		if p.Categoria != nil {
			cat = q(*p.Categoria)
		}
		sku := fmt.Sprintf("%s-%03d", strings.ToUpper(t.Slug[:3]), i+1)
		fmt.Fprintf(b, `INSERT INTO productos (tienda_id, id, nombre, descripcion, categoria, precio, costo, sku, stock, foto, fotos, canales,
	publicado, creado_en, opciones, slug, oferta_precio, oferta_hasta)
VALUES (%s, %s, %s, %s, %s, %d, %d, %s, %d, %s, ARRAY[%s], '{"web": true, "pos": true}'::jsonb, true, %s, %s,
	app.slug_producto(%s::uuid, %s, %s::uuid), %s, %s);
`, q(tid), q(pid), q(p.Nombre), q(desc), cat, precio, costo, q(sku), stock, q(p.Foto), q(p.Foto),
			ts(creada.Add(time.Duration(3+i)*time.Hour)), j(opciones), q(tid), q(p.Nombre), q(pid), oferta, ofertaHasta)
		pr := producto{id: pid, nombre: p.Nombre, foto: p.Foto, precio: precio}
		if p.Anterior != nil {
			pr.precio = p.Precio
		}
		for k, v := range vars {
			vid := uuid(r)
			vp := "NULL"
			precioV := pr.precio
			if mult != nil {
				precioV = redondear(int64(float64(pr.precio) * mult[k]))
				vp = fmt.Sprint(precioV)
			}
			vstock := r.IntN(12)
			if r.IntN(10) == 0 {
				vstock = 0
			}
			var partes []string
			for _, op := range opciones {
				n := op["nombre"].(string)
				partes = append(partes, n+" "+v[0][n])
			}
			fmt.Fprintf(b, "INSERT INTO variantes (tienda_id, producto_id, id, valores, sku, precio, stock, creada) VALUES (%s, %s, %s, %s, %s, %s, %d, %s);\n",
				q(tid), q(pid), q(vid), j(v[0]), q(fmt.Sprintf("%s-%02d", sku, k+1)), vp, vstock, ts(creada.Add(time.Duration(3+i)*time.Hour)))
			pr.variantes = append(pr.variantes, variante{id: vid, texto: strings.Join(partes, " · "), precio: precioV})
		}
		prods = append(prods, pr)
	}

	// Clientes.
	var clis []cliente
	for i := 0; i < t.Clientes; i++ {
		n, a := nombres[r.IntN(len(nombres))], apellidos[r.IntN(len(apellidos))]
		cl := cliente{id: uuid(r), nombre: n + " " + a, ciudad: ciudadDe(t, r), telefono: telefonoDe(t, r)}
		if r.IntN(3) > 0 {
			e := strings.ToLower(sinTildes(n)+"."+sinTildes(a)) + fmt.Sprint(r.IntN(90)+10) + "@example.com"
			cl.email = &e
		}
		email := "NULL"
		if cl.email != nil {
			email = q(*cl.email)
		}
		alta := creada.Add(time.Duration(r.IntN(int(ahora.Sub(creada).Hours()))) * time.Hour)
		fmt.Fprintf(b, "INSERT INTO clientes (id, tienda_id, nombre, email, telefono, ciudad, provincia, origen, creado_en) VALUES (%s, %s, %s, %s, %s, %s, %s, 'panel', %s);\n",
			q(cl.id), q(tid), q(cl.nombre), email, q(cl.telefono), q(cl.ciudad), q(t.Provincia), ts(alta))
		clis = append(clis, cl)
	}

	// Pedidos: de la apertura a hoy, más los fines de semana y un poco más cada mes.
	numero := 1000
	type vendido struct {
		numero   int
		producto producto
		cliente  cliente
		cuando   time.Time
	}
	var entregados []vendido
	envios := map[string]int64{"andreani": 6900, "correo": 5400, "moto": 3500, "mercadoenvios": 0, "retiro": 0, "oca": 6200}
	for dia := creada.AddDate(0, 0, 3); !dia.After(ahora); dia = dia.AddDate(0, 0, 1) {
		avance := float64(dia.Sub(creada)) / float64(ahora.Sub(creada))
		cuantos := t.PorDia[0] + r.IntN(t.PorDia[1]-t.PorDia[0]+1)
		if wd := dia.Weekday(); wd == time.Saturday || wd == time.Sunday {
			cuantos = int(float64(cuantos) * t.FinDeSemana)
		}
		cuantos = int(float64(cuantos) * (0.7 + 0.6*avance))
		for k := 0; k < cuantos; k++ {
			cuando := time.Date(dia.Year(), dia.Month(), dia.Day(), 9+r.IntN(12), r.IntN(60), r.IntN(60), 0, dia.Location())
			if cuando.After(ahora) {
				continue
			}
			numero++
			canal, pago := elegir(r, t.Canales), elegir(r, t.Pagos)
			envio := elegir(r, t.Envios)
			if canal == "pos" {
				envio, pago = "retiro", []string{"efectivo", "tarjeta", "mp", "modo"}[r.IntN(4)]
			}
			if pago == "efectivo" && canal != "pos" {
				envio = "retiro"
			}
			cl := clis[r.IntN(len(clis))]
			var items []pedidos.Item
			var sub int64
			nItems := 1 + r.IntN(3)
			if t.Rubro == "moda" || t.Rubro == "gastronomia" {
				nItems = 1 + r.IntN(2)
			}
			var primero producto
			for x := 0; x < nItems; x++ {
				p := elegirProducto(t, prods, r)
				if x == 0 {
					primero = p
				}
				it := pedidos.Item{ProductoID: &p.id, Nombre: p.nombre, Foto: &p.foto, Cantidad: 1, Precio: p.precio}
				switch {
				case t.Rubro == "almacen" && r.IntN(3) == 0:
					it.Cantidad = 2 + r.IntN(2)
				case t.Rubro == "gastronomia" && len(p.variantes) == 0 && r.IntN(4) == 0:
					it.Cantidad = 2
				}
				if len(p.variantes) > 0 {
					v := p.variantes[r.IntN(len(p.variantes))]
					it.VarianteID, it.Variante, it.Precio = &v.id, &v.texto, v.precio
				}
				items = append(items, it)
				sub += it.Precio * int64(it.Cantidad)
			}
			costoEnvio := envios[envio]
			edad := ahora.Sub(cuando)
			estado, pasos := estadoPara(edad, envio, r)
			if canal == "pos" {
				// Lo que se vende en el mostrador se cobra y se entrega en el momento.
				estado, pasos = "Entregado", []string{"Para preparar", "Entregado"}
			}
			estadosEn := map[string]string{}
			for i, e := range pasos {
				estadosEn[e] = cuando.Add(time.Duration(i*(2+r.IntN(20))) * time.Hour).UTC().Format(time.RFC3339)
			}
			estadoPago := "Pendiente"
			if estado != "Sin pagar" && estado != "Cancelado" {
				estadoPago = map[string]string{"mp": "Aprobado", "transf": "Acreditada"}[pago]
				if estadoPago == "" {
					estadoPago = "Cobrado"
				}
			}
			ciudad := cl.ciudad
			if envio == "retiro" {
				ciudad = ""
			}
			codigo := "NULL"
			if canal == "web" {
				codigo = q(fmt.Sprintf("%016x%08x", r.Uint64(), r.Uint32()))
			}
			comp := pedidos.Comprador{Nombre: cl.nombre, Email: cl.email, Telefono: &cl.telefono}
			fmt.Fprintf(b, `INSERT INTO pedidos (tienda_id, numero, cliente_id, comprador, canal, pago, estado_pago, envio, ciudad_envio, estado, estados_en,
	items, subtotal, costo_envio, total, creado_en, codigo_publico)
VALUES (%s, %d, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %d, %d, %d, %s, %s);
`, q(tid), numero, q(cl.id), j(comp), q(canal), q(pago), q(estadoPago), q(envio), q(ciudad), q(estado), j(estadosEn),
				j(items), sub, costoEnvio, sub+costoEnvio, ts(cuando), codigo)
			if estado == "Entregado" && canal == "web" {
				entregados = append(entregados, vendido{numero, primero, cl, cuando})
			}
		}
	}

	// Reseñas de compras entregadas, con alguna respuesta de la tienda.
	r.Shuffle(len(entregados), func(a, z int) { entregados[a], entregados[z] = entregados[z], entregados[a] })
	for i, v := range entregados {
		if i >= 22 {
			break
		}
		txt := resenas[t.Rubro][i%len(resenas[t.Rubro])]
		estrellas := []int{5, 5, 5, 5, 4, 4, 5, 3}[r.IntN(8)]
		resp := "NULL"
		if i%3 == 0 {
			resp = q("¡Gracias, " + strings.Split(v.cliente.nombre, " ")[0] + "! Nos alegra mucho que te haya gustado.")
		}
		autor := strings.Split(v.cliente.nombre, " ")
		fmt.Fprintf(b, "INSERT INTO resenas (tienda_id, pedido_numero, producto_id, autor, estrellas, texto, origen, respuesta, creada) VALUES (%s, %d, %s, %s, %d, %s, 'Tienda', %s, %s);\n",
			q(tid), v.numero, q(v.producto.id), q(autor[0]+" "+string([]rune(autor[1])[:1])+"."), estrellas, q(txt), resp, ts(v.cuando.AddDate(0, 0, 4+r.IntN(6))))
	}

	// Preguntas en la tienda online: las viejas respondidas, la última sin responder.
	for i, pq := range preguntas[t.Rubro] {
		p := prods[(i*5+1)%len(prods)]
		cuando := ahora.Add(-time.Duration(2+i*53) * time.Hour)
		resp, en := "NULL", "NULL"
		if i > 0 {
			resp, en = q(pq[1]), ts(cuando.Add(3*time.Hour))
		}
		fmt.Fprintf(b, "INSERT INTO preguntas (tienda_id, producto_id, nombre, pregunta, respuesta, respondida_en, creada) VALUES (%s, %s, %s, %s, %s, %s, %s);\n",
			q(tid), q(p.id), q(nombres[(i*7+3)%len(nombres)]), q(pq[0]), resp, en, ts(cuando))
	}

	// Visitas a la tienda online, día por día, y a cada producto.
	for dia := creada.AddDate(0, 0, 3); !dia.After(ahora); dia = dia.AddDate(0, 0, 1) {
		avance := float64(dia.Sub(creada)) / float64(ahora.Sub(creada))
		vistas := int(float64(90+r.IntN(160)) * (0.6 + 0.9*avance))
		fmt.Fprintf(b, "INSERT INTO visitas (tienda_id, dia, pagina, vistas, visitantes) VALUES (%s, %s, 'tienda', %d, %d);\n",
			q(tid), q(dia.Format("2006-01-02")), vistas, vistas*6/10)
		for _, p := range prods {
			if r.IntN(3) == 0 {
				continue
			}
			v := 2 + r.IntN(vistas/12+1)
			fmt.Fprintf(b, "INSERT INTO visitas (tienda_id, dia, pagina, producto_id, vistas, visitantes) VALUES (%s, %s, 'producto', %s, %d, %d);\n",
				q(tid), q(dia.Format("2006-01-02")), q(p.id), v, v*7/10)
		}
	}

	// Cupones.
	fmt.Fprintf(b, `INSERT INTO cupones (tienda_id, codigo, tipo, valor, descripcion, usos, vendio, desconto, canales, creado_en) VALUES
	(%s, 'BIENVENIDA10', 'pct', 10, '10%% en tu primera compra', %d, %d, %d, '{web}', %s),
	(%s, 'ENVIOGRATIS', 'envio', 0, 'Envío gratis en compras grandes', %d, %d, %d, '{web}', %s);
`, q(tid), 30+r.IntN(40), 2_000_000+r.Int64N(3_000_000), 180_000+r.Int64N(200_000), ts(creada.AddDate(0, 0, 5)),
		q(tid), 12+r.IntN(20), 1_200_000+r.Int64N(2_000_000), 60_000+r.Int64N(90_000), ts(creada.AddDate(0, 1, 0)))
	return nil
}

// elegirProducto: en el café casi todo sale de la carta; el café en grano (los últimos seis del
// catálogo) es una de cada doce compras.
func elegirProducto(t Tienda, prods []producto, r *rand.Rand) producto {
	if t.Rubro == "gastronomia" && len(prods) > 6 {
		if r.IntN(12) == 0 {
			return prods[len(prods)-6+r.IntN(6)]
		}
		return prods[r.IntN(len(prods)-6)]
	}
	return prods[r.IntN(len(prods))]
}

// estadoPara: lo viejo ya se entregó (o se canceló); lo de los últimos días está en camino.
func estadoPara(edad time.Duration, envio string, r *rand.Rand) (string, []string) {
	c := pedidos.CircuitoDe(envio)
	dias := edad.Hours() / 24
	if r.IntN(40) == 0 && dias > 1 {
		return "Cancelado", []string{"Sin pagar", "Cancelado"}
	}
	var i int
	switch {
	case dias > 6:
		i = len(c) - 1
	case dias > 3:
		i = len(c) - 2 + r.IntN(2)
	case dias > 1:
		i = 1 + r.IntN(len(c)-2)
	default:
		i = r.IntN(3)
	}
	return c[i], c[:i+1]
}

func ciudadDe(t Tienda, r *rand.Rand) string {
	if t.Provincia == "Córdoba" {
		return []string{"Córdoba", "Córdoba", "Villa Carlos Paz", "Río Cuarto", "Alta Gracia", "Villa Allende"}[r.IntN(6)]
	}
	return []string{"CABA", "CABA", "CABA", "Vicente López", "San Isidro", "Avellaneda", "Lanús", "Quilmes", "La Plata", "Morón"}[r.IntN(10)]
}

// telefonoDe: números con la característica 5555, que no se asignan a líneas reales.
func telefonoDe(t Tienda, r *rand.Rand) string {
	area := "11"
	if t.Provincia == "Córdoba" {
		area = "351"
	}
	return fmt.Sprintf("+54 9 %s 5555-%04d", area, 1000+r.IntN(9000))
}

func sinTildes(s string) string {
	return strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ñ", "n", "Á", "a", "É", "e", "Í", "i", "Ó", "o", "Ú", "u", " ", "").Replace(s)
}

// descripcion: el detalle del diseño, completado según el rubro.
func descripcion(t Tienda, p ProductoEjemplo, detalle string) string {
	switch t.Rubro {
	case "moda":
		return "Confeccionado en nuestro taller de Buenos Aires, en series cortas. Mirá la guía de talles antes de comprar: si no te queda, lo cambiás dentro de los 30 días."
	case "gastronomia":
		if detalle != "" {
			return detalle + "."
		}
	case "almacen":
		if detalle != "" {
			return "Presentación de " + detalle + ". Lo fraccionamos en el día, en bolsa compostable. De productores de Córdoba y alrededores."
		}
	}
	return detalle
}

var nombres = []string{"Lucía", "Martina", "Sofía", "Valentina", "Camila", "Julieta", "Florencia", "Agustina", "Paula", "Carolina",
	"Mariana", "Victoria", "Micaela", "Rocío", "Belén", "Natalia", "Juan", "Matías", "Nicolás", "Santiago", "Tomás", "Federico",
	"Martín", "Lucas", "Ignacio", "Diego", "Facundo", "Gonzalo", "Ezequiel", "Pablo", "Emilia", "Guadalupe", "Celeste", "Andrés"}

var apellidos = []string{"González", "Rodríguez", "Fernández", "López", "Martínez", "Pérez", "Gómez", "Sánchez", "Romero", "Díaz",
	"Álvarez", "Torres", "Ruiz", "Ramírez", "Flores", "Acosta", "Benítez", "Medina", "Herrera", "Aguirre", "Pereyra", "Molina",
	"Castro", "Ortiz", "Silva", "Núñez", "Rojas", "Ledesma", "Sosa", "Vega", "Ferreyra", "Cabrera", "Quiroga", "Peralta"}

var resenas = map[string][]string{
	"moda": {
		"Hermosa la tela y la terminación. Me quedó perfecto el talle.",
		"Llegó en dos días y súper bien empaquetado. Ya es mi segunda compra.",
		"La calidad se nota al toque. Vale cada peso.",
		"Me ayudaron por WhatsApp con el talle y acertaron.",
		"El color es igual al de la foto. Muy conforme.",
		"Lo cambié por otro talle sin vueltas. Excelente atención.",
		"Lindo pero me quedó un poco grande; igual lo uso un montón.",
	},
	"gastronomia": {
		"El mejor flat white de Palermo, sin exagerar.",
		"Pedí café en grano para la cafetera y llegó recién tostado.",
		"Las medialunas son una locura. Volvemos todas las semanas.",
		"Rapidísimo el pedido para retirar, estaba listo al llegar.",
		"El Yirgacheffe es floral y delicado, tal cual la descripción.",
		"Muy buena atención y el lugar es hermoso.",
		"Rico todo, el tostado en masa madre un poco chico.",
	},
	"almacen": {
		"Todo fresquísimo y bien fraccionado. Recomiendo.",
		"La miel de monte es increíble, se nota que es de verdad.",
		"Llegó en el día con la moto. Muy prolijos.",
		"Los precios están muy bien comparado con el súper.",
		"Las almendras son enormes. Ya repetí el pedido.",
		"Me encanta que usen bolsas compostables.",
		"Faltaba un producto pero me avisaron y me lo mandaron al otro día.",
	},
}

var preguntas = map[string][][2]string{
	"moda": {
		{"¿Tienen el talle M en otro color?", ""},
		{"¿Hacen envíos a Rosario?", "¡Hola! Sí, enviamos a todo el país por Andreani y Correo Argentino."},
		{"¿El talle es grande o normal?", "Es de calce normal. Si dudás entre dos talles, te recomendamos el más grande."},
		{"¿Se puede retirar por el local?", "¡Claro! Lo retirás en Gurruchaga 1840 de lunes a sábado de 11 a 20."},
		{"¿Aceptan cuotas sin interés?", "Sí, hasta 3 cuotas sin interés con Mercado Pago."},
	},
	"gastronomia": {
		{"¿El café en grano sale molido para moka?", ""},
		{"¿Tienen opciones sin TACC?", "Sí, los alfajores de maicena son sin TACC y tenemos budín sin TACC los fines de semana."},
		{"¿Hacen envíos de café en grano?", "Sí, por moto en CABA y alrededores. Lo tostamos los martes y viernes."},
		{"¿Cuánto dura el café una vez abierto?", "Lo ideal es consumirlo dentro del mes, en un frasco hermético y lejos de la luz."},
	},
	"almacen": {
		{"¿Venden yerba sin palo?", ""},
		{"¿Hacen envíos a Villa Carlos Paz?", "¡Sí! Los jueves enviamos a Carlos Paz y alrededores."},
		{"¿La miel es pura?", "Es miel de monte de productores de Traslasierra, sin agregados."},
		{"¿Tienen precios por mayor?", "Sí, desde 5 kg por producto. Escribinos por WhatsApp y te pasamos la lista."},
		{"¿Las almendras son tostadas?", "Son naturales, sin tostar ni salar."},
	},
}
