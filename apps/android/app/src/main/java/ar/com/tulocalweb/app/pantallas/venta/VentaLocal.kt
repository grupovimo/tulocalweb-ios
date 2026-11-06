package ar.com.tulocalweb.app.pantallas.venta

import android.media.AudioManager
import android.media.ToneGenerator
import androidx.compose.animation.AnimatedContent
import androidx.compose.animation.core.RepeatMode
import androidx.compose.animation.core.animateFloat
import androidx.compose.animation.core.infiniteRepeatable
import androidx.compose.animation.core.rememberInfiniteTransition
import androidx.compose.animation.core.tween
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.slideInHorizontally
import androidx.compose.animation.togetherWith
import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateListOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.clipToBounds
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Path
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.graphics.StrokeJoin
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.layout.layout
import androidx.compose.ui.platform.LocalFocusManager
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.app.LocalNavegacion
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.PaginaProductos
import ar.com.tulocalweb.app.nucleo.ProductoFila
import ar.com.tulocalweb.app.nucleo.opcional
import ar.com.tulocalweb.app.tema.*
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

// Clon de Pantallas/Venta/VentaLocal.swift.

/** El ticket de una venta en el local. */
data class LineaTicket(val producto: ProductoFila, val variante: ProductoFila.Variante?, val cantidad: Int) {
    val id: String get() = producto.id + (variante?.let { ":" + it.id } ?: "")
    val precio: Int get() = variante?.precio ?: producto.precio
    val nombre: String get() = variante?.let { "${producto.nombre} · ${it.texto}" } ?: producto.nombre
}

/** El pitido de leer un código (AudioServicesPlaySystemSound(1057) en iOS). */
private val pitido by lazy { runCatching { ToneGenerator(AudioManager.STREAM_NOTIFICATION, 60) }.getOrNull() }
private fun sonar() { pitido?.startTone(ToneGenerator.TONE_PROP_BEEP, 80) }

/** 4a · Venta rápida: cámara arriba (300) con visor y línea de escaneo, la hoja del ticket que se superpone 20, buscador,
 *  filas con stepper (la recién escaneada resaltada) y «Cobrar» con el total. Sigue a 4b (medios), 4c (QR),
 *  4d (transferencia) y 4e (confirmada). */
@Composable
fun VentaLocal() {
    val estado = LocalEstado.current
    val nav = LocalNavegacion.current
    val lineas = remember { mutableStateListOf<LineaTicket>() }
    var paso by remember { mutableStateOf("ticket") }
    var hecha by remember { mutableStateOf<VentaHecha?>(null) }
    val total = lineas.sumOf { it.precio * it.cantidad }
    val cerrar = { nav.venta = false }

    Box(Modifier.fillMaxSize().background(C.fondo)) {
        AnimatedContent(Triple(hecha != null, paso, hecha), transitionSpec = {
            if (targetState.second == "cobrar" && initialState.second == "ticket") slideInHorizontally { it } togetherWith fadeOut()
            else fadeIn(tween(250)) togetherWith fadeOut(tween(250))
        }, label = "venta") { (conHecha, p, h) ->
            when {
                conHecha && h != null -> VentaConfirmada(h, { lineas.clear(); hecha = null; paso = "ticket" }) { cerrar() }
                p == "cobrar" -> Cobrar(lineas.toList(), total, { paso = "ticket" }) { v ->
                    Haptico.exito()
                    hecha = v
                    estado.version += 1
                    estado.tarea { estado.refrescarResumen() }
                }
                else -> Ticket(lineas, total, cerrar) { paso = "cobrar" }
            }
        }
    }
}

