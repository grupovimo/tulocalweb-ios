import Foundation

// Las ventanas (popups) de la tienda: diseno.ventanas (docs/especificaciones/tienda-promociones.md). Lee y escribe
// el JSON de la especificación; las opciones, los campos y las plantillas salen de bloques.json › ventanas.

/// Una ventana: qué muestra (contenido y sus datos), cómo se ve (variante), cuándo sale, a quién, dónde y entre qué fechas.
struct Ventana: Equatable, Identifiable {
    var id: String
    var nombre: String
    var activa = true
    var variante = "centrada"
    var contenido = "promo"
    var datos: [String: ValorJSON] = [:]
    var cuando = "entrar"
    /// Segundos o % de la página, según el disparo (sólo los que traen rango).
    var valor: Int?
    var cada = "sesion"
    var dias: Int?
    var publico = "todos"
    /// Vacío: en todas.
    var paginas: [String] = []
    var dispositivos = "todos"
    var desde: String?
    var hasta: String?

    private static var cat: CatalogoBloques.Ventanas? { CatalogoBloques.shared.ventanas }

    init(id: String, nombre: String) {
        self.id = id
        self.nombre = nombre
    }

    init(json v: ValorJSON) {
        id = v["id"]?.texto ?? Bloque.nuevoId()
        nombre = v["nombre"]?.texto ?? ""
        activa = v["activa"]?.bool ?? true
        variante = v["variante"]?.texto ?? "centrada"
        contenido = v["contenido"]?.texto ?? "promo"
        datos = v["datos"]?.objeto ?? [:]
        cuando = v["disparo"]?["cuando"]?.texto ?? "entrar"
        valor = v["disparo"]?["valor"]?.entero
        cada = v["frecuencia"]?["cada"]?.texto ?? "sesion"
        dias = v["frecuencia"]?["dias"]?.entero
        publico = v["publico"]?.texto ?? "todos"
        paginas = v["paginas"]?.lista?.compactMap(\.texto) ?? []
        dispositivos = v["dispositivos"]?.texto ?? "todos"
        desde = v["desde"]?.texto
        hasta = v["hasta"]?.texto
    }

    /// Para el PUT: `valor` y `dias` sólo en los disparos y frecuencias que traen rango (con el de por defecto si falta).
    var json: ValorJSON {
        var disparo: [String: ValorJSON] = ["cuando": .texto(cuando)]
        if let r = Self.cat?.disparo(cuando)?.valor { disparo["valor"] = .numero(Double(min(max(valor ?? r.porDefecto, r.min), r.max))) }
        var frecuencia: [String: ValorJSON] = ["cada": .texto(cada)]
        if let r = Self.cat?.frecuencia(cada)?.dias { frecuencia["dias"] = .numero(Double(min(max(dias ?? r.porDefecto, r.min), r.max))) }
        let n = nombre.trimmingCharacters(in: .whitespaces)
        var o: [String: ValorJSON] = [
            "id": .texto(id), "nombre": .texto(n.isEmpty ? "Ventana" : String(n.prefix(40))), "activa": .booleano(activa),
            "variante": .texto(variante), "contenido": .texto(contenido), "datos": .objeto(datos),
            "disparo": .objeto(disparo), "frecuencia": .objeto(frecuencia), "publico": .texto(publico),
            "paginas": .lista(paginas.map { .texto($0) }), "dispositivos": .texto(dispositivos)
        ]
        if let desde, !desde.isEmpty { o["desde"] = .texto(desde) }
        if let hasta, !hasta.isEmpty { o["hasta"] = .texto(hasta) }
        return .objeto(o)
    }

    /// Una en blanco: una promo centrada al entrar, una vez por visita.
    static func enBlanco(id: String) -> Ventana { Ventana(id: id, nombre: "Ventana nueva") }

    /// Desde una plantilla del catálogo: id nuevo, el monto del envío gratis en los textos y, si la trae, la cuenta
    /// regresiva a tantos días.
    static func dePlantilla(_ p: CatalogoBloques.Ventanas.Plantilla, id: String, envioGratis: Int?, ahora: Date = Date()) -> Ventana {
        let monto = envioGratis.map { Formato.pesos($0) } ?? "un monto"
        var v = Ventana(json: reemplazar(p.ventana, "{envioGratis}", monto))
        v.id = id
        if let d = p.hastaEnDias {
            let f = Calendar.current.date(byAdding: .day, value: d, to: ahora) ?? ahora
            v.datos["hasta"] = .texto(isoAR(f))
        }
        return v
    }

    private static func reemplazar(_ v: ValorJSON, _ a: String, _ b: String) -> ValorJSON {
        switch v {
        case let .texto(s): return .texto(s.replacingOccurrences(of: a, with: b))
        case let .lista(l): return .lista(l.map { reemplazar($0, a, b) })
        case let .objeto(o): return .objeto(o.mapValues { reemplazar($0, a, b) })
        default: return v
        }
    }

    static let formatoAR: ISO8601DateFormatter = {
        let f = ISO8601DateFormatter()
        f.formatOptions = [.withInternetDateTime]
        f.timeZone = TimeZone(identifier: "America/Argentina/Buenos_Aires")
        return f
    }()

    static func isoAR(_ f: Date) -> String { formatoAR.string(from: f) }

    /// Al cambiar de contenido quedan sólo los datos que el nuevo también tiene (título, texto, foto…).
    mutating func cambiarContenido(_ c: String) {
        guard c != contenido else { return }
        let campos = Set(Self.cat?.contenido(c)?.campos.map(\.id) ?? [])
        datos = datos.filter { campos.contains($0.key) }
        contenido = c
    }
}

