package ar.com.tulocalweb.app.app

import androidx.activity.compose.BackHandler
import androidx.compose.animation.AnimatedContent
import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.core.spring
import androidx.compose.animation.core.tween
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.scaleIn
import androidx.compose.animation.scaleOut
import androidx.compose.animation.slideInHorizontally
import androidx.compose.animation.slideInVertically
import androidx.compose.animation.slideOutHorizontally
import androidx.compose.animation.slideOutVertically
import androidx.compose.animation.togetherWith
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.RowScope
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.defaultMinSize
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.SideEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.key
import androidx.compose.runtime.mutableStateListOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.runtime.snapshots.SnapshotStateList
import androidx.compose.runtime.staticCompositionLocalOf
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.selected
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.zIndex
import ar.com.tulocalweb.app.nucleo.Tincho
import ar.com.tulocalweb.app.pantallas.avisos.Avisos
import ar.com.tulocalweb.app.pantallas.buscar.Buscar
import ar.com.tulocalweb.app.pantallas.clientes.ClienteFila
import ar.com.tulocalweb.app.pantallas.clientes.Clientes
import ar.com.tulocalweb.app.pantallas.clientes.FichaCliente
import ar.com.tulocalweb.app.pantallas.compartir.CompartirTienda
import ar.com.tulocalweb.app.pantallas.inicio.Inicio
import ar.com.tulocalweb.app.pantallas.mas.*
import ar.com.tulocalweb.app.pantallas.pedidos.DetallePedido
import ar.com.tulocalweb.app.pantallas.pedidos.ListaPedidos
import ar.com.tulocalweb.app.pantallas.pedidos.PedidoManual
import ar.com.tulocalweb.app.pantallas.productos.EditorProducto
import ar.com.tulocalweb.app.pantallas.productos.ListaProductos
import ar.com.tulocalweb.app.pantallas.tincho.TinchoChat
import ar.com.tulocalweb.app.pantallas.venta.VentaLocal
import ar.com.tulocalweb.app.tema.*
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

// Clon de App/Principal.swift.

enum class Pestana { inicio, pedidos, productos, tincho, mas }

/** Las pantallas a las que se entra desde una lista (push). */
sealed class Destino {
    data class pedido(val n: Int) : Destino()
    data object avisos : Destino()
    data object cuenta : Destino()
    data object clientes : Destino()
    data class cliente(val c: ClienteFila) : Destino()
    data object caja : Destino()
    data object carritos : Destino()
    data object resenas : Destino()
    data object devoluciones : Destino()
    data object cupones : Destino()
    data object finanzas : Destino()
    data object facturacion : Destino()
    data object plan : Destino()
    data object integraciones : Destino()
    data object usuarios : Destino()
    data class editorProducto(val id: String) : Destino()
    data class nuevoProducto(val sku: String?) : Destino()
    data object pedidoManual : Destino()
    data object preferenciasAvisos : Destino()
    data object negocio : Destino()
    data object links : Destino()
    data object inventario : Destino()
    data object ayuda : Destino()
    data object miTienda : Destino()
    data object estadisticas : Destino()
    data object compartir : Destino()
    data object importar : Destino()
    data object telegram : Destino()
    data object medicion : Destino()
    data object ia : Destino()
    /** Un artículo de Ayuda (NavigationLink { ArticuloAyuda } en iOS). */
    data class articuloAyuda(val slug: String) : Destino()
}

/** La navegación de la app: la pestaña, el camino de cada una y los filtros con los que se abre una lista desde otro
 *  lado (Inicio › Para hacer hoy). */
class Navegacion {
    var pestana by mutableStateOf(Pestana.inicio)
    val inicio: SnapshotStateList<Destino> = mutableStateListOf()
    val pedidos: SnapshotStateList<Destino> = mutableStateListOf()
    val productos: SnapshotStateList<Destino> = mutableStateListOf()
    val tincho: SnapshotStateList<Destino> = mutableStateListOf()
    val mas: SnapshotStateList<Destino> = mutableStateListOf()
    /** Al abrir Pedidos con un filtro (estado de la API) o Productos con un segmento. */
    var filtroPedidos by mutableStateOf<String?>(null)
    var segmentoProductos by mutableStateOf<String?>(null)
    /** La hoja «+» (2e). */
    var crear by mutableStateOf(false)
    /** La venta en el local (4a), a pantalla completa. */
    var venta by mutableStateOf(false)
    /** El buscador (secciones, acciones y preguntarle a Tincho). */
    var buscar by mutableStateOf(false)
    /** Una pregunta para Tincho tocada desde otra pantalla (la toma el chat al abrirse). */
    var preguntaTincho by mutableStateOf<String?>(null)
    /** Una página web dentro de la app (lo que vive sólo en la web). */
    var web by mutableStateOf<String?>(null)

