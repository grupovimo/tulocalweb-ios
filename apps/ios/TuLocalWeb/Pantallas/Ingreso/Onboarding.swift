import SwiftUI

/// El onboarding de una tienda recién creada, entero en el celular (como /onboarding en la web): tu
/// negocio (rubro), traer tu tienda de otra plataforma, plantilla y dirección, cobros y ¡listo! Cada
/// paso se guarda en la tienda (PUT /v1/panel/onboarding), así se retoma donde quedó, también desde la
/// web. «Lo hago después» deja seguir al panel en cualquier paso.
struct Onboarding: View {
    @Environment(Estado.self) private var estado

    enum Paso: Int, CaseIterable {
        case negocio, importar, tienda, cobros, listo
        var titulo: String {
            switch self {
            case .negocio: "¿Qué vendés?"
            case .importar: "¿Ya vendés en otra plataforma?"
            case .tienda: "Elegí cómo se ve tu tienda"
            case .cobros: "¿Cómo te pagan?"
            case .listo: "¡Tu tienda está online!"
            }
        }
        var bajada: String {
            switch self {
            case .negocio: "Así armamos tu tienda con lo que necesita tu rubro."
            case .importar: "Traemos tus productos, clientes y pedidos sin cargar nada a mano."
            case .tienda: "Una plantilla para empezar y la dirección de tu tienda. Las cambiás cuando quieras."
            case .cobros: "Conectá Mercado Pago para cobrar con tarjeta y QR, o dejá tus datos para transferencias."
            case .listo: "Compartí el link y empezá a vender."
            }
        }
    }

    struct Rubro: Decodable, Identifiable { var id: String; var nombre: String; var plantilla: String }

    @State private var paso: Paso = .negocio
    @State private var cargado = false
    @State private var guardando = false
    @State private var rubros: [Rubro] = []
    @State private var rubro = ""
    @State private var plantilla = ""
    @State private var traer: Bool?
    @State private var slug = ""
    @State private var disponible: String?
    @State private var slugOk = true
    @State private var mpConectado = false
    @State private var conectando = false
    @State private var alias = ""
    @State private var cbu = ""
    @State private var titular = ""
    @State private var negocio: [String: ValorJSON] = [:]
    @State private var qr = false
    private let plantillas = PlantillaTienda.todas()

    var body: some View {
        VStack(spacing: 0) {
            cabecera
            ScrollView {
                VStack(alignment: .leading, spacing: 16) {
                    VStack(alignment: .leading, spacing: 6) {
                        Text(paso.titulo).estilo(.saludo).foregroundStyle(C.texto)
                            .accessibilityAddTraits(.isHeader)
                        Text(Formato.sinViudas(paso.bajada)).estilo(.body16).foregroundStyle(C.texto2)
                    }
                    if !cargado {
                        EsqueletoFilas(filas: 4, miniatura: 0)
                    } else {
                        switch paso {
                        case .negocio: pasoNegocio
                        case .importar: pasoImportar
                        case .tienda: pasoTienda
                        case .cobros: pasoCobros
                        case .listo: pasoListo
                        }
                    }
                }
                .padding(.horizontal, 24)
                .padding(.top, 8)
                .padding(.bottom, 24)
            }
            .scrollDismissesKeyboard(.interactively)
        }
        .background(C.fondo.ignoresSafeArea())
        .barraAccion {
            Button {
                Task { await seguir() }
            } label: {
                if guardando { ProgressView().tint(.white) } else { Text(paso == .listo ? "Ir a mi panel" : "Seguir") }
            }
            .buttonStyle(BotonPrincipal())
            .disabled(guardando || !puedeSeguir)
            if paso != .listo {
                Button("Lo hago después") { Task { await despues() } }
                    .buttonStyle(BotonTexto(color: C.texto2))
            }
        }
        .sheet(isPresented: $qr) {
            if let t = estado.tienda { QRTienda(tienda: t).hojaTLW([.large]) }
        }
        .task { await cargar() }
    }

    // MARK: Cabecera (volver y avance)

    private var cabecera: some View {
        HStack(spacing: 14) {
            Button {
                if let p = Paso(rawValue: paso.rawValue - 1) { withAnimation { paso = p } }
            } label: {
                Image(systemName: "chevron.left").font(.system(size: 17, weight: .semibold)).foregroundStyle(C.texto)
                    .frame(width: 40, height: 40)
            }
            .opacity(paso == .negocio || paso == .listo ? 0 : 1)
            .disabled(paso == .negocio || paso == .listo)
            .accessibilityLabel("Volver al paso anterior")
            GeometryReader { g in
                ZStack(alignment: .leading) {
                    Capsule().fill(C.pista)
                    Capsule().fill(C.acento).frame(width: g.size.width * CGFloat(paso.rawValue + 1) / CGFloat(Paso.allCases.count))
                }
            }
            .frame(height: 6)
            .accessibilityLabel("Paso \(paso.rawValue + 1) de \(Paso.allCases.count)")
            Text("\(paso.rawValue + 1)/\(Paso.allCases.count)").estilo(.footnoteFuerte).foregroundStyle(C.texto2).monospacedDigit()
        }
        .padding(.horizontal, 16)
        .padding(.top, 8)
        .padding(.bottom, 8)
    }

