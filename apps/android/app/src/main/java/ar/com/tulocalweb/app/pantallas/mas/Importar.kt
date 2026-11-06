package ar.com.tulocalweb.app.pantallas.mas

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.semantics.selected
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.nucleo.ErrorApi
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.opcional
import ar.com.tulocalweb.app.tema.*
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import kotlinx.serialization.Serializable

// Clon de Pantallas/Mas/Importar.swift: traer la tienda de otra plataforma (Tiendanube, Shopify, WooCommerce,
// Empretienda): productos, clientes, pedidos y redirecciones de los links viejos. La usan el onboarding y Más.

@Serializable
data class Importacion(
    val id: String = "",
    val plataforma: String = "",
    val urlOrigen: String = "",
    val estado: String = "",
    val etapa: String = "",
    val progreso: Progreso = Progreso(),
    val errores: List<ErrorImportacion> = emptyList(),
) {
    @Serializable data class Progreso(val productos: Int = 0, val clientes: Int = 0, val pedidos: Int = 0, val redirecciones: Int = 0, val repetidos: Int = 0, val recortados: Int = 0)
    @Serializable data class ErrorImportacion(val mensaje: String? = null)

    val enCurso: Boolean get() = estado == "pendiente" || estado == "corriendo"
}

enum class PlataformaOrigen {
    tiendanube, shopify, woocommerce, empretienda;

    val nombre: String get() = when (this) { tiendanube -> "Tiendanube"; shopify -> "Shopify"; woocommerce -> "WooCommerce"; empretienda -> "Empretienda" }
    /** Tiendanube se conecta con su cuenta; las demás, con el link de la tienda. */
    val conLink: Boolean get() = this != tiendanube
    val ejemplo: String get() = when (this) { tiendanube -> ""; shopify -> "mitienda.myshopify.com"; woocommerce -> "www.mitienda.com.ar"; empretienda -> "mitienda.empretienda.com.ar" }

    companion object { fun de(s: String) = entries.firstOrNull { it.name == s } }
}

/** Elegir de dónde, arrancar y ver el avance en vivo. */
@Composable
fun TraerTienda(compacto: Boolean = false) {
    val estado = LocalEstado.current
    val alcance = rememberCoroutineScope()
    var plataforma by remember { mutableStateOf<PlataformaOrigen?>(null) }
    var link by remember { mutableStateOf("") }
    var actual by remember { mutableStateOf<Importacion?>(null) }
    var arrancando by remember { mutableStateOf(false) }
    var error by remember { mutableStateOf<String?>(null) }

    /** Mira la importación en curso cada 3 segundos mientras la pantalla está abierta. */
    suspend fun seguir() {
        while (true) {
            val a = estado.api.opcional<Importacion>("/v1/panel/importaciones/actual")
            if (a != null) {
                if (actual?.enCurso == true && a.estado == "lista") { Haptico.exito(); estado.version += 1 }
                actual = a
                if (!a.enCurso) return
            } else if (actual == null) return
            delay(3000)
        }
    }

    @Serializable data class Cuerpo(val plataforma: String, val url: String)

    suspend fun arrancar(p: PlataformaOrigen) {
        error = null
        arrancando = true
        try {
            if (p == PlataformaOrigen.tiendanube) {
                if (!ConectarCuenta.importarTiendanube(estado.api)) { error = "No se completó la conexión con Tiendanube. Probá de nuevo."; return }
            } else {
                var u = link.trim()
                if (!u.lowercase().startsWith("http")) u = "https://$u"
                try {
                    actual = estado.api.enviar<Importacion, Cuerpo>("POST", "/v1/panel/importaciones", Cuerpo(p.name, u))
                } catch (e: ErrorApi) { error = e.detalle; return } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
                    error = "No pudimos arrancar. Revisá el link y probá de nuevo."; return
                }
            }
        } finally { arrancando = false }
        Haptico.exito()
        seguir()
    }

    LaunchedEffect(Unit) { seguir() }

    Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
        val a = actual
        if (a != null && (a.enCurso || a.estado == "lista")) {
            Progreso(a)
        } else {
            PlataformaOrigen.entries.chunked(2).forEach { par ->
                Row(horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                    par.forEach { p ->
                        val forma = RoundedCornerShape(14.dp)
                        Row(Modifier.weight(1f).heightIn(min = 56.dp).tarjeta(radio = 14.dp).border(1.5.dp, if (plataforma == p) C.acento else Color.Transparent, forma)
                            .clickable { Haptico.seleccion(); plataforma = p; error = null }.semantics { selected = plataforma == p }.padding(12.dp),
                            horizontalArrangement = Arrangement.spacedBy(10.dp), verticalAlignment = Alignment.CenterVertically) {
                            MarcaPlataforma(p)
                            Texto(p.nombre, Estilo.fila, C.texto, lineas = 1, escalaMinima = 0.8f)
                        }
                    }
                }
            }
            plataforma?.let { p ->
                if (p.conLink) {
                    CampoTexto("El link de tu tienda en ${p.nombre}", link, { link = it }, teclado = KeyboardType.Uri)
                    Texto(Formato.sinViudas("Traemos productos con fotos, variantes y stock. Los links viejos redirigen solos a tu tienda nueva."), Estilo.footnote, C.texto2)
                } else {
                    Texto(Formato.sinViudas("Entrás con tu cuenta de Tiendanube y traemos productos, clientes y pedidos."), Estilo.footnote, C.texto2)
                }
                error?.let { MensajeError(it) }
                BotonPrincipal(if (p.conLink) "Traer mis productos" else "Conectar Tiendanube", { alcance.launch { arrancar(p) } },
                    habilitado = !(arrancando || (p.conLink && link.trim().length < 4)), cargando = arrancando)
            }
        }
    }
}

