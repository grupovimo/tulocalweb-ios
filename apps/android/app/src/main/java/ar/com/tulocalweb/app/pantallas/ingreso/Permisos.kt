package ar.com.tulocalweb.app.pantallas.ingreso

import android.Manifest
import androidx.compose.animation.AnimatedContent
import androidx.compose.animation.core.animateDpAsState
import androidx.compose.animation.core.tween
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.slideInHorizontally
import androidx.compose.animation.togetherWith
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.R
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.Tincho
import ar.com.tulocalweb.app.tema.*
import kotlinx.coroutines.launch

// Clon de Pantallas/Ingreso/Permisos.swift.

private data class Paso(val tincho: Tincho, val tam: Dp, val titulo: String, val texto: String, val boton: String)

/** 1e–1g · Los pre-permisos: notificaciones, cámara y la huella. Cada botón principal dispara el diálogo del sistema;
 *  «Ahora no» y «Saltear» siguen sin pedirlo. */
@Composable
fun Permisos() {
    val estado = LocalEstado.current
    val alcance = rememberCoroutineScope()
    var paso by remember { mutableIntStateOf(0) }
    val ps = remember {
        buildList {
            add(Paso(Tincho.hablando, 210.dp, "Te aviso cada venta al instante",
                "Cuando entra un pedido o se acredita un pago, te llega un aviso. Vos elegís cuáles.", "Activar avisos"))
            add(Paso(Tincho.foto, 240.dp, "Fotos y códigos, con la cámara",
                "La usamos para sacarle fotos a tus productos y para escanear códigos de barras en el local. Nada más.", "Permitir la cámara"))
            if (estado.hayBiometria()) add(Paso(Tincho.seguridad, 240.dp, "Entrá con tu huella o tu cara",
                "Es más rápido que la contraseña. También te lo pedimos para cosas delicadas, como cancelar un pedido.", "Usar tu huella o tu cara"))
        }
    }
    val p = ps[minOf(paso, ps.size - 1)]

    fun siguiente() { if (paso + 1 < ps.size) paso += 1 else estado.terminarPermisos() }

    suspend fun aceptar() {
        when (paso) {
            0 -> estado.pedirPermisoPush()
            1 -> estado.alPedirPermiso(Manifest.permission.CAMERA)
            else -> if (estado.confirmarIdentidad("Para entrar a tu tienda con tu huella o tu cara")) estado.ponerFaceID(true)
        }
        siguiente()
    }

    Column(Modifier.fillMaxSize().background(C.fondo).statusBarsPadding().navigationBarsPadding()) {
        Row(Modifier.fillMaxWidth().padding(horizontal = 24.dp, vertical = 10.dp), verticalAlignment = Alignment.CenterVertically) {
            Row(Modifier.weight(1f).semantics { contentDescription = "Paso ${paso + 1} de ${ps.size}" }, horizontalArrangement = Arrangement.spacedBy(6.dp)) {
                ps.indices.forEach { i ->
                    val w by animateDpAsState(if (i == paso) 22.dp else 8.dp, tween(250), label = "punto")
                    Box(Modifier.width(w).height(5.dp).clip(Capsula).background(if (i == paso) C.primarioTexto else tema(0x3A1D12, 0xFBF7F2, alfa = 0.18f).actual))
                }
            }
            Texto("Saltear", Estilo.subheadFuerte, C.texto2, Modifier.clickable { estado.terminarPermisos() })
        }
        AnimatedContent(paso, Modifier.weight(1f), transitionSpec = { (slideInHorizontally { it } + fadeIn()) togetherWith fadeOut() }, label = "paso") { _ ->
            Column(Modifier.fillMaxWidth().verticalScroll(rememberScrollState()).padding(horizontal = 28.dp).padding(top = 8.dp),
                horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(14.dp)) {
                ImagenTincho(p.tincho, Modifier.size(p.tam))
                Texto(p.titulo, Estilo.tituloPaso, C.texto, alinear = TextAlign.Center)
                Texto(Formato.sinViudas(p.texto), Estilo.body16, C.texto2, alinear = TextAlign.Center, interlineado = 3f)
                if (paso == 0) Box(Modifier.padding(top = 14.dp)) { MockPush() }
            }
        }
        Column(Modifier.padding(horizontal = 24.dp).padding(bottom = 16.dp), verticalArrangement = Arrangement.spacedBy(10.dp)) {
            BotonPrincipal(p.boton, { alcance.launch { aceptar() } })
            BotonTexto("Ahora no", { siguiente() })
        }
    }
}

/** El push de muestra de 1e. */
@Composable
private fun MockPush() {
    val forma = RoundedCornerShape(20.dp)
    Row(Modifier.fillMaxWidth().sombra(forma, hex(0x2A140C, 0.2f), 14.dp, y = 10.dp).clip(forma).background(C.superficie.copy(alpha = 0.85f))
        .padding(horizontal = 14.dp, vertical = 12.dp), horizontalArrangement = Arrangement.spacedBy(11.dp), verticalAlignment = Alignment.Top) {
        Imagen(R.drawable.isotipo, Modifier.size(38.dp).clip(RoundedCornerShape(9.dp)))
        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(1.dp)) {
            Row {
                Texto("Venta nueva · ${Formato.pesos(24020)}", Estilo.subheadFuerte, C.texto, Modifier.weight(1f))
                Texto("ahora", Estilo.footnote, C.texto2)
            }
            Texto("Martina compró Yerba Lumbre suave y 2 productos más.", Estilo.subhead, C.texto)
        }
    }
}
