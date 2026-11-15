/**
 * El diseño tal como se guarda: el editor deja a medio hacer lo que la persona está escribiendo
 * (una pregunta sin respuesta, un botón sin texto), pero el PUT (la web y Go) sólo acepta lo
 * completo. Acá se saca lo incompleto o inválido y se dejan sólo los campos del catálogo, así
 * nunca da 400 por algo que el editor todavía muestra.
 */
import { defDe, LIMITES, PAGINAS, type Campo } from './catalogo.ts';
import { URL_VIDEO } from './campos.ts';
import { bloquesDe, type Bloque } from './bloques.ts';
import type { DisenoV2, EstiloV2 } from './diseno.ts';

const LINK =
	/^(inicio|catalogo|nosotros|turnos|suscripciones|cat:.{1,60}|prod:[0-9A-Za-z-]{1,64}|https:\/\/[^\s"'<>]{3,})$/;
const ID_PRODUCTO = /^[0-9A-Za-z-]{1,64}$/;
// Las fotos subidas vuelven con https en producción; en local, con http.
const FOTO = /^https?:\/\/[^\s"'<>]{3,}$/;
const FECHA = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}(:\d{2}(\.\d+)?)?(Z|[+-]\d{2}:\d{2})$/;

const sinAngulos = (s: string) => s.replace(/[<>]/g, '');
/** Un texto de una línea: sin < ni >, sin saltos, recortado. */
export const linea = (s: unknown, max = 200) =>
	typeof s === 'string'
		? sinAngulos(s)
				.replace(/\s*[\r\n]+\s*/g, ' ')
				.trim()
				.slice(0, max)
		: '';
const largo = (s: unknown, max = 600) =>
	typeof s === 'string' ? sinAngulos(s).trim().slice(0, max) : '';
export const linkValido = (a: unknown): a is string => typeof a === 'string' && LINK.test(a);

/** El valor de un campo listo para guardar, o undefined si no va. */
export function valorCampo(c: Campo, v: unknown): unknown {
	switch (c.tipo) {
		case 'texto':
			return linea(v, c.max) || undefined;
		case 'textoLargo':
			return largo(v, c.max) || undefined;
		case 'foto':
			return typeof v === 'string' && FOTO.test(v) && v.length <= 300 ? v : undefined;
		case 'fotos': {
			const f = Array.isArray(v)
				? v.filter((x) => typeof x === 'string' && FOTO.test(x) && x.length <= 300)
				: [];
			return f.length ? f.slice(0, c.max ?? LIMITES.itemsPorLista) : undefined;
		}
		case 'link':
			return linkValido(v) ? v : undefined;
		case 'boton': {
			const b = (v ?? {}) as { t?: unknown; a?: unknown };
			const t = linea(b.t, 30);
			return t && linkValido(b.a) ? { t, a: b.a } : undefined;
		}
		case 'fondo':
		case 'opcion':
		case 'icono':
			return typeof v === 'string' && v ? v : undefined;
		case 'fecha':
			return typeof v === 'string' && FECHA.test(v) && Number.isFinite(Date.parse(v))
				? v
				: undefined;
		case 'numero': {
			if (typeof v !== 'number' || !Number.isFinite(v)) return undefined;
			const n = Math.round(v);
			return Math.min(c.max ?? n, Math.max(c.min ?? n, n));
		}
		case 'si':
			return typeof v === 'boolean' ? v : undefined;
		case 'video':
			return typeof v === 'string' && URL_VIDEO.test(v) ? v : undefined;
		case 'producto':
		case 'categoria':
			return typeof v === 'string' && v && (c.tipo === 'categoria' || ID_PRODUCTO.test(v))
				? v.slice(0, c.tipo === 'categoria' ? 60 : 64)
				: undefined;
		case 'elegidos': {
			const ids = Array.isArray(v)
				? [...new Set(v.filter((x) => typeof x === 'string' && ID_PRODUCTO.test(x)))]
				: [];
			const r = ids.slice(0, c.max ?? 24);
			return r.length && r.length >= (c.min ?? 0) ? r : undefined;
		}
		case 'categorias': {
			const r = Array.isArray(v)
				? v.filter((x) => typeof x === 'string' && x && x.length <= 60)
				: [];
			return r.length ? r.slice(0, LIMITES.itemsPorLista) : undefined;
		}
		case 'productos': {
			const p = { ...(c.porDefecto as object), ...(v as object) } as {
				fuente?: string;
				cantidad?: number;
				ids?: unknown;
				categoria?: unknown;
			};
			const r: Record<string, unknown> = {
				fuente: p.fuente || 'vendidos',
				cantidad: Math.min(24, Math.max(1, Math.round(Number(p.cantidad) || 8)))
			};
			if (r.fuente === 'elegidos') {
				const ids = valorCampo({ id: 'ids', tipo: 'elegidos', nombre: '' }, p.ids);
				if (ids) r.ids = ids;
			}
			if (r.fuente === 'categoria' && typeof p.categoria === 'string' && p.categoria)
				r.categoria = p.categoria.slice(0, 60);
			return r;
		}
		case 'lista': {
			const items = Array.isArray(v)
				? v.map((x) => objeto(c.item ?? [], x)).filter((x): x is Record<string, unknown> => !!x)
				: [];
			const r = items.slice(0, c.max ?? LIMITES.itemsPorLista);
			// Una lista con mínimo que no lo cumple no va (vacía, toma lo de la tienda).
			return r.length && r.length >= (c.min ?? 0) ? r : undefined;
		}
		case 'grupo': {
			const o = objeto(c.item ?? [], v, false);
			return o && Object.keys(o).length ? o : undefined;
		}
		case 'tabla':
			return tablaParaGuardar(v);
	}
}

/**
 * Una tabla lista para guardar: celdas en una línea y recortadas, todas las filas del ancho del
 * encabezado, sin filas vacías y con los topes del catálogo. Sin encabezado y al menos una fila, no va.
 */
export function tablaParaGuardar(v: unknown): string[][] | undefined {
	if (!Array.isArray(v)) return undefined;
	const filas = v
		.filter((f): f is unknown[] => Array.isArray(f))
		.map((f) => f.map((x) => linea(x, LIMITES.letrasCelda)));
	const ancho = Math.min(filas[0]?.length ?? 0, LIMITES.columnasTabla);
	if (!ancho) return undefined;
	const parejas = filas.map((f) => Array.from({ length: ancho }, (_, j) => f[j] ?? ''));
	const [cab, ...cuerpo] = parejas;
	const r = [cab, ...cuerpo.filter((f) => f.some(Boolean))].slice(0, LIMITES.filasTabla);
	return r.length >= 2 ? r : undefined;
}

/**
 * Un objeto con sólo esos campos. Con `exigir`, si falta un requerido no va (un ítem de lista
 * incompleto).
 */
function objeto(campos: Campo[], v: unknown, exigir = true): Record<string, unknown> | null {
	if (!v || typeof v !== 'object' || Array.isArray(v)) return null;
	const x = v as Record<string, unknown>;
	const r: Record<string, unknown> = {};
	for (const c of campos) {
		const val = valorCampo(c, x[c.id]);
		if (val === undefined) {
			if (exigir && c.requerido) return null;
			continue;
		}
		r[c.id] = val;
	}
	return r;
}

export function bloqueParaGuardar(b: Bloque): Bloque | null {
	const def = defDe(b.tipo);
	if (!def) return null;
	return {
		id: b.id,
		tipo: b.tipo,
		variante: def.variantes.includes(b.variante) ? b.variante : def.variantes[0],
		...(b.oculto ? { oculto: true } : {}),
		datos: objeto(def.campos, b.datos, false) ?? {}
	};
}

const lista = (l: Bloque[] | undefined) =>
	(l ?? [])
		.map(bloqueParaGuardar)
		.filter((b): b is Bloque => !!b)
		.slice(0, LIMITES.bloquesPorPagina);

function estiloParaGuardar(e: EstiloV2 | undefined): EstiloV2 | undefined {
	if (!e) return undefined;
	const r: EstiloV2 = { ...e };
	// v1: pasan a la barra y a los títulos de los bloques.
	delete r.anuncio;
	delete r.anuncioLink;
	delete r.textos;
	if (r.barra) {
		const mensajes = (r.barra.mensajes ?? [])
			.map((m) => {
				const texto = linea(m.texto, 100);
				return texto ? { texto, ...(linkValido(m.link) ? { link: m.link } : {}) } : null;
			})
			.filter((m) => !!m)
			.slice(0, 5);
		if (mensajes.length)
			r.barra = {
				variante: r.barra.variante === 'rotativa' ? 'rotativa' : 'fija',
				mensajes,
				...(r.barra.fondo ? { fondo: r.barra.fondo } : {})
			};
		else delete r.barra;
	}
	if (r.menu) {
		const menu = r.menu
			.map((i) => ({ t: linea(i.t, 30), a: i.a }))
			.filter((i) => i.t && linkValido(i.a))
			.slice(0, 8);
		if (menu.length) r.menu = menu;
		else delete r.menu;
	}
	if (r.leyendaPie !== undefined) r.leyendaPie = linea(r.leyendaPie, 160) || undefined;
	if (r.redes) {
		const redes = Object.fromEntries(
			Object.entries(r.redes)
				.map(([k, x]) => [k, linea(x, k === 'email' ? 120 : 100)])
				.filter(([, x]) => x)
		);
		r.redes = Object.keys(redes).length ? redes : undefined;
	}
	for (const k of Object.keys(r) as (keyof EstiloV2)[])
		if (r[k] === undefined || r[k] === '') delete r[k];
	return Object.keys(r).length ? r : undefined;
}

/** Las opciones de las páginas: sólo las del catálogo. */
function paginasParaGuardar(d: DisenoV2): DisenoV2['paginas'] {
	const p = d.paginas;
	if (!p) return undefined;
	const r: Record<string, unknown> = {};
	for (const id of ['catalogo', 'producto', 'carrito'] as const) {
		const x = p[id] as Record<string, unknown> | undefined;
		if (!x) continue;
		const def = PAGINAS[id];
		const o: Record<string, unknown> = {};
		if (typeof x.variante === 'string' && def.variantes?.includes(x.variante))
			o.variante = x.variante;
		for (const c of def.campos ?? []) {
			const v = valorCampo(c, x[c.id]);
			if (v !== undefined) o[c.id] = v;
		}
		if (id === 'producto') {
			const b = lista(p.producto?.bloques);
			if (b.length) o.bloques = b;
		}
		if (Object.keys(o).length) r[id] = o;
	}
	const nosotros = lista(p.nosotros?.bloques);
	if (nosotros.length) r.nosotros = { bloques: nosotros };
	return Object.keys(r).length ? (r as DisenoV2['paginas']) : undefined;
}

const MARCA = [
	['titulo', 80],
	['lema', 120],
	['compartirTitulo', 70],
	['compartirDescripcion', 160]
] as const;

export function paraGuardar(d: DisenoV2): DisenoV2 {
	const r: DisenoV2 = { version: 2, paleta: d.paleta, inicio: lista(d.inicio) };
	if (d.plantilla) r.plantilla = d.plantilla;
	for (const [k, max] of MARCA) {
		const t = linea(d[k], max);
		if (t) r[k] = t;
	}
	for (const k of ['portada', 'logo', 'compartirFoto'] as const)
		if (d[k] && FOTO.test(d[k]!)) r[k] = d[k];
	const estilo = estiloParaGuardar(d.estilo);
	if (estilo) r.estilo = estilo;
	const paginas = paginasParaGuardar(d);
	if (paginas) r.paginas = paginas;
	return r;
}

const ID_BLOQUE = /^[a-z0-9]{6,16}$/;
/** Lo que no se arregla limpiando: ids repetidos o un bloque que va una sola vez. */
export function problema(d: DisenoV2): string | null {
	const ids = new Set<string>();
	for (const p of ['inicio', 'producto', 'nosotros'] as const) {
		const cuenta = new Map<string, number>();
		for (const b of bloquesDe(d, p)) {
			if (!ID_BLOQUE.test(b.id) || ids.has(b.id)) return 'Hay dos bloques con el mismo id.';
			ids.add(b.id);
			const def = defDe(b.tipo);
			if (!def) return 'Uno de los bloques ya no existe: borralo.';
			const n = (cuenta.get(b.tipo) ?? 0) + 1;
			cuenta.set(b.tipo, n);
			if (def.max && n > def.max) return `${def.nombre} va una sola vez por página.`;
		}
	}
	return null;
}
