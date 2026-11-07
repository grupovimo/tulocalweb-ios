package ar.com.tulocalweb.app

import androidx.compose.foundation.layout.Column
import ar.com.tulocalweb.app.nucleo.Web
import ar.com.tulocalweb.app.nucleo.DatosSesion
import androidx.compose.ui.unit.dp
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.ColumnScope
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.mutableStateOf
import androidx.compose.ui.Modifier
import androidx.compose.ui.test.junit4.createComposeRule
import androidx.compose.ui.test.onRoot
import org.robolectric.RuntimeEnvironment
import ar.com.tulocalweb.app.app.Estado
import ar.com.tulocalweb.app.app.LocalEstado
import ar.com.tulocalweb.app.app.LocalNavegacion
import ar.com.tulocalweb.app.app.Navegacion
import ar.com.tulocalweb.app.app.Principal
import ar.com.tulocalweb.app.nucleo.Acceso
import ar.com.tulocalweb.app.nucleo.Api
import ar.com.tulocalweb.app.nucleo.Preferencias
import ar.com.tulocalweb.app.nucleo.Tablero
import ar.com.tulocalweb.app.nucleo.TipoProducto
import ar.com.tulocalweb.app.nucleo.AtributoProducto
import ar.com.tulocalweb.app.pantallas.avisos.Avisos
import ar.com.tulocalweb.app.pantallas.buscar.Buscar
import ar.com.tulocalweb.app.pantallas.clientes.Clientes
import ar.com.tulocalweb.app.pantallas.compartir.CompartirTienda
import ar.com.tulocalweb.app.pantallas.compartir.HojaHistoria
import ar.com.tulocalweb.app.pantallas.compartir.TarjetaHistoria
import ar.com.tulocalweb.app.pantallas.ingreso.*
import ar.com.tulocalweb.app.pantallas.inicio.Inicio
import ar.com.tulocalweb.app.pantallas.mas.*
import ar.com.tulocalweb.app.pantallas.pedidos.DetallePedido
import ar.com.tulocalweb.app.pantallas.pedidos.ListaPedidos
import ar.com.tulocalweb.app.pantallas.pedidos.PedidoManual
import ar.com.tulocalweb.app.pantallas.productos.EditorProducto
import ar.com.tulocalweb.app.pantallas.productos.ListaProductos
import ar.com.tulocalweb.app.pantallas.tincho.RespuestaTincho
import ar.com.tulocalweb.app.pantallas.tincho.TinchoChat
import ar.com.tulocalweb.app.pantallas.tincho.VistaBloque
import ar.com.tulocalweb.app.pantallas.tincho.corto
import ar.com.tulocalweb.app.pantallas.venta.VentaLocal
import ar.com.tulocalweb.app.tema.*
import com.github.takahirom.roborazzi.captureRoboImage
import okhttp3.Interceptor
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Protocol
import okhttp3.Response
import okhttp3.ResponseBody.Companion.toResponseBody
import kotlinx.serialization.json.JsonElement
import kotlinx.serialization.json.jsonObject
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Rule
import org.junit.Test
import org.junit.runner.RunWith
import org.robolectric.RobolectricTestRunner
import org.robolectric.annotation.Config
import org.robolectric.annotation.GraphicsMode

/** La API simulada (ApiFalsa en iOS): responde por ruta con pruebas/respuestas.json, el mismo archivo que lee iOS.
 *  Lo que no está: GET → 404, lo demás → 200 con `{}`. */
class ApiFalsa(private val tabla: Map<String, String?>) : Interceptor {
    override fun intercept(chain: Interceptor.Chain): Response {
        val req = chain.request()
        val ruta = req.url.encodedPath
        val cuerpo = tabla["${req.method} $ruta"] ?: if (req.method == "GET") tabla[ruta] else "{}"
        val codigo = if (cuerpo == null) 404 else 200
        return Response.Builder().request(req).protocol(Protocol.HTTP_1_1).code(codigo).message(if (codigo == 200) "OK" else "No existe")
            .body((cuerpo ?: "{}").toResponseBody("application/json".toMediaType())).build()
    }

    companion object {
        /** Las respuestas compartidas con iOS (apps/pruebas/respuestas.json): «MÉTODO ruta» o sólo la ruta. */
        fun respuestas(): Map<String, String> {
            val raiz = Api.json.parseToJsonElement(ApiFalsa::class.java.getResource("/pruebas/respuestas.json")!!.readText()).jsonObject
            return raiz.filterKeys { !it.startsWith("_") }.mapValues { Api.json.encodeToString(JsonElement.serializer(), it.value) }
        }
    }
}

