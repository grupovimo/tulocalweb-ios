/**
 * Cuándo sale una ventana en la tienda (tienda-promociones.md › Reglas), sin DOM: qué ventana
 * califica en esta página, cómo se dibuja en el celular, la memoria de lo ya visto (localStorage y
 * sessionStorage, siempre con try/catch) y la cuenta regresiva. Los disparadores y el dibujo están
 * en marco/Ventanas.svelte y marco/Ventana.svelte.
 */
import { completa, type PaginaVentana, type VarianteVentana, type Ventana } from './promos.ts';

/** Las páginas de la tienda que no son del catálogo de ventanas pero donde pueden salir (si «todas»). */
const OTRAS = ['turnos', 'suscripciones', 'suscripcion'];
const DE_VENTANA: PaginaVentana[] = ['inicio', 'catalogo', 'producto', 'carrito', 'nosotros'];

/**
 * La página como la entienden las ventanas: una del catálogo, 'otra' (sale sólo si la ventana va en
 * todas) o null: nunca sale (regla 2: checkout, listo, legales, arrepentimiento, newsletter…).
 */
export function paginaDeVentana(pagina: string): PaginaVentana | 'otra' | null {
	if ((DE_VENTANA as string[]).includes(pagina)) return pagina as PaginaVentana;
	return OTRAS.includes(pagina) ? 'otra' : null;
}

/** Lo que recuerda el dispositivo (localStorage): visitas, cuándo vio cada una y qué completó. */
export type Memoria = {
	visitas: number;
	vistas: Record<string, number>;
	/** Captaciones y ruletas ya completadas: no se vuelven a ver (regla 5). */
	hechas: string[];
};
/** Lo de esta visita (sessionStorage). */
export type Sesion = { vistas: string[] };

export type Ocasion = {
	pagina: string;
	movil: boolean;
	ahora: number;
	memoria: Memoria;
	sesion: Sesion;
};

const DIA = 86_400_000;

/** Si la ventana califica en esta ocasión (todo menos el disparo). */
export function califica(w: Ventana, o: Ocasion): boolean {
	if (!w.activa || !completa(w)) return false;
	const p = paginaDeVentana(o.pagina);
	if (!p) return false;
	if (w.paginas.length && (p === 'otra' || !w.paginas.includes(p))) return false;
	if (w.dispositivos === 'celular' && !o.movil) return false;
	if (w.dispositivos === 'compu' && o.movil) return false;
	if (w.desde && o.ahora < Date.parse(w.desde)) return false;
	if (w.hasta && o.ahora > Date.parse(w.hasta)) return false;
	// La cuenta regresiva del contenido: vencida, no sale más.
	const h = w.datos.hasta;
	if (typeof h === 'string' && o.ahora > Date.parse(h)) return false;
	if (w.publico === 'nuevos' && o.memoria.visitas > 1) return false;
	if (w.publico === 'recurrentes' && o.memoria.visitas < 2) return false;
	if ((w.contenido === 'captacion' || w.contenido === 'ruleta') && o.memoria.hechas.includes(w.id))
		return false;
	const vista = o.memoria.vistas[w.id];
	switch (w.frecuencia.cada) {
		case 'sesion':
			return !o.sesion.vistas.includes(w.id);
		case 'dias':
			return !vista || o.ahora - vista >= (w.frecuencia.dias ?? 7) * DIA;
		case 'unaVez':
			return !vista;
		case 'siempre':
			return true;
	}
}

/** Regla 1: una sola por visita a la página; si varias califican, la primera de la lista. */
export const elegir = (ventanas: Ventana[], o: Ocasion): Ventana | null =>
	ventanas.find((w) => califica(w, o)) ?? null;

/**
 * Cómo se dibuja (regla 3): en el celular, la completa siempre y la centrada que sale al entrar
 * suben como hoja desde abajo (deja ver la tienda); la lateral no entra en un costado y también.
 */
export function varianteEn(
	w: Pick<Ventana, 'variante' | 'disparo'>,
	movil: boolean
): VarianteVentana {
	if (!movil) return w.variante;
	if (w.variante === 'completa' || w.variante === 'lateral') return 'inferior';
	if (w.variante === 'centrada' && w.disparo.cuando === 'entrar') return 'inferior';
	return w.variante;
}

// ---------------------------------------------------------------------------------------------
// Memoria (con try/catch: en modo privado o con el almacenamiento bloqueado, todo sigue andando)