// MARK: - Qué le falta

/// Por qué una ventana no va a salir en la tienda, en una frase. Las mismas reglas que Go (Ventana.Completa) y la web:
/// los requeridos del contenido cargados, el premio con lo suyo y, en la ruleta, algún peso mayor a 0.
enum FaltaVentana {
    static func de(_ v: Ventana, ahora: Date = Date()) -> String? {
        guard let cont = CatalogoBloques.shared.ventanas?.contenido(v.contenido) else { return nil }
        for c in cont.campos {
            let x = v.datos[c.id]
            if c.requerido == true, x?.vacio ?? true {
                if c.tipo == "cupon" { return "Elegí el cupón." }
                if c.tipo == "premio" { return "Elegí el premio." }
                if c.tipo == "lista" { return "Sumá \(c.nombre.lowercased())." }
                return "Completá «\(c.nombre)»."
            }
            if c.tipo == "premio", let f = dePremio(x) { return f }
            if c.tipo == "lista", let min = c.min, (x?.lista?.count ?? 0) < min { return "Sumá al menos \(min) en «\(c.nombre)»." }
        }
        if v.contenido == "ruleta" {
            let porciones = v.datos["porciones"]?.lista ?? []
            for (i, p) in porciones.enumerated() {
                if (p["texto"]?.texto ?? "").trimmingCharacters(in: .whitespaces).isEmpty { return "Poné el texto de la porción \(i + 1)." }
                if let f = dePremio(p["premio"]) { return "Porción \(i + 1): \(f.prefix(1).lowercased())\(f.dropFirst())" }
            }
            if porciones.reduce(0.0, { $0 + ($1["peso"]?.numero ?? 0) }) <= 0 { return "Poné un peso mayor a 0 en alguna porción." }
        }
        if let h = v.datos["hasta"]?.texto, let f = Fecha.iso(h), f < ahora { return "La cuenta regresiva ya terminó: cambiá la fecha." }
        if let h = v.hasta, let f = Fecha.iso(h), f < ahora { return "Ya pasó la fecha en que termina." }
        if let d = v.desde, let h = v.hasta, let fd = Fecha.iso(d), let fh = Fecha.iso(h), fh <= fd { return "La fecha de fin va después de la de inicio." }
        return nil
    }

    /// El premio trae lo que necesita para entregarse.
    static func dePremio(_ v: ValorJSON?) -> String? {
        guard let p = v?.objeto, let tipo = p["tipo"]?.texto else { return "Elegí el premio." }
        switch tipo {
        case "cuponFijo": return (p["cupon"]?.vacio ?? true) ? "Elegí el cupón del premio." : nil
        case "cuponUnico":
            let d = p["descuento"]?.texto ?? ""
            if d == "envioGratis" { return nil }
            if d != "porcentaje" && d != "monto" { return "Elegí qué descuento da el premio." }
            return p["valor"] == nil ? "Poné de cuánto es el descuento." : nil
        case "descarga": return (p["archivo"]?.vacio ?? true) ? "Poné el link del archivo para descargar." : nil
        case "nada": return nil
        default: return "Elegí el premio."
        }
    }
}

// MARK: - Métricas

/// Lo que pasó con una ventana en el período (GET /panel/promos/metricas): vistas, cierres, envíos, clics y ventas.
struct MetricaVentana: Equatable {
    var vistas = 0
    var cierres = 0
    var envios = 0
    var clics = 0
    var ventas = 0
    var monto = 0

    /// Qué parte de los que la vieron hizo algo: dejó el mail (captación, ruleta) o tocó (promo, cupón).
    func tasa(_ contenido: String) -> Double? {
        guard vistas > 0 else { return nil }
        let n = contenido == "captacion" || contenido == "ruleta" ? envios : clics
        return Double(n) / Double(vistas) * 100
    }

    /// Por ventana, sumando los días. Acepta una lista de filas ({ventana|id, dia?, vistas…}), {ventanas: […]} o
    /// un objeto por id.
    static func leer(_ v: ValorJSON) -> [String: MetricaVentana] {
        var r: [String: MetricaVentana] = [:]
        func sumar(_ id: String, _ o: ValorJSON) {
            // Con el detalle por día, se suman los días (la fila puede traer además el total).
            if let dias = o["dias"]?.lista { for d in dias { sumar(id, d) }; return }
            var m = r[id] ?? MetricaVentana()
            m.vistas += o["vistas"]?.entero ?? 0
            m.cierres += o["cierres"]?.entero ?? 0
            m.envios += o["envios"]?.entero ?? 0
            m.clics += o["clics"]?.entero ?? 0
            m.ventas += o["ventas"]?.entero ?? 0
            m.monto += o["monto"]?.entero ?? 0
            r[id] = m
        }
        func filas(_ l: [ValorJSON]) {
            for f in l { if let id = f["ventana"]?.texto ?? f["id"]?.texto { sumar(id, f) } }
        }
        switch v {
        case let .lista(l): filas(l)
        case let .objeto(o):
            if let l = o["ventanas"]?.lista { filas(l) } else if let l = o["filas"]?.lista { filas(l) } else if case let .objeto(p)? = o["ventanas"] {
                for (id, x) in p { sumar(id, x) }
            } else {
                for (id, x) in o where x.objeto != nil { sumar(id, x) }
            }
        default: break
        }
        return r
    }
}
