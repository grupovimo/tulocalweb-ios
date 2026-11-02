import SwiftUI

// Traer la tienda de otra plataforma (Tiendanube, Shopify, WooCommerce, Empretienda): productos,
// clientes, pedidos y redirecciones de los links viejos. La usan el onboarding y Más.

struct Importacion: Decodable, Equatable {
    struct Progreso: Decodable, Equatable {
        var productos: Int
        var clientes: Int
        var pedidos: Int
        var redirecciones: Int
        var repetidos: Int
        var recortados: Int
    }
    struct ErrorImportacion: Decodable, Equatable { var mensaje: String? }
    var id: String
    var plataforma: String
    var urlOrigen: String
    var estado: String
    var etapa: String
    var progreso: Progreso
    var errores: [ErrorImportacion]

    var enCurso: Bool { estado == "pendiente" || estado == "corriendo" }
}

enum PlataformaOrigen: String, CaseIterable, Identifiable {
    case tiendanube, shopify, woocommerce, empretienda

    var id: String { rawValue }
    var nombre: String {
        switch self {
        case .tiendanube: "Tiendanube"
        case .shopify: "Shopify"
        case .woocommerce: "WooCommerce"
        case .empretienda: "Empretienda"
        }
    }
    /// Tiendanube se conecta con su cuenta; las demás, con el link de la tienda.
    var conLink: Bool { self != .tiendanube }
    var ejemplo: String {
        switch self {
        case .tiendanube: ""
        case .shopify: "mitienda.myshopify.com"
        case .woocommerce: "www.mitienda.com.ar"
        case .empretienda: "mitienda.empretienda.com.ar"
        }
    }
}

