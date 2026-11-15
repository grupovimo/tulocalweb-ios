import type { Cookies } from '@sveltejs/kit';

/**
 * De dónde viene quien se crea la cuenta (/empezar: QR de flyers, roll-up, one-pager, links de
 * partners; /l/: los anuncios). Va en una cookie de 90 días y se guarda en la tienda al llegar al onboarding.
 */
export const COOKIE_ORIGEN = 'mostrador-origen';

export type Origen = {
	fuente?: string;
	medio?: string;
	campana?: string;
	/** utm_content: el anuncio puntual (en Meta, su id). */
	contenido?: string;
	ref?: string;
	cupon?: string;
	llego?: string;
};

const valido = (v: string | null) => (v && /^[A-Za-z0-9_.-]{1,40}$/.test(v) ? v : undefined);

export function origenDe(url: URL, ahora = new Date()): Origen | null {
	const q = url.searchParams;
	const o: Origen = {
		fuente: valido(q.get('utm_source')),
		medio: valido(q.get('utm_medium')),
		campana: valido(q.get('utm_campaign')),
		contenido: valido(q.get('utm_content')),
		ref: valido(q.get('ref')),
		cupon: valido(q.get('cupon'))?.toUpperCase()
	};
	const limpio = Object.fromEntries(Object.entries(o).filter(([, v]) => v)) as Origen;
	if (!Object.keys(limpio).length) return null;
	return { ...limpio, llego: ahora.toISOString() };
}

export function guardarOrigen(cookies: Cookies, o: Origen) {
	cookies.set(COOKIE_ORIGEN, JSON.stringify(o), {
		path: '/',
		httpOnly: true,
		sameSite: 'lax',
		maxAge: 90 * 86_400
	});
}

export function leerOrigen(cookies: Cookies): Origen | null {
	try {
		const o = JSON.parse(cookies.get(COOKIE_ORIGEN) ?? 'null') as unknown;
		if (!o || typeof o !== 'object' || Array.isArray(o)) return null;
		// Sólo los campos conocidos y válidos; sin ninguno de atribución, no hay origen.
		const r = o as Record<string, unknown>;
		const campo = (k: string) => (typeof r[k] === 'string' ? valido(r[k] as string) : undefined);
		const limpio = Object.fromEntries(
			(['fuente', 'medio', 'campana', 'contenido', 'ref', 'cupon'] as const)
				.map((k) => [k, campo(k)])
				.filter(([, v]) => v)
		) as Origen;
		if (!Object.keys(limpio).length) return null;
		const llego = typeof r.llego === 'string' && r.llego.length <= 40 ? r.llego : undefined;
		return llego ? { ...limpio, llego } : limpio;
	} catch {
		return null;
	}
}
