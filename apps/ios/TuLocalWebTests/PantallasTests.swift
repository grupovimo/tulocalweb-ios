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

    override func setUp() {
        let c = "{\"actual\": 12, \"anterior\": 4}"
        ApiFalsa.respuestas = [
            "/v1/panel/tablero": """
            {"periodo": "7d", "desde": "2026-10-25T00:00:00Z", "hasta": "2026-11-01T00:00:00Z", "montos": true,
             "kpis": {"pedidos": \(c), "clientesNuevos": \(c)}, "ultimosPedidos": [],
             "paraHacer": {"preparar": 0, "transferencias": 0, "sinStock": 0, "sinStockEnMl": 0},
             "hornero": {"carritos": 0, "ideasHechas": []}, "masVendidos": [],
             "visitas": {"desde": "2026-10-01", "hoy": \(c), "dias7": \(c), "dias30": \(c), "visitantes": \(c),
                         "vistas": \(c), "masVistos": []}}
            """,
            "/v1/panel/listas/productos": """
            {"total": 2, "segmentos": {"active": 2}, "filas": [
             {"id": "p1", "nombre": "Mate de calabaza con virola", "foto": "https://example.com/a.jpg", "precio": 18500,
              "stock": 3, "canales": ["web"], "publicado": true, "estado": "active", "variantes": []},
             {"id": "p2", "nombre": "Yerba", "precio": 4200, "stock": 0, "canales": [], "publicado": true,
              "estado": "out", "variantes": []}]}
            """,
            "/v1/panel/cupones": """
            {"cupones": [{"id": "c1", "codigo": "HOLA10", "tipo": "pct", "valor": 10, "descripcion": "",
              "estado": "activo", "usos": 0, "limite": 0, "vendio": 0}],
             "kpis": {"activos": 1, "usosMes": 0, "vendido": 0, "descontado": 0}}
            """,
            "/v1/panel/mi-tienda": "{\"diseno\": {\"portada\": null, \"logo\": null}}",
            "/v1/rubros": """
            [{"id": "indumentaria", "nombre": "Indumentaria", "plantilla": "ceibo", "tienda": "Ceibo"},
             {"id": "gastronomia", "nombre": "Gastronomía", "plantilla": "miga", "tienda": "Miga"}]
            """,
            "/v1/panel/onboarding": "{\"plantilla\": \"ceibo\", \"elecciones\": {\"rubro\": \"indumentaria\", \"pasoApp\": 2}, \"terminado\": false}",
            "/v1/panel/importaciones/actual": """
            {"id": "i1", "plataforma": "shopify", "urlOrigen": "https://x.myshopify.com", "estado": "corriendo", "etapa": "productos",
             "progreso": {"productos": 12, "clientes": 0, "pedidos": 0, "redirecciones": 0, "repetidos": 1, "recortados": 0},
             "errores": [], "creada": "2026-11-02T03:00:00Z", "terminada": null}
            """,
            "/v1/panel/telegram": "{\"vinculado\": false}",
            "/v1/panel/integraciones/medicion": "{\"ga4\": \"G-ABC1234\", \"pixel\": null, \"feeds\": {\"gm\": true, \"meta\": false}}",
            "/v1/panel/ia": """
            {"permisos": {"leer": true, "escribir": true}, "conexiones": [{"id": "c1", "cliente": "Claude", "persona": "Nahuel",
             "acciones": true, "creadaEn": "2026-11-01T10:00:00Z", "usadaEn": "2026-11-02T01:00:00Z"}]}
            """,
            "/v1/panel/productos/p1": """
            {"id": "p1", "nombre": "Mesa ratona", "descripcion": "", "categoria": null, "precio": 90000, "costo": 0,
             "sku": "", "stock": 2, "foto": null, "fotos": [], "canales": {"web": true}, "publicado": true,
             "vendidos30d": 0, "slug": "mesa-ratona", "tipo": "hogar", "atributos": {"peso": "8500", "alto": "45"},
             "opciones": [{"nombre": "Color", "valores": ["Natural", "Negro"]}],
             "variantes": [{"id": "v1", "valores": {"Color": "Natural"}, "sku": "", "precio": null, "stock": 1, "foto": null, "activa": true},
                           {"id": "v2", "valores": {"Color": "Negro"}, "sku": "", "precio": null, "stock": 1, "foto": null, "activa": true}]}
            """,
        ]
    }

    private func montar<V: View>(_ vista: V, segundos: Double = 4) {
        let conf = URLSessionConfiguration.ephemeral
        conf.protocolClasses = [ApiFalsa.self]
        let estado = Estado(api: Api(base: URL(string: "https://api.prueba")!, token: "x", http: URLSession(configuration: conf)))
        estado.acceso = Acceso(tiendaId: "t1", rol: "owner", permisos: ["catalogo", "ventas"], tiendas: [Self.tienda])
        let raiz = NavigationStack { vista }
            .environment(estado)
            .environment(Navegacion())
        let ventana = UIWindow(frame: CGRect(x: 0, y: 0, width: 390, height: 844))
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
    func testPantallaListaPedidos() { montar(ListaPedidos()) }
    func testPantallaListaProductos() { montar(ListaProductos()) }
    func testPantallaTinchoChat() { montar(TinchoChat()) }
    func testPantallaMas() { montar(Mas()) }
    func testPantallaDetallePedido() { montar(DetallePedido(numero: 1048)) }
    func testPantallaAvisos() { montar(Avisos()) }
    func testPantallaCuenta() { montar(Cuenta()) }
    func testPantallaClientes() { montar(Clientes()) }
    func testPantallaCajaDelDia() { montar(CajaDelDia()) }
    func testPantallaCarritosAbandonados() { montar(CarritosAbandonados()) }
    func testPantallaResenasYPreguntas() { montar(ResenasYPreguntas()) }
    func testPantallaDevoluciones() { montar(Devoluciones()) }
    func testPantallaCupones() { montar(Cupones()) }
    func testPantallaFinanzas() { montar(Finanzas()) }
    func testPantallaFacturacionArca() { montar(FacturacionArca()) }
    func testPantallaPlanCuenta() { montar(PlanCuenta()) }
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
