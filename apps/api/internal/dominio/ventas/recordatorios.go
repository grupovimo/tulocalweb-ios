package ventas

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cupones"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/auditoria"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

// Recordatorios del carrito abandonado (docs/decisiones.md §62, migración 00084): hasta tres mails
// con su demora desde que dejó el carrito, como Shopify. El segundo o el tercero pueden llevar un
// cupón único de un uso (origen «carrito:<id>», como los de las ventanas). Se corta si compra o si
// se da de baja.
type Recordatorios struct {
	Activo bool                `json:"activo"`
	Pasos  [3]PasoRecordatorio `json:"pasos"`
}

// PasoRecordatorio: uno de los tres mails.
type PasoRecordatorio struct {
	Activo bool `json:"activo"`
	// Horas desde que dejó el carrito (1 a 168, en orden creciente).
	Horas int        `json:"horas"`
	Cupon *CuponPaso `json:"cupon"`
}

// CuponPaso: el descuento del cupón único (el de cupones.Unico): porcentaje (1–90), monto (pesos,
// 100–100000) o envioGratis; vence a los Dias (1–30).
type CuponPaso struct {
	Descuento string `json:"descuento"`
	Valor     int    `json:"valor"`
	Dias      int    `json:"dias"`
}

// RecordatoriosPorDefecto: 1 h, 24 h y 72 h, sin cupón, apagados hasta que la tienda los prenda.
var RecordatoriosPorDefecto = Recordatorios{Pasos: [3]PasoRecordatorio{{Activo: true, Horas: 1}, {Activo: true, Horas: 24}, {Activo: true, Horas: 72}}}

// ErrRecordatorios: la configuración no va (el mensaje es para la persona).
type ErrRecordatorios struct{ Motivo string }

func (e ErrRecordatorios) Error() string { return e.Motivo }

// Validar revisa los tres pasos.
func (r Recordatorios) Validar() error {
	previo := 0
	for i, p := range r.Pasos {
		if p.Horas < 1 || p.Horas > 168 {
			return ErrRecordatorios{"Cada recordatorio sale entre 1 hora y 7 días después."}
		}
		if p.Horas <= previo {
			return ErrRecordatorios{"Cada recordatorio tiene que salir después del anterior."}
		}
		previo = p.Horas
		if p.Cupon == nil {
			continue
		}
		if i == 0 {
			return ErrRecordatorios{"El cupón va en el segundo o en el tercer recordatorio."}
		}
		if _, _, ok := cupones.TipoYDescripcion(p.Cupon.Descuento, p.Cupon.Valor); !ok {
			return ErrRecordatorios{"Revisá el descuento del cupón: de 1 a 90 %, de $ 100 a $ 100.000 o envío gratis."}
		}
		if p.Cupon.Dias < 1 || p.Cupon.Dias > 30 {
			return ErrRecordatorios{"El cupón vence entre 1 y 30 días después."}
		}
	}
	return nil
}

// LeerRecordatorios: la configuración de la tienda.
func LeerRecordatorios(ctx context.Context, pool tx, tiendaID string) (Recordatorios, error) {
	var r Recordatorios
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var err error
		r, err = recordatoriosEnTx(ctx, t, tiendaID)
		return err
	})
	return r, err
}

func recordatoriosEnTx(ctx context.Context, t pgx.Tx, tiendaID string) (Recordatorios, error) {
	r := RecordatoriosPorDefecto
	err := t.QueryRow(ctx, `SELECT carritos_recordatorios FROM tiendas WHERE id = $1`, tiendaID).Scan(&r)
	return r, err
}

// GuardarRecordatorios valida y guarda; el interruptor también queda en carritos_automaticos.email
// (lo leen las apps de antes).
func GuardarRecordatorios(ctx context.Context, pool tx, tiendaID string, r Recordatorios) error {
	if err := r.Validar(); err != nil {
		return err
	}
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		antes, err := recordatoriosEnTx(ctx, t, tiendaID)
		if err != nil {
			return err
		}
		if _, err := t.Exec(ctx, `UPDATE tiendas SET carritos_recordatorios = $2,
			carritos_automaticos = carritos_automaticos || jsonb_build_object('email', $3::boolean) WHERE id = $1`,
			tiendaID, r, r.Activo); err != nil {
			return err
		}
		texto := "cambió los recordatorios del carrito abandonado"
		switch {
		case r.Activo && !antes.Activo:
			texto = "prendió los recordatorios del carrito abandonado"
		case !r.Activo && antes.Activo:
			texto = "apagó los recordatorios del carrito abandonado"
		}
		return auditoria.Registrar(ctx, t, tiendaID, auditoria.Registro{Accion: "carritos.recordatorios", Entidad: "carritos",
			Texto: texto, Antes: antes, Despues: r})
	})
}

