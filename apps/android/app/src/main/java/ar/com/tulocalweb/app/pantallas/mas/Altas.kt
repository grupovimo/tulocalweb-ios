package ar.com.tulocalweb.app.pantallas.mas

import android.net.Uri
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ColumnScope
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.foundation.verticalScroll
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.runtime.snapshotFlow
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.nucleo.Api
import ar.com.tulocalweb.app.nucleo.ErrorApi
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.archivoParaCompartir
import ar.com.tulocalweb.app.nucleo.compartirArchivo
import ar.com.tulocalweb.app.nucleo.intentar
import ar.com.tulocalweb.app.nucleo.opcional
import ar.com.tulocalweb.app.tema.*
import kotlinx.coroutines.CompletableDeferred
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch
import kotlinx.serialization.Serializable

// Clon de Pantallas/Mas/Altas.swift: lo que antes pedía la computadora, ahora desde el celular: conectar Mercado Pago,
// Mercado Libre y Correo Argentino, crear cupones y cargar el certificado de ARCA.

// MARK: - Conectar una cuenta (OAuth)

/** Abre el login del proveedor en la hoja web del sistema (Custom Tab); la web devuelve el código a la app
 *  (ar.com.tulocalweb.app://…) y la app lo canjea con su token. */
object ConectarCuenta {
    /** La Custom Tab (la pone MainActivity). */
    var abrirPestana: (String) -> Unit = {}
    private var esperando: CompletableDeferred<String?>? = null

    /** MainActivity: llegó la vuelta (ar.com.tulocalweb.app://integraciones/… o …://importar/tiendanube). false si nadie
     *  la esperaba (el sistema cerró la app mientras estaba la pestaña). */
    fun vuelta(url: String): Boolean = esperando?.complete(url) ?: false

    /** La vuelta llegó a una app que se volvió a abrir de cero: se canjea el código igual y se refresca lo que se ve. */
    suspend fun retomar(u: Uri, estado: ar.com.tulocalweb.app.app.Estado) {
        snapshotFlow { estado.fase }.first { it != ar.com.tulocalweb.app.app.Estado.Fase.arrancando }
        if (estado.api.token == null) return
        val codigo = u.getQueryParameter("code")
        val est = u.getQueryParameter("state")
        val proveedor = u.pathSegments.firstOrNull().orEmpty()
        val nombre = when (proveedor) { "mercadopago" -> "Mercado Pago"; "mercadolibre" -> "Mercado Libre"; "tiendanube" -> "Tiendanube"; else -> "La cuenta" }
        if (codigo != null && est != null && proveedor.matches(Regex("^[a-z]+$"))) {
            val ok = if (u.host == "importar") {
                runCatching { estado.api.enviar<Importacion, V>("POST", "/v1/panel/importaciones/$proveedor/vuelta", V(codigo, est)) }.isSuccess
            } else intentar { estado.api.hacer("POST", "/v1/panel/integraciones/$proveedor/vuelta", V(codigo, est)) }
            if (ok) estado.avisar(if (u.host == "importar") "Empezamos a traer tu tienda de $nombre." else "$nombre quedó conectada.")
        }
        estado.version += 1
    }

    /** MainActivity: la Custom Tab se cerró sin vuelta. */
    fun cancelado() { esperando?.complete(null) }

    private suspend fun abrirYEsperar(url: String): Uri? {
        val d = CompletableDeferred<String?>()
        esperando = d
        abrirPestana(url)
        val v = d.await()
        esperando = null
        return v?.let { Uri.parse(it) }
    }

    @Serializable private data class U(val url: String)
    @Serializable private data class V(val codigo: String, val estado: String)
    @Serializable private data class Conectada(val conectada: Boolean? = null)

