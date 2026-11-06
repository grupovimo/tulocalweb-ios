package ar.com.tulocalweb.app.tema

import androidx.compose.animation.AnimatedContent
import androidx.compose.animation.core.Animatable
import androidx.compose.animation.core.AnimationSpec
import androidx.compose.animation.core.CubicBezierEasing
import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.tween
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.slideInVertically
import androidx.compose.animation.slideOutVertically
import androidx.compose.animation.togetherWith
import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.gestures.detectDragGestures
import androidx.compose.foundation.gestures.detectTapGestures
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.rotate
import androidx.compose.ui.draw.scale
import androidx.compose.ui.geometry.CornerRadius
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.geometry.Size
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Path
import androidx.compose.ui.graphics.PathEffect
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.graphics.drawscope.DrawScope
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.layout.onSizeChanged
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.semantics.stateDescription
import androidx.compose.ui.text.TextMeasurer
import androidx.compose.ui.text.drawText
import androidx.compose.ui.text.rememberTextMeasurer
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.IntOffset
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.Tablero
import kotlinx.coroutines.delay
import java.time.Instant
import kotlin.math.atan2
import kotlin.math.ceil
import kotlin.math.floor
import kotlin.math.log10
import kotlin.math.max
import kotlin.math.min
import kotlin.math.pow
import kotlin.math.roundToInt

// Clon de Tema/Graficos.swift: los mismos gráficos (en iOS con Swift Charts) dibujados con Canvas y el movimiento de
// la app. Todos crecen desde cero la primera vez, se tocan o se arrastran para ver cada valor (con un toque háptico) y
// respetan «Reducir movimiento».

// MARK: - Movimiento

object Mov {
    private val suave = CubicBezierEasing(0.25f, 0.1f, 0.25f, 1f)
    /** Entrada de tarjetas: suave y corta. */
    fun <T> entrada(): androidx.compose.animation.core.FiniteAnimationSpec<T> = tween(500, easing = suave)
    /** Gráficos y barras que crecen. */
    fun <T> crecer(): androidx.compose.animation.core.FiniteAnimationSpec<T> = tween(900, easing = suave)
    /** Cambios por un toque (elegir, filtrar). */
    fun <T> toque(): androidx.compose.animation.core.FiniteAnimationSpec<T> = tween(250, easing = CubicBezierEasing(0.2f, 0.8f, 0.2f, 1f))
}

/** Entra subiendo apenas y apareciendo, con un retraso para escalonar una grilla (una sola vez). */
@Composable
fun Modifier.aparecer(retraso: Double = 0.0): Modifier {
    if (!animacionesActivas()) return this
    var visto by rememberSaveable { mutableStateOf(false) }
    val t = remember { Animatable(if (visto) 1f else 0f) }
    LaunchedEffect(Unit) {
        if (!visto) { delay((retraso * 1000).toLong()); t.animateTo(1f, Mov.entrada()); visto = true }
    }
    return this.graphicsLayer { alpha = t.value; translationY = (1 - t.value) * 14.dp.toPx() }
}

/** 0 → 1 la primera vez que aparece (1 de entrada con «Reducir movimiento»). */
@Composable
fun crecer(retraso: Double = 0.1): Float {
    val anima = animacionesActivas()
    val a = remember { Animatable(if (anima) 0f else 1f) }
    LaunchedEffect(Unit) { if (a.value < 1f) { delay((retraso * 1000).toLong()); a.animateTo(1f, Mov.crecer()) } }
    return a.value
}

/** Un número que cambia contando (contentTransition(.numericText()) en iOS). */
@Composable
fun NumeroAnimado(texto: String, estilo: Estilo, color: Color, modifier: Modifier = Modifier, escalaMinima: Float? = null) {
    AnimatedContent(texto, modifier, transitionSpec = { (slideInVertically { it / 2 } + fadeIn()) togetherWith (slideOutVertically { -it / 2 } + fadeOut()) }, label = "numero") {
        Texto(it, estilo, color, lineas = 1, numeros = true, escalaMinima = escalaMinima)
    }
}

/** El globo con el valor del punto elegido. */
data class LineaGlobo(val etiqueta: String, val valor: String, val color: Color)

