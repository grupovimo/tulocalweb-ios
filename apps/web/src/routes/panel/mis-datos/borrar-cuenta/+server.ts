import { error } from '@sveltejs/kit';
import { cuentasDe, usaSandbox } from '#lib/server/datos/index.ts';
import { borrarSesion } from '#lib/server/auth/sesion.ts';
import { ErrorCuenta } from '#lib/server/auth/cuentas.ts';
import type { RequestHandler } from './$types';

/** Borrar la cuenta de la persona (con su email como confirmación). Google Play pide esta URL. */
export const POST: RequestHandler = async (event) => {
	const token = event.locals.token;
	if (usaSandbox() || !token) error(400, 'En la demo no hay una cuenta para borrar.');
	const c = (await event.request.json().catch(() => null)) as { confirmacion?: unknown } | null;
	try {
		await cuentasDe(event).borrarCuenta(token, String(c?.confirmacion ?? ''));
	} catch (e) {
		if (e instanceof ErrorCuenta && e.estado < 500) error(400, e.message);
		throw e;
	}
	borrarSesion(event.cookies);
	return new Response(null, { status: 204 });
};
