import SwiftUI

struct VentaHecha: Equatable {
    var numero: Int
    var total: Int
    var medio: String
    var vuelto: Int?
    var cuando = Date()
}

/// 4b · Cobrar: el total (48/700), la grilla 2 × 2 de medios (el elegido con borde de 2
/// `primary-text`) y lo de cada medio: efectivo con «¿Con cuánto paga?», atajos y el vuelto (4b),
/// transferencia con alias y CVU (4d), QR de Mercado Pago que pasa solo a confirmada al acreditarse
/// (4c). La venta se crea con POST /v1/panel/pedidos (canal pos, retiro).
struct Cobrar: View {
    var lineas: [LineaTicket]
    var total: Int
    var volver: () -> Void
    var alCobrar: (VentaHecha) -> Void
    @Environment(Estado.self) private var estado
    @Environment(\.openURL) private var abrir
    @State private var medio = "efectivo"
    @State private var paga = 0
    @State private var cobrando = false
    @State private var negocio: Negocio?
    @State private var qr: QRCobro?
    @State private var editarNegocio = false
    @State private var armarLink = false

    struct Negocio: Decodable { var alias: String?; var cbu: String?; var titular: String?; var nombre: String? }
    struct QRCobro: Identifiable { var numero: Int; var referencia: String; var datos: String; var id: String { referencia } }

    var body: some View {
        VStack(spacing: 0) {
            HStack {
                Button(action: volver) {
                    HStack(spacing: 4) {
                        Image(systemName: "chevron.left").font(.system(size: 17, weight: .semibold))
                        Text("Ticket")
                    }
                    .font(Estilo.body.font).foregroundStyle(C.primarioTexto)
                }
                Spacer()
                Text("Cobrar").estilo(.headline).foregroundStyle(C.texto)
                Spacer()
                Color.clear.frame(width: 60)
            }
            .padding(.leading, 10).padding(.trailing, 16).frame(height: 44)
            ScrollView {
                VStack(spacing: 16) {
                    VStack(spacing: 2) {
                        Text("Total a cobrar").estilo(.subhead).foregroundStyle(C.texto2)
                        Text(Formato.pesos(total)).estilo(Estilo(tamano: 48, peso: .bold, tracking: -0.035, relativo: .largeTitle))
                            .monospacedDigit().foregroundStyle(C.texto)
                    }
                    LazyVGrid(columns: [GridItem(.flexible(), spacing: 10), GridItem(.flexible())], spacing: 10) {
                        tile("efectivo", "Efectivo", "banknote")
                        tile("transf", "Transferencia", "arrow.left.arrow.right")
                        tile("mp", "QR Mercado Pago", "qrcode")
                        tile("link", "Link de pago", "link")
                    }
                    switch medio {
                    case "efectivo": efectivo
                    case "transf": transferencia
                    case "mp":
                        Text(Formato.sinViudas("Mostrale el QR: lo escanea con la app de Mercado Pago o la de su banco, y se confirma solo."))
                            .estilo(.subhead).foregroundStyle(C.texto2).multilineTextAlignment(.center).padding(.horizontal, 10)
                    default:
                        Text(Formato.sinViudas("El link de pago se arma en el panel web y se manda por WhatsApp."))
                            .estilo(.subhead).foregroundStyle(C.texto2).multilineTextAlignment(.center).padding(.horizontal, 10)
                    }
                }
                .padding(.horizontal, 20).padding(.top, 12).padding(.bottom, 20)
            }
            .scrollDismissesKeyboard(.interactively)
        }
        .safeAreaInset(edge: .bottom, spacing: 0) {
            Button {
                Task { await confirmar() }
            } label: {
                if cobrando { ProgressView().tint(.white) } else { Text(textoBoton) }
            }
            .buttonStyle(BotonPrincipal(alto: 56))
            .disabled(cobrando || (medio == "efectivo" && paga > 0 && paga < total))
            .padding(.horizontal, 20).padding(.bottom, 12)
        }
        .background(C.fondo.ignoresSafeArea())
        .task {
            paga = total
            struct Conf: Decodable { var negocio: Negocio? }
            if let c: Conf = try? await estado.api.get("/v1/panel/configuracion") { negocio = c.negocio }
        }
        .sheet(isPresented: $editarNegocio, onDismiss: {
            Task {
                struct Conf: Decodable { var negocio: Negocio? }
                if let c: Conf = try? await estado.api.get("/v1/panel/configuracion") { negocio = c.negocio }
            }
        }) {
            NavigationStack { NegocioCobro() }.environment(estado)
        }
        .sheet(isPresented: $armarLink) {
            NavigationStack { LinksPago(montoInicial: total) }.environment(estado)
        }
        .fullScreenCover(item: $qr) { q in
            CobroQR(qr: q, total: total, nombreTienda: estado.tienda?.nombre ?? "") { ok in
                qr = nil
                if ok { alCobrar(VentaHecha(numero: q.numero, total: total, medio: "QR Mercado Pago")) }
            }
        }
    }

