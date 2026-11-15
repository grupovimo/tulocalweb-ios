package ar.com.tulocalweb.app.pantallas.mas

import android.os.SystemClock
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import ar.com.tulocalweb.app.nucleo.AjustesPagina
import ar.com.tulocalweb.app.nucleo.Bloque
import ar.com.tulocalweb.app.nucleo.CatalogoBloques
import ar.com.tulocalweb.app.nucleo.Diseno
import ar.com.tulocalweb.app.nucleo.Historial
import ar.com.tulocalweb.app.nucleo.ValorJSON
import ar.com.tulocalweb.app.nucleo.VentanaTienda

// Clon de Pantallas/Mas/EditorTienda.swift.

/** El estado del editor de la tienda: el diseño, la página elegida, el bloque abierto y el historial para deshacer y
 *  rehacer. Todas las ediciones pasan por `cambiar`, así cada paso se puede deshacer. */
class EditorTienda {
    var diseno by mutableStateOf(Diseno())
    /** Lo último guardado (para saber si hay cambios sin guardar). */
    var guardado by mutableStateOf(Diseno())
    var plantillaTienda by mutableStateOf("")
    var pagina by mutableStateOf("inicio")
    /** El bloque resaltado en la lista y en la vista previa. */
    var seleccion by mutableStateOf<String?>(null)
    /** El bloque con el panel abierto. */
    var panel by mutableStateOf<String?>(null)
    /** Dónde va el bloque que se elija en la biblioteca. */
    var insertarEn by mutableStateOf<Int?>(null)
    /** La ventana (popup) con el panel abierto. */
    var ventanaAbierta by mutableStateOf<String?>(null)
    /** La hoja de plantillas para crear una ventana. */
    var creandoVentana by mutableStateOf(false)
    var historial by mutableStateOf(Historial())
        private set
    /** Cambia con cada paso del historial (Historial no es observable). */
    var pasos by mutableStateOf(0)
        private set
    private var ultimaClave: String? = null
    private var ultimaVez = 0L

    private val cat = CatalogoBloques.shared

    val plantilla: String get() = diseno.plantilla ?: plantillaTienda
    val sinGuardar: Boolean get() = diseno != guardado
    val bloques: List<Bloque> get() = diseno.bloques(pagina)
    val paginaConBloques: Boolean get() = pagina == "inicio" || cat.paginas[pagina]?.bloques == true

    fun empezar(d: Diseno, plantilla: String) {
        diseno = d
        guardado = d
        plantillaTienda = plantilla
        historial = Historial()
        pasos += 1
    }

    // MARK: Cambios con historial

    /** Aplica un cambio guardando cómo estaba. Con `clave`, los cambios seguidos de lo mismo (escribir en un campo)
     *  quedan en un solo paso. */
    fun cambiar(clave: String? = null, f: (Diseno) -> Diseno) {
        val d = f(diseno)
        if (d == diseno) return
        val ahora = SystemClock.uptimeMillis()
        if (clave == null || clave != ultimaClave || ahora - ultimaVez > 1500) historial.guardar(diseno)
        ultimaClave = clave
        ultimaVez = ahora
        diseno = d
        pasos += 1
    }

    fun deshacer() {
        val d = historial.deshacer(diseno) ?: return
        ultimaClave = null
        diseno = d
        pasos += 1
        cerrarSiNoEsta()
    }

    fun rehacer() {
        val d = historial.rehacer(diseno) ?: return
        ultimaClave = null
        diseno = d
        pasos += 1
        cerrarSiNoEsta()
    }

    private fun cerrarSiNoEsta() {
        panel?.let { if (diseno.buscar(it) == null) panel = null }
        seleccion?.let { if (diseno.buscar(it) == null) seleccion = null }
        ventanaAbierta?.let { if (ventana(it) == null) ventanaAbierta = null }
    }

    // MARK: Bloques de la página elegida

    fun bloque(id: String): Bloque? = diseno.buscar(id)?.second

    /** Cambia un bloque esté en la página que esté. */
    fun editar(id: String, clave: String? = null, f: (Bloque) -> Bloque) {
        val p = diseno.buscar(id)?.first ?: return
        cambiar(clave) { d -> d.ponerBloques(d.bloques(p).map { if (it.id == id) f(it) else it }, p) }
    }

    /** Si se puede sumar un bloque de este tipo a la página (tope por tipo y por página). */
    fun puedeAgregar(t: CatalogoBloques.Tipo): Boolean {
        if (bloques.size >= cat.limites.bloquesPorPagina) return false
        t.max?.let { max -> return bloques.count { it.tipo == t.tipo } < max }
        return true
    }

    /** Suma un bloque y abre su panel (con `abrir` false, lo abre quien llama: después de cerrar la biblioteca). */
    fun agregar(t: CatalogoBloques.Tipo, en: Int, abrir: Boolean = true): String {
        var b = Bloque.nuevo(t)
        while (b.id in diseno.idsUsados) b = b.copy(id = Bloque.nuevoId())
        val nuevo = b
        val p = pagina
        cambiar { d ->
            val l = d.bloques(p).toMutableList()
            l.add(en.coerceIn(0, l.size), nuevo)
            d.ponerBloques(l, p)
        }
        seleccion = nuevo.id
        if (abrir) panel = nuevo.id
        return nuevo.id
    }

