import SwiftUI

/// 6a · Chat con Tincho. Al abrir, cómo viene el día (cifras y pendientes). Cada respuesta trae el
/// texto y, cuando corresponde, gráficos, tablas, listas, botones que llevan a la sección y preguntas
/// para seguir. El motor vive en la web (/panel/hornero, formato «bloques»), que la app llama con su
/// token; las preguntas de todos los días se responden con los datos, sin IA.
struct TinchoChat: View {
    @Environment(Estado.self) private var estado
    @Environment(Navegacion.self) private var nav
    @State private var mensajes: [Mensaje] = []
    @State private var texto = ""
    @State private var escribiendo = false
    @State private var sinConexion = false
    @State private var saludado = false
    @FocusState private var foco: Bool

    struct Mensaje: Identifiable, Equatable {
        let id = UUID()
        var deTincho: Bool
        var texto: String
        var respuesta: RespuestaTincho? = nil
        /// Preguntas que se mandan solas (el resumen al abrir): no se muestran como burbuja.
        var oculto = false
    }

    private let iniciales = ["¿Cómo vengo hoy?", "¿Qué tengo pendiente?", "¿Qué tengo que reponer?", "¿Qué es lo que más se vende?",
                             "Ideas para vender más", "¿Cómo creo un cupón?"]

    var body: some View {
        VStack(spacing: 0) {
            encabezado
            ScrollViewReader { lector in
                ScrollView {
                    LazyVStack(alignment: .leading, spacing: 14) {
                        burbujaTincho(Mensaje(deTincho: true, texto: saludo))
                        ForEach(mensajes.filter { !$0.oculto }) { m in
                            Group {
                                if m.deTincho { burbujaTincho(m) } else { burbujaPersona(m) }
                            }
                            .id(m.id)
                            .transition(.asymmetric(insertion: .move(edge: .bottom).combined(with: .opacity), removal: .opacity))
                        }
                        if escribiendo { escribiendoAhora.id("escribiendo") }
                        if !escribiendo { sugerencias.id("sugerencias") }
                    }
                    .padding(16)
                    .animation(.spring(duration: 0.35), value: mensajes)
                }
                .scrollDismissesKeyboard(.interactively)
                .onChange(of: mensajes) { _, m in
                    withAnimation { lector.scrollTo(m.last(where: { !$0.oculto })?.id, anchor: .top) }
                }
                .onChange(of: escribiendo) { _, e in
                    if e { withAnimation { lector.scrollTo("escribiendo", anchor: .bottom) } }
                }
            }
            entrada
        }
        .background(C.fondo.ignoresSafeArea())
        .toolbar(.hidden, for: .navigationBar)
        .safeAreaPadding(.bottom, M.tabBar)
        .task {
            guard !saludado else { return }
            saludado = true
            if let p = nav.preguntaTincho {
                nav.preguntaTincho = nil
                preguntar(p)
            } else {
                preguntar("¿Cómo vengo hoy?", oculta: true)
            }
        }
        .onChange(of: nav.preguntaTincho) { _, p in
            guard let p else { return }
            nav.preguntaTincho = nil
            if escribiendo {
                Task { while escribiendo { try? await Task.sleep(for: .milliseconds(200)) }; preguntar(p) }
            } else {
                preguntar(p)
            }
        }
    }

    private var saludo: String {
        let nombre = estado.datos?.nombreDePila
        return "¡Hola\(nombre.map { ", \($0)" } ?? "")! Soy Tincho: conozco tus ventas, tu stock y tus clientes al día. Mirá cómo venís:"
    }

    // MARK: Piezas

    private var encabezado: some View {
        HStack(spacing: 10) {
            Image(Tincho.avatar.imagen).resizable().scaledToFill().frame(width: 38, height: 38)
                .background(C.crema).clipShape(Circle())
                .saturation(sinConexion ? 0 : 1)
            VStack(alignment: .leading, spacing: 1) {
                Text("Tincho").estilo(.headline).foregroundStyle(C.texto)
                HStack(spacing: 5) {
                    Circle().fill(sinConexion ? C.errorTinta : C.ok).frame(width: 6, height: 6)
                    Text(sinConexion ? "Sin conexión" : "Conoce tu tienda al día").estilo(.footnote)
                        .foregroundStyle(sinConexion ? C.errorTinta : C.texto2)
                }
            }
            Spacer()
            if mensajes.contains(where: { !$0.oculto }) {
                Button {
                    // Queda el resumen del día (la pregunta oculta y su respuesta).
                    withAnimation { mensajes = Array(mensajes.prefix(2)) }
                } label: {
                    Image(systemName: "square.and.pencil").font(.system(size: 17, weight: .medium)).foregroundStyle(C.primarioTexto)
                        .frame(width: 40, height: 40)
                }
                .accessibilityLabel("Charla nueva")
            }
        }
        .padding(.horizontal, 16).padding(.vertical, 8)
        .overlay(alignment: .bottom) { Rectangle().fill(C.linea).frame(height: 0.5) }
    }

