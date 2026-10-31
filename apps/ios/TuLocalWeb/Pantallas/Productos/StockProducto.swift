import SwiftUI

/// 11m · Stock: segmentado Ajuste / Ingreso / Conteo, stepper grande de 52 con el valor relativo
/// («+24») y «Quedan 2 → van a quedar 26». Con variantes, primero se elige cuál.
/// POST /v1/panel/listas/inventario/acciones (ajustar, ingreso, conteo).
struct StockProducto: View {
    var producto: ProductoFila
    var alGuardar: () async -> Void
    @Environment(\.dismiss) private var cerrar
    @Environment(Estado.self) private var estado
    @State private var modo = "ingreso"
    @State private var cantidad = 0
    @State private var variante: String?
    @State private var guardando = false

    private var variantes: [ProductoFila.Variante] { producto.variantes.filter(\.activa) }
    private var actual: Int {
        if let v = variantes.first(where: { $0.id == variante }) { return v.stock }
        return producto.stock
    }
    private var queda: Int {
        switch modo {
        case "conteo": return max(0, cantidad)
        default: return max(0, actual + cantidad)
        }
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 14) {
            Grabber().frame(maxWidth: .infinity)
            VStack(alignment: .leading, spacing: 3) {
                Text("Stock de \(producto.nombre)").estilo(.tituloHojaChico).foregroundStyle(C.texto).lineLimit(2)
                if actual == 0 {
                    Text("Sin stock: sigue visible en la tienda").estilo(.label).foregroundStyle(C.errorTinta)
                } else if actual < ListaProductos.stockBajo {
                    Text("Stock bajo: avisamos con menos de \(ListaProductos.stockBajo)").estilo(.label).foregroundStyle(C.avisoTinta)
                }
            }
            .padding(.horizontal, 4).padding(.top, 6)
            if !variantes.isEmpty {
                ScrollView(.horizontal) {
                    HStack(spacing: 8) {
                        ForEach(variantes) { v in
                            ChipFiltro(texto: v.texto, cuenta: v.stock, cuentaAlerta: v.stock == 0, activo: variante == v.id) {
                                variante = v.id
                                cantidad = modo == "conteo" ? v.stock : 0
                            }
                        }
                    }
                    .padding(.horizontal, 20)
                }
                .scrollIndicators(.hidden)
                .padding(.horizontal, -20)
            }
            Segmentado(opciones: [("ajuste", "Ajuste"), ("ingreso", "Ingreso"), ("conteo", "Conteo")], elegido: $modo)
            VStack(spacing: 12) {
                Text(pregunta).estilo(.subhead).foregroundStyle(C.texto2)
                HStack(spacing: 18) {
                    paso("minus", -1)
                    Text(valor)
                        .estilo(Estilo(tamano: 52, peso: .bold, tracking: -0.03, relativo: .largeTitle))
                        .monospacedDigit().foregroundStyle(C.texto)
                        .frame(minWidth: 70)
                        .contentTransition(.numericText())
                        .accessibilityLabel(valor)
                    paso("plus", 1)
                }
                (Text("Quedan ") + Text("\(actual)").fontWeight(.bold) + Text(" → van a quedar ") + Text("\(queda)").fontWeight(.bold))
                    .estilo(.subhead).monospacedDigit().foregroundStyle(C.texto)
            }
            .padding(18)
            .frame(maxWidth: .infinity)
            .tarjeta()
            if let costo = producto.costo, costo > 0 {
                HStack {
                    Text("Costo por unidad").estilo(.body16).foregroundStyle(C.texto)
                    Spacer()
                    Text(Formato.pesos(costo)).estilo(.body16).monospacedDigit().foregroundStyle(C.texto2)
                }
                .padding(.horizontal, 16).padding(.vertical, 13)
                .tarjeta()
            }
            Spacer(minLength: 0)
            Button {
                Task { await guardar() }
            } label: {
                if guardando { ProgressView().tint(.white) } else { Text(boton) }
            }
            .buttonStyle(BotonPrincipal(alto: 56))
            .disabled(guardando || (modo != "conteo" && cantidad == 0) || (!variantes.isEmpty && variante == nil))
            .padding(.bottom, 12)
        }
        .padding(.horizontal, 20)
        .onAppear { variante = variantes.first?.id }
        .onChange(of: modo) { _, m in cantidad = m == "conteo" ? actual : 0 }
        .sensoryFeedback(.selection, trigger: cantidad)
    }

    private var pregunta: String {
        switch modo {
        case "ajuste": return "¿Cuánto sumás o restás?"
        case "conteo": return "¿Cuántos contaste?"
        default: return "¿Cuántos entraron?"
        }
    }

    private var valor: String {
        if modo == "conteo" { return "\(cantidad)" }
        return cantidad > 0 ? "+\(cantidad)" : (cantidad < 0 ? "−\(-cantidad)" : "0")
    }

    private var boton: String {
        switch modo {
        case "ajuste": return "Guardar ajuste"
        case "conteo": return "Guardar conteo"
        default: return "Cargar ingreso"
        }
    }

    private func paso(_ simbolo: String, _ d: Int) -> some View {
        let minimo = modo == "ingreso" ? 0 : (modo == "conteo" ? 0 : -actual)
        let deshabilitado = d < 0 && cantidad <= minimo
        return Button {
            cantidad += d
        } label: {
            Image(systemName: simbolo).font(.system(size: 22, weight: .medium)).foregroundStyle(C.primarioTexto)
                .frame(width: 52, height: 52).background(C.pista, in: Circle())
        }
        .buttonStyle(.plain)
        .disabled(deshabilitado)
        .opacity(deshabilitado ? 0.45 : 1)
        .buttonRepeatBehavior(.enabled)
        .accessibilityLabel(d > 0 ? "Sumar uno" : "Restar uno")
    }

    private func guardar() async {
        struct Ajuste: Encodable { var ids: [String]; var accion: String; var valor: Int }
        struct Conteo: Encodable {
            struct Item: Encodable { var id: String; var cantidad: Int }
            var ids: [String]; var accion = "conteo"; var valor: [Item]
        }
        struct R: Decodable { var hechos: Int }
        let id = variante.map { "\(producto.id):\($0)" } ?? producto.id
        guardando = true
        defer { guardando = false }
        do {
            let r: R
            if modo == "conteo" {
                r = try await estado.api.enviar("POST", "/v1/panel/listas/inventario/acciones",
                                                Conteo(ids: [id], valor: [.init(id: id, cantidad: cantidad)]))
            } else {
                r = try await estado.api.enviar("POST", "/v1/panel/listas/inventario/acciones",
                                                Ajuste(ids: [id], accion: modo == "ingreso" ? "ingreso" : "ajustar", valor: cantidad))
            }
            Haptico.exito()
            cerrar()
            estado.avisar(r.hechos > 0 ? "Stock de \(producto.nombre): \(queda)" : "No cambió el stock")
            estado.version += 1
            await alGuardar()
        } catch let e as ErrorApi {
            Haptico.error()
            estado.avisar(e.detalle, ok: false)
        } catch {}
    }
}
