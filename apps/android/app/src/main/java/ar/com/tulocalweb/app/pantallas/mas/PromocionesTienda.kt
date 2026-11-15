package ar.com.tulocalweb.app.pantallas.mas

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.semantics.clearAndSetSemantics
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.nucleo.CatalogoBloques
import ar.com.tulocalweb.app.nucleo.FaltaVentana
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.MetricaVentana
import ar.com.tulocalweb.app.nucleo.Valor
import ar.com.tulocalweb.app.nucleo.Ventana
import ar.com.tulocalweb.app.nucleo.get
import ar.com.tulocalweb.app.nucleo.texto
import ar.com.tulocalweb.app.tema.*
import kotlinx.coroutines.delay

// Clon de Pantallas/Mas/PromocionesTienda.swift.
// Mi tienda › Promociones (docs/especificaciones/tienda-promociones.md § 4): las ventanas (popups) en orden, crear desde
// plantillas, el panel de cada una generado desde bloques.json › ventanas, sus métricas y las herramientas de conversión
// (estilo.conversion). Igual que el editor web.

/** La pestaña Promociones del editor. */
@Composable
fun PestanaPromociones(editor: EditorTienda, verEnVista: (String) -> Unit) {
    val datos = LocalDatosEditor.current
    val cat = CatalogoBloques.shared.ventanas
    Column(verticalArrangement = Arrangement.spacedBy(14.dp)) {
        TarjetaSeccion(cat?.nombre ?: "Ventanas", "sparkles") {
            val ventanas = editor.diseno.ventanas
            if (ventanas.isEmpty()) {
                Column(Modifier.fillMaxWidth().padding(vertical = 8.dp), horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(12.dp)) {
                    MiniaturaVentana("centrada", 92.dp)
                    Texto("Todavía no tenés ventanas", Estilo.headline, C.texto)
                    Texto(cat?.descripcion ?: "Avisos que se abren solos.", Estilo.subhead, C.texto2, alinear = TextAlign.Center)
                    BotonPrincipal("Crear la primera", { editor.creandoVentana = true })
                }
            } else {
                Texto("Si varias pueden salir a la vez, sale la primera de la lista.", Estilo.footnote, C.texto2)
                ventanas.forEachIndexed { i, v ->
                    FilaVentana(editor, v, datos.metricas[v.id], primera = i == 0, ultima = i == ventanas.size - 1, verEnVista = verEnVista)
                }
                BotonSecundario("Crear una ventana", { editor.creandoVentana = true }, Modifier.fillMaxWidth(), simbolo = "plus", habilitado = editor.puedeSumarVentana, lleno = true)
                if (!editor.puedeSumarVentana) Texto("Llegaste al tope de ${editor.topeVentanas} ventanas.", Estilo.footnote, C.texto2)
                else if (datos.metricasCargadas) Texto("Las cifras son de los últimos 30 días.", Estilo.footnote, C.texto3)
            }
        }
        HerramientasConversion(editor)
    }
}

/** «320 vistas · 6 % dejó el mail · 4 ventas». */
fun cifrasVentana(m: MetricaVentana, contenido: String): String {
    val p = mutableListOf("${Formato.numero(m.vistas)} ${if (m.vistas == 1) "vista" else "vistas"}")
    m.tasa(contenido)?.let { t -> p += "${Formato.porcentaje(Math.round(t).toDouble())} ${if (contenido == "captacion" || contenido == "ruleta") "dejó el mail" else "tocó"}" }
    if (m.ventas > 0) p += "${m.ventas} ${if (m.ventas == 1) "venta" else "ventas"}"
    return p.joinToString(" · ")
}

