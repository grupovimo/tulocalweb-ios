package facturacion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/arca"
)

const maxIntentos = 8

// espera crece con los intentos y se acota a cinco minutos: una caída de ARCA de más de dos horas
// no es un reintento, es un incidente que se ve en el panel.
func espera(intentos int) time.Duration {
	d := time.Duration(1<<min(intentos, 10)) * 15 * time.Second
	return min(d, 5*time.Minute)
}

type pendiente struct {
	id              string
	pedido          *int
	tipo, pv        int
	numeroIntentado *int
	receptor        struct {
		Nombre    string `json:"nombre"`
		DocTipo   int    `json:"docTipo"`
		DocNro    any    `json:"docNro"`
		Condicion int    `json:"condicion"`
	}
	total    int64
	asociado *struct{ tipo, pv, numero int }
	intentos int
}

// Drenar emite los pendientes de las tiendas que tienen (para el worker). Una tienda con problemas
// (certificado vencido, ARCA caído) no frena a las demás.
func (s *Servicio) Drenar(ctx context.Context, ahora time.Time) error {
	var tiendas []string
	err := s.Pool.EnTx(ctx, func(t pgx.Tx) error {
		filas, err := t.Query(ctx, `SELECT id::text FROM app.tiendas_con_facturas_pendientes(20) id`)
		if err != nil {
			return err
		}
		defer filas.Close()
		for filas.Next() {
			var id string
			if err := filas.Scan(&id); err != nil {
				return err
			}
			tiendas = append(tiendas, id)
		}
		return filas.Err()
	})
	if err != nil {
		return err
	}
	for _, t := range tiendas {
		if err := s.ProcesarTienda(ctx, t, ahora); err != nil {
			slog.Warn("facturacion: tienda", "tienda", t, "error", err)
		}
	}
	return nil
}

// ProcesarTienda emite los pendientes de una tienda, de a uno y en orden (la numeración es
// correlativa por tipo y punto de venta). Un lock de la tienda evita que dos procesos numeren a la vez.
func (s *Servicio) ProcesarTienda(ctx context.Context, tienda string, ahora time.Time) error {
	conn, err := s.Pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	var ok bool
	if err := conn.QueryRow(ctx, `SELECT pg_try_advisory_lock(hashtext('facturas:' || $1))`, tienda).Scan(&ok); err != nil {
		return err
	}
	if !ok {
		return nil // otro proceso ya la está emitiendo
	}
	defer conn.Exec(context.WithoutCancel(ctx), `SELECT pg_advisory_unlock(hashtext('facturas:' || $1))`, tienda) //nolint:errcheck

	var cuit string
	var clave []byte
	var cert *string
	var amb arca.Ambiente
	var pendientes []pendiente
	err = s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		var err error
		cuit, clave, cert, amb, _, err = s.clave(ctx, t, tienda)
		if err != nil {
			return err
		}
		filas, err := t.Query(ctx, `SELECT c.id::text, c.pedido_numero, c.tipo, c.punto_venta, c.numero_intentado, c.receptor, c.total, c.intentos,
				a.tipo, a.punto_venta, a.numero
			FROM comprobantes c LEFT JOIN comprobantes a ON a.id = c.asociado
			WHERE c.tienda_id = $1 AND c.estado = 'pendiente' AND c.reintentar_en <= $2 AND c.intentos < $3
			ORDER BY c.creado_en LIMIT 20`, tienda, ahora, maxIntentos)
		if err != nil {
			return err
		}
		defer filas.Close()
		for filas.Next() {
			var p pendiente
			var rec []byte
			var at, apv, an *int
			if err := filas.Scan(&p.id, &p.pedido, &p.tipo, &p.pv, &p.numeroIntentado, &rec, &p.total, &p.intentos, &at, &apv, &an); err != nil {
				return err
			}
			_ = json.Unmarshal(rec, &p.receptor)
			if at != nil && apv != nil && an != nil {
				p.asociado = &struct{ tipo, pv, numero int }{*at, *apv, *an}
			}
			pendientes = append(pendientes, p)
		}
		return filas.Err()
	})
	if errors.Is(err, ErrSinARCA) || (err == nil && cert == nil) {
		return s.marcarTodos(ctx, tienda, "La tienda todavía no conectó ARCA.")
	}
	if err != nil || len(pendientes) == 0 {
		return err
	}
	tk, err := s.ticket(ctx, tienda, clave, *cert, amb, ahora)
	if err != nil {
		for _, p := range pendientes {
			s.fallo(ctx, tienda, p, err)
		}
		return err
	}
	for _, p := range pendientes {
		s.emitir(ctx, tienda, cuit, amb, tk, p, ahora)
	}
	return nil
}

