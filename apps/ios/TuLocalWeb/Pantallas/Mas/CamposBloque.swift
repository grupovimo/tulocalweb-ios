import PhotosUI
import SwiftUI

// Los campos del panel de un bloque, generados desde bloques.json › campos (tiposCampo): un editor por tipo.
// Lo vacío se saca de los datos, así la tienda usa lo de por defecto o lo sugerido.

/// Lo que los campos necesitan de la tienda: productos para buscar, categorías y el contexto de los avisos.
@Observable
final class DatosEditor {
    var contexto = ContextoTienda()
    /// Los primeros productos (para elegir sin buscar) y todos los que se fueron viendo, por id.
    var productos: [ProductoFila] = []
    var conocidos: [String: ProductoFila] = [:]

    func cargarProductos(_ api: Api) async {
        let q = [URLQueryItem(name: "orden", value: "sold"), URLQueryItem(name: "cuantos", value: "60")]
        guard let l: PaginaProductos = try? await api.get("/v1/panel/listas/productos", query: q) else { return }
        productos = l.filas
        for p in l.filas { conocidos[p.id] = p }
    }

    func buscar(_ texto: String, _ api: Api) async -> [ProductoFila] {
        let t = texto.trimmingCharacters(in: .whitespaces)
        guard !t.isEmpty else { return productos }
        let q = [URLQueryItem(name: "q", value: t), URLQueryItem(name: "cuantos", value: "40")]
        guard let l: PaginaProductos = try? await api.get("/v1/panel/listas/productos", query: q) else {
            return productos.filter { $0.nombre.localizedCaseInsensitiveContains(t) }
        }
        for p in l.filas { conocidos[p.id] = p }
        return l.filas
    }

    func nombre(_ id: String) -> String { conocidos[id]?.nombre ?? "Un producto" }
}

/// Todos los campos de un bloque (o de un elemento de una lista).
struct EditorCampos: View {
    var campos: [CatalogoBloques.Campo]
    @Binding var datos: [String: ValorJSON]

    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            ForEach(campos) { c in
                EditorCampo(campo: c, valor: Binding(
                    get: { datos[c.id] },
                    set: { v in
                        if let v, !v.vacio || v.bool != nil { datos[c.id] = v } else { datos[c.id] = nil }
                    }
                ))
            }
        }
    }
}

/// Un campo: elige el editor por su tipo.
struct EditorCampo: View {
    var campo: CatalogoBloques.Campo
    @Binding var valor: ValorJSON?
    @Environment(DatosEditor.self) private var datos
    private let cat = CatalogoBloques.shared

    private var texto: Binding<String> {
        Binding(get: { valor?.texto ?? "" }, set: { v in valor = .texto(campo.max.map { String(v.prefix($0)) } ?? v) })
    }

    /// Lo que se ve si se deja vacío.
    private var sugerencia: String {
        if let d = campo.porDefecto?.texto { return d }
        switch campo.sugerido {
        case "nombre": return datos.contexto.nombre.isEmpty ? "El nombre de tu tienda" : datos.contexto.nombre
        case "lema": return "La frase de tu tienda"
        default: return ""
        }
    }