@Composable
fun GloboGrafico(titulo: String, lineas: List<LineaGlobo>, modifier: Modifier = Modifier) {
    val forma = RoundedCornerShape(10.dp)
    Column(
        modifier.sombra(forma, hex(0x2A140C, 0.12f), 8.dp, y = 3.dp).clip(forma).background(C.superficie).border(1.dp, C.anillo, forma)
            .padding(horizontal = 10.dp, vertical = 7.dp),
        verticalArrangement = Arrangement.spacedBy(2.dp),
    ) {
        if (titulo.isNotEmpty()) Texto(titulo, Estilo.chip, C.texto2)
        lineas.forEach { l ->
            Row(horizontalArrangement = Arrangement.spacedBy(5.dp), verticalAlignment = Alignment.CenterVertically) {
                if (l.etiqueta.isNotEmpty()) Texto(l.etiqueta, Estilo.chip, C.texto3)
                Texto(l.valor, Estilo.footnoteFuerte, l.color, numeros = true)
            }
        }
    }
}

/** El globo arriba del punto elegido, sin salirse del gráfico (overflowResolution .fit(to: .chart)). */
@Composable
private fun GloboEn(x: Float, anchoTotal: Float, contenido: @Composable () -> Unit) {
    var ancho by remember { mutableStateOf(0) }
    val densidad = LocalDensity.current
    Box(Modifier.onSizeChanged { ancho = it.width }.offset {
        IntOffset((x - ancho / 2f).coerceIn(0f, max(0f, anchoTotal - ancho)).roundToInt(), 0)
    }) { contenido() }
}

/** Elegir un punto tocando o arrastrando (chartXSelection), con un toque háptico al cambiar. */
private fun Modifier.seleccionX(n: Int, ancho: () -> Float, alElegir: (Int?) -> Unit): Modifier = this
    .pointerInput(n) {
        detectTapGestures { o -> alElegir(indiceEn(o.x, ancho(), n)) }
    }
    .pointerInput(n) {
        detectDragGestures(onDragEnd = { alElegir(null) }, onDragCancel = { alElegir(null) }) { c, _ -> alElegir(indiceEn(c.position.x, ancho(), n)) }
    }

private fun indiceEn(x: Float, ancho: Float, n: Int): Int? {
    if (n < 1 || ancho <= 0f) return null
    if (n == 1) return 0
    return ((x / ancho) * (n - 1)).roundToInt().coerceIn(0, n - 1)
}

/** Una curva suave por los puntos (interpolationMethod(.catmullRom)). */
private fun curva(puntos: List<Offset>): Path {
    val p = Path()
    if (puntos.isEmpty()) return p
    p.moveTo(puntos[0].x, puntos[0].y)
    for (i in 0 until puntos.size - 1) {
        val p0 = puntos[max(0, i - 1)]; val p1 = puntos[i]; val p2 = puntos[i + 1]; val p3 = puntos[min(puntos.size - 1, i + 2)]
        p.cubicTo(p1.x + (p2.x - p0.x) / 6f, p1.y + (p2.y - p0.y) / 6f, p2.x - (p3.x - p1.x) / 6f, p2.y - (p3.y - p1.y) / 6f, p2.x, p2.y)
    }
    return p
}

/** Valores «lindos» para un eje (AxisMarks automático). */
private fun marcasEje(minimo: Double, maximo: Double, cantidad: Int): List<Double> {
    val rango = maximo - minimo
    if (rango <= 0) return listOf(minimo)
    val crudo = rango / cantidad
    val potencia = 10.0.pow(floor(log10(crudo)))
    val paso = listOf(1.0, 2.0, 2.5, 5.0, 10.0).map { it * potencia }.first { it >= crudo }
    val desde = ceil(minimo / paso) * paso
    return generateSequence(desde) { it + paso }.takeWhile { it <= maximo + 1e-9 }.toList()
}

private fun DrawScope.textoEje(m: TextMeasurer, s: String, color: Color, x: Float, y: Float, centrado: Boolean = false, derecha: Boolean = false) {
    val r = m.measure(s, Estilo(11f, Peso.regular).ts(true).copy(color = color))
    val dx = when { centrado -> x - r.size.width / 2f; derecha -> x - r.size.width; else -> x }
    drawText(r, topLeft = Offset(dx, y - r.size.height / 2f))
}

