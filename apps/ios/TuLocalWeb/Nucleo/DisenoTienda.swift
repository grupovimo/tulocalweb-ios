import Foundation

// El diseño v2 de la tienda online (docs/especificaciones/tienda-bloques.md): las páginas armadas con
// bloques. Lee y escribe exactamente el JSON de la especificación; lo que la app no edita viaja tal cual.

/// Un bloque de una página: el tipo (bloques.json), su variante y sus datos (lo que falta toma porDefecto).
struct Bloque: Codable, Equatable, Identifiable {
    var id: String
    var tipo: String
    var variante: String
    var oculto = false
    var datos: [String: ValorJSON] = [:]

    init(id: String, tipo: String, variante: String, oculto: Bool = false, datos: [String: ValorJSON] = [:]) {
        self.id = id
        self.tipo = tipo
        self.variante = variante
        self.oculto = oculto
        self.datos = datos
    }

    private enum Claves: String, CodingKey { case id, tipo, variante, oculto, datos }

    init(from decoder: Decoder) throws {
        let c = try decoder.container(keyedBy: Claves.self)
        id = try c.decode(String.self, forKey: .id)
        tipo = try c.decode(String.self, forKey: .tipo)
        variante = try c.decodeIfPresent(String.self, forKey: .variante) ?? ""
        oculto = try c.decodeIfPresent(Bool.self, forKey: .oculto) ?? false
        datos = try c.decodeIfPresent([String: ValorJSON].self, forKey: .datos) ?? [:]
    }

    func encode(to encoder: Encoder) throws {
        var c = encoder.container(keyedBy: Claves.self)
        try c.encode(id, forKey: .id)
        try c.encode(tipo, forKey: .tipo)
        try c.encode(variante, forKey: .variante)
        if oculto { try c.encode(true, forKey: .oculto) }
        try c.encode(datos, forKey: .datos)
    }

    /// Un id nuevo: [a-z0-9]{8}, como pide la especificación (6 a 16).
    static func nuevoId() -> String {
        let letras = Array("abcdefghijklmnopqrstuvwxyz0123456789")
        return String((0..<8).map { _ in letras.randomElement()! })
    }

    /// Un bloque recién agregado: la primera variante y los datos vacíos (la tienda usa lo de por defecto).
    static func nuevo(_ tipo: CatalogoBloques.Tipo) -> Bloque {
        Bloque(id: nuevoId(), tipo: tipo.tipo, variante: tipo.variantes.first ?? "")
    }
}

/// Los ajustes de una página que no es el inicio (paginas.catalogo, producto, carrito, nosotros): su variante,
/// los campos de bloques.json › paginas y, en Ficha y Nosotros, sus bloques.
struct AjustesPagina: Equatable {
    var variante: String?
    var bloques: [Bloque]?
    var campos: [String: ValorJSON] = [:]

    init() {}

    init(json: ValorJSON?) {
        guard case let .objeto(o)? = json else { return }
        variante = o["variante"]?.texto
        bloques = o["bloques"].flatMap { $0.como([Bloque].self) }
        campos = o.filter { $0.key != "variante" && $0.key != "bloques" }
    }

    var json: ValorJSON {
        var o = campos
        if let variante { o["variante"] = .texto(variante) }
        if let bloques { o["bloques"] = ValorJSON.de(bloques) }
        return .objeto(o)
    }
}

struct Diseno: Equatable {
    var plantilla: String?
    var paleta = 0
    var titulo = ""
    var lema = ""
    var portada = ""
    var logo = ""
    var compartirTitulo = ""
    var compartirDescripcion = ""
    var estilo = EstiloTienda()
    var inicio: [Bloque] = []
    var paginas: [String: AjustesPagina] = [:]
    /// Las ventanas (popups), en orden: si varias califican, sale la primera.
    var ventanas: [Ventana] = []
    /// Lo que la app no edita (compartirFoto y lo que venga): vuelve igual en el PUT.
    var resto: [String: ValorJSON] = [:]

    private static let propias: Set<String> = [
        "version", "plantilla", "paleta", "titulo", "lema", "portada", "logo", "compartirTitulo", "compartirDescripcion",
        "estilo", "inicio", "paginas", "secciones", "ventanas"
    ]

    init() {}

