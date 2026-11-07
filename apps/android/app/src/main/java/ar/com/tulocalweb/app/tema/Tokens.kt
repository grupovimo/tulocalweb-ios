package ar.com.tulocalweb.app.tema

import android.view.HapticFeedbackConstants
import android.view.View
import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.ui.Modifier
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.ReadOnlyComposable
import androidx.compose.runtime.staticCompositionLocalOf
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.Font
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.LineBreak
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.em
import androidx.compose.ui.unit.sp
import ar.com.tulocalweb.app.R
import java.lang.ref.WeakReference

// Clon de Tema/Tokens.swift: los tokens del handoff (docs/handoff/app-ios/README.md · Design Tokens), claro y oscuro.
// Un punto de iOS es un dp acá (y un sp en el texto, que escala con el tamaño de letra del sistema).

fun hex(v: Long, alfa: Float = 1f): Color = Color(0xFF000000 or v).copy(alpha = alfa)

/** Si se ve en oscuro (lo fija TemaTLW: el sistema o lo que se eligió en Cuenta). */
val LocalOscuro = staticCompositionLocalOf { false }

/** Un color que cambia con el modo oscuro (Color.tema en iOS). */
class Tono(val claro: Color, val oscuro: Color) {
    val actual: Color @Composable @ReadOnlyComposable get() = if (LocalOscuro.current) oscuro else claro
    fun en(oscuro: Boolean): Color = if (oscuro) this.oscuro else claro
}

fun tema(claro: Long, oscuro: Long, alfa: Float = 1f, alfaOscuro: Float? = null) = Tono(hex(claro, alfa), hex(oscuro, alfaOscuro ?: alfa))

/** El tema: «sistema» sigue al teléfono (como @AppStorage("tema") en iOS), «claro» y «oscuro» lo fijan. */
@Composable
fun TemaTLW(tema: String = "sistema", contenido: @Composable () -> Unit) {
    val oscuro = when (tema) { "oscuro" -> true; "claro" -> false; else -> isSystemInDarkTheme() }
    CompositionLocalProvider(LocalOscuro provides oscuro) {
        // El ancho de la pantalla, para los márgenes del handoff (SE / estándar / Pro Max).
        // El fondo lo pinta la app y no la ventana: la ventana sigue el tema del sistema y, con «Oscuro» elegido en la
        // app y el teléfono en claro, las pantallas sin fondo propio (Inicio, Pedidos, Productos) quedaban crema
        // con letra clara.
        BoxWithConstraints(Modifier.fillMaxSize().background(C.fondo)) {
            CompositionLocalProvider(LocalAnchoPantalla provides maxWidth.value, content = contenido)
        }
    }
}

/** Los tonos del handoff, por nombre (los de C, sin resolver). */
object T {
    val fondo = tema(0xF7F1EA, 0x22140D)
    val elevado = tema(0xFBF7F2, 0x2E1B12)
    val superficie = tema(0xFFFFFF, 0x2E1B12)
    val texto = tema(0x2A140C, 0xFBF7F2)
    val texto2 = tema(0x6B5244, 0xCDB8A8)
    val texto3 = tema(0x7D6656, 0xA8907F)
    val chevron = tema(0xB5A193, 0x7D6656)
    val primarioTexto = tema(0xB0441D, 0xE07A48)
    val primarioTinta = tema(0x8E3A16, 0xF3A57E)
    val crema = tema(0xF8E6D4, 0x4A2A1A)
    val crema2 = tema(0xFBF3EA, 0x3A2216)
    val pista = tema(0xF3E9DF, 0x3A2419)
    val pistaMeta = tema(0xF3E6DA, 0x3A2419)
    val ok = tema(0x2F8F63, 0x3FA877)
    val okFondo = tema(0xE1F1E8, 0x3FA877, alfaOscuro = 0.18f)
    val okTinta = tema(0x1F6B48, 0x6FD0A0)
    val aviso = tema(0xC2861B, 0xE3A93B)
    val avisoFondo = tema(0xFBEFD8, 0xE3A93B, alfaOscuro = 0.18f)
    val avisoTinta = tema(0x7A4F08, 0xF0C267)
    val bannerTinta = tema(0x6E4707, 0xF0C267)
    val errorFondo = tema(0xFBE4DA, 0xD2582F, alfaOscuro = 0.2f)
    val errorTinta = tema(0xA23F1C, 0xF08A63)
    val neutroFondo = tema(0xEFE6DF, 0xFBF7F2, alfaOscuro = 0.1f)
    val neutroTinta = tema(0x5A4033, 0xCDB8A8)
    val apagadoFondo = tema(0xEEE8E3, 0xFBF7F2, alfaOscuro = 0.08f)
    val linea = tema(0x3A1D12, 0xFBF7F2, alfa = 0.09f, alfaOscuro = 0.08f)
    val linea2 = tema(0x3A1D12, 0xFBF7F2, alfa = 0.16f, alfaOscuro = 0.12f)
    val anillo = tema(0x3A1D12, 0xFBF7F2, alfa = 0.06f, alfaOscuro = 0.08f)
    val segmentado = tema(0x3A1D12, 0xFBF7F2, alfa = 0.07f, alfaOscuro = 0.07f)
    val segmentoElegido = tema(0xFFFFFF, 0x4A3125)
    val rayado1 = tema(0xF3E3D2, 0x3A2419)
    val rayado2 = tema(0xFAEEE2, 0x33201A)
    val esqueleto1 = tema(0xEFE6DD, 0x3A2419)
    val esqueleto2 = tema(0xF3ECE5, 0x33201A)
    val canalLocal = tema(0x6B5244, 0xCDB8A8)
    val primarioArriba = tema(0xC24E25, 0xD9652F)
}

