import SwiftUI

@main
struct TuLocalWebApp: App {
    @UIApplicationDelegateAdaptor(Delegado.self) private var delegado
    @State private var estado = Estado()
    @Environment(\.scenePhase) private var fase
    @AppStorage("tema") private var tema = "sistema"
    @State private var enSegundoPlano: Date?

    init() {
        Apariencia.aplicar()
    }

    var body: some Scene {
        WindowGroup {
            GeometryReader { g in
                Raiz()
                    .environment(estado)
                    .environment(\.anchoPantalla, g.size.width)
            }
            .preferredColorScheme(tema == "claro" ? .light : (tema == "oscuro" ? .dark : nil))
            .tint(C.primarioTexto)
            .onAppear { delegado.estado = estado }
            .onOpenURL { url in
                if let r = Ruta(texto: url.absoluteString) { estado.rutaPendiente = r }
            }
            .onChange(of: fase) { _, nueva in
                switch nueva {
                case .background:
                    enSegundoPlano = Date()
                case .active:
                    // Con Face ID prendido, volver después de un minuto pide identificarse.
                    if estado.conFaceID, estado.fase == .adentro, let d = enSegundoPlano, Date().timeIntervalSince(d) > 60 {
                        estado.bloqueada = true
                    }
                    enSegundoPlano = nil
                    if estado.fase == .adentro { Task { await estado.refrescarResumen() } }
                default:
                    break
                }
            }
        }
    }
}

/// Qué se ve según la sesión: splash, ingreso, selector de tienda, permisos o el panel.
struct Raiz: View {
    @Environment(Estado.self) private var estado

    var body: some View {
        ZStack {
            switch estado.fase {
            case .arrancando:
                Splash()
            case .afuera:
                Ingresar()
                    .transition(.opacity)
            case .eligiendoTienda:
                Ingresar(fondoDeSelector: true)
            case .permisos:
                Permisos()
                    .transition(.opacity)
            case .onboarding:
                Onboarding()
                    .transition(.opacity)
            case .adentro:
                Principal()
                    .transition(.opacity)
            }
            if estado.bloqueada {
                Bloqueo().transition(.opacity).zIndex(2)
            }
        }
        .animation(.easeInOut(duration: 0.25), value: estado.fase)
        .animation(.easeInOut(duration: 0.2), value: estado.bloqueada)
        .task { await estado.arrancar() }
    }
}
