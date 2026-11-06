package ar.com.tulocalweb.app.pantallas.compartir

import android.content.Context
import android.graphics.Bitmap
import android.graphics.BitmapFactory
import android.graphics.LinearGradient
import android.graphics.Paint
import android.graphics.RectF
import android.graphics.Shader
import android.text.Layout
import android.text.StaticLayout
import android.text.TextPaint
import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxSize
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
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.asImageBitmap
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.SpanStyle
import androidx.compose.ui.text.buildAnnotatedString
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.withStyle
import androidx.compose.ui.unit.dp
import androidx.core.content.res.ResourcesCompat
import ar.com.tulocalweb.app.R
import ar.com.tulocalweb.app.app.Destino
import ar.com.tulocalweb.app.app.Estado
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.app.LocalNavegacion
import ar.com.tulocalweb.app.nucleo.Acceso
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.PaginaProductos
import ar.com.tulocalweb.app.nucleo.ProductoFila
import ar.com.tulocalweb.app.nucleo.Tablero
import ar.com.tulocalweb.app.nucleo.Tincho
import ar.com.tulocalweb.app.nucleo.ValorJSON
import ar.com.tulocalweb.app.nucleo.abrirUrl
import ar.com.tulocalweb.app.nucleo.archivoParaCompartir
import ar.com.tulocalweb.app.nucleo.compartirArchivo
import ar.com.tulocalweb.app.nucleo.compartirTexto
import ar.com.tulocalweb.app.nucleo.get
import ar.com.tulocalweb.app.nucleo.objeto
import ar.com.tulocalweb.app.nucleo.opcional
import ar.com.tulocalweb.app.nucleo.texto
import ar.com.tulocalweb.app.pantallas.mas.CuponFila
import ar.com.tulocalweb.app.pantallas.mas.DatosCupones
import ar.com.tulocalweb.app.pantallas.mas.GrupoFilas
import ar.com.tulocalweb.app.pantallas.mas.PantallaSeccion
import ar.com.tulocalweb.app.pantallas.mas.QRTienda
import ar.com.tulocalweb.app.pantallas.mas.resumenCupon
import ar.com.tulocalweb.app.tema.*
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import java.io.FileOutputStream
import java.net.URL
import kotlin.math.roundToInt

// Clon de Pantallas/Compartir/Compartir.swift.

/** Compartir la tienda: copiar el link, mandarlo por WhatsApp, la imagen para estados e historias, el QR, mensajes
 *  listos y el link de cada producto y cupón. Lo usan Inicio, Más, Productos y Mi tienda, para que el link no quede
 *  escondido. */
object Compartir {
    fun copiar(ctx: Context, url: String, estado: Estado, aviso: String = "Copiamos el link de tu tienda") {
        ar.com.tulocalweb.app.nucleo.copiar(ctx, url)
        Haptico.exito()
        estado.avisar(aviso)
    }

    fun copiarTexto(ctx: Context, texto: String, estado: Estado, aviso: String = "Copiamos el mensaje") {
        ar.com.tulocalweb.app.nucleo.copiar(ctx, texto)
        Haptico.exito()
        estado.avisar(aviso)
    }

    /** WhatsApp con el mensaje escrito; la persona elige a quién (o «Mi estado»). */
    fun whatsapp(texto: String): String = "https://wa.me/?text=" + java.net.URLEncoder.encode(texto, "UTF-8").replace("+", "%20")

    fun mensajeTienda(t: Acceso.Tienda): String = "¡Hola! Te comparto mi tienda online, ${t.nombre}. Mirá lo que hay y pedí desde ahí: ${t.url}"

    fun mensajeProducto(nombre: String, precio: Int?, url: String): String = "$nombre${precio?.let { " a ${Formato.pesos(it)}" } ?: ""}. Lo pedís acá: $url"

    data class Mensaje(val id: String, val titulo: String, val texto: String)

