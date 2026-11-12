package ar.com.tulocalweb.app.tema

import android.graphics.Bitmap
import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.drawBehind
import androidx.compose.ui.graphics.BlendMode
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.FilterQuality
import androidx.compose.ui.graphics.ImageBitmap
import androidx.compose.ui.graphics.ImageShader
import androidx.compose.ui.graphics.ShaderBrush
import androidx.compose.ui.graphics.Shape
import androidx.compose.ui.graphics.TileMode
import androidx.compose.ui.graphics.asImageBitmap
import androidx.compose.ui.graphics.drawOutline
import androidx.compose.ui.graphics.drawscope.DrawScope
import androidx.compose.ui.graphics.drawscope.translate
import androidx.compose.ui.graphics.toArgb
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.res.imageResource
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import androidx.compose.foundation.shape.RoundedCornerShape
import ar.com.tulocalweb.app.R
import com.google.zxing.BarcodeFormat
import com.google.zxing.EncodeHintType
import com.google.zxing.qrcode.QRCodeWriter
import com.google.zxing.qrcode.decoder.ErrorCorrectionLevel
import kotlin.math.pow
import kotlin.math.roundToInt

// Clon de Tema/Relieve.swift: piezas con cuerpo para las pantallas con más vida (el Inicio «La vidriera de tu local»,
// handoff docs/handoff/inicio-a): el grano, la tarjeta con luz arriba y sombra en capas, el vidrio sobre un color, la
// paleta que sale del color de la tienda y su QR.

// MARK: - Grano

/** El pincel del grano (drawable grano: ruido de 128 px en mosaico). */
@Composable
fun pincelGrano(): Brush {
    val img = ImageBitmap.imageResource(R.drawable.grano)
    return remember(img) { ShaderBrush(ImageShader(img, TileMode.Repeated, TileMode.Repeated)) }
}

/** Pinta el grano encima de lo que ya está dibujado (Grano en iOS). Sobre un color va en `Overlay` (0,2 en la
 *  vidriera); sobre el papel o un botón blanco, en `Multiply` (0,06). */
fun DrawScope.grano(pincel: Brush, opacidad: Float = 0.2f, mezcla: BlendMode = BlendMode.Overlay) {
    drawRect(pincel, alpha = opacidad, blendMode = mezcla)
}

/** El grano detrás del contenido (sobre el fondo que ya tenga el modificador de afuera). */
@Composable
fun Modifier.grano(opacidad: Float = 0.2f, mezcla: BlendMode = BlendMode.Overlay): Modifier {
    val pincel = pincelGrano()
    return this.drawBehind { grano(pincel, opacidad, mezcla) }
}

// MARK: - Relieve

/** La única receta de sombra en capas (la del handoff: `inset 0 1px 0 #FFF, 0 1px 0 #E3D6C8, 0 12px 24px -12px
 *  rgba(42,20,12,.22)`): una luz de 1 arriba, un filo de 1 abajo y una sombra larga y corrida hacia abajo. En oscuro, la
 *  luz es apenas un brillo y la sombra, negra. */
object Relieve {
    val luz = tema(0xFFFFFF, 0xFFFFFF, alfa = 1f, alfaOscuro = 0.1f)
    val anillo = tema(0xFFFFFF, 0xFFFFFF, alfa = 0f, alfaOscuro = 0.06f)
    val filo = tema(0xE3D6C8, 0x000000, alfaOscuro = 0.35f)
    val sombra = tema(0x2A140C, 0x000000, alfa = 0.14f, alfaOscuro = 0.45f)
    /** Radio y caída de la sombra: la de siempre y la de la tarjeta de Tincho (`0 16px 32px -12px .28`). */
    fun caida(fuerte: Boolean): Pair<Dp, Dp> = if (fuerte) 14.dp to 14.dp else 10.dp to 10.dp
}

/** Tarjeta con relieve: fondo, luz arriba, filo abajo y sombra en capas (radio 24 en tarjetas). `luz` en false para las
 *  tarjetas oscuras (la meta), que no llevan el brillo de arriba. */
