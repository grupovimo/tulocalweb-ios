import type { Reroute } from '@sveltejs/kit/hooks';
import {
	esDominioPropio,
	GLOBALES_EN_DOMINIO,
	pathInterno,
	slugDeSubdominio
} from '#lib/tienda/dominio.ts';

/** En el navegador el dominio no cambia: se pregunta una vez. */
let slugNavegador: Promise<string | null> | undefined;

/**
 * Un dominio propio (mitienda.com.ar) muestra su tienda: /producto/mate se resuelve como
 * /t/<slug>/producto/mate. La tienda del dominio la dice /api/dominio (cacheada).
 */
export const reroute: Reroute = async ({ url, fetch }) => {
	// <slug>.tulocalweb.com.ar: la tienda sale del host, sin preguntar.
	const sub = slugDeSubdominio(url.hostname);
	if (sub) return pathInterno(sub, url.pathname);
	if (!esDominioPropio(url.hostname) || GLOBALES_EN_DOMINIO.test(url.pathname)) return;
	const pedir = () =>
		fetch(new URL('/api/dominio', url))
			.then((r) => (r.ok ? (r.json() as Promise<{ slug: string | null }>) : { slug: null }))
			.then((r) => r.slug);
	const slug = await (typeof window === 'undefined' ? pedir() : (slugNavegador ??= pedir()));
	return slug ? pathInterno(slug, url.pathname) : undefined;
};
