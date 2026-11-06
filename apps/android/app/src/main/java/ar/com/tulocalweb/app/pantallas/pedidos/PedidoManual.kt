package ar.com.tulocalweb.app.pantallas.pedidos

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.rememberScrollState
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
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.app.Destino
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.app.LocalNavegacion
import ar.com.tulocalweb.app.app.Pestana
import ar.com.tulocalweb.app.nucleo.ErrorApi
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.PaginaProductos
import ar.com.tulocalweb.app.nucleo.ProductoFila
import ar.com.tulocalweb.app.nucleo.ValorJSON
import ar.com.tulocalweb.app.nucleo.entero
import ar.com.tulocalweb.app.nucleo.get
import ar.com.tulocalweb.app.nucleo.opcional
import ar.com.tulocalweb.app.tema.*
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import kotlinx.serialization.Serializable

// Clon de Pantallas/Pedidos/PedidoManual.swift.

private data class ItemManual(val producto: ProductoFila, val variante: ProductoFila.Variante?, val cantidad: Int) {
    val id: String get() = producto.id + (variante?.id ?: "")
    val precio: Int get() = variante?.precio ?: producto.precio
}

private val canalesManual = listOf("wa" to "WhatsApp", "ig" to "Instagram", "web" to "Teléfono / otro")
private val pagosManual = listOf("transf" to "Transferencia", "efectivo" to "Efectivo", "mp" to "Mercado Pago", "tarjeta" to "Tarjeta")
private val enviosManual = listOf("retiro" to "Retira", "moto" to "Moto", "correo" to "Correo")

/** 11a · Pedido manual: para ventas por WhatsApp, Instagram o teléfono. Productos (con variante si tiene), cliente, por
 *  dónde llegó, cómo paga, la entrega y si ya está pagado. */
@Composable
fun PedidoManual() {
    val estado = LocalEstado.current
    val nav = LocalNavegacion.current
    val alcance = rememberCoroutineScope()
    val items = remember { mutableStateListOf<ItemManual>() }
    var buscar by remember { mutableStateOf(false) }
    var cliente by remember { mutableStateOf("") }
    var telefono by remember { mutableStateOf("") }
    var nota by remember { mutableStateOf("") }
    var canal by remember { mutableStateOf("wa") }
    var pago by remember { mutableStateOf("transf") }
    var envio by remember { mutableStateOf("retiro") }
    var pagado by remember { mutableStateOf(false) }
    var guardando by remember { mutableStateOf(false) }
    val total = items.sumOf { it.precio * it.cantidad }

    @Serializable data class It(val productoId: String, val varianteId: String?, val cantidad: Int)
    @Serializable data class Cuerpo(val items: List<It>, val pago: String, val canal: String, val envio: String,
                                    val cliente: String?, val telefono: String?, val nota: String?, val pagado: Boolean)

    suspend fun crear() {
        fun vacio(s: String) = s.trim().ifEmpty { null }
        guardando = true
        try {
            val r: ValorJSON = estado.api.enviar("POST", "/v1/panel/pedidos", Cuerpo(
                items.map { It(it.producto.id, it.variante?.id, it.cantidad) }, pago, canal, envio, vacio(cliente), vacio(telefono), vacio(nota), pagado))
            Haptico.exito()
            estado.avisar("Pedido creado")
            val n = r["numero"]?.entero
            nav.poner(Pestana.pedidos, if (n != null) listOf(Destino.pedido(n)) else emptyList())
        } catch (e: ErrorApi) { estado.avisar(e.detalle, ok = false) } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
            estado.avisar("No pudimos crear el pedido", ok = false)
        } finally { guardando = false }
    }

    @Composable
    fun CampoManual(titulo: String, texto: String, alCambiar: (String) -> Unit, teclado: KeyboardType = KeyboardType.Text) {
        val forma = RoundedCornerShape(14.dp)
        Entrada(texto, alCambiar, titulo, Modifier.fillMaxWidth().clip(forma).background(C.superficie).border(1.dp, C.linea2, forma).padding(horizontal = 14.dp, vertical = 12.dp),
            estilo = Estilo.body16, teclado = teclado)
    }

    ConBarraAccion({
        BotonPrincipal(if (guardando) "Creando…" else "Crear pedido", { alcance.launch { crear() } }, habilitado = items.isNotEmpty() && !guardando)
    }, Modifier.background(C.fondo).imePadding()) { abajo ->
        Column(Modifier.fillMaxSize().verticalScroll(rememberScrollState()).margenPantalla().padding(bottom = 32.dp + abajo), verticalArrangement = Arrangement.spacedBy(14.dp)) {
            Texto("Pedido manual", Estilo.tituloGrande, C.texto)
            EtiquetaSeccion("Productos", Modifier.padding(horizontal = 4.dp))
            if (items.isEmpty()) {
                BotonSecundario("Agregar productos", { buscar = true }, lleno = true, simbolo = "plus")
            } else {
                Column(Modifier.fillMaxWidth().tarjeta()) {
                    items.toList().forEachIndexed { i, it ->
                        if (i > 0) Separador(70.dp)
                        Row(Modifier.padding(horizontal = 14.dp, vertical = 10.dp), horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
                            Miniatura(it.producto.foto, 44.dp)
                            Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                                Texto(it.producto.nombre, Estilo.fila, C.texto, lineas = 1)
                                Texto(listOfNotNull(it.variante?.texto, Formato.pesos(it.precio)).joinToString(" · "), Estilo.footnote, C.texto2)
                            }
                            Pasador(it.cantidad, { n -> if (n <= 0) items.removeAt(i) else items[i] = it.copy(cantidad = n) })
                            Texto("${it.cantidad}", Estilo.fila, C.texto, Modifier.widthIn(min = 22.dp), numeros = true)
                        }
                    }
                }
                Row(Modifier.padding(horizontal = 4.dp), verticalAlignment = Alignment.CenterVertically) {
                    Texto("Agregar más", Estilo.subheadFuerte, C.primarioTexto, Modifier.weight(1f).clickable { buscar = true })
                    Texto("Total ${Formato.pesos(total)}", Estilo.headline, C.texto, numeros = true)
                }
            }
            EtiquetaSeccion("Cliente", Modifier.padding(horizontal = 4.dp).padding(top = 6.dp))
            CampoManual("Nombre", cliente, { cliente = it })
            CampoManual("WhatsApp o teléfono", telefono, { telefono = it }, KeyboardType.Phone)
            EtiquetaSeccion("Por dónde llegó", Modifier.padding(horizontal = 4.dp).padding(top = 6.dp))
            Segmentado(canalesManual, canal, { canal = it })
            EtiquetaSeccion("Cómo paga", Modifier.padding(horizontal = 4.dp).padding(top = 6.dp))
            Segmentado(pagosManual, pago, { pago = it })
            Box(Modifier.fillMaxWidth().tarjeta().padding(horizontal = 16.dp, vertical = 10.dp)) { Interruptor("Ya está pagado", pagado, { pagado = it }) }
            EtiquetaSeccion("Entrega", Modifier.padding(horizontal = 4.dp).padding(top = 6.dp))
            Segmentado(enviosManual, envio, { envio = it })
            CampoManual("Nota (opcional)", nota, { nota = it })
            if (!pagado) {
                Texto(Formato.sinViudas("Queda «Sin pagar»: desde el pedido le mandás el link de pago o confirmás la transferencia."), Estilo.footnote, C.texto3, Modifier.padding(horizontal = 4.dp))
            }
        }
    }
    Hoja(buscar, { buscar = false }) {
        ElegirProducto { p, v ->
            val i = items.indexOfFirst { it.producto.id == p.id && it.variante?.id == v?.id }
            if (i >= 0) items[i] = items[i].copy(cantidad = items[i].cantidad + 1) else items.add(ItemManual(p, v, 1))
        }
    }
}

