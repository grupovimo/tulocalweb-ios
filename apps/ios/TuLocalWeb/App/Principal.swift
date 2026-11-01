import Observation
import SafariServices
import SwiftUI

enum Pestana: Hashable, CaseIterable {
    case inicio, pedidos, productos, tincho, mas
}

/// Las pantallas a las que se entra desde una lista (push).
enum Destino: Hashable {
    case pedido(Int)
    case avisos
    case cuenta
    case clientes
    case cliente(ClienteFila)
    case caja, carritos, resenas, devoluciones, cupones, finanzas, facturacion, plan, integraciones, usuarios
}

/// La navegación de la app: la pestaña, el camino de cada una y los filtros con los que se abre una
/// lista desde otro lado (Inicio › Para hacer hoy).
@Observable
@MainActor
final class Navegacion {
    var pestana: Pestana = .inicio
    var inicio: [Destino] = []
    var pedidos: [Destino] = []
    var productos: [Destino] = []
    var tincho: [Destino] = []
    var mas: [Destino] = []
    /// Al abrir Pedidos con un filtro (estado de la API) o Productos con un segmento.
    var filtroPedidos: String?
    var segmentoProductos: String?
    /// La hoja «+» (2e).
    var crear = false
    /// La venta en el local (4a), a pantalla completa.
    var venta = false
    /// Una página web dentro de la app (lo que vive sólo en la web).
    var web: URL?

    var caminoActual: [Destino] {
        switch pestana {
        case .inicio: return inicio
        case .pedidos: return pedidos
        case .productos: return productos
        case .tincho: return tincho
        case .mas: return mas
        }
    }

    func ir(_ ruta: Ruta) {
        switch ruta {
        case let .pedido(n):
            pestana = .pedidos
            pedidos = [.pedido(n)]
        case .producto:
            pestana = .productos
            productos = []
        case .avisos:
            pestana = .inicio
            inicio = [.avisos]
        }
    }

    /// Tocar la pestaña activa vuelve a la raíz.
    func tocar(_ p: Pestana) {
        if p == pestana {
            switch p {
            case .inicio: inicio = []
            case .pedidos: pedidos = []
            case .productos: productos = []
            case .tincho: tincho = []
            case .mas: mas = []
            }
        }
        pestana = p
    }
}

/// El panel: 5 pestañas con su barra (la del handoff, no la del sistema), el «+» en Inicio y
/// Productos, y los toasts arriba de la barra.
struct Principal: View {
    @Environment(Estado.self) private var estado
    @State private var nav = Navegacion()

    var body: some View {
        @Bindable var nav = nav
        let enRaiz = nav.caminoActual.isEmpty
        ZStack(alignment: .bottom) {
            TabView(selection: $nav.pestana) {
                pila($nav.inicio) { Inicio() }.tag(Pestana.inicio)
                pila($nav.pedidos) { ListaPedidos() }.tag(Pestana.pedidos)
                pila($nav.productos) { ListaProductos() }.tag(Pestana.productos)
                pila($nav.tincho) { TinchoChat() }.tag(Pestana.tincho)
                pila($nav.mas) { Mas() }.tag(Pestana.mas)
            }
            .toolbar(.hidden, for: .tabBar)

            if enRaiz && (nav.pestana == .inicio || nav.pestana == .productos) {
                BotonMas { nav.crear = true }
                    .frame(maxWidth: .infinity, alignment: .trailing)
                    .padding(.trailing, 20)
                    .padding(.bottom, M.tabBar + 16)
                    .transition(.scale.combined(with: .opacity))
            }

            if let aviso = estado.aviso {
                VistaAviso(aviso: aviso) { estado.aviso = nil }
                    .padding(.horizontal, 16)
                    .padding(.bottom, enRaiz ? M.tabBar + 17 : 84)
                    .transition(.move(edge: .bottom).combined(with: .opacity))
                    .zIndex(3)
            }

            if enRaiz {
                BarraTabs(pestana: nav.pestana, insignia: estado.resumen?.insignias.pedidos ?? 0) { nav.tocar($0) }
                    .transition(.move(edge: .bottom))
            }
        }
        .animation(.easeOut(duration: 0.2), value: enRaiz)
        .animation(.spring(duration: 0.3), value: estado.aviso)
        .ignoresSafeArea(.keyboard, edges: .bottom)
        .environment(nav)
        .sheet(isPresented: $nav.crear) {
            HojaCrear().hojaTLW([.height(420)])
        }
        .fullScreenCover(isPresented: $nav.venta) {
            VentaLocal()
                .overlay(alignment: .bottom) {
                    if let aviso = estado.aviso {
                        VistaAviso(aviso: aviso) { estado.aviso = nil }
                            .padding(.horizontal, 16).padding(.bottom, 84)
                            .transition(.move(edge: .bottom).combined(with: .opacity))
                    }
                }
                .animation(.spring(duration: 0.3), value: estado.aviso)
        }
        .sheet(item: Binding(get: { nav.web.map(PaginaWeb.init) }, set: { nav.web = $0?.url })) { p in
            Safari(url: p.url).ignoresSafeArea()
        }
        .onChange(of: estado.rutaPendiente) { _, r in
            if let r {
                nav.ir(r)
                estado.rutaPendiente = nil
            }
        }
        .onAppear {
            if let r = estado.rutaPendiente {
                nav.ir(r)
                estado.rutaPendiente = nil
            }
        }
    }

