import SwiftUI

/// El gráfico de ventas de 2a: el período en `accent` (3 pt, relleno degradé 22 % → 0 y punto final
/// de 6 con borde blanco de 3) y el anterior punteado (`#8A7364`, 2 pt, 4/5). Líneas guía a 30, 70 y
/// 110 de 130.
struct GraficoVentas: View {
    var actual: [Double]
    var anterior: [Double]
    var alto: CGFloat = 130

    var body: some View {
        Canvas { ctx, tam in
            let guias = Path { p in
                for y in [30.0, 70.0, 110.0] {
                    p.move(to: CGPoint(x: 0, y: y / 130 * tam.height))
                    p.addLine(to: CGPoint(x: tam.width, y: y / 130 * tam.height))
                }
            }
            ctx.stroke(guias, with: .color(Color.tema(0x3A1D12, 0xFBF7F2, alfa: 0.07)), lineWidth: 1)

            let maximo = max(actual.max() ?? 0, anterior.max() ?? 0, 1)
            let base = 110 / 130 * tam.height
            let techo = 20 / 130 * tam.height
            func puntos(_ s: [Double]) -> [CGPoint] {
                guard s.count > 1 else { return [] }
                return s.enumerated().map { i, v in
                    CGPoint(x: 2 + CGFloat(i) / CGFloat(s.count - 1) * (tam.width - 4),
                            y: base - CGFloat(v / maximo) * (base - techo))
                }
            }
            let pa = puntos(anterior)
            if pa.count > 1 {
                ctx.stroke(GraficoVentas.curva(pa), with: .color(Color(hex: 0x8A7364)),
                           style: StrokeStyle(lineWidth: 2, lineCap: .round, dash: [4, 5]))
            }
            let pt = puntos(actual)
            if pt.count > 1, let ultimo = pt.last {
                var relleno = GraficoVentas.curva(pt)
                relleno.addLine(to: CGPoint(x: ultimo.x, y: base))
                relleno.addLine(to: CGPoint(x: pt[0].x, y: base))
                relleno.closeSubpath()
                ctx.fill(relleno, with: .linearGradient(
                    Gradient(colors: [C.acento.opacity(0.22), C.acento.opacity(0)]),
                    startPoint: CGPoint(x: 0, y: techo), endPoint: CGPoint(x: 0, y: base)))
                ctx.stroke(GraficoVentas.curva(pt), with: .color(C.acento), style: StrokeStyle(lineWidth: 3, lineCap: .round))
                let punto = Path(ellipseIn: CGRect(x: ultimo.x - 7.5, y: ultimo.y - 7.5, width: 15, height: 15))
                ctx.fill(punto, with: .color(.white))
                ctx.fill(Path(ellipseIn: CGRect(x: ultimo.x - 4.5, y: ultimo.y - 4.5, width: 9, height: 9)), with: .color(C.acento))
            }
        }
        .frame(height: alto)
        .accessibilityHidden(true)
    }

    /// Curva suave por los puntos (Catmull-Rom a Bézier).
    static func curva(_ p: [CGPoint]) -> Path {
        var path = Path()
        guard let primero = p.first else { return path }
        path.move(to: primero)
        for i in 0..<(p.count - 1) {
            let p0 = i > 0 ? p[i - 1] : p[i]
            let p1 = p[i]
            let p2 = p[i + 1]
            let p3 = i + 2 < p.count ? p[i + 2] : p2
            let c1 = CGPoint(x: p1.x + (p2.x - p0.x) / 6, y: p1.y + (p2.y - p0.y) / 6)
            let c2 = CGPoint(x: p2.x - (p3.x - p1.x) / 6, y: p2.y - (p3.y - p1.y) / 6)
            path.addCurve(to: p2, control1: c1, control2: c2)
        }
        return path
    }
}
