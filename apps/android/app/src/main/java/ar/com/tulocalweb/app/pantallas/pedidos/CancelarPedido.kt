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

// Clon de Pantallas/Pedidos/CancelarPedido.swift.

/** 3c · «¿Cancelás el pedido #1048?»: aviso de reintegro manual si ya se cobró, y el botón destructivo con la huella
 *  antes de ejecutar. (La API todavía no guarda el motivo ni avisa por mail: esas dos piezas de la hoja quedan para
 *  cuando las tenga, ver docs/decisiones.md.) */
@Composable
fun CancelarPedido(pedido: PedidoDetalle, alCancelar: suspend () -> Unit) {
    val estado = LocalEstado.current
    val cerrar = LocalCerrarHoja.current
    val alcance = rememberCoroutineScope()
    var cancelando by remember { mutableStateOf(false) }
    val cobrado = pedido.estadoPago in listOf("Acreditada", "Acreditado", "Aprobado", "Cobrado")
    val nombre = pedido.comprador.nombre.split(" ").firstOrNull() ?: pedido.comprador.nombre
    val explicacion = when {
        cobrado && pedido.pago == "mp" -> "$nombre pagó con Mercado Pago: la devolución de ${Formato.pesos(pedido.total)} la hacés desde Mercado Pago. Esto no se puede deshacer."
        cobrado -> "$nombre pagó por ${Marcas.pago(pedido.pago).lowercase()}: la devolución de ${Formato.pesos(pedido.total)} la hacés vos. Esto no se puede deshacer."
        else -> "Todavía no estaba cobrado. El stock vuelve a la tienda. Esto no se puede deshacer."
    }

    suspend fun confirmar() {
        if (!estado.confirmarIdentidad("Para cancelar el pedido #${pedido.numero}")) return
        cancelando = true
        try {
            PedidosApi.cambiarEstado(estado.api, pedido.numero, "Cancelado")
            Haptico.exito()
            cerrar()
            estado.avisar("Cancelaste el pedido #${pedido.numero}")
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
            BotonDestructivo("Cancelar pedido", { alcance.launch { confirmar() } }, habilitado = !cancelando, simbolo = "faceid", cargando = cancelando)
            BotonTexto("Volver", { cerrar() }, alto = 48.dp)
        }
    }
}
