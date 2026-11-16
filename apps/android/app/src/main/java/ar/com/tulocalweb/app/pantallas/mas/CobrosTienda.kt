package ar.com.tulocalweb.app.pantallas.mas

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.padding
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
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.app.Destino
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.app.LocalNavegacion
import ar.com.tulocalweb.app.nucleo.ErrorApi
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.abrirEnNavegador
import ar.com.tulocalweb.app.tema.*
import kotlinx.coroutines.launch
import kotlinx.serialization.Serializable

// Clon de Pantallas/Mas/CobrosTienda.swift.

/** Lo que se guarda (CobrosTienda del contrato). */
@Serializable
data class ConfigCobros(val cuotas: Int = 0, val cuotasDesde: Int = 30000, val impuestos: ImpuestosCobros = ImpuestosCobros()) {
    @Serializable
    data class ImpuestosCobros(val mostrar: Boolean = false)
}

/** Configuracion.cobros: lo elegido y si hay Mercado Pago conectado. */
@Serializable
data class CobrosConfiguracion(val cuotas: Int = 0, val cuotasDesde: Int = 30000,
    val impuestos: ConfigCobros.ImpuestosCobros = ConfigCobros.ImpuestosCobros(), val mercadoPago: Boolean = false)

@Serializable
data class ConfiguracionCobros(val cobros: CobrosConfiguracion? = null)

/** Las cuotas que se pueden ofrecer (0: ninguna). */
private val opcionesCuotas = listOf(0 to "Ninguna", 3 to "3", 6 to "6", 9 to "9", 12 to "12")
private const val maximoDesde = 100_000_000
/** Donde la tienda activa en Mercado Pago las cuotas sin interés que absorbe. */
private const val costosMercadoPago = "https://www.mercadopago.com.ar/costs-section"

private fun textoMonto(monto: Int) = if (monto > 0) monto.toString() else ""

/**
 * Cobros (Configuración › Cobros del panel): las cuotas sin interés (ninguna, 3, 6, 9 o 12, desde un monto) y el precio
 * sin impuestos nacionales (Ley 27.743). Las cuotas y el interruptor se guardan al toque, con «Deshacer»; el «desde»,
 * con su botón. Las reglas son las de Go (dominio/cobros): sin Mercado Pago conectado la tienda no las ofrece aunque
 * estén elegidas, y el monto va de $ 0 a $ 100.000.000.
 */
