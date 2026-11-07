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
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalFocusManager
import androidx.compose.ui.text.SpanStyle
import androidx.compose.ui.text.buildAnnotatedString
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.withStyle
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.app.Destino
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.app.LocalNavegacion
import ar.com.tulocalweb.app.nucleo.Api
import ar.com.tulocalweb.app.nucleo.ErrorApi
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.Guardado
import ar.com.tulocalweb.app.nucleo.Marcas
import ar.com.tulocalweb.app.nucleo.PaginaPedidos
import ar.com.tulocalweb.app.nucleo.PedidoDetalle
import ar.com.tulocalweb.app.nucleo.PedidoFila
import ar.com.tulocalweb.app.nucleo.Tincho
import ar.com.tulocalweb.app.nucleo.abrirUrl
import ar.com.tulocalweb.app.nucleo.errorDe
import ar.com.tulocalweb.app.nucleo.intentar
import ar.com.tulocalweb.app.nucleo.opcional
import ar.com.tulocalweb.app.tema.*
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import kotlinx.serialization.Serializable
import java.time.Instant

// Clon de Pantallas/Pedidos/ListaPedidos.swift.

private data class DatosPedidos(
    val filas: List<PedidoFila>,
    val total: Int,
    val etapas: Map<String, Int>,
    /** La suma en pesos de los pedidos del filtro (0 si el rol no ve montos). */
    val suma: Int = 0,
)

private val etapasPedidos: List<Pair<String?, String>> = listOf(
    null to "Todos", "Para preparar" to "Para preparar", "Sin pagar" to "Sin pagar",
    "Listo para despachar" to "Listo para despachar", "En camino" to "En camino",
    "Entregado" to "Entregados", "Cancelado" to "Cancelados",
)

/** 3a · Pedidos: buscador, chips por etapa con contador, secciones por día y filas con swipe (izquierda: «Más» y el
 *  siguiente estado; derecha: «Cobrado» en los sin pagar), toast con «Deshacer». 3d vacío con filtro, 12f error, 5a sin
 *  conexión. */
