package ar.com.tulocalweb.app.pantallas.mas

import android.net.Uri
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.PickVisualMediaRequest
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.imePadding
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
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateMapOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.runtime.setValue
import androidx.compose.runtime.staticCompositionLocalOf
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.nucleo.Api
import ar.com.tulocalweb.app.nucleo.CatalogoBloques
import ar.com.tulocalweb.app.nucleo.ContextoTienda
import ar.com.tulocalweb.app.nucleo.Fecha
import ar.com.tulocalweb.app.nucleo.Fotos
import ar.com.tulocalweb.app.nucleo.MetricaVentana
import ar.com.tulocalweb.app.nucleo.VentanaTienda
import ar.com.tulocalweb.app.nucleo.numero
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.PaginaProductos
import ar.com.tulocalweb.app.nucleo.ProductoFila
import ar.com.tulocalweb.app.nucleo.Valor
import ar.com.tulocalweb.app.nucleo.ValorJSON
import ar.com.tulocalweb.app.nucleo.Web
import ar.com.tulocalweb.app.nucleo.bool
import ar.com.tulocalweb.app.nucleo.entero
import ar.com.tulocalweb.app.nucleo.get
import ar.com.tulocalweb.app.nucleo.lista
import ar.com.tulocalweb.app.nucleo.objeto
import ar.com.tulocalweb.app.nucleo.opcional
import ar.com.tulocalweb.app.nucleo.texto
import ar.com.tulocalweb.app.nucleo.vacio
import ar.com.tulocalweb.app.tema.*
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import java.time.Instant
import java.time.LocalDate
import java.time.ZoneId
import java.time.ZonedDateTime
import java.time.format.DateTimeFormatter

// Clon de Pantallas/Mas/CamposBloque.swift.
// Los campos del panel de un bloque, generados desde bloques.json › campos (tiposCampo): un editor por tipo. Lo vacío
// se saca de los datos, así la tienda usa lo de por defecto o lo sugerido.

/** Lo que los campos necesitan de la tienda: productos para buscar, categorías y el contexto de los avisos. */
class DatosEditor {
    var contexto by mutableStateOf(ContextoTienda())
    /** Los primeros productos (para elegir sin buscar) y todos los que se fueron viendo, por id. */
    var productos by mutableStateOf<List<ProductoFila>>(emptyList())
    val conocidos = mutableStateMapOf<String, ProductoFila>()

    /** WhatsApp y dirección del local, para los avisos. Se vuelve a leer al volver de Datos del negocio. */
    suspend fun cargarNegocio(api: Api) {
        val n = api.opcional<ValorJSON>("/v1/panel/configuracion")?.get("negocio") ?: return
        contexto = contexto.copy(
            whatsapp = !(n["whatsapp"]?.texto ?: "").isBlank(),
            direccionLocal = !(n["direccion"]?.texto ?: "").isBlank(),
        )
    }

    suspend fun cargarProductos(api: Api) {
        val l = api.opcional<PaginaProductos>("/v1/panel/listas/productos", listOf("orden" to "sold", "cuantos" to "60")) ?: return
        productos = l.filas
        l.filas.forEach { conocidos[it.id] = it }
    }

    suspend fun buscar(texto: String, api: Api): List<ProductoFila> {
        val t = texto.trim()
        if (t.isEmpty()) return productos
        val l = api.opcional<PaginaProductos>("/v1/panel/listas/productos", listOf("q" to t, "cuantos" to "40"))
            ?: return productos.filter { it.nombre.contains(t, ignoreCase = true) }
        l.filas.forEach { conocidos[it.id] = it }
        return l.filas
    }

    fun nombre(id: String): String = conocidos[id]?.nombre ?: "Un producto"

    // MARK: Promociones

    /** Un cupón de la tienda, para elegirlo en una ventana. */
    data class CuponTienda(val codigo: String, val descripcion: String, val activo: Boolean)
    var cupones by mutableStateOf<List<CuponTienda>>(emptyList())
    /** Desde cuánto el envío es gratis (Cupones › automáticas), para las plantillas. */
    var envioGratis by mutableStateOf<Int?>(null)
    /** Por ventana, los últimos 30 días. */
    var metricas by mutableStateOf<Map<String, MetricaVentana>>(emptyMap())
    var metricasCargadas by mutableStateOf(false)

    suspend fun cargarPromos(api: Api) {
        val cs = api.opcional<ValorJSON>("/v1/panel/cupones")
        cupones = (cs?.get("cupones")?.lista ?: emptyList()).mapNotNull { x ->
            val codigo = x["codigo"]?.texto ?: return@mapNotNull null
            if (x["estado"]?.texto == "terminado") null else CuponTienda(codigo, x["descripcion"]?.texto ?: "", x["estado"]?.texto == "activo")
        }
        envioGratis = api.opcional<ValorJSON>("/v1/panel/configuracion/entregas")?.get("gratisDesde")?.entero
    }

    suspend fun cargarMetricas(api: Api) {
        val hoy = Instant.now()
        val desde = hoy.atZone(VentanaTienda.zonaAR).minusDays(29)
        api.opcional<ValorJSON>("/v1/panel/promos/metricas", listOf("desde" to VentanaTienda.isoAR(desde).take(10), "hasta" to VentanaTienda.isoAR(hoy).take(10)))
            ?.let { metricas = MetricaVentana.leer(it) }
        metricasCargadas = true
    }
}

val LocalDatosEditor = staticCompositionLocalOf { DatosEditor() }

/** Todos los campos de un bloque (o de un elemento de una lista). */
@Composable
fun EditorCampos(campos: List<CatalogoBloques.Campo>, datos: Map<String, ValorJSON>, alCambiar: (Map<String, ValorJSON>) -> Unit) {
    val actuales by rememberUpdatedState(datos)
    Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
        campos.forEach { c ->
            Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
                EditorCampo(c, datos[c.id]) { v ->
                    val d = actuales.toMutableMap()
                    if (v != null && (!v.vacio || v.bool != null)) d[c.id] = v else d.remove(c.id)
                    alCambiar(d)
                }
                val d = c.descripcion
                if (d != null && c.tipo != "grupo") Texto(d, Estilo.footnote, C.texto2)
            }
        }
    }
}