    fun camino(p: Pestana): SnapshotStateList<Destino> = when (p) {
        Pestana.inicio -> inicio
        Pestana.pedidos -> pedidos
        Pestana.productos -> productos
        Pestana.tincho -> tincho
        Pestana.mas -> mas
    }

    val caminoActual: SnapshotStateList<Destino> get() = camino(pestana)

    /** Entrar a una pantalla en la pestaña que se ve (NavigationLink / path.append). */
    fun abrir(d: Destino) { caminoActual.add(d) }

    /** Volver una pantalla (dismiss). */
    fun volver() { caminoActual.removeLastOrNull() }

    /** Reemplaza el camino de una pestaña (nav.pedidos = [d]). */
    fun poner(p: Pestana, camino: List<Destino>) { camino(p).apply { clear(); addAll(camino) } }

    fun ir(ruta: Ruta) {
        when (ruta) {
            is Ruta.pedido -> { pestana = Pestana.pedidos; poner(Pestana.pedidos, listOf(Destino.pedido(ruta.n))) }
            is Ruta.producto -> { pestana = Pestana.productos; poner(Pestana.productos, listOf(Destino.editorProducto(ruta.id))) }
            Ruta.productos -> { pestana = Pestana.productos; poner(Pestana.productos, emptyList()) }
            is Ruta.destino -> { pestana = Pestana.inicio; camino(Pestana.inicio).add(ruta.d) }
            is Ruta.web -> web = ruta.url
            Ruta.avisos -> { pestana = Pestana.inicio; poner(Pestana.inicio, listOf(Destino.avisos)) }
            Ruta.pedidos -> { pestana = Pestana.pedidos; poner(Pestana.pedidos, emptyList()) }
            Ruta.inicio -> { pestana = Pestana.inicio; poner(Pestana.inicio, emptyList()) }
            Ruta.compartir -> { pestana = Pestana.inicio; poner(Pestana.inicio, listOf(Destino.compartir)) }
        }
    }

    /** Tocar la pestaña activa vuelve a la raíz. */
    fun tocar(p: Pestana) {
        if (p == pestana) camino(p).clear()
        pestana = p
    }
}

val LocalNavegacion = staticCompositionLocalOf<Navegacion> { error("Falta la Navegacion") }

/** La barra de navegación de una pantalla de adentro (la nativa de iOS): lo que la pantalla le pone con BarraDe. */
class Barra {
    var titulo by mutableStateOf<String?>(null)
    var acciones by mutableStateOf<(@Composable RowScope.() -> Unit)?>(null)
    var izquierda by mutableStateOf<(@Composable RowScope.() -> Unit)?>(null)
    var oculta by mutableStateOf(false)
}

val LocalBarra = staticCompositionLocalOf { Barra() }

/** .navigationTitle / .toolbar / .toolbar(.hidden) de la pantalla que se ve. `izquierda` reemplaza el «‹ volver». */
@Composable
fun BarraDe(titulo: String? = null, oculta: Boolean = false, izquierda: (@Composable RowScope.() -> Unit)? = null, acciones: (@Composable RowScope.() -> Unit)? = null) {
    val b = LocalBarra.current
    SideEffect { b.titulo = titulo; b.oculta = oculta; b.acciones = acciones; b.izquierda = izquierda }
}

/** Un botón de texto de la barra (Button("Listo") en un ToolbarItem): `primary-text`, 17. */
@Composable
fun BotonBarra(texto: String, fuerte: Boolean = false, habilitado: Boolean = true, alTocar: () -> Unit) {
    Box(Modifier.defaultMinSize(minHeight = 44.dp).clickable(enabled = habilitado, onClick = alTocar).padding(horizontal = 8.dp).alpha(if (habilitado) 1f else 0.4f), contentAlignment = Alignment.Center) {
        Texto(texto, if (fuerte) Estilo.headline else Estilo.body, C.primarioTexto)
    }
}

