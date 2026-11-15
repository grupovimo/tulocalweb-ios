import SwiftUI

/// La miniatura de un bloque o de una variante, dibujada como en docs/handoff/tienda-bloques/Bloques.dc.html:
/// rayas para los textos, bloques grises para las fotos y el acento ámbar para los botones de promo.
struct MiniaturaBloque: View {
    var tipo: String
    var variante: String
    var ancho: CGFloat = 120

    private static let foto = Color(hex: 0xD7D3CA)
    private static let linea = Color(hex: 0x3A3936)
    private static let suave = Color(hex: 0xB9B5AC)
    private static let gris = Color(hex: 0x8B887F)
    private static let ambar = Color(hex: 0xE7B04A)
    private static let borde = Color(hex: 0xE2DFD8)

    /// Qué dibujo le toca a cada variante (los nombres de Bloques.dc.html).
    enum Dibujo: Equatable {
        case foto, partido, partidoIzquierda, collage, texto, carrusel, banda, contador, dos, iconos, grilla(Int, Int)
        case look, lista, mapa, acordeon, tarjetas, logos, input, circulos, video, miniaturas, panel
    }

    static func forma(_ tipo: String, _ variante: String) -> Dibujo {
        switch (tipo, variante) {
        case ("textoImagen", "izquierda"): return .partidoIzquierda
        case (_, "partida"), (_, "partido"), (_, "derecha"), ("video", "conTexto"): return .partido
        case ("video", _): return .video
        case ("combo", _): return .look
        case (_, "foto"), (_, "fondo"): return .foto
        case (_, "collage"): return .collage
        case ("newsletter", "centrado"): return .input
        case (_, "texto"), (_, "cita"), (_, "centrado"): return .texto
        case ("catalogo", "grilla"): return .grilla(4, 8)
        case ("producto", "grilla"): return .grilla(2, 4)
        case ("producto", "miniaturas"): return .miniaturas
        case ("carrito", "panel"): return .panel
        case ("carrito", "pagina"): return .lista
        case (_, "carrusel"): return .carrusel
        case ("oferta", "productos"): return .contador
        case (_, "banda"), ("barra", _): return .banda
        case (_, "doble"): return .dos
        case ("beneficios", "fila"): return .iconos
        case ("productos", "grilla"): return .grilla(4, 4)
        case ("categorias", "tarjetas"): return .grilla(3, 3)
        case ("menu", "columnas"): return .grilla(2, 4)
        case ("galeria", "fila"): return .grilla(6, 6)
        case ("galeria", "mosaico"): return .grilla(3, 6)
        case (_, "tarjetas"), (_, "tarjeta"): return .tarjetas
        case (_, "circulos"): return .circulos
        case (_, "mapa"): return .mapa
        case (_, "acordeon"): return .acordeon
        case ("logos", _): return .logos
        default: return .lista
        }
    }

    private var alto: CGFloat { (ancho * 0.66).rounded() }
    private var adentro: CGFloat { ancho - 12 }

    var body: some View {
        dibujo
            .padding(6)
            .frame(width: ancho, height: alto, alignment: .topLeading)
            .background(.white, in: RoundedRectangle(cornerRadius: 8, style: .continuous))
            .overlay(RoundedRectangle(cornerRadius: 8, style: .continuous).strokeBorder(Self.borde))
            .clipped()
            .accessibilityHidden(true)
    }

    // MARK: Piezas

    private func linea(_ f: CGFloat, alto a: CGFloat = 5, color: Color = Self.linea) -> some View {
        Capsule().fill(color).frame(width: adentro * f, height: a)
    }

    private func suave(_ f: CGFloat) -> some View { linea(f, alto: 4, color: Self.suave) }

    private func boton(_ color: Color = Self.linea) -> some View {
        RoundedRectangle(cornerRadius: 2).fill(color).frame(width: adentro * 0.3, height: 8)
    }

    private var foto: some View { RoundedRectangle(cornerRadius: 3).fill(Self.foto) }

    private func textos(_ f: CGFloat = 0.45, boton b: Bool = true) -> some View {
        VStack(alignment: .leading, spacing: 4) {
            linea(f, alto: 7)
            linea(f * 0.7)
            suave(f * 0.85)
            if b { boton().padding(.top, 2) }
        }
    }

    // MARK: Dibujos

