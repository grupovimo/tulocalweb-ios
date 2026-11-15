package ar.com.tulocalweb.app.pantallas.mas

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.gestures.detectDragGesturesAfterLongPress
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxWithConstraints
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
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.key
import androidx.compose.runtime.mutableFloatStateOf
import androidx.compose.runtime.mutableStateMapOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.layout.onSizeChanged
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.unit.dp
import androidx.compose.ui.zIndex
import ar.com.tulocalweb.app.app.Destino
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.app.LocalNavegacion
import ar.com.tulocalweb.app.nucleo.Bloque
import ar.com.tulocalweb.app.nucleo.CatalogoBloques
import ar.com.tulocalweb.app.nucleo.FaltaBloque
import ar.com.tulocalweb.app.nucleo.texto
import ar.com.tulocalweb.app.tema.*
import kotlinx.coroutines.delay

// Clon de Pantallas/Mas/EditorBloques.swift.
// Mi tienda › Páginas: elegir la página, su lista de bloques (arrastrar, agregar entre medio, duplicar, ocultar,
// borrar, deshacer y rehacer), la biblioteca y el panel de cada bloque.

/** La pestaña Páginas del editor. */
@Composable
fun ListaBloques(editor: EditorTienda) {
    val cat = CatalogoBloques.shared
    @Suppress("UNUSED_VARIABLE") val pasos = editor.pasos // el historial no es observable: esto redibuja deshacer y rehacer
    Column(verticalArrangement = Arrangement.spacedBy(14.dp)) {
        Row(Modifier.fillMaxWidth().horizontalScroll(rememberScrollState()), horizontalArrangement = Arrangement.spacedBy(6.dp)) {
            CatalogoBloques.ordenPaginas.forEach { p ->
                ChipFiltro(cat.paginas[p]?.nombre?.replace("Ficha de producto", "Ficha") ?: p, cuenta = editor.diseno.bloques(p).size, activo = editor.pagina == p) {
                    editor.pagina = p
                    editor.seleccion = null
                }
            }
        }
        Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            Texto(ayuda(editor.pagina), Estilo.footnote, C.texto2, Modifier.weight(1f))
            BotonIcono("arrow.uturn.backward", "Deshacer", editor.historial.puedeDeshacer) { editor.deshacer() }
            BotonIcono("arrow.uturn.forward", "Rehacer", editor.historial.puedeRehacer) { editor.rehacer() }
        }
        if (cat.paginas[editor.pagina]?.variantes != null || cat.paginas[editor.pagina]?.campos != null) AjustesPaginaVista(editor, editor.pagina)
        if (editor.paginaConBloques) Lista(editor)
    }
}

private fun ayuda(pagina: String): String = when (pagina) {
    "inicio" -> "Tocá un bloque para editarlo. Mantenelo apretado para moverlo."
    "producto" -> "Cómo se ve cada producto. Abajo de la ficha podés sumar bloques."
    "nosotros" -> "Contá quién está detrás de tu tienda."
    else -> "Cómo se ve esta página en tu tienda."
}

