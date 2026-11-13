package ar.com.tulocalweb.app.pantallas.productos

import android.Manifest
import android.content.ActivityNotFoundException
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
import androidx.compose.foundation.layout.ColumnScope
import androidx.compose.foundation.layout.IntrinsicSize
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxHeight
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
import androidx.compose.runtime.mutableStateListOf
import androidx.compose.runtime.mutableStateMapOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.drawBehind
import androidx.compose.ui.geometry.CornerRadius
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.PathEffect
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.selected
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.input.KeyboardCapitalization
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.app.BarraDe
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.app.LocalNavegacion
import ar.com.tulocalweb.app.nucleo.AtributoProducto
import ar.com.tulocalweb.app.nucleo.ErrorApi
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.Fotos
import ar.com.tulocalweb.app.nucleo.TipoProducto
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
import kotlinx.coroutines.launch
import kotlin.math.roundToInt

// Clon de Pantallas/Productos/EditorProducto.swift.

private data class Opcion(val nombre: String, val valores: List<String>)
private data class Combinacion(val stock: String = "", val precio: String = "", val sku: String = "")

/** Sin tipo elegido, las de siempre. */
private val curvas: List<Pair<String, List<String>>> = listOf("Talle" to listOf("XS", "S", "M", "L", "XL", "XXL"), "Color" to emptyList(), "Presentación" to emptyList())

/** 11c · Cargar o editar un producto: primero qué es (ropa, comida, un mueble…), y según eso sus datos (talle y color,
 *  contenido, peso y medidas…); fotos (cámara o galería), nombre, precio, costo, stock, SKU, categoría y descripción.
 *  Edita el producto entero que da la API, así que lo que la app no muestra (opciones, variantes, canales) vuelve tal
 *  cual. */
