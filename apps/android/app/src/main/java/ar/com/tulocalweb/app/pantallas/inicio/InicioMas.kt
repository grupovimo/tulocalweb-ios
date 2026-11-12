package ar.com.tulocalweb.app.pantallas.inicio

import androidx.compose.animation.core.RepeatMode
import androidx.compose.animation.core.animateFloat
import androidx.compose.animation.core.infiniteRepeatable
import androidx.compose.animation.core.rememberInfiniteTransition
import androidx.compose.animation.core.tween
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.drawBehind
import androidx.compose.ui.draw.drawWithContent
import androidx.compose.ui.draw.rotate
import androidx.compose.ui.draw.scale
import androidx.compose.ui.geometry.CornerRadius
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.geometry.Size
import androidx.compose.ui.graphics.BlendMode
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.graphics.CompositingStrategy
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.heading
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.semantics.stateDescription
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.app.Destino
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.app.LocalNavegacion
import ar.com.tulocalweb.app.app.Navegacion
import ar.com.tulocalweb.app.app.Pestana
import ar.com.tulocalweb.app.app.Ruta
import ar.com.tulocalweb.app.nucleo.Acceso
import ar.com.tulocalweb.app.nucleo.Fecha
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.Marcas
import ar.com.tulocalweb.app.nucleo.PaginaProductos
import ar.com.tulocalweb.app.nucleo.Preferencias
import ar.com.tulocalweb.app.nucleo.Tablero
import ar.com.tulocalweb.app.nucleo.Tincho
import ar.com.tulocalweb.app.nucleo.compartirTexto
import ar.com.tulocalweb.app.nucleo.entero
import ar.com.tulocalweb.app.nucleo.get
import ar.com.tulocalweb.app.nucleo.opcional
import ar.com.tulocalweb.app.nucleo.texto
import ar.com.tulocalweb.app.pantallas.mas.ChipSimple
import ar.com.tulocalweb.app.pantallas.mas.Estrellas
import ar.com.tulocalweb.app.tema.*
import kotlinx.serialization.Serializable
import java.time.Instant
import kotlin.math.abs
import kotlin.math.roundToInt

// Clon de Pantallas/Inicio/InicioMas.swift. Lo que va debajo de la vidriera (docs/handoff/inicio-a): las acciones en un
// bento, «Para hacer hoy», «De un vistazo» de costado (la meta, la caja del local, por dónde compran, quién está ahora,
// las visitas y las liquidaciones), los últimos pedidos, lo más vendido y, al final, lo más visto, las reseñas, los
// envíos y las integraciones. Todo con tarjetas con relieve (tema/Relieve.kt).

/** Título de sección del Inicio (22/700) con un dato o una acción a la derecha. */
@Composable
fun TituloInicio(texto: String, detalle: String? = null, accion: String? = null, tinta: Color = C.primarioTexto, alTocar: (() -> Unit)? = null) {
    Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
        Texto(texto, Estilo.seccion, C.texto, Modifier.weight(1f).semantics { heading() })
        Spacer(Modifier.width(8.dp))
        if (accion != null && alTocar != null) {
            Tocable(alTocar, Modifier.heightIn(min = 44.dp)) {
                Box(Modifier.heightIn(min = 44.dp), contentAlignment = Alignment.Center) { Texto(accion, Estilo.subheadFuerte, tinta) }
            }
        } else if (detalle != null) {
            Texto(detalle, Estilo(14f, Peso.medium), C.texto2)
        }
    }
}

// MARK: - Acciones

/** El bento de 2 columnas: «Subir producto» alta con foto, «Cobrar» con el QR real de la tienda y «Compartir» con una
 *  miniatura de la vidriera. */
@Composable
fun AccionesInicio(t: Tablero, paleta: PaletaTienda) {
    val estado = LocalEstado.current
    val nav = LocalNavegacion.current
    Row(Modifier.fillMaxWidth().height(252.dp), horizontalArrangement = Arrangement.spacedBy(12.dp)) {
        Subir(t, paleta, Modifier.weight(1f).fillMaxHeight())
        Column(Modifier.weight(1f).fillMaxHeight(), verticalArrangement = Arrangement.spacedBy(12.dp)) {
            val puede = estado.puede("pedidos")
            val tienda = estado.tienda
            Tocable({ cobrar(puede, tienda, nav) }, Modifier.weight(1f).fillMaxWidth()) {
                Box(Modifier.fillMaxSize().relieve()) {
                    if (tienda != null) {
                        val forma = RoundedCornerShape(14.dp)
                        Box(Modifier.align(Alignment.TopEnd).offset(x = 6.dp, y = (-6).dp).rotate(-8f).size(76.dp)
                            .sombra(forma, hex(0x2A140C, 0.25f), 6.dp, y = 6.dp).clip(forma).background(Color.White).border(1.dp, hex(0xEADFD3), forma).padding(8.dp)) {
                            CodigoQR(tienda.url, paleta.tintaBlanco, Modifier.fillMaxSize(), "QR de ${tienda.direccion}")
                        }
                    }
                    Pie(if (puede) "Cobrar" else "Tu tienda", if (puede) "Link o QR al toque" else "Mirala como tus clientes")
                }
            }
            val v = t.visitas
            Tocable({ nav.inicio.add(Destino.compartir) }, Modifier.weight(1f).fillMaxWidth()) {
                Box(Modifier.fillMaxSize().relieve()) {
                    MiniVidriera(tienda, t.masVendidos.map { it.foto }, paleta,
                        Modifier.align(Alignment.TopEnd).offset(x = (-8).dp, y = (-14).dp).rotate(7f))
                    Pie("Compartir", if (v != null && v.desde != null) "${Formato.cantidad(v.hoy.actual.toInt(), "visita", "visitas")} hoy" else "Tu vidriera")
                }
            }
        }
    }
}

