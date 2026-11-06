package ar.com.tulocalweb.app.nucleo

import ar.com.tulocalweb.app.R
import kotlinx.serialization.Serializable
import java.time.Instant

// Clon de Nucleo/Modelos.swift: lo que la app lee de la API (packages/contrato/openapi.yaml). Sólo los campos que usa;
// lo demás se ignora. Los montos vienen en pesos enteros.

// MARK: - Sesión

@Serializable
data class SesionNueva(
    val token: String = "",
    val venceEn: FechaApi? = null,
    val tiendaId: String = "",
    /** Sólo en el ingreso con Clerk: la cuenta se creó recién (sigue al onboarding). */
    val nueva: Boolean? = null,
    /** La cuenta tiene la verificación en dos pasos: no hay token todavía, sino este desafío, que se completa con el
     *  código de la app autenticadora (POST /v1/auth/segundo-factor). */
    val segundoFactor: String? = null,
)

/** GET /v1/auth/segundo-factor. */
@Serializable
data class EstadoDosPasos(val activo: Boolean = false, val codigosRestantes: Int = 0)

@Serializable
data class DatosSesion(val usuario: Usuario, val tiendaId: String, val rol: String, val venceEn: FechaApi) {
    @Serializable
    data class Usuario(val id: String, val nombre: String, val email: String, val foto: String? = null)

    /** El nombre de pila, para «Hola, Sofía». */
    val nombreDePila: String get() = usuario.nombre.split(" ").firstOrNull { it.isNotEmpty() } ?: usuario.nombre
}

@Serializable
data class Acceso(val tiendaId: String, val rol: String, val permisos: List<String> = emptyList(), val tiendas: List<Tienda> = emptyList()) {
    @Serializable
    data class Tienda(
        val id: String,
        val nombre: String,
        val slug: String,
        val rol: String = "",
        /** El logo que se cargó en Mi tienda (foto subida o URL); null: sus iniciales. */
        val logo: String? = null,
    ) {
        /** «almacen-lumbre.tulocalweb.com.ar». */
        val direccion: String get() = "$slug.tulocalweb.com.ar"
        /** El link de la tienda, para copiar y compartir. */
        val url: String get() = "https://$direccion"
        /** El link de un producto (la tienda acepta el id y redirige al slug). */
        fun url(producto: String): String = "$url/producto/$producto"
    }

    val tienda: Tienda? get() = tiendas.firstOrNull { it.id == tiendaId }
    fun puede(p: String): Boolean = p in permisos
}

/** «dueña», «administradora»… como lo dice el handoff (1d). */
fun nombreDelRol(rol: String): String = when (rol) {
    "duenio" -> "dueña"
    "admin" -> "administración"
    "ventas" -> "ventas"
    "local" -> "local"
    else -> rol
}

@Serializable
data class Proveedores(val clerk: Boolean = false, val clerkPublicable: String? = null)

// MARK: - Inicio (tablero)

@Serializable
data class Comparado(val actual: Double = 0.0, val anterior: Double = 0.0) {
    /** Variación en %, o null si antes no había nada con qué comparar. */
    val variacion: Double? get() = if (anterior > 0) (actual - anterior) / anterior * 100 else null
}

