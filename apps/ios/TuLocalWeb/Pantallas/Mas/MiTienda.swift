import PhotosUI
import SwiftUI

/// Editar la tienda online desde el celular, igual que Mi tienda en el panel web: plantilla, colores,
/// tipografías, logos, portada, secciones, menú, títulos, anuncio, pie, redes, cómo se ve al
/// compartir y la dirección. Se guarda como borrador, se mira en la vista previa real y se publica.
struct MiTiendaEditor: View {
    @Environment(Estado.self) private var estado
    @Environment(\.openURL) private var abrir
    @State private var diseno: [String: ValorJSON] = [:]
    @State private var cargado = false
    @State private var plantilla = ""
    @State private var paleta = 0
    @State private var titulo = ""
    @State private var lema = ""
    @State private var portada = ""
    @State private var logo = ""
    @State private var compartirTitulo = ""
    @State private var compartirDescripcion = ""
    @State private var secciones: [Seccion] = []
    @State private var direccion = ""
    @State private var slug = ""
    @State private var sinPublicar = false
    @State private var guardando = false
    @State private var fotoPortada: PhotosPickerItem?
    @State private var fotoLogo: PhotosPickerItem?
    @State private var fotoLogoOscuro: PhotosPickerItem?
    @State private var fotoIcono: PhotosPickerItem?
    @State private var subiendo: String?
    @State private var estilo = EstiloTienda()
    @State private var categorias: [String] = []
    @State private var previa = false
    @State private var vuelta = 0
    private let plantillas = PlantillaTienda.todas()