// MARK: - Ventas

/** El gráfico de ventas de 2a: el período en `accent` (3 pt, relleno degradé 22 % → 0 y punto final de 9 con borde
 *  blanco de 3) y el anterior punteado (`#8A7364`, 2 pt, 4/5), con tres guías. Arrastrando se ve cuánto se vendió en
 *  cada punto y cuánto en el mismo punto del período anterior. */
@Composable
fun GraficoVentas(actual: List<Double>, anterior: List<Double>, alto: Dp = 130.dp, etiqueta: ((Int) -> String)? = null) {
    var elegido by remember { mutableStateOf<Int?>(null) }
    val factor = crecer()
    val f by animateFloatAsState(1f, Mov.crecer(), label = "datos")
    val tope = maxOf(actual.maxOrNull() ?: 0.0, anterior.maxOrNull() ?: 0.0, 1.0) * 1.08
    val largo = maxOf(actual.size, anterior.size, 2)
    val guia = tema(0x3A1D12, 0xFBF7F2, alfa = 0.07f).actual
    val acento = C.acento; val linea2 = C.linea2
    var ancho by remember { mutableStateOf(0f) }
    BoxWithConstraints(
        Modifier.fillMaxWidth().height(alto).onSizeChanged { ancho = it.width.toFloat() }
            .seleccionX(actual.size, { ancho }) { if (it != elegido) { elegido = it; if (it != null) Haptico.seleccion() } }
            .semantics { contentDescription = "Gráfico de ventas"; stateDescription = actual.lastOrNull()?.let { "Último punto: ${Formato.pesos(it)}" } ?: "" },
    ) {
        Canvas(Modifier.fillMaxWidth().height(alto)) {
            fun px(i: Int) = size.width * i / (largo - 1)
            fun py(v: Double) = size.height * (1 - (v * factor * f / tope).toFloat())
            listOf(0.15, 0.55, 0.95).forEach { g ->
                val y = size.height * (1 - (g / 1.08).toFloat())
                drawLine(guia, Offset(0f, y), Offset(size.width, y), 1.dp.toPx())
            }
            if (anterior.size > 1) {
                drawPath(curva(anterior.mapIndexed { i, v -> Offset(px(i), py(v)) }), hex(0x8A7364),
                    style = Stroke(2.dp.toPx(), cap = StrokeCap.Round, pathEffect = PathEffect.dashPathEffect(floatArrayOf(4.dp.toPx(), 5.dp.toPx()))))
            }
            if (actual.isNotEmpty()) {
                val pts = actual.mapIndexed { i, v -> Offset(px(i), py(v)) }
                val area = curva(pts).apply { lineTo(pts.last().x, size.height); lineTo(pts.first().x, size.height); close() }
                drawPath(area, Brush.verticalGradient(listOf(acento.copy(alpha = 0.22f), acento.copy(alpha = 0f))))
                drawPath(curva(pts), acento, style = Stroke(3.dp.toPx(), cap = StrokeCap.Round))
                val e = elegido
                val punto = if (e != null && e in actual.indices) {
                    drawLine(linea2, Offset(px(e), 0f), Offset(px(e), size.height), 1.dp.toPx(), pathEffect = PathEffect.dashPathEffect(floatArrayOf(3.dp.toPx(), 3.dp.toPx())))
                    pts[e]
                } else pts.last()
                drawCircle(Color.White, 6.dp.toPx(), punto)
                drawCircle(acento, 4.5.dp.toPx(), punto)
            }
        }
        val e = elegido
        if (e != null && e in actual.indices) {
            val lineas = mutableListOf(LineaGlobo("", Formato.pesos(actual[e]), C.texto))
            if (e in anterior.indices) lineas.add(LineaGlobo("Antes", Formato.pesos(anterior[e]), C.texto2))
            GloboEn(ancho * e / (largo - 1), ancho) { GloboGrafico(etiqueta?.invoke(e) ?: "", lineas) }
        }
    }
}

