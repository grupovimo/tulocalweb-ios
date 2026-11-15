package avisos

import (
	"strings"
	"testing"
)

// Con Mercado Pago la confirmación sale recién con el pago acreditado (00085): es un solo mail, con
// el detalle y el pago adentro. Con transferencia sigue pidiendo el comprobante.
func TestConfirmacionConElPago(t *testing.T) {
	s := &Servicio{Sitio: "https://tulocalweb.com.ar"}
	base := pedido{numero: 1042, email: "lucia@ejemplo.com", nombre: "Lucía", tienda: "Luna", slug: "luna",
		items: []item{{Nombre: "Mate", Cantidad: 1, Precio: 10000}}, subtotal: 10000, total: 10000}
	mp := base
	mp.pago, mp.estadoPago = "mp", "Aprobado"
	m, ok := s.mail(mp, "confirmacion")
	if !ok || m.Asunto != "Tu pedido #1042 en Luna" {
		t.Fatalf("mercado pago: %v %q", ok, m.Asunto)
	}
	for _, want := range []string{"¡Gracias por tu compra!", "Se acreditó tu pago con Mercado Pago", "1 × Mate — $ 10.000", "Total — $ 10.000"} {
		if !strings.Contains(m.Texto, want) {
			t.Errorf("falta %q en:\n%s", want, m.Texto)
		}
	}
	tr := base
	tr.pago, tr.estadoPago, tr.alias = "transf", "Pendiente", "luna.mates"
	m, _ = s.mail(tr, "confirmacion")
	if strings.Contains(m.Texto, "Se acreditó") || !strings.Contains(m.Texto, "Para confirmarlo, transferí $ 10.000") {
		t.Errorf("transferencia:\n%s", m.Texto)
	}
}
