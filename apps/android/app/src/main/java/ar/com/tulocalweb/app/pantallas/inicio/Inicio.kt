package ar.com.tulocalweb.app.pantallas.inicio

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.IntrinsicSize
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
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
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Path
import androidx.compose.ui.graphics.PathEffect
import androidx.compose.ui.graphics.drawscope.clipRect
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.heading
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.AnnotatedString
import androidx.compose.ui.text.SpanStyle
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.buildAnnotatedString
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextDecoration
import androidx.compose.ui.text.withStyle
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.em
import androidx.compose.ui.unit.sp
import ar.com.tulocalweb.app.app.Destino
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.app.LocalNavegacion
import ar.com.tulocalweb.app.app.Pestana
import ar.com.tulocalweb.app.nucleo.Comparado
import ar.com.tulocalweb.app.nucleo.ErrorApi
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.Guardado
import ar.com.tulocalweb.app.nucleo.Marcas
import ar.com.tulocalweb.app.nucleo.PaginaPedidos
import ar.com.tulocalweb.app.nucleo.PaginaProductos
import ar.com.tulocalweb.app.nucleo.Preferencias
import ar.com.tulocalweb.app.nucleo.Tablero
import ar.com.tulocalweb.app.nucleo.Tincho
import ar.com.tulocalweb.app.nucleo.compartirTexto
import ar.com.tulocalweb.app.nucleo.errorDe
import ar.com.tulocalweb.app.nucleo.opcional
import ar.com.tulocalweb.app.pantallas.compartir.TarjetaTiendaOnline
import ar.com.tulocalweb.app.tema.*
import kotlinx.coroutines.launch
import kotlinx.serialization.Serializable
import java.time.Instant
import java.time.YearMonth
import kotlin.math.abs
import kotlin.math.max
import kotlin.math.roundToInt

// Clon de Pantallas/Inicio/Inicio.swift.

/** 2a · Inicio: cómo viene el día (ventas con gráfico, indicadores), lo que hay para hacer hoy, la meta del mes, ideas de
 *  Tincho, últimos pedidos, lo más vendido y canales. Con /v1/panel/tablero. */
@Composable
fun Inicio() {
    val estado = LocalEstado.current
    val nav = LocalNavegacion.current
    val ancho = LocalAnchoPantalla.current
    var periodo by remember { mutableStateOf("hoy") }
    var carga by remember { mutableStateOf<Carga<Tablero>>(Carga.cargando) }
    var masViejo by remember { mutableStateOf<Instant?>(null) }
    var sinLeer by remember { mutableStateOf(0) }
    var editarMeta by remember { mutableStateOf(false) }
    var ideasOcultas by remember { mutableStateOf(setOf<String>()) }
    val se = ancho <= 375f

    @Serializable data class SinLeer(val sinLeer: Int = 0)

    suspend fun cargar() {
        try {
            val t: Tablero = estado.api.get("/v1/panel/tablero", listOf("periodo" to periodo))
            carga = Carga.listo(t)
            Guardado.guardar(t, "tablero-$periodo")
            if (t.paraHacer.preparar > 0) {
                val l = estado.api.opcional<PaginaPedidos>("/v1/panel/pedidos", listOf("estado" to "Para preparar", "orden" to "viejos", "cuantos" to "1"))
                masViejo = l?.filas?.firstOrNull()?.creado
            }
            estado.api.opcional<SinLeer>("/v1/panel/notificaciones/sin-leer")?.let { sinLeer = it.sinLeer }
        } catch (e: kotlinx.coroutines.CancellationException) {
            throw e
        } catch (e: Exception) {
            val x = errorDe(e)
            val guardado = if (x.sinConexion) Guardado.leer<Tablero>("tablero-$periodo") else null
            if (guardado != null) carga = Carga.listo(guardado.first)
            else if (carga.valor == null || e !is ErrorApi) carga = Carga.error(x)
        }
    }

    Refrescable({ cargar() }) {
        Column(Modifier.fillMaxSize().verticalScroll(rememberScrollState()).padding(bottom = 90.dp)) {
            Encabezado(carga.valor?.hola ?: estado.datos?.nombreDePila ?: "", sinLeer, se)
            Column(Modifier.padding(top = 18.dp).margenPantalla(), verticalArrangement = Arrangement.spacedBy(14.dp)) {
                AccionesRapidas()
                BarraPreguntarTincho()
                Segmentado(listOf("hoy" to "Hoy", "7d" to "7 días", "30d" to "30 días"), periodo, { periodo = it })
                when (val c = carga) {
                    Carga.cargando -> EsqueletoInicio()
                    is Carga.error -> Box(Modifier.padding(top = 30.dp)) { EstadoError("cómo viene el día", c.e) { estado.tarea { cargar() } } }
                    is Carga.listo -> Contenido(c.v, periodo, se, masViejo, ideasOcultas, { ideasOcultas = ideasOcultas + it }, { editarMeta = true })
                }
            }
        }
    }
    LaunchedEffect(periodo, estado.version) { cargar() }
    Hoja(editarMeta, { editarMeta = false }, altoFijo = 470.dp) {
        carga.valor?.let { t -> EditarMeta(t.meta?.objetivo ?: 0, t.meta?.actual ?: 0) { cargar() } }
    }
}

// MARK: Encabezado

