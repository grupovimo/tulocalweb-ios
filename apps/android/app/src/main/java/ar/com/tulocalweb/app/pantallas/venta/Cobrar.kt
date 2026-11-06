package ar.com.tulocalweb.app.pantallas.venta

import android.app.Activity
import android.content.Context
import android.content.ContextWrapper
import android.view.WindowManager
import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.core.RepeatMode
import androidx.compose.animation.core.animateFloat
import androidx.compose.animation.core.infiniteRepeatable
import androidx.compose.animation.core.rememberInfiniteTransition
import androidx.compose.animation.core.tween
import androidx.compose.animation.slideInVertically
import androidx.compose.animation.slideOutVertically
import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
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
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.text.selection.SelectionContainer
import androidx.compose.foundation.verticalScroll
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.drawBehind
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.FilterQuality
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.selected
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.R
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.nucleo.ErrorApi
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.Tincho
import ar.com.tulocalweb.app.nucleo.compartirTexto
import ar.com.tulocalweb.app.nucleo.copiar
import ar.com.tulocalweb.app.nucleo.intentar
import ar.com.tulocalweb.app.nucleo.opcional
import ar.com.tulocalweb.app.pantallas.mas.LinksPago
import ar.com.tulocalweb.app.pantallas.mas.NegocioCobro
import ar.com.tulocalweb.app.pantallas.mas.QRTienda
import ar.com.tulocalweb.app.pantallas.pedidos.PedidosApi
import ar.com.tulocalweb.app.tema.*
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import kotlinx.serialization.Serializable
import java.time.Instant

// Clon de Pantallas/Venta/Cobrar.swift.

data class VentaHecha(val numero: Int, val total: Int, val medio: String, val vuelto: Int? = null, val cuando: Instant = Instant.now())

@Serializable data class NegocioCobroDatos(val alias: String? = null, val cbu: String? = null, val titular: String? = null, val nombre: String? = null)
data class QRCobro(val numero: Int, val referencia: String, val datos: String)

/** 4b · Cobrar: el total (48/700), la grilla 2 × 2 de medios (el elegido con borde de 2 `primary-text`) y lo de cada
 *  medio: efectivo con «¿Con cuánto paga?», atajos y el vuelto (4b), transferencia con alias y CVU (4d), QR de Mercado
 *  Pago que pasa solo a confirmada al acreditarse (4c). La venta se crea con POST /v1/panel/pedidos (canal pos, retiro). */
