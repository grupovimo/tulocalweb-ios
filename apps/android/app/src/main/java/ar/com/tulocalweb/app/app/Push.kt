package ar.com.tulocalweb.app.app

import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.net.Uri
import android.os.Build
import androidx.core.app.NotificationCompat
import androidx.core.app.NotificationManagerCompat
import androidx.core.content.ContextCompat
import ar.com.tulocalweb.app.BuildConfig
import ar.com.tulocalweb.app.MainActivity
import ar.com.tulocalweb.app.R
import ar.com.tulocalweb.app.compartido.MarcarPreparado
import ar.com.tulocalweb.app.compartido.SesionCompartida
import ar.com.tulocalweb.app.nucleo.ActividadEnApp
import ar.com.tulocalweb.app.widgets.Widgets
import com.google.firebase.FirebaseApp
import com.google.firebase.FirebaseOptions
import com.google.firebase.messaging.FirebaseMessaging
import com.google.firebase.messaging.FirebaseMessagingService
import com.google.firebase.messaging.RemoteMessage
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.launch
import kotlinx.coroutines.tasks.await
import kotlinx.serialization.Serializable

// Clon de App/Push.swift (el Delegado de APNs) con FCM: registra el token en /v1/panel/dispositivos, muestra los
// avisos con la app abierta y abre la pantalla exacta al tocarlos (`ruta` del aviso, §40).

/** FCM sin google-services.json (como Tengo Libre y Caigo): FirebaseApp se arma a mano con tlw.firebase.* de
 *  local.properties. Sin esos datos el push queda apagado y la app anda igual. */
object Push {
    /** Los canales: los avisos de la tienda (ventas, preguntas, stock, ingresos) y los pedidos por preparar (sin sonido). */
    const val CANAL_AVISOS = "avisos"
    const val CANAL_PREPARAR = "preparar"

    val disponible: Boolean
        get() = BuildConfig.FIREBASE_APP_ID.isNotBlank() && BuildConfig.FIREBASE_API_KEY.isNotBlank() &&
            BuildConfig.FIREBASE_PROYECTO.isNotBlank() && BuildConfig.FIREBASE_REMITENTE.isNotBlank()

    fun iniciar(ctx: Context) {
        val nm = ctx.getSystemService(NotificationManager::class.java)
        nm.createNotificationChannel(NotificationChannel(CANAL_AVISOS, "Avisos de la tienda", NotificationManager.IMPORTANCE_HIGH).apply {
            description = "Ventas nuevas, preguntas de Mercado Libre, productos sin stock e ingresos a tu cuenta."
        })
        nm.createNotificationChannel(NotificationChannel(CANAL_PREPARAR, "Pedidos por preparar", NotificationManager.IMPORTANCE_LOW).apply {
            description = "Cuántos pedidos te quedan por preparar hoy."
            setShowBadge(false)
        })
        if (!disponible || FirebaseApp.getApps(ctx).isNotEmpty()) return
        runCatching {
            FirebaseApp.initializeApp(ctx, FirebaseOptions.Builder()
                .setApplicationId(BuildConfig.FIREBASE_APP_ID).setApiKey(BuildConfig.FIREBASE_API_KEY)
                .setProjectId(BuildConfig.FIREBASE_PROYECTO).setGcmSenderId(BuildConfig.FIREBASE_REMITENTE).build())
        }
    }

    /** El token de FCM de este teléfono (null sin push). */
    suspend fun token(ctx: Context): String? {
        if (!disponible || FirebaseApp.getApps(ctx).isEmpty()) return null
        return runCatching { FirebaseMessaging.getInstance().token.await() }.getOrNull()
    }

    /** Android 13+ pide permiso para mostrar avisos; antes viene dado. */
    fun necesitaPermiso(ctx: Context): Boolean =
        Build.VERSION.SDK_INT >= 33 && ContextCompat.checkSelfPermission(ctx, android.Manifest.permission.POST_NOTIFICATIONS) != android.content.pm.PackageManager.PERMISSION_GRANTED

    fun hayPermiso(ctx: Context): Boolean = !necesitaPermiso(ctx) && NotificationManagerCompat.from(ctx).areNotificationsEnabled()

    /** Lo que abre un aviso al tocarlo: la app con la ruta (tulocalweb://…) y la tienda del aviso. */
    fun abrir(ctx: Context, ruta: String, codigo: Int, tienda: String? = null, accion: String? = null): PendingIntent {
        val i = Intent(ctx, MainActivity::class.java).apply {
            action = Intent.ACTION_VIEW
            data = Uri.parse("tulocalweb://" + ruta.ifEmpty { "inicio" })
            flags = Intent.FLAG_ACTIVITY_SINGLE_TOP or Intent.FLAG_ACTIVITY_CLEAR_TOP
            putExtra("tienda", tienda)
            putExtra("accion", accion)
        }
        return PendingIntent.getActivity(ctx, codigo, i, PendingIntent.FLAG_IMMUTABLE or PendingIntent.FLAG_UPDATE_CURRENT)
    }

