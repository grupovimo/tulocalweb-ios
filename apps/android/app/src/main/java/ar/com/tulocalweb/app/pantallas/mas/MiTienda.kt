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
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.nucleo.CatalogoBloques
import ar.com.tulocalweb.app.nucleo.ContextoTienda
import ar.com.tulocalweb.app.nucleo.Diseno
import ar.com.tulocalweb.app.nucleo.ErrorApi
import ar.com.tulocalweb.app.nucleo.errorDe
import ar.com.tulocalweb.app.nucleo.Fotos
import ar.com.tulocalweb.app.nucleo.PresetTienda
import ar.com.tulocalweb.app.nucleo.Recursos
import ar.com.tulocalweb.app.nucleo.Valor
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
import kotlinx.coroutines.launch
import kotlinx.serialization.Serializable
import kotlin.math.roundToInt

// Clon de Pantallas/Mas/MiTienda.swift.

/** Editar la tienda online desde el celular, igual que Mi tienda en el panel web (diseño v2, en bloques): la vista
 *  previa real arriba; Páginas (Inicio, Catálogo, Ficha, Carrito y Nosotros con sus bloques), Estilo (punto de
 *  partida, dirección, colores, tipografías, encabezado, pie, menú, barra y redes), Promociones (ventanas y
 *  herramientas para vender más) y Marca (logos, al compartir y la dirección). Se guarda como borrador y se publica. */
