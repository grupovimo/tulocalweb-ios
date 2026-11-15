package ar.com.tulocalweb.app.pantallas.mas

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxScope
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ColumnScope
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.RowScope
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.tema.*

// Clon de Pantallas/Mas/MiniaturasBloques.swift.

/** El dibujo de cada variante (docs/handoff/tienda-bloques/Bloques.dc.html): una tarjeta blanca con rayas y cajas
 *  grises. Es el mismo dibujo en claro y en oscuro, como un papelito de la tienda. */
enum class Dibujo { foto, partido, partidoIzquierda, collage, texto, carrusel, banda, contador, barra, dos, iconos, grilla, look, lista, mapa, acordeon, tarjetas, logos, input, circulos, video }

/** Qué dibujo lleva cada variante, y para las grillas cuántas columnas y celdas. */
data class MiniaturaVariante(val dibujo: Dibujo, val columnas: Int = 0, val celdas: Int = 0) {
    companion object {
        fun de(tipo: String, variante: String): MiniaturaVariante {
            fun g(c: Int, k: Int) = MiniaturaVariante(Dibujo.grilla, c, k)
            val d = Dibujo.entries
            return when ("$tipo.$variante") {
                "portada.partida" -> MiniaturaVariante(Dibujo.partido)
                "banner.doble" -> MiniaturaVariante(Dibujo.dos)
                "oferta.productos" -> MiniaturaVariante(Dibujo.contador)
                "barra.fija", "barra.rotativa" -> MiniaturaVariante(Dibujo.barra)
                "textoImagen.derecha" -> MiniaturaVariante(Dibujo.partido)
                "textoImagen.izquierda" -> MiniaturaVariante(Dibujo.partidoIzquierda)
                "textoImagen.fondo" -> MiniaturaVariante(Dibujo.foto)
                "video.ancho" -> MiniaturaVariante(Dibujo.video)
                "video.conTexto" -> MiniaturaVariante(Dibujo.partido)
                "beneficios.fila" -> MiniaturaVariante(Dibujo.iconos)
                "productos.grilla" -> g(4, 4)
                "combo.tarjeta" -> MiniaturaVariante(Dibujo.look)
                "categorias.tarjetas" -> g(3, 3)
                "menu.columnas" -> g(2, 4)
                "local.tarjeta" -> MiniaturaVariante(Dibujo.tarjetas)
                "resenas.cita" -> MiniaturaVariante(Dibujo.texto)
                "logos.fila" -> MiniaturaVariante(Dibujo.logos)
                "galeria.fila" -> g(6, 6)
                "galeria.mosaico" -> g(3, 6)
                "newsletter.centrado" -> MiniaturaVariante(Dibujo.input)
                "whatsapp.tarjeta" -> MiniaturaVariante(Dibujo.tarjetas)
                "catalogo.grilla" -> g(4, 8)
                "producto.grilla" -> g(2, 4)
                "producto.miniaturas" -> MiniaturaVariante(Dibujo.partidoIzquierda)
                "carrito.pagina" -> MiniaturaVariante(Dibujo.lista)
                "carrito.panel" -> MiniaturaVariante(Dibujo.dos)
                else -> MiniaturaVariante(d.firstOrNull { it.name == variante } ?: Dibujo.tarjetas)
            }
        }
    }
}

private val papel = hex(0xFFFFFF)
private val bordePapel = hex(0xE2DFD8)
private val caja = hex(0xD7D3CA)
private val raya = hex(0x3A3936)
private val rayaSuave = hex(0xB9B5AC)
private val gris = hex(0x8B887F)
private val marca = hex(0xE7B04A)

@Composable private fun Caja(m: Modifier, color: Color = caja, radio: Dp = 3.dp) = Box(m.clip(RoundedCornerShape(radio)).background(color))
@Composable private fun Raya(ancho: Float = 1f, alto: Dp = 5.dp, color: Color = raya) = Box(Modifier.fillMaxWidth(ancho).height(alto).clip(RoundedCornerShape(3.dp)).background(color))
@Composable private fun Suave(ancho: Float = 1f, color: Color = rayaSuave) = Raya(ancho, 4.dp, color)
@Composable private fun Boton(ancho: Dp = 30.dp, color: Color = raya) = Box(Modifier.size(ancho, 9.dp).clip(RoundedCornerShape(2.dp)).background(color))

