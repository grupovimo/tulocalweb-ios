import UIKit
import UserNotifications

/// El delegado de la app para APNs: recibe el token (y lo registra en /v1/panel/dispositivos), muestra
/// los avisos con la app abierta y abre la pantalla exacta al tocarlos (`ruta` del payload, §40).
final class Delegado: NSObject, UIApplicationDelegate, UNUserNotificationCenterDelegate {
    /// Lo asigna la app en `.onAppear`. Con la app cerrada, el aviso tocado llega antes: queda en `pendiente`
    /// y se atiende apenas está (si no, se perdía y la app abría en Inicio).
    weak var estado: Estado? {
        didSet {
            guard let estado, let p = pendiente else { return }
            pendiente = nil
            atender(p, en: estado)
        }
    }
    private var pendiente: Tocado?

    /// Un aviso tocado: adónde lleva, de qué tienda y si fue «Marcar como preparado».
    private struct Tocado { var ruta: Ruta; var tienda: String?; var marcar: Bool }

    func application(_ application: UIApplication,
                     didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]? = nil) -> Bool {
        UNUserNotificationCenter.current().delegate = self
        // Las acciones del aviso ampliado (7b): «Ver pedido» abre la app; «Marcar como preparado»
        // pide desbloquear el iPhone.
        let ver = UNNotificationAction(identifier: "ver", title: "Ver pedido", options: [.foreground])
        let preparado = UNNotificationAction(identifier: "preparado", title: "Marcar como preparado",
                                             options: [.authenticationRequired, .foreground])
        UNUserNotificationCenter.current().setNotificationCategories([
            UNNotificationCategory(identifier: "venta", actions: [ver, preparado], intentIdentifiers: []),
        ])
        return true
    }

    func application(_ application: UIApplication, didRegisterForRemoteNotificationsWithDeviceToken deviceToken: Data) {
        let hex = deviceToken.map { String(format: "%02x", $0) }.joined()
        Task { @MainActor in await self.estado?.registrarTokenPush(hex) }
    }

    func application(_ application: UIApplication, didFailToRegisterForRemoteNotificationsWithError error: Error) {
        #if DEBUG
        print("APNs: \(error.localizedDescription)")
        #endif
    }

    // Con la app abierta, el aviso se ve igual (banner, sonido y lista).
    func userNotificationCenter(_ center: UNUserNotificationCenter, willPresent notification: UNNotification) async
        -> UNNotificationPresentationOptions {
        await MainActor.run {
            Task { await self.estado?.refrescarResumen() }
            self.estado?.version += 1
        }
        return [.banner, .sound, .list]
    }

    func userNotificationCenter(_ center: UNUserNotificationCenter, didReceive response: UNNotificationResponse) async {
        let info = response.notification.request.content.userInfo
        guard let texto = info["ruta"] as? String, let ruta = Ruta(texto: texto) else { return }
        // La tienda del aviso: con varias, primero se pasa a esa y después se abre la ruta.
        let tocado = Tocado(ruta: ruta, tienda: info["tienda"] as? String, marcar: response.actionIdentifier == "preparado")
        await MainActor.run {
            if let estado = self.estado { self.atender(tocado, en: estado) } else { self.pendiente = tocado }
        }
    }

    private func atender(_ t: Tocado, en estado: Estado) {
        Task { @MainActor in
            guard await estado.pasarATiendaDelAviso(t.tienda) else { return }
            if t.marcar, case let .pedido(n) = t.ruta {
                struct Cuerpo: Encodable { var estado: String }
                do {
                    try await estado.api.hacer("POST", "/v1/panel/pedidos/\(n)/estado", Cuerpo(estado: "Listo para despachar"))
                    estado.version += 1
                    estado.avisar("#\(n) listo")
                } catch {
                    estado.rutaPendiente = t.ruta
                }
            } else {
                estado.rutaPendiente = t.ruta
            }
        }
    }
}
