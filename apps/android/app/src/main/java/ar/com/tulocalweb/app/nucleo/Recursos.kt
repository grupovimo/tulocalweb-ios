package ar.com.tulocalweb.app.nucleo

import android.content.Context

/** Los archivos que vienen con la app (Bundle.main en iOS: ayuda.json, plantillas.json, estilo.json, en assets/). */
object Recursos {
    internal var contexto: Context? = null

    fun texto(nombre: String): String? = runCatching { contexto?.assets?.open(nombre)?.bufferedReader()?.use { it.readText() } }.getOrNull()

    inline fun <reified T> leer(nombre: String): T? = texto(nombre)?.let { runCatching { Api.json.decodeFromString<T>(it) }.getOrNull() }
}
