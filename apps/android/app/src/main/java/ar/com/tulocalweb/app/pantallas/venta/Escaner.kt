package ar.com.tulocalweb.app.pantallas.venta

import android.Manifest
import android.content.pm.PackageManager
import androidx.camera.core.Camera
import androidx.camera.core.CameraSelector
import androidx.camera.core.ImageAnalysis
import androidx.camera.core.Preview
import androidx.camera.lifecycle.ProcessCameraProvider
import androidx.camera.view.PreviewView
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.viewinterop.AndroidView
import androidx.core.content.ContextCompat
import androidx.lifecycle.compose.LocalLifecycleOwner
import ar.com.tulocalweb.app.app.LocalEstado
import com.google.mlkit.vision.barcode.BarcodeScannerOptions
import com.google.mlkit.vision.barcode.BarcodeScanning
import com.google.mlkit.vision.barcode.common.Barcode
import com.google.mlkit.vision.common.InputImage
import java.util.concurrent.Executors

// Clon de Pantallas/Venta/Escaner.swift: con CameraX y el lector de códigos de Google (ML Kit) en lugar de AVFoundation.

/** La cámara para escanear códigos de barras (EAN, UPC, Code 128, Code 39, QR). Llama a `alLeer` con cada código nuevo
 *  (el mismo código no se repite hasta pasado un segundo y medio). Pide el permiso de la cámara si falta. */
@androidx.annotation.OptIn(androidx.camera.core.ExperimentalGetImage::class)
@Composable
fun Escaner(linterna: Boolean, modifier: Modifier = Modifier.fillMaxSize(), alLeer: (String) -> Unit) {
    val ctx = LocalContext.current
    val estado = LocalEstado.current
    val ciclo = LocalLifecycleOwner.current
    val leer by rememberUpdatedState(alLeer)
    var permiso by remember { mutableStateOf(ContextCompat.checkSelfPermission(ctx, Manifest.permission.CAMERA) == PackageManager.PERMISSION_GRANTED) }
    var camara by remember { mutableStateOf<Camera?>(null) }
    val vista = remember { PreviewView(ctx).apply { scaleType = PreviewView.ScaleType.FILL_CENTER } }
    val hilo = remember { Executors.newSingleThreadExecutor() }

    // El hilo del análisis se cierra al salir del escáner.
    DisposableEffect(Unit) { onDispose { hilo.shutdown() } }

    LaunchedEffect(Unit) { if (!permiso) permiso = estado.alPedirPermiso(Manifest.permission.CAMERA) }

    DisposableEffect(permiso) {
        if (!permiso) return@DisposableEffect onDispose {}
        // ML Kit se inicia solo con su ContentProvider; por si no corrió (los tests), se lo inicia acá.
        com.google.mlkit.common.sdkinternal.MlKitContext.initializeIfNeeded(ctx.applicationContext)
        val lector = BarcodeScanning.getClient(BarcodeScannerOptions.Builder().setBarcodeFormats(
            Barcode.FORMAT_EAN_13, Barcode.FORMAT_EAN_8, Barcode.FORMAT_UPC_E, Barcode.FORMAT_UPC_A, Barcode.FORMAT_CODE_128, Barcode.FORMAT_CODE_39, Barcode.FORMAT_QR_CODE,
        ).build())
        var ultimo: Pair<String, Long>? = null
        val futuro = ProcessCameraProvider.getInstance(ctx)
        futuro.addListener({
            val proveedor = futuro.get()
            val previa = Preview.Builder().build().also { it.surfaceProvider = vista.surfaceProvider }
            val analisis = ImageAnalysis.Builder().setBackpressureStrategy(ImageAnalysis.STRATEGY_KEEP_ONLY_LATEST).build()
            analisis.setAnalyzer(hilo) { cuadro ->
                val img = cuadro.image
                if (img == null) { cuadro.close(); return@setAnalyzer }
                lector.process(InputImage.fromMediaImage(img, cuadro.imageInfo.rotationDegrees))
                    .addOnSuccessListener { codigos ->
                        val c = codigos.firstOrNull()?.rawValue ?: return@addOnSuccessListener
                        val u = ultimo
                        val ahora = System.currentTimeMillis()
                        if (u != null && u.first == c && ahora - u.second < 1500) return@addOnSuccessListener
                        ultimo = c to ahora
                        leer(c)
                    }
                    .addOnCompleteListener { cuadro.close() }
            }
            runCatching {
                proveedor.unbindAll()
                camara = proveedor.bindToLifecycle(ciclo, CameraSelector.DEFAULT_BACK_CAMERA, previa, analisis)
            }
        }, ContextCompat.getMainExecutor(ctx))
        onDispose {
            runCatching { futuro.get().unbindAll() }
            lector.close()
            camara = null
        }
    }

    LaunchedEffect(linterna, camara) { camara?.let { c -> if (c.cameraInfo.hasFlashUnit()) c.cameraControl.enableTorch(linterna) } }

    AndroidView({ vista }, modifier)
}