    private func pila<V: View>(_ camino: Binding<[Destino]>, @ViewBuilder raiz: () -> V) -> some View {
        NavigationStack(path: camino) {
            raiz()
                .safeAreaInset(edge: .bottom, spacing: 0) { Color.clear.frame(height: M.tabBar) }
                .navigationDestination(for: Destino.self) { d in
                    switch d {
                    case let .pedido(n): DetallePedido(numero: n)
                    case .avisos: Avisos()
                    case .cuenta: Cuenta()
                    case .clientes: Clientes()
                    case let .cliente(c): FichaCliente(cliente: c)
                    case .caja: CajaDelDia()
                    case .carritos: CarritosAbandonados()
                    case .resenas: ResenasYPreguntas()
                    case .devoluciones: Devoluciones()
                    case .cupones: Cupones()
                    case .finanzas: Finanzas()
                    case .facturacion: FacturacionArca()
                    case .plan: PlanCuenta()
                    case .integraciones: Integraciones()
                    case .usuarios: UsuariosYPermisos()
                    }
                }
        }
    }
}

struct PaginaWeb: Identifiable {
    var url: URL
    var id: String { url.absoluteString }
}

/// SFSafariViewController: lo que vive sólo en la web (editor, dominio, invitar).
struct Safari: UIViewControllerRepresentable {
    var url: URL
    func makeUIViewController(context: Context) -> SFSafariViewController {
        let c = SFSafariViewController(url: url)
        c.preferredControlTintColor = .tema(0xB0441D, 0xE07A48)
        return c
    }
    func updateUIViewController(_ c: SFSafariViewController, context: Context) {}
}

/// La barra de pestañas del handoff: 49 + safe area, `bg-elevated` al 92 % con blur, borde de 0,5.
struct BarraTabs: View {
    var pestana: Pestana
    var insignia: Int
    var tocar: (Pestana) -> Void
    @Environment(\.colorScheme) private var esquema

    var body: some View {
        HStack(spacing: 0) {
            item(.inicio, "Inicio", simbolo: "house")
            item(.pedidos, "Pedidos", simbolo: "list.bullet.rectangle.portrait", insignia: insignia)
            item(.productos, "Productos", simbolo: "shippingbox")
            item(.tincho, "Tincho", simbolo: nil)
            item(.mas, "Más", simbolo: "ellipsis")
        }
        .padding(.horizontal, 6)
        .padding(.top, 6)
        .frame(height: M.tabBar, alignment: .top)
        .background {
            ZStack {
                Rectangle().fill(.ultraThinMaterial)
                Rectangle().fill(Color.tema(0xFBF7F2, 0x22140D, alfa: 0.92))
            }
            .overlay(alignment: .top) { Rectangle().fill(C.linea2).frame(height: 0.5) }
            .ignoresSafeArea(edges: .bottom)
        }
    }

