package ar.com.tulocalweb.app.pantallas.ingreso

import androidx.compose.animation.AnimatedVisibility
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
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
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
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.heading
import androidx.compose.ui.semantics.selected
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.nucleo.ErrorApi
import ar.com.tulocalweb.app.nucleo.Formato
import ar.com.tulocalweb.app.nucleo.ValorJSON
import ar.com.tulocalweb.app.nucleo.Valor
import ar.com.tulocalweb.app.nucleo.bool
import ar.com.tulocalweb.app.nucleo.compartirTexto
import ar.com.tulocalweb.app.nucleo.entero
import ar.com.tulocalweb.app.nucleo.get
import ar.com.tulocalweb.app.nucleo.objeto
import ar.com.tulocalweb.app.nucleo.opcional
import ar.com.tulocalweb.app.nucleo.texto
import ar.com.tulocalweb.app.pantallas.compartir.Compartir
import ar.com.tulocalweb.app.pantallas.mas.CampoTexto
import ar.com.tulocalweb.app.pantallas.mas.ChipSimple
import ar.com.tulocalweb.app.pantallas.mas.ConectarCuenta
import ar.com.tulocalweb.app.pantallas.mas.PlantillaTienda
import ar.com.tulocalweb.app.pantallas.mas.QRTienda
import ar.com.tulocalweb.app.pantallas.mas.TraerTienda
import ar.com.tulocalweb.app.tema.*
import coil3.compose.AsyncImage
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import kotlinx.serialization.Serializable

// Clon de Pantallas/Ingreso/Onboarding.swift.

private enum class PasoOnb(val titulo: String, val bajada: String) {
    negocio("¿Qué vendés?", "Así armamos tu tienda con lo que necesita tu rubro."),
    importar("¿Ya vendés en otra plataforma?", "Traemos tus productos, clientes y pedidos sin cargar nada a mano."),
    tienda("Elegí cómo se ve tu tienda", "Una plantilla para empezar y la dirección de tu tienda. Las cambiás cuando quieras."),
    cobros("¿Cómo te pagan?", "Conectá Mercado Pago para cobrar con tarjeta y QR, o dejá tus datos para transferencias."),
    listo("¡Tu tienda está online!", "Compartí el link y empezá a vender."),
}

@Serializable private data class Rubro(val id: String, val nombre: String, val plantilla: String = "")

/** El onboarding de una tienda recién creada, entero en el celular (como /onboarding en la web): tu negocio (rubro),
 *  traer tu tienda de otra plataforma, plantilla y dirección, cobros y ¡listo! Cada paso se guarda en la tienda
 *  (PUT /v1/panel/onboarding), así se retoma donde quedó, también desde la web. «Lo hago después» deja seguir al panel
 *  en cualquier paso. */
