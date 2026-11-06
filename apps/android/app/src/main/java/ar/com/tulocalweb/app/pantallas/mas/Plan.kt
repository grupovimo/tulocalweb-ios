package ar.com.tulocalweb.app.pantallas.mas

import androidx.compose.animation.AnimatedContent
import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.slideInHorizontally
import androidx.compose.animation.slideOutHorizontally
import androidx.compose.animation.togetherWith
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.semantics.selected
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.nucleo.ErrorApi
import ar.com.tulocalweb.app.nucleo.Fecha
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.Tincho
import ar.com.tulocalweb.app.tema.*
import kotlinx.coroutines.launch
import kotlinx.serialization.Serializable
import java.time.Instant
import kotlin.math.roundToInt

// Clon de Pantallas/Mas/Plan.swift: el plan de Tu Local Web (§56), igual que Plan y facturación del panel: elegir un
// plan no cobra nada; el resumen muestra qué se paga hoy (línea por línea), desde cuándo rige y cuánto se cobra después,
// y recién «Confirmar» pasa por Mercado Pago. Las cuentas vienen hechas de la API (abono.Cotizar).

/** Los planes (apps/web/src/lib/panel/planes.ts): nombre, para quién y qué incluye. Los precios salen de las
 *  cotizaciones de la API. */
object CatalogoPlanes {
    data class Plan(val id: String, val nombre: String, val lema: String, val mensual: Int, val incluye: List<String>)
    val todos = listOf(
        Plan("gratis", "Gratis", "Para probar y empezar", 0, listOf("Productos ilimitados", "Tincho: 100 respuestas con IA", "1 persona · subdominio")),
        Plan("emprende", "Emprende", "Para vender todos los días", 19_900, listOf("Dominio propio y factura ARCA", "Todos los medios de pago y envío", "Tincho: 1.000 respuestas con IA", "2 personas")),
        Plan("pro", "Pro", "Para vender en todos tus canales", 49_900, listOf("Mercado Libre, Instagram y local", "Acciones masivas y estadísticas", "Tincho: 3.000 respuestas con IA", "5 personas con permisos")),
        Plan("negocio", "Negocio", "Para varias sucursales", 149_900, listOf("Sucursales y cajas ilimitadas", "API, webhooks y mayoristas", "Tincho: 15.000 respuestas con IA", "Personas ilimitadas")),
    )
    fun de(id: String): Plan = todos.firstOrNull { it.id == id } ?: todos[0]
    /** Por mes: el anual es 25 % menos (cobrado una vez por año). */
    fun porMes(p: Plan, anual: Boolean): Int = if (anual) (p.mensual * 0.75).roundToInt() else p.mensual
}

/** Lo que cuesta pasar a un plan y ciclo (packages/contrato: Cotizacion). */
@Serializable
data class CotizacionPlan(
    val plan: String = "", val ciclo: String = "", val tipo: String = "", val hoy: Int = 0, val rige: String = "", val hasta: String? = null,
    val cobro: Cobro? = null, val porCiclo: Int = 0, val lineas: List<Linea> = emptyList(), val pago: String = "", val inicio: String? = null,
) {
    @Serializable data class Cobro(val fecha: String, val importe: Int = 0)
    @Serializable data class Linea(val texto: String, val importe: Int = 0)

    val sinCambio: Boolean get() = tipo == "actual" || tipo == "programado"
    val pasaPorMercadoPago: Boolean get() = pago == "suscripcion" || pago == "diferencia"
}