@Composable
private fun Encabezado(nombre: String, sinLeer: Int, se: Boolean) {
    val estado = LocalEstado.current
    val nav = LocalNavegacion.current
    Row(Modifier.fillMaxWidth().padding(top = 6.dp).margenPantalla(), verticalAlignment = Alignment.Bottom) {
        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
            Texto(Formato.fechaLarga(Instant.now()), Estilo.subhead, C.texto2)
            Texto("Hola, $nombre", if (se) Estilo(28f, Peso.bold, -0.025f) else Estilo.saludo, C.texto,
                Modifier.semantics { heading() }, lineas = 1, escalaMinima = 0.8f)
        }
        Spacer(Modifier.width(8.dp))
        Row(Modifier.padding(bottom = 2.dp), horizontalArrangement = Arrangement.spacedBy(10.dp - 4.dp), verticalAlignment = Alignment.CenterVertically) {
            BotonRedondo("magnifyingglass", "Buscar") { nav.buscar = true }
            BotonRedondo("bell", if (sinLeer > 0) "Avisos, $sinLeer sin leer" else "Avisos", sinLeer > 0) { nav.inicio.add(Destino.avisos) }
            Box(Modifier.clickable { nav.tocar(Pestana.mas) }.semantics { contentDescription = "Tu tienda, ${estado.tienda?.nombre ?: ""}" }) {
                AvatarTienda(estado.tienda, 40.dp, 20.dp)
            }
        }
    }
}

// MARK: Acciones rápidas

@Composable
private fun AccionesRapidas() {
    val estado = LocalEstado.current
    val nav = LocalNavegacion.current
    Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
        AccionRapida("barcode.viewfinder", "Vender", Modifier.weight(1f)) { nav.venta = true }
        AccionRapida("camera", "Producto", Modifier.weight(1f)) { nav.inicio.add(Destino.nuevoProducto(null)) }
        if (estado.puede("pedidos")) AccionRapida("link", "Cobrar", Modifier.weight(1f)) { nav.inicio.add(Destino.links) }
        AccionRapida("square.and.arrow.up", "Compartir", Modifier.weight(1f)) { nav.inicio.add(Destino.compartir) }
    }
}

// MARK: Contenido

@Composable
private fun Contenido(t: Tablero, periodo: String, se: Boolean, masViejo: Instant?, ocultas: Set<String>, ocultar: (String) -> Unit, editarMeta: () -> Unit) {
    val estado = LocalEstado.current
    if (t.ultimosPedidos.isEmpty() && t.masVendidos.isEmpty() && (t.ventas?.actual ?: 0.0) == 0.0 && periodo == "hoy"
        && !PrimerosPasos.ocultos(estado.tienda?.id)) {
        SinVentas()
        TarjetaTiendaOnline(t.visitas, periodo)
    } else {
        val v = t.ventas
        if (t.montos && v != null) TarjetaVentas(t, v, periodo, se)
        Indicadores(t, periodo)
        TarjetaTiendaOnline(t.visitas, periodo)
        ParaHacer(t, masViejo)
        if (t.montos) t.meta?.let { Meta(it, editarMeta) }
        Ideas(t, ocultas, ocultar)
        if (t.ultimosPedidos.isNotEmpty()) Ultimos(t)
        t.enVivo?.let { if (it.personas > 0 || it.carritos > 0) TarjetaEnVivo(it, t.montos) }
        if (t.montos) t.caja?.let { if (it.abierta != null || it.total > 0) TarjetaCajaLocal(it) }
        if (t.montos) t.liquidaciones?.let { if (it.isNotEmpty()) TarjetaLiquidaciones(it) }
        if (estado.puede("cuenta")) TarjetaIntegraciones()
        if (t.masVendidos.isNotEmpty()) MasVendidos(t)
        t.visitas?.let { if (it.masVistos.isNotEmpty()) TarjetaMasVistos(it) }
        t.resenas?.let { if (it.nuevas.isNotEmpty()) TarjetaResenas(it) }
        t.envios?.let { if (it.paquetes > 0 || it.retiros.isNotEmpty()) TarjetaEnvios(it) }
        val canales = t.canales
        if (t.montos && !canales.isNullOrEmpty() && v != null && v.actual > 0) Canales(canales, v.actual)
    }
}

private fun textoComparacion(periodo: String) = when (periodo) {
    "hoy" -> "vs. ayer a esta hora"
    "7d" -> "vs. los 7 días anteriores"
    else -> "vs. los 30 días anteriores"
}

