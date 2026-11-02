import Charts
import SwiftUI

/// Lo que responde Tincho (POST /panel/hornero con formato «bloques»): el texto y, cuando
/// corresponde, cifras, gráficos, tablas, listas, una barra de progreso, botones y sugerencias.
/// Lo arma el motor de la web (apps/web/src/lib/server/hornero/motor/respuesta.ts).
struct RespuestaTincho: Decodable, Equatable {
    var texto: String
    var bloques: [BloqueTincho]?
    var acciones: [AccionTincho]?
    var sugerencias: [String]?
}

struct AccionTincho: Decodable, Equatable, Hashable { var texto: String; var ir: String }

enum BloqueTincho: Decodable, Equatable {
    struct Cifra: Decodable, Equatable { var titulo: String; var valor: String; var detalle: String?; var tono: String? }
    struct Barra: Decodable, Equatable { var etiqueta: String; var valor: Double; var destacado: Bool? }
    struct Punto: Decodable, Equatable { var etiqueta: String; var valor: Double }
    struct Item: Decodable, Equatable { var titulo: String; var detalle: String?; var insignia: String?; var tono: String? }

    case cifras([Cifra])
    case barras(titulo: String, unidad: String, items: [Barra])
    case linea(titulo: String, unidad: String, puntos: [Punto], anterior: [Double]?)
    case tabla(titulo: String, columnas: [String], filas: [[String]])
    case lista(titulo: String, items: [Item])
    case progreso(titulo: String, actual: Double, objetivo: Double, unidad: String, detalle: String?)
    case otro

    private enum K: String, CodingKey { case tipo, titulo, unidad, items, puntos, anterior, columnas, filas, actual, objetivo, detalle }

    init(from decoder: Decoder) throws {
        let c = try decoder.container(keyedBy: K.self)
        let titulo = (try? c.decode(String.self, forKey: .titulo)) ?? ""
        let unidad = (try? c.decode(String.self, forKey: .unidad)) ?? "numero"
        switch try c.decode(String.self, forKey: .tipo) {
        case "cifras": self = .cifras(try c.decode([Cifra].self, forKey: .items))
        case "barras": self = .barras(titulo: titulo, unidad: unidad, items: try c.decode([Barra].self, forKey: .items))
        case "linea":
            self = .linea(titulo: titulo, unidad: unidad, puntos: try c.decode([Punto].self, forKey: .puntos),
                          anterior: try? c.decode([Double].self, forKey: .anterior))
        case "tabla": self = .tabla(titulo: titulo, columnas: try c.decode([String].self, forKey: .columnas), filas: try c.decode([[String]].self, forKey: .filas))
        case "lista": self = .lista(titulo: titulo, items: try c.decode([Item].self, forKey: .items))
        case "progreso":
            self = .progreso(titulo: titulo, actual: try c.decode(Double.self, forKey: .actual), objetivo: try c.decode(Double.self, forKey: .objetivo),
                             unidad: unidad, detalle: try? c.decode(String.self, forKey: .detalle))
        default: self = .otro
        }
    }
}

// MARK: - Dibujo

/// Un bloque de la respuesta, en una tarjeta.
struct VistaBloque: View {
    var bloque: BloqueTincho

    var body: some View {
        switch bloque {
        case let .cifras(items): cifras(items)
        case let .barras(titulo, unidad, items): tarjeta(titulo) { barras(unidad, items) }
        case let .linea(titulo, unidad, puntos, anterior): tarjeta(titulo) { linea(unidad, puntos, anterior) }
        case let .tabla(titulo, columnas, filas): tarjeta(titulo) { tabla(columnas, filas) }
        case let .lista(titulo, items): tarjeta(titulo) { lista(items) }
        case let .progreso(titulo, actual, objetivo, unidad, detalle): tarjeta(titulo) { progreso(actual, objetivo, unidad, detalle) }
        case .otro: EmptyView()
        }
    }

    private func tarjeta<V: View>(_ titulo: String, @ViewBuilder _ contenido: () -> V) -> some View {
        VStack(alignment: .leading, spacing: 10) {
            if !titulo.isEmpty { Text(titulo).estilo(.footnoteFuerte).foregroundStyle(C.texto2) }
            contenido()
        }
        .padding(14)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(C.superficie, in: RoundedRectangle(cornerRadius: 18, style: .continuous))
        .overlay(RoundedRectangle(cornerRadius: 18, style: .continuous).strokeBorder(C.anillo, lineWidth: 1))
    }

    private func cifras(_ items: [BloqueTincho.Cifra]) -> some View {
        LazyVGrid(columns: [GridItem(.flexible(), spacing: 8), GridItem(.flexible(), spacing: 8)], spacing: 8) {
            ForEach(Array(items.enumerated()), id: \.offset) { _, c in
                VStack(alignment: .leading, spacing: 3) {
                    Text(c.titulo).estilo(.footnote).foregroundStyle(C.texto2).lineLimit(1)
                    Text(c.valor).estilo(.headline).monospacedDigit().foregroundStyle(C.texto)
                        .lineLimit(1).minimumScaleFactor(0.6)
                    if let d = c.detalle {
                        Text(d).estilo(.footnoteFuerte).foregroundStyle(VistaBloque.tinta(c.tono))
                    }
                }
                .frame(maxWidth: .infinity, alignment: .leading)
                .padding(12)
                .background(C.superficie, in: RoundedRectangle(cornerRadius: 14, style: .continuous))
                .overlay(RoundedRectangle(cornerRadius: 14, style: .continuous).strokeBorder(C.anillo, lineWidth: 1))
            }
        }
    }

