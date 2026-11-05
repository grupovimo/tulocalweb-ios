import SwiftUI

// El plan de Tu Local Web (§56), igual que Plan y facturación del panel: elegir un plan no cobra nada; el
// resumen muestra qué se paga hoy (línea por línea), desde cuándo rige y cuánto se cobra después, y
// recién «Confirmar» pasa por Mercado Pago. Las cuentas vienen hechas de la API (abono.Cotizar).

/// Los planes (apps/web/src/lib/panel/planes.ts): nombre, para quién y qué incluye. Los precios salen de
/// las cotizaciones de la API.
enum CatalogoPlanes {
    struct Plan: Identifiable {
        var id: String
        var nombre: String
        var lema: String
        var mensual: Int
        var incluye: [String]
    }
    static let todos: [Plan] = [
        Plan(id: "gratis", nombre: "Gratis", lema: "Para probar y empezar", mensual: 0,
             incluye: ["Productos ilimitados", "Tincho: 100 respuestas con IA", "1 persona · subdominio"]),
        Plan(id: "emprende", nombre: "Emprende", lema: "Para vender todos los días", mensual: 19_900,
             incluye: ["Dominio propio y factura ARCA", "Todos los medios de pago y envío", "Tincho: 1.000 respuestas con IA", "2 personas"]),
        Plan(id: "pro", nombre: "Pro", lema: "Para vender en todos tus canales", mensual: 49_900,
             incluye: ["Mercado Libre, Instagram y local", "Acciones masivas y estadísticas", "Tincho: 3.000 respuestas con IA", "5 personas con permisos"]),
        Plan(id: "negocio", nombre: "Negocio", lema: "Para varias sucursales", mensual: 149_900,
             incluye: ["Sucursales y cajas ilimitadas", "API, webhooks y mayoristas", "Tincho: 15.000 respuestas con IA", "Personas ilimitadas"]),
    ]
    static func de(_ id: String) -> Plan { todos.first { $0.id == id } ?? todos[0] }
    /// Por mes: el anual es 25 % menos (cobrado una vez por año).
    static func porMes(_ p: Plan, anual: Bool) -> Int { anual ? Int((Double(p.mensual) * 0.75).rounded()) : p.mensual }
}

/// Lo que cuesta pasar a un plan y ciclo (packages/contrato: Cotizacion).
struct CotizacionPlan: Decodable {
    struct Cobro: Decodable { var fecha: String; var importe: Int }
    struct Linea: Decodable, Identifiable { var texto: String; var importe: Int; var id: String { texto } }
    var plan: String
    var ciclo: String
    var tipo: String
    var hoy: Int
    var rige: String
    var hasta: String?
    var cobro: Cobro?
    var porCiclo: Int
    var lineas: [Linea]
    var pago: String
    var inicio: String?

    var sinCambio: Bool { tipo == "actual" || tipo == "programado" }
    var pasaPorMercadoPago: Bool { pago == "suscripcion" || pago == "diferencia" }
}

// MARK: - Pantalla

struct PlanCuenta: View {
    @Environment(Estado.self) private var estado
    @State private var hoja: String?

    struct Datos: Decodable {
        struct Uso: Decodable { var productos: Int; var personas: Int; var hornero: Int }
        struct Medio: Decodable { var marca: String; var ultimos: String; var vence: String }
        struct Factura: Decodable, Identifiable {
            var numero: String; var detalle: String; var total: Int; var fecha: String?
            var id: String { numero }
        }
        struct Programado: Decodable { var plan: String; var ciclo: String; var desde: String }
        struct Abono: Decodable { var activo: Bool; var proximo: String?; var programado: Programado? }
        var actual: String
        var ciclo: String
        var enPrueba: Bool
        var diasPrueba: Int
        var finPrueba: String?
        var proximoCobro: String?
        var abono: Abono?
        var cotizaciones: [String: CotizacionPlan]?
        var uso: Uso
        var medio: Medio?
        var facturas: [Factura]

