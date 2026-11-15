import Foundation
import SwiftUI

/// El estado del editor de la tienda: el diseño, la página elegida, el bloque abierto y el historial para
/// deshacer y rehacer. Todas las ediciones pasan por `cambiar`, así cada paso se puede deshacer.
@Observable
final class EditorTienda {
    var diseno = Diseno()
    /// Lo último guardado (para saber si hay cambios sin guardar).
    var guardado = Diseno()
    var plantillaTienda = ""
    var pagina = "inicio"
    /// El bloque resaltado en la lista y en la vista previa.
    var seleccion: String?
    /// El bloque con el panel abierto.
    var panel: String?
    /// Dónde va el bloque que se elija en la biblioteca.
    var insertarEn: Int?
    private(set) var historial = Historial()
    private var ultimaClave: String?
    private var ultimaVez = Date.distantPast

    private let cat = CatalogoBloques.shared

    var plantilla: String { diseno.plantilla ?? plantillaTienda }
    var sinGuardar: Bool { diseno != guardado }
    var bloques: [Bloque] { diseno.bloques(pagina) }
    var paginaConBloques: Bool { pagina == "inicio" || cat.paginas[pagina]?.bloques == true }

    func empezar(_ d: Diseno, plantilla: String) {
        diseno = d
        guardado = d
        plantillaTienda = plantilla
        historial = Historial()
    }

    // MARK: Cambios con historial

    /// Aplica un cambio guardando cómo estaba. Con `clave`, los cambios seguidos de lo mismo (escribir en un
    /// campo) quedan en un solo paso.
    func cambiar(clave: String? = nil, _ f: (inout Diseno) -> Void) {
        var d = diseno
        f(&d)
        guard d != diseno else { return }
        let ahora = Date()
        if clave == nil || clave != ultimaClave || ahora.timeIntervalSince(ultimaVez) > 1.5 {
            historial.guardar(diseno)
        }
        ultimaClave = clave
        ultimaVez = ahora
        diseno = d
    }

    func deshacer() {
        guard let d = historial.deshacer(diseno) else { return }
        ultimaClave = nil
        diseno = d
        cerrarSiNoEsta()
    }

    func rehacer() {
        guard let d = historial.rehacer(diseno) else { return }
        ultimaClave = nil
        diseno = d
        cerrarSiNoEsta()
    }

    private func cerrarSiNoEsta() {
        if let p = panel, diseno.buscar(p) == nil { panel = nil }
        if let s = seleccion, diseno.buscar(s) == nil { seleccion = nil }
    }

    /// Un campo del diseño para un control (Marca, Estilo), con su paso en el historial.
    func campo<T>(_ k: WritableKeyPath<Diseno, T>, clave: String) -> Binding<T> {
        Binding(get: { self.diseno[keyPath: k] }, set: { v in self.cambiar(clave: clave) { $0[keyPath: k] = v } })
    }

    // MARK: Bloques de la página elegida

    func bloque(_ id: String) -> Bloque? { diseno.buscar(id)?.bloque }

    /// Cambia un bloque esté en la página que esté.
    func editar(_ id: String, clave: String? = nil, _ f: @escaping (inout Bloque) -> Void) {
        guard let p = diseno.buscar(id)?.pagina else { return }
        cambiar(clave: clave) { d in
            var l = d.bloques(p)
            if let i = l.firstIndex(where: { $0.id == id }) { f(&l[i]) }
            d.ponerBloques(l, en: p)
        }
    }

    /// Si se puede sumar un bloque de este tipo a la página (tope por tipo y por página).
    func puedeAgregar(_ t: CatalogoBloques.Tipo) -> Bool {
        guard bloques.count < cat.limites.bloquesPorPagina else { return false }
        if let max = t.max { return bloques.filter { $0.tipo == t.tipo }.count < max }
        return true
    }

    /// Suma un bloque y abre su panel (con `abrir` false, lo abre quien llama: después de cerrar la biblioteca).
    @discardableResult
    func agregar(_ t: CatalogoBloques.Tipo, en posicion: Int, abrir: Bool = true) -> String {
        var b = Bloque.nuevo(t)
        while diseno.idsUsados.contains(b.id) { b.id = Bloque.nuevoId() }
        let nuevo = b
        cambiar { d in
            var l = d.bloques(self.pagina)
            l.insert(nuevo, at: min(max(0, posicion), l.count))
            d.ponerBloques(l, en: self.pagina)
        }
        seleccion = nuevo.id
        if abrir { panel = nuevo.id }
        return nuevo.id
    }

    func duplicar(_ id: String) {
        guard let r = diseno.buscar(id), let t = cat.tipo(r.bloque.tipo), puedeAgregar(t) || r.pagina != pagina else { return }
        let p = r.pagina
        var copia = r.bloque
        copia.id = Bloque.nuevoId()
        while diseno.idsUsados.contains(copia.id) { copia.id = Bloque.nuevoId() }
        let nueva = copia
        cambiar { d in
            var l = d.bloques(p)
            if let i = l.firstIndex(where: { $0.id == id }) { l.insert(nueva, at: i + 1) }
            d.ponerBloques(l, en: p)
        }
        seleccion = nueva.id
    }

    func alternarOculto(_ id: String) { editar(id) { $0.oculto.toggle() } }

    func borrar(_ id: String) {
        guard let p = diseno.buscar(id)?.pagina else { return }
        cambiar { d in d.ponerBloques(d.bloques(p).filter { $0.id != id }, en: p) }
        if panel == id { panel = nil }
        if seleccion == id { seleccion = nil }
    }

    /// Mueve un bloque arriba del otro (arrastrar y soltar); sin destino, al final.
    func mover(_ id: String, antesDe destino: String?) {
        guard id != destino else { return }
        cambiar { d in
            var l = d.bloques(self.pagina)
            guard let i = l.firstIndex(where: { $0.id == id }) else { return }
            let b = l.remove(at: i)
            let j = destino.flatMap { x in l.firstIndex { $0.id == x } } ?? l.count
            l.insert(b, at: j)
            d.ponerBloques(l, en: self.pagina)
        }
    }

    func mover(_ id: String, _ paso: Int) {
        cambiar { d in
            var l = d.bloques(self.pagina)
            guard let i = l.firstIndex(where: { $0.id == id }), l.indices.contains(i + paso) else { return }
            l.swapAt(i, i + paso)
            d.ponerBloques(l, en: self.pagina)
        }
    }

    func abrir(_ id: String) {
        if let p = diseno.buscar(id)?.pagina, p != pagina { pagina = p }
        seleccion = id
        panel = id
    }

    // MARK: Página y punto de partida

    func ajustes(_ p: String) -> AjustesPagina { diseno.ajustes(p) }

    func editarAjustes(_ p: String, clave: String? = nil, _ f: @escaping (inout AjustesPagina) -> Void) {
        cambiar(clave: clave) { d in
            var a = d.ajustes(p)
            f(&a)
            d.paginas[p] = a
        }
    }

    func aplicarPreset(_ id: String) {
        cambiar { $0.aplicarPreset(id) }
    }
}
