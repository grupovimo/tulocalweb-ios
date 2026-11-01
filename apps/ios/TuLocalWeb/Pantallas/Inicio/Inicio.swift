import SwiftUI

/// 2a · Inicio: cómo viene el día (ventas con gráfico, indicadores), lo que hay para hacer hoy, la
/// meta del mes, ideas de Tincho, últimos pedidos, lo más vendido y canales. Con /v1/panel/tablero.
struct Inicio: View {
    @Environment(Estado.self) private var estado
    @Environment(Navegacion.self) private var nav
    @Environment(\.margen) private var margen
    @Environment(\.anchoPantalla) private var ancho
    @State private var periodo = "hoy"
    @State private var carga: Carga<Tablero> = .cargando
    @State private var masViejo: Date?
    @State private var sinLeer = 0
    @State private var editarMeta = false
    @State private var ideasOcultas: Set<String> = []

    private var se: Bool { ancho <= 375 }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 0) {
                encabezado
                VStack(alignment: .leading, spacing: 14) {
                    Segmentado(opciones: [("hoy", "Hoy"), ("7d", "7 días"), ("30d", "30 días")], elegido: $periodo)
                    switch carga {
                    case .cargando:
                        esqueleto
                    case let .error(e):
                        EstadoError(que: "cómo viene el día", error: e) { Task { await cargar() } }
                            .padding(.top, 30)
                    case let .listo(t):
                        contenido(t)
                    }
                }
                .padding(.top, 18)
                .margenPantalla()
            }
            .padding(.bottom, 90)
        }
        .scrollIndicators(.hidden)
        .background(C.fondo.ignoresSafeArea())
        .toolbar(.hidden, for: .navigationBar)
        .navigationTitle("Inicio")
        .refreshable { await cargar() }
        .task(id: "\(periodo)-\(estado.version)") { await cargar() }
        .sheet(isPresented: $editarMeta) {
            if let t = carga.valor {
                EditarMeta(actual: t.meta?.objetivo ?? 0, vendido: t.meta?.actual ?? 0) { await cargar() }
                    .hojaTLW([.height(470)])
            }
        }
    }

    // MARK: Encabezado

    private var encabezado: some View {
        HStack(alignment: .bottom) {
            VStack(alignment: .leading, spacing: 2) {
                Text(Formato.fechaLarga(Date())).estilo(.subhead).foregroundStyle(C.texto2)
                Text("Hola, \(carga.valor?.hola ?? estado.datos?.nombreDePila ?? "")")
                    .estilo(se ? Estilo(tamano: 28, peso: .bold, tracking: -0.025, relativo: .title) : .saludo)
                    .foregroundStyle(C.texto)
                    .lineLimit(1).minimumScaleFactor(0.8)
                    .accessibilityAddTraits(.isHeader)
            }
            Spacer(minLength: 8)
            HStack(spacing: 10) {
                BotonRedondo(simbolo: "bell", etiqueta: sinLeer > 0 ? "Avisos, \(sinLeer) sin leer" : "Avisos", punto: sinLeer > 0) {
                    nav.inicio.append(.avisos)
                }
                .frame(width: 40, height: 40)
                Button { nav.tocar(.mas) } label: {
                    AvatarIniciales(nombre: estado.tienda?.nombre ?? "", tam: 40)
                        .font(Estilo(tamano: 14, peso: .bold, relativo: .body).font)
                }
                .buttonStyle(.plain)
                .accessibilityLabel("Tu tienda, \(estado.tienda?.nombre ?? "")")
            }
            .padding(.bottom, 2)
        }
        .padding(.top, 6)
        .margenPantalla()
    }

    // MARK: Contenido

    @ViewBuilder
    private func contenido(_ t: Tablero) -> some View {
        if t.ultimosPedidos.isEmpty && t.masVendidos.isEmpty && (t.ventas?.actual ?? 0) == 0 && periodo == "hoy" {
            SinVentas()
        } else {
            conVentas(t)
        }
    }

    @ViewBuilder
    private func conVentas(_ t: Tablero) -> some View {
        if t.montos, let v = t.ventas {
            tarjetaVentas(t, v)
        }
        indicadores(t)
        paraHacer(t)
        if t.montos, let m = t.meta { meta(m) }
        ideas(t)
        if !t.ultimosPedidos.isEmpty { ultimos(t) }
        if !t.masVendidos.isEmpty { masVendidos(t) }
        if t.montos, let canales = t.canales, !canales.isEmpty, let v = t.ventas, v.actual > 0 {
            canalesVista(canales, total: v.actual)
        }
    }

    private var textoComparacion: String {
        switch periodo {
        case "hoy": return "vs. ayer a esta hora"
        case "7d": return "vs. los 7 días anteriores"
        default: return "vs. los 30 días anteriores"
        }
    }

    private func tarjetaVentas(_ t: Tablero, _ v: Comparado) -> some View {
        VStack(alignment: .leading, spacing: 4) {
            Text(periodo == "hoy" ? "Ventas de hoy" : (periodo == "7d" ? "Ventas de 7 días" : "Ventas de 30 días"))
                .estilo(.subheadFuerte).foregroundStyle(C.texto2)
            Text(Formato.pesos(v.actual))
                .estilo(se ? Estilo(tamano: 38, peso: .bold, tracking: -0.035, relativo: .largeTitle) : .montoHeroe)
                .monospacedDigit().foregroundStyle(C.texto)
                .lineLimit(1).minimumScaleFactor(0.6)
            HStack(spacing: 8) {
                if let d = v.variacion {
                    Delta(variacion: d)
                } else if v.actual > 0 {
                    Text("Primeras ventas").estilo(.texto14).fontWeight(.bold)
                        .padding(.horizontal, 8).padding(.vertical, 3)
                        .background(C.okFondo, in: Capsule()).foregroundStyle(C.okTinta)
                }
                Text(textoComparacion).estilo(.texto14).foregroundStyle(C.texto2)
            }
            if !se, let s = t.serie {
                GraficoVentas(actual: s.actual, anterior: s.anterior)
                    .padding(.top, 12).padding(.horizontal, -2)
                ejes(t)
                HStack(spacing: 16) {
                    HStack(spacing: 6) {
                        Capsule().fill(C.acento).frame(width: 14, height: 3)
                        Text(periodo == "hoy" ? "Hoy" : "Este período")
                    }
                    HStack(spacing: 6) {
                        Punteado().frame(width: 14, height: 2)
                        Text("\(periodo == "hoy" ? "Ayer" : "Anterior") · \(Formato.pesos(v.anterior))")
                    }
                }
                .estilo(.footnote).foregroundStyle(C.texto2).monospacedDigit()
                .padding(.top, 8)
            }
        }
        .padding(.horizontal, 18).padding(.top, 18).padding(.bottom, 14)
        .frame(maxWidth: .infinity, alignment: .leading)
        .tarjeta(radio: 22)
        .accessibilityElement(children: .combine)
    }

    private func ejes(_ t: Tablero) -> some View {
        let total = t.hasta.timeIntervalSince(t.desde)
        let marcas = (0..<4).map { i in t.desde.addingTimeInterval(total * (Double(i) * 2 + 1) / 8) }
        return HStack {
            ForEach(Array(marcas.enumerated()), id: \.offset) { i, f in
                if i > 0 { Spacer() }
                Text(periodo == "hoy" ? "\(Formato.calendario.component(.hour, from: f))\(Formato.fino)h" : diaCorto(f))
            }
        }
        .estilo(.chip).fontWeight(.regular).monospacedDigit().foregroundStyle(C.texto2)
        .padding(.horizontal, 2).padding(.top, 2)
    }

    private func diaCorto(_ f: Date) -> String {
        let c = Formato.calendario
        return "\(c.component(.day, from: f))/\(c.component(.month, from: f))"
    }

    private func indicadores(_ t: Tablero) -> some View {
        let comp = periodo == "hoy" ? "vs. ayer" : "vs. antes"
        return LazyVGrid(columns: Array(repeating: GridItem(.flexible(), spacing: 10), count: 3), spacing: 10) {
            Indicador(titulo: "Pedidos", valor: Formato.numero(Int(t.kpis.pedidos.actual)),
                      delta: diferencia(t.kpis.pedidos, comp), positivo: t.kpis.pedidos.actual >= t.kpis.pedidos.anterior)
            if t.montos, let tk = t.kpis.ticket {
                Indicador(titulo: "Ticket prom.", valor: Formato.pesos(tk.actual),
                          delta: tk.variacion.map { porcentajeConSigno($0) }, positivo: tk.actual >= tk.anterior)
            }
            Indicador(titulo: "Clientes nuevos", valor: Formato.numero(Int(t.kpis.clientesNuevos.actual)),
                      delta: diferencia(t.kpis.clientesNuevos, comp), positivo: t.kpis.clientesNuevos.actual >= t.kpis.clientesNuevos.anterior)
        }
    }

    private func diferencia(_ c: Comparado, _ comp: String) -> String? {
        let d = Int(c.actual - c.anterior)
        if d == 0 { return "igual \(comp)" }
        return "\(d > 0 ? "+" : "−")\(abs(d)) \(comp)"
    }

    private func porcentajeConSigno(_ p: Double) -> String {
        let r = (p * 10).rounded() / 10
        return "\(r >= 0 ? "+" : "−")\(Formato.porcentaje(abs(r)))"
    }

    // MARK: Para hacer hoy

    private func filasPendientes(_ ph: Tablero.ParaHacer) -> [AnyView] {
        var filas: [AnyView] = []
        if ph.preparar > 0 {
            filas.append(AnyView(FilaPendiente(n: ph.preparar, fondo: C.avisoFondo, tinta: C.avisoTinta,
                                               titulo: "Pedidos por preparar",
                                               detalle: masViejo.map { "El más viejo entró \(Formato.hace($0))" } ?? "Tocá para verlos") {
                nav.filtroPedidos = "Para preparar"
                nav.tocar(.pedidos)
            }))
        }
        if ph.transferencias > 0 {
            filas.append(AnyView(FilaPendiente(n: ph.transferencias, fondo: C.crema, tinta: C.primarioTinta,
                                               titulo: "Transferencias por confirmar",
                                               detalle: ph.transferenciasMonto.map { "\(Formato.pesos($0)) en total" } ?? "Revisá que hayan llegado") {
                nav.filtroPedidos = "Sin pagar"
                nav.tocar(.pedidos)
            }))
        }
        if ph.sinStock > 0 {
            filas.append(AnyView(FilaPendiente(n: ph.sinStock, fondo: C.errorFondo, tinta: C.errorTinta,
                                               titulo: "Productos sin stock",
                                               detalle: "Siguen visibles en la tienda") {
                nav.segmentoProductos = "out"
                nav.tocar(.productos)
            }))
        }
        return filas
    }

    @ViewBuilder
    private func paraHacer(_ t: Tablero) -> some View {
        let filas = filasPendientes(t.paraHacer)
        TituloSeccion(texto: "Para hacer hoy").padding(.top, 14)
        VStack(spacing: 0) {
            if filas.isEmpty {
                HStack(spacing: 13) {
                    Image(systemName: "checkmark").font(.system(size: 16, weight: .bold)).foregroundStyle(C.okTinta)
                        .frame(width: 40, height: 40).background(C.okFondo, in: RoundedRectangle(cornerRadius: 12, style: .continuous))
                    VStack(alignment: .leading, spacing: 1) {
                        Text("Todo al día").estilo(.fila).foregroundStyle(C.texto)
                        Text("No hay pedidos ni transferencias esperando.").estilo(.footnote).foregroundStyle(C.texto2)
                    }
                    Spacer(minLength: 0)
                }
                .padding(.horizontal, 16).padding(.vertical, 13)
            } else {
                ForEach(Array(filas.enumerated()), id: \.offset) { i, f in
                    if i > 0 { Separador(inset: 69) }
                    f
                }
            }
        }
        .tarjeta()
    }

    // MARK: Meta

    private func meta(_ m: Tablero.Meta) -> some View {
        let hoy = Date()
        let c = Formato.calendario
        let diasMes = c.range(of: .day, in: .month, for: hoy)?.count ?? 30
        let quedan = diasMes - c.component(.day, from: hoy) + 1
        let pct = m.objetivo > 0 ? Double(m.actual) / Double(m.objetivo) : 0
        let proy = m.objetivo > 0 ? Double(m.proyeccion) / Double(m.objetivo) : 0
        return VStack(alignment: .leading, spacing: 10) {
            HStack(alignment: .firstTextBaseline) {
                Text("Meta de \(Formato.mesDe(hoy))").estilo(.subheadFuerte).foregroundStyle(C.texto2)
                Spacer()
                Button(m.objetivo > 0 ? "Editar" : "Ponerla") { editarMeta = true }
                    .estilo(.subheadFuerte).foregroundStyle(C.primarioTexto)
            }
            if m.objetivo > 0 {
                (Text(Formato.pesos(m.actual)).estiloTexto(.montoMeta).foregroundColor(C.texto)
                    + Text("  de \(Formato.pesos(m.objetivo))").estiloTexto(.subhead).foregroundColor(C.texto2))
                    .monospacedDigit()
                    .lineLimit(1).minimumScaleFactor(0.7)
                GeometryReader { g in
                    ZStack(alignment: .leading) {
                        Capsule().fill(C.pistaMeta)
                        RayadoMeta().frame(width: g.size.width * min(1, proy))
                        Capsule().fill(LinearGradient(colors: [C.acento, Color(hex: 0xB0441D)], startPoint: .leading, endPoint: .trailing))
                            .frame(width: max(12, g.size.width * min(1, pct)))
                    }
                    .clipShape(Capsule())
                }
                .frame(height: 12)
                HStack {
                    (Text(Formato.porcentaje((pct * 100).rounded())).fontWeight(.bold).foregroundColor(C.texto) + Text(" vendido"))
                    Spacer()
                    Text(quedan == 1 ? "Queda 1 día" : "Quedan \(quedan) días")
                }
                .estilo(.footnote).foregroundStyle(C.texto2).monospacedDigit()
                (Text("Al ritmo de estos días cerrás el mes en ")
                    + Text(Formato.pesos(m.proyeccion)).fontWeight(.bold).foregroundColor(C.texto) + Text("."))
                    .estilo(.texto14).foregroundStyle(C.texto2).lineSpacing(2)
            } else {
                Text(Formato.sinViudas("Poné cuánto querés vender este mes y te decimos cómo venís cada día."))
                    .estilo(.texto14).foregroundStyle(C.texto2)
            }
        }
        .padding(18)
        .tarjeta(radio: 22)
        .padding(.top, 6)
    }

    // MARK: Ideas de Tincho

    @ViewBuilder
    private func ideas(_ t: Tablero) -> some View {
        let lista = ideasDe(t).filter { !ideasOcultas.contains($0.id) }
        if !lista.isEmpty {
            HStack(spacing: 8) {
                Image(Tincho.avatar.imagen).resizable().scaledToFill().frame(width: 28, height: 28)
                    .background(C.crema).clipShape(Circle())
                Text("Ideas de Tincho").estilo(.tituloSeccion).foregroundStyle(C.texto)
            }
            .padding(.top, 14)
            ScrollView(.horizontal) {
                HStack(spacing: 10) {
                    ForEach(lista) { idea in
                        VStack(alignment: .leading, spacing: 12) {
                            Text(idea.texto).estilo(.tarjetaTincho).foregroundStyle(C.texto).lineSpacing(2)
                                .fixedSize(horizontal: false, vertical: true)
                            Spacer(minLength: 0)
                            HStack(spacing: 8) {
                                Button(idea.boton) { Task { await idea.accion() } }
                                    .font(Estilo.subheadFuerte.font).foregroundStyle(.white)
                                    .padding(.horizontal, 16).frame(height: 38)
                                    .background(C.primario, in: RoundedRectangle(cornerRadius: 11, style: .continuous))
                                Button("Ahora no") { withAnimation { _ = ideasOcultas.insert(idea.id) } }
                                    .font(Estilo.subheadFuerte.font).foregroundStyle(C.primarioTexto)
                                    .padding(.horizontal, 12).frame(height: 38)
                            }
                        }
                        .padding(16)
                        .frame(width: min(300, ancho - 2 * margen - 30), alignment: .leading)
                        .frame(maxHeight: .infinity, alignment: .top)
                        .background(C.crema2, in: RoundedRectangle(cornerRadius: 20, style: .continuous))
                        .overlay(RoundedRectangle(cornerRadius: 20, style: .continuous).strokeBorder(Color(hex: 0xB0441D, alfa: 0.16), lineWidth: 1))
                    }
                }
                .padding(.horizontal, margen)
            }
            .scrollIndicators(.hidden)
            .padding(.horizontal, -margen)
        }
    }

    private struct Idea: Identifiable {
        var id: String
        var texto: AttributedString
        var boton: String
        var accion: () async -> Void
    }

    private func ideasDe(_ t: Tablero) -> [Idea] {
        var r: [Idea] = []
        if t.hornero.carritos > 0, !t.hornero.ideasHechas.contains("carritos") {
            var s = AttributedString(t.hornero.carritos == 1 ? "Hay 1 carrito abandonado" : "Hay \(t.hornero.carritos) carritos abandonados")
            if let m = t.hornero.carritosMonto, t.montos {
                s += AttributedString(" por ")
                var monto = AttributedString(Formato.pesos(m))
                monto.font = Estilo(tamano: 16, peso: .bold, relativo: .body).font
                s += monto
            }
            s += AttributedString(". ¿Les escribo para recordarles?")
            r.append(Idea(id: "carritos", texto: s, boton: "Escribiles") { await recordarCarritos() })
        }
        if t.paraHacer.sinStock > 0, !t.hornero.ideasHechas.contains("stock") {
            let n = t.paraHacer.sinStock
            let s = AttributedString(n == 1
                ? "Un producto se quedó sin stock y sigue visible en la tienda. ¿Le cargás un ingreso?"
                : "\(n) productos se quedaron sin stock y siguen visibles en la tienda. ¿Les cargás un ingreso?")
            r.append(Idea(id: "stock", texto: s, boton: "Cargar") {
                nav.segmentoProductos = "out"
                nav.tocar(.productos)
            })
        }
        return r
    }

    private func recordarCarritos() async {
        struct R: Decodable { var enviados: Int }
        do {
            let r: R = try await estado.api.enviar("POST", "/v1/panel/carritos/recordar", [String: String]())
            Haptico.exito()
            estado.avisar(r.enviados == 1 ? "Le escribimos a 1 cliente" : "Les escribimos a \(r.enviados) clientes")
            ideasOcultas.insert("carritos")
        } catch let e as ErrorApi {
            estado.avisar(e.detalle, ok: false)
        } catch {}
    }

    // MARK: Últimos pedidos

    private func ultimos(_ t: Tablero) -> some View {
        VStack(alignment: .leading, spacing: 14) {
            TituloSeccion(texto: "Últimos pedidos", accion: "Ver todos") {
                nav.filtroPedidos = nil
                nav.tocar(.pedidos)
            }
            .padding(.top, 14)
            VStack(spacing: 0) {
                ForEach(Array(t.ultimosPedidos.prefix(5).enumerated()), id: \.element.id) { i, p in
                    if i > 0 { Separador(inset: 68) }
                    Button {
                        nav.pestana = .pedidos
                        nav.pedidos = [.pedido(p.numero)]
                    } label: {
                        HStack(spacing: 12) {
                            Miniatura(foto: p.foto, tam: 42, radio: 11)
                            VStack(alignment: .leading, spacing: 2) {
                                Text(p.cliente).estilo(.fila).foregroundStyle(C.texto).lineLimit(1)
                                Text("#\(p.numero) · \(Marcas.canalCorto(p.canal)) · \(Formato.horaDe(p.creado))")
                                    .estilo(.footnote).foregroundStyle(C.texto2).lineLimit(1)
                            }
                            Spacer(minLength: 4)
                            VStack(alignment: .trailing, spacing: 4) {
                                if let total = p.total {
                                    Text(Formato.pesos(total)).estilo(.fila).monospacedDigit().foregroundStyle(C.texto)
                                }
                                ChipEstado(estado: p.estado)
                            }
                        }
                        .padding(.horizontal, 14).padding(.vertical, 12)
                        .contentShape(Rectangle())
                    }
                    .buttonStyle(.plain)
                }
            }
            .tarjeta()
        }
    }

    // MARK: Lo más vendido

    private func masVendidos(_ t: Tablero) -> some View {
        VStack(alignment: .leading, spacing: 14) {
            HStack(alignment: .firstTextBaseline) {
                Text("Lo más vendido").estilo(.tituloSeccion).foregroundStyle(C.texto)
                Spacer()
                Text("Esta semana").estilo(.subhead).foregroundStyle(C.texto2)
            }
            .padding(.top, 14)
            ScrollView(.horizontal) {
                HStack(alignment: .top, spacing: 10) {
                    ForEach(t.masVendidos) { p in
                        VStack(alignment: .leading, spacing: 7) {
                            Miniatura(foto: p.foto, tam: 136, alto: 120, radio: 16)
                            Text(p.nombre).estilo(Estilo(tamano: 14, peso: .semibold, relativo: .footnote))
                                .foregroundStyle(C.texto).lineLimit(2).lineSpacing(1)
                            Text(Formato.cantidad(p.unidades, "unidad", "unidades")).estilo(.footnote).monospacedDigit().foregroundStyle(C.texto2)
                        }
                        .frame(width: 136, alignment: .leading)
                    }
                }
                .padding(.horizontal, margen)
            }
            .scrollIndicators(.hidden)
            .padding(.horizontal, -margen)
        }
    }

    // MARK: Canales

    private func canalesVista(_ canales: [Tablero.Canal], total: Double) -> some View {
        VStack(alignment: .leading, spacing: 14) {
            Text("Canales").estilo(.tituloSeccion).foregroundStyle(C.texto).padding(.top, 14)
            VStack(alignment: .leading, spacing: 14) {
                GeometryReader { g in
                    HStack(spacing: 2) {
                        ForEach(canales, id: \.canal) { c in
                            Rectangle().fill(Marcas.colorCanal(c.canal))
                                .frame(width: max(2, (g.size.width - CGFloat(canales.count - 1) * 2) * c.parte))
                        }
                    }
                }
                .frame(height: 14)
                .clipShape(RoundedRectangle(cornerRadius: 7, style: .continuous))
                VStack(spacing: 9) {
                    ForEach(canales, id: \.canal) { c in
                        HStack(spacing: 9) {
                            RoundedRectangle(cornerRadius: 3).fill(Marcas.colorCanal(c.canal)).frame(width: 10, height: 10)
                            Text(Marcas.canal(c.canal)).frame(maxWidth: .infinity, alignment: .leading)
                            Text(Formato.pesos(total * c.parte)).foregroundStyle(C.texto2)
                            Text(Formato.porcentaje((c.parte * 100).rounded())).fontWeight(.bold).frame(width: 52, alignment: .trailing)
                        }
                        .estilo(.subhead).monospacedDigit().foregroundStyle(C.texto)
                    }
                }
            }
            .padding(16)
            .tarjeta()
        }
    }

    // MARK: Cargando (2c)

    private var esqueleto: some View {
        VStack(alignment: .leading, spacing: 14) {
            VStack(alignment: .leading, spacing: 12) {
                Esqueleto(ancho: 110, alto: 14)
                Esqueleto(ancho: 220, alto: 42, radio: 10)
                Esqueleto(ancho: 160, alto: 20)
                Esqueleto(alto: 130, radio: 14, suave: true)
            }
            .padding(18).tarjeta(radio: 22)
            HStack(spacing: 10) {
                ForEach(0..<3, id: \.self) { _ in Color.clear.frame(height: 84).tarjeta(radio: 18) }
            }
            Esqueleto(ancho: 150, alto: 20).padding(.top, 14)
            EsqueletoFilas(filas: 2, miniatura: 40)
        }
    }

    // MARK: Datos

    private func cargar() async {
        do {
            let t: Tablero = try await estado.api.get("/v1/panel/tablero", query: [URLQueryItem(name: "periodo", value: periodo)])
            carga = .listo(t)
            Guardado.guardar(t, clave: "tablero-\(periodo)")
            if t.paraHacer.preparar > 0 {
                let l: PaginaPedidos? = try? await estado.api.get("/v1/panel/pedidos", query: [
                    URLQueryItem(name: "estado", value: "Para preparar"), URLQueryItem(name: "orden", value: "viejos"),
                    URLQueryItem(name: "cuantos", value: "1"),
                ])
                masViejo = l?.filas.first?.creado
            }
            struct SinLeer: Decodable { var sinLeer: Int }
            if let s: SinLeer = try? await estado.api.get("/v1/panel/notificaciones/sin-leer") { sinLeer = s.sinLeer }
        } catch is CancellationError {
        } catch let e as ErrorApi {
            if e.sinConexion, let (t, _) = Guardado.leer(Tablero.self, clave: "tablero-\(periodo)") {
                carga = .listo(t)
            } else if carga.valor == nil {
                carga = .error(e)
            }
        } catch {
            carga = .error(.desconocido)
        }
    }
}

