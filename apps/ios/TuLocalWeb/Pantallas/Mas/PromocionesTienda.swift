import SwiftUI

// Mi tienda › Promociones (docs/especificaciones/tienda-promociones.md § 4): las ventanas (popups) en orden, crear
// desde plantillas, el panel de cada una generado desde bloques.json › ventanas, sus métricas y las herramientas de
// conversión (estilo.conversion). Igual que el editor web.

/// La pestaña Promociones del editor.
struct PestanaPromociones: View {
    var editor: EditorTienda
    /// Abre la ventana en la vista previa de arriba.
    var verEnVista: (String) -> Void
    @Environment(DatosEditor.self) private var datos
    @Environment(Estado.self) private var estado
    private let cat = CatalogoBloques.shared.ventanas

    var body: some View {
        VStack(alignment: .leading, spacing: 14) {
            TarjetaSeccion(titulo: cat?.nombre ?? "Ventanas", simbolo: "sparkles") {
                let ventanas = editor.diseno.ventanas
                if ventanas.isEmpty {
                    vacio
                } else {
                    Text("Si varias pueden salir a la vez, sale la primera de la lista.").estilo(.footnote).foregroundStyle(C.texto2)
                    ForEach(Array(ventanas.enumerated()), id: \.element.id) { i, v in
                        FilaVentana(editor: editor, ventana: v, metrica: datos.metricas[v.id], primera: i == 0, ultima: i == ventanas.count - 1,
                                    verEnVista: verEnVista)
                    }
                    Button { editor.creandoVentana = true } label: { Label("Crear una ventana", systemImage: "plus").frame(maxWidth: .infinity) }
                        .buttonStyle(BotonSecundario()).disabled(!editor.puedeSumarVentana)
                    if !editor.puedeSumarVentana {
                        Text("Llegaste al tope de \(editor.topeVentanas) ventanas.").estilo(.footnote).foregroundStyle(C.texto2)
                    } else if datos.metricasCargadas {
                        Text("Las cifras son de los últimos 30 días.").estilo(.footnote).foregroundStyle(C.texto3)
                    }
                }
            }
            HerramientasConversion(editor: editor)
        }
    }

    private var vacio: some View {
        VStack(spacing: 12) {
            MiniaturaVentana(variante: "centrada", ancho: 92)
            Text("Todavía no tenés ventanas").estilo(.headline).foregroundStyle(C.texto)
            Text(cat?.descripcion ?? "Avisos que se abren solos.").estilo(.subhead).foregroundStyle(C.texto2).multilineTextAlignment(.center)
            Button("Crear la primera") { editor.creandoVentana = true }.buttonStyle(BotonPrincipal())
        }
        .frame(maxWidth: .infinity).padding(.vertical, 8)
    }
}

/// Una ventana en la lista: miniatura, nombre, qué muestra y cuándo, qué le falta o cómo le va, el interruptor y el menú.
struct FilaVentana: View {
    var editor: EditorTienda
    var ventana: VentanaTienda
    var metrica: MetricaVentana?
    var primera: Bool
    var ultima: Bool
    var verEnVista: (String) -> Void
    @Environment(Estado.self) private var estado
    private let cat = CatalogoBloques.shared.ventanas

