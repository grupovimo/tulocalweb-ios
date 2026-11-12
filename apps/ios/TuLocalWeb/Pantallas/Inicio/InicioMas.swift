import SwiftUI

// Lo que va debajo de la vidriera (docs/handoff/inicio-a): las acciones en un bento, «Para hacer hoy»,
// «De un vistazo» de costado (la meta, la caja del local, por dónde compran, quién está ahora, las
// visitas y las liquidaciones), los últimos pedidos, lo más vendido y, al final, lo más visto, las
// reseñas, los envíos y las integraciones. Todo con tarjetas con relieve (Tema/Relieve.swift).

/// Título de sección del Inicio (22/700) con un dato o una acción a la derecha.
struct TituloInicio: View {
    var texto: String
    var detalle: String? = nil
    var accion: String? = nil
    var tinta: Color = C.primarioTexto
    var alTocar: (() -> Void)? = nil
    var body: some View {
        HStack(alignment: .firstTextBaseline) {
            Text(texto).estilo(.seccion).foregroundStyle(C.texto).accessibilityAddTraits(.isHeader)
            Spacer(minLength: 8)
            if let accion, let alTocar {
                Button(accion, action: alTocar)
                    .estilo(.subheadFuerte).foregroundStyle(tinta)
                    .frame(minHeight: 44).contentShape(Rectangle())
            } else if let detalle {
                Text(detalle).estilo(Estilo(tamano: 14, peso: .medium, relativo: .footnote)).foregroundStyle(C.texto2)
            }
        }
    }
}

// MARK: - Acciones

/// El bento de 2 columnas: «Subir producto» alta con foto, «Cobrar» con el QR real de la tienda y
/// «Compartir» con una miniatura de la vidriera.
struct AccionesInicio: View {
    var tablero: Tablero
    var paleta: PaletaTienda
    @Environment(Estado.self) private var estado
    @Environment(Navegacion.self) private var nav

    var body: some View {
        HStack(spacing: 12) {
            subir.frame(maxWidth: .infinity)
            VStack(spacing: 12) {
                cobrar
                compartir
            }
            .frame(maxWidth: .infinity)
        }
        .frame(height: 252)
    }

    private var subir: some View {
        Button { nav.inicio.append(.nuevoProducto(nil)) } label: {
            VStack(alignment: .leading, spacing: 12) {
                ZStack(alignment: .bottomTrailing) {
                    if let f = tablero.masVendidos.first(where: { $0.foto != nil })?.foto {
                        Foto(foto: f, ancho: 500)
                    } else {
                        Color.tema(0xE9DCCD, 0x3A2419)
                            .overlay(Image(Tincho.foto.imagen).resizable().scaledToFit().padding(12))
                    }
                    // El marco blanco del visor y lo de afuera apenas más oscuro.
                    ZStack {
                        Color.black.opacity(0.12)
                        RoundedRectangle(cornerRadius: 10, style: .continuous).padding(10).blendMode(.destinationOut)
                    }
                    .compositingGroup()
                    RoundedRectangle(cornerRadius: 10, style: .continuous).strokeBorder(.white.opacity(0.85), lineWidth: 2).padding(10)
                    Image(systemName: "camera.fill").font(.system(size: 17, weight: .semibold))
                        .foregroundStyle(paleta.sobreAcento)
                        .frame(width: 40, height: 40)
                        .background(paleta.acento, in: Circle())
                        .overlay(Circle().strokeBorder(LinearGradient(colors: [.white.opacity(0.3), .clear], startPoint: .top, endPoint: .center), lineWidth: 1))
                        .shadow(color: .black.opacity(0.35), radius: 6, y: 5)
                        .padding(10)
                }
                .frame(maxWidth: .infinity, maxHeight: .infinity)
                .clipShape(RoundedRectangle(cornerRadius: 16, style: .continuous))
                .accessibilityHidden(true)
                VStack(alignment: .leading, spacing: 2) {
                    Text("Subir producto").estilo(Estilo(tamano: 17, peso: .bold, tracking: -0.012, relativo: .headline))
                    Text(Formato.sinViudas("Sacale una foto y Tincho escribe el resto"))
                        .estilo(.footnote).foregroundStyle(C.texto2).lineSpacing(1)
                        .fixedSize(horizontal: false, vertical: true)
                }
                .padding(.horizontal, 2)
            }
            .foregroundStyle(C.texto)
            .padding(12)
            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .leading)
            .relieve()
        }
        .buttonStyle(BotonHundido())
    }

    private var cobrar: some View {
        let puede = estado.puede("pedidos")
        return Button {
            if puede { nav.inicio.append(.links) } else if let t = estado.tienda { nav.web = t.url }
        } label: {
            pie(puede ? "Cobrar" : "Tu tienda", puede ? "Link o QR al toque" : "Mirala como tus clientes")
                .overlay(alignment: .topTrailing) {
                    if let t = estado.tienda {
                        CodigoQR(texto: t.url.absoluteString, color: paleta.tintaBlanco)
                            .padding(8)
                            .frame(width: 76, height: 76)
                            .background(.white, in: RoundedRectangle(cornerRadius: 14, style: .continuous))
                            .overlay(RoundedRectangle(cornerRadius: 14, style: .continuous).strokeBorder(Color(hex: 0xEADFD3), lineWidth: 1))
                            .rotationEffect(.degrees(-8))
                            .shadow(color: Color(hex: 0x2A140C, alfa: 0.25), radius: 6, y: 6)
                            .offset(x: 6, y: -6)
                            .accessibilityElement()
                            .accessibilityLabel("QR de \(t.direccion)")
                    }
                }
                .relieve()
        }
        .buttonStyle(BotonHundido())
    }

    private var compartir: some View {
        let visitas: String = if let v = tablero.visitas, v.desde != nil {
            "\(Formato.cantidad(Int(v.hoy.actual), "visita", "visitas")) hoy"
        } else {
            "Tu vidriera"
        }
        return Button { nav.inicio.append(.compartir) } label: {
            pie("Compartir", visitas)
                .overlay(alignment: .topTrailing) {
                    MiniVidriera(tienda: estado.tienda, fotos: tablero.masVendidos.map(\.foto), paleta: paleta)
                        .rotationEffect(.degrees(7))
                        .offset(x: -8, y: -14)
                }
                .relieve()
        }
        .buttonStyle(BotonHundido())
    }

    private func pie(_ titulo: String, _ detalle: String) -> some View {
        VStack(alignment: .leading, spacing: 2) {
            Text(titulo).estilo(Estilo(tamano: 17, peso: .bold, tracking: -0.012, relativo: .headline))
            Text(detalle).estilo(.footnote).foregroundStyle(C.texto2).monospacedDigit().lineLimit(1).minimumScaleFactor(0.85)
        }
        .foregroundStyle(C.texto)
        .padding(14)
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .bottomLeading)
    }
}