    var body: some View {
        switch campo.tipo {
        case "texto":
            CampoLinea(titulo: campo.nombre, texto: texto, sugerencia: sugerencia, max: campo.max)
        case "textoLargo":
            CampoLinea(titulo: campo.nombre, texto: texto, sugerencia: sugerencia, max: campo.max, largo: true)
        case "video":
            VStack(alignment: .leading, spacing: 6) {
                CampoLinea(titulo: campo.nombre, texto: texto, sugerencia: "https://youtube.com/…", teclado: .URL)
                if let v = valor?.texto, !v.isEmpty, !EditorCampo.esVideo(v) {
                    Text("Pegá un link de YouTube o de Vimeo.").estilo(.footnote).foregroundStyle(C.errorTinta)
                }
            }
        case "foto":
            SubidorFotos(titulo: campo.nombre, max: 1, fotos: Binding(
                get: { valor?.texto.map { [$0] } ?? [] },
                set: { valor = $0.first.map { .texto($0) } }
            ))
        case "fotos":
            SubidorFotos(titulo: campo.nombre, max: campo.max ?? 12, fotos: Binding(
                get: { valor?.lista?.compactMap(\.texto) ?? [] },
                set: { valor = .lista($0.map { .texto($0) }) }
            ))
        case "link":
            SelectorLink(titulo: campo.nombre, link: texto)
        case "boton":
            EditorBoton(titulo: campo.nombre, valor: $valor)
        case "fondo":
            SelectorOpcion(titulo: campo.nombre, opciones: cat.fondos.opciones, nombres: cat.fondos.nombres,
                           elegido: Binding(get: { valor?.texto ?? campo.porDefecto?.texto ?? "normal" }, set: { valor = .texto($0) }))
        case "opcion":
            SelectorOpcion(titulo: campo.nombre, opciones: campo.opciones ?? [], nombres: campo.nombres ?? campo.opciones ?? [],
                           elegido: Binding(get: { valor?.texto ?? campo.porDefecto?.texto ?? campo.opciones?.first ?? "" }, set: { valor = .texto($0) }))
        case "fecha":
            EditorFecha(titulo: campo.nombre, valor: $valor)
        case "numero":
            let n = valor?.entero ?? campo.porDefecto?.entero ?? campo.min ?? 0
            Stepper(value: Binding(get: { n }, set: { valor = .numero(Double($0)) }), in: (campo.min ?? 0)...(campo.max ?? 99)) {
                HStack {
                    Text(campo.nombre).estilo(.body16).foregroundStyle(C.texto)
                    Spacer()
                    Text("\(n)").estilo(.fila).monospacedDigit().foregroundStyle(C.texto)
                }
            }
        case "si":
            Interruptor(texto: campo.nombre, prendido: Binding(
                get: { valor?.bool ?? campo.porDefecto?.bool ?? false }, set: { valor = .booleano($0) }))
        case "icono":
            SelectorIcono(titulo: campo.nombre, elegido: texto)
        case "producto":
            SelectorProductos(titulo: campo.nombre, max: 1, ids: Binding(
                get: { valor?.texto.map { [$0] } ?? [] }, set: { valor = $0.first.map { .texto($0) } }))
        case "elegidos":
            SelectorProductos(titulo: campo.nombre, min: campo.min ?? 0, max: campo.max ?? 12, ids: Binding(
                get: { valor?.lista?.compactMap(\.texto) ?? [] }, set: { valor = .lista($0.map { .texto($0) }) }))
        case "productos":
            EditorFuenteProductos(campo: campo, valor: $valor)
        case "categoria":
            SelectorCategoria(titulo: campo.nombre, elegida: texto)
        case "categorias":
            SelectorCategorias(titulo: campo.nombre, elegidas: Binding(
                get: { valor?.lista?.compactMap(\.texto) ?? [] }, set: { valor = .lista($0.map { .texto($0) }) }))
        case "lista":
            AnyView(EditorLista(campo: campo, valor: $valor))
        case "grupo":
            AnyView(
                VStack(alignment: .leading, spacing: 12) {
                    Text(campo.nombre).estilo(.subheadFuerte).foregroundStyle(C.texto)
                    EditorCampos(campos: campo.item ?? [], datos: Binding(
                        get: { valor?.objeto ?? [:] }, set: { valor = .objeto($0) }))
                }
                .padding(12)
                .overlay(RoundedRectangle(cornerRadius: 14, style: .continuous).strokeBorder(C.linea2))
            )
        default:
            EmptyView()
        }
    }

    static func esVideo(_ s: String) -> Bool {
        guard let u = URL(string: s), u.scheme == "https", let h = u.host?.lowercased() else { return false }
        return ["youtube.com", "www.youtube.com", "m.youtube.com", "youtu.be", "vimeo.com", "www.vimeo.com", "player.vimeo.com"].contains(h)
    }
}

// MARK: - Texto

/// Un texto con su título arriba y, de placeholder, lo que se ve si queda vacío.
struct CampoLinea: View {
    var titulo: String
    @Binding var texto: String
    var sugerencia = ""
    var max: Int?
    var largo = false
    var teclado: UIKeyboardType = .default

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            HStack {
                Text(titulo).estilo(.footnoteFuerte).foregroundStyle(C.texto2)
                Spacer()
                if let max, texto.count > max * 3 / 4 {
                    Text("\(texto.count)/\(max)").estilo(.footnote).monospacedDigit().foregroundStyle(C.texto3)
                }
            }
            TextField("", text: $texto, prompt: Text(sugerencia.isEmpty ? titulo : sugerencia).foregroundStyle(C.texto3),
                      axis: largo ? .vertical : .horizontal)
                .lineLimit(largo ? 3...8 : 1...1)
                .keyboardType(teclado)
                .textInputAutocapitalization(teclado == .URL ? .never : .sentences)
                .autocorrectionDisabled(teclado == .URL)
                .estilo(.body16).foregroundStyle(C.texto)
                .padding(.horizontal, 14).padding(.vertical, 13)
                .background(C.superficie, in: RoundedRectangle(cornerRadius: 14, style: .continuous))
                .overlay(RoundedRectangle(cornerRadius: 14, style: .continuous).strokeBorder(C.linea2))
        }
    }
}

