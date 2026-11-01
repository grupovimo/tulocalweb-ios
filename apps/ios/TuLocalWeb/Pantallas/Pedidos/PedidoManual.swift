import SwiftUI

/// 11a · Pedido manual: para ventas por WhatsApp, Instagram o teléfono. Productos (con variante si
/// tiene), cliente, por dónde llegó, cómo paga, la entrega y si ya está pagado.
struct PedidoManual: View {
    @Environment(Estado.self) private var estado
    @Environment(Navegacion.self) private var nav
    @State private var items: [Item] = []
    @State private var buscar = false
    @State private var cliente = ""
    @State private var telefono = ""
    @State private var nota = ""
    @State private var canal = "wa"
    @State private var pago = "transf"
    @State private var envio = "retiro"
    @State private var pagado = false
    @State private var guardando = false

    struct Item: Identifiable, Equatable {
        var producto: ProductoFila
        var variante: ProductoFila.Variante?
        var cantidad: Int
        var id: String { producto.id + (variante?.id ?? "") }
        var precio: Int { variante?.precio ?? producto.precio }
    }

    private static let canales = [("wa", "WhatsApp"), ("ig", "Instagram"), ("web", "Teléfono / otro")]
    private static let pagos = [("transf", "Transferencia"), ("efectivo", "Efectivo"), ("mp", "Mercado Pago"), ("tarjeta", "Tarjeta")]
    private static let envios = [("retiro", "Retira"), ("moto", "Moto"), ("correo", "Correo")]

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 14) {
                Text("Pedido manual").estilo(.tituloGrande).foregroundStyle(C.texto)
                EtiquetaSeccion(texto: "Productos").padding(.horizontal, 4)
                if items.isEmpty {
                    Button { buscar = true } label: {
                        Label("Agregar productos", systemImage: "plus")
                    }
                    .buttonStyle(BotonSecundario()).frame(maxWidth: .infinity)
                } else {
                    VStack(spacing: 0) {
                        ForEach(Array(items.enumerated()), id: \.element.id) { i, it in
                            if i > 0 { Separador(inset: 70) }
                            HStack(spacing: 12) {
                                Miniatura(foto: it.producto.foto, tam: 44)
                                VStack(alignment: .leading, spacing: 2) {
                                    Text(it.producto.nombre).estilo(.fila).foregroundStyle(C.texto).lineLimit(1)
                                    Text([it.variante?.texto, Formato.pesos(it.precio)].compactMap { $0 }.joined(separator: " · "))
                                        .estilo(.footnote).foregroundStyle(C.texto2)
                                }
                                Spacer(minLength: 6)
                                Stepper("\(it.cantidad)", value: Binding(get: { items[i].cantidad }, set: { n in
                                    if n <= 0 { items.remove(at: i) } else { items[i].cantidad = n }
                                }), in: 0...999)
                                .labelsHidden()
                                Text("\(it.cantidad)").estilo(.fila).monospacedDigit().frame(minWidth: 22)
                            }
                            .padding(.horizontal, 14).padding(.vertical, 10)
                        }
                    }
                    .tarjeta()
                    HStack {
                        Button("Agregar más") { buscar = true }.estilo(.subheadFuerte).foregroundStyle(C.primarioTexto)
                        Spacer()
                        Text("Total \(Formato.pesos(total))").estilo(.headline).monospacedDigit().foregroundStyle(C.texto)
                    }
                    .padding(.horizontal, 4)
                }
                EtiquetaSeccion(texto: "Cliente").padding(.horizontal, 4).padding(.top, 6)
                campo("Nombre", $cliente)
                campo("WhatsApp o teléfono", $telefono, teclado: .phonePad)
                EtiquetaSeccion(texto: "Por dónde llegó").padding(.horizontal, 4).padding(.top, 6)
                Segmentado(opciones: PedidoManual.canales, elegido: $canal)
                EtiquetaSeccion(texto: "Cómo paga").padding(.horizontal, 4).padding(.top, 6)
                Segmentado(opciones: PedidoManual.pagos, elegido: $pago)
                Toggle(isOn: $pagado) {
                    Text("Ya está pagado").estilo(.body16).foregroundStyle(C.texto)
                }
                .tint(C.ok).padding(.horizontal, 16).padding(.vertical, 10).tarjeta()
                EtiquetaSeccion(texto: "Entrega").padding(.horizontal, 4).padding(.top, 6)
                Segmentado(opciones: PedidoManual.envios, elegido: $envio)
                campo("Nota (opcional)", $nota)
                if !pagado {
                    Text(Formato.sinViudas("Queda «Sin pagar»: desde el pedido le mandás el link de pago o confirmás la transferencia."))
                        .estilo(.footnote).foregroundStyle(C.texto3).padding(.horizontal, 4)
                }
            }
            .margenPantalla()
            .padding(.bottom, 32)
        }
        .scrollDismissesKeyboard(.interactively)
        .barraAccion {
            Button(guardando ? "Creando…" : "Crear pedido") { Task { await crear() } }
                .buttonStyle(BotonPrincipal())
                .disabled(items.isEmpty || guardando)
        }
        .background(C.fondo.ignoresSafeArea())
        .navigationBarTitleDisplayMode(.inline)
        .sheet(isPresented: $buscar) {
            ElegirProducto { p, v in
                if let i = items.firstIndex(where: { $0.producto.id == p.id && $0.variante?.id == v?.id }) {
                    items[i].cantidad += 1
                } else {
                    items.append(Item(producto: p, variante: v, cantidad: 1))
                }
            }
            .environment(estado)
            .hojaTLW([.large])
        }
    }

    private var total: Int { items.reduce(0) { $0 + $1.precio * $1.cantidad } }

    private func campo(_ titulo: String, _ texto: Binding<String>, teclado: UIKeyboardType = .default) -> some View {
        TextField(titulo, text: texto)
            .keyboardType(teclado)
            .estilo(.body16).foregroundStyle(C.texto)
            .padding(.horizontal, 14).padding(.vertical, 12)
            .background(C.superficie, in: RoundedRectangle(cornerRadius: 14, style: .continuous))
            .overlay(RoundedRectangle(cornerRadius: 14, style: .continuous).strokeBorder(C.linea2, lineWidth: 1))
    }

    private func crear() async {
        struct It: Encodable { var productoId: String; var varianteId: String?; var cantidad: Int }
        struct Cuerpo: Encodable {
            var items: [It]; var pago: String; var canal: String; var envio: String
            var cliente: String?; var telefono: String?; var nota: String?; var pagado: Bool
        }
        let vacio: (String) -> String? = { let t = $0.trimmingCharacters(in: .whitespaces); return t.isEmpty ? nil : t }
        guardando = true
        defer { guardando = false }
        do {
            let r: ValorJSON = try await estado.api.enviar("POST", "/v1/panel/pedidos", Cuerpo(
                items: items.map { It(productoId: $0.producto.id, varianteId: $0.variante?.id, cantidad: $0.cantidad) },
                pago: pago, canal: canal, envio: envio, cliente: vacio(cliente), telefono: vacio(telefono),
                nota: vacio(nota), pagado: pagado))
            Haptico.exito()
            estado.avisar("Pedido creado")
            if let n = r["numero"]?.entero {
                nav.pedidos = [.pedido(n)]
            } else {
                nav.pedidos = []
            }
        } catch let e as ErrorApi {
            estado.avisar(e.detalle, ok: false)
        } catch {
            estado.avisar("No pudimos crear el pedido", ok: false)
        }
    }
}

