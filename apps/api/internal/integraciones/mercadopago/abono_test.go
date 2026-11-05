package mercadopago_test

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/abono"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/mercadopago"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/cifrado"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

// El abono con prorrateo (§56) de punta a punta: alta, subir pagando la diferencia, bajar y volver
// atrás sin cobrar, pasar a gratis sin perder lo pagado, el cambio programado y los cobros del débito.
func TestAbonoConProrrateo(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	var ana string
	if err := b.App.EnTx(ctx, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `INSERT INTO auth.usuarios (email, nombre) VALUES ('ana+luna@luna.com', 'Ana') RETURNING id::text`).Scan(&ana)
	}); err != nil {
		t.Fatal(err)
	}
	f := &mpFalso{cuerpos: map[string]map[string]any{}, pagos: map[string]map[string]any{}, susc: map[string]map[string]any{},
		autoriz: map[string]map[string]any{}, usuario: 777, refreshOK: true}
	srv := httptest.NewServer(f)
	defer srv.Close()
	k := make([]byte, 32)
	_, _ = rand.Read(k)
	cif, _ := cifrado.Nuevo(base64.StdEncoding.EncodeToString(k))
	// La prueba gratis ya terminó.
	ahora := time.Now().Add(20 * 24 * time.Hour)
	s := &mercadopago.Servicio{Pool: b.App, Cifra: cif, Cliente: &mercadopago.Cliente{Base: srv.URL},
		Ahora:  func() time.Time { return ahora },
		Config: mercadopago.Config{AppID: "123", Secreto: "sec", TokenPlataforma: "APP_USR-plataforma", SecretoWebhook: "whsec", Sitio: "https://tulocalweb.com.ar"}}
	leer := func() (plan string, prog *string, proximo *time.Time) {
		_ = b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
			return tx.QueryRow(ctx, `SELECT plan, plan_programado, abono_proximo FROM tiendas WHERE id = $1`, luna).Scan(&plan, &prog, &proximo)
		})
		return
	}
	cobros := func() int {
		var n int
		_ = b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
			return tx.QueryRow(ctx, `SELECT count(*) FROM abono_cobros WHERE tienda_id = $1`, luna).Scan(&n)
		})
		return n
	}

	// Alta sin prueba: el débito arranca ya, sin free_trial.
	u, err := s.CambiarAbono(ctx, luna, ana, "emprende", "mes")
	if err != nil || !strings.Contains(u, "preapproval_id=sus-2") {
		t.Fatalf("alta: %q %v", u, err)
	}
	rec := f.cuerpos["POST /preapproval"]["auto_recurring"].(map[string]any)
	if _, hay := rec["free_trial"]; hay || rec["transaction_amount"].(float64) != 19900 {
		t.Fatalf("débito: %v", rec)
	}
	proximo := ahora.AddDate(0, 1, 0)
	f.susc["sus-2"] = map[string]any{"id": "sus-2", "status": "authorized", "external_reference": luna,
		"reason": "Tu Local Web Emprende (mensual)", "next_payment_date": proximo.Format("2006-01-02T15:04:05.000-07:00")}
	if err := s.VolverAbono(ctx, luna, "", "sus-2"); err != nil {
		t.Fatal(err)
	}
	if p, _, prox := leer(); p != "emprende" || prox == nil || prox.Sub(proximo).Abs() > time.Second {
		t.Fatalf("después de autorizar: %s %v", p, prox)
	}

	// Sube a los 10 días: la diferencia por lo que queda, con un pago único de la plataforma.
	ahora = ahora.Add(10 * 24 * time.Hour)
	u, err = s.CambiarAbono(ctx, luna, ana, "pro", "mes")
	if err != nil || !strings.Contains(u, "pref_id=pref-1") {
		t.Fatalf("subir: %q %v", u, err)
	}
	var e abono.Estado
	_ = b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error { e, err = abono.Leer(ctx, tx, luna); return err })
	c, _ := abono.Cotizar(e, "pro", "mes", ahora)
	pref := f.cuerpos["POST /checkout/preferences"]
	item := pref["items"].([]any)[0].(map[string]any)
	if c.Tipo != "sube" || c.Hoy <= 0 || item["unit_price"].(float64) != float64(c.Hoy) {
		t.Fatalf("diferencia: %+v %v", c, item)
	}
	if p, _, _ := leer(); p != "emprende" {
		t.Fatal("subió antes de pagar")
	}
	// Un pago por menos no aplica nada; el bueno, sí (y una sola vez).
	f.pagos["900"] = map[string]any{"id": 900, "status": "approved", "external_reference": pref["external_reference"], "transaction_amount": c.Hoy - 1}
	_ = s.ConciliarPagoAbono(ctx, "900")
	if p, _, _ := leer(); p != "emprende" {
		t.Fatal("aplicó un pago incompleto")
	}
	f.pagos["901"] = map[string]any{"id": 901, "status": "approved", "external_reference": pref["external_reference"], "transaction_amount": c.Hoy}
	if err := s.VolverAbono(ctx, luna, "901", ""); err != nil {
		t.Fatal(err)
	}
	_ = s.ConciliarPagoAbono(ctx, "901")
	if p, _, _ := leer(); p != "pro" || cobros() != 1 {
		t.Fatalf("después de pagar la diferencia: %s, %d cobros", p, cobros())
	}
	if f.susc["sus-2"]["transaction_amount"].(float64) != 49900 {
		t.Fatalf("el débito no subió: %v", f.susc["sus-2"])
	}
	// La vuelta de otra tienda no concilia el pago ajeno.
	sol := pruebadb.CrearTienda(t, b.App, "sol")
	if err := s.VolverAbono(ctx, sol, "901", ""); err == nil {
		t.Fatal("otra tienda concilió el pago")
	}

	// Baja: nada hoy, sigue con Pro hasta el próximo cobro y el débito baja de monto.
	if u, err := s.CambiarAbono(ctx, luna, ana, "emprende", "mes"); err != nil || u != "" {
		t.Fatalf("bajar: %q %v", u, err)
	}
	if p, prog, _ := leer(); p != "pro" || prog == nil || *prog != "emprende" || f.susc["sus-2"]["transaction_amount"].(float64) != 19900 {
		t.Fatalf("baja programada: %s %v %v", p, prog, f.susc["sus-2"])
	}
	// Con la baja programada, otro plan no; volver al vigente la deshace.
	if _, err := s.CambiarAbono(ctx, luna, ana, "negocio", "mes"); err == nil {
		t.Fatal("dejó elegir otro plan con un cambio programado")
	}
	if u, err := s.CambiarAbono(ctx, luna, ana, "pro", "mes"); err != nil || u != "" {
		t.Fatalf("deshacer: %q %v", u, err)
	}
	if _, prog, _ := leer(); prog != nil || f.susc["sus-2"]["transaction_amount"].(float64) != 49900 {
		t.Fatalf("no se deshizo: %v", prog)
	}

	// A gratis: el débito se da de baja, lo pagado sigue. El aviso de la baja no lo pisa.
	if u, err := s.CambiarAbono(ctx, luna, ana, "gratis", "mes"); err != nil || u != "" {
		t.Fatalf("gratis: %q %v", u, err)
	}
	if p, prog, _ := leer(); p != "pro" || prog == nil || *prog != "gratis" || f.susc["sus-2"]["status"] != "cancelled" {
		t.Fatalf("gratis programado: %s %v %v", p, prog, f.susc["sus-2"])
	}
	_ = s.ConciliarSuscripcion(ctx, "sus-2")
	if p, _, _ := leer(); p != "pro" {
		t.Fatal("el aviso de la baja sacó lo pagado")
	}

	// Llega la fecha: el worker aplica el cambio programado.
	_ = b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE tiendas SET programado_desde = now() - interval '1 minute' WHERE id = $1`, luna)
		return err
	})
	if n, err := abono.AplicarProgramados(ctx, b.App); err != nil || n != 1 {
		t.Fatalf("aplicar programados: %d %v", n, err)
	}
	if p, prog, prox := leer(); p != "gratis" || prog != nil || prox != nil {
		t.Fatalf("después de la fecha: %s %v %v", p, prog, prox)
	}

	// Cada cobro del débito queda en Facturas, una vez.
	f.autoriz["77"] = map[string]any{"id": 77, "preapproval_id": "sus-2", "status": "processed", "transaction_amount": 49900,
		"payment": map[string]any{"id": 1, "status": "approved"}}
	for range 2 {
		if err := s.ConciliarCobroAbono(ctx, "77"); err != nil {
			t.Fatal(err)
		}
	}
	if cobros() != 2 {
		t.Fatalf("cobros: %d", cobros())
	}
}