@Composable
fun MiTiendaEditor() {
    val estado = LocalEstado.current
    val ctx = LocalContext.current
    val alcance = rememberCoroutineScope()
    // Lo que edita la persona sobrevive a salir y volver de la pantalla (ver Estado.miTienda).
    val (editor, datos) = remember {
        val clave = estado.tienda?.slug ?: ""
        estado.miTienda?.takeIf { it.first == clave }?.second
            ?: (EditorTienda() to DatosEditor()).also { estado.miTienda = clave to it }
    }
    var cargado by remember { mutableStateOf(false) }
    /** Si no se pudo traer la tienda: sin el diseño real no se guarda (el PUT lo reemplaza entero). */
    var errorCarga by remember { mutableStateOf<ErrorApi?>(null) }
    var pestana by remember { mutableStateOf("paginas") }
    var direccion by remember { mutableStateOf("") }
    var slug by remember { mutableStateOf("") }
    var sinPublicar by remember { mutableStateOf(false) }
    var guardando by remember { mutableStateOf(false) }
    var subiendo by remember { mutableStateOf<String?>(null) }
    var categorias by remember { mutableStateOf<List<String>>(emptyList()) }
    var previaGrande by remember { mutableStateOf(false) }
    var vuelta by remember { mutableIntStateOf(0) }
    /** La vista previa responde al protocolo: los cambios se ven sin guardar. */
    var viva by remember { mutableStateOf(false) }
    var presetPedido by remember { mutableStateOf<PlantillaTienda?>(null) }
    /** La ventana que se pidió ver en la vista previa (la vez cambia en cada pedido). */
    var ventanaVista by remember { mutableStateOf<Pair<String, Int>?>(null) }
    val desplazamiento = rememberScrollState()
    val plantillas = remember { PlantillaTienda.todas() }

    val urlTienda: String? = if (direccion.isEmpty()) estado.tienda?.url else (if (direccion.startsWith("http")) direccion else "https://$direccion")

    suspend fun subir(u: Uri?, que: String): String? {
        u ?: return null
        subiendo = que
        try {
            val jpg = Fotos.achicar(ctx, u, 1600, 85) ?: return null
            return Web.subirFoto(jpg, estado.api.token)
        } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
            estado.avisar("No pudimos subir la foto", ok = false); return null
        } finally { subiendo = null }
    }

    // MARK: Cargar y guardar

    suspend fun cargar() {
        errorCarga = null
        val v = try { estado.api.get<ValorJSON>("/v1/panel/mi-tienda") } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
            // Sin el diseño de verdad no se habilita guardar: publicaría una tienda vacía.
            errorCarga = errorDe(e); return
        }
        val plantilla = v["plantillaId"]?.texto ?: v["plantilla"]?.texto ?: ""
        // Al volver con cambios sin guardar, se quedan los de la persona.
        if (!editor.sinGuardar) editor.empezar(Diseno.desde(v["diseno"]?.objeto ?: emptyMap(), plantilla), plantilla)
        direccion = v["direccion"]?.texto ?: ""
        sinPublicar = v["sinPublicar"]?.bool ?: false
        categorias = v["vista"]?.get("categorias")?.lista?.mapNotNull { it.texto } ?: emptyList()
        val productos = v["vista"]?.get("productos")?.lista ?: emptyList()
        datos.contexto = ContextoTienda(
            nombre = estado.tienda?.nombre ?: "",
            categorias = categorias,
            productos = productos.size,
            resenas = v["vista"]?.get("resena")?.objeto != null,
            fotos = !(v["vista"]?.get("fotos")?.lista ?: emptyList()).isEmpty() || productos.any { !(it["foto"]?.texto ?: "").isEmpty() },
        )
        slug = estado.tienda?.slug ?: ""
        cargado = true
        // Lo que sólo hace falta para los avisos y los selectores: después de mostrar el editor.
        datos.cargarNegocio(estado.api)
        datos.cargarProductos(estado.api)
    }

    suspend fun guardar(publicar: Boolean, avisar: Boolean = true): Boolean {
        if (!cargado) return false // sin la tienda traída, el PUT pisaría el diseño con uno vacío
        val d = editor.diseno
        val cuerpo = Valor.objeto(d.json)
        if (cuerpo.toString().toByteArray().size > CatalogoBloques.shared.limites.bytesDiseno) {
            estado.avisar("Tu tienda quedó muy grande: sacá algunos bloques o textos largos", ok = false)
            return false
        }
        guardando = true
        try {
            if (editor.sinGuardar || !publicar) {
                estado.api.hacer("PUT", "/v1/panel/mi-tienda", cuerpo)
                editor.guardado = d
                // El logo nuevo se ve enseguida en Inicio y en Más (en vez de las iniciales).
                if (d.logo != (estado.tienda?.logo ?: "")) estado.tarea { estado.refrescarAcceso() }
            }
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

    /** Sin la vista conectada: guarda el borrador y recarga la tienda con él. */
    suspend fun mirar(grande: Boolean) {
        if (!viva || grande) {
            if (editor.sinGuardar && !guardar(publicar = false, avisar = false)) return
            if (!viva) vuelta += 1
        }
        if (grande) previaGrande = true
    }

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

    fun nombreDireccion(plantilla: String): String {
        val d = PresetTienda.todos[plantilla]?.direccion ?: ""
        return CatalogoBloques.shared.estilo["direccion"]?.let { CatalogoBloques.nombre(it, d) } ?: d
    }

    /** «Ver en la vista previa» de una ventana: sube a la vista y la abre con el borrador. Sin la vista conectada, primero
     *  guarda el borrador y la recarga (la ventana se abre cuando la página avisa que está lista). */
    fun verVentana(id: String) {
        editor.ventanaAbierta = null
        ventanaVista = id to ((ventanaVista?.second ?: 0) + 1)
        alcance.launch { desplazamiento.animateScrollTo(0) }
        if (!viva) alcance.launch { mirar(grande = false) }
    }

    // Cupones, envío gratis y métricas sólo hacen falta en Promociones: se traen la primera vez que se abre.
    LaunchedEffect(pestana) {
        if (pestana == "promos" && !datos.metricasCargadas) { datos.cargarPromos(estado.api); datos.cargarMetricas(estado.api) }
    }

    fun cambiarPreset(p: PlantillaTienda) {
        editor.aplicarPreset(p.id)
        Haptico.exito()
        estado.avisar("Pasamos a «${p.nombre}»", deshacer = { editor.deshacer() })
    }

    val elegirLogo = rememberLauncherForActivityResult(ActivityResultContracts.PickVisualMedia()) { u ->
        alcance.launch { subir(u, "logo")?.let { url -> editor.cambiar { it.copy(logo = url) } } }
    }
    val elegirLogoOscuro = rememberLauncherForActivityResult(ActivityResultContracts.PickVisualMedia()) { u ->
        alcance.launch { subir(u, "logo para fondos oscuros")?.let { url -> editor.cambiar { it.copy(estilo = it.estilo.copy(logoOscuro = url)) } } }
    }
    val elegirIcono = rememberLauncherForActivityResult(ActivityResultContracts.PickVisualMedia()) { u ->
        alcance.launch { subir(u, "ícono de la pestaña")?.let { url -> editor.cambiar { it.copy(estilo = it.estilo.copy(icono = url)) } } }
    }
    val soloImagenes = PickVisualMediaRequest(ActivityResultContracts.PickVisualMedia.ImageOnly)

    @Composable
    fun SelectorFoto(titulo: String, url: String, modifier: Modifier, oscuro: Boolean = false, elegir: () -> Unit) {
        Column(modifier.clip(RoundedCornerShape(14.dp)).background(C.crema2).clickable(onClick = elegir).padding(8.dp), horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.spacedBy(6.dp)) {
            when {
                subiendo == titulo.lowercase() -> Box(Modifier.height(60.dp), contentAlignment = Alignment.Center) { Rueda(C.acento) }
                url.isNotEmpty() -> AsyncImage(url, null, Modifier.fillMaxWidth().height(60.dp).clip(RoundedCornerShape(10.dp))
                    .background(if (oscuro) hex(0x1C1A17) else Color.Transparent).padding(4.dp), contentScale = ContentScale.Fit)
                else -> Box(Modifier.height(60.dp), contentAlignment = Alignment.Center) { Icono("photo.badge.plus", 22f, C.primarioTexto) }
            }
            Texto(titulo, Estilo.footnoteFuerte, C.primarioTinta, alinear = TextAlign.Center, lineas = 2, escalaMinima = 0.85f)
        }
    }

    LaunchedEffect(Unit) { cargar() }

    CompositionLocalProvider(LocalDatosEditor provides datos) {
        ConBarraAccion({
            if (cargado) {
                BotonPrincipal(if (guardando) "Publicando…" else "Publicar cambios", { alcance.launch { guardar(publicar = true) } }, habilitado = !guardando && subiendo == null)
                BotonTexto(if (editor.sinGuardar) "Guardar borrador" else "Borrador guardado", { alcance.launch { guardar(publicar = false) } },
                    habilitado = !guardando && subiendo == null && editor.sinGuardar)
            }
        }, Modifier.background(C.fondo).imePadding()) { abajo ->
            Column(Modifier.fillMaxSize().verticalScroll(desplazamiento).margenPantalla().padding(bottom = 32.dp + abajo), verticalArrangement = Arrangement.spacedBy(14.dp)) {
                Texto("Mi tienda", Estilo.tituloGrande, C.texto)
                val error = errorCarga
                if (error != null) Box(Modifier.padding(top = 60.dp)) { EstadoError("tu tienda", error) { alcance.launch { cargar() } } }
                else if (!cargado) EsqueletoFilas(5, 0.dp)
                else {
                    // MARK: Vista previa
                    Column(Modifier.fillMaxWidth().tarjeta(radio = 20.dp)) {
                        VistaPreviaTienda(estado.api.token, vuelta, Modifier.fillMaxWidth().height(380.dp).clip(RoundedCornerShape(topStart = 20.dp, topEnd = 20.dp)).background(C.pista),
                            pagina = editor.pagina, diseno = editor.diseno.textoJSON, seleccion = editor.seleccion, ventana = ventanaVista, marcos = true,
                            alTocarBloque = { id -> editor.abrir(id); pestana = "paginas" }, alCambiarViva = { viva = it })
                        Row(Modifier.fillMaxWidth().padding(12.dp), horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
                            when {
                                editor.sinGuardar && !viva -> ChipSimple("Guardá para verlo", C.avisoFondo, C.avisoTinta)
                                sinPublicar || editor.sinGuardar -> ChipSimple("Hay cambios sin publicar", C.avisoFondo, C.avisoTinta)
                                else -> ChipSimple("Publicada", C.okFondo, C.okTinta)
                            }
                            Spacer(Modifier.weight(1f))
                            if (!viva) {
                                Texto("Actualizar", Estilo.subheadFuerte, C.primarioTexto, Modifier.alpha(if (guardando || subiendo != null) 0.4f else 1f)
                                    .clickable(enabled = !guardando && subiendo == null) { alcance.launch { mirar(grande = false) } })
                            }
                            Box(Modifier.clickable { alcance.launch { mirar(grande = true) } }.semantics { contentDescription = "Ver en grande" }) {
                                Icono("arrow.up.left.and.arrow.down.right", 15f, C.primarioTexto)
                            }
                            Texto("Ver tienda", Estilo.subheadFuerte, C.primarioTexto, Modifier.clickable { urlTienda?.let { abrirEnNavegador(ctx, it) } })
                        }
                        if (urlTienda != null && !sinPublicar && !editor.sinGuardar) {
                            Row(Modifier.padding(horizontal = 12.dp).padding(bottom = 12.dp), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                                BotonTonal("Copiar link", { Compartir.copiar(ctx, urlTienda, estado) }, Modifier.weight(1f), alto = 38.dp, lleno = true, simbolo = "doc.on.doc")
                                estado.tienda?.let { t ->
                                    BotonTonal("Compartir", { compartirTexto(ctx, Compartir.mensajeTienda(t)) }, Modifier.weight(1f), alto = 38.dp, lleno = true, simbolo = "square.and.arrow.up")
                                }
                            }
                        }
                    }
                    Segmentado(listOf("paginas" to "Páginas", "estilo" to "Estilo", "promos" to "Promociones", "marca" to "Marca"), pestana, { pestana = it })
                    when (pestana) {
                        "estilo" -> {
                            // MARK: Estilo
                            TarjetaSeccion("Punto de partida", "rectangle.3.group") {
                                Texto("Cada uno trae su dirección, sus colores y sus bloques. Cambiarlo conserva tus textos y fotos.", Estilo.footnote, C.texto2)
                                Row(Modifier.fillMaxWidth().horizontalScroll(rememberScrollState()), horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                                    plantillas.forEach { p ->
                                        Column(Modifier.width(120.dp).clickable { if (p.id != editor.plantilla) presetPedido = p }, verticalArrangement = Arrangement.spacedBy(6.dp)) {
                                            val forma = RoundedCornerShape(14.dp)
                                            Box(Modifier.size(120.dp, 150.dp).clip(forma).background(hex(p.colorFondo)).border(3.dp, if (editor.plantilla == p.id) C.acento else Color.Transparent, forma)) {
                                                AsyncImage(p.img, null, Modifier.fillMaxSize().clip(forma), contentScale = ContentScale.Crop)
                                                Texto(p.lema, Estilo(11f, Peso.semibold), Color.White, Modifier.align(Alignment.BottomStart).padding(6.dp))
                                            }
                                            Texto(p.nombre, Estilo.fila, C.texto)
                                            Texto(nombreDireccion(p.id), Estilo.footnote, C.texto2)
                                        }
                                    }
                                }
                            }
                            EditorEstilo(editor.diseno.estilo, { e -> editor.cambiar("estilo") { it.copy(estilo = e) } }, editor.diseno.paleta,
                                { p -> editor.cambiar("paleta") { it.copy(paleta = p) } }, editor.plantilla, categorias)
                        }
                        "marca" -> {
                            // MARK: Marca
                            val d = editor.diseno
                            TarjetaSeccion("Nombre y logo", "seal") {
                                CampoTexto("Nombre que se ve (vacío: el de la tienda)", d.titulo, { v -> editor.cambiar("titulo") { it.copy(titulo = v) } })
                                CampoTexto("Frase", d.lema, { v -> editor.cambiar("lema") { it.copy(lema = v) } })
                                Row(horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                                    SelectorFoto("Logo", d.logo, Modifier.weight(1f)) { elegirLogo.launch(soloImagenes); estado.abrioOtraApp = true }
                                    SelectorFoto("Logo para fondos oscuros", d.estilo.logoOscuro ?: "", Modifier.weight(1f), oscuro = true) { elegirLogoOscuro.launch(soloImagenes); estado.abrioOtraApp = true }
                                    SelectorFoto("Ícono de la pestaña", d.estilo.icono ?: "", Modifier.weight(1f)) { elegirIcono.launch(soloImagenes); estado.abrioOtraApp = true }
                                }
                                if (d.logo.isNotEmpty()) {
                                    Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
                                        Texto("Tamaño del logo · ${d.estilo.logoAlto?.let { "$it px" } ?: "el de la dirección"}", Estilo.footnoteFuerte, C.texto2)
                                        Deslizador((d.estilo.logoAlto ?: 36).toFloat(), { v -> editor.cambiar("logoAlto") { it.copy(estilo = it.estilo.copy(logoAlto = v.roundToInt())) } }, 24f..72f, 2f)
                                    }
                                }
                                Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
                                    if (d.logo.isNotEmpty()) Texto("Quitar logo", Estilo.footnoteFuerte, C.primarioTexto, Modifier.clickable { editor.cambiar { it.copy(logo = "") } })
                                    Spacer(Modifier.weight(1f))
                                    if (!d.estilo.logoOscuro.isNullOrEmpty()) Texto("Quitar logo oscuro", Estilo.footnoteFuerte, C.primarioTexto,
                                        Modifier.clickable { editor.cambiar { it.copy(estilo = it.estilo.copy(logoOscuro = null)) } })
                                    Spacer(Modifier.weight(1f))
                                    if (!d.estilo.icono.isNullOrEmpty()) Texto("Quitar ícono", Estilo.footnoteFuerte, C.primarioTexto,
                                        Modifier.clickable { editor.cambiar { it.copy(estilo = it.estilo.copy(icono = null)) } })
                                }
                            }
                            TarjetaSeccion("Al compartir el link", "square.and.arrow.up") {
                                CampoTexto("Título (vacío: el sugerido)", d.compartirTitulo, { v -> editor.cambiar("compartirTitulo") { it.copy(compartirTitulo = v) } })
                                CampoTexto("Descripción", d.compartirDescripcion, { v -> editor.cambiar("compartirDescripcion") { it.copy(compartirDescripcion = v) } })
                            }
                            TarjetaSeccion("Dirección de tu tienda", "link") {
                                SelectionContainer { Texto(direccion, Estilo.texto14, C.primarioTexto) }
                                CampoTexto("Nombre en la dirección (letras, números y guiones)", slug, { slug = it })
                                BotonSecundario("Cambiar dirección", { alcance.launch { cambiarDireccion() } }, habilitado = slug.isNotEmpty() && !direccion.contains(slug))
                            }
                        }
                        "promos" -> PestanaPromociones(editor) { verVentana(it) }
                        else -> ListaBloques(editor)
                    }
                }
            }
        }
        Hoja(editor.panel != null, { editor.panel = null }, listOf(Altura.media, Altura.grande)) {
            editor.panel?.let { id -> PanelBloque(editor, id) { editor.panel = null } }
        }
        Hoja(editor.insertarEn != null, { editor.insertarEn = null }) {
            BibliotecaBloques(editor) { editor.insertarEn = null }
        }
        Hoja(editor.ventanaAbierta != null, { editor.ventanaAbierta = null }, listOf(Altura.media, Altura.grande)) {
            editor.ventanaAbierta?.let { id -> PanelVentana(editor, id, { verVentana(it) }) { editor.ventanaAbierta = null } }
        }
        Hoja(editor.creandoVentana, { editor.creandoVentana = false }) {
            PlantillasVentana(editor) { editor.creandoVentana = false }
        }
        // La vista previa en grande (la hoja con NavigationStack de iOS).
        Hoja(previaGrande, { previaGrande = false }) {
            Column(Modifier.fillMaxSize()) {
                Box(Modifier.fillMaxWidth().height(44.dp).padding(horizontal = 16.dp)) {
                    Texto("Así se ve", Estilo.headline, C.texto, Modifier.align(Alignment.Center))
                    Texto("Listo", Estilo(17f, Peso.semibold), C.primarioTexto, Modifier.align(Alignment.CenterEnd).clickable { previaGrande = false })
                }
                VistaPreviaTienda(estado.api.token, vuelta, Modifier.fillMaxSize(), pagina = editor.pagina, diseno = editor.diseno.textoJSON)
            }
        }
        val p = presetPedido
        Confirmacion(p != null, p?.let { "¿Pasar a «${it.nombre}»?" } ?: "",
            "Cambian la dirección, el orden y el diseño de los bloques. Tus textos y fotos quedan, y podés deshacerlo.",
            listOfNotNull(p?.let { OpcionConfirmacion("Pasar a ${it.nombre}") { cambiarPreset(it) } }), { presetPedido = null }, volver = "Cancelar")
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
