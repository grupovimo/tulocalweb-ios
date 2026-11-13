package ar.com.tulocalweb.app.pantallas.clientes

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.IntrinsicSize
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.input.KeyboardType
import ar.com.tulocalweb.app.pantallas.mas.CampoTexto
import kotlinx.coroutines.launch
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.semantics.heading
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.app.Destino
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.app.LocalNavegacion
import ar.com.tulocalweb.app.nucleo.ErrorApi
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.PaginaPedidos
import ar.com.tulocalweb.app.nucleo.PedidoFila
import ar.com.tulocalweb.app.nucleo.Tincho
import ar.com.tulocalweb.app.nucleo.abrirUrl
import ar.com.tulocalweb.app.nucleo.errorDe
import ar.com.tulocalweb.app.nucleo.opcional
import ar.com.tulocalweb.app.pantallas.inicio.Indicador
import ar.com.tulocalweb.app.pantallas.pedidos.urlWhatsApp
import ar.com.tulocalweb.app.tema.*
import kotlinx.coroutines.delay
import kotlinx.serialization.Serializable

// Clon de Pantallas/Clientes/Clientes.swift.

/** Una fila de /v1/panel/listas/clientes. */
@Serializable
data class ClienteFila(
    val id: String,
    val nombre: String = "",
    val telefono: String? = null,
    val email: String? = null,
    val ciudad: String? = null,
    val gastado: Int = 0,
    val pedidos: Int = 0,
    /** Días desde la última compra. */
    val ultima: Double? = null,
    val seg: String? = null,
    val etiquetas: List<String>? = null,
)

@Serializable
data class PaginaClientes(val total: Int = 0, val filas: List<ClienteFila> = emptyList())

/** «hoy», «ayer», «hace 3 días». */
fun haceDias(d: Double?): String {
    if (d == null) return ""
    val n = kotlin.math.floor(d).toInt()
    return when { n < 1 -> "hoy"; n == 1 -> "ayer"; else -> "hace $n días" }
}

private val ordenesClientes = listOf("ltv" to "Más compras", "last" to "Compra más reciente", "orders" to "Más pedidos", "az" to "Nombre")

/** 12a · Clientes: buscador, orden y filas con avatar, pedidos, la última compra y el total. */
@Composable
fun Clientes() {
    val estado = LocalEstado.current
    val nav = LocalNavegacion.current
    var busqueda by remember { mutableStateOf("") }
    var orden by remember { mutableStateOf("ltv") }
    var carga by remember { mutableStateOf<Carga<PaginaClientes>>(Carga.cargando) }

    suspend fun cargar() {
        val q = buildList { add("orden" to orden); add("cuantos" to "100"); if (busqueda.isNotEmpty()) add("q" to busqueda) }
        try { carga = Carga.listo(estado.api.get("/v1/panel/listas/clientes", q)) } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
            if (carga.valor == null || e !is ErrorApi) carga = Carga.error(errorDe(e))
        }
    }

    Refrescable({ cargar() }) {
        LazyColumn(Modifier.fillMaxSize().background(C.fondo), contentPadding = PaddingValues(bottom = 24.dp)) {
            item { Texto("Clientes", Estilo.tituloGrande, C.texto, Modifier.margenPantalla().semantics { heading() }) }
            item { Box(Modifier.padding(top = 12.dp).margenPantalla()) { Buscador("Nombre, teléfono o ciudad", busqueda, { busqueda = it }) } }
            carga.valor?.let { d ->
                if (d.filas.isNotEmpty()) item {
                    Row(Modifier.padding(horizontal = margen + 4.dp).padding(top = 16.dp, bottom = 8.dp), verticalAlignment = Alignment.CenterVertically) {
                        Texto(Formato.cantidad(d.total, "cliente", "clientes"), Estilo.texto14, C.texto2, Modifier.weight(1f))
                        MenuBoton(ordenesClientes.map { (v, t) -> OpcionDeMenu(t, if (orden == v) "checkmark" else "") { orden = v } }, "Orden") {
                            Row(horizontalArrangement = Arrangement.spacedBy(4.dp), verticalAlignment = Alignment.CenterVertically) {
                                Texto(ordenesClientes.firstOrNull { it.first == orden }?.second ?: "", Estilo.texto14.con(Peso.semibold), C.primarioTexto)
                                Icono("chevron.down", 11f, C.primarioTexto)
                            }
                        }
                    }
                }
            }
            when (val c = carga) {
                Carga.cargando -> item { Box(Modifier.margenPantalla().padding(top = 24.dp)) { EsqueletoFilas(5, 42.dp) } }
                is Carga.error -> item { Box(Modifier.padding(top = 60.dp)) { EstadoError("los clientes", c.e) { estado.tarea { cargar() } } } }
                is Carga.listo -> if (c.v.filas.isEmpty()) item {
                    Box(Modifier.padding(top = 50.dp)) {
                        EstadoTincho(if (busqueda.isEmpty()) Tincho.saludo else Tincho.lupa,
                            if (busqueda.isEmpty()) "Todavía no tenés clientes" else "No encontramos «$busqueda»",
                            if (busqueda.isEmpty()) "Con la primera venta aparece acá, con todo lo que compró." else "Probá con otra parte del nombre o el teléfono.")
                    }
                } else item {
                    Column(Modifier.margenPantalla().fillMaxWidth().tarjeta()) {
                        c.v.filas.forEachIndexed { i, cl ->
                            if (i > 0) Separador(68.dp)
                            Box(Modifier.clickable { nav.abrir(Destino.cliente(cl)) }) { FilaCliente(cl) }
                        }
                    }
                }
            }
        }
    }
    LaunchedEffect(orden, busqueda) {
        if (carga.valor != null) delay(250)
        cargar()
    }
}

