package ar.com.tulocalweb.app.pantallas.mas

import android.net.Uri
import android.os.SystemClock
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
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.selection.SelectionContainer
import androidx.compose.foundation.verticalScroll
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateListOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.nucleo.ErrorApi
import ar.com.tulocalweb.app.nucleo.Fotos
import ar.com.tulocalweb.app.nucleo.PaginaProductos
import ar.com.tulocalweb.app.nucleo.ProductoFila
import ar.com.tulocalweb.app.nucleo.Recursos
import ar.com.tulocalweb.app.nucleo.Tincho
import ar.com.tulocalweb.app.nucleo.ValorJSON
import ar.com.tulocalweb.app.nucleo.Web
import ar.com.tulocalweb.app.nucleo.abrirEnNavegador
import ar.com.tulocalweb.app.nucleo.bool
import ar.com.tulocalweb.app.nucleo.compartirTexto
import ar.com.tulocalweb.app.nucleo.get
import ar.com.tulocalweb.app.nucleo.lista
import ar.com.tulocalweb.app.nucleo.objeto
import ar.com.tulocalweb.app.nucleo.opcional
import ar.com.tulocalweb.app.nucleo.texto
import ar.com.tulocalweb.app.pantallas.compartir.Compartir
import ar.com.tulocalweb.app.tema.*
import coil3.compose.AsyncImage
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import kotlinx.serialization.Serializable
import kotlin.math.roundToInt

// Clon de Pantallas/Mas/MiTienda.swift.

/** Deshacer y rehacer: cada cambio guarda el diseño de antes. Lo que se escribe en un mismo campo seguido cuenta
 *  como un solo cambio. */
private class Historia {
    val atras = mutableStateListOf<DisenoTienda>()
    val adelante = mutableStateListOf<DisenoTienda>()
    private var clave: String? = null
    private var cuando = 0L

    fun antes(actual: DisenoTienda, nuevaClave: String?) {
        val ahora = SystemClock.uptimeMillis()
        val seguido = nuevaClave != null && nuevaClave == clave && ahora - cuando < 1500
        clave = nuevaClave; cuando = ahora
        adelante.clear()
        if (seguido) return
        atras.add(actual)
        if (atras.size > 60) atras.removeAt(0)
    }

    fun deshacer(actual: DisenoTienda): DisenoTienda? {
        val d = atras.removeLastOrNull() ?: return null
        adelante.add(actual); clave = null
        return d
    }

    fun rehacer(actual: DisenoTienda): DisenoTienda? {
        val d = adelante.removeLastOrNull() ?: return null
        atras.add(actual); clave = null
        return d
    }
}

/** Editar la tienda online desde el celular, igual que Mi tienda en el panel web: las páginas armadas con bloques
 *  (agregar, ordenar, ocultar, duplicar y borrar, con deshacer), el estilo (dirección, forma, colores, letra, barra,
 *  menú y redes), la marca, cómo se ve al compartir y la dirección. Se guarda como borrador, se mira en la vista previa
 *  real (que se actualiza sin recargar y abre el bloque que se toca) y se publica. */