    // MARK: Pasos

    private var pasoNegocio: some View {
        LazyVGrid(columns: [GridItem(.flexible(), spacing: 10), GridItem(.flexible(), spacing: 10)], spacing: 10) {
            ForEach(rubros) { r in
                Button {
                    Haptico.seleccion()
                    rubro = r.id
                    if plantillas.contains(where: { $0.id == r.plantilla }) { plantilla = r.plantilla }
                } label: {
                    Text(r.nombre).estilo(.fila).foregroundStyle(C.texto)
                        .multilineTextAlignment(.leading)
                        .frame(maxWidth: .infinity, minHeight: 54, alignment: .leading)
                        .padding(.horizontal, 14)
                        .tarjeta(radio: 14)
                        .overlay(RoundedRectangle(cornerRadius: 14, style: .continuous)
                            .strokeBorder(rubro == r.id ? C.acento : .clear, lineWidth: 1.5))
                }
                .buttonStyle(.plain)
                .accessibilityAddTraits(rubro == r.id ? .isSelected : [])
            }
        }
    }

    private var pasoImportar: some View {
        VStack(alignment: .leading, spacing: 12) {
            opcion("Sí, quiero traer mi tienda", "Tiendanube, Shopify, WooCommerce o Empretienda", simbolo: "arrow.down.doc", elegida: traer == true) { traer = true }
            opcion("No, empiezo de cero", "Cargás tus productos con una foto, desde el celular", simbolo: "sparkles", elegida: traer == false) { traer = false }
            if traer == true {
                TraerTienda().padding(.top, 4)
            }
        }
    }

    private var pasoTienda: some View {
        VStack(alignment: .leading, spacing: 14) {
            ScrollView(.horizontal, showsIndicators: false) {
                HStack(spacing: 10) {
                    ForEach(plantillas) { p in
                        Button { Haptico.seleccion(); plantilla = p.id } label: {
                            VStack(alignment: .leading, spacing: 6) {
                                AsyncImage(url: URL(string: p.img)) { i in i.resizable().scaledToFill() } placeholder: { Color(hex: p.colorFondo) }
                                    .frame(width: 128, height: 164).clipped()
                                    .clipShape(RoundedRectangle(cornerRadius: 14, style: .continuous))
                                    .overlay(RoundedRectangle(cornerRadius: 14, style: .continuous)
                                        .strokeBorder(plantilla == p.id ? C.acento : .clear, lineWidth: 3))
                                Text(p.nombre).estilo(.fila).foregroundStyle(C.texto)
                                Text(p.rubro).estilo(.footnote).foregroundStyle(C.texto2).lineLimit(1)
                            }
                            .frame(width: 128)
                        }
                        .buttonStyle(.plain)
                        .accessibilityAddTraits(plantilla == p.id ? .isSelected : [])
                    }
                }
                .padding(.horizontal, 24)
            }
            .padding(.horizontal, -24)
            VStack(alignment: .leading, spacing: 6) {
                Text("La dirección de tu tienda").estilo(.footnoteFuerte).foregroundStyle(C.texto2)
                HStack(spacing: 0) {
                    TextField("mitienda", text: $slug)
                        .textInputAutocapitalization(.never).autocorrectionDisabled().keyboardType(.URL)
                        .onChange(of: slug) { _, s in
                            let limpio = String(s.lowercased().filter { $0.isLetter && $0.isASCII || $0.isNumber || $0 == "-" }.prefix(60))
                            if limpio != s { slug = limpio }
                            Task { await revisarSlug() }
                        }
                    Text(".tulocalweb.com.ar").foregroundStyle(C.texto2)
                }
                .estilo(.body16).foregroundStyle(C.texto)
                .padding(.horizontal, 14).padding(.vertical, 12)
                .background(C.superficie, in: RoundedRectangle(cornerRadius: 14, style: .continuous))
                .overlay(RoundedRectangle(cornerRadius: 14, style: .continuous).strokeBorder(slugOk ? C.linea2 : C.errorTinta, lineWidth: 1))
                if let disponible {
                    Text(disponible).estilo(.footnote).foregroundStyle(slugOk ? C.okTinta : C.errorTinta)
                }
            }
        }
    }

