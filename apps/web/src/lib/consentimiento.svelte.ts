/**
 * Elección de cookies del visitante, como en el handoff: localStorage['tlw-cookies'] guarda
 * 'all', 'nec' o un JSON {an, mk}. Analíticas (GA4) y marketing (Meta Pixel) sólo se cargan si están
 * permitidas; Cloudflare Web Analytics no usa cookies y va siempre.
 */
export type Eleccion = { an: boolean; mk: boolean };

const CLAVE = 'tlw-cookies';
/** La clave de antes del cambio de marca: se pasa a la nueva la primera vez que se lee. */
const CLAVE_VIEJA = 'mostrador-cookies';

/** Si el banner está a la vista: el widget de Tincho espera a que se cierre para no taparlo. */
export const banner = $state({ abierto: false });

export function leerEleccion(): Eleccion | null {
	let v: string | null;
	try {
		v = localStorage.getItem(CLAVE);
		const vieja = v ? null : localStorage.getItem(CLAVE_VIEJA);
		if (vieja) {
			localStorage.setItem(CLAVE, vieja);
			localStorage.removeItem(CLAVE_VIEJA);
			v = vieja;
		}
	} catch {
		return null;
	}
	if (!v) return null;
	if (v === 'all') return { an: true, mk: true };
	if (v === 'nec') return { an: false, mk: false };
	try {
		const o = JSON.parse(v) as Partial<Eleccion>;
		return { an: !!o.an, mk: !!o.mk };
	} catch {
		return null;
	}
}

/**
 * La cookie que lee el servidor al dar de alta la cuenta: con 'si', el alta va además a la API de
 * conversiones de Meta (#lib/server/meta.ts). Refleja el permiso de marketing.
 */
export const COOKIE_MARKETING = 'tlw-marketing';

export function guardarEleccion(v: 'all' | 'nec' | Eleccion) {
	try {
		localStorage.setItem(CLAVE, typeof v === 'string' ? v : JSON.stringify(v));
	} catch {
		// Sin almacenamiento (modo privado): la elección vale para esta visita.
	}
	const mk = v === 'all' || (typeof v === 'object' && v.mk);
	try {
		document.cookie = `${COOKIE_MARKETING}=${mk ? 'si' : 'no'}; path=/; max-age=31536000; samesite=lax; secure`;
	} catch {
		// Sin cookies: el alta se mide sólo desde el navegador.
	}
	window.dispatchEvent(new CustomEvent('mostrador:cookies'));
}
