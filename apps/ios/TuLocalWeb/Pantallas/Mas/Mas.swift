import CoreImage
import SwiftUI

/// 8a · Más: la tarjeta de la tienda (cambiar, ver, compartir, QR) y los grupos Ventas y clientes,
/// Plata y Tu cuenta, con «Cerrar sesión». Lo que todavía no tiene pantalla en la app abre el panel
/// web adentro (SFSafariViewController), como lo que vive sólo en la web.
struct Mas: View {
    @Environment(Estado.self) private var estado
    @Environment(Navegacion.self) private var nav
    @State private var cambiarTienda = false
    @State private var qr = false
    @State private var salir = false
    @State private var cajaAbierta: Bool?

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 8) {
                EncabezadoGrande(titulo: "Más").padding(.top, 4)
                if let t = estado.tienda { tarjetaTienda(t).padding(.top, 4) }
                grupo("Ventas y clientes", tono: (C.pista, C.primarioTexto), filas: [
                    Fila(texto: "Clientes", simbolo: "person.2", destino: .clientes, permiso: "clientes"),
                    Fila(texto: "Caja del día", simbolo: "banknote", ruta: "ventas-en-local", permiso: "pedidos",
                         detalle: cajaAbierta.map { $0 ? ("Abierta", C.okTinta) : ("Cerrada", C.texto2) }),
                    Fila(texto: "Carritos abandonados", simbolo: "cart", ruta: "carritos-abandonados", permiso: "clientes",
                         insignia: estado.resumen?.insignias.carritos ?? 0),
                    Fila(texto: "Reseñas y preguntas", simbolo: "star.bubble", ruta: "resenas-y-preguntas", permiso: "clientes"),
                    Fila(texto: "Devoluciones", simbolo: "arrow.uturn.backward", ruta: "devoluciones", permiso: "pedidos"),
                    Fila(texto: "Cupones", simbolo: "ticket", ruta: "cupones", permiso: "catalogo"),
                ])
                grupo("Plata", tono: (C.okFondo, C.okTinta), filas: [
                    Fila(texto: "Finanzas", simbolo: "chart.line.uptrend.xyaxis", ruta: "finanzas", permiso: "finanzas"),
                    Fila(texto: "Facturación de ARCA", simbolo: "doc.text", ruta: "facturacion", permiso: "finanzas"),
                ])
                grupo("Tu cuenta", tono: (C.avisoFondo, C.avisoTinta), filas: [
                    Fila(texto: "Plan", simbolo: "creditcard", ruta: "plan", permiso: "cuenta", detalle: plan),
                    Fila(texto: "Integraciones", simbolo: "puzzlepiece.extension", ruta: "integraciones", permiso: "cuenta"),
                    Fila(texto: "Usuarios y permisos", simbolo: "person.badge.key", ruta: "usuarios-y-permisos", permiso: "cuenta"),
                    Fila(texto: "Perfil, seguridad y tema", simbolo: "person.crop.circle", destino: .cuenta),
                    Fila(texto: "Ayuda", simbolo: "questionmark.circle", ruta: "ayuda"),
                ])
                Button("Cerrar sesión") { salir = true }
                    .font(Estilo.fila.font).foregroundStyle(C.errorTinta)
                    .frame(maxWidth: .infinity, minHeight: 52)
                    .background(C.superficie, in: RoundedRectangle(cornerRadius: 16, style: .continuous))
                    .overlay(RoundedRectangle(cornerRadius: 16, style: .continuous).strokeBorder(C.anillo, lineWidth: 1))
                    .padding(.top, 8)
                Text("Tu Local Web \(Bundle.main.object(forInfoDictionaryKey: "CFBundleShortVersionString") as? String ?? "")")
                    .estilo(.footnote).foregroundStyle(C.texto3).frame(maxWidth: .infinity).padding(.top, 8)
            }
            .margenPantalla()
            .padding(.bottom, 24)
        }
        .background(C.fondo.ignoresSafeArea())
        .toolbar(.hidden, for: .navigationBar)
        .navigationTitle("Más")
        .task { await cargarCaja() }
        .sheet(isPresented: $cambiarTienda) {
            ElegirTienda { cambiarTienda = false }
                .presentationDetents([.height(520), .large])
                .presentationCornerRadius(30)
                .presentationBackground(C.elevado)
                .presentationDragIndicator(.hidden)
        }
        .sheet(isPresented: $qr) {
            if let t = estado.tienda { QRTienda(tienda: t).hojaTLW([.medium]) }
        }
        .confirmationDialog("¿Cerrás la sesión en este iPhone?", isPresented: $salir, titleVisibility: .visible) {
            Button("Cerrar sesión", role: .destructive) { Task { await estado.salir() } }
        } message: {
            Text("Dejás de recibir los avisos acá hasta que vuelvas a entrar.")
        }
    }

    private var plan: (String, Color)? {
        guard let p = estado.resumen?.prueba, p.diasRestantes > 0 else { return nil }
        return ("Prueba · \(p.diasRestantes) \(p.diasRestantes == 1 ? "día" : "días")", C.avisoTinta)
    }

    private func tarjetaTienda(_ t: Acceso.Tienda) -> some View {
        let url = URL(string: "https://\(t.direccion)")!
        return VStack(spacing: 14) {
            Button {
                if (estado.acceso?.tiendas.count ?? 0) > 1 { cambiarTienda = true }
            } label: {
                HStack(spacing: 12) {
                    AvatarIniciales(nombre: t.nombre, tam: 50, radio: 14)
                        .font(Estilo(tamano: 17, peso: .bold, relativo: .body).font)
                    VStack(alignment: .leading, spacing: 2) {
                        Text(t.nombre).estilo(.nombreTienda).foregroundStyle(C.texto).lineLimit(1)
                        Text(t.direccion).estilo(.texto14).foregroundStyle(C.texto2).lineLimit(1)
                    }
                    Spacer(minLength: 4)
                    if (estado.acceso?.tiendas.count ?? 0) > 1 {
                        Text("Cambiar").estilo(.label).foregroundStyle(C.primarioTexto)
                    }
                }
                .contentShape(Rectangle())
            }
            .buttonStyle(.plain)
            HStack(spacing: 8) {
                AccionRapida(simbolo: "arrow.up.right.square", texto: "Ver tienda", alto: 54) { nav.web = url }
                ShareLink(item: url, message: Text("Mirá mi tienda: \(t.nombre)")) {
                    VStack(spacing: 3) {
                        Image(systemName: "square.and.arrow.up").font(.system(size: 18)).foregroundStyle(C.primarioTexto)
                        Text("Compartir").estilo(.footnoteFuerte).foregroundStyle(C.primarioTinta)
                    }
                    .frame(maxWidth: .infinity, minHeight: 54)
                    .background(C.crema2, in: RoundedRectangle(cornerRadius: 14, style: .continuous))
                }
                .buttonStyle(.plain)
                AccionRapida(simbolo: "qrcode", texto: "QR", alto: 54) { qr = true }
            }
        }
        .padding(16)
        .tarjeta(radio: 22)
    }

    struct Fila: Identifiable {
        var texto: String
        var simbolo: String
        var ruta: String? = nil
        var destino: Destino? = nil
        var permiso: String? = nil
        var insignia = 0
        var detalle: (String, Color)? = nil
        var id: String { texto }
    }

    @ViewBuilder
    private func grupo(_ titulo: String, tono: (Color, Color), filas: [Fila]) -> some View {
        let visibles = filas.filter { $0.permiso.map(estado.puede) ?? true }
        if !visibles.isEmpty {
            EtiquetaSeccion(texto: titulo).padding(.horizontal, 4).padding(.top, 12)
            VStack(spacing: 0) {
                ForEach(Array(visibles.enumerated()), id: \.element.id) { i, f in
                    if i > 0 { Separador(inset: 58) }
                    Button { abrir(f) } label: {
                        HStack(spacing: 12) {
                            Image(systemName: f.simbolo).font(.system(size: 14, weight: .semibold)).foregroundStyle(tono.1)
                                .frame(width: 30, height: 30)
                                .background(tono.0, in: RoundedRectangle(cornerRadius: 8, style: .continuous))
                            Text(f.texto).estilo(.body16).foregroundStyle(C.texto)
                            Spacer(minLength: 4)
                            if let d = f.detalle {
                                Text(d.0).estilo(.label).foregroundStyle(d.1)
                            }
                            if f.insignia > 0 { Insignia(n: f.insignia) }
                            Chevron(ancho: 8)
                        }
                        .padding(.horizontal, 16).padding(.vertical, 12)
                        .contentShape(Rectangle())
                    }
                    .buttonStyle(.plain)
                }
            }
            .tarjeta()
        }
    }

    private func abrir(_ f: Fila) {
        if let d = f.destino {
            nav.mas.append(d)
        } else if let r = f.ruta {
            nav.web = URL(string: r == "ayuda" ? "https://tulocalweb.com.ar/ayuda" : "https://tulocalweb.com.ar/panel/\(r)")
        }
    }

    private func cargarCaja() async {
        struct T: Decodable { struct Caja: Decodable { var abierta: Date? }; var caja: Caja? }
        guard estado.puede("finanzas") else { return }
        if let t: T = try? await estado.api.get("/v1/panel/tablero", query: [URLQueryItem(name: "periodo", value: "hoy")]) {
            cajaAbierta = t.caja.map { $0.abierta != nil }
        }
    }
}

