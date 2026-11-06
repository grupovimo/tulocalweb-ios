package ar.com.tulocalweb.app.pantallas.mas

import android.content.Context
import android.graphics.Paint
import android.graphics.Typeface
import android.graphics.pdf.PdfDocument
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.app.Destino
import ar.com.tulocalweb.app.app.LocalNavegacion
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.PedidoDetalle
import ar.com.tulocalweb.app.nucleo.Recursos
import ar.com.tulocalweb.app.nucleo.Tincho
import ar.com.tulocalweb.app.nucleo.ValorJSON
import ar.com.tulocalweb.app.nucleo.abrirUrl
import ar.com.tulocalweb.app.nucleo.archivoParaCompartir
import ar.com.tulocalweb.app.nucleo.lista
import ar.com.tulocalweb.app.nucleo.texto
import ar.com.tulocalweb.app.pantallas.buscar.normal
import ar.com.tulocalweb.app.tema.*
import kotlinx.serialization.Serializable
import java.io.File

// Clon de Pantallas/Mas/Ayuda.swift.

/** El centro de ayuda adentro de la app: los mismos artículos que tulocalweb.com.ar/ayuda, que vienen en assets/ayuda.json
 *  (se regeneran desde apps/web/src/lib/ayuda/articulos.ts). */
@Composable
fun CentroAyuda() {
    val nav = LocalNavegacion.current
    val ctx = LocalContext.current
    var busqueda by remember { mutableStateOf("") }
    val datos = remember { DatosAyuda.cargar() }

    @Composable
    fun Lista(arts: List<DatosAyuda.Articulo>) {
        Column(Modifier.fillMaxWidth().tarjeta()) {
            arts.forEachIndexed { i, a ->
                if (i > 0) Separador()
                Row(Modifier.fillMaxWidth().clickable { nav.abrir(Destino.articuloAyuda(a.slug)) }.padding(horizontal = 16.dp, vertical = 12.dp), verticalAlignment = Alignment.CenterVertically) {
                    Column(Modifier.weight(1f).padding(end = 6.dp), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                        Texto(a.titulo, Estilo.fila, C.texto)
                        Texto(a.resumen, Estilo.footnote, C.texto2, lineas = 2)
                    }
                    Chevron(8f)
                }
            }
        }
    }

    Column(Modifier.fillMaxSize().background(C.fondo).verticalScroll(rememberScrollState()).margenPantalla().padding(bottom = 24.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
        Texto("Ayuda", Estilo.tituloGrande, C.texto)
        Buscador("Buscá: envíos, cobrar, factura…", busqueda, { busqueda = it })
        if (busqueda.trim().isEmpty()) {
            datos.categorias.forEach { c ->
                val arts = datos.articulos.filter { it.categoria == c.slug }
                if (arts.isNotEmpty()) {
                    EtiquetaSeccion(c.nombre, Modifier.padding(horizontal = 4.dp).padding(top = 8.dp))
                    Lista(arts)
                }
            }
        } else {
            val q = normal(busqueda)
            val arts = datos.articulos.filter { normal(it.titulo + " " + it.resumen).contains(q) }
            if (arts.isEmpty()) Box(Modifier.padding(top = 30.dp)) { EstadoTincho(Tincho.lupa, "No encontramos «$busqueda»", "Probá con otra palabra o escribinos.") }
            else Lista(arts)
        }
        EtiquetaSeccion("¿No lo encontraste?", Modifier.padding(horizontal = 4.dp).padding(top = 10.dp))
        Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            AccionRapida("envelope", "Escribinos", Modifier.weight(1f), 58.dp) { abrirUrl(ctx, "mailto:hola@tulocalweb.com.ar?subject=Ayuda%20desde%20la%20app") }
        }
    }
}