@Composable
private fun Ticket(lineas: androidx.compose.runtime.snapshots.SnapshotStateList<LineaTicket>, total: Int, cerrar: () -> Unit, cobrar: () -> Unit) {
    val estado = LocalEstado.current
    val alcance = rememberCoroutineScope()
    val foco = LocalFocusManager.current
    var recien by remember { mutableStateOf<String?>(null) }
    var busqueda by remember { mutableStateOf("") }
    var resultados by remember { mutableStateOf<List<ProductoFila>>(emptyList()) }
    var linterna by remember { mutableStateOf(false) }
    var elegirVariante by remember { mutableStateOf<ProductoFila?>(null) }
    val unidades = lineas.sumOf { it.cantidad }

    fun sumar(p: ProductoFila, v: ProductoFila.Variante?) {
        Haptico.seleccion()
        val nueva = LineaTicket(p, v, 1)
        val i = lineas.indexOfFirst { it.id == nueva.id }
        if (i >= 0) lineas[i] = lineas[i].copy(cantidad = lineas[i].cantidad + 1) else lineas.add(nueva)
        recien = nueva.id
    }

    fun agregar(p: ProductoFila) {
        busqueda = ""
        foco.clearFocus()
        if (p.variantes.any { it.activa }) elegirVariante = p else sumar(p, null)
    }

    fun cambiar(l: LineaTicket, d: Int) {
        val i = lineas.indexOfFirst { it.id == l.id }
        if (i < 0) return
        Haptico.seleccion()
        val n = lineas[i].cantidad + d
        if (n <= 0) lineas.removeAt(i) else lineas[i] = lineas[i].copy(cantidad = n)
    }

    suspend fun escaneado(codigo: String) {
        val l = estado.api.opcional<PaginaProductos>("/v1/panel/listas/productos", listOf("q" to codigo, "cuantos" to "5")) ?: return
        // El SKU exacto primero (del producto o de una variante).
        val exacto = l.filas.firstOrNull { it.sku == codigo }
        val conVariante = l.filas.firstOrNull { p -> p.variantes.any { it.sku == codigo } }
        when {
            exacto != null -> { sonar(); sumar(exacto, null) }
            conVariante != null -> { sonar(); sumar(conVariante, conVariante.variantes.first { it.sku == codigo }) }
            l.filas.size == 1 -> { sonar(); agregar(l.filas[0]) }
            else -> { Haptico.error(); estado.avisar("No encontramos el código $codigo", ok = false) }
        }
    }

    LaunchedEffect(busqueda) {
        val q = busqueda.trim()
        if (q.isEmpty()) { resultados = emptyList(); return@LaunchedEffect }
        delay(250)
        estado.api.opcional<PaginaProductos>("/v1/panel/listas/productos", listOf("q" to q, "cuantos" to "20", "seg" to "active,out"))?.let { resultados = it.filas }
    }

    Column(Modifier.fillMaxSize().imePadding()) {
        // La cámara llega hasta arriba de todo, pero los botones quedan debajo de la barra de estado.
        Box(Modifier.fillMaxWidth().background(Color.Black).clipToBounds()) {
            Escaner(linterna, Modifier.matchParentSize()) { c -> alcance.launch { escaneado(c) } }
            Box(Modifier.fillMaxWidth().statusBarsPadding().height(250.dp)) {
                Visor(Modifier.align(Alignment.Center).size(260.dp, 116.dp))
                Row(Modifier.fillMaxWidth().padding(horizontal = 16.dp).padding(top = 6.dp), verticalAlignment = Alignment.CenterVertically) {
                    BotonCamara("xmark", "Cerrar", cerrar)
                    Texto("Nueva venta", Estilo.headline, hex(0xFBF7F2), Modifier.weight(1f), alinear = TextAlign.Center)
                    BotonCamara(if (linterna) "bolt.fill" else "bolt", "Linterna") { linterna = !linterna }
                }
                Texto("Apuntá al código de barras", Estilo.texto14, hex(0xFBF7F2, 0.85f), Modifier.align(Alignment.BottomCenter).padding(bottom = 36.dp))
            }
        }
        // .padding(.top, -20) de iOS: la hoja sube 20 sobre la cámara y además crece esos 20 (offset sólo la corría).
        Column(Modifier.weight(1f).layout { medible, r ->
            val extra = 20.dp.roundToPx()
            val p = medible.measure(r.copy(minHeight = r.minHeight + extra, maxHeight = if (r.hasBoundedHeight) r.maxHeight + extra else r.maxHeight))
            layout(p.width, p.height - extra) { p.place(0, -extra) }
        }.clip(RoundedCornerShape(topStart = 22.dp, topEnd = 22.dp)).background(C.fondo)
            .padding(horizontal = 20.dp).padding(top = 14.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
            Box(Modifier.clip(RoundedCornerShape(12.dp)).background(C.superficie)) { Buscador("Buscá por nombre o código", busqueda, { busqueda = it }) }
            if (resultados.isNotEmpty() && busqueda.isNotEmpty()) {
                Column(Modifier.heightIn(max = 260.dp).verticalScroll(rememberScrollState())) {
                    Column(Modifier.fillMaxWidth().tarjeta()) {
                        resultados.forEachIndexed { i, p ->
                            if (i > 0) Separador(62.dp)
                            Row(Modifier.fillMaxWidth().clickable { agregar(p) }.padding(horizontal = 12.dp, vertical = 9.dp), horizontalArrangement = Arrangement.spacedBy(12.dp),
                                verticalAlignment = Alignment.CenterVertically) {
                                Miniatura(p.foto, 38.dp, radio = 10.dp)
                                Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(1.dp)) {
                                    Texto(p.nombre, Estilo.fila, C.texto, lineas = 1)
                                    Texto("${Formato.pesos(p.precio)} · ${if (p.stock == 0) "sin stock" else "${p.stock}${Formato.fino}u."}", Estilo.footnote,
                                        if (p.stock == 0) C.errorTinta else C.texto2, numeros = true)
                                }
                                Icono("plus.circle.fill", 22f, C.primarioTexto)
                            }
                        }
                    }
                }
            }
            Row(Modifier.padding(horizontal = 4.dp).padding(top = 2.dp), verticalAlignment = Alignment.CenterVertically) {
                Texto("Ticket · ${Formato.cantidad(unidades, "producto", "productos")}", Estilo.label, C.texto2, Modifier.weight(1f))
                if (lineas.isNotEmpty()) Texto("Vaciar", Estilo.label, C.primarioTexto, Modifier.clickable { lineas.clear() })
            }
            Column(Modifier.weight(1f).verticalScroll(rememberScrollState())) {
                if (lineas.isEmpty()) {
                    Texto(Formato.sinViudas("Escaneá un código o buscá un producto para empezar el ticket."), Estilo.subhead, C.texto2,
                        Modifier.fillMaxWidth().padding(top = 24.dp).padding(horizontal = 20.dp), alinear = TextAlign.Center)
                } else {
                    Column(Modifier.fillMaxWidth().tarjeta()) {
                        lineas.toList().forEachIndexed { i, l ->
                            if (i > 0) Separador(14.dp)
                            Row(Modifier.fillMaxWidth().background(if (l.id == recien) tema(0xFFFBF6, 0x3A2216).actual else Color.Transparent)
                                .padding(horizontal = 14.dp, vertical = 11.dp), horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
                                Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                                    Texto(l.nombre, Estilo.fila, C.texto, lineas = 2)
                                    Texto(if (l.id == recien) "${Formato.pesos(l.precio)} · recién escaneado"
                                        else if (l.cantidad > 1) "${Formato.pesos(l.precio)} c/u · ${Formato.pesos(l.precio * l.cantidad)}" else Formato.pesos(l.precio),
                                        Estilo.footnote, C.texto2, numeros = true)
                                }
                                Row(Modifier.clip(Capsula).background(C.pista), verticalAlignment = Alignment.CenterVertically) {
                                    Box(Modifier.size(34.dp).clickable { cambiar(l, -1) }.semantics { contentDescription = "Uno menos" }, contentAlignment = Alignment.Center) {
                                        Texto("−", Estilo(20f, Peso.regular), C.primarioTexto)
                                    }
                                    Texto("${l.cantidad}", Estilo(16f, Peso.bold), C.primarioTexto, Modifier.widthIn(min = 18.dp), alinear = TextAlign.Center, numeros = true)
                                    Box(Modifier.size(34.dp).clickable { cambiar(l, 1) }.semantics { contentDescription = "Uno más" }, contentAlignment = Alignment.Center) {
                                        Texto("+", Estilo(20f, Peso.regular), C.primarioTexto)
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
        Column(Modifier.fillMaxWidth().background(tema(0xFBF7F2, 0x22140D, alfa = 0.94f).actual)) {
            Box(Modifier.fillMaxWidth().height(0.5.dp).background(C.linea2))
            BotonPrincipal("", cobrar, Modifier.navigationBarsPadding().padding(horizontal = 20.dp).padding(top = 12.dp, bottom = 8.dp), alto = 56.dp, habilitado = lineas.isNotEmpty()) {
                Row(Modifier.weight(1f).padding(horizontal = 0.dp), verticalAlignment = Alignment.CenterVertically) {
                    Texto("Cobrar", Estilo.headline, Color.White, Modifier.weight(1f))
                    Texto(Formato.pesos(total), Estilo(20f, Peso.bold), Color.White, numeros = true)
                }
            }
        }
    }
    val p = elegirVariante
    Confirmacion(p != null, "¿Cuál?", opciones = p?.variantes?.filter { it.activa }?.map { v ->
        OpcionConfirmacion("${v.texto} · ${v.stock}${Formato.fino}u.") { sumar(p, v) }
    } ?: emptyList(), alCerrar = { elegirVariante = null }, volver = "Cancelar", verTitulo = false)
}

/** Botón redondo sobre la cámara (34, blanco al 16 %). */
@Composable
fun BotonCamara(simbolo: String, etiqueta: String, accion: () -> Unit) {
    Box(Modifier.size(44.dp).clickable(onClick = accion).semantics { contentDescription = etiqueta }, contentAlignment = Alignment.Center) {
        Box(Modifier.size(34.dp).clip(CircleShape).background(hex(0xFBF7F2, 0.16f)), contentAlignment = Alignment.Center) { Icono(simbolo, 14f, hex(0xFBF7F2)) }
    }
}

/** El visor: esquinas blancas de 4 y la línea de escaneo `#E07A48` con brillo. */
@Composable
fun Visor(modifier: Modifier = Modifier) {
    val y = if (animacionesActivas()) rememberInfiniteTransition(label = "visor").animateFloat(0.2f, 0.8f, infiniteRepeatable(tween(1600), RepeatMode.Reverse), label = "y").value else 0.5f
    val linea = hex(0xE07A48)
    Canvas(modifier) {
        val l = 26.dp.toPx(); val w = size.width; val h = size.height
        val p = Path().apply {
            moveTo(0f, l); lineTo(0f, 0f); lineTo(l, 0f)
            moveTo(w - l, 0f); lineTo(w, 0f); lineTo(w, l)
            moveTo(w, h - l); lineTo(w, h); lineTo(w - l, h)
            moveTo(l, h); lineTo(0f, h); lineTo(0f, h - l)
        }
        drawPath(p, Color.White, style = Stroke(4.dp.toPx(), cap = StrokeCap.Round, join = StrokeJoin.Round))
        val m = 12.dp.toPx()
        drawLine(linea.copy(alpha = 0.35f), Offset(m, h * y), Offset(w - m, h * y), 8.dp.toPx(), cap = StrokeCap.Round)
        drawLine(linea, Offset(m, h * y), Offset(w - m, h * y), 2.dp.toPx(), cap = StrokeCap.Round)
    }
}