/// La vidriera en chiquito: la franja del acento, el logo y dos fotos de productos.
struct MiniVidriera: View {
    var tienda: Acceso.Tienda?
    var fotos: [String?]
    var paleta: PaletaTienda
    var body: some View {
        ZStack(alignment: .topLeading) {
            Color.white
            paleta.acento.frame(height: 30)
            Group {
                if let u = urlDeFoto(tienda?.logo, ancho: 60) {
                    AsyncImage(url: u) { f in f.image?.resizable().scaledToFit().padding(2) }
                } else {
                    Text(String(Formato.iniciales(tienda?.nombre ?? "").prefix(1)))
                        .font(Estilo(tamano: 11, peso: .bold, relativo: .caption2).font)
                        .foregroundStyle(paleta.tintaBlanco)
                }
            }
            .frame(width: 20, height: 20)
            .background(.white, in: RoundedRectangle(cornerRadius: 6, style: .continuous))
            .shadow(color: .black.opacity(0.25), radius: 1.5, y: 1)
            .offset(x: 7, y: 19)
            Foto(foto: fotos.first ?? nil, ancho: 90).frame(width: 20, height: 20)
                .clipShape(RoundedRectangle(cornerRadius: 5, style: .continuous)).offset(x: 7, y: 45)
            Foto(foto: fotos.dropFirst().first ?? nil, ancho: 90).frame(width: 20, height: 20)
                .clipShape(RoundedRectangle(cornerRadius: 5, style: .continuous)).offset(x: 31, y: 45)
            Capsule().fill(Color(hex: 0xEFE6DC)).frame(width: 44, height: 4).offset(x: 7, y: 70)
        }
        .frame(width: 58, height: 80)
        .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
        .overlay(RoundedRectangle(cornerRadius: 12, style: .continuous).strokeBorder(Color(hex: 0xEADFD3), lineWidth: 1))
        .shadow(color: Color(hex: 0x2A140C, alfa: 0.25), radius: 6, y: 6)
        .accessibilityHidden(true)
    }
}

/// Las acciones del día en cero (ACero): sacar una foto, vender en el local y cobrar, en tres.
struct AccionesCero: View {
    var paleta: PaletaTienda
    @Environment(Estado.self) private var estado
    @Environment(Navegacion.self) private var nav

    var body: some View {
        let puede = estado.puede("pedidos")
        HStack(spacing: 10) {
            tile("camera", "Producto", "Sacale una foto") { nav.inicio.append(.nuevoProducto(nil)) }
            tile("barcode.viewfinder", "Vender", "En el local") { nav.venta = true }
            if puede {
                tile("link", "Cobrar", "Link o QR") { nav.inicio.append(.links) }
            } else {
                tile("safari", "Tu tienda", "Mirala") { if let t = estado.tienda { nav.web = t.url } }
            }
        }
    }

    private func tile(_ simbolo: String, _ titulo: String, _ detalle: String, _ accion: @escaping () -> Void) -> some View {
        Button(action: accion) {
            VStack(alignment: .leading, spacing: 0) {
                Image(systemName: simbolo).font(.system(size: 17, weight: .semibold))
                    .foregroundStyle(paleta.tinta)
                    .frame(width: 36, height: 36)
                    .background(paleta.claro, in: RoundedRectangle(cornerRadius: 12, style: .continuous))
                Spacer(minLength: 8)
                Text(titulo).estilo(Estilo(tamano: 16, peso: .bold, relativo: .body)).foregroundStyle(C.texto)
                    .lineLimit(1).minimumScaleFactor(0.85)
                Text(detalle).estilo(.footnote).foregroundStyle(C.texto2).lineLimit(1).minimumScaleFactor(0.85)
            }
            .padding(12)
            .frame(maxWidth: .infinity, minHeight: 104, alignment: .leading)
            .relieve()
        }
        .buttonStyle(BotonHundido())
    }
}

/// «Para tu primera venta» (los primeros pasos de 2d con el estilo de ACero): el avance y cada paso
/// con su acción. Ningún paso es obligatorio: la tarjeta se puede cerrar y no vuelve.
struct PasosPrimeraVenta: View {
    var paleta: PaletaTienda
    @Environment(Estado.self) private var estado
    @Environment(Navegacion.self) private var nav
    @State private var conProductos: Bool?
    @State private var mpConectada = false
    @AppStorage(PrimerosPasos.claveCompartido) private var compartidos = ""
    @AppStorage(PrimerosPasos.claveOcultos) private var ocultos = ""

