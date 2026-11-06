import SwiftUI

/// Hoja del ingreso con la verificación en dos pasos: la contraseña (o Apple, Google) ya estuvo bien y
/// falta el código de seis números de la app autenticadora, o uno de recuperación.
struct CodigoIngreso: View {
    @Environment(Estado.self) private var estado
    @Environment(\.dismiss) private var cerrar
    @State private var codigo = ""
    @State private var error: String?
    @State private var cargando = false
    @FocusState private var foco: Bool

    var body: some View {
        VStack(alignment: .leading, spacing: 18) {
            Grabber().frame(maxWidth: .infinity)
            VStack(alignment: .leading, spacing: 6) {
                Text("Un paso más").estilo(.tituloHoja).foregroundStyle(C.texto)
                Text(Formato.sinViudas("Escribí el código de seis números que muestra tu app de autenticación."))
                    .estilo(.subhead).foregroundStyle(C.texto2)
                    .fixedSize(horizontal: false, vertical: true)
            }
            VStack(alignment: .leading, spacing: 8) {
                CampoCodigo(codigo: $codigo, error: error != nil, foco: $foco) { Task { await entrar() } }
                if let error { MensajeError(texto: error) }
            }
            Button {
                Task { await entrar() }
            } label: {
                if cargando { ProgressView().tint(.white) } else { Text("Entrar") }
            }
            .buttonStyle(BotonPrincipal())
            .disabled(cargando || codigo.filter { !$0.isWhitespace }.count < 6)
            Text(Formato.sinViudas("¿No tenés el teléfono a mano? Usá uno de tus códigos de recuperación."))
                .estilo(.footnote).foregroundStyle(C.texto2)
            Spacer(minLength: 0)
        }
        .padding(.horizontal, 20)
        .onAppear { foco = true }
    }

    private func entrar() async {
        error = nil
        cargando = true
        defer { cargando = false }
        do {
            try await estado.completarSegundoFactor(codigo)
            Haptico.exito()
        } catch let e as ErrorApi where e.estado == 401 {
            Haptico.error()
            estado.avisar(e.detalle, ok: false)
            cerrar()
        } catch let e as ErrorApi {
            Haptico.error()
            error = e.estado == 400 ? "El código no es correcto. Probá con el que muestra ahora." : e.detalle
            codigo = ""
        } catch {
            Haptico.error()
            self.error = ErrorApi.desconocido.detalle
        }
    }
}

/// El campo grande del código (números, o letras para los de recuperación).
struct CampoCodigo: View {
    @Binding var codigo: String
    var error = false
    var foco: FocusState<Bool>.Binding
    var alEnviar: () -> Void

    var body: some View {
        Campo(label: "Código", error: error, enfocado: foco.wrappedValue) {
            TextField("", text: $codigo, prompt: Text("123 456").foregroundStyle(C.texto3))
                .font(.system(size: 22, weight: .semibold, design: .monospaced))
                .foregroundStyle(C.texto)
                .multilineTextAlignment(.center)
                .textContentType(.oneTimeCode)
                .keyboardType(.asciiCapable)
                .textInputAutocapitalization(.never).autocorrectionDisabled()
                .focused(foco)
                .submitLabel(.go)
                .onSubmit(alEnviar)
                .onChange(of: codigo) { _, v in if v.count > 12 { codigo = String(v.prefix(12)) } }
        }
    }
}

/// Cuenta › Verificación en dos pasos: prenderla (con la app autenticadora en este mismo iPhone o
/// escaneando el QR desde otro), guardar los códigos de recuperación, pedir nuevos o apagarla.
struct DosPasos: View {
    @Environment(Estado.self) private var estado
    @Environment(\.dismiss) private var cerrar
    var estadoInicial: EstadoDosPasos?

    private enum Paso: Equatable { case inicio, alta, codigos, apagar, nuevos }
    @State private var paso: Paso = .inicio
    @State private var activo = false
    @State private var restantes = 0
    @State private var alta: Alta?
    @State private var codigo = ""
    @State private var codigos: [String] = []
    @State private var error: String?
    @State private var cargando = false
    @FocusState private var foco: Bool

