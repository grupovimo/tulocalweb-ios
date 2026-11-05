import SwiftUI

// Las secciones de «Más» con pantalla propia (antes abrían el panel web): Caja del día, Carritos
// abandonados, Reseñas y preguntas, Devoluciones, Cupones, Finanzas, Facturación de ARCA, Plan,
// Integraciones y Usuarios y permisos. Todas con las piezas del handoff y la API del panel.

// MARK: - Armazón común

/// Pantalla de sección: título grande, carga, error y «tirar para actualizar».
struct PantallaSeccion<T, Contenido: View>: View {
    var titulo: String
    var que: String
    var cargar: () async throws -> T
    @ViewBuilder var contenido: (T, @escaping () async -> Void) -> Contenido
    @State private var carga: Carga<T> = .cargando

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 12) {
                Text(titulo).estilo(.tituloGrande).foregroundStyle(C.texto)
                    .accessibilityAddTraits(.isHeader)
                switch carga {
                case .cargando:
                    EsqueletoFilas(filas: 5, miniatura: 42).padding(.top, 12)
                case let .error(e):
                    EstadoError(que: que, error: e) { Task { await actualizar() } }.padding(.top, 60)
                case let .listo(d):
                    contenido(d, actualizar)
                }
            }
            .margenPantalla()
            .padding(.bottom, 24)
        }
        .background(C.fondo.ignoresSafeArea())
        .navigationBarTitleDisplayMode(.inline)
        .refreshable { await actualizar() }
        .task { await actualizar() }
    }

    private func actualizar() async {
        do {
            carga = .listo(try await cargar())
        } catch is CancellationError {
        } catch let e as ErrorApi {
            if carga.valor == nil { carga = .error(e) }
        } catch {
            if carga.valor == nil { carga = .error(.desconocido) }
        }
    }
}

/// Grupo de filas en una tarjeta, con separadores.
struct GrupoFilas<Elemento: Identifiable, Fila: View>: View {
    var elementos: [Elemento]
    var inset: CGFloat = 16
    @ViewBuilder var fila: (Elemento) -> Fila
    var body: some View {
        VStack(spacing: 0) {
            ForEach(Array(elementos.enumerated()), id: \.element.id) { i, e in
                if i > 0 { Separador(inset: inset) }
                fila(e)
            }
        }
        .tarjeta()
    }
}

/// Fila de dato: etiqueta a la izquierda, valor a la derecha.
struct FilaDato: View {
    var etiqueta: String
    var valor: String
    var color: Color = C.texto
    var body: some View {
        HStack {
            Text(etiqueta).estilo(.body16).foregroundStyle(C.texto2)
            Spacer(minLength: 8)
            Text(valor).estilo(.fila).monospacedDigit().foregroundStyle(color).multilineTextAlignment(.trailing)
        }
        .padding(.horizontal, 16).padding(.vertical, 13)
    }
}

/// Chip de estado chico con colores propios.
struct ChipSimple: View {
    var texto: String
    var fondo: Color
    var tinta: Color
    var body: some View {
        Text(texto).font(Estilo(tamano: 12, peso: .bold, relativo: .caption).font)
            .foregroundStyle(tinta)
            .padding(.horizontal, 9).padding(.vertical, 3)
            .background(fondo, in: Capsule())
    }
}

/// Grilla de indicadores de 2 columnas.
struct Indicadores: View {
    var items: [(String, String)]
    var body: some View {
        LazyVGrid(columns: [GridItem(.flexible(), spacing: 10), GridItem(.flexible())], spacing: 10) {
            ForEach(Array(items.enumerated()), id: \.offset) { _, i in
                Indicador(titulo: i.0, valor: i.1, delta: nil, positivo: true)
            }
        }
    }
}

private struct Vacio: Encodable {}

// MARK: - Caja del día

struct CajaDelDia: View {
    @Environment(Estado.self) private var estado
    @Environment(Navegacion.self) private var nav

    struct Tablero: Decodable {
        struct Caja: Decodable {
            struct Medio: Decodable, Identifiable { var medio: String; var monto: Int; var id: String { medio } }
            var abierta: Date?
            var total: Int
            var medios: [Medio]
        }
        struct Serie: Decodable { var actual: [Double]; var anterior: [Double] }
        var caja: Caja?
        var serie: Serie?
    }

    var body: some View {
        PantallaSeccion(titulo: "Caja del día", que: "la caja") {
            try await estado.api.get("/v1/panel/tablero", query: [URLQueryItem(name: "periodo", value: "hoy")]) as Tablero
        } contenido: { t, _ in
            let caja = t.caja
            TarjetaHeroe(etiqueta: "Vendido hoy en el local", valor: Formato.pesos(caja?.total ?? 0),
                         detalle: caja?.abierta.map { "Caja abierta desde las \(Formato.horaDe($0))" } ?? "La caja está cerrada",
                         tincho: .caja)
            if let s = t.serie, s.actual.contains(where: { $0 > 0 }) {
                TarjetaSeccion(titulo: "Cómo viene el día", simbolo: "chart.xyaxis.line") {
                    GraficoVentas(actual: s.actual, anterior: s.anterior, alto: 140)
                }
            }
            if let medios = caja?.medios, !medios.isEmpty {
                TarjetaSeccion(titulo: "Por medio de pago", simbolo: "creditcard") {
                    DonaMedios(partes: medios.map { .init(nombre: CajaDelDia.nombre($0.medio), monto: $0.monto) })
                }
            }
            Button { nav.venta = true } label: { Label("Vender en el local", systemImage: "barcode.viewfinder") }
                .buttonStyle(BotonPrincipal()).padding(.top, 4)
        }
    }

