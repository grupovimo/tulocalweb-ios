import { error, json } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import { PlanSuscripcion } from '#lib/server/suscripciones-esquemas.ts';
import { cuerpo } from '#lib/server/turnos-esquemas.ts';
import type { RequestHandler } from './$types';

export const PUT: RequestHandler = async ({ locals, params, request }) => {
	const r = await datosDe(locals).guardarPlanSuscripcion(
		params.id,
		await cuerpo(request, PlanSuscripcion)
	);
	if (r === 'invalido') error(400, 'Revisá los productos del plan.');
	if (!r) error(404, 'Ese plan ya no existe.');
	return json(r);
};

export const DELETE: RequestHandler = async ({ locals, params }) => {
	const r = await datosDe(locals).borrarPlanSuscripcion(params.id);
	if (r === 'no-existe') error(404, 'Ese plan ya no existe.');
	if (r === 'no-se-puede') error(409, 'Este plan tiene suscriptores: pausalo en vez de borrarlo.');
	return new Response(null, { status: 204 });
};
