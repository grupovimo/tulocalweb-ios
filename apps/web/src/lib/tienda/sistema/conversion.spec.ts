import { describe, expect, test } from 'vitest';
import type { ProductoPublico } from '../tipos.ts';
import { etiquetasDe } from './conversion.ts';
import { conversionDe } from './promos.ts';

const p = (x: Partial<ProductoPublico> = {}): ProductoPublico => ({
	id: 'p1',
	nombre: 'Mate Imperial',
	descripcion: '',
	categoria: 'Mates',
	precio: 8000,
	anterior: null,
	foto: null,
	etiqueta: null,
	stock: 10,
	fotos: [],
	opciones: [],
	variantes: [],
	...x
});
const ahora = Date.parse('2026-11-15T12:00:00Z');
const t = { masVendidos: ['p9', 'p1', 'p2', 'p3'], topVendidos: ['p1'] };
const textos = (x: ProductoPublico, c = conversionDe({})) =>
	etiquetasDe(x, t, c, ahora).map((e) => e.texto);

describe('etiquetas automáticas', () => {
	test('oferta, nuevo, últimas y más vendido, en ese orden', () => {
		expect(textos(p({ anterior: 10000, alta: '2026-11-10T00:00:00Z', stock: 2 }))).toEqual([
			'−20\u00a0%',
			'Nuevo',
			'Últimas 2',
			'Más vendido'
		]);
	});
	test('nuevo: hasta 14 días', () => {
		expect(textos(p({ id: 'x', alta: '2026-10-20T00:00:00Z' }))).toEqual([]);
	});
	test('apagadas en la conversión, no salen', () => {
		const c = conversionDe({
			conversion: { etiquetas: { nuevo: false, ultimas: false, masVendido: false, oferta: false } }
		});
		expect(textos(p({ anterior: 10000, alta: '2026-11-10T00:00:00Z', stock: 1 }), c)).toEqual([]);
	});
	test('sin stock o próximamente no llevan otra', () => {
		expect(textos(p({ stock: 0, anterior: 10000 }))).toEqual(['Sin stock']);
	});
	test('la de la tienda no se repite con una automática', () => {
		expect(textos(p({ id: 'x', etiqueta: 'Nuevo', alta: '2026-11-14T00:00:00Z' }))).toEqual([
			'Nuevo'
		]);
	});
	test('«Últimos» de la tienda y «Últimas 2» automática no van juntas', () => {
		expect(textos(p({ id: 'x', etiqueta: 'Últimos', stock: 2 }))).toEqual(['Últimos']);
	});
	test('últimas: desde el stock bajo elegido', () => {
		const c = conversionDe({ conversion: { stockBajo: 5 } });
		expect(textos(p({ id: 'x', stock: 5 }), c)).toEqual(['Últimas 5']);
	});
	test('Nuevo no sale si casi todos serían nuevos: no distingue', () => {
		const alta = '2026-11-10T00:00:00Z';
		const productos = [
			p({ id: 'a', alta }),
			p({ id: 'b', alta }),
			p({ id: 'c', alta }),
			p({ id: 'd' })
		];
		const de = (x: ProductoPublico) =>
			etiquetasDe(x, { ...t, productos }, conversionDe({}), ahora).map((e) => e.texto);
		expect(de(productos[0])).toEqual([]);
		const pocos = [p({ id: 'a', alta }), p({ id: 'b' }), p({ id: 'c' }), p({ id: 'd' })];
		expect(
			etiquetasDe(pocos[0], { ...t, productos: pocos }, conversionDe({}), ahora).map((e) => e.texto)
		).toEqual(['Nuevo']);
	});
	test('Más vendido no sale con 1 a 3 productos', () => {
		const tres = [p({ id: 'p1' }), p({ id: 'p2' }), p({ id: 'p3' })];
		const de = (productos: ProductoPublico[]) =>
			etiquetasDe(productos[0], { ...t, productos }, conversionDe({}), ahora).map((e) => e.texto);
		expect(de(tres)).toEqual([]);
		expect(de([...tres, p({ id: 'p4' })])).toEqual(['Más vendido']);
	});
});
