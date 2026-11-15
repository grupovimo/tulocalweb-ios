/**
 * Las ventanas (popups) y las herramientas de conversión de la tienda
 * (docs/especificaciones/tienda-promociones.md), desde packages/contrato/bloques.json › ventanas y
 * › conversion: tipos, validación (la misma de Go) y si una ventana está completa para salir. Acá no
 * se escribe a mano ningún límite ni ninguna opción.
 */
import * as v from 'valibot';
import CATALOGO from '@mostrador/contrato/bloques.json';
import type { ConversionGuardada, EstiloTienda } from '../estilo/estilo.ts';
import { esquemaCampo, LIMITES, type Campo, type TipoPremio } from './catalogo.ts';

export type VarianteVentana = 'centrada' | 'lateral' | 'completa' | 'inferior';
export type ContenidoVentana = 'promo' | 'cupon' | 'captacion' | 'ruleta';
export type Disparo = 'entrar' | 'segundos' | 'scroll' | 'salida' | 'carrito';
export type Frecuencia = 'sesion' | 'dias' | 'siempre' | 'unaVez';
export type Publico = 'todos' | 'nuevos' | 'recurrentes';
export type PaginaVentana = 'inicio' | 'catalogo' | 'producto' | 'carrito' | 'nosotros';
export type Dispositivos = 'todos' | 'celular' | 'compu';

export type Premio = {
	tipo: TipoPremio;
	cupon?: string;
	descuento?: 'porcentaje' | 'monto' | 'envioGratis';
	valor?: number;
	dias?: number;
	archivo?: string;
	nombreArchivo?: string;
};

export type Ventana = {
	/** [a-z0-9]{6,16}, único. */
	id: string;
	/** Interno, para el panel. */
	nombre: string;
	activa: boolean;
	variante: VarianteVentana;
	contenido: ContenidoVentana;
	/** Los campos del contenido (catálogo). */
	datos: Record<string, unknown>;
	disparo: { cuando: Disparo; valor?: number };
	frecuencia: { cada: Frecuencia; dias?: number };
	publico: Publico;
	/** Vacío: en todas. */
	paginas: PaginaVentana[];
	dispositivos: Dispositivos;
	desde?: string;
	hasta?: string;
};

type Rango = { min: number; max: number; porDefecto: number; unidad?: string };
export type DefContenido = {
	tipo: ContenidoVentana;
	nombre: string;
	descripcion: string;
	campos: Campo[];
};
export type DefDisparo = { cuando: Disparo; nombre: string; descripcion: string; valor?: Rango };
export type DefFrecuencia = { cada: Frecuencia; nombre: string; descripcion: string; dias?: Rango };
export type OpcionesConNombres<T extends string> = {
	opciones: T[];
	nombres: string[];
	porDefecto?: T;
	descripcion?: string;
};
export type PlantillaVentana = {
	id: string;
	nombre: string;
	descripcion: string;
	/** Los campos que la persona tiene que elegir (un cupón suyo, el archivo del premio). */
	completar?: string[];
	/** datos.hasta = ahora + estos días. */
	hastaEnDias?: number;
	ventana: Omit<Ventana, 'id'>;
};

export const VENTANAS = CATALOGO.ventanas as unknown as {
	nombre: string;
	descripcion: string;
	variantes: VarianteVentana[];
	nombresVariantes: string[];
	descripcionesVariantes: string[];
	campos: Campo[];
	contenidos: DefContenido[];
	disparos: DefDisparo[];
	frecuencias: DefFrecuencia[];
	publicos: OpcionesConNombres<Publico>;
	paginas: OpcionesConNombres<PaginaVentana>;
	dispositivos: OpcionesConNombres<Dispositivos>;
	plantillas: PlantillaVentana[];
};
export const MAX_VENTANAS: number = (LIMITES as { ventanas?: number }).ventanas ?? 8;
export const CONTENIDO = Object.fromEntries(VENTANAS.contenidos.map((c) => [c.tipo, c])) as Record<
	ContenidoVentana,
	DefContenido
