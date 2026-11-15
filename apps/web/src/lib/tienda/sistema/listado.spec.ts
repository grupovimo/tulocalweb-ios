import { describe, expect, it } from 'vitest';
import type { ProductoPublico, TiendaPublica } from '../tipos.ts';
import {
	filtrar,
	hayFiltros,
	hrefFiltros,
	leerFiltros,
	porCategoria,
	SIN_FILTROS
} from './listado.ts';

const prod = (id: string, p: Partial<ProductoPublico> = {}): ProductoPublico => ({
	id,
	nombre: id,
	descripcion: '',
	categoria: 'Yerbas',
	precio: 1000,
	anterior: null,
	foto: null,
	etiqueta: null,
	stock: 5,
	fotos: [],
	opciones: [],
	variantes: [],
	...p
});

const tienda = {
	categorias: [
		{ nombre: 'Yerbas', cantidad: 3, foto: null },
		{ nombre: 'Mates', cantidad: 2, foto: null }
	],
	masVendidos: ['b', 'a'],
	productos: [
		prod('a', { nombre: 'Yerba Barbacuá', precio: 3000 }),
		prod('b', { nombre: 'Café de especialidad', precio: 9000, anterior: 12000 }),
		prod('c', { nombre: 'Mate de calabaza', categoria: 'Mates', precio: 5000, stock: 0 }),
		prod('d', { nombre: 'Bombilla', categoria: 'Mates', precio: 2000, mayorista: 1500 })
	]
} as unknown as TiendaPublica;

describe('listado del catálogo', () => {
	it('lee los filtros de la URL e ignora lo que no vale', () => {
		const f = leerFiltros(
			new URLSearchParams('cat=Mates&cat=Nada&q=%20mate%20&orden=mayor&ofertas=1&min=1.000'),
			tienda
		);
		expect(f).toEqual({
			...SIN_FILTROS,
			cats: ['Mates'],
			q: 'mate',
			orden: 'mayor',
			ofertas: true,
			min: 1000
		});
		expect(leerFiltros(new URLSearchParams('orden=raro'), tienda).orden).toBe('vendidos');
	});

	it('arma el link sin lo que está por defecto', () => {
		expect(hrefFiltros('/t/x', SIN_FILTROS)).toBe('/t/x/catalogo');
		expect(hrefFiltros('', { ...SIN_FILTROS, cats: ['Mates', 'Yerbas'], stock: true })).toBe(
			'/catalogo?cat=Mates&cat=Yerbas&stock=1'
		);
	});

	it('busca sin tildes y ordena con los agotados al final', () => {
		expect(filtrar(tienda, { ...SIN_FILTROS, q: 'cafe' }).map((p) => p.id)).toEqual(['b']);
		expect(filtrar(tienda, SIN_FILTROS).map((p) => p.id)).toEqual(['b', 'a', 'd', 'c']);
		expect(filtrar(tienda, { ...SIN_FILTROS, orden: 'menor' }).map((p) => p.id)).toEqual([
			'd',
			'a',
			'b',
			'c'
		]);
	});

	it('filtra por ofertas, stock, por mayor y precio', () => {
		const ids = (f: Partial<typeof SIN_FILTROS>) =>
			filtrar(tienda, { ...SIN_FILTROS, ...f }).map((p) => p.id);
		expect(ids({ ofertas: true })).toEqual(['b']);
		expect(ids({ stock: true })).not.toContain('c');
		expect(ids({ mayor: true })).toEqual(['d']);
		expect(ids({ min: 2500, max: 6000 }).sort()).toEqual(['a', 'c']);
		expect(hayFiltros(SIN_FILTROS)).toBe(false);
	});

	it('cuenta por categoría con los demás filtros', () => {
		expect(Object.fromEntries(porCategoria(tienda, { ...SIN_FILTROS, stock: true }))).toEqual({
			Yerbas: 2,
			Mates: 1
		});
	});
});
