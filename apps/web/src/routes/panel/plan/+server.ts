import { error, json } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import { ErrorCotizar } from '#lib/panel/cotizar.ts';
import type { RequestHandler } from './$types';

const PLANES = ['gratis', 'emprende', 'pro', 'negocio'] as const;

/**
 * Cambiar de plan (y de ciclo) con su cotización (§56). Cuando hay que pasar por Mercado Pago (cargar la
 * tarjeta del débito o pagar la diferencia prorrateada) devuelve { url }; si quedó hecho o programado,
 * 204. Con un cambio programado, sólo se puede deshacerlo: 409.
 */
export const PUT: RequestHandler = async ({ locals, request }) => {
	const c = (await request.json().catch(() => null)) as { plan?: unknown; ciclo?: unknown } | null;
	const plan = PLANES.find((p) => p === c?.plan);
	if (!plan) error(400, 'Elegí un plan.');
	const ciclo = c?.ciclo === 'mes' || c?.ciclo === 'anio' ? c.ciclo : undefined;
	const r = await datosDe(locals)
		.cambiarPlan(plan, ciclo)
		.catch((e: unknown) => {
			if (e instanceof ErrorCotizar && e.motivo === 'programado')
				error(409, 'Ya tenés un cambio de plan programado: deshacelo para elegir otro.');
			throw e;
		});
	if (r?.url) {
		if (!/^https:\/\/(www\.)?mercadopago\.com(\.ar)?\//.test(r.url))
			error(502, 'Respuesta inesperada.');
		return json({ url: r.url });
	}
	return new Response(null, { status: 204 });
};
