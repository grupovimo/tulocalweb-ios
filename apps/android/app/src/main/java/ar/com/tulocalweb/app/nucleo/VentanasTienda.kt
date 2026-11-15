package ar.com.tulocalweb.app.nucleo

import kotlinx.serialization.json.JsonArray
import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.JsonPrimitive
import java.time.Instant
import java.time.ZoneId
import java.time.ZonedDateTime
import java.time.format.DateTimeFormatter

// Clon de Nucleo/VentanasTienda.swift.
// Las ventanas (popups) de la tienda: diseno.ventanas (docs/especificaciones/tienda-promociones.md). Lee y escribe el
// JSON de la especificación; las opciones, los campos y las plantillas salen de bloques.json › ventanas.

/** Una ventana: qué muestra (contenido y sus datos), cómo se ve (variante), cuándo sale, a quién, dónde y entre qué fechas. */
data class VentanaTienda(
    val id: String,
    val nombre: String,
    val activa: Boolean = true,
    val variante: String = "centrada",
    val contenido: String = "promo",
    val datos: Map<String, ValorJSON> = emptyMap(),
    val cuando: String = "entrar",
    /** Segundos o % de la página, según el disparo (sólo los que traen rango). */
    val valor: Int? = null,
    val cada: String = "sesion",
    val dias: Int? = null,
    val publico: String = "todos",
    /** Vacío: en todas. */
    val paginas: List<String> = emptyList(),
    val dispositivos: String = "todos",
    val desde: String? = null,
    val hasta: String? = null,
) {
    /** Para el PUT: `valor` y `dias` sólo en los disparos y frecuencias que traen rango (con el de por defecto si falta). */
    val json: ValorJSON get() {
        val cat = CatalogoBloques.shared.ventanas
        val disparo = mutableMapOf<String, ValorJSON>("cuando" to Valor.texto(cuando))
        cat?.disparo(cuando)?.valor?.let { r -> disparo["valor"] = Valor.numero((valor ?: r.porDefecto).coerceIn(r.min, r.max)) }
        val frecuencia = mutableMapOf<String, ValorJSON>("cada" to Valor.texto(cada))
        cat?.frecuencia(cada)?.dias?.let { r -> frecuencia["dias"] = Valor.numero((dias ?: r.porDefecto).coerceIn(r.min, r.max)) }
        val n = nombre.trim()
        val o = mutableMapOf(
            "id" to Valor.texto(id), "nombre" to Valor.texto(if (n.isEmpty()) "Ventana" else n.take(40)), "activa" to Valor.booleano(activa),
            "variante" to Valor.texto(variante), "contenido" to Valor.texto(contenido), "datos" to Valor.objeto(datos),
            "disparo" to Valor.objeto(disparo), "frecuencia" to Valor.objeto(frecuencia), "publico" to Valor.texto(publico),
            "paginas" to Valor.lista(paginas.map { Valor.texto(it) }), "dispositivos" to Valor.texto(dispositivos),
        )
        if (!desde.isNullOrEmpty()) o["desde"] = Valor.texto(desde)
        if (!hasta.isNullOrEmpty()) o["hasta"] = Valor.texto(hasta)
        return Valor.objeto(o)
    }

    /** Al cambiar de contenido quedan sólo los datos que el nuevo también tiene (título, texto, foto…). */
    fun conContenido(c: String): VentanaTienda {
        if (c == contenido) return this
        val campos = CatalogoBloques.shared.ventanas?.contenido(c)?.campos?.map { it.id }?.toSet() ?: emptySet()
        return copy(contenido = c, datos = datos.filterKeys { it in campos })
    }

    companion object {
        fun desde(v: ValorJSON): VentanaTienda = VentanaTienda(
            id = v["id"]?.texto ?: Bloque.nuevoId(),
            nombre = v["nombre"]?.texto ?: "",
            activa = v["activa"]?.bool ?: true,
            variante = v["variante"]?.texto ?: "centrada",
            contenido = v["contenido"]?.texto ?: "promo",
            datos = v["datos"]?.objeto ?: emptyMap(),
            cuando = v["disparo"]?.get("cuando")?.texto ?: "entrar",
            valor = v["disparo"]?.get("valor")?.entero,
            cada = v["frecuencia"]?.get("cada")?.texto ?: "sesion",
            dias = v["frecuencia"]?.get("dias")?.entero,
            publico = v["publico"]?.texto ?: "todos",
            paginas = v["paginas"]?.lista?.mapNotNull { it.texto } ?: emptyList(),
            dispositivos = v["dispositivos"]?.texto ?: "todos",
            desde = v["desde"]?.texto,
            hasta = v["hasta"]?.texto,
        )

        /** Una en blanco: una promo centrada al entrar, una vez por visita. */
        fun enBlanco(id: String): VentanaTienda = VentanaTienda(id, "Ventana nueva")

        /** Desde una plantilla del catálogo: id nuevo, el monto del envío gratis en los textos y, si la trae, la cuenta
         *  regresiva a tantos días. */
        fun dePlantilla(p: CatalogoBloques.Ventanas.Plantilla, id: String, envioGratis: Int?, ahora: Instant = Instant.now()): VentanaTienda {
            val monto = envioGratis?.let { Formato.pesos(it) } ?: "un monto"
            var v = desde(reemplazar(p.ventana, "{envioGratis}", monto)).copy(id = id)
            p.hastaEnDias?.let { d -> v = v.copy(datos = v.datos + ("hasta" to Valor.texto(isoAR(ahora.atZone(zonaAR).plusDays(d.toLong()))))) }
            return v
        }

        private fun reemplazar(v: ValorJSON, a: String, b: String): ValorJSON = when {
            v is JsonPrimitive && v.isString -> JsonPrimitive(v.content.replace(a, b))
            v is JsonArray -> JsonArray(v.map { reemplazar(it, a, b) })
            v is JsonObject -> JsonObject(v.mapValues { reemplazar(it.value, a, b) })
            else -> v
        }

        val zonaAR: ZoneId = ZoneId.of("America/Argentina/Buenos_Aires")
        fun isoAR(f: ZonedDateTime): String = f.withZoneSameInstant(zonaAR).withNano(0).format(DateTimeFormatter.ISO_OFFSET_DATE_TIME)
        fun isoAR(f: Instant): String = isoAR(f.atZone(zonaAR))
    }
}

