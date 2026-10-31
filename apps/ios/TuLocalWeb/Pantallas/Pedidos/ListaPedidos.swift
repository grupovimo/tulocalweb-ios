import SwiftUI

/// 3a · Pedidos: buscador, chips por etapa con contador, secciones por día y filas con swipe
/// (izquierda: «Más» y el siguiente estado; derecha: «Cobrado» en los sin pagar), toast con
/// «Deshacer». 3d vacío con filtro, 12f error, 5a sin conexión.
struct ListaPedidos: View {
    @Environment(Estado.self) private var estado
    @Environment(Navegacion.self) private var nav
    @Environment(\.margen) private var margen
    @State private var filtro: String?
    @State private var orden = "nuevos"
    @State private var busqueda = ""
    @State private var carga: Carga<ListaPedidos.Datos> = .cargando
    @State private var guardadoEn: Date?
    @State private var cargandoMas = false
    @State private var masDe: PedidoFila?

    struct Datos {
        var filas: [PedidoFila]
        var total: Int
        var etapas: [String: Int]
    }

    private static let etapas: [(String?, String)] = [
        (nil, "Todos"), ("Para preparar", "Para preparar"), ("Sin pagar", "Sin pagar"),
        ("Listo para despachar", "Listo para despachar"), ("En camino", "En camino"),
        ("Entregado", "Entregados"), ("Cancelado", "Cancelados"),
    ]

    var body: some View {
        ScrollView {
            LazyVStack(alignment: .leading, spacing: 0, pinnedViews: []) {
                if let guardadoEn {
                    BannerSinConexion(guardado: guardadoEn).padding(.horizontal, 16).padding(.top, 2).padding(.bottom, 8)
                }
                EncabezadoGrande(titulo: "Pedidos") {
                    Menu {
                        Picker("Orden", selection: $orden) {
                            Text("Más nuevos primero").tag("nuevos")
                            Text("Más urgentes primero").tag("urgentes")
                            Text("Mayor monto primero").tag("monto")
                            Text("Más viejos primero").tag("viejos")
                        }
                    } label: {
                        Image(systemName: "line.3.horizontal.decrease")
                            .font(.system(size: 17, weight: .semibold)).foregroundStyle(C.texto)
                            .frame(width: 40, height: 40)
                            .background(C.superficie, in: Circle())
                            .overlay(Circle().strokeBorder(C.linea, lineWidth: 1))
                    }
                    .accessibilityLabel("Ordenar")
                }
                .padding(.top, 4)
                .margenPantalla()
                Buscador(placeholder: "Número, cliente o producto", texto: $busqueda)
                    .padding(.top, 12).margenPantalla()
                chips.padding(.top, 12)
                contenido
            }
            .padding(.bottom, 24)
        }
        .scrollDismissesKeyboard(.immediately)
        .background(C.fondo.ignoresSafeArea())
        .toolbar(.hidden, for: .navigationBar)
        .navigationTitle("Pedidos")
        .refreshable { await cargar() }
        .task(id: "\(filtro ?? "")|\(orden)|\(estado.version)") { await cargar() }
        .task(id: busqueda) {
            guard !busqueda.isEmpty || carga.valor != nil else { return }
            try? await Task.sleep(for: .milliseconds(300))
            guard !Task.isCancelled else { return }
            await cargar()
        }
        .onAppear(perform: tomarFiltro)
        .onChange(of: nav.filtroPedidos) { _, _ in tomarFiltro() }
        .confirmationDialog(masDe.map { "Pedido #\($0.numero)" } ?? "", isPresented: Binding(get: { masDe != nil }, set: { if !$0 { masDe = nil } }),
                            titleVisibility: .visible, presenting: masDe) { p in
            Button("Ver el pedido") { nav.pedidos.append(.pedido(p.numero)) }
            if let tel = p.telefonoWhatsApp {
                Button("Escribirle por WhatsApp") { UIApplication.shared.open(tel) }
            }
            if let ant = Marcas.anterior(p.estado, envio: p.envio), p.estado != "Cancelado" {
                Button("Volver a «\(ChipEstado.texto(ant, envio: p.envio))»") { Task { await mover(p, a: ant) } }
            }
            if p.estado != "Entregado" && p.estado != "Cancelado" {
                Button("Cancelar pedido", role: .destructive) { nav.pedidos.append(.pedido(p.numero)) }
            }
        }
    }

    private func tomarFiltro() {
        if let f = nav.filtroPedidos {
            filtro = f
            nav.filtroPedidos = nil
        }
    }

    // MARK: Chips