    var body: some View {
        let mp = mpConectada || (estado.resumen?.canales?.contains("mp") ?? false)
        let id = estado.tienda?.id ?? ""
        let compartio = PrimerosPasos.contiene(compartidos, id)
        let hechos = 1 + (conProductos == true ? 1 : 0) + (mp ? 1 : 0) + (compartio ? 1 : 0)
        VStack(alignment: .leading, spacing: 12) {
            HStack(alignment: .firstTextBaseline) {
                Text("Para tu primera venta").estilo(.seccion).foregroundStyle(C.texto).accessibilityAddTraits(.isHeader)
                Spacer(minLength: 8)
                Text("\(hechos) de 4").estilo(Estilo(tamano: 14, peso: .bold, relativo: .footnote)).foregroundStyle(paleta.tinta).monospacedDigit()
            }
            GeometryReader { g in
                Capsule().fill(C.pista)
                    .overlay(alignment: .leading) { Capsule().fill(paleta.acento).frame(width: g.size.width * CGFloat(hechos) / 4) }
            }
            .frame(height: 8)
            .accessibilityElement().accessibilityLabel("Avance").accessibilityValue("\(hechos) de 4")
            VStack(spacing: 0) {
                paso("Creaste tu tienda", hecho: true, accion: nil) {}
                paso("Cargá tu primer producto", hecho: conProductos == true, accion: "Cargar") { nav.inicio.append(.nuevoProducto(nil)) }
                paso("Conectá Mercado Pago", hecho: mp, accion: "Conectar") { nav.inicio.append(.integraciones) }
                ShareLink(item: estado.tienda?.url ?? URL(string: "https://tulocalweb.com.ar")!) {
                    filaPaso("Mandá el link a tus clientes", hecho: compartio, accion: "Compartir")
                }
                .buttonStyle(.plain)
                // Tocar «Compartir» alcanza para darlo por hecho (iOS no avisa si se mandó).
                .simultaneousGesture(TapGesture().onEnded { compartidos = PrimerosPasos.sumar(compartidos, id) })
            }
            Button(hechos == 4 ? "Listo, ir al tablero" : "Ocultar estos pasos") {
                withAnimation { ocultos = PrimerosPasos.sumar(ocultos, id) }
                estado.version += 1
            }
            .buttonStyle(BotonTexto(alto: 44, color: paleta.tinta))
        }
        .padding(16)
        .relieve()
        .task(id: estado.version) {
            struct MP: Decodable { var conectada: Bool? }
            if let m: MP = try? await estado.api.get("/v1/panel/integraciones/mercadopago") { mpConectada = m.conectada ?? false }
            if let l: PaginaProductos = try? await estado.api.get("/v1/panel/listas/productos", query: [URLQueryItem(name: "cuantos", value: "1")]) {
                conProductos = l.total > 0
            }
        }
    }

    private func paso(_ t: String, hecho: Bool, accion: String?, _ alTocar: @escaping () -> Void) -> some View {
        Button(action: alTocar) { filaPaso(t, hecho: hecho, accion: accion) }
            .buttonStyle(.plain).disabled(hecho || accion == nil)
    }

    private func filaPaso(_ t: String, hecho: Bool, accion: String?) -> some View {
        HStack(spacing: 12) {
            ZStack {
                if hecho {
                    Circle().fill(C.ok)
                    Image(systemName: "checkmark").font(.system(size: 12, weight: .heavy)).foregroundStyle(.white)
                } else {
                    Circle().fill(C.superficie)
                    Circle().strokeBorder(Color.tema(0xD8C9BB, 0x7D6656), lineWidth: 2)
                }
            }
            .frame(width: 26, height: 26)
            .accessibilityHidden(true)
            Text(t).estilo(.fila).foregroundStyle(hecho ? C.texto2 : C.texto)
                .frame(maxWidth: .infinity, alignment: .leading)
            if !hecho, let accion {
                Text(accion).estilo(Estilo(tamano: 14, peso: .bold, relativo: .footnote)).foregroundStyle(paleta.tinta)
            }
        }
        .frame(minHeight: 44)
        .contentShape(Rectangle())
        .accessibilityElement(children: .combine)
        .accessibilityValue(hecho ? "Hecho" : "")
    }
}

// MARK: - Para hacer hoy

/// «Para hacer hoy»: filas de 68 con la insignia del número y la acción a la derecha; la primera en el
/// acento. Si no hay nada urgente o la tienda no tuvo visitas hoy, invita a compartir.
struct ParaHacerHoy: View {
    var tablero: Tablero
    var masViejo: Date?
    var paleta: PaletaTienda
    var cero: Bool
    @Environment(Estado.self) private var estado
    @Environment(Navegacion.self) private var nav

    enum Tono { case claro, aviso, ok }

    struct Fila {
        var n: Int?
        var simbolo: String?
        var titulo: String
        var detalle: String
        var accion: String?
        var tono: Tono
        var ir: (() -> Void)?
    }

    private var filas: [Fila] {
        let ph = tablero.paraHacer
        var r: [Fila] = []
        if ph.preparar > 0 {
            r.append(Fila(n: ph.preparar, titulo: "Pedidos por preparar",
                          detalle: masViejo.map { "El más viejo entró \(Formato.hace($0))" } ?? "Tocá para verlos",
                          accion: "Preparar", tono: .claro) {
                nav.filtroPedidos = "Para preparar"
                nav.tocar(.pedidos)
            })
        }
        if ph.transferencias > 0 {
            r.append(Fila(n: ph.transferencias, titulo: "Transferencias por confirmar",
                          detalle: ph.transferenciasMonto.map { "\(Formato.pesos($0)) en total" } ?? "Revisá que hayan llegado",
                          accion: "Confirmar", tono: .claro) {
                nav.filtroPedidos = "Sin pagar"
                nav.tocar(.pedidos)
            })
        }
        if ph.sinStock > 0 {
            r.append(Fila(n: ph.sinStock, titulo: ph.sinStock == 1 ? "Producto sin stock" : "Productos sin stock",
                          detalle: "Siguen visibles en la tienda", accion: "Reponer", tono: .aviso) {
                nav.segmentoProductos = "out"
                nav.tocar(.productos)
            })
        }
        if let n = tablero.resenas?.nuevas.count, n > 0, estado.puede("clientes") {
            r.append(Fila(n: n, titulo: n == 1 ? "Reseña sin responder" : "Reseñas sin responder",
                          detalle: "Contestarlas suma confianza", accion: "Responder", tono: .claro) {
                nav.inicio.append(.resenas)
            })
        }
        // «Para hacer» nunca queda en «Todo al día» sin una acción (en cero, compartir ya está arriba).
        let hoy = Int(tablero.visitas?.hoy.actual ?? 0)
        if !cero, r.isEmpty || hoy == 0 {
            r.append(Fila(simbolo: "square.and.arrow.up", titulo: "Compartí tu tienda",
                          detalle: hoy == 0 ? "Hoy todavía no tuvo visitas" : "Más visitas, más ventas",
                          accion: "Compartir", tono: .claro) {
                nav.inicio.append(.compartir)
            })
        }
        return r
    }