    /** «Preparado» de los pedidos por preparar, sin abrir la app (MarcarPreparado, un LiveActivityIntent en iOS). */
    fun accionPreparado(ctx: Context, numero: Int, aviso: Int, actividad: Boolean = false): PendingIntent {
        val i = Intent(ctx, AccionesPush::class.java).apply {
            action = "ar.com.tulocalweb.app.PREPARADO"
            putExtra("numero", numero); putExtra("aviso", aviso); putExtra("actividad", actividad)
        }
        return PendingIntent.getBroadcast(ctx, ("preparado$numero$actividad").hashCode(), i, PendingIntent.FLAG_IMMUTABLE or PendingIntent.FLAG_UPDATE_CURRENT)
    }

    /** Arma y muestra un aviso: título, cuerpo, agrupado por tienda y, en las ventas (categoría «venta»), «Ver pedido» y
     *  «Marcar como preparado» (7b). */
    fun mostrar(ctx: Context, d: Map<String, String>) {
        val titulo = d["titulo"] ?: return
        val cuerpo = d["cuerpo"].orEmpty()
        val ruta = d["ruta"].orEmpty()
        val tienda = d["tienda"]
        val id = (ruta.ifEmpty { titulo + cuerpo }).hashCode()
        val b = NotificationCompat.Builder(ctx, CANAL_AVISOS)
            .setSmallIcon(R.drawable.ic_aviso)
            .setColor(0xFFD9652F.toInt())
            .setContentTitle(titulo)
            .setContentText(cuerpo)
            .setStyle(NotificationCompat.BigTextStyle().bigText(cuerpo))
            .setAutoCancel(true)
            .setGroup(d["hilo"] ?: "tienda-$tienda")
            .setContentIntent(abrir(ctx, ruta, id, tienda))
        val numero = Regex("^pedido/(\\d+)$").find(ruta)?.groupValues?.get(1)?.toIntOrNull()
        if (d["categoria"] == "venta" && numero != null) {
            b.addAction(0, "Ver pedido", abrir(ctx, ruta, id + 1, tienda))
            // Abre la app y lo marca ahí, con el toast «#N listo» (en iOS la acción es .foreground y pide desbloquear).
            b.addAction(0, "Marcar como preparado", abrir(ctx, ruta, id + 2, tienda, accion = "preparado"))
        }
        if (!hayPermiso(ctx)) return
        runCatching { NotificationManagerCompat.from(ctx).notify(id, b.build()) }
    }
}

/** Recibe los avisos de FCM (vienen sólo con datos: la app arma la notificación con sus acciones) y los pedidos por
 *  preparar que mueve el servidor. */
class ServicioPush : FirebaseMessagingService() {
    private val alcance = CoroutineScope(SupervisorJob() + Dispatchers.IO)

    override fun onNewToken(token: String) {
        val api = SesionCompartida.api(this) ?: return
        alcance.launch { runCatching { api.hacer("POST", "/v1/panel/dispositivos", Estado.Dispositivo(token = token, modelo = (Build.MODEL ?: "Android").take(60), version = BuildConfig.VERSION_NAME)) } }
    }

    override fun onMessageReceived(m: RemoteMessage) {
        val d = m.data.toMutableMap()
        m.notification?.let { n -> d.putIfAbsent("titulo", n.title.orEmpty()); d.putIfAbsent("cuerpo", n.body.orEmpty()) }
        if (d["tipo"] == "actividad") { ActividadEnApp.desdePush(applicationContext, d); return }
        Push.mostrar(applicationContext, d)
        // Con la app abierta: el resumen y las listas al día (willPresent en iOS).
        EnPrimerPlano.alAviso?.invoke()
    }
}

/** Lo que hace la app abierta cuando llega un aviso (refrescar, como willPresent en iOS). */
object EnPrimerPlano { @Volatile var alAviso: (() -> Unit)? = null }

/** «Marcar como preparado» desde el aviso o desde los pedidos por preparar, sin abrir la app. */
class AccionesPush : BroadcastReceiver() {
    @Serializable private data class Cuerpo(val estado: String)

    override fun onReceive(ctx: Context, i: Intent) {
        val numero = i.getIntExtra("numero", 0).takeIf { it > 0 } ?: return
        val aviso = i.getIntExtra("aviso", 0)
        val actividad = i.getBooleanExtra("actividad", false)
        val pendiente = goAsync()
        CoroutineScope(Dispatchers.IO).launch {
            val ok = runCatching {
                if (actividad) MarcarPreparado.hacer(ctx, numero)
                else SesionCompartida.api(ctx)?.hacer("POST", "/v1/panel/pedidos/$numero/estado", Cuerpo("Listo para despachar"))
            }.isSuccess
            if (ok && aviso != 0) NotificationManagerCompat.from(ctx).cancel(aviso)
            // Si no se pudo, se abre el pedido (como iOS: rutaPendiente).
            if (!ok) runCatching { Push.abrir(ctx, "pedido/$numero", numero).send() }
            Widgets.actualizar(ctx)
            EnPrimerPlano.alAviso?.invoke()
            pendiente.finish()
        }
    }
}
