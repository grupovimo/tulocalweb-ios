package push_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/push"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

func (e *escenario) actividad(t *testing.T, rol, cuerpo string) int {
	t.Helper()
	r := httptest.NewRequest("PUT", "/v1/panel/dispositivos/actividad", strings.NewReader(cuerpo))
	r.Header.Set("Authorization", "Bearer "+e.ses[rol])
	r.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.h.ServeHTTP(rec, r)
	return rec.Code
}

func TestPayloadActividad(t *testing.T) {
	n, c := 1048, "Lucía"
	e := push.EstadoActividad{PorPreparar: 3, PreparadosHoy: 2, UltimoNumero: &n, UltimoCliente: &c}
	ahora := time.Unix(1_800_000_000, 0)
	var p struct {
		Aps map[string]any `json:"aps"`
	}
	if err := json.Unmarshal(push.PayloadActividad("start", e, "Luna", ahora), &p); err != nil {
		t.Fatal(err)
	}
	if p.Aps["event"] != "start" || p.Aps["attributes-type"] != "PedidosPorPreparar" || p.Aps["alert"] == nil {
		t.Fatalf("start: %v", p.Aps)
	}
	cs := p.Aps["content-state"].(map[string]any)
	if cs["porPreparar"] != 3.0 || cs["ultimoNumero"] != 1048.0 || cs["ultimoCliente"] != "Lucía" || cs["preparadosHoy"] != 2.0 {
		t.Fatalf("content-state: %v", cs)
	}
	var fin struct {
		Aps map[string]any `json:"aps"`
	}
	_ = json.Unmarshal(push.PayloadActividad("end", push.EstadoActividad{}, "Luna", ahora), &fin)
	if fin.Aps["event"] != "end" || fin.Aps["dismissal-date"] == nil || fin.Aps["attributes"] != nil {
		t.Fatalf("end: %v", fin.Aps)
	}
}

// Con una venta nueva, el worker actualiza la actividad de quien tiene Pedidos y la arranca en el
// iPhone que sólo tiene el token de inicio (Local también prepara pedidos: le llega).
func TestActividadConVentaNueva(t *testing.T) {
	e := armar(t)
	ctx := context.Background()
	for rol, c := range map[string]string{"duenio": "a", "ventas": "c", "local": "d"} {
		e.registrar(t, rol, token(c))
	}
	actividad := strings.Repeat("ab", 40)
	inicio := strings.Repeat("cd", 40)
	if code := e.actividad(t, "duenio", `{"token":"`+token("a")+`","actividad":"`+actividad+`"}`); code != http.StatusNoContent {
		t.Fatalf("guardar actividad: %d", code)
	}
	if code := e.actividad(t, "ventas", `{"token":"`+token("c")+`","inicio":"`+inicio+`"}`); code != http.StatusNoContent {
		t.Fatalf("guardar inicio: %d", code)
	}
	if code := e.actividad(t, "local", `{"token":"`+token("d")+`","actividad":"`+strings.Repeat("ef", 40)+`"}`); code != http.StatusNoContent {
		t.Fatalf("guardar local: %d", code)
	}
	// Un token ajeno no se puede tocar ni con una sesión válida.
	if code := e.actividad(t, "ventas", `{"token":"`+token("a")+`","actividad":""}`); code != http.StatusNoContent {
		t.Fatalf("token ajeno: %d", code)
	}
	if n := e.cuantos(t, "token_actividad = $1", actividad); n != 1 {
		t.Fatal("otra persona borró la actividad del dueño")
	}
	if code := e.actividad(t, "duenio", `{"token":"no-hex"}`); code != http.StatusBadRequest && code != http.StatusUnprocessableEntity {
		t.Fatalf("token inválido: %d", code)
	}

	f := &apnsFalso{}
	s := &push.Servicio{Pool: e.b.App, APNs: f, Log: silencio()}
	w := &outbox.Worker{Pool: e.b.Worker, Manejadores: s.Manejadores()}
	vender(t, e)
	// El pedido de prueba puede quedar «Sin pagar» (transferencia): se lo pasa a preparar.
	e.admin(t, `UPDATE pedidos SET estado = 'Para preparar' WHERE tienda_id = $1`, e.tienda)
	if _, err := w.Pasada(ctx); err != nil {
		t.Fatal(err)
	}
	eventos := map[string]string{}
	for _, x := range f.envios {
		if ev, ok := x.aviso["aps"].(map[string]any)["event"].(string); ok {
			eventos[x.token] = ev
		}
	}
	if eventos[actividad] != "update" || eventos[inicio] != "start" || eventos[strings.Repeat("ef", 40)] != "update" || len(eventos) != 3 {
		t.Fatalf("eventos de actividad: %v", eventos)
	}
}
