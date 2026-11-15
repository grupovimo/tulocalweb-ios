import { env } from 'cloudflare:workers';
import { error, json } from '@sveltejs/kit';
import { LIMITES_DOCUMENTO, subirDocumento } from '#lib/server/archivos.ts';
import { datosDe, usaSandbox } from '#lib/server/datos/index.ts';
import { golpe } from '#lib/server/limite.ts';
import { dbSandbox } from '#lib/server/sandbox/sandbox.ts';
import type { RequestHandler } from './$types';

/**
 * Subir un archivo para descargar (el premio «descarga» de una captación): el cuerpo es el PDF,
 * EPUB o ZIP tal cual. El tipo se mira por los primeros bytes, no por lo que dice quien sube.
 * Devuelve su ruta en /f/ (con el origen, una URL https estable para `premio.archivo`).
 */
export const POST: RequestHandler = async (event) => {
	datosDe(event.locals); // sin sesión, a ingresar
	const largo = Number(event.request.headers.get('content-length') ?? 0);
	if (largo > LIMITES_DOCUMENTO.archivo) error(413, 'El archivo pesa más de 20 MB.');
	const bytes = new Uint8Array(await event.request.arrayBuffer());
	const db = dbSandbox(env.SANDBOX);
	const tienda = (await datosDe(event.locals).tienda()).id;
	if (!(await golpe(db, `archivos:${tienda}`, 20, 3600)))
		error(429, 'Subiste muchos archivos seguidos. Probá en un rato.');
	const sandbox = usaSandbox() ? (event.locals.sandbox?.id ?? null) : null;
	const clave = await subirDocumento(db, tienda, bytes, sandbox);
	return json({ ruta: `/f/${clave}` }, { status: 201 });
};