@Serializable
data class DatosPlan(
    val actual: String = "gratis", val ciclo: String = "mes", val enPrueba: Boolean = false, val diasPrueba: Int = 0, val finPrueba: String? = null,
    val proximoCobro: String? = null, val abono: Abono? = null, val cotizaciones: Map<String, CotizacionPlan>? = null, val beneficio: Beneficio? = null,
    val beneficioUsado: Boolean? = null, val uso: Uso = Uso(), val medio: Medio? = null, val facturas: List<Factura> = emptyList(),
) {
    @Serializable data class Uso(val productos: Int = 0, val personas: Int = 0, val hornero: Int = 0)
    @Serializable data class Medio(val marca: String = "", val ultimos: String = "", val vence: String = "")
    @Serializable data class Factura(val numero: String, val detalle: String = "", val total: Int = 0, val fecha: String? = null)
    @Serializable data class Programado(val plan: String, val ciclo: String = "", val desde: String = "")
    @Serializable data class Abono(val activo: Boolean = false, val proximo: String? = null, val programado: Programado? = null)
    /** El descuento para quedarse (§57) que queda en los próximos cobros. */
    @Serializable data class Beneficio(val pct: Int, val cobros: Int)

    fun cotizacion(plan: String, ciclo: String): CotizacionPlan? = cotizaciones?.get("$plan/${if (plan == "gratis") this.ciclo else ciclo}")

    /** Lo que se debita por ciclo con el plan vigente, sin descuento. */
    val debito: Int get() = cotizacion(actual, ciclo)?.porCiclo ?: 0

    companion object {
        fun conDescuento(importe: Int, b: Beneficio?): Int {
            if (b == null || b.cobros <= 0 || b.pct <= 0) return importe
            return (importe * (100 - b.pct) / 100.0).roundToInt()
        }
    }
}

// MARK: - Pantalla

@Composable
fun PlanCuenta() {
    val estado = LocalEstado.current
    var hoja by remember { mutableStateOf<String?>(null) }
    var cancelar by remember { mutableStateOf(false) }
    var datos by remember { mutableStateOf<DatosPlan?>(null) }
    var actualizarHoja by remember { mutableStateOf<(suspend () -> Unit)?>(null) }

    fun detalleHeroe(d: DatosPlan): String {
        if (d.enPrueba) return "Te quedan ${Formato.cantidad(d.diasPrueba, "día", "días")} de prueba"
        val p = d.abono?.programado
        if (p != null && p.plan == "gratis") return "Se da de baja el ${Formato.diaMes(p.desde)}"
        if (d.actual == "gratis") return "Sin cobros"
        d.proximoCobro?.let { return "Próximo cobro: ${Formato.diaMes(it)}" }
        return "Sin débito cargado"
    }

    PantallaSeccion("Plan", "tu plan", { estado.api.get<DatosPlan>("/v1/panel/cuenta/plan") }) { d, actualizar ->
        datos = d; actualizarHoja = actualizar
        val nombre = CatalogoPlanes.de(d.actual).nombre
        val ciclo = if (d.actual == "gratis") "" else (if (d.ciclo == "anio") " anual" else " mensual")
        TarjetaHeroe(if (d.enPrueba) "Estás probando gratis" else "Tu plan", "Plan $nombre$ciclo", detalleHeroe(d), Tincho.celebrando)
        d.abono?.programado?.let { p -> AvisoProgramado(p, nombre) { hoja = d.actual } }
        d.beneficio?.let { b ->
            Row(Modifier.fillMaxWidth().aparecer(0.05).clip(RoundedCornerShape(18.dp)).background(C.okFondo.copy(alpha = C.okFondo.alpha * 0.5f)).padding(14.dp),
                horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
                Box(Modifier.size(36.dp).clip(CircleShape).background(C.okFondo), contentAlignment = Alignment.Center) { Icono("percent", 18f, C.okTinta) }
                Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                    Texto("${b.pct} % off en ${if (b.cobros == 1) "tu próximo cobro" else "tus próximos ${b.cobros} cobros"}", Estilo.subheadFuerte, C.texto)
                    Texto(Formato.sinViudas("Pagás ${Formato.pesos(DatosPlan.conDescuento(d.debito, b))} en vez de ${Formato.pesos(d.debito)}. Después vuelve solo al precio del plan."),
                        Estilo.footnote, C.texto2)
                }
            }
        }
        GrillaIndicadores(listOf(
            IndicadorIcono("shippingbox.fill", "Productos", "${d.uso.productos}"),
            IndicadorIcono("person.2.fill", "Personas del equipo", "${d.uso.personas}", tono = T.okFondo to T.okTinta),
            IndicadorIcono("sparkles", "Respuestas de Tincho este mes", "${d.uso.hornero}", tono = T.avisoFondo to T.avisoTinta),
        ))
        d.medio?.let { m ->
            GrupoFilas(listOf("Medio de pago" to "${m.marca} ·· ${m.ultimos}", "Vence" to m.vence), { it.first }) { FilaDato(it.first, it.second) }
        }
        if (d.facturas.isNotEmpty()) {
            EtiquetaSeccion("Facturas", Modifier.padding(horizontal = 4.dp).padding(top = 8.dp))
            GrupoFilas(d.facturas, { it.numero }) { f -> FilaDato(f.fecha?.let { "${f.detalle} · ${Formato.diaMes(it)}" } ?: f.detalle, Formato.pesos(f.total)) }
        }
        BotonPrincipal("Cambiar de plan", { hoja = d.actual }, Modifier.padding(top = 6.dp), alto = 48.dp)
        if (d.actual != "gratis" && d.abono?.programado?.plan != "gratis") BotonTexto("Cancelar plan", { cancelar = true }, alto = 44.dp, color = C.errorTinta)
    }
    Hoja(hoja != null, { hoja = null }) { datos?.let { d -> HojaPlan(d, hoja) { actualizarHoja?.invoke() } } }
    Hoja(cancelar, { cancelar = false }) { datos?.let { d -> HojaCancelar(d) { actualizarHoja?.invoke() } } }
}

