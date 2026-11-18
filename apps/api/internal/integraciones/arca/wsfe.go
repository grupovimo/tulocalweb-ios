package arca

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Tipos de comprobante (tabla de ARCA): una factura A es el 1 y una B es el 6; confundirlos emite
// otro tipo del que dice la pantalla.
const (
	FacturaA     = 1
	NotaCreditoA = 3
	FacturaB     = 6
	NotaCreditoB = 8
	FacturaC     = 11
	NotaCreditoC = 13
)

// Documento de quien recibe la factura.
const (
	DocCUIT          = 80
	DocDNI           = 96
	DocConsumidorFin = 99
)

// Condición frente al IVA de quien recibe (obligatoria desde 2025).
const (
	CondResponsableInscripto = 1
	CondExento               = 4
	CondConsumidorFinal      = 5
	CondMonotributo          = 6
)

// Comprobante a autorizar. Los precios de una tienda argentina son finales (con IVA), así que el
// neto se saca dividiendo y el redondeo cae en el IVA: la suma cierra exacta.
type Comprobante struct {
	Tipo         int
	PuntoVenta   int
	DocTipo      int
	DocNro       int64
	CondReceptor int
	Fecha        time.Time
	// Total en pesos (los precios de Tu Local Web son enteros); se manda con dos decimales.
	Total int64
	// Para notas de crédito: el comprobante que corrige.
	Asociado *Asociado
}

type Asociado struct{ Tipo, PuntoVenta, Numero int }

// Importes en centavos: neto + iva = total.
func (c Comprobante) importes() (neto, iva, total int64) {
	total = c.Total * 100
	if c.Tipo == FacturaC || c.Tipo == NotaCreditoC {
		return total, 0, total
	}
	neto = (total*100 + 60) / 121
	return neto, total - neto, total
}

func dec(centavos int64) string { return fmt.Sprintf("%d.%02d", centavos/100, centavos%100) }

func esA(t int) bool { return t == FacturaA || t == NotaCreditoA }

// Validar lo que ARCA rechazaría de entrada.
func (c Comprobante) Validar() error {
	switch {
	case c.Total <= 0:
		return errors.New("arca: el total tiene que ser mayor a cero")
	case c.PuntoVenta < 1 || c.PuntoVenta > 99998:
		return errors.New("arca: punto de venta inválido")
	case esA(c.Tipo) && (c.DocTipo != DocCUIT || !CUITValido(fmt.Sprint(c.DocNro))):
		return errors.New("arca: la factura A necesita el CUIT de quien recibe")
	case (c.Tipo == NotaCreditoA || c.Tipo == NotaCreditoB || c.Tipo == NotaCreditoC) && c.Asociado == nil:
		return errors.New("arca: la nota de crédito necesita el comprobante que corrige")
	}
	return nil
}

func auth(t Ticket, cuit string) string {
	return fmt.Sprintf(`<ar:Auth><ar:Token>%s</ar:Token><ar:Sign>%s</ar:Sign><ar:Cuit>%s</ar:Cuit></ar:Auth>`,
		escapar(t.Token), escapar(t.Sign), cuit)
}

func escapar(s string) string {
	var b strings.Builder
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}

func sobre(cuerpo string) []byte {
	return []byte(`<?xml version="1.0" encoding="UTF-8"?><soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:ar="http://ar.gov.afip.dif.FEV1/"><soapenv:Header/><soapenv:Body>` +
		cuerpo + `</soapenv:Body></soapenv:Envelope>`)
}

type errsARCA struct {
	Err []struct {
		Code int    `xml:"Code"`
		Msg  string `xml:"Msg"`
	} `xml:"Err"`
}

func (e errsARCA) primero() error {
	if len(e.Err) == 0 {
		return nil
	}
	return ErrARCA{Codigo: fmt.Sprint(e.Err[0].Code), Mensaje: e.Err[0].Msg}
}

