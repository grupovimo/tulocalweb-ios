/**
 * Suscripciones de la tienda online (§51): el alta desde la página del plan y el link privado
 * (/t/[tienda]/suscripcion/[codigo]). En la demo las atiende el sandbox del visitante (sin cobrar);
 * con DATA_SOURCE=api, Go (/v1/publico/tiendas/{slug}/suscripciones). El navegador nunca habla con Go.
 */
import { env } from 'cloudflare:workers';
import type { SuscripcionPublica } from '#lib/tienda/tipos.ts';
import { ENTIDAD } from '../sandbox/suscripciones.ts';
import { buscarCambio, DatosSandbox, dbSandbox, type Sandbox } from '../sandbox/sandbox.ts';
import type { TiendaResuelta } from './publica.ts';

const api = (ruta: string, init?: RequestInit) => fetch(`${String(env.API_URL)}${ruta}`, init);
const enc = encodeURIComponent;

export type Alta = {
	planId: string;
	nombre: string;
	email: string;
	telefono: string;
	entrega: 'envio' | 'retiro';
	direccion: string;
	cp: string;
};

/** El alta: el link privado y, con la API, adónde autorizar el débito en Mercado Pago. */
export async function suscribir(
	r: TiendaResuelta,
	a: Alta,
	/** IP de quien se suscribe: Go la usa (con la clave del BFF) para el tope de altas. */
	ip = ''
): Promise<
	{ codigo: string; url?: string } | { error: string; estado: 400 | 404 | 409 | 429 | 502 }
> {
	if (!r.datos || !r.sandboxId) {
		const res = await api(`/v1/publico/tiendas/${enc(r.tienda.slug)}/suscripciones`, {
			method: 'POST',
			headers: {
				'content-type': 'application/json',
				...(ip ? { 'x-mostrador-bff': String(env.CLAVE_BFF ?? ''), 'x-cliente-ip': ip } : {})
			},
			body: JSON.stringify(a)
		});
		if (!res.ok) {
			const e = (await res.json().catch(() => null)) as { detail?: string } | null;
			const estado = ([404, 409, 429, 502] as const).find((x) => x === res.status) ?? 400;
			return { estado, error: e?.detail ?? 'No pudimos sumarte. Probá de nuevo.' };
		}
		return (await res.json()) as { codigo: string; url?: string };
	}
	const s = await r.datos.suscribirPublico(a);
	if (s === 'no-existe') return { estado: 404, error: 'Este plan ya no está disponible.' };
	if (s === 'cupo') return { estado: 409, error: 'Este plan no tiene más lugar por ahora.' };
	if (s === 'invalido') return { estado: 400, error: 'Elegí cómo lo recibís.' };
	return s;
}

const CODIGO = /^[A-Za-z0-9_-]{16,64}$/;

/** En la demo, el sandbox dueño de la suscripción por su código (el link anda sin la cookie). */
async function sandboxDe(slug: string, codigo: string) {
	const db = dbSandbox(env.SANDBOX);
	const c = await buscarCambio<{ id: string }>(db, ENTIDAD, `c:${codigo}`);
	if (!c) return null;
	const datos = new DatosSandbox(db, c.sb as Sandbox);
	return (await datos.tienda()).slug === slug ? datos : null;
}

/** La suscripción por su link privado. */
export async function suscripcionDe(
	slug: string,
	codigo: string,
	sandbox: boolean
): Promise<SuscripcionPublica | null> {
	if (!CODIGO.test(codigo)) return null;
	if (!sandbox) {
		const r = await api(`/v1/publico/tiendas/${enc(slug)}/suscripciones/${enc(codigo)}`);
		if (r.status === 404) return null;
		if (!r.ok) throw new Error(`suscripciones: la API respondió ${r.status}`);
		return (await r.json()) as SuscripcionPublica;
	}
	return (await sandboxDe(slug, codigo))?.suscripcionPublica(codigo) ?? null;
}

export type Accion = 'pausar' | 'reanudar' | 'cancelar';

/** Pausar, reanudar o cancelar desde el link privado. */
export async function cambiar(
	slug: string,
	codigo: string,
	accion: Accion,
	sandbox: boolean
): Promise<SuscripcionPublica | 'no-existe' | 'estado' | 'mercadopago'> {
	if (!CODIGO.test(codigo)) return 'no-existe';
	if (!sandbox) {
		const r = await api(`/v1/publico/tiendas/${enc(slug)}/suscripciones/${enc(codigo)}/${accion}`, {
			method: 'POST'
		});
		if (r.status === 404) return 'no-existe';
		if (r.status === 409) return 'estado';
		if (r.status === 502) return 'mercadopago';
		if (!r.ok) throw new Error(`suscripciones: la API respondió ${r.status}`);
		return (await r.json()) as SuscripcionPublica;
	}
	const d = await sandboxDe(slug, codigo);
	return d ? d.cambiarSuscripcionPublica(codigo, accion) : 'no-existe';
}
