package ar.com.tulocalweb.app.pantallas.mas

import android.net.Uri
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.PickVisualMediaRequest
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.gestures.detectDragGestures
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
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.key
import androidx.compose.runtime.mutableFloatStateOf
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
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.layout.onSizeChanged
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import androidx.compose.ui.zIndex
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.ProductoFila
import ar.com.tulocalweb.app.nucleo.Valor
import ar.com.tulocalweb.app.nucleo.ValorJSON
import ar.com.tulocalweb.app.nucleo.bool
import ar.com.tulocalweb.app.nucleo.entero
import ar.com.tulocalweb.app.nucleo.get
import ar.com.tulocalweb.app.nucleo.lista
import ar.com.tulocalweb.app.nucleo.objeto
import ar.com.tulocalweb.app.nucleo.texto
import ar.com.tulocalweb.app.tema.*
import coil3.compose.AsyncImage
import kotlinx.coroutines.launch
import kotlinx.serialization.json.JsonObject
import java.time.LocalDate
import java.time.OffsetDateTime
import java.time.ZoneOffset
import java.time.format.DateTimeFormatter

// Clon de Pantallas/Mas/EditorBloques.swift.

/** Lo que necesitan los campos de un bloque: la tienda, sus categorías y productos, y cómo subir una foto. */
class ContextoCampos(
    val nombreTienda: String = "",
    val lema: String = "",
    val categorias: List<String> = emptyList(),
    val productos: List<ProductoFila> = emptyList(),
    val hayResenas: Boolean = true,
    val subir: suspend (Uri) -> String? = { null },
)

val LocalContextoCampos = staticCompositionLocalOf { ContextoCampos() }

/** Los íconos de los beneficios (bloques.json › iconos) con su símbolo y su nombre. */
val iconosBloque = mapOf(
    "envio" to ("shippingbox" to "Envío"), "cuotas" to ("creditcard" to "Cuotas"), "transferencia" to ("banknote" to "Transferencia"),
    "cambios" to ("arrow.triangle.2.circlepath" to "Cambios"), "retiro" to ("storefront" to "Retiro"), "seguro" to ("checkmark.shield" to "Compra segura"),
    "regalo" to ("gift" to "Regalo"), "hecho-a-mano" to ("hand.raised" to "Hecho a mano"), "reloj" to ("clock" to "Horario"), "estrella" to ("star" to "Destacado"),
)

/** Las páginas de la tienda a las que puede llevar un link (tiposCampo › link). */
private val paginasLink = listOf("inicio" to "Inicio", "catalogo" to "Catálogo", "nosotros" to "Nosotros", "turnos" to "Turnos", "suscripciones" to "Suscripciones")

// MARK: - Qué le falta a un bloque

/** Por qué un bloque no se va a ver en la tienda (regla «Nada vacío»), con los mismos motivos que el editor web
 *  (apps/web/src/lib/panel/mi-tienda/faltan.ts › queFalta). Vacío: se ve. */
fun faltaEn(b: Bloque, ctx: ContextoCampos): List<String> = listOfNotNull(queFalta(b, ctx))

private fun queFalta(b: Bloque, ctx: ContextoCampos): String? {
    val tipo = CatalogoBloques.compartido.tipo(b.tipo) ?: return "Este bloque ya no existe: borralo."
    for (c in tipo.campos) {
        val v = b[c.id]
        val n = v?.lista?.size ?: 0
        if (c.requerido && (v == null || v.vacio) && c.porDefecto == null) return "Le falta: ${c.nombre.lowercase()}."
        if ((c.min ?: 0) > 0 && n in 1 until (c.min ?: 0)) return "${c.nombre}: van por lo menos ${c.min}."
        if (c.tipo == "lista" && (c.min ?: 0) > 0 && n == 0) return "Sumá ${c.nombre.lowercase()}."
    }
    fun productos(f: Map<String, ValorJSON>?): String? {
        if (ctx.productos.isEmpty()) return "Todavía no cargaste productos."
        when (f?.get("fuente")?.texto) {
            "elegidos" -> if (f?.get("ids")?.lista?.mapNotNull { it.texto }?.none { id -> ctx.productos.any { it.id == id } } != false) return "Elegí qué productos van."
            "categoria" -> if (f?.get("categoria")?.texto.isNullOrEmpty()) return "Elegí la categoría."
        }
        return null
    }
    val d = b.datos
    return when (b.tipo) {
        "productos" -> productos(d["productos"]?.objeto ?: tipo.campos.firstOrNull { it.id == "productos" }?.porDefecto?.objeto)
        "oferta" -> {
            val hasta = d["hasta"]?.texto?.let { runCatching { OffsetDateTime.parse(it) }.getOrNull() }
            if (hasta != null && !hasta.isAfter(OffsetDateTime.now())) "La oferta ya terminó: cambiá la fecha."
            else productos(d["productos"]?.objeto ?: mapOf("fuente" to Valor.texto("ofertas")))
        }
        "destacado" -> if (ctx.productos.any { it.id == d["producto"]?.texto }) null else "Ese producto ya no está en la tienda: elegí otro."
        "combo" -> if ((d["productos"]?.lista?.mapNotNull { it.texto }?.count { id -> ctx.productos.any { it.id == id } } ?: 0) >= 2) null else "Elegí por lo menos 2 productos."
        "categorias" -> if (ctx.categorias.isNotEmpty()) null else "Tus productos todavía no tienen categorías."
        "menu" -> if (ctx.productos.isNotEmpty()) null else "Todavía no cargaste productos."
        "resenas" -> if (ctx.hayResenas) null else "Todavía no tenés reseñas: aparecen con la primera."
        "galeria" -> if ((d["fotos"] == null || d["fotos"]!!.vacio) && ctx.productos.none { !it.foto.isNullOrEmpty() }) "Subí fotos o cargá productos con foto." else null
        "video" -> if (Regex("^https://(www\\.)?(youtube\\.com|youtu\\.be|vimeo\\.com)/").containsMatchIn(d["url"]?.texto ?: "")) null else "El link tiene que ser de YouTube o Vimeo."
        else -> null
    }
}

