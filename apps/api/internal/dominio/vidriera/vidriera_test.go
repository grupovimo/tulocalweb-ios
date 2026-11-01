package vidriera_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/configuracion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cupones"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/vidriera"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

type mpFalso struct{ pedidos []int }

func (m *mpFalso) PagarPedido(_ context.Context, _ string, numero int, vuelta string) (string, error) {
	m.pedidos = append(m.pedidos, numero)
	return "https://www.mercadopago.com.ar/checkout/v1/redirect?pref_id=x&vuelta=" + vuelta, nil
}

func compra(id string, cantidad int, pago string) vidriera.Compra {
	c := vidriera.Compra{Email: "lucia@ejemplo.com", Nombre: "Lucía Pérez", Telefono: "11 5555-0101", CP: "1414",
		Direccion: "Gorriti 4500", Envio: "retiro", Pago: pago}
	c.Carrito = append(c.Carrito, vidriera.LineaCarrito{ID: id, Cantidad: cantidad})
	return c
}

func TestTiendaOnline(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	sol := pruebadb.CrearTienda(t, b.App, "sol")
	// Sin alias ni CBU no se ofrece transferencia: la tienda los carga.
	if err := configuracion.GuardarNegocio(ctx, b.App, luna, configuracion.Negocio{Nombre: "Luna", Alias: "luna.tienda"}); err != nil {
		t.Fatal(err)
	}
	stock := 3
	cat := "Mates"
	p, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Mate de calabaza", Categoria: &cat, Precio: 10000, Stock: &stock})
	if err != nil {
		t.Fatal(err)
	}

	tienda, err := vidriera.Publica(ctx, b.App, "luna")
	if err != nil || len(tienda.Productos) != 1 || len(tienda.Categorias) != 1 || tienda.Categorias[0].Cantidad != 1 {
		t.Fatalf("pública: %+v %v", tienda, err)
	}
	if _, err := vidriera.Publica(ctx, b.App, "no-existe"); !errors.Is(err, vidriera.ErrNoExiste) {
		t.Fatal("una tienda que no existe")
	}

	// Las promociones arrancan apagadas: sin prender la de transferencia, no hay descuento.
	if tienda.Cobros["transferencia"] != 0 || tienda.Envios.GratisDesde != nil {
		t.Fatalf("promociones prendidas de entrada: %+v %+v", tienda.Cobros, tienda.Envios)
	}
	if err := cupones.FijarAutomatica(ctx, b.App, luna, "transf", true); err != nil {
		t.Fatal(err)
	}
	// Transferencia: precio del servidor, 15 % off y el código público (Correo Argentino se prueba en
	// integraciones/correoargentino, con la cotización de la API).
	lunes := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)
	r, err := vidriera.Comprar(ctx, b.App, nil, "https://tulocalweb.com.ar", "luna", compra(p.ID, 2, "transferencia"), lunes)
	if err != nil {
		t.Fatal(err)
	}
	if want := int64(20000 - 3000); r.Total != want || r.URL != "" || len(r.Codigo) < 16 {
		t.Fatalf("resultado: %+v (quería total %d)", r, want)
	}
	c, err := vidriera.VerPedido(ctx, b.App, "luna", r.Codigo)
	if err != nil || c.Numero != r.Numero || c.Total != r.Total || c.Nombre != "Lucía" || c.Pago != "transferencia" {
		t.Fatalf("confirmación: %+v %v", c, err)
	}
	if _, err := vidriera.VerPedido(ctx, b.App, "sol", r.Codigo); !errors.Is(err, vidriera.ErrNoExiste) {
		t.Fatal("otra tienda vio el pedido")
	}
	var quedan int
	_ = b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT stock FROM productos WHERE id::text = $1`, p.ID).Scan(&quedan)
	})
	if quedan != 1 {
		t.Fatalf("stock: %d", quedan)
	}

	// Sin stock, carrito de otra tienda, y Mercado Pago sin cuenta conectada: no se arma el pedido.
	if _, err := vidriera.Comprar(ctx, b.App, nil, "", "luna", compra(p.ID, 5, "transferencia"), lunes); !errors.Is(err, vidriera.ErrSinStock) {
		t.Fatalf("sin stock: %v", err)
	}
	if _, err := vidriera.Comprar(ctx, b.App, nil, "", "sol", compra(p.ID, 1, "transferencia"), lunes); !errors.Is(err, vidriera.ErrInvalido) {
		t.Fatalf("producto de otra tienda: %v", err)
	}
	mp := &mpFalso{}
	if _, err := vidriera.Comprar(ctx, b.App, mp, "", "luna", compra(p.ID, 1, "mercadopago"), lunes); !errors.Is(err, vidriera.ErrInvalido) {
		t.Fatalf("Mercado Pago sin conectar: %v", err)
	}

	// Una tienda cerrada no se ve.
	_ = b.App.EnTienda(ctx, sol, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE tiendas SET cerrada_en = now() WHERE id = $1`, sol)
		return err
	})
	if _, err := vidriera.Publica(ctx, b.App, "sol"); !errors.Is(err, vidriera.ErrNoExiste) {
		t.Fatal("se ve una tienda cerrada")
	}
}