    /** Los mensajes listos para copiar o mandar (el link siempre al final, así WhatsApp arma la vista previa). */
    fun mensajes(t: Acceso.Tienda): List<Mensaje> {
        val u = t.url
        return listOf(
            Mensaje("abrimos", "Para avisar que abriste", "¡Ya tenemos tienda online! 🎉 Ahora podés ver todo lo de ${t.nombre} y pedir desde el celu: $u"),
            Mensaje("novedades", "Para tus estados", "Llegaron novedades a ${t.nombre} 👀 Pasá a mirarlas: $u"),
            Mensaje("bio", "Para la bio de Instagram", "Pedí online 👉 $u"),
            Mensaje("cliente", "Para un cliente", "¡Gracias por tu compra! La próxima podés pedir directo desde acá: $u"),
        )
    }
}

private data class DatosCompartir(val visitas: Tablero.Visitas?, val productos: List<ProductoFila>, val cupones: List<CuponFila>)

/** La pantalla «Compartí tu tienda». */
@Composable
fun CompartirTienda() {
    val estado = LocalEstado.current
    val nav = LocalNavegacion.current
    val ctx = LocalContext.current
    val t = estado.tienda ?: return
    var qr by remember { mutableStateOf(false) }
    var historia by remember { mutableStateOf(false) }
    val conCupones = estado.puede("catalogo")

    PantallaSeccion("Compartí tu tienda", "tu tienda", {
        // Uno detrás del otro, como iOS (allá `async let` cerraba la app en Release).
        val v = estado.api.opcional<Tablero>("/v1/panel/tablero", listOf("periodo" to "7d"))
        val p = estado.api.opcional<PaginaProductos>("/v1/panel/listas/productos", listOf("seg" to "active", "orden" to "sold", "cuantos" to "12"))?.filas ?: emptyList()
        val c = if (conCupones) estado.api.opcional<DatosCupones>("/v1/panel/cupones")?.cupones?.filter { it.estado == "activo" } ?: emptyList() else emptyList()
        DatosCompartir(v?.visitas, p, c)
    }) { d, _ ->
        // El link
        TarjetaHeroe("El link de tu tienda", t.nombre, Formato.sinViudas(textoVisitas(d.visitas)), Tincho.whatsapp) {
            Texto(t.direccion, Estilo.subheadFuerte, Color.White, Modifier.padding(end = 70.dp), lineas = 1, escalaMinima = 0.7f)
            Row(Modifier.padding(top = 6.dp), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                Row(Modifier.height(38.dp).clip(RoundedCornerShape(11.dp)).background(Color.White).clickable { Compartir.copiar(ctx, t.url, estado) }.padding(horizontal = 14.dp),
                    horizontalArrangement = Arrangement.spacedBy(6.dp), verticalAlignment = Alignment.CenterVertically) {
                    Icono("doc.on.doc", 14f, C.primarioTinta); Texto("Copiar link", Estilo.label, C.primarioTinta)
                }
                Row(Modifier.height(38.dp).clip(RoundedCornerShape(11.dp)).background(Color.White.copy(alpha = 0.18f)).clickable { nav.web = t.url }.padding(horizontal = 14.dp),
                    horizontalArrangement = Arrangement.spacedBy(6.dp), verticalAlignment = Alignment.CenterVertically) {
                    Icono("arrow.up.right.square", 14f, Color.White); Texto("Ver", Estilo.label, Color.White)
                }
            }
        }
        // Por dónde
        Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
            Box(Modifier.padding(top = 10.dp)) { TituloSeccion("Mandala por") }
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                AccionRapida("message", "WhatsApp", Modifier.weight(1f)) { abrirUrl(ctx, Compartir.whatsapp(Compartir.mensajeTienda(t))) }
                AccionRapida("photo.on.rectangle", "Historia", Modifier.weight(1f)) { historia = true }
                AccionRapida("qrcode", "QR", Modifier.weight(1f)) { qr = true }
                AccionRapida("square.and.arrow.up", "Más", Modifier.weight(1f)) { compartirTexto(ctx, Compartir.mensajeTienda(t)) }
            }
        }
        // Mensajes listos
        Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
            Box(Modifier.padding(top = 10.dp)) { TituloSeccion("Mensajes listos") }
            Column(Modifier.fillMaxWidth().tarjeta()) {
                Compartir.mensajes(t).forEachIndexed { i, m ->
                    if (i > 0) Separador()
                    Column(Modifier.fillMaxWidth().padding(16.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                        Texto(m.titulo, Estilo.footnoteFuerte, C.texto2)
                        Texto(m.texto, Estilo.texto14, C.texto)
                        Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                            BotonTonal("Copiar", { Compartir.copiarTexto(ctx, m.texto, estado) })
                            BotonTonal("WhatsApp", { abrirUrl(ctx, Compartir.whatsapp(m.texto)) })
                        }
                    }
                }
            }
        }
        // Productos
        Box(Modifier.padding(top = 10.dp)) { TituloSeccion("Compartí un producto") }
        if (d.productos.isEmpty()) {
            Row(Modifier.fillMaxWidth().tarjeta().padding(horizontal = 16.dp, vertical = 13.dp), horizontalArrangement = Arrangement.spacedBy(13.dp), verticalAlignment = Alignment.CenterVertically) {
                Box(Modifier.size(40.dp).clip(RoundedCornerShape(12.dp)).background(C.crema), contentAlignment = Alignment.Center) { Icono("shippingbox", 16f, C.primarioTexto) }
                Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(1.dp)) {
                    Texto("Todavía no hay productos publicados", Estilo.fila, C.texto)
                    Texto("Con una foto alcanza", Estilo.footnote, C.texto2)
                }
                BotonTonal("Cargar", { nav.abrir(Destino.nuevoProducto(null)) })
            }
        } else {
            Row(Modifier.fillMaxWidth().sinMargen().horizontalScroll(rememberScrollState()).padding(horizontal = margen), horizontalArrangement = Arrangement.spacedBy(10.dp), verticalAlignment = Alignment.Top) {
                d.productos.forEach { p ->
                    val u = t.url(p.id)
                    Column(Modifier.width(140.dp), verticalArrangement = Arrangement.spacedBy(7.dp)) {
                        Miniatura(p.foto, 140.dp, 120.dp, 16.dp)
                        Box(Modifier.height(40.dp)) { Texto(p.nombre, Estilo(14f, Peso.semibold), C.texto, lineas = 2) }
                        Texto(Formato.pesos(p.precio), Estilo.footnote, C.texto2, numeros = true)
                        Row(horizontalArrangement = Arrangement.spacedBy(6.dp)) {
                            Box(Modifier.weight(1f).semantics { contentDescription = "Compartir ${p.nombre}" }) {
                                BotonTonal("", { compartirTexto(ctx, Compartir.mensajeProducto(p.nombre, p.precio, u)) }, Modifier.fillMaxWidth(), simbolo = "square.and.arrow.up")
                            }
                            Box(Modifier.weight(1f).semantics { contentDescription = "Copiar el link de ${p.nombre}" }) {
                                BotonTonal("", { Compartir.copiar(ctx, u, estado, "Copiamos el link del producto") }, Modifier.fillMaxWidth(), simbolo = "doc.on.doc")
                            }
                        }
                    }
                }
            }
        }
        // Cupones
        if (d.cupones.isNotEmpty()) {
            Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
                Box(Modifier.padding(top = 10.dp)) { TituloSeccion("Tus cupones activos") }
                GrupoFilas(d.cupones, { it.id }) { c ->
                    val texto = "Usá el código ${c.codigo} en ${t.nombre} y tenés ${resumenCupon(c).lowercase()}: ${t.url}"
                    Row(Modifier.fillMaxWidth().padding(14.dp), horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
                        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(3.dp)) {
                            Texto(c.codigo, Estilo(16f, Peso.bold), C.texto, monoespaciado = true)
                            Texto(resumenCupon(c), Estilo.footnote, C.texto2)
                        }
                        BotonTonal("Compartir", { compartirTexto(ctx, texto) })
                    }
                }
            }
        }
    }
    Hoja(qr, { qr = false }) { QRTienda(t) }
    Hoja(historia, { historia = false }) { HojaHistoria(t) }
}

