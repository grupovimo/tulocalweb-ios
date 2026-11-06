package ar.com.tulocalweb.app.pantallas.mas

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ColumnScope
import androidx.compose.foundation.layout.IntrinsicSize
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
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
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.heading
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.app.Destino
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.app.LocalNavegacion
import ar.com.tulocalweb.app.nucleo.ErrorApi
import ar.com.tulocalweb.app.nucleo.FechaApi
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.Tincho
import ar.com.tulocalweb.app.nucleo.errorDe
import ar.com.tulocalweb.app.nucleo.intentar
import ar.com.tulocalweb.app.nucleo.opcional
import ar.com.tulocalweb.app.pantallas.inicio.Indicador
import ar.com.tulocalweb.app.tema.*
import kotlinx.coroutines.launch
import kotlinx.serialization.Serializable
import java.time.Duration
import java.time.Instant

// Clon de Pantallas/Mas/Secciones.swift: las secciones de «Más» con pantalla propia (antes abrían el panel web): Caja del
// día, Carritos abandonados, Reseñas y preguntas, Devoluciones, Cupones, Finanzas, Facturación de ARCA, Integraciones y
// Usuarios y permisos. Todas con las piezas del handoff y la API del panel.

// MARK: - Armazón común

/** Pantalla de sección: título grande, carga, error y «tirar para actualizar». `contenido` recibe los datos y cómo
 *  volver a pedirlos. */
@Composable
fun <T> PantallaSeccion(titulo: String, que: String, cargar: suspend () -> T, contenido: @Composable ColumnScope.(T, suspend () -> Unit) -> Unit) {
    val estado = LocalEstado.current
    var carga by remember { mutableStateOf<Carga<T>>(Carga.cargando) }
    suspend fun actualizar() {
        try { carga = Carga.listo(cargar()) } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
            if (carga.valor == null) carga = Carga.error(errorDe(e))
        }
    }
    Refrescable({ actualizar() }) {
        Column(Modifier.fillMaxSize().background(C.fondo).verticalScroll(rememberScrollState()).margenPantalla().padding(bottom = 24.dp).imePadding(),
            verticalArrangement = Arrangement.spacedBy(12.dp)) {
            Texto(titulo, Estilo.tituloGrande, C.texto, Modifier.semantics { heading() })
            when (val c = carga) {
                Carga.cargando -> Box(Modifier.padding(top = 12.dp)) { EsqueletoFilas(5, 42.dp) }
                is Carga.error -> Box(Modifier.padding(top = 60.dp)) { EstadoError(que, c.e) { estado.tarea { actualizar() } } }
                is Carga.listo -> contenido(c.v) { actualizar() }
            }
        }
    }
    LaunchedEffect(Unit) { actualizar() }
}

/** Grupo de filas en una tarjeta, con separadores. */
@Composable
fun <E> GrupoFilas(elementos: List<E>, id: (E) -> Any, inset: Dp = 16.dp, fila: @Composable (E) -> Unit) {
    Column(Modifier.fillMaxWidth().tarjeta()) {
        elementos.forEachIndexed { i, e ->
            if (i > 0) Separador(inset)
            androidx.compose.runtime.key(id(e)) { fila(e) }
        }
    }
}

/** Fila de dato: etiqueta a la izquierda, valor a la derecha. */
@Composable
fun FilaDato(etiqueta: String, valor: String, color: Color = C.texto) {
    Row(Modifier.fillMaxWidth().padding(horizontal = 16.dp, vertical = 13.dp), verticalAlignment = Alignment.CenterVertically) {
        Texto(etiqueta, Estilo.body16, C.texto2, Modifier.padding(end = 8.dp))
        Texto(valor, Estilo.fila, color, Modifier.weight(1f), alinear = TextAlign.End, numeros = true)
    }
}

/** Chip de estado chico con colores propios. */
@Composable
fun ChipSimple(texto: String, fondo: Color, tinta: Color) {
    Texto(texto, Estilo(12f, Peso.bold), tinta, Modifier.clip(Capsula).background(fondo).padding(horizontal = 9.dp, vertical = 3.dp), lineas = 1)
}

/** Grilla de indicadores de 2 columnas. */
@Composable
fun Indicadores(items: List<Pair<String, String>>) {
    Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
        items.chunked(2).forEach { par ->
            Row(Modifier.height(IntrinsicSize.Max), horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                par.forEach { Indicador(it.first, it.second, null, true, Modifier.weight(1f)) }
                if (par.size == 1) Spacer(Modifier.weight(1f))
            }
        }
    }
}

// MARK: - Caja del día

@Composable
fun CajaDelDia() {
    val estado = LocalEstado.current
    val nav = LocalNavegacion.current
    @Serializable data class Medio(val medio: String, val monto: Int)
    @Serializable data class Caja(val abierta: FechaApi? = null, val total: Int = 0, val medios: List<Medio> = emptyList())
    @Serializable data class Serie(val actual: List<Double> = emptyList(), val anterior: List<Double> = emptyList())
    @Serializable data class Tablero(val caja: Caja? = null, val serie: Serie? = null)

    PantallaSeccion("Caja del día", "la caja", { estado.api.get<Tablero>("/v1/panel/tablero", listOf("periodo" to "hoy")) }) { t, _ ->
        val caja = t.caja
        TarjetaHeroe("Vendido hoy en el local", Formato.pesos(caja?.total ?: 0),
            caja?.abierta?.let { "Caja abierta desde las ${Formato.horaDe(it)}" } ?: "La caja está cerrada", Tincho.caja)
        val s = t.serie
        if (s != null && s.actual.any { it > 0 }) {
            TarjetaSeccion("Cómo viene el día", "chart.xyaxis.line") { GraficoVentas(s.actual, s.anterior, alto = 140.dp) }
        }
        val medios = caja?.medios
        if (!medios.isNullOrEmpty()) {
            TarjetaSeccion("Por medio de pago", "creditcard") { DonaMedios(medios.map { ParteDona(nombreMedio(it.medio), it.monto) }) }
        }
        BotonPrincipal("Vender en el local", { nav.venta = true }, Modifier.padding(top = 4.dp), simbolo = "barcode.viewfinder")
    }
}

