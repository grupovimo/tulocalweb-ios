import SwiftUI

/// Cobros (Configuración › Cobros del panel): las cuotas sin interés (ninguna, 3, 6, 9 o 12, desde un monto)
/// y el precio sin impuestos nacionales (Ley 27.743). Las cuotas y el interruptor se guardan al toque, con
/// «Deshacer»; el «desde», con su botón. Las reglas son las de Go (dominio/cobros): sin Mercado Pago
/// conectado la tienda no las ofrece aunque estén elegidas, y el monto va de $ 0 a $ 100.000.000.
struct CobrosTienda: View {
    @Environment(Estado.self) private var estado
    @Environment(Navegacion.self) private var nav
    @Environment(\.openURL) private var abrirWeb

    /// Lo que se guarda (CobrosTienda del contrato).
    struct Cobros: Codable, Equatable {
        struct Impuestos: Codable, Equatable { var mostrar: Bool }
        var cuotas: Int
        var cuotasDesde: Int
        var impuestos: Impuestos
    }

    /// Las cuotas que se pueden ofrecer (0: ninguna).
    static let cuotas: [(Int, String)] = [(0, "Ninguna"), (3, "3"), (6, "6"), (9, "9"), (12, "12")]
    static let maximoDesde = 100_000_000
    /// Donde la tienda activa en Mercado Pago las cuotas sin interés que absorbe.
    static let costosMercadoPago = URL(string: "https://www.mercadopago.com.ar/costs-section")!

