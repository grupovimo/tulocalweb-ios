package ar.com.tulocalweb.app.pantallas.mas

import androidx.compose.animation.AnimatedContent
import androidx.compose.animation.core.tween
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.togetherWith
import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ColumnScope
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
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
import androidx.compose.ui.draw.clip
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.FilterQuality
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.nucleo.ErrorApi
import ar.com.tulocalweb.app.nucleo.EstadoDosPasos
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.abrirUrl
import ar.com.tulocalweb.app.nucleo.compartirTexto
import ar.com.tulocalweb.app.nucleo.copiar
import ar.com.tulocalweb.app.tema.*
import kotlinx.coroutines.launch
import kotlinx.serialization.Serializable

// Clon de Pantallas/Mas/DosPasos.swift.

/** Hoja del ingreso con la verificación en dos pasos: la contraseña (o Google) ya estuvo bien y falta el código de
 *  seis números de la app autenticadora, o uno de recuperación. */
@Composable
fun CodigoIngreso() {
    val estado = LocalEstado.current
    val cerrar = LocalCerrarHoja.current
    val alcance = rememberCoroutineScope()
    var codigo by remember { mutableStateOf("") }
    var error by remember { mutableStateOf<String?>(null) }
    var cargando by remember { mutableStateOf(false) }
    val foco = remember { FocusRequester() }

    suspend fun entrar() {
        error = null
        cargando = true
        try {
            estado.completarSegundoFactor(codigo)
            Haptico.exito()
        } catch (e: ErrorApi) {
            Haptico.error()
            if (e.estado == 401) { estado.avisar(e.detalle, ok = false); cerrar() }
            else { error = if (e.estado == 400) "El código no es correcto. Probá con el que muestra ahora." else e.detalle; codigo = "" }
        } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
            Haptico.error()
            error = ErrorApi.desconocido.detalle
        } finally { cargando = false }
    }

    Column(Modifier.fillMaxWidth().imePadding().padding(horizontal = 20.dp), verticalArrangement = Arrangement.spacedBy(18.dp)) {
        Grabber()
        Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
            Texto("Un paso más", Estilo.tituloHoja, C.texto)
            Texto(Formato.sinViudas("Escribí el código de seis números que muestra tu app de autenticación."), Estilo.subhead, C.texto2)
        }
        Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
            CampoCodigo(codigo, { codigo = it }, error != null, foco) { alcance.launch { entrar() } }
            error?.let { MensajeError(it) }
        }
        BotonPrincipal("Entrar", { alcance.launch { entrar() } }, habilitado = !cargando && codigo.count { !it.isWhitespace() } >= 6, cargando = cargando)
        Texto(Formato.sinViudas("¿No tenés el teléfono a mano? Usá uno de tus códigos de recuperación."), Estilo.footnote, C.texto2)
    }
    LaunchedEffect(Unit) { runCatching { foco.requestFocus() } }
}

/** El campo grande del código (números, o letras para los de recuperación). */
@Composable
fun CampoCodigo(codigo: String, alCambiar: (String) -> Unit, error: Boolean = false, foco: FocusRequester, alEnviar: () -> Unit) {
    var enfocado by remember { mutableStateOf(false) }
    Campo("Código", error, enfocado) {
        Entrada(
            codigo, { alCambiar(it.take(12)) }, "123 456", Modifier.weight(1f).focusRequester(foco),
            estilo = Estilo(22f, Peso.semibold), alinear = TextAlign.Center, teclado = KeyboardType.Ascii,
            accionTeclado = ImeAction.Go, alEnviar = alEnviar, sinCorrector = true, alEnfocar = { enfocado = it }, monoespaciado = true,
        )
    }
}

/** Cuenta › Verificación en dos pasos: prenderla (con la app autenticadora en este mismo teléfono o escaneando el QR
 *  desde otro), guardar los códigos de recuperación, pedir nuevos o apagarla. */
