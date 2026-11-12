import SwiftUI

// La vidriera del Inicio (AFinal, ACero y AOscuro de docs/handoff/inicio-a): el color de la tienda
// con grano y dos halos, el logo, el estado, buscar y avisos en vidrio, el período, lo vendido con
// el gráfico de hoy contra ayer y «Vender ahora». Al bajar se pliega a BarraVidriera (AScroll). La
// tarjeta de Tincho va montada sobre su borde.

/// La vidriera: arriba de todo en el Inicio, con lo que haya (cargando, error o el tablero).
struct Vidriera: View {
    var carga: Carga<Tablero>
    @Binding var periodo: String
    var sinLeer: Int
    var paleta: PaletaTienda
    var se: Bool
    @Environment(Estado.self) private var estado
    @Environment(Navegacion.self) private var nav
    @Environment(\.margen) private var margen

    /// El día en $0 (o sin pedidos para quien no ve montos): la vidriera invita a compartir (ACero).
    static func enCero(_ t: Tablero) -> Bool {
        t.montos ? (t.ventas?.actual ?? 0) == 0 : t.kpis.pedidos.actual == 0
    }

    var body: some View {
        let t = carga.valor
        let cero = t.map(Self.enCero) ?? false
        let conTarjeta = (t != nil && !cero) || t == nil && carga.esCargando
        VStack(alignment: .leading, spacing: 16) {
            arriba(t)
            SelectorPeriodo(periodo: $periodo, paleta: paleta)
            switch carga {
            case .cargando: esqueleto
            case .error: EmptyView()
            case let .listo(t):
                if cero { enCero(t) } else { conVentas(t) }
            }
        }
        .padding(.horizontal, margen)
        .padding(.top, 12)
        .padding(.bottom, conTarjeta ? 48 : 24)
        .foregroundStyle(paleta.sobre)
        .background { FondoVidriera(paleta: paleta) }
    }

    private var nombre: String {
        (carga.valor?.hola ?? estado.datos?.nombreDePila ?? "").trimmingCharacters(in: .whitespaces)
    }

    private var hola: String { nombre.isEmpty ? "Hola" : "Hola, \(nombre)" }

    private var cuando: String {
        periodo == "hoy" ? "Hoy" : (periodo == "7d" ? "En 7 días" : "En 30 días")
    }

    // MARK: Arriba

