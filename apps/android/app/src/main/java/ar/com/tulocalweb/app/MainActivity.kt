package ar.com.tulocalweb.app

import android.Manifest
import android.content.Intent
import android.content.pm.PackageManager
import android.net.Uri
import android.os.Build
import android.os.Bundle
import androidx.activity.SystemBarStyle
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.activity.result.contract.ActivityResultContracts
import androidx.biometric.BiometricManager
import androidx.biometric.BiometricManager.Authenticators.BIOMETRIC_STRONG
import androidx.biometric.BiometricManager.Authenticators.BIOMETRIC_WEAK
import androidx.biometric.BiometricManager.Authenticators.DEVICE_CREDENTIAL
import androidx.biometric.BiometricPrompt
import androidx.browser.customtabs.CustomTabColorSchemeParams
import androidx.browser.customtabs.CustomTabsIntent
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.snapshotFlow
import androidx.core.content.ContextCompat
import androidx.fragment.app.FragmentActivity
import androidx.lifecycle.lifecycleScope
import ar.com.tulocalweb.app.app.EnPrimerPlano
import ar.com.tulocalweb.app.app.Estado
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.app.Raiz
import ar.com.tulocalweb.app.app.Ruta
import ar.com.tulocalweb.app.app.TuLocalWebApp
import ar.com.tulocalweb.app.nucleo.Clerk
import ar.com.tulocalweb.app.pantallas.mas.ConectarCuenta
import ar.com.tulocalweb.app.nucleo.intentar
import ar.com.tulocalweb.app.tema.Haptico
import ar.com.tulocalweb.app.tema.TemaTLW
import kotlinx.coroutines.CompletableDeferred
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch
import kotlinx.serialization.Serializable
import java.lang.ref.WeakReference

/** La única actividad (el WindowGroup de TuLocalWebApp.swift y el Delegado de App/Push.swift): arma el Estado, recibe
 *  los links, la vuelta de Clerk y de conectar cuentas, el permiso de avisos y la huella, y vuelve a bloquear la app
 *  si estuvo más de un minuto afuera con la huella prendida. Es FragmentActivity por BiometricPrompt. */
