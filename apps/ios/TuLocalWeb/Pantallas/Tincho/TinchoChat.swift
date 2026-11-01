import SwiftUI

/// 6a · Chat con Tincho: el saludo con cómo viene el día, sugerencias en chips, burbujas y la respuesta
/// en streaming. El modelo vive en la web (/panel/hornero), que la app llama con su token.
struct TinchoChat: View {
    @Environment(Estado.self) private var estado
    @State private var mensajes: [Mensaje] = []
    @State private var texto = ""
    @State private var escribiendo = false
    @State private var hola: String?
    @State private var sinConexion = false
    @FocusState private var foco: Bool

    struct Mensaje: Identifiable, Equatable {
        let id = UUID()
        var deTincho: Bool
        var texto: String
    }

    private let sugerencias = [
        "¿Cómo vengo hoy?", "¿Qué productos se están por quedar sin stock?",
        "¿Quiénes son mis mejores clientes?", "Ideas para vender más esta semana",
    ]

    var body: some View {
        VStack(spacing: 0) {
            encabezado
            ScrollViewReader { lector in
                ScrollView {
                    LazyVStack(alignment: .leading, spacing: 12) {
                        burbuja(Mensaje(deTincho: true, texto: hola ?? "¡Hola! Soy Tincho. Conozco tus ventas, tu stock y tus clientes. ¿En qué te doy una mano?"))
                        ForEach(mensajes) { m in burbuja(m).id(m.id) }
                        if escribiendo && (mensajes.last?.deTincho != true || mensajes.last?.texto.isEmpty == true) {
                            HStack(spacing: 6) {
                                Image(Tincho.avatar.imagen).resizable().scaledToFill().frame(width: 28, height: 28)
                                    .background(C.crema).clipShape(Circle())
                                ProgressView().tint(C.acento)
                            }
                            .id("escribiendo")
                        }
                    }
                    .padding(16)
                }
                .scrollDismissesKeyboard(.interactively)
                .onChange(of: mensajes) { _, m in
                    withAnimation { lector.scrollTo(m.last?.id, anchor: .bottom) }
                }
            }
            if mensajes.isEmpty {
                ScrollView(.horizontal, showsIndicators: false) {
                    HStack(spacing: 8) {
                        ForEach(sugerencias, id: \.self) { s in
                            Button { enviar(s) } label: {
                                Text(s).estilo(.footnoteFuerte).foregroundStyle(C.primarioTinta)
                                    .padding(.horizontal, 12).padding(.vertical, 8)
                                    .background(C.crema2, in: Capsule())
                                    .overlay(Capsule().strokeBorder(C.linea2, lineWidth: 1))
                            }
                            .buttonStyle(.plain)
                        }
                    }
                    .padding(.horizontal, 16)
                }
                .padding(.bottom, 8)
            }
            entrada
        }
        .background(C.fondo.ignoresSafeArea())
        .toolbar(.hidden, for: .navigationBar)
        .safeAreaPadding(.bottom, M.tabBar)
        .task { await saludar() }
    }

    private var encabezado: some View {
        HStack(spacing: 10) {
            Image(Tincho.avatar.imagen).resizable().scaledToFill().frame(width: 38, height: 38)
                .background(C.crema).clipShape(Circle())
                .saturation(sinConexion ? 0 : 1)
            VStack(alignment: .leading, spacing: 1) {
                Text("Tincho").estilo(.headline).foregroundStyle(C.texto)
                Text(sinConexion ? "Sin conexión" : "Conoce tu tienda al día").estilo(.footnote)
                    .foregroundStyle(sinConexion ? C.errorTinta : C.texto2)
            }
            Spacer()
            if !mensajes.isEmpty {
                Button { mensajes = [] } label: {
                    Image(systemName: "square.and.pencil").font(.system(size: 17, weight: .medium)).foregroundStyle(C.primarioTexto)
                }
                .accessibilityLabel("Charla nueva")
            }
        }
        .padding(.horizontal, 16).padding(.vertical, 8)
        .overlay(alignment: .bottom) { Rectangle().fill(C.linea).frame(height: 0.5) }
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
                .onSubmit { enviar(texto) }
            Button { enviar(texto) } label: {
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

    @ViewBuilder
    private func burbuja(_ m: Mensaje) -> some View {
        if m.deTincho {
            if !m.texto.isEmpty {
                HStack(alignment: .bottom, spacing: 8) {
                    Image(Tincho.avatar.imagen).resizable().scaledToFill().frame(width: 28, height: 28)
                        .background(C.crema).clipShape(Circle())
                    Text(LocalizedStringKey(m.texto)).estilo(.body16).foregroundStyle(C.texto)
                        .textSelection(.enabled)
                        .padding(.horizontal, 14).padding(.vertical, 10)
                        .background(C.superficie, in: UnevenRoundedRectangle(topLeadingRadius: 20, bottomLeadingRadius: 6, bottomTrailingRadius: 20, topTrailingRadius: 20, style: .continuous))
                        .overlay(UnevenRoundedRectangle(topLeadingRadius: 20, bottomLeadingRadius: 6, bottomTrailingRadius: 20, topTrailingRadius: 20, style: .continuous).strokeBorder(C.anillo, lineWidth: 1))
                    Spacer(minLength: 40)
                }
            }
        } else {
            HStack {
                Spacer(minLength: 60)
                Text(m.texto).estilo(.body16).foregroundStyle(.white)
                    .padding(.horizontal, 14).padding(.vertical, 10)
                    .background(C.primario, in: UnevenRoundedRectangle(topLeadingRadius: 20, bottomLeadingRadius: 20, bottomTrailingRadius: 6, topTrailingRadius: 20, style: .continuous))
            }
        }
    }

    private func enviar(_ t: String) {
        let pregunta = t.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !pregunta.isEmpty, !escribiendo else { return }
        texto = ""
        mensajes.append(Mensaje(deTincho: false, texto: pregunta))
        let historial = mensajes.map { (rol: $0.deTincho ? "hornero" : "persona", texto: $0.texto) }
        mensajes.append(Mensaje(deTincho: true, texto: ""))
        escribiendo = true
        Task {
            do {
                for try await pedazo in Web.tincho(historial, token: estado.api.token) {
                    if let i = mensajes.indices.last { mensajes[i].texto += pedazo }
                }
                sinConexion = false
            } catch {
                sinConexion = true
                if let i = mensajes.indices.last, mensajes[i].texto.isEmpty {
                    mensajes[i].texto = "No pude responderte ahora. Revisá la conexión y probá de nuevo."
                }
            }
            escribiendo = false
        }
    }

    private func saludar() async {
        guard hola == nil else { return }
        var r = URLRequest(url: Web.base.appending(path: "panel/hornero"))
        if let t = estado.api.token { r.setValue("Bearer \(t)", forHTTPHeaderField: "Authorization") }
        if let (d, _) = try? await URLSession.shared.data(for: r),
           let v = try? JSONDecoder().decode(ValorJSON.self, from: d), let h = v["hola"]?.texto, !h.isEmpty {
            hola = h
        }
    }
}
