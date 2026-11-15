import SwiftUI

// Mi tienda › Páginas: elegir la página, su lista de bloques (arrastrar, agregar entre medio, duplicar,
// ocultar, borrar, deshacer y rehacer), la biblioteca y el panel de cada bloque.

/// La pestaña Páginas del editor.
struct ListaBloques: View {
    var editor: EditorTienda
    @Environment(DatosEditor.self) private var datos
    @Environment(Estado.self) private var estado
    private let cat = CatalogoBloques.shared

    var body: some View {
        VStack(alignment: .leading, spacing: 14) {
            ScrollView(.horizontal, showsIndicators: false) {
                HStack(spacing: 6) {
                    ForEach(CatalogoBloques.ordenPaginas, id: \.self) { p in
                        ChipFiltro(texto: cat.paginas[p]?.nombre.replacingOccurrences(of: "Ficha de producto", with: "Ficha") ?? p,
                                   cuenta: editor.diseno.bloques(p).count, activo: editor.pagina == p) {
                            editor.pagina = p
                            editor.seleccion = nil
                        }
                    }
                }
            }
            HStack(spacing: 8) {
                Text(ayuda).estilo(.footnote).foregroundStyle(C.texto2)
                Spacer(minLength: 0)
                Button { editor.deshacer() } label: { Image(systemName: "arrow.uturn.backward") }
                    .disabled(!editor.historial.puedeDeshacer).accessibilityLabel("Deshacer")
                Button { editor.rehacer() } label: { Image(systemName: "arrow.uturn.forward") }
                    .disabled(!editor.historial.puedeRehacer).accessibilityLabel("Rehacer")
            }
            .font(.system(size: 16, weight: .semibold)).foregroundStyle(C.primarioTexto)
            .buttonStyle(BotonIcono())

            if cat.paginas[editor.pagina]?.variantes != nil || cat.paginas[editor.pagina]?.campos != nil {
                AjustesPaginaVista(editor: editor, pagina: editor.pagina)
            }
            if editor.paginaConBloques { lista }
        }
    }

    private var ayuda: String {
        switch editor.pagina {
        case "inicio": return "Tocá un bloque para editarlo. Mantenelo apretado para moverlo."
        case "producto": return "Cómo se ve cada producto. Abajo de la ficha podés sumar bloques."
        case "nosotros": return "Contá quién está detrás de tu tienda."
        default: return "Cómo se ve esta página en tu tienda."
        }
    }

    @ViewBuilder private var lista: some View {
        let bloques = editor.bloques
        VStack(spacing: 0) {
            if editor.pagina == "producto" {
                Text("Debajo de la ficha").estilo(.subheadFuerte).foregroundStyle(C.texto)
                    .frame(maxWidth: .infinity, alignment: .leading).padding(.bottom, 8)
            }
            if bloques.isEmpty {
                vacio
            } else {
                ForEach(Array(bloques.enumerated()), id: \.element.id) { i, b in
                    insertar(i)
                    FilaBloque(editor: editor, bloque: b, falta: FaltaBloque.de(b, datos.contexto), primero: i == 0, ultimo: i == bloques.count - 1)
                }
                insertar(bloques.count)
                Button { editor.insertarEn = bloques.count } label: { Label("Sumar un bloque", systemImage: "plus") }
                    .buttonStyle(BotonSecundario()).padding(.top, 4)
                    .disabled(bloques.count >= cat.limites.bloquesPorPagina)
                if bloques.count >= cat.limites.bloquesPorPagina {
                    Text("Llegaste al tope de \(cat.limites.bloquesPorPagina) bloques en esta página.").estilo(.footnote).foregroundStyle(C.texto2)
                }
            }
        }
    }

