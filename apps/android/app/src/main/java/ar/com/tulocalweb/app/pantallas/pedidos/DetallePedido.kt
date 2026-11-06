package ar.com.tulocalweb.app.pantallas.pedidos

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.IntrinsicSize
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.selection.SelectionContainer
import androidx.compose.foundation.verticalScroll
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.app.BarraDe
import ar.com.tulocalweb.app.app.BotonBarraIcono
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.nucleo.ErrorApi
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.Marcas
import ar.com.tulocalweb.app.nucleo.PedidoDetalle
import ar.com.tulocalweb.app.nucleo.abrirUrl
import ar.com.tulocalweb.app.nucleo.compartirArchivo
import ar.com.tulocalweb.app.nucleo.compartirTexto
import ar.com.tulocalweb.app.nucleo.copiar
import ar.com.tulocalweb.app.nucleo.errorDe
import ar.com.tulocalweb.app.nucleo.intentar
import ar.com.tulocalweb.app.pantallas.mas.EtiquetaEnvio
import ar.com.tulocalweb.app.tema.*
import kotlinx.coroutines.launch
import kotlinx.serialization.Serializable
import java.time.Instant

// Clon de Pantallas/Pedidos/DetallePedido.swift.

/** 3b · Detalle de un pedido: estado y monto, cliente con acciones, productos con el desglose que cierra, pago, entrega
 *  con su línea de tiempo, factura y comprobantes, nota, «Cancelar pedido» y la barra de acción fija con el siguiente
 *  estado (según el tipo de entrega). */