@Composable
fun FilaCliente(c: ClienteFila) {
    val (fondo, tinta) = coloresAvatar(c.nombre)
    Row(Modifier.fillMaxWidth().padding(horizontal = 14.dp, vertical = 12.dp), horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
        AvatarIniciales(c.nombre, 42.dp, fondo = fondo, tinta = tinta)
        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
            Texto(c.nombre, Estilo.fila, C.texto, lineas = 1)
            Texto("${Formato.cantidad(c.pedidos, "pedido", "pedidos")} · última ${haceDias(c.ultima)}", Estilo.footnote, C.texto2)
        }
        Texto(Formato.pesos(c.gastado), Estilo(15f, Peso.bold), C.texto, numeros = true)
    }
}

/** 5e · Ficha de cliente: avatar 76, nombre, antigüedad y etiquetas; WhatsApp / Llamar / Mail; 4 indicadores y el
 *  historial de pedidos. */
@Composable
fun FichaCliente(cliente: ClienteFila) {
    val estado = LocalEstado.current
    val nav = LocalNavegacion.current
    val ctx = LocalContext.current
    var pedidos by remember { mutableStateOf<List<PedidoFila>>(emptyList()) }
    var corregir by remember { mutableStateOf(false) }
    var borrar by remember { mutableStateOf(false) }
    val (fondo, tinta) = coloresAvatar(cliente.nombre)

    LaunchedEffect(Unit) {
        estado.api.opcional<PaginaPedidos>("/v1/panel/pedidos", listOf("q" to cliente.nombre, "cuantos" to "50"))?.let { l -> pedidos = l.filas.filter { it.cliente == cliente.nombre } }
    }

    @Composable
    fun Accion(simbolo: String, texto: String, url: String?, modifier: Modifier) {
        Column(modifier.heightIn(min = 58.dp).tarjeta(radio = 14.dp).deshabilitado(url == null, 0.45f).clickable { url?.let { abrirUrl(ctx, it) } },
            horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(3.dp, Alignment.CenterVertically)) {
            Icono(simbolo, 18f, C.primarioTexto)
            Texto(texto, Estilo.footnoteFuerte, C.primarioTinta)
        }
    }

    Column(Modifier.fillMaxSize().background(C.fondo).verticalScroll(rememberScrollState()).padding(top = 4.dp).margenPantalla().padding(bottom = 24.dp),
        verticalArrangement = Arrangement.spacedBy(14.dp)) {
        Column(Modifier.fillMaxWidth(), horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(6.dp)) {
            AvatarIniciales(cliente.nombre, 76.dp, fondo = fondo, tinta = tinta)
            Texto(cliente.nombre, Estilo(26f, Peso.bold, -0.02f), C.texto, alinear = TextAlign.Center)
            pedidos.lastOrNull()?.creado?.let { primero ->
                Texto("Cliente desde ${Formato.mesDe(primero)} de ${Formato.enArgentina(primero).year}", Estilo.texto14, C.texto2)
            }
            val etiquetas = listOfNotNull(if (cliente.pedidos >= 3) "Frecuente" else null, cliente.ciudad?.ifEmpty { null })
            if (etiquetas.isNotEmpty()) {
                Row(Modifier.padding(top = 2.dp), horizontalArrangement = Arrangement.spacedBy(6.dp)) {
                    etiquetas.forEachIndexed { i, e ->
                        val frecuente = i == 0 && cliente.pedidos >= 3
                        Texto(e, Estilo(12f, Peso.bold), if (frecuente) C.primarioTinta else C.neutroTinta,
                            Modifier.clip(Capsula).background(if (frecuente) C.crema else C.pista).padding(horizontal = 9.dp, vertical = 3.dp))
                    }
                }
            }
        }
        Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            Accion("message", "WhatsApp", urlWhatsApp(cliente.telefono), Modifier.weight(1f))
            Accion("phone", "Llamar", cliente.telefono?.let { t -> "tel:" + t.filter { it.isDigit() || it == '+' } }, Modifier.weight(1f))
            Accion("envelope", "Mail", cliente.email?.let { "mailto:$it" }, Modifier.weight(1f))
        }
        Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
            Row(Modifier.height(IntrinsicSize.Max), horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                Indicador("Total comprado", Formato.pesos(cliente.gastado), null, true, Modifier.weight(1f))
                Indicador("Pedidos", "${cliente.pedidos}", null, true, Modifier.weight(1f))
            }
            Row(Modifier.height(IntrinsicSize.Max), horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                Indicador("Ticket promedio", Formato.pesos(if (cliente.pedidos > 0) cliente.gastado / cliente.pedidos else 0), null, true, Modifier.weight(1f))
                Indicador("Última compra", haceDias(cliente.ultima), null, true, Modifier.weight(1f))
            }
        }
        if (pedidos.isNotEmpty()) {
            EtiquetaSeccion("Pedidos", Modifier.padding(horizontal = 4.dp).padding(top = 4.dp))
            Column(Modifier.fillMaxWidth().tarjeta()) {
                pedidos.forEachIndexed { i, p ->
                    if (i > 0) Separador()
                    Row(Modifier.fillMaxWidth().clickable { nav.abrir(Destino.pedido(p.numero)) }.padding(horizontal = 16.dp, vertical = 12.dp), verticalAlignment = Alignment.CenterVertically) {
                        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                            Texto("#${p.numero}", Estilo.fila, C.texto)
                            Texto(Formato.cuando(p.creado), Estilo.footnote, C.texto2)
                        }
                        Column(horizontalAlignment = Alignment.End, verticalArrangement = Arrangement.spacedBy(4.dp)) {
                            Texto(Formato.pesos(p.total), Estilo.fila, C.texto, numeros = true)
                            ChipEstado(p.estado, p.envio)
                        }
                    }
                }
            }
        }
        // Lo que pide la ley de datos personales: corregirlos o borrarlos cuando el cliente lo pide (como la web).
        EtiquetaSeccion("Sus datos", Modifier.padding(horizontal = 4.dp).padding(top = 4.dp))
        Column(Modifier.fillMaxWidth().tarjeta()) {
            FilaDatos("square.and.pencil", "Corregir", C.texto) { corregir = true }
            Separador()
            FilaDatos("trash", "Borrar sus datos", C.errorTinta) { borrar = true }
        }
    }
    Hoja(corregir, { corregir = false }) { HojaCorregirCliente(cliente) { nav.volver() } }
    Hoja(borrar, { borrar = false }) { HojaBorrarCliente(cliente) { nav.volver() } }
}

