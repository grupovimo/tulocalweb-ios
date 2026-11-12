package ar.com.tulocalweb.app.pantallas.inicio

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.selection.selectable
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.runtime.Composable
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.drawBehind
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.BlendMode
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.layout.Layout
import androidx.compose.ui.layout.layout
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.app.Destino
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.app.LocalNavegacion
import ar.com.tulocalweb.app.app.Pestana
import ar.com.tulocalweb.app.nucleo.Acceso
import ar.com.tulocalweb.app.nucleo.Comparado
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.Tablero
import ar.com.tulocalweb.app.nucleo.Tincho
import ar.com.tulocalweb.app.tema.*
import kotlinx.coroutines.launch
import java.time.Instant
import kotlin.math.abs
import kotlin.math.max
import kotlin.math.roundToInt

// Clon de Pantallas/Inicio/Vidriera.swift. La vidriera del Inicio (AFinal, ACero y AOscuro de docs/handoff/inicio-a):
// el color de la tienda con grano y dos halos, el logo, el estado, buscar y avisos en vidrio, el período, lo vendido con
// el gráfico de hoy contra ayer y «Vender ahora». Al bajar se pliega a BarraVidriera (AScroll). La tarjeta de Tincho va
// montada sobre su borde.

/** Lo que la vidriera calcula del tablero (los `static` de Vidriera en iOS). */
object Vidriera {
    /** El día en $0 (o sin pedidos para quien no ve montos): la vidriera invita a compartir (ACero). */
    fun enCero(t: Tablero): Boolean = if (t.montos) (t.ventas?.actual ?: 0.0) == 0.0 else t.kpis.pedidos.actual == 0.0

    /** Lo vendido; quien no ve montos ve los pedidos. */
    fun monto(t: Tablero): String {
        val v = t.ventas
        if (t.montos && v != null) return Formato.pesos(v.actual)
        return Formato.cantidad(t.kpis.pedidos.actual.toInt(), "pedido", "pedidos")
    }

    /** «Abierta · 3 mirando ahora» si hay gente en vivo; si no, las visitas de hoy. */
    fun estadoTienda(t: Tablero?): String {
        if (t == null) return "Abierta"
        val vivo = t.enVivo
        if (vivo != null && vivo.personas > 0) return "Abierta · ${Formato.numero(vivo.personas)} mirando ahora"
        val v = t.visitas
        if (v == null || v.desde == null) return "Abierta"
        val n = v.hoy.actual.toInt()
        if (n == 0) return "Abierta · todavía sin visitas hoy"
        return "Abierta · ${if (n == 1) "1 visita" else "${Formato.numero(n)} visitas"} hoy"
    }

    /** «8 h · 12 h · 16 h · 20 h» (hoy) o «5/11 · 7/11 · …»: cuatro marcas parejas del período. */
    fun marcas(t: Tablero, porHora: Boolean): String {
        val total = t.hasta.toEpochMilli() - t.desde.toEpochMilli()
        return (0 until 4).joinToString(" · ") { i ->
            val f = Formato.enArgentina(Instant.ofEpochMilli(t.desde.toEpochMilli() + (total * (i * 2 + 1) / 8)))
            if (porHora) "${f.hour}${Formato.fino}h" else "${f.dayOfMonth}/${f.monthValue}"
        }
    }

    fun cuando(periodo: String) = if (periodo == "hoy") "Hoy" else if (periodo == "7d") "En 7 días" else "En 30 días"
}

