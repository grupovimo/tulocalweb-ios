import Charts
import SwiftUI

/// Estadísticas (Informes del panel web): el período elegido (ventas, pedidos, ticket, clientes nuevos,
/// canales y lo más vendido, de /v1/panel/tablero) y los meses (/v1/panel/informes), con gráficos.
struct Estadisticas: View {
    @Environment(Estado.self) private var estado
    @State private var periodo = "30d"

    struct Informes: Decodable {
        struct Series: Decodable { var ventas: [Int]; var productos: [Int]; var clientes: [Int]; var margen: [Int?] }
        struct Kpis: Decodable {
            var ventas: Int
            var ventasVs: Double?
            var ticket: Int
            var ticketVs: Double?
            var margen: Int?
            var ahorro: Int
            var mesAnterior: String
        }
        var meses: [String]
        var series: Series
        var kpis: Kpis
        var insights: [String]
    }
    struct Datos { var tablero: Tablero; var informes: Informes? }
    struct Mes: Identifiable { var nombre: String; var monto: Int; var id: String { nombre } }

    var body: some View {
        PantallaSeccion(titulo: "Estadísticas", que: "las estadísticas") { () async throws -> Datos in
            async let t: Tablero = estado.api.get("/v1/panel/tablero", query: [URLQueryItem(name: "periodo", value: periodo)])
            async let i: Informes? = try? await estado.api.get("/v1/panel/informes")
            return try await Datos(tablero: t, informes: i)
        } contenido: { d, actualizar in
            Segmentado(opciones: [("hoy", "Hoy"), ("7d", "7 días"), ("30d", "30 días")], elegido: $periodo)
                .onChange(of: periodo) { Task { await actualizar() } }
            let t = d.tablero
            TarjetaHeroe(etiqueta: "Ventas · \(nombrePeriodo)", valor: Formato.pesos(Int(t.ventas?.actual ?? 0)),
                         detalle: variacion(t.ventas), tincho: .grafico)
            if let s = t.serie, s.actual.contains(where: { $0 > 0 }) {
                TarjetaSeccion(titulo: "Cómo vienen las ventas", simbolo: "chart.xyaxis.line") {
                    GraficoVentas(actual: s.actual, anterior: s.anterior, alto: 150)
                }
            }
            GrillaIndicadores(items: [
                IndicadorIcono(simbolo: "bag.fill", titulo: "Pedidos", valor: "\(Int(t.kpis.pedidos.actual))",
                               detalle: variacion(t.kpis.pedidos)),
                IndicadorIcono(simbolo: "receipt", titulo: "Ticket promedio", valor: Formato.pesos(Int(t.kpis.ticket?.actual ?? 0)),
                               detalle: variacion(t.kpis.ticket), tono: (C.avisoFondo, C.avisoTinta)),
                IndicadorIcono(simbolo: "person.crop.circle.badge.plus", titulo: "Clientes nuevos",
                               valor: "\(Int(t.kpis.clientesNuevos.actual))", detalle: variacion(t.kpis.clientesNuevos),
                               tono: (C.okFondo, C.okTinta)),
                IndicadorIcono(simbolo: "percent", titulo: "Margen del mes", valor: d.informes?.kpis.margen.map { "\($0) %" } ?? "—",
                               tono: (C.errorFondo, C.errorTinta)),
            ])
            if let canales = t.canales, !canales.isEmpty {
                TarjetaSeccion(titulo: "Por dónde te compran", simbolo: "point.3.connected.trianglepath.dotted") {
                    DonaMedios(partes: canales.map { .init(nombre: Marcas.canal($0.canal), monto: Int(($0.parte * 1000).rounded())) })
                }
            }
            visitas(t)
            if !t.masVendidos.isEmpty {
                TarjetaSeccion(titulo: "Lo más vendido", simbolo: "flame") {
                    let tope = max(1, t.masVendidos.map(\.unidades).max() ?? 1)
                    VStack(spacing: 12) {
                        ForEach(t.masVendidos) { p in
                            HStack(spacing: 10) {
                                Miniatura(foto: p.foto, tam: 38, radio: 10)
                                VStack(alignment: .leading, spacing: 4) {
                                    HStack {
                                        Text(p.nombre).estilo(.fila).foregroundStyle(C.texto).lineLimit(1)
                                        Spacer(minLength: 6)
                                        Text(Formato.cantidad(p.unidades, "unidad", "unidades")).estilo(.footnoteFuerte)
                                            .monospacedDigit().foregroundStyle(C.texto2)
                                    }
                                    BarraProgreso(valor: Double(p.unidades) / Double(tope), alto: 7)
                                }
                            }
                        }
                    }
                }
            }
            if let i = d.informes, i.meses.count == i.series.ventas.count, i.series.ventas.contains(where: { $0 > 0 }) {
                TarjetaSeccion(titulo: "Mes a mes", simbolo: "calendar") {
                    Chart(zip(i.meses, i.series.ventas).map { Mes(nombre: $0, monto: $1) }) { m in
                        BarMark(x: .value("Mes", m.nombre), y: .value("Ventas", m.monto))
                            .foregroundStyle(LinearGradient(colors: [C.acento, Color(hex: 0xB0441D)], startPoint: .top, endPoint: .bottom))
                            .cornerRadius(5)
                    }
                    .chartYAxis {
                        AxisMarks(position: .leading) { v in
                            AxisGridLine().foregroundStyle(C.linea)
                            AxisValueLabel { if let n = v.as(Int.self) { Text(Formato.compacto(n)).font(.caption2).foregroundStyle(C.texto3) } }
                        }
                    }
                    .frame(height: 170)
                    if i.kpis.ahorro > 0 {
                        Text("Te ahorraste \(Formato.pesos(i.kpis.ahorro)) en comisiones este mes")
                            .estilo(.footnoteFuerte).foregroundStyle(C.okTinta)
                    }
                }
            }
            if let i = d.informes, !i.insights.isEmpty {
                TarjetaSeccion(titulo: "Lo que ve Tincho", simbolo: "sparkles") {
                    VStack(alignment: .leading, spacing: 10) {
                        ForEach(i.insights, id: \.self) { s in
                            HStack(alignment: .firstTextBaseline, spacing: 8) {
                                Image(systemName: "lightbulb.fill").font(.system(size: 12)).foregroundStyle(C.aviso)
                                Text(Formato.sinViudas(s)).estilo(.texto14).foregroundStyle(C.texto)
                                    .fixedSize(horizontal: false, vertical: true)
                            }
                        }
                    }
                }
            }
        }
    }