    private var pasoCobros: some View {
        VStack(alignment: .leading, spacing: 14) {
            HStack(spacing: 12) {
                Image(systemName: "creditcard.fill").font(.system(size: 16, weight: .semibold)).foregroundStyle(Color(hex: 0x009EE3))
                    .frame(width: 40, height: 40).background(Color(hex: 0x009EE3).opacity(0.12), in: RoundedRectangle(cornerRadius: 12, style: .continuous))
                VStack(alignment: .leading, spacing: 2) {
                    Text("Mercado Pago").estilo(.fila).foregroundStyle(C.texto)
                    Text(mpConectado ? "Conectado: ya cobrás con tarjeta y QR" : "Tarjeta, cuotas, QR y link de pago").estilo(.footnote).foregroundStyle(C.texto2)
                }
                Spacer(minLength: 6)
                if mpConectado {
                    ChipSimple(texto: "Conectada", fondo: C.okFondo, tinta: C.okTinta)
                } else if conectando {
                    ProgressView().tint(C.acento)
                } else {
                    Button("Conectar") {
                        Task {
                            conectando = true
                            mpConectado = await ConectarCuenta.compartido.conectar("mercadopago", api: estado.api)
                            conectando = false
                            if mpConectado { Haptico.exito() }
                        }
                    }
                    .buttonStyle(BotonTonal())
                }
            }
            .padding(14)
            .tarjeta(radio: 16)
            Text("Transferencias").estilo(.headline).foregroundStyle(C.texto).padding(.top, 4)
            CampoTexto(titulo: "Alias", texto: $alias)
            CampoTexto(titulo: "CBU o CVU", texto: $cbu, teclado: .numberPad)
            CampoTexto(titulo: "Titular de la cuenta", texto: $titular)
            Text(Formato.sinViudas("Aparecen en la tienda cuando alguien elige pagar con transferencia."))
                .estilo(.footnote).foregroundStyle(C.texto2)
        }
    }

    private var pasoListo: some View {
        VStack(alignment: .leading, spacing: 14) {
            if let t = estado.tienda {
                HStack(spacing: 12) {
                    AvatarTienda(tienda: t, tam: 48, radio: 14)
                    VStack(alignment: .leading, spacing: 2) {
                        Text(t.nombre).estilo(.fila).foregroundStyle(C.texto)
                        Text(t.direccion).estilo(.footnote).foregroundStyle(C.primarioTexto).lineLimit(1).minimumScaleFactor(0.8)
                    }
                }
                .padding(14)
                .frame(maxWidth: .infinity, alignment: .leading)
                .tarjeta(radio: 16)
                HStack(spacing: 8) {
                    AccionRapida(simbolo: "doc.on.doc", texto: "Copiar") { Compartir.copiar(t.url, estado) }
                    ShareLink(item: t.url, message: Text(Compartir.mensajeTienda(t))) {
                        VStack(spacing: 3) {
                            Image(systemName: "square.and.arrow.up").font(.system(size: 18)).foregroundStyle(C.primarioTexto)
                            Text("Compartir").estilo(.footnoteFuerte).foregroundStyle(C.primarioTinta)
                        }
                        .frame(maxWidth: .infinity, minHeight: 58)
                        .background(C.crema2, in: RoundedRectangle(cornerRadius: 14, style: .continuous))
                    }
                    .buttonStyle(.plain)
                    AccionRapida(simbolo: "qrcode", texto: "QR") { qr = true }
                }
            }
            Text(Formato.sinViudas("En Inicio te esperan los primeros pasos: cargar tus productos con una foto y compartir el link."))
                .estilo(.footnote).foregroundStyle(C.texto2)
        }
    }

    private func opcion(_ t: String, _ d: String, simbolo: String, elegida: Bool, _ accion: @escaping () -> Void) -> some View {
        Button {
            Haptico.seleccion()
            withAnimation(.easeOut(duration: 0.2)) { accion() }
        } label: {
            HStack(spacing: 12) {
                Image(systemName: simbolo).font(.system(size: 16, weight: .semibold)).foregroundStyle(C.primarioTexto)
                    .frame(width: 40, height: 40).background(C.crema, in: RoundedRectangle(cornerRadius: 12, style: .continuous))
                VStack(alignment: .leading, spacing: 2) {
                    Text(t).estilo(.fila).foregroundStyle(C.texto)
                    Text(d).estilo(.footnote).foregroundStyle(C.texto2).multilineTextAlignment(.leading)
                }
                Spacer(minLength: 0)
            }
            .padding(14)
            .tarjeta(radio: 16)
            .overlay(RoundedRectangle(cornerRadius: 16, style: .continuous).strokeBorder(elegida ? C.acento : .clear, lineWidth: 1.5))
        }
        .buttonStyle(.plain)
        .accessibilityAddTraits(elegida ? .isSelected : [])
    }