    static func nombre(_ medio: String) -> String {
        switch medio {
        case "efectivo": "Efectivo"
        case "tarjeta": "Tarjeta"
        case "mp": "Mercado Pago"
        case "modo": "MODO"
        case "transf": "Transferencia"
        default: Formato.mayuscula(medio)
        }
    }
}

// MARK: - Carritos abandonados

struct CarritosAbandonados: View {
    @Environment(Estado.self) private var estado

    struct Datos: Decodable {
        struct Carrito: Decodable, Identifiable {
            var id: String
            var nombre: String
            var foto: String?
            var desde: String
            var creado: Date
            var cantidad: Int
            var total: Int
            var recordado: Bool
        }
        var carritos: [Carrito]
        var abiertos: Int
        var monto: Int
        var recuperado: Int
        var tasa: Double
    }

    var body: some View {
        PantallaSeccion(titulo: "Carritos abandonados", que: "los carritos") {
            try await estado.api.get("/v1/panel/carritos") as Datos
        } contenido: { d, actualizar in
            TarjetaHeroe(etiqueta: "En juego en carritos abiertos", valor: Formato.pesos(d.monto),
                         detalle: Formato.cantidad(d.abiertos, "carrito para recordar", "carritos para recordar"), tincho: .whatsapp)
            TarjetaSeccion(titulo: "Lo que ya recuperaste", simbolo: "arrow.uturn.backward.circle") {
                HStack(alignment: .firstTextBaseline) {
                    Text(Formato.pesos(d.recuperado)).estilo(.indicador).monospacedDigit().foregroundStyle(C.okTinta)
                    Spacer()
                    Text("\(Formato.porcentaje(d.tasa > 1 ? d.tasa : d.tasa * 100)) recuperado").estilo(.footnoteFuerte).foregroundStyle(C.texto2)
                }
                BarraProgreso(valor: d.tasa > 1 ? d.tasa / 100 : d.tasa)
            }
            if d.carritos.isEmpty {
                EstadoTincho(tincho: .durmiendo, titulo: "No hay carritos abandonados",
                             texto: "Cuando alguien deja productos sin comprar, aparece acá para que le recuerdes.")
                    .padding(.top, 30)
            } else {
                GrupoFilas(elementos: d.carritos, inset: 70) { c in
                    HStack(spacing: 12) {
                        Miniatura(foto: c.foto, tam: 44)
                        VStack(alignment: .leading, spacing: 2) {
                            Text(c.nombre).estilo(.fila).foregroundStyle(C.texto).lineLimit(1)
                            Text("\(Formato.cantidad(c.cantidad, "producto", "productos")) · \(Formato.hace(c.creado)) · \(c.desde)")
                                .estilo(.footnote).foregroundStyle(C.texto2).lineLimit(1)
                        }
                        Spacer(minLength: 6)
                        VStack(alignment: .trailing, spacing: 6) {
                            Text(Formato.pesos(c.total)).estilo(.fila).monospacedDigit().foregroundStyle(C.texto)
                            if c.recordado {
                                ChipSimple(texto: "Recordado", fondo: C.okFondo, tinta: C.okTinta)
                            } else {
                                Button("Recordar") {
                                    Task {
                                        do {
                                            try await estado.api.hacer("POST", "/v1/panel/carritos/\(c.id)/recordar")
                                            estado.avisar("Le mandamos el recordatorio")
                                            await actualizar()
                                        } catch { estado.avisar("No pudimos mandar el recordatorio", ok: false) }
                                    }
                                }
                                .buttonStyle(BotonTonal(alto: 28))
                            }
                        }
                    }
                    .padding(.horizontal, 14).padding(.vertical, 12)
                }
            }
        }
    }
}

// MARK: - Reseñas y preguntas

struct ResenasYPreguntas: View {
    @Environment(Estado.self) private var estado
    @State private var vista = "preguntas"
    @State private var respondiendo: Pregunta?
    @State private var respondiendoResena: Resena?

    struct Pregunta: Decodable, Identifiable, Hashable {
        var id: String
        var producto: String
        var foto: String?
        var pregunta: String
        var creada: Date
        var sugerida: String
        var respuesta: String?
    }
    struct Resena: Decodable, Identifiable, Hashable {
        var id: String
        var autor: String
        var estrellas: Int
        var texto: String
        var origen: String
        var creada: Date
        var producto: String?
        var conCompra: Bool
        var oculta: Bool
        var respuesta: String?
    }
    struct Datos: Decodable {
        var preguntas: [Pregunta]
        var resenas: [Resena]
        var promedio: Double?
        var total: Int
        var porEstrellas: [Int]?
    }