/** La vidriera: arriba de todo en el Inicio, con lo que haya (cargando, error o el tablero). */
@Composable
fun Vidriera(carga: Carga<Tablero>, periodo: String, alElegir: (String) -> Unit, sinLeer: Int, paleta: PaletaTienda, se: Boolean, modifier: Modifier = Modifier) {
    val estado = LocalEstado.current
    val t = carga.valor
    val cero = t?.let { Vidriera.enCero(it) } ?: false
    val conTarjeta = (t != null && !cero) || carga == Carga.cargando
    val arriba = WindowInsets.statusBars.asPaddingValues().calculateTopPadding()
    val nombre = (t?.hola ?: estado.datos?.nombreDePila ?: "").trim()
    val hola = if (nombre.isEmpty()) "Hola" else "Hola, $nombre"
    Column(
        modifier.fillMaxWidth().fondoVidriera(paleta, pincelGrano(), arriba = arriba)
            .padding(start = margen, end = margen, top = arriba + 12.dp, bottom = if (conTarjeta) 48.dp else 24.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp),
    ) {
        Arriba(t, sinLeer, paleta)
        SelectorPeriodo(periodo, alElegir, paleta)
        when (carga) {
            Carga.cargando -> EsqueletoVidriera(paleta, se)
            is Carga.error -> {}
            is Carga.listo -> if (cero) EnCero(carga.v, periodo, hola, paleta, se) else ConVentas(carga.v, periodo, hola, paleta, se)
        }
    }
}

// MARK: Arriba

@Composable
private fun Arriba(t: Tablero?, sinLeer: Int, paleta: PaletaTienda) {
    val estado = LocalEstado.current
    val nav = LocalNavegacion.current
    Row(horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
        Tocable({ nav.tocar(Pestana.mas) }, descripcion = "Tu tienda, ${estado.tienda?.nombre ?: ""}") { LogoTienda(estado.tienda, 48.dp, 16.dp, paleta) }
        Column(Modifier.weight(1f).semantics(mergeDescendants = true) {}, verticalArrangement = Arrangement.spacedBy(2.dp)) {
            Texto(estado.tienda?.nombre ?: "Tu tienda", Estilo(22f, Peso.bold, -0.018f), paleta.sobre, lineas = 1, escalaMinima = 0.75f)
            Row(horizontalArrangement = Arrangement.spacedBy(6.dp), verticalAlignment = Alignment.CenterVertically) {
                Box(Modifier.size(8.dp).drawBehind { drawCircle(hex(0x57D9A0, 0.28f), 7.dp.toPx()) }.clip(CircleShape).background(hex(0x57D9A0)))
                Texto(Vidriera.estadoTienda(t), Estilo(13f, Peso.medium), paleta.sobre2, lineas = 1, numeros = true, escalaMinima = 0.85f)
            }
        }
        BotonVidrio("magnifyingglass", "Buscar", paleta = paleta) { nav.buscar = true }
        BotonVidrio("bell", if (sinLeer > 0) "Avisos, $sinLeer sin leer" else "Avisos", sinLeer > 0, paleta) { nav.inicio.add(Destino.avisos) }
    }
}

// MARK: Con ventas

@Composable
private fun ConVentas(t: Tablero, periodo: String, hola: String, paleta: PaletaTienda, se: Boolean) {
    val nav = LocalNavegacion.current
    Column(Modifier.semantics(mergeDescendants = true) {}, verticalArrangement = Arrangement.spacedBy(6.dp)) {
        Texto("$hola. ${Vidriera.cuando(periodo)} ${if (t.montos) "vendiste" else "tuviste"}", Estilo(15f, Peso.medium), paleta.sobre2)
        NumeroAnimado(Vidriera.monto(t), Estilo(if (se) 40f else 46f, Peso.bold, -0.039f), paleta.sobre, escalaMinima = 0.6f)
        LineaDelta(t, periodo, paleta, Modifier.padding(top = 2.dp))
    }
    val s = t.serie
    val v = t.ventas
    if (t.montos && s != null && v != null) Grafico(t, s, v, periodo, paleta, se)
    BotonVidriera("Vender ahora", "barcode.viewfinder", paleta) { nav.venta = true }
}

private fun contra(periodo: String) = if (periodo == "hoy") "vs. ayer" else if (periodo == "7d") "vs. 7 días antes" else "vs. 30 días antes"