/** El nombre corto de un bloque en la lista: el título que le pusieron o el del tipo. */
fun nombreDe(b: Bloque): String = CatalogoBloques.compartido.tipo(b.tipo)?.nombre ?: b.tipo

// MARK: - Lista de bloques de una página

/** Los bloques de una página: arrastrar del asa para ordenar, tocar para editar y el menú con lo demás. */
@Composable
fun ListaBloques(
    bloques: List<Bloque>,
    alMover: (desde: Int, hasta: Int) -> Unit,
    alEditar: (Bloque) -> Unit,
    alAgregar: (posicion: Int) -> Unit,
    alDuplicar: (Bloque) -> Unit,
    alOcultar: (Bloque) -> Unit,
    alBorrar: (Bloque) -> Unit,
) {
    val ctx = LocalContextoCampos.current
    var arrastrado by remember { mutableStateOf<String?>(null) }
    var corrimiento by remember { mutableFloatStateOf(0f) }
    val altos = remember { mutableStateMapOf<String, Int>() }
    val actuales by rememberUpdatedState(bloques)
    val mover by rememberUpdatedState(alMover)

    Column(Modifier.fillMaxWidth().tarjeta(radio = 20.dp)) {
        bloques.forEachIndexed { i, b ->
            key(b.id) {
                val tipo = CatalogoBloques.compartido.tipo(b.tipo)
                val faltan = faltaEn(b, ctx)
                val esEste = arrastrado == b.id
                Column(Modifier.zIndex(if (esEste) 1f else 0f).onSizeChanged { altos[b.id] = it.height }
                    .graphicsLayer { translationY = if (esEste) corrimiento else 0f; shadowElevation = if (esEste) 12f else 0f }
                    .background(if (esEste) C.elevado else androidx.compose.ui.graphics.Color.Transparent)) {
                    if (i > 0) Separador()
                    Row(Modifier.fillMaxWidth().heightIn(min = 64.dp).clickable { alEditar(b) }.padding(start = 4.dp, end = 8.dp, top = 8.dp, bottom = 8.dp),
                        horizontalArrangement = Arrangement.spacedBy(10.dp), verticalAlignment = Alignment.CenterVertically) {
                        // El asa: se arrastra enseguida (sin mantener apretado), como el modo edición de iOS.
                        Box(Modifier.size(36.dp, 48.dp).semantics { contentDescription = "Mover ${tipo?.nombre ?: b.tipo}" }.pointerInput(b.id) {
                            detectDragGestures(
                                onDragStart = { arrastrado = b.id; corrimiento = 0f; Haptico.impacto() },
                                onDragEnd = { arrastrado = null; corrimiento = 0f },
                                onDragCancel = { arrastrado = null; corrimiento = 0f },
                            ) { c, d ->
                                c.consume()
                                corrimiento += d.y
                                val lista = actuales
                                val j = lista.indexOfFirst { it.id == b.id }
                                if (corrimiento > 0 && j in 0 until lista.size - 1) {
                                    val h = altos[lista[j + 1].id] ?: return@detectDragGestures
                                    if (corrimiento > h / 2f) { mover(j, j + 1); corrimiento -= h; Haptico.seleccion() }
                                } else if (corrimiento < 0 && j > 0) {
                                    val h = altos[lista[j - 1].id] ?: return@detectDragGestures
                                    if (-corrimiento > h / 2f) { mover(j, j - 1); corrimiento += h; Haptico.seleccion() }
                                }
                            }
                        }, contentAlignment = Alignment.Center) { Icono("line.3.horizontal", 15f, C.texto3) }
                        MiniaturaBloque(b.tipo, b.variante, Modifier.width(64.dp).alpha(if (b.visible) 1f else 0.45f), alto = 44.dp)
                        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                            Texto(b["titulo"]?.texto?.takeIf { it.isNotBlank() } ?: nombreDe(b), Estilo.fila, if (b.visible) C.texto else C.texto3, lineas = 1)
                            Texto(if (!b.visible) "Oculto" else tipo?.nombreVariante(b.variante) ?: b.variante, Estilo.footnote, C.texto2, lineas = 1)
                            if (b.visible && faltan.isNotEmpty()) Texto(faltan.first(), Estilo.footnoteFuerte, C.avisoTinta, lineas = 2)
                        }
                        MenuBoton(listOf(
                            OpcionDeMenu("Editar", "square.and.pencil") { alEditar(b) },
                            OpcionDeMenu("Agregar un bloque abajo", "plus") { alAgregar(i + 1) },
                            OpcionDeMenu("Duplicar", "plus.square.on.square") { alDuplicar(b) },
                            OpcionDeMenu(if (b.visible) "Ocultar" else "Mostrar", if (b.visible) "eye.slash" else "eye") { alOcultar(b) },
                            OpcionDeMenu("Borrar", "trash", destructiva = true) { alBorrar(b) },
                        ), "Más opciones de ${tipo?.nombre ?: b.tipo}") {
                            Box(Modifier.size(36.dp), contentAlignment = Alignment.Center) { Icono("ellipsis", 17f, C.primarioTexto) }
                        }
                    }
                }
            }
        }
    }
}

