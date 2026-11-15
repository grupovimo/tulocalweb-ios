package cupones

import (
	"context"
	"crypto/rand"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// Unico: un cupón de un solo uso que genera una ventana de la tienda (captación o ruleta,
// docs/especificaciones/tienda-promociones.md). Descuento es el del contrato: porcentaje (Valor 1–90),
// monto (Valor en pesos, 100–100000) o envioGratis. Vence a los Dias (1–90). Origen: «ventana:<id>».
type Unico struct {
	Prefijo   string
	Descuento string
	Valor     int
	Dias      int
	Origen    string
	// Nombre de la ventana, para la descripción del panel.
	Nombre string
}

// Origen: «ventana:<id>» o, el de un recordatorio del carrito abandonado, «carrito:<uuid>».
var reOrigen = regexp.MustCompile(`^(ventana:[a-z0-9]{6,16}|carrito:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})$`)

// alfabeto sin letras que se confunden (0/O, 1/I/L).
const alfabeto = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"

func azar(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	for i := range b {
		b[i] = alfabeto[int(b[i])%len(alfabeto)]
	}
	return string(b)
}

// TipoYDescripcion: el tipo de la tabla y el texto corto del premio («10 % off», «$ 2.000 off»,
// «Envío gratis»).
func TipoYDescripcion(descuento string, valor int) (string, string, bool) {
	switch descuento {
	case "porcentaje":
		return "pct", fmt.Sprintf("%d %% off", valor), valor >= 1 && valor <= 90
	case "monto":
		return "fijo", pesos(int64(valor)) + " off", valor >= 100 && valor <= 100000
	case "envioGratis":
		return "envio", "Envío gratis", true
	}
	return "", "", false
}

// CrearUnico guarda el cupón en la transacción (con la tienda fijada) y devuelve su código y cuándo
// vence. El código no repite ninguno vigente de la tienda.
func CrearUnico(ctx context.Context, t pgx.Tx, tiendaID string, u Unico, ahora time.Time) (string, time.Time, error) {
	tipo, desc, ok := TipoYDescripcion(u.Descuento, u.Valor)
	if !ok || !reOrigen.MatchString(u.Origen) {
		return "", time.Time{}, ErrInvalido
	}
	if u.Dias < 1 || u.Dias > 90 {
		u.Dias = 7
	}
	valor := u.Valor
	if tipo == "envio" {
		valor = 0
	}
	prefijo := strings.Map(func(r rune) rune {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return -1
	}, strings.ToUpper(u.Prefijo))
	if len(prefijo) > 6 {
		prefijo = prefijo[:6]
	}
	if len(prefijo) < 2 {
		prefijo = "PROMO"
	}
	vence := ahora.Add(time.Duration(u.Dias) * 24 * time.Hour)
	nombre := strings.TrimSpace(u.Nombre)
	if nombre != "" {
		desc += " · " + nombre
	}
	for range 5 {
		codigo := prefijo + "-" + azar(8)
		var repetido bool
		if err := t.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM cupones WHERE tienda_id = $1 AND codigo = $2)`, tiendaID, codigo).
			Scan(&repetido); err != nil {
			return "", time.Time{}, err
		}
		if repetido {
			continue
		}
		_, err := t.Exec(ctx, `INSERT INTO cupones (tienda_id, codigo, tipo, valor, descripcion, limite, canales, termina, minimo,
				quien, por_cliente, alcance, opciones, origen)
			VALUES ($1, $2, $3, $4, $5, 1, '{web}', $6, 0, 'todos', '1', '{"tipo": "todo"}', '{}', $7)`,
			tiendaID, codigo, tipo, valor, desc, vence, u.Origen)
		return codigo, vence, err
	}
	return "", time.Time{}, fmt.Errorf("cupones: no salió un código libre")
}
