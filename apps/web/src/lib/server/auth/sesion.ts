import type { Cookies } from '@sveltejs/kit';
import { dev } from '$app/env';

/**
 * Cookie de sesión del panel. `__Host-` en producción: sólo HTTPS, sin Domain y con Path=/, así no
 * la puede pisar un subdominio. En desarrollo (http://localhost) va sin el prefijo.
 */
export const COOKIE_SESION = dev ? 'tlw_sesion' : '__Host-tlw_sesion';
/** El nombre de antes del cambio de marca: se sigue aceptando hasta que venza, así nadie se desloguea. */
export const COOKIE_SESION_VIEJA = dev ? 'mostrador_sesion' : '__Host-mostrador_sesion';

export function guardarSesion(cookies: Cookies, token: string, venceEn: string) {
	cookies.delete(COOKIE_SESION_VIEJA, { path: '/', secure: !dev });
	cookies.set(COOKIE_SESION, token, {
		path: '/',
		httpOnly: true,
		secure: !dev,
		sameSite: 'lax',
		expires: new Date(venceEn)
	});
}

/** El desafío de la verificación en dos pasos, entre la contraseña y el código (10 minutos). */
export const COOKIE_DESAFIO = dev ? 'tlw_desafio' : '__Host-tlw_desafio';

/**
 * Lo que termina cada ingreso: guarda la sesión o, si la cuenta tiene la verificación en dos pasos,
 * el desafío y a dónde seguir. Devuelve true cuando falta el código: hay que ir a
 * /ingresar?m=codigo (el redirect lo hace quien llama, fuera de su try).
 */
export function guardarIngreso(
	cookies: Cookies,
	s: { token: string; venceEn: string; segundoFactor?: string },
	hacia: string
): boolean {
	if (!s.segundoFactor) {
		guardarSesion(cookies, s.token, s.venceEn);
		return false;
	}
	cookies.set(COOKIE_DESAFIO, JSON.stringify({ d: s.segundoFactor, h: hacia }), {
		path: '/',
		httpOnly: true,
		secure: !dev,
		sameSite: 'lax',
		expires: new Date(s.venceEn)
	});
	return true;
}

/** El desafío pendiente y a dónde seguir (null si no hay o venció). */
export function leerDesafio(cookies: Cookies): { d: string; h: string } | null {
	try {
		const v = JSON.parse(cookies.get(COOKIE_DESAFIO) ?? '');
		if (typeof v?.d !== 'string' || !v.d) return null;
		// Sólo destinos de este sitio.
		const h = typeof v.h === 'string' && /^\/(?!\/)/.test(v.h) ? v.h : '/panel';
		return { d: v.d, h };
	} catch {
		return null;
	}
}

export function borrarDesafio(cookies: Cookies) {
	cookies.delete(COOKIE_DESAFIO, { path: '/', secure: !dev });
}

export function borrarSesion(cookies: Cookies) {
	cookies.delete(COOKIE_SESION, { path: '/', secure: !dev });
	cookies.delete(COOKIE_SESION_VIEJA, { path: '/', secure: !dev });
}