@Composable
fun EditorProducto(id: String?, sku: String? = null) {
    val estado = LocalEstado.current
    val nav = LocalNavegacion.current
    val ctx = LocalContext.current
    val alcance = rememberCoroutineScope()
    var original by remember { mutableStateOf<Map<String, ValorJSON>>(emptyMap()) }
    var cargando by remember { mutableStateOf(true) }
    var guardando by remember { mutableStateOf(false) }
    var nombre by remember { mutableStateOf("") }
    var precio by remember { mutableStateOf("") }
    var costo by remember { mutableStateOf("") }
    var stock by remember { mutableStateOf("") }
    var codigo by remember { mutableStateOf("") }
    var categoria by remember { mutableStateOf("") }
    var descripcion by remember { mutableStateOf("") }
    val fotos = remember { mutableStateListOf<String>() }
    var subiendo by remember { mutableIntStateOf(0) }
    var borrar by remember { mutableStateOf(false) }
    val opciones = remember { mutableStateListOf<Opcion>() }
    val variantes = remember { mutableStateMapOf<String, Combinacion>() }
    var teniaOpciones by remember { mutableStateOf(false) }
    val nuevoValor = remember { mutableStateMapOf<Int, String>() }
    var tipo by remember { mutableStateOf<TipoProducto?>(null) }
    val atributos = remember { mutableStateMapOf<String, String>() }
    var eligiendoTipo by remember { mutableStateOf(false) }

    fun num(s: String) = s.filter { it.isDigit() }.toIntOrNull()
    val valido = nombre.trim().isNotEmpty() && (num(precio) ?: 0) > 0
    val margen: String? = run {
        val p = num(precio) ?: 0; val c = num(costo) ?: 0
        if (p > 0 && c > 0 && p > c) "Ganás ${Formato.pesos(p - c)} por unidad (${((p - c).toDouble() / p * 100).roundToInt()} %)" else null
    }

    fun clave(c: Map<String, String>) = opciones.joinToString(" · ") { c[it.nombre] ?: "" }

    /** Las combinaciones de todas las opciones, en orden («M · Rojo»). */
    val combinaciones: List<Map<String, String>> = run {
        val validas = opciones.filter { it.nombre.trim().isNotEmpty() && it.valores.isNotEmpty() }
        if (validas.isEmpty()) emptyList() else {
            var out = listOf(emptyMap<String, String>())
            for (o in validas) out = out.flatMap { base -> o.valores.map { v -> base + (o.nombre to v) } }
            out.take(100)
        }
    }

    suspend fun cargar() {
        try {
            if (id == null || original.isNotEmpty()) {
                if (sku != null && codigo.isEmpty()) codigo = sku
                if (id == null && tipo == null) eligiendoTipo = true
                return
            }
            val v = estado.api.opcional<ValorJSON>("/v1/panel/productos/$id")
            val o = v?.objeto
            if (v == null || o == null) { estado.avisar("No pudimos traer el producto", ok = false); return }
            original = o
            nombre = v["nombre"]?.texto ?: ""
            precio = v["precio"]?.entero?.toString() ?: ""
            costo = v["costo"]?.entero?.let { if (it > 0) it.toString() else "" } ?: ""
            stock = v["stock"]?.entero?.toString() ?: ""
            codigo = v["sku"]?.texto ?: ""
            categoria = v["categoria"]?.texto ?: ""
            descripcion = v["descripcion"]?.texto ?: ""
            fotos.clear(); fotos.addAll(v["fotos"]?.lista?.mapNotNull { it.texto } ?: listOfNotNull(v["foto"]?.texto))
            opciones.clear(); opciones.addAll(v["opciones"]?.lista?.mapNotNull { x -> x["nombre"]?.texto?.let { n -> Opcion(n, x["valores"]?.lista?.mapNotNull { it.texto } ?: emptyList()) } } ?: emptyList())
            teniaOpciones = opciones.isNotEmpty()
            tipo = TipoProducto.de(v["tipo"]?.texto) ?: TipoProducto.adivinar(opciones.associate { it.nombre to it.valores })
            v["atributos"]?.objeto?.let { a -> a.forEach { (k, x) -> x.texto?.let { atributos[k] = it } } }
            for (x in v["variantes"]?.lista ?: emptyList()) {
                val vals = x["valores"]?.objeto ?: continue
                val c = vals.mapNotNull { (k, y) -> y.texto?.let { k to it } }.toMap()
                variantes[clave(c)] = Combinacion(x["stock"]?.entero?.toString() ?: "", x["precio"]?.entero?.toString() ?: "", x["sku"]?.texto ?: "")
            }
        } finally { cargando = false }
    }

    suspend fun subir(uri: Uri) {
        val datos = Fotos.achicar(ctx, uri, 1600, 85) ?: return
        subiendo += 1
        try { fotos.add(Web.subirFoto(datos, estado.api.token)) } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
            estado.avisar("No pudimos subir la foto", ok = false)
        } finally { subiendo -= 1 }
    }

    suspend fun guardar() {
        guardando = true
        try {
            val o = original.toMutableMap()
            // Lo que la API da de más (id, slug, publicado, vendidos…) no se manda: sólo lo editable.
            listOf("id", "slug", "publicado", "vendidos30d", "estado").forEach { o.remove(it) }
            o["nombre"] = Valor.texto(nombre.trim())
            o["precio"] = Valor.numero(num(precio) ?: 0)
            o["costo"] = Valor.numero(num(costo) ?: 0)
            val combos = combinaciones
            if (combos.isEmpty()) {
                o["stock"] = Valor.numero(num(stock) ?: 0)
                if (teniaOpciones) { o["opciones"] = Valor.lista(emptyList()); o["variantes"] = Valor.lista(emptyList()) } else { o.remove("opciones"); o.remove("variantes") }
            } else {
                o["opciones"] = Valor.lista(opciones.filter { it.nombre.isNotEmpty() && it.valores.isNotEmpty() }.map {
                    Valor.objeto(mapOf("nombre" to Valor.texto(it.nombre), "valores" to Valor.lista(it.valores.map { v -> Valor.texto(v) })))
                })
                o["variantes"] = Valor.lista(combos.map { c ->
                    val x = variantes[clave(c)] ?: Combinacion()
                    val v = mutableMapOf("valores" to Valor.objeto(c.mapValues { Valor.texto(it.value) }), "stock" to Valor.numero(num(x.stock) ?: 0))
                    v["precio"] = num(x.precio)?.let { Valor.numero(it) } ?: Valor.nulo
                    if (x.sku.isNotEmpty()) v["sku"] = Valor.texto(x.sku)
                    Valor.objeto(v)
                })
                o.remove("stock")
            }
            o["sku"] = Valor.texto(codigo.trim())
            o["categoria"] = categoria.trim().ifEmpty { null }?.let { Valor.texto(it) } ?: Valor.nulo
            o["descripcion"] = Valor.texto(descripcion)
            o["fotos"] = Valor.lista(fotos.map { Valor.texto(it) })
            o["foto"] = fotos.firstOrNull()?.let { Valor.texto(it) } ?: Valor.nulo
            tipo?.let { t ->
                o["tipo"] = Valor.texto(t.name)
                // Sólo los datos de este tipo: lo cargado con otro tipo se va al cambiarlo.
                val campos = t.campos.map { it.name }.toSet()
                o["atributos"] = Valor.objeto(atributos.filter { it.key in campos && it.value.trim().isNotEmpty() }.mapValues { Valor.texto(it.value.trim()) })
            }
            if (id != null) {
                estado.api.hacer("PUT", "/v1/panel/productos/$id", Valor.objeto(o))
                estado.avisar("Guardamos los cambios")
            } else {
                estado.api.hacer("POST", "/v1/panel/productos", Valor.objeto(o))
                estado.avisar("¡Producto cargado!")
            }
            Haptico.exito()
            nav.volver()
        } catch (e: ErrorApi) { estado.avisar(e.detalle, ok = false) } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
            estado.avisar("No pudimos guardarlo", ok = false)
        } finally { guardando = false }
    }

    suspend fun eliminar() {
        val i = id ?: return
        try {
            estado.api.hacer("DELETE", "/v1/panel/productos/$i")
            estado.avisar("Borramos el producto")
            nav.volver()
        } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) { estado.avisar("No pudimos borrarlo", ok = false) }
    }

    // La cámara del sistema (UIImagePickerController) y la galería (PhotosPicker, hasta 8).
    // Guardado aparte: si el sistema cierra la app con la cámara abierta, la foto vuelve igual.
    var fotoCamara by rememberSaveable { mutableStateOf<Uri?>(null) }
    val camara = rememberLauncherForActivityResult(ActivityResultContracts.TakePicture()) { ok ->
        val u = fotoCamara
        if (ok && u != null) alcance.launch { subir(u) }
    }
    val galeria = rememberLauncherForActivityResult(ActivityResultContracts.PickMultipleVisualMedia(8)) { uris ->
        val libres = 8 - fotos.size - subiendo
        alcance.launch { uris.take(maxOf(0, libres)).forEach { subir(it) } }
    }

    val t = estado.tienda
    val pub = original["publicado"]?.bool != false
    BarraDe(acciones = if (id != null && !cargando && t != null && pub) {
        {
            val u = t.url(original["slug"]?.texto ?: id)
            MenuBoton(listOf(
                OpcionDeMenu("Compartir", "square.and.arrow.up") { compartirTexto(ctx, Compartir.mensajeProducto(nombre, num(precio), u)) },
                OpcionDeMenu("Copiar link", "doc.on.doc") { Compartir.copiar(ctx, u, estado, "Copiamos el link del producto") },
                OpcionDeMenu("Ver en la tienda", "arrow.up.right.square") { abrirEnNavegador(ctx, u) },
            ), "Compartir el producto") { Box(Modifier.size(44.dp), contentAlignment = Alignment.Center) { Icono("square.and.arrow.up", 19f, C.primarioTexto) } }
        }
    } else null)

    @Composable
    fun Campo(titulo: String, modifier: Modifier = Modifier, contenido: @Composable () -> Unit) {
        val forma = RoundedCornerShape(14.dp)
        Column(modifier.fillMaxWidth(), verticalArrangement = Arrangement.spacedBy(6.dp)) {
            Texto(titulo, Estilo.footnoteFuerte, C.texto2)
            Box(Modifier.fillMaxWidth().clip(forma).background(C.superficie).border(1.dp, C.linea2, forma).padding(horizontal = 14.dp, vertical = 12.dp)) { contenido() }
        }
    }

    @Composable
    fun EntradaAtributo(a: AtributoProducto, ejemplo: String? = null, modifier: Modifier = Modifier) {
        val valor = atributos[a.name] ?: ""
        val cambiar = { n: String -> atributos[a.name] = if (a.unidad == null) n.take(60) else n.filter { it.isDigit() }.take(8) }
        Row(modifier.semantics { contentDescription = a.unidad?.let { "${a.nombre} en ${if (it == "g") "gramos" else "centímetros"}" } ?: a.nombre },
            horizontalArrangement = Arrangement.spacedBy(6.dp), verticalAlignment = Alignment.CenterVertically) {
            if (a.unidad == null) Entrada(valor, cambiar, ejemplo ?: a.ejemplo, Modifier.weight(1f), estilo = Estilo.body16)
            else {
                Entrada(valor, cambiar, ejemplo ?: "0", Modifier.weight(1f), estilo = Estilo.body16, teclado = KeyboardType.Number)
                if (!a.esMedida) Texto(a.unidad ?: "", Estilo.footnoteFuerte, C.texto2)
            }
        }
    }

    ConBarraAccion({
        if (!eligiendoTipo) BotonPrincipal(if (guardando) "Guardando…" else (if (id == null) "Cargar producto" else "Guardar cambios"), { alcance.launch { guardar() } },
            habilitado = valido && !guardando && subiendo == 0)
    }, Modifier.background(C.fondo).imePadding()) { abajo ->
        Column(Modifier.fillMaxSize().verticalScroll(rememberScrollState()).margenPantalla().padding(bottom = 32.dp + abajo), verticalArrangement = Arrangement.spacedBy(14.dp)) {
            Texto(if (id == null) "Producto nuevo" else "Editar producto", Estilo.tituloGrande, C.texto)
            if (cargando) {
                Box(Modifier.padding(top = 12.dp)) { EsqueletoFilas(4, 0.dp) }
            } else if (eligiendoTipo) {
                // MARK: Qué es
                Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
                    Texto("¿Qué vas a vender?", Estilo.headline, C.texto)
                    Texto(Formato.sinViudas("Así te pedimos sólo los datos que van con tu producto."), Estilo.footnote, C.texto2)
                    TipoProducto.entries.chunked(2).forEach { par ->
                        Row(Modifier.height(IntrinsicSize.Min), horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                            par.forEach { x ->
                                val forma = RoundedCornerShape(16.dp)
                                Column(Modifier.weight(1f).fillMaxHeight().heightIn(min = 128.dp).tarjeta(radio = 16.dp).border(1.5.dp, if (tipo == x) C.acento else Color.Transparent, forma)
                                    .clickable { Haptico.seleccion(); tipo = x; eligiendoTipo = false }.semantics { selected = tipo == x }.padding(14.dp),
                                    verticalArrangement = Arrangement.spacedBy(6.dp)) {
                                    Box(Modifier.size(40.dp).clip(RoundedCornerShape(12.dp)).background(C.crema), contentAlignment = Alignment.Center) { Icono(x.simbolo, 20f, C.primarioTexto) }
                                    Texto(x.nombre, Estilo.subheadFuerte, C.texto)
                                    Texto(x.ejemplo, Estilo.footnote, C.texto2)
                                }
                            }
                            if (par.size == 1) Spacer(Modifier.weight(1f))
                        }
                    }
                }
            } else {
                // Fila del tipo
                Row(Modifier.fillMaxWidth().tarjeta(radio = 16.dp).padding(horizontal = 14.dp, vertical = 10.dp), horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
                    Box(Modifier.size(38.dp).clip(RoundedCornerShape(11.dp)).background(C.crema), contentAlignment = Alignment.Center) { Icono(tipo?.simbolo ?: "questionmark", 16f, C.primarioTexto) }
                    Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(1.dp)) {
                        Texto("Tipo de producto", Estilo.footnote, C.texto2)
                        Texto(tipo?.nombre ?: "Sin elegir", Estilo.fila, C.texto)
                    }
                    BotonTonal(if (tipo == null) "Elegir" else "Cambiar", { eligiendoTipo = true })
                }
                // Fotos
                Row(Modifier.fillMaxWidth().horizontalScroll(rememberScrollState()), horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                    fotos.toList().forEach { f ->
                        Box {
                            Miniatura(f, 96.dp, radio = 16.dp)
                            Box(Modifier.align(Alignment.TopEnd).padding(5.dp).size(22.dp).clip(CircleShape).background(Color.Black.copy(alpha = 0.55f))
                                .clickable { fotos.remove(f) }.semantics { contentDescription = "Sacar foto" }, contentAlignment = Alignment.Center) { Icono("xmark", 10f, Color.White) }
                        }
                    }
                    repeat(subiendo) { Box(Modifier.size(96.dp).clip(RoundedCornerShape(16.dp)).background(C.pista), contentAlignment = Alignment.Center) { Rueda(C.acento) } }
                    if (fotos.size + subiendo < 8) {
                        BotonFoto("camera", "Sacar foto") {
                            alcance.launch {
                                if (!estado.alPedirPermiso(Manifest.permission.CAMERA)) {
                                    estado.avisar("Para sacar fotos, dale permiso a la cámara en los ajustes del teléfono.", ok = false); return@launch
                                }
                                try {
                                    val (_, u) = Fotos.archivoCamara(ctx); fotoCamara = u; estado.abrioOtraApp = true; camara.launch(u)
                                } catch (e: ActivityNotFoundException) {
                                    estado.avisar("No encontramos una app de cámara. Elegí una foto de la galería.", ok = false)
                                } catch (e: SecurityException) {
                                    estado.avisar("Para sacar fotos, dale permiso a la cámara en los ajustes del teléfono.", ok = false)
                                }
                            }
                        }
                        BotonFoto("photo.on.rectangle", "Galería") { estado.abrioOtraApp = true; galeria.launch(PickVisualMediaRequest(ActivityResultContracts.PickVisualMedia.ImageOnly)) }
                    }
                }
                Campo("Nombre") { Entrada(nombre, { nombre = it }, ejemploNombre(tipo), estilo = Estilo.body16) }
                Row(horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                    Campo("Precio", Modifier.weight(1f)) { Entrada(precio, { precio = it }, "0", estilo = Estilo.body16, teclado = KeyboardType.Number) }
                    Campo("Costo (opcional)", Modifier.weight(1f)) { Entrada(costo, { costo = it }, "0", estilo = Estilo.body16, teclado = KeyboardType.Number) }
                }
                margen?.let { Texto(it, Estilo.footnoteFuerte, C.okTinta, Modifier.padding(horizontal = 4.dp)) }
                if (opciones.isEmpty()) Campo("Stock") { Entrada(stock, { stock = it }, "0", estilo = Estilo.body16, teclado = KeyboardType.Number) }
                // Datos del tipo: texto libre de a uno, peso con su unidad y las tres medidas juntas.
                tipo?.let { tp ->
                    val campos = tp.campos
                    Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
                        Texto("Datos del producto", Estilo.headline, C.texto, Modifier.padding(top = 4.dp))
                        campos.filter { !it.esMedida }.forEach { a -> Campo(a.nombre) { EntradaAtributo(a) } }
                        if (AtributoProducto.alto in campos) {
                            Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
                                Texto("Medidas en cm (alto × ancho × largo)", Estilo.footnoteFuerte, C.texto2)
                                Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                                    listOf(AtributoProducto.alto, AtributoProducto.ancho, AtributoProducto.largo).forEach { a ->
                                        val forma = RoundedCornerShape(14.dp)
                                        EntradaAtributo(a, a.nombre, Modifier.weight(1f).clip(forma).background(C.superficie).border(1.dp, C.linea2, forma).padding(12.dp))
                                    }
                                }
                            }
                        }
                        if (AtributoProducto.peso in campos) Texto(Formato.sinViudas("El peso y las medidas sirven para cotizar el envío."), Estilo.footnote, C.texto2, Modifier.padding(horizontal = 4.dp))
                    }
                }
                // Variantes y stock
                Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
                    val sugeridas = (tipo?.opciones ?: curvas).filter { s -> opciones.none { it.nombre == s.first } }
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        Texto("Variantes y stock", Estilo.headline, C.texto, Modifier.weight(1f))
                        if (opciones.size < 3) {
                            MenuBoton(sugeridas.map { c -> OpcionDeMenu(c.first, "") { if (opciones.none { it.nombre == c.first }) opciones.add(Opcion(c.first, c.second)) } } +
                                OpcionDeMenu("Otra opción", "") { opciones.add(Opcion("", emptyList())) }, "Agregar") {
                                Row(horizontalArrangement = Arrangement.spacedBy(4.dp), verticalAlignment = Alignment.CenterVertically) {
                                    Icono("plus", 14f, C.primarioTexto); Texto("Agregar", Estilo.subheadFuerte, C.primarioTexto)
                                }
                            }
                        }
                    }
                    if (opciones.isEmpty()) Texto(Formato.sinViudas(textoVariantes(tipo)), Estilo.footnote, C.texto2)
                    opciones.indices.toList().forEach { i ->
                        val o = opciones.getOrNull(i) ?: return@forEach
                        Column(Modifier.fillMaxWidth().tarjeta(radio = 16.dp).padding(14.dp), verticalArrangement = Arrangement.spacedBy(10.dp)) {
                            Row(verticalAlignment = Alignment.CenterVertically) {
                                Entrada(o.nombre, { opciones[i] = o.copy(nombre = it) }, "Nombre (Talle, Color…)", Modifier.weight(1f), estilo = Estilo.fila)
                                Box(Modifier.size(36.dp).clickable { opciones.removeAt(i) }.semantics { contentDescription = "Sacar opción" }, contentAlignment = Alignment.Center) { Icono("trash", 16f, C.errorTinta) }
                            }
                            FlujoChips(o.valores) { v -> opciones[i] = o.copy(valores = o.valores - v) }
                            val sumarValor = {
                                val v = (nuevoValor[i] ?: "").trim()
                                if (v.isNotEmpty() && o.valores.none { it.equals(v, ignoreCase = true) }) { opciones[i] = o.copy(valores = o.valores + v.take(20)); nuevoValor[i] = "" }
                            }
                            Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                                Entrada(nuevoValor[i] ?: "", { nuevoValor[i] = it }, "+ Valor (ej.: ${ejemploValor(o.nombre)})", Modifier.weight(1f), estilo = Estilo.body16,
                                    accionTeclado = ImeAction.Done, alEnviar = sumarValor)
                                BotonTonal("Agregar", sumarValor, habilitado = (nuevoValor[i] ?: "").trim().isNotEmpty())
                            }
                        }
                    }
                    if (combinaciones.isNotEmpty()) {
                        Column(Modifier.fillMaxWidth().tarjeta(radio = 16.dp)) {
                            Row(Modifier.padding(horizontal = 14.dp, vertical = 10.dp)) {
                                Texto("Variante", Estilo.footnoteFuerte, C.texto2, Modifier.weight(1f))
                                Texto("Stock", Estilo.footnoteFuerte, C.texto2, Modifier.width(64.dp), alinear = TextAlign.Center)
                                Texto("Precio", Estilo.footnoteFuerte, C.texto2, Modifier.width(92.dp), alinear = TextAlign.Center)
                            }
                            combinaciones.forEach { c ->
                                val k = clave(c)
                                Separador()
                                Row(Modifier.padding(horizontal = 14.dp, vertical = 8.dp), verticalAlignment = Alignment.CenterVertically) {
                                    Texto(k, Estilo.fila, C.texto, Modifier.weight(1f), lineas = 1)
                                    val x = variantes[k] ?: Combinacion()
                                    Entrada(x.stock, { variantes[k] = x.copy(stock = it) }, "0", Modifier.width(64.dp).clip(RoundedCornerShape(8.dp))
                                        .background(if ((x.stock.toIntOrNull() ?: 0) == 0) C.errorFondo else C.pista).padding(vertical = 6.dp),
                                        estilo = Estilo.body16, teclado = KeyboardType.Number, alinear = TextAlign.Center, numeros = true)
                                    Spacer(Modifier.width(8.dp))
                                    Entrada(x.precio, { variantes[k] = (variantes[k] ?: Combinacion()).copy(precio = it) }, precio.ifEmpty { "Igual" },
                                        Modifier.width(92.dp).clip(RoundedCornerShape(8.dp)).background(C.pista).padding(vertical = 6.dp),
                                        estilo = Estilo.body16, teclado = KeyboardType.Number, alinear = TextAlign.Center, numeros = true)
                                }
                            }
                        }
                        Texto("${Formato.cantidad(combinaciones.size, "combinación", "combinaciones")} · stock total ${combinaciones.sumOf { (variantes[clave(it)]?.stock?.toIntOrNull() ?: 0) }}",
                            Estilo.footnote, C.texto2, Modifier.padding(horizontal = 4.dp))
                    }
                }
                Row(horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                    Campo("SKU o código", Modifier.weight(1f)) { Entrada(codigo, { codigo = it }, "Opcional", estilo = Estilo.body16, mayusculas = KeyboardCapitalization.Characters) }
                    Campo("Categoría", Modifier.weight(1f)) { Entrada(categoria, { categoria = it }, "Opcional", estilo = Estilo.body16) }
                }
                Campo("Descripción") { Entrada(descripcion, { descripcion = it }, "Contá cómo es, de qué está hecho, medidas…", estilo = Estilo.body16, lineas = 3, lineasMaximas = 8) }
                if (id != null) BotonTexto("Borrar producto", { borrar = true }, color = C.errorTinta)
            }
        }
    }
    LaunchedEffect(Unit) { cargar() }
    Confirmacion(borrar, "¿Borrás «$nombre»?", "Deja de verse en la tienda. Los pedidos que ya tiene no cambian.",
        listOf(OpcionConfirmacion("Borrar", destructiva = true) { alcance.launch { eliminar() } }), { borrar = false }, volver = "Cancelar")
}

