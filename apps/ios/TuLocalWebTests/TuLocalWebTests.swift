import XCTest
@testable import TuLocalWeb

final class FormatoTests: XCTestCase {
    func testPesosConEspacioFinoYPuntoDeMiles() {
        XCTAssertEqual(Formato.pesos(184350), "$\u{202F}184.350")
        XCTAssertEqual(Formato.pesos(-2400), "−$\u{202F}2.400")
        XCTAssertEqual(Formato.pesos(0), "$\u{202F}0")
        XCTAssertEqual(Formato.pesos(3_000_000), "$\u{202F}3.000.000")
    }

    func testPorcentajeConComa() {
        XCTAssertEqual(Formato.porcentaje(12.4), "12,4\u{202F}%")
        XCTAssertEqual(Formato.porcentaje(73), "73\u{202F}%")
    }

    func testIniciales() {
        XCTAssertEqual(Formato.iniciales("Almacén Lumbre"), "AL")
        XCTAssertEqual(Formato.iniciales("martina"), "M")
        XCTAssertEqual(Formato.iniciales(""), "TL")
    }

    func testSinViudasPegaLaUltimaPalabraYLasCortas() {
        let s = Formato.sinViudas("Después cambiás de tienda desde Más, tocando el nombre.")
        XCTAssertTrue(s.hasSuffix("el\u{00A0}nombre."))
        XCTAssertTrue(s.contains("de\u{00A0}tienda"))
    }

    func testFechasDeGo() {
        XCTAssertNotNil(Fecha.iso("2026-10-31T14:32:05.123456789-03:00"))
        XCTAssertNotNil(Fecha.iso("2026-10-31T14:32:05Z"))
        XCTAssertNotNil(Fecha.iso("2026-10-31"))
        XCTAssertNil(Fecha.iso("ayer"))
    }
}

final class RutaTests: XCTestCase {
    func testRutasDelPushYDeLosLinks() {
        XCTAssertEqual(Ruta(texto: "pedido/1048"), .pedido(1048))
        XCTAssertEqual(Ruta(texto: "tulocalweb://pedido/1048"), .pedido(1048))
        XCTAssertEqual(Ruta(texto: "tulocalweb://producto/abc"), .producto("abc"))
        XCTAssertNil(Ruta(texto: "pedido/x"))
        // Los push de seguridad y de preguntas de Mercado Libre abren su pantalla; lo desconocido, nada.
        XCTAssertEqual(Ruta(texto: "cuenta/seguridad"), .destino(.cuenta))
        XCTAssertEqual(Ruta(texto: "preguntas/ml/q1"), .destino(.resenas))
        XCTAssertNil(Ruta(texto: "cualquiera/cosa"))
    }
}

final class ClerkTests: XCTestCase {
    func testLaFrontendAPISaleDeLaClavePublicable() {
        // base64("clerk.grupovimo.com.ar$")
        XCTAssertEqual(Clerk.fapiDe("pk_live_Y2xlcmsuZ3J1cG92aW1vLmNvbS5hciQ")?.absoluteString, "https://clerk.grupovimo.com.ar")
        XCTAssertNil(Clerk.fapiDe("sk_live_xxx"))
        XCTAssertNil(Clerk.fapiDe(""))
        XCTAssertNil(Clerk.fapiDe(nil))
    }
}

final class PedidosTests: XCTestCase {
    func testElSiguienteEstadoDependeDelEnvio() {
        XCTAssertEqual(Marcas.siguiente("Para preparar", envio: "andreani")?.boton, "Listo para despachar")
        XCTAssertEqual(Marcas.siguiente("Para preparar", envio: "retiro")?.boton, "Listo para retirar")
        XCTAssertEqual(Marcas.siguiente("Listo para despachar", envio: "retiro")?.estado, "Entregado")
        XCTAssertEqual(Marcas.siguiente("Listo para despachar", envio: "moto")?.estado, "En camino")
        XCTAssertNil(Marcas.siguiente("Entregado", envio: "moto"))
        XCTAssertEqual(Marcas.anterior("En camino", envio: "moto"), "Listo para despachar")
        XCTAssertEqual(Marcas.anterior("Entregado", envio: "retiro"), "Listo para despachar")
    }

    func testChipsDeRetiro() {
        XCTAssertEqual(ChipEstado.texto("Listo para despachar", envio: "retiro"), "Listo para retirar")
        XCTAssertEqual(ChipEstado.texto("Entregado", envio: "retiro"), "Retirado")
        XCTAssertEqual(ChipEstado.texto("Entregado", envio: "andreani"), "Entregado")
    }

    func testMensajeDeIntentos() {
        XCTAssertTrue(Ingresar.mensajeClave(restantes: 2).hasSuffix("Te quedan 2 intentos antes del bloqueo."))
        XCTAssertTrue(Ingresar.mensajeClave(restantes: 1).contains("Te queda 1 intento"))
        XCTAssertTrue(Ingresar.mensajeClave(restantes: 0).contains("pausamos el ingreso"))
    }

    func testWhatsAppArgentino() {
        XCTAssertEqual(urlWhatsApp("11 5832-4410")?.absoluteString, "https://wa.me/5491158324410")
        XCTAssertEqual(urlWhatsApp("+54 9 11 5832-4410")?.absoluteString, "https://wa.me/5491158324410")
        XCTAssertNil(urlWhatsApp("123"))
    }

    func testDecodificaUnPedidoDeLaAPI() throws {
        let json = """
        {"numero":1048,"cliente":"Martina Gómez","foto":null,"ciudad":"CABA","canal":"web","pago":"transf","envio":"moto",
         "estado":"Para preparar","total":24020,"creado":"2026-10-30T14:32:00.123456-03:00",
         "productos":[{"nombre":"Yerba","foto":null}],"cantidadProductos":5,"vip":false,"hornero":false,"mayorista":false,
         "telefono":"+54 9 11 5832-4410","codigo":"abc"}
        """
        let p = try Api.decodificador.decode(PedidoFila.self, from: Data(json.utf8))
        XCTAssertEqual(p.numero, 1048)
        XCTAssertEqual(p.productos.count, 1)
    }
}