/** Buscar un producto (y su variante) para agregarlo. */
@Composable
fun ElegirProducto(alElegir: (ProductoFila, ProductoFila.Variante?) -> Unit) {
    val estado = LocalEstado.current
    val cerrar = LocalCerrarHoja.current
    var busqueda by remember { mutableStateOf("") }
    var filas by remember { mutableStateOf<List<ProductoFila>>(emptyList()) }
    var variantesDe by remember { mutableStateOf<ProductoFila?>(null) }

    LaunchedEffect(busqueda) {
        delay(250)
        val q = buildList { add("cuantos" to "40"); add("seg" to "active,out"); if (busqueda.isNotEmpty()) add("q" to busqueda) }
        estado.api.opcional<PaginaProductos>("/v1/panel/listas/productos", q)?.let { filas = it.filas }
    }

    Column(Modifier.fillMaxSize().imePadding().padding(horizontal = 20.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
        Grabber()
        Texto("Agregar productos", Estilo.tituloHojaChico, C.texto)
        Buscador("Nombre, SKU o código", busqueda, { busqueda = it })
        Column(Modifier.weight(1f).verticalScroll(rememberScrollState())) {
            Column(Modifier.fillMaxWidth().tarjeta()) {
                filas.forEachIndexed { i, p ->
                    if (i > 0) Separador(70.dp)
                    Row(Modifier.fillMaxWidth().clickable {
                        if (p.variantes.isEmpty()) { alElegir(p, null); estado.avisar("Agregamos ${p.nombre}") } else variantesDe = p
                    }.padding(horizontal = 14.dp, vertical = 10.dp), horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
                        Miniatura(p.foto, 44.dp)
                        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                            Texto(p.nombre, Estilo.fila, C.texto, lineas = 1)
                            Texto(if (p.variantes.isEmpty()) "Stock ${p.stock}" else Formato.cantidad(p.variantes.size, "variante", "variantes"), Estilo.footnote, C.texto2)
                        }
                        Texto(Formato.pesos(p.precio), Estilo.fila, C.texto, numeros = true)
                    }
                }
            }
        }
        BotonPrincipal("Listo", { cerrar() }, Modifier.padding(bottom = 8.dp))
    }
    val p = variantesDe
    Confirmacion(p != null, "Elegí la variante", opciones = p?.variantes?.filter { it.activa }?.map { v ->
        OpcionConfirmacion("${v.texto} · ${Formato.pesos(v.precio ?: p.precio)}") { alElegir(p, v); estado.avisar("Agregamos ${p.nombre}") }
    } ?: emptyList(), alCerrar = { variantesDe = null }, volver = "Cancelar")
}
