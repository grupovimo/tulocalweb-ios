import XCTest
@testable import TuLocalWeb

/// El diseño v2 de la tienda: los casos compartidos de la migración (packages/contrato/casos/diseno-v2.json),
/// el catálogo de bloques y los cambios del editor.
final class DisenoTests: XCTestCase {
    private struct Caso: Decodable { var nombre: String; var plantilla: String; var v1: ValorJSON; var v2: ValorJSON }
    private struct Casos: Decodable { var casos: [Caso] }

    func testCasosDeLaMigracion() throws {
        let url = try XCTUnwrap(Bundle(for: DisenoTests.self).url(forResource: "diseno-v2", withExtension: "json"))
        let casos = try JSONDecoder().decode(Casos.self, from: Data(contentsOf: url)).casos
        XCTAssertFalse(casos.isEmpty)
        for c in casos {
            let d = Diseno(json: c.v1.objeto ?? [:], plantillaTienda: c.plantilla)
            XCTAssertEqual(ValorJSON.objeto(d.json), c.v2, c.nombre)
        }
    }

    func testCatalogoYPresets() {
        let cat = CatalogoBloques.shared
        XCTAssertGreaterThan(cat.bloques.count, 15)
        XCTAssertEqual(PresetTienda.todos.count, 17)
        for t in cat.bloques {
            XCTAssertEqual(t.variantes.count, t.nombresVariantes.count, t.tipo)
            XCTAssertNotNil(cat.grupos.first { $0.id == t.grupo }, t.tipo)
        }
        for (id, p) in PresetTienda.todos {
            for b in p.inicio { XCTAssertNotNil(cat.tipo(b.tipo)?.variantes.firstIndex(of: b.variante), "\(id) \(b.tipo)") }
        }
    }

    func testCambiarDePresetConservaLosDatos() {
        var d = Diseno(json: [:], plantillaTienda: "ceibo")
        d.inicio[0].datos["titulo"] = .texto("Lumbre")
        d.inicio.append(Bloque(id: "abc12345", tipo: "banner", variante: "banda", datos: ["titulo": .texto("Hot Sale")]))
        d.estilo.direccion = "pedido"
        d.aplicarPreset("andes")
        XCTAssertEqual(d.plantilla, "andes")
        XCTAssertNil(d.estilo.direccion)
        XCTAssertEqual(d.inicio.first { $0.tipo == "portada" }?.datos["titulo"], .texto("Lumbre"))
        XCTAssertEqual(d.inicio.last?.id, "abc12345")
        XCTAssertEqual(Set(d.inicio.map(\.id)).count, d.inicio.count)
    }

    func testEditorDeshaceYRehace() {
        let e = EditorTienda()
        e.empezar(Diseno(json: [:], plantillaTienda: "litoral"), plantilla: "litoral")
        let antes = e.bloques.count
        let banner = CatalogoBloques.shared.tipo("banner")!
        let id = e.agregar(banner, en: 1, abrir: false)
        XCTAssertEqual(e.bloques.count, antes + 1)
        XCTAssertEqual(e.bloques[1].id, id)
        e.duplicar(id)
        e.borrar(id)
        XCTAssertEqual(e.bloques.count, antes + 1)
        e.deshacer()
        e.deshacer()
        XCTAssertEqual(e.bloques.count, antes + 1)
        e.deshacer()
        XCTAssertEqual(e.bloques.count, antes)
        e.rehacer()
        XCTAssertEqual(e.bloques[1].id, id)
        e.mover(id, antesDe: nil)
        XCTAssertEqual(e.bloques.last?.id, id)
        XCTAssertTrue(e.sinGuardar)
    }

    func testLoQueLeFaltaAUnBloque() {
        var ctx = ContextoTienda(nombre: "Lumbre", productos: 3)
        XCTAssertEqual(FaltaBloque.de(Bloque(id: "w", tipo: "whatsapp", variante: "banda"), ctx)?.enNegocio, true)
        ctx.whatsapp = true
        XCTAssertNil(FaltaBloque.de(Bloque(id: "w", tipo: "whatsapp", variante: "banda"), ctx))
        XCTAssertNotNil(FaltaBloque.de(Bloque(id: "b", tipo: "banner", variante: "banda"), ctx))
        XCTAssertNil(FaltaBloque.de(Bloque(id: "b", tipo: "banner", variante: "banda", datos: ["titulo": .texto("Hot Sale")]), ctx))
        XCTAssertNotNil(FaltaBloque.de(Bloque(id: "p", tipo: "preguntas", variante: "acordeon"), ctx))
    }
}
