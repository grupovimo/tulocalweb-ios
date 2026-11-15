import { env } from 'cloudflare:workers';
import type { Cookies } from '@sveltejs/kit';
import { COOKIE_MARKETING } from '#lib/consentimiento.svelte.ts';
import type { Origen } from './origen.ts';

/**
 * El alta de una cuenta que vino de un anuncio, a la API de conversiones de Meta
 * (CompleteRegistration), con el píxel de Tu Local Web. Va desde el servidor porque el navegador de
 * Instagram y los bloqueadores se comen buena parte del píxel. Sólo si quien se registra aceptó las
 * cookies de marketing y si el píxel y el token están cargados; nunca frena el onboarding.
 */
const VERSION = 'v23.0';

export type Alta = {
	tiendaId: string;
	email: string;
	origen: Origen;
	ip: string;
	agente: string;
	fbp?: string;
	fbc?: string;
	url: string;
	hora: number;
};

async function sha256(s: string) {
	const b = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(s));
	return [...new Uint8Array(b)].map((x) => x.toString(16).padStart(2, '0')).join('');
}

/** El cuerpo del evento, sin red (para probarlo). El email va normalizado y hasheado, como pide Meta. */
export async function eventoAlta(a: Alta) {
	const usuario: Record<string, unknown> = { em: [await sha256(a.email.trim().toLowerCase())] };
	for (const [k, v] of [
		['client_ip_address', a.ip],
		['client_user_agent', a.agente],
		['fbp', a.fbp],
		['fbc', a.fbc]
	] as const)
		if (v) usuario[k] = v;
	return {
		event_name: 'CompleteRegistration',
		event_time: a.hora,
		// Una sola vez por tienda aunque se reintente.
		event_id: `alta-${a.tiendaId}`,
		action_source: 'website',
		event_source_url: a.url,
		user_data: usuario,
		custom_data: {
			status: 'registrada',
			...(a.origen.campana ? { content_name: a.origen.campana } : {})
		}
	};
}

export type Config = { pixel: string; token: string; prueba?: string };

/** El píxel y el token de Tu Local Web; null si falta alguno. */
export function config(): Config | null {
	const pixel = String(env.META_PIXEL_ID ?? '');
	const token = String(env.META_CAPI_TOKEN ?? '');
	const prueba = env.META_CODIGO_PRUEBA ? String(env.META_CODIGO_PRUEBA) : undefined;
	return pixel && token ? { pixel, token, prueba } : null;
}

/** Si el alta se puede medir: hay píxel y token, y la persona aceptó el marketing. */
export const medible = (cookies: Pick<Cookies, 'get'>, c = config()) =>
	!!c && cookies.get(COOKIE_MARKETING) === 'si';

export async function mandarAlta(a: Alta, c = config(), f: typeof fetch = fetch) {
	if (!c) return;
	const cuerpo: Record<string, unknown> = { data: [await eventoAlta(a)], access_token: c.token };
	if (c.prueba) cuerpo.test_event_code = c.prueba;
	const r = await f(`https://graph.facebook.com/${VERSION}/${c.pixel}/events`, {
		method: 'POST',
		headers: { 'content-type': 'application/json' },
		body: JSON.stringify(cuerpo),
		signal: AbortSignal.timeout(8000)
	}).catch(() => null);
	// Sin el detalle: la respuesta de Meta puede repetir datos del evento.
	if (!r?.ok) console.warn('meta: el alta no se mandó', r?.status ?? 'sin respuesta');
}
