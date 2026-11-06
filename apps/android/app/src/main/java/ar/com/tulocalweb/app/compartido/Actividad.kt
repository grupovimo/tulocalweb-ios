package ar.com.tulocalweb.app.compartido

import android.content.Context
import ar.com.tulocalweb.app.nucleo.ActividadEnApp
import kotlinx.serialization.Serializable

// Clon de Compartido/Actividad.swift: 9a/9b · los pedidos por preparar (la Live Activity de iOS): viven mientras haya
// alguno. En Android es una notificación en curso (nucleo/ActividadEnApp.kt).

/** Lo que muestra (PedidosPorPreparar.ContentState en iOS; push.EstadoActividad en Go). */
@Serializable
data class PedidosPorPreparar(val porPreparar: Int, val preparadosHoy: Int, val ultimoNumero: Int? = null, val ultimoCliente: String? = null)

/** «Preparado» desde los pedidos por preparar: pasa el último pedido a «Listo para despachar» (MarcarPreparado). */
object MarcarPreparado {
    @Serializable private data class Cuerpo(val estado: String)

    suspend fun hacer(ctx: Context, numero: Int) {
        val api = SesionCompartida.api(ctx) ?: return
        api.hacer("POST", "/v1/panel/pedidos/$numero/estado", Cuerpo("Listo para despachar"))
        val datos = DatosHoy.traer(ctx)
        val previo = ActividadEnApp.estadoActual
        val nuevo = PedidosPorPreparar(datos.porPreparar, (previo?.preparadosHoy ?: 0) + 1, datos.ultimoPorPreparar?.numero, datos.ultimoPorPreparar?.cliente)
        if (datos.porPreparar == 0) ActividadEnApp.cerrar(ctx, nuevo) else ActividadEnApp.mostrar(ctx, ActividadEnApp.tiendaActual ?: "Tu Local Web", nuevo)
    }
}
