import SwiftUI

/// Los recordatorios del carrito abandonado (§62), como el panel (/panel/carritos-abandonados): hasta
/// tres mails desde que dejó el carrito (por defecto 1 h, 24 h y 72 h), con un cupón único opcional en
/// el segundo o el tercero. Se cortan si compra o se da de baja. Las reglas son las de Go
/// (ventas/recordatorios.go) y las valida la API.
struct RecordatoriosCarrito: Codable, Equatable {
    struct Cupon: Codable, Equatable {
        var descuento: String
        var valor: Int
        var dias: Int
    }
    struct Paso: Codable, Equatable {
        var activo: Bool
        var horas: Int
        var cupon: Cupon?
    }
    var activo: Bool
    var pasos: [Paso]

    static let orden = ["Primero", "Segundo", "Tercero"]
    /// Las demoras que se ofrecen (en horas), de 1 hora a 7 días.
    static let demoras = [1, 2, 3, 6, 12, 24, 36, 48, 72, 96, 120, 168]
    /// Los cupones que se ofrecen en un toque.
    static let cupones: [(String, Cupon?)] = [
        ("Sin cupón", nil),
        ("5 % off", Cupon(descuento: "porcentaje", valor: 5, dias: 3)),
        ("10 % off", Cupon(descuento: "porcentaje", valor: 10, dias: 3)),
        ("15 % off", Cupon(descuento: "porcentaje", valor: 15, dias: 3)),
        ("Envío gratis", Cupon(descuento: "envioGratis", valor: 0, dias: 3)),
    ]

    static func demora(_ h: Int) -> String {
        if h < 24 { return h == 1 ? "1 hora" : "\(h) horas" }
        return h % 24 == 0 ? (h == 24 ? "1 día" : "\(h / 24) días") : "\(h) horas"
    }

    static func cupon(_ c: Cupon?) -> String {
        guard let c else { return "Sin cupón" }
        switch c.descuento {
        case "envioGratis": return "envío gratis"
        case "porcentaje": return "\(c.valor) % off"
        default: return "\(Formato.pesos(c.valor)) off"
        }
    }

    static func resumen(_ p: Paso) -> String {
        "\(demora(p.horas)) después" + (p.cupon.map { _ in " · \(cupon(p.cupon))" } ?? "")
    }
}

/// La hoja «Configurar demoras y cupones».
struct HojaRecordatoriosCarrito: View {
    var inicial: RecordatoriosCarrito
    var guardar: (RecordatoriosCarrito) async -> Bool
    @Environment(\.dismiss) private var cerrar
    @State private var r: RecordatoriosCarrito?
    @State private var guardando = false

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 14) {
                Grabber()
                Text("Recordatorios del carrito").estilo(.tituloHojaChico).foregroundStyle(C.texto)
                Text(Formato.sinViudas("Salen desde que dejó el carrito. Si compra o se da de baja, no le llega ninguno más."))
                    .estilo(.subhead).foregroundStyle(C.texto2)
                if let r {
                    ForEach(r.pasos.indices, id: \.self) { i in paso(i, r.pasos[i]) }
                    Label(Formato.sinViudas("El cupón es único para cada carrito, vale 3 días y se aplica solo con el link del mail."),
                          systemImage: "ticket")
                        .estilo(.footnote).foregroundStyle(C.texto2)
                    Button(guardando ? "Guardando…" : "Guardar") {
                        Task {
                            guardando = true
                            if await guardar(r) { cerrar() }
                            guardando = false
                        }
                    }
                    .buttonStyle(BotonPrincipal()).disabled(guardando)
                }
            }
            .padding(.horizontal, 20)
            .padding(.bottom, 24)
        }
        .onAppear { if r == nil { r = inicial } }
    }

    private func paso(_ i: Int, _ p: RecordatoriosCarrito.Paso) -> some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack {
                Text(RecordatoriosCarrito.orden[i]).estilo(.fila).foregroundStyle(C.texto)
                Spacer()
                Toggle(RecordatoriosCarrito.orden[i], isOn: Binding(get: { r?.pasos[i].activo ?? false }, set: { r?.pasos[i].activo = $0 }))
                    .labelsHidden().tint(C.ok)
            }
            HStack(spacing: 8) {
                Menu {
                    ForEach(RecordatoriosCarrito.demoras, id: \.self) { h in
                        Button("\(RecordatoriosCarrito.demora(h)) después") { r?.pasos[i].horas = h }
                    }
                } label: {
                    etiquetaMenu("\(RecordatoriosCarrito.demora(p.horas)) después")
                }
                if i > 0 {
                    Menu {
                        ForEach(RecordatoriosCarrito.cupones.indices, id: \.self) { k in
                            Button(RecordatoriosCarrito.cupones[k].0) { r?.pasos[i].cupon = RecordatoriosCarrito.cupones[k].1 }
                        }
                    } label: {
                        etiquetaMenu(p.cupon == nil ? "Sin cupón" : RecordatoriosCarrito.cupon(p.cupon))
                    }
                }
            }
        }
        .padding(14).tarjeta()
    }

    private func etiquetaMenu(_ t: String) -> some View {
        HStack(spacing: 6) {
            Text(t).estilo(.texto14).foregroundStyle(C.texto).lineLimit(1)
            Image(systemName: "chevron.up.chevron.down").font(.system(size: 11, weight: .semibold)).foregroundStyle(C.texto2)
        }
        .padding(.horizontal, 12).frame(minHeight: 44)
        .background(C.superficie, in: RoundedRectangle(cornerRadius: 12, style: .continuous))
        .overlay(RoundedRectangle(cornerRadius: 12, style: .continuous).strokeBorder(C.linea2, lineWidth: 1))
    }
}
