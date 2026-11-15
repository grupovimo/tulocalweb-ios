/**
 * Las ventanas de la tienda online (docs/especificaciones/tienda-promociones.md): captar, girar la
 * ruleta, mandar las métricas y leer las compras recientes. Con la API real lo hace Go
 * (dominio/promos, con sus topes por IP y tienda); en la demo, el sandbox del visitante con las
 * mismas reglas (sandbox/promos.ts) y en la vitrina de una plantilla el premio se muestra sin
 * guardar nada.
 */
import { env } from 'cloudflare:workers';
import { ClienteBff, ErrorApi } from '../bff.ts';
import { dbSandbox } from '../sandbox/sandbox.ts';
import { golpe } from '../limite.ts';
import type { CompraReciente, ResultadoPromo } from '../sandbox/promos.ts';
import type { TiendaResuelta } from './publica.ts';

export type PedidoPromo = {
	ventana: string;
	email: string;
	nombre?: string;
	whatsapp?: string;
	consentimiento: boolean;
};
export type Evento = { ventana: string; tipo: 'vista' | 'cierre' | 'clic'; n?: number };

class PromosApi extends ClienteBff {
	participar(slug: string, que: 'captar' | 'ruleta', p: PedidoPromo) {
		return this.pedir<ResultadoPromo>('POST', `/v1/publico/tiendas/${slug}/promos/${que}`, p);
	}
	eventos(slug: string, eventos: Evento[]) {
		return this.pedir<void>('POST', `/v1/publico/tiendas/${slug}/promos/eventos`, { eventos });
	}
	recientes(slug: string) {
		return this.pedir<{ compras: CompraReciente[] }>(
			'GET',
			`/v1/publico/tiendas/${slug}/promos/recientes`
		);
	}
}

const api = (cliente: { ip: string; agente: string }) =>
	new PromosApi(String(env.API_URL), String(env.CLAVE_BFF ?? ''), cliente);

export type Participacion =
	| { ok: true; resultado: ResultadoPromo }
	| { ok: false; estado: number; error: string; campo?: string };

const MENSAJES = {
	invalido: { estado: 400, error: 'Revisá el email.', campo: 'email' },
	'no-existe': { estado: 404, error: 'Esta promo no existe.' },
	terminada: { estado: 409, error: 'Esta promo ya terminó. ¡Gracias por tu interés!' }
} as const;

/** Captar (contenido captacion) o girar la ruleta. */
export async function participar(
	r: TiendaResuelta,
	cliente: { ip: string; agente: string },
	que: 'captar' | 'ruleta',
	p: PedidoPromo
): Promise<Participacion> {
	if (!p.consentimiento)
		return {
			ok: false,
			estado: 400,
			error: 'Marcá la casilla para aceptar que te escriban.',
			campo: 'consentimiento'
		};
	if (!r.datos) {
		try {
			return { ok: true, resultado: await api(cliente).participar(r.tienda.slug, que, p) };
		} catch (e) {
			if (e instanceof ErrorApi) return { ok: false, estado: e.estado, error: e.message };
			throw e;
		}
	}
	// En la demo, como mucho 20 por hora por visitante (cada uno escribe en la D1).
	if (r.sandboxId && !(await golpe(dbSandbox(env.SANDBOX), `promos:${r.sandboxId}`, 20, 3600)))
		return {
			ok: false,
			estado: 429,
			error: 'Recibimos varios pedidos seguidos. Probá en un rato.'
		};
	const res = await r.datos.participarPromo(
		que === 'captar' ? 'captacion' : 'ruleta',
		p,
		!r.sandboxId
	);
	if (typeof res === 'string') return { ok: false, ...MENSAJES[res] };
	return { ok: true, resultado: res };
}

/** Las métricas en lote. En la vitrina no se guarda nada; los errores no le importan a la tienda. */
export async function registrarEventos(
	r: TiendaResuelta,
	cliente: { ip: string; agente: string },
	eventos: Evento[]
) {
	if (!eventos.length) return;
	if (!r.datos) {
		try {
			await api(cliente).eventos(r.tienda.slug, eventos.slice(0, 50));
		} catch (e) {
			if (!(e instanceof ErrorApi)) throw e;
		}
		return;
	}
	if (!r.sandboxId) return;
	if (!(await golpe(dbSandbox(env.SANDBOX), `promos-eventos:${r.sandboxId}`, 120, 3600))) return;
	await r.datos.eventosPromo(eventos);
}

/** La prueba social: compras pagadas de los últimos 7 días (vacío si la tienda no la prendió). */
export async function recientes(
	r: TiendaResuelta,
	cliente: { ip: string; agente: string }
): Promise<CompraReciente[]> {
	if (!r.datos) {
		try {
			return (await api(cliente).recientes(r.tienda.slug)).compras;
		} catch (e) {
			if (e instanceof ErrorApi) return [];
			throw e;
		}
	}
	return r.datos.promosRecientes();
}
