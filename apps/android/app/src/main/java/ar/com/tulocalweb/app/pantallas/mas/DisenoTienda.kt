package ar.com.tulocalweb.app.pantallas.mas

import ar.com.tulocalweb.app.nucleo.Api
import ar.com.tulocalweb.app.nucleo.Recursos
import ar.com.tulocalweb.app.nucleo.Valor
import ar.com.tulocalweb.app.nucleo.ValorJSON
import ar.com.tulocalweb.app.nucleo.bool
import ar.com.tulocalweb.app.nucleo.entero
import ar.com.tulocalweb.app.nucleo.lista
import ar.com.tulocalweb.app.nucleo.objeto
import ar.com.tulocalweb.app.nucleo.texto
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.JsonElement
import kotlinx.serialization.json.JsonObject

// Clon de Pantallas/Mas/DisenoTienda.swift.

/** El diseño v2 de la tienda (docs/especificaciones/tienda-bloques.md): la marca, el estilo, los bloques de Inicio y
 *  las páginas. Lee y escribe exactamente el JSON de la especificación; lo que la app no conoce viaja en `extra`. */
data class DisenoTienda(
    val plantilla: String = "",
    val paleta: Int = 0,
    val titulo: String = "",
    val lema: String = "",
    val portada: String = "",
    val logo: String = "",
    val compartirTitulo: String = "",
    val compartirDescripcion: String = "",
    val compartirFoto: String = "",
    val estilo: EstiloTienda = EstiloTienda(),
    val inicio: List<Bloque> = emptyList(),
    val paginas: PaginasTienda = PaginasTienda(),
    val extra: Map<String, ValorJSON> = emptyMap(),
) {
    /** Los bloques de una página (inicio, producto o nosotros). */
    fun bloques(pagina: String): List<Bloque> = when (pagina) {
        "inicio" -> inicio
        "producto" -> paginas.producto?.bloques ?: emptyList()
        "nosotros" -> paginas.nosotros?.bloques ?: emptyList()
        else -> emptyList()
    }

    fun conBloques(pagina: String, b: List<Bloque>): DisenoTienda = when (pagina) {
        "inicio" -> copy(inicio = b)
        "producto" -> copy(paginas = paginas.copy(producto = (paginas.producto ?: PaginasTienda.Producto()).copy(bloques = b)))
        "nosotros" -> copy(paginas = paginas.copy(nosotros = PaginasTienda.Nosotros(b)))
        else -> this
    }

    /** Todos los bloques del diseño (para que los id no se repitan). */
    val todos: List<Bloque> get() = inicio + (paginas.producto?.bloques ?: emptyList()) + (paginas.nosotros?.bloques ?: emptyList())

    /** Dónde está un bloque: la página y la posición. */
    fun ubicar(id: String): Pair<String, Int>? {
        for (p in listOf("inicio", "producto", "nosotros")) {
            val i = bloques(p).indexOfFirst { it.id == id }
            if (i >= 0) return p to i
        }
        return null
    }

    /** El JSON que va en el PUT /v1/panel/mi-tienda (y en la vista previa). */
    val json: ValorJSON get() {
        val d = extra.toMutableMap()
        d.remove("secciones")
        d["version"] = Valor.numero(2)
        d["plantilla"] = Valor.texto(plantilla)
        d["paleta"] = Valor.numero(paleta)
        d["titulo"] = Valor.texto(titulo)
        d["lema"] = Valor.texto(lema)
        d["portada"] = Valor.texto(portada)
        d["logo"] = Valor.texto(logo)
        d["compartirTitulo"] = Valor.texto(compartirTitulo)
        d["compartirDescripcion"] = Valor.texto(compartirDescripcion)
        if (compartirFoto.isNotEmpty()) d["compartirFoto"] = Valor.texto(compartirFoto) else d.remove("compartirFoto")
        d["estilo"] = estilo.json
        d["inicio"] = Valor.lista(inicio.map { it.json })
        val p = paginas.json
        if (p.isEmpty()) d.remove("paginas") else d["paginas"] = Valor.objeto(p)
        return Valor.objeto(d)
    }

    companion object {
        private val conocidos = setOf("version", "plantilla", "paleta", "titulo", "lema", "portada", "logo", "compartirTitulo",
            "compartirDescripcion", "compartirFoto", "estilo", "inicio", "paginas", "secciones")

        /** Lee el diseño del borrador. Un diseño v1 (o vacío) se migra al vuelo, igual que la web y Go. */
        fun desde(v: ValorJSON?, plantilla: String): DisenoTienda {
            val o = Migrar.aV2(v?.objeto, plantilla)
            return DisenoTienda(
                plantilla = o["plantilla"]?.texto ?: plantilla,
                paleta = o["paleta"]?.entero ?: 0,
                titulo = o["titulo"]?.texto ?: "",
                lema = o["lema"]?.texto ?: "",
                portada = o["portada"]?.texto ?: "",
                logo = o["logo"]?.texto ?: "",
                compartirTitulo = o["compartirTitulo"]?.texto ?: "",
                compartirDescripcion = o["compartirDescripcion"]?.texto ?: "",
                compartirFoto = o["compartirFoto"]?.texto ?: "",
                estilo = EstiloTienda.desde(o["estilo"]),
                inicio = Bloque.lista(o["inicio"]),
                paginas = PaginasTienda.desde(o["paginas"]),
                extra = o.filterKeys { it !in conocidos },
            )
        }
    }
}

