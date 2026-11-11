package ar.com.tulocalweb.app.pantallas.avisos

import android.net.Uri
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.heading
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.app.BarraDe
import ar.com.tulocalweb.app.app.BotonBarra
import ar.com.tulocalweb.app.app.Enlaces
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.app.LocalNavegacion
import ar.com.tulocalweb.app.app.Pestana
import ar.com.tulocalweb.app.nucleo.ErrorApi
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.Notificacion
import ar.com.tulocalweb.app.nucleo.PaginaNotificaciones
import ar.com.tulocalweb.app.nucleo.Tincho
import ar.com.tulocalweb.app.nucleo.errorDe
import ar.com.tulocalweb.app.nucleo.intentar
import ar.com.tulocalweb.app.nucleo.opcional
import ar.com.tulocalweb.app.tema.*
import kotlinx.coroutines.launch

// Clon de Pantallas/Avisos/Avisos.swift.

/** 7c · Centro de avisos: agrupado por día, punto de color por tipo, punto `accent` de no leído y «Leí todo». Con
 *  /v1/panel/notificaciones (la campanita del panel, §45). */
@Composable
fun Avisos() {
    val estado = LocalEstado.current
    val nav = LocalNavegacion.current
    val alcance = rememberCoroutineScope()
    var carga by remember { mutableStateOf<Carga<List<Notificacion>>>(Carga.cargando) }
    var siguiente by remember { mutableStateOf<String?>(null) }

    suspend fun cargar() {
        try {
            val p: PaginaNotificaciones = estado.api.get("/v1/panel/notificaciones", listOf("limite" to "30"))
            carga = Carga.listo(p.notificaciones)
            siguiente = p.siguiente
        } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
            if (carga.valor == null || e !is ErrorApi) carga = Carga.error(errorDe(e))
        }
    }

    suspend fun cargarMas() {
        val s = siguiente ?: return
        val l = carga.valor ?: return
        val p = estado.api.opcional<PaginaNotificaciones>("/v1/panel/notificaciones", listOf("limite" to "30", "cursor" to s))
        if (p != null) { carga = Carga.listo(l + p.notificaciones); siguiente = p.siguiente } else siguiente = null
    }

    suspend fun tocar(n: Notificacion) {
        val l = carga.valor
        if (!n.leida && l != null) {
            carga = Carga.listo(l.map { if (it.id == n.id) it.copy(leida = true) else it })
            intentar { estado.api.hacer("POST", "/v1/panel/notificaciones/${n.id}/leida") }
        }
        // La ruta es la del panel web (/panel/pedidos?pedido=1048): la abre Enlaces, como un link.
        val ruta = n.ruta ?: return
        Enlaces.ruta(Uri.parse("https://tulocalweb.com.ar$ruta"))?.let { nav.ir(it) }
    }

    suspend fun leerTodo() {
        val l = carga.valor ?: return
        carga = Carga.listo(l.map { it.copy(leida = true) })
        intentar { estado.api.hacer("POST", "/v1/panel/notificaciones/leidas") }
        Haptico.exito()
    }

    BarraDe(acciones = { BotonBarra("Leí todo", habilitado = carga.valor?.any { !it.leida } ?: false) { alcance.launch { leerTodo() } } })

    Refrescable({ cargar() }) {
        LazyColumn(Modifier.fillMaxSize().background(C.fondo), contentPadding = PaddingValues(bottom = 24.dp)) {
            item { Texto("Avisos", Estilo.tituloGrande, C.texto, Modifier.margenPantalla().semantics { heading() }) }
            when (val c = carga) {
                Carga.cargando -> item { Box(Modifier.margenPantalla().padding(top = 24.dp)) { EsqueletoFilas(4, 36.dp) } }
                is Carga.error -> item { Box(Modifier.padding(top = 60.dp)) { EstadoError("los avisos", c.e) { estado.tarea { cargar() } } } }
                is Carga.listo -> if (c.v.isEmpty()) item {
                    Box(Modifier.padding(top = 60.dp)) { EstadoTincho(Tincho.durmiendo, "No tenés avisos", "Cuando entre una venta o pase algo en tu tienda, te avisamos acá.") }
                } else {
                    secciones(c.v).forEach { (titulo, items) ->
                        item(key = "t-$titulo") { Texto(titulo, Estilo.label, C.texto2, Modifier.padding(top = 14.dp, bottom = 6.dp).margenPantalla()) }
                        item(key = "s-$titulo") {
                            Column(Modifier.margenPantalla().fillMaxWidth().tarjeta()) {
                                items.forEachIndexed { i, n ->
                                    if (i > 0) Separador(62.dp)
                                    Box(Modifier.clickable { alcance.launch { tocar(n) } }) { FilaAviso(n) }
                                }
                            }
                        }
                    }
                    if (siguiente != null) item {
                        Box(Modifier.fillMaxWidth().padding(20.dp), contentAlignment = Alignment.Center) { Rueda() }
                        LaunchedEffect(c.v.size) { cargarMas() }
                    }
                }
            }
        }
    }
    LaunchedEffect(estado.version) { cargar() }
}

@Composable
private fun FilaAviso(n: Notificacion) {
    val (fondo, punto) = coloresAviso(n.tipo)
    Row(Modifier.fillMaxWidth().padding(horizontal = 14.dp, vertical = 12.dp), horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.Top) {
        Box(Modifier.size(36.dp).clip(CircleShape).background(fondo), contentAlignment = Alignment.Center) { Box(Modifier.size(10.dp).clip(CircleShape).background(punto)) }
        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Texto(n.titulo, Estilo(15f, if (n.leida) Peso.semibold else Peso.bold), C.texto, Modifier.weight(1f).padding(end = 6.dp))
                Texto(Formato.horaDe(n.creada), Estilo.footnote, C.texto2, numeros = true)
            }
            Texto(n.detalle, Estilo.texto14, C.neutroTinta, interlineado = 1f)
        }
        if (!n.leida) Box(Modifier.padding(top = 5.dp).size(9.dp).clip(CircleShape).background(C.acento).semantics { contentDescription = "Sin leer" })
    }
}

/** Fondo y punto por tipo (lib/marcas.ts): ventas y pagos en verde, avisos en ámbar, errores en terracota. */
@Composable
fun coloresAviso(tipo: String): Pair<Color, Color> = when (tipo) {
    "web", "mp", "ok", "pos" -> C.okFondo to C.ok
    "transf", "ml", "warn", "star" -> C.avisoFondo to C.aviso
    "err", "arca" -> C.errorFondo to C.error
    else -> C.pista to C.texto3
}

private fun secciones(l: List<Notificacion>): List<Pair<String, List<Notificacion>>> {
    val r = linkedMapOf<String, MutableList<Notificacion>>()
    l.forEach { n -> r.getOrPut(Formato.seccion(n.creada)) { mutableListOf() }.add(n) }
    return r.map { it.key to it.value }
}