@Composable
private fun TarjetaVentas(t: Tablero, v: Comparado, periodo: String, se: Boolean) {
    Column(Modifier.fillMaxWidth().tarjeta(radio = 22.dp).padding(start = 18.dp, end = 18.dp, top = 18.dp, bottom = 14.dp), verticalArrangement = Arrangement.spacedBy(4.dp)) {
        Texto(if (periodo == "hoy") "Ventas de hoy" else if (periodo == "7d") "Ventas de 7 días" else "Ventas de 30 días", Estilo.subheadFuerte, C.texto2)
        NumeroAnimado(Formato.pesos(v.actual), if (se) Estilo(38f, Peso.bold, -0.035f) else Estilo.montoHeroe, C.texto, escalaMinima = 0.6f)
        Row(horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.CenterVertically) {
            val d = v.variacion
            if (d != null) Delta(d)
            else if (v.actual > 0) Texto("Primeras ventas", Estilo.texto14.con(Peso.bold), C.okTinta,
                Modifier.clip(Capsula).background(C.okFondo).padding(horizontal = 8.dp, vertical = 3.dp))
            Texto(textoComparacion(periodo), Estilo.texto14, C.texto2)
        }
        val s = t.serie
        if (!se && s != null) {
            Box(Modifier.padding(top = 12.dp).offset(x = (-2).dp)) {
                GraficoVentas(s.actual, s.anterior, etiqueta = EtiquetasVentas.etiquetas(t.desde, t.hasta, s.actual.size, periodo == "hoy"))
            }
            Ejes(t, periodo)
            Row(Modifier.padding(top = 8.dp), horizontalArrangement = Arrangement.spacedBy(16.dp)) {
                Row(horizontalArrangement = Arrangement.spacedBy(6.dp), verticalAlignment = Alignment.CenterVertically) {
                    Box(Modifier.size(14.dp, 3.dp).clip(Capsula).background(C.acento))
                    Texto(if (periodo == "hoy") "Hoy" else "Este período", Estilo.footnote, C.texto2)
                }
                Row(horizontalArrangement = Arrangement.spacedBy(6.dp), verticalAlignment = Alignment.CenterVertically) {
                    Punteado(Modifier.size(14.dp, 2.dp))
                    Texto("${if (periodo == "hoy") "Ayer" else "Anterior"} · ${Formato.pesos(v.anterior)}", Estilo.footnote, C.texto2, numeros = true)
                }
            }
        }
    }
}

@Composable
private fun Ejes(t: Tablero, periodo: String) {
    val total = t.hasta.toEpochMilli() - t.desde.toEpochMilli()
    val marcas = (0 until 4).map { i -> Instant.ofEpochMilli(t.desde.toEpochMilli() + (total * (i * 2 + 1) / 8)) }
    Row(Modifier.fillMaxWidth().padding(horizontal = 2.dp).padding(top = 2.dp)) {
        marcas.forEachIndexed { i, f ->
            if (i > 0) Spacer(Modifier.weight(1f))
            val z = Formato.enArgentina(f)
            Texto(if (periodo == "hoy") "${z.hour}${Formato.fino}h" else "${z.dayOfMonth}/${z.monthValue}", Estilo.chip.con(Peso.regular), C.texto2, numeros = true)
        }
    }
}

@Composable
private fun Indicadores(t: Tablero, periodo: String) {
    val comp = if (periodo == "hoy") "vs. ayer" else "vs. antes"
    fun diferencia(c: Comparado): String? {
        val d = (c.actual - c.anterior).toInt()
        if (d == 0) return "igual $comp"
        return "${if (d > 0) "+" else "−"}${abs(d)} $comp"
    }
    fun porcentajeConSigno(p: Double): String {
        val r = (p * 10).roundToInt() / 10.0
        return "${if (r >= 0) "+" else "−"}${Formato.porcentaje(abs(r))}"
    }
    Row(Modifier.fillMaxWidth().height(IntrinsicSize.Max), horizontalArrangement = Arrangement.spacedBy(10.dp)) {
        Indicador("Pedidos", Formato.numero(t.kpis.pedidos.actual.toInt()), diferencia(t.kpis.pedidos), t.kpis.pedidos.actual >= t.kpis.pedidos.anterior, Modifier.weight(1f))
        val tk = t.kpis.ticket
        if (t.montos && tk != null) Indicador("Ticket prom.", Formato.pesos(tk.actual), tk.variacion?.let { porcentajeConSigno(it) }, tk.actual >= tk.anterior, Modifier.weight(1f))
        Indicador("Clientes nuevos", Formato.numero(t.kpis.clientesNuevos.actual.toInt()), diferencia(t.kpis.clientesNuevos),
            t.kpis.clientesNuevos.actual >= t.kpis.clientesNuevos.anterior, Modifier.weight(1f))
        // La grilla de iOS es de tres columnas: con dos indicadores queda una vacía.
        if (!(t.montos && tk != null)) Spacer(Modifier.weight(1f))
    }
}

// MARK: Para hacer hoy

private data class Pendiente(val n: Int, val fondo: Tono, val tinta: Tono, val titulo: String, val detalle: String, val accion: () -> Unit)