/** Cobrar con un link o el QR; quien no cobra ve la tienda. */
private fun cobrar(puede: Boolean, tienda: Acceso.Tienda?, nav: Navegacion) {
    if (puede) nav.inicio.add(Destino.links) else if (tienda != null) nav.web = tienda.url
}

@Composable
private fun BoxScope.Pie(titulo: String, detalle: String) {
    Column(Modifier.align(Alignment.BottomStart).padding(14.dp), verticalArrangement = Arrangement.spacedBy(2.dp)) {
        Texto(titulo, Estilo(17f, Peso.bold, -0.012f), C.texto)
        Texto(detalle, Estilo.footnote, C.texto2, lineas = 1, numeros = true, escalaMinima = 0.85f)
    }
}

@Composable
private fun Subir(t: Tablero, paleta: PaletaTienda, modifier: Modifier) {
    val nav = LocalNavegacion.current
    Tocable({ nav.inicio.add(Destino.nuevoProducto(null)) }, modifier) {
        Column(Modifier.fillMaxSize().relieve().padding(12.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
            val forma16 = RoundedCornerShape(16.dp)
            Box(Modifier.weight(1f).fillMaxWidth().clip(forma16)) {
                val foto = t.masVendidos.firstOrNull { it.foto != null }?.foto
                if (foto != null) Foto(foto, Modifier.matchParentSize(), 500)
                else Box(Modifier.matchParentSize().background(tema(0xE9DCCD, 0x3A2419).actual), contentAlignment = Alignment.Center) {
                    ImagenTincho(Tincho.foto, Modifier.fillMaxSize().padding(12.dp))
                }
                // El marco blanco del visor y lo de afuera apenas más oscuro.
                Box(Modifier.matchParentSize().graphicsLayer { compositingStrategy = CompositingStrategy.Offscreen }.drawWithContent {
                    drawRect(Color.Black.copy(alpha = 0.12f))
                    val m = 10.dp.toPx()
                    drawRoundRect(Color.Black, Offset(m, m), Size(size.width - 2 * m, size.height - 2 * m), CornerRadius(m), blendMode = BlendMode.Clear)
                })
                Box(Modifier.matchParentSize().padding(10.dp).border(2.dp, Color.White.copy(alpha = 0.85f), RoundedCornerShape(10.dp)))
                Box(Modifier.align(Alignment.BottomEnd).padding(10.dp).size(40.dp).sombra(CircleShape, Color.Black.copy(alpha = 0.35f), 6.dp, y = 5.dp)
                    .clip(CircleShape).background(paleta.acento)
                    .border(1.dp, Brush.verticalGradient(0f to Color.White.copy(alpha = 0.3f), 0.5f to Color.Transparent), CircleShape),
                    contentAlignment = Alignment.Center) { Icono("camera.fill", 17f, paleta.sobreAcento) }
            }
            Column(Modifier.padding(horizontal = 2.dp), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                Texto("Subir producto", Estilo(17f, Peso.bold, -0.012f), C.texto)
                Texto(Formato.sinViudas("Sacale una foto y Tincho escribe el resto"), Estilo.footnote, C.texto2, interlineado = 1f)
            }
        }
    }
}

/** La vidriera en chiquito: la franja del acento, el logo y dos fotos de productos. */
@Composable
fun MiniVidriera(tienda: Acceso.Tienda?, fotos: List<String?>, paleta: PaletaTienda, modifier: Modifier = Modifier) {
    val forma = RoundedCornerShape(12.dp)
    Box(modifier.size(58.dp, 80.dp).sombra(forma, hex(0x2A140C, 0.25f), 6.dp, y = 6.dp).clip(forma).background(Color.White).border(1.dp, hex(0xEADFD3), forma)) {
        Box(Modifier.fillMaxWidth().height(30.dp).background(paleta.acento))
        val logo = urlDeFoto(tienda?.logo, 60)
        Box(Modifier.offset(7.dp, 19.dp).size(20.dp).sombra(RoundedCornerShape(6.dp), Color.Black.copy(alpha = 0.25f), 1.5.dp, y = 1.dp)
            .clip(RoundedCornerShape(6.dp)).background(Color.White), contentAlignment = Alignment.Center) {
            if (logo != null) coil3.compose.AsyncImage(logo, null, Modifier.fillMaxSize().padding(2.dp), contentScale = ContentScale.Fit)
            else Texto(Formato.iniciales(tienda?.nombre ?: "").take(1), Estilo(11f, Peso.bold), paleta.tintaBlanco)
        }
        Foto(fotos.getOrNull(0), Modifier.offset(7.dp, 45.dp).size(20.dp).clip(RoundedCornerShape(5.dp)), 90)
        Foto(fotos.getOrNull(1), Modifier.offset(31.dp, 45.dp).size(20.dp).clip(RoundedCornerShape(5.dp)), 90)
        Box(Modifier.offset(7.dp, 70.dp).size(44.dp, 4.dp).clip(Capsula).background(hex(0xEFE6DC)))
    }
}

/** Las acciones del día en cero (ACero): sacar una foto, vender en el local y cobrar, en tres. */
@Composable
fun AccionesCero(paleta: PaletaTienda) {
    val estado = LocalEstado.current
    val nav = LocalNavegacion.current
    val puede = estado.puede("pedidos")
    Row(Modifier.fillMaxWidth().height(IntrinsicSize.Max), horizontalArrangement = Arrangement.spacedBy(10.dp)) {
        AccionCero("camera", "Producto", "Sacale una foto", paleta, Modifier.weight(1f)) { nav.inicio.add(Destino.nuevoProducto(null)) }
        AccionCero("barcode.viewfinder", "Vender", "En el local", paleta, Modifier.weight(1f)) { nav.venta = true }
        if (puede) AccionCero("link", "Cobrar", "Link o QR", paleta, Modifier.weight(1f)) { nav.inicio.add(Destino.links) }
        else AccionCero("safari", "Tu tienda", "Mirala", paleta, Modifier.weight(1f)) { estado.tienda?.let { nav.web = it.url } }
    }
}

@Composable
private fun AccionCero(simbolo: String, titulo: String, detalle: String, paleta: PaletaTienda, modifier: Modifier, accion: () -> Unit) {
    Tocable(accion, modifier.fillMaxHeight()) {
        Column(Modifier.fillMaxSize().heightIn(min = 104.dp).relieve().padding(12.dp)) {
            Box(Modifier.size(36.dp).clip(RoundedCornerShape(12.dp)).background(paleta.claro), contentAlignment = Alignment.Center) { Icono(simbolo, 17f, paleta.tinta) }
            Spacer(Modifier.weight(1f).heightIn(min = 8.dp))
            Texto(titulo, Estilo(16f, Peso.bold), C.texto, lineas = 1, escalaMinima = 0.85f)
            Texto(detalle, Estilo.footnote, C.texto2, lineas = 1, escalaMinima = 0.85f)
        }
    }
}

/** «Para tu primera venta» (los primeros pasos de 2d con el estilo de ACero): el avance y cada paso con su acción.
 *  Ningún paso es obligatorio: la tarjeta se puede cerrar y no vuelve. */
@Composable
fun PasosPrimeraVenta(paleta: PaletaTienda) {
    val estado = LocalEstado.current
    val nav = LocalNavegacion.current
    val ctx = LocalContext.current
    var conProductos by remember { mutableStateOf<Boolean?>(null) }
    var mpConectada by remember { mutableStateOf(false) }
    var compartidos by remember { mutableStateOf(Preferencias.texto(PrimerosPasos.claveCompartido) ?: "") }
    val mp = mpConectada || (estado.resumen?.canales?.contains("mp") ?: false)
    val id = estado.tienda?.id ?: ""
    val compartio = PrimerosPasos.contiene(compartidos, id)
    val hechos = 1 + (if (conProductos == true) 1 else 0) + (if (mp) 1 else 0) + (if (compartio) 1 else 0)

    @Serializable data class MP(val conectada: Boolean? = null)
    LaunchedEffect(estado.version) {
        estado.api.opcional<MP>("/v1/panel/integraciones/mercadopago")?.let { mpConectada = it.conectada ?: false }
        estado.api.opcional<PaginaProductos>("/v1/panel/listas/productos", listOf("cuantos" to "1"))?.let { conProductos = it.total > 0 }
    }

    @Composable
    fun Paso(t: String, hecho: Boolean, accion: String?, alTocar: (() -> Unit)?) {
        Row(Modifier.fillMaxWidth().heightIn(min = 44.dp)
            .then(if (!hecho && alTocar != null) Modifier.clickable(onClick = alTocar) else Modifier)
            .semantics(mergeDescendants = true) { if (hecho) stateDescription = "Hecho" },
            horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
            if (hecho) {
                Box(Modifier.size(26.dp).clip(CircleShape).background(C.ok), contentAlignment = Alignment.Center) { Icono("checkmark", 12f, Color.White) }
            } else {
                Box(Modifier.size(26.dp).clip(CircleShape).background(C.superficie).border(2.dp, tema(0xD8C9BB, 0x7D6656).actual, CircleShape))
            }
            Texto(t, Estilo.fila, if (hecho) C.texto2 else C.texto, Modifier.weight(1f))
            if (!hecho && accion != null) Texto(accion, Estilo(14f, Peso.bold), paleta.tinta)
        }
    }

    Column(Modifier.fillMaxWidth().relieve().padding(16.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
        Row(verticalAlignment = Alignment.CenterVertically) {
            Texto("Para tu primera venta", Estilo.seccion, C.texto, Modifier.weight(1f).semantics { heading() })
            Texto("$hechos de 4", Estilo(14f, Peso.bold), paleta.tinta, numeros = true)
        }
        BoxWithConstraints(Modifier.fillMaxWidth().height(8.dp).clip(Capsula).background(C.pista).semantics { contentDescription = "Avance"; stateDescription = "$hechos de 4" }) {
            Box(Modifier.width(maxWidth * hechos / 4).fillMaxHeight().clip(Capsula).background(paleta.acento))
        }
        Column {
            Paso("Creaste tu tienda", true, null, null)
            Paso("Cargá tu primer producto", conProductos == true, "Cargar") { nav.inicio.add(Destino.nuevoProducto(null)) }
            Paso("Conectá Mercado Pago", mp, "Conectar") { nav.inicio.add(Destino.integraciones) }
            // Tocar «Compartir» alcanza para darlo por hecho (no hay forma de saber si se mandó).
            Paso("Mandá el link a tus clientes", compartio, "Compartir") {
                compartidos = PrimerosPasos.sumar(compartidos, id)
                Preferencias.ponerTexto(PrimerosPasos.claveCompartido, compartidos)
                compartirTexto(ctx, estado.tienda?.url ?: "https://tulocalweb.com.ar")
            }
        }
        BotonTexto(if (hechos == 4) "Listo, ir al tablero" else "Ocultar estos pasos", {
            Preferencias.ponerTexto(PrimerosPasos.claveOcultos, PrimerosPasos.sumar(Preferencias.texto(PrimerosPasos.claveOcultos) ?: "", id))
            estado.version += 1
        }, alto = 44.dp, color = paleta.tinta)
    }
}

// MARK: - Para hacer hoy

private enum class Tono { claro, aviso, ok }

private class FilaHacer(val n: Int? = null, val simbolo: String? = null, val titulo: String, val detalle: String, val accion: String? = null,
                        val tono: Tono, val ir: (() -> Unit)? = null)

/** «Para hacer hoy»: filas de 68 con la insignia del número y la acción a la derecha; la primera en el acento. Si no hay
 *  nada urgente o la tienda no tuvo visitas hoy, invita a compartir. */
@Composable
fun ParaHacerHoy(t: Tablero, masViejo: Instant?, paleta: PaletaTienda, cero: Boolean) {
    val estado = LocalEstado.current
    val nav = LocalNavegacion.current
    val ph = t.paraHacer
    val lista = buildList {
        if (ph.preparar > 0) add(FilaHacer(ph.preparar, titulo = "Pedidos por preparar",
            detalle = masViejo?.let { "El más viejo entró ${Formato.hace(it)}" } ?: "Tocá para verlos", accion = "Preparar", tono = Tono.claro) {
            nav.filtroPedidos = "Para preparar"; nav.tocar(Pestana.pedidos)
        })
        if (ph.transferencias > 0) add(FilaHacer(ph.transferencias, titulo = "Transferencias por confirmar",
            detalle = ph.transferenciasMonto?.let { "${Formato.pesos(it)} en total" } ?: "Revisá que hayan llegado", accion = "Confirmar", tono = Tono.claro) {
            nav.filtroPedidos = "Sin pagar"; nav.tocar(Pestana.pedidos)
        })
        if (ph.sinStock > 0) add(FilaHacer(ph.sinStock, titulo = if (ph.sinStock == 1) "Producto sin stock" else "Productos sin stock",
            detalle = "Siguen visibles en la tienda", accion = "Reponer", tono = Tono.aviso) {
            nav.segmentoProductos = "out"; nav.tocar(Pestana.productos)
        })
        val n = t.resenas?.nuevas?.size ?: 0
        if (n > 0 && estado.puede("clientes")) add(FilaHacer(n, titulo = if (n == 1) "Reseña sin responder" else "Reseñas sin responder",
            detalle = "Contestarlas suma confianza", accion = "Responder", tono = Tono.claro) { nav.inicio.add(Destino.resenas) })
        // «Para hacer» nunca queda en «Todo al día» sin una acción (en cero, compartir ya está arriba).
        val hoy = (t.visitas?.hoy?.actual ?: 0.0).toInt()
        if (!cero && (isEmpty() || hoy == 0)) add(FilaHacer(simbolo = "square.and.arrow.up", titulo = "Compartí tu tienda",
            detalle = if (hoy == 0) "Hoy todavía no tuvo visitas" else "Más visitas, más ventas", accion = "Compartir", tono = Tono.claro) {
            nav.inicio.add(Destino.compartir)
        })
    }
    if (lista.isEmpty() && cero) return
    Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
        TituloInicio("Para hacer hoy", if (lista.isEmpty()) "Todo al día" else if (lista.size == 1) "1 cosa" else "${lista.size} cosas")
        Column(Modifier.fillMaxWidth().relieve()) {
            if (lista.isEmpty()) FilaParaHacer(FilaHacer(simbolo = "checkmark", titulo = "Todo al día", detalle = "No hay pedidos ni transferencias esperando.", tono = Tono.ok), false, paleta)
            lista.forEachIndexed { i, f ->
                if (i > 0) Separador(70.dp)
                FilaParaHacer(f, i == 0, paleta)
            }
        }
    }
}

@Composable
private fun FilaParaHacer(f: FilaHacer, primera: Boolean, paleta: PaletaTienda) {
    val (fondo, tinta) = when {
        primera -> paleta.acento to paleta.sobreAcento
        f.tono == Tono.aviso -> C.avisoFondo to C.avisoTinta
        f.tono == Tono.ok -> C.okFondo to C.okTinta
        else -> paleta.claro to paleta.tinta
    }
    val forma = RoundedCornerShape(12.dp)
    val ir = f.ir
    Row(Modifier.fillMaxWidth().heightIn(min = 68.dp)
        .then(if (ir != null) Modifier.clickable(onClick = ir) else Modifier)
        .semantics(mergeDescendants = true) {}
        .padding(horizontal = 16.dp, vertical = 12.dp), horizontalArrangement = Arrangement.spacedBy(14.dp), verticalAlignment = Alignment.CenterVertically) {
        Box(Modifier.size(40.dp).clip(forma).background(fondo)
            .then(if (primera) Modifier.border(1.dp, Brush.verticalGradient(0f to Color.White.copy(alpha = 0.25f), 0.5f to Color.Transparent), forma) else Modifier),
            contentAlignment = Alignment.Center) {
            if (f.n != null) Texto(Formato.numero(f.n), Estilo(17f, Peso.bold), tinta, lineas = 1, numeros = true, escalaMinima = 0.6f)
            else if (f.simbolo != null) Icono(f.simbolo, 16f, tinta)
        }
        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
            Texto(f.titulo, Estilo.fila, C.texto)
            Texto(f.detalle, Estilo.footnote, C.texto2, numeros = true)
        }
        if (f.accion != null) Texto(f.accion, Estilo(14f, Peso.bold), paleta.tinta)
    }
}

