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
                Text(valor).estilo(.montoPantalla).monospacedDigit().foregroundStyle(.white)
                    .lineLimit(1).minimumScaleFactor(0.6)
                if let detalle {
                    Text(detalle).estilo(.texto14).foregroundStyle(.white.opacity(0.9))
                        .padding(.trailing, tincho == nil ? 0 : 90)
                }
                pie()
            }
            .frame(maxWidth: .infinity, alignment: .leading)
        }
        .padding(18)
        .background(
            ZStack {
                C.primario
                RadialGradient(colors: [.white.opacity(0.18), .clear], center: .topTrailing, startRadius: 10, endRadius: 260)
            },
            in: RoundedRectangle(cornerRadius: 24, style: .continuous))
        .clipShape(RoundedRectangle(cornerRadius: 24, style: .continuous))
        .shadow(color: C.sombraPrimario, radius: 14, y: 8)
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
    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Image(systemName: simbolo).font(.system(size: 15, weight: .semibold)).foregroundStyle(tono.1)
                .frame(width: 32, height: 32)
                .background(tono.0, in: RoundedRectangle(cornerRadius: 9, style: .continuous))
            Text(valor).estilo(.indicador).monospacedDigit().foregroundStyle(C.texto)
                .lineLimit(1).minimumScaleFactor(0.6)
            Text(titulo).estilo(.footnote).foregroundStyle(C.texto2).lineLimit(2)
            if let detalle, !detalle.isEmpty {
                Text(detalle).estilo(.chip).foregroundStyle(C.texto3).lineLimit(2)
            }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(14)
        .tarjeta(radio: 18)
    }
}

struct GrillaIndicadores: View {
    var items: [IndicadorIcono]
    var body: some View {
        LazyVGrid(columns: [GridItem(.flexible(), spacing: 10), GridItem(.flexible())], spacing: 10) {
            ForEach(Array(items.enumerated()), id: \.offset) { _, i in i }
        }
    }
}

/// Barra de progreso con el degradé `accent → primary`.
struct BarraProgreso: View {
    var valor: Double
    var alto: CGFloat = 10
    var fondo: Color = C.pistaMeta
    var body: some View {
        GeometryReader { g in
            ZStack(alignment: .leading) {
                Capsule().fill(fondo)
                Capsule().fill(LinearGradient(colors: [C.acento, Color(hex: 0xB0441D)], startPoint: .leading, endPoint: .trailing))
                    .frame(width: max(alto, g.size.width * min(1, max(0, valor))))
            }
        }
        .frame(height: alto)
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
    }
}

// MARK: - Gráficos

/// Flujo de caja: lo que entra y lo que sale por día (barras enfrentadas).
struct GraficoFlujo: View {
    struct Dia: Identifiable { var dia: Int; var entra: Int; var sale: Int; var id: Int { dia } }
    var dias: [Dia]
    var body: some View {
        Chart {
            ForEach(dias) { d in
                BarMark(x: .value("Día", d.dia), y: .value("Entra", d.entra))
                    .foregroundStyle(C.ok).cornerRadius(3)
                BarMark(x: .value("Día", d.dia), y: .value("Sale", -d.sale))
                    .foregroundStyle(C.error.opacity(0.8)).cornerRadius(3)
            }
            RuleMark(y: .value("Cero", 0)).foregroundStyle(C.linea2)
        }
        .chartYAxis {
            AxisMarks(position: .leading) { v in
                AxisGridLine().foregroundStyle(C.linea)
                AxisValueLabel { if let n = v.as(Int.self) { Text(Formato.compacto(n)).font(.caption2).foregroundStyle(C.texto3) } }
            }
        }
        .chartXAxis {
            AxisMarks(values: .automatic(desiredCount: 7)) { v in
                AxisValueLabel { if let n = v.as(Int.self) { Text("\(n)").font(.caption2).foregroundStyle(C.texto3) } }
            }
        }
        .frame(height: 170)
    }
}

/// Dona de medios de pago.
struct DonaMedios: View {
    struct Parte: Identifiable { var nombre: String; var monto: Int; var id: String { nombre } }
    var partes: [Parte]
    static let colores: [Color] = [Color(hex: 0xD9652F), Color(hex: 0xE3A93B), Color(hex: 0x6B5244), Color(hex: 0x2F8F63), Color(hex: 0x8E3A16), Color(hex: 0xB5A193)]

    var body: some View {
        let total = max(1, partes.reduce(0) { $0 + $1.monto })
        HStack(spacing: 18) {
            Chart(Array(partes.enumerated()), id: \.element.id) { i, p in
                SectorMark(angle: .value("Monto", p.monto), innerRadius: .ratio(0.62), angularInset: 2)
                    .foregroundStyle(DonaMedios.colores[i % DonaMedios.colores.count])
                    .cornerRadius(4)
            }
            .frame(width: 120, height: 120)
            VStack(alignment: .leading, spacing: 8) {
                ForEach(Array(partes.enumerated()), id: \.element.id) { i, p in
                    HStack(spacing: 8) {
                        Circle().fill(DonaMedios.colores[i % DonaMedios.colores.count]).frame(width: 9, height: 9)
                        Text(p.nombre).estilo(.texto14).foregroundStyle(C.texto).lineLimit(1)
                        Spacer(minLength: 4)
                        Text("\(Int((Double(p.monto) / Double(total) * 100).rounded())) %").estilo(.footnoteFuerte)
                            .monospacedDigit().foregroundStyle(C.texto2)
                    }
                }
            }
        }
    }
}

/// Barras horizontales por cantidad de estrellas (5 a 1).
struct BarrasEstrellas: View {
    var porEstrellas: [Int]
    var body: some View {
        let total = max(1, porEstrellas.reduce(0, +))
        VStack(spacing: 6) {
            ForEach((1...5).reversed(), id: \.self) { n in
                let c = porEstrellas.indices.contains(n - 1) ? porEstrellas[n - 1] : 0
                HStack(spacing: 8) {
                    Text("\(n)").estilo(.footnoteFuerte).foregroundStyle(C.texto2).frame(width: 12)
                    Image(systemName: "star.fill").font(.system(size: 10)).foregroundStyle(C.aviso)
                    BarraProgreso(valor: Double(c) / Double(total), alto: 8)
                    Text("\(c)").estilo(.footnote).monospacedDigit().foregroundStyle(C.texto3).frame(width: 28, alignment: .trailing)
                }
            }
        }
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
