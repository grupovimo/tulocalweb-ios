import Foundation

/// Los links de tulocalweb.com.ar que abren la app (Universal Links), y las rutas del panel que mandan los avisos.
/// Qué paths abren la app lo declara `apps/pruebas/enlaces.json` (de ahí sale el AASA de la web); acá, a qué
/// pantalla lleva cada uno. Lo que la app no tiene se abre en la web dentro de la app, nunca queda en nada.
enum Enlaces {
    static func ruta(_ u: URL) -> Ruta? {
        guard let host = u.host(), host == "tulocalweb.com.ar" || host == "www.tulocalweb.com.ar" else { return nil }
        let query = URLComponents(url: u, resolvingAgainstBaseURL: false)?.queryItems ?? []
        func valor(_ k: String) -> String? { query.first { $0.name == k }.map { $0.value ?? "" } }
        if valor("web") != nil { return .web(u) }
        let p = u.pathComponents.filter { $0 != "/" }
        if p.first == "ayuda" { return .destino(.ayuda) }
        guard p.first == "panel" else { return .web(u) }
        let seccion = p.count > 1 ? p[1] : nil
        let id = p.count > 2 ? p[2] : nil
        switch seccion {
        case nil, "para-hacer-hoy": return .inicio
        case "pedidos":
            if let n = valor("pedido").flatMap(Int.init) ?? id.flatMap(Int.init) { return .pedido(n) }
            if id == "nuevo" { return .destino(.pedidoManual) }
            return id == nil ? .pedidos : .web(u)
        case "productos":
            if let e = valor("editar") ?? id { return .producto(e) }
            return .productos
        case "configuracion": return valor("tab") == "seguridad" ? .destino(.cuenta) : .destino(.negocio)
        default:
            return secciones.first { $0.0 == seccion }.map { .destino($0.1) } ?? .web(u)
        }
    }

    /// Las secciones del panel que tienen pantalla en la app (/panel/<sección>).
    private static let secciones: [(String, Destino)] = [
        ("inventario", .inventario), ("ventas-en-local", .caja), ("clientes", .clientes),
        ("carritos-abandonados", .carritos), ("resenas-y-preguntas", .resenas), ("devoluciones", .devoluciones),
        ("cupones", .cupones), ("finanzas", .finanzas), ("facturacion", .facturacion), ("plan-y-facturacion", .plan),
        ("integraciones", .integraciones), ("usuarios-y-permisos", .usuarios), ("mis-datos", .cuenta),
        ("notificaciones", .preferenciasAvisos), ("informes", .estadisticas), ("mi-tienda", .miTienda),
        ("links", .links), ("importar", .importar),
    ]
}

extension Ruta {
    /// Cómo se escribe la ruta en `apps/pruebas/enlaces.json` («pedido:1048», «destino:cupones», «web-en-app»).
    var clave: String {
        switch self {
        case .pedido(let n): "pedido:\(n)"
        case .producto(let id): "producto:\(id)"
        case .avisos: "avisos"
        case .pedidos: "pedidos"
        case .inicio: "inicio"
        case .compartir: "compartir"
        case .productos: "productos"
        case .web: "web-en-app"
        case .destino(let d): "destino:\(String(describing: d))"
        }
    }
}
