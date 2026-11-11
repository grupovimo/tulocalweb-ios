import XCTest
@testable import TuLocalWeb

/// Cada link de `apps/pruebas/enlaces.json` que abre la app lleva a su pantalla (o a la web dentro de la app),
/// nunca a nada. La misma lista la sirve la web en el AASA y la prueba Android en EnlacesTest.kt.
final class EnlacesTests: XCTestCase {
    private struct Lista: Decodable { var dominio: String; var ejemplos: [String: String] }

    func testCadaLinkAbreSuPantalla() throws {
        let url = try XCTUnwrap(Bundle(for: EnlacesTests.self).url(forResource: "enlaces", withExtension: "json"),
                                "Falta enlaces.json en el paquete de pruebas")
        let lista = try JSONDecoder().decode(Lista.self, from: Data(contentsOf: url))
        for (camino, destino) in lista.ejemplos where destino != "web" {
            let r = Ruta(texto: "https://\(lista.dominio)\(camino)")
            XCTAssertEqual(r?.clave, destino, camino)
        }
    }

    func testOtroDominioNoEsDeLaApp() {
        XCTAssertNil(Ruta(texto: "https://otro.com/panel/pedidos/1"))
    }
}
