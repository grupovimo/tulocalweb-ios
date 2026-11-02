import SwiftUI

/// 1a · Splash: isotipo 132 centrado y el wordmark a 64 del borde de abajo.
struct Splash: View {
    var body: some View {
        ZStack {
            C.fondo.ignoresSafeArea()
            Image("Marca").resizable().scaledToFit().frame(width: 132, height: 132)
                .accessibilityLabel("Tu Local Web")
            VStack {
                Spacer()
                Image("Wordmark").resizable().scaledToFit().frame(width: 140)
                    .accessibilityHidden(true)
                    .padding(.bottom, 64 - 34)
            }
        }
    }
}

/// 1b · «Entrá a tu tienda»: email y contraseña, Apple y Google (por Clerk). Con la contraseña mal
/// dice cuántos intentos quedan; al quinto, 1c.
struct Ingresar: View {
    var fondoDeSelector = false
    @Environment(Estado.self) private var estado
    @State private var email = ""
    @State private var clave = ""
    @State private var mostrar = false
    @State private var error: String?
    @State private var errorEmail: String?
    @State private var cargando = false
    @State private var bloqueadoHasta: Date?
    @State private var crear = false
    @FocusState private var foco: CampoIngreso?

    enum CampoIngreso { case email, clave }

    private var conClerk: Bool { Clerk.fapiDe(estado.proveedores?.clerkPublicable) != nil }

    var body: some View {
        ZStack {
            C.fondo.ignoresSafeArea()
            if let hasta = bloqueadoHasta {
                Bloqueado(hasta: hasta, email: email) { bloqueadoHasta = nil }
                    .transition(.opacity)
            } else {
                formulario
            }
        }
        .animation(.easeInOut(duration: 0.2), value: bloqueadoHasta)
        .sheet(isPresented: .constant(estado.fase == .eligiendoTienda)) {
            ElegirTienda()
                .presentationDetents([.height(520), .large])
                .presentationCornerRadius(30)
                .presentationBackground(C.elevado)
                .presentationDragIndicator(.hidden)
                .interactiveDismissDisabled()
        }
        .task { await estado.cargarProveedores() }
    }