    var body: some View {
        PantallaSeccion(titulo: "Reseñas y preguntas", que: "las reseñas") {
            try await estado.api.get("/v1/panel/resenas") as Datos
        } contenido: { d, actualizar in
            Segmentado(opciones: [("preguntas", "Preguntas"), ("resenas", "Reseñas")], elegido: $vista)
            if vista == "preguntas" {
                if d.preguntas.isEmpty {
                    EstadoTincho(tincho: .saludo, titulo: "No hay preguntas", texto: "Cuando te pregunten algo de un producto, lo respondés acá.")
                        .padding(.top, 30)
                } else {
                    GrupoFilas(elementos: d.preguntas) { p in
                        Button { if p.respuesta == nil { respondiendo = p } } label: {
                            VStack(alignment: .leading, spacing: 6) {
                                HStack {
                                    Text(p.producto).estilo(.footnoteFuerte).foregroundStyle(C.texto2).lineLimit(1)
                                    Spacer()
                                    Text(Formato.hace(p.creada)).estilo(.footnote).foregroundStyle(C.texto3)
                                }
                                Text(p.pregunta).estilo(.body16).foregroundStyle(C.texto).fixedSize(horizontal: false, vertical: true)
                                if let r = p.respuesta {
                                    Text(r).estilo(.texto14).foregroundStyle(C.texto2).padding(10)
                                        .frame(maxWidth: .infinity, alignment: .leading)
                                        .background(C.crema2, in: RoundedRectangle(cornerRadius: 12, style: .continuous))
                                } else {
                                    Text("Responder").estilo(.subheadFuerte).foregroundStyle(C.primarioTexto)
                                }
                            }
                            .padding(14).contentShape(Rectangle())
                        }
                        .buttonStyle(.plain)
                    }
                }
            } else {
                if let prom = d.promedio {
                    TarjetaSeccion(titulo: "Lo que dicen de tu tienda", simbolo: "star.bubble") {
                        HStack(alignment: .center, spacing: 18) {
                            VStack(spacing: 4) {
                                Text(String(format: "%.1f", prom).replacingOccurrences(of: ".", with: ","))
                                    .estilo(.montoPantalla).foregroundStyle(C.texto)
                                Estrellas(n: Int(prom.rounded()))
                                Text(Formato.cantidad(d.total, "reseña", "reseñas")).estilo(.footnote).foregroundStyle(C.texto2)
                            }
                            BarrasEstrellas(porEstrellas: d.porEstrellas ?? [])
                        }
                    }
                }
                if d.resenas.isEmpty {
                    EstadoTincho(tincho: .celebrando, titulo: "Todavía no hay reseñas", texto: "Las de la tienda, Mercado Libre y Google aparecen acá.")
                        .padding(.top, 30)
                } else {
                    GrupoFilas(elementos: d.resenas) { r in
                        VStack(alignment: .leading, spacing: 6) {
                            HStack {
                                Estrellas(n: r.estrellas)
                                Text(r.autor).estilo(.footnoteFuerte).foregroundStyle(C.texto)
                                Spacer()
                                if r.oculta { ChipSimple(texto: "Oculta", fondo: C.pista, tinta: C.texto2) }
                            }
                            Text(r.texto).estilo(.body16).foregroundStyle(C.texto).fixedSize(horizontal: false, vertical: true)
                            Text([r.producto, r.origen, Formato.dia(r.creada), r.conCompra ? "compró" : nil].compactMap { $0 }.joined(separator: " · "))
                                .estilo(.footnote).foregroundStyle(C.texto3)
                            if let resp = r.respuesta {
                                Text(resp).estilo(.texto14).foregroundStyle(C.texto2).padding(10)
                                    .frame(maxWidth: .infinity, alignment: .leading)
                                    .background(C.crema2, in: RoundedRectangle(cornerRadius: 12, style: .continuous))
                            }
                            HStack(spacing: 16) {
                                if r.respuesta == nil {
                                    Button("Responder") { respondiendoResena = r }
                                }
                                Button(r.oculta ? "Mostrar" : "Ocultar") {
                                    Task {
                                        struct C2: Encodable { var accion: String }
                                        try? await estado.api.hacer("POST", "/v1/panel/resenas/\(r.id)", C2(accion: r.oculta ? "mostrar" : "ocultar"))
                                        await actualizar()
                                    }
                                }
                            }
                            .estilo(.subheadFuerte).foregroundStyle(C.primarioTexto).buttonStyle(.plain)
                        }
                        .padding(14)
                    }
                }
            }
            Color.clear.frame(height: 0)
                .sheet(item: $respondiendo) { p in
                    HojaRespuesta(titulo: p.pregunta, sugerida: p.sugerida) { texto in
                        struct C2: Encodable { var respuesta: String }
                        try await estado.api.hacer("POST", "/v1/panel/preguntas/\(p.id)", C2(respuesta: texto))
                        await actualizar()
                    }
                    .hojaTLW([.medium, .large])
                }
                .sheet(item: $respondiendoResena) { r in
                    HojaRespuesta(titulo: r.texto, sugerida: "") { texto in
                        struct C2: Encodable { var accion: String; var respuesta: String }
                        try await estado.api.hacer("POST", "/v1/panel/resenas/\(r.id)", C2(accion: "responder", respuesta: texto))
                        await actualizar()
                    }
                    .hojaTLW([.medium, .large])
                }
        }
    }
}

struct Estrellas: View {
    var n: Int
    var body: some View {
        HStack(spacing: 1) {
            ForEach(0..<5, id: \.self) { i in
                Image(systemName: i < n ? "star.fill" : "star").font(.system(size: 11)).foregroundStyle(C.aviso)
            }
        }
        .accessibilityLabel("\(n) de 5 estrellas")
    }
}

/// Hoja para responder una pregunta o una reseña (con la sugerencia de Hornero precargada).
struct HojaRespuesta: View {
    var titulo: String
    var sugerida: String
    var enviar: (String) async throws -> Void
    @Environment(\.dismiss) private var cerrar
    @Environment(Estado.self) private var estado
    @State private var texto = ""
    @State private var enviando = false

    var body: some View {
        VStack(alignment: .leading, spacing: 14) {
            Grabber()
            Text(titulo).estilo(.tituloHojaChico).foregroundStyle(C.texto).lineLimit(3)
            TextEditor(text: $texto)
                .scrollContentBackground(.hidden)
                .estilo(.body16).foregroundStyle(C.texto)
                .frame(minHeight: 120)
                .padding(10)
                .background(C.superficie, in: RoundedRectangle(cornerRadius: 14, style: .continuous))
                .overlay(RoundedRectangle(cornerRadius: 14, style: .continuous).strokeBorder(C.linea2, lineWidth: 1))
            Button(enviando ? "Publicando…" : "Publicar respuesta") {
                Task {
                    enviando = true
                    do {
                        try await enviar(texto.trimmingCharacters(in: .whitespacesAndNewlines))
                        estado.avisar("Respuesta publicada")
                        cerrar()
                    } catch { estado.avisar("No pudimos publicarla", ok: false) }
                    enviando = false
                }
            }
            .buttonStyle(BotonPrincipal())
            .disabled(texto.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty || enviando)
            Spacer(minLength: 0)
        }
        .padding(.horizontal, 20)
        .onAppear { if texto.isEmpty { texto = sugerida } }
    }
}