// MARK: - Biblioteca

/** Elegir un bloque para sumar: por grupo, con la miniatura de su primera variante. Los que ya llegaron a su máximo
 *  en la página no se pueden sumar. */
@Composable
fun BibliotecaBloques(pagina: String, bloques: List<Bloque>, alElegir: (TipoBloque) -> Unit) {
    val cat = CatalogoBloques.compartido
    val lleno = bloques.size >= cat.limites.bloquesPorPagina
    Column(Modifier.fillMaxSize()) {
        Grabber()
        Column(Modifier.weight(1f).verticalScroll(rememberScrollState()).padding(horizontal = 20.dp).padding(top = 8.dp, bottom = 24.dp), verticalArrangement = Arrangement.spacedBy(14.dp)) {
            Texto("Sumar un bloque", Estilo.tituloHojaChico, C.texto)
            Texto("Cada bloque toma los colores y la letra de tu tienda. Después lo podés mover, ocultar o cambiar de variante.", Estilo.subhead, C.texto2)
            if (lleno) ChipSimple("Llegaste a ${cat.limites.bloquesPorPagina} bloques en esta página", C.avisoFondo, C.avisoTinta)
            cat.grupos.forEach { g ->
                val tipos = cat.bloques.filter { it.grupo == g.id }
                if (tipos.isNotEmpty()) {
                    Texto(g.nombre, Estilo.headline, C.texto)
                    tipos.chunked(2).forEach { par ->
                        Row(horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                            par.forEach { t ->
                                val usados = bloques.count { it.tipo == t.tipo }
                                val puede = !lleno && (t.max == null || usados < t.max)
                                Column(Modifier.weight(1f).alpha(if (puede) 1f else 0.45f).clip(RoundedCornerShape(14.dp)).clickable(enabled = puede) { alElegir(t) }
                                    .semantics { contentDescription = "Sumar ${t.nombre}" }, verticalArrangement = Arrangement.spacedBy(4.dp)) {
                                    MiniaturaBloque(t.tipo, t.variantes.firstOrNull() ?: "", Modifier.fillMaxWidth(), alto = 84.dp)
                                    Texto(t.nombre, Estilo.subheadFuerte, C.texto, lineas = 2)
                                    Texto(if (!puede && t.max != null && usados >= t.max) "Ya está en la página" else t.descripcion, Estilo.footnote, C.texto2, lineas = 3)
                                }
                            }
                            if (par.size == 1) Spacer(Modifier.weight(1f))
                        }
                    }
                }
            }
        }
    }
}

// MARK: - Panel del bloque

/** Editar un bloque: la variante con miniaturas, los campos de bloques.json y lo que le falta. */
@Composable
fun PanelBloque(
    bloque: Bloque,
    alCambiar: (Bloque, clave: String?) -> Unit,
    alDuplicar: () -> Unit,
    alBorrar: () -> Unit,
    alCerrar: () -> Unit,
) {
    val ctx = LocalContextoCampos.current
    val tipo = CatalogoBloques.compartido.tipo(bloque.tipo)
    val faltan = faltaEn(bloque, ctx)
    Column(Modifier.fillMaxSize().imePadding()) {
        Grabber()
        Row(Modifier.fillMaxWidth().padding(horizontal = 20.dp).padding(top = 6.dp), verticalAlignment = Alignment.CenterVertically) {
            Texto(tipo?.nombre ?: bloque.tipo, Estilo.tituloHojaChico, C.texto, Modifier.weight(1f))
            Texto("Listo", Estilo(17f, Peso.bold), C.primarioTexto, Modifier.clickable(onClick = alCerrar).padding(8.dp))
        }
        Column(Modifier.weight(1f).verticalScroll(rememberScrollState()).padding(horizontal = 20.dp).padding(top = 4.dp, bottom = 32.dp), verticalArrangement = Arrangement.spacedBy(14.dp)) {
            if (tipo == null) {
                Texto("Este bloque es de una versión más nueva. Actualizá la app para editarlo.", Estilo.subhead, C.texto2)
                return@Column
            }
            Texto(tipo.descripcion, Estilo.subhead, C.texto2)
            if (faltan.isNotEmpty()) Column(Modifier.fillMaxWidth().clip(RoundedCornerShape(14.dp)).background(C.avisoFondo).padding(12.dp), verticalArrangement = Arrangement.spacedBy(4.dp)) {
                Texto("Todavía no se ve en tu tienda", Estilo.subheadFuerte, C.avisoTinta)
                faltan.forEach { Texto(it, Estilo.footnote, C.avisoTinta) }
            }
            if (tipo.variantes.size > 1) {
                Texto("Variante", Estilo.footnoteFuerte, C.texto2)
                SelectorVariante(bloque.tipo, tipo.variantes, tipo.nombresVariantes, bloque.variante) { alCambiar(bloque.copy(variante = it), null) }
            }
            tipo.campos.forEach { c ->
                CampoDeBloque(c, bloque[c.id]) { v -> alCambiar(bloque.con(c.id, v), "${bloque.id}.${c.id}") }
            }
            Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                BotonTonal(if (bloque.visible) "Ocultar" else "Mostrar", { alCambiar(bloque.copy(oculto = if (bloque.visible) true else null), null) }, Modifier.weight(1f), alto = 40.dp,
                    lleno = true, simbolo = if (bloque.visible) "eye.slash" else "eye")
                BotonTonal("Duplicar", alDuplicar, Modifier.weight(1f), alto = 40.dp, lleno = true, simbolo = "plus.square.on.square")
            }
            BotonTexto("Borrar el bloque", alBorrar, color = C.error, simbolo = "trash")
        }
    }
}

