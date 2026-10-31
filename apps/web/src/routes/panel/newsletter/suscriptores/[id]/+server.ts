import { error } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** Sacar a alguien de la lista del newsletter. */
export const DELETE: RequestHandler = async ({ locals, params }) => {
	const r = await datosDe(locals).borrarSuscriptor(params.id);
	if (r === 'no-existe') error(404, 'Ese suscripto ya no está en la lista.');
	return new Response(null, { status: 204 });
};
