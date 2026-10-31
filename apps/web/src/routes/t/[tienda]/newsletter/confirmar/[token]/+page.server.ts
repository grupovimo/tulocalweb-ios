import { error, fail } from '@sveltejs/kit';
import { clienteDe } from '#lib/server/datos/index.ts';
import { COOKIE } from '#lib/server/sandbox/sandbox.ts';
import { accionToken } from '#lib/server/tienda/newsletter.ts';
import { resolverTienda } from '#lib/server/tienda/publica.ts';
import type { Actions, PageServerLoad } from './$types';

/**
 * El link del mail de confirmación (§50). Confirma con un botón y no al abrir: los antivirus de
 * correo abren los links solos y no tienen que suscribir a nadie.
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
		if (!(await accionToken(r, clienteDe(event), 'confirmar', params.token)))
			return fail(404, { invalido: true });
		return { listo: true };
	}
};
