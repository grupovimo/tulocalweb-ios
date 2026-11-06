package ar.com.tulocalweb.app.widgets

import android.content.Context
import android.content.Intent
import android.content.res.Configuration
import android.graphics.Bitmap
import android.graphics.Canvas
import android.graphics.Paint
import android.graphics.Path
import android.net.Uri
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.DpSize
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.glance.GlanceId
import androidx.glance.GlanceModifier
import androidx.glance.Image
import androidx.glance.ImageProvider
import androidx.glance.LocalSize
import androidx.glance.action.clickable
import androidx.glance.appwidget.GlanceAppWidget
import androidx.glance.appwidget.GlanceAppWidgetReceiver
import androidx.glance.appwidget.SizeMode
import androidx.glance.appwidget.action.actionStartActivity
import androidx.glance.appwidget.cornerRadius
import androidx.glance.appwidget.provideContent
import androidx.glance.appwidget.updateAll
import androidx.glance.background
import androidx.glance.layout.Box
import androidx.glance.layout.Column
import androidx.glance.layout.ContentScale
import androidx.glance.layout.Row
import androidx.glance.layout.Spacer
import androidx.glance.layout.fillMaxHeight
import androidx.glance.layout.fillMaxSize
import androidx.glance.layout.fillMaxWidth
import androidx.glance.layout.height
import androidx.glance.layout.padding
import androidx.glance.layout.width
import androidx.glance.text.FontWeight
import androidx.glance.text.Text
import androidx.glance.text.TextStyle
import androidx.glance.unit.ColorProvider
import ar.com.tulocalweb.app.MainActivity
import ar.com.tulocalweb.app.compartido.ColorWidget
import ar.com.tulocalweb.app.compartido.DatosHoy
import ar.com.tulocalweb.app.compartido.PesosWidget
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.launch
import kotlin.math.abs

// Clon de TuLocalWebWidgets/Widgets.swift con Glance: «Ventas de hoy» (chico y mediano). En los teléfonos Android no
// hay Live Activities: lo de ActividadPedidos lo muestra la notificación en curso (nucleo/ActividadEnApp.kt).

object Widgets {
    private val alcance = CoroutineScope(SupervisorJob() + Dispatchers.Default)

    /** Vuelve a dibujar los widgets (WidgetCenter.reloadAllTimelines en iOS). */
    fun actualizar(ctx: Context) { alcance.launch { runCatching { VentasDeHoy().updateAll(ctx.applicationContext) } } }
}

/** Los colores de ColorWidget según el modo del teléfono (los widgets de iOS siguen el modo del sistema). */
private class Paleta(ctx: Context) {
    private val oscuro = (ctx.resources.configuration.uiMode and Configuration.UI_MODE_NIGHT_MASK) == Configuration.UI_MODE_NIGHT_YES
    private fun c(p: Pair<Long, Long>, alfa: Float = 1f) = Color(0xFF000000 or (if (oscuro) p.second else p.first)).copy(alpha = alfa)
    val fondo = c(ColorWidget.fondo)
    val texto = c(ColorWidget.texto)
    fun texto2(alfa: Float = 1f) = c(ColorWidget.texto2, alfa)
    val acento = c(ColorWidget.acento)
    val primario = c(ColorWidget.primario)
    val ok = c(ColorWidget.ok)
    val error = c(ColorWidget.error)
}

private fun cp(c: Color) = ColorProvider(c)

/** «Ventas de hoy»: cómo viene el día y los pedidos por preparar. */
class VentasDeHoy : GlanceAppWidget() {
    companion object {
        private val CHICO = DpSize(110.dp, 110.dp)
        private val MEDIANO = DpSize(250.dp, 110.dp)
    }

    override val sizeMode = SizeMode.Responsive(setOf(CHICO, MEDIANO))

    override suspend fun provideGlance(context: Context, id: GlanceId) {
        val d = runCatching { DatosHoy.traer(context) }.getOrDefault(DatosHoy.sinSesion)
        provideContent { Vista(context, d) }
    }

    private fun abrir(ctx: Context, ruta: String) =
        actionStartActivity(Intent(ctx, MainActivity::class.java).setAction(Intent.ACTION_VIEW).setData(Uri.parse("tulocalweb://$ruta")))

