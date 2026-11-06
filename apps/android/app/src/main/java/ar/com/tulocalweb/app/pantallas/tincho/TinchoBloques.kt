package ar.com.tulocalweb.app.pantallas.tincho

import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ColumnScope
import androidx.compose.foundation.layout.ExperimentalLayoutApi
import androidx.compose.foundation.layout.FlowRow
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Path
import androidx.compose.ui.graphics.PathEffect
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.drawText
import androidx.compose.ui.text.rememberTextMeasurer
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.ValorJSON
import ar.com.tulocalweb.app.tema.*
import kotlinx.serialization.KSerializer
import kotlinx.serialization.Serializable
import kotlinx.serialization.descriptors.SerialDescriptor
import kotlinx.serialization.encoding.Decoder
import kotlinx.serialization.encoding.Encoder
import kotlinx.serialization.json.JsonDecoder
import kotlinx.serialization.json.JsonElement
import kotlinx.serialization.json.JsonNull
import kotlinx.serialization.json.decodeFromJsonElement
import kotlinx.serialization.json.jsonObject
import kotlinx.serialization.json.jsonPrimitive
import kotlinx.serialization.json.contentOrNull
import kotlin.math.max
import kotlin.math.roundToInt

// Clon de Pantallas/Tincho/TinchoBloques.swift.

/** Lo que responde Tincho (POST /panel/hornero con formato «bloques»): el texto y, cuando corresponde, cifras, gráficos,
 *  tablas, listas, una barra de progreso, botones y sugerencias. Lo arma el motor de la web
 *  (apps/web/src/lib/server/hornero/motor/respuesta.ts). */
@Serializable
data class RespuestaTincho(
    val texto: String = "",
    val bloques: List<@Serializable(with = BloqueSerializer::class) BloqueTincho>? = null,
    val acciones: List<AccionTincho>? = null,
    val sugerencias: List<String>? = null,
)

@Serializable data class AccionTincho(val texto: String, val ir: String)

sealed class BloqueTincho {
    @Serializable data class Cifra(val titulo: String = "", val valor: String = "", val detalle: String? = null, val tono: String? = null)
    @Serializable data class Barra(val etiqueta: String = "", val valor: Double = 0.0, val destacado: Boolean? = null)
    @Serializable data class Punto(val etiqueta: String = "", val valor: Double = 0.0)
    @Serializable data class Item(val titulo: String = "", val detalle: String? = null, val insignia: String? = null, val tono: String? = null)

    data class cifras(val items: List<Cifra>) : BloqueTincho()
    data class barras(val titulo: String, val unidad: String, val items: List<Barra>) : BloqueTincho()
    data class linea(val titulo: String, val unidad: String, val puntos: List<Punto>, val anterior: List<Double>?) : BloqueTincho()
    data class tabla(val titulo: String, val columnas: List<String>, val filas: List<List<String>>) : BloqueTincho()
    data class lista(val titulo: String, val items: List<Item>) : BloqueTincho()
    data class progreso(val titulo: String, val actual: Double, val objetivo: Double, val unidad: String, val detalle: String?) : BloqueTincho()
    data object otro : BloqueTincho()
}