>;
export const DISPARO = Object.fromEntries(VENTANAS.disparos.map((d) => [d.cuando, d])) as Record<
	Disparo,
	DefDisparo
>;
export const FRECUENCIA = Object.fromEntries(
	VENTANAS.frecuencias.map((f) => [f.cada, f])
) as Record<Frecuencia, DefFrecuencia>;

const ID = /^[a-z0-9]{6,16}$/;
const CLAVES = [
	'id',
	'nombre',
	'activa',
	'variante',
	'contenido',
	'datos',
	'disparo',
	'frecuencia',
	'publico',
	'paginas',
	'dispositivos',
	'desde',
	'hasta'
];
const ISO = v.pipe(v.string(), v.isoTimestamp());
const nombreVentana = VENTANAS.campos.find((c) => c.id === 'nombre')!;

const DATOS: Record<ContenidoVentana, v.GenericSchema> = Object.fromEntries(
	VENTANAS.contenidos.map((c) => [
		c.tipo,
		v.strictObject(Object.fromEntries(c.campos.map((x) => [x.id, v.optional(esquemaCampo(x))])))
	])
) as unknown as Record<ContenidoVentana, v.GenericSchema>;

const enRango = (x: unknown, r: Rango | undefined) =>
	typeof x === 'number' && Number.isInteger(x) && !!r && x >= r.min && x <= r.max;

/**
 * Una ventana válida (o null): variante, contenido, disparo, frecuencia, público, páginas y
 * dispositivos del catálogo, y los datos del contenido con sus límites. Puede estar incompleta
 * (el borrador se guarda a medias): si sale lo decide `completa`.
 */
export function validarVentana(x: unknown): Ventana | null {
	if (!x || typeof x !== 'object') return null;
	const w = x as Record<string, unknown>;
	if (Object.keys(w).some((k) => !CLAVES.includes(k))) return null;
	if (typeof w.id !== 'string' || !ID.test(w.id)) return null;
	if (!v.safeParse(esquemaCampo(nombreVentana), w.nombre).success || !w.nombre) return null;
	if (typeof w.activa !== 'boolean') return null;
	if (!VENTANAS.variantes.includes(w.variante as VarianteVentana)) return null;
	const contenido = w.contenido as ContenidoVentana;
	if (!CONTENIDO[contenido]) return null;
	const datos = v.safeParse(DATOS[contenido], w.datos ?? {});
	if (!datos.success) return null;
	const d = (w.disparo ?? {}) as { cuando?: unknown; valor?: unknown };
	const dd = DISPARO[d.cuando as Disparo];
	if (!dd || (d.valor !== undefined && (!dd.valor || !enRango(d.valor, dd.valor)))) return null;
	const f = (w.frecuencia ?? {}) as { cada?: unknown; dias?: unknown };
	const ff = FRECUENCIA[f.cada as Frecuencia];
	if (!ff || (f.dias !== undefined && (!ff.dias || !enRango(f.dias, ff.dias)))) return null;
	if (!VENTANAS.publicos.opciones.includes(w.publico as Publico)) return null;
	if (!VENTANAS.dispositivos.opciones.includes(w.dispositivos as Dispositivos)) return null;
	const paginas = w.paginas ?? [];
	if (
		!Array.isArray(paginas) ||
		new Set(paginas).size !== paginas.length ||
		!paginas.every((p) => VENTANAS.paginas.opciones.includes(p))
	)
		return null;
	for (const k of ['desde', 'hasta'] as const)
		if (w[k] !== undefined && !v.safeParse(ISO, w[k]).success) return null;
	if (w.desde && w.hasta && Date.parse(w.desde as string) >= Date.parse(w.hasta as string))
		return null;
	return {
		id: w.id,
		nombre: w.nombre as string,
		activa: w.activa,
		variante: w.variante as VarianteVentana,
		contenido,
		datos: datos.output as Record<string, unknown>,
		disparo: {
			cuando: dd.cuando,
			...(dd.valor ? { valor: (d.valor as number) ?? dd.valor.porDefecto } : {})
		},
		frecuencia: {
			cada: ff.cada,
			...(ff.dias ? { dias: (f.dias as number) ?? ff.dias.porDefecto } : {})
		},
		publico: w.publico as Publico,
		paginas: paginas as PaginaVentana[],
		dispositivos: w.dispositivos as Dispositivos,
		...(w.desde ? { desde: w.desde as string } : {}),
		...(w.hasta ? { hasta: w.hasta as string } : {})
	};
}