    private func arriba(_ t: Tablero?) -> some View {
        HStack(spacing: 12) {
            Button { nav.tocar(.mas) } label: { LogoTienda(tienda: estado.tienda, tam: 48, radio: 16, paleta: paleta) }
                .buttonStyle(BotonHundido())
                .accessibilityLabel("Tu tienda, \(estado.tienda?.nombre ?? "")")
            VStack(alignment: .leading, spacing: 2) {
                Text(estado.tienda?.nombre ?? "Tu tienda")
                    .estilo(Estilo(tamano: 22, peso: .bold, tracking: -0.018, relativo: .title2))
                    .lineLimit(1).minimumScaleFactor(0.75)
                HStack(spacing: 6) {
                    Circle().fill(Color(hex: 0x57D9A0)).frame(width: 8, height: 8)
                        .background(Circle().fill(Color(hex: 0x57D9A0, alfa: 0.28)).padding(-3))
                    Text(Self.estadoTienda(t))
                        .estilo(Estilo(tamano: 13, peso: .medium, relativo: .footnote))
                        .foregroundStyle(paleta.sobre2).monospacedDigit()
                        .lineLimit(1).minimumScaleFactor(0.85)
                }
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            .accessibilityElement(children: .combine)
            BotonVidrio(simbolo: "magnifyingglass", etiqueta: "Buscar", paleta: paleta) { nav.buscar = true }
            BotonVidrio(simbolo: "bell", etiqueta: sinLeer > 0 ? "Avisos, \(sinLeer) sin leer" : "Avisos", punto: sinLeer > 0, paleta: paleta) {
                nav.inicio.append(.avisos)
            }
        }
    }

    /// «Abierta · 3 mirando ahora» si hay gente en vivo; si no, las visitas de hoy.
    static func estadoTienda(_ t: Tablero?) -> String {
        guard let t else { return "Abierta" }
        if let v = t.enVivo, v.personas > 0 { return "Abierta · \(Formato.numero(v.personas)) mirando ahora" }
        guard let v = t.visitas, v.desde != nil else { return "Abierta" }
        let n = Int(v.hoy.actual)
        if n == 0 { return "Abierta · todavía sin visitas hoy" }
        return "Abierta · \(n == 1 ? "1 visita" : "\(Formato.numero(n)) visitas") hoy"
    }

    // MARK: Con ventas

    @ViewBuilder
    private func conVentas(_ t: Tablero) -> some View {
        VStack(alignment: .leading, spacing: 6) {
            Text("\(hola). \(cuando) \(t.montos ? "vendiste" : "tuviste")")
                .estilo(Estilo(tamano: 15, peso: .medium, relativo: .subheadline)).foregroundStyle(paleta.sobre2)
            Text(Self.monto(t))
                .estilo(se ? Estilo(tamano: 40, peso: .bold, tracking: -0.039, relativo: .largeTitle)
                           : Estilo(tamano: 46, peso: .bold, tracking: -0.039, relativo: .largeTitle))
                .monospacedDigit().lineLimit(1).minimumScaleFactor(0.6)
                .contentTransition(.numericText())
                .animation(Mov.toque, value: Self.monto(t))
            lineaDelta(t).padding(.top, 2)
        }
        .accessibilityElement(children: .combine)
        if t.montos, let s = t.serie, let v = t.ventas { grafico(t, s, v) }
        BotonVidriera(texto: "Vender ahora", simbolo: "barcode.viewfinder", paleta: paleta) { nav.venta = true }
    }

    /// Lo vendido; quien no ve montos ve los pedidos.
    static func monto(_ t: Tablero) -> String {
        if t.montos, let v = t.ventas { return Formato.pesos(v.actual) }
        return Formato.cantidad(Int(t.kpis.pedidos.actual), "pedido", "pedidos")
    }

    private var contra: String {
        periodo == "hoy" ? "vs. ayer" : (periodo == "7d" ? "vs. 7 días antes" : "vs. 30 días antes")
    }

    private func lineaDelta(_ t: Tablero) -> some View {
        let c = t.montos ? t.ventas : t.kpis.pedidos
        let pedidos = Formato.cantidad(Int(t.kpis.pedidos.actual), "pedido", "pedidos")
        let detalle: String = if t.montos, let tk = t.kpis.ticket, tk.actual > 0 {
            "\(pedidos) · ticket \(Formato.pesos(tk.actual))"
        } else if t.montos {
            pedidos
        } else {
            Formato.cantidad(Int(t.kpis.clientesNuevos.actual), "cliente nuevo", "clientes nuevos")
        }
        return ViewThatFits(in: .horizontal) {
            HStack(spacing: 8) { pildora(c); Text(detalle).lineLimit(1) }
            VStack(alignment: .leading, spacing: 6) { pildora(c); Text(detalle) }
        }
        .estilo(Estilo(tamano: 14, peso: .medium, relativo: .footnote)).monospacedDigit()
    }

    @ViewBuilder
    private func pildora(_ c: Comparado?) -> some View {
        if let c, let d = c.variacion {
            let r = Int(abs(d).rounded())
            Text("\(d >= 0 ? "▲" : "▼") \(Formato.porcentaje(Double(r))) \(contra)")
                .fontWeight(.bold).lineLimit(1)
                .padding(.horizontal, 9).padding(.vertical, 3)
                .vidrio(Capsule(), paleta)
                .accessibilityLabel("\(d >= 0 ? "Subió" : "Bajó") \(Formato.porcentaje(Double(r))) \(contra)")
        } else if let c, c.actual > 0 {
            Text("Primeras ventas").fontWeight(.bold).lineLimit(1)
                .padding(.horizontal, 9).padding(.vertical, 3)
                .vidrio(Capsule(), paleta)
        }
    }

    private func grafico(_ t: Tablero, _ s: Tablero.Serie, _ v: Comparado) -> some View {
        let hoy = periodo == "hoy"
        let colores = ColoresGrafico(linea: paleta.linea, anterior: paleta.sobreSuave, relleno: paleta.velo,
                                     guia: paleta.borde, bordePunto: paleta.linea.opacity(0.25))
        let leyenda = HStack(spacing: 12) {
            HStack(spacing: 6) {
                Capsule().fill(paleta.linea).frame(width: 14, height: 3)
                Text(hoy ? "Hoy" : "Este período")
            }
            HStack(spacing: 6) {
                Punteado(color: paleta.sobreSuave).frame(width: 14, height: 2)
                Text("\(hoy ? "Ayer" : "Antes") \(Formato.pesos(v.anterior))")
            }
        }
        return VStack(alignment: .leading, spacing: 6) {
            GraficoVentas(actual: s.actual, anterior: s.anterior, alto: se ? 64 : 88,
                          etiqueta: GraficoVentas.etiquetas(desde: t.desde, hasta: t.hasta, puntos: s.actual.count, porHora: hoy),
                          colores: colores,
                          descripcion: "Ventas \(hoy ? "de hoy por hora" : "del período") comparadas con \(hoy ? "ayer" : "el período anterior"): \(Formato.pesos(v.actual)) contra \(Formato.pesos(v.anterior))")
                .padding(.horizontal, -2)
            ViewThatFits(in: .horizontal) {
                HStack {
                    leyenda
                    Spacer(minLength: 8)
                    Text(Self.marcas(t, porHora: hoy)).lineLimit(1)
                }
                leyenda
            }
            .estilo(Estilo(tamano: 12, peso: .medium, relativo: .caption))
            .foregroundStyle(paleta.sobre2).monospacedDigit()
            .accessibilityHidden(true)
        }
    }

    /// «8 h · 12 h · 16 h · 20 h» (hoy) o «5/11 · 7/11 · …»: cuatro marcas parejas del período.
    static func marcas(_ t: Tablero, porHora: Bool) -> String {
        let total = t.hasta.timeIntervalSince(t.desde)
        let c = Formato.calendario
        return (0..<4).map { i in
            let f = t.desde.addingTimeInterval(total * (Double(i) * 2 + 1) / 8)
            return porHora ? "\(c.component(.hour, from: f))\(Formato.fino)h" : "\(c.component(.day, from: f))/\(c.component(.month, from: f))"
        }
        .joined(separator: " · ")
    }

    // MARK: En cero (ACero)

    @ViewBuilder
    private func enCero(_ t: Tablero) -> some View {
        VStack(alignment: .leading, spacing: 4) {
            Text(hola).estilo(Estilo(tamano: 15, peso: .medium, relativo: .subheadline)).foregroundStyle(paleta.sobre2)
            Text(Formato.sinViudas(titularCero(t)))
                .estilo(Estilo(tamano: se ? 26 : 30, peso: .bold, tracking: -0.027, relativo: .title))
                .lineSpacing(0)
                .fixedSize(horizontal: false, vertical: true)
        }
        .accessibilityElement(children: .combine)
        BotonVidriera(texto: "Compartir tu vidriera", simbolo: "square.and.arrow.up", paleta: paleta) { nav.inicio.append(.compartir) }
    }

    private func titularCero(_ t: Tablero) -> String {
        guard let v = t.visitas, v.desde != nil else { return "Tu local está abierto. Compartilo para recibir las primeras visitas." }
        let n = Int(periodo == "hoy" ? v.hoy.actual : v.visitantes.actual)
        if n == 0 { return "Tu local está abierto. \(cuando) todavía no te visitaron: compartilo." }
        return "Tu local está abierto. \(cuando) te \(n == 1 ? "visitó 1 persona" : "visitaron \(Formato.numero(n)) personas")."
    }

    // MARK: Cargando

    private var esqueleto: some View {
        VStack(alignment: .leading, spacing: 16) {
            VStack(alignment: .leading, spacing: 10) {
                bloque(150, 15)
                bloque(230, 46, radio: 12)
                bloque(190, 22)
            }
            bloque(nil, se ? 64 : 88, radio: 14)
            RoundedRectangle(cornerRadius: 20, style: .continuous).fill(paleta.velo).frame(height: 60)
        }
        .accessibilityElement()
        .accessibilityLabel("Cargando")
    }

    private func bloque(_ ancho: CGFloat?, _ alto: CGFloat, radio: CGFloat? = nil) -> some View {
        RoundedRectangle(cornerRadius: radio ?? alto / 2, style: .continuous).fill(paleta.velo)
            .frame(width: ancho, height: alto)
    }
}

extension Carga {
    var esCargando: Bool {
        if case .cargando = self { return true }
        return false
    }
}

// MARK: - Fondo

/// El rectángulo de la vidriera, redondeado abajo y estirado 600 hacia arriba: cubre la barra de
/// estado y lo que asoma al tirar para refrescar.
struct FormaVidriera: Shape {
    var radio: CGFloat = 32
    func path(in r: CGRect) -> Path {
        UnevenRoundedRectangle(bottomLeadingRadius: radio, bottomTrailingRadius: radio, style: .continuous)
            .path(in: CGRect(x: r.minX, y: r.minY - 600, width: r.width, height: r.height + 600))
    }
}

/// El fondo de la vidriera: el color (en oscuro, un degradé hacia el fondo), dos halos y el grano.
struct FondoVidriera: View {
    var paleta: PaletaTienda
    var radio: CGFloat = 32
    var halos = true
    var body: some View {
        LinearGradient(colors: [paleta.fondoArriba, paleta.fondoAbajo], startPoint: UnitPoint(x: 0.32, y: 0), endPoint: UnitPoint(x: 0.68, y: 1))
            .padding(.top, -600)
            .overlay(alignment: .topTrailing) {
                if halos { halo(paleta.luz, 248).offset(x: 72, y: -100) }
            }
            .overlay(alignment: .bottomLeading) {
                if halos { halo(paleta.sombraHalo, 224).offset(x: -64, y: 112) }
            }
            .overlay { Grano(opacidad: paleta.oscuro ? 0.18 : 0.2).padding(.top, -600) }
            .clipShape(FormaVidriera(radio: radio))
            .accessibilityHidden(true)
    }

