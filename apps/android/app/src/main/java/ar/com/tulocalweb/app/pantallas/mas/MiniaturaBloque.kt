package ar.com.tulocalweb.app.pantallas.mas

import androidx.compose.foundation.Canvas
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
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.geometry.Size
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.semantics.clearAndSetSemantics
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.tema.*

// Clon de Pantallas/Mas/MiniaturaBloque.swift.

/** La miniatura de un bloque o de una variante, dibujada como en docs/handoff/tienda-bloques/Bloques.dc.html: rayas
 *  para los textos, bloques grises para las fotos y el acento ámbar para los botones de promo. */
@Composable
fun MiniaturaBloque(tipo: String, variante: String, ancho: Dp = 120.dp, modifier: Modifier = Modifier) {
    val alto = (ancho.value * 0.66f).let { Math.round(it).toFloat() }.dp
    val m = MiniaturaPiezas(ancho - 12.dp, alto - 12.dp, tipo)
    val forma = RoundedCornerShape(8.dp)
    Box(modifier.size(ancho, alto).clip(forma).background(Color.White).border(1.dp, MiniaturaPiezas.borde, forma).padding(6.dp).clearAndSetSemantics { }) {
        with(m) { Dibujar(MiniaturaPiezas.forma(tipo, variante)) }
    }
}

/** Qué dibujo le toca a cada variante (los nombres de Bloques.dc.html). */
sealed class Dibujo {
    object foto : Dibujo(); object partido : Dibujo(); object partidoIzquierda : Dibujo(); object collage : Dibujo(); object texto : Dibujo()
    object carrusel : Dibujo(); object banda : Dibujo(); object contador : Dibujo(); object dos : Dibujo(); object iconos : Dibujo()
    data class grilla(val columnas: Int, val total: Int) : Dibujo()
    object look : Dibujo(); object lista : Dibujo(); object mapa : Dibujo(); object acordeon : Dibujo(); object tarjetas : Dibujo(); object logos : Dibujo()
    object input : Dibujo(); object circulos : Dibujo(); object video : Dibujo(); object miniaturas : Dibujo(); object panel : Dibujo()
}

/** Las piezas de las miniaturas (rayas, fotos, botones) medidas contra el ancho de adentro. */
private class MiniaturaPiezas(val adentro: Dp, val h: Dp, val tipo: String) {
    companion object {
        val foto = hex(0xD7D3CA)
        val linea = hex(0x3A3936)
        val suave = hex(0xB9B5AC)
        val gris = hex(0x8B887F)
        val ambar = hex(0xE7B04A)
        val borde = hex(0xE2DFD8)

        fun forma(tipo: String, variante: String): Dibujo = when {
            tipo == "textoImagen" && variante == "izquierda" -> Dibujo.partidoIzquierda
            variante == "partida" || variante == "partido" || variante == "derecha" || (tipo == "video" && variante == "conTexto") -> Dibujo.partido
            tipo == "video" -> Dibujo.video
            tipo == "combo" -> Dibujo.look
            variante == "foto" || variante == "fondo" -> Dibujo.foto
            variante == "collage" -> Dibujo.collage
            tipo == "newsletter" && variante == "centrado" -> Dibujo.input
            variante == "texto" || variante == "cita" || variante == "centrado" -> Dibujo.texto
            tipo == "catalogo" && variante == "grilla" -> Dibujo.grilla(4, 8)
            tipo == "producto" && variante == "grilla" -> Dibujo.grilla(2, 4)
            tipo == "producto" && variante == "miniaturas" -> Dibujo.miniaturas
            tipo == "carrito" && variante == "panel" -> Dibujo.panel
            tipo == "carrito" && variante == "pagina" -> Dibujo.lista
            variante == "carrusel" -> Dibujo.carrusel
            tipo == "oferta" && variante == "productos" -> Dibujo.contador
            variante == "banda" || tipo == "barra" -> Dibujo.banda
            variante == "doble" -> Dibujo.dos
            tipo == "beneficios" && variante == "fila" -> Dibujo.iconos
            tipo == "productos" && variante == "grilla" -> Dibujo.grilla(4, 4)
            tipo == "categorias" && variante == "tarjetas" -> Dibujo.grilla(3, 3)
            tipo == "menu" && variante == "columnas" -> Dibujo.grilla(2, 4)
            tipo == "galeria" && variante == "fila" -> Dibujo.grilla(6, 6)
            tipo == "galeria" && variante == "mosaico" -> Dibujo.grilla(3, 6)
            variante == "tarjetas" || variante == "tarjeta" -> Dibujo.tarjetas
            variante == "circulos" -> Dibujo.circulos
            variante == "mapa" -> Dibujo.mapa
            variante == "acordeon" -> Dibujo.acordeon
            tipo == "logos" -> Dibujo.logos
            else -> Dibujo.lista
        }
    }