/** Una ventana en la lista: miniatura, nombre, qué muestra y cuándo, qué le falta o cómo le va, el interruptor y el menú. */
@Composable
fun FilaVentana(editor: EditorTienda, ventana: Ventana, metrica: MetricaVentana?, primera: Boolean, ultima: Boolean, verEnVista: (String) -> Unit) {
    val estado = LocalEstado.current
    val cat = CatalogoBloques.shared.ventanas
    val falta = FaltaVentana.de(ventana)
    val c = cat?.contenido(ventana.contenido)?.nombre ?: ventana.contenido
    val d = cat?.disparo(ventana.cuando)?.nombre ?: ventana.cuando
    val subtitulo = if (ventana.activa) "$c · $d" else "Apagada · $c"
    val forma = RoundedCornerShape(16.dp)
    fun borrar() {
        val nombre = ventana.nombre
        editor.borrarVentana(ventana.id)
        Haptico.impacto()
        estado.avisar("Borramos «$nombre»", deshacer = { editor.deshacer() })
    }
    Row(Modifier.fillMaxWidth().alpha(if (ventana.activa) 1f else 0.6f).clip(forma).background(C.superficie).border(1.dp, C.linea, forma)
        .padding(start = 12.dp, top = 10.dp, bottom = 10.dp, end = 4.dp), horizontalArrangement = Arrangement.spacedBy(10.dp), verticalAlignment = Alignment.CenterVertically) {
        Row(Modifier.weight(1f).clickable { editor.ventanaAbierta = ventana.id }, horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
            MiniaturaVentana(ventana.variante, 46.dp)
            Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(3.dp)) {
                Texto(ventana.nombre.ifEmpty { "Sin nombre" }, Estilo.fila, C.texto, lineas = 1)
                Texto(subtitulo, Estilo.footnote, C.texto2, lineas = 1)
                if (falta != null && ventana.activa) {
                    Row(horizontalArrangement = Arrangement.spacedBy(4.dp), verticalAlignment = Alignment.CenterVertically) {
                        Icono("exclamationmark.circle.fill", 12f, C.avisoTinta)
                        Texto(falta, Estilo.footnoteFuerte, C.avisoTinta, lineas = 2)
                    }
                } else if (metrica != null && metrica.vistas > 0) {
                    Texto(cifrasVentana(metrica, ventana.contenido), Estilo.footnote, C.texto3, lineas = 1, numeros = true)
                }
            }
        }
        Box(Modifier.semantics { contentDescription = if (ventana.activa) "Apagar ${ventana.nombre}" else "Prender ${ventana.nombre}" }) {
            Llave(ventana.activa, { Haptico.seleccion(); editor.alternarVentana(ventana.id) })
        }
        MenuBoton(
            listOfNotNull(
                OpcionDeMenu("Editar", "pencil") { editor.ventanaAbierta = ventana.id },
                OpcionDeMenu("Ver en la vista previa", "eye") { verEnVista(ventana.id) },
                if (editor.puedeSumarVentana) OpcionDeMenu("Duplicar", "plus.square.on.square") { editor.duplicarVentana(ventana.id) } else null,
                if (!primera) OpcionDeMenu("Subir", "arrow.up") { editor.moverVentana(ventana.id, -1) } else null,
                if (!ultima) OpcionDeMenu("Bajar", "arrow.down") { editor.moverVentana(ventana.id, 1) } else null,
                OpcionDeMenu("Borrar", "trash", destructiva = true) { borrar() },
            ),
            "Más acciones de ${ventana.nombre}",
        ) {
            Box(Modifier.size(44.dp), contentAlignment = Alignment.Center) { Icono("ellipsis", 17f, C.texto2) }
        }
    }
}

/** Cómo se ve una ventana sobre la tienda: centrada, en un costado, completa o desde abajo. */
@Composable
fun MiniaturaVentana(variante: String, ancho: Dp) {
    val alto = ancho * 1.3f
    val r = RoundedCornerShape(ancho * 0.12f)
    val t = RoundedCornerShape(ancho * 0.08f)
    Box(Modifier.size(ancho, alto).clip(r).background(C.crema2).clearAndSetSemantics {}) {
        Column(Modifier.padding(ancho * 0.12f), verticalArrangement = Arrangement.spacedBy(ancho * 0.07f)) {
            repeat(4) { i ->
                Box(Modifier.size(ancho * (if (i % 2 == 0) 0.62f else 0.42f), maxOf(2.dp, ancho * 0.06f)).clip(Capsula).background(C.linea2))
            }
        }
        if (variante != "lateral") Box(Modifier.fillMaxSize().background(Color.Black.copy(alpha = 0.18f)))
        @Composable
        fun Contenido(m: Modifier) {
            Column(m, horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(ancho * 0.05f, Alignment.CenterVertically)) {
                Box(Modifier.size(ancho * 0.3f, maxOf(2.dp, ancho * 0.05f)).clip(Capsula).background(C.texto.copy(alpha = 0.7f)))
                Box(Modifier.size(ancho * 0.24f, maxOf(3.dp, ancho * 0.09f)).clip(Capsula).background(C.acento))
            }
        }
        when (variante) {
            "lateral" -> Contenido(Modifier.align(Alignment.BottomEnd).padding(ancho * 0.07f).size(ancho * 0.48f, alto * 0.24f).clip(t).background(C.superficie).border(1.dp, C.linea2, t))
            "completa" -> Contenido(Modifier.fillMaxSize().padding(ancho * 0.05f).clip(t).background(C.superficie))
            "inferior" -> Contenido(Modifier.align(Alignment.BottomCenter).fillMaxWidth().height(alto * 0.42f)
                .clip(RoundedCornerShape(topStart = ancho * 0.12f, topEnd = ancho * 0.12f)).background(C.superficie))
            else -> Contenido(Modifier.align(Alignment.Center).size(ancho * 0.66f, alto * 0.42f).clip(t).background(C.superficie))
        }
    }
}

