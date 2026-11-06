package ar.com.tulocalweb.app.pantallas.inicio

import androidx.compose.animation.core.RepeatMode
import androidx.compose.animation.core.animateFloat
import androidx.compose.animation.core.infiniteRepeatable
import androidx.compose.animation.core.rememberInfiniteTransition
import androidx.compose.animation.core.tween
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.scale
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.semantics.heading
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.SpanStyle
import androidx.compose.ui.text.buildAnnotatedString
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.withStyle
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.app.Destino
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.app.LocalNavegacion
import ar.com.tulocalweb.app.app.Pestana
import ar.com.tulocalweb.app.nucleo.Fecha
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.Marcas
import ar.com.tulocalweb.app.nucleo.Tablero
import ar.com.tulocalweb.app.nucleo.entero
import ar.com.tulocalweb.app.nucleo.get
import ar.com.tulocalweb.app.nucleo.opcional
import ar.com.tulocalweb.app.nucleo.texto
import ar.com.tulocalweb.app.pantallas.mas.ChipSimple
import ar.com.tulocalweb.app.pantallas.mas.Estrellas
import ar.com.tulocalweb.app.tema.*
import kotlinx.serialization.Serializable

// Clon de Pantallas/Inicio/InicioMas.swift: lo que el Inicio de la web mostraba y la app no: preguntarle a Tincho, quién
// está ahora en la tienda, la caja del local, las próximas liquidaciones, las integraciones, lo más visto, las reseñas
// nuevas y los envíos de hoy. Cada tarjeta aparece sólo si tiene algo que mostrar.

/** La barra «Preguntale a Tincho» (como la del Inicio de la web): lleva a Tincho con la pregunta. */
@Composable
fun BarraPreguntarTincho() {
    val nav = LocalNavegacion.current
    val ejemplos = listOf("¿Cuánto vendí ayer?", "¿Qué tengo que reponer?", "¿Cómo vengo esta semana?", "Ideas para vender más")
    Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
        val forma = RoundedCornerShape(16.dp)
        Row(Modifier.fillMaxWidth().clip(forma).background(Brush.verticalGradient(listOf(T.primarioArriba.actual.copy(alpha = 0.09f), hex(0xB0441D, 0.09f))))
            .clickable { nav.pestana = Pestana.tincho }.padding(12.dp), horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
            Box(Modifier.size(36.dp).clip(RoundedCornerShape(11.dp)).background(Brush.linearGradient(listOf(hex(0xE8875F), hex(0xC24E25)))), contentAlignment = Alignment.Center) {
                Icono("sparkles", 16f, Color.White)
            }
            Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(1.dp)) {
                Texto("Preguntale a Tincho", Estilo.fila, C.texto)
                Texto("«¿Cuánto vendí ayer por Mercado Libre?»", Estilo.footnote, C.texto2, lineas = 1)
            }
            Icono("arrow.up.circle.fill", 22f, C.primarioTexto)
        }
        Row(Modifier.fillMaxWidth().horizontalScroll(rememberScrollState()), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            ejemplos.forEach { e ->
                Texto(e, Estilo.footnoteFuerte, C.primarioTinta, Modifier.clip(Capsula).background(C.crema2).border(1.dp, C.linea2, Capsula)
                    .clickable { nav.preguntaTincho = e; nav.pestana = Pestana.tincho }.padding(horizontal = 12.dp, vertical = 7.dp), lineas = 1)
            }
        }
    }
}

