package ar.com.tulocalweb.app.pantallas.tincho

import android.Manifest
import android.content.Context
import android.content.Intent
import android.os.Bundle
import android.speech.RecognitionListener
import android.speech.RecognizerIntent
import android.speech.SpeechRecognizer
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import ar.com.tulocalweb.app.app.Estado
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

/** 6a · El micrófono del chat de Tincho (clon de Pantallas/Tincho/Dictado.swift): dictado en español (Argentina) que va
 *  escribiendo en la entrada mientras se habla. Pide el permiso del micrófono la primera vez; se corta solo con 2 s de
 *  silencio o al tocar de nuevo. Usa el reconocedor del teléfono (en el aparato si lo tiene). */
class Dictado(private val ctx: Context, private val estado: Estado) {
    var escuchando by mutableStateOf(false)
    var texto by mutableStateOf("")
    var error by mutableStateOf<String?>(null)

    private var reconocedor: SpeechRecognizer? = null
    private var silencio: Job? = null
    private val alcance = CoroutineScope(Dispatchers.Main)

    suspend fun alternar() {
        if (escuchando) { parar(); return }
        error = null
        if (!estado.alPedirPermiso(Manifest.permission.RECORD_AUDIO) || !SpeechRecognizer.isRecognitionAvailable(ctx)) {
            error = "Para dictar, habilitá el micrófono y el reconocimiento de voz en Ajustes."
            return
        }
        try { empezar() } catch (e: Exception) {
            parar()
            error = "No pudimos usar el micrófono ahora."
        }
    }

    /** El del aparato no tiene bajado el español: desde ahí se usa el de red. */
    private var sinAparato = false

    private fun empezar() {
        val enAparato = !sinAparato && android.os.Build.VERSION.SDK_INT >= 31 && SpeechRecognizer.isOnDeviceRecognitionAvailable(ctx)
        val r = if (enAparato) SpeechRecognizer.createOnDeviceSpeechRecognizer(ctx) else SpeechRecognizer.createSpeechRecognizer(ctx)
        reconocedor = r
        r.setRecognitionListener(object : RecognitionListener {
            override fun onReadyForSpeech(params: Bundle?) {}
            override fun onBeginningOfSpeech() {}
            override fun onRmsChanged(rmsdB: Float) {}
            override fun onBufferReceived(buffer: ByteArray?) {}
            override fun onEndOfSpeech() {}
            override fun onEvent(eventType: Int, params: Bundle?) {}
            override fun onError(e: Int) {
                if (reconocedor !== r) return
                parar()
                val idioma = e == SpeechRecognizer.ERROR_LANGUAGE_NOT_SUPPORTED || e == SpeechRecognizer.ERROR_LANGUAGE_UNAVAILABLE
                if (idioma && enAparato) {
                    // Sin el español en el aparato (Pixel, Samsung): se prueba con el reconocedor de red.
                    sinAparato = true
                    try { empezar() } catch (x: Exception) { error = "No pudimos usar el micrófono ahora." }
                    return
                }
                error = when (e) {
                    SpeechRecognizer.ERROR_NO_MATCH, SpeechRecognizer.ERROR_SPEECH_TIMEOUT -> "No te escuchamos. Tocá el micrófono y hablá de nuevo."
                    SpeechRecognizer.ERROR_NETWORK, SpeechRecognizer.ERROR_NETWORK_TIMEOUT, SpeechRecognizer.ERROR_SERVER, SpeechRecognizer.ERROR_SERVER_DISCONNECTED ->
                        "Para dictar hace falta internet. Revisá la conexión."
                    SpeechRecognizer.ERROR_INSUFFICIENT_PERMISSIONS -> "Para dictar, habilitá el micrófono en los ajustes del teléfono."
                    SpeechRecognizer.ERROR_LANGUAGE_NOT_SUPPORTED, SpeechRecognizer.ERROR_LANGUAGE_UNAVAILABLE -> "Tu teléfono no tiene el dictado en español. Escribile a Tincho."
                    SpeechRecognizer.ERROR_RECOGNIZER_BUSY, SpeechRecognizer.ERROR_TOO_MANY_REQUESTS -> "El dictado está ocupado. Probá de nuevo en un rato."
                    else -> "No pudimos usar el micrófono ahora."
                }
            }
            override fun onPartialResults(p: Bundle?) {
                p?.getStringArrayList(SpeechRecognizer.RESULTS_RECOGNITION)?.firstOrNull()?.let { texto = it; reiniciarSilencio() }
            }
            override fun onResults(p: Bundle?) {
                p?.getStringArrayList(SpeechRecognizer.RESULTS_RECOGNITION)?.firstOrNull()?.let { texto = it }
                parar()
            }
        })
        texto = ""
        escuchando = true
        r.startListening(Intent(RecognizerIntent.ACTION_RECOGNIZE_SPEECH).apply {
            putExtra(RecognizerIntent.EXTRA_LANGUAGE_MODEL, RecognizerIntent.LANGUAGE_MODEL_FREE_FORM)
            putExtra(RecognizerIntent.EXTRA_LANGUAGE, "es-AR")
            putExtra(RecognizerIntent.EXTRA_PARTIAL_RESULTS, true)
            putExtra(RecognizerIntent.EXTRA_PREFER_OFFLINE, enAparato)
        })
        reiniciarSilencio()
    }

    private fun reiniciarSilencio() {
        silencio?.cancel()
        silencio = alcance.launch { delay(2000); parar() }
    }

    fun parar() {
        silencio?.cancel()
        val r = reconocedor
        reconocedor = null
        r?.let { runCatching { it.stopListening(); it.destroy() } }
        escuchando = false
    }
}