/** Las variantes de un bloque (o de una página) con su miniatura, en una tira que se corre. */
@Composable
fun SelectorVariante(tipo: String, variantes: List<String>, nombres: List<String>, elegida: String, alElegir: (String) -> Unit) {
    Row(Modifier.fillMaxWidth().horizontalScroll(rememberScrollState()), horizontalArrangement = Arrangement.spacedBy(10.dp)) {
        variantes.forEachIndexed { i, v ->
            val on = v == elegida
            Column(Modifier.width(112.dp).clip(RoundedCornerShape(10.dp)).clickable { alElegir(v); Haptico.seleccion() }
                .semantics { contentDescription = nombres.getOrNull(i) ?: v }, verticalArrangement = Arrangement.spacedBy(4.dp)) {
                MiniaturaBloque(tipo, v, Modifier.fillMaxWidth(), alto = 72.dp, elegida = on)
                Texto(nombres.getOrNull(i) ?: v, if (on) Estilo.footnoteFuerte else Estilo.footnote, if (on) C.texto else C.texto2, lineas = 2)
            }
        }
    }
}

// MARK: - Ajustes de página

/** Catálogo, Ficha y Carrito: la variante y los ajustes de bloques.json › paginas. */
@Composable
fun AjustesPagina(pagina: String, paginas: PaginasTienda, alCambiar: (PaginasTienda) -> Unit) {
    val p = CatalogoBloques.compartido.paginas[pagina] ?: return
    TarjetaSeccion("Cómo se ve ${p.nombre.lowercase().let { if (pagina == "producto") "la ficha" else "el ${it}" }}", "rectangle.3.group") {
        if (p.variantes.isNotEmpty()) {
            val actual = paginas.valor(pagina, "variante")?.texto ?: p.variantes.first()
            SelectorVariante(pagina, p.variantes, p.nombresVariantes, actual) { alCambiar(paginas.con(pagina, "variante", Valor.texto(it))) }
        }
        p.campos.forEach { c ->
            CampoDeBloque(c, paginas.valor(pagina, c.id)) { v -> alCambiar(paginas.con(pagina, c.id, v ?: c.porDefecto ?: Valor.nulo)) }
        }
    }
}

// MARK: - Campos (todos los tiposCampo de bloques.json)

