import { error, json, type RequestEvent } from '@sveltejs/kit';
import * as v from 'valibot';
import { clienteDe } from '#lib/server/datos/index.ts';
import { COOKIE } from '#lib/server/sandbox/sandbox.ts';
import { participar } from '#lib/server/tienda/promos.ts';
import { resolverTienda } from '#lib/server/tienda/publica.ts';

const Pedido = v.object({
	ventana: v.pipe(v.string(), v.regex(/^[a-z0-9]{6,16}$/)),
	email: v.pipe(v.string(), v.trim(), v.toLowerCase(), v.email(), v.maxLength(120)),
	nombre: v.optional(v.pipe(v.string(), v.trim(), v.maxLength(60))),
	whatsapp: v.optional(v.pipe(v.string(), v.trim(), v.maxLength(24))),
	consentimiento: v.optional(v.boolean(), false)
});

/**
 * Una ventana de captación o la ruleta (docs/especificaciones/tienda-promociones.md): devuelve el
 * premio (y en la ruleta la porción que salió). Una vez por email y ventana.
 */
export async function manejar(event: RequestEvent<{ tienda: string }>, que: 'captar' | 'ruleta') {
	const { params, cookies, request } = event;
	const texto = await request.text();
	if (texto.length > 2048) error(413, 'Demasiado largo.');
	let cuerpo: unknown = null;
	try {
		cuerpo = JSON.parse(texto);
	} catch {
		// queda null y falla la validación
	}
	const b = v.safeParse(Pedido, cuerpo);
	if (!b.success) return json({ error: 'Revisá el email.', campo: 'email' }, { status: 400 });
	const r = await resolverTienda(params.tienda, cookies.get(COOKIE));
	if (!r) error(404, 'Esta tienda no existe.');
	const res = await participar(r, clienteDe(event), que, b.output);
	if (!res.ok) return json({ error: res.error, campo: res.campo ?? null }, { status: res.estado });
	return json(res.resultado);
}