// MARK: - Plantillas

fun simboloContenido(contenido: String): String = when (contenido) {
    "cupon" -> "ticket"
    "captacion" -> "envelope"
    "ruleta" -> "gift"
    else -> "megaphone"
}

/** Crear una ventana: las plantillas del catálogo y «En blanco». */
@Composable
fun PlantillasVentana(editor: EditorTienda, cerrar: () -> Unit) {
    val estado = LocalEstado.current
    val datos = LocalDatosEditor.current
    val cat = CatalogoBloques.shared.ventanas
    fun crear(v: Ventana) {
        Haptico.seleccion()
        cerrar()
        editor.creandoVentana = false
        val id = editor.agregarVentana(v, abrir = false) ?: return
        estado.tarea { delay(500); editor.ventanaAbierta = id }
    }
    @Composable
    fun Tarjeta(titulo: String, texto: String, simbolo: String, variante: String, accion: () -> Unit) {
        val forma = RoundedCornerShape(16.dp)
        Row(Modifier.fillMaxWidth().clip(forma).background(C.superficie).border(1.dp, C.linea, forma).clickable(onClick = accion).padding(12.dp),
            horizontalArrangement = Arrangement.spacedBy(14.dp), verticalAlignment = Alignment.CenterVertically) {
            MiniaturaVentana(variante, 46.dp)
            Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(3.dp)) {
                Row(horizontalArrangement = Arrangement.spacedBy(6.dp), verticalAlignment = Alignment.CenterVertically) {
                    Icono(simbolo, 14f, C.texto)
                    Texto(titulo, Estilo.fila, C.texto)
                }
                Texto(texto, Estilo.footnote, C.texto2)
            }
            Chevron()
        }
    }
    Column(Modifier.fillMaxSize().background(C.fondo)) {
        Grabber()
        Box(Modifier.fillMaxWidth().height(44.dp).padding(horizontal = 16.dp)) {
            Texto("Cancelar", Estilo(17f, Peso.regular), C.primarioTexto, Modifier.align(Alignment.CenterStart).clickable(onClick = cerrar))
            Texto("Crear una ventana", Estilo.headline, C.texto, Modifier.align(Alignment.Center))
        }
        Column(Modifier.weight(1f).verticalScroll(rememberScrollState()).margenPantalla().padding(vertical = 12.dp), verticalArrangement = Arrangement.spacedBy(10.dp)) {
            Texto("Elegí una para empezar: después cambiás todo.", Estilo.subhead, C.texto2)
            (cat?.plantillas ?: emptyList()).forEach { p ->
                Tarjeta(p.nombre, p.descripcion, simboloContenido(p.ventana["contenido"]?.texto ?: "promo"), p.ventana["variante"]?.texto ?: "centrada") {
                    crear(Ventana.dePlantilla(p, editor.idNuevo(), datos.envioGratis))
                }
            }
            Tarjeta("En blanco", "Una promo centrada al entrar. La armás a tu gusto.", "plus", "centrada") { crear(Ventana.enBlanco(editor.idNuevo())) }
        }
    }
}

// MARK: - Panel de la ventana