/** Un campo de bloques.json con su control. `valor` nulo: lo que el campo trae por defecto. */
@Composable
fun CampoDeBloque(c: CampoBloque, valor: ValorJSON?, alCambiar: (ValorJSON?) -> Unit) {
    val ctx = LocalContextoCampos.current
    val v = valor ?: c.porDefecto
    when (c.tipo) {
        "texto" -> {
            val ayuda = when (c.sugerido) { "nombre" -> "vacío: ${ctx.nombreTienda.ifEmpty { "el nombre de tu tienda" }}"; "lema" -> "vacío: tu lema"; else -> null }
            CampoTexto(c.nombre + (ayuda?.let { " ($it)" } ?: ""), v?.texto ?: "", { alCambiar(Valor.texto(it.take(c.max ?: 200))) })
        }
        "textoLargo" -> Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
            Texto(c.nombre, Estilo.footnoteFuerte, C.texto2)
            EditorTexto(v?.texto ?: "", { alCambiar(Valor.texto(it.take(c.max ?: 600))) }, placeholder = c.nombre)
        }
        "foto" -> CampoFoto(c.nombre, v?.texto ?: "") { alCambiar(it?.let { u -> Valor.texto(u) }) }
        "fotos" -> CampoFotos(c.nombre, v?.lista?.mapNotNull { it.texto } ?: emptyList(), c.max ?: 6) { alCambiar(Valor.lista(it.map { u -> Valor.texto(u) })) }
        "link" -> CampoLink(c.nombre, v?.texto ?: "") { alCambiar(Valor.texto(it)) }
        "boton" -> Column(Modifier.fillMaxWidth().clip(RoundedCornerShape(14.dp)).background(C.crema2).padding(12.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
            val t = v?.get("t")?.texto ?: ""; val a = v?.get("a")?.texto ?: ""
            Texto(c.nombre, Estilo.subheadFuerte, C.texto)
            CampoTexto("Texto del botón", t, { alCambiar(Valor.objeto(mapOf("t" to Valor.texto(it.take(30)), "a" to Valor.texto(a.ifEmpty { "catalogo" })))) })
            CampoLink("Adónde lleva", a) { alCambiar(Valor.objeto(mapOf("t" to Valor.texto(t), "a" to Valor.texto(it)))) }
            if (t.isEmpty()) Texto("Sin texto, el botón no se muestra.", Estilo.footnote, C.texto2)
        }
        "fondo" -> CampoOpcion(c.nombre, CatalogoBloques.compartido.fondos.pares, v?.texto ?: CatalogoBloques.compartido.fondos.opciones.firstOrNull() ?: "") { alCambiar(Valor.texto(it)) }
        "opcion" -> CampoOpcion(c.nombre, (c.opciones ?: emptyList()).mapIndexed { i, o -> o to (c.nombres?.getOrNull(i) ?: o) }, v?.texto ?: c.opciones?.firstOrNull() ?: "") { alCambiar(Valor.texto(it)) }
        "fecha" -> CampoFecha(c.nombre, v?.texto) { alCambiar(Valor.texto(it)) }
        "numero" -> Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
            val n = v?.entero ?: c.min ?: 0
            Texto("${c.nombre}: $n", Estilo.body16, C.texto, Modifier.weight(1f))
            Pasador(n, { alCambiar(Valor.numero(it)) }, (c.min ?: 0)..(c.max ?: 99))
        }
        "si" -> Interruptor(c.nombre, v?.bool ?: false, { alCambiar(Valor.booleano(it)) })
        "icono" -> Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
            Texto(c.nombre, Estilo.footnoteFuerte, C.texto2)
            Row(Modifier.fillMaxWidth().horizontalScroll(rememberScrollState()), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                CatalogoBloques.compartido.iconos.forEach { i ->
                    val on = v?.texto == i
                    val (simbolo, nombre) = iconosBloque[i] ?: ("star" to i)
                    Column(Modifier.width(76.dp).clip(RoundedCornerShape(12.dp)).background(if (on) C.crema else C.crema2).border(if (on) 2.dp else 0.dp, if (on) C.acento else C.crema2, RoundedCornerShape(12.dp))
                        .clickable { alCambiar(Valor.texto(i)) }.padding(vertical = 10.dp), horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(4.dp)) {
                        Icono(simbolo, 18f, C.primarioTexto)
                        Texto(nombre, Estilo.footnote, C.texto, lineas = 1, escalaMinima = 0.8f)
                    }
                }
            }
        }
        "video" -> Column(verticalArrangement = Arrangement.spacedBy(4.dp)) {
            CampoTexto(c.nombre, v?.texto ?: "", { alCambiar(Valor.texto(it.trim())) }, teclado = KeyboardType.Uri)
            Texto("Pegá el link de YouTube o Vimeo.", Estilo.footnote, C.texto2)
        }
        "producto" -> CampoProductos(c.nombre, listOfNotNull(v?.texto), 1) { alCambiar(it.firstOrNull()?.let { id -> Valor.texto(id) }) }
        "elegidos" -> CampoProductos(c.nombre, v?.lista?.mapNotNull { it.texto } ?: emptyList(), c.max ?: 12) { alCambiar(Valor.lista(it.map { id -> Valor.texto(id) })) }
        "productos" -> CampoFuenteProductos(c.nombre, v?.objeto ?: emptyMap()) { alCambiar(Valor.objeto(it)) }
        "categoria" -> Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
            Texto(c.nombre, Estilo.body16, C.texto, Modifier.weight(1f))
            if (ctx.categorias.isEmpty()) Texto("Sin categorías", Estilo.subhead, C.texto3)
            else SelectorMenu(listOf("" to "Elegir…") + ctx.categorias.map { it to it }, v?.texto ?: "", { alCambiar(Valor.texto(it)) }, c.nombre)
        }
        "categorias" -> Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
            val elegidas = v?.lista?.mapNotNull { it.texto } ?: emptyList()
            Texto(c.nombre + if (elegidas.isEmpty()) " (vacío: todas)" else "", Estilo.footnoteFuerte, C.texto2)
            if (ctx.categorias.isEmpty()) Texto("Cuando tus productos tengan categoría, las elegís acá.", Estilo.footnote, C.texto2)
            Row(Modifier.fillMaxWidth().horizontalScroll(rememberScrollState()), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                ctx.categorias.forEach { cat ->
                    ChipFiltro(cat, activo = cat in elegidas) { alCambiar(Valor.lista((if (cat in elegidas) elegidas - cat else elegidas + cat).map { Valor.texto(it) })) }
                }
            }
        }
        "lista" -> CampoLista(c, v?.lista ?: emptyList(), alCambiar)
        "grupo" -> Column(Modifier.fillMaxWidth().clip(RoundedCornerShape(14.dp)).background(C.crema2).padding(12.dp), verticalArrangement = Arrangement.spacedBy(10.dp)) {
            val o = v?.objeto ?: emptyMap()
            Texto(c.nombre, Estilo.subheadFuerte, C.texto)
            c.item?.forEach { sub ->
                CampoDeBloque(sub, o[sub.id]) { nv ->
                    val m = o.toMutableMap(); if (nv == null || nv.vacio) m.remove(sub.id) else m[sub.id] = nv
                    alCambiar(if (m.isEmpty()) null else Valor.objeto(m))
                }
            }
        }
        else -> Texto("${c.nombre}: este campo se edita desde la compu.", Estilo.footnote, C.texto2)
    }
}

/** Una opción entre pocas: chips que se corren. */
@Composable
private fun CampoOpcion(nombre: String, opciones: List<Pair<String, String>>, elegida: String, alElegir: (String) -> Unit) {
    Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
        Texto(nombre, Estilo.footnoteFuerte, C.texto2)
        Row(Modifier.fillMaxWidth().horizontalScroll(rememberScrollState()), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            opciones.forEach { (o, t) -> ChipFiltro(t, activo = o == elegida) { alElegir(o) } }
        }
    }
}

