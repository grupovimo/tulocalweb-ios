import AVFoundation
import LocalAuthentication
import SwiftUI

/// 1e–1g · Los pre-permisos: notificaciones, cámara y Face ID. Cada botón principal dispara el
/// diálogo del sistema; «Ahora no» y «Saltear» siguen sin pedirlo.
struct Permisos: View {
    @Environment(Estado.self) private var estado
    @State private var paso = 0

    private struct Paso {
        var tincho: Tincho
        var tam: CGFloat
        var titulo: String
        var texto: String
        var boton: String
    }

    private var pasos: [Paso] {
        var p = [
            Paso(tincho: .hablando, tam: 210, titulo: "Te aviso cada venta al instante",
                 texto: "Cuando entra un pedido o se acredita un pago, te llega un aviso. Vos elegís cuáles.",
                 boton: "Activar avisos"),
            Paso(tincho: .foto, tam: 240, titulo: "Fotos y códigos, con la cámara",
                 texto: "La usamos para sacarle fotos a tus productos y para escanear códigos de barras en el local. Nada más.",
                 boton: "Permitir la cámara"),
        ]
        if LAContext().canEvaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, error: nil) {
            p.append(Paso(tincho: .seguridad, tam: 240, titulo: "Entrá con Face ID",
                          texto: "Es más rápido que la contraseña. También te lo pedimos para cosas delicadas, como cancelar un pedido.",
                          boton: "Usar Face ID"))
        }
        return p
    }

    var body: some View {
        let ps = pasos
        let p = ps[min(paso, ps.count - 1)]
        VStack(spacing: 0) {
            HStack {
                HStack(spacing: 6) {
                    ForEach(0..<ps.count, id: \.self) { i in
                        Capsule().fill(i == paso ? C.primarioTexto : Color.tema(0x3A1D12, 0xFBF7F2, alfa: 0.18))
                            .frame(width: i == paso ? 22 : 8, height: 5)
                    }
                }
                .accessibilityElement().accessibilityLabel("Paso \(paso + 1) de \(ps.count)")
                Spacer()
                Button("Saltear") { estado.terminarPermisos() }
                    .estilo(.subheadFuerte).foregroundStyle(C.texto2)
            }
            .padding(.horizontal, 24).padding(.vertical, 10)

            ScrollView {
                VStack(spacing: 14) {
                    Image(p.tincho.imagen).resizable().scaledToFit().frame(width: p.tam, height: p.tam)
                        .accessibilityHidden(true)
                    Text(p.titulo).estilo(.tituloPaso).foregroundStyle(C.texto).multilineTextAlignment(.center)
                    Text(Formato.sinViudas(p.texto)).estilo(.body16).foregroundStyle(C.texto2)
                        .multilineTextAlignment(.center).lineSpacing(3)
                        .fixedSize(horizontal: false, vertical: true)
                    if paso == 0 { mockPush.padding(.top, 14) }
                }
                .padding(.horizontal, 28).padding(.top, 8)
                .id(paso)
                .transition(.asymmetric(insertion: .move(edge: .trailing).combined(with: .opacity), removal: .opacity))
            }
            .scrollBounceBehavior(.basedOnSize)

            VStack(spacing: 10) {
                Button(p.boton) { Task { await aceptar(ps.count) } }.buttonStyle(BotonPrincipal())
                Button("Ahora no") { siguiente(ps.count) }.buttonStyle(BotonTexto())
            }
            .padding(.horizontal, 24).padding(.bottom, 16)
        }
        .background(C.fondo.ignoresSafeArea())
        .animation(.easeInOut(duration: 0.25), value: paso)
    }

    /// El push de muestra de 1e.
    private var mockPush: some View {
        HStack(alignment: .top, spacing: 11) {
            Image("Isotipo").resizable().frame(width: 38, height: 38)
                .clipShape(RoundedRectangle(cornerRadius: 9, style: .continuous))
            VStack(alignment: .leading, spacing: 1) {
                HStack {
                    Text("Venta nueva · \(Formato.pesos(24020))").estilo(.subheadFuerte)
                    Spacer()
                    Text("ahora").estilo(.footnote).foregroundStyle(C.texto2)
                }
                Text("Martina compró Yerba Lumbre suave y 2 productos más.").estilo(.subhead)
                    .fixedSize(horizontal: false, vertical: true)
            }
            .foregroundStyle(C.texto)
        }
        .padding(.horizontal, 14).padding(.vertical, 12)
        .background(C.superficie.opacity(0.85), in: RoundedRectangle(cornerRadius: 20, style: .continuous))
        .shadow(color: Color(hex: 0x2A140C, alfa: 0.2), radius: 14, y: 10)
        .accessibilityHidden(true)
    }

    private func aceptar(_ total: Int) async {
        switch paso {
        case 0:
            _ = await estado.pedirPermisoPush()
        case 1:
            _ = await AVCaptureDevice.requestAccess(for: .video)
        default:
            if await estado.confirmarIdentidad("Para entrar a tu tienda con Face ID") { estado.conFaceID = true }
        }
        siguiente(total)
    }

    private func siguiente(_ total: Int) {
        if paso + 1 < total { paso += 1 } else { estado.terminarPermisos() }
    }
}