// MARK: - Qué le falta

/** Por qué una ventana no va a salir en la tienda, en una frase. Las mismas reglas que Go (VentanaTienda.Completa) y la web: los
 *  requeridos del contenido cargados, el premio con lo suyo y, en la ruleta, algún peso mayor a 0. */
object FaltaVentana {
    fun de(v: VentanaTienda, ahora: Instant = Instant.now()): String? {
        val cont = CatalogoBloques.shared.ventanas?.contenido(v.contenido) ?: return null
        for (c in cont.campos) {
            val x = v.datos[c.id]
            if (c.requerido == true && (x?.vacio ?: true)) {
                return when (c.tipo) {
                    "cupon" -> "Elegí el cupón."
                    "premio" -> "Elegí el premio."
                    "lista" -> "Sumá ${c.nombre.lowercase()}."
                    else -> "Completá «${c.nombre}»."
                }
            }
            if (c.tipo == "premio") dePremio(x)?.let { return it }
            val min = c.min
            if (c.tipo == "lista" && min != null && (x?.lista?.size ?: 0) < min) return "Sumá al menos $min en «${c.nombre}»."
        }
        if (v.contenido == "ruleta") {
            val porciones = v.datos["porciones"]?.lista ?: emptyList()
            porciones.forEachIndexed { i, p ->
                if ((p["texto"]?.texto ?: "").isBlank()) return "Poné el texto de la porción ${i + 1}."
                dePremio(p["premio"])?.let { f -> return "Porción ${i + 1}: ${f.take(1).lowercase()}${f.drop(1)}" }
            }
            if (porciones.sumOf { it["peso"]?.numero ?: 0.0 } <= 0) return "Poné un peso mayor a 0 en alguna porción."
        }
        v.datos["hasta"]?.texto?.let { Fecha.iso(it) }?.let { if (it.isBefore(ahora)) return "La cuenta regresiva ya terminó: cambiá la fecha." }
        v.hasta?.let { Fecha.iso(it) }?.let { if (it.isBefore(ahora)) return "Ya pasó la fecha en que termina." }
        val fd = v.desde?.let { Fecha.iso(it) }
        val fh = v.hasta?.let { Fecha.iso(it) }
        if (fd != null && fh != null && !fh.isAfter(fd)) return "La fecha de fin va después de la de inicio."
        return null
    }