@Composable
private fun FilaDatos(simbolo: String, texto: String, color: Color, accion: () -> Unit) {
    Row(Modifier.fillMaxWidth().clickable(onClick = accion).padding(horizontal = 16.dp, vertical = 13.dp), horizontalArrangement = Arrangement.spacedBy(10.dp), verticalAlignment = Alignment.CenterVertically) {
        Icono(simbolo, 16f, color)
        Texto(texto, Estilo.body16, color, Modifier.weight(1f))
        Chevron(8f)
    }
}

/** POST /v1/panel/listas/clientes/acciones con un solo cliente (editar o suprimir). */
@Serializable private data class AccionCliente(val ids: List<String>, val accion: String, val valor: DatosCliente? = null)
@Serializable data class DatosCliente(val nombre: String, val email: String, val telefono: String)
@Serializable private data class Hechos(val hechos: Int = 0)

/** Lo que se escribió para corregir: null si está bien, o el mismo aviso que la web. */
fun errorDatosCliente(nombre: String, telefono: String): String? = when {
    nombre.trim().length < 2 -> "Poné el nombre del cliente."
    telefono.isNotBlank() && telefono.count { it.isDigit() } < 8 -> "El WhatsApp tiene que tener la característica: 11 5555-1234."
    else -> null
}

