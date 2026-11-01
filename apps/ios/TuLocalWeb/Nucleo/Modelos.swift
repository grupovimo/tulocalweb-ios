import Foundation

// Lo que la app lee de la API (packages/contrato/openapi.yaml). Sólo los campos que usa; lo demás se
// ignora. Los montos vienen en pesos enteros.

// MARK: - Sesión

struct SesionNueva: Codable {
    var token: String
    var venceEn: Date
    var tiendaId: String
    /// Sólo en el ingreso con Clerk: la cuenta se creó recién (sigue al onboarding, en la web).
    var nueva: Bool?
}

struct DatosSesion: Codable, Equatable {
    struct Usuario: Codable, Equatable {
        var id: String
        var nombre: String
        var email: String
        var foto: String?
    }
    var usuario: Usuario
    var tiendaId: String
    var rol: String
    var venceEn: Date

    /// El nombre de pila, para «Hola, Sofía».
    var nombreDePila: String {
        String(usuario.nombre.split(separator: " ").first ?? Substring(usuario.nombre))
    }
}

struct Acceso: Codable, Equatable {
    struct Tienda: Codable, Equatable, Identifiable {
        var id: String
        var nombre: String
        var slug: String
        var rol: String

        /// «almacen-lumbre.tulocalweb.com.ar».
        var direccion: String { "\(slug).tulocalweb.com.ar" }
    }
    var tiendaId: String
    var rol: String
    var permisos: [String]
    var tiendas: [Tienda]

    var tienda: Tienda? { tiendas.first { $0.id == tiendaId } }
    func puede(_ p: String) -> Bool { permisos.contains(p) }
}

/// «dueña», «administradora»… como lo dice el handoff (1d).
func nombreDelRol(_ rol: String) -> String {
    switch rol {
    case "duenio": return "dueña"
    case "admin": return "administración"
    case "ventas": return "ventas"
    case "local": return "local"
    default: return rol
    }
}

struct Proveedores: Codable {
    var clerk: Bool
    var clerkPublicable: String?
}

// MARK: - Inicio (tablero)

struct Comparado: Codable, Equatable {
    var actual: Double
    var anterior: Double

    /// Variación en %, o nil si antes no había nada con qué comparar.
    var variacion: Double? {
        anterior > 0 ? (actual - anterior) / anterior * 100 : nil
    }
}

struct Tablero: Codable {
    struct Serie: Codable { var actual: [Double]; var anterior: [Double] }
    struct Kpis: Codable {
        var pedidos: Comparado
        var ticket: Comparado?
        var clientesNuevos: Comparado
    }
    struct Canal: Codable { var canal: String; var parte: Double }
    struct Meta: Codable { var objetivo: Int; var actual: Int; var proyeccion: Int }
    struct ParaHacer: Codable {
        var preparar: Int
        var transferencias: Int
        var transferenciasMonto: Int?
        var sinStock: Int
        var sinStockEnMl: Int
    }
    struct Hornero: Codable {
        var carritos: Int
        var carritosMonto: Int?
        var ideasHechas: [String]
    }
    struct MasVendido: Codable, Identifiable {
        var productoId: String
        var nombre: String
        var foto: String?
        var unidades: Int
        var id: String { productoId }
    }
    var periodo: String
    var desde: Date
    var hasta: Date
    var hola: String?
    var montos: Bool
    var ventas: Comparado?
    var serie: Serie?
    var kpis: Kpis
    var canales: [Canal]?
    var meta: Meta?
    var ultimosPedidos: [PedidoBreve]
    var paraHacer: ParaHacer
    var hornero: Hornero
    var masVendidos: [MasVendido]
}

struct PedidoBreve: Codable, Identifiable {
    var numero: Int
    var cliente: String
    var foto: String?
    var canal: String
    var estado: String
    var total: Int?
    var creado: Date
    var id: Int { numero }
}

