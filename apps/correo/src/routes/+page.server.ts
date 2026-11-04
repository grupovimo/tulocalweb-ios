import { error, fail } from '@sveltejs/kit';
import {
	accion,
	bandeja,
	CARPETAS,
	mensaje,
	type CarpetaId,
	type Pestana,
	vaciarPapelera
} from '#lib/server/bandeja.ts';
import { ErrorJmap, Jmap } from '#lib/server/jmap.ts';
import { guardarFiltros, leerFiltros } from '#lib/server/ajustes.ts';
import { marca } from '#lib/server/marca.ts';
import type { Actions, PageServerLoad } from './$types';

const PESTANAS: Pestana[] = ['todos', 'sinleer'];

export const load: PageServerLoad = async ({ locals, url, depends, cookies }) => {
	depends('correo:bandeja');
	const j = new Jmap(locals.cred!);
	const c = url.searchParams.get('c');
	const carpeta: CarpetaId = CARPETAS.some((x) => x.id === c) ? (c as CarpetaId) : 'entrada';
	const p = url.searchParams.get('p') as Pestana;
	const pestana = PESTANAS.includes(p) ? p : 'todos';
	const casilla = url.searchParams.get('cas');
	const buscar = (url.searchParams.get('q') ?? '').slice(0, 200);
	const m = url.searchParams.get('m');
	try {
		const [b, leido] = await Promise.all([
			bandeja(j, { carpeta, casilla, pestana, buscar }),
			m ? mensaje(j, m, url.searchParams.get('img') === '1').catch(() => null) : null
		]);
		const tema: 'claro' | 'oscuro' = cookies.get('correo_tema') === 'claro' ? 'claro' : 'oscuro';
		return { ...b, carpeta, casilla, pestana, buscar, leido, marca: marca(), tema };
	} catch (e) {
		console.error('bandeja:', e instanceof Error ? e.message : e);
		if (e instanceof ErrorJmap) error(e.estado === 401 ? 401 : 502, e.message);
		throw e;
	}
};

const QUE = [
	'archivar',
	'borrar',
	'destacar',
	'quitar-destacado',
	'no-leido',
	'leido',
	'spam',
	'entrada',
	'etiquetar',
	'desetiquetar',
	'mover',
	'cancelar-envio'
] as const;

export const actions: Actions = {
	accion: async ({ locals, request }) => {
		const f = await request.formData();
		const que = String(f.get('que')) as (typeof QUE)[number];
		const ids = f.getAll('id').map(String).filter(Boolean);
		if (!QUE.includes(que) || !ids.length) return fail(400, { error: 'Pedido inválido.' });
		try {
			await accion(new Jmap(locals.cred!), ids, que, String(f.get('valor') ?? ''));
		} catch (e) {
			return fail(502, { error: e instanceof Error ? e.message : 'No se pudo.' });
		}
		return { ok: que };
	},
	/** Vaciar la papelera (de la casilla que se está viendo, o de todas). */
	vaciar: async ({ locals, request }) => {
		const casilla = String((await request.formData()).get('casilla') ?? '') || null;
		try {
			const n = await vaciarPapelera(new Jmap(locals.cred!), casilla);
			return { ok: 'vaciar', n };
		} catch (e) {
			return fail(502, { error: e instanceof Error ? e.message : 'No se pudo.' });
		}
	},
	/** Bloquear a un remitente: una regla más de la casilla propia que lo manda a Spam. */
	bloquear: async ({ locals, request }) => {
		const email = String((await request.formData()).get('email') ?? '')
			.trim()
			.toLowerCase();
		if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email))
			return fail(400, { error: 'Remitente inválido.' });
		const j = new Jmap(locals.cred!);
		try {
			const f = await leerFiltros(j);
			if (!f.reglas.some((x) => x.si === 'de' && x.valor === email && x.accion === 'mover'))
				f.reglas.push({ si: 'de', valor: email, accion: 'mover', destino: 'Junk Mail' });
			await guardarFiltros(j, f);
		} catch (e) {
			return fail(502, { error: e instanceof Error ? e.message : 'No se pudo.' });
		}
		return { ok: 'bloquear' };
	}
};
