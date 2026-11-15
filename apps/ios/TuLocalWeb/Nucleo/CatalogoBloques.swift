import Foundation

/// El catálogo único de bloques de la tienda (Recursos/bloques.json, copia de packages/contrato/bloques.json
/// hecha por scripts/estilo/para-ios.ts) y los puntos de partida (Recursos/presets.json). Los campos del
/// editor salen de acá: sumar un bloque o un campo en el contrato lo suma en la app sin tocar código.
struct CatalogoBloques: Decodable {
    struct Limites: Decodable {
        var bloquesPorPagina: Int
        var itemsPorLista: Int
        var bytesDiseno: Int
        /// Los topes de un campo tabla (la guía de talles).
        var filasTabla: Int?
        var columnasTabla: Int?
        var letrasCelda: Int?
        /// Cuántas ventanas (popups) puede tener la tienda.
        var ventanas: Int?
    }
    struct Opciones: Decodable { var opciones: [String]; var nombres: [String]; var porDefecto: String?; var descripcion: String? }
    struct Grupo: Decodable, Identifiable { var id: String; var nombre: String }

    /// Un campo de un bloque (tiposCampo en bloques.json): texto, foto, link, lista…
    struct Campo: Decodable, Identifiable {
        var id: String
        var tipo: String
        var nombre: String
        var max: Int?
        var min: Int?
        var requerido: Bool?
        var porDefecto: ValorJSON?
        /// «nombre» o «lema»: si está vacío, la tienda usa ese dato.
        var sugerido: String?
        /// Los campos de cada elemento (lista) o del objeto (grupo).
        var item: [Campo]?
        var opciones: [String]?
        var nombres: [String]?
        /// La explicación corta que va abajo del campo.
        var descripcion: String?
    }

    struct Tipo: Decodable, Identifiable {
        var tipo: String
        var grupo: String
        var nombre: String
        var max: Int?
        var descripcion: String
        var variantes: [String]
        var nombresVariantes: [String]
        var campos: [Campo]
        var id: String { tipo }

        func nombreVariante(_ v: String) -> String {
            variantes.firstIndex(of: v).map { nombresVariantes[$0] } ?? v
        }
    }

    struct Pagina: Decodable {
        var nombre: String
        var variantes: [String]?
        var nombresVariantes: [String]?
        var campos: [Campo]?
        var bloques: Bool?
    }

    struct Barra: Decodable {
        var nombre: String
        var descripcion: String
        var variantes: [String]
        var nombresVariantes: [String]
        var campos: [Campo]
    }

    var limites: Limites
    var estilo: [String: Opciones]
    var fondos: Opciones
    var fuentesProductos: Opciones
    var iconos: [String]
    var paginas: [String: Pagina]
    var grupos: [Grupo]
    var bloques: [Tipo]
    var barraAnuncios: Barra
    /// Las ventanas (popups) y las herramientas de conversión (docs/especificaciones/tienda-promociones.md).
    var ventanas: Ventanas?
    var conversion: Conversion?

    struct Ventanas: Decodable {
        struct Contenido: Decodable, Identifiable { var tipo: String; var nombre: String; var descripcion: String; var campos: [Campo]; var id: String { tipo } }
        struct Rango: Decodable { var min: Int; var max: Int; var porDefecto: Int; var unidad: String? }
        struct Disparo: Decodable, Identifiable { var cuando: String; var nombre: String; var descripcion: String; var valor: Rango?; var id: String { cuando } }
        struct Frecuencia: Decodable, Identifiable { var cada: String; var nombre: String; var descripcion: String; var dias: Rango?; var id: String { cada } }
        struct Plantilla: Decodable, Identifiable {
            var id: String
            var nombre: String
            var descripcion: String
            /// Los campos que la persona tiene que completar (un cupón suyo, el archivo).
            var completar: [String]?
            var hastaEnDias: Int?
            var ventana: ValorJSON
        }
        var nombre: String
        var descripcion: String
        var variantes: [String]
        var nombresVariantes: [String]
        var descripcionesVariantes: [String]?
        var campos: [Campo]
        var contenidos: [Contenido]
        var disparos: [Disparo]
        var frecuencias: [Frecuencia]
        var publicos: Opciones
        var paginas: Opciones
        var dispositivos: Opciones
        var plantillas: [Plantilla]

        func contenido(_ t: String) -> Contenido? { contenidos.first { $0.tipo == t } }
        func disparo(_ c: String) -> Disparo? { disparos.first { $0.cuando == c } }
        func frecuencia(_ c: String) -> Frecuencia? { frecuencias.first { $0.cada == c } }
        func nombreVariante(_ v: String) -> String { variantes.firstIndex(of: v).map { nombresVariantes[$0] } ?? v }
    }

    struct Conversion: Decodable { var nombre: String; var descripcion: String; var campos: [Campo] }

    /// Las páginas del editor, en orden (el JSON es un objeto y no guarda el orden).
    static let ordenPaginas = ["inicio", "catalogo", "producto", "carrito", "nosotros"]

    static let shared: CatalogoBloques = {
        guard let u = Bundle.main.url(forResource: "bloques", withExtension: "json"), let d = try? Data(contentsOf: u),
              let v = try? JSONDecoder().decode(CatalogoBloques.self, from: d) else {
            return CatalogoBloques(
                limites: .init(bloquesPorPagina: 30, itemsPorLista: 12, bytesDiseno: 65536), estilo: [:],
                fondos: .init(opciones: [], nombres: []), fuentesProductos: .init(opciones: [], nombres: []), iconos: [],
                paginas: [:], grupos: [], bloques: [],
                barraAnuncios: .init(nombre: "Barra de anuncios", descripcion: "", variantes: ["fija"], nombresVariantes: ["Fija"], campos: [])
            )
        }
        return v
    }()

    func tipo(_ t: String) -> Tipo? { bloques.first { $0.tipo == t } }
    func bloques(grupo: String) -> [Tipo] { bloques.filter { $0.grupo == grupo } }

    /// El nombre de una opción (fondos, fuentes de productos, estilo) o la misma si no está.
    static func nombre(_ o: Opciones, _ v: String) -> String {
        o.opciones.firstIndex(of: v).map { o.nombres[$0] } ?? v
    }

    /// Los ajustes nuevos de una página, con la primera variante y lo que trae cada campo por defecto.
    func ajustesNuevos(_ pagina: String) -> AjustesPagina {
        var a = AjustesPagina()
        guard let p = paginas[pagina] else { return a }
        a.variante = p.variantes?.first
        for c in p.campos ?? [] { if let d = c.porDefecto { a.campos[c.id] = d } }
        if p.bloques == true { a.bloques = [] }
        return a
    }
}

/// Un punto de partida (una de las 17 plantillas): dirección, estilo base, orden y variantes de los bloques.
struct PresetTienda: Decodable {
    var direccion: String
    var estilo: [String: String]
    var orden: [String]
    var variantes: [String: String]
    var gastronomia: Bool?

    private struct Archivo: Decodable { var presets: [String: PresetTienda] }

    static let todos: [String: PresetTienda] = {
        guard let u = Bundle.main.url(forResource: "presets", withExtension: "json"), let d = try? Data(contentsOf: u),
              let v = try? JSONDecoder().decode(Archivo.self, from: d) else { return [:] }
        return v.presets
    }()

    /// Los bloques del inicio recién armado con este punto de partida (id = tipo, como la migración).
    var inicio: [Bloque] {
        orden.map { Bloque(id: $0, tipo: $0, variante: variantes[$0] ?? CatalogoBloques.shared.tipo($0)?.variantes.first ?? "", datos: [:]) }
    }
}
