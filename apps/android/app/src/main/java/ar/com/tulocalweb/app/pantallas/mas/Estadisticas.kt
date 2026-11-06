package ar.com.tulocalweb.app.pantallas.mas

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.app.Destino
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.app.LocalNavegacion
import ar.com.tulocalweb.app.nucleo.Comparado
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.Marcas
import ar.com.tulocalweb.app.nucleo.Tablero
import ar.com.tulocalweb.app.nucleo.Tincho
import ar.com.tulocalweb.app.nucleo.opcional
import ar.com.tulocalweb.app.tema.*
import kotlinx.serialization.Serializable
import kotlin.math.abs
import kotlin.math.roundToInt

// Clon de Pantallas/Mas/Estadisticas.swift.

@Serializable
private data class Informes(val meses: List<String> = emptyList(), val series: Series = Series(), val kpis: Kpis = Kpis(), val insights: List<String> = emptyList()) {
    @Serializable data class Series(val ventas: List<Int> = emptyList(), val productos: List<Int> = emptyList(), val clientes: List<Int> = emptyList(),
                                    val margen: List<Int?> = emptyList(), val visitas: List<Int?>? = null)
    @Serializable data class Kpis(val ventas: Int = 0, val ventasVs: Double? = null, val ticket: Int = 0, val ticketVs: Double? = null, val margen: Int? = null,
                                  val ahorro: Int = 0, val mesAnterior: String = "")
}

private data class DatosEstadisticas(val tablero: Tablero, val informes: Informes?)

private fun variacion(c: Comparado?): String? {
    val v = c?.variacion ?: return null
    val r = v.roundToInt()
    return if (r == 0) "Igual que el período anterior" else "${if (r > 0) "▲" else "▼"} ${abs(r)} % contra el período anterior"
}

/** Estadísticas (Informes del panel web): el período elegido (ventas, pedidos, ticket, clientes nuevos, canales y lo más
 *  vendido, de /v1/panel/tablero) y los meses (/v1/panel/informes), con gráficos. */
