import Foundation
import Security
import SwiftUI

// Lo que comparten la app y la extensión de widgets: la sesión (en el llavero compartido), los datos
// del día, los colores y la Live Activity de los pedidos por preparar.

/// La sesión guardada por la app, leída desde el grupo de llavero compartido (TLWGrupoLlavero).
enum SesionCompartida {
    static let servicio = "ar.com.tulocalweb.app.sesion"
    static var grupo: String? { Bundle.main.object(forInfoDictionaryKey: "TLWGrupoLlavero") as? String }

    static func token() -> String? {
        var q: [String: Any] = [kSecClass as String: kSecClassGenericPassword,
                                kSecAttrService as String: servicio,
                                kSecAttrAccount as String: "token",
                                kSecReturnData as String: true,
                                kSecMatchLimit as String: kSecMatchLimitOne]
        if let grupo { q[kSecAttrAccessGroup as String] = grupo }
        var r: CFTypeRef?
        guard SecItemCopyMatching(q as CFDictionary, &r) == errSecSuccess, let d = r as? Data else { return nil }
        return String(data: d, encoding: .utf8)
    }

    static var api: URL {
        (Bundle.main.object(forInfoDictionaryKey: "TLWApi") as? String).flatMap(URL.init(string:))
            ?? URL(string: "https://api.tulocalweb.com.ar")!
    }

    /// GET a la API con la sesión compartida.
    static func get(_ ruta: String, query: [URLQueryItem] = []) async throws -> Data {
        var c = URLComponents(url: api.appendingPathComponent(ruta), resolvingAgainstBaseURL: false)!
        if !query.isEmpty { c.queryItems = query }
        var r = URLRequest(url: c.url!, timeoutInterval: 15)
        r.setValue("application/json", forHTTPHeaderField: "Accept")
        if let t = token() { r.setValue("Bearer \(t)", forHTTPHeaderField: "Authorization") }
        let (d, resp) = try await URLSession.shared.data(for: r)
        guard let h = resp as? HTTPURLResponse, (200..<300).contains(h.statusCode) else { throw URLError(.badServerResponse) }
        return d
    }

    static func post(_ ruta: String, json: [String: String]) async throws {
        var r = URLRequest(url: api.appendingPathComponent(ruta), timeoutInterval: 15)
        r.httpMethod = "POST"
        r.setValue("application/json", forHTTPHeaderField: "Content-Type")
        if let t = token() { r.setValue("Bearer \(t)", forHTTPHeaderField: "Authorization") }
        r.httpBody = try JSONEncoder().encode(json)
        let (_, resp) = try await URLSession.shared.data(for: r)
        guard let h = resp as? HTTPURLResponse, (200..<300).contains(h.statusCode) else { throw URLError(.badServerResponse) }
    }
}

/// Cómo viene el día, para los widgets.
struct DatosHoy: Codable, Hashable {
    struct Pedido: Codable, Hashable { var numero: Int; var cliente: String; var total: Int?; var estado: String }
    var conSesion: Bool
    var montos: Bool
    var ventasHoy: Double
    var ventasAyer: Double
    var serie: [Double]
    var pedidosHoy: Int
    var porPreparar: Int
    var ultimo: Pedido?
    /// El más nuevo de los que esperan preparación (para la Live Activity).
    var ultimoPorPreparar: Pedido?

    var delta: Double? { ventasAyer > 0 ? (ventasHoy - ventasAyer) / ventasAyer * 100 : nil }

    static let ejemplo = DatosHoy(conSesion: true, montos: true, ventasHoy: 184_350, ventasAyer: 164_000,
                                  serie: [0, 8, 15, 22, 40, 52, 61, 80, 95, 110, 128, 150, 170, 184],
                                  pedidosHoy: 12, porPreparar: 3,
                                  ultimo: Pedido(numero: 1048, cliente: "Lucía Fernández", total: 23_400, estado: "Para preparar"),
                                  ultimoPorPreparar: Pedido(numero: 1048, cliente: "Lucía Fernández", total: 23_400, estado: "Para preparar"))
    static let sinSesion = DatosHoy(conSesion: false, montos: false, ventasHoy: 0, ventasAyer: 0, serie: [], pedidosHoy: 0, porPreparar: 0, ultimo: nil, ultimoPorPreparar: nil)

