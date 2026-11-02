import SwiftUI

/// El buscador (como el ⌘K del panel web): preguntarle a Tincho, las acciones de todos los días y
/// todas las secciones. Escribiendo se filtra (sin tildes ni mayúsculas); «Enter» le pregunta a Tincho.
struct Buscar: View {
    @Environment(Estado.self) private var estado
    @Environment(Navegacion.self) private var nav
    @Environment(\.dismiss) private var cerrar
    @State private var q = ""
    @FocusState private var foco: Bool

    struct Item: Identifiable {
        var titulo: String
        var detalle: String
        var simbolo: String
        var grupo: String
        var permiso: String? = nil
        var accion: (Navegacion) -> Void
        var id: String { grupo + titulo }
    }

    private var items: [Item] {
        let acciones: [Item] = [
            Item(titulo: "Crear pedido", detalle: "Manual, por WhatsApp o en el local", simbolo: "plus", grupo: "Acciones", permiso: "pedidos") { $0.inicio.append(.pedidoManual); $0.pestana = .inicio },
            Item(titulo: "Crear link de pago", detalle: "Cobrá por WhatsApp en segundos", simbolo: "link", grupo: "Acciones", permiso: "pedidos") { $0.inicio.append(.links); $0.pestana = .inicio },
            Item(titulo: "Cargar producto con foto", detalle: "Subí las fotos y completá nombre y precio", simbolo: "camera", grupo: "Acciones", permiso: "catalogo") { $0.inicio.append(.nuevoProducto(nil)); $0.pestana = .inicio },
            Item(titulo: "Cobrar en el local", detalle: "Punto de venta", simbolo: "barcode.viewfinder", grupo: "Acciones") { $0.venta = true },
            Item(titulo: "Compartir mi tienda", detalle: "El link, por WhatsApp o redes", simbolo: "storefront", grupo: "Acciones") { $0.inicio.append(.compartir); $0.pestana = .inicio },
        ]
        func ir(_ titulo: String, _ detalle: String, _ simbolo: String, _ d: Destino, _ permiso: String? = nil) -> Item {
            Item(titulo: titulo, detalle: detalle, simbolo: simbolo, grupo: "Ir a", permiso: permiso) { $0.inicio.append(d); $0.pestana = .inicio }
        }
        let secciones: [Item] = [
            Item(titulo: "Resumen", detalle: "Inicio", simbolo: "square.grid.2x2", grupo: "Ir a") { $0.inicio = []; $0.pestana = .inicio },
            Item(titulo: "Pedidos", detalle: "Ventas", simbolo: "list.bullet.rectangle", grupo: "Ir a", permiso: "pedidos") { $0.filtroPedidos = nil; $0.pedidos = []; $0.pestana = .pedidos },
            Item(titulo: "Para preparar", detalle: "Pedidos", simbolo: "shippingbox", grupo: "Ir a", permiso: "pedidos") { $0.filtroPedidos = "Para preparar"; $0.pestana = .pedidos },
            Item(titulo: "Productos", detalle: "Catálogo", simbolo: "cube.box", grupo: "Ir a") { $0.productos = []; $0.pestana = .productos },
            Item(titulo: "Sin stock", detalle: "Productos", simbolo: "exclamationmark.triangle", grupo: "Ir a") { $0.segmentoProductos = "out"; $0.pestana = .productos },
            ir("Inventario", "Catálogo", "barcode.viewfinder", .inventario, "catalogo"),
            ir("Traer mi tienda", "Catálogo · Tiendanube, Shopify…", "arrow.down.doc", .importar, "catalogo"),
            ir("Clientes", "Ventas", "person.2", .clientes, "clientes"),
            ir("Caja del día", "Ventas", "banknote", .caja, "pedidos"),
            ir("Carritos abandonados", "Crecer", "cart", .carritos, "clientes"),
            ir("Reseñas y preguntas", "Crecer", "star.bubble", .resenas, "clientes"),
            ir("Devoluciones", "Ventas", "arrow.uturn.backward", .devoluciones, "pedidos"),
            ir("Cupones", "Crecer", "ticket", .cupones, "catalogo"),
            ir("Links de pago", "Ventas", "link", .links, "pedidos"),
            ir("Estadísticas", "Crecer", "chart.bar.xaxis", .estadisticas, "pedidos"),
            ir("Compartí tu tienda", "Tienda online", "square.and.arrow.up", .compartir),
            ir("Editar mi tienda", "Plantilla, colores, logo", "paintbrush.pointed", .miTienda, "cuenta"),
            ir("Finanzas", "Plata", "chart.line.uptrend.xyaxis", .finanzas, "finanzas"),
            ir("Facturación de ARCA", "Plata", "doc.text", .facturacion, "finanzas"),
            ir("Plan", "Cuenta", "creditcard", .plan, "cuenta"),
            ir("Datos del negocio y cobro", "Configuración · alias, envíos", "building.2", .negocio, "cuenta"),
            ir("Integraciones", "Mercado Pago, Mercado Libre, Correo", "puzzlepiece.extension", .integraciones, "cuenta"),
            ir("Tincho en Telegram", "Integraciones", "paperplane", .telegram, "cuenta"),
            ir("Google y Meta", "Integraciones · Analytics, píxel", "chart.xyaxis.line", .medicion, "cuenta"),
            ir("Tu IA", "Integraciones · Claude, ChatGPT", "sparkles", .ia, "cuenta"),
            ir("Usuarios y permisos", "Equipo", "person.badge.key", .usuarios, "cuenta"),
            ir("Notificaciones", "Qué te avisamos", "bell.badge", .preferenciasAvisos),
            ir("Perfil, seguridad y tema", "Tu cuenta", "person.crop.circle", .cuenta),
            ir("Avisos", "Lo último que pasó", "bell", .avisos),
            ir("Ayuda", "Preguntas frecuentes", "questionmark.circle", .ayuda),
        ]
        return (acciones + secciones).filter { $0.permiso.map(estado.puede) ?? true }
    }

