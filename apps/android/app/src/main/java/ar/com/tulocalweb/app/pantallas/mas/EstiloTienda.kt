package ar.com.tulocalweb.app.pantallas.mas

import android.annotation.SuppressLint
import android.webkit.WebView
import android.webkit.WebViewClient
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ColumnScope
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import androidx.compose.ui.viewinterop.AndroidView
import ar.com.tulocalweb.app.nucleo.Api
import ar.com.tulocalweb.app.nucleo.Recursos
import ar.com.tulocalweb.app.nucleo.Valor
import ar.com.tulocalweb.app.nucleo.ValorJSON
import ar.com.tulocalweb.app.nucleo.Web
import ar.com.tulocalweb.app.tema.*
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.JsonElement
import kotlin.math.pow
import kotlin.math.roundToInt

// Clon de Pantallas/Mas/EstiloTienda.swift.

/** Mi tienda › Estilo (components/schemas/EstiloTienda), igual que en la web (apps/web/src/lib/tienda/estilo/estilo.ts):
 *  colores por rol y por zona, tipografías, logos, anuncio, títulos de las secciones, redes y el menú. Todo opcional:
 *  vacío es lo de la plantilla. */
@Serializable
data class EstiloTienda(
    val colores: Map<String, String>? = null,
    val fuenteTitulos: String? = null,
    val fuenteTexto: String? = null,
    val logoOscuro: String? = null,
    val icono: String? = null,
    val logoAlto: Int? = null,
    val anuncio: String? = null,
    val anuncioLink: String? = null,
    val leyendaPie: String? = null,
    val textos: Map<String, String>? = null,
    val redes: Map<String, String>? = null,
    val menu: List<ItemMenu>? = null,
) {
    @Serializable data class ItemMenu(val t: String, val a: String)

    /** Para mandarlo dentro del diseño: sin lo vacío (Go lo limpia igual). */
    val json: ValorJSON get() {
        val e = copy(
            colores = colores?.filter { it.value.isNotEmpty() },
            textos = textos?.filter { it.value.trim().isNotEmpty() },
            redes = redes?.filter { it.value.trim().isNotEmpty() },
            logoAlto = if (logoAlto == 0) null else logoAlto,
        )
        return runCatching { Api.json.encodeToJsonElement(serializer(), e) }.getOrDefault(Valor.objeto(emptyMap()))
    }

    companion object {
        fun desde(v: ValorJSON?): EstiloTienda = v?.let { runCatching { Api.json.decodeFromJsonElement(serializer(), it) }.getOrNull() } ?: EstiloTienda()
    }
}

/** Lo que comparte con la web (assets/estilo.json, de scripts/estilo/para-ios.ts). */
@Serializable
data class CatalogoEstilo(
    val fuentes: List<Fuente> = emptyList(),
    val combinaciones: List<Combinacion> = emptyList(),
    val roles: Map<String, Map<String, String>> = emptyMap(),
    val acentos: Map<String, List<String>> = emptyMap(),
    val paginas: List<Pagina> = emptyList(),
    val maxMenu: Int = 8,
) {
    @Serializable data class Fuente(val id: String, val nombre: String, val tipo: String = "")
    @Serializable data class Combinacion(val id: String, val nombre: String, val colores: Map<String, String> = emptyMap())
    @Serializable data class Pagina(val id: String, val nombre: String)

    companion object {
        val compartido: CatalogoEstilo by lazy { Recursos.leer<CatalogoEstilo>("estilo.json") ?: CatalogoEstilo() }

        val rolesColor = listOf(
            "fondo" to "Fondo", "superficie" to "Tarjetas", "texto" to "Texto", "textoSuave" to "Texto suave",
            "borde" to "Bordes", "acento" to "Botones y precios", "sobreAcento" to "Texto de los botones",
            "encabezado" to "Fondo del encabezado", "encabezadoTexto" to "Texto del encabezado",
            "pie" to "Fondo del pie", "pieTexto" to "Texto del pie",
        )
        val titulos = listOf("cats" to "Categorías", "prods" to "Productos destacados", "revs" to "Reseñas", "ig" to "Fotos")
        data class Red(val id: String, val nombre: String, val ejemplo: String, val teclado: KeyboardType)
        val redes = listOf(
            Red("instagram", "Instagram", "@tutienda", KeyboardType.Text), Red("tiktok", "TikTok", "@tutienda", KeyboardType.Text),
            Red("facebook", "Facebook", "facebook.com/tutienda", KeyboardType.Uri), Red("youtube", "YouTube", "@tucanal", KeyboardType.Text),
            Red("email", "Email", "hola@tutienda.com", KeyboardType.Email), Red("telefono", "Teléfono", "11 5555-5555", KeyboardType.Phone),
        )
    }
}

