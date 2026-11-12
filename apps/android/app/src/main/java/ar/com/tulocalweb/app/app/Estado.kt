package ar.com.tulocalweb.app.app

import android.content.Context
import android.os.Build
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import androidx.compose.runtime.snapshotFlow
import androidx.compose.runtime.staticCompositionLocalOf
import ar.com.tulocalweb.app.BuildConfig
import ar.com.tulocalweb.app.nucleo.Acceso
import ar.com.tulocalweb.app.nucleo.ActividadEnApp
import ar.com.tulocalweb.app.nucleo.Api
import ar.com.tulocalweb.app.nucleo.Clerk
import ar.com.tulocalweb.app.nucleo.DatosSesion
import ar.com.tulocalweb.app.nucleo.ErrorApi
import ar.com.tulocalweb.app.nucleo.errorDe
import ar.com.tulocalweb.app.nucleo.Guardado
import ar.com.tulocalweb.app.nucleo.Llavero
import ar.com.tulocalweb.app.nucleo.Preferencias
import ar.com.tulocalweb.app.nucleo.Proveedores
import ar.com.tulocalweb.app.nucleo.ResumenPanel
import ar.com.tulocalweb.app.nucleo.SesionNueva
import ar.com.tulocalweb.app.nucleo.intentar
import ar.com.tulocalweb.app.nucleo.opcional
import ar.com.tulocalweb.app.tema.Aviso
import kotlinx.coroutines.CompletableDeferred
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.async
import kotlinx.coroutines.coroutineScope
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch
import kotlinx.serialization.Serializable

/** Adónde lleva un push, un widget o un link `tulocalweb://`. */
sealed class Ruta {
    data class pedido(val n: Int) : Ruta()
    data class producto(val id: String) : Ruta()
    data object avisos : Ruta()
    data object pedidos : Ruta()
    data object inicio : Ruta()
    data object compartir : Ruta()
    data object productos : Ruta()
    /** Una sección del panel que la app tiene (clientes, cupones, finanzas…), sobre Inicio. */
    data class destino(val d: Destino) : Ruta()
    /** Un link de tulocalweb.com.ar que la app no tiene: se abre la web dentro de la app. */
    data class web(val url: String) : Ruta()

    companion object {
        /** `pedido/1048` (el `ruta` del push), `tulocalweb://pedido/1048` o un link del panel (Vinculos). */
        fun de(texto: String): Ruta? {
            var t = texto
            val u = runCatching { android.net.Uri.parse(texto) }.getOrNull()
            if (u?.scheme == "tulocalweb") t = (listOf(u.host ?: "") + u.pathSegments).joinToString("/")
            else if (u?.scheme == "https" || u?.scheme == "http") return Vinculos.ruta(u)
            val partes = t.split("/").filter { it.isNotEmpty() }
            return when (partes.firstOrNull()) {
                "pedido" -> partes.getOrNull(1)?.toIntOrNull()?.let { pedido(it) }
                "producto" -> partes.getOrNull(1)?.let { producto(it) }
                "avisos", "notificaciones" -> avisos
                "pedidos" -> pedidos
                "inicio" -> inicio
                "compartir" -> compartir
                // Los push de preguntas de Mercado Libre y de seguridad de la cuenta.
                "preguntas" -> destino(Destino.resenas)
                "cuenta" -> destino(Destino.cuenta)
                else -> null
            }
        }
    }
}

/** El estado de la app (clon de App/Estado.swift): la sesión (token en el Llavero), la tienda y sus permisos, el
 *  resumen del panel (insignias) y los avisos (toasts). Lo comparten todas las pantallas. */
class Estado(api: Api? = null) {
    enum class Fase {
        arrancando,
        afuera,
        eligiendoTienda,
        permisos,
        /** La tienda recién creada: el onboarding (tu negocio, traer tu tienda, plantilla, cobros). */
        onboarding,
        adentro,
    }