object EtiquetasVentas {
    /** Los nombres de los puntos de un período: la hora («14 h») hoy, el día («5/11») si no. */
    fun etiquetas(desde: Instant, hasta: Instant, puntos: Int, porHora: Boolean): (Int) -> String {
        val total = hasta.toEpochMilli() - desde.toEpochMilli()
        return { i ->
            val f = Formato.enArgentina(Instant.ofEpochMilli(desde.toEpochMilli() + (total * i / max(1, puntos - 1))))
            if (porHora) "${f.hour}${Formato.fino}h" else "${f.dayOfMonth}/${f.monthValue}"
        }
    }
}

// MARK: - Flujo de caja

data class DiaFlujo(val dia: Int, val entra: Int, val sale: Int)

/** Flujo de caja: lo que entra y lo que sale por día (barras enfrentadas). Tocando un día se ve el detalle. */
@Composable
fun GraficoFlujo(dias: List<DiaFlujo>) {
    var elegido by remember { mutableStateOf<Int?>(null) }
    val factor = crecer()
    val medidor = rememberTextMeasurer()
    val ok = C.ok; val error = C.error; val linea = C.linea; val linea2 = C.linea2; val texto3 = C.texto3
    val maxEntra = dias.maxOfOrNull { it.entra } ?: 0
    val maxSale = dias.maxOfOrNull { it.sale } ?: 0
    val marcas = marcasEje(-maxSale.toDouble(), maxOf(1, maxEntra).toDouble(), 4)
    val arriba = maxOf(marcas.maxOrNull() ?: 1.0, maxEntra.toDouble(), 1.0)
    val abajo = minOf(marcas.minOrNull() ?: 0.0, -maxSale.toDouble(), 0.0)
    var ancho by remember { mutableStateOf(0f) }
    val izquierda = 44f
    BoxWithConstraints(Modifier.fillMaxWidth().height(170.dp)) {
        val densidad = LocalDensity.current.density
        Canvas(
            Modifier.fillMaxWidth().height(170.dp).onSizeChanged { ancho = it.width - izquierda * densidad }
                .pointerInput(dias) {
                    detectTapGestures { o ->
                        val i = indiceEn(o.x - izquierda * density - (ancho / dias.size.coerceAtLeast(1)) / 2, ancho - ancho / dias.size.coerceAtLeast(1), dias.size)
                        val n = i?.let { dias[it].dia }
                        elegido = if (n == elegido) null else n; Haptico.seleccion()
                    }
                },
        ) {
            val iz = izquierda.dp.toPx()
            val altoEje = 16.dp.toPx()
            val h = size.height - altoEje
            fun y(v: Double) = (h * (arriba - v) / (arriba - abajo)).toFloat()
            marcas.forEach { m ->
                drawLine(linea, Offset(iz, y(m)), Offset(size.width, y(m)), 1f)
                textoEje(medidor, Formato.compacto(m.roundToInt()), texto3, 0f, y(m))
            }
            val w = (size.width - iz) / dias.size.coerceAtLeast(1)
            dias.forEachIndexed { i, d ->
                val apagado = elegido != null && elegido != d.dia
                val x = iz + w * i + w * 0.2f
                val bw = w * 0.6f
                val y0 = y(0.0)
                val ye = y(d.entra * factor.toDouble())
                drawRoundRect(ok.copy(alpha = if (apagado) 0.3f else 1f), Offset(x, ye), Size(bw, y0 - ye), CornerRadius(3.dp.toPx()))
                val ys = y(-d.sale * factor.toDouble())
                drawRoundRect(error.copy(alpha = if (apagado) 0.25f else 0.8f), Offset(x, y0), Size(bw, ys - y0), CornerRadius(3.dp.toPx()))
            }
            drawLine(linea2, Offset(iz, y(0.0)), Offset(size.width, y(0.0)), 1f)
            val cada = max(1, ceil(dias.size / 7.0).toInt())
            dias.forEachIndexed { i, d -> if (i % cada == 0) textoEje(medidor, "${d.dia}", texto3, iz + w * i + w / 2, size.height - altoEje / 2, centrado = true) }
        }
        val e = elegido
        val d = dias.firstOrNull { it.dia == e }
        if (d != null) {
            val i = dias.indexOf(d)
            val w = ancho / dias.size.coerceAtLeast(1)
            GloboEn(izquierda * LocalDensity.current.density + w * i + w / 2, ancho + izquierda * LocalDensity.current.density) {
                GloboGrafico("Día $e", listOf(LineaGlobo("Entra", Formato.pesos(d.entra), C.okTinta), LineaGlobo("Sale", Formato.pesos(d.sale), C.errorTinta)))
            }
        }
    }
}

