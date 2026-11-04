import SwiftUI
import UIKit
import WebKit

/// Mi tienda › Estilo (components/schemas/EstiloTienda), igual que en la web
/// (apps/web/src/lib/tienda/estilo/estilo.ts): colores por rol y por zona, tipografías, logos,
/// anuncio, títulos de las secciones, redes y el menú. Todo opcional: vacío es lo de la plantilla.
struct EstiloTienda: Codable, Equatable {
    var colores: [String: String]?
    var fuenteTitulos: String?
    var fuenteTexto: String?
    var logoOscuro: String?
    var icono: String?
    var logoAlto: Int?
    var anuncio: String?
    var anuncioLink: String?
    var leyendaPie: String?
    var textos: [String: String]?
    var redes: [String: String]?
    var menu: [ItemMenu]?

    struct ItemMenu: Codable, Equatable, Hashable { var t: String; var a: String }

    /// Para mandarlo dentro del diseño: sin lo vacío (Go lo limpia igual).
    var json: ValorJSON {
        var e = self
        e.colores = e.colores?.filter { !$0.value.isEmpty }
        e.textos = e.textos?.filter { !$0.value.trimmingCharacters(in: .whitespaces).isEmpty }
        e.redes = e.redes?.filter { !$0.value.trimmingCharacters(in: .whitespaces).isEmpty }
        if e.logoAlto == 0 { e.logoAlto = nil }
        guard let d = try? JSONEncoder().encode(e), let v = try? JSONDecoder().decode(ValorJSON.self, from: d) else { return .objeto([:]) }
        return v
    }

    static func desde(_ v: ValorJSON?) -> EstiloTienda {
        guard let v, let d = try? JSONEncoder().encode(v), let e = try? JSONDecoder().decode(EstiloTienda.self, from: d) else { return EstiloTienda() }
        return e
    }
}

/// Lo que comparte con la web (Recursos/estilo.json, de scripts/estilo/para-ios.ts).
struct CatalogoEstilo: Decodable {
    struct Fuente: Decodable, Identifiable { var id: String; var nombre: String; var tipo: String }
    struct Combinacion: Decodable, Identifiable { var id: String; var nombre: String; var colores: [String: String] }
    struct Pagina: Decodable, Identifiable { var id: String; var nombre: String }
    var fuentes: [Fuente]
    var combinaciones: [Combinacion]
    var roles: [String: [String: String]]
    var acentos: [String: [String]]
    var paginas: [Pagina]
    var maxMenu: Int

    static let shared: CatalogoEstilo = {
        guard let u = Bundle.main.url(forResource: "estilo", withExtension: "json"), let d = try? Data(contentsOf: u),
              let v = try? JSONDecoder().decode(CatalogoEstilo.self, from: d) else {
            return CatalogoEstilo(fuentes: [], combinaciones: [], roles: [:], acentos: [:], paginas: [], maxMenu: 8)
        }
        return v
    }()

    static let rolesColor: [(String, String)] = [
        ("fondo", "Fondo"), ("superficie", "Tarjetas"), ("texto", "Texto"), ("textoSuave", "Texto suave"),
        ("borde", "Bordes"), ("acento", "Botones y precios"), ("sobreAcento", "Texto de los botones"),
        ("encabezado", "Fondo del encabezado"), ("encabezadoTexto", "Texto del encabezado"),
        ("pie", "Fondo del pie"), ("pieTexto", "Texto del pie")
    ]
    static let titulos: [(String, String)] = [("cats", "Categorías"), ("prods", "Productos destacados"), ("revs", "Reseñas"), ("ig", "Fotos")]
    static let redes: [(String, String, String, UIKeyboardType)] = [
        ("instagram", "Instagram", "@tutienda", .default), ("tiktok", "TikTok", "@tutienda", .default),
        ("facebook", "Facebook", "facebook.com/tutienda", .URL), ("youtube", "YouTube", "@tucanal", .default),
        ("email", "Email", "hola@tutienda.com", .emailAddress), ("telefono", "Teléfono", "11 5555-5555", .phonePad)
    ]
}