    @State private var c: Cobros?
    @State private var mercadoPago = true
    @State private var desde = ""
    @State private var guardando = false

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 12) {
                Text("Cobros").estilo(.tituloGrande).foregroundStyle(C.texto)
                Text(Formato.sinViudas("Las cuotas sin interés y el precio sin impuestos que ve quien compra."))
                    .estilo(.subhead).foregroundStyle(C.texto2)
                if let c {
                    tarjetaCuotas(c)
                    tarjetaImpuestos(c)
                } else {
                    EsqueletoFilas(filas: 4, miniatura: 0)
                }
            }
            .margenPantalla()
            .padding(.bottom, 32)
        }
        .scrollDismissesKeyboard(.interactively)
        .barraAccion {
            if cambioDesde {
                Button(guardando ? "Guardando…" : "Guardar monto") { Task { await guardarDesde() } }
                    .buttonStyle(BotonPrincipal()).disabled(guardando)
                Button("Descartar") { desde = CobrosTienda.texto(c?.cuotasDesde ?? 0) }
                    .buttonStyle(BotonTexto())
            }
        }
        .background(C.fondo.ignoresSafeArea())
        .navigationBarTitleDisplayMode(.inline)
        .task { await cargar() }
    }

    private var montoDesde: Int { Int(desde.filter(\.isNumber).prefix(12)) ?? 0 }
    private var cambioDesde: Bool { (c?.cuotas ?? 0) > 0 && montoDesde != c?.cuotasDesde }

    private static func texto(_ monto: Int) -> String { monto > 0 ? String(monto) : "" }

    @ViewBuilder
    private func tarjetaCuotas(_ v: Cobros) -> some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack(spacing: 12) {
                IconoOpcion(simbolo: "creditcard")
                VStack(alignment: .leading, spacing: 2) {
                    Text("Cuotas sin interés").estilo(.fila).foregroundStyle(C.texto)
                    Text(v.cuotas > 0 ? "Hasta \(v.cuotas) cuotas con Mercado Pago" : "Tu tienda no las ofrece")
                        .estilo(.footnote).foregroundStyle(C.texto2)
                }
                Spacer(minLength: 0)
            }
            if !mercadoPago {
                VStack(alignment: .leading, spacing: 8) {
                    Text(Formato.sinViudas("Se ofrecen sólo cobrando con Mercado Pago: conectalo en Integraciones."))
                        .estilo(.footnote).foregroundStyle(C.avisoTinta)
                    Button("Ir a Integraciones") { nav.abrir(.integraciones) }.buttonStyle(BotonTonal())
                }
                .padding(12)
                .frame(maxWidth: .infinity, alignment: .leading)
                .background(C.avisoFondo, in: RoundedRectangle(cornerRadius: 12, style: .continuous))
            }
            Segmentado(opciones: CobrosTienda.cuotas, elegido: Binding(get: { c?.cuotas ?? 0 }, set: { n in
                Task { await elegirCuotas(n) }
            }))
            .accessibilityLabel("Cuotas sin interés")
            if v.cuotas > 0 {
                VStack(alignment: .leading, spacing: 6) {
                    Text("Desde $").estilo(.footnoteFuerte).foregroundStyle(C.texto2)
                    TextField("En cualquier compra", text: $desde)
                        .keyboardType(.numberPad)
                        .estilo(.body16).foregroundStyle(C.texto)
                        .padding(.horizontal, 14).padding(.vertical, 12)
                        .background(C.superficie, in: RoundedRectangle(cornerRadius: 14, style: .continuous))
                        .overlay(RoundedRectangle(cornerRadius: 14, style: .continuous).strokeBorder(C.linea2, lineWidth: 1))
                }
            }
            Text(Formato.sinViudas("El costo de las cuotas lo absorbe tu tienda: quien compra paga el mismo precio en todas. Activalas también en tu cuenta de Mercado Pago, en Costos y cuotas, con las mismas cuotas y el mismo monto."))
                .estilo(.footnote).foregroundStyle(C.texto2)
            Button { abrirWeb(CobrosTienda.costosMercadoPago) } label: {
                Label("Costos y cuotas en Mercado Pago", systemImage: "arrow.up.right.square")
            }
            .buttonStyle(BotonTexto(alto: 44))
        }
        .padding(16).tarjeta()
    }

    private func tarjetaImpuestos(_ v: Cobros) -> some View {
        HStack(spacing: 12) {
            IconoOpcion(simbolo: "receipt")
            VStack(alignment: .leading, spacing: 2) {
                Text("Mostrar precio sin impuestos nacionales").estilo(.fila).foregroundStyle(C.texto)
                Text(Formato.sinViudas("Lo pide la Ley 27.743 a quien discrimina IVA. Cada producto tiene su IVA en su ficha."))
                    .estilo(.footnote).foregroundStyle(C.texto2)
            }
            Spacer(minLength: 6)
            Toggle("Mostrar precio sin impuestos nacionales", isOn: Binding(get: { c?.impuestos.mostrar ?? false }, set: { m in
                Task { await elegirImpuestos(m) }
            }))
            .labelsHidden().tint(C.ok)
        }
        .padding(16).tarjeta()
    }

    private func cargar() async {
        struct Conf: Decodable {
            struct Cob: Decodable { var cuotas: Int; var cuotasDesde: Int; var impuestos: Cobros.Impuestos; var mercadoPago: Bool? }
            var cobros: Cob?
        }
        do {
            let v: Conf = try await estado.api.get("/v1/panel/configuracion")
            let x = v.cobros
            c = Cobros(cuotas: x?.cuotas ?? 0, cuotasDesde: x?.cuotasDesde ?? 30000, impuestos: x?.impuestos ?? .init(mostrar: false))
            mercadoPago = x?.mercadoPago ?? false
            desde = CobrosTienda.texto(c?.cuotasDesde ?? 0)
        } catch {
            estado.avisar("No pudimos traer los cobros", ok: false)
        }
    }

    private func elegirCuotas(_ n: Int) async {
        guard var v = c, v.cuotas != n else { return }
        v.cuotas = n
        await guardar(v, aviso: n > 0 ? "\(n) cuotas sin interés desde \(Formato.pesos(v.cuotasDesde))" : "Tu tienda ya no ofrece cuotas sin interés")
    }

    private func elegirImpuestos(_ m: Bool) async {
        guard var v = c, v.impuestos.mostrar != m else { return }
        v.impuestos.mostrar = m
        await guardar(v, aviso: m ? "Precio sin impuestos a la vista" : "Tu tienda ya no muestra el precio sin impuestos")
    }

    private func guardarDesde() async {
        guard var v = c else { return }
        guard montoDesde <= CobrosTienda.maximoDesde else {
            estado.avisar("El monto va de $ 0 a $ 100.000.000.", ok: false)
            return
        }
        v.cuotasDesde = montoDesde
        guardando = true
        defer { guardando = false }
        if await guardar(v, aviso: nil) {
            desde = CobrosTienda.texto(v.cuotasDesde)
            estado.avisar("Monto guardado: tu tienda ya usa el nuevo mínimo")
        }
    }

    /// Guarda al toque (se ve enseguida y vuelve atrás si falla). Con `aviso`, el toast trae «Deshacer».
    @discardableResult
    private func guardar(_ v: Cobros, aviso: String?, deshaciendo: Bool = false) async -> Bool {
        guard let antes = c else { return false }
        c = v
        do {
            try await estado.api.hacer("PUT", "/v1/panel/configuracion/cobros", v)
            guard !deshaciendo else { return true }
            Haptico.exito()
            if let aviso {
                estado.avisar(aviso, deshacer: {
                    Task { await guardar(antes, aviso: nil, deshaciendo: true) }
                })
            }
            return true
        } catch let x as ErrorApi {
            c = antes
            estado.avisar(x.detalle, ok: false)
        } catch {
            c = antes
            estado.avisar("No pudimos guardar los cobros", ok: false)
        }
        return false
    }
}