// MARK: - Fotos

/// Una o varias fotos: se eligen del carrete, se suben con POST /panel/fotos y se pueden quitar.
struct SubidorFotos: View {
    var titulo: String
    var max: Int
    @Binding var fotos: [String]
    @Environment(Estado.self) private var estado
    @State private var elegidas: [PhotosPickerItem] = []
    @State private var subiendo = 0

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Text(titulo).estilo(.footnoteFuerte).foregroundStyle(C.texto2)
                Spacer()
                if max > 1 { Text("\(fotos.count) de \(max)").estilo(.footnote).monospacedDigit().foregroundStyle(C.texto3) }
            }
            ScrollView(.horizontal, showsIndicators: false) {
                HStack(spacing: 8) {
                    ForEach(Array(fotos.enumerated()), id: \.offset) { i, f in
                        Miniatura(foto: f, tam: 76, radio: 12)
                            .overlay(alignment: .topTrailing) {
                                Button { fotos.remove(at: i) } label: {
                                    Image(systemName: "xmark").font(.system(size: 10, weight: .bold)).foregroundStyle(.white)
                                        .frame(width: 22, height: 22).background(.black.opacity(0.6), in: Circle())
                                }
                                .padding(4)
                                .accessibilityLabel("Quitar la foto \(i + 1)")
                            }
                    }
                    ForEach(0..<subiendo, id: \.self) { _ in
                        ProgressView().tint(C.acento).frame(width: 76, height: 76).background(C.crema2, in: RoundedRectangle(cornerRadius: 12))
                    }
                    if fotos.count + subiendo < max {
                        PhotosPicker(selection: $elegidas, maxSelectionCount: max - fotos.count - subiendo, matching: .images) {
                            VStack(spacing: 4) {
                                Image(systemName: "photo.badge.plus").font(.system(size: 20))
                                Text(fotos.isEmpty ? "Elegir" : "Sumar").estilo(.footnoteFuerte)
                            }
                            .foregroundStyle(C.primarioTinta)
                            .frame(width: 76, height: 76)
                            .background(C.crema2, in: RoundedRectangle(cornerRadius: 12, style: .continuous))
                        }
                    }
                }
            }
        }
        .onChange(of: elegidas) { _, items in
            guard !items.isEmpty else { return }
            elegidas = []
            Task { await subir(items) }
        }
    }

    private func subir(_ items: [PhotosPickerItem]) async {
        subiendo += items.count
        for item in items {
            defer { subiendo -= 1 }
            guard let d = try? await item.loadTransferable(type: Data.self), let jpg = Web.jpegParaSubir(d) else { continue }
            do {
                let u = try await Web.subirFoto(jpg, token: estado.api.token)
                if fotos.count < max { fotos.append(u) } else { fotos = [u] }
            } catch {
                estado.avisar("No pudimos subir la foto", ok: false)
            }
        }
    }
}

// MARK: - Links y botones

/// Adónde lleva un link: una página, una categoría, un producto o un link propio (tiposCampo › link).
struct SelectorLink: View {
    var titulo: String
    @Binding var link: String
    @Environment(DatosEditor.self) private var datos
    @Environment(Estado.self) private var estado
    @State private var buscando = false