    struct Seccion: Identifiable, Equatable { var id: String; var visible: Bool }
    static let nombres = ["hero": "Portada", "cats": "Categorías", "prods": "Productos", "revs": "Reseñas", "ig": "Instagram", "wa": "Botón de WhatsApp"]

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 14) {
                Text("Mi tienda").estilo(.tituloGrande).foregroundStyle(C.texto)
                if !cargado {
                    EsqueletoFilas(filas: 5, miniatura: 0)
                } else {
                    vistaPrevia
                    TarjetaSeccion(titulo: "Plantilla", simbolo: "rectangle.3.group") {
                        ScrollView(.horizontal, showsIndicators: false) {
                            HStack(spacing: 10) {
                                ForEach(plantillas) { p in
                                    Button { plantilla = p.id } label: {
                                        VStack(alignment: .leading, spacing: 6) {
                                            AsyncImage(url: URL(string: p.img)) { i in i.resizable().scaledToFill() } placeholder: { Color(hex: p.colorFondo) }
                                                .frame(width: 120, height: 150).clipped()
                                                .overlay(alignment: .bottomLeading) {
                                                    Text(p.lema).font(.caption2.weight(.semibold)).foregroundStyle(.white)
                                                        .padding(6).shadow(radius: 3)
                                                }
                                                .clipShape(RoundedRectangle(cornerRadius: 14, style: .continuous))
                                                .overlay(RoundedRectangle(cornerRadius: 14, style: .continuous)
                                                    .strokeBorder(plantilla == p.id ? C.acento : .clear, lineWidth: 3))
                                            Text(p.nombre).estilo(.fila).foregroundStyle(C.texto)
                                            Text(p.rubro).estilo(.footnote).foregroundStyle(C.texto2)
                                        }
                                        .frame(width: 120)
                                    }
                                    .buttonStyle(.plain)
                                }
                            }
                        }
                    }
                    TarjetaSeccion(titulo: "Portada", simbolo: "photo.on.rectangle") {
                        CampoTexto(titulo: "Título (vacío: el nombre de la tienda)", texto: $titulo)
                        CampoTexto(titulo: "Lema", texto: $lema)
                        HStack(spacing: 10) {
                            selectorFoto("Foto de portada", url: portada, item: $fotoPortada)
                            selectorFoto("Logo", url: logo, item: $fotoLogo)
                        }
                    }
                    TarjetaSeccion(titulo: "Logos", simbolo: "seal") {
                        HStack(spacing: 10) {
                            selectorFoto("Logo para fondos oscuros", url: estilo.logoOscuro ?? "", item: $fotoLogoOscuro, oscuro: true)
                            selectorFoto("Ícono de la pestaña", url: estilo.icono ?? "", item: $fotoIcono)
                        }
                        if !logo.isEmpty {
                            VStack(alignment: .leading, spacing: 6) {
                                Text("Tamaño del logo · \(estilo.logoAlto.map { "\($0) px" } ?? "el de la plantilla")")
                                    .estilo(.footnoteFuerte).foregroundStyle(C.texto2)
                                Slider(value: Binding(
                                    get: { Double(estilo.logoAlto ?? 36) },
                                    set: { estilo.logoAlto = Int($0) }
                                ), in: 24...72, step: 2).tint(C.acento)
                            }
                        }
                        if !(estilo.logoOscuro ?? "").isEmpty || !(estilo.icono ?? "").isEmpty {
                            HStack {
                                if !(estilo.logoOscuro ?? "").isEmpty { Button("Quitar logo oscuro") { estilo.logoOscuro = nil } }
                                Spacer()
                                if !(estilo.icono ?? "").isEmpty { Button("Quitar ícono") { estilo.icono = nil } }
                            }
                            .buttonStyle(BotonTexto())
                        }
                    }
                    TarjetaSeccion(titulo: "Secciones", simbolo: "list.bullet.rectangle") {
                        VStack(spacing: 0) {
                            ForEach(Array(secciones.enumerated()), id: \.element.id) { i, s in
                                if i > 0 { Separador() }
                                HStack(spacing: 10) {
                                    VStack(spacing: 2) {
                                        Button { mover(i, -1) } label: { Image(systemName: "chevron.up") }.disabled(i == 0)
                                        Button { mover(i, 1) } label: { Image(systemName: "chevron.down") }.disabled(i == secciones.count - 1)
                                    }
                                    .font(.system(size: 12, weight: .bold)).foregroundStyle(C.primarioTexto).buttonStyle(.plain)
                                    Text(MiTiendaEditor.nombres[s.id] ?? s.id).estilo(.body16).foregroundStyle(C.texto)
                                    Spacer()
                                    Toggle("", isOn: $secciones[i].visible).labelsHidden().tint(C.ok)
                                }
                                .padding(.vertical, 8)
                            }
                        }
                    }
                    EditorEstilo(estilo: $estilo, paleta: $paleta, plantilla: plantilla, categorias: categorias)
                    TarjetaSeccion(titulo: "Al compartir el link", simbolo: "square.and.arrow.up") {
                        CampoTexto(titulo: "Título (vacío: el sugerido)", texto: $compartirTitulo)
                        CampoTexto(titulo: "Descripción", texto: $compartirDescripcion)
                    }
                    TarjetaSeccion(titulo: "Dirección de tu tienda", simbolo: "link") {
                        Text(direccion).estilo(.texto14).foregroundStyle(C.primarioTexto).textSelection(.enabled)
                        CampoTexto(titulo: "Nombre en la dirección (letras, números y guiones)", texto: $slug)
                        Button("Cambiar dirección") { Task { await cambiarDireccion() } }
                            .buttonStyle(BotonSecundario())
                            .disabled(slug.isEmpty || direccion.contains(slug))
                    }
                }
            }
            .margenPantalla()
            .padding(.bottom, 32)
        }
        .scrollDismissesKeyboard(.interactively)
        .barraAccion {
            if cargado {
                Button(guardando ? "Publicando…" : "Publicar cambios") { Task { await guardar(publicar: true) } }
                    .buttonStyle(BotonPrincipal()).disabled(guardando || subiendo != nil)
                Button("Guardar borrador") { Task { await guardar(publicar: false) } }
                    .buttonStyle(BotonTexto()).disabled(guardando || subiendo != nil)
            }
        }
        .background(C.fondo.ignoresSafeArea())
        .navigationBarTitleDisplayMode(.inline)
        .task { await cargar() }
        .onChange(of: fotoPortada) { _, i in Task { if let u = await subir(i, que: "portada") { portada = u } } }
        .onChange(of: fotoLogo) { _, i in Task { if let u = await subir(i, que: "logo") { logo = u } } }
        .onChange(of: fotoLogoOscuro) { _, i in Task { if let u = await subir(i, que: "logo para fondos oscuros") { estilo.logoOscuro = u } } }
        .onChange(of: fotoIcono) { _, i in Task { if let u = await subir(i, que: "ícono de la pestaña") { estilo.icono = u } } }
        .sheet(isPresented: $previa) {
            NavigationStack {
                VistaPreviaTienda(token: estado.api.token, vuelta: vuelta)
                    .ignoresSafeArea(edges: .bottom)
                    .navigationTitle("Así se ve")
                    .navigationBarTitleDisplayMode(.inline)
                    .toolbar { ToolbarItem(placement: .confirmationAction) { Button("Listo") { previa = false } } }
            }
        }
    }

    private var vistaPrevia: some View {
        let p = plantillas.first { $0.id == plantilla }
        return VStack(alignment: .leading, spacing: 0) {
            ZStack(alignment: .bottomLeading) {
                // La foto llena un lugar fijo y se recorta ahí: suelta, una muy ancha ensanchaba todo el editor.
                Color.clear.frame(height: 170).frame(maxWidth: .infinity)
                    .overlay {
                        AsyncImage(url: URL(string: portada.isEmpty ? (p?.img ?? "") : portada)) { i in i.resizable().scaledToFill() }
                            placeholder: { Color(hex: p?.colorFondo ?? 0xF4EFEA) }
                    }
                    .clipped()
                LinearGradient(colors: [.clear, .black.opacity(0.55)], startPoint: .center, endPoint: .bottom)
                HStack(alignment: .bottom, spacing: 10) {
                    if !logo.isEmpty {
                        AsyncImage(url: URL(string: logo)) { i in i.resizable().scaledToFit() } placeholder: { Color.clear }
                            .frame(width: 44, height: 44).background(.white, in: RoundedRectangle(cornerRadius: 10))
                    }
                    VStack(alignment: .leading, spacing: 2) {
                        Text(titulo.isEmpty ? (estado.tienda?.nombre ?? "") : titulo).font(.title3.bold()).foregroundStyle(.white)
                        if !lema.isEmpty { Text(lema).font(.footnote).foregroundStyle(.white.opacity(0.9)) }
                    }
                }
                .padding(14)
            }
            HStack {
                if sinPublicar {
                    ChipSimple(texto: "Hay cambios sin publicar", fondo: C.avisoFondo, tinta: C.avisoTinta)
                } else {
                    ChipSimple(texto: "Publicada", fondo: C.okFondo, tinta: C.okTinta)
                }
                Spacer()
                Button("Vista previa") { Task { await mirar() } }
                    .estilo(.subheadFuerte).foregroundStyle(C.primarioTexto)
                    .disabled(guardando || subiendo != nil)
                Button("Ver tienda") { if let u = urlTienda { abrir(u) } }
                    .estilo(.subheadFuerte).foregroundStyle(C.primarioTexto)
            }
            .padding(12)
            if let u = urlTienda, !sinPublicar {
                HStack(spacing: 8) {
                    Button { Compartir.copiar(u, estado) } label: { Label("Copiar link", systemImage: "doc.on.doc").frame(maxWidth: .infinity) }
                    if let t = estado.tienda {
                        ShareLink(item: u, message: Text(Compartir.mensajeTienda(t))) {
                            Label("Compartir", systemImage: "square.and.arrow.up").frame(maxWidth: .infinity)
                        }
                    }
                }
                .buttonStyle(BotonTonal(alto: 38))
                .padding(.horizontal, 12).padding(.bottom, 12)
            }
        }
        .tarjeta(radio: 20)
    }

    private var urlTienda: URL? {
        direccion.isEmpty ? estado.tienda?.url : URL(string: direccion.hasPrefix("http") ? direccion : "https://\(direccion)")
    }

    private func selectorFoto(_ titulo: String, url: String, item: Binding<PhotosPickerItem?>, oscuro: Bool = false) -> some View {
        PhotosPicker(selection: item, matching: .images) {
            VStack(spacing: 6) {
                if subiendo == titulo.lowercased() {
                    ProgressView().tint(C.acento).frame(height: 70)
                } else if !url.isEmpty {
                    Color.clear.frame(height: 70).frame(maxWidth: .infinity)
                        .overlay { AsyncImage(url: URL(string: url)) { i in i.resizable().scaledToFill() } placeholder: { C.pista } }
                        .clipped()
                        .background(oscuro ? Color(hex: 0x1C1A17) : .clear)
                        .clipShape(RoundedRectangle(cornerRadius: 10, style: .continuous))
                } else {
                    Image(systemName: "photo.badge.plus").font(.system(size: 22)).foregroundStyle(C.primarioTexto).frame(height: 70)
                }
                Text(titulo).estilo(.footnoteFuerte).foregroundStyle(C.primarioTinta).multilineTextAlignment(.center)
            }
            .frame(maxWidth: .infinity).padding(8)
            .background(C.crema2, in: RoundedRectangle(cornerRadius: 14, style: .continuous))
        }
    }

    private func mover(_ i: Int, _ d: Int) {
        let j = i + d
        guard secciones.indices.contains(j) else { return }
        secciones.swapAt(i, j)
    }

    private func subir(_ item: PhotosPickerItem?, que: String) async -> String? {
        guard let item, let d = try? await item.loadTransferable(type: Data.self), let img = UIImage(data: d) else { return nil }
        subiendo = que == "portada" ? "foto de portada" : que
        defer { subiendo = nil }
        let escala = min(1, 1600 / max(img.size.width, img.size.height))
        let tam = CGSize(width: img.size.width * escala, height: img.size.height * escala)
        let chica = UIGraphicsImageRenderer(size: tam).image { _ in img.draw(in: CGRect(origin: .zero, size: tam)) }
        guard let jpg = chica.jpegData(compressionQuality: 0.85) else { return nil }
        do { return try await Web.subirFoto(jpg, token: estado.api.token) } catch {
            estado.avisar("No pudimos subir la foto", ok: false); return nil
        }
    }

    private func cargar() async {
        guard let v: ValorJSON = try? await estado.api.get("/v1/panel/mi-tienda") else {
            estado.avisar("No pudimos traer tu tienda", ok: false); cargado = true; return
        }
        if case let .objeto(d)? = v["diseno"] { diseno = d }
        plantilla = v["plantillaId"]?.texto ?? ""
        direccion = v["direccion"]?.texto ?? ""
        sinPublicar = v["sinPublicar"]?.bool ?? false
        paleta = diseno["paleta"]?.entero ?? 0
        titulo = diseno["titulo"]?.texto ?? ""
        lema = diseno["lema"]?.texto ?? ""
        portada = diseno["portada"]?.texto ?? ""
        logo = diseno["logo"]?.texto ?? ""
        compartirTitulo = diseno["compartirTitulo"]?.texto ?? ""
        compartirDescripcion = diseno["compartirDescripcion"]?.texto ?? ""
        secciones = diseno["secciones"]?.lista?.compactMap { s in
            s["id"]?.texto.map { Seccion(id: $0, visible: s["visible"]?.bool ?? true) }
        } ?? []
        estilo = EstiloTienda.desde(diseno["estilo"])
        categorias = v["vista"]?["categorias"]?.lista?.compactMap { $0.texto } ?? []
        slug = estado.tienda?.slug ?? ""
        cargado = true
    }

    @discardableResult
    private func guardar(publicar: Bool, avisar: Bool = true) async -> Bool {
        guardando = true
        defer { guardando = false }
        var d = diseno
        d["plantilla"] = .texto(plantilla)
        d["paleta"] = .numero(Double(paleta))
        d["titulo"] = .texto(titulo)
        d["lema"] = .texto(lema)
        d["portada"] = .texto(portada)
        d["logo"] = .texto(logo)
        d["compartirTitulo"] = .texto(compartirTitulo)
        d["compartirDescripcion"] = .texto(compartirDescripcion)
        d["secciones"] = .lista(secciones.map { .objeto(["id": .texto($0.id), "visible": .booleano($0.visible)]) })
        d["estilo"] = estilo.json
        do {
            try await estado.api.hacer("PUT", "/v1/panel/mi-tienda", ValorJSON.objeto(d))
            // El logo nuevo se ve enseguida en Inicio y en Más (en vez de las iniciales).
            if d["logo"]?.texto != estado.tienda?.logo { Task { await estado.refrescarAcceso() } }
            diseno = d
            if publicar {
                try await estado.api.hacer("POST", "/v1/panel/mi-tienda/publicar")
                sinPublicar = false
                Haptico.exito()
                estado.avisar("¡Publicamos los cambios! Ahora compartí el link")
            } else {
                sinPublicar = true
                if avisar { estado.avisar("Guardamos el borrador") }
            }
            return true
        } catch let e as ErrorApi {
            estado.avisar(e.detalle, ok: false)
        } catch {
            estado.avisar("No pudimos guardar", ok: false)
        }
        return false
    }

    /// Guarda el borrador y abre la tienda de verdad con él.
    private func mirar() async {
        if await guardar(publicar: false, avisar: false) {
            vuelta += 1
            previa = true
        }
    }

    private func cambiarDireccion() async {
        struct C2: Encodable { var slug: String }
        let s = slug.lowercased().trimmingCharacters(in: .whitespaces)
        do {
            try await estado.api.hacer("PUT", "/v1/panel/tienda/direccion", C2(slug: s))
            estado.avisar("Listo: tu tienda tiene dirección nueva")
            if let v: ValorJSON = try? await estado.api.get("/v1/panel/mi-tienda") { direccion = v["direccion"]?.texto ?? direccion }
        } catch let e as ErrorApi {
            estado.avisar(e.detalle, ok: false)
        } catch {
            estado.avisar("No pudimos cambiarla", ok: false)
        }
    }
}

/// Las plantillas de tienda (Recursos/plantillas.json, generado desde apps/web/src/lib/plantillas.ts).
struct PlantillaTienda: Decodable, Identifiable {
    var id: String
    var nombre: String
    var rubro: String
    var img: String
    var bg: String
    var lema: String
    var colorFondo: UInt32 { UInt32(bg.dropFirst(), radix: 16) ?? 0xF4EFEA }

    static func todas() -> [PlantillaTienda] {
        guard let u = Bundle.main.url(forResource: "plantillas", withExtension: "json"), let d = try? Data(contentsOf: u),
              let v = try? JSONDecoder().decode([PlantillaTienda].self, from: d) else { return [] }
        return v
    }
}