/** Corregir datos: se corrigen en su ficha y en sus pedidos (accion «editar»). */
@Composable
fun HojaCorregirCliente(cliente: ClienteFila, alTerminar: () -> Unit) {
    val estado = LocalEstado.current
    val cerrar = LocalCerrarHoja.current
    val alcance = rememberCoroutineScope()
    var nombre by remember { mutableStateOf(cliente.nombre) }
    var telefono by remember { mutableStateOf(cliente.telefono.orEmpty()) }
    var email by remember { mutableStateOf(cliente.email.orEmpty()) }
    var enviando by remember { mutableStateOf(false) }
    var error by remember { mutableStateOf<String?>(null) }

    suspend fun guardar() {
        error = errorDatosCliente(nombre, telefono)
        if (error != null) return
        enviando = true
        try {
            estado.api.enviar<Hechos, AccionCliente>("POST", "/v1/panel/listas/clientes/acciones",
                AccionCliente(listOf(cliente.id), "editar", DatosCliente(nombre.trim(), email.trim(), telefono.trim())))
            Haptico.exito()
            cerrar()
            estado.avisar("Datos corregidos en su ficha y en sus pedidos")
            estado.version += 1
            alTerminar()
        } catch (e: ErrorApi) {
            Haptico.error()
            error = e.detalle
        } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
            error = ErrorApi.desconocido.detalle
        } finally { enviando = false }
    }

    Column(Modifier.fillMaxSize().imePadding()) {
        Grabber()
        Column(Modifier.weight(1f).verticalScroll(rememberScrollState()).padding(horizontal = 20.dp).padding(top = 8.dp, bottom = 16.dp), verticalArrangement = Arrangement.spacedBy(14.dp)) {
            Box(Modifier.size(60.dp).clip(CircleShape).background(C.crema), contentAlignment = Alignment.Center) { Icono("square.and.pencil", 26f, C.primarioTinta) }
            Texto("Corregir datos", Estilo.tituloHojaChico, C.texto)
            Texto(Formato.sinViudas("Se corrigen en su ficha y en sus pedidos. Las facturas ya emitidas no cambian."), Estilo.texto14, C.texto2)
            CampoTexto("Nombre y apellido", nombre, { nombre = it.take(80) })
            CampoTexto("WhatsApp", telefono, { telefono = it.take(30) }, teclado = KeyboardType.Phone)
            CampoTexto("Email", email, { email = it.take(120) }, teclado = KeyboardType.Email)
            error?.let { MensajeError(it) }
        }
        Column(Modifier.padding(horizontal = 20.dp).padding(bottom = 12.dp), verticalArrangement = Arrangement.spacedBy(4.dp)) {
            BotonPrincipal("Guardar", { alcance.launch { guardar() } }, habilitado = !enviando, cargando = enviando)
            BotonTexto("Cancelar", { cerrar() }, alto = 44.dp)
        }
    }
}

/** Borrar sus datos (accion «suprimir»): la ficha y el contacto se borran, y sus pedidos quedan como venta a
 *  «Consumidor final». */
@Composable
fun HojaBorrarCliente(cliente: ClienteFila, alTerminar: () -> Unit) {
    val estado = LocalEstado.current
    val cerrar = LocalCerrarHoja.current
    val alcance = rememberCoroutineScope()
    var enviando by remember { mutableStateOf(false) }
    var error by remember { mutableStateOf<String?>(null) }

    suspend fun borrar() {
        enviando = true
        error = null
        try {
            estado.api.enviar<Hechos, AccionCliente>("POST", "/v1/panel/listas/clientes/acciones", AccionCliente(listOf(cliente.id), "suprimir"))
            Haptico.exito()
            cerrar()
            estado.avisar("Datos borrados: sus ventas quedan como «Consumidor final»")
            estado.version += 1
            alTerminar()
        } catch (e: ErrorApi) {
            Haptico.error()
            error = e.detalle
        } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
            error = ErrorApi.desconocido.detalle
        } finally { enviando = false }
    }

    Column(Modifier.fillMaxSize()) {
        Grabber()
        Column(Modifier.weight(1f).verticalScroll(rememberScrollState()).padding(horizontal = 20.dp).padding(top = 8.dp, bottom = 16.dp), verticalArrangement = Arrangement.spacedBy(14.dp)) {
            Box(Modifier.size(60.dp).clip(CircleShape).background(C.errorFondo), contentAlignment = Alignment.Center) { Icono("trash", 26f, C.errorTinta) }
            Texto(Formato.sinViudas("¿Borrar los datos de ${cliente.nombre}?"), Estilo.tituloHojaChico, C.texto)
            Texto(Formato.sinViudas("Es para cuando te lo pide. Borramos su ficha y su contacto, y sus pedidos quedan como venta a «Consumidor final». Las facturas emitidas se guardan como salieron, porque la ley lo pide. No se puede deshacer."),
                Estilo.texto14, C.texto)
            error?.let { MensajeError(it) }
        }
        Column(Modifier.padding(horizontal = 20.dp).padding(bottom = 12.dp), verticalArrangement = Arrangement.spacedBy(4.dp)) {
            BotonDestructivo(if (enviando) "Un momento…" else "Borrar sus datos", { alcance.launch { borrar() } }, habilitado = !enviando)
            BotonTexto("Cancelar", { cerrar() }, alto = 44.dp)
        }
    }
}