// MARK: - Colores

object Hex {
    fun color(s: String?): Color? {
        if (s == null || s.length != 7 || !s.startsWith("#")) return null
        return s.drop(1).toLongOrNull(16)?.let { hex(it) }
    }

    /** Contraste WCAG 2 entre dos #RRGGBB (1 a 21). */
    fun contraste(a: String, b: String): Double {
        fun lum(h: String): Double {
            val n = h.drop(1).toLongOrNull(16) ?: return 0.0
            val c = listOf((n shr 16) and 0xFF, (n shr 8) and 0xFF, n and 0xFF).map { v ->
                val x = v / 255.0
                if (x <= 0.03928) x / 12.92 else ((x + 0.055) / 1.055).pow(2.4)
            }
            return 0.2126 * c[0] + 0.7152 * c[1] + 0.0722 * c[2]
        }
        val x = lum(a); val y = lum(b)
        return (maxOf(x, y) + 0.05) / (minOf(x, y) + 0.05)
    }
}

/** ColorPicker de iOS en un renglón: el nombre y la muestra; tocándola, una grilla de colores y el código. */
@Composable
fun SelectorDeColor(nombre: String, hexActual: String, alCambiar: (String) -> Unit) {
    var abierto by remember { mutableStateOf(false) }
    Row(Modifier.fillMaxWidth().clickable { abierto = true }, verticalAlignment = Alignment.CenterVertically) {
        Texto(nombre, Estilo.body16, C.texto, Modifier.weight(1f))
        Box(Modifier.size(28.dp).clip(CircleShape).background(Hex.color(hexActual) ?: Color.White).border(1.dp, C.texto3.copy(alpha = 0.6f), CircleShape) /* con anillo marcado: un color oscuro no se pierde en la tarjeta oscura */
            .semantics { contentDescription = "Elegir $nombre" })
    }
    val grilla = listOf(
        "#000000", "#17171A", "#3A2A1C", "#57534E", "#9CA3AF", "#E7E5E0", "#FAFAF7", "#FFFFFF",
        "#B0441D", "#D9652F", "#E3A93B", "#C9821A", "#8B5E34", "#F6EEDF", "#FBF3EA", "#F4EFEA",
        "#1F6B48", "#2F8F63", "#84CC16", "#14B8A6", "#0EA5E9", "#2F80ED", "#1D4ED8", "#2C3357",
        "#7A5CFA", "#6D28D9", "#EC4899", "#BE185D", "#D9534F", "#B91C1C", "#F97316", "#EEF2EA",
    )
    Hoja(abierto, { abierto = false }, listOf(Altura.media)) {
        var texto by remember(hexActual) { mutableStateOf(hexActual.uppercase()) }
        Column(Modifier.fillMaxWidth().padding(horizontal = 20.dp), verticalArrangement = Arrangement.spacedBy(14.dp)) {
            Grabber()
            Row(verticalAlignment = Alignment.CenterVertically) {
                Texto(nombre, Estilo.tituloHojaChico, C.texto, Modifier.weight(1f))
                Texto("Listo", Estilo(17f, Peso.bold), C.primarioTexto, Modifier.alpha(if (Hex.color(texto) != null) 1f else 0.4f)
                    .clickable(enabled = Hex.color(texto) != null) { alCambiar(texto.uppercase()); abierto = false })
            }
            grilla.chunked(8).forEach { fila ->
                Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
                    fila.forEach { c ->
                        val on = texto.equals(c, true)
                        Box(Modifier.size(34.dp).clip(CircleShape).background(Hex.color(c) ?: Color.Gray)
                            .border(if (on) 3.dp else 1.dp, if (on) C.texto else C.texto3.copy(alpha = 0.6f), CircleShape).clickable { texto = c })
                    }
                }
            }
            CampoTexto("Código (#RRGGBB)", texto, { texto = it.uppercase().take(7) })
        }
    }
}

