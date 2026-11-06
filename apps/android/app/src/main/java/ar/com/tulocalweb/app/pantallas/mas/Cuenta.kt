package ar.com.tulocalweb.app.pantallas.mas

import android.content.Intent
import android.provider.Settings
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
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
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.app.BarraDe
import ar.com.tulocalweb.app.app.Destino
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.app.LocalNavegacion
import ar.com.tulocalweb.app.app.Push
import ar.com.tulocalweb.app.nucleo.ErrorApi
import ar.com.tulocalweb.app.nucleo.EstadoDosPasos
import ar.com.tulocalweb.app.nucleo.abrirUrl
import ar.com.tulocalweb.app.nucleo.opcional
import ar.com.tulocalweb.app.tema.*
import kotlinx.coroutines.launch
import kotlinx.serialization.Serializable

// Clon de Pantallas/Mas/Cuenta.swift.

/** 11h · Cuenta: la persona, seguridad (la huella, contraseña, cerrar otras sesiones), tema (Claro / Oscuro / Sistema),
 *  avisos y ayuda. */
@Composable
fun Cuenta() {
    val estado = LocalEstado.current
    val nav = LocalNavegacion.current
    val ctx = LocalContext.current
    val alcance = rememberCoroutineScope()
    var avisosPermitidos by remember { mutableStateOf<Boolean?>(null) }
    var dosPasos by remember { mutableStateOf<EstadoDosPasos?>(null) }
    var verDosPasos by remember { mutableStateOf(false) }

    fun leerAvisos() { avisosPermitidos = Push.hayPermiso(ctx) }

    LaunchedEffect(Unit) {
        leerAvisos()
        dosPasos = estado.api.opcional("/v1/auth/segundo-factor")
    }

    suspend fun cambiarFaceID(v: Boolean) {
        if (v == estado.conFaceID) return
        if (estado.confirmarIdentidad(if (v) "Para entrar a tu tienda con tu huella o tu cara" else "Para dejar de usar tu huella o tu cara")) estado.ponerFaceID(v)
    }

    @Serializable data class Cuerpo(val email: String)
    suspend fun cambiarClave() {
        val email = estado.datos?.usuario?.email ?: return
        try {
            estado.api.hacer("POST", "/v1/auth/recuperar", Cuerpo(email))
            estado.avisar("Te mandamos un link a $email para cambiarla.")
        } catch (e: ErrorApi) { estado.avisar(e.detalle, ok = false) } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {}
    }

    @Serializable data class R(val cerradas: Int = 0)
    suspend fun cerrarOtras() {
        try {
            val r: R = estado.api.enviar("DELETE", "/v1/panel/sesiones", emptyMap<String, String>())
            Haptico.exito()
            estado.avisar(if (r.cerradas == 0) "No había otras sesiones abiertas" else if (r.cerradas == 1) "Cerramos 1 sesión" else "Cerramos ${r.cerradas} sesiones")
        } catch (e: ErrorApi) { estado.avisar(e.detalle, ok = false) } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {}
    }

    @Composable
    fun Fila(texto: String, detalle: String? = null, accion: () -> Unit) {
        Row(Modifier.fillMaxWidth().clickable(onClick = accion).padding(horizontal = 16.dp, vertical = 13.dp), horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.CenterVertically) {
            Texto(texto, Estilo.body16, C.texto, Modifier.weight(1f))
            if (detalle != null) Texto(detalle, Estilo.body16, C.texto2)
            Chevron(8f)
        }
    }

    BarraDe("Cuenta")
    Column(Modifier.fillMaxSize().background(C.fondo).verticalScroll(rememberScrollState()).padding(top = 10.dp).margenPantalla().padding(bottom = 24.dp),
        verticalArrangement = Arrangement.spacedBy(8.dp)) {
        estado.datos?.let { d ->
            Row(Modifier.fillMaxWidth().tarjeta().padding(horizontal = 16.dp, vertical = 14.dp), horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
                AvatarIniciales(d.usuario.nombre, 52.dp, fondo = C.crema, tinta = C.primarioTinta)
                Column(verticalArrangement = Arrangement.spacedBy(2.dp)) {
                    Texto(d.usuario.nombre, Estilo.headline, C.texto)
                    Texto(d.usuario.email, Estilo.texto14, C.texto2)
                }
            }
        }
        EtiquetaSeccion("Seguridad", Modifier.padding(horizontal = 4.dp).padding(top = 10.dp))
        Column(Modifier.fillMaxWidth().tarjeta()) {
            if (estado.hayBiometria()) {
                Box(Modifier.padding(horizontal = 16.dp, vertical = 10.dp)) { Interruptor("Entrar con tu huella o tu cara", estado.conFaceID, { v -> alcance.launch { cambiarFaceID(v) } }) }
                Separador()
            }
            Fila("Verificación en dos pasos", dosPasos?.let { if (it.activo) "Prendida" else "Apagada" }) { verDosPasos = true }
            Separador()
            Fila("Cambiar contraseña") { alcance.launch { cambiarClave() } }
            Separador()
            Fila("Cerrar otras sesiones") { alcance.launch { cerrarOtras() } }
        }
        EtiquetaSeccion("Avisos", Modifier.padding(horizontal = 4.dp).padding(top = 10.dp))
        Column(Modifier.fillMaxWidth().tarjeta()) {
            Fila("Avisos en este teléfono", avisosPermitidos?.let { if (it) "Prendidos" else "Apagados" }) {
                if (avisosPermitidos == false && !Push.necesitaPermiso(ctx)) {
                    runCatching { ctx.startActivity(Intent(Settings.ACTION_APP_NOTIFICATION_SETTINGS).putExtra(Settings.EXTRA_APP_PACKAGE, ctx.packageName).addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)) }
                } else alcance.launch { estado.pedirPermisoPush(); leerAvisos() }
            }
            Separador()
            Fila("Qué te avisamos") { nav.mas.add(Destino.preferenciasAvisos) }
        }
        EtiquetaSeccion("Tema", Modifier.padding(horizontal = 4.dp).padding(top = 10.dp))
        Segmentado(listOf("claro" to "Claro", "oscuro" to "Oscuro", "sistema" to "Sistema"), estado.tema, { estado.ponerTema(it) })
        EtiquetaSeccion("Ayuda", Modifier.padding(horizontal = 4.dp).padding(top = 10.dp))
        Column(Modifier.fillMaxWidth().tarjeta()) {
            Fila("Artículos de ayuda") { nav.mas.add(Destino.ayuda) }
            Separador()
            Fila("Escribirle a soporte") { abrirUrl(ctx, "mailto:hola@tulocalweb.com.ar?subject=Ayuda%20desde%20la%20app") }
        }
    }
    Hoja(verDosPasos, { verDosPasos = false; estado.tarea { dosPasos = estado.api.opcional("/v1/auth/segundo-factor") } }) { DosPasos(dosPasos) }
}
