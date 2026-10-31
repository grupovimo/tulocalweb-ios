import SwiftUI

/// 5a · Productos: buscador, chips (Todos, Publicados, Pausados, Sin stock), orden, filas con
/// precio, stock (bajo en `warn-ink`, cero en `error-ink`) y etiquetas de canal; los pausados al
/// 75 %. Tocar una fila abre el stock (11m). Sin conexión, lo guardado con el banner.
struct ListaProductos: View {
    @Environment(Estado.self) private var estado
    @Environment(Navegacion.self) private var nav
    @Environment(\.margen) private var margen
    @State private var segmento: String?
    @State private var orden = "sold"
    @State private var busqueda = ""
    @State private var carga: Carga<PaginaProductos> = .cargando
    @State private var guardadoEn: Date?
    @State private var stockDe: ProductoFila?
    @State private var cargandoMas = false

    /// Stock bajo: con menos de 5 se avisa (11m).
    static let stockBajo = 5

    private static let segmentos: [(String?, String)] = [
        (nil, "Todos"), ("active,out", "Publicados"), ("paused,draft", "Pausados"), ("out", "Sin stock"),
    ]
    private static let ordenes = [("sold", "Más vendidos"), ("stock", "Menos stock"), ("price", "Mayor precio"), ("margin", "Mayor margen")]

    var body: some View {
        ScrollView {
            LazyVStack(alignment: .leading, spacing: 0) {
                if let guardadoEn {
                    BannerSinConexion(guardado: guardadoEn).padding(.horizontal, 16).padding(.top, 2)
                }
                EncabezadoGrande(titulo: "Productos") {
                    BotonRedondo(simbolo: "barcode.viewfinder", etiqueta: "Escanear un código") {
                        nav.web = URL(string: "https://tulocalweb.com.ar/panel/inventario")
                    }
                    .frame(width: 40, height: 40)
                }
                .padding(.top, 8)
                .margenPantalla()
                Buscador(placeholder: "Nombre, SKU o código", texto: $busqueda)
                    .padding(.top, 12).margenPantalla()
                chips.padding(.top, 12)
                if let d = carga.valor, !d.filas.isEmpty {
                    HStack {
                        Text(Formato.cantidad(d.total, "producto", "productos"))
                        Spacer()
                        Menu {
                            Picker("Orden", selection: $orden) {
                                ForEach(ListaProductos.ordenes, id: \.0) { Text($0.1).tag($0.0) }
                            }
                        } label: {
                            HStack(spacing: 4) {
                                Text(ListaProductos.ordenes.first { $0.0 == orden }?.1 ?? "")
                                Image(systemName: "chevron.down").font(.system(size: 11, weight: .bold))
                            }
                            .fontWeight(.semibold).foregroundStyle(C.primarioTexto)
                        }
                    }
                    .estilo(.texto14).foregroundStyle(C.texto2)
                    .padding(.horizontal, margen + 4).padding(.top, 14).padding(.bottom, 8)
                }
                contenido
            }
            .padding(.bottom, 90)
        }
        .scrollDismissesKeyboard(.immediately)
        .background(C.fondo.ignoresSafeArea())
        .toolbar(.hidden, for: .navigationBar)
        .navigationTitle("Productos")
        .refreshable { await cargar() }
        .task(id: "\(segmento ?? "")|\(orden)|\(estado.version)") { await cargar() }
        .task(id: busqueda) {
            guard carga.valor != nil else { return }
            try? await Task.sleep(for: .milliseconds(300))
            guard !Task.isCancelled else { return }
            await cargar()
        }
        .onAppear(perform: tomarSegmento)
        .onChange(of: nav.segmentoProductos) { _, _ in tomarSegmento() }
        .sheet(item: $stockDe) { p in
            StockProducto(producto: p) { await cargar() }
                .hojaTLW([.large])
        }
    }

    private func tomarSegmento() {
        if let s = nav.segmentoProductos {
            segmento = s
            nav.segmentoProductos = nil
        }
    }

    private var chips: some View {
        let seg = carga.valor?.segmentos ?? [:]
        let total = seg.values.reduce(0, +)
        return ScrollView(.horizontal) {
            HStack(spacing: 8) {
                ForEach(ListaProductos.segmentos, id: \.1) { s in
                    let n: Int? = s.0 == nil ? total : (s.0 == "out" ? seg["out"] : nil)
                    ChipFiltro(texto: s.1, cuenta: n, cuentaAlerta: s.0 == "out", activo: segmento == s.0) {
                        segmento = s.0
                        Haptico.seleccion()
                    }
                }
            }
            .padding(.horizontal, margen)
        }
        .scrollIndicators(.hidden)
    }