    private func barras(_ unidad: String, _ items: [BloqueTincho.Barra]) -> some View {
        Chart(Array(items.enumerated()), id: \.offset) { i, b in
            BarMark(x: .value("Valor", b.valor), y: .value("Qué", "\(i)·\(b.etiqueta)"))
                .foregroundStyle(b.destacado == true ? C.acento : C.acento.opacity(0.45))
                .cornerRadius(5)
                .annotation(position: .trailing, alignment: .leading) {
                    Text(VistaBloque.corto(b.valor, unidad)).font(.caption2.weight(.semibold)).foregroundStyle(C.texto2)
                }
        }
        .chartYAxis {
            AxisMarks { v in
                AxisValueLabel {
                    if let s = v.as(String.self) {
                        Text(String(s.split(separator: "·", maxSplits: 1).last ?? "")).font(.caption2).foregroundStyle(C.texto2).lineLimit(1)
                    }
                }
            }
        }
        .chartXAxis(.hidden)
        .chartXScale(domain: 0...((items.map(\.valor).max() ?? 1) * 1.25))
        .frame(height: CGFloat(max(items.count, 2)) * 30)
        .accessibilityElement(children: .combine)
        .accessibilityLabel(items.map { "\($0.etiqueta): \(VistaBloque.corto($0.valor, unidad))" }.joined(separator: ", "))
    }

    private func linea(_ unidad: String, _ puntos: [BloqueTincho.Punto], _ anterior: [Double]?) -> some View {
        Chart {
            if let anterior, anterior.count == puntos.count {
                ForEach(Array(puntos.enumerated()), id: \.offset) { i, p in
                    LineMark(x: .value("Día", "\(i)·\(p.etiqueta)"), y: .value("Antes", anterior[i]), series: .value("Serie", "Antes"))
                        .foregroundStyle(C.texto3.opacity(0.7))
                        .lineStyle(StrokeStyle(lineWidth: 1.5, dash: [4, 4]))
                        .interpolationMethod(.monotone)
                }
            }
            ForEach(Array(puntos.enumerated()), id: \.offset) { i, p in
                AreaMark(x: .value("Día", "\(i)·\(p.etiqueta)"), y: .value("Valor", p.valor))
                    .foregroundStyle(LinearGradient(colors: [C.acento.opacity(0.25), C.acento.opacity(0)], startPoint: .top, endPoint: .bottom))
                    .interpolationMethod(.monotone)
                LineMark(x: .value("Día", "\(i)·\(p.etiqueta)"), y: .value("Valor", p.valor), series: .value("Serie", "Ahora"))
                    .foregroundStyle(C.acento)
                    .lineStyle(StrokeStyle(lineWidth: 2.5))
                    .interpolationMethod(.monotone)
            }
        }
        .chartXAxis {
            AxisMarks { v in
                AxisValueLabel {
                    if let s = v.as(String.self) {
                        Text(String(s.split(separator: "·", maxSplits: 1).last ?? "")).font(.caption2).foregroundStyle(C.texto2)
                    }
                }
            }
        }
        .chartYAxis {
            AxisMarks(position: .leading, values: .automatic(desiredCount: 3)) { v in
                AxisGridLine().foregroundStyle(C.linea)
                AxisValueLabel { if let d = v.as(Double.self) { Text(VistaBloque.corto(d, unidad)).font(.caption2).foregroundStyle(C.texto3) } }
            }
        }
        .frame(height: 150)
        .accessibilityElement(children: .combine)
        .accessibilityLabel(puntos.map { "\($0.etiqueta): \(VistaBloque.corto($0.valor, unidad))" }.joined(separator: ", "))
    }

    private func tabla(_ columnas: [String], _ filas: [[String]]) -> some View {
        ScrollView(.horizontal, showsIndicators: false) {
            Grid(alignment: .leading, horizontalSpacing: 14, verticalSpacing: 8) {
                GridRow {
                    ForEach(Array(columnas.enumerated()), id: \.offset) { i, c in
                        Text(c).estilo(.footnoteFuerte).foregroundStyle(C.texto2)
                            .gridColumnAlignment(i == 0 ? .leading : .trailing)
                    }
                }
                Divider().gridCellUnsizedAxes(.horizontal)
                ForEach(Array(filas.enumerated()), id: \.offset) { _, f in
                    GridRow {
                        ForEach(Array(f.enumerated()), id: \.offset) { i, v in
                            Text(v).estilo(i == 0 ? .fila : .texto14).monospacedDigit()
                                .foregroundStyle(v == "Sin stock" ? C.errorTinta : C.texto)
                                .lineLimit(2)
                                .frame(maxWidth: i == 0 ? 170 : nil, alignment: i == 0 ? .leading : .trailing)
                        }
                    }
                }
            }
        }
    }