// MARK: - Devoluciones

struct Devoluciones: View {
    @Environment(Estado.self) private var estado

    struct Devolucion: Decodable, Identifiable {
        var id: String
        var producto: String
        var foto: String?
        var cliente: String
        var pedido: Int
        var creada: Date
        var motivo: String
        var monto: Int
        var resolucion: String
        var estado: String
    }
    struct Datos: Decodable { var devoluciones: [Devolucion]; var tasa: Double }

    var body: some View {
        PantallaSeccion(titulo: "Devoluciones", que: "las devoluciones") {
            try await estado.api.get("/v1/panel/devoluciones") as Datos
        } contenido: { d, actualizar in
            if d.devoluciones.isEmpty {
                EstadoTincho(tincho: .exito, titulo: "No hay devoluciones", texto: "Cuando alguien use el botón de arrepentimiento, lo resolvés acá.")
                    .padding(.top, 30)
            } else {
                let pendientes = d.devoluciones.filter { $0.estado == "Pendiente" }
                TarjetaHeroe(etiqueta: "Devoluciones para resolver", valor: "\(pendientes.count)",
                             detalle: "\(Formato.pesos(pendientes.reduce(0) { $0 + $1.monto })) en juego · \(Formato.porcentaje(d.tasa > 1 ? d.tasa : d.tasa * 100)) de los pedidos",
                             tincho: .envio)
                EtiquetaSeccion(texto: "Pedidos de devolución").padding(.horizontal, 4).padding(.top, 4)
                GrupoFilas(elementos: d.devoluciones) { v in
                    VStack(alignment: .leading, spacing: 8) {
                        HStack(spacing: 12) {
                            Miniatura(foto: v.foto, tam: 44)
                            VStack(alignment: .leading, spacing: 2) {
                                Text(v.producto).estilo(.fila).foregroundStyle(C.texto).lineLimit(1)
                                Text("\(v.cliente) · pedido #\(v.pedido) · \(Formato.dia(v.creada))")
                                    .estilo(.footnote).foregroundStyle(C.texto2).lineLimit(1)
                            }
                            Spacer(minLength: 6)
                            Text(Formato.pesos(v.monto)).estilo(.fila).monospacedDigit().foregroundStyle(C.texto)
                        }
                        Text(v.motivo).estilo(.texto14).foregroundStyle(C.texto2).fixedSize(horizontal: false, vertical: true)
                        if v.estado == "Pendiente" {
                            HStack(spacing: 10) {
                                Button("Aprobar") { Task { await accion(v, "aprobar", actualizar) } }
                                    .buttonStyle(BotonTonal())
                                Button("Ofrecer cambio") { Task { await accion(v, "cambio", actualizar) } }
                                    .buttonStyle(BotonSecundario(alto: 34))
                            }
                        } else {
                            ChipSimple(texto: v.estado, fondo: v.estado == "Aprobada" ? C.okFondo : C.pista,
                                       tinta: v.estado == "Aprobada" ? C.okTinta : C.texto2)
                        }
                    }
                    .padding(14)
                }
            }
        }
    }

    private func accion(_ v: Devolucion, _ a: String, _ actualizar: () async -> Void) async {
        struct C2: Encodable { var accion: String }
        do {
            try await estado.api.hacer("POST", "/v1/panel/devoluciones/\(v.id)", C2(accion: a))
            estado.avisar(a == "aprobar" ? "Devolución aprobada" : "Le ofrecimos el cambio") {
                Task { try? await estado.api.hacer("POST", "/v1/panel/devoluciones/\(v.id)", C2(accion: "deshacer")) }
            }
            await actualizar()
        } catch { estado.avisar("No pudimos guardarlo", ok: false) }
    }
}

// MARK: - Cupones

struct Cupones: View {
    @Environment(Estado.self) private var estado
    @State private var nuevo = false

    struct Cupon: Decodable, Identifiable {
        var id: String
        var codigo: String
        var tipo: String
        var valor: Int
        var descripcion: String
        var estado: String
        var usos: Int
        var limite: Int
        var vendio: Int
    }
    struct Datos: Decodable {
        struct Kpis: Decodable { var activos: Int; var usosMes: Int; var vendido: Int; var descontado: Int }
        var cupones: [Cupon]
        var kpis: Kpis
    }