// MARK: - De un vistazo

/** «De un vistazo»: tarjetas de 164 × 156 que se deslizan de costado. Cada una aparece si tiene algo. */
@Composable
fun DeUnVistazo(t: Tablero, periodo: String, paleta: PaletaTienda, editarMeta: () -> Unit) {
    val nav = LocalNavegacion.current
    val m = margen
    val vivo = t.enVivo?.takeIf { it.personas > 0 || it.carritos > 0 }
    val meta = if (t.montos) t.meta else null
    val caja = if (t.montos) t.caja else null
    val canales = if (t.montos && (t.ventas?.actual ?: 0.0) > 0) t.canales ?: emptyList() else emptyList()
    val liquidaciones = if (t.montos) t.liquidaciones ?: emptyList() else emptyList()
    val visitas = t.visitas
    if (vivo == null && meta == null && caja == null && canales.isEmpty() && visitas == null && liquidaciones.isEmpty()) return
    Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
        TituloInicio("De un vistazo")
        Row(Modifier.fillMaxWidth().sinMargen().horizontalScroll(rememberScrollState()).padding(horizontal = m).padding(bottom = 24.dp),
            horizontalArrangement = Arrangement.spacedBy(12.dp)) {
            if (vivo != null) EnVivo(vivo) { if (vivo.carritos > 0) nav.inicio.add(Destino.carritos) else nav.inicio.add(Destino.estadisticas) }
            if (meta != null) TarjetaMeta(meta, paleta, editarMeta)
            if (caja != null) Vistazo({ nav.inicio.add(Destino.caja) }) {
                Etiqueta("Caja del local")
                Spacer(Modifier.weight(1f).heightIn(min = 4.dp))
                Grande(Formato.pesos(caja.total))
                Spacer(Modifier.weight(1f).heightIn(min = 4.dp))
                if (caja.total > 0) Chip("Cobrado hoy", C.okFondo, C.okTinta) else Chip("Sin cobros hoy", C.neutroFondo, C.neutroTinta)
            }
            if (canales.isNotEmpty()) TarjetaCanales(canales, paleta) { nav.inicio.add(Destino.estadisticas) }
            if (visitas != null) TarjetaVisitas(visitas, periodo, paleta) { if (visitas.desde == null) nav.inicio.add(Destino.compartir) else nav.inicio.add(Destino.estadisticas) }
            liquidaciones.firstOrNull()?.let { l ->
                val mas = liquidaciones.size - 1
                Vistazo({ nav.inicio.add(Destino.finanzas) }) {
                    Etiqueta(if (mas > 0) "Próximas liquidaciones" else "Próxima liquidación")
                    Spacer(Modifier.weight(1f).heightIn(min = 4.dp))
                    Grande(Formato.pesos(l.monto), C.okTinta)
                    Spacer(Modifier.weight(1f).heightIn(min = 4.dp))
                    Texto("${Marcas.pago(l.medio)} · ${Fecha.iso(l.fecha)?.let { Formato.dia(it) } ?: l.fecha}", Estilo(12f, Peso.regular), C.texto2, lineas = 1, escalaMinima = 0.8f)
                    if (mas > 0) Texto(if (mas == 1) "y 1 más" else "y $mas más", Estilo(12f, Peso.regular), C.texto2)
                }
            }
        }
    }
}

