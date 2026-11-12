package ar.com.tulocalweb.app.pantallas.tincho

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.core.RepeatMode
import androidx.compose.animation.core.animateFloat
import androidx.compose.animation.core.infiniteRepeatable
import androidx.compose.animation.core.rememberInfiniteTransition
import androidx.compose.animation.core.tween
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.selection.SelectionContainer
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateListOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.ColorFilter
import androidx.compose.ui.graphics.ColorMatrix
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.AnnotatedString
import androidx.compose.ui.text.SpanStyle
import androidx.compose.ui.text.buildAnnotatedString
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.withStyle
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.app.Destino
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.app.LocalNavegacion
import ar.com.tulocalweb.app.app.Pestana
import ar.com.tulocalweb.app.nucleo.Tincho
import ar.com.tulocalweb.app.nucleo.Web
import ar.com.tulocalweb.app.nucleo.copiar
import ar.com.tulocalweb.app.tema.*
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch
import java.util.UUID

// Clon de Pantallas/Tincho/TinchoChat.swift.

private data class Mensaje(
    val deTincho: Boolean,
    val texto: String,
    val respuesta: RespuestaTincho? = null,
    /** Preguntas que se mandan solas (el resumen al abrir): no se muestran como burbuja. */
    val oculto: Boolean = false,
    val id: String = UUID.randomUUID().toString(),
)

private val iniciales = listOf("¿Cómo vengo hoy?", "¿Qué tengo pendiente?", "¿Qué tengo que reponer?", "¿Qué es lo que más se vende?", "Ideas para vender más", "¿Cómo creo un cupón?")

private val formaTincho = RoundedCornerShape(topStart = 20.dp, topEnd = 20.dp, bottomEnd = 20.dp, bottomStart = 6.dp)
private val formaPersona = RoundedCornerShape(topStart = 20.dp, topEnd = 20.dp, bottomEnd = 6.dp, bottomStart = 20.dp)

/** 6a · Chat con Tincho. Al abrir, cómo viene el día (cifras y pendientes). Cada respuesta trae el texto y, cuando
 *  corresponde, gráficos, tablas, listas, botones que llevan a la sección y preguntas para seguir. El motor vive en la
 *  web (/panel/hornero, formato «bloques»), que la app llama con su token; las preguntas de todos los días se responden
 *  con los datos, sin IA. */
