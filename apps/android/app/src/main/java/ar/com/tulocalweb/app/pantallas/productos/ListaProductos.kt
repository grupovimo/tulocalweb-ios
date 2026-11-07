package ar.com.tulocalweb.app.pantallas.productos

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
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
import ar.com.tulocalweb.app.nucleo.ErrorApi
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.Guardado
import ar.com.tulocalweb.app.nucleo.PaginaProductos
import ar.com.tulocalweb.app.nucleo.ProductoFila
import ar.com.tulocalweb.app.nucleo.Tincho
import ar.com.tulocalweb.app.nucleo.compartirTexto
import ar.com.tulocalweb.app.nucleo.errorDe
import ar.com.tulocalweb.app.nucleo.opcional
import ar.com.tulocalweb.app.pantallas.compartir.Compartir
import ar.com.tulocalweb.app.tema.*
import kotlinx.coroutines.delay
import java.time.Instant

// Clon de Pantallas/Productos/ListaProductos.swift.

/** Stock bajo: con menos de 5 se avisa (11m). */
const val stockBajo = 5

private val segmentosProductos: List<Pair<String?, String>> = listOf(null to "Todos", "active,out" to "Publicados", "paused,draft" to "Pausados", "out" to "Sin stock")
private val ordenesProductos = listOf("sold" to "Más vendidos", "stock" to "Menos stock", "price" to "Mayor precio", "margin" to "Mayor margen")

/** 5a · Productos: buscador, chips (Todos, Publicados, Pausados, Sin stock), orden, filas con precio, stock (bajo en
 *  `warn-ink`, cero en `error-ink`) y etiquetas de canal; los pausados al 75 %. Tocar una fila abre el editor; mantenerla
 *  apretada, el stock (11m). Sin conexión, lo guardado con el banner. */
