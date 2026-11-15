package ar.com.tulocalweb.app.pantallas.mas

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.tema.*
import kotlinx.coroutines.launch
import kotlinx.serialization.Serializable

/**
 * Los recordatorios del carrito abandonado (§62), como el panel (/panel/carritos-abandonados): hasta tres mails desde que
 * dejó el carrito (por defecto 1 h, 24 h y 72 h), con un cupón único opcional en el segundo o el tercero. Se cortan si
 * compra o se da de baja. Las reglas son las de Go (ventas/recordatorios.go) y las valida la API.
 */
@Serializable
data class RecordatoriosCarrito(val activo: Boolean = false, val pasos: List<Paso> = emptyList()) {
    @Serializable data class Cupon(val descuento: String, val valor: Int = 0, val dias: Int = 3)
    @Serializable data class Paso(val activo: Boolean = true, val horas: Int = 1, val cupon: Cupon? = null)

    companion object {
        val orden = listOf("Primero", "Segundo", "Tercero")
        /** Las demoras que se ofrecen (en horas), de 1 hora a 7 días. */
        val demoras = listOf(1, 2, 3, 6, 12, 24, 36, 48, 72, 96, 120, 168)
        /** Los cupones que se ofrecen en un toque. */
        val cupones: List<Pair<String, Cupon?>> = listOf(
            "Sin cupón" to null,
            "5 % off" to Cupon("porcentaje", 5, 3),
            "10 % off" to Cupon("porcentaje", 10, 3),
            "15 % off" to Cupon("porcentaje", 15, 3),
            "Envío gratis" to Cupon("envioGratis", 0, 3),
        )

        fun demora(h: Int): String = when {
            h < 24 -> if (h == 1) "1 hora" else "$h horas"
            h % 24 == 0 -> if (h == 24) "1 día" else "${h / 24} días"
            else -> "$h horas"
        }

        fun cupon(c: Cupon?): String = when {
            c == null -> "Sin cupón"
            c.descuento == "envioGratis" -> "envío gratis"
            c.descuento == "porcentaje" -> "${c.valor} % off"
            else -> "${Formato.pesos(c.valor)} off"
        }

        fun resumen(p: Paso): String = "${demora(p.horas)} después" + (if (p.cupon != null) " · ${cupon(p.cupon)}" else "")
    }
}

/** La hoja «Configurar demoras y cupones». */
@Composable
fun HojaRecordatoriosCarrito(inicial: RecordatoriosCarrito, guardar: suspend (RecordatoriosCarrito) -> Boolean) {
    val cerrar = LocalCerrarHoja.current
    val alcance = rememberCoroutineScope()
    var r by remember { mutableStateOf(inicial) }
    var guardando by remember { mutableStateOf(false) }
    fun cambiar(i: Int, f: (RecordatoriosCarrito.Paso) -> RecordatoriosCarrito.Paso) {
        r = r.copy(pasos = r.pasos.mapIndexed { j, p -> if (j == i) f(p) else p })
    }
    Column(Modifier.fillMaxWidth().verticalScroll(rememberScrollState()).padding(horizontal = 20.dp).padding(bottom = 24.dp), verticalArrangement = Arrangement.spacedBy(14.dp)) {
        Grabber()
        Texto("Recordatorios del carrito", Estilo.tituloHojaChico, C.texto)
        Texto(Formato.sinViudas("Salen desde que dejó el carrito. Si compra o se da de baja, no le llega ninguno más."), Estilo.subhead, C.texto2)
        r.pasos.forEachIndexed { i, p ->
            Column(Modifier.fillMaxWidth().tarjeta().padding(14.dp), verticalArrangement = Arrangement.spacedBy(10.dp)) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Texto(RecordatoriosCarrito.orden[i], Estilo.fila, C.texto, Modifier.weight(1f))
                    Llave(p.activo, { on -> cambiar(i) { it.copy(activo = on) } })
                }
                Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    MenuBoton(RecordatoriosCarrito.demoras.map { h ->
                        OpcionDeMenu("${RecordatoriosCarrito.demora(h)} después", if (h == p.horas) "checkmark" else "") { cambiar(i) { it.copy(horas = h) } }
                    }, "Cuándo sale") { EtiquetaMenu("${RecordatoriosCarrito.demora(p.horas)} después") }
                    if (i > 0) MenuBoton(RecordatoriosCarrito.cupones.map { (t, c) ->
                        OpcionDeMenu(t, if (c == p.cupon) "checkmark" else "") { cambiar(i) { it.copy(cupon = c) } }
                    }, "Cupón") { EtiquetaMenu(RecordatoriosCarrito.cupon(p.cupon)) }
                }
            }
        }
        Row(horizontalArrangement = Arrangement.spacedBy(6.dp)) {
            Icono("ticket", 14f, C.texto2)
            Texto(Formato.sinViudas("El cupón es único para cada carrito, vale 3 días y se aplica solo con el link del mail."), Estilo.footnote, C.texto2)
        }
        BotonPrincipal(if (guardando) "Guardando…" else "Guardar", {
            alcance.launch {
                guardando = true
                if (guardar(r)) cerrar()
                guardando = false
            }
        }, habilitado = !guardando)
    }
}

@Composable
private fun EtiquetaMenu(t: String) {
    val forma = RoundedCornerShape(12.dp)
    Row(Modifier.heightIn(min = 44.dp).clip(forma).background(C.superficie).border(1.dp, C.linea2, forma).padding(horizontal = 12.dp),
        horizontalArrangement = Arrangement.spacedBy(6.dp), verticalAlignment = Alignment.CenterVertically) {
        Texto(t, Estilo.texto14, C.texto, lineas = 1)
        Icono("chevron.up.chevron.down", 11f, C.texto2)
    }
}
