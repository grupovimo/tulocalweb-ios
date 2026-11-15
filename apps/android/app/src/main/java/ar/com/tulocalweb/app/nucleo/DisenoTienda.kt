package ar.com.tulocalweb.app.nucleo

import ar.com.tulocalweb.app.pantallas.mas.EstiloTienda
import kotlinx.serialization.KSerializer
import kotlinx.serialization.Serializable
import kotlinx.serialization.builtins.ListSerializer
import kotlinx.serialization.json.JsonNull
import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.JsonPrimitive
import kotlinx.serialization.json.buildJsonObject
import kotlinx.serialization.json.put
import java.time.Instant

// Clon de Nucleo/DisenoTienda.swift.
// El diseño v2 de la tienda online (docs/especificaciones/tienda-bloques.md): las páginas armadas con bloques. Lee y
// escribe exactamente el JSON de la especificación; lo que la app no edita viaja tal cual.

/** Un bloque de una página: el tipo (bloques.json), su variante y sus datos (lo que falta toma porDefecto). */
@Serializable(with = Bloque.Serializador::class)
data class Bloque(
    val id: String,
    val tipo: String,
    val variante: String,
    val oculto: Boolean = false,
    val datos: Map<String, ValorJSON> = emptyMap(),
) {
    companion object {
        /** Un id nuevo: [a-z0-9]{8}, como pide la especificación (6 a 16). */
        fun nuevoId(): String {
            val letras = "abcdefghijklmnopqrstuvwxyz0123456789"
            return (0 until 8).map { letras.random() }.joinToString("")
        }

        /** Un bloque recién agregado: la primera variante y los datos vacíos (la tienda usa lo de por defecto). */
        fun nuevo(tipo: CatalogoBloques.Tipo): Bloque = Bloque(nuevoId(), tipo.tipo, tipo.variantes.firstOrNull() ?: "")
    }

    /** `oculto` va sólo si es true y `datos` siempre, como en la especificación. */
    object Serializador : KSerializer<Bloque> {
        override val descriptor = JsonObject.serializer().descriptor
        override fun serialize(encoder: kotlinx.serialization.encoding.Encoder, value: Bloque) {
            encoder.encodeSerializableValue(JsonObject.serializer(), buildJsonObject {
                put("id", value.id); put("tipo", value.tipo); put("variante", value.variante)
                if (value.oculto) put("oculto", true)
                put("datos", JsonObject(value.datos))
            })
        }
        override fun deserialize(decoder: kotlinx.serialization.encoding.Decoder): Bloque {
            val o = decoder.decodeSerializableValue(JsonObject.serializer())
            return Bloque(o["id"]?.texto ?: "", o["tipo"]?.texto ?: "", o["variante"]?.texto ?: "", o["oculto"]?.bool ?: false, o["datos"]?.objeto ?: emptyMap())
        }
    }
}

/** Los ajustes de una página que no es el inicio (paginas.catalogo, producto, carrito, nosotros): su variante, los
 *  campos de bloques.json › paginas y, en Ficha y Nosotros, sus bloques. */
data class AjustesPagina(
    val variante: String? = null,
    val bloques: List<Bloque>? = null,
    val campos: Map<String, ValorJSON> = emptyMap(),
) {
    val json: ValorJSON get() {
        val o = campos.toMutableMap()
        variante?.let { o["variante"] = Valor.texto(it) }
        bloques?.let { o["bloques"] = valorDe(ListSerializer(Bloque.serializer()), it) }
        return Valor.objeto(o)
    }

    companion object {
        fun desde(json: ValorJSON?): AjustesPagina {
            val o = json?.objeto ?: return AjustesPagina()
            return AjustesPagina(o["variante"]?.texto, o["bloques"]?.como(ListSerializer(Bloque.serializer())), o.filterKeys { it != "variante" && it != "bloques" })
        }
    }
}