@Composable
private fun Vistazo(accion: () -> Unit, modifier: Modifier = Modifier, contenido: @Composable ColumnScope.() -> Unit) {
    Tocable(accion, modifier) {
        Column(Modifier.size(164.dp, 156.dp).relieve().padding(16.dp), content = contenido)
    }
}

@Composable
private fun Etiqueta(s: String, color: Color = C.texto2) = Texto(s, Estilo.footnoteFuerte, color, lineas = 1, escalaMinima = 0.85f)

@Composable
private fun Grande(s: String, color: Color = C.texto) = Texto(s, Estilo(26f, Peso.bold, -0.023f), color, lineas = 1, numeros = true, escalaMinima = 0.6f)

@Composable
private fun Chip(s: String, fondo: Color, tinta: Color) {
    Texto(s, Estilo(12f, Peso.bold), tinta, Modifier.clip(Capsula).background(fondo).padding(horizontal = 8.dp, vertical = 3.dp), lineas = 1, escalaMinima = 0.8f)
}

@Composable
private fun EnVivo(v: Tablero.EnVivo, accion: () -> Unit) {
    Vistazo(accion) {
        Row(horizontalArrangement = Arrangement.spacedBy(6.dp), verticalAlignment = Alignment.CenterVertically) {
            Latido()
            Etiqueta("Ahora en tu tienda", C.okTinta)
        }
        Spacer(Modifier.weight(1f).heightIn(min = 4.dp))
        Row(horizontalArrangement = Arrangement.spacedBy(6.dp), verticalAlignment = Alignment.Bottom) {
            Grande(Formato.numero(maxOf(0, v.personas)))
            Texto("mirando", Estilo.footnote, C.texto2, Modifier.padding(bottom = 5.dp))
        }
        Spacer(Modifier.weight(1f).heightIn(min = 4.dp))
        val p = v.paginas.firstOrNull()
        if (v.carritos > 0) Chip(if (v.carritos == 1) "1 carrito armándose" else "${v.carritos} carritos armándose", C.avisoFondo, C.avisoTinta)
        else if (p != null) Texto("Miran ${p.nombre}", Estilo(12f, Peso.regular), C.texto2, lineas = 2)
    }
}