@Composable
fun Modifier.relieve(radio: Dp = 24.dp, fondo: Color = C.superficie, fuerte: Boolean = false, luz: Boolean = true): Modifier {
    val forma = RoundedCornerShape(radio)
    val (r, y) = Relieve.caida(fuerte)
    val filo = Relieve.filo.actual
    val brillo = Relieve.luz.actual
    return this
        .sombra(forma, Relieve.sombra.actual, r, y = y)
        .drawBehind { translate(top = 1.dp.toPx()) { drawOutline(forma.createOutline(size, layoutDirection, this), filo) } }
        .clip(forma)
        .background(fondo)
        .border(1.dp, Relieve.anillo.actual, forma)
        .then(if (luz) Modifier.border(1.dp, Brush.verticalGradient(0f to brillo, 0.12f to brillo.copy(alpha = 0f)), forma) else Modifier)
}

/** Vidrio sobre el color de la tienda: el velo y un borde de 1 (`veloV` y `bordeV` del handoff). */
fun Modifier.vidrio(forma: Shape, paleta: PaletaTienda): Modifier = this.clip(forma).background(paleta.velo).border(1.dp, paleta.borde, forma)

// MARK: - Paleta de la tienda

/** Los colores que salen del de la tienda (`acento` y `sobreAcento` de /v1/panel/tablero), como los deriva el handoff:
 *  el velo, el borde, la tinta, el claro… En oscuro, la vidriera se apaga hacia el fondo y el texto encima es el del
 *  tema. Si el acento es claro (Pampa, #E6FF4D), el texto va oscuro. */