@Composable
private fun Lista(editor: EditorTienda) {
    val cat = CatalogoBloques.shared
    val datos = LocalDatosEditor.current
    val bloques = editor.bloques
    // Arrastrar: el bloque sigue al dedo y cambia de lugar al pasar la mitad del vecino.
    var arrastrado by remember { mutableStateOf<String?>(null) }
    var corrimiento by remember { mutableFloatStateOf(0f) }
    val altos = remember { mutableStateMapOf<String, Int>() }
    Column {
        if (editor.pagina == "producto") Texto("Debajo de la ficha", Estilo.subheadFuerte, C.texto, Modifier.padding(bottom = 8.dp))
        if (bloques.isEmpty()) {
            Vacio(editor)
        } else {
            bloques.forEachIndexed { i, b ->
                key(b.id) {
                    Insertar(editor, i)
                    val esEste = arrastrado == b.id
                    Box(Modifier.zIndex(if (esEste) 1f else 0f).onSizeChanged { altos[b.id] = it.height + 30 }
                        .graphicsLayer { translationY = if (esEste) corrimiento else 0f; shadowElevation = if (esEste) 16f else 0f; shape = RoundedCornerShape(16.dp); clip = esEste }
                        .pointerInput(b.id) {
                            detectDragGesturesAfterLongPress(
                                onDragStart = { arrastrado = b.id; corrimiento = 0f; Haptico.impacto() },
                                onDragEnd = { arrastrado = null; corrimiento = 0f },
                                onDragCancel = { arrastrado = null; corrimiento = 0f },
                            ) { cambio, d ->
                                cambio.consume()
                                corrimiento += d.y
                                val l = editor.bloques
                                val j = l.indexOfFirst { it.id == b.id }
                                if (j < 0) return@detectDragGesturesAfterLongPress
                                if (corrimiento > 0 && j < l.size - 1) {
                                    val h = altos[l[j + 1].id] ?: return@detectDragGesturesAfterLongPress
                                    if (corrimiento > h / 2f) { editor.mover(b.id, l.getOrNull(j + 2)?.id); corrimiento -= h; Haptico.seleccion() }
                                } else if (corrimiento < 0 && j > 0) {
                                    val h = altos[l[j - 1].id] ?: return@detectDragGesturesAfterLongPress
                                    if (-corrimiento > h / 2f) { editor.mover(b.id, l[j - 1].id); corrimiento += h; Haptico.seleccion() }
                                }
                            }
                        }) {
                        FilaBloque(editor, b, FaltaBloque.de(b, datos.contexto), primero = i == 0, ultimo = i == bloques.size - 1)
                    }
                }
            }
            Insertar(editor, bloques.size)
            BotonSecundario("Sumar un bloque", { editor.insertarEn = bloques.size }, Modifier.padding(top = 4.dp), simbolo = "plus",
                habilitado = bloques.size < cat.limites.bloquesPorPagina)
            if (bloques.size >= cat.limites.bloquesPorPagina) {
                Texto("Llegaste al tope de ${cat.limites.bloquesPorPagina} bloques en esta página.", Estilo.footnote, C.texto2)
            }
        }
    }
}

@Composable
private fun Vacio(editor: EditorTienda) {
    val cat = CatalogoBloques.shared
    val sugerido = if (editor.pagina == "nosotros") "textoImagen" else if (editor.pagina == "producto") "preguntas" else "portada"
    Column(Modifier.fillMaxWidth().tarjeta(radio = 20.dp).padding(18.dp), horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(12.dp)) {
        MiniaturaBloque(sugerido, cat.tipo(sugerido)?.variantes?.firstOrNull() ?: "", 140.dp)
        Texto(if (editor.pagina == "nosotros") "Nosotros está vacía" else "Esta página no tiene bloques", Estilo.headline, C.texto)
        Texto(if (editor.pagina == "nosotros") "Empezá con un texto con foto: tu historia, el taller o el equipo."
            else "Sumá el primero: preguntas frecuentes o beneficios ayudan a decidir.", Estilo.subhead, C.texto2, alinear = androidx.compose.ui.text.style.TextAlign.Center)
        cat.tipo(sugerido)?.let { t -> BotonPrincipal("Sumar ${t.nombre.lowercase()}", { editor.agregar(t, 0) }) }
        BotonTexto("Ver todos los bloques", { editor.insertarEn = 0 }, alto = 40.dp)
    }
}

/** El «+» entre dos bloques, para sumar uno justo ahí. */
@Composable
private fun Insertar(editor: EditorTienda, i: Int) {
    Row(Modifier.fillMaxWidth().height(30.dp).clickable { editor.insertarEn = i }
        .semantics { contentDescription = if (i == 0) "Sumar un bloque al principio" else "Sumar un bloque acá" },
        horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.CenterVertically) {
        Box(Modifier.weight(1f).height(1.dp).background(C.linea))
        Box(Modifier.size(24.dp).clip(CircleShape).background(C.crema), contentAlignment = Alignment.Center) { Icono("plus", 11f, C.primarioTexto) }
        Box(Modifier.weight(1f).height(1.dp).background(C.linea))
    }
}