/** Una foto: tocar para elegirla de la galería (se sube a /panel/fotos) y quitarla. */
@Composable
fun CampoFoto(nombre: String, url: String, alCambiar: (String?) -> Unit) {
    val estado = LocalEstado.current
    val ctx = LocalContextoCampos.current
    val alcance = rememberCoroutineScope()
    var subiendo by remember { mutableStateOf(false) }
    val elegir = rememberLauncherForActivityResult(ActivityResultContracts.PickVisualMedia()) { u ->
        u ?: return@rememberLauncherForActivityResult
        alcance.launch { subiendo = true; ctx.subir(u)?.let { alCambiar(it) }; subiendo = false }
    }
    Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
        Box(Modifier.size(72.dp).clip(RoundedCornerShape(12.dp)).background(C.crema2).clickable {
            elegir.launch(PickVisualMediaRequest(ActivityResultContracts.PickVisualMedia.ImageOnly)); estado.abrioOtraApp = true
        }.semantics { contentDescription = "Elegir $nombre" }, contentAlignment = Alignment.Center) {
            when {
                subiendo -> Rueda(C.acento)
                url.isNotEmpty() -> AsyncImage(url, null, Modifier.fillMaxSize(), contentScale = ContentScale.Crop)
                else -> Icono("photo.badge.plus", 22f, C.primarioTexto)
            }
        }
        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
            Texto(nombre, Estilo.subheadFuerte, C.texto)
            Texto(if (url.isEmpty()) "Tocá para elegir una foto" else "Tocá la foto para cambiarla", Estilo.footnote, C.texto2)
        }
        if (url.isNotEmpty()) Box(Modifier.size(36.dp).clickable { alCambiar(null) }.semantics { contentDescription = "Quitar $nombre" }, contentAlignment = Alignment.Center) {
            Icono("trash", 15f, C.primarioTexto)
        }
    }
}

/** Varias fotos (hasta `max`): agregar, quitar y correr a la izquierda. */
@Composable
fun CampoFotos(nombre: String, fotos: List<String>, max: Int, alCambiar: (List<String>) -> Unit) {
    val estado = LocalEstado.current
    val ctx = LocalContextoCampos.current
    val alcance = rememberCoroutineScope()
    var subiendo by remember { mutableStateOf(false) }
    val actuales by rememberUpdatedState(fotos)
    val elegir = rememberLauncherForActivityResult(ActivityResultContracts.PickVisualMedia()) { u ->
        u ?: return@rememberLauncherForActivityResult
        alcance.launch { subiendo = true; ctx.subir(u)?.let { alCambiar((actuales + it).take(max)) }; subiendo = false }
    }
    Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
        Texto("$nombre · ${fotos.size} de $max", Estilo.footnoteFuerte, C.texto2)
        Row(Modifier.fillMaxWidth().horizontalScroll(rememberScrollState()), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            fotos.forEachIndexed { i, f ->
                Box(Modifier.size(80.dp).clip(RoundedCornerShape(12.dp))) {
                    AsyncImage(f, null, Modifier.fillMaxSize().clickable(enabled = i > 0) {
                        val l = fotos.toMutableList(); val x = l.removeAt(i); l.add(i - 1, x); alCambiar(l)
                    }, contentScale = ContentScale.Crop)
                    Box(Modifier.align(Alignment.TopEnd).padding(4.dp).size(26.dp).clip(RoundedCornerShape(13.dp)).background(C.velo).clickable { alCambiar(fotos.filterIndexed { j, _ -> j != i }) }
                        .semantics { contentDescription = "Quitar la foto ${i + 1}" }, contentAlignment = Alignment.Center) { Icono("xmark", 11f, androidx.compose.ui.graphics.Color.White) }
                }
            }
            if (fotos.size < max) Box(Modifier.size(80.dp).clip(RoundedCornerShape(12.dp)).background(C.crema2).clickable {
                elegir.launch(PickVisualMediaRequest(ActivityResultContracts.PickVisualMedia.ImageOnly)); estado.abrioOtraApp = true
            }.semantics { contentDescription = "Agregar una foto" }, contentAlignment = Alignment.Center) {
                if (subiendo) Rueda(C.acento) else Icono("photo.badge.plus", 22f, C.primarioTexto)
            }
        }
        if (fotos.size > 1) Texto("Tocá una foto para pasarla adelante.", Estilo.footnote, C.texto2)
    }
}

/** Un link de la tienda: una página, una categoría, un producto o uno propio (https://). */
@Composable
fun CampoLink(nombre: String, valor: String, alCambiar: (String) -> Unit) {
    val ctx = LocalContextoCampos.current
    val propio = valor.startsWith("https://") || valor.startsWith("http://")
    val opciones = listOf("" to "Sin link") + paginasLink + ctx.categorias.map { "cat:$it" to "Categoría: $it" } +
        ctx.productos.take(30).map { "prod:${it.id}" to "Producto: ${it.nombre}" } + listOf("https://" to "Un link propio…")
    val elegido = if (propio) "https://" else valor
    Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
        Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
            Texto(nombre, Estilo.body16, C.texto, Modifier.weight(1f))
            SelectorMenu(if (opciones.any { it.first == elegido }) opciones else opciones + (elegido to elegido), elegido, { alCambiar(it) }, nombre)
        }
        if (propio) CampoTexto("Dirección (https://…)", valor, { alCambiar(it.trim()) }, teclado = KeyboardType.Uri)
    }
}

