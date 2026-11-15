import SwiftUI

// Newsletter: los contactos de la tienda con su origen (el pie de la tienda, el checkout o una ventana
// de Promociones: «Ventana: Bienvenida 10 %», «Ruleta»), el WhatsApp si lo dejaron y un filtro por
// origen. GET /v1/panel/newsletter (trae las ventanas para nombrar el origen).

struct NewsletterDatos: Decodable {
    struct Suscriptor: Decodable, Identifiable {
        var id: String
        var email: String
        var origen: String
        var nombre: String?
        var whatsapp: String?
        var confirmado: Bool
        var creado: Date
    }
    struct VentanaOrigen: Decodable {
        var id: String
        var nombre: String
        var contenido: String
    }
    var suscriptores: [Suscriptor]
    var confirmados: Int
    var pendientes: Int
    var ventanas: [VentanaOrigen]?

    /// El origen como lo lee la tienda (origenDe en la web).
    func origen(_ o: String) -> String {
        switch o {
        case "tienda": return "Tienda online"
        case "checkout": return "Checkout"
        default:
            guard o.hasPrefix("ventana:") else { return o }
            let id = String(o.dropFirst(8))
            guard let v = ventanas?.first(where: { $0.id == id }), !v.nombre.isEmpty else { return "Ventana" }
            if v.contenido == "ruleta" {
                return v.nombre.localizedCaseInsensitiveContains("ruleta") ? v.nombre : "Ruleta: \(v.nombre)"
            }
            return "Ventana: \(v.nombre)"
        }
    }

    /// Los orígenes que hay, con cuántos contactos trae cada uno (los que más primero).
    var origenes: [(clave: String, nombre: String, cuenta: Int)] {
        var cuentas: [String: Int] = [:]
        for s in suscriptores { cuentas[s.origen, default: 0] += 1 }
        return cuentas.map { (clave: $0.key, nombre: origen($0.key), cuenta: $0.value) }
            .sorted { $0.cuenta != $1.cuenta ? $0.cuenta > $1.cuenta : $0.nombre < $1.nombre }
    }
}

struct Newsletter: View {
    @Environment(Estado.self) private var estado
    @Environment(\.openURL) private var abrir
    @State private var filtro: String?

    var body: some View {
        PantallaSeccion(titulo: "Newsletter", que: "tu newsletter") {
            try await estado.api.get("/v1/panel/newsletter") as NewsletterDatos
        } contenido: { d, _ in
            TarjetaHeroe(etiqueta: "Suscriptos confirmados", valor: "\(d.confirmados)",
                         detalle: d.pendientes > 0 ? Formato.cantidad(d.pendientes, "sin confirmar el mail", "sin confirmar el mail") : nil,
                         tincho: .portapapeles)
            if d.suscriptores.isEmpty {
                EstadoTincho(tincho: .durmiendo, titulo: "Todavía no hay contactos",
                             texto: "Se suman desde el pie de tu tienda, al comprar o con una ventana de Promociones.")
                    .padding(.top, 30)
            } else {
                let origenes = d.origenes
                if origenes.count > 1 {
                    ScrollView(.horizontal, showsIndicators: false) {
                        HStack(spacing: 8) {
                            ChipFiltro(texto: "Todos", cuenta: d.suscriptores.count, activo: filtro == nil) { filtro = nil; Haptico.seleccion() }
                            ForEach(origenes, id: \.clave) { o in
                                ChipFiltro(texto: o.nombre, cuenta: o.cuenta, activo: filtro == o.clave) { filtro = o.clave; Haptico.seleccion() }
                            }
                        }
                    }
                    .scrollClipDisabled()
                }
                let lista = d.suscriptores.filter { filtro == nil || $0.origen == filtro }
                GrupoFilas(elementos: lista, inset: 14) { s in
                    fila(s, d)
                }
            }
        }
    }

    private func fila(_ s: NewsletterDatos.Suscriptor, _ d: NewsletterDatos) -> some View {
        let nombre = s.nombre.flatMap { $0.isEmpty ? nil : $0 }
        return HStack(spacing: 12) {
            VStack(alignment: .leading, spacing: 2) {
                Text(nombre ?? s.email).estilo(.fila).foregroundStyle(C.texto).lineLimit(1)
                if nombre != nil {
                    Text(s.email).estilo(.footnote).foregroundStyle(C.texto2).lineLimit(1)
                }
                Text("\(d.origen(s.origen)) · \(Formato.hace(s.creado))").estilo(.footnote).foregroundStyle(C.texto2).lineLimit(2)
            }
            Spacer(minLength: 6)
            VStack(alignment: .trailing, spacing: 6) {
                if !s.confirmado { ChipSimple(texto: "Sin confirmar", fondo: C.avisoFondo, tinta: C.avisoTinta) }
                if let w = s.whatsapp, let u = urlWhatsApp(w) {
                    Button { abrir(u) } label: { Label("WhatsApp", systemImage: "message") }
                        .buttonStyle(BotonTonal(alto: 32))
                        .accessibilityLabel("Escribirle por WhatsApp al \(w)")
                }
            }
        }
        .padding(.horizontal, 14).padding(.vertical, 12)
    }
}