/** Un botón con ícono de la barra (los ToolbarItem con Image). */
@Composable
fun BotonBarraIcono(simbolo: String, etiqueta: String, alTocar: () -> Unit) {
    Box(Modifier.size(44.dp).clickable(onClick = alTocar).semantics { contentDescription = etiqueta }, contentAlignment = Alignment.Center) {
        Icono(simbolo, 19f, C.primarioTexto)
    }
}

/** El título que deja una pantalla para el «‹ volver» de la siguiente (su navigationTitle). */
private fun tituloDe(d: Destino?, p: Pestana): String? = when (d) {
    null -> when (p) { Pestana.inicio -> "Inicio"; Pestana.pedidos -> "Pedidos"; Pestana.productos -> "Productos"; Pestana.tincho -> null; Pestana.mas -> "Más" }
    is Destino.pedido -> "#${d.n}"
    Destino.cuenta -> "Cuenta"
    else -> null
}

/** La barra nativa de las pantallas de adentro, con la tipografía y los colores del handoff: back con chevron y el
 *  título del padre en `primary-text`, título 17/600 (Apariencia en iOS). */
@Composable
private fun BarraNavegacion(b: Barra, atras: String?, volver: () -> Unit) {
    if (b.oculta) return
    Box(Modifier.fillMaxWidth().background(C.fondo).statusBarsPadding().height(44.dp)) {
        Row(Modifier.align(Alignment.CenterStart).padding(start = 4.dp), verticalAlignment = Alignment.CenterVertically) {
            val izq = b.izquierda
            if (izq != null) izq() else {
                Row(Modifier.defaultMinSize(minHeight = 44.dp).clickable(onClick = volver).padding(horizontal = 4.dp).semantics { contentDescription = atras ?: "Atrás" },
                    verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(3.dp)) {
                    Icono("chevron.left", 19f, C.primarioTexto)
                    if (atras != null) Texto(atras, Estilo.body, C.primarioTexto, Modifier.widthIn(max = 120.dp), lineas = 1)
                }
            }
        }
        b.titulo?.let { Texto(it, Estilo.headline, C.texto, Modifier.align(Alignment.Center).widthIn(max = 200.dp), lineas = 1, alinear = TextAlign.Center) }
        Row(Modifier.align(Alignment.CenterEnd).padding(end = 8.dp), verticalAlignment = Alignment.CenterVertically) { b.acciones?.invoke(this) }
    }
}

/** El panel: 5 pestañas con su barra (la del handoff, no la del sistema), el «+» en Productos y los toasts
 *  arriba de la barra. */
