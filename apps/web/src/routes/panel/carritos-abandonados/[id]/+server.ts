import { error } from '@sveltejs/kit';
import type { RecordatoriosCarritos } from '#lib/panel/recordatorios.ts';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** Los tres recordatorios por mail (§62): /panel/carritos-abandonados/recordatorios. */
export const PUT: RequestHandler = async ({ locals, params, request }) => {
	if (params.id !== 'recordatorios') error(404, 'No existe.');
	const r = (await request.json().catch(() => null)) as RecordatoriosCarritos | null;
	if (!r || typeof r !== 'object' || !Array.isArray(r.pasos))
		error(400, 'Revisá los recordatorios.');
	const mal = await datosDe(locals).guardarRecordatorios({
		activo: r.activo === true,
		pasos: r.pasos.slice(0, 3).map((p, i) => ({
			activo: p?.activo === true,
			horas: Number(p?.horas),
			cupon:
				i > 0 && p?.cupon
					? {
							descuento: p.cupon.descuento,
							valor: Number(p.cupon.valor) || 0,
							dias: Number(p.cupon.dias) || 3
						}
					: null
		}))
	});
	if (mal) error(400, mal);
	return new Response(null, { status: 204 });
};

/** "Recordar" un carrito (id), "todos" o guardar los automáticos ("automaticos"). */
export const POST: RequestHandler = async ({ locals, params, request }) => {
	const datos = datosDe(locals);
	if (params.id === 'todos') return Response.json({ enviados: await datos.recordarCarritos() });
	if (params.id === 'automaticos') {
		const a = (await request.json().catch(() => null)) as {
			hora?: unknown;
			dia?: unknown;
			email?: unknown;
		} | null;
		await datos.guardarAutomaticos({
			hora: a?.hora === true,
			dia: a?.dia === true,
			email: a?.email === true
		});
		return new Response(null, { status: 204 });
	}
	const r = await datos.recordarCarrito(params.id);
	if (r === 'no-existe') error(404, 'Ese carrito no existe.');
	if (r === 'no-se-puede') error(409, 'Ya le mandaste el recordatorio.');
	return new Response(null, { status: 204 });
};
