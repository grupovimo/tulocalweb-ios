import { env } from 'cloudflare:workers';
import { error } from '@sveltejs/kit';
import { angulo } from '#lib/anuncios/angulos.ts';
import { guardarOrigen, leerOrigen, origenDe } from '#lib/server/origen.ts';
import type { PageServerLoad } from './$types';

/**
 * Las páginas de los anuncios (/l/avisos, /l/una-caja, /l/pasate). Al llegar se anota el origen
 * (los utm del anuncio; sin ellos, la página misma), como en /empezar: el primero manda.
 */
export const load: PageServerLoad = ({ params, url, cookies }) => {
	const a = angulo(params.angulo);
	if (!a) error(404, 'No encontramos esta página.');
	if (!leerOrigen(cookies))
		guardarOrigen(
			cookies,
			origenDe(url) ?? { medio: 'aterrizaje', campana: a.id, llego: new Date().toISOString() }
		);
	return { a, pixel: env.META_PIXEL_ID || null };
};