/** Un campo: elige el editor por su tipo. */
@Composable
fun EditorCampo(campo: CatalogoBloques.Campo, valor: ValorJSON?, alCambiar: (ValorJSON?) -> Unit) {
    val datos = LocalDatosEditor.current
    val cat = CatalogoBloques.shared
    val texto = valor?.texto ?: ""
    fun ponerTexto(v: String) = alCambiar(Valor.texto(campo.max?.let { v.take(it) } ?: v))
    /** Lo que se ve si se deja vacío. */
    val sugerencia = campo.porDefecto?.texto ?: when (campo.sugerido) {
        "nombre" -> datos.contexto.nombre.ifEmpty { "El nombre de tu tienda" }
        "lema" -> "La frase de tu tienda"
        else -> ""
    }
    when (campo.tipo) {
        "texto" -> CampoLinea(campo.nombre, texto, ::ponerTexto, sugerencia, campo.max)
        "textoLargo" -> CampoLinea(campo.nombre, texto, ::ponerTexto, sugerencia, campo.max, largo = true)
        "video" -> Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
            CampoLinea(campo.nombre, texto, ::ponerTexto, "https://youtube.com/…", teclado = KeyboardType.Uri)
            if (texto.isNotEmpty() && !esVideo(texto)) Texto("Pegá un link de YouTube o de Vimeo.", Estilo.footnote, C.errorTinta)
        }
        "foto" -> SubidorFotos(campo.nombre, 1, listOfNotNull(valor?.texto)) { alCambiar(it.firstOrNull()?.let { u -> Valor.texto(u) }) }
        "fotos" -> SubidorFotos(campo.nombre, campo.max ?: 12, valor?.lista?.mapNotNull { it.texto } ?: emptyList()) { alCambiar(Valor.lista(it.map { u -> Valor.texto(u) })) }
        "link" -> SelectorLink(campo.nombre, texto, ::ponerTexto)
        "boton" -> EditorBoton(campo.nombre, valor, alCambiar)
        "fondo" -> SelectorOpcion(campo.nombre, cat.fondos.opciones, cat.fondos.nombres, valor?.texto ?: campo.porDefecto?.texto ?: "normal") { alCambiar(Valor.texto(it)) }
        "opcion" -> SelectorOpcion(campo.nombre, campo.opciones ?: emptyList(), campo.nombres ?: campo.opciones ?: emptyList(),
            valor?.texto ?: campo.porDefecto?.texto ?: campo.opciones?.firstOrNull() ?: "") { alCambiar(Valor.texto(it)) }
        "fecha" -> EditorFecha(campo.nombre, valor, alCambiar)
        "numero" -> {
            val n = valor?.entero ?: campo.porDefecto?.entero ?: campo.min ?: 0
            Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
                Texto(campo.nombre, Estilo.body16, C.texto, Modifier.weight(1f))
                Texto("$n", Estilo.fila, C.texto, Modifier.padding(end = 10.dp), numeros = true)
                Pasador(n, { alCambiar(Valor.numero(it)) }, (campo.min ?: 0)..(campo.max ?: 99))
            }
        }
        "si" -> Interruptor(campo.nombre, valor?.bool ?: campo.porDefecto?.bool ?: false, { alCambiar(Valor.booleano(it)) })
        "icono" -> SelectorIcono(campo.nombre, texto, ::ponerTexto)
        "producto" -> SelectorProductos(campo.nombre, max = 1, ids = listOfNotNull(valor?.texto)) { alCambiar(it.firstOrNull()?.let { id -> Valor.texto(id) }) }
        "elegidos" -> SelectorProductos(campo.nombre, campo.min ?: 0, campo.max ?: 12, valor?.lista?.mapNotNull { it.texto } ?: emptyList()) {
            alCambiar(Valor.lista(it.map { id -> Valor.texto(id) }))
        }
        "productos" -> EditorFuenteProductos(campo, valor, alCambiar)
        "categoria" -> SelectorCategoria(campo.nombre, texto, ::ponerTexto)
        "categorias" -> SelectorCategorias(campo.nombre, valor?.lista?.mapNotNull { it.texto } ?: emptyList()) { alCambiar(Valor.lista(it.map { c -> Valor.texto(c) })) }
        "lista" -> EditorLista(campo, valor, alCambiar)
        "grupo" -> Column(Modifier.fillMaxWidth().border(1.dp, C.linea2, RoundedCornerShape(14.dp)).padding(12.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
            Texto(campo.nombre, Estilo.subheadFuerte, C.texto)
            campo.descripcion?.let { Texto(it, Estilo.footnote, C.texto2) }
            EditorCampos(camposDelGrupo(campo, valor), valor?.objeto ?: emptyMap()) { alCambiar(Valor.objeto(it)) }
        }
        "tabla" -> EditorTabla(campo.nombre, valor, alCambiar)
        "cupon" -> SelectorCupon(campo.nombre, texto, ::ponerTexto)
        "premio" -> EditorPremio(campo, valor, alCambiar)
        "opciones" -> SelectorOpciones(campo.nombre, campo.opciones ?: emptyList(), campo.nombres ?: campo.opciones ?: emptyList(),
            valor?.lista?.mapNotNull { it.texto } ?: emptyList()) { alCambiar(Valor.lista(it.map { o -> Valor.texto(o) })) }
        else -> {}
    }
}

/** Un grupo que empieza con «activa» apagada muestra sólo el interruptor (compras recientes). */
private fun camposDelGrupo(campo: CatalogoBloques.Campo, valor: ValorJSON?): List<CatalogoBloques.Campo> {
    val items = campo.item ?: emptyList()
    val p = items.firstOrNull() ?: return items
    if (p.id != "activa" || p.tipo != "si") return items
    val on = valor?.get("activa")?.bool ?: campo.porDefecto?.get("activa")?.bool ?: p.porDefecto?.bool ?: false
    return if (on) items else listOf(p)
}