    var body: some View {
        let falta = FaltaVentana.de(ventana)
        HStack(spacing: 10) {
            Button { editor.ventanaAbierta = ventana.id } label: {
                HStack(spacing: 12) {
                    MiniaturaVentana(variante: ventana.variante, ancho: 46)
                    VStack(alignment: .leading, spacing: 3) {
                        Text(ventana.nombre.isEmpty ? "Sin nombre" : ventana.nombre).estilo(.fila).foregroundStyle(C.texto).lineLimit(1)
                        Text(subtitulo).estilo(.footnote).foregroundStyle(C.texto2).lineLimit(1)
                        if let falta, ventana.activa {
                            Label(falta, systemImage: "exclamationmark.circle.fill").estilo(.footnoteFuerte).foregroundStyle(C.avisoTinta).lineLimit(2)
                        } else if let m = metrica, m.vistas > 0 {
                            Text(FilaVentana.cifras(m, ventana.contenido)).estilo(.footnote).monospacedDigit().foregroundStyle(C.texto3).lineLimit(1)
                        }
                    }
                    Spacer(minLength: 0)
                }
                .contentShape(Rectangle())
            }
            .buttonStyle(.plain)
            Toggle("", isOn: Binding(get: { ventana.activa }, set: { _ in Haptico.seleccion(); editor.alternarVentana(ventana.id) }))
                .labelsHidden().tint(C.ok)
                .accessibilityLabel(ventana.activa ? "Apagar \(ventana.nombre)" : "Prender \(ventana.nombre)")
            Menu {
                Button { editor.ventanaAbierta = ventana.id } label: { Label("Editar", systemImage: "pencil") }
                Button { verEnVista(ventana.id) } label: { Label("Ver en la vista previa", systemImage: "eye") }
                if editor.puedeSumarVentana {
                    Button { editor.duplicarVentana(ventana.id) } label: { Label("Duplicar", systemImage: "plus.square.on.square") }
                }
                if !primera { Button { editor.moverVentana(ventana.id, -1) } label: { Label("Subir", systemImage: "arrow.up") } }
                if !ultima { Button { editor.moverVentana(ventana.id, 1) } label: { Label("Bajar", systemImage: "arrow.down") } }
                Button(role: .destructive) { borrar() } label: { Label("Borrar", systemImage: "trash") }
            } label: {
                Image(systemName: "ellipsis").font(.system(size: 17, weight: .bold)).foregroundStyle(C.texto2)
                    .frame(width: 44, height: 44).contentShape(Rectangle())
            }
            .accessibilityLabel("Más acciones de \(ventana.nombre)")
        }
        .padding(.leading, 12).padding(.vertical, 10).padding(.trailing, 4)
        .background(C.superficie, in: RoundedRectangle(cornerRadius: 16, style: .continuous))
        .overlay(RoundedRectangle(cornerRadius: 16, style: .continuous).strokeBorder(C.linea))
        .opacity(ventana.activa ? 1 : 0.6)
    }

    private var subtitulo: String {
        let c = cat?.contenido(ventana.contenido)?.nombre ?? ventana.contenido
        let d = cat?.disparo(ventana.cuando)?.nombre ?? ventana.cuando
        return ventana.activa ? "\(c) · \(d)" : "Apagada · \(c)"
    }

    /// «320 vistas · 6 % dejó el mail · 4 ventas».
    static func cifras(_ m: MetricaVentana, _ contenido: String) -> String {
        var p = ["\(Formato.numero(m.vistas)) \(m.vistas == 1 ? "vista" : "vistas")"]
        if let t = m.tasa(contenido) {
            p.append("\(Formato.porcentaje(t.rounded())) \(contenido == "captacion" || contenido == "ruleta" ? "dejó el mail" : "tocó")")
        }
        if m.ventas > 0 { p.append("\(m.ventas) \(m.ventas == 1 ? "venta" : "ventas")") }
        return p.joined(separator: " · ")
    }

    private func borrar() {
        let nombre = ventana.nombre
        editor.borrarVentana(ventana.id)
        Haptico.impacto()
        estado.avisar("Borramos «\(nombre)»", deshacer: { [editor] in editor.deshacer() })
    }
}

/// Cómo se ve una ventana sobre la tienda: centrada, en un costado, completa o desde abajo.
struct MiniaturaVentana: View {
    var variante: String
    var ancho: CGFloat

    var body: some View {
        let alto = ancho * 1.3
        let r = ancho * 0.12
        ZStack {
            RoundedRectangle(cornerRadius: r, style: .continuous).fill(C.crema2)
            VStack(alignment: .leading, spacing: ancho * 0.07) {
                ForEach(0..<4, id: \.self) { i in
                    Capsule().fill(C.linea2).frame(width: ancho * (i % 2 == 0 ? 0.62 : 0.42), height: max(2, ancho * 0.06))
                }
                Spacer(minLength: 0)
            }
            .padding(ancho * 0.12).frame(maxWidth: .infinity, alignment: .leading)
            if variante != "lateral" { Color.black.opacity(0.18).clipShape(RoundedRectangle(cornerRadius: r, style: .continuous)) }
            tarjeta(alto: alto)
        }
        .frame(width: ancho, height: alto)
        .accessibilityHidden(true)
    }