    var body: some View {
        let lista = filas
        if !lista.isEmpty || !cero {
            VStack(alignment: .leading, spacing: 12) {
                TituloInicio(texto: "Para hacer hoy",
                             detalle: lista.isEmpty ? "Todo al día" : (lista.count == 1 ? "1 cosa" : "\(lista.count) cosas"))
                VStack(spacing: 0) {
                    if lista.isEmpty {
                        fila(Fila(simbolo: "checkmark", titulo: "Todo al día",
                                  detalle: "No hay pedidos ni transferencias esperando.", tono: .ok), primera: false)
                    }
                    ForEach(Array(lista.enumerated()), id: \.offset) { i, f in
                        if i > 0 { Separador(inset: 70) }
                        fila(f, primera: i == 0)
                    }
                }
                .relieve()
            }
        }
    }

    /// La insignia: la primera fila en el acento; las demás en el claro, aviso u ok.
    private func colores(_ tono: Tono, primera: Bool) -> (Color, Color) {
        if primera { return (paleta.acento, paleta.sobreAcento) }
        switch tono {
        case .claro: return (paleta.claro, paleta.tinta)
        case .aviso: return (C.avisoFondo, C.avisoTinta)
        case .ok: return (C.okFondo, C.okTinta)
        }
    }

    @ViewBuilder
    private func fila(_ f: Fila, primera: Bool) -> some View {
        let tono = colores(f.tono, primera: primera)
        let contenido = HStack(spacing: 14) {
            Group {
                if let n = f.n {
                    Text(Formato.numero(n)).font(Estilo(tamano: 17, peso: .bold, relativo: .headline).font).monospacedDigit()
                        .lineLimit(1).minimumScaleFactor(0.6)
                } else if let s = f.simbolo {
                    Image(systemName: s).font(.system(size: 16, weight: .bold))
                }
            }
            .foregroundStyle(tono.1)
            .frame(width: 40, height: 40)
            .background(tono.0, in: RoundedRectangle(cornerRadius: 12, style: .continuous))
            .overlay {
                if primera {
                    RoundedRectangle(cornerRadius: 12, style: .continuous)
                        .strokeBorder(LinearGradient(colors: [.white.opacity(0.25), .clear], startPoint: .top, endPoint: .center), lineWidth: 1)
                }
            }
            VStack(alignment: .leading, spacing: 2) {
                Text(f.titulo).estilo(.fila).foregroundStyle(C.texto).fixedSize(horizontal: false, vertical: true)
                Text(f.detalle).estilo(.footnote).foregroundStyle(C.texto2).monospacedDigit().fixedSize(horizontal: false, vertical: true)
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            if let a = f.accion {
                Text(a).estilo(Estilo(tamano: 14, peso: .bold, relativo: .footnote)).foregroundStyle(paleta.tinta)
            }
        }
        .padding(.horizontal, 16).padding(.vertical, 12)
        .frame(minHeight: 68)
        .contentShape(Rectangle())
        if let ir = f.ir {
            Button(action: ir) { contenido }.buttonStyle(.plain)
                .accessibilityElement(children: .combine)
        } else {
            contenido.accessibilityElement(children: .combine)
        }
    }
}

// MARK: - De un vistazo

/// «De un vistazo»: tarjetas de 164 × 156 que se deslizan de costado. Cada una aparece si tiene algo.
struct DeUnVistazo: View {
    var tablero: Tablero
    var periodo: String
    var paleta: PaletaTienda
    var editarMeta: () -> Void
    @Environment(Navegacion.self) private var nav
    @Environment(\.margen) private var margen

    var body: some View {
        let t = tablero
        let vivo = t.enVivo.flatMap { $0.personas > 0 || $0.carritos > 0 ? $0 : nil }
        let meta = t.montos ? t.meta : nil
        let caja = t.montos ? t.caja : nil
        let canales = t.montos && (t.ventas?.actual ?? 0) > 0 ? (t.canales ?? []) : []
        let liquidaciones = t.montos ? (t.liquidaciones ?? []) : []
        if vivo != nil || meta != nil || caja != nil || !canales.isEmpty || t.visitas != nil || !liquidaciones.isEmpty {
            VStack(alignment: .leading, spacing: 12) {
                TituloInicio(texto: "De un vistazo")
                ScrollView(.horizontal) {
                    HStack(spacing: 12) {
                        if let vivo { enVivo(vivo) }
                        if let meta { tarjetaMeta(meta) }
                        if let caja { tarjetaCaja(caja) }
                        if !canales.isEmpty { tarjetaCanales(canales) }
                        if let v = t.visitas { tarjetaVisitas(v) }
                        if let l = liquidaciones.first { tarjetaLiquidacion(l, mas: liquidaciones.count - 1) }
                    }
                    .padding(.horizontal, margen)
                }
                .scrollIndicators(.hidden)
                .scrollClipDisabled()
                .padding(.horizontal, -margen)
            }
        }
    }

