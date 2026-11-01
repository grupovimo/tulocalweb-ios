import AuthenticationServices
import SwiftUI
import UniformTypeIdentifiers

// Lo que antes pedía la computadora, ahora desde el celular: conectar Mercado Pago, Mercado Libre y
// Correo Argentino, crear cupones, cambiar de plan y cargar el certificado de ARCA.

// MARK: - Conectar una cuenta (OAuth)

/// Abre el login del proveedor; la web devuelve el código a la app (ar.com.tulocalweb.app://…) y la
/// app lo canja con su token.
@MainActor
final class ConectarCuenta: NSObject, ASWebAuthenticationPresentationContextProviding {
    static let compartido = ConectarCuenta()
    private var sesion: ASWebAuthenticationSession?

    func presentationAnchor(for _: ASWebAuthenticationSession) -> ASPresentationAnchor {
        UIApplication.shared.connectedScenes.compactMap { ($0 as? UIWindowScene)?.keyWindow }.first ?? ASPresentationAnchor()
    }

    /// `proveedor`: mercadopago o mercadolibre. Devuelve true si quedó conectada.
    func conectar(_ proveedor: String, api: Api) async -> Bool {
        struct U: Decodable { var url: String }
        guard let u: U = try? await api.enviar("POST", "/v1/panel/integraciones/\(proveedor)/conectar", [String: String]()),
              let url = URL(string: u.url) else { return false }
        let vuelta: URL? = await withCheckedContinuation { cont in
            let s = ASWebAuthenticationSession(url: url, callbackURLScheme: "ar.com.tulocalweb.app") { v, _ in cont.resume(returning: v) }
            s.presentationContextProvider = self
            // Efímera: sin las cookies de Safari. Con la sesión de la web abierta en Safari, la vuelta la
            // canjeaba la web y dejaba a la persona en el panel web, sin volver a la app.
            s.prefersEphemeralWebBrowserSession = true
            sesion = s
            s.start()
        }
        sesion = nil
        // Si la vuelta no llegó a la app pero la cuenta quedó conectada igual, es un éxito.
        struct Conectada: Decodable { var conectada: Bool? }
        func yaConectada() async -> Bool {
            ((try? await api.get("/v1/panel/integraciones/\(proveedor)") as Conectada)?.conectada) ?? false
        }
        guard let vuelta, let c = URLComponents(url: vuelta, resolvingAgainstBaseURL: false),
              let codigo = c.queryItems?.first(where: { $0.name == "code" })?.value,
              let estado = c.queryItems?.first(where: { $0.name == "state" })?.value else { return await yaConectada() }
        struct V: Encodable { var codigo: String; var estado: String }
        do {
            try await api.hacer("POST", "/v1/panel/integraciones/\(proveedor)/vuelta", V(codigo: codigo, estado: estado))
            return true
        } catch { return await yaConectada() }
    }

    /// Abre una página de pago (la suscripción al plan) y vuelve cuando se cierra.
    func abrir(_ url: URL) async {
        await withCheckedContinuation { (cont: CheckedContinuation<Void, Never>) in
            let s = ASWebAuthenticationSession(url: url, callbackURLScheme: "ar.com.tulocalweb.app") { _, _ in cont.resume() }
            s.presentationContextProvider = self
            sesion = s
            s.start()
        }
        sesion = nil
    }
}

// MARK: - Piezas de formulario

struct CampoTexto: View {
    var titulo: String
    @Binding var texto: String
    var teclado: UIKeyboardType = .default
    var seguro = false
    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text(titulo).estilo(.footnoteFuerte).foregroundStyle(C.texto2)
            Group {
                if seguro { SecureField(titulo, text: $texto) } else { TextField(titulo, text: $texto) }
            }
            .keyboardType(teclado)
            .textInputAutocapitalization(teclado == .emailAddress ? .never : .sentences)
            .autocorrectionDisabled(teclado != .default)
            .estilo(.body16).foregroundStyle(C.texto)
            .padding(.horizontal, 14).padding(.vertical, 12)
            .background(C.superficie, in: RoundedRectangle(cornerRadius: 14, style: .continuous))
            .overlay(RoundedRectangle(cornerRadius: 14, style: .continuous).strokeBorder(C.linea2, lineWidth: 1))
        }
        .frame(maxWidth: .infinity, alignment: .leading)
    }
}

/// Hoja con título, contenido con scroll y el botón principal abajo.
struct HojaFormulario<Contenido: View>: View {
    var titulo: String
    var boton: String
    var habilitado: Bool
    var accion: () async -> Bool
    @ViewBuilder var contenido: () -> Contenido
    @Environment(\.dismiss) private var cerrar
    @State private var enviando = false

