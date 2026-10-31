import SwiftUI

/// 7c · Centro de avisos: agrupado por día, punto de color por tipo, punto `accent` de no leído y
/// «Leí todo». Con /v1/panel/notificaciones (la campanita del panel, §45).
struct Avisos: View {
    @Environment(Estado.self) private var estado
    @Environment(Navegacion.self) private var nav
    @Environment(\.openURL) private var abrir
    @State private var carga: Carga<[Notificacion]> = .cargando
    @State private var siguiente: String?

    var body: some View {
        ScrollView {
            LazyVStack(alignment: .leading, spacing: 0) {
                Text("Avisos").estilo(.tituloGrande).foregroundStyle(C.texto).margenPantalla()
                    .accessibilityAddTraits(.isHeader)
                switch carga {
                case .cargando:
                    EsqueletoFilas(filas: 4, miniatura: 36).margenPantalla().padding(.top, 24)
                case let .error(e):
                    EstadoError(que: "los avisos", error: e) { Task { await cargar() } }.padding(.top, 60)
                case let .listo(lista):
                    if lista.isEmpty {
                        EstadoTincho(tincho: .durmiendo, titulo: "No tenés avisos",
                                     texto: "Cuando entre una venta o pase algo en tu tienda, te avisamos acá.")
                            .padding(.top, 60)
                    } else {
                        ForEach(secciones(lista), id: \.0) { titulo, items in
                            Text(titulo).estilo(.label).foregroundStyle(C.texto2)
                                .padding(.top, 14).padding(.bottom, 6).margenPantalla()
                            VStack(spacing: 0) {
                                ForEach(Array(items.enumerated()), id: \.element.id) { i, n in
                                    if i > 0 { Separador(inset: 62) }
                                    Button { Task { await tocar(n) } } label: { fila(n) }.buttonStyle(.plain)
                                }
                            }
                            .tarjeta()
                            .margenPantalla()
                        }
                        if siguiente != nil {
                            ProgressView().frame(maxWidth: .infinity).padding(20)
                                .onAppear { Task { await cargarMas() } }
                        }
                    }
                }
            }
            .padding(.bottom, 24)
        }
        .background(C.fondo.ignoresSafeArea())
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .topBarTrailing) {
                Button("Leí todo") { Task { await leerTodo() } }
                    .disabled(!(carga.valor?.contains { !$0.leida } ?? false))
            }
        }
        .refreshable { await cargar() }
        .task(id: estado.version) { await cargar() }
    }

    private func fila(_ n: Notificacion) -> some View {
        let (fondo, punto) = Avisos.colores(n.tipo)
        return HStack(alignment: .top, spacing: 12) {
            Circle().fill(fondo).frame(width: 36, height: 36)
                .overlay(Circle().fill(punto).frame(width: 10, height: 10))
            VStack(alignment: .leading, spacing: 2) {
                HStack(alignment: .firstTextBaseline) {
                    Text(n.titulo).font(Estilo(tamano: 15, peso: n.leida ? .semibold : .bold, relativo: .subheadline).font)
                        .foregroundStyle(C.texto)
                    Spacer(minLength: 6)
                    Text(Formato.horaDe(n.creada)).estilo(.footnote).monospacedDigit().foregroundStyle(C.texto2)
                }
                Text(n.detalle).estilo(.texto14).foregroundStyle(C.neutroTinta).lineSpacing(1)
                    .fixedSize(horizontal: false, vertical: true)
            }
            if !n.leida {
                Circle().fill(C.acento).frame(width: 9, height: 9).padding(.top, 5)
                    .accessibilityLabel("Sin leer")
            }
        }
        .padding(.horizontal, 14).padding(.vertical, 12)
        .contentShape(Rectangle())
        .accessibilityElement(children: .combine)
    }

    /// Fondo y punto por tipo (lib/marcas.ts): ventas y pagos en verde, avisos en ámbar, errores en
    /// terracota.
    static func colores(_ tipo: String) -> (Color, Color) {
        switch tipo {
        case "web", "mp", "ok", "pos": return (C.okFondo, C.ok)
        case "transf", "ml", "warn", "star": return (C.avisoFondo, C.aviso)
        case "err", "arca": return (C.errorFondo, C.error)
        default: return (C.pista, C.texto3)
        }
    }

    private func secciones(_ l: [Notificacion]) -> [(String, [Notificacion])] {
        var r: [(String, [Notificacion])] = []
        for n in l {
            let t = Formato.seccion(n.creada)
            if let i = r.firstIndex(where: { $0.0 == t }) { r[i].1.append(n) } else { r.append((t, [n])) }
        }
        return r
    }

    private func cargar() async {
        do {
            let p: PaginaNotificaciones = try await estado.api.get("/v1/panel/notificaciones", query: [URLQueryItem(name: "limite", value: "30")])
            carga = .listo(p.notificaciones)
            siguiente = p.siguiente
        } catch is CancellationError {
        } catch let e as ErrorApi {
            if carga.valor == nil { carga = .error(e) }
        } catch {
            carga = .error(.desconocido)
        }
    }

    private func cargarMas() async {
        guard let s = siguiente, var l = carga.valor else { return }
        if let p: PaginaNotificaciones = try? await estado.api.get("/v1/panel/notificaciones", query: [
            URLQueryItem(name: "limite", value: "30"), URLQueryItem(name: "cursor", value: s),
        ]) {
            l += p.notificaciones
            carga = .listo(l)
            siguiente = p.siguiente
        } else {
            siguiente = nil
        }
    }

    private func tocar(_ n: Notificacion) async {
        if !n.leida, var l = carga.valor, let i = l.firstIndex(of: n) {
            l[i].leida = true
            carga = .listo(l)
            try? await estado.api.hacer("POST", "/v1/panel/notificaciones/\(n.id)/leida")
        }
        // La ruta es la del panel web (/panel/pedidos/1048): las que tienen pantalla en la app se abren acá.
        guard let ruta = n.ruta else { return }
        let partes = ruta.split(separator: "/").map(String.init)
        if partes.count >= 3, partes[0] == "panel", partes[1] == "pedidos", let num = Int(partes[2]) {
            nav.inicio.append(.pedido(num))
        } else if ruta.hasPrefix("/panel/pedidos") {
            nav.tocar(.pedidos)
        } else if ruta.hasPrefix("/panel/productos") || ruta.hasPrefix("/panel/inventario") {
            nav.tocar(.productos)
        } else if let u = URL(string: "https://tulocalweb.com.ar\(ruta)") {
            nav.web = u
        }
    }

    private func leerTodo() async {
        guard var l = carga.valor else { return }
        for i in l.indices { l[i].leida = true }
        carga = .listo(l)
        try? await estado.api.hacer("POST", "/v1/panel/notificaciones/leidas")
        Haptico.exito()
    }
}