@Composable
private fun ParaHacer(t: Tablero, masViejo: Instant?) {
    val estado = LocalEstado.current
    val nav = LocalNavegacion.current
    val ph = t.paraHacer
    val filas = buildList {
        if (ph.preparar > 0) add(Pendiente(ph.preparar, T.avisoFondo, T.avisoTinta, "Pedidos por preparar",
            masViejo?.let { "El más viejo entró ${Formato.hace(it)}" } ?: "Tocá para verlos") { nav.filtroPedidos = "Para preparar"; nav.tocar(Pestana.pedidos) })
        if (ph.transferencias > 0) add(Pendiente(ph.transferencias, T.crema, T.primarioTinta, "Transferencias por confirmar",
            ph.transferenciasMonto?.let { "${Formato.pesos(it)} en total" } ?: "Revisá que hayan llegado") { nav.filtroPedidos = "Sin pagar"; nav.tocar(Pestana.pedidos) })
        if (ph.sinStock > 0) add(Pendiente(ph.sinStock, T.errorFondo, T.errorTinta, "Productos sin stock", "Siguen visibles en la tienda") {
            nav.segmentoProductos = "out"; nav.tocar(Pestana.productos)
        })
        val n = t.resenas?.nuevas?.size ?: 0
        if (n > 0 && estado.puede("clientes")) add(Pendiente(n, T.crema, T.primarioTinta, if (n == 1) "Reseña sin responder" else "Reseñas sin responder",
            "Contestarlas suma confianza") { nav.inicio.add(Destino.resenas) })
    }
    val compartir = filas.isEmpty() || (t.visitas?.hoy?.actual ?: 0.0).toInt() == 0
    Box(Modifier.padding(top = 14.dp)) { TituloSeccion("Para hacer hoy") }
    Column(Modifier.fillMaxWidth().tarjeta()) {
        if (filas.isEmpty()) {
            Row(Modifier.fillMaxWidth().padding(horizontal = 16.dp, vertical = 13.dp), horizontalArrangement = Arrangement.spacedBy(13.dp), verticalAlignment = Alignment.CenterVertically) {
                Box(Modifier.size(40.dp).clip(RoundedCornerShape(12.dp)).background(C.okFondo), contentAlignment = Alignment.Center) { Icono("checkmark", 16f, C.okTinta) }
                Column(verticalArrangement = Arrangement.spacedBy(1.dp)) {
                    Texto("Todo al día", Estilo.fila, C.texto)
                    Texto("No hay pedidos ni transferencias esperando.", Estilo.footnote, C.texto2)
                }
            }
        } else filas.forEachIndexed { i, f ->
            if (i > 0) Separador(69.dp)
            FilaPendiente(f.n, f.fondo.actual, f.tinta.actual, f.titulo, f.detalle, f.accion)
        }
        if (compartir) {
            Separador(69.dp)
            // La fila de compartir cuando no hay nada urgente o la tienda no tuvo visitas hoy: «Para hacer» nunca queda en
            // «Todo al día» sin una acción.
            val hoy = (t.visitas?.hoy?.actual ?: 0.0).toInt()
            Row(Modifier.fillMaxWidth().clickable { nav.inicio.add(Destino.compartir) }.padding(horizontal = 16.dp, vertical = 13.dp),
                horizontalArrangement = Arrangement.spacedBy(13.dp), verticalAlignment = Alignment.CenterVertically) {
                Box(Modifier.size(40.dp).clip(RoundedCornerShape(12.dp)).background(C.crema), contentAlignment = Alignment.Center) { Icono("square.and.arrow.up", 16f, C.primarioTexto) }
                Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(1.dp)) {
                    Texto("Compartí tu tienda", Estilo.fila, C.texto)
                    Texto(if (hoy == 0) "Hoy todavía no tuvo visitas" else "Más visitas, más ventas", Estilo.footnote, C.texto2)
                }
                Chevron()
            }
        }
    }
}

// MARK: Meta

@Composable
private fun Meta(m: Tablero.Meta, editar: () -> Unit) {
    val hoy = Formato.enArgentina(Instant.now())
    val diasMes = YearMonth.from(hoy).lengthOfMonth()
    val quedan = diasMes - hoy.dayOfMonth + 1
    val pct = if (m.objetivo > 0) m.actual.toDouble() / m.objetivo else 0.0
    val proy = if (m.objetivo > 0) m.proyeccion.toDouble() / m.objetivo else 0.0
    Column(Modifier.padding(top = 6.dp).fillMaxWidth().tarjeta(radio = 22.dp).padding(18.dp), verticalArrangement = Arrangement.spacedBy(10.dp)) {
        Row(verticalAlignment = Alignment.Bottom) {
            Texto("Meta de ${Formato.mesDe(Instant.now())}", Estilo.subheadFuerte, C.texto2, Modifier.weight(1f))
            Texto(if (m.objetivo > 0) "Editar" else "Ponerla", Estilo.subheadFuerte, C.primarioTexto, Modifier.clickable(onClick = editar))
        }
        if (m.objetivo > 0) {
            val t2 = C.texto2
            // .minimumScaleFactor(0.7): el texto de abajo va en em para que se achique junto con el monto.
            TextoRico(buildAnnotatedString {
                append(Formato.pesos(m.actual))
                withStyle(Estilo.subhead.ts(true).toSpanStyle().copy(color = t2, fontSize = (Estilo.subhead.tamano / Estilo.montoMeta.tamano).em)) {
                    append("  de ${Formato.pesos(m.objetivo)}")
                }
            }, Estilo.montoMeta, C.texto, lineas = 1, numeros = true, escalaMinima = 0.7f)
            BoxWithConstraints(Modifier.fillMaxWidth().height(12.dp).clip(Capsula)) {
                Box(Modifier.fillMaxSize().background(C.pistaMeta))
                RayadoMeta(Modifier.width(maxWidth * minOf(1.0, proy).toFloat()).fillMaxHeight())
                Box(Modifier.width(maxOf(12.dp, maxWidth * minOf(1.0, pct).toFloat())).fillMaxHeight().clip(Capsula)
                    .background(Brush.horizontalGradient(listOf(C.acento, hex(0xB0441D)))))
            }
            Row {
                val tx = C.texto
                TextoRico(buildAnnotatedString {
                    withStyle(SpanStyle(fontWeight = FontWeight.Bold, color = tx)) { append(Formato.porcentaje(Math.round(pct * 100).toDouble())) }
                    append(" vendido")
                }, Estilo.footnote, C.texto2, Modifier.weight(1f), numeros = true)
                Texto(if (quedan == 1) "Queda 1 día" else "Quedan $quedan días", Estilo.footnote, C.texto2, numeros = true)
            }
            val tx = C.texto
            TextoRico(buildAnnotatedString {
                append("Al ritmo de estos días cerrás el mes en ")
                withStyle(SpanStyle(fontWeight = FontWeight.Bold, color = tx)) { append(Formato.pesos(m.proyeccion)) }
                append(".")
            }, Estilo.texto14, C.texto2)
        } else {
            Texto(Formato.sinViudas("Poné cuánto querés vender este mes y te decimos cómo venís cada día."), Estilo.texto14, C.texto2)
        }
    }
}