    private var vacio: some View {
        VStack(spacing: 12) {
            MiniaturaBloque(tipo: sugerido, variante: cat.tipo(sugerido)?.variantes.first ?? "", ancho: 140)
            Text(editor.pagina == "nosotros" ? "Nosotros está vacía" : "Esta página no tiene bloques")
                .estilo(.headline).foregroundStyle(C.texto)
            Text(editor.pagina == "nosotros" ? "Empezá con un texto con foto: tu historia, el taller o el equipo."
                 : "Sumá el primero: preguntas frecuentes o beneficios ayudan a decidir.")
                .estilo(.subhead).foregroundStyle(C.texto2).multilineTextAlignment(.center)
            if let t = cat.tipo(sugerido) {
                Button("Sumar \(t.nombre.lowercased())") { editor.agregar(t, en: 0) }.buttonStyle(BotonPrincipal())
            }
            Button("Ver todos los bloques") { editor.insertarEn = 0 }.buttonStyle(BotonTexto(alto: 40))
        }
        .padding(18).frame(maxWidth: .infinity)
        .tarjeta(radio: 20)
    }

    private var sugerido: String { editor.pagina == "nosotros" ? "textoImagen" : (editor.pagina == "producto" ? "preguntas" : "portada") }

    /// El «+» entre dos bloques, para sumar uno justo ahí.
    private func insertar(_ i: Int) -> some View {
        Button { editor.insertarEn = i } label: {
            HStack(spacing: 8) {
                Rectangle().fill(C.linea).frame(height: 1)
                Image(systemName: "plus").font(.system(size: 11, weight: .bold)).foregroundStyle(C.primarioTexto)
                    .frame(width: 24, height: 24).background(C.crema, in: Circle())
                Rectangle().fill(C.linea).frame(height: 1)
            }
            .frame(height: 30).contentShape(Rectangle())
        }
        .buttonStyle(.plain)
        .accessibilityLabel(i == 0 ? "Sumar un bloque al principio" : "Sumar un bloque acá")
        .dropDestination(for: String.self) { ids, _ in
            guard let id = ids.first else { return false }
            let l = editor.bloques
            editor.mover(id, antesDe: l.indices.contains(i) ? l[i].id : nil)
            Haptico.seleccion()
            return true
        }
    }
}

/// Un botón de ícono de 44 × 44 (deshacer, rehacer, menú del bloque).
struct BotonIcono: ButtonStyle {
    func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .frame(width: 44, height: 44)
            .contentShape(Rectangle())
            .opacity(configuration.isPressed ? 0.5 : 1)
    }
}

/// Una fila de la lista: miniatura, nombre, variante, qué le falta y el menú.
struct FilaBloque: View {
    var editor: EditorTienda
    var bloque: Bloque
    var falta: FaltaBloque?
    var primero: Bool
    var ultimo: Bool
    @Environment(Estado.self) private var estado

    private var tipo: CatalogoBloques.Tipo? { CatalogoBloques.shared.tipo(bloque.tipo) }

