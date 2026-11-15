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
	test('últimas: desde el stock bajo elegido', () => {
		const c = conversionDe({ conversion: { stockBajo: 5 } });
		expect(textos(p({ id: 'x', stock: 5 }), c)).toEqual(['Últimas 5']);
	});
});