@Composable
fun Cobrar(lineas: List<LineaTicket>, total: Int, volver: () -> Unit, alCobrar: (VentaHecha) -> Unit) {
    val estado = LocalEstado.current
    val ctx = LocalContext.current
    val alcance = rememberCoroutineScope()
    var medio by remember { mutableStateOf("efectivo") }
    var paga by remember { mutableIntStateOf(total) }
    var cobrando by remember { mutableStateOf(false) }
    var negocio by remember { mutableStateOf<NegocioCobroDatos?>(null) }
    var qr by remember { mutableStateOf<QRCobro?>(null) }
    var editarNegocio by remember { mutableStateOf(false) }
    var armarLink by remember { mutableStateOf(false) }

    @Serializable data class Conf(val negocio: NegocioCobroDatos? = null)
    suspend fun traerNegocio() { estado.api.opcional<Conf>("/v1/panel/configuracion")?.let { negocio = it.negocio } }
    LaunchedEffect(Unit) { paga = total; traerNegocio() }

    val textoBoton = when (medio) { "transf" -> "Ya me transfirió"; "mp" -> "Mostrar el QR"; "link" -> "Armar el link de pago"; else -> "Confirmar cobro" }

    @Serializable data class Item(val productoId: String, val varianteId: String?, val cantidad: Int)
    @Serializable data class Cuerpo(val items: List<Item>, val pago: String, val canal: String = "pos", val envio: String = "retiro", val pagado: Boolean)
    @Serializable data class Creado(val numero: Int, val total: Int)
    @Serializable data class Q(val referencia: String, val qr: String)

    suspend fun confirmar() {
        if (medio == "link") { armarLink = true; return }
        cobrando = true
        try {
            val items = lineas.map { Item(it.producto.id, it.variante?.id, it.cantidad) }
            val c: Creado = estado.api.enviar("POST", "/v1/panel/pedidos", Cuerpo(items, medio, pagado = medio != "mp"))
            if (medio == "mp") {
                try {
                    val q: Q = estado.api.enviar("POST", "/v1/panel/pedidos/${c.numero}/qr", emptyMap<String, String>())
                    qr = QRCobro(c.numero, q.referencia, q.qr)
                } catch (e: ErrorApi) {
                    // Sin Mercado Pago conectado: la venta no queda colgada.
                    intentar { PedidosApi.cambiarEstado(estado.api, c.numero, "Cancelado") }
                    estado.avisar(e.detalle, ok = false)
                }
                return
            }
            alCobrar(VentaHecha(c.numero, c.total, if (medio == "efectivo") "Efectivo" else "Transferencia", if (medio == "efectivo") maxOf(0, paga - c.total) else null))
        } catch (e: ErrorApi) { Haptico.error(); estado.avisar(e.detalle, ok = false) } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
        } finally { cobrando = false }
    }

    @Composable
    fun Tile(id: String, texto: String, simbolo: String, modifier: Modifier) {
        val activo = medio == id
        val forma = RoundedCornerShape(16.dp)
        Row(modifier.height(64.dp).clip(forma).background(C.superficie).border(if (activo) 2.dp else 1.dp, if (activo) C.primarioTexto else tema(0x3A1D12, 0xFBF7F2, alfa = 0.12f).actual, forma)
            .clickable { medio = id; Haptico.seleccion() }.semantics { selected = activo }.padding(horizontal = 14.dp),
            horizontalArrangement = Arrangement.spacedBy(10.dp), verticalAlignment = Alignment.CenterVertically) {
            Box(Modifier.size(30.dp).clip(RoundedCornerShape(9.dp)).background(if (activo) C.crema else C.pista), contentAlignment = Alignment.Center) { Icono(simbolo, 14f, C.primarioTexto) }
            Texto(texto, Estilo.fila, C.texto, lineas = 2)
        }
    }

    @Composable
    fun Dato(titulo: String, valor: String, grande: Boolean, copiarTexto: String? = null) {
        Row(Modifier.fillMaxWidth().padding(horizontal = 16.dp, vertical = 12.dp), verticalAlignment = Alignment.CenterVertically) {
            Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                Texto(titulo, Estilo.footnoteFuerte, C.texto2)
                SelectionContainer { Texto(valor, if (grande) Estilo(26f, Peso.bold, -0.02f) else Estilo.body16, C.texto, numeros = true) }
            }
            BotonTonal("Copiar", { copiar(ctx, copiarTexto ?: valor); Haptico.exito(); estado.avisar("Copiamos el ${titulo.lowercase()}") })
        }
    }

    Box(Modifier.fillMaxSize().background(C.fondo)) {
        Column(Modifier.fillMaxSize().statusBarsPadding().imePadding()) {
            Box(Modifier.fillMaxWidth().height(44.dp).padding(start = 10.dp, end = 16.dp)) {
                Row(Modifier.align(Alignment.CenterStart).clickable(onClick = volver), horizontalArrangement = Arrangement.spacedBy(4.dp), verticalAlignment = Alignment.CenterVertically) {
                    Icono("chevron.left", 17f, C.primarioTexto); Texto("Ticket", Estilo.body, C.primarioTexto)
                }
                Texto("Cobrar", Estilo.headline, C.texto, Modifier.align(Alignment.Center))
            }
            Column(Modifier.weight(1f).verticalScroll(rememberScrollState()).padding(horizontal = 20.dp).padding(top = 12.dp, bottom = 20.dp),
                horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(16.dp)) {
                Column(horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(2.dp)) {
                    Texto("Total a cobrar", Estilo.subhead, C.texto2)
                    Texto(Formato.pesos(total), Estilo(48f, Peso.bold, -0.035f), C.texto, numeros = true)
                }
                Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
                    Row(horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                        Tile("efectivo", "Efectivo", "banknote", Modifier.weight(1f)); Tile("transf", "Transferencia", "arrow.left.arrow.right", Modifier.weight(1f))
                    }
                    Row(horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                        Tile("mp", "QR Mercado Pago", "qrcode", Modifier.weight(1f)); Tile("link", "Link de pago", "link", Modifier.weight(1f))
                    }
                }
                when (medio) {
                    "efectivo" -> {
                        // MARK: Efectivo (4b)
                        val redondos = billetes(total)
                        val vuelto = maxOf(0, paga - total)
                        Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
                            Column(Modifier.fillMaxWidth().tarjeta().padding(16.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
                                Texto("¿Con cuánto paga?", Estilo.subheadFuerte, C.texto2)
                                val grande = Estilo(34f, Peso.bold, -0.02f)
                                val acento = C.acento
                                Row(Modifier.fillMaxWidth().padding(bottom = 6.dp).drawBehind { drawLine(acento, Offset(0f, size.height + 6.dp.toPx()), Offset(size.width, size.height + 6.dp.toPx()), 2.dp.toPx()) },
                                    verticalAlignment = Alignment.CenterVertically) {
                                    Texto("$${Formato.fino}", grande, C.texto, numeros = true)
                                    BasicTextField(if (paga > 0) Formato.numero(paga) else "", { paga = it.filter { c -> c.isDigit() }.toIntOrNull() ?: 0 }, Modifier.weight(1f),
                                        textStyle = grande.ts(true).copy(color = C.texto), singleLine = true, keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number),
                                        cursorBrush = SolidColor(C.acento))
                                }
                                Row(Modifier.fillMaxWidth().horizontalScroll(rememberScrollState()), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                                    ChipFiltro("Justo", activo = paga == total) { paga = total }
                                    redondos.forEach { b -> ChipFiltro(Formato.pesos(b), activo = paga == b) { paga = b } }
                                }
                            }
                            Row(Modifier.fillMaxWidth().clip(RoundedCornerShape(20.dp)).background(C.okFondo).padding(horizontal = 18.dp, vertical = 16.dp), verticalAlignment = Alignment.CenterVertically) {
                                Texto("Vuelto", Estilo.headline, C.okTinta, Modifier.weight(1f))
                                Texto(Formato.pesos(vuelto), Estilo(32f, Peso.bold, -0.02f), C.okTinta, numeros = true)
                            }
                            if (paga in 1 until total) MensajeError("Faltan ${Formato.pesos(total - paga)} para el total.")
                        }
                    }
                    "transf" -> {
                        // MARK: Transferencia (4d)
                        val n = negocio
                        if (n != null && (!n.alias.isNullOrEmpty() || !n.cbu.isNullOrEmpty())) {
                            Column(Modifier.fillMaxWidth().tarjeta()) {
                                if (!n.alias.isNullOrEmpty()) Dato("Alias", n.alias, true)
                                if (!n.cbu.isNullOrEmpty()) { Separador(); Dato("CVU", agrupar(n.cbu), false, n.cbu) }
                                if (!n.titular.isNullOrEmpty()) {
                                    Separador()
                                    Column(Modifier.fillMaxWidth().padding(horizontal = 16.dp, vertical = 12.dp), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                                        Texto("Titular", Estilo.footnoteFuerte, C.texto2)
                                        Texto(n.titular, Estilo.body16, C.texto)
                                    }
                                }
                            }
                            BotonSecundario("Compartir datos", {
                                compartirTexto(ctx, "Para pagar ${Formato.pesos(total)} a ${estado.tienda?.nombre ?: ""}:\nAlias: ${n.alias ?: ""}\nCVU: ${n.cbu ?: ""}\nTitular: ${n.titular ?: ""}")
                            }, simbolo = "square.and.arrow.up")
                        } else {
                            EstadoTincho(Tincho.portapapeles, "Falta tu alias", "Cargá el alias o el CVU una vez y te lo mostramos acá para cada cobro.", 120.dp,
                                accion = "Cargar el alias", alTocar = { editarNegocio = true })
                        }
                    }
                    "mp" -> Texto(Formato.sinViudas("Mostrale el QR: lo escanea con la app de Mercado Pago o la de su banco, y se confirma solo."), Estilo.subhead, C.texto2,
                        Modifier.padding(horizontal = 10.dp), alinear = TextAlign.Center)
                    else -> Texto(Formato.sinViudas("Armás el link y lo mandás por WhatsApp: se paga con Mercado Pago."), Estilo.subhead, C.texto2,
                        Modifier.padding(horizontal = 10.dp), alinear = TextAlign.Center)
                }
            }
            BotonPrincipal(textoBoton, { alcance.launch { confirmar() } }, Modifier.navigationBarsPadding().padding(horizontal = 20.dp).padding(bottom = 12.dp), alto = 56.dp,
                habilitado = !(cobrando || (medio == "efectivo" && paga in 1 until total)), cargando = cobrando)
        }
        AnimatedVisibility(qr != null, enter = slideInVertically { it }, exit = slideOutVertically { it }) {
            qr?.let { q ->
                CobroQR(q, total, estado.tienda?.nombre ?: "") { ok ->
                    qr = null
                    if (ok) alCobrar(VentaHecha(q.numero, total, "QR Mercado Pago"))
                }
            }
        }
    }
    Hoja(editarNegocio, { editarNegocio = false; alcance.launch { traerNegocio() } }) { NegocioCobro() }
    Hoja(armarLink, { armarLink = false }) { LinksPago(montoInicial = total) }
}

