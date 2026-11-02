import SwiftUI

// Las integraciones que estaban sólo en la web: Tincho en Telegram, medición (Google Analytics,
// Google Shopping y Meta) y las IA conectadas por MCP.

// MARK: - Tincho en Telegram

struct TelegramTincho: View {
    @Environment(Estado.self) private var estado
    @Environment(\.openURL) private var abrir
    @State private var vinculando = false

    /// El bot de Tincho (apps/web/src/lib/server/telegram.ts › BOT).
    static let bot = "mostradorhornero_bot"

    struct Datos: Decodable { var vinculado: Bool; var nombre: String?; var desde: Date? }

    var body: some View {
        PantallaSeccion(titulo: "Tincho en Telegram", que: "Telegram") {
            try await estado.api.get("/v1/panel/telegram") as Datos
        } contenido: { d, actualizar in
            TarjetaHeroe(etiqueta: "Telegram", valor: d.vinculado ? "Vinculado" : "Sin vincular",
                         detalle: d.vinculado
                             ? "Te avisa de cada venta y le preguntás lo que quieras\(d.nombre.map { " · \($0)" } ?? "")."
                             : "Tincho te avisa de cada venta y te contesta por chat, sin abrir la app.",
                         tincho: .hablando)
            if d.vinculado {
                Button("Desvincular el chat", role: .destructive) {
                    Task {
                        do {
                            try await estado.api.hacer("DELETE", "/v1/panel/telegram")
                            estado.avisar("Desvinculamos Telegram")
                            await actualizar()
                        } catch { estado.avisar("No pudimos desvincularlo", ok: false) }
                    }
                }
                .buttonStyle(BotonSecundario())
            } else {
                Button {
                    Task { await vincular(actualizar) }
                } label: {
                    if vinculando { ProgressView().tint(.white) } else { Text("Abrir Telegram y vincular") }
                }
                .buttonStyle(BotonPrincipal())
                .disabled(vinculando)
                Text(Formato.sinViudas("Se abre el chat con Tincho: tocá Iniciar y listo. El link sirve 10 minutos."))
                    .estilo(.footnote).foregroundStyle(C.texto2)
            }
        }
    }

    private func vincular(_ actualizar: @escaping () async -> Void) async {
        struct Codigo: Decodable { var codigo: String }
        vinculando = true
        defer { vinculando = false }
        do {
            let c: Codigo = try await estado.api.enviar("POST", "/v1/panel/telegram/codigo", [String: String]())
            if let u = URL(string: "https://t.me/\(TelegramTincho.bot)?start=\(c.codigo)") { abrir(u) }
        } catch let e as ErrorApi {
            estado.avisar(e.detalle, ok: false)
        } catch {
            estado.avisar("No pudimos hablar con Telegram. Probá de nuevo en un rato.", ok: false)
        }
    }
}

// MARK: - Medición: Google Analytics, Google Shopping y Meta

struct MedicionTienda: View {
    @Environment(Estado.self) private var estado
    @State private var ga4 = ""
    @State private var pixel = ""
    @State private var gm = false
    @State private var meta = false
    @State private var cargado = false
    @State private var guardando = false

    struct Medicion: Codable {
        struct Feeds: Codable { var gm: Bool; var meta: Bool }
        var ga4: String?
        var pixel: String?
        var feeds: Feeds
    }

