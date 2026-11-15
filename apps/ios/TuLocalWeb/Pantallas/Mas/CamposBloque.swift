import PhotosUI
import SwiftUI
import UniformTypeIdentifiers

// Los campos del panel de un bloque, generados desde bloques.json › campos (tiposCampo): un editor por tipo.
// Lo vacío se saca de los datos, así la tienda usa lo de por defecto o lo sugerido.

/// Lo que los campos necesitan de la tienda: productos para buscar, categorías y el contexto de los avisos.
@Observable
final class DatosEditor {
    var contexto = ContextoTienda()
    /// Los primeros productos (para elegir sin buscar) y todos los que se fueron viendo, por id.
    var productos: [ProductoFila] = []
    var conocidos: [String: ProductoFila] = [:]

    /// WhatsApp y dirección del local, para los avisos. Se vuelve a leer al volver de Datos del negocio.
    func cargarNegocio(_ api: Api) async {
        guard let c: ValorJSON = try? await api.get("/v1/panel/configuracion"), let n = c["negocio"] else { return }
        contexto.whatsapp = !(n["whatsapp"]?.texto ?? "").trimmingCharacters(in: .whitespaces).isEmpty
        contexto.direccionLocal = !(n["direccion"]?.texto ?? "").trimmingCharacters(in: .whitespaces).isEmpty
    }

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

    // MARK: Promociones

    /// Un cupón de la tienda, para elegirlo en una ventana.
    struct CuponTienda: Equatable, Identifiable { var codigo: String; var descripcion: String; var activo: Bool; var id: String { codigo } }
    var cupones: [CuponTienda] = []
    /// Desde cuánto el envío es gratis (Cupones › automáticas), para las plantillas.
    var envioGratis: Int?
    /// Por ventana, los últimos 30 días.
    var metricas: [String: MetricaVentana] = [:]
    var metricasCargadas = false

    func cargarPromos(_ api: Api) async {
        let cs: ValorJSON? = try? await api.get("/v1/panel/cupones")
        cupones = (cs?["cupones"]?.lista ?? []).compactMap { x in
            guard let codigo = x["codigo"]?.texto, x["estado"]?.texto != "terminado" else { return nil }
            return CuponTienda(codigo: codigo, descripcion: x["descripcion"]?.texto ?? "", activo: x["estado"]?.texto == "activo")
        }
        let en: ValorJSON? = try? await api.get("/v1/panel/configuracion/entregas")
        envioGratis = en?["gratisDesde"]?.entero
    }

    func cargarMetricas(_ api: Api) async {
        let hoy = Date()
        let desde = Calendar.current.date(byAdding: .day, value: -29, to: hoy) ?? hoy
        let q = [URLQueryItem(name: "desde", value: Self.dia(desde)), URLQueryItem(name: "hasta", value: Self.dia(hoy))]
        if let v: ValorJSON = try? await api.get("/v1/panel/promos/metricas", query: q) { metricas = MetricaVentana.leer(v) }
        metricasCargadas = true
    }

    private static func dia(_ f: Date) -> String { String(VentanaTienda.isoAR(f).prefix(10)) }
}

