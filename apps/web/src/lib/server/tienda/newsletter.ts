/**
 * El newsletter desde la tienda online (docs/decisiones.md §50): sumarse (pie o checkout), confirmar
 * con el link del mail y la baja de un clic. Con la API real lo hace Go (con sus topes por IP, tienda
 * y email); en la demo, el sandbox del visitante, sin mails: la tienda muestra el link de confirmar.
 */
import { env } from 'cloudflare:workers';
import { ClienteBff, ErrorApi } from '../bff.ts';
import { dbSandbox } from '../sandbox/sandbox.ts';
import { golpe } from '../limite.ts';
import type { TiendaResuelta } from './publica.ts';

class NewsletterApi extends ClienteBff {
	suscribir(slug: string, email: string) {
		return this.pedir<void>('POST', `/v1/publico/tiendas/${slug}/newsletter`, { email });
	}
	accion(slug: string, accion: 'confirmar' | 'baja', token: string) {
		return this.pedir<void>('POST', `/v1/publico/tiendas/${slug}/newsletter/${accion}`, { token });
	}
}

const api = (cliente: { ip: string; agente: string }) =>
	new NewsletterApi(String(env.API_URL), String(env.CLAVE_BFF ?? ''), cliente);

export type Suscripcion =
	{ ok: true; demo?: { link: string | null } } | { ok: false; estado: number; error: string };

/**
 * Suma el email como pendiente. En la demo devuelve el link de confirmar (no sale ningún mail);
 * `link` null: ya estaba confirmado.
 */
export async function suscribir(
	r: TiendaResuelta,
	cliente: { ip: string; agente: string },
	email: string
): Promise<Suscripcion> {
	if (r.tienda.vitrina)
		return { ok: false, estado: 409, error: 'Es una tienda de ejemplo: no guarda emails.' };
	if (!r.datos || !r.sandboxId) {
		try {
			await api(cliente).suscribir(r.tienda.slug, email);
			return { ok: true };
		} catch (e) {
			if (e instanceof ErrorApi) return { ok: false, estado: e.estado, error: e.message };
			throw e;
		}
	}
	// En la demo, como mucho 20 por hora por visitante (cada uno escribe en la D1).
	if (!(await golpe(dbSandbox(env.SANDBOX), `newsletter:${r.sandboxId}`, 20, 3600)))
		return {
			ok: false,
			estado: 429,
			error: 'Recibimos varios pedidos seguidos. Probá en un rato.'
		};
	const s = await r.datos.suscribirNewsletter(email, 'tienda');
	if (!s) return { ok: false, estado: 400, error: 'Revisá el email.' };
	return {
		ok: true,
		demo: { link: s.confirmado ? null : `/t/${r.tienda.slug}/newsletter/confirmar/${s.token}` }
	};
}

/** La casilla «Quiero recibir novedades» del checkout de la demo (con la API la manda el pedido). */
export async function suscribirDesdeCheckout(r: TiendaResuelta, email: string) {
	if (r.datos && r.sandboxId) await r.datos.suscribirNewsletter(email, 'checkout');
}

const tokenValido = (t: string) => /^[A-Za-z0-9_-]{20,64}$/.test(t);

/** Confirmar o darse de baja con el token del mail. false: el link ya no sirve. */
export async function accionToken(
	r: TiendaResuelta,
	cliente: { ip: string; agente: string },
	accion: 'confirmar' | 'baja',
	token: string
): Promise<boolean> {
	if (!tokenValido(token) || r.tienda.vitrina) return false;
	if (!r.datos || !r.sandboxId) {
		try {
			await api(cliente).accion(r.tienda.slug, accion, token);
			return true;
		} catch (e) {
			if (e instanceof ErrorApi && e.estado === 404) return false;
			throw e;
		}
	}
	return accion === 'confirmar'
		? r.datos.confirmarNewsletter(token)
		: r.datos.bajaNewsletter(token);
}