    private var feed: URL? { estado.tienda.map { URL(string: "https://tulocalweb.com.ar/t/\($0.slug)/feed.xml")! } }
    private var gaOk: Bool { ga4.isEmpty || ga4.range(of: "^G-[A-Z0-9]{6,12}$", options: .regularExpression) != nil }
    private var pixelOk: Bool { pixel.isEmpty || pixel.range(of: "^[0-9]{15,16}$", options: .regularExpression) != nil }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 14) {
                Text("Medición y catálogos").estilo(.tituloGrande).foregroundStyle(C.texto)
                if !cargado {
                    EsqueletoFilas(filas: 4, miniatura: 0)
                } else {
                    TarjetaSeccion(titulo: "Google Analytics 4", simbolo: "chart.xyaxis.line") {
                        CampoTexto(titulo: "ID de medición (G-XXXXXXX)", texto: $ga4)
                            .onChange(of: ga4) { _, v in if v != v.uppercased() { ga4 = v.uppercased() } }
                        if !gaOk { MensajeError(texto: "Empieza con G- y sigue con 6 a 12 letras o números.") }
                    }
                    TarjetaSeccion(titulo: "Píxel de Meta", simbolo: "megaphone") {
                        CampoTexto(titulo: "ID del píxel (15 o 16 números)", texto: $pixel, teclado: .numberPad)
                        if !pixelOk { MensajeError(texto: "Son 15 o 16 números.") }
                    }
                    TarjetaSeccion(titulo: "Catálogo para Google y Meta", simbolo: "list.bullet.rectangle") {
                        Text(Formato.sinViudas("Pegá este link como feed en Google Merchant Center y en Meta Commerce Manager: precio, oferta y stock se actualizan solos cada hora."))
                            .estilo(.footnote).foregroundStyle(C.texto2)
                        if let feed {
                            Text(feed.absoluteString).estilo(.texto14).foregroundStyle(C.primarioTexto).textSelection(.enabled)
                            Button("Copiar el link del feed") { Compartir.copiar(feed, estado, aviso: "Copiamos el link del feed") }
                                .buttonStyle(BotonTonal())
                        }
                        Toggle("Lo cargué en Google Merchant Center", isOn: $gm).tint(C.acento)
                        Toggle("Lo cargué en Meta Commerce Manager", isOn: $meta).tint(C.acento)
                    }
                    .estilo(.body16)
                }
            }
            .margenPantalla()
            .padding(.bottom, 32)
        }
        .scrollDismissesKeyboard(.interactively)
        .background(C.fondo.ignoresSafeArea())
        .navigationBarTitleDisplayMode(.inline)
        .barraAccion {
            Button(guardando ? "Guardando…" : "Guardar") { Task { await guardar() } }
                .buttonStyle(BotonPrincipal())
                .disabled(!cargado || guardando || !gaOk || !pixelOk)
        }
        .task { await cargar() }
    }

    private func cargar() async {
        defer { cargado = true }
        if let m: Medicion = try? await estado.api.get("/v1/panel/integraciones/medicion") {
            ga4 = m.ga4 ?? ""
            pixel = m.pixel ?? ""
            gm = m.feeds.gm
            meta = m.feeds.meta
        }
    }

    private func guardar() async {
        guardando = true
        defer { guardando = false }
        let m = Medicion(ga4: ga4.isEmpty ? nil : ga4, pixel: pixel.isEmpty ? nil : pixel, feeds: .init(gm: gm, meta: meta))
        do {
            try await estado.api.hacer("PUT", "/v1/panel/integraciones/medicion", m)
            Haptico.exito()
            estado.avisar("Guardamos la medición")
        } catch let e as ErrorApi {
            estado.avisar(e.detalle, ok: false)
        } catch {
            estado.avisar("No pudimos guardar", ok: false)
        }
    }
}

// MARK: - IA conectadas (MCP)

struct IAConectadas: View {
    @Environment(Estado.self) private var estado

    struct Conexion: Decodable, Identifiable {
        var id: String
        var cliente: String
        var persona: String
        var acciones: Bool
        var usadaEn: Date?
    }
    struct Datos: Decodable { var permisos: [String: Bool]; var conexiones: [Conexion] }

    static let direccion = "https://tulocalweb.com.ar/api/mcp"
    static let permisos: [(id: String, titulo: String, detalle: String)] = [
        ("escribir", "Cargar y editar productos", "Precios, stock y fotos"),
        ("responder", "Responder mensajes", "WhatsApp e Instagram, con tu tono"),
        ("plata", "Reembolsos y descuentos", "Crear cupones de descuento"),
    ]