func TestCotizar(t *testing.T) {
	lunes := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)
	gratis := int64(100000)
	e := vidriera.Envios{GratisDesde: &gratis, Retiro: true, RetiroDireccion: "Gorriti 4500", RetiroHorario: "Lun a vie de 10 a 19",
		Moto: true, MotoCosto: 3500, MotoZona: "Palermo y Villa Crespo", MotoDemora: "En el día", Correo: true}
	o := vidriera.Cotizar("1414", 5000, e, lunes)
	// Moto (con el costo y la zona de la tienda) y retiro. Correo Argentino no sale de una tabla: lo
	// cotiza la API (integraciones/correoargentino).
	if len(o) != 2 || o[0].ID != "moto" || o[0].Precio != 3500 || o[0].Detalle != "Llega a Palermo y Villa Crespo · En el día" ||
		o[1].ID != "retiro" || o[1].Detalle != "Gorriti 4500 · Lun a vie de 10 a 19" {
		t.Fatalf("con CP: %+v", o)
	}
	// Desde el monto del envío gratis, la moto también es gratis.
	if o := vidriera.Cotizar("5000", 150000, e, lunes); len(o) != 2 || o[0].Precio != 0 {
		t.Fatalf("envío gratis: %+v", o)
	}
	// Sin CP no hay adónde mandar la moto: sólo retiro.
	if o := vidriera.Cotizar("x", 1, e, lunes); len(o) != 1 || o[0].ID != "retiro" {
		t.Fatalf("CP inválido: %+v", o)
	}
	// Sin retiro ni moto prendidos, nada.
	if o := vidriera.Cotizar("1414", 1, vidriera.Envios{Correo: true}, lunes); len(o) != 0 {
		t.Fatalf("todo apagado: %+v", o)
	}
	if o := vidriera.Cotizar("", 1, vidriera.Envios{Retiro: true}, lunes); len(o) != 1 || o[0].Detalle != "Te avisamos cuando esté listo" {
		t.Fatalf("retiro sin dirección: %+v", o)
	}
}

// La política de cambios y devoluciones de Configuración › Negocio llega a la tienda, sin espacios
// de más; más de 1000 caracteres no se guarda.
func TestPoliticaDeCambios(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	if tienda, err := vidriera.Publica(ctx, b.App, "luna"); err != nil || tienda.Politica != "" {
		t.Fatalf("de entrada: %q %v", tienda.Politica, err)
	}
	politica := "Cambios dentro de los 30 días con el ticket."
	if err := configuracion.GuardarNegocio(ctx, b.App, luna, configuracion.Negocio{Nombre: "Luna", Politica: "  " + politica + "\n"}); err != nil {
		t.Fatal(err)
	}
	if tienda, err := vidriera.Publica(ctx, b.App, "luna"); err != nil || tienda.Politica != politica {
		t.Fatalf("guardada: %q %v", tienda.Politica, err)
	}
	if err := configuracion.GuardarNegocio(ctx, b.App, luna, configuracion.Negocio{Nombre: "Luna", Politica: strings.Repeat("ñ", 1000)}); err != nil {
		t.Fatalf("1000 caracteres (2000 bytes): %v", err)
	}
	if err := configuracion.GuardarNegocio(ctx, b.App, luna, configuracion.Negocio{Nombre: "Luna", Politica: strings.Repeat("a", 1001)}); !errors.Is(err, configuracion.ErrInvalido) {
		t.Fatalf("1001 caracteres: %v", err)
	}
}

