package ar.com.tulocalweb.app.nucleo

import android.content.ClipData
import android.content.ClipboardManager
import android.content.Context
import android.content.Intent
import android.net.Uri
import androidx.browser.customtabs.CustomTabsIntent
import androidx.core.content.FileProvider
import java.io.File

// Lo que en iOS es openURL, ShareLink, UIPasteboard y SFSafariViewController.

/** Abre una página en la hoja web del sistema (Custom Tab; SFSafariViewController / openURL de https en iOS). */
fun abrirEnNavegador(ctx: Context, url: String) {
    runCatching {
        CustomTabsIntent.Builder().setShowTitle(true).build().apply { intent.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK) }.launchUrl(ctx, Uri.parse(url))
    }.onFailure { abrirUrl(ctx, url) }
}

/** openURL de un link cualquiera (tel:, mailto:, wa.me, la tienda en el navegador). Llamar abre el marcador. */
fun abrirUrl(ctx: Context, url: String) {
    val u = Uri.parse(url)
    val i = if (u.scheme == "tel") Intent(Intent.ACTION_DIAL, u) else Intent(Intent.ACTION_VIEW, u)
    runCatching { ctx.startActivity(i.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)) }
}

/** ShareLink de un texto (o un link). */
fun compartirTexto(ctx: Context, texto: String, asunto: String? = null) {
    val i = Intent(Intent.ACTION_SEND).setType("text/plain").putExtra(Intent.EXTRA_TEXT, texto)
    asunto?.let { i.putExtra(Intent.EXTRA_SUBJECT, it) }
    runCatching { ctx.startActivity(Intent.createChooser(i, null).addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)) }
}

/** Un archivo para compartir, en la caché (cache/compartir, que expone el FileProvider). */
fun archivoParaCompartir(ctx: Context, nombre: String): File = File(ctx.cacheDir, "compartir").apply { mkdirs() }.let { File(it, nombre) }

/** ShareLink de un archivo (la imagen para historias, el pedido .csr de ARCA, la etiqueta en PDF). */
fun compartirArchivo(ctx: Context, archivo: File, tipo: String, texto: String? = null) {
    val uri = FileProvider.getUriForFile(ctx, "ar.com.tulocalweb.app.archivos", archivo)
    val i = Intent(Intent.ACTION_SEND).setType(tipo).putExtra(Intent.EXTRA_STREAM, uri).addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION)
    texto?.let { i.putExtra(Intent.EXTRA_TEXT, it) }
    runCatching { ctx.startActivity(Intent.createChooser(i, null).addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)) }
}

/** UIPasteboard.general.string = … */
fun copiar(ctx: Context, texto: String) {
    ctx.getSystemService(ClipboardManager::class.java)?.setPrimaryClip(ClipData.newPlainText("Tu Local Web", texto))
}