/** Un botón de ícono de 44 × 44 (deshacer, rehacer, menú del bloque). */
@Composable
fun BotonIcono(simbolo: String, descripcion: String, habilitado: Boolean = true, accion: () -> Unit) {
    Box(Modifier.size(44.dp).alpha(if (habilitado) 1f else 0.35f).clickable(enabled = habilitado, onClick = accion).semantics { contentDescription = descripcion },
        contentAlignment = Alignment.Center) { Icono(simbolo, 16f, C.primarioTexto) }
}

/** Una fila de la lista: miniatura, nombre, variante, qué le falta y el menú. */
@Composable
fun FilaBloque(editor: EditorTienda, bloque: Bloque, falta: FaltaBloque?, primero: Boolean, ultimo: Boolean) {
    val estado = LocalEstado.current
    val tipo = CatalogoBloques.shared.tipo(bloque.tipo)
    val elegido = editor.seleccion == bloque.id
    val forma = RoundedCornerShape(16.dp)
    val v = tipo?.nombreVariante(bloque.variante) ?: bloque.variante
    val subtitulo = if (bloque.oculto) "Oculto · $v" else v + (bloque.datos["titulo"]?.texto?.let { " · $it" } ?: "")
    fun borrar() {
        val nombre = tipo?.nombre ?: "el bloque"
        editor.borrar(bloque.id)
        Haptico.impacto()
        estado.avisar("Borramos «$nombre»", deshacer = { editor.deshacer() })
    }
    Row(Modifier.fillMaxWidth().alpha(if (bloque.oculto) 0.6f else 1f).clip(forma).background(C.superficie)
        .border(if (elegido) 2.dp else 1.dp, if (elegido) C.acento else C.linea, forma).padding(start = 12.dp, top = 10.dp, bottom = 10.dp, end = 4.dp),
        horizontalArrangement = Arrangement.spacedBy(10.dp), verticalAlignment = Alignment.CenterVertically) {
        Row(Modifier.weight(1f).clickable { editor.abrir(bloque.id) }, horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
            Icono("line.3.horizontal", 13f, C.texto3)
            MiniaturaBloque(bloque.tipo, bloque.variante, 66.dp)
            Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(3.dp)) {
                Texto(tipo?.nombre ?: bloque.tipo, Estilo.fila, C.texto, lineas = 1)
                Texto(subtitulo, Estilo.footnote, C.texto2, lineas = 1)
                if (falta != null && !bloque.oculto) {
                    Row(horizontalArrangement = Arrangement.spacedBy(4.dp), verticalAlignment = Alignment.Top) {
                        Icono("exclamationmark.circle.fill", 12f, C.avisoTinta, Modifier.padding(top = 1.dp))
                        Texto(falta.texto, Estilo.footnoteFuerte, C.avisoTinta, lineas = 2)
                    }
                }
            }
        }
        MenuBoton(listOfNotNull(
            OpcionDeMenu("Editar", "pencil") { editor.abrir(bloque.id) },
            OpcionDeMenu("Duplicar", "plus.square.on.square") { editor.duplicar(bloque.id) },
            OpcionDeMenu(if (bloque.oculto) "Mostrar" else "Ocultar", if (bloque.oculto) "eye" else "eye.slash") { editor.alternarOculto(bloque.id) },
            if (!primero) OpcionDeMenu("Subir", "arrow.up") { editor.mover(bloque.id, -1) } else null,
            if (!ultimo) OpcionDeMenu("Bajar", "arrow.down") { editor.mover(bloque.id, 1) } else null,
            OpcionDeMenu("Borrar", "trash", destructiva = true) { borrar() },
        ), "Más acciones de ${tipo?.nombre ?: "este bloque"}") {
            Box(Modifier.size(44.dp), contentAlignment = Alignment.Center) { Icono("ellipsis", 17f, C.texto2) }
        }
    }
}

// MARK: - Biblioteca