// UltimoAutorizado: el último número autorizado de ese tipo en ese punto de venta.
func UltimoAutorizado(ctx context.Context, cli *http.Client, dir Direcciones, t Ticket, cuit string, pv, tipo int) (int, error) {
	b, err := soap(ctx, cli, dir.WSFE, "http://ar.gov.afip.dif.FEV1/FECompUltimoAutorizado", sobre(fmt.Sprintf(
		`<ar:FECompUltimoAutorizado>%s<ar:PtoVta>%d</ar:PtoVta><ar:CbteTipo>%d</ar:CbteTipo></ar:FECompUltimoAutorizado>`,
		auth(t, cuit), pv, tipo)))
	if err != nil {
		return 0, err
	}
	var r struct {
		Nro    int      `xml:"Body>FECompUltimoAutorizadoResponse>FECompUltimoAutorizadoResult>CbteNro"`
		Errors errsARCA `xml:"Body>FECompUltimoAutorizadoResponse>FECompUltimoAutorizadoResult>Errors"`
	}
	if err := xml.Unmarshal(b, &r); err != nil {
		return 0, err
	}
	if err := r.Errors.primero(); err != nil {
		return 0, err
	}
	return r.Nro, nil
}

// Autorizado es la respuesta aprobada: el CAE y su vencimiento.
type Autorizado struct {
	Numero        int
	CAE           string
	VenceCAE      time.Time
	Observaciones []string
}

// Rechazado: ARCA contestó (HTTP 200) pero no autorizó. No es un error de quien llamó: es el
// estado del comprobante, con el código de ARCA, que es lo único accionable (10016 ≠ 10048).
type Rechazado struct {
	Codigo  string
	Mensaje string
}

func (r Rechazado) Error() string { return "arca: rechazado " + r.Codigo + " " + r.Mensaje }

// SolicitarCAE pide la autorización del comprobante con el número indicado (el siguiente al último).
func SolicitarCAE(ctx context.Context, cli *http.Client, dir Direcciones, t Ticket, cuit string, c Comprobante, numero int) (Autorizado, error) {
	if err := c.Validar(); err != nil {
		return Autorizado{}, err
	}
	neto, iva, total := c.importes()
	var ivas, asoc string
	if iva > 0 {
		ivas = fmt.Sprintf(`<ar:Iva><ar:AlicIva><ar:Id>5</ar:Id><ar:BaseImp>%s</ar:BaseImp><ar:Importe>%s</ar:Importe></ar:AlicIva></ar:Iva>`, dec(neto), dec(iva))
	}
	if c.Asociado != nil {
		asoc = fmt.Sprintf(`<ar:CbtesAsoc><ar:CbteAsoc><ar:Tipo>%d</ar:Tipo><ar:PtoVta>%d</ar:PtoVta><ar:Nro>%d</ar:Nro><ar:Cuit>%s</ar:Cuit></ar:CbteAsoc></ar:CbtesAsoc>`,
			c.Asociado.Tipo, c.Asociado.PuntoVenta, c.Asociado.Numero, cuit)
	}
	det := fmt.Sprintf(`<ar:FECAEDetRequest><ar:Concepto>1</ar:Concepto><ar:DocTipo>%d</ar:DocTipo><ar:DocNro>%d</ar:DocNro><ar:CbteDesde>%d</ar:CbteDesde><ar:CbteHasta>%d</ar:CbteHasta><ar:CbteFch>%s</ar:CbteFch><ar:ImpTotal>%s</ar:ImpTotal><ar:ImpTotConc>0.00</ar:ImpTotConc><ar:ImpNeto>%s</ar:ImpNeto><ar:ImpOpEx>0.00</ar:ImpOpEx><ar:ImpTrib>0.00</ar:ImpTrib><ar:ImpIVA>%s</ar:ImpIVA><ar:MonId>PES</ar:MonId><ar:MonCotiz>1</ar:MonCotiz><ar:CondicionIVAReceptorId>%d</ar:CondicionIVAReceptorId>%s%s</ar:FECAEDetRequest>`,
		c.DocTipo, c.DocNro, numero, numero, c.Fecha.In(arg).Format("20060102"), dec(total), dec(neto), dec(iva), c.CondReceptor, asoc, ivas)
	b, err := soap(ctx, cli, dir.WSFE, "http://ar.gov.afip.dif.FEV1/FECAESolicitar", sobre(fmt.Sprintf(
		`<ar:FECAESolicitar>%s<ar:FeCAEReq><ar:FeCabReq><ar:CantReg>1</ar:CantReg><ar:PtoVta>%d</ar:PtoVta><ar:CbteTipo>%d</ar:CbteTipo></ar:FeCabReq><ar:FeDetReq>%s</ar:FeDetReq></ar:FeCAEReq></ar:FECAESolicitar>`,
		auth(t, cuit), c.PuntoVenta, c.Tipo, det)))
	if err != nil {
		return Autorizado{}, err
	}
	var r struct {
		Resultado string   `xml:"Body>FECAESolicitarResponse>FECAESolicitarResult>FeCabResp>Resultado"`
		Errors    errsARCA `xml:"Body>FECAESolicitarResponse>FECAESolicitarResult>Errors"`
		Det       struct {
			Resultado string `xml:"Resultado"`
			CAE       string `xml:"CAE"`
			Vence     string `xml:"CAEFchVto"`
			Desde     int    `xml:"CbteDesde"`
			Obs       []struct {
				Code int    `xml:"Code"`
				Msg  string `xml:"Msg"`
			} `xml:"Observaciones>Obs"`
		} `xml:"Body>FECAESolicitarResponse>FECAESolicitarResult>FeDetResp>FECAEDetResponse"`
	}
	if err := xml.Unmarshal(b, &r); err != nil {
		return Autorizado{}, err
	}
	// Un 200 no es un comprobante autorizado: manda el Resultado (A, R o P).
	if r.Det.Resultado != "A" || r.Det.CAE == "" {
		if e := r.Errors.primero(); e != nil {
			var ea ErrARCA
			errors.As(e, &ea)
			return Autorizado{}, Rechazado{Codigo: ea.Codigo, Mensaje: ea.Mensaje}
		}
		if len(r.Det.Obs) > 0 {
			return Autorizado{}, Rechazado{Codigo: fmt.Sprint(r.Det.Obs[0].Code), Mensaje: r.Det.Obs[0].Msg}
		}
		return Autorizado{}, Rechazado{Codigo: "R", Mensaje: "ARCA no autorizó el comprobante."}
	}
	vence, _ := time.ParseInLocation("20060102", r.Det.Vence, arg)
	a := Autorizado{Numero: r.Det.Desde, CAE: r.Det.CAE, VenceCAE: vence}
	if a.Numero == 0 {
		a.Numero = numero
	}
	for _, o := range r.Det.Obs {
		a.Observaciones = append(a.Observaciones, fmt.Sprintf("%d %s", o.Code, o.Msg))
	}
	return a, nil
}