/** Clon de TuLocalWebTests/PantallasTests.swift: monta cada pantalla contra la API simulada, a 390 × 844 dp como la
 *  ventana de iOS, y guarda la captura en claro y oscuro (build/capturas/<Nombre>-claro|oscuro.png). Falla si alguna
 *  pantalla se rompe al montarse. Al sumar una pantalla, sumarla acá (y en iOS). */
@RunWith(RobolectricTestRunner::class)
@GraphicsMode(GraphicsMode.Mode.NATIVE)
@Config(sdk = [35], qualifiers = "w390dp-h844dp-xxhdpi")
class PantallasTest {
    @get:Rule val compose = createComposeRule()

    companion object {
        val tienda = Acceso.Tienda("t1", "Almacén Lumbre", "almacen-lumbre", "owner", "https://api.prueba/logo.png")
        val plan: String get() = ApiFalsa.respuestas()["/v1/panel/cuenta/plan"] ?: "{}"
        /** Una respuesta de Tincho con todos los bloques (y uno que no existe, que se ignora). */
        val respuestaTincho = """
        {"texto": "Hoy llevás **$ 539.000** en 15 pedidos.", "bloques": [
          {"tipo": "cifras", "items": [{"titulo": "Ventas", "valor": "$ 539.000", "detalle": "▲ 53 %", "tono": "ok"}, {"titulo": "Pedidos", "valor": "15"}]},
          {"tipo": "barras", "titulo": "Más vendidos", "unidad": "unidades", "items": [{"etiqueta": "Yerba", "valor": 208, "destacado": true}, {"etiqueta": "Mate", "valor": 120}]},
          {"tipo": "linea", "titulo": "Ventas por día", "unidad": "pesos", "puntos": [{"etiqueta": "lun 26", "valor": 500000}, {"etiqueta": "mar 27", "valor": 776400}, {"etiqueta": "mié 28", "valor": 344700}], "anterior": [400000, 600000, 300000]},
          {"tipo": "tabla", "titulo": "Detalle", "columnas": ["Producto", "Stock"], "filas": [["Yerba", "17"], ["Kit", "Sin stock"]]},
          {"tipo": "lista", "titulo": "Para reponer", "items": [{"titulo": "Kit matero", "detalle": "Se vende", "insignia": "Sin stock", "tono": "mal"}]},
          {"tipo": "progreso", "titulo": "Meta del mes", "actual": 11011500, "objetivo": 21600000, "unidad": "pesos", "detalle": "Proyección"},
          {"tipo": "nuevo-que-no-existe"}
        ], "acciones": [{"texto": "Ver pedidos", "ir": "pedidos"}], "sugerencias": ["¿Y ayer?"]}
        """
    }

    private fun estado(extra: Map<String, String?> = emptyMap()): Estado {
        Preferencias.borrarTodo()
        val api = Api("https://api.prueba", "x", OkHttpClient.Builder().addInterceptor(ApiFalsa(ApiFalsa.respuestas() + extra)).build())
        val e = Estado(api)
        e.contexto = RuntimeEnvironment.getApplication()
        e.acceso = Acceso("t1", "owner", listOf("catalogo", "ventas"), listOf(tienda))
        e.datos = DatosSesion(DatosSesion.Usuario("u1", "Martina Gómez", "martina@almacenlumbre.com.ar"), "t1", "owner", java.time.Instant.parse("2026-12-31T00:00:00Z"))
        // Tincho (la web, /panel/hornero) contesta con la misma API simulada.
        Web.clientePrueba = OkHttpClient.Builder().addInterceptor(ApiFalsa(ApiFalsa.respuestas() + extra)).build()
        return e
    }

    /** Deja correr la carga: las respuestas llegan por el hilo de OkHttp y vuelven al principal. */
    private fun asentar(ms: Long) {
        repeat(12) {
            Thread.sleep(60)
            compose.mainClock.advanceTimeBy(ms / 12)
            compose.waitForIdle()
        }
    }

    /** Monta la vista con el Estado y una Navegacion (NavigationStack en iOS) y captura claro y oscuro. */
    private fun montar(nombre: String, e: Estado = estado(), espera: Long = 1500, vista: @Composable () -> Unit) {
        val tema = mutableStateOf("claro")
        val nav = Navegacion()
        compose.setContent {
            TemaTLW(tema.value) { CompositionLocalProvider(LocalEstado provides e, LocalNavegacion provides nav) { vista() } }
        }
        for (t in listOf("claro", "oscuro")) {
            tema.value = t
            asentar(espera)
            compose.onRoot().captureRoboImage("build/capturas/$nombre-$t.png")
        }
    }