private fun ejemploNombre(tipo: TipoProducto?) = when (tipo) {
    TipoProducto.ropa -> "Ej.: Remera de algodón"
    TipoProducto.calzado -> "Ej.: Zapatillas urbanas"
    TipoProducto.alimentos -> "Ej.: Yerba orgánica"
    TipoProducto.belleza -> "Ej.: Crema de manos"
    TipoProducto.hogar -> "Ej.: Mesa ratona de pino"
    TipoProducto.tecnologia -> "Ej.: Auriculares inalámbricos"
    else -> "Ej.: Mate de calabaza"
}

private fun textoVariantes(tipo: TipoProducto?): String {
    if (tipo == null) return "¿Viene en talles o colores? Agregá las opciones y cargás el stock de cada combinación."
    val ns = tipo.opciones.map { it.first.lowercase() }
    return "¿Tiene opciones como ${ns.joinToString(" o ")}? Agregalas y cargás el stock de cada combinación."
}

private fun ejemploValor(opcion: String) = when (opcion.lowercase()) {
    "color" -> "Rojo"
    "talle" -> "M"
    "presentación" -> "500 g"
    "sabor" -> "Limón"
    "fragancia" -> "Vainilla"
    "tamaño" -> "Grande"
    "capacidad" -> "128 GB"
    else -> "Clásico"
}

