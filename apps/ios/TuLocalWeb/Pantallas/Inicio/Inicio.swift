import SwiftUI

/// 2a · Inicio, «La vidriera de tu local» (docs/handoff/inicio-a): arriba la vidriera con el color de
/// la tienda (cuánto se vendió, el gráfico de hoy contra ayer y «Vender ahora»), la tarjeta de Tincho
/// montada sobre su borde, las acciones en un bento, lo que hay para hacer hoy, un vistazo de costado,
/// los últimos pedidos, lo más vendido y el resto del tablero. Al bajar, la vidriera se pliega a una
/// barra. Con /v1/panel/tablero.
struct Inicio: View {
    @Environment(Estado.self) private var estado
    @Environment(Navegacion.self) private var nav
    @Environment(\.anchoPantalla) private var ancho
    @Environment(\.colorScheme) private var esquema
    @State private var periodo = "hoy"
    @State private var carga: Carga<Tablero> = .cargando
    @State private var masViejo: Date?
    @State private var sinLeer = 0
    @State private var editarMeta = false
    @State private var ideasOcultas: Set<String> = []
    @State private var plegada = false
    /// El color de la última carga («#RRGGBB|#RRGGBB»), para que el esqueleto ya salga con el de la tienda.
    @AppStorage("inicio.colores") private var colores = ""

    private var se: Bool { ancho <= 375 }

    private var paleta: PaletaTienda {
        let guardados = colores.split(separator: "|").map(String.init)
        return PaletaTienda(acento: carga.valor?.acento ?? guardados.first, sobre: carga.valor?.sobreAcento ?? guardados.last,
                            oscuro: esquema == .dark)
    }

    var body: some View {
        let p = paleta
        ScrollView {
            VStack(alignment: .leading, spacing: 0) {
                Vidriera(carga: carga, periodo: $periodo, sinLeer: sinLeer, paleta: p, se: se)
                    .onGeometryChange(for: Bool.self) { $0.frame(in: .scrollView).maxY < 140 } action: { v in
                        withAnimation(Mov.toque) { plegada = v }
                    }
                switch carga {
                case .cargando:
                    EsqueletoInicio(paleta: p)
                case let .error(e):
                    EstadoError(que: "cómo viene el día", error: e) { Task { await cargar() } }
                        .padding(.top, 30).margenPantalla()
                case let .listo(t):
                    contenido(t, p)
                }
            }
            .padding(.bottom, 32)
        }
        .scrollIndicators(.hidden)
        .background(C.fondo.overlay(Grano(opacidad: 0.06, mezcla: .multiply)).ignoresSafeArea())
        .overlay(alignment: .top) {
            if plegada, let t = carga.valor {
                BarraVidriera(tablero: t, periodo: periodo, paleta: p)
                    .transition(.move(edge: .top).combined(with: .opacity))
            }
        }
        .toolbar(.hidden, for: .navigationBar)
        // La barra de estado con el color de la vidriera: íconos claros sobre un acento oscuro.
        .toolbarColorScheme(paleta.claroEncima ? .dark : nil, for: .navigationBar)
        .navigationTitle("Inicio")
        .refreshable { await cargar() }
        .task(id: "\(periodo)-\(estado.version)") { await cargar() }
        .sheet(isPresented: $editarMeta) {
            if let t = carga.valor {
                EditarMeta(actual: t.meta?.objetivo ?? 0, vendido: t.meta?.actual ?? 0) { await cargar() }
                    .hojaTLW([.height(470)])
            }
        }
    }

    // MARK: Contenido