/** Los billetes redondos que siguen al total: 18.900 → 20.000, 25.000, 30.000. */
fun billetes(total: Int): List<Int> {
    val r = mutableListOf<Int>()
    for (paso in listOf(1000, 5000, 10000, 20000)) {
        val v = ((total / paso) + 1) * paso
        if (v > total && v !in r) r.add(v)
    }
    return r.sorted().take(3)
}

fun agrupar(cbu: String): String {
    if (cbu.length != 22) return cbu
    return listOf(cbu.substring(0, 10), cbu.substring(10, 14), cbu.substring(14, 18), cbu.substring(18, 22)).joinToString(" ")
}

private fun Context.actividad(): Activity? = when (this) { is Activity -> this; is ContextWrapper -> baseContext.actividad(); else -> null }

/** 4c · El QR de Mercado Pago, para girar el teléfono: nombre de la tienda, monto 52/700, QR de 290 en tarjeta blanca de
 *  radio 28 con el isotipo al centro y «Esperando el pago…» con el punto `warn` pulsante. Consulta cada 3 s; al
 *  acreditarse pasa solo a confirmada. Brillo al máximo mientras se ve. */
@Composable
fun CobroQR(qr: QRCobro, total: Int, nombreTienda: String, terminar: (Boolean) -> Unit) {
    val estado = LocalEstado.current
    val ctx = LocalContext.current
    val alcance = rememberCoroutineScope()
    val pulso = rememberInfiniteTransition(label = "pulso").animateFloat(1f, 0.35f, infiniteRepeatable(tween(800), RepeatMode.Reverse), label = "p").value

    suspend fun cancelar() {
        intentar { estado.api.hacer("DELETE", "/v1/panel/cobros/qr/${qr.referencia}") }
        intentar { PedidosApi.cambiarEstado(estado.api, qr.numero, "Cancelado") }
        terminar(false)
    }

    DisposableEffect(Unit) {
        val v = ctx.actividad()?.window
        val antes = v?.attributes?.screenBrightness ?: WindowManager.LayoutParams.BRIGHTNESS_OVERRIDE_NONE
        v?.attributes = v?.attributes?.apply { screenBrightness = 1f }
        onDispose { v?.attributes = v?.attributes?.apply { screenBrightness = antes } }
    }

    @Serializable data class E(val estado: String)
    LaunchedEffect(qr.referencia) {
        while (true) {
            delay(3000)
            val e = estado.api.opcional<E>("/v1/panel/cobros/qr/${qr.referencia}") ?: continue
            if (e.estado in listOf("processed", "approved", "paid")) { terminar(true); return@LaunchedEffect }
            if (e.estado in listOf("canceled", "expired", "refunded")) { estado.avisar("El pago no se completó.", ok = false); cancelar(); return@LaunchedEffect }
        }
    }

    Column(Modifier.fillMaxSize().background(C.fondo).statusBarsPadding().navigationBarsPadding().padding(horizontal = 20.dp),
        horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(18.dp)) {
        Box(Modifier.fillMaxWidth().height(44.dp).padding(horizontal = 16.dp), contentAlignment = Alignment.CenterStart) {
            Texto("Cancelar", Estilo.body, C.primarioTexto, Modifier.clickable { alcance.launch { cancelar() } })
        }
        Spacer(Modifier.weight(1f))
        Texto(nombreTienda, Estilo.headline, C.texto2)
        Texto(Formato.pesos(total), Estilo.montoCobro, C.texto, numeros = true)
        QRTienda.imagen(qr.datos)?.let { img ->
            val forma = RoundedCornerShape(28.dp)
            Box(Modifier.sombra(forma, hex(0x2A140C, 0.35f), 25.dp, y = 20.dp).clip(forma).background(Color.White).padding(18.dp)
                .semantics { contentDescription = "Código QR para pagar ${Formato.pesos(total)}" }, contentAlignment = Alignment.Center) {
                Image(img, null, Modifier.size(254.dp), filterQuality = FilterQuality.None)
                Box(Modifier.clip(RoundedCornerShape(10.dp)).background(Color.White).padding(6.dp)) { Imagen(R.drawable.marca, Modifier.size(48.dp)) }
            }
        }
        Texto(Formato.sinViudas("Escaneá con la app de Mercado Pago o tu banco"), Estilo.subhead, C.texto2, alinear = TextAlign.Center)
        Row(horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.CenterVertically) {
            Box(Modifier.size(9.dp).alpha(pulso).clip(CircleShape).background(C.aviso))
            Texto("Esperando el pago…", Estilo.subheadFuerte, C.avisoTinta)
        }
        Spacer(Modifier.weight(1f))
    }
}