    var body: some View {
        let elegido = editor.seleccion == bloque.id
        HStack(spacing: 10) {
            Button { editor.abrir(bloque.id) } label: {
                HStack(spacing: 12) {
                    Image(systemName: "line.3.horizontal").font(.system(size: 13, weight: .semibold)).foregroundStyle(C.texto3)
                    MiniaturaBloque(tipo: bloque.tipo, variante: bloque.variante, ancho: 66)
                    VStack(alignment: .leading, spacing: 3) {
                        Text(tipo?.nombre ?? bloque.tipo).estilo(.fila).foregroundStyle(C.texto).lineLimit(1)
                        Text(subtitulo).estilo(.footnote).foregroundStyle(C.texto2).lineLimit(1)
                        if let falta, !bloque.oculto {
                            Label(falta.texto, systemImage: "exclamationmark.circle.fill")
                                .estilo(.footnoteFuerte).foregroundStyle(C.avisoTinta).lineLimit(2)
                        }
                    }
                    Spacer(minLength: 0)
                }
                .contentShape(Rectangle())
            }
            .buttonStyle(.plain)
            Menu {
                Button { editor.abrir(bloque.id) } label: { Label("Editar", systemImage: "pencil") }
                Button { editor.duplicar(bloque.id) } label: { Label("Duplicar", systemImage: "plus.square.on.square") }
                Button { editor.alternarOculto(bloque.id) } label: {
                    Label(bloque.oculto ? "Mostrar" : "Ocultar", systemImage: bloque.oculto ? "eye" : "eye.slash")
                }
                if !primero { Button { editor.mover(bloque.id, -1) } label: { Label("Subir", systemImage: "arrow.up") } }
                if !ultimo { Button { editor.mover(bloque.id, 1) } label: { Label("Bajar", systemImage: "arrow.down") } }
                Button(role: .destructive) { borrar() } label: { Label("Borrar", systemImage: "trash") }
            } label: {
                Image(systemName: "ellipsis").font(.system(size: 17, weight: .bold)).foregroundStyle(C.texto2)
                    .frame(width: 44, height: 44).contentShape(Rectangle())
            }
            .accessibilityLabel("Más acciones de \(tipo?.nombre ?? "este bloque")")
        }
        .padding(.leading, 12).padding(.vertical, 10).padding(.trailing, 4)
        .background(C.superficie, in: RoundedRectangle(cornerRadius: 16, style: .continuous))
        .overlay(RoundedRectangle(cornerRadius: 16, style: .continuous).strokeBorder(elegido ? C.acento : C.linea, lineWidth: elegido ? 2 : 1))
        .opacity(bloque.oculto ? 0.6 : 1)
        .draggable(bloque.id) {
            MiniaturaBloque(tipo: bloque.tipo, variante: bloque.variante, ancho: 120)
        }
        .dropDestination(for: String.self) { ids, _ in
            guard let id = ids.first else { return false }
            editor.mover(id, antesDe: bloque.id)
            Haptico.seleccion()
            return true
        }
    }

    private var subtitulo: String {
        let v = tipo?.nombreVariante(bloque.variante) ?? bloque.variante
        let t = bloque.datos["titulo"]?.texto.map { " · \($0)" } ?? ""
        return bloque.oculto ? "Oculto · \(v)" : "\(v)\(t)"
    }

    private func borrar() {
        let nombre = tipo?.nombre ?? "el bloque"
        editor.borrar(bloque.id)
        Haptico.impacto()
        estado.avisar("Borramos «\(nombre)»", deshacer: { [editor] in editor.deshacer() })
    }
}

// MARK: - Biblioteca

/// Los bloques para sumar, por grupo y con su miniatura.
struct BibliotecaBloques: View {
    var editor: EditorTienda
    @Environment(\.dismiss) private var cerrar
    private let cat = CatalogoBloques.shared

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(alignment: .leading, spacing: 22) {
                    ForEach(cat.grupos) { g in
                        let tipos = cat.bloques(grupo: g.id)
                        if !tipos.isEmpty {
                            VStack(alignment: .leading, spacing: 10) {
                                Text(g.nombre).estilo(.headline).foregroundStyle(C.texto)
                                LazyVGrid(columns: [GridItem(.flexible(), spacing: 10), GridItem(.flexible(), spacing: 10)], spacing: 10) {
                                    ForEach(tipos) { t in tarjeta(t) }
                                }
                            }
                        }
                    }
                }
                .margenPantalla().padding(.vertical, 12)
            }
            .background(C.fondo.ignoresSafeArea())
            .navigationTitle("Sumar un bloque")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar { ToolbarItem(placement: .cancellationAction) { Button("Cancelar") { cerrar() } } }
        }
    }

    private func tarjeta(_ t: CatalogoBloques.Tipo) -> some View {
        let puede = editor.puedeAgregar(t)
        return Button {
            let en = editor.insertarEn ?? editor.bloques.count
            cerrar()
            Haptico.seleccion()
            editor.insertarEn = nil
            let id = editor.agregar(t, en: en, abrir: false)
            DispatchQueue.main.asyncAfter(deadline: .now() + 0.5) { [editor] in editor.panel = id }
        } label: {
            VStack(alignment: .leading, spacing: 6) {
                GeometryReader { g in MiniaturaBloque(tipo: t.tipo, variante: t.variantes.first ?? "", ancho: g.size.width) }
                    .aspectRatio(1 / 0.66, contentMode: .fit)
                Text(t.nombre).estilo(.subheadFuerte).foregroundStyle(C.texto).lineLimit(2)
                Text(puede ? t.descripcion : "Ya está en esta página").estilo(.footnote).foregroundStyle(C.texto2)
                    .lineLimit(3).fixedSize(horizontal: false, vertical: true)
                Spacer(minLength: 0)
            }
            .padding(10)
            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
            .background(C.superficie, in: RoundedRectangle(cornerRadius: 16, style: .continuous))
            .overlay(RoundedRectangle(cornerRadius: 16, style: .continuous).strokeBorder(C.linea))
            .opacity(puede ? 1 : 0.5)
        }
        .buttonStyle(.plain)
        .disabled(!puede)
    }
}

