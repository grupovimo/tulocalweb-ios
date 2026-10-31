import LocalAuthentication
import SwiftUI
import UserNotifications

/// 11h · Cuenta: la persona, seguridad (Face ID, contraseña, cerrar otras sesiones), tema
/// (Claro / Oscuro / Sistema), avisos y ayuda.
struct Cuenta: View {
    @Environment(Estado.self) private var estado
    @Environment(Navegacion.self) private var nav
    @AppStorage("tema") private var tema = "sistema"
    @State private var faceID = false
    @State private var avisosPermitidos: Bool?

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 8) {
                if let d = estado.datos {
                    HStack(spacing: 12) {
                        AvatarIniciales(nombre: d.usuario.nombre, tam: 52, fondo: C.crema, tinta: C.primarioTinta)
                        VStack(alignment: .leading, spacing: 2) {
                            Text(d.usuario.nombre).estilo(.headline).foregroundStyle(C.texto)
                            Text(d.usuario.email).estilo(.texto14).foregroundStyle(C.texto2)
                        }
                        Spacer(minLength: 0)
                    }
                    .padding(.horizontal, 16).padding(.vertical, 14)
                    .tarjeta()
                }
                EtiquetaSeccion(texto: "Seguridad").padding(.horizontal, 4).padding(.top, 10)
                VStack(spacing: 0) {
                    if LAContext().canEvaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, error: nil) {
                        Interruptor(texto: "Entrar con Face ID", prendido: $faceID)
                            .padding(.horizontal, 16).padding(.vertical, 10)
                            .onChange(of: faceID) { _, v in Task { await cambiarFaceID(v) } }
                        Separador()
                    }
                    fila("Cambiar contraseña") { Task { await cambiarClave() } }
                    Separador()
                    fila("Cerrar otras sesiones") { Task { await cerrarOtras() } }
                }
                .tarjeta()
                EtiquetaSeccion(texto: "Avisos").padding(.horizontal, 4).padding(.top, 10)
                VStack(spacing: 0) {
                    fila("Avisos en este iPhone", detalle: avisosPermitidos.map { $0 ? "Prendidos" : "Apagados" }) {
                        if avisosPermitidos == false, let u = URL(string: UIApplication.openNotificationSettingsURLString) {
                            UIApplication.shared.open(u)
                        } else {
                            Task {
                                _ = await estado.pedirPermisoPush()
                                await leerAvisos()
                            }
                        }
                    }
                    Separador()
                    fila("Qué te avisamos") { nav.web = URL(string: "https://tulocalweb.com.ar/panel/notificaciones") }
                }
                .tarjeta()
                EtiquetaSeccion(texto: "Tema").padding(.horizontal, 4).padding(.top, 10)
                Segmentado(opciones: [("claro", "Claro"), ("oscuro", "Oscuro"), ("sistema", "Sistema")], elegido: $tema)
                EtiquetaSeccion(texto: "Ayuda").padding(.horizontal, 4).padding(.top, 10)
                VStack(spacing: 0) {
                    fila("Artículos de ayuda") { nav.web = URL(string: "https://tulocalweb.com.ar/ayuda") }
                    Separador()
                    fila("Escribirle a soporte") { nav.web = URL(string: "https://tulocalweb.com.ar/contacto") }
                }
                .tarjeta()
            }
            .padding(.top, 10)
            .margenPantalla()
            .padding(.bottom, 24)
        }
        .background(C.fondo.ignoresSafeArea())
        .navigationTitle("Cuenta")
        .navigationBarTitleDisplayMode(.inline)
        .onAppear { faceID = estado.conFaceID }
        .task { await leerAvisos() }
    }

    private func fila(_ texto: String, detalle: String? = nil, accion: @escaping () -> Void) -> some View {
        Button(action: accion) {
            HStack {
                Text(texto).estilo(.body16).foregroundStyle(C.texto)
                Spacer()
                if let detalle { Text(detalle).estilo(.body16).foregroundStyle(C.texto2) }
                Chevron(ancho: 8)
            }
            .padding(.horizontal, 16).padding(.vertical, 13)
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
    }

    private func leerAvisos() async {
        let a = await UNUserNotificationCenter.current().notificationSettings()
        avisosPermitidos = a.authorizationStatus == .authorized || a.authorizationStatus == .provisional
    }

    private func cambiarFaceID(_ v: Bool) async {
        guard v != estado.conFaceID else { return }
        if await estado.confirmarIdentidad(v ? "Para entrar a tu tienda con Face ID" : "Para dejar de usar Face ID") {
            estado.conFaceID = v
        } else {
            faceID = estado.conFaceID
        }
    }

    private func cambiarClave() async {
        guard let email = estado.datos?.usuario.email else { return }
        struct Cuerpo: Encodable { var email: String }
        do {
            try await estado.api.hacer("POST", "/v1/auth/recuperar", Cuerpo(email: email))
            estado.avisar("Te mandamos un link a \(email) para cambiarla.")
        } catch let e as ErrorApi {
            estado.avisar(e.detalle, ok: false)
        } catch {}
    }

    private func cerrarOtras() async {
        struct R: Decodable { var cerradas: Int }
        do {
            let r: R = try await estado.api.enviar("DELETE", "/v1/panel/sesiones", [String: String]())
            Haptico.exito()
            estado.avisar(r.cerradas == 0 ? "No había otras sesiones abiertas" :
                (r.cerradas == 1 ? "Cerramos 1 sesión" : "Cerramos \(r.cerradas) sesiones"))
        } catch let e as ErrorApi {
            estado.avisar(e.detalle, ok: false)
        } catch {}
    }
}

/// La pestaña Tincho. El chat con IA vive todavía en el panel web (Workers AI); hasta que la API lo
/// sirva, la app lo dice y lo abre allá.
struct TinchoChat: View {
    @Environment(Navegacion.self) private var nav
    var body: some View {
        VStack(spacing: 0) {
            HStack(spacing: 10) {
                Image(Tincho.avatar.imagen).resizable().scaledToFill().frame(width: 38, height: 38)
                    .background(C.crema).clipShape(Circle())
                VStack(alignment: .leading, spacing: 1) {
                    Text("Tincho").estilo(.headline).foregroundStyle(C.texto)
                    Text("Conoce tu tienda al día").estilo(.footnote).foregroundStyle(C.texto2)
                }
                Spacer()
            }
            .padding(.horizontal, 16).padding(.vertical, 8)
            .overlay(alignment: .bottom) { Rectangle().fill(C.linea).frame(height: 0.5) }
            Spacer()
            EstadoTincho(tincho: .pensando, titulo: "Tincho llega pronto a la app",
                         texto: "Mientras tanto, preguntale lo que quieras desde el panel web: conoce tus ventas, tu stock y tus clientes.",
                         tam: 210, accion: "Hablar con Tincho en la web", principal: true) {
                nav.web = URL(string: "https://tulocalweb.com.ar/panel/hornero")
            }
            Spacer()
        }
        .background(C.fondo.ignoresSafeArea())
        .toolbar(.hidden, for: .navigationBar)
    }
}
