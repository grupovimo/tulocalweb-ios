import SwiftUI

// MARK: - Qué te avisamos

/// Los avisos y por dónde llegan (panel, WhatsApp, mail, celular), como en Configuración › Notificaciones.
struct PreferenciasAvisos: View {
    @Environment(Estado.self) private var estado
    @State private var matriz: [String: Bool] = [:]
    @State private var silencio = false
    @State private var resumen: ValorJSON = .booleano(false)
    @State private var cargado = false

    struct Fila: Identifiable { var id: String; var titulo: String; var detalle: String; var defecto: [Bool] }
    static let canales = [("panel", "Panel"), ("whatsapp", "WhatsApp"), ("email", "Mail"), ("celular", "Celular")]
    static let grupos: [(String, [Fila])] = [
        ("Ventas", [
            Fila(id: "venta", titulo: "Venta nueva", detalle: "Cada pedido pagado", defecto: [true, true, true, true]),
            Fila(id: "pago", titulo: "Pago pendiente o rechazado", detalle: "Transferencias y tarjetas", defecto: [true, true, true, false]),
            Fila(id: "carrito", titulo: "Carrito abandonado recuperado", detalle: "Cuando Tincho cierra la venta", defecto: [true, false, false, true]),
        ]),
        ("Operación", [
            Fila(id: "sinstock", titulo: "Sin stock", detalle: "Una venta lo dejó en 0", defecto: [true, false, true, true]),
            Fila(id: "stock", titulo: "Stock bajo", detalle: "Menos de 5 unidades", defecto: [true, false, true, true]),
            Fila(id: "demora", titulo: "Envío demorado", detalle: "Más de 48 h sin movimiento", defecto: [true, true, false, true]),
            Fila(id: "arca", titulo: "Certificado de ARCA por vencer", detalle: "15 días antes", defecto: [true, true, true, false]),
        ]),
        ("Cuenta y seguridad", [
            Fila(id: "sesion", titulo: "Inicio de sesión nuevo", detalle: "Desde otro dispositivo", defecto: [true, true, true, false]),
            Fila(id: "cobro", titulo: "Cobro del plan", detalle: "Factura de Tu Local Web", defecto: [false, false, true, false]),
            Fila(id: "novedades", titulo: "Novedades de Tu Local Web", detalle: "Una vez por mes", defecto: [false, false, true, false]),
        ]),
    ]

    struct Prefs: Codable { var matriz: [String: Bool]; var silencio: Bool; var resumen: ValorJSON }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 12) {
                Text("Qué te avisamos").estilo(.tituloGrande).foregroundStyle(C.texto)
                if !cargado {
                    EsqueletoFilas(filas: 5, miniatura: 0)
                } else {
                    Toggle(isOn: Binding(get: { silencio }, set: { silencio = $0; Task { await guardar() } })) {
                        VStack(alignment: .leading, spacing: 2) {
                            Text("Modo silencio").estilo(.body16).foregroundStyle(C.texto)
                            Text("Pausa todos los avisos al celular").estilo(.footnote).foregroundStyle(C.texto2)
                        }
                    }
                    .tint(C.ok).padding(.horizontal, 16).padding(.vertical, 10).tarjeta()
                    ForEach(PreferenciasAvisos.grupos, id: \.0) { g in
                        EtiquetaSeccion(texto: g.0).padding(.horizontal, 4).padding(.top, 8)
                        VStack(spacing: 0) {
                            ForEach(Array(g.1.enumerated()), id: \.element.id) { i, f in
                                if i > 0 { Separador() }
                                VStack(alignment: .leading, spacing: 8) {
                                    VStack(alignment: .leading, spacing: 2) {
                                        Text(f.titulo).estilo(.fila).foregroundStyle(C.texto)
                                        Text(f.detalle).estilo(.footnote).foregroundStyle(C.texto2)
                                    }
                                    HStack(spacing: 6) {
                                        ForEach(Array(PreferenciasAvisos.canales.enumerated()), id: \.offset) { j, c in
                                            let on = activo(f, j)
                                            Button {
                                                matriz["\(f.id):\(c.0)"] = !on
                                                Task { await guardar() }
                                            } label: {
                                                Text(c.1).estilo(.footnoteFuerte)
                                                    .foregroundStyle(on ? C.primarioTinta : C.texto3)
                                                    .padding(.horizontal, 10).padding(.vertical, 6)
                                                    .background(on ? C.crema : C.pista, in: Capsule())
                                            }
                                            .buttonStyle(.plain)
                                            .accessibilityAddTraits(on ? .isSelected : [])
                                        }
                                    }
                                }
                                .padding(14)
                            }
                        }
                        .tarjeta()
                    }
                }
            }
            .margenPantalla()
            .padding(.bottom, 24)
        }
        .background(C.fondo.ignoresSafeArea())
        .navigationBarTitleDisplayMode(.inline)
        .task {
            if let p: Prefs = try? await estado.api.get("/v1/panel/notificaciones/preferencias") {
                matriz = p.matriz; silencio = p.silencio; resumen = p.resumen
            }
            cargado = true
        }
    }

    private func activo(_ f: Fila, _ j: Int) -> Bool {
        matriz["\(f.id):\(PreferenciasAvisos.canales[j].0)"] ?? f.defecto[j]
    }

    private func guardar() async {
        do {
            try await estado.api.hacer("PUT", "/v1/panel/notificaciones/preferencias", Prefs(matriz: matriz, silencio: silencio, resumen: resumen))
        } catch {
            estado.avisar("No pudimos guardar los avisos", ok: false)
        }
    }
}

