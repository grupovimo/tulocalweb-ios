import SwiftUI

/// 3c · «¿Cancelás el pedido #1048?»: aviso de reintegro manual si ya se cobró, y el botón
/// destructivo con Face ID antes de ejecutar. (La API todavía no guarda el motivo ni avisa por mail:
/// esas dos piezas de la hoja quedan para cuando las tenga, ver docs/decisiones.md.)
struct CancelarPedido: View {
    var pedido: PedidoDetalle
    var alCancelar: () async -> Void
    @Environment(\.dismiss) private var cerrar
    @Environment(Estado.self) private var estado
    @State private var cancelando = false

    private var cobrado: Bool {
        ["Acreditada", "Acreditado", "Aprobado", "Cobrado"].contains(pedido.estadoPago)
    }

    private var explicacion: String {
        let nombre = pedido.comprador.nombre.split(separator: " ").first.map(String.init) ?? pedido.comprador.nombre
        if cobrado && pedido.pago == "mp" {
            return "\(nombre) pagó con Mercado Pago: la devolución de \(Formato.pesos(pedido.total)) la hacés desde Mercado Pago. Esto no se puede deshacer."
        }
        if cobrado {
            return "\(nombre) pagó por \(Marcas.pago(pedido.pago).lowercased()): la devolución de \(Formato.pesos(pedido.total)) la hacés vos. Esto no se puede deshacer."
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
                    Task { await confirmar() }
                } label: {
                    HStack(spacing: 8) {
                        if cancelando { ProgressView() } else { Image(systemName: "faceid") }
                        Text("Cancelar pedido")
                    }
                }
                .buttonStyle(BotonDestructivo())
                .disabled(cancelando)
                Button("Volver") { cerrar() }.buttonStyle(BotonTexto(alto: 48))
            }
            .padding(.bottom, 12)
        }
        .padding(.horizontal, 20)
    }

    private func confirmar() async {
        guard await estado.confirmarIdentidad("Para cancelar el pedido #\(pedido.numero)") else { return }
        cancelando = true
        defer { cancelando = false }
        do {
            try await PedidosApi.cambiarEstado(estado.api, numero: pedido.numero, a: "Cancelado")
            Haptico.exito()
            cerrar()
            estado.avisar("Cancelaste el pedido #\(pedido.numero)")
            estado.version += 1
            await alCancelar()
            await estado.refrescarResumen()
        } catch let e as ErrorApi {
            Haptico.error()
            estado.avisar(e.detalle, ok: false)
        } catch {}
    }
}