/// Chip de variación: `ok-bg` «▲ 12,4 %» o `error-bg` «▼ 3 %».
struct Delta: View {
    var variacion: Double
    var body: some View {
        let sube = variacion >= 0
        let r = (abs(variacion) * 10).rounded() / 10
        Text("\(sube ? "▲" : "▼") \(Formato.porcentaje(r))")
            .estilo(.texto14).fontWeight(.bold).monospacedDigit()
            .foregroundStyle(sube ? C.okTinta : C.errorTinta)
            .padding(.horizontal, 8).padding(.vertical, 3)
            .background(sube ? C.okFondo : C.errorFondo, in: Capsule())
            .accessibilityLabel(sube ? "Subió \(Formato.porcentaje(r))" : "Bajó \(Formato.porcentaje(r))")
    }
}

/// Indicador chico (radio 18): título 13, valor 22/700 y delta 12/600 coloreado.
struct Indicador: View {
    var titulo: String
    var valor: String
    var delta: String?
    var positivo: Bool
    var body: some View {
        VStack(alignment: .leading, spacing: 2) {
            Text(titulo).estilo(.footnote).foregroundStyle(C.texto2).lineLimit(2)
            Text(valor).estilo(.indicador).monospacedDigit().foregroundStyle(C.texto)
                .lineLimit(1).minimumScaleFactor(0.6)
            if let delta {
                Text(delta).estilo(.chip).foregroundStyle(positivo ? C.okTinta : C.errorTinta).lineLimit(1).minimumScaleFactor(0.8)
            }
        }
        .padding(.horizontal, 13).padding(.vertical, 12)
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
        .tarjeta(radio: 18)
        .accessibilityElement(children: .combine)
    }
}