    private struct Alta: Decodable, Equatable { var secreto: String; var uri: String }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 18) {
                Grabber().frame(maxWidth: .infinity)
                switch paso {
                case .inicio: inicio
                case .alta: escanear
                case .codigos: recuperacion
                case .apagar, .nuevos: pedirCodigo
                }
            }
            .padding(.horizontal, 20)
            .padding(.bottom, 24)
        }
        .scrollBounceBehavior(.basedOnSize)
        .onAppear {
            activo = estadoInicial?.activo ?? false
            restantes = estadoInicial?.codigosRestantes ?? 0
        }
        .animation(.easeInOut(duration: 0.2), value: paso)
    }

    private func titulo(_ t: String, _ d: String) -> some View {
        VStack(alignment: .leading, spacing: 6) {
            Text(Formato.sinViudas(t)).estilo(.tituloHoja).foregroundStyle(C.texto)
            Text(Formato.sinViudas(d)).estilo(.subhead).foregroundStyle(C.texto2)
                .fixedSize(horizontal: false, vertical: true)
        }
    }

    @ViewBuilder private var inicio: some View {
        if activo {
            titulo("Verificación en dos pasos prendida",
                   "Cada vez que entres a tu cuenta te vamos a pedir el código de tu app de autenticación. Te quedan \(restantes) códigos de recuperación.")
            Button("Pedir códigos de recuperación nuevos") { empezar(.nuevos) }
                .buttonStyle(BotonSecundario())
            Button("Apagarla") { empezar(.apagar) }
                .estilo(.subheadFuerte).foregroundStyle(C.errorTinta)
                .frame(maxWidth: .infinity)
        } else {
            titulo("Verificación en dos pasos",
                   "Además de la contraseña, al entrar te pedimos un código que cambia cada 30 segundos en tu celular. Si alguien conoce tu contraseña, igual no puede entrar.")
            Button {
                Task { await activar() }
            } label: {
                if cargando { ProgressView().tint(.white) } else { Text("Prenderla") }
            }
            .buttonStyle(BotonPrincipal())
            .disabled(cargando)
            if let error { MensajeError(texto: error) }
        }
    }

    @ViewBuilder private var escanear: some View {
        titulo("Sumala a tu app de autenticación",
               "Usá la app de Contraseñas del iPhone, Google Authenticator o la que prefieras. Después escribí el código que te muestra.")
        if let alta {
            if let u = URL(string: alta.uri) {
                Button("Abrir en mi app de autenticación") { UIApplication.shared.open(u) }
                    .buttonStyle(BotonSecundario())
            }
            if let img = QRTienda.imagen(alta.uri) {
                VStack(spacing: 8) {
                    Image(uiImage: img).interpolation(.none).resizable().scaledToFit()
                        .frame(width: 180, height: 180)
                        .padding(10).background(.white, in: RoundedRectangle(cornerRadius: 14, style: .continuous))
                        .accessibilityLabel("Código QR para tu app de autenticación")
                    Text("O escanealo desde otro teléfono").estilo(.footnote).foregroundStyle(C.texto2)
                }
                .frame(maxWidth: .infinity)
            }
            Button {
                UIPasteboard.general.string = alta.secreto
                estado.avisar("Clave copiada")
            } label: {
                Text("Copiar la clave para cargarla a mano").estilo(.subheadFuerte).foregroundStyle(C.primarioTexto)
            }
            .frame(maxWidth: .infinity)
        }
        campoYBoton("Prenderla") { await conCodigo("confirmar") }
    }

    @ViewBuilder private var recuperacion: some View {
        titulo("Guardá tus códigos de recuperación",
               "Si perdés el celular, cada uno te deja entrar una vez. Es la única vez que los vas a ver.")
        LazyVGrid(columns: [GridItem(.flexible()), GridItem(.flexible())], spacing: 10) {
            ForEach(codigos, id: \.self) { c in
                Text(c).font(.system(size: 16, weight: .medium, design: .monospaced)).foregroundStyle(C.texto)
                    .frame(maxWidth: .infinity).padding(.vertical, 10)
                    .background(C.superficie, in: RoundedRectangle(cornerRadius: 10, style: .continuous))
            }
        }
        .textSelection(.enabled)
        ShareLink(item: "Códigos de recuperación de Tu Local Web\n\n" + codigos.joined(separator: "\n")) {
            Text("Compartir o guardar")
        }
        .buttonStyle(BotonSecundario())
        Button("Ya los guardé") { cerrar() }.buttonStyle(BotonPrincipal())
    }

    @ViewBuilder private var pedirCodigo: some View {
        if paso == .apagar {
            titulo("Apagar la verificación en dos pasos", "Escribí el código de tu app o uno de recuperación.")
        } else {
            titulo("Códigos de recuperación nuevos", "Escribí el código de tu app. Los códigos de antes dejan de servir.")
        }
        campoYBoton(paso == .apagar ? "Apagarla" : "Generarlos") {
            await conCodigo(paso == .apagar ? "desactivar" : "codigos")
        }
        Button("Cancelar") { paso = .inicio }
            .estilo(.subheadFuerte).foregroundStyle(C.texto2).frame(maxWidth: .infinity)
    }

    private func campoYBoton(_ texto: String, accion: @escaping () async -> Void) -> some View {
        VStack(alignment: .leading, spacing: 14) {
            VStack(alignment: .leading, spacing: 8) {
                CampoCodigo(codigo: $codigo, error: error != nil, foco: $foco) { Task { await accion() } }
                if let error { MensajeError(texto: error) }
            }
            Button {
                Task { await accion() }
            } label: {
                if cargando { ProgressView().tint(.white) } else { Text(texto) }
            }
            .buttonStyle(BotonPrincipal())
            .disabled(cargando || codigo.filter { !$0.isWhitespace }.count < 6)
        }
    }

    private func empezar(_ p: Paso) {
        codigo = ""
        error = nil
        paso = p
        foco = true
    }

    private func activar() async {
        error = nil
        cargando = true
        defer { cargando = false }
        do {
            alta = try await estado.api.enviar("POST", "/v1/auth/segundo-factor/activar", [String: String]())
            empezar(.alta)
        } catch let e as ErrorApi {
            error = e.detalle
        } catch {
            self.error = ErrorApi.desconocido.detalle
        }
    }

    private func conCodigo(_ accion: String) async {
        struct Cuerpo: Encodable { var codigo: String }
        struct Codigos: Decodable { var codigos: [String] }
        error = nil
        cargando = true
        defer { cargando = false }
        do {
            if accion == "desactivar" {
                try await estado.api.hacer("POST", "/v1/auth/segundo-factor/desactivar", Cuerpo(codigo: codigo))
                Haptico.exito()
                estado.avisar("Verificación en dos pasos apagada")
                cerrar()
                return
            }
            let r: Codigos = try await estado.api.enviar("POST", "/v1/auth/segundo-factor/\(accion)", Cuerpo(codigo: codigo))
            Haptico.exito()
            activo = true
            codigos = r.codigos
            restantes = r.codigos.count
            paso = .codigos
        } catch let e as ErrorApi {
            Haptico.error()
            error = e.estado == 400 ? "El código no es correcto. Probá con el que muestra ahora." : e.detalle
            codigo = ""
        } catch {
            Haptico.error()
            self.error = ErrorApi.desconocido.detalle
        }
    }
}