@Composable
private fun BotonFoto(simbolo: String, texto: String, alTocar: () -> Unit) {
    val linea = C.linea2
    Column(Modifier.size(96.dp).clip(RoundedCornerShape(16.dp)).background(C.crema2).drawBehind {
        drawRoundRect(linea, cornerRadius = CornerRadius(16.dp.toPx()), style = Stroke(1.dp.toPx(), pathEffect = PathEffect.dashPathEffect(floatArrayOf(5.dp.toPx(), 4.dp.toPx()))))
    }.clickable(onClick = alTocar), horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(6.dp, Alignment.CenterVertically)) {
        Icono(simbolo, 22f, C.primarioTexto)
        Texto(texto, Estilo.footnoteFuerte, C.primarioTinta)
    }
}

/** Los valores de una opción como píldoras, con «×» para sacarlos. */
@Composable
fun FlujoChips(valores: List<String>, sacar: (String) -> Unit) {
    Row(Modifier.fillMaxWidth().horizontalScroll(rememberScrollState()), horizontalArrangement = Arrangement.spacedBy(6.dp)) {
        valores.forEach { v ->
            Row(Modifier.clip(Capsula).background(C.crema).padding(horizontal = 10.dp, vertical = 6.dp), horizontalArrangement = Arrangement.spacedBy(4.dp), verticalAlignment = Alignment.CenterVertically) {
                Texto(v, Estilo.footnoteFuerte, C.primarioTinta)
                Box(Modifier.clickable { sacar(v) }.semantics { contentDescription = "Sacar $v" }) { Icono("xmark", 9f, C.primarioTinta) }
            }
        }
    }
}