data class Diseno(
    val plantilla: String? = null,
    val paleta: Int = 0,
    val titulo: String = "",
    val lema: String = "",
    val portada: String = "",
    val logo: String = "",
    val compartirTitulo: String = "",
    val compartirDescripcion: String = "",
    val estilo: EstiloTienda = EstiloTienda(),
    val inicio: List<Bloque> = emptyList(),
    val paginas: Map<String, AjustesPagina> = emptyMap(),
    /** Lo que la app no edita (compartirFoto y lo que venga): vuelve igual en el PUT. */
    val resto: Map<String, ValorJSON> = emptyMap(),
) {
    /** Para el PUT: version 2, sin lo vacío y sin nada de v1. */
    val json: Map<String, ValorJSON> get() {
        val o = resto.toMutableMap()
        o["version"] = Valor.numero(2)
        if (!plantilla.isNullOrEmpty()) o["plantilla"] = Valor.texto(plantilla)
        o["paleta"] = Valor.numero(paleta)
        listOf("titulo" to titulo, "lema" to lema, "portada" to portada, "logo" to logo, "compartirTitulo" to compartirTitulo,
            "compartirDescripcion" to compartirDescripcion).forEach { (k, v) -> if (v.isNotBlank()) o[k] = Valor.texto(v) }
        val e = estilo.json
        if (!e.objeto.isNullOrEmpty()) o["estilo"] = e
        o["inicio"] = valorDe(ListSerializer(Bloque.serializer()), inicio)
        if (paginas.isNotEmpty()) o["paginas"] = Valor.objeto(paginas.mapValues { it.value.json })
        return o
    }

    /** El JSON en texto, para la vista previa (evaluateJavascript). */
    val textoJSON: String get() =
        // U+2028 y U+2029 valen en JSON pero cortan un literal de JavaScript.
        Valor.objeto(json).toString().replace("\u2028", "\\u2028").replace("\u2029", "\\u2029")

    // MARK: Bloques por página

    fun bloques(pagina: String): List<Bloque> = if (pagina == "inicio") inicio else (paginas[pagina]?.bloques ?: emptyList())

    fun ponerBloques(b: List<Bloque>, en: String): Diseno {
        if (en == "inicio") return copy(inicio = b)
        val a = (paginas[en] ?: CatalogoBloques.shared.ajustesNuevos(en)).copy(bloques = b)
        return copy(paginas = paginas + (en to a))
    }

    fun ajustes(pagina: String): AjustesPagina = paginas[pagina] ?: CatalogoBloques.shared.ajustesNuevos(pagina)

    val idsUsados: Set<String> get() = CatalogoBloques.ordenPaginas.flatMap { p -> bloques(p).map { it.id } }.toSet()

    /** Busca un bloque en cualquier página (lo que toca la persona en la vista previa). */
    fun buscar(id: String): Pair<String, Bloque>? {
        for (p in CatalogoBloques.ordenPaginas) bloques(p).firstOrNull { it.id == id }?.let { return p to it }
        return null
    }

    // MARK: Punto de partida

    /** Cambia de punto de partida: la dirección, el estilo base y el orden del preset, conservando los datos de los
     *  bloques del mismo tipo (y los que el preset no trae, al final). Se deshace con el historial. */
    fun aplicarPreset(id: String): Diseno {
        val p = PresetTienda.todos[id] ?: return this
        // Lo que define el preset vuelve a ser el del preset (el diseño guarda sólo lo que cambió la tienda).
        val e = estilo.copy(direccion = null, forma = null, densidad = null, ancho = null, fotoProducto = null, encabezado = null, pie = null,
            fuenteTitulos = null, fuenteTexto = null)
        val viejos = inicio.toMutableList()
        val usados = CatalogoBloques.ordenPaginas.filter { it != "inicio" }.flatMap { q -> bloques(q).map { it.id } }.toMutableSet()
        val nuevos = mutableListOf<Bloque>()
        for (b in p.inicio) {
            val i = viejos.indexOfFirst { it.tipo == b.tipo }
            if (i >= 0) {
                nuevos += viejos.removeAt(i).copy(variante = b.variante)
            } else {
                var n = b
                if (n.id in usados || viejos.any { it.id == n.id }) n = n.copy(id = Bloque.nuevoId())
                nuevos += n
            }
            usados += nuevos.last().id
        }
        return copy(plantilla = id, paleta = 0, estilo = e, inicio = nuevos + viejos)
    }

    companion object {
        private val propias = setOf("version", "plantilla", "paleta", "titulo", "lema", "portada", "logo", "compartirTitulo", "compartirDescripcion",
            "estilo", "inicio", "paginas", "secciones")

        /** Desde el JSON del GET. La API lo manda ya en v2; si llega uno viejo, se migra acá igual que en la web. */
        fun desde(o: Map<String, ValorJSON>, plantillaTienda: String): Diseno {
            val d = Diseno(
                resto = o.filterKeys { it !in propias },
                plantilla = o["plantilla"]?.texto,
                paleta = o["paleta"]?.entero ?: 0,
                titulo = o["titulo"]?.texto ?: "",
                lema = o["lema"]?.texto ?: "",
                portada = o["portada"]?.texto ?: "",
                logo = o["logo"]?.texto ?: "",
                compartirTitulo = o["compartirTitulo"]?.texto ?: "",
                compartirDescripcion = o["compartirDescripcion"]?.texto ?: "",
                estilo = EstiloTienda.desde(o["estilo"]),
            )
            return if (o["version"]?.entero == 2) {
                d.copy(inicio = o["inicio"]?.como(ListSerializer(Bloque.serializer())) ?: emptyList(),
                    paginas = o["paginas"]?.objeto?.mapValues { AjustesPagina.desde(it.value) } ?: emptyMap())
            } else d.migrar(o, d.plantilla ?: plantillaTienda)
        }
    }

    // MARK: Migración v1 → v2 (por si la API manda uno viejo; los mismos casos que casos/diseno-v2.json)

    private fun migrar(o: Map<String, ValorJSON>, plantilla: String): Diseno {
        val p = PresetTienda.todos[plantilla] ?: PresetTienda.todos["litoral"]
        val deSeccion = mapOf("hero" to "portada", "cats" to "categorias", "prods" to "productos", "revs" to "resenas", "ig" to "galeria", "wa" to "whatsapp")
        val secciones: List<Pair<String, Boolean>> = o["secciones"]?.lista?.mapNotNull { s ->
            val t = s["id"]?.texto?.let { deSeccion[it] } ?: return@mapNotNull null
            t to (s["visible"]?.bool ?: true)
        } ?: emptyList()
        val textos = mutableMapOf<String, String>()
        o["estilo"]?.get("textos")?.objeto?.forEach { (k, v) -> val t = deSeccion[k]; val s = v.texto; if (t != null && !s.isNullOrEmpty()) textos[t] = s }
        val orden = p?.orden ?: emptyList()
        // El orden de la tienda si lo cambió (filtrado a los del preset); si no, el del preset.
        val deFabrica = listOf("portada", "categorias", "productos", "resenas", "galeria", "whatsapp")
        val tipos = secciones.map { it.first }
        val tocado = tipos != deFabrica.filter { t -> t in tipos }
        val propio = tipos.filter { it in orden }
        val final = if (tocado && propio.size == orden.size) propio else orden
        val inicio = final.map { t ->
            val datos = mutableMapOf<String, ValorJSON>()
            if (t == "portada") {
                if (titulo.isNotEmpty()) datos["titulo"] = Valor.texto(titulo)
                if (lema.isNotEmpty()) datos["texto"] = Valor.texto(lema)
                if (portada.isNotEmpty()) datos["fotos"] = Valor.lista(listOf(Valor.texto(portada)))
            }
            textos[t]?.let { datos["titulo"] = Valor.texto(it) }
            val visible = secciones.firstOrNull { it.first == t }?.second ?: true
            Bloque(t, t, p?.variantes?.get(t) ?: CatalogoBloques.shared.tipo(t)?.variantes?.firstOrNull() ?: "", !visible, datos)
        }
        var e = estilo
        if (secciones.any { it.first == "whatsapp" && !it.second }) e = e.copy(whatsappFlotante = false)
        val a = e.anuncio
        if (!a.isNullOrBlank()) {
            val link = e.anuncioLink?.ifEmpty { null }
            e = e.copy(barra = EstiloTienda.Barra("fija", listOf(EstiloTienda.Barra.Mensaje(a, link))))
        }
        return copy(inicio = inicio, estilo = e.copy(anuncio = null, anuncioLink = null, textos = null))
    }
}