// MARK: Ideas de Tincho

private data class Idea(val id: String, val texto: AnnotatedString, val boton: String, val accion: suspend () -> Unit)

@Composable
private fun Ideas(t: Tablero, ocultas: Set<String>, ocultar: (String) -> Unit) {
    val estado = LocalEstado.current
    val nav = LocalNavegacion.current
    val alcance = rememberCoroutineScope()
    val ancho = LocalAnchoPantalla.current
    val m = margen

    @Serializable data class R(val enviados: Int = 0)

    suspend fun recordarCarritos() {
        try {
            val r: R = estado.api.enviar("POST", "/v1/panel/carritos/recordar", emptyMap<String, String>())
            Haptico.exito()
            if (r.enviados == 0) {
                // Ya se les mandó en las últimas 24 horas (no se repite para no molestar).
                estado.avisar("Ya les mandamos un mail hace menos de un día. Mañana podés volver a recordarles.")
            } else {
                estado.avisar(if (r.enviados == 1) "Listo: le mandamos un mail a 1 cliente" else "Listo: les mandamos un mail a ${r.enviados} clientes")
            }
            ocultar("carritos")
            estado.version += 1
        } catch (e: ErrorApi) { estado.avisar(e.detalle, ok = false) } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {}
    }

    val lista = buildList {
        // Ya recordados: el mail salió y no se repite antes de 24 h, así que la idea no vuelve.
        if (t.hornero.carritos > 0 && "carritos" !in t.hornero.ideasHechas && t.hornero.recordados != true) {
            val s = buildAnnotatedString {
                append(if (t.hornero.carritos == 1) "Hay 1 carrito abandonado" else "Hay ${t.hornero.carritos} carritos abandonados")
                val monto = t.hornero.carritosMonto
                if (monto != null && t.montos) {
                    append(" por ")
                    withStyle(SpanStyle(fontWeight = FontWeight.Bold)) { append(Formato.pesos(monto)) }
                }
                append(". ¿Les mando un mail para recordarles lo que dejaron?")
            }
            add(Idea("carritos", s, "Sí, mandales") { recordarCarritos() })
        }
        if (t.paraHacer.sinStock > 0 && "stock" !in t.hornero.ideasHechas) {
            val n = t.paraHacer.sinStock
            add(Idea("stock", AnnotatedString(if (n == 1) "Un producto se quedó sin stock y sigue visible en la tienda. ¿Le cargás un ingreso?"
                else "$n productos se quedaron sin stock y siguen visibles en la tienda. ¿Les cargás un ingreso?"), "Cargar") {
                nav.segmentoProductos = "out"; nav.tocar(Pestana.productos)
            })
        }
        val v = t.visitas
        if (v != null && "compartir" !in t.hornero.ideasHechas) {
            val n = v.dias7.actual.toInt()
            if (v.desde == null || n < 30) {
                add(Idea("compartir", AnnotatedString(if (v.desde == null || n == 0) "Tu tienda todavía no tuvo visitas. ¿Te armo una imagen con el QR para tus estados de WhatsApp?"
                    else "En 7 días te visitaron ${if (n == 1) "1 persona" else "$n personas"}. Con una historia o un estado llegás a muchas más. ¿La armamos?"), "Armarla") {
                    nav.inicio.add(Destino.compartir)
                })
            }
        }
    }.filter { it.id !in ocultas }
    if (lista.isEmpty()) return
    Row(Modifier.padding(top = 14.dp), horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.CenterVertically) {
        ImagenTincho(Tincho.avatar, Modifier.size(28.dp).clip(CircleShape).background(C.crema), ContentScale.Crop)
        Texto("Ideas de Tincho", Estilo.tituloSeccion, C.texto)
    }
    Row(Modifier.fillMaxWidth().sinMargen().horizontalScroll(rememberScrollState()).height(IntrinsicSize.Max)
        .padding(horizontal = m), horizontalArrangement = Arrangement.spacedBy(10.dp)) {
        lista.forEach { idea ->
            val forma = RoundedCornerShape(20.dp)
            Column(Modifier.width(minOf(300.dp, ancho.dp - m * 2 - 30.dp)).fillMaxHeight().clip(forma).background(C.crema2)
                .border(1.dp, hex(0xB0441D, 0.16f), forma).padding(16.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
                TextoRico(idea.texto, Estilo.tarjetaTincho, C.texto)
                Spacer(Modifier.weight(1f))
                Row(horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.CenterVertically) {
                    Box(Modifier.height(38.dp).clip(RoundedCornerShape(11.dp)).background(C.primario).clickable { alcance.launch { idea.accion() } }.padding(horizontal = 16.dp),
                        contentAlignment = Alignment.Center) { Texto(idea.boton, Estilo.subheadFuerte, Color.White) }
                    Box(Modifier.height(38.dp).clickable { ocultar(idea.id) }.padding(horizontal = 12.dp), contentAlignment = Alignment.Center) {
                        Texto("Ahora no", Estilo.subheadFuerte, C.primarioTexto)
                    }
                }
            }
        }
    }
}

// MARK: Últimos pedidos

@Composable
private fun Ultimos(t: Tablero) {
    val nav = LocalNavegacion.current
    Column(verticalArrangement = Arrangement.spacedBy(14.dp)) {
        Box(Modifier.padding(top = 14.dp)) { TituloSeccion("Últimos pedidos", "Ver todos") { nav.filtroPedidos = null; nav.tocar(Pestana.pedidos) } }
        Column(Modifier.fillMaxWidth().tarjeta()) {
            t.ultimosPedidos.take(5).forEachIndexed { i, p ->
                if (i > 0) Separador(68.dp)
                Row(Modifier.fillMaxWidth().clickable { nav.pestana = Pestana.pedidos; nav.poner(Pestana.pedidos, listOf(Destino.pedido(p.numero))) }
                    .padding(horizontal = 14.dp, vertical = 12.dp), horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
                    Miniatura(p.foto, 42.dp, radio = 11.dp)
                    Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                        Texto(p.cliente, Estilo.fila, C.texto, lineas = 1)
                        Texto("#${p.numero} · ${Marcas.canalCorto(p.canal)} · ${Formato.horaDe(p.creado)}", Estilo.footnote, C.texto2, lineas = 1)
                    }
                    Column(horizontalAlignment = Alignment.End, verticalArrangement = Arrangement.spacedBy(4.dp)) {
                        p.total?.let { Texto(Formato.pesos(it), Estilo.fila, C.texto, numeros = true) }
                        ChipEstado(p.estado)
                    }
                }
            }
        }
    }
}

