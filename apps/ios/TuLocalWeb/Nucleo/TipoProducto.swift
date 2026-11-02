import Foundation

/// Qué es el producto (API: migración 00071). El editor pregunta primero y según eso pide sus datos:
/// la ropa, talle y color; la comida, el contenido; un mueble, peso y medidas. La misma lista está en
/// Go (internal/dominio/listas/tipo.go) y en la web (src/lib/tienda/tipo-producto.ts).
enum TipoProducto: String, CaseIterable, Identifiable {
    case ropa, calzado, alimentos, belleza, hogar, tecnologia, otro

    var id: String { rawValue }

    var nombre: String {
        switch self {
        case .ropa: "Ropa y accesorios"
        case .calzado: "Calzado"
        case .alimentos: "Comida y bebida"
        case .belleza: "Belleza y cuidado"
        case .hogar: "Hogar, deco y objetos"
        case .tecnologia: "Tecnología"
        case .otro: "Otra cosa"
        }
    }

    var ejemplo: String {
        switch self {
        case .ropa: "Remeras, buzos, carteras"
        case .calzado: "Zapatillas, botas, sandalias"
        case .alimentos: "Yerba, vinos, budines"
        case .belleza: "Cremas, perfumes, maquillaje"
        case .hogar: "Muebles, mates, macetas"
        case .tecnologia: "Celulares, auriculares, cables"
        case .otro: "Lo que no entra en ninguno"
        }
    }

    var simbolo: String {
        switch self {
        case .ropa: "tshirt"
        case .calzado: "shoe"
        case .alimentos: "fork.knife"
        case .belleza: "sparkles"
        case .hogar: "sofa"
        case .tecnologia: "iphone"
        case .otro: "ellipsis.circle"
        }
    }

    /// Los datos que pide, en orden (alto, ancho y largo van juntos como «medidas»).
    var campos: [AtributoProducto] {
        switch self {
        case .ropa, .calzado: [.material, .peso]
        case .alimentos, .belleza: [.contenido, .marca, .peso]
        case .hogar: [.material, .peso, .alto, .ancho, .largo]
        case .tecnologia: [.marca, .modelo, .garantia, .peso, .alto, .ancho, .largo]
        case .otro: [.peso, .alto, .ancho, .largo]
        }
    }

    /// Las opciones de variante que propone, con valores para arrancar.
    var opciones: [(nombre: String, valores: [String])] {
        switch self {
        case .ropa: [("Talle", ["XS", "S", "M", "L", "XL", "XXL"]), ("Color", [])]
        case .calzado: [("Talle", ["35", "36", "37", "38", "39", "40", "41", "42", "43", "44"]), ("Color", [])]
        case .alimentos: [("Presentación", []), ("Sabor", [])]
        case .belleza: [("Presentación", []), ("Fragancia", [])]
        case .hogar: [("Color", []), ("Tamaño", ["Chico", "Mediano", "Grande"])]
        case .tecnologia: [("Color", []), ("Capacidad", [])]
        case .otro: [("Modelo", []), ("Color", [])]
        }
    }

    /// Para los productos de antes, sin tipo: lo que dicen sus opciones.
    static func adivinar(_ opciones: [String: [String]]) -> TipoProducto? {
        guard let talle = opciones.first(where: { $0.key.lowercased().hasPrefix("talle") })?.value else { return nil }
        let calzado = !talle.isEmpty && talle.allSatisfy { Int($0).map { (34...46).contains($0) } ?? false }
        return calzado ? .calzado : .ropa
    }
}

enum AtributoProducto: String {
    case peso, alto, ancho, largo, contenido, material, marca, modelo, garantia

    var nombre: String {
        switch self {
        case .peso: "Peso"
        case .alto: "Alto"
        case .ancho: "Ancho"
        case .largo: "Largo"
        case .contenido: "Contenido"
        case .material: "Material"
        case .marca: "Marca"
        case .modelo: "Modelo"
        case .garantia: "Garantía"
        }
    }

    /// La unidad de los numéricos (gramos y centímetros); nil: texto libre.
    var unidad: String? {
        switch self {
        case .peso: "g"
        case .alto, .ancho, .largo: "cm"
        default: nil
        }
    }

    var ejemplo: String {
        switch self {
        case .contenido: "Ej.: 500 g o 750 ml"
        case .material: "Ej.: algodón"
        case .marca: "La del fabricante o la tuya"
        case .modelo: "Ej.: A15"
        case .garantia: "Ej.: 6 meses"
        default: "0"
        }
    }

    var esMedida: Bool { self == .alto || self == .ancho || self == .largo }
}