// Lo de Envíos y retiro llega a la tienda y al pedido: sin retiro no se puede retirar, y la moto
// cobra su costo y queda como envío (con la dirección de entrega).
func TestEntregasConfiguradas(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	if err := configuracion.GuardarNegocio(ctx, b.App, luna, configuracion.Negocio{Nombre: "Luna", Alias: "luna.tienda",
		Direccion: "Gorriti 4500", Horario: "Lun a vie de 10 a 19"}); err != nil {
		t.Fatal(err)
	}
	stock := 10
	p, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Mate", Precio: 10000, Stock: &stock})
	if err != nil {
		t.Fatal(err)
	}
	lunes := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)

	// Sin configurar: retiro con la dirección y el horario del negocio, sin moto.
	tienda, err := vidriera.Publica(ctx, b.App, "luna")
	if err != nil || !tienda.Envios.Retiro || tienda.Envios.RetiroDireccion != "Gorriti 4500" ||
		tienda.Envios.RetiroHorario != "Lun a vie de 10 a 19" || tienda.Envios.Moto {
		t.Fatalf("de entrada: %+v %v", tienda.Envios, err)
	}
	moto := compra(p.ID, 1, "transferencia")
	moto.Envio = "moto"
	if _, err := vidriera.Comprar(ctx, b.App, nil, "", "luna", moto, lunes); !errors.Is(err, vidriera.ErrInvalido) {
		t.Fatalf("moto sin prender: %v", err)
	}

	// Sólo moto: el retiro deja de ofrecerse y la moto cobra lo que puso la tienda.
	if err := configuracion.GuardarEntregas(ctx, b.App, luna, configuracion.Entregas{Moto: true, MotoCosto: 3500,
		MotoZona: "Palermo", MotoDemora: "En el día"}); err != nil {
		t.Fatal(err)
	}
	tienda, _ = vidriera.Publica(ctx, b.App, "luna")
	if tienda.Envios.Retiro || tienda.Envios.RetiroDireccion != "" || !tienda.Envios.Moto || tienda.Envios.MotoCosto != 3500 ||
		tienda.Envios.MotoZona != "Palermo" {
		t.Fatalf("sólo moto: %+v", tienda.Envios)
	}
	if _, err := vidriera.Comprar(ctx, b.App, nil, "", "luna", compra(p.ID, 1, "transferencia"), lunes); !errors.Is(err, vidriera.ErrInvalido) {
		t.Fatalf("retiro apagado: %v", err)
	}
	r, err := vidriera.Comprar(ctx, b.App, nil, "", "luna", moto, lunes)
	if err != nil || r.Total != 10000+3500 {
		t.Fatalf("moto: %+v %v", r, err)
	}
	var envio, direccion string
	var costo int64
	_ = b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT envio, costo_envio, coalesce(entrega->>'direccion', '') FROM pedidos WHERE numero = $1`, r.Numero).
			Scan(&envio, &costo, &direccion)
	})
	if envio != "moto" || costo != 3500 || direccion != "Gorriti 4500" {
		t.Fatalf("pedido con moto: %s %d %q", envio, costo, direccion)
	}
}

func TestCompraConCupon(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	// Con datos para transferir: sin ellos la tienda no ofrece transferencia.
	if err := configuracion.GuardarNegocio(ctx, b.App, luna, configuracion.Negocio{Nombre: "Luna", Alias: "luna.tienda"}); err != nil {
		t.Fatal(err)
	}
	stock := 50
	p, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Mate", Precio: 10000, Stock: &stock})
	if err != nil {
		t.Fatal(err)
	}
	if err := cupones.FijarAutomatica(ctx, b.App, luna, "transf", true); err != nil {
		t.Fatal(err)
	}
	ahora := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)
	nuevo := func(codigo, tipo string, valor int, limite int, porCliente string) {
		t.Helper()
		if _, err := cupones.Crear(ctx, b.App, luna, cupones.Nuevo{Codigo: codigo, Tipo: tipo, Valor: valor, Alcance: "todo",
			Quien: "todos", Limite: limite, PorCliente: porCliente, Vigencia: "sin", Canales: []string{"web"}}, ahora); err != nil {
			t.Fatal(err)
		}
	}
	nuevo("DIEZ", "pct", 10, 0, "1")
	nuevo("UNO", "fijo", 1000, 1, "libre")
	nuevo("ENVIO", "envio", 0, 0, "libre")

	// 10 % sobre $ 20.000 con transferencia: 2.000 de cupón y 15 % de lo que queda (2.700). Sin
	// "combinable" en el cupón, la transferencia no se suma: sólo el cupón.
	c := compra(p.ID, 2, "transferencia")
	c.Cupon = "diez"
	r, err := vidriera.Comprar(ctx, b.App, nil, "", "luna", c, ahora)
	if err != nil {
		t.Fatal(err)
	}
	if want := int64(20000 - 2000); r.Total != want {
		t.Fatalf("total con cupón: %d, quería %d", r.Total, want)
	}
	// Una vez por cliente: el mismo email ya no puede.
	var ec vidriera.ErrCupon
	if _, err := vidriera.Comprar(ctx, b.App, nil, "", "luna", c, ahora); !errors.As(err, &ec) || ec.Motivo != "Ya usaste este cupón." {
		t.Fatalf("segundo uso: %v", err)
	}
	// Límite total de 1: el segundo, aunque sea otra persona, ya no.
	c2 := compra(p.ID, 1, "mercadopago")
	c2.Pago, c2.Cupon = "transferencia", "UNO"
	if _, err := vidriera.Comprar(ctx, b.App, nil, "", "luna", c2, ahora); err != nil {
		t.Fatal(err)
	}
	c2.Email = "otra@ejemplo.com"
	if _, err := vidriera.Comprar(ctx, b.App, nil, "", "luna", c2, ahora); !errors.As(err, &ec) {
		t.Fatalf("límite: %v", err)
	}
	// Envío gratis y código inexistente.
	c3 := compra(p.ID, 1, "transferencia")
	c3.Email, c3.Cupon = "tres@ejemplo.com", "ENVIO"
	r3, err := vidriera.Comprar(ctx, b.App, nil, "", "luna", c3, ahora)
	// No combinable: el envío sale gratis pero no se suma el 15 % de transferencia.
	if err != nil || r3.Total != 10000 {
		t.Fatalf("envío gratis: %+v %v", r3, err)
	}
	c3.Cupon = "NOEXISTE"
	if _, err := vidriera.Comprar(ctx, b.App, nil, "", "luna", c3, ahora); !errors.As(err, &ec) || !strings.Contains(ec.Motivo, "no existe") {
		t.Fatalf("inexistente: %v", err)
	}
	// El cupón suma usos y lo que descontó.
	var usos int
	var desconto int64
	_ = b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT usos, desconto FROM cupones WHERE codigo = 'DIEZ'`).Scan(&usos, &desconto)
	})
	if usos != 1 || desconto != 2000 {
		t.Fatalf("contadores: %d %d", usos, desconto)
	}
}

