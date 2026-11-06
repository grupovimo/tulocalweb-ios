package ar.com.tulocalweb.app.nucleo

import ar.com.tulocalweb.app.BuildConfig
import ar.com.tulocalweb.app.pantallas.tincho.RespuestaTincho
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow
import kotlinx.coroutines.flow.flowOn
import kotlinx.coroutines.withContext
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.JsonArray
import kotlinx.serialization.json.JsonElement
import kotlinx.serialization.json.JsonNull
import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.JsonPrimitive
import kotlinx.serialization.json.booleanOrNull
import kotlinx.serialization.json.doubleOrNull
import kotlinx.serialization.json.jsonPrimitive
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody
import java.nio.ByteBuffer
import java.nio.CharBuffer
import java.nio.charset.CodingErrorAction
import java.util.concurrent.TimeUnit

/** Un JSON cualquiera: para editar un producto sin perder los campos que la app no muestra (opciones, variantes,
 *  canales) al mandarlo de vuelta (ValorJSON en iOS). */
typealias ValorJSON = JsonElement

val ValorJSON.texto: String? get() = (this as? JsonPrimitive)?.takeIf { it.isString }?.content
val ValorJSON.entero: Int? get() = (this as? JsonPrimitive)?.takeIf { !it.isString }?.doubleOrNull?.toInt()
val ValorJSON.numero: Double? get() = (this as? JsonPrimitive)?.takeIf { !it.isString }?.doubleOrNull
val ValorJSON.bool: Boolean? get() = (this as? JsonPrimitive)?.takeIf { !it.isString }?.booleanOrNull
val ValorJSON.lista: List<ValorJSON>? get() = this as? JsonArray
val ValorJSON.objeto: Map<String, ValorJSON>? get() = this as? JsonObject
operator fun ValorJSON.get(k: String): ValorJSON? = (this as? JsonObject)?.get(k)

/** Para armar un ValorJSON a mano (los .texto, .numero… de iOS). */
object Valor {
    fun texto(s: String): ValorJSON = JsonPrimitive(s)
    fun numero(n: Number): ValorJSON = JsonPrimitive(n)
    fun booleano(b: Boolean): ValorJSON = JsonPrimitive(b)
    fun lista(l: List<ValorJSON>): ValorJSON = JsonArray(l)
    fun objeto(o: Map<String, ValorJSON>): ValorJSON = JsonObject(o)
    val nulo: ValorJSON = JsonNull
}

/** Lo que la app usa de la web (no de la API): subir fotos y hablar con Tincho. La web acepta el mismo token de la
 *  API en esas dos rutas (hooks.server.ts, RUTAS_APP). */
object Web {
    val base: String = BuildConfig.WEB.ifEmpty { "https://tulocalweb.com.ar" }.trimEnd('/')
    private val http by lazy { Api.cliente() }

    /** Sube una foto (JPEG) y devuelve su URL pública (https://tulocalweb.com.ar/f/…). */
    suspend fun subirFoto(datos: ByteArray, token: String?): String = withContext(Dispatchers.IO) {
        val r = Request.Builder().url("$base/panel/fotos").post(datos.toRequestBody("image/jpeg".toMediaType()))
        token?.let { r.header("Authorization", "Bearer $it") }
        val texto = try {
            http.newCall(r.build()).execute().use { if (!it.isSuccessful) throw ErrorApi.desconocido; it.body?.string().orEmpty() }
        } catch (e: java.io.IOException) { throw ErrorApi.sinRed }
        @Serializable data class R(val ruta: String)
        val ruta = runCatching { Api.json.decodeFromString<R>(texto).ruta }.getOrNull() ?: throw ErrorApi.desconocido
        if (ruta.startsWith("http")) ruta else base + ruta
    }

    @Serializable private data class M(val rol: String, val texto: String)
    @Serializable private data class CuerpoBloques(val mensajes: List<M>, val formato: String = "bloques")
    @Serializable private data class Cuerpo(val mensajes: List<M>)

    /** Tincho con bloques (cifras, gráficos, tablas, botones y sugerencias): la respuesta entera en JSON. */
    suspend fun tinchoConBloques(mensajes: List<Pair<String, String>>, token: String?): RespuestaTincho = withContext(Dispatchers.IO) {
        val c = CuerpoBloques(mensajes.takeLast(12).map { M(it.first, it.second.take(1000)) })
        val r = Request.Builder().url("$base/panel/hornero").post(Api.json.encodeToString(c).toRequestBody("application/json".toMediaType()))
        token?.let { r.header("Authorization", "Bearer $it") }
        val largo = http.newBuilder().readTimeout(40, TimeUnit.SECONDS).callTimeout(40, TimeUnit.SECONDS).build()
        val texto = try {
            largo.newCall(r.build()).execute().use { if (!it.isSuccessful) throw ErrorApi.desconocido; it.body?.string().orEmpty() }
        } catch (e: java.io.IOException) { throw ErrorApi.sinRed }
        runCatching { Api.json.decodeFromString<RespuestaTincho>(texto) }.getOrNull() ?: throw ErrorApi.desconocido
    }

    /** Una pregunta a Tincho: va devolviendo el texto a medida que llega. */
    fun tincho(mensajes: List<Pair<String, String>>, token: String?): Flow<String> = flow {
        val c = Cuerpo(mensajes.takeLast(12).map { M(it.first, it.second) })
        val r = Request.Builder().url("$base/panel/hornero").post(Api.json.encodeToString(c).toRequestBody("application/json".toMediaType()))
        token?.let { r.header("Authorization", "Bearer $it") }
        val largo = http.newBuilder().readTimeout(0, TimeUnit.MILLISECONDS).build()
        largo.newCall(r.build()).execute().use { res ->
            if (!res.isSuccessful) throw ErrorApi.desconocido
            val fuente = res.body?.byteStream() ?: throw ErrorApi.desconocido
            // Se manda cuando el pedazo es UTF-8 válido (no se corta una letra a la mitad).
            val decodificador = Charsets.UTF_8.newDecoder().onMalformedInput(CodingErrorAction.REPORT)
            val buf = ByteArray(1024)
            var acumulado = ByteArray(0)
            while (true) {
                val n = fuente.read(buf)
                if (n < 0) break
                acumulado += buf.copyOf(n)
                val s = runCatching { decodificador.decode(ByteBuffer.wrap(acumulado)).toString() }.getOrNull()
                if (s != null) { emit(s); acumulado = ByteArray(0) }
            }
        }
    }.flowOn(Dispatchers.IO)
}