    private var nombre: String {
        if link.isEmpty { return "Sin link" }
        if let p = CatalogoEstilo.shared.paginas.first(where: { $0.id == link }) { return p.nombre }
        if link.hasPrefix("cat:") { return "Categoría: \(link.dropFirst(4))" }
        if link.hasPrefix("prod:") { return "Producto: \(datos.nombre(String(link.dropFirst(5))))" }
        return "Un link propio"
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text(titulo).estilo(.footnoteFuerte).foregroundStyle(C.texto2)
            Menu {
                Button("Sin link") { link = "" }
                Section("Páginas") {
                    ForEach(CatalogoEstilo.shared.paginas) { p in Button(p.nombre) { link = p.id } }
                }
                if !datos.contexto.categorias.isEmpty {
                    Section("Categorías") {
                        ForEach(datos.contexto.categorias, id: \.self) { c in Button(c) { link = "cat:\(c)" } }
                    }
                }
                Button("Un producto…") { buscando = true }
                Button("Un link propio…") { link = "https://" }
            } label: {
                HStack {
                    Text(nombre).estilo(.body16).foregroundStyle(link.isEmpty ? C.texto3 : C.texto).lineLimit(1)
                    Spacer()
                    Image(systemName: "chevron.up.chevron.down").font(.system(size: 12, weight: .semibold)).foregroundStyle(C.texto2)
                }
                .padding(.horizontal, 14).frame(height: 48)
                .background(C.superficie, in: RoundedRectangle(cornerRadius: 14, style: .continuous))
                .overlay(RoundedRectangle(cornerRadius: 14, style: .continuous).strokeBorder(C.linea2))
            }
            if link.hasPrefix("http") {
                CampoLinea(titulo: "Link", texto: Binding(get: { link }, set: { link = $0.trimmingCharacters(in: .whitespaces) }),
                           sugerencia: "https://", teclado: .URL)
                if !link.hasPrefix("https://") { Text("Tiene que empezar con https://").estilo(.footnote).foregroundStyle(C.errorTinta) }
            }
        }
        .sheet(isPresented: $buscando) {
            BuscadorProductos(max: 1, elegidos: Binding(
                get: { link.hasPrefix("prod:") ? [String(link.dropFirst(5))] : [] },
                set: { ids in if let id = ids.first { link = "prod:\(id)" } }
            ))
            .environment(estado).environment(datos)
        }
    }
}

/// Un botón: el texto (hasta 30) y adónde lleva.
struct EditorBoton: View {
    var titulo: String
    @Binding var valor: ValorJSON?

    private func parte(_ k: String) -> Binding<String> {
        Binding(get: { valor?[k]?.texto ?? "" }, set: { v in
            var o = valor?.objeto ?? [:]
            o[k] = v.isEmpty ? nil : .texto(k == "t" ? String(v.prefix(30)) : v)
            valor = o.isEmpty ? nil : .objeto(o)
        })
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            CampoLinea(titulo: "\(titulo): texto", texto: parte("t"), sugerencia: "Ver productos", max: 30)
            if !(valor?["t"]?.texto ?? "").isEmpty {
                SelectorLink(titulo: "Adónde lleva", link: parte("a"))
            }
        }
    }
}

// MARK: - Opciones, fecha e íconos

/// Una opción entre pocas: segmentado si son hasta tres cortas, menú si no.
struct SelectorOpcion: View {
    var titulo: String
    var opciones: [String]
    var nombres: [String]
    @Binding var elegido: String

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text(titulo).estilo(.footnoteFuerte).foregroundStyle(C.texto2)
            if opciones.count <= 3, nombres.allSatisfy({ $0.count <= 14 }) {
                Segmentado(opciones: Array(zip(opciones, nombres)), elegido: $elegido)
            } else {
                Picker(titulo, selection: $elegido) {
                    ForEach(Array(zip(opciones, nombres)), id: \.0) { o, n in Text(n).tag(o) }
                }
                .pickerStyle(.menu).tint(C.primarioTexto)
            }
        }
    }
}

/// Fecha y hora (ISO 8601 con la hora de Argentina).
struct EditorFecha: View {
    var titulo: String
    @Binding var valor: ValorJSON?

    private static let formato: ISO8601DateFormatter = {
        let f = ISO8601DateFormatter()
        f.formatOptions = [.withInternetDateTime]
        f.timeZone = TimeZone(identifier: "America/Argentina/Buenos_Aires")
        return f
    }()

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            if let v = valor?.texto, let f = Fecha.iso(v) {
                DatePicker(titulo, selection: Binding(get: { f }, set: { valor = .texto(Self.formato.string(from: $0)) }))
                    .estilo(.body16).foregroundStyle(C.texto).tint(C.acento)
                if f < Date() { Text("Ya pasó: la oferta no se muestra.").estilo(.footnote).foregroundStyle(C.errorTinta) }
            } else {
                Text(titulo).estilo(.footnoteFuerte).foregroundStyle(C.texto2)
                Button {
                    let enUnaSemana = Calendar.current.date(byAdding: .day, value: 7, to: Date()) ?? Date()
                    valor = .texto(Self.formato.string(from: Calendar.current.startOfDay(for: enUnaSemana).addingTimeInterval(-60)))
                } label: { Label("Elegir cuándo termina", systemImage: "calendar") }
                    .buttonStyle(BotonTonal(alto: 40))
            }
        }
    }
}

