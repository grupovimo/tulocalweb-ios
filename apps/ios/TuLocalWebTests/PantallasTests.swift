import SwiftUI
import XCTest
@testable import TuLocalWeb

/// Responde a la app con datos fijos, por ruta.
final class ApiFalsa: URLProtocol {
    nonisolated(unsafe) static var respuestas: [String: String] = [:]
    override class func canInit(with request: URLRequest) -> Bool { true }
    override class func canonicalRequest(for request: URLRequest) -> URLRequest { request }
    override func startLoading() {
        let ruta = request.url?.path ?? ""
        let cuerpo = ApiFalsa.respuestas[ruta]
        let r = HTTPURLResponse(url: request.url!, statusCode: cuerpo == nil ? 404 : 200, httpVersion: nil,
                                headerFields: ["content-type": "application/json"])!
        client?.urlProtocol(self, didReceive: r, cacheStoragePolicy: .notAllowed)
        client?.urlProtocol(self, didLoad: Data((cuerpo ?? "{}").utf8))
        client?.urlProtocolDidFinishLoading(self)
    }
    override func stopLoading() {}
}

/// Monta pantallas enteras en una ventana para que un crash al abrirlas salte en el CI
/// (en la Mac no se usa el simulador).
@MainActor
final class PantallasTests: XCTestCase {
    static let tienda = Acceso.Tienda(id: "t1", nombre: "Almacén Lumbre", slug: "almacen-lumbre", rol: "owner", logo: "https://api.prueba/logo.png")

    /// Las respuestas compartidas con Android (apps/pruebas/respuestas.json): «MÉTODO ruta» o sólo la ruta.
    static func respuestas() -> [String: String] {
        guard let url = Bundle(for: PantallasTests.self).url(forResource: "respuestas", withExtension: "json"),
              let datos = try? Data(contentsOf: url),
              let raiz = try? JSONSerialization.jsonObject(with: datos) as? [String: Any] else { return [:] }
        var r: [String: String] = [:]
        for (k, v) in raiz where !k.hasPrefix("_") {
            if let d = try? JSONSerialization.data(withJSONObject: v, options: [.fragmentsAllowed]) { r[k] = String(decoding: d, as: UTF8.self) }
        }
        return r
    }

    override func setUp() {
        ApiFalsa.respuestas = Self.respuestas()
    }

    private func montar<V: View>(_ vista: V, segundos: Double = 4, oscuro: Bool = false) {
        let conf = URLSessionConfiguration.ephemeral
        conf.protocolClasses = [ApiFalsa.self]
        let estado = Estado(api: Api(base: URL(string: "https://api.prueba")!, token: "x", http: URLSession(configuration: conf)))
        estado.acceso = Acceso(tiendaId: "t1", rol: "owner", permisos: ["catalogo", "ventas"], tiendas: [Self.tienda])
        let raiz = NavigationStack { vista }
            .environment(estado)
            .environment(Navegacion())
        let ventana = UIWindow(frame: CGRect(x: 0, y: 0, width: 390, height: 844))
        ventana.overrideUserInterfaceStyle = oscuro ? .dark : .light
        ventana.rootViewController = UIHostingController(rootView: raiz)
        ventana.makeKeyAndVisible()
        RunLoop.main.run(until: Date().addingTimeInterval(segundos))
        ventana.isHidden = true
    }