/** El cambio que rige en el próximo cobro, con «Deshacer». */
@Composable
private fun AvisoProgramado(programado: DatosPlan.Programado, actual: String, deshacer: () -> Unit) {
    val nuevo = CatalogoPlanes.de(programado.plan).nombre
    val forma = RoundedCornerShape(18.dp)
    Row(Modifier.fillMaxWidth().clip(forma).background(C.crema2).border(1.dp, C.linea2, forma).padding(14.dp), horizontalArrangement = Arrangement.spacedBy(12.dp),
        verticalAlignment = Alignment.CenterVertically) {
        Icono("calendar.badge.clock", 20f, C.primarioTexto)
        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
            Texto(if (programado.plan == "gratis") "Pasás al plan Gratis el ${Formato.diaMes(programado.desde)}" else "Pasás a $nuevo el ${Formato.diaMes(programado.desde)}",
                Estilo.subheadFuerte, C.texto)
            Texto(Formato.sinViudas("Hasta ese día seguís con $actual, que ya está pago."), Estilo.footnote, C.texto2)
        }
        BotonTonal("Deshacer", deshacer, alto = 40.dp)
    }
}

// MARK: - Elegir y confirmar

@Composable
fun HojaPlan(datos: DatosPlan, preelegido: String? = null, listo: suspend () -> Unit) {
    val estado = LocalEstado.current
    val cerrar = LocalCerrarHoja.current
    val alcance = rememberCoroutineScope()
    var plan by remember { mutableStateOf(preelegido ?: datos.actual) }
    var ciclo by remember { mutableStateOf(datos.ciclo) }
    var enviando by remember { mutableStateOf(false) }
    val cot = datos.cotizacion(plan, ciclo)
    val hayCambio = cot?.let { !it.sinCambio } ?: false

    fun boton(c: CotizacionPlan): String = when {
        c.pago == "diferencia" || (c.pago == "suscripcion" && c.hoy > 0) -> "Pagar con Mercado Pago"
        c.pago == "suscripcion" -> "Seguir a Mercado Pago"
        c.tipo == "mantener" -> "Deshacer el cambio"
        else -> if (c.plan == "gratis") "Pasar a Gratis" else "Confirmar el cambio"
    }

    @Serializable data class S(val plan: String, val ciclo: String)
    @Serializable data class U(val url: String = "")

    suspend fun confirmar(c: CotizacionPlan) {
        enviando = true
        try {
            val u: U = estado.api.enviar("POST", "/v1/panel/cuenta/abono", S(c.plan, c.ciclo))
            val nombre = CatalogoPlanes.de(c.plan).nombre
            if (u.url.isNotEmpty()) {
                // Mercado Pago: cargar la tarjeta del débito o pagar la diferencia. El plan cambia cuando avisa.
                ConectarCuenta.abrir(u.url)
                listo()
                estado.avisar("Si completaste el pago, en unos segundos tenés $nombre.")
            } else {
                listo()
                Haptico.exito()
                when (c.tipo) {
                    "baja" -> estado.avisar("Pasás a $nombre el ${Formato.diaMes(c.rige)}. Hasta ahí seguís con lo que pagaste.")
                    "gratis" -> estado.avisar(c.hasta?.let { "Pasás a Gratis el ${Formato.diaMes(it)}. No te cobramos más." } ?: "Pasaste al plan Gratis.")
                    "mantener" -> estado.avisar("Seguís con ${CatalogoPlanes.de(datos.actual).nombre}.")
                    else -> estado.avisar("Ya tenés el plan $nombre.")
                }
            }
            cerrar()
        } catch (e: ErrorApi) { Haptico.error(); estado.avisar(e.detalle, ok = false) } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
            estado.avisar("No se pudo cambiar el plan. Probá de nuevo.", ok = false)
        } finally { enviando = false }
    }

    Column(Modifier.fillMaxSize()) {
        Grabber()
        Column(Modifier.weight(1f).verticalScroll(rememberScrollState()).padding(horizontal = 20.dp).padding(top = 8.dp, bottom = 16.dp), verticalArrangement = Arrangement.spacedBy(14.dp)) {
            Texto("Elegí tu plan", Estilo.tituloHojaChico, C.texto)
            Texto(Formato.sinViudas("Nada se cobra hasta que confirmes. Precios en pesos, con IVA incluido y sin permanencia."), Estilo.subhead, C.texto2)
            Segmentado(listOf("mes" to "Por mes", "anio" to "Por año · −25 %"), ciclo, { ciclo = it })
            CatalogoPlanes.todos.forEach { p ->
                TarjetaPlan(p, ciclo == "anio", plan == p.id, datos.actual == p.id, datos.cotizacion(p.id, ciclo) != null) {
                    if (datos.cotizacion(p.id, ciclo) == null) estado.avisar("Tenés un cambio programado: deshacelo para elegir otro plan.", ok = false)
                    else { plan = p.id; Haptico.seleccion() }
                }
            }
            if (cot != null && hayCambio) ResumenCambio(cot, datos)
            else if (cot != null && cot.tipo == "programado") Texto("Ese cambio ya está programado: rige el ${Formato.diaMes(cot.rige)}.", Estilo.footnote, C.texto2)
        }
        if (cot != null && hayCambio) BarraConfirmar(cot, boton(cot), enviando) { alcance.launch { confirmar(cot) } }
    }
}

