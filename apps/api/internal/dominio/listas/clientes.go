package listas

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type tx interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
}

// cliente es lo que se junta de un cliente antes de armar su fila: de los pedidos, del historial
// importado y de su ficha.
type cliente struct {
	clave, nombre, ciudad, canal string
	foto, tel                    *string
	pedidos                      int
	ultima                       time.Time
	gastado                      int64
	mayorista                    bool
	fechas                       []time.Time
}

// claveCliente: el email en minúsculas o, sin email, el nombre (como agrupa la lista). Es el id
// de la fila y lo que recibe el historial.
const claveCliente = `coalesce(lower(nullif(%s, '')), %s)`

// Clientes arma la lista uniendo, por email en minúsculas, los clientes de los pedidos (agrupados
// por email, o por nombre si no hay), las fichas propias (importadas o cargadas en el panel, aunque
// no tengan pedidos) y el historial importado, que cuenta para pedidos, gastado, última compra y
// segmento.
// anonimo: la venta sin datos de quien compra (el local cobra a "Consumidor final"). No es un
// cliente: contarla como tal sumaba un «Consumidor final» a la lista con cada venta de mostrador.
const anonimo = `(comprador->>'nombre' = 'Consumidor final' AND coalesce(comprador->>'email', '') = ''
	AND coalesce(comprador->>'telefono', '') = '')`

