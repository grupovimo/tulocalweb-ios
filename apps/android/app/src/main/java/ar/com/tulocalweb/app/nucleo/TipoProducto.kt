package ar.com.tulocalweb.app.nucleo

/** Qué es el producto (API: migración 00071), clon de Nucleo/TipoProducto.swift. El editor pregunta primero y según
 *  eso pide sus datos: la ropa, talle y color; la comida, el contenido; un mueble, peso y medidas. La misma lista está
 *  en Go (internal/dominio/listas/tipo.go) y en la web (src/lib/tienda/tipo-producto.ts). */
enum class TipoProducto {
    ropa, calzado, alimentos, belleza, hogar, tecnologia, otro;

    val id: String get() = name

    val nombre: String get() = when (this) {
        ropa -> "Ropa y accesorios"
        calzado -> "Calzado"
        alimentos -> "Comida y bebida"
        belleza -> "Belleza y cuidado"
        hogar -> "Hogar, deco y objetos"
        tecnologia -> "Tecnología"
        otro -> "Otra cosa"
    }

    val ejemplo: String get() = when (this) {
        ropa -> "Remeras, buzos, carteras"
        calzado -> "Zapatillas, botas, sandalias"
        alimentos -> "Yerba, vinos, budines"
        belleza -> "Cremas, perfumes, maquillaje"
        hogar -> "Muebles, mates, macetas"
        tecnologia -> "Celulares, auriculares, cables"
        otro -> "Lo que no entra en ninguno"
    }

    val simbolo: String get() = when (this) {
        ropa -> "tshirt"
        calzado -> "shoe"
        alimentos -> "fork.knife"
        belleza -> "sparkles"
        hogar -> "sofa"
        tecnologia -> "iphone"
        otro -> "ellipsis.circle"
    }

    /** Los datos que pide, en orden (alto, ancho y largo van juntos como «medidas»). */
    val campos: List<AtributoProducto> get() = when (this) {
        ropa, calzado -> listOf(AtributoProducto.material, AtributoProducto.peso)
        alimentos, belleza -> listOf(AtributoProducto.contenido, AtributoProducto.marca, AtributoProducto.peso)
        hogar -> listOf(AtributoProducto.material, AtributoProducto.peso, AtributoProducto.alto, AtributoProducto.ancho, AtributoProducto.largo)
        tecnologia -> listOf(AtributoProducto.marca, AtributoProducto.modelo, AtributoProducto.garantia, AtributoProducto.peso, AtributoProducto.alto, AtributoProducto.ancho, AtributoProducto.largo)
        otro -> listOf(AtributoProducto.peso, AtributoProducto.alto, AtributoProducto.ancho, AtributoProducto.largo)
    }

    /** Las opciones de variante que propone, con valores para arrancar. */
    val opciones: List<Pair<String, List<String>>> get() = when (this) {
        ropa -> listOf("Talle" to listOf("XS", "S", "M", "L", "XL", "XXL"), "Color" to emptyList())
        calzado -> listOf("Talle" to listOf("35", "36", "37", "38", "39", "40", "41", "42", "43", "44"), "Color" to emptyList())
        alimentos -> listOf("Presentación" to emptyList(), "Sabor" to emptyList())
        belleza -> listOf("Presentación" to emptyList(), "Fragancia" to emptyList())
        hogar -> listOf("Color" to emptyList(), "Tamaño" to listOf("Chico", "Mediano", "Grande"))
        tecnologia -> listOf("Color" to emptyList(), "Capacidad" to emptyList())
        otro -> listOf("Modelo" to emptyList(), "Color" to emptyList())
    }

    companion object {
        fun de(id: String?): TipoProducto? = entries.firstOrNull { it.name == id }

        /** Para los productos de antes, sin tipo: lo que dicen sus opciones. */
        fun adivinar(opciones: Map<String, List<String>>): TipoProducto? {
            val talle = opciones.entries.firstOrNull { it.key.lowercase().startsWith("talle") }?.value ?: return null
            val esCalzado = talle.isNotEmpty() && talle.all { v -> v.toIntOrNull()?.let { it in 34..46 } ?: false }
            return if (esCalzado) TipoProducto.calzado else TipoProducto.ropa
        }
    }
}

enum class AtributoProducto {
    peso, alto, ancho, largo, contenido, material, marca, modelo, garantia;

    val nombre: String get() = when (this) {
        peso -> "Peso"
        alto -> "Alto"
        ancho -> "Ancho"
        largo -> "Largo"
        contenido -> "Contenido"
        material -> "Material"
        marca -> "Marca"
        modelo -> "Modelo"
        garantia -> "Garantía"
    }

    /** La unidad de los numéricos (gramos y centímetros); null: texto libre. */
    val unidad: String? get() = when (this) {
        peso -> "g"
        alto, ancho, largo -> "cm"
        else -> null
    }

    val ejemplo: String get() = when (this) {
        contenido -> "Ej.: 500 g o 750 ml"
        material -> "Ej.: algodón"
        marca -> "La del fabricante o la tuya"
        modelo -> "Ej.: A15"
        garantia -> "Ej.: 6 meses"
        else -> "0"
    }

    val esMedida: Boolean get() = this == alto || this == ancho || this == largo
}