@Composable
fun DetallePedido(numero: Int) {
    val estado = LocalEstado.current
    val ctx = LocalContext.current
    val alcance = rememberCoroutineScope()
    var carga by remember { mutableStateOf<Carga<PedidoDetalle>>(Carga.cargando) }
    var moviendo by remember { mutableStateOf(false) }
    var cancelar by remember { mutableStateOf(false) }
    var facturando by remember { mutableStateOf(false) }

    fun puedeCancelar(p: PedidoDetalle) = p.estado != "Entregado" && p.estado != "Cancelado" && estado.puede("pedidos")
    fun urlPublica(p: PedidoDetalle): String? {
        val c = p.codigo ?: return null
        val slug = estado.tienda?.slug ?: return null
        return "https://$slug.tulocalweb.com.ar/pedido/$c"
    }
    fun textoCompartir(p: PedidoDetalle): String {
        var t = "Pedido #${p.numero} · ${p.comprador.nombre} · ${Formato.pesos(p.total)}\n"
        t += p.items.joinToString("\n") { "${it.cantidad} × ${it.nombre}" }
        urlPublica(p)?.let { t += "\n$it" }
        return t
    }

    suspend fun cargar() {
        try {
            carga = Carga.listo(estado.api.get("/v1/panel/pedidos/$numero"))
        } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
            if (carga.valor == null || e !is ErrorApi) carga = Carga.error(errorDe(e))
        }
    }

    suspend fun mover(p: PedidoDetalle, nuevo: String) {
        moviendo = true
        try {
            PedidosApi.cambiarEstado(estado.api, p.numero, nuevo)
            Haptico.impacto()
            val antes = p.estado
            estado.avisar("#${p.numero} ${Marcas.dicho(nuevo, p.envio)}") {
                estado.tarea { intentar { PedidosApi.cambiarEstado(estado.api, p.numero, antes) }; estado.version += 1 }
            }
            estado.version += 1
            estado.refrescarResumen()
        } catch (e: ErrorApi) {
            Haptico.error(); estado.avisar(e.detalle, ok = false); cargar()
        } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
        } finally { moviendo = false }
    }

    @Serializable data class CuerpoLote(val numeros: List<Int>, val accion: String = "facturar")
    @Serializable data class R(val hechos: Int = 0)

    suspend fun facturar(p: PedidoDetalle) {
        facturando = true
        try {
            val r: R = estado.api.enviar("POST", "/v1/panel/pedidos/lote", CuerpoLote(listOf(p.numero)))
            if (r.hechos > 0) { Haptico.exito(); estado.avisar("Pedimos la factura a ARCA"); cargar() }
            else estado.avisar("No se pudo facturar: revisá ARCA en Más › Facturación de ARCA.", ok = false)
        } catch (e: ErrorApi) { estado.avisar(e.detalle, ok = false) } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
        } finally { facturando = false }
    }

    val pv = carga.valor
    BarraDe("#$numero", acciones = pv?.let { p ->
        {
            BotonBarraIcono("square.and.arrow.up", "Compartir") { compartirTexto(ctx, textoCompartir(p)) }
            MenuBoton(buildList {
                urlPublica(p)?.let { u -> add(OpcionDeMenu("Copiar el link del pedido", "link") { copiar(ctx, u); estado.avisar("Copiamos el link del pedido") }) }
                if (puedeCancelar(p)) add(OpcionDeMenu("Cancelar pedido", "xmark.circle", destructiva = true) { cancelar = true })
            }, "Más acciones") { Box(Modifier.size(44.dp), contentAlignment = Alignment.Center) { Icono("ellipsis", 19f, C.primarioTexto) } }
        }
    })

    Box(Modifier.fillMaxSize().background(C.fondo)) {
        when (val c = carga) {
            Carga.cargando -> Box(Modifier.padding(20.dp)) { EsqueletoFilas(4) }
            is Carga.error -> Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) { EstadoError("el pedido", c.e) { estado.tarea { cargar() } } }
            is Carga.listo -> {
                val p = c.v
                Column(Modifier.fillMaxSize()) {
                    Refrescable({ cargar() }, Modifier.weight(1f)) {
                        Column(Modifier.fillMaxSize().verticalScroll(rememberScrollState()).padding(top = 10.dp).margenPantalla().padding(bottom = 16.dp),
                            verticalArrangement = Arrangement.spacedBy(14.dp)) {
                            Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
                                ChipEstado(p.estado, p.envio, grande = true)
                                Texto(Formato.pesos(p.total), Estilo.montoPantalla, C.texto, numeros = true)
                                Texto("${Formato.cuando(p.creado)} · ${Marcas.canal(p.canal)} · ${Formato.cantidad(p.cantidadProductos, "producto", "productos")}", Estilo.subhead, C.texto2)
                            }
                            Cliente(p)
                            EtiquetaSeccion("Productos", Modifier.padding(horizontal = 4.dp).padding(top = 6.dp))
                            Productos(p)
                            EtiquetaSeccion("Pago", Modifier.padding(horizontal = 4.dp).padding(top = 6.dp))
                            Pago(p)
                            EtiquetaSeccion("Entrega", Modifier.padding(horizontal = 4.dp).padding(top = 6.dp))
                            Entrega(p)
                            EtiquetaSeccion("Factura y comprobantes", Modifier.padding(horizontal = 4.dp).padding(top = 6.dp))
                            Factura(p, urlPublica(p), facturando) { alcance.launch { facturar(p) } }
                            val nota = p.nota
                            if (!nota.isNullOrEmpty()) {
                                EtiquetaSeccion("Notas internas", Modifier.padding(horizontal = 4.dp).padding(top = 6.dp))
                                Box(Modifier.fillMaxWidth().tarjeta().padding(horizontal = 16.dp, vertical = 14.dp)) {
                                    Column(Modifier.fillMaxWidth().clip(RoundedCornerShape(14.dp)).background(C.crema2).padding(horizontal = 12.dp, vertical = 10.dp),
                                        verticalArrangement = Arrangement.spacedBy(3.dp)) {
                                        Texto(nota, Estilo.subhead, C.texto, interlineado = 2f)
                                        Texto("Tincho · de la charla con quien compró", Estilo.chip.con(Peso.regular), C.texto2)
                                    }
                                }
                            }
                            if (puedeCancelar(p)) BotonTexto("Cancelar pedido", { cancelar = true }, alto = 50.dp, color = C.errorTinta)
                        }
                    }
                    val sig = Marcas.siguiente(p.estado, p.envio)
                    if (sig != null) {
                        Column(Modifier.fillMaxWidth().background(tema(0xFBF7F2, 0x22140D, alfa = 0.94f).actual)) {
                            Box(Modifier.fillMaxWidth().height(0.5.dp).background(C.linea2))
                            BotonPrincipal(sig.boton, { alcance.launch { mover(p, sig.estado) } },
                                Modifier.navigationBarsPadding().padding(horizontal = 20.dp).padding(top = 12.dp, bottom = 8.dp), habilitado = !moviendo, cargando = moviendo)
                        }
                    }
                }
            }
        }
    }
    LaunchedEffect(estado.version) { cargar() }
    Hoja(cancelar, { cancelar = false }) { carga.valor?.let { p -> CancelarPedido(p) { cargar() } } }
}