    var body: some View {
        PantallaSeccion(titulo: "Cupones", que: "los cupones") {
            try await estado.api.get("/v1/panel/cupones") as Datos
        } contenido: { d, actualizar in
            Button { nuevo = true } label: { Label("Crear cupón", systemImage: "plus") }
                .buttonStyle(BotonPrincipal(alto: 48))
                .sheet(isPresented: $nuevo) { HojaCuponNuevo { await actualizar() }.hojaTLW([.large]) }
            TarjetaHeroe(etiqueta: "Vendido con cupones este mes", valor: Formato.pesos(d.kpis.vendido),
                         detalle: "\(Formato.cantidad(d.kpis.usosMes, "uso", "usos")) · \(Formato.pesos(d.kpis.descontado)) de descuento", tincho: .ticket)
            GrillaIndicadores(items: [
                IndicadorIcono(simbolo: "ticket.fill", titulo: "Cupones activos", valor: "\(d.kpis.activos)"),
                IndicadorIcono(simbolo: "percent", titulo: "Descontado", valor: Formato.pesos(d.kpis.descontado), tono: (C.okFondo, C.okTinta)),
            ])
            if d.cupones.isEmpty {
                EstadoTincho(tincho: .ticket, titulo: "Todavía no tenés cupones", texto: "Un descuento con código para tus clientes: lo creás en un minuto.")
                    .padding(.top, 30)
            } else {
                GrupoFilas(elementos: d.cupones) { c in
                    HStack(spacing: 12) {
                        VStack(alignment: .leading, spacing: 3) {
                            Text(c.codigo).font(Estilo(tamano: 16, peso: .bold, relativo: .body).font).monospaced()
                                .foregroundStyle(C.texto)
                            Text(c.descripcion.isEmpty ? Cupones.resumen(c) : c.descripcion)
                                .estilo(.footnote).foregroundStyle(C.texto2).lineLimit(2)
                            Text("\(Formato.cantidad(c.usos, "uso", "usos"))\(c.limite > 0 ? " de \(c.limite)" : "") · vendió \(Formato.pesos(c.vendio))")
                                .estilo(.footnote).foregroundStyle(C.texto3)
                        }
                        Spacer(minLength: 6)
                        if c.estado == "activo" || c.estado == "pausado" {
                            Toggle("", isOn: Binding(get: { c.estado == "activo" }, set: { on in
                                Task {
                                    struct C2: Encodable { var accion: String }
                                    try? await estado.api.hacer("POST", "/v1/panel/cupones/\(c.id)", C2(accion: on ? "activar" : "pausar"))
                                    await actualizar()
                                }
                            }))
                            .labelsHidden().tint(C.ok)
                        } else {
                            ChipSimple(texto: Formato.mayuscula(c.estado), fondo: C.pista, tinta: C.texto2)
                        }
                    }
                    .padding(14)
                }
            }
        }
    }

    static func resumen(_ c: Cupon) -> String {
        switch c.tipo {
        case "pct": "\(c.valor) % de descuento"
        case "fijo": "\(Formato.pesos(c.valor)) de descuento"
        case "envio": "Envío gratis"
        default: "Regalo"
        }
    }
}

// MARK: - Finanzas

struct Finanzas: View {
    @Environment(Estado.self) private var estado

    struct Datos: Decodable {
        struct Kpis: Decodable {
            var aLiquidar: Int?
            var aLiquidarDetalle: String?
            var gastos: Int?
            var gastosDetalle: String?
            var ganancia: Int?
            var gananciaDetalle: String?
            var margen: Int?
            var mes: String
        }
        struct Movimiento: Decodable, Identifiable {
            var id: String
            var tipo: String
            var titulo: String
            var detalle: String
            var monto: Int
        }
        struct Flujo: Decodable { var dia: Int; var entra: Int; var sale: Int }
        struct Caja: Decodable { var ventas: Int; var gastos: Int; var meta: String; var metaPct: Int; var falta: Int }
        var vendido30: Int?
        var disponible: Int?
        var cuenta: String?
        var kpis: Kpis
        var flujo: [Flujo]?
        var caja: Caja?
        var movimientos: [Movimiento]
    }

    var body: some View {
        PantallaSeccion(titulo: "Finanzas", que: "las finanzas") {
            try await estado.api.get("/v1/panel/finanzas") as Datos
        } contenido: { d, _ in
            TarjetaHeroe(etiqueta: d.disponible != nil ? "Disponible en Mercado Pago" : "Vendido en los últimos 30 días",
                         valor: Formato.pesos(d.disponible ?? d.vendido30 ?? 0),
                         detalle: d.cuenta ?? d.vendido30.map { d.disponible != nil ? "Vendiste \(Formato.pesos($0)) en 30 días" : "" },
                         tincho: .grafico)
            EtiquetaSeccion(texto: Formato.mayuscula(d.kpis.mes)).padding(.horizontal, 4).padding(.top, 4)
            GrillaIndicadores(items: [
                IndicadorIcono(simbolo: "clock.arrow.circlepath", titulo: "A liquidar", valor: d.kpis.aLiquidar.map(Formato.pesos) ?? "—",
                               detalle: d.kpis.aLiquidarDetalle),
                IndicadorIcono(simbolo: "arrow.down.right", titulo: "Gastos", valor: d.kpis.gastos.map(Formato.pesos) ?? "—",
                               detalle: d.kpis.gastosDetalle, tono: (C.errorFondo, C.errorTinta)),
                IndicadorIcono(simbolo: "arrow.up.right", titulo: "Ganancia", valor: d.kpis.ganancia.map(Formato.pesos) ?? "—",
                               detalle: d.kpis.gananciaDetalle, tono: (C.okFondo, C.okTinta)),
                IndicadorIcono(simbolo: "percent", titulo: "Margen", valor: d.kpis.margen.map { "\($0) %" } ?? "—",
                               tono: (C.avisoFondo, C.avisoTinta)),
            ])
            if let c = d.caja, c.metaPct > 0 || !c.meta.isEmpty {
                TarjetaSeccion(titulo: "Meta del mes", simbolo: "flag.checkered") {
                    HStack(spacing: 16) {
                        AnilloAvance(valor: Double(c.metaPct) / 100)
                        VStack(alignment: .leading, spacing: 4) {
                            Text(c.meta).estilo(.fila).foregroundStyle(C.texto).lineLimit(2)
                            if c.falta > 0 {
                                Text("Faltan \(Formato.pesos(c.falta))").estilo(.footnote).monospacedDigit().foregroundStyle(C.texto2)
                            } else {
                                Text("¡Llegaste a la meta!").estilo(.footnoteFuerte).foregroundStyle(C.okTinta)
                            }
                        }
                        Spacer(minLength: 0)
                    }
                }
            }
            if let f = d.flujo, f.contains(where: { $0.entra > 0 || $0.sale > 0 }) {
                TarjetaSeccion(titulo: "Lo que entra y sale (próximos 14 días)", simbolo: "chart.bar.xaxis") {
                    GraficoFlujo(dias: f.map { .init(dia: $0.dia, entra: $0.entra, sale: $0.sale) })
                    HStack(spacing: 14) {
                        Label("Entra", systemImage: "circle.fill").foregroundStyle(C.ok)
                        Label("Sale", systemImage: "circle.fill").foregroundStyle(C.error)
                    }
                    .labelStyle(.titleAndIcon).font(.caption2.weight(.semibold))
                }
                let flujo = f.map { GraficoFlujo.Dia(dia: $0.dia, entra: $0.entra, sale: $0.sale) }
                let neto = flujo.reduce(0) { $0 + $1.entra - $1.sale }
                TarjetaSeccion(titulo: "Cómo queda la caja", simbolo: "chart.line.uptrend.xyaxis") {
                    GraficoSaldo(flujo: flujo)
                    Text(Formato.sinViudas(neto >= 0
                        ? "Si entra y sale lo previsto, en 14 días sumás \(Formato.pesos(neto))."
                        : "Si entra y sale lo previsto, en 14 días salen \(Formato.pesos(-neto)) más de lo que entra."))
                        .estilo(.footnote).foregroundStyle(neto >= 0 ? C.okTinta : C.errorTinta)
                }
            }
            if !d.movimientos.isEmpty {
                EtiquetaSeccion(texto: "Movimientos").padding(.horizontal, 4).padding(.top, 8)
                GrupoFilas(elementos: d.movimientos, inset: 62) { m in
                    HStack(spacing: 12) {
                        Image(systemName: simboloMovimiento(m.tipo)).font(.system(size: 14, weight: .semibold))
                            .foregroundStyle(m.monto < 0 ? C.errorTinta : C.okTinta)
                            .frame(width: 34, height: 34)
                            .background(m.monto < 0 ? C.errorFondo : C.okFondo, in: RoundedRectangle(cornerRadius: 10, style: .continuous))
                        VStack(alignment: .leading, spacing: 2) {
                            Text(m.titulo).estilo(.fila).foregroundStyle(C.texto).lineLimit(1)
                            Text(m.detalle).estilo(.footnote).foregroundStyle(C.texto2).lineLimit(1)
                        }
                        Spacer(minLength: 6)
                        Text((m.monto < 0 ? "−" : "+") + Formato.pesos(abs(m.monto)))
                            .estilo(.fila).monospacedDigit().foregroundStyle(m.monto < 0 ? C.errorTinta : C.okTinta)
                    }
                    .padding(.horizontal, 14).padding(.vertical, 12)
                }
            }
        }
    }
}

