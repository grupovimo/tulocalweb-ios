import Foundation

/// Un JSON cualquiera: para editar un producto sin perder los campos que la app no muestra (opciones,
/// variantes, canales) al mandarlo de vuelta.
enum ValorJSON: Codable, Equatable {
    case texto(String), numero(Double), booleano(Bool), lista([ValorJSON]), objeto([String: ValorJSON]), nulo

    init(from decoder: Decoder) throws {
        let c = try decoder.singleValueContainer()
        if c.decodeNil() { self = .nulo }
        else if let b = try? c.decode(Bool.self) { self = .booleano(b) }
        else if let n = try? c.decode(Double.self) { self = .numero(n) }
        else if let s = try? c.decode(String.self) { self = .texto(s) }
        else if let l = try? c.decode([ValorJSON].self) { self = .lista(l) }
        else { self = .objeto(try c.decode([String: ValorJSON].self)) }
    }

    func encode(to encoder: Encoder) throws {
        var c = encoder.singleValueContainer()
        switch self {
        case let .texto(s): try c.encode(s)
        case let .numero(n):
            if n == n.rounded(), abs(n) < 1e15 { try c.encode(Int(n)) } else { try c.encode(n) }
        case let .booleano(b): try c.encode(b)
        case let .lista(l): try c.encode(l)
        case let .objeto(o): try c.encode(o)
        case .nulo: try c.encodeNil()
        }
    }

    var texto: String? { if case let .texto(s) = self { return s }; return nil }
    var entero: Int? { if case let .numero(n) = self { return Int(n) }; return nil }
    var bool: Bool? { if case let .booleano(b) = self { return b }; return nil }
    var lista: [ValorJSON]? { if case let .lista(l) = self { return l }; return nil }
    subscript(_ k: String) -> ValorJSON? { if case let .objeto(o) = self { return o[k] }; return nil }
}

/// Lo que la app usa de la web (no de la API): subir fotos y hablar con Tincho. La web acepta el
/// mismo token de la API en esas dos rutas (hooks.server.ts, RUTAS_APP).
enum Web {
    static let base = URL(string: "https://tulocalweb.com.ar")!

    /// Sube una foto (JPEG) y devuelve su URL pública (https://tulocalweb.com.ar/f/…).
    static func subirFoto(_ datos: Data, token: String?) async throws -> String {
        var r = URLRequest(url: base.appending(path: "panel/fotos"))
        r.httpMethod = "POST"
        r.setValue("image/jpeg", forHTTPHeaderField: "Content-Type")
        if let token { r.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization") }
        r.httpBody = datos
        let (d, resp) = try await URLSession.shared.data(for: r)
        guard let h = resp as? HTTPURLResponse, (200..<300).contains(h.statusCode) else {
            throw ErrorApi.desconocido
        }
        struct R: Decodable { var ruta: String }
        let ruta = try JSONDecoder().decode(R.self, from: d).ruta
        return ruta.hasPrefix("http") ? ruta : base.absoluteString + ruta
    }

    /// Una pregunta a Tincho: va devolviendo el texto a medida que llega.
    /// Tincho con bloques (cifras, gráficos, tablas, botones y sugerencias): la respuesta entera en JSON.
    static func tinchoConBloques(_ mensajes: [(rol: String, texto: String)], token: String?) async throws -> RespuestaTincho {
        var r = URLRequest(url: base.appending(path: "panel/hornero"), timeoutInterval: 40)
        r.httpMethod = "POST"
        r.setValue("application/json", forHTTPHeaderField: "Content-Type")
        if let token { r.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization") }
        struct M: Encodable { var rol: String; var texto: String }
        struct C: Encodable { var mensajes: [M]; var formato = "bloques" }
        r.httpBody = try JSONEncoder().encode(C(mensajes: mensajes.suffix(12).map { M(rol: $0.rol, texto: String($0.texto.prefix(1000))) }))
        let (d, resp) = try await URLSession.shared.data(for: r)
        guard let h = resp as? HTTPURLResponse, (200..<300).contains(h.statusCode) else { throw ErrorApi.desconocido }
        return try JSONDecoder().decode(RespuestaTincho.self, from: d)
    }

    static func tincho(_ mensajes: [(rol: String, texto: String)], token: String?) -> AsyncThrowingStream<String, Error> {
        AsyncThrowingStream { cont in
            let tarea = Task {
                do {
                    var r = URLRequest(url: base.appending(path: "panel/hornero"))
                    r.httpMethod = "POST"
                    r.setValue("application/json", forHTTPHeaderField: "Content-Type")
                    if let token { r.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization") }
                    struct M: Encodable { var rol: String; var texto: String }
                    struct C: Encodable { var mensajes: [M] }
                    r.httpBody = try JSONEncoder().encode(C(mensajes: mensajes.suffix(12).map { M(rol: $0.rol, texto: $0.texto) }))
                    let (bytes, resp) = try await URLSession.shared.bytes(for: r)
                    guard let h = resp as? HTTPURLResponse, (200..<300).contains(h.statusCode) else {
                        throw ErrorApi.desconocido
                    }
                    var acumulado = Data()
                    for try await b in bytes {
                        acumulado.append(b)
                        // Se manda cuando el pedazo es UTF-8 válido (no se corta una letra a la mitad).
                        if let s = String(data: acumulado, encoding: .utf8) {
                            cont.yield(s)
                            acumulado.removeAll()
                        }
                    }
                    cont.finish()
                } catch {
                    cont.finish(throwing: error)
                }
            }
            cont.onTermination = { _ in tarea.cancel() }
        }
    }
}
