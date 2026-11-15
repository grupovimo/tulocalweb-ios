package ar.com.tulocalweb.app.pantallas.mas

import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.nucleo.FechaApi
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.Tincho
import ar.com.tulocalweb.app.nucleo.abrirUrl
import ar.com.tulocalweb.app.pantallas.pedidos.urlWhatsApp
import ar.com.tulocalweb.app.tema.*
import kotlinx.serialization.Serializable
import java.time.Instant

// Newsletter: los contactos de la tienda con su origen (el pie de la tienda, el checkout o una ventana de Promociones:
// «Ventana: Bienvenida 10 %», «Ruleta»), el WhatsApp si lo dejaron y un filtro por origen. GET /v1/panel/newsletter
// (trae las ventanas para nombrar el origen).

@Serializable
data class NewsletterDatos(
    val suscriptores: List<Suscriptor> = emptyList(),
    val confirmados: Int = 0,
    val pendientes: Int = 0,
    val ventanas: List<VentanaOrigen>? = null,
) {
    @Serializable
    data class Suscriptor(val id: String, val email: String = "", val origen: String = "", val nombre: String? = null, val whatsapp: String? = null,
                          val confirmado: Boolean = false, val creado: FechaApi = Instant.EPOCH)
    @Serializable
    data class VentanaOrigen(val id: String, val nombre: String = "", val contenido: String = "")

    /** El origen como lo lee la tienda (origenDe en la web). */
    fun origen(o: String): String = when (o) {
        "tienda" -> "Tienda online"
        "checkout" -> "Checkout"
        else -> {
            if (!o.startsWith("ventana:")) o
            else {
                val v = ventanas?.firstOrNull { it.id == o.drop(8) }
                when {
                    v == null || v.nombre.isEmpty() -> "Ventana"
                    v.contenido == "ruleta" -> if (v.nombre.contains("ruleta", ignoreCase = true)) v.nombre else "Ruleta: ${v.nombre}"
                    else -> "Ventana: ${v.nombre}"
                }
            }
        }
    }

    data class Origen(val clave: String, val nombre: String, val cuenta: Int)

    /** Los orígenes que hay, con cuántos contactos trae cada uno (los que más primero). */
    val origenes: List<Origen> get() = suscriptores.groupingBy { it.origen }.eachCount()
        .map { Origen(it.key, origen(it.key), it.value) }
        .sortedWith(compareByDescending<Origen> { it.cuenta }.thenBy { it.nombre })
}

@Composable
fun Newsletter() {
    val estado = LocalEstado.current
    val ctx = LocalContext.current
    var filtro by remember { mutableStateOf<String?>(null) }

    PantallaSeccion("Newsletter", "tu newsletter", { estado.api.get<NewsletterDatos>("/v1/panel/newsletter") }) { d, _ ->
        TarjetaHeroe("Suscriptos confirmados", "${d.confirmados}",
            if (d.pendientes > 0) Formato.cantidad(d.pendientes, "sin confirmar el mail", "sin confirmar el mail") else null, Tincho.portapapeles)
        if (d.suscriptores.isEmpty()) {
            Box(Modifier.padding(top = 30.dp)) {
                EstadoTincho(Tincho.durmiendo, "Todavía no hay contactos", "Se suman desde el pie de tu tienda, al comprar o con una ventana de Promociones.")
            }
        } else {
            val origenes = d.origenes
            if (origenes.size > 1) {
                Row(Modifier.fillMaxWidth().horizontalScroll(rememberScrollState()), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    ChipFiltro("Todos", d.suscriptores.size, activo = filtro == null) { filtro = null; Haptico.seleccion() }
                    origenes.forEach { o -> ChipFiltro(o.nombre, o.cuenta, activo = filtro == o.clave) { filtro = o.clave; Haptico.seleccion() } }
                }
            }
            val lista = d.suscriptores.filter { filtro == null || it.origen == filtro }
            GrupoFilas(lista, { it.id }, 14.dp) { s ->
                val nombre = s.nombre?.ifEmpty { null }
                Row(Modifier.fillMaxWidth().padding(horizontal = 14.dp, vertical = 12.dp), horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
                    Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                        Texto(nombre ?: s.email, Estilo.fila, C.texto, lineas = 1)
                        if (nombre != null) Texto(s.email, Estilo.footnote, C.texto2, lineas = 1)
                        Texto("${d.origen(s.origen)} · ${Formato.hace(s.creado)}", Estilo.footnote, C.texto2, lineas = 2)
                    }
                    Column(horizontalAlignment = Alignment.End, verticalArrangement = Arrangement.spacedBy(6.dp)) {
                        if (!s.confirmado) ChipSimple("Sin confirmar", C.avisoFondo, C.avisoTinta)
                        val w = s.whatsapp
                        val u = urlWhatsApp(w)
                        if (w != null && u != null) {
                            BotonTonal("WhatsApp", { abrirUrl(ctx, u) }, Modifier.semantics { contentDescription = "Escribirle por WhatsApp al $w" },
                                alto = 32.dp, simbolo = "message")
                        }
                    }
                }
            }
        }
    }
}