/** Todo lo de una ventana: nombre, qué muestra, diseño, contenido, cuándo sale, a quién, dónde y fechas. */
@Composable
fun PanelVentana(editor: EditorTienda, id: String, verEnVista: (String) -> Unit, cerrar: () -> Unit) {
    val datos = LocalDatosEditor.current
    val estado = LocalEstado.current
    val cat = CatalogoBloques.shared.ventanas
    val v = editor.ventana(id)
    fun editar(clave: String? = null, f: (Ventana) -> Ventana) = editor.editarVentana(id, clave, f)
    Column(Modifier.fillMaxSize().background(C.elevado).imePadding()) {
        Grabber()
        Box(Modifier.fillMaxWidth().height(44.dp).padding(horizontal = 16.dp)) {
            Texto("Listo", Estilo(17f, Peso.semibold), C.primarioTexto, Modifier.align(Alignment.CenterEnd).clickable(onClick = cerrar))
        }
        if (v == null || cat == null) {
            Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) { Texto("Esta ventana ya no está.", Estilo.body16, C.texto2) }
            return@Column
        }
        val cont = cat.contenido(v.contenido)
        @Composable
        fun Seccion(titulo: String, contenido: @Composable () -> Unit) {
            Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
                Texto(titulo, Estilo.headline, C.texto)
                contenido()
            }
        }
        @Composable
        fun Rango(titulo: String, r: CatalogoBloques.Ventanas.Rango, valor: Int?, poner: (Int) -> Unit) {
            val n = valor ?: r.porDefecto
            Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
                Texto(titulo, Estilo.body16, C.texto, Modifier.weight(1f))
                Texto(if (r.unidad == "%") "$n %" else "$n", Estilo.fila, C.texto, Modifier.padding(end = 10.dp), numeros = true)
                Pasador(n, { poner(it.coerceIn(r.min, r.max)) }, r.min..r.max, if (r.max - r.min > 40) 5 else 1)
            }
        }
        @Composable
        fun Cifra(valor: String, texto: String, m: Modifier) {
            Column(m.clip(RoundedCornerShape(14.dp)).background(C.superficie).padding(12.dp), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                Texto(valor, Estilo.headline, C.texto, lineas = 1, numeros = true, escalaMinima = 0.8f)
                Texto(texto, Estilo.footnote, C.texto2, lineas = 1, escalaMinima = 0.8f)
            }
        }
        Column(Modifier.weight(1f).verticalScroll(rememberScrollState()).padding(20.dp), verticalArrangement = Arrangement.spacedBy(22.dp)) {
            Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
                CampoLinea(cat.campos.firstOrNull()?.nombre ?: "Nombre", v.nombre, { n -> editar("ventana-nombre-$id") { it.copy(nombre = n.take(40)) } }, "Bienvenida 10 %", 40)
                Interruptor("Prendida", v.activa, { a -> editar { it.copy(activa = a) } })
                val f = FaltaVentana.de(v)
                if (!v.activa) AvisoPanel("Está apagada: no sale en tu tienda.", "eye.slash")
                else if (f != null) AvisoPanel("$f Hasta entonces no sale.", "exclamationmark.circle.fill")
                BotonTonal("Ver en la vista previa", { cerrar(); verEnVista(id) }, Modifier.fillMaxWidth(), alto = 44.dp, simbolo = "eye", lleno = true)
            }
            val m = datos.metricas[id]
            if (m != null && m.vistas > 0) {
                val capta = v.contenido == "captacion" || v.contenido == "ruleta"
                Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
                    Texto("Últimos 30 días", Estilo.footnoteFuerte, C.texto2)
                    Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                        Cifra(Formato.numero(m.vistas), if (m.vistas == 1) "vista" else "vistas", Modifier.weight(1f))
                        Cifra(m.tasa(v.contenido)?.let { Formato.porcentaje(Math.round(it).toDouble()) } ?: "—", if (capta) "dejó el mail" else "tocó", Modifier.weight(1f))
                        Cifra("${m.ventas}", if (m.ventas == 1) "venta" else "ventas", Modifier.weight(1f))
                    }
                    if (m.monto > 0) Texto("Vendió ${Formato.pesos(m.monto)} con su cupón.", Estilo.footnote, C.texto2)
                }
            }
            Seccion("Qué muestra") {
                cat.contenidos.forEach { c ->
                    FilaOpcion(c.nombre, c.descripcion, v.contenido == c.tipo, simboloContenido(c.tipo)) { editar { it.conContenido(c.tipo) } }
                }
            }
            Seccion("Diseño") {
                Row(Modifier.fillMaxWidth().horizontalScroll(rememberScrollState()), horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                    cat.variantes.forEachIndexed { i, va ->
                        val on = v.variante == va
                        val nombre = cat.nombresVariantes.getOrNull(i) ?: va
                        Column(Modifier.clickable { Haptico.seleccion(); editar { it.copy(variante = va) } }.semantics { contentDescription = nombre }.padding(3.dp),
                            verticalArrangement = Arrangement.spacedBy(6.dp)) {
                            Box(Modifier.border(2.5.dp, if (on) C.acento else Color.Transparent, RoundedCornerShape(11.dp)).padding(3.dp)) { MiniaturaVentana(va, 76.dp) }
                            Texto(nombre, Estilo.footnoteFuerte, if (on) C.texto else C.texto2, Modifier.width(82.dp), lineas = 2)
                        }
                    }
                }
                val i = cat.variantes.indexOf(v.variante)
                cat.descripcionesVariantes?.getOrNull(i)?.let { Texto(it, Estilo.footnote, C.texto2) }
            }
            if (cont != null) {
                Seccion(cont.nombre) {
                    EditorCampos(cont.campos, v.datos) { d -> editar("ventana-$id") { it.copy(datos = d) } }
                }
            }
            Seccion("Cuándo sale") {
                cat.disparos.forEach { d -> FilaOpcion(d.nombre, d.descripcion, v.cuando == d.cuando) { editar { it.copy(cuando = d.cuando) } } }
                cat.disparo(v.cuando)?.valor?.let { r ->
                    Rango(if (r.unidad == "%") "Qué parte de la página" else "Cuántos segundos", r, v.valor) { n -> editar("ventana-valor-$id") { it.copy(valor = n) } }
                }
                Texto("Cuántas veces", Estilo.subheadFuerte, C.texto, Modifier.padding(top = 6.dp))
                cat.frecuencias.forEach { f -> FilaOpcion(f.nombre, f.descripcion, v.cada == f.cada) { editar { it.copy(cada = f.cada) } } }
                cat.frecuencia(v.cada)?.dias?.let { r -> Rango("Cada cuántos días", r, v.dias) { n -> editar("ventana-dias-$id") { it.copy(dias = n) } } }
            }
            Seccion("A quién") {
                SelectorOpcion("Quiénes la ven", cat.publicos.opciones, cat.publicos.nombres, v.publico) { p -> editar { it.copy(publico = p) } }
                SelectorOpcion("En qué pantallas", cat.dispositivos.opciones, cat.dispositivos.nombres, v.dispositivos) { p -> editar { it.copy(dispositivos = p) } }
            }
            Seccion("Dónde") {
                SelectorOpciones("En qué páginas", cat.paginas.opciones, cat.paginas.nombres, v.paginas) { p -> editar { it.copy(paginas = p) } }
                cat.paginas.descripcion?.let { Texto(it, Estilo.footnote, C.texto2) }
            }
            Seccion("Fechas") {
                EditorFecha("Empieza", v.desde?.let { Valor.texto(it) }, { f -> editar { it.copy(desde = f?.texto) } }, boton = "Elegir cuándo empieza", pasada = "Ya empezó.", quitable = true)
                EditorFecha("Termina", v.hasta?.let { Valor.texto(it) }, { f -> editar { it.copy(hasta = f?.texto) } }, boton = "Elegir cuándo termina", quitable = true)
                Texto("Sin fechas, sale mientras esté prendida.", Estilo.footnote, C.texto2)
            }
            Separador(0.dp)
            BotonTonal("Duplicar", { editor.duplicarVentana(id); estado.avisar("Duplicamos «${v.nombre}»: quedó apagada") }, Modifier.fillMaxWidth(), alto = 44.dp,
                simbolo = "plus.square.on.square", habilitado = editor.puedeSumarVentana, lleno = true)
            BotonTexto("Borrar esta ventana", {
                val nombre = v.nombre
                cerrar()
                editor.borrarVentana(id)
                estado.avisar("Borramos «$nombre»", deshacer = { editor.deshacer() })
            }, alto = 44.dp, color = C.errorTinta)
        }
    }
}

