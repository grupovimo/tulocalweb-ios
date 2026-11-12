package ar.com.tulocalweb.app.pantallas.inicio

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.slideInVertically
import androidx.compose.animation.slideOutVertically
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.core.view.WindowCompat
import androidx.compose.ui.platform.LocalView
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.derivedStateOf
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.BlendMode
import androidx.compose.ui.layout.onSizeChanged
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.text.SpanStyle
import androidx.compose.ui.text.AnnotatedString
import androidx.compose.ui.text.buildAnnotatedString
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.withStyle
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.app.Destino
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.app.LocalNavegacion
import ar.com.tulocalweb.app.app.Pestana
import ar.com.tulocalweb.app.nucleo.ErrorApi
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.Guardado
import ar.com.tulocalweb.app.nucleo.PaginaPedidos
import ar.com.tulocalweb.app.nucleo.Preferencias
import ar.com.tulocalweb.app.nucleo.Tablero
import ar.com.tulocalweb.app.nucleo.errorDe
import ar.com.tulocalweb.app.nucleo.opcional
import ar.com.tulocalweb.app.tema.*
import kotlinx.serialization.Serializable
import java.time.Instant

// Clon de Pantallas/Inicio/Inicio.swift.

/** 2a · Inicio, «La vidriera de tu local» (docs/handoff/inicio-a): arriba la vidriera con el color de la tienda (cuánto
 *  se vendió, el gráfico de hoy contra ayer y «Vender ahora»), la tarjeta de Tincho montada sobre su borde, las acciones
 *  en un bento, lo que hay para hacer hoy, un vistazo de costado, los últimos pedidos, lo más vendido y el resto del
 *  tablero. Al bajar, la vidriera se pliega a una barra. Con /v1/panel/tablero. */
