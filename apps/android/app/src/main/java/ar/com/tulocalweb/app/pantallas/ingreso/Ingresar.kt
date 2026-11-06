package ar.com.tulocalweb.app.pantallas.ingreso

import androidx.compose.animation.AnimatedContent
import androidx.compose.animation.core.tween
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.togetherWith
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
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableLongStateOf
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
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.heading
import androidx.compose.ui.semantics.selected
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.SpanStyle
import androidx.compose.ui.text.buildAnnotatedString
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.withStyle
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.R
import ar.com.tulocalweb.app.app.Estado
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.nucleo.Acceso
import ar.com.tulocalweb.app.nucleo.Clerk
import ar.com.tulocalweb.app.nucleo.ErrorApi
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.Tincho
import ar.com.tulocalweb.app.nucleo.intentar
import ar.com.tulocalweb.app.nucleo.nombreDelRol
import ar.com.tulocalweb.app.pantallas.mas.CodigoIngreso
import ar.com.tulocalweb.app.tema.*
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import kotlinx.serialization.Serializable

// Clon de Pantallas/Ingreso/Ingresar.swift.

/** 1a · Splash: isotipo 132 centrado y el wordmark a 64 del borde de abajo. */
@Composable
fun Splash() {
    Box(Modifier.fillMaxSize().background(C.fondo)) {
        Imagen(R.drawable.marca, Modifier.align(Alignment.Center).size(132.dp), descripcion = "Tu Local Web")
        Imagen(wordmark, Modifier.align(Alignment.BottomCenter).navigationBarsPadding().padding(bottom = (64 - 34).dp).width(140.dp))
    }
}

/** 1b · «Entrá a tu tienda»: email y contraseña, y Google (por Clerk). Con la contraseña mal dice cuántos intentos
 *  quedan; al quinto, 1c. En Android no va «Seguir con Apple» (decisión de Nahuel, 6/11). */