// MARK: - Datos del negocio y de cobro

/// Los datos del negocio y los de cobro por transferencia (alias, CBU/CVU, titular).
struct NegocioCobro: View {
    @Environment(Estado.self) private var estado
    @Environment(\.dismiss) private var cerrar
    @State private var negocio: [String: ValorJSON] = [:]
    @State private var cargado = false
    @State private var guardando = false
    @State private var campos: [String: String] = [:]

    private static let lista: [(String, String, UIKeyboardType)] = [
        ("nombre", "Nombre del negocio", .default), ("whatsapp", "WhatsApp", .phonePad), ("email", "Mail", .emailAddress),
        ("direccion", "Dirección", .default), ("horario", "Horario", .default),
        ("alias", "Alias para transferencias", .default), ("cbu", "CBU o CVU (22 números)", .numberPad), ("titular", "Titular de la cuenta", .default),
    ]

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 12) {
                Text("Datos del negocio").estilo(.tituloGrande).foregroundStyle(C.texto)
                if !cargado {
                    EsqueletoFilas(filas: 6, miniatura: 0)
                } else {
                    ForEach(NegocioCobro.lista, id: \.0) { c in
                        if c.0 == "alias" {
                            EtiquetaSeccion(texto: "Cobro por transferencia").padding(.horizontal, 4).padding(.top, 10)
                        }
                        VStack(alignment: .leading, spacing: 6) {
                            Text(c.1).estilo(.footnoteFuerte).foregroundStyle(C.texto2)
                            TextField(c.1, text: Binding(get: { campos[c.0] ?? "" }, set: { campos[c.0] = $0 }))
                                .keyboardType(c.2)
                                .textInputAutocapitalization(c.0 == "email" || c.0 == "alias" ? .never : .sentences)
                                .autocorrectionDisabled(c.0 == "email" || c.0 == "alias" || c.0 == "cbu")
                                .estilo(.body16).foregroundStyle(C.texto)
                                .padding(.horizontal, 14).padding(.vertical, 12)
                                .background(C.superficie, in: RoundedRectangle(cornerRadius: 14, style: .continuous))
                                .overlay(RoundedRectangle(cornerRadius: 14, style: .continuous).strokeBorder(C.linea2, lineWidth: 1))
                        }
                    }
                    Button(guardando ? "Guardando…" : "Guardar") { Task { await guardar() } }
                        .buttonStyle(BotonPrincipal()).disabled(guardando).padding(.top, 8)
                }
            }
            .margenPantalla()
            .padding(.bottom, 32)
        }
        .scrollDismissesKeyboard(.interactively)
        .background(C.fondo.ignoresSafeArea())
        .navigationBarTitleDisplayMode(.inline)
        .task {
            if let v: ValorJSON = try? await estado.api.get("/v1/panel/configuracion"), case let .objeto(n)? = v["negocio"] {
                negocio = n
                for (k, _, _) in NegocioCobro.lista { campos[k] = n[k]?.texto ?? "" }
            }
            cargado = true
        }
    }

    private func guardar() async {
        var n = negocio
        for (k, _, _) in NegocioCobro.lista {
            var t = (campos[k] ?? "").trimmingCharacters(in: .whitespaces)
            if k == "cbu" { t = t.filter(\.isNumber) }
            n[k] = .texto(t)
        }
        for k in ["razonSocial", "cuit", "condicion"] where n[k] == nil { n[k] = .texto("") }
        guardando = true
        defer { guardando = false }
        do {
            try await estado.api.hacer("PUT", "/v1/panel/configuracion/negocio", ValorJSON.objeto(n))
            negocio = n
            Haptico.exito()
            estado.avisar("Guardamos los datos")
            cerrar()
        } catch let e as ErrorApi {
            estado.avisar(e.detalle, ok: false)
        } catch {
            estado.avisar("No pudimos guardarlos", ok: false)
        }
    }
}

