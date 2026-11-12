import CoreImage
import SwiftUI
import UIKit

// Piezas con cuerpo para las pantallas con más vida (el Inicio «La vidriera de tu local», handoff
// docs/handoff/inicio-a): el grano, la tarjeta con luz arriba y sombra en capas, el vidrio sobre un
// color, la paleta que sale del color de la tienda y su QR. La misma receta está en tema/Relieve.kt.

// MARK: - Grano

/// La textura de grano (Assets › Grano: ruido de 128 px en mosaico) con su mezcla y opacidad. Sobre un
/// color va en `overlay` (0,2 en la vidriera); sobre el papel o un botón blanco, en `multiply` (0,06).
struct Grano: View {
    var opacidad: Double = 0.2
    var mezcla: BlendMode = .overlay
    var body: some View {
        Image("Grano").resizable(resizingMode: .tile)
            .blendMode(mezcla).opacity(opacidad)
            .allowsHitTesting(false).accessibilityHidden(true)
    }
}

// MARK: - Relieve

/// La única receta de sombra en capas (la del handoff: `inset 0 1px 0 #FFF, 0 1px 0 #E3D6C8,
/// 0 12px 24px -12px rgba(42,20,12,.22)`): una luz de 1 arriba, un filo de 1 abajo y una sombra larga y
/// corrida hacia abajo. En oscuro, la luz es apenas un brillo y la sombra, negra.
enum Relieve {
    static let luz = Color.tema(0xFFFFFF, 0xFFFFFF, alfa: 1, alfaOscuro: 0.1)
    static let anillo = Color.tema(0xFFFFFF, 0xFFFFFF, alfa: 0, alfaOscuro: 0.06)
    static let filo = Color.tema(0xE3D6C8, 0x000000, alfaOscuro: 0.35)
    static let sombra = Color.tema(0x2A140C, 0x000000, alfa: 0.14, alfaOscuro: 0.45)
    /// Radio y caída de la sombra: la de siempre y la de la tarjeta de Tincho (`0 16px 32px -12px .28`).
    static func caida(fuerte: Bool) -> (radio: CGFloat, y: CGFloat) { fuerte ? (14, 14) : (10, 10) }
}

extension View {
    /// Tarjeta con relieve: fondo, luz arriba, filo abajo y sombra en capas (radio 24 en tarjetas).
    /// `luz` en false para las tarjetas oscuras (la meta), que no llevan el brillo de arriba.
    func relieve(radio: CGFloat = 24, fondo: Color = C.superficie, fuerte: Bool = false, luz: Bool = true) -> some View {
        let forma = RoundedRectangle(cornerRadius: radio, style: .continuous)
        let c = Relieve.caida(fuerte: fuerte)
        return self
            .background(fondo, in: forma)
            .overlay(forma.strokeBorder(Relieve.anillo, lineWidth: 1))
            .overlay(forma.strokeBorder(LinearGradient(colors: [Relieve.luz.opacity(luz ? 1 : 0), Relieve.luz.opacity(0)],
                                                       startPoint: .top, endPoint: UnitPoint(x: 0.5, y: 0.12)), lineWidth: 1))
            .clipShape(forma)
            .background {
                forma.fill(Relieve.filo).offset(y: 1)
                    .shadow(color: Relieve.sombra, radius: c.radio, y: c.y)
            }
    }

    /// Vidrio sobre el color de la tienda: el velo y un borde de 1 (`veloV` y `bordeV` del handoff).
    func vidrio<S: InsettableShape>(_ forma: S, _ paleta: PaletaTienda) -> some View {
        background(paleta.velo, in: forma).overlay(forma.strokeBorder(paleta.borde, lineWidth: 1))
    }
}

// MARK: - Paleta de la tienda

/// Los colores que salen del de la tienda (`acento` y `sobreAcento` de /v1/panel/tablero), como los
/// deriva el handoff: el velo, el borde, la tinta, el claro… En oscuro, la vidriera se apaga hacia el
/// fondo y el texto encima es el del tema. Si el acento es claro (Pampa, #E6FF4D), el texto va oscuro.
struct PaletaTienda: Equatable {
    /// El acento y el texto que va encima de él (botones y la insignia en acento).
    var acento: Color
    var sobreAcento: Color
    /// El fondo de la vidriera (en claro, el acento; en oscuro, un degradé hacia el fondo).
    var fondoArriba: Color
    var fondoAbajo: Color
    /// El texto sobre la vidriera, el secundario y lo punteado.
    var sobre: Color
    var sobre2: Color
    var sobreSuave: Color
    var velo: Color
    var borde: Color
    /// Los halos (en oscuro, uno solo con el acento).
    var luz: Color
    var sombraHalo: Color
    /// El acento para texto sobre blanco o superficie (links, «Preparar», el número de la fila).
    var tinta: Color
    /// La tinta sobre blanco siempre, en los dos modos (el QR y la miniatura, que son blancos).
    var tintaBlanco: Color
    /// Un fondo suave del acento (píldoras, insignias, «Preguntale»).
    var claro: Color
    var medio: Color
    var meta: Color
    /// La línea del gráfico de la vidriera.
    var linea: Color
    var oscuro: Bool
    /// Si el texto sobre el acento es claro (acento oscuro).
    var claroEncima: Bool

    static let porDefecto: UInt32 = 0xD9652F

