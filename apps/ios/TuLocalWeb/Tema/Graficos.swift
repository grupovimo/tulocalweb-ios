import Charts
import SwiftUI

// Gráficos con Swift Charts y el movimiento de la app. Todos crecen desde cero la primera vez, se
// tocan o se arrastran para ver cada valor (con un toque háptico) y respetan «Reducir movimiento».

// MARK: - Movimiento

enum Mov {
    /// Entrada de tarjetas: suave y corta.
    static let entrada = Animation.smooth(duration: 0.5)
    /// Gráficos y barras que crecen.
    static let crecer = Animation.smooth(duration: 0.9)
    /// Cambios por un toque (elegir, filtrar).
    static let toque = Animation.snappy(duration: 0.25)
}

/// Entra subiendo apenas y apareciendo, con un retraso para escalonar una grilla.
private struct Aparecer: ViewModifier {
    var retraso: Double
    @State private var visto = false
    @Environment(\.accessibilityReduceMotion) private var sinMovimiento
    func body(content: Content) -> some View {
        content
            .opacity(visto || sinMovimiento ? 1 : 0)
            .offset(y: visto || sinMovimiento ? 0 : 14)
            .onAppear {
                guard !visto else { return }
                withAnimation(Mov.entrada.delay(retraso)) { visto = true }
            }
    }
}

extension View {
    func aparecer(_ retraso: Double = 0) -> some View { modifier(Aparecer(retraso: retraso)) }
}

/// 0 → 1 la primera vez que aparece (1 de entrada con «Reducir movimiento»).
private struct Crecimiento: ViewModifier {
    @Binding var factor: Double
    var retraso: Double
    @Environment(\.accessibilityReduceMotion) private var sinMovimiento
    func body(content: Content) -> some View {
        content.onAppear {
            guard factor < 1 else { return }
            if sinMovimiento { factor = 1 } else { withAnimation(Mov.crecer.delay(retraso)) { factor = 1 } }
        }
    }
}

extension View {
    fileprivate func crecer(_ factor: Binding<Double>, retraso: Double = 0.1) -> some View {
        modifier(Crecimiento(factor: factor, retraso: retraso))
    }
}

/// El globo con el valor del punto elegido.
struct GloboGrafico: View {
    var titulo: String
    var lineas: [(String, String, Color)]
    var body: some View {
        VStack(alignment: .leading, spacing: 2) {
            Text(titulo).estilo(.chip).foregroundStyle(C.texto2)
            ForEach(Array(lineas.enumerated()), id: \.offset) { _, l in
                HStack(spacing: 5) {
                    if !l.0.isEmpty { Text(l.0).estilo(.chip).foregroundStyle(C.texto3) }
                    Text(l.1).estilo(.footnoteFuerte).monospacedDigit().foregroundStyle(l.2)
                }
            }
        }
        .padding(.horizontal, 10).padding(.vertical, 7)
        .background(C.superficie, in: RoundedRectangle(cornerRadius: 10, style: .continuous))
        .overlay(RoundedRectangle(cornerRadius: 10, style: .continuous).strokeBorder(C.anillo, lineWidth: 1))
        .shadow(color: Color(hex: 0x2A140C, alfa: 0.12), radius: 8, y: 3)
    }
}

private let fitDentro = AnnotationOverflowResolution(x: .fit(to: .chart), y: .fit(to: .chart))

// MARK: - Ventas

/// Los colores del gráfico de ventas: los de siempre, o los de encima del color de la tienda (la
/// vidriera del Inicio: línea y punto en «sobre», relleno en velo, guías en el borde del vidrio).
struct ColoresGrafico {
    var linea: Color = C.acento
    var anterior: Color = Color(hex: 0x8A7364)
    /// nil: el degradé de la línea (22 % → 0).
    var relleno: Color? = nil
    var guia: Color = Color.tema(0x3A1D12, 0xFBF7F2, alfa: 0.07)
    var bordePunto: Color = .white
    static let estandar = ColoresGrafico()
}