/** BloqueTincho desde el JSON según su «tipo» (lo que no se conoce, «otro»). */
object BloqueSerializer : KSerializer<BloqueTincho> {
    override val descriptor: SerialDescriptor = JsonElement.serializer().descriptor
    override fun serialize(encoder: Encoder, value: BloqueTincho) = encoder.encodeSerializableValue(JsonElement.serializer(), JsonNull)
    override fun deserialize(decoder: Decoder): BloqueTincho {
        val j = (decoder as JsonDecoder).json
        val o = decoder.decodeJsonElement().jsonObject
        fun s(k: String) = o[k]?.jsonPrimitive?.contentOrNull
        val titulo = s("titulo") ?: ""
        val unidad = s("unidad") ?: "numero"
        return runCatching {
            when (s("tipo")) {
                "cifras" -> BloqueTincho.cifras(j.decodeFromJsonElement(o["items"]!!))
                "barras" -> BloqueTincho.barras(titulo, unidad, j.decodeFromJsonElement(o["items"]!!))
                "linea" -> BloqueTincho.linea(titulo, unidad, j.decodeFromJsonElement(o["puntos"]!!), o["anterior"]?.let { runCatching { j.decodeFromJsonElement<List<Double>>(it) }.getOrNull() })
                "tabla" -> BloqueTincho.tabla(titulo, j.decodeFromJsonElement(o["columnas"]!!), j.decodeFromJsonElement(o["filas"]!!))
                "lista" -> BloqueTincho.lista(titulo, j.decodeFromJsonElement(o["items"]!!))
                "progreso" -> BloqueTincho.progreso(titulo, s("actual")!!.toDouble(), s("objetivo")!!.toDouble(), unidad, s("detalle"))
                else -> BloqueTincho.otro
            }
        }.getOrDefault(BloqueTincho.otro)
    }
}

// MARK: - Dibujo

/** Un bloque de la respuesta, en una tarjeta. */
@Composable
fun VistaBloque(bloque: BloqueTincho) {
    when (bloque) {
        is BloqueTincho.cifras -> Cifras(bloque.items)
        is BloqueTincho.barras -> TarjetaBloque(bloque.titulo) { Barras(bloque.unidad, bloque.items) }
        is BloqueTincho.linea -> TarjetaBloque(bloque.titulo) { Linea(bloque.unidad, bloque.puntos, bloque.anterior) }
        is BloqueTincho.tabla -> TarjetaBloque(bloque.titulo) { Tabla(bloque.columnas, bloque.filas) }
        is BloqueTincho.lista -> TarjetaBloque(bloque.titulo) { Lista(bloque.items) }
        is BloqueTincho.progreso -> TarjetaBloque(bloque.titulo) { Progreso(bloque.actual, bloque.objetivo, bloque.unidad, bloque.detalle) }
        BloqueTincho.otro -> {}
    }
}

@Composable
private fun TarjetaBloque(titulo: String, contenido: @Composable ColumnScope.() -> Unit) {
    val forma = RoundedCornerShape(18.dp)
    Column(Modifier.fillMaxWidth().clip(forma).background(C.superficie).border(1.dp, C.anillo, forma).padding(14.dp), verticalArrangement = Arrangement.spacedBy(10.dp)) {
        if (titulo.isNotEmpty()) Texto(titulo, Estilo.footnoteFuerte, C.texto2)
        contenido()
    }
}

@Composable
private fun Cifras(items: List<BloqueTincho.Cifra>) {
    Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
        items.chunked(2).forEach { par ->
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                par.forEach { c ->
                    val forma = RoundedCornerShape(14.dp)
                    Column(Modifier.weight(1f).clip(forma).background(C.superficie).border(1.dp, C.anillo, forma).padding(12.dp), verticalArrangement = Arrangement.spacedBy(3.dp)) {
                        Texto(c.titulo, Estilo.footnote, C.texto2, lineas = 1)
                        Texto(c.valor, Estilo.headline, C.texto, lineas = 1, numeros = true, escalaMinima = 0.6f)
                        c.detalle?.let { Texto(it, Estilo.footnoteFuerte, tintaTono(c.tono)) }
                    }
                }
                if (par.size == 1) Spacer(Modifier.weight(1f))
            }
        }
    }
}

