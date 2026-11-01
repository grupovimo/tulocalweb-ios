import AudioToolbox
import SwiftUI

/// El ticket de una venta en el local.
struct LineaTicket: Identifiable, Equatable {
    var producto: ProductoFila
    var variante: ProductoFila.Variante?
    var cantidad: Int
    var id: String { producto.id + (variante.map { ":" + $0.id } ?? "") }
    var precio: Int { variante?.precio ?? producto.precio }
    var nombre: String { variante.map { "\(producto.nombre) · \($0.texto)" } ?? producto.nombre }
}

/// 4a · Venta rápida: cámara arriba (300) con visor y línea de escaneo, la hoja del ticket que se
/// superpone 20, buscador, filas con stepper (la recién escaneada resaltada) y «Cobrar» con el
/// total. Sigue a 4b (medios), 4c (QR), 4d (transferencia) y 4e (confirmada).
struct VentaLocal: View {
    @Environment(\.dismiss) private var cerrar
    @Environment(Estado.self) private var estado
    @State private var lineas: [LineaTicket] = []
    @State private var recien: String?
    @State private var busqueda = ""
    @State private var resultados: [ProductoFila] = []
    @State private var linterna = false
    @State private var elegirVariante: ProductoFila?
    @State private var paso: Paso = .ticket
    @State private var hecha: VentaHecha?
    @FocusState private var buscando: Bool

    enum Paso { case ticket, cobrar }

    var total: Int { lineas.reduce(0) { $0 + $1.precio * $1.cantidad } }
    var unidades: Int { lineas.reduce(0) { $0 + $1.cantidad } }

    var body: some View {
        ZStack {
            if let hecha {
                VentaConfirmada(venta: hecha) {
                    lineas = []
                    self.hecha = nil
                    paso = .ticket
                } listo: { cerrar() }
                .transition(.opacity)
            } else if paso == .cobrar {
                Cobrar(lineas: lineas, total: total, volver: { paso = .ticket }) { v in
                    Haptico.exito()
                    withAnimation { hecha = v }
                    estado.version += 1
                    Task { await estado.refrescarResumen() }
                }
                .transition(.move(edge: .trailing))
            } else {
                ticket
            }
        }
        .animation(.easeInOut(duration: 0.25), value: paso)
        .background(C.fondo.ignoresSafeArea())
    }

    // MARK: Ticket

    private var ticket: some View {
        VStack(spacing: 0) {
            camara
            VStack(spacing: 12) {
                Buscador(placeholder: "Buscá por nombre o código", texto: $busqueda)
                    .focused($buscando)
                    .background(C.superficie, in: RoundedRectangle(cornerRadius: 12, style: .continuous))
                if !resultados.isEmpty && !busqueda.isEmpty {
                    ScrollView {
                        VStack(spacing: 0) {
                            ForEach(Array(resultados.enumerated()), id: \.element.id) { i, p in
                                if i > 0 { Separador(inset: 62) }
                                Button { agregar(p) } label: {
                                    HStack(spacing: 12) {
                                        Miniatura(foto: p.foto, tam: 38, radio: 10)
                                        VStack(alignment: .leading, spacing: 1) {
                                            Text(p.nombre).estilo(.fila).foregroundStyle(C.texto).lineLimit(1)
                                            Text("\(Formato.pesos(p.precio)) · \(p.stock == 0 ? "sin stock" : "\(p.stock)\(Formato.fino)u.")")
                                                .estilo(.footnote).monospacedDigit().foregroundStyle(p.stock == 0 ? C.errorTinta : C.texto2)
                                        }
                                        Spacer()
                                        Image(systemName: "plus.circle.fill").font(.system(size: 22)).foregroundStyle(C.primarioTexto)
                                    }
                                    .padding(.horizontal, 12).padding(.vertical, 9).contentShape(Rectangle())
                                }
                                .buttonStyle(.plain)
                            }
                        }
                        .tarjeta()
                    }
                    .frame(maxHeight: 260)
                }
                HStack {
                    Text("Ticket · \(Formato.cantidad(unidades, "producto", "productos"))")
                    Spacer()
                    if !lineas.isEmpty {
                        Button("Vaciar") { withAnimation { lineas = [] } }.foregroundStyle(C.primarioTexto)
                    }
                }
                .estilo(.label).foregroundStyle(C.texto2).padding(.horizontal, 4).padding(.top, 2)
                ScrollView {
                    if lineas.isEmpty {
                        Text(Formato.sinViudas("Escaneá un código o buscá un producto para empezar el ticket."))
                            .estilo(.subhead).foregroundStyle(C.texto2).multilineTextAlignment(.center)
                            .padding(.top, 24).padding(.horizontal, 20)
                    } else {
                        VStack(spacing: 0) {
                            ForEach(Array(lineas.enumerated()), id: \.element.id) { i, l in
                                if i > 0 { Separador(inset: 14) }
                                fila(l)
                            }
                        }
                        .tarjeta()
                    }
                }
                .scrollDismissesKeyboard(.immediately)
            }
            .padding(.horizontal, 20).padding(.top, 14)
            .background(C.fondo, in: UnevenRoundedRectangle(topLeadingRadius: 22, topTrailingRadius: 22, style: .continuous))
            .padding(.top, -20)
        }
        .safeAreaInset(edge: .bottom, spacing: 0) {
            Button { paso = .cobrar } label: {
                HStack {
                    Text("Cobrar")
                    Spacer()
                    Text(Formato.pesos(total)).font(Estilo(tamano: 20, peso: .bold, relativo: .headline).font).monospacedDigit()
                }
                .padding(.horizontal, 20)
            }
            .buttonStyle(BotonPrincipal(alto: 56))
            .disabled(lineas.isEmpty)
            .padding(.horizontal, 20).padding(.top, 12).padding(.bottom, 8)
            .background {
                Color.tema(0xFBF7F2, 0x22140D, alfa: 0.94)
                    .overlay(alignment: .top) { Rectangle().fill(C.linea2).frame(height: 0.5) }
                    .ignoresSafeArea(edges: .bottom)
            }
        }
        .task(id: busqueda) { await buscar() }
        .confirmationDialog("¿Cuál?", isPresented: Binding(get: { elegirVariante != nil }, set: { if !$0 { elegirVariante = nil } }),
                            presenting: elegirVariante) { p in
            ForEach(p.variantes.filter(\.activa)) { v in
                Button("\(v.texto) · \(v.stock)\(Formato.fino)u.") { sumar(p, v) }
            }
        }
    }

