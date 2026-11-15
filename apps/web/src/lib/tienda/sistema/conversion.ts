/**
 * Las herramientas para vender más (estilo.conversion, tienda-promociones.md › 2) que se calculan
 * con lo que ya trae la tienda: las etiquetas automáticas de cada producto, el aviso de stock bajo y
 * «Sumá también» del carrito. Sin DOM: lo prueban conversion.spec.ts.
 */
import { proximamente } from '../lanzamiento.ts';
import { tarjetaDe } from '../seleccion.svelte.ts';
import type { ProductoPublico, TiendaPublica } from '../tipos.ts';
import { comboDelProducto } from '../combos.ts';
import { comboDe } from './datos.ts';
import { descuento } from './formato.ts';
import { disenoDeTienda } from './migrar.ts';
import { conversionDe, type Conversion } from './promos.ts';

export const conversionDeTienda = (t: TiendaPublica): Conversion =>
	conversionDe(disenoDeTienda(t).estilo);

export type Etiqueta = { texto: string; tipo: 'off' | 'propia' | 'nuevo' | 'ultimas' | 'top' };

const DIA = 86_400_000;
/** Los ids de los 3 más vendidos de 30 días (la API los manda; si no, los más vendidos). */
export const top3 = (t: Pick<TiendaPublica, 'masVendidos' | 'topVendidos'>) =>
	(t.topVendidos ?? t.masVendidos ?? []).slice(0, 3);

/** Desde cuántas unidades se avisa que quedan pocas (0: nunca). */
export const umbralStock = (c: Conversion) => c.stockBajo;

/**
 * Las etiquetas de un producto, en orden de importancia: la oferta (−X %), la propia de la tienda,
 * Nuevo (cargado hace 14 días o menos), Últimas N (con poco stock) y Más vendido (top 3 de 30
 * días). Próximamente o sin stock no llevan otra.
 */
export function etiquetasDe(
	p: ProductoPublico,
	t: Pick<TiendaPublica, 'masVendidos' | 'topVendidos'>,
	c: Conversion,
	ahora = Date.now()
): Etiqueta[] {
	if (proximamente(p)) return [{ texto: 'Próximamente', tipo: 'propia' }];
	const tj = tarjetaDe(p);
	if (tj.agotado) return [{ texto: 'Sin stock', tipo: 'propia' }];
	const r: Etiqueta[] = [];
	const off = tj.desde ? '' : descuento(tj.precio, p.anterior);
	if (off && c.etiquetas.oferta) r.push({ texto: off, tipo: 'off' });
	const propia = p.etiqueta?.trim();
	if (propia) r.push({ texto: propia, tipo: 'propia' });
	const alta = p.alta ? Date.parse(p.alta) : NaN;
	if (c.etiquetas.nuevo && Number.isFinite(alta) && ahora - alta <= 14 * DIA && ahora >= alta)
		r.push({ texto: 'Nuevo', tipo: 'nuevo' });
	const tope = c.stockBajo || 3;
	if (c.etiquetas.ultimas && p.stock > 0 && p.stock <= tope)
		r.push({ texto: p.stock === 1 ? 'Última' : `Últimas ${p.stock}`, tipo: 'ultimas' });
	if (c.etiquetas.masVendido && top3(t).includes(p.id))
		r.push({ texto: 'Más vendido', tipo: 'top' });
	// La propia de la tienda no se repite con una automática parecida («Más vendida» y «Más vendido»).
	const raiz = (s: string) => s.toLowerCase().replace(/[aeo]s?$/, '');
	return r.filter((e, i) => r.findIndex((x) => raiz(x.texto) === raiz(e.texto)) === i);
}

/**
 * «Sumá también» del carrito (hasta 3): primero lo que completa un combo de lo que ya está en el
 * carrito, después de la misma categoría y, si faltan, los más vendidos. Sólo lo que se puede
 * sumar con un toque o elegir (con stock, sin lo que ya está).
 */
export function sumaTambien(t: TiendaPublica, enCarrito: string[], tope = 3): ProductoPublico[] {
	const diseno = disenoDeTienda(t);
	const ya = new Set(enCarrito);
	const ok = (p: ProductoPublico | undefined): p is ProductoPublico =>
		!!p && !ya.has(p.id) && !tarjetaDe(p).agotado && !proximamente(p);
	const out: ProductoPublico[] = [];
	const sumar = (p: ProductoPublico | undefined) => {
		if (ok(p) && !out.includes(p) && out.length < tope) out.push(p);
	};
	for (const id of enCarrito) {
		const b = comboDelProducto(diseno, id);
		if (b) for (const p of comboDe(t, b)) sumar(p);
	}
	const cats = new Set(
		enCarrito.map((id) => t.productos.find((p) => p.id === id)?.categoria).filter(Boolean)
	);
	for (const p of t.productos) if (p.categoria && cats.has(p.categoria)) sumar(p);
	for (const id of t.masVendidos ?? []) sumar(t.productos.find((p) => p.id === id));
	return out;
}