    @Composable
    private fun Vista(ctx: Context, d: DatosHoy) {
        val p = Paleta(ctx)
        val mediano = LocalSize.current.width >= MEDIANO.width
        Box(GlanceModifier.fillMaxSize().background(p.fondo).cornerRadius(22.dp).padding(16.dp).clickable(abrir(ctx, "inicio"))) {
            when {
                !d.conSesion -> Column(GlanceModifier.fillMaxSize()) {
                    Text("Tu Local Web", style = TextStyle(cp(p.primario), 12.sp, FontWeight.Bold))
                    Spacer(GlanceModifier.defaultWeight())
                    Text("Entrá a la app para ver tus ventas acá.", style = TextStyle(cp(p.texto2()), 13.sp))
                }
                mediano -> Row(GlanceModifier.fillMaxSize()) {
                    Box(GlanceModifier.defaultWeight().fillMaxHeight()) { Hoy(ctx, d, p) }
                    Spacer(GlanceModifier.width(14.dp))
                    Box(GlanceModifier.width(1.dp).fillMaxHeight().background(p.texto2(0.18f))) {}
                    Spacer(GlanceModifier.width(14.dp))
                    Column(GlanceModifier.defaultWeight().fillMaxHeight()) {
                        Text("Por preparar", style = TextStyle(cp(p.texto2()), 12.sp, FontWeight.Medium))
                        Spacer(GlanceModifier.height(6.dp))
                        Text("${d.porPreparar}", style = TextStyle(cp(if (d.porPreparar > 0) p.acento else p.texto), 30.sp, FontWeight.Bold))
                        Spacer(GlanceModifier.defaultWeight())
                        d.ultimo?.let { u ->
                            Column(GlanceModifier.clickable(abrir(ctx, "pedido/${u.numero}"))) {
                                Text("Último · #${u.numero}", style = TextStyle(cp(p.texto2()), 11.sp, FontWeight.Medium))
                                Text(u.cliente, style = TextStyle(cp(p.texto), 13.sp, FontWeight.Medium), maxLines = 1)
                                if (d.montos && u.total != null) Text(PesosWidget.texto(u.total.toDouble()), style = TextStyle(cp(p.texto2()), 13.sp))
                            }
                        }
                    }
                }
                else -> Hoy(ctx, d, p)
            }
        }
    }

    @Composable
    private fun Hoy(ctx: Context, d: DatosHoy, p: Paleta) {
        Column(GlanceModifier.fillMaxSize()) {
            Text("Hoy", style = TextStyle(cp(p.texto2()), 12.sp, FontWeight.Medium))
            Spacer(GlanceModifier.height(4.dp))
            if (d.montos) {
                Text(PesosWidget.texto(d.ventasHoy), style = TextStyle(cp(p.texto), 24.sp, FontWeight.Bold), maxLines = 1)
                d.delta?.let { x ->
                    Text("${if (x >= 0) "▲" else "▼"} ${String.format(java.util.Locale.US, "%.1f", abs(x)).replace(".", ",")} % vs. ayer",
                        style = TextStyle(cp(if (x >= 0) p.ok else p.error), 11.sp, FontWeight.Medium))
                }
            } else {
                Text("${d.pedidosHoy}", style = TextStyle(cp(p.texto), 28.sp, FontWeight.Bold))
                Text(if (d.pedidosHoy == 1) "pedido" else "pedidos", style = TextStyle(cp(p.texto2()), 11.sp))
            }
            Spacer(GlanceModifier.defaultWeight())
            Image(ImageProvider(linea(ctx, d.serie, p.acento)), null, GlanceModifier.fillMaxWidth().height(30.dp), contentScale = ContentScale.FillBounds)
        }
    }

    /** Sparkline de las ventas acumuladas del día (Linea en iOS): Glance no dibuja trazos, va como imagen. */
    private fun linea(ctx: Context, valores: List<Double>, color: Color): Bitmap {
        val den = ctx.resources.displayMetrics.density
        val w = (200 * den).toInt(); val h = (30 * den).toInt()
        val b = Bitmap.createBitmap(w, h, Bitmap.Config.ARGB_8888)
        val pintura = Paint(Paint.ANTI_ALIAS_FLAG).apply {
            style = Paint.Style.STROKE; strokeWidth = 2.5f * den; strokeCap = Paint.Cap.ROUND; strokeJoin = Paint.Join.ROUND
            this.color = android.graphics.Color.argb((color.alpha * 255).toInt(), (color.red * 255).toInt(), (color.green * 255).toInt(), (color.blue * 255).toInt())
        }
        val m = pintura.strokeWidth / 2
        val ancho = w - 2 * m; val alto = h - 2 * m
        val camino = Path()
        val max = valores.maxOrNull() ?: 0.0
        if (valores.size <= 1 || max <= 0) { camino.moveTo(m, m + alto); camino.lineTo(m + ancho, m + alto) }
        else valores.forEachIndexed { i, v ->
            val x = m + ancho * i / (valores.size - 1); val y = (m + alto - alto * (v / max)).toFloat()
            if (i == 0) camino.moveTo(x, y) else camino.lineTo(x, y)
        }
        Canvas(b).drawPath(camino, pintura)
        return b
    }
}

class VentasDeHoyReceiver : GlanceAppWidgetReceiver() {
    override val glanceAppWidget: GlanceAppWidget = VentasDeHoy()
}
