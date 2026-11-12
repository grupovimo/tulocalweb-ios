package ar.com.tulocalweb.app.app

import android.app.Application
import androidx.compose.animation.AnimatedContent
import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.core.tween
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.togetherWith
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.compose.ui.zIndex
import ar.com.tulocalweb.app.nucleo.Guardado
import ar.com.tulocalweb.app.nucleo.Llavero
import ar.com.tulocalweb.app.nucleo.Preferencias
import ar.com.tulocalweb.app.nucleo.Recursos
import ar.com.tulocalweb.app.pantallas.ingreso.Bloqueo
import ar.com.tulocalweb.app.pantallas.ingreso.Ingresar
import ar.com.tulocalweb.app.pantallas.ingreso.Onboarding
import ar.com.tulocalweb.app.pantallas.ingreso.Permisos
import ar.com.tulocalweb.app.pantallas.ingreso.Splash
import ar.com.tulocalweb.app.tema.VistaAviso

// Clon de App/TuLocalWebApp.swift: la app (la actividad está en MainActivity) y la raíz que elige qué se ve.

/** La aplicación: deja listos el Llavero, las preferencias y lo guardado antes que cualquier pantalla, servicio de
 *  avisos o widget (en iOS es el proceso entero). */
class TuLocalWebApp : Application() {
    override fun onCreate() {
        super.onCreate()
        Llavero.contexto = applicationContext
        Preferencias.contexto = applicationContext
        Guardado.contexto = applicationContext
        Recursos.contexto = applicationContext
        Push.iniciar(applicationContext)
        estado = Estado().also { it.contexto = applicationContext }
    }

    companion object {
        /** El Estado de la app: vive lo que vive el proceso, así que girar, cambiar el tema o el tamaño de letra no
         *  vuelve al splash ni pide la huella de nuevo. */
        lateinit var estado: Estado
            private set
    }
}

/** Qué se ve según la sesión: splash, ingreso, selector de tienda, permisos o el panel. */
@Composable
fun Raiz(estado: Estado) {
    Box(Modifier.fillMaxSize()) {
        AnimatedContent(estado.fase, transitionSpec = { fadeIn(tween(250)) togetherWith fadeOut(tween(250)) }, label = "fase") { f ->
            when (f) {
                Estado.Fase.arrancando -> Splash()
                Estado.Fase.afuera -> Ingresar()
                Estado.Fase.eligiendoTienda -> Ingresar(fondoDeSelector = true)
                Estado.Fase.permisos -> Permisos()
                Estado.Fase.onboarding -> Onboarding()
                Estado.Fase.adentro -> Principal()
            }
        }
        AnimatedVisibility(estado.bloqueada, Modifier.zIndex(2f), enter = fadeIn(tween(200)), exit = fadeOut(tween(200))) { Bloqueo() }
        // Afuera del panel (la sesión venció, la cuenta se borró, falló el ingreso al volver de Google) los avisos
        // también se ven; adentro los muestra Principal.
        AnimatedVisibility(estado.fase != Estado.Fase.adentro && estado.aviso != null, Modifier.align(Alignment.BottomCenter).zIndex(3f).navigationBarsPadding()
            .padding(start = 16.dp, end = 16.dp, bottom = 24.dp), enter = fadeIn(tween(200)), exit = fadeOut(tween(200))) {
            estado.aviso?.let { a -> VistaAviso(a) { estado.aviso = null } }
        }
    }
    // Si la actividad se recrea, el Estado ya arrancó.
    LaunchedEffect(Unit) { if (estado.fase == Estado.Fase.arrancando) estado.arrancar() }
}
