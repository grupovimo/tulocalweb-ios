import Foundation
import LocalAuthentication
import Observation
import SwiftUI
import UIKit
import UserNotifications

/// Adónde lleva un push, un widget o un link `tulocalweb://`.
enum Ruta: Hashable {
    case pedido(Int)
    case producto(String)
    case avisos
    case pedidos
    case inicio
    case compartir
    case productos
    /// Una sección del panel que la app tiene (clientes, cupones, finanzas…), sobre Inicio.
    case destino(Destino)
    /// Un link de tulocalweb.com.ar que la app no tiene: se abre la web dentro de la app.
    case web(URL)

    /// `pedido/1048` (el `ruta` del push), `tulocalweb://pedido/1048` o un link del panel (Vinculos).
    init?(texto: String) {
        var t = texto
        if let u = URL(string: texto), u.scheme == "tulocalweb" {
            t = ([u.host ?? ""] + u.pathComponents.filter { $0 != "/" }).joined(separator: "/")
        } else if let u = URL(string: texto), u.scheme == "https" || u.scheme == "http" {
            guard let r = Vinculos.ruta(u) else { return nil }
            self = r
            return
        }
        let partes = t.split(separator: "/").map(String.init)
        guard let primera = partes.first else { return nil }
        switch primera {
        case "pedido": guard partes.count > 1, let n = Int(partes[1]) else { return nil }; self = .pedido(n)
        case "producto": guard partes.count > 1 else { return nil }; self = .producto(partes[1])
        case "avisos", "notificaciones": self = .avisos
        case "pedidos": self = .pedidos
        case "inicio": self = .inicio
        case "compartir": self = .compartir
        // Los push de preguntas de Mercado Libre y de seguridad de la cuenta.
        case "preguntas": self = .destino(.resenas)
        case "cuenta": self = .destino(.cuenta)
        default: return nil
        }
    }
}

/// El estado de la app: la sesión (token en el Keychain), la tienda y sus permisos, el resumen del
/// panel (insignias) y los avisos (toasts). Lo comparten todas las pantallas.
@Observable
@MainActor
final class Estado {
    enum Fase: Equatable {
        case arrancando
        case afuera
        case eligiendoTienda
        case permisos
        /// La tienda recién creada: el onboarding (tu negocio, traer tu tienda, plantilla, cobros).
        case onboarding
        case adentro
    }

    var fase: Fase = .arrancando
    /// Ingreso a medias con la verificación en dos pasos: falta el código (hoja en Ingresar).
    var desafio: String?
    let api: Api
    var datos: DatosSesion?
    var acceso: Acceso?
    var resumen: ResumenPanel?
    var proveedores: Proveedores?
    var aviso: Aviso?
    var rutaPendiente: Ruta?
    /// Pantalla bloqueada con Face ID hasta que la persona se identifique.
    var bloqueada = false
    /// Para que las listas se vuelvan a pedir después de un cambio (estado de un pedido, stock).
    var version = 0
    /// Token de APNs en hexadecimal, cuando iOS lo da.
    private(set) var tokenPush: String?
    private var avisoTarea: Task<Void, Never>?

    /// Ya pasó por los pre-permisos (1e–1g).
    var permisosVistos: Bool {
        get { UserDefaults.standard.bool(forKey: "permisosVistos") }
        set { UserDefaults.standard.set(newValue, forKey: "permisosVistos") }
    }
    /// Una cuenta creada desde la app (o nueva con Apple/Google) que todavía no terminó el onboarding.
    /// Sólo así se muestra: a quien ya usaba la tienda no se le pide de nuevo.
    var onboardingPendiente: Bool {
        get { UserDefaults.standard.bool(forKey: "onboardingPendiente") }
        set { UserDefaults.standard.set(newValue, forKey: "onboardingPendiente") }
    }
    /// Entrar con Face ID (1g, y en Perfil, seguridad y tema).
    var conFaceID: Bool {
        get { UserDefaults.standard.bool(forKey: "faceID") }
        set { UserDefaults.standard.set(newValue, forKey: "faceID") }
    }

    init(api: Api? = nil) {
        self.api = api ?? Api(base: Api.urlConfigurada(), token: Llavero.leer("token"))
        self.api.alPerderSesion = { [weak self] in self?.sesionPerdida() }
    }

    var tienda: Acceso.Tienda? { acceso?.tienda }
    func puede(_ permiso: String) -> Bool { acceso?.puede(permiso) ?? false }

    // MARK: Arranque

