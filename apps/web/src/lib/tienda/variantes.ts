/**
 * Variantes de un producto (Talle M · Negro): lógica pura que comparten la tienda, el carrito, el
 * panel y el sandbox. Un producto sin opciones no tiene variantes y se vende como siempre.
 */
import type { OpcionProducto } from './tipos.ts';

export const MAX_OPCIONES = 3;
export const MAX_VALORES = 20;
export const MAX_FOTOS = 8;
/** Lo mismo que admite Mercado Libre por publicación (y la API Go). */
export const MAX_VARIANTES = 100;

export type Valores = Record<string, string>;
type ConValores = { valores: Valores };
type Vendible = ConValores & { precio: number | null; stock: number };

/** Lo que manda el editor por cada combinación: se identifica por sus valores (el id no cuenta). */
export type VarianteEditable = {
	valores: Valores;
	sku?: string;
	precio?: number | null;
	/** null: el mayorista del producto (§49). */
	precioMayorista?: number | null;
	stock?: number;
	foto?: string | null;
	activa?: boolean;
};

/** La variante completa del panel y la API. */
export type Variante = {
	id: string;
	valores: Valores;
	sku: string;
	precio: number | null;
	/** null (o sin venir): el mayorista del producto (§49). */
	precioMayorista?: number | null;
	stock: number;
	foto: string | null;
	activa: boolean;
};

const coincide = (v: ConValores, elegidos: Valores, salvo?: string) =>
	Object.entries(elegidos).every(([k, x]) => k === salvo || v.valores[k] === x);

/**
 * El texto legible de una variante, como lo guarda el pedido: "Talle M · Color Negro" (nombre y
 * valor de cada opción, en su orden). Es la misma regla que la API Go.
 */
export function textoVariante(opciones: OpcionProducto[], valores: Valores) {
	return opciones
		.filter((o) => valores[o.nombre])
		.map((o) => `${o.nombre} ${valores[o.nombre]}`)
		.join(' · ');
}

/** Para comparar nombres y valores repetidos: sin mayúsculas ni tildes. */
export const normal = (s: string) => s.trim().normalize('NFD').replace(/\p{M}/gu, '').toLowerCase();

/** La variante de los valores elegidos (null si falta elegir alguna opción o no existe). */
export function elegirVariante<V extends ConValores>(
	opciones: OpcionProducto[],
	variantes: V[],
	elegidos: Valores
): V | null {
	if (!opciones.length || opciones.some((o) => !elegidos[o.nombre])) return null;
	return (
		variantes.find((v) => opciones.every((o) => v.valores[o.nombre] === elegidos[o.nombre])) ?? null
	);
}

/**
 * Para cada valor de una opción, si se puede comprar con lo demás que ya se eligió: hay una
 * variante con ese valor, los otros valores elegidos y stock.
 */
export function valoresDisponibles(
	opcion: OpcionProducto,
	variantes: Vendible[],
	elegidos: Valores
): Record<string, boolean> {
	return Object.fromEntries(
		opcion.valores.map((x) => [
			x,
			variantes.some(
				(v) => v.valores[opcion.nombre] === x && v.stock > 0 && coincide(v, elegidos, opcion.nombre)
			)
		])
	);
}

/** Precio de una variante: el suyo o, si no tiene, el del producto. */
export const precioDe = (v: { precio: number | null }, base: number) => v.precio ?? base;

/**
 * El precio que se muestra en la tarjeta: con variantes de precios distintos, "desde" el menor de
 * las que tienen stock (o de todas, si no queda ninguna).
 */
export function precioDesde(p: { precio: number; variantes: Vendible[] }) {
	if (!p.variantes.length) return { precio: p.precio, desde: false };
	const con = p.variantes.filter((v) => v.stock > 0);
	const precios = (con.length ? con : p.variantes).map((v) => precioDe(v, p.precio));
	const todos = new Set(p.variantes.map((v) => precioDe(v, p.precio)));
	return { precio: Math.min(...precios), desde: todos.size > 1 };
}

/** Sin stock: sin opciones, el del producto; con opciones, todas sus variantes. */
export const agotado = (p: { stock: number; opciones: OpcionProducto[]; variantes: Vendible[] }) =>
	p.opciones.length ? !p.variantes.some((v) => v.stock > 0) : p.stock <= 0;

/** Con opciones, el stock del producto es la suma de sus variantes activas. */
export const stockTotal = (variantes: { stock: number; activa?: boolean }[]) =>
	variantes.reduce((s, v) => s + (v.activa === false ? 0 : v.stock), 0);

// "una" para las opciones femeninas (Presentación, Medida); "un" para las demás.
const MASCULINAS_EN_A = new Set(['aroma', 'sistema', 'tema', 'idioma', 'clima', 'programa']);
function articulo(nombre: string) {
	const n = nombre.trim().toLowerCase();
	if (MASCULINAS_EN_A.has(n)) return 'un';
	return /(a|ión|ion|dad|tud|ez)$/.test(n) ? 'una' : 'un';
}

/** Lo que falta elegir para poder agregar: "Elegí un talle" (null si ya está todo). */
export function faltaElegir(opciones: OpcionProducto[], elegidos: Valores) {
	const o = opciones.find((x) => !elegidos[x.nombre]);
	return o ? `Elegí ${articulo(o.nombre)} ${o.nombre.trim().toLowerCase()}` : null;
}