    private func halo(_ color: Color, _ tam: CGFloat) -> some View {
        Circle().fill(RadialGradient(colors: [color, color.opacity(0)], center: .center, startRadius: 0, endRadius: tam / 2))
            .frame(width: tam, height: tam)
    }
}

// MARK: - Piezas de la vidriera

/// El logo de la tienda (Tienda.logo, entero sobre blanco) o su inicial; en oscuro, la inicial va en
/// el acento.
struct LogoTienda: View {
    var tienda: Acceso.Tienda?
    var tam: CGFloat
    var radio: CGFloat
    var paleta: PaletaTienda

    var body: some View {
        let forma = RoundedRectangle(cornerRadius: radio, style: .continuous)
        Group {
            if let u = urlDeFoto(tienda?.logo, ancho: Int(tam * 3)) {
                AsyncImage(url: u) { fase in
                    if let img = fase.image {
                        img.resizable().scaledToFit().padding(tam * 0.1)
                            .frame(width: tam, height: tam).background(Color.white)
                    } else {
                        inicial
                    }
                }
            } else {
                inicial
            }
        }
        .frame(width: tam, height: tam)
        .clipShape(forma)
        .overlay(forma.strokeBorder(.white.opacity(0.6), lineWidth: 1).mask(LinearGradient(colors: [.white, .clear], startPoint: .top, endPoint: .center)))
        .shadow(color: .black.opacity(0.3), radius: 8, y: 6)
        .accessibilityHidden(true)
    }