/** Los bloques para sumar, por grupo y con su miniatura. */
@Composable
fun BibliotecaBloques(editor: EditorTienda, cerrar: () -> Unit) {
    val cat = CatalogoBloques.shared
    val estado = LocalEstado.current
    Column(Modifier.fillMaxSize().background(C.fondo)) {
        Grabber()
        Box(Modifier.fillMaxWidth().height(44.dp).padding(horizontal = 16.dp)) {
            Texto("Cancelar", Estilo(17f, Peso.regular), C.primarioTexto, Modifier.align(Alignment.CenterStart).clickable(onClick = cerrar))
            Texto("Sumar un bloque", Estilo.headline, C.texto, Modifier.align(Alignment.Center))
        }
        Column(Modifier.weight(1f).verticalScroll(rememberScrollState()).margenPantalla().padding(vertical = 12.dp), verticalArrangement = Arrangement.spacedBy(22.dp)) {
            cat.grupos.forEach { g ->
                val tipos = cat.bloques(g.id)
                if (tipos.isNotEmpty()) Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
                    Texto(g.nombre, Estilo.headline, C.texto)
                    tipos.chunked(2).forEach { par ->
                        Row(Modifier.fillMaxWidth().height(androidx.compose.foundation.layout.IntrinsicSize.Max), horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                            par.forEach { t ->
                                Tarjeta(editor, t, Modifier.weight(1f).fillMaxSize()) {
                                    val en = editor.insertarEn ?: editor.bloques.size
                                    cerrar()
                                    Haptico.seleccion()
                                    editor.insertarEn = null
                                    val id = editor.agregar(t, en, abrir = false)
                                    estado.tarea { delay(500); editor.panel = id }
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

@Composable
private fun Tarjeta(editor: EditorTienda, t: CatalogoBloques.Tipo, modifier: Modifier, elegir: () -> Unit) {
    val puede = editor.puedeAgregar(t)
    val forma = RoundedCornerShape(16.dp)
    Column(modifier.alpha(if (puede) 1f else 0.5f).clip(forma).background(C.superficie).border(1.dp, C.linea, forma).clickable(enabled = puede, onClick = elegir).padding(10.dp),
        verticalArrangement = Arrangement.spacedBy(6.dp)) {
        BoxWithConstraints(Modifier.fillMaxWidth()) { MiniaturaBloque(t.tipo, t.variantes.firstOrNull() ?: "", this.maxWidth) }
        Texto(t.nombre, Estilo.subheadFuerte, C.texto, lineas = 2)
        Texto(if (puede) t.descripcion else "Ya está en esta página", Estilo.footnote, C.texto2, lineas = 3)
    }
}

// MARK: - Panel del bloque

/** La variante con miniaturas, qué le falta y los campos generados desde bloques.json. */
@Composable
fun PanelBloque(editor: EditorTienda, id: String, cerrar: () -> Unit) {
    val datos = LocalDatosEditor.current
    val estado = LocalEstado.current
    val nav = LocalNavegacion.current
    val b = editor.bloque(id)
    val t = b?.let { CatalogoBloques.shared.tipo(it.tipo) }
    Column(Modifier.fillMaxSize().background(C.elevado).imePadding()) {
        Grabber()
        Box(Modifier.fillMaxWidth().height(44.dp).padding(horizontal = 16.dp)) {
            Texto("Listo", Estilo(17f, Peso.semibold), C.primarioTexto, Modifier.align(Alignment.CenterEnd).clickable(onClick = cerrar))
        }
        if (b == null || t == null) {
            Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) { Texto("Este bloque ya no está.", Estilo.body16, C.texto2) }
            return@Column
        }
        Column(Modifier.weight(1f).verticalScroll(rememberScrollState()).padding(20.dp), verticalArrangement = Arrangement.spacedBy(18.dp)) {
            Column(verticalArrangement = Arrangement.spacedBy(4.dp)) {
                Texto(t.nombre, Estilo.tituloHojaChico, C.texto)
                Texto(t.descripcion, Estilo.subhead, C.texto2)
            }
            val f = FaltaBloque.de(b, datos.contexto)
            if (b.oculto) {
                AvisoPanel("Está oculto: no se ve en tu tienda.", "eye.slash") { BotonTonal("Mostrar", { editor.alternarOculto(id) }, alto = 34.dp) }
            } else if (f != null) {
                AvisoPanel(f.texto, "exclamationmark.circle.fill") {
                    if (f.enNegocio) BotonTonal("Cargar ahora", { cerrar(); nav.abrir(Destino.negocio) }, alto = 34.dp)
                }
            }
            if (t.variantes.size > 1) {
                SelectorVariante(t.tipo, t.variantes, t.nombresVariantes, b.variante) { v -> editor.editar(id) { it.copy(variante = v) } }
            }
            EditorCampos(t.campos, b.datos) { d -> editor.editar(id, "bloque-$id") { it.copy(datos = d) } }
            Separador(0.dp)
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                BotonTonal(if (b.oculto) "Mostrar" else "Ocultar", { editor.alternarOculto(id) }, Modifier.weight(1f), alto = 44.dp, simbolo = if (b.oculto) "eye" else "eye.slash", lleno = true)
                BotonTonal("Duplicar", { editor.duplicar(id) }, Modifier.weight(1f), alto = 44.dp, simbolo = "plus.square.on.square", lleno = true)
            }
            BotonTexto("Borrar este bloque", {
                val nombre = t.nombre
                cerrar()
                editor.borrar(id)
                estado.avisar("Borramos «$nombre»", deshacer = { editor.deshacer() })
            }, alto = 44.dp, color = C.errorTinta)
        }
    }
}

/** El aviso del panel: por qué no se ve y, si aplica, el botón para arreglarlo. */
@Composable
fun AvisoPanel(texto: String, simbolo: String, accion: @Composable () -> Unit = {}) {
    Row(Modifier.fillMaxWidth().clip(RoundedCornerShape(14.dp)).background(C.avisoFondo).padding(12.dp), horizontalArrangement = Arrangement.spacedBy(10.dp),
        verticalAlignment = Alignment.CenterVertically) {
        Icono(simbolo, 16f, C.avisoTinta)
        Texto(texto, Estilo.subheadFuerte, C.avisoTinta, Modifier.weight(1f))
        accion()
    }
}

/** Las variantes de un bloque o de una página, con su miniatura. */
@Composable
fun SelectorVariante(tipo: String, variantes: List<String>, nombres: List<String>, elegida: String, alElegir: (String) -> Unit) {
    Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
        Texto("Diseño", Estilo.footnoteFuerte, C.texto2)
        Row(Modifier.fillMaxWidth().horizontalScroll(rememberScrollState()).padding(vertical = 2.dp), horizontalArrangement = Arrangement.spacedBy(10.dp)) {
            variantes.forEachIndexed { i, v ->
                val on = elegida == v
                Column(Modifier.clickable { Haptico.seleccion(); alElegir(v) }.semantics { contentDescription = nombres.getOrNull(i) ?: v }.padding(2.dp),
                    verticalArrangement = Arrangement.spacedBy(6.dp)) {
                    Box(Modifier.border(2.5.dp, if (on) C.acento else Color.Transparent, RoundedCornerShape(10.dp)).padding(2.dp)) {
                        MiniaturaBloque(tipo, v, 112.dp)
                    }
                    Texto(nombres.getOrNull(i) ?: v, Estilo.footnoteFuerte, if (on) C.texto else C.texto2, Modifier.width(116.dp))
                }
            }
        }
    }
}

// MARK: - Ajustes de la página

/** Catálogo, Ficha y Carrito: la variante y los campos de bloques.json › paginas. */
@Composable
fun AjustesPaginaVista(editor: EditorTienda, pagina: String) {
    val p = CatalogoBloques.shared.paginas[pagina]
    TarjetaSeccion(p?.nombre ?: pagina, "slider.horizontal.3") {
        val v = p?.variantes
        if (v != null && v.size > 1) {
            SelectorVariante(pagina, v, p?.nombresVariantes ?: v, editor.ajustes(pagina).variante ?: v[0]) { n -> editor.editarAjustes(pagina) { it.copy(variante = n) } }
        }
        val campos = p?.campos
        if (!campos.isNullOrEmpty()) {
            EditorCampos(campos, editor.ajustes(pagina).campos) { d -> editor.editarAjustes(pagina, "pagina-$pagina") { it.copy(campos = d) } }
        }
    }
}