fun esVideo(s: String): Boolean {
    val u = runCatching { Uri.parse(s) }.getOrNull() ?: return false
    if (u.scheme != "https") return false
    return u.host?.lowercase() in listOf("youtube.com", "www.youtube.com", "m.youtube.com", "youtu.be", "vimeo.com", "www.vimeo.com", "player.vimeo.com")
}

// MARK: - Texto

/** Un texto con su título arriba y, de placeholder, lo que se ve si queda vacío. */
@Composable
fun CampoLinea(titulo: String, texto: String, alCambiar: (String) -> Unit, sugerencia: String = "", max: Int? = null, largo: Boolean = false,
               teclado: KeyboardType = KeyboardType.Text) {
    val forma = RoundedCornerShape(14.dp)
    Column(Modifier.fillMaxWidth(), verticalArrangement = Arrangement.spacedBy(6.dp)) {
        Row {
            Texto(titulo, Estilo.footnoteFuerte, C.texto2, Modifier.weight(1f))
            if (max != null && texto.length > max * 3 / 4) Texto("${texto.length}/$max", Estilo.footnote, C.texto3, numeros = true)
        }
        Entrada(texto, alCambiar, sugerencia.ifEmpty { titulo },
            Modifier.fillMaxWidth().clip(forma).background(C.superficie).border(1.dp, C.linea2, forma).padding(horizontal = 14.dp, vertical = 13.dp),
            estilo = Estilo.body16, teclado = teclado, sinCorrector = teclado == KeyboardType.Uri,
            lineas = if (largo) 3 else 1, lineasMaximas = if (largo) 8 else 1)
    }
}

// MARK: - Fotos

/** Una o varias fotos: se eligen de la galería, se suben con POST /panel/fotos y se pueden quitar. */
@Composable
fun SubidorFotos(titulo: String, max: Int, fotos: List<String>, alCambiar: (List<String>) -> Unit) {
    val estado = LocalEstado.current
    val ctx = LocalContext.current
    val alcance = rememberCoroutineScope()
    var subiendo by remember { mutableIntStateOf(0) }
    val actuales by rememberUpdatedState(fotos)
    val cambiar by rememberUpdatedState(alCambiar)
    fun subir(items: List<Uri>) {
        if (items.isEmpty()) return
        subiendo += items.size
        alcance.launch {
            for (u in items) {
                try {
                    val jpg = Fotos.achicar(ctx, u, 1600, 85) ?: continue
                    val url = Web.subirFoto(jpg, estado.api.token)
                    cambiar(if (actuales.size < max) actuales + url else listOf(url))
                } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
                    estado.avisar("No pudimos subir la foto", ok = false)
                } finally { subiendo -= 1 }
            }
        }
    }
    val una = rememberLauncherForActivityResult(ActivityResultContracts.PickVisualMedia()) { u -> subir(listOfNotNull(u)) }
    val varias = rememberLauncherForActivityResult(ActivityResultContracts.PickMultipleVisualMedia(maxOf(2, max))) { l -> subir(l.take(maxOf(0, max - actuales.size))) }
    Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
        Row {
            Texto(titulo, Estilo.footnoteFuerte, C.texto2, Modifier.weight(1f))
            if (max > 1) Texto("${fotos.size} de $max", Estilo.footnote, C.texto3, numeros = true)
        }
        Row(Modifier.fillMaxWidth().horizontalScroll(rememberScrollState()), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            fotos.forEachIndexed { i, f ->
                Box(Modifier.size(76.dp)) {
                    Miniatura(f, 76.dp, radio = 12.dp)
                    Box(Modifier.align(Alignment.TopEnd).padding(4.dp).size(22.dp).clip(CircleShape).background(Color.Black.copy(alpha = 0.6f))
                        .clickable { cambiar(fotos.filterIndexed { j, _ -> j != i }) }.semantics { contentDescription = "Quitar la foto ${i + 1}" },
                        contentAlignment = Alignment.Center) { Icono("xmark", 10f, Color.White) }
                }
            }
            repeat(subiendo) { Box(Modifier.size(76.dp).clip(RoundedCornerShape(12.dp)).background(C.crema2), contentAlignment = Alignment.Center) { Rueda(C.acento) } }
            if (fotos.size + subiendo < max) {
                Column(Modifier.size(76.dp).clip(RoundedCornerShape(12.dp)).background(C.crema2).clickable {
                    val pedido = PickVisualMediaRequest(ActivityResultContracts.PickVisualMedia.ImageOnly)
                    if (max - fotos.size - subiendo > 1) varias.launch(pedido) else una.launch(pedido)
                    estado.abrioOtraApp = true
                }, horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(4.dp, Alignment.CenterVertically)) {
                    Icono("photo.badge.plus", 20f, C.primarioTinta)
                    Texto(if (fotos.isEmpty()) "Elegir" else "Sumar", Estilo.footnoteFuerte, C.primarioTinta)
                }
            }
        }
    }
}

// MARK: - Links y botones