/** Fecha y hora (ISO 8601, hora de Argentina). */
@Composable
private fun CampoFecha(nombre: String, valor: String?, alCambiar: (String) -> Unit) {
    val ar = ZoneOffset.ofHours(-3)
    val f = valor?.let { runCatching { OffsetDateTime.parse(it).withOffsetSameInstant(ar) }.getOrNull() }
        ?: OffsetDateTime.now(ar).plusDays(3).withHour(23).withMinute(59).withSecond(0).withNano(0)
    fun poner(d: LocalDate, minutos: Int) = alCambiar(d.atTime(minutos / 60, minutos % 60).atOffset(ar).format(DateTimeFormatter.ISO_OFFSET_DATE_TIME))
    Column(Modifier.fillMaxWidth().clip(RoundedCornerShape(14.dp)).background(C.crema2)) {
        RenglonFecha(nombre, f.toLocalDate(), { poner(it, f.hour * 60 + f.minute) }, minimo = LocalDate.now())
        RenglonHora("Hora", f.hour * 60 + f.minute) { poner(f.toLocalDate(), it) }
        if (valor == null) Texto("Elegí cuándo termina: hasta entonces se ve la cuenta regresiva.", Estilo.footnote, C.texto2, Modifier.padding(horizontal = 16.dp).padding(bottom = 10.dp))
    }
}

/** De dónde salen los productos: la fuente, cuántos y, si son elegidos a mano o de una categoría, cuáles. */
@Composable
private fun CampoFuenteProductos(nombre: String, o: Map<String, ValorJSON>, alCambiar: (Map<String, ValorJSON>) -> Unit) {
    val ctx = LocalContextoCampos.current
    val fuentes = CatalogoBloques.compartido.fuentesProductos
    val fuente = o["fuente"]?.texto ?: "vendidos"
    val cantidad = o["cantidad"]?.entero ?: 8
    fun con(k: String, v: ValorJSON?) = alCambiar(o.toMutableMap().apply { if (v == null) remove(k) else put(k, v) })
    Column(Modifier.fillMaxWidth().clip(RoundedCornerShape(14.dp)).background(C.crema2).padding(12.dp), verticalArrangement = Arrangement.spacedBy(10.dp)) {
        Texto(nombre, Estilo.subheadFuerte, C.texto)
        Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
            Texto("Cuáles", Estilo.body16, C.texto, Modifier.weight(1f))
            SelectorMenu(fuentes.pares, fuente, { f ->
                val m = o.toMutableMap(); m["fuente"] = Valor.texto(f); if (f != "elegidos") m.remove("ids"); if (f != "categoria") m.remove("categoria"); alCambiar(m)
            }, "Cuáles productos")
        }
        if (fuente != "elegidos") Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
            Texto("Cuántos: $cantidad", Estilo.body16, C.texto, Modifier.weight(1f))
            Pasador(cantidad, { con("cantidad", Valor.numero(it)) }, 1..24)
        }
        if (fuente == "elegidos") CampoProductos("Los productos", o["ids"]?.lista?.mapNotNull { it.texto } ?: emptyList(), 24) { ids ->
            val m = o.toMutableMap(); m["ids"] = Valor.lista(ids.map { Valor.texto(it) }); m["cantidad"] = Valor.numero(ids.size.coerceAtLeast(1)); alCambiar(m)
        }
        if (fuente == "categoria") Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
            Texto("Categoría", Estilo.body16, C.texto, Modifier.weight(1f))
            if (ctx.categorias.isEmpty()) Texto("Sin categorías", Estilo.subhead, C.texto3)
            else SelectorMenu(listOf("" to "Elegir…") + ctx.categorias.map { it to it }, o["categoria"]?.texto ?: "", { con("categoria", Valor.texto(it)) }, "Categoría")
        }
        val ayuda = when (fuente) {
            "vendidos" -> "Se arma solo con lo que más se vende."
            "nuevos" -> "Los últimos que cargaste."
            "ofertas" -> "Los que tienen precio de oferta."
            "categoria" -> if (o["categoria"]?.texto.isNullOrEmpty()) "Elegí la categoría." else "Los de esa categoría."
            else -> if (o["ids"]?.lista.isNullOrEmpty()) "Elegí los productos a mano." else "En el orden en que los elegiste."
        }
        Texto(ayuda, Estilo.footnote, C.texto2)
    }
}

