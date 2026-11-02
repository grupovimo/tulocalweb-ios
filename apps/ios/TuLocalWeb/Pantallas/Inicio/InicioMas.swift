import SwiftUI

// Lo que el Inicio de la web mostraba y la app no: preguntarle a Tincho, quién está ahora en la
// tienda, la caja del local, las próximas liquidaciones, las integraciones, lo más visto, las reseñas
// nuevas y los envíos de hoy. Cada tarjeta aparece sólo si tiene algo que mostrar.

/// La barra «Preguntale a Tincho» (como la del Inicio de la web): lleva a Tincho con la pregunta.
struct BarraPreguntarTincho: View {
    @Environment(Navegacion.self) private var nav
    private let ejemplos = ["¿Cuánto vendí ayer?", "¿Qué tengo que reponer?", "¿Cómo vengo esta semana?", "Ideas para vender más"]

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            Button { nav.pestana = .tincho } label: {
                HStack(spacing: 12) {
                    Image(systemName: "sparkles").font(.system(size: 16, weight: .semibold)).foregroundStyle(.white)
                        .frame(width: 36, height: 36)
                        .background(LinearGradient(colors: [Color(hex: 0xE8875F), Color(hex: 0xC24E25)], startPoint: .topLeading, endPoint: .bottomTrailing),
                                    in: RoundedRectangle(cornerRadius: 11, style: .continuous))
                    VStack(alignment: .leading, spacing: 1) {
                        Text("Preguntale a Tincho").estilo(.fila).foregroundStyle(C.texto)
                        Text("«¿Cuánto vendí ayer por Mercado Libre?»").estilo(.footnote).foregroundStyle(C.texto2).lineLimit(1)
                    }
                    Spacer(minLength: 4)
                    Image(systemName: "arrow.up.circle.fill").font(.system(size: 22)).foregroundStyle(C.primarioTexto)
                }
                .padding(12)
                .background(C.primario.opacity(0.09), in: RoundedRectangle(cornerRadius: 16, style: .continuous))
            }
            .buttonStyle(.plain)
            ScrollView(.horizontal, showsIndicators: false) {
                HStack(spacing: 8) {
                    ForEach(ejemplos, id: \.self) { e in
                        Button {
                            nav.preguntaTincho = e
                            nav.pestana = .tincho
                        } label: {
                            Text(e).estilo(.footnoteFuerte).foregroundStyle(C.primarioTinta)
                                .padding(.horizontal, 12).padding(.vertical, 7)
                                .background(C.crema2, in: Capsule())
                                .overlay(Capsule().strokeBorder(C.linea2, lineWidth: 1))
                        }
                        .buttonStyle(.plain)
                    }
                }
            }
        }
    }
}

/// «Ahora en tu tienda»: cuántas personas están mirando, en qué páginas y los carritos armándose.
struct TarjetaEnVivo: View {
    var vivo: Tablero.EnVivo
    var montos: Bool
    @State private var latido = false

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Text("Ahora en tu tienda").estilo(.tituloSeccion).foregroundStyle(C.texto).accessibilityAddTraits(.isHeader)
                Spacer()
                HStack(spacing: 6) {
                    Circle().fill(C.ok).frame(width: 8, height: 8)
                        .scaleEffect(latido ? 1.35 : 0.85).opacity(latido ? 0.6 : 1)
                        .animation(.easeInOut(duration: 0.9).repeatForever(autoreverses: true), value: latido)
                    Text("En vivo").estilo(.footnoteFuerte).foregroundStyle(C.okTinta)
                }
            }
            .padding(.top, 14)
            VStack(alignment: .leading, spacing: 10) {
                HStack(alignment: .firstTextBaseline, spacing: 8) {
                    Text("\(max(0, vivo.personas))").estilo(.saludo).monospacedDigit().foregroundStyle(C.texto)
                    Text(vivo.personas == 1 ? "persona mirando" : "personas mirando").estilo(.subhead).foregroundStyle(C.texto2)
                }
                if !vivo.paginas.isEmpty {
                    VStack(spacing: 0) {
                        ForEach(Array(vivo.paginas.prefix(4).enumerated()), id: \.offset) { i, p in
                            if i > 0 { Separador(inset: 0) }
                            HStack {
                                Text(p.nombre).estilo(.texto14).foregroundStyle(C.texto).lineLimit(1)
                                Spacer()
                                Text("\(p.personas)").estilo(.fila).monospacedDigit().foregroundStyle(C.texto2)
                            }
                            .padding(.vertical, 8)
                        }
                    }
                }
                if vivo.carritos > 0 {
                    (Text(Formato.cantidad(vivo.carritos, "carrito", "carritos")).fontWeight(.bold)
                        + Text(" armándose ahora")
                        + (montos ? Text(vivo.carritosMonto.map { " por \(Formato.pesos($0))" } ?? "").fontWeight(.bold) : Text("")))
                        .estilo(.texto14).foregroundStyle(C.texto)
                        .padding(10).frame(maxWidth: .infinity, alignment: .leading)
                        .background(C.avisoFondo, in: RoundedRectangle(cornerRadius: 12, style: .continuous))
                }
            }
            .padding(16)
            .tarjeta()
        }
        .onAppear { latido = true }
    }
}