// MARK: - Colores

enum Hex {
    static func color(_ s: String?) -> Color? {
        guard let s, s.count == 7, s.hasPrefix("#"), let n = UInt32(s.dropFirst(), radix: 16) else { return nil }
        return Color(hex: n)
    }
    static func de(_ c: Color) -> String {
        var r: CGFloat = 0, g: CGFloat = 0, b: CGFloat = 0, a: CGFloat = 0
        UIColor(c).getRed(&r, green: &g, blue: &b, alpha: &a)
        let v = [r, g, b].map { max(0, min(255, Int(($0 * 255).rounded()))) }
        return String(format: "#%02X%02X%02X", v[0], v[1], v[2])
    }
    /// Contraste WCAG 2 entre dos #RRGGBB (1 a 21).
    static func contraste(_ a: String, _ b: String) -> Double {
        func lum(_ h: String) -> Double {
            guard let n = UInt32(h.dropFirst(), radix: 16) else { return 0 }
            let c = [Double((n >> 16) & 0xFF), Double((n >> 8) & 0xFF), Double(n & 0xFF)].map { v -> Double in
                let x = v / 255
                return x <= 0.03928 ? x / 12.92 : pow((x + 0.055) / 1.055, 2.4)
            }
            return 0.2126 * c[0] + 0.7152 * c[1] + 0.0722 * c[2]
        }
        let (x, y) = (lum(a), lum(b))
        return (max(x, y) + 0.05) / (min(x, y) + 0.05)
    }
}

/// Los grupos del estilo, para Mi tienda (los mismos de la web).
struct EditorEstilo: View {
    @Binding var estilo: EstiloTienda
    @Binding var paleta: Int
    var plantilla: String
    var categorias: [String]
    @State private var abierto: String? = "colores"
    private let cat = CatalogoEstilo.shared

    private var original: [String: String] { cat.roles[plantilla] ?? cat.roles["litoral"] ?? [:] }

    /// El color que se ve: el elegido, el acento rápido o el de la plantilla (encabezado y pie siguen a su rol).
    func visto(_ rol: String) -> String {
        if let c = estilo.colores?[rol], c.count == 7 { return c }
        if rol == "acento", let a = cat.acentos[plantilla], a.indices.contains(paleta) { return a[paleta] }
        if rol == "encabezado" { return visto("fondo") }
        if rol == "encabezadoTexto" { return visto("texto") }
        if rol == "pie" || rol == "pieTexto", let o = original[rol] {
            if o == original["texto"] { return visto("texto") }
            if o == original["fondo"] { return visto("fondo") }
        }
        return original[rol] ?? "#FFFFFF"
    }

    private func poner(_ c: [String: String]) {
        var colores = estilo.colores ?? [:]
        for (k, v) in c { colores[k] = v.isEmpty ? nil : v }
        estilo.colores = colores.isEmpty ? nil : colores
    }

    private var avisos: [String] {
        [
            Hex.contraste(visto("texto"), visto("fondo")) < 4.5 ? "El texto casi no se lee sobre el fondo." : nil,
            Hex.contraste(visto("sobreAcento"), visto("acento")) < 4.5 ? "El texto de los botones casi no se lee." : nil,
            Hex.contraste(visto("pieTexto"), visto("pie")) < 4.5 ? "El texto del pie casi no se lee." : nil
        ].compactMap { $0 }
    }

    var body: some View {
        VStack(spacing: 14) {
            grupo("colores", "Colores", "paintpalette") { colores }
            grupo("fuentes", "Tipografías", "textformat") { fuentes }
            grupo("menu", "Menú", "line.3.horizontal") { menu }
            grupo("titulos", "Títulos de las secciones", "textformat.size") { titulos }
            grupo("anuncio", "Anuncio y pie", "megaphone") { anuncioYPie }
            grupo("redes", "Redes y contacto", "at") { redes }
        }
    }