    // MARK: Piezas

    @Composable fun Linea(f: Float, alto: Dp = 5.dp, color: Color = linea) = Box(Modifier.size(adentro * f, alto).clip(CircleShape).background(color))
    @Composable fun Suave(f: Float) = Linea(f, 4.dp, suave)
    @Composable fun Boton(color: Color = linea) = Box(Modifier.size(adentro * 0.3f, 8.dp).clip(RoundedCornerShape(2.dp)).background(color))
    @Composable fun Foto(m: Modifier) = Box(m.clip(RoundedCornerShape(3.dp)).background(foto))
    @Composable fun Caja(m: Modifier, color: Color, radio: Dp = 3.dp) = Box(m.clip(RoundedCornerShape(radio)).background(color))

    @Composable fun Textos(m: Modifier, f: Float = 0.45f, boton: Boolean = true) {
        Column(m, verticalArrangement = Arrangement.spacedBy(4.dp, Alignment.CenterVertically)) {
            Linea(f, 7.dp); Linea(f * 0.7f); Suave(f * 0.85f)
            if (boton) Box(Modifier.padding(top = 2.dp)) { Boton() }
        }
    }

    @Composable fun VideoCaja(m: Modifier) {
        Box(m.clip(RoundedCornerShape(3.dp)).background(gris), contentAlignment = Alignment.Center) {
            Box(Modifier.size(20.dp).clip(CircleShape).background(Color.White), contentAlignment = Alignment.Center) { Texto("▶", Estilo(8f, Peso.semibold), linea) }
        }
    }

    /** Celdas de producto (foto, nombre y precio) o de foto sola, en filas completas. */
    @Composable fun Celdas(cols: Int, total: Int, alto: Dp) {
        val filas = maxOf(1, total / cols)
        val conTexto = cols <= 4 && tipo != "galeria"
        Column(Modifier.fillMaxWidth().height(alto), verticalArrangement = Arrangement.spacedBy(4.dp)) {
            repeat(filas) {
                Row(Modifier.weight(1f), horizontalArrangement = Arrangement.spacedBy(4.dp)) {
                    repeat(cols) {
                        Column(Modifier.weight(1f).fillMaxHeight(), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                            Foto(Modifier.weight(1f).fillMaxWidth())
                            if (conTexto) {
                                Box(Modifier.fillMaxWidth().height(3.dp).clip(CircleShape).background(suave))
                                Box(Modifier.size(adentro / cols * 0.5f, 3.dp).clip(CircleShape).background(linea))
                            }
                        }
                    }
                }
            }
        }
    }

    // MARK: Dibujos