@Composable
private fun TarjetaMeta(m: Tablero.Meta, paleta: PaletaTienda, editar: () -> Unit) {
    val pct = if (m.objetivo > 0) m.actual.toDouble() / m.objetivo else 0.0
    val suave = hex(0xCDB8A8)
    val pincel = pincelGrano()
    val pista = hex(0x4A2D1F)
    Tocable(editar) {
        Column(Modifier.size(164.dp, 156.dp).relieve(fondo = tema(0x2A140C, 0x3A2216).actual, luz = false)
            .drawBehind { grano(pincel, 0.14f, BlendMode.SrcOver) }.padding(16.dp)) {
            Etiqueta("Meta de ${Formato.mesDe(Instant.now())}", suave)
            Spacer(Modifier.weight(1f).heightIn(min = 4.dp))
            Row(horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically,
                modifier = Modifier.semantics(mergeDescendants = true) {
                    contentDescription = if (m.objetivo > 0) "${Formato.porcentaje(Math.round(pct * 100).toDouble())} de la meta" else "Sin meta"
                }) {
                val color = paleta.meta
                androidx.compose.foundation.Canvas(Modifier.size(56.dp)) {
                    val g = 7.dp.toPx()
                    val tam = Size(size.width - g, size.height - g)
                    val o = Offset(g / 2, g / 2)
                    drawArc(pista, 0f, 360f, false, o, tam, style = Stroke(g))
                    drawArc(color, -90f, (360 * minOf(1.0, pct)).toFloat(), false, o, tam, style = Stroke(g, cap = StrokeCap.Round))
                }
                Grande(if (m.objetivo > 0) Formato.porcentaje(Math.round(pct * 100).toDouble()) else "—", hex(0xFBF7F2))
            }
            Spacer(Modifier.weight(1f).heightIn(min = 4.dp))
            Texto(if (m.objetivo > 0) "${Formato.compacto(m.actual)} de ${Formato.compacto(m.objetivo)}" else "Tocá para ponerla",
                Estilo(12f, Peso.regular), suave, lineas = 1, numeros = true, escalaMinima = 0.8f)
        }
    }
}