/// Todos los campos de un bloque (o de un elemento de una lista).
struct EditorCampos: View {
    var campos: [CatalogoBloques.Campo]
    @Binding var datos: [String: ValorJSON]

    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            ForEach(campos) { c in
                VStack(alignment: .leading, spacing: 6) {
                    EditorCampo(campo: c, valor: Binding(
                        get: { datos[c.id] },
                        set: { v in
                            if let v, !v.vacio || v.bool != nil { datos[c.id] = v } else { datos[c.id] = nil }
                        }
                    ))
                    if let d = c.descripcion, c.tipo != "grupo" {
                        Text(d).estilo(.footnote).foregroundStyle(C.texto2).fixedSize(horizontal: false, vertical: true)
                    }
                }
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
                    if let d = campo.descripcion { Text(d).estilo(.footnote).foregroundStyle(C.texto2).fixedSize(horizontal: false, vertical: true) }
                    EditorCampos(campos: camposDelGrupo, datos: Binding(
                        get: { valor?.objeto ?? [:] }, set: { valor = .objeto($0) }))
                }
                .padding(12)
                .overlay(RoundedRectangle(cornerRadius: 14, style: .continuous).strokeBorder(C.linea2))
            )
        case "tabla":
            AnyView(EditorTabla(titulo: campo.nombre, valor: $valor))
        case "cupon":
            AnyView(SelectorCupon(titulo: campo.nombre, codigo: texto))
        case "premio":
            AnyView(EditorPremio(campo: campo, valor: $valor))
        case "opciones":
            AnyView(SelectorOpciones(titulo: campo.nombre, opciones: campo.opciones ?? [], nombres: campo.nombres ?? campo.opciones ?? [],
                                     elegidas: Binding(get: { valor?.lista?.compactMap(\.texto) ?? [] }, set: { valor = .lista($0.map { .texto($0) }) })))
        default:
            EmptyView()
        }
    }

    /// Un grupo que empieza con «activa» apagada muestra sólo el interruptor (compras recientes).
    private var camposDelGrupo: [CatalogoBloques.Campo] {
        let items = campo.item ?? []
        guard let p = items.first, p.id == "activa", p.tipo == "si" else { return items }
        let on = valor?["activa"]?.bool ?? campo.porDefecto?["activa"]?.bool ?? p.porDefecto?.bool ?? false
        return on ? items : [p]
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
    var boton = "Elegir cuándo termina"
    var pasada = "Ya pasó: no se muestra."
    /// Con «Quitar» para dejarla sin fecha.
    var quitable = false

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
                HStack {
                    if f < Date() { Text(pasada).estilo(.footnote).foregroundStyle(C.errorTinta) }
                    Spacer()
                    if quitable { Button("Quitar") { valor = nil }.estilo(.footnoteFuerte).foregroundStyle(C.primarioTexto) }
                }
            } else {
                Text(titulo).estilo(.footnoteFuerte).foregroundStyle(C.texto2)
                Button {
                    let enUnaSemana = Calendar.current.date(byAdding: .day, value: 7, to: Date()) ?? Date()
                    valor = .texto(Self.formato.string(from: Calendar.current.startOfDay(for: enUnaSemana).addingTimeInterval(-60)))
                } label: { Label(boton, systemImage: "calendar") }
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

    /// Un elemento nuevo con los números por defecto ya puestos (el peso de una porción cuenta aunque no se toque).
    private var nuevo: [String: ValorJSON] {
        var o: [String: ValorJSON] = [:]
        for c in campo.item ?? [] where c.tipo == "numero" { if let d = c.porDefecto { o[c.id] = d } }
        return o
    }

    /// Con peso (las porciones de la ruleta): cuántas veces de cada 100 sale cada una.
    private func chance(_ i: Int) -> String? {
        guard (campo.item ?? []).contains(where: { $0.id == "peso" }) else { return nil }
        let pesos = items.map { $0["peso"]?.numero ?? 0 }
        let total = pesos.reduce(0, +)
        guard total > 0, pesos.indices.contains(i), pesos[i] > 0 else { return "Nunca sale" }
        return "Sale \(Int((pesos[i] / total * 100).rounded())) de cada 100"
    }

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
                        if let c = chance(i) { Text(c).estilo(.footnote).monospacedDigit().foregroundStyle(C.texto3).lineLimit(1) }
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
                Button { poner(items + [nuevo]) } label: { Label("Sumar", systemImage: "plus") }
                    .buttonStyle(BotonTonal(alto: 40))
            }
        }
    }
}

// MARK: - Tabla

/// Una tabla (tiposCampo › tabla; la guía de talles): filas × columnas de textos cortos, la primera fila es el
/// encabezado. Siempre queda pareja, con el encabezado y al menos una fila, dentro de los topes del catálogo.
struct EditorTabla: View {
    var titulo: String
    @Binding var valor: ValorJSON?
    private let lim = CatalogoBloques.shared.limites

    private var filas: [[String]] { valor?.lista?.map { $0.lista?.map { $0.texto ?? "" } ?? [] } ?? [] }
    private var ancho: Int { filas.first?.count ?? 0 }

    private func poner(_ f: [[String]]?) { valor = f.map { .lista($0.map { .lista($0.map { .texto($0) }) }) } }

    /// Una celda: en una línea, sin < ni >, hasta `letrasCelda` letras (como las cuenta Go).
    private func limpia(_ s: String) -> String {
        let t = s.filter { $0 != "<" && $0 != ">" && !$0.isNewline }
        return String(String.UnicodeScalarView(t.unicodeScalars.prefix(lim.letrasCelda ?? 30)))
    }

    private func celda(_ i: Int, _ j: Int) -> Binding<String> {
        Binding(get: { filas.indices.contains(i) && filas[i].indices.contains(j) ? filas[i][j] : "" }, set: { v in
            var f = filas
            guard f.indices.contains(i), f[i].indices.contains(j) else { return }
            f[i][j] = limpia(v)
            poner(f)
        })
    }