/** Las ventanas que valen de una lista: sin las inválidas, sin ids repetidos y hasta el tope. */
export function validarVentanas(lista: unknown): Ventana[] {
	if (!Array.isArray(lista)) return [];
	const ids = new Set<string>();
	const out: Ventana[] = [];
	for (const x of lista) {
		const w = validarVentana(x);
		if (!w || ids.has(w.id)) continue;
		ids.add(w.id);
		out.push(w);
		if (out.length >= MAX_VENTANAS) break;
	}
	return out;
}

/** Como Go: la lista entera vale (todas válidas, ids sin repetir, hasta el tope). */
export function ventanasValidas(lista: unknown): boolean {
	if (!Array.isArray(lista) || lista.length > MAX_VENTANAS) return false;
	const ids = new Set<string>();
	for (const x of lista) {
		const w = validarVentana(x);
		if (!w || ids.has(w.id)) return false;
		ids.add(w.id);
	}
	return true;
}

const lleno = (x: unknown) =>
	x !== undefined && x !== null && x !== '' && !(Array.isArray(x) && !x.length);

/** Si un premio tiene lo suyo (cuponFijo con cupón, cuponUnico con descuento y valor…). */
export function premioCompleto(p: unknown): boolean {
	if (!p || typeof p !== 'object') return false;
	const x = p as Premio;
	switch (x.tipo) {
		case 'nada':
			return true;
		case 'cuponFijo':
			return !!x.cupon;
		case 'cuponUnico':
			return x.descuento === 'envioGratis' || (!!x.descuento && typeof x.valor === 'number');
		case 'descarga':
			return !!x.archivo;
		default:
			return false;
	}
}

/**
 * Lo que le falta a una ventana para salir en la tienda (null: está completa): los campos
 * requeridos del contenido, el premio con lo suyo y, en la ruleta, de 4 a 8 porciones con al menos
 * una con peso mayor a 0. Lo usan la tienda (no dibuja lo incompleto) y el editor (lo dice).
 */
export function queFaltaVentana(w: Ventana): string | null {
	const def = CONTENIDO[w.contenido];
	for (const c of def.campos) {
		// Lo requerido tiene que estar cargado (el porDefecto de un requerido no alcanza).
		if (c.requerido && !lleno(w.datos[c.id])) return `Falta ${c.nombre.toLowerCase()}.`;
		const x = w.datos[c.id] ?? c.porDefecto;
		if (c.tipo === 'premio' && lleno(x) && !premioCompleto(x)) return 'Completá el premio.';
		if (c.tipo === 'lista' && c.requerido && Array.isArray(x)) {
			if (x.length < (c.min ?? 0)) return `Van al menos ${c.min} ${c.nombre.toLowerCase()}.`;
			const items = x as Record<string, unknown>[];
			if (items.some((i) => (c.item ?? []).some((ic) => ic.requerido && !lleno(i[ic.id]))))
				return `Completá ${c.nombre.toLowerCase()}.`;
			if (items.some((i) => i.premio !== undefined && !premioCompleto(i.premio)))
				return 'Completá el premio de cada porción.';
			if (w.contenido === 'ruleta' && !items.some((i) => Number(i.peso) > 0))
				return 'Al menos una porción tiene que tener peso.';
		}
	}
	return null;
}

export const completa = (w: Ventana) => queFaltaVentana(w) === null;