func Clientes(ctx context.Context, pool tx, tiendaID string, ahora time.Time) ([]Registro, error) {
	por := map[string]*cliente{}
	orden := []string{}
	de := func(clave string) *cliente {
		c := por[clave]
		if c == nil {
			c = &cliente{clave: clave}
			por[clave] = c
			orden = append(orden, clave)
		}
		return c
	}
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		filas, err := t.Query(ctx, `
			SELECT `+fmt.Sprintf(claveCliente, "comprador->>'email'", "comprador->>'nombre'")+` AS clave,
				(array_agg(comprador->>'nombre' ORDER BY creado_en DESC))[1],
				(array_agg(comprador->>'foto' ORDER BY creado_en DESC))[1],
				(array_agg(comprador->>'telefono' ORDER BY creado_en DESC))[1],
				(array_agg(ciudad_envio ORDER BY creado_en DESC))[1],
				(array_agg(canal ORDER BY creado_en))[1],
				count(*), max(creado_en), sum(total), bool_or(mayorista),
				array_agg(creado_en)
			FROM pedidos WHERE tienda_id = $1 AND estado <> 'Cancelado' AND NOT `+anonimo+`
			GROUP BY 1`, tiendaID)
		if err != nil {
			return err
		}
		for filas.Next() {
			var clave string
			var c cliente
			if err := filas.Scan(&clave, &c.nombre, &c.foto, &c.tel, &c.ciudad, &c.canal, &c.pedidos, &c.ultima, &c.gastado,
				&c.mayorista, &c.fechas); err != nil {
				return err
			}
			c.clave = clave
			*de(clave) = c
		}
		if err := filas.Err(); err != nil {
			return err
		}
		// El historial importado: suma pedidos, gastado y fechas al cliente del mismo email.
		filas, err = t.Query(ctx, `
			SELECT `+fmt.Sprintf(claveCliente, "cliente_email", "nullif(cliente_nombre, '')")+` AS clave,
				(array_agg(cliente_nombre ORDER BY fecha DESC))[1], count(*), max(fecha), sum(total), array_agg(fecha)
			FROM historial_pedidos WHERE tienda_id = $1 AND (cliente_email IS NOT NULL OR cliente_nombre <> '')
			GROUP BY 1`, tiendaID)
		if err != nil {
			return err
		}
		for filas.Next() {
			var clave, nombre string
			var n int
			var ultima time.Time
			var gastado int64
			var fechas []time.Time
			if err := filas.Scan(&clave, &nombre, &n, &ultima, &gastado, &fechas); err != nil {
				return err
			}
			c := de(clave)
			if c.nombre == "" {
				c.nombre, c.canal = nombre, "web"
			}
			c.pedidos += n
			c.gastado += gastado
			c.fechas = append(c.fechas, fechas...)
			if ultima.After(c.ultima) {
				c.ultima = ultima
			}
		}
		if err := filas.Err(); err != nil {
			return err
		}
		// Las fichas: completan teléfono y ciudad, y las que no tienen compras entran igual. Sin
		// email, su id es el de la ficha.
		filas, err = t.Query(ctx, `SELECT coalesce(lower(email), id::text), nombre, telefono, ciudad, origen, creado_en
			FROM clientes WHERE tienda_id = $1 ORDER BY creado_en`, tiendaID)
		if err != nil {
			return err
		}
		for filas.Next() {
			var clave, nombre, ciudad, origen string
			var tel *string
			var creado time.Time
			if err := filas.Scan(&clave, &nombre, &tel, &ciudad, &origen, &creado); err != nil {
				return err
			}
			c := de(clave)
			if c.nombre == "" {
				c.nombre = nombre
				// Cargado en el panel: llegó por el local, como en la demo; importado, por la tienda online.
				c.canal = "web"
				if origen == "panel" {
					c.canal = "pos"
				}
			}
			if c.tel == nil || *c.tel == "" {
				c.tel = tel
			}
			if c.ciudad == "" {
				c.ciudad = ciudad
			}
			if c.ultima.IsZero() {
				c.ultima = creado
			}
		}
		return filas.Err()
	})
	if err != nil {
		return nil, err
	}
	regs := make([]Registro, 0, len(orden))
	for _, k := range orden {
		c := por[k]
		dias := max(0, ahora.Sub(c.ultima).Hours()/24)
		meses := make([]int, 12)
		for _, f := range c.fechas {
			k := 11 - int(ahora.Sub(f).Hours()/24/30)
			if k >= 0 && k < 12 && meses[k] < 3 {
				meses[k]++
			}
		}
		h := fnv.New32a()
		_, _ = h.Write([]byte(c.clave))
		// El email es la clave cuando la hay (para corregirlo desde la ficha).
		var email *string
		if strings.Contains(c.clave, "@") {
			email = &c.clave
		}
		regs = append(regs, Registro{
			"id": c.clave, "nombre": c.nombre, "email": email, "foto": c.foto, "ini": int(h.Sum32() % 5), "ciudad": c.ciudad,
			"canal": c.canal, "telefono": c.tel, "pedidos": c.pedidos, "ultima": dias, "gastado": c.gastado,
			"seg": segmento(c.pedidos, dias), "meses": meses, "cuit": c.mayorista, "wa": c.tel != nil && *c.tel != "",
			"cupon": false, "etiquetas": []string{},
		})
	}
	return regs, nil
}

// PedidoHistorial es un pedido importado de otra plataforma (components/schemas/PedidoHistorial).
type PedidoHistorial struct {
	ID         string          `json:"id"`
	Plataforma string          `json:"plataforma"`
	Numero     string          `json:"numero"`
	Fecha      time.Time       `json:"fecha"`
	Total      int64           `json:"total"`
	Items      json.RawMessage `json:"items"`
	Estado     string          `json:"estado"`
}