@Composable
fun ListaPedidos() {
    val estado = LocalEstado.current
    val nav = LocalNavegacion.current
    val ctx = LocalContext.current
    val alcance = rememberCoroutineScope()
    val foco = LocalFocusManager.current
    var filtro by remember { mutableStateOf<String?>(null) }
    var orden by remember { mutableStateOf("nuevos") }
    var busqueda by remember { mutableStateOf("") }
    var carga by remember { mutableStateOf<Carga<DatosPedidos>>(Carga.cargando) }
    var guardadoEn by remember { mutableStateOf<Instant?>(null) }
    var cargandoMas by remember { mutableStateOf(false) }
    var masDe by remember { mutableStateOf<PedidoFila?>(null) }
    // «Cancelar pedido» desde «Más»: el pedido entero para la hoja de cancelar.
    var cancelando by remember { mutableStateOf<PedidoDetalle?>(null) }
    val lista = rememberLazyListState()
    val chipsEstado = rememberLazyListState()

    fun query(desde: Int): List<Pair<String, String>> = buildList {
        add("orden" to orden); add("desde" to desde.toString()); add("cuantos" to "60")
        filtro?.let { add("estado" to it) }
        val b = busqueda.trim()
        if (b.isNotEmpty()) add("q" to b)
    }
    val clave = "pedidos-${filtro ?: "todos"}-$orden"

    suspend fun cargar() {
        try {
            val l: PaginaPedidos = estado.api.get("/v1/panel/pedidos", query(0))
            carga = Carga.listo(DatosPedidos(l.filas, l.total, l.etapas, l.suma))
            guardadoEn = null
            if (busqueda.isEmpty()) Guardado.guardar(l, clave)
        } catch (e: kotlinx.coroutines.CancellationException) {
            throw e
        } catch (e: Exception) {
            val x = errorDe(e)
            val g = if (x.sinConexion) Guardado.leer<PaginaPedidos>(clave) else null
            when {
                g != null -> { carga = Carga.listo(DatosPedidos(g.first.filas, g.first.filas.size, g.first.etapas)); guardadoEn = g.second }
                carga is Carga.listo && x.sinConexion -> estado.avisar("Sin conexión. Probá de nuevo en un rato.", ok = false)
                else -> carga = Carga.error(x)
            }
        }
    }

    suspend fun cargarMas() {
        val d = carga.valor ?: return
        if (cargandoMas) return
        cargandoMas = true
        try {
            estado.api.opcional<PaginaPedidos>("/v1/panel/pedidos", query(d.filas.size))?.let { l ->
                val ya = d.filas.map { it.numero }.toSet()
                carga = Carga.listo(d.copy(filas = d.filas + l.filas.filter { it.numero !in ya }, total = l.total))
            }
        } finally { cargandoMas = false }
    }

    /** Cambio de estado optimista: se ve al instante, vuelve atrás si la API no lo acepta, y el toast ofrece «Deshacer»
     *  (el estado anterior). */
    suspend fun abrirCancelar(p: PedidoFila) {
        try {
            cancelando = estado.api.get<PedidoDetalle>("/v1/panel/pedidos/${p.numero}")
        } catch (e: ErrorApi) {
            Haptico.error()
            estado.avisar(e.detalle, ok = false)
        } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {}
    }

    suspend fun mover(p: PedidoFila, nuevo: String, texto: String? = null) {
        val d = carga.valor ?: return
        val i = d.filas.indexOfFirst { it.numero == p.numero }
        if (i < 0) return
        val antes = d.filas[i]
        Haptico.impacto()
        var filas = d.filas.toMutableList().also { it[i] = antes.copy(estado = nuevo) }.toList()
        // Si quedó afuera del filtro, sale de la lista.
        val f = filtro
        if (f != null && f != nuevo) filas = filas.filterIndexed { j, _ -> j != i }
        val etapas = d.etapas.toMutableMap()
        etapas[antes.estado] = (etapas[antes.estado] ?: 1) - 1
        etapas[nuevo] = (etapas[nuevo] ?: 0) + 1
        carga = Carga.listo(d.copy(filas = filas, etapas = etapas))
        try {
            PedidosApi.cambiarEstado(estado.api, p.numero, nuevo)
            estado.avisar(texto ?: "#${p.numero} ${Marcas.dicho(nuevo, p.envio)}") {
                estado.tarea {
                    intentar { PedidosApi.cambiarEstado(estado.api, p.numero, antes.estado) }
                    cargar()
                    estado.refrescarResumen()
                }
            }
            estado.refrescarResumen()
        } catch (e: ErrorApi) {
            Haptico.error()
            estado.avisar(e.detalle, ok = false)
            cargar()
        } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {}
    }

    fun tomarFiltro() { nav.filtroPedidos?.let { filtro = it; nav.filtroPedidos = null } }

    // La carga se lee acá y no adentro de la lista: así, cuando llegan los pedidos, se arma la lista de nuevo entera
    // (leída sólo adentro, el primer render quedaba con el buscador repetido en el lugar del resumen).
    val cargaVista = carga
    Refrescable({ cargar() }) {
        LazyColumn(Modifier.fillMaxSize(), state = lista, contentPadding = androidx.compose.foundation.layout.PaddingValues(bottom = 24.dp)) {
            guardadoEn?.let { g -> item { Box(Modifier.padding(horizontal = 16.dp).padding(top = 2.dp, bottom = 8.dp)) { BannerSinConexion(g) } } }
            item {
                Box(Modifier.padding(top = 4.dp).margenPantalla()) {
                    EncabezadoGrande("Pedidos") {
                        BotonRedondo("magnifyingglass", "Buscar") { nav.buscar = true }
                        MenuBoton(listOf("nuevos" to "Más nuevos primero", "urgentes" to "Más urgentes primero", "monto" to "Mayor monto primero", "viejos" to "Más viejos primero")
                            .map { (v, t) -> OpcionDeMenu(t, if (orden == v) "checkmark" else "", accion = { orden = v }) }, "Ordenar") {
                            Box(Modifier.size(44.dp), contentAlignment = Alignment.Center) {
                                Box(Modifier.size(40.dp).clip(CircleShape).background(C.superficie).border(1.dp, C.linea, CircleShape), contentAlignment = Alignment.Center) {
                                    Icono("line.3.horizontal.decrease", 17f, C.texto)
                                }
                            }
                        }
                    }
                }
            }
            cargaVista.valor?.let { d ->
                item { Box(Modifier.padding(top = 2.dp).margenPantalla()) { ResumenPedidos(d) } }
                item {
                    val e = d.etapas
                    Box(Modifier.padding(top = 12.dp)) {
                        TiraEstados(listOf(
                            EstadoTira("Sin pagar", e["Sin pagar"] ?: 0, C.aviso, "Sin pagar"),
                            EstadoTira("Para preparar", e["Para preparar"] ?: 0, C.acento, "Para preparar"),
                            EstadoTira("Listo para despachar", e["Listo para despachar"] ?: 0, hex(0xC9962C), "Listo para despachar"),
                            EstadoTira("En camino", e["En camino"] ?: 0, hex(0x4A7FB5), "En camino"),
                            EstadoTira("Entregados", e["Entregado"] ?: 0, C.ok, "Entregado"),
                        ), filtro) { filtro = it }
                    }
                }
            }
            item { Box(Modifier.padding(top = 12.dp).margenPantalla()) { Buscador("Número, cliente o producto", busqueda, { busqueda = it }) } }
            item {
                val etapas = carga.valor?.etapas ?: emptyMap()
                val total = etapas.values.sum()
                LazyRow(Modifier.padding(top = 12.dp), state = chipsEstado, contentPadding = androidx.compose.foundation.layout.PaddingValues(horizontal = margen),
                    horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    items(etapasPedidos, key = { it.second }) { e ->
                        val n = if (e.first == null) total else (etapas[e.first] ?: 0)
                        val mostrarN = e.first == null || e.first == "Para preparar" || e.first == "Sin pagar"
                        ChipFiltro(e.second, if (mostrarN) n else null, activo = filtro == e.first) { filtro = e.first; Haptico.seleccion() }
                    }
                }
            }
            when (val c = cargaVista) {
                Carga.cargando -> item { Box(Modifier.margenPantalla().padding(top = 24.dp)) { EsqueletoFilas(5) } }
                is Carga.error -> item { Box(Modifier.padding(top = 60.dp)) { EstadoError("los pedidos", c.e) { estado.tarea { cargar() } } } }
                is Carga.listo -> {
                    val d = c.v
                    if (d.filas.isEmpty()) item { Box(Modifier.padding(top = 50.dp)) { VacioPedidos(busqueda, filtro) { filtro = null } } }
                    else {
                        secciones(d.filas).forEach { (titulo, filas) ->
                            item(key = "t-$titulo") { Texto(titulo, Estilo.label, C.texto2, Modifier.padding(top = 18.dp, bottom = 6.dp).margenPantalla()) }
                            item(key = "s-$titulo") {
                                Column(Modifier.margenPantalla().fillMaxWidth().tarjeta()) {
                                    filas.forEachIndexed { i, p ->
                                        if (i > 0) Separador(74.dp)
                                        FilaConAcciones(p, { masDe = p }, { n, t -> alcance.launch { mover(p, n, t) } }) { nav.pedidos.add(Destino.pedido(p.numero)) }
                                    }
                                }
                            }
                        }
                        if (d.filas.size < d.total) item {
                            Box(Modifier.fillMaxWidth().padding(20.dp), contentAlignment = Alignment.Center) { Rueda() }
                            LaunchedEffect(d.filas.size) { cargarMas() }
                        }
                    }
                }
            }
        }
    }

    LaunchedEffect(filtro, orden, estado.version) { cargar() }
    LaunchedEffect(busqueda) {
        if (busqueda.isEmpty() && carga.valor == null) return@LaunchedEffect
        delay(300)
        cargar()
    }
    LaunchedEffect(nav.filtroPedidos) { tomarFiltro() }
    LaunchedEffect(filtro) {
        val i = etapasPedidos.indexOfFirst { it.first == filtro }.coerceAtLeast(0)
        runCatching { chipsEstado.animateScrollToItem(maxOf(0, i - 1)) }
    }
    LaunchedEffect(lista.isScrollInProgress) { if (lista.isScrollInProgress) foco.clearFocus() }

    val p = masDe
    Confirmacion(p != null, p?.let { "Pedido #${it.numero}" } ?: "", opciones = buildList {
        if (p == null) return@buildList
        add(OpcionConfirmacion("Ver el pedido") { nav.pedidos.add(Destino.pedido(p.numero)) })
        p.telefonoWhatsApp?.let { u -> add(OpcionConfirmacion("Escribirle por WhatsApp") { abrirUrl(ctx, u) }) }
        val ant = Marcas.anterior(p.estado, p.envio)
        if (ant != null && p.estado != "Cancelado") add(OpcionConfirmacion("Volver a «${ChipEstado.texto(ant, p.envio)}»") { alcance.launch { mover(p, ant) } })
        if (p.estado != "Entregado" && p.estado != "Cancelado") add(OpcionConfirmacion("Cancelar pedido", destructiva = true) { alcance.launch { abrirCancelar(p) } })
    }, alCerrar = { masDe = null }, volver = "Cancelar")
    Hoja(cancelando != null, { cancelando = null }) {
        cancelando?.let { c -> CancelarPedido(c) { cargar(); estado.refrescarResumen() } }
    }
}

