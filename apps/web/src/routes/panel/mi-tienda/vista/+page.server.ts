import { tiendaDelBorrador } from '#lib/server/compartir/borrador.ts';
import { paginaVista } from '#lib/panel/mi-tienda/vista.ts';
import { migrar } from '#lib/tienda/sistema/migrar.ts';
import type { PageServerLoad } from './$types';

/**
 * La tienda de verdad con el borrador de Mi tienda, para la vista previa del editor (un iframe) y
 * la de las apps (con Bearer). `pagina` elige qué se ve; con `marcos=1` cada bloque se puede tocar
 * para abrir su panel. Después, el editor manda los cambios por postMessage sin recargar.
 */
export const load: PageServerLoad = async ({ locals, cookies, setHeaders, url }) => {
	setHeaders({ 'cache-control': 'private, no-store' });
	const { mi, tienda } = await tiendaDelBorrador(locals, cookies);
	return {
		tienda,
		diseno: migrar(mi.diseno, mi.plantillaId),
		plantilla: mi.plantillaId,
		pagina: paginaVista(url.searchParams.get('pagina')),
		marcos: url.searchParams.get('marcos') === '1',
		// La ficha se muestra con el producto más vendido (o el primero).
		producto:
			tienda.productos.find((p) => p.id === tienda.masVendidos[0]) ?? tienda.productos[0] ?? null
	};
};