    @Composable fun BoxScope.Dibujar(d: Dibujo) {
        when (d) {
            Dibujo.foto -> {
                Foto(Modifier.fillMaxSize())
                Column(Modifier.align(Alignment.BottomStart).padding(8.dp), verticalArrangement = Arrangement.spacedBy(4.dp)) {
                    Linea(0.45f, 6.dp, Color.White); Linea(0.3f, color = Color.White); Boton(Color.White)
                }
            }
            Dibujo.partido, Dibujo.partidoIzquierda -> Row(Modifier.fillMaxSize(), horizontalArrangement = Arrangement.spacedBy(5.dp)) {
                if (d == Dibujo.partidoIzquierda) Foto(Modifier.weight(1f).fillMaxHeight())
                Textos(Modifier.weight(1f).fillMaxHeight())
                if (d == Dibujo.partido) { if (tipo == "video") VideoCaja(Modifier.weight(1f).fillMaxHeight()) else Foto(Modifier.weight(1f).fillMaxHeight()) }
            }
            Dibujo.collage -> Row(Modifier.fillMaxSize(), horizontalArrangement = Arrangement.spacedBy(4.dp)) {
                Box(Modifier.weight(1f).fillMaxHeight().clip(RoundedCornerShape(3.dp)).background(linea)) {
                    Column(Modifier.align(Alignment.BottomStart).padding(5.dp), verticalArrangement = Arrangement.spacedBy(3.dp)) {
                        Box(Modifier.size(adentro * 0.25f, 4.dp).clip(CircleShape).background(Color.White))
                        Box(Modifier.size(adentro * 0.18f, 6.dp).clip(RoundedCornerShape(2.dp)).background(ambar))
                    }
                }
                Foto(Modifier.weight(1f).fillMaxHeight())
                Column(Modifier.weight(1f).fillMaxHeight(), verticalArrangement = Arrangement.spacedBy(4.dp)) {
                    Foto(Modifier.weight(1f).fillMaxWidth()); Caja(Modifier.weight(1f).fillMaxWidth(), ambar)
                }
            }
            Dibujo.texto -> Column(Modifier.fillMaxSize(), verticalArrangement = Arrangement.spacedBy(5.dp, Alignment.CenterVertically), horizontalAlignment = Alignment.CenterHorizontally) {
                Linea(0.75f, 7.dp); Linea(0.5f, 7.dp); Suave(0.55f); Boton()
            }
            Dibujo.carrusel -> {
                Foto(Modifier.fillMaxSize())
                Row(Modifier.align(Alignment.BottomCenter).padding(bottom = 6.dp), horizontalArrangement = Arrangement.spacedBy(3.dp), verticalAlignment = Alignment.CenterVertically) {
                    Box(Modifier.size(12.dp, 3.dp).clip(CircleShape).background(linea))
                    Box(Modifier.size(3.dp).clip(CircleShape).background(gris)); Box(Modifier.size(3.dp).clip(CircleShape).background(gris))
                }
            }
            Dibujo.banda -> Row(Modifier.fillMaxSize().clip(RoundedCornerShape(5.dp)).background(linea).padding(10.dp), verticalAlignment = Alignment.CenterVertically) {
                Column(verticalArrangement = Arrangement.spacedBy(4.dp)) { Linea(0.4f, 6.dp, Color.White); Linea(0.3f, 4.dp, gris) }
                Spacer(Modifier.weight(1f)); Boton(ambar)
            }
            Dibujo.contador -> Column(Modifier.fillMaxSize(), verticalArrangement = Arrangement.spacedBy(5.dp)) {
                Row(horizontalArrangement = Arrangement.spacedBy(3.dp), verticalAlignment = Alignment.CenterVertically) {
                    Linea(0.35f); Spacer(Modifier.weight(1f))
                    repeat(3) { Caja(Modifier.size(12.dp, 10.dp), linea, 2.dp) }
                }
                Celdas(4, 4, h - 15.dp)
            }
            Dibujo.dos -> Row(Modifier.fillMaxSize(), horizontalArrangement = Arrangement.spacedBy(4.dp)) {
                Foto(Modifier.weight(1f).fillMaxHeight()); Caja(Modifier.weight(1f).fillMaxHeight(), linea)
            }
            Dibujo.iconos -> Row(Modifier.fillMaxSize(), horizontalArrangement = Arrangement.spacedBy(4.dp), verticalAlignment = Alignment.CenterVertically) {
                repeat(4) {
                    Column(Modifier.weight(1f), horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(4.dp)) {
                        Box(Modifier.size(14.dp).border(1.5.dp, linea, RoundedCornerShape(4.dp)))
                        Box(Modifier.fillMaxWidth().height(4.dp).clip(CircleShape).background(suave))
                    }
                }
            }
            is Dibujo.grilla -> Celdas(d.columnas, d.total, h)
            Dibujo.look -> Row(Modifier.fillMaxSize(), horizontalArrangement = Arrangement.spacedBy(4.dp)) {
                repeat(3) { Foto(Modifier.weight(1f).fillMaxHeight()) }
                Column(Modifier.weight(1f).fillMaxHeight().border(1.5.dp, linea, RoundedCornerShape(3.dp)).padding(3.dp), verticalArrangement = Arrangement.spacedBy(3.dp, Alignment.Bottom)) {
                    Box(Modifier.fillMaxWidth().height(4.dp).clip(CircleShape).background(linea))
                    Box(Modifier.fillMaxWidth().height(7.dp).clip(RoundedCornerShape(2.dp)).background(linea))
                }
            }
            Dibujo.lista -> Column(Modifier.fillMaxSize(), verticalArrangement = Arrangement.spacedBy(5.dp)) {
                repeat(3) { i ->
                    Row(Modifier.weight(1f).fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(5.dp), verticalAlignment = Alignment.CenterVertically) {
                        Column(verticalArrangement = Arrangement.spacedBy(3.dp)) { Linea(if (i == 1) 0.35f else 0.45f, 4.dp); Suave(0.6f) }
                        Spacer(Modifier.weight(1f))
                        Foto(Modifier.size(h / 3.6f))
                    }
                }
            }
            Dibujo.mapa -> Row(Modifier.fillMaxSize(), horizontalArrangement = Arrangement.spacedBy(5.dp)) {
                Column(Modifier.weight(1f).fillMaxHeight(), verticalArrangement = Arrangement.spacedBy(4.dp, Alignment.CenterVertically)) {
                    Linea(0.38f, 6.dp); Suave(0.4f); Suave(0.3f); Suave(0.35f)
                }
                Box(Modifier.weight(1f).fillMaxHeight().clip(RoundedCornerShape(3.dp)).background(hex(0xE6E1D6)), contentAlignment = Alignment.Center) {
                    Canvas(Modifier.fillMaxSize()) {
                        val a = 2.dp.toPx(); val paso = (adentro * 0.13f).toPx()
                        for (i in 1 until 4) drawRect(Color.White, Offset(i * paso, 0f), Size(a, size.height))
                        for (i in 1 until 3) drawRect(Color.White, Offset(0f, i * size.height / 3), Size(size.width, a))
                    }
                    Box(Modifier.size(9.dp).clip(CircleShape).background(ambar))
                }
            }
            Dibujo.acordeon -> Column(Modifier.fillMaxSize(), verticalArrangement = Arrangement.spacedBy(4.dp)) {
                repeat(3) { i ->
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        Linea(if (i == 1) 0.45f else 0.55f, 4.dp); Spacer(Modifier.weight(1f)); Texto(if (i == 0) "−" else "+", Estilo(8f, Peso.regular), linea)
                    }
                    if (i == 0) Suave(0.8f)
                    Box(Modifier.fillMaxWidth().height(1.dp).background(borde))
                }
            }
            Dibujo.tarjetas -> Row(Modifier.fillMaxSize(), horizontalArrangement = Arrangement.spacedBy(4.dp)) {
                repeat(3) {
                    Column(Modifier.weight(1f).fillMaxHeight().border(1.dp, borde, RoundedCornerShape(3.dp)).padding(4.dp), verticalArrangement = Arrangement.spacedBy(3.dp)) {
                        Box(Modifier.size(adentro * 0.12f, 4.dp).clip(CircleShape).background(ambar))
                        Box(Modifier.fillMaxWidth().height(4.dp).clip(CircleShape).background(linea))
                        Box(Modifier.fillMaxWidth().height(3.dp).clip(CircleShape).background(suave))
                        Box(Modifier.size(adentro * 0.15f, 3.dp).clip(CircleShape).background(suave))
                    }
                }
            }
            Dibujo.logos -> Column(Modifier.fillMaxSize(), verticalArrangement = Arrangement.spacedBy(9.dp, Alignment.CenterVertically), horizontalAlignment = Alignment.CenterHorizontally) {
                Suave(0.4f)
                Row(horizontalArrangement = Arrangement.spacedBy(6.dp)) {
                    listOf(0.22f, 0.17f, 0.25f, 0.19f).forEach { f -> Caja(Modifier.size(adentro * f * 0.9f, 9.dp), gris, 2.dp) }
                }
            }
            Dibujo.input -> Column(Modifier.fillMaxSize(), verticalArrangement = Arrangement.spacedBy(5.dp, Alignment.CenterVertically), horizontalAlignment = Alignment.CenterHorizontally) {
                Linea(0.55f, 6.dp); Suave(0.4f)
                Row(horizontalArrangement = Arrangement.spacedBy(3.dp)) {
                    Box(Modifier.size(adentro * 0.42f, 11.dp).border(1.dp, gris, RoundedCornerShape(2.dp)))
                    Caja(Modifier.size(adentro * 0.2f, 11.dp), linea, 2.dp)
                }
            }
            Dibujo.circulos -> Row(Modifier.fillMaxSize(), verticalAlignment = Alignment.CenterVertically) {
                repeat(4) {
                    Column(Modifier.weight(1f), horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(4.dp)) {
                        Box(Modifier.size(adentro * 0.2f).clip(CircleShape).background(foto))
                        Box(Modifier.size(adentro * 0.15f, 4.dp).clip(CircleShape).background(suave))
                    }
                }
            }
            Dibujo.video -> VideoCaja(Modifier.fillMaxSize())
            Dibujo.miniaturas -> Row(Modifier.fillMaxSize(), horizontalArrangement = Arrangement.spacedBy(4.dp)) {
                Column(Modifier.width(adentro * 0.12f).fillMaxHeight(), verticalArrangement = Arrangement.spacedBy(3.dp)) {
                    repeat(3) { Foto(Modifier.weight(1f).fillMaxWidth()) }
                }
                Foto(Modifier.weight(1f).fillMaxHeight())
                Textos(Modifier.weight(1f).fillMaxHeight(), 0.3f)
            }
            Dibujo.panel -> Row(Modifier.fillMaxSize(), horizontalArrangement = Arrangement.spacedBy(4.dp)) {
                Foto(Modifier.weight(1f).fillMaxHeight().alpha(0.5f))
                Column(Modifier.width(adentro * 0.38f).fillMaxHeight().border(1.dp, borde, RoundedCornerShape(3.dp)).padding(4.dp), verticalArrangement = Arrangement.spacedBy(4.dp)) {
                    Linea(0.25f, 5.dp)
                    repeat(2) { Row(horizontalArrangement = Arrangement.spacedBy(3.dp), verticalAlignment = Alignment.CenterVertically) { Foto(Modifier.size(12.dp)); Suave(0.15f) } }
                    Spacer(Modifier.weight(1f))
                    Box(Modifier.fillMaxWidth().height(8.dp).clip(RoundedCornerShape(2.dp)).background(linea))
                }
            }
        }
    }
}

