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
            _ = try? Activity.request(attributes: PedidosPorPreparar(tienda: tienda),
                                      content: ActivityContent(state: estado, staleDate: nil))
        } else {
            for a in actuales where a.content.state != estado {
                await a.update(ActivityContent(state: estado, staleDate: nil))
            }
        }
    }

    static func terminarTodo() {
        WidgetCenter.shared.reloadAllTimelines()
        Task {
            for a in Activity<PedidosPorPreparar>.activities { await a.end(nil, dismissalPolicy: .immediate) }
        }
    }
}
