import SwiftUI

/// 3b · Detalle de un pedido: estado y monto, cliente con acciones, productos con el desglose que
/// cierra, pago, entrega con su línea de tiempo, factura y comprobantes, nota, «Cancelar pedido» y la
/// barra de acción fija con el siguiente estado (según el tipo de entrega).
struct DetallePedido: View {
    var numero: Int
    @Environment(Estado.self) private var estado
    @Environment(\.openURL) private var abrir
    @State private var carga: Carga<PedidoDetalle> = .cargando
    @State private var moviendo = false
    @State private var cancelar = false
    @State private var facturando = false

    var body: some View {
        Group {
            switch carga {
            case .cargando:
                ScrollView { EsqueletoFilas(filas: 4).padding(20) }
            case let .error(e):
                EstadoError(que: "el pedido", error: e) { Task { await cargar() } }
                    .frame(maxHeight: .infinity)
            case let .listo(p):
                contenido(p)
            }
        }
        .background(C.fondo.ignoresSafeArea())
        .navigationTitle("#\(numero)")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            if let p = carga.valor {
                ToolbarItemGroup(placement: .topBarTrailing) {
                    ShareLink(item: textoCompartir(p)) {
                        Image(systemName: "square.and.arrow.up")
                    }
                    .accessibilityLabel("Compartir")
                    Menu {
                        if let u = urlPublica(p) {
                            Button { UIPasteboard.general.url = u; estado.avisar("Copiamos el link del pedido") } label: {
                                Label("Copiar el link del pedido", systemImage: "link")
                            }
                        }
                        Button { abrir(URL(string: "https://tulocalweb.com.ar/panel/pedidos/\(numero)")!) } label: {
                            Label("Abrir en el panel web", systemImage: "arrow.up.right.square")
                        }
                        if puedeCancelar(p) {
                            Button(role: .destructive) { cancelar = true } label: {
                                Label("Cancelar pedido", systemImage: "xmark.circle")
                            }
                        }
                    } label: {
                        Image(systemName: "ellipsis")
                    }
                    .accessibilityLabel("Más acciones")
                }
            }
        }
        .task(id: estado.version) { await cargar() }
        .sheet(isPresented: $cancelar) {
            if let p = carga.valor {
                CancelarPedido(pedido: p) { await cargar() }
                    .hojaTLW([.large])
            }
        }
    }

    // MARK: Contenido

    private func contenido(_ p: PedidoDetalle) -> some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 14) {
                VStack(alignment: .leading, spacing: 6) {
                    ChipEstado(estado: p.estado, envio: p.envio, grande: true)
                    Text(Formato.pesos(p.total)).estilo(.montoPantalla).monospacedDigit().foregroundStyle(C.texto)
                    Text("\(Formato.cuando(p.creado)) · \(Marcas.canal(p.canal)) · \(Formato.cantidad(p.cantidadProductos, "producto", "productos"))")
                        .estilo(.subhead).foregroundStyle(C.texto2)
                }
                .accessibilityElement(children: .combine)
                cliente(p)
                EtiquetaSeccion(texto: "Productos").padding(.horizontal, 4).padding(.top, 6)
                productos(p)
                EtiquetaSeccion(texto: "Pago").padding(.horizontal, 4).padding(.top, 6)
                pago(p)
                EtiquetaSeccion(texto: "Entrega").padding(.horizontal, 4).padding(.top, 6)
                entrega(p)
                EtiquetaSeccion(texto: "Factura y comprobantes").padding(.horizontal, 4).padding(.top, 6)
                factura(p)
                if let nota = p.nota, !nota.isEmpty {
                    EtiquetaSeccion(texto: "Notas internas").padding(.horizontal, 4).padding(.top, 6)
                    VStack(alignment: .leading, spacing: 3) {
                        Text(nota).estilo(.subhead).foregroundStyle(C.texto).lineSpacing(2)
                        Text("Tincho · de la charla con quien compró").estilo(.chip).fontWeight(.regular).foregroundStyle(C.texto2)
                    }
                    .padding(.horizontal, 12).padding(.vertical, 10)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .background(C.crema2, in: RoundedRectangle(cornerRadius: 14, style: .continuous))
                    .padding(.horizontal, 16).padding(.vertical, 14)
                    .tarjeta()
                }
                if puedeCancelar(p) {
                    Button("Cancelar pedido") { cancelar = true }
                        .buttonStyle(BotonTexto(alto: 50, color: C.errorTinta))
                        .font(Estilo.fila.font)
                }
            }
            .padding(.top, 10)
            .margenPantalla()
            .padding(.bottom, 16)
        }
        .safeAreaInset(edge: .bottom, spacing: 0) {
            if let sig = Marcas.siguiente(p.estado, envio: p.envio) {
                Button {
                    Task { await mover(p, a: sig.estado) }
                } label: {
                    if moviendo { ProgressView().tint(.white) } else { Text(sig.boton) }
                }
                .buttonStyle(BotonPrincipal())
                .disabled(moviendo)
                .padding(.horizontal, 20).padding(.top, 12).padding(.bottom, 8)
                .background {
                    ZStack {
                        Rectangle().fill(.ultraThinMaterial)
                        Rectangle().fill(Color.tema(0xFBF7F2, 0x22140D, alfa: 0.94))
                    }
                    .overlay(alignment: .top) { Rectangle().fill(C.linea2).frame(height: 0.5) }
                    .ignoresSafeArea(edges: .bottom)
                }
            }
        }
        .refreshable { await cargar() }
    }

    private func cliente(_ p: PedidoDetalle) -> some View {
        let c = p.comprador
        let (fondo, tinta) = (Color.tema(0xE7EFE9, 0x3FA877, alfaOscuro: 0.18), C.okTinta)
        return VStack(alignment: .leading, spacing: 14) {
            HStack(spacing: 12) {
                AvatarIniciales(nombre: c.nombre, tam: 46, fondo: fondo, tinta: tinta)
                VStack(alignment: .leading, spacing: 2) {
                    Text(c.nombre).estilo(.headline).foregroundStyle(C.texto)
                    Text(antiguedad(c)).estilo(.footnote).foregroundStyle(C.texto2)
                }
                Spacer(minLength: 0)
            }
            if c.telefono != nil || c.email != nil {
                VStack(alignment: .leading, spacing: 3) {
                    if let t = c.telefono { Text(t).foregroundStyle(C.texto).textSelection(.enabled) }
                    if let e = c.email { Text(e).foregroundStyle(C.texto2).textSelection(.enabled) }
                }
                .estilo(.subhead).monospacedDigit()
            }
            HStack(spacing: 8) {
                AccionRapida(simbolo: "phone", texto: "Llamar") {
                    if let t = c.telefono?.filter({ $0.isNumber || $0 == "+" }), let u = URL(string: "tel:\(t)") { abrir(u) }
                }
                .disabled(c.telefono == nil).opacity(c.telefono == nil ? 0.45 : 1)
                AccionRapida(simbolo: "message", texto: "WhatsApp") {
                    if let u = urlWhatsApp(c.telefono, texto: "Hola \(c.nombre.split(separator: " ").first ?? ""), te escribo por tu pedido #\(p.numero).") { abrir(u) }
                }
                .disabled(urlWhatsApp(c.telefono) == nil).opacity(urlWhatsApp(c.telefono) == nil ? 0.45 : 1)
                AccionRapida(simbolo: "doc.on.doc", texto: "Copiar mail") {
                    if let e = c.email {
                        UIPasteboard.general.string = e
                        Haptico.exito()
                        estado.avisar("Copiamos el mail")
                    }
                }
                .disabled(c.email == nil).opacity(c.email == nil ? 0.45 : 1)
            }
        }
        .padding(16)
        .tarjeta()
    }

    private func antiguedad(_ c: PedidoDetalle.Comprador) -> String {
        let mes = Formato.mesDe(c.desde)
        if c.pedidos <= 1 { return "Primer pedido" }
        let n = c.pedidos
        let ord = [1: "1.er", 3: "3.er"][n] ?? "\(n).º"
        let anio = Formato.calendario.component(.year, from: c.desde) == Formato.calendario.component(.year, from: Date())
        return "\(ord) pedido · cliente desde \(mes)\(anio ? "" : " de \(Formato.calendario.component(.year, from: c.desde))")"
    }

    private func productos(_ p: PedidoDetalle) -> some View {
        VStack(spacing: 0) {
            ForEach(Array(p.items.enumerated()), id: \.offset) { i, it in
                if i > 0 { Separador(inset: 78) }
                HStack(spacing: 12) {
                    Miniatura(foto: it.foto, tam: 52, radio: 12)
                    VStack(alignment: .leading, spacing: 2) {
                        Text(it.nombre).estilo(.fila).foregroundStyle(C.texto).lineLimit(2)
                        Text([it.variante, "\(it.cantidad) × \(Formato.pesos(it.precio))"].compactMap { $0 }.joined(separator: " · "))
                            .estilo(.footnote).monospacedDigit().foregroundStyle(C.texto2)
                    }
                    Spacer(minLength: 6)
                    Text(Formato.pesos(it.precio * it.cantidad)).estilo(.fila).monospacedDigit().foregroundStyle(C.texto)
                }
                .padding(.horizontal, 14).padding(.vertical, 12)
                .accessibilityElement(children: .combine)
            }
            Rectangle().fill(C.linea).frame(height: 1)
            VStack(spacing: 8) {
                linea("Subtotal", Formato.pesos(p.subtotal))
                ForEach(Array(p.descuentos.enumerated()), id: \.offset) { _, d in
                    linea(d.concepto, Formato.pesos(-d.monto), color: C.okTinta)
                }
                if p.envio != "retiro" {
                    linea("Envío", p.costoEnvio > 0 ? Formato.pesos(p.costoEnvio) : "Gratis")
                }
                Rectangle().fill(C.linea).frame(height: 1).padding(.vertical, 4)
                HStack {
                    Text("Total")
                    Spacer()
                    Text(Formato.pesos(p.total))
                }
                .estilo(.total).monospacedDigit().foregroundStyle(C.texto)
            }
            .padding(.horizontal, 16).padding(.vertical, 14)
        }
        .tarjeta()
    }

    private func linea(_ a: String, _ b: String, color: Color = C.texto) -> some View {
        HStack(alignment: .firstTextBaseline) {
            Text(a).foregroundStyle(C.texto2)
            Spacer(minLength: 8)
            Text(b).foregroundStyle(color)
        }
        .estilo(.subhead).monospacedDigit()
    }

    private func pago(_ p: PedidoDetalle) -> some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack {
                Text(Marcas.pago(p.pago)).estilo(.fila).foregroundStyle(C.texto)
                Spacer()
                ChipPago(estado: p.estadoPago)
            }
            if p.estado == "Sin pagar" {
                Text(p.pago == "transf"
                     ? Formato.sinViudas("Cuando veas la transferencia en tu cuenta, marcalo como cobrado.")
                     : Formato.sinViudas("Todavía no se acreditó. Si te pagó por otro lado, marcalo como cobrado."))
                    .estilo(.texto14).foregroundStyle(C.texto2)
            }
        }
        .padding(.horizontal, 16).padding(.vertical, 14)
        .tarjeta()
    }

    private func entrega(_ p: PedidoDetalle) -> some View {
        let retiro = p.envio == "retiro"
        let actual = p.historial.firstIndex { !$0.hecho }
        return VStack(alignment: .leading, spacing: 14) {
            VStack(alignment: .leading, spacing: 3) {
                Text(retiro ? "Retiro en el local" : "Envío a domicilio").estilo(.fila).foregroundStyle(C.texto)
                if let e = p.entrega {
                    Text([e.direccion, [e.localidad, e.provincia].compactMap { $0 }.joined(separator: ", ") + " · \(e.cp)"]
                        .filter { !$0.isEmpty }.joined(separator: "\n"))
                        .estilo(.subhead).foregroundStyle(C.texto).lineSpacing(2)
                }
                let detalle = [retiro ? nil : Marcas.envio(p.envio), p.llega.map { "llega \(Formato.dia($0))" },
                               p.seguimiento.map { "seguimiento \($0)" }].compactMap { $0 }
                if !detalle.isEmpty {
                    Text(Formato.mayuscula(detalle.joined(separator: " · "))).estilo(.footnote).foregroundStyle(C.texto2)
                }
            }
            VStack(alignment: .leading, spacing: 0) {
                ForEach(Array(p.historial.enumerated()), id: \.offset) { i, paso in
                    let esActual = i == actual && p.estado != "Cancelado"
                    let ultimo = i == p.historial.count - 1
                    HStack(alignment: .top, spacing: 12) {
                        VStack(spacing: 0) {
                            ZStack {
                                if paso.hecho {
                                    Circle().fill(C.ok)
                                } else if esActual {
                                    Circle().fill(C.superficie)
                                    Circle().strokeBorder(C.acento, lineWidth: 4)
                                } else {
                                    Circle().strokeBorder(Color.tema(0x3A1D12, 0xFBF7F2, alfa: 0.18), lineWidth: 2)
                                }
                            }
                            .frame(width: 14, height: 14)
                            if !ultimo {
                                Rectangle().fill(paso.hecho ? C.ok : Color.tema(0x3A1D12, 0xFBF7F2, alfa: 0.12))
                                    .frame(width: 2).frame(minHeight: 20, maxHeight: .infinity)
                            }
                        }
                        HStack(alignment: .firstTextBaseline) {
                            Text(paso.texto)
                                .font((esActual ? Estilo(tamano: 15, peso: .bold, relativo: .subheadline) : .subhead).font)
                                .foregroundStyle(paso.hecho || esActual ? C.texto : C.texto3)
                            Spacer()
                            if let c = paso.cuando, paso.hecho {
                                Text(Formato.calendario.isDateInToday(c) ? Formato.horaDe(c) : Formato.cuando(c))
                                    .estilo(.subhead).monospacedDigit().foregroundStyle(C.texto2)
                            }
                        }
                        .padding(.top, -3)
                        .padding(.bottom, ultimo ? 0 : 12)
                    }
                    .fixedSize(horizontal: false, vertical: true)
                }
            }
            .accessibilityElement(children: .combine)
        }
        .padding(.horizontal, 16).padding(.vertical, 14)
        .frame(maxWidth: .infinity, alignment: .leading)
        .tarjeta()
    }

    private func factura(_ p: PedidoDetalle) -> some View {
        VStack(spacing: 0) {
            HStack(spacing: 12) {
                VStack(alignment: .leading, spacing: 2) {
                    Text("Factura de ARCA").estilo(.fila).foregroundStyle(C.texto)
                    Text(p.factura ?? "Todavía no la emitiste").estilo(.footnote).foregroundStyle(C.texto2)
                }
                Spacer()
                if p.factura == nil, p.estado != "Cancelado", estado.puede("finanzas") {
                    Button {
                        Task { await facturar(p) }
                    } label: {
                        if facturando { ProgressView() } else { Text("Emitir factura") }
                    }
                    .buttonStyle(BotonTonal())
                    .disabled(facturando)
                }
            }
            .padding(.horizontal, 16).padding(.vertical, 14)
            if let u = urlPublica(p) {
                Separador()
                ShareLink(item: u) {
                    HStack {
                        Text("Comprobante del pedido").estilo(.body16).foregroundStyle(C.texto)
                        Spacer()
                        Image(systemName: "square.and.arrow.up").foregroundStyle(C.primarioTexto)
                    }
                    .padding(.horizontal, 16).padding(.vertical, 14)
                    .contentShape(Rectangle())
                }
                .buttonStyle(.plain)
            }
            if p.envio != "retiro", p.entrega != nil {
                Separador()
                Button {
                    abrir(URL(string: "https://tulocalweb.com.ar/panel/pedidos/imprimir?numeros=\(p.numero)")!)
                } label: {
                    HStack {
                        Text("Etiqueta de envío").estilo(.body16).foregroundStyle(C.texto)
                        Spacer()
                        Image(systemName: "square.and.arrow.up").foregroundStyle(C.primarioTexto)
                    }
                    .padding(.horizontal, 16).padding(.vertical, 14)
                    .contentShape(Rectangle())
                }
                .buttonStyle(.plain)
            }
        }
        .tarjeta()
    }

    // MARK: Acciones

    private func puedeCancelar(_ p: PedidoDetalle) -> Bool {
        p.estado != "Entregado" && p.estado != "Cancelado" && estado.puede("pedidos")
    }

    private func urlPublica(_ p: PedidoDetalle) -> URL? {
        guard let c = p.codigo, let slug = estado.tienda?.slug else { return nil }
        return URL(string: "https://tulocalweb.com.ar/t/\(slug)/pedido/\(c)")
    }

    private func textoCompartir(_ p: PedidoDetalle) -> String {
        var t = "Pedido #\(p.numero) · \(p.comprador.nombre) · \(Formato.pesos(p.total))\n"
        t += p.items.map { "\($0.cantidad) × \($0.nombre)" }.joined(separator: "\n")
        if let u = urlPublica(p) { t += "\n\(u.absoluteString)" }
        return t
    }

    private func cargar() async {
        do {
            let p: PedidoDetalle = try await estado.api.get("/v1/panel/pedidos/\(numero)")
            carga = .listo(p)
        } catch is CancellationError {
        } catch let e as ErrorApi {
            if carga.valor == nil { carga = .error(e) }
        } catch {
            carga = .error(.desconocido)
        }
    }

    private func mover(_ p: PedidoDetalle, a nuevo: String) async {
        moviendo = true
        defer { moviendo = false }
        do {
            try await PedidosApi.cambiarEstado(estado.api, numero: p.numero, a: nuevo)
            Haptico.impacto()
            let antes = p.estado
            estado.avisar("#\(p.numero) \(Marcas.dicho(nuevo, envio: p.envio))") {
                Task {
                    try? await PedidosApi.cambiarEstado(estado.api, numero: p.numero, a: antes)
                    estado.version += 1
                }
            }
            estado.version += 1
            await estado.refrescarResumen()
        } catch let e as ErrorApi {
            Haptico.error()
            estado.avisar(e.detalle, ok: false)
            await cargar()
        } catch {}
    }

    private func facturar(_ p: PedidoDetalle) async {
        struct Cuerpo: Encodable { var numeros: [Int]; var accion = "facturar" }
        struct R: Decodable { var hechos: Int }
        facturando = true
        defer { facturando = false }
        do {
            let r: R = try await estado.api.enviar("POST", "/v1/panel/pedidos/lote", Cuerpo(numeros: [p.numero]))
            if r.hechos > 0 {
                Haptico.exito()
                estado.avisar("Pedimos la factura a ARCA")
                await cargar()
            } else {
                estado.avisar("No se pudo facturar: revisá ARCA en el panel web.", ok: false)
            }
        } catch let e as ErrorApi {
            estado.avisar(e.detalle, ok: false)
        } catch {}
    }
}