@Composable
private fun Barras(unidad: String, items: List<BloqueTincho.Barra>) {
    val tope = (items.maxOfOrNull { it.valor } ?: 1.0).coerceAtLeast(1e-9) * 1.25
    Column(Modifier.semantics { contentDescription = items.joinToString(", ") { "${it.etiqueta}: ${corto(it.valor, unidad)}" } }, verticalArrangement = Arrangement.spacedBy(6.dp)) {
        items.forEach { b ->
            Row(Modifier.height(24.dp), verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                Texto(b.etiqueta, Estilo(11f, Peso.regular), C.texto2, Modifier.width(84.dp), lineas = 1)
                BoxWithConstraints(Modifier.weight(1f)) {
                    Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(4.dp)) {
                        Box(Modifier.width(this@BoxWithConstraints.maxWidth * (b.valor / tope).toFloat()).height(20.dp).clip(RoundedCornerShape(5.dp))
                            .background(if (b.destacado == true) C.acento else C.acento.copy(alpha = 0.45f)))
                        Texto(corto(b.valor, unidad), Estilo(11f, Peso.semibold), C.texto2, lineas = 1)
                    }
                }
            }
        }
    }
}

@Composable
private fun Linea(unidad: String, puntos: List<BloqueTincho.Punto>, anterior: List<Double>?) {
    val medidor = rememberTextMeasurer()
    val acento = C.acento; val gris = C.texto3.copy(alpha = 0.7f); val linea = C.linea; val texto2 = C.texto2; val texto3 = C.texto3
    val tope = max(puntos.maxOfOrNull { it.valor } ?: 0.0, anterior?.maxOrNull() ?: 0.0).coerceAtLeast(1.0)
    Canvas(Modifier.fillMaxWidth().height(150.dp).semantics { contentDescription = puntos.joinToString(", ") { "${it.etiqueta}: ${corto(it.valor, unidad)}" } }) {
        val iz = 40.dp.toPx(); val ab = 16.dp.toPx()
        val w = size.width - iz; val h = size.height - ab
        fun x(i: Int) = iz + if (puntos.size <= 1) w / 2 else w * i / (puntos.size - 1)
        fun y(v: Double) = (h * (1 - v / tope)).toFloat()
        listOf(0.0, tope / 2, tope).forEach { m ->
            drawLine(linea, Offset(iz, y(m)), Offset(size.width, y(m)), 1f)
            val r = medidor.measure(corto(m, unidad), Estilo(11f, Peso.regular).ts(true).copy(color = texto3))
            drawText(r, topLeft = Offset(0f, y(m) - r.size.height / 2))
        }
        if (anterior != null && anterior.size == puntos.size && anterior.size > 1) {
            val p = Path().apply { anterior.forEachIndexed { i, v -> if (i == 0) moveTo(x(i), y(v)) else lineTo(x(i), y(v)) } }
            drawPath(p, gris, style = Stroke(1.5.dp.toPx(), pathEffect = PathEffect.dashPathEffect(floatArrayOf(4.dp.toPx(), 4.dp.toPx()))))
        }
        if (puntos.size > 1) {
            val p = Path().apply { puntos.forEachIndexed { i, v -> if (i == 0) moveTo(x(i), y(v.valor)) else lineTo(x(i), y(v.valor)) } }
            val area = Path().apply { addPath(p); lineTo(x(puntos.size - 1), h); lineTo(x(0), h); close() }
            drawPath(area, Brush.verticalGradient(listOf(acento.copy(alpha = 0.25f), acento.copy(alpha = 0f))))
            drawPath(p, acento, style = Stroke(2.5.dp.toPx()))
        }
        val cada = max(1, (puntos.size + 5) / 6)
        puntos.forEachIndexed { i, pt ->
            if (i % cada == 0) {
                val r = medidor.measure(pt.etiqueta, Estilo(11f, Peso.regular).ts().copy(color = texto2))
                drawText(r, topLeft = Offset(x(i) - r.size.width / 2, h + 2.dp.toPx()))
            }
        }
    }
}

@Composable
private fun Tabla(columnas: List<String>, filas: List<List<String>>) {
    Row(Modifier.fillMaxWidth().horizontalScroll(rememberScrollState()), horizontalArrangement = Arrangement.spacedBy(14.dp)) {
        columnas.forEachIndexed { i, c ->
            Column(horizontalAlignment = if (i == 0) Alignment.Start else Alignment.End, verticalArrangement = Arrangement.spacedBy(8.dp)) {
                Texto(c, Estilo.footnoteFuerte, C.texto2)
                Box(Modifier.width(1.dp).height(1.dp))
                filas.forEach { f ->
                    val v = f.getOrNull(i) ?: ""
                    Texto(v, if (i == 0) Estilo.fila else Estilo.texto14, if (v == "Sin stock") C.errorTinta else C.texto,
                        if (i == 0) Modifier.widthIn(max = 170.dp) else Modifier, lineas = 2, numeros = true, alinear = if (i == 0) TextAlign.Start else TextAlign.End)
                }
            }
        }
    }
}