/** Adónde lleva un link: una página, una categoría, un producto o un link propio (tiposCampo › link). */
@Composable
fun SelectorLink(titulo: String, link: String, alCambiar: (String) -> Unit) {
    val datos = LocalDatosEditor.current
    var buscando by remember { mutableStateOf(false) }
    val paginas = CatalogoEstilo.compartido.paginas
    val nombre = when {
        link.isEmpty() -> "Sin link"
        paginas.any { it.id == link } -> paginas.first { it.id == link }.nombre
        link.startsWith("cat:") -> "Categoría: ${link.drop(4)}"
        link.startsWith("prod:") -> "Producto: ${datos.nombre(link.drop(5))}"
        else -> "Un link propio"
    }
    val forma = RoundedCornerShape(14.dp)
    Column(Modifier.fillMaxWidth(), verticalArrangement = Arrangement.spacedBy(6.dp)) {
        Texto(titulo, Estilo.footnoteFuerte, C.texto2)
        MenuBoton(
            listOf(OpcionDeMenu("Sin link", "") { alCambiar("") }) +
                paginas.map { p -> OpcionDeMenu(p.nombre, "") { alCambiar(p.id) } } +
                datos.contexto.categorias.map { c -> OpcionDeMenu(c, "tag") { alCambiar("cat:$c") } } +
                listOf(OpcionDeMenu("Un producto…", "magnifyingglass") { buscando = true }, OpcionDeMenu("Un link propio…", "link") { alCambiar("https://") }),
            titulo,
        ) {
            Row(Modifier.fillMaxWidth().height(48.dp).clip(forma).background(C.superficie).border(1.dp, C.linea2, forma).padding(horizontal = 14.dp),
                verticalAlignment = Alignment.CenterVertically) {
                Texto(nombre, Estilo.body16, if (link.isEmpty()) C.texto3 else C.texto, Modifier.weight(1f), lineas = 1)
                Icono("chevron.up.chevron.down", 12f, C.texto2)
            }
        }
        if (link.startsWith("http")) {
            CampoLinea("Link", link, { alCambiar(it.trim()) }, "https://", teclado = KeyboardType.Uri)
            if (!link.startsWith("https://")) Texto("Tiene que empezar con https://", Estilo.footnote, C.errorTinta)
        }
    }
    Hoja(buscando, { buscando = false }) {
        BuscadorProductos(1, if (link.startsWith("prod:")) listOf(link.drop(5)) else emptyList(), { ids -> ids.firstOrNull()?.let { alCambiar("prod:$it") } }) { buscando = false }
    }
}

/** Un botón: el texto (hasta 30) y adónde lleva. */
@Composable
fun EditorBoton(titulo: String, valor: ValorJSON?, alCambiar: (ValorJSON?) -> Unit) {
    fun poner(k: String, v: String) {
        val o = (valor?.objeto ?: emptyMap()).toMutableMap()
        if (v.isEmpty()) o.remove(k) else o[k] = Valor.texto(if (k == "t") v.take(30) else v)
        alCambiar(if (o.isEmpty()) null else Valor.objeto(o))
    }
    Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
        CampoLinea("$titulo: texto", valor?.get("t")?.texto ?: "", { poner("t", it) }, "Ver productos", 30)
        if (!(valor?.get("t")?.texto ?: "").isEmpty()) SelectorLink("Adónde lleva", valor?.get("a")?.texto ?: "") { poner("a", it) }
    }
}

// MARK: - Opciones, fecha e íconos

/** Una opción entre pocas: segmentado si son hasta tres cortas, menú si no. */
@Composable
fun SelectorOpcion(titulo: String, opciones: List<String>, nombres: List<String>, elegido: String, alElegir: (String) -> Unit) {
    val pares = opciones.mapIndexed { i, o -> o to (nombres.getOrNull(i) ?: o) }
    Column(Modifier.fillMaxWidth(), verticalArrangement = Arrangement.spacedBy(6.dp)) {
        Texto(titulo, Estilo.footnoteFuerte, C.texto2)
        if (opciones.size <= 3 && nombres.all { it.length <= 14 }) Segmentado(pares, elegido, alElegir)
        else SelectorMenu(pares, elegido, alElegir, titulo)
    }
}

private val zonaAR: ZoneId = ZoneId.of("America/Argentina/Buenos_Aires")
private fun isoAR(f: ZonedDateTime): String = f.withZoneSameInstant(zonaAR).withNano(0).format(DateTimeFormatter.ISO_OFFSET_DATE_TIME)

/** Fecha y hora (ISO 8601 con la hora de Argentina). */
@Composable
fun EditorFecha(titulo: String, valor: ValorJSON?, alCambiar: (ValorJSON?) -> Unit, boton: String = "Elegir cuándo termina",
                pasada: String = "Ya pasó: no se muestra.", /** Con «Quitar» para dejarla sin fecha. */ quitable: Boolean = false) {
    val f = valor?.texto?.let { Fecha.iso(it) }?.atZone(zonaAR)
    Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
        if (f != null) {
            Column(Modifier.fillMaxWidth().clip(RoundedCornerShape(14.dp)).background(C.superficie)) {
                RenglonFecha(titulo, f.toLocalDate(), { d -> alCambiar(Valor.texto(isoAR(d.atTime(f.toLocalTime()).atZone(zonaAR)))) })
                RenglonHora("Hora", f.hour * 60 + f.minute) { m -> alCambiar(Valor.texto(isoAR(f.withHour(m / 60).withMinute(m % 60).withSecond(0)))) }
            }
            Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
                if (f.toInstant().isBefore(Instant.now())) Texto(pasada, Estilo.footnote, C.errorTinta)
                Spacer(Modifier.weight(1f))
                if (quitable) Texto("Quitar", Estilo.footnoteFuerte, C.primarioTexto, Modifier.clickable { alCambiar(null) })
            }
        } else {
            Texto(titulo, Estilo.footnoteFuerte, C.texto2)
            BotonTonal(boton, {
                val enUnaSemana = LocalDate.now(zonaAR).plusDays(7).atStartOfDay(zonaAR).minusMinutes(1)
                alCambiar(Valor.texto(isoAR(enUnaSemana)))
            }, alto = 40.dp, simbolo = "calendar")
        }
    }
}