fun nombreMedio(medio: String): String = when (medio) {
    "efectivo" -> "Efectivo"
    "tarjeta" -> "Tarjeta"
    "mp" -> "Mercado Pago"
    "modo" -> "MODO"
    "transf" -> "Transferencia"
    else -> Formato.mayuscula(medio)
}

// MARK: - Carritos abandonados

@Composable
fun CarritosAbandonados() {
    val estado = LocalEstado.current
    @Serializable data class Carrito(val id: String, val nombre: String = "", val foto: String? = null, val desde: String = "", val creado: FechaApi = Instant.EPOCH,
                                     val cantidad: Int = 0, val total: Int = 0, val recordado: Boolean = false)
    @Serializable data class Datos(val carritos: List<Carrito> = emptyList(), val abiertos: Int = 0, val monto: Int = 0, val recuperado: Int = 0, val tasa: Double = 0.0)

    PantallaSeccion("Carritos abandonados", "los carritos", { estado.api.get<Datos>("/v1/panel/carritos") }) { d, actualizar ->
        TarjetaHeroe("En juego en carritos abiertos", Formato.pesos(d.monto), Formato.cantidad(d.abiertos, "carrito para recordar", "carritos para recordar"), Tincho.whatsapp)
        TarjetaSeccion("Lo que ya recuperaste", "arrow.uturn.backward.circle") {
            Row(verticalAlignment = Alignment.Bottom) {
                Texto(Formato.pesos(d.recuperado), Estilo.indicador, C.okTinta, Modifier.weight(1f), numeros = true)
                Texto("${Formato.porcentaje(if (d.tasa > 1) d.tasa else d.tasa * 100)} recuperado", Estilo.footnoteFuerte, C.texto2)
            }
            BarraProgreso(if (d.tasa > 1) d.tasa / 100 else d.tasa)
        }
        if (d.carritos.isEmpty()) {
            Box(Modifier.padding(top = 30.dp)) { EstadoTincho(Tincho.durmiendo, "No hay carritos abandonados", "Cuando alguien deja productos sin comprar, aparece acá para que le recuerdes.") }
        } else {
            GrupoFilas(d.carritos, { it.id }, 70.dp) { c ->
                Row(Modifier.fillMaxWidth().padding(horizontal = 14.dp, vertical = 12.dp), horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
                    Miniatura(c.foto, 44.dp)
                    Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                        Texto(c.nombre, Estilo.fila, C.texto, lineas = 1)
                        Texto("${Formato.cantidad(c.cantidad, "producto", "productos")} · ${Formato.hace(c.creado)} · ${c.desde}", Estilo.footnote, C.texto2, lineas = 1)
                    }
                    Column(horizontalAlignment = Alignment.End, verticalArrangement = Arrangement.spacedBy(6.dp)) {
                        Texto(Formato.pesos(c.total), Estilo.fila, C.texto, numeros = true)
                        if (c.recordado) ChipSimple("Recordado", C.okFondo, C.okTinta)
                        else BotonTonal("Recordar", {
                            estado.tarea {
                                try {
                                    estado.api.hacer("POST", "/v1/panel/carritos/${c.id}/recordar")
                                    estado.avisar("Le mandamos el recordatorio")
                                    actualizar()
                                } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) { estado.avisar("No pudimos mandar el recordatorio", ok = false) }
                            }
                        }, alto = 28.dp)
                    }
                }
            }
        }
    }
}

// MARK: - Reseñas y preguntas

@Serializable data class Pregunta(val id: String, val producto: String = "", val foto: String? = null, val pregunta: String = "", val creada: FechaApi = Instant.EPOCH,
                                  val sugerida: String = "", val respuesta: String? = null)
@Serializable data class Resena(val id: String, val autor: String = "", val estrellas: Int = 0, val texto: String = "", val origen: String = "", val creada: FechaApi = Instant.EPOCH,
                                val producto: String? = null, val conCompra: Boolean = false, val oculta: Boolean = false, val respuesta: String? = null)