/// El QR de la tienda (para mostrar en el mostrador).
struct QRTienda: View {
    var tienda: Acceso.Tienda
    var body: some View {
        VStack(spacing: 16) {
            Grabber()
            Text(tienda.nombre).estilo(.tituloHojaChico).foregroundStyle(C.texto)
            if let img = QRTienda.imagen("https://\(tienda.direccion)") {
                Image(uiImage: img).interpolation(.none).resizable().scaledToFit()
                    .frame(width: 220, height: 220)
                    .padding(18)
                    .background(.white, in: RoundedRectangle(cornerRadius: 28, style: .continuous))
                    .shadow(color: Color(hex: 0x2A140C, alfa: 0.2), radius: 20, y: 14)
                    .accessibilityLabel("Código QR de \(tienda.direccion)")
            }
            Text(tienda.direccion).estilo(.subhead).foregroundStyle(C.texto2)
            Spacer(minLength: 0)
        }
        .padding(.horizontal, 20)
    }

    static func imagen(_ texto: String) -> UIImage? {
        let f = CIFilter(name: "CIQRCodeGenerator")
        f?.setValue(Data(texto.utf8), forKey: "inputMessage")
        f?.setValue("M", forKey: "inputCorrectionLevel")
        guard let out = f?.outputImage?.transformed(by: CGAffineTransform(scaleX: 10, y: 10)),
              let cg = CIContext().createCGImage(out, from: out.extent) else { return nil }
        return UIImage(cgImage: cg)
    }
}