// Historial devuelve los pedidos importados de un cliente de la lista (por su id: el email en
// minúsculas, el nombre si compró sin email, o el id de su ficha), del más nuevo al más viejo.
func Historial(ctx context.Context, pool tx, tiendaID, id string) ([]PedidoHistorial, error) {
	l := []PedidoHistorial{}
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		// El id de una ficha sin email no tiene historial propio; con email, se busca por el email.
		var email *string
		_ = t.QueryRow(ctx, `SELECT lower(email) FROM clientes WHERE tienda_id = $1 AND id::text = $2`, tiendaID, id).Scan(&email)
		clave := id
		if email != nil {
			clave = *email
		}
		filas, err := t.Query(ctx, `SELECT id::text, plataforma, numero_externo, fecha, total, items, estado
			FROM historial_pedidos WHERE tienda_id = $1 AND `+fmt.Sprintf(claveCliente, "cliente_email", "nullif(cliente_nombre, '')")+` = $2
			ORDER BY fecha DESC LIMIT 200`, tiendaID, clave)
		if err != nil {
			return err
		}
		for filas.Next() {
			var p PedidoHistorial
			if err := filas.Scan(&p.ID, &p.Plataforma, &p.Numero, &p.Fecha, &p.Total, &p.Items, &p.Estado); err != nil {
				return err
			}
			l = append(l, p)
		}
		return filas.Err()
	})
	return l, err
}

// CrearCliente guarda la ficha de un cliente cargado en el panel (el alta de Clientes). Con un
// email que ya tiene ficha no crea otra: devuelve 0.
func CrearCliente(ctx context.Context, pool tx, tiendaID string, v json.RawMessage) (int, error) {
	var c struct{ Nombre, Telefono, Email, Ciudad string }
	if json.Unmarshal(v, &c) != nil {
		return 0, nil
	}
	nombre := strings.TrimSpace(c.Nombre)
	if n := len([]rune(nombre)); n < 2 || n > 80 {
		return 0, nil
	}
	nulo := func(s string, max int) any {
		s = strings.TrimSpace(s)
		if s == "" || len([]rune(s)) > max {
			return nil
		}
		return s
	}
	email := nulo(strings.ToLower(c.Email), 120)
	if e, ok := email.(string); ok && (len(e) < 3 || !strings.Contains(e, "@")) {
		email = nil
	}
	ciudad, _ := nulo(c.Ciudad, 60).(string)
	n := 0
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		tag, err := t.Exec(ctx, `INSERT INTO clientes (tienda_id, nombre, email, telefono, ciudad, origen)
			VALUES ($1, $2, $3, $4, $5, 'panel') ON CONFLICT DO NOTHING`, tiendaID, nombre, email, nulo(c.Telefono, 30), ciudad)
		n = int(tag.RowsAffected())
		return err
	})
	return n, err
}

// segmento: los mismos cortes que el handoff (PanelListas).
func segmento(pedidos int, dias float64) string {
	switch {
	case pedidos >= 12 && dias < 60:
		return "vip"
	case pedidos >= 4 && dias < 90:
		return "freq"
	case pedidos <= 1 && dias < 30:
		return "new"
	case dias > 180:
		return "sleep"
	case dias > 75:
		return "risk"
	}
	return "freq"
}

// DefClientes: cómo filtra y ordena Clientes.
var DefClientes = Def{
	Buscar: func(r Registro) string {
		t, _ := r["telefono"].(*string)
		tel := ""
		if t != nil {
			tel = *t
		}
		return strings.Join([]string{r["nombre"].(string), r["ciudad"].(string), tel}, " ")
	},
	Segmento: "seg", Facetas: []string{"canal", "ciudad"}, Rango: "gastado", Banderas: []string{"wa", "cuit"},
	Ordenes: []Orden{
		{"ltv", func(a, b Registro) bool { return numero(a["gastado"]) > numero(b["gastado"]) }},
		{"last", func(a, b Registro) bool { return numero(a["ultima"]) < numero(b["ultima"]) }},
		{"orders", func(a, b Registro) bool { return numero(a["pedidos"]) > numero(b["pedidos"]) }},
		{"az", func(a, b Registro) bool { return normal(a["nombre"].(string)) < normal(b["nombre"].(string)) }},
	},
	Resumen: func(todos, l []Registro) map[string]float64 {
		r := map[string]float64{}
		for _, x := range l {
			r["gastado"] += numero(x["gastado"])
		}
		for _, x := range todos {
			if x["seg"] == "risk" {
				r["enRiesgo"]++
			}
		}
		return r
	},
}