@Composable
fun ResenasYPreguntas() {
    val estado = LocalEstado.current
    var vista by remember { mutableStateOf("preguntas") }
    var respondiendo by remember { mutableStateOf<Pregunta?>(null) }
    var respondiendoResena by remember { mutableStateOf<Resena?>(null) }
    var actualizarHoja by remember { mutableStateOf<(suspend () -> Unit)?>(null) }
    @Serializable data class Datos(val preguntas: List<Pregunta> = emptyList(), val resenas: List<Resena> = emptyList(), val promedio: Double? = null, val total: Int = 0, val porEstrellas: List<Int>? = null)
    @Serializable data class C2(val accion: String)

    PantallaSeccion("Reseñas y preguntas", "las reseñas", { estado.api.get<Datos>("/v1/panel/resenas") }) { d, actualizar ->
        actualizarHoja = actualizar
        Segmentado(listOf("preguntas" to "Preguntas", "resenas" to "Reseñas"), vista, { vista = it })
        if (vista == "preguntas") {
            if (d.preguntas.isEmpty()) {
                Box(Modifier.padding(top = 30.dp)) { EstadoTincho(Tincho.saludo, "No hay preguntas", "Cuando te pregunten algo de un producto, lo respondés acá.") }
            } else GrupoFilas(d.preguntas, { it.id }) { p ->
                Column(Modifier.fillMaxWidth().clickable { if (p.respuesta == null) respondiendo = p }.padding(14.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
                    Row {
                        Texto(p.producto, Estilo.footnoteFuerte, C.texto2, Modifier.weight(1f), lineas = 1)
                        Texto(Formato.hace(p.creada), Estilo.footnote, C.texto3)
                    }
                    Texto(p.pregunta, Estilo.body16, C.texto)
                    val r = p.respuesta
                    if (r != null) Texto(r, Estilo.texto14, C.texto2, Modifier.fillMaxWidth().clip(RoundedCornerShape(12.dp)).background(C.crema2).padding(10.dp))
                    else Texto("Responder", Estilo.subheadFuerte, C.primarioTexto)
                }
            }
        } else {
            d.promedio?.let { prom ->
                TarjetaSeccion("Lo que dicen de tu tienda", "star.bubble") {
                    Row(horizontalArrangement = Arrangement.spacedBy(18.dp), verticalAlignment = Alignment.CenterVertically) {
                        Column(horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(4.dp)) {
                            Texto(String.format(java.util.Locale.US, "%.1f", prom).replace(".", ","), Estilo.montoPantalla, C.texto)
                            Estrellas(Math.round(prom).toInt())
                            Texto(Formato.cantidad(d.total, "reseña", "reseñas"), Estilo.footnote, C.texto2)
                        }
                        Box(Modifier.weight(1f)) { BarrasEstrellas(d.porEstrellas ?: emptyList()) }
                    }
                }
            }
            if (d.resenas.isEmpty()) {
                Box(Modifier.padding(top = 30.dp)) { EstadoTincho(Tincho.celebrando, "Todavía no hay reseñas", "Las de la tienda, Mercado Libre y Google aparecen acá.") }
            } else GrupoFilas(d.resenas, { it.id }) { r ->
                Column(Modifier.fillMaxWidth().padding(14.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
                    Row(horizontalArrangement = Arrangement.spacedBy(6.dp), verticalAlignment = Alignment.CenterVertically) {
                        Estrellas(r.estrellas)
                        Texto(r.autor, Estilo.footnoteFuerte, C.texto, Modifier.weight(1f))
                        if (r.oculta) ChipSimple("Oculta", C.pista, C.texto2)
                    }
                    Texto(r.texto, Estilo.body16, C.texto)
                    Texto(listOfNotNull(r.producto, r.origen, Formato.dia(r.creada), if (r.conCompra) "compró" else null).joinToString(" · "), Estilo.footnote, C.texto3)
                    r.respuesta?.let { Texto(it, Estilo.texto14, C.texto2, Modifier.fillMaxWidth().clip(RoundedCornerShape(12.dp)).background(C.crema2).padding(10.dp)) }
                    Row(horizontalArrangement = Arrangement.spacedBy(16.dp)) {
                        if (r.respuesta == null) Texto("Responder", Estilo.subheadFuerte, C.primarioTexto, Modifier.clickable { respondiendoResena = r })
                        Texto(if (r.oculta) "Mostrar" else "Ocultar", Estilo.subheadFuerte, C.primarioTexto, Modifier.clickable {
                            estado.tarea { intentar { estado.api.hacer("POST", "/v1/panel/resenas/${r.id}", C2(if (r.oculta) "mostrar" else "ocultar")) }; actualizar() }
                        })
                    }
                }
            }
        }
    }
    @Serializable data class R1(val respuesta: String)
    @Serializable data class R2(val accion: String, val respuesta: String)
    Hoja(respondiendo != null, { respondiendo = null }, listOf(Altura.media, Altura.grande)) {
        respondiendo?.let { p -> HojaRespuesta(p.pregunta, p.sugerida) { t -> estado.api.hacer("POST", "/v1/panel/preguntas/${p.id}", R1(t)); actualizarHoja?.invoke() } }
    }
    Hoja(respondiendoResena != null, { respondiendoResena = null }, listOf(Altura.media, Altura.grande)) {
        respondiendoResena?.let { r -> HojaRespuesta(r.texto, "") { t -> estado.api.hacer("POST", "/v1/panel/resenas/${r.id}", R2("responder", t)); actualizarHoja?.invoke() } }
    }
}

@Composable
fun Estrellas(n: Int) {
    Row(Modifier.semantics { contentDescription = "$n de 5 estrellas" }, horizontalArrangement = Arrangement.spacedBy(1.dp)) {
        repeat(5) { i -> Icono(if (i < n) "star.fill" else "star", 11f, C.aviso) }
    }
}

/** Hoja para responder una pregunta o una reseña (con la sugerencia de Hornero precargada). */
@Composable
fun HojaRespuesta(titulo: String, sugerida: String, enviar: suspend (String) -> Unit) {
    val estado = LocalEstado.current
    val cerrar = LocalCerrarHoja.current
    val alcance = rememberCoroutineScope()
    var texto by remember { mutableStateOf(sugerida) }
    var enviando by remember { mutableStateOf(false) }
    Column(Modifier.fillMaxSize().imePadding().padding(horizontal = 20.dp), verticalArrangement = Arrangement.spacedBy(14.dp)) {
        Grabber()
        Texto(titulo, Estilo.tituloHojaChico, C.texto, lineas = 3)
        EditorTexto(texto, { texto = it }, 140.dp)
        BotonPrincipal(if (enviando) "Publicando…" else "Publicar respuesta", {
            alcance.launch {
                enviando = true
                try { enviar(texto.trim()); estado.avisar("Respuesta publicada"); cerrar() } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
                    estado.avisar("No pudimos publicarla", ok = false)
                }
                enviando = false
            }
        }, habilitado = texto.trim().isNotEmpty() && !enviando)
    }
}

// MARK: - Devoluciones

@Composable
fun Devoluciones() {
    val estado = LocalEstado.current
    @Serializable data class Devolucion(val id: String, val producto: String = "", val foto: String? = null, val cliente: String = "", val pedido: Int = 0,
                                        val creada: FechaApi = Instant.EPOCH, val motivo: String = "", val monto: Int = 0, val resolucion: String = "", val estado: String = "")
    @Serializable data class Datos(val devoluciones: List<Devolucion> = emptyList(), val tasa: Double = 0.0)
    @Serializable data class C2(val accion: String)

    suspend fun accion(v: Devolucion, a: String, actualizar: suspend () -> Unit) {
        try {
            estado.api.hacer("POST", "/v1/panel/devoluciones/${v.id}", C2(a))
            estado.avisar(if (a == "aprobar") "Devolución aprobada" else "Le ofrecimos el cambio") {
                estado.tarea { intentar { estado.api.hacer("POST", "/v1/panel/devoluciones/${v.id}", C2("deshacer")) } }
            }
            actualizar()
        } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) { estado.avisar("No pudimos guardarlo", ok = false) }
    }

    PantallaSeccion("Devoluciones", "las devoluciones", { estado.api.get<Datos>("/v1/panel/devoluciones") }) { d, actualizar ->
        if (d.devoluciones.isEmpty()) {
            Box(Modifier.padding(top = 30.dp)) { EstadoTincho(Tincho.exito, "No hay devoluciones", "Cuando alguien use el botón de arrepentimiento, lo resolvés acá.") }
        } else {
            val pendientes = d.devoluciones.filter { it.estado == "Pendiente" }
            TarjetaHeroe("Devoluciones para resolver", "${pendientes.size}",
                "${Formato.pesos(pendientes.sumOf { it.monto })} en juego · ${Formato.porcentaje(if (d.tasa > 1) d.tasa else d.tasa * 100)} de los pedidos", Tincho.envio)
            EtiquetaSeccion("Pedidos de devolución", Modifier.padding(horizontal = 4.dp).padding(top = 4.dp))
            GrupoFilas(d.devoluciones, { it.id }) { v ->
                Column(Modifier.fillMaxWidth().padding(14.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                    Row(horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
                        Miniatura(v.foto, 44.dp)
                        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                            Texto(v.producto, Estilo.fila, C.texto, lineas = 1)
                            Texto("${v.cliente} · pedido #${v.pedido} · ${Formato.dia(v.creada)}", Estilo.footnote, C.texto2, lineas = 1)
                        }
                        Texto(Formato.pesos(v.monto), Estilo.fila, C.texto, numeros = true)
                    }
                    Texto(v.motivo, Estilo.texto14, C.texto2)
                    if (v.estado == "Pendiente") {
                        Row(horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                            BotonTonal("Aprobar", { estado.tarea { accion(v, "aprobar", actualizar) } })
                            BotonSecundario("Ofrecer cambio", { estado.tarea { accion(v, "cambio", actualizar) } }, alto = 34.dp)
                        }
                    } else {
                        ChipSimple(v.estado, if (v.estado == "Aprobada") C.okFondo else C.pista, if (v.estado == "Aprobada") C.okTinta else C.texto2)
                    }
                }
            }
        }
    }
}