    @ViewBuilder
    private var contenido: some View {
        switch carga {
        case .cargando:
            EsqueletoFilas(filas: 5, miniatura: 56).margenPantalla().padding(.top, 24)
        case let .error(e):
            EstadoError(que: "los productos", error: e) { Task { await cargar() } }.padding(.top, 60)
        case let .listo(d):
            if d.filas.isEmpty {
                vacio.padding(.top, 50)
            } else {
                VStack(spacing: 0) {
                    ForEach(Array(d.filas.enumerated()), id: \.element.id) { i, p in
                        if i > 0 { Separador(inset: 82) }
                        Button { stockDe = p } label: { FilaProducto(p: p) }
                            .buttonStyle(.plain)
                    }
                }
                .tarjeta()
                .margenPantalla()
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
                         texto: "Probá con otra parte del nombre o con el SKU.")
        } else if segmento == "out" {
            EstadoTincho(tincho: .exito, titulo: "No hay productos sin stock",
                         texto: "Todo lo que publicaste se puede comprar.", accion: "Ver todos") { segmento = nil }
        } else if segmento != nil {
            EstadoTincho(tincho: .caja, titulo: "No hay productos acá", texto: "Probá con otro filtro.",
                         accion: "Ver todos") { segmento = nil }
        } else {
            EstadoTincho(tincho: .foto, titulo: "Cargá tu primer producto",
                         texto: "Con una foto y el precio alcanza para empezar a vender.",
                         accion: "Cargar producto", principal: true) { nav.crear = true }
        }
    }

    private func query(desde: Int) -> [URLQueryItem] {
        var q = [URLQueryItem(name: "orden", value: orden), URLQueryItem(name: "desde", value: String(desde)),
                 URLQueryItem(name: "cuantos", value: "60")]
        if let segmento { q.append(URLQueryItem(name: "seg", value: segmento)) }
        let b = busqueda.trimmingCharacters(in: .whitespaces)
        if !b.isEmpty { q.append(URLQueryItem(name: "q", value: b)) }
        return q
    }

    private var clave: String { "productos-\(segmento ?? "todos")-\(orden)" }

    private func cargar() async {
        do {
            let l: PaginaProductos = try await estado.api.get("/v1/panel/listas/productos", query: query(desde: 0))
            carga = .listo(l)
            guardadoEn = nil
            if busqueda.isEmpty { Guardado.guardar(l, clave: clave) }
        } catch is CancellationError {
        } catch let e as ErrorApi {
            if e.sinConexion, let (l, f) = Guardado.leer(PaginaProductos.self, clave: clave) {
                carga = .listo(l)
                guardadoEn = f
            } else if carga.valor == nil || !e.sinConexion {
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
        if let l: PaginaProductos = try? await estado.api.get("/v1/panel/listas/productos", query: query(desde: d.filas.count)) {
            let ya = Set(d.filas.map(\.id))
            d.filas += l.filas.filter { !ya.contains($0.id) }
            d.total = l.total
            carga = .listo(d)
        }
    }
}

/// La fila de un producto (thumbnail 56).
struct FilaProducto: View {
    var p: ProductoFila
    var body: some View {
        let pausado = !p.publicado
        HStack(spacing: 12) {
            Miniatura(foto: p.foto, tam: 56, radio: 13)
            VStack(alignment: .leading, spacing: 3) {
                Text(p.nombre).estilo(.fila).foregroundStyle(C.texto).lineLimit(2)
                (Text(Formato.pesos(p.precio)).fontWeight(.bold).foregroundColor(C.texto) + Text(" · ") + stock)
                    .estilo(.texto14).monospacedDigit().foregroundStyle(C.texto2)
                HStack(spacing: 4) {
                    if pausado {
                        EtiquetaCanal(texto: "En pausa", fondo: C.apagadoFondo, tinta: C.neutroTinta)
                    } else {
                        if p.canales.contains("web") || p.canales.isEmpty {
                            EtiquetaCanal(texto: "Web", fondo: C.pista, tinta: C.neutroTinta)
                        }
                        if p.canales.contains("ml") {
                            EtiquetaCanal(texto: "ML", fondo: C.avisoFondo, tinta: C.bannerTinta)
                        }
                    }
                }
            }
            Spacer(minLength: 0)
        }
        .padding(.horizontal, 14).padding(.vertical, 12)
        .opacity(pausado ? 0.75 : 1)
        .contentShape(Rectangle())
        .accessibilityElement(children: .combine)
    }

    private var stock: Text {
        if p.stock == 0 {
            return Text("Sin stock").fontWeight(.semibold).foregroundColor(C.errorTinta)
        }
        if p.stock < ListaProductos.stockBajo {
            return Text("Quedan \(p.stock)").fontWeight(.semibold).foregroundColor(C.avisoTinta)
        }
        return Text("\(Formato.numero(p.stock))\(Formato.fino)u.")
    }
}