/** Un bloque: `id` único en el diseño, `tipo` y `variante` de bloques.json y sus `datos` (lo que falta toma
 *  porDefecto o lo sugerido). */
@Serializable
data class Bloque(val id: String, val tipo: String, val variante: String, val oculto: Boolean? = null, val datos: JsonObject = JsonObject(emptyMap())) {
    val visible: Boolean get() = oculto != true

    operator fun get(campo: String): ValorJSON? = datos[campo]

    fun con(campo: String, valor: ValorJSON?): Bloque {
        val d = datos.toMutableMap()
        if (valor == null || valor.vacio) d.remove(campo) else d[campo] = valor
        return copy(datos = JsonObject(d))
    }

    val json: ValorJSON get() = runCatching { Api.json.encodeToJsonElement(serializer(), copy(oculto = if (oculto == true) true else null)) }.getOrDefault(Valor.nulo)

    companion object {
        fun lista(v: ValorJSON?): List<Bloque> = v?.lista?.mapNotNull { runCatching { Api.json.decodeFromJsonElement(serializer(), it) }.getOrNull() } ?: emptyList()

        /** Un id nuevo: [a-z0-9]{8}, distinto de los que ya están. */
        fun nuevoId(usados: Set<String>): String {
            val letras = "abcdefghijklmnopqrstuvwxyz0123456789"
            while (true) {
                val id = (1..8).map { letras.random() }.joinToString("")
                if (id !in usados) return id
            }
        }
    }
}

/** Lo vacío no se guarda: un texto en blanco, una lista u objeto sin nada. */
val ValorJSON.vacio: Boolean get() = when {
    this is kotlinx.serialization.json.JsonNull -> true
    texto != null -> texto!!.isBlank()
    lista != null -> lista!!.isEmpty()
    objeto != null -> objeto!!.values.all { it.vacio }
    else -> false
}

/** Las páginas además de Inicio. */
data class PaginasTienda(val catalogo: Catalogo? = null, val producto: Producto? = null, val carrito: Carrito? = null, val nosotros: Nosotros? = null) {
    @Serializable data class Catalogo(val variante: String = "grilla", val filtros: String = "costado", val columnas: Int = 4)
    @Serializable data class Producto(val variante: String = "grilla", val relacionados: Boolean = true, val calcularEnvio: Boolean = true, val bloques: List<Bloque> = emptyList())
    @Serializable data class Carrito(val variante: String = "pagina", val envioGratisDesde: Boolean = true)
    @Serializable data class Nosotros(val bloques: List<Bloque> = emptyList())