@Composable
private fun Lista(items: List<BloqueTincho.Item>) {
    Column {
        items.forEachIndexed { i, it ->
            if (i > 0) Separador(0.dp)
            Row(Modifier.padding(vertical = 9.dp), horizontalArrangement = Arrangement.spacedBy(10.dp), verticalAlignment = Alignment.CenterVertically) {
                Box(Modifier.size(7.dp).clip(CircleShape).background(tintaTono(it.tono)))
                Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                    Texto(it.titulo, Estilo.fila, C.texto)
                    it.detalle?.let { d -> Texto(d, Estilo.footnote, C.texto2) }
                }
                it.insignia?.let { s ->
                    Texto(s, Estilo.footnoteFuerte, tintaTono(it.tono), Modifier.clip(Capsula).background(fondoTono(it.tono)).padding(horizontal = 8.dp, vertical = 4.dp))
                }
            }
        }
    }
}

@Composable
private fun Progreso(actual: Double, objetivo: Double, unidad: String, detalle: String?) {
    val f = if (objetivo > 0) minOf(1.0, actual / objetivo) else 0.0
    Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
        Row(verticalAlignment = Alignment.Bottom, horizontalArrangement = Arrangement.spacedBy(6.dp)) {
            Texto(largo(actual, unidad), Estilo.headline, C.texto, numeros = true)
            Texto("de ${largo(objetivo, unidad)}", Estilo.footnote, C.texto2, Modifier.weight(1f))
            Texto("${(f * 100).roundToInt()} %", Estilo.footnoteFuerte, C.primarioTexto)
        }
        BoxWithConstraints(Modifier.fillMaxWidth().height(8.dp).clip(Capsula).background(C.pista)) {
            Box(Modifier.width(maxWidth * f.toFloat()).fillMaxHeight().clip(Capsula).background(C.acento))
        }
        detalle?.let { Texto(it, Estilo.footnote, C.texto2) }
    }
}

// MARK: Formato

@Composable
fun tintaTono(tono: String?): Color = when (tono) { "ok" -> C.okTinta; "mal" -> C.errorTinta; "aviso" -> C.avisoTinta; else -> C.texto2 }

@Composable
fun fondoTono(tono: String?): Color = when (tono) { "ok" -> C.okFondo; "mal" -> C.errorFondo; "aviso" -> C.avisoFondo; else -> C.pista }

/** Para ejes y etiquetas: «$ 1,2 M», «$ 340 mil», «12». */
fun corto(v: Double, unidad: String): String {
    if (unidad != "pesos") return Formato.numero(v.roundToInt())
    if (v >= 1_000_000) return "$ " + java.text.DecimalFormat("#,##0.#", java.text.DecimalFormatSymbols(Formato.argentina)).format(v / 1_000_000) + " M"
    if (v >= 10_000) return "$ ${(v / 1000).roundToInt()} mil"
    return Formato.pesos(v.roundToInt())
}

fun largo(v: Double, unidad: String): String = if (unidad == "pesos") Formato.pesos(v.roundToInt()) else Formato.numero(v.roundToInt())

/** Píldoras que bajan de renglón cuando no entran (las sugerencias de Tincho). */
@OptIn(ExperimentalLayoutApi::class)
@Composable
fun Fluir(espacio: Dp = 8.dp, contenido: @Composable () -> Unit) {
    FlowRow(horizontalArrangement = Arrangement.spacedBy(espacio), verticalArrangement = Arrangement.spacedBy(espacio)) { contenido() }
}