/// El gráfico de ventas de 2a: el período en `accent` (3 pt, relleno degradé 22 % → 0 y punto final
/// de 9 con borde blanco de 3) y el anterior punteado (`#8A7364`, 2 pt, 4/5), con tres guías.
/// Arrastrando se ve cuánto se vendió en cada punto y cuánto en el mismo punto del período anterior.
struct GraficoVentas: View {
    var actual: [Double]
    var anterior: [Double]
    var alto: CGFloat = 130
    /// El nombre de cada punto al tocarlo (el día o la hora); nil: sin título.
    var etiqueta: ((Int) -> String)? = nil
    var colores: ColoresGrafico = .estandar
    /// Lo que dice VoiceOver del gráfico (nil: «Gráfico de ventas»).
    var descripcion: String? = nil

    @State private var elegido: Int?
    @State private var factor = 0.0

    var body: some View {
        let tope = max(actual.max() ?? 0, anterior.max() ?? 0, 1)
        let largo = max(actual.count, anterior.count, 2)
        Chart {
            ForEach(Array(anterior.enumerated()), id: \.offset) { i, v in
                LineMark(x: .value("Punto", i), y: .value("Monto", v * factor), series: .value("Serie", "Anterior"))
                    .foregroundStyle(colores.anterior)
                    .lineStyle(StrokeStyle(lineWidth: 2, lineCap: .round, dash: [4, 5]))
                    .interpolationMethod(.catmullRom)
            }
            ForEach(Array(actual.enumerated()), id: \.offset) { i, v in
                AreaMark(x: .value("Punto", i), y: .value("Monto", v * factor))
                    .foregroundStyle(LinearGradient(colors: colores.relleno.map { [$0, $0] } ?? [colores.linea.opacity(0.22), colores.linea.opacity(0)],
                                                    startPoint: .top, endPoint: .bottom))
                    .interpolationMethod(.catmullRom)
                LineMark(x: .value("Punto", i), y: .value("Monto", v * factor), series: .value("Serie", "Actual"))
                    .foregroundStyle(colores.linea)
                    .lineStyle(StrokeStyle(lineWidth: 3, lineCap: .round))
                    .interpolationMethod(.catmullRom)
            }
            if let e = elegido, actual.indices.contains(e) {
                RuleMark(x: .value("Punto", e))
                    .foregroundStyle(C.linea2)
                    .lineStyle(StrokeStyle(lineWidth: 1, dash: [3, 3]))
                    .annotation(position: .top, spacing: 0, overflowResolution: fitDentro) { globo(e) }
                PointMark(x: .value("Punto", e), y: .value("Monto", actual[e] * factor))
                    .symbol { punto }
            } else if let u = actual.indices.last {
                PointMark(x: .value("Punto", u), y: .value("Monto", actual[u] * factor))
                    .symbol { punto }
            }
        }
        .chartXScale(domain: 0...(largo - 1))
        .chartYScale(domain: 0...(tope * 1.08))
        .chartXAxis(.hidden)
        .chartYAxis {
            AxisMarks(values: [tope * 0.15, tope * 0.55, tope * 0.95]) { _ in
                AxisGridLine(stroke: StrokeStyle(lineWidth: 1)).foregroundStyle(colores.guia)
            }
        }
        .chartLegend(.hidden)
        .chartXSelection(value: $elegido)
        .sensoryFeedback(.selection, trigger: elegido)
        .frame(height: alto)
        .crecer($factor)
        .animation(Mov.crecer, value: actual)
        .accessibilityElement()
        .accessibilityLabel(descripcion ?? "Gráfico de ventas")
        .accessibilityValue(actual.last.map { "Último punto: \(Formato.pesos($0))" } ?? "")
    }

    private var punto: some View {
        Circle().fill(colores.linea).frame(width: 9, height: 9)
            .overlay(Circle().stroke(colores.bordePunto, lineWidth: 3))
    }

    private func globo(_ i: Int) -> some View {
        var lineas: [(String, String, Color)] = [("", Formato.pesos(actual[i]), C.texto)]
        if anterior.indices.contains(i) { lineas.append(("Antes", Formato.pesos(anterior[i]), C.texto2)) }
        return GloboGrafico(titulo: etiqueta?(i) ?? "", lineas: lineas)
    }

