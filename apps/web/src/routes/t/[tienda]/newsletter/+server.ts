import { error, json } from '@sveltejs/kit';
import * as v from 'valibot';
import { clienteDe } from '#lib/server/datos/index.ts';
import { COOKIE } from '#lib/server/sandbox/sandbox.ts';
import { suscribir } from '#lib/server/tienda/newsletter.ts';
import { resolverTienda } from '#lib/server/tienda/publica.ts';
import type { RequestHandler } from './$types';

const Alta = v.object({
	email: v.pipe(v.string(), v.trim(), v.toLowerCase(), v.email(), v.maxLength(120))
});

/**
 * El bloque de newsletter del pie de la tienda (§50): queda pendiente hasta que confirme por mail.
 * En la demo no sale el mail: la respuesta trae el link de confirmar.
 */
export const POST: RequestHandler = async (event) => {
	const { params, cookies, request } = event;
	const texto = await request.text();
	if (texto.length > 2048) error(413, 'Demasiado largo.');
	let cuerpo: unknown = null;
	try {
		cuerpo = JSON.parse(texto);
	} catch {
		// queda null y falla la validación
	}
	const b = v.safeParse(Alta, cuerpo);
	if (!b.success) return json({ error: 'Revisá el email.' }, { status: 400 });
	const r = await resolverTienda(params.tienda, cookies.get(COOKIE));
	if (!r) error(404, 'Esta tienda no existe.');
	const s = await suscribir(r, clienteDe(event), b.output.email);
	if (!s.ok) return json({ error: s.error }, { status: s.estado });
	return json({ demo: s.demo ?? null }, { status: 202 });
};