/** Cómo queda la plata día a día si entra y sale lo previsto (lo acumulado del flujo). */
@Composable
fun GraficoSaldo(flujo: List<DiaFlujo>) {
    var suma = 0
    val dias = flujo.map { d -> suma += d.entra - d.sale; d.dia to suma }
    var elegido by remember { mutableStateOf<Int?>(null) }
    val factor = crecer()
    val medidor = rememberTextMeasurer()
    val positivo = (dias.lastOrNull()?.second ?: 0) >= 0
    val tono = if (positivo) C.ok else C.error
    val linea = C.linea; val linea2 = C.linea2; val texto3 = C.texto3
    val minimo = minOf(0, dias.minOfOrNull { it.second } ?: 0).toDouble()
    val maximo = maxOf(1, dias.maxOfOrNull { it.second } ?: 0).toDouble()
    val marcas = marcasEje(minimo, maximo, 4)
    val arriba = maxOf(marcas.maxOrNull() ?: maximo, maximo)
    val abajo = minOf(marcas.minOrNull() ?: minimo, minimo)
    var ancho by remember { mutableStateOf(0f) }
    val izquierda = 44f
    BoxWithConstraints(Modifier.fillMaxWidth().height(140.dp)) {
        val densidad = LocalDensity.current.density
        Canvas(
            Modifier.fillMaxWidth().height(140.dp).onSizeChanged { ancho = it.width - izquierda * densidad }
                .pointerInput(dias) {
                    detectTapGestures { o -> elegido = indiceEn(o.x - izquierda * density, ancho, dias.size)?.let { dias[it].first }; Haptico.seleccion() }
                }
                .pointerInput(dias) {
                    detectDragGestures(onDragEnd = { elegido = null }) { c, _ -> elegido = indiceEn(c.position.x - izquierda * density, ancho, dias.size)?.let { dias[it].first } }
                },
        ) {
            val iz = izquierda.dp.toPx()
            val altoEje = 16.dp.toPx()
            val h = size.height - altoEje
            fun y(v: Double) = (h * (arriba - v) / (arriba - abajo)).toFloat()
            fun x(i: Int) = iz + (size.width - iz) * i / max(1, dias.size - 1)
            marcas.forEach { m ->
                drawLine(linea, Offset(iz, y(m)), Offset(size.width, y(m)), 1f)
                textoEje(medidor, Formato.compacto(m.roundToInt()), texto3, 0f, y(m))
            }
            if (dias.isNotEmpty()) {
                val pts = dias.mapIndexed { i, d -> Offset(x(i), y(d.second * factor.toDouble())) }
                val area = curva(pts).apply { lineTo(pts.last().x, y(0.0)); lineTo(pts.first().x, y(0.0)); close() }
                drawPath(area, Brush.verticalGradient(listOf(tono.copy(alpha = 0.22f), tono.copy(alpha = 0f))))
                drawPath(curva(pts), tono, style = Stroke(2.5.dp.toPx(), cap = StrokeCap.Round))
                val e = dias.indexOfFirst { it.first == elegido }
                if (e >= 0) {
                    drawLine(linea2, Offset(x(e), 0f), Offset(x(e), h), 1.dp.toPx(), pathEffect = PathEffect.dashPathEffect(floatArrayOf(3.dp.toPx(), 3.dp.toPx())))
                    drawCircle(tono, 4.5.dp.toPx(), pts[e])
                }
            }
            drawLine(linea2, Offset(iz, y(0.0)), Offset(size.width, y(0.0)), 1f)
            val cada = max(1, ceil(dias.size / 7.0).toInt())
            dias.forEachIndexed { i, d -> if (i % cada == 0) textoEje(medidor, "${d.first}", texto3, x(i), size.height - altoEje / 2, centrado = true) }
        }
        val e = dias.indexOfFirst { it.first == elegido }
        if (e >= 0) {
            val d = dias[e]
            val dens = LocalDensity.current.density
            GloboEn(izquierda * dens + ancho * e / max(1, dias.size - 1), ancho + izquierda * dens) {
                GloboGrafico("Hasta el día ${d.first}", listOf(LineaGlobo("", (if (d.second < 0) "−" else "+") + Formato.pesos(kotlin.math.abs(d.second)), if (d.second < 0) C.errorTinta else C.okTinta)))
            }
        }
    }
}