/** Los íconos de bloques.json › iconos, con su símbolo y su nombre. */
@Composable
fun SelectorIcono(titulo: String, elegido: String, alElegir: (String) -> Unit) {
    Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
        Texto(titulo, Estilo.footnoteFuerte, C.texto2)
        CatalogoBloques.shared.iconos.chunked(5).forEach { fila ->
            Row(horizontalArrangement = Arrangement.spacedBy(6.dp)) {
                fila.forEach { i ->
                    val (simbolo, nombre) = simbolosIcono[i] ?: ("circle" to i)
                    val on = elegido == i
                    val forma = RoundedCornerShape(10.dp)
                    Column(Modifier.weight(1f).heightIn(min = 48.dp).clip(forma).background(if (on) C.texto else C.superficie).border(1.dp, C.linea2, forma)
                        .clickable { alElegir(i) }.semantics { contentDescription = nombre }.padding(vertical = 6.dp),
                        horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(3.dp, Alignment.CenterVertically)) {
                        Icono(simbolo, 17f, if (on) C.elevado else C.texto)
                        Texto(nombre, Estilo(9f, Peso.semibold), if (on) C.elevado else C.texto, lineas = 1, escalaMinima = 0.7f)
                    }
                }
                repeat(5 - fila.size) { Spacer(Modifier.weight(1f)) }
            }
        }
    }
}

val simbolosIcono = mapOf(
    "envio" to ("shippingbox" to "Envío"), "cuotas" to ("creditcard" to "Cuotas"), "transferencia" to ("building.columns" to "Transferencia"),
    "cambios" to ("arrow.triangle.2.circlepath" to "Cambios"), "retiro" to ("storefront" to "Retiro"), "seguro" to ("lock.shield" to "Compra segura"),
    "regalo" to ("gift" to "Regalo"), "hecho-a-mano" to ("hand.raised" to "Hecho a mano"), "reloj" to ("clock" to "Horario"),
    "estrella" to ("star" to "Destacado"),
)

// MARK: - Productos y categorías

/** Uno o varios productos de la tienda, con el buscador del catálogo. */
@Composable
fun SelectorProductos(titulo: String, min: Int = 0, max: Int, ids: List<String>, alCambiar: (List<String>) -> Unit) {
    val datos = LocalDatosEditor.current
    var buscando by remember { mutableStateOf(false) }
    Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
        Row {
            Texto(titulo, Estilo.footnoteFuerte, C.texto2, Modifier.weight(1f))
            if (max > 1) Texto(if (min > 0) "${ids.size} · de $min a $max" else "${ids.size} de $max", Estilo.footnote, C.texto3)
        }
        ids.forEach { id ->
            Row(horizontalArrangement = Arrangement.spacedBy(10.dp), verticalAlignment = Alignment.CenterVertically) {
                Miniatura(datos.conocidos[id]?.foto, 40.dp, radio = 10.dp)
                Texto(datos.nombre(id), Estilo.body16, C.texto, Modifier.weight(1f), lineas = 1)
                Box(Modifier.size(36.dp).clickable { alCambiar(ids.filter { it != id }) }.semantics { contentDescription = "Quitar ${datos.nombre(id)}" },
                    contentAlignment = Alignment.Center) { Icono("xmark.circle.fill", 18f, C.texto3) }
            }
        }
        if (ids.size < max || max == 1) {
            BotonTonal(if (ids.isEmpty()) (if (max == 1) "Elegir producto" else "Elegir productos") else (if (max == 1) "Cambiar" else "Sumar productos"),
                { buscando = true }, alto = 40.dp, simbolo = "magnifyingglass")
        }
    }
    Hoja(buscando, { buscando = false }) { BuscadorProductos(max, ids, alCambiar) { buscando = false } }
}

/** El catálogo de la tienda para elegir: buscador, foto, nombre y precio. */
@Composable
fun BuscadorProductos(max: Int, elegidos: List<String>, alCambiar: (List<String>) -> Unit, alCerrar: () -> Unit) {
    val estado = LocalEstado.current
    val datos = LocalDatosEditor.current
    var texto by remember { mutableStateOf("") }
    var lista by remember { mutableStateOf<List<ProductoFila>>(emptyList()) }
    var cargando by remember { mutableStateOf(true) }
    var marcados by remember { mutableStateOf(elegidos) }
    LaunchedEffect(texto) {
        if (texto.isNotEmpty()) delay(300)
        lista = datos.buscar(texto, estado.api)
        cargando = false
    }
    fun tocar(p: ProductoFila) {
        Haptico.seleccion()
        datos.conocidos[p.id] = p
        if (max == 1) { alCambiar(listOf(p.id)); alCerrar(); return }
        marcados = if (p.id in marcados) marcados - p.id else if (marcados.size < max) marcados + p.id else marcados
        alCambiar(marcados)
    }
    Column(Modifier.fillMaxSize().imePadding()) {
        Grabber()
        Box(Modifier.fillMaxWidth().height(44.dp).padding(horizontal = 16.dp)) {
            Texto(if (max == 1) "Elegí un producto" else "Elegí hasta $max", Estilo.headline, C.texto, Modifier.align(Alignment.Center))
            Texto("Listo", Estilo(17f, Peso.semibold), C.primarioTexto, Modifier.align(Alignment.CenterEnd).clickable(onClick = alCerrar))
        }
        Buscador("Buscá por nombre", texto, { texto = it }, Modifier.padding(horizontal = 20.dp, vertical = 8.dp))
        if (!cargando && lista.isEmpty()) {
            Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                Texto(if (texto.isEmpty()) "Todavía no cargaste productos." else "No encontramos «$texto».", Estilo.body16, C.texto2)
            }
        } else {
            Column(Modifier.weight(1f).verticalScroll(rememberScrollState())) {
                lista.forEach { p ->
                    Row(Modifier.fillMaxWidth().clickable { tocar(p) }.padding(vertical = 8.dp, horizontal = 20.dp),
                        horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
                        Miniatura(p.foto, 48.dp, radio = 12.dp)
                        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                            Texto(p.nombre, Estilo.fila, C.texto, lineas = 2)
                            Texto(Formato.pesos(p.precio), Estilo.texto14, C.texto2, numeros = true)
                        }
                        if (p.id in marcados) Icono("checkmark.circle.fill", 22f, C.ok)
                    }
                    Separador(80.dp)
                }
            }
        }
    }
}

