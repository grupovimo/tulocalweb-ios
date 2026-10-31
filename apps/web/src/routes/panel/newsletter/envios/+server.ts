import { error, json } from '@sveltejs/kit';
import { validarNovedad } from '#lib/panel/newsletter.ts';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** «Mandar novedad»: a los suscriptos confirmados, de a tandas y con el tope diario (en Go). */
export const POST: RequestHandler = async ({ locals, request }) => {
	const v = validarNovedad(await request.json().catch(() => null));
	if (!v.ok) error(400, v.error);
	const e = await datosDe(locals).mandarNovedad(v.novedad);
	if (e === 'sin-suscriptos')
		error(409, 'Todavía no hay suscriptos confirmados: la novedad no tendría a quién llegar.');
	if ('error' in e) error(400, e.error);
	return json(e, { status: 201 });
};