/** «Ahora en tu tienda»: cuántas personas están mirando, en qué páginas y los carritos armándose. */
@Composable
fun TarjetaEnVivo(vivo: Tablero.EnVivo, montos: Boolean) {
    val latido = if (animacionesActivas()) rememberInfiniteTransition(label = "latido").animateFloat(0f, 1f, infiniteRepeatable(tween(900), RepeatMode.Reverse), label = "l").value else 0f
    Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
        Row(Modifier.padding(top = 14.dp), verticalAlignment = Alignment.CenterVertically) {
            Texto("Ahora en tu tienda", Estilo.tituloSeccion, C.texto, Modifier.weight(1f).semantics { heading() })
            Row(horizontalArrangement = Arrangement.spacedBy(6.dp), verticalAlignment = Alignment.CenterVertically) {
                Box(Modifier.size(8.dp).scale(0.85f + 0.5f * latido).alpha(1f - 0.4f * latido).clip(CircleShape).background(C.ok))
                Texto("En vivo", Estilo.footnoteFuerte, C.okTinta)
            }
        }
        Column(Modifier.fillMaxWidth().tarjeta().padding(16.dp), verticalArrangement = Arrangement.spacedBy(10.dp)) {
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.Bottom) {
                Texto("${maxOf(0, vivo.personas)}", Estilo.saludo, C.texto, numeros = true)
                Texto(if (vivo.personas == 1) "persona mirando" else "personas mirando", Estilo.subhead, C.texto2, Modifier.padding(bottom = 5.dp))
            }
            if (vivo.paginas.isNotEmpty()) {
                Column {
                    vivo.paginas.take(4).forEachIndexed { i, p ->
                        if (i > 0) Separador(0.dp)
                        Row(Modifier.padding(vertical = 8.dp)) {
                            Texto(p.nombre, Estilo.texto14, C.texto, Modifier.weight(1f), lineas = 1)
                            Texto("${p.personas}", Estilo.fila, C.texto2, numeros = true)
                        }
                    }
                }
            }
            if (vivo.carritos > 0) {
                TextoRico(buildAnnotatedString {
                    withStyle(SpanStyle(fontWeight = FontWeight.Bold)) { append(Formato.cantidad(vivo.carritos, "carrito", "carritos")) }
                    append(" armándose ahora")
                    if (montos) vivo.carritosMonto?.let { withStyle(SpanStyle(fontWeight = FontWeight.Bold)) { append(" por ${Formato.pesos(it)}") } }
                }, Estilo.texto14, C.texto, Modifier.fillMaxWidth().clip(RoundedCornerShape(12.dp)).background(C.avisoFondo).padding(10.dp))
            }
        }
    }
}

/** «Caja del local»: lo cobrado hoy en el punto de venta, por medio. */
@Composable
fun TarjetaCajaLocal(caja: Tablero.Caja) {
    val nav = LocalNavegacion.current
    Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
        Box(Modifier.padding(top = 14.dp)) { TituloSeccion("Caja del local", if (caja.abierta == null) "Abrir" else "Ver") { nav.inicio.add(Destino.caja) } }
        Column(Modifier.fillMaxWidth().tarjeta().padding(16.dp), verticalArrangement = Arrangement.spacedBy(10.dp)) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Texto(Formato.pesos(caja.total), Estilo.montoPantalla, C.texto, Modifier.weight(1f), numeros = true)
                ChipSimple(if (caja.abierta == null) "Cerrada" else "Abierta", if (caja.abierta == null) C.pista else C.okFondo, if (caja.abierta == null) C.texto2 else C.okTinta)
            }
            if (caja.medios.isNotEmpty()) {
                val total = maxOf(1, caja.medios.sumOf { it.monto })
                BoxWithConstraints(Modifier.fillMaxWidth().height(8.dp).clip(Capsula)) {
                    Row(horizontalArrangement = Arrangement.spacedBy(2.dp)) {
                        caja.medios.forEachIndexed { i, m ->
                            Box(Modifier.width(maxOf(4.dp, this@BoxWithConstraints.maxWidth * m.monto / total)).fillMaxHeight().background(C.acento.copy(alpha = 1f - i * 0.22f)))
                        }
                    }
                }
                caja.medios.forEachIndexed { i, m ->
                    Row(horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.CenterVertically) {
                        Box(Modifier.size(8.dp).clip(CircleShape).background(C.acento.copy(alpha = 1f - i * 0.22f)))
                        Texto(Marcas.pago(m.medio), Estilo.texto14, C.texto, Modifier.weight(1f))
                        Texto(Formato.pesos(m.monto), Estilo.fila, C.texto, numeros = true)
                    }
                }
            }
        }
    }
}

