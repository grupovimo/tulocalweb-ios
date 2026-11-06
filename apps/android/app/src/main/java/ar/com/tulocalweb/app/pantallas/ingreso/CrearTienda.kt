package ar.com.tulocalweb.app.pantallas.ingreso

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
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
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.semantics.heading
import androidx.compose.ui.semantics.selected
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.SpanStyle
import androidx.compose.ui.text.buildAnnotatedString
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.withStyle
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.nucleo.Clerk
import ar.com.tulocalweb.app.nucleo.ErrorApi
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.abrirEnNavegador
import ar.com.tulocalweb.app.pantallas.mas.CampoTexto
import ar.com.tulocalweb.app.tema.*
import kotlinx.coroutines.launch

// Clon de Pantallas/Ingreso/CrearTienda.swift.

/** Crear la cuenta y la tienda desde el celular (POST /v1/auth/registro), como el registro de la web: tu nombre, el de
 *  la tienda, email y contraseña (8 o más), con los términos. También con Google. Después sigue el onboarding. */
@Composable
fun CrearTienda() {
    val estado = LocalEstado.current
    val cerrar = LocalCerrarHoja.current
    val ctx = LocalContext.current
    val alcance = rememberCoroutineScope()
    var nombre by remember { mutableStateOf("") }
    var tienda by remember { mutableStateOf("") }
    var email by remember { mutableStateOf("") }
    var clave by remember { mutableStateOf("") }
    var terminos by remember { mutableStateOf(false) }
    var cargando by remember { mutableStateOf(false) }
    var error by remember { mutableStateOf<String?>(null) }
    val conClerk = Clerk.fapiDe(estado.proveedores?.clerkPublicable) != null
    val valido = nombre.trim().length >= 2 && tienda.trim().length >= 2 && email.contains("@") && clave.length >= 8 && terminos

    /** Como las barras de la web: corta, débil, buena, fuerte. */
    @Composable
    fun fuerza(): Pair<String, Color> {
        val tipos = listOf(clave.any { it.isLowerCase() }, clave.any { it.isUpperCase() }, clave.any { it.isDigit() }, clave.any { !it.isLetterOrDigit() }).count { it }
        if (clave.length < 8) return "Le faltan ${8 - clave.length} caracteres" to C.errorTinta
        if (clave.length >= 12 && tipos >= 3) return "Fuerte" to C.okTinta
        return if (tipos >= 2) "Buena" to C.okTinta else "Débil: sumale números o símbolos" to C.avisoTinta
    }

    suspend fun crearCuenta() {
        error = null
        cargando = true
        try {
            estado.registrar(nombre.trim(), tienda.trim(), email.trim().lowercase(), clave)
            Haptico.exito()
            cerrar()
        } catch (e: ErrorApi) {
            Haptico.error()
            error = if (e.estado == 409) "Ese email ya tiene cuenta. Entrá con él o recuperá la contraseña." else e.detalle
        } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
            Haptico.error()
            error = ErrorApi.desconocido.detalle
        } finally { cargando = false }
    }

    suspend fun social() {
        error = null
        cargando = true
        try {
            estado.ingresarConGoogle()
            cerrar()
        } catch (e: ErrorApi) { error = e.detalle } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
            // Cancelado por la persona: nada que decir.
        } finally { cargando = false }
    }

    Column(Modifier.fillMaxWidth().imePadding().verticalScroll(rememberScrollState()).padding(horizontal = 22.dp).padding(bottom = 24.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp)) {
        Grabber()
        Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
            Texto("Creá tu tienda", Estilo.tituloHojaChico, C.texto, Modifier.semantics { heading() })
            Texto(Formato.sinViudas("Gratis para arrancar. En dos minutos tenés tu link para vender."), Estilo.body16, C.texto2)
        }
        if (conClerk) {
            BotonGoogle(!cargando) { alcance.launch { social() } }
            Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                Box(Modifier.weight(1f).height(1.dp).background(C.linea2))
                Texto("o con tu email", Estilo.texto14, C.texto2, lineas = 1)
                Box(Modifier.weight(1f).height(1.dp).background(C.linea2))
            }
        }
        CampoTexto("Tu nombre", nombre, { nombre = it })
        CampoTexto("El nombre de tu tienda", tienda, { tienda = it })
        CampoTexto("Email", email, { email = it }, teclado = KeyboardType.Email)
        Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
            CampoTexto("Contraseña (8 o más)", clave, { clave = it }, seguro = true)
            if (clave.isNotEmpty()) { val (t, c) = fuerza(); Texto(t, Estilo.footnote, c) }
        }
        Row(Modifier.fillMaxWidth().clickable { terminos = !terminos; Haptico.seleccion() }.semantics { selected = terminos },
            horizontalArrangement = Arrangement.spacedBy(10.dp), verticalAlignment = Alignment.Top) {
            Icono(if (terminos) "checkmark.square.fill" else "square", 20f, if (terminos) C.acento else C.texto3)
            val p = C.primarioTexto
            TextoRico(buildAnnotatedString {
                append("Acepto los ")
                withStyle(SpanStyle(color = p, fontWeight = FontWeight.SemiBold)) { append("Términos y condiciones") }
                append(" y la ")
                withStyle(SpanStyle(color = p, fontWeight = FontWeight.SemiBold)) { append("Política de privacidad") }
                append(".")
            }, Estilo.texto14, C.texto2)
        }
        Row(horizontalArrangement = Arrangement.spacedBy(16.dp)) {
            Texto("Leer los términos", Estilo.footnoteFuerte, C.primarioTexto, Modifier.clickable { abrirEnNavegador(ctx, "https://tulocalweb.com.ar/legales/terminos") })
            Texto("Leer la privacidad", Estilo.footnoteFuerte, C.primarioTexto, Modifier.clickable { abrirEnNavegador(ctx, "https://tulocalweb.com.ar/legales/privacidad") })
        }
        error?.let { MensajeError(it) }
        BotonPrincipal("Crear mi tienda", { alcance.launch { crearCuenta() } }, habilitado = valido && !cargando, cargando = cargando)
    }
    LaunchedEffect(Unit) { estado.cargarProveedores() }
}