    private var inicial: some View {
        Text(String(Formato.iniciales(tienda?.nombre ?? "").prefix(1)))
            .font(Estilo(tamano: tam * 0.46, peso: .bold, relativo: .title2).font)
            .foregroundStyle(paleta.oscuro ? paleta.sobreAcento : paleta.tintaBlanco)
            .frame(width: tam, height: tam)
            .background(paleta.oscuro ? paleta.acento : .white)
    }
}

/// Botón redondo de 44 en vidrio (buscar, avisos), con el punto de «sin leer».
struct BotonVidrio: View {
    var simbolo: String
    var etiqueta: String
    var punto = false
    var paleta: PaletaTienda
    var accion: () -> Void
    var body: some View {
        Button(action: accion) {
            Image(systemName: simbolo).font(.system(size: 18, weight: .semibold))
                .foregroundStyle(paleta.sobre)
                .frame(width: 44, height: 44)
                .vidrio(Circle(), paleta)
                .overlay(alignment: .topTrailing) {
                    if punto {
                        Circle().fill(paleta.oscuro ? paleta.tinta : Color(hex: 0xFFC24D)).frame(width: 9, height: 9)
                            .overlay(Circle().strokeBorder(paleta.fondoArriba, lineWidth: 2).padding(-2))
                            .offset(x: -10, y: 9)
                    }
                }
                .contentShape(Circle())
        }
        .buttonStyle(BotonHundido())
        .accessibilityLabel(etiqueta)
    }
}

/// Hoy / 7 días / 30 días en vidrio (reemplaza al Segmentado en el Inicio). 44 de alto con el vidrio.
struct SelectorPeriodo: View {
    @Binding var periodo: String
    var paleta: PaletaTienda
    @Namespace private var ns
    private let opciones = [("hoy", "Hoy"), ("7d", "7 días"), ("30d", "30 días")]