// MARK: - Dona

data class ParteDona(val nombre: String, val monto: Int)

/** Dona de medios de pago (o canales): tocando una porción, o su renglón, se destaca y se ve su parte en el centro. */
@Composable
fun DonaMedios(partes: List<ParteDona>) {
    val colores = listOf(hex(0xD9652F), hex(0xE3A93B), hex(0x6B5244), hex(0x2F8F63), hex(0x8E3A16), hex(0xB5A193))
    var elegida by remember { mutableStateOf<Int?>(null) }
    val factor = crecer()
    val total = maxOf(1, partes.sumOf { it.monto })
    val destacada = elegida ?: partes.indices.maxByOrNull { partes[it].monto }
    fun porcentaje(n: Int) = (n.toDouble() / total * 100).roundToInt()
    fun elegir(i: Int) { elegida = if (elegida == i) null else i; Haptico.seleccion() }
    Row(horizontalArrangement = Arrangement.spacedBy(18.dp), verticalAlignment = Alignment.CenterVertically) {
        Box(
            Modifier.size(120.dp).scale(0.6f + 0.4f * factor).rotate(-90f * (1 - factor)).alpha(factor),
            contentAlignment = Alignment.Center,
        ) {
            Canvas(Modifier.size(120.dp).pointerInput(partes) {
                detectTapGestures { o ->
                    val c = Offset(size.width / 2f, size.height / 2f)
                    var a = Math.toDegrees(atan2((o.y - c.y).toDouble(), (o.x - c.x).toDouble())) + 90
                    if (a < 0) a += 360
                    var suma = 0.0
                    for ((i, p) in partes.withIndex()) {
                        suma += p.monto.toDouble() / total * 360
                        if (a <= suma) { elegir(i); break }
                    }
                }
            }) {
                var inicio = -90f
                val r = size.minDimension / 2
                partes.forEachIndexed { i, p ->
                    val barrido = p.monto.toFloat() / total * 360f
                    val on = elegida == i
                    val externo = r * (if (on) 1f else 0.94f)
                    val interno = r * (if (on) 0.58f else 0.62f)
                    val grosor = externo - interno
                    val radio = interno + grosor / 2
                    val color = colores[i % colores.size].copy(alpha = if (elegida == null || on) 1f else 0.4f)
                    drawArc(color, inicio + 1f, (barrido - 2f).coerceAtLeast(0.5f), false,
                        topLeft = Offset(center.x - radio, center.y - radio), size = Size(radio * 2, radio * 2), style = Stroke(grosor))
                    inicio += barrido
                }
            }
            if (destacada != null && destacada in partes.indices) {
                Column(Modifier.width(64.dp), horizontalAlignment = Alignment.CenterHorizontally) {
                    NumeroAnimado("${porcentaje(partes[destacada].monto)} %", Estilo.headline, C.texto)
                    Texto(partes[destacada].nombre, Estilo.chip, C.texto2, lineas = 1, escalaMinima = 0.7f, alinear = TextAlign.Center)
                }
            }
        }
        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(8.dp)) {
            partes.forEachIndexed { i, p ->
                Row(
                    Modifier.fillMaxWidth().alpha(if (elegida == null || elegida == i) 1f else 0.5f).clickable { elegir(i) },
                    horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.CenterVertically,
                ) {
                    Box(Modifier.size(9.dp).clip(CircleShape).background(colores[i % colores.size]))
                    Texto(p.nombre, Estilo.texto14, C.texto, Modifier.weight(1f), lineas = 1)
                    Texto("${porcentaje(p.monto)} %", Estilo.footnoteFuerte, C.texto2, numeros = true)
                }
            }
        }
    }
}

