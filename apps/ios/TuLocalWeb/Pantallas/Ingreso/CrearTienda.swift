import SwiftUI

/// Crear la cuenta y la tienda desde el celular (POST /v1/auth/registro), como el registro de la web:
/// tu nombre, el de la tienda, email y contraseña (8 o más), con los términos. También con Apple o
/// Google. Después sigue el onboarding.
struct CrearTienda: View {
    @Environment(Estado.self) private var estado
    @Environment(\.dismiss) private var cerrar
    @Environment(\.openURL) private var abrir
    @State private var nombre = ""
    @State private var tienda = ""
    @State private var email = ""
    @State private var clave = ""
    @State private var terminos = false
    @State private var cargando = false
    @State private var error: String?

    private var conClerk: Bool { Clerk.fapiDe(estado.proveedores?.clerkPublicable) != nil }
    private var valido: Bool {
        nombre.trimmingCharacters(in: .whitespaces).count >= 2 && tienda.trimmingCharacters(in: .whitespaces).count >= 2
            && email.contains("@") && clave.count >= 8 && terminos
    }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                Grabber()
                VStack(alignment: .leading, spacing: 6) {
                    Text("Creá tu tienda").estilo(.tituloHojaChico).foregroundStyle(C.texto)
                        .accessibilityAddTraits(.isHeader)
                    Text(Formato.sinViudas("Gratis para arrancar. En dos minutos tenés tu link para vender."))
                        .estilo(.body16).foregroundStyle(C.texto2)
                }
                if conClerk {
                    VStack(spacing: 10) {
                        Button { Task { await social(apple: true) } } label: {
                            HStack(spacing: 8) {
                                Image(systemName: "apple.logo").font(.system(size: 18, weight: .medium))
                                Text("Seguir con Apple")
                            }
                            .estilo(.headline).foregroundStyle(.white)
                            .frame(maxWidth: .infinity, minHeight: 52)
                            .background(.black, in: RoundedRectangle(cornerRadius: 14, style: .continuous))
                        }
                        Button { Task { await social(apple: false) } } label: {
                            HStack(spacing: 10) {
                                Image("LogoGoogle").resizable().scaledToFit().frame(width: 20, height: 20).accessibilityHidden(true)
                                Text("Seguir con Google").foregroundStyle(C.texto)
                            }
                            .estilo(.headline)
                            .frame(maxWidth: .infinity, minHeight: 52)
                            .background(C.superficie, in: RoundedRectangle(cornerRadius: 14, style: .continuous))
                            .overlay(RoundedRectangle(cornerRadius: 14, style: .continuous).strokeBorder(C.linea2, lineWidth: 1))
                        }
                    }
                    .buttonStyle(.plain)
                    .disabled(cargando)
                    HStack(spacing: 12) {
                        Rectangle().fill(C.linea2).frame(height: 1)
                        Text("o con tu email").estilo(.texto14).foregroundStyle(C.texto2).fixedSize()
                        Rectangle().fill(C.linea2).frame(height: 1)
                    }
                }
                CampoTexto(titulo: "Tu nombre", texto: $nombre)
                    .textContentType(.name)
                CampoTexto(titulo: "El nombre de tu tienda", texto: $tienda)
                    .textContentType(.organizationName)
                CampoTexto(titulo: "Email", texto: $email, teclado: .emailAddress)
                    .textContentType(.username)
                VStack(alignment: .leading, spacing: 6) {
                    CampoTexto(titulo: "Contraseña (8 o más)", texto: $clave, seguro: true)
                        .textContentType(.newPassword)
                    if !clave.isEmpty {
                        Text(fuerza.0).estilo(.footnote).foregroundStyle(fuerza.1)
                    }
                }
                Button { terminos.toggle(); Haptico.seleccion() } label: {
                    HStack(alignment: .top, spacing: 10) {
                        Image(systemName: terminos ? "checkmark.square.fill" : "square")
                            .font(.system(size: 20)).foregroundStyle(terminos ? C.acento : C.texto3)
                        (Text("Acepto los ") + Text("Términos y condiciones").foregroundColor(C.primarioTexto).fontWeight(.semibold)
                            + Text(" y la ") + Text("Política de privacidad").foregroundColor(C.primarioTexto).fontWeight(.semibold) + Text("."))
                            .estilo(.texto14).foregroundStyle(C.texto2)
                            .multilineTextAlignment(.leading)
                    }
                }
                .buttonStyle(.plain)
                .accessibilityAddTraits(terminos ? .isSelected : [])
                HStack(spacing: 16) {
                    Button("Leer los términos") { abrir(URL(string: "https://tulocalweb.com.ar/legales/terminos")!) }
                    Button("Leer la privacidad") { abrir(URL(string: "https://tulocalweb.com.ar/legales/privacidad")!) }
                }
                .estilo(.footnoteFuerte).foregroundStyle(C.primarioTexto)
                if let error { MensajeError(texto: error) }
                Button {
                    Task { await crearCuenta() }
                } label: {
                    if cargando { ProgressView().tint(.white) } else { Text("Crear mi tienda") }
                }
                .buttonStyle(BotonPrincipal())
                .disabled(!valido || cargando)
            }
            .padding(.horizontal, 22)
            .padding(.bottom, 24)
        }
        .scrollDismissesKeyboard(.interactively)
        .task { await estado.cargarProveedores() }
    }

    /// Como las barras de la web: corta, débil, buena, fuerte.
    private var fuerza: (String, Color) {
        let tipos = [clave.contains(where: \.isLowercase), clave.contains(where: \.isUppercase),
                     clave.contains(where: \.isNumber), clave.contains(where: { !$0.isLetter && !$0.isNumber })].filter { $0 }.count
        if clave.count < 8 { return ("Le faltan \(8 - clave.count) caracteres", C.errorTinta) }
        if clave.count >= 12 && tipos >= 3 { return ("Fuerte", C.okTinta) }
        return tipos >= 2 ? ("Buena", C.okTinta) : ("Débil: sumale números o símbolos", C.avisoTinta)
    }

    private func crearCuenta() async {
        error = nil
        cargando = true
        defer { cargando = false }
        do {
            try await estado.registrar(nombre: nombre.trimmingCharacters(in: .whitespaces),
                                       tienda: tienda.trimmingCharacters(in: .whitespaces),
                                       email: email.trimmingCharacters(in: .whitespaces).lowercased(), clave: clave)
            Haptico.exito()
            cerrar()
        } catch let e as ErrorApi {
            Haptico.error()
            error = e.estado == 409 ? "Ese email ya tiene cuenta. Entrá con él o recuperá la contraseña." : e.detalle
        } catch {
            Haptico.error()
            self.error = ErrorApi.desconocido.detalle
        }
    }

    private func social(apple: Bool) async {
        error = nil
        cargando = true
        defer { cargando = false }
        do {
            if apple { try await estado.ingresarConApple() } else { try await estado.ingresarConGoogle() }
            cerrar()
        } catch let e as ErrorApi {
            error = e.detalle
        } catch {
            // Cancelado por la persona: nada que decir.
        }
    }
}
