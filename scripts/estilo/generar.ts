/**
 * Estilo de la tienda (Mi tienda › Colores y Tipografías): prepara las plantillas con archivo propio
 * para que la tienda cambie sus colores por rol y sus fuentes, sin tocar cómo se ven por defecto.
 *
 * Cada color de la plantilla pasa a `var(--tc-<clave>, <el original>)` y cada fuente a
 * `var(--tf-titulos|texto, <la original>)`. Si la tienda no cambió nada, las variables no existen y
 * la plantilla se ve igual que siempre. Para cada tono se calcula de qué dos roles sale (una mezcla
 * en OKLab, como `color-mix(in oklab, …)`), y eso queda en lib/tienda/estilo/mapas.ts: con los
 * colores elegidos, lib/tienda/estilo/css.ts arma los --tc-* de nuevo. Los tonos que no salen de
 * los roles (verde de «en stock», rojo de error, WhatsApp, Mercado Pago) quedan fijos.
 *
 * Se puede volver a correr (por ejemplo, después de sumar colores a una plantilla): lo que ya tiene
 * su variable no se envuelve de nuevo, pero entra igual al mapa.
 *   bun scripts/estilo/generar.ts
 */
import { readFileSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';

const RAIZ = join(import.meta.dir, '../..');
const DIR = 'apps/web/src/lib/tienda/plantillas';

export type Rol =
	| 'fondo'
	| 'superficie'
	| 'texto'
	| 'textoSuave'
	| 'borde'
	| 'acento'
	| 'sobreAcento';

type Def = {
	archivo: string;
	roles: Record<Rol, string>;
	/** Fondo y texto originales del pie (si es una banda propia). */
	pie?: { fondo: string; texto: string };
	fuentes: { titulos: string | null; texto: string };
	/** El fondo del encabezado cuando no sale del fondo de la página (Pampa: negro). */
	encabezado?: string;
	/** Los verdes de «en stock» y del precio con transferencia: rol «ok», que se ajusta solo. */
	ok?: string[];
};

// Los roles de cada plantilla, leídos de su diseño (scripts/estilo/analizar.ts lista sus colores).
const PLANTILLAS: Record<string, Def> = {
	andes: {
		archivo: 'Andes',
		ok: ['#14894B', '#2F7A43'],
		roles: { fondo: '#F3F5F8', superficie: '#FFFFFF', texto: '#0E1116', textoSuave: '#69717E', borde: '#E3E8EF', acento: '#2F6BFF', sobreAcento: '#FFFFFF' },
		pie: { fondo: '#0E1116', texto: '#FFFFFF' },
		fuentes: { titulos: 'Sora', texto: 'Instrument Sans' }
	},
	ceibo: {
		archivo: 'Ceibo',
		ok: ['#2F6B47', '#2F7A43'],
		roles: { fondo: '#F4EFEA', superficie: '#F8F4EF', texto: '#1C1A17', textoSuave: '#6E655C', borde: '#E2D9CF', acento: '#9C4A2F', sobreAcento: '#F4EFEA' },
		fuentes: { titulos: 'Instrument Serif', texto: 'Instrument Sans' }
	},
	malbec: {
		archivo: 'Malbec',
		ok: ['#9FD3A8', '#2F7D4F', '#2F7A43'],
		roles: { fondo: '#1E0E12', superficie: '#24121A', texto: '#F4E9E1', textoSuave: '#BFA9A3', borde: '#3A2026', acento: '#E8B86A', sobreAcento: '#1E0E12' },
		fuentes: { titulos: 'Cormorant Garamond', texto: 'Instrument Sans' }
	},
	miga: {
		archivo: 'Miga',
		ok: ['#2F7A43'],
		roles: { fondo: '#F4EADB', superficie: '#FFFFFF', texto: '#2A1C12', textoSuave: '#7A5F4C', borde: '#E2CDB6', acento: '#D9452B', sobreAcento: '#F4EADB' },
		pie: { fondo: '#2A1C12', texto: '#F4EADB' },
		fuentes: { titulos: 'Schibsted Grotesk', texto: 'Instrument Sans' }
	},
	'miga-clasica': {
		archivo: 'MigaClasica',
		ok: ['#2F7A43'],
		roles: { fondo: '#FBF1E6', superficie: '#FFFFFF', texto: '#3A2414', textoSuave: '#7A5F4C', borde: '#E2CDB6', acento: '#B5651D', sobreAcento: '#FBF1E6' },
		pie: { fondo: '#3A2414', texto: '#FBF1E6' },
		fuentes: { titulos: 'Young Serif', texto: 'Instrument Sans' }
	},
	mostaza: {
		archivo: 'Mostaza',
		ok: ['#2F6B3A', '#1F6F3A', '#2F7A43'],
		roles: { fondo: '#FBF3E4', superficie: '#FFFFFF', texto: '#3B2416', textoSuave: '#7A5A44', borde: '#E9D6B8', acento: '#C9821A', sobreAcento: '#2A170C' },
		pie: { fondo: '#3B2416', texto: '#FBF3E4' },
		fuentes: { titulos: 'Bricolage Grotesque', texto: 'Instrument Sans' }
	},
	pampa: {
		archivo: 'Pampa',
		encabezado: '#111111',
		ok: ['#B9F27C', '#2F7A43'],
		roles: { fondo: '#0B0B0B', superficie: '#161616', texto: '#F2F2F2', textoSuave: '#8A8A8A', borde: '#2A2A2A', acento: '#E6FF4D', sobreAcento: '#111111' },
		fuentes: { titulos: 'Archivo Black', texto: 'Archivo' }
	},
	quebracho: {
		archivo: 'Quebracho',
		ok: ['#2F7A43'],
		roles: { fondo: '#EFE9E1', superficie: '#F7F3EE', texto: '#2B2520', textoSuave: '#6F6358', borde: '#D8CCBC', acento: '#4F5B3A', sobreAcento: '#FFFFFF' },
		pie: { fondo: '#2B2520', texto: '#EFE9E1' },
		fuentes: { titulos: 'Newsreader', texto: 'Familjen Grotesk' }
	},
	rosedal: {
		archivo: 'Rosedal',
		ok: ['#2F7D55', '#2F7A43'],
		roles: { fondo: '#FBF4F1', superficie: '#FFFFFF', texto: '#3A1F25', textoSuave: '#8C6A70', borde: '#EAD3CF', acento: '#B4475B', sobreAcento: '#FFFFFF' },
		pie: { fondo: '#F3D9DE', texto: '#3A1F25' },
		fuentes: { titulos: 'Gloock', texto: 'Hanken Grotesk' }
	},
	tijera: {
		archivo: 'Tijera',
		ok: ['#2F7D4F', '#2F7A43'],
		roles: { fondo: '#F6F4F1', superficie: '#FFFFFF', texto: '#1E1B2E', textoSuave: '#5A566B', borde: '#E4E1EA', acento: '#5B3FD6', sobreAcento: '#FFFFFF' },
		pie: { fondo: '#1E1B2E', texto: '#F6F4F1' },
		fuentes: { titulos: 'Syne', texto: 'Instrument Sans' }
	},
	yerbal: {
		archivo: 'Yerbal',
		ok: ['#1F6F3A', '#2F7A43'],
		roles: { fondo: '#EEF1E8', superficie: '#FFFFFF', texto: '#1E2B22', textoSuave: '#6B7A6E', borde: '#DDE8DF', acento: '#2F5D50', sobreAcento: '#FFFFFF' },
		pie: { fondo: '#1E2B22', texto: '#EEF1E8' },
		fuentes: { titulos: null, texto: 'Schibsted Grotesk' }
	}
};

// Prioridad cuando un mismo color es dos roles (Miga: el texto sobre el acento es el fondo).
const ROLES: Rol[] = ['fondo', 'texto', 'acento', 'superficie', 'textoSuave', 'borde', 'sobreAcento'];

/** Un tono: mezcla de dos roles (p % del primero) y, si tenía transparencia, su alfa (0–1). */
export type Tono = { a: Rol | 'ok'; b: Rol | null; p: number; alfa: number | null };

// ── OKLab ────────────────────────────────────────────────────────────────────────────────────
type Lab = [number, number, number];
function rgb(hex: string): [number, number, number] {
	return [1, 3, 5].map((i) => parseInt(hex.slice(i, i + 2), 16) / 255) as [number, number, number];
}
function lab(hex: string): Lab {
	const [r, g, b] = rgb(hex).map((v) => (v <= 0.04045 ? v / 12.92 : ((v + 0.055) / 1.055) ** 2.4));
	const l = Math.cbrt(0.4122214708 * r + 0.5363325363 * g + 0.0514459929 * b);
	const m = Math.cbrt(0.2119034982 * r + 0.6806995451 * g + 0.1073969566 * b);
	const s = Math.cbrt(0.0883024619 * r + 0.2817188376 * g + 0.6299787005 * b);
	return [
		0.2104542553 * l + 0.793617785 * m - 0.0040720468 * s,
		1.9779984951 * l - 2.428592205 * m + 0.4505937099 * s,
		0.0259040371 * l + 0.7827717662 * m - 0.808675766 * s
	];
}
const dist = (x: Lab, y: Lab) => Math.hypot(x[0] - y[0], x[1] - y[1], x[2] - y[2]);

/** De qué roles sale un color (null: es un color propio, queda fijo). */
function tono(hex: string, roles: Record<Rol, string>): Omit<Tono, 'alfa'> | null {
	const c = lab(hex);
	for (const r of ROLES) if (roles[r].toUpperCase() === hex) return { a: r, b: null, p: 100 };
	let mejor: { a: Rol; b: Rol; p: number; e: number } | null = null;
	for (const a of ROLES)
		for (const b of ROLES) {
			if (a >= b) continue;
			const A = lab(roles[a]);
			const B = lab(roles[b]);
			const d: Lab = [A[0] - B[0], A[1] - B[1], A[2] - B[2]];
			const n = d[0] ** 2 + d[1] ** 2 + d[2] ** 2;
			if (n < 1e-6) continue;
			const t = ((c[0] - B[0]) * d[0] + (c[1] - B[1]) * d[1] + (c[2] - B[2]) * d[2]) / n;
			if (t < 0 || t > 1) continue;
			const e = dist(c, [B[0] + t * d[0], B[1] + t * d[1], B[2] + t * d[2]]);
			if (!mejor || e < mejor.e) mejor = { a, b, p: Math.round(t * 1000) / 10, e };
		}
	// 0,02 en OKLab es apenas perceptible: más lejos, el color no sale de los roles.
	return mejor && mejor.e < 0.02 ? { a: mejor.a, b: mejor.b, p: mejor.p } : null;
}

const hex6 = (h: string) =>
	(h.length === 4 ? '#' + [...h.slice(1)].map((c) => c + c).join('') : h).toUpperCase();
const aHex = (r: number, g: number, b: number) =>
	'#' + [r, g, b].map((v) => Math.round(v).toString(16).padStart(2, '0')).join('').toUpperCase();

const FIJOS = new Set(['#2A140C', '#FCEDE5']);

// Colores que no se tocan: el acento ya va por var(--acento, …) y los atributos de SVG no aceptan var().
const COLOR = /(var\(--acento,\s*)?(#[0-9a-fA-F]{6}\b|#[0-9a-fA-F]{3}\b|rgba?\(\s*(\d+)\s*,\s*(\d+)\s*,\s*(\d+)\s*(?:,\s*([\d.]+%?)\s*)?\))/g;
const ATRIBUTO = /(fill|stroke|stop-color|flood-color|color)="$/;

type Mapa = {
	roles: Record<Rol, string>;
	ok: string | null;
	encabezado: string | null;
	pie: { fondo: string; texto: string } | null;
	fuentes: { titulos: string | null; texto: string };
	tonos: Record<string, Tono>;
};

const mapas: Record<string, Mapa> = {};
for (const [id, def] of Object.entries(PLANTILLAS)) {
	const ruta = `${DIR}/${def.archivo}.svelte`;
	let src = readFileSync(join(RAIZ, ruta), 'utf8');
	const tonos: Record<string, Tono> = {};
	let cambios = 0;
	src = src.replace(COLOR, (todo, acento, color, r, g, b, a, pos: number, s: string) => {
		if (acento) return todo;
		if (ATRIBUTO.test(s.slice(Math.max(0, pos - 20), pos))) return todo;
		// En <link href> o URLs no es un color.
		if (/%23|href=|url\(/.test(s.slice(Math.max(0, pos - 12), pos))) return todo;
		let base: string;
		let alfa: number | null = null;
		if (r !== undefined) {
			base = aHex(+r, +g, +b);
			if (a !== undefined) alfa = a.endsWith('%') ? parseFloat(a) / 100 : parseFloat(a);
			if (alfa === 1) alfa = null;
		} else base = hex6(color);
		// La marca de Tu Local Web (la pastilla del pie) no cambia con la tienda.
		if (FIJOS.has(base)) return todo;
		const t = def.ok?.includes(base) ? { a: 'ok' as const, b: null, p: 100 } : tono(base, def.roles);
		if (!t) return todo;
		const clave = base.slice(1).toLowerCase() + (alfa === null ? '' : '-' + Math.round(alfa * 1000));
		tonos[clave] = { ...t, alfa };
		if (/var\(--tc-[\w-]+,\s*$/.test(s.slice(Math.max(0, pos - 30), pos))) return todo;
		cambios++;
		return `var(--tc-${clave}, ${color})`;
	});
	// Las fuentes: la de títulos y la de texto, cada una con su variable.
	const fuente = (nombre: string, v: string) => {
		const re = new RegExp(`(?<!var\\(--tf-[a-z]+,\\s*)'${nombre}'`, 'g');
		src = src.replace(re, `var(--tf-${v},'${nombre}')`);
	};
	if (def.fuentes.titulos) fuente(def.fuentes.titulos, 'titulos');
	fuente(def.fuentes.texto, 'texto');
	// El encabezado y el pie, para que tengan sus propios colores.
	if (!src.includes('data-t-enc')) src = src.replace(/<header(\s)/, '<header data-t-enc$1');
	if (!src.includes('data-t-pie')) src = src.replace(/<footer(\s)/, '<footer data-t-pie$1');
	writeFileSync(join(RAIZ, ruta), src);
	mapas[id] = { roles: def.roles, ok: def.ok?.[0] ?? null, encabezado: def.encabezado ?? null, pie: def.pie ?? null, fuentes: def.fuentes, tonos };
	console.log(`${id}: ${cambios} colores con variable, ${Object.keys(tonos).length} tonos`);
}

const salida = `// Generado por scripts/estilo/generar.ts: no editar a mano.
import type { MapaEstilo } from './tipos.ts';

export const MAPAS: Record<string, MapaEstilo> = ${JSON.stringify(mapas, null, '\t')};
`;
writeFileSync(join(RAIZ, 'apps/web/src/lib/tienda/estilo/mapas.ts'), salida);