data class PaletaTienda(
    /** El acento y el texto que va encima de él (botones y la insignia en acento). */
    val acento: Color,
    val sobreAcento: Color,
    /** El fondo de la vidriera (en claro, el acento; en oscuro, un degradé hacia el fondo). */
    val fondoArriba: Color,
    val fondoAbajo: Color,
    /** El texto sobre la vidriera, el secundario y lo punteado. */
    val sobre: Color,
    val sobre2: Color,
    val sobreSuave: Color,
    val velo: Color,
    val borde: Color,
    /** Los halos (en oscuro, uno solo con el acento). */
    val luz: Color,
    val sombraHalo: Color,
    /** El acento para texto sobre blanco o superficie (links, «Preparar», el número de la fila). */
    val tinta: Color,
    /** La tinta sobre blanco siempre, en los dos modos (el QR y la miniatura, que son blancos). */
    val tintaBlanco: Color,
    /** Un fondo suave del acento (píldoras, insignias, «Preguntale»). */
    val claro: Color,
    val medio: Color,
    val meta: Color,
    /** La línea del gráfico de la vidriera. */
    val linea: Color,
    val oscuro: Boolean,
    /** Si el texto sobre el acento es claro (acento oscuro). */
    val claroEncima: Boolean,
) {
    companion object {
        const val porDefecto = 0xD9652FL

        operator fun invoke(acento: String?, sobre: String?, oscuro: Boolean): PaletaTienda {
            val a = leer(acento) ?: porDefecto
            val s = leer(sobre) ?: if (luminancia(a) < 0.3) 0xFFFFFFL else 0x1E0E12L
            val claroEncima = luminancia(s) > 0.5
            val blanco = hex(if (claroEncima) mezclar(a, 0x000000, 0.2) else mezclar(a, 0x000000, 0.6))
            val medio = hex(mezclar(a, 0xFFFFFF, 0.45))
            return if (oscuro) {
                val luzTienda = hex(mezclar(a, 0xFFFFFF, 0.5))
                PaletaTienda(
                    acento = hex(a), sobreAcento = hex(s),
                    fondoArriba = hex(mezclar(a, 0x22140D, 0.55)), fondoAbajo = hex(mezclar(a, 0x22140D, 0.82)),
                    sobre = hex(0xFBF7F2), sobre2 = hex(0xCDB8A8), sobreSuave = hex(0xFBF7F2, 0.35f),
                    velo = hex(0xFFFFFF, 0.06f), borde = hex(0xFFFFFF, 0.14f),
                    luz = hex(a, 0.28f), sombraHalo = Color.Transparent,
                    tinta = luzTienda, tintaBlanco = blanco, claro = hex(mezclar(a, 0x2E1B12, 0.6)), medio = medio,
                    meta = luzTienda, linea = luzTienda, oscuro = true, claroEncima = claroEncima,
                )
            } else {
                PaletaTienda(
                    acento = hex(a), sobreAcento = hex(s),
                    fondoArriba = hex(a), fondoAbajo = hex(a),
                    sobre = hex(s), sobre2 = hex(s, 0.88f), sobreSuave = hex(s, if (claroEncima) 0.55f else 0.45f),
                    velo = hex(s, if (claroEncima) 0.14f else 0.08f), borde = hex(s, if (claroEncima) 0.28f else 0.16f),
                    luz = hex(mezclar(a, 0xFFFFFF, 0.45), 0.55f), sombraHalo = hex(mezclar(a, 0x000000, 0.55), 0.6f),
                    tinta = blanco, tintaBlanco = blanco, claro = hex(mezclar(a, 0xFFFFFF, 0.86)), medio = medio,
                    meta = if (claroEncima) hex(mezclar(a, 0xFFFFFF, 0.45)) else hex(a), linea = hex(s),
                    oscuro = false, claroEncima = claroEncima,
                )
            }
        }

        /** «#RRGGBB» → 0xRRGGBB (null si no es un color). */
        fun leer(s: String?): Long? {
            if (s == null || s.length != 7 || !s.startsWith("#")) return null
            return s.substring(1).toLongOrNull(16)
        }

        /** Mezcla `t` (0–1) de `con` sobre `hex`, canal por canal (el `mezclar` del handoff). */
        fun mezclar(hex: Long, con: Long, t: Double): Long = listOf(16, 8, 0).fold(0L) { r, s ->
            val a = ((hex shr s) and 0xFF).toDouble()
            val b = ((con shr s) and 0xFF).toDouble()
            r or ((a * (1 - t) + b * t).roundToInt().toLong() shl s)
        }

        /** Luminancia relativa (WCAG 2). */
        fun luminancia(hex: Long): Double {
            val c = listOf(16, 8, 0).map { s ->
                val v = ((hex shr s) and 0xFF) / 255.0
                if (v <= 0.03928) v / 12.92 else ((v + 0.055) / 1.055).pow(2.4)
            }
            return 0.2126 * c[0] + 0.7152 * c[1] + 0.0722 * c[2]
        }
    }
}

// MARK: - QR

/** El QR de un texto (el link de la tienda) en `color` sobre transparente, generado en el teléfono con ZXing (la misma
 *  librería que QRTienda) y nítido a cualquier tamaño. */
object CodigoQR {
    /** Los módulos en `color` y el resto transparente, un píxel por módulo. */
    fun mascara(texto: String, color: Int): ImageBitmap? = runCatching {
        val m = QRCodeWriter().encode(texto, BarcodeFormat.QR_CODE, 0, 0, mapOf(EncodeHintType.ERROR_CORRECTION to ErrorCorrectionLevel.M, EncodeHintType.MARGIN to 0))
        val px = IntArray(m.width * m.height) { i -> if (m[i % m.width, i / m.width]) color else 0 }
        Bitmap.createBitmap(px, m.width, m.height, Bitmap.Config.ARGB_8888).asImageBitmap()
    }.getOrNull()
}

@Composable
fun CodigoQR(texto: String, color: Color, modifier: Modifier = Modifier, descripcion: String? = null) {
    val img = remember(texto, color) { CodigoQR.mascara(texto, color.toArgb()) } ?: return
    Image(img, null, modifier.then(if (descripcion != null) Modifier.semantics { contentDescription = descripcion } else Modifier),
        contentScale = ContentScale.Fit, filterQuality = FilterQuality.None)
}
