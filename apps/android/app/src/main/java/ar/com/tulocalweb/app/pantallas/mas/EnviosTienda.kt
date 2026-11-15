package ar.com.tulocalweb.app.pantallas.mas

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ColumnScope
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.imePadding
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
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.nucleo.ErrorApi
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.tema.*
import kotlinx.coroutines.launch
import kotlinx.serialization.Serializable

/**
 * Envíos y retiro (Configuración › Envíos del panel, /panel/envios/configurar): retiro en el local, envío propio en moto,
 * Correo Argentino (se conecta en Integraciones) y el envío gratis desde el monto de la tienda. Las reglas son las de Go
 * (configuracion/entregas.go): al menos una forma de entrega, la moto con su zona y el envío gratis desde $ 1.000.
 */
@Serializable
data class EntregasTienda(
    val retiro: Boolean = true,
    val retiroDireccion: String = "",
    val retiroHorario: String = "",
    val moto: Boolean = false,
    val motoCosto: Int = 0,
    val motoZona: String = "",
    val motoDemora: String = "",
    val correo: Boolean? = null,
    val envioGratis: Boolean? = null,
    val envioGratisMonto: Int? = null,
)

@Composable
fun EnviosTienda() {
    val estado = LocalEstado.current
    val cerrar = LocalCerrarHoja.current
    val alcance = rememberCoroutineScope()
    var e by remember { mutableStateOf<EntregasTienda?>(null) }
    var costo by remember { mutableStateOf("") }
    var monto by remember { mutableStateOf("") }
    var guardando by remember { mutableStateOf(false) }

    LaunchedEffect(Unit) {
        try {
            val v = estado.api.get<EntregasTienda>("/v1/panel/configuracion/entregas")
            costo = v.motoCosto.toString()
            monto = (v.envioGratisMonto ?: 40000).toString()
            e = v
        } catch (x: kotlinx.coroutines.CancellationException) { throw x } catch (x: Exception) {
            estado.avisar("No pudimos traer los envíos", ok = false)
        }
    }

    suspend fun guardar() {
        val v = e ?: return
        val listo = v.copy(motoCosto = costo.filter { it.isDigit() }.toIntOrNull() ?: 0,
            envioGratisMonto = monto.filter { it.isDigit() }.toIntOrNull() ?: 0, correo = null)
        if (listo.envioGratis == true && (listo.envioGratisMonto ?: 0) < 1000) {
            estado.avisar("El envío gratis va desde $ 1.000 hasta $ 100.000.000.", ok = false)
            return
        }
        guardando = true
        try {
            estado.api.hacer("PUT", "/v1/panel/configuracion/entregas", listo)
            Haptico.exito()
            estado.avisar("Envíos y retiro guardados")
            cerrar()
        } catch (x: ErrorApi) { estado.avisar(x.detalle, ok = false) } catch (x: kotlinx.coroutines.CancellationException) { throw x } catch (x: Exception) {
            estado.avisar("No pudimos guardarlos", ok = false)
        } finally { guardando = false }
    }

    ConBarraAccion({
        if (e != null) BotonPrincipal(if (guardando) "Guardando…" else "Guardar", { alcance.launch { guardar() } }, habilitado = !guardando)
    }, Modifier.background(C.fondo).imePadding()) { abajo ->
        Column(Modifier.fillMaxSize().verticalScroll(rememberScrollState()).margenPantalla().padding(bottom = 32.dp + abajo), verticalArrangement = Arrangement.spacedBy(12.dp)) {
            Texto("Envíos y retiro", Estilo.tituloGrande, C.texto)
            Texto(Formato.sinViudas("Cómo les llega lo que compran en tu tienda."), Estilo.subhead, C.texto2)
            val v = e
            if (v == null) EsqueletoFilas(5, 0.dp)
            else {
                OpcionEnvio("Retiro en el local", "Sin costo · te avisamos cuando está listo", "storefront", v.retiro, { e = v.copy(retiro = it) }) {
                    if (v.retiro) {
                        CampoEnvio("Dirección", v.retiroDireccion, { e = e?.copy(retiroDireccion = it) }, "Calle, número y localidad")
                        CampoEnvio("Horario", v.retiroHorario, { e = e?.copy(retiroHorario = it) }, "Lunes a viernes de 10 a 19")
                    }
                }
                OpcionEnvio("Envío propio en moto", "Con tu moto o tu cadete, a un costo fijo", "paperplane", v.moto, { e = v.copy(moto = it) }) {
                    if (v.moto) {
                        CampoEnvio("Costo en pesos", costo, { costo = it }, "3500", KeyboardType.Number)
                        CampoEnvio("Cuánto tarda", v.motoDemora, { e = e?.copy(motoDemora = it) }, "En el día")
                        CampoEnvio("Hasta dónde llega", v.motoZona, { e = e?.copy(motoZona = it) }, "Palermo, Belgrano, Núñez")
                    }
                }
                val conectado = v.correo == true
                Row(Modifier.fillMaxWidth().tarjeta().padding(16.dp), horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
                    IconoOpcion("envelope")
                    Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                        Texto("Correo Argentino", Estilo.fila, C.texto)
                        Texto(if (conectado) "Cotiza con tu cuenta de MiCorreo" else "Conectalo en Integraciones para enviar a todo el país", Estilo.footnote, C.texto2)
                    }
                    ChipSimple(if (conectado) "Conectado" else "Sin conectar", if (conectado) C.okFondo else C.apagadoFondo, if (conectado) C.okTinta else C.neutroTinta)
                }
                val gratis = v.envioGratis == true
                OpcionEnvio("Envío gratis", if (gratis) "En moto y por Correo, desde el monto que elijas" else "Para que sumen algo más al carrito", "gift", gratis,
                    { e = v.copy(envioGratis = it) }) {
                    if (gratis) {
                        CampoEnvio("Desde qué monto", monto, { monto = it }, "40000", KeyboardType.Number)
                        Texto(Formato.sinViudas("La tienda muestra cuánto les falta en la bolsa y en el carrito, y el envío sale gratis al pagar."), Estilo.footnote, C.texto2)
                    }
                }
            }
        }
    }
}