/** «Próximas liquidaciones»: cuándo y cuánto deposita cada medio. */
@Composable
fun TarjetaLiquidaciones(liquidaciones: List<Tablero.Liquidacion>) {
    val nav = LocalNavegacion.current
    Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
        Box(Modifier.padding(top = 14.dp)) { TituloSeccion("Próximas liquidaciones", "Finanzas") { nav.inicio.add(Destino.finanzas) } }
        Column(Modifier.fillMaxWidth().tarjeta()) {
            liquidaciones.take(4).forEachIndexed { i, l ->
                if (i > 0) Separador(16.dp)
                Row(Modifier.padding(horizontal = 16.dp, vertical = 12.dp), verticalAlignment = Alignment.CenterVertically) {
                    Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                        Texto(Marcas.pago(l.medio), Estilo.fila, C.texto)
                        Texto(Fecha.iso(l.fecha)?.let { Formato.dia(it) } ?: l.fecha, Estilo.footnote, C.texto2)
                    }
                    Texto(Formato.pesos(l.monto), Estilo.fila, C.okTinta, numeros = true)
                }
            }
        }
    }
}

/** «Más vistos»: los productos que más miraron en la tienda online. */
@Composable
fun TarjetaMasVistos(visitas: Tablero.Visitas) {
    Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
        Row(Modifier.padding(top = 14.dp), verticalAlignment = Alignment.Bottom) {
            Texto("Más vistos", Estilo.tituloSeccion, C.texto, Modifier.weight(1f).semantics { heading() })
            Texto("En la tienda online", Estilo.subhead, C.texto2)
        }
        val max = (visitas.masVistos.maxOfOrNull { it.vistas } ?: 1).coerceAtLeast(1).toFloat()
        Column(Modifier.fillMaxWidth().tarjeta()) {
            visitas.masVistos.take(5).forEachIndexed { i, p ->
                if (i > 0) Separador(68.dp)
                Row(Modifier.padding(horizontal = 14.dp, vertical = 11.dp), horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
                    Miniatura(p.foto, 42.dp, radio = 11.dp)
                    Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(5.dp)) {
                        Texto(p.nombre, Estilo.fila, C.texto, lineas = 1)
                        BoxWithConstraints(Modifier.fillMaxWidth().height(5.dp)) {
                            Box(Modifier.width(maxWidth * (p.vistas / max)).fillMaxHeight().clip(Capsula).background(C.acento.copy(alpha = 0.5f)))
                        }
                    }
                    Texto(Formato.numero(p.vistas), Estilo.fila, C.texto2, numeros = true)
                }
            }
        }
    }
}

/** «Reseñas nuevas»: las últimas sin responder. */
@Composable
fun TarjetaResenas(resenas: Tablero.Resenas) {
    val nav = LocalNavegacion.current
    Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
        Box(Modifier.padding(top = 14.dp)) { TituloSeccion("Reseñas nuevas", "Responder") { nav.inicio.add(Destino.resenas) } }
        Column(Modifier.fillMaxWidth().tarjeta()) {
            resenas.promedio?.let { p ->
                Row(Modifier.padding(horizontal = 16.dp, vertical = 12.dp), horizontalArrangement = Arrangement.spacedBy(6.dp), verticalAlignment = Alignment.CenterVertically) {
                    Texto(String.format(Formato.argentina, "%.1f", p), Estilo.headline, C.texto)
                    Estrellas(Math.round(p).toInt())
                    Texto("promedio", Estilo.footnote, C.texto2)
                }
            }
            resenas.nuevas.take(3).forEach { r ->
                Separador(16.dp)
                Column(Modifier.padding(horizontal = 16.dp, vertical = 12.dp), verticalArrangement = Arrangement.spacedBy(4.dp)) {
                    Row(horizontalArrangement = Arrangement.spacedBy(6.dp), verticalAlignment = Alignment.CenterVertically) {
                        Estrellas(r["estrellas"]?.entero ?: 0)
                        Texto(r["autor"]?.texto ?: "", Estilo.footnoteFuerte, C.texto)
                        Spacer(Modifier.weight(1f))
                        Texto(r["producto"]?.texto ?: "", Estilo.footnote, C.texto2, lineas = 1)
                    }
                    Texto("«${r["texto"]?.texto ?: ""}»", Estilo.texto14, C.texto, lineas = 3)
                }
            }
        }
    }
}

