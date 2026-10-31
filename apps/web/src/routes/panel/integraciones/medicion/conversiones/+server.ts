import { error, type RequestEvent } from '@sveltejs/kit';
import * as v from 'valibot';
import { ErrorApi } from '#lib/server/bff.ts';
import { conversionesDe } from '#lib/server/conversiones.ts';
import { usaSandbox } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** Los mismos límites que Go (conversiones.Guardar). */
const Conexion = v.object({
	token: v.pipe(
		v.string(),
		v.trim(),
		v.regex(/^([A-Za-z0-9_-]{20,500})?$/, 'Ese no parece un token de Meta: copialo entero.')
	),
	codigoPrueba: v.pipe(
		v.string(),
		v.trim(),
		v.toUpperCase(),
		v.regex(/^(TEST[0-9A-Z]{1,20})?$/, 'El código de prueba empieza con TEST.')
	)
});

function sesion(event: RequestEvent) {
	if (usaSandbox())
		error(409, 'Es una tienda de ejemplo: la API de conversiones se conecta con tu píxel real.');
	if (!event.locals.token) error(401, 'Tu sesión venció. Volvé a ingresar.');
	return event.locals.token;
}

function falla(e: unknown): never {
	if (e instanceof ErrorApi) error(e.estado, e.message);
	console.error('conversiones', e instanceof Error ? e.message : e);
	error(503, 'No pudimos guardar. Probá de nuevo en un rato.');
}

/** Conecta (Go comprueba el token con Meta) o cambia el código de prueba. */
export const PUT: RequestHandler = async (event) => {
	const s = sesion(event);
	const b = v.safeParse(Conexion, await event.request.json().catch(() => null));
	if (!b.success) error(400, b.issues[0]?.message ?? 'Revisá los datos.');
	await conversionesDe(event).guardar(s, b.output.token, b.output.codigoPrueba).catch(falla);
	return new Response(null, { status: 204 });
};

export const DELETE: RequestHandler = async (event) => {
	await conversionesDe(event).desconectar(sesion(event)).catch(falla);
	return new Response(null, { status: 204 });
};