@Composable
private fun OpcionEnvio(titulo: String, detalle: String, simbolo: String, prendido: Boolean, alCambiar: (Boolean) -> Unit, contenido: @Composable ColumnScope.() -> Unit) {
    Column(Modifier.fillMaxWidth().tarjeta().padding(16.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
        Row(horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
            IconoOpcion(simbolo)
            Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                Texto(titulo, Estilo.fila, C.texto)
                Texto(Formato.sinViudas(detalle), Estilo.footnote, C.texto2)
            }
            Llave(prendido, alCambiar)
        }
        contenido()
    }
}

@Composable
private fun CampoEnvio(titulo: String, texto: String, alCambiar: (String) -> Unit, ejemplo: String, teclado: KeyboardType = KeyboardType.Text) {
    Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
        Texto(titulo, Estilo.footnoteFuerte, C.texto2)
        val forma = RoundedCornerShape(14.dp)
        Entrada(texto, alCambiar, ejemplo, Modifier.fillMaxWidth().clip(forma).background(C.superficie).border(1.dp, C.linea2, forma)
            .padding(horizontal = 14.dp, vertical = 12.dp), estilo = Estilo.body16, teclado = teclado)
    }
}

/** El ícono redondo de cada opción (como las de Configuración del panel). */
@Composable
private fun IconoOpcion(simbolo: String) {
    Box(Modifier.size(40.dp).clip(RoundedCornerShape(12.dp)).background(C.crema), contentAlignment = Alignment.Center) {
        Icono(simbolo, 16f, C.primarioTexto)
    }
}
