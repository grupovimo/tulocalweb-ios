import SwiftUI

/// Envíos y retiro (Configuración › Envíos del panel, /panel/envios/configurar): retiro en el local,
/// envío propio en moto, Correo Argentino (se conecta en Integraciones) y el envío gratis desde el
/// monto de la tienda. Las reglas son las de Go (configuracion/entregas.go): al menos una forma de
/// entrega, la moto con su zona y el envío gratis desde $ 1.000.
struct EnviosTienda: View {
    @Environment(Estado.self) private var estado
    @Environment(\.dismiss) private var cerrar

    struct Entregas: Codable {
        var retiro: Bool
        var retiroDireccion: String
        var retiroHorario: String
        var moto: Bool
        var motoCosto: Int
        var motoZona: String
        var motoDemora: String
        var correo: Bool?
        var envioGratis: Bool?
        var envioGratisMonto: Int?
    }

    @State private var e: Entregas?
    @State private var costo = ""
    @State private var monto = ""
    @State private var guardando = false

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 12) {
                Text("Envíos y retiro").estilo(.tituloGrande).foregroundStyle(C.texto)
                Text(Formato.sinViudas("Cómo les llega lo que compran en tu tienda.")).estilo(.subhead).foregroundStyle(C.texto2)
                if let e {
                    formulario(e)
                } else {
                    EsqueletoFilas(filas: 5, miniatura: 0)
                }
            }
            .margenPantalla()
            .padding(.bottom, 32)
        }
        .scrollDismissesKeyboard(.interactively)
        .barraAccion {
            if e != nil {
                Button(guardando ? "Guardando…" : "Guardar") { Task { await guardar() } }
                    .buttonStyle(BotonPrincipal()).disabled(guardando)
            }
        }
        .background(C.fondo.ignoresSafeArea())
        .navigationBarTitleDisplayMode(.inline)
        .task { await cargar() }
    }

    @ViewBuilder
    private func formulario(_ v: Entregas) -> some View {
        opcion("Retiro en el local", "Sin costo · te avisamos cuando está listo", simbolo: "storefront",
               prendido: Binding(get: { e?.retiro ?? false }, set: { e?.retiro = $0 })) {
            if v.retiro {
                campo("Dirección", Binding(get: { e?.retiroDireccion ?? "" }, set: { e?.retiroDireccion = $0 }), "Calle, número y localidad")
                campo("Horario", Binding(get: { e?.retiroHorario ?? "" }, set: { e?.retiroHorario = $0 }), "Lunes a viernes de 10 a 19")
            }
        }
        opcion("Envío propio en moto", "Con tu moto o tu cadete, a un costo fijo", simbolo: "paperplane",
               prendido: Binding(get: { e?.moto ?? false }, set: { e?.moto = $0 })) {
            if v.moto {
                campo("Costo en pesos", $costo, "3500", teclado: .numberPad)
                campo("Cuánto tarda", Binding(get: { e?.motoDemora ?? "" }, set: { e?.motoDemora = $0 }), "En el día")
                campo("Hasta dónde llega", Binding(get: { e?.motoZona ?? "" }, set: { e?.motoZona = $0 }), "Palermo, Belgrano, Núñez")
            }
        }
        HStack(spacing: 12) {
            IconoOpcion(simbolo: "envelope")
            VStack(alignment: .leading, spacing: 2) {
                Text("Correo Argentino").estilo(.fila).foregroundStyle(C.texto)
                Text(v.correo == true ? "Cotiza con tu cuenta de MiCorreo" : "Conectalo en Integraciones para enviar a todo el país")
                    .estilo(.footnote).foregroundStyle(C.texto2)
            }
            Spacer(minLength: 6)
            ChipSimple(texto: v.correo == true ? "Conectado" : "Sin conectar",
                       fondo: v.correo == true ? C.okFondo : C.apagadoFondo, tinta: v.correo == true ? C.okTinta : C.neutroTinta)
        }
        .padding(16).tarjeta()
        opcion("Envío gratis", (v.envioGratis ?? false) ? "En moto y por Correo, desde el monto que elijas" : "Para que sumen algo más al carrito",
               simbolo: "gift", prendido: Binding(get: { e?.envioGratis ?? false }, set: { e?.envioGratis = $0 })) {
            if v.envioGratis == true {
                campo("Desde qué monto", $monto, "40000", teclado: .numberPad)
                Text(Formato.sinViudas("La tienda muestra cuánto les falta en la bolsa y en el carrito, y el envío sale gratis al pagar."))
                    .estilo(.footnote).foregroundStyle(C.texto2)
            }
        }
    }

    private func opcion<Contenido: View>(_ titulo: String, _ detalle: String, simbolo: String, prendido: Binding<Bool>,
                                         @ViewBuilder contenido: () -> Contenido) -> some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack(spacing: 12) {
                IconoOpcion(simbolo: simbolo)
                VStack(alignment: .leading, spacing: 2) {
                    Text(titulo).estilo(.fila).foregroundStyle(C.texto)
                    Text(Formato.sinViudas(detalle)).estilo(.footnote).foregroundStyle(C.texto2)
                }
                Spacer(minLength: 6)
                Toggle(titulo, isOn: prendido).labelsHidden().tint(C.ok)
            }
            contenido()
        }
        .padding(16).tarjeta()
    }

    private func campo(_ titulo: String, _ texto: Binding<String>, _ ejemplo: String, teclado: UIKeyboardType = .default) -> some View {
        VStack(alignment: .leading, spacing: 6) {
            Text(titulo).estilo(.footnoteFuerte).foregroundStyle(C.texto2)
            TextField(ejemplo, text: texto)
                .keyboardType(teclado)
                .estilo(.body16).foregroundStyle(C.texto)
                .padding(.horizontal, 14).padding(.vertical, 12)
                .background(C.superficie, in: RoundedRectangle(cornerRadius: 14, style: .continuous))
                .overlay(RoundedRectangle(cornerRadius: 14, style: .continuous).strokeBorder(C.linea2, lineWidth: 1))
        }
    }

    private func cargar() async {
        do {
            let v: Entregas = try await estado.api.get("/v1/panel/configuracion/entregas")
            costo = String(v.motoCosto)
            monto = String(v.envioGratisMonto ?? 40000)
            e = v
        } catch {
            estado.avisar("No pudimos traer los envíos", ok: false)
        }
    }

    private func guardar() async {
        guard var v = e else { return }
        v.motoCosto = Int(costo.filter(\.isNumber)) ?? 0
        v.envioGratisMonto = Int(monto.filter(\.isNumber)) ?? 0
        v.correo = nil
        if v.envioGratis == true, (v.envioGratisMonto ?? 0) < 1000 {
            estado.avisar("El envío gratis va desde $ 1.000 hasta $ 100.000.000.", ok: false)
            return
        }
        guardando = true
        defer { guardando = false }
        do {
            try await estado.api.hacer("PUT", "/v1/panel/configuracion/entregas", v)
            Haptico.exito()
            estado.avisar("Envíos y retiro guardados")
            cerrar()
        } catch let x as ErrorApi {
            estado.avisar(x.detalle, ok: false)
        } catch {
            estado.avisar("No pudimos guardarlos", ok: false)
        }
    }
}

/// El ícono redondo de cada opción (como las de Configuración del panel).
struct IconoOpcion: View {
    var simbolo: String
    var body: some View {
        Image(systemName: simbolo)
            .font(.system(size: 16, weight: .semibold))
            .foregroundStyle(C.primarioTexto)
            .frame(width: 40, height: 40)
            .background(C.crema, in: RoundedRectangle(cornerRadius: 12, style: .continuous))
            .accessibilityHidden(true)
    }
}