    var body: some View {
        HStack(spacing: 2) {
            ForEach(opciones, id: \.0) { op in
                let activo = op.0 == periodo
                Button {
                    withAnimation(.snappy(duration: 0.22)) { periodo = op.0 }
                    Haptico.seleccion()
                } label: {
                    Text(op.1)
                        .estilo(Estilo(tamano: 14, peso: activo ? .bold : .semibold, relativo: .footnote))
                        .foregroundStyle(activo ? paleta.tintaBlanco : paleta.sobre)
                        .lineLimit(1)
                        .padding(.horizontal, 14).frame(height: 38)
                        .background {
                            if activo {
                                RoundedRectangle(cornerRadius: 11, style: .continuous).fill(.white)
                                    .shadow(color: .black.opacity(0.25), radius: 3, y: 2)
                                    .matchedGeometryEffect(id: "periodo", in: ns)
                            }
                        }
                        .contentShape(Rectangle())
                }
                .buttonStyle(.plain)
                .accessibilityAddTraits(activo ? [.isSelected] : [])
            }
        }
        .padding(3)
        .vidrio(RoundedRectangle(cornerRadius: 14, style: .continuous), paleta)
        .accessibilityElement(children: .contain)
        .accessibilityLabel("Período")
    }
}

/// El botón grande de la vidriera: blanco con un degradé sutil, relieve, grano y sombra; en oscuro, el
/// acento con su texto.
struct BotonVidriera: View {
    var texto: String
    var simbolo: String
    var paleta: PaletaTienda
    var alto: CGFloat = 60
    var accion: () -> Void

