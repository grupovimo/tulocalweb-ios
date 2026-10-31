import SwiftUI

/// Una fila de /v1/panel/listas/clientes.
struct ClienteFila: Codable, Hashable, Identifiable {
    var id: String
    var nombre: String
    var telefono: String?
    var email: String?
    var ciudad: String?
    var gastado: Int
    var pedidos: Int
    /// Días desde la última compra.
    var ultima: Double?
    var seg: String?
    var etiquetas: [String]?
}

struct PaginaClientes: Codable {
    var total: Int
    var filas: [ClienteFila]
}

/// «hoy», «ayer», «hace 3 días».
func haceDias(_ d: Double?) -> String {
    guard let d else { return "" }
    let n = Int(d.rounded(.down))
    switch n {
    case ..<1: return "hoy"
    case 1: return "ayer"
    default: return "hace \(n) días"
    }
}

/// 12a · Clientes: buscador, orden y filas con avatar, pedidos, la última compra y el total.
struct Clientes: View {
    @Environment(Estado.self) private var estado
    @Environment(\.margen) private var margen
    @State private var busqueda = ""
    @State private var orden = "ltv"
    @State private var carga: Carga<PaginaClientes> = .cargando

    private static let ordenes = [("ltv", "Más compras"), ("last", "Compra más reciente"), ("orders", "Más pedidos"), ("az", "Nombre")]

    var body: some View {
        ScrollView {
            LazyVStack(alignment: .leading, spacing: 0) {
                Text("Clientes").estilo(.tituloGrande).foregroundStyle(C.texto).margenPantalla()
                    .accessibilityAddTraits(.isHeader)
                Buscador(placeholder: "Nombre, teléfono o ciudad", texto: $busqueda).padding(.top, 12).margenPantalla()
                if let d = carga.valor, !d.filas.isEmpty {
                    HStack {
                        Text(Formato.cantidad(d.total, "cliente", "clientes"))
                        Spacer()
                        Menu {
                            Picker("Orden", selection: $orden) { ForEach(Clientes.ordenes, id: \.0) { Text($0.1).tag($0.0) } }
                        } label: {
                            HStack(spacing: 4) {
                                Text(Clientes.ordenes.first { $0.0 == orden }?.1 ?? "")
                                Image(systemName: "chevron.down").font(.system(size: 11, weight: .bold))
                            }
                            .fontWeight(.semibold).foregroundStyle(C.primarioTexto)
                        }
                    }
                    .estilo(.texto14).foregroundStyle(C.texto2)
                    .padding(.horizontal, margen + 4).padding(.top, 16).padding(.bottom, 8)
                }
                switch carga {
                case .cargando:
                    EsqueletoFilas(filas: 5, miniatura: 42).margenPantalla().padding(.top, 24)
                case let .error(e):
                    EstadoError(que: "los clientes", error: e) { Task { await cargar() } }.padding(.top, 60)
                case let .listo(d):
                    if d.filas.isEmpty {
                        EstadoTincho(tincho: busqueda.isEmpty ? .saludo : .lupa,
                                     titulo: busqueda.isEmpty ? "Todavía no tenés clientes" : "No encontramos «\(busqueda)»",
                                     texto: busqueda.isEmpty ? "Con la primera venta aparece acá, con todo lo que compró." : "Probá con otra parte del nombre o el teléfono.")
                            .padding(.top, 50)
                    } else {
                        VStack(spacing: 0) {
                            ForEach(Array(d.filas.enumerated()), id: \.element.id) { i, c in
                                if i > 0 { Separador(inset: 68) }
                                NavigationLink(value: Destino.cliente(c)) { FilaCliente(c: c) }.buttonStyle(.plain)
                            }
                        }
                        .tarjeta()
                        .margenPantalla()
                    }
                }
            }
            .padding(.bottom, 24)
        }
        .background(C.fondo.ignoresSafeArea())
        .navigationBarTitleDisplayMode(.inline)
        .refreshable { await cargar() }
        .task(id: "\(orden)|\(busqueda)") {
            if carga.valor != nil { try? await Task.sleep(for: .milliseconds(250)) }
            guard !Task.isCancelled else { return }
            await cargar()
        }
    }

    private func cargar() async {
        var q = [URLQueryItem(name: "orden", value: orden), URLQueryItem(name: "cuantos", value: "100")]
        if !busqueda.isEmpty { q.append(URLQueryItem(name: "q", value: busqueda)) }
        do {
            carga = .listo(try await estado.api.get("/v1/panel/listas/clientes", query: q))
        } catch is CancellationError {
        } catch let e as ErrorApi {
            if carga.valor == nil { carga = .error(e) }
        } catch {
            carga = .error(.desconocido)
        }
    }
}

struct FilaCliente: View {
    var c: ClienteFila
    var body: some View {
        let (fondo, tinta) = coloresAvatar(c.nombre)
        HStack(spacing: 12) {
            AvatarIniciales(nombre: c.nombre, tam: 42, fondo: fondo, tinta: tinta)
            VStack(alignment: .leading, spacing: 2) {
                Text(c.nombre).estilo(.fila).foregroundStyle(C.texto).lineLimit(1)
                Text("\(Formato.cantidad(c.pedidos, "pedido", "pedidos")) · última \(haceDias(c.ultima))")
                    .estilo(.footnote).foregroundStyle(C.texto2)
            }
            Spacer(minLength: 6)
            Text(Formato.pesos(c.gastado)).font(Estilo(tamano: 15, peso: .bold, relativo: .subheadline).font)
                .monospacedDigit().foregroundStyle(C.texto)
        }
        .padding(.horizontal, 14).padding(.vertical, 12)
        .contentShape(Rectangle())
        .accessibilityElement(children: .combine)
    }
}