@OptIn(ExperimentalLayoutApi::class)
@Composable
private fun LineaDelta(t: Tablero, periodo: String, paleta: PaletaTienda, modifier: Modifier = Modifier) {
    val c = if (t.montos) t.ventas else t.kpis.pedidos
    val pedidos = Formato.cantidad(t.kpis.pedidos.actual.toInt(), "pedido", "pedidos")
    val tk = t.kpis.ticket
    val detalle = if (t.montos && tk != null && tk.actual > 0) "$pedidos · ticket ${Formato.pesos(tk.actual)}"
        else if (t.montos) pedidos
        else Formato.cantidad(t.kpis.clientesNuevos.actual.toInt(), "cliente nuevo", "clientes nuevos")
    val estilo = Estilo(14f, Peso.medium)
    FlowRow(modifier, horizontalArrangement = Arrangement.spacedBy(8.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
        Pildora(c, periodo, paleta)
        Texto(detalle, estilo, paleta.sobre2, lineas = 1, numeros = true)
    }
}

@Composable
private fun Pildora(c: Comparado?, periodo: String, paleta: PaletaTienda) {
    val d = c?.variacion
    val texto = when {
        d != null -> "${if (d >= 0) "▲" else "▼"} ${Formato.porcentaje(abs(d).roundToInt().toDouble())} ${contra(periodo)}"
        c != null && c.actual > 0 -> "Primeras ventas"
        else -> return
    }
    val etiqueta = if (d != null) "${if (d >= 0) "Subió" else "Bajó"} ${Formato.porcentaje(abs(d).roundToInt().toDouble())} ${contra(periodo)}" else texto
    Box(Modifier.vidrio(Capsula, paleta).padding(horizontal = 9.dp, vertical = 3.dp).semantics { contentDescription = etiqueta }) {
        Texto(texto, Estilo(14f, Peso.bold), paleta.sobre, lineas = 1, numeros = true)
    }
}

@Composable
private fun Grafico(t: Tablero, s: Tablero.Serie, v: Comparado, periodo: String, paleta: PaletaTienda, se: Boolean) {
    val hoy = periodo == "hoy"
    val colores = ColoresGrafico(paleta.linea, paleta.sobreSuave, paleta.velo, paleta.borde, paleta.linea.copy(alpha = 0.25f))
    val estilo = Estilo(12f, Peso.medium)
    Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
        Box(Modifier.offset(x = (-2).dp)) {
            GraficoVentas(s.actual, s.anterior, if (se) 64.dp else 88.dp, EtiquetasVentas.etiquetas(t.desde, t.hasta, s.actual.size, hoy), colores,
                "Ventas ${if (hoy) "de hoy por hora" else "del período"} comparadas con ${if (hoy) "ayer" else "el período anterior"}: ${Formato.pesos(v.actual)} contra ${Formato.pesos(v.anterior)}")
        }
        SiEntra({
            Row(horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
                Row(horizontalArrangement = Arrangement.spacedBy(6.dp), verticalAlignment = Alignment.CenterVertically) {
                    Box(Modifier.size(14.dp, 3.dp).clip(Capsula).background(paleta.linea))
                    Texto(if (hoy) "Hoy" else "Este período", estilo, paleta.sobre2, lineas = 1)
                }
                Row(horizontalArrangement = Arrangement.spacedBy(6.dp), verticalAlignment = Alignment.CenterVertically) {
                    Punteado(Modifier.size(14.dp, 2.dp), paleta.sobreSuave)
                    Texto("${if (hoy) "Ayer" else "Antes"} ${Formato.pesos(v.anterior)}", estilo, paleta.sobre2, lineas = 1, numeros = true)
                }
            }
        }, { Texto(Vidriera.marcas(t, hoy), estilo, paleta.sobre2, lineas = 1, numeros = true) })
    }
}

/** Lo principal a la izquierda y lo extra a la derecha si entra en el renglón; si no, sólo lo principal
 *  (ViewThatFits en iOS). */