    func arrancar() async {
        guard api.token != nil else {
            fase = .afuera
            return
        }
        do {
            try await cargarSesion()
            if conFaceID { bloqueada = true }
            // Se cerró la app a mitad del onboarding: se retoma.
            if onboardingPendiente && acceso?.rol == "duenio" {
                fase = .onboarding
                return
            }
            fase = .adentro
            await refrescarResumen()
            await pedirTokenPushSiCorresponde()
        } catch let e as ErrorApi where e.sinConexion {
            // Sin internet se entra igual con lo guardado (handoff 5a).
            datos = Guardado.leer(DatosSesion.self, clave: "sesion")?.0
            acceso = Guardado.leer(Acceso.self, clave: "acceso")?.0
            if conFaceID { bloqueada = true }
            fase = datos == nil ? .afuera : .adentro
        } catch {
            olvidarSesion()
            fase = .afuera
        }
    }

    private func cargarSesion() async throws {
        async let d: DatosSesion = api.get("/v1/auth/sesion")
        async let a: Acceso = api.get("/v1/auth/tiendas")
        let (dd, aa) = try await (d, a)
        datos = dd
        acceso = aa
        Guardado.guardar(dd, clave: "sesion")
        Guardado.guardar(aa, clave: "acceso")
    }

    /// Vuelve a pedir las tiendas de la persona (por ejemplo, después de cambiar el logo en Mi tienda).
    func refrescarAcceso() async {
        guard let a: Acceso = try? await api.get("/v1/auth/tiendas") else { return }
        acceso = a
        Guardado.guardar(a, clave: "acceso")
    }

    func cargarProveedores() async {
        if proveedores != nil { return }
        proveedores = try? await api.get("/v1/auth/proveedores")
    }

    // MARK: Ingreso

    func ingresar(email: String, clave: String) async throws {
        struct Cuerpo: Encodable { var email: String; var clave: String; var recordarme = true }
        let s: SesionNueva = try await api.enviar("POST", "/v1/auth/ingresar", Cuerpo(email: email, clave: clave))
        try await entrar(s)
    }

    func ingresarConApple() async throws {
        await cargarProveedores()
        guard let clerk = Clerk(publicable: proveedores?.clerkPublicable) else { throw Clerk.ErrorClerk(codigo: "sin_clave") }
        let tokenApple = try await HojaApple().pedirToken()
        let jwt = try await clerk.conApple(tokenApple: tokenApple)
        try await canjearClerk(jwt)
    }

    func ingresarConGoogle() async throws {
        await cargarProveedores()
        guard let clerk = Clerk(publicable: proveedores?.clerkPublicable) else { throw Clerk.ErrorClerk(codigo: "sin_clave") }
        let jwt = try await clerk.conGoogle()
        try await canjearClerk(jwt)
    }

    private func canjearClerk(_ jwt: String) async throws {
        struct Cuerpo: Encodable { var token: String }
        let s: SesionNueva = try await api.enviar("POST", "/v1/auth/clerk", Cuerpo(token: jwt))
        if s.nueva == true { onboardingPendiente = true }
        try await entrar(s)
    }

    /// Crear la cuenta y la tienda desde la app (POST /v1/auth/registro); sigue el onboarding.
    func registrar(nombre: String, tienda: String, email: String, clave: String) async throws {
        struct Cuerpo: Encodable { var nombre: String; var tienda: String; var email: String; var clave: String }
        let s: SesionNueva = try await api.enviar("POST", "/v1/auth/registro", Cuerpo(nombre: nombre, tienda: tienda, email: email, clave: clave))
        onboardingPendiente = true
        try await entrar(s)
    }

    /// Completa el ingreso con el código de la app autenticadora (o uno de recuperación). Si el
    /// desafío venció, se vuelve al formulario.
    func completarSegundoFactor(_ codigo: String) async throws {
        guard let d = desafio else { return }
        struct Cuerpo: Encodable { var desafio: String; var codigo: String }
        do {
            let s: SesionNueva = try await api.enviar("POST", "/v1/auth/segundo-factor", Cuerpo(desafio: d, codigo: codigo))
            desafio = nil
            try await entrar(s)
        } catch let e as ErrorApi where e.estado == 401 {
            desafio = nil
            throw e
        }
    }

    private func entrar(_ s: SesionNueva) async throws {
        if let d = s.segundoFactor, !d.isEmpty {
            desafio = d
            return
        }
        Llavero.guardar(s.token, cuenta: "token")
        api.token = s.token
        Guardado.borrarTodo()
        try await cargarSesion()
        if (acceso?.tiendas.count ?? 0) > 1 {
            fase = .eligiendoTienda
        } else {
            seguirDespuesDeElegir()
        }
    }

    func seguirDespuesDeElegir() {
        if onboardingPendiente && acceso?.rol == "duenio" {
            fase = .onboarding
        } else if permisosVistos {
            fase = .adentro
            Task {
                await refrescarResumen()
                await pedirTokenPushSiCorresponde()
            }
        } else {
            fase = .permisos
        }
    }

    /// Fin del onboarding (o «lo hago después»): sigue a los permisos o al panel.
    func terminarOnboarding() {
        onboardingPendiente = false
        version += 1
        if permisosVistos {
            fase = .adentro
            Task {
                await refrescarResumen()
                await pedirTokenPushSiCorresponde()
            }
        } else {
            fase = .permisos
        }
    }