@Composable
fun Ingresar(fondoDeSelector: Boolean = false) {
    val estado = LocalEstado.current
    val alcance = rememberCoroutineScope()
    var email by remember { mutableStateOf("") }
    var clave by remember { mutableStateOf("") }
    var mostrar by remember { mutableStateOf(false) }
    var error by remember { mutableStateOf<String?>(null) }
    var errorEmail by remember { mutableStateOf<String?>(null) }
    var cargando by remember { mutableStateOf(false) }
    var bloqueadoHasta by remember { mutableStateOf<Long?>(null) }
    var crear by remember { mutableStateOf(false) }
    var focoEmail by remember { mutableStateOf(false) }
    var focoClave by remember { mutableStateOf(false) }
    val pedirEmail = remember { FocusRequester() }
    val pedirClave = remember { FocusRequester() }
    val conClerk = Clerk.fapiDe(estado.proveedores?.clerkPublicable) != null

    suspend fun entrar() {
        error = null
        errorEmail = null
        val e = email.trim()
        if (!e.contains("@")) {
            errorEmail = "Escribí el email con el que entrás a tu tienda."
            Haptico.error()
            runCatching { pedirEmail.requestFocus() }
            return
        }
        if (clave.isEmpty()) {
            error = "Escribí tu contraseña."
            Haptico.error()
            runCatching { pedirClave.requestFocus() }
            return
        }
        cargando = true
        try {
            estado.ingresar(e, clave)
            Haptico.exito()
        } catch (x: ErrorApi) {
            Haptico.error()
            when {
                x.estado == 401 && x.codigo == "credenciales_invalidas" -> {
                    error = mensajeClave(x.restantes)
                    clave = ""
                    runCatching { pedirClave.requestFocus() }
                }
                x.estado == 429 -> bloqueadoHasta = System.currentTimeMillis() + (x.reintentarEn ?: 900) * 1000L
                else -> error = x.detalle
            }
        } catch (x: kotlinx.coroutines.CancellationException) { throw x } catch (x: Exception) {
            Haptico.error()
            error = ErrorApi.desconocido.detalle
        } finally { cargando = false }
    }

    @Serializable data class CuerpoRecuperar(val email: String)

    suspend fun recuperar() {
        val e = email.trim()
        if (!e.contains("@")) {
            errorEmail = "Escribí tu email y te mandamos un link para cambiar la contraseña."
            runCatching { pedirEmail.requestFocus() }
            return
        }
        try {
            estado.api.hacer("POST", "/v1/auth/recuperar", CuerpoRecuperar(e))
            estado.avisar("Si hay una cuenta con ese email, te llega el link en un minuto.")
        } catch (x: ErrorApi) { estado.avisar(x.detalle, ok = false) } catch (x: kotlinx.coroutines.CancellationException) { throw x } catch (x: Exception) {}
    }

    suspend fun social() {
        error = null
        cargando = true
        try {
            estado.ingresarConGoogle()
            Haptico.exito()
        } catch (e: Clerk.ErrorClerk) {
            if (e.codigo != "cancelado") { Haptico.error(); error = mensajeClerk(e.codigo, apple = false) }
        } catch (e: ErrorApi) {
            Haptico.error(); error = e.detalle
        } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
            Haptico.error(); error = "No pudimos completar el ingreso. Probá de nuevo."
        } finally { cargando = false }
    }

    Box(Modifier.fillMaxSize().background(C.fondo)) {
        AnimatedContent(bloqueadoHasta, transitionSpec = { fadeIn(tween(200)) togetherWith fadeOut(tween(200)) }, label = "bloqueo") { hasta ->
            if (hasta != null) {
                Bloqueado(hasta, email) { bloqueadoHasta = null }
            } else {
                Column(Modifier.fillMaxSize().statusBarsPadding().imePadding()) {
                    Column(Modifier.weight(1f).verticalScroll(rememberScrollState()).padding(horizontal = 24.dp).padding(top = 22.dp, bottom = 24.dp),
                        verticalArrangement = Arrangement.spacedBy(22.dp)) {
                        Imagen(R.drawable.marca, Modifier.offset(x = (-4).dp).size(52.dp))
                        Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
                            Texto("Entrá a tu tienda", Estilo.saludo, C.texto, Modifier.semantics { heading() })
                            Texto("Ventas, pedidos y stock, desde el mostrador.", Estilo.body16, C.texto2)
                        }
                        Column(verticalArrangement = Arrangement.spacedBy(14.dp)) {
                            Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                                Campo("Email", errorEmail != null, focoEmail) {
                                    Entrada(email, { email = it }, "tu@email.com", Modifier.weight(1f).focusRequester(pedirEmail),
                                        teclado = KeyboardType.Email, accionTeclado = ImeAction.Next, alEnviar = { runCatching { pedirClave.requestFocus() } },
                                        sinCorrector = true, alEnfocar = { focoEmail = it })
                                }
                                errorEmail?.let { MensajeError(it) }
                            }
                            Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                                Campo("Contraseña", error != null, focoClave) {
                                    Entrada(clave, { clave = it }, "", Modifier.weight(1f).focusRequester(pedirClave), secreto = !mostrar,
                                        teclado = KeyboardType.Password, accionTeclado = ImeAction.Go, alEnviar = { alcance.launch { entrar() } },
                                        sinCorrector = true, alEnfocar = { focoClave = it })
                                    Texto(if (mostrar) "Ocultar" else "Mostrar", Estilo.subheadFuerte, C.texto2, Modifier.clickable { mostrar = !mostrar }.padding(start = 8.dp))
                                }
                                error?.let { MensajeError(it) }
                            }
                        }
                        Texto("Olvidé mi contraseña", Estilo.subheadFuerte, C.primarioTexto,
                            Modifier.fillMaxWidth().offset(y = (-8).dp).clickable { alcance.launch { recuperar() } }, alinear = TextAlign.End)
                        BotonPrincipal("Entrar", { alcance.launch { entrar() } }, habilitado = !cargando, cargando = cargando)
                        if (conClerk) {
                            Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                                Box(Modifier.weight(1f).height(1.dp).background(C.linea2))
                                Texto("o", Estilo.texto14, C.texto2)
                                Box(Modifier.weight(1f).height(1.dp).background(C.linea2))
                            }
                            BotonGoogle(!cargando) { alcance.launch { social() } }
                        }
                    }
                    // safeAreaInset(edge: .bottom): «¿Todavía no tenés tienda? Creala gratis».
                    Box(Modifier.fillMaxWidth().background(C.fondo).navigationBarsPadding().padding(bottom = 4.dp).heightIn(min = 44.dp).clickable { crear = true },
                        contentAlignment = Alignment.Center) {
                        val primario = C.primarioTexto
                        TextoRico(buildAnnotatedString {
                            append("¿Todavía no tenés tienda? ")
                            withStyle(SpanStyle(color = primario, fontWeight = FontWeight.SemiBold)) { append("Creala gratis") }
                        }, Estilo.texto14, C.texto2, alinear = TextAlign.Center)
                    }
                }
            }
        }
        Hoja(crear, { crear = false }) { CrearTienda() }
        // 1d · «¿A qué tienda entrás?»: no se puede cerrar sin elegir (interactiveDismissDisabled).
        Hoja(estado.fase == Estado.Fase.eligiendoTienda, {}, altoFijo = 520.dp, estirable = true) { ElegirTienda() }
        Hoja(estado.desafio != null, { estado.desafio = null }) { CodigoIngreso() }
    }
    LaunchedEffect(Unit) { estado.cargarProveedores() }
}