        func cotizacion(_ plan: String, _ ciclo: String) -> CotizacionPlan? {
            cotizaciones?["\(plan)/\(plan == "gratis" ? self.ciclo : ciclo)"]
        }
    }

    var body: some View {
        PantallaSeccion(titulo: "Plan", que: "tu plan") {
            try await estado.api.get("/v1/panel/cuenta/plan") as Datos
        } contenido: { d, actualizar in
            let nombre = CatalogoPlanes.de(d.actual).nombre
            let ciclo = d.actual == "gratis" ? "" : (d.ciclo == "anio" ? " anual" : " mensual")
            TarjetaHeroe(etiqueta: d.enPrueba ? "Estás probando gratis" : "Tu plan", valor: "Plan \(nombre)\(ciclo)",
                         detalle: detalleHeroe(d), tincho: .celebrando)
            if let p = d.abono?.programado {
                AvisoProgramado(programado: p, actual: nombre) { hoja = d.actual }
            }
            GrillaIndicadores(items: [
                IndicadorIcono(simbolo: "shippingbox.fill", titulo: "Productos", valor: "\(d.uso.productos)"),
                IndicadorIcono(simbolo: "person.2.fill", titulo: "Personas del equipo", valor: "\(d.uso.personas)", tono: (C.okFondo, C.okTinta)),
                IndicadorIcono(simbolo: "sparkles", titulo: "Respuestas de Tincho este mes", valor: "\(d.uso.hornero)", tono: (C.avisoFondo, C.avisoTinta)),
            ])
            if let m = d.medio {
                GrupoFilas(elementos: [FilaId("Medio de pago", "\(m.marca) ·· \(m.ultimos)"), FilaId("Vence", m.vence)]) {
                    FilaDato(etiqueta: $0.etiqueta, valor: $0.valor)
                }
            }
            if !d.facturas.isEmpty {
                EtiquetaSeccion(texto: "Facturas").padding(.horizontal, 4).padding(.top, 8)
                GrupoFilas(elementos: d.facturas) { f in
                    FilaDato(etiqueta: f.fecha.map { "\(f.detalle) · \(Formato.diaMes($0))" } ?? f.detalle, valor: Formato.pesos(f.total))
                }
            }
            Button("Cambiar de plan") { hoja = d.actual }
                .buttonStyle(BotonPrincipal(alto: 48)).padding(.top, 6)
                .sheet(item: Binding(get: { hoja.map { Elegido(id: $0) } }, set: { hoja = $0?.id })) { e in
                    HojaPlan(datos: d, preelegido: e.id) { await actualizar() }.hojaTLW([.large])
                }
        }
    }

    private struct Elegido: Identifiable { var id: String }

    private func detalleHeroe(_ d: Datos) -> String {
        if d.enPrueba { return "Te quedan \(Formato.cantidad(d.diasPrueba, "día", "días")) de prueba" }
        if let p = d.abono?.programado, p.plan == "gratis" { return "Se da de baja el \(Formato.diaMes(p.desde))" }
        if d.actual == "gratis" { return "Sin cobros" }
        if let c = d.proximoCobro { return "Próximo cobro: \(Formato.diaMes(c))" }
        return "Sin débito cargado"
    }
}

/// El cambio que rige en el próximo cobro, con «Deshacer».
private struct AvisoProgramado: View {
    var programado: PlanCuenta.Datos.Programado
    var actual: String
    var deshacer: () -> Void
    var body: some View {
        let nuevo = CatalogoPlanes.de(programado.plan).nombre
        HStack(spacing: 12) {
            Image(systemName: "calendar.badge.clock").font(.system(size: 20, weight: .semibold)).foregroundStyle(C.primarioTexto)
            VStack(alignment: .leading, spacing: 2) {
                Text(programado.plan == "gratis" ? "Pasás al plan Gratis el \(Formato.diaMes(programado.desde))"
                     : "Pasás a \(nuevo) el \(Formato.diaMes(programado.desde))")
                    .estilo(.subheadFuerte).foregroundStyle(C.texto)
                Text(Formato.sinViudas("Hasta ese día seguís con \(actual), que ya está pago."))
                    .estilo(.footnote).foregroundStyle(C.texto2)
            }
            Spacer(minLength: 0)
            Button("Deshacer", action: deshacer).buttonStyle(BotonTonal(alto: 40))
        }
        .padding(14)
        .background(C.crema2, in: RoundedRectangle(cornerRadius: 18, style: .continuous))
        .overlay(RoundedRectangle(cornerRadius: 18, style: .continuous).strokeBorder(C.linea2, lineWidth: 1))
    }
}

