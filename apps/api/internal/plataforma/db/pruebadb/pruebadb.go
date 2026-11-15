// Package pruebadb arma una base de Postgres nueva por test: roles, base propia y migraciones.
// Los tests de datos corren contra Postgres de verdad, nunca contra un doble.
//
// Usa MOSTRADOR_BASE_DE_DATOS_TEST (superusuario). Sin Postgres, el test se saltea en local y
// falla en CI.
package pruebadb

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/url"
	"os"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db"
)

const porDefecto = "postgres://mostrador@127.0.0.1:5440/mostrador?sslmode=disable"

// Bases son las conexiones de cada rol a la base del test.
type Bases struct {
	App, Worker, Staff *db.Pool
	// AdminURL es la URL de superusuario a la base del test (para inspeccionar el catálogo).
	AdminURL string
	// MigradorURL: la del rol que corre las migraciones y los comandos de mostradorctl.
	MigradorURL string
}

var rolesListos sync.Once

// Nueva crea la base del test y la borra al terminar.
func Nueva(t testing.TB) *Bases {
	t.Helper()
	ctx := context.Background()
	admin := os.Getenv("MOSTRADOR_BASE_DE_DATOS_TEST")
	if admin == "" {
		admin = porDefecto
	}
	conn, err := pgx.Connect(ctx, admin)
	if err != nil {
		if os.Getenv("CI") != "" {
			t.Fatalf("sin Postgres en CI: %v", err)
		}
		t.Skipf("sin Postgres local (bun run db:up): %v", err)
	}
	defer conn.Close(ctx)

	// sync.Once cubre un paquete, pero `go test ./...` corre los paquetes en procesos paralelos que
	// crean los mismos roles a la vez ("tuple concurrently updated"): un advisory lock los pone en fila.
	var errRoles error
	rolesListos.Do(func() {
		if _, errRoles = conn.Exec(ctx, "SELECT pg_advisory_lock(735100)"); errRoles != nil {
			return
		}
		defer func() { _, _ = conn.Exec(ctx, "SELECT pg_advisory_unlock(735100)") }()
		errRoles = db.CrearRoles(ctx, admin, nil)
	})
	if errRoles != nil {
		t.Fatal(errRoles)
	}

	b := make([]byte, 6)
	_, _ = rand.Read(b)
	nombre := "prueba_" + hex.EncodeToString(b)
	if err := db.CrearBase(ctx, admin, nombre); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		c, err := pgx.Connect(context.Background(), admin)
		if err != nil {
			return
		}
		defer c.Close(context.Background())
		_, _ = c.Exec(context.Background(), "DROP DATABASE "+pgx.Identifier{nombre}.Sanitize()+" WITH (FORCE)")
	})

	migrador := conRol(t, admin, "mostrador_migrador", nombre)
	if err := db.Migrar(ctx, migrador); err != nil {
		t.Fatalf("migraciones: %v", err)
	}
	abrir := func(rol string) *db.Pool {
		p, err := db.Abrir(ctx, conRol(t, admin, rol, nombre))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(p.Close)
		return p
	}
	u, _ := url.Parse(admin)
	u.Path = "/" + nombre
	return &Bases{
		App:         abrir("mostrador_app"),
		Worker:      abrir("mostrador_worker"),
		Staff:       abrir("mostrador_staff"),
		AdminURL:    u.String(),
		MigradorURL: migrador,
	}
}

func conRol(t testing.TB, admin, rol, base string) string {
	t.Helper()
	u, err := url.Parse(admin)
	if err != nil {
		t.Fatal(err)
	}
	u.User = url.User(rol)
	u.Path = "/" + base
	return u.String()
}

// CrearTienda da de alta una tienda con el rol de la app y devuelve su id.
func CrearTienda(t testing.TB, app *db.Pool, slug string) string {
	t.Helper()
	var id string
	err := app.EnTx(context.Background(), func(tx pgx.Tx) error {
		return tx.QueryRow(context.Background(), "SELECT app.crear_tienda($1, $2, 'regionales', 'litoral')", slug, "Tienda "+slug).Scan(&id)
	})
	if err != nil {
		t.Fatal(err)
	}
	return id
}