@Composable
private fun Cliente(p: PedidoDetalle) {
    val estado = LocalEstado.current
    val ctx = LocalContext.current
    val c = p.comprador
    Column(Modifier.fillMaxWidth().tarjeta().padding(16.dp), verticalArrangement = Arrangement.spacedBy(14.dp)) {
        Row(horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
            AvatarIniciales(c.nombre, 46.dp, fondo = tema(0xE7EFE9, 0x3FA877, alfaOscuro = 0.18f).actual, tinta = C.okTinta)
            Column(verticalArrangement = Arrangement.spacedBy(2.dp)) {
                Texto(c.nombre, Estilo.headline, C.texto)
                Texto(antiguedad(c), Estilo.footnote, C.texto2)
            }
        }
        if (c.telefono != null || c.email != null) {
            SelectionContainer {
                Column(verticalArrangement = Arrangement.spacedBy(3.dp)) {
                    c.telefono?.let { Texto(it, Estilo.subhead, C.texto, numeros = true) }
                    c.email?.let { Texto(it, Estilo.subhead, C.texto2, numeros = true) }
                }
            }
        }
        Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            val wa = urlWhatsApp(c.telefono, "Hola ${c.nombre.split(" ").firstOrNull() ?: ""}, te escribo por tu pedido #${p.numero}.")
            AccionRapida("phone", "Llamar", Modifier.weight(1f).deshabilitado(c.telefono == null, 0.45f)) {
                c.telefono?.filter { it.isDigit() || it == '+' }?.let { abrirUrl(ctx, "tel:$it") }
            }
            AccionRapida("message", "WhatsApp", Modifier.weight(1f).deshabilitado(urlWhatsApp(c.telefono) == null, 0.45f)) { wa?.let { abrirUrl(ctx, it) } }
            AccionRapida("doc.on.doc", "Copiar mail", Modifier.weight(1f).deshabilitado(c.email == null, 0.45f)) {
                c.email?.let { copiar(ctx, it); Haptico.exito(); estado.avisar("Copiamos el mail") }
            }
        }
    }
}

private fun antiguedad(c: PedidoDetalle.Comprador): String {
    val mes = Formato.mesDe(c.desde)
    if (c.pedidos <= 1) return "Primer pedido"
    val n = c.pedidos
    val ord = mapOf(1 to "1.er", 3 to "3.er")[n] ?: "$n.º"
    val anioDesde = Formato.enArgentina(c.desde).year
    val esteAnio = anioDesde == Formato.enArgentina(Instant.now()).year
    return "$ord pedido · cliente desde $mes${if (esteAnio) "" else " de $anioDesde"}"
}

@Composable
private fun Productos(p: PedidoDetalle) {
    Column(Modifier.fillMaxWidth().tarjeta()) {
        p.items.forEachIndexed { i, it ->
            if (i > 0) Separador(78.dp)
            Row(Modifier.padding(horizontal = 14.dp, vertical = 12.dp), horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
                Miniatura(it.foto, 52.dp, radio = 12.dp)
                Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                    Texto(it.nombre, Estilo.fila, C.texto, lineas = 2)
                    Texto(listOfNotNull(it.variante, "${it.cantidad} × ${Formato.pesos(it.precio)}").joinToString(" · "), Estilo.footnote, C.texto2, numeros = true)
                }
                Texto(Formato.pesos(it.precio * it.cantidad), Estilo.fila, C.texto, numeros = true)
            }
        }
        Box(Modifier.fillMaxWidth().height(1.dp).background(C.linea))
        Column(Modifier.padding(horizontal = 16.dp, vertical = 14.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
            Linea("Subtotal", Formato.pesos(p.subtotal))
            p.descuentos.forEach { d -> Linea(d.concepto, Formato.pesos(-d.monto), C.okTinta) }
            if (p.envio != "retiro") Linea("Envío", if (p.costoEnvio > 0) Formato.pesos(p.costoEnvio) else "Gratis")
            Box(Modifier.padding(vertical = 4.dp).fillMaxWidth().height(1.dp).background(C.linea))
            Row {
                Texto("Total", Estilo.total, C.texto, Modifier.weight(1f))
                Texto(Formato.pesos(p.total), Estilo.total, C.texto, numeros = true)
            }
        }
    }
}

@Composable
private fun Linea(a: String, b: String, color: Color = C.texto) {
    Row {
        Texto(a, Estilo.subhead, C.texto2, Modifier.weight(1f).padding(end = 8.dp), numeros = true)
        Texto(b, Estilo.subhead, color, numeros = true)
    }
}

@Composable
private fun Pago(p: PedidoDetalle) {
    Column(Modifier.fillMaxWidth().tarjeta().padding(horizontal = 16.dp, vertical = 14.dp), verticalArrangement = Arrangement.spacedBy(10.dp)) {
        Row(verticalAlignment = Alignment.CenterVertically) {
            Texto(Marcas.pago(p.pago), Estilo.fila, C.texto, Modifier.weight(1f))
            ChipPago(p.estadoPago)
        }
        if (p.estado == "Sin pagar") {
            Texto(if (p.pago == "transf") Formato.sinViudas("Cuando veas la transferencia en tu cuenta, marcalo como cobrado.")
                else Formato.sinViudas("Todavía no se acreditó. Si te pagó por otro lado, marcalo como cobrado."), Estilo.texto14, C.texto2)
        }
    }
}