    var body: some View {
        VStack(spacing: 0) {
            Grabber()
            ScrollView {
                VStack(alignment: .leading, spacing: 12) {
                    Text(titulo).estilo(.tituloHojaChico).foregroundStyle(C.texto)
                    contenido()
                }
                .padding(.horizontal, 20).padding(.top, 8).padding(.bottom, 16)
            }
            .scrollDismissesKeyboard(.interactively)
            Button(enviando ? "Un momento…" : boton) {
                Task {
                    enviando = true
                    if await accion() { cerrar() }
                    enviando = false
                }
            }
            .buttonStyle(BotonPrincipal())
            .disabled(!habilitado || enviando)
            .padding(.horizontal, 20).padding(.bottom, 12)
        }
    }
}

// MARK: - Correo Argentino

struct HojaCorreo: View {
    var listo: () async -> Void
    @Environment(Estado.self) private var estado
    @State private var email = ""
    @State private var clave = ""
    @State private var calle = ""
    @State private var altura = ""
    @State private var cp = ""
    @State private var localidad = ""
    @State private var provincia = "B"
    @State private var peso = "500"

    static let provincias = [("C", "CABA"), ("B", "Buenos Aires"), ("X", "Córdoba"), ("S", "Santa Fe"), ("M", "Mendoza"),
                             ("T", "Tucumán"), ("E", "Entre Ríos"), ("Q", "Neuquén"), ("R", "Río Negro"), ("U", "Chubut"),
                             ("A", "Salta"), ("Y", "Jujuy"), ("H", "Chaco"), ("N", "Misiones"), ("W", "Corrientes"),
                             ("G", "Santiago del Estero"), ("J", "San Juan"), ("D", "San Luis"), ("K", "Catamarca"),
                             ("F", "La Rioja"), ("L", "La Pampa"), ("P", "Formosa"), ("Z", "Santa Cruz"), ("V", "Tierra del Fuego")]

    var body: some View {
        HojaFormulario(titulo: "Conectar Correo Argentino", boton: "Conectar",
                       habilitado: email.contains("@") && !clave.isEmpty && !calle.isEmpty && !altura.isEmpty && cp.count == 4 && !localidad.isEmpty) {
            struct Origen: Encodable { var cp, calle, altura, localidad, provincia: String }
            struct Paquete: Encodable { var peso, alto, ancho, largo: Int }
            struct Alta: Encodable { var email, clave: String; var origen: Origen; var paquete: Paquete }
            do {
                try await estado.api.hacer("POST", "/v1/panel/integraciones/correo", Alta(
                    email: email.trimmingCharacters(in: .whitespaces), clave: clave,
                    origen: Origen(cp: cp, calle: calle, altura: altura, localidad: localidad, provincia: provincia),
                    paquete: Paquete(peso: Int(peso) ?? 500, alto: 10, ancho: 20, largo: 30)))
                estado.avisar("Correo Argentino conectado")
                await listo()
                return true
            } catch let e as ErrorApi {
                estado.avisar(e.detalle, ok: false); return false
            } catch { return false }
        } contenido: {
            Text(Formato.sinViudas("Con tu cuenta de MiCorreo: las etiquetas y el seguimiento salen solos."))
                .estilo(.subhead).foregroundStyle(C.texto2)
            CampoTexto(titulo: "Mail de MiCorreo", texto: $email, teclado: .emailAddress)
            CampoTexto(titulo: "Contraseña de MiCorreo", texto: $clave, seguro: true)
            EtiquetaSeccion(texto: "Desde dónde despachás").padding(.top, 6)
            HStack(spacing: 10) {
                CampoTexto(titulo: "Calle", texto: $calle)
                CampoTexto(titulo: "Altura", texto: $altura, teclado: .numberPad).frame(width: 100)
            }
            HStack(spacing: 10) {
                CampoTexto(titulo: "Código postal", texto: $cp, teclado: .numberPad).frame(width: 130)
                CampoTexto(titulo: "Localidad", texto: $localidad)
            }
            Picker("Provincia", selection: $provincia) {
                ForEach(HojaCorreo.provincias, id: \.0) { Text($0.1).tag($0.0) }
            }
            .tint(C.primarioTexto)
            CampoTexto(titulo: "Peso habitual de un paquete (gramos)", texto: $peso, teclado: .numberPad)
        }
    }
}

// MARK: - Cupón nuevo

struct HojaCuponNuevo: View {
    var listo: () async -> Void
    @Environment(Estado.self) private var estado
    @State private var codigo = ""
    @State private var tipo = "pct"
    @State private var valor = "10"
    @State private var minimo = ""
    @State private var vigencia = "30"