    /** Piezas sueltas (gráficos, bloques de Tincho, el editor de estilo): con fondo, margen y aire entre una y otra,
     *  como quedan adentro de una pantalla. */
    @Composable private fun EnHoja(contenido: @Composable ColumnScope.() -> Unit) =
        Column(Modifier.fillMaxSize().background(C.fondo).verticalScroll(rememberScrollState()).padding(20.dp), verticalArrangement = Arrangement.spacedBy(16.dp)) { contenido() }

    /** Lo que en la app va en una Hoja: a pantalla completa, con el fondo de la hoja. */
    @Composable private fun ComoHoja(contenido: @Composable () -> Unit) = Box(Modifier.fillMaxSize().background(C.elevado).statusBarsPadding()) { contenido() }

    @Test fun compartirTienda() = montar("CompartirTienda") { CompartirTienda() }
    @Test fun compartirTiendaSinProductos() = montar("CompartirTiendaSinProductos",
        estado(mapOf("/v1/panel/listas/productos" to """{"total": 0, "segmentos": {}, "filas": []}""", "/v1/panel/cupones" to null))) { CompartirTienda() }
    // Todas las pantallas de la app con la API simulada (lo que no está, 404): un crash al abrirlas salta acá.
    @Test fun pantallaInicio() = montar("Inicio") { Inicio() }
    @Test fun pantallaListaPedidos() = montar("ListaPedidos", espera = 3000) { ListaPedidos() }
    @Test fun pantallaListaProductos() = montar("ListaProductos") { ListaProductos() }
    @Test fun pantallaTinchoChat() = montar("TinchoChat", estado(mapOf("POST /panel/hornero" to respuestaTincho)), espera = 2500) { TinchoChat() }
    @Test fun pantallaMas() = montar("Mas") { Mas() }
    @Test fun pantallaDetallePedido() = montar("DetallePedido") { DetallePedido(1048) }
    @Test fun pantallaAvisos() = montar("Avisos") { Avisos() }
    @Test fun pantallaCuenta() = montar("Cuenta") { Cuenta() }
    @Test fun pantallaClientes() = montar("Clientes") { Clientes() }
    @Test fun pantallaCajaDelDia() = montar("CajaDelDia") { CajaDelDia() }
    @Test fun pantallaCarritosAbandonados() = montar("CarritosAbandonados") { CarritosAbandonados() }
    @Test fun pantallaResenasYPreguntas() = montar("ResenasYPreguntas") { ResenasYPreguntas() }
    @Test fun pantallaDevoluciones() = montar("Devoluciones") { Devoluciones() }
    @Test fun pantallaCupones() = montar("Cupones") { Cupones() }
    @Test fun pantallaFinanzas() = montar("Finanzas") { Finanzas() }
    @Test fun pantallaFacturacionArca() = montar("FacturacionArca") { FacturacionArca() }
    @Test fun pantallaPlanCuenta() = montar("PlanCuenta") { PlanCuenta() }

    /** El plan con prorrateo (§56): la respuesta de la API se decodifica y la hoja con el resumen se arma. */
    @Test fun hojaPlanConResumen() {
        val d = Api.json.decodeFromString(DatosPlan.serializer(), plan)
        assertEquals(14516, d.cotizacion("pro", "mes")?.hoy)
        assertEquals("gratis", d.cotizacion("gratis", "anio")?.tipo) // el gratis se cotiza con el ciclo vigente
        assertTrue(ar.com.tulocalweb.app.nucleo.Formato.diaMes("2026-11-20T00:00:00Z").isNotEmpty())
        // El descuento para quedarse (§57): mitad del débito mientras queden cobros.
        assertEquals(9950, DatosPlan.conDescuento(19900, d.beneficio))
        assertEquals(19900, DatosPlan.conDescuento(19900, null))
        montar("HojaPlanPro") { ComoHoja { HojaPlan(d, "pro") {} } }
    }
    @Test fun hojaPlanGratis() = montar("HojaPlanGratis") { ComoHoja { HojaPlan(Api.json.decodeFromString(DatosPlan.serializer(), plan), "gratis") {} } }
    @Test fun hojaCancelar() = montar("HojaCancelar") { ComoHoja { HojaCancelar(Api.json.decodeFromString(DatosPlan.serializer(), plan)) {} } }

