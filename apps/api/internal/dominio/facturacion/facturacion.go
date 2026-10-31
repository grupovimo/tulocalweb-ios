// Package facturacion: los comprobantes de cada tienda con ARCA. La configuración (clave, pedido de
// certificado, certificado), la factura manual, la lista del panel y el drenador que pide los CAE.
// Las ventas cobradas se encolan solas en la base (trigger app.encolar_factura, misma transacción
// que el cobro); acá se emiten.
package facturacion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/arca"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/auditoria"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/cifrado"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db"
)

var (
	ErrInvalido    = errors.New("facturacion: inválido")
	ErrSinPermiso  = errors.New("facturacion: sin permiso")
	ErrSinARCA     = errors.New("facturacion: la tienda no conectó ARCA")
	ErrSinCifrado  = errors.New("facturacion: falta la clave de cifrado del servidor")
	ErrCertificado = errors.New("facturacion: certificado")
)

type Servicio struct {
	Pool  *db.Pool
	Cifra *cifrado.Cifrador
	HTTP  *http.Client
	// Direcciones de ARCA por ambiente (los tests apuntan a un ARCA falso).
	Direcciones func(arca.Ambiente) arca.Direcciones
}

func (s *Servicio) cli() *http.Client {
	if s.HTTP != nil {
		return s.HTTP
	}
	return &http.Client{Timeout: 30 * time.Second}
}

func (s *Servicio) dir(a arca.Ambiente) arca.Direcciones {
	if s.Direcciones != nil {
		return s.Direcciones(a)
	}
	return arca.DireccionesDe(a)
}

func contexto(tienda, que string) string { return "arca:" + que + ":" + tienda }

// Config es lo que ve el panel (nunca la clave).
type Config struct {
	Conectada        bool       `json:"conectada"`
	CUIT             string     `json:"cuit"`
	RazonSocial      string     `json:"razonSocial"`
	Condicion        string     `json:"condicion"`
	PuntoVenta       int        `json:"puntoVenta"`
	Ambiente         string     `json:"ambiente"`
	Certificado      bool       `json:"certificado"`
	CertificadoVence *time.Time `json:"certificadoVence"`
	FacturarAlCobrar bool       `json:"facturarAlCobrar"`
}

