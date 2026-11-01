import ActivityKit
import Foundation
import WidgetKit

/// La app arranca, actualiza y cierra la Live Activity de los pedidos por preparar cada vez que
/// refresca el resumen (al abrir, al volver y con cada push de venta), y recarga los widgets.
@MainActor
enum ActividadEnApp {
    static func sincronizar(tienda: String) async {
        WidgetCenter.shared.reloadAllTimelines()
        guard ActivityAuthorizationInfo().areActivitiesEnabled else { return }
        let datos = await DatosHoy.traer()
        guard datos.conSesion else { return }
        let actuales = Activity<PedidosPorPreparar>.activities
        let previo = actuales.first?.content.state
        let estado = PedidosPorPreparar.ContentState(
            porPreparar: datos.porPreparar,
            preparadosHoy: max(previo?.preparadosHoy ?? 0, previo.map { max(0, $0.porPreparar - datos.porPreparar) + $0.preparadosHoy } ?? 0),
            ultimoNumero: datos.ultimoPorPreparar?.numero, ultimoCliente: datos.ultimoPorPreparar?.cliente)
        if datos.porPreparar == 0 {
            for a in actuales { await a.end(ActivityContent(state: estado, staleDate: nil), dismissalPolicy: .default) }
        } else if actuales.isEmpty {
            if let a = try? Activity.request(attributes: PedidosPorPreparar(tienda: tienda),
                                             content: ActivityContent(state: estado, staleDate: nil), pushType: .token) {
                escuchar(a)
            }
        } else {
            for a in actuales where a.content.state != estado {
                await a.update(ActivityContent(state: estado, staleDate: nil))
            }
        }
    }

    /// El servidor actualiza la actividad por APNs con la app cerrada (§54): se le pasa su token.
    static var api: Api?
    static var tokenPush: String?
    private static var escuchando: Set<String> = []

    static func escuchar(_ a: Activity<PedidosPorPreparar>) {
        guard escuchando.insert(a.id).inserted else { return }
        Task {
            for await t in a.pushTokenUpdates {
                await avisar(actividad: t.map { String(format: "%02x", $0) }.joined(), inicio: nil)
            }
        }
        Task {
            for await e in a.activityStateUpdates where e == .ended || e == .dismissed {
                await avisar(actividad: "", inicio: nil)
            }
        }
    }

    /// El token para arrancar la actividad desde el servidor (iOS 17.2+) y las actividades que ya
    /// estaban (de antes de abrir la app).
    static func arrancarEscucha() {
        for a in Activity<PedidosPorPreparar>.activities { escuchar(a) }
        if #available(iOS 17.2, *) {
            Task {
                for await t in Activity<PedidosPorPreparar>.pushToStartTokenUpdates {
                    await avisar(actividad: nil, inicio: t.map { String(format: "%02x", $0) }.joined())
                }
            }
        }
    }

    private static func avisar(actividad: String?, inicio: String?) async {
        guard let api, let tokenPush else { return }
        struct Cuerpo: Encodable { var token: String; var actividad: String?; var inicio: String? }
        var actual = actividad
        if actual == nil { actual = Activity<PedidosPorPreparar>.activities.first?.pushToken.map { $0.map { String(format: "%02x", $0) }.joined() } }
        try? await api.hacer("PUT", "/v1/panel/dispositivos/actividad", Cuerpo(token: tokenPush, actividad: actual ?? "", inicio: inicio))
    }

    static func terminarTodo() {
        WidgetCenter.shared.reloadAllTimelines()
        Task {
            for a in Activity<PedidosPorPreparar>.activities { await a.end(nil, dismissalPolicy: .immediate) }
        }
    }
}