    private var camara: some View {
        ZStack {
            // Visor de esquinas y línea de escaneo.
            Visor().frame(width: 260, height: 116)
            VStack {
                HStack {
                    BotonCamara(simbolo: "xmark", etiqueta: "Cerrar") { cerrar() }
                    Spacer()
                    Text("Nueva venta").estilo(.headline).foregroundStyle(Color(hex: 0xFBF7F2))
                    Spacer()
                    BotonCamara(simbolo: linterna ? "bolt.fill" : "bolt", etiqueta: "Linterna") { linterna.toggle() }
                }
                .padding(.horizontal, 16).padding(.top, 6)
                Spacer()
                Text("Apuntá al código de barras").estilo(.texto14).foregroundStyle(Color(hex: 0xFBF7F2, alfa: 0.85))
                    .padding(.bottom, 36)
            }
        }
        .frame(height: 250)
        .frame(maxWidth: .infinity)
        // La cámara llega hasta arriba de todo, pero los botones quedan debajo de la barra de estado
        // (antes iban debajo de la hora y la X no se podía tocar).
        .background {
            ZStack {
                Color.black
                Escaner(linterna: linterna) { codigo in Task { await escaneado(codigo) } }
            }
            .clipped()
            .ignoresSafeArea(edges: .top)
        }
    }

    private func fila(_ l: LineaTicket) -> some View {
        HStack(spacing: 12) {
            VStack(alignment: .leading, spacing: 2) {
                Text(l.nombre).estilo(.fila).foregroundStyle(C.texto).lineLimit(2)
                Text(l.id == recien ? "\(Formato.pesos(l.precio)) · recién escaneado"
                     : (l.cantidad > 1 ? "\(Formato.pesos(l.precio)) c/u · \(Formato.pesos(l.precio * l.cantidad))" : Formato.pesos(l.precio)))
                    .estilo(.footnote).monospacedDigit().foregroundStyle(C.texto2)
            }
            Spacer(minLength: 6)
            HStack(spacing: 0) {
                Button { cambiar(l, -1) } label: { Text("−").frame(width: 34, height: 34) }
                Text("\(l.cantidad)").font(Estilo(tamano: 16, peso: .bold, relativo: .body).font).monospacedDigit().frame(minWidth: 18)
                Button { cambiar(l, 1) } label: { Text("+").frame(width: 34, height: 34) }
            }
            .font(.system(size: 20)).foregroundStyle(C.primarioTexto)
            .background(C.pista, in: Capsule())
            .buttonStyle(.plain)
        }
        .padding(.horizontal, 14).padding(.vertical, 11)
        .background(l.id == recien ? Color.tema(0xFFFBF6, 0x3A2216) : .clear)
        .accessibilityElement(children: .combine)
        .accessibilityAdjustableAction { d in cambiar(l, d == .increment ? 1 : -1) }
    }

