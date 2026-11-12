import SwiftUI
import UIKit

// Los tokens del handoff (docs/handoff/app-ios/README.md · Design Tokens), claro y oscuro.

extension UIColor {
    convenience init(hex: UInt32, alfa: CGFloat = 1) {
        self.init(red: CGFloat((hex >> 16) & 0xFF) / 255, green: CGFloat((hex >> 8) & 0xFF) / 255,
                  blue: CGFloat(hex & 0xFF) / 255, alpha: alfa)
    }
}

extension Color {
    init(hex: UInt32, alfa: Double = 1) {
        self.init(uiColor: UIColor(hex: hex, alfa: alfa))
    }

    /// Un color que cambia con el modo oscuro.
    static func tema(_ claro: UInt32, _ oscuro: UInt32, alfa: CGFloat = 1, alfaOscuro: CGFloat? = nil) -> Color {
        Color(uiColor: .tema(claro, oscuro, alfa: alfa, alfaOscuro: alfaOscuro))
    }
}

extension UIColor {
    static func tema(_ claro: UInt32, _ oscuro: UInt32, alfa: CGFloat = 1, alfaOscuro: CGFloat? = nil) -> UIColor {
        UIColor { t in
            t.userInterfaceStyle == .dark ? UIColor(hex: oscuro, alfa: alfaOscuro ?? alfa) : UIColor(hex: claro, alfa: alfa)
        }
    }
}

/// Colores del handoff. Los nombres siguen la tabla del README.
enum C {
    static let fondo = Color.tema(0xF7F1EA, 0x22140D)
    static let elevado = Color.tema(0xFBF7F2, 0x2E1B12)
    static let superficie = Color.tema(0xFFFFFF, 0x2E1B12)
    static let texto = Color.tema(0x2A140C, 0xFBF7F2)
    static let texto2 = Color.tema(0x6B5244, 0xCDB8A8)
    static let texto3 = Color.tema(0x7D6656, 0xA8907F)
    static let chevron = Color.tema(0xB5A193, 0x7D6656)
    static let acento = Color(hex: 0xD9652F)
    static let primarioTexto = Color.tema(0xB0441D, 0xE07A48)
    static let primarioTinta = Color.tema(0x8E3A16, 0xF3A57E)
    static let crema = Color.tema(0xF8E6D4, 0x4A2A1A)
    static let crema2 = Color.tema(0xFBF3EA, 0x3A2216)
    static let pista = Color.tema(0xF3E9DF, 0x3A2419)
    static let pistaMeta = Color.tema(0xF3E6DA, 0x3A2419)
    static let ok = Color.tema(0x2F8F63, 0x3FA877)
    static let okFondo = Color.tema(0xE1F1E8, 0x3FA877, alfaOscuro: 0.18)
    static let okTinta = Color.tema(0x1F6B48, 0x6FD0A0)
    static let aviso = Color.tema(0xC2861B, 0xE3A93B)
    static let avisoFondo = Color.tema(0xFBEFD8, 0xE3A93B, alfaOscuro: 0.18)
    static let avisoTinta = Color.tema(0x7A4F08, 0xF0C267)
    static let bannerTinta = Color.tema(0x6E4707, 0xF0C267)
    static let error = Color(hex: 0xD2582F)
    static let errorFondo = Color.tema(0xFBE4DA, 0xD2582F, alfaOscuro: 0.2)
    static let errorTinta = Color.tema(0xA23F1C, 0xF08A63)
    static let neutroFondo = Color.tema(0xEFE6DF, 0xFBF7F2, alfaOscuro: 0.1)
    static let neutroTinta = Color.tema(0x5A4033, 0xCDB8A8)
    static let apagadoFondo = Color.tema(0xEEE8E3, 0xFBF7F2, alfaOscuro: 0.08)
    static let linea = Color.tema(0x3A1D12, 0xFBF7F2, alfa: 0.09, alfaOscuro: 0.08)
    static let linea2 = Color.tema(0x3A1D12, 0xFBF7F2, alfa: 0.16, alfaOscuro: 0.12)
    static let anillo = Color.tema(0x3A1D12, 0xFBF7F2, alfa: 0.06, alfaOscuro: 0.08)
    static let segmentado = Color.tema(0x3A1D12, 0xFBF7F2, alfa: 0.07, alfaOscuro: 0.07)
    static let segmentoElegido = Color.tema(0xFFFFFF, 0x4A3125)
    static let velo = Color(hex: 0x2A140C, alfa: 0.38)
    static let rayado1 = Color.tema(0xF3E3D2, 0x3A2419)
    static let rayado2 = Color.tema(0xFAEEE2, 0x33201A)
    static let esqueleto1 = Color.tema(0xEFE6DD, 0x3A2419)
    static let esqueleto2 = Color.tema(0xF3ECE5, 0x33201A)
    static let canalWeb = Color(hex: 0xD9652F)
    static let canalML = Color(hex: 0xE3A93B)
    static let canalLocal = Color.tema(0x6B5244, 0xCDB8A8)
    /// El gradiente del botón principal, la burbuja de Tincho y el «+».
    static let primario = LinearGradient(colors: [Color.tema(0xC24E25, 0xD9652F), Color(hex: 0xB0441D)],
                                         startPoint: .top, endPoint: .bottom)
    static let sombraPrimario = Color(hex: 0xB0441D, alfa: 0.7)
}