    @ViewBuilder private var dibujo: some View {
        let h = alto - 12
        switch Self.forma(tipo, variante) {
        case .foto:
            ZStack(alignment: .bottomLeading) {
                foto
                VStack(alignment: .leading, spacing: 4) {
                    linea(0.45, alto: 6, color: .white)
                    linea(0.3, color: .white)
                    boton(.white)
                }
                .padding(8)
            }
        case .partido, .partidoIzquierda:
            HStack(spacing: 5) {
                if Self.forma(tipo, variante) == .partidoIzquierda { foto }
                textos().frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .leading)
                if Self.forma(tipo, variante) == .partido {
                    if tipo == "video" { videoCaja } else { foto }
                }
            }
        case .collage:
            HStack(spacing: 4) {
                RoundedRectangle(cornerRadius: 3).fill(Self.linea)
                    .overlay(alignment: .bottomLeading) {
                        VStack(alignment: .leading, spacing: 3) {
                            Capsule().fill(.white).frame(width: adentro * 0.25, height: 4)
                            RoundedRectangle(cornerRadius: 2).fill(Self.ambar).frame(width: adentro * 0.18, height: 6)
                        }
                        .padding(5)
                    }
                foto
                VStack(spacing: 4) { foto; RoundedRectangle(cornerRadius: 3).fill(Self.ambar) }
            }
        case .texto:
            VStack(spacing: 5) {
                linea(0.75, alto: 7)
                linea(0.5, alto: 7)
                suave(0.55)
                boton()
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
        case .carrusel:
            ZStack(alignment: .bottom) {
                foto
                HStack(spacing: 3) {
                    Capsule().fill(Self.linea).frame(width: 12, height: 3)
                    Circle().fill(Self.gris).frame(width: 3, height: 3)
                    Circle().fill(Self.gris).frame(width: 3, height: 3)
                }
                .padding(.bottom, 6)
            }
        case .banda:
            HStack {
                VStack(alignment: .leading, spacing: 4) {
                    linea(0.4, alto: 6, color: .white)
                    linea(0.3, alto: 4, color: Self.gris)
                }
                Spacer(minLength: 0)
                boton(Self.ambar)
            }
            .padding(10)
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .background(Self.linea, in: RoundedRectangle(cornerRadius: 5))
        case .contador:
            VStack(spacing: 5) {
                HStack(spacing: 3) {
                    linea(0.35)
                    Spacer(minLength: 0)
                    ForEach(0..<3, id: \.self) { _ in RoundedRectangle(cornerRadius: 2).fill(Self.linea).frame(width: 12, height: 10) }
                }
                celdas(4, 4, alto: h - 15)
            }
        case .dos:
            HStack(spacing: 4) { foto; RoundedRectangle(cornerRadius: 3).fill(Self.linea) }
        case .iconos:
            HStack(spacing: 4) {
                ForEach(0..<4, id: \.self) { _ in
                    VStack(spacing: 4) {
                        RoundedRectangle(cornerRadius: 4).strokeBorder(Self.linea, lineWidth: 1.5).frame(width: 14, height: 14)
                        Capsule().fill(Self.suave).frame(height: 4)
                    }
                    .frame(maxWidth: .infinity)
                }
            }
            .frame(maxHeight: .infinity)
        case let .grilla(cols, total):
            celdas(cols, total, alto: h)
        case .look:
            HStack(spacing: 4) {
                foto; foto; foto
                VStack(alignment: .leading, spacing: 3) {
                    Spacer(minLength: 0)
                    Capsule().fill(Self.linea).frame(height: 4)
                    RoundedRectangle(cornerRadius: 2).fill(Self.linea).frame(height: 7)
                }
                .padding(3)
                .overlay(RoundedRectangle(cornerRadius: 3).strokeBorder(Self.linea, lineWidth: 1.5))
            }
        case .lista:
            VStack(spacing: 5) {
                ForEach(0..<3, id: \.self) { i in
                    HStack(spacing: 5) {
                        VStack(alignment: .leading, spacing: 3) { linea(i == 1 ? 0.35 : 0.45, alto: 4); suave(0.6) }
                        Spacer(minLength: 0)
                        foto.frame(width: h / 3.6, height: h / 3.6)
                    }
                    .frame(maxHeight: .infinity)
                }
            }
        case .mapa:
            HStack(spacing: 5) {
                VStack(alignment: .leading, spacing: 4) { linea(0.38, alto: 6); suave(0.4); suave(0.3); suave(0.35) }
                    .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .leading)
                ZStack {
                    RoundedRectangle(cornerRadius: 3).fill(Color(hex: 0xE6E1D6))
                    Path { p in
                        for i in 1..<4 { let x = CGFloat(i) * adentro * 0.13; p.addRect(CGRect(x: x, y: 0, width: 2, height: h)) }
                        for i in 1..<3 { let y = CGFloat(i) * h / 3; p.addRect(CGRect(x: 0, y: y, width: adentro, height: 2)) }
                    }
                    .fill(.white)
                    Circle().fill(Self.ambar).frame(width: 9, height: 9)
                }
                .clipShape(RoundedRectangle(cornerRadius: 3))
            }
        case .acordeon:
            VStack(alignment: .leading, spacing: 4) {
                ForEach(0..<3, id: \.self) { i in
                    HStack { linea(i == 1 ? 0.45 : 0.55, alto: 4); Spacer(minLength: 0); Text(i == 0 ? "−" : "+").font(.system(size: 8)) }
                    if i == 0 { suave(0.8) }
                    Rectangle().fill(Self.borde).frame(height: 1)
                }
            }
            .foregroundStyle(Self.linea)
        case .tarjetas:
            HStack(spacing: 4) {
                ForEach(0..<3, id: \.self) { _ in
                    VStack(alignment: .leading, spacing: 3) {
                        Capsule().fill(Self.ambar).frame(width: adentro * 0.12, height: 4)
                        Capsule().fill(Self.linea).frame(height: 4)
                        Capsule().fill(Self.suave).frame(height: 3)
                        Capsule().fill(Self.suave).frame(width: adentro * 0.15, height: 3)
                        Spacer(minLength: 0)
                    }
                    .padding(4)
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
                    .overlay(RoundedRectangle(cornerRadius: 3).strokeBorder(Self.borde))
                }
            }
        case .logos:
            VStack(spacing: 9) {
                suave(0.4)
                HStack(spacing: 6) {
                    ForEach([0.22, 0.17, 0.25, 0.19] as [CGFloat], id: \.self) { f in
                        RoundedRectangle(cornerRadius: 2).fill(Self.gris).frame(width: adentro * f * 0.9, height: 9)
                    }
                }
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
        case .input:
            VStack(spacing: 5) {
                linea(0.55, alto: 6)
                suave(0.4)
                HStack(spacing: 3) {
                    RoundedRectangle(cornerRadius: 2).strokeBorder(Self.gris).frame(width: adentro * 0.42, height: 11)
                    RoundedRectangle(cornerRadius: 2).fill(Self.linea).frame(width: adentro * 0.2, height: 11)
                }
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
        case .circulos:
            HStack {
                ForEach(0..<4, id: \.self) { _ in
                    VStack(spacing: 4) {
                        Circle().fill(Self.foto).frame(width: adentro * 0.2, height: adentro * 0.2)
                        Capsule().fill(Self.suave).frame(width: adentro * 0.15, height: 4)
                    }
                    .frame(maxWidth: .infinity)
                }
            }
            .frame(maxHeight: .infinity)
        case .video:
            videoCaja
        case .miniaturas:
            HStack(spacing: 4) {
                VStack(spacing: 3) { foto; foto; foto }.frame(width: adentro * 0.12)
                foto
                textos(0.3).frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .leading)
            }
        case .panel:
            HStack(spacing: 4) {
                foto.opacity(0.5)
                VStack(alignment: .leading, spacing: 4) {
                    linea(0.25, alto: 5)
                    ForEach(0..<2, id: \.self) { _ in HStack(spacing: 3) { foto.frame(width: 12, height: 12); suave(0.15) } }
                    Spacer(minLength: 0)
                    RoundedRectangle(cornerRadius: 2).fill(Self.linea).frame(height: 8)
                }
                .padding(4)
                .frame(width: adentro * 0.38)
                .overlay(RoundedRectangle(cornerRadius: 3).strokeBorder(Self.borde))
            }
        }
    }

    private var videoCaja: some View {
        RoundedRectangle(cornerRadius: 3).fill(Self.gris)
            .overlay {
                Circle().fill(.white).frame(width: 20, height: 20)
                    .overlay(Image(systemName: "play.fill").font(.system(size: 8)).foregroundStyle(Self.linea))
            }
    }

    /// Celdas de producto (foto, nombre y precio) o de foto sola, en filas completas.
    private func celdas(_ cols: Int, _ total: Int, alto h: CGFloat) -> some View {
        let filas = max(1, total / cols)
        let conTexto = cols <= 4 && tipo != "galeria"
        return VStack(spacing: 4) {
            ForEach(0..<filas, id: \.self) { _ in
                HStack(spacing: 4) {
                    ForEach(0..<cols, id: \.self) { _ in
                        VStack(alignment: .leading, spacing: 2) {
                            foto
                            if conTexto {
                                Capsule().fill(Self.suave).frame(height: 3)
                                Capsule().fill(Self.linea).frame(width: adentro / CGFloat(cols) * 0.5, height: 3)
                            }
                        }
                    }
                }
            }
        }
        .frame(height: h)
    }
}