// MARK: Lo más vendido

@Composable
private fun MasVendidos(t: Tablero) {
    val m = margen
    val ancho = LocalAnchoPantalla.current
    Column(verticalArrangement = Arrangement.spacedBy(14.dp)) {
        Row(Modifier.padding(top = 14.dp), verticalAlignment = Alignment.Bottom) {
            Texto("Lo más vendido", Estilo.tituloSeccion, C.texto, Modifier.weight(1f))
            Texto("Esta semana", Estilo.subhead, C.texto2)
        }
        Row(Modifier.fillMaxWidth().sinMargen().horizontalScroll(rememberScrollState()).padding(horizontal = m), horizontalArrangement = Arrangement.spacedBy(10.dp), verticalAlignment = Alignment.Top) {
            t.masVendidos.forEach { p ->
                Column(Modifier.width(136.dp), verticalArrangement = Arrangement.spacedBy(7.dp)) {
                    Miniatura(p.foto, 136.dp, 120.dp, 16.dp)
                    Texto(p.nombre, Estilo(14f, Peso.semibold), C.texto, lineas = 2, interlineado = 1f)
                    Texto(Formato.cantidad(p.unidades, "unidad", "unidades"), Estilo.footnote, C.texto2, numeros = true)
                }
            }
        }
    }
}

// MARK: Canales

@Composable
private fun Canales(canales: List<Tablero.Canal>, total: Double) {
    Column(verticalArrangement = Arrangement.spacedBy(14.dp)) {
        Texto("Canales", Estilo.tituloSeccion, C.texto, Modifier.padding(top = 14.dp))
        Column(Modifier.fillMaxWidth().tarjeta().padding(16.dp), verticalArrangement = Arrangement.spacedBy(14.dp)) {
            BoxWithConstraints(Modifier.fillMaxWidth().height(14.dp).clip(RoundedCornerShape(7.dp))) {
                Row(horizontalArrangement = Arrangement.spacedBy(2.dp)) {
                    canales.forEach { c ->
                        Box(Modifier.width(maxOf(2.dp, (this@BoxWithConstraints.maxWidth - 2.dp * (canales.size - 1)) * c.parte.toFloat())).fillMaxHeight().background(Marcas.colorCanal(c.canal)))
                    }
                }
            }
            Column(verticalArrangement = Arrangement.spacedBy(9.dp)) {
                canales.forEach { c ->
                    Row(horizontalArrangement = Arrangement.spacedBy(9.dp), verticalAlignment = Alignment.CenterVertically) {
                        Box(Modifier.size(10.dp).clip(RoundedCornerShape(3.dp)).background(Marcas.colorCanal(c.canal)))
                        Texto(Marcas.canal(c.canal), Estilo.subhead, C.texto, Modifier.weight(1f), numeros = true)
                        Texto(Formato.pesos(total * c.parte), Estilo.subhead, C.texto2, numeros = true)
                        Texto(Formato.porcentaje(Math.round(c.parte * 100).toDouble()), Estilo.subhead.con(Peso.bold), C.texto, Modifier.width(52.dp),
                            alinear = androidx.compose.ui.text.style.TextAlign.End, numeros = true)
                    }
                }
            }
        }
    }
}