    private func tarjeta<Contenido: View>(_ accion: @escaping () -> Void, @ViewBuilder _ contenido: () -> Contenido) -> some View {
        Button(action: accion) {
            VStack(alignment: .leading, spacing: 0) { contenido() }
                .foregroundStyle(C.texto)
                .padding(16)
                .frame(width: 164, height: 156, alignment: .topLeading)
                .relieve()
        }
        .buttonStyle(BotonHundido())
    }

    private func etiqueta(_ s: String, _ color: Color = C.texto2) -> some View {
        Text(s).estilo(.footnoteFuerte).foregroundStyle(color).lineLimit(1).minimumScaleFactor(0.85)
    }

    private func grande(_ s: String, _ color: Color = C.texto) -> some View {
        Text(s).estilo(Estilo(tamano: 26, peso: .bold, tracking: -0.023, relativo: .title2)).monospacedDigit()
            .foregroundStyle(color).lineLimit(1).minimumScaleFactor(0.6)
    }

    private func chip(_ s: String, _ fondo: Color, _ tinta: Color) -> some View {
        Text(s).estilo(Estilo(tamano: 12, peso: .bold, relativo: .caption)).foregroundStyle(tinta)
            .lineLimit(1).minimumScaleFactor(0.8)
            .padding(.horizontal, 8).padding(.vertical, 3)
            .background(fondo, in: Capsule())
    }

    private func enVivo(_ v: Tablero.EnVivo) -> some View {
        tarjeta({ if v.carritos > 0 { nav.inicio.append(.carritos) } else { nav.inicio.append(.estadisticas) } }) {
            HStack(spacing: 6) {
                Latido()
                etiqueta("Ahora en tu tienda", C.okTinta)
            }
            Spacer(minLength: 4)
            HStack(alignment: .firstTextBaseline, spacing: 6) {
                grande(Formato.numero(max(0, v.personas)))
                Text("mirando").estilo(.footnote).foregroundStyle(C.texto2)
            }
            Spacer(minLength: 4)
            if v.carritos > 0 {
                chip(v.carritos == 1 ? "1 carrito armándose" : "\(v.carritos) carritos armándose", C.avisoFondo, C.avisoTinta)
            } else if let p = v.paginas.first {
                Text("Miran \(p.nombre)").estilo(Estilo(tamano: 12, peso: .regular, relativo: .caption)).foregroundStyle(C.texto2).lineLimit(2)
            }
        }
    }

    private func tarjetaMeta(_ m: Tablero.Meta) -> some View {
        let pct = m.objetivo > 0 ? Double(m.actual) / Double(m.objetivo) : 0
        let suave = Color(hex: 0xCDB8A8)
        return Button(action: editarMeta) {
            VStack(alignment: .leading, spacing: 0) {
                etiqueta("Meta de \(Formato.mesDe(Date()))", suave)
                Spacer(minLength: 4)
                HStack(spacing: 12) {
                    ZStack {
                        Circle().stroke(Color(hex: 0x4A2D1F), lineWidth: 7)
                        Circle().trim(from: 0, to: min(1, pct))
                            .stroke(paleta.meta, style: StrokeStyle(lineWidth: 7, lineCap: .round))
                            .rotationEffect(.degrees(-90))
                    }
                    .padding(3.5)
                    .frame(width: 56, height: 56)
                    grande(m.objetivo > 0 ? Formato.porcentaje((pct * 100).rounded()) : "—", Color(hex: 0xFBF7F2))
                }
                .accessibilityElement()
                .accessibilityLabel(m.objetivo > 0 ? "\(Formato.porcentaje((pct * 100).rounded())) de la meta" : "Sin meta")
                Spacer(minLength: 4)
                Text(m.objetivo > 0 ? "\(Formato.compacto(m.actual)) de \(Formato.compacto(m.objetivo))" : "Tocá para ponerla")
                    .estilo(Estilo(tamano: 12, peso: .regular, relativo: .caption)).foregroundStyle(suave)
                    .monospacedDigit().lineLimit(1).minimumScaleFactor(0.8)
            }
            .padding(16)
            .frame(width: 164, height: 156, alignment: .topLeading)
            .background { Grano(opacidad: 0.14, mezcla: .normal) }
            .relieve(fondo: Color.tema(0x2A140C, 0x3A2216), luz: false)
        }
        .buttonStyle(BotonHundido())
        .accessibilityHint(m.objetivo > 0 ? "Editar la meta" : "Poner una meta")
    }

    private func tarjetaCaja(_ c: Tablero.Caja) -> some View {
        tarjeta({ nav.inicio.append(.caja) }) {
            etiqueta("Caja del local")
            Spacer(minLength: 4)
            grande(Formato.pesos(c.total))
            Spacer(minLength: 4)
            if c.total > 0 {
                chip("Cobrado hoy", C.okFondo, C.okTinta)
            } else {
                chip("Sin cobros hoy", C.neutroFondo, C.neutroTinta)
            }
        }
    }

    private func tarjetaCanales(_ canales: [Tablero.Canal]) -> some View {
        let partes = Self.partes(canales)
        let colores = [paleta.acento, paleta.medio, Color(hex: 0xE3A93B), Color(hex: 0xD8C9BB)]
        let textos = partes.map { "\($0.0) \(Formato.porcentaje(($0.1 * 100).rounded()))" }
        return tarjeta({ nav.inicio.append(.estadisticas) }) {
            etiqueta("Por dónde compran")
            Spacer(minLength: 4)
            GeometryReader { g in
                HStack(spacing: 2) {
                    ForEach(Array(partes.enumerated()), id: \.offset) { i, p in
                        Rectangle().fill(colores[i % colores.count])
                            .frame(width: max(3, (g.size.width - CGFloat(partes.count - 1) * 2) * p.1))
                    }
                }
            }
            .frame(height: 12)
            .clipShape(Capsule())
            Spacer(minLength: 4)
            Text([textos.prefix(2).joined(separator: " · "), textos.dropFirst(2).joined(separator: " · ")]
                .filter { !$0.isEmpty }.joined(separator: "\n"))
                .estilo(Estilo(tamano: 12, peso: .regular, relativo: .caption)).foregroundStyle(C.texto2)
                .monospacedDigit().lineSpacing(2).lineLimit(2).minimumScaleFactor(0.8)
        }
        .accessibilityElement(children: .combine)
        .accessibilityLabel("Por dónde compran: \(textos.joined(separator: ", "))")
    }

