import SwiftUI
import WebKit

/// La tienda de verdad con el borrador (la misma página que el editor web, /panel/mi-tienda/vista, con el
/// token de la app). Habla con la página con el protocolo de docs/especificaciones/tienda-bloques.md:
/// - la página avisa `{tipo:'lista'}` por `webkit.messageHandlers.tienda` y recién ahí se le manda algo;
/// - `window.tiendaVista.mostrar(<diseño>, '<pagina>')` dibuja el borrador sin guardar ni recargar;
/// - `window.tiendaVista.seleccionar('<id>')` resalta el bloque elegido en la lista;
/// - al tocar un bloque (con `marcos=1`) avisa `{tipo:'bloque', id}` y el editor abre su panel.
/// Si la página todavía no responde (no avisa «lista»), queda como antes: se recarga con `vuelta`.
struct VistaPreviaTienda: UIViewRepresentable {
    var token: String?
    var vuelta: Int
    var pagina = "inicio"
    /// El borrador sin guardar, en JSON (Diseno.textoJSON); nil: lo guardado.
    var diseno: String?
    var seleccion: String?
    var marcos = false
    var alTocarBloque: ((String) -> Void)?
    /// true cuando la página responde al protocolo (los cambios se ven sin guardar).
    var alCambiarViva: ((Bool) -> Void)?

    func makeCoordinator() -> Coordinador { Coordinador() }

    func makeUIView(context: Context) -> WKWebView {
        let conf = WKWebViewConfiguration()
        conf.userContentController.add(PuenteDebil(context.coordinator), name: "tienda")
        let w = WKWebView(frame: .zero, configuration: conf)
        w.isOpaque = false
        w.backgroundColor = .clear
        context.coordinator.web = w
        return w
    }

    static func dismantleUIView(_ w: WKWebView, coordinator: Coordinador) {
        w.configuration.userContentController.removeScriptMessageHandler(forName: "tienda")
        coordinator.reloj?.invalidate()
    }

    func updateUIView(_ w: WKWebView, context: Context) {
        let c = context.coordinator
        c.alTocarBloque = alTocarBloque
        c.alCambiarViva = alCambiarViva
        if let diseno { c.pendiente = diseno }
        if c.cargada != vuelta || c.pagina != pagina {
            if c.lista, c.cargada == vuelta, diseno != nil {
                // Otra página con la vista ya conectada: se le pide que la dibuje, sin recargar.
                c.pagina = pagina
                c.enviar(siempre: true)
            } else {
                c.cargada = vuelta
                c.pagina = pagina
                c.cargar(token: token, marcos: marcos)
            }
        } else if diseno != nil {
            c.programar()
        }
        if seleccion != c.seleccionada {
            c.seleccionada = seleccion
            if let s = seleccion { c.seleccionar(s) }
        }
    }

    final class Coordinador: NSObject, WKScriptMessageHandler {
        weak var web: WKWebView?
        var cargada = -1
        var pagina = "inicio"
        var lista = false
        var pendiente: String?
        var enviado: String?
        var seleccionada: String?
        var reloj: Timer?
        var alTocarBloque: ((String) -> Void)?
        var alCambiarViva: ((Bool) -> Void)?

        func cargar(token: String?, marcos: Bool) {
            lista = false
            enviado = nil
            let avisar = alCambiarViva
            DispatchQueue.main.async { avisar?(false) }
            var u = URLComponents(url: Web.base.appending(path: "panel/mi-tienda/vista"), resolvingAgainstBaseURL: false)
            u?.queryItems = [URLQueryItem(name: "pagina", value: pagina)] + (marcos ? [URLQueryItem(name: "marcos", value: "1")] : [])
            guard let url = u?.url else { return }
            var r = URLRequest(url: url)
            r.cachePolicy = .reloadIgnoringLocalCacheData
            if let token { r.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization") }
            web?.load(r)
        }

        func userContentController(_ c: WKUserContentController, didReceive m: WKScriptMessage) {
            guard let o = m.body as? [String: Any], let tipo = o["tipo"] as? String else { return }
            switch tipo {
            case "lista":
                lista = true
                alCambiarViva?(true)
                enviar(siempre: true)
                if let s = seleccionada { seleccionar(s) }
            case "bloque":
                if let id = o["id"] as? String { Haptico.seleccion(); alTocarBloque?(id) }
            default: break
            }
        }

        /// Mientras se escribe, manda el borrador un cuarto de segundo después de la última tecla.
        func programar() {
            guard lista, pendiente != enviado else { return }
            reloj?.invalidate()
            reloj = Timer.scheduledTimer(withTimeInterval: 0.25, repeats: false) { [weak self] _ in self?.enviar() }
        }

        func enviar(siempre: Bool = false) {
            guard lista, let d = pendiente, siempre || d != enviado else { return }
            enviado = d
            web?.evaluateJavaScript("window.tiendaVista && window.tiendaVista.mostrar(\(d), \(Self.js(pagina)))")
        }

        func seleccionar(_ id: String) {
            guard lista else { return }
            web?.evaluateJavaScript("window.tiendaVista && window.tiendaVista.seleccionar(\(Self.js(id)))")
        }

        /// Un texto como literal de JavaScript (comillas y escapes de JSON).
        static func js(_ s: String) -> String {
            guard let d = try? JSONEncoder().encode(s), let t = String(data: d, encoding: .utf8) else { return "''" }
            return t
        }
    }

    /// WKUserContentController retiene a su handler: este puente evita el ciclo con el coordinador.
    final class PuenteDebil: NSObject, WKScriptMessageHandler {
        weak var destino: WKScriptMessageHandler?
        init(_ d: WKScriptMessageHandler) { destino = d }
        func userContentController(_ c: WKUserContentController, didReceive m: WKScriptMessage) {
            destino?.userContentController(c, didReceive: m)
        }
    }
}
