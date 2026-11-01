import ActivityKit
import AppIntents
import SwiftUI
import WidgetKit

@main
struct WidgetsTuLocalWeb: WidgetBundle {
    var body: some Widget {
        VentasDeHoy()
        ActividadPedidos()
    }
}

// MARK: - 9a · Widget de ventas de hoy (chico y mediano)

struct EntradaHoy: TimelineEntry {
    var date: Date
    var datos: DatosHoy
}

struct ProveedorHoy: TimelineProvider {
    func placeholder(in _: Context) -> EntradaHoy { EntradaHoy(date: .now, datos: .ejemplo) }

    func getSnapshot(in c: Context, completion: @escaping (EntradaHoy) -> Void) {
        if c.isPreview { completion(EntradaHoy(date: .now, datos: .ejemplo)); return }
        Task { completion(EntradaHoy(date: .now, datos: await DatosHoy.traer())) }
    }

    func getTimeline(in _: Context, completion: @escaping (Timeline<EntradaHoy>) -> Void) {
        Task {
            let e = EntradaHoy(date: .now, datos: await DatosHoy.traer())
            // Cada 15 minutos; la app además lo recarga al entrar una venta.
            completion(Timeline(entries: [e], policy: .after(.now.addingTimeInterval(15 * 60))))
        }
    }
}

struct VentasDeHoy: Widget {
    var body: some WidgetConfiguration {
        StaticConfiguration(kind: "VentasDeHoy", provider: ProveedorHoy()) { e in
            VistaVentasDeHoy(datos: e.datos)
                .containerBackground(ColorWidget.fondo, for: .widget)
                .widgetURL(URL(string: "tulocalweb://inicio"))
        }
        .configurationDisplayName("Ventas de hoy")
        .description("Cómo viene el día y los pedidos por preparar.")
        .supportedFamilies([.systemSmall, .systemMedium])
    }
}

struct VistaVentasDeHoy: View {
    var datos: DatosHoy
    @Environment(\.widgetFamily) private var familia