    var fase by mutableStateOf(Fase.arrancando)
    /** Ingreso a medias con la verificación en dos pasos: falta el código (hoja en Ingresar). */
    var desafio by mutableStateOf<String?>(null)
    val api: Api = api ?: Api(Api.urlConfigurada(), Llavero.leer("token"))
    var datos by mutableStateOf<DatosSesion?>(null)
    var acceso by mutableStateOf<Acceso?>(null)
    var resumen by mutableStateOf<ResumenPanel?>(null)
    var proveedores by mutableStateOf<Proveedores?>(null)
    var aviso by mutableStateOf<Aviso?>(null)
    var rutaPendiente by mutableStateOf<Ruta?>(null)
    /** Pantalla bloqueada con la huella o la cara hasta que la persona se identifique. */
    var bloqueada by mutableStateOf(false)
    /** Para que las listas se vuelvan a pedir después de un cambio (estado de un pedido, stock). */
    var version by mutableIntStateOf(0)
    /** Token de FCM, cuando Android lo da. */
    var tokenPush by mutableStateOf<String?>(null)
        private set
    private var avisoTarea: Job? = null
    private val alcance = CoroutineScope(SupervisorJob() + Dispatchers.Main.immediate)

    // Lo que pone MainActivity: lo que en iOS da el sistema (hojas web, permisos, Face ID, compartir).
    var contexto: Context? = null
    /** Abre una Custom Tab (la hoja web del sistema). */
    var abrirPestana: (String) -> Unit = {}
    /** El diálogo del sistema para los avisos. */
    var alPedirPermisoPush: suspend () -> Boolean = { true }
    /** Un permiso del sistema (la cámara, el micrófono): AVCaptureDevice.requestAccess en iOS. */
    var alPedirPermiso: suspend (String) -> Boolean = { true }
    /** La huella, la cara o el PIN del teléfono (LAContext.evaluatePolicy). */
    var alConfirmarIdentidad: suspend (String) -> Boolean = { true }
    /** La app abrió una actividad del sistema (la cámara): a la vuelta no se bloquea aunque haya tardado. */
    var abrioOtraApp = false
    /** Si el teléfono tiene huella o cara cargada (canEvaluatePolicy(.deviceOwnerAuthenticationWithBiometrics)). */
    var hayBiometria: () -> Boolean = { false }

    /** Ya pasó por los pre-permisos (1e–1g). */
    var permisosVistos: Boolean
        get() = Preferencias.bool("permisosVistos")
        set(v) = Preferencias.ponerBool("permisosVistos", v)
    /** Una cuenta creada desde la app (o nueva con Apple/Google) que todavía no terminó el onboarding. Sólo así se
     *  muestra: a quien ya usaba la tienda no se le pide de nuevo. */
    var onboardingPendiente: Boolean
        get() = Preferencias.bool("onboardingPendiente")
        set(v) = Preferencias.ponerBool("onboardingPendiente", v)
    /** Entrar con la huella o la cara (1g, y en Perfil, seguridad y tema). */
    var conFaceID by mutableStateOf(Preferencias.bool("faceID"))
    fun ponerFaceID(v: Boolean) { conFaceID = v; Preferencias.ponerBool("faceID", v) }
    /** El tema elegido en Cuenta: sistema, claro u oscuro (@AppStorage("tema")). */
    var tema by mutableStateOf(Preferencias.texto("tema") ?: "sistema")
    fun ponerTema(v: String) { tema = v; Preferencias.ponerTexto("tema", v) }

    init {
        this.api.alPerderSesion = { alcance.launch { sesionPerdida() } }
    }

    val tienda: Acceso.Tienda? get() = acceso?.tienda
    fun puede(permiso: String): Boolean = acceso?.puede(permiso) ?: false

    /** Corre algo en el alcance de la app (Task { … } en iOS). */
    fun tarea(bloque: suspend CoroutineScope.() -> Unit): Job = alcance.launch(block = bloque)

    // MARK: Arranque

    suspend fun arrancar() {
        if (api.token == null) { fase = Fase.afuera; return }
        try {
            cargarSesion()
            if (conFaceID) bloqueada = true
            // Se cerró la app a mitad del onboarding: se retoma.
            if (onboardingPendiente && acceso?.rol == "duenio") { fase = Fase.onboarding; return }
            fase = Fase.adentro
            refrescarResumen()
            pedirTokenPushSiCorresponde()
        } catch (e: kotlinx.coroutines.CancellationException) {
            throw e
        } catch (e: Exception) {
            val error = errorDe(e)
            // Sólo un 401 cierra la sesión. Sin internet, un 500 o un deploy en curso: se entra con lo guardado.
            if (error.sinSesion || error.estado == 401 || error.codigo == "cuenta_borrada") { olvidarSesion(); fase = Fase.afuera; return }
            entrarConLoGuardado(error)
        }
    }

