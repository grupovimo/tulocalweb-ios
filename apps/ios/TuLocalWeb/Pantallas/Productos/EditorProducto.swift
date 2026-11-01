import PhotosUI
import SwiftUI

/// 11c · Cargar o editar un producto: fotos (cámara o galería), nombre, precio, costo, stock, SKU,
/// categoría y descripción. Edita el producto entero que da la API, así que lo que la app no muestra
/// (opciones, variantes, canales) vuelve tal cual.
struct EditorProducto: View {
    var id: String?
    var sku: String? = nil
    @Environment(Estado.self) private var estado
    @Environment(\.dismiss) private var cerrar
    @State private var original: [String: ValorJSON] = [:]
    @State private var cargando = true
    @State private var guardando = false
    @State private var nombre = ""
    @State private var precio = ""
    @State private var costo = ""
    @State private var stock = ""
    @State private var codigo = ""
    @State private var categoria = ""
    @State private var descripcion = ""
    @State private var fotos: [String] = []
    @State private var subiendo = 0
    @State private var conVariantes = false
    @State private var elegidas: [PhotosPickerItem] = []
    @State private var camara = false
    @State private var borrar = false
    @State private var opciones: [Opcion] = []
    @State private var variantes: [String: Combinacion] = [:]
    @State private var teniaOpciones = false
    @State private var nuevoValor: [Int: String] = [:]