    private var filtrados: [Item] {
        let t = Buscar.normal(q)
        guard !t.isEmpty else { return items }
        let palabras = t.split(separator: " ").map(String.init)
        return items.filter { i in
            let texto = Buscar.normal("\(i.titulo) \(i.detalle)")
            return palabras.allSatisfy { texto.contains($0) }
        }
    }

    static func normal(_ s: String) -> String {
        s.folding(options: [.diacriticInsensitive, .caseInsensitive], locale: Locale(identifier: "es_AR"))
            .trimmingCharacters(in: .whitespaces)
    }

    var body: some View {
        VStack(spacing: 0) {
            HStack(spacing: 12) {
                Image(systemName: "magnifyingglass").font(.system(size: 18)).foregroundStyle(C.texto2)
                TextField("Buscá una sección o pregunta", text: $q)
                    .estilo(.body).foregroundStyle(C.texto)
                    .focused($foco)
                    .submitLabel(.search)
                    .autocorrectionDisabled()
                    .onSubmit { preguntar() }
                Button { cerrar() } label: {
                    Image(systemName: "xmark").font(.system(size: 15, weight: .semibold)).foregroundStyle(C.texto2)
                        .frame(width: 44, height: 44)
                }
                .accessibilityLabel("Cerrar")
            }
            .padding(.leading, 18).padding(.trailing, 6).padding(.vertical, 6)
            Rectangle().fill(C.linea).frame(height: 0.5)
            ScrollView {
                LazyVStack(alignment: .leading, spacing: 2) {
                    Button { preguntar() } label: {
                        HStack(spacing: 12) {
                            Image(systemName: "sparkles").font(.system(size: 16, weight: .semibold)).foregroundStyle(.white)
                                .frame(width: 36, height: 36)
                                .background(LinearGradient(colors: [Color(hex: 0xE8875F), Color(hex: 0xC24E25)], startPoint: .topLeading, endPoint: .bottomTrailing),
                                            in: RoundedRectangle(cornerRadius: 11, style: .continuous))
                            VStack(alignment: .leading, spacing: 1) {
                                Text("Preguntale a Tincho").estilo(.fila).foregroundStyle(C.texto)
                                Text(q.isEmpty ? "«¿Cuánto vendí ayer por Mercado Libre?»" : "«\(q)»")
                                    .estilo(.footnote).foregroundStyle(C.texto2).lineLimit(1)
                            }
                            Spacer(minLength: 0)
                        }
                        .padding(12)
                        .background(C.primario.opacity(0.09), in: RoundedRectangle(cornerRadius: 14, style: .continuous))
                    }
                    .buttonStyle(.plain)
                    .padding(.bottom, 4)
                    let lista = filtrados
                    ForEach(Array(lista.enumerated()), id: \.element.id) { i, it in
                        if i == 0 || lista[i - 1].grupo != it.grupo {
                            Text(it.grupo).estilo(.footnoteFuerte).foregroundStyle(C.texto3)
                                .padding(.horizontal, 12).padding(.top, 12).padding(.bottom, 4)
                        }
                        Button {
                            Haptico.seleccion()
                            cerrar()
                            it.accion(nav)
                        } label: {
                            HStack(spacing: 12) {
                                Image(systemName: it.simbolo).font(.system(size: 15, weight: .semibold)).foregroundStyle(C.primarioTexto)
                                    .frame(width: 34, height: 34)
                                    .background(C.primario.opacity(0.1), in: RoundedRectangle(cornerRadius: 10, style: .continuous))
                                VStack(alignment: .leading, spacing: 1) {
                                    Text(it.titulo).estilo(.fila).foregroundStyle(C.texto)
                                    Text(it.detalle).estilo(.footnote).foregroundStyle(C.texto2).lineLimit(1)
                                }
                                Spacer(minLength: 0)
                            }
                            .padding(.horizontal, 12).padding(.vertical, 8)
                            .contentShape(Rectangle())
                        }
                        .buttonStyle(.plain)
                    }
                    if lista.isEmpty {
                        Text(Formato.sinViudas("Ninguna sección se llama «\(q)». Tocá «Preguntale a Tincho» y te respondo."))
                            .estilo(.texto14).foregroundStyle(C.texto2).multilineTextAlignment(.center)
                            .frame(maxWidth: .infinity).padding(28)
                    }
                }
                .padding(10)
            }
            .scrollDismissesKeyboard(.interactively)
        }
        .background(C.elevado)
        .onAppear { foco = true }
    }

    private func preguntar() {
        let p = q.trimmingCharacters(in: .whitespacesAndNewlines)
        cerrar()
        if !p.isEmpty { nav.preguntaTincho = p }
        nav.pestana = .tincho
    }
}