    private func item(_ p: Pestana, _ texto: String, simbolo: String?, insignia: Int = 0) -> some View {
        let activo = p == pestana
        return Button {
            tocar(p)
            Haptico.seleccion()
        } label: {
            VStack(spacing: 3) {
                Group {
                    if let simbolo {
                        Image(systemName: simbolo)
                            .font(.system(size: 21, weight: .regular))
                            .frame(width: 27, height: 27)
                    } else {
                        Image(Tincho.avatar.imagen).resizable().scaledToFill()
                            .frame(width: 23, height: 23)
                            .background(C.crema)
                            .clipShape(Circle())
                            .padding(2)
                            .overlay(Circle().strokeBorder(activo ? C.primarioTexto : .clear, lineWidth: 2))
                            .frame(width: 27, height: 27)
                    }
                }
                .overlay(alignment: .topLeading) {
                    if insignia > 0 {
                        Text("\(insignia)")
                            .font(Estilo(tamano: 11.5, peso: .bold, relativo: .caption2).font)
                            .monospacedDigit()
                            .foregroundStyle(.white)
                            .padding(.horizontal, 5)
                            .frame(minWidth: 18, minHeight: 18)
                            .background(C.error, in: Capsule())
                            .overlay(Capsule().strokeBorder(Color.tema(0xFBF7F2, 0x22140D), lineWidth: 2).padding(-2))
                            .offset(x: 18, y: -5)
                    }
                }
                Text(texto).estilo(.tab)
            }
            .foregroundStyle(activo ? C.primarioTexto : C.texto3)
            .frame(maxWidth: .infinity)
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
        .accessibilityLabel(insignia > 0 ? "\(texto), \(insignia) por preparar" : texto)
        .accessibilityAddTraits(activo ? .isSelected : [])
    }
}

/// El «+» (58 Ø, gradiente, sombra).
struct BotonMas: View {
    var accion: () -> Void
    var body: some View {
        Button(action: accion) {
            Image(systemName: "plus").font(.system(size: 22, weight: .semibold)).foregroundStyle(.white)
                .frame(width: 58, height: 58)
                .background(C.primario, in: Circle())
                .shadow(color: C.sombraPrimario.opacity(0.7), radius: 10, y: 8)
        }
        .buttonStyle(.plain)
        .accessibilityLabel("Crear")
    }
}

/// 2e · La hoja «+»: Venta en el local, Producto con la cámara, Pedido manual.
struct HojaCrear: View {
    @Environment(\.dismiss) private var cerrar
    @Environment(Estado.self) private var estado
    @Environment(Navegacion.self) private var nav

    var body: some View {
        VStack(spacing: 10) {
            Grabber()
            HStack {
                Text("Crear").estilo(.tituloHojaChico).foregroundStyle(C.texto)
                Spacer()
                Button { cerrar() } label: {
                    Image(systemName: "xmark").font(.system(size: 12, weight: .bold)).foregroundStyle(C.texto2)
                        .frame(width: 30, height: 30)
                        .background(Color.tema(0x3A1D12, 0xFBF7F2, alfa: 0.08, alfaOscuro: 0.1), in: Circle())
                }
                .accessibilityLabel("Cerrar")
            }
            .padding(.horizontal, 6).padding(.top, 6).padding(.bottom, 0)
            opcion("Venta en el local", "Escaneá, armá el ticket y cobrá", simbolo: "barcode.viewfinder", principal: true) {
                cerrar()
                DispatchQueue.main.asyncAfter(deadline: .now() + 0.35) { nav.venta = true }
            }
            opcion("Producto con la cámara", "Sacale fotos y cargalo en un minuto", simbolo: "camera", principal: false) {
                web("productos")
            }
            opcion("Pedido manual", "Para ventas por WhatsApp o teléfono", simbolo: "doc.text", principal: false) {
                web("pedidos/nuevo")
            }
            Spacer(minLength: 0)
        }
        .padding(.horizontal, 20)
    }

    /// Mientras estas pantallas no están en la app, se abren las del panel web.
    private func web(_ ruta: String) {
        cerrar()
        nav.web = URL(string: "https://tulocalweb.com.ar/panel/\(ruta)")
    }

    private func opcion(_ titulo: String, _ texto: String, simbolo: String, principal: Bool, accion: @escaping () -> Void) -> some View {
        Button(action: accion) {
            HStack(spacing: 14) {
                Image(systemName: simbolo).font(.system(size: 21, weight: .medium))
                    .foregroundStyle(principal ? .white : C.primarioTexto)
                    .frame(width: 48, height: 48)
                    .background {
                        if principal {
                            RoundedRectangle(cornerRadius: 14, style: .continuous).fill(C.primario)
                        } else {
                            RoundedRectangle(cornerRadius: 14, style: .continuous).fill(C.crema)
                        }
                    }
                VStack(alignment: .leading, spacing: 2) {
                    Text(titulo).estilo(.headline).foregroundStyle(C.texto)
                    Text(texto).estilo(.texto14).foregroundStyle(C.texto2)
                }
                Spacer(minLength: 0)
            }
            .padding(16)
            .tarjeta()
        }
        .buttonStyle(.plain)
    }
}