@Serializable
data class Tablero(
    val periodo: String = "",
    val desde: FechaApi = Instant.EPOCH,
    val hasta: FechaApi = Instant.EPOCH,
    val hola: String? = null,
    val montos: Boolean = false,
    val ventas: Comparado? = null,
    val serie: Serie? = null,
    val kpis: Kpis = Kpis(),
    val canales: List<Canal>? = null,
    val meta: Meta? = null,
    val ultimosPedidos: List<PedidoBreve> = emptyList(),
    val paraHacer: ParaHacer = ParaHacer(),
    val hornero: Hornero = Hornero(),
    val masVendidos: List<MasVendido> = emptyList(),
    val visitas: Visitas? = null,
    val resenas: Resenas? = null,
    // Lo que el Inicio de la web ya mostraba (opcional: la API lo manda según el rol y lo conectado).
    val enVivo: EnVivo? = null,
    val caja: Caja? = null,
    val liquidaciones: List<Liquidacion>? = null,
    val bandeja: List<Chat>? = null,
    val integraciones: List<Integracion>? = null,
    val envios: Envios? = null,
) {
    @Serializable data class Serie(val actual: List<Double> = emptyList(), val anterior: List<Double> = emptyList())

    @Serializable
    data class Kpis(
        val pedidos: Comparado = Comparado(),
        val ticket: Comparado? = null,
        val clientesNuevos: Comparado = Comparado(),
        /** Pedidos web por cada 100 visitas (en cero mientras no se midan visitas). */
        val conversion: Comparado? = null,
    )

    @Serializable data class Canal(val canal: String, val parte: Double)
    @Serializable data class Meta(val objetivo: Int, val actual: Int, val proyeccion: Int)

    @Serializable
    data class ParaHacer(val preparar: Int = 0, val transferencias: Int = 0, val transferenciasMonto: Int? = null, val sinStock: Int = 0, val sinStockEnMl: Int = 0)

    @Serializable
    data class Hornero(
        val carritos: Int = 0,
        val carritosMonto: Int? = null,
        /** Alguno de los carritos ya recibió el mail de recordatorio. */
        val recordados: Boolean? = null,
        val ideasHechas: List<String> = emptyList(),
    )

    @Serializable
    data class MasVendido(val productoId: String, val nombre: String, val foto: String? = null, val unidades: Int = 0) {
        val id: String get() = productoId
    }

    @Serializable
    data class EnVivo(val personas: Int = 0, val paginas: List<Pagina> = emptyList(), val carritos: Int = 0, val carritosMonto: Int? = null) {
        @Serializable data class Pagina(val nombre: String, val personas: Int)
    }

    @Serializable
    data class Caja(val abierta: String? = null, val total: Int = 0, val medios: List<Medio> = emptyList()) {
        @Serializable data class Medio(val medio: String, val monto: Int)
    }

    @Serializable data class Liquidacion(val medio: String, val fecha: String, val monto: Int)

    @Serializable
    data class Chat(val nombre: String, val foto: String? = null, val mensaje: String = "", val creado: FechaApi = Instant.EPOCH, val atiende: String = "", val sinLeer: Int = 0)

    @Serializable data class Integracion(val id: String, val valor: Int)

    @Serializable
    data class Envios(val paquetes: Int = 0, val retiros: List<Retiro> = emptyList()) {
        @Serializable data class Retiro(val hora: String, val transporte: String, val paquetes: Int, val ciudad: String? = null, val proximo: Boolean = false)
    }

    /** Las visitas de la tienda online (personas distintas por día). `desde` en null: todavía no hay. */
    @Serializable
    data class Visitas(
        val desde: String? = null,
        val hoy: Comparado = Comparado(),
        val dias7: Comparado = Comparado(),
        val dias30: Comparado = Comparado(),
        /** Las del período del tablero. */
        val visitantes: Comparado = Comparado(),
        val vistas: Comparado = Comparado(),
        val masVistos: List<Producto> = emptyList(),
        val paginas: List<Pagina>? = null,
    ) {
        @Serializable
        data class Producto(val productoId: String, val nombre: String, val foto: String? = null, val vistas: Int = 0, val visitantes: Int = 0) {
            val id: String get() = productoId
        }
        @Serializable data class Pagina(val pagina: String, val vistas: Int)
    }

    @Serializable
    data class Resenas(
        val promedio: Double? = null,
        /** Las últimas sin responder (hasta tres). */
        val nuevas: List<ValorJSON> = emptyList(),
    )
}

@Serializable
data class PedidoBreve(
    val numero: Int,
    val cliente: String = "",
    val foto: String? = null,
    val canal: String = "",
    val estado: String = "",
    val total: Int? = null,
    val creado: FechaApi = Instant.EPOCH,
) {
    val id: Int get() = numero
}

@Serializable
data class ResumenPanel(
    val insignias: Insignias = Insignias(),
    val prueba: Prueba = Prueba(),
    /** Canales conectados (wa, ml, mp, ig, tg). */
    val canales: List<String>? = null,
) {
    @Serializable data class Insignias(val pedidos: Int = 0, val carritos: Int = 0)
    @Serializable data class Prueba(val plan: String = "", val diasRestantes: Int = 0, val diasTotales: Int = 0)
}

// MARK: - Pedidos

@Serializable
data class PedidoFila(
    val numero: Int,
    val cliente: String = "",
    val foto: String? = null,
    val canal: String = "",
    val pago: String = "",
    val envio: String = "",
    val estado: String = "",
    val total: Int = 0,
    val creado: FechaApi = Instant.EPOCH,
    val productos: List<Producto> = emptyList(),
    val cantidadProductos: Int = 0,
    /** El WhatsApp de quien compró, tal como lo cargó. */
    val telefono: String? = null,
) {
    @Serializable data class Producto(val nombre: String, val foto: String? = null)

    val id: Int get() = numero
    /** La foto del primer producto, para el cuadrito de la fila. */
    val fotoProducto: String? get() = productos.firstOrNull()?.foto
}