@Composable
fun CobrosTienda() {
    val estado = LocalEstado.current
    val nav = LocalNavegacion.current
    val ctx = LocalContext.current
    val alcance = rememberCoroutineScope()
    var c by remember { mutableStateOf<ConfigCobros?>(null) }
    var mercadoPago by remember { mutableStateOf(true) }
    var desde by remember { mutableStateOf("") }
    var guardando by remember { mutableStateOf(false) }

    LaunchedEffect(Unit) {
        try {
            val x = estado.api.get<ConfiguracionCobros>("/v1/panel/configuracion").cobros
            val v = ConfigCobros(x?.cuotas ?: 0, x?.cuotasDesde ?: 30000, x?.impuestos ?: ConfigCobros.ImpuestosCobros())
            c = v
            mercadoPago = x?.mercadoPago ?: false
            desde = textoMonto(v.cuotasDesde)
        } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
            estado.avisar("No pudimos traer los cobros", ok = false)
        }
    }

    val montoDesde = desde.filter { it.isDigit() }.take(12).toLongOrNull()?.coerceAtMost(Int.MAX_VALUE.toLong())?.toInt() ?: 0
    val cambioDesde = (c?.cuotas ?: 0) > 0 && montoDesde != c?.cuotasDesde

    /** Guarda al toque (se ve enseguida y vuelve atrás si falla). Con `aviso`, el toast trae «Deshacer». */
    suspend fun guardar(v: ConfigCobros, aviso: String?, deshaciendo: Boolean = false): Boolean {
        val antes = c ?: return false
        c = v
        try {
            estado.api.hacer("PUT", "/v1/panel/configuracion/cobros", v)
            if (deshaciendo) return true
            Haptico.exito()
            if (aviso != null) estado.avisar(aviso, deshacer = { alcance.launch { guardar(antes, null, deshaciendo = true) } })
            return true
        } catch (x: ErrorApi) {
            c = antes
            estado.avisar(x.detalle, ok = false)
        } catch (x: kotlinx.coroutines.CancellationException) { throw x } catch (x: Exception) {
            c = antes
            estado.avisar("No pudimos guardar los cobros", ok = false)
        }
        return false
    }

    suspend fun elegirCuotas(n: Int) {
        val v = c ?: return
        if (v.cuotas == n) return
        guardar(v.copy(cuotas = n), if (n > 0) "$n cuotas sin interés desde ${Formato.pesos(v.cuotasDesde)}" else "Tu tienda ya no ofrece cuotas sin interés")
    }

    suspend fun elegirImpuestos(m: Boolean) {
        val v = c ?: return
        if (v.impuestos.mostrar == m) return
        guardar(v.copy(impuestos = ConfigCobros.ImpuestosCobros(m)), if (m) "Precio sin impuestos a la vista" else "Tu tienda ya no muestra el precio sin impuestos")
    }

    suspend fun guardarDesde() {
        val v = c ?: return
        if (montoDesde > maximoDesde) {
            estado.avisar("El monto va de $ 0 a $ 100.000.000.", ok = false)
            return
        }
        val nuevo = v.copy(cuotasDesde = montoDesde)
        guardando = true
        try {
            if (guardar(nuevo, null)) {
                desde = textoMonto(nuevo.cuotasDesde)
                estado.avisar("Monto guardado: tu tienda ya usa el nuevo mínimo")
            }
        } finally { guardando = false }
    }

    ConBarraAccion({
        if (cambioDesde) {
            BotonPrincipal(if (guardando) "Guardando…" else "Guardar monto", { alcance.launch { guardarDesde() } }, habilitado = !guardando)
            BotonTexto("Descartar", { desde = textoMonto(c?.cuotasDesde ?: 0) })
        }
    }, Modifier.background(C.fondo).imePadding()) { abajo ->
        Column(Modifier.fillMaxSize().verticalScroll(rememberScrollState()).margenPantalla().padding(bottom = 32.dp + abajo), verticalArrangement = Arrangement.spacedBy(12.dp)) {
            Texto("Cobros", Estilo.tituloGrande, C.texto)
            Texto(Formato.sinViudas("Las cuotas sin interés y el precio sin impuestos que ve quien compra."), Estilo.subhead, C.texto2)
            val v = c
            if (v == null) EsqueletoFilas(4, 0.dp)
            else {
                // Cuotas sin interés
                Column(Modifier.fillMaxWidth().tarjeta().padding(16.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
                    Row(horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
                        IconoOpcion("creditcard")
                        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                            Texto("Cuotas sin interés", Estilo.fila, C.texto)
                            Texto(if (v.cuotas > 0) "Hasta ${v.cuotas} cuotas con Mercado Pago" else "Tu tienda no las ofrece", Estilo.footnote, C.texto2)
                        }
                    }
                    if (!mercadoPago) {
                        Column(Modifier.fillMaxWidth().clip(RoundedCornerShape(12.dp)).background(C.avisoFondo).padding(12.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                            Texto(Formato.sinViudas("Se ofrecen sólo cobrando con Mercado Pago: conectalo en Integraciones."), Estilo.footnote, C.avisoTinta)
                            BotonTonal("Ir a Integraciones", { nav.abrir(Destino.integraciones) })
                        }
                    }
                    Segmentado(opcionesCuotas, v.cuotas, { n -> alcance.launch { elegirCuotas(n) } },
                        Modifier.semantics { contentDescription = "Cuotas sin interés" })
                    if (v.cuotas > 0) {
                        Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
                            Texto("Desde $", Estilo.footnoteFuerte, C.texto2)
                            val forma = RoundedCornerShape(14.dp)
                            Entrada(desde, { desde = it }, "En cualquier compra", Modifier.fillMaxWidth().clip(forma).background(C.superficie).border(1.dp, C.linea2, forma)
                                .padding(horizontal = 14.dp, vertical = 12.dp), estilo = Estilo.body16, teclado = KeyboardType.Number)
                        }
                    }
                    Texto(Formato.sinViudas("El costo de las cuotas lo absorbe tu tienda: quien compra paga el mismo precio en todas. Activalas también en tu cuenta de Mercado Pago, en Costos y cuotas, con las mismas cuotas y el mismo monto."),
                        Estilo.footnote, C.texto2)
                    BotonTexto("Costos y cuotas en Mercado Pago", { abrirEnNavegador(ctx, costosMercadoPago) }, alto = 44.dp, simbolo = "arrow.up.right.square")
                }
                // Precio sin impuestos nacionales
                Row(Modifier.fillMaxWidth().tarjeta().padding(16.dp), horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
                    IconoOpcion("receipt")
                    Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                        Texto("Mostrar precio sin impuestos nacionales", Estilo.fila, C.texto)
                        Texto(Formato.sinViudas("Lo pide la Ley 27.743 a quien discrimina IVA. Cada producto tiene su IVA en su ficha."), Estilo.footnote, C.texto2)
                    }
                    Llave(v.impuestos.mostrar, { m -> alcance.launch { elegirImpuestos(m) } })
                }
            }
        }
    }
}
