import { error, json } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

export const POST: RequestHandler = async ({ locals, params }) => {
	const r = await datosDe(locals).cancelarSuscriptor(params.id);
	if (r === 'no-existe') error(404, 'Esa suscripción ya no existe.');
	if (r === 'mercadopago')
		error(
			502,
			'Mercado Pago no respondió: la suscripción sigue activa. Probá de nuevo en un rato.'
		);
	return json(r);
};