    private var textoBoton: String {
        switch medio {
        case "transf": return "Ya me transfirió"
        case "mp": return "Mostrar el QR"
        case "link": return "Armar el link de pago"
        default: return "Confirmar cobro"
        }
    }

    private func tile(_ id: String, _ texto: String, _ simbolo: String) -> some View {
        let activo = medio == id
        return Button {
            medio = id
            Haptico.seleccion()
        } label: {
            HStack(spacing: 10) {
                Image(systemName: simbolo).font(.system(size: 14, weight: .semibold)).foregroundStyle(C.primarioTexto)
                    .frame(width: 30, height: 30)
                    .background(activo ? C.crema : C.pista, in: RoundedRectangle(cornerRadius: 9, style: .continuous))
                Text(texto).estilo(.fila).foregroundStyle(C.texto).lineLimit(2).multilineTextAlignment(.leading)
                Spacer(minLength: 0)
            }
            .padding(.horizontal, 14)
            .frame(height: 64)
            .background(C.superficie, in: RoundedRectangle(cornerRadius: 16, style: .continuous))
            .overlay(RoundedRectangle(cornerRadius: 16, style: .continuous)
                .strokeBorder(activo ? C.primarioTexto : Color.tema(0x3A1D12, 0xFBF7F2, alfa: 0.12), lineWidth: activo ? 2 : 1))
        }
        .buttonStyle(.plain)
        .accessibilityAddTraits(activo ? .isSelected : [])
    }

    // MARK: Efectivo (4b)

    private var efectivo: some View {
        let redondos = Cobrar.billetes(para: total)
        let vuelto = max(0, paga - total)
        return VStack(spacing: 10) {
            VStack(alignment: .leading, spacing: 12) {
                Text("¿Con cuánto paga?").estilo(.subheadFuerte).foregroundStyle(C.texto2)
                HStack(spacing: 0) {
                    Text("$\(Formato.fino)")
                    TextField("", text: Binding(get: { paga > 0 ? Formato.numero(paga) : "" },
                                                set: { paga = Int($0.filter(\.isNumber)) ?? 0 }))
                        .keyboardType(.numberPad)
                }
                .estilo(Estilo(tamano: 34, peso: .bold, tracking: -0.02, relativo: .largeTitle)).monospacedDigit()
                .foregroundStyle(C.texto)
                .padding(.bottom, 6)
                .overlay(alignment: .bottom) { Rectangle().fill(C.acento).frame(height: 2) }
                ScrollView(.horizontal) {
                    HStack(spacing: 8) {
                        ChipFiltro(texto: "Justo", activo: paga == total) { paga = total }
                        ForEach(redondos, id: \.self) { b in
                            ChipFiltro(texto: Formato.pesos(b), activo: paga == b) { paga = b }
                        }
                    }
                }
                .scrollIndicators(.hidden)
            }
            .padding(16)
            .tarjeta()
            HStack {
                Text("Vuelto").estilo(.headline)
                Spacer()
                Text(Formato.pesos(vuelto)).estilo(Estilo(tamano: 32, peso: .bold, tracking: -0.02, relativo: .title)).monospacedDigit()
            }
            .foregroundStyle(C.okTinta)
            .padding(.horizontal, 18).padding(.vertical, 16)
            .background(C.okFondo, in: RoundedRectangle(cornerRadius: 20, style: .continuous))
            if paga > 0 && paga < total {
                MensajeError(texto: "Faltan \(Formato.pesos(total - paga)) para el total.")
            }
        }
    }