/** Una tarjeta de plan elegible. */
@Composable
private fun TarjetaPlan(plan: CatalogoPlanes.Plan, anual: Boolean, elegido: Boolean, actual: Boolean, permitido: Boolean, tocar: () -> Unit) {
    val forma = RoundedCornerShape(20.dp)
    Column(Modifier.fillMaxWidth().alpha(if (permitido) 1f else 0.5f).clip(forma).background(C.superficie).border(if (elegido) 2.dp else 1.dp, if (elegido) C.acento else C.anillo, forma)
        .clickable(onClick = tocar).semantics { selected = elegido }.padding(16.dp), verticalArrangement = Arrangement.spacedBy(10.dp)) {
        Row(verticalAlignment = Alignment.Top) {
            Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                Row(horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.CenterVertically) {
                    Texto(plan.nombre, Estilo.nombreTienda, C.texto)
                    if (actual) Texto("Tu plan", Estilo.chip, Color.White, Modifier.clip(Capsula).background(C.primario).padding(horizontal = 8.dp, vertical = 3.dp))
                }
                Texto(plan.lema, Estilo.footnote, C.texto2)
            }
            Icono(if (elegido) "checkmark.circle.fill" else "circle", 22f, if (elegido) C.primarioTexto else C.linea2)
        }
        if (plan.mensual == 0) {
            Column {
                Texto("Gratis", Estilo.montoMeta, C.texto)
                Texto("Para siempre, sin tarjeta", Estilo.footnote, C.texto3)
            }
        } else {
            val mes = CatalogoPlanes.porMes(plan, anual)
            Column {
                Row(horizontalArrangement = Arrangement.spacedBy(4.dp), verticalAlignment = Alignment.Bottom) {
                    Texto(Formato.pesos(mes), Estilo.montoMeta, C.texto)
                    Texto("/ mes", Estilo.footnote, C.texto2, Modifier.padding(bottom = 4.dp))
                }
                Texto(if (anual) "${Formato.pesos(mes * 12)} por año · ahorrás ${Formato.pesos((plan.mensual - mes) * 12)}" else "IVA incluido · sin permanencia", Estilo.footnote, C.texto3)
            }
        }
        Column(verticalArrangement = Arrangement.spacedBy(5.dp)) {
            plan.incluye.forEach { f ->
                Row(horizontalArrangement = Arrangement.spacedBy(7.dp), verticalAlignment = Alignment.CenterVertically) {
                    Icono("checkmark", 11f, C.ok)
                    Texto(f, Estilo.footnote, C.texto)
                }
            }
        }
    }
}