/** La miniatura de una variante. `alto` es el de la tarjeta; el ancho lo pone quien la usa. */
@Composable
fun MiniaturaBloque(tipo: String, variante: String, modifier: Modifier = Modifier, alto: Dp = 72.dp, elegida: Boolean = false) {
    val m = MiniaturaVariante.de(tipo, variante)
    val forma = RoundedCornerShape(8.dp)
    Box(modifier.height(alto).clip(forma).background(papel).border(if (elegida) 2.dp else 1.dp, if (elegida) C.acento else bordePapel, forma).padding(7.dp)) {
        Dibujar(m)
    }
}

@Composable private fun Fila(c: @Composable RowScope.() -> Unit) = Row(Modifier.fillMaxSize(), horizontalArrangement = Arrangement.spacedBy(5.dp), content = c)
@Composable private fun Columna(m: Modifier, v: Arrangement.Vertical, h: Alignment.Horizontal, c: @Composable ColumnScope.() -> Unit) =
    Column(m, verticalArrangement = v, horizontalAlignment = h, content = c)

@Composable
private fun BoxScope.Dibujar(m: MiniaturaVariante) {
    when (m.dibujo) {
        Dibujo.foto -> {
            Caja(Modifier.fillMaxSize())
            Column(Modifier.align(Alignment.BottomStart).padding(6.dp).fillMaxWidth(0.5f), verticalArrangement = Arrangement.spacedBy(3.dp)) {
                Raya(1f, 6.dp, papel); Raya(0.7f, 5.dp, papel); Boton(color = papel)
            }
        }
        Dibujo.partido, Dibujo.partidoIzquierda -> Fila {
            if (m.dibujo == Dibujo.partidoIzquierda) Caja(Modifier.weight(1f).fillMaxHeight())
            Columna(Modifier.weight(1f).fillMaxHeight(), Arrangement.spacedBy(3.dp, Alignment.CenterVertically), Alignment.Start) {
                Raya(1f, 6.dp); Raya(0.7f); Suave(0.85f); Spacer(Modifier.height(2.dp)); Boton()
            }
            if (m.dibujo == Dibujo.partido) Caja(Modifier.weight(1f).fillMaxHeight())
        }
        Dibujo.collage -> Fila {
            Columna(Modifier.weight(1.2f).fillMaxHeight().clip(RoundedCornerShape(3.dp)).background(raya).padding(5.dp), Arrangement.spacedBy(3.dp, Alignment.Bottom), Alignment.Start) {
                Raya(1f, 5.dp, papel); Boton(24.dp, marca)
            }
            Caja(Modifier.weight(1f).fillMaxHeight())
            Columna(Modifier.weight(0.8f).fillMaxHeight(), Arrangement.spacedBy(4.dp), Alignment.Start) {
                Caja(Modifier.weight(1f).fillMaxWidth()); Caja(Modifier.weight(1f).fillMaxWidth(), marca)
            }
        }
        Dibujo.texto -> Columna(Modifier.fillMaxSize(), Arrangement.spacedBy(4.dp, Alignment.CenterVertically), Alignment.CenterHorizontally) {
            Raya(0.8f, 7.dp); Raya(0.55f, 7.dp); Suave(0.6f); Boton()
        }
        Dibujo.carrusel -> {
            Caja(Modifier.fillMaxSize())
            Row(Modifier.align(Alignment.BottomCenter).padding(bottom = 6.dp), horizontalArrangement = Arrangement.spacedBy(3.dp)) {
                Box(Modifier.size(10.dp, 3.dp).clip(CircleShape).background(raya)); Box(Modifier.size(3.dp).clip(CircleShape).background(gris)); Box(Modifier.size(3.dp).clip(CircleShape).background(gris))
            }
            Texto("‹", Estilo(11f, Peso.semibold), raya, Modifier.align(Alignment.CenterStart).padding(start = 3.dp))
            Texto("›", Estilo(11f, Peso.semibold), raya, Modifier.align(Alignment.CenterEnd).padding(end = 3.dp))
        }
        Dibujo.banda -> Row(Modifier.fillMaxSize().clip(RoundedCornerShape(5.dp)).background(raya).padding(9.dp), verticalAlignment = Alignment.CenterVertically) {
            Column(Modifier.weight(0.55f), verticalArrangement = Arrangement.spacedBy(3.dp)) { Raya(1f, 6.dp, papel); Suave(1f, gris) }
            Spacer(Modifier.weight(0.45f)); Boton(color = marca)
        }
        Dibujo.contador -> Columna(Modifier.fillMaxSize(), Arrangement.spacedBy(5.dp), Alignment.Start) {
            Row(horizontalArrangement = Arrangement.spacedBy(3.dp), verticalAlignment = Alignment.CenterVertically) {
                Box(Modifier.weight(1f)) { Raya(0.6f) }
                repeat(3) { Caja(Modifier.size(11.dp, 10.dp), raya, 2.dp) }
            }
            Row(Modifier.weight(1f), horizontalArrangement = Arrangement.spacedBy(4.dp)) { repeat(4) { Caja(Modifier.weight(1f).fillMaxHeight()) } }
        }
        Dibujo.barra -> Columna(Modifier.fillMaxSize(), Arrangement.spacedBy(4.dp), Alignment.Start) {
            Box(Modifier.fillMaxWidth().height(13.dp).clip(RoundedCornerShape(3.dp)).background(raya), contentAlignment = Alignment.Center) { Box(Modifier.fillMaxWidth(0.5f)) { Suave(1f, papel) } }
            Box(Modifier.fillMaxWidth().height(10.dp).border(width = 0.dp, color = Color.Transparent)) { Box(Modifier.align(Alignment.BottomStart).fillMaxWidth().height(1.dp).background(bordePapel)) }
            Caja(Modifier.weight(1f).fillMaxWidth().alpha(0.5f))
        }
        Dibujo.dos -> Fila { Caja(Modifier.weight(1f).fillMaxHeight()); Caja(Modifier.weight(1f).fillMaxHeight(), raya) }
        Dibujo.iconos -> Row(Modifier.fillMaxSize(), horizontalArrangement = Arrangement.spacedBy(4.dp), verticalAlignment = Alignment.CenterVertically) {
            repeat(4) {
                Column(Modifier.weight(1f), horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(3.dp)) {
                    Box(Modifier.size(13.dp).clip(RoundedCornerShape(4.dp)).border(1.5.dp, raya, RoundedCornerShape(4.dp))); Suave(0.8f)
                }
            }
        }
        Dibujo.grilla -> {
            val filas = ((m.celdas + m.columnas - 1) / m.columnas).coerceAtLeast(1)
            Column(Modifier.fillMaxSize(), verticalArrangement = Arrangement.spacedBy(4.dp)) {
                repeat(filas) { f ->
                    Row(Modifier.weight(1f), horizontalArrangement = Arrangement.spacedBy(4.dp)) {
                        repeat(m.columnas) { c ->
                            if (f * m.columnas + c < m.celdas) Column(Modifier.weight(1f).fillMaxHeight(), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                                Caja(Modifier.weight(1f).fillMaxWidth()); if (filas == 1) { Suave(); Raya(0.5f, 4.dp) }
                            } else Spacer(Modifier.weight(1f))
                        }
                    }
                }
            }
        }
        Dibujo.look -> Fila {
            repeat(3) { Caja(Modifier.weight(1f).fillMaxHeight()) }
            Columna(Modifier.weight(1f).fillMaxHeight().border(1.5.dp, raya, RoundedCornerShape(3.dp)).padding(3.dp), Arrangement.spacedBy(3.dp, Alignment.Bottom), Alignment.Start) {
                Raya(0.7f, 4.dp); Box(Modifier.fillMaxWidth().height(7.dp).clip(RoundedCornerShape(2.dp)).background(raya))
            }
        }
        Dibujo.lista -> Column(Modifier.fillMaxSize(), verticalArrangement = Arrangement.spacedBy(4.dp, Alignment.CenterVertically)) {
            listOf(0.55f, 0.45f, 0.6f).forEach { a ->
                Row(Modifier.weight(1f), horizontalArrangement = Arrangement.spacedBy(4.dp), verticalAlignment = Alignment.CenterVertically) {
                    Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) { Raya(a, 4.dp); Suave(a + 0.25f) }
                    Caja(Modifier.fillMaxHeight().width(18.dp))
                }
            }
        }
        Dibujo.mapa -> Fila {
            Columna(Modifier.weight(1f).fillMaxHeight(), Arrangement.spacedBy(3.dp, Alignment.CenterVertically), Alignment.Start) { Raya(1f, 6.dp); Suave(); Suave(0.7f); Suave(0.8f) }
            Box(Modifier.weight(1.2f).fillMaxHeight().clip(RoundedCornerShape(3.dp)).background(hex(0xE6E1D6))) {
                Row(Modifier.fillMaxSize(), horizontalArrangement = Arrangement.SpaceEvenly) { repeat(3) { Box(Modifier.fillMaxHeight().width(2.dp).background(papel)) } }
                Column(Modifier.fillMaxSize(), verticalArrangement = Arrangement.SpaceEvenly) { repeat(2) { Box(Modifier.fillMaxWidth().height(2.dp).background(papel)) } }
                Box(Modifier.align(Alignment.Center).size(8.dp).clip(CircleShape).background(raya))
            }
        }
        Dibujo.acordeon -> Column(Modifier.fillMaxSize(), verticalArrangement = Arrangement.SpaceEvenly) {
            listOf(0.6f to "−", 0.5f to "+", 0.65f to "+").forEach { (a, s) ->
                Row(verticalAlignment = Alignment.CenterVertically) { Box(Modifier.weight(1f)) { Raya(a, 4.dp) }; Texto(s, Estilo(9f, Peso.semibold), raya) }
                Box(Modifier.fillMaxWidth().height(1.dp).background(bordePapel))
            }
        }
        Dibujo.tarjetas -> Fila {
            repeat(3) {
                Columna(Modifier.weight(1f).fillMaxHeight().border(1.dp, bordePapel, RoundedCornerShape(3.dp)).padding(4.dp), Arrangement.spacedBy(3.dp), Alignment.Start) {
                    Suave(0.5f, marca); Raya(1f, 4.dp); Suave(); Suave(0.6f)
                }
            }
        }
        Dibujo.logos -> Columna(Modifier.fillMaxSize(), Arrangement.spacedBy(8.dp, Alignment.CenterVertically), Alignment.CenterHorizontally) {
            Suave(0.4f)
            Row(horizontalArrangement = Arrangement.spacedBy(6.dp)) { listOf(20, 16, 22, 17).forEach { Caja(Modifier.size(it.dp, 8.dp), gris, 2.dp) } }
        }
        Dibujo.input -> Columna(Modifier.fillMaxSize(), Arrangement.spacedBy(5.dp, Alignment.CenterVertically), Alignment.CenterHorizontally) {
            Raya(0.6f, 6.dp); Suave(0.45f)
            Row(Modifier.fillMaxWidth(0.7f), horizontalArrangement = Arrangement.spacedBy(3.dp)) {
                Box(Modifier.weight(1f).height(10.dp).border(1.dp, gris, RoundedCornerShape(2.dp))); Box(Modifier.size(20.dp, 10.dp).clip(RoundedCornerShape(2.dp)).background(raya))
            }
        }
        Dibujo.circulos -> Row(Modifier.fillMaxSize(), horizontalArrangement = Arrangement.SpaceAround, verticalAlignment = Alignment.CenterVertically) {
            repeat(4) {
                Column(horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(3.dp)) {
                    Box(Modifier.size(24.dp).clip(CircleShape).background(caja)); Box(Modifier.width(18.dp)) { Suave() }
                }
            }
        }
        Dibujo.video -> Box(Modifier.fillMaxSize().clip(RoundedCornerShape(3.dp)).background(gris), contentAlignment = Alignment.Center) {
            Box(Modifier.size(20.dp).clip(CircleShape).background(papel), contentAlignment = Alignment.Center) { Texto("▶", Estilo(8f, Peso.semibold), raya) }
        }
    }
}