    private var formulario: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 22) {
                Image("Marca").resizable().frame(width: 52, height: 52).padding(.leading, -4)
                    .accessibilityHidden(true)
                VStack(alignment: .leading, spacing: 6) {
                    Text("Entrá a tu tienda").estilo(.saludo).foregroundStyle(C.texto)
                        .accessibilityAddTraits(.isHeader)
                    Text("Ventas, pedidos y stock, desde el mostrador.").estilo(.body16).foregroundStyle(C.texto2)
                }
                VStack(alignment: .leading, spacing: 14) {
                    VStack(alignment: .leading, spacing: 8) {
                        Campo(label: "Email", error: errorEmail != nil, enfocado: foco == .email) {
                            TextField("", text: $email, prompt: Text("tu@email.com").foregroundStyle(C.texto3))
                                .estilo(.body).foregroundStyle(C.texto)
                                .keyboardType(.emailAddress).textContentType(.username)
                                .textInputAutocapitalization(.never).autocorrectionDisabled()
                                .focused($foco, equals: .email)
                                .submitLabel(.next)
                                .onSubmit { foco = .clave }
                        }
                        if let errorEmail { MensajeError(texto: errorEmail) }
                    }
                    VStack(alignment: .leading, spacing: 8) {
                        Campo(label: "Contraseña", error: error != nil, enfocado: foco == .clave) {
                            Group {
                                if mostrar {
                                    TextField("", text: $clave)
                                } else {
                                    SecureField("", text: $clave)
                                }
                            }
                            .estilo(.body).foregroundStyle(C.texto)
                            .textContentType(.password)
                            .textInputAutocapitalization(.never).autocorrectionDisabled()
                            .focused($foco, equals: .clave)
                            .submitLabel(.go)
                            .onSubmit { Task { await entrar() } }
                            Button(mostrar ? "Ocultar" : "Mostrar") { mostrar.toggle() }
                                .estilo(.subheadFuerte).foregroundStyle(C.texto2)
                        }
                        if let error { MensajeError(texto: error) }
                    }
                }
                Button("Olvidé mi contraseña") { Task { await recuperar() } }
                    .estilo(.subheadFuerte).foregroundStyle(C.primarioTexto)
                    .frame(maxWidth: .infinity, alignment: .trailing)
                    .padding(.top, -8)
                Button {
                    Task { await entrar() }
                } label: {
                    if cargando { ProgressView().tint(.white) } else { Text("Entrar") }
                }
                .buttonStyle(BotonPrincipal())
                .disabled(cargando)
                if conClerk {
                    HStack(spacing: 12) {
                        Rectangle().fill(C.linea2).frame(height: 1)
                        Text("o").estilo(.texto14).foregroundStyle(C.texto2)
                        Rectangle().fill(C.linea2).frame(height: 1)
                    }
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
                                Image("LogoGoogle").resizable().scaledToFit().frame(width: 20, height: 20)
                                    .accessibilityHidden(true)
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
                }
            }
            .padding(.horizontal, 24)
            .padding(.top, 22)
            .padding(.bottom, 24)
        }
        .scrollDismissesKeyboard(.interactively)
        .safeAreaInset(edge: .bottom) {
            Button { crear = true } label: {
                (Text("¿Todavía no tenés tienda? ").foregroundColor(C.texto2)
                    + Text("Creala gratis").foregroundColor(C.primarioTexto).fontWeight(.semibold))
                    .estilo(.texto14)
                    .multilineTextAlignment(.center)
                    .frame(maxWidth: .infinity, minHeight: 44)
            }
            .buttonStyle(.plain)
            .padding(.bottom, 4)
            .background(C.fondo)
        }
        .sheet(isPresented: $crear) { CrearTienda().hojaTLW([.large]) }
    }

    private func entrar() async {
        error = nil
        errorEmail = nil
        let e = email.trimmingCharacters(in: .whitespaces)
        guard e.contains("@") else {
            errorEmail = "Escribí el email con el que entrás a tu tienda."
            Haptico.error()
            foco = .email
            return
        }
        guard !clave.isEmpty else {
            error = "Escribí tu contraseña."
            Haptico.error()
            foco = .clave
            return
        }
        cargando = true
        defer { cargando = false }
        do {
            try await estado.ingresar(email: e, clave: clave)
            Haptico.exito()
        } catch let e as ErrorApi {
            Haptico.error()
            switch (e.estado, e.codigo) {
            case (401, "credenciales_invalidas"):
                error = Ingresar.mensajeClave(restantes: e.restantes)
                clave = ""
                foco = .clave
            case (429, _):
                bloqueadoHasta = Date().addingTimeInterval(TimeInterval(e.reintentarEn ?? 900))
            default:
                error = e.detalle
            }
        } catch {
            Haptico.error()
            self.error = ErrorApi.desconocido.detalle
        }
    }

    /// «La contraseña no coincide. Te quedan 2 intentos antes del bloqueo.» La API no dice si el
    /// email tiene cuenta (para no delatarlo), así que se nombran los dos.
    static func mensajeClave(restantes: Int?) -> String {
        let base = "El email o la contraseña no coinciden."
        switch restantes {
        case .none: return base
        case .some(0): return base + " Si volvés a fallar, pausamos el ingreso 15 minutos."
        case .some(1): return base + " Te queda 1 intento antes del bloqueo."
        case let .some(n): return base + " Te quedan \(n) intentos antes del bloqueo."
        }
    }

    private func recuperar() async {
        let e = email.trimmingCharacters(in: .whitespaces)
        guard e.contains("@") else {
            errorEmail = "Escribí tu email y te mandamos un link para cambiar la contraseña."
            foco = .email
            return
        }
        struct Cuerpo: Encodable { var email: String }
        do {
            try await estado.api.hacer("POST", "/v1/auth/recuperar", Cuerpo(email: e))
            estado.avisar("Si hay una cuenta con ese email, te llega el link en un minuto.")
        } catch let e as ErrorApi {
            estado.avisar(e.detalle, ok: false)
        } catch {}
    }

    private func social(apple: Bool) async {
        error = nil
        cargando = true
        defer { cargando = false }
        do {
            if apple { try await estado.ingresarConApple() } else { try await estado.ingresarConGoogle() }
            Haptico.exito()
        } catch let e as Clerk.ErrorClerk where e.codigo == "cancelado" {
            return
        } catch let e as ErrorApi {
            Haptico.error()
            error = e.detalle
        } catch {
            Haptico.error()
            self.error = "No pudimos completar el ingreso. Probá de nuevo."
        }
    }
}

/// 1c · Ingreso pausado: Tincho preocupado y la cuenta regresiva (cada segundo).
struct Bloqueado: View {
    var hasta: Date
    var email: String
    var volver: () -> Void
    @Environment(Estado.self) private var estado

    var body: some View {
        VStack(spacing: 0) {
            VStack(spacing: 14) {
                Image(Tincho.preocupado.imagen).resizable().scaledToFit().frame(width: 200, height: 200)
                    .accessibilityHidden(true)
                Text("Pausamos el ingreso por 15 minutos").estilo(.tituloBloqueo).foregroundStyle(C.texto)
                    .multilineTextAlignment(.center)
                Text(Formato.sinViudas("Hubo 5 intentos con una contraseña equivocada. Para cuidar tu tienda, esperá o cambiala ahora."))
                    .estilo(.body16).foregroundStyle(C.texto2).multilineTextAlignment(.center).lineSpacing(3)
                TimelineView(.periodic(from: .now, by: 1)) { ctx in
                    let s = max(0, Int(hasta.timeIntervalSince(ctx.date)))
                    Text("Podés volver a probar en \(s / 60):\(String(format: "%02d", s % 60))")
                        .estilo(.subheadFuerte).monospacedDigit().foregroundStyle(C.avisoTinta)
                        .padding(.horizontal, 14).padding(.vertical, 8)
                        .background(C.avisoFondo, in: Capsule())
                        .onChange(of: s == 0) { _, listo in if listo { volver() } }
                }
                .padding(.top, 6)
            }
            .padding(.horizontal, 32)
            .frame(maxHeight: .infinity)
            VStack(spacing: 10) {
                Button("Cambiar contraseña") { Task { await cambiar() } }.buttonStyle(BotonPrincipal())
                Button("Volver", action: volver).buttonStyle(BotonTexto())
            }
            .padding(.horizontal, 24)
            .padding(.bottom, 16)
        }
        .background(C.fondo.ignoresSafeArea())
    }

