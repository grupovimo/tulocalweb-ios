import { error } from '@sveltejs/kit';
import * as v from 'valibot';
import { cbuValido } from '#lib/panel/cbu.ts';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

const txt = (max: number) => v.pipe(v.string(), v.trim(), v.maxLength(max));
const Negocio = v.object({
	nombre: v.pipe(txt(80), v.minLength(1, 'Poné el nombre del negocio.')),
	razonSocial: txt(120),
	cuit: v.pipe(txt(13), v.regex(/^$|^\d{2}-?\d{8}-?\d$/, 'El CUIT tiene 11 números.')),
	condicion: txt(60),
	direccion: txt(160),
	whatsapp: txt(30),
	email: v.pipe(txt(120), v.regex(/^$|^[^\s@]+@[^\s@]+\.[^\s@]+$/, 'Revisá el email.')),
	horario: txt(80),
	// Para cobrar por transferencia en la tienda online (sin alias ni CBU no se ofrece).
	alias: v.optional(
		v.pipe(
			txt(20),
			v.regex(
				/^$|^[a-zA-Z0-9.-]{6,20}$/,
				'El alias tiene de 6 a 20 letras, números, puntos o guiones.'
			)
		),
		''
	),
	cbu: v.optional(
		v.pipe(
			txt(26),
			v.transform((s) => s.replace(/\s/g, '')),
			v.check(
				(s) => s === '' || cbuValido(s),
				'Revisá el CBU/CVU: tiene 22 números y algún dígito no coincide.'
			)
		),
		''
	),
	titular: v.optional(txt(80), ''),
	// Cambios y devoluciones, a la vista en la tienda (vacía: el texto legal base).
	politica: v.optional(
		v.pipe(v.string(), v.trim(), v.maxLength(1000, 'La política es muy larga (hasta 1000).')),
		''
	)
});

/** Datos del negocio. */
export const PUT: RequestHandler = async ({ locals, request }) => {
	const r = v.safeParse(Negocio, await request.json().catch(() => null));
	if (!r.success) error(400, r.issues[0].message);
	await datosDe(locals).guardarNegocio(r.output);
	return new Response(null, { status: 204 });
};