/** Deshacer y rehacer del editor: fotos del diseño entero (es chico, hasta 64 KB). */
class Historial {
    var pasado: List<Diseno> = emptyList()
        private set
    var futuro: List<Diseno> = emptyList()
        private set

    val puedeDeshacer: Boolean get() = pasado.isNotEmpty()
    val puedeRehacer: Boolean get() = futuro.isNotEmpty()

    /** Antes de cambiar algo: guarda cómo estaba. */
    fun guardar(d: Diseno) {
        if (pasado.lastOrNull() != d) pasado = pasado + d
        if (pasado.size > 60) pasado = pasado.drop(1)
        futuro = emptyList()
    }

    /** Si al final no cambió nada (se abrió un panel y se cerró), no deja un paso vacío. */
    fun descartarSiIgual(d: Diseno) { if (pasado.lastOrNull() == d) pasado = pasado.dropLast(1) }

    fun deshacer(actual: Diseno): Diseno? {
        val d = pasado.lastOrNull() ?: return null
        pasado = pasado.dropLast(1); futuro = futuro + actual
        return d
    }

    fun rehacer(actual: Diseno): Diseno? {
        val d = futuro.lastOrNull() ?: return null
        futuro = futuro.dropLast(1); pasado = pasado + actual
        return d
    }
}

// MARK: - Qué le falta a un bloque

/** Lo que el editor sabe de la tienda para avisar por qué un bloque no se va a ver. */
data class ContextoTienda(
    val nombre: String = "",
    val whatsapp: Boolean = false,
    val direccionLocal: Boolean = false,
    val categorias: List<String> = emptyList(),
    val productos: Int = 0,
    val resenas: Boolean = false,
    val fotos: Boolean = false,
)

