import SwiftUI

/// Compartir la tienda: copiar el link, mandarlo por WhatsApp, la imagen para estados e historias,
/// el QR, mensajes listos y el link de cada producto y cupón. Lo usan Inicio, Más, Productos y Mi
/// tienda, para que el link no quede escondido.
@MainActor
enum Compartir {
    static func copiar(_ url: URL, _ estado: Estado, aviso: String = "Copiamos el link de tu tienda") {
        UIPasteboard.general.string = url.absoluteString
        Haptico.exito()
        estado.avisar(aviso)
    }

    static func copiar(texto: String, _ estado: Estado, aviso: String = "Copiamos el mensaje") {
        UIPasteboard.general.string = texto
        Haptico.exito()
        estado.avisar(aviso)
    }

    /// WhatsApp con el mensaje escrito; la persona elige a quién (o «Mi estado»).
    static func whatsapp(_ texto: String) -> URL {
        var c = URLComponents(string: "https://wa.me/")!
        c.queryItems = [URLQueryItem(name: "text", value: texto)]
        return c.url!
    }

    static func mensajeTienda(_ t: Acceso.Tienda) -> String {
        "¡Hola! Te comparto mi tienda online, \(t.nombre). Mirá lo que hay y pedí desde ahí: \(t.url.absoluteString)"
    }

    static func mensajeProducto(_ nombre: String, precio: Int?, url: URL) -> String {
        let p = precio.map { " a \(Formato.pesos($0))" } ?? ""
        return "\(nombre)\(p). Lo pedís acá: \(url.absoluteString)"
    }

    /// Los mensajes listos para copiar o mandar (el link siempre al final, así WhatsApp arma la vista previa).
    static func mensajes(_ t: Acceso.Tienda) -> [(id: String, titulo: String, texto: String)] {
        let u = t.url.absoluteString
        return [
            ("abrimos", "Para avisar que abriste", "¡Ya tenemos tienda online! 🎉 Ahora podés ver todo lo de \(t.nombre) y pedir desde el celu: \(u)"),
            ("novedades", "Para tus estados", "Llegaron novedades a \(t.nombre) 👀 Pasá a mirarlas: \(u)"),
            ("bio", "Para la bio de Instagram", "Pedí online 👉 \(u)"),
            ("cliente", "Para un cliente", "¡Gracias por tu compra! La próxima podés pedir directo desde acá: \(u)"),
        ]
    }
}

/// La pantalla «Compartí tu tienda».
struct CompartirTienda: View {
    @Environment(Estado.self) private var estado
    @Environment(Navegacion.self) private var nav
    @Environment(\.openURL) private var abrir
    @Environment(\.margen) private var margen
    @State private var qr = false
    @State private var historia = false

    struct Datos {
        var visitas: Tablero.Visitas?
        var productos: [ProductoFila]
        var cupones: [Cupones.Cupon]
    }

    var body: some View {
        if let t = estado.tienda {
            let conCupones = estado.puede("catalogo")
            PantallaSeccion(titulo: "Compartí tu tienda", que: "tu tienda") { () async throws -> Datos in
                // Uno detrás del otro, sin `async let`: en Release (Xcode 26.3) el async let dentro de este
                // closure rompía el asignador de tareas (swift_task_dealloc) y la app se cerraba.
                let v = await tablero()
                let p = await masVendidos()
                let c = await cuponesActivos(conCupones)
                return Datos(visitas: v?.visitas, productos: p, cupones: c)
            } contenido: { d, _ in
                heroe(t, d.visitas)
                canales(t)
                mensajes(t)
                productos(t, d.productos)
                if !d.cupones.isEmpty { cupones(t, d.cupones) }
            }
            .sheet(isPresented: $qr) { QRTienda(tienda: t).hojaTLW([.large]) }
            .sheet(isPresented: $historia) { HojaHistoria(tienda: t).hojaTLW([.large]) }
        }
    }

    private func tablero() async -> Tablero? {
        try? await estado.api.get("/v1/panel/tablero", query: [URLQueryItem(name: "periodo", value: "7d")])
    }

    private func masVendidos() async -> [ProductoFila] {
        let p: PaginaProductos? = try? await estado.api.get("/v1/panel/listas/productos", query: [
            URLQueryItem(name: "seg", value: "active"), URLQueryItem(name: "orden", value: "sold"),
            URLQueryItem(name: "cuantos", value: "12"),
        ])
        return p?.filas ?? []
    }