    @Test fun graficos() {
        val flujo = listOf(DiaFlujo(1, 100, 30), DiaFlujo(2, 0, 80))
        montar("Graficos", espera = 1000) {
            EnHoja {
                GraficoVentas(listOf(0.0, 3.0, 2.0, 5.0), listOf(1.0, 1.0, 2.0, 2.0))
                GraficoFlujo(flujo)
                GraficoSaldo(flujo)
                DonaMedios(listOf(ParteDona("Efectivo", 3), ParteDona("Mercado Pago", 7)))
                GraficoMeses(listOf("sep", "oct", "nov"), listOf(SerieMeses("ventas", "Ventas", listOf(1, 2, null), { "$it" }, { "$it" })))
                EmbudoVisitas(listOf(Tablero.Visitas.Pagina("inicio", 100), Tablero.Visitas.Pagina("checkout", 4)))
                BarrasEstrellas(listOf(0, 1, 0, 3, 9))
                AnilloAvance(0.62)
            }
        }
    }

    @Test fun pantallaUsuariosYPermisos() = montar("UsuariosYPermisos") { UsuariosYPermisos() }
    @Test fun pantallaEditorProductoExistente() = montar("EditorProductoExistente") { EditorProducto("p1") }
    @Test fun pantallaPedidoManual() = montar("PedidoManual") { PedidoManual() }
    @Test fun pantallaPreferenciasAvisos() = montar("PreferenciasAvisos") { PreferenciasAvisos() }
    @Test fun pantallaNegocioCobro() = montar("NegocioCobro") { NegocioCobro() }
    @Test fun pantallaLinksPago() = montar("LinksPago") { LinksPago() }
    @Test fun pantallaCentroAyuda() = montar("CentroAyuda") { CentroAyuda() }
    @Test fun pantallaMiTiendaEditor() = montar("MiTiendaEditor") { MiTiendaEditor() }
    @Test fun onboarding() = montar("Onboarding") { Onboarding() }
    @Test fun crearTienda() = montar("CrearTienda") { ComoHoja { CrearTienda() } }
    @Test fun importarTienda() = montar("ImportarTienda") { ImportarTienda() }
    @Test fun telegram() = montar("Telegram") { TelegramTincho() }
    @Test fun medicion() = montar("Medicion") { MedicionTienda() }
    @Test fun iaConectadas() = montar("IAConectadas") { IAConectadas() }

    @Test fun bloquesDeTincho() {

        val r = Api.json.decodeFromString(RespuestaTincho.serializer(), respuestaTincho)
        assertEquals(7, r.bloques?.size)
        assertEquals("$ 1,2 M", corto(1_250_000.0, "pesos"))
        montar("BloquesDeTincho", espera = 1000) { EnHoja { (r.bloques ?: emptyList()).forEach { VistaBloque(it) } } }
    }

    @Test fun estadisticas() = montar("Estadisticas") { Estadisticas() }
    @Test fun integraciones() = montar("Integraciones") { Integraciones() }
    @Test fun hojaHistoria() = montar("HojaHistoria") { ComoHoja { HojaHistoria(tienda) } }
    @Test fun editorProductoNuevoPreguntaElTipo() = montar("EditorProductoNuevo") { EditorProducto(null) }

    @Test fun tipoDeLosProductosDeAntes() {
        assertEquals(TipoProducto.ropa, TipoProducto.adivinar(mapOf("Talle" to listOf("S", "M"))))
        assertEquals(TipoProducto.calzado, TipoProducto.adivinar(mapOf("Talle" to listOf("38", "39"), "Color" to listOf("Negro"))))
        assertNull(TipoProducto.adivinar(mapOf("Sabor" to listOf("Limón"))))
        assertTrue(TipoProducto.hogar.campos.contains(AtributoProducto.alto))
    }

    @Test fun tarjetaHistoria() {
        assertNotNull(TarjetaHistoria.dibujar(RuntimeEnvironment.getApplication(), tienda, null, null))
    }

    // Sólo en Android (para la galería): lo que iOS no monta en sus pruebas.
    @Test fun principal() = montar("Principal") { Principal() }
    @Test fun splash() = montar("Splash", espera = 1000) { Splash() }
    @Test fun ingresar() = montar("Ingresar", espera = 1000) { Ingresar() }
    @Test fun elegirTienda() = montar("ElegirTienda", estado().also { it.acceso = Acceso("t1", "owner", tiendas = listOf(tienda, tienda.copy(id = "t2", nombre = "Lumbre Palermo", slug = "lumbre-palermo"))) }) { ComoHoja { ElegirTienda() } }
    @Test fun permisos() = montar("Permisos", espera = 1000) { Permisos() }
    @Test fun bloqueo() = montar("Bloqueo", espera = 1000) { Bloqueo() }
    @Test fun buscar() = montar("Buscar", espera = 1000) { Buscar() }
    @Test fun ventaLocal() = montar("VentaLocal") { VentaLocal() }
    @Test fun estiloTienda() = montar("EstiloTienda", espera = 1000) { EnHoja { EditorEstilo(EstiloTienda(), {}, 0, {}, "litoral", listOf("Mates")) } }
}
