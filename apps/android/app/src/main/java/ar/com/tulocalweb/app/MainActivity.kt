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
import androidx.core.content.ContextCompat
import androidx.fragment.app.FragmentActivity
import androidx.lifecycle.lifecycleScope
import ar.com.tulocalweb.app.app.EnPrimerPlano
import ar.com.tulocalweb.app.app.Estado
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.app.Raiz
import ar.com.tulocalweb.app.app.Ruta
import ar.com.tulocalweb.app.nucleo.Clerk
import ar.com.tulocalweb.app.pantallas.mas.ConectarCuenta
import ar.com.tulocalweb.app.nucleo.intentar
import ar.com.tulocalweb.app.tema.Haptico
import ar.com.tulocalweb.app.tema.TemaTLW
import kotlinx.coroutines.CompletableDeferred
import kotlinx.coroutines.launch
import kotlinx.serialization.Serializable
import java.lang.ref.WeakReference

/** La única actividad (el WindowGroup de TuLocalWebApp.swift y el Delegado de App/Push.swift): arma el Estado, recibe
 *  los links, la vuelta de Clerk y de conectar cuentas, el permiso de avisos y la huella, y vuelve a bloquear la app
 *  si estuvo más de un minuto afuera con la huella prendida. Es FragmentActivity por BiometricPrompt. */
class MainActivity : FragmentActivity() {
    private lateinit var estado: Estado
    /** Si la Custom Tab volvió con una vuelta (si no, la persona la cerró: «cancelado» en iOS). */
    private var conVuelta = false
    private var abrioPestana = false
    private var enSegundoPlano: Long? = null
    private var permiso: CompletableDeferred<Boolean>? = null
    private val pedirPermiso = registerForActivityResult(ActivityResultContracts.RequestPermission()) { ok -> permiso?.complete(ok); permiso = null }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge(statusBarStyle = SystemBarStyle.auto(android.graphics.Color.TRANSPARENT, android.graphics.Color.TRANSPARENT))
        Haptico.vista = WeakReference(window.decorView)
        estado = Estado()
        estado.contexto = applicationContext
        estado.abrirPestana = { url -> abrirPestana(url) }
        ConectarCuenta.abrirPestana = { url -> abrirPestana(url) }
        estado.alPedirPermisoPush = { pedirPermisoAvisos() }
        estado.alPedirPermiso = { p -> pedir(p) }
        estado.alConfirmarIdentidad = { motivo -> confirmar(motivo) }
        estado.hayBiometria = { BiometricManager.from(this).canAuthenticate(BIOMETRIC_STRONG or BIOMETRIC_WEAK) == BiometricManager.BIOMETRIC_SUCCESS }
        // Con la app abierta, un aviso que llega pone al día el resumen y las listas (willPresent en iOS).
        EnPrimerPlano.alAviso = { lifecycleScope.launch { estado.refrescarResumen(); estado.version += 1 } }
        atender(intent)
        setContent {
            TemaTLW(estado.tema) {
                CompositionLocalProvider(LocalEstado provides estado) { Raiz(estado) }
            }
        }
    }

    override fun onNewIntent(intent: Intent) {
        super.onNewIntent(intent)
        atender(intent)
    }

    override fun onStop() {
        super.onStop()
        enSegundoPlano = System.currentTimeMillis()
    }

    override fun onStart() {
        super.onStart()
        if (!::estado.isInitialized) return
        // Con la huella prendida, volver después de un minuto pide identificarse.
        val d = enSegundoPlano
        if (estado.conFaceID && estado.fase == Estado.Fase.adentro && d != null && System.currentTimeMillis() - d > 60_000 && !abrioPestana) estado.bloqueada = true
        enSegundoPlano = null
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
            if (u.host == "integraciones" || u.host == "importar") ConectarCuenta.vuelta(u.toString())
            else estado.vueltaClerk(u.getQueryParameter("rotating_token_nonce"))
            return
        }
        val r = Ruta.de(u.toString()) ?: return
        // «Marcar como preparado» desde el aviso de una venta: se marca y se avisa; si falla, se abre el pedido.
        if (i.getStringExtra("accion") == "preparado" && r is Ruta.pedido) {
            i.removeExtra("accion")
            lifecycleScope.launch {
                if (intentar { estado.api.hacer("POST", "/v1/panel/pedidos/${r.n}/estado", Cuerpo("Listo para despachar")) }) {
                    estado.version += 1
                    estado.avisar("#${r.n} listo")
                } else estado.rutaPendiente = r
            }
            return
        }
        estado.rutaPendiente = r
    }

    /** Un permiso del sistema: true si ya estaba o si lo dio. */
    private suspend fun pedir(p: String): Boolean {
        if (checkSelfPermission(p) == PackageManager.PERMISSION_GRANTED) return true
        val d = CompletableDeferred<Boolean>()
        permiso = d
        pedirPermiso.launch(p)
        return d.await()
    }

    private suspend fun pedirPermisoAvisos(): Boolean {
        if (Build.VERSION.SDK_INT < 33) return true
        if (checkSelfPermission(Manifest.permission.POST_NOTIFICATIONS) == PackageManager.PERMISSION_GRANTED) return true
        val d = CompletableDeferred<Boolean>()
        permiso = d
        pedirPermiso.launch(Manifest.permission.POST_NOTIFICATIONS)
        return d.await()
    }

    /** La huella, la cara o el PIN (.deviceOwnerAuthentication): sin nada configurado, deja pasar (como iOS). */
    private suspend fun confirmar(motivo: String): Boolean {
        val permitidos = if (Build.VERSION.SDK_INT >= 30) BIOMETRIC_STRONG or BIOMETRIC_WEAK or DEVICE_CREDENTIAL else BIOMETRIC_WEAK or DEVICE_CREDENTIAL
        if (BiometricManager.from(this).canAuthenticate(permitidos) != BiometricManager.BIOMETRIC_SUCCESS) return true
        val d = CompletableDeferred<Boolean>()
        val p = BiometricPrompt(this, ContextCompat.getMainExecutor(this), object : BiometricPrompt.AuthenticationCallback() {
            override fun onAuthenticationSucceeded(result: BiometricPrompt.AuthenticationResult) { d.complete(true) }
            override fun onAuthenticationError(errorCode: Int, errString: CharSequence) { d.complete(false) }
        })
        p.authenticate(BiometricPrompt.PromptInfo.Builder().setTitle("Tu Local Web").setSubtitle(motivo).setAllowedAuthenticators(permitidos).build())
        return d.await()
    }
}