func (s *Servicio) Leer(ctx context.Context, tienda string) (Config, error) {
	var c Config
	var cert *string
	err := s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT cuit, razon_social, condicion, punto_venta, ambiente, certificado, certificado_vence, facturar_al_cobrar
			FROM arca_config WHERE tienda_id = $1`, tienda).
			Scan(&c.CUIT, &c.RazonSocial, &c.Condicion, &c.PuntoVenta, &c.Ambiente, &cert, &c.CertificadoVence, &c.FacturarAlCobrar)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Config{FacturarAlCobrar: true, Ambiente: string(arca.Homologacion)}, nil
	}
	c.Certificado = cert != nil
	c.Conectada = c.Certificado
	return c, err
}

// Alta son los datos para generar la clave y el pedido de certificado.
type Alta struct {
	CUIT        string `json:"cuit"`
	RazonSocial string `json:"razonSocial"`
	Condicion   string `json:"condicion"`
	PuntoVenta  int    `json:"puntoVenta"`
	Ambiente    string `json:"ambiente"`
}

func duenia(rol string) bool { return rol == "duenio" || rol == "admin" }

// GenerarPedido crea la clave (cifrada, no sale nunca) y devuelve el CSR para subir a ARCA. Si ya
// había una clave, la reemplaza: el certificado anterior deja de servir.
func (s *Servicio) GenerarPedido(ctx context.Context, tienda, rol string, a Alta) ([]byte, error) {
	if !duenia(rol) {
		return nil, ErrSinPermiso
	}
	if s.Cifra == nil {
		return nil, ErrSinCifrado
	}
	a.CUIT = strings.NewReplacer("-", "", " ", "", ".", "").Replace(a.CUIT)
	if !arca.CUITValido(a.CUIT) || (a.Condicion != "monotributo" && a.Condicion != "responsable_inscripto") ||
		a.PuntoVenta < 1 || a.PuntoVenta > 99998 || (a.Ambiente != "homologacion" && a.Ambiente != "produccion") {
		return nil, ErrInvalido
	}
	clave, err := arca.NuevaClave()
	if err != nil {
		return nil, err
	}
	csr, err := arca.PedidoCertificado(clave, a.CUIT, a.RazonSocial, "mostrador")
	if err != nil {
		return nil, ErrInvalido
	}
	cif, err := s.Cifra.Cifrar(contexto(tienda, "clave"), clave)
	if err != nil {
		return nil, err
	}
	err = s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		if _, err := t.Exec(ctx, `DELETE FROM arca_tickets WHERE tienda_id = $1`, tienda); err != nil {
			return err
		}
		_, err := t.Exec(ctx, `INSERT INTO arca_config (tienda_id, cuit, razon_social, condicion, punto_venta, ambiente, clave)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (tienda_id) DO UPDATE SET cuit = $2, razon_social = $3, condicion = $4, punto_venta = $5,
				ambiente = $6, clave = $7, certificado = NULL, certificado_vence = NULL, aviso_vence = NULL, actualizada_en = now()`,
			tienda, a.CUIT, strings.TrimSpace(a.RazonSocial), a.Condicion, a.PuntoVenta, a.Ambiente, cif)
		if err != nil {
			return err
		}
		return auditoria.Registrar(ctx, t, tienda, auditoria.Registro{Accion: "integracion.arca_pedido", Entidad: "integracion",
			EntidadID: "arca", Texto: fmt.Sprintf("empezó a conectar ARCA (CUIT %s, punto de venta %d)", auditoria.Final(a.CUIT), a.PuntoVenta)})
	})
	return csr, err
}

func (s *Servicio) clave(ctx context.Context, t pgx.Tx, tienda string) (cuit string, clave []byte, cert *string, amb arca.Ambiente, pv int, err error) {
	var cif []byte
	var a string
	err = t.QueryRow(ctx, `SELECT cuit, clave, certificado, ambiente, punto_venta FROM arca_config WHERE tienda_id = $1`, tienda).
		Scan(&cuit, &cif, &cert, &a, &pv)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil, nil, "", 0, ErrSinARCA
	}
	if err != nil {
		return
	}
	if s.Cifra == nil {
		return "", nil, nil, "", 0, ErrSinCifrado
	}
	clave, err = s.Cifra.Descifrar(contexto(tienda, "clave"), cif)
	return cuit, clave, cert, arca.Ambiente(a), pv, err
}

// GuardarCertificado recibe el PEM que devolvió ARCA y comprueba que sea de esta clave y este CUIT.
func (s *Servicio) GuardarCertificado(ctx context.Context, tienda, rol, certPEM string, ahora time.Time) (arca.DatosCertificado, error) {
	var d arca.DatosCertificado
	if !duenia(rol) {
		return d, ErrSinPermiso
	}
	err := s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		cuit, clave, _, _, _, err := s.clave(ctx, t, tienda)
		if err != nil {
			return err
		}
		d, err = arca.RevisarCertificado([]byte(strings.TrimSpace(certPEM)), clave, cuit, ahora)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrCertificado, err)
		}
		if _, err := t.Exec(ctx, `DELETE FROM arca_tickets WHERE tienda_id = $1`, tienda); err != nil {
			return err
		}
		_, err = t.Exec(ctx, `UPDATE arca_config SET certificado = $2, certificado_vence = $3, aviso_vence = NULL, actualizada_en = now() WHERE tienda_id = $1`,
			tienda, strings.TrimSpace(certPEM), d.Vence)
		if err != nil {
			return err
		}
		return auditoria.Registrar(ctx, t, tienda, auditoria.Registro{Accion: "integracion.arca_conectar", Entidad: "integracion",
			EntidadID: "arca", Texto: "conectó ARCA: subió el certificado (vence el " + d.Vence.Format("02/01/2006") + ")"})
	})
	return d, err
}

// Ajustes que se cambian sin tocar el certificado.
type Ajustes struct {
	PuntoVenta       *int  `json:"puntoVenta"`
	FacturarAlCobrar *bool `json:"facturarAlCobrar"`
}

func (s *Servicio) Ajustar(ctx context.Context, tienda, rol string, a Ajustes) error {
	if !duenia(rol) {
		return ErrSinPermiso
	}
	if a.PuntoVenta != nil && (*a.PuntoVenta < 1 || *a.PuntoVenta > 99998) {
		return ErrInvalido
	}
	return s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		tag, err := t.Exec(ctx, `UPDATE arca_config SET punto_venta = coalesce($2, punto_venta),
			facturar_al_cobrar = coalesce($3, facturar_al_cobrar), actualizada_en = now() WHERE tienda_id = $1`,
			tienda, a.PuntoVenta, a.FacturarAlCobrar)
		if err == nil && tag.RowsAffected() == 0 {
			return ErrSinARCA
		}
		if err != nil {
			return err
		}
		var cambios []string
		if a.PuntoVenta != nil {
			cambios = append(cambios, fmt.Sprintf("punto de venta %d", *a.PuntoVenta))
		}
		if a.FacturarAlCobrar != nil {
			cambios = append(cambios, map[bool]string{true: "factura al cobrar", false: "no factura al cobrar"}[*a.FacturarAlCobrar])
		}
		if len(cambios) == 0 {
			return nil
		}
		return auditoria.Registrar(ctx, t, tienda, auditoria.Registro{Accion: "integracion.arca_ajustes", Entidad: "integracion",
			EntidadID: "arca", Texto: "cambió los ajustes de ARCA: " + strings.Join(cambios, ", ")})
	})
}

// ticket devuelve un ticket vigente para el servicio: el guardado o uno nuevo del WSAA.
func (s *Servicio) ticket(ctx context.Context, tienda string, clave []byte, cert string, amb arca.Ambiente, ahora time.Time) (arca.Ticket, error) {
	var tk arca.Ticket
	var cif []byte
	err := s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT ticket FROM arca_tickets WHERE tienda_id = $1 AND servicio = 'wsfe' AND vence > $2`,
			tienda, ahora.Add(10*time.Minute)).Scan(&cif)
	})
	if err == nil {
		plano, err := s.Cifra.Descifrar(contexto(tienda, "ticket"), cif)
		if err == nil && json.Unmarshal(plano, &tk) == nil && tk.Vigente(ahora) {
			return tk, nil
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return tk, err
	}
	tk, err = arca.PedirTicket(ctx, s.cli(), s.dir(amb), []byte(cert), clave, "wsfe", ahora)
	if err != nil {
		return tk, err
	}
	plano, _ := json.Marshal(tk)
	cif, err = s.Cifra.Cifrar(contexto(tienda, "ticket"), plano)
	if err != nil {
		return tk, err
	}
	err = s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `INSERT INTO arca_tickets (tienda_id, servicio, ticket, vence) VALUES ($1, 'wsfe', $2, $3)
			ON CONFLICT (tienda_id, servicio) DO UPDATE SET ticket = $2, vence = $3`, tienda, cif, tk.Vence)
		return err
	})
	return tk, err
}