@Composable
fun Estadisticas() {
    val estado = LocalEstado.current
    val nav = LocalNavegacion.current
    var periodo by remember { mutableStateOf("30d") }
    val nombrePeriodo = when (periodo) { "hoy" -> "hoy"; "7d" -> "últimos 7 días"; else -> "últimos 30 días" }

    PantallaSeccion("Estadísticas", "las estadísticas", {
        // Uno detrás del otro (en iOS `async let` cerraba la app en Release).
        val t: Tablero = estado.api.get("/v1/panel/tablero", listOf("periodo" to periodo))
        DatosEstadisticas(t, estado.api.opcional("/v1/panel/informes"))
    }) { d, actualizar ->
        var anterior by remember { mutableStateOf(periodo) }
        LaunchedEffect(periodo) { if (periodo != anterior) { anterior = periodo; actualizar() } }
        Segmentado(listOf("hoy" to "Hoy", "7d" to "7 días", "30d" to "30 días"), periodo, { periodo = it })
        val t = d.tablero
        TarjetaHeroe("Ventas · $nombrePeriodo", Formato.pesos((t.ventas?.actual ?: 0.0).toInt()), variacion(t.ventas), Tincho.grafico)
        val s = t.serie
        if (s != null && s.actual.any { it > 0 }) {
            TarjetaSeccion("Cómo vienen las ventas", "chart.xyaxis.line") {
                GraficoVentas(s.actual, s.anterior, 150.dp, EtiquetasVentas.etiquetas(t.desde, t.hasta, s.actual.size, periodo == "hoy"))
            }
        }
        GrillaIndicadores(listOf(
            IndicadorIcono("bag.fill", "Pedidos", "${t.kpis.pedidos.actual.toInt()}", variacion(t.kpis.pedidos)),
            IndicadorIcono("receipt", "Ticket promedio", Formato.pesos((t.kpis.ticket?.actual ?: 0.0).toInt()), variacion(t.kpis.ticket), T.avisoFondo to T.avisoTinta),
            IndicadorIcono("person.crop.circle.badge.plus", "Clientes nuevos", "${t.kpis.clientesNuevos.actual.toInt()}", variacion(t.kpis.clientesNuevos), T.okFondo to T.okTinta),
            IndicadorIcono("percent", "Margen del mes", d.informes?.kpis?.margen?.let { "$it %" } ?: "—", tono = T.errorFondo to T.errorTinta),
        ))
        val canales = t.canales
        if (!canales.isNullOrEmpty()) {
            TarjetaSeccion("Por dónde te compran", "point.3.connected.trianglepath.dotted") {
                DonaMedios(canales.map { ParteDona(Marcas.canal(it.canal), (it.parte * 1000).roundToInt()) })
            }
        }
        // Las visitas de la tienda online: personas, páginas vistas, conversión y lo más mirado.
        t.visitas?.let { v ->
            if (v.desde == null) {
                TarjetaSeccion("Visitas de tu tienda", "eye") {
                    Texto(Formato.sinViudas("Todavía no hay visitas medidas. Compartí el link y acá vas a ver quién entra."), Estilo.texto14, C.texto2)
                    BotonTonal("Compartir mi tienda", { nav.abrir(Destino.compartir) })
                }
            } else {
                GrillaIndicadores(listOf(
                    IndicadorIcono("eye.fill", "Visitas", Formato.numero(v.visitantes.actual.toInt()), variacion(v.visitantes)),
                    IndicadorIcono("cart.fill", "Visitas que compran", Formato.porcentaje(Math.round((t.kpis.conversion?.actual ?: 0.0) * 10) / 10.0),
                        variacion(t.kpis.conversion), T.okFondo to T.okTinta),
                ))
                val p = v.paginas
                if (p != null && p.any { it.vistas > 0 }) {
                    TarjetaSeccion("Cómo recorren tu tienda", "line.3.horizontal.decrease") {
                        EmbudoVisitas(p)
                        Texto(Formato.sinViudas("Páginas vistas en el período: de las que entran al inicio, cuántas llegan a cada paso."), Estilo.footnote, C.texto2)
                    }
                }
                if (v.masVistos.isNotEmpty()) {
                    TarjetaSeccion("Lo más mirado", "eye") {
                        val tope = maxOf(1, v.masVistos.maxOf { it.visitantes })
                        Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
                            v.masVistos.forEach { x -> FilaBarra(x.foto, x.nombre, Formato.cantidad(x.visitantes, "persona", "personas"), x.visitantes.toDouble() / tope) }
                        }
                    }
                }
            }
        }
        if (t.masVendidos.isNotEmpty()) {
            TarjetaSeccion("Lo más vendido", "flame") {
                val tope = maxOf(1, t.masVendidos.maxOf { it.unidades })
                Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
                    t.masVendidos.forEach { x -> FilaBarra(x.foto, x.nombre, Formato.cantidad(x.unidades, "unidad", "unidades"), x.unidades.toDouble() / tope) }
                }
            }
        }
        val i = d.informes
        if (i != null && i.meses.size == i.series.ventas.size && i.series.ventas.any { it > 0 }) {
            TarjetaSeccion("Mes a mes", "calendar") {
                GraficoMeses(i.meses, series(i))
                if (i.kpis.ahorro > 0) Texto("Te ahorraste ${Formato.pesos(i.kpis.ahorro)} en comisiones este mes", Estilo.footnoteFuerte, C.okTinta)
            }
        }
        if (i != null && i.insights.isNotEmpty()) {
            TarjetaSeccion("Lo que ve Tincho", "sparkles") {
                Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
                    i.insights.forEach { s ->
                        Row(horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.Top) {
                            Icono("lightbulb.fill", 12f, C.aviso, Modifier.padding(top = 2.dp))
                            Texto(Formato.sinViudas(s), Estilo.texto14, C.texto, Modifier.weight(1f))
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun FilaBarra(foto: String?, nombre: String, cuanto: String, valor: Double) {
    Row(horizontalArrangement = Arrangement.spacedBy(10.dp), verticalAlignment = Alignment.CenterVertically) {
        Miniatura(foto, 38.dp, radio = 10.dp)
        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(4.dp)) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Texto(nombre, Estilo.fila, C.texto, Modifier.weight(1f).padding(end = 6.dp), lineas = 1)
                Texto(cuanto, Estilo.footnoteFuerte, C.texto2, numeros = true)
            }
            BarraProgreso(valor, 7.dp)
        }
    }
}

/** Las series de «Mes a mes»: ventas, clientes nuevos, visitas (si se miden) y margen. */
private fun series(i: Informes): List<SerieMeses> {
    val cuenta: (Int) -> String = { Formato.numero(it) }
    val s = mutableListOf(
        SerieMeses("ventas", "Ventas", i.series.ventas, { Formato.pesos(it) }, { Formato.compacto(it) }),
        SerieMeses("clientes", "Clientes", i.series.clientes, { Formato.cantidad(it, "cliente nuevo", "clientes nuevos") }, cuenta),
    )
    val v = i.series.visitas
    if (v != null && v.any { (it ?: 0) > 0 }) s.add(SerieMeses("visitas", "Visitas", v, { Formato.cantidad(it, "visita", "visitas") }, cuenta))
    if (i.series.margen.any { it != null }) s.add(SerieMeses("margen", "Margen", i.series.margen, { "$it %" }, { "$it %" }))
    return s.filter { it.valores.size == i.meses.size }
}