    /// Los nombres de los puntos de un período: la hora («14 h») hoy, el día («5/11») si no.
    static func etiquetas(desde: Date, hasta: Date, puntos: Int, porHora: Bool) -> (Int) -> String {
        let total = hasta.timeIntervalSince(desde)
        return { i in
            let f = desde.addingTimeInterval(total * Double(i) / Double(max(1, puntos - 1)))
            let c = Formato.calendario
            return porHora ? "\(c.component(.hour, from: f))\(Formato.fino)h"
                : "\(c.component(.day, from: f))/\(c.component(.month, from: f))"
        }
    }
}

// MARK: - Flujo de caja

/// Flujo de caja: lo que entra y lo que sale por día (barras enfrentadas). Tocando un día se ve el
/// detalle.
struct GraficoFlujo: View {
    struct Dia: Identifiable { var dia: Int; var entra: Int; var sale: Int; var id: Int { dia } }
    var dias: [Dia]
    @State private var elegido: Int?
    @State private var factor = 0.0

    var body: some View {
        Chart {
            ForEach(dias) { d in
                let apagado = elegido != nil && elegido != d.dia
                BarMark(x: .value("Día", d.dia), y: .value("Entra", Double(d.entra) * factor))
                    .foregroundStyle(C.ok.opacity(apagado ? 0.3 : 1)).cornerRadius(3)
                BarMark(x: .value("Día", d.dia), y: .value("Sale", -Double(d.sale) * factor))
                    .foregroundStyle(C.error.opacity(apagado ? 0.25 : 0.8)).cornerRadius(3)
            }
            RuleMark(y: .value("Cero", 0)).foregroundStyle(C.linea2)
            if let e = elegido, let d = dias.first(where: { $0.dia == e }) {
                RuleMark(x: .value("Día", e))
                    .foregroundStyle(.clear)
                    .annotation(position: .top, spacing: 0, overflowResolution: fitDentro) {
                        GloboGrafico(titulo: "Día \(e)", lineas: [
                            ("Entra", Formato.pesos(d.entra), C.okTinta),
                            ("Sale", Formato.pesos(d.sale), C.errorTinta),
                        ])
                    }
            }
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
        .chartXSelection(value: $elegido)
        .sensoryFeedback(.selection, trigger: elegido)
        .animation(Mov.toque, value: elegido)
        .frame(height: 170)
        .crecer($factor)
    }
}

/// Cómo queda la plata día a día si entra y sale lo previsto (lo acumulado del flujo).
struct GraficoSaldo: View {
    struct Dia: Identifiable { var dia: Int; var saldo: Int; var id: Int { dia } }
    var dias: [Dia]
    @State private var elegido: Int?
    @State private var factor = 0.0

    init(flujo: [GraficoFlujo.Dia]) {
        var suma = 0
        dias = flujo.map { d in
            suma += d.entra - d.sale
            return Dia(dia: d.dia, saldo: suma)
        }
    }

