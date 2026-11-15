import { error, json } from '@sveltejs/kit';
import { clienteDe } from '#lib/server/datos/index.ts';
import { COOKIE } from '#lib/server/sandbox/sandbox.ts';
import { recientes } from '#lib/server/tienda/promos.ts';
import { resolverTienda } from '#lib/server/tienda/publica.ts';
import type { RequestHandler } from './$types';

/**
 * La prueba social: { compras: [{ nombre, provincia, producto, productoId, foto, cuando }] }, sólo
 * pedidos reales pagados de los últimos 7 días (vacío si la tienda no la prendió). 5 minutos de caché.
 */
export const GET: RequestHandler = async (event) => {
	const r = await resolverTienda(event.params.tienda, event.cookies.get(COOKIE));
	if (!r) error(404, 'Esta tienda no existe.');
	const compras = await recientes(r, clienteDe(event));
	// La demo es de cada visitante: no va a una caché compartida.
	return json(
		{ compras },
		{ headers: { 'cache-control': r.sandboxId ? 'private, max-age=300' : 'public, max-age=300' } }
	);
};
