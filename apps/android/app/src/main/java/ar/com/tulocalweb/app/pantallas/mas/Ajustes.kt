package ar.com.tulocalweb.app.pantallas.mas

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
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.selection.SelectionContainer
import androidx.compose.foundation.verticalScroll
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateMapOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.selected
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.input.KeyboardCapitalization
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.app.Destino
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.app.LocalNavegacion
import ar.com.tulocalweb.app.nucleo.ErrorApi
import ar.com.tulocalweb.app.nucleo.FechaApi
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.PaginaProductos
import ar.com.tulocalweb.app.nucleo.Valor
import ar.com.tulocalweb.app.nucleo.ValorJSON
import ar.com.tulocalweb.app.nucleo.Web
import ar.com.tulocalweb.app.nucleo.compartirTexto
import ar.com.tulocalweb.app.nucleo.copiar
import ar.com.tulocalweb.app.nucleo.get
import ar.com.tulocalweb.app.nucleo.objeto
import ar.com.tulocalweb.app.nucleo.opcional
import ar.com.tulocalweb.app.nucleo.texto
import ar.com.tulocalweb.app.pantallas.venta.Escaner
import ar.com.tulocalweb.app.tema.*
import kotlinx.coroutines.launch
import kotlinx.serialization.Serializable
import java.time.Instant

// Clon de Pantallas/Mas/Ajustes.swift.

// MARK: - Qué te avisamos

private data class FilaAviso(val id: String, val titulo: String, val detalle: String, val defecto: List<Boolean>)
private val canalesAvisos = listOf("panel" to "Panel", "whatsapp" to "WhatsApp", "email" to "Mail", "celular" to "Celular")
private val gruposAvisos: List<Pair<String, List<FilaAviso>>> = listOf(
    "Ventas" to listOf(
        FilaAviso("venta", "Venta nueva", "Cada pedido pagado", listOf(true, true, true, true)),
        FilaAviso("pago", "Pago pendiente o rechazado", "Transferencias y tarjetas", listOf(true, true, true, false)),
        FilaAviso("carrito", "Carrito abandonado recuperado", "Cuando Tincho cierra la venta", listOf(true, false, false, true)),
    ),
    "Operación" to listOf(
        FilaAviso("sinstock", "Sin stock", "Una venta lo dejó en 0", listOf(true, false, true, true)),
        FilaAviso("stock", "Stock bajo", "Menos de 5 unidades", listOf(true, false, true, true)),
        FilaAviso("demora", "Envío demorado", "Más de 48 h sin movimiento", listOf(true, true, false, true)),
        FilaAviso("arca", "Certificado de ARCA por vencer", "15 días antes", listOf(true, true, true, false)),
    ),
    "Cuenta y seguridad" to listOf(
        FilaAviso("sesion", "Inicio de sesión nuevo", "Desde otro dispositivo", listOf(true, true, true, false)),
        FilaAviso("cobro", "Cobro del plan", "Factura de Tu Local Web", listOf(false, false, true, false)),
        FilaAviso("novedades", "Novedades de Tu Local Web", "Una vez por mes", listOf(false, false, true, false)),
    ),
)

@Serializable private data class Prefs(val matriz: Map<String, Boolean> = emptyMap(), val silencio: Boolean = false, val resumen: ValorJSON = Valor.booleano(false))