// MARK: - Cupones

@Serializable data class CuponFila(val id: String, val codigo: String = "", val tipo: String = "", val valor: Int = 0, val descripcion: String = "", val estado: String = "",
                                   val usos: Int = 0, val limite: Int = 0, val vendio: Int = 0)
@Serializable data class DatosCupones(val cupones: List<CuponFila> = emptyList(), val kpis: Kpis = Kpis()) {
    @Serializable data class Kpis(val activos: Int = 0, val usosMes: Int = 0, val vendido: Int = 0, val descontado: Int = 0)
}

fun resumenCupon(c: CuponFila): String = when (c.tipo) {
    "pct" -> "${c.valor} % de descuento"
    "fijo" -> "${Formato.pesos(c.valor)} de descuento"
    "envio" -> "Envío gratis"
    else -> "Regalo"
}

@Composable
fun Cupones() {
    val estado = LocalEstado.current
    var nuevo by remember { mutableStateOf(false) }
    var actualizarHoja by remember { mutableStateOf<(suspend () -> Unit)?>(null) }
    @Serializable data class C2(val accion: String)

    PantallaSeccion("Cupones", "los cupones", { estado.api.get<DatosCupones>("/v1/panel/cupones") }) { d, actualizar ->
        actualizarHoja = actualizar
        BotonPrincipal("Crear cupón", { nuevo = true }, alto = 48.dp, simbolo = "plus")
        TarjetaHeroe("Vendido con cupones este mes", Formato.pesos(d.kpis.vendido), "${Formato.cantidad(d.kpis.usosMes, "uso", "usos")} · ${Formato.pesos(d.kpis.descontado)} de descuento", Tincho.ticket)
        GrillaIndicadores(listOf(
            IndicadorIcono("ticket.fill", "Cupones activos", "${d.kpis.activos}"),
            IndicadorIcono("percent", "Descontado", Formato.pesos(d.kpis.descontado), tono = T.okFondo to T.okTinta),
        ))
        if (d.cupones.isEmpty()) {
            Box(Modifier.padding(top = 30.dp)) { EstadoTincho(Tincho.ticket, "Todavía no tenés cupones", "Un descuento con código para tus clientes: lo creás en un minuto.") }
        } else GrupoFilas(d.cupones, { it.id }) { c ->
            Row(Modifier.fillMaxWidth().padding(14.dp), horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
                Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(3.dp)) {
                    Texto(c.codigo, Estilo(16f, Peso.bold), C.texto)
                    Texto(c.descripcion.ifEmpty { resumenCupon(c) }, Estilo.footnote, C.texto2, lineas = 2)
                    Texto("${Formato.cantidad(c.usos, "uso", "usos")}${if (c.limite > 0) " de ${c.limite}" else ""} · vendió ${Formato.pesos(c.vendio)}", Estilo.footnote, C.texto3)
                }
                if (c.estado == "activo" || c.estado == "pausado") {
                    Llave(c.estado == "activo", { on -> estado.tarea { intentar { estado.api.hacer("POST", "/v1/panel/cupones/${c.id}", C2(if (on) "activar" else "pausar")) }; actualizar() } })
                } else ChipSimple(Formato.mayuscula(c.estado), C.pista, C.texto2)
            }
        }
    }
    Hoja(nuevo, { nuevo = false }) { HojaCuponNuevo { actualizarHoja?.invoke() } }
}