/** Hasta cuatro partes: las tres primeras y el resto junto en «Otros». */
fun partesCanales(canales: List<Tablero.Canal>): List<Pair<String, Double>> {
    val orden = canales.sortedByDescending { it.parte }
    if (orden.size <= 4) return orden.map { Marcas.canalCorto(it.canal) to it.parte }
    return orden.take(3).map { Marcas.canalCorto(it.canal) to it.parte } + ("Otros" to orden.drop(3).sumOf { it.parte })
}

@Composable
private fun TarjetaCanales(canales: List<Tablero.Canal>, paleta: PaletaTienda, accion: () -> Unit) {
    val partes = partesCanales(canales)
    val colores = listOf(paleta.acento, paleta.medio, hex(0xE3A93B), hex(0xD8C9BB))
    val textos = partes.map { "${it.first} ${Formato.porcentaje(Math.round(it.second * 100).toDouble())}" }
    Vistazo(accion, Modifier.semantics(mergeDescendants = true) { contentDescription = "Por dónde compran: ${textos.joinToString(", ")}" }) {
        Etiqueta("Por dónde compran")
        Spacer(Modifier.weight(1f).heightIn(min = 4.dp))
        BoxWithConstraints(Modifier.fillMaxWidth().height(12.dp).clip(Capsula)) {
            Row(horizontalArrangement = Arrangement.spacedBy(2.dp)) {
                partes.forEachIndexed { i, p ->
                    Box(Modifier.width(maxOf(3.dp, (this@BoxWithConstraints.maxWidth - 2.dp * (partes.size - 1)) * p.second.toFloat())).fillMaxHeight()
                        .background(colores[i % colores.size]))
                }
            }
        }
        Spacer(Modifier.weight(1f).heightIn(min = 4.dp))
        Texto(listOf(textos.take(2).joinToString(" · "), textos.drop(2).joinToString(" · ")).filter { it.isNotEmpty() }.joinToString("\n"),
            Estilo(12f, Peso.regular), C.texto2, lineas = 2, numeros = true, escalaMinima = 0.8f, interlineado = 2f)
    }
}

@Composable
private fun TarjetaVisitas(v: Tablero.Visitas, periodo: String, paleta: PaletaTienda, accion: () -> Unit) {
    val n = v.visitantes.actual.toInt()
    val cuando = if (periodo == "hoy") "hoy" else if (periodo == "7d") "en 7 días" else "en 30 días"
    Vistazo(accion) {
        Etiqueta("Tu tienda online")
        Spacer(Modifier.weight(1f).heightIn(min = 4.dp))
        if (v.desde == null) {
            Texto(Formato.sinViudas("Todavía no tuvo visitas"), Estilo(17f, Peso.bold), C.texto)
            Spacer(Modifier.weight(1f).heightIn(min = 4.dp))
            Chip("Compartila", paleta.claro, paleta.tinta)
        } else {
            Grande(Formato.numero(n))
            Texto("${if (n == 1) "visita" else "visitas"} $cuando", Estilo.footnote, C.texto2)
            Spacer(Modifier.weight(1f).heightIn(min = 4.dp))
            val d = v.visitantes.variacion
            if (d != null) {
                val r = abs(d).roundToInt()
                Chip("${if (d >= 0) "▲" else "▼"} ${Formato.porcentaje(r.toDouble())}", if (d >= 0) C.okFondo else C.errorFondo, if (d >= 0) C.okTinta else C.errorTinta)
            }
        }
    }
}

/** El punto verde que late («En vivo»). */
@Composable
fun Latido() {
    val anima = animacionesActivas()
    val latido = if (anima) rememberInfiniteTransition(label = "latido").animateFloat(0f, 1f, infiniteRepeatable(tween(900), RepeatMode.Reverse), label = "l").value else 0f
    Box(Modifier.size(8.dp).scale(0.85f + 0.5f * latido).alpha(1f - 0.4f * latido).clip(CircleShape).background(C.ok))
}

// MARK: - Últimos pedidos y lo más vendido

@Composable
fun UltimosPedidos(t: Tablero, paleta: PaletaTienda) {
    val nav = LocalNavegacion.current
    Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
        TituloInicio("Últimos pedidos", accion = "Ver todos", tinta = paleta.tinta) { nav.filtroPedidos = null; nav.tocar(Pestana.pedidos) }
        Column(Modifier.fillMaxWidth().relieve()) {
            t.ultimosPedidos.take(5).forEachIndexed { i, p ->
                if (i > 0) Separador(68.dp)
                Row(Modifier.fillMaxWidth().heightIn(min = 68.dp)
                    .clickable { nav.pestana = Pestana.pedidos; nav.poner(Pestana.pedidos, listOf(Destino.pedido(p.numero))) }
                    .semantics(mergeDescendants = true) {}
                    .padding(horizontal = 16.dp, vertical = 12.dp), horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
                    AvatarIniciales(p.cliente, 40.dp, fondo = paleta.claro, tinta = paleta.tinta)
                    Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                        Texto(p.cliente, Estilo.fila, C.texto, lineas = 1)
                        Texto("#${p.numero} · ${Marcas.canalCorto(p.canal)} · ${Formato.hace(p.creado)}", Estilo.footnote, C.texto2, lineas = 1, numeros = true)
                    }
                    Column(horizontalAlignment = Alignment.End, verticalArrangement = Arrangement.spacedBy(4.dp)) {
                        p.total?.let { Texto(Formato.pesos(it), Estilo(15f, Peso.bold), C.texto, numeros = true) }
                        ChipEstado(p.estado)
                    }
                }
            }
        }
    }
}