    var body: some View {
        PantallaSeccion(titulo: "Tu IA", que: "las IA conectadas") {
            try await estado.api.get("/v1/panel/ia") as Datos
        } contenido: { d, actualizar in
            Text(Formato.sinViudas("Manejá Tu Local Web desde Claude, ChatGPT u otra IA: preguntale por tus ventas o pedile que cargue productos."))
                .estilo(.body16).foregroundStyle(C.texto2)
            TarjetaSeccion(titulo: "Cómo se conecta", simbolo: "link") {
                Text("1. En tu IA, agregá un conector o servidor MCP con esta dirección:").estilo(.texto14).foregroundStyle(C.texto)
                Text(IAConectadas.direccion).estilo(.texto14).foregroundStyle(C.primarioTexto).textSelection(.enabled)
                Button("Copiar la dirección") { Compartir.copiar(URL(string: IAConectadas.direccion)!, estado, aviso: "Copiamos la dirección") }
                    .buttonStyle(BotonTonal())
                Text("2. Tocá Conectar: la IA te trae a Tu Local Web para que permitas el acceso.").estilo(.texto14).foregroundStyle(C.texto)
            }
            EtiquetaSeccion(texto: "Qué puede hacer").padding(.horizontal, 4).padding(.top, 8)
            VStack(spacing: 0) {
                FilaPermiso(titulo: "Ver ventas, stock y clientes", detalle: "Sólo lectura, siempre activo", valor: true, fijo: true) { _ in }
                ForEach(IAConectadas.permisos, id: \.id) { p in
                    Separador()
                    FilaPermiso(titulo: p.titulo, detalle: p.detalle, valor: d.permisos[p.id] ?? false, fijo: !estado.puede("cuenta")) { v in
                        struct V: Encodable { var valor: Bool }
                        do {
                            try await estado.api.hacer("PUT", "/v1/panel/ia/permisos/\(p.id)", V(valor: v))
                        } catch {
                            estado.avisar("No pudimos cambiar el permiso", ok: false)
                        }
                        await actualizar()
                    }
                }
            }
            .tarjeta()
            if !d.conexiones.isEmpty {
                EtiquetaSeccion(texto: "Conectadas").padding(.horizontal, 4).padding(.top, 8)
                GrupoFilas(elementos: d.conexiones) { c in
                    HStack {
                        VStack(alignment: .leading, spacing: 2) {
                            Text(c.cliente).estilo(.fila).foregroundStyle(C.texto)
                            Text("\(c.persona)\(c.usadaEn.map { " · usada \(Formato.hace($0))" } ?? "")")
                                .estilo(.footnote).foregroundStyle(C.texto2).lineLimit(1)
                        }
                        Spacer(minLength: 6)
                        Button("Desconectar", role: .destructive) {
                            Task {
                                do {
                                    try await estado.api.hacer("DELETE", "/v1/panel/ia/conexiones/\(c.id)")
                                    estado.avisar("Desconectamos \(c.cliente)")
                                    await actualizar()
                                } catch let e as ErrorApi {
                                    estado.avisar(e.detalle, ok: false)
                                } catch {
                                    estado.avisar("No pudimos desconectarla", ok: false)
                                }
                            }
                        }
                        .buttonStyle(BotonTonal())
                    }
                    .padding(.horizontal, 16).padding(.vertical, 12)
                }
            }
        }
    }
}

private struct FilaPermiso: View {
    var titulo: String
    var detalle: String
    var valor: Bool
    var fijo = false
    var cambiar: (Bool) async -> Void
    @State private var local: Bool?

    var body: some View {
        Toggle(isOn: Binding(get: { local ?? valor }, set: { v in local = v; Task { await cambiar(v); local = nil } })) {
            VStack(alignment: .leading, spacing: 2) {
                Text(titulo).estilo(.fila).foregroundStyle(C.texto)
                Text(detalle).estilo(.footnote).foregroundStyle(C.texto2)
            }
        }
        .tint(C.acento)
        .disabled(fijo)
        .padding(.horizontal, 16).padding(.vertical, 12)
    }
}
