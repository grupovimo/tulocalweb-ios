package facturacion_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/facturacion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/arca"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/arca/arcatest"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/cifrado"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

// firmarComoARCA hace de ARCA firmando el pedido de certificado con una CA de homologación.
func firmarComoARCA(t *testing.T, csrPEM []byte) string {
	t.Helper()
	b, _ := pem.Decode(csrPEM)
	csr, err := x509.ParseCertificateRequest(b.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	caK, _ := rsa.GenerateKey(rand.Reader, 2048)
	ca := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "Computadores Test"}, IsCA: true,
		BasicConstraintsValid: true, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(48 * time.Hour)}
	der, err := x509.CreateCertificate(rand.Reader, &x509.Certificate{SerialNumber: big.NewInt(99), Subject: csr.Subject,
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(24 * time.Hour)}, ca, csr.PublicKey, caK)
	if err != nil {
		t.Fatal(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
}

func cifrador(t *testing.T) *cifrado.Cifrador {
	k := make([]byte, 32)
	_, _ = rand.Read(k)
	c, err := cifrado.Nuevo(base64.StdEncoding.EncodeToString(k))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

type fila struct {
	estado, motivo string
	numero         *int
	cae            *string
}

func comprobante(t *testing.T, b *pruebadb.Bases, tienda, donde string, args ...any) fila {
	t.Helper()
	var f fila
	err := b.App.EnTienda(context.Background(), tienda, func(tx pgx.Tx) error {
		return tx.QueryRow(context.Background(), `SELECT estado, coalesce(motivo, ''), numero, cae FROM comprobantes WHERE `+donde+` ORDER BY creado_en DESC LIMIT 1`, args...).
			Scan(&f.estado, &f.motivo, &f.numero, &f.cae)
	})
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func TestFacturacionConARCA(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	sol := pruebadb.CrearTienda(t, b.App, "sol")
	falso := arcatest.Nuevo()
	defer falso.Close()
	s := &facturacion.Servicio{Pool: b.App, Cifra: cifrador(t), HTTP: falso.Client(),
		Direcciones: func(arca.Ambiente) arca.Direcciones { return arca.Direcciones{WSAA: falso.WSAA(), WSFE: falso.WSFE()} }}
	// El reloj real (un minuto adelante): la cola encola con now() de Postgres; con una fecha fija, lo
	// encolado quedaba en el futuro del test y no se drenaba.
	ahora := time.Now().UTC().Add(time.Minute)
	stock := 100
	prod, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Mate", Precio: 1000, Stock: &stock})
	if err != nil {
		t.Fatal(err)
	}
	// Una venta cobrada por Mercado Pago, por el camino real (entra "Aprobado").
	vender := func(nombre string, total int64) int {
		t.Helper()
		n := pedidos.PedidoNuevo{Canal: "web", Pago: "mp", Envio: "retiro", Cliente: &nombre, Precios: map[string]int64{prod.ID: total}}
		n.Items = append(n.Items, pedidos.Linea{ProductoID: prod.ID, Cantidad: 1})
		numero, _, err := pedidos.CrearDesdeProductos(ctx, b.App, luna, n)
		if err != nil {
			t.Fatal(err)
		}
		return numero
	}

	// Sin ARCA: la venta cobrada queda "no facturada" con su motivo, nunca sin rastro.
	p0 := vender("Ana", 5000)
	if f := comprobante(t, b, luna, "pedido_numero = $1", p0); f.estado != "no_facturado" || !strings.Contains(f.motivo, "no conectó ARCA") {
		t.Fatalf("sin ARCA: %+v", f)
	}

	// Conectar: sólo dueña o admin; clave y pedido de certificado; ARCA lo firma; se guarda.
	alta := facturacion.Alta{CUIT: "20-44859234-1", RazonSocial: "Nahuel Santillán", Condicion: "monotributo", PuntoVenta: 3, Ambiente: "homologacion"}
	if _, err := s.GenerarPedido(ctx, luna, "ventas", alta); err != facturacion.ErrSinPermiso {
		t.Fatalf("ventas generó la clave: %v", err)
	}
	csr, err := s.GenerarPedido(ctx, luna, "duenio", alta)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.GuardarCertificado(ctx, luna, "duenio", "no es un certificado", ahora); err == nil {
		t.Fatal("aceptó cualquier cosa")
	}
	if _, err := s.GuardarCertificado(ctx, luna, "duenio", firmarComoARCA(t, csr), ahora); err != nil {
		t.Fatal(err)
	}
	if c, _ := s.Leer(ctx, luna); !c.Conectada || c.CUIT != "20448592341" || c.PuntoVenta != 3 {
		t.Fatalf("config: %+v", c)
	}
	if n, err := s.Probar(ctx, luna, ahora); err != nil || n != 0 {
		t.Fatalf("probar: %d %v", n, err)
	}

	// Venta cobrada → se encola en la misma transacción → el drenador pide el CAE (factura C).
	p1 := vender("Lucía", 12100)
	if f := comprobante(t, b, luna, "pedido_numero = $1", p1); f.estado != "pendiente" {
		t.Fatalf("encolada: %+v", f)
	}
	if err := s.Drenar(ctx, ahora); err != nil {
		t.Fatal(err)
	}
	f1 := comprobante(t, b, luna, "pedido_numero = $1", p1)
	if f1.estado != "emitido" || f1.numero == nil || *f1.numero != 1 || f1.cae == nil {
		t.Fatalf("emitida: %+v", f1)
	}
	if !strings.Contains(falso.UltimoPedido, "<ar:CbteTipo>11</ar:CbteTipo>") || !strings.Contains(falso.UltimoPedido, "<ar:ImpTotal>12100.00</ar:ImpTotal>") {
		t.Fatalf("lo que salió: %s", falso.UltimoPedido)
	}
	var factura string
	_ = b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT coalesce(factura, '') FROM pedidos WHERE numero = $1`, p1).Scan(&factura)
	})
	if factura != "C 0003-00000001" {
		t.Fatalf("pedido: %q", factura)
	}

	// Factura manual; la A no la puede emitir un monotributista.
	if _, err := s.EmitirManual(ctx, luna, facturacion.Manual{Cliente: "Ferretería", Tipo: "A", CUIT: "30500010912", Total: 1000}, ahora); err != facturacion.ErrInvalido {
		t.Fatalf("A de monotributo: %v", err)
	}
	if _, err := s.EmitirManual(ctx, luna, facturacion.Manual{Cliente: "Consumidor final", Concepto: "Clases", Total: 8000}, ahora); err != nil {
		t.Fatal(err)
	}
	if f := comprobante(t, b, luna, "concepto = 'Clases'"); f.estado != "emitido" || *f.numero != 2 {
		t.Fatalf("manual: %+v", f)
	}

	// La respuesta de ARCA se corta después de autorizar: el reintento recupera el CAE con
	// FECompConsultar y NO emite otro comprobante por la misma venta.
	falso.CortarDespuesDeAutorizar = true
	_, _ = s.EmitirManual(ctx, luna, facturacion.Manual{Cliente: "Corte", Total: 3000}, ahora)
	if f := comprobante(t, b, luna, "receptor->>'nombre' = 'Corte'"); f.estado != "pendiente" {
		t.Fatalf("corte: %+v", f)
	}
	var id string
	_ = b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT id::text FROM comprobantes WHERE receptor->>'nombre' = 'Corte'`).Scan(&id)
	})
	if _, err := s.Reintentar(ctx, luna, []string{id}, ahora); err != nil {
		t.Fatal(err)
	}
	if f := comprobante(t, b, luna, "receptor->>'nombre' = 'Corte'"); f.estado != "emitido" || *f.numero != 3 {
		t.Fatalf("recuperada: %+v", f)
	}
	if falso.Ultimo[arca.FacturaC] != 3 {
		t.Fatalf("se emitió de más: último %d", falso.Ultimo[arca.FacturaC])
	}

	// Nota de crédito: anula la factura del pedido (tipo 13, asociada) y la lista la muestra anulada.
	var idP1 string
	_ = b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT id::text FROM comprobantes WHERE pedido_numero = $1`, p1).Scan(&idP1)
	})
	if err := s.NotaDeCredito(ctx, luna, idP1, ahora); err != nil {
		t.Fatal(err)
	}
	if err := s.NotaDeCredito(ctx, luna, idP1, ahora); err == nil {
		t.Fatal("se anuló dos veces")
	}
	// Desde la lista, el error se lee tal cual en el panel (no el de la factura A).
	if _, err := (facturacion.ParaListas{S: s}).Anular(ctx, luna, []string{idP1}, ahora); err == nil || !strings.Contains(err.Error(), "no se puede anular") {
		t.Fatalf("anular dos veces desde la lista: %v", err)
	}
	regs, err := listas.Comprobantes(ctx, b.App, luna, ahora)
	if err != nil {
		t.Fatal(err)
	}
	estados := map[string]string{}
	for _, r := range regs {
		estados[r["numero"].(string)] = r["estado"].(string)
	}
	if estados["C 0003-00000001"] != "void" || estados["NC C 0003-00000001"] != "ok" || estados["—"] != "nofac" {
		t.Fatalf("lista: %v", estados)
	}

	// Aislamiento: otra tienda no ve nada.
	if r, _ := listas.Comprobantes(ctx, b.App, sol, ahora); len(r) != 0 {
		t.Fatalf("otra tienda ve %d comprobantes", len(r))
	}
}

// Responsable inscripta: la venta online con CUIT sale como factura A a ese CUIT; sin CUIT, B a
// consumidor final.
func TestFacturaADesdeElCheckout(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	falso := arcatest.Nuevo()
	defer falso.Close()
	s := &facturacion.Servicio{Pool: b.App, Cifra: cifrador(t), HTTP: falso.Client(),
		Direcciones: func(arca.Ambiente) arca.Direcciones { return arca.Direcciones{WSAA: falso.WSAA(), WSFE: falso.WSFE()} }}
	ahora := time.Now().UTC().Add(time.Minute)
	alta := facturacion.Alta{CUIT: "20-44859234-1", RazonSocial: "Luna SRL", Condicion: "responsable_inscripto", PuntoVenta: 2, Ambiente: "homologacion"}
	csr, err := s.GenerarPedido(ctx, luna, "duenio", alta)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.GuardarCertificado(ctx, luna, "duenio", firmarComoARCA(t, csr), ahora); err != nil {
		t.Fatal(err)
	}
	stock := 100
	prod, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Mate", Precio: 12100, Stock: &stock})
	if err != nil {
		t.Fatal(err)
	}
	vender := func(nombre string, f *pedidos.Factura) {
		t.Helper()
		n := pedidos.PedidoNuevo{Canal: "web", Pago: "mp", Envio: "retiro", Cliente: &nombre, Factura: f, Precios: map[string]int64{prod.ID: 12100}}
		n.Items = append(n.Items, pedidos.Linea{ProductoID: prod.ID, Cantidad: 1})
		if _, _, err := pedidos.CrearDesdeProductos(ctx, b.App, luna, n); err != nil {
			t.Fatal(err)
		}
		if err := s.Drenar(ctx, ahora); err != nil {
			t.Fatal(err)
		}
	}

	vender("Ana", &pedidos.Factura{CUIT: "30500010912", RazonSocial: "Ferretería El Tornillo SA"})
	for _, quiere := range []string{"<ar:CbteTipo>1</ar:CbteTipo>", "<ar:DocTipo>80</ar:DocTipo>", "<ar:DocNro>30500010912</ar:DocNro>",
		"<ar:CondicionIVAReceptorId>1</ar:CondicionIVAReceptorId>", "<ar:ImpNeto>10000.00</ar:ImpNeto>"} {
		if !strings.Contains(falso.UltimoPedido, quiere) {
			t.Fatalf("factura A sin %s: %s", quiere, falso.UltimoPedido)
		}
	}
	if f := comprobante(t, b, luna, "receptor->>'nombre' = 'Ferretería El Tornillo SA'"); f.estado != "emitido" {
		t.Fatalf("A: %+v", f)
	}

	vender("Beto", nil)
	if !strings.Contains(falso.UltimoPedido, "<ar:CbteTipo>6</ar:CbteTipo>") || !strings.Contains(falso.UltimoPedido, "<ar:DocTipo>99</ar:DocTipo>") {
		t.Fatalf("B a consumidor final: %s", falso.UltimoPedido)
	}
}

// El certificado por vencer avisa una vez por umbral (30, 7, al vencer); uno nuevo arranca de cero.
func TestAvisoDeVencimientoDelCertificado(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	sol := pruebadb.CrearTienda(t, b.App, "sol")
	s := &facturacion.Servicio{Pool: b.App, Cifra: cifrador(t)}
	ahora := time.Now().UTC()
	conectar := func() {
		t.Helper()
		alta := facturacion.Alta{CUIT: "20-44859234-1", RazonSocial: "Luna", Condicion: "monotributo", PuntoVenta: 2, Ambiente: "homologacion"}
		csr, err := s.GenerarPedido(ctx, luna, "duenio", alta)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := s.GuardarCertificado(ctx, luna, "duenio", firmarComoARCA(t, csr), ahora); err != nil {
			t.Fatal(err)
		}
	}
	vencerEn := func(d time.Duration) {
		t.Helper()
		if err := b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `UPDATE arca_config SET certificado_vence = $1`, ahora.Add(d))
			return err
		}); err != nil {
			t.Fatal(err)
		}
	}
	avisos := func() []int {
		t.Helper()
		if err := s.AvisarVencimientos(ctx, ahora); err != nil {
			t.Fatal(err)
		}
		filas, err := b.Worker.Query(ctx, `SELECT (datos->>'dias')::int FROM outbox WHERE tipo = $1 ORDER BY id`, facturacion.TipoVence)
		if err != nil {
			t.Fatal(err)
		}
		ds, err := pgx.CollectRows(filas, pgx.RowTo[int])
		if err != nil {
			t.Fatal(err)
		}
		return ds
	}
	conectar()
	vencerEn(90 * 24 * time.Hour)
	if d := avisos(); len(d) != 0 {
		t.Fatalf("a 90 días avisó: %v", d)
	}
	vencerEn(20 * 24 * time.Hour)
	avisos()
	if d := avisos(); len(d) != 1 || d[0] != 20 {
		t.Fatalf("a 20 días (una sola vez): %v", d)
	}
	vencerEn(5*24*time.Hour - time.Hour)
	if d := avisos(); len(d) != 2 || d[1] != 5 {
		t.Fatalf("a 5 días: %v", d)
	}
	vencerEn(-time.Hour)
	avisos()
	if d := avisos(); len(d) != 3 || d[2] != 0 {
		t.Fatalf("vencido: %v", d)
	}
	// Certificado nuevo: vuelve a avisar cuando le toque.
	conectar()
	vencerEn(10 * 24 * time.Hour)
	if d := avisos(); len(d) != 4 || d[3] != 10 {
		t.Fatalf("con el certificado nuevo: %v", d)
	}
	// La otra tienda, sin ARCA, nunca.
	var otra int
	_ = b.Worker.QueryRow(ctx, `SELECT count(*) FROM outbox WHERE tipo = $1 AND tienda_id = $2`, facturacion.TipoVence, sol).Scan(&otra)
	if otra != 0 {
		t.Fatalf("avisó a una tienda sin ARCA: %d", otra)
	}
}