@Composable
fun MiTiendaEditor() {
    val estado = LocalEstado.current
    val ctx = LocalContext.current
    val alcance = rememberCoroutineScope()
    var diseno by remember { mutableStateOf(DisenoTienda()) }
    val historia = remember { Historia() }
    var cargado by remember { mutableStateOf(false) }
    var direccion by remember { mutableStateOf("") }
    var slug by remember { mutableStateOf("") }
    var sinPublicar by remember { mutableStateOf(false) }
    var guardando by remember { mutableStateOf(false) }
    var subiendo by remember { mutableStateOf<String?>(null) }
    var categorias by remember { mutableStateOf<List<String>>(emptyList()) }
    var productos by remember { mutableStateOf<List<ProductoFila>>(emptyList()) }
    var hayResenas by remember { mutableStateOf(false) }
    var pestana by remember { mutableStateOf("paginas") }
    var pagina by remember { mutableStateOf("inicio") }
    var editando by remember { mutableStateOf<String?>(null) }
    var agregarEn by remember { mutableStateOf<Int?>(null) }
    var presetPedido by remember { mutableStateOf<String?>(null) }
    var previa by remember { mutableStateOf(false) }
    var vuelta by remember { mutableIntStateOf(0) }
    /** El diseño que tiene la vista previa (si la página no sabe dibujar sin recargar, se recarga al cambiar). */
    var enLaVista by remember { mutableStateOf<DisenoTienda?>(null) }
    val puente = remember { PuenteVista() }
    val plantillas = remember { PlantillaTienda.todas() }
    val cat = CatalogoBloques.compartido

    val urlTienda: String? = if (direccion.isEmpty()) estado.tienda?.url else (if (direccion.startsWith("http")) direccion else "https://$direccion")

    fun cambiar(nuevo: DisenoTienda, clave: String? = null) {
        if (nuevo == diseno) return
        historia.antes(diseno, clave)
        diseno = nuevo
    }
    fun deshacer() { historia.deshacer(diseno)?.let { diseno = it; Haptico.seleccion() } }
    fun rehacer() { historia.rehacer(diseno)?.let { diseno = it; Haptico.seleccion() } }

    suspend fun subir(u: Uri?, que: String): String? {
        u ?: return null
        subiendo = if (que == "portada") "foto de portada" else que
        try {
            val jpg = Fotos.achicar(ctx, u, 1600, 85) ?: return null
            return Web.subirFoto(jpg, estado.api.token)
        } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
            estado.avisar("No pudimos subir la foto", ok = false); return null
        } finally { subiendo = null }
    }

    suspend fun cargar() {
        val v = estado.api.opcional<ValorJSON>("/v1/panel/mi-tienda")
        if (v == null) { estado.avisar("No pudimos traer tu tienda", ok = false); cargado = true; return }
        val plantilla = v["plantillaId"]?.texto ?: ""
        diseno = DisenoTienda.desde(v["diseno"], plantilla).let { if (it.plantilla.isEmpty()) it.copy(plantilla = plantilla) else it }
        historia.atras.clear(); historia.adelante.clear()
        direccion = v["direccion"]?.texto ?: ""
        sinPublicar = v["sinPublicar"]?.bool ?: false
        categorias = v["vista"]?.get("categorias")?.lista?.mapNotNull { it.texto } ?: emptyList()
        hayResenas = v["vista"]?.get("resena")?.objeto != null
        slug = estado.tienda?.slug ?: ""
        cargado = true
        productos = estado.api.opcional<PaginaProductos>("/v1/panel/listas/productos", listOf("cuantos" to "100"))?.filas ?: emptyList()
    }

    suspend fun guardar(publicar: Boolean, avisar: Boolean = true): Boolean {
        val cuerpo = diseno.json
        if (cuerpo.toString().toByteArray().size > cat.limites.bytesDiseno) {
            estado.avisar("Tu tienda tiene demasiados bloques para guardarla: sacá algunos y probá de nuevo", ok = false); return false
        }
        guardando = true
        try {
            estado.api.hacer("PUT", "/v1/panel/mi-tienda", cuerpo)
            // El logo nuevo se ve enseguida en Inicio y en Más (en vez de las iniciales).
            if (diseno.logo != (estado.tienda?.logo ?: "")) estado.tarea { estado.refrescarAcceso() }
            if (publicar) {
                estado.api.hacer("POST", "/v1/panel/mi-tienda/publicar")
                sinPublicar = false
                Haptico.exito()
                estado.avisar("¡Publicamos los cambios! Ahora compartí el link")
            } else {
                sinPublicar = true
                if (avisar) estado.avisar("Guardamos el borrador")
            }
            return true
        } catch (e: ErrorApi) { estado.avisar(e.detalle, ok = false) } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
            estado.avisar("No pudimos guardar", ok = false)
        } finally { guardando = false }
        return false
    }

    /** Guarda el borrador y abre la tienda de verdad con él. */
    suspend fun mirar() { if (guardar(publicar = false, avisar = false)) { enLaVista = diseno; vuelta += 1; previa = true } }

    @Serializable data class C2(val slug: String)
    suspend fun cambiarDireccion() {
        val s = slug.lowercase().trim()
        try {
            estado.api.hacer("PUT", "/v1/panel/tienda/direccion", C2(s))
            estado.avisar("Listo: tu tienda tiene dirección nueva")
            estado.api.opcional<ValorJSON>("/v1/panel/mi-tienda")?.let { direccion = it["direccion"]?.texto ?: direccion }
        } catch (e: ErrorApi) { estado.avisar(e.detalle, ok = false) } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
            estado.avisar("No pudimos cambiarla", ok = false)
        }
    }

    // Los bloques de la página elegida.
    fun bloquesDe(p: String) = diseno.bloques(p)
    fun ponerBloques(p: String, l: List<Bloque>, clave: String? = null) = cambiar(diseno.conBloques(p, l), clave)
    fun usados() = diseno.todos.map { it.id }.toSet()
    fun reemplazar(b: Bloque, clave: String?) {
        val (p, i) = diseno.ubicar(b.id) ?: return
        ponerBloques(p, bloquesDe(p).toMutableList().also { it[i] = b }, clave)
    }
    fun agregar(t: TipoBloque, en: Int) {
        val l = bloquesDe(pagina)
        val b = Bloque(Bloque.nuevoId(usados()), t.tipo, t.variantes.firstOrNull() ?: "")
        ponerBloques(pagina, l.take(en) + b + l.drop(en))
        Haptico.exito()
        editando = b.id
    }
    fun duplicar(b: Bloque) {
        val (p, i) = diseno.ubicar(b.id) ?: return
        val l = bloquesDe(p)
        if (l.size >= cat.limites.bloquesPorPagina) { estado.avisar("Cada página va hasta ${cat.limites.bloquesPorPagina} bloques.", ok = false); return }
        val t = cat.tipo(b.tipo)
        if (t?.max != null && l.count { it.tipo == b.tipo } >= t.max) { estado.avisar("Ya hay un bloque de ${t.nombre.lowercase()}.", ok = false); return }
        ponerBloques(p, l.take(i + 1) + b.copy(id = Bloque.nuevoId(usados())) + l.drop(i + 1))
        estado.avisar("Duplicamos «${nombreDe(b)}»")
    }
    fun alternar(b: Bloque) = reemplazar(b.copy(oculto = if (b.visible) true else null), null)
    fun borrar(b: Bloque) {
        val (p, _) = diseno.ubicar(b.id) ?: return
        val antes = diseno
        if (editando == b.id) editando = null
        ponerBloques(p, bloquesDe(p).filter { it.id != b.id })
        estado.avisar("Borramos «${nombreDe(b)}»", deshacer = { cambiar(antes) })
    }

    val elegirLogo = rememberLauncherForActivityResult(ActivityResultContracts.PickVisualMedia()) { u -> alcance.launch { subir(u, "logo")?.let { cambiar(diseno.copy(logo = it)) } } }
    val elegirLogoOscuro = rememberLauncherForActivityResult(ActivityResultContracts.PickVisualMedia()) { u ->
        alcance.launch { subir(u, "logo para fondos oscuros")?.let { cambiar(diseno.copy(estilo = diseno.estilo.copy(logoOscuro = it))) } }
    }
    val elegirIcono = rememberLauncherForActivityResult(ActivityResultContracts.PickVisualMedia()) { u ->
        alcance.launch { subir(u, "ícono de la pestaña")?.let { cambiar(diseno.copy(estilo = diseno.estilo.copy(icono = it))) } }
    }
    val soloImagenes = PickVisualMediaRequest(ActivityResultContracts.PickVisualMedia.ImageOnly)

    @Composable
    fun SelectorFoto(titulo: String, url: String, modifier: Modifier, oscuro: Boolean = false, elegir: () -> Unit) {
        Column(modifier.clip(RoundedCornerShape(14.dp)).background(C.crema2).clickable(onClick = elegir).padding(8.dp), horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.spacedBy(6.dp)) {
            when {
                subiendo == titulo.lowercase() -> Box(Modifier.height(70.dp), contentAlignment = Alignment.Center) { Rueda(C.acento) }
                url.isNotEmpty() -> AsyncImage(url, null, Modifier.fillMaxWidth().height(70.dp).clip(RoundedCornerShape(10.dp)).background(if (oscuro) hex(0x1C1A17) else Color.Transparent),
                    contentScale = ContentScale.Crop)
                else -> Box(Modifier.height(70.dp), contentAlignment = Alignment.Center) { Icono("photo.badge.plus", 22f, C.primarioTexto) }
            }
            Texto(titulo, Estilo.footnoteFuerte, C.primarioTinta, alinear = TextAlign.Center)
        }
    }

    LaunchedEffect(Unit) { cargar() }

    // La vista previa abierta sigue al editor: sin recargar si la página sabe (tiendaVista.mostrar); si no, guarda el
    // borrador y recarga, como antes.
    LaunchedEffect(diseno, pagina, previa, puente.lista) {
        if (!previa) return@LaunchedEffect
        if (puente.lista) { puente.mostrar(diseno.json, pagina); enLaVista = diseno }
        else if (enLaVista != null && enLaVista != diseno) {
            delay(1500)
            if (guardar(publicar = false, avisar = false)) { enLaVista = diseno; vuelta += 1 }
        }
    }
    LaunchedEffect(editando, previa) { editando?.let { if (previa) puente.seleccionar(it) } }

    val contexto = ContextoCampos(estado.tienda?.nombre ?: "", diseno.lema, categorias, productos, hayResenas) { u -> subir(u, "foto") }

    CompositionLocalProvider(LocalContextoCampos provides contexto) {
        ConBarraAccion({
            if (cargado) {
                BotonPrincipal(if (guardando) "Publicando…" else "Publicar cambios", { alcance.launch { guardar(publicar = true) } }, habilitado = !guardando && subiendo == null)
                BotonTexto("Guardar borrador", { alcance.launch { guardar(publicar = false) } }, habilitado = !guardando && subiendo == null)
            }
        }, Modifier.background(C.fondo).imePadding()) { abajo ->
            Column(Modifier.fillMaxSize().verticalScroll(rememberScrollState()).margenPantalla().padding(bottom = 32.dp + abajo), verticalArrangement = Arrangement.spacedBy(14.dp)) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Texto("Mi tienda", Estilo.tituloGrande, C.texto, Modifier.weight(1f))
                    if (cargado) {
                        Box(Modifier.size(40.dp).alpha(if (historia.atras.isEmpty()) 0.35f else 1f).clickable(enabled = historia.atras.isNotEmpty()) { deshacer() }
                            .semantics { contentDescription = "Deshacer" }, contentAlignment = Alignment.Center) { Icono("arrow.uturn.backward", 18f, C.primarioTexto) }
                        Box(Modifier.size(40.dp).alpha(if (historia.adelante.isEmpty()) 0.35f else 1f).clickable(enabled = historia.adelante.isNotEmpty()) { rehacer() }
                            .semantics { contentDescription = "Rehacer" }, contentAlignment = Alignment.Center) { Icono("arrow.uturn.forward", 18f, C.primarioTexto) }
                    }
                }
                if (!cargado) EsqueletoFilas(5, 0.dp)
                else {
                    // La tarjeta de arriba: cómo arranca la tienda, si hay cambios sin publicar, la vista previa y el link.
                    val p = plantillas.firstOrNull { it.id == diseno.plantilla }
                    val foto = diseno.inicio.firstOrNull { it.tipo == "portada" }?.get("fotos")?.lista?.firstOrNull()?.texto ?: diseno.portada.ifEmpty { p?.img ?: "" }
                    Column(Modifier.fillMaxWidth().tarjeta(radio = 20.dp)) {
                        Box(Modifier.fillMaxWidth().height(150.dp).background(hex(p?.colorFondo ?: 0xF4EFEA))) {
                            AsyncImage(foto, null, Modifier.fillMaxSize(), contentScale = ContentScale.Crop)
                            Box(Modifier.fillMaxSize().background(Brush.verticalGradient(0.5f to Color.Transparent, 1f to Color.Black.copy(alpha = 0.55f))))
                            Row(Modifier.align(Alignment.BottomStart).padding(14.dp), horizontalArrangement = Arrangement.spacedBy(10.dp), verticalAlignment = Alignment.Bottom) {
                                if (diseno.logo.isNotEmpty()) AsyncImage(diseno.logo, null, Modifier.size(44.dp).clip(RoundedCornerShape(10.dp)).background(Color.White), contentScale = ContentScale.Fit)
                                Column(verticalArrangement = Arrangement.spacedBy(2.dp)) {
                                    Texto(diseno.titulo.ifEmpty { estado.tienda?.nombre ?: "" }, Estilo(20f, Peso.bold), Color.White)
                                    if (diseno.lema.isNotEmpty()) Texto(diseno.lema, Estilo.footnote, Color.White.copy(alpha = 0.9f))
                                }
                            }
                        }
                        Row(Modifier.fillMaxWidth().padding(12.dp), horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
                            if (sinPublicar) ChipSimple("Hay cambios sin publicar", C.avisoFondo, C.avisoTinta) else ChipSimple("Publicada", C.okFondo, C.okTinta)
                            Spacer(Modifier.weight(1f))
                            Texto("Vista previa", Estilo.subheadFuerte, C.primarioTexto, Modifier.alpha(if (guardando || subiendo != null) 0.4f else 1f)
                                .clickable(enabled = !guardando && subiendo == null) { alcance.launch { mirar() } })
                            Texto("Ver tienda", Estilo.subheadFuerte, C.primarioTexto, Modifier.clickable { urlTienda?.let { abrirEnNavegador(ctx, it) } })
                        }
                        if (urlTienda != null && !sinPublicar) {
                            Row(Modifier.padding(horizontal = 12.dp).padding(bottom = 12.dp), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                                BotonTonal("Copiar link", { Compartir.copiar(ctx, urlTienda, estado) }, Modifier.weight(1f), alto = 38.dp, lleno = true, simbolo = "doc.on.doc")
                                estado.tienda?.let { t ->
                                    BotonTonal("Compartir", { compartirTexto(ctx, Compartir.mensajeTienda(t)) }, Modifier.weight(1f), alto = 38.dp, lleno = true, simbolo = "square.and.arrow.up")
                                }
                            }
                        }
                    }
                    Segmentado(listOf("paginas" to "Páginas", "estilo" to "Estilo", "datos" to "Datos"), pestana, { pestana = it })
                    when (pestana) {
                        "paginas" -> PestanaPaginas(diseno, pagina, { pagina = it }, contexto,
                            alCambiarPaginas = { cambiar(diseno.copy(paginas = it), "paginas") },
                            alMover = { de, a -> ponerBloques(pagina, bloquesDe(pagina).toMutableList().also { l -> l.add(a, l.removeAt(de)) }, "mover.$pagina") },
                            alEditar = { editando = it.id }, alAgregar = { agregarEn = it }, alDuplicar = { duplicar(it) }, alOcultar = { alternar(it) }, alBorrar = { borrar(it) })
                        "estilo" -> {
                            TarjetaSeccion("Punto de partida", "rectangle.3.group") {
                                Texto("Cada plantilla trae su dirección, sus colores y sus bloques. Al cambiarla conservamos lo que cargaste.", Estilo.footnote, C.texto2)
                                Row(Modifier.fillMaxWidth().horizontalScroll(rememberScrollState()), horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                                    plantillas.forEach { pl ->
                                        Column(Modifier.width(120.dp).clickable { if (pl.id != diseno.plantilla) presetPedido = pl.id }, verticalArrangement = Arrangement.spacedBy(6.dp)) {
                                            val forma = RoundedCornerShape(14.dp)
                                            Box(Modifier.size(120.dp, 150.dp).clip(forma).background(hex(pl.colorFondo)).border(3.dp, if (diseno.plantilla == pl.id) C.acento else Color.Transparent, forma)) {
                                                AsyncImage(pl.img, null, Modifier.fillMaxSize().clip(forma), contentScale = ContentScale.Crop)
                                                Texto(pl.lema, Estilo(11f, Peso.semibold), Color.White, Modifier.align(Alignment.BottomStart).padding(6.dp))
                                            }
                                            Texto(pl.nombre, Estilo.fila, C.texto)
                                            Texto(pl.rubro, Estilo.footnote, C.texto2)
                                        }
                                    }
                                }
                            }
                            EditorEstilo(diseno.estilo, { cambiar(diseno.copy(estilo = it), "estilo") }, diseno.paleta, { cambiar(diseno.copy(paleta = it), "paleta") }, diseno.plantilla, categorias)
                        }
                        else -> {
                            TarjetaSeccion("Tu marca", "seal") {
                                CampoTexto("Nombre en la tienda (vacío: el de tu negocio)", diseno.titulo, { cambiar(diseno.copy(titulo = it.take(80)), "titulo") })
                                CampoTexto("Lema", diseno.lema, { cambiar(diseno.copy(lema = it.take(160)), "lema") })
                                Texto("La portada usa estos dos si no le escribís otros.", Estilo.footnote, C.texto2)
                                Row(horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                                    SelectorFoto("Logo", diseno.logo, Modifier.weight(1f)) { elegirLogo.launch(soloImagenes); estado.abrioOtraApp = true }
                                    SelectorFoto("Logo para fondos oscuros", diseno.estilo.logoOscuro ?: "", Modifier.weight(1f), oscuro = true) { elegirLogoOscuro.launch(soloImagenes); estado.abrioOtraApp = true }
                                    SelectorFoto("Ícono de la pestaña", diseno.estilo.icono ?: "", Modifier.weight(1f)) { elegirIcono.launch(soloImagenes); estado.abrioOtraApp = true }
                                }
                                if (diseno.logo.isNotEmpty()) {
                                    Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
                                        Texto("Tamaño del logo · ${diseno.estilo.logoAlto?.let { "$it px" } ?: "el de la plantilla"}", Estilo.footnoteFuerte, C.texto2)
                                        Deslizador((diseno.estilo.logoAlto ?: 36).toFloat(), { cambiar(diseno.copy(estilo = diseno.estilo.copy(logoAlto = it.roundToInt())), "logoAlto") }, 24f..72f, 2f)
                                    }
                                }
                                if (diseno.logo.isNotEmpty() || !diseno.estilo.logoOscuro.isNullOrEmpty() || !diseno.estilo.icono.isNullOrEmpty()) {
                                    Row(Modifier.fillMaxWidth()) {
                                        if (diseno.logo.isNotEmpty()) BotonTexto("Quitar logo", { cambiar(diseno.copy(logo = "")) }, Modifier.weight(1f))
                                        if (!diseno.estilo.logoOscuro.isNullOrEmpty()) BotonTexto("Quitar oscuro", { cambiar(diseno.copy(estilo = diseno.estilo.copy(logoOscuro = null))) }, Modifier.weight(1f))
                                        if (!diseno.estilo.icono.isNullOrEmpty()) BotonTexto("Quitar ícono", { cambiar(diseno.copy(estilo = diseno.estilo.copy(icono = null))) }, Modifier.weight(1f))
                                    }
                                }
                            }
                            TarjetaSeccion("Al compartir el link", "square.and.arrow.up") {
                                CampoTexto("Título (vacío: el sugerido)", diseno.compartirTitulo, { cambiar(diseno.copy(compartirTitulo = it), "compartirTitulo") })
                                CampoTexto("Descripción", diseno.compartirDescripcion, { cambiar(diseno.copy(compartirDescripcion = it), "compartirDescripcion") })
                                CampoFoto("Foto (vacío: la de la portada)", diseno.compartirFoto) { cambiar(diseno.copy(compartirFoto = it ?: "")) }
                            }
                            TarjetaSeccion("Dirección de tu tienda", "link") {
                                SelectionContainer { Texto(direccion, Estilo.texto14, C.primarioTexto) }
                                CampoTexto("Nombre en la dirección (letras, números y guiones)", slug, { slug = it })
                                BotonSecundario("Cambiar dirección", { alcance.launch { cambiarDireccion() } }, habilitado = slug.isNotEmpty() && !direccion.contains(slug))
                            }
                        }
                    }
                }
            }
        }

        // La biblioteca para sumar un bloque.
        Hoja(agregarEn != null, { agregarEn = null }) {
            BibliotecaBloques(pagina, bloquesDe(pagina)) { t -> val en = agregarEn ?: bloquesDe(pagina).size; agregarEn = null; agregar(t, en) }
        }
        // El panel del bloque (desde la lista o tocándolo en la vista previa).
        Hoja(editando != null, { editando = null }) {
            val b = editando?.let { id -> diseno.ubicar(id)?.let { (p, i) -> diseno.bloques(p)[i] } }
            if (b != null) PanelBloque(b, { nb, clave -> reemplazar(nb, clave) }, { duplicar(b) }, { borrar(b) }, { editando = null })
        }
        // Cambiar el punto de partida.
        Confirmacion(presetPedido != null, "¿Cambiar a ${plantillas.firstOrNull { it.id == presetPedido }?.nombre ?: "esta plantilla"}?",
            "Toma su dirección, sus colores y el orden de sus bloques. Lo que cargaste en cada bloque se queda y lo podés deshacer.",
            listOf(OpcionConfirmacion("Cambiar") {
                presetPedido?.let { id ->
                    cambiar(diseno.conPreset(id))
                    estado.avisar("Listo: tu tienda arranca de ${plantillas.firstOrNull { it.id == id }?.nombre ?: id}", deshacer = { deshacer() })
                }
            }), { presetPedido = null })
        // La vista previa real, a pantalla completa con «Listo» (la hoja con NavigationStack de iOS).
        Hoja(previa, { previa = false; enLaVista = null }) {
            Column(Modifier.fillMaxSize()) {
                Box(Modifier.fillMaxWidth().height(44.dp).padding(horizontal = 16.dp)) {
                    Texto("Así se ve", Estilo.headline, C.texto, Modifier.align(Alignment.Center))
                    Texto("Listo", Estilo(17f, Peso.semibold), C.primarioTexto, Modifier.align(Alignment.CenterEnd).clickable { previa = false; enLaVista = null })
                }
                Row(Modifier.fillMaxWidth().horizontalScroll(rememberScrollState()).padding(horizontal = 16.dp, vertical = 6.dp), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    CatalogoBloques.paginasEditor.forEach { (id, nombre) -> ChipFiltro(nombre, activo = pagina == id) { pagina = id } }
                }
                Texto(if (puente.lista) "Tocá un bloque para editarlo." else "Guardamos el borrador para mostrártelo.", Estilo.footnote, C.texto2, Modifier.padding(horizontal = 16.dp))
                VistaPreviaTienda(estado.api.token, vuelta, Modifier.fillMaxSize().padding(top = 6.dp), pagina, puente,
                    alTocarBloque = { id -> if (diseno.ubicar(id) != null) editando = id },
                    alEstarLista = { puente.mostrar(diseno.json, pagina); enLaVista = diseno })
            }
        }
    }
}