@Composable
fun TinchoChat() {
    val estado = LocalEstado.current
    val nav = LocalNavegacion.current
    val ctx = LocalContext.current
    val alcance = rememberCoroutineScope()
    val mensajes = remember { mutableStateListOf<Mensaje>() }
    var texto by remember { mutableStateOf("") }
    var escribiendo by remember { mutableStateOf(false) }
    var sinConexion by remember { mutableStateOf(false) }
    var saludado by remember { mutableStateOf(false) }
    var tarea by remember { mutableStateOf<Job?>(null) }
    var ultimaPregunta by remember { mutableStateOf<String?>(null) }
    /** La respuesta nueva aparece de a poco, con cursor (6a): su id y cuántos caracteres van. */
    var revelando by remember { mutableStateOf<Pair<String, Int>?>(null) }
    val dictado = remember { Dictado(ctx, estado) }
    val lista = rememberLazyListState()
    val anima = animacionesActivas()

    val saludo = "¡Hola${estado.datos?.nombreDePila?.let { ", $it" } ?: ""}! Soy Tincho: conozco tus ventas, tu stock y tus clientes al día. Mirá cómo venís:"

    /** La respuesta aparece de a poco (de a palabras), y después sus gráficos y botones. */
    suspend fun revelar(m: Mensaje) {
        val total = (m.respuesta?.texto ?: m.texto).length
        if (total == 0 || !anima) return
        revelando = m.id to 0
        var n = 0
        while (n < total) {
            n = minOf(total, n + maxOf(3, total / 60))
            revelando = m.id to n
            delay(18)
        }
        revelando = null
    }

    /** Pide la respuesta con todo el historial (la última pregunta ya está en mensajes). */
    fun responder(pregunta: String) {
        val historial = mensajes.map { (if (it.deTincho) "hornero" else "persona") to (it.respuesta?.texto ?: it.texto) }
        ultimaPregunta = pregunta
        escribiendo = true
        sinConexion = false
        tarea = alcance.launch {
            try {
                val r = Web.tinchoConBloques(historial, estado.api.token)
                val m = Mensaje(true, r.texto, r)
                mensajes.add(m)
                escribiendo = false
                revelar(m)
            } catch (e: kotlinx.coroutines.CancellationException) {
                throw e
            } catch (e: Exception) {
                sinConexion = true
                escribiendo = false
            }
        }
    }

    fun preguntar(t: String, oculta: Boolean = false) {
        val pregunta = t.trim()
        if (pregunta.isEmpty() || escribiendo) return
        dictado.parar()
        if (!oculta) { texto = ""; Haptico.seleccion() }
        mensajes.add(Mensaje(false, pregunta, oculto = oculta))
        responder(pregunta)
    }

    fun detener() {
        tarea?.cancel()
        tarea = null
        escribiendo = false
        revelando = null
        Haptico.seleccion()
    }

    fun reintentar(p: String) { if (!escribiendo) responder(p) }

    /** Los botones de Tincho llevan a la pantalla de la app. */
    /** Manda la pregunta y la respuesta a seguridad@ por mail (la API no tiene dónde recibirlo). Sin app de mail, se
     *  copia el texto y se avisa a qué casilla mandarlo. */
    fun reportar(m: Mensaje) {
        Haptico.seleccion()
        val i = mensajes.indexOfFirst { it.id == m.id }.takeIf { it >= 0 } ?: mensajes.size
        val pregunta = mensajes.take(i).lastOrNull { !it.deTincho }?.texto.orEmpty()
        val cuerpo = cuerpoReporte(pregunta, m.respuesta?.texto ?: m.texto, estado.tienda?.id)
        val mail = android.content.Intent(android.content.Intent.ACTION_SENDTO, android.net.Uri.parse(mailReporte(cuerpo)))
            .addFlags(android.content.Intent.FLAG_ACTIVITY_NEW_TASK)
        if (runCatching { ctx.startActivity(mail) }.isFailure) {
            copiar(ctx, cuerpo)
            estado.avisar("Copiamos la respuesta. Mandala a $CASILLA_REPORTE para que la revisemos.")
        }
    }

    fun ir(destino: String) {
        Haptico.seleccion()
        when (destino) {
            "pedidos" -> { nav.filtroPedidos = null; nav.tocar(Pestana.pedidos) }
            "preparar" -> { nav.filtroPedidos = "Para preparar"; nav.pestana = Pestana.pedidos }
            "transferencias" -> { nav.filtroPedidos = "Sin pagar"; nav.pestana = Pestana.pedidos }
            "productos" -> nav.pestana = Pestana.productos
            "sin-stock" -> { nav.segmentoProductos = "out"; nav.pestana = Pestana.productos }
            "venta" -> nav.venta = true
            else -> {
                val d: Destino? = when (destino) {
                    "nuevo-producto" -> Destino.nuevoProducto(null)
                    "importar" -> Destino.importar
                    "clientes" -> Destino.clientes
                    "cupones" -> Destino.cupones
                    "carritos" -> Destino.carritos
                    "compartir" -> Destino.compartir
                    "estadisticas" -> Destino.estadisticas
                    "integraciones" -> Destino.integraciones
                    "mi-tienda" -> Destino.miTienda
                    "resenas" -> Destino.resenas
                    "negocio" -> Destino.negocio
                    "caja" -> Destino.caja
                    "finanzas" -> Destino.finanzas
                    "facturacion" -> Destino.facturacion
                    "equipo" -> Destino.usuarios
                    "plan" -> Destino.plan
                    "links" -> Destino.links
                    "devoluciones" -> Destino.devoluciones
                    else -> null
                }
                d?.let { nav.tincho.add(it) }
            }
        }
    }

    LaunchedEffect(Unit) {
        if (saludado) return@LaunchedEffect
        saludado = true
        val p = nav.preguntaTincho
        if (p != null) { nav.preguntaTincho = null; preguntar(p) } else preguntar("¿Cómo vengo hoy?", oculta = true)
    }
    LaunchedEffect(nav.preguntaTincho) {
        val p = nav.preguntaTincho ?: return@LaunchedEffect
        if (!saludado) return@LaunchedEffect
        nav.preguntaTincho = null
        while (escribiendo) delay(200)
        preguntar(p)
    }
    LaunchedEffect(dictado.texto) { if (dictado.escuchando || dictado.texto.isNotEmpty()) texto = dictado.texto }
    LaunchedEffect(dictado.error) { dictado.error?.let { estado.avisar(it, ok = false) } }
    DisposableEffect(Unit) { onDispose { dictado.parar() } }
    val visibles = mensajes.filter { !it.oculto }
    LaunchedEffect(visibles.size) { if (visibles.isNotEmpty()) runCatching { lista.animateScrollToItem(visibles.size) } }
    LaunchedEffect(escribiendo) { if (escribiendo) runCatching { lista.animateScrollToItem(visibles.size + 1) } }

    Column(Modifier.fillMaxSize().background(C.fondo).imePadding()) {
        // Encabezado
        Box(Modifier.fillMaxWidth()) {
            Row(Modifier.fillMaxWidth().padding(horizontal = 16.dp, vertical = 8.dp), horizontalArrangement = Arrangement.spacedBy(10.dp), verticalAlignment = Alignment.CenterVertically) {
                androidx.compose.foundation.Image(painterResource(Tincho.avatar.imagen), null, Modifier.size(38.dp).clip(CircleShape).background(C.crema),
                    contentScale = ContentScale.Crop, colorFilter = if (sinConexion) ColorFilter.colorMatrix(ColorMatrix().apply { setToSaturation(0f) }) else null)
                Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(1.dp)) {
                    Texto("Tincho", Estilo.headline, C.texto)
                    Row(horizontalArrangement = Arrangement.spacedBy(5.dp), verticalAlignment = Alignment.CenterVertically) {
                        Box(Modifier.size(6.dp).clip(CircleShape).background(if (sinConexion) C.errorTinta else C.ok))
                        Texto(if (sinConexion) "Sin conexión" else "Conoce tu tienda al día", Estilo.footnote, if (sinConexion) C.errorTinta else C.texto2)
                    }
                }
                if (mensajes.any { !it.oculto }) {
                    // Queda el resumen del día (la pregunta oculta y su respuesta).
                    Box(Modifier.size(40.dp).clickable { val dos = mensajes.take(2); mensajes.clear(); mensajes.addAll(dos) }.semantics { contentDescription = "Charla nueva" },
                        contentAlignment = Alignment.Center) { Icono("square.and.pencil", 17f, C.primarioTexto) }
                }
            }
            Box(Modifier.align(Alignment.BottomCenter).fillMaxWidth().height(0.5.dp).background(C.linea))
        }
        LazyColumn(Modifier.weight(1f).fillMaxWidth(), state = lista, contentPadding = PaddingValues(16.dp), verticalArrangement = Arrangement.spacedBy(14.dp)) {
            item { BurbujaTincho(Mensaje(true, saludo), revelando, ::ir, ::reportar) }
            items(visibles, key = { it.id }) { m ->
                Box(Modifier.animateItem()) { if (m.deTincho) BurbujaTincho(m, revelando, ::ir, ::reportar) else BurbujaPersona(m) }
            }
            if (escribiendo) item(key = "escribiendo") { EscribiendoAhora() }
            if (sinConexion && !escribiendo) item(key = "desenchufado") {
                Column(Modifier.fillMaxWidth().padding(vertical = 12.dp), horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(10.dp)) {
                    ImagenTincho(Tincho.desenchufado, Modifier.height(150.dp))
                    Texto("Sin conexión", Estilo.headline, C.errorTinta)
                    Texto("No pude responderte. Revisá la conexión y probá de nuevo.", Estilo.subhead, C.texto2, alinear = TextAlign.Center)
                    ultimaPregunta?.let { p -> BotonTonal("Reintentar", { reintentar(p) }, alto = 40.dp) }
                }
            }
        }
        AnimatedVisibility(!escribiendo && !sinConexion, enter = fadeIn(), exit = fadeOut()) {
            // 6a: las sugerencias en chips arriba de la entrada, a mano del pulgar.
            val ultimas = mensajes.lastOrNull { it.deTincho }?.respuesta?.sugerencias ?: emptyList()
            LazyRow(Modifier.fillMaxWidth().background(C.elevado).padding(top = 8.dp), contentPadding = PaddingValues(horizontal = 16.dp), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                items(ultimas.ifEmpty { iniciales }) { s ->
                    Texto(s, Estilo.footnoteFuerte, C.primarioTinta, Modifier.clip(Capsula).background(C.crema2).border(1.dp, C.linea2, Capsula)
                        .clickable { preguntar(s) }.padding(horizontal = 12.dp, vertical = 8.dp), lineas = 1)
                }
            }
        }
        // Entrada
        Row(Modifier.fillMaxWidth().background(C.elevado).padding(horizontal = 16.dp, vertical = 8.dp), horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.CenterVertically) {
            val forma = RoundedCornerShape(22.dp)
            Entrada(texto, { texto = it }, "Preguntale a Tincho", Modifier.weight(1f).clip(forma).background(C.superficie).border(1.dp, C.linea2, forma)
                .padding(horizontal = 16.dp, vertical = 11.dp), estilo = Estilo.body16, lineas = 1, lineasMaximas = 4, accionTeclado = ImeAction.Send, alEnviar = { preguntar(texto) })
            when {
                escribiendo -> BotonRedondoChat("stop.fill", "Detener", true) { detener() }
                texto.trim().isEmpty() || dictado.escuchando -> BotonRedondoChat(if (dictado.escuchando) "waveform" else "mic.fill",
                    if (dictado.escuchando) "Dejar de dictar" else "Dictar", dictado.escuchando, pulsar = dictado.escuchando) {
                    Haptico.seleccion(); alcance.launch { dictado.alternar() }
                }
                else -> BotonRedondoChat("arrow.up", "Enviar", true) { preguntar(texto) }
            }
        }
    }
}