    /// Hasta cuatro partes: las tres primeras y el resto junto en «Otros».
    static func partes(_ canales: [Tablero.Canal]) -> [(String, Double)] {
        let orden = canales.sorted { $0.parte > $1.parte }
        if orden.count <= 4 { return orden.map { (Marcas.canalCorto($0.canal), $0.parte) } }
        return orden.prefix(3).map { (Marcas.canalCorto($0.canal), $0.parte) } + [("Otros", orden.dropFirst(3).map(\.parte).reduce(0, +))]
    }

    private func tarjetaVisitas(_ v: Tablero.Visitas) -> some View {
        let n = Int(v.visitantes.actual)
        let cuando = periodo == "hoy" ? "hoy" : (periodo == "7d" ? "en 7 días" : "en 30 días")
        return tarjeta({ v.desde == nil ? nav.inicio.append(.compartir) : nav.inicio.append(.estadisticas) }) {
            etiqueta("Tu tienda online")
            Spacer(minLength: 4)
            if v.desde == nil {
                Text(Formato.sinViudas("Todavía no tuvo visitas"))
                    .estilo(Estilo(tamano: 17, peso: .bold, relativo: .headline)).fixedSize(horizontal: false, vertical: true)
                Spacer(minLength: 4)
                chip("Compartila", paleta.claro, paleta.tinta)
            } else {
                grande(Formato.numero(n))
                Text("\(n == 1 ? "visita" : "visitas") \(cuando)").estilo(.footnote).foregroundStyle(C.texto2)
                Spacer(minLength: 4)
                if let d = v.visitantes.variacion {
                    let r = Int(abs(d).rounded())
                    chip("\(d >= 0 ? "▲" : "▼") \(Formato.porcentaje(Double(r)))", d >= 0 ? C.okFondo : C.errorFondo, d >= 0 ? C.okTinta : C.errorTinta)
                }
            }
        }
    }

    private func tarjetaLiquidacion(_ l: Tablero.Liquidacion, mas: Int) -> some View {
        tarjeta({ nav.inicio.append(.finanzas) }) {
            etiqueta(mas > 0 ? "Próximas liquidaciones" : "Próxima liquidación")
            Spacer(minLength: 4)
            grande(Formato.pesos(l.monto), C.okTinta)
            Spacer(minLength: 4)
            Text("\(Marcas.pago(l.medio)) · \(Fecha.iso(l.fecha).map { Formato.dia($0) } ?? l.fecha)")
                .estilo(Estilo(tamano: 12, peso: .regular, relativo: .caption)).foregroundStyle(C.texto2).lineLimit(1).minimumScaleFactor(0.8)
            if mas > 0 {
                Text(mas == 1 ? "y 1 más" : "y \(mas) más").estilo(Estilo(tamano: 12, peso: .regular, relativo: .caption)).foregroundStyle(C.texto2)
            }
        }
    }
}

/// El punto verde que late («En vivo»).
struct Latido: View {
    @State private var latido = false
    @Environment(\.accessibilityReduceMotion) private var reducir
    var body: some View {
        Circle().fill(C.ok).frame(width: 8, height: 8)
            .scaleEffect(latido ? 1.35 : 0.85).opacity(latido ? 0.6 : 1)
            .animation(reducir ? nil : .easeInOut(duration: 0.9).repeatForever(autoreverses: true), value: latido)
            .onAppear { latido = true }
            .accessibilityHidden(true)
    }
}

// MARK: - Últimos pedidos y lo más vendido

struct UltimosPedidos: View {
    var tablero: Tablero
    var paleta: PaletaTienda
    @Environment(Navegacion.self) private var nav

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            TituloInicio(texto: "Últimos pedidos", accion: "Ver todos", tinta: paleta.tinta) {
                nav.filtroPedidos = nil
                nav.tocar(.pedidos)
            }
            VStack(spacing: 0) {
                ForEach(Array(tablero.ultimosPedidos.prefix(5).enumerated()), id: \.element.id) { i, p in
                    if i > 0 { Separador(inset: 68) }
                    Button {
                        nav.pestana = .pedidos
                        nav.pedidos = [.pedido(p.numero)]
                    } label: {
                        HStack(spacing: 12) {
                            AvatarIniciales(nombre: p.cliente, tam: 40, fondo: paleta.claro, tinta: paleta.tinta)
                            VStack(alignment: .leading, spacing: 2) {
                                Text(p.cliente).estilo(.fila).foregroundStyle(C.texto).lineLimit(1)
                                Text("#\(p.numero) · \(Marcas.canalCorto(p.canal)) · \(Formato.hace(p.creado))")
                                    .estilo(.footnote).foregroundStyle(C.texto2).monospacedDigit().lineLimit(1)
                            }
                            Spacer(minLength: 4)
                            VStack(alignment: .trailing, spacing: 4) {
                                if let total = p.total {
                                    Text(Formato.pesos(total)).estilo(Estilo(tamano: 15, peso: .bold, relativo: .subheadline))
                                        .monospacedDigit().foregroundStyle(C.texto)
                                }
                                ChipEstado(estado: p.estado)
                            }
                        }
                        .padding(.horizontal, 16).padding(.vertical, 12)
                        .frame(minHeight: 68)
                        .contentShape(Rectangle())
                    }
                    .buttonStyle(.plain)
                    .accessibilityElement(children: .combine)
                }
            }
            .relieve()
        }
    }
}