const clave = (slug: string) => `tlw-promos:${slug}`;
const vacia = (): Memoria => ({ visitas: 0, vistas: {}, hechas: [] });

function leer<T>(s: Storage | undefined, k: string): T | null {
	try {
		const x = s?.getItem(k);
		return x ? (JSON.parse(x) as T) : null;
	} catch {
		return null;
	}
}
function escribir(s: Storage | undefined, k: string, x: unknown) {
	try {
		s?.setItem(k, JSON.stringify(x));
	} catch {
		/* sin almacenamiento: se vuelve a mostrar, nada más */
	}
}
const local = () => (typeof localStorage === 'undefined' ? undefined : safe(() => localStorage));
const sesionS = () =>
	typeof sessionStorage === 'undefined' ? undefined : safe(() => sessionStorage);
function safe<T>(f: () => T): T | undefined {
	try {
		return f();
	} catch {
		return undefined;
	}
}

export function memoriaDe(slug: string): Memoria {
	const m = leer<Partial<Memoria>>(local(), clave(slug));
	return {
		visitas: typeof m?.visitas === 'number' ? m.visitas : 0,
		vistas: m?.vistas && typeof m.vistas === 'object' ? m.vistas : {},
		hechas: Array.isArray(m?.hechas) ? m.hechas : []
	};
}
export function sesionDe(slug: string): Sesion {
	const s = leer<Partial<Sesion>>(sesionS(), clave(slug));
	return { vistas: Array.isArray(s?.vistas) ? s.vistas : [] };
}

/** La primera página de una visita suma una visita (para «primera visita» y «los que vuelven»). */
export function contarVisita(slug: string): Memoria {
	const m = memoriaDe(slug);
	if (leer(sesionS(), `${clave(slug)}:visita`)) return m.visitas ? m : { ...m, visitas: 1 };
	escribir(sesionS(), `${clave(slug)}:visita`, 1);
	const n = { ...m, visitas: m.visitas + 1 };
	escribir(local(), clave(slug), n);
	return n;
}

export function anotarVista(slug: string, id: string, ahora = Date.now()) {
	const m = memoriaDe(slug);
	escribir(local(), clave(slug), { ...m, vistas: { ...m.vistas, [id]: ahora } });
	const s = sesionDe(slug);
	if (!s.vistas.includes(id)) escribir(sesionS(), clave(slug), { vistas: [...s.vistas, id] });
}

export function anotarHecha(slug: string, id: string) {
	const m = memoriaDe(slug);
	if (!m.hechas.includes(id)) escribir(local(), clave(slug), { ...m, hechas: [...m.hechas, id] });
}

export { vacia as memoriaVacia };

// ---------------------------------------------------------------------------------------------
// Cuenta regresiva

export type Faltan = { d: number; h: number; m: number; s: number };

/** Cuánto falta hasta `hasta` (ISO), o null si ya pasó o no es una fecha. */
export function faltan(hasta: unknown, ahora: number): Faltan | null {
	const t = typeof hasta === 'string' ? Date.parse(hasta) : NaN;
	if (!Number.isFinite(t) || t <= ahora) return null;
	let x = Math.floor((t - ahora) / 1000);
	const d = Math.floor(x / 86400);
	x -= d * 86400;
	const h = Math.floor(x / 3600);
	x -= h * 3600;
	const m = Math.floor(x / 60);
	return { d, h, m, s: x - m * 60 };
}

const dos = (n: number) => String(n).padStart(2, '0');
/** «2 d 04:10:05», «04:10:05»: corto, para la barra y las ventanas. */
export function textoFaltan(f: Faltan): string {
	const reloj = `${dos(f.h)}:${dos(f.m)}:${dos(f.s)}`;
	return f.d ? `${f.d}\u00a0d ${reloj}` : reloj;
}
/** Lo que lee un lector de pantalla: «Faltan 2 días y 4 horas». */
export function textoFaltanLargo(f: Faltan): string {
	const p = (n: number, a: string, b: string) => `${n} ${n === 1 ? a : b}`;
	if (f.d) return `Faltan ${p(f.d, 'día', 'días')} y ${p(f.h, 'hora', 'horas')}`;
	if (f.h) return `Faltan ${p(f.h, 'hora', 'horas')} y ${p(f.m, 'minuto', 'minutos')}`;
	return `Faltan ${p(Math.max(1, f.m), 'minuto', 'minutos')}`;
}
