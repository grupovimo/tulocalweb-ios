import { error } from '@sveltejs/kit';
import { clienteDe } from '#lib/server/datos/index.ts';
import { COOKIE } from '#lib/server/sandbox/sandbox.ts';
import { bajaCarrito } from '#lib/server/tienda/carritos.ts';
import { resolverTienda } from '#lib/server/tienda/publica.ts';
import type { Actions, PageServerLoad } from './$types';

/**
 * «No quiero más estos mails» de los recordatorios del carrito (§62). Es también la URL de
 * List-Unsubscribe: el POST de un clic del servidor de correo (RFC 8058) llega a esta acción sin
 * Origin (exceptuada del CSRF: el token es la credencial) y da de baja sin más pasos.
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
		await bajaCarrito(r, clienteDe(event), params.token);
		return { listo: true };
	}
};
