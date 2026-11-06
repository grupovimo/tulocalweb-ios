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

    private fun empezar() {
        val r = if (android.os.Build.VERSION.SDK_INT >= 31 && SpeechRecognizer.isOnDeviceRecognitionAvailable(ctx)) SpeechRecognizer.createOnDeviceSpeechRecognizer(ctx)
        else SpeechRecognizer.createSpeechRecognizer(ctx)
        reconocedor = r
        r.setRecognitionListener(object : RecognitionListener {
            override fun onReadyForSpeech(params: Bundle?) {}
            override fun onBeginningOfSpeech() {}
            override fun onRmsChanged(rmsdB: Float) {}
            override fun onBufferReceived(buffer: ByteArray?) {}
            override fun onEndOfSpeech() {}
            override fun onEvent(eventType: Int, params: Bundle?) {}
            override fun onError(e: Int) { parar() }
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
            putExtra(RecognizerIntent.EXTRA_PREFER_OFFLINE, true)
        })
        reiniciarSilencio()
    }

    private fun reiniciarSilencio() {
        silencio?.cancel()
        silencio = alcance.launch { delay(2000); parar() }
    }

    fun parar() {
        silencio?.cancel()
        reconocedor?.let { runCatching { it.stopListening(); it.destroy() } }
        reconocedor = null
        escuchando = false
    }
}