/** «Envíos de hoy»: paquetes y retiros de los transportes. */
@Composable
fun TarjetaEnvios(envios: Tablero.Envios) {
    val nav = LocalNavegacion.current
    Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
        Box(Modifier.padding(top = 14.dp)) { TituloSeccion("Envíos de hoy", "Preparar") { nav.filtroPedidos = "Para preparar"; nav.pestana = Pestana.pedidos } }
        Column(Modifier.fillMaxWidth().tarjeta()) {
            Row(Modifier.padding(horizontal = 16.dp, vertical = 12.dp), horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.CenterVertically) {
                Icono("shippingbox.fill", 17f, C.primarioTexto)
                Texto(Formato.cantidad(envios.paquetes, "paquete para despachar", "paquetes para despachar"), Estilo.fila, C.texto)
            }
            envios.retiros.forEach { r ->
                Separador(16.dp)
                Row(Modifier.padding(horizontal = 16.dp, vertical = 11.dp), horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.CenterVertically) {
                    Texto(r.hora, Estilo.fila, if (r.proximo) C.primarioTexto else C.texto, Modifier.width(54.dp), numeros = true)
                    Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(1.dp)) {
                        Texto(Formato.mayuscula(r.transporte), Estilo.texto14, C.texto)
                        r.ciudad?.let { Texto(it, Estilo.footnote, C.texto2) }
                    }
                    Texto(Formato.cantidad(r.paquetes, "paquete", "paquetes"), Estilo.footnote, C.texto2)
                    if (r.proximo) ChipSimple("Próximo", C.avisoFondo, C.avisoTinta)
                }
            }
        }
    }
}

/** «Integraciones»: lo conectado (Mercado Pago, Mercado Libre, Correo) y un acceso para sumar más. */
@Composable
fun TarjetaIntegraciones() {
    val estado = LocalEstado.current
    val nav = LocalNavegacion.current
    var conectadas by remember { mutableStateOf<List<Pair<String, String>>?>(null) }

    @Serializable data class I(val conectada: Boolean? = null, val usuario: String? = null, val apodo: String? = null, val email: String? = null)

    LaunchedEffect(Unit) {
        if (conectadas != null) return@LaunchedEffect
        val c = mutableListOf<Pair<String, String>>()
        for ((id, nombre) in listOf("mercadopago" to "Mercado Pago", "mercadolibre" to "Mercado Libre", "correo" to "Correo Argentino")) {
            val i = estado.api.opcional<I>("/v1/panel/integraciones/$id")
            if (i?.conectada == true) c.add(nombre to (i.usuario ?: i.apodo ?: i.email ?: "Conectada"))
        }
        conectadas = c
    }

    Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
        Box(Modifier.padding(top = 14.dp)) { TituloSeccion("Integraciones", "Explorar") { nav.inicio.add(Destino.integraciones) } }
        Column(Modifier.fillMaxWidth().tarjeta()) {
            val c = conectadas
            when {
                c != null && c.isNotEmpty() -> c.forEachIndexed { i, x ->
                    if (i > 0) Separador(16.dp)
                    Row(Modifier.padding(horizontal = 16.dp, vertical = 12.dp), horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.CenterVertically) {
                        Texto(x.first, Estilo.fila, C.texto, Modifier.weight(1f))
                        Texto(x.second, Estilo.footnote, C.texto2, lineas = 1)
                        Box(Modifier.size(8.dp).clip(CircleShape).background(C.ok))
                    }
                }
                c != null -> Texto(Formato.sinViudas("Conectá Mercado Pago, Mercado Libre y Correo Argentino para cobrar, publicar y enviar."), Estilo.texto14, C.texto2,
                    Modifier.fillMaxWidth().clickable { nav.inicio.add(Destino.integraciones) }.padding(16.dp))
                else -> Box(Modifier.padding(16.dp)) { Rueda(C.acento) }
            }
        }
    }
}