// MARK: - Elegir y confirmar

struct HojaPlan: View {
    var datos: PlanCuenta.Datos
    var listo: () async -> Void
    @Environment(Estado.self) private var estado
    @Environment(\.dismiss) private var cerrar
    @State private var plan: String
    @State private var ciclo: String
    @State private var enviando = false

    init(datos: PlanCuenta.Datos, preelegido: String? = nil, listo: @escaping () async -> Void) {
        self.datos = datos
        self.listo = listo
        _plan = State(initialValue: preelegido ?? datos.actual)
        _ciclo = State(initialValue: datos.ciclo)
    }

    private var cot: CotizacionPlan? { datos.cotizacion(plan, ciclo) }
    private var hayCambio: Bool { cot.map { !$0.sinCambio } ?? false }

    var body: some View {
        VStack(spacing: 0) {
            Grabber()
            ScrollView {
                VStack(alignment: .leading, spacing: 14) {
                    Text("Elegí tu plan").estilo(.tituloHojaChico).foregroundStyle(C.texto)
                    Text(Formato.sinViudas("Nada se cobra hasta que confirmes. Precios en pesos, con IVA incluido y sin permanencia."))
                        .estilo(.subhead).foregroundStyle(C.texto2)
                    Segmentado(opciones: [("mes", "Por mes"), ("anio", "Por año · −25 %")], elegido: $ciclo)
                    ForEach(CatalogoPlanes.todos) { p in
                        TarjetaPlan(plan: p, anual: ciclo == "anio", elegido: plan == p.id, actual: datos.actual == p.id,
                                    permitido: datos.cotizacion(p.id, ciclo) != nil) {
                            if datos.cotizacion(p.id, ciclo) == nil {
                                estado.avisar("Tenés un cambio programado: deshacelo para elegir otro plan.", ok: false)
                                return
                            }
                            withAnimation(.snappy(duration: 0.22)) { plan = p.id }
                            Haptico.seleccion()
                        }
                    }
                    if let c = cot, hayCambio {
                        ResumenCambio(c: c, datos: datos).transition(.opacity)
                    } else if let c = cot, c.tipo == "programado" {
                        Text("Ese cambio ya está programado: rige el \(Formato.diaMes(c.rige)).")
                            .estilo(.footnote).foregroundStyle(C.texto2)
                    }
                }
                .padding(.horizontal, 20).padding(.top, 8).padding(.bottom, 16)
            }
            if let c = cot, hayCambio {
                BarraConfirmar(c: c, boton: boton(c), enviando: enviando) { Task { await confirmar(c) } }
            }
        }
    }

    private func boton(_ c: CotizacionPlan) -> String {
        if c.pago == "diferencia" || (c.pago == "suscripcion" && c.hoy > 0) { return "Pagar con Mercado Pago" }
        if c.pago == "suscripcion" { return "Seguir a Mercado Pago" }
        if c.tipo == "mantener" { return "Deshacer el cambio" }
        return c.plan == "gratis" ? "Pasar a Gratis" : "Confirmar el cambio"
    }

