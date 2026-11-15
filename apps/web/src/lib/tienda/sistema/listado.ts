/**
 * El catálogo de la tienda: qué productos se ven con lo que pide la URL (categorías, búsqueda,
 * orden, ofertas, con stock, por mayor y precio) y cómo se arma cada link de filtro. Los filtros van
 * en la URL (?cat=…&q=…&orden=…&ofertas=1&stock=1&mayor=1&min=…&max=…), así un link de «Ver
 * ofertas» o de una categoría llega ya filtrado y el navegador vuelve atrás bien.
 */
import { proximamente } from '../lanzamiento.ts';
import { tarjetaDe } from '../seleccion.svelte.ts';
import type { ProductoPublico, TiendaPublica } from '../tipos.ts';

export const ORDENES = [
	['vendidos', 'Más vendidos'],
	['nuevos', 'Novedades'],
	['menor', 'Menor precio'],
	['mayor', 'Mayor precio'],
	['az', 'De la A a la Z']
] as const;
export type Orden = (typeof ORDENES)[number][0];

export type Filtros = {
	cats: string[];
	q: string;
	orden: Orden;
	ofertas: boolean;
	stock: boolean;
	mayor: boolean;
	min: number | null;
	max: number | null;
};

export const SIN_FILTROS: Filtros = {
	cats: [],
	q: '',
	orden: 'vendidos',
	ofertas: false,
	stock: false,
	mayor: false,
	min: null,
	max: null
};

const numero = (s: string | null) => {
	const n = Number((s ?? '').replace(/\D/g, ''));
	return s && Number.isFinite(n) && n > 0 ? n : null;
};

/** Los filtros de la URL (lo que no vale, se ignora). Las categorías que no existen no filtran. */
export function leerFiltros(
	q: Pick<URLSearchParams, 'get' | 'getAll' | 'has'>,
	t: Pick<TiendaPublica, 'categorias'>
): Filtros {
	const nombres = new Set(t.categorias.map((c) => c.nombre));
	const orden = q.get('orden') ?? '';
	return {
		cats: [...new Set(q.getAll('cat'))].filter((c) => nombres.has(c)),
		q: (q.get('q') ?? '').trim().slice(0, 60),
		orden: ORDENES.some(([k]) => k === orden) ? (orden as Orden) : 'vendidos',
		ofertas: q.has('ofertas'),
		stock: q.has('stock'),
		mayor: q.has('mayor'),
		min: numero(q.get('min')),
		max: numero(q.get('max'))
	};
}

/** La URL del catálogo con esos filtros (lo que está por defecto no va). */
export function hrefFiltros(base: string, f: Filtros): string {
	const q = new URLSearchParams();
	for (const c of f.cats) q.append('cat', c);
	if (f.q) q.set('q', f.q);
	if (f.orden !== 'vendidos') q.set('orden', f.orden);
	if (f.ofertas) q.set('ofertas', '1');
	if (f.stock) q.set('stock', '1');
	if (f.mayor) q.set('mayor', '1');
	if (f.min) q.set('min', String(f.min));
	if (f.max) q.set('max', String(f.max));
	const s = q.toString();
	return `${base}/catalogo${s ? `?${s}` : ''}`;
}

/** Sin tildes ni mayúsculas, para buscar «cafe» y encontrar «Café». */
export const limpio = (s: string) => s.normalize('NFD').replace(/[̀-ͯ]/g, '').toLowerCase();

export const enOferta = (p: ProductoPublico) => !!p.anterior && p.anterior > p.precio;

/** ¿Hay algún filtro puesto (sin contar el orden)? */
export const hayFiltros = (f: Filtros) =>
	!!(f.cats.length || f.q || f.ofertas || f.stock || f.mayor || f.min || f.max);

/** Los productos que se ven con esos filtros, en el orden pedido. */
export function filtrar(t: TiendaPublica, f: Filtros): ProductoPublico[] {
	const palabras = limpio(f.q).split(/\s+/).filter(Boolean);
	const lista = t.productos.filter((p) => {
		if (!p.nombre) return false;
		if (f.cats.length && !f.cats.includes(p.categoria ?? '')) return false;
		if (f.ofertas && !enOferta(p)) return false;
		if (f.stock && (tarjetaDe(p).agotado || proximamente(p))) return false;
		if (f.mayor && !(p.mayorista && p.mayorista < p.precio)) return false;
		const precio = tarjetaDe(p).precio;
		if (f.min && precio < f.min) return false;
		if (f.max && precio > f.max) return false;
		if (palabras.length) {
			const texto = limpio(`${p.nombre} ${p.categoria ?? ''} ${p.etiqueta ?? ''} ${p.descripcion}`);
			if (!palabras.every((w) => texto.includes(w))) return false;
		}
		return true;
	});
	const pos = new Map(t.masVendidos.map((id, i) => [id, i]));
	const orden = new Map(t.productos.map((p, i) => [p.id, i]));
	const precio = (p: ProductoPublico) => tarjetaDe(p).precio;
	const por: Record<Orden, (a: ProductoPublico, b: ProductoPublico) => number> = {
		vendidos: (a, b) => (pos.get(a.id) ?? 1e9) - (pos.get(b.id) ?? 1e9),
		// La tienda los manda del más nuevo al más viejo; los lanzamientos, primero.
		nuevos: (a, b) =>
			Number(proximamente(b)) - Number(proximamente(a)) ||
			(orden.get(a.id) ?? 0) - (orden.get(b.id) ?? 0),
		menor: (a, b) => precio(a) - precio(b),
		mayor: (a, b) => precio(b) - precio(a),
		az: (a, b) => a.nombre.localeCompare(b.nombre, 'es')
	};
	// Los agotados, al final (salvo en la A a la Z).
	const agotado = (p: ProductoPublico) => Number(tarjetaDe(p).agotado);
	return [...lista].sort(
		(a, b) => (f.orden === 'az' ? 0 : agotado(a) - agotado(b)) || por[f.orden](a, b)
	);
}

/** Cuántos productos de cada categoría quedan con los demás filtros (para los contadores). */
export function porCategoria(t: TiendaPublica, f: Filtros): Map<string, number> {
	const sin = filtrar(t, { ...f, cats: [] });
	const n = new Map<string, number>();
	for (const p of sin) if (p.categoria) n.set(p.categoria, (n.get(p.categoria) ?? 0) + 1);
	return n;
}
