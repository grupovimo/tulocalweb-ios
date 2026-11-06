// Comando mostradorctl: tareas de operación de la base.
//
//	mostradorctl bootstrap   crea roles y base (superusuario: MOSTRADOR_BASE_DE_DATOS_ADMIN)
//	mostradorctl migrar      aplica las migraciones (MOSTRADOR_BASE_DE_DATOS_MIGRADOR)
//	mostradorctl partner-demo <email>
//	                         demo del panel de partners para una cuenta que ya existe (sólo local:
//	                         superusuario de MOSTRADOR_BASE_DE_DATOS_ADMIN sobre la base del migrador)
//	mostradorctl equipo-alta <email> <nombre> <rol>
//	                         da de alta (o actualiza) a alguien del equipo para /staff; la clave sale de
//	                         MOSTRADOR_CLAVE_NUEVA y el rol es soporte, administracion o duenio
//	mostradorctl dominio regalar|activar|quitar …
//	                         el dominio propio de una tienda en Cloudflare (ver dominio.go)
//	mostradorctl demo-tiendas <email> [borrar]
//	                         escribe el SQL de las tres tiendas de demostración (dominio/demo) para la
//	                         cuenta de ese email, o el que las borra. Se corre como superusuario:
//	                         mostradorctl demo-tiendas x@y.com | psql -U postgres -d mostrador
//
// Las claves de los roles salen de MOSTRADOR_CLAVE_<ROL> (vacías en desarrollo).
package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/demo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/partners"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/config"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "uso: mostradorctl bootstrap|migrar|partner-demo <email>|equipo-alta")
		os.Exit(2)
	}
	// Las demos no tocan la base desde acá: escriben SQL (ver dominio/demo).
	if os.Args[1] == "demo-tiendas" {
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "uso: mostradorctl demo-tiendas <email> [borrar]")
			os.Exit(2)
		}
		if len(os.Args) > 3 && os.Args[3] == "borrar" {
			fmt.Print("BEGIN;\n" + demo.Borrar() + "COMMIT;\n")
			return
		}
		s, err := demo.SQL(os.Args[2], time.Now())
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		fmt.Print(s)
		return
	}
	if err := correr(os.Args[1]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func correr(cmd string) error {
	cfg, err := config.Leer()
	if err != nil {
		return err
	}
	ctx := context.Background()
	switch cmd {
	case "bootstrap":
		if err := cfg.Exigir("BaseAdmin", "BaseMigrador"); err != nil {
			return err
		}
		u, err := url.Parse(cfg.BaseMigrador)
		if err != nil {
			return err
		}
		claves := db.Claves{}
		for _, r := range db.Roles {
			claves[r] = os.Getenv("MOSTRADOR_CLAVE_" + strings.ToUpper(strings.TrimPrefix(r, "mostrador_")))
		}
		if err := db.Bootstrap(ctx, cfg.BaseAdmin, strings.TrimPrefix(u.Path, "/"), claves); err != nil {
			return err
		}
		fmt.Println("roles y base listos")
	case "migrar":
		if err := cfg.Exigir("BaseMigrador"); err != nil {
			return err
		}
		if err := db.Migrar(ctx, cfg.BaseMigrador); err != nil {
			return err
		}
		fmt.Println("migraciones aplicadas")
	case "partner-demo":
		if len(os.Args) < 3 {
			return fmt.Errorf("falta el email: mostradorctl partner-demo nombre@mail.com")
		}
		if !cfg.Desarrollo() {
			return fmt.Errorf("la demo de partners es sólo para desarrollo")
		}
		if err := cfg.Exigir("BaseAdmin", "BaseMigrador"); err != nil {
			return err
		}
		admin, err := url.Parse(cfg.BaseAdmin)
		if err != nil {
			return err
		}
		mig, err := url.Parse(cfg.BaseMigrador)
		if err != nil {
			return err
		}
		admin.Path = mig.Path
		conn, err := pgx.Connect(ctx, admin.String())
		if err != nil {
			return err
		}
		defer conn.Close(ctx)
		if _, err := partners.Sembrar(ctx, conn, os.Args[2], time.Now()); err != nil {
			return err
		}
		fmt.Println("demo de partners lista: /partners/panel con", os.Args[2])
	case "equipo-alta":
		if len(os.Args) != 5 {
			return fmt.Errorf("uso: mostradorctl equipo-alta <email> <nombre> <soporte|administracion|duenio>")
		}
		if err := cfg.Exigir("BaseMigrador"); err != nil {
			return err
		}
		return altaEquipo(ctx, cfg.BaseMigrador, os.Args[2], os.Args[3], os.Args[4], os.Getenv("MOSTRADOR_CLAVE_NUEVA"))
	case "dominio":
		if err := cfg.Exigir("BaseMigrador"); err != nil {
			return err
		}
		return dominio(ctx, cfg.BaseMigrador, os.Args[2:])
	default:
		return fmt.Errorf("comando desconocido %q", cmd)
	}
	return nil
}

// altaEquipo crea o actualiza a una persona del equipo (con el migrador: la API no puede).
func altaEquipo(ctx context.Context, url, email, nombre, rol, clave string) error {
	if rol != "soporte" && rol != "administracion" && rol != "duenio" {
		return fmt.Errorf("rol %q: tiene que ser soporte, administracion o duenio", rol)
	}
	if len([]rune(clave)) < 12 {
		return fmt.Errorf("poné la clave en MOSTRADOR_CLAVE_NUEVA (al menos 12 caracteres)")
	}
	hash, err := cuentas.HashClave(clave)
	if err != nil {
		return err
	}
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	_, err = conn.Exec(ctx, `INSERT INTO equipo.miembros (email, nombre, clave_hash, rol) VALUES (lower($1), $2, $3, $4)
		ON CONFLICT (email) DO UPDATE SET nombre = excluded.nombre, clave_hash = excluded.clave_hash, rol = excluded.rol, activo = true`,
		strings.TrimSpace(email), strings.TrimSpace(nombre), hash, rol)
	if err == nil {
		fmt.Println("listo:", email, "es", rol)
	}
	return err
}
