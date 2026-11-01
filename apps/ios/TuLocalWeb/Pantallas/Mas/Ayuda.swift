import SwiftUI

/// El centro de ayuda adentro de la app: los mismos artículos que tulocalweb.com.ar/ayuda, que vienen
/// en Recursos/ayuda.json (se regeneran desde apps/web/src/lib/ayuda/articulos.ts).
struct CentroAyuda: View {
    @Environment(\.openURL) private var abrir
    @State private var busqueda = ""
    private let datos = DatosAyuda.cargar()

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 12) {
                Text("Ayuda").estilo(.tituloGrande).foregroundStyle(C.texto)
                Buscador(placeholder: "Buscá: envíos, cobrar, factura…", texto: $busqueda)
                if busqueda.trimmingCharacters(in: .whitespaces).isEmpty {
                    ForEach(datos.categorias) { c in
                        let arts = datos.articulos.filter { $0.categoria == c.slug }
                        if !arts.isEmpty {
                            EtiquetaSeccion(texto: c.nombre).padding(.horizontal, 4).padding(.top, 8)
                            lista(arts)
                        }
                    }
                } else {
                    let q = busqueda.lowercased().folding(options: .diacriticInsensitive, locale: nil)
                    let arts = datos.articulos.filter {
                        ($0.titulo + " " + $0.resumen).lowercased().folding(options: .diacriticInsensitive, locale: nil).contains(q)
                    }
                    if arts.isEmpty {
                        EstadoTincho(tincho: .lupa, titulo: "No encontramos «\(busqueda)»", texto: "Probá con otra palabra o escribinos.")
                            .padding(.top, 30)
                    } else {
                        lista(arts)
                    }
                }
                EtiquetaSeccion(texto: "¿No lo encontraste?").padding(.horizontal, 4).padding(.top, 10)
                HStack(spacing: 8) {
                    AccionRapida(simbolo: "envelope", texto: "Escribinos", alto: 58) {
                        if let u = URL(string: "mailto:hola@tulocalweb.com.ar?subject=Ayuda%20desde%20la%20app") { abrir(u) }
                    }
                }
            }
            .margenPantalla()
            .padding(.bottom, 24)
        }
        .background(C.fondo.ignoresSafeArea())
        .navigationBarTitleDisplayMode(.inline)
    }

    private func lista(_ arts: [DatosAyuda.Articulo]) -> some View {
        VStack(spacing: 0) {
            ForEach(Array(arts.enumerated()), id: \.element.id) { i, a in
                if i > 0 { Separador() }
                NavigationLink {
                    ArticuloAyuda(a: a)
                } label: {
                    HStack {
                        VStack(alignment: .leading, spacing: 2) {
                            Text(a.titulo).estilo(.fila).foregroundStyle(C.texto)
                            Text(a.resumen).estilo(.footnote).foregroundStyle(C.texto2).lineLimit(2)
                        }
                        Spacer(minLength: 6)
                        Chevron(ancho: 8)
                    }
                    .padding(.horizontal, 16).padding(.vertical, 12).contentShape(Rectangle())
                }
                .buttonStyle(.plain)
            }
        }
        .tarjeta()
    }
}

struct ArticuloAyuda: View {
    var a: DatosAyuda.Articulo
    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 12) {
                Text(Formato.sinViudas(a.titulo)).estilo(.tituloHoja).foregroundStyle(C.texto)
                Text(Formato.sinViudas(a.resumen)).estilo(.subhead).foregroundStyle(C.texto2)
                ForEach(Array(a.cuerpo.enumerated()), id: \.offset) { _, b in bloque(b) }
            }
            .margenPantalla()
            .padding(.bottom, 32)
        }
        .background(C.fondo.ignoresSafeArea())
        .navigationBarTitleDisplayMode(.inline)
    }

    @ViewBuilder
    private func bloque(_ b: DatosAyuda.Bloque) -> some View {
        switch b.t {
        case "h":
            Text(b.texto).estilo(.headline).foregroundStyle(C.texto).padding(.top, 6)
        case "pasos":
            VStack(alignment: .leading, spacing: 8) {
                ForEach(Array(b.lista.enumerated()), id: \.offset) { i, x in
                    HStack(alignment: .top, spacing: 10) {
                        Text("\(i + 1)").font(Estilo(tamano: 13, peso: .bold, relativo: .caption).font)
                            .foregroundStyle(C.primarioTinta).frame(width: 24, height: 24).background(C.crema, in: Circle())
                        Text(LocalizedStringKey(x)).estilo(.body16).foregroundStyle(C.texto).fixedSize(horizontal: false, vertical: true)
                    }
                }
            }
        case "lista":
            VStack(alignment: .leading, spacing: 6) {
                ForEach(Array(b.lista.enumerated()), id: \.offset) { _, x in
                    HStack(alignment: .top, spacing: 8) {
                        Text("•").foregroundStyle(C.acento)
                        Text(LocalizedStringKey(x)).estilo(.body16).foregroundStyle(C.texto).fixedSize(horizontal: false, vertical: true)
                    }
                }
            }
        case "nota":
            Text(LocalizedStringKey(b.texto)).estilo(.texto14).foregroundStyle(C.avisoTinta)
                .padding(12).frame(maxWidth: .infinity, alignment: .leading)
                .background(C.avisoFondo, in: RoundedRectangle(cornerRadius: 12, style: .continuous))
        default:
            Text(LocalizedStringKey(b.texto)).estilo(.body16).foregroundStyle(C.texto).fixedSize(horizontal: false, vertical: true)
        }
    }
}

