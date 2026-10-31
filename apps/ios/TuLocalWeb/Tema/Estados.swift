import SwiftUI

// Los estados de cada pantalla (handoff: «Interactions & Behavior · Estados»): cargando con
// esqueletos de forma real (2c), vacío con Tincho y la primera acción (3d), error con Tincho y
// «Reintentar» (12f), sin conexión con el banner y la hora de lo guardado (5a).

/// Lo que puede pasar al traer algo de la API.
enum Carga<T> {
    case cargando
    case listo(T)
    case error(ErrorApi)

    var valor: T? {
        if case let .listo(v) = self { return v }
        return nil
    }
}

/// Tincho con un título, un texto y una o dos acciones.
struct EstadoTincho: View {
    var tincho: Tincho
    var titulo: String
    var texto: String
    var tam: CGFloat = 190
    var secundaria: String?
    var alTocarSecundaria: (() -> Void)?
    var accion: String?
    var principal = false
    var alTocar: (() -> Void)?

    var body: some View {
        VStack(spacing: 12) {
            Image(tincho.imagen).resizable().scaledToFit().frame(width: tam, height: tam)
                .accessibilityHidden(true)
            Text(Formato.sinViudas(titulo)).estilo(.tituloHojaChico).foregroundStyle(C.texto)
                .multilineTextAlignment(.center)
            Text(Formato.sinViudas(texto)).estilo(.subhead).foregroundStyle(C.texto2)
                .multilineTextAlignment(.center).lineSpacing(2)
                .fixedSize(horizontal: false, vertical: true)
            if let accion, let alTocar {
                Group {
                    if principal {
                        Button(accion, action: alTocar).buttonStyle(BotonPrincipal(alto: 48)).fixedSize()
                    } else {
                        Button(accion, action: alTocar).buttonStyle(BotonSecundario())
                    }
                }
                .padding(.top, 8)
            }
            if let secundaria, let alTocarSecundaria {
                Button(secundaria, action: alTocarSecundaria)
                    .estilo(.subheadFuerte).foregroundStyle(C.primarioTexto).padding(.top, 4)
            }
        }
        .padding(.horizontal, 36)
        .frame(maxWidth: .infinity)
    }
}

/// El error de traer una lista (12f): «No pudimos traer los pedidos».
struct EstadoError: View {
    var que: String
    var error: ErrorApi
    var reintentar: () -> Void
    var body: some View {
        if error.sinConexion {
            EstadoTincho(tincho: .desenchufado, titulo: "Sin conexión",
                         texto: "No hay internet. Cuando vuelva, probá de nuevo.",
                         accion: "Reintentar", principal: true, alTocar: reintentar)
        } else {
            EstadoTincho(tincho: .preocupado, titulo: "No pudimos traer \(que)",
                         texto: "Puede ser un problema nuestro. Probá de nuevo en un ratito.",
                         secundaria: "Ver estado del servicio",
                         alTocarSecundaria: { UIApplication.shared.open(URL(string: "https://tulocalweb.com.ar/estado")!) },
                         accion: "Reintentar", principal: true, alTocar: reintentar)
        }
    }
}

/// Banner `warn-bg` de sin conexión (5a).
struct BannerSinConexion: View {
    var guardado: Date
    var body: some View {
        HStack(spacing: 8) {
            Image(systemName: "wifi.slash").font(.system(size: 14, weight: .semibold))
            Text("Sin conexión. Ves lo guardado \(Formato.cuando(guardado)).")
                .estilo(.texto14)
                .frame(maxWidth: .infinity, alignment: .leading)
        }
        .foregroundStyle(C.bannerTinta)
        .padding(.horizontal, 12).padding(.vertical, 9)
        .background(C.avisoFondo, in: RoundedRectangle(cornerRadius: 12, style: .continuous))
        .accessibilityElement(children: .combine)
    }
}