/// Elegir de dónde, arrancar y ver el avance en vivo. `alTerminar` avisa cuando arrancó (o terminó).
struct TraerTienda: View {
    var compacto = false
    @Environment(Estado.self) private var estado
    @State private var plataforma: PlataformaOrigen?
    @State private var link = ""
    @State private var actual: Importacion?
    @State private var arrancando = false
    @State private var error: String?

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            if let a = actual, a.enCurso || a.estado == "lista" {
                progreso(a)
            } else {
                LazyVGrid(columns: [GridItem(.flexible(), spacing: 10), GridItem(.flexible(), spacing: 10)], spacing: 10) {
                    ForEach(PlataformaOrigen.allCases) { p in
                        Button {
                            Haptico.seleccion()
                            plataforma = p
                            error = nil
                        } label: {
                            HStack(spacing: 10) {
                                MarcaPlataforma(plataforma: p)
                                Text(p.nombre).estilo(.fila).foregroundStyle(C.texto).lineLimit(1).minimumScaleFactor(0.8)
                                Spacer(minLength: 0)
                            }
                            .padding(12)
                            .frame(maxWidth: .infinity, minHeight: 56)
                            .tarjeta(radio: 14)
                            .overlay(RoundedRectangle(cornerRadius: 14, style: .continuous)
                                .strokeBorder(plataforma == p ? C.acento : .clear, lineWidth: 1.5))
                        }
                        .buttonStyle(.plain)
                        .accessibilityAddTraits(plataforma == p ? .isSelected : [])
                    }
                }
                if let p = plataforma {
                    if p.conLink {
                        CampoTexto(titulo: "El link de tu tienda en \(p.nombre)", texto: $link, teclado: .URL)
                        Text(Formato.sinViudas("Traemos productos con fotos, variantes y stock. Los links viejos redirigen solos a tu tienda nueva."))
                            .estilo(.footnote).foregroundStyle(C.texto2)
                    } else {
                        Text(Formato.sinViudas("Entrás con tu cuenta de Tiendanube y traemos productos, clientes y pedidos."))
                            .estilo(.footnote).foregroundStyle(C.texto2)
                    }
                    if let error { MensajeError(texto: error) }
                    Button {
                        Task { await arrancar(p) }
                    } label: {
                        if arrancando { ProgressView().tint(.white) } else { Text(p.conLink ? "Traer mis productos" : "Conectar Tiendanube") }
                    }
                    .buttonStyle(BotonPrincipal())
                    .disabled(arrancando || (p.conLink && link.trimmingCharacters(in: .whitespaces).count < 4))
                }
            }
        }
        .task { await seguir() }
    }

    private func progreso(_ a: Importacion) -> some View {
        let lista = a.estado == "lista"
        let nombre = PlataformaOrigen(rawValue: a.plataforma)?.nombre ?? Formato.mayuscula(a.plataforma)
        return VStack(alignment: .leading, spacing: 10) {
            HStack(spacing: 10) {
                if lista {
                    Image(systemName: "checkmark.circle.fill").font(.system(size: 22)).foregroundStyle(C.ok)
                } else {
                    ProgressView().tint(C.acento)
                }
                VStack(alignment: .leading, spacing: 1) {
                    Text(lista ? "Listo: trajimos tu tienda de \(nombre)" : "Trayendo tu tienda de \(nombre)…")
                        .estilo(.fila).foregroundStyle(C.texto)
                    Text(lista ? "Revisá los productos en la pestaña Productos" : "Podés seguir mientras tanto: avanza sola")
                        .estilo(.footnote).foregroundStyle(C.texto2)
                }
            }
            HStack(spacing: 8) {
                cifra(a.progreso.productos, "productos")
                cifra(a.progreso.clientes, "clientes")
                cifra(a.progreso.pedidos, "pedidos")
            }
            if a.progreso.repetidos > 0 {
                Text("\(Formato.cantidad(a.progreso.repetidos, "ya estaba", "ya estaban")) y no se duplicó.")
                    .estilo(.footnote).foregroundStyle(C.texto2)
            }
        }
        .padding(14)
        .tarjeta(radio: 16)
    }

    private func cifra(_ n: Int, _ que: String) -> some View {
        VStack(alignment: .leading, spacing: 1) {
            Text(Formato.numero(n)).estilo(.headline).monospacedDigit().foregroundStyle(C.texto)
            Text(que).estilo(.footnote).foregroundStyle(C.texto2)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(10)
        .background(C.pista, in: RoundedRectangle(cornerRadius: 12, style: .continuous))
    }

    /// Mira la importación en curso cada 3 segundos mientras la pantalla está abierta.
    private func seguir() async {
        while !Task.isCancelled {
            if let a: Importacion = try? await estado.api.get("/v1/panel/importaciones/actual") {
                if actual?.enCurso == true && a.estado == "lista" {
                    Haptico.exito()
                    estado.version += 1
                }
                actual = a
                if !a.enCurso { return }
            } else if actual == nil {
                return
            }
            try? await Task.sleep(for: .seconds(3))
        }
    }

    private func arrancar(_ p: PlataformaOrigen) async {
        error = nil
        arrancando = true
        defer { arrancando = false }
        if p == .tiendanube {
            guard await ConectarCuenta.compartido.importarTiendanube(api: estado.api) else {
                error = "No se completó la conexión con Tiendanube. Probá de nuevo."
                return
            }
        } else {
            var u = link.trimmingCharacters(in: .whitespaces)
            if !u.lowercased().hasPrefix("http") { u = "https://" + u }
            struct Cuerpo: Encodable { var plataforma: String; var url: String }
            do {
                let a: Importacion = try await estado.api.enviar("POST", "/v1/panel/importaciones", Cuerpo(plataforma: p.rawValue, url: u))
                actual = a
            } catch let e as ErrorApi {
                error = e.detalle
                return
            } catch {
                self.error = "No pudimos arrancar. Revisá el link y probá de nuevo."
                return
            }
        }
        Haptico.exito()
        await seguir()
    }
}

/// La inicial de la plataforma en su color (sin logos de terceros dibujados a mano).
struct MarcaPlataforma: View {
    var plataforma: PlataformaOrigen
    var body: some View {
        let (fondo, letra): (Color, String) = switch plataforma {
        case .tiendanube: (Color(hex: 0x2C3357), "T")
        case .shopify: (Color(hex: 0x5E8E3E), "S")
        case .woocommerce: (Color(hex: 0x7F54B3), "W")
        case .empretienda: (Color(hex: 0xE4572E), "E")
        }
        Text(letra).font(.system(size: 15, weight: .bold)).foregroundStyle(.white)
            .frame(width: 30, height: 30)
            .background(fondo, in: RoundedRectangle(cornerRadius: 8, style: .continuous))
            .accessibilityHidden(true)
    }
}

/// Más › Traer mi tienda de otra plataforma.
struct ImportarTienda: View {
    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 14) {
                Text("Traer mi tienda").estilo(.tituloGrande).foregroundStyle(C.texto)
                Text(Formato.sinViudas("¿Vendías en otra plataforma? Traé tus productos, clientes y pedidos en un par de minutos."))
                    .estilo(.body16).foregroundStyle(C.texto2)
                TraerTienda()
            }
            .margenPantalla()
            .padding(.bottom, 32)
        }
        .background(C.fondo.ignoresSafeArea())
        .navigationBarTitleDisplayMode(.inline)
    }
}