    /// Desde el JSON del GET. La API lo manda ya en v2; si llega uno viejo, se migra acá igual que en la web.
    init(json o: [String: ValorJSON], plantillaTienda: String) {
        resto = o.filter { !Diseno.propias.contains($0.key) }
        plantilla = o["plantilla"]?.texto
        paleta = o["paleta"]?.entero ?? 0
        titulo = o["titulo"]?.texto ?? ""
        lema = o["lema"]?.texto ?? ""
        portada = o["portada"]?.texto ?? ""
        logo = o["logo"]?.texto ?? ""
        compartirTitulo = o["compartirTitulo"]?.texto ?? ""
        compartirDescripcion = o["compartirDescripcion"]?.texto ?? ""
        estilo = EstiloTienda.desde(o["estilo"])
        ventanas = o["ventanas"]?.lista?.map { Ventana(json: $0) } ?? []
        if o["version"]?.entero == 2 {
            inicio = o["inicio"]?.como([Bloque].self) ?? []
            if case let .objeto(p)? = o["paginas"] {
                for (k, v) in p { paginas[k] = AjustesPagina(json: v) }
            }
        } else {
            migrar(o, plantilla: plantilla ?? plantillaTienda)
        }
    }

    /// Para el PUT: version 2, sin lo vacío y sin nada de v1.
    var json: [String: ValorJSON] {
        var o = resto
        o["version"] = .numero(2)
        if let plantilla, !plantilla.isEmpty { o["plantilla"] = .texto(plantilla) }
        o["paleta"] = .numero(Double(paleta))
        let textos = [("titulo", titulo), ("lema", lema), ("portada", portada), ("logo", logo),
                      ("compartirTitulo", compartirTitulo), ("compartirDescripcion", compartirDescripcion)]
        for (k, v) in textos where !v.trimmingCharacters(in: .whitespaces).isEmpty { o[k] = .texto(v) }
        let e = estilo.json
        if case let .objeto(eo) = e, !eo.isEmpty { o["estilo"] = e }
        o["inicio"] = ValorJSON.de(inicio)
        if !paginas.isEmpty { o["paginas"] = .objeto(paginas.mapValues(\.json)) }
        if !ventanas.isEmpty { o["ventanas"] = .lista(ventanas.map(\.json)) }
        return o
    }

    /// El JSON en texto, para la vista previa (evaluateJavaScript).
    var textoJSON: String {
        guard let d = try? JSONEncoder().encode(ValorJSON.objeto(json)), let s = String(data: d, encoding: .utf8) else { return "{}" }
        // U+2028 y U+2029 valen en JSON pero cortan un literal de JavaScript.
        return s.replacingOccurrences(of: "\u{2028}", with: "\\u2028").replacingOccurrences(of: "\u{2029}", with: "\\u2029")
    }

    // MARK: Bloques por página

    func bloques(_ pagina: String) -> [Bloque] {
        pagina == "inicio" ? inicio : (paginas[pagina]?.bloques ?? [])
    }

    mutating func ponerBloques(_ b: [Bloque], en pagina: String) {
        if pagina == "inicio" { inicio = b; return }
        var a = paginas[pagina] ?? CatalogoBloques.shared.ajustesNuevos(pagina)
        a.bloques = b
        paginas[pagina] = a
    }

    func ajustes(_ pagina: String) -> AjustesPagina {
        paginas[pagina] ?? CatalogoBloques.shared.ajustesNuevos(pagina)
    }

    var idsUsados: Set<String> {
        Set(CatalogoBloques.ordenPaginas.flatMap { bloques($0).map(\.id) })
    }

    /// Busca un bloque en cualquier página (lo que toca la persona en la vista previa).
    func buscar(_ id: String) -> (pagina: String, bloque: Bloque)? {
        for p in CatalogoBloques.ordenPaginas { if let b = bloques(p).first(where: { $0.id == id }) { return (p, b) } }
        return nil
    }

    // MARK: Punto de partida

    /// Cambia de punto de partida: la dirección, el estilo base y el orden del preset, conservando los datos de
    /// los bloques del mismo tipo (y los que el preset no trae, al final). Se deshace con el historial.
    mutating func aplicarPreset(_ id: String) {
        guard let p = PresetTienda.todos[id] else { return }
        plantilla = id
        paleta = 0
        // Lo que define el preset vuelve a ser el del preset (el diseño guarda sólo lo que cambió la tienda).
        estilo.direccion = nil
        estilo.forma = nil
        estilo.densidad = nil
        estilo.ancho = nil
        estilo.fotoProducto = nil
        estilo.encabezado = nil
        estilo.pie = nil
        estilo.fuenteTitulos = nil
        estilo.fuenteTexto = nil
        var viejos = inicio
        var usados = Set(CatalogoBloques.ordenPaginas.filter { $0 != "inicio" }.flatMap { bloques($0).map(\.id) })
        var nuevos: [Bloque] = []
        for b in p.inicio {
            if let i = viejos.firstIndex(where: { $0.tipo == b.tipo }) {
                var v = viejos.remove(at: i)
                v.variante = b.variante
                nuevos.append(v)
            } else {
                var n = b
                if usados.contains(n.id) || viejos.contains(where: { $0.id == n.id }) { n.id = Bloque.nuevoId() }
                nuevos.append(n)
            }
            usados.insert(nuevos[nuevos.count - 1].id)
        }
        inicio = nuevos + viejos
    }