class MainActivity : FragmentActivity() {
    private val estado: Estado get() = TuLocalWebApp.estado
    private var huella: BiometricPrompt? = null
    // El resultado del permiso llega a la actividad nueva si esta se recreó: lo que se espera vive en el proceso.
    private val pedirPermiso = registerForActivityResult(ActivityResultContracts.RequestPermission()) { ok -> permiso?.complete(ok); permiso = null }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge(statusBarStyle = SystemBarStyle.auto(android.graphics.Color.TRANSPARENT, android.graphics.Color.TRANSPARENT))
        Haptico.vista = WeakReference(window.decorView)
        estado.abrirPestana = { url -> abrirPestana(url) }
        ConectarCuenta.abrirPestana = { url -> abrirPestana(url) }
        estado.alPedirPermisoPush = { pedirPermisoAvisos() }
        estado.alPedirPermiso = { p -> pedir(p) }
        estado.alConfirmarIdentidad = { motivo -> confirmar(motivo) }
        estado.hayBiometria = { BiometricManager.from(this).canAuthenticate(BIOMETRIC_STRONG or BIOMETRIC_WEAK) == BiometricManager.BIOMETRIC_SUCCESS }
        // Con la app abierta, un aviso que llega pone al día el resumen y las listas (willPresent en iOS).
        EnPrimerPlano.alAviso = { lifecycleScope.launch { estado.refrescarResumen(); estado.version += 1 } }
        // Recreada (tema, letra, pantalla dividida) o restaurada: el intent ya se atendió.
        if (savedInstanceState == null) atender(intent)
        setContent {
            TemaTLW(estado.tema) {
                CompositionLocalProvider(LocalEstado provides estado) { Raiz(estado) }
            }
        }
    }

    override fun onDestroy() {
        // La huella a medias no queda colgada: la pantalla de bloqueo la vuelve a pedir.
        huella?.let { runCatching { it.cancelAuthentication() } }
        huella = null
        identidad?.complete(false); identidad = null
        if (!isChangingConfigurations) { permiso?.complete(false); permiso = null }
        super.onDestroy()
    }

    override fun onNewIntent(intent: Intent) {
        super.onNewIntent(intent)
        atender(intent)
    }

    override fun onStop() {
        super.onStop()
        if (!isChangingConfigurations) enSegundoPlano = System.currentTimeMillis()
    }

    override fun onStart() {
        super.onStart()
        // Con la huella prendida, volver después de un minuto pide identificarse.
        val d = enSegundoPlano
        if (estado.conFaceID && estado.fase == Estado.Fase.adentro && d != null && System.currentTimeMillis() - d > 60_000 && !abrioPestana && !estado.abrioOtraApp) estado.bloqueada = true
        enSegundoPlano = null
        estado.abrioOtraApp = false
        if (estado.fase == Estado.Fase.adentro) lifecycleScope.launch { estado.refrescarResumen() }
    }

    override fun onResume() {
        super.onResume()
        // Se cerró la Custom Tab sin terminar: como cancelar la hoja web en iOS.
        if (abrioPestana && !conVuelta) { estado.clerkCancelado(); ConectarCuenta.cancelado() }
        abrioPestana = false
    }

    private fun abrirPestana(url: String) {
        conVuelta = false
        abrioPestana = true
        val colores = CustomTabColorSchemeParams.Builder().setToolbarColor(0xFFF7F1EA.toInt()).build()
        CustomTabsIntent.Builder().setShowTitle(true).setDefaultColorSchemeParams(colores).build().launchUrl(this, Uri.parse(url))
    }

    @Serializable private data class Cuerpo(val estado: String)

    /** La vuelta de Clerk (ar.com.tulocalweb.app://callback), la de conectar una cuenta
     *  (ar.com.tulocalweb.app://integraciones/<proveedor>) o un link tulocalweb:// (avisos y widgets). */
    private fun atender(i: Intent?) {
        val u = i?.data ?: return
        if (u.scheme == Clerk.ESQUEMA) {
            conVuelta = true
            if (u.host == "integraciones" || u.host == "importar") {
                // Si el sistema cerró la app mientras estaba la pestaña, nadie la espera: se canjea igual.
                if (!ConectarCuenta.vuelta(u.toString())) estado.tarea { ConectarCuenta.retomar(u, estado) }
            } else estado.vueltaClerk(u.getQueryParameter("rotating_token_nonce"))
            return
        }
        val r = Ruta.de(u.toString()) ?: return
        // El aviso es de otra tienda: se cambia antes de abrir la ruta.
        val tienda = i.getStringExtra("tienda")?.takeIf { it.isNotBlank() }
        i.removeExtra("tienda")
        val preparado = i.getStringExtra("accion") == "preparado" && r is Ruta.pedido
        i.removeExtra("accion")
        lifecycleScope.launch {
            // Abierta desde el aviso con la app cerrada: primero termina de cargar la sesión.
            if (tienda != null) snapshotFlow { estado.fase }.first { it != Estado.Fase.arrancando }
            if (tienda != null && estado.acceso?.tiendas?.any { it.id == tienda } == true && tienda != estado.acceso?.tiendaId) {
                if (!intentar { estado.elegirTienda(tienda) }) { estado.avisar("No pudimos cambiar de tienda. Probá de nuevo.", ok = false); return@launch }
            }
            // «Marcar como preparado» desde el aviso de una venta: se marca y se avisa; si falla, se abre el pedido.
            if (preparado && r is Ruta.pedido) {
                if (intentar { estado.api.hacer("POST", "/v1/panel/pedidos/${r.n}/estado", Cuerpo("Listo para despachar")) }) {
                    estado.version += 1
                    estado.avisar("#${r.n} listo")
                } else estado.rutaPendiente = r
                return@launch
            }
            estado.rutaPendiente = r
        }
    }

    /** Un permiso del sistema: true si ya estaba o si lo dio. */
    private suspend fun pedir(p: String): Boolean {
        if (checkSelfPermission(p) == PackageManager.PERMISSION_GRANTED) return true
        permiso?.complete(false)
        val d = CompletableDeferred<Boolean>()
        permiso = d
        runCatching { pedirPermiso.launch(p) }.onFailure { permiso = null; return false }
        return d.await()
    }

    private suspend fun pedirPermisoAvisos(): Boolean {
        if (Build.VERSION.SDK_INT < 33) return true
        return pedir(Manifest.permission.POST_NOTIFICATIONS)
    }

    /** La huella, la cara o el PIN (.deviceOwnerAuthentication): sin nada configurado, deja pasar (como iOS). */
    private suspend fun confirmar(motivo: String): Boolean {
        val permitidos = if (Build.VERSION.SDK_INT >= 30) BIOMETRIC_STRONG or BIOMETRIC_WEAK or DEVICE_CREDENTIAL else BIOMETRIC_WEAK or DEVICE_CREDENTIAL
        if (BiometricManager.from(this).canAuthenticate(permitidos) != BiometricManager.BIOMETRIC_SUCCESS) return true
        identidad?.complete(false)
        val d = CompletableDeferred<Boolean>()
        identidad = d
        val p = BiometricPrompt(this, ContextCompat.getMainExecutor(this), object : BiometricPrompt.AuthenticationCallback() {
            override fun onAuthenticationSucceeded(result: BiometricPrompt.AuthenticationResult) { d.complete(true) }
            override fun onAuthenticationError(errorCode: Int, errString: CharSequence) { d.complete(false) }
        })
        huella = p
        p.authenticate(BiometricPrompt.PromptInfo.Builder().setTitle("Tu Local Web").setSubtitle(motivo).setAllowedAuthenticators(permitidos).build())
        return try { d.await() } finally { if (identidad === d) identidad = null; if (huella === p) huella = null }
    }

    companion object {
        /** Si la Custom Tab volvió con una vuelta (si no, la persona la cerró: «cancelado» en iOS). Viven en el proceso:
         *  si la actividad se recrea con la pestaña abierta, lo que espera la vuelta no queda colgado. */
        private var conVuelta = false
        private var abrioPestana = false
        /** Cuándo pasó a segundo plano: también si se cerró con «atrás» y se volvió a abrir (el Estado sigue vivo). */
        private var enSegundoPlano: Long? = null
        private var permiso: CompletableDeferred<Boolean>? = null
        private var identidad: CompletableDeferred<Boolean>? = null
    }
}