    init(acento: String?, sobre: String?, oscuro: Bool) {
        let a = PaletaTienda.leer(acento) ?? PaletaTienda.porDefecto
        let s = PaletaTienda.leer(sobre) ?? (PaletaTienda.luminancia(a) < 0.3 ? 0xFFFFFF : 0x1E0E12)
        let claroEncima = PaletaTienda.luminancia(s) > 0.5
        let mezclar = PaletaTienda.mezclar
        self.oscuro = oscuro
        self.claroEncima = claroEncima
        self.acento = Color(hex: a)
        sobreAcento = Color(hex: s)
        let blanco = Color(hex: claroEncima ? mezclar(a, 0x000000, 0.2) : mezclar(a, 0x000000, 0.6))
        tintaBlanco = blanco
        medio = Color(hex: mezclar(a, 0xFFFFFF, 0.45))
        if oscuro {
            fondoArriba = Color(hex: mezclar(a, 0x22140D, 0.55))
            fondoAbajo = Color(hex: mezclar(a, 0x22140D, 0.82))
            sobre = Color(hex: 0xFBF7F2)
            sobre2 = Color(hex: 0xCDB8A8)
            sobreSuave = Color(hex: 0xFBF7F2, alfa: 0.35)
            velo = Color(hex: 0xFFFFFF, alfa: 0.06)
            borde = Color(hex: 0xFFFFFF, alfa: 0.14)
            luz = Color(hex: a, alfa: 0.28)
            sombraHalo = .clear
            let luzTienda = Color(hex: mezclar(a, 0xFFFFFF, 0.5))
            tinta = luzTienda
            claro = Color(hex: mezclar(a, 0x2E1B12, 0.6))
            meta = luzTienda
            linea = luzTienda
        } else {
            fondoArriba = Color(hex: a)
            fondoAbajo = Color(hex: a)
            sobre = Color(hex: s)
            sobre2 = Color(hex: s, alfa: 0.88)
            sobreSuave = Color(hex: s, alfa: claroEncima ? 0.55 : 0.45)
            velo = Color(hex: s, alfa: claroEncima ? 0.14 : 0.08)
            borde = Color(hex: s, alfa: claroEncima ? 0.28 : 0.16)
            luz = Color(hex: mezclar(a, 0xFFFFFF, 0.45), alfa: 0.55)
            sombraHalo = Color(hex: mezclar(a, 0x000000, 0.55), alfa: 0.6)
            tinta = blanco
            claro = Color(hex: mezclar(a, 0xFFFFFF, 0.86))
            meta = claroEncima ? Color(hex: mezclar(a, 0xFFFFFF, 0.45)) : Color(hex: a)
            linea = Color(hex: s)
        }
    }

    /// «#RRGGBB» → 0xRRGGBB (nil si no es un color).
    static func leer(_ s: String?) -> UInt32? {
        guard let s, s.count == 7, s.hasPrefix("#") else { return nil }
        return UInt32(s.dropFirst(), radix: 16)
    }

    /// Mezcla `t` (0–1) de `con` sobre `hex`, canal por canal (el `mezclar` del handoff).
    static func mezclar(_ hex: UInt32, _ con: UInt32, _ t: Double) -> UInt32 {
        [16, 8, 0].reduce(UInt32(0)) { r, s in
            let a = Double((hex >> s) & 0xFF), b = Double((con >> s) & 0xFF)
            return r | (UInt32((a * (1 - t) + b * t).rounded()) << s)
        }
    }

    /// Luminancia relativa (WCAG 2).
    static func luminancia(_ hex: UInt32) -> Double {
        let c = [16, 8, 0].map { s -> Double in
            let v = Double((hex >> UInt32(s)) & 0xFF) / 255
            return v <= 0.03928 ? v / 12.92 : pow((v + 0.055) / 1.055, 2.4)
        }
        return 0.2126 * c[0] + 0.7152 * c[1] + 0.0722 * c[2]
    }
}

// MARK: - QR

/// El QR de un texto (el link de la tienda) en `color` sobre transparente, generado en el teléfono con
/// CIQRCodeGenerator y nítido a cualquier tamaño.
struct CodigoQR: View {
    var texto: String
    var color: Color
    var body: some View {
        if let img = CodigoQR.mascara(texto) {
            Image(uiImage: img).renderingMode(.template).interpolation(.none).resizable().scaledToFit()
                .foregroundStyle(color)
        }
    }

    private static let guardados = NSCache<NSString, UIImage>()

    /// Los módulos opacos y el resto transparente (invertido y pasado a alfa), un píxel por módulo.
    static func mascara(_ texto: String) -> UIImage? {
        if let img = guardados.object(forKey: texto as NSString) { return img }
        guard let f = CIFilter(name: "CIQRCodeGenerator") else { return nil }
        f.setValue(Data(texto.utf8), forKey: "inputMessage")
        f.setValue("M", forKey: "inputCorrectionLevel")
        guard let qr = f.outputImage,
              let invertido = CIFilter(name: "CIColorInvert", parameters: [kCIInputImageKey: qr])?.outputImage,
              let alfa = CIFilter(name: "CIMaskToAlpha", parameters: [kCIInputImageKey: invertido])?.outputImage,
              let cg = CIContext().createCGImage(alfa, from: alfa.extent) else { return nil }
        let img = UIImage(cgImage: cg)
        guardados.setObject(img, forKey: texto as NSString)
        return img
    }
}

// MARK: - Foto

/// Una foto que llena lo que se le dé (la de un producto), con el rayado mientras carga o si no tiene.
struct Foto: View {
    var foto: String?
    var ancho = 900
    var body: some View {
        Rayado()
            .overlay {
                if let u = urlDeFoto(foto, ancho: ancho) {
                    AsyncImage(url: u, transaction: Transaction(animation: .easeOut(duration: 0.2))) { fase in
                        if let img = fase.image { img.resizable().scaledToFill() }
                    }
                }
            }
            .clipped()
            .accessibilityHidden(true)
    }
}
