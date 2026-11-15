package ar.com.tulocalweb.app.nucleo

import kotlinx.serialization.Serializable

// Clon de Nucleo/CatalogoBloques.swift.

/** El catálogo único de bloques de la tienda (assets/bloques.json, copia de packages/contrato/bloques.json hecha por
 *  scripts/estilo/para-ios.ts) y los puntos de partida (assets/presets.json). Los campos del editor salen de acá:
 *  sumar un bloque o un campo en el contrato lo suma en la app sin tocar código. */
@Serializable
data class CatalogoBloques(
    val limites: Limites = Limites(),
    val estilo: Map<String, Opciones> = emptyMap(),
    val fondos: Opciones = Opciones(),
    val fuentesProductos: Opciones = Opciones(),
    val iconos: List<String> = emptyList(),
    val paginas: Map<String, Pagina> = emptyMap(),
    val grupos: List<Grupo> = emptyList(),
    val bloques: List<Tipo> = emptyList(),
    val barraAnuncios: Barra = Barra(nombre = "Barra de anuncios", variantes = listOf("fija"), nombresVariantes = listOf("Fija")),
) {
    @Serializable data class Limites(
        val bloquesPorPagina: Int = 30, val itemsPorLista: Int = 12, val bytesDiseno: Int = 65536,
        /** Los topes de un campo tabla (la guía de talles). */
        val filasTabla: Int = 8, val columnasTabla: Int = 8, val letrasCelda: Int = 30,
    )
    @Serializable data class Opciones(val opciones: List<String> = emptyList(), val nombres: List<String> = emptyList(), val porDefecto: String? = null)
    @Serializable data class Grupo(val id: String, val nombre: String)

    /** Un campo de un bloque (tiposCampo en bloques.json): texto, foto, link, lista… */
    @Serializable
    data class Campo(
        val id: String,
        val tipo: String,
        val nombre: String = "",
        val max: Int? = null,
        val min: Int? = null,
        val requerido: Boolean? = null,
        val porDefecto: ValorJSON? = null,
        /** «nombre» o «lema»: si está vacío, la tienda usa ese dato. */
        val sugerido: String? = null,
        /** Los campos de cada elemento (lista) o del objeto (grupo). */
        val item: List<Campo>? = null,
        val opciones: List<String>? = null,
        val nombres: List<String>? = null,
    )

    @Serializable
    data class Tipo(
        val tipo: String,
        val grupo: String = "",
        val nombre: String = "",
        val max: Int? = null,
        val descripcion: String = "",
        val variantes: List<String> = emptyList(),
        val nombresVariantes: List<String> = emptyList(),
        val campos: List<Campo> = emptyList(),
    ) {
        val id: String get() = tipo
        fun nombreVariante(v: String): String = variantes.indexOf(v).takeIf { it >= 0 }?.let { nombresVariantes.getOrNull(it) } ?: v
    }

    @Serializable
    data class Pagina(
        val nombre: String = "",
        val variantes: List<String>? = null,
        val nombresVariantes: List<String>? = null,
        val campos: List<Campo>? = null,
        val bloques: Boolean? = null,
    )

    @Serializable
    data class Barra(
        val nombre: String = "",
        val descripcion: String = "",
        val variantes: List<String> = emptyList(),
        val nombresVariantes: List<String> = emptyList(),
        val campos: List<Campo> = emptyList(),
    )

    fun tipo(t: String): Tipo? = bloques.firstOrNull { it.tipo == t }
    fun bloques(grupo: String): List<Tipo> = bloques.filter { it.grupo == grupo }

    /** Los ajustes nuevos de una página, con la primera variante y lo que trae cada campo por defecto. */
    fun ajustesNuevos(pagina: String): AjustesPagina {
        val p = paginas[pagina] ?: return AjustesPagina()
        val campos = (p.campos ?: emptyList()).mapNotNull { c -> c.porDefecto?.let { c.id to it } }.toMap()
        return AjustesPagina(variante = p.variantes?.firstOrNull(), bloques = if (p.bloques == true) emptyList() else null, campos = campos)
    }

    companion object {
        /** Las páginas del editor, en orden (el JSON es un objeto y no guarda el orden). */
        val ordenPaginas = listOf("inicio", "catalogo", "producto", "carrito", "nosotros")

        val shared: CatalogoBloques by lazy { Recursos.leer<CatalogoBloques>("bloques.json") ?: CatalogoBloques() }

        /** El nombre de una opción (fondos, fuentes de productos, estilo) o la misma si no está. */
        fun nombre(o: Opciones, v: String): String = o.opciones.indexOf(v).takeIf { it >= 0 }?.let { o.nombres.getOrNull(it) } ?: v
    }
}

/** Un punto de partida (una de las 17 plantillas): dirección, estilo base, orden y variantes de los bloques. */
@Serializable
data class PresetTienda(
    val direccion: String = "vidriera",
    val estilo: Map<String, String> = emptyMap(),
    val orden: List<String> = emptyList(),
    val variantes: Map<String, String> = emptyMap(),
    val gastronomia: Boolean? = null,
) {
    @Serializable private data class Archivo(val presets: Map<String, PresetTienda> = emptyMap())

    /** Los bloques del inicio recién armado con este punto de partida (id = tipo, como la migración). */
    val inicio: List<Bloque> get() = orden.map { Bloque(it, it, variantes[it] ?: CatalogoBloques.shared.tipo(it)?.variantes?.firstOrNull() ?: "") }

    companion object {
        val todos: Map<String, PresetTienda> by lazy { Recursos.leer<Archivo>("presets.json")?.presets ?: emptyMap() }
    }
}