    private func confirmar(_ c: CotizacionPlan) async {
        struct S: Encodable { var plan: String; var ciclo: String }
        struct U: Decodable { var url: String }
        enviando = true
        defer { enviando = false }
        do {
            let u: U = try await estado.api.enviar("POST", "/v1/panel/cuenta/abono", S(plan: c.plan, ciclo: c.ciclo))
            let nombre = CatalogoPlanes.de(c.plan).nombre
            if let url = URL(string: u.url), !u.url.isEmpty {
                // Mercado Pago: cargar la tarjeta del débito o pagar la diferencia. El plan cambia cuando avisa.
                await ConectarCuenta.compartido.abrir(url)
                await listo()
                estado.avisar("Si completaste el pago, en unos segundos tenés \(nombre).")
            } else {
                await listo()
                Haptico.exito()
                switch c.tipo {
                case "baja": estado.avisar("Pasás a \(nombre) el \(Formato.diaMes(c.rige)). Hasta ahí seguís con lo que pagaste.")
                case "gratis": estado.avisar(c.hasta.map { "Pasás a Gratis el \(Formato.diaMes($0)). No te cobramos más." } ?? "Pasaste al plan Gratis.")
                case "mantener": estado.avisar("Seguís con \(CatalogoPlanes.de(datos.actual).nombre).")
                default: estado.avisar("Ya tenés el plan \(nombre).")
                }
            }
            cerrar()
        } catch let e as ErrorApi {
            Haptico.error()
            estado.avisar(e.detalle, ok: false)
        } catch {
            estado.avisar("No se pudo cambiar el plan. Probá de nuevo.", ok: false)
        }
    }
}

/// Una tarjeta de plan elegible.
private struct TarjetaPlan: View {
    var plan: CatalogoPlanes.Plan
    var anual: Bool
    var elegido: Bool
    var actual: Bool
    var permitido: Bool
    var tocar: () -> Void

    var body: some View {
        Button(action: tocar) {
            VStack(alignment: .leading, spacing: 10) {
                HStack(alignment: .top) {
                    VStack(alignment: .leading, spacing: 2) {
                        HStack(spacing: 8) {
                            Text(plan.nombre).estilo(.nombreTienda).foregroundStyle(C.texto)
                            if actual { ChipTexto(texto: "Tu plan") }
                        }
                        Text(plan.lema).estilo(.footnote).foregroundStyle(C.texto2)
                    }
                    Spacer(minLength: 8)
                    Image(systemName: elegido ? "checkmark.circle.fill" : "circle")
                        .font(.system(size: 22, weight: .semibold))
                        .foregroundStyle(elegido ? C.primarioTexto : C.linea2)
                }
                precio
                VStack(alignment: .leading, spacing: 5) {
                    ForEach(plan.incluye, id: \.self) { f in
                        HStack(alignment: .firstTextBaseline, spacing: 7) {
                            Image(systemName: "checkmark").font(.system(size: 11, weight: .bold)).foregroundStyle(C.ok)
                            Text(f).estilo(.footnote).foregroundStyle(C.texto)
                        }
                    }
                }
            }
            .padding(16)
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(C.superficie, in: RoundedRectangle(cornerRadius: 20, style: .continuous))
            .overlay(RoundedRectangle(cornerRadius: 20, style: .continuous)
                .strokeBorder(elegido ? C.acento : C.anillo, lineWidth: elegido ? 2 : 1))
            .opacity(permitido ? 1 : 0.5)
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
        .accessibilityAddTraits(elegido ? .isSelected : [])
    }

    @ViewBuilder private var precio: some View {
        if plan.mensual == 0 {
            VStack(alignment: .leading, spacing: 0) {
                Text("Gratis").estilo(.montoMeta).foregroundStyle(C.texto)
                Text("Para siempre, sin tarjeta").estilo(.footnote).foregroundStyle(C.texto3)
            }
        } else {
            let mes = CatalogoPlanes.porMes(plan, anual: anual)
            VStack(alignment: .leading, spacing: 0) {
                HStack(alignment: .firstTextBaseline, spacing: 4) {
                    Text(Formato.pesos(mes)).estilo(.montoMeta).foregroundStyle(C.texto)
                    Text("/ mes").estilo(.footnote).foregroundStyle(C.texto2)
                }
                Text(anual ? "\(Formato.pesos(mes * 12)) por año · ahorrás \(Formato.pesos((plan.mensual - mes) * 12))"
                     : "IVA incluido · sin permanencia")
                    .estilo(.footnote).foregroundStyle(C.texto3)
            }
        }
    }
}

private struct ChipTexto: View {
    var texto: String
    var body: some View {
        Text(texto).estilo(.chip).foregroundStyle(.white)
            .padding(.horizontal, 8).padding(.vertical, 3)
            .background(C.primario, in: Capsule())
    }
}

/// El resumen: lo de hoy (con su detalle), lo de después, desde cuándo rige y cómo se paga.
private struct ResumenCambio: View {
    var c: CotizacionPlan
    var datos: PlanCuenta.Datos

