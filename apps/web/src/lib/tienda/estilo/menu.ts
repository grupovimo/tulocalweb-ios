import type { TiendaPublica } from '../tipos.ts';

/**
 * Mi tienda › Menú: las pestañas que arma la tienda (texto y adónde lleva). Sin pestañas propias,
 * va el menú de la plantilla. `a` es una página de la tienda, una categoría (`cat:<nombre>`) o un
 * link propio (https://…).
 */
export type ItemMenu = { t: string; a: string };

export const MAX_MENU = 8;
export const PAGINAS = ['inicio', 'catalogo', 'nosotros', 'turnos', 'suscripciones'] as const;
export const NOMBRE_PAGINA: Record<(typeof PAGINAS)[number], string> = {
	inicio: 'Inicio',
	catalogo: 'Catálogo',
	nosotros: 'Nosotros',
	turnos: 'Turnos',
	suscripciones: 'Suscripciones'
};

/** El link de una pestaña (null si el destino no es válido). */
export function hrefDe(a: string, base: string): string | null {
	if ((PAGINAS as readonly string[]).includes(a))
		return a === 'inicio' ? base || '/' : `${base}/${a}`;
	if (a.startsWith('cat:') && a.length > 4)
		return `${base}/catalogo?${new URLSearchParams({ cat: a.slice(4) })}`;
	if (/^https:\/\/[^\s"'<>]+$/.test(a)) return a;
	return null;
}

/** El menú propio de la tienda como lo usan las plantillas ({t, href, on}), o null. */
export function menuPropio(
	tienda: Pick<TiendaPublica, 'estilo'>,
	base: string
): { t: string; href: string; on: boolean }[] | null {
	const items = (tienda.estilo?.menu ?? [])
		.slice(0, MAX_MENU)
		.map((i) => ({ t: i.t.trim(), href: hrefDe(i.a, base) ?? '', on: false }))
		.filter((i) => i.t && i.href);
	return items.length ? items : null;
}