@Composable
private fun Progreso(a: Importacion) {
    val lista = a.estado == "lista"
    val nombre = PlataformaOrigen.de(a.plataforma)?.nombre ?: Formato.mayuscula(a.plataforma)
    @Composable
    fun Cifra(n: Int, que: String, modifier: Modifier) {
        Column(modifier.clip(RoundedCornerShape(12.dp)).background(C.pista).padding(10.dp), verticalArrangement = Arrangement.spacedBy(1.dp)) {
            Texto(Formato.numero(n), Estilo.headline, C.texto, numeros = true)
            Texto(que, Estilo.footnote, C.texto2)
        }
    }
    Column(Modifier.fillMaxWidth().tarjeta(radio = 16.dp).padding(14.dp), verticalArrangement = Arrangement.spacedBy(10.dp)) {
        Row(horizontalArrangement = Arrangement.spacedBy(10.dp), verticalAlignment = Alignment.CenterVertically) {
            if (lista) Icono("checkmark.circle.fill", 22f, C.ok) else Rueda(C.acento)
            Column(verticalArrangement = Arrangement.spacedBy(1.dp)) {
                Texto(if (lista) "Listo: trajimos tu tienda de $nombre" else "Trayendo tu tienda de $nombre…", Estilo.fila, C.texto)
                Texto(if (lista) "Revisá los productos en la pestaña Productos" else "Podés seguir mientras tanto: avanza sola", Estilo.footnote, C.texto2)
            }
        }
        Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            Cifra(a.progreso.productos, "productos", Modifier.weight(1f))
            Cifra(a.progreso.clientes, "clientes", Modifier.weight(1f))
            Cifra(a.progreso.pedidos, "pedidos", Modifier.weight(1f))
        }
        if (a.progreso.repetidos > 0) Texto("${Formato.cantidad(a.progreso.repetidos, "ya estaba", "ya estaban")} y no se duplicó.", Estilo.footnote, C.texto2)
    }
}

/** La inicial de la plataforma en su color (sin logos de terceros dibujados a mano). */
@Composable
fun MarcaPlataforma(plataforma: PlataformaOrigen) {
    val (fondo, letra) = when (plataforma) {
        PlataformaOrigen.tiendanube -> hex(0x2C3357) to "T"
        PlataformaOrigen.shopify -> hex(0x5E8E3E) to "S"
        PlataformaOrigen.woocommerce -> hex(0x7F54B3) to "W"
        PlataformaOrigen.empretienda -> hex(0xE4572E) to "E"
    }
    Box(Modifier.size(30.dp).clip(RoundedCornerShape(8.dp)).background(fondo), contentAlignment = Alignment.Center) { Texto(letra, Estilo(15f, Peso.bold), Color.White) }
}

/** Más › Traer mi tienda de otra plataforma. */
@Composable
fun ImportarTienda() {
    Column(Modifier.fillMaxSize().background(C.fondo).verticalScroll(rememberScrollState()).margenPantalla().padding(bottom = 32.dp), verticalArrangement = Arrangement.spacedBy(14.dp)) {
        Texto("Traer mi tienda", Estilo.tituloGrande, C.texto)
        Texto(Formato.sinViudas("¿Vendías en otra plataforma? Traé tus productos, clientes y pedidos en un par de minutos."), Estilo.body16, C.texto2)
        TraerTienda()
    }
}