    var body: some View {
        let actual = CatalogoPlanes.de(datos.actual).nombre
        let nuevo = CatalogoPlanes.de(c.plan)
        TarjetaSeccion(titulo: "Resumen del cambio", simbolo: "cart.fill") {
            HStack(spacing: 8) {
                Text(actual).estilo(.label).foregroundStyle(C.texto2)
                Image(systemName: "arrow.right").font(.system(size: 12, weight: .bold)).foregroundStyle(C.texto3)
                Text(nuevo.nombre + (c.plan == "gratis" ? "" : (c.ciclo == "anio" ? " anual" : " mensual")))
                    .estilo(.label).foregroundStyle(C.primarioTexto)
            }
            Divider()
            Text("HOY PAGÁS").estilo(.chip).foregroundStyle(C.texto3)
            if c.lineas.isEmpty {
                Text(Formato.sinViudas(nadaHoy(actual: actual))).estilo(.footnote).foregroundStyle(C.texto2)
            } else {
                ForEach(c.lineas) { l in
                    HStack {
                        Text(l.texto).estilo(.texto14).foregroundStyle(l.importe < 0 ? C.okTinta : C.texto)
                        Spacer()
                        Text((l.importe < 0 ? "−" : "") + Formato.pesos(abs(l.importe)))
                            .estilo(.texto14).monospacedDigit().foregroundStyle(l.importe < 0 ? C.okTinta : C.texto)
                    }
                }
            }
            HStack(alignment: .firstTextBaseline) {
                Text("Total hoy").estilo(.headline).foregroundStyle(C.texto)
                Spacer()
                Text(Formato.pesos(c.hoy)).estilo(.montoMeta).monospacedDigit().foregroundStyle(C.texto)
            }
            Text("IVA incluido").estilo(.footnote).foregroundStyle(C.texto3).frame(maxWidth: .infinity, alignment: .trailing)
            Divider()
            Text("DESPUÉS").estilo(.chip).foregroundStyle(C.texto3)
            if let cobro = c.cobro {
                Dato(simbolo: "calendar", texto: proximoCobro(cobro))
                Dato(simbolo: "arrow.triangle.2.circlepath", texto: debito)
            } else {
                Dato(simbolo: "calendar", texto: sinCobros)
            }
            Dato(simbolo: "clock", texto: cuandoRige(nuevo.nombre))
            if !como.isEmpty {
                HStack(spacing: 10) {
                    Image(systemName: c.pasaPorMercadoPago ? "creditcard.fill" : "checkmark.shield.fill")
                        .font(.system(size: 15, weight: .semibold)).foregroundStyle(C.okTinta)
                        .frame(width: 30, height: 30).background(C.okFondo, in: RoundedRectangle(cornerRadius: 9, style: .continuous))
                    Text(Formato.sinViudas(como)).estilo(.footnote).foregroundStyle(C.texto2)
                }
                .padding(12)
                .background(C.crema2, in: RoundedRectangle(cornerRadius: 14, style: .continuous))
            }
            Text("Tu Local Web nunca ve tu tarjeta.").estilo(.footnote).foregroundStyle(C.texto3)
        }
    }