    val json: Map<String, ValorJSON> get() = buildMap {
        catalogo?.let { put("catalogo", Api.json.encodeToJsonElement(Catalogo.serializer(), it)) }
        producto?.let { put("producto", Valor.objeto(mapOf("variante" to Valor.texto(it.variante), "relacionados" to Valor.booleano(it.relacionados),
            "calcularEnvio" to Valor.booleano(it.calcularEnvio), "bloques" to Valor.lista(it.bloques.map { b -> b.json })))) }
        carrito?.let { put("carrito", Api.json.encodeToJsonElement(Carrito.serializer(), it)) }
        nosotros?.let { put("nosotros", Valor.objeto(mapOf("bloques" to Valor.lista(it.bloques.map { b -> b.json })))) }
    }

    /** Un ajuste de la página (variante o campo de bloques.json › paginas). */
    fun valor(pagina: String, campo: String): ValorJSON? = json[pagina]?.objeto?.get(campo)

    fun con(pagina: String, campo: String, v: ValorJSON): PaginasTienda {
        fun t() = v.texto ?: ""
        fun b() = v.bool ?: true
        return when (pagina) {
            "catalogo" -> (catalogo ?: Catalogo()).let { c ->
                copy(catalogo = when (campo) { "variante" -> c.copy(variante = t()); "filtros" -> c.copy(filtros = t()); "columnas" -> c.copy(columnas = v.entero ?: 4); else -> c })
            }
            "producto" -> (producto ?: Producto()).let { p ->
                copy(producto = when (campo) { "variante" -> p.copy(variante = t()); "relacionados" -> p.copy(relacionados = b()); "calcularEnvio" -> p.copy(calcularEnvio = b()); else -> p })
            }
            "carrito" -> (carrito ?: Carrito()).let { c ->
                copy(carrito = when (campo) { "variante" -> c.copy(variante = t()); "envioGratisDesde" -> c.copy(envioGratisDesde = b()); else -> c })
            }
            else -> this
        }
    }

    companion object {
        fun desde(v: ValorJSON?): PaginasTienda {
            val o = v?.objeto ?: return PaginasTienda()
            fun <T> leer(k: String, s: kotlinx.serialization.KSerializer<T>): T? = o[k]?.let { runCatching { Api.json.decodeFromJsonElement(s, it) }.getOrNull() }
            return PaginasTienda(leer("catalogo", Catalogo.serializer()), leer("producto", Producto.serializer()), leer("carrito", Carrito.serializer()), leer("nosotros", Nosotros.serializer()))
        }
    }
}

// MARK: - Catálogo (assets/bloques.json, el mismo de packages/contrato)

@Serializable
data class CatalogoBloques(
    val version: Int = 1,
    val limites: Limites = Limites(),
    val estilo: Map<String, Opciones> = emptyMap(),
    val fondos: Opciones = Opciones(),
    val fuentesProductos: Opciones = Opciones(),
    val iconos: List<String> = emptyList(),
    val paginas: Map<String, Pagina> = emptyMap(),
    val grupos: List<Grupo> = emptyList(),
    val bloques: List<TipoBloque> = emptyList(),
    val tiposCampo: Map<String, String> = emptyMap(),
    val barraAnuncios: TipoBloque = TipoBloque(),
) {
    @Serializable data class Limites(val bloquesPorPagina: Int = 30, val itemsPorLista: Int = 12, val bytesDiseno: Int = 65536)
    @Serializable data class Opciones(val opciones: List<String> = emptyList(), val nombres: List<String> = emptyList(), val porDefecto: String? = null) {
        val pares: List<Pair<String, String>> get() = opciones.mapIndexed { i, o -> o to (nombres.getOrNull(i) ?: o) }
        fun nombre(o: String): String = nombres.getOrNull(opciones.indexOf(o)) ?: o
    }
    @Serializable data class Pagina(val nombre: String = "", val bloques: Boolean = false, val variantes: List<String> = emptyList(), val nombresVariantes: List<String> = emptyList(),
        val campos: List<CampoBloque> = emptyList())
    @Serializable data class Grupo(val id: String, val nombre: String)

    fun tipo(t: String): TipoBloque? = bloques.firstOrNull { it.tipo == t }

    companion object {
        val compartido: CatalogoBloques by lazy { Recursos.leer<CatalogoBloques>("bloques.json") ?: CatalogoBloques() }

        /** Las páginas que se editan, en el orden del editor. */
        val paginasEditor = listOf("inicio" to "Inicio", "catalogo" to "Catálogo", "producto" to "Ficha", "carrito" to "Carrito", "nosotros" to "Nosotros")
    }
}

