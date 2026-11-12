import LocalAuthentication
import SwiftUI
import UserNotifications

/// 11h · Cuenta: la persona, seguridad (Face ID, contraseña, cerrar otras sesiones), tema
/// (Claro / Oscuro / Sistema), avisos, ayuda y borrar la cuenta (guía 5.1.1(v) de Apple).
struct Cuenta: View {
    @Environment(Estado.self) private var estado
    @Environment(Navegacion.self) private var nav
    @AppStorage("tema") private var tema = "sistema"
    @State private var faceID = false
    @State private var avisosPermitidos: Bool?
    /// Todavía no se le preguntó (notDetermined): tocar la fila pide el permiso en vez de ir a Ajustes.
    @State private var avisosSinPreguntar = false
    @State private var dosPasos: EstadoDosPasos?
    @State private var verDosPasos = false
    @State private var borrar = false

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
                    fila("Verificación en dos pasos", detalle: dosPasos.map { $0.activo ? "Prendida" : "Apagada" }) {
                        verDosPasos = true
                    }
                    Separador()
                    fila("Cambiar contraseña") { Task { await cambiarClave() } }
                    Separador()
                    fila("Cerrar otras sesiones") { Task { await cerrarOtras() } }
                }
                .tarjeta()
                EtiquetaSeccion(texto: "Avisos").padding(.horizontal, 4).padding(.top, 10)
                VStack(spacing: 0) {
                    fila("Avisos en este iPhone", detalle: avisosPermitidos.map { $0 ? "Prendidos" : "Apagados" }) {
                        if avisosPermitidos == false, !avisosSinPreguntar, let u = URL(string: UIApplication.openNotificationSettingsURLString) {
                            UIApplication.shared.open(u)
                        } else {
                            Task {
                                _ = await estado.pedirPermisoPush()
                                await leerAvisos()
                            }
                        }
                    }
                    Separador()
                    fila("Qué te avisamos") { nav.mas.append(.preferenciasAvisos) }
                }
                .tarjeta()
                EtiquetaSeccion(texto: "Tema").padding(.horizontal, 4).padding(.top, 10)
                Segmentado(opciones: [("claro", "Claro"), ("oscuro", "Oscuro"), ("sistema", "Sistema")], elegido: $tema)
                EtiquetaSeccion(texto: "Ayuda").padding(.horizontal, 4).padding(.top, 10)
                VStack(spacing: 0) {
                    fila("Artículos de ayuda") { nav.mas.append(.ayuda) }
                    Separador()
                    fila("Escribirle a soporte") {
                        if let u = URL(string: "mailto:hola@tulocalweb.com.ar?subject=Ayuda%20desde%20la%20app") { UIApplication.shared.open(u) }
                    }
                }
                .tarjeta()
                if estado.datos != nil {
                    EtiquetaSeccion(texto: "Tu cuenta").padding(.horizontal, 4).padding(.top, 10)
                    Button { borrar = true } label: {
                        HStack {
                            Text("Borrar mi cuenta").estilo(.body16).foregroundStyle(C.errorTinta)
                            Spacer()
                            Chevron(ancho: 8)
                        }
                        .padding(.horizontal, 16).padding(.vertical, 13)
                        .contentShape(Rectangle())
                    }
                    .buttonStyle(.plain)
                    .tarjeta()
                }
            }
            .padding(.top, 10)
            .margenPantalla()
            .padding(.bottom, 24)
        }
        .background(C.fondo.ignoresSafeArea())
        .navigationTitle("Cuenta")
        .navigationBarTitleDisplayMode(.inline)
        .onAppear { faceID = estado.conFaceID }
        .task {
            await leerAvisos()
            dosPasos = try? await estado.api.get("/v1/auth/segundo-factor")
        }
        .sheet(isPresented: $verDosPasos, onDismiss: {
            Task { dosPasos = try? await estado.api.get("/v1/auth/segundo-factor") }
        }) {
            DosPasos(estadoInicial: dosPasos).hojaTLW([.large])
        }
        .sheet(isPresented: $borrar) {
            HojaBorrarCuenta().hojaTLW([.large])
        }
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
        avisosSinPreguntar = a.authorizationStatus == .notDetermined
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