// MARK: - Mes a mes

/** Una serie del gráfico por mes. `formato`: el valor en el globo; `eje`: las marcas del eje. */
data class SerieMeses(val id: String, val nombre: String, val valores: List<Int?>, val formato: (Int) -> String, val eje: (Int) -> String)

/** Barras por mes con varias series para elegir (ventas, clientes, visitas…). Tocando un mes se ve el valor; el último
 *  mes va destacado. */
@Composable
fun GraficoMeses(meses: List<String>, series: List<SerieMeses>) {
    var serieId by remember { mutableStateOf("") }
    var elegido by remember { mutableStateOf<String?>(null) }
    val factor = crecer()
    val serie = series.firstOrNull { it.id == serieId } ?: series.first()
    val barras = meses.zip(serie.valores).mapNotNull { (m, v) -> v?.let { m to it } }
    val cambio by animateFloatAsState(1f, Mov.crecer(), label = "serie")
    val medidor = rememberTextMeasurer()
    val acento = C.acento; val linea = C.linea; val texto3 = C.texto3
    Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
        if (series.size > 1) {
            Segmentado(series.map { it.id to it.nombre }, if (serieId.isEmpty()) series[0].id else serieId, { elegido = null; serieId = it })
        }
        val tope = maxOf(1, barras.maxOfOrNull { it.second } ?: 1).toDouble()
        val marcas = marcasEje(0.0, tope, 4)
        val arriba = maxOf(marcas.maxOrNull() ?: tope, tope) * 1.12
        var ancho by remember { mutableStateOf(0f) }
        val izquierda = 44f
        BoxWithConstraints(Modifier.fillMaxWidth().height(180.dp)) {
            val dens = LocalDensity.current.density
            Canvas(
                Modifier.fillMaxWidth().height(180.dp).onSizeChanged { ancho = it.width - izquierda * dens }
                    .pointerInput(barras) {
                        detectTapGestures { o ->
                            val w = ancho / barras.size.coerceAtLeast(1)
                            val i = ((o.x - izquierda * density) / w).toInt().coerceIn(0, (barras.size - 1).coerceAtLeast(0))
                            val m = barras.getOrNull(i)?.first
                            elegido = if (m == elegido) null else m; Haptico.seleccion()
                        }
                    },
            ) {
                val iz = izquierda.dp.toPx()
                val altoEje = 16.dp.toPx()
                val h = size.height - altoEje
                fun y(v: Double) = (h * (1 - v / arriba)).toFloat()
                marcas.forEach { m ->
                    drawLine(linea, Offset(iz, y(m)), Offset(size.width, y(m)), 1f)
                    textoEje(medidor, serie.eje(m.roundToInt()), texto3, 0f, y(m))
                }
                val w = (size.width - iz) / barras.size.coerceAtLeast(1)
                barras.forEachIndexed { i, (m, v) ->
                    val ultimo = i == barras.size - 1
                    val x = iz + w * i + w * 0.18f
                    val bw = w * 0.64f
                    val yv = y(v * factor * cambio.toDouble())
                    val brocha = if (ultimo || elegido == m) Brush.verticalGradient(listOf(acento, hex(0xB0441D)), yv, h)
                    else Brush.verticalGradient(listOf(acento.copy(alpha = if (elegido == null) 0.45f else 0.25f), acento.copy(alpha = if (elegido == null) 0.45f else 0.25f)))
                    drawRoundRect(brocha, Offset(x, yv), Size(bw, h - yv), CornerRadius(5.dp.toPx()))
                    textoEje(medidor, m, texto3, iz + w * i + w / 2, size.height - altoEje / 2, centrado = true)
                }
            }
            val i = barras.indexOfFirst { it.first == elegido }
            if (i >= 0) {
                val w = ancho / barras.size.coerceAtLeast(1)
                GloboEn(izquierda * dens + w * i + w / 2, ancho + izquierda * dens) {
                    GloboGrafico(barras[i].first, listOf(LineaGlobo("", serie.formato(barras[i].second), C.texto)))
                }
            }
        }
    }
}