private fun textoVisitas(v: Tablero.Visitas?): String {
    if (v == null || v.desde == null) return "Compartila y empezá a sumar visitas."
    val n = v.dias7.actual.toInt()
    if (n == 0) return "Nadie la visitó en 7 días: es el momento de compartirla."
    return if (n == 1) "1 persona la visitó en los últimos 7 días." else "${Formato.numero(n)} personas la visitaron en los últimos 7 días."
}

// MARK: - Imagen para estados e historias

/** La hoja con la imagen 9:16 (1080 × 1920) para estados de WhatsApp e historias de Instagram: la portada de Mi tienda
 *  (si hay), el nombre, el QR y el link. */
@Composable
fun HojaHistoria(tienda: Acceso.Tienda) {
    val estado = LocalEstado.current
    val ctx = LocalContext.current
    var imagen by remember { mutableStateOf<Bitmap?>(null) }

    LaunchedEffect(Unit) {
        var portada: Bitmap? = null; var logo: Bitmap? = null
        estado.api.opcional<ValorJSON>("/v1/panel/mi-tienda")?.get("diseno")?.objeto?.let { d ->
            portada = bajar(d["portada"]?.texto); logo = bajar(d["logo"]?.texto)
        }
        imagen = withContext(Dispatchers.Default) { TarjetaHistoria.dibujar(ctx, tienda, portada, logo) }
    }

    Column(Modifier.fillMaxSize().padding(horizontal = 20.dp).padding(bottom = 12.dp), horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(14.dp)) {
        Grabber()
        Texto("Para tus estados e historias", Estilo.tituloHojaChico, C.texto)
        Box(Modifier.weight(1f), contentAlignment = Alignment.Center) {
            val img = imagen
            if (img != null) {
                val forma = RoundedCornerShape(18.dp)
                Image(img.asImageBitmap(), "Imagen de ${tienda.nombre} con el QR y el link",
                    Modifier.aspectRatio(9f / 16f).sombra(forma, hex(0x2A140C, 0.2f), 16.dp, y = 10.dp).clip(forma))
            } else {
                Box(Modifier.aspectRatio(9f / 16f).clip(RoundedCornerShape(18.dp)).background(C.pista), contentAlignment = Alignment.Center) { Rueda(C.acento) }
            }
        }
        imagen?.let { img ->
            BotonPrincipal("Compartir o guardar", {
                val f = archivoParaCompartir(ctx, "historia-${tienda.slug}.png")
                FileOutputStream(f).use { img.compress(Bitmap.CompressFormat.PNG, 100, it) }
                compartirArchivo(ctx, f, "image/png")
            }, simbolo = "square.and.arrow.up")
        }
        Texto(Formato.sinViudas("Subila a tu estado de WhatsApp o a tu historia y pegá el link en la bio."), Estilo.footnote, C.texto2, alinear = TextAlign.Center)
    }
}