@Composable
private fun SiEntra(principal: @Composable () -> Unit, extra: @Composable () -> Unit) {
    Layout(contents = listOf(principal, extra), modifier = Modifier.fillMaxWidth()) { (a, b), c ->
        val libre = c.copy(minWidth = 0)
        val pa = a.first().measure(libre)
        val pb = b.first().measure(libre)
        val entra = pa.width + 8.dp.roundToPx() + pb.width <= c.maxWidth
        val alto = if (entra) max(pa.height, pb.height) else pa.height
        layout(c.maxWidth, alto) {
            pa.place(0, (alto - pa.height) / 2)
            if (entra) pb.place(c.maxWidth - pb.width, (alto - pb.height) / 2)
        }
    }
}

// MARK: En cero (ACero)

@Composable
private fun EnCero(t: Tablero, periodo: String, hola: String, paleta: PaletaTienda, se: Boolean) {
    val nav = LocalNavegacion.current
    Column(Modifier.semantics(mergeDescendants = true) {}, verticalArrangement = Arrangement.spacedBy(4.dp)) {
        Texto(hola, Estilo(15f, Peso.medium), paleta.sobre2)
        Texto(Formato.sinViudas(titularCero(t, periodo)), Estilo(if (se) 26f else 30f, Peso.bold, -0.027f), paleta.sobre)
    }
    BotonVidriera("Compartir tu vidriera", "square.and.arrow.up", paleta) { nav.inicio.add(Destino.compartir) }
}

private fun titularCero(t: Tablero, periodo: String): String {
    val v = t.visitas
    if (v == null || v.desde == null) return "Tu local está abierto. Compartilo para recibir las primeras visitas."
    val n = (if (periodo == "hoy") v.hoy.actual else v.visitantes.actual).toInt()
    val cuando = Vidriera.cuando(periodo)
    if (n == 0) return "Tu local está abierto. $cuando todavía no te visitaron: compartilo."
    return "Tu local está abierto. $cuando te ${if (n == 1) "visitó 1 persona" else "visitaron ${Formato.numero(n)} personas"}."
}

// MARK: Cargando

@Composable
private fun EsqueletoVidriera(paleta: PaletaTienda, se: Boolean) {
    Column(Modifier.semantics { contentDescription = "Cargando" }, verticalArrangement = Arrangement.spacedBy(16.dp)) {
        Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
            Bloque(paleta, 150.dp, 15.dp)
            Bloque(paleta, 230.dp, 46.dp, 12.dp)
            Bloque(paleta, 190.dp, 22.dp)
        }
        Bloque(paleta, null, if (se) 64.dp else 88.dp, 14.dp)
        Bloque(paleta, null, 60.dp, 20.dp)
    }
}

@Composable
private fun Bloque(paleta: PaletaTienda, ancho: Dp?, alto: Dp, radio: Dp? = null) {
    Box((if (ancho != null) Modifier.width(ancho) else Modifier.fillMaxWidth()).height(alto).clip(RoundedCornerShape(radio ?: (alto / 2))).background(paleta.velo))
}

// MARK: - Fondo

/** El fondo de la vidriera: el color (en oscuro, un degradé hacia el fondo), dos halos y el grano, redondeado abajo.
 *  `arriba`: la barra de estado, que la vidriera cubre. */
fun Modifier.fondoVidriera(paleta: PaletaTienda, pincel: Brush, radio: Dp = 32.dp, halos: Boolean = true, arriba: Dp = 0.dp): Modifier =
    this.clip(RoundedCornerShape(bottomStart = radio, bottomEnd = radio)).drawBehind {
        drawRect(Brush.linearGradient(listOf(paleta.fondoArriba, paleta.fondoAbajo), Offset(size.width * 0.32f, 0f), Offset(size.width * 0.68f, size.height)))
        if (halos) {
            val r1 = 124.dp.toPx()
            val c1 = Offset(size.width - 52.dp.toPx(), arriba.toPx() + 24.dp.toPx())
            drawCircle(Brush.radialGradient(listOf(paleta.luz, paleta.luz.copy(alpha = 0f)), c1, r1), r1, c1)
            val r2 = 112.dp.toPx()
            val c2 = Offset(48.dp.toPx(), size.height)
            drawCircle(Brush.radialGradient(listOf(paleta.sombraHalo, paleta.sombraHalo.copy(alpha = 0f)), c2, r2), r2, c2)
        }
        grano(pincel, if (paleta.oscuro) 0.18f else 0.2f)
    }

