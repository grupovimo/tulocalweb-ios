import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';

/** La pestaña «Suscripciones» del menú: todos los planes. Sin planes, no existe. */
export const load: PageLoad = async ({ parent }) => {
	const { tienda } = await parent();
	if (!tienda.suscripciones.length) error(404, 'Esta tienda no tiene suscripciones.');
	return {};
};
