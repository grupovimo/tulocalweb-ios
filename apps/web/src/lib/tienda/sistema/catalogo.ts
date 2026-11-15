/**
 * El catálogo de bloques de la tienda (docs/especificaciones/tienda-bloques.md): tipos y validación
 * armados desde packages/contrato/bloques.json, el mismo archivo que valida Go (mitienda/bloques.go)
 * y que leen las apps. Agregar un bloque o un campo es tocar ese JSON: acá no se escribe a mano
 * ningún límite.
 */
import * as v from 'valibot';
import CATALOGO from '@mostrador/contrato/bloques.json';
import type { EstiloTienda } from '../estilo/estilo.ts';

export type TipoCampo =
	| 'texto'
	| 'textoLargo'
	| 'foto'
	| 'fotos'
	| 'link'
	| 'boton'
	| 'fondo'
	| 'fecha'
	| 'numero'
	| 'si'
	| 'opcion'
	| 'icono'
	| 'video'
	| 'producto'
	| 'elegidos'
	| 'productos'
	| 'categoria'
	| 'categorias'
	| 'lista'
	| 'grupo';

export type Campo = {
	id: string;
	tipo: TipoCampo;
	nombre: string;
	max?: number;
	min?: number;
	requerido?: boolean;
	porDefecto?: unknown;
	/** Lo que toma si está vacío: «nombre» o «lema» de la tienda (lo resuelve datos.ts). */
	sugerido?: 'nombre' | 'lema';
	opciones?: string[];
	item?: Campo[];
};

export type DefBloque = {
	tipo: TipoBloque;
	grupo: string;
	nombre: string;
	descripcion: string;
	max?: number;
	variantes: string[];
	nombresVariantes: string[];
	campos: Campo[];
};

export const TIPOS = [
	'portada',
	'banner',
	'oferta',
	'textoImagen',
	'video',
	'beneficios',
	'productos',
	'destacado',
	'combo',
	'categorias',
	'menu',
	'local',
	'resenas',
	'preguntas',
	'logos',
	'galeria',
	'newsletter',
	'whatsapp',
	'turnos'
] as const;
export type TipoBloque = (typeof TIPOS)[number];

export const LIMITES = CATALOGO.limites;
export const BLOQUES = CATALOGO.bloques as unknown as DefBloque[];
export const DEF: Record<TipoBloque, DefBloque> = Object.fromEntries(
	BLOQUES.map((b) => [b.tipo, b])
) as Record<TipoBloque, DefBloque>;
export const FONDOS = CATALOGO.fondos.opciones as Fondo[];
export const FUENTES_PRODUCTOS = CATALOGO.fuentesProductos.opciones as FuenteProductos[];
export const ICONOS = CATALOGO.iconos as Icono[];
export const OPCIONES_ESTILO = CATALOGO.estilo;

export type Fondo = 'normal' | 'superficie' | 'acento' | 'oscuro';
export type FuenteProductos = 'vendidos' | 'nuevos' | 'ofertas' | 'elegidos' | 'categoria';
export type Icono =
	| 'envio'
	| 'cuotas'
	| 'transferencia'
	| 'cambios'
	| 'retiro'
	| 'seguro'
	| 'regalo'
	| 'hecho-a-mano'
	| 'reloj'
	| 'estrella';
export type Direccion = 'vidriera' | 'mostrador' | 'pedido';
export type BotonDatos = { t: string; a: string };
export type SeleccionProductos = {
	fuente: FuenteProductos;
	cantidad: number;
	ids?: string[];
	categoria?: string;
};

export type Bloque = {
	/** [a-z0-9]{6,16}, único en el diseño (lo genera el editor; los migrados usan el tipo). */
	id: string;
	tipo: TipoBloque;
	variante: string;
	oculto?: boolean;
	/** Los campos del tipo; lo que falta toma porDefecto o lo sugerido. */
	datos: Record<string, unknown>;
};

/** Lo que el diseño v2 suma al estilo de Mi tienda (los campos ya están en EstiloTienda). */
export type EstiloV2 = EstiloTienda;