    private func cambiar() async {
        struct Cuerpo: Encodable { var email: String }
        try? await estado.api.hacer("POST", "/v1/auth/recuperar", Cuerpo(email: email))
        estado.avisar("Te mandamos un link a \(email) para cambiarla.")
    }
}

/// 1d · «¿A qué tienda entrás?»: hoja de altura fija con las tiendas de la cuenta.
struct ElegirTienda: View {
    @Environment(Estado.self) private var estado
    @State private var elegida: String?
    @State private var cargando = false
    var alElegir: (() -> Void)?

    var body: some View {
        let tiendas = estado.acceso?.tiendas ?? []
        VStack(spacing: 0) {
            Grabber()
            VStack(alignment: .leading, spacing: 4) {
                Text("¿A qué tienda entrás?").estilo(.tituloHoja).foregroundStyle(C.texto)
                Text("Tenés \(tiendas.count) tiendas con esta cuenta.").estilo(.subhead).foregroundStyle(C.texto2)
            }
            .frame(maxWidth: .infinity, alignment: .leading)
            .padding(.horizontal, 4).padding(.top, 22).padding(.bottom, 18)
            ScrollView {
                VStack(spacing: 0) {
                    ForEach(Array(tiendas.enumerated()), id: \.element.id) { i, t in
                        if i > 0 { Separador(inset: 76) }
                        filaTienda(t)
                    }
                }
                .tarjeta()
                Text(Formato.sinViudas("Después cambiás de tienda desde Más, tocando el nombre."))
                    .estilo(.footnote).foregroundStyle(C.texto2)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .padding(.horizontal, 4).padding(.top, 12)
            }
            .scrollBounceBehavior(.basedOnSize)
            let nombre = tiendas.first { $0.id == elegidaId }?.nombre ?? ""
            Button {
                Task { await entrar() }
            } label: {
                if cargando { ProgressView().tint(.white) } else { Text("Entrar a \(nombre)").lineLimit(1) }
            }
            .buttonStyle(BotonPrincipal())
            .disabled(cargando)
            .padding(.bottom, 12)
        }
        .padding(.horizontal, 20)
        .onAppear { elegida = elegida ?? estado.acceso?.tiendaId }
    }

    private var elegidaId: String { elegida ?? estado.acceso?.tiendaId ?? "" }

    private func filaTienda(_ t: Acceso.Tienda) -> some View {
        let (fondo, tinta) = coloresAvatar(t.nombre)
        let activa = t.id == elegidaId
        return Button {
            elegida = t.id
            Haptico.seleccion()
        } label: {
            HStack(spacing: 14) {
                AvatarTienda(tienda: t, tam: 46, radio: 13, fondo: fondo, tinta: tinta)
                VStack(alignment: .leading, spacing: 2) {
                    Text(t.nombre).estilo(.headline).foregroundStyle(C.texto)
                    Text("\(t.direccion) · \(nombreDelRol(t.rol))").estilo(.texto14).foregroundStyle(C.texto2).lineLimit(1)
                }
                Spacer(minLength: 0)
                ZStack {
                    if activa {
                        Circle().fill(C.primarioTexto)
                        Image(systemName: "checkmark").font(.system(size: 11, weight: .heavy)).foregroundStyle(.white)
                    } else {
                        Circle().strokeBorder(Color.tema(0x3A1D12, 0xFBF7F2, alfa: 0.2), lineWidth: 1.5)
                    }
                }
                .frame(width: 24, height: 24)
            }
            .padding(.horizontal, 16).padding(.vertical, 14)
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
        .accessibilityAddTraits(activa ? .isSelected : [])
    }

    private func entrar() async {
        cargando = true
        defer { cargando = false }
        do {
            try await estado.elegirTienda(elegidaId)
            if let alElegir { alElegir() } else { estado.seguirDespuesDeElegir() }
        } catch let e as ErrorApi {
            estado.avisar(e.detalle, ok: false)
        } catch {}
    }
}

/// Pantalla de Face ID al volver (1g prendido).
struct Bloqueo: View {
    @Environment(Estado.self) private var estado
    var body: some View {
        VStack(spacing: 22) {
            Spacer()
            Image("Marca").resizable().frame(width: 96, height: 96).accessibilityHidden(true)
            Text("Entrá con Face ID").estilo(.tituloPaso).foregroundStyle(C.texto)
            Spacer()
            Button("Usar Face ID") { Task { await estado.desbloquear() } }
                .buttonStyle(BotonPrincipal())
                .padding(.horizontal, 24).padding(.bottom, 16)
        }
        .frame(maxWidth: .infinity)
        .background(C.fondo.ignoresSafeArea())
        .task { await estado.desbloquear() }
    }
}