    /** Sin conexión (handoff 5a) o con la API caída: se entra con lo último que se guardó y se avisa. */
    private fun entrarConLoGuardado(error: ErrorApi) {
        datos = Guardado.leer<DatosSesion>("sesion")?.first
        acceso = Guardado.leer<Acceso>("acceso")?.first
        if (datos == null || acceso == null) {
            // Sin nada guardado no hay qué mostrar: la sesión queda y se vuelve a probar al abrir la app.
            fase = Fase.afuera
            avisar(if (error.sinConexion) "No hay internet. Revisá la conexión y probá de nuevo." else "No pudimos cargar tu tienda. Probá de nuevo en un rato.", ok = false)
            return
        }
        if (conFaceID) bloqueada = true
        fase = Fase.adentro
        if (!error.sinConexion) avisar("No pudimos actualizar tu tienda. Ves lo último que se guardó.", ok = false)
        tarea { refrescarResumen(); pedirTokenPushSiCorresponde() }
    }

    private suspend fun cargarSesion() = coroutineScope {
        val d = async { api.get<DatosSesion>("/v1/auth/sesion") }
        val a = async { api.get<Acceso>("/v1/auth/tiendas") }
        val dd = d.await(); val aa = a.await()
        datos = dd
        acceso = aa
        Guardado.guardar(dd, "sesion")
        Guardado.guardar(aa, "acceso")
    }

    /** Vuelve a pedir las tiendas de la persona (por ejemplo, después de cambiar el logo en Mi tienda). */
    suspend fun refrescarAcceso() {
        val a = api.opcional<Acceso>("/v1/auth/tiendas") ?: return
        acceso = a
        Guardado.guardar(a, "acceso")
    }

    suspend fun cargarProveedores() {
        if (proveedores != null) return
        proveedores = api.opcional("/v1/auth/proveedores")
    }

    // MARK: Ingreso

    @Serializable private data class CuerpoIngreso(val email: String, val clave: String, val recordarme: Boolean = true)

    suspend fun ingresar(email: String, clave: String) {
        val s: SesionNueva = api.enviar("POST", "/v1/auth/ingresar", CuerpoIngreso(email, clave))
        entrar(s)
    }

    /** La vuelta de la Custom Tab de Clerk: el nonce, o null si la persona la cerró (cancelado). */
    private var vuelta: CompletableDeferred<String?>? = null
    private var clerk: Clerk? = null

    /** Apple: en Android no hay hoja nativa; va por Clerk con la hoja web (oauth_apple), igual que Google. */
    suspend fun ingresarConApple() = ingresarConClerk("apple")

    suspend fun ingresarConGoogle() = ingresarConClerk("google")

    private suspend fun ingresarConClerk(proveedor: String) {
        cargarProveedores()
        val ctx = contexto ?: throw Clerk.ErrorClerk("sin_clave")
        val c = Clerk(ctx, api.http, proveedores?.clerkPublicable.orEmpty()).takeIf { it.disponible } ?: throw Clerk.ErrorClerk("sin_clave")
        clerk = c
        val url = c.iniciar(proveedor)
        val d = CompletableDeferred<String?>()
        vuelta = d
        abrirPestana(url)
        val nonce = d.await()
        vuelta = null
        if (nonce == CANCELADO) throw Clerk.ErrorClerk("cancelado")
        val jwt = c.completar(nonce)
        canjearClerk(jwt)
    }