    /// Los billetes redondos que siguen al total: 18.900 → 20.000, 25.000, 30.000.
    static func billetes(para total: Int) -> [Int] {
        var r: [Int] = []
        for paso in [1000, 5000, 10000, 20000] {
            let v = ((total / paso) + 1) * paso
            if v > total, !r.contains(v) { r.append(v) }
        }
        return Array(r.sorted().prefix(3))
    }

    // MARK: Transferencia (4d)

    @ViewBuilder
    private var transferencia: some View {
        if let n = negocio, (n.alias ?? "").isEmpty == false || (n.cbu ?? "").isEmpty == false {
            VStack(spacing: 0) {
                if let a = n.alias, !a.isEmpty { dato("Alias", a, grande: true) }
                if let c = n.cbu, !c.isEmpty {
                    Separador()
                    dato("CVU", Cobrar.agrupar(c), grande: false, copiar: c)
                }
                if let t = n.titular, !t.isEmpty {
                    Separador()
                    VStack(alignment: .leading, spacing: 2) {
                        Text("Titular").estilo(.footnoteFuerte).foregroundStyle(C.texto2)
                        Text(t).estilo(.body16).foregroundStyle(C.texto)
                    }
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .padding(.horizontal, 16).padding(.vertical, 12)
                }
            }
            .tarjeta()
            ShareLink(item: "Para pagar \(Formato.pesos(total)) a \(estado.tienda?.nombre ?? ""):\nAlias: \(n.alias ?? "")\nCVU: \(n.cbu ?? "")\nTitular: \(n.titular ?? "")") {
                Label("Compartir datos", systemImage: "square.and.arrow.up")
            }
            .buttonStyle(BotonSecundario())
        } else {
            EstadoTincho(tincho: .portapapeles, titulo: "Falta tu alias",
                         texto: "Cargá el alias o el CVU una vez y te lo mostramos acá para cada cobro.", tam: 120,
                         accion: "Cargar el alias") { editarNegocio = true }
        }
    }

    static func agrupar(_ cbu: String) -> String {
        guard cbu.count == 22 else { return cbu }
        let s = Array(cbu)
        return [String(s[0..<10]), String(s[10..<14]), String(s[14..<18]), String(s[18..<22])].joined(separator: " ")
    }

    private func dato(_ titulo: String, _ valor: String, grande: Bool, copiar: String? = nil) -> some View {
        HStack(alignment: .center) {
            VStack(alignment: .leading, spacing: 2) {
                Text(titulo).estilo(.footnoteFuerte).foregroundStyle(C.texto2)
                Text(valor).estilo(grande ? Estilo(tamano: 26, peso: .bold, tracking: -0.02, relativo: .title2) : .body16)
                    .monospacedDigit().foregroundStyle(C.texto).textSelection(.enabled)
            }
            Spacer()
            Button("Copiar") {
                UIPasteboard.general.string = copiar ?? valor
                Haptico.exito()
                estado.avisar("Copiamos el \(titulo.lowercased())")
            }
            .buttonStyle(BotonTonal())
        }
        .padding(.horizontal, 16).padding(.vertical, 12)
    }

    // MARK: Confirmar