// MARK: - Panel del bloque

/// La variante con miniaturas, qué le falta y los campos generados desde bloques.json.
struct PanelBloque: View {
    var editor: EditorTienda
    var id: String
    @Environment(DatosEditor.self) private var datos
    @Environment(Estado.self) private var estado
    @Environment(\.dismiss) private var cerrar

    var body: some View {
        NavigationStack {
            Group {
                if let b = editor.bloque(id), let t = CatalogoBloques.shared.tipo(b.tipo) {
                    contenido(b, t)
                } else {
                    Text("Este bloque ya no está.").estilo(.body16).foregroundStyle(C.texto2)
                        .frame(maxWidth: .infinity, maxHeight: .infinity)
                }
            }
            .background(C.elevado.ignoresSafeArea())
            .navigationBarTitleDisplayMode(.inline)
            .toolbar { ToolbarItem(placement: .confirmationAction) { Button("Listo") { cerrar() } } }
        }
    }

    private func contenido(_ b: Bloque, _ t: CatalogoBloques.Tipo) -> some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 18) {
                VStack(alignment: .leading, spacing: 4) {
                    Text(t.nombre).estilo(.tituloHojaChico).foregroundStyle(C.texto)
                    Text(t.descripcion).estilo(.subhead).foregroundStyle(C.texto2)
                }
                if b.oculto {
                    AvisoPanel(texto: "Está oculto: no se ve en tu tienda.", simbolo: "eye.slash") {
                        Button("Mostrar") { editor.alternarOculto(id) }.buttonStyle(BotonTonal(alto: 34))
                    }
                } else if let f = FaltaBloque.de(b, datos.contexto) {
                    AvisoPanel(texto: f.texto, simbolo: "exclamationmark.circle.fill") {
                        if f.enNegocio {
                            NavigationLink("Cargar ahora") {
                                NegocioCobro().onDisappear { Task { await datos.cargarNegocio(estado.api) } }
                            }.buttonStyle(BotonTonal(alto: 34))
                        }
                    }
                }
                if t.variantes.count > 1 {
                    SelectorVariante(tipo: t.tipo, variantes: t.variantes, nombres: t.nombresVariantes, elegida: Binding(
                        get: { editor.bloque(id)?.variante ?? b.variante },
                        set: { v in editor.editar(id) { $0.variante = v } }
                    ))
                }
                EditorCampos(campos: t.campos, datos: Binding(
                    get: { editor.bloque(id)?.datos ?? [:] },
                    set: { d in editor.editar(id, clave: "bloque-\(id)") { $0.datos = d } }
                ))
                Separador()
                HStack(spacing: 8) {
                    Button { editor.alternarOculto(id) } label: {
                        Label(b.oculto ? "Mostrar" : "Ocultar", systemImage: b.oculto ? "eye" : "eye.slash").frame(maxWidth: .infinity)
                    }
                    Button { editor.duplicar(id) } label: { Label("Duplicar", systemImage: "plus.square.on.square").frame(maxWidth: .infinity) }
                }
                .buttonStyle(BotonTonal(alto: 44))
                Button("Borrar este bloque") {
                    let nombre = t.nombre
                    cerrar()
                    editor.borrar(id)
                    estado.avisar("Borramos «\(nombre)»", deshacer: { [editor] in editor.deshacer() })
                }
                .buttonStyle(BotonTexto(alto: 44, color: C.errorTinta))
            }
            .padding(20)
        }
        .scrollDismissesKeyboard(.interactively)
    }
}