/// Los íconos de bloques.json › iconos, con su símbolo y su nombre.
struct SelectorIcono: View {
    var titulo: String
    @Binding var elegido: String

    static let simbolos: [String: (String, String)] = [
        "envio": ("shippingbox", "Envío"), "cuotas": ("creditcard", "Cuotas"), "transferencia": ("building.columns", "Transferencia"),
        "cambios": ("arrow.triangle.2.circlepath", "Cambios"), "retiro": ("storefront", "Retiro"), "seguro": ("lock.shield", "Compra segura"),
        "regalo": ("gift", "Regalo"), "hecho-a-mano": ("hand.raised", "Hecho a mano"), "reloj": ("clock", "Horario"),
        "estrella": ("star", "Destacado")
    ]

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text(titulo).estilo(.footnoteFuerte).foregroundStyle(C.texto2)
            LazyVGrid(columns: Array(repeating: GridItem(.flexible(), spacing: 6), count: 5), spacing: 6) {
                ForEach(CatalogoBloques.shared.iconos, id: \.self) { i in
                    let s = Self.simbolos[i] ?? ("circle", i)
                    Button { elegido = i } label: {
                        VStack(spacing: 3) {
                            Image(systemName: s.0).font(.system(size: 17, weight: .medium))
                            Text(s.1).font(.system(size: 9, weight: .semibold)).lineLimit(1).minimumScaleFactor(0.7)
                        }
                        .foregroundStyle(elegido == i ? C.elevado : C.texto)
                        .frame(maxWidth: .infinity, minHeight: 48)
                        .background(elegido == i ? C.texto : C.superficie, in: RoundedRectangle(cornerRadius: 10, style: .continuous))
                        .overlay(RoundedRectangle(cornerRadius: 10, style: .continuous).strokeBorder(C.linea2))
                    }
                    .buttonStyle(.plain)
                    .accessibilityLabel(s.1)
                    .accessibilityAddTraits(elegido == i ? .isSelected : [])
                }
            }
        }
    }
}

// MARK: - Productos y categorías

/// Uno o varios productos de la tienda, con el buscador del catálogo.
struct SelectorProductos: View {
    var titulo: String
    var min = 0
    var max: Int
    @Binding var ids: [String]
    @Environment(DatosEditor.self) private var datos
    @Environment(Estado.self) private var estado
    @State private var buscando = false

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Text(titulo).estilo(.footnoteFuerte).foregroundStyle(C.texto2)
                Spacer()
                if max > 1 { Text(min > 0 ? "\(ids.count) · de \(min) a \(max)" : "\(ids.count) de \(max)").estilo(.footnote).foregroundStyle(C.texto3) }
            }
            ForEach(ids, id: \.self) { id in
                HStack(spacing: 10) {
                    Miniatura(foto: datos.conocidos[id]?.foto, tam: 40, radio: 10)
                    Text(datos.nombre(id)).estilo(.body16).foregroundStyle(C.texto).lineLimit(1)
                    Spacer()
                    Button { ids.removeAll { $0 == id } } label: { Image(systemName: "xmark.circle.fill").foregroundStyle(C.texto3) }
                        .accessibilityLabel("Quitar \(datos.nombre(id))")
                }
            }
            if ids.count < max || max == 1 {
                Button { buscando = true } label: {
                    Label(ids.isEmpty ? (max == 1 ? "Elegir producto" : "Elegir productos") : (max == 1 ? "Cambiar" : "Sumar productos"),
                          systemImage: "magnifyingglass")
                }
                .buttonStyle(BotonTonal(alto: 40))
            }
        }
        .sheet(isPresented: $buscando) { BuscadorProductos(max: max, elegidos: $ids).environment(estado).environment(datos) }
    }
}

/// El catálogo de la tienda para elegir: buscador, foto, nombre y precio.
struct BuscadorProductos: View {
    var max: Int
    @Binding var elegidos: [String]
    @Environment(Estado.self) private var estado
    @Environment(DatosEditor.self) private var datos
    @Environment(\.dismiss) private var cerrar
    @State private var texto = ""
    @State private var lista: [ProductoFila] = []
    @State private var cargando = true

