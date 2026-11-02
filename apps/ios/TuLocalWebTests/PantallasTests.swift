import SwiftUI
import XCTest
@testable import TuLocalWeb

/// Monta pantallas enteras en una ventana para que un crash al abrirlas salte en el CI
/// (en la Mac no se usa el simulador).
@MainActor
final class PantallasTests: XCTestCase {
    private func montar<V: View>(_ vista: V, segundos: Double = 3) {
        let estado = Estado(api: Api(base: URL(string: "http://127.0.0.1:9")!, token: "x"))
        estado.acceso = Acceso(tiendaId: "t1", rol: "owner", permisos: ["catalogo", "ventas"],
                               tiendas: [.init(id: "t1", nombre: "Almacén Lumbre", slug: "almacen-lumbre", rol: "owner")])
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
    func testHojaHistoria() {
        montar(HojaHistoria(tienda: .init(id: "t1", nombre: "Almacén Lumbre", slug: "almacen-lumbre", rol: "owner")))
    }
}