// MARK: Cargando (2c)

@Composable
private fun EsqueletoInicio() {
    Column(verticalArrangement = Arrangement.spacedBy(14.dp)) {
        Column(Modifier.fillMaxWidth().tarjeta(radio = 22.dp).padding(18.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
            Esqueleto(110.dp, 14.dp)
            Esqueleto(220.dp, 42.dp, 10.dp)
            Esqueleto(160.dp, 20.dp)
            Esqueleto(null, 130.dp, 14.dp, suave = true)
        }
        Row(horizontalArrangement = Arrangement.spacedBy(10.dp)) {
            repeat(3) { Box(Modifier.weight(1f).height(84.dp).tarjeta(radio = 18.dp)) }
        }
        Box(Modifier.padding(top = 14.dp)) { Esqueleto(150.dp, 20.dp) }
        EsqueletoFilas(2, 40.dp)
    }
}

/** Chip de variación: `ok-bg` «▲ 12,4 %» o `error-bg` «▼ 3 %». */
@Composable
fun Delta(variacion: Double) {
    val sube = variacion >= 0
    val r = (abs(variacion) * 10).roundToInt() / 10.0
    Texto("${if (sube) "▲" else "▼"} ${Formato.porcentaje(r)}", Estilo.texto14.con(Peso.bold), if (sube) C.okTinta else C.errorTinta,
        Modifier.clip(Capsula).background(if (sube) C.okFondo else C.errorFondo).padding(horizontal = 8.dp, vertical = 3.dp)
            .semantics { contentDescription = if (sube) "Subió ${Formato.porcentaje(r)}" else "Bajó ${Formato.porcentaje(r)}" }, numeros = true)
}

/** Indicador chico (radio 18): título 13, valor 22/700 y delta 12/600 coloreado. */
@Composable
fun Indicador(titulo: String, valor: String, delta: String?, positivo: Boolean, modifier: Modifier = Modifier) {
    Column(modifier.fillMaxHeight().tarjeta(radio = 18.dp).padding(horizontal = 13.dp, vertical = 12.dp), verticalArrangement = Arrangement.spacedBy(2.dp)) {
        Texto(titulo, Estilo.footnote, C.texto2, lineas = 2)
        Texto(valor, Estilo.indicador, C.texto, lineas = 1, numeros = true, escalaMinima = 0.6f)
        if (delta != null) Texto(delta, Estilo.chip, if (positivo) C.okTinta else C.errorTinta, lineas = 1, escalaMinima = 0.8f)
    }
}

/** Fila de «Para hacer hoy» con el cuadro numérico de 40. */
@Composable
fun FilaPendiente(n: Int, fondo: Color, tinta: Color, titulo: String, detalle: String, accion: () -> Unit) {
    Row(Modifier.fillMaxWidth().clickable(onClick = accion).padding(horizontal = 16.dp, vertical = 13.dp), horizontalArrangement = Arrangement.spacedBy(13.dp),
        verticalAlignment = Alignment.CenterVertically) {
        Box(Modifier.size(40.dp).clip(RoundedCornerShape(12.dp)).background(fondo), contentAlignment = Alignment.Center) {
            Texto("$n", Estilo(18f, Peso.bold), tinta, numeros = true)
        }
        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(1.dp)) {
            Texto(titulo, Estilo.fila, C.texto)
            Texto(detalle, Estilo.footnote, C.texto2, numeros = true)
        }
        Chevron()
    }
}

/** Línea punteada de la leyenda. */
@Composable
fun Punteado(modifier: Modifier = Modifier) {
    Canvas(modifier) {
        drawLine(hex(0x8A7364), Offset(0f, size.height / 2), Offset(size.width, size.height / 2), 2.dp.toPx(),
            pathEffect = PathEffect.dashPathEffect(floatArrayOf(3.dp.toPx(), 2.dp.toPx())))
    }
}

/** La proyección rayada detrás de la barra de la meta. */
@Composable
fun RayadoMeta(modifier: Modifier = Modifier) {
    val acento = C.acento
    Canvas(modifier) {
        drawRect(acento.copy(alpha = 0.15f))
        clipRect {
            var x = -size.height
            val g = 4.dp.toPx()
            while (x < size.width) {
                val p = Path().apply { moveTo(x, size.height); lineTo(x + size.height, 0f); lineTo(x + size.height + g, 0f); lineTo(x + g, size.height); close() }
                drawPath(p, acento.copy(alpha = 0.35f))
                x += 8.dp.toPx() * 1.4142f
            }
        }
    }
}

/** 2d · Sin ventas todavía: Tincho saludando, «Tu tienda ya está online» y los pasos (el hecho tachado con check verde;
 *  el siguiente con «Cargar»). */
