package ar.com.tulocalweb.app.pantallas.mas

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.text.selection.SelectionContainer
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
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.nucleo.ErrorApi
import ar.com.tulocalweb.app.nucleo.FechaApi
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.Tincho
import ar.com.tulocalweb.app.nucleo.abrirUrl
import ar.com.tulocalweb.app.nucleo.intentar
import ar.com.tulocalweb.app.nucleo.opcional
import ar.com.tulocalweb.app.pantallas.compartir.Compartir
import ar.com.tulocalweb.app.tema.*
import kotlinx.coroutines.launch
import kotlinx.serialization.Serializable

// Clon de Pantallas/Mas/IntegracionesMas.swift: las integraciones que estaban sólo en la web: Tincho en Telegram, medición
// (Google Analytics, Google Shopping y Meta) y las IA conectadas por MCP.

// MARK: - Tincho en Telegram

/** El bot de Tincho (apps/web/src/lib/server/telegram.ts › BOT). */
const val botTelegram = "mostradorhornero_bot"

@Composable
fun TelegramTincho() {
    val estado = LocalEstado.current
    val ctx = LocalContext.current
    var vinculando by remember { mutableStateOf(false) }
    @Serializable data class Datos(val vinculado: Boolean = false, val nombre: String? = null, val desde: FechaApi? = null)
    @Serializable data class Codigo(val codigo: String)

    suspend fun vincular() {
        vinculando = true
        try {
            val c: Codigo = estado.api.enviar("POST", "/v1/panel/telegram/codigo", emptyMap<String, String>())
            abrirUrl(ctx, "https://t.me/$botTelegram?start=${c.codigo}")
        } catch (e: ErrorApi) { estado.avisar(e.detalle, ok = false) } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
            estado.avisar("No pudimos hablar con Telegram. Probá de nuevo en un rato.", ok = false)
        } finally { vinculando = false }
    }

    PantallaSeccion("Tincho en Telegram", "Telegram", { estado.api.get<Datos>("/v1/panel/telegram") }) { d, actualizar ->
        TarjetaHeroe("Telegram", if (d.vinculado) "Vinculado" else "Sin vincular",
            if (d.vinculado) "Te avisa de cada venta y le preguntás lo que quieras${d.nombre?.let { " · $it" } ?: ""}."
            else "Tincho te avisa de cada venta y te contesta por chat, sin abrir la app.", Tincho.hablando)
        if (d.vinculado) {
            BotonSecundario("Desvincular el chat", {
                estado.tarea {
                    try { estado.api.hacer("DELETE", "/v1/panel/telegram"); estado.avisar("Desvinculamos Telegram"); actualizar() } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
                        estado.avisar("No pudimos desvincularlo", ok = false)
                    }
                }
            })
        } else {
            BotonPrincipal("Abrir Telegram y vincular", { estado.tarea { vincular() } }, habilitado = !vinculando, cargando = vinculando)
            Texto(Formato.sinViudas("Se abre el chat con Tincho: tocá Iniciar y listo. El link sirve 10 minutos."), Estilo.footnote, C.texto2)
        }
    }
}

// MARK: - Medición: Google Analytics, Google Shopping y Meta

@Serializable private data class Feeds(val gm: Boolean = false, val meta: Boolean = false)
@Serializable private data class Medicion(val ga4: String? = null, val pixel: String? = null, val feeds: Feeds = Feeds())