    @ViewBuilder private func tarjeta(alto: CGFloat) -> some View {
        let t = RoundedRectangle(cornerRadius: ancho * 0.08, style: .continuous)
        let contenido = VStack(spacing: ancho * 0.05) {
            Capsule().fill(C.texto.opacity(0.7)).frame(width: ancho * 0.3, height: max(2, ancho * 0.05))
            Capsule().fill(C.acento).frame(width: ancho * 0.24, height: max(3, ancho * 0.09))
        }
        switch variante {
        case "lateral":
            contenido.frame(width: ancho * 0.48, height: alto * 0.24).background(C.superficie, in: t)
                .overlay(t.strokeBorder(C.linea2)).frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .bottomTrailing)
                .padding(ancho * 0.07)
        case "completa":
            contenido.frame(maxWidth: .infinity, maxHeight: .infinity).background(C.superficie, in: t).padding(ancho * 0.05)
        case "inferior":
            contenido.frame(maxWidth: .infinity).frame(height: alto * 0.42)
                .background(C.superficie, in: UnevenRoundedRectangle(topLeadingRadius: ancho * 0.12, topTrailingRadius: ancho * 0.12, style: .continuous))
                .frame(maxHeight: .infinity, alignment: .bottom)
        default:
            contenido.frame(width: ancho * 0.66, height: alto * 0.42).background(C.superficie, in: t)
        }
    }
}

// MARK: - Plantillas

/// Crear una ventana: las plantillas del catálogo y «En blanco».
struct PlantillasVentana: View {
    var editor: EditorTienda
    @Environment(DatosEditor.self) private var datos
    @Environment(\.dismiss) private var cerrar
    private let cat = CatalogoBloques.shared.ventanas

    static func simbolo(_ contenido: String) -> String {
        switch contenido {
        case "cupon": return "ticket"
        case "captacion": return "envelope"
        case "ruleta": return "gift"
        default: return "megaphone"
        }
    }

    var body: some View {
        NavigationStack {
            ScrollView {
                VStack(alignment: .leading, spacing: 10) {
                    Text("Elegí una para empezar: después cambiás todo.").estilo(.subhead).foregroundStyle(C.texto2)
                    ForEach(cat?.plantillas ?? []) { p in
                        let contenido = p.ventana["contenido"]?.texto ?? "promo"
                        tarjeta(p.nombre, p.descripcion, simbolo: Self.simbolo(contenido), variante: p.ventana["variante"]?.texto ?? "centrada") {
                            crear(VentanaTienda.dePlantilla(p, id: editor.idNuevo(), envioGratis: datos.envioGratis))
                        }
                    }
                    tarjeta("En blanco", "Una promo centrada al entrar. La armás a tu gusto.", simbolo: "plus", variante: "centrada") {
                        crear(VentanaTienda.enBlanco(id: editor.idNuevo()))
                    }
                }
                .margenPantalla().padding(.vertical, 12)
            }
            .background(C.fondo.ignoresSafeArea())
            .navigationTitle("Crear una ventana")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar { ToolbarItem(placement: .cancellationAction) { Button("Cancelar") { cerrar() } } }
        }
    }

    private func tarjeta(_ titulo: String, _ texto: String, simbolo: String, variante: String, _ accion: @escaping () -> Void) -> some View {
        Button(action: accion) {
            HStack(spacing: 14) {
                MiniaturaVentana(variante: variante, ancho: 46)
                VStack(alignment: .leading, spacing: 3) {
                    Label(titulo, systemImage: simbolo).estilo(.fila).foregroundStyle(C.texto)
                    Text(texto).estilo(.footnote).foregroundStyle(C.texto2).fixedSize(horizontal: false, vertical: true)
                }
                Spacer(minLength: 0)
                Chevron()
            }
            .padding(12)
            .background(C.superficie, in: RoundedRectangle(cornerRadius: 16, style: .continuous))
            .overlay(RoundedRectangle(cornerRadius: 16, style: .continuous).strokeBorder(C.linea))
        }
        .buttonStyle(.plain)
    }

    private func crear(_ v: VentanaTienda) {
        Haptico.seleccion()
        cerrar()
        editor.creandoVentana = false
        guard let id = editor.agregarVentana(v, abrir: false) else { return }
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.5) { [editor] in editor.ventanaAbierta = id }
    }
}