    static func traer() async -> DatosHoy {
        guard SesionCompartida.token() != nil else { return .sinSesion }
        struct Comparado: Decodable { var actual: Double; var anterior: Double }
        struct Serie: Decodable { var actual: [Double] }
        struct Kpis: Decodable { var pedidos: Comparado }
        struct Breve: Decodable { var numero: Int; var cliente: String; var total: Int?; var estado: String }
        struct Tablero: Decodable { var montos: Bool; var ventas: Comparado?; var serie: Serie?; var kpis: Kpis; var ultimosPedidos: [Breve] }
        struct Insignias: Decodable { var pedidos: Int }
        struct Resumen: Decodable { var insignias: Insignias }
        async let t = try? JSONDecoder().decode(Tablero.self, from: await SesionCompartida.get("/v1/panel/tablero", query: [URLQueryItem(name: "periodo", value: "hoy")]))
        async let r = try? JSONDecoder().decode(Resumen.self, from: await SesionCompartida.get("/v1/panel/resumen"))
        guard let tab = await t else { return .sinSesion }
        let res = await r
        let u = tab.ultimosPedidos.first
        let pp = tab.ultimosPedidos.first { $0.estado == "Para preparar" }
        return DatosHoy(conSesion: true, montos: tab.montos, ventasHoy: tab.ventas?.actual ?? 0, ventasAyer: tab.ventas?.anterior ?? 0,
                        serie: tab.serie?.actual ?? [], pedidosHoy: Int(tab.kpis.pedidos.actual),
                        porPreparar: res?.insignias.pedidos ?? 0,
                        ultimo: u.map { Pedido(numero: $0.numero, cliente: $0.cliente, total: $0.total, estado: $0.estado) },
                        ultimoPorPreparar: pp.map { Pedido(numero: $0.numero, cliente: $0.cliente, total: $0.total, estado: $0.estado) })
    }
}

/// Colores del handoff para los widgets (sin las fuentes de la app).
enum ColorWidget {
    static let fondo = Color(light: 0xFBF7F2, dark: 0x22140D)
    static let texto = Color(light: 0x2A140C, dark: 0xFBF7F2)
    static let texto2 = Color(light: 0x6B5244, dark: 0xCDB8A8)
    static let acento = Color(light: 0xD9652F, dark: 0xE07A48)
    static let primario = Color(light: 0xB0441D, dark: 0xE07A48)
    static let ok = Color(light: 0x1F6B48, dark: 0x6FD0A0)
    static let error = Color(light: 0xA23F1C, dark: 0xF08A63)
}

extension Color {
    init(light: UInt32, dark: UInt32) {
        func comp(_ v: UInt32) -> (Double, Double, Double) {
            (Double((v >> 16) & 0xFF) / 255, Double((v >> 8) & 0xFF) / 255, Double(v & 0xFF) / 255)
        }
        let l = comp(light), d = comp(dark)
        self.init(uiColor: UIColor { t in
            let c = t.userInterfaceStyle == .dark ? d : l
            return UIColor(red: c.0, green: c.1, blue: c.2, alpha: 1)
        })
    }
}

enum PesosWidget {
    static func texto(_ n: Double) -> String {
        let f = NumberFormatter()
        f.numberStyle = .decimal
        f.locale = Locale(identifier: "es_AR")
        f.maximumFractionDigits = 0
        return "$\u{202F}" + (f.string(from: NSNumber(value: n.rounded())) ?? "\(Int(n))")
    }
}