/** «Seguir con Google»: blanco con borde, el logo y el texto. */
@Composable
fun BotonGoogle(habilitado: Boolean, alTocar: () -> Unit) {
    val forma = RoundedCornerShape(14.dp)
    Tocable(alTocar, Modifier.fillMaxWidth(), habilitado, hundir = false) {
        Row(Modifier.fillMaxWidth().heightIn(min = 52.dp).clip(forma).background(C.superficie).border(1.dp, C.linea2, forma),
            horizontalArrangement = Arrangement.spacedBy(10.dp, Alignment.CenterHorizontally), verticalAlignment = Alignment.CenterVertically) {
            Imagen(R.drawable.logo_google, Modifier.size(20.dp))
            Texto("Seguir con Google", Estilo.headline, C.texto)
        }
    }
}

/** «La contraseña no coincide. Te quedan 2 intentos antes del bloqueo.» La API no dice si el email tiene cuenta (para
 *  no delatarlo), así que se nombran los dos. */
fun mensajeClave(restantes: Int?): String {
    val base = "El email o la contraseña no coinciden."
    return when (restantes) {
        null -> base
        0 -> "$base Si volvés a fallar, pausamos el ingreso 15 minutos."
        1 -> "$base Te queda 1 intento antes del bloqueo."
        else -> "$base Te quedan $restantes intentos antes del bloqueo."
    }
}

/** Lo que se le dice a la persona cuando Clerk no deja entrar (antes se veía como si no pasara nada). */
fun mensajeClerk(codigo: String, apple: Boolean): String {
    val quien = if (apple) "Apple" else "Google"
    if (codigo.startsWith("sin_sesion_missing_requirements")) {
        return "Tu cuenta de $quien no compartió el email. Probá de nuevo y elegí compartirlo, o entrá con tu email."
    }
    return when (codigo) {
        "authorization_invalid", "oauth_token_invalid", "oauth_identification_claim_missing" ->
            "$quien no pudo confirmar tu cuenta. Probá de nuevo en un rato o entrá con tu email."
        "sin_clave", "clerk_no_responde" -> "No pudimos conectar con el ingreso. Revisá tu conexión y probá de nuevo."
        else -> "No pudimos completar el ingreso con $quien ($codigo). Probá de nuevo o entrá con tu email."
    }
}

/** 1c · Ingreso pausado: Tincho preocupado y la cuenta regresiva (cada segundo). */
@Composable
fun Bloqueado(hasta: Long, email: String, volver: () -> Unit) {
    val estado = LocalEstado.current
    val alcance = rememberCoroutineScope()
    var ahora by remember { mutableLongStateOf(System.currentTimeMillis()) }
    LaunchedEffect(hasta) {
        while (true) {
            ahora = System.currentTimeMillis()
            if (ahora >= hasta) { volver(); break }
            delay(1000)
        }
    }
    @Serializable data class Cuerpo(val email: String)
    Column(Modifier.fillMaxSize().background(C.fondo).statusBarsPadding().navigationBarsPadding()) {
        Column(Modifier.weight(1f).fillMaxWidth().padding(horizontal = 32.dp), horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.spacedBy(14.dp, Alignment.CenterVertically)) {
            ImagenTincho(Tincho.preocupado, Modifier.size(200.dp))
            Texto("Pausamos el ingreso por 15 minutos", Estilo.tituloBloqueo, C.texto, alinear = TextAlign.Center)
            Texto(Formato.sinViudas("Hubo 5 intentos con una contraseña equivocada. Para cuidar tu tienda, esperá o cambiala ahora."),
                Estilo.body16, C.texto2, alinear = TextAlign.Center, interlineado = 3f)
            val s = maxOf(0L, (hasta - ahora) / 1000)
            Texto("Podés volver a probar en ${s / 60}:${"%02d".format(s % 60)}", Estilo.subheadFuerte, C.avisoTinta,
                Modifier.padding(top = 6.dp).clip(Capsula).background(C.avisoFondo).padding(horizontal = 14.dp, vertical = 8.dp), numeros = true)
        }
        Column(Modifier.padding(horizontal = 24.dp).padding(bottom = 16.dp), verticalArrangement = Arrangement.spacedBy(10.dp)) {
            BotonPrincipal("Cambiar contraseña", {
                alcance.launch {
                    intentar { estado.api.hacer("POST", "/v1/auth/recuperar", Cuerpo(email)) }
                    estado.avisar("Te mandamos un link a $email para cambiarla.")
                }
            })
            BotonTexto("Volver", volver)
        }
    }
}

