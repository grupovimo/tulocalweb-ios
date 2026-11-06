package ar.com.tulocalweb.app.pantallas.productos

import androidx.compose.foundation.background
import androidx.compose.foundation.gestures.awaitEachGesture
import androidx.compose.foundation.gestures.awaitFirstDown
import androidx.compose.foundation.gestures.waitForUpOrCancellation
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.SpanStyle
import androidx.compose.ui.text.buildAnnotatedString
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.withStyle
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.nucleo.ErrorApi
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.ProductoFila
import ar.com.tulocalweb.app.tema.*
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import kotlinx.serialization.Serializable

// Clon de Pantallas/Productos/StockProducto.swift.

/** 11m · Stock: segmentado Ajuste / Ingreso / Conteo, stepper grande de 52 con el valor relativo («+24») y «Quedan 2 →
 *  van a quedar 26». Con variantes, primero se elige cuál. POST /v1/panel/listas/inventario/acciones (ajustar, ingreso,
 *  conteo). */
@Composable
fun StockProducto(producto: ProductoFila, alGuardar: suspend () -> Unit) {
    val estado = LocalEstado.current
    val cerrar = LocalCerrarHoja.current
    val alcance = rememberCoroutineScope()
    val variantes = producto.variantes.filter { it.activa }
    var modo by remember { mutableStateOf("ingreso") }
    var cantidad by remember { mutableIntStateOf(0) }
    var variante by remember { mutableStateOf(variantes.firstOrNull()?.id) }
    var guardando by remember { mutableStateOf(false) }
    val actual = variantes.firstOrNull { it.id == variante }?.stock ?: producto.stock
    val queda = if (modo == "conteo") maxOf(0, cantidad) else maxOf(0, actual + cantidad)
    val pregunta = when (modo) { "ajuste" -> "¿Cuánto sumás o restás?"; "conteo" -> "¿Cuántos contaste?"; else -> "¿Cuántos entraron?" }
    val valor = if (modo == "conteo") "$cantidad" else if (cantidad > 0) "+$cantidad" else if (cantidad < 0) "−${-cantidad}" else "0"
    val boton = when (modo) { "ajuste" -> "Guardar ajuste"; "conteo" -> "Guardar conteo"; else -> "Cargar ingreso" }

    @Serializable data class Ajuste(val ids: List<String>, val accion: String, val valor: Int)
    @Serializable data class Item(val id: String, val cantidad: Int)
    @Serializable data class Conteo(val ids: List<String>, val accion: String = "conteo", val valor: List<Item>)
    @Serializable data class R(val hechos: Int = 0)

    suspend fun guardar() {
        val id = variante?.let { "${producto.id}:$it" } ?: producto.id
        guardando = true
        try {
            val r: R = if (modo == "conteo") estado.api.enviar("POST", "/v1/panel/listas/inventario/acciones", Conteo(listOf(id), valor = listOf(Item(id, cantidad))))
            else estado.api.enviar("POST", "/v1/panel/listas/inventario/acciones", Ajuste(listOf(id), if (modo == "ingreso") "ingreso" else "ajustar", cantidad))
            Haptico.exito()
            cerrar()
            estado.avisar(if (r.hechos > 0) "Stock de ${producto.nombre}: $queda" else "No cambió el stock")
            estado.version += 1
            alGuardar()
        } catch (e: ErrorApi) { Haptico.error(); estado.avisar(e.detalle, ok = false) } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
        } finally { guardando = false }
    }

    @Composable
    fun Paso(simbolo: String, d: Int) {
        val minimo = if (modo == "ingreso") 0 else if (modo == "conteo") 0 else -actual
        val deshabilitado = d < 0 && cantidad <= minimo
        val cambiar by rememberUpdatedState { if (!(d < 0 && cantidad <= minimo)) { cantidad += d; Haptico.seleccion() } }
        // .buttonRepeatBehavior(.enabled): mantener apretado sigue sumando.
        Box(Modifier.size(52.dp).alpha(if (deshabilitado) 0.45f else 1f).clip(CircleShape).background(C.pista)
            .semantics { contentDescription = if (d > 0) "Sumar uno" else "Restar uno" }
            .pointerInput(d) {
                awaitEachGesture {
                    awaitFirstDown()
                    cambiar()
                    val repetir = alcance.launch { delay(400); while (true) { cambiar(); delay(80) } }
                    waitForUpOrCancellation()
                    repetir.cancel()
                }
            }, contentAlignment = Alignment.Center) {
            Icono(simbolo, 22f, C.primarioTexto)
        }
    }

    Column(Modifier.fillMaxSize().padding(horizontal = 20.dp), verticalArrangement = Arrangement.spacedBy(14.dp)) {
        Grabber()
        Column(Modifier.padding(horizontal = 4.dp).padding(top = 6.dp), verticalArrangement = Arrangement.spacedBy(3.dp)) {
            Texto("Stock de ${producto.nombre}", Estilo.tituloHojaChico, C.texto, lineas = 2)
            if (actual == 0) Texto("Sin stock: sigue visible en la tienda", Estilo.label, C.errorTinta)
            else if (actual < stockBajo) Texto("Stock bajo: avisamos con menos de $stockBajo", Estilo.label, C.avisoTinta)
        }
        if (variantes.isNotEmpty()) {
            LazyRow(Modifier.sangrar(20.dp), contentPadding = androidx.compose.foundation.layout.PaddingValues(horizontal = 20.dp), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                items(variantes, key = { it.id }) { v ->
                    ChipFiltro(v.texto, v.stock, cuentaAlerta = v.stock == 0, activo = variante == v.id) { variante = v.id; cantidad = if (modo == "conteo") v.stock else 0 }
                }
            }
        }
        Segmentado(listOf("ajuste" to "Ajuste", "ingreso" to "Ingreso", "conteo" to "Conteo"), modo, { modo = it; cantidad = if (it == "conteo") actual else 0 })
        Column(Modifier.fillMaxWidth().tarjeta().padding(18.dp), horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(12.dp)) {
            Texto(pregunta, Estilo.subhead, C.texto2)
            Row(horizontalArrangement = Arrangement.spacedBy(18.dp), verticalAlignment = Alignment.CenterVertically) {
                Paso("minus", -1)
                NumeroAnimado(valor, Estilo(52f, Peso.bold, -0.03f), C.texto, Modifier.widthIn(min = 70.dp))
                Paso("plus", 1)
            }
            val tx = C.texto
            TextoRico(buildAnnotatedString {
                append("Quedan "); withStyle(SpanStyle(fontWeight = FontWeight.Bold)) { append("$actual") }
                append(" → van a quedar "); withStyle(SpanStyle(fontWeight = FontWeight.Bold)) { append("$queda") }
            }, Estilo.subhead, tx, numeros = true, alinear = TextAlign.Center)
        }
        val costo = producto.costo
        if (costo != null && costo > 0) {
            Row(Modifier.fillMaxWidth().tarjeta().padding(horizontal = 16.dp, vertical = 13.dp)) {
                Texto("Costo por unidad", Estilo.body16, C.texto, Modifier.weight(1f))
                Texto(Formato.pesos(costo), Estilo.body16, C.texto2, numeros = true)
            }
        }
        Spacer(Modifier.weight(1f))
        BotonPrincipal(boton, { alcance.launch { guardar() } }, Modifier.padding(bottom = 12.dp), alto = 56.dp,
            habilitado = !(guardando || (modo != "conteo" && cantidad == 0) || (variantes.isNotEmpty() && variante == null)), cargando = guardando)
    }
}

