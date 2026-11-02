package configuracion_test

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/mitienda"
	"net/url"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/configuracion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

const sitio = "https://tulocalweb.test"

// El equipo de punta a punta: invitar, reenviar, aceptar con una cuenta que ya existe, crear la
// cuenta desde la invitación, cambiar el rol, quitar (cierra sus sesiones en esa tienda) y que a
// la dueña no se la toque.
func TestEquipo(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	s := &cuentas.Servicio{Pool: b.App, Correo: &correo.Memoria{}, Sitio: sitio}
	cl := cuentas.Cliente{IP: "190.210.32.1", Agente: "Mozilla/5.0 (Macintosh) Chrome/130"}
	ana, err := s.Registrar(ctx, cuentas.Alta{Nombre: "Ana", Tienda: "Sol", Email: "ana@sol.com", Clave: "clave-larga-1"}, cl)
	if err != nil {
		t.Fatal(err)
	}
	leo, err := s.Registrar(ctx, cuentas.Alta{Nombre: "Leo", Tienda: "Leo Deco", Email: "leo@deco.com", Clave: "clave-larga-2"}, cl)
	if err != nil {
		t.Fatal(err)
	}
	id := ana.TiendaID
	usuario := func(email string) string {
		var u string
		if err := b.App.QueryRow(ctx, `SELECT id::text FROM auth.usuarios WHERE email = $1`, email).Scan(&u); err != nil {
			t.Fatal(err)
		}
		return u
	}
	anaID, leoID := usuario("ana@sol.com"), usuario("leo@deco.com")
	q := configuracion.Quien{UsuarioID: anaID, Rol: "duenio"}
	// El último mail de invitación del outbox (lo lee el worker; la app no ve los eventos sin tienda).
	ultimo := func(para string) (cuentas.MailInvitacion, string) {
		t.Helper()
		var raw []byte
		if err := b.Worker.QueryRow(ctx, `SELECT datos FROM outbox WHERE tipo = 'cuenta.invitacion' AND datos->>'para' = $1
			ORDER BY creado_en DESC, id DESC LIMIT 1`, para).Scan(&raw); err != nil {
			t.Fatal(err)
		}
		var d cuentas.MailInvitacion
		_ = json.Unmarshal(raw, &d)
		u, err := url.Parse(d.Link)
		if err != nil || u.Path != "/invitacion" {
			t.Fatalf("link: %q", d.Link)
		}
		return d, u.Query().Get("token")
	}

	if _, err := configuracion.Invitar(ctx, b.App, sitio, id, q, "leo@deco", "ventas"); !errors.Is(err, configuracion.ErrInvalido) {
		t.Fatalf("un email mal escrito: %v", err)
	}
	if _, err := configuracion.Invitar(ctx, b.App, sitio, id, q, "otra@sol.com", "duenio"); !errors.Is(err, configuracion.ErrInvalido) {
		t.Fatalf("invitó como dueña: %v", err)
	}
	m, err := configuracion.Invitar(ctx, b.App, sitio, id, q, "Leo@Deco.com", "ventas")
	if err != nil || !m.Pendiente || m.Vence == nil || time.Until(*m.Vence) < 6*24*time.Hour {
		t.Fatalf("invitar: %+v %v", m, err)
	}
	if _, err := configuracion.Invitar(ctx, b.App, sitio, id, q, "leo@deco.com", "ventas"); !errors.Is(err, configuracion.ErrRepetido) {
		t.Fatalf("dos invitaciones al mismo email: %v", err)
	}
	mail, viejo := ultimo("leo@deco.com")
	if mail.Tienda != "Sol" || mail.Invita != "Ana" || mail.Rol != "Ventas" || mail.Vence == "" {
		t.Fatalf("mail: %+v", mail)
	}

	t.Run("reenviar deja sin efecto el link anterior", func(t *testing.T) {
		if err := configuracion.Reenviar(ctx, b.App, sitio, id, q, m.ID); err != nil {
			t.Fatal(err)
		}
		if _, err := s.VerInvitacion(ctx, viejo); !errors.Is(err, cuentas.ErrTokenInvalido) {
			t.Fatalf("el link viejo sigue: %v", err)
		}
	})
	_, token := ultimo("leo@deco.com")
	inv, err := s.VerInvitacion(ctx, token)
	if err != nil || inv.Tienda != "Sol" || inv.Rol != "ventas" || inv.Email != "leo@deco.com" || !inv.TieneCuenta || inv.Invita != "Ana" {
		t.Fatalf("ver: %+v %v", inv, err)
	}

	t.Run("otra cuenta no puede usar el link de Leo", func(t *testing.T) {
		mia, err := s.Registrar(ctx, cuentas.Alta{Nombre: "Mía", Tienda: "Mía", Email: "mia@otra.com", Clave: "clave-larga-3"}, cl)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := s.AceptarInvitacion(ctx, mia.Token, token); !errors.Is(err, cuentas.ErrOtroEmail) {
			t.Fatalf("aceptar con otra cuenta: %v", err)
		}
		if _, err := s.VerInvitacion(ctx, token); err != nil {
			t.Fatalf("el link de Leo se gastó: %v", err)
		}
	})
	t.Run("con cuenta: se suma y la sesión pasa a esa tienda", func(t *testing.T) {
		tienda, err := s.AceptarInvitacion(ctx, leo.Token, token)
		if err != nil || tienda != id {
			t.Fatalf("aceptar: %q %v", tienda, err)
		}
		if _, err := s.AceptarInvitacion(ctx, leo.Token, token); !errors.Is(err, cuentas.ErrTokenInvalido) {
			t.Fatalf("se usó dos veces: %v", err)
		}
		// La campanita de la dueña (§45): un aviso, con la persona que se sumó.
		var avisos int
		if err := b.App.EnTienda(ctx, id, func(tx pgx.Tx) error {
			return tx.QueryRow(ctx, `SELECT count(*) FROM outbox WHERE tipo = 'equipo.aceptada' AND datos->>'usuario' = (SELECT id::text FROM auth.usuarios WHERE email = 'leo@deco.com')`).Scan(&avisos)
		}); err != nil || avisos != 1 {
			t.Fatalf("aviso de invitación aceptada: %d %v", avisos, err)
		}
		d, err := s.Sesion(ctx, leo.Token)
		if err != nil || d.TiendaID != id || d.Rol != "ventas" {
			t.Fatalf("sesión: %+v %v", d, err)
		}
		// El logo de Mi tienda viaja con las tiendas de la persona (la app lo muestra en vez de las iniciales).
		dis := mitienda.Inicial()
		dis.Logo = "https://tulocalweb.com.ar/f/fotos/logo.webp"
		if err := mitienda.Guardar(ctx, b.App, leo.TiendaID, dis); err != nil {
			t.Fatal(err)
		}
		a, err := s.AccesoDe(ctx, d)
		if err != nil || len(a.Tiendas) != 2 || a.Tiendas[0].ID != leo.TiendaID || a.Tiendas[1].Rol != "ventas" ||
			len(a.Permisos) != 4 {
			t.Fatalf("acceso: %+v %v", a, err)
		}
		if a.Tiendas[0].Logo == nil || *a.Tiendas[0].Logo != dis.Logo || a.Tiendas[1].Logo != nil {
			t.Fatalf("logo de las tiendas: %v %v", a.Tiendas[0].Logo, a.Tiendas[1].Logo)
		}
		if err := s.ElegirTienda(ctx, leo.Token, leo.TiendaID); err != nil {
			t.Fatal(err)
		}
		if d, _ := s.Sesion(ctx, leo.Token); d.TiendaID != leo.TiendaID || d.Rol != "duenio" {
			t.Fatalf("eligió su tienda: %+v", d)
		}
		if err := s.ElegirTienda(ctx, ana.Token, leo.TiendaID); !errors.Is(err, cuentas.ErrNoEsTuya) {
			t.Fatalf("entró a una tienda ajena: %v", err)
		}
		if err := s.ElegirTienda(ctx, leo.Token, id); err != nil {
			t.Fatal(err)
		}
	})

	var miaID string
	t.Run("sin cuenta: la crea sin tienda propia", func(t *testing.T) {
		if _, err := configuracion.Invitar(ctx, b.App, sitio, id, q, "mia@mail.com", "local"); err != nil {
			t.Fatal(err)
		}
		_, tok := ultimo("mia@mail.com")
		if inv, _ := s.VerInvitacion(ctx, tok); inv.TieneCuenta {
			t.Fatal("dice que tiene cuenta")
		}
		ses, err := s.RegistrarConInvitacion(ctx, tok, "Mía", "clave-larga-3", cl)
		if err != nil || ses.TiendaID != id {
			t.Fatalf("registro: %+v %v", ses, err)
		}
		d, _ := s.Sesion(ctx, ses.Token)
		a, _ := s.AccesoDe(ctx, d)
		if d.Rol != "local" || len(a.Tiendas) != 1 {
			t.Fatalf("una sola tienda, la de la invitación: %+v %+v", d, a)
		}
		if _, err := s.Ingresar(ctx, "mia@mail.com", "clave-larga-3", false, cl); err != nil {
			t.Fatalf("entra con su clave: %v", err)
		}
		miaID = usuario("mia@mail.com")
	})

	t.Run("roles: la dueña no se toca", func(t *testing.T) {
		if err := configuracion.CambiarRol(ctx, b.App, id, q, miaID, "admin"); err != nil {
			t.Fatal(err)
		}
		if err := configuracion.CambiarRol(ctx, b.App, id, q, anaID, "admin"); !errors.Is(err, configuracion.ErrVos) {
			t.Fatalf("se cambió el rol a sí misma: %v", err)
		}
		otra := configuracion.Quien{UsuarioID: miaID, Rol: "admin"}
		if err := configuracion.CambiarRol(ctx, b.App, id, otra, anaID, "ventas"); !errors.Is(err, configuracion.ErrDuenia) {
			t.Fatalf("degradó a la dueña: %v", err)
		}
		if err := configuracion.Quitar(ctx, b.App, id, otra, anaID); !errors.Is(err, configuracion.ErrDuenia) {
			t.Fatalf("quitó a la dueña: %v", err)
		}
		if err := configuracion.CambiarRol(ctx, b.App, id, q, miaID, "duenio"); !errors.Is(err, configuracion.ErrInvalido) {
			t.Fatalf("pasó a alguien a dueña: %v", err)
		}
	})

	t.Run("quitar cierra sus sesiones en esta tienda", func(t *testing.T) {
		if err := configuracion.Quitar(ctx, b.App, id, q, leoID); err != nil {
			t.Fatal(err)
		}
		if _, err := s.Sesion(ctx, leo.Token); !errors.Is(err, cuentas.ErrSinSesion) {
			t.Fatalf("la sesión en Sol sigue: %v", err)
		}
		if err := configuracion.Quitar(ctx, b.App, id, q, leoID); !errors.Is(err, configuracion.ErrNoExiste) {
			t.Fatalf("quitó dos veces: %v", err)
		}
		if ses, err := s.Ingresar(ctx, "leo@deco.com", "clave-larga-2", false, cl); err != nil || ses.TiendaID != leo.TiendaID {
			t.Fatalf("su cuenta y su tienda siguen: %+v %v", ses, err)
		}
	})

	t.Run("cancelar la invitación anula el link", func(t *testing.T) {
		m, err := configuracion.Invitar(ctx, b.App, sitio, id, q, "nadie@mail.com", "ventas")
		if err != nil {
			t.Fatal(err)
		}
		_, tok := ultimo("nadie@mail.com")
		if err := configuracion.CambiarRol(ctx, b.App, id, q, m.ID, "local"); err != nil {
			t.Fatalf("cambiar el rol de la pendiente: %v", err)
		}
		if inv, _ := s.VerInvitacion(ctx, tok); inv.Rol != "local" {
			t.Fatalf("el rol nuevo: %+v", inv)
		}
		if err := configuracion.Quitar(ctx, b.App, id, q, m.ID); err != nil {
			t.Fatal(err)
		}
		if _, err := s.VerInvitacion(ctx, tok); !errors.Is(err, cuentas.ErrTokenInvalido) {
			t.Fatalf("el link sigue: %v", err)
		}
	})

	t.Run("quien no es dueña no ve el equipo ni los datos de cobro", func(t *testing.T) {
		if err := configuracion.GuardarNegocio(ctx, b.App, id, configuracion.Negocio{Nombre: "Sol", Alias: "sol.deco.mp"}); err != nil {
			t.Fatal(err)
		}
		c, err := configuracion.Leer(ctx, b.App, nil, id, configuracion.Quien{UsuarioID: miaID, Rol: "local"}, time.Now())
		if err != nil {
			t.Fatal(err)
		}
		eq := c["equipo"].([]configuracion.Miembro)
		if n := c["negocio"].(configuracion.Negocio); n.Alias != "" || n.Nombre != "Sol" || len(eq) != 1 || !eq[0].Vos {
			t.Fatalf("local ve: %+v %+v", n, eq)
		}
		c, _ = configuracion.Leer(ctx, b.App, nil, id, q, time.Now())
		if n := c["negocio"].(configuracion.Negocio); n.Alias != "sol.deco.mp" || len(c["equipo"].([]configuracion.Miembro)) != 2 {
			t.Fatalf("la dueña ve: %+v", c)
		}
	})

	p, err := configuracion.Plan(ctx, b.App, id, time.Now())
	if err != nil || p["uso"].(map[string]any)["personas"] != 2 {
		t.Fatalf("plan: %+v %v", p, err)
	}
}