/** 4e · ¡Venta cobrada!: Tincho celebrando, el monto 52/700, el detalle, «Emitir factura» / «Compartir ticket» y «Nueva
 *  venta» + «Listo». Háptico de éxito. */
@Composable
fun VentaConfirmada(venta: VentaHecha, nueva: () -> Unit, listo: () -> Unit) {
    val estado = LocalEstado.current
    val ctx = LocalContext.current
    val alcance = rememberCoroutineScope()
    var facturada by remember { mutableStateOf(false) }
    val detalle = buildList {
        add(venta.medio)
        venta.vuelto?.let { if (it > 0) add("vuelto ${Formato.pesos(it)}") }
        add(Formato.cuando(venta.cuando))
    }.joinToString(" · ")
    LaunchedEffect(venta) { Haptico.exito() }

    @Serializable data class Cuerpo(val numeros: List<Int>, val accion: String = "facturar")
    @Serializable data class R(val hechos: Int = 0)

    suspend fun facturar() {
        try {
            val r: R = estado.api.enviar("POST", "/v1/panel/pedidos/lote", Cuerpo(listOf(venta.numero)))
            facturada = r.hechos > 0
            estado.avisar(if (r.hechos > 0) "Pedimos la factura a ARCA" else "Conectá ARCA en Más › Facturación de ARCA para facturar.", ok = r.hechos > 0)
        } catch (e: ErrorApi) { estado.avisar(e.detalle, ok = false) } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {}
    }

    @Composable
    fun Acceso(t: String, s: String, modifier: Modifier, habilitado: Boolean = true, a: () -> Unit) {
        Column(modifier.heightIn(min = 78.dp).tarjeta(radio = 18.dp).clickable(enabled = habilitado, onClick = a).alpha(if (habilitado) 1f else 0.5f),
            horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(6.dp, Alignment.CenterVertically)) {
            Icono(s, 20f, C.primarioTexto)
            Texto(t, Estilo.subheadFuerte, C.texto)
        }
    }

    Column(Modifier.fillMaxSize().background(C.fondo).statusBarsPadding().navigationBarsPadding()) {
        Column(Modifier.weight(1f).verticalScroll(rememberScrollState()).padding(horizontal = 28.dp).padding(top = 30.dp),
            horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(8.dp)) {
            ImagenTincho(Tincho.celebrando, Modifier.size(220.dp))
            Texto("¡Venta cobrada!", Estilo.headline, C.okTinta)
            Texto(Formato.pesos(venta.total), Estilo.montoCobro, C.texto, numeros = true)
            Texto(detalle, Estilo.subhead, C.texto2, alinear = TextAlign.Center, numeros = true)
            Row(Modifier.padding(top = 22.dp), horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                Acceso(if (facturada) "Factura pedida" else "Emitir factura", "doc.text", Modifier.weight(1f), !facturada) { alcance.launch { facturar() } }
                Acceso("Compartir ticket", "square.and.arrow.up", Modifier.weight(1f)) {
                    compartirTexto(ctx, "Ticket #${venta.numero} · ${estado.tienda?.nombre ?: ""}\nTotal ${Formato.pesos(venta.total)} · ${venta.medio}")
                }
            }
        }
        Column(Modifier.padding(horizontal = 20.dp).padding(bottom = 12.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
            BotonPrincipal("Nueva venta", nueva, alto = 56.dp)
            BotonTexto("Listo", listo, alto = 48.dp)
        }
    }
}
