/**
 * Dominios propios de las tiendas (mitienda.com.ar). Todo host que no es de Tu Local Web es el de
 * una tienda: el hook `reroute` (src/hooks.ts) lleva sus paths a /t/<slug> y el servidor no deja
 * ver nada más en ese dominio (ni el panel ni la landing).
 */
export function esDominioPropio(host: string): boolean {
	const h = host.toLowerCase();
	return !(
		h === 'localhost' ||
		h === '127.0.0.1' ||
		h === 'tulocalweb.com.ar' ||
		h.endsWith('.tulocalweb.com.ar') ||
		h.endsWith('.workers.dev')
	);
}

/** Hosts de la plataforma que no son tiendas. */
const NO_TIENDA = new Set([
	'www',
	'api',
	'webmail',
	'mail',
	'correo',
	'panel',
	'app',
	'staff',
	'admin',
	'static',
	'cdn'
]);

/**
 * La tienda de un subdominio de la plataforma: <slug>.tulocalweb.com.ar (docs/decisiones.md §55).
 * null si el host no es un subdominio de tienda.
 */
export function slugDeSubdominio(host: string): string | null {
	const m = /^([a-z0-9][a-z0-9-]{0,58}[a-z0-9])\.tulocalweb\.com\.ar$/.exec(host.toLowerCase());
	if (!m || NO_TIENDA.has(m[1])) return null;
	return m[1];
}

/** Paths que en un dominio propio se sirven tal cual, sin pasarlos a /t/<slug>. */
export const GLOBALES_EN_DOMINIO = /^\/(api\/dominio|legales)(\/|$)/;

/** El path interno (/t/<slug>/…) de un path pedido en el dominio de la tienda `slug`. */
export function pathInterno(slug: string, path: string): string | undefined {
	if (GLOBALES_EN_DOMINIO.test(path)) return;
	const base = `/t/${slug}`;
	if (path === base || path.startsWith(`${base}/`)) return;
	return path === '/' ? base : base + path;
}
