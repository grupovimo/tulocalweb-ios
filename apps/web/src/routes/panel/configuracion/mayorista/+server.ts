import { error } from '@sveltejs/kit';
import * as v from 'valibot';
import { datosDe } from '#lib/server/datos/index.ts';
import { MAX_MONTO, MAX_UNIDADES, problemaMayorista } from '#lib/tienda/mayorista.ts';
import type { RequestHandler } from './$types';

const Minimo = (max: number) =>
	v.nullable(v.pipe(v.number(), v.integer(), v.minValue(1), v.maxValue(max)));
const Mayorista = v.object({
	activa: v.boolean(),
	unidades: Minimo(MAX_UNIDADES),
	monto: Minimo(MAX_MONTO)
});

/** Venta mayorista (§49): prendida o apagada y el mínimo del carrito (Go y el sandbox lo validan). */
export const PUT: RequestHandler = async ({ locals, request }) => {
	const r = v.safeParse(Mayorista, await request.json().catch(() => null));
	if (!r.success) error(400, 'Revisá el mínimo: de 2 a 10.000 unidades y hasta $ 100.000.000.');
	const mal = problemaMayorista(r.output);
	if (mal) error(400, mal);
	await datosDe(locals).guardarMayorista(r.output);
	return new Response(null, { status: 204 });
};