/// 5e · Ficha de cliente: avatar 76, nombre, antigüedad y etiquetas; WhatsApp / Llamar / Mail; 4
/// indicadores y el historial de pedidos.
struct FichaCliente: View {
    var cliente: ClienteFila
    @Environment(Estado.self) private var estado
    @Environment(\.openURL) private var abrir
    @State private var pedidos: [PedidoFila] = []

    var body: some View {
        let (fondo, tinta) = coloresAvatar(cliente.nombre)
        ScrollView {
            VStack(alignment: .leading, spacing: 14) {
                VStack(spacing: 6) {
                    AvatarIniciales(nombre: cliente.nombre, tam: 76, fondo: fondo, tinta: tinta)
                    Text(cliente.nombre).estilo(Estilo(tamano: 26, peso: .bold, tracking: -0.02, relativo: .title)).foregroundStyle(C.texto)
                        .multilineTextAlignment(.center)
                    if let primero = pedidos.last?.creado {
                        Text("Cliente desde \(Formato.mesDe(primero)) de \(Formato.calendario.component(.year, from: primero))")
                            .estilo(.texto14).foregroundStyle(C.texto2)
                    }
                    let etiquetas = [cliente.pedidos >= 3 ? "Frecuente" : nil, cliente.ciudad.flatMap { $0.isEmpty ? nil : $0 }].compactMap { $0 }
                    if !etiquetas.isEmpty {
                        HStack(spacing: 6) {
                            ForEach(Array(etiquetas.enumerated()), id: \.offset) { i, e in
                                Text(e).font(Estilo(tamano: 12, peso: .bold, relativo: .caption).font)
                                    .foregroundStyle(i == 0 && cliente.pedidos >= 3 ? C.primarioTinta : C.neutroTinta)
                                    .padding(.horizontal, 9).padding(.vertical, 3)
                                    .background(i == 0 && cliente.pedidos >= 3 ? C.crema : C.pista, in: Capsule())
                            }
                        }
                        .padding(.top, 2)
                    }
                }
                .frame(maxWidth: .infinity)
                HStack(spacing: 8) {
                    accion("message", "WhatsApp", url: urlWhatsApp(cliente.telefono))
                    accion("phone", "Llamar", url: cliente.telefono.flatMap { URL(string: "tel:\($0.filter { $0.isNumber || $0 == "+" })") })
                    accion("envelope", "Mail", url: cliente.email.flatMap { URL(string: "mailto:\($0)") })
                }
                LazyVGrid(columns: [GridItem(.flexible(), spacing: 10), GridItem(.flexible())], spacing: 10) {
                    Indicador(titulo: "Total comprado", valor: Formato.pesos(cliente.gastado), delta: nil, positivo: true)
                    Indicador(titulo: "Pedidos", valor: "\(cliente.pedidos)", delta: nil, positivo: true)
                    Indicador(titulo: "Ticket promedio", valor: Formato.pesos(cliente.pedidos > 0 ? cliente.gastado / cliente.pedidos : 0), delta: nil, positivo: true)
                    Indicador(titulo: "Última compra", valor: haceDias(cliente.ultima), delta: nil, positivo: true)
                }
                if !pedidos.isEmpty {
                    EtiquetaSeccion(texto: "Pedidos").padding(.horizontal, 4).padding(.top, 4)
                    VStack(spacing: 0) {
                        ForEach(Array(pedidos.enumerated()), id: \.element.id) { i, p in
                            if i > 0 { Separador() }
                            NavigationLink(value: Destino.pedido(p.numero)) {
                                HStack {
                                    VStack(alignment: .leading, spacing: 2) {
                                        Text("#\(p.numero)").estilo(.fila).foregroundStyle(C.texto)
                                        Text(Formato.cuando(p.creado)).estilo(.footnote).foregroundStyle(C.texto2)
                                    }
                                    Spacer()
                                    VStack(alignment: .trailing, spacing: 4) {
                                        Text(Formato.pesos(p.total)).estilo(.fila).monospacedDigit().foregroundStyle(C.texto)
                                        ChipEstado(estado: p.estado, envio: p.envio)
                                    }
                                }
                                .padding(.horizontal, 16).padding(.vertical, 12)
                                .contentShape(Rectangle())
                            }
                            .buttonStyle(.plain)
                        }
                    }
                    .tarjeta()
                }
            }
            .padding(.top, 4)
            .margenPantalla()
            .padding(.bottom, 24)
        }
        .background(C.fondo.ignoresSafeArea())
        .navigationBarTitleDisplayMode(.inline)
        .task {
            if let l: PaginaPedidos = try? await estado.api.get("/v1/panel/pedidos", query: [
                URLQueryItem(name: "q", value: cliente.nombre), URLQueryItem(name: "cuantos", value: "50"),
            ]) {
                pedidos = l.filas.filter { $0.cliente == cliente.nombre }
            }
        }
    }

    private func accion(_ simbolo: String, _ texto: String, url: URL?) -> some View {
        Button { if let url { abrir(url) } } label: {
            VStack(spacing: 3) {
                Image(systemName: simbolo).font(.system(size: 18)).foregroundStyle(C.primarioTexto)
                Text(texto).estilo(.footnoteFuerte).foregroundStyle(C.primarioTinta)
            }
            .frame(maxWidth: .infinity, minHeight: 58)
            .tarjeta(radio: 14)
        }
        .buttonStyle(.plain)
        .disabled(url == nil).opacity(url == nil ? 0.45 : 1)
    }
}
