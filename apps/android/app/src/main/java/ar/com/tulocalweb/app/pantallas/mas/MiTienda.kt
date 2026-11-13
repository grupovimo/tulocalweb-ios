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
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.selection.SelectionContainer
import androidx.compose.foundation.verticalScroll
import androidx.compose.runtime.Composable
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
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.nucleo.ErrorApi
import ar.com.tulocalweb.app.nucleo.Fotos
import ar.com.tulocalweb.app.nucleo.Recursos
import ar.com.tulocalweb.app.nucleo.Valor
import ar.com.tulocalweb.app.nucleo.ValorJSON
import ar.com.tulocalweb.app.nucleo.Web
import ar.com.tulocalweb.app.nucleo.abrirEnNavegador
import ar.com.tulocalweb.app.nucleo.bool
import ar.com.tulocalweb.app.nucleo.compartirTexto
import ar.com.tulocalweb.app.nucleo.entero
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

private data class Seccion(val id: String, val visible: Boolean)
private val nombresSecciones = mapOf("hero" to "Portada", "cats" to "Categorías", "prods" to "Productos", "revs" to "Reseñas", "ig" to "Instagram", "wa" to "Botón de WhatsApp")

/** Editar la tienda online desde el celular, igual que Mi tienda en el panel web: plantilla, colores, tipografías,
 *  logos, portada, secciones, menú, títulos, anuncio, pie, redes, cómo se ve al compartir y la dirección. Se guarda como
 *  borrador, se mira en la vista previa real y se publica. */