// MARK: - Panel de la ventana

/// Todo lo de una ventana: nombre, qué muestra, diseño, contenido, cuándo sale, a quién, dónde y fechas.
struct PanelVentana: View {
    var editor: EditorTienda
    var id: String
    var verEnVista: (String) -> Void
    @Environment(DatosEditor.self) private var datos
    @Environment(Estado.self) private var estado
    @Environment(\.dismiss) private var cerrar
    private let cat = CatalogoBloques.shared.ventanas

    var body: some View {
        NavigationStack {
            Group {
                if let v = editor.ventana(id), let cat {
                    contenido(v, cat)
                } else {
                    Text("Esta ventana ya no está.").estilo(.body16).foregroundStyle(C.texto2).frame(maxWidth: .infinity, maxHeight: .infinity)
                }
            }
            .background(C.elevado.ignoresSafeArea())
            .navigationBarTitleDisplayMode(.inline)
            .toolbar { ToolbarItem(placement: .confirmationAction) { Button("Listo") { cerrar() } } }
        }
    }

    private func editar(_ clave: String? = nil, _ f: @escaping (inout VentanaTienda) -> Void) { editor.editarVentana(id, clave: clave, f) }

    private func contenido(_ v: VentanaTienda, _ cat: CatalogoBloques.Ventanas) -> some View {
        let cont = cat.contenido(v.contenido)
        return ScrollView {
            VStack(alignment: .leading, spacing: 22) {
                VStack(alignment: .leading, spacing: 12) {
                    CampoLinea(titulo: cat.campos.first?.nombre ?? "Nombre", texto: Binding(
                        get: { editor.ventana(id)?.nombre ?? "" }, set: { n in editar("ventana-nombre-\(id)") { $0.nombre = String(n.prefix(40)) } }),
                               sugerencia: "Bienvenida 10 %", max: 40)
                    Interruptor(texto: "Prendida", prendido: Binding(get: { v.activa }, set: { a in editar { $0.activa = a } }))
                    if !v.activa {
                        AvisoPanel(texto: "Está apagada: no sale en tu tienda.", simbolo: "eye.slash") {}
                    } else if let f = FaltaVentana.de(v) {
                        AvisoPanel(texto: "\(f) Hasta entonces no sale.", simbolo: "exclamationmark.circle.fill") {}
                    }
                    Button { cerrar(); verEnVista(id) } label: { Label("Ver en la vista previa", systemImage: "eye").frame(maxWidth: .infinity) }
                        .buttonStyle(BotonTonal(alto: 44))
                }
                if let m = datos.metricas[id], m.vistas > 0 { metricas(m, v.contenido) }

                seccion("Qué muestra") {
                    ForEach(cat.contenidos) { c in
                        FilaOpcion(titulo: c.nombre, detalle: c.descripcion, simbolo: PlantillasVentana.simbolo(c.tipo), elegida: v.contenido == c.tipo) {
                            editar { $0.cambiarContenido(c.tipo) }
                        }
                    }
                }
                seccion("Diseño") {
                    ScrollView(.horizontal, showsIndicators: false) {
                        HStack(spacing: 10) {
                            ForEach(Array(cat.variantes.enumerated()), id: \.element) { i, va in
                                Button { Haptico.seleccion(); editar { $0.variante = va } } label: {
                                    VStack(alignment: .leading, spacing: 6) {
                                        MiniaturaVentana(variante: va, ancho: 76)
                                            .overlay(RoundedRectangle(cornerRadius: 11, style: .continuous)
                                                .strokeBorder(v.variante == va ? C.acento : .clear, lineWidth: 2.5).padding(-3))
                                        Text(cat.nombresVariantes.indices.contains(i) ? cat.nombresVariantes[i] : va).estilo(.footnoteFuerte)
                                            .foregroundStyle(v.variante == va ? C.texto : C.texto2).lineLimit(2).frame(width: 80, alignment: .leading)
                                    }
                                    .padding(3)
                                }
                                .buttonStyle(.plain)
                                .accessibilityAddTraits(v.variante == va ? .isSelected : [])
                            }
                        }
                    }
                    if let i = cat.variantes.firstIndex(of: v.variante), let d = cat.descripcionesVariantes, d.indices.contains(i) {
                        Text(d[i]).estilo(.footnote).foregroundStyle(C.texto2)
                    }
                }
                if let cont {
                    seccion(cont.nombre) {
                        EditorCampos(campos: cont.campos, datos: Binding(
                            get: { editor.ventana(id)?.datos ?? [:] },
                            set: { d in editar("ventana-\(id)") { $0.datos = d } }))
                    }
                }
                seccion("Cuándo sale") {
                    ForEach(cat.disparos) { d in
                        FilaOpcion(titulo: d.nombre, detalle: d.descripcion, elegida: v.cuando == d.cuando) { editar { $0.cuando = d.cuando } }
                    }
                    if let r = cat.disparo(v.cuando)?.valor {
                        rango(r.unidad == "%" ? "Qué parte de la página" : "Cuántos segundos", r, v.valor) { n in editar("ventana-valor-\(id)") { $0.valor = n } }
                    }
                    Text("Cuántas veces").estilo(.subheadFuerte).foregroundStyle(C.texto).padding(.top, 6)
                    ForEach(cat.frecuencias) { f in
                        FilaOpcion(titulo: f.nombre, detalle: f.descripcion, elegida: v.cada == f.cada) { editar { $0.cada = f.cada } }
                    }
                    if let r = cat.frecuencia(v.cada)?.dias {
                        rango("Cada cuántos días", r, v.dias) { n in editar("ventana-dias-\(id)") { $0.dias = n } }
                    }
                }
                seccion("A quién") {
                    SelectorOpcion(titulo: "Quiénes la ven", opciones: cat.publicos.opciones, nombres: cat.publicos.nombres,
                                   elegido: Binding(get: { v.publico }, set: { p in editar { $0.publico = p } }))
                    SelectorOpcion(titulo: "En qué pantallas", opciones: cat.dispositivos.opciones, nombres: cat.dispositivos.nombres,
                                   elegido: Binding(get: { v.dispositivos }, set: { p in editar { $0.dispositivos = p } }))
                }
                seccion("Dónde") {
                    SelectorOpciones(titulo: "En qué páginas", opciones: cat.paginas.opciones, nombres: cat.paginas.nombres,
                                     elegidas: Binding(get: { v.paginas }, set: { p in editar { $0.paginas = p } }))
                    if let d = cat.paginas.descripcion { Text(d).estilo(.footnote).foregroundStyle(C.texto2) }
                }
                seccion("Fechas") {
                    EditorFecha(titulo: "Empieza", valor: Binding(get: { v.desde.map { ValorJSON.texto($0) } }, set: { f in editar { $0.desde = f?.texto } }),
                                boton: "Elegir cuándo empieza", pasada: "Ya empezó.", quitable: true)
                    EditorFecha(titulo: "Termina", valor: Binding(get: { v.hasta.map { ValorJSON.texto($0) } }, set: { f in editar { $0.hasta = f?.texto } }),
                                boton: "Elegir cuándo termina", quitable: true)
                    Text("Sin fechas, sale mientras esté prendida.").estilo(.footnote).foregroundStyle(C.texto2)
                }
                Separador(inset: 0)
                Button { editor.duplicarVentana(id); estado.avisar("Duplicamos «\(v.nombre)»: quedó apagada") } label: {
                    Label("Duplicar", systemImage: "plus.square.on.square").frame(maxWidth: .infinity)
                }
                .buttonStyle(BotonTonal(alto: 44)).disabled(!editor.puedeSumarVentana)
                Button("Borrar esta ventana") {
                    let nombre = v.nombre
                    cerrar()
                    editor.borrarVentana(id)
                    estado.avisar("Borramos «\(nombre)»", deshacer: { [editor] in editor.deshacer() })
                }
                .buttonStyle(BotonTexto(alto: 44, color: C.errorTinta))
            }
            .padding(20)
        }
        .scrollDismissesKeyboard(.interactively)
    }