    var body: some View {
        if !datos.conSesion {
            VStack(alignment: .leading, spacing: 6) {
                Text("Tu Local Web").font(.caption.bold()).foregroundStyle(ColorWidget.primario)
                Spacer()
                Text("Entrá a la app para ver tus ventas acá.").font(.footnote).foregroundStyle(ColorWidget.texto2)
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
        } else if familia == .systemMedium {
            HStack(spacing: 14) {
                hoy
                Rectangle().fill(ColorWidget.texto2.opacity(0.18)).frame(width: 1)
                VStack(alignment: .leading, spacing: 6) {
                    Text("Por preparar").font(.caption.weight(.semibold)).foregroundStyle(ColorWidget.texto2)
                    Text("\(datos.porPreparar)").font(.system(size: 30, weight: .bold, design: .rounded)).monospacedDigit()
                        .foregroundStyle(datos.porPreparar > 0 ? ColorWidget.acento : ColorWidget.texto)
                    Spacer(minLength: 0)
                    if let u = datos.ultimo {
                        Link(destination: URL(string: "tulocalweb://pedido/\(u.numero)")!) {
                            VStack(alignment: .leading, spacing: 1) {
                                Text("Último · #\(u.numero)").font(.caption2.weight(.semibold)).foregroundStyle(ColorWidget.texto2)
                                Text(u.cliente).font(.footnote.weight(.semibold)).foregroundStyle(ColorWidget.texto).lineLimit(1)
                                if datos.montos, let t = u.total {
                                    Text(PesosWidget.texto(Double(t))).font(.footnote).monospacedDigit().foregroundStyle(ColorWidget.texto2)
                                }
                            }
                        }
                    }
                }
                .frame(maxWidth: .infinity, alignment: .leading)
            }
        } else {
            hoy
        }
    }

    private var hoy: some View {
        VStack(alignment: .leading, spacing: 4) {
            Text("Hoy").font(.caption.weight(.semibold)).foregroundStyle(ColorWidget.texto2)
            if datos.montos {
                Text(PesosWidget.texto(datos.ventasHoy)).font(.system(size: 24, weight: .bold, design: .rounded))
                    .monospacedDigit().foregroundStyle(ColorWidget.texto).minimumScaleFactor(0.6).lineLimit(1)
                if let d = datos.delta {
                    Text("\(d >= 0 ? "▲" : "▼") \(String(format: "%.1f", abs(d)).replacingOccurrences(of: ".", with: ",")) % vs. ayer")
                        .font(.caption2.weight(.semibold)).foregroundStyle(d >= 0 ? ColorWidget.ok : ColorWidget.error)
                }
            } else {
                Text("\(datos.pedidosHoy)").font(.system(size: 28, weight: .bold, design: .rounded)).monospacedDigit()
                    .foregroundStyle(ColorWidget.texto)
                Text(datos.pedidosHoy == 1 ? "pedido" : "pedidos").font(.caption2).foregroundStyle(ColorWidget.texto2)
            }
            Spacer(minLength: 2)
            Linea(valores: datos.serie).stroke(ColorWidget.acento, style: StrokeStyle(lineWidth: 2.5, lineCap: .round, lineJoin: .round))
                .frame(height: 30)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
    }
}

/// Sparkline de las ventas acumuladas del día.
struct Linea: Shape {
    var valores: [Double]
    func path(in r: CGRect) -> Path {
        var p = Path()
        guard valores.count > 1, let max = valores.max(), max > 0 else {
            p.move(to: CGPoint(x: r.minX, y: r.maxY)); p.addLine(to: CGPoint(x: r.maxX, y: r.maxY)); return p
        }
        for (i, v) in valores.enumerated() {
            let pt = CGPoint(x: r.minX + r.width * CGFloat(i) / CGFloat(valores.count - 1), y: r.maxY - r.height * CGFloat(v / max))
            if i == 0 { p.move(to: pt) } else { p.addLine(to: pt) }
        }
        return p
    }
}

// MARK: - 9a/9b · Live Activity de los pedidos por preparar

struct ActividadPedidos: Widget {
    var body: some WidgetConfiguration {
        ActivityConfiguration(for: PedidosPorPreparar.self) { ctx in
            // Pantalla bloqueada.
            VStack(alignment: .leading, spacing: 10) {
                HStack {
                    Text(ctx.attributes.tienda).font(.caption.weight(.semibold)).foregroundStyle(ColorWidget.texto2)
                    Spacer()
                    Text("Por preparar").font(.caption.weight(.semibold)).foregroundStyle(ColorWidget.texto2)
                }
                HStack(alignment: .firstTextBaseline) {
                    Text("\(ctx.state.porPreparar)").font(.system(size: 34, weight: .bold, design: .rounded))
                        .monospacedDigit().foregroundStyle(ColorWidget.acento)
                    if let c = ctx.state.ultimoCliente, let n = ctx.state.ultimoNumero {
                        Text("Último: #\(n) · \(c)").font(.footnote).foregroundStyle(ColorWidget.texto).lineLimit(1)
                    }
                    Spacer()
                }
                Progreso(hechos: ctx.state.preparadosHoy, pendientes: ctx.state.porPreparar)
                if let n = ctx.state.ultimoNumero {
                    HStack(spacing: 8) {
                        Link(destination: URL(string: "tulocalweb://pedido/\(n)")!) {
                            Text("Ver pedido").font(.footnote.weight(.semibold)).frame(maxWidth: .infinity).padding(.vertical, 8)
                                .background(ColorWidget.texto.opacity(0.08), in: Capsule())
                        }
                        Button(intent: MarcarPreparado(numero: n)) {
                            Text("Preparado").font(.footnote.weight(.semibold)).frame(maxWidth: .infinity).padding(.vertical, 8)
                        }
                        .buttonStyle(.plain)
                        .foregroundStyle(.white)
                        .background(ColorWidget.primario, in: Capsule())
                    }
                    .foregroundStyle(ColorWidget.texto)
                }
            }
            .padding(16)
            .activityBackgroundTint(ColorWidget.fondo)
            .activitySystemActionForegroundColor(ColorWidget.texto)
        } dynamicIsland: { ctx in
            DynamicIsland {
                DynamicIslandExpandedRegion(.leading) {
                    Label("\(ctx.state.porPreparar)", systemImage: "shippingbox.fill")
                        .font(.title2.bold()).monospacedDigit().foregroundStyle(ColorWidget.acento)
                }
                DynamicIslandExpandedRegion(.trailing) {
                    Text("por preparar").font(.caption).foregroundStyle(.secondary)
                }
                DynamicIslandExpandedRegion(.bottom) {
                    if let n = ctx.state.ultimoNumero {
                        HStack(spacing: 8) {
                            Text("#\(n) · \(ctx.state.ultimoCliente ?? "")").font(.footnote).lineLimit(1)
                            Spacer()
                            Button(intent: MarcarPreparado(numero: n)) { Text("Preparado").font(.footnote.weight(.semibold)) }
                                .tint(ColorWidget.acento)
                        }
                    }
                }
            } compactLeading: {
                Image(systemName: "shippingbox.fill").foregroundStyle(ColorWidget.acento)
            } compactTrailing: {
                Text("\(ctx.state.porPreparar)").monospacedDigit().foregroundStyle(ColorWidget.acento)
            } minimal: {
                Text("\(ctx.state.porPreparar)").monospacedDigit().foregroundStyle(ColorWidget.acento)
            }
            .widgetURL(URL(string: "tulocalweb://pedidos"))
        }
    }
}

/// La barra del día: preparados vs. por preparar.
struct Progreso: View {
    var hechos: Int
    var pendientes: Int
    var body: some View {
        GeometryReader { g in
            let total = max(1, hechos + pendientes)
            ZStack(alignment: .leading) {
                Capsule().fill(ColorWidget.texto.opacity(0.1))
                Capsule().fill(ColorWidget.ok).frame(width: g.size.width * CGFloat(hechos) / CGFloat(total))
            }
        }
        .frame(height: 6)
    }
}