@Composable
private fun BotonRedondoChat(simbolo: String, etiqueta: String, lleno: Boolean, pulsar: Boolean = false, accion: () -> Unit) {
    val a = if (pulsar) rememberInfiniteTransition(label = "onda").animateFloat(0.5f, 1f, infiniteRepeatable(tween(500), RepeatMode.Reverse), label = "a").value else 1f
    Box(Modifier.size(44.dp).clip(CircleShape).then(if (lleno) Modifier.background(C.primario) else Modifier.background(C.crema))
        .clickable(onClick = accion).semantics { contentDescription = etiqueta }, contentAlignment = Alignment.Center) {
        Icono(simbolo, 17f, if (lleno) Color.White else C.primarioTexto, Modifier.alpha(a))
    }
}

@Composable
private fun Avatar() {
    ImagenTincho(Tincho.avatar, Modifier.size(28.dp).clip(CircleShape).background(C.crema), ContentScale.Crop)
}

@Composable
private fun EscribiendoAhora() {
    // 6b: tres puntos accent con opacidades 1 / ,6 / ,3 que van rotando.
    val fase = rememberInfiniteTransition(label = "puntos").animateFloat(0f, 3f, infiniteRepeatable(tween(1200)), label = "f").value.toInt() % 3
    Row(horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.Bottom) {
        Avatar()
        Row(Modifier.clip(formaTincho).background(C.superficie).padding(horizontal = 16.dp, vertical = 14.dp).semantics { contentDescription = "Tincho está escribiendo" },
            horizontalArrangement = Arrangement.spacedBy(4.dp)) {
            repeat(3) { i -> Box(Modifier.size(7.dp).alpha(listOf(1f, 0.6f, 0.3f)[(fase + i) % 3]).clip(CircleShape).background(C.acento)) }
        }
    }
}