    private var chips: some View {
        let etapas = carga.valor?.etapas ?? [:]
        let total = etapas.values.reduce(0, +)
        return ScrollViewReader { lector in
            ScrollView(.horizontal) {
                HStack(spacing: 8) {
                    ForEach(ListaPedidos.etapas, id: \.1) { e in
                        let n = e.0 == nil ? total : (etapas[e.0!] ?? 0)
                        let mostrarN = e.0 == nil || e.0 == "Para preparar" || e.0 == "Sin pagar"
                        ChipFiltro(texto: e.1, cuenta: mostrarN ? n : nil, activo: filtro == e.0) {
                            filtro = e.0
                            Haptico.seleccion()
                        }
                        .id(e.1)
                    }
                }
                .padding(.horizontal, margen)
            }
            .scrollIndicators(.hidden)
            .onChange(of: filtro) { _, f in
                withAnimation { lector.scrollTo(ListaPedidos.etapas.first { $0.0 == f }?.1 ?? "Todos", anchor: .center) }
            }
        }
    }

    // MARK: Contenido

    @ViewBuilder
    private var contenido: some View {
        switch carga {
        case .cargando:
            EsqueletoFilas(filas: 5).margenPantalla().padding(.top, 24)
        case let .error(e):
            EstadoError(que: "los pedidos", error: e) { Task { await cargar() } }.padding(.top, 60)
        case let .listo(d):
            if d.filas.isEmpty {
                vacio.padding(.top, 50)
            } else {
                ForEach(secciones(d.filas), id: \.0) { titulo, filas in
                    Text(titulo).estilo(.label).foregroundStyle(C.texto2)
                        .padding(.top, 18).padding(.bottom, 6).margenPantalla()
                    VStack(spacing: 0) {
                        ForEach(Array(filas.enumerated()), id: \.element.id) { i, p in
                            if i > 0 { Separador(inset: 74) }
                            fila(p)
                        }
                    }
                    .tarjeta()
                    .margenPantalla()
                }
                if d.filas.count < d.total {
                    ProgressView().frame(maxWidth: .infinity).padding(20)
                        .onAppear { Task { await cargarMas() } }
                }
            }
        }
    }

    @ViewBuilder
    private var vacio: some View {
        if !busqueda.isEmpty {
            EstadoTincho(tincho: .lupa, titulo: "No encontramos «\(busqueda)»",
                         texto: "Probá con el número del pedido, el nombre de quien compró o un producto.")
        } else if let f = filtro {
            let (tincho, titulo, texto): (Tincho, String, String) = {
                switch f {
                case "En camino": return (.envio, "No hay pedidos en camino", "Cuando despaches uno, lo ves acá hasta que llegue.")
                case "Para preparar": return (.caja, "No hay pedidos por preparar", "Cuando entre uno nuevo, te avisamos al instante.")
                case "Sin pagar": return (.ticket, "No hay pedidos sin pagar", "Todos los pedidos están cobrados.")
                case "Listo para despachar": return (.caja, "No hay pedidos listos", "Los que prepares quedan acá hasta que salgan.")
                default: return (.portapapeles, "No hay pedidos \(ChipEstado.texto(f, envio: "").lowercased())", "Cuando haya, los ves acá.")
                }
            }()
            EstadoTincho(tincho: tincho, titulo: titulo, texto: texto, accion: "Ver todos los pedidos") { filtro = nil }
        } else {
            EstadoTincho(tincho: .durmiendo, titulo: "Todavía no tenés pedidos",
                         texto: "Compartí el link de tu tienda y el primero llega solo.")
        }
    }

    private func fila(_ p: PedidoFila) -> some View {
        var derecha: [AccionFila] = []
        var izquierda: [AccionFila] = []
        derecha.append(AccionFila(texto: "Más", simbolo: "ellipsis", color: Color(hex: 0x7D6656)) { masDe = p })
        if p.estado != "Sin pagar", let sig = Marcas.siguiente(p.estado, envio: p.envio) {
            derecha.append(AccionFila(texto: sig.corto, simbolo: "checkmark", color: Color(hex: 0xB0441D)) {
                Task { await mover(p, a: sig.estado) }
            })
        }
        if p.estado == "Sin pagar" {
            izquierda.append(AccionFila(texto: "Cobrado", simbolo: "banknote", color: C.ok) {
                Task { await mover(p, a: "Para preparar", texto: "#\(p.numero) cobrado") }
            })
        }
        return FilaDeslizable(izquierda: izquierda, derecha: derecha) {
            Button {
                nav.pedidos.append(.pedido(p.numero))
            } label: {
                FilaPedido(p: p)
            }
            .buttonStyle(.plain)
        }
    }

    private func secciones(_ filas: [PedidoFila]) -> [(String, [PedidoFila])] {
        var r: [(String, [PedidoFila])] = []
        for f in filas {
            let t = Formato.seccion(f.creado)
            if let i = r.firstIndex(where: { $0.0 == t }) { r[i].1.append(f) } else { r.append((t, [f])) }
        }
        return r
    }

    // MARK: Datos

    private func query(desde: Int) -> [URLQueryItem] {
        var q = [URLQueryItem(name: "orden", value: orden), URLQueryItem(name: "desde", value: String(desde)),
                 URLQueryItem(name: "cuantos", value: "60")]
        if let filtro { q.append(URLQueryItem(name: "estado", value: filtro)) }
        let b = busqueda.trimmingCharacters(in: .whitespaces)
        if !b.isEmpty { q.append(URLQueryItem(name: "q", value: b)) }
        return q
    }