@Composable
fun LoMasVendido(t: Tablero) {
    val nav = LocalNavegacion.current
    val m = margen
    Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
        TituloInicio("Lo más vendido", if (t.periodo == "hoy") "Hoy" else if (t.periodo == "7d") "En 7 días" else "En 30 días")
        Row(Modifier.fillMaxWidth().sinMargen().horizontalScroll(rememberScrollState()).padding(horizontal = m).padding(bottom = 24.dp),
            horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.Top) {
            t.masVendidos.forEachIndexed { i, p ->
                Tocable({ nav.ir(Ruta.producto(p.productoId)) }, Modifier.semantics(mergeDescendants = true) {
                    contentDescription = "Puesto ${i + 1}: ${p.nombre}, ${Formato.cantidad(p.unidades, "vendido", "vendidos")}"
                }) {
                    Column(Modifier.width(148.dp).relieve()) {
                        Box(Modifier.size(148.dp, 112.dp)) {
                            Foto(p.foto, Modifier.matchParentSize(), 450)
                            Box(Modifier.padding(8.dp).height(24.dp).clip(Capsula).background(hex(0xFFFDF9, 0.92f)).padding(horizontal = 8.dp), contentAlignment = Alignment.Center) {
                                Texto("#${i + 1}", Estilo(12f, Peso.bold), hex(0x2A140C))
                            }
                        }
                        Column(Modifier.padding(start = 12.dp, end = 12.dp, top = 10.dp, bottom = 12.dp), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                            // Dos renglones siempre (lineLimit(2, reservesSpace:) en iOS): las tarjetas quedan parejas.
                            Box(Modifier.heightIn(min = 36.dp)) { Texto(p.nombre, Estilo(14f, Peso.semibold), C.texto, lineas = 2) }
                            Texto(Formato.cantidad(p.unidades, "vendido", "vendidos"), Estilo.footnote, C.texto2, numeros = true)
                        }
                    }
                }
            }
        }
    }
}

// MARK: - Al final

/** «Más vistos»: los productos que más miraron en la tienda online. */
@Composable
fun TarjetaMasVistos(visitas: Tablero.Visitas, paleta: PaletaTienda) {
    Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
        TituloInicio("Más vistos", "En la tienda online")
        val max = (visitas.masVistos.maxOfOrNull { it.vistas } ?: 1).coerceAtLeast(1).toFloat()
        Column(Modifier.fillMaxWidth().relieve()) {
            visitas.masVistos.take(5).forEachIndexed { i, p ->
                if (i > 0) Separador(70.dp)
                Row(Modifier.heightIn(min = 68.dp).semantics(mergeDescendants = true) {}.padding(horizontal = 16.dp, vertical = 12.dp),
                    horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
                    Miniatura(p.foto, 42.dp, radio = 12.dp)
                    Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(6.dp)) {
                        Texto(p.nombre, Estilo.fila, C.texto, lineas = 1)
                        BoxWithConstraints(Modifier.fillMaxWidth().height(5.dp)) {
                            Box(Modifier.width(maxWidth * (p.vistas / max)).fillMaxHeight().clip(Capsula).background(paleta.acento.copy(alpha = 0.55f)))
                        }
                    }
                    Texto(Formato.numero(p.vistas), Estilo.fila, C.texto2, numeros = true)
                }
            }
        }
    }
}

/** «Reseñas nuevas»: las últimas sin responder. */
@Composable
fun TarjetaResenas(resenas: Tablero.Resenas, paleta: PaletaTienda) {
    val nav = LocalNavegacion.current
    Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
        TituloInicio("Reseñas nuevas", accion = "Responder", tinta = paleta.tinta) { nav.inicio.add(Destino.resenas) }
        Column(Modifier.fillMaxWidth().relieve()) {
            resenas.promedio?.let { p ->
                Row(Modifier.padding(horizontal = 16.dp, vertical = 14.dp), horizontalArrangement = Arrangement.spacedBy(6.dp), verticalAlignment = Alignment.CenterVertically) {
                    Texto(String.format(Formato.argentina, "%.1f", p), Estilo.headline, C.texto)
                    Estrellas(Math.round(p).toInt())
                    Texto("promedio", Estilo.footnote, C.texto2)
                }
            }
            resenas.nuevas.take(3).forEachIndexed { i, r ->
                if (i > 0 || resenas.promedio != null) Separador(16.dp)
                Column(Modifier.padding(horizontal = 16.dp, vertical = 12.dp), verticalArrangement = Arrangement.spacedBy(4.dp)) {
                    Row(horizontalArrangement = Arrangement.spacedBy(6.dp), verticalAlignment = Alignment.CenterVertically) {
                        Estrellas(r["estrellas"]?.entero ?: 0)
                        Texto(r["autor"]?.texto ?: "", Estilo.footnoteFuerte, C.texto)
                        Spacer(Modifier.weight(1f))
                        Texto(r["producto"]?.texto ?: "", Estilo.footnote, C.texto2, lineas = 1)
                    }
                    Texto("«${r["texto"]?.texto ?: ""}»", Estilo.texto14, C.texto, lineas = 3)
                }
            }
        }
    }
}

