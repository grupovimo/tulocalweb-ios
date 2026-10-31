import { error, json } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import { PlanSuscripcion } from '#lib/server/suscripciones-esquemas.ts';
import { cuerpo } from '#lib/server/turnos-esquemas.ts';
import type { RequestHandler } from './$types';

export const POST: RequestHandler = async ({ locals, request }) => {
	const r = await datosDe(locals).guardarPlanSuscripcion(
		null,
		await cuerpo(request, PlanSuscripcion)
	);
	if (r === 'invalido' || !r) error(400, 'Revisá los productos del plan.');
	return json(r, { status: 201 });
};
