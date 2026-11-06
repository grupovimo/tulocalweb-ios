package ar.com.tulocalweb.app.tema

import androidx.compose.animation.core.Animatable
import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ColumnScope
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.semantics.stateDescription
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.Tincho
import kotlinx.coroutines.delay
import kotlin.math.abs
import kotlin.math.roundToInt

// Clon de Tema/Decoracion.swift: piezas para que las pantallas de sección no queden planas: la tarjeta héroe con el
// degradé terracota, indicadores con ícono, barras de progreso y gráficos con los colores del handoff.

/** Tarjeta destacada: degradé `primary`, monto grande en blanco, un renglón abajo y Tincho opcional. */
@Composable
fun TarjetaHeroe(
    etiqueta: String,
    valor: String,
    detalle: String? = null,
    tincho: Tincho? = null,
    pie: @Composable ColumnScope.() -> Unit = {},
) {
    val forma = RoundedCornerShape(24.dp)
    val brocha = C.primario
    Box(
        Modifier.fillMaxWidth().aparecer().sombra(forma, C.sombraPrimario, 14.dp, y = 8.dp).clip(forma).background(brocha)
            .brilloRadial(Color.White.copy(alpha = 0.18f), Offset(1f, 0f), 260f),
    ) {
        if (tincho != null) {
            ImagenTincho(tincho, Modifier.align(Alignment.BottomEnd).offset(x = 14.dp, y = 18.dp).size(110.dp).alpha(0.95f))
        }
        Column(Modifier.fillMaxWidth().padding(18.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
            Texto(etiqueta, Estilo.footnoteFuerte, Color.White.copy(alpha = 0.85f))
            NumeroAnimado(valor, Estilo.montoPantalla, Color.White, escalaMinima = 0.6f)
            if (detalle != null) {
                Texto(detalle, Estilo.texto14, Color.White.copy(alpha = 0.9f), Modifier.padding(end = if (tincho == null) 0.dp else 90.dp))
            }
            pie()
        }
    }
}

/** Indicador con ícono en un cuadradito de color y un renglón de detalle. */
data class IndicadorIcono(
    val simbolo: String,
    val titulo: String,
    val valor: String,
    val detalle: String? = null,
    val tono: Pair<Tono, Tono>? = null,
)

@Composable
fun VistaIndicadorIcono(i: IndicadorIcono, modifier: Modifier = Modifier) {
    val fondo = i.tono?.first?.actual ?: C.crema
    val tinta = i.tono?.second?.actual ?: C.primarioTexto
    Column(modifier.fillMaxWidth().tarjeta(radio = 18.dp).padding(14.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
        Box(Modifier.size(32.dp).clip(RoundedCornerShape(9.dp)).background(fondo), contentAlignment = Alignment.Center) {
            Icono(i.simbolo, 15f, tinta)
        }
        NumeroAnimado(i.valor, Estilo.indicador, C.texto, escalaMinima = 0.6f)
        Texto(i.titulo, Estilo.footnote, C.texto2, lineas = 2)
        if (!i.detalle.isNullOrEmpty()) Texto(i.detalle, Estilo.chip, C.texto3, lineas = 2)
    }
}

@Composable
fun GrillaIndicadores(items: List<IndicadorIcono>) {
    Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
        items.chunked(2).forEachIndexed { f, par ->
            Row(horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                par.forEachIndexed { c, i -> VistaIndicadorIcono(i, Modifier.weight(1f).aparecer(0.05 + (f * 2 + c) * 0.05)) }
                if (par.size == 1) Spacer(Modifier.weight(1f))
            }
        }
    }
}

/** Barra de progreso con el degradé `accent → primary`. */
@Composable
fun BarraProgreso(valor: Double, alto: Dp = 10.dp, fondo: Color = C.pistaMeta) {
    val anima = animacionesActivas()
    val lleno = remember { Animatable(if (anima) 0f else 1f) }
    LaunchedEffect(Unit) { delay(150); lleno.animateTo(1f, Mov.crecer()) }
    val v by animateFloatAsState(valor.coerceIn(0.0, 1.0).toFloat(), Mov.crecer(), label = "progreso")
    BoxWithConstraints(Modifier.fillMaxWidth().height(alto).semantics { stateDescription = "${(valor * 100).roundToInt()} %" }) {
        Box(Modifier.fillMaxWidth().height(alto).clip(Capsula).background(fondo))
        Box(Modifier.width(maxOf(alto, maxWidth * v * lleno.value)).height(alto).clip(Capsula)
            .background(Brush.horizontalGradient(listOf(C.acento, hex(0xB0441D)))))
    }
}

/** Tarjeta con título de sección adentro. */
@Composable
fun TarjetaSeccion(titulo: String, simbolo: String? = null, contenido: @Composable ColumnScope.() -> Unit) {
    Column(Modifier.fillMaxWidth().aparecer(0.12).tarjeta(radio = 20.dp).padding(16.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
        Row(horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.CenterVertically) {
            if (simbolo != null) Icono(simbolo, 13f, C.primarioTexto)
            Texto(titulo, Estilo.headline, C.texto)
        }
        contenido()
    }
}

/** «$ 1,2 M», «$ 350 mil», «$ 900». */
fun Formato.compacto(n: Int): String {
    val a = abs(n)
    val signo = if (n < 0) "−" else ""
    if (a >= 1_000_000) return signo + "$" + fino + "%.1f".format(java.util.Locale.US, a / 1_000_000.0).replace(".", ",") + " M"
    if (a >= 1_000) return signo + "$" + fino + "${a / 1_000} mil"
    return signo + "$" + fino + "$a"
}

/** Ícono por tipo de movimiento de plata. */
fun simboloMovimiento(tipo: String): String = when (tipo) {
    "mp", "payway", "modo", "uala" -> "creditcard.fill"
    "andreani", "oca", "correo", "moto" -> "shippingbox.fill"
    "transf" -> "arrow.left.arrow.right"
    "gasto" -> "cart.fill"
    "retiro" -> "banknote.fill"
    else -> "circle.fill"
}