@Composable
fun SinVentas() {
    val estado = LocalEstado.current
    val nav = LocalNavegacion.current
    val ctx = LocalContext.current
    var conProductos by remember { mutableStateOf<Boolean?>(null) }
    var mpConectada by remember { mutableStateOf(false) }
    var compartidos by remember { mutableStateOf(Preferencias.texto(PrimerosPasos.claveCompartido) ?: "") }
    val mp = mpConectada || (estado.resumen?.canales?.contains("mp") ?: false)
    val url = "https://${estado.tienda?.direccion ?: "tulocalweb.com.ar"}"
    val id = estado.tienda?.id ?: ""
    val compartio = PrimerosPasos.contiene(compartidos, id)

    @Serializable data class MP(val conectada: Boolean? = null)
    LaunchedEffect(estado.version) {
        estado.api.opcional<MP>("/v1/panel/integraciones/mercadopago")?.let { mpConectada = it.conectada ?: false }
        estado.api.opcional<PaginaProductos>("/v1/panel/listas/productos", listOf("cuantos" to "1"))?.let { conProductos = it.total > 0 }
    }

    @Composable
    fun FilaPaso(n: Int, t: String, d: String?, hecho: Boolean, boton: String?, accion: (() -> Unit)?) {
        Row(Modifier.fillMaxWidth().then(if (accion != null && !hecho) Modifier.clickable(onClick = accion) else Modifier).padding(horizontal = 16.dp, vertical = 14.dp),
            horizontalArrangement = Arrangement.spacedBy(13.dp), verticalAlignment = Alignment.CenterVertically) {
            Box(Modifier.size(28.dp), contentAlignment = Alignment.Center) {
                if (hecho) {
                    Box(Modifier.size(28.dp).clip(CircleShape).background(C.ok), contentAlignment = Alignment.Center) { Icono("checkmark", 12f, Color.White) }
                } else {
                    Box(Modifier.size(28.dp).border(2.dp, tema(0x3A1D12, 0xFBF7F2, alfa = 0.2f).actual, CircleShape), contentAlignment = Alignment.Center) {
                        Texto("$n", Estilo(13f, Peso.bold), C.texto2)
                    }
                }
            }
            Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(1.dp)) {
                if (hecho) {
                    androidx.compose.foundation.text.BasicText(t, style = Estilo.body16.ts().copy(color = C.texto2, textDecoration = TextDecoration.LineThrough))
                } else {
                    Texto(t, Estilo.fila, C.texto)
                    if (d != null) Texto(d, Estilo.footnote, C.texto2)
                }
            }
            if (boton != null) {
                Box(Modifier.height(34.dp).clip(RoundedCornerShape(10.dp)).background(C.primario).padding(horizontal = 14.dp), contentAlignment = Alignment.Center) {
                    Texto(boton, Estilo.label, Color.White)
                }
            }
        }
    }

    val forma = RoundedCornerShape(24.dp)
    Column(Modifier.fillMaxWidth().clip(forma).background(C.crema2).border(1.dp, hex(0xB0441D, 0.14f), forma).padding(horizontal = 20.dp).padding(top = 20.dp, bottom = 18.dp),
        horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(8.dp)) {
        ImagenTincho(Tincho.saludo, Modifier.size(150.dp))
        Texto("Tu tienda ya está online", Estilo.tituloHoja, C.texto, alinear = androidx.compose.ui.text.style.TextAlign.Center)
        Texto(Formato.sinViudas("Falta la primera venta. Con estos pasos, la tenés cerca."), Estilo.subhead, C.texto2, alinear = androidx.compose.ui.text.style.TextAlign.Center)
        Column(Modifier.padding(top = 8.dp).fillMaxWidth().tarjeta()) {
            FilaPaso(1, "Creaste la tienda", null, true, null, null)
            Separador(57.dp)
            FilaPaso(2, "Cargá tu primer producto", "Con una foto alcanza", conProductos == true, if (conProductos == false) "Cargar" else null) { nav.inicio.add(Destino.nuevoProducto(null)) }
            Separador(57.dp)
            FilaPaso(3, "Conectá Mercado Pago", "Para cobrar con tarjeta y QR", mp, if (conProductos == true && !mp) "Conectar" else null) { nav.inicio.add(Destino.integraciones) }
            Separador(57.dp)
            // Tocar «Compartir» alcanza para darlo por hecho (no hay forma de saber si se mandó).
            FilaPaso(4, "Compartí el link", "Por WhatsApp o en tus redes", compartio, null) {
                compartidos = PrimerosPasos.sumar(compartidos, id)
                Preferencias.ponerTexto(PrimerosPasos.claveCompartido, compartidos)
                compartirTexto(ctx, url)
            }
        }
        // Ningún paso es obligatorio: la tarjeta se puede cerrar y no vuelve.
        BotonTexto(if ((conProductos == true && mp) || compartio) "Listo, ir al tablero" else "Ocultar estos pasos", {
            Preferencias.ponerTexto(PrimerosPasos.claveOcultos, PrimerosPasos.sumar(Preferencias.texto(PrimerosPasos.claveOcultos) ?: "", id))
            estado.version += 1
        }, Modifier.padding(top = 4.dp))
    }
}

/** Los primeros pasos de Inicio: qué tiendas ya compartieron el link o cerraron la tarjeta (en el teléfono, por tienda). */
object PrimerosPasos {
    const val claveCompartido = "pasos.compartido"
    const val claveOcultos = "pasos.ocultos"

    fun contiene(lista: String, id: String): Boolean = id.isNotEmpty() && id in lista.split(",")

    fun sumar(lista: String, id: String): String {
        if (id.isEmpty() || contiene(lista, id)) return lista
        return if (lista.isEmpty()) id else "$lista,$id"
    }

    fun ocultos(id: String?): Boolean = contiene(Preferencias.texto(claveOcultos) ?: "", id ?: "")
}