// MARK: - Finanzas

@Composable
fun Finanzas() {
    val estado = LocalEstado.current
    @Serializable data class Kpis(val aLiquidar: Int? = null, val aLiquidarDetalle: String? = null, val gastos: Int? = null, val gastosDetalle: String? = null,
                                  val ganancia: Int? = null, val gananciaDetalle: String? = null, val margen: Int? = null, val mes: String = "")
    @Serializable data class Movimiento(val id: String, val tipo: String = "", val titulo: String = "", val detalle: String = "", val monto: Int = 0)
    @Serializable data class Flujo(val dia: Int, val entra: Int = 0, val sale: Int = 0)
    @Serializable data class Caja(val ventas: Int = 0, val gastos: Int = 0, val meta: String = "", val metaPct: Int = 0, val falta: Int = 0)
    @Serializable data class Datos(val vendido30: Int? = null, val disponible: Int? = null, val cuenta: String? = null, val kpis: Kpis = Kpis(), val flujo: List<Flujo>? = null,
                                   val caja: Caja? = null, val movimientos: List<Movimiento> = emptyList())

    PantallaSeccion("Finanzas", "las finanzas", { estado.api.get<Datos>("/v1/panel/finanzas") }) { d, _ ->
        TarjetaHeroe(if (d.disponible != null) "Disponible en Mercado Pago" else "Vendido en los últimos 30 días", Formato.pesos(d.disponible ?: d.vendido30 ?: 0),
            d.cuenta ?: d.vendido30?.let { if (d.disponible != null) "Vendiste ${Formato.pesos(it)} en 30 días" else "" }, Tincho.grafico)
        EtiquetaSeccion(Formato.mayuscula(d.kpis.mes), Modifier.padding(horizontal = 4.dp).padding(top = 4.dp))
        GrillaIndicadores(listOf(
            IndicadorIcono("clock.arrow.circlepath", "A liquidar", d.kpis.aLiquidar?.let { Formato.pesos(it) } ?: "—", d.kpis.aLiquidarDetalle),
            IndicadorIcono("arrow.down.right", "Gastos", d.kpis.gastos?.let { Formato.pesos(it) } ?: "—", d.kpis.gastosDetalle, T.errorFondo to T.errorTinta),
            IndicadorIcono("arrow.up.right", "Ganancia", d.kpis.ganancia?.let { Formato.pesos(it) } ?: "—", d.kpis.gananciaDetalle, T.okFondo to T.okTinta),
            IndicadorIcono("percent", "Margen", d.kpis.margen?.let { "$it %" } ?: "—", tono = T.avisoFondo to T.avisoTinta),
        ))
        val c = d.caja
        if (c != null && (c.metaPct > 0 || c.meta.isNotEmpty())) {
            TarjetaSeccion("Meta del mes", "flag.checkered") {
                Row(horizontalArrangement = Arrangement.spacedBy(16.dp), verticalAlignment = Alignment.CenterVertically) {
                    AnilloAvance(c.metaPct / 100.0)
                    Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(4.dp)) {
                        Texto(c.meta, Estilo.fila, C.texto, lineas = 2)
                        if (c.falta > 0) Texto("Faltan ${Formato.pesos(c.falta)}", Estilo.footnote, C.texto2, numeros = true)
                        else Texto("¡Llegaste a la meta!", Estilo.footnoteFuerte, C.okTinta)
                    }
                }
            }
        }
        val f = d.flujo
        if (f != null && f.any { it.entra > 0 || it.sale > 0 }) {
            val flujo = f.map { DiaFlujo(it.dia, it.entra, it.sale) }
            TarjetaSeccion("Lo que entra y sale (próximos 14 días)", "chart.bar.xaxis") {
                GraficoFlujo(flujo)
                Row(horizontalArrangement = Arrangement.spacedBy(14.dp)) {
                    Row(horizontalArrangement = Arrangement.spacedBy(4.dp), verticalAlignment = Alignment.CenterVertically) { Icono("circle.fill", 9f, C.ok); Texto("Entra", Estilo(11f, Peso.semibold), C.ok) }
                    Row(horizontalArrangement = Arrangement.spacedBy(4.dp), verticalAlignment = Alignment.CenterVertically) { Icono("circle.fill", 9f, C.error); Texto("Sale", Estilo(11f, Peso.semibold), C.error) }
                }
            }
            val neto = flujo.sumOf { it.entra - it.sale }
            TarjetaSeccion("Cómo queda la caja", "chart.line.uptrend.xyaxis") {
                GraficoSaldo(flujo)
                Texto(Formato.sinViudas(if (neto >= 0) "Si entra y sale lo previsto, en 14 días sumás ${Formato.pesos(neto)}."
                    else "Si entra y sale lo previsto, en 14 días salen ${Formato.pesos(-neto)} más de lo que entra."), Estilo.footnote, if (neto >= 0) C.okTinta else C.errorTinta)
            }
        }
        if (d.movimientos.isNotEmpty()) {
            EtiquetaSeccion("Movimientos", Modifier.padding(horizontal = 4.dp).padding(top = 8.dp))
            GrupoFilas(d.movimientos, { it.id }, 62.dp) { m ->
                val neg = m.monto < 0
                Row(Modifier.fillMaxWidth().padding(horizontal = 14.dp, vertical = 12.dp), horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
                    Box(Modifier.size(34.dp).clip(RoundedCornerShape(10.dp)).background(if (neg) C.errorFondo else C.okFondo), contentAlignment = Alignment.Center) {
                        Icono(simboloMovimiento(m.tipo), 14f, if (neg) C.errorTinta else C.okTinta)
                    }
                    Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                        Texto(m.titulo, Estilo.fila, C.texto, lineas = 1)
                        Texto(m.detalle, Estilo.footnote, C.texto2, lineas = 1)
                    }
                    Texto((if (neg) "−" else "+") + Formato.pesos(kotlin.math.abs(m.monto)), Estilo.fila, if (neg) C.errorTinta else C.okTinta, numeros = true)
                }
            }
        }
    }
}

