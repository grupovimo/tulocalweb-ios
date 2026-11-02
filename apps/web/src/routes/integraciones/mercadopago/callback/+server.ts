import { redirect } from '@sveltejs/kit';
import { datosDe, usaSandbox } from '#lib/server/datos/index.ts';
import { leerVuelta } from '#lib/server/volver.ts';
import type { RequestHandler } from './$types';

/**
 * Vuelta del OAuth de Mercado Pago (la URL registrada en la aplicación). El state de un solo uso y el
 * canje del código los hace la API; acá sólo se pasa y se vuelve al panel con el resultado.
 */
export const GET: RequestHandler = async ({ locals, url, cookies }) => {
	const codigo = url.searchParams.get('code') ?? '';
	const estado = url.searchParams.get('state') ?? '';
	// Sin sesión de la web: la conexión la arrancó la app de iPhone (que no comparte la cookie). Se le
	// devuelve el código a la app, que lo canjea con su token (ConectarCuenta en la app).
	if (!locals.token && !usaSandbox() && codigo && estado)
		redirect(
			303,
			`ar.com.tulocalweb.app://integraciones/mercadopago?code=${encodeURIComponent(codigo)}&state=${encodeURIComponent(estado)}`,
			// Un esquema de la app: SvelteKit 3 sólo deja salir del sitio a propósito.
			{ external: true }
		);
	// La persona canceló en Mercado Pago, o falta algo.
	const r =
		url.searchParams.has('error') || !codigo || !estado
			? 'no-se-puede'
			: await datosDe(locals).volverMercadoPago(codigo.slice(0, 512), estado.slice(0, 512));
	redirect(303, leerVuelta(cookies, 'mercadopago', r === 'ok'));
};
