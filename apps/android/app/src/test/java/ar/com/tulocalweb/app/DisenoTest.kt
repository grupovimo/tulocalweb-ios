package ar.com.tulocalweb.app

import ar.com.tulocalweb.app.nucleo.Api
import ar.com.tulocalweb.app.nucleo.Bloque
import ar.com.tulocalweb.app.nucleo.CatalogoBloques
import ar.com.tulocalweb.app.nucleo.ContextoTienda
import ar.com.tulocalweb.app.nucleo.Diseno
import ar.com.tulocalweb.app.nucleo.FaltaBloque
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
}
