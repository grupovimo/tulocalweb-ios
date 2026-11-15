import { error } from '@sveltejs/kit';
import * as v from 'valibot';
import { clienteDe } from '#lib/server/datos/index.ts';
import { COOKIE } from '#lib/server/sandbox/sandbox.ts';
import { registrarEventos } from '#lib/server/tienda/promos.ts';
import { resolverTienda } from '#lib/server/tienda/publica.ts';
import type { RequestHandler } from './$types';

const Lote = v.object({
	eventos: v.pipe(
		v.array(
			v.object({
				ventana: v.pipe(v.string(), v.regex(/^[a-z0-9]{6,16}$/)),
				tipo: v.picklist(['vista', 'cierre', 'clic']),
				n: v.optional(v.pipe(v.number(), v.integer(), v.minValue(1), v.maxValue(20)))
			})
		),
		v.maxLength(50)
	)
});

/**
 * Vistas, cierres y clics de las ventanas, en lote con navigator.sendBeacon (llega como text/plain).
 * Responde 204 si el lote tiene forma: a la tienda no le sirve saber si se guardó.
 */
export const POST: RequestHandler = async (event) => {
	const { params, cookies, request, platform } = event;
	const texto = await request.text();
	if (texto.length > 8192) error(413, 'Demasiado largo.');
	let cuerpo: unknown = null;
	try {
		cuerpo = JSON.parse(texto);
	} catch {
		// queda null y falla la validación
	}
	const b = v.safeParse(Lote, cuerpo);
	if (!b.success) error(400, 'Mandá hasta 50 eventos.');
	const r = await resolverTienda(params.tienda, cookies.get(COOKIE));
	if (!r) error(404, 'Esta tienda no existe.');
	const lote = registrarEventos(r, clienteDe(event), b.output.eventos);
	// No hace falta esperar la escritura para contestar.
	if (platform?.ctx) platform.ctx.waitUntil(lote);
	else await lote;
	return new Response(null, { status: 204 });
};