    private func confirmar() async {
        if medio == "link" {
            armarLink = true
            return
        }
        struct Item: Encodable { var productoId: String; var varianteId: String?; var cantidad: Int }
        struct Cuerpo: Encodable {
            var items: [Item]; var pago: String; var canal = "pos"; var envio = "retiro"; var pagado: Bool
        }
        struct Creado: Decodable { var numero: Int; var total: Int }
        cobrando = true
        defer { cobrando = false }
        let items = lineas.map { Item(productoId: $0.producto.id, varianteId: $0.variante?.id, cantidad: $0.cantidad) }
        do {
            let c: Creado = try await estado.api.enviar("POST", "/v1/panel/pedidos",
                                                        Cuerpo(items: items, pago: medio, pagado: medio != "mp"))
            if medio == "mp" {
                struct Q: Decodable { var referencia: String; var qr: String }
                do {
                    let q: Q = try await estado.api.enviar("POST", "/v1/panel/pedidos/\(c.numero)/qr", [String: String]())
                    qr = QRCobro(numero: c.numero, referencia: q.referencia, datos: q.qr)
                } catch let e as ErrorApi {
                    // Sin Mercado Pago conectado: la venta no queda colgada.
                    try? await PedidosApi.cambiarEstado(estado.api, numero: c.numero, a: "Cancelado")
                    estado.avisar(e.detalle, ok: false)
                }
                return
            }
            alCobrar(VentaHecha(numero: c.numero, total: c.total, medio: medio == "efectivo" ? "Efectivo" : "Transferencia",
                                vuelto: medio == "efectivo" ? max(0, paga - c.total) : nil))
        } catch let e as ErrorApi {
            Haptico.error()
            estado.avisar(e.detalle, ok: false)
        } catch {}
    }
}

/// 4c · El QR de Mercado Pago, para girar el teléfono: nombre de la tienda, monto 52/700, QR de 290 en
/// tarjeta blanca de radio 28 con el isotipo al centro y «Esperando el pago…» con el punto `warn`
/// pulsante. Consulta cada 3 s; al acreditarse pasa solo a confirmada. Brillo al máximo mientras se ve.
struct CobroQR: View {
    var qr: Cobrar.QRCobro
    var total: Int
    var nombreTienda: String
    var terminar: (Bool) -> Void
    @Environment(Estado.self) private var estado
    @State private var pulso = false
    @State private var brillo: CGFloat = UIScreen.main.brightness

    var body: some View {
        VStack(spacing: 18) {
            HStack {
                Button("Cancelar") { Task { await cancelar() } }.font(Estilo.body.font).foregroundStyle(C.primarioTexto)
                Spacer()
            }
            .padding(.horizontal, 16).frame(height: 44)
            Spacer(minLength: 0)
            Text(nombreTienda).estilo(.headline).foregroundStyle(C.texto2)
            Text(Formato.pesos(total)).estilo(.montoCobro).monospacedDigit().foregroundStyle(C.texto)
            if let img = QRTienda.imagen(qr.datos) {
                Image(uiImage: img).interpolation(.none).resizable().scaledToFit()
                    .frame(width: 254, height: 254)
                    .overlay {
                        Image("Marca").resizable().frame(width: 48, height: 48).padding(6).background(.white, in: RoundedRectangle(cornerRadius: 10))
                    }
                    .padding(18)
                    .background(.white, in: RoundedRectangle(cornerRadius: 28, style: .continuous))
                    .shadow(color: Color(hex: 0x2A140C, alfa: 0.35), radius: 25, y: 20)
                    .accessibilityLabel("Código QR para pagar \(Formato.pesos(total))")
            }
            Text(Formato.sinViudas("Escaneá con la app de Mercado Pago o tu banco")).estilo(.subhead).foregroundStyle(C.texto2)
                .multilineTextAlignment(.center)
            HStack(spacing: 8) {
                Circle().fill(C.aviso).frame(width: 9, height: 9).opacity(pulso ? 0.35 : 1)
                Text("Esperando el pago…").estilo(.subheadFuerte).foregroundStyle(C.avisoTinta)
            }
            Spacer(minLength: 0)
        }
        .padding(.horizontal, 20)
        .background(C.fondo.ignoresSafeArea())
        .onAppear {
            brillo = UIScreen.main.brightness
            UIScreen.main.brightness = 1
            withAnimation(.easeInOut(duration: 0.8).repeatForever()) { pulso = true }
        }
        .onDisappear { UIScreen.main.brightness = brillo }
        .task { await esperar() }
    }

    private func esperar() async {
        struct E: Decodable { var estado: String }
        while !Task.isCancelled {
            try? await Task.sleep(for: .seconds(3))
            guard !Task.isCancelled else { return }
            if let e: E = try? await estado.api.get("/v1/panel/cobros/qr/\(qr.referencia)") {
                if ["processed", "approved", "paid"].contains(e.estado) {
                    terminar(true)
                    return
                }
                if ["canceled", "expired", "refunded"].contains(e.estado) {
                    estado.avisar("El pago no se completó.", ok: false)
                    await cancelar()
                    return
                }
            }
        }
    }