    /// Las visitas de la tienda online: personas, páginas vistas, conversión y lo más mirado.
    @ViewBuilder
    private func visitas(_ t: Tablero) -> some View {
        if let v = t.visitas {
            if v.desde == nil {
                TarjetaSeccion(titulo: "Visitas de tu tienda", simbolo: "eye") {
                    Text(Formato.sinViudas("Todavía no hay visitas medidas. Compartí el link y acá vas a ver quién entra."))
                        .estilo(.texto14).foregroundStyle(C.texto2)
                    NavigationLink(value: Destino.compartir) { Text("Compartir mi tienda") }.buttonStyle(BotonTonal())
                }
            } else {
                GrillaIndicadores(items: [
                    IndicadorIcono(simbolo: "eye.fill", titulo: "Visitas", valor: Formato.numero(Int(v.visitantes.actual)),
                                   detalle: variacion(v.visitantes)),
                    IndicadorIcono(simbolo: "cart.fill", titulo: "Visitas que compran", valor: Formato.porcentaje(((t.kpis.conversion?.actual ?? 0) * 10).rounded() / 10),
                                   detalle: variacion(t.kpis.conversion), tono: (C.okFondo, C.okTinta)),
                ])
                if !v.masVistos.isEmpty {
                    TarjetaSeccion(titulo: "Lo más mirado", simbolo: "eye") {
                        let tope = max(1, v.masVistos.map(\.visitantes).max() ?? 1)
                        VStack(spacing: 12) {
                            ForEach(v.masVistos) { p in
                                HStack(spacing: 10) {
                                    Miniatura(foto: p.foto, tam: 38, radio: 10)
                                    VStack(alignment: .leading, spacing: 4) {
                                        HStack {
                                            Text(p.nombre).estilo(.fila).foregroundStyle(C.texto).lineLimit(1)
                                            Spacer(minLength: 6)
                                            Text(Formato.cantidad(p.visitantes, "persona", "personas")).estilo(.footnoteFuerte)
                                                .monospacedDigit().foregroundStyle(C.texto2)
                                        }
                                        BarraProgreso(valor: Double(p.visitantes) / Double(tope), alto: 7)
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
    }

    private var nombrePeriodo: String {
        switch periodo { case "hoy": "hoy"; case "7d": "últimos 7 días"; default: "últimos 30 días" }
    }

    private func variacion(_ c: Comparado?) -> String? {
        guard let v = c?.variacion else { return nil }
        let r = Int(v.rounded())
        return r == 0 ? "Igual que el período anterior" : "\(r > 0 ? "▲" : "▼") \(abs(r)) % contra el período anterior"
    }
}