    var body: some View {
        let positivo = (dias.last?.saldo ?? 0) >= 0
        let tono = positivo ? C.ok : C.error
        Chart {
            ForEach(dias) { d in
                AreaMark(x: .value("Día", d.dia), y: .value("Saldo", Double(d.saldo) * factor))
                    .foregroundStyle(LinearGradient(colors: [tono.opacity(0.22), tono.opacity(0)], startPoint: .top, endPoint: .bottom))
                    .interpolationMethod(.monotone)
                LineMark(x: .value("Día", d.dia), y: .value("Saldo", Double(d.saldo) * factor))
                    .foregroundStyle(tono)
                    .lineStyle(StrokeStyle(lineWidth: 2.5, lineCap: .round))
                    .interpolationMethod(.monotone)
            }
            RuleMark(y: .value("Cero", 0)).foregroundStyle(C.linea2)
            if let e = elegido, let d = dias.first(where: { $0.dia == e }) {
                RuleMark(x: .value("Día", e))
                    .foregroundStyle(C.linea2)
                    .lineStyle(StrokeStyle(lineWidth: 1, dash: [3, 3]))
                    .annotation(position: .top, spacing: 0, overflowResolution: fitDentro) {
                        GloboGrafico(titulo: "Hasta el día \(e)",
                                     lineas: [("", (d.saldo < 0 ? "−" : "+") + Formato.pesos(abs(d.saldo)), d.saldo < 0 ? C.errorTinta : C.okTinta)])
                    }
                PointMark(x: .value("Día", e), y: .value("Saldo", Double(d.saldo) * factor))
                    .foregroundStyle(tono).symbolSize(60)
            }
        }
        .chartYAxis {
            AxisMarks(position: .leading, values: .automatic(desiredCount: 4)) { v in
                AxisGridLine().foregroundStyle(C.linea)
                AxisValueLabel { if let n = v.as(Int.self) { Text(Formato.compacto(n)).font(.caption2).foregroundStyle(C.texto3) } }
            }
        }
        .chartXAxis {
            AxisMarks(values: .automatic(desiredCount: 7)) { v in
                AxisValueLabel { if let n = v.as(Int.self) { Text("\(n)").font(.caption2).foregroundStyle(C.texto3) } }
            }
        }
        .chartXSelection(value: $elegido)
        .sensoryFeedback(.selection, trigger: elegido)
        .frame(height: 140)
        .crecer($factor)
    }
}

// MARK: - Dona

/// Dona de medios de pago (o canales): tocando una porción, o su renglón, se destaca y se ve su parte
/// en el centro.
struct DonaMedios: View {
    struct Parte: Identifiable { var nombre: String; var monto: Int; var id: String { nombre } }
    var partes: [Parte]
    static let colores: [Color] = [Color(hex: 0xD9652F), Color(hex: 0xE3A93B), Color(hex: 0x6B5244), Color(hex: 0x2F8F63), Color(hex: 0x8E3A16), Color(hex: 0xB5A193)]

    @State private var angulo: Int?
    @State private var elegida: Int?
    @State private var factor = 0.0

    var body: some View {
        let total = max(1, partes.reduce(0) { $0 + $1.monto })
        let destacada = elegida ?? partes.indices.max(by: { partes[$0].monto < partes[$1].monto })
        HStack(spacing: 18) {
            ZStack {
                Chart(Array(partes.enumerated()), id: \.element.id) { i, p in
                    SectorMark(angle: .value("Monto", p.monto),
                               innerRadius: .ratio(elegida == i ? 0.58 : 0.62),
                               outerRadius: .ratio(elegida == i ? 1 : 0.94),
                               angularInset: 2)
                        .foregroundStyle(DonaMedios.colores[i % DonaMedios.colores.count])
                        .opacity(elegida == nil || elegida == i ? 1 : 0.4)
                        .cornerRadius(4)
                }
                .chartAngleSelection(value: $angulo)
                .chartLegend(.hidden)
                if let d = destacada, partes.indices.contains(d) {
                    VStack(spacing: 0) {
                        Text("\(porcentaje(partes[d].monto, total)) %").estilo(.headline).monospacedDigit()
                            .foregroundStyle(C.texto).contentTransition(.numericText())
                        Text(partes[d].nombre).estilo(.chip).foregroundStyle(C.texto2).lineLimit(1)
                            .minimumScaleFactor(0.7)
                    }
                    .frame(width: 64)
                }
            }
            .frame(width: 120, height: 120)
            .scaleEffect(0.6 + 0.4 * factor)
            .rotationEffect(.degrees(-90 * (1 - factor)))
            .opacity(factor)
            VStack(alignment: .leading, spacing: 8) {
                ForEach(Array(partes.enumerated()), id: \.element.id) { i, p in
                    Button {
                        withAnimation(Mov.toque) { elegida = elegida == i ? nil : i }
                    } label: {
                        HStack(spacing: 8) {
                            Circle().fill(DonaMedios.colores[i % DonaMedios.colores.count]).frame(width: 9, height: 9)
                            Text(p.nombre).estilo(.texto14).foregroundStyle(C.texto).lineLimit(1)
                            Spacer(minLength: 4)
                            Text("\(porcentaje(p.monto, total)) %").estilo(.footnoteFuerte)
                                .monospacedDigit().foregroundStyle(C.texto2)
                        }
                        .opacity(elegida == nil || elegida == i ? 1 : 0.5)
                        .contentShape(Rectangle())
                    }
                    .buttonStyle(.plain)
                }
            }
        }
        .onChange(of: angulo) { _, a in
            guard let a else { return }
            var suma = 0
            for (i, p) in partes.enumerated() {
                suma += p.monto
                if a <= suma {
                    withAnimation(Mov.toque) { elegida = elegida == i ? nil : i }
                    break
                }
            }
        }
        .sensoryFeedback(.selection, trigger: elegida)
        .crecer($factor)
    }