    // MARK: Migración v1 → v2 (por si la API manda uno viejo; los mismos casos que casos/diseno-v2.json)

    private mutating func migrar(_ o: [String: ValorJSON], plantilla id: String) {
        let p = PresetTienda.todos[id] ?? PresetTienda.todos["litoral"]
        let deSeccion = ["hero": "portada", "cats": "categorias", "prods": "productos", "revs": "resenas", "ig": "galeria", "wa": "whatsapp"]
        let secciones: [(tipo: String, visible: Bool)] = o["secciones"]?.lista?.compactMap { s in
            guard let sid = s["id"]?.texto, let t = deSeccion[sid] else { return nil }
            return (t, s["visible"]?.bool ?? true)
        } ?? []
        var textos: [String: String] = [:]
        if case let .objeto(t)? = o["estilo"]?["textos"] {
            for (k, v) in t { if let tipo = deSeccion[k], let s = v.texto, !s.isEmpty { textos[tipo] = s } }
        }
        let orden = p?.orden ?? []
        // El orden de la tienda si lo cambió (filtrado a los del preset); si no, el del preset.
        let deFabrica = ["portada", "categorias", "productos", "resenas", "galeria", "whatsapp"]
        let tocado = secciones.map(\.tipo) != deFabrica.filter { t in secciones.contains { $0.tipo == t } }
        let propio = secciones.map(\.tipo).filter { orden.contains($0) }
        let final = tocado && propio.count == orden.count ? propio : orden
        inicio = final.map { t in
            var datos: [String: ValorJSON] = [:]
            if t == "portada" {
                if !titulo.isEmpty { datos["titulo"] = .texto(titulo) }
                if !lema.isEmpty { datos["texto"] = .texto(lema) }
                if !portada.isEmpty { datos["fotos"] = .lista([.texto(portada)]) }
            }
            if let s = textos[t] { datos["titulo"] = .texto(s) }
            let visible = secciones.first { $0.tipo == t }?.visible ?? true
            return Bloque(id: t, tipo: t, variante: p?.variantes[t] ?? CatalogoBloques.shared.tipo(t)?.variantes.first ?? "",
                          oculto: !visible, datos: datos)
        }
        if secciones.contains(where: { $0.tipo == "whatsapp" && !$0.visible }) { estilo.whatsappFlotante = false }
        if let a = estilo.anuncio, !a.trimmingCharacters(in: .whitespaces).isEmpty {
            let link = estilo.anuncioLink.flatMap { $0.isEmpty ? nil : $0 }
            estilo.barra = .init(variante: "fija", mensajes: [.init(texto: a, link: link)])
        }
        estilo.anuncio = nil
        estilo.anuncioLink = nil
        estilo.textos = nil
    }
}

/// Deshacer y rehacer del editor: fotos del diseño entero (es chico, hasta 64 KB).
struct Historial {
    private(set) var pasado: [Diseno] = []
    private(set) var futuro: [Diseno] = []

    var puedeDeshacer: Bool { !pasado.isEmpty }
    var puedeRehacer: Bool { !futuro.isEmpty }

    /// Antes de cambiar algo: guarda cómo estaba.
    mutating func guardar(_ d: Diseno) {
        if pasado.last != d { pasado.append(d) }
        if pasado.count > 60 { pasado.removeFirst() }
        futuro = []
    }

    /// Si al final no cambió nada (se abrió un panel y se cerró), no deja un paso vacío.
    mutating func descartarSiIgual(_ d: Diseno) {
        if pasado.last == d { pasado.removeLast() }
    }

    mutating func deshacer(_ actual: Diseno) -> Diseno? {
        guard let d = pasado.popLast() else { return nil }
        futuro.append(actual)
        return d
    }

    mutating func rehacer(_ actual: Diseno) -> Diseno? {
        guard let d = futuro.popLast() else { return nil }
        pasado.append(actual)
        return d
    }
}

// MARK: - Qué le falta a un bloque

/// Lo que el editor sabe de la tienda para avisar por qué un bloque no se va a ver.
struct ContextoTienda: Equatable {
    var nombre = ""
    var whatsapp = false
    var direccionLocal = false
    var categorias: [String] = []
    var productos = 0
    var resenas = false
    var fotos = false
}