/// Borrar la cuenta (DELETE /v1/auth/cuenta): la cuenta se desactiva al instante, se cierran las tiendas
/// donde es la única dueña y la API borra sus datos personales a los 30 días. Para confirmar, escribe
/// su mail; después, Face ID o el código.
struct HojaBorrarCuenta: View {
    @Environment(Estado.self) private var estado
    @Environment(\.dismiss) private var cerrar
    @State private var confirmacion = ""
    @State private var enviando = false
    @State private var error: String?

    private var email: String { estado.datos?.usuario.email ?? "" }
    /// Las tiendas que se cierran con la cuenta: las que tiene como dueña. Si hay otra persona dueña,
    /// esa tienda sigue (lo resuelve la API).
    private var propias: [String] {
        (estado.acceso?.tiendas ?? []).filter { $0.rol == "duenio" }.map(\.nombre)
    }
    static func coincide(_ escrito: String, _ email: String) -> Bool {
        !email.isEmpty && escrito.trimmingCharacters(in: .whitespacesAndNewlines).lowercased() == email.lowercased()
    }

    var body: some View {
        VStack(spacing: 0) {
            Grabber()
            ScrollView {
                VStack(alignment: .leading, spacing: 14) {
                    Image(systemName: "person.crop.circle.badge.xmark").font(.system(size: 28, weight: .bold))
                        .foregroundStyle(C.errorTinta)
                        .frame(width: 60, height: 60).background(C.errorFondo, in: Circle())
                    Text("Borrar mi cuenta").estilo(.tituloHojaChico).foregroundStyle(C.texto)
                    VStack(alignment: .leading, spacing: 10) {
                        Label(Formato.sinViudas("Tu cuenta se desactiva ahora y tus datos personales se borran a los 30 días."), systemImage: "person.fill.xmark")
                        Label(Formato.sinViudas(cierre), systemImage: "storefront")
                        Label(Formato.sinViudas("En esos 30 días la recuperás escribiendo a hola@tulocalweb.com.ar. Después no se puede deshacer."),
                              systemImage: "exclamationmark.triangle")
                    }
                    .estilo(.texto14).foregroundStyle(C.texto)
                    Text(Formato.sinViudas("Para confirmar, escribí tu email: \(email)"))
                        .estilo(.footnote).foregroundStyle(C.texto2)
                    CampoTexto(titulo: "Tu email", texto: $confirmacion, teclado: .emailAddress)
                    if let error { MensajeError(texto: error) }
                }
                .padding(.horizontal, 20).padding(.top, 8).padding(.bottom, 16)
            }
            .scrollDismissesKeyboard(.interactively)
            VStack(spacing: 4) {
                Button(enviando ? "Un momento…" : "Borrar mi cuenta") { Task { await borrar() } }
                    .buttonStyle(BotonDestructivo())
                    .disabled(!Self.coincide(confirmacion, email) || enviando)
                    .opacity(Self.coincide(confirmacion, email) ? 1 : 0.5)
                Button("Me quedo") { cerrar() }.buttonStyle(BotonTexto(alto: 44))
            }
            .padding(.horizontal, 20).padding(.bottom, 12)
        }
    }

    private var cierre: String {
        switch propias.count {
        case 0: return "Las tiendas donde trabajás siguen andando: sólo dejás de tener acceso."
        case 1: return "Si sos la única persona dueña de \(propias[0]), la tienda se cierra con sus productos y pedidos."
        default: return "Las tiendas donde sos la única persona dueña (\(propias.joined(separator: ", "))) se cierran con sus productos y pedidos."
        }
    }

    private func borrar() async {
        guard Self.coincide(confirmacion, email) else { return }
        guard await estado.confirmarIdentidad("Para borrar tu cuenta") else { return }
        enviando = true
        error = nil
        defer { enviando = false }
        do {
            // La API compara con el mail tal cual: va el de la sesión, no lo que se tipeó.
            try await estado.borrarCuenta(confirmacion: email)
            Haptico.exito()
            cerrar()
        } catch let e as ErrorApi {
            Haptico.error()
            error = e.detalle
        } catch {
            self.error = ErrorApi.desconocido.detalle
        }
    }
}