struct ResumenPanel: Codable {
    struct Insignias: Codable { var pedidos: Int; var carritos: Int }
    struct Prueba: Codable { var plan: String; var diasRestantes: Int; var diasTotales: Int }
    var insignias: Insignias
    var prueba: Prueba
    /// Canales conectados (wa, ml, mp, ig, tg).
    var canales: [String]?
}

// MARK: - Pedidos

struct PedidoFila: Codable, Identifiable, Equatable {
    struct Producto: Codable, Equatable { var nombre: String; var foto: String? }
    var numero: Int
    var cliente: String
    var foto: String?
    var canal: String
    var pago: String
    var envio: String
    var estado: String
    var total: Int
    var creado: Date
    var productos: [Producto]
    var cantidadProductos: Int
    /// El WhatsApp de quien compró, tal como lo cargó.
    var telefono: String?
    var id: Int { numero }

    /// La foto del primer producto, para el cuadrito de la fila.
    var fotoProducto: String? { productos.first?.foto }
}

struct PaginaPedidos: Codable {
    var total: Int
    var suma: Int
    var filas: [PedidoFila]
    var etapas: [String: Int]
}

struct PedidoDetalle: Codable {
    struct Item: Codable, Identifiable {
        var productoId: String?
        var nombre: String
        var foto: String?
        var cantidad: Int
        var precio: Int
        var varianteId: String?
        var variante: String?
        var id: String { (productoId ?? nombre) + (varianteId ?? "") }
    }
    struct Descuento: Codable { var concepto: String; var monto: Int }
    struct Comprador: Codable {
        var nombre: String
        var telefono: String?
        var email: String?
        var desde: Date
        var pedidos: Int
        var gastado: Int
    }
    struct Entrega: Codable {
        var direccion: String
        var cp: String
        var provincia: String?
        var localidad: String?
        var sucursalNombre: String?
    }
    struct Paso: Codable { var texto: String; var cuando: Date?; var hecho: Bool }
    struct Sena: Codable { var monto: Int?; var pagada: Bool? }

    var numero: Int
    var cliente: String
    var canal: String
    var pago: String
    var envio: String
    var estado: String
    var total: Int
    var creado: Date
    var cantidadProductos: Int
    var items: [Item]
    var subtotal: Int
    var costoEnvio: Int
    var descuentos: [Descuento]
    var comprador: Comprador
    var estadoPago: String
    var factura: String?
    var seguimiento: String?
    var llega: Date?
    var nota: String?
    var entrega: Entrega?
    var historial: [Paso]
    /// Código público de la confirmación (/t/<tienda>/pedido/<codigo>); null si se cargó a mano.
    var codigo: String?
}

// MARK: - Productos

/// Una fila de la lista de productos (/v1/panel/listas/productos).
struct ProductoFila: Codable, Identifiable, Equatable {
    struct Variante: Codable, Equatable, Identifiable {
        var id: String
        var texto: String
        var sku: String?
        var precio: Int?
        var stock: Int
        var activa: Bool
    }
    var id: String
    var nombre: String
    var foto: String?
    var sku: String?
    var precio: Int
    var costo: Int?
    var stock: Int
    var canales: [String]
    var publicado: Bool
    var estado: String
    var vendidos30: Int?
    var variantes: [Variante]
}

struct PaginaProductos: Codable {
    var total: Int
    var filas: [ProductoFila]
    var segmentos: [String: Int]
}

// MARK: - Avisos

struct Notificacion: Codable, Identifiable, Equatable {
    var id: String
    var tipo: String
    var titulo: String
    var detalle: String
    var creada: Date
    var leida: Bool
    var ruta: String?
}

struct PaginaNotificaciones: Codable {
    var notificaciones: [Notificacion]
    var sinLeer: Int
    var siguiente: String?
}

/// Las ilustraciones de Tincho (Assets/Tincho).
enum Tincho: String {
    case avatar, saludo, hablando, pensando, preocupado, celebrando, desenchufado, foto, seguridad, grafico
    case caja, envio, asomado, durmiendo, exito, lupa, portapapeles, ticket, whatsapp

    var imagen: String { "Tincho/\(rawValue)" }
}