// pasoQueToca: el último paso prendido que ya venció y todavía no salió (si el worker estuvo parado,
// sale uno solo, el más reciente). -1: ninguno.
func pasoQueToca(r Recordatorios, paso int, actualizado, ahora time.Time) int {
	toca := -1
	for i := paso; i < len(r.Pasos); i++ {
		p := r.Pasos[i]
		if p.Activo && !actualizado.Add(time.Duration(p.Horas)*time.Hour).After(ahora) {
			toca = i
		}
	}
	return toca
}

// Separación mínima entre un recordatorio a mano y uno automático, para no mandar dos seguidos.
const separacion = 6 * time.Hour

// RecordatoriosAutomaticos encola los recordatorios que vencieron, de todas las tiendas que los
// prendieron (el worker, cada 5 minutos). Cada carrito se bloquea y avanza su paso en la misma
// transacción que encola el mail y crea el cupón: dos workers no mandan dos veces.
func RecordatoriosAutomaticos(ctx context.Context, pool interface {
	tx
	EnTx(context.Context, func(pgx.Tx) error) error
}, ahora time.Time) (int, error) {
	type par struct{ tienda, id string }
	var pares []par
	err := pool.EnTx(ctx, func(t pgx.Tx) error {
		filas, err := t.Query(ctx, `SELECT tienda_id::text, id::text FROM app.carritos_para_recordar(200)`)
		if err != nil {
			return err
		}
		defer filas.Close()
		for filas.Next() {
			var p par
			if err := filas.Scan(&p.tienda, &p.id); err != nil {
				return err
			}
			pares = append(pares, p)
		}
		return filas.Err()
	})
	if err != nil {
		return 0, err
	}
	n := 0
	for _, p := range pares {
		ok, err := recordarPaso(ctx, pool, p.tienda, p.id, ahora)
		if err != nil {
			return n, err
		}
		if ok {
			n++
		}
	}
	return n, nil
}

func recordarPaso(ctx context.Context, pool tx, tienda, id string, ahora time.Time) (bool, error) {
	enviado := false
	err := pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		r, err := recordatoriosEnTx(ctx, t, tienda)
		if err != nil || !r.Activo {
			return err
		}
		var paso int
		var actualizado time.Time
		var recordado *time.Time
		var cupon *string
		err = t.QueryRow(ctx, `SELECT paso, actualizado, recordado_en, cupon FROM carritos
			WHERE id = $1::uuid AND recuperado_en IS NULL AND baja_en IS NULL FOR UPDATE`, id).Scan(&paso, &actualizado, &recordado, &cupon)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		i := pasoQueToca(r, paso, actualizado, ahora)
		if i < 0 || (recordado != nil && ahora.Sub(*recordado) < separacion) {
			return nil
		}
		datos := map[string]any{"id": id, "paso": i + 1}
		if c := r.Pasos[i].Cupon; c != nil {
			if cupon == nil {
				codigo, _, err := cupones.CrearUnico(ctx, t, tienda, cupones.Unico{Prefijo: "VUELVE", Descuento: c.Descuento,
					Valor: c.Valor, Dias: c.Dias, Origen: "carrito:" + id, Nombre: "Carrito abandonado"}, ahora)
				if err != nil {
					return fmt.Errorf("carritos: cupón del recordatorio: %w", err)
				}
				cupon = &codigo
			}
			datos["cupon"] = *cupon
		}
		if _, err := t.Exec(ctx, `UPDATE carritos SET paso = $2, recordado_en = $3, recordatorios = recordatorios + 1, cupon = $4
			WHERE id = $1::uuid`, id, i+1, ahora, cupon); err != nil {
			return err
		}
		enviado = true
		return outbox.Publicar(ctx, t, tienda, "carrito.recordar", datos)
	})
	return enviado, err
}

// BajaCarrito: quien recibió un recordatorio pidió no recibir más. Corta los de ese email en la
// tienda, también los de un carrito que arme después. Idempotente; un token que no existe no avisa
// nada (la respuesta es la misma).
func BajaCarrito(ctx context.Context, pool tx, tienda, token string) error {
	return pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `UPDATE carritos SET baja_en = coalesce(baja_en, now())
			WHERE lower(email) = (SELECT lower(email) FROM carritos WHERE token = $1)`, token)
		return err
	})
}