@Composable
fun MiTiendaEditor() {
    val estado = LocalEstado.current
    val ctx = LocalContext.current
    val alcance = rememberCoroutineScope()
    var diseno by remember { mutableStateOf<Map<String, ValorJSON>>(emptyMap()) }
    var cargado by remember { mutableStateOf(false) }
    var plantilla by remember { mutableStateOf("") }
    var paleta by remember { mutableIntStateOf(0) }
    var titulo by remember { mutableStateOf("") }
    var lema by remember { mutableStateOf("") }
    var portada by remember { mutableStateOf("") }
    var logo by remember { mutableStateOf("") }
    var compartirTitulo by remember { mutableStateOf("") }
    var compartirDescripcion by remember { mutableStateOf("") }
    val secciones = remember { mutableStateListOf<Seccion>() }
    var direccion by remember { mutableStateOf("") }
    var slug by remember { mutableStateOf("") }
    var sinPublicar by remember { mutableStateOf(false) }
    var guardando by remember { mutableStateOf(false) }
    var subiendo by remember { mutableStateOf<String?>(null) }
    var estilo by remember { mutableStateOf(EstiloTienda()) }
    var categorias by remember { mutableStateOf<List<String>>(emptyList()) }
    var previa by remember { mutableStateOf(false) }
    var vuelta by remember { mutableIntStateOf(0) }
    val plantillas = remember { PlantillaTienda.todas() }

    val urlTienda: String? = if (direccion.isEmpty()) estado.tienda?.url else (if (direccion.startsWith("http")) direccion else "https://$direccion")

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
        v["diseno"]?.objeto?.let { diseno = it }
        plantilla = v["plantillaId"]?.texto ?: ""
        direccion = v["direccion"]?.texto ?: ""
        sinPublicar = v["sinPublicar"]?.bool ?: false
        paleta = diseno["paleta"]?.entero ?: 0
        titulo = diseno["titulo"]?.texto ?: ""
        lema = diseno["lema"]?.texto ?: ""
        portada = diseno["portada"]?.texto ?: ""
        logo = diseno["logo"]?.texto ?: ""
        compartirTitulo = diseno["compartirTitulo"]?.texto ?: ""
        compartirDescripcion = diseno["compartirDescripcion"]?.texto ?: ""
        secciones.clear(); secciones.addAll(diseno["secciones"]?.lista?.mapNotNull { s -> s["id"]?.texto?.let { Seccion(it, s["visible"]?.bool ?: true) } } ?: emptyList())
        estilo = EstiloTienda.desde(diseno["estilo"])
        categorias = v["vista"]?.get("categorias")?.lista?.mapNotNull { it.texto } ?: emptyList()
        slug = estado.tienda?.slug ?: ""
        cargado = true
    }

    suspend fun guardar(publicar: Boolean, avisar: Boolean = true): Boolean {
        guardando = true
        try {
            val d = diseno.toMutableMap()
            d["plantilla"] = Valor.texto(plantilla)
            d["paleta"] = Valor.numero(paleta)
            d["titulo"] = Valor.texto(titulo)
            d["lema"] = Valor.texto(lema)
            d["portada"] = Valor.texto(portada)
            d["logo"] = Valor.texto(logo)
            d["compartirTitulo"] = Valor.texto(compartirTitulo)
            d["compartirDescripcion"] = Valor.texto(compartirDescripcion)
            d["secciones"] = Valor.lista(secciones.map { Valor.objeto(mapOf("id" to Valor.texto(it.id), "visible" to Valor.booleano(it.visible))) })
            d["estilo"] = estilo.json
            estado.api.hacer("PUT", "/v1/panel/mi-tienda", Valor.objeto(d))
            // El logo nuevo se ve enseguida en Inicio y en Más (en vez de las iniciales).
            if (d["logo"]?.texto != estado.tienda?.logo) estado.tarea { estado.refrescarAcceso() }
            diseno = d
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
    suspend fun mirar() { if (guardar(publicar = false, avisar = false)) { vuelta += 1; previa = true } }

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

    val elegirPortada = rememberLauncherForActivityResult(ActivityResultContracts.PickVisualMedia()) { u -> alcance.launch { subir(u, "portada")?.let { portada = it } } }
    val elegirLogo = rememberLauncherForActivityResult(ActivityResultContracts.PickVisualMedia()) { u -> alcance.launch { subir(u, "logo")?.let { logo = it } } }
    val elegirLogoOscuro = rememberLauncherForActivityResult(ActivityResultContracts.PickVisualMedia()) { u -> alcance.launch { subir(u, "logo para fondos oscuros")?.let { estilo = estilo.copy(logoOscuro = it) } } }
    val elegirIcono = rememberLauncherForActivityResult(ActivityResultContracts.PickVisualMedia()) { u -> alcance.launch { subir(u, "ícono de la pestaña")?.let { estilo = estilo.copy(icono = it) } } }
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

    ConBarraAccion({
        if (cargado) {
            BotonPrincipal(if (guardando) "Publicando…" else "Publicar cambios", { alcance.launch { guardar(publicar = true) } }, habilitado = !guardando && subiendo == null)
            BotonTexto("Guardar borrador", { alcance.launch { guardar(publicar = false) } }, habilitado = !guardando && subiendo == null)
        }
    }, Modifier.background(C.fondo).imePadding()) { abajo ->
        Column(Modifier.fillMaxSize().verticalScroll(rememberScrollState()).margenPantalla().padding(bottom = 32.dp + abajo), verticalArrangement = Arrangement.spacedBy(14.dp)) {
            Texto("Mi tienda", Estilo.tituloGrande, C.texto)
            if (!cargado) EsqueletoFilas(5, 0.dp)
            else {
                // La vista previa
                val p = plantillas.firstOrNull { it.id == plantilla }
                Column(Modifier.fillMaxWidth().tarjeta(radio = 20.dp)) {
                    Box(Modifier.fillMaxWidth().height(170.dp).background(hex(p?.colorFondo ?: 0xF4EFEA))) {
                        AsyncImage(portada.ifEmpty { p?.img ?: "" }, null, Modifier.fillMaxSize(), contentScale = ContentScale.Crop)
                        Box(Modifier.fillMaxSize().background(Brush.verticalGradient(0.5f to Color.Transparent, 1f to Color.Black.copy(alpha = 0.55f))))
                        Row(Modifier.align(Alignment.BottomStart).padding(14.dp), horizontalArrangement = Arrangement.spacedBy(10.dp), verticalAlignment = Alignment.Bottom) {
                            if (logo.isNotEmpty()) AsyncImage(logo, null, Modifier.size(44.dp).clip(RoundedCornerShape(10.dp)).background(Color.White), contentScale = ContentScale.Fit)
                            Column(verticalArrangement = Arrangement.spacedBy(2.dp)) {
                                Texto(titulo.ifEmpty { estado.tienda?.nombre ?: "" }, Estilo(20f, Peso.bold), Color.White)
                                if (lema.isNotEmpty()) Texto(lema, Estilo.footnote, Color.White.copy(alpha = 0.9f))
                            }
                        }
                    }
                    Row(Modifier.fillMaxWidth().padding(12.dp), horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
                        if (sinPublicar) ChipSimple("Hay cambios sin publicar", C.avisoFondo, C.avisoTinta) else ChipSimple("Publicada", C.okFondo, C.okTinta)
                        androidx.compose.foundation.layout.Spacer(Modifier.weight(1f))
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
                TarjetaSeccion("Plantilla", "rectangle.3.group") {
                    Row(Modifier.fillMaxWidth().horizontalScroll(rememberScrollState()), horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                        plantillas.forEach { pl ->
                            Column(Modifier.width(120.dp).clickable { plantilla = pl.id }, verticalArrangement = Arrangement.spacedBy(6.dp)) {
                                val forma = RoundedCornerShape(14.dp)
                                Box(Modifier.size(120.dp, 150.dp).clip(forma).background(hex(pl.colorFondo)).border(3.dp, if (plantilla == pl.id) C.acento else Color.Transparent, forma)) {
                                    AsyncImage(pl.img, null, Modifier.fillMaxSize().clip(forma), contentScale = ContentScale.Crop)
                                    Texto(pl.lema, Estilo(11f, Peso.semibold), Color.White, Modifier.align(Alignment.BottomStart).padding(6.dp))
                                }
                                Texto(pl.nombre, Estilo.fila, C.texto)
                                Texto(pl.rubro, Estilo.footnote, C.texto2)
                            }
                        }
                    }
                }
                TarjetaSeccion("Portada", "photo.on.rectangle") {
                    CampoTexto("Título (vacío: el nombre de la tienda)", titulo, { titulo = it })
                    CampoTexto("Lema", lema, { lema = it })
                    Row(horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                        SelectorFoto("Foto de portada", portada, Modifier.weight(1f)) { elegirPortada.launch(soloImagenes); estado.abrioOtraApp = true }
                        SelectorFoto("Logo", logo, Modifier.weight(1f)) { elegirLogo.launch(soloImagenes); estado.abrioOtraApp = true }
                    }
                }
                TarjetaSeccion("Logos", "seal") {
                    Row(horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                        SelectorFoto("Logo para fondos oscuros", estilo.logoOscuro ?: "", Modifier.weight(1f), oscuro = true) { elegirLogoOscuro.launch(soloImagenes); estado.abrioOtraApp = true }
                        SelectorFoto("Ícono de la pestaña", estilo.icono ?: "", Modifier.weight(1f)) { elegirIcono.launch(soloImagenes); estado.abrioOtraApp = true }
                    }
                    if (logo.isNotEmpty()) {
                        Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
                            Texto("Tamaño del logo · ${estilo.logoAlto?.let { "$it px" } ?: "el de la plantilla"}", Estilo.footnoteFuerte, C.texto2)
                            Deslizador((estilo.logoAlto ?: 36).toFloat(), { estilo = estilo.copy(logoAlto = it.roundToInt()) }, 24f..72f, 2f)
                        }
                    }
                    if (!estilo.logoOscuro.isNullOrEmpty() || !estilo.icono.isNullOrEmpty()) {
                        Row(Modifier.fillMaxWidth()) {
                            if (!estilo.logoOscuro.isNullOrEmpty()) BotonTexto("Quitar logo oscuro", { estilo = estilo.copy(logoOscuro = null) }, Modifier.weight(1f))
                            if (!estilo.icono.isNullOrEmpty()) BotonTexto("Quitar ícono", { estilo = estilo.copy(icono = null) }, Modifier.weight(1f))
                        }
                    }
                }
                TarjetaSeccion("Secciones", "list.bullet.rectangle") {
                    Column {
                        secciones.toList().forEachIndexed { i, s ->
                            if (i > 0) Separador()
                            Row(Modifier.padding(vertical = 8.dp), horizontalArrangement = Arrangement.spacedBy(10.dp), verticalAlignment = Alignment.CenterVertically) {
                                Column(verticalArrangement = Arrangement.spacedBy(2.dp)) {
                                    Box(Modifier.alpha(if (i == 0) 0.3f else 1f).clickable(enabled = i > 0) { val x = secciones[i]; secciones[i] = secciones[i - 1]; secciones[i - 1] = x }) {
                                        Icono("chevron.up", 12f, C.primarioTexto)
                                    }
                                    Box(Modifier.alpha(if (i == secciones.size - 1) 0.3f else 1f).clickable(enabled = i < secciones.size - 1) {
                                        val x = secciones[i]; secciones[i] = secciones[i + 1]; secciones[i + 1] = x
                                    }) { Icono("chevron.down", 12f, C.primarioTexto) }
                                }
                                Texto(nombresSecciones[s.id] ?: s.id, Estilo.body16, C.texto, Modifier.weight(1f))
                                Llave(s.visible, { v -> secciones[i] = s.copy(visible = v) })
                            }
                        }
                    }
                }
                EditorEstilo(estilo, { estilo = it }, paleta, { paleta = it }, plantilla, categorias)
                TarjetaSeccion("Al compartir el link", "square.and.arrow.up") {
                    CampoTexto("Título (vacío: el sugerido)", compartirTitulo, { compartirTitulo = it })
                    CampoTexto("Descripción", compartirDescripcion, { compartirDescripcion = it })
                }
                TarjetaSeccion("Dirección de tu tienda", "link") {
                    SelectionContainer { Texto(direccion, Estilo.texto14, C.primarioTexto) }
                    CampoTexto("Nombre en la dirección (letras, números y guiones)", slug, { slug = it })
                    BotonSecundario("Cambiar dirección", { alcance.launch { cambiarDireccion() } }, habilitado = slug.isNotEmpty() && !direccion.contains(slug))
                }
            }
        }
    }
    // La vista previa real, a pantalla completa con «Listo» (la hoja con NavigationStack de iOS).
    Hoja(previa, { previa = false }) {
        Column(Modifier.fillMaxSize()) {
            Box(Modifier.fillMaxWidth().height(44.dp).padding(horizontal = 16.dp)) {
                Texto("Así se ve", Estilo.headline, C.texto, Modifier.align(Alignment.Center))
                Texto("Listo", Estilo(17f, Peso.semibold), C.primarioTexto, Modifier.align(Alignment.CenterEnd).clickable { previa = false })
            }
            VistaPreviaTienda(estado.api.token, vuelta, Modifier.fillMaxSize())
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