private suspend fun bajar(s: String?): Bitmap? {
    if (s.isNullOrEmpty()) return null
    return withContext(Dispatchers.IO) { runCatching { URL(s).openStream().use { BitmapFactory.decodeStream(it) } }.getOrNull() }
}

/** La imagen en sí (360 × 640 puntos, a 3×). Colores fijos: no depende del modo oscuro. */
object TarjetaHistoria {
    fun dibujar(ctx: Context, tienda: Acceso.Tienda, portada: Bitmap?, logo: Bitmap?): Bitmap {
        val e = 3f
        val ancho = (360 * e).toInt(); val alto = (640 * e).toInt()
        val bmp = Bitmap.createBitmap(ancho, alto, Bitmap.Config.ARGB_8888)
        val c = android.graphics.Canvas(bmp)
        val p = Paint(Paint.ANTI_ALIAS_FLAG)
        p.shader = LinearGradient(0f, 0f, 0f, alto.toFloat(), 0xFFD9652F.toInt(), 0xFFB0441D.toInt(), Shader.TileMode.CLAMP)
        c.drawRect(0f, 0f, ancho.toFloat(), alto.toFloat(), p)
        p.shader = null
        val arriba = 300 * e
        if (portada != null) {
            // scaledToFill, recortada a 360 × 300.
            val esc = maxOf(ancho / portada.width.toFloat(), arriba / portada.height)
            val w = portada.width * esc; val h = portada.height * esc
            c.save(); c.clipRect(0f, 0f, ancho.toFloat(), arriba)
            c.drawBitmap(portada, null, RectF((ancho - w) / 2, (arriba - h) / 2, (ancho + w) / 2, (arriba + h) / 2), p)
            p.shader = LinearGradient(0f, arriba / 2, 0f, arriba, 0x00B0441D, 0xE6B0441D.toInt(), Shader.TileMode.CLAMP)
            c.drawRect(0f, arriba / 2, ancho.toFloat(), arriba, p); p.shader = null
            c.restore()
        } else {
            val t = BitmapFactory.decodeResource(ctx.resources, Tincho.saludo.imagen)
            val h = 220 * e; val w = t.width * h / t.height
            c.drawBitmap(t, null, RectF((ancho - w) / 2, arriba - 20 * e - h, (ancho + w) / 2, arriba - 20 * e), p)
        }
        fun fuente(id: Int) = ResourcesCompat.getFont(ctx, id)
        var y = arriba + (if (portada == null) 0f else 14 * e)
        if (logo != null) {
            val lado = 76 * e
            p.color = android.graphics.Color.WHITE
            c.drawRoundRect(RectF((ancho - lado) / 2, y, (ancho + lado) / 2, y + lado), 16 * e, 16 * e, p)
            c.drawBitmap(logo, null, RectF((ancho - lado) / 2 + 6 * e, y + 6 * e, (ancho + lado) / 2 - 6 * e, y + lado - 6 * e), p)
            y += lado + 12 * e
        }
        fun texto(s: String, tam: Float, id: Int, color: Int, lineas: Int = 1): Float {
            val tp = TextPaint(Paint.ANTI_ALIAS_FLAG).apply { textSize = tam * e; typeface = fuente(id); this.color = color }
            val l = StaticLayout.Builder.obtain(s, 0, s.length, tp, (ancho - 48 * e).toInt()).setAlignment(Layout.Alignment.ALIGN_CENTER).setMaxLines(lineas)
                .setEllipsize(android.text.TextUtils.TruncateAt.END).build()
            c.save(); c.translate(24 * e, y); l.draw(c); c.restore()
            return l.height.toFloat()
        }
        y += texto(tienda.nombre, 30f, R.font.instrument_bold, android.graphics.Color.WHITE, 2) + 12 * e
        y += texto("Ya podés comprar online", 17f, R.font.instrument_semibold, 0xE6FFFFFF.toInt()) + 16 * e
        QRTienda.bitmap(tienda.url)?.let { q ->
            val lado = 150 * e
            p.color = android.graphics.Color.WHITE
            c.drawRoundRect(RectF((ancho - lado) / 2, y, (ancho + lado) / 2, y + lado), 18 * e, 18 * e, p)
            p.isFilterBitmap = false
            c.drawBitmap(q, null, RectF((ancho - lado) / 2 + 10 * e, y + 10 * e, (ancho + lado) / 2 - 10 * e, y + lado - 10 * e), p)
            p.isFilterBitmap = true
            y += lado + 12 * e
        }
        // La dirección en una pastilla crema.
        val tp = TextPaint(Paint.ANTI_ALIAS_FLAG).apply { textSize = 15 * e; typeface = fuente(R.font.instrument_bold); color = 0xFF8A3414.toInt() }
        var dir = tienda.direccion
        while (tp.measureText(dir) > ancho - 80 * e && tp.textSize > 9 * e) tp.textSize -= e
        val wDir = tp.measureText(dir) + 32 * e; val hDir = tp.textSize + 18 * e
        p.color = 0xFFFBF3EA.toInt()
        c.drawRoundRect(RectF((ancho - wDir) / 2, y, (ancho + wDir) / 2, y + hDir), hDir / 2, hDir / 2, p)
        c.drawText(dir, (ancho - tp.measureText(dir)) / 2, y + hDir / 2 - (tp.descent() + tp.ascent()) / 2, tp)
        // Al pie.
        val pie = TextPaint(Paint.ANTI_ALIAS_FLAG).apply { textSize = 12 * e; typeface = fuente(R.font.instrument_medium); color = 0xBFFFFFFF.toInt() }
        val t = "Hecha con Tu Local Web"
        c.drawText(t, (ancho - pie.measureText(t)) / 2, alto - 18 * e - pie.descent(), pie)
        return bmp
    }
}