func (s *Servicio) emitir(ctx context.Context, tienda, cuit string, amb arca.Ambiente, tk arca.Ticket, p pendiente, ahora time.Time) {
	dir := s.dir(amb)
	ult, err := arca.UltimoAutorizado(ctx, s.cli(), dir, tk, cuit, p.pv, p.tipo)
	if err != nil {
		s.fallo(ctx, tienda, p, err)
		return
	}
	c := arca.Comprobante{Tipo: p.tipo, PuntoVenta: p.pv, DocTipo: p.receptor.DocTipo, CondReceptor: p.receptor.Condicion, Fecha: ahora, Total: p.total}
	if c.DocTipo == 0 {
		c.DocTipo = arca.DocConsumidorFin
	}
	if c.CondReceptor == 0 {
		c.CondReceptor = arca.CondConsumidorFinal
	}
	if n, err := strconv.ParseInt(fmt.Sprint(p.receptor.DocNro), 10, 64); err == nil {
		c.DocNro = n
	}
	if p.asociado != nil {
		c.Asociado = &arca.Asociado{Tipo: p.asociado.tipo, PuntoVenta: p.asociado.pv, Numero: p.asociado.numero}
	}
	// La respuesta de un intento anterior no llegó: si ARCA ya numeró ese comprobante, se recupera
	// con su CAE en vez de emitir otro por la misma venta.
	if p.numeroIntentado != nil && ult >= *p.numeroIntentado {
		q, err := arca.Consultar(ctx, s.cli(), dir, tk, cuit, p.pv, p.tipo, *p.numeroIntentado)
		if err != nil {
			s.fallo(ctx, tienda, p, err)
			return
		}
		// Además del total: el mismo receptor, y que ningún otro comprobante de la tienda ya tenga ese
		// número (si el intento no se numeró y otro de la tanda tomó el número, era la factura de otra
		// venta con el mismo total, y quedaban dos pedidos con la misma factura).
		otro := false
		if err := s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
			return t.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM comprobantes WHERE id::text <> $1 AND punto_venta = $2 AND tipo = $3
				AND (numero = $4 OR (estado = 'pendiente' AND numero_intentado = $4)))`, p.id, p.pv, p.tipo, *p.numeroIntentado).Scan(&otro)
		}); err != nil {
			s.fallo(ctx, tienda, p, err)
			return
		}
		if q.CAE != "" && mismoTotal(q.Total, c.Total) && q.DocNro == c.DocNro && !otro {
			s.emitido(ctx, tienda, p, arca.Autorizado{Numero: *p.numeroIntentado, CAE: q.CAE, VenceCAE: q.VenceCAE})
			return
		}
		s.rechazado(ctx, tienda, p, arca.Rechazado{Codigo: "revisar",
			Mensaje: fmt.Sprintf("ARCA tiene otro comprobante con el número %d: revisalo antes de reintentar.", *p.numeroIntentado)})
		return
	}
	numero := ult + 1
	if err := s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `UPDATE comprobantes SET numero_intentado = $2 WHERE id::text = $1`, p.id, numero)
		return err
	}); err != nil {
		s.fallo(ctx, tienda, p, err)
		return
	}
	a, err := arca.SolicitarCAE(ctx, s.cli(), dir, tk, cuit, c, numero)
	var r arca.Rechazado
	switch {
	case errors.As(err, &r):
		s.rechazado(ctx, tienda, p, r)
	case err != nil:
		s.fallo(ctx, tienda, p, err)
	default:
		s.emitido(ctx, tienda, p, a)
	}
}

func (s *Servicio) emitido(ctx context.Context, tienda string, p pendiente, a arca.Autorizado) {
	err := s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		if _, err := t.Exec(ctx, `UPDATE comprobantes SET estado = 'emitido', numero = $2, cae = $3, cae_vence = $4, emitido_en = now(),
				motivo = NULL, codigo_arca = NULL WHERE id::text = $1`, p.id, a.Numero, a.CAE, a.VenceCAE); err != nil {
			return err
		}
		if p.pedido == nil {
			return nil
		}
		_, err := t.Exec(ctx, `UPDATE pedidos SET facturado = true, factura = $2 WHERE tienda_id = $3 AND numero = $1`,
			*p.pedido, fmt.Sprintf("%s %04d-%08d", letra(p.tipo), p.pv, a.Numero), tienda)
		return err
	})
	if err != nil {
		slog.Error("facturacion: guardar emitido", "error", err, "cae", a.CAE)
	}
}

func (s *Servicio) rechazado(ctx context.Context, tienda string, p pendiente, r arca.Rechazado) {
	_ = s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `UPDATE comprobantes SET estado = 'rechazado', motivo = left($2, 300), codigo_arca = $3,
			numero_intentado = NULL WHERE id::text = $1`, p.id, r.Mensaje, r.Codigo)
		return err
	})
}

func (s *Servicio) fallo(ctx context.Context, tienda string, p pendiente, err error) {
	motivo := "No pudimos hablar con ARCA: reintentamos solos."
	var e arca.ErrARCA
	if errors.As(err, &e) {
		motivo = "ARCA: " + e.Mensaje
	} else if errors.Is(err, ErrSinCifrado) {
		motivo = "Falta configurar el cifrado del servidor."
	}
	_ = s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `UPDATE comprobantes SET intentos = intentos + 1, reintentar_en = now() + $2::interval, motivo = left($3, 300)
			WHERE id::text = $1`, p.id, espera(p.intentos+1).String(), motivo)
		return err
	})
}

// marcarTodos deja "no facturado" con su motivo los pendientes de una tienda que ya no tiene ARCA.
func (s *Servicio) marcarTodos(ctx context.Context, tienda, motivo string) error {
	return s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `UPDATE comprobantes SET estado = 'no_facturado', motivo = $2 WHERE tienda_id = $1 AND estado = 'pendiente'`,
			tienda, motivo)
		return err
	})
}

// mismoTotal compara el ImpTotal que devuelve ARCA ("12100", "12100.0", "12100.00") con el total en
// pesos, por valor y no por texto.
func mismoTotal(texto string, pesos int64) bool {
	v, err := strconv.ParseFloat(strings.TrimSpace(texto), 64)
	return err == nil && !math.IsInf(v, 0) && !math.IsNaN(v) && int64(math.Round(v*100)) == pesos*100
}