    private func cambiar(_ l: LineaTicket, _ d: Int) {
        guard let i = lineas.firstIndex(of: l) else { return }
        Haptico.seleccion()
        withAnimation(.snappy) {
            lineas[i].cantidad += d
            if lineas[i].cantidad <= 0 { lineas.remove(at: i) }
        }
    }

    private func agregar(_ p: ProductoFila) {
        busqueda = ""
        buscando = false
        if p.variantes.contains(where: \.activa) { elegirVariante = p } else { sumar(p, nil) }
    }

    private func sumar(_ p: ProductoFila, _ v: ProductoFila.Variante?) {
        Haptico.seleccion()
        let nueva = LineaTicket(producto: p, variante: v, cantidad: 1)
        withAnimation(.snappy) {
            if let i = lineas.firstIndex(where: { $0.id == nueva.id }) {
                lineas[i].cantidad += 1
            } else {
                lineas.append(nueva)
            }
            recien = nueva.id
        }
    }

    private func buscar() async {
        let q = busqueda.trimmingCharacters(in: .whitespaces)
        guard !q.isEmpty else { resultados = []; return }
        try? await Task.sleep(for: .milliseconds(250))
        guard !Task.isCancelled else { return }
        if let l: PaginaProductos = try? await estado.api.get("/v1/panel/listas/productos", query: [
            URLQueryItem(name: "q", value: q), URLQueryItem(name: "cuantos", value: "20"), URLQueryItem(name: "seg", value: "active,out"),
        ]) {
            resultados = l.filas
        }
    }

    private func escaneado(_ codigo: String) async {
        guard let l: PaginaProductos = try? await estado.api.get("/v1/panel/listas/productos", query: [
            URLQueryItem(name: "q", value: codigo), URLQueryItem(name: "cuantos", value: "5"),
        ]) else { return }
        // El SKU exacto primero (del producto o de una variante).
        if let p = l.filas.first(where: { $0.sku == codigo }) {
            AudioServicesPlaySystemSound(1057)
            sumar(p, nil)
        } else if let p = l.filas.first(where: { $0.variantes.contains { $0.sku == codigo } }),
                  let v = p.variantes.first(where: { $0.sku == codigo }) {
            AudioServicesPlaySystemSound(1057)
            sumar(p, v)
        } else if l.filas.count == 1, let p = l.filas.first {
            AudioServicesPlaySystemSound(1057)
            agregar(p)
        } else {
            Haptico.error()
            estado.avisar("No encontramos el código \(codigo)", ok: false)
        }
    }
}


/// Botón redondo sobre la cámara (34, blanco al 16 %).
struct BotonCamara: View {
    var simbolo: String
    var etiqueta: String
    var accion: () -> Void
    var body: some View {
        Button(action: accion) {
            Image(systemName: simbolo).font(.system(size: 14, weight: .bold)).foregroundStyle(Color(hex: 0xFBF7F2))
                .frame(width: 34, height: 34)
                .background(Color(hex: 0xFBF7F2, alfa: 0.16), in: Circle())
                .frame(width: 44, height: 44)
        }
        .accessibilityLabel(etiqueta)
    }
}

/// El visor: esquinas blancas de 4 y la línea de escaneo `#E07A48` con brillo.
struct Visor: View {
    @Environment(\.accessibilityReduceMotion) private var reducir
    @State private var y: CGFloat = 0.2
    var body: some View {
        GeometryReader { g in
            ZStack {
                Path { p in
                    let l: CGFloat = 26
                    let w = g.size.width, h = g.size.height
                    p.move(to: CGPoint(x: 0, y: l)); p.addLine(to: .zero); p.addLine(to: CGPoint(x: l, y: 0))
                    p.move(to: CGPoint(x: w - l, y: 0)); p.addLine(to: CGPoint(x: w, y: 0)); p.addLine(to: CGPoint(x: w, y: l))
                    p.move(to: CGPoint(x: w, y: h - l)); p.addLine(to: CGPoint(x: w, y: h)); p.addLine(to: CGPoint(x: w - l, y: h))
                    p.move(to: CGPoint(x: l, y: h)); p.addLine(to: CGPoint(x: 0, y: h)); p.addLine(to: CGPoint(x: 0, y: h - l))
                }
                .stroke(.white, style: StrokeStyle(lineWidth: 4, lineCap: .round, lineJoin: .round))
                Capsule().fill(Color(hex: 0xE07A48)).frame(height: 2)
                    .shadow(color: Color(hex: 0xE07A48), radius: 6)
                    .padding(.horizontal, 12)
                    .position(x: g.size.width / 2, y: g.size.height * y)
            }
        }
        .onAppear {
            guard !reducir else { y = 0.5; return }
            withAnimation(.easeInOut(duration: 1.6).repeatForever()) { y = 0.8 }
        }
        .accessibilityHidden(true)
    }
}