    private var clave: String { "pedidos-\(filtro ?? "todos")-\(orden)" }

    private func cargar() async {
        do {
            let l: PaginaPedidos = try await estado.api.get("/v1/panel/pedidos", query: query(desde: 0))
            carga = .listo(Datos(filas: l.filas, total: l.total, etapas: l.etapas))
            guardadoEn = nil
            if busqueda.isEmpty { Guardado.guardar(l, clave: clave) }
        } catch is CancellationError {
        } catch let e as ErrorApi {
            if e.sinConexion, let (l, f) = Guardado.leer(PaginaPedidos.self, clave: clave) {
                carga = .listo(Datos(filas: l.filas, total: l.filas.count, etapas: l.etapas))
                guardadoEn = f
            } else if case .listo = carga, e.sinConexion {
                estado.avisar("Sin conexión. Probá de nuevo en un rato.", ok: false)
            } else {
                carga = .error(e)
            }
        } catch {
            carga = .error(.desconocido)
        }
    }

    private func cargarMas() async {
        guard !cargandoMas, var d = carga.valor else { return }
        cargandoMas = true
        defer { cargandoMas = false }
        if let l: PaginaPedidos = try? await estado.api.get("/v1/panel/pedidos", query: query(desde: d.filas.count)) {
            let ya = Set(d.filas.map(\.numero))
            d.filas += l.filas.filter { !ya.contains($0.numero) }
            d.total = l.total
            carga = .listo(d)
        }
    }

    /// Cambio de estado optimista: se ve al instante, vuelve atrás si la API no lo acepta, y el toast
    /// ofrece «Deshacer» (el estado anterior).
    private func mover(_ p: PedidoFila, a nuevo: String, texto: String? = nil) async {
        guard var d = carga.valor, let i = d.filas.firstIndex(where: { $0.numero == p.numero }) else { return }
        let antes = d.filas[i]
        Haptico.impacto()
        d.filas[i].estado = nuevo
        // Si quedó afuera del filtro, sale de la lista.
        if let filtro, filtro != nuevo { d.filas.remove(at: i) }
        d.etapas[antes.estado, default: 1] -= 1
        d.etapas[nuevo, default: 0] += 1
        withAnimation(.snappy) { carga = .listo(d) }
        do {
            try await PedidosApi.cambiarEstado(estado.api, numero: p.numero, a: nuevo)
            estado.avisar(texto ?? "#\(p.numero) \(Marcas.dicho(nuevo, envio: p.envio))") {
                Task {
                    try? await PedidosApi.cambiarEstado(estado.api, numero: p.numero, a: antes.estado)
                    await cargar()
                    await estado.refrescarResumen()
                }
            }
            await estado.refrescarResumen()
        } catch let e as ErrorApi {
            Haptico.error()
            estado.avisar(e.detalle, ok: false)
            await cargar()
        } catch {}
    }
}

/// La fila de un pedido (thumbnail 48, nombre y monto 16/600, meta 13 y chip).
struct FilaPedido: View {
    var p: PedidoFila
    var body: some View {
        HStack(spacing: 12) {
            Miniatura(foto: p.fotoProducto ?? p.foto, tam: 48, radio: 12)
            VStack(alignment: .leading, spacing: 5) {
                HStack(alignment: .firstTextBaseline) {
                    Text(p.cliente).estilo(.fila).foregroundStyle(C.texto).lineLimit(1)
                    Spacer(minLength: 8)
                    Text(Formato.pesos(p.total)).estilo(.fila).monospacedDigit().foregroundStyle(C.texto)
                }
                HStack {
                    Text("#\(p.numero) · \(Marcas.canalCorto(p.canal)) · \(Formato.horaDe(p.creado))")
                        .estilo(.footnote).foregroundStyle(C.texto2).lineLimit(1)
                    Spacer(minLength: 6)
                    ChipEstado(estado: p.estado, envio: p.envio)
                }
            }
        }
        .padding(.horizontal, 14).padding(.vertical, 13)
        .contentShape(Rectangle())
        .accessibilityElement(children: .combine)
    }
}

extension PedidoFila {
    var telefonoWhatsApp: URL? { urlWhatsApp(telefono) }
}

/// wa.me con el número en formato internacional (Argentina: 54 9 + área + número).
func urlWhatsApp(_ telefono: String?, texto: String? = nil) -> URL? {
    guard let t = telefono else { return nil }
    var d = t.filter(\.isNumber)
    guard d.count >= 8 else { return nil }
    if d.hasPrefix("0") { d.removeFirst() }
    if !d.hasPrefix("54") { d = "549" + d }
    var c = URLComponents(string: "https://wa.me/\(d)")!
    if let texto { c.queryItems = [URLQueryItem(name: "text", value: texto)] }
    return c.url
}

enum PedidosApi {
    static func cambiarEstado(_ api: Api, numero: Int, a estado: String) async throws {
        struct Cuerpo: Encodable { var estado: String }
        try await api.hacer("POST", "/v1/panel/pedidos/\(numero)/estado", Cuerpo(estado: estado))
    }
}