    var body: some View {
        HojaFormulario(titulo: "Cupón nuevo", boton: "Crear cupón",
                       habilitado: codigo.count >= 3 && (tipo == "envio" || (Int(valor) ?? 0) > 0)) {
            struct Opciones: Encodable { var hornero = false; var combinable = false; var unicos = false; var link = false }
            struct Nuevo: Encodable {
                var codigo, tipo: String; var valor: Int; var alcance = "todo"; var quien = "todos"
                var minimo: Int; var limite = 0; var porCliente = "1"; var vigencia: String
                var canales = ["web"]; var opciones = Opciones()
            }
            do {
                try await estado.api.hacer("POST", "/v1/panel/cupones", Nuevo(
                    codigo: codigo.uppercased().filter { $0.isLetter || $0.isNumber }, tipo: tipo,
                    valor: tipo == "envio" ? 0 : (Int(valor) ?? 0), minimo: Int(minimo.filter(\.isNumber)) ?? 0, vigencia: vigencia))
                Haptico.exito()
                estado.avisar("Cupón creado")
                await listo()
                return true
            } catch let e as ErrorApi {
                estado.avisar(e.detalle, ok: false); return false
            } catch { return false }
        } contenido: {
            CampoTexto(titulo: "Código", texto: $codigo)
            Segmentado(opciones: [("pct", "%"), ("fijo", "$"), ("envio", "Envío gratis")], elegido: $tipo)
            if tipo != "envio" {
                CampoTexto(titulo: tipo == "pct" ? "Porcentaje de descuento" : "Monto de descuento", texto: $valor, teclado: .numberPad)
            }
            CampoTexto(titulo: "Compra mínima (opcional)", texto: $minimo, teclado: .numberPad)
            EtiquetaSeccion(texto: "Vigencia").padding(.top, 4)
            Segmentado(opciones: [("7", "7 días"), ("30", "30 días"), ("sin", "Sin fin")], elegido: $vigencia)
        }
    }
}

// MARK: - Plan

struct HojaPlan: View {
    var actual: String
    var listo: () async -> Void
    @Environment(Estado.self) private var estado
    @State private var plan = "emprende"
    @State private var ciclo = "mes"

    static let planes = [("emprende", "Emprende"), ("pro", "Pro"), ("negocio", "Negocio")]

    var body: some View {
        HojaFormulario(titulo: "Cambiar de plan", boton: plan == actual ? "Es tu plan actual" : "Seguir al pago", habilitado: plan != actual) {
            struct S: Encodable { var plan: String; var ciclo: String }
            struct U: Decodable { var url: String }
            do {
                let u: U = try await estado.api.enviar("POST", "/v1/panel/cuenta/suscripcion", S(plan: plan, ciclo: ciclo))
                if let url = URL(string: u.url) { await ConectarCuenta.compartido.abrir(url) }
                await listo()
                return true
            } catch let e as ErrorApi {
                estado.avisar(e.detalle, ok: false); return false
            } catch { return false }
        } contenido: {
            Text(Formato.sinViudas("El cobro va por Mercado Pago, con débito automático. Podés cambiar o cancelar cuando quieras."))
                .estilo(.subhead).foregroundStyle(C.texto2)
            Segmentado(opciones: HojaPlan.planes, elegido: $plan)
            Segmentado(opciones: [("mes", "Por mes"), ("anio", "Por año")], elegido: $ciclo)
            Text("Los precios y lo que incluye cada plan están en tulocalweb.com.ar/precios.")
                .estilo(.footnote).foregroundStyle(C.texto3)
        }
        .onAppear { if HojaPlan.planes.contains(where: { $0.0 == actual }) { plan = actual } }
    }
}

// MARK: - ARCA

/// Alta de ARCA en el celular: los datos, el pedido de certificado para cargar en ARCA y el
/// certificado que ARCA devuelve (desde Archivos o pegado).
struct HojaArca: View {
    var listo: () async -> Void
    @Environment(Estado.self) private var estado
    @Environment(\.dismiss) private var cerrar
    @State private var cuit = ""
    @State private var razon = ""
    @State private var condicion = "monotributo"
    @State private var puntoVenta = "1"
    @State private var pedido: URL?
    @State private var certificado = ""
    @State private var importar = false
    @State private var enviando = false

