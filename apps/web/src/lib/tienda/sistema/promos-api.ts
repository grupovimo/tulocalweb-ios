/**
 * Lo que la tienda le pide a la API por las promociones (tienda-promociones.md › API): captar un
 * contacto, girar la ruleta, los eventos de las ventanas en lote y las compras recientes. Todo va a
 * /t/<slug>/promos/… (en un dominio propio también, como la newsletter).
 */
import type { Premio } from './promos.ts';

/** El premio que devuelve el servidor al captar o al girar (components/schemas/PromoPremio). */
export type PremioGanado = {
	tipo: Premio['tipo'];
	codigo?: string;
	descripcion?: string;
	descarga?: string;
	nombreArchivo?: string;
	vence?: string;
};

export type Captado = { premio: PremioGanado; repetido?: boolean; demo?: boolean };
export type Girado = Captado & { porcion?: number; texto?: string };
export type Fallo = { error: string };

async function pedir<T>(url: string, cuerpo: unknown): Promise<T | Fallo> {
	try {
		const r = await fetch(url, {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify(cuerpo)
		});
		const j = (await r.json().catch(() => null)) as
			(T & { message?: string; error?: string; detail?: string; title?: string }) | null;
		if (!r.ok || !j)
			return {
				error:
					j?.error ??
					j?.detail ??
					j?.message ??
					j?.title ??
					(r.status === 429
						? 'Probaste muchas veces seguidas. Esperá unos minutos.'
						: 'No pudimos anotarte. Probá de nuevo.')
			};
		return j;
	} catch {
		return { error: 'No hay conexión. Probá de nuevo.' };
	}
}

export type Contacto = {
	ventana: string;
	email: string;
	nombre?: string;
	whatsapp?: string;
	consentimiento: true;
};

export const captar = (slug: string, c: Contacto) => pedir<Captado>(`/t/${slug}/promos/captar`, c);
export const girar = (slug: string, c: Contacto) => pedir<Girado>(`/t/${slug}/promos/ruleta`, c);
export const esFallo = (x: unknown): x is Fallo =>
	!!x && typeof x === 'object' && typeof (x as Fallo).error === 'string';

// ---------------------------------------------------------------------------------------------
// Eventos: se juntan por (ventana, tipo) y salen en lote con sendBeacon al esconder la página.
// Los envíos (captación y ruleta) los cuenta el servidor al captar: acá no se mandan.

export type TipoEvento = 'vista' | 'cierre' | 'clic';
type Evento = { ventana: string; tipo: TipoEvento; n: number };

let fila: { slug: string; eventos: Evento[] } | null = null;
let escuchando = false;

export function mandarEventos() {
	if (!fila?.eventos.length) return;
	const { slug, eventos } = fila;
	fila = { slug, eventos: [] };
	const url = `/t/${slug}/promos/eventos`;
	const cuerpo = JSON.stringify({ eventos: eventos.slice(0, 50) });
	try {
		if (navigator.sendBeacon?.(url, new Blob([cuerpo], { type: 'application/json' }))) return;
	} catch {
		/* sin beacon: fetch con keepalive */
	}
	fetch(url, {
		method: 'POST',
		headers: { 'content-type': 'application/json' },
		body: cuerpo,
		keepalive: true
	}).catch(() => {});
}

export function evento(slug: string, ventana: string, tipo: TipoEvento) {
	if (typeof window === 'undefined') return;
	if (fila && fila.slug !== slug) mandarEventos();
	fila ??= { slug, eventos: [] };
	const e = fila.eventos.find((x) => x.ventana === ventana && x.tipo === tipo && x.n < 20);
	if (e) e.n++;
	else fila.eventos.push({ ventana, tipo, n: 1 });
	if (!escuchando) {
		escuchando = true;
		addEventListener('pagehide', mandarEventos);
		document.addEventListener('visibilitychange', () => {
			if (document.visibilityState === 'hidden') mandarEventos();
		});
	}
	if (fila.eventos.length >= 40) mandarEventos();
}

// ---------------------------------------------------------------------------------------------
// Compras recientes (prueba social): sólo pedidos reales pagados de los últimos 7 días.

export type Reciente = {
	nombre: string;
	/** '' si no se sabe (no se inventa). */
	provincia: string;
	producto: string;
	productoId: string | null;
	foto: string | null;
	cuando: string;
};

export async function recientes(slug: string): Promise<Reciente[]> {
	try {
		const r = await fetch(`/t/${slug}/promos/recientes`);
		if (!r.ok) return [];
		const j = (await r.json()) as { compras?: Reciente[] };
		return (j.compras ?? []).filter(
			(x) => x && typeof x.nombre === 'string' && x.nombre && typeof x.producto === 'string'
		);
	} catch {
		return [];
	}
}

/** «hace 2 h», «hace 3 días»: cuánto pasó desde la compra. */
export function hace(iso: string, ahora = Date.now()): string {
	const m = Math.max(1, Math.round((ahora - Date.parse(iso)) / 60_000));
	if (!Number.isFinite(m)) return '';
	if (m < 60) return `hace ${m} min`;
	const h = Math.round(m / 60);
	if (h < 24) return `hace ${h} h`;
	const d = Math.round(h / 24);
	return d === 1 ? 'ayer' : `hace ${d} días`;
}

/** Deja el cupón anotado para el carrito y el checkout (como el link de un cupón: CampoCupon). */
export function anotarCupon(slug: string, base: string, codigo: string) {
	if (!/^[A-Z0-9-]{3,20}$/.test(codigo)) return;
	try {
		document.cookie = `cupon-${slug}=${codigo}; path=${base || '/'}; max-age=${7 * 86_400}; samesite=lax`;
	} catch {
		/* sin cookies: queda el «Copiar» */
	}
}