    private var avatar: some View {
        Image(Tincho.avatar.imagen).resizable().scaledToFill().frame(width: 28, height: 28)
            .background(C.crema).clipShape(Circle())
    }

    private var escribiendoAhora: some View {
        HStack(alignment: .bottom, spacing: 8) {
            avatar
            HStack(spacing: 4) {
                ForEach(0..<3) { i in
                    Circle().fill(C.texto3).frame(width: 7, height: 7)
                        .phaseAnimator([0.3, 1]) { c, f in c.opacity(f) } animation: { _ in .easeInOut(duration: 0.5).delay(Double(i) * 0.15) }
                }
            }
            .padding(.horizontal, 16).padding(.vertical, 14)
            .background(C.superficie, in: UnevenRoundedRectangle(topLeadingRadius: 20, bottomLeadingRadius: 6, bottomTrailingRadius: 20, topTrailingRadius: 20, style: .continuous))
            .accessibilityLabel("Tincho está escribiendo")
        }
    }

    @ViewBuilder
    private func burbujaTincho(_ m: Mensaje) -> some View {
        let texto = m.respuesta?.texto ?? m.texto
        VStack(alignment: .leading, spacing: 10) {
            if !texto.isEmpty {
                HStack(alignment: .bottom, spacing: 8) {
                    avatar
                    Text(Self.conFormato(texto)).estilo(.body16).foregroundStyle(C.texto)
                        .textSelection(.enabled)
                        .padding(.horizontal, 14).padding(.vertical, 10)
                        .background(C.superficie, in: UnevenRoundedRectangle(topLeadingRadius: 20, bottomLeadingRadius: 6, bottomTrailingRadius: 20, topTrailingRadius: 20, style: .continuous))
                        .overlay(UnevenRoundedRectangle(topLeadingRadius: 20, bottomLeadingRadius: 6, bottomTrailingRadius: 20, topTrailingRadius: 20, style: .continuous).strokeBorder(C.anillo, lineWidth: 1))
                    Spacer(minLength: 24)
                }
            }
            if let r = m.respuesta {
                VStack(alignment: .leading, spacing: 10) {
                    ForEach(Array((r.bloques ?? []).enumerated()), id: \.offset) { _, b in VistaBloque(bloque: b) }
                    if let acciones = r.acciones, !acciones.isEmpty {
                        Fluir {
                            ForEach(acciones, id: \.self) { a in
                                Button { ir(a.ir) } label: {
                                    HStack(spacing: 6) {
                                        Text(a.texto)
                                        Image(systemName: "arrow.right").font(.system(size: 11, weight: .bold))
                                    }
                                    .estilo(.footnoteFuerte).foregroundStyle(.white)
                                    .padding(.horizontal, 14).frame(height: 36)
                                    .background(C.primario, in: Capsule())
                                }
                                .buttonStyle(.plain)
                            }
                        }
                    }
                }
                .padding(.leading, 36)
            }
        }
    }

    private func burbujaPersona(_ m: Mensaje) -> some View {
        HStack {
            Spacer(minLength: 60)
            Text(m.texto).estilo(.body16).foregroundStyle(.white)
                .padding(.horizontal, 14).padding(.vertical, 10)
                .background(C.primario, in: UnevenRoundedRectangle(topLeadingRadius: 20, bottomLeadingRadius: 20, bottomTrailingRadius: 6, topTrailingRadius: 20, style: .continuous))
        }
    }