/// «Caja del local»: lo cobrado hoy en el punto de venta, por medio.
struct TarjetaCajaLocal: View {
    var caja: Tablero.Caja
    @Environment(Navegacion.self) private var nav

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            TituloSeccion(texto: "Caja del local", accion: caja.abierta == nil ? "Abrir" : "Ver") { nav.inicio.append(.caja) }
                .padding(.top, 14)
            VStack(alignment: .leading, spacing: 10) {
                HStack(alignment: .firstTextBaseline) {
                    Text(Formato.pesos(caja.total)).estilo(.montoPantalla).monospacedDigit().foregroundStyle(C.texto)
                    Spacer()
                    ChipSimple(texto: caja.abierta == nil ? "Cerrada" : "Abierta",
                               fondo: caja.abierta == nil ? C.pista : C.okFondo, tinta: caja.abierta == nil ? C.texto2 : C.okTinta)
                }
                if !caja.medios.isEmpty {
                    let total = max(1, caja.medios.map(\.monto).reduce(0, +))
                    GeometryReader { g in
                        HStack(spacing: 2) {
                            ForEach(Array(caja.medios.enumerated()), id: \.offset) { i, m in
                                Rectangle().fill(C.acento.opacity(1 - Double(i) * 0.22))
                                    .frame(width: max(4, g.size.width * CGFloat(m.monto) / CGFloat(total)))
                            }
                        }
                        .clipShape(Capsule())
                    }
                    .frame(height: 8)
                    ForEach(Array(caja.medios.enumerated()), id: \.offset) { i, m in
                        HStack(spacing: 8) {
                            Circle().fill(C.acento.opacity(1 - Double(i) * 0.22)).frame(width: 8, height: 8)
                            Text(Marcas.pago(m.medio)).estilo(.texto14).foregroundStyle(C.texto)
                            Spacer()
                            Text(Formato.pesos(m.monto)).estilo(.fila).monospacedDigit().foregroundStyle(C.texto)
                        }
                    }
                }
            }
            .padding(16)
            .tarjeta()
        }
    }
}

/// «Próximas liquidaciones»: cuándo y cuánto deposita cada medio.
struct TarjetaLiquidaciones: View {
    var liquidaciones: [Tablero.Liquidacion]
    @Environment(Navegacion.self) private var nav

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            TituloSeccion(texto: "Próximas liquidaciones", accion: "Finanzas") { nav.inicio.append(.finanzas) }
                .padding(.top, 14)
            VStack(spacing: 0) {
                ForEach(Array(liquidaciones.prefix(4).enumerated()), id: \.offset) { i, l in
                    if i > 0 { Separador(inset: 16) }
                    HStack {
                        VStack(alignment: .leading, spacing: 2) {
                            Text(Marcas.pago(l.medio)).estilo(.fila).foregroundStyle(C.texto)
                            Text(Fecha.iso(l.fecha).map { Formato.dia($0) } ?? l.fecha).estilo(.footnote).foregroundStyle(C.texto2)
                        }
                        Spacer()
                        Text(Formato.pesos(l.monto)).estilo(.fila).monospacedDigit().foregroundStyle(C.okTinta)
                    }
                    .padding(.horizontal, 16).padding(.vertical, 12)
                }
            }
            .tarjeta()
        }
    }
}