/** El resumen: lo de hoy (con su detalle), lo de después, desde cuándo rige y cómo se paga. */
@Composable
private fun ResumenCambio(c: CotizacionPlan, datos: DatosPlan) {
    val actual = CatalogoPlanes.de(datos.actual).nombre
    val nuevo = CatalogoPlanes.de(c.plan)
    val rige = Fecha.iso(c.rige)?.takeIf { it.isAfter(Instant.now().plusSeconds(60)) }?.let { "el ${Formato.diaMes(c.rige)}" } ?: "hoy, al confirmar"
    val nadaHoy = when (c.tipo) {
        "prueba" -> "Estás en la prueba gratis: no pagás nada hasta el ${Formato.diaMes(datos.finPrueba)}."
        "baja", "ciclo" -> "Ya pagaste hasta el ${Formato.diaMes(c.hasta)}: seguís con $actual hasta ese día."
        "gratis" -> c.hasta?.let { "Ya pagaste hasta el ${Formato.diaMes(it)}: seguís con $actual hasta ese día." } ?: "El plan Gratis no tiene cobros."
        "mantener" -> "Deshacés el cambio programado y seguís con $actual."
        else -> ""
    }
    val como = when (c.pago) {
        "suscripcion" -> "Cargás la tarjeta en Mercado Pago y queda el débito automático. Lo das de baja cuando quieras."
        "diferencia" -> "Pagás la diferencia en Mercado Pago (tarjeta, dinero en cuenta o efectivo). Tu débito se actualiza solo."
        "actualizar" -> "Actualizamos el monto de tu débito de Mercado Pago. No tenés que cargar nada."
        "cancelar" -> "Damos de baja tu débito de Mercado Pago: no te cobramos más."
        else -> ""
    }
    TarjetaSeccion("Resumen del cambio", "cart.fill") {
        Row(horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.CenterVertically) {
            Texto(actual, Estilo.label, C.texto2)
            Icono("arrow.right", 12f, C.texto3)
            Texto(nuevo.nombre + (if (c.plan == "gratis") "" else (if (c.ciclo == "anio") " anual" else " mensual")), Estilo.label, C.primarioTexto)
        }
        Box(Modifier.fillMaxWidth().height(1.dp).background(C.linea))
        Texto("HOY PAGÁS", Estilo.chip, C.texto3)
        if (c.lineas.isEmpty()) Texto(Formato.sinViudas(nadaHoy), Estilo.footnote, C.texto2)
        else c.lineas.forEach { l ->
            Row {
                Texto(l.texto, Estilo.texto14, if (l.importe < 0) C.okTinta else C.texto, Modifier.weight(1f))
                Texto((if (l.importe < 0) "−" else "") + Formato.pesos(kotlin.math.abs(l.importe)), Estilo.texto14, if (l.importe < 0) C.okTinta else C.texto, numeros = true)
            }
        }
        Row(verticalAlignment = Alignment.Bottom) {
            Texto("Total hoy", Estilo.headline, C.texto, Modifier.weight(1f))
            Texto(Formato.pesos(c.hoy), Estilo.montoMeta, C.texto, numeros = true)
        }
        Texto("IVA incluido", Estilo.footnote, C.texto3, Modifier.fillMaxWidth(), alinear = TextAlign.End)
        Box(Modifier.fillMaxWidth().height(1.dp).background(C.linea))
        Texto("DESPUÉS", Estilo.chip, C.texto3)
        val cobro = c.cobro
        if (cobro != null) {
            val desde = if (c.tipo in listOf("sube", "baja", "ciclo", "mantener")) "Desde el próximo cobro" else "Primer cobro"
            val anual = c.ciclo == "anio"
            var t = "$desde, el ${Formato.diaMes(cobro.fecha)}: ${Formato.pesos(c.porCiclo)} ${if (anual) "por año" else "por mes"}"
            if (anual) t += " (${Formato.pesos(c.porCiclo / 12)} por mes)"
            Dato("calendar", t)
            Dato("arrow.triangle.2.circlepath", "Débito automático ${if (c.ciclo == "anio") "cada año" else "cada mes"}. Sin permanencia: lo das de baja cuando quieras.")
        } else {
            Dato("calendar", "${c.hasta?.let { "Desde el ${Formato.diaMes(it)}" } ?: "Desde hoy"}: plan Gratis, sin cobros.")
        }
        Dato("clock", "${if (c.plan == "gratis") "Pasás a Gratis" else "Tenés ${nuevo.nombre}"}: $rige")
        if (como.isNotEmpty()) {
            Row(Modifier.fillMaxWidth().clip(RoundedCornerShape(14.dp)).background(C.crema2).padding(12.dp), horizontalArrangement = Arrangement.spacedBy(10.dp), verticalAlignment = Alignment.CenterVertically) {
                Box(Modifier.size(30.dp).clip(RoundedCornerShape(9.dp)).background(C.okFondo), contentAlignment = Alignment.Center) {
                    Icono(if (c.pasaPorMercadoPago) "creditcard.fill" else "checkmark.shield.fill", 15f, C.okTinta)
                }
                Texto(Formato.sinViudas(como), Estilo.footnote, C.texto2, Modifier.weight(1f))
            }
        }
        Texto("Tu Local Web nunca ve tu tarjeta.", Estilo.footnote, C.texto3)
    }
}