// Dos compras a la vez del mismo email con un cupón de una vez por cliente: sólo una lo usa. Y el
// cupón de otra tienda con el mismo código no se toca.
func TestCuponPorClienteConcurrenteYAislado(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	ahora := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)
	ids := map[string]string{}
	for _, slug := range []string{"luna", "sol"} {
		id := pruebadb.CrearTienda(t, b.App, slug)
		if err := configuracion.GuardarNegocio(ctx, b.App, id, configuracion.Negocio{Nombre: slug, Alias: slug + ".tienda"}); err != nil {
			t.Fatal(err)
		}
		if _, err := cupones.Crear(ctx, b.App, id, cupones.Nuevo{Codigo: "DIEZ", Tipo: "pct", Valor: 10, Alcance: "todo",
			Quien: "todos", PorCliente: "1", Vigencia: "sin", Canales: []string{"web"}}, ahora); err != nil {
			t.Fatal(err)
		}
		ids[slug] = id
	}
	stock := 50
	p, err := listas.Guardar(ctx, b.App, ids["luna"], "", listas.Editable{Nombre: "Mate", Precio: 10000, Stock: &stock})
	if err != nil {
		t.Fatal(err)
	}
	errs := make(chan error, 2)
	for range 2 {
		go func() {
			c := compra(p.ID, 1, "transferencia")
			c.Cupon = "DIEZ"
			_, err := vidriera.Comprar(ctx, b.App, nil, "", "luna", c, ahora)
			errs <- err
		}()
	}
	ok := 0
	for range 2 {
		if err := <-errs; err == nil {
			ok++
		} else if ec := (vidriera.ErrCupon{}); !errors.As(err, &ec) {
			t.Fatal(err)
		}
	}
	if ok != 1 {
		t.Fatalf("compras con el cupón: %d, quería 1", ok)
	}
	var usos int
	_ = b.App.EnTienda(ctx, ids["sol"], func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT usos FROM cupones WHERE codigo = 'DIEZ'`).Scan(&usos)
	})
	if usos != 0 {
		t.Fatalf("el cupón de sol sumó %d usos", usos)
	}
}

func TestTiendaConVariantes(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	// Con datos para transferir (sin ellos la tienda no ofrece transferencia).
	if err := configuracion.GuardarNegocio(ctx, b.App, luna, configuracion.Negocio{Nombre: "Luna", Alias: "luna.tienda"}); err != nil {
		t.Fatal(err)
	}
	if err := cupones.FijarAutomatica(ctx, b.App, luna, "transf", true); err != nil {
		t.Fatal(err)
	}
	pruebadb.CrearTienda(t, b.App, "sol")
	cinco, uno, precio, no := 5, 1, int64(26000), false
	p, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Buzo frisa", Precio: 24000,
		Fotos:    []string{"https://f/buzo.jpg", "https://f/buzo-gris.jpg"},
		Opciones: []listas.Opcion{{Nombre: "Talle", Valores: []string{"M", "L", "XL"}}},
		Variantes: []listas.VarianteEditable{
			{Valores: map[string]string{"Talle": "M"}, Stock: &cinco},
			{Valores: map[string]string{"Talle": "L"}, Stock: &uno, Precio: &precio},
			{Valores: map[string]string{"Talle": "XL"}, Stock: &cinco, Activa: &no},
		}})
	if err != nil {
		t.Fatal(err)
	}
	id := map[string]string{}
	for _, v := range p.Variantes {
		id[v.Valores["Talle"]] = v.ID
	}

	// La tienda ve las fotos, las opciones y sólo las variantes activas (sin SKU); el stock es la suma.
	tienda, err := vidriera.Publica(ctx, b.App, "luna")
	if err != nil {
		t.Fatal(err)
	}
	pp := tienda.Productos[0]
	if len(pp.Fotos) != 2 || *pp.Foto != "https://f/buzo.jpg" || len(pp.Opciones) != 1 || len(pp.Variantes) != 2 || pp.Stock != 6 {
		t.Fatalf("producto público: %+v", pp)
	}
	if pp.Variantes[0].Valores["Talle"] != "M" || pp.Variantes[0].Precio != nil || *pp.Variantes[1].Precio != 26000 {
		t.Fatalf("variantes públicas: %+v", pp.Variantes)
	}
	if otra, _ := vidriera.Publica(ctx, b.App, "sol"); len(otra.Productos) != 0 {
		t.Fatal("otra tienda ve los productos")
	}

	// Comprar por variante: el precio de la variante (o el del producto) y descuenta esa variante.
	lunes := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)
	c := compra(p.ID, 1, "transferencia")
	c.Envio = "retiro"
	c.Carrito[0].VarianteID = id["L"]
	c.Carrito = append(c.Carrito, vidriera.LineaCarrito{ID: p.ID, VarianteID: id["M"], Cantidad: 2})
	r, err := vidriera.Comprar(ctx, b.App, nil, "", "luna", c, lunes)
	if err != nil {
		t.Fatal(err)
	}
	if want := int64(26000+48000) * 85 / 100; r.Total != want {
		t.Fatalf("total: %d, quería %d", r.Total, want)
	}
	tienda, _ = vidriera.Publica(ctx, b.App, "luna")
	if pp = tienda.Productos[0]; pp.Stock != 3 || pp.Variantes[1].Stock != 0 {
		t.Fatalf("después de comprar: %+v", pp)
	}

	// Sin variante, con una inactiva, o la de otro producto: no se arma el pedido.
	for nombre, v := range map[string]string{"sin variante": "", "inactiva": id["XL"], "inventada": "x"} {
		c := compra(p.ID, 1, "transferencia")
		c.Envio, c.Carrito[0].VarianteID = "retiro", v
		if _, err := vidriera.Comprar(ctx, b.App, nil, "", "luna", c, lunes); !errors.Is(err, vidriera.ErrInvalido) {
			t.Errorf("%s: %v", nombre, err)
		}
	}
	// La última L ya se vendió.
	c = compra(p.ID, 1, "transferencia")
	c.Envio, c.Carrito[0].VarianteID = "retiro", id["L"]
	if _, err := vidriera.Comprar(ctx, b.App, nil, "", "luna", c, lunes); !errors.Is(err, vidriera.ErrSinStock) {
		t.Fatalf("sin stock: %v", err)
	}
}

// El link de cada producto sale del nombre (sin tildes ni signos), no se repite en la tienda y no
// cambia si el producto se renombra; otra tienda puede usar el mismo.
func TestSlugDeProductos(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	sol := pruebadb.CrearTienda(t, b.App, "sol")
	stock := 1
	nuevo := func(tienda, nombre string) listas.Producto {
		t.Helper()
		p, err := listas.Guardar(ctx, b.App, tienda, "", listas.Editable{Nombre: nombre, Precio: 1000, Stock: &stock})
		if err != nil {
			t.Fatal(err)
		}
		return p
	}
	primero := nuevo(luna, "Mate de Calabaza & Bombilla ¡Ñandú!")
	nuevo(luna, "Mate de calabaza y bombilla ñandú")
	nuevo(sol, "Mate de calabaza y bombilla ñandú")
	if _, err := listas.Guardar(ctx, b.App, luna, primero.ID, listas.Editable{Nombre: "Otro nombre", Precio: 1000, Stock: &stock}); err != nil {
		t.Fatal(err)
	}
	slugs := func(tienda string) map[string]bool {
		t.Helper()
		ti, err := vidriera.Publica(ctx, b.App, tienda)
		if err != nil {
			t.Fatal(err)
		}
		out := map[string]bool{}
		for _, p := range ti.Productos {
			out[p.Slug] = true
		}
		return out
	}
	if s := slugs("luna"); !s["mate-de-calabaza-y-bombilla-nandu"] || !s["mate-de-calabaza-y-bombilla-nandu-2"] || len(s) != 2 {
		t.Fatalf("luna: %v", s)
	}
	if s := slugs("sol"); !s["mate-de-calabaza-y-bombilla-nandu"] {
		t.Fatalf("sol: %v", s)
	}
}

// Un dominio propio muestra su tienda sólo verificado ("www." incluido); otro, o uno sin
// verificar, no muestra ninguna.
func TestSlugDeDominio(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	for _, d := range []string{"lunamates.com.ar", "pendiente.com.ar"} {
		if err := configuracion.ConectarDominio(ctx, b.App, luna, d); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := vidriera.SlugDeDominio(ctx, b.App, "lunamates.com.ar"); !errors.Is(err, vidriera.ErrNoExiste) {
		t.Fatalf("sin verificar: %v", err)
	}
	if err := b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE dominios SET verificado_en = now() WHERE nombre = 'lunamates.com.ar'`)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	for _, d := range []string{"lunamates.com.ar", "www.lunamates.com.ar", "LunaMates.com.ar"} {
		if s, err := vidriera.SlugDeDominio(ctx, b.App, d); err != nil || s != "luna" {
			t.Fatalf("%s: %q %v", d, s, err)
		}
	}
	// Elegido como principal, la tienda pública lo trae (para redirigir y para el canonical).
	if ti, err := vidriera.Publica(ctx, b.App, "luna"); err != nil || ti.Dominio != nil {
		t.Fatalf("sin principal: %v %v", ti.Dominio, err)
	}
	if err := configuracion.DominioPrincipal(ctx, b.App, luna, "lunamates.com.ar", "https://tulocalweb.com.ar"); err != nil {
		t.Fatal(err)
	}
	if ti, err := vidriera.Publica(ctx, b.App, "luna"); err != nil || ti.Dominio == nil || *ti.Dominio != "lunamates.com.ar" {
		t.Fatalf("principal: %v %v", ti.Dominio, err)
	}
	for _, d := range []string{"pendiente.com.ar", "otro.com.ar", "", "../x"} {
		if _, err := vidriera.SlugDeDominio(ctx, b.App, d); !errors.Is(err, vidriera.ErrNoExiste) {
			t.Fatalf("%q: %v", d, err)
		}
	}
}