/** Los avisos y por dónde llegan (panel, WhatsApp, mail, celular), como en Configuración › Notificaciones. */
@Composable
fun PreferenciasAvisos() {
    val estado = LocalEstado.current
    val alcance = rememberCoroutineScope()
    val matriz = remember { mutableStateMapOf<String, Boolean>() }
    var silencio by remember { mutableStateOf(false) }
    var resumen by remember { mutableStateOf<ValorJSON>(Valor.booleano(false)) }
    var cargado by remember { mutableStateOf(false) }

    suspend fun guardar() {
        try { estado.api.hacer("PUT", "/v1/panel/notificaciones/preferencias", Prefs(matriz.toMap(), silencio, resumen)) } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
            estado.avisar("No pudimos guardar los avisos", ok = false)
        }
    }

    LaunchedEffect(Unit) {
        estado.api.opcional<Prefs>("/v1/panel/notificaciones/preferencias")?.let { p -> matriz.putAll(p.matriz); silencio = p.silencio; resumen = p.resumen }
        cargado = true
    }

    Column(Modifier.fillMaxSize().background(C.fondo).verticalScroll(rememberScrollState()).margenPantalla().padding(bottom = 24.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
        Texto("Qué te avisamos", Estilo.tituloGrande, C.texto)
        if (!cargado) EsqueletoFilas(5, 0.dp)
        else {
            Row(Modifier.fillMaxWidth().tarjeta().clickable { silencio = !silencio; alcance.launch { guardar() } }.padding(horizontal = 16.dp, vertical = 10.dp),
                verticalAlignment = Alignment.CenterVertically) {
                Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                    Texto("Modo silencio", Estilo.body16, C.texto)
                    Texto("Pausa todos los avisos al celular", Estilo.footnote, C.texto2)
                }
                Llave(silencio, { silencio = it; alcance.launch { guardar() } })
            }
            gruposAvisos.forEach { (titulo, filas) ->
                EtiquetaSeccion(titulo, Modifier.padding(horizontal = 4.dp).padding(top = 8.dp))
                Column(Modifier.fillMaxWidth().tarjeta()) {
                    filas.forEachIndexed { i, f ->
                        if (i > 0) Separador()
                        Column(Modifier.fillMaxWidth().padding(14.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                            Column(verticalArrangement = Arrangement.spacedBy(2.dp)) {
                                Texto(f.titulo, Estilo.fila, C.texto)
                                Texto(f.detalle, Estilo.footnote, C.texto2)
                            }
                            Row(horizontalArrangement = Arrangement.spacedBy(6.dp)) {
                                canalesAvisos.forEachIndexed { j, c ->
                                    val on = matriz["${f.id}:${c.first}"] ?: f.defecto[j]
                                    Texto(c.second, Estilo.footnoteFuerte, if (on) C.primarioTinta else C.texto3,
                                        Modifier.clip(Capsula).background(if (on) C.crema else C.pista).clickable {
                                            matriz["${f.id}:${c.first}"] = !on
                                            alcance.launch { guardar() }
                                        }.semantics { selected = on }.padding(horizontal = 10.dp, vertical = 6.dp))
                                }
                            }
                        }
                    }
                }
            }
        }
    }
}

// MARK: - Datos del negocio y de cobro

private val camposNegocio: List<Triple<String, String, KeyboardType>> = listOf(
    Triple("nombre", "Nombre del negocio", KeyboardType.Text), Triple("whatsapp", "WhatsApp", KeyboardType.Phone), Triple("email", "Mail", KeyboardType.Email),
    Triple("direccion", "Dirección", KeyboardType.Text), Triple("horario", "Horario", KeyboardType.Text),
    Triple("alias", "Alias para transferencias", KeyboardType.Text), Triple("cbu", "CBU o CVU (22 números)", KeyboardType.Number), Triple("titular", "Titular de la cuenta", KeyboardType.Text),
)