    private func pista(_ i: Int, _ j: Int) -> String {
        i == 0 ? (j == 0 ? "Talle" : "Medida") : (j == 0 ? "S" : "")
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            Text(titulo).estilo(.footnoteFuerte).foregroundStyle(C.texto2)
            if ancho == 0 {
                Button { poner([["Talle", ""], ["", ""]]) } label: { Label("Armar la tabla", systemImage: "square.grid.2x2") }
                    .buttonStyle(BotonTonal(alto: 40))
            } else {
                ScrollView(.horizontal, showsIndicators: false) {
                    VStack(alignment: .leading, spacing: 6) {
                        ForEach(filas.indices, id: \.self) { i in
                            HStack(spacing: 6) {
                                ForEach(0..<ancho, id: \.self) { j in
                                    TextField("", text: celda(i, j), prompt: Text(pista(i, j)).foregroundStyle(C.texto3))
                                        .estilo(i == 0 ? .footnoteFuerte : .body16).foregroundStyle(C.texto)
                                        .accessibilityLabel(i == 0 ? "Encabezado, columna \(j + 1)" : "Fila \(i), columna \(j + 1)")
                                        .padding(.horizontal, 10).frame(width: 92, height: 44)
                                        .background(i == 0 ? C.crema2 : C.superficie, in: RoundedRectangle(cornerRadius: 12, style: .continuous))
                                        .overlay(RoundedRectangle(cornerRadius: 12, style: .continuous).strokeBorder(C.linea2))
                                }
                                if i > 0 {
                                    Button { var f = filas; f.remove(at: i); poner(f) } label: { Image(systemName: "trash") }
                                        .font(.system(size: 14, weight: .semibold)).foregroundStyle(C.primarioTexto).buttonStyle(.plain)
                                        .frame(width: 44, height: 44).disabled(filas.count <= 2).opacity(filas.count <= 2 ? 0.35 : 1)
                                        .accessibilityLabel("Quitar la fila \(i)")
                                }
                            }
                        }
                    }
                }
                Text("La primera fila es el encabezado. Hasta \(lim.filasTabla ?? 8) filas y \(lim.columnasTabla ?? 8) columnas.")
                    .estilo(.footnote).foregroundStyle(C.texto2)
                ScrollView(.horizontal, showsIndicators: false) {
                    HStack(spacing: 8) {
                        Button { poner(filas + [Array(repeating: "", count: ancho)]) } label: { Label("Fila", systemImage: "plus") }
                            .disabled(filas.count >= (lim.filasTabla ?? 8)).accessibilityLabel("Sumar una fila")
                        Button { poner(filas.map { $0 + [""] }) } label: { Label("Columna", systemImage: "plus") }
                            .disabled(ancho >= (lim.columnasTabla ?? 8)).accessibilityLabel("Sumar una columna")
                        Button { poner(filas.map { Array($0.dropLast()) }) } label: { Label("Columna", systemImage: "minus") }
                            .disabled(ancho <= 1).accessibilityLabel("Quitar la última columna")
                        Button { poner(nil) } label: { Label("Borrar la tabla", systemImage: "trash") }
                    }
                    .buttonStyle(BotonTonal(alto: 40))
                }
            }
        }
    }
}
// MARK: - Promociones: cupón, premio y opciones

/// Un cupón de la tienda (tiposCampo › cupon): se elige de la lista de Cupones.
struct SelectorCupon: View {
    var titulo: String
    @Binding var codigo: String
    @Environment(DatosEditor.self) private var datos

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text(titulo).estilo(.footnoteFuerte).foregroundStyle(C.texto2)
            if datos.cupones.isEmpty && codigo.isEmpty {
                Text("Todavía no tenés cupones. Crealos en Más › Cupones.").estilo(.footnote).foregroundStyle(C.avisoTinta)
            } else {
                Picker(titulo, selection: $codigo) {
                    Text("Elegí uno").tag("")
                    ForEach(datos.cupones) { c in
                        Text(c.descripcion.isEmpty ? c.codigo : "\(c.codigo) · \(c.descripcion)").tag(c.codigo)
                    }
                    if !codigo.isEmpty, !datos.cupones.contains(where: { $0.codigo == codigo }) { Text(codigo).tag(codigo) }
                }
                .pickerStyle(.menu).tint(C.primarioTexto)
                if let c = datos.cupones.first(where: { $0.codigo == codigo }), !c.activo {
                    Text("Este cupón está pausado: activalo en Cupones para que funcione.").estilo(.footnote).foregroundStyle(C.avisoTinta)
                }
            }
        }
    }
}