@Composable
fun DosPasos(estadoInicial: EstadoDosPasos?) {
    val estado = LocalEstado.current
    val cerrar = LocalCerrarHoja.current
    val ctx = LocalContext.current
    val alcance = rememberCoroutineScope()

    var paso by remember { mutableStateOf("inicio") }
    var activo by remember { mutableStateOf(estadoInicial?.activo ?: false) }
    var restantes by remember { mutableStateOf(estadoInicial?.codigosRestantes ?: 0) }
    @Serializable data class Alta(val secreto: String, val uri: String)
    var alta by remember { mutableStateOf<Alta?>(null) }
    var codigo by remember { mutableStateOf("") }
    var codigos by remember { mutableStateOf<List<String>>(emptyList()) }
    var error by remember { mutableStateOf<String?>(null) }
    var cargando by remember { mutableStateOf(false) }
    val foco = remember { FocusRequester() }

    fun empezar(p: String) {
        codigo = ""
        error = null
        paso = p
        alcance.launch { kotlinx.coroutines.delay(250); runCatching { foco.requestFocus() } }
    }

    suspend fun activar() {
        error = null
        cargando = true
        try {
            alta = estado.api.enviar<Alta, Map<String, String>>("POST", "/v1/auth/segundo-factor/activar", emptyMap())
            empezar("alta")
        } catch (e: ErrorApi) { error = e.detalle } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) { error = ErrorApi.desconocido.detalle } finally { cargando = false }
    }

    @Serializable data class Cuerpo(val codigo: String)
    @Serializable data class Codigos(val codigos: List<String>)

    suspend fun conCodigo(accion: String) {
        error = null
        cargando = true
        try {
            if (accion == "desactivar") {
                estado.api.hacer("POST", "/v1/auth/segundo-factor/desactivar", Cuerpo(codigo))
                Haptico.exito()
                estado.avisar("Verificación en dos pasos apagada")
                cerrar()
                return
            }
            val r: Codigos = estado.api.enviar("POST", "/v1/auth/segundo-factor/$accion", Cuerpo(codigo))
            Haptico.exito()
            activo = true
            codigos = r.codigos
            restantes = r.codigos.size
            paso = "codigos"
        } catch (e: ErrorApi) {
            Haptico.error()
            error = if (e.estado == 400) "El código no es correcto. Probá con el que muestra ahora." else e.detalle
            codigo = ""
        } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
            Haptico.error()
            error = ErrorApi.desconocido.detalle
        } finally { cargando = false }
    }

    @Composable
    fun Titulo(t: String, d: String) {
        Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
            Texto(Formato.sinViudas(t), Estilo.tituloHoja, C.texto)
            Texto(Formato.sinViudas(d), Estilo.subhead, C.texto2)
        }
    }

    @Composable
    fun ColumnScope.CampoYBoton(texto: String, accion: suspend () -> Unit) {
        Column(verticalArrangement = Arrangement.spacedBy(14.dp)) {
            Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                CampoCodigo(codigo, { codigo = it }, error != null, foco) { alcance.launch { accion() } }
                error?.let { MensajeError(it) }
            }
            BotonPrincipal(texto, { alcance.launch { accion() } }, habilitado = !cargando && codigo.count { !it.isWhitespace() } >= 6, cargando = cargando)
        }
    }

    Column(Modifier.fillMaxWidth().imePadding().verticalScroll(rememberScrollState()).padding(horizontal = 20.dp).padding(bottom = 24.dp)) {
        Grabber()
        Spacer(Modifier.size(18.dp))
        AnimatedContent(paso, transitionSpec = { fadeIn(tween(200)) togetherWith fadeOut(tween(200)) }, label = "paso") { p ->
            Column(verticalArrangement = Arrangement.spacedBy(18.dp)) {
                when (p) {
                    "inicio" -> if (activo) {
                        Titulo("Verificación en dos pasos prendida",
                            "Cada vez que entres a tu cuenta te vamos a pedir el código de tu app de autenticación. Te quedan $restantes códigos de recuperación.")
                        BotonSecundario("Pedir códigos de recuperación nuevos", { empezar("nuevos") })
                        Texto("Apagarla", Estilo.subheadFuerte, C.errorTinta, Modifier.fillMaxWidth().clickable { empezar("apagar") }, alinear = TextAlign.Center)
                    } else {
                        Titulo("Verificación en dos pasos",
                            "Además de la contraseña, al entrar te pedimos un código que cambia cada 30 segundos en tu celular. Si alguien conoce tu contraseña, igual no puede entrar.")
                        BotonPrincipal("Prenderla", { alcance.launch { activar() } }, habilitado = !cargando, cargando = cargando)
                        error?.let { MensajeError(it) }
                    }
                    "alta" -> {
                        Titulo("Sumala a tu app de autenticación",
                            "Usá la app de Contraseñas del teléfono, Google Authenticator o la que prefieras. Después escribí el código que te muestra.")
                        alta?.let { a ->
                            BotonSecundario("Abrir en mi app de autenticación", { abrirUrl(ctx, a.uri) })
                            QRTienda.imagen(a.uri)?.let { img ->
                                Column(Modifier.fillMaxWidth(), horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(8.dp)) {
                                    Box(Modifier.clip(RoundedCornerShape(14.dp)).background(Color.White).padding(10.dp)) {
                                        Image(img, null, Modifier.size(180.dp).semantics { contentDescription = "Código QR para tu app de autenticación" }, filterQuality = FilterQuality.None)
                                    }
                                    Texto("O escanealo desde otro teléfono", Estilo.footnote, C.texto2)
                                }
                            }
                            Texto("Copiar la clave para cargarla a mano", Estilo.subheadFuerte, C.primarioTexto,
                                Modifier.fillMaxWidth().clickable { copiar(ctx, a.secreto); estado.avisar("Clave copiada") }, alinear = TextAlign.Center)
                        }
                        CampoYBoton("Prenderla") { conCodigo("confirmar") }
                    }
                    "codigos" -> {
                        Titulo("Guardá tus códigos de recuperación", "Si perdés el celular, cada uno te deja entrar una vez. Es la única vez que los vas a ver.")
                        SelectionContainer {
                            Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
                                codigos.chunked(2).forEach { par ->
                                    Row(horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                                        par.forEach { c ->
                                            Box(Modifier.weight(1f).clip(RoundedCornerShape(10.dp)).background(C.superficie).padding(vertical = 10.dp), contentAlignment = Alignment.Center) {
                                                androidx.compose.foundation.text.BasicText(c, style = TextStyle(fontFamily = FontFamily.Monospace, fontWeight = FontWeight.Medium, fontSize = 16.sp, color = C.texto))
                                            }
                                        }
                                        if (par.size == 1) Spacer(Modifier.weight(1f))
                                    }
                                }
                            }
                        }
                        BotonSecundario("Compartir o guardar", { compartirTexto(ctx, "Códigos de recuperación de Tu Local Web\n\n" + codigos.joinToString("\n")) })
                        BotonPrincipal("Ya los guardé", { cerrar() })
                    }
                    else -> {
                        if (p == "apagar") Titulo("Apagar la verificación en dos pasos", "Escribí el código de tu app o uno de recuperación.")
                        else Titulo("Códigos de recuperación nuevos", "Escribí el código de tu app. Los códigos de antes dejan de servir.")
                        CampoYBoton(if (p == "apagar") "Apagarla" else "Generarlos") { conCodigo(if (p == "apagar") "desactivar" else "codigos") }
                        Texto("Cancelar", Estilo.subheadFuerte, C.texto2, Modifier.fillMaxWidth().clickable { paso = "inicio" }, alinear = TextAlign.Center)
                    }
                }
            }
        }
    }
}