    private func lista(_ items: [BloqueTincho.Item]) -> some View {
        VStack(alignment: .leading, spacing: 0) {
            ForEach(Array(items.enumerated()), id: \.offset) { i, it in
                if i > 0 { Separador(inset: 0) }
                HStack(alignment: .center, spacing: 10) {
                    Circle().fill(VistaBloque.tinta(it.tono)).frame(width: 7, height: 7)
                    VStack(alignment: .leading, spacing: 2) {
                        Text(it.titulo).estilo(.fila).foregroundStyle(C.texto).fixedSize(horizontal: false, vertical: true)
                        if let d = it.detalle {
                            Text(d).estilo(.footnote).foregroundStyle(C.texto2).fixedSize(horizontal: false, vertical: true)
                        }
                    }
                    Spacer(minLength: 6)
                    if let s = it.insignia {
                        Text(s).estilo(.footnoteFuerte).foregroundStyle(VistaBloque.tinta(it.tono))
                            .padding(.horizontal, 8).padding(.vertical, 4)
                            .background(VistaBloque.fondo(it.tono), in: Capsule())
                    }
                }
                .padding(.vertical, 9)
            }
        }
    }

    private func progreso(_ actual: Double, _ objetivo: Double, _ unidad: String, _ detalle: String?) -> some View {
        let f = objetivo > 0 ? min(1, actual / objetivo) : 0
        return VStack(alignment: .leading, spacing: 8) {
            HStack(alignment: .firstTextBaseline) {
                Text(VistaBloque.largo(actual, unidad)).estilo(.headline).monospacedDigit().foregroundStyle(C.texto)
                Text("de \(VistaBloque.largo(objetivo, unidad))").estilo(.footnote).foregroundStyle(C.texto2)
                Spacer()
                Text("\(Int((f * 100).rounded()))\u{202F}%").estilo(.footnoteFuerte).foregroundStyle(C.primarioTexto)
            }
            GeometryReader { g in
                ZStack(alignment: .leading) {
                    Capsule().fill(C.pista)
                    Capsule().fill(C.acento).frame(width: g.size.width * f)
                }
            }
            .frame(height: 8)
            if let detalle { Text(detalle).estilo(.footnote).foregroundStyle(C.texto2) }
        }
    }

    // MARK: Formato

    static func tinta(_ tono: String?) -> Color {
        switch tono {
        case "ok": C.okTinta
        case "mal": C.errorTinta
        case "aviso": C.avisoTinta
        default: C.texto2
        }
    }

    static func fondo(_ tono: String?) -> Color {
        switch tono {
        case "ok": C.okFondo
        case "mal": C.errorFondo
        case "aviso": C.avisoFondo
        default: C.pista
        }
    }

    /// Para ejes y etiquetas: «$ 1,2 M», «$ 340 mil», «12».
    static func corto(_ v: Double, _ unidad: String) -> String {
        guard unidad == "pesos" else { return Formato.numero(Int(v.rounded())) }
        if v >= 1_000_000 { return "$\u{202F}\((v / 1_000_000).formatted(.number.precision(.fractionLength(0...1)).locale(Locale(identifier: "es_AR")))) M" }
        if v >= 10_000 { return "$\u{202F}\(Int((v / 1000).rounded())) mil" }
        return Formato.pesos(Int(v.rounded()))
    }

    static func largo(_ v: Double, _ unidad: String) -> String {
        unidad == "pesos" ? Formato.pesos(Int(v.rounded())) : Formato.numero(Int(v.rounded()))
    }
}

/// Píldoras que bajan de renglón cuando no entran (las sugerencias de Tincho).
struct Fluir: Layout {
    var espacio: CGFloat = 8

    func sizeThatFits(proposal: ProposedViewSize, subviews: Subviews, cache _: inout ()) -> CGSize {
        let ancho = proposal.width ?? .infinity
        var x: CGFloat = 0, y: CGFloat = 0, alto: CGFloat = 0, maxX: CGFloat = 0
        for s in subviews {
            let t = s.sizeThatFits(.init(width: ancho, height: nil))
            if x > 0 && x + t.width > ancho { x = 0; y += alto + espacio; alto = 0 }
            x += t.width + espacio
            alto = max(alto, t.height)
            maxX = max(maxX, x - espacio)
        }
        return CGSize(width: proposal.width ?? maxX, height: y + alto)
    }

    func placeSubviews(in bounds: CGRect, proposal _: ProposedViewSize, subviews: Subviews, cache _: inout ()) {
        var x = bounds.minX, y = bounds.minY, alto: CGFloat = 0
        for s in subviews {
            let t = s.sizeThatFits(.init(width: bounds.width, height: nil))
            if x > bounds.minX && x + t.width > bounds.maxX { x = bounds.minX; y += alto + espacio; alto = 0 }
            s.place(at: CGPoint(x: x, y: y), proposal: .init(width: min(t.width, bounds.width), height: t.height))
            x += t.width + espacio
            alto = max(alto, t.height)
        }
    }
}