/** Qué productos muestra un bloque: más vendidos, novedades, ofertas, elegidos o de una categoría. */
@Composable
fun EditorFuenteProductos(campo: CatalogoBloques.Campo, valor: ValorJSON?, alCambiar: (ValorJSON?) -> Unit) {
    val fuentes = CatalogoBloques.shared.fuentesProductos
    val actual = valor?.objeto ?: campo.porDefecto?.objeto ?: mapOf("fuente" to Valor.texto("vendidos"), "cantidad" to Valor.numero(8))
    fun poner(k: String, v: ValorJSON?) {
        val o = actual.toMutableMap()
        if (v == null) o.remove(k) else o[k] = v
        alCambiar(Valor.objeto(o))
    }
    val fuente = actual["fuente"]?.texto ?: "vendidos"
    val cantidad = actual["cantidad"]?.entero ?: 8
    Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
        SelectorOpcion(campo.nombre, fuentes.opciones, fuentes.nombres, fuente) { poner("fuente", Valor.texto(it)) }
        if (fuente == "elegidos") {
            SelectorProductos("Los productos", max = 24, ids = actual["ids"]?.lista?.mapNotNull { it.texto } ?: emptyList()) { ids ->
                poner("ids", if (ids.isEmpty()) null else Valor.lista(ids.map { Valor.texto(it) }))
            }
        } else {
            if (fuente == "categoria") SelectorCategoria("La categoría", actual["categoria"]?.texto ?: "") { poner("categoria", if (it.isEmpty()) null else Valor.texto(it)) }
            Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
                Texto("Cuántos", Estilo.body16, C.texto, Modifier.weight(1f))
                Texto("$cantidad", Estilo.fila, C.texto, Modifier.padding(end = 10.dp), numeros = true)
                Pasador(cantidad, { poner("cantidad", Valor.numero(it)) }, 1..24)
            }
        }
    }
}

@Composable
fun SelectorCategoria(titulo: String, elegida: String, alElegir: (String) -> Unit) {
    val datos = LocalDatosEditor.current
    Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
        Texto(titulo, Estilo.footnoteFuerte, C.texto2)
        if (datos.contexto.categorias.isEmpty()) Texto("Todavía no hay categorías: ponéselas a tus productos.", Estilo.footnote, C.texto2)
        else SelectorMenu(listOf("" to "Elegí una") + datos.contexto.categorias.map { it to it }, elegida, alElegir, titulo)
    }
}

/** Varias categorías (vacío: todas), como chips que se prenden. */
@Composable
fun SelectorCategorias(titulo: String, elegidas: List<String>, alCambiar: (List<String>) -> Unit) {
    val datos = LocalDatosEditor.current
    Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
        Texto(titulo, Estilo.footnoteFuerte, C.texto2)
        if (datos.contexto.categorias.isEmpty()) Texto("Todavía no hay categorías: ponéselas a tus productos.", Estilo.footnote, C.texto2)
        Row(Modifier.fillMaxWidth().horizontalScroll(rememberScrollState()), horizontalArrangement = Arrangement.spacedBy(6.dp)) {
            datos.contexto.categorias.forEach { c ->
                ChipFiltro(c, activo = c in elegidas) { alCambiar(if (c in elegidas) elegidas - c else elegidas + c) }
            }
        }
    }
}

// MARK: - Listas

/** Una lista de elementos con los campos de `item` (beneficios, preguntas, horarios, logos…). */
@Composable
fun EditorLista(campo: CatalogoBloques.Campo, valor: ValorJSON?, alCambiar: (ValorJSON?) -> Unit) {
    val items = valor?.lista?.map { it.objeto ?: emptyMap() } ?: emptyList()
    val max = campo.max ?: CatalogoBloques.shared.limites.itemsPorLista
    fun poner(l: List<Map<String, ValorJSON>>) = alCambiar(if (l.isEmpty()) null else Valor.lista(l.map { Valor.objeto(it) }))
    /** Un elemento nuevo con los números por defecto ya puestos (el peso de una porción cuenta aunque no se toque). */
    val nuevo = (campo.item ?: emptyList()).filter { it.tipo == "numero" }.mapNotNull { c -> c.porDefecto?.let { c.id to it } }.toMap()
    /** Con peso (las porciones de la ruleta): cuántas veces de cada 100 sale cada una. */
    fun chance(i: Int): String? {
        if ((campo.item ?: emptyList()).none { it.id == "peso" }) return null
        val pesos = items.map { it["peso"]?.numero ?: 0.0 }
        val total = pesos.sum()
        if (total <= 0 || i !in pesos.indices || pesos[i] <= 0) return "Nunca sale"
        return "Sale ${Math.round(pesos[i] / total * 100)} de cada 100"
    }
    Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
        Row {
            Texto(campo.nombre, Estilo.subheadFuerte, C.texto, Modifier.weight(1f))
            Texto("${items.size} de $max", Estilo.footnote, C.texto3, numeros = true)
        }
        items.forEachIndexed { i, item ->
            val forma = RoundedCornerShape(14.dp)
            Column(Modifier.fillMaxWidth().clip(forma).background(C.crema2.copy(alpha = 0.5f)).border(1.dp, C.linea, forma).padding(12.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp)) {
                Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(4.dp)) {
                    Texto("${i + 1}", Estilo.footnoteFuerte, C.texto2)
                    Texto(chance(i) ?: "", Estilo.footnote, C.texto3, Modifier.weight(1f).padding(start = 10.dp), lineas = 1, numeros = true)
                    Box(Modifier.size(36.dp).alpha(if (i == 0) 0.35f else 1f).clickable(enabled = i > 0) {
                        val l = items.toMutableList(); val x = l[i]; l[i] = l[i - 1]; l[i - 1] = x; poner(l)
                    }.semantics { contentDescription = "Subir" }, contentAlignment = Alignment.Center) { Icono("chevron.up", 14f, C.primarioTexto) }
                    Box(Modifier.size(36.dp).alpha(if (i == items.size - 1) 0.35f else 1f).clickable(enabled = i < items.size - 1) {
                        val l = items.toMutableList(); val x = l[i]; l[i] = l[i + 1]; l[i + 1] = x; poner(l)
                    }.semantics { contentDescription = "Bajar" }, contentAlignment = Alignment.Center) { Icono("chevron.down", 14f, C.primarioTexto) }
                    Box(Modifier.size(36.dp).clickable { poner(items.filterIndexed { j, _ -> j != i }) }.semantics { contentDescription = "Quitar" },
                        contentAlignment = Alignment.Center) { Icono("trash", 14f, C.primarioTexto) }
                }
                EditorCampos(campo.item ?: emptyList(), item) { v -> val l = items.toMutableList(); if (i in l.indices) { l[i] = v; poner(l) } }
            }
        }
        if (items.isEmpty()) Texto("Todavía no hay ninguno.", Estilo.footnote, C.texto2)
        if (items.size < max) BotonTonal("Sumar", { poner(items + listOf(nuevo)) }, alto = 40.dp, simbolo = "plus")
    }
}

