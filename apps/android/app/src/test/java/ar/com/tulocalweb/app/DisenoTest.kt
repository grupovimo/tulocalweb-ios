package ar.com.tulocalweb.app

import ar.com.tulocalweb.app.nucleo.Api
import ar.com.tulocalweb.app.nucleo.Bloque
import ar.com.tulocalweb.app.nucleo.CatalogoBloques
import ar.com.tulocalweb.app.nucleo.ContextoTienda
import ar.com.tulocalweb.app.nucleo.Diseno
import ar.com.tulocalweb.app.nucleo.FaltaBloque
import ar.com.tulocalweb.app.nucleo.FaltaVentana
import ar.com.tulocalweb.app.nucleo.Fecha
import ar.com.tulocalweb.app.nucleo.Ventana
import ar.com.tulocalweb.app.nucleo.bool
import ar.com.tulocalweb.app.nucleo.get
import ar.com.tulocalweb.app.nucleo.lista
import ar.com.tulocalweb.app.pantallas.mas.EstiloTienda
import ar.com.tulocalweb.app.nucleo.PresetTienda
import ar.com.tulocalweb.app.nucleo.Valor
import ar.com.tulocalweb.app.nucleo.objeto
import ar.com.tulocalweb.app.nucleo.texto
import ar.com.tulocalweb.app.pantallas.mas.EditorTienda
import kotlinx.serialization.json.jsonArray
import kotlinx.serialization.json.jsonObject
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner
import org.robolectric.annotation.Config

// Clon de TuLocalWebTests/DisenoTests.swift: el diseño v2 de la tienda, los casos compartidos de la migración
// (packages/contrato/casos/diseno-v2.json), el catálogo de bloques y los cambios del editor.
@RunWith(RobolectricTestRunner::class)
@Config(sdk = [35])
class DisenoTest {
    @Test fun casosDeLaMigracion() {
        val raiz = Api.json.parseToJsonElement(javaClass.getResource("/diseno-v2.json")!!.readText()).jsonObject
        val casos = raiz["casos"]!!.jsonArray
        assertFalse(casos.isEmpty())
        for (c in casos) {
            val o = c.jsonObject
            val d = Diseno.desde(o["v1"]?.objeto ?: emptyMap(), o["plantilla"]!!.texto!!)
            assertEquals(o["nombre"]!!.texto, o["v2"], Valor.objeto(d.json))
        }
    }

    @Test fun catalogoYPresets() {
        val cat = CatalogoBloques.shared
        assertTrue(cat.bloques.size > 15)
        assertEquals(17, PresetTienda.todos.size)
        for (t in cat.bloques) {
            assertEquals(t.tipo, t.variantes.size, t.nombresVariantes.size)
            assertNotNull(t.tipo, cat.grupos.firstOrNull { it.id == t.grupo })
        }
        for ((id, p) in PresetTienda.todos) {
            for (b in p.inicio) assertTrue("$id ${b.tipo}", cat.tipo(b.tipo)?.variantes?.contains(b.variante) == true)
        }
    }

    @Test fun cambiarDePresetConservaLosDatos() {
        var d = Diseno.desde(emptyMap(), "ceibo")
        d = d.copy(inicio = d.inicio.mapIndexed { i, b -> if (i == 0) b.copy(datos = b.datos + ("titulo" to Valor.texto("Lumbre"))) else b } +
            Bloque("abc12345", "banner", "banda", datos = mapOf("titulo" to Valor.texto("Hot Sale"))))
        d = d.copy(estilo = d.estilo.copy(direccion = "pedido"))
        d = d.aplicarPreset("andes")
        assertEquals("andes", d.plantilla)
        assertNull(d.estilo.direccion)
        assertEquals(Valor.texto("Lumbre"), d.inicio.firstOrNull { it.tipo == "portada" }?.datos?.get("titulo"))
        assertEquals("abc12345", d.inicio.last().id)
        assertEquals(d.inicio.size, d.inicio.map { it.id }.toSet().size)
    }