// Probar la conexión: pide (o reusa) el ticket y el último número autorizado.
func (s *Servicio) Probar(ctx context.Context, tienda string, ahora time.Time) (int, error) {
	var cuit string
	var clave []byte
	var cert *string
	var amb arca.Ambiente
	var pv int
	err := s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		var err error
		cuit, clave, cert, amb, pv, err = s.clave(ctx, t, tienda)
		return err
	})
	if err != nil {
		return 0, err
	}
	if cert == nil {
		return 0, ErrSinARCA
	}
	tk, err := s.ticket(ctx, tienda, clave, *cert, amb, ahora)
	if err != nil {
		return 0, err
	}
	c, err := s.Leer(ctx, tienda)
	if err != nil {
		return 0, err
	}
	return arca.UltimoAutorizado(ctx, s.cli(), s.dir(amb), tk, cuit, pv, tipoFactura(c.Condicion, false))
}

func tipoFactura(condicion string, receptorRI bool) int {
	switch {
	case condicion != "responsable_inscripto":
		return arca.FacturaC
	case receptorRI:
		return arca.FacturaA
	}
	return arca.FacturaB
}

func notaDe(tipo int) int {
	switch tipo {
	case arca.FacturaA:
		return arca.NotaCreditoA
	case arca.FacturaB:
		return arca.NotaCreditoB
	}
	return arca.NotaCreditoC
}