// MARK: - Tabla

/** Una tabla (tiposCampo › tabla; la guía de talles): filas × columnas de textos cortos, la primera fila es el
 *  encabezado. Siempre queda pareja, con el encabezado y al menos una fila, dentro de los topes del catálogo. */
@Composable
fun EditorTabla(titulo: String, valor: ValorJSON?, alCambiar: (ValorJSON?) -> Unit) {
    val lim = CatalogoBloques.shared.limites
    val filas = valor?.lista?.map { f -> f.lista?.map { it.texto ?: "" } ?: emptyList() } ?: emptyList()
    val ancho = filas.firstOrNull()?.size ?: 0
    fun poner(f: List<List<String>>?) = alCambiar(f?.let { Valor.lista(it.map { r -> Valor.lista(r.map { c -> Valor.texto(c) }) }) })
    /** Una celda: en una línea, sin < ni >, hasta `letrasCelda` letras (como las cuenta Go). */
    fun limpia(s: String): String {
        val t = s.filter { it != '<' && it != '>' && it != '\n' && it != '\r' }
        val n = lim.letrasCelda
        return if (t.codePointCount(0, t.length) <= n) t else t.substring(0, t.offsetByCodePoints(0, n))
    }
    fun pista(i: Int, j: Int) = if (i == 0) (if (j == 0) "Talle" else "Medida") else (if (j == 0) "S" else "")
    Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
        Texto(titulo, Estilo.footnoteFuerte, C.texto2)
        if (ancho == 0) {
            BotonTonal("Armar la tabla", { poner(listOf(listOf("Talle", ""), listOf("", ""))) }, alto = 40.dp, simbolo = "square.grid.2x2")
        } else {
            val forma = RoundedCornerShape(12.dp)
            Column(Modifier.horizontalScroll(rememberScrollState()), verticalArrangement = Arrangement.spacedBy(6.dp)) {
                filas.forEachIndexed { i, fila ->
                    Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(6.dp)) {
                        for (j in 0 until ancho) {
                            val rotulo = if (i == 0) "Encabezado, columna ${j + 1}" else "Fila $i, columna ${j + 1}"
                            Entrada(fila.getOrElse(j) { "" }, { v ->
                                poner(filas.mapIndexed { a, f -> if (a == i) f.mapIndexed { b, x -> if (b == j) limpia(v) else x } else f })
                            }, pista(i, j),
                                Modifier.width(92.dp).heightIn(min = 44.dp).clip(forma).background(if (i == 0) C.crema2 else C.superficie)
                                    .border(1.dp, C.linea2, forma).padding(horizontal = 10.dp, vertical = 12.dp).semantics { contentDescription = rotulo },
                                estilo = if (i == 0) Estilo.footnoteFuerte else Estilo.body16)
                        }
                        if (i > 0) {
                            val puede = filas.size > 2
                            Box(Modifier.size(44.dp).alpha(if (puede) 1f else 0.35f).clickable(enabled = puede) { poner(filas.filterIndexed { k, _ -> k != i }) }
                                .semantics { contentDescription = "Quitar la fila $i" }, contentAlignment = Alignment.Center) { Icono("trash", 14f, C.primarioTexto) }
                        }
                    }
                }
            }
            Texto("La primera fila es el encabezado. Hasta ${lim.filasTabla} filas y ${lim.columnasTabla} columnas.", Estilo.footnote, C.texto2)
            Row(Modifier.horizontalScroll(rememberScrollState()), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                BotonTonal("Fila", { poner(filas + listOf(List(ancho) { "" })) }, Modifier.semantics { contentDescription = "Sumar una fila" }, alto = 40.dp, simbolo = "plus", habilitado = filas.size < lim.filasTabla)
                BotonTonal("Columna", { poner(filas.map { it + "" }) }, Modifier.semantics { contentDescription = "Sumar una columna" }, alto = 40.dp, simbolo = "plus", habilitado = ancho < lim.columnasTabla)
                BotonTonal("Columna", { poner(filas.map { it.dropLast(1) }) }, Modifier.semantics { contentDescription = "Quitar la última columna" }, alto = 40.dp, simbolo = "minus", habilitado = ancho > 1)
                BotonTonal("Borrar la tabla", { poner(null) }, alto = 40.dp, simbolo = "trash")
            }
        }
    }
}
// MARK: - Promociones: cupón, premio y opciones

/** Un cupón de la tienda (tiposCampo › cupon): se elige de la lista de Cupones. */
@Composable
fun SelectorCupon(titulo: String, codigo: String, alElegir: (String) -> Unit) {
    val datos = LocalDatosEditor.current
    Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
        Texto(titulo, Estilo.footnoteFuerte, C.texto2)
        if (datos.cupones.isEmpty() && codigo.isEmpty()) {
            Texto("Todavía no tenés cupones. Crealos en Más › Cupones.", Estilo.footnote, C.avisoTinta)
        } else {
            val opciones = listOf("" to "Elegí uno") + datos.cupones.map { c -> c.codigo to (if (c.descripcion.isEmpty()) c.codigo else "${c.codigo} · ${c.descripcion}") } +
                (if (codigo.isNotEmpty() && datos.cupones.none { it.codigo == codigo }) listOf(codigo to codigo) else emptyList())
            SelectorMenu(opciones, codigo, alElegir, titulo)
            if (datos.cupones.firstOrNull { it.codigo == codigo }?.activo == false) {
                Texto("Este cupón está pausado: activalo en Cupones para que funcione.", Estilo.footnote, C.avisoTinta)
            }
        }
    }
}