/** La tarjeta de la tienda online en Inicio: el link, las visitas del período y las acciones para compartirla (Copiar,
 *  WhatsApp, QR y la pantalla «Compartí tu tienda»). */
@Composable
fun TarjetaTiendaOnline(visitas: Tablero.Visitas?, periodo: String) {
    val estado = LocalEstado.current
    val nav = LocalNavegacion.current
    val ctx = LocalContext.current
    val t = estado.tienda ?: return
    var qr by remember { mutableStateOf(false) }
    Column(Modifier.fillMaxWidth().tarjeta(radio = 22.dp).padding(16.dp), verticalArrangement = Arrangement.spacedBy(14.dp)) {
        Row(horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
            AvatarTienda(t, 44.dp, 13.dp)
            Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                Row(horizontalArrangement = Arrangement.spacedBy(6.dp), verticalAlignment = Alignment.CenterVertically) {
                    Box(Modifier.size(7.dp).clip(CircleShape).background(C.ok))
                    Texto("Tu tienda está online", Estilo.footnoteFuerte, C.okTinta)
                }
                Texto(t.direccion, Estilo.fila, C.texto, lineas = 1, escalaMinima = 0.8f)
            }
            Box(Modifier.size(40.dp).clickable { nav.web = t.url }.semantics { contentDescription = "Ver tu tienda" }, contentAlignment = Alignment.Center) {
                Icono("arrow.up.right.square", 19f, C.primarioTexto)
            }
        }
        if (visitas != null) {
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.CenterVertically) {
                Icono("eye", 13f, C.primarioTexto)
                if (visitas.desde == null) Texto("Compartila para empezar a sumar visitas", Estilo.texto14, C.texto2)
                else {
                    val n = visitas.visitantes.actual.toInt()
                    val cuando = if (periodo == "hoy") "hoy" else if (periodo == "7d") "en 7 días" else "en 30 días"
                    val tx = C.texto; val ok = C.okTinta; val err = C.errorTinta
                    TextoRico(buildAnnotatedString {
                        withStyle(SpanStyle(fontWeight = FontWeight.Bold, color = tx)) { append(if (n == 1) "1 visita" else "${Formato.numero(n)} visitas") }
                        append(" $cuando")
                        visitas.visitantes.variacion?.let { d ->
                            val r = d.roundToInt()
                            withStyle(SpanStyle(fontWeight = FontWeight.Bold, color = if (r >= 0) ok else err)) { append(if (r >= 0) "  ▲ $r %" else "  ▼ ${kotlin.math.abs(r)} %") }
                        }
                    }, Estilo.texto14, C.texto2, numeros = true)
                }
            }
        }
        Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            AccionRapida("doc.on.doc", "Copiar", Modifier.weight(1f), 54.dp) { Compartir.copiar(ctx, t.url, estado) }
            AccionRapida("message", "WhatsApp", Modifier.weight(1f), 54.dp) { abrirUrl(ctx, Compartir.whatsapp(Compartir.mensajeTienda(t))) }
            AccionRapida("qrcode", "QR", Modifier.weight(1f), 54.dp) { qr = true }
            AccionRapida("square.and.arrow.up", "Compartir", Modifier.weight(1f), 54.dp) { nav.inicio.add(Destino.compartir) }
        }
    }
    Hoja(qr, { qr = false }) { QRTienda(t) }
}