@Composable
fun ListaProductos() {
    val estado = LocalEstado.current
    val nav = LocalNavegacion.current
    val ctx = LocalContext.current
    val foco = LocalFocusManager.current
    var segmento by remember { mutableStateOf<String?>(null) }
    var orden by remember { mutableStateOf("sold") }
    var busqueda by remember { mutableStateOf("") }
    var carga by remember { mutableStateOf<Carga<PaginaProductos>>(Carga.cargando) }
    var guardadoEn by remember { mutableStateOf<Instant?>(null) }
    var stockDe by remember { mutableStateOf<ProductoFila?>(null) }
    var cargandoMas by remember { mutableStateOf(false) }
    val lista = rememberLazyListState()

    fun query(desde: Int): List<Pair<String, String>> = buildList {
        add("orden" to orden); add("desde" to desde.toString()); add("cuantos" to "60")
        segmento?.let { add("seg" to it) }
        val b = busqueda.trim()
        if (b.isNotEmpty()) add("q" to b)
    }
    val clave = "productos-${segmento ?: "todos"}-$orden"

    suspend fun cargar() {
        try {
            val l: PaginaProductos = estado.api.get("/v1/panel/listas/productos", query(0))
            carga = Carga.listo(l)
            guardadoEn = null
            if (busqueda.isEmpty()) Guardado.guardar(l, clave)
        } catch (e: kotlinx.coroutines.CancellationException) {
            throw e
        } catch (e: Exception) {
            val x = errorDe(e)
            val g = if (x.sinConexion) Guardado.leer<PaginaProductos>(clave) else null
            if (g != null) { carga = Carga.listo(g.first); guardadoEn = g.second }
            else if (carga.valor == null || !x.sinConexion) carga = Carga.error(x)
        }
    }

    suspend fun cargarMas() {
        val d = carga.valor ?: return
        if (cargandoMas) return
        cargandoMas = true
        try {
            estado.api.opcional<PaginaProductos>("/v1/panel/listas/productos", query(d.filas.size))?.let { l ->
                val ya = d.filas.map { it.id }.toSet()
                carga = Carga.listo(d.copy(filas = d.filas + l.filas.filter { it.id !in ya }, total = l.total))
            }
        } finally { cargandoMas = false }
    }

    Refrescable({ cargar() }) {
        LazyColumn(Modifier.fillMaxSize(), state = lista, contentPadding = PaddingValues(bottom = 90.dp)) {
            guardadoEn?.let { g -> item { Box(Modifier.padding(horizontal = 16.dp).padding(top = 2.dp)) { BannerSinConexion(g) } } }
            item {
                Box(Modifier.padding(top = 8.dp).margenPantalla()) {
                    EncabezadoGrande("Productos") {
                        BotonRedondo("magnifyingglass", "Buscar") { nav.buscar = true }
                        BotonRedondo("barcode.viewfinder", "Escanear un código") { nav.productos.add(Destino.inventario) }
                    }
                }
            }
            if (estado.puede("catalogo")) item {
                Row(Modifier.padding(top = 12.dp).margenPantalla(), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    Row(Modifier.height(42.dp).clip(Capsula).background(C.primario).clickable { nav.productos.add(Destino.nuevoProducto(null)) }.padding(horizontal = 16.dp),
                        verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(6.dp)) {
                        Icono("plus", 15f, Color.White); Texto("Nuevo producto", Estilo.subheadFuerte, Color.White)
                    }
                    Row(Modifier.height(42.dp).clip(Capsula).border(1.dp, C.linea2, Capsula).clickable { nav.productos.add(Destino.importar) }.padding(horizontal = 16.dp),
                        verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(6.dp)) {
                        Icono("arrow.down.doc", 15f, C.texto); Texto("Traer mi tienda", Estilo.subheadFuerte, C.texto)
                    }
                }
            }
            carga.valor?.segmentos?.let { seg ->
                item {
                    Box(Modifier.padding(top = 12.dp)) {
                        TiraEstados(listOf(
                            EstadoTira("Publicados", (seg["active"] ?: 0) + (seg["out"] ?: 0), C.ok, "active,out"),
                            EstadoTira("Sin stock", seg["out"] ?: 0, C.errorTinta, "out"),
                            EstadoTira("Pausados", (seg["paused"] ?: 0) + (seg["draft"] ?: 0), C.texto3, "paused,draft"),
                        ), segmento) { segmento = it }
                    }
                }
            }
            item { Box(Modifier.padding(top = 12.dp).margenPantalla()) { Buscador("Nombre, SKU o código", busqueda, { busqueda = it }) } }
            item {
                val seg = carga.valor?.segmentos ?: emptyMap()
                val total = seg.values.sum()
                LazyRow(Modifier.padding(top = 12.dp), contentPadding = PaddingValues(horizontal = margen), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    items(segmentosProductos, key = { it.second }) { s ->
                        val n: Int? = if (s.first == null) total else if (s.first == "out") seg["out"] else null
                        ChipFiltro(s.second, n, cuentaAlerta = s.first == "out", activo = segmento == s.first) { segmento = s.first; Haptico.seleccion() }
                    }
                }
            }
            carga.valor?.let { d ->
                if (d.filas.isNotEmpty()) item {
                    Row(Modifier.padding(horizontal = margen + 4.dp).padding(top = 14.dp, bottom = 8.dp), verticalAlignment = Alignment.CenterVertically) {
                        Texto(Formato.cantidad(d.total, "producto", "productos"), Estilo.texto14, C.texto2, Modifier.weight(1f))
                        MenuBoton(ordenesProductos.map { (v, t) -> OpcionDeMenu(t, if (orden == v) "checkmark" else "") { orden = v } }, "Orden") {
                            Row(horizontalArrangement = Arrangement.spacedBy(4.dp), verticalAlignment = Alignment.CenterVertically) {
                                Texto(ordenesProductos.firstOrNull { it.first == orden }?.second ?: "", Estilo.texto14.con(Peso.semibold), C.primarioTexto)
                                Icono("chevron.down", 11f, C.primarioTexto)
                            }
                        }
                    }
                }
            }
            when (val c = carga) {
                Carga.cargando -> item { Box(Modifier.margenPantalla().padding(top = 24.dp)) { EsqueletoFilas(5, 56.dp) } }
                is Carga.error -> item { Box(Modifier.padding(top = 60.dp)) { EstadoError("los productos", c.e) { estado.tarea { cargar() } } } }
                is Carga.listo -> {
                    val d = c.v
                    if (d.filas.isEmpty()) item { Box(Modifier.padding(top = 50.dp)) { VacioProductos(busqueda, segmento, { segmento = null }) { nav.productos.add(Destino.nuevoProducto(null)) } } }
                    else {
                        item {
                            Column(Modifier.margenPantalla().fillMaxWidth().tarjeta()) {
                                d.filas.forEachIndexed { i, p ->
                                    if (i > 0) Separador(82.dp)
                                    val t = estado.tienda
                                    MenuContextual(buildList {
                                        add(OpcionDeMenu("Ajustar stock", "shippingbox") { stockDe = p })
                                        if (t != null && p.publicado) {
                                            val u = t.url(p.id)
                                            add(OpcionDeMenu("Compartir", "square.and.arrow.up") { compartirTexto(ctx, Compartir.mensajeProducto(p.nombre, p.precio, u)) })
                                            add(OpcionDeMenu("Copiar link", "doc.on.doc") { Compartir.copiar(ctx, u, estado, "Copiamos el link del producto") })
                                        }
                                    }, { nav.productos.add(Destino.editorProducto(p.id)) }) { FilaProducto(p) }
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

    LaunchedEffect(segmento, orden, estado.version) { cargar() }
    LaunchedEffect(busqueda) {
        if (carga.valor == null) return@LaunchedEffect
        delay(300)
        cargar()
    }
    LaunchedEffect(nav.segmentoProductos) { nav.segmentoProductos?.let { segmento = it; nav.segmentoProductos = null } }
    LaunchedEffect(lista.isScrollInProgress) { if (lista.isScrollInProgress) foco.clearFocus() }
    Hoja(stockDe != null, { stockDe = null }) { stockDe?.let { p -> StockProducto(p) { cargar() } } }
}

@Composable
private fun VacioProductos(busqueda: String, segmento: String?, verTodos: () -> Unit, cargar: () -> Unit) {
    when {
        busqueda.isNotEmpty() -> EstadoTincho(Tincho.lupa, "No encontramos «$busqueda»", "Probá con otra parte del nombre o con el SKU.")
        segmento == "out" -> EstadoTincho(Tincho.exito, "No hay productos sin stock", "Todo lo que publicaste se puede comprar.", accion = "Ver todos", alTocar = verTodos)
        segmento != null -> EstadoTincho(Tincho.caja, "No hay productos acá", "Probá con otro filtro.", accion = "Ver todos", alTocar = verTodos)
        else -> EstadoTincho(Tincho.foto, "Cargá tu primer producto", "Con una foto y el precio alcanza para empezar a vender.", accion = "Cargar producto", principal = true, alTocar = cargar)
    }
}

/** La fila de un producto (thumbnail 56). */
@Composable
fun FilaProducto(p: ProductoFila) {
    val pausado = !p.publicado
    Row(Modifier.fillMaxWidth().padding(horizontal = 14.dp, vertical = 12.dp).alpha(if (pausado) 0.75f else 1f), horizontalArrangement = Arrangement.spacedBy(12.dp),
        verticalAlignment = Alignment.CenterVertically) {
        Miniatura(p.foto, 56.dp, radio = 13.dp)
        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(3.dp)) {
            Texto(p.nombre, Estilo.fila, C.texto, lineas = 2)
            val tx = C.texto; val err = C.errorTinta; val av = C.avisoTinta
            TextoRico(buildAnnotatedString {
                withStyle(SpanStyle(fontWeight = FontWeight.Bold, color = tx)) { append(Formato.pesos(p.precio)) }
                append(" · ")
                when {
                    p.stock == 0 -> withStyle(SpanStyle(fontWeight = FontWeight.SemiBold, color = err)) { append("Sin stock") }
                    p.stock < stockBajo -> withStyle(SpanStyle(fontWeight = FontWeight.SemiBold, color = av)) { append("Quedan ${p.stock}") }
                    else -> append("${Formato.numero(p.stock)}${Formato.fino}u.")
                }
            }, Estilo.texto14, C.texto2, numeros = true)
            Row(horizontalArrangement = Arrangement.spacedBy(4.dp)) {
                if (pausado) EtiquetaCanal("En pausa", C.apagadoFondo, C.neutroTinta)
                else {
                    if ("web" in p.canales || p.canales.isEmpty()) EtiquetaCanal("Web", C.pista, C.neutroTinta)
                    if ("ml" in p.canales) EtiquetaCanal("ML", C.avisoFondo, C.bannerTinta)
                }
            }
        }
    }
}
