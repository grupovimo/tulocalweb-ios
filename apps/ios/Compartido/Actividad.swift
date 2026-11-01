import ActivityKit
import AppIntents
import Foundation

/// 9a/9b · La Live Activity de los pedidos por preparar: vive mientras haya alguno.
struct PedidosPorPreparar: ActivityAttributes {
    struct ContentState: Codable, Hashable {
        var porPreparar: Int
        var preparadosHoy: Int
        var ultimoNumero: Int?
        var ultimoCliente: String?
    }
    var tienda: String
}

/// «Preparado» desde la Live Activity: pasa el último pedido a «Listo para despachar».
struct MarcarPreparado: LiveActivityIntent {
    static var title: LocalizedStringResource = "Marcar pedido como preparado"
    static var openAppWhenRun = false

    @Parameter(title: "Pedido") var numero: Int

    init() {}
    init(numero: Int) { self.numero = numero }

    func perform() async throws -> some IntentResult {
        try await SesionCompartida.post("/v1/panel/pedidos/\(numero)/estado", json: ["estado": "Listo para despachar"])
        let datos = await DatosHoy.traer()
        for a in Activity<PedidosPorPreparar>.activities {
            let previo = a.content.state
            let nuevo = PedidosPorPreparar.ContentState(
                porPreparar: datos.porPreparar, preparadosHoy: previo.preparadosHoy + 1,
                ultimoNumero: datos.ultimoPorPreparar?.numero, ultimoCliente: datos.ultimoPorPreparar?.cliente)
            if datos.porPreparar == 0 {
                await a.end(ActivityContent(state: nuevo, staleDate: nil), dismissalPolicy: .default)
            } else {
                await a.update(ActivityContent(state: nuevo, staleDate: nil))
            }
        }
        return .result()
    }
}