    private func cuponesActivos(_ conCupones: Bool) async -> [Cupones.Cupon] {
        guard conCupones else { return [] }
        let c: Cupones.Datos? = try? await estado.api.get("/v1/panel/cupones")
        return c?.cupones.filter { $0.estado == "activo" } ?? []
    }

    // MARK: El link

    private func heroe(_ t: Acceso.Tienda, _ v: Tablero.Visitas?) -> some View {
        TarjetaHeroe(etiqueta: "El link de tu tienda", valor: t.nombre, detalle: Formato.sinViudas(textoVisitas(v)), tincho: .whatsapp) {
            Text(t.direccion).estilo(.subheadFuerte).foregroundStyle(.white)
                .lineLimit(1).minimumScaleFactor(0.7)
                .padding(.trailing, 70)
            HStack(spacing: 8) {
                Button { Compartir.copiar(t.url, estado) } label: {
                    // El fondo es blanco siempre: la tinta va fija (la de oscuro, clarita, no se leía sobre blanco).
                    Label("Copiar link", systemImage: "doc.on.doc")
                        .estilo(.label).foregroundStyle(Color(hex: 0x8E3A16))
                        .padding(.horizontal, 14).frame(height: 38)
                        .background(.white, in: RoundedRectangle(cornerRadius: 11, style: .continuous))
                }
                Button { nav.web = t.url } label: {
                    Label("Ver", systemImage: "arrow.up.right.square")
                        .estilo(.label).foregroundStyle(.white)
                        .padding(.horizontal, 14).frame(height: 38)
                        .background(.white.opacity(0.18), in: RoundedRectangle(cornerRadius: 11, style: .continuous))
                }
            }
            .buttonStyle(.plain)
            .padding(.top, 6)
        }
    }

    private func textoVisitas(_ v: Tablero.Visitas?) -> String {
        guard let v, v.desde != nil else { return "Compartila y empezá a sumar visitas." }
        let n = Int(v.dias7.actual)
        if n == 0 { return "Nadie la visitó en 7 días: es el momento de compartirla." }
        return n == 1 ? "1 persona la visitó en los últimos 7 días." : "\(Formato.numero(n)) personas la visitaron en los últimos 7 días."
    }

    // MARK: Por dónde

    private func canales(_ t: Acceso.Tienda) -> some View {
        VStack(alignment: .leading, spacing: 10) {
            TituloSeccion(texto: "Mandala por").padding(.top, 10)
            HStack(spacing: 8) {
                AccionRapida(simbolo: "message", texto: "WhatsApp") { abrir(Compartir.whatsapp(Compartir.mensajeTienda(t))) }
                AccionRapida(simbolo: "photo.on.rectangle", texto: "Historia") { historia = true }
                AccionRapida(simbolo: "qrcode", texto: "QR") { qr = true }
                ShareLink(item: t.url, message: Text(Compartir.mensajeTienda(t))) {
                    VStack(spacing: 3) {
                        Image(systemName: "square.and.arrow.up").font(.system(size: 18)).foregroundStyle(C.primarioTexto)
                        Text("Más").estilo(.footnoteFuerte).foregroundStyle(C.primarioTinta)
                    }
                    .frame(maxWidth: .infinity, minHeight: 58)
                    .background(C.crema2, in: RoundedRectangle(cornerRadius: 14, style: .continuous))
                }
                .buttonStyle(.plain)
            }
        }
    }

    // MARK: Mensajes listos

    private func mensajes(_ t: Acceso.Tienda) -> some View {
        VStack(alignment: .leading, spacing: 10) {
            TituloSeccion(texto: "Mensajes listos").padding(.top, 10)
            VStack(spacing: 0) {
                ForEach(Array(Compartir.mensajes(t).enumerated()), id: \.element.id) { i, m in
                    if i > 0 { Separador() }
                    VStack(alignment: .leading, spacing: 8) {
                        Text(m.titulo).estilo(.footnoteFuerte).foregroundStyle(C.texto2)
                        Text(m.texto).estilo(.texto14).foregroundStyle(C.texto).fixedSize(horizontal: false, vertical: true)
                        HStack(spacing: 8) {
                            Button("Copiar") { Compartir.copiar(texto: m.texto, estado) }
                            Button("WhatsApp") { abrir(Compartir.whatsapp(m.texto)) }
                        }
                        .buttonStyle(BotonTonal())
                    }
                    .padding(16)
                    .frame(maxWidth: .infinity, alignment: .leading)
                }
            }
            .tarjeta()
        }
    }

    // MARK: Productos

