import { error, json } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

const FECHA = /^\d{4}-\d{2}-\d{2}$/;

/**
 * Mi tienda › Promociones: los números de las ventanas (vistas, cierres, clics, envíos, ventas y
 * monto) por ventana y por día, entre `desde` y `hasta` (AAAA-MM-DD; vacío: los últimos 30 días).
 */
export const GET: RequestHandler = async ({ locals, url }) => {
	const desde = url.searchParams.get('desde');
	const hasta = url.searchParams.get('hasta');
	if ((desde && !FECHA.test(desde)) || (hasta && !FECHA.test(hasta)))
		error(400, 'Elegí las fechas como AAAA-MM-DD.');
	const m = await datosDe(locals).metricasPromos(desde, hasta);
	if (!m) error(400, 'Elegí un rango de fechas de hasta un año.');
	return json(m, { headers: { 'cache-control': 'private, no-store' } });
};
