import { error, json } from '@sveltejs/kit';
import { ErrorApi } from '#lib/server/bff.ts';
import { cuentasDe, usaSandbox } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/**
 * Verificación en dos pasos de la persona (docs/decisiones.md §58), contra /v1/auth/segundo-factor.
 * En la demo no hay cuentas: se muestra apagada y sin poder prenderse.
 */
export const GET: RequestHandler = async (event) => {
	if (usaSandbox() || !event.locals.token)
		return json({ activo: false, desde: null, codigosRestantes: 0, disponible: false });
	try {
		const e = await cuentasDe(event).estadoSegundoFactor(event.locals.token);
		return json({ ...e, disponible: true });
	} catch (e) {
		if (!(e instanceof ErrorApi)) throw e;
		error(e.estado, e.message);
	}
};

/** {accion: 'activar'} da el QR; confirmar, codigos y desactivar llevan el código. */
export const POST: RequestHandler = async (event) => {
	if (usaSandbox() || !event.locals.token) error(409, 'En la demo no se puede prender.');
	const { accion, codigo } = (await event.request.json().catch(() => ({}))) as {
		accion?: string;
		codigo?: string;
	};
	const c = cuentasDe(event);
	try {
		if (accion === 'activar') return json(await c.activarSegundoFactor(event.locals.token));
		if (accion === 'confirmar' || accion === 'codigos' || accion === 'desactivar') {
			const r = await c.conCodigoSegundoFactor(accion, String(codigo ?? ''), event.locals.token);
			return json(r ?? {});
		}
	} catch (e) {
		if (!(e instanceof ErrorApi)) throw e;
		error(e.estado, e.message);
	}
	error(400, 'Acción desconocida.');
};