/// El aviso del panel: por qué no se ve y, si aplica, el botón para arreglarlo.
struct AvisoPanel<Accion: View>: View {
    var texto: String
    var simbolo: String
    @ViewBuilder var accion: () -> Accion
    var body: some View {
        HStack(alignment: .center, spacing: 10) {
            Image(systemName: simbolo).font(.system(size: 16, weight: .semibold)).foregroundStyle(C.avisoTinta)
            Text(texto).estilo(.subheadFuerte).foregroundStyle(C.avisoTinta).fixedSize(horizontal: false, vertical: true)
            Spacer(minLength: 0)
            accion()
        }
        .padding(12)
        .background(C.avisoFondo, in: RoundedRectangle(cornerRadius: 14, style: .continuous))
    }
}

/// Las variantes de un bloque o de una página, con su miniatura.
struct SelectorVariante: View {
    var tipo: String
    var variantes: [String]
    var nombres: [String]
    @Binding var elegida: String

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Diseño").estilo(.footnoteFuerte).foregroundStyle(C.texto2)
            ScrollView(.horizontal, showsIndicators: false) {
                HStack(spacing: 10) {
                    ForEach(Array(variantes.enumerated()), id: \.element) { i, v in
                        Button {
                            Haptico.seleccion()
                            elegida = v
                        } label: {
                            VStack(alignment: .leading, spacing: 6) {
                                MiniaturaBloque(tipo: tipo, variante: v, ancho: 112)
                                    .overlay(RoundedRectangle(cornerRadius: 9, style: .continuous)
                                        .strokeBorder(elegida == v ? C.acento : .clear, lineWidth: 2.5).padding(-2))
                                Text(nombres.indices.contains(i) ? nombres[i] : v).estilo(.footnoteFuerte)
                                    .foregroundStyle(elegida == v ? C.texto : C.texto2)
                            }
                            .padding(2)
                        }
                        .buttonStyle(.plain)
                        .accessibilityAddTraits(elegida == v ? .isSelected : [])
                    }
                }
                .padding(.vertical, 2)
            }
        }
    }
}

// MARK: - Ajustes de la página

/// Catálogo, Ficha y Carrito: la variante y los campos de bloques.json › paginas.
struct AjustesPaginaVista: View {
    var editor: EditorTienda
    var pagina: String

    var body: some View {
        let p = CatalogoBloques.shared.paginas[pagina]
        TarjetaSeccion(titulo: p?.nombre ?? pagina, simbolo: "slider.horizontal.3") {
            if let v = p?.variantes, v.count > 1 {
                SelectorVariante(tipo: pagina, variantes: v, nombres: p?.nombresVariantes ?? v, elegida: Binding(
                    get: { editor.ajustes(pagina).variante ?? v[0] },
                    set: { n in editor.editarAjustes(pagina) { $0.variante = n } }
                ))
            }
            if let campos = p?.campos, !campos.isEmpty {
                EditorCampos(campos: campos, datos: Binding(
                    get: { editor.ajustes(pagina).campos },
                    set: { d in editor.editarAjustes(pagina, clave: "pagina-\(pagina)") { $0.campos = d } }
                ))
            }
        }
    }
}