/** «Envíos de hoy»: paquetes y retiros de los transportes. */
@Composable
fun TarjetaEnvios(envios: Tablero.Envios, paleta: PaletaTienda) {
    val nav = LocalNavegacion.current
    Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
        TituloInicio("Envíos de hoy", accion = "Preparar", tinta = paleta.tinta) { nav.filtroPedidos = "Para preparar"; nav.pestana = Pestana.pedidos }
        Column(Modifier.fillMaxWidth().relieve()) {
            Row(Modifier.padding(horizontal = 16.dp, vertical = 12.dp), horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
                Box(Modifier.size(40.dp).clip(RoundedCornerShape(12.dp)).background(paleta.claro), contentAlignment = Alignment.Center) { Icono("shippingbox.fill", 16f, paleta.tinta) }
                Texto(Formato.cantidad(envios.paquetes, "paquete para despachar", "paquetes para despachar"), Estilo.fila, C.texto, Modifier.weight(1f))
            }
            envios.retiros.forEach { r ->
                Separador(16.dp)
                Row(Modifier.padding(horizontal = 16.dp, vertical = 12.dp), horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.CenterVertically) {
                    Texto(r.hora, Estilo.fila, if (r.proximo) paleta.tinta else C.texto, Modifier.width(54.dp), numeros = true)
                    Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(1.dp)) {
                        Texto(Formato.mayuscula(r.transporte), Estilo.texto14, C.texto)
                        r.ciudad?.let { Texto(it, Estilo.footnote, C.texto2) }
                    }
                    Texto(Formato.cantidad(r.paquetes, "paquete", "paquetes"), Estilo.footnote, C.texto2)
                    if (r.proximo) ChipSimple("Próximo", C.avisoFondo, C.avisoTinta)
                }
            }
        }
    }
}

/** «Integraciones»: lo conectado (Mercado Pago, Mercado Libre, Correo) y un acceso para sumar más. */
@Composable
fun TarjetaIntegraciones(paleta: PaletaTienda) {
    val estado = LocalEstado.current
    val nav = LocalNavegacion.current
    var conectadas by remember { mutableStateOf<List<Pair<String, String>>?>(null) }

    @Serializable data class I(val conectada: Boolean? = null, val usuario: String? = null, val apodo: String? = null, val email: String? = null)

    LaunchedEffect(Unit) {
        if (conectadas != null) return@LaunchedEffect
        val c = mutableListOf<Pair<String, String>>()
        for ((id, nombre) in listOf("mercadopago" to "Mercado Pago", "mercadolibre" to "Mercado Libre", "correo" to "Correo Argentino")) {
            val i = estado.api.opcional<I>("/v1/panel/integraciones/$id")
            if (i?.conectada == true) c.add(nombre to (i.usuario ?: i.apodo ?: i.email ?: "Conectada"))
        }
        conectadas = c
    }

    Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
        TituloInicio("Integraciones", accion = "Explorar", tinta = paleta.tinta) { nav.inicio.add(Destino.integraciones) }
        Column(Modifier.fillMaxWidth().relieve()) {
            val c = conectadas
            when {
                c != null && c.isNotEmpty() -> c.forEachIndexed { i, x ->
                    if (i > 0) Separador(16.dp)
                    Row(Modifier.padding(horizontal = 16.dp, vertical = 14.dp), horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.CenterVertically) {
                        Texto(x.first, Estilo.fila, C.texto, Modifier.weight(1f))
                        Texto(x.second, Estilo.footnote, C.texto2, lineas = 1)
                        Box(Modifier.size(8.dp).clip(CircleShape).background(C.ok))
                    }
                }
                c != null -> Texto(Formato.sinViudas("Conectá Mercado Pago, Mercado Libre y Correo Argentino para cobrar, publicar y enviar."), Estilo.texto14, C.texto2,
                    Modifier.fillMaxWidth().clickable { nav.inicio.add(Destino.integraciones) }.padding(16.dp))
                else -> Box(Modifier.padding(16.dp)) { Rueda(paleta.acento) }
            }
        }
    }
}

// MARK: - Cargando (2c)

/** El esqueleto del Inicio, con la misma forma: la tarjeta de Tincho, el bento y una lista. */
@Composable
fun EsqueletoInicio(paleta: PaletaTienda) {
    Column(Modifier.semantics { contentDescription = "Cargando cómo viene el día" }) {
        Row(Modifier.subir(28.dp).padding(horizontal = 16.dp).fillMaxWidth().relieve(fondo = tema(0xFFFDF9, 0x2E1B12).actual, fuerte = true).padding(14.dp),
            horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.Top) {
            Esqueleto(48.dp, 48.dp, 24.dp)
            Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                Esqueleto(null, 14.dp)
                Esqueleto(180.dp, 14.dp, suave = true)
                Esqueleto(120.dp, 36.dp, 18.dp)
            }
        }
        Column(Modifier.padding(top = 24.dp).margenPantalla(), verticalArrangement = Arrangement.spacedBy(32.dp)) {
            Row(Modifier.fillMaxWidth().height(252.dp), horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                Box(Modifier.weight(1f).fillMaxHeight().relieve().padding(12.dp)) { Esqueleto(null, 150.dp, 16.dp, suave = true) }
                Column(Modifier.weight(1f).fillMaxHeight(), verticalArrangement = Arrangement.spacedBy(12.dp)) {
                    Box(Modifier.weight(1f).fillMaxWidth().relieve())
                    Box(Modifier.weight(1f).fillMaxWidth().relieve())
                }
            }
            Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
                Esqueleto(150.dp, 22.dp)
                EsqueletoFilas(3, 40.dp)
            }
        }
    }
}

/** Los primeros pasos de Inicio: qué tiendas ya compartieron el link o cerraron la tarjeta (en el teléfono, por tienda). */
object PrimerosPasos {
    const val claveCompartido = "pasos.compartido"
    const val claveOcultos = "pasos.ocultos"

    fun contiene(lista: String, id: String): Boolean = id.isNotEmpty() && id in lista.split(",")

    fun sumar(lista: String, id: String): String {
        if (id.isEmpty() || contiene(lista, id)) return lista
        return if (lista.isEmpty()) id else "$lista,$id"
    }

    fun ocultos(id: String?): Boolean = contiene(Preferencias.texto(claveOcultos) ?: "", id ?: "")
}