// MARK: - Facturación de ARCA

struct FacturacionArca: View {
    @Environment(Estado.self) private var estado
    @State private var alta = false

    struct Datos: Decodable {
        var conectada: Bool
        var cuit: String
        var razonSocial: String
        var condicion: String
        var puntoVenta: Int
        var ambiente: String
        var certificado: Bool
        var certificadoVence: Date?
        var facturarAlCobrar: Bool
    }

    var body: some View {
        PantallaSeccion(titulo: "Facturación de ARCA", que: "la facturación") {
            try await estado.api.get("/v1/panel/arca") as Datos
        } contenido: { d, actualizar in
            if !d.conectada {
                EstadoTincho(tincho: .portapapeles, titulo: "Todavía no conectaste ARCA",
                             texto: "Para facturar hace falta el certificado de ARCA. Lo cargás una sola vez, en tres pasos.",
                             accion: "Conectar ARCA", principal: true) { alta = true }
                    .padding(.top, 20)
            } else {
                TarjetaHeroe(etiqueta: "ARCA conectada", valor: d.razonSocial.isEmpty ? d.cuit : d.razonSocial,
                             detalle: "CUIT \(d.cuit) · punto de venta \(d.puntoVenta)", tincho: .portapapeles)
                if let vence = d.certificadoVence {
                    let dias = Calendar.current.dateComponents([.day], from: .now, to: vence).day ?? 0
                    TarjetaSeccion(titulo: "Certificado", simbolo: "checkmark.seal") {
                        HStack(alignment: .firstTextBaseline) {
                            Text(dias > 0 ? Formato.cantidad(dias, "día", "días") : "Vencido").estilo(.indicador)
                                .foregroundStyle(dias < 30 ? C.errorTinta : C.texto)
                            Text(dias > 0 ? "hasta que venza" : "renovalo para seguir facturando").estilo(.footnote).foregroundStyle(C.texto2)
                            Spacer()
                        }
                        BarraProgreso(valor: Double(max(0, min(730, dias))) / 730)
                    }
                }
                GrupoFilas(elementos: [
                    FilaId("Razón social", d.razonSocial), FilaId("CUIT", d.cuit), FilaId("Condición", d.condicion),
                    FilaId("Punto de venta", "\(d.puntoVenta)"),
                    FilaId("Ambiente", d.ambiente == "produccion" ? "Producción" : "Homologación"),
                    FilaId("Certificado", d.certificadoVence.map { "vence el \(Formato.dia($0))" } ?? (d.certificado ? "cargado" : "falta")),
                ]) { f in FilaDato(etiqueta: f.etiqueta, valor: f.valor) }
                HStack {
                    Text("Facturar al cobrar").estilo(.body16).foregroundStyle(C.texto)
                    Spacer()
                    Toggle("", isOn: Binding(get: { d.facturarAlCobrar }, set: { on in
                        Task {
                            struct C2: Encodable { var facturarAlCobrar: Bool }
                            do {
                                try await estado.api.hacer("PUT", "/v1/panel/arca/ajustes", C2(facturarAlCobrar: on))
                                await actualizar()
                            } catch { estado.avisar("No pudimos guardarlo", ok: false) }
                        }
                    }))
                    .labelsHidden().tint(C.ok)
                }
                .padding(.horizontal, 16).padding(.vertical, 12).tarjeta()
                Button("Renovar el certificado") { alta = true }.buttonStyle(BotonSecundario()).frame(maxWidth: .infinity)
            }
            Color.clear.frame(height: 0)
                .sheet(isPresented: $alta) { HojaArca { await actualizar() }.hojaTLW([.large]) }
        }
    }
}

