package ar.com.tulocalweb.app.pantallas.inicio

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.IntrinsicSize
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.drawBehind
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.nucleo.ErrorApi
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.tema.*
import kotlinx.coroutines.launch
import kotlinx.serialization.Serializable
import java.time.Instant
import java.time.YearMonth

// Clon de Pantallas/Inicio/EditarMeta.swift.

/** 12e · Editar la meta del mes: hoja de altura fija con el monto, atajos («Igual que el mes pasado», «+10 %», «+20 %»)
 *  y el promedio diario que hace falta. PUT /v1/panel/meta. */
@Composable
fun EditarMeta(actual: Int, vendido: Int, alGuardar: suspend () -> Unit) {
    val estado = LocalEstado.current
    val cerrar = LocalCerrarHoja.current
    val alcance = rememberCoroutineScope()
    var texto by remember { mutableStateOf(if (actual > 0) actual.toString() else "") }
    var atajo by remember { mutableStateOf<Int?>(null) }
    var guardando by remember { mutableStateOf(false) }
    val foco = remember { FocusRequester() }
    val monto = texto.filter { it.isDigit() }.toIntOrNull() ?: 0
    val hoy = Formato.enArgentina(Instant.now())
    val dias = YearMonth.from(hoy).lengthOfMonth()
    val quedan = maxOf(1, dias - hoy.dayOfMonth + 1)
    val pasaron = maxOf(1, hoy.dayOfMonth - 1)
    val base = if (actual > 0) actual else maxOf(vendido, 1)

    @Serializable data class Cuerpo(val objetivo: Int)

    suspend fun guardar() {
        guardando = true
        try {
            estado.api.hacer("PUT", "/v1/panel/meta", Cuerpo(monto))
            Haptico.exito()
            cerrar()
            alGuardar()
            estado.avisar(if (monto > 0) "Meta guardada" else "Sacaste la meta del mes")
        } catch (e: ErrorApi) { estado.avisar(e.detalle, ok = false) } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
        } finally { guardando = false }
    }

    @Composable
    fun Chip(t: String, id: Int, accion: () -> Unit) {
        ChipFiltro(t, activo = atajo == id) { accion(); atajo = id; Haptico.seleccion() }
    }

    val grande = Estilo(44f, Peso.bold, -0.03f)
    Column(Modifier.fillMaxWidth().imePadding().padding(horizontal = 20.dp), verticalArrangement = Arrangement.spacedBy(14.dp)) {
        Grabber()
        Row(Modifier.padding(top = 6.dp), verticalAlignment = Alignment.CenterVertically) {
            Texto("Cancelar", Estilo.body, C.primarioTexto, Modifier.clickable { cerrar() })
            Texto("Meta de ${Formato.mesDe(Instant.now())}", Estilo.headline, C.texto, Modifier.weight(1f), alinear = TextAlign.Center)
            Texto("Listo", Estilo(17f, Peso.bold), C.primarioTexto, Modifier.alpha(if (guardando) 0.4f else 1f).clickable(enabled = !guardando) { alcance.launch { guardar() } })
        }
        Column(Modifier.fillMaxWidth().tarjeta().padding(18.dp), horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(6.dp)) {
            Texto("¿Cuánto querés vender?", Estilo.subhead, C.texto2)
            val acento = C.acento
            Row(Modifier.padding(horizontal = 6.dp).padding(bottom = 4.dp).drawBehind {
                drawLine(acento, Offset(0f, size.height), Offset(size.width, size.height), 2.dp.toPx())
            }, verticalAlignment = Alignment.CenterVertically) {
                Texto("$${Formato.fino}", grande, C.texto, numeros = true)
                Box(Modifier.width(IntrinsicSize.Min)) {
                    val mostrado = if (monto > 0) Formato.numero(monto) else ""
                    if (mostrado.isEmpty()) Texto("0", grande, C.texto3, numeros = true)
                    BasicTextField(mostrado, { texto = it; atajo = null }, Modifier.focusRequester(foco).width(IntrinsicSize.Min),
                        textStyle = grande.ts(true).copy(color = C.texto), singleLine = true,
                        keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number), cursorBrush = SolidColor(C.acento))
                }
            }
            Row(Modifier.padding(top = 6.dp), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                if (vendido > 0) Chip("Igual que hoy", 0) { texto = (vendido * dias / pasaron).toString() }
                Chip("+10${Formato.fino}%", 10) { texto = (base * 110 / 100).toString() }
                Chip("+20${Formato.fino}%", 20) { texto = (base * 120 / 100).toString() }
            }
            if (monto > 0) {
                val falta = maxOf(0, monto - vendido)
                Texto(Formato.sinViudas("Son unos ${Formato.pesos(falta / quedan)} por día. Este mes promediás ${Formato.pesos(vendido / pasaron)}."),
                    Estilo.texto14, C.texto2, Modifier.padding(horizontal = 10.dp).padding(top = 6.dp), alinear = TextAlign.Center)
            }
        }
    }
    LaunchedEffect(Unit) { runCatching { foco.requestFocus() } }
}