    private func cancelar() async {
        try? await estado.api.hacer("DELETE", "/v1/panel/cobros/qr/\(qr.referencia)")
        try? await PedidosApi.cambiarEstado(estado.api, numero: qr.numero, a: "Cancelado")
        terminar(false)
    }
}

/// 4e · ¡Venta cobrada!: Tincho celebrando, el monto 52/700, el detalle, «Emitir factura» /
/// «Compartir ticket» y «Nueva venta» + «Listo». Haptic de éxito.
struct VentaConfirmada: View {
    var venta: VentaHecha
    var nueva: () -> Void
    var listo: () -> Void
    @Environment(Estado.self) private var estado
    @State private var facturada = false

    var body: some View {
        VStack(spacing: 0) {
            ScrollView {
                VStack(spacing: 8) {
                    Image(Tincho.celebrando.imagen).resizable().scaledToFit().frame(width: 220, height: 220)
                        .accessibilityHidden(true)
                    Text("¡Venta cobrada!").estilo(.headline).foregroundStyle(C.okTinta)
                    Text(Formato.pesos(venta.total)).estilo(.montoCobro).monospacedDigit().foregroundStyle(C.texto)
                    Text(detalle).estilo(.subhead).monospacedDigit().foregroundStyle(C.texto2).multilineTextAlignment(.center)
                    HStack(spacing: 10) {
                        acceso(facturada ? "Factura pedida" : "Emitir factura", "doc.text") { Task { await facturar() } }
                            .disabled(facturada)
                        ShareLink(item: "Ticket #\(venta.numero) · \(estado.tienda?.nombre ?? "")\nTotal \(Formato.pesos(venta.total)) · \(venta.medio)") {
                            accesoVista("Compartir ticket", "square.and.arrow.up")
                        }
                        .buttonStyle(.plain)
                    }
                    .padding(.top, 22)
                }
                .padding(.horizontal, 28).padding(.top, 30)
            }
            VStack(spacing: 6) {
                Button("Nueva venta", action: nueva).buttonStyle(BotonPrincipal(alto: 56))
                Button("Listo", action: listo).buttonStyle(BotonTexto(alto: 48))
            }
            .padding(.horizontal, 20).padding(.bottom, 12)
        }
        .background(C.fondo.ignoresSafeArea())
        .sensoryFeedback(.success, trigger: venta)
    }

    private var detalle: String {
        var p = [venta.medio]
        if let v = venta.vuelto, v > 0 { p.append("vuelto \(Formato.pesos(v))") }
        p.append(Formato.cuando(venta.cuando))
        return p.joined(separator: " · ")
    }

    private func acceso(_ t: String, _ s: String, _ a: @escaping () -> Void) -> some View {
        Button(action: a) { accesoVista(t, s) }.buttonStyle(.plain)
    }

    private func accesoVista(_ t: String, _ s: String) -> some View {
        VStack(spacing: 6) {
            Image(systemName: s).font(.system(size: 20)).foregroundStyle(C.primarioTexto)
            Text(t).estilo(.subheadFuerte).foregroundStyle(C.texto)
        }
        .frame(maxWidth: .infinity, minHeight: 78)
        .tarjeta(radio: 18)
    }

    private func facturar() async {
        struct Cuerpo: Encodable { var numeros: [Int]; var accion = "facturar" }
        struct R: Decodable { var hechos: Int }
        do {
            let r: R = try await estado.api.enviar("POST", "/v1/panel/pedidos/lote", Cuerpo(numeros: [venta.numero]))
            facturada = r.hechos > 0
            estado.avisar(r.hechos > 0 ? "Pedimos la factura a ARCA" : "Conectá ARCA en el panel web para facturar.", ok: r.hechos > 0)
        } catch let e as ErrorApi {
            estado.avisar(e.detalle, ok: false)
        } catch {}
    }
}