func letra(tipo int) string {
	switch tipo {
	case arca.FacturaA, arca.NotaCreditoA:
		return "A"
	case arca.FacturaB, arca.NotaCreditoB:
		return "B"
	}
	return "C"
}

func importes(tipo int, total int64) (neto, iva int64) {
	if tipo == arca.FacturaC || tipo == arca.NotaCreditoC {
		return total * 100, 0
	}
	neto = (total*10000 + 60) / 121
	return neto, total*100 - neto
}

// Manual es la factura que se carga a mano desde el panel.
type Manual struct {
	Cliente  string `json:"cliente"`
	Tipo     string `json:"tipo"`
	CUIT     string `json:"cuit"`
	Concepto string `json:"concepto"`
	Total    int64  `json:"total"`
}

// EmitirManual encola la factura y la procesa en el momento (el panel quiere saber cómo salió).
func (s *Servicio) EmitirManual(ctx context.Context, tienda string, m Manual, ahora time.Time) (string, error) {
	c, err := s.Leer(ctx, tienda)
	if err != nil {
		return "", err
	}
	if !c.Conectada {
		return "", ErrSinARCA
	}
	m.Cliente, m.Concepto = strings.TrimSpace(m.Cliente), strings.TrimSpace(m.Concepto)
	m.CUIT = strings.NewReplacer("-", "", " ", "", ".", "").Replace(m.CUIT)
	if len(m.Cliente) < 2 || len(m.Cliente) > 120 || m.Total <= 0 || m.Total > 1_000_000_000 || len(m.Concepto) > 200 {
		return "", ErrInvalido
	}
	ri := m.Tipo == "A"
	if ri && (c.Condicion != "responsable_inscripto" || !arca.CUITValido(m.CUIT)) {
		return "", ErrInvalido
	}
	tipo := tipoFactura(c.Condicion, ri)
	receptor := map[string]any{"nombre": m.Cliente, "docTipo": arca.DocConsumidorFin, "docNro": 0, "condicion": arca.CondConsumidorFinal}
	if m.CUIT != "" && arca.CUITValido(m.CUIT) {
		receptor["docTipo"], receptor["docNro"] = arca.DocCUIT, m.CUIT
		if ri {
			receptor["condicion"] = arca.CondResponsableInscripto
		}
	}
	rec, _ := json.Marshal(receptor)
	neto, iva := importes(tipo, m.Total)
	var id string
	err = s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `INSERT INTO comprobantes (tienda_id, tipo, punto_venta, estado, receptor, concepto, total, neto, iva)
			VALUES ($1, $2, $3, 'pendiente', $4, nullif($5, ''), $6, $7, $8) RETURNING id::text`,
			tienda, tipo, c.PuntoVenta, rec, m.Concepto, m.Total, neto, iva).Scan(&id)
	})
	if err != nil {
		return "", err
	}
	if err := s.ProcesarTienda(ctx, tienda, ahora); err != nil {
		return id, err
	}
	return id, nil
}

