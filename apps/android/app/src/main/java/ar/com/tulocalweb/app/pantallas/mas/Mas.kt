package ar.com.tulocalweb.app.pantallas.mas

import android.content.Context
import android.graphics.Bitmap
import android.graphics.Paint
import android.text.Layout
import android.text.StaticLayout
import android.text.TextPaint
import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
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
import androidx.compose.ui.graphics.FilterQuality
import androidx.compose.ui.graphics.ImageBitmap
import androidx.compose.ui.graphics.asImageBitmap
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.core.content.res.ResourcesCompat
import ar.com.tulocalweb.app.BuildConfig
import ar.com.tulocalweb.app.R
import ar.com.tulocalweb.app.app.Destino
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.app.LocalNavegacion
import ar.com.tulocalweb.app.nucleo.Acceso
import ar.com.tulocalweb.app.nucleo.FechaApi
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.archivoParaCompartir
import ar.com.tulocalweb.app.nucleo.compartirArchivo
import ar.com.tulocalweb.app.nucleo.opcional
import ar.com.tulocalweb.app.pantallas.compartir.Compartir
import ar.com.tulocalweb.app.pantallas.ingreso.ElegirTienda
import ar.com.tulocalweb.app.tema.*
import com.google.zxing.BarcodeFormat
import com.google.zxing.EncodeHintType
import com.google.zxing.qrcode.QRCodeWriter
import com.google.zxing.qrcode.decoder.ErrorCorrectionLevel
import kotlinx.coroutines.launch
import kotlinx.serialization.Serializable
import java.io.FileOutputStream

// Clon de Pantallas/Mas/Mas.swift.

private data class FilaMas(
    val texto: String,
    val simbolo: String,
    val destino: Destino,
    val permiso: String? = null,
    val insignia: Int = 0,
    val detalle: Pair<String, Color>? = null,
)

/** 8a · Más: la tarjeta de la tienda (cambiar, ver, compartir, QR) y los grupos Ventas y clientes, Plata y Tu cuenta,
 *  con «Cerrar sesión». Todo tiene pantalla propia en la app. */