/** Una opción con su explicación, que se elige tocándola (contenido, disparo, frecuencia). */
@Composable
fun FilaOpcion(titulo: String, detalle: String, elegida: Boolean, simbolo: String? = null, accion: () -> Unit) {
    val forma = RoundedCornerShape(14.dp)
    Row(Modifier.fillMaxWidth().clip(forma).background(C.superficie).border(if (elegida) 2.dp else 1.dp, if (elegida) C.acento else C.linea, forma)
        .clickable { Haptico.seleccion(); accion() }.padding(12.dp), horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
        if (simbolo != null) Box(Modifier.width(22.dp), contentAlignment = Alignment.Center) { Icono(simbolo, 15f, C.primarioTexto) }
        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
            Texto(titulo, Estilo.fila, C.texto)
            Texto(detalle, Estilo.footnote, C.texto2)
        }
        Icono(if (elegida) "checkmark.circle.fill" else "circle", 20f, if (elegida) C.acento else C.linea2)
    }
}

// MARK: - Herramientas de conversión

/** estilo.conversion: interruptores con su explicación corta, generados desde bloques.json › conversion. */
@Composable
fun HerramientasConversion(editor: EditorTienda) {
    val c = CatalogoBloques.shared.conversion ?: return
    TarjetaSeccion(c.nombre, "bolt") {
        Texto(c.descripcion, Estilo.footnote, C.texto2)
        EditorCampos(c.campos, editor.diseno.estilo.conversion ?: emptyMap()) { editor.editarConversion(it) }
    }
}