    @ViewBuilder
    private func productos(_ t: Acceso.Tienda, _ lista: [ProductoFila]) -> some View {
        TituloSeccion(texto: "Compartí un producto").padding(.top, 10)
        if lista.isEmpty {
            HStack(spacing: 13) {
                Image(systemName: "shippingbox").font(.system(size: 16, weight: .semibold)).foregroundStyle(C.primarioTexto)
                    .frame(width: 40, height: 40).background(C.crema, in: RoundedRectangle(cornerRadius: 12, style: .continuous))
                VStack(alignment: .leading, spacing: 1) {
                    Text("Todavía no hay productos publicados").estilo(.fila).foregroundStyle(C.texto)
                    Text("Con una foto alcanza").estilo(.footnote).foregroundStyle(C.texto2)
                }
                Spacer(minLength: 4)
                NavigationLink(value: Destino.nuevoProducto(nil)) { Text("Cargar") }.buttonStyle(BotonTonal())
            }
            .padding(.horizontal, 16).padding(.vertical, 13)
            .tarjeta()
        } else {
            ScrollView(.horizontal) {
                HStack(alignment: .top, spacing: 10) {
                    ForEach(lista) { p in
                        let u = t.url(producto: p.id)
                        VStack(alignment: .leading, spacing: 7) {
                            Miniatura(foto: p.foto, tam: 140, alto: 120, radio: 16)
                            Text(p.nombre).estilo(Estilo(tamano: 14, peso: .semibold, relativo: .footnote))
                                .foregroundStyle(C.texto).lineLimit(2, reservesSpace: true)
                            Text(Formato.pesos(p.precio)).estilo(.footnote).monospacedDigit().foregroundStyle(C.texto2)
                            HStack(spacing: 6) {
                                ShareLink(item: u, message: Text(Compartir.mensajeProducto(p.nombre, precio: p.precio, url: u))) {
                                    Image(systemName: "square.and.arrow.up").frame(maxWidth: .infinity)
                                }
                                .accessibilityLabel("Compartir \(p.nombre)")
                                Button { Compartir.copiar(u, estado, aviso: "Copiamos el link del producto") } label: {
                                    Image(systemName: "doc.on.doc").frame(maxWidth: .infinity)
                                }
                                .accessibilityLabel("Copiar el link de \(p.nombre)")
                            }
                            .buttonStyle(BotonTonal())
                        }
                        .frame(width: 140, alignment: .leading)
                    }
                }
                .padding(.horizontal, margen)
            }
            .scrollIndicators(.hidden)
            .padding(.horizontal, -margen)
        }
    }

    // MARK: Cupones

    private func cupones(_ t: Acceso.Tienda, _ lista: [Cupones.Cupon]) -> some View {
        VStack(alignment: .leading, spacing: 10) {
            TituloSeccion(texto: "Tus cupones activos").padding(.top, 10)
            GrupoFilas(elementos: lista) { c in
                let texto = "Usá el código \(c.codigo) en \(t.nombre) y tenés \(Cupones.resumen(c).lowercased()): \(t.url.absoluteString)"
                HStack(spacing: 12) {
                    VStack(alignment: .leading, spacing: 3) {
                        Text(c.codigo).font(Estilo(tamano: 16, peso: .bold, relativo: .body).font).monospaced().foregroundStyle(C.texto)
                        Text(Cupones.resumen(c)).estilo(.footnote).foregroundStyle(C.texto2)
                    }
                    Spacer(minLength: 6)
                    ShareLink(item: t.url, message: Text(texto)) { Text("Compartir") }
                        .buttonStyle(BotonTonal())
                }
                .padding(14)
            }
        }
    }
}

// MARK: - Imagen para estados e historias

/// La hoja con la imagen 9:16 (1080 × 1920) para estados de WhatsApp e historias de Instagram:
/// la portada de Mi tienda (si hay), el nombre, el QR y el link.
struct HojaHistoria: View {
    var tienda: Acceso.Tienda
    @Environment(Estado.self) private var estado
    @State private var portada: UIImage?
    @State private var logo: UIImage?
    @State private var imagen: UIImage?

