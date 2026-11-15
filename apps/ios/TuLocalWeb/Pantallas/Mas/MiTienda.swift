import PhotosUI
import SwiftUI

/// Editar la tienda online desde el celular, igual que Mi tienda en el panel web (diseño v2, en bloques):
/// la vista previa real arriba; Páginas (Inicio, Catálogo, Ficha, Carrito y Nosotros con sus bloques),
/// Estilo (punto de partida, dirección, colores, tipografías, encabezado, pie, menú, barra y redes) y
/// Marca (logos, al compartir y la dirección). Se guarda como borrador y se publica.
struct MiTiendaEditor: View {
    @Environment(Estado.self) private var estado
    @Environment(\.openURL) private var abrir
    @State private var editor = EditorTienda()
    @State private var datos = DatosEditor()
    @State private var cargado = false
    @State private var pestana = "paginas"
    @State private var direccion = ""
    @State private var slug = ""
    @State private var sinPublicar = false
    @State private var guardando = false
    @State private var fotoLogo: PhotosPickerItem?
    @State private var fotoLogoOscuro: PhotosPickerItem?
    @State private var fotoIcono: PhotosPickerItem?
    @State private var subiendo: String?
    @State private var categorias: [String] = []
    @State private var previaGrande = false
    @State private var vuelta = 0
    /// La vista previa responde al protocolo: los cambios se ven sin guardar.
    @State private var viva = false
    @State private var presetPedido: PlantillaTienda?
    private let plantillas = PlantillaTienda.todas()

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 14) {
                Text("Mi tienda").estilo(.tituloGrande).foregroundStyle(C.texto)
                if !cargado {
                    EsqueletoFilas(filas: 5, miniatura: 0)
                } else {
                    vistaPrevia
                    Segmentado(opciones: [("paginas", "Páginas"), ("estilo", "Estilo"), ("marca", "Marca")], elegido: $pestana)
                    switch pestana {
                    case "estilo": pestanaEstilo
                    case "marca": pestanaMarca
                    default: ListaBloques(editor: editor)
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
                Button(editor.sinGuardar ? "Guardar borrador" : "Borrador guardado") { Task { await guardar(publicar: false) } }
                    .buttonStyle(BotonTexto()).disabled(guardando || subiendo != nil || !editor.sinGuardar)
            }
        }
        .background(C.fondo.ignoresSafeArea())
        .navigationBarTitleDisplayMode(.inline)
        .environment(datos)
        .task { await cargar() }
        .onChange(of: fotoLogo) { _, i in Task { if let u = await subir(i, que: "logo") { editor.cambiar { $0.logo = u } } } }
        .onChange(of: fotoLogoOscuro) { _, i in Task { if let u = await subir(i, que: "logo para fondos oscuros") { editor.cambiar { $0.estilo.logoOscuro = u } } } }
        .onChange(of: fotoIcono) { _, i in Task { if let u = await subir(i, que: "ícono de la pestaña") { editor.cambiar { $0.estilo.icono = u } } } }
        .sheet(isPresented: Binding(get: { editor.panel != nil }, set: { if !$0 { editor.panel = nil } })) {
            if let id = editor.panel {
                PanelBloque(editor: editor, id: id)
                    .environment(estado).environment(datos)
                    .hojaTLW([.medium, .large])
                    .presentationBackgroundInteraction(.enabled(upThrough: .medium))
            }
        }
        .sheet(isPresented: Binding(get: { editor.insertarEn != nil }, set: { if !$0 { editor.insertarEn = nil } })) {
            BibliotecaBloques(editor: editor).hojaTLW([.large])
        }
        .sheet(isPresented: $previaGrande) {
            NavigationStack {
                VistaPreviaTienda(token: estado.api.token, vuelta: vuelta, pagina: editor.pagina,
                                  diseno: editor.diseno.textoJSON)
                    .ignoresSafeArea(edges: .bottom)
                    .navigationTitle("Así se ve")
                    .navigationBarTitleDisplayMode(.inline)
                    .toolbar { ToolbarItem(placement: .confirmationAction) { Button("Listo") { previaGrande = false } } }
            }
        }
        .confirmationDialog(presetPedido.map { "¿Pasar a «\($0.nombre)»?" } ?? "", isPresented: Binding(
            get: { presetPedido != nil }, set: { if !$0 { presetPedido = nil } }), titleVisibility: .visible, presenting: presetPedido) { p in
            Button("Pasar a \(p.nombre)") { cambiarPreset(p) }
            Button("Cancelar", role: .cancel) {}
        } message: { _ in
            Text("Cambian la dirección, el orden y el diseño de los bloques. Tus textos y fotos quedan, y podés deshacerlo.")
        }
    }

    // MARK: Vista previa

    private var vistaPrevia: some View {
        VStack(alignment: .leading, spacing: 0) {
            VistaPreviaTienda(token: estado.api.token, vuelta: vuelta, pagina: editor.pagina, diseno: editor.diseno.textoJSON,
                              seleccion: editor.seleccion, marcos: true,
                              alTocarBloque: { id in editor.abrir(id); pestana = "paginas" },
                              alCambiarViva: { viva = $0 })
                .frame(height: 380)
                .background(C.pista)
                .clipShape(UnevenRoundedRectangle(topLeadingRadius: 20, topTrailingRadius: 20, style: .continuous))
            HStack(spacing: 12) {
                if editor.sinGuardar && !viva {
                    ChipSimple(texto: "Guardá para verlo", fondo: C.avisoFondo, tinta: C.avisoTinta)
                } else if sinPublicar || editor.sinGuardar {
                    ChipSimple(texto: "Hay cambios sin publicar", fondo: C.avisoFondo, tinta: C.avisoTinta)
                } else {
                    ChipSimple(texto: "Publicada", fondo: C.okFondo, tinta: C.okTinta)
                }
                Spacer()
                if !viva {
                    Button("Actualizar") { Task { await mirar(grande: false) } }
                        .estilo(.subheadFuerte).foregroundStyle(C.primarioTexto)
                        .disabled(guardando || subiendo != nil)
                }
                Button { Task { await mirar(grande: true) } } label: { Image(systemName: "arrow.up.left.and.arrow.down.right") }
                    .foregroundStyle(C.primarioTexto).accessibilityLabel("Ver en grande")
                Button("Ver tienda") { if let u = urlTienda { abrir(u) } }
                    .estilo(.subheadFuerte).foregroundStyle(C.primarioTexto)
            }
            .padding(12)
            if let u = urlTienda, !sinPublicar, !editor.sinGuardar {
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

    // MARK: Estilo

    private var pestanaEstilo: some View {
        VStack(alignment: .leading, spacing: 14) {
            TarjetaSeccion(titulo: "Punto de partida", simbolo: "rectangle.3.group") {
                Text("Cada uno trae su dirección, sus colores y sus bloques. Cambiarlo conserva tus textos y fotos.")
                    .estilo(.footnote).foregroundStyle(C.texto2)
                ScrollView(.horizontal, showsIndicators: false) {
                    HStack(spacing: 10) {
                        ForEach(plantillas) { p in
                            Button { if p.id != editor.plantilla { presetPedido = p } } label: {
                                VStack(alignment: .leading, spacing: 6) {
                                    AsyncImage(url: URL(string: p.img)) { i in i.resizable().scaledToFill() } placeholder: { Color(hex: p.colorFondo) }
                                        .frame(width: 120, height: 150).clipped()
                                        .overlay(alignment: .bottomLeading) {
                                            Text(p.lema).font(.caption2.weight(.semibold)).foregroundStyle(.white)
                                                .padding(6).shadow(radius: 3)
                                        }
                                        .clipShape(RoundedRectangle(cornerRadius: 14, style: .continuous))
                                        .overlay(RoundedRectangle(cornerRadius: 14, style: .continuous)
                                            .strokeBorder(editor.plantilla == p.id ? C.acento : .clear, lineWidth: 3))
                                    Text(p.nombre).estilo(.fila).foregroundStyle(C.texto)
                                    Text(nombreDireccion(p.id)).estilo(.footnote).foregroundStyle(C.texto2)
                                }
                                .frame(width: 120)
                            }
                            .buttonStyle(.plain)
                            .accessibilityAddTraits(editor.plantilla == p.id ? .isSelected : [])
                        }
                    }
                }
            }
            EditorEstilo(estilo: editor.campo(\.estilo, clave: "estilo"), paleta: editor.campo(\.paleta, clave: "paleta"),
                         plantilla: editor.plantilla, categorias: categorias)
        }
    }

    private func nombreDireccion(_ plantilla: String) -> String {
        let d = PresetTienda.todos[plantilla]?.direccion ?? ""
        let o = CatalogoBloques.shared.estilo["direccion"]
        return o.map { CatalogoBloques.nombre($0, d) } ?? d
    }

    private func cambiarPreset(_ p: PlantillaTienda) {
        editor.aplicarPreset(p.id)
        Haptico.exito()
        estado.avisar("Pasamos a «\(p.nombre)»", deshacer: { [editor] in editor.deshacer() })
    }

    // MARK: Marca

    private var pestanaMarca: some View {
        VStack(alignment: .leading, spacing: 14) {
            TarjetaSeccion(titulo: "Nombre y logo", simbolo: "seal") {
                CampoTexto(titulo: "Nombre que se ve (vacío: el de la tienda)", texto: editor.campo(\.titulo, clave: "titulo"))
                CampoTexto(titulo: "Frase", texto: editor.campo(\.lema, clave: "lema"))
                HStack(spacing: 10) {
                    selectorFoto("Logo", url: editor.diseno.logo, item: $fotoLogo)
                    selectorFoto("Logo para fondos oscuros", url: editor.diseno.estilo.logoOscuro ?? "", item: $fotoLogoOscuro, oscuro: true)
                    selectorFoto("Ícono de la pestaña", url: editor.diseno.estilo.icono ?? "", item: $fotoIcono)
                }
                if !editor.diseno.logo.isEmpty {
                    VStack(alignment: .leading, spacing: 6) {
                        Text("Tamaño del logo · \(editor.diseno.estilo.logoAlto.map { "\($0) px" } ?? "el de la dirección")")
                            .estilo(.footnoteFuerte).foregroundStyle(C.texto2)
                        Slider(value: Binding(
                            get: { Double(editor.diseno.estilo.logoAlto ?? 36) },
                            set: { v in editor.cambiar(clave: "logoAlto") { $0.estilo.logoAlto = Int(v) } }
                        ), in: 24...72, step: 2).tint(C.acento)
                    }
                }
                HStack {
                    if !editor.diseno.logo.isEmpty { Button("Quitar logo") { editor.cambiar { $0.logo = "" } } }
                    Spacer()
                    if !(editor.diseno.estilo.logoOscuro ?? "").isEmpty { Button("Quitar logo oscuro") { editor.cambiar { $0.estilo.logoOscuro = nil } } }
                    Spacer()
                    if !(editor.diseno.estilo.icono ?? "").isEmpty { Button("Quitar ícono") { editor.cambiar { $0.estilo.icono = nil } } }
                }
                .estilo(.footnoteFuerte).foregroundStyle(C.primarioTexto)
            }
            TarjetaSeccion(titulo: "Al compartir el link", simbolo: "square.and.arrow.up") {
                CampoTexto(titulo: "Título (vacío: el sugerido)", texto: editor.campo(\.compartirTitulo, clave: "compartirTitulo"))
                CampoTexto(titulo: "Descripción", texto: editor.campo(\.compartirDescripcion, clave: "compartirDescripcion"))
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

    private func selectorFoto(_ titulo: String, url: String, item: Binding<PhotosPickerItem?>, oscuro: Bool = false) -> some View {
        PhotosPicker(selection: item, matching: .images) {
            VStack(spacing: 6) {
                if subiendo == titulo.lowercased() {
                    ProgressView().tint(C.acento).frame(height: 60)
                } else if !url.isEmpty {
                    Color.clear.frame(height: 60).frame(maxWidth: .infinity)
                        .overlay { AsyncImage(url: URL(string: url)) { i in i.resizable().scaledToFit() } placeholder: { C.pista } }
                        .padding(4)
                        .background(oscuro ? Color(hex: 0x1C1A17) : .clear)
                        .clipShape(RoundedRectangle(cornerRadius: 10, style: .continuous))
                } else {
                    Image(systemName: "photo.badge.plus").font(.system(size: 22)).foregroundStyle(C.primarioTexto).frame(height: 60)
                }
                Text(titulo).estilo(.footnoteFuerte).foregroundStyle(C.primarioTinta).multilineTextAlignment(.center)
                    .lineLimit(2).minimumScaleFactor(0.85)
            }
            .frame(maxWidth: .infinity).padding(8)
            .background(C.crema2, in: RoundedRectangle(cornerRadius: 14, style: .continuous))
        }
    }

    private func subir(_ item: PhotosPickerItem?, que: String) async -> String? {
        guard let item, let d = try? await item.loadTransferable(type: Data.self), let jpg = Web.jpegParaSubir(d) else { return nil }
        subiendo = que
        defer { subiendo = nil }
        do { return try await Web.subirFoto(jpg, token: estado.api.token) } catch {
            estado.avisar("No pudimos subir la foto", ok: false); return nil
        }
    }

    // MARK: Cargar y guardar

    private func cargar() async {
        guard let v: ValorJSON = try? await estado.api.get("/v1/panel/mi-tienda") else {
            estado.avisar("No pudimos traer tu tienda", ok: false); cargado = true; return
        }
        let plantilla = v["plantillaId"]?.texto ?? v["plantilla"]?.texto ?? ""
        editor.empezar(Diseno(json: v["diseno"]?.objeto ?? [:], plantillaTienda: plantilla), plantilla: plantilla)
        direccion = v["direccion"]?.texto ?? ""
        sinPublicar = v["sinPublicar"]?.bool ?? false
        categorias = v["vista"]?["categorias"]?.lista?.compactMap { $0.texto } ?? []
        let productos = v["vista"]?["productos"]?.lista ?? []
        datos.contexto = ContextoTienda(
            nombre: estado.tienda?.nombre ?? "",
            categorias: categorias,
            productos: productos.count,
            resenas: v["vista"]?["resena"]?.objeto != nil,
            fotos: !(v["vista"]?["fotos"]?.lista ?? []).isEmpty || productos.contains { !($0["foto"]?.texto ?? "").isEmpty }
        )
        slug = estado.tienda?.slug ?? ""
        cargado = true
        // Lo que sólo hace falta para los avisos y los selectores: después de mostrar el editor.
        await datos.cargarNegocio(estado.api)
        await datos.cargarProductos(estado.api)
    }

    @discardableResult
    private func guardar(publicar: Bool, avisar: Bool = true) async -> Bool {
        let d = editor.diseno
        let cuerpo = ValorJSON.objeto(d.json)
        if let bytes = try? JSONEncoder().encode(cuerpo), bytes.count > CatalogoBloques.shared.limites.bytesDiseno {
            estado.avisar("Tu tienda quedó muy grande: sacá algunos bloques o textos largos", ok: false)
            return false
        }
        guardando = true
        defer { guardando = false }
        do {
            if editor.sinGuardar || !publicar {
                try await estado.api.hacer("PUT", "/v1/panel/mi-tienda", cuerpo)
                editor.guardado = d
                // El logo nuevo se ve enseguida en Inicio y en Más (en vez de las iniciales).
                if d.logo != (estado.tienda?.logo ?? "") { Task { await estado.refrescarAcceso() } }
            }
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

    /// Sin la vista conectada: guarda el borrador y recarga la tienda con él.
    private func mirar(grande: Bool) async {
        if !viva || grande {
            if editor.sinGuardar, !(await guardar(publicar: false, avisar: false)) { return }
            if !viva { vuelta += 1 }
        }
        if grande { previaGrande = true }
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