/** Colores del handoff. Los nombres siguen la tabla del README. */
object C {
    val fondo: Color @Composable @ReadOnlyComposable get() = T.fondo.actual
    val elevado: Color @Composable @ReadOnlyComposable get() = T.elevado.actual
    val superficie: Color @Composable @ReadOnlyComposable get() = T.superficie.actual
    val texto: Color @Composable @ReadOnlyComposable get() = T.texto.actual
    val texto2: Color @Composable @ReadOnlyComposable get() = T.texto2.actual
    val texto3: Color @Composable @ReadOnlyComposable get() = T.texto3.actual
    val chevron: Color @Composable @ReadOnlyComposable get() = T.chevron.actual
    val acento = hex(0xD9652F)
    val primarioTexto: Color @Composable @ReadOnlyComposable get() = T.primarioTexto.actual
    val primarioTinta: Color @Composable @ReadOnlyComposable get() = T.primarioTinta.actual
    val crema: Color @Composable @ReadOnlyComposable get() = T.crema.actual
    val crema2: Color @Composable @ReadOnlyComposable get() = T.crema2.actual
    val pista: Color @Composable @ReadOnlyComposable get() = T.pista.actual
    val pistaMeta: Color @Composable @ReadOnlyComposable get() = T.pistaMeta.actual
    val ok: Color @Composable @ReadOnlyComposable get() = T.ok.actual
    val okFondo: Color @Composable @ReadOnlyComposable get() = T.okFondo.actual
    val okTinta: Color @Composable @ReadOnlyComposable get() = T.okTinta.actual
    val aviso: Color @Composable @ReadOnlyComposable get() = T.aviso.actual
    val avisoFondo: Color @Composable @ReadOnlyComposable get() = T.avisoFondo.actual
    val avisoTinta: Color @Composable @ReadOnlyComposable get() = T.avisoTinta.actual
    val bannerTinta: Color @Composable @ReadOnlyComposable get() = T.bannerTinta.actual
    val error = hex(0xD2582F)
    val errorFondo: Color @Composable @ReadOnlyComposable get() = T.errorFondo.actual
    val errorTinta: Color @Composable @ReadOnlyComposable get() = T.errorTinta.actual
    val neutroFondo: Color @Composable @ReadOnlyComposable get() = T.neutroFondo.actual
    val neutroTinta: Color @Composable @ReadOnlyComposable get() = T.neutroTinta.actual
    val apagadoFondo: Color @Composable @ReadOnlyComposable get() = T.apagadoFondo.actual
    val linea: Color @Composable @ReadOnlyComposable get() = T.linea.actual
    val linea2: Color @Composable @ReadOnlyComposable get() = T.linea2.actual
    val anillo: Color @Composable @ReadOnlyComposable get() = T.anillo.actual
    val segmentado: Color @Composable @ReadOnlyComposable get() = T.segmentado.actual
    val segmentoElegido: Color @Composable @ReadOnlyComposable get() = T.segmentoElegido.actual
    val velo = hex(0x2A140C, 0.38f)
    val rayado1: Color @Composable @ReadOnlyComposable get() = T.rayado1.actual
    val rayado2: Color @Composable @ReadOnlyComposable get() = T.rayado2.actual
    val esqueleto1: Color @Composable @ReadOnlyComposable get() = T.esqueleto1.actual
    val esqueleto2: Color @Composable @ReadOnlyComposable get() = T.esqueleto2.actual
    val canalWeb = hex(0xD9652F)
    val canalML = hex(0xE3A93B)
    val canalLocal: Color @Composable @ReadOnlyComposable get() = T.canalLocal.actual
    /** El gradiente del botón principal, la burbuja de Tincho y el «+». */
    val primario: Brush @Composable @ReadOnlyComposable get() = Brush.verticalGradient(listOf(T.primarioArriba.actual, hex(0xB0441D)))
    val sombraPrimario = hex(0xB0441D, 0.7f)
}