@Composable
private fun BurbujaTincho(m: Mensaje, revelando: Pair<String, Int>?, ir: (String) -> Unit, reportar: (Mensaje) -> Unit) {
    val texto = m.respuesta?.texto ?: m.texto
    Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
        if (texto.isNotEmpty()) {
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.Bottom) {
                Avatar()
                SelectionContainer(Modifier.weight(1f, fill = false)) {
                    val acento = C.acento
                    val t = if (revelando != null && revelando.first == m.id) buildAnnotatedString {
                        append(texto.take(revelando.second)); withStyle(SpanStyle(color = acento)) { append("▍") }
                    } else conFormato(texto)
                    TextoRico(t, Estilo.body16, C.texto, Modifier.clip(formaTincho).background(C.superficie).border(1.dp, C.anillo, formaTincho).padding(horizontal = 14.dp, vertical = 10.dp))
                }
                Spacer(Modifier.width(24.dp))
            }
        }
        val r = m.respuesta
        if (r != null && revelando?.first != m.id) {
            Column(Modifier.padding(start = 36.dp), verticalArrangement = Arrangement.spacedBy(10.dp)) {
                (r.bloques ?: emptyList()).forEach { VistaBloque(it) }
                val acciones = r.acciones
                if (!acciones.isNullOrEmpty()) {
                    Fluir {
                        acciones.forEach { a ->
                            Row(Modifier.height(36.dp).clip(Capsula).background(C.primario).clickable { ir(a.ir) }.padding(horizontal = 14.dp),
                                horizontalArrangement = Arrangement.spacedBy(6.dp), verticalAlignment = Alignment.CenterVertically) {
                                Texto(a.texto, Estilo.footnoteFuerte, Color.White)
                                Icono("arrow.right", 11f, Color.White)
                            }
                        }
                    }
                }
                // Las respuestas son de IA: cualquiera se puede reportar (políticas de Google y Apple).
                Row(Modifier.clip(Capsula).clickable { reportar(m) }.padding(vertical = 4.dp, horizontal = 2.dp),
                    horizontalArrangement = Arrangement.spacedBy(5.dp), verticalAlignment = Alignment.CenterVertically) {
                    Icono("flag", 12f, C.texto3)
                    Texto("Reportar respuesta", Estilo.footnote, C.texto3)
                }
            }
        }
    }
}