@Composable
fun Onboarding() {
    val estado = LocalEstado.current
    val ctx = LocalContext.current
    val alcance = rememberCoroutineScope()
    var paso by remember { mutableStateOf(PasoOnb.negocio) }
    var cargado by remember { mutableStateOf(false) }
    var guardando by remember { mutableStateOf(false) }
    var rubros by remember { mutableStateOf<List<Rubro>>(emptyList()) }
    var rubro by remember { mutableStateOf("") }
    var plantilla by remember { mutableStateOf("") }
    var traer by remember { mutableStateOf<Boolean?>(null) }
    var slug by remember { mutableStateOf("") }
    var disponible by remember { mutableStateOf<String?>(null) }
    var slugOk by remember { mutableStateOf(true) }
    var mpConectado by remember { mutableStateOf(false) }
    var conectando by remember { mutableStateOf(false) }
    var alias by remember { mutableStateOf("") }
    var cbu by remember { mutableStateOf("") }
    var titular by remember { mutableStateOf("") }
    var negocio by remember { mutableStateOf<Map<String, ValorJSON>>(emptyMap()) }
    var qr by remember { mutableStateOf(false) }
    val plantillas = remember { PlantillaTienda.todas() }
    val total = PasoOnb.entries.size

    val puedeSeguir = when (paso) {
        PasoOnb.negocio -> rubro.isNotEmpty()
        PasoOnb.importar -> traer != null
        PasoOnb.tienda -> plantilla.isNotEmpty() && slugOk && slug.length >= 2
        else -> true
    }

    @Serializable data class MP(val conectada: Boolean? = null)

    suspend fun cargar() {
        try {
            slug = estado.tienda?.slug ?: ""
            estado.api.opcional<List<Rubro>>("/v1/rubros")?.let { rubros = it }
            estado.api.opcional<ValorJSON>("/v1/panel/onboarding")?.let { v ->
                plantilla = v["plantilla"]?.texto ?: ""
                v["elecciones"]?.objeto?.let { e ->
                    rubro = e["rubro"]?.texto ?: ""
                    e["pasoApp"]?.entero?.let { p -> PasoOnb.entries.getOrNull(p)?.let { paso = it } }
                    e["traer"]?.bool?.let { traer = it }
                }
            }
            if (plantilla.isEmpty()) plantilla = plantillas.firstOrNull()?.id ?: ""
            estado.api.opcional<ValorJSON>("/v1/panel/configuracion")?.get("negocio")?.objeto?.let { n ->
                negocio = n
                alias = n["alias"]?.texto ?: ""
                cbu = n["cbu"]?.texto ?: ""
                titular = n["titular"]?.texto ?: ""
            }
            mpConectado = estado.api.opcional<MP>("/v1/panel/integraciones/mercadopago")?.conectada ?: false
        } finally { cargado = true }
    }

    @Serializable data class D(val disponible: Boolean, val mensaje: String)

    suspend fun revisarSlug() {
        val s = slug
        if (s.length < 2) { slugOk = false; disponible = "Al menos 2 letras o números."; return }
        if (s == estado.tienda?.slug) { slugOk = true; disponible = null; return }
        delay(350)
        if (s != slug) return
        val d = estado.api.opcional<D>("/v1/panel/tienda/direccion", listOf("slug" to s))
        if (d != null && s == slug) { slugOk = d.disponible; disponible = d.mensaje }
    }

    /** Guarda lo elegido (y en qué paso va) en la tienda. */
    suspend fun guardar(terminado: Boolean = false) {
        val e = mutableMapOf("rubro" to Valor.texto(rubro), "pasoApp" to Valor.numero(paso.ordinal))
        traer?.let { e["traer"] = Valor.booleano(it) }
        val o = mutableMapOf("elecciones" to Valor.objeto(e), "terminado" to Valor.booleano(terminado))
        if (plantilla.isNotEmpty()) o["plantilla"] = Valor.texto(plantilla)
        estado.api.hacer("PUT", "/v1/panel/onboarding", Valor.objeto(o))
    }

    @Serializable data class C2(val slug: String)

    suspend fun seguir() {
        guardando = true
        try {
            when (paso) {
                PasoOnb.tienda -> if (slug != estado.tienda?.slug) {
                    estado.api.hacer("PUT", "/v1/panel/tienda/direccion", C2(slug))
                    estado.refrescarAcceso()
                }
                PasoOnb.cobros -> {
                    val datos = listOf(alias, cbu, titular).map { it.trim() }
                    if (datos != listOf(negocio["alias"]?.texto ?: "", negocio["cbu"]?.texto ?: "", negocio["titular"]?.texto ?: "")) {
                        val n = negocio.toMutableMap()
                        n["alias"] = Valor.texto(datos[0]); n["cbu"] = Valor.texto(datos[1]); n["titular"] = Valor.texto(datos[2])
                        estado.api.hacer("PUT", "/v1/panel/configuracion/negocio", Valor.objeto(n))
                        negocio = n
                    }
                }
                PasoOnb.listo -> {
                    guardar(terminado = true)
                    Haptico.exito()
                    estado.terminarOnboarding()
                    return
                }
                else -> {}
            }
            PasoOnb.entries.getOrNull(paso.ordinal + 1)?.let { p ->
                paso = p
                guardar(terminado = p == PasoOnb.listo)
                if (p == PasoOnb.listo) Haptico.exito()
            }
        } catch (e: ErrorApi) {
            Haptico.error()
            estado.avisar(e.detalle, ok = false)
        } catch (e: kotlinx.coroutines.CancellationException) { throw e } catch (e: Exception) {
            Haptico.error()
            estado.avisar("No pudimos guardar. Revisá tu conexión y probá de nuevo.", ok = false)
        } finally { guardando = false }
    }

    suspend fun despues() {
        runCatching { guardar() }
        estado.terminarOnboarding()
    }

    @Composable
    fun Opcion(t: String, d: String, simbolo: String, elegida: Boolean, accion: () -> Unit) {
        val forma = RoundedCornerShape(16.dp)
        Row(Modifier.fillMaxWidth().tarjeta(radio = 16.dp).border(1.5.dp, if (elegida) C.acento else Color.Transparent, forma)
            .clickable { Haptico.seleccion(); accion() }.semantics { selected = elegida }.padding(14.dp),
            horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
            Box(Modifier.size(40.dp).clip(RoundedCornerShape(12.dp)).background(C.crema), contentAlignment = Alignment.Center) { Icono(simbolo, 16f, C.primarioTexto) }
            Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                Texto(t, Estilo.fila, C.texto)
                Texto(d, Estilo.footnote, C.texto2)
            }
        }
    }

    ConBarraAccion({
        BotonPrincipal(if (paso == PasoOnb.listo) "Ir a mi panel" else "Seguir", { alcance.launch { seguir() } }, habilitado = !guardando && puedeSeguir, cargando = guardando)
        if (paso != PasoOnb.listo) BotonTexto("Lo hago después", { alcance.launch { despues() } }, color = C.texto2)
    }, Modifier.background(C.fondo).statusBarsPadding().imePadding()) { abajo ->
        Column(Modifier.fillMaxSize()) {
            // Cabecera (volver y avance)
            Row(Modifier.fillMaxWidth().padding(horizontal = 16.dp).padding(top = 8.dp, bottom = 8.dp), horizontalArrangement = Arrangement.spacedBy(14.dp), verticalAlignment = Alignment.CenterVertically) {
                val sinVolver = paso == PasoOnb.negocio || paso == PasoOnb.listo
                Box(Modifier.size(40.dp).alpha(if (sinVolver) 0f else 1f).clickable(enabled = !sinVolver) { PasoOnb.entries.getOrNull(paso.ordinal - 1)?.let { paso = it } }
                    .semantics { contentDescription = "Volver al paso anterior" }, contentAlignment = Alignment.Center) {
                    Icono("chevron.left", 17f, C.texto)
                }
                BoxWithConstraints(Modifier.weight(1f).height(6.dp).semantics { contentDescription = "Paso ${paso.ordinal + 1} de $total" }) {
                    Box(Modifier.fillMaxSize().clip(Capsula).background(C.pista))
                    Box(Modifier.width(maxWidth * (paso.ordinal + 1) / total).fillMaxHeight().clip(Capsula).background(C.acento))
                }
                Texto("${paso.ordinal + 1}/$total", Estilo.footnoteFuerte, C.texto2, numeros = true)
            }
            Column(Modifier.weight(1f).verticalScroll(rememberScrollState()).padding(horizontal = 24.dp).padding(top = 8.dp, bottom = 24.dp + abajo),
                verticalArrangement = Arrangement.spacedBy(16.dp)) {
                Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
                    Texto(paso.titulo, Estilo.saludo, C.texto, Modifier.semantics { heading() })
                    Texto(Formato.sinViudas(paso.bajada), Estilo.body16, C.texto2)
                }
                if (!cargado) {
                    EsqueletoFilas(4, 0.dp)
                } else when (paso) {
                    PasoOnb.negocio -> Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
                        rubros.chunked(2).forEach { par ->
                            Row(horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                                par.forEach { r ->
                                    val forma = RoundedCornerShape(14.dp)
                                    Box(Modifier.weight(1f).heightIn(min = 54.dp).tarjeta(radio = 14.dp).border(1.5.dp, if (rubro == r.id) C.acento else Color.Transparent, forma)
                                        .clickable {
                                            Haptico.seleccion()
                                            rubro = r.id
                                            if (plantillas.any { it.id == r.plantilla }) plantilla = r.plantilla
                                        }.semantics { selected = rubro == r.id }.padding(horizontal = 14.dp), contentAlignment = Alignment.CenterStart) {
                                        Texto(r.nombre, Estilo.fila, C.texto)
                                    }
                                }
                                if (par.size == 1) Spacer(Modifier.weight(1f))
                            }
                        }
                    }
                    PasoOnb.importar -> Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
                        Opcion("Sí, quiero traer mi tienda", "Tiendanube, Shopify, WooCommerce o Empretienda", "arrow.down.doc", traer == true) { traer = true }
                        Opcion("No, empiezo de cero", "Cargás tus productos con una foto, desde el celular", "sparkles", traer == false) { traer = false }
                        AnimatedVisibility(traer == true) { Box(Modifier.padding(top = 4.dp)) { TraerTienda() } }
                    }
                    PasoOnb.tienda -> Column(verticalArrangement = Arrangement.spacedBy(14.dp)) {
                        Row(Modifier.padding(horizontal = 0.dp).horizontalScroll(rememberScrollState()), horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                            plantillas.forEach { p ->
                                Column(Modifier.width(128.dp).clickable { Haptico.seleccion(); plantilla = p.id }.semantics { selected = plantilla == p.id },
                                    verticalArrangement = Arrangement.spacedBy(6.dp)) {
                                    val forma = RoundedCornerShape(14.dp)
                                    Box(Modifier.size(128.dp, 164.dp).clip(forma).background(hex(p.colorFondo))
                                        .border(3.dp, if (plantilla == p.id) C.acento else Color.Transparent, forma)) {
                                        AsyncImage(p.img, null, Modifier.fillMaxSize().clip(forma), contentScale = ContentScale.Crop)
                                    }
                                    Texto(p.nombre, Estilo.fila, C.texto)
                                    Texto(p.rubro, Estilo.footnote, C.texto2, lineas = 1)
                                }
                            }
                        }
                        Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
                            Texto("La dirección de tu tienda", Estilo.footnoteFuerte, C.texto2)
                            val forma = RoundedCornerShape(14.dp)
                            Row(Modifier.fillMaxWidth().clip(forma).background(C.superficie).border(1.dp, if (slugOk) C.linea2 else C.errorTinta, forma)
                                .padding(horizontal = 14.dp, vertical = 12.dp), verticalAlignment = Alignment.CenterVertically) {
                                Entrada(slug, { s ->
                                    slug = s.lowercase().filter { (it.isLetter() && it.code < 128) || it.isDigit() || it == '-' }.take(60)
                                    alcance.launch { revisarSlug() }
                                }, "mitienda", Modifier.weight(1f), estilo = Estilo.body16, teclado = KeyboardType.Uri, sinCorrector = true)
                                Texto(".tulocalweb.com.ar", Estilo.body16, C.texto2)
                            }
                            disponible?.let { Texto(it, Estilo.footnote, if (slugOk) C.okTinta else C.errorTinta) }
                        }
                    }
                    PasoOnb.cobros -> Column(verticalArrangement = Arrangement.spacedBy(14.dp)) {
                        Row(Modifier.fillMaxWidth().tarjeta(radio = 16.dp).padding(14.dp), horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
                            Box(Modifier.size(40.dp).clip(RoundedCornerShape(12.dp)).background(hex(0x009EE3, 0.12f)), contentAlignment = Alignment.Center) {
                                Icono("creditcard.fill", 16f, hex(0x009EE3))
                            }
                            Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
                                Texto("Mercado Pago", Estilo.fila, C.texto)
                                Texto(if (mpConectado) "Conectado: ya cobrás con tarjeta y QR" else "Tarjeta, cuotas, QR y link de pago", Estilo.footnote, C.texto2)
                            }
                            when {
                                mpConectado -> ChipSimple("Conectada", C.okFondo, C.okTinta)
                                conectando -> Rueda(C.acento)
                                else -> BotonTonal("Conectar", {
                                    alcance.launch {
                                        conectando = true
                                        mpConectado = ConectarCuenta.conectar("mercadopago", estado.api)
                                        conectando = false
                                        if (mpConectado) Haptico.exito()
                                    }
                                })
                            }
                        }
                        Texto("Transferencias", Estilo.headline, C.texto, Modifier.padding(top = 4.dp))
                        CampoTexto("Alias", alias, { alias = it })
                        CampoTexto("CBU o CVU", cbu, { cbu = it }, teclado = KeyboardType.Number)
                        CampoTexto("Titular de la cuenta", titular, { titular = it })
                        Texto(Formato.sinViudas("Aparecen en la tienda cuando alguien elige pagar con transferencia."), Estilo.footnote, C.texto2)
                    }
                    PasoOnb.listo -> Column(verticalArrangement = Arrangement.spacedBy(14.dp)) {
                        estado.tienda?.let { t ->
                            Row(Modifier.fillMaxWidth().tarjeta(radio = 16.dp).padding(14.dp), horizontalArrangement = Arrangement.spacedBy(12.dp), verticalAlignment = Alignment.CenterVertically) {
                                AvatarTienda(t, 48.dp, 14.dp)
                                Column(verticalArrangement = Arrangement.spacedBy(2.dp)) {
                                    Texto(t.nombre, Estilo.fila, C.texto)
                                    Texto(t.direccion, Estilo.footnote, C.primarioTexto, lineas = 1, escalaMinima = 0.8f)
                                }
                            }
                            Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                                AccionRapida("doc.on.doc", "Copiar", Modifier.weight(1f)) { Compartir.copiar(ctx, t.url, estado) }
                                AccionRapida("square.and.arrow.up", "Compartir", Modifier.weight(1f)) { compartirTexto(ctx, Compartir.mensajeTienda(t) + "\n" + t.url) }
                                AccionRapida("qrcode", "QR", Modifier.weight(1f)) { qr = true }
                            }
                        }
                        Texto(Formato.sinViudas("En Inicio te esperan los primeros pasos: cargar tus productos con una foto y compartir el link."), Estilo.footnote, C.texto2)
                    }
                }
            }
        }
    }
    Hoja(qr, { qr = false }) { estado.tienda?.let { QRTienda(it) } }
    LaunchedEffect(Unit) { cargar() }
}
