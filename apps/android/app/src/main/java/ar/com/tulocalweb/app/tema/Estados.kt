package ar.com.tulocalweb.app.tema

import androidx.compose.animation.core.LinearEasing
import androidx.compose.animation.core.animateFloat
import androidx.compose.animation.core.infiniteRepeatable
import androidx.compose.animation.core.rememberInfiniteTransition
import androidx.compose.animation.core.tween
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.drawBehind
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.nucleo.ErrorApi
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.Tincho
import ar.com.tulocalweb.app.nucleo.abrirEnNavegador
import java.time.Instant
import java.util.UUID

// Clon de Tema/Estados.swift: los estados de cada pantalla (handoff: «Interactions & Behavior · Estados»): cargando
// con esqueletos de forma real (2c), vacío con Tincho y la primera acción (3d), error con Tincho y «Reintentar» (12f),
// sin conexión con el banner y la hora de lo guardado (5a).

/** Lo que puede pasar al traer algo de la API. */
sealed class Carga<out T> {
    data object cargando : Carga<Nothing>()
    data class listo<T>(val v: T) : Carga<T>()
    data class error(val e: ErrorApi) : Carga<Nothing>()

    val valor: T? get() = (this as? listo<T>)?.v
}

/** Tincho con un título, un texto y una o dos acciones. */
@Composable
fun EstadoTincho(
    tincho: Tincho,
    titulo: String,
    texto: String,
    tam: Dp = 190.dp,
    secundaria: String? = null,
    alTocarSecundaria: (() -> Unit)? = null,
    accion: String? = null,
    principal: Boolean = false,
    alTocar: (() -> Unit)? = null,
) {
    Column(Modifier.fillMaxWidth().padding(horizontal = 36.dp), horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(12.dp)) {
        ImagenTincho(tincho, Modifier.size(tam))
        Texto(Formato.sinViudas(titulo), Estilo.tituloHojaChico, C.texto, alinear = TextAlign.Center)
        Texto(Formato.sinViudas(texto), Estilo.subhead, C.texto2, alinear = TextAlign.Center, interlineado = 2f)
        if (accion != null && alTocar != null) {
            Box(Modifier.padding(top = 8.dp)) {
                if (principal) BotonPrincipal(accion, alTocar, alto = 48.dp, lleno = false)
                else BotonSecundario(accion, alTocar)
            }
        }
        if (secundaria != null && alTocarSecundaria != null) {
            Texto(secundaria, Estilo.subheadFuerte, C.primarioTexto, Modifier.padding(top = 4.dp).clickable(onClick = alTocarSecundaria))
        }
    }
}

/** El error de traer una lista (12f): «No pudimos traer los pedidos». */
@Composable
fun EstadoError(que: String, error: ErrorApi, reintentar: () -> Unit) {
    val ctx = LocalContext.current
    if (error.sinConexion) {
        EstadoTincho(Tincho.desenchufado, "Sin conexión", "No hay internet. Cuando vuelva, probá de nuevo.",
            accion = "Reintentar", principal = true, alTocar = reintentar)
    } else {
        EstadoTincho(Tincho.preocupado, "No pudimos traer $que", "Puede ser un problema nuestro. Probá de nuevo en un ratito.",
            secundaria = "Ver estado del servicio", alTocarSecundaria = { abrirEnNavegador(ctx, "https://tulocalweb.com.ar/estado") },
            accion = "Reintentar", principal = true, alTocar = reintentar)
    }
}

/** Banner `warn-bg` de sin conexión (5a). */
@Composable
fun BannerSinConexion(guardado: Instant) {
    Row(
        Modifier.fillMaxWidth().clip(RoundedCornerShape(12.dp)).background(C.avisoFondo).padding(horizontal = 12.dp, vertical = 9.dp),
        horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.CenterVertically,
    ) {
        Icono("wifi.slash", 14f, C.bannerTinta)
        Texto("Sin conexión. Ves lo guardado ${Formato.cuando(guardado)}.", Estilo.texto14, C.bannerTinta, Modifier.weight(1f))
    }
}

/** Bloque de esqueleto con el brillo horizontal de 1,2 s (respeta Reducir movimiento). */
@Composable
fun Esqueleto(ancho: Dp? = null, alto: Dp, radio: Dp? = null, suave: Boolean = false, modifier: Modifier = Modifier) {
    val base = if (suave) C.esqueleto2 else C.esqueleto1
    val brillo = C.elevado.copy(alpha = 0.9f)
    val anima = animacionesActivas()
    val fase = if (anima) rememberInfiniteTransition(label = "esqueleto").animateFloat(-1f, 1f, infiniteRepeatable(tween(1200, easing = LinearEasing)), label = "f").value else -1f
    Box(
        modifier.then(if (ancho != null) Modifier.width(ancho) else Modifier.fillMaxWidth()).height(alto)
            .clip(RoundedCornerShape(radio ?: (alto / 2))).background(base)
            .drawBehind {
                if (anima) drawRect(Brush.horizontalGradient(listOf(base, brillo, base), startX = size.width * fase, endX = size.width * (fase + 1)))
            },
    )
}