// NotaDeCredito anula (entera) una factura emitida con su nota de crédito.
func (s *Servicio) NotaDeCredito(ctx context.Context, tienda, id string, ahora time.Time) error {
	err := s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		tag, err := t.Exec(ctx, `INSERT INTO comprobantes (tienda_id, tipo, punto_venta, estado, receptor, concepto, total, neto, iva, asociado)
			SELECT tienda_id, CASE tipo WHEN 1 THEN 3 WHEN 6 THEN 8 ELSE 13 END, punto_venta, 'pendiente', receptor,
				'Anula ' || CASE tipo WHEN 1 THEN 'A' WHEN 6 THEN 'B' ELSE 'C' END || ' ' || lpad(punto_venta::text, 4, '0') || '-' || lpad(numero::text, 8, '0'),
				total, neto, iva, id
			FROM comprobantes c WHERE id::text = $1 AND estado = 'emitido' AND tipo IN (1, 6, 11)
				AND NOT EXISTS (SELECT 1 FROM comprobantes n WHERE n.asociado = c.id)`, id)
		if err == nil && tag.RowsAffected() == 0 {
			return ErrInvalido
		}
		return err
	})
	if err != nil {
		return err
	}
	return s.ProcesarTienda(ctx, tienda, ahora)
}

// Reintentar vuelve a poner en la cola los pendientes y rechazados elegidos.
func (s *Servicio) Reintentar(ctx context.Context, tienda string, ids []string, ahora time.Time) (int, error) {
	var n int64
	err := s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		tag, err := t.Exec(ctx, `UPDATE comprobantes SET estado = 'pendiente', intentos = 0, reintentar_en = now(), motivo = NULL, codigo_arca = NULL
			WHERE id::text = ANY($1) AND estado IN ('pendiente', 'rechazado')`, ids)
		n = tag.RowsAffected()
		return err
	})
	if err != nil || n == 0 {
		return int(n), err
	}
	return int(n), s.ProcesarTienda(ctx, tienda, ahora)
}

// ParaListas adapta el servicio a las acciones de la lista de Facturación, con errores que se
// pueden mostrar tal cual.
type ParaListas struct{ S *Servicio }

func legible(err error) error {
	var r arca.Rechazado
	var e arca.ErrARCA
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ErrSinARCA):
		return errors.New("Primero conectá ARCA con tu certificado (Integraciones › ARCA).")
	case errors.Is(err, ErrInvalido):
		return errors.New("Revisá los datos: la factura A necesita un CUIT válido y ser responsable inscripto.")
	case errors.As(err, &r):
		return errors.New("ARCA la rechazó: " + r.Mensaje)
	case errors.As(err, &e):
		return errors.New("ARCA no respondió bien: " + e.Mensaje)
	}
	return errors.New("No pudimos emitirla. Probá de nuevo en un rato.")
}

func (p ParaListas) EmitirManual(ctx context.Context, tienda string, valor json.RawMessage, ahora time.Time) (int, error) {
	var m Manual
	if json.Unmarshal(valor, &m) != nil {
		return 0, legible(ErrInvalido)
	}
	id, err := p.S.EmitirManual(ctx, tienda, m, ahora)
	if err != nil {
		return 0, legible(err)
	}
	// Se encoló; si ARCA la rechazó en el momento, lo dice.
	var estado, motivo string
	_ = p.S.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT estado, coalesce(motivo, '') FROM comprobantes WHERE id::text = $1`, id).Scan(&estado, &motivo)
	})
	if estado == "rechazado" {
		return 0, errors.New("ARCA la rechazó: " + motivo)
	}
	return 1, nil
}

func (p ParaListas) Reintentar(ctx context.Context, tienda string, ids []string, ahora time.Time) (int, error) {
	n, err := p.S.Reintentar(ctx, tienda, ids, ahora)
	return n, legible(err)
}

func (p ParaListas) Anular(ctx context.Context, tienda string, ids []string, ahora time.Time) (int, error) {
	n := 0
	for _, id := range ids {
		if err := p.S.NotaDeCredito(ctx, tienda, id, ahora); err != nil {
			if errors.Is(err, ErrInvalido) {
				return n, errors.New("Esa factura no se puede anular: tiene que estar autorizada y sin otra nota de crédito.")
			}
			return n, legible(err)
		}
		n++
	}
	return n, nil
}
