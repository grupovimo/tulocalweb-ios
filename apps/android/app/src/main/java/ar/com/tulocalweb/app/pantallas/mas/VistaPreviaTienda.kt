package ar.com.tulocalweb.app.pantallas.mas

import android.annotation.SuppressLint
import android.os.Handler
import android.os.Looper
import android.webkit.JavascriptInterface
import android.webkit.WebSettings
import android.webkit.WebView
import android.webkit.WebViewClient
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.viewinterop.AndroidView
import ar.com.tulocalweb.app.nucleo.Api
import ar.com.tulocalweb.app.nucleo.Web
import ar.com.tulocalweb.app.nucleo.get
import ar.com.tulocalweb.app.nucleo.texto
import ar.com.tulocalweb.app.tema.Haptico
import kotlinx.serialization.json.JsonPrimitive

// Clon de Pantallas/Mas/VistaPreviaTienda.swift.

/** La tienda de verdad con el borrador (la misma página que el editor web, /panel/mi-tienda/vista, con el token de la
 *  app). Habla con la página con el protocolo de docs/especificaciones/tienda-bloques.md:
 *  - la página avisa `{tipo:'lista'}` por `window.TiendaPuente` y recién ahí se le manda algo;
 *  - `window.tiendaVista.mostrar(<diseño>, '<pagina>')` dibuja el borrador sin guardar ni recargar;
 *  - `window.tiendaVista.seleccionar('<id>')` resalta el bloque elegido en la lista;
 *  - al tocar un bloque (con `marcos=1`) avisa `TiendaPuente.bloque(id)` y el editor abre su panel.
 *  Si la página todavía no responde (no avisa «lista»), queda como antes: se recarga con `vuelta`. */
@SuppressLint("SetJavaScriptEnabled", "JavascriptInterface")
@Composable
fun VistaPreviaTienda(
    token: String?,
    vuelta: Int,
    modifier: Modifier = Modifier,
    pagina: String = "inicio",
    /** El borrador sin guardar, en JSON (Diseno.textoJSON); null: lo guardado. */
    diseno: String? = null,
    seleccion: String? = null,
    marcos: Boolean = false,
    alTocarBloque: ((String) -> Unit)? = null,
    /** true cuando la página responde al protocolo (los cambios se ven sin guardar). */
    alCambiarViva: ((Boolean) -> Unit)? = null,
) {
    val c = remember { Coordinador() }
    c.alTocarBloque = alTocarBloque
    c.alCambiarViva = alCambiarViva
    AndroidView({ ctx ->
        WebView(ctx).apply {
            settings.javaScriptEnabled = true
            settings.domStorageEnabled = true
            settings.cacheMode = WebSettings.LOAD_NO_CACHE
            setBackgroundColor(android.graphics.Color.TRANSPARENT)
            addJavascriptInterface(Puente(c), "TiendaPuente")
            webViewClient = object : WebViewClient() {
                override fun onPageFinished(view: WebView?, url: String?) {
                    // Por si la página no avisa: si ya sabe dibujar sin recargar, cuenta como lista.
                    view?.evaluateJavascript("!!(window.tiendaVista && window.tiendaVista.mostrar)") { r -> if (r == "true" && !c.lista) c.recibir("lista", null) }
                }
            }
            c.web = this
        }
    }, modifier, onRelease = { c.reloj.removeCallbacksAndMessages(null); c.web = null }, update = { w ->
        c.web = w
        if (diseno != null) c.pendiente = diseno
        if (c.cargada != vuelta || c.pagina != pagina) {
            if (c.lista && c.cargada == vuelta && diseno != null) {
                // Otra página con la vista ya conectada: se le pide que la dibuje, sin recargar.
                c.pagina = pagina
                c.enviar(siempre = true)
            } else {
                c.cargada = vuelta
                c.pagina = pagina
                c.cargar(token, marcos)
            }
        } else if (diseno != null) {
            c.programar()
        }
        if (seleccion != c.seleccionada) {
            c.seleccionada = seleccion
            seleccion?.let { c.seleccionar(it) }
        }
    })
}

private class Coordinador {
    var web: WebView? = null
    var cargada = -1
    var pagina = "inicio"
    var lista = false
    var pendiente: String? = null
    var enviado: String? = null
    var seleccionada: String? = null
    val reloj = Handler(Looper.getMainLooper())
    var alTocarBloque: ((String) -> Unit)? = null
    var alCambiarViva: ((Boolean) -> Unit)? = null

    fun cargar(token: String?, marcos: Boolean) {
        lista = false
        enviado = null
        val avisar = alCambiarViva
        reloj.post { avisar?.invoke(false) }
        val url = "${Web.base}/panel/mi-tienda/vista?pagina=$pagina" + (if (marcos) "&marcos=1" else "")
        web?.loadUrl(url, token?.let { mapOf("Authorization" to "Bearer $it") } ?: emptyMap())
    }

    /** Lo que avisa la página (ya en el hilo principal). */
    fun recibir(tipo: String, id: String?) {
        when (tipo) {
            "lista" -> {
                lista = true
                alCambiarViva?.invoke(true)
                enviar(siempre = true)
                seleccionada?.let { seleccionar(it) }
            }
            "bloque" -> if (id != null) { Haptico.seleccion(); alTocarBloque?.invoke(id) }
        }
    }

    /** Mientras se escribe, manda el borrador un cuarto de segundo después de la última tecla. */
    fun programar() {
        if (!lista || pendiente == enviado) return
        reloj.removeCallbacksAndMessages(null)
        reloj.postDelayed({ enviar() }, 250)
    }

    fun enviar(siempre: Boolean = false) {
        val d = pendiente ?: return
        if (!lista || (!siempre && d == enviado)) return
        enviado = d
        web?.evaluateJavascript("window.tiendaVista && window.tiendaVista.mostrar($d, ${js(pagina)})", null)
    }

    fun seleccionar(id: String) {
        if (!lista) return
        web?.evaluateJavascript("window.tiendaVista && window.tiendaVista.seleccionar(${js(id)})", null)
    }

    /** Un texto como literal de JavaScript (comillas y escapes de JSON). */
    fun js(s: String): String = JsonPrimitive(s).toString()
}

/** window.TiendaPuente: corre en otro hilo, así que pasa todo al principal. */
private class Puente(private val c: Coordinador) {
    private val principal = Handler(Looper.getMainLooper())

    @JavascriptInterface fun bloque(id: String) { principal.post { c.recibir("bloque", id) } }
    @JavascriptInterface fun lista() { principal.post { c.recibir("lista", null) } }

    /** El mensaje entero en JSON ({ tipo: 'bloque', id } o { tipo: 'lista' }), como en iOS. */
    @JavascriptInterface fun postMessage(mensaje: String) {
        val o = runCatching { Api.json.parseToJsonElement(mensaje) }.getOrNull() ?: return
        val tipo = o["tipo"]?.texto ?: return
        val id = o["id"]?.texto
        principal.post { c.recibir(tipo, id) }
    }
}
