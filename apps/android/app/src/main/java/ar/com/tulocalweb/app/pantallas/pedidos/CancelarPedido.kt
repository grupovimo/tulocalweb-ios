package ar.com.tulocalweb.app.pantallas.pedidos

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.nucleo.ErrorApi
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.Marcas
import ar.com.tulocalweb.app.nucleo.PedidoDetalle
import ar.com.tulocalweb.app.tema.*
import kotlinx.coroutines.launch
import java.util.UUID

// Clon de Pantallas/Pedidos/CancelarPedido.swift.

/** 3c · «¿Cancelás el pedido #1048?»: qué pasa con lo cobrado (también la seña) y el botón destructivo con la huella
 *  antes de ejecutar. Si se cobró con Mercado Pago y la persona maneja las finanzas, devuelve el pago desde acá
 *  (POST …/reembolso) y después cancela. (La API todavía no guarda el motivo ni avisa por mail: esas dos piezas de la hoja quedan para
 *  cuando las tenga, ver docs/decisiones.md.) */
@Composable
fun CancelarPedido(pedido: PedidoDetalle, alCancelar: suspend () -> Unit) {
    val estado = LocalEstado.current
    val cerrar = LocalCerrarHoja.current
    val alcance = rememberCoroutineScope()
    var cancelando by remember { mutableStateOf(false) }
    /** Ya se devolvió (si después falló cancelar, el reintento no vuelve a devolver). */
    var devuelto by remember { mutableStateOf(false) }
    /** Una sola clave por hoja: si se reintenta, Mercado Pago no devuelve dos veces. */
    val clave = remember { UUID.randomUUID().toString() }
    // «Seña» es cobrado también: se pagó una parte (la API la cuenta igual, sinpago.go).
    val sena = pedido.estadoPago == "Seña"
    val cobrado = pedido.estadoPago in listOf("Acreditada", "Acreditado", "Aprobado", "Cobrado", "Seña")
    val monto = if (sena) pedido.sena?.monto ?: pedido.total else pedido.total
    val puedeDevolver = cobrado && !devuelto && pedido.pago == "mp" && estado.puede("finanzas")
    val nombre = pedido.comprador.nombre.split(" ").firstOrNull() ?: pedido.comprador.nombre
    val que = if (sena) "una seña de ${Formato.pesos(monto)}" else Formato.pesos(monto)
    val explicacion = when {
        pedido.estadoPago == "Devuelto" -> "Ya le devolviste el pago a $nombre. El stock vuelve a la tienda. Esto no se puede deshacer."
        cobrado && pedido.pago == "mp" && devuelto -> "Ya le devolviste $que a $nombre. Falta cancelar el pedido. Esto no se puede deshacer."
        cobrado && pedido.pago == "mp" && puedeDevolver -> "$nombre pagó $que con Mercado Pago. Se lo devolvés desde acá al cancelar. Esto no se puede deshacer."
        cobrado && pedido.pago == "mp" -> "$nombre pagó $que con Mercado Pago: la devolución la hace quien maneja las finanzas de la tienda. Esto no se puede deshacer."
        cobrado -> "$nombre pagó $que por ${Marcas.pago(pedido.pago).lowercase()}: la devolución la hacés vos. Esto no se puede deshacer."
        else -> "Todavía no estaba cobrado. El stock vuelve a la tienda. Esto no se puede deshacer."
    }

    suspend fun confirmar(devolver: Boolean) {
        if (!estado.confirmarIdentidad("Para cancelar el pedido #${pedido.numero}")) return
        cancelando = true
        try {
            // Primero la devolución: si falla, el pedido sigue como estaba.
            if (devolver) {
                estado.api.hacer("POST", "/v1/panel/pedidos/${pedido.numero}/reembolso", mapOf("Idempotency-Key" to clave))
                devuelto = true
            }
            PedidosApi.cambiarEstado(estado.api, pedido.numero, "Cancelado")
            Haptico.exito()
            cerrar()
            estado.avisar(if (devolver) "Devolviste ${Formato.pesos(monto)} y cancelaste el pedido #${pedido.numero}" else "Cancelaste el pedido #${pedido.numero}")
            estado.version += 1
            alCancelar()
            estado.refrescarResumen()
        } catch (e: ErrorApi) { Haptico.error(); estado.avisar(e.detalle, ok = false) } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
        } finally { cancelando = false }
    }

    Column(Modifier.fillMaxSize().padding(horizontal = 20.dp), verticalArrangement = Arrangement.spacedBy(14.dp)) {
        Grabber()
        Column(Modifier.padding(horizontal = 4.dp).padding(top = 8.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
            Texto("¿Cancelás el pedido #${pedido.numero}?", Estilo.tituloHoja, C.texto)
            Texto(Formato.sinViudas(explicacion), Estilo.subhead, C.texto2, interlineado = 2f)
        }
        Column(Modifier.fillMaxWidth().tarjeta()) {
            pedido.items.take(4).forEachIndexed { i, it ->
                if (i > 0) Separador(66.dp)
                Row(Modifier.padding(horizontal = 14.dp, vertical = 10.dp), horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
                    Miniatura(it.foto, 40.dp, radio = 11.dp)
                    Texto("${it.cantidad} × ${it.nombre}", Estilo.body16, C.texto, lineas = 1)
                }
            }
        }
        Spacer(Modifier.weight(1f))
        Column(Modifier.padding(bottom = 12.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
            BotonDestructivo(if (puedeDevolver) "Devolver ${Formato.pesos(monto)} y cancelar" else "Cancelar pedido", { alcance.launch { confirmar(puedeDevolver) } },
                habilitado = !cancelando, simbolo = "faceid", cargando = cancelando)
            if (puedeDevolver) BotonTexto("Cancelar sin devolver", { alcance.launch { confirmar(false) } }, alto = 48.dp, color = C.errorTinta, habilitado = !cancelando)
            BotonTexto("Volver", { cerrar() }, alto = 48.dp)
        }
    }
}