/// «Más vistos»: los productos que más miraron en la tienda online.
struct TarjetaMasVistos: View {
    var visitas: Tablero.Visitas

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack(alignment: .firstTextBaseline) {
                Text("Más vistos").estilo(.tituloSeccion).foregroundStyle(C.texto).accessibilityAddTraits(.isHeader)
                Spacer()
                Text("En la tienda online").estilo(.subhead).foregroundStyle(C.texto2)
            }
            .padding(.top, 14)
            let max = Double(visitas.masVistos.map(\.vistas).max() ?? 1)
            VStack(spacing: 0) {
                ForEach(Array(visitas.masVistos.prefix(5).enumerated()), id: \.element.id) { i, p in
                    if i > 0 { Separador(inset: 68) }
                    HStack(spacing: 12) {
                        Miniatura(foto: p.foto, tam: 42, radio: 11)
                        VStack(alignment: .leading, spacing: 5) {
                            Text(p.nombre).estilo(.fila).foregroundStyle(C.texto).lineLimit(1)
                            GeometryReader { g in
                                Capsule().fill(C.acento.opacity(0.5)).frame(width: g.size.width * CGFloat(Double(p.vistas) / max))
                            }
                            .frame(height: 5)
                        }
                        Text("\(Formato.numero(p.vistas))").estilo(.fila).monospacedDigit().foregroundStyle(C.texto2)
                    }
                    .padding(.horizontal, 14).padding(.vertical, 11)
                }
            }
            .tarjeta()
        }
    }
}

/// «Reseñas nuevas»: las últimas sin responder.
struct TarjetaResenas: View {
    var resenas: Tablero.Resenas
    @Environment(Navegacion.self) private var nav

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            TituloSeccion(texto: "Reseñas nuevas", accion: "Responder") { nav.inicio.append(.resenas) }
                .padding(.top, 14)
            VStack(alignment: .leading, spacing: 0) {
                if let p = resenas.promedio {
                    HStack(spacing: 6) {
                        Text(p.formatted(.number.precision(.fractionLength(1)).locale(Locale(identifier: "es_AR"))))
                            .estilo(.headline).foregroundStyle(C.texto)
                        Estrellas(n: Int(p.rounded()))
                        Text("promedio").estilo(.footnote).foregroundStyle(C.texto2)
                    }
                    .padding(.horizontal, 16).padding(.vertical, 12)
                }
                ForEach(Array(resenas.nuevas.prefix(3).enumerated()), id: \.offset) { _, r in
                    Separador(inset: 16)
                    VStack(alignment: .leading, spacing: 4) {
                        HStack(spacing: 6) {
                            Estrellas(n: r["estrellas"]?.entero ?? 0)
                            Text(r["autor"]?.texto ?? "").estilo(.footnoteFuerte).foregroundStyle(C.texto)
                            Spacer()
                            Text(r["producto"]?.texto ?? "").estilo(.footnote).foregroundStyle(C.texto2).lineLimit(1)
                        }
                        Text("«\(r["texto"]?.texto ?? "")»").estilo(.texto14).foregroundStyle(C.texto).lineLimit(3)
                    }
                    .padding(.horizontal, 16).padding(.vertical, 12)
                }
            }
            .tarjeta()
        }
    }
}