    struct Opcion: Equatable { var nombre: String; var valores: [String] }
    struct Combinacion: Equatable { var stock = ""; var precio = ""; var sku = "" }
    static let curvas: [(String, [String])] = [
        ("Talle (letras)", ["XS", "S", "M", "L", "XL", "XXL"]), ("Talle (números)", ["36", "38", "40", "42", "44", "46"]),
        ("Talle (calzado)", ["35", "36", "37", "38", "39", "40", "41", "42", "43", "44"]), ("Color", []), ("Presentación", []),
    ]

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 14) {
                Text(id == nil ? "Producto nuevo" : "Editar producto").estilo(.tituloGrande).foregroundStyle(C.texto)
                if cargando {
                    EsqueletoFilas(filas: 4, miniatura: 0).padding(.top, 12)
                } else {
                    seccionFotos
                    campo("Nombre") { TextField("Ej.: Remera de algodón", text: $nombre) }
                    HStack(spacing: 10) {
                        campo("Precio") { TextField("0", text: $precio).keyboardType(.numberPad) }
                        campo("Costo (opcional)") { TextField("0", text: $costo).keyboardType(.numberPad) }
                    }
                    if let m = margen {
                        Text(m).estilo(.footnoteFuerte).foregroundStyle(C.okTinta).padding(.horizontal, 4)
                    }
                    if opciones.isEmpty {
                        campo("Stock") { TextField("0", text: $stock).keyboardType(.numberPad) }
                    }
                    seccionVariantes
                    HStack(spacing: 10) {
                        campo("SKU o código") { TextField("Opcional", text: $codigo).textInputAutocapitalization(.characters) }
                        campo("Categoría") { TextField("Opcional", text: $categoria) }
                    }
                    campo("Descripción") {
                        TextField("Contá cómo es, de qué está hecho, medidas…", text: $descripcion, axis: .vertical)
                            .lineLimit(3...8)
                    }
                    if id != nil {
                        Button("Borrar producto") { borrar = true }
                            .buttonStyle(BotonTexto(color: C.errorTinta))
                    }
                }
            }
            .margenPantalla()
            .padding(.bottom, 32)
        }
        .scrollDismissesKeyboard(.interactively)
        .barraAccion {
            Button(guardando ? "Guardando…" : (id == nil ? "Cargar producto" : "Guardar cambios")) {
                Task { await guardar() }
            }
            .buttonStyle(BotonPrincipal())
            .disabled(!valido || guardando || subiendo > 0)
        }
        .background(C.fondo.ignoresSafeArea())
        .navigationBarTitleDisplayMode(.inline)
        .task { await cargar() }
        .onChange(of: elegidas) { _, items in
            guard !items.isEmpty else { return }
            Task {
                for item in items {
                    if let d = try? await item.loadTransferable(type: Data.self), let img = UIImage(data: d) { await subir(img) }
                }
                elegidas = []
            }
        }
        .fullScreenCover(isPresented: $camara) {
            Camara { img in Task { await subir(img) } }.ignoresSafeArea()
        }
        .confirmationDialog("¿Borrás «\(nombre)»?", isPresented: $borrar, titleVisibility: .visible) {
            Button("Borrar", role: .destructive) { Task { await eliminar() } }
        } message: {
            Text("Deja de verse en la tienda. Los pedidos que ya tiene no cambian.")
        }
    }

    private var margen: String? {
        let p = Int(precio.filter(\.isNumber)) ?? 0, c = Int(costo.filter(\.isNumber)) ?? 0
        guard p > 0, c > 0, p > c else { return nil }
        return "Ganás \(Formato.pesos(p - c)) por unidad (\(Int((Double(p - c) / Double(p) * 100).rounded())) %)"
    }

    /// Las combinaciones de todas las opciones, en orden («M · Rojo»).
    private var combinaciones: [[String: String]] {
        let validas = opciones.filter { !$0.nombre.trimmingCharacters(in: .whitespaces).isEmpty && !$0.valores.isEmpty }
        guard !validas.isEmpty else { return [] }
        var out: [[String: String]] = [[:]]
        for o in validas {
            out = out.flatMap { base in o.valores.map { v in var b = base; b[o.nombre] = v; return b } }
        }
        return Array(out.prefix(100))
    }

    private func clave(_ c: [String: String]) -> String {
        opciones.map { c[$0.nombre] ?? "" }.joined(separator: " · ")
    }

    @ViewBuilder
    private var seccionVariantes: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Text("Talles, colores y más").estilo(.headline).foregroundStyle(C.texto)
                Spacer()
                if opciones.count < 3 {
                    Menu {
                        ForEach(EditorProducto.curvas, id: \.0) { c in
                            Button(c.0) {
                                let nombre = c.0.hasPrefix("Talle") ? "Talle" : c.0
                                if !opciones.contains(where: { $0.nombre == nombre }) { opciones.append(Opcion(nombre: nombre, valores: c.1)) }
                            }
                        }
                        Button("Otra opción") { opciones.append(Opcion(nombre: "", valores: [])) }
                    } label: {
                        Label("Agregar", systemImage: "plus").estilo(.subheadFuerte).foregroundStyle(C.primarioTexto)
                    }
                }
            }
            if opciones.isEmpty {
                Text(Formato.sinViudas("¿Viene en talles o colores? Agregá las opciones y cargás el stock de cada combinación."))
                    .estilo(.footnote).foregroundStyle(C.texto2)
            }
            ForEach(opciones.indices, id: \.self) { i in
                VStack(alignment: .leading, spacing: 10) {
                    HStack {
                        TextField("Nombre (Talle, Color…)", text: $opciones[i].nombre).estilo(.fila).foregroundStyle(C.texto)
                        Button { opciones.remove(at: i) } label: { Image(systemName: "trash").foregroundStyle(C.errorTinta) }
                            .accessibilityLabel("Sacar opción")
                    }
                    FlujoChips(valores: opciones[i].valores) { v in opciones[i].valores.removeAll { $0 == v } }
                    HStack {
                        TextField("+ Valor (ej.: Rojo)", text: Binding(get: { nuevoValor[i] ?? "" }, set: { nuevoValor[i] = $0 }))
                            .estilo(.body16).submitLabel(.done)
                            .onSubmit { sumarValor(i) }
                        Button("Agregar") { sumarValor(i) }.buttonStyle(BotonTonal())
                            .disabled((nuevoValor[i] ?? "").trimmingCharacters(in: .whitespaces).isEmpty)
                    }
                }
                .padding(14).tarjeta(radio: 16)
            }
            if !combinaciones.isEmpty {
                VStack(spacing: 0) {
                    HStack {
                        Text("Variante").frame(maxWidth: .infinity, alignment: .leading)
                        Text("Stock").frame(width: 64)
                        Text("Precio").frame(width: 92)
                    }
                    .estilo(.footnoteFuerte).foregroundStyle(C.texto2).padding(.horizontal, 14).padding(.vertical, 10)
                    ForEach(combinaciones, id: \.self) { c in
                        let k = clave(c)
                        Separador()
                        HStack {
                            Text(k).estilo(.fila).foregroundStyle(C.texto).frame(maxWidth: .infinity, alignment: .leading).lineLimit(1)
                            TextField("0", text: Binding(get: { variantes[k]?.stock ?? "" }, set: { variantes[k, default: Combinacion()].stock = $0 }))
                                .keyboardType(.numberPad).multilineTextAlignment(.center).frame(width: 64)
                                .padding(.vertical, 6)
                                .background((Int(variantes[k]?.stock ?? "") ?? 0) == 0 ? C.errorFondo : C.pista, in: RoundedRectangle(cornerRadius: 8))
                            TextField(precio.isEmpty ? "Igual" : precio, text: Binding(get: { variantes[k]?.precio ?? "" }, set: { variantes[k, default: Combinacion()].precio = $0 }))
                                .keyboardType(.numberPad).multilineTextAlignment(.center).frame(width: 92)
                                .padding(.vertical, 6).background(C.pista, in: RoundedRectangle(cornerRadius: 8))
                        }
                        .estilo(.body16).monospacedDigit()
                        .padding(.horizontal, 14).padding(.vertical, 8)
                    }
                }
                .tarjeta(radio: 16)
                Text("\(Formato.cantidad(combinaciones.count, "combinación", "combinaciones")) · stock total \(combinaciones.reduce(0) { $0 + (Int(variantes[clave($1)]?.stock ?? "") ?? 0) })")
                    .estilo(.footnote).foregroundStyle(C.texto2).padding(.horizontal, 4)
            }
        }
    }

    private func sumarValor(_ i: Int) {
        let v = (nuevoValor[i] ?? "").trimmingCharacters(in: .whitespaces)
        guard !v.isEmpty, !opciones[i].valores.contains(where: { $0.caseInsensitiveCompare(v) == .orderedSame }) else { return }
        opciones[i].valores.append(String(v.prefix(20)))
        nuevoValor[i] = ""
    }

    private var valido: Bool {
        !nombre.trimmingCharacters(in: .whitespaces).isEmpty && (Int(precio.filter(\.isNumber)) ?? 0) > 0
    }

    private var seccionFotos: some View {
        ScrollView(.horizontal, showsIndicators: false) {
            HStack(spacing: 10) {
                ForEach(fotos, id: \.self) { f in
                    Miniatura(foto: f, tam: 96, radio: 16)
                        .overlay(alignment: .topTrailing) {
                            Button { fotos.removeAll { $0 == f } } label: {
                                Image(systemName: "xmark").font(.system(size: 10, weight: .bold)).foregroundStyle(.white)
                                    .frame(width: 22, height: 22).background(.black.opacity(0.55), in: Circle())
                            }
                            .padding(5)
                            .accessibilityLabel("Sacar foto")
                        }
                }
                ForEach(0..<subiendo, id: \.self) { _ in
                    RoundedRectangle(cornerRadius: 16, style: .continuous).fill(C.pista)
                        .frame(width: 96, height: 96).overlay(ProgressView().tint(C.acento))
                }
                if fotos.count + subiendo < 8 {
                    Button { camara = true } label: { botonFoto("camera", "Sacar foto") }.buttonStyle(.plain)
                    PhotosPicker(selection: $elegidas, maxSelectionCount: 8 - fotos.count - subiendo, matching: .images) {
                        botonFoto("photo.on.rectangle", "Galería")
                    }
                }
            }
        }
    }

    private func botonFoto(_ simbolo: String, _ texto: String) -> some View {
        VStack(spacing: 6) {
            Image(systemName: simbolo).font(.system(size: 22, weight: .medium)).foregroundStyle(C.primarioTexto)
            Text(texto).estilo(.footnoteFuerte).foregroundStyle(C.primarioTinta)
        }
        .frame(width: 96, height: 96)
        .background(C.crema2, in: RoundedRectangle(cornerRadius: 16, style: .continuous))
        .overlay(RoundedRectangle(cornerRadius: 16, style: .continuous).strokeBorder(C.linea2, style: StrokeStyle(lineWidth: 1, dash: [5, 4])))
    }

    private func campo<V: View>(_ titulo: String, @ViewBuilder _ contenido: () -> V) -> some View {
        VStack(alignment: .leading, spacing: 6) {
            Text(titulo).estilo(.footnoteFuerte).foregroundStyle(C.texto2)
            contenido()
                .estilo(.body16).foregroundStyle(C.texto)
                .padding(.horizontal, 14).padding(.vertical, 12)
                .background(C.superficie, in: RoundedRectangle(cornerRadius: 14, style: .continuous))
                .overlay(RoundedRectangle(cornerRadius: 14, style: .continuous).strokeBorder(C.linea2, lineWidth: 1))
        }
        .frame(maxWidth: .infinity, alignment: .leading)
    }

    private func cargar() async {
        defer { cargando = false }
        guard let id, original.isEmpty else {
            if let sku, codigo.isEmpty { codigo = sku }
            return
        }
        guard let v: ValorJSON = try? await estado.api.get("/v1/panel/productos/\(id)"), case let .objeto(o) = v else {
            estado.avisar("No pudimos traer el producto", ok: false)
            return
        }
        original = o
        nombre = v["nombre"]?.texto ?? ""
        precio = v["precio"]?.entero.map(String.init) ?? ""
        costo = v["costo"]?.entero.map { $0 > 0 ? String($0) : "" } ?? ""
        stock = v["stock"]?.entero.map(String.init) ?? ""
        codigo = v["sku"]?.texto ?? ""
        categoria = v["categoria"]?.texto ?? ""
        descripcion = v["descripcion"]?.texto ?? ""
        fotos = v["fotos"]?.lista?.compactMap(\.texto) ?? (v["foto"]?.texto.map { [$0] } ?? [])
        conVariantes = !(v["variantes"]?.lista?.isEmpty ?? true)
        opciones = v["opciones"]?.lista?.compactMap { o in
            guard let n = o["nombre"]?.texto else { return nil }
            return Opcion(nombre: n, valores: o["valores"]?.lista?.compactMap(\.texto) ?? [])
        } ?? []
        teniaOpciones = !opciones.isEmpty
        for x in v["variantes"]?.lista ?? [] {
            guard case let .objeto(vals)? = x["valores"] else { continue }
            let c = vals.compactMapValues(\.texto)
            variantes[clave(c)] = Combinacion(stock: x["stock"]?.entero.map(String.init) ?? "",
                                              precio: x["precio"]?.entero.map(String.init) ?? "", sku: x["sku"]?.texto ?? "")
        }
    }

    private func subir(_ img: UIImage) async {
        let lado: CGFloat = 1600
        let escala = min(1, lado / max(img.size.width, img.size.height))
        let tam = CGSize(width: img.size.width * escala, height: img.size.height * escala)
        let chica = UIGraphicsImageRenderer(size: tam).image { _ in img.draw(in: CGRect(origin: .zero, size: tam)) }
        guard let datos = chica.jpegData(compressionQuality: 0.85) else { return }
        subiendo += 1
        defer { subiendo -= 1 }
        do {
            fotos.append(try await Web.subirFoto(datos, token: estado.api.token))
        } catch {
            estado.avisar("No pudimos subir la foto", ok: false)
        }
    }

    private func guardar() async {
        guardando = true
        defer { guardando = false }
        var o = original
        // Lo que la API da de más (id, slug, publicado, vendidos…) no se manda: sólo lo editable.
        for k in ["id", "slug", "publicado", "vendidos30d", "estado"] { o.removeValue(forKey: k) }
        o["nombre"] = .texto(nombre.trimmingCharacters(in: .whitespaces))
        o["precio"] = .numero(Double(Int(precio.filter(\.isNumber)) ?? 0))
        o["costo"] = .numero(Double(Int(costo.filter(\.isNumber)) ?? 0))
        let combos = combinaciones
        if combos.isEmpty {
            o["stock"] = .numero(Double(Int(stock.filter(\.isNumber)) ?? 0))
            if teniaOpciones { o["opciones"] = .lista([]); o["variantes"] = .lista([]) } else { o.removeValue(forKey: "opciones"); o.removeValue(forKey: "variantes") }
        } else {
            o["opciones"] = .lista(opciones.filter { !$0.nombre.isEmpty && !$0.valores.isEmpty }.map {
                .objeto(["nombre": .texto($0.nombre), "valores": .lista($0.valores.map { .texto($0) })])
            })
            o["variantes"] = .lista(combos.map { c in
                let x = variantes[clave(c)] ?? Combinacion()
                var v: [String: ValorJSON] = ["valores": .objeto(c.mapValues { .texto($0) }),
                                              "stock": .numero(Double(Int(x.stock.filter(\.isNumber)) ?? 0))]
                v["precio"] = Int(x.precio.filter(\.isNumber)).map { .numero(Double($0)) } ?? .nulo
                if !x.sku.isEmpty { v["sku"] = .texto(x.sku) }
                return .objeto(v)
            })
            o.removeValue(forKey: "stock")
        }
        o["sku"] = .texto(codigo.trimmingCharacters(in: .whitespaces))
        o["categoria"] = categoria.trimmingCharacters(in: .whitespaces).isEmpty ? .nulo : .texto(categoria.trimmingCharacters(in: .whitespaces))
        o["descripcion"] = .texto(descripcion)
        o["fotos"] = .lista(fotos.map { .texto($0) })
        o["foto"] = fotos.first.map { .texto($0) } ?? .nulo
        do {
            if let id {
                try await estado.api.hacer("PUT", "/v1/panel/productos/\(id)", ValorJSON.objeto(o))
                estado.avisar("Guardamos los cambios")
            } else {
                try await estado.api.hacer("POST", "/v1/panel/productos", ValorJSON.objeto(o))
                estado.avisar("¡Producto cargado!")
            }
            Haptico.exito()
            cerrar()
        } catch let e as ErrorApi {
            estado.avisar(e.detalle, ok: false)
        } catch {
            estado.avisar("No pudimos guardarlo", ok: false)
        }
    }

    private func eliminar() async {
        guard let id else { return }
        do {
            try await estado.api.hacer("DELETE", "/v1/panel/productos/\(id)")
            estado.avisar("Borramos el producto")
            cerrar()
        } catch {
            estado.avisar("No pudimos borrarlo", ok: false)
        }
    }
}