/** Los datos del negocio y los de cobro por transferencia (alias, CBU/CVU, titular). */
@Composable
fun NegocioCobro() {
    val estado = LocalEstado.current
    val cerrar = LocalCerrarHoja.current
    val alcance = rememberCoroutineScope()
    var negocio by remember { mutableStateOf<Map<String, ValorJSON>>(emptyMap()) }
    var cargado by remember { mutableStateOf(false) }
    var guardando by remember { mutableStateOf(false) }
    val campos = remember { mutableStateMapOf<String, String>() }

    LaunchedEffect(Unit) {
        estado.api.opcional<ValorJSON>("/v1/panel/configuracion")?.get("negocio")?.objeto?.let { n ->
            negocio = n
            camposNegocio.forEach { (k, _, _) -> campos[k] = n[k]?.texto ?: "" }
        }
        cargado = true
    }

    suspend fun guardar() {
        val n = negocio.toMutableMap()
        camposNegocio.forEach { (k, _, _) ->
            var t = (campos[k] ?: "").trim()
            if (k == "cbu") t = t.filter { it.isDigit() }
            n[k] = Valor.texto(t)
        }
        listOf("razonSocial", "cuit", "condicion").forEach { k -> if (n[k] == null) n[k] = Valor.texto("") }
        guardando = true
        try {
            estado.api.hacer("PUT", "/v1/panel/configuracion/negocio", Valor.objeto(n))
            negocio = n
            Haptico.exito()
            estado.avisar("Guardamos los datos")
            cerrar()
        } catch (e: ErrorApi) { estado.avisar(e.detalle, ok = false) } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
            estado.avisar("No pudimos guardarlos", ok = false)
        } finally { guardando = false }
    }

    ConBarraAccion({
        if (cargado) BotonPrincipal(if (guardando) "Guardando…" else "Guardar", { alcance.launch { guardar() } }, habilitado = !guardando)
    }, Modifier.background(C.fondo).imePadding()) { abajo ->
        Column(Modifier.fillMaxSize().verticalScroll(rememberScrollState()).margenPantalla().padding(bottom = 32.dp + abajo), verticalArrangement = Arrangement.spacedBy(12.dp)) {
            Texto("Datos del negocio", Estilo.tituloGrande, C.texto)
            if (!cargado) EsqueletoFilas(6, 0.dp)
            else camposNegocio.forEach { (k, titulo, teclado) ->
                if (k == "alias") EtiquetaSeccion("Cobro por transferencia", Modifier.padding(horizontal = 4.dp).padding(top = 10.dp))
                Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
                    Texto(titulo, Estilo.footnoteFuerte, C.texto2)
                    val forma = RoundedCornerShape(14.dp)
                    Entrada(campos[k] ?: "", { campos[k] = it }, titulo, Modifier.fillMaxWidth().clip(forma).background(C.superficie).border(1.dp, C.linea2, forma)
                        .padding(horizontal = 14.dp, vertical = 12.dp), estilo = Estilo.body16, teclado = teclado,
                        mayusculas = if (k == "email" || k == "alias") KeyboardCapitalization.None else KeyboardCapitalization.Sentences,
                        sinCorrector = k == "email" || k == "alias" || k == "cbu")
                }
            }
        }
    }
}

// MARK: - Links de pago

/** Un link de pago por un monto, para mandar por WhatsApp. */
@Composable
fun LinksPago(montoInicial: Int? = null) {
    val estado = LocalEstado.current
    val ctx = LocalContext.current
    val alcance = rememberCoroutineScope()
    var monto by remember { mutableStateOf(montoInicial?.toString() ?: "") }
    var concepto by remember { mutableStateOf("") }
    var cuotas by remember { mutableStateOf(false) }
    var link by remember { mutableStateOf<String?>(null) }
    var creando by remember { mutableStateOf(false) }

    @Serializable data class Cuerpo(val monto: Int, val concepto: String, val cuotas: Boolean)
    @Serializable data class R(val codigo: String)

    suspend fun crear() {
        creando = true
        try {
            val r: R = estado.api.enviar("POST", "/v1/panel/links", Cuerpo(monto.filter { it.isDigit() }.toIntOrNull() ?: 0, concepto.trim(), cuotas))
            link = "${Web.base}/p/${r.codigo}"
            Haptico.exito()
        } catch (e: ErrorApi) { estado.avisar(e.detalle, ok = false) } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
            estado.avisar("No pudimos crear el link", ok = false)
        } finally { creando = false }
    }

    val forma = RoundedCornerShape(14.dp)
    Column(Modifier.fillMaxSize().background(C.fondo).imePadding().verticalScroll(rememberScrollState()).margenPantalla().padding(bottom = 32.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
        Texto("Link de pago", Estilo.tituloGrande, C.texto)
        Texto(Formato.sinViudas("Se paga con Mercado Pago: tarjeta, dinero en cuenta o efectivo. La plata llega a tu cuenta."), Estilo.subhead, C.texto2)
        Entrada(monto, { monto = it }, "Monto", Modifier.fillMaxWidth().clip(forma).background(C.superficie).border(1.dp, C.linea2, forma).padding(horizontal = 14.dp, vertical = 10.dp),
            estilo = Estilo.montoPantalla, teclado = KeyboardType.Number, numeros = true)
        Entrada(concepto, { concepto = it }, "Concepto (ej.: Seña mesa de roble)", Modifier.fillMaxWidth().clip(forma).background(C.superficie).border(1.dp, C.linea2, forma)
            .padding(horizontal = 14.dp, vertical = 12.dp), estilo = Estilo.body16)
        Box(Modifier.fillMaxWidth().tarjeta().padding(horizontal = 16.dp, vertical = 10.dp)) { Interruptor("Permitir cuotas", cuotas, { cuotas = it }) }
        link?.let { l ->
            Column(Modifier.fillMaxWidth().tarjeta().padding(16.dp), verticalArrangement = Arrangement.spacedBy(10.dp)) {
                SelectionContainer { Texto(l, Estilo.texto14, C.primarioTexto) }
                Row(horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                    BotonTonal("Copiar", { copiar(ctx, l); Haptico.exito(); estado.avisar("Copiamos el link") })
                    BotonTonal("Compartir", { compartirTexto(ctx, "Para pagar ${concepto.ifEmpty { "tu compra" }}: $l") }, simbolo = "square.and.arrow.up")
                }
            }
        }
        BotonPrincipal(if (creando) "Creando…" else (if (link == null) "Crear link" else "Crear otro"), { alcance.launch { crear() } },
            habilitado = (monto.filter { it.isDigit() }.toIntOrNull() ?: 0) > 0 && !creando)
    }
}