/// Fila de «Para hacer hoy» con el cuadro numérico de 40.
struct FilaPendiente: View {
    var n: Int
    var fondo: Color
    var tinta: Color
    var titulo: String
    var detalle: String
    var accion: () -> Void
    var body: some View {
        Button(action: accion) {
            HStack(spacing: 13) {
                Text("\(n)").font(Estilo(tamano: 18, peso: .bold, relativo: .headline).font).monospacedDigit()
                    .foregroundStyle(tinta)
                    .frame(width: 40, height: 40)
                    .background(fondo, in: RoundedRectangle(cornerRadius: 12, style: .continuous))
                VStack(alignment: .leading, spacing: 1) {
                    Text(titulo).estilo(.fila).foregroundStyle(C.texto)
                    Text(detalle).estilo(.footnote).monospacedDigit().foregroundStyle(C.texto2)
                }
                Spacer(minLength: 4)
                Chevron()
            }
            .padding(.horizontal, 16).padding(.vertical, 13)
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
    }
}

/// Línea punteada de la leyenda.
struct Punteado: View {
    var body: some View {
        Canvas { ctx, tam in
            var p = Path()
            p.move(to: CGPoint(x: 0, y: tam.height / 2))
            p.addLine(to: CGPoint(x: tam.width, y: tam.height / 2))
            ctx.stroke(p, with: .color(Color(hex: 0x8A7364)), style: StrokeStyle(lineWidth: 2, dash: [3, 2]))
        }
    }
}

/// La proyección rayada detrás de la barra de la meta.
struct RayadoMeta: View {
    var body: some View {
        Canvas { ctx, tam in
            ctx.fill(Path(CGRect(origin: .zero, size: tam)), with: .color(C.acento.opacity(0.15)))
            var x: CGFloat = -tam.height
            while x < tam.width {
                var p = Path()
                p.move(to: CGPoint(x: x, y: tam.height))
                p.addLine(to: CGPoint(x: x + tam.height, y: 0))
                p.addLine(to: CGPoint(x: x + tam.height + 4, y: 0))
                p.addLine(to: CGPoint(x: x + 4, y: tam.height))
                p.closeSubpath()
                ctx.fill(p, with: .color(C.acento.opacity(0.35)))
                x += 8 * 1.4142
            }
        }
    }
}

/// 2d · Sin ventas todavía: Tincho saludando, «Tu tienda ya está online» y los pasos (el hecho
/// tachado con check verde; el siguiente con «Cargar»).
struct SinVentas: View {
    @Environment(Estado.self) private var estado
    @Environment(Navegacion.self) private var nav
    @State private var conProductos: Bool?