// MARK: - Embudo de visitas

/** Cómo recorren la tienda: de cada 100 que entran, cuántos ven el catálogo, un producto, el carrito y el pago. */
@Composable
fun EmbudoVisitas(paginas: List<Tablero.Visitas.Pagina>) {
    val orden = listOf("inicio" to "Inicio", "catalogo" to "Catálogo", "producto" to "Producto", "carrito" to "Carrito", "checkout" to "Pago")
    val pasos = orden.mapNotNull { (id, nombre) -> paginas.firstOrNull { it.pagina == id }?.let { nombre to it.vistas } }
    val factor = crecer()
    val tope = maxOf(1, pasos.maxOfOrNull { it.second } ?: 1)
    Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
        pasos.forEachIndexed { i, (nombre, vistas) ->
            Row(Modifier.fillMaxWidth().height(30.dp), verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                Texto(nombre, Estilo.footnoteFuerte, C.texto, Modifier.width(72.dp), lineas = 1)
                BoxWithConstraints(Modifier.weight(1f)) {
                    val w = maxWidth / 1.45f * (vistas.toFloat() / tope) * factor
                    Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(6.dp)) {
                        Box(Modifier.width(w).height(26.dp).clip(RoundedCornerShape(6.dp)).background(C.acento.copy(alpha = 1f - i * 0.15f)))
                        Texto("${Formato.numero(vistas)} · ${(vistas.toDouble() / tope * 100).roundToInt()} %", Estilo.chip, C.texto2, Modifier.alpha(factor), numeros = true, lineas = 1)
                    }
                }
            }
        }
    }
}

// MARK: - Barras

/** Barras horizontales por cantidad de estrellas (5 a 1). */
@Composable
fun BarrasEstrellas(porEstrellas: List<Int>) {
    val factor = crecer()
    val filas = (5 downTo 1).map { n -> n to (porEstrellas.getOrNull(n - 1) ?: 0) }
    val tope = maxOf(1, filas.maxOfOrNull { it.second } ?: 1)
    Column(Modifier.height(150.dp), verticalArrangement = Arrangement.SpaceEvenly) {
        filas.forEach { (n, cantidad) ->
            Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                Texto("$n ★", Estilo.footnoteFuerte, C.texto2, Modifier.width(30.dp))
                BoxWithConstraints(Modifier.weight(1f)) {
                    val w = maxWidth / 1.15f * (cantidad.toFloat() / tope) * factor
                    Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(6.dp)) {
                        Box(Modifier.width(w).height(18.dp).clip(RoundedCornerShape(4.dp)).background(Brush.horizontalGradient(listOf(C.acento, hex(0xB0441D)))))
                        Texto("$cantidad", Estilo.footnote, C.texto3, Modifier.alpha(factor), numeros = true)
                    }
                }
            }
        }
    }
}

/** Anillo de avance (la meta del mes): se llena al aparecer. */
@Composable
fun AnilloAvance(valor: Double, tam: Dp = 76.dp, grosor: Dp = 9.dp) {
    val factor = crecer(0.2)
    val v = valor.coerceIn(0.0, 1.0).toFloat()
    val pista = C.pistaMeta; val acento = C.acento
    Box(Modifier.size(tam).semantics { contentDescription = "Avance"; stateDescription = "${(valor * 100).roundToInt()} %" }, contentAlignment = Alignment.Center) {
        Canvas(Modifier.size(tam)) {
            val g = grosor.toPx()
            val r = size.minDimension / 2 - g / 2
            val tl = Offset(center.x - r, center.y - r)
            drawCircle(pista, r, style = Stroke(g))
            drawArc(Brush.sweepGradient(listOf(acento, hex(0xB0441D), acento)), -90f, 360f * v * factor, false, tl, Size(r * 2, r * 2), style = Stroke(g, cap = StrokeCap.Round))
        }
        NumeroAnimado("${(valor * 100 * factor).roundToInt()} %", Estilo.headline, C.texto)
    }
}