    // MARK: Lógica

    private var puedeSeguir: Bool {
        switch paso {
        case .negocio: !rubro.isEmpty
        case .importar: traer != nil
        case .tienda: !plantilla.isEmpty && slugOk && slug.count >= 2
        default: true
        }
    }

    private func cargar() async {
        defer { cargado = true }
        slug = estado.tienda?.slug ?? ""
        if let r: [Rubro] = try? await estado.api.get("/v1/rubros") { rubros = r }
        if let v: ValorJSON = try? await estado.api.get("/v1/panel/onboarding") {
            plantilla = v["plantilla"]?.texto ?? ""
            if case let .objeto(e)? = v["elecciones"] {
                rubro = e["rubro"]?.texto ?? ""
                if let p = e["pasoApp"]?.entero, let pp = Paso(rawValue: p) { paso = pp }
                if let b = e["traer"]?.bool { traer = b }
            }
        }
        if plantilla.isEmpty { plantilla = plantillas.first?.id ?? "" }
        if case let .objeto(n)? = (try? await estado.api.get("/v1/panel/configuracion") as ValorJSON)?["negocio"] {
            negocio = n
            alias = n["alias"]?.texto ?? ""
            cbu = n["cbu"]?.texto ?? ""
            titular = n["titular"]?.texto ?? ""
        }
        struct MP: Decodable { var conectada: Bool? }
        mpConectado = ((try? await estado.api.get("/v1/panel/integraciones/mercadopago") as MP)?.conectada) ?? false
    }

    private func revisarSlug() async {
        let s = slug
        guard s.count >= 2 else { slugOk = false; disponible = "Al menos 2 letras o números."; return }
        if s == estado.tienda?.slug { slugOk = true; disponible = nil; return }
        try? await Task.sleep(for: .milliseconds(350))
        guard s == slug else { return }
        struct D: Decodable { var disponible: Bool; var mensaje: String }
        if let d: D = try? await estado.api.get("/v1/panel/tienda/direccion", query: [URLQueryItem(name: "slug", value: s)]), s == slug {
            slugOk = d.disponible
            disponible = d.mensaje
        }
    }

    /// Guarda lo elegido (y en qué paso va) en la tienda.
    private func guardar(terminado: Bool = false) async throws {
        var e: [String: ValorJSON] = ["rubro": .texto(rubro), "pasoApp": .numero(Double(paso.rawValue))]
        if let traer { e["traer"] = .booleano(traer) }
        var o: [String: ValorJSON] = ["elecciones": .objeto(e), "terminado": .booleano(terminado)]
        if !plantilla.isEmpty { o["plantilla"] = .texto(plantilla) }
        try await estado.api.hacer("PUT", "/v1/panel/onboarding", ValorJSON.objeto(o))
    }

    private func seguir() async {
        guardando = true
        defer { guardando = false }
        do {
            switch paso {
            case .tienda:
                if slug != estado.tienda?.slug {
                    struct C2: Encodable { var slug: String }
                    try await estado.api.hacer("PUT", "/v1/panel/tienda/direccion", C2(slug: slug))
                    await estado.refrescarAcceso()
                }
            case .cobros:
                let datos = [alias, cbu, titular].map { $0.trimmingCharacters(in: .whitespaces) }
                if datos != [negocio["alias"]?.texto ?? "", negocio["cbu"]?.texto ?? "", negocio["titular"]?.texto ?? ""] {
                    var n = negocio
                    n["alias"] = .texto(datos[0]); n["cbu"] = .texto(datos[1]); n["titular"] = .texto(datos[2])
                    try await estado.api.hacer("PUT", "/v1/panel/configuracion/negocio", ValorJSON.objeto(n))
                    negocio = n
                }
            case .listo:
                try await guardar(terminado: true)
                Haptico.exito()
                estado.terminarOnboarding()
                return
            default:
                break
            }
            if let p = Paso(rawValue: paso.rawValue + 1) {
                withAnimation(.easeInOut(duration: 0.25)) { paso = p }
                try await guardar(terminado: p == .listo)
                if p == .listo { Haptico.exito() }
            }
        } catch let e as ErrorApi {
            Haptico.error()
            estado.avisar(e.detalle, ok: false)
        } catch {
            Haptico.error()
            estado.avisar("No pudimos guardar. Revisá tu conexión y probá de nuevo.", ok: false)
        }
    }

    private func despues() async {
        try? await guardar()
        estado.terminarOnboarding()
    }
}