/** 1d · «¿A qué tienda entrás?»: hoja de altura fija con las tiendas de la cuenta. */
@Composable
fun ElegirTienda(alElegir: (() -> Unit)? = null) {
    val estado = LocalEstado.current
    val alcance = rememberCoroutineScope()
    var elegida by remember { mutableStateOf<String?>(estado.acceso?.tiendaId) }
    var cargando by remember { mutableStateOf(false) }
    val tiendas = estado.acceso?.tiendas ?: emptyList()
    val elegidaId = elegida ?: estado.acceso?.tiendaId ?: ""

    suspend fun entrar() {
        cargando = true
        try {
            estado.elegirTienda(elegidaId)
            if (alElegir != null) alElegir() else estado.seguirDespuesDeElegir()
        } catch (e: ErrorApi) { estado.avisar(e.detalle, ok = false) } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
        } finally { cargando = false }
    }

    Column(Modifier.fillMaxSize().padding(horizontal = 20.dp)) {
        Grabber()
        Column(Modifier.fillMaxWidth().padding(horizontal = 4.dp).padding(top = 22.dp, bottom = 18.dp), verticalArrangement = Arrangement.spacedBy(4.dp)) {
            Texto("¿A qué tienda entrás?", Estilo.tituloHoja, C.texto)
            Texto("Tenés ${tiendas.size} tiendas con esta cuenta.", Estilo.subhead, C.texto2)
        }
        Column(Modifier.weight(1f).verticalScroll(rememberScrollState())) {
            Column(Modifier.fillMaxWidth().tarjeta()) {
                tiendas.forEachIndexed { i, t ->
                    if (i > 0) Separador(76.dp)
                    FilaTienda(t, t.id == elegidaId) { elegida = t.id; Haptico.seleccion() }
                }
            }
            Texto(Formato.sinViudas("Después cambiás de tienda desde Más, tocando el nombre."), Estilo.footnote, C.texto2,
                Modifier.fillMaxWidth().padding(horizontal = 4.dp).padding(top = 12.dp))
        }
        val nombre = tiendas.firstOrNull { it.id == elegidaId }?.nombre ?: ""
        BotonPrincipal("Entrar a $nombre", { alcance.launch { entrar() } }, Modifier.padding(bottom = 12.dp), habilitado = !cargando, cargando = cargando)
    }
}

@Composable
private fun FilaTienda(t: Acceso.Tienda, activa: Boolean, alTocar: () -> Unit) {
    val (fondo, tinta) = coloresAvatar(t.nombre)
    Row(Modifier.fillMaxWidth().clickable(onClick = alTocar).semantics { selected = activa }.padding(horizontal = 16.dp, vertical = 14.dp),
        horizontalArrangement = Arrangement.spacedBy(14.dp), verticalAlignment = Alignment.CenterVertically) {
        AvatarTienda(t, 46.dp, 13.dp, fondo, tinta)
        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
            Texto(t.nombre, Estilo.headline, C.texto)
            Texto("${t.direccion} · ${nombreDelRol(t.rol)}", Estilo.texto14, C.texto2, lineas = 1)
        }
        Box(Modifier.size(24.dp), contentAlignment = Alignment.Center) {
            if (activa) {
                Box(Modifier.size(24.dp).clip(CircleShape).background(C.primarioTexto), contentAlignment = Alignment.Center) { Icono("checkmark", 11f, Color.White) }
            } else {
                Box(Modifier.size(24.dp).border(1.5.dp, tema(0x3A1D12, 0xFBF7F2, alfa = 0.2f).actual, CircleShape))
            }
        }
    }
}

/** La pantalla de la huella al volver (1g prendido). */
@Composable
fun Bloqueo() {
    val estado = LocalEstado.current
    val alcance = rememberCoroutineScope()
    Column(Modifier.fillMaxSize().background(C.fondo).statusBarsPadding().navigationBarsPadding(), horizontalAlignment = Alignment.CenterHorizontally) {
        Spacer(Modifier.weight(1f))
        Imagen(R.drawable.marca, Modifier.size(96.dp))
        Spacer(Modifier.height(22.dp))
        Texto("Entrá con tu huella o tu cara", Estilo.tituloPaso, C.texto, alinear = TextAlign.Center, modifier = Modifier.padding(horizontal = 24.dp))
        Spacer(Modifier.weight(1f))
        BotonPrincipal("Desbloquear", { alcance.launch { estado.desbloquear() } }, Modifier.padding(horizontal = 24.dp).padding(bottom = 16.dp))
    }
    LaunchedEffect(Unit) { estado.desbloquear() }
}