/** «12 pedidos · $ 340.000 · ticket promedio $ 28.300». */
@Composable
private fun ResumenPedidos(d: DatosPedidos) {
    val total = d.etapas.values.sum()
    val suma = if (d.suma > 0) d.suma else null
    TextoRico(buildAnnotatedString {
        withStyle(SpanStyle(fontWeight = FontWeight.Bold)) { append(Formato.numero(total)) }
        append(if (total == 1) " pedido" else " pedidos")
        if (suma != null) append(" · ${Formato.pesos(suma)}" + (if (total > 0) " · ticket promedio ${Formato.pesos(suma / maxOf(1, total))}" else ""))
    }, Estilo.subhead, C.texto2)
}

@Composable
private fun VacioPedidos(busqueda: String, filtro: String?, verTodos: () -> Unit) {
    when {
        busqueda.isNotEmpty() -> EstadoTincho(Tincho.lupa, "No encontramos «$busqueda»", "Probá con el número del pedido, el nombre de quien compró o un producto.")
        filtro != null -> {
            val (tincho, titulo, texto) = when (filtro) {
                "En camino" -> Triple(Tincho.envio, "No hay pedidos en camino", "Cuando despaches uno, lo ves acá hasta que llegue.")
                "Para preparar" -> Triple(Tincho.caja, "No hay pedidos por preparar", "Cuando entre uno nuevo, te avisamos al instante.")
                "Sin pagar" -> Triple(Tincho.ticket, "No hay pedidos sin pagar", "Todos los pedidos están cobrados.")
                "Listo para despachar" -> Triple(Tincho.caja, "No hay pedidos listos", "Los que prepares quedan acá hasta que salgan.")
                else -> Triple(Tincho.portapapeles, "No hay pedidos ${ChipEstado.texto(filtro, "").lowercase()}", "Cuando haya, los ves acá.")
            }
            EstadoTincho(tincho, titulo, texto, accion = "Ver todos los pedidos", alTocar = verTodos)
        }
        else -> EstadoTincho(Tincho.durmiendo, "Todavía no tenés pedidos", "Compartí el link de tu tienda y el primero llega solo.")
    }
}