// MARK: - Inventario

/** Escanear un código: si el producto existe, se edita; si no, se carga con ese código. */
@Composable
fun EscanearInventario() {
    val estado = LocalEstado.current
    val nav = LocalNavegacion.current
    val alcance = rememberCoroutineScope()
    var linterna by remember { mutableStateOf(false) }
    var buscando by remember { mutableStateOf(false) }
    @Serializable data class Movimiento(val tipo: String = "", val texto: String = "", val cantidad: Int = 0, val en: FechaApi = Instant.EPOCH)
    var movimientos by remember { mutableStateOf<List<Movimiento>>(emptyList()) }

    LaunchedEffect(Unit) { estado.api.opcional<List<Movimiento>>("/v1/panel/inventario/movimientos")?.let { movimientos = it } }

    suspend fun abrir(codigo: String) {
        buscando = true
        try {
            Haptico.seleccion()
            val l = estado.api.opcional<PaginaProductos>("/v1/panel/listas/productos", listOf("q" to codigo, "cuantos" to "5"))
            val p = l?.filas?.firstOrNull { it.sku == codigo } ?: l?.filas?.firstOrNull()
            // En la pestaña en que se abrió (en iOS iba siempre a Productos y desde Más no se veía).
            nav.abrir(if (p != null) Destino.editorProducto(p.id) else Destino.nuevoProducto(codigo))
        } finally { buscando = false }
    }

    Column(Modifier.fillMaxSize().background(C.fondo).verticalScroll(rememberScrollState()).margenPantalla().padding(bottom = 24.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
        Texto("Inventario", Estilo.tituloGrande, C.texto)
        Box(Modifier.fillMaxWidth().height(260.dp).clip(RoundedCornerShape(20.dp)).background(Color.Black)) {
            Escaner(linterna, Modifier.fillMaxSize()) { c -> if (!buscando) alcance.launch { abrir(c) } }
            Box(Modifier.align(Alignment.BottomEnd).padding(12.dp).size(44.dp).clip(CircleShape).background(Color.Black.copy(alpha = 0.45f))
                .clickable { linterna = !linterna }.semantics { contentDescription = "Linterna" }, contentAlignment = Alignment.Center) {
                Icono(if (linterna) "flashlight.on.fill" else "flashlight.off.fill", 18f, Color.White)
            }
        }
        Texto(Formato.sinViudas("Apuntá al código de barras: si el producto ya existe lo abrimos, y si no, lo cargás con ese código."), Estilo.footnote, C.texto2)
        if (movimientos.isNotEmpty()) {
            EtiquetaSeccion("Últimos movimientos de stock", Modifier.padding(horizontal = 4.dp).padding(top = 8.dp))
            Column(Modifier.fillMaxWidth().tarjeta()) {
                movimientos.take(30).forEachIndexed { i, m ->
                    if (i > 0) Separador()
                    Row(Modifier.fillMaxWidth().padding(horizontal = 16.dp, vertical = 12.dp), verticalAlignment = Alignment.CenterVertically) {
                        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                            Texto(m.texto, Estilo.fila, C.texto, lineas = 2)
                            Texto(Formato.cuando(m.en), Estilo.footnote, C.texto2)
                        }
                        Texto(if (m.cantidad > 0) "+${m.cantidad}" else "${m.cantidad}", Estilo.fila, if (m.cantidad < 0) C.errorTinta else C.okTinta, numeros = true)
                    }
                }
            }
        }
    }
}