/** Elegir productos de la tienda (uno o varios, hasta `max`), con buscador. */
@Composable
fun CampoProductos(nombre: String, ids: List<String>, max: Int, alCambiar: (List<String>) -> Unit) {
    val ctx = LocalContextoCampos.current
    var abierto by remember { mutableStateOf(false) }
    val elegidos = ids.mapNotNull { id -> ctx.productos.firstOrNull { it.id == id } }
    Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
        Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
            Texto(nombre, Estilo.body16, C.texto, Modifier.weight(1f))
            Texto(if (ids.isEmpty()) "Elegir" else "Cambiar", Estilo.subheadFuerte, C.primarioTexto, Modifier.clickable { abierto = true }.padding(8.dp))
        }
        if (ctx.productos.isEmpty()) Texto("Todavía no cargaste productos.", Estilo.footnote, C.texto2)
        elegidos.forEach { p ->
            Row(horizontalArrangement = Arrangement.spacedBy(10.dp), verticalAlignment = Alignment.CenterVertically) {
                Miniatura(p.foto, 36.dp, radio = 8.dp)
                Texto(p.nombre, Estilo.subhead, C.texto, Modifier.weight(1f), lineas = 1)
                Texto(Formato.pesos(p.precio), Estilo.footnote, C.texto2, numeros = true)
            }
        }
    }
    Hoja(abierto, { abierto = false }) {
        var q by remember { mutableStateOf("") }
        var marcados by remember(ids) { mutableStateOf(ids) }
        Column(Modifier.fillMaxSize()) {
            Grabber()
            Row(Modifier.fillMaxWidth().padding(horizontal = 20.dp), verticalAlignment = Alignment.CenterVertically) {
                Texto(nombre, Estilo.tituloHojaChico, C.texto, Modifier.weight(1f))
                Texto("Listo", Estilo(17f, Peso.bold), C.primarioTexto, Modifier.clickable { alCambiar(marcados); abierto = false }.padding(8.dp))
            }
            Buscador("Buscar productos", q, { q = it }, Modifier.padding(horizontal = 20.dp, vertical = 8.dp))
            if (max > 1) Texto("${marcados.size} de $max elegidos", Estilo.footnote, C.texto2, Modifier.padding(horizontal = 20.dp))
            Column(Modifier.weight(1f).verticalScroll(rememberScrollState()).padding(horizontal = 20.dp).padding(bottom = 24.dp)) {
                ctx.productos.filter { q.isBlank() || it.nombre.contains(q.trim(), ignoreCase = true) }.forEach { p ->
                    val on = p.id in marcados
                    Row(Modifier.fillMaxWidth().heightIn(min = 56.dp).clickable {
                        marcados = when {
                            on -> marcados - p.id
                            max == 1 -> listOf(p.id)
                            marcados.size < max -> marcados + p.id
                            else -> marcados
                        }
                        if (max == 1 && !on) { alCambiar(marcados); abierto = false }
                    }, horizontalArrangement = Arrangement.spacedBy(10.dp), verticalAlignment = Alignment.CenterVertically) {
                        Miniatura(p.foto, 40.dp, radio = 10.dp)
                        Texto(p.nombre, Estilo.body16, C.texto, Modifier.weight(1f), lineas = 2)
                        Icono(if (on) "checkmark.circle.fill" else "circle", 20f, if (on) C.acento else C.texto3)
                    }
                    Separador(0.dp)
                }
            }
        }
    }
}

/** Una lista de cosas con los mismos campos (beneficios, preguntas, horarios…): agregar, ordenar y quitar. */
@Composable
private fun CampoLista(c: CampoBloque, items: List<ValorJSON>, alCambiar: (ValorJSON?) -> Unit) {
    val max = c.max ?: CatalogoBloques.compartido.limites.itemsPorLista
    val nombreItem = c.item?.firstOrNull()?.nombre?.lowercase() ?: "uno"
    fun poner(l: List<ValorJSON>) = alCambiar(if (l.isEmpty()) null else Valor.lista(l))
    Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
        Texto("${c.nombre} · ${items.size} de $max", Estilo.footnoteFuerte, C.texto2)
        if (items.isEmpty()) Texto(if ((c.min ?: 0) > 0) "Sumá el primero para que se vea." else "Vacío: lo arma la tienda sola.", Estilo.footnote, C.texto2)
        items.forEachIndexed { i, it ->
            val o = it.objeto ?: emptyMap()
            Column(Modifier.fillMaxWidth().clip(RoundedCornerShape(14.dp)).background(C.crema2).padding(12.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Texto("${i + 1}", Estilo.subheadFuerte, C.texto2, Modifier.weight(1f))
                    Box(Modifier.size(36.dp).alpha(if (i == 0) 0.3f else 1f).clickable(enabled = i > 0) {
                        val l = items.toMutableList(); val x = l.removeAt(i); l.add(i - 1, x); poner(l)
                    }.semantics { contentDescription = "Subir" }, contentAlignment = Alignment.Center) { Icono("chevron.up", 13f, C.primarioTexto) }
                    Box(Modifier.size(36.dp).alpha(if (i == items.size - 1) 0.3f else 1f).clickable(enabled = i < items.size - 1) {
                        val l = items.toMutableList(); val x = l.removeAt(i); l.add(i + 1, x); poner(l)
                    }.semantics { contentDescription = "Bajar" }, contentAlignment = Alignment.Center) { Icono("chevron.down", 13f, C.primarioTexto) }
                    Box(Modifier.size(36.dp).clickable { poner(items.filterIndexed { j, _ -> j != i }) }.semantics { contentDescription = "Quitar" }, contentAlignment = Alignment.Center) {
                        Icono("trash", 14f, C.primarioTexto)
                    }
                }
                c.item?.forEach { sub ->
                    CampoDeBloque(sub, o[sub.id]) { nv ->
                        val m = o.toMutableMap(); if (nv == null || nv.vacio) m.remove(sub.id) else m[sub.id] = nv
                        val l = items.toMutableList(); l[i] = JsonObject(m); poner(l)
                    }
                }
            }
        }
        if (items.size < max) BotonTonal("Agregar $nombreItem", { poner(items + JsonObject(emptyMap())) }, alto = 40.dp, simbolo = "plus")
    }
}