    fun duplicar(id: String) {
        val (p, b) = diseno.buscar(id) ?: return
        val t = cat.tipo(b.tipo) ?: return
        if (!puedeAgregar(t) && p == pagina) return
        var copia = b.copy(id = Bloque.nuevoId())
        while (copia.id in diseno.idsUsados) copia = copia.copy(id = Bloque.nuevoId())
        val nueva = copia
        cambiar { d ->
            val l = d.bloques(p).toMutableList()
            val i = l.indexOfFirst { it.id == id }
            if (i >= 0) l.add(i + 1, nueva)
            d.ponerBloques(l, p)
        }
        seleccion = nueva.id
    }

    fun alternarOculto(id: String) = editar(id) { it.copy(oculto = !it.oculto) }

    fun borrar(id: String) {
        val p = diseno.buscar(id)?.first ?: return
        cambiar { d -> d.ponerBloques(d.bloques(p).filter { it.id != id }, p) }
        if (panel == id) panel = null
        if (seleccion == id) seleccion = null
    }

    /** Mueve un bloque arriba del otro (arrastrar y soltar); sin destino, al final. */
    fun mover(id: String, antesDe: String?) {
        if (id == antesDe) return
        val p = pagina
        cambiar("mover") { d ->
            val l = d.bloques(p).toMutableList()
            val i = l.indexOfFirst { it.id == id }
            if (i < 0) return@cambiar d
            val b = l.removeAt(i)
            val j = antesDe?.let { x -> l.indexOfFirst { it.id == x }.takeIf { it >= 0 } } ?: l.size
            l.add(j, b)
            d.ponerBloques(l, p)
        }
    }

    fun mover(id: String, paso: Int) {
        val p = pagina
        cambiar { d ->
            val l = d.bloques(p).toMutableList()
            val i = l.indexOfFirst { it.id == id }
            if (i < 0 || i + paso !in l.indices) return@cambiar d
            val x = l[i]; l[i] = l[i + paso]; l[i + paso] = x
            d.ponerBloques(l, p)
        }
    }

    fun abrir(id: String) {
        diseno.buscar(id)?.first?.let { if (it != pagina) pagina = it }
        seleccion = id
        panel = id
    }

    // MARK: Página y punto de partida

    fun ajustes(p: String): AjustesPagina = diseno.ajustes(p)

    fun editarAjustes(p: String, clave: String? = null, f: (AjustesPagina) -> AjustesPagina) {
        cambiar(clave) { d -> d.copy(paginas = d.paginas + (p to f(d.ajustes(p)))) }
    }

    fun aplicarPreset(id: String) = cambiar { it.aplicarPreset(id) }

    // MARK: Ventanas (Promociones)

    val topeVentanas: Int get() = cat.limites.ventanas
    val puedeSumarVentana: Boolean get() = diseno.ventanas.size < topeVentanas

    fun ventana(id: String): VentanaTienda? = diseno.ventanas.firstOrNull { it.id == id }

    /** Un id que no usa ningún bloque ni otra ventana. */
    fun idNuevo(): String {
        var id = Bloque.nuevoId()
        while (id in diseno.idsUsados || ventana(id) != null) id = Bloque.nuevoId()
        return id
    }

    /** Suma una ventana al final y abre su panel. */
    fun agregarVentana(v: VentanaTienda, abrir: Boolean = true): String? {
        if (!puedeSumarVentana) return null
        cambiar { it.copy(ventanas = it.ventanas + v) }
        if (abrir) ventanaAbierta = v.id
        return v.id
    }

    fun editarVentana(id: String, clave: String? = null, f: (VentanaTienda) -> VentanaTienda) {
        cambiar(clave) { d -> d.copy(ventanas = d.ventanas.map { if (it.id == id) f(it) else it }) }
    }

    fun alternarVentana(id: String) = editarVentana(id) { it.copy(activa = !it.activa) }

    fun duplicarVentana(id: String) {
        if (!puedeSumarVentana) return
        val v = ventana(id) ?: return
        val nueva = v.copy(id = idNuevo(), nombre = "${v.nombre} (copia)".take(40), activa = false)
        cambiar { d ->
            val l = d.ventanas.toMutableList()
            val i = l.indexOfFirst { it.id == id }
            if (i >= 0) l.add(i + 1, nueva)
            d.copy(ventanas = l)
        }
    }

    fun borrarVentana(id: String) {
        cambiar { d -> d.copy(ventanas = d.ventanas.filter { it.id != id }) }
        if (ventanaAbierta == id) ventanaAbierta = null
    }

    fun moverVentana(id: String, paso: Int) {
        cambiar { d ->
            val l = d.ventanas.toMutableList()
            val i = l.indexOfFirst { it.id == id }
            if (i < 0 || i + paso !in l.indices) return@cambiar d
            val x = l[i]; l[i] = l[i + paso]; l[i + paso] = x
            d.copy(ventanas = l)
        }
    }

    /** Las herramientas de conversión (estilo.conversion), con su paso en el historial. */
    fun editarConversion(c: Map<String, ValorJSON>) {
        cambiar("conversion") { it.copy(estilo = it.estilo.copy(conversion = c.ifEmpty { null })) }
    }
}