    var body: some View {
        let forma = RoundedRectangle(cornerRadius: 20, style: .continuous)
        Button(action: accion) {
            HStack(spacing: 10) {
                Image(systemName: simbolo).font(.system(size: 20, weight: .semibold))
                Text(texto).estilo(Estilo(tamano: 18, peso: .bold, tracking: -0.011, relativo: .headline))
                    .lineLimit(1).minimumScaleFactor(0.8)
            }
            .foregroundStyle(paleta.oscuro ? paleta.sobreAcento : paleta.tintaBlanco)
            .frame(maxWidth: .infinity).frame(height: alto)
            .background {
                if paleta.oscuro {
                    forma.fill(paleta.acento)
                } else {
                    forma.fill(LinearGradient(colors: [.white, Color(hex: 0xF6EFE7)], startPoint: .top, endPoint: .bottom))
                }
            }
            .overlay { Grano(opacidad: 0.08, mezcla: .multiply) }
            .overlay(forma.strokeBorder(LinearGradient(colors: [.white.opacity(paleta.oscuro ? 0.2 : 1), .clear],
                                                       startPoint: .top, endPoint: UnitPoint(x: 0.5, y: 0.2)), lineWidth: 1))
            .overlay(forma.strokeBorder(LinearGradient(colors: [.clear, Color(hex: 0x2A140C, alfa: 0.08)],
                                                       startPoint: UnitPoint(x: 0.5, y: 0.8), endPoint: .bottom), lineWidth: 2))
            .clipShape(forma)
            .shadow(color: .black.opacity(paleta.oscuro ? 0.4 : 0.32), radius: 14, y: 12)
        }
        .buttonStyle(BotonHundido())
    }
}

/// Se hunde apenas al apretar (sin cambiar el color): los botones con relieve.
struct BotonHundido: ButtonStyle {
    func makeBody(configuration: Configuration) -> some View {
        configuration.label.hundir(configuration.isPressed)
    }
}

/// Línea punteada de la leyenda del gráfico.
struct Punteado: View {
    var color: Color = Color(hex: 0x8A7364)
    var body: some View {
        Canvas { ctx, tam in
            var p = Path()
            p.move(to: CGPoint(x: 0, y: tam.height / 2))
            p.addLine(to: CGPoint(x: tam.width, y: tam.height / 2))
            ctx.stroke(p, with: .color(color), style: StrokeStyle(lineWidth: 2, dash: [3, 2]))
        }
    }
}

// MARK: - Plegada (AScroll)

/// La vidriera plegada al bajar: el logo, lo vendido y un «Vender» compacto.
struct BarraVidriera: View {
    var tablero: Tablero
    var periodo: String
    var paleta: PaletaTienda
    @Environment(Estado.self) private var estado
    @Environment(Navegacion.self) private var nav

    var body: some View {
        HStack(spacing: 10) {
            LogoTienda(tienda: estado.tienda, tam: 38, radio: 13, paleta: paleta)
            VStack(alignment: .leading, spacing: 0) {
                Text("\(estado.tienda?.nombre ?? "Tu tienda") · \(periodo == "hoy" ? "hoy" : (periodo == "7d" ? "7 días" : "30 días"))")
                    .estilo(Estilo(tamano: 13, peso: .medium, relativo: .footnote)).foregroundStyle(paleta.sobre2)
                    .lineLimit(1)
                Text(Vidriera.monto(tablero))
                    .estilo(Estilo(tamano: 21, peso: .bold, tracking: -0.024, relativo: .title3))
                    .monospacedDigit().lineLimit(1).minimumScaleFactor(0.7)
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            .accessibilityElement(children: .combine)
            Button { nav.venta = true } label: {
                HStack(spacing: 8) {
                    Image(systemName: "barcode.viewfinder").font(.system(size: 16, weight: .semibold))
                    Text("Vender").estilo(Estilo(tamano: 16, peso: .bold, relativo: .body))
                }
                .foregroundStyle(paleta.oscuro ? paleta.sobreAcento : paleta.tintaBlanco)
                .padding(.horizontal, 16).frame(height: 44)
                .background(paleta.oscuro ? paleta.acento : .white, in: Capsule())
                .shadow(color: .black.opacity(0.2), radius: 7, y: 6)
            }
            .buttonStyle(BotonHundido())
        }
        .foregroundStyle(paleta.sobre)
        .padding(.horizontal, 16).padding(.top, 8).padding(.bottom, 14)
        .background {
            FondoVidriera(paleta: paleta, radio: 26, halos: false)
                .shadow(color: Color(hex: 0x2A140C, alfa: 0.22), radius: 12, y: 10)
        }
    }
}

// MARK: - Tincho

/// Una idea de Tincho para la tarjeta del Inicio: el texto, la acción principal (si hay) y la pregunta
/// con la que «Preguntale» abre el chat.
struct IdeaTincho: Identifiable {
    var id: String
    var texto: AttributedString
    var boton: String?
    var pregunta: String
    var accion: () async -> Void
}

/// La tarjeta de Tincho montada sobre el borde de la vidriera (reemplaza a «Preguntale a Tincho» y a
/// las ideas de abajo).
struct TarjetaTincho: View {
    var idea: IdeaTincho
    var paleta: PaletaTienda
    @Environment(Navegacion.self) private var nav