    @ViewBuilder
    private func contenido(_ t: Tablero, _ p: PaletaTienda) -> some View {
        let cero = Vidriera.enCero(t)
        if !cero, let idea = ideaPrincipal(t) {
            TarjetaTincho(idea: idea, paleta: p)
                .padding(.horizontal, 16).padding(.top, -28)
        }
        VStack(alignment: .leading, spacing: 32) {
            if cero {
                AccionesCero(paleta: p)
                if !PrimerosPasos.ocultos(estado.tienda?.id) { PasosPrimeraVenta(paleta: p) }
            } else {
                AccionesInicio(tablero: t, paleta: p)
            }
            ParaHacerHoy(tablero: t, masViejo: masViejo, paleta: p, cero: cero)
            DeUnVistazo(tablero: t, periodo: periodo, paleta: p) { editarMeta = true }
            if !t.ultimosPedidos.isEmpty { UltimosPedidos(tablero: t, paleta: p) }
            if !t.masVendidos.isEmpty { LoMasVendido(tablero: t) }
            if let v = t.visitas, !v.masVistos.isEmpty { TarjetaMasVistos(visitas: v, paleta: p) }
            if let r = t.resenas, !r.nuevas.isEmpty { TarjetaResenas(resenas: r, paleta: p) }
            if let e = t.envios, e.paquetes > 0 || !e.retiros.isEmpty { TarjetaEnvios(envios: e, paleta: p) }
            if estado.puede("cuenta") { TarjetaIntegraciones(paleta: p) }
        }
        .padding(.top, 24)
        .margenPantalla()
    }

    // MARK: Tincho

    /// La idea más relevante de Tincho: carritos por recordar, después stock, después compartir; si no
    /// hay ninguna, lo más vendido o una invitación a preguntarle.
    private func ideaPrincipal(_ t: Tablero) -> IdeaTincho? {
        var r: [IdeaTincho] = []
        let negrita = Estilo(tamano: 16, peso: .bold, relativo: .body).font
        // Ya recordados: el mail salió y no se repite antes de 24 h, así que la idea no vuelve.
        if t.hornero.carritos > 0, !t.hornero.ideasHechas.contains("carritos"), t.hornero.recordados != true {
            var s = AttributedString(t.hornero.carritos == 1 ? "Hay 1 carrito abandonado" : "Hay \(t.hornero.carritos) carritos abandonados")
            if let m = t.hornero.carritosMonto, t.montos {
                s += AttributedString(" por ")
                var monto = AttributedString(Formato.pesos(m))
                monto.font = negrita
                s += monto
            }
            s += AttributedString(Formato.sinViudas(". ¿Les mando un mail para recordarles lo que dejaron?"))
            r.append(IdeaTincho(id: "carritos", texto: s, boton: "Sí, mandales", pregunta: "¿Qué dejaron en los carritos abandonados?") {
                await recordarCarritos()
            })
        }
        if t.paraHacer.sinStock > 0, !t.hornero.ideasHechas.contains("stock") {
            let n = t.paraHacer.sinStock
            let s = AttributedString(Formato.sinViudas(n == 1
                ? "Un producto se quedó sin stock y sigue visible en la tienda. ¿Le cargás un ingreso?"
                : "\(n) productos se quedaron sin stock y siguen visibles en la tienda. ¿Les cargás un ingreso?"))
            r.append(IdeaTincho(id: "stock", texto: s, boton: "Reponer", pregunta: "¿Qué productos tengo que reponer?") {
                nav.segmentoProductos = "out"
                nav.tocar(.productos)
            })
        }
        if let v = t.visitas, !t.hornero.ideasHechas.contains("compartir") {
            let n = Int(v.dias7.actual)
            if v.desde == nil || n < 30 {
                let s = AttributedString(Formato.sinViudas(v.desde == nil || n == 0
                    ? "Tu tienda todavía no tuvo visitas. ¿Te armo una imagen con el QR para tus estados de WhatsApp?"
                    : "En 7 días te visitaron \(n == 1 ? "1 persona" : "\(n) personas"). Con una historia o un estado llegás a muchas más. ¿La armamos?"))
                r.append(IdeaTincho(id: "compartir", texto: s, boton: "Armarla", pregunta: "¿Cómo consigo más visitas en mi tienda?") {
                    nav.inicio.append(.compartir)
                })
            }
        }
        if let p = t.masVendidos.first {
            var s = AttributedString(p.nombre)
            s.font = negrita
            s += AttributedString(Formato.sinViudas(" es lo que más sale: \(Formato.cantidad(p.unidades, "unidad", "unidades")). Compartilo para que llegue a más gente."))
            r.append(IdeaTincho(id: "estrella", texto: s, boton: "Compartir", pregunta: "¿Cómo viene \(p.nombre)?") {
                nav.inicio.append(.compartir)
            })
        }
        r.append(IdeaTincho(id: "preguntar", texto: AttributedString(Formato.sinViudas("Preguntame lo que quieras de tu negocio: cuánto vendiste, qué reponer o cómo vienen tus clientes.")),
                            boton: nil, pregunta: "¿Cómo vengo esta semana?") {})
        return r.first { !ideasOcultas.contains($0.id) }
    }