    private func seccion<V: View>(_ titulo: String, @ViewBuilder _ v: () -> V) -> some View {
        VStack(alignment: .leading, spacing: 10) {
            Text(titulo).estilo(.headline).foregroundStyle(C.texto)
            v()
        }
    }

    private func rango(_ titulo: String, _ r: CatalogoBloques.Ventanas.Rango, _ valor: Int?, _ poner: @escaping (Int) -> Void) -> some View {
        let n = valor ?? r.porDefecto
        let paso = r.max - r.min > 40 ? 5 : 1
        return Stepper(value: Binding(get: { n }, set: { poner(min(max($0, r.min), r.max)) }), in: r.min...r.max, step: paso) {
            HStack {
                Text(titulo).estilo(.body16).foregroundStyle(C.texto)
                Spacer()
                Text(r.unidad == "%" ? "\(n) %" : "\(n)").estilo(.fila).monospacedDigit().foregroundStyle(C.texto)
            }
        }
    }

    private func metricas(_ m: MetricaVentana, _ contenido: String) -> some View {
        let capta = contenido == "captacion" || contenido == "ruleta"
        return VStack(alignment: .leading, spacing: 10) {
            Text("Últimos 30 días").estilo(.footnoteFuerte).foregroundStyle(C.texto2)
            HStack(spacing: 8) {
                cifra("\(Formato.numero(m.vistas))", m.vistas == 1 ? "vista" : "vistas")
                cifra(m.tasa(contenido).map { Formato.porcentaje($0.rounded()) } ?? "—", capta ? "dejó el mail" : "tocó")
                cifra("\(m.ventas)", m.ventas == 1 ? "venta" : "ventas")
            }
            if m.monto > 0 { Text("Vendió \(Formato.pesos(m.monto)) con su cupón.").estilo(.footnote).foregroundStyle(C.texto2) }
        }
    }