    /// Las de la última respuesta; al empezar, las de siempre.
    private var sugerencias: some View {
        let ultimas = mensajes.last(where: { $0.deTincho })?.respuesta?.sugerencias ?? []
        let lista = ultimas.isEmpty ? iniciales : ultimas
        return Fluir {
            ForEach(lista, id: \.self) { s in
                Button { preguntar(s) } label: {
                    Text(s).estilo(.footnoteFuerte).foregroundStyle(C.primarioTinta)
                        .padding(.horizontal, 12).padding(.vertical, 8)
                        .background(C.crema2, in: Capsule())
                        .overlay(Capsule().strokeBorder(C.linea2, lineWidth: 1))
                }
                .buttonStyle(.plain)
            }
        }
        .padding(.leading, 36)
    }

    private var entrada: some View {
        HStack(spacing: 8) {
            TextField("Preguntale a Tincho", text: $texto, axis: .vertical)
                .lineLimit(1...4)
                .focused($foco)
                .estilo(.body16).foregroundStyle(C.texto)
                .padding(.horizontal, 16).padding(.vertical, 11)
                .background(C.superficie, in: RoundedRectangle(cornerRadius: 22, style: .continuous))
                .overlay(RoundedRectangle(cornerRadius: 22, style: .continuous).strokeBorder(C.linea2, lineWidth: 1))
                .submitLabel(.send)
                .onSubmit { preguntar(texto) }
            Button { preguntar(texto) } label: {
                Image(systemName: "arrow.up").font(.system(size: 17, weight: .bold)).foregroundStyle(.white)
                    .frame(width: 44, height: 44)
                    .background(C.primario, in: Circle())
            }
            .buttonStyle(.plain)
            .disabled(texto.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty || escribiendo)
            .opacity(texto.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty || escribiendo ? 0.5 : 1)
            .accessibilityLabel("Enviar")
        }
        .padding(.horizontal, 16).padding(.vertical, 8)
        .background(C.elevado)
    }

    // MARK: Lógica

    /// **negritas** del texto de Tincho (sin tomar el resto como Markdown).
    static func conFormato(_ t: String) -> AttributedString {
        (try? AttributedString(markdown: t, options: .init(interpretedSyntax: .inlineOnlyPreservingWhitespace))) ?? AttributedString(t)
    }

    private func preguntar(_ t: String, oculta: Bool = false) {
        let pregunta = t.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !pregunta.isEmpty, !escribiendo else { return }
        if !oculta { texto = ""; Haptico.seleccion() }
        mensajes.append(Mensaje(deTincho: false, texto: pregunta, oculto: oculta))
        let historial = mensajes.map { (rol: $0.deTincho ? "hornero" : "persona", texto: $0.respuesta?.texto ?? $0.texto) }
        escribiendo = true
        Task {
            do {
                let r = try await Web.tinchoConBloques(historial, token: estado.api.token)
                mensajes.append(Mensaje(deTincho: true, texto: r.texto, respuesta: r))
                sinConexion = false
            } catch {
                sinConexion = true
                if !oculta {
                    mensajes.append(Mensaje(deTincho: true, texto: "No pude responderte ahora. Revisá la conexión y probá de nuevo."))
                }
            }
            escribiendo = false
        }
    }

    /// Los botones de Tincho llevan a la pantalla de la app.
    private func ir(_ destino: String) {
        Haptico.seleccion()
        switch destino {
        case "pedidos": nav.filtroPedidos = nil; nav.tocar(.pedidos)
        case "preparar": nav.filtroPedidos = "Para preparar"; nav.pestana = .pedidos
        case "transferencias": nav.filtroPedidos = "Sin pagar"; nav.pestana = .pedidos
        case "productos": nav.pestana = .productos
        case "sin-stock": nav.segmentoProductos = "out"; nav.pestana = .productos
        case "venta": nav.venta = true
        default:
            let d: Destino? = switch destino {
            case "nuevo-producto": .nuevoProducto(nil)
            case "importar": .importar
            case "clientes": .clientes
            case "cupones": .cupones
            case "carritos": .carritos
            case "compartir": .compartir
            case "estadisticas": .estadisticas
            case "integraciones": .integraciones
            case "mi-tienda": .miTienda
            case "resenas": .resenas
            case "negocio": .negocio
            case "caja": .caja
            case "finanzas": .finanzas
            case "facturacion": .facturacion
            case "equipo": .usuarios
            case "plan": .plan
            case "links": .links
            case "devoluciones": .devoluciones
            default: nil
            }
            if let d { nav.tincho.append(d) }
        }
    }
}