    var body: some View {
        NavigationStack {
            VStack(spacing: 12) {
                Buscador(placeholder: "Buscá por nombre", texto: $texto).padding(.horizontal, 20)
                if !cargando, lista.isEmpty {
                    Text(texto.isEmpty ? "Todavía no cargaste productos." : "No encontramos «\(texto)».")
                        .estilo(.body16).foregroundStyle(C.texto2).frame(maxHeight: .infinity)
                } else {
                    ScrollView {
                        LazyVStack(spacing: 0) {
                            ForEach(lista) { p in
                                Button { tocar(p) } label: {
                                    HStack(spacing: 12) {
                                        Miniatura(foto: p.foto, tam: 48, radio: 12)
                                        VStack(alignment: .leading, spacing: 2) {
                                            Text(p.nombre).estilo(.fila).foregroundStyle(C.texto).lineLimit(2)
                                            Text(Formato.pesos(p.precio)).estilo(.texto14).monospacedDigit().foregroundStyle(C.texto2)
                                        }
                                        Spacer()
                                        if elegidos.contains(p.id) {
                                            Image(systemName: "checkmark.circle.fill").font(.system(size: 22)).foregroundStyle(C.ok)
                                        }
                                    }
                                    .padding(.vertical, 8).padding(.horizontal, 20)
                                    .contentShape(Rectangle())
                                }
                                .buttonStyle(.plain)
                                Separador().padding(.leading, 80)
                            }
                        }
                    }
                }
            }
            .padding(.top, 8)
            .background(C.fondo.ignoresSafeArea())
            .navigationTitle(max == 1 ? "Elegí un producto" : "Elegí hasta \(max)")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar { ToolbarItem(placement: .confirmationAction) { Button("Listo") { cerrar() } } }
            .task(id: texto) {
                if !texto.isEmpty { try? await Task.sleep(for: .milliseconds(300)) }
                if Task.isCancelled { return }
                lista = await datos.buscar(texto, estado.api)
                cargando = false
            }
        }
    }

    private func tocar(_ p: ProductoFila) {
        Haptico.seleccion()
        datos.conocidos[p.id] = p
        if max == 1 { elegidos = [p.id]; cerrar(); return }
        if let i = elegidos.firstIndex(of: p.id) { elegidos.remove(at: i) } else if elegidos.count < max { elegidos.append(p.id) }
    }
}

/// Qué productos muestra un bloque: más vendidos, novedades, ofertas, elegidos o de una categoría.
struct EditorFuenteProductos: View {
    var campo: CatalogoBloques.Campo
    @Binding var valor: ValorJSON?
    private let fuentes = CatalogoBloques.shared.fuentesProductos

    private var actual: [String: ValorJSON] { valor?.objeto ?? campo.porDefecto?.objeto ?? ["fuente": .texto("vendidos"), "cantidad": .numero(8)] }

    private func poner(_ k: String, _ v: ValorJSON?) {
        var o = actual
        o[k] = v
        valor = .objeto(o)
    }

    var body: some View {
        let fuente = actual["fuente"]?.texto ?? "vendidos"
        let cantidad = actual["cantidad"]?.entero ?? 8
        VStack(alignment: .leading, spacing: 12) {
            SelectorOpcion(titulo: campo.nombre, opciones: fuentes.opciones, nombres: fuentes.nombres,
                           elegido: Binding(get: { fuente }, set: { poner("fuente", .texto($0)) }))
            if fuente == "elegidos" {
                SelectorProductos(titulo: "Los productos", max: 24, ids: Binding(
                    get: { actual["ids"]?.lista?.compactMap(\.texto) ?? [] },
                    set: { ids in poner("ids", ids.isEmpty ? nil : .lista(ids.map { .texto($0) })) }))
            } else {
                if fuente == "categoria" {
                    SelectorCategoria(titulo: "La categoría", elegida: Binding(
                        get: { actual["categoria"]?.texto ?? "" }, set: { poner("categoria", $0.isEmpty ? nil : .texto($0)) }))
                }
                Stepper(value: Binding(get: { cantidad }, set: { poner("cantidad", .numero(Double($0))) }), in: 1...24) {
                    HStack {
                        Text("Cuántos").estilo(.body16).foregroundStyle(C.texto)
                        Spacer()
                        Text("\(cantidad)").estilo(.fila).monospacedDigit().foregroundStyle(C.texto)
                    }
                }
            }
        }
    }
}

