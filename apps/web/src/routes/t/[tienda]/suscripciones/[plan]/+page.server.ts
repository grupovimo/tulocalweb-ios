import { error, fail, redirect } from '@sveltejs/kit';
import * as v from 'valibot';
import { irAfuera } from '#lib/server/afuera.ts';
import { usaSandbox } from '#lib/server/datos/index.ts';
import { COOKIE } from '#lib/server/sandbox/sandbox.ts';
import { AltaSuscripcion } from '#lib/server/suscripciones-esquemas.ts';
import { resolverTienda } from '#lib/server/tienda/publica.ts';
import { suscribir } from '#lib/server/tienda/suscripciones.ts';
import type { Actions, PageServerLoad } from './$types';

/** La página de un plan de suscripción con su alta (§51). */
export const load: PageServerLoad = async ({ parent, params }) => {
	const { tienda } = await parent();
	const plan = tienda.suscripciones.find((p) => p.id === params.plan);
	if (!plan) error(404, 'Este plan ya no está disponible.');
	return { plan, demo: usaSandbox() };
};

export const actions: Actions = {
	default: async ({ request, params, cookies, getClientAddress }) => {
		const r = await resolverTienda(params.tienda, cookies.get(COOKIE));
		if (!r || !r.tienda.suscripciones.some((p) => p.id === params.plan))
			error(404, 'Este plan ya no está disponible.');
		if (r.tienda.vitrina)
			return fail(409, { error: 'Es una tienda de ejemplo: no se suscribe de verdad.' });
		const b = v.safeParse(AltaSuscripcion, {
			...Object.fromEntries(await request.formData()),
			planId: params.plan
		});
		if (!b.success) return fail(400, { error: b.issues[0].message });
		// El precio sale del plan en el servidor (Go o el sandbox); el tope de altas lo pone Go.
		const ip = request.headers.get('cf-connecting-ip') ?? getClientAddress();
		const res = await suscribir(r, b.output, ip);
		if ('error' in res) return fail(res.estado, { error: res.error });
		if (res.url) irAfuera(res.url);
		redirect(303, `/t/${r.tienda.slug}/suscripcion/${res.codigo}?alta=1`);
	}
};