@Composable
fun MedicionTienda() {
    val estado = LocalEstado.current
    val ctx = LocalContext.current
    val alcance = rememberCoroutineScope()
    var ga4 by remember { mutableStateOf("") }
    var pixel by remember { mutableStateOf("") }
    var gm by remember { mutableStateOf(false) }
    var meta by remember { mutableStateOf(false) }
    var cargado by remember { mutableStateOf(false) }
    var guardando by remember { mutableStateOf(false) }
    val feed = estado.tienda?.let { "https://tulocalweb.com.ar/t/${it.slug}/feed.xml" }
    val gaOk = ga4.isEmpty() || Regex("^G-[A-Z0-9]{6,12}$").matches(ga4)
    val pixelOk = pixel.isEmpty() || Regex("^[0-9]{15,16}$").matches(pixel)

    LaunchedEffect(Unit) {
        estado.api.opcional<Medicion>("/v1/panel/integraciones/medicion")?.let { m -> ga4 = m.ga4 ?: ""; pixel = m.pixel ?: ""; gm = m.feeds.gm; meta = m.feeds.meta }
        cargado = true
    }

    suspend fun guardar() {
        guardando = true
        try {
            estado.api.hacer("PUT", "/v1/panel/integraciones/medicion", Medicion(ga4.ifEmpty { null }, pixel.ifEmpty { null }, Feeds(gm, meta)))
            Haptico.exito()
            estado.avisar("Guardamos la medición")
        } catch (e: ErrorApi) { estado.avisar(e.detalle, ok = false) } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
            estado.avisar("No pudimos guardar", ok = false)
        } finally { guardando = false }
    }

    @Composable
    fun FilaLlave(texto: String, valor: Boolean, cambiar: (Boolean) -> Unit) {
        Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
            Texto(texto, Estilo.body16, C.texto, Modifier.weight(1f))
            Llave(valor, cambiar, C.acento)
        }
    }

    ConBarraAccion({
        BotonPrincipal(if (guardando) "Guardando…" else "Guardar", { alcance.launch { guardar() } }, habilitado = cargado && !guardando && gaOk && pixelOk)
    }, Modifier.background(C.fondo).imePadding()) { abajo ->
        Column(Modifier.fillMaxSize().verticalScroll(rememberScrollState()).margenPantalla().padding(bottom = 32.dp + abajo), verticalArrangement = Arrangement.spacedBy(14.dp)) {
            Texto("Medición y catálogos", Estilo.tituloGrande, C.texto)
            if (!cargado) EsqueletoFilas(4, 0.dp)
            else {
                TarjetaSeccion("Google Analytics 4", "chart.xyaxis.line") {
                    CampoTexto("ID de medición (G-XXXXXXX)", ga4, { ga4 = it.uppercase() })
                    if (!gaOk) MensajeError("Empieza con G- y sigue con 6 a 12 letras o números.")
                }
                TarjetaSeccion("Píxel de Meta", "megaphone") {
                    CampoTexto("ID del píxel (15 o 16 números)", pixel, { pixel = it }, teclado = KeyboardType.Number)
                    if (!pixelOk) MensajeError("Son 15 o 16 números.")
                }
                TarjetaSeccion("Catálogo para Google y Meta", "list.bullet.rectangle") {
                    Texto(Formato.sinViudas("Pegá este link como feed en Google Merchant Center y en Meta Commerce Manager: precio, oferta y stock se actualizan solos cada hora."), Estilo.footnote, C.texto2)
                    if (feed != null) {
                        SelectionContainer { Texto(feed, Estilo.texto14, C.primarioTexto) }
                        BotonTonal("Copiar el link del feed", { Compartir.copiar(ctx, feed, estado, "Copiamos el link del feed") })
                    }
                    FilaLlave("Lo cargué en Google Merchant Center", gm) { gm = it }
                    FilaLlave("Lo cargué en Meta Commerce Manager", meta) { meta = it }
                }
            }
        }
    }
}

// MARK: - IA conectadas (MCP)

const val direccionMcp = "https://tulocalweb.com.ar/api/mcp"
private val permisosIA = listOf(Triple("escribir", "Cargar y editar productos", "Precios, stock y fotos"),
    Triple("responder", "Responder mensajes", "WhatsApp e Instagram, con tu tono"), Triple("plata", "Reembolsos y descuentos", "Crear cupones de descuento"))