    private func porcentaje(_ n: Int, _ total: Int) -> Int { Int((Double(n) / Double(total) * 100).rounded()) }
}

// MARK: - Mes a mes

/// Barras por mes con varias series para elegir (ventas, clientes, visitas…). Tocando un mes se ve el
/// valor; el último mes va destacado.
struct GraficoMeses: View {
    struct Serie: Identifiable {
        var id: String
        var nombre: String
        var valores: [Int?]
        /// El valor en el globo; el eje usa `eje`.
        var formato: (Int) -> String
        var eje: (Int) -> String
    }
    struct Barra: Identifiable { var mes: String; var valor: Int; var id: String { mes } }

    var meses: [String]
    var series: [Serie]
    @State private var serieId: String = ""
    @State private var elegido: String?
    @State private var factor = 0.0

    var body: some View {
        let serie = series.first(where: { $0.id == serieId }) ?? series[0]
        let barras = zip(meses, serie.valores).compactMap { m, v in v.map { Barra(mes: m, valor: $0) } }
        VStack(alignment: .leading, spacing: 12) {
            if series.count > 1 {
                Segmentado(opciones: series.map { ($0.id, $0.nombre) }, elegido: Binding(
                    get: { serieId.isEmpty ? series[0].id : serieId },
                    set: { nuevo in elegido = nil; serieId = nuevo }))
            }
            Chart(barras) { b in
                let ultimo = b.mes == barras.last?.mes
                BarMark(x: .value("Mes", b.mes), y: .value(serie.nombre, Double(b.valor) * factor))
                    .foregroundStyle(ultimo || elegido == b.mes
                                     ? AnyShapeStyle(LinearGradient(colors: [C.acento, Color(hex: 0xB0441D)], startPoint: .top, endPoint: .bottom))
                                     : AnyShapeStyle(C.acento.opacity(elegido == nil ? 0.45 : 0.25)))
                    .cornerRadius(5)
                    .annotation(position: .top, spacing: 4, overflowResolution: fitDentro) {
                        if elegido == b.mes {
                            GloboGrafico(titulo: b.mes, lineas: [("", serie.formato(b.valor), C.texto)])
                        }
                    }
            }
            .chartYAxis {
                AxisMarks(position: .leading, values: .automatic(desiredCount: 4)) { v in
                    AxisGridLine().foregroundStyle(C.linea)
                    AxisValueLabel { if let n = v.as(Int.self) { Text(serie.eje(n)).font(.caption2).foregroundStyle(C.texto3) } }
                }
            }
            .chartXAxis {
                AxisMarks { _ in AxisValueLabel().font(.caption2).foregroundStyle(C.texto3) }
            }
            .chartXSelection(value: $elegido)
            .sensoryFeedback(.selection, trigger: elegido)
            .animation(Mov.crecer, value: serieId)
            .animation(Mov.toque, value: elegido)
            .frame(height: 180)
            .crecer($factor)
        }
    }
}

// MARK: - Embudo de visitas

/// Cómo recorren la tienda: de cada 100 que entran, cuántos ven el catálogo, un producto, el carrito
/// y el pago.
struct EmbudoVisitas: View {
    struct Paso: Identifiable { var nombre: String; var vistas: Int; var id: String { nombre } }
    var pasos: [Paso]
    @State private var factor = 0.0

