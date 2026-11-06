package ar.com.tulocalweb.app.compartido

import android.content.Context
import ar.com.tulocalweb.app.nucleo.Api
import ar.com.tulocalweb.app.nucleo.Llavero
import ar.com.tulocalweb.app.nucleo.opcional
import kotlinx.coroutines.async
import kotlinx.coroutines.coroutineScope
import kotlinx.serialization.Serializable

// Clon de Compartido/Hoy.swift: lo que comparten la app, los widgets y los pedidos por preparar: la sesión, los datos
// del día y los colores. En Android corren en el mismo proceso: la sesión es la del Llavero.

/** La sesión guardada por la app, para usarla fuera de la app abierta (widgets, acciones de los avisos). */
object SesionCompartida {
    fun token(ctx: Context): String? {
        Llavero.contexto = Llavero.contexto ?: ctx.applicationContext
        return Llavero.leer("token")
    }

    /** Una Api con la sesión compartida (null sin sesión). */
    fun api(ctx: Context): Api? = token(ctx)?.let { Api(Api.urlConfigurada(), it) }
}

/** Cómo viene el día, para los widgets. */
@Serializable
data class DatosHoy(
    val conSesion: Boolean,
    val montos: Boolean,
    val ventasHoy: Double,
    val ventasAyer: Double,
    val serie: List<Double>,
    val pedidosHoy: Int,
    val porPreparar: Int,
    val ultimo: Pedido? = null,
    /** El más nuevo de los que esperan preparación (para los pedidos por preparar). */
    val ultimoPorPreparar: Pedido? = null,
) {
    @Serializable data class Pedido(val numero: Int, val cliente: String, val total: Int? = null, val estado: String)

    val delta: Double? get() = if (ventasAyer > 0) (ventasHoy - ventasAyer) / ventasAyer * 100 else null

    companion object {
        val ejemplo = DatosHoy(true, true, 184_350.0, 164_000.0,
            listOf(0.0, 8.0, 15.0, 22.0, 40.0, 52.0, 61.0, 80.0, 95.0, 110.0, 128.0, 150.0, 170.0, 184.0), 12, 3,
            Pedido(1048, "Lucía Fernández", 23_400, "Para preparar"), Pedido(1048, "Lucía Fernández", 23_400, "Para preparar"))
        val sinSesion = DatosHoy(false, false, 0.0, 0.0, emptyList(), 0, 0)

        @Serializable private data class Comparado(val actual: Double = 0.0, val anterior: Double = 0.0)
        @Serializable private data class Serie(val actual: List<Double> = emptyList())
        @Serializable private data class Kpis(val pedidos: Comparado = Comparado())
        @Serializable private data class Breve(val numero: Int, val cliente: String = "", val total: Int? = null, val estado: String = "")
        @Serializable private data class Tablero(val montos: Boolean = false, val ventas: Comparado? = null, val serie: Serie? = null, val kpis: Kpis = Kpis(), val ultimosPedidos: List<Breve> = emptyList())
        @Serializable private data class Insignias(val pedidos: Int = 0)
        @Serializable private data class Resumen(val insignias: Insignias = Insignias())

        suspend fun traer(ctx: Context): DatosHoy = coroutineScope {
            val api = SesionCompartida.api(ctx) ?: return@coroutineScope sinSesion
            val t = async { api.opcional<Tablero>("/v1/panel/tablero", listOf("periodo" to "hoy")) }
            val r = async { api.opcional<Resumen>("/v1/panel/resumen") }
            val tab = t.await() ?: return@coroutineScope sinSesion
            val res = r.await()
            val u = tab.ultimosPedidos.firstOrNull()
            val pp = tab.ultimosPedidos.firstOrNull { it.estado == "Para preparar" }
            DatosHoy(true, tab.montos, tab.ventas?.actual ?: 0.0, tab.ventas?.anterior ?: 0.0, tab.serie?.actual ?: emptyList(),
                tab.kpis.pedidos.actual.toInt(), res?.insignias?.pedidos ?: 0,
                u?.let { Pedido(it.numero, it.cliente, it.total, it.estado) }, pp?.let { Pedido(it.numero, it.cliente, it.total, it.estado) })
        }
    }
}

/** Colores del handoff para los widgets (sin las fuentes de la app). */
object ColorWidget {
    val fondo = 0xFBF7F2L to 0x22140DL
    val texto = 0x2A140CL to 0xFBF7F2L
    val texto2 = 0x6B5244L to 0xCDB8A8L
    val acento = 0xD9652FL to 0xE07A48L
    val primario = 0xB0441DL to 0xE07A48L
    val ok = 0x1F6B48L to 0x6FD0A0L
    val error = 0xA23F1CL to 0xF08A63L
}

object PesosWidget {
    fun texto(n: Double): String = "$ " + java.text.DecimalFormat("#,##0", java.text.DecimalFormatSymbols(java.util.Locale.forLanguageTag("es-AR"))).format(Math.round(n))
}