    /** `proveedor`: mercadopago o mercadolibre. Devuelve true si quedó conectada. */
    suspend fun conectar(proveedor: String, api: Api): Boolean {
        val u = runCatching { api.enviar<U, Map<String, String>>("POST", "/v1/panel/integraciones/$proveedor/conectar", emptyMap()) }.getOrNull() ?: return false
        val vuelta = abrirYEsperar(u.url)
        // Si la vuelta no llegó a la app pero la cuenta quedó conectada igual, es un éxito.
        suspend fun yaConectada(): Boolean = api.opcional<Conectada>("/v1/panel/integraciones/$proveedor")?.conectada ?: false
        val codigo = vuelta?.getQueryParameter("code")
        val estado = vuelta?.getQueryParameter("state")
        if (codigo == null || estado == null) return yaConectada()
        return try { api.hacer("POST", "/v1/panel/integraciones/$proveedor/vuelta", V(codigo, estado)); true } catch (e: Exception) { yaConectada() }
    }

    /** Traer la tienda de Tiendanube: su login; la web devuelve el código a la app (ar.com.tulocalweb.app://importar/tiendanube)
     *  y la app lo canjea, lo que arranca la importación. */
    suspend fun importarTiendanube(api: Api): Boolean {
        val u = runCatching { api.enviar<U, Map<String, String>>("POST", "/v1/panel/importaciones/tiendanube/conectar", emptyMap()) }.getOrNull() ?: return false
        val vuelta = abrirYEsperar(u.url)
        // Si la vuelta no llegó a la app pero la importación arrancó igual, es un éxito.
        suspend fun arranco(): Boolean = api.opcional<Importacion>("/v1/panel/importaciones/actual")?.enCurso ?: false
        val codigo = vuelta?.getQueryParameter("code")
        val estado = vuelta?.getQueryParameter("state")
        if (codigo == null || estado == null) return arranco()
        return try { api.enviar<Importacion, V>("POST", "/v1/panel/importaciones/tiendanube/vuelta", V(codigo, estado)); true } catch (e: Exception) { arranco() }
    }
}

// MARK: - Piezas de formulario

@Composable
fun CampoTexto(titulo: String, texto: String, alCambiar: (String) -> Unit, modifier: Modifier = Modifier, teclado: KeyboardType = KeyboardType.Text, seguro: Boolean = false) {
    val forma = RoundedCornerShape(14.dp)
    Column(modifier.fillMaxWidth(), verticalArrangement = Arrangement.spacedBy(6.dp)) {
        Texto(titulo, Estilo.footnoteFuerte, C.texto2)
        Entrada(
            texto, alCambiar, titulo,
            Modifier.fillMaxWidth().clip(forma).background(C.superficie).border(1.dp, C.linea2, forma).padding(horizontal = 14.dp, vertical = 12.dp),
            estilo = Estilo.body16, teclado = teclado, secreto = seguro,
            sinCorrector = teclado != KeyboardType.Text,
        )
    }
}

/** Un TextEditor: varias líneas sobre `surface` con borde `line-2`. */
@Composable
fun EditorTexto(texto: String, alCambiar: (String) -> Unit, alto: Dp = 110.dp, estilo: TextStyle? = null, placeholder: String = "") {
    val forma = RoundedCornerShape(12.dp)
    Box(Modifier.fillMaxWidth().height(alto).clip(forma).background(C.superficie).border(1.dp, C.linea2, forma).padding(8.dp)) {
        if (estilo != null) {
            BasicTextField(texto, alCambiar, Modifier.fillMaxSize(), textStyle = estilo.copy(color = C.texto), cursorBrush = SolidColor(C.acento))
        } else {
            Entrada(texto, alCambiar, placeholder, Modifier.fillMaxSize(), estilo = Estilo.body16, lineas = 4, lineasMaximas = 40)
        }
    }
}