    /** MainActivity: llegó la vuelta (`ar.com.tulocalweb.app://callback`). Si el sistema cerró la app mientras estaba
     *  la pestaña, nadie la espera: se termina el ingreso con lo que Clerk dejó guardado (tlw_clerk). */
    fun vueltaClerk(nonce: String?) {
        val d = vuelta
        if (d != null) { d.complete(nonce ?: ""); return }
        val ctx = contexto ?: return
        if (!Clerk.hayIntento(ctx)) return
        tarea {
            snapshotFlow { fase }.first { it != Fase.arrancando }
            if (fase != Fase.afuera) return@tarea
            try {
                cargarProveedores()
                val c = Clerk(ctx, api.http, proveedores?.clerkPublicable.orEmpty()).takeIf { it.disponible } ?: throw Clerk.ErrorClerk("sin_clave")
                canjearClerk(c.completar(nonce))
            } catch (e: kotlinx.coroutines.CancellationException) {
                throw e
            } catch (e: ErrorApi) {
                avisar(e.detalle, ok = false)
            } catch (e: Exception) {
                avisar("No pudimos terminar el ingreso con Google. Probá de nuevo.", ok = false)
            }
        }
    }

    /** MainActivity: la Custom Tab se cerró sin vuelta (ASWebAuthenticationSessionError.canceledLogin). */
    fun clerkCancelado() { vuelta?.complete(CANCELADO) }
    val esperandoClerk: Boolean get() = vuelta != null

    @Serializable private data class CuerpoClerk(val token: String)

    private suspend fun canjearClerk(jwt: String) {
        val s: SesionNueva = api.enviar("POST", "/v1/auth/clerk", CuerpoClerk(jwt))
        if (s.nueva == true) onboardingPendiente = true
        entrar(s)
    }

    @Serializable private data class CuerpoRegistro(val nombre: String, val tienda: String, val email: String, val clave: String)

    /** Crear la cuenta y la tienda desde la app (POST /v1/auth/registro); sigue el onboarding. */
    suspend fun registrar(nombre: String, tienda: String, email: String, clave: String) {
        val s: SesionNueva = api.enviar("POST", "/v1/auth/registro", CuerpoRegistro(nombre, tienda, email, clave))
        onboardingPendiente = true
        entrar(s)
    }

    @Serializable private data class CuerpoSegundo(val desafio: String, val codigo: String)

    /** Completa el ingreso con el código de la app autenticadora (o uno de recuperación). Si el desafío venció, se
     *  vuelve al formulario. */
    suspend fun completarSegundoFactor(codigo: String) {
        val d = desafio ?: return
        try {
            val s: SesionNueva = api.enviar("POST", "/v1/auth/segundo-factor", CuerpoSegundo(d, codigo))
            desafio = null
            entrar(s)
        } catch (e: ErrorApi) {
            if (e.estado == 401) desafio = null
            throw e
        }
    }

    private suspend fun entrar(s: SesionNueva) {
        val d = s.segundoFactor
        if (!d.isNullOrEmpty()) { desafio = d; return }
        Llavero.guardar(s.token, "token")
        api.token = s.token
        Guardado.borrarTodo()
        cargarSesion()
        if ((acceso?.tiendas?.size ?: 0) > 1) fase = Fase.eligiendoTienda else seguirDespuesDeElegir()
    }

    fun seguirDespuesDeElegir() {
        if (onboardingPendiente && acceso?.rol == "duenio") {
            fase = Fase.onboarding
        } else if (permisosVistos) {
            fase = Fase.adentro
            tarea { refrescarResumen(); pedirTokenPushSiCorresponde() }
        } else {
            fase = Fase.permisos
        }
    }

    /** Fin del onboarding (o «lo hago después»): sigue a los permisos o al panel. */
    fun terminarOnboarding() {
        onboardingPendiente = false
        version += 1
        if (permisosVistos) {
            fase = Fase.adentro
            tarea { refrescarResumen(); pedirTokenPushSiCorresponde() }
        } else {
            fase = Fase.permisos
        }
    }

    fun terminarPermisos() {
        permisosVistos = true
        fase = Fase.adentro
        tarea { refrescarResumen() }
    }

    @Serializable private data class CuerpoTienda(val tiendaId: String)

    suspend fun elegirTienda(id: String) {
        if (id != acceso?.tiendaId) {
            api.hacer("PUT", "/v1/auth/tienda", CuerpoTienda(id))
            Guardado.borrarTodo()
            cargarSesion()
            version += 1
            refrescarResumen()
        }
    }

    // MARK: Salir