@Composable
private fun Entrega(p: PedidoDetalle) {
    val retiro = p.envio == "retiro"
    val actual = p.historial.indexOfFirst { !it.hecho }
    Column(Modifier.fillMaxWidth().tarjeta().padding(horizontal = 16.dp, vertical = 14.dp), verticalArrangement = Arrangement.spacedBy(14.dp)) {
        Column(verticalArrangement = Arrangement.spacedBy(3.dp)) {
            Texto(if (retiro) "Retiro en el local" else "Envío a domicilio", Estilo.fila, C.texto)
            p.entrega?.let { e ->
                Texto(listOf(e.direccion, listOfNotNull(e.localidad, e.provincia).joinToString(", ") + " · ${e.cp}").filter { it.isNotEmpty() }.joinToString("\n"),
                    Estilo.subhead, C.texto, interlineado = 2f)
            }
            val detalle = listOfNotNull(if (retiro) null else Marcas.envio(p.envio), p.llega?.let { "llega ${Formato.dia(it)}" }, p.seguimiento?.let { "seguimiento $it" })
            if (detalle.isNotEmpty()) Texto(Formato.mayuscula(detalle.joinToString(" · ")), Estilo.footnote, C.texto2)
        }
        Column {
            p.historial.forEachIndexed { i, paso ->
                val esActual = i == actual && p.estado != "Cancelado"
                val ultimo = i == p.historial.size - 1
                Row(Modifier.height(IntrinsicSize.Min), horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                    Column(Modifier.fillMaxHeight(), horizontalAlignment = Alignment.CenterHorizontally) {
                        val gris = tema(0x3A1D12, 0xFBF7F2, alfa = 0.18f).actual
                        when {
                            paso.hecho -> Box(Modifier.size(14.dp).clip(CircleShape).background(C.ok))
                            esActual -> Box(Modifier.size(14.dp).clip(CircleShape).background(C.superficie).border(4.dp, C.acento, CircleShape))
                            else -> Box(Modifier.size(14.dp).border(2.dp, gris, CircleShape))
                        }
                        if (!ultimo) Box(Modifier.width(2.dp).weight(1f).heightIn(min = 20.dp).background(if (paso.hecho) C.ok else tema(0x3A1D12, 0xFBF7F2, alfa = 0.12f).actual))
                    }
                    Row(Modifier.weight(1f).offset(y = (-3).dp).padding(bottom = if (ultimo) 0.dp else 12.dp)) {
                        Texto(paso.texto, if (esActual) Estilo(15f, Peso.bold) else Estilo.subhead, if (paso.hecho || esActual) C.texto else C.texto3, Modifier.weight(1f))
                        val cu = paso.cuando
                        if (cu != null && paso.hecho) {
                            Texto(if (Formato.mismoDia(cu, Instant.now())) Formato.horaDe(cu) else Formato.cuando(cu), Estilo.subhead, C.texto2, numeros = true)
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun Factura(p: PedidoDetalle, url: String?, facturando: Boolean, facturar: () -> Unit) {
    val estado = LocalEstado.current
    val ctx = LocalContext.current
    Column(Modifier.fillMaxWidth().tarjeta()) {
        Row(Modifier.padding(horizontal = 16.dp, vertical = 14.dp), horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
            Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                Texto("Factura de ARCA", Estilo.fila, C.texto)
                Texto(p.factura ?: "Todavía no la emitiste", Estilo.footnote, C.texto2)
            }
            if (p.factura == null && p.estado != "Cancelado" && estado.puede("finanzas")) {
                if (facturando) Rueda() else BotonTonal("Emitir factura", facturar)
            }
        }
        if (url != null) {
            Separador()
            Row(Modifier.fillMaxWidth().clickable { compartirTexto(ctx, url) }.padding(horizontal = 16.dp, vertical = 14.dp), verticalAlignment = Alignment.CenterVertically) {
                Texto("Comprobante del pedido", Estilo.body16, C.texto, Modifier.weight(1f))
                Icono("square.and.arrow.up", 17f, C.primarioTexto)
            }
        }
        if (p.envio != "retiro" && p.entrega != null) {
            Separador()
            Row(Modifier.fillMaxWidth().clickable {
                EtiquetaEnvio.pdf(ctx, p, estado.tienda?.nombre ?: "")?.let { compartirArchivo(ctx, it, "application/pdf") }
            }.padding(horizontal = 16.dp, vertical = 14.dp), verticalAlignment = Alignment.CenterVertically) {
                Texto("Etiqueta de envío", Estilo.body16, C.texto, Modifier.weight(1f))
                Icono("printer", 17f, C.primarioTexto)
            }
        }
    }
}