    func testCompartirTienda() { montar(CompartirTienda()) }
    func testCompartirTiendaSinProductos() {
        ApiFalsa.respuestas["/v1/panel/listas/productos"] = "{\"total\": 0, \"segmentos\": {}, \"filas\": []}"
        ApiFalsa.respuestas["/v1/panel/cupones"] = nil
        montar(CompartirTienda())
    }
    // Todas las pantallas de la app, en Release y con la API simulada (lo que no está, 404): un crash al abrirlas salta acá.
    func testPantallaInicio() { montar(Inicio()) }
    /// El Inicio «La vidriera» en sus estados: lleno (arriba), el día en $0 con un acento claro, oscuro y cargando.
    func testPantallaInicioEnCero() {
        ApiFalsa.respuestas["/v1/panel/tablero"] = ApiFalsa.respuestas["/v1/panel/tablero#cero"]
        montar(Inicio())
    }
    func testPantallaInicioOscuro() { montar(Inicio(), oscuro: true) }
    func testPantallaInicioCargando() {
        let paleta = PaletaTienda(acento: "#2F5D50", sobre: "#FFFFFF", oscuro: false)
        montar(ScrollView {
            VStack(spacing: 0) {
                Vidriera(carga: .cargando, periodo: .constant("hoy"), sinLeer: 0, paleta: paleta, se: false)
                EsqueletoInicio(paleta: paleta)
            }
        }, segundos: 1)
    }
    /// La paleta de la tienda: el texto encima sale de `sobreAcento` (oscuro sobre Pampa) y los tonos se mezclan como en el handoff.
    func testPaletaTienda() {
        XCTAssertFalse(PaletaTienda(acento: "#E6FF4D", sobre: "#111111", oscuro: false).claroEncima)
        XCTAssertTrue(PaletaTienda(acento: "#2F5D50", sobre: "#FFFFFF", oscuro: false).claroEncima)
        XCTAssertTrue(PaletaTienda(acento: nil, sobre: nil, oscuro: true).claroEncima)
        XCTAssertEqual(PaletaTienda.mezclar(0x000000, 0xFFFFFF, 0.5), 0x808080)
        XCTAssertEqual(PaletaTienda.leer("#2F5D50"), 0x2F5D50)
        XCTAssertNil(PaletaTienda.leer("verde"))
        XCTAssertNotNil(CodigoQR.mascara("https://almacen-lumbre.tulocalweb.com.ar"))
    }
    func testPantallaListaPedidos() { montar(ListaPedidos()) }
    func testPantallaListaProductos() { montar(ListaProductos()) }
    func testPantallaTinchoChat() { montar(TinchoChat()) }
    /// «Reportar respuesta»: el mail a seguridad@ lleva la pregunta y la respuesta, bien codificadas.
    func testReportarRespuestaDeTincho() throws {
        let cuerpo = TinchoChat.cuerpoReporte(pregunta: "¿Cómo vengo?", respuesta: "Vas 50 % & bien = ok", tienda: "t1")
        XCTAssertTrue(cuerpo.contains("Pregunta: ¿Cómo vengo?"))
        XCTAssertTrue(cuerpo.contains("Tienda: t1"))
        let url = try XCTUnwrap(TinchoChat.mailReporte(cuerpo))
        XCTAssertEqual(url.scheme, "mailto")
        XCTAssertTrue(url.absoluteString.hasPrefix("mailto:seguridad@tulocalweb.com.ar?subject="))
        XCTAssertFalse(url.absoluteString.contains("& bien"))
    }
    func testPantallaMas() { montar(Mas()) }
    func testPantallaDetallePedido() { montar(DetallePedido(numero: 1048)) }
    func testPantallaAvisos() { montar(Avisos()) }
    func testPantallaCuenta() { montar(Cuenta()) }
    /// Borrar la cuenta (guía 5.1.1(v)): la hoja se arma y sólo habilita con el mail de la persona.
    func testHojaBorrarCuenta() {
        montar(HojaBorrarCuenta())
        XCTAssertTrue(HojaBorrarCuenta.coincide(" Ana@Lumbre.com ", "ana@lumbre.com"))
        XCTAssertFalse(HojaBorrarCuenta.coincide("ana@lumbre", "ana@lumbre.com"))
        XCTAssertFalse(HojaBorrarCuenta.coincide("", ""))
    }
    func testPantallaClientes() { montar(Clientes()) }
    private var martina: ClienteFila {
        ClienteFila(id: "c1", nombre: "Martina Gómez", telefono: "+5492235551234", email: "martina@correo.com",
                    ciudad: "Mar del Plata", gastado: 96500, pedidos: 4, ultima: 3)
    }
    func testPantallaFichaCliente() { montar(FichaCliente(cliente: martina)) }
    func testHojaCorregirCliente() { montar(HojaCorregirCliente(cliente: martina) {}) }
    func testHojaBorrarCliente() { montar(HojaBorrarCliente(cliente: martina) {}) }
    /// Corregir un cliente: las mismas validaciones que la web antes de mandar «editar».
    func testCorregirClienteValida() {
        XCTAssertEqual(errorDatosCliente(nombre: " a ", telefono: ""), "Poné el nombre del cliente.")
        XCTAssertEqual(errorDatosCliente(nombre: "Martina", telefono: "5551"), "El WhatsApp tiene que tener la característica: 11 5555-1234.")
        XCTAssertNil(errorDatosCliente(nombre: "Martina Gómez", telefono: ""))
        XCTAssertNil(errorDatosCliente(nombre: "Martina Gómez", telefono: "11 5555-1234"))
    }
    func testPantallaCajaDelDia() { montar(CajaDelDia()) }
    func testPantallaCarritosAbandonados() { montar(CarritosAbandonados()) }
    func testPantallaResenasYPreguntas() { montar(ResenasYPreguntas()) }
    func testPantallaDevoluciones() { montar(Devoluciones()) }
    func testPantallaCupones() { montar(Cupones()) }
    func testPantallaFinanzas() { montar(Finanzas()) }
    func testPantallaFacturacionArca() { montar(FacturacionArca()) }
    func testPantallaPlanCuenta() { montar(PlanCuenta()) }
    /// El plan con prorrateo (§56): la respuesta de la API se decodifica y la hoja con el resumen se arma.
    static var plan: String { Self.respuestas()["/v1/panel/cuenta/plan"] ?? "{}" }
    func testHojaPlanConResumen() throws {
        let d = try JSONDecoder().decode(PlanCuenta.Datos.self, from: Data(Self.plan.utf8))
        XCTAssertEqual(d.cotizacion("pro", "mes")?.hoy, 14516)
        XCTAssertEqual(d.cotizacion("gratis", "anio")?.tipo, "gratis") // el gratis se cotiza con el ciclo vigente
        XCTAssertEqual(Formato.diaMes("2026-11-20T00:00:00Z").isEmpty, false)
        // Guía 3.1.1: desde la app no se pasa a un plan pago; sólo a Gratis o quedarse con el actual.
        XCTAssertFalse(HojaPlan.permitido(d, "pro"))
        XCTAssertFalse(HojaPlan.permitido(d, "negocio"))
        XCTAssertTrue(HojaPlan.permitido(d, "gratis"))
        XCTAssertTrue(HojaPlan.permitido(d, "emprende"))
        montar(HojaPlan(datos: d, preelegido: "pro") {})
        montar(HojaPlan(datos: d, preelegido: "gratis") {})
        // El descuento para quedarse (§57): mitad del débito mientras queden cobros.
        XCTAssertEqual(PlanCuenta.Datos.conDescuento(19900, d.beneficio), 9950)
        XCTAssertEqual(PlanCuenta.Datos.conDescuento(19900, nil), 19900)
        montar(HojaCancelar(datos: d) {})
    }
    func testGraficos() {
        let flujo: [GraficoFlujo.Dia] = [.init(dia: 1, entra: 100, sale: 30), .init(dia: 2, entra: 0, sale: 80)]
        montar(GraficoVentas(actual: [0, 3, 2, 5], anterior: [1, 1, 2, 2]), segundos: 1)
        montar(GraficoFlujo(dias: flujo), segundos: 1)
        montar(GraficoSaldo(flujo: flujo), segundos: 1)
        montar(DonaMedios(partes: [.init(nombre: "Efectivo", monto: 3), .init(nombre: "Mercado Pago", monto: 7)]), segundos: 1)
        montar(GraficoMeses(meses: ["sep", "oct", "nov"], series: [
            .init(id: "ventas", nombre: "Ventas", valores: [1, 2, nil], formato: { "\($0)" }, eje: { "\($0)" }),
        ]), segundos: 1)
        montar(EmbudoVisitas(paginas: [.init(pagina: "inicio", vistas: 100), .init(pagina: "checkout", vistas: 4)]), segundos: 1)
        montar(BarrasEstrellas(porEstrellas: [0, 1, 0, 3, 9]), segundos: 1)
        montar(AnilloAvance(valor: 0.62), segundos: 1)
    }
    func testPantallaUsuariosYPermisos() { montar(UsuariosYPermisos()) }
    func testPantallaEditorProductoExistente() { montar(EditorProducto(id: "p1")) }
    func testPantallaPedidoManual() { montar(PedidoManual()) }
    func testPantallaPreferenciasAvisos() { montar(PreferenciasAvisos()) }
    func testPantallaNegocioCobro() { montar(NegocioCobro()) }
    func testPantallaLinksPago() { montar(LinksPago()) }
    func testPantallaCentroAyuda() { montar(CentroAyuda()) }
    func testPantallaMiTiendaEditor() { montar(MiTiendaEditor()) }
    func testOnboarding() { montar(Onboarding()) }
    func testCrearTienda() { montar(CrearTienda()) }
    func testImportarTienda() { montar(ImportarTienda()) }
    func testTelegram() { montar(TelegramTincho()) }
    func testMedicion() { montar(MedicionTienda()) }
    func testIAConectadas() { montar(IAConectadas()) }
    func testBloquesDeTincho() throws {
        let json = """
        {"texto": "Hoy llevás **$ 539.000** en 15 pedidos.", "bloques": [
          {"tipo": "cifras", "items": [{"titulo": "Ventas", "valor": "$ 539.000", "detalle": "▲ 53 %", "tono": "ok"}, {"titulo": "Pedidos", "valor": "15"}]},
          {"tipo": "barras", "titulo": "Más vendidos", "unidad": "unidades", "items": [{"etiqueta": "Yerba", "valor": 208, "destacado": true}, {"etiqueta": "Mate", "valor": 120}]},
          {"tipo": "linea", "titulo": "Ventas por día", "unidad": "pesos", "puntos": [{"etiqueta": "lun 26", "valor": 500000}, {"etiqueta": "mar 27", "valor": 776400}, {"etiqueta": "mié 28", "valor": 344700}], "anterior": [400000, 600000, 300000]},
          {"tipo": "tabla", "titulo": "Detalle", "columnas": ["Producto", "Stock"], "filas": [["Yerba", "17"], ["Kit", "Sin stock"]]},
          {"tipo": "lista", "titulo": "Para reponer", "items": [{"titulo": "Kit matero", "detalle": "Se vende", "insignia": "Sin stock", "tono": "mal"}]},
          {"tipo": "progreso", "titulo": "Meta del mes", "actual": 11011500, "objetivo": 21600000, "unidad": "pesos", "detalle": "Proyección"},
          {"tipo": "nuevo-que-no-existe"}
        ], "acciones": [{"texto": "Ver pedidos", "ir": "pedidos"}], "sugerencias": ["¿Y ayer?"]}
        """
        let r = try JSONDecoder().decode(RespuestaTincho.self, from: Data(json.utf8))
        XCTAssertEqual(r.bloques?.count, 7)
        XCTAssertEqual(VistaBloque.corto(1_250_000, "pesos"), "$\u{202F}1,2 M")
        montar(ScrollView { VStack { ForEach(Array((r.bloques ?? []).enumerated()), id: \.offset) { _, b in VistaBloque(bloque: b) } } })
    }
    func testEstadisticas() { montar(Estadisticas()) }
    func testIntegraciones() { montar(Integraciones()) }
    func testHojaHistoria() { montar(HojaHistoria(tienda: Self.tienda)) }
    func testEditorProductoNuevoPreguntaElTipo() { montar(EditorProducto(id: nil)) }
    func testTipoDeLosProductosDeAntes() {
        XCTAssertEqual(TipoProducto.adivinar(["Talle": ["S", "M"]]), .ropa)
        XCTAssertEqual(TipoProducto.adivinar(["Talle": ["38", "39"], "Color": ["Negro"]]), .calzado)
        XCTAssertNil(TipoProducto.adivinar(["Sabor": ["Limón"]]))
        XCTAssertTrue(TipoProducto.hogar.campos.contains(.alto))
    }
    func testTarjetaHistoria() {
        let r = ImageRenderer(content: TarjetaHistoria(tienda: Self.tienda, portada: nil, logo: nil))
        r.scale = 3
        XCTAssertNotNil(r.uiImage)
    }
}