/// Pesos de Instrument Sans incluidos en el bundle.
enum Peso {
    case regular, medium, semibold, bold

    var nombre: String {
        switch self {
        case .regular: return "InstrumentSans-Regular"
        case .medium: return "InstrumentSans-Medium"
        case .semibold: return "InstrumentSans-SemiBold"
        case .bold: return "InstrumentSans-Bold"
        }
    }

    var ui: UIFont.Weight {
        switch self {
        case .regular: return .regular
        case .medium: return .medium
        case .semibold: return .semibold
        case .bold: return .bold
        }
    }
}

/// Un estilo de texto del handoff: tamaño, peso, tracking (en em) y el estilo de Dynamic Type al que
/// escala.
struct Estilo {
    var tamano: CGFloat
    var peso: Peso
    var tracking: CGFloat = 0
    var relativo: Font.TextStyle

    var font: Font { .custom(peso.nombre, size: tamano, relativeTo: relativo) }

    static let montoHeroe = Estilo(tamano: 46, peso: .bold, tracking: -0.035, relativo: .largeTitle)
    static let montoCobro = Estilo(tamano: 52, peso: .bold, tracking: -0.035, relativo: .largeTitle)
    static let montoPantalla = Estilo(tamano: 40, peso: .bold, tracking: -0.03, relativo: .largeTitle)
    static let tituloGrande = Estilo(tamano: 34, peso: .bold, tracking: -0.025, relativo: .largeTitle)
    static let saludo = Estilo(tamano: 32, peso: .bold, tracking: -0.025, relativo: .title)
    static let tituloPaso = Estilo(tamano: 30, peso: .bold, tracking: -0.025, relativo: .title)
    static let tituloBloqueo = Estilo(tamano: 28, peso: .bold, tracking: -0.02, relativo: .title)
    static let montoMeta = Estilo(tamano: 28, peso: .bold, tracking: -0.025, relativo: .title)
    static let tituloHoja = Estilo(tamano: 24, peso: .bold, tracking: -0.02, relativo: .title2)
    static let tituloHojaChico = Estilo(tamano: 22, peso: .bold, tracking: -0.02, relativo: .title2)
    static let tituloSeccion = Estilo(tamano: 21, peso: .bold, tracking: -0.02, relativo: .title3)
    /// Las secciones del Inicio «La vidriera» (22/700).
    static let seccion = Estilo(tamano: 22, peso: .bold, tracking: -0.018, relativo: .title2)
    static let indicador = Estilo(tamano: 22, peso: .bold, tracking: -0.02, relativo: .title3)
    static let nombreTienda = Estilo(tamano: 18, peso: .bold, relativo: .headline)
    static let total = Estilo(tamano: 18, peso: .bold, relativo: .headline)
    static let headline = Estilo(tamano: 17, peso: .semibold, relativo: .headline)
    static let body = Estilo(tamano: 17, peso: .regular, relativo: .body)
    static let fila = Estilo(tamano: 16, peso: .semibold, relativo: .body)
    static let body16 = Estilo(tamano: 16, peso: .regular, relativo: .body)
    static let tarjetaTincho = Estilo(tamano: 16, peso: .medium, relativo: .body)
    static let subhead = Estilo(tamano: 15, peso: .regular, relativo: .subheadline)
    static let subheadFuerte = Estilo(tamano: 15, peso: .semibold, relativo: .subheadline)
    static let label = Estilo(tamano: 14, peso: .semibold, relativo: .footnote)
    static let texto14 = Estilo(tamano: 14, peso: .regular, relativo: .footnote)
    static let footnote = Estilo(tamano: 13, peso: .regular, relativo: .footnote)
    static let footnoteFuerte = Estilo(tamano: 13, peso: .semibold, relativo: .footnote)
    static let chip = Estilo(tamano: 12, peso: .semibold, relativo: .caption)
    static let etiqueta = Estilo(tamano: 11, peso: .bold, relativo: .caption2)
    static let tab = Estilo(tamano: 10.5, peso: .semibold, relativo: .caption2)
}