/// Bloque de esqueleto con el brillo horizontal de 1,2 s (respeta Reducir movimiento).
struct Esqueleto: View {
    var ancho: CGFloat? = nil
    var alto: CGFloat
    var radio: CGFloat? = nil
    var suave = false
    @Environment(\.accessibilityReduceMotion) private var reducir
    @State private var fase: CGFloat = -1

    var body: some View {
        let base = suave ? C.esqueleto2 : C.esqueleto1
        RoundedRectangle(cornerRadius: radio ?? alto / 2, style: .continuous)
            .fill(base)
            .overlay {
                if !reducir {
                    GeometryReader { g in
                        LinearGradient(colors: [base, C.elevado.opacity(0.9), base], startPoint: .leading, endPoint: .trailing)
                            .frame(width: g.size.width)
                            .offset(x: fase * g.size.width)
                    }
                    .clipShape(RoundedRectangle(cornerRadius: radio ?? alto / 2, style: .continuous))
                }
            }
            .frame(width: ancho, height: alto)
            .onAppear {
                guard !reducir else { return }
                withAnimation(.linear(duration: 1.2).repeatForever(autoreverses: false)) { fase = 1 }
            }
            .accessibilityHidden(true)
    }
}

/// Filas de lista mientras cargan.
struct EsqueletoFilas: View {
    var filas = 4
    var miniatura: CGFloat = 48
    var body: some View {
        VStack(spacing: 18) {
            ForEach(0..<filas, id: \.self) { i in
                HStack(spacing: 13) {
                    Esqueleto(ancho: miniatura, alto: miniatura, radio: 12)
                    VStack(alignment: .leading, spacing: 7) {
                        GeometryReader { g in
                            VStack(alignment: .leading, spacing: 7) {
                                Esqueleto(ancho: g.size.width * (i.isMultiple(of: 2) ? 0.7 : 0.8), alto: 13)
                                Esqueleto(ancho: g.size.width * (i.isMultiple(of: 2) ? 0.45 : 0.35), alto: 11, suave: true)
                            }
                        }
                        .frame(height: 31)
                    }
                }
            }
        }
        .padding(.horizontal, 16).padding(.vertical, 14)
        .tarjeta()
        .accessibilityLabel("Cargando")
    }
}

// MARK: - Toast

/// Toast inferior oscuro (52 de alto, 4 s) con «Deshacer» cuando aplica.
struct Aviso: Identifiable, Equatable {
    let id = UUID()
    var texto: String
    var ok = true
    var deshacer: (() -> Void)?

    static func == (a: Aviso, b: Aviso) -> Bool { a.id == b.id }
}

struct VistaAviso: View {
    var aviso: Aviso
    var cerrar: () -> Void
    var body: some View {
        HStack(spacing: 10) {
            Image(systemName: aviso.ok ? "checkmark" : "exclamationmark")
                .font(.system(size: 11, weight: .heavy)).foregroundStyle(.white)
                .frame(width: 22, height: 22)
                .background(aviso.ok ? Color(hex: 0x3FA877) : C.error, in: Circle())
            Text(aviso.texto).estilo(.subhead).foregroundStyle(Color(hex: 0xFBF7F2))
                .lineLimit(2)
                .frame(maxWidth: .infinity, alignment: .leading)
            if let d = aviso.deshacer {
                Button("Deshacer") {
                    d()
                    cerrar()
                }
                .font(Estilo(tamano: 15, peso: .bold, relativo: .subheadline).font)
                .foregroundStyle(Color(hex: 0xF3A57E))
            }
        }
        .padding(.horizontal, 16)
        .frame(minHeight: 52)
        .background(Color.tema(0x2A140C, 0x3A2419), in: RoundedRectangle(cornerRadius: 16, style: .continuous))
        .shadow(color: Color(hex: 0x2A140C, alfa: 0.4), radius: 12, y: 10)
        .accessibilityElement(children: .contain)
    }
}