    private func cifra(_ valor: String, _ texto: String) -> some View {
        VStack(alignment: .leading, spacing: 2) {
            Text(valor).estilo(.headline).monospacedDigit().foregroundStyle(C.texto).lineLimit(1).minimumScaleFactor(0.8)
            Text(texto).estilo(.footnote).foregroundStyle(C.texto2).lineLimit(1).minimumScaleFactor(0.8)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(12)
        .background(C.superficie, in: RoundedRectangle(cornerRadius: 14, style: .continuous))
    }
}

/// Una opción con su explicación, que se elige tocándola (contenido, disparo, frecuencia).
struct FilaOpcion: View {
    var titulo: String
    var detalle: String
    var simbolo: String?
    var elegida: Bool
    var accion: () -> Void

    var body: some View {
        Button { Haptico.seleccion(); accion() } label: {
            HStack(spacing: 12) {
                if let simbolo {
                    Image(systemName: simbolo).font(.system(size: 15, weight: .semibold)).foregroundStyle(C.primarioTexto).frame(width: 22)
                }
                VStack(alignment: .leading, spacing: 2) {
                    Text(titulo).estilo(.fila).foregroundStyle(C.texto)
                    Text(detalle).estilo(.footnote).foregroundStyle(C.texto2).fixedSize(horizontal: false, vertical: true)
                }
                Spacer(minLength: 0)
                Image(systemName: elegida ? "checkmark.circle.fill" : "circle").font(.system(size: 20)).foregroundStyle(elegida ? C.acento : C.linea2)
            }
            .padding(12)
            .background(C.superficie, in: RoundedRectangle(cornerRadius: 14, style: .continuous))
            .overlay(RoundedRectangle(cornerRadius: 14, style: .continuous).strokeBorder(elegida ? C.acento : C.linea, lineWidth: elegida ? 2 : 1))
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
        .accessibilityAddTraits(elegida ? .isSelected : [])
    }
}

// MARK: - Herramientas de conversión

/// estilo.conversion: interruptores con su explicación corta, generados desde bloques.json › conversion.
struct HerramientasConversion: View {
    var editor: EditorTienda

    var body: some View {
        if let c = CatalogoBloques.shared.conversion {
            TarjetaSeccion(titulo: c.nombre, simbolo: "bolt") {
                Text(c.descripcion).estilo(.footnote).foregroundStyle(C.texto2)
                EditorCampos(campos: c.campos, datos: Binding(
                    get: { editor.diseno.estilo.conversion ?? [:] },
                    set: { editor.editarConversion($0) }))
            }
        }
    }
}