extension View {
    /// Aplica un estilo del handoff (fuente, tracking).
    func estilo(_ e: Estilo) -> some View {
        font(e.font).tracking(e.tamano * e.tracking)
    }
}

extension Text {
    /// Para concatenar Text con estilos distintos.
    func estiloTexto(_ e: Estilo) -> Text {
        font(e.font).tracking(e.tamano * e.tracking)
    }
}

/// Medidas del handoff.
enum M {
    /// Margen horizontal: 20 (16 en SE, 24 en Pro Max).
    static func margen(_ ancho: CGFloat) -> CGFloat {
        ancho <= 375 ? 16 : (ancho >= 428 ? 24 : 20)
    }
    static let tabBar: CGFloat = 49
}

/// El ancho de la pantalla para los márgenes del handoff (SE / estándar / Pro Max).
private struct AnchoPantallaKey: EnvironmentKey {
    static let defaultValue: CGFloat = 393
}

extension EnvironmentValues {
    var anchoPantalla: CGFloat {
        get { self[AnchoPantallaKey.self] }
        set { self[AnchoPantallaKey.self] = newValue }
    }
    var margen: CGFloat { M.margen(anchoPantalla) }
}

/// UIKit: la barra de navegación nativa (pantallas de detalle) con la tipografía y los colores del
/// handoff: back con chevron y el título del padre en `primary-text`, título 17/600.
enum Apariencia {
    @MainActor
    static func aplicar() {
        let a = UINavigationBarAppearance()
        a.configureWithOpaqueBackground()
        a.backgroundColor = .tema(0xF7F1EA, 0x22140D)
        a.shadowColor = .clear
        let titulo = UIFont(name: Peso.semibold.nombre, size: 17) ?? .systemFont(ofSize: 17, weight: .semibold)
        a.titleTextAttributes = [.font: titulo, .foregroundColor: UIColor.tema(0x2A140C, 0xFBF7F2)]
        let grande = UIFont(name: Peso.bold.nombre, size: 34) ?? .systemFont(ofSize: 34, weight: .bold)
        a.largeTitleTextAttributes = [.font: grande, .kern: -0.85, .foregroundColor: UIColor.tema(0x2A140C, 0xFBF7F2)]
        let boton = UIBarButtonItemAppearance()
        let fuenteBoton = UIFont(name: Peso.regular.nombre, size: 17) ?? .systemFont(ofSize: 17)
        boton.normal.titleTextAttributes = [.font: fuenteBoton, .foregroundColor: UIColor.tema(0xB0441D, 0xE07A48)]
        a.buttonAppearance = boton
        a.backButtonAppearance = boton
        UINavigationBar.appearance().standardAppearance = a
        UINavigationBar.appearance().scrollEdgeAppearance = a
        UINavigationBar.appearance().compactAppearance = a
        UINavigationBar.appearance().tintColor = .tema(0xB0441D, 0xE07A48)
        UIRefreshControl.appearance().tintColor = .tema(0x7D6656, 0xA8907F)
    }
}