// MARK: - Piezas de la vidriera

/** El logo de la tienda (Tienda.logo, entero sobre blanco) o su inicial; en oscuro, la inicial va en el acento. */
@Composable
fun LogoTienda(tienda: Acceso.Tienda?, tam: Dp, radio: Dp, paleta: PaletaTienda) {
    val forma = RoundedCornerShape(radio)
    val u = urlDeFoto(tienda?.logo, (tam.value * 3).toInt())
    val inicial = @Composable {
        Box(Modifier.fillMaxSize().background(if (paleta.oscuro) paleta.acento else Color.White), contentAlignment = Alignment.Center) {
            Texto(Formato.iniciales(tienda?.nombre ?: "").take(1), Estilo(tam.value * 0.46f, Peso.bold), if (paleta.oscuro) paleta.sobreAcento else paleta.tintaBlanco)
        }
    }
    Box(Modifier.size(tam).sombra(forma, Color.Black.copy(alpha = 0.3f), 8.dp, y = 6.dp).clip(forma).background(Color.White)
        .border(1.dp, Brush.verticalGradient(0f to Color.White.copy(alpha = 0.6f), 0.5f to Color.Transparent), forma), contentAlignment = Alignment.Center) {
        if (u != null) {
            coil3.compose.SubcomposeAsyncImage(u, null, Modifier.fillMaxSize().padding(tam * 0.1f), contentScale = ContentScale.Fit,
                error = { inicial() }, loading = { inicial() })
        } else inicial()
    }
}

/** Botón redondo de 44 en vidrio (buscar, avisos), con el punto de «sin leer». */
@Composable
fun BotonVidrio(simbolo: String, etiqueta: String, punto: Boolean = false, paleta: PaletaTienda, accion: () -> Unit) {
    Tocable(accion, Modifier.size(44.dp), descripcion = etiqueta) {
        Box(Modifier.size(44.dp)) {
            Box(Modifier.fillMaxSize().vidrio(CircleShape, paleta), contentAlignment = Alignment.Center) { Icono(simbolo, 18f, paleta.sobre) }
            if (punto) {
                Box(Modifier.align(Alignment.TopEnd).offset(x = (-8).dp, y = 7.dp).size(13.dp).clip(CircleShape).background(paleta.fondoArriba).padding(2.dp)
                    .clip(CircleShape).background(if (paleta.oscuro) paleta.tinta else hex(0xFFC24D)))
            }
        }
    }
}

/** Hoy / 7 días / 30 días en vidrio (reemplaza al Segmentado en el Inicio). 44 de alto con el vidrio. */
@Composable
fun SelectorPeriodo(periodo: String, alElegir: (String) -> Unit, paleta: PaletaTienda) {
    val forma = RoundedCornerShape(11.dp)
    Row(Modifier.vidrio(RoundedCornerShape(14.dp), paleta).padding(3.dp).semantics { contentDescription = "Período" },
        horizontalArrangement = Arrangement.spacedBy(2.dp)) {
        listOf("hoy" to "Hoy", "7d" to "7 días", "30d" to "30 días").forEach { (valor, nombre) ->
            val activo = valor == periodo
            Box(Modifier.height(38.dp)
                .then(if (activo) Modifier.sombra(forma, Color.Black.copy(alpha = 0.25f), 3.dp, y = 2.dp).clip(forma).background(Color.White) else Modifier)
                .selectable(activo, role = Role.RadioButton) { alElegir(valor); Haptico.seleccion() }
                .padding(horizontal = 14.dp), contentAlignment = Alignment.Center) {
                Texto(nombre, Estilo(14f, if (activo) Peso.bold else Peso.semibold), if (activo) paleta.tintaBlanco else paleta.sobre, lineas = 1)
            }
        }
    }
}

