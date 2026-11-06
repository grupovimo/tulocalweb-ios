import { redirect } from '@sveltejs/kit';
import { ErrorApi } from '#lib/server/bff.ts';
import { guardarIngreso } from '#lib/server/auth/sesion.ts';
import { esProveedor } from '#lib/server/auth/social.ts';
import { cuentasDe, usaSandbox } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/**
 * Vuelta del proveedor (redirect_uri de Google, Mercado Pago y Mercado Libre). Canjea en la API,
 * guarda la sesión y sigue: una cuenta nueva va al onboarding; una que ya existía, al panel.
 */
export const GET: RequestHandler = async (event) => {
	const { params, url } = event;
	const codigo = url.searchParams.get('code') ?? '';
	const estado = url.searchParams.get('state') ?? '';
	if (usaSandbox() || !esProveedor(params.proveedor)) redirect(303, '/ingresar');
	// La persona canceló en el proveedor o este devolvió un error.
	if (!codigo || !estado) redirect(303, '/ingresar?error=cancelado');
	let destino = '/panel';
	try {
		const r = await cuentasDe(event).volverOAuth(params.proveedor, codigo, estado);
		if (r.nueva) destino = '/onboarding?paso=1';
		if (guardarIngreso(event.cookies, r, destino)) destino = '/ingresar?m=codigo';
	} catch (e) {
		if (!(e instanceof ErrorApi)) throw e;
		redirect(303, `/ingresar?error=${encodeURIComponent(e.codigo)}`);
	}
	redirect(303, destino);
};