struct DatosAyuda: Decodable {
    struct Categoria: Decodable, Identifiable { var slug: String; var nombre: String; var id: String { slug } }
    struct Bloque: Decodable {
        var t: String
        var x: ValorJSON
        var texto: String { x.texto ?? "" }
        var lista: [String] { x.lista?.compactMap(\.texto) ?? [] }
    }
    struct Articulo: Decodable, Identifiable {
        var slug: String
        var categoria: String
        var titulo: String
        var resumen: String
        var cuerpo: [Bloque]
        var id: String { slug }
    }
    var categorias: [Categoria]
    var articulos: [Articulo]

    static func cargar() -> DatosAyuda {
        guard let u = Bundle.main.url(forResource: "ayuda", withExtension: "json"),
              let d = try? Data(contentsOf: u),
              let v = try? JSONDecoder().decode(DatosAyuda.self, from: d) else { return DatosAyuda(categorias: [], articulos: []) }
        return v
    }
}

/// La etiqueta de envío (10 × 15 cm) armada en el celular, para imprimir o compartir como PDF.
struct EtiquetaEnvio: View {
    var p: PedidoDetalle
    var tienda: String
    var body: some View {
        VStack(alignment: .leading, spacing: 14) {
            Text("Pedido #\(p.numero)").font(.system(size: 30, weight: .bold))
            VStack(alignment: .leading, spacing: 4) {
                Text("PARA").font(.system(size: 12, weight: .bold)).foregroundStyle(.gray)
                Text(p.cliente).font(.system(size: 22, weight: .bold))
                if let e = p.entrega {
                    Text(e.direccion).font(.system(size: 18))
                    Text([e.localidad, e.provincia, "CP \(e.cp)"].compactMap { $0 }.joined(separator: " · ")).font(.system(size: 16))
                    if let s = e.sucursalNombre { Text("Sucursal: \(s)").font(.system(size: 16)) }
                }
            }
            Divider()
            VStack(alignment: .leading, spacing: 4) {
                Text("DE").font(.system(size: 12, weight: .bold)).foregroundStyle(.gray)
                Text(tienda).font(.system(size: 16, weight: .semibold))
            }
            Spacer()
            Text("PED-\(p.numero)").font(.system(size: 26, weight: .heavy, design: .monospaced))
                .frame(maxWidth: .infinity).padding(.vertical, 10).overlay(Rectangle().stroke(.black, lineWidth: 2))
        }
        .padding(24)
        .frame(width: 283, height: 425, alignment: .topLeading)
        .background(.white)
        .foregroundStyle(.black)
    }

    /// El PDF de la etiqueta, en un archivo temporal.
    @MainActor
    static func pdf(_ p: PedidoDetalle, tienda: String) -> URL? {
        let url = FileManager.default.temporaryDirectory.appendingPathComponent("etiqueta-\(p.numero).pdf")
        let r = ImageRenderer(content: EtiquetaEnvio(p: p, tienda: tienda))
        var ok = false
        r.render { tam, dibujar in
            var caja = CGRect(origin: .zero, size: tam)
            guard let ctx = CGContext(url as CFURL, mediaBox: &caja, nil) else { return }
            ctx.beginPDFPage(nil)
            dibujar(ctx)
            ctx.endPDFPage()
            ctx.closePDF()
            ok = true
        }
        return ok ? url : nil
    }
}
