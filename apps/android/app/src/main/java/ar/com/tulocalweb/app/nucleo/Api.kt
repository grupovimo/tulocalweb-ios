package ar.com.tulocalweb.app.nucleo

import android.content.Context
import ar.com.tulocalweb.app.BuildConfig
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.suspendCancellableCoroutine
import kotlinx.coroutines.withContext
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json
import okhttp3.Call
import okhttp3.Callback
import okhttp3.HttpUrl.Companion.toHttpUrl
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody
import okhttp3.Response
import java.io.File
import java.io.IOException
import java.net.URLEncoder
import java.time.Instant
import java.util.concurrent.TimeUnit
import kotlin.coroutines.resume
import kotlin.coroutines.resumeWithException

/** Un error de la API (application/problem+json) o de la red, con lo que hay que decirle a la persona. */
data class ErrorApi(
    val estado: Int,
    val codigo: String,
    val titulo: String,
    val detalle: String,
    /** Ingreso con contraseña: intentos que quedan antes de pausarlo. */
    val restantes: Int? = null,
    /** Segundos para volver a probar (429). */
    val reintentarEn: Int? = null,
) : Exception(detalle) {
    val sinConexion: Boolean get() = codigo == "sin_conexion"
    val sinSesion: Boolean get() = estado == 401 && codigo == "sin_sesion"

    companion object {
        val sinRed = ErrorApi(0, "sin_conexion", "Sin conexión", "No hay internet. Revisá la conexión y probá de nuevo.")
        val desconocido = ErrorApi(0, "interno", "Algo salió mal", "Probá de nuevo en un rato.")
    }
}

/** El ErrorApi de cualquier error (lo que en iOS es `catch let e as ErrorApi` y si no `.desconocido`). */
fun errorDe(e: Throwable): ErrorApi = e as? ErrorApi ?: ErrorApi.desconocido

private suspend fun Call.esperar(): Response = suspendCancellableCoroutine { cont ->
    enqueue(object : Callback {
        override fun onFailure(call: Call, e: IOException) = cont.resumeWithException(e)
        override fun onResponse(call: Call, response: Response) = cont.resume(response)
    })
    cont.invokeOnCancellation { runCatching { cancel() } }
}

/** El cliente de la API de Tu Local Web (apps/api, contrato en packages/contrato/openapi.yaml), clon de
 *  Nucleo/Api.swift. Va directo contra api.tulocalweb.com.ar con la sesión propia en `Authorization: Bearer`. */
class Api(val base: String, @Volatile var token: String? = null, val http: OkHttpClient = cliente()) {
    /** Se llama cuando la API dice que la sesión ya no vale (401 sin_sesion): la app vuelve al ingreso. */
    var alPerderSesion: (suspend () -> Unit)? = null

    @Serializable
    private data class Problema(val status: Int? = null, val code: String? = null, val title: String? = null, val detail: String? = null, val restantes: Int? = null)

    suspend fun pedir(metodo: String, ruta: String, query: List<Pair<String, String>> = emptyList(), cuerpo: ByteArray? = null): String {
        val url = (base.trimEnd('/') + ruta).toHttpUrl().newBuilder().apply { query.forEach { (k, v) -> addQueryParameter(k, v) } }.build()
        val r = Request.Builder().url(url).header("Accept", "application/json")
        token?.let { r.header("Authorization", "Bearer $it") }
        val cuerpoHttp = cuerpo?.toRequestBody("application/json".toMediaType())
            ?: if (metodo in listOf("POST", "PUT", "PATCH")) ByteArray(0).toRequestBody(null) else null
        r.method(metodo, cuerpoHttp)
        val (codigo, texto, reintentar) = try {
            withContext(Dispatchers.IO) {
                http.newCall(r.build()).esperar().use { Triple(it.code, it.body?.string().orEmpty(), it.header("Retry-After")) }
            }
        } catch (e: CancellationException) {
            throw e
        } catch (e: IOException) {
            throw ErrorApi.sinRed
        }
        if (codigo in 200..299) return texto
        val p = runCatching { json.decodeFromString<Problema>(texto) }.getOrElse { Problema() }
        val e = ErrorApi(codigo, p.code ?: "interno", p.title ?: "Algo salió mal", p.detail ?: "Probá de nuevo en un rato.", p.restantes, reintentar?.toIntOrNull())
        if (e.sinSesion && token != null) alPerderSesion?.invoke()
        throw e
    }