@Composable
fun Principal() {
    val estado = LocalEstado.current
    val nav = remember { Navegacion() }
    val enRaiz = nav.caminoActual.isEmpty()

    CompositionLocalProvider(LocalNavegacion provides nav) {
        Box(Modifier.fillMaxSize().background(C.fondo)) {
            // Las cinco pilas viven a la vez (cada una guarda su camino) y se ve sólo la de la pestaña elegida.
            val visitadas = remember { mutableStateListOf(nav.pestana) }
            if (nav.pestana !in visitadas) visitadas.add(nav.pestana)
            Pestana.entries.filter { it in visitadas }.forEach { p ->
                key(p) {
                    val activa = nav.pestana == p
                    Box(Modifier.fillMaxSize().zIndex(if (activa) 1f else 0f).alpha(if (activa) 1f else 0f).then(if (activa) Modifier else Modifier.tragarToques())) {
                        Pila(nav, p)
                    }
                }
            }
            BackHandler(nav.caminoActual.isNotEmpty()) { nav.volver() }

            // En Inicio no va: «Vender ahora» ya está en la vidriera.
            AnimatedVisibility(enRaiz && nav.pestana == Pestana.productos,
                Modifier.align(Alignment.BottomEnd).zIndex(2f).navigationBarsPadding().padding(end = 20.dp, bottom = M.tabBar + 16.dp),
                enter = scaleIn() + fadeIn(), exit = scaleOut() + fadeOut()) {
                BotonMas { nav.crear = true }
            }

            AnimatedVisibility(estado.aviso != null, Modifier.align(Alignment.BottomCenter).zIndex(3f).navigationBarsPadding()
                .padding(start = 16.dp, end = 16.dp, bottom = if (enRaiz) M.tabBar + 17.dp else 84.dp),
                enter = slideInVertically(spring()) { it } + fadeIn(), exit = slideOutVertically(spring()) { it } + fadeOut()) {
                estado.aviso?.let { a -> VistaAviso(a) { estado.aviso = null } }
            }

            AnimatedVisibility(enRaiz, Modifier.align(Alignment.BottomCenter).zIndex(2f),
                enter = slideInVertically(tween(200)) { it }, exit = slideOutVertically(tween(200)) { it }) {
                BarraTabs(nav.pestana, estado.resumen?.insignias?.pedidos ?: 0) { nav.tocar(it) }
            }

            Hoja(nav.buscar, { nav.buscar = false }, radio = 28.dp) { Buscar() }
            Hoja(nav.crear, { nav.crear = false }, altoFijo = 420.dp) { HojaCrear() }
            // La venta en el local, a pantalla completa (fullScreenCover), con sus toasts.
            AnimatedVisibility(nav.venta, Modifier.zIndex(5f), enter = slideInVertically(tween(350)) { it }, exit = slideOutVertically(tween(300)) { it }) {
                Box(Modifier.fillMaxSize().background(C.fondo)) {
                    VentaLocal()
                    AnimatedVisibility(estado.aviso != null, Modifier.align(Alignment.BottomCenter).navigationBarsPadding().padding(start = 16.dp, end = 16.dp, bottom = 84.dp),
                        enter = slideInVertically(spring()) { it } + fadeIn(), exit = slideOutVertically(spring()) { it } + fadeOut()) {
                        estado.aviso?.let { a -> VistaAviso(a) { estado.aviso = null } }
                    }
                    BackHandler { nav.venta = false }
                }
            }
        }
    }
    // Lo que vive sólo en la web, en la hoja web del sistema (SFSafariViewController en iOS).
    LaunchedEffect(nav.web) { nav.web?.let { estado.abrirPestana(it); nav.web = null } }
    LaunchedEffect(estado.rutaPendiente) {
        estado.rutaPendiente?.let { nav.ir(it); estado.rutaPendiente = null }
    }
}

/** Una pila (NavigationStack): la raíz sin barra y las de adentro con la barra nativa. */
@Composable
private fun Pila(nav: Navegacion, p: Pestana) {
    val camino = nav.camino(p)
    AnimatedContent(camino.size to camino.lastOrNull(), transitionSpec = {
        if (targetState.first > initialState.first) slideInHorizontally { it } togetherWith slideOutHorizontally { -it / 3 }
        else slideInHorizontally { -it / 3 } togetherWith slideOutHorizontally { it }
    }, label = "pila") { (n, arriba) ->
        if (arriba == null) {
            // El Inicio lleva la vidriera hasta arriba de todo, debajo de la barra de estado (como en iOS).
            Box(Modifier.fillMaxSize().background(C.fondo).then(if (p == Pestana.inicio) Modifier else Modifier.statusBarsPadding())) {
                // La raíz deja lugar para la barra de pestañas (safeAreaInset).
                Box(Modifier.fillMaxSize().navigationBarsPadding().padding(bottom = M.tabBar)) {
                    when (p) {
                        Pestana.inicio -> Inicio()
                        Pestana.pedidos -> ListaPedidos()
                        Pestana.productos -> ListaProductos()
                        Pestana.tincho -> TinchoChat()
                        Pestana.mas -> Mas()
                    }
                }
            }
        } else {
            val barra = remember(n, arriba) { Barra() }
            val atras = tituloDe(camino.getOrNull(n - 2), p)
            // dismiss() en una pantalla apilada vuelve atrás; dentro de una hoja, la cierra (Hoja pisa esto).
            CompositionLocalProvider(LocalBarra provides barra, LocalCerrarHoja provides { nav.volver() }) {
                Column(Modifier.fillMaxSize().background(C.fondo)) {
                    BarraNavegacion(barra, atras) { nav.volver() }
                    Box(Modifier.weight(1f).fillMaxWidth().then(if (barra.oculta) Modifier.statusBarsPadding() else Modifier)) { Destinos(arriba) }
                }
            }
        }
    }
}