/// La cámara del sistema para sacar una foto.
struct Camara: UIViewControllerRepresentable {
    var alSacar: (UIImage) -> Void
    @Environment(\.dismiss) private var cerrar

    func makeUIViewController(context: Context) -> UIImagePickerController {
        let c = UIImagePickerController()
        c.sourceType = UIImagePickerController.isSourceTypeAvailable(.camera) ? .camera : .photoLibrary
        c.delegate = context.coordinator
        return c
    }

    func updateUIViewController(_: UIImagePickerController, context _: Context) {}

    func makeCoordinator() -> Coordinador { Coordinador(self) }

    final class Coordinador: NSObject, UIImagePickerControllerDelegate, UINavigationControllerDelegate {
        let padre: Camara
        init(_ p: Camara) { padre = p }
        func imagePickerController(_: UIImagePickerController, didFinishPickingMediaWithInfo info: [UIImagePickerController.InfoKey: Any]) {
            if let img = info[.originalImage] as? UIImage { padre.alSacar(img) }
            padre.cerrar()
        }
        func imagePickerControllerDidCancel(_: UIImagePickerController) { padre.cerrar() }
    }
}


/// Los valores de una opción como píldoras, con «×» para sacarlos.
struct FlujoChips: View {
    var valores: [String]
    var sacar: (String) -> Void
    var body: some View {
        ScrollView(.horizontal, showsIndicators: false) {
            HStack(spacing: 6) {
                ForEach(valores, id: \.self) { v in
                    HStack(spacing: 4) {
                        Text(v).estilo(.footnoteFuerte).foregroundStyle(C.primarioTinta)
                        Button { sacar(v) } label: { Image(systemName: "xmark").font(.system(size: 9, weight: .bold)).foregroundStyle(C.primarioTinta) }
                            .accessibilityLabel("Sacar \(v)")
                    }
                    .padding(.horizontal, 10).padding(.vertical, 6)
                    .background(C.crema, in: Capsule())
                }
            }
        }
    }
}