    func terminarPermisos() {
        permisosVistos = true
        fase = .adentro
        Task { await refrescarResumen() }
    }

    func elegirTienda(_ id: String) async throws {
        struct Cuerpo: Encodable { var tiendaId: String }
        if id != acceso?.tiendaId {
            try await api.hacer("PUT", "/v1/auth/tienda", Cuerpo(tiendaId: id))
            Guardado.borrarTodo()
            try await cargarSesion()
            version += 1
            await refrescarResumen()
        }
    }

    // MARK: Salir

    func salir() async {
        if tokenPush != nil { try? await api.hacer("DELETE", "/v1/panel/dispositivos") }
        try? await api.hacer("POST", "/v1/auth/salir")
        olvidarSesion()
        fase = .afuera
    }

    private func sesionPerdida() {
        guard fase == .adentro || fase == .eligiendoTienda || fase == .permisos || fase == .onboarding else { return }
        olvidarSesion()
        fase = .afuera
        avisar("Tu sesión venció. Entrá de nuevo.", ok: false)
    }

    private func olvidarSesion() {
        Llavero.borrar("token")
        api.token = nil
        onboardingPendiente = false
        ActividadEnApp.terminarTodo()
        datos = nil
        acceso = nil
        resumen = nil
        bloqueada = false
        Guardado.borrarTodo()
        UNUserNotificationCenter.current().setBadgeCount(0) { _ in }
    }

    // MARK: Resumen e insignias

    func refrescarResumen() async {
        guard api.token != nil else { return }
        if let r: ResumenPanel = try? await api.get("/v1/panel/resumen") {
            resumen = r
            try? await UNUserNotificationCenter.current().setBadgeCount(r.insignias.pedidos)
            await ActividadEnApp.sincronizar(tienda: tienda?.nombre ?? "Tu Local Web")
        }
    }

    // MARK: Avisos (toasts)

    func avisar(_ texto: String, ok: Bool = true, deshacer: (() -> Void)? = nil) {
        aviso = Aviso(texto: texto, ok: ok, deshacer: deshacer)
        avisoTarea?.cancel()
        avisoTarea = Task { [weak self] in
            try? await Task.sleep(for: .seconds(4))
            guard !Task.isCancelled else { return }
            withAnimation(.easeOut(duration: 0.2)) { self?.aviso = nil }
        }
    }

    // MARK: Push

    /// Pide el token de APNs si la persona ya dio permiso (el permiso se pide en 1e).
    func pedirTokenPushSiCorresponde() async {
        let ajustes = await UNUserNotificationCenter.current().notificationSettings()
        if ajustes.authorizationStatus == .authorized || ajustes.authorizationStatus == .provisional {
            UIApplication.shared.registerForRemoteNotifications()
        }
    }

    /// 1e: el diálogo del sistema y, si dice que sí, el registro.
    func pedirPermisoPush() async -> Bool {
        let ok = (try? await UNUserNotificationCenter.current().requestAuthorization(options: [.alert, .sound, .badge])) ?? false
        if ok { UIApplication.shared.registerForRemoteNotifications() }
        return ok
    }

    func registrarTokenPush(_ hex: String) async {
        tokenPush = hex
        ActividadEnApp.api = api
        ActividadEnApp.tokenPush = hex
        ActividadEnApp.arrancarEscucha()
        guard api.token != nil else { return }
        struct Dispositivo: Encodable {
            var token: String
            var entorno: String
            var modelo: String
            var version: String
        }
        #if DEBUG
        let entorno = "sandbox"
        #else
        let entorno = "produccion"
        #endif
        var info = utsname()
        uname(&info)
        let modelo = withUnsafeBytes(of: &info.machine) { b in
            String(decoding: b.prefix { $0 != 0 }, as: UTF8.self)
        }
        let version = Bundle.main.object(forInfoDictionaryKey: "CFBundleShortVersionString") as? String ?? ""
        try? await api.hacer("POST", "/v1/panel/dispositivos",
                             Dispositivo(token: hex, entorno: entorno, modelo: String(modelo.prefix(60)), version: version))
    }

    // MARK: Face ID

    /// Face ID (o el código del iPhone) para entrar y para lo delicado (cancelar un pedido).
    func confirmarIdentidad(_ motivo: String) async -> Bool {
        let ctx = LAContext()
        var error: NSError?
        guard ctx.canEvaluatePolicy(.deviceOwnerAuthentication, error: &error) else { return true }
        return (try? await ctx.evaluatePolicy(.deviceOwnerAuthentication, localizedReason: motivo)) ?? false
    }

    func desbloquear() async {
        if await confirmarIdentidad("Entrá a tu tienda") { bloqueada = false }
    }
}