@Composable
private fun Destinos(d: Destino) {
    when (d) {
        is Destino.pedido -> DetallePedido(d.n)
        Destino.avisos -> Avisos()
        Destino.cuenta -> Cuenta()
        Destino.clientes -> Clientes()
        is Destino.cliente -> FichaCliente(d.c)
        Destino.caja -> CajaDelDia()
        Destino.carritos -> CarritosAbandonados()
        Destino.resenas -> ResenasYPreguntas()
        Destino.devoluciones -> Devoluciones()
        Destino.cupones -> Cupones()
        Destino.finanzas -> Finanzas()
        Destino.facturacion -> FacturacionArca()
        Destino.plan -> PlanCuenta()
        Destino.integraciones -> Integraciones()
        Destino.usuarios -> UsuariosYPermisos()
        is Destino.editorProducto -> EditorProducto(d.id)
        is Destino.nuevoProducto -> EditorProducto(null, sku = d.sku)
        Destino.pedidoManual -> PedidoManual()
        Destino.preferenciasAvisos -> PreferenciasAvisos()
        Destino.negocio -> NegocioCobro()
        Destino.links -> LinksPago()
        Destino.inventario -> EscanearInventario()
        Destino.ayuda -> CentroAyuda()
        Destino.miTienda -> MiTiendaEditor()
        Destino.estadisticas -> Estadisticas()
        Destino.compartir -> CompartirTienda()
        Destino.importar -> ImportarTienda()
        Destino.telegram -> TelegramTincho()
        Destino.medicion -> MedicionTienda()
        Destino.ia -> IAConectadas()
        is Destino.articuloAyuda -> ArticuloAyuda(d.slug)
    }
}

/** Una pestaña que no se ve no recibe toques (allowsHitTesting(false)). */
private fun Modifier.tragarToques(): Modifier = this.deshabilitado(true, 0f)

/** La barra de pestañas del handoff: 49 + safe area, `bg-elevated` al 92 %, borde de 0,5. */
@Composable
fun BarraTabs(pestana: Pestana, insignia: Int, tocar: (Pestana) -> Unit) {
    val fondo = tema(0xFBF7F2, 0x22140D, alfa = 0.92f).actual
    Column(Modifier.fillMaxWidth().background(fondo)) {
        Box(Modifier.fillMaxWidth().height(0.5.dp).background(C.linea2))
        Row(Modifier.fillMaxWidth().navigationBarsPadding().height(M.tabBar - 0.5.dp).padding(horizontal = 6.dp).padding(top = 6.dp), verticalAlignment = Alignment.Top) {
            Item(pestana, Pestana.inicio, "Inicio", "house", 0, tocar)
            Item(pestana, Pestana.pedidos, "Pedidos", "list.bullet.rectangle.portrait", insignia, tocar)
            Item(pestana, Pestana.productos, "Productos", "shippingbox", 0, tocar)
            Item(pestana, Pestana.tincho, "Tincho", null, 0, tocar)
            Item(pestana, Pestana.mas, "Más", "ellipsis", 0, tocar)
        }
    }
}

@Composable
private fun RowScope.Item(actual: Pestana, p: Pestana, texto: String, simbolo: String?, insignia: Int, tocar: (Pestana) -> Unit) {
    val activo = p == actual
    val color = if (activo) C.primarioTexto else C.texto3
    Column(
        Modifier.weight(1f).clickable(remember { MutableInteractionSource() }, null) { tocar(p); Haptico.seleccion() }
            .semantics { contentDescription = if (insignia > 0) "$texto, $insignia por preparar" else texto; selected = activo },
        horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(3.dp),
    ) {
        Box(Modifier.size(27.dp), contentAlignment = Alignment.Center) {
            if (simbolo != null) Icono(simbolo, 21f, color)
            else Box(Modifier.size(27.dp).clip(CircleShape).border(2.dp, if (activo) C.primarioTexto else Color.Transparent, CircleShape).padding(2.dp)) {
                ImagenTincho(Tincho.avatar, Modifier.size(23.dp).clip(CircleShape).background(C.crema), ContentScale.Crop)
            }
            if (insignia > 0) {
                val borde = tema(0xFBF7F2, 0x22140D).actual
                Box(Modifier.align(Alignment.TopStart).offset(x = 18.dp, y = (-5).dp).clip(RoundedCornerShape(50)).background(borde).padding(2.dp)) {
                    Box(Modifier.defaultMinSize(18.dp, 18.dp).clip(RoundedCornerShape(50)).background(C.error).padding(horizontal = 5.dp), contentAlignment = Alignment.Center) {
                        Texto("$insignia", Estilo(11.5f, Peso.bold), Color.White, numeros = true)
                    }
                }
            }
        }
        Texto(texto, Estilo.tab, color, lineas = 1)
    }
}