@Composable
fun Mas() {
    val estado = LocalEstado.current
    val nav = LocalNavegacion.current
    val ctx = LocalContext.current
    var cambiarTienda by remember { mutableStateOf(false) }
    var qr by remember { mutableStateOf(false) }
    var salir by remember { mutableStateOf(false) }
    var cajaAbierta by remember { mutableStateOf<Boolean?>(null) }

    @Serializable data class Caja(val abierta: FechaApi? = null)
    @Serializable data class T(val caja: Caja? = null)
    LaunchedEffect(Unit) {
        if (!estado.puede("finanzas")) return@LaunchedEffect
        estado.api.opcional<T>("/v1/panel/tablero", listOf("periodo" to "hoy"))?.let { t -> cajaAbierta = t.caja?.let { it.abierta != null } }
    }

    val plan: Pair<String, Color>? = estado.resumen?.prueba?.takeIf { it.diasRestantes > 0 }?.let { p ->
        "Prueba · ${p.diasRestantes} ${if (p.diasRestantes == 1) "día" else "días"}" to C.avisoTinta
    }
    val okTinta = C.okTinta; val texto2 = C.texto2

    @Composable
    fun Grupo(titulo: String, tono: Pair<Color, Color>, filas: List<FilaMas>) {
        val visibles = filas.filter { f -> f.permiso?.let { estado.puede(it) } ?: true }
        if (visibles.isEmpty()) return
        EtiquetaSeccion(titulo, Modifier.padding(horizontal = 4.dp).padding(top = 12.dp))
        Column(Modifier.fillMaxWidth().tarjeta()) {
            visibles.forEachIndexed { i, f ->
                if (i > 0) Separador(58.dp)
                Row(Modifier.fillMaxWidth().clickable { nav.mas.add(f.destino) }.padding(horizontal = 16.dp, vertical = 12.dp),
                    horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
                    Box(Modifier.size(30.dp).clip(RoundedCornerShape(8.dp)).background(tono.first), contentAlignment = Alignment.Center) { Icono(f.simbolo, 14f, tono.second) }
                    Texto(f.texto, Estilo.body16, C.texto, Modifier.weight(1f))
                    f.detalle?.let { Texto(it.first, Estilo.label, it.second) }
                    if (f.insignia > 0) Insignia(f.insignia)
                    Chevron(8f)
                }
            }
        }
    }

    Column(Modifier.fillMaxSize().background(C.fondo).verticalScroll(rememberScrollState()).margenPantalla().padding(bottom = 24.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
        Box(Modifier.padding(top = 4.dp)) { EncabezadoGrande("Más") }
        estado.tienda?.let { t ->
            val varias = (estado.acceso?.tiendas?.size ?: 0) > 1
            Column(Modifier.padding(top = 4.dp).fillMaxWidth().tarjeta(radio = 22.dp).padding(16.dp), verticalArrangement = Arrangement.spacedBy(14.dp)) {
                Row(Modifier.fillMaxWidth().clickable { if (varias) cambiarTienda = true }, horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
                    AvatarTienda(t, 50.dp, 14.dp)
                    Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                        Texto(t.nombre, Estilo.nombreTienda, C.texto, lineas = 1)
                        Texto(t.direccion, Estilo.texto14, C.texto2, lineas = 1)
                    }
                    if (varias) Texto("Cambiar", Estilo.label, C.primarioTexto)
                }
                Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    AccionRapida("arrow.up.right.square", "Ver", Modifier.weight(1f), 54.dp) { nav.web = t.url }
                    AccionRapida("doc.on.doc", "Copiar", Modifier.weight(1f), 54.dp) { Compartir.copiar(ctx, t.url, estado) }
                    AccionRapida("square.and.arrow.up", "Compartir", Modifier.weight(1f), 54.dp) { nav.mas.add(Destino.compartir) }
                    AccionRapida("qrcode", "QR", Modifier.weight(1f), 54.dp) { qr = true }
                }
            }
        }
        Grupo("Tu tienda online", C.crema to C.primarioTexto, listOf(
            FilaMas("Compartí tu tienda", "square.and.arrow.up", Destino.compartir),
            FilaMas("Editar mi tienda", "paintbrush.pointed", Destino.miTienda, "cuenta"),
            FilaMas("Estadísticas", "chart.bar.xaxis", Destino.estadisticas, "pedidos"),
            FilaMas("Traer mi tienda de otra plataforma", "arrow.down.doc", Destino.importar, "catalogo"),
        ))
        Grupo("Ventas y clientes", C.pista to C.primarioTexto, listOf(
            FilaMas("Clientes", "person.2", Destino.clientes, "clientes"),
            FilaMas("Caja del día", "banknote", Destino.caja, "pedidos", detalle = cajaAbierta?.let { if (it) "Abierta" to okTinta else "Cerrada" to texto2 }),
            FilaMas("Carritos abandonados", "cart", Destino.carritos, "clientes", insignia = estado.resumen?.insignias?.carritos ?: 0),
            FilaMas("Reseñas y preguntas", "star.bubble", Destino.resenas, "clientes"),
            FilaMas("Devoluciones", "arrow.uturn.backward", Destino.devoluciones, "pedidos"),
            FilaMas("Cupones", "ticket", Destino.cupones, "catalogo"),
            FilaMas("Links de pago", "link", Destino.links, "pedidos"),
            FilaMas("Inventario", "barcode.viewfinder", Destino.inventario, "catalogo"),
        ))
        Grupo("Plata", C.okFondo to C.okTinta, listOf(
            FilaMas("Finanzas", "chart.line.uptrend.xyaxis", Destino.finanzas, "finanzas"),
            FilaMas("Facturación de ARCA", "doc.text", Destino.facturacion, "finanzas"),
        ))
        Grupo("Tu cuenta", C.avisoFondo to C.avisoTinta, listOf(
            FilaMas("Plan", "creditcard", Destino.plan, "cuenta", detalle = plan),
            FilaMas("Datos del negocio y cobro", "building.2", Destino.negocio, "cuenta"),
            FilaMas("Integraciones", "puzzlepiece.extension", Destino.integraciones, "cuenta"),
            FilaMas("Usuarios y permisos", "person.badge.key", Destino.usuarios, "cuenta"),
            FilaMas("Perfil, seguridad y tema", "person.crop.circle", Destino.cuenta),
            FilaMas("Ayuda", "questionmark.circle", Destino.ayuda),
        ))
        val forma = RoundedCornerShape(16.dp)
        Box(Modifier.padding(top = 8.dp).fillMaxWidth().heightIn(min = 52.dp).clip(forma).background(C.superficie).border(1.dp, C.anillo, forma).clickable { salir = true },
            contentAlignment = Alignment.Center) { Texto("Cerrar sesión", Estilo.fila, C.errorTinta) }
        Texto("Tu Local Web ${BuildConfig.VERSION_NAME}", Estilo.footnote, C.texto3, Modifier.fillMaxWidth().padding(top = 8.dp), alinear = TextAlign.Center)
    }
    Hoja(cambiarTienda, { cambiarTienda = false }, altoFijo = 520.dp, estirable = true) { ElegirTienda { cambiarTienda = false } }
    Hoja(qr, { qr = false }) { estado.tienda?.let { QRTienda(it) } }
    Confirmacion(salir, "¿Cerrás la sesión en este teléfono?", "Dejás de recibir los avisos acá hasta que vuelvas a entrar.",
        listOf(OpcionConfirmacion("Cerrar sesión", destructiva = true) { estado.tarea { estado.salir() } }), { salir = false }, volver = "Cancelar")
}