    suspend fun salir() {
        if (tokenPush != null) intentar { api.hacer("DELETE", "/v1/panel/dispositivos") }
        intentar { api.hacer("POST", "/v1/auth/salir") }
        olvidarSesion()
        fase = Fase.afuera
    }

    @Serializable private data class CuerpoBorrar(val confirmacion: String)

    /** Borra la cuenta de la persona (DELETE /v1/auth/cuenta, con su mail tal cual como confirmación): la API la desactiva,
     *  borra sus datos y cierra las tiendas donde es la única dueña. Después la app olvida la sesión y vuelve a Ingresar. */
    suspend fun borrarCuenta(email: String) {
        api.hacer("DELETE", "/v1/auth/cuenta", CuerpoBorrar(email))
        olvidarSesion()
        permisosVistos = false
        ponerFaceID(false)
        fase = Fase.afuera
        avisar("Borramos tu cuenta y tus datos.")
    }

    private fun sesionPerdida() {
        if (fase !in listOf(Fase.adentro, Fase.eligiendoTienda, Fase.permisos, Fase.onboarding)) return
        olvidarSesion()
        fase = Fase.afuera
        avisar("Tu sesión venció. Entrá de nuevo.", ok = false)
    }

    private fun olvidarSesion() {
        Llavero.borrar("token")
        api.token = null
        onboardingPendiente = false
        contexto?.let { ActividadEnApp.terminarTodo(it) }
        datos = null
        acceso = null
        resumen = null
        bloqueada = false
        Guardado.borrarTodo()
    }

    // MARK: Resumen e insignias

    suspend fun refrescarResumen() {
        if (api.token == null) return
        val r = api.opcional<ResumenPanel>("/v1/panel/resumen") ?: return
        resumen = r
        // El contador del ícono en Android sale de los avisos (no se fija a mano, como setBadgeCount en iOS).
        contexto?.let { ActividadEnApp.sincronizar(it, tienda?.nombre ?: "Tu Local Web") }
    }

    // MARK: Avisos (toasts)

    fun avisar(texto: String, ok: Boolean = true, deshacer: (() -> Unit)? = null) {
        aviso = Aviso(texto, ok, deshacer)
        avisoTarea?.cancel()
        avisoTarea = alcance.launch {
            delay(4000)
            aviso = null
        }
    }

    // MARK: Push

    /** Pide el token de FCM si la persona ya dio permiso (el permiso se pide en 1e). */
    suspend fun pedirTokenPushSiCorresponde() {
        val ctx = contexto ?: return
        if (!Push.hayPermiso(ctx)) return
        Push.token(ctx)?.let { registrarTokenPush(it) }
    }

    /** 1e: el diálogo del sistema y, si dice que sí, el registro. */
    suspend fun pedirPermisoPush(): Boolean {
        val ok = alPedirPermisoPush()
        if (ok) pedirTokenPushSiCorresponde()
        return ok
    }

    @Serializable
    data class Dispositivo(
        val token: String,
        val entorno: String = if (BuildConfig.DEBUG) "sandbox" else "produccion",
        val modelo: String,
        val version: String,
        /** Android: el token es de FCM (la API lo manda por Firebase y no por APNs). */
        val plataforma: String = "android",
    )

    suspend fun registrarTokenPush(token: String) {
        tokenPush = token
        ActividadEnApp.api = api
        ActividadEnApp.tokenPush = token
        contexto?.let { ActividadEnApp.arrancarEscucha(it) }
        if (api.token == null) return
        intentar { api.hacer("POST", "/v1/panel/dispositivos", Dispositivo(token = token, modelo = (Build.MODEL ?: "Android").take(60), version = BuildConfig.VERSION_NAME)) }
    }

    // MARK: Face ID

    /** La huella, la cara o el PIN del teléfono para entrar y para lo delicado (cancelar un pedido). */
    suspend fun confirmarIdentidad(motivo: String): Boolean = alConfirmarIdentidad(motivo)

    suspend fun desbloquear() {
        if (confirmarIdentidad("Entrá a tu tienda")) bloqueada = false
    }

    companion object {
        private const val CANCELADO = "\u0000cancelado"
    }
}

val LocalEstado = staticCompositionLocalOf<Estado> { error("Falta el Estado") }
