import { env } from 'cloudflare:workers';
import { error } from '@sveltejs/kit';
import { CLAVE, CLAVE_DOCUMENTO } from '#lib/server/archivos.ts';
import type { RequestHandler } from './$types';

/**
 * Las fotos de productos y los archivos para descargar (R2; en el VPS, una carpeta en disco). La
 * clave lleva un uuid: no cambia nunca, se cachea un año.
 */
export const GET: RequestHandler = async ({ params, request }) => {
	const documento = CLAVE_DOCUMENTO.test(params.clave);
	if (!documento && !CLAVE.test(params.clave)) error(404, 'No existe.');
	const o = await env.FILES.get(params.clave, { onlyIf: request.headers });
	if (!o) error(404, 'No existe.');
	const h = new Headers();
	o.writeHttpMetadata(h);
	h.set('etag', o.httpEtag);
	h.set('cache-control', 'public, max-age=31536000, immutable');
	h.set('x-content-type-options', 'nosniff');
	if (documento) {
		// Un PDF se abre en el visor del navegador; EPUB y ZIP se bajan. Nada de lo que traiga
		// adentro corre con el origen de la tienda.
		h.set('content-security-policy', 'sandbox');
		if (!params.clave.endsWith('.pdf')) h.set('content-disposition', 'attachment');
	}
	// Sin cuerpo: se cumplió el If-None-Match.
	if (!('body' in o)) return new Response(null, { status: 304, headers: h });
	return new Response(o.body, { headers: h });
};