@Composable
private fun FilaConAcciones(p: PedidoFila, mas: () -> Unit, mover: (String, String?) -> Unit, abrir: () -> Unit) {
    val derecha = buildList {
        add(AccionFila("Más", "ellipsis", hex(0x7D6656), mas))
        val sig = Marcas.siguiente(p.estado, p.envio)
        if (p.estado != "Sin pagar" && sig != null) add(AccionFila(sig.corto, "checkmark", hex(0xB0441D)) { mover(sig.estado, null) })
    }
    val izquierda = if (p.estado == "Sin pagar") listOf(AccionFila("Cobrado", "banknote", T.ok.claro) { mover("Para preparar", "#${p.numero} cobrado") }) else emptyList()
    FilaDeslizable(izquierda, derecha) {
        Box(Modifier.clickable(onClick = abrir)) { FilaPedido(p) }
    }
}

private fun secciones(filas: List<PedidoFila>): List<Pair<String, List<PedidoFila>>> {
    val r = linkedMapOf<String, MutableList<PedidoFila>>()
    filas.forEach { f -> r.getOrPut(Formato.seccion(f.creado)) { mutableListOf() }.add(f) }
    return r.map { it.key to it.value }
}

/** La fila de un pedido (thumbnail 48, nombre y monto 16/600, meta 13 y chip). */
@Composable
fun FilaPedido(p: PedidoFila) {
    Row(Modifier.fillMaxWidth().padding(horizontal = 14.dp, vertical = 13.dp), horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
        Miniatura(p.fotoProducto ?: p.foto, 48.dp, radio = 12.dp)
        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(5.dp)) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Texto(p.cliente, Estilo.fila, C.texto, Modifier.weight(1f).padding(end = 8.dp), lineas = 1)
                Texto(Formato.pesos(p.total), Estilo.fila, C.texto, numeros = true)
            }
            Row(verticalAlignment = Alignment.CenterVertically) {
                Texto("#${p.numero} · ${Marcas.canalCorto(p.canal)} · ${Formato.horaDe(p.creado)}", Estilo.footnote, C.texto2, Modifier.weight(1f).padding(end = 6.dp), lineas = 1)
                ChipEstado(p.estado, p.envio)
            }
        }
    }
}

val PedidoFila.telefonoWhatsApp: String? get() = urlWhatsApp(telefono)

/** wa.me con el número en formato internacional (Argentina: 54 9 + área + número). */
fun urlWhatsApp(telefono: String?, texto: String? = null): String? {
    val t = telefono ?: return null
    var d = t.filter { it.isDigit() }
    if (d.length < 8) return null
    if (d.startsWith("0")) d = d.drop(1)
    if (!d.startsWith("54")) d = "549$d"
    return "https://wa.me/$d" + (texto?.let { "?text=" + java.net.URLEncoder.encode(it, "UTF-8").replace("+", "%20") } ?: "")
}

object PedidosApi {
    @Serializable private data class Cuerpo(val estado: String)

    suspend fun cambiarEstado(api: Api, numero: Int, a: String) {
        api.hacer("POST", "/v1/panel/pedidos/$numero/estado", Cuerpo(a))
    }
}