// MARK: - Links de pago

/// Un link de pago por un monto, para mandar por WhatsApp.
struct LinksPago: View {
    var montoInicial: Int? = nil
    @Environment(Estado.self) private var estado
    @State private var monto = ""
    @State private var concepto = ""
    @State private var cuotas = false
    @State private var link: URL?
    @State private var creando = false

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 12) {
                Text("Link de pago").estilo(.tituloGrande).foregroundStyle(C.texto)
                Text(Formato.sinViudas("Se paga con Mercado Pago: tarjeta, dinero en cuenta o efectivo. La plata llega a tu cuenta."))
                    .estilo(.subhead).foregroundStyle(C.texto2)
                TextField("Monto", text: $monto).keyboardType(.numberPad)
                    .font(Estilo.montoPantalla.font).monospacedDigit().foregroundStyle(C.texto)
                    .padding(.horizontal, 14).padding(.vertical, 10)
                    .background(C.superficie, in: RoundedRectangle(cornerRadius: 14, style: .continuous))
                    .overlay(RoundedRectangle(cornerRadius: 14, style: .continuous).strokeBorder(C.linea2, lineWidth: 1))
                TextField("Concepto (ej.: Seña mesa de roble)", text: $concepto)
                    .estilo(.body16).foregroundStyle(C.texto)
                    .padding(.horizontal, 14).padding(.vertical, 12)
                    .background(C.superficie, in: RoundedRectangle(cornerRadius: 14, style: .continuous))
                    .overlay(RoundedRectangle(cornerRadius: 14, style: .continuous).strokeBorder(C.linea2, lineWidth: 1))
                Toggle(isOn: $cuotas) { Text("Permitir cuotas").estilo(.body16).foregroundStyle(C.texto) }
                    .tint(C.ok).padding(.horizontal, 16).padding(.vertical, 10).tarjeta()
                if let link {
                    VStack(alignment: .leading, spacing: 10) {
                        Text(link.absoluteString).estilo(.texto14).foregroundStyle(C.primarioTexto).textSelection(.enabled)
                        HStack(spacing: 10) {
                            Button("Copiar") {
                                UIPasteboard.general.string = link.absoluteString
                                Haptico.exito()
                                estado.avisar("Copiamos el link")
                            }
                            .buttonStyle(BotonTonal())
                            ShareLink(item: link, message: Text("Para pagar \(concepto.isEmpty ? "tu compra" : concepto):")) {
                                Label("Compartir", systemImage: "square.and.arrow.up")
                            }
                            .buttonStyle(BotonTonal())
                        }
                    }
                    .padding(16).frame(maxWidth: .infinity, alignment: .leading).tarjeta()
                }
                Button(creando ? "Creando…" : (link == nil ? "Crear link" : "Crear otro")) { Task { await crear() } }
                    .buttonStyle(BotonPrincipal())
                    .disabled((Int(monto.filter(\.isNumber)) ?? 0) <= 0 || creando)
            }
            .margenPantalla()
            .padding(.bottom, 32)
        }
        .scrollDismissesKeyboard(.interactively)
        .background(C.fondo.ignoresSafeArea())
        .navigationBarTitleDisplayMode(.inline)
        .onAppear { if monto.isEmpty, let montoInicial { monto = String(montoInicial) } }
    }

    private func crear() async {
        struct Cuerpo: Encodable { var monto: Int; var concepto: String; var cuotas: Bool }
        struct R: Decodable { var codigo: String }
        creando = true
        defer { creando = false }
        do {
            let r: R = try await estado.api.enviar("POST", "/v1/panel/links", Cuerpo(
                monto: Int(monto.filter(\.isNumber)) ?? 0, concepto: concepto.trimmingCharacters(in: .whitespaces), cuotas: cuotas))
            link = Web.base.appending(path: "p/\(r.codigo)")
            Haptico.exito()
        } catch let e as ErrorApi {
            estado.avisar(e.detalle, ok: false)
        } catch {
            estado.avisar("No pudimos crear el link", ok: false)
        }
    }
}