@Composable
private fun BurbujaPersona(m: Mensaje) {
    Row(Modifier.fillMaxWidth()) {
        Spacer(Modifier.weight(1f).widthIn(min = 40.dp))
        Texto(m.texto, Estilo.body16, Color.White, Modifier.widthIn(max = 260.dp).clip(formaPersona).background(C.primario).padding(horizontal = 14.dp, vertical = 10.dp))
    }
}

const val CASILLA_REPORTE = "seguridad@tulocalweb.com.ar"

/** El mail del reporte: la pregunta, la respuesta y la tienda (como cuerpoReporte en iOS). */
fun cuerpoReporte(pregunta: String, respuesta: String, tienda: String?): String {
    var t = "Quiero reportar esta respuesta de Tincho.\n\nMotivo: \n\n"
    if (pregunta.isNotEmpty()) t += "Pregunta: ${pregunta.take(500)}\n\n"
    t += "Respuesta: ${respuesta.take(1500)}"
    if (tienda != null) t += "\n\nTienda: $tienda"
    return t
}

fun mailReporte(cuerpo: String): String =
    "mailto:$CASILLA_REPORTE?subject=${android.net.Uri.encode("Reporte de una respuesta de Tincho")}&body=${android.net.Uri.encode(cuerpo)}"

/** **negritas** del texto de Tincho (sin tomar el resto como Markdown). */
fun conFormato(t: String): AnnotatedString = buildAnnotatedString {
    var i = 0
    while (i < t.length) {
        val ini = t.indexOf("**", i)
        if (ini < 0) { append(t.substring(i)); break }
        val fin = t.indexOf("**", ini + 2)
        if (fin < 0) { append(t.substring(i)); break }
        append(t.substring(i, ini))
        withStyle(SpanStyle(fontWeight = FontWeight.Bold)) { append(t.substring(ini + 2, fin)) }
        i = fin + 2
    }
}