/** Un tipo de bloque del catálogo. */
@Serializable
data class TipoBloque(
    val tipo: String = "",
    val grupo: String = "",
    val nombre: String = "",
    val max: Int? = null,
    val descripcion: String = "",
    val variantes: List<String> = emptyList(),
    val nombresVariantes: List<String> = emptyList(),
    val campos: List<CampoBloque> = emptyList(),
) {
    fun nombreVariante(v: String): String = nombresVariantes.getOrNull(variantes.indexOf(v)) ?: v
}

/** Un campo de un bloque (bloques.json › tiposCampo dice qué guarda cada tipo). */
@Serializable
data class CampoBloque(
    val id: String,
    val tipo: String,
    val nombre: String = "",
    val max: Int? = null,
    val min: Int? = null,
    val porDefecto: JsonElement? = null,
    val sugerido: String? = null,
    val requerido: Boolean = false,
    val item: List<CampoBloque>? = null,
    val opciones: List<String>? = null,
    val nombres: List<String>? = null,
)

// MARK: - Presets (assets/presets.json: las 17 plantillas como punto de partida)

@Serializable
data class PresetTienda(
    val direccion: String = "vidriera",
    val estilo: Map<String, String> = emptyMap(),
    val orden: List<String> = emptyList(),
    val variantes: Map<String, String> = emptyMap(),
    val gastronomia: Boolean = false,
) {
    @Serializable private data class Archivo(val presets: Map<String, PresetTienda> = emptyMap())

    /** Los bloques iniciales del preset (id = tipo, como la migración). */
    fun inicial(): List<Bloque> = orden.map { t -> Bloque(t, t, variantes[t] ?: CatalogoBloques.compartido.tipo(t)?.variantes?.firstOrNull() ?: "") }

    companion object {
        val todos: Map<String, PresetTienda> by lazy { Recursos.leer<Archivo>("presets.json")?.presets ?: emptyMap() }
        fun de(plantilla: String): PresetTienda = todos[plantilla] ?: todos["litoral"] ?: PresetTienda()
    }
}

/** Lo que cambia al pasar a otro preset: el orden y las variantes del preset nuevo, conservando los datos de los
 *  bloques del mismo tipo (y los bloques que el preset no trae, al final). El estilo que define el preset vuelve a
 *  ser el suyo. */
fun DisenoTienda.conPreset(plantilla: String): DisenoTienda {
    val p = PresetTienda.de(plantilla)
    val usados = mutableSetOf<String>()
    val nuevo = p.orden.filter { CatalogoBloques.compartido.tipo(it) != null }.map { t ->
        val viejo = inicio.firstOrNull { it.tipo == t && it.id !in usados }
        if (viejo != null) { usados += viejo.id; viejo.copy(variante = p.variantes[t] ?: viejo.variante, oculto = null) }
        else Bloque(Bloque.nuevoId(todos.map { it.id }.toSet() + usados).also { usados += it }, t, p.variantes[t] ?: CatalogoBloques.compartido.tipo(t)?.variantes?.firstOrNull() ?: "")
    } + inicio.filter { it.id !in usados }
    // Lo que trae el preset (el look) se suelta; la marca (logos, menú, redes, barra) queda.
    val e = estilo.copy(direccion = null, forma = null, densidad = null, ancho = null, fotoProducto = null, encabezado = null, pie = null, fuenteTitulos = null,
        fuenteTexto = null, colores = null)
    return copy(plantilla = plantilla, inicio = nuevo, estilo = e, paleta = 0)
}