/// Por qué un bloque no se va a dibujar (regla «Nada vacío»), en una frase corta.
struct FaltaBloque: Equatable {
    var texto: String
    /// Se arregla en Más › Datos del negocio.
    var enNegocio = false

    static func de(_ b: Bloque, _ ctx: ContextoTienda) -> FaltaBloque? {
        guard let t = CatalogoBloques.shared.tipo(b.tipo) else { return nil }
        if let f = deCampos(t.campos, b.datos) { return f }
        let sinProductos = FaltaBloque(texto: "Cargá tu primer producto para que se vea.")
        switch b.tipo {
        case "whatsapp":
            if !ctx.whatsapp { return .init(texto: "Cargá tu WhatsApp en Datos del negocio para que se vea.", enNegocio: true) }
        case "local":
            if !ctx.direccionLocal { return .init(texto: "Cargá la dirección del local en Datos del negocio.", enNegocio: true) }
        case "resenas":
            if !ctx.resenas { return .init(texto: "Se va a ver cuando llegue tu primera reseña.") }
        case "categorias", "menu":
            if ctx.productos == 0 { return sinProductos }
            if ctx.categorias.isEmpty { return .init(texto: "Poné una categoría a tus productos para que se vea.") }
        case "galeria":
            if (b.datos["fotos"]?.lista ?? []).isEmpty, !ctx.fotos { return .init(texto: "Subí fotos o cargá productos con foto.") }
        case "productos", "oferta", "destacado", "combo":
            if ctx.productos == 0 { return sinProductos }
            if case let .objeto(p)? = b.datos["productos"] {
                let fuente = p["fuente"]?.texto
                if fuente == "categoria", (p["categoria"]?.texto ?? "").isEmpty { return .init(texto: "Elegí de qué categoría son los productos.") }
                if fuente == "elegidos", (p["ids"]?.lista ?? []).isEmpty { return .init(texto: "Elegí los productos que van.") }
            }
            if b.tipo == "oferta", let h = b.datos["hasta"]?.texto, let f = Fecha.iso(h), f < Date() {
                return .init(texto: "La oferta ya terminó: poné otra fecha.")
            }
        case "banner":
            if b.variante == "doble", (b.datos["segundo"]?["titulo"]?.texto ?? "").isEmpty {
                return .init(texto: "Escribí el título de la segunda promo (la variante Doble lleva dos).")
            }
        default: break
        }
        return nil
    }

    private static func deCampos(_ campos: [CatalogoBloques.Campo], _ datos: [String: ValorJSON]) -> FaltaBloque? {
        for c in campos {
            let v = datos[c.id] ?? c.porDefecto
            if c.requerido == true, v?.vacio ?? true {
                return .init(texto: c.tipo == "producto" || c.tipo == "elegidos" ? "Elegí «\(c.nombre.lowercased())»." : "Completá «\(c.nombre)».")
            }
            if let min = c.min, c.tipo == "lista" || c.tipo == "elegidos", (v?.lista?.count ?? 0) < min {
                return .init(texto: min == 1 ? "Sumá al menos uno en «\(c.nombre)»." : "Sumá al menos \(min) en «\(c.nombre)».")
            }
            if c.tipo == "lista", let item = c.item {
                for e in v?.lista ?? [] {
                    if case let .objeto(o) = e, let f = deCampos(item, o) {
                        let resto = f.texto.prefix(1).lowercased() + String(f.texto.dropFirst())
                        return .init(texto: "En «\(c.nombre)»: \(resto)")
                    }
                }
            }
        }
        return nil
    }
}

// MARK: - ValorJSON

extension ValorJSON {
    /// Vacío para la tienda: sin texto, lista vacía u objeto sin nada adentro.
    var vacio: Bool {
        switch self {
        case .nulo: return true
        case let .texto(s): return s.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
        case let .lista(l): return l.isEmpty
        case let .objeto(o): return o.values.allSatisfy(\.vacio)
        default: return false
        }
    }

    var numero: Double? { if case let .numero(n) = self { return n }; return nil }
    var objeto: [String: ValorJSON]? { if case let .objeto(o) = self { return o }; return nil }

    static func de<T: Encodable>(_ v: T) -> ValorJSON {
        guard let d = try? JSONEncoder().encode(v), let j = try? JSONDecoder().decode(ValorJSON.self, from: d) else { return .nulo }
        return j
    }

    func como<T: Decodable>(_ tipo: T.Type) -> T? {
        guard let d = try? JSONEncoder().encode(self) else { return nil }
        return try? JSONDecoder().decode(tipo, from: d)
    }
}