/** Filas de lista mientras cargan. */
@Composable
fun EsqueletoFilas(filas: Int = 4, miniatura: Dp = 48.dp, modifier: Modifier = Modifier) {
    Column(
        modifier.fillMaxWidth().tarjeta().padding(horizontal = 16.dp, vertical = 14.dp).semantics { contentDescription = "Cargando" },
        verticalArrangement = Arrangement.spacedBy(18.dp),
    ) {
        repeat(filas) { i ->
            Row(horizontalArrangement = Arrangement.spacedBy(13.dp), verticalAlignment = Alignment.CenterVertically) {
                Esqueleto(miniatura, miniatura, 12.dp)
                BoxWithConstraints(Modifier.weight(1f).heightIn(min = 31.dp)) {
                    Column(verticalArrangement = Arrangement.spacedBy(7.dp)) {
                        Esqueleto(this@BoxWithConstraints.maxWidth * (if (i % 2 == 0) 0.7f else 0.8f), 13.dp)
                        Esqueleto(this@BoxWithConstraints.maxWidth * (if (i % 2 == 0) 0.45f else 0.35f), 11.dp, suave = true)
                    }
                }
            }
        }
    }
}

// MARK: - Toast

/** Toast inferior oscuro (52 de alto, 4 s) con «Deshacer» cuando aplica. */
data class Aviso(val texto: String, val ok: Boolean = true, val deshacer: (() -> Unit)? = null, val id: String = UUID.randomUUID().toString())

@Composable
fun VistaAviso(aviso: Aviso, cerrar: () -> Unit) {
    val forma = RoundedCornerShape(16.dp)
    Row(
        Modifier.fillMaxWidth().sombra(forma, hex(0x2A140C, 0.4f), 12.dp, y = 10.dp).clip(forma).background(tema(0x2A140C, 0x3A2419).actual)
            .heightIn(min = 52.dp).padding(horizontal = 16.dp),
        horizontalArrangement = Arrangement.spacedBy(10.dp), verticalAlignment = Alignment.CenterVertically,
    ) {
        Box(Modifier.size(22.dp).clip(CircleShape).background(if (aviso.ok) hex(0x3FA877) else C.error), contentAlignment = Alignment.Center) {
            Icono(if (aviso.ok) "checkmark" else "exclamationmark", 11f, Color.White)
        }
        Texto(aviso.texto, Estilo.subhead, hex(0xFBF7F2), Modifier.weight(1f).padding(vertical = 8.dp), lineas = 2)
        aviso.deshacer?.let { d ->
            Texto("Deshacer", Estilo(15f, Peso.bold), hex(0xF3A57E), Modifier.clickable { d(); cerrar() }.padding(vertical = 12.dp))
        }
    }
}

/** Un punto de luz que late (los «en vivo» y el «Escuchando…» de Tincho). */
@Composable
fun Pulso(color: Color, tam: Dp = 8.dp) {
    val anima = animacionesActivas()
    val a = if (anima) rememberInfiniteTransition(label = "pulso").animateFloat(0.35f, 1f, infiniteRepeatable(tween(900), androidx.compose.animation.core.RepeatMode.Reverse), label = "a").value else 1f
    Box(Modifier.size(tam).clip(CircleShape).background(color.copy(alpha = a)))
}

/** Sombra del círculo (para centrar algo con un brillo detrás). */
fun Modifier.brilloRadial(color: Color, centro: Offset = Offset(1f, 0f), radio: Float = 260f): Modifier = drawBehind {
    drawRect(Brush.radialGradient(listOf(color, Color.Transparent), Offset(size.width * centro.x, size.height * centro.y), radio * density))
}

/** ProgressView(): la ruedita que gira. */
@Composable
fun Rueda(color: Color = C.texto3, tam: Dp = 20.dp) {
    val giro = rememberInfiniteTransition(label = "rueda").animateFloat(0f, 360f, infiniteRepeatable(tween(900, easing = LinearEasing)), label = "g").value
    androidx.compose.foundation.Canvas(Modifier.size(tam).semantics { contentDescription = "Cargando" }) {
        val g = size.minDimension * 0.12f
        drawArc(color.copy(alpha = 0.25f), 0f, 360f, false, style = androidx.compose.ui.graphics.drawscope.Stroke(g))
        drawArc(color, giro, 100f, false, style = androidx.compose.ui.graphics.drawscope.Stroke(g, cap = androidx.compose.ui.graphics.StrokeCap.Round))
    }
}