    private func recordarCarritos() async {
        struct R: Decodable { var enviados: Int }
        do {
            let r: R = try await estado.api.enviar("POST", "/v1/panel/carritos/recordar", [String: String]())
            Haptico.exito()
            if r.enviados == 0 {
                // Ya se les mandó en las últimas 24 horas (no se repite para no molestar).
                estado.avisar("Ya les mandamos un mail hace menos de un día. Mañana podés volver a recordarles.")
            } else {
                estado.avisar(r.enviados == 1 ? "Listo: le mandamos un mail a 1 cliente" : "Listo: les mandamos un mail a \(r.enviados) clientes")
            }
            ideasOcultas.insert("carritos")
            estado.version += 1
        } catch let e as ErrorApi {
            estado.avisar(e.detalle, ok: false)
        } catch {}
    }

    // MARK: Datos

    private func cargar() async {
        do {
            let t: Tablero = try await estado.api.get("/v1/panel/tablero", query: [URLQueryItem(name: "periodo", value: periodo)])
            carga = .listo(t)
            Guardado.guardar(t, clave: "tablero-\(periodo)")
            if let a = t.acento, let s = t.sobreAcento { colores = "\(a)|\(s)" }
            if t.paraHacer.preparar > 0 {
                let l: PaginaPedidos? = try? await estado.api.get("/v1/panel/pedidos", query: [
                    URLQueryItem(name: "estado", value: "Para preparar"), URLQueryItem(name: "orden", value: "viejos"),
                    URLQueryItem(name: "cuantos", value: "1"),
                ])
                masViejo = l?.filas.first?.creado
            }
            struct SinLeer: Decodable { var sinLeer: Int }
            if let s: SinLeer = try? await estado.api.get("/v1/panel/notificaciones/sin-leer") { sinLeer = s.sinLeer }
        } catch is CancellationError {
        } catch let e as ErrorApi {
            if e.sinConexion, let (t, _) = Guardado.leer(Tablero.self, clave: "tablero-\(periodo)") {
                carga = .listo(t)
            } else if carga.valor == nil {
                carga = .error(e)
            }
        } catch {
            carga = .error(.desconocido)
        }
    }
}

/// Indicador chico (radio 18): título 13, valor 22/700 y delta 12/600 coloreado. Lo usan Clientes y las
/// secciones de Más.
struct Indicador: View {
    var titulo: String
    var valor: String
    var delta: String?
    var positivo: Bool
    var body: some View {
        VStack(alignment: .leading, spacing: 2) {
            Text(titulo).estilo(.footnote).foregroundStyle(C.texto2).lineLimit(2)
            Text(valor).estilo(.indicador).monospacedDigit().foregroundStyle(C.texto)
                .lineLimit(1).minimumScaleFactor(0.6)
            if let delta {
                Text(delta).estilo(.chip).foregroundStyle(positivo ? C.okTinta : C.errorTinta).lineLimit(1).minimumScaleFactor(0.8)
            }
        }
        .padding(.horizontal, 13).padding(.vertical, 12)
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
        .tarjeta(radio: 18)
        .accessibilityElement(children: .combine)
    }
}

/// Los primeros pasos de Inicio: qué tiendas ya compartieron el link o cerraron la tarjeta (en el
/// teléfono, por tienda).
enum PrimerosPasos {
    static let claveCompartido = "pasos.compartido"
    static let claveOcultos = "pasos.ocultos"

    static func contiene(_ lista: String, _ id: String) -> Bool {
        !id.isEmpty && lista.split(separator: ",").contains(Substring(id))
    }

    static func sumar(_ lista: String, _ id: String) -> String {
        guard !id.isEmpty, !contiene(lista, id) else { return lista }
        return lista.isEmpty ? id : lista + "," + id
    }

    static func ocultos(_ id: String?) -> Bool {
        contiene(UserDefaults.standard.string(forKey: claveOcultos) ?? "", id ?? "")
    }
}
