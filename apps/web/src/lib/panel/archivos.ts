/**
 * Subir un archivo para descargar (el premio «descarga» de una captación): PDF, EPUB o ZIP de hasta
 * 20 MB, tal cual. El servidor mira el tipo real por los primeros bytes (lib/server/archivos.ts).
 */
export const MAXIMO_ARCHIVO = 20 * 1024 * 1024;
export const ACEPTA_ARCHIVO =
	'.pdf,.epub,.zip,application/pdf,application/epub+zip,application/zip';

/** El nombre que ve quien lo baja: sin la extensión, hasta 60 letras. */
export function nombreDeArchivo(nombre: string): string {
	return nombre
		.replace(/\.(pdf|epub|zip)$/i, '')
		.replace(/[_-]+/g, ' ')
		.replace(/\s+/g, ' ')
		.trim()
		.slice(0, 60);
}

/** Sube el archivo y devuelve su URL pública. Los errores vienen con el mensaje para mostrar. */
export async function subirArchivo(archivo: Blob): Promise<string> {
	if (archivo.size > MAXIMO_ARCHIVO) throw new Error('El archivo pesa más de 20 MB.');
	const r = await fetch('/panel/archivos', {
		method: 'POST',
		headers: { 'content-type': 'application/octet-stream' },
		body: archivo
	});
	if (!r.ok)
		throw new Error(
			((await r.json().catch(() => null)) as { message?: string } | null)?.message ??
				'No pudimos subir el archivo.'
		);
	return new URL(((await r.json()) as { ruta: string }).ruta, location.origin).href;
}