// MARK: - Inventario

/// Escanear un código: si el producto existe, se edita; si no, se carga con ese código.
struct EscanearInventario: View {
    @Environment(Estado.self) private var estado
    @Environment(Navegacion.self) private var nav
    @State private var linterna = false
    @State private var buscando = false
    @State private var movimientos: [Movimiento] = []

    struct Movimiento: Decodable, Identifiable {
        var tipo: String
        var texto: String
        var cantidad: Int
        var en: Date
        var id: String { "\(en.timeIntervalSince1970)\(texto)" }
    }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 12) {
                Text("Inventario").estilo(.tituloGrande).foregroundStyle(C.texto)
                ZStack(alignment: .bottomTrailing) {
                    Escaner(linterna: linterna) { codigo in
                        guard !buscando else { return }
                        Task { await abrir(codigo) }
                    }
                    .frame(height: 260)
                    .clipShape(RoundedRectangle(cornerRadius: 20, style: .continuous))
                    Button { linterna.toggle() } label: {
                        Image(systemName: linterna ? "flashlight.on.fill" : "flashlight.off.fill")
                            .font(.system(size: 18)).foregroundStyle(.white)
                            .frame(width: 44, height: 44).background(.black.opacity(0.45), in: Circle())
                    }
                    .padding(12)
                    .accessibilityLabel("Linterna")
                }
                Text(Formato.sinViudas("Apuntá al código de barras: si el producto ya existe lo abrimos, y si no, lo cargás con ese código."))
                    .estilo(.footnote).foregroundStyle(C.texto2)
                if !movimientos.isEmpty {
                    EtiquetaSeccion(texto: "Últimos movimientos de stock").padding(.horizontal, 4).padding(.top, 8)
                    VStack(spacing: 0) {
                        ForEach(Array(movimientos.prefix(30).enumerated()), id: \.element.id) { i, m in
                            if i > 0 { Separador() }
                            HStack {
                                VStack(alignment: .leading, spacing: 2) {
                                    Text(m.texto).estilo(.fila).foregroundStyle(C.texto).lineLimit(2)
                                    Text(Formato.cuando(m.en)).estilo(.footnote).foregroundStyle(C.texto2)
                                }
                                Spacer()
                                Text(m.cantidad > 0 ? "+\(m.cantidad)" : "\(m.cantidad)")
                                    .estilo(.fila).monospacedDigit().foregroundStyle(m.cantidad < 0 ? C.errorTinta : C.okTinta)
                            }
                            .padding(.horizontal, 16).padding(.vertical, 12)
                        }
                    }
                    .tarjeta()
                }
            }
            .margenPantalla()
            .padding(.bottom, 24)
        }
        .background(C.fondo.ignoresSafeArea())
        .navigationBarTitleDisplayMode(.inline)
        .task {
            if let m: [Movimiento] = try? await estado.api.get("/v1/panel/inventario/movimientos") { movimientos = m }
        }
    }

    private func abrir(_ codigo: String) async {
        buscando = true
        defer { buscando = false }
        Haptico.seleccion()
        let l: PaginaProductos? = try? await estado.api.get("/v1/panel/listas/productos", query: [
            URLQueryItem(name: "q", value: codigo), URLQueryItem(name: "cuantos", value: "5"),
        ])
        if let p = l?.filas.first(where: { $0.sku == codigo }) ?? l?.filas.first {
            nav.productos.append(.editorProducto(p.id))
        } else {
            nav.productos.append(.nuevoProducto(codigo))
        }
    }
}