/** Por qué un bloque no se va a dibujar (regla «Nada vacío»), en una frase corta. */
data class FaltaBloque(
    val texto: String,
    /** Se arregla en Más › Datos del negocio. */
    val enNegocio: Boolean = false,
) {
    companion object {
        fun de(b: Bloque, ctx: ContextoTienda): FaltaBloque? {
            val t = CatalogoBloques.shared.tipo(b.tipo) ?: return null
            deCampos(t.campos, b.datos)?.let { return it }
            val sinProductos = FaltaBloque("Cargá tu primer producto para que se vea.")
            when (b.tipo) {
                "whatsapp" -> if (!ctx.whatsapp) return FaltaBloque("Cargá tu WhatsApp en Datos del negocio para que se vea.", enNegocio = true)
                "local" -> if (!ctx.direccionLocal) return FaltaBloque("Cargá la dirección del local en Datos del negocio.", enNegocio = true)
                "resenas" -> if (!ctx.resenas) return FaltaBloque("Se va a ver cuando llegue tu primera reseña.")
                "categorias", "menu" -> {
                    if (ctx.productos == 0) return sinProductos
                    if (ctx.categorias.isEmpty()) return FaltaBloque("Poné una categoría a tus productos para que se vea.")
                }
                "galeria" -> if ((b.datos["fotos"]?.lista ?: emptyList()).isEmpty() && !ctx.fotos) return FaltaBloque("Subí fotos o cargá productos con foto.")
                "productos", "oferta", "destacado", "combo" -> {
                    if (ctx.productos == 0) return sinProductos
                    b.datos["productos"]?.objeto?.let { p ->
                        val fuente = p["fuente"]?.texto
                        if (fuente == "categoria" && (p["categoria"]?.texto ?: "").isEmpty()) return FaltaBloque("Elegí de qué categoría son los productos.")
                        if (fuente == "elegidos" && (p["ids"]?.lista ?: emptyList()).isEmpty()) return FaltaBloque("Elegí los productos que van.")
                    }
                    if (b.tipo == "oferta") {
                        val f = b.datos["hasta"]?.texto?.let { Fecha.iso(it) }
                        if (f != null && f.isBefore(Instant.now())) return FaltaBloque("La oferta ya terminó: poné otra fecha.")
                    }
                }
                "banner" -> if (b.variante == "doble" && (b.datos["segundo"]?.get("titulo")?.texto ?: "").isEmpty()) {
                    return FaltaBloque("Escribí el título de la segunda promo (la variante Doble lleva dos).")
                }
            }
            return null
        }

        private fun deCampos(campos: List<CatalogoBloques.Campo>, datos: Map<String, ValorJSON>): FaltaBloque? {
            for (c in campos) {
                val v = datos[c.id] ?: c.porDefecto
                if (c.requerido == true && (v?.vacio ?: true)) {
                    return FaltaBloque(if (c.tipo == "producto" || c.tipo == "elegidos") "Elegí «${c.nombre.lowercase()}»." else "Completá «${c.nombre}».")
                }
                val min = c.min
                if (min != null && (c.tipo == "lista" || c.tipo == "elegidos") && (v?.lista?.size ?: 0) < min) {
                    return FaltaBloque(if (min == 1) "Sumá al menos uno en «${c.nombre}»." else "Sumá al menos $min en «${c.nombre}».")
                }
                val item = c.item
                if (c.tipo == "lista" && item != null) {
                    for (e in v?.lista ?: emptyList()) {
                        val o = e.objeto ?: continue
                        val f = deCampos(item, o) ?: continue
                        val resto = f.texto.take(1).lowercase() + f.texto.drop(1)
                        return FaltaBloque("En «${c.nombre}»: $resto")
                    }
                }
            }
            return null
        }
    }
}

// MARK: - ValorJSON

/** Vacío para la tienda: sin texto, lista vacía u objeto sin nada adentro. */
val ValorJSON.vacio: Boolean get() = when {
    this is JsonNull -> true
    this is JsonPrimitive && isString -> content.isBlank()
    this is kotlinx.serialization.json.JsonArray -> isEmpty()
    this is JsonObject -> values.all { it.vacio }
    else -> false
}

fun <T> valorDe(s: KSerializer<T>, v: T): ValorJSON = runCatching { Api.json.encodeToJsonElement(s, v) }.getOrDefault(JsonNull)

fun <T> ValorJSON.como(s: KSerializer<T>): T? = runCatching { Api.json.decodeFromJsonElement(s, this) }.getOrNull()