    var body: some View {
        VStack(spacing: 14) {
            Grabber()
            Text("Para tus estados e historias").estilo(.tituloHojaChico).foregroundStyle(C.texto)
            Group {
                if let imagen {
                    Image(uiImage: imagen).resizable().scaledToFit()
                        .clipShape(RoundedRectangle(cornerRadius: 18, style: .continuous))
                        .shadow(color: Color(hex: 0x2A140C, alfa: 0.2), radius: 16, y: 10)
                        .accessibilityLabel("Imagen de \(tienda.nombre) con el QR y el link")
                } else {
                    RoundedRectangle(cornerRadius: 18, style: .continuous).fill(C.pista)
                        .aspectRatio(9 / 16, contentMode: .fit)
                        .overlay(ProgressView().tint(C.acento))
                }
            }
            .frame(maxHeight: .infinity)
            if let imagen {
                ShareLink(item: Image(uiImage: imagen), preview: SharePreview(tienda.nombre, image: Image(uiImage: imagen))) {
                    Label("Compartir o guardar", systemImage: "square.and.arrow.up")
                }
                .buttonStyle(BotonPrincipal())
            }
            Text(Formato.sinViudas("Subila a tu estado de WhatsApp o a tu historia y pegá el link en la bio."))
                .estilo(.footnote).foregroundStyle(C.texto2).multilineTextAlignment(.center)
        }
        .padding(.horizontal, 20).padding(.bottom, 12)
        .task { await armar() }
    }

    private func armar() async {
        if let v: ValorJSON = try? await estado.api.get("/v1/panel/mi-tienda"), case let .objeto(d)? = v["diseno"] {
            portada = await bajar(d["portada"]?.texto)
            logo = await bajar(d["logo"]?.texto)
        }
        let r = ImageRenderer(content: TarjetaHistoria(tienda: tienda, portada: portada, logo: logo))
        r.scale = 3
        imagen = r.uiImage
    }

    private func bajar(_ s: String?) async -> UIImage? {
        guard let s, !s.isEmpty, let u = URL(string: s), let r = try? await URLSession.shared.data(from: u) else { return nil }
        return UIImage(data: r.0)
    }
}

/// La imagen en sí (360 × 640 puntos, se renderiza a 3×). Colores fijos: no depende del modo oscuro.
struct TarjetaHistoria: View {
    var tienda: Acceso.Tienda
    var portada: UIImage?
    var logo: UIImage?

    var body: some View {
        VStack(spacing: 0) {
            ZStack(alignment: .bottom) {
                if let portada {
                    Image(uiImage: portada).resizable().scaledToFill().frame(width: 360, height: 300).clipped()
                    LinearGradient(colors: [.clear, Color(hex: 0xB0441D, alfa: 0.9)], startPoint: .center, endPoint: .bottom)
                } else {
                    Image(Tincho.saludo.imagen).resizable().scaledToFit().frame(height: 220).padding(.bottom, 20)
                }
            }
            .frame(width: 360, height: 300)
            VStack(spacing: 12) {
                if let logo {
                    Image(uiImage: logo).resizable().scaledToFit().frame(width: 64, height: 64)
                        .padding(6).background(.white, in: RoundedRectangle(cornerRadius: 16, style: .continuous))
                }
                Text(tienda.nombre).font(Estilo(tamano: 30, peso: .bold, tracking: -0.02, relativo: .title).font)
                    .foregroundStyle(.white).multilineTextAlignment(.center).lineLimit(2).minimumScaleFactor(0.6)
                Text("Ya podés comprar online").font(Estilo(tamano: 17, peso: .semibold, relativo: .body).font)
                    .foregroundStyle(.white.opacity(0.9))
                if let qr = QRTienda.imagen(tienda.url.absoluteString) {
                    Image(uiImage: qr).interpolation(.none).resizable().scaledToFit()
                        .frame(width: 130, height: 130).padding(10)
                        .background(.white, in: RoundedRectangle(cornerRadius: 18, style: .continuous))
                        .padding(.top, 4)
                }
                Text(tienda.direccion).font(Estilo(tamano: 15, peso: .bold, relativo: .body).font)
                    .foregroundStyle(Color(hex: 0x8A3414))
                    .lineLimit(1).minimumScaleFactor(0.6)
                    .padding(.horizontal, 16).padding(.vertical, 9)
                    .background(Color(hex: 0xFBF3EA), in: Capsule())
            }
            .padding(.horizontal, 24)
            .padding(.top, portada == nil ? 0 : 14)
            Spacer(minLength: 0)
            Text("Hecha con Tu Local Web").font(Estilo(tamano: 12, peso: .medium, relativo: .caption).font)
                .foregroundStyle(.white.opacity(0.75)).padding(.bottom, 18)
        }
        .frame(width: 360, height: 640)
        .background(
            LinearGradient(colors: [Color(hex: 0xD9652F), Color(hex: 0xB0441D)], startPoint: .top, endPoint: .bottom)
        )
        .environment(\.colorScheme, .light)
    }
}
