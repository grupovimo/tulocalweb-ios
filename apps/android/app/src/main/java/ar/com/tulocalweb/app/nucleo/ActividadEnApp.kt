package ar.com.tulocalweb.app.nucleo

import android.content.Context
import android.os.Bundle
import androidx.core.app.NotificationCompat
import androidx.core.app.NotificationManagerCompat
import ar.com.tulocalweb.app.R
import ar.com.tulocalweb.app.app.Push
import ar.com.tulocalweb.app.compartido.DatosHoy
import ar.com.tulocalweb.app.compartido.PedidosPorPreparar
import ar.com.tulocalweb.app.widgets.Widgets
import kotlinx.serialization.Serializable

// Clon de Nucleo/ActividadEnApp.swift. La Live Activity de los pedidos por preparar es, en Android, una notificación
// en curso con el número, el último pedido, la barra del día y «Ver pedido» / «Preparado» (en Android 16 se pide como
// Live Update: arriba en la barra y en la pantalla bloqueada).

/** La app arranca, actualiza y cierra los pedidos por preparar cada vez que refresca el resumen (al abrir, al volver y
 *  con cada push de venta), y recarga los widgets. */
object ActividadEnApp {
    private const val AVISO = 9001
    /** Lo que se está mostrando (null: no hay). */
    @Volatile var estadoActual: PedidosPorPreparar? = null
    @Volatile var tiendaActual: String? = null

    suspend fun sincronizar(ctx: Context, tienda: String) {
        Widgets.actualizar(ctx)
        if (!Push.hayPermiso(ctx)) return
        val datos = DatosHoy.traer(ctx)
        if (!datos.conSesion) return
        val previo = estadoActual
        val estado = PedidosPorPreparar(
            datos.porPreparar,
            maxOf(previo?.preparadosHoy ?: 0, previo?.let { maxOf(0, it.porPreparar - datos.porPreparar) + it.preparadosHoy } ?: 0),
            datos.ultimoPorPreparar?.numero, datos.ultimoPorPreparar?.cliente,
        )
        if (datos.porPreparar == 0) cerrar(ctx, estado)
        else if (previo != estado) mostrar(ctx, tienda, estado)
        avisar(ctx)
    }

    /** El servidor la actualiza por FCM con la app cerrada (§54): se le pasa el token del teléfono como el de la
     *  actividad y el de «arrancarla». */
    var api: Api? = null
    var tokenPush: String? = null

    @Serializable private data class Cuerpo(val token: String, val actividad: String?, val inicio: String?)

    /** El token para arrancarla desde el servidor y la que ya estaba (de antes de abrir la app). */
    suspend fun arrancarEscucha(ctx: Context) = avisar(ctx)

    private suspend fun avisar(ctx: Context) {
        val api = api ?: return
        val t = tokenPush ?: return
        intentar { api.hacer("PUT", "/v1/panel/dispositivos/actividad", Cuerpo(t, if (estadoActual != null) t else "", t)) }
    }

    /** Lo que manda el servidor por FCM: {tipo: actividad, tienda, porPreparar, preparadosHoy, ultimoNumero, ultimoCliente, terminar}. */
    fun desdePush(ctx: Context, d: Map<String, String>) {
        val e = PedidosPorPreparar(d["porPreparar"]?.toIntOrNull() ?: 0, d["preparadosHoy"]?.toIntOrNull() ?: 0,
            d["ultimoNumero"]?.toIntOrNull(), d["ultimoCliente"]?.takeIf { it.isNotEmpty() })
        if (d["terminar"] == "1" || e.porPreparar == 0) cerrar(ctx, e) else mostrar(ctx, d["tienda"] ?: tiendaActual ?: "Tu Local Web", e)
        Widgets.actualizar(ctx)
    }

    fun mostrar(ctx: Context, tienda: String, e: PedidosPorPreparar) {
        if (!Push.hayPermiso(ctx)) return
        estadoActual = e
        tiendaActual = tienda
        val ultimo = if (e.ultimoNumero != null && e.ultimoCliente != null) "Último: #${e.ultimoNumero} · ${e.ultimoCliente}" else ""
        val b = NotificationCompat.Builder(ctx, Push.CANAL_PREPARAR)
            .setSmallIcon(R.drawable.ic_aviso)
            .setColor(0xFFD9652F.toInt())
            .setSubText(tienda)
            .setContentTitle("${e.porPreparar} por preparar")
            .setContentText(ultimo)
            .setProgress(maxOf(1, e.preparadosHoy + e.porPreparar), e.preparadosHoy, false)
            .setOngoing(true)
            .setOnlyAlertOnce(true)
            .setShowWhen(false)
            .setCategory(NotificationCompat.CATEGORY_PROGRESS)
            .setVisibility(NotificationCompat.VISIBILITY_PUBLIC)
            .setContentIntent(Push.abrir(ctx, "pedidos", AVISO))
            // Android 16: pedirla como Live Update (EXTRA_REQUEST_PROMOTED_ONGOING); antes se ignora.
            .addExtras(Bundle().apply { putBoolean("android.requestPromotedOngoing", true) })
        e.ultimoNumero?.let { n ->
            b.addAction(0, "Ver pedido", Push.abrir(ctx, "pedido/$n", AVISO + 1))
            b.addAction(0, "Preparado", Push.accionPreparado(ctx, n, 0, actividad = true))
        }
        runCatching { NotificationManagerCompat.from(ctx).notify(AVISO, b.build()) }
    }

    fun cerrar(ctx: Context, e: PedidosPorPreparar? = null) {
        estadoActual = null
        NotificationManagerCompat.from(ctx).cancel(AVISO)
    }

    fun terminarTodo(ctx: Context) {
        Widgets.actualizar(ctx)
        cerrar(ctx)
    }
}