/** El «+» (58 Ø, gradiente, sombra). */
@Composable
fun BotonMas(accion: () -> Unit) {
    val brocha = C.primario
    Box(Modifier.size(58.dp).sombra(CircleShape, C.sombraPrimario.copy(alpha = 0.49f), 10.dp, y = 8.dp).clip(CircleShape).background(brocha)
        .clickable(onClick = accion).semantics { contentDescription = "Crear" }, contentAlignment = Alignment.Center) {
        Icono("plus", 22f, Color.White)
    }
}

/** 2e · La hoja «+»: Venta en el local, Producto con la cámara, Pedido manual. */
@Composable
fun HojaCrear() {
    val nav = LocalNavegacion.current
    val alcance = rememberCoroutineScope()
    fun cerrar() { nav.crear = false }

    /** Cierra la hoja y abre la pantalla en su pestaña. */
    fun ir(p: Pestana, d: Destino) {
        cerrar()
        alcance.launch {
            delay(300)
            nav.pestana = p
            nav.poner(p, listOf(d))
        }
    }

    Column(Modifier.fillMaxWidth().padding(horizontal = 20.dp), verticalArrangement = Arrangement.spacedBy(10.dp)) {
        Grabber()
        Row(Modifier.fillMaxWidth().padding(horizontal = 6.dp).padding(top = 6.dp), verticalAlignment = Alignment.CenterVertically) {
            Texto("Crear", Estilo.tituloHojaChico, C.texto, Modifier.weight(1f))
            Box(Modifier.size(30.dp).clip(CircleShape).background(tema(0x3A1D12, 0xFBF7F2, alfa = 0.08f, alfaOscuro = 0.1f).actual)
                .clickable { cerrar() }.semantics { contentDescription = "Cerrar" }, contentAlignment = Alignment.Center) {
                Icono("xmark", 12f, C.texto2)
            }
        }
        OpcionCrear("Venta en el local", "Escaneá, armá el ticket y cobrá", "barcode.viewfinder", true) {
            cerrar()
            alcance.launch { delay(350); nav.venta = true }
        }
        OpcionCrear("Producto con la cámara", "Sacale fotos y cargalo en un minuto", "camera", false) { ir(Pestana.productos, Destino.nuevoProducto(null)) }
        OpcionCrear("Pedido manual", "Para ventas por WhatsApp o teléfono", "doc.text", false) { ir(Pestana.pedidos, Destino.pedidoManual) }
        Spacer(Modifier.height(0.dp))
    }
}

@Composable
private fun OpcionCrear(titulo: String, texto: String, simbolo: String, principal: Boolean, accion: () -> Unit) {
    Tocable(accion, Modifier.fillMaxWidth(), hundir = false) {
        Row(Modifier.fillMaxWidth().tarjeta().padding(16.dp), horizontalArrangement = Arrangement.spacedBy(14.dp), verticalAlignment = Alignment.CenterVertically) {
            val forma = RoundedCornerShape(14.dp)
            Box(Modifier.size(48.dp).clip(forma).then(if (principal) Modifier.background(C.primario) else Modifier.background(C.crema)), contentAlignment = Alignment.Center) {
                Icono(simbolo, 21f, if (principal) Color.White else C.primarioTexto)
            }
            Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                Texto(titulo, Estilo.headline, C.texto)
                Texto(texto, Estilo.texto14, C.texto2)
            }
        }
    }
}