// MARK: - Facturación de ARCA

@Composable
fun FacturacionArca() {
    val estado = LocalEstado.current
    var alta by remember { mutableStateOf(false) }
    var actualizarHoja by remember { mutableStateOf<(suspend () -> Unit)?>(null) }
    @Serializable data class Datos(val conectada: Boolean = false, val cuit: String = "", val razonSocial: String = "", val condicion: String = "", val puntoVenta: Int = 0,
                                   val ambiente: String = "", val certificado: Boolean = false, val certificadoVence: FechaApi? = null, val facturarAlCobrar: Boolean = false)
    @Serializable data class C2(val facturarAlCobrar: Boolean)

    PantallaSeccion("Facturación de ARCA", "la facturación", { estado.api.get<Datos>("/v1/panel/arca") }) { d, actualizar ->
        actualizarHoja = actualizar
        if (!d.conectada) {
            Box(Modifier.padding(top = 20.dp)) {
                EstadoTincho(Tincho.portapapeles, "Todavía no conectaste ARCA", "Para facturar hace falta el certificado de ARCA. Lo cargás una sola vez, en tres pasos.",
                    accion = "Conectar ARCA", principal = true, alTocar = { alta = true })
            }
        } else {
            TarjetaHeroe("ARCA conectada", d.razonSocial.ifEmpty { d.cuit }, "CUIT ${d.cuit} · punto de venta ${d.puntoVenta}", Tincho.portapapeles)
            d.certificadoVence?.let { vence ->
                val dias = Duration.between(Instant.now(), vence).toDays().toInt()
                TarjetaSeccion("Certificado", "checkmark.seal") {
                    Row(horizontalArrangement = Arrangement.spacedBy(6.dp), verticalAlignment = Alignment.Bottom) {
                        Texto(if (dias > 0) Formato.cantidad(dias, "día", "días") else "Vencido", Estilo.indicador, if (dias < 30) C.errorTinta else C.texto)
                        Texto(if (dias > 0) "hasta que venza" else "renovalo para seguir facturando", Estilo.footnote, C.texto2, Modifier.padding(bottom = 3.dp))
                    }
                    BarraProgreso(maxOf(0, minOf(730, dias)) / 730.0)
                }
            }
            GrupoFilas(listOf(
                "Razón social" to d.razonSocial, "CUIT" to d.cuit, "Condición" to d.condicion, "Punto de venta" to "${d.puntoVenta}",
                "Ambiente" to (if (d.ambiente == "produccion") "Producción" else "Homologación"),
                "Certificado" to (d.certificadoVence?.let { "vence el ${Formato.dia(it)}" } ?: if (d.certificado) "cargado" else "falta"),
            ), { it.first }) { f -> FilaDato(f.first, f.second) }
            Row(Modifier.fillMaxWidth().tarjeta().padding(horizontal = 16.dp, vertical = 12.dp), verticalAlignment = Alignment.CenterVertically) {
                Texto("Facturar al cobrar", Estilo.body16, C.texto, Modifier.weight(1f))
                Llave(d.facturarAlCobrar, { on ->
                    estado.tarea {
                        try { estado.api.hacer("PUT", "/v1/panel/arca/ajustes", C2(on)); actualizar() } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
                            estado.avisar("No pudimos guardarlo", ok = false)
                        }
                    }
                })
            }
            BotonSecundario("Renovar el certificado", { alta = true }, lleno = true)
        }
    }
    Hoja(alta, { alta = false }) { HojaArca { actualizarHoja?.invoke() } }
}