/// «Envíos de hoy»: paquetes y retiros de los transportes.
struct TarjetaEnvios: View {
    var envios: Tablero.Envios
    @Environment(Navegacion.self) private var nav

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            TituloSeccion(texto: "Envíos de hoy", accion: "Preparar") {
                nav.filtroPedidos = "Para preparar"
                nav.pestana = .pedidos
            }
            .padding(.top, 14)
            VStack(spacing: 0) {
                HStack {
                    Image(systemName: "shippingbox.fill").foregroundStyle(C.primarioTexto)
                    Text(Formato.cantidad(envios.paquetes, "paquete para despachar", "paquetes para despachar")).estilo(.fila).foregroundStyle(C.texto)
                    Spacer()
                }
                .padding(.horizontal, 16).padding(.vertical, 12)
                ForEach(Array(envios.retiros.enumerated()), id: \.offset) { _, r in
                    Separador(inset: 16)
                    HStack {
                        Text(r.hora).estilo(.fila).monospacedDigit().foregroundStyle(r.proximo ? C.primarioTexto : C.texto)
                            .frame(width: 54, alignment: .leading)
                        VStack(alignment: .leading, spacing: 1) {
                            Text(Formato.mayuscula(r.transporte)).estilo(.texto14).foregroundStyle(C.texto)
                            if let c = r.ciudad { Text(c).estilo(.footnote).foregroundStyle(C.texto2) }
                        }
                        Spacer()
                        Text(Formato.cantidad(r.paquetes, "paquete", "paquetes")).estilo(.footnote).foregroundStyle(C.texto2)
                        if r.proximo { ChipSimple(texto: "Próximo", fondo: C.avisoFondo, tinta: C.avisoTinta) }
                    }
                    .padding(.horizontal, 16).padding(.vertical, 11)
                }
            }
            .tarjeta()
        }
    }
}

/// «Integraciones»: lo conectado (Mercado Pago, Mercado Libre, Correo) y un acceso para sumar más.
struct TarjetaIntegraciones: View {
    @Environment(Estado.self) private var estado
    @Environment(Navegacion.self) private var nav
    @State private var conectadas: [(String, String)]?

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            TituloSeccion(texto: "Integraciones", accion: "Explorar") { nav.inicio.append(.integraciones) }
                .padding(.top, 14)
            VStack(spacing: 0) {
                if let c = conectadas, !c.isEmpty {
                    ForEach(Array(c.enumerated()), id: \.offset) { i, x in
                        if i > 0 { Separador(inset: 16) }
                        HStack {
                            Text(x.0).estilo(.fila).foregroundStyle(C.texto)
                            Spacer()
                            Text(x.1).estilo(.footnote).foregroundStyle(C.texto2).lineLimit(1)
                            Circle().fill(C.ok).frame(width: 8, height: 8)
                        }
                        .padding(.horizontal, 16).padding(.vertical, 12)
                    }
                } else if conectadas != nil {
                    Button { nav.inicio.append(.integraciones) } label: {
                        Text(Formato.sinViudas("Conectá Mercado Pago, Mercado Libre y Correo Argentino para cobrar, publicar y enviar."))
                            .estilo(.texto14).foregroundStyle(C.texto2).multilineTextAlignment(.leading)
                            .padding(16).frame(maxWidth: .infinity, alignment: .leading)
                    }
                    .buttonStyle(.plain)
                } else {
                    ProgressView().tint(C.acento).padding(16)
                }
            }
            .tarjeta()
        }
        .task { await cargar() }
    }

    private func cargar() async {
        guard conectadas == nil else { return }
        struct I: Decodable { var conectada: Bool?; var usuario: String?; var apodo: String?; var email: String? }
        var c: [(String, String)] = []
        for (id, nombre) in [("mercadopago", "Mercado Pago"), ("mercadolibre", "Mercado Libre"), ("correo", "Correo Argentino")] {
            if let i: I = try? await estado.api.get("/v1/panel/integraciones/\(id)"), i.conectada == true {
                c.append((nombre, i.usuario ?? i.apodo ?? i.email ?? "Conectada"))
            }
        }
        conectadas = c
    }
}