/** Los grupos del estilo, para Mi tienda (los mismos de la web). */
@Composable
fun EditorEstilo(estilo: EstiloTienda, alCambiar: (EstiloTienda) -> Unit, paleta: Int, alCambiarPaleta: (Int) -> Unit, plantilla: String, categorias: List<String>) {
    var abierto by remember { mutableStateOf<String?>("colores") }
    val cat = CatalogoEstilo.compartido
    val original = cat.roles[plantilla] ?: cat.roles["litoral"] ?: emptyMap()

    /** El color que se ve: el elegido, el acento rápido o el de la plantilla (encabezado y pie siguen a su rol). */
    fun visto(rol: String): String {
        estilo.colores?.get(rol)?.let { if (it.length == 7) return it }
        if (rol == "acento") cat.acentos[plantilla]?.getOrNull(paleta)?.let { return it }
        if (rol == "encabezado") return visto("fondo")
        if (rol == "encabezadoTexto") return visto("texto")
        if (rol == "pie" || rol == "pieTexto") original[rol]?.let { o ->
            if (o == original["texto"]) return visto("texto")
            if (o == original["fondo"]) return visto("fondo")
        }
        return original[rol] ?: "#FFFFFF"
    }

    fun poner(c: Map<String, String>) {
        val colores = (estilo.colores ?: emptyMap()).toMutableMap()
        c.forEach { (k, v) -> if (v.isEmpty()) colores.remove(k) else colores[k] = v }
        alCambiar(estilo.copy(colores = colores.ifEmpty { null }))
    }

    val avisos = listOfNotNull(
        if (Hex.contraste(visto("texto"), visto("fondo")) < 4.5) "El texto casi no se lee sobre el fondo." else null,
        if (Hex.contraste(visto("sobreAcento"), visto("acento")) < 4.5) "El texto de los botones casi no se lee." else null,
        if (Hex.contraste(visto("pieTexto"), visto("pie")) < 4.5) "El texto del pie casi no se lee." else null,
    )

    @Composable
    fun Grupo(id: String, titulo: String, simbolo: String, contenido: @Composable ColumnScope.() -> Unit) {
        TarjetaSeccion(titulo, simbolo) {
            Row(Modifier.fillMaxWidth().clickable { abierto = if (abierto == id) null else id }, verticalAlignment = Alignment.CenterVertically) {
                Texto(if (abierto == id) "Ocultar" else "Editar", Estilo.subheadFuerte, C.primarioTexto, Modifier.weight(1f))
                Icono(if (abierto == id) "chevron.up" else "chevron.down", 12f, C.texto2)
            }
            if (abierto == id) Column(verticalArrangement = Arrangement.spacedBy(12.dp), content = contenido)
        }
    }

    val destinos = cat.paginas.map { it.id to it.nombre } + categorias.map { "cat:$it" to "Categoría: $it" } + listOf("link" to "Un link propio…")

    Column(verticalArrangement = Arrangement.spacedBy(14.dp)) {
        Grupo("colores", "Colores", "paintpalette") {
            Texto("Acento rápido", Estilo.footnoteFuerte, C.texto2)
            Row(horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                (cat.acentos[plantilla] ?: emptyList()).forEachIndexed { i, a ->
                    val on = paleta == i && estilo.colores?.get("acento") == null
                    Box(Modifier.size(36.dp).clip(CircleShape).background(Hex.color(a) ?: Color.Gray).border(if (on) 3.dp else 0.dp, if (on) C.texto else Color.Transparent, CircleShape)
                        .clickable { alCambiarPaleta(i); poner(mapOf("acento" to "", "sobreAcento" to "")) }
                        .semantics { contentDescription = if (i == 0) "El de la plantilla" else "Color ${i + 1}" })
                }
            }
            Texto("Combinaciones", Estilo.footnoteFuerte, C.texto2)
            cat.combinaciones.chunked(2).forEach { par ->
                Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    par.forEach { c ->
                        Row(Modifier.weight(1f).clip(RoundedCornerShape(12.dp)).border(1.dp, C.linea2, RoundedCornerShape(12.dp)).clickable {
                            val nuevo = c.colores.toMutableMap()
                            listOf("encabezado", "encabezadoTexto", "pie", "pieTexto").forEach { nuevo[it] = "" }
                            poner(nuevo)
                        }.padding(8.dp), horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.CenterVertically) {
                            Row(Modifier.clip(RoundedCornerShape(7.dp)).background(Hex.color(c.colores["fondo"]) ?: Color.White).padding(4.dp), horizontalArrangement = Arrangement.spacedBy(2.dp)) {
                                listOf(c.colores["superficie"], c.colores["texto"], visto("acento")).forEach { h ->
                                    Box(Modifier.size(7.dp, 14.dp).clip(RoundedCornerShape(2.dp)).background(Hex.color(h) ?: Color.Gray))
                                }
                            }
                            Texto(c.nombre, Estilo.subheadFuerte, C.texto, Modifier.weight(1f), lineas = 1)
                        }
                    }
                    if (par.size == 1) Spacer(Modifier.weight(1f))
                }
            }
            Texto("Todos los colores", Estilo.footnoteFuerte, C.texto2)
            Column {
                CatalogoEstilo.rolesColor.forEach { (rol, nombre) ->
                    Row(Modifier.padding(vertical = 6.dp), horizontalArrangement = Arrangement.spacedBy(10.dp), verticalAlignment = Alignment.CenterVertically) {
                        Box(Modifier.weight(1f)) { SelectorDeColor(nombre, visto(rol)) { poner(mapOf(rol to it)) } }
                        if (estilo.colores?.get(rol) != null) {
                            Box(Modifier.clickable { poner(mapOf(rol to "")) }.semantics { contentDescription = "Volver ${nombre.lowercase()} al de la plantilla" }) {
                                Icono("arrow.uturn.backward", 13f, C.texto2)
                            }
                        }
                    }
                }
            }
            avisos.forEach { ChipSimple(it, C.avisoFondo, C.avisoTinta) }
            if (estilo.colores != null || paleta != 0) BotonTexto("Volver a los colores de la plantilla", { alCambiar(estilo.copy(colores = null)); alCambiarPaleta(0) })
        }
        Grupo("fuentes", "Tipografías", "textformat") {
            val opciones = listOf("" to "La de la plantilla") + cat.fuentes.map { it.id to it.nombre }
            Row(verticalAlignment = Alignment.CenterVertically) {
                Texto("Títulos", Estilo.body16, C.texto, Modifier.weight(1f))
                SelectorMenu(opciones, estilo.fuenteTitulos ?: "", { alCambiar(estilo.copy(fuenteTitulos = it.ifEmpty { null })) }, "Títulos")
            }
            Row(verticalAlignment = Alignment.CenterVertically) {
                Texto("Texto", Estilo.body16, C.texto, Modifier.weight(1f))
                SelectorMenu(opciones, estilo.fuenteTexto ?: "", { alCambiar(estilo.copy(fuenteTexto = it.ifEmpty { null })) }, "Texto")
            }
        }
        Grupo("menu", "Menú", "line.3.horizontal") {
            val items = estilo.menu ?: emptyList()
            fun cambiarItem(i: Int, f: (EstiloTienda.ItemMenu) -> EstiloTienda.ItemMenu) {
                val m = (estilo.menu ?: return).toMutableList()
                if (i !in m.indices) return
                m[i] = f(m[i]); alCambiar(estilo.copy(menu = m))
            }
            Texto(if (items.isEmpty()) "Va el menú de la plantilla. Armá el tuyo para sumar pestañas como Nosotros." else "Así se ve el menú de tu tienda.", Estilo.footnote, C.texto2)
            items.forEachIndexed { i, it ->
                Column(Modifier.padding(vertical = 6.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                    Row(horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.CenterVertically) {
                        Entrada(it.t, { v -> cambiarItem(i) { x -> x.copy(t = v.take(30)) } }, "Texto de la pestaña",
                            Modifier.weight(1f).clip(RoundedCornerShape(10.dp)).background(C.crema2).padding(10.dp), estilo = Estilo.body16)
                        Box(Modifier.size(32.dp).alpha(if (i == 0) 0.4f else 1f).clickable(enabled = i > 0) {
                            val m = items.toMutableList(); val x = m[i]; m[i] = m[i - 1]; m[i - 1] = x; alCambiar(estilo.copy(menu = m))
                        }.semantics { contentDescription = "Subir ${it.t}" }, contentAlignment = Alignment.Center) { Icono("chevron.up", 14f, C.primarioTexto) }
                        Box(Modifier.size(32.dp).clickable { alCambiar(estilo.copy(menu = items.filterIndexed { j, _ -> j != i })) }
                            .semantics { contentDescription = "Quitar ${it.t}" }, contentAlignment = Alignment.Center) { Icono("trash", 14f, C.primarioTexto) }
                    }
                    SelectorMenu(destinos, if (it.a.startsWith("https://")) "link" else it.a, { d ->
                        fun nombre(a: String) = destinos.firstOrNull { x -> x.first == a }?.second?.replace("Categoría: ", "")
                        cambiarItem(i) { x ->
                            val viejo = nombre(x.a)
                            val a = if (d == "link") "https://" else d
                            val t = if (d != "link" && (x.t.isEmpty() || x.t == viejo)) nombre(d) ?: x.t else x.t
                            x.copy(t = t, a = a)
                        }
                    }, "Adónde lleva")
                    if (it.a.startsWith("https://")) {
                        Entrada(it.a, { v -> cambiarItem(i) { x -> x.copy(a = v.trim()) } }, "https://", Modifier.fillMaxWidth().clip(RoundedCornerShape(10.dp)).background(C.crema2).padding(10.dp),
                            estilo = Estilo.body16, teclado = KeyboardType.Uri, sinCorrector = true)
                    }
                }
                if (i < items.size - 1) Separador()
            }
            if (items.size < cat.maxMenu) {
                BotonTonal("Agregar pestaña", {
                    val usadas = items.map { it.a }.toSet()
                    val p = cat.paginas.firstOrNull { it.id !in usadas } ?: cat.paginas.getOrNull(1) ?: return@BotonTonal
                    alCambiar(estilo.copy(menu = items + EstiloTienda.ItemMenu(p.nombre, p.id)))
                }, alto = 40.dp, simbolo = "plus")
            }
            if (items.isNotEmpty()) BotonTexto("Volver al menú de la plantilla", { alCambiar(estilo.copy(menu = null)) })
        }
        Grupo("titulos", "Títulos de las secciones", "textformat.size") {
            CatalogoEstilo.titulos.forEach { (id, nombre) ->
                CampoTexto("$nombre (vacío: el de la plantilla)", estilo.textos?.get(id) ?: "", { v -> alCambiar(estilo.copy(textos = (estilo.textos ?: emptyMap()) + (id to v.take(60)))) })
            }
        }
        Grupo("anuncio", "Anuncio y pie", "megaphone") {
            CampoTexto("Barra de anuncio (arriba de todo)", estilo.anuncio ?: "", { alCambiar(estilo.copy(anuncio = it.take(120))) })
            CampoTexto("Link del anuncio (opcional)", estilo.anuncioLink ?: "", { alCambiar(estilo.copy(anuncioLink = it.trim())) }, teclado = KeyboardType.Uri)
            CampoTexto("Línea del pie", estilo.leyendaPie ?: "", { alCambiar(estilo.copy(leyendaPie = it.take(160))) })
        }
        Grupo("redes", "Redes y contacto", "at") {
            CatalogoEstilo.redes.forEach { r ->
                CampoTexto(r.nombre, estilo.redes?.get(r.id) ?: "", { v -> alCambiar(estilo.copy(redes = (estilo.redes ?: emptyMap()) + (r.id to v))) }, teclado = r.teclado)
            }
            Texto("Se muestran en el pie de tu tienda.", Estilo.footnote, C.texto2)
        }
    }
}

// MARK: - Vista previa

/** La tienda de verdad con el borrador (la misma página que el editor web, /panel/mi-tienda/vista, con el token de la
 *  app). La página misma no deja navegar los links. */
@SuppressLint("SetJavaScriptEnabled")
@Composable
fun VistaPreviaTienda(token: String?, vuelta: Int, modifier: Modifier = Modifier) {
    var cargada by remember { mutableStateOf(-1) }
    AndroidView({ c ->
        WebView(c).apply {
            settings.javaScriptEnabled = true
            settings.domStorageEnabled = true
            setBackgroundColor(android.graphics.Color.TRANSPARENT)
            webViewClient = WebViewClient()
        }
    }, modifier, update = { w ->
        if (cargada != vuelta) {
            cargada = vuelta
            w.clearCache(true)
            w.loadUrl("${Web.base}/panel/mi-tienda/vista", token?.let { mapOf("Authorization" to "Bearer $it") } ?: emptyMap())
        }
    })
}