/// Buscar un producto (y su variante) para agregarlo.
struct ElegirProducto: View {
    var alElegir: (ProductoFila, ProductoFila.Variante?) -> Void
    @Environment(Estado.self) private var estado
    @Environment(\.dismiss) private var cerrar
    @State private var busqueda = ""
    @State private var filas: [ProductoFila] = []
    @State private var variantesDe: ProductoFila?

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Grabber()
            Text("Agregar productos").estilo(.tituloHojaChico).foregroundStyle(C.texto)
            Buscador(placeholder: "Nombre, SKU o código", texto: $busqueda)
            ScrollView {
                VStack(spacing: 0) {
                    ForEach(Array(filas.enumerated()), id: \.element.id) { i, p in
                        if i > 0 { Separador(inset: 70) }
                        Button {
                            if p.variantes.isEmpty { alElegir(p, nil); estado.avisar("Agregamos \(p.nombre)") } else { variantesDe = p }
                        } label: {
                            HStack(spacing: 12) {
                                Miniatura(foto: p.foto, tam: 44)
                                VStack(alignment: .leading, spacing: 2) {
                                    Text(p.nombre).estilo(.fila).foregroundStyle(C.texto).lineLimit(1)
                                    Text(p.variantes.isEmpty ? "Stock \(p.stock)" : Formato.cantidad(p.variantes.count, "variante", "variantes"))
                                        .estilo(.footnote).foregroundStyle(C.texto2)
                                }
                                Spacer()
                                Text(Formato.pesos(p.precio)).estilo(.fila).monospacedDigit().foregroundStyle(C.texto)
                            }
                            .padding(.horizontal, 14).padding(.vertical, 10).contentShape(Rectangle())
                        }
                        .buttonStyle(.plain)
                    }
                }
                .tarjeta()
            }
            Button("Listo") { cerrar() }.buttonStyle(BotonPrincipal())
        }
        .padding(.horizontal, 20)
        .task(id: busqueda) {
            try? await Task.sleep(for: .milliseconds(250))
            guard !Task.isCancelled else { return }
            var q = [URLQueryItem(name: "cuantos", value: "40"), URLQueryItem(name: "seg", value: "active,out")]
            if !busqueda.isEmpty { q.append(URLQueryItem(name: "q", value: busqueda)) }
            if let l: PaginaProductos = try? await estado.api.get("/v1/panel/listas/productos", query: q) { filas = l.filas }
        }
        .confirmationDialog("Elegí la variante", isPresented: Binding(get: { variantesDe != nil }, set: { if !$0 { variantesDe = nil } })) {
            if let p = variantesDe {
                ForEach(p.variantes.filter(\.activa)) { v in
                    Button("\(v.texto) · \(Formato.pesos(v.precio ?? p.precio))") { alElegir(p, v); estado.avisar("Agregamos \(p.nombre)") }
                }
            }
        }
    }
}