    /** El premio trae lo que necesita para entregarse. */
    fun dePremio(v: ValorJSON?): String? {
        val p = v?.objeto ?: return "Elegí el premio."
        return when (p["tipo"]?.texto) {
            "cuponFijo" -> if (p["cupon"]?.vacio ?: true) "Elegí el cupón del premio." else null
            "cuponUnico" -> {
                val d = p["descuento"]?.texto ?: ""
                when {
                    d == "envioGratis" -> null
                    d != "porcentaje" && d != "monto" -> "Elegí qué descuento da el premio."
                    p["valor"] == null -> "Poné de cuánto es el descuento."
                    else -> null
                }
            }
            "descarga" -> if (p["archivo"]?.vacio ?: true) "Poné el link del archivo para descargar." else null
            "nada" -> null
            else -> "Elegí el premio."
        }
    }
}

// MARK: - Métricas

/** Lo que pasó con una ventana en el período (GET /panel/promos/metricas): vistas, cierres, envíos, clics y ventas. */
data class MetricaVentana(
    val vistas: Int = 0,
    val cierres: Int = 0,
    val envios: Int = 0,
    val clics: Int = 0,
    val ventas: Int = 0,
    val monto: Int = 0,
) {
    /** Qué parte de los que la vieron hizo algo: dejó el mail (captación, ruleta) o tocó (promo, cupón). */
    fun tasa(contenido: String): Double? {
        if (vistas <= 0) return null
        val n = if (contenido == "captacion" || contenido == "ruleta") envios else clics
        return n.toDouble() / vistas * 100
    }

    companion object {
        /** Por ventana (GET /panel/promos/metricas › ventanas[]: id, vistas, cierres, envios, clics, ventas, monto, dias[]).
         *  También acepta una lista de filas por día o un objeto por id. */
        fun leer(v: ValorJSON): Map<String, MetricaVentana> {
            val r = mutableMapOf<String, MetricaVentana>()
            fun sumar(id: String, o: ValorJSON) {
                // La fila de la API (PromoMetricaVentana) trae el total del período; si sólo trae los días, se suman.
                if (o["vistas"] == null) o["dias"]?.lista?.let { dias -> dias.forEach { sumar(id, it) }; return }
                val m = r[id] ?: MetricaVentana()
                r[id] = m.copy(
                    vistas = m.vistas + (o["vistas"]?.entero ?: 0), cierres = m.cierres + (o["cierres"]?.entero ?: 0),
                    envios = m.envios + (o["envios"]?.entero ?: 0), clics = m.clics + (o["clics"]?.entero ?: 0),
                    ventas = m.ventas + (o["ventas"]?.entero ?: 0), monto = m.monto + (o["monto"]?.entero ?: 0),
                )
            }
            fun filas(l: List<ValorJSON>) = l.forEach { f -> (f["ventana"]?.texto ?: f["id"]?.texto)?.let { sumar(it, f) } }
            when (v) {
                is JsonArray -> filas(v)
                is JsonObject -> {
                    val vs = v["ventanas"]
                    when {
                        vs is JsonArray -> filas(vs)
                        v["filas"] is JsonArray -> filas(v["filas"] as JsonArray)
                        vs is JsonObject -> vs.forEach { (id, x) -> sumar(id, x) }
                        else -> v.forEach { (id, x) -> if (x is JsonObject) sumar(id, x) }
                    }
                }
                else -> {}
            }
            return r
        }
    }
}