/// El premio de una captación o de una porción de la ruleta (tiposCampo › premio).
struct EditorPremio: View {
    var campo: CatalogoBloques.Campo
    @Binding var valor: ValorJSON?
    @Environment(Estado.self) private var estado
    @State private var eligiendo = false
    @State private var subiendo = false

    static let nombres = ["cuponFijo": "Un cupón tuyo", "cuponUnico": "Un código único", "descarga": "Una descarga", "nada": "Nada, sólo sumarse"]
    static let descuentos = [("porcentaje", "Porcentaje"), ("monto", "Monto fijo"), ("envioGratis", "Envío gratis")]

    private var tipos: [String] { campo.opciones ?? ["cuponFijo", "cuponUnico", "descarga", "nada"] }
    private var actual: [String: ValorJSON] { valor?.objeto ?? campo.porDefecto?.objeto ?? [:] }
    private var tipo: String { actual["tipo"]?.texto ?? "" }

    private func poner(_ k: String, _ v: ValorJSON?) {
        var o = actual
        o[k] = v
        valor = .objeto(o)
    }

    /// Al cambiar de tipo quedan sólo las claves del nuevo (Go ignora las otras, pero así no confunden).
    private func cambiarTipo(_ t: String) {
        var o: [String: ValorJSON] = ["tipo": .texto(t)]
        if t == "cuponUnico" { o["descuento"] = .texto("porcentaje"); o["valor"] = .numero(10); o["dias"] = .numero(7) }
        if t == "descarga", let n = actual["nombreArchivo"] { o["nombreArchivo"] = n }
        valor = .objeto(o)
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            SelectorOpcion(titulo: campo.nombre, opciones: tipos, nombres: tipos.map { Self.nombres[$0] ?? $0 },
                           elegido: Binding(get: { tipo }, set: { cambiarTipo($0) }))
            detalle
        }
        .padding(12)
        .overlay(RoundedRectangle(cornerRadius: 14, style: .continuous).strokeBorder(C.linea2))
        .fileImporter(isPresented: $eligiendo, allowedContentTypes: [.pdf, .epub, .zip]) { r in
            if case let .success(u) = r { Task { await subir(u) } }
        }
    }

    /// Si el archivo es uno subido a Tu Local Web (y no un link de afuera).
    private static func esSubido(_ a: String) -> Bool { a.contains("/f/archivos/") }

    /// Sube el archivo elegido y completa `archivo` (y `nombreArchivo` si estaba vacío).
    private func subir(_ u: URL) async {
        let acceso = u.startAccessingSecurityScopedResource()
        defer { if acceso { u.stopAccessingSecurityScopedResource() } }
        guard let d = try? Data(contentsOf: u) else { estado.avisar("No pudimos leer el archivo", ok: false); return }
        guard d.count <= Web.maximoArchivo else { estado.avisar("El archivo pesa más de 20 MB", ok: false); return }
        subiendo = true
        defer { subiendo = false }
        do {
            let url = try await Web.subirArchivo(d, token: estado.api.token)
            var o = actual
            o["archivo"] = .texto(url)
            if (o["nombreArchivo"]?.texto ?? "").isEmpty { o["nombreArchivo"] = .texto(Web.nombreDeArchivo(u.lastPathComponent)) }
            valor = .objeto(o)
            Haptico.exito()
        } catch let e as Web.ErrorSubida {
            estado.avisar(e.mensaje, ok: false)
        } catch {
            estado.avisar("No pudimos subir el archivo", ok: false)
        }
    }

    @ViewBuilder private var detalle: some View {
        switch tipo {
        case "cuponFijo":
            SelectorCupon(titulo: "El cupón", codigo: Binding(get: { actual["cupon"]?.texto ?? "" }, set: { poner("cupon", $0.isEmpty ? nil : .texto($0)) }))
        case "cuponUnico":
            let d = actual["descuento"]?.texto ?? "porcentaje"
            SelectorOpcion(titulo: "Qué descuento", opciones: Self.descuentos.map(\.0), nombres: Self.descuentos.map(\.1),
                           elegido: Binding(get: { d }, set: { n in
                               var o = actual
                               o["descuento"] = .texto(n)
                               o["valor"] = n == "envioGratis" ? nil : .numero(n == "monto" ? 2000 : 10)
                               valor = .objeto(o)
                           }))
            if d == "porcentaje" {
                numero("Porcentaje", clave: "valor", rango: 1...90, paso: 5, porDefecto: 10) { "\($0) %" }
            } else if d == "monto" {
                numero("Monto", clave: "valor", rango: 100...100_000, paso: 500, porDefecto: 2000) { Formato.pesos($0) }
            }
            numero("Vence a los", clave: "dias", rango: 1...90, paso: 1, porDefecto: 7) { $0 == 1 ? "1 día" : "\($0) días" }
            Text("Cada persona recibe su propio código, de un solo uso.").estilo(.footnote).foregroundStyle(C.texto2)
        case "descarga":
            let a = actual["archivo"]?.texto ?? ""
            let nombre = actual["nombreArchivo"]?.texto ?? ""
            if Self.esSubido(a) {
                HStack(spacing: 10) {
                    Image(systemName: "doc").font(.system(size: 18)).foregroundStyle(C.primarioTinta)
                    Text(nombre.isEmpty ? "Tu archivo" : nombre)
                        .estilo(.fila).foregroundStyle(C.texto).lineLimit(1)
                    Spacer(minLength: 0)
                    Button { poner("archivo", nil) } label: {
                        Image(systemName: "trash").font(.system(size: 15)).foregroundStyle(C.texto2).frame(width: 44, height: 44)
                    }
                    .accessibilityLabel("Quitar el archivo")
                }
                .padding(.leading, 12)
                .overlay(RoundedRectangle(cornerRadius: 12, style: .continuous).strokeBorder(C.linea2))
            }
            HStack(spacing: 10) {
                Button { eligiendo = true } label: {
                    HStack(spacing: 6) {
                        if subiendo { ProgressView().tint(C.primarioTinta) } else { Image(systemName: "arrow.up") }
                        Text(subiendo ? "Subiendo…" : Self.esSubido(a) ? "Cambiar el archivo" : "Subir archivo")
                    }
                }
                .buttonStyle(BotonTonal(alto: 44))
                .disabled(subiendo)
                Text("PDF, EPUB o ZIP de hasta 20 MB.").estilo(.footnote).foregroundStyle(C.texto2)
            }
            if !Self.esSubido(a) {
                CampoLinea(titulo: "O pegá un link", texto: Binding(
                    get: { actual["archivo"]?.texto ?? "" },
                    set: { v in let t = v.trimmingCharacters(in: .whitespaces); poner("archivo", t.isEmpty ? nil : .texto(String(t.prefix(300)))) }),
                           sugerencia: "https://", teclado: .URL)
                if !a.isEmpty, !a.hasPrefix("https://") {
                    Text("Tiene que empezar con https://").estilo(.footnote).foregroundStyle(C.errorTinta)
                }
            }
            CampoLinea(titulo: "Cómo se llama", texto: Binding(
                get: { actual["nombreArchivo"]?.texto ?? "" }, set: { poner("nombreArchivo", $0.isEmpty ? nil : .texto(String($0.prefix(60)))) }),
                       sugerencia: "Guía", max: 60)
        case "nada":
            Text("Sólo se suma a tu newsletter.").estilo(.footnote).foregroundStyle(C.texto2)
        default:
            EmptyView()
        }
    }

    private func numero(_ titulo: String, clave: String, rango: ClosedRange<Int>, paso: Int, porDefecto: Int, _ texto: @escaping (Int) -> String) -> some View {
        let n = actual[clave]?.entero ?? porDefecto
        return Stepper(value: Binding(get: { n }, set: { poner(clave, .numero(Double(min(max($0, rango.lowerBound), rango.upperBound)))) }),
                       in: rango, step: paso) {
            HStack {
                Text(titulo).estilo(.body16).foregroundStyle(C.texto)
                Spacer()
                Text(texto(n)).estilo(.fila).monospacedDigit().foregroundStyle(C.texto)
            }
        }
    }
}

/// Varias opciones de una lista (tiposCampo › opciones; vacía = todas), como chips que se prenden.
struct SelectorOpciones: View {
    var titulo: String
    var opciones: [String]
    var nombres: [String]
    @Binding var elegidas: [String]

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Text(titulo).estilo(.footnoteFuerte).foregroundStyle(C.texto2)
                Spacer()
                if elegidas.isEmpty { Text("Todas").estilo(.footnote).foregroundStyle(C.texto3) }
            }
            ScrollView(.horizontal, showsIndicators: false) {
                HStack(spacing: 6) {
                    ForEach(Array(opciones.enumerated()), id: \.element) { i, o in
                        ChipFiltro(texto: nombres.indices.contains(i) ? nombres[i] : o, activo: elegidas.contains(o)) {
                            Haptico.seleccion()
                            // En el orden del catálogo, sin repetir.
                            let n = elegidas.contains(o) ? elegidas.filter { $0 != o } : elegidas + [o]
                            elegidas = opciones.filter { n.contains($0) }
                        }
                    }
                }
            }
        }
    }
}