// El plan (PlanCuenta) está en Plan.kt.

// MARK: - Integraciones

/** Las que tienen su pantalla: Telegram, medición y las IA. */
private val otrasIntegraciones = listOf(
    Triple(Destino.telegram, "Tincho en Telegram", "Avisos de ventas y consultas por chat") to "paperplane",
    Triple(Destino.medicion, "Google y Meta", "Analytics, píxel y catálogo de productos") to "chart.xyaxis.line",
    Triple(Destino.ia, "Tu IA (Claude, ChatGPT)", "Manejá la tienda desde tu asistente") to "sparkles",
)

@Composable
fun Integraciones() {
    val estado = LocalEstado.current
    val nav = LocalNavegacion.current
    var correo by remember { mutableStateOf(false) }
    var conectando by remember { mutableStateOf<String?>(null) }
    var actualizarHoja by remember { mutableStateOf<(suspend () -> Unit)?>(null) }
    // Todo opcional: sin cuenta conectada la API no manda usuario, apodo ni mail.
    @Serializable data class MP(val conectada: Boolean? = null, val usuario: String? = null, val reautorizar: Boolean? = null)
    @Serializable data class ML(val conectada: Boolean? = null, val apodo: String? = null, val reautorizar: Boolean? = null)
    @Serializable data class Correo(val conectada: Boolean? = null, val email: String? = null)
    data class Datos(val mp: MP?, val ml: ML?, val correo: Correo?)
    data class FilaIntegracion(val id: String, val nombre: String, val detalle: String, val conectada: Boolean, val alerta: Boolean)

    PantallaSeccion("Integraciones", "las integraciones", {
        // Uno detrás del otro (en iOS `async let` cerraba la app en Release).
        Datos(estado.api.opcional("/v1/panel/integraciones/mercadopago"), estado.api.opcional("/v1/panel/integraciones/mercadolibre"), estado.api.opcional("/v1/panel/integraciones/correo"))
    }) { d, actualizar ->
        actualizarHoja = actualizar
        val mp = d.mp?.conectada == true; val ml = d.ml?.conectada == true; val co = d.correo?.conectada == true
        val filas = listOf(
            FilaIntegracion("mp", "Mercado Pago", if (mp) (d.mp?.usuario ?: "Conectada") else "Cobrá con tarjeta, QR y link", mp, d.mp?.reautorizar == true),
            FilaIntegracion("ml", "Mercado Libre", if (ml) (d.ml?.apodo ?: "Conectada") else "Publicá y sincronizá el stock", ml, d.ml?.reautorizar == true),
            FilaIntegracion("correo", "Correo Argentino", if (co) (d.correo?.email ?: "Conectada") else "Etiquetas y seguimiento de envíos", co, false),
        )
        GrupoFilas(filas, { it.id }) { f ->
            Row(Modifier.fillMaxWidth().padding(horizontal = 16.dp, vertical = 12.dp), verticalAlignment = Alignment.CenterVertically) {
                Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                    Texto(f.nombre, Estilo.fila, C.texto)
                    Texto(f.detalle, Estilo.footnote, C.texto2, lineas = 1)
                }
                when {
                    f.conectada && !f.alerta -> ChipSimple("Conectada", C.okFondo, C.okTinta)
                    conectando == f.id -> Rueda(C.acento)
                    else -> BotonTonal(if (f.alerta) "Volver a conectar" else "Conectar", {
                        if (f.id == "correo") { correo = true } else estado.tarea {
                            conectando = f.id
                            val ok = ConectarCuenta.conectar(if (f.id == "mp") "mercadopago" else "mercadolibre", estado.api)
                            conectando = null
                            estado.avisar(if (ok) "${f.nombre} conectado" else "No se completó la conexión", ok = ok)
                            if (ok) estado.version += 1
                            actualizar()
                        }
                    })
                }
            }
        }
        EtiquetaSeccion("Más integraciones", Modifier.padding(horizontal = 4.dp).padding(top = 8.dp))
        Column(Modifier.fillMaxWidth().tarjeta()) {
            otrasIntegraciones.forEachIndexed { i, (o, simbolo) ->
                if (i > 0) Separador()
                Row(Modifier.fillMaxWidth().clickable { nav.abrir(o.first) }.padding(horizontal = 16.dp, vertical = 12.dp), horizontalArrangement = Arrangement.spacedBy(12.dp),
                    verticalAlignment = Alignment.CenterVertically) {
                    Box(Modifier.size(34.dp).clip(RoundedCornerShape(10.dp)).background(C.crema), contentAlignment = Alignment.Center) { Icono(simbolo, 15f, C.primarioTexto) }
                    Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                        Texto(o.second, Estilo.fila, C.texto)
                        Texto(o.third, Estilo.footnote, C.texto2, lineas = 1)
                    }
                    Chevron(8f)
                }
            }
        }
    }
    Hoja(correo, { correo = false }) { HojaCorreo { actualizarHoja?.invoke() } }
}