@Composable
fun IAConectadas() {
    val estado = LocalEstado.current
    val ctx = LocalContext.current
    @Serializable data class Conexion(val id: String, val cliente: String = "", val persona: String = "", val acciones: Boolean = false, val usadaEn: FechaApi? = null)
    @Serializable data class Datos(val permisos: Map<String, Boolean> = emptyMap(), val conexiones: List<Conexion> = emptyList())
    @Serializable data class V(val valor: Boolean)

    PantallaSeccion("Tu IA", "las IA conectadas", { estado.api.get<Datos>("/v1/panel/ia") }) { d, actualizar ->
        Texto(Formato.sinViudas("Manejá Tu Local Web desde Claude, ChatGPT u otra IA: preguntale por tus ventas o pedile que cargue productos."), Estilo.body16, C.texto2)
        TarjetaSeccion("Cómo se conecta", "link") {
            Texto("1. En tu IA, agregá un conector o servidor MCP con esta dirección:", Estilo.texto14, C.texto)
            SelectionContainer { Texto(direccionMcp, Estilo.texto14, C.primarioTexto) }
            BotonTonal("Copiar la dirección", { Compartir.copiar(ctx, direccionMcp, estado, "Copiamos la dirección") })
            Texto("2. Tocá Conectar: la IA te trae a Tu Local Web para que permitas el acceso.", Estilo.texto14, C.texto)
        }
        EtiquetaSeccion("Qué puede hacer", Modifier.padding(horizontal = 4.dp).padding(top = 8.dp))
        Column(Modifier.fillMaxWidth().tarjeta()) {
            FilaPermiso("Ver ventas, stock y clientes", "Sólo lectura, siempre activo", true, fijo = true) {}
            permisosIA.forEach { (id, titulo, detalle) ->
                Separador()
                FilaPermiso(titulo, detalle, d.permisos[id] ?: false, fijo = !estado.puede("cuenta")) { v ->
                    try { estado.api.hacer("PUT", "/v1/panel/ia/permisos/$id", V(v)) } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
                        estado.avisar("No pudimos cambiar el permiso", ok = false)
                    }
                    actualizar()
                }
            }
        }
        if (d.conexiones.isNotEmpty()) {
            EtiquetaSeccion("Conectadas", Modifier.padding(horizontal = 4.dp).padding(top = 8.dp))
            GrupoFilas(d.conexiones, { it.id }) { c ->
                Row(Modifier.fillMaxWidth().padding(horizontal = 16.dp, vertical = 12.dp), verticalAlignment = Alignment.CenterVertically) {
                    Column(Modifier.weight(1f).padding(end = 6.dp), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                        Texto(c.cliente, Estilo.fila, C.texto)
                        Texto("${c.persona}${c.usadaEn?.let { " · usada ${Formato.hace(it)}" } ?: ""}", Estilo.footnote, C.texto2, lineas = 1)
                    }
                    BotonTonal("Desconectar", {
                        estado.tarea {
                            try { estado.api.hacer("DELETE", "/v1/panel/ia/conexiones/${c.id}"); estado.avisar("Desconectamos ${c.cliente}"); actualizar() }
                            catch (e: ErrorApi) { estado.avisar(e.detalle, ok = false) } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
                                estado.avisar("No pudimos desconectarla", ok = false)
                            }
                        }
                    })
                }
            }
        }
    }
}

@Composable
private fun FilaPermiso(titulo: String, detalle: String, valor: Boolean, fijo: Boolean = false, cambiar: suspend (Boolean) -> Unit) {
    val estado = LocalEstado.current
    var local by remember { mutableStateOf<Boolean?>(null) }
    Row(Modifier.fillMaxWidth().deshabilitado(fijo, 0.5f).padding(horizontal = 16.dp, vertical = 12.dp), verticalAlignment = Alignment.CenterVertically) {
        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
            Texto(titulo, Estilo.fila, C.texto)
            Texto(detalle, Estilo.footnote, C.texto2)
        }
        Llave(local ?: valor, { v -> local = v; estado.tarea { cambiar(v); local = null } }, C.acento, habilitado = !fijo)
    }
}