/** Páginas › los bloques de la página elegida (Inicio, Ficha y Nosotros) o sus ajustes (Catálogo, Ficha y Carrito). */
@Composable
private fun PestanaPaginas(
    diseno: DisenoTienda,
    pagina: String,
    alElegirPagina: (String) -> Unit,
    contexto: ContextoCampos,
    alCambiarPaginas: (PaginasTienda) -> Unit,
    alMover: (Int, Int) -> Unit,
    alEditar: (Bloque) -> Unit,
    alAgregar: (Int) -> Unit,
    alDuplicar: (Bloque) -> Unit,
    alOcultar: (Bloque) -> Unit,
    alBorrar: (Bloque) -> Unit,
) {
    val cat = CatalogoBloques.compartido
    val conBloques = cat.paginas[pagina]?.bloques == true
    val bloques = diseno.bloques(pagina)
    Column(verticalArrangement = Arrangement.spacedBy(14.dp)) {
        Row(Modifier.fillMaxWidth().horizontalScroll(rememberScrollState()), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            CatalogoBloques.paginasEditor.forEach { (id, nombre) -> ChipFiltro(nombre, activo = pagina == id) { alElegirPagina(id) } }
        }
        if (pagina != "inicio" && pagina != "nosotros") AjustesPagina(pagina, diseno.paginas, alCambiarPaginas)
        if (conBloques) {
            val titulo = when (pagina) { "producto" -> "Debajo de la ficha"; "nosotros" -> "Tu historia"; else -> "Bloques de la portada" }
            Row(verticalAlignment = Alignment.CenterVertically) {
                Texto(titulo, Estilo.tituloSeccion, C.texto, Modifier.weight(1f))
                Texto("${bloques.size}", Estilo.subheadFuerte, C.texto3, numeros = true)
            }
            val sinVerse = bloques.count { it.visible && faltaEn(it, contexto).isNotEmpty() }
            if (sinVerse > 0) ChipSimple(if (sinVerse == 1) "Un bloque todavía no se ve: tocalo para completarlo" else "$sinVerse bloques todavía no se ven: tocalos para completarlos", C.avisoFondo, C.avisoTinta)
            if (bloques.isEmpty()) {
                Column(Modifier.fillMaxWidth().tarjeta(radio = 20.dp).padding(20.dp), horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(10.dp)) {
                    ImagenTincho(Tincho.pensando, Modifier.size(84.dp))
                    Texto(when (pagina) {
                        "producto" -> "Abajo de cada producto podés sumar preguntas frecuentes, beneficios o reseñas."
                        "nosotros" -> "Contá quién está detrás: tu historia, el local, el equipo y las preguntas que más te hacen."
                        else -> "Tu portada está vacía. Sumá el primer bloque: una portada con foto, tus productos o un banner."
                    }, Estilo.subhead, C.texto2, alinear = TextAlign.Center)
                    BotonPrincipal("Sumar un bloque", { alAgregar(0) }, simbolo = "plus")
                }
            } else {
                Texto("Arrastrá ≡ para ordenar. Tocá un bloque para editarlo.", Estilo.footnote, C.texto2)
                ListaBloques(bloques, alMover, alEditar, alAgregar, alDuplicar, alOcultar, alBorrar)
                BotonSecundario("Sumar un bloque", { alAgregar(bloques.size) }, simbolo = "plus",
                    habilitado = bloques.size < cat.limites.bloquesPorPagina)
            }
        }
    }
}

/** Las plantillas de tienda (assets/plantillas.json, generado desde apps/web/src/lib/plantillas.ts). */
@Serializable
data class PlantillaTienda(val id: String, val nombre: String = "", val rubro: String = "", val img: String = "", val bg: String = "#F4EFEA", val lema: String = "") {
    val colorFondo: Long get() = bg.drop(1).toLongOrNull(16) ?: 0xF4EFEA

    companion object {
        fun todas(): List<PlantillaTienda> = Recursos.leer<List<PlantillaTienda>>("plantillas.json") ?: emptyList()
    }
}
