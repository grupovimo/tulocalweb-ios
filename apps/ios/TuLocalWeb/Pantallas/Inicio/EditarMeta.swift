import SwiftUI

/// 12e · Editar la meta del mes: hoja de altura fija con el monto, atajos («Igual que el mes
/// pasado», «+10 %», «+20 %») y el promedio diario que hace falta. PUT /v1/panel/meta.
struct EditarMeta: View {
    var actual: Int
    var vendido: Int
    var alGuardar: () async -> Void
    @Environment(\.dismiss) private var cerrar
    @Environment(Estado.self) private var estado
    @State private var texto = ""
    @State private var atajo: Int?
    @State private var guardando = false
    @FocusState private var foco: Bool

    private var monto: Int { Int(texto.filter(\.isNumber)) ?? 0 }

    var body: some View {
        let hoy = Date()
        let c = Formato.calendario
        let dias = c.range(of: .day, in: .month, for: hoy)?.count ?? 30
        let quedan = max(1, dias - c.component(.day, from: hoy) + 1)
        let pasaron = max(1, c.component(.day, from: hoy) - 1)
        VStack(spacing: 14) {
            Grabber()
            HStack {
                Button("Cancelar") { cerrar() }.font(Estilo.body.font).foregroundStyle(C.primarioTexto)
                Spacer()
                Text("Meta de \(Formato.mesDe(hoy))").estilo(.headline).foregroundStyle(C.texto)
                Spacer()
                Button("Listo") { Task { await guardar() } }
                    .font(Estilo(tamano: 17, peso: .bold, relativo: .headline).font).foregroundStyle(C.primarioTexto)
                    .disabled(guardando)
            }
            .padding(.top, 6)
            VStack(spacing: 6) {
                Text("¿Cuánto querés vender?").estilo(.subhead).foregroundStyle(C.texto2)
                HStack(spacing: 0) {
                    Text("$\(Formato.fino)").estilo(Estilo(tamano: 44, peso: .bold, tracking: -0.03, relativo: .largeTitle))
                    TextField("0", text: Binding(get: { monto > 0 ? Formato.numero(monto) : "" },
                                                 set: { texto = $0; atajo = nil }))
                        .estilo(Estilo(tamano: 44, peso: .bold, tracking: -0.03, relativo: .largeTitle))
                        .keyboardType(.numberPad)
                        .focused($foco)
                        .fixedSize()
                }
                .monospacedDigit().foregroundStyle(C.texto)
                .padding(.horizontal, 6).padding(.bottom, 4)
                .overlay(alignment: .bottom) { Rectangle().fill(C.acento).frame(height: 2) }
                HStack(spacing: 8) {
                    if vendido > 0 {
                        chip("A este ritmo", 0) { texto = String(vendido * dias / pasaron) }
                    }
                    chip("+10\(Formato.fino)%", 10) { texto = String(base * 110 / 100) }
                    chip("+20\(Formato.fino)%", 20) { texto = String(base * 120 / 100) }
                }
                .padding(.top, 6)
                if monto > 0 {
                    let falta = max(0, monto - vendido)
                    Text(Formato.sinViudas("Son unos \(Formato.pesos(falta / quedan)) por día. Este mes promediás \(Formato.pesos(vendido / pasaron))."))
                        .estilo(.texto14).foregroundStyle(C.texto2).multilineTextAlignment(.center)
                        .padding(.horizontal, 10).padding(.top, 6)
                }
            }
            .padding(18)
            .frame(maxWidth: .infinity)
            .tarjeta()
            Spacer(minLength: 0)
        }
        .padding(.horizontal, 20)
        .onAppear {
            texto = actual > 0 ? String(actual) : ""
            foco = true
        }
    }

    private var base: Int { actual > 0 ? actual : max(vendido, 1) }

    private func chip(_ t: String, _ id: Int, _ accion: @escaping () -> Void) -> some View {
        ChipFiltro(texto: t, activo: atajo == id) {
            accion()
            atajo = id
            Haptico.seleccion()
        }
    }

    private func guardar() async {
        struct Cuerpo: Encodable { var objetivo: Int }
        guardando = true
        defer { guardando = false }
        do {
            try await estado.api.hacer("PUT", "/v1/panel/meta", Cuerpo(objetivo: monto))
            Haptico.exito()
            cerrar()
            await alGuardar()
            estado.avisar(monto > 0 ? "Meta guardada" : "Sacaste la meta del mes")
        } catch let e as ErrorApi {
            estado.avisar(e.detalle, ok: false)
        } catch {}
    }
}