@Composable
private fun Dato(simbolo: String, texto: String) {
    Row(horizontalArrangement = Arrangement.spacedBy(9.dp), verticalAlignment = Alignment.Top) {
        Icono(simbolo, 13f, C.primarioTexto, Modifier.padding(top = 2.dp))
        Texto(Formato.sinViudas(texto), Estilo.footnote, C.texto2, Modifier.weight(1f))
    }
}

/** El total de hoy y el botón, siempre a mano abajo de la hoja. */
@Composable
private fun BarraConfirmar(c: CotizacionPlan, boton: String, enviando: Boolean, accion: () -> Unit) {
    Column(Modifier.fillMaxWidth().background(C.fondo)) {
        Box(Modifier.fillMaxWidth().height(1.dp).background(C.linea))
        Row(Modifier.fillMaxWidth().padding(horizontal = 20.dp).padding(top = 10.dp, bottom = 12.dp), horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
            Column {
                Texto("Hoy pagás", Estilo.footnote, C.texto2)
                Texto(Formato.pesos(c.hoy), Estilo.total, C.texto, numeros = true)
            }
            BotonPrincipal(if (enviando) "Un momento…" else boton, accion, Modifier.weight(1f), habilitado = !enviando, simbolo = if (c.pasaPorMercadoPago) "lock.fill" else null)
        }
    }
}

// MARK: - Cancelar