    suspend inline fun <reified T> get(ruta: String, query: List<Pair<String, String>> = emptyList()): T = decodificar(pedir("GET", ruta, query))

    suspend inline fun <reified T, reified B> enviar(metodo: String, ruta: String, cuerpo: B): T =
        decodificar(pedir(metodo, ruta, cuerpo = json.encodeToString(cuerpo).toByteArray()))

    /** Un pedido sin respuesta (204). */
    suspend inline fun <reified B> hacer(metodo: String, ruta: String, cuerpo: B) {
        pedir(metodo, ruta, cuerpo = json.encodeToString(cuerpo).toByteArray())
    }

    suspend fun hacer(metodo: String, ruta: String) { pedir(metodo, ruta) }

    inline fun <reified T> decodificar(d: String): T = try {
        json.decodeFromString<T>(d)
    } catch (e: Exception) {
        if (BuildConfig.DEBUG) android.util.Log.w("Api", "no pude leer ${T::class.simpleName}: $e")
        throw ErrorApi.desconocido
    }

    companion object {
        /** La API de la configuración (BuildConfig.API, de tlw.api en local.properties). */
        fun urlConfigurada(): String = BuildConfig.API.ifEmpty { "https://api.tulocalweb.com.ar" }

        val json = Json { ignoreUnknownKeys = true; explicitNulls = false; coerceInputValues = true; encodeDefaults = true }

        fun cliente(): OkHttpClient = OkHttpClient.Builder()
            .connectTimeout(20, TimeUnit.SECONDS).readTimeout(20, TimeUnit.SECONDS).writeTimeout(60, TimeUnit.SECONDS)
            .build()
    }
}

/** Un GET que puede fallar sin cortar todo (try? en iOS). */
suspend inline fun <reified T> Api.opcional(ruta: String, query: List<Pair<String, String>> = emptyList()): T? =
    try { get<T>(ruta, query) } catch (e: CancellationException) { throw e } catch (e: Exception) { null }

/** Un pedido que puede fallar sin cortar todo (try? await api.hacer en iOS). */
suspend inline fun intentar(bloque: () -> Unit): Boolean =
    try { bloque(); true } catch (e: CancellationException) { throw e } catch (e: Exception) { false }

/** Un cuerpo vacío (`{}`). */
@Serializable
class Vacio

/** Lo último que se trajo de cada lista, para verlo sin conexión (handoff 5a: «Ves lo guardado hoy, 14:20»). Un
 *  archivo por clave en la caché, con la fecha en que se guardó. */
object Guardado {
    internal var contexto: Context? = null
    private val carpeta: File? get() = contexto?.let { File(it.cacheDir, "guardado").apply { mkdirs() } }
    private fun archivo(clave: String) = carpeta?.let { File(it, URLEncoder.encode(clave, "UTF-8")) }

    inline fun <reified T> guardar(valor: T, clave: String) {
        val texto = runCatching { Api.json.encodeToString(valor) }.getOrNull() ?: return
        escribir(clave, texto)
    }

    fun escribir(clave: String, texto: String) { runCatching { archivo(clave)?.writeText(texto) } }

    inline fun <reified T> leer(clave: String): Pair<T, Instant>? {
        val (texto, fecha) = crudo(clave) ?: return null
        return runCatching { Api.json.decodeFromString<T>(texto) to fecha }.getOrNull()
    }

    fun crudo(clave: String): Pair<String, Instant>? {
        val f = archivo(clave)?.takeIf { it.exists() } ?: return null
        return runCatching { f.readText() to Instant.ofEpochMilli(f.lastModified()) }.getOrNull()
    }

    /** Al cerrar sesión o cambiar de tienda no queda nada de la anterior. */
    fun borrarTodo() { carpeta?.deleteRecursively() }
}