struct LoMasVendido: View {
    var tablero: Tablero
    @Environment(Navegacion.self) private var nav
    @Environment(\.margen) private var margen

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            TituloInicio(texto: "Lo más vendido",
                         detalle: tablero.periodo == "hoy" ? "Hoy" : (tablero.periodo == "7d" ? "En 7 días" : "En 30 días"))
            ScrollView(.horizontal) {
                HStack(alignment: .top, spacing: 12) {
                    ForEach(Array(tablero.masVendidos.enumerated()), id: \.element.id) { i, p in
                        Button { nav.ir(.producto(p.productoId)) } label: {
                            VStack(alignment: .leading, spacing: 0) {
                                Foto(foto: p.foto, ancho: 450)
                                    .frame(width: 148, height: 112)
                                    .overlay(alignment: .topLeading) {
                                        Text("#\(i + 1)").estilo(Estilo(tamano: 12, peso: .bold, relativo: .caption))
                                            .foregroundStyle(Color(hex: 0x2A140C))
                                            .padding(.horizontal, 8).frame(height: 24)
                                            .background(Color(hex: 0xFFFDF9, alfa: 0.92), in: Capsule())
                                            .padding(8)
                                    }
                                VStack(alignment: .leading, spacing: 2) {
                                    Text(p.nombre).estilo(Estilo(tamano: 14, peso: .semibold, relativo: .footnote))
                                        .foregroundStyle(C.texto).lineLimit(2, reservesSpace: true)
                                    Text(Formato.cantidad(p.unidades, "vendido", "vendidos")).estilo(.footnote)
                                        .foregroundStyle(C.texto2).monospacedDigit()
                                }
                                .padding(.horizontal, 12).padding(.top, 10).padding(.bottom, 12)
                            }
                            .frame(width: 148, alignment: .leading)
                            .relieve()
                        }
                        .buttonStyle(BotonHundido())
                        .accessibilityElement(children: .combine)
                        .accessibilityLabel("Puesto \(i + 1): \(p.nombre), \(Formato.cantidad(p.unidades, "vendido", "vendidos"))")
                    }
                }
                .padding(.horizontal, margen)
            }
            .scrollIndicators(.hidden)
            .scrollClipDisabled()
            .padding(.horizontal, -margen)
        }
    }
}

// MARK: - Al final

/// «Más vistos»: los productos que más miraron en la tienda online.
struct TarjetaMasVistos: View {
    var visitas: Tablero.Visitas
    var paleta: PaletaTienda

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            TituloInicio(texto: "Más vistos", detalle: "En la tienda online")
            let max = Double(visitas.masVistos.map(\.vistas).max() ?? 1)
            VStack(spacing: 0) {
                ForEach(Array(visitas.masVistos.prefix(5).enumerated()), id: \.element.id) { i, p in
                    if i > 0 { Separador(inset: 70) }
                    HStack(spacing: 12) {
                        Miniatura(foto: p.foto, tam: 42, radio: 12)
                        VStack(alignment: .leading, spacing: 6) {
                            Text(p.nombre).estilo(.fila).foregroundStyle(C.texto).lineLimit(1)
                            GeometryReader { g in
                                Capsule().fill(paleta.acento.opacity(0.55)).frame(width: g.size.width * CGFloat(Double(p.vistas) / max))
                            }
                            .frame(height: 5)
                        }
                        Text(Formato.numero(p.vistas)).estilo(.fila).monospacedDigit().foregroundStyle(C.texto2)
                    }
                    .padding(.horizontal, 16).padding(.vertical, 12)
                    .frame(minHeight: 68)
                    .accessibilityElement(children: .combine)
                }
            }
            .relieve()
        }
    }
}

/// «Reseñas nuevas»: las últimas sin responder.
struct TarjetaResenas: View {
    var resenas: Tablero.Resenas
    var paleta: PaletaTienda
    @Environment(Navegacion.self) private var nav

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            TituloInicio(texto: "Reseñas nuevas", accion: "Responder", tinta: paleta.tinta) { nav.inicio.append(.resenas) }
            VStack(alignment: .leading, spacing: 0) {
                if let p = resenas.promedio {
                    HStack(spacing: 6) {
                        Text(p.formatted(.number.precision(.fractionLength(1)).locale(Locale(identifier: "es_AR"))))
                            .estilo(.headline).foregroundStyle(C.texto)
                        Estrellas(n: Int(p.rounded()))
                        Text("promedio").estilo(.footnote).foregroundStyle(C.texto2)
                    }
                    .padding(.horizontal, 16).padding(.vertical, 14)
                }
                ForEach(Array(resenas.nuevas.prefix(3).enumerated()), id: \.offset) { i, r in
                    if i > 0 || resenas.promedio != nil { Separador(inset: 16) }
                    VStack(alignment: .leading, spacing: 4) {
                        HStack(spacing: 6) {
                            Estrellas(n: r["estrellas"]?.entero ?? 0)
                            Text(r["autor"]?.texto ?? "").estilo(.footnoteFuerte).foregroundStyle(C.texto)
                            Spacer()
                            Text(r["producto"]?.texto ?? "").estilo(.footnote).foregroundStyle(C.texto2).lineLimit(1)
                        }
                        Text("«\(r["texto"]?.texto ?? "")»").estilo(.texto14).foregroundStyle(C.texto).lineLimit(3)
                    }
                    .padding(.horizontal, 16).padding(.vertical, 12)
                }
            }
            .relieve()
        }
    }
}

