import { describe, expect, it } from 'vitest';
import type { ProductoPublico, TiendaPublica } from '../tipos.ts';
import type { Bloque } from './catalogo.ts';
import { categoriasDe, fotosGaleria, fotosPortada, fotosUsadas } from './datos.ts';

const prod = (id: string, p: Partial<ProductoPublico> = {}): ProductoPublico => ({
	id,
	nombre: id,
	descripcion: '',
	categoria: id < 'e' ? 'Abrigos' : 'Básicos',
	precio: 1000,
	anterior: null,
	foto: `f-${id}`,
	etiqueta: null,
	stock: 5,
	fotos: [],
	opciones: [],
	variantes: [],
	...p
});

const conProductos = (productos: ProductoPublico[]) =>
	({
		slug: 'x',
		nombre: 'X',
		categorias: [
			{ nombre: 'Abrigos', cantidad: 4, foto: 'f-a' },
			{ nombre: 'Básicos', cantidad: 4, foto: 'f-e' }
		],
		masVendidos: productos.map((p) => p.id),
		productos,
		fotos: productos.slice(0, 6).flatMap((p) => (p.foto ? [p.foto] : [])),
		portada: { titulo: '', lema: '', foto: null }
	}) as unknown as TiendaPublica;

const b = (id: string, tipo: string, variante: string, datos = {}): Bloque =>
	({ id, tipo, variante, datos }) as Bloque;

const pagina = [
	b('por', 'portada', 'partida'),
	b('cat', 'categorias', 'tarjetas'),
	b('pro', 'productos', 'grilla', { productos: { fuente: 'vendidos', cantidad: 8 } }),
	b('gal', 'galeria', 'fila')
];

describe('fotos de la página sin repetir', () => {
	it('la galería automática no se dibuja si sólo repetiría la grilla', () => {
		const t = conProductos('abcdefgh'.split('').map((id) => prod(id)));
		const usadas = fotosUsadas(t, pagina);
		expect(fotosGaleria(t, pagina[3], usadas.get('gal'))).toEqual([]);
		// Fuera de una página (sin usadas), sigue como antes.
		expect(fotosGaleria(t, pagina[3]).length).toBe(6);
	});

	it('la portada y las categorías prefieren las fotos que la grilla no muestra', () => {
		const t = conProductos(
			'abcdefgh'.split('').map((id) => prod(id, { fotos: [`f-${id}`, `g-${id}`] }))
		);
		const usadas = fotosUsadas(t, pagina);
		const portada = fotosPortada(t, pagina[0], usadas.get('por'));
		expect(portada.slice(0, 3).every((f) => f.startsWith('g-'))).toBe(true);
		const cats = categoriasDe(t, undefined, usadas.get('cat'));
		expect(cats.map((c) => c.foto)).toEqual(['g-d', 'g-e']);
		// La galería se arma con las segundas fotos que no tomaron ni la portada ni las categorías.
		const gal = fotosGaleria(t, pagina[3], usadas.get('gal'));
		expect(gal.length).toBe(3);
		expect(gal.some((f) => portada.slice(0, 3).includes(f) || f.startsWith('f-'))).toBe(false);
	});

	it('las fotos propias de un bloque se respetan aunque se repitan', () => {
		const t = conProductos('abcd'.split('').map((id) => prod(id)));
		const gal = b('gal', 'galeria', 'fila', { fotos: ['f-a', 'f-b', 'f-c'] });
		const usadas = fotosUsadas(t, [pagina[2], gal]);
		expect(fotosGaleria(t, gal, usadas.get('gal'))).toEqual(['f-a', 'f-b', 'f-c']);
	});

	it('sin otras fotos, la portada usa igual las de los productos', () => {
		const t = conProductos('abcd'.split('').map((id) => prod(id)));
		const usadas = fotosUsadas(t, pagina);
		expect(fotosPortada(t, pagina[0], usadas.get('por')).length).toBe(4);
	});
});
