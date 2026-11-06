package ar.com.tulocalweb.app

import ar.com.tulocalweb.app.app.Ruta
import ar.com.tulocalweb.app.nucleo.Api
import ar.com.tulocalweb.app.nucleo.Clerk
import ar.com.tulocalweb.app.nucleo.Fecha
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.Marcas
import ar.com.tulocalweb.app.nucleo.PedidoFila
import ar.com.tulocalweb.app.pantallas.ingreso.mensajeClave
import ar.com.tulocalweb.app.pantallas.pedidos.urlWhatsApp
import ar.com.tulocalweb.app.tema.ChipEstado
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner

// Clon de TuLocalWebTests/TuLocalWebTests.swift: los mismos casos, con los mismos nombres de grupo.
@RunWith(RobolectricTestRunner::class)
class NucleoTest {
    // FormatoTests
    @Test fun pesosConEspacioFinoYPuntoDeMiles() {
        assertEquals("$ 184.350", Formato.pesos(184350))
        assertEquals("−$ 2.400", Formato.pesos(-2400))
        assertEquals("$ 0", Formato.pesos(0))
        assertEquals("$ 3.000.000", Formato.pesos(3_000_000))
    }

    @Test fun porcentajeConComa() {
        assertEquals("12,4 %", Formato.porcentaje(12.4))
        assertEquals("73 %", Formato.porcentaje(73.0))
    }

    @Test fun iniciales() {
        assertEquals("AL", Formato.iniciales("Almacén Lumbre"))
        assertEquals("M", Formato.iniciales("martina"))
        assertEquals("TL", Formato.iniciales(""))
    }

    @Test fun sinViudasPegaLaUltimaPalabraYLasCortas() {
        val s = Formato.sinViudas("Después cambiás de tienda desde Más, tocando el nombre.")
        assertTrue(s.endsWith("el nombre."))
        assertTrue(s.contains("de tienda"))
    }

    @Test fun fechasDeGo() {
        assertNotNull(Fecha.iso("2026-10-31T14:32:05.123456789-03:00"))
        assertNotNull(Fecha.iso("2026-10-31T14:32:05Z"))
        assertNotNull(Fecha.iso("2026-10-31"))
        assertNull(Fecha.iso("ayer"))
    }

    // RutaTests
    @Test fun rutasDelPushYDeLosLinks() {
        assertEquals(Ruta.pedido(1048), Ruta.de("pedido/1048"))
        assertEquals(Ruta.pedido(1048), Ruta.de("tulocalweb://pedido/1048"))
        assertEquals(Ruta.producto("abc"), Ruta.de("tulocalweb://producto/abc"))
        assertNull(Ruta.de("pedido/x"))
        assertNull(Ruta.de("cuenta/seguridad"))
    }

    // ClerkTests
    @Test fun laFrontendAPISaleDeLaClavePublicable() {
        // base64("clerk.grupovimo.com.ar$")
        assertEquals("https://clerk.grupovimo.com.ar", Clerk.fapiDe("pk_live_Y2xlcmsuZ3J1cG92aW1vLmNvbS5hciQ"))
        assertNull(Clerk.fapiDe("sk_live_xxx"))
        assertNull(Clerk.fapiDe(""))
        assertNull(Clerk.fapiDe(null))
    }

    // PedidosTests
    @Test fun elSiguienteEstadoDependeDelEnvio() {
        assertEquals("Listo para despachar", Marcas.siguiente("Para preparar", "andreani")?.boton)
        assertEquals("Listo para retirar", Marcas.siguiente("Para preparar", "retiro")?.boton)
        assertEquals("Entregado", Marcas.siguiente("Listo para despachar", "retiro")?.estado)
        assertEquals("En camino", Marcas.siguiente("Listo para despachar", "moto")?.estado)
        assertNull(Marcas.siguiente("Entregado", "moto"))
        assertEquals("Listo para despachar", Marcas.anterior("En camino", "moto"))
        assertEquals("Listo para despachar", Marcas.anterior("Entregado", "retiro"))
    }

    @Test fun chipsDeRetiro() {
        assertEquals("Listo para retirar", ChipEstado.texto("Listo para despachar", "retiro"))
        assertEquals("Retirado", ChipEstado.texto("Entregado", "retiro"))
        assertEquals("Entregado", ChipEstado.texto("Entregado", "andreani"))
    }

    @Test fun mensajeDeIntentos() {
        assertTrue(mensajeClave(2).endsWith("Te quedan 2 intentos antes del bloqueo."))
        assertTrue(mensajeClave(1).contains("Te queda 1 intento"))
        assertTrue(mensajeClave(0).contains("pausamos el ingreso"))
    }

    @Test fun whatsAppArgentino() {
        assertEquals("https://wa.me/5491158324410", urlWhatsApp("11 5832-4410"))
        assertEquals("https://wa.me/5491158324410", urlWhatsApp("+54 9 11 5832-4410"))
        assertNull(urlWhatsApp("123"))
    }

    @Test fun decodificaUnPedidoDeLaAPI() {
        val json = """
        {"numero":1048,"cliente":"Martina Gómez","foto":null,"ciudad":"CABA","canal":"web","pago":"transf","envio":"moto",
         "estado":"Para preparar","total":24020,"creado":"2026-10-30T14:32:00.123456-03:00",
         "productos":[{"nombre":"Yerba","foto":null}],"cantidadProductos":5,"vip":false,"hornero":false,"mayorista":false,
         "telefono":"+54 9 11 5832-4410","codigo":"abc"}
        """
        val p = Api.json.decodeFromString(PedidoFila.serializer(), json)
        assertEquals(1048, p.numero)
        assertEquals(1, p.productos.size)
    }
}