// MARK: - Usuarios y permisos

val rolesEquipo = listOf("admin" to "Administración", "ventas" to "Ventas", "local" to "Local")
fun nombreRolEquipo(r: String): String = rolesEquipo.firstOrNull { it.first == r }?.second ?: if (r.startsWith("due")) "Dueña" else Formato.mayuscula(r)

@Composable
fun UsuariosYPermisos() {
    val estado = LocalEstado.current
    var invitar by remember { mutableStateOf(false) }
    var actualizarHoja by remember { mutableStateOf<(suspend () -> Unit)?>(null) }
    @Serializable data class Miembro(val id: String, val nombre: String = "", val contacto: String = "", val rol: String = "", val pendiente: Boolean = false, val vos: Boolean = false)
    @Serializable data class Datos(val equipo: List<Miembro> = emptyList())
    @Serializable data class C2(val rol: String)

    PantallaSeccion("Usuarios y permisos", "el equipo", { estado.api.get<Datos>("/v1/panel/configuracion") }) { d, actualizar ->
        actualizarHoja = actualizar
        val activos = d.equipo.count { !it.pendiente }
        GrillaIndicadores(listOf(
            IndicadorIcono("person.2.fill", "En el equipo", "$activos"),
            IndicadorIcono("envelope.badge", "Invitaciones pendientes", "${d.equipo.size - activos}", tono = T.avisoFondo to T.avisoTinta),
        ))
        GrupoFilas(d.equipo, { it.id }, 68.dp) { m ->
            Row(Modifier.fillMaxWidth().padding(horizontal = 14.dp, vertical = 12.dp), horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
                AvatarIniciales(m.nombre.ifEmpty { m.contacto }, 42.dp)
                Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                    Texto(if (m.nombre.isEmpty()) m.contacto else m.nombre + (if (m.vos) " (vos)" else ""), Estilo.fila, C.texto, lineas = 1)
                    Texto(if (m.pendiente) "Invitación pendiente · ${m.contacto}" else m.contacto, Estilo.footnote, C.texto2, lineas = 1)
                }
                if (m.vos || m.rol == "duena" || m.rol == "dueña") {
                    Texto(nombreRolEquipo(m.rol), Estilo.label, C.texto2)
                } else {
                    MenuBoton(buildList {
                        rolesEquipo.forEach { r ->
                            add(OpcionDeMenu(r.second, "") {
                                estado.tarea {
                                    try { estado.api.hacer("PUT", "/v1/panel/equipo/${m.id}", C2(r.first)); actualizar() } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
                                        estado.avisar("No pudimos cambiar el rol", ok = false)
                                    }
                                }
                            })
                        }
                        if (m.pendiente) add(OpcionDeMenu("Reenviar invitación", "") {
                            estado.tarea { intentar { estado.api.hacer("POST", "/v1/panel/equipo/${m.id}/reenviar") }; estado.avisar("Invitación reenviada") }
                        })
                        add(OpcionDeMenu("Quitar del equipo", "", destructiva = true) {
                            estado.tarea {
                                try { estado.api.hacer("DELETE", "/v1/panel/equipo/${m.id}"); actualizar() } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
                                    estado.avisar("No pudimos quitarlo", ok = false)
                                }
                            }
                        })
                    }, "Rol de ${m.nombre}") {
                        Row(horizontalArrangement = Arrangement.spacedBy(4.dp), verticalAlignment = Alignment.CenterVertically) {
                            Texto(nombreRolEquipo(m.rol), Estilo.label, C.primarioTexto)
                            Icono("chevron.down", 11f, C.primarioTexto)
                        }
                    }
                }
            }
        }
        BotonSecundario("Invitar a alguien", { invitar = true }, Modifier.padding(top = 4.dp), lleno = true)
    }
    Hoja(invitar, { invitar = false }, listOf(Altura.media)) { HojaInvitar { actualizarHoja?.invoke() } }
}

@Composable
fun HojaInvitar(listo: suspend () -> Unit) {
    val estado = LocalEstado.current
    val cerrar = LocalCerrarHoja.current
    val alcance = rememberCoroutineScope()
    var email by remember { mutableStateOf("") }
    var rol by remember { mutableStateOf("ventas") }
    var enviando by remember { mutableStateOf(false) }
    @Serializable data class C2(val contacto: String, val rol: String)
    Column(Modifier.fillMaxSize().imePadding().padding(horizontal = 20.dp), verticalArrangement = Arrangement.spacedBy(14.dp)) {
        Grabber()
        Texto("Invitar a alguien", Estilo.tituloHojaChico, C.texto)
        val forma = RoundedCornerShape(14.dp)
        Entrada(email, { email = it }, "Email", Modifier.fillMaxWidth().clip(forma).background(C.superficie).border(1.dp, C.linea2, forma).padding(14.dp),
            estilo = Estilo.body16, teclado = KeyboardType.Email)
        Segmentado(rolesEquipo, rol, { rol = it })
        BotonPrincipal(if (enviando) "Enviando…" else "Mandar invitación", {
            alcance.launch {
                enviando = true
                try {
                    estado.api.hacer("POST", "/v1/panel/equipo", C2(email.trim(), rol))
                    estado.avisar("Le mandamos la invitación")
                    listo()
                    cerrar()
                } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) { estado.avisar("No pudimos invitarla", ok = false) }
                enviando = false
            }
        }, habilitado = email.contains("@") && !enviando)
    }
}
