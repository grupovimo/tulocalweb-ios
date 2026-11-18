import SwiftUI

/// 3c · «¿Cancelás el pedido #1048?»: qué pasa con lo cobrado (también la seña) y el botón destructivo
/// con Face ID antes de ejecutar. Si se cobró con Mercado Pago y la persona maneja las finanzas, devuelve
/// el pago desde acá (POST …/reembolso) y después cancela. (La API todavía no guarda el motivo ni avisa por mail:
/// esas dos piezas de la hoja quedan para cuando las tenga, ver docs/decisiones.md.)
struct CancelarPedido: View {
    var pedido: PedidoDetalle
    var alCancelar: () async -> Void
    @Environment(\.dismiss) private var cerrar
    @Environment(Estado.self) private var estado
    @State private var cancelando = false
    /// Ya se devolvió (si después falló cancelar, el reintento no vuelve a devolver).
    @State private var devuelto = false
    /// Una sola clave por hoja: si se reintenta, Mercado Pago no devuelve dos veces.
    @State private var clave = UUID().uuidString

    /// «Seña» es cobrado también: se pagó una parte (la API la cuenta igual, sinpago.go).
    private var sena: Bool { pedido.estadoPago == "Seña" }
    private var cobrado: Bool {
        ["Acreditada", "Acreditado", "Aprobado", "Cobrado", "Seña"].contains(pedido.estadoPago)
    }
    private var monto: Int { sena ? (pedido.sena?.monto ?? pedido.total) : pedido.total }
    private var puedeDevolver: Bool { cobrado && !devuelto && pedido.pago == "mp" && estado.puede("finanzas") }

    private var explicacion: String {
        let nombre = pedido.comprador.nombre.split(separator: " ").first.map(String.init) ?? pedido.comprador.nombre
        let que = sena ? "una seña de \(Formato.pesos(monto))" : Formato.pesos(monto)
        if pedido.estadoPago == "Devuelto" {
            return "Ya le devolviste el pago a \(nombre). El stock vuelve a la tienda. Esto no se puede deshacer."
        }
        if cobrado && pedido.pago == "mp" {
            if devuelto { return "Ya le devolviste \(que) a \(nombre). Falta cancelar el pedido. Esto no se puede deshacer." }
            if puedeDevolver { return "\(nombre) pagó \(que) con Mercado Pago. Se lo devolvés desde acá al cancelar. Esto no se puede deshacer." }
            return "\(nombre) pagó \(que) con Mercado Pago: la devolución la hace quien maneja las finanzas de la tienda. Esto no se puede deshacer."
        }
        if cobrado {
            return "\(nombre) pagó \(que) por \(Marcas.pago(pedido.pago).lowercased()): la devolución la hacés vos. Esto no se puede deshacer."
        }
        return "Todavía no estaba cobrado. El stock vuelve a la tienda. Esto no se puede deshacer."
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 14) {
            Grabber().frame(maxWidth: .infinity)
            VStack(alignment: .leading, spacing: 6) {
                Text("¿Cancelás el pedido #\(pedido.numero)?").estilo(.tituloHoja).foregroundStyle(C.texto)
                Text(Formato.sinViudas(explicacion)).estilo(.subhead).foregroundStyle(C.texto2).lineSpacing(2)
                    .fixedSize(horizontal: false, vertical: true)
            }
            .padding(.horizontal, 4).padding(.top, 8)
            VStack(spacing: 0) {
                ForEach(Array(pedido.items.prefix(4).enumerated()), id: \.offset) { i, it in
                    if i > 0 { Separador(inset: 66) }
                    HStack(spacing: 12) {
                        Miniatura(foto: it.foto, tam: 40, radio: 11)
                        Text("\(it.cantidad) × \(it.nombre)").estilo(.body16).foregroundStyle(C.texto).lineLimit(1)
                        Spacer(minLength: 0)
                    }
                    .padding(.horizontal, 14).padding(.vertical, 10)
                }
            }
            .tarjeta()
            Spacer(minLength: 0)
            VStack(spacing: 8) {
                Button {
                    Task { await confirmar(devolver: puedeDevolver) }
                } label: {
                    HStack(spacing: 8) {
                        if cancelando { ProgressView() } else { Image(systemName: "faceid") }
                        Text(puedeDevolver ? "Devolver \(Formato.pesos(monto)) y cancelar" : "Cancelar pedido")
                    }
                }
                .buttonStyle(BotonDestructivo())
                .disabled(cancelando)
                if puedeDevolver {
                    Button("Cancelar sin devolver") { Task { await confirmar(devolver: false) } }
                        .buttonStyle(BotonTexto(alto: 48, color: C.errorTinta)).disabled(cancelando)
                }
                Button("Volver") { cerrar() }.buttonStyle(BotonTexto(alto: 48))
            }
            .padding(.bottom, 12)
        }
        .padding(.horizontal, 20)
    }

    private func confirmar(devolver: Bool) async {
        guard await estado.confirmarIdentidad("Para cancelar el pedido #\(pedido.numero)") else { return }
        cancelando = true
        defer { cancelando = false }
        do {
            // Primero la devolución: si falla, el pedido sigue como estaba.
            if devolver {
                try await estado.api.hacer("POST", "/v1/panel/pedidos/\(pedido.numero)/reembolso", encabezados: ["Idempotency-Key": clave])
                devuelto = true
            }
            try await PedidosApi.cambiarEstado(estado.api, numero: pedido.numero, a: "Cancelado")
            Haptico.exito()
            cerrar()
            estado.avisar(devolver ? "Devolviste \(Formato.pesos(monto)) y cancelaste el pedido #\(pedido.numero)" : "Cancelaste el pedido #\(pedido.numero)")
            estado.version += 1
            await alCancelar()
            await estado.refrescarResumen()
        } catch let e as ErrorApi {
            Haptico.error()
            estado.avisar(e.detalle, ok: false)
        } catch {}
    }
}