struct SelectorCategoria: View {
    var titulo: String
    @Binding var elegida: String
    @Environment(DatosEditor.self) private var datos

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text(titulo).estilo(.footnoteFuerte).foregroundStyle(C.texto2)
            if datos.contexto.categorias.isEmpty {
                Text("Todavía no hay categorías: ponéselas a tus productos.").estilo(.footnote).foregroundStyle(C.texto2)
            } else {
                Picker(titulo, selection: $elegida) {
                    Text("Elegí una").tag("")
                    ForEach(datos.contexto.categorias, id: \.self) { Text($0).tag($0) }
                }
                .pickerStyle(.menu).tint(C.primarioTexto)
            }
        }
    }
}

/// Varias categorías (vacío: todas), como chips que se prenden.
struct SelectorCategorias: View {
    var titulo: String
    @Binding var elegidas: [String]
    @Environment(DatosEditor.self) private var datos

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(titulo).estilo(.footnoteFuerte).foregroundStyle(C.texto2)
            if datos.contexto.categorias.isEmpty {
                Text("Todavía no hay categorías: ponéselas a tus productos.").estilo(.footnote).foregroundStyle(C.texto2)
            }
            ScrollView(.horizontal, showsIndicators: false) {
                HStack(spacing: 6) {
                    ForEach(datos.contexto.categorias, id: \.self) { c in
                        ChipFiltro(texto: c, activo: elegidas.contains(c)) {
                            if let i = elegidas.firstIndex(of: c) { elegidas.remove(at: i) } else { elegidas.append(c) }
                        }
                    }
                }
            }
        }
    }
}

// MARK: - Listas

/// Una lista de elementos con los campos de `item` (beneficios, preguntas, horarios, logos…).
struct EditorLista: View {
    var campo: CatalogoBloques.Campo
    @Binding var valor: ValorJSON?

    private var items: [[String: ValorJSON]] { valor?.lista?.map { $0.objeto ?? [:] } ?? [] }
    private var max: Int { campo.max ?? CatalogoBloques.shared.limites.itemsPorLista }

    private func poner(_ l: [[String: ValorJSON]]) { valor = l.isEmpty ? nil : .lista(l.map { .objeto($0) }) }

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack {
                Text(campo.nombre).estilo(.subheadFuerte).foregroundStyle(C.texto)
                Spacer()
                Text("\(items.count) de \(max)").estilo(.footnote).monospacedDigit().foregroundStyle(C.texto3)
            }
            ForEach(Array(items.enumerated()), id: \.offset) { i, item in
                VStack(alignment: .leading, spacing: 12) {
                    HStack(spacing: 14) {
                        Text("\(i + 1)").estilo(.footnoteFuerte).foregroundStyle(C.texto2)
                        Spacer()
                        Button { var l = items; l.swapAt(i, i - 1); poner(l) } label: { Image(systemName: "chevron.up") }
                            .disabled(i == 0).accessibilityLabel("Subir")
                        Button { var l = items; l.swapAt(i, i + 1); poner(l) } label: { Image(systemName: "chevron.down") }
                            .disabled(i == items.count - 1).accessibilityLabel("Bajar")
                        Button { var l = items; l.remove(at: i); poner(l) } label: { Image(systemName: "trash") }
                            .accessibilityLabel("Quitar")
                    }
                    .font(.system(size: 14, weight: .semibold)).foregroundStyle(C.primarioTexto).buttonStyle(.plain)
                    EditorCampos(campos: campo.item ?? [], datos: Binding(
                        get: { items.indices.contains(i) ? items[i] : [:] },
                        set: { v in var l = items; if l.indices.contains(i) { l[i] = v; poner(l) } }))
                }
                .padding(12)
                .background(C.crema2.opacity(0.5), in: RoundedRectangle(cornerRadius: 14, style: .continuous))
                .overlay(RoundedRectangle(cornerRadius: 14, style: .continuous).strokeBorder(C.linea))
            }
            if items.isEmpty {
                Text("Todavía no hay ninguno.").estilo(.footnote).foregroundStyle(C.texto2)
            }
            if items.count < max {
                Button { poner(items + [[:]]) } label: { Label("Sumar", systemImage: "plus") }
                    .buttonStyle(BotonTonal(alto: 40))
            }
        }
    }
}
