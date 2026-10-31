import { error, fail } from '@sveltejs/kit';
import { clienteDe } from '#lib/server/datos/index.ts';
import { COOKIE } from '#lib/server/sandbox/sandbox.ts';
import { accionToken } from '#lib/server/tienda/newsletter.ts';
import { resolverTienda } from '#lib/server/tienda/publica.ts';
import type { Actions, PageServerLoad } from './$types';

/**
 * La baja del newsletter (§50). Es la URL de List-Unsubscribe de cada mail: el POST de un clic de
 * Gmail o Yahoo (List-Unsubscribe=One-Click, RFC 8058) llega a esta misma acción, sin Origin (está
 * exceptuada del CSRF: el token es la credencial) y da de baja sin más pasos.
 */
export const load: PageServerLoad = async ({ params, cookies }) => {
	const r = await resolverTienda(params.tienda, cookies.get(COOKIE));
	if (!r) error(404, 'Esta tienda no existe.');
	return { tienda: r.tienda, token: params.token };
};

export const actions: Actions = {
	default: async (event) => {
		const { params, cookies } = event;
		const r = await resolverTienda(params.tienda, cookies.get(COOKIE));
		if (!r) error(404, 'Esta tienda no existe.');
		if (!(await accionToken(r, clienteDe(event), 'baja', params.token)))
			return fail(404, { invalido: true });
		return { listo: true };
	}
};
