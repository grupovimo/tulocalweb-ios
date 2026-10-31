import { error, json } from '@sveltejs/kit';
import * as v from 'valibot';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

const Fecha = v.nullable(v.pipe(v.string(), v.isoTimestamp('Revisá la fecha.')));
const Pct = v.nullable(
	v.pipe(v.number(), v.integer(), v.minValue(1, 'La seña va de 1 % a 100 %.'), v.maxValue(100))
);

const Config = v.object({
	lanzamiento: Fecha,
	reserva: Pct,
	encargo: Pct,
	ofertaPrecio: v.nullable(
		v.pipe(v.number(), v.integer(), v.minValue(1), v.maxValue(1_000_000_000))
	),
	ofertaHasta: Fecha
});

/** Lanzamiento, señas y oferta con fin del producto (§52), con avisos y reservas. */
export const GET: RequestHandler = async ({ locals, params }) => {
	const l = await datosDe(locals).lanzamiento(params.id);
	if (!l) error(404, 'Ese producto no existe.');
	return json(l, { headers: { 'cache-control': 'no-store' } });
};

export const PUT: RequestHandler = async ({ locals, params, request }) => {
	const r = v.safeParse(Config, await request.json().catch(() => null));
	if (!r.success) error(400, r.issues[0].message);
	const l = await datosDe(locals).guardarLanzamiento(params.id, r.output);
	if (!l) error(404, 'Ese producto no existe.');
	if ('error' in l) error(400, l.error);
	return json(l);
};