@Composable
fun ArticuloAyuda(slug: String) {
    val a = remember(slug) { DatosAyuda.cargar().articulos.firstOrNull { it.slug == slug } } ?: return
    Column(Modifier.fillMaxSize().background(C.fondo).verticalScroll(rememberScrollState()).margenPantalla().padding(bottom = 32.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
        Texto(Formato.sinViudas(a.titulo), Estilo.tituloHoja, C.texto)
        Texto(Formato.sinViudas(a.resumen), Estilo.subhead, C.texto2)
        a.cuerpo.forEach { b ->
            when (b.t) {
                "h" -> Texto(b.texto, Estilo.headline, C.texto, Modifier.padding(top = 6.dp))
                "pasos" -> Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                    b.lista.forEachIndexed { i, x ->
                        Row(horizontalArrangement = Arrangement.spacedBy(10.dp), verticalAlignment = Alignment.Top) {
                            Box(Modifier.size(24.dp).clip(CircleShape).background(C.crema), contentAlignment = Alignment.Center) { Texto("${i + 1}", Estilo(13f, Peso.bold), C.primarioTinta) }
                            Texto(x, Estilo.body16, C.texto, Modifier.weight(1f))
                        }
                    }
                }
                "lista" -> Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
                    b.lista.forEach { x ->
                        Row(horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.Top) {
                            Texto("•", Estilo.body16, C.acento)
                            Texto(x, Estilo.body16, C.texto, Modifier.weight(1f))
                        }
                    }
                }
                "nota" -> Texto(b.texto, Estilo.texto14, C.avisoTinta, Modifier.fillMaxWidth().clip(RoundedCornerShape(12.dp)).background(C.avisoFondo).padding(12.dp))
                else -> Texto(b.texto, Estilo.body16, C.texto)
            }
        }
    }
}

@Serializable
data class DatosAyuda(val categorias: List<Categoria> = emptyList(), val articulos: List<Articulo> = emptyList()) {
    @Serializable data class Categoria(val slug: String, val nombre: String)
    @Serializable data class Bloque(val t: String, val x: ValorJSON? = null) {
        val texto: String get() = x?.texto ?: ""
        val lista: List<String> get() = x?.lista?.mapNotNull { it.texto } ?: emptyList()
    }
    @Serializable data class Articulo(val slug: String, val categoria: String = "", val titulo: String = "", val resumen: String = "", val cuerpo: List<Bloque> = emptyList())

    companion object {
        private var cache: DatosAyuda? = null
        fun cargar(): DatosAyuda = cache ?: (Recursos.leer<DatosAyuda>("ayuda.json") ?: DatosAyuda()).also { cache = it }
    }
}

/** La etiqueta de envío (10 × 15 cm) armada en el celular, para imprimir o compartir como PDF. */
object EtiquetaEnvio {
    /** El PDF de la etiqueta (283 × 425 puntos), en un archivo para compartir. */
    fun pdf(ctx: Context, p: PedidoDetalle, tienda: String): File? = runCatching {
        val doc = PdfDocument()
        val pagina = doc.startPage(PdfDocument.PageInfo.Builder(283, 425, 1).create())
        val c = pagina.canvas
        fun pintura(tam: Float, peso: Int = Typeface.NORMAL, gris: Boolean = false, mono: Boolean = false) = Paint(Paint.ANTI_ALIAS_FLAG).apply {
            textSize = tam; color = if (gris) android.graphics.Color.GRAY else android.graphics.Color.BLACK
            typeface = Typeface.create(if (mono) Typeface.MONOSPACE else Typeface.DEFAULT, peso)
        }
        var y = 24f
        fun linea(t: String, p: Paint, despues: Float = 4f) { y += -p.ascent(); c.drawText(t, 24f, y, p); y += p.descent() + despues }
        linea("Pedido #${p.numero}", pintura(30f, Typeface.BOLD), 14f)
        linea("PARA", pintura(12f, Typeface.BOLD, gris = true))
        linea(p.cliente, pintura(22f, Typeface.BOLD))
        p.entrega?.let { e ->
            linea(e.direccion, pintura(18f))
            linea(listOfNotNull(e.localidad, e.provincia, "CP ${e.cp}").joinToString(" · "), pintura(16f))
            e.sucursalNombre?.let { linea("Sucursal: $it", pintura(16f)) }
        }
        y += 10f
        c.drawLine(24f, y, 259f, y, Paint().apply { color = android.graphics.Color.LTGRAY; strokeWidth = 1f })
        y += 14f
        linea("DE", pintura(12f, Typeface.BOLD, gris = true))
        linea(tienda, pintura(16f, Typeface.BOLD))
        // El código abajo, en un recuadro.
        val cod = pintura(26f, Typeface.BOLD, mono = true)
        val t = "PED-${p.numero}"
        val alto = -cod.ascent() + cod.descent() + 20f
        val arriba = 425f - 24f - alto
        c.drawRect(24f, arriba, 259f, arriba + alto, Paint().apply { style = Paint.Style.STROKE; strokeWidth = 2f; color = android.graphics.Color.BLACK })
        c.drawText(t, (283f - cod.measureText(t)) / 2, arriba + 10f - cod.ascent(), cod)
        doc.finishPage(pagina)
        val f = archivoParaCompartir(ctx, "etiqueta-${p.numero}.pdf")
        f.outputStream().use { doc.writeTo(it) }
        doc.close()
        f
    }.getOrNull()
}