    var body: some View {
        let mp = estado.resumen?.canales?.contains("mp") ?? false
        let url = URL(string: "https://\(estado.tienda?.direccion ?? "tulocalweb.com.ar")")!
        VStack(spacing: 8) {
            Image(Tincho.saludo.imagen).resizable().scaledToFit().frame(width: 150, height: 150).accessibilityHidden(true)
            Text("Tu tienda ya está online").estilo(.tituloHoja).foregroundStyle(C.texto).multilineTextAlignment(.center)
            Text(Formato.sinViudas("Falta la primera venta. Con estos pasos, la tenés cerca."))
                .estilo(.subhead).foregroundStyle(C.texto2).multilineTextAlignment(.center)
            VStack(spacing: 0) {
                paso(1, "Creaste la tienda", nil, hecho: true, boton: nil) {}
                Separador(inset: 57)
                paso(2, "Cargá tu primer producto", "Con una foto alcanza", hecho: conProductos == true,
                     boton: conProductos == false ? "Cargar" : nil) { nav.inicio.append(.nuevoProducto(nil)) }
                Separador(inset: 57)
                paso(3, "Conectá Mercado Pago", "Para cobrar con tarjeta y QR", hecho: mp,
                     boton: conProductos == true && !mp ? "Conectar" : nil) {
                    nav.inicio.append(.integraciones)
                }
                Separador(inset: 57)
                ShareLink(item: url) {
                    filaPaso(4, "Compartí el link", "Por WhatsApp o en tus redes", hecho: false, boton: nil)
                }
                .buttonStyle(.plain)
            }
            .tarjeta()
            .padding(.top, 8)
        }
        .padding(.horizontal, 20).padding(.top, 20).padding(.bottom, 18)
        .background(C.crema2, in: RoundedRectangle(cornerRadius: 24, style: .continuous))
        .overlay(RoundedRectangle(cornerRadius: 24, style: .continuous).strokeBorder(Color(hex: 0xB0441D, alfa: 0.14), lineWidth: 1))
        .task {
            if let l: PaginaProductos = try? await estado.api.get("/v1/panel/listas/productos", query: [URLQueryItem(name: "cuantos", value: "1")]) {
                conProductos = l.total > 0
            }
        }
    }