    private func grupo<V: View>(_ id: String, _ titulo: String, _ simbolo: String, @ViewBuilder _ v: @escaping () -> V) -> some View {
        TarjetaSeccion(titulo: titulo, simbolo: simbolo) {
            Button {
                withAnimation(.snappy) { abierto = abierto == id ? nil : id }
            } label: {
                HStack {
                    Text(abierto == id ? "Ocultar" : "Editar").estilo(.subheadFuerte).foregroundStyle(C.primarioTexto)
                    Spacer()
                    Image(systemName: abierto == id ? "chevron.up" : "chevron.down")
                        .font(.system(size: 12, weight: .bold)).foregroundStyle(C.texto2)
                }
                .contentShape(Rectangle())
            }
            .buttonStyle(.plain)
            if abierto == id { v() }
        }
    }

    // MARK: Colores
    private var colores: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("Acento rápido").estilo(.footnoteFuerte).foregroundStyle(C.texto2)
            HStack(spacing: 10) {
                ForEach(Array((cat.acentos[plantilla] ?? []).enumerated()), id: \.offset) { i, a in
                    Button {
                        paleta = i
                        poner(["acento": "", "sobreAcento": ""])
                    } label: {
                        Circle().fill(Hex.color(a) ?? .gray).frame(width: 36, height: 36)
                            .overlay(Circle().strokeBorder(C.texto, lineWidth: paleta == i && estilo.colores?["acento"] == nil ? 3 : 0))
                    }
                    .accessibilityLabel(i == 0 ? "El de la plantilla" : "Color \(i + 1)")
                }
            }
            Text("Combinaciones").estilo(.footnoteFuerte).foregroundStyle(C.texto2)
            LazyVGrid(columns: [GridItem(.flexible()), GridItem(.flexible())], spacing: 8) {
                ForEach(cat.combinaciones) { c in
                    Button {
                        var nuevo = c.colores
                        for z in ["encabezado", "encabezadoTexto", "pie", "pieTexto"] { nuevo[z] = "" }
                        poner(nuevo)
                    } label: {
                        HStack(spacing: 8) {
                            HStack(spacing: 2) {
                                ForEach(["superficie", "texto"], id: \.self) { r in
                                    RoundedRectangle(cornerRadius: 2).fill(Hex.color(c.colores[r]) ?? .gray).frame(width: 7, height: 14)
                                }
                                RoundedRectangle(cornerRadius: 2).fill(Hex.color(visto("acento")) ?? .gray).frame(width: 7, height: 14)
                            }
                            .padding(4).background(Hex.color(c.colores["fondo"]) ?? .white, in: RoundedRectangle(cornerRadius: 7))
                            Text(c.nombre).estilo(.subheadFuerte).foregroundStyle(C.texto).lineLimit(1)
                            Spacer(minLength: 0)
                        }
                        .padding(8).frame(maxWidth: .infinity)
                        .overlay(RoundedRectangle(cornerRadius: 12).strokeBorder(C.linea2))
                    }
                    .buttonStyle(.plain)
                }
            }
            Text("Todos los colores").estilo(.footnoteFuerte).foregroundStyle(C.texto2)
            VStack(spacing: 0) {
                ForEach(CatalogoEstilo.rolesColor, id: \.0) { rol, nombre in
                    HStack(spacing: 10) {
                        ColorPicker(nombre, selection: Binding(
                            get: { Hex.color(visto(rol)) ?? .white },
                            set: { poner([rol: Hex.de($0)]) }
                        ), supportsOpacity: false)
                        .estilo(.body16).foregroundStyle(C.texto)
                        if estilo.colores?[rol] != nil {
                            Button { poner([rol: ""]) } label: {
                                Image(systemName: "arrow.uturn.backward").font(.system(size: 13, weight: .semibold))
                            }
                            .foregroundStyle(C.texto2).accessibilityLabel("Volver \(nombre.lowercased()) al de la plantilla")
                        }
                    }
                    .padding(.vertical, 6)
                }
            }
            ForEach(avisos, id: \.self) { a in
                ChipSimple(texto: a, fondo: C.avisoFondo, tinta: C.avisoTinta)
            }
            if estilo.colores != nil || paleta != 0 {
                Button("Volver a los colores de la plantilla") { estilo.colores = nil; paleta = 0 }
                    .buttonStyle(BotonTexto())
            }
        }
    }

    // MARK: Tipografías
    private var fuentes: some View {
        VStack(alignment: .leading, spacing: 12) {
            selectorFuente("Títulos", seleccion: $estilo.fuenteTitulos)
            selectorFuente("Texto", seleccion: $estilo.fuenteTexto)
        }
    }

    private func selectorFuente(_ titulo: String, seleccion: Binding<String?>) -> some View {
        HStack {
            Text(titulo).estilo(.body16).foregroundStyle(C.texto)
            Spacer()
            Picker(titulo, selection: Binding(get: { seleccion.wrappedValue ?? "" }, set: { seleccion.wrappedValue = $0.isEmpty ? nil : $0 })) {
                Text("La de la plantilla").tag("")
                ForEach(cat.fuentes) { f in Text(f.nombre).tag(f.id) }
            }
            .pickerStyle(.menu).tint(C.primarioTexto)
        }
    }

    // MARK: Menú
    private var destinos: [(String, String)] {
        cat.paginas.map { ($0.id, $0.nombre) } + categorias.map { ("cat:\($0)", "Categoría: \($0)") } + [("link", "Un link propio…")]
    }

    private var menu: some View {
        let items = estilo.menu ?? []
        return VStack(alignment: .leading, spacing: 10) {
            Text(items.isEmpty ? "Va el menú de la plantilla. Armá el tuyo para sumar pestañas como Nosotros." : "Así se ve el menú de tu tienda.")
                .estilo(.footnote).foregroundStyle(C.texto2)
            ForEach(Array(items.enumerated()), id: \.offset) { i, it in
                VStack(alignment: .leading, spacing: 8) {
                    HStack(spacing: 8) {
                        TextField("Texto de la pestaña", text: Binding(
                            get: { it.t },
                            set: { v in cambiarItem(i) { $0.t = String(v.prefix(30)) } }
                        ))
                        .estilo(.body16).foregroundStyle(C.texto)
                        .padding(10).background(C.crema2, in: RoundedRectangle(cornerRadius: 10))
                        Button { moverItem(i) } label: { Image(systemName: "chevron.up") }.disabled(i == 0)
                            .accessibilityLabel("Subir \(it.t)")
                        Button { estilo.menu = items.enumerated().filter { $0.offset != i }.map(\.element) } label: { Image(systemName: "trash") }
                            .accessibilityLabel("Quitar \(it.t)")
                    }
                    .font(.system(size: 14, weight: .semibold)).foregroundStyle(C.primarioTexto)
                    Picker("Adónde lleva", selection: Binding(
                        get: { it.a.hasPrefix("https://") ? "link" : it.a },
                        set: { d in cambiarDestino(i, d) }
                    )) {
                        ForEach(destinos, id: \.0) { d in Text(d.1).tag(d.0) }
                    }
                    .pickerStyle(.menu).tint(C.primarioTexto)
                    if it.a.hasPrefix("https://") {
                        TextField("https://", text: Binding(get: { it.a }, set: { v in cambiarItem(i) { $0.a = v.trimmingCharacters(in: .whitespaces) } }))
                            .keyboardType(.URL).textInputAutocapitalization(.never).autocorrectionDisabled()
                            .estilo(.body16).padding(10).background(C.crema2, in: RoundedRectangle(cornerRadius: 10))
                    }
                }
                .padding(.vertical, 6)
                if i < items.count - 1 { Separador() }
            }
            if items.count < cat.maxMenu {
                Button {
                    let usadas = Set(items.map(\.a))
                    let p = cat.paginas.first { !usadas.contains($0.id) } ?? cat.paginas[1]
                    estilo.menu = items + [.init(t: p.nombre, a: p.id)]
                } label: { Label("Agregar pestaña", systemImage: "plus") }
                    .buttonStyle(BotonTonal(alto: 40))
            }
            if !items.isEmpty {
                Button("Volver al menú de la plantilla") { estilo.menu = nil }.buttonStyle(BotonTexto())
            }
        }
    }

    private func cambiarItem(_ i: Int, _ f: (inout EstiloTienda.ItemMenu) -> Void) {
        guard var m = estilo.menu, m.indices.contains(i) else { return }
        f(&m[i])
        estilo.menu = m
    }

    private func moverItem(_ i: Int) {
        guard var m = estilo.menu, i > 0 else { return }
        m.swapAt(i, i - 1)
        estilo.menu = m
    }

    private func cambiarDestino(_ i: Int, _ d: String) {
        let nombre = { (a: String) in self.destinos.first { $0.0 == a }?.1.replacingOccurrences(of: "Categoría: ", with: "") }
        cambiarItem(i) { it in
            let viejo = nombre(it.a)
            it.a = d == "link" ? "https://" : d
            if d != "link", it.t.isEmpty || it.t == viejo { it.t = nombre(d) ?? it.t }
        }
    }

    // MARK: Textos
    private var titulos: some View {
        VStack(alignment: .leading, spacing: 12) {
            ForEach(CatalogoEstilo.titulos, id: \.0) { id, nombre in
                CampoTexto(titulo: "\(nombre) (vacío: el de la plantilla)", texto: Binding(
                    get: { estilo.textos?[id] ?? "" },
                    set: { v in var t = estilo.textos ?? [:]; t[id] = String(v.prefix(60)); estilo.textos = t }
                ))
            }
        }
    }

    private var anuncioYPie: some View {
        VStack(alignment: .leading, spacing: 12) {
            CampoTexto(titulo: "Barra de anuncio (arriba de todo)", texto: Binding(
                get: { estilo.anuncio ?? "" }, set: { estilo.anuncio = String($0.prefix(120)) }))
            CampoTexto(titulo: "Link del anuncio (opcional)", texto: Binding(
                get: { estilo.anuncioLink ?? "" }, set: { estilo.anuncioLink = $0.trimmingCharacters(in: .whitespaces) }), teclado: .URL)
            CampoTexto(titulo: "Línea del pie", texto: Binding(
                get: { estilo.leyendaPie ?? "" }, set: { estilo.leyendaPie = String($0.prefix(160)) }))
        }
    }

    private var redes: some View {
        VStack(alignment: .leading, spacing: 12) {
            ForEach(CatalogoEstilo.redes, id: \.0) { id, nombre, _, teclado in
                CampoTexto(titulo: nombre, texto: Binding(
                    get: { estilo.redes?[id] ?? "" },
                    set: { v in var r = estilo.redes ?? [:]; r[id] = v; estilo.redes = r }
                ), teclado: teclado)
            }
            Text("Se muestran en el pie de tu tienda.").estilo(.footnote).foregroundStyle(C.texto2)
        }
    }
}

// MARK: - Vista previa

/// La tienda de verdad con el borrador (la misma página que el editor web, /panel/mi-tienda/vista,
/// con el token de la app). La página misma no deja navegar los links.
struct VistaPreviaTienda: UIViewRepresentable {
    var token: String?
    var vuelta: Int

    final class Coordinador { var cargada = -1 }

    func makeCoordinator() -> Coordinador { Coordinador() }

    func makeUIView(context: Context) -> WKWebView {
        let w = WKWebView(frame: .zero, configuration: WKWebViewConfiguration())
        w.isOpaque = false
        w.backgroundColor = .clear
        return w
    }

    func updateUIView(_ w: WKWebView, context: Context) {
        guard context.coordinator.cargada != vuelta else { return }
        context.coordinator.cargada = vuelta
        var r = URLRequest(url: Web.base.appending(path: "panel/mi-tienda/vista"))
        r.cachePolicy = .reloadIgnoringLocalCacheData
        if let token { r.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization") }
        w.load(r)
    }
}