@Composable
fun Inicio() {
    val estado = LocalEstado.current
    val ancho = LocalAnchoPantalla.current
    val oscuro = LocalOscuro.current
    var periodo by remember { mutableStateOf("hoy") }
    var carga by remember { mutableStateOf<Carga<Tablero>>(Carga.cargando) }
    var masViejo by remember { mutableStateOf<Instant?>(null) }
    var sinLeer by remember { mutableStateOf(0) }
    var editarMeta by remember { mutableStateOf(false) }
    var ideasOcultas by remember { mutableStateOf(setOf<String>()) }
    // El color de la última carga («#RRGGBB|#RRGGBB»), para que el esqueleto ya salga con el de la tienda.
    var colores by remember { mutableStateOf(Preferencias.texto("inicio.colores") ?: "") }
    val se = ancho <= 375f
    val t = carga.valor
    val paleta = remember(t?.acento, t?.sobreAcento, colores, oscuro) {
        val guardados = colores.split("|").filter { it.isNotEmpty() }
        PaletaTienda(acento = t?.acento ?: guardados.firstOrNull(), sobre = t?.sobreAcento ?: guardados.lastOrNull(), oscuro = oscuro)
    }
    // La barra de estado toma el color de la vidriera: íconos claros sobre un acento oscuro. Sólo
    // mientras Inicio está a la vista (las pestañas quedan compuestas aunque no se vean).
    val nav = LocalNavegacion.current
    val vista = LocalView.current
    val aLaVista = nav.pestana == Pestana.inicio && nav.inicio.isEmpty()
    DisposableEffect(aLaVista, paleta.claroEncima, oscuro) {
        val ventana = (vista.context as? android.app.Activity)?.window
        val control = ventana?.let { WindowCompat.getInsetsController(it, vista) }
        if (aLaVista && control != null) control.isAppearanceLightStatusBars = !paleta.claroEncima && !oscuro
        onDispose { if (aLaVista && control != null) control.isAppearanceLightStatusBars = !oscuro }
    }
    val scroll = rememberScrollState()
    var altoVidriera by remember { mutableIntStateOf(0) }
    val umbral = with(LocalDensity.current) { 140.dp.toPx() }
    val plegada by remember { derivedStateOf { altoVidriera > 0 && scroll.value > altoVidriera - umbral } }

    @Serializable data class SinLeer(val sinLeer: Int = 0)

    suspend fun cargar() {
        try {
            val r: Tablero = estado.api.get("/v1/panel/tablero", listOf("periodo" to periodo))
            carga = Carga.listo(r)
            Guardado.guardar(r, "tablero-$periodo")
            val a = r.acento
            val s = r.sobreAcento
            if (a != null && s != null) {
                colores = "$a|$s"
                Preferencias.ponerTexto("inicio.colores", colores)
            }
            if (r.paraHacer.preparar > 0) {
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

    Box(Modifier.fillMaxSize().grano(0.06f, BlendMode.Multiply)) {
        Refrescable({ cargar() }) {
            Column(Modifier.fillMaxSize().verticalScroll(scroll).padding(bottom = 32.dp)) {
                Vidriera(carga, periodo, { periodo = it }, sinLeer, paleta, se, Modifier.onSizeChanged { altoVidriera = it.height })
                when (val c = carga) {
                    Carga.cargando -> EsqueletoInicio(paleta)
                    is Carga.error -> Box(Modifier.padding(top = 30.dp).margenPantalla()) { EstadoError("cómo viene el día", c.e) { estado.tarea { cargar() } } }
                    is Carga.listo -> Contenido(c.v, periodo, masViejo, paleta, ideasOcultas, { ideasOcultas = ideasOcultas + it }, { editarMeta = true })
                }
            }
        }
        AnimatedVisibility(plegada && t != null, enter = slideInVertically { -it } + fadeIn(), exit = slideOutVertically { -it } + fadeOut()) {
            t?.let { BarraVidriera(it, periodo, paleta) }
        }
    }
    LaunchedEffect(periodo, estado.version) { cargar() }
    Hoja(editarMeta, { editarMeta = false }, altoFijo = 470.dp) {
        carga.valor?.let { v -> EditarMeta(v.meta?.objetivo ?: 0, v.meta?.actual ?: 0) { cargar() } }
    }
}

// MARK: Contenido

@Composable
private fun Contenido(t: Tablero, periodo: String, masViejo: Instant?, paleta: PaletaTienda, ocultas: Set<String>, ocultar: (String) -> Unit, editarMeta: () -> Unit) {
    val estado = LocalEstado.current
    val cero = Vidriera.enCero(t)
    if (!cero) ideaPrincipal(t, ocultas, ocultar)?.let { TarjetaTincho(it, paleta, Modifier.subir(28.dp).padding(horizontal = 16.dp)) }
    Column(Modifier.fillMaxWidth().padding(top = 24.dp).margenPantalla(), verticalArrangement = Arrangement.spacedBy(32.dp)) {
        if (cero) {
            AccionesCero(paleta)
            if (!PrimerosPasos.ocultos(estado.tienda?.id)) PasosPrimeraVenta(paleta)
        } else {
            AccionesInicio(t, paleta)
        }
        ParaHacerHoy(t, masViejo, paleta, cero)
        DeUnVistazo(t, periodo, paleta, editarMeta)
        if (t.ultimosPedidos.isNotEmpty()) UltimosPedidos(t, paleta)
        if (t.masVendidos.isNotEmpty()) LoMasVendido(t)
        t.visitas?.let { if (it.masVistos.isNotEmpty()) TarjetaMasVistos(it, paleta) }
        t.resenas?.let { if (it.nuevas.isNotEmpty()) TarjetaResenas(it, paleta) }
        t.envios?.let { if (it.paquetes > 0 || it.retiros.isNotEmpty()) TarjetaEnvios(it, paleta) }
        if (estado.puede("cuenta")) TarjetaIntegraciones(paleta)
    }
}

// MARK: Tincho

/** La idea más relevante de Tincho: carritos por recordar, después stock, después compartir; si no hay ninguna, lo más
 *  vendido o una invitación a preguntarle. */
@Composable
private fun ideaPrincipal(t: Tablero, ocultas: Set<String>, ocultar: (String) -> Unit): IdeaTincho? {
    val estado = LocalEstado.current
    val nav = LocalNavegacion.current

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
            add(IdeaTincho("carritos", s, "Sí, mandales", "¿Qué dejaron en los carritos abandonados?") { recordarCarritos() })
        }
        if (t.paraHacer.sinStock > 0 && "stock" !in t.hornero.ideasHechas) {
            val n = t.paraHacer.sinStock
            add(IdeaTincho("stock", AnnotatedString(if (n == 1) "Un producto se quedó sin stock y sigue visible en la tienda. ¿Le cargás un ingreso?"
                else "$n productos se quedaron sin stock y siguen visibles en la tienda. ¿Les cargás un ingreso?"), "Reponer", "¿Qué productos tengo que reponer?") {
                nav.segmentoProductos = "out"; nav.tocar(Pestana.productos)
            })
        }
        val v = t.visitas
        if (v != null && "compartir" !in t.hornero.ideasHechas) {
            val n = v.dias7.actual.toInt()
            if (v.desde == null || n < 30) {
                add(IdeaTincho("compartir", AnnotatedString(if (v.desde == null || n == 0) "Tu tienda todavía no tuvo visitas. ¿Te armo una imagen con el QR para tus estados de WhatsApp?"
                    else "En 7 días te visitaron ${if (n == 1) "1 persona" else "$n personas"}. Con una historia o un estado llegás a muchas más. ¿La armamos?"),
                    "Armarla", "¿Cómo consigo más visitas en mi tienda?") { nav.inicio.add(Destino.compartir) })
            }
        }
        t.masVendidos.firstOrNull()?.let { p ->
            val s = buildAnnotatedString {
                withStyle(SpanStyle(fontWeight = FontWeight.Bold)) { append(p.nombre) }
                append(" es lo que más sale: ${Formato.cantidad(p.unidades, "unidad", "unidades")}. Compartilo para que llegue a más gente.")
            }
            add(IdeaTincho("estrella", s, "Compartir", "¿Cómo viene ${p.nombre}?") { nav.inicio.add(Destino.compartir) })
        }
        add(IdeaTincho("preguntar", AnnotatedString("Preguntame lo que quieras de tu negocio: cuánto vendiste, qué reponer o cómo vienen tus clientes."),
            null, "¿Cómo vengo esta semana?") {})
    }
    return lista.firstOrNull { it.id !in ocultas }
}

/** Indicador chico (radio 18): título 13, valor 22/700 y delta 12/600 coloreado. Lo usan Clientes y las secciones de
 *  Más. */
@Composable
fun Indicador(titulo: String, valor: String, delta: String?, positivo: Boolean, modifier: Modifier = Modifier) {
    Column(modifier.fillMaxHeight().tarjeta(radio = 18.dp).padding(horizontal = 13.dp, vertical = 12.dp), verticalArrangement = Arrangement.spacedBy(2.dp)) {
        Texto(titulo, Estilo.footnote, C.texto2, lineas = 2)
        Texto(valor, Estilo.indicador, C.texto, lineas = 1, numeros = true, escalaMinima = 0.6f)
        if (delta != null) Texto(delta, Estilo.chip, if (positivo) C.okTinta else C.errorTinta, lineas = 1, escalaMinima = 0.8f)
    }
}
