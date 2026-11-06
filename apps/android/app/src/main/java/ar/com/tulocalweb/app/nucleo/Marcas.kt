package ar.com.tulocalweb.app.nucleo

import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color
import ar.com.tulocalweb.app.tema.C
import ar.com.tulocalweb.app.tema.ChipEstado
import ar.com.tulocalweb.app.tema.hex

/** Nombres y colores de canales, pagos y envíos (como apps/web/src/lib/marcas.ts), clon de Nucleo/Marcas.swift. */
object Marcas {
    /** Como `capitalized` de Swift: cada palabra con mayúscula. */
    private fun capitalizado(s: String) = s.split(" ").joinToString(" ") { w -> w.lowercase().replaceFirstChar { it.titlecase() } }

    /** Corto, para la meta de una fila: «Web», «ML», «Local». */
    fun canalCorto(c: String): String = when (c) {
        "web" -> "Web"
        "ml" -> "ML"
        "pos" -> "Local"
        "wa" -> "WhatsApp"
        "ig" -> "Instagram"
        "tg" -> "Telegram"
        else -> capitalizado(c)
    }

    /** Largo: «Tienda web», «Mercado Libre». */
    fun canal(c: String): String = when (c) {
        "web" -> "Tienda web"
        "ml" -> "Mercado Libre"
        "pos" -> "Local"
        else -> canalCorto(c)
    }

    @Composable
    fun colorCanal(c: String): Color = when (c) {
        "web" -> C.canalWeb
        "ml" -> C.canalML
        "pos" -> C.canalLocal
        "wa" -> hex(0x2F8F63)
        "ig" -> hex(0xC24E25)
        else -> C.texto3
    }

    fun pago(p: String): String = when (p) {
        "mp" -> "Mercado Pago"
        "tarjeta" -> "Tarjeta"
        "payway" -> "Payway"
        "modo" -> "MODO"
        "uala" -> "Ualá"
        "transf" -> "Transferencia"
        "efectivo" -> "Efectivo"
        else -> capitalizado(p)
    }

    fun envio(e: String): String = when (e) {
        "andreani" -> "Andreani"
        "oca" -> "OCA"
        "correo" -> "Correo Argentino"
        "moto" -> "Envío propio"
        "retiro" -> "Retiro en el local"
        else -> capitalizado(e)
    }

    data class Siguiente(val estado: String, val boton: String, val corto: String)

    /** El siguiente estado de un pedido y el texto del botón (la barra de acción de 3b y el swipe de 3a). null si ya
     *  terminó. */
    fun siguiente(estado: String, envio: String): Siguiente? {
        val retiro = envio == "retiro"
        return when (estado) {
            "Sin pagar" -> Siguiente("Para preparar", "Marcar como cobrado", "Cobrado")
            "Para preparar" -> Siguiente("Listo para despachar", if (retiro) "Listo para retirar" else "Listo para despachar", "Listo")
            "Listo para despachar" -> if (retiro) Siguiente("Entregado", "Marcar como retirado", "Retirado") else Siguiente("En camino", "Marcar como despachado", "Despachado")
            "En camino" -> Siguiente("Entregado", "Marcar como entregado", "Entregado")
            else -> null
        }
    }

    /** El estado de antes (para «Deshacer»). */
    fun anterior(estado: String, envio: String): String? {
        val circuito = if (envio == "retiro") listOf("Sin pagar", "Para preparar", "Listo para despachar", "Entregado")
        else listOf("Sin pagar", "Para preparar", "Listo para despachar", "En camino", "Entregado")
        val i = circuito.indexOf(estado)
        return if (i > 0) circuito[i - 1] else null
    }

    /** Cómo queda dicho el cambio en el toast: «#1043 listo para retirar». */
    fun dicho(estado: String, envio: String): String = ChipEstado.texto(estado, envio).lowercase()
}
