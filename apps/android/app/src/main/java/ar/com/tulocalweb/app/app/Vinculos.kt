package ar.com.tulocalweb.app.app

import android.net.Uri

/** Los links de tulocalweb.com.ar que abren la app (App Links) y las rutas del panel que mandan los avisos (clon de
 *  App/Vinculos.swift). Qué paths abren la app lo declara `apps/pruebas/enlaces.json` (de ahí sale el assetlinks de la
 *  web y el manifest lo repite); acá, a qué pantalla lleva cada uno. Lo que la app no tiene se abre en la web dentro
 *  de la app, nunca queda en nada. */
object Vinculos {
    fun ruta(u: Uri): Ruta? {
        val host = u.host ?: return null
        if (host != "tulocalweb.com.ar" && host != "www.tulocalweb.com.ar") return null
        val texto = u.toString()
        fun valor(k: String): String? = if (k in u.queryParameterNames) u.getQueryParameter(k).orEmpty() else null
        if (valor("web") != null) return Ruta.web(texto)
        val p = u.pathSegments
        if (p.firstOrNull() == "ayuda") return Ruta.destino(Destino.ayuda)
        if (p.firstOrNull() != "panel") return Ruta.web(texto)
        val seccion = p.getOrNull(1)
        val id = p.getOrNull(2)
        return when (seccion) {
            null, "para-hacer-hoy" -> Ruta.inicio
            "pedidos" -> {
                val n = valor("pedido")?.toIntOrNull() ?: id?.toIntOrNull()
                when {
                    n != null -> Ruta.pedido(n)
                    id == "nuevo" -> Ruta.destino(Destino.pedidoManual)
                    id == null -> Ruta.pedidos
                    else -> Ruta.web(texto)
                }
            }
            "productos" -> (valor("editar") ?: id)?.let { Ruta.producto(it) } ?: Ruta.productos
            "configuracion" -> Ruta.destino(if (valor("tab") == "seguridad") Destino.cuenta else Destino.negocio)
            else -> secciones.firstOrNull { it.first == seccion }?.let { Ruta.destino(it.second) } ?: Ruta.web(texto)
        }
    }

    /** Las secciones del panel que tienen pantalla en la app (/panel/<sección>). */
    private val secciones = listOf(
        "inventario" to Destino.inventario, "ventas-en-local" to Destino.caja, "clientes" to Destino.clientes,
        "carritos-abandonados" to Destino.carritos, "resenas-y-preguntas" to Destino.resenas, "devoluciones" to Destino.devoluciones,
        "cupones" to Destino.cupones, "finanzas" to Destino.finanzas, "facturacion" to Destino.facturacion, "plan-y-facturacion" to Destino.plan,
        "integraciones" to Destino.integraciones, "usuarios-y-permisos" to Destino.usuarios, "mis-datos" to Destino.cuenta,
        "notificaciones" to Destino.preferenciasAvisos, "informes" to Destino.estadisticas, "mi-tienda" to Destino.miTienda,
        "links" to Destino.links, "importar" to Destino.importar,
    )
}

/** Cómo se escribe la ruta en `apps/pruebas/enlaces.json` («pedido:1048», «destino:cupones», «web-en-app»). */
val Ruta.clave: String get() = when (this) {
    is Ruta.pedido -> "pedido:$n"
    is Ruta.producto -> "producto:$id"
    Ruta.avisos -> "avisos"
    Ruta.pedidos -> "pedidos"
    Ruta.inicio -> "inicio"
    Ruta.compartir -> "compartir"
    Ruta.productos -> "productos"
    is Ruta.web -> "web-en-app"
    is Ruta.destino -> "destino:" + d.toString()
}
