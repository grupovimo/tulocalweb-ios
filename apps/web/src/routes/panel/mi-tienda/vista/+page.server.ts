import { tiendaDelBorrador } from '#lib/server/compartir/borrador.ts';
import type { PageServerLoad } from './$types';

/**
 * La tienda de verdad con el borrador de Mi tienda (colores, tipografías, logos, secciones, textos),
 * para la vista previa del editor (un iframe que se recarga en cada guardado) y la de la app.
 */
export const load: PageServerLoad = async ({ locals, cookies, setHeaders }) => {
	setHeaders({ 'cache-control': 'private, no-store' });
	const { tienda } = await tiendaDelBorrador(locals, cookies);
	return { tienda };
};