/** El QR de la tienda (para mostrar en el mostrador, imprimirlo o mandarlo como imagen). */
@Composable
fun QRTienda(tienda: Acceso.Tienda) {
    val estado = LocalEstado.current
    val ctx = LocalContext.current
    var cartel by remember { mutableStateOf<Bitmap?>(null) }
    LaunchedEffect(Unit) { cartel = QRTienda.cartel(ctx, tienda) }
    Column(Modifier.fillMaxSize().padding(horizontal = 20.dp).padding(bottom = 12.dp), horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(16.dp)) {
        Grabber()
        Texto(tienda.nombre, Estilo.tituloHojaChico, C.texto)
        Texto(Formato.sinViudas("Que lo escaneen con la cámara y entran directo a tu tienda."), Estilo.subhead, C.texto2, alinear = TextAlign.Center)
        QRTienda.imagen(tienda.url)?.let { img ->
            val forma = RoundedCornerShape(28.dp)
            Box(Modifier.sombra(forma, hex(0x2A140C, 0.2f), 20.dp, y = 14.dp).clip(forma).background(Color.White).padding(18.dp)) {
                Image(img, null, Modifier.size(220.dp).semantics { contentDescription = "Código QR de ${tienda.direccion}" }, filterQuality = FilterQuality.None)
            }
        }
        Texto(tienda.direccion, Estilo.subheadFuerte, C.primarioTexto)
        Spacer(Modifier.weight(1f))
        cartel?.let { c ->
            BotonPrincipal("Imprimir, guardar o mandar", {
                val f = archivoParaCompartir(ctx, "qr-${tienda.slug}.png")
                FileOutputStream(f).use { c.compress(Bitmap.CompressFormat.PNG, 100, it) }
                compartirArchivo(ctx, f, "image/png")
            }, simbolo = "printer")
        }
        BotonTexto("Copiar link", { Compartir.copiar(ctx, tienda.url, estado) }, simbolo = "doc.on.doc")
    }
}

object QRTienda {
    /** El QR en blanco y negro, un píxel por módulo (se agranda sin suavizar). */
    fun bitmap(texto: String): Bitmap? = runCatching {
        val m = QRCodeWriter().encode(texto, BarcodeFormat.QR_CODE, 0, 0, mapOf(EncodeHintType.ERROR_CORRECTION to ErrorCorrectionLevel.M, EncodeHintType.MARGIN to 0))
        val esc = 10
        val bmp = Bitmap.createBitmap(m.width * esc, m.height * esc, Bitmap.Config.ARGB_8888)
        for (y in 0 until m.height) for (x in 0 until m.width) {
            val c = if (m[x, y]) android.graphics.Color.BLACK else android.graphics.Color.WHITE
            for (dy in 0 until esc) for (dx in 0 until esc) bmp.setPixel(x * esc + dx, y * esc + dy, c)
        }
        bmp
    }.getOrNull()

    fun imagen(texto: String): ImageBitmap? = bitmap(texto)?.asImageBitmap()

    /** El cartel para imprimir: el QR grande con el nombre y el link, en blanco (A6 vertical). */
    fun cartel(ctx: Context, t: Acceso.Tienda): Bitmap? {
        val qr = bitmap(t.url) ?: return null
        val e = 3f
        val ancho = (400 * e).toInt(); val alto = (566 * e).toInt()
        val bmp = Bitmap.createBitmap(ancho, alto, Bitmap.Config.ARGB_8888)
        val c = android.graphics.Canvas(bmp)
        c.drawColor(android.graphics.Color.WHITE)
        fun fuente(id: Int) = ResourcesCompat.getFont(ctx, id)
        val textos = mutableListOf<StaticLayout>()
        fun armar(s: String, tam: Float, id: Int, color: Int, lineas: Int = 3): StaticLayout {
            val tp = TextPaint(Paint.ANTI_ALIAS_FLAG).apply { textSize = tam * e; typeface = fuente(id); this.color = color }
            return StaticLayout.Builder.obtain(s, 0, s.length, tp, (ancho - 64 * e).toInt()).setAlignment(Layout.Alignment.ALIGN_CENTER).setMaxLines(lineas).build()
        }
        val nombre = armar(t.nombre, 28f, R.font.instrument_bold, 0xFF2A140C.toInt())
        val bajada = armar("Escaneá y comprá online", 17f, R.font.instrument_semibold, 0xFF6B5446.toInt())
        val dir = armar(t.direccion, 16f, R.font.instrument_bold, 0xFFB0441D.toInt(), 1)
        textos += listOf(nombre, bajada)
        val ladoQr = 260 * e
        val espacio = 18 * e
        val total = nombre.height + bajada.height + ladoQr + dir.height + espacio * 3
        var y = (alto - total) / 2
        fun poner(l: StaticLayout) { c.save(); c.translate(32 * e, y); l.draw(c); c.restore(); y += l.height + espacio }
        poner(nombre); poner(bajada)
        val p = Paint().apply { isFilterBitmap = false }
        c.drawBitmap(qr, null, android.graphics.RectF((ancho - ladoQr) / 2, y, (ancho + ladoQr) / 2, y + ladoQr), p)
        y += ladoQr + espacio
        poner(dir)
        return bmp
    }
}
