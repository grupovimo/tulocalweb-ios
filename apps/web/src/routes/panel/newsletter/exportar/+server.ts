import { filasCsv } from '#lib/panel/newsletter.ts';
import { datosDe } from '#lib/server/datos/index.ts';
import { csv } from '#lib/server/zip.ts';
import type { RequestHandler } from './$types';

/** Los suscriptos en CSV (con BOM y punto y coma: así lo abre bien Excel en castellano). */
export const GET: RequestHandler = async ({ locals }) => {
	const datos = datosDe(locals);
	const [n, mi] = await Promise.all([datos.newsletter(), datos.miTienda().catch(() => null)]);
	// El origen de los captados por una ventana va con su nombre.
	const d = mi?.diseno as { ventanas?: { id: string; nombre: string }[] } | null | undefined;
	const ventanas = Object.fromEntries((d?.ventanas ?? []).map((w) => [w.id, w.nombre]));
	const dia = new Date(Date.now() - 3 * 3_600_000).toISOString().slice(0, 10);
	return new Response(csv(filasCsv(n.suscriptores, ventanas), 'excel'), {
		headers: {
			'content-type': 'text/csv; charset=utf-8',
			'content-disposition': `attachment; filename="newsletter-${dia}.csv"`,
			'cache-control': 'no-store'
		}
	});
};