// Consultado es lo que ARCA tiene de un comprobante ya emitido.
type Consultado struct {
	CAE      string
	VenceCAE time.Time
	Total    string
	// DocNro: el documento del receptor (0 si es consumidor final sin identificar).
	DocNro int64
}

// Consultar trae un comprobante por tipo, punto de venta y número (FECompConsultar). Sirve para
// recuperar el CAE cuando ARCA autorizó pero la respuesta no llegó: reintentar a ciegas emitiría
// otro comprobante por la misma venta.
func Consultar(ctx context.Context, cli *http.Client, dir Direcciones, t Ticket, cuit string, pv, tipo, numero int) (Consultado, error) {
	b, err := soap(ctx, cli, dir.WSFE, "http://ar.gov.afip.dif.FEV1/FECompConsultar", sobre(fmt.Sprintf(
		`<ar:FECompConsultar>%s<ar:FeCompConsReq><ar:CbteTipo>%d</ar:CbteTipo><ar:CbteNro>%d</ar:CbteNro><ar:PtoVta>%d</ar:PtoVta></ar:FeCompConsReq></ar:FECompConsultar>`,
		auth(t, cuit), tipo, numero, pv)))
	if err != nil {
		return Consultado{}, err
	}
	var r struct {
		CAE    string   `xml:"Body>FECompConsultarResponse>FECompConsultarResult>ResultGet>CodAutorizacion"`
		Vence  string   `xml:"Body>FECompConsultarResponse>FECompConsultarResult>ResultGet>FchVto"`
		Total  string   `xml:"Body>FECompConsultarResponse>FECompConsultarResult>ResultGet>ImpTotal"`
		DocNro int64    `xml:"Body>FECompConsultarResponse>FECompConsultarResult>ResultGet>DocNro"`
		Errors errsARCA `xml:"Body>FECompConsultarResponse>FECompConsultarResult>Errors"`
	}
	if err := xml.Unmarshal(b, &r); err != nil {
		return Consultado{}, err
	}
	if err := r.Errors.primero(); err != nil {
		return Consultado{}, err
	}
	v, _ := time.ParseInLocation("20060102", r.Vence, arg)
	return Consultado{CAE: r.CAE, VenceCAE: v, Total: r.Total, DocNro: r.DocNro}, nil
}

// TotalDec es el total como lo manda (y lo devuelve) ARCA: "12100.00".
func (c Comprobante) TotalDec() string { _, _, t := c.importes(); return dec(t) }
