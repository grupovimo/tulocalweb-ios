import { filasCsv } from '#lib/panel/newsletter.ts';
import { datosDe } from '#lib/server/datos/index.ts';
import { csv } from '#lib/server/zip.ts';
import type { RequestHandler } from './$types';

/** Los suscriptos en CSV (con BOM y punto y coma: así lo abre bien Excel en castellano). */
export const GET: RequestHandler = async ({ locals }) => {
	const n = await datosDe(locals).newsletter();
	const dia = new Date(Date.now() - 3 * 3_600_000).toISOString().slice(0, 10);
	return new Response(csv(filasCsv(n.suscriptores), 'excel'), {
		headers: {
			'content-type': 'text/csv; charset=utf-8',
			'content-disposition': `attachment; filename="newsletter-${dia}.csv"`,
			'cache-control': 'no-store'
		}
	});
};