/** El botón grande de la vidriera: blanco con un degradé sutil, relieve, grano y sombra; en oscuro, el acento con su
 *  texto. */
@Composable
fun BotonVidriera(texto: String, simbolo: String, paleta: PaletaTienda, alto: Dp = 60.dp, accion: () -> Unit) {
    val forma = RoundedCornerShape(20.dp)
    val pincel = pincelGrano()
    val tinta = if (paleta.oscuro) paleta.sobreAcento else paleta.tintaBlanco
    Tocable(accion, Modifier.fillMaxWidth().height(alto)) {
        Row(Modifier.fillMaxSize().sombra(forma, Color.Black.copy(alpha = if (paleta.oscuro) 0.4f else 0.32f), 14.dp, y = 12.dp).clip(forma)
            .background(if (paleta.oscuro) SolidColor(paleta.acento) else Brush.verticalGradient(listOf(Color.White, hex(0xF6EFE7))))
            .drawBehind { grano(pincel, 0.08f, BlendMode.Multiply) }
            .border(1.dp, Brush.verticalGradient(0f to Color.White.copy(alpha = if (paleta.oscuro) 0.2f else 1f), 0.2f to Color.Transparent), forma)
            .border(2.dp, Brush.verticalGradient(0.8f to Color.Transparent, 1f to hex(0x2A140C, 0.08f)), forma),
            horizontalArrangement = Arrangement.spacedBy(10.dp, Alignment.CenterHorizontally), verticalAlignment = Alignment.CenterVertically) {
            Icono(simbolo, 20f, tinta)
            Texto(texto, Estilo(18f, Peso.bold, -0.011f), tinta, lineas = 1, escalaMinima = 0.8f)
        }
    }
}

/** Línea punteada de la leyenda del gráfico. */
@Composable
fun Punteado(modifier: Modifier = Modifier, color: Color = hex(0x8A7364)) {
    androidx.compose.foundation.Canvas(modifier) {
        drawLine(color, Offset(0f, size.height / 2), Offset(size.width, size.height / 2), 2.dp.toPx(),
            pathEffect = androidx.compose.ui.graphics.PathEffect.dashPathEffect(floatArrayOf(3.dp.toPx(), 2.dp.toPx())))
    }
}

/** Sube un bloque `n` sin dejar el hueco (padding negativo en iOS): la tarjeta montada sobre el borde de la vidriera. */
fun Modifier.subir(n: Dp): Modifier = this.layout { medible, restricciones ->
    val p = medible.measure(restricciones)
    val d = n.roundToPx()
    layout(p.width, max(0, p.height - d)) { p.place(0, -d) }
}

// MARK: - Plegada (AScroll)

/** La vidriera plegada al bajar: el logo, lo vendido y un «Vender» compacto. */
@Composable
fun BarraVidriera(t: Tablero, periodo: String, paleta: PaletaTienda) {
    val estado = LocalEstado.current
    val nav = LocalNavegacion.current
    val arriba = WindowInsets.statusBars.asPaddingValues().calculateTopPadding()
    val tinta = if (paleta.oscuro) paleta.sobreAcento else paleta.tintaBlanco
    Row(Modifier.fillMaxWidth().sombra(RoundedCornerShape(bottomStart = 26.dp, bottomEnd = 26.dp), hex(0x2A140C, 0.22f), 12.dp, y = 10.dp)
        .fondoVidriera(paleta, pincelGrano(), 26.dp, halos = false)
        .padding(start = 16.dp, end = 16.dp, top = arriba + 8.dp, bottom = 14.dp),
        horizontalArrangement = Arrangement.spacedBy(10.dp), verticalAlignment = Alignment.CenterVertically) {
        LogoTienda(estado.tienda, 38.dp, 13.dp, paleta)
        Column(Modifier.weight(1f).semantics(mergeDescendants = true) {}) {
            Texto("${estado.tienda?.nombre ?: "Tu tienda"} · ${if (periodo == "hoy") "hoy" else if (periodo == "7d") "7 días" else "30 días"}",
                Estilo(13f, Peso.medium), paleta.sobre2, lineas = 1)
            Texto(Vidriera.monto(t), Estilo(21f, Peso.bold, -0.024f), paleta.sobre, lineas = 1, numeros = true, escalaMinima = 0.7f)
        }
        Tocable({ nav.venta = true }) {
            Row(Modifier.height(44.dp).sombra(Capsula, Color.Black.copy(alpha = 0.2f), 7.dp, y = 6.dp).clip(Capsula)
                .background(if (paleta.oscuro) paleta.acento else Color.White).padding(horizontal = 16.dp),
                horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.CenterVertically) {
                Icono("barcode.viewfinder", 16f, tinta)
                Texto("Vender", Estilo(16f, Peso.bold), tinta, lineas = 1)
            }
        }
    }
}