/** Todas las combinaciones de valores, en el orden de las opciones. */
export function combinaciones(opciones: OpcionProducto[]): Valores[] {
	if (!opciones.length) return [];
	return opciones.reduce<Valores[]>(
		(acc, o) => acc.flatMap((v) => o.valores.map((x) => ({ ...v, [o.nombre]: x }))),
		[{}]
	);
}

const mismaCombinacion = (a: Valores, b: Valores, opciones: OpcionProducto[]) =>
	Object.keys(a).length === opciones.length && opciones.every((o) => a[o.nombre] === b[o.nombre]);

/**
 * Las variantes para unas opciones nuevas: las combinaciones que ya existían conservan su stock,
 * SKU, precio y foto; las nuevas arrancan con stock 0 y activas; las que dejan de existir se van.
 */
export function generarVariantes(
	opciones: OpcionProducto[],
	existentes: Variante[],
	nuevoId: (i: number) => string = () => ''
): Variante[] {
	return combinaciones(opciones).map(
		(valores, i) =>
			existentes.find((v) => mismaCombinacion(v.valores, valores, opciones)) ?? {
				id: nuevoId(i),
				valores,
				sku: '',
				precio: null,
				stock: 0,
				foto: null,
				activa: true
			}
	);
}

/** Revisa las opciones como las valida la API. Devuelve el error para mostrar, o null. */
export function errorOpciones(opciones: OpcionProducto[]): string | null {
	if (opciones.length > MAX_OPCIONES) return `Hasta ${MAX_OPCIONES} opciones por producto.`;
	const nombres = new Set<string>();
	for (const o of opciones) {
		const n = o.nombre.trim();
		if (!n || n.length > 20) return 'Cada opción necesita un nombre de hasta 20 letras.';
		if (nombres.has(normal(n))) return `La opción "${n}" está repetida.`;
		nombres.add(normal(n));
		if (!o.valores.length) return `Sumale al menos un valor a "${n}".`;
		if (o.valores.length > MAX_VALORES) return `Hasta ${MAX_VALORES} valores por opción.`;
		const vistos = new Set<string>();
		for (const v of o.valores) {
			const x = v.trim();
			if (!x || x.length > 20) return `Los valores de "${n}" van de 1 a 20 letras.`;
			if (vistos.has(normal(x))) return `"${x}" está repetido en "${n}".`;
			vistos.add(normal(x));
		}
	}
	if (combinaciones(opciones).length > MAX_VARIANTES)
		return `Son demasiadas combinaciones (hasta ${MAX_VARIANTES}).`;
	return null;
}

/** Revisa que las variantes correspondan a las opciones (una clave por opción, valores que existen). */
export function errorVariantes(
	opciones: OpcionProducto[],
	variantes: { valores: Valores; foto?: string | null }[],
	fotos: string[]
): string | null {
	if (!opciones.length) return variantes.length ? 'Sin opciones no hay variantes.' : null;
	const claves = new Set<string>();
	for (const v of variantes) {
		const ks = Object.keys(v.valores);
		if (
			ks.length !== opciones.length ||
			opciones.some((o) => !o.valores.includes(v.valores[o.nombre]))
		)
			return 'Hay una variante que no corresponde a las opciones.';
		const clave = opciones.map((o) => v.valores[o.nombre]).join('\u0000');
		if (claves.has(clave)) return 'Hay dos variantes con los mismos valores.';
		claves.add(clave);
		if (v.foto && !fotos.includes(v.foto))
			return 'La foto de una variante tiene que estar en la galería.';
	}
	return null;
}

/**
 * La foto de un valor (para swatches): sólo si todas las variantes con ese valor tienen la misma.
 * Así el color lleva su foto y el talle no (cada talle viene en varios colores).
 */
export function fotoDelValor(
	variantes: (ConValores & { foto: string | null })[],
	opcion: string,
	valor: string
) {
	const fotos = new Set(variantes.filter((v) => v.valores[opcion] === valor).map((v) => v.foto));
	const [f] = fotos;
	return fotos.size === 1 ? (f ?? null) : null;
}

const COLORES: Record<string, string> = {
	negro: '#2B2622',
	blanco: '#F7F5F0',
	crudo: '#EDE4D3',
	natural: '#D9C3A0',
	arena: '#D8C3A5',
	beige: '#D9C7AE',
	nude: '#E3BFA6',
	camel: '#B98B5E',
	suela: '#9A6236',
	marron: '#6B4A33',
	chocolate: '#4E3426',
	gris: '#8E8E8C',
	'gris melange': '#A7A6A3',
	'gris topo': '#7C7268',
	verde: '#6E7B6A',
	'verde oliva': '#6B6B3A',
	azul: '#2F4A6B',
	'azul marino': '#1F2A44',
	celeste: '#9CC3DE',
	rojo: '#A8322D',
	bordo: '#6B1F2A',
	rosa: '#E4A9B4',
	amarillo: '#E7C552',
	naranja: '#D9772F',
	violeta: '#6E4F8A',
	lila: '#B7A2CF',
	dorado: '#C9A44C',
	plateado: '#C0C0C0',
	carey: '#6A4128'
};

/** El color de un valor de la opción Color ("Negro", "Gris melange"), o null si no es conocido. */
export function colorDe(opcion: string, valor: string) {
	if (!/^colou?r(es)?$/.test(normal(opcion))) return null;
	return COLORES[normal(valor)] ?? null;
}