/** Pesos de Instrument Sans incluidos en la app. */
enum class Peso(val peso: FontWeight) { regular(FontWeight.Normal), medium(FontWeight.Medium), semibold(FontWeight.SemiBold), bold(FontWeight.Bold) }

/** La letra: Instrument Sans 400–700 (las mismas instancias estáticas que iOS). */
val InstrumentSans = FontFamily(
    Font(R.font.instrument_regular, FontWeight.Normal),
    Font(R.font.instrument_medium, FontWeight.Medium),
    Font(R.font.instrument_semibold, FontWeight.SemiBold),
    Font(R.font.instrument_bold, FontWeight.Bold),
)

/** Un estilo de texto del handoff: tamaño, peso y tracking (en em). Escala con el tamaño de letra del sistema. */
data class Estilo(val tamano: Float, val peso: Peso, val tracking: Float = 0f) {
    /** El TextStyle de Compose. `numeros`: dígitos tabulares (.monospacedDigit()). */
    fun ts(numeros: Boolean = false): TextStyle = TextStyle(
        fontFamily = InstrumentSans,
        fontSize = tamano.sp,
        fontWeight = peso.peso,
        letterSpacing = tracking.em,
        fontFeatureSettings = if (numeros) "tnum" else null,
        lineBreak = LineBreak.Paragraph,
    )

    fun con(peso: Peso) = copy(peso = peso)

    companion object {
        val montoHeroe = Estilo(46f, Peso.bold, -0.035f)
        val montoCobro = Estilo(52f, Peso.bold, -0.035f)
        val montoPantalla = Estilo(40f, Peso.bold, -0.03f)
        val tituloGrande = Estilo(34f, Peso.bold, -0.025f)
        val saludo = Estilo(32f, Peso.bold, -0.025f)
        val tituloPaso = Estilo(30f, Peso.bold, -0.025f)
        val tituloBloqueo = Estilo(28f, Peso.bold, -0.02f)
        val montoMeta = Estilo(28f, Peso.bold, -0.025f)
        val tituloHoja = Estilo(24f, Peso.bold, -0.02f)
        val tituloHojaChico = Estilo(22f, Peso.bold, -0.02f)
        val tituloSeccion = Estilo(21f, Peso.bold, -0.02f)
        val indicador = Estilo(22f, Peso.bold, -0.02f)
        val nombreTienda = Estilo(18f, Peso.bold)
        val total = Estilo(18f, Peso.bold)
        val headline = Estilo(17f, Peso.semibold)
        val body = Estilo(17f, Peso.regular)
        val fila = Estilo(16f, Peso.semibold)
        val body16 = Estilo(16f, Peso.regular)
        val tarjetaTincho = Estilo(16f, Peso.medium)
        val subhead = Estilo(15f, Peso.regular)
        val subheadFuerte = Estilo(15f, Peso.semibold)
        val label = Estilo(14f, Peso.semibold)
        val texto14 = Estilo(14f, Peso.regular)
        val footnote = Estilo(13f, Peso.regular)
        val footnoteFuerte = Estilo(13f, Peso.semibold)
        val chip = Estilo(12f, Peso.semibold)
        val etiqueta = Estilo(11f, Peso.bold)
        val tab = Estilo(10.5f, Peso.semibold)
    }
}

/** Medidas del handoff. */
object M {
    /** Margen horizontal: 20 (16 en SE, 24 en Pro Max). */
    fun margen(ancho: Float): Dp = if (ancho <= 375f) 16.dp else if (ancho >= 428f) 24.dp else 20.dp
    val tabBar = 49.dp
}

/** El ancho de la pantalla en dp para los márgenes del handoff (anchoPantalla en iOS). */
val LocalAnchoPantalla = staticCompositionLocalOf { 393f }

/** El margen de la pantalla (20, o 16 en SE). */
val margen: Dp @Composable @ReadOnlyComposable get() = M.margen(LocalAnchoPantalla.current)

/** Los hápticos de iOS con su equivalente de Android. La vista se fija una vez en MainActivity. */
object Haptico {
    internal var vista: WeakReference<View>? = null
    private fun h(c: Int) { vista?.get()?.performHapticFeedback(c) }
    fun exito() = h(if (android.os.Build.VERSION.SDK_INT >= 30) HapticFeedbackConstants.CONFIRM else HapticFeedbackConstants.VIRTUAL_KEY)
    fun error() = h(if (android.os.Build.VERSION.SDK_INT >= 30) HapticFeedbackConstants.REJECT else HapticFeedbackConstants.LONG_PRESS)
    fun impacto() = h(HapticFeedbackConstants.VIRTUAL_KEY)
    fun seleccion() = h(if (android.os.Build.VERSION.SDK_INT >= 34) HapticFeedbackConstants.SEGMENT_TICK else HapticFeedbackConstants.CLOCK_TICK)
}