private val nombresPremio = mapOf("cuponFijo" to "Un cupón tuyo", "cuponUnico" to "Un código único", "descarga" to "Una descarga", "nada" to "Nada, sólo sumarse")
private val descuentosPremio = listOf("porcentaje" to "Porcentaje", "monto" to "Monto fijo", "envioGratis" to "Envío gratis")

/** El premio de una captación o de una porción de la ruleta (tiposCampo › premio). */
@Composable
fun EditorPremio(campo: CatalogoBloques.Campo, valor: ValorJSON?, alCambiar: (ValorJSON?) -> Unit) {
    val tipos = campo.opciones ?: listOf("cuponFijo", "cuponUnico", "descarga", "nada")
    val actual = valor?.objeto ?: campo.porDefecto?.objeto ?: emptyMap()
    val tipo = actual["tipo"]?.texto ?: ""
    fun poner(k: String, v: ValorJSON?) {
        val o = actual.toMutableMap()
        if (v == null) o.remove(k) else o[k] = v
        alCambiar(Valor.objeto(o))
    }
    /** Al cambiar de tipo quedan sólo las claves del nuevo (Go ignora las otras, pero así no confunden). */
    fun cambiarTipo(t: String) {
        val o = mutableMapOf<String, ValorJSON>("tipo" to Valor.texto(t))
        if (t == "cuponUnico") { o["descuento"] = Valor.texto("porcentaje"); o["valor"] = Valor.numero(10); o["dias"] = Valor.numero(7) }
        if (t == "descarga") actual["nombreArchivo"]?.let { o["nombreArchivo"] = it }
        alCambiar(Valor.objeto(o))
    }
    @Composable
    fun Numero(titulo: String, clave: String, rango: IntRange, paso: Int, porDefecto: Int, texto: (Int) -> String) {
        val n = actual[clave]?.entero ?: porDefecto
        Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
            Texto(titulo, Estilo.body16, C.texto, Modifier.weight(1f))
            Texto(texto(n), Estilo.fila, C.texto, Modifier.padding(end = 10.dp), numeros = true)
            Pasador(n, { poner(clave, Valor.numero(it.coerceIn(rango.first, rango.last))) }, rango, paso)
        }
    }
    Column(Modifier.fillMaxWidth().border(1.dp, C.linea2, RoundedCornerShape(14.dp)).padding(12.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
        SelectorOpcion(campo.nombre, tipos, tipos.map { nombresPremio[it] ?: it }, tipo) { cambiarTipo(it) }
        when (tipo) {
            "cuponFijo" -> SelectorCupon("El cupón", actual["cupon"]?.texto ?: "") { poner("cupon", if (it.isEmpty()) null else Valor.texto(it)) }
            "cuponUnico" -> {
                val d = actual["descuento"]?.texto ?: "porcentaje"
                SelectorOpcion("Qué descuento", descuentosPremio.map { it.first }, descuentosPremio.map { it.second }, d) { n ->
                    val o = actual.toMutableMap()
                    o["descuento"] = Valor.texto(n)
                    if (n == "envioGratis") o.remove("valor") else o["valor"] = Valor.numero(if (n == "monto") 2000 else 10)
                    alCambiar(Valor.objeto(o))
                }
                if (d == "porcentaje") Numero("Porcentaje", "valor", 1..90, 5, 10) { "$it %" }
                else if (d == "monto") Numero("Monto", "valor", 100..100_000, 500, 2000) { Formato.pesos(it) }
                Numero("Vence a los", "dias", 1..90, 1, 7) { if (it == 1) "1 día" else "$it días" }
                Texto("Cada persona recibe su propio código, de un solo uso.", Estilo.footnote, C.texto2)
            }
            "descarga" -> {
                val a = actual["archivo"]?.texto ?: ""
                CampoLinea("Link del archivo (PDF, guía, catálogo)", a, { v -> val t = v.trim(); poner("archivo", if (t.isEmpty()) null else Valor.texto(t.take(300))) },
                    "https://", teclado = KeyboardType.Uri)
                if (a.isNotEmpty() && !a.startsWith("https://")) Texto("Tiene que empezar con https://", Estilo.footnote, C.errorTinta)
                CampoLinea("Cómo se llama", actual["nombreArchivo"]?.texto ?: "", { v -> poner("nombreArchivo", if (v.isEmpty()) null else Valor.texto(v.take(60))) }, "Guía", 60)
            }
            "nada" -> Texto("Sólo se suma a tu newsletter.", Estilo.footnote, C.texto2)
        }
    }
}

/** Varias opciones de una lista (tiposCampo › opciones; vacía = todas), como chips que se prenden. */
@Composable
fun SelectorOpciones(titulo: String, opciones: List<String>, nombres: List<String>, elegidas: List<String>, alCambiar: (List<String>) -> Unit) {
    Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
        Row {
            Texto(titulo, Estilo.footnoteFuerte, C.texto2, Modifier.weight(1f))
            if (elegidas.isEmpty()) Texto("Todas", Estilo.footnote, C.texto3)
        }
        Row(Modifier.fillMaxWidth().horizontalScroll(rememberScrollState()), horizontalArrangement = Arrangement.spacedBy(6.dp)) {
            opciones.forEachIndexed { i, o ->
                ChipFiltro(nombres.getOrNull(i) ?: o, activo = o in elegidas) {
                    Haptico.seleccion()
                    // En el orden del catálogo, sin repetir.
                    val n = if (o in elegidas) elegidas - o else elegidas + o
                    alCambiar(opciones.filter { it in n })
                }
            }
        }
    }
}
