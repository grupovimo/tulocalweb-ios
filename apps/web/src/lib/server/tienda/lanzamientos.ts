/**
 * «Avisame cuando salga» y la reserva o el encargo con seña de la tienda online (§52), en
 * /t/[tienda]/aviso y /t/[tienda]/sena. Con la API, Go valida, pone el tope por IP, calcula la seña
 * y arma el pago con la cuenta de Mercado Pago de la tienda. En la demo se guarda en el sandbox y
 * la seña se da por pagada (la demo no cobra).
 */
import { env } from 'cloudflare:workers';
import { json, type RequestEvent } from '@sveltejs/kit';
import * as v from 'valibot';
import { ErrorApi } from '../bff.ts';
import { usaSandbox, vidrieraDe } from '../datos/index.ts';
import { golpe } from '../limite.ts';
import { COOKIE, dbSandbox } from '../sandbox/sandbox.ts';
import { codigoPedido } from './pedido.ts';
import { NuevaSena, NuevoAviso } from './esquemas.ts';
import { resolverTienda } from './publica.ts';

type Evento = RequestEvent<{ tienda: string }>;

const error = (estado: number, mensaje: string) => json({ error: mensaje }, { status: estado });
const MUCHAS = 'Mandaste muchas seguidas. Probá de nuevo en un rato.';

/** Diez por hora por IP, como Go (con la API, el tope lo pone Go). */
async function dentroDelTope(event: Evento, tipo: string) {
	let ip = 'sin-ip';
	try {
		ip = event.request.headers.get('cf-connecting-ip') ?? event.getClientAddress();
	} catch {
		// Sin dirección (tests): cuenta como una sola.
	}
	return golpe(dbSandbox(env.SANDBOX), `tienda:${tipo}:${ip}`, 10, 3600);
}

async function preparar<E extends v.GenericSchema>(event: Evento, tipo: string, esquema: E) {
	const r = await resolverTienda(event.params.tienda, event.cookies.get(COOKIE));
	if (!r) return error(404, 'Esta tienda no existe.');
	if (r.tienda.vitrina) return error(409, 'Es una tienda de ejemplo: no se reserva de verdad.');
	const b = v.safeParse(esquema, await event.request.json().catch(() => null));
	if (!b.success) return error(400, b.issues[0].message);
	if (usaSandbox() && !(await dentroDelTope(event, tipo))) return error(429, MUCHAS);
	return { r, datos: b.output as v.InferOutput<E> };
}

function deLaApi(e: unknown) {
	if (!(e instanceof ErrorApi)) throw e;
	if (e.estado === 409 || e.estado === 404) return error(409, e.message);
	if (e.estado === 429) return error(429, MUCHAS);
	return error(e.estado >= 500 ? 503 : 400, e.message);
}

export async function avisar(event: Evento) {
	const p = await preparar(event, 'aviso', NuevoAviso);
	if (p instanceof Response) return p;
	const { r, datos } = p;
	if (r.datos) {
		const res = await r.datos.avisarLanzamiento(datos.productoId, datos.email);
		if (res === 'no-existe') return error(404, 'Ese producto ya no está publicado.');
		if (res === 'ya-salio') return error(409, 'Este producto ya está a la venta.');
		if (res === 'lleno') return error(409, 'No podemos anotar más avisos para este producto.');
		return json({ ok: true }, { status: 201 });
	}
	try {
		await vidrieraDe(event).aviso(r.tienda.slug, datos.productoId, datos.email);
		return json({ ok: true }, { status: 201 });
	} catch (e) {
		return deLaApi(e);
	}
}

export async function senar(event: Evento) {
	const p = await preparar(event, 'sena', NuevaSena);
	if (p instanceof Response) return p;
	const { r, datos } = p;
	if (r.datos && r.sandboxId) {
		const res = await r.datos.senar(datos);
		if (res === 'sin-stock') return error(409, 'No queda stock para reservar.');
		if (res === 'sin-sena')
			return error(409, 'Este producto no se reserva ni se encarga con seña.');
		if (res === 'invalido') return error(400, 'Elegí de nuevo la variante.');
		const codigo = await codigoPedido(r.sandboxId, res.numero);
		return json({ url: `/t/${r.tienda.slug}/pedido/${codigo}?sena=demo`, ...res }, { status: 201 });
	}
	try {
		const res = await vidrieraDe(event).sena(r.tienda.slug, datos);
		return json(res, { status: 201 });
	} catch (e) {
		return deLaApi(e);
	}
}