@Serializable
data class PaginaPedidos(val total: Int = 0, val suma: Int = 0, val filas: List<PedidoFila> = emptyList(), val etapas: Map<String, Int> = emptyMap())

@Serializable
data class PedidoDetalle(
    val numero: Int,
    val cliente: String = "",
    val canal: String = "",
    val pago: String = "",
    val envio: String = "",
    val estado: String = "",
    val total: Int = 0,
    val creado: FechaApi = Instant.EPOCH,
    val cantidadProductos: Int = 0,
    val items: List<Item> = emptyList(),
    val subtotal: Int = 0,
    val costoEnvio: Int = 0,
    val descuentos: List<Descuento> = emptyList(),
    val comprador: Comprador = Comprador(),
    val estadoPago: String = "",
    val factura: String? = null,
    val seguimiento: String? = null,
    val llega: FechaApi? = null,
    val nota: String? = null,
    val entrega: Entrega? = null,
    val historial: List<Paso> = emptyList(),
    /** Código público de la confirmación (/t/<tienda>/pedido/<codigo>); null si se cargó a mano. */
    val codigo: String? = null,
) {
    @Serializable
    data class Item(
        val productoId: String? = null,
        val nombre: String,
        val foto: String? = null,
        val cantidad: Int = 1,
        val precio: Int = 0,
        val varianteId: String? = null,
        val variante: String? = null,
    ) {
        val id: String get() = (productoId ?: nombre) + (varianteId ?: "")
    }

    @Serializable data class Descuento(val concepto: String, val monto: Int)

    @Serializable
    data class Comprador(
        val nombre: String = "",
        val telefono: String? = null,
        val email: String? = null,
        val desde: FechaApi = Instant.EPOCH,
        val pedidos: Int = 0,
        val gastado: Int = 0,
    )

    @Serializable
    data class Entrega(val direccion: String = "", val cp: String = "", val provincia: String? = null, val localidad: String? = null, val sucursalNombre: String? = null)

    @Serializable data class Paso(val texto: String, val cuando: FechaApi? = null, val hecho: Boolean = false)
    @Serializable data class Sena(val monto: Int? = null, val pagada: Boolean? = null)
}

// MARK: - Productos

/** Una fila de la lista de productos (/v1/panel/listas/productos). */
@Serializable
data class ProductoFila(
    val id: String,
    val nombre: String = "",
    val foto: String? = null,
    val sku: String? = null,
    val precio: Int = 0,
    val costo: Int? = null,
    val stock: Int = 0,
    val canales: List<String> = emptyList(),
    val publicado: Boolean = false,
    val estado: String = "",
    val vendidos30: Int? = null,
    val variantes: List<Variante> = emptyList(),
) {
    @Serializable
    data class Variante(val id: String, val texto: String = "", val sku: String? = null, val precio: Int? = null, val stock: Int = 0, val activa: Boolean = true)
}

@Serializable
data class PaginaProductos(val total: Int = 0, val filas: List<ProductoFila> = emptyList(), val segmentos: Map<String, Int> = emptyMap())

// MARK: - Avisos

@Serializable
data class Notificacion(
    val id: String,
    val tipo: String = "",
    val titulo: String = "",
    val detalle: String = "",
    val creada: FechaApi = Instant.EPOCH,
    val leida: Boolean = false,
    val ruta: String? = null,
)

@Serializable
data class PaginaNotificaciones(val notificaciones: List<Notificacion> = emptyList(), val sinLeer: Int = 0, val siguiente: String? = null)

/** Las ilustraciones de Tincho (Assets/Tincho). */
enum class Tincho(val imagen: Int) {
    avatar(R.drawable.tincho_avatar), saludo(R.drawable.tincho_saludo), hablando(R.drawable.tincho_hablando),
    pensando(R.drawable.tincho_pensando), preocupado(R.drawable.tincho_preocupado), celebrando(R.drawable.tincho_celebrando),
    desenchufado(R.drawable.tincho_desenchufado), foto(R.drawable.tincho_foto), seguridad(R.drawable.tincho_seguridad),
    grafico(R.drawable.tincho_grafico), caja(R.drawable.tincho_caja), envio(R.drawable.tincho_envio),
    asomado(R.drawable.tincho_asomado), durmiendo(R.drawable.tincho_durmiendo), exito(R.drawable.tincho_exito),
    lupa(R.drawable.tincho_lupa), portapapeles(R.drawable.tincho_portapapeles), ticket(R.drawable.tincho_ticket),
    whatsapp(R.drawable.tincho_whatsapp),
}