// MARK: - Migración v1 → v2 (la misma de migrar.ts y mitienda.Migrar; casos en packages/contrato/casos/diseno-v2.json)

object Migrar {
    private val deSeccion = mapOf("hero" to "portada", "cats" to "categorias", "prods" to "productos", "revs" to "resenas", "ig" to "galeria")
    private val ordenV1 = listOf("hero", "cats", "prods", "revs", "ig", "wa")

    fun aV2(v1: Map<String, ValorJSON>?, plantilla: String): Map<String, ValorJSON> {
        if (v1 != null && (v1["version"]?.entero ?: 1) >= 2) return v1
        val preset = PresetTienda.de(v1?.get("plantilla")?.texto?.takeIf { PresetTienda.todos.containsKey(it) } ?: plantilla)
        val d = (v1 ?: emptyMap()).toMutableMap()
        val estilo = (d["estilo"]?.objeto ?: emptyMap()).toMutableMap()
        val textos = estilo["textos"]?.objeto ?: emptyMap()
        val secciones = d["secciones"]?.lista?.mapNotNull { s -> s.objeto?.get("id")?.texto?.let { it to (s.objeto?.get("visible")?.bool ?: true) } } ?: emptyList()
        val visible = secciones.toMap()

        // El orden: el del preset, salvo que la tienda haya reordenado sus secciones.
        val ids = secciones.map { it.first }.filter { it in ordenV1 }
        val reordenada = ids.isNotEmpty() && ids != ordenV1.filter { it in ids }
        val orden = if (!reordenada) preset.orden else {
            val suyo = ids.mapNotNull { deSeccion[it] }.filter { it in preset.orden }
            suyo + preset.orden.filter { it !in suyo }
        }
        val inicio = orden.map { t ->
            val s = deSeccion.entries.firstOrNull { it.value == t }?.key
            val datos = mutableMapOf<String, ValorJSON>()
            if (t == "portada") {
                d["titulo"]?.texto?.takeIf { it.isNotBlank() }?.let { datos["titulo"] = Valor.texto(it) }
                d["lema"]?.texto?.takeIf { it.isNotBlank() }?.let { datos["texto"] = Valor.texto(it) }
                d["portada"]?.texto?.takeIf { it.isNotBlank() }?.let { datos["fotos"] = Valor.lista(listOf(Valor.texto(it))) }
            }
            s?.let { textos[it]?.texto?.takeIf { x -> x.isNotBlank() } }?.let { datos["titulo"] = Valor.texto(it) }
            Bloque(t, t, preset.variantes[t] ?: CatalogoBloques.compartido.tipo(t)?.variantes?.firstOrNull() ?: "", if (s != null && visible[s] == false) true else null, JsonObject(datos)).json
        }
        if (visible["wa"] == false) estilo["whatsappFlotante"] = Valor.booleano(false)
        estilo["anuncio"]?.texto?.takeIf { it.isNotBlank() }?.let { a ->
            val m = mutableMapOf("texto" to Valor.texto(a))
            estilo["anuncioLink"]?.texto?.takeIf { it.isNotBlank() }?.let { m["link"] = Valor.texto(it) }
            estilo["barra"] = Valor.objeto(mapOf("variante" to Valor.texto("fija"), "mensajes" to Valor.lista(listOf(Valor.objeto(m)))))
        }
        estilo.remove("textos"); estilo.remove("anuncio"); estilo.remove("anuncioLink")
        d.remove("secciones")
        d["version"] = Valor.numero(2)
        d["paleta"] = d["paleta"] ?: Valor.numero(0)
        if (estilo.isEmpty()) d.remove("estilo") else d["estilo"] = Valor.objeto(estilo)
        d["inicio"] = Valor.lista(inicio)
        return d
    }
}