struct FilaId: Identifiable {
    var etiqueta: String
    var valor: String
    var id: String { etiqueta }
    init(_ e: String, _ v: String) { etiqueta = e; valor = v }
}

// El plan (PlanCuenta) está en Plan.swift.

// MARK: - Integraciones

struct Integraciones: View {
    @Environment(Estado.self) private var estado
    @State private var correo = false
    @State private var conectando: String?

    // Todo opcional: sin cuenta conectada la API no manda usuario, apodo ni mail.
    struct MP: Decodable { var conectada: Bool?; var usuario: String?; var reautorizar: Bool? }
    struct ML: Decodable { var conectada: Bool?; var apodo: String?; var reautorizar: Bool? }
    struct Correo: Decodable { var conectada: Bool?; var email: String? }
    struct Datos { var mp: MP?; var ml: ML?; var correo: Correo? }

    var body: some View {
        PantallaSeccion(titulo: "Integraciones", que: "las integraciones") { () async throws -> Datos in
            // Sin `async let` en la carga de PantallaSeccion: en Release cerraba la app (ver Compartir).
            let mp: MP? = try? await estado.api.get("/v1/panel/integraciones/mercadopago")
            let ml: ML? = try? await estado.api.get("/v1/panel/integraciones/mercadolibre")
            let co: Correo? = try? await estado.api.get("/v1/panel/integraciones/correo")
            return Datos(mp: mp, ml: ml, correo: co)
        } contenido: { d, actualizar in
            let mp = d.mp?.conectada == true, ml = d.ml?.conectada == true, co = d.correo?.conectada == true
            let filas = [
                FilaIntegracion(id: "mp", nombre: "Mercado Pago", detalle: mp ? (d.mp?.usuario ?? "Conectada") : "Cobrá con tarjeta, QR y link",
                                conectada: mp, alerta: d.mp?.reautorizar == true),
                FilaIntegracion(id: "ml", nombre: "Mercado Libre", detalle: ml ? (d.ml?.apodo ?? "Conectada") : "Publicá y sincronizá el stock",
                                conectada: ml, alerta: d.ml?.reautorizar == true),
                FilaIntegracion(id: "correo", nombre: "Correo Argentino", detalle: co ? (d.correo?.email ?? "Conectada") : "Etiquetas y seguimiento de envíos",
                                conectada: co, alerta: false),
            ]
            GrupoFilas(elementos: filas) { f in
                HStack {
                    VStack(alignment: .leading, spacing: 2) {
                        Text(f.nombre).estilo(.fila).foregroundStyle(C.texto)
                        Text(f.detalle).estilo(.footnote).foregroundStyle(C.texto2).lineLimit(1)
                    }
                    Spacer()
                    if f.conectada && !f.alerta {
                        ChipSimple(texto: "Conectada", fondo: C.okFondo, tinta: C.okTinta)
                    } else if conectando == f.id {
                        ProgressView().tint(C.acento)
                    } else {
                        Button(f.alerta ? "Volver a conectar" : "Conectar") {
                            if f.id == "correo" { correo = true; return }
                            Task {
                                conectando = f.id
                                let ok = await ConectarCuenta.compartido.conectar(f.id == "mp" ? "mercadopago" : "mercadolibre", api: estado.api)
                                conectando = nil
                                estado.avisar(ok ? "\(f.nombre) conectado" : "No se completó la conexión", ok: ok)
                                if ok { estado.version += 1 }
                                await actualizar()
                            }
                        }
                        .buttonStyle(BotonTonal())
                    }
                }
                .padding(.horizontal, 16).padding(.vertical, 12)
            }
            Color.clear.frame(height: 0)
                .sheet(isPresented: $correo) { HojaCorreo { await actualizar() }.hojaTLW([.large]) }
            EtiquetaSeccion(texto: "Más integraciones").padding(.horizontal, 4).padding(.top, 8)
            VStack(spacing: 0) {
                ForEach(Array(Integraciones.otras.enumerated()), id: \.element.destino) { i, o in
                    if i > 0 { Separador() }
                    NavigationLink(value: o.destino) {
                        HStack(spacing: 12) {
                            Image(systemName: o.simbolo).font(.system(size: 15, weight: .semibold)).foregroundStyle(C.primarioTexto)
                                .frame(width: 34, height: 34).background(C.crema, in: RoundedRectangle(cornerRadius: 10, style: .continuous))
                            VStack(alignment: .leading, spacing: 2) {
                                Text(o.nombre).estilo(.fila).foregroundStyle(C.texto)
                                Text(o.detalle).estilo(.footnote).foregroundStyle(C.texto2).lineLimit(1)
                            }
                            Spacer(minLength: 4)
                            Chevron(ancho: 8)
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

    struct FilaIntegracion: Identifiable { var id: String; var nombre: String; var detalle: String; var conectada: Bool; var alerta: Bool }

    /// Las que tienen su pantalla: Telegram, medición y las IA.
    static let otras: [(destino: Destino, nombre: String, detalle: String, simbolo: String)] = [
        (.telegram, "Tincho en Telegram", "Avisos de ventas y consultas por chat", "paperplane"),
        (.medicion, "Google y Meta", "Analytics, píxel y catálogo de productos", "chart.xyaxis.line"),
        (.ia, "Tu IA (Claude, ChatGPT)", "Manejá la tienda desde tu asistente", "sparkles"),
    ]
}

// MARK: - Usuarios y permisos

struct UsuariosYPermisos: View {
    @Environment(Estado.self) private var estado
    @State private var invitar = false

    struct Miembro: Decodable, Identifiable {
        var id: String
        var nombre: String
        var contacto: String
        var rol: String
        var pendiente: Bool
        var vos: Bool
    }
    struct Datos: Decodable { var equipo: [Miembro] }

    static let roles = [("admin", "Administración"), ("ventas", "Ventas"), ("local", "Local")]

    var body: some View {
        PantallaSeccion(titulo: "Usuarios y permisos", que: "el equipo") {
            try await estado.api.get("/v1/panel/configuracion") as Datos
        } contenido: { d, actualizar in
            let activos = d.equipo.filter { !$0.pendiente }.count
            GrillaIndicadores(items: [
                IndicadorIcono(simbolo: "person.2.fill", titulo: "En el equipo", valor: "\(activos)"),
                IndicadorIcono(simbolo: "envelope.badge", titulo: "Invitaciones pendientes", valor: "\(d.equipo.count - activos)",
                               tono: (C.avisoFondo, C.avisoTinta)),
            ])
            GrupoFilas(elementos: d.equipo, inset: 68) { m in
                HStack(spacing: 12) {
                    AvatarIniciales(nombre: m.nombre.isEmpty ? m.contacto : m.nombre, tam: 42)
                    VStack(alignment: .leading, spacing: 2) {
                        Text(m.nombre.isEmpty ? m.contacto : m.nombre + (m.vos ? " (vos)" : ""))
                            .estilo(.fila).foregroundStyle(C.texto).lineLimit(1)
                        Text(m.pendiente ? "Invitación pendiente · \(m.contacto)" : m.contacto)
                            .estilo(.footnote).foregroundStyle(C.texto2).lineLimit(1)
                    }
                    Spacer(minLength: 6)
                    if m.vos || m.rol == "duena" || m.rol == "dueña" {
                        Text(UsuariosYPermisos.nombreRol(m.rol)).estilo(.label).foregroundStyle(C.texto2)
                    } else {
                        Menu {
                            ForEach(UsuariosYPermisos.roles, id: \.0) { r in
                                Button(r.1) { Task { await cambiar(m, rol: r.0, actualizar) } }
                            }
                            if m.pendiente {
                                Button("Reenviar invitación") {
                                    Task { try? await estado.api.hacer("POST", "/v1/panel/equipo/\(m.id)/reenviar"); estado.avisar("Invitación reenviada") }
                                }
                            }
                            Button("Quitar del equipo", role: .destructive) {
                                Task {
                                    do { try await estado.api.hacer("DELETE", "/v1/panel/equipo/\(m.id)"); await actualizar() }
                                    catch { estado.avisar("No pudimos quitarlo", ok: false) }
                                }
                            }
                        } label: {
                            HStack(spacing: 4) {
                                Text(UsuariosYPermisos.nombreRol(m.rol))
                                Image(systemName: "chevron.down").font(.system(size: 11, weight: .bold))
                            }
                            .estilo(.label).foregroundStyle(C.primarioTexto)
                        }
                    }
                }
                .padding(.horizontal, 14).padding(.vertical, 12)
            }
            Button("Invitar a alguien") { invitar = true }
                .buttonStyle(BotonSecundario()).frame(maxWidth: .infinity).padding(.top, 4)
                .sheet(isPresented: $invitar) {
                    HojaInvitar { await actualizar() }.hojaTLW([.medium])
                }
        }
    }

    static func nombreRol(_ r: String) -> String {
        roles.first { $0.0 == r }?.1 ?? (r.hasPrefix("due") ? "Dueña" : Formato.mayuscula(r))
    }

    private func cambiar(_ m: Miembro, rol: String, _ actualizar: () async -> Void) async {
        struct C2: Encodable { var rol: String }
        do {
            try await estado.api.hacer("PUT", "/v1/panel/equipo/\(m.id)", C2(rol: rol))
            await actualizar()
        } catch { estado.avisar("No pudimos cambiar el rol", ok: false) }
    }
}

struct HojaInvitar: View {
    var listo: () async -> Void
    @Environment(\.dismiss) private var cerrar
    @Environment(Estado.self) private var estado
    @State private var email = ""
    @State private var rol = "ventas"
    @State private var enviando = false

    var body: some View {
        VStack(alignment: .leading, spacing: 14) {
            Grabber()
            Text("Invitar a alguien").estilo(.tituloHojaChico).foregroundStyle(C.texto)
            TextField("Email", text: $email)
                .keyboardType(.emailAddress).textInputAutocapitalization(.never).autocorrectionDisabled()
                .estilo(.body16).padding(14)
                .background(C.superficie, in: RoundedRectangle(cornerRadius: 14, style: .continuous))
                .overlay(RoundedRectangle(cornerRadius: 14, style: .continuous).strokeBorder(C.linea2, lineWidth: 1))
            Segmentado(opciones: UsuariosYPermisos.roles.map { ($0.0, $0.1) }, elegido: $rol)
            Button(enviando ? "Enviando…" : "Mandar invitación") {
                Task {
                    enviando = true
                    struct C2: Encodable { var contacto: String; var rol: String }
                    do {
                        try await estado.api.hacer("POST", "/v1/panel/equipo", C2(contacto: email.trimmingCharacters(in: .whitespaces), rol: rol))
                        estado.avisar("Le mandamos la invitación")
                        await listo()
                        cerrar()
                    } catch { estado.avisar("No pudimos invitarla", ok: false) }
                    enviando = false
                }
            }
            .buttonStyle(BotonPrincipal())
            .disabled(!email.contains("@") || enviando)
            Spacer(minLength: 0)
        }
        .padding(.horizontal, 20)
    }
}