/** El valor de un campo de la ventana o su porDefecto del catálogo. */
export function datoVentana<T = unknown>(w: Pick<Ventana, 'contenido' | 'datos'>, id: string) {
	const x = w.datos[id];
	if (lleno(x)) return x as T;
	return CONTENIDO[w.contenido]?.campos.find((c) => c.id === id)?.porDefecto as T | undefined;
}

// ---------------------------------------------------------------------------------------------
// Herramientas de conversión (estilo.conversion)

export type Conversion = {
	envioGratis: boolean;
	pruebaSocial: { activa: boolean; cada: number; paginas: PaginaVentana[] };
	etiquetas: { nuevo: boolean; oferta: boolean; ultimas: boolean; masVendido: boolean };
	stockBajo: number;
	compraFija: boolean;
	sumaEnCarrito: boolean;
	favoritos: boolean;
};

export const CONVERSION = CATALOGO.conversion as unknown as {
	nombre: string;
	descripcion: string;
	campos: Campo[];
};

/**
 * La conversión de la tienda: cada campo del catálogo con lo guardado si vale, o su porDefecto (en
 * los grupos, campo por campo). Lo que no vale se descarta en vez de romper la tienda.
 */
export function conversionDe(e: Pick<EstiloTienda, 'conversion'> | null | undefined): Conversion {
	const g = (e?.conversion ?? {}) as Record<string, unknown>;
	const r: Record<string, unknown> = {};
	for (const c of CONVERSION.campos) {
		if (c.tipo === 'grupo') {
			const x = (g[c.id] ?? {}) as Record<string, unknown>;
			const pd = (c.porDefecto ?? {}) as Record<string, unknown>;
			r[c.id] = Object.fromEntries(
				(c.item ?? []).map((i) => [
					i.id,
					x[i.id] !== undefined && v.safeParse(esquemaCampo(i), x[i.id]).success
						? x[i.id]
						: (i.porDefecto ?? pd[i.id])
				])
			);
		} else
			r[c.id] =
				g[c.id] !== undefined && v.safeParse(esquemaCampo(c), g[c.id]).success
					? g[c.id]
					: c.porDefecto;
	}
	return r as Conversion;
}

/** Como Go: estilo.conversion vale entera (sin claves desconocidas, cada campo en su rango). */
export function conversionValida(x: unknown): boolean {
	if (!x || typeof x !== 'object' || Array.isArray(x)) return false;
	const g = x as Record<string, unknown>;
	for (const k of Object.keys(g)) {
		const c = CONVERSION.campos.find((y) => y.id === k);
		if (!c) return false;
		if (c.tipo === 'grupo') {
			const o = g[k];
			if (!o || typeof o !== 'object' || Array.isArray(o)) return false;
			for (const [ik, iv] of Object.entries(o)) {
				const i = (c.item ?? []).find((y) => y.id === ik);
				if (!i || !v.safeParse(esquemaCampo(i), iv).success) return false;
			}
		} else if (!v.safeParse(esquemaCampo(c), g[k]).success) return false;
	}
	return true;
}

/** estilo.conversion lista para guardar: sólo los campos del catálogo que valen. */
export function conversionParaGuardar(x: unknown): ConversionGuardada | undefined {
	if (!x || typeof x !== 'object') return undefined;
	const g = x as Record<string, unknown>;
	const r: Record<string, unknown> = {};
	for (const c of CONVERSION.campos) {
		if (g[c.id] === undefined) continue;
		if (c.tipo === 'grupo') {
			const o = (g[c.id] ?? {}) as Record<string, unknown>;
			const s = Object.fromEntries(
				(c.item ?? [])
					.filter((i) => o[i.id] !== undefined && v.safeParse(esquemaCampo(i), o[i.id]).success)
					.map((i) => [i.id, o[i.id]])
			);
			if (Object.keys(s).length) r[c.id] = s;
		} else if (v.safeParse(esquemaCampo(c), g[c.id]).success) r[c.id] = g[c.id];
	}
	return Object.keys(r).length ? (r as ConversionGuardada) : undefined;
}