    init(paginas: [Tablero.Visitas.Pagina]) {
        let orden = [("inicio", "Inicio"), ("catalogo", "Catálogo"), ("producto", "Producto"), ("carrito", "Carrito"), ("checkout", "Pago")]
        pasos = orden.compactMap { id, nombre in
            paginas.first(where: { $0.pagina == id }).map { Paso(nombre: nombre, vistas: $0.vistas) }
        }
    }

    var body: some View {
        let tope = max(1, pasos.map(\.vistas).max() ?? 1)
        Chart(Array(pasos.enumerated()), id: \.element.id) { i, p in
            BarMark(x: .value("Vistas", Double(p.vistas) * factor), y: .value("Página", p.nombre))
                .foregroundStyle(C.acento.opacity(1 - Double(i) * 0.15))
                .cornerRadius(6)
                .annotation(position: .trailing, spacing: 6) {
                    Text("\(Formato.numero(p.vistas)) · \(Int((Double(p.vistas) / Double(tope) * 100).rounded())) %")
                        .estilo(.chip).monospacedDigit().foregroundStyle(C.texto2)
                        .opacity(factor)
                }
        }
        .chartXScale(domain: 0...(Double(tope) * 1.45))
        .chartXAxis(.hidden)
        .chartYAxis {
            AxisMarks { _ in AxisValueLabel().font(.footnote.weight(.semibold)).foregroundStyle(C.texto) }
        }
        .frame(height: CGFloat(pasos.count) * 38)
        .crecer($factor)
    }
}

// MARK: - Barras

/// Barras horizontales por cantidad de estrellas (5 a 1).
struct BarrasEstrellas: View {
    var porEstrellas: [Int]
    @State private var factor = 0.0
    struct Fila: Identifiable { var estrellas: Int; var cantidad: Int; var id: Int { estrellas } }

    var body: some View {
        let filas = (1...5).reversed().map { n in Fila(estrellas: n, cantidad: porEstrellas.indices.contains(n - 1) ? porEstrellas[n - 1] : 0) }
        let tope = max(1, filas.map(\.cantidad).max() ?? 1)
        Chart(filas) { f in
            BarMark(x: .value("Cantidad", Double(f.cantidad) * factor), y: .value("Estrellas", "\(f.estrellas) ★"))
                .foregroundStyle(LinearGradient(colors: [C.acento, Color(hex: 0xB0441D)], startPoint: .leading, endPoint: .trailing))
                .cornerRadius(4)
                .annotation(position: .trailing, spacing: 6) {
                    Text("\(f.cantidad)").estilo(.footnote).monospacedDigit().foregroundStyle(C.texto3).opacity(factor)
                }
        }
        .chartXScale(domain: 0...(Double(tope) * 1.15))
        .chartXAxis(.hidden)
        .chartYAxis {
            AxisMarks { _ in AxisValueLabel().font(.footnote.weight(.semibold)).foregroundStyle(C.texto2) }
        }
        .frame(height: 150)
        .crecer($factor)
    }
}

/// Anillo de avance (la meta del mes): se llena al aparecer.
struct AnilloAvance: View {
    var valor: Double
    var tam: CGFloat = 76
    var grosor: CGFloat = 9
    @State private var factor = 0.0

    var body: some View {
        let v = min(1, max(0, valor))
        ZStack {
            Circle().stroke(C.pistaMeta, lineWidth: grosor)
            Circle().trim(from: 0, to: v * factor)
                .stroke(AngularGradient(colors: [C.acento, Color(hex: 0xB0441D), C.acento], center: .center),
                        style: StrokeStyle(lineWidth: grosor, lineCap: .round))
                .rotationEffect(.degrees(-90))
            Text("\(Int((valor * 100 * factor).rounded())) %").estilo(.headline).monospacedDigit().foregroundStyle(C.texto)
                .contentTransition(.numericText())
        }
        .frame(width: tam, height: tam)
        .crecer($factor, retraso: 0.2)
        .accessibilityElement()
        .accessibilityLabel("Avance")
        .accessibilityValue("\(Int((valor * 100).rounded())) %")
    }
}
