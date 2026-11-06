package cuentas

import (
	"testing"
	"time"
)

// Los vectores del RFC 6238 (apéndice B, SHA-1), en sus últimos seis números.
func TestCodigoTOTPdelRFC(t *testing.T) {
	secreto := []byte("12345678901234567890")
	for _, c := range []struct {
		t   int64
		cod string
	}{{59, "287082"}, {1111111109, "081804"}, {1111111111, "050471"}, {1234567890, "005924"}, {2000000000, "279037"}} {
		if got := codigoTOTP(secreto, c.t/30); got != c.cod {
			t.Errorf("t=%d: %s, esperaba %s", c.t, got, c.cod)
		}
	}
	ahora := time.Unix(1111111111, 0)
	if !totpValido(secreto, "050 471", ahora) || !totpValido(secreto, codigoTOTP(secreto, ahora.Unix()/30-1), ahora) {
		t.Error("tiene que aceptar el código actual (con espacios) y el del paso anterior")
	}
	if totpValido(secreto, codigoTOTP(secreto, ahora.Unix()/30-3), ahora) || totpValido(secreto, "12345", ahora) {
		t.Error("no tiene que aceptar un código viejo ni uno corto")
	}
	if intervaloAGo("8760:00:00") != "8760h00m00s" {
		t.Error(intervaloAGo("8760:00:00"))
	}
}
