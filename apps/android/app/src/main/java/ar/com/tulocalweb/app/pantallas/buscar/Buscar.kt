package ar.com.tulocalweb.app.pantallas.buscar

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.itemsIndexed
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
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.app.Destino
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.app.LocalNavegacion
import ar.com.tulocalweb.app.app.Navegacion
import ar.com.tulocalweb.app.app.Pestana
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.tema.*
import java.text.Normalizer

// Clon de Pantallas/Buscar/Buscar.swift.

private data class ItemBuscar(val titulo: String, val detalle: String, val simbolo: String, val grupo: String, val permiso: String? = null, val accion: (Navegacion) -> Unit) {
    val id: String get() = grupo + titulo
}

/** Sin tildes ni mayúsculas (folding .diacriticInsensitive, .caseInsensitive). */
fun normal(s: String): String = Normalizer.normalize(s, Normalizer.Form.NFD).replace(Regex("\\p{Mn}+"), "").lowercase(Formato.argentina).trim()

/** El buscador (como el ⌘K del panel web): preguntarle a Tincho, las acciones de todos los días y todas las secciones.
 *  Escribiendo se filtra (sin tildes ni mayúsculas); «Enter» le pregunta a Tincho. */
@Composable
fun Buscar() {
    val estado = LocalEstado.current
    val nav = LocalNavegacion.current
    val cerrar = LocalCerrarHoja.current
    var q by remember { mutableStateOf("") }
    val foco = remember { FocusRequester() }

    fun ir(titulo: String, detalle: String, simbolo: String, d: Destino, permiso: String? = null) =
        ItemBuscar(titulo, detalle, simbolo, "Ir a", permiso) { it.inicio.add(d); it.pestana = Pestana.inicio }

    val items = remember(estado.acceso) {
        val acciones = listOf(
            ItemBuscar("Crear pedido", "Manual, por WhatsApp o en el local", "plus", "Acciones", "pedidos") { it.inicio.add(Destino.pedidoManual); it.pestana = Pestana.inicio },
            ItemBuscar("Crear link de pago", "Cobrá por WhatsApp en segundos", "link", "Acciones", "pedidos") { it.inicio.add(Destino.links); it.pestana = Pestana.inicio },
            ItemBuscar("Cargar producto con foto", "Subí las fotos y completá nombre y precio", "camera", "Acciones", "catalogo") { it.inicio.add(Destino.nuevoProducto(null)); it.pestana = Pestana.inicio },
            ItemBuscar("Cobrar en el local", "Punto de venta", "barcode.viewfinder", "Acciones") { it.venta = true },
            ItemBuscar("Compartir mi tienda", "El link, por WhatsApp o redes", "storefront", "Acciones") { it.inicio.add(Destino.compartir); it.pestana = Pestana.inicio },
        )
        val secciones = listOf(
            ItemBuscar("Resumen", "Inicio", "square.grid.2x2", "Ir a") { it.poner(Pestana.inicio, emptyList()); it.pestana = Pestana.inicio },
            ItemBuscar("Pedidos", "Ventas", "list.bullet.rectangle", "Ir a", "pedidos") { it.filtroPedidos = null; it.poner(Pestana.pedidos, emptyList()); it.pestana = Pestana.pedidos },
            ItemBuscar("Para preparar", "Pedidos", "shippingbox", "Ir a", "pedidos") { it.filtroPedidos = "Para preparar"; it.pestana = Pestana.pedidos },
            ItemBuscar("Productos", "Catálogo", "cube.box", "Ir a") { it.poner(Pestana.productos, emptyList()); it.pestana = Pestana.productos },
            ItemBuscar("Sin stock", "Productos", "exclamationmark.triangle", "Ir a") { it.segmentoProductos = "out"; it.pestana = Pestana.productos },
            ir("Inventario", "Catálogo", "barcode.viewfinder", Destino.inventario, "catalogo"),
            ir("Traer mi tienda", "Catálogo · Tiendanube, Shopify…", "arrow.down.doc", Destino.importar, "catalogo"),
            ir("Clientes", "Ventas", "person.2", Destino.clientes, "clientes"),
            ir("Caja del día", "Ventas", "banknote", Destino.caja, "pedidos"),
            ir("Carritos abandonados", "Crecer", "cart", Destino.carritos, "clientes"),
            ir("Reseñas y preguntas", "Crecer", "star.bubble", Destino.resenas, "clientes"),
            ir("Devoluciones", "Ventas", "arrow.uturn.backward", Destino.devoluciones, "pedidos"),
            ir("Cupones", "Crecer", "ticket", Destino.cupones, "catalogo"),
            ir("Links de pago", "Ventas", "link", Destino.links, "pedidos"),
            ir("Estadísticas", "Crecer", "chart.bar.xaxis", Destino.estadisticas, "pedidos"),
            ir("Compartí tu tienda", "Tienda online", "square.and.arrow.up", Destino.compartir),
            ir("Editar mi tienda", "Plantilla, colores, logo", "paintbrush.pointed", Destino.miTienda, "cuenta"),
            ir("Finanzas", "Plata", "chart.line.uptrend.xyaxis", Destino.finanzas, "finanzas"),
            ir("Facturación de ARCA", "Plata", "doc.text", Destino.facturacion, "finanzas"),
            ir("Plan", "Cuenta", "creditcard", Destino.plan, "cuenta"),
            ir("Datos del negocio y cobro", "Configuración · alias, envíos", "building.2", Destino.negocio, "cuenta"),
            ir("Integraciones", "Mercado Pago, Mercado Libre, Correo", "puzzlepiece.extension", Destino.integraciones, "cuenta"),
            ir("Tincho en Telegram", "Integraciones", "paperplane", Destino.telegram, "cuenta"),
            ir("Google y Meta", "Integraciones · Analytics, píxel", "chart.xyaxis.line", Destino.medicion, "cuenta"),
            ir("Tu IA", "Integraciones · Claude, ChatGPT", "sparkles", Destino.ia, "cuenta"),
            ir("Usuarios y permisos", "Equipo", "person.badge.key", Destino.usuarios, "cuenta"),
            ir("Notificaciones", "Qué te avisamos", "bell.badge", Destino.preferenciasAvisos),
            ir("Perfil, seguridad y tema", "Tu cuenta", "person.crop.circle", Destino.cuenta),
            ir("Avisos", "Lo último que pasó", "bell", Destino.avisos),
            ir("Ayuda", "Preguntas frecuentes", "questionmark.circle", Destino.ayuda),
        )
        (acciones + secciones).filter { i -> i.permiso?.let { estado.puede(it) } ?: true }
    }
    val filtrados = run {
        val t = normal(q)
        if (t.isEmpty()) items else {
            val palabras = t.split(" ").filter { it.isNotEmpty() }
            items.filter { i -> val texto = normal("${i.titulo} ${i.detalle}"); palabras.all { texto.contains(it) } }
        }
    }

    fun preguntar() {
        val p = q.trim()
        cerrar()
        if (p.isNotEmpty()) nav.preguntaTincho = p
        nav.pestana = Pestana.tincho
    }

    Column(Modifier.fillMaxSize().background(C.elevado).imePadding()) {
        Row(Modifier.fillMaxWidth().padding(start = 18.dp, end = 6.dp).padding(vertical = 6.dp), horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
            Icono("magnifyingglass", 18f, C.texto2)
            Entrada(q, { q = it }, "Buscá una sección o pregunta", Modifier.weight(1f).focusRequester(foco), accionTeclado = ImeAction.Search, alEnviar = { preguntar() }, sinCorrector = true)
            Box(Modifier.size(44.dp).clickable { cerrar() }.semantics { contentDescription = "Cerrar" }, contentAlignment = Alignment.Center) { Icono("xmark", 15f, C.texto2) }
        }
        Box(Modifier.fillMaxWidth().height(0.5.dp).background(C.linea))
        LazyColumn(Modifier.weight(1f), contentPadding = PaddingValues(10.dp), verticalArrangement = Arrangement.spacedBy(2.dp)) {
            item {
                Row(Modifier.padding(bottom = 4.dp).fillMaxWidth().clip(RoundedCornerShape(14.dp))
                    .background(Brush.verticalGradient(listOf(T.primarioArriba.actual.copy(alpha = 0.09f), hex(0xB0441D, 0.09f)))).clickable { preguntar() }.padding(12.dp),
                    horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
                    Box(Modifier.size(36.dp).clip(RoundedCornerShape(11.dp)).background(Brush.linearGradient(listOf(hex(0xE8875F), hex(0xC24E25)))), contentAlignment = Alignment.Center) {
                        Icono("sparkles", 16f, Color.White)
                    }
                    Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(1.dp)) {
                        Texto("Preguntale a Tincho", Estilo.fila, C.texto)
                        Texto(if (q.isEmpty()) "«¿Cuánto vendí ayer por Mercado Libre?»" else "«$q»", Estilo.footnote, C.texto2, lineas = 1)
                    }
                }
            }
            itemsIndexed(filtrados, key = { _, it -> it.id }) { i, it ->
                Column {
                    if (i == 0 || filtrados[i - 1].grupo != it.grupo) {
                        Texto(it.grupo, Estilo.footnoteFuerte, C.texto3, Modifier.padding(horizontal = 12.dp).padding(top = 12.dp, bottom = 4.dp))
                    }
                    Row(Modifier.fillMaxWidth().clickable { Haptico.seleccion(); cerrar(); it.accion(nav) }.padding(horizontal = 12.dp, vertical = 8.dp),
                        horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
                        Box(Modifier.size(34.dp).clip(RoundedCornerShape(10.dp)).background(hex(0xB0441D, 0.1f)), contentAlignment = Alignment.Center) { Icono(it.simbolo, 15f, C.primarioTexto) }
                        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(1.dp)) {
                            Texto(it.titulo, Estilo.fila, C.texto)
                            Texto(it.detalle, Estilo.footnote, C.texto2, lineas = 1)
                        }
                    }
                }
            }
            if (filtrados.isEmpty()) item {
                Texto(Formato.sinViudas("Ninguna sección se llama «$q». Tocá «Preguntale a Tincho» y te respondo."), Estilo.texto14, C.texto2,
                    Modifier.fillMaxWidth().padding(28.dp), alinear = TextAlign.Center)
            }
        }
    }
    LaunchedEffect(Unit) { runCatching { foco.requestFocus() } }
}