    var body: some View {
        VStack(spacing: 0) {
            Grabber()
            ScrollView {
                VStack(alignment: .leading, spacing: 12) {
                    Text("Conectar ARCA").estilo(.tituloHojaChico).foregroundStyle(C.texto)
                    paso(1, "Tus datos fiscales")
                    CampoTexto(titulo: "CUIT (11 números)", texto: $cuit, teclado: .numberPad)
                    CampoTexto(titulo: "Razón social", texto: $razon)
                    Segmentado(opciones: [("monotributo", "Monotributo"), ("responsable_inscripto", "Resp. inscripto")], elegido: $condicion)
                    CampoTexto(titulo: "Punto de venta (web services)", texto: $puntoVenta, teclado: .numberPad)
                    if let pedido {
                        paso(2, "Pedí el certificado en ARCA")
                        Text(Formato.sinViudas("Compartí el pedido a tu mail o a Archivos. En arca.gob.ar › Administración de certificados digitales, creá un alias, subí este archivo y bajá el certificado que te da."))
                            .estilo(.texto14).foregroundStyle(C.texto2)
                        ShareLink(item: pedido) { Label("Compartir el pedido (.csr)", systemImage: "square.and.arrow.up") }
                            .buttonStyle(BotonSecundario())
                        paso(3, "Cargá el certificado")
                        Button { importar = true } label: { Label("Elegir el archivo .crt", systemImage: "doc") }
                            .buttonStyle(BotonSecundario())
                        Text("o pegalo acá:").estilo(.footnote).foregroundStyle(C.texto3)
                        TextEditor(text: $certificado).font(.system(size: 11, design: .monospaced)).frame(height: 110)
                            .scrollContentBackground(.hidden).padding(8)
                            .background(C.superficie, in: RoundedRectangle(cornerRadius: 12, style: .continuous))
                            .overlay(RoundedRectangle(cornerRadius: 12, style: .continuous).strokeBorder(C.linea2, lineWidth: 1))
                    }
                }
                .padding(.horizontal, 20).padding(.top, 8).padding(.bottom, 16)
            }
            .scrollDismissesKeyboard(.interactively)
            Button(enviando ? "Un momento…" : (pedido == nil ? "Generar el pedido" : "Guardar el certificado")) {
                Task { await siguiente() }
            }
            .buttonStyle(BotonPrincipal())
            .disabled(enviando || (pedido == nil ? (cuit.filter(\.isNumber).count != 11 || razon.isEmpty) : !certificado.contains("BEGIN CERTIFICATE")))
            .padding(.horizontal, 20).padding(.bottom, 12)
        }
        .fileImporter(isPresented: $importar, allowedContentTypes: [.x509Certificate, .data, .plainText]) { r in
            if case let .success(u) = r {
                let acceso = u.startAccessingSecurityScopedResource()
                defer { if acceso { u.stopAccessingSecurityScopedResource() } }
                if let t = try? String(contentsOf: u, encoding: .utf8) { certificado = t }
                else if let d = try? Data(contentsOf: u) {
                    certificado = "-----BEGIN CERTIFICATE-----\n" + d.base64EncodedString(options: .lineLength64Characters) + "\n-----END CERTIFICATE-----"
                }
            }
        }
    }

    private func paso(_ n: Int, _ t: String) -> some View {
        HStack(spacing: 8) {
            Text("\(n)").font(Estilo(tamano: 13, peso: .bold, relativo: .caption).font)
                .foregroundStyle(C.primarioTinta).frame(width: 24, height: 24).background(C.crema, in: Circle())
            Text(t).estilo(.headline).foregroundStyle(C.texto)
        }
        .padding(.top, 6)
    }

    private func siguiente() async {
        enviando = true
        defer { enviando = false }
        do {
            if pedido == nil {
                struct Alta: Encodable { var cuit, razonSocial, condicion: String; var puntoVenta: Int; var ambiente = "produccion" }
                struct R: Decodable { var pedido: String }
                let r: R = try await estado.api.enviar("POST", "/v1/panel/arca/pedido", Alta(
                    cuit: cuit.filter(\.isNumber), razonSocial: razon, condicion: condicion, puntoVenta: Int(puntoVenta) ?? 1))
                let u = FileManager.default.temporaryDirectory.appendingPathComponent("pedido-arca-\(cuit.filter(\.isNumber)).csr")
                try r.pedido.write(to: u, atomically: true, encoding: .utf8)
                pedido = u
            } else {
                struct C2: Encodable { var certificado: String }
                try await estado.api.hacer("PUT", "/v1/panel/arca/certificado", C2(certificado: certificado.trimmingCharacters(in: .whitespacesAndNewlines)))
                Haptico.exito()
                estado.avisar("ARCA conectada: ya podés facturar")
                await listo()
                cerrar()
            }
        } catch let e as ErrorApi {
            estado.avisar(e.detalle, ok: false)
        } catch {
            estado.avisar("No pudimos guardarlo", ok: false)
        }
    }
}
