import Charts
import SwiftUI

// Piezas para que las pantallas de sección no queden planas: la tarjeta héroe con el degradé
// terracota, indicadores con ícono, barras de progreso y gráficos (Swift Charts) con los colores
// del handoff.

/// Tarjeta destacada: degradé `primary`, monto grande en blanco, un renglón abajo y Tincho opcional.
struct TarjetaHeroe<Pie: View>: View {
    var etiqueta: String
    var valor: String
    var detalle: String? = nil
    var tincho: Tincho? = nil
    @ViewBuilder var pie: () -> Pie

    var body: some View {
        ZStack(alignment: .bottomTrailing) {
            if let tincho {
                Image(tincho.imagen).resizable().scaledToFit().frame(width: 110, height: 110)
                    .offset(x: 14, y: 18).opacity(0.95).accessibilityHidden(true)
            }
            VStack(alignment: .leading, spacing: 6) {
                Text(etiqueta).estilo(.footnoteFuerte).foregroundStyle(.white.opacity(0.85))
                // Con Tincho, el valor le deja lugar (como el detalle): «Plan Emprende mensual» le pasaba por encima.
                Text(valor).estilo(.montoPantalla).monospacedDigit().foregroundStyle(.white)
                    .lineLimit(1).minimumScaleFactor(tincho == nil ? 0.6 : 0.5)
                    .padding(.trailing, tincho == nil ? 0 : 90)
                    .contentTransition(.numericText())
                    .animation(Mov.toque, value: valor)
                if let detalle {
                    Text(detalle).estilo(.texto14).foregroundStyle(.white.opacity(0.9))
                        .padding(.trailing, tincho == nil ? 0 : 90)
                }
                pie()
            }
            .frame(maxWidth: .infinity, alignment: .leading)
        }
        .padding(18)
        .background {
            ZStack {
                C.primario
                RadialGradient(colors: [.white.opacity(0.18), .clear], center: .topTrailing, startRadius: 10, endRadius: 260)
            }
        }
        .clipShape(RoundedRectangle(cornerRadius: 24, style: .continuous))
        .shadow(color: C.sombraPrimario, radius: 14, y: 8)
        .aparecer()
    }
}

extension TarjetaHeroe where Pie == EmptyView {
    init(etiqueta: String, valor: String, detalle: String? = nil, tincho: Tincho? = nil) {
        self.init(etiqueta: etiqueta, valor: valor, detalle: detalle, tincho: tincho) { EmptyView() }
    }
}

/// Indicador con ícono en un cuadradito de color y un renglón de detalle.
struct IndicadorIcono: View {
    var simbolo: String
    var titulo: String
    var valor: String
    var detalle: String? = nil
    var tono: (Color, Color) = (C.crema, C.primarioTexto)
    @State private var visto = false
    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Image(systemName: simbolo).font(.system(size: 15, weight: .semibold)).foregroundStyle(tono.1)
                .symbolEffect(.bounce, value: visto)
                .frame(width: 32, height: 32)
                .background(tono.0, in: RoundedRectangle(cornerRadius: 9, style: .continuous))
            Text(valor).estilo(.indicador).monospacedDigit().foregroundStyle(C.texto)
                .lineLimit(1).minimumScaleFactor(0.6)
                .contentTransition(.numericText())
                .animation(Mov.toque, value: valor)
            Text(titulo).estilo(.footnote).foregroundStyle(C.texto2).lineLimit(2)
            if let detalle, !detalle.isEmpty {
                Text(detalle).estilo(.chip).foregroundStyle(C.texto3).lineLimit(2)
            }
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
        .padding(14)
        .tarjeta(radio: 18)
        .onAppear { visto = true }
    }
}

struct GrillaIndicadores: View {
    var items: [IndicadorIcono]
    var body: some View {
        // De a dos por fila y con el mismo alto (la que tiene detalle no queda más alta que su vecina).
        FilasParejas(cantidad: items.count, espacio: 10) { n in
            items[n].aparecer(0.05 + Double(n) * 0.05)
        }
    }
}

/// Barra de progreso con el degradé `accent → primary`.
struct BarraProgreso: View {
    var valor: Double
    var alto: CGFloat = 10
    var fondo: Color = C.pistaMeta
    @State private var lleno = false
    @Environment(\.accessibilityReduceMotion) private var sinMovimiento
    var body: some View {
        GeometryReader { g in
            ZStack(alignment: .leading) {
                Capsule().fill(fondo)
                Capsule().fill(LinearGradient(colors: [C.acento, Color(hex: 0xB0441D)], startPoint: .leading, endPoint: .trailing))
                    .frame(width: max(alto, g.size.width * min(1, max(0, valor)) * (lleno || sinMovimiento ? 1 : 0)))
            }
        }
        .frame(height: alto)
        .animation(Mov.crecer, value: valor)
        .onAppear { withAnimation(Mov.crecer.delay(0.15)) { lleno = true } }
        .accessibilityValue("\(Int((valor * 100).rounded())) %")
    }
}

/// Tarjeta con título de sección adentro.
struct TarjetaSeccion<Contenido: View>: View {
    var titulo: String
    var simbolo: String? = nil
    @ViewBuilder var contenido: () -> Contenido
    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack(spacing: 8) {
                if let simbolo {
                    Image(systemName: simbolo).font(.system(size: 13, weight: .semibold)).foregroundStyle(C.primarioTexto)
                }
                Text(titulo).estilo(.headline).foregroundStyle(C.texto)
            }
            contenido()
        }
        .padding(16)
        .frame(maxWidth: .infinity, alignment: .leading)
        .tarjeta(radio: 20)
        .aparecer(0.12)
    }
}

extension Formato {
    /// «$ 1,2 M», «$ 350 mil», «$ 900».
    static func compacto(_ n: Int) -> String {
        let a = abs(n)
        let signo = n < 0 ? "−" : ""
        if a >= 1_000_000 { return signo + "$" + fino + String(format: "%.1f", Double(a) / 1_000_000).replacingOccurrences(of: ".", with: ",") + " M" }
        if a >= 1_000 { return signo + "$" + fino + "\(a / 1_000) mil" }
        return signo + "$" + fino + "\(a)"
    }
}

/// Ícono por tipo de movimiento de plata.
func simboloMovimiento(_ tipo: String) -> String {
    switch tipo {
    case "mp", "payway", "modo", "uala": "creditcard.fill"
    case "andreani", "oca", "correo", "moto": "shippingbox.fill"
    case "transf": "arrow.left.arrow.right"
    case "gasto": "cart.fill"
    case "retiro": "banknote.fill"
    default: "circle.fill"
    }
}
