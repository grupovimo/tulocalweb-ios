import { error } from '@sveltejs/kit';
import * as v from 'valibot';
import { datosDe } from '#lib/server/datos/index.ts';
import { CUOTAS, MAX_DESDE, problemaCobros } from '#lib/tienda/cobros.ts';
import type { RequestHandler } from './$types';

const Cobros = v.object({
	cuotas: v.picklist(CUOTAS),
	cuotasDesde: v.pipe(v.number(), v.integer(), v.minValue(0), v.maxValue(MAX_DESDE)),
	impuestos: v.object({ mostrar: v.boolean() })
});

/** Cobros: cuotas sin interés (cuántas y desde cuánto) y el precio sin impuestos (Go y el sandbox lo validan). */
export const PUT: RequestHandler = async ({ locals, request }) => {
	const r = v.safeParse(Cobros, await request.json().catch(() => null));
	if (!r.success)
		error(400, 'Revisá las cuotas (ninguna, 3, 6, 9 o 12) y el monto: hasta $ 100.000.000.');
	const mal = problemaCobros(r.output);
	if (mal) error(400, mal);
	await datosDe(locals).guardarCobros(r.output);
	return new Response(null, { status: 204 });
};