    private func proximoCobro(_ cobro: CotizacionPlan.Cobro) -> String {
        let desde = ["sube", "baja", "ciclo", "mantener"].contains(c.tipo) ? "Desde el próximo cobro" : "Primer cobro"
        let anual = c.ciclo == "anio"
        var t = "\(desde), el \(Formato.diaMes(cobro.fecha)): \(Formato.pesos(c.porCiclo)) \(anual ? "por año" : "por mes")"
        if anual { t += " (\(Formato.pesos(c.porCiclo / 12)) por mes)" }
        return t
    }

    private var debito: String {
        let cada = c.ciclo == "anio" ? "cada año" : "cada mes"
        return "Débito automático \(cada). Sin permanencia: lo das de baja cuando quieras."
    }

    private var sinCobros: String {
        let desde = c.hasta.map { "Desde el \(Formato.diaMes($0))" } ?? "Desde hoy"
        return "\(desde): plan Gratis, sin cobros."
    }

    private func cuandoRige(_ nombre: String) -> String {
        let que = c.plan == "gratis" ? "Pasás a Gratis" : "Tenés \(nombre)"
        return "\(que): \(rige)"
    }

    private var rige: String {
        guard let f = Fecha.iso(c.rige), f > Date().addingTimeInterval(60) else { return "hoy, al confirmar" }
        return "el \(Formato.diaMes(c.rige))"
    }

    private func nadaHoy(actual: String) -> String {
        switch c.tipo {
        case "prueba": return "Estás en la prueba gratis: no pagás nada hasta el \(Formato.diaMes(datos.finPrueba))."
        case "baja", "ciclo": return "Ya pagaste hasta el \(Formato.diaMes(c.hasta)): seguís con \(actual) hasta ese día."
        case "gratis": return c.hasta.map { "Ya pagaste hasta el \(Formato.diaMes($0)): seguís con \(actual) hasta ese día." } ?? "El plan Gratis no tiene cobros."
        case "mantener": return "Deshacés el cambio programado y seguís con \(actual)."
        default: return ""
        }
    }

    private var como: String {
        switch c.pago {
        case "suscripcion": return "Cargás la tarjeta en Mercado Pago y queda el débito automático. Lo das de baja cuando quieras."
        case "diferencia": return "Pagás la diferencia en Mercado Pago (tarjeta, dinero en cuenta o efectivo). Tu débito se actualiza solo."
        case "actualizar": return "Actualizamos el monto de tu débito de Mercado Pago. No tenés que cargar nada."
        case "cancelar": return "Damos de baja tu débito de Mercado Pago: no te cobramos más."
        default: return ""
        }
    }
}

private struct Dato: View {
    var simbolo: String
    var texto: String
    var body: some View {
        HStack(alignment: .firstTextBaseline, spacing: 9) {
            Image(systemName: simbolo).font(.system(size: 13, weight: .semibold)).foregroundStyle(C.primarioTexto)
            Text(Formato.sinViudas(texto)).estilo(.footnote).foregroundStyle(C.texto2)
        }
    }
}

/// El total de hoy y el botón, siempre a mano abajo de la hoja.
private struct BarraConfirmar: View {
    var c: CotizacionPlan
    var boton: String
    var enviando: Bool
    var accion: () -> Void
    var body: some View {
        HStack(spacing: 12) {
            VStack(alignment: .leading, spacing: 0) {
                Text("Hoy pagás").estilo(.footnote).foregroundStyle(C.texto2)
                Text(Formato.pesos(c.hoy)).estilo(.total).monospacedDigit().foregroundStyle(C.texto)
            }
            Button(action: accion) {
                HStack(spacing: 6) {
                    if c.pasaPorMercadoPago { Image(systemName: "lock.fill").font(.system(size: 13, weight: .bold)) }
                    Text(enviando ? "Un momento…" : boton)
                }
            }
            .buttonStyle(BotonPrincipal(alto: 52))
            .disabled(enviando)
        }
        .padding(.horizontal, 20).padding(.top, 10).padding(.bottom, 12)
        .background(C.fondo)
        .overlay(alignment: .top) { Rectangle().fill(C.linea).frame(height: 1) }
    }
}