export type PaginasDiseno = {
	catalogo?: { variante: 'grilla' | 'lista'; filtros: 'costado' | 'arriba'; columnas: number };
	producto?: {
		variante: 'grilla' | 'miniaturas' | 'carrusel';
		relacionados: boolean;
		calcularEnvio: boolean;
		bloques: Bloque[];
	};
	carrito?: { variante: 'pagina' | 'panel'; envioGratisDesde: boolean };
	nosotros?: { bloques: Bloque[] };
};

/** El diseño de la tienda, versión 2: la portada en bloques y el estilo ampliado. */
export type DisenoV2 = {
	version: 2;
	plantilla?: string;
	paleta: number;
	titulo?: string;
	lema?: string;
	portada?: string;
	logo?: string;
	compartirTitulo?: string;
	compartirDescripcion?: string;
	compartirFoto?: string;
	estilo?: EstiloV2;
	inicio: Bloque[];
	paginas?: PaginasDiseno;
};

// ---------------------------------------------------------------------------------------------
// Validación (valibot), generada campo por campo desde el catálogo.

const URL_HTTPS = /^https:\/\/[^\s"'<>]{3,}$/;
const LINK = /^(inicio|catalogo|nosotros|turnos|suscripciones|cat:[^<>\n]{1,60}|prod:[0-9A-Za-z-]{1,64}|https:\/\/[^\s"'<>]{3,})$/;
const VIDEO =
	/^https:\/\/(youtube\.com|www\.youtube\.com|youtu\.be|vimeo\.com|www\.vimeo\.com|player\.vimeo\.com)\/[^\s"'<>]+$/;

// Igual que Go (mitienda/bloques.go): sin < ni >, y los textos de una línea sin saltos.
const texto = (max = 200, largo = false) =>
	v.pipe(v.string(), v.maxLength(max), v.regex(largo ? /^[^<>]*$/ : /^[^<>\r\n]*$/));
const lleno = (s: v.GenericSchema<string, string>, requerido?: boolean) =>
	requerido ? v.pipe(s, v.minLength(1)) : s;
const foto = v.pipe(v.string(), v.maxLength(300), v.regex(URL_HTTPS));
const link = v.pipe(v.string(), v.maxLength(300), v.regex(LINK));
const id = v.pipe(v.string(), v.regex(/^[0-9A-Za-z-]{1,64}$/));

function esquemaCampo(c: Campo): v.GenericSchema {
	switch (c.tipo) {
		case 'texto':
			return lleno(texto(c.max), c.requerido);
		case 'textoLargo':
			return lleno(texto(c.max, true), c.requerido);
		case 'foto':
			return foto;
		case 'fotos':
			return v.pipe(v.array(foto), v.maxLength(c.max ?? LIMITES.itemsPorLista));
		case 'link':
			return link;
		case 'boton':
			return v.strictObject({ t: v.pipe(texto(30), v.minLength(1)), a: link });
		case 'fondo':
			return v.picklist(FONDOS);
		case 'fecha':
			return v.pipe(v.string(), v.isoTimestamp());
		case 'numero':
			return v.pipe(
				v.number(),
				v.integer(),
				v.minValue(c.min ?? 0),
				v.maxValue(c.max ?? Number.MAX_SAFE_INTEGER)
			);
		case 'si':
			return v.boolean();
		case 'opcion':
			return v.picklist(c.opciones ?? []);
		case 'icono':
			return v.picklist(ICONOS);
		case 'video':
			return v.pipe(v.string(), v.maxLength(300), v.regex(VIDEO));
		case 'producto':
			return id;
		case 'categoria':
			return lleno(texto(60), true);
		case 'elegidos':
			return v.pipe(v.array(id), v.minLength(c.min ?? 0), v.maxLength(c.max ?? 24));
		case 'categorias':
			return v.pipe(v.array(texto(60)), v.maxLength(LIMITES.itemsPorLista));
		case 'productos':
			return v.strictObject({
				fuente: v.picklist(FUENTES_PRODUCTOS),
				cantidad: v.pipe(v.number(), v.integer(), v.minValue(1), v.maxValue(24)),
				ids: v.optional(v.pipe(v.array(id), v.maxLength(24))),
				categoria: v.optional(texto(60))
			});
		case 'lista':
			return v.pipe(
				v.array(objeto(c.item ?? [])),
				v.minLength(c.min ?? 0),
				v.maxLength(c.max ?? LIMITES.itemsPorLista)
			);
		case 'grupo':
			return objeto(c.item ?? []);
	}
}

/**
 * Un objeto con esos campos (los desconocidos se rechazan): los requeridos tienen que venir, y no
 * vacíos; los demás son opcionales.
 */
function objeto(campos: Campo[]) {
	return v.strictObject(
		Object.fromEntries(
			campos.map((c) => [c.id, c.requerido ? esquemaCampo(c) : v.optional(esquemaCampo(c))])
		)
	);
}

/** Los datos de cada tipo: en el diseño guardado todo es opcional (el editor dice qué falta). */
const DATOS: Record<TipoBloque, v.GenericSchema> = Object.fromEntries(
	BLOQUES.map((b) => [
		b.tipo,
		v.strictObject(Object.fromEntries(b.campos.map((c) => [c.id, v.optional(esquemaCampo(c))])))
	])
) as unknown as Record<TipoBloque, v.GenericSchema>;

const ID_BLOQUE = /^[a-z0-9]{6,16}$/;

/** Un bloque válido (o null): tipo y variante del catálogo, datos con sus límites. */
export function validarBloque(x: unknown): Bloque | null {
	if (!x || typeof x !== 'object') return null;
	const b = x as Record<string, unknown>;
	const tipo = b.tipo as TipoBloque;
	const def = DEF[tipo];
	if (!def || typeof b.id !== 'string' || !ID_BLOQUE.test(b.id)) return null;
	if (typeof b.variante !== 'string' || !def.variantes.includes(b.variante)) return null;
	const datos = v.safeParse(DATOS[tipo], b.datos ?? {});
	if (!datos.success) return null;
	return {
		id: b.id,
		tipo,
		variante: b.variante,
		...(b.oculto === true ? { oculto: true } : {}),
		datos: datos.output as Record<string, unknown>
	};
}

/**
 * Los bloques que se pueden dibujar de una lista: los inválidos se descartan (la tienda nunca se
 * rompe por un dato raro), los ids repetidos también, y se respeta el tope por página y el máximo de
 * cada tipo (una sola portada).
 */
export function validarBloques(lista: unknown): Bloque[] {
	if (!Array.isArray(lista)) return [];
	const ids = new Set<string>();
	const cuenta = new Map<TipoBloque, number>();
	const out: Bloque[] = [];
	for (const x of lista) {
		const b = validarBloque(x);
		if (!b || ids.has(b.id)) continue;
		const n = (cuenta.get(b.tipo) ?? 0) + 1;
		if (DEF[b.tipo].max && n > DEF[b.tipo].max!) continue;
		ids.add(b.id);
		cuenta.set(b.tipo, n);
		out.push(b);
		if (out.length >= LIMITES.bloquesPorPagina) break;
	}
	return out;
}

/** El valor de un campo del bloque, o su valor por defecto del catálogo. */
export function campo<T = unknown>(b: Pick<Bloque, 'tipo' | 'datos'>, idCampo: string): T | undefined {
	const x = b.datos[idCampo];
	if (x !== undefined && x !== null && x !== '') return x as T;
	return DEF[b.tipo]?.campos.find((c) => c.id === idCampo)?.porDefecto as T | undefined;
}

/** Un texto del bloque recortado (vacío: el por defecto del catálogo, o ''). */
export function textoDe(b: Pick<Bloque, 'tipo' | 'datos'>, idCampo: string): string {
	const x = campo<string>(b, idCampo);
	return typeof x === 'string' ? x.trim() : '';
}
