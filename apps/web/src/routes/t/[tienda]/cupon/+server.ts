import { error, json } from '@sveltejs/kit';
import * as v from 'valibot';
import { env } from 'cloudflare:workers';
import { golpe } from '#lib/server/limite.ts';
import { COOKIE, dbSandbox } from '#lib/server/sandbox/sandbox.ts';
import { evaluarCupon } from '#lib/server/tienda/cupon.ts';
import { resolverTienda } from '#lib/server/tienda/publica.ts';
import type { RequestHandler } from './$types';

const Pedido = v.object({
	codigo: v.pipe(v.string(), v.maxLength(30)),
	email: v.optional(v.pipe(v.string(), v.maxLength(120)), ''),
	carrito: v.pipe(
		v.array(
			v.object({
				id: v.pipe(v.string(), v.maxLength(80)),
				// La variante cuenta para el precio (el suyo o el mayorista, §49).
				varianteId: v.optional(v.nullable(v.pipe(v.string(), v.maxLength(120))), null),
				cantidad: v.pipe(v.number(), v.integer(), v.minValue(1), v.maxValue(99))
			})
		),
		v.maxLength(50)
	)
});

/** Vista previa del cupón en el carrito y el checkout: al comprar se vuelve a evaluar. */
export const POST: RequestHandler = async ({ params, request, cookies, getClientAddress }) => {
	// Tope por IP: que no se puedan adivinar códigos probando.
	const ip = request.headers.get('cf-connecting-ip') ?? getClientAddress();
	if (!(await golpe(dbSandbox(env.SANDBOX), `cupon:${ip}`, 30, 600)))
		return json({ ok: false, motivo: 'Probaste muchos cupones seguidos. Esperá unos minutos.' });
	const r = await resolverTienda(params.tienda, cookies.get(COOKIE));
	if (!r) error(404, 'Esta tienda no existe.');
	const b = v.safeParse(Pedido, await request.json().catch(() => null));
	if (!b.success) error(400, 'Pedido inválido.');
	return json(await evaluarCupon(r, b.output.codigo, b.output.email, b.output.carrito));
};