    var body: some View {
        HStack(alignment: .top, spacing: 12) {
            Image(Tincho.avatar.imagen).resizable().scaledToFill()
                .frame(width: 48, height: 48)
                .background(Color.tema(0xF3E3D3, 0x4A2A1A))
                .clipShape(Circle())
                .accessibilityHidden(true)
            VStack(alignment: .leading, spacing: 6) {
                Text(idea.texto)
                    .estilo(Estilo(tamano: 16, peso: .regular, relativo: .body)).lineSpacing(4)
                    .foregroundStyle(C.texto)
                    .fixedSize(horizontal: false, vertical: true)
                    .accessibilityLabel(Text("Tincho: ") + Text(idea.texto))
                HStack(spacing: 8) {
                    if let b = idea.boton {
                        BotonChico(texto: b, fondo: paleta.acento, tinta: paleta.sobreAcento, relieve: true) {
                            Task { await idea.accion() }
                        }
                    }
                    BotonChico(texto: "Preguntale", fondo: idea.boton == nil ? paleta.acento : paleta.claro,
                               tinta: idea.boton == nil ? paleta.sobreAcento : paleta.tinta, relieve: idea.boton == nil) {
                        nav.preguntaTincho = idea.pregunta
                        nav.pestana = .tincho
                    }
                    .accessibilityHint("Abre Tincho con la pregunta: \(idea.pregunta)")
                }
            }
        }
        .padding(.top, 14).padding(.bottom, 10).padding(.leading, 12).padding(.trailing, 14)
        .frame(maxWidth: .infinity, alignment: .leading)
        .relieve(radio: 24, fondo: Color.tema(0xFFFDF9, 0x2E1B12), fuerte: true)
    }
}

/// Botón chico en píldora (36 visibles, 44 para tocar): las acciones de Tincho.
struct BotonChico: View {
    var texto: String
    var fondo: Color
    var tinta: Color
    var relieve = false
    var accion: () -> Void
    var body: some View {
        Button(action: accion) {
            Text(texto).estilo(Estilo(tamano: 14, peso: .bold, relativo: .footnote))
                .foregroundStyle(tinta).lineLimit(1)
                .padding(.horizontal, 14).frame(height: 36)
                .background(fondo, in: Capsule())
                .overlay {
                    if relieve {
                        Capsule().strokeBorder(LinearGradient(colors: [.white.opacity(0.25), .clear], startPoint: .top, endPoint: .center), lineWidth: 1)
                    }
                }
                .shadow(color: relieve ? fondo.opacity(0.5) : .clear, radius: 5, y: 4)
                .frame(minHeight: 44)
                .contentShape(Rectangle())
        }
        .buttonStyle(BotonHundido())
    }
}