/** Hoja con título, contenido con scroll y el botón principal abajo. */
@Composable
fun HojaFormulario(titulo: String, boton: String, habilitado: Boolean, accion: suspend () -> Boolean, contenido: @Composable ColumnScope.() -> Unit) {
    val cerrar = LocalCerrarHoja.current
    val alcance = rememberCoroutineScope()
    var enviando by remember { mutableStateOf(false) }
    Column(Modifier.fillMaxSize().imePadding()) {
        Grabber()
        Column(Modifier.weight(1f).verticalScroll(rememberScrollState()).padding(horizontal = 20.dp).padding(top = 8.dp, bottom = 16.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
            Texto(titulo, Estilo.tituloHojaChico, C.texto)
            contenido()
        }
        BotonPrincipal(if (enviando) "Un momento…" else boton, {
            alcance.launch {
                enviando = true
                if (accion()) cerrar()
                enviando = false
            }
        }, Modifier.padding(horizontal = 20.dp).padding(bottom = 12.dp), habilitado = habilitado && !enviando)
    }
}

// MARK: - Correo Argentino

@Composable
fun HojaCorreo(listo: suspend () -> Unit) {
    val estado = LocalEstado.current
    var email by remember { mutableStateOf("") }
    var clave by remember { mutableStateOf("") }
    var calle by remember { mutableStateOf("") }
    var altura by remember { mutableStateOf("") }
    var cp by remember { mutableStateOf("") }
    var localidad by remember { mutableStateOf("") }
    var provincia by remember { mutableStateOf("B") }
    var peso by remember { mutableStateOf("500") }

    @Serializable data class Origen(val cp: String, val calle: String, val altura: String, val localidad: String, val provincia: String)
    @Serializable data class Paquete(val peso: Int, val alto: Int, val ancho: Int, val largo: Int)
    @Serializable data class Alta(val email: String, val clave: String, val origen: Origen, val paquete: Paquete)

    HojaFormulario("Conectar Correo Argentino", "Conectar",
        email.contains("@") && clave.isNotEmpty() && calle.isNotEmpty() && altura.isNotEmpty() && cp.length == 4 && localidad.isNotEmpty(), {
        try {
            estado.api.hacer("POST", "/v1/panel/integraciones/correo", Alta(email.trim(), clave,
                Origen(cp, calle, altura, localidad, provincia), Paquete(peso.toIntOrNull() ?: 500, 10, 20, 30)))
            estado.avisar("Correo Argentino conectado")
            listo()
            true
        } catch (e: ErrorApi) { estado.avisar(e.detalle, ok = false); false } catch (e: Exception) { false }
    }) {
        Texto(Formato.sinViudas("Con tu cuenta de MiCorreo: las etiquetas y el seguimiento salen solos."), Estilo.subhead, C.texto2)
        CampoTexto("Mail de MiCorreo", email, { email = it }, teclado = KeyboardType.Email)
        CampoTexto("Contraseña de MiCorreo", clave, { clave = it }, seguro = true)
        EtiquetaSeccion("Desde dónde despachás", Modifier.padding(top = 6.dp))
        Row(horizontalArrangement = Arrangement.spacedBy(10.dp)) {
            CampoTexto("Calle", calle, { calle = it }, Modifier.weight(1f))
            CampoTexto("Altura", altura, { altura = it }, Modifier.width(100.dp), teclado = KeyboardType.Number)
        }
        Row(horizontalArrangement = Arrangement.spacedBy(10.dp)) {
            CampoTexto("Código postal", cp, { cp = it }, Modifier.width(130.dp), teclado = KeyboardType.Number)
            CampoTexto("Localidad", localidad, { localidad = it }, Modifier.weight(1f))
        }
        // El Picker de menú fuera de un Form muestra sólo lo elegido (sin el título).
        SelectorMenu(provincias, provincia, { provincia = it }, "Provincia")
        CampoTexto("Peso habitual de un paquete (gramos)", peso, { peso = it }, teclado = KeyboardType.Number)
    }
}

val provincias = listOf("C" to "CABA", "B" to "Buenos Aires", "X" to "Córdoba", "S" to "Santa Fe", "M" to "Mendoza",
    "T" to "Tucumán", "E" to "Entre Ríos", "Q" to "Neuquén", "R" to "Río Negro", "U" to "Chubut",
    "A" to "Salta", "Y" to "Jujuy", "H" to "Chaco", "N" to "Misiones", "W" to "Corrientes",
    "G" to "Santiago del Estero", "J" to "San Juan", "D" to "San Luis", "K" to "Catamarca",
    "F" to "La Rioja", "L" to "La Pampa", "P" to "Formosa", "Z" to "Santa Cruz", "V" to "Tierra del Fuego")

// MARK: - Cupón nuevo

@Composable
fun HojaCuponNuevo(listo: suspend () -> Unit) {
    val estado = LocalEstado.current
    var codigo by remember { mutableStateOf("") }
    var tipo by remember { mutableStateOf("pct") }
    var valor by remember { mutableStateOf("10") }
    var minimo by remember { mutableStateOf("") }
    var vigencia by remember { mutableStateOf("30") }

    @Serializable data class Opciones(val hornero: Boolean = false, val combinable: Boolean = false, val unicos: Boolean = false, val link: Boolean = false)
    @Serializable data class Nuevo(
        val codigo: String, val tipo: String, val valor: Int, val alcance: String = "todo", val quien: String = "todos",
        val minimo: Int, val limite: Int = 0, val porCliente: String = "1", val vigencia: String,
        val canales: List<String> = listOf("web"), val opciones: Opciones = Opciones(),
    )

    HojaFormulario("Cupón nuevo", "Crear cupón", codigo.length >= 3 && (tipo == "envio" || (valor.toIntOrNull() ?: 0) > 0), {
        try {
            estado.api.hacer("POST", "/v1/panel/cupones", Nuevo(
                codigo.uppercase().filter { it.isLetterOrDigit() }, tipo,
                if (tipo == "envio") 0 else (valor.toIntOrNull() ?: 0), minimo = minimo.filter { it.isDigit() }.toIntOrNull() ?: 0, vigencia = vigencia))
            Haptico.exito()
            estado.avisar("Cupón creado")
            listo()
            true
        } catch (e: ErrorApi) { estado.avisar(e.detalle, ok = false); false } catch (e: Exception) { false }
    }) {
        CampoTexto("Código", codigo, { codigo = it })
        Segmentado(listOf("pct" to "%", "fijo" to "$", "envio" to "Envío gratis"), tipo, { tipo = it })
        if (tipo != "envio") CampoTexto(if (tipo == "pct") "Porcentaje de descuento" else "Monto de descuento", valor, { valor = it }, teclado = KeyboardType.Number)
        CampoTexto("Compra mínima (opcional)", minimo, { minimo = it }, teclado = KeyboardType.Number)
        EtiquetaSeccion("Vigencia", Modifier.padding(top = 4.dp))
        Segmentado(listOf("7" to "7 días", "30" to "30 días", "sin" to "Sin fin"), vigencia, { vigencia = it })
    }
}

// MARK: - ARCA

/** Alta de ARCA en el celular: los datos, el pedido de certificado para cargar en ARCA y el certificado que ARCA
 *  devuelve (desde Archivos o pegado). */
@Composable
fun HojaArca(listo: suspend () -> Unit) {
    val estado = LocalEstado.current
    val cerrar = LocalCerrarHoja.current
    val ctx = LocalContext.current
    val alcance = rememberCoroutineScope()
    var cuit by remember { mutableStateOf("") }
    var razon by remember { mutableStateOf("") }
    var condicion by remember { mutableStateOf("monotributo") }
    var puntoVenta by remember { mutableStateOf("1") }
    var pedido by remember { mutableStateOf<java.io.File?>(null) }
    var certificado by remember { mutableStateOf("") }
    var enviando by remember { mutableStateOf(false) }
    val importar = rememberLauncherForActivityResult(ActivityResultContracts.OpenDocument()) { u ->
        if (u == null) return@rememberLauncherForActivityResult
        val d = runCatching { ctx.contentResolver.openInputStream(u)?.use { it.readBytes() } }.getOrNull() ?: return@rememberLauncherForActivityResult
        val t = String(d, Charsets.UTF_8)
        certificado = if (t.contains("BEGIN")) t
        else "-----BEGIN CERTIFICATE-----\n" + android.util.Base64.encodeToString(d, android.util.Base64.DEFAULT).trim() + "\n-----END CERTIFICATE-----"
    }

    @Composable
    fun Paso(n: Int, t: String) {
        Row(Modifier.padding(top = 6.dp), horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.CenterVertically) {
            Box(Modifier.size(24.dp).clip(CircleShape).background(C.crema), contentAlignment = Alignment.Center) { Texto("$n", Estilo(13f, Peso.bold), C.primarioTinta) }
            Texto(t, Estilo.headline, C.texto)
        }
    }

    @Serializable data class Alta(val cuit: String, val razonSocial: String, val condicion: String, val puntoVenta: Int, val ambiente: String = "produccion")
    @Serializable data class R(val pedido: String)
    @Serializable data class C2(val certificado: String)

    suspend fun siguiente() {
        enviando = true
        try {
            if (pedido == null) {
                val r: R = estado.api.enviar("POST", "/v1/panel/arca/pedido", Alta(cuit.filter { it.isDigit() }, razon, condicion, puntoVenta.toIntOrNull() ?: 1))
                val u = archivoParaCompartir(ctx, "pedido-arca-${cuit.filter { it.isDigit() }}.csr")
                u.writeText(r.pedido)
                pedido = u
            } else {
                estado.api.hacer("PUT", "/v1/panel/arca/certificado", C2(certificado.trim()))
                Haptico.exito()
                estado.avisar("ARCA conectada: ya podés facturar")
                listo()
                cerrar()
            }
        } catch (e: ErrorApi) {
            estado.avisar(e.detalle, ok = false)
        } catch (e: kotlinx.coroutines.CancellationException) {
            throw e
        } catch (e: Exception) {
            estado.avisar("No pudimos guardarlo", ok = false)
        } finally { enviando = false }
    }

    Column(Modifier.fillMaxSize().imePadding()) {
        Grabber()
        Column(Modifier.weight(1f).verticalScroll(rememberScrollState()).padding(horizontal = 20.dp).padding(top = 8.dp, bottom = 16.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
            Texto("Conectar ARCA", Estilo.tituloHojaChico, C.texto)
            Paso(1, "Tus datos fiscales")
            CampoTexto("CUIT (11 números)", cuit, { cuit = it }, teclado = KeyboardType.Number)
            CampoTexto("Razón social", razon, { razon = it })
            Segmentado(listOf("monotributo" to "Monotributo", "responsable_inscripto" to "Resp. inscripto"), condicion, { condicion = it })
            CampoTexto("Punto de venta (web services)", puntoVenta, { puntoVenta = it }, teclado = KeyboardType.Number)
            val p = pedido
            if (p != null) {
                Paso(2, "Pedí el certificado en ARCA")
                Texto(Formato.sinViudas("Compartí el pedido a tu mail o a Archivos. En arca.gob.ar › Administración de certificados digitales, creá un alias, subí este archivo y bajá el certificado que te da."),
                    Estilo.texto14, C.texto2)
                BotonSecundario("Compartir el pedido (.csr)", { compartirArchivo(ctx, p, "application/pkcs10") }, simbolo = "square.and.arrow.up")
                Paso(3, "Cargá el certificado")
                BotonSecundario("Elegir el archivo .crt", { importar.launch(arrayOf("*/*")) }, simbolo = "doc")
                Texto("o pegalo acá:", Estilo.footnote, C.texto3)
                EditorTexto(certificado, { certificado = it }, estilo = TextStyle(fontFamily = FontFamily.Monospace, fontSize = 11.sp))
            }
        }
        BotonPrincipal(if (enviando) "Un momento…" else (if (pedido == null) "Generar el pedido" else "Guardar el certificado"),
            { alcance.launch { siguiente() } }, Modifier.padding(horizontal = 20.dp).padding(bottom = 12.dp),
            habilitado = !(enviando || (if (pedido == null) (cuit.filter { it.isDigit() }.length != 11 || razon.isEmpty()) else !certificado.contains("BEGIN CERTIFICATE"))))
    }
}