// MARK: - Tincho

/** Una idea de Tincho para la tarjeta del Inicio: el texto, la acción principal (si hay) y la pregunta con la que
 *  «Preguntale» abre el chat. */
class IdeaTincho(val id: String, val texto: androidx.compose.ui.text.AnnotatedString, val boton: String?, val pregunta: String, val accion: suspend () -> Unit)

/** La tarjeta de Tincho montada sobre el borde de la vidriera (reemplaza a «Preguntale a Tincho» y a las ideas de
 *  abajo). */
@Composable
fun TarjetaTincho(idea: IdeaTincho, paleta: PaletaTienda, modifier: Modifier = Modifier) {
    val nav = LocalNavegacion.current
    val alcance = rememberCoroutineScope()
    Row(modifier.fillMaxWidth().relieve(24.dp, tema(0xFFFDF9, 0x2E1B12).actual, fuerte = true)
        .padding(start = 12.dp, end = 14.dp, top = 14.dp, bottom = 10.dp),
        horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.Top) {
        ImagenTincho(Tincho.avatar, Modifier.size(48.dp).clip(CircleShape).background(tema(0xF3E3D3, 0x4A2A1A).actual), ContentScale.Crop)
        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(6.dp)) {
            TextoRico(idea.texto, Estilo(16f, Peso.regular), C.texto, Modifier.semantics { contentDescription = "Tincho: ${idea.texto.text}" }, interlineado = 4f)
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                idea.boton?.let { b -> BotonChico(b, paleta.acento, paleta.sobreAcento, relieve = true) { alcance.launch { idea.accion() } } }
                BotonChico("Preguntale", if (idea.boton == null) paleta.acento else paleta.claro, if (idea.boton == null) paleta.sobreAcento else paleta.tinta,
                    relieve = idea.boton == null) {
                    nav.preguntaTincho = idea.pregunta
                    nav.pestana = Pestana.tincho
                }
            }
        }
    }
}

/** Botón chico en píldora (36 visibles, 44 para tocar): las acciones de Tincho. */
@Composable
fun BotonChico(texto: String, fondo: Color, tinta: Color, relieve: Boolean = false, accion: () -> Unit) {
    Tocable(accion) {
        Box(Modifier.heightIn(min = 44.dp), contentAlignment = Alignment.Center) {
            Box(Modifier.height(36.dp)
                .then(if (relieve) Modifier.sombra(Capsula, fondo.copy(alpha = 0.5f), 5.dp, y = 4.dp) else Modifier)
                .clip(Capsula).background(fondo)
                .then(if (relieve) Modifier.border(1.dp, Brush.verticalGradient(0f to Color.White.copy(alpha = 0.25f), 0.5f to Color.Transparent), Capsula) else Modifier)
                .padding(horizontal = 14.dp), contentAlignment = Alignment.Center) {
                Texto(texto, Estilo(14f, Peso.bold), tinta, lineas = 1)
            }
        }
    }
}