/// «Envíos de hoy»: paquetes y retiros de los transportes.
struct TarjetaEnvios: View {
    var envios: Tablero.Envios
    var paleta: PaletaTienda
    @Environment(Navegacion.self) private var nav

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            TituloInicio(texto: "Envíos de hoy", accion: "Preparar", tinta: paleta.tinta) {
                nav.filtroPedidos = "Para preparar"
                nav.pestana = .pedidos
            }
            VStack(spacing: 0) {
                HStack(spacing: 12) {
                    Image(systemName: "shippingbox.fill").font(.system(size: 16, weight: .semibold)).foregroundStyle(paleta.tinta)
                        .frame(width: 40, height: 40)
                        .background(paleta.claro, in: RoundedRectangle(cornerRadius: 12, style: .continuous))
                    Text(Formato.cantidad(envios.paquetes, "paquete para despachar", "paquetes para despachar")).estilo(.fila).foregroundStyle(C.texto)
                    Spacer(minLength: 0)
                }
                .padding(.horizontal, 16).padding(.vertical, 12)
                ForEach(Array(envios.retiros.enumerated()), id: \.offset) { _, r in
                    Separador(inset: 16)
                    HStack {
                        Text(r.hora).estilo(.fila).monospacedDigit().foregroundStyle(r.proximo ? paleta.tinta : C.texto)
                            .frame(width: 54, alignment: .leading)
                        VStack(alignment: .leading, spacing: 1) {
                            Text(Formato.mayuscula(r.transporte)).estilo(.texto14).foregroundStyle(C.texto)
                            if let c = r.ciudad { Text(c).estilo(.footnote).foregroundStyle(C.texto2) }
                        }
                        Spacer()
                        Text(Formato.cantidad(r.paquetes, "paquete", "paquetes")).estilo(.footnote).foregroundStyle(C.texto2)
                        if r.proximo { ChipSimple(texto: "Próximo", fondo: C.avisoFondo, tinta: C.avisoTinta) }
                    }
                    .padding(.horizontal, 16).padding(.vertical, 12)
                }
            }
            .relieve()
        }
    }
}

/// «Integraciones»: lo conectado (Mercado Pago, Mercado Libre, Correo) y un acceso para sumar más.
struct TarjetaIntegraciones: View {
    var paleta: PaletaTienda
    @Environment(Estado.self) private var estado
    @Environment(Navegacion.self) private var nav
    @State private var conectadas: [(String, String)]?

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            TituloInicio(texto: "Integraciones", accion: "Explorar", tinta: paleta.tinta) { nav.inicio.append(.integraciones) }
            VStack(spacing: 0) {
                if let c = conectadas, !c.isEmpty {
                    ForEach(Array(c.enumerated()), id: \.offset) { i, x in
                        if i > 0 { Separador(inset: 16) }
                        HStack {
                            Text(x.0).estilo(.fila).foregroundStyle(C.texto)
                            Spacer()
                            Text(x.1).estilo(.footnote).foregroundStyle(C.texto2).lineLimit(1)
                            Circle().fill(C.ok).frame(width: 8, height: 8)
                        }
                        .padding(.horizontal, 16).padding(.vertical, 14)
                    }
                } else if conectadas != nil {
                    Button { nav.inicio.append(.integraciones) } label: {
                        Text(Formato.sinViudas("Conectá Mercado Pago, Mercado Libre y Correo Argentino para cobrar, publicar y enviar."))
                            .estilo(.texto14).foregroundStyle(C.texto2).multilineTextAlignment(.leading)
                            .padding(16).frame(maxWidth: .infinity, alignment: .leading)
                    }
                    .buttonStyle(.plain)
                } else {
                    ProgressView().tint(paleta.acento).padding(16)
                }
            }
            .relieve()
        }
        .task { await cargar() }
    }

    private func cargar() async {
        guard conectadas == nil else { return }
        struct I: Decodable { var conectada: Bool?; var usuario: String?; var apodo: String?; var email: String? }
        var c: [(String, String)] = []
        for (id, nombre) in [("mercadopago", "Mercado Pago"), ("mercadolibre", "Mercado Libre"), ("correo", "Correo Argentino")] {
            if let i: I = try? await estado.api.get("/v1/panel/integraciones/\(id)"), i.conectada == true {
                c.append((nombre, i.usuario ?? i.apodo ?? i.email ?? "Conectada"))
            }
        }
        conectadas = c
    }
}

// MARK: - Cargando (2c)

/// El esqueleto del Inicio, con la misma forma: la tarjeta de Tincho, el bento y una lista.
struct EsqueletoInicio: View {
    var paleta: PaletaTienda
    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            HStack(alignment: .top, spacing: 12) {
                Esqueleto(ancho: 48, alto: 48, radio: 24)
                VStack(alignment: .leading, spacing: 8) {
                    Esqueleto(alto: 14)
                    Esqueleto(ancho: 180, alto: 14, suave: true)
                    Esqueleto(ancho: 120, alto: 36, radio: 18)
                }
            }
            .padding(14)
            .relieve(fondo: Color.tema(0xFFFDF9, 0x2E1B12), fuerte: true)
            .padding(.horizontal, 16).padding(.top, -28)
            VStack(alignment: .leading, spacing: 32) {
                HStack(spacing: 12) {
                    Color.clear.relieve().overlay(alignment: .top) { Esqueleto(alto: 150, radio: 16, suave: true).padding(12) }
                    VStack(spacing: 12) {
                        Color.clear.relieve()
                        Color.clear.relieve()
                    }
                }
                .frame(height: 252)
                VStack(alignment: .leading, spacing: 12) {
                    Esqueleto(ancho: 150, alto: 22)
                    EsqueletoFilas(filas: 3, miniatura: 40)
                }
            }
            .padding(.top, 24)
            .margenPantalla()
        }
        .accessibilityElement()
        .accessibilityLabel("Cargando cómo viene el día")
    }
}