    @Test fun editorDeshaceYRehace() {
        val e = EditorTienda()
        e.empezar(Diseno.desde(emptyMap(), "litoral"), "litoral")
        val antes = e.bloques.size
        val banner = CatalogoBloques.shared.tipo("banner")!!
        val id = e.agregar(banner, 1, abrir = false)
        assertEquals(antes + 1, e.bloques.size)
        assertEquals(id, e.bloques[1].id)
        e.duplicar(id)
        e.borrar(id)
        assertEquals(antes + 1, e.bloques.size)
        e.deshacer()
        e.deshacer()
        assertEquals(antes + 1, e.bloques.size)
        e.deshacer()
        assertEquals(antes, e.bloques.size)
        e.rehacer()
        assertEquals(id, e.bloques[1].id)
        e.mover(id, antesDe = null)
        assertEquals(id, e.bloques.last().id)
        assertTrue(e.sinGuardar)
    }

    @Test fun loQueLeFaltaAUnBloque() {
        var ctx = ContextoTienda(nombre = "Lumbre", productos = 3)
        assertEquals(true, FaltaBloque.de(Bloque("w", "whatsapp", "banda"), ctx)?.enNegocio)
        ctx = ctx.copy(whatsapp = true)
        assertNull(FaltaBloque.de(Bloque("w", "whatsapp", "banda"), ctx))
        assertNotNull(FaltaBloque.de(Bloque("b", "banner", "banda"), ctx))
        assertNull(FaltaBloque.de(Bloque("b", "banner", "banda", datos = mapOf("titulo" to Valor.texto("Hot Sale"))), ctx))
        assertNotNull(FaltaBloque.de(Bloque("p", "preguntas", "acordeon"), ctx))
    }

    // MARK: Ventanas (packages/contrato/casos/ventanas.json, los mismos que corren Go y la web)

    @Test fun ventanasCompletasComoGo() {
        val raiz = Api.json.parseToJsonElement(javaClass.getResource("/ventanas.json")!!.readText())
        val casos = raiz["ventanas"]?.lista ?: emptyList()
        assertFalse(casos.isEmpty())
        val ahora = Fecha.iso("2026-11-15T12:00:00-03:00")!!
        for (c in casos) {
            if (c["valida"]?.bool != true) continue
            val completa = c["completa"]?.bool ?: continue
            val v = c["ventana"] ?: continue
            val nombre = c["nombre"]?.texto ?: ""
            val ventana = Ventana.desde(v)
            assertEquals(nombre, completa, FaltaVentana.de(ventana, ahora) == null)
            // Leer y volver a escribir no cambia lo que Go valida.
            assertEquals(nombre, ventana.json, Ventana.desde(ventana.json).json)
        }
    }

    @Test fun plantillasDeVentanas() {
        val cat = CatalogoBloques.shared.ventanas!!
        assertEquals(7, cat.plantillas.size)
        val e = EditorTienda()
        e.empezar(Diseno.desde(emptyMap(), "litoral"), "litoral")
        for (p in cat.plantillas) {
            val v = Ventana.dePlantilla(p, e.idNuevo(), 50000)
            assertNotNull(p.id, cat.contenido(v.contenido))
            assertFalse(p.id, v.json.toString().contains("{envioGratis}"))
            e.agregarVentana(v, abrir = false)
        }
        assertEquals(7, e.diseno.ventanas.size)
        assertEquals(7, e.diseno.ventanas.map { it.id }.toSet().size)
        val id = e.diseno.ventanas[0].id
        e.duplicarVentana(id)
        assertEquals(8, e.diseno.ventanas.size)
        assertFalse(e.puedeSumarVentana)
        e.moverVentana(id, 1)
        assertEquals(id, e.diseno.ventanas[1].id)
        e.borrarVentana(id)
        e.deshacer()
        assertEquals(id, e.diseno.ventanas[1].id)
        val d = Diseno.desde(e.diseno.json, "litoral")
        assertEquals(e.diseno.ventanas, d.ventanas)
    }

    @Test fun conversionYBarraViajanEnElEstilo() {
        val d = Diseno.desde(emptyMap(), "litoral").let {
            it.copy(estilo = it.estilo.copy(conversion = mapOf("favoritos" to Valor.booleano(true), "stockBajo" to Valor.numero(5)),
                barra = EstiloTienda.Barra("fija", listOf(EstiloTienda.Barra.Mensaje("Hot Sale", hasta = "2026-11-20T23:59:00-03:00")))))
        }
        val otra = Diseno.desde(d.json, "litoral")
        assertEquals(Valor.booleano(true), otra.estilo.conversion?.get("favoritos"))
        assertEquals("2026-11-20T23:59:00-03:00", otra.estilo.barra?.mensajes?.firstOrNull()?.hasta)
    }
}
