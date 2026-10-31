import SwiftUI

/// Nombres y colores de canales, pagos y envíos (como apps/web/src/lib/marcas.ts).
enum Marcas {
    /// Corto, para la meta de una fila: «Web», «ML», «Local».
    static func canalCorto(_ c: String) -> String {
        switch c {
        case "web": return "Web"
        case "ml": return "ML"
        case "pos": return "Local"
        case "wa": return "WhatsApp"
        case "ig": return "Instagram"
        case "tg": return "Telegram"
        default: return c.capitalized
        }
    }

    /// Largo: «Tienda web», «Mercado Libre».
    static func canal(_ c: String) -> String {
        switch c {
        case "web": return "Tienda web"
        case "ml": return "Mercado Libre"
        case "pos": return "Local"
        default: return canalCorto(c)
        }
    }

    static func colorCanal(_ c: String) -> Color {
        switch c {
        case "web": return C.canalWeb
        case "ml": return C.canalML
        case "pos": return C.canalLocal
        case "wa": return Color(hex: 0x2F8F63)
        case "ig": return Color(hex: 0xC24E25)
        default: return C.texto3
        }
    }

    static func pago(_ p: String) -> String {
        switch p {
        case "mp": return "Mercado Pago"
        case "tarjeta": return "Tarjeta"
        case "payway": return "Payway"
        case "modo": return "MODO"
        case "uala": return "Ualá"
        case "transf": return "Transferencia"
        case "efectivo": return "Efectivo"
        default: return p.capitalized
        }
    }

    static func envio(_ e: String) -> String {
        switch e {
        case "andreani": return "Andreani"
        case "oca": return "OCA"
        case "correo": return "Correo Argentino"
        case "moto": return "Envío propio"
        case "retiro": return "Retiro en el local"
        default: return e.capitalized
        }
    }

    /// El siguiente estado de un pedido y el texto del botón (la barra de acción de 3b y el swipe de
    /// 3a). nil si ya terminó.
    static func siguiente(_ estado: String, envio: String) -> (estado: String, boton: String, corto: String)? {
        let retiro = envio == "retiro"
        switch estado {
        case "Sin pagar": return ("Para preparar", "Marcar como cobrado", "Cobrado")
        case "Para preparar":
            return ("Listo para despachar", retiro ? "Listo para retirar" : "Listo para despachar", "Listo")
        case "Listo para despachar":
            return retiro ? ("Entregado", "Marcar como retirado", "Retirado") : ("En camino", "Marcar como despachado", "Despachado")
        case "En camino": return ("Entregado", "Marcar como entregado", "Entregado")
        default: return nil
        }
    }

    /// El estado de antes (para «Deshacer»).
    static func anterior(_ estado: String, envio: String) -> String? {
        let circuito = envio == "retiro"
            ? ["Sin pagar", "Para preparar", "Listo para despachar", "Entregado"]
            : ["Sin pagar", "Para preparar", "Listo para despachar", "En camino", "Entregado"]
        guard let i = circuito.firstIndex(of: estado), i > 0 else { return nil }
        return circuito[i - 1]
    }

    /// Cómo queda dicho el cambio en el toast: «#1043 listo para retirar».
    static func dicho(_ estado: String, envio: String) -> String {
        ChipEstado.texto(estado, envio: envio).lowercased()
    }
}