private val motivosBaja = listOf("Es caro para lo que vendo", "Me falta una función", "Voy a pausar el negocio", "Me voy a otra plataforma", "Otro motivo")
/** Los descuentos por motivo (abono.Ofertas de la API). */
private val beneficiosBaja = mapOf(0 to DatosPlan.Beneficio(50, 3), 4 to DatosPlan.Beneficio(50, 1))
private data class Oferta(val titulo: String, val texto: String, val simbolo: String, val boton: String)

/** Cancelar el plan como en el panel web: el motivo, la oferta para quedarse (§57: un descuento que baja el débito de
 *  verdad, una vez por tienda, o que alguien del equipo escriba) y la baja, que rige cuando termina lo pagado. */
@Composable
fun HojaCancelar(datos: DatosPlan, listo: suspend () -> Unit) {
    val estado = LocalEstado.current
    val cerrar = LocalCerrarHoja.current
    val alcance = rememberCoroutineScope()
    var paso by remember { mutableIntStateOf(0) }
    var motivo by remember { mutableIntStateOf(0) }
    var enviando by remember { mutableStateOf(false) }

    val oferta: Oferta = run {
        val debito = datos.debito
        val b = beneficiosBaja[motivo]
        if (datos.beneficioUsado != true && b != null) {
            val con = Formato.pesos(DatosPlan.conDescuento(debito, b))
            if (b.cobros == 1) Oferta("Tu próximo cobro a mitad de precio", "Si algo no anduvo bien, queremos saberlo. Mientras tanto, tu próximo cobro es de $con en vez de ${Formato.pesos(debito)}.",
                "heart.circle.fill", "Me quedo con el descuento")
            else Oferta("Te dejamos 3 cobros a mitad de precio", "Seguí con todo por $con en tus próximos 3 cobros. Después vuelve a ${Formato.pesos(debito)}.", "percent", "Quiero el descuento")
        } else when (motivo) {
            2 -> Oferta("Mejor pausala", "Pasás al plan Gratis: tu tienda queda guardada y la reactivás cuando quieras.", "pause.circle.fill", "Pasar a Gratis")
            3 -> Oferta("Antes de irte, hablemos", "Contanos qué te ofrecen en la otra plataforma: vemos cómo ayudarte a quedarte o a mudarte sin perder nada.",
                "bubble.left.and.bubble.right.fill", "Que me contacten")
            else -> Oferta("Contanos qué te falta", "Muchas funciones se pueden activar o están por salir. Una persona del equipo te escribe hoy.", "bubble.left.and.bubble.right.fill", "Que me contacten")
        }
    }

    @Serializable data class Cuerpo(val motivo: Int, val accion: String)
    @Serializable data class Respuesta(val beneficio: DatosPlan.Beneficio? = null)
    @Serializable data class S(val plan: String, val ciclo: String)
    @Serializable data class U(val url: String = "")

    suspend fun aceptar() {
        // Pausar es pasar a Gratis: la misma confirmación de la baja.
        if (motivo == 2) { paso = 2; return }
        enviando = true
        try {
            val r: Respuesta = estado.api.enviar("POST", "/v1/panel/cuenta/cancelacion", Cuerpo(motivo, "oferta"))
            listo()
            Haptico.exito()
            val b = r.beneficio
            estado.avisar(if (b != null) "¡Gracias por quedarte! ${b.pct} % off en ${if (b.cobros == 1) "tu próximo cobro" else "tus próximos ${b.cobros} cobros"}." else "Listo, una persona del equipo te escribe hoy.")
            cerrar()
        } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) { estado.avisar("No se pudo. Probá de nuevo.", ok = false) } finally { enviando = false }
    }

    suspend fun darDeBaja() {
        enviando = true
        try {
            estado.api.hacer("POST", "/v1/panel/cuenta/cancelacion", Cuerpo(motivo, "cancelar"))
            estado.api.enviar<U, S>("POST", "/v1/panel/cuenta/abono", S("gratis", datos.ciclo))
            listo()
            Haptico.exito()
            val hasta = datos.cotizacion("gratis", datos.ciclo)?.hasta
            estado.avisar(hasta?.let { "Plan cancelado. Seguís con todo hasta el ${Formato.diaMes(it)}." } ?: "Pasaste al plan Gratis.")
            cerrar()
        } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) { estado.avisar("No se pudo. Probá de nuevo.", ok = false) } finally { enviando = false }
    }

    Column(Modifier.fillMaxSize().background(C.fondo)) {
        // La barra de la hoja (NavigationStack con «Cerrar» y el título).
        Box(Modifier.fillMaxWidth().height(44.dp).padding(horizontal = 16.dp)) {
            Texto("Cerrar", Estilo.body, C.primarioTexto, Modifier.align(Alignment.CenterStart).clickable { cerrar() })
            Texto("Cancelar el plan", Estilo.headline, C.texto, Modifier.align(Alignment.Center))
        }
        Column(Modifier.weight(1f).verticalScroll(rememberScrollState()).margenPantalla().padding(vertical = 16.dp), verticalArrangement = Arrangement.spacedBy(16.dp)) {
            Row(horizontalArrangement = Arrangement.spacedBy(6.dp)) {
                repeat(3) { i -> Box(Modifier.weight(1f).height(4.dp).clip(Capsula).background(if (i <= paso) C.acento else C.pistaMeta)) }
            }
            AnimatedContent(paso, transitionSpec = { (slideInHorizontally { it } + fadeIn()) togetherWith (slideOutHorizontally { -it } + fadeOut()) }, label = "paso") { p ->
                Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
                    when (p) {
                        0 -> {
                            Texto("¿Por qué te vas?", Estilo.tituloHojaChico, C.texto)
                            Texto("Nos ayuda a mejorar.", Estilo.footnote, C.texto2)
                            Column(Modifier.fillMaxWidth().tarjeta()) {
                                motivosBaja.forEachIndexed { i, m ->
                                    if (i > 0) Separador(16.dp)
                                    Row(Modifier.fillMaxWidth().clickable { motivo = i; Haptico.seleccion() }.semantics { selected = motivo == i }.padding(horizontal = 16.dp, vertical = 13.dp),
                                        verticalAlignment = Alignment.CenterVertically) {
                                        Texto(m, Estilo.body16, C.texto, Modifier.weight(1f))
                                        Icono(if (motivo == i) "largecircle.fill.circle" else "circle", 18f, if (motivo == i) C.acento else C.texto3)
                                    }
                                }
                            }
                            BotonPrincipal("Continuar", { paso = 1 })
                            BotonTexto("Me quedo", { cerrar() })
                        }
                        1 -> {
                            val o = oferta
                            Box(Modifier.size(64.dp).clip(CircleShape).background(C.crema), contentAlignment = Alignment.Center) { Icono(o.simbolo, 30f, C.acento) }
                            Texto(o.titulo, Estilo.tituloHojaChico, C.texto)
                            Texto(Formato.sinViudas(o.texto), Estilo.body16, C.texto2)
                            BotonPrincipal(o.boton, { alcance.launch { aceptar() } }, habilitado = !enviando)
                            BotonTexto("Cancelar igual", { paso = 2 })
                        }
                        else -> {
                            val hasta = datos.cotizacion("gratis", datos.ciclo)?.hasta
                            Texto("Confirmá la cancelación", Estilo.tituloHojaChico, C.texto)
                            Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
                                Etiqueta("calendar", hasta?.let { "Seguís con todo hasta el ${Formato.diaMes(it)}" } ?: "Pasás al plan Gratis ahora")
                                Etiqueta("creditcard", "No te cobramos más")
                                Etiqueta("checkmark.shield", "Tus productos y pedidos quedan guardados")
                            }
                            BotonDestructivo("Cancelar mi plan", { alcance.launch { darDeBaja() } }, habilitado = !enviando)
                            BotonTexto("Volver", { cerrar() })
                        }
                    }
                }
            }
        }
    }
}

/** Label(texto, systemImage:) de SwiftUI. */
@Composable
private fun Etiqueta(simbolo: String, texto: String) {
    Row(horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.CenterVertically) {
        Icono(simbolo, 14f, C.texto)
        Texto(texto, Estilo.texto14, C.texto)
    }
}