    private func paso(_ n: Int, _ t: String, _ d: String?, hecho: Bool, boton: String?, _ accion: @escaping () -> Void) -> some View {
        Button(action: accion) { filaPaso(n, t, d, hecho: hecho, boton: boton) }
            .buttonStyle(.plain).disabled(hecho)
    }

    private func filaPaso(_ n: Int, _ t: String, _ d: String?, hecho: Bool, boton: String?) -> some View {
        HStack(spacing: 13) {
            ZStack {
                if hecho {
                    Circle().fill(C.ok)
                    Image(systemName: "checkmark").font(.system(size: 12, weight: .heavy)).foregroundStyle(.white)
                } else {
                    Circle().strokeBorder(Color.tema(0x3A1D12, 0xFBF7F2, alfa: 0.2), lineWidth: 2)
                    Text("\(n)").font(Estilo(tamano: 13, peso: .bold, relativo: .footnote).font).foregroundStyle(C.texto2)
                }
            }
            .frame(width: 28, height: 28)
            VStack(alignment: .leading, spacing: 1) {
                if hecho {
                    Text(t).estilo(.body16).strikethrough().foregroundStyle(C.texto2)
                } else {
                    Text(t).estilo(.fila).foregroundStyle(C.texto)
                    if let d { Text(d).estilo(.footnote).foregroundStyle(C.texto2) }
                }
            }
            Spacer(minLength: 0)
            if let boton {
                Text(boton).estilo(.label).foregroundStyle(.white)
                    .padding(.horizontal, 14).frame(height: 34)
                    .background(C.primario, in: RoundedRectangle(cornerRadius: 10, style: .continuous))
            }
        }
        .padding(.horizontal, 16).padding(.vertical, 14)
        .contentShape(Rectangle())
    }
}
