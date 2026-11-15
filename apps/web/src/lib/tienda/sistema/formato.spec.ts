import { describe, expect, test } from 'vitest';
import {
	conTransferencia,
	cuotas,
	descuento,
	miles,
	precio,
	promedio,
	sinViuda,
	urlFoto
} from './formato.ts';

describe('formato', () => {
	test('precios con punto de miles y espacio duro', () => {
		expect(precio(186000)).toBe('$\u00a0186.000');
		expect(precio(18000.4)).toBe('$\u00a018.000');
		expect(precio(950)).toBe('$\u00a0950');
		expect(miles(1234567)).toBe('1.234.567');
	});
	test('promedio con un decimal', () => {
		expect(promedio(4.8)).toBe('4,8');
		expect(promedio(4.455)).toBe('4,5');
		expect(promedio(5)).toBe('5,0');
	});
	test('cuotas y transferencia', () => {
		expect(cuotas(186000, 3)).toBe('3 cuotas de $\u00a062.000');
		expect(cuotas(186000, 6, true)).toBe('6 cuotas sin interés de $\u00a031.000');
		expect(cuotas(186000, 1)).toBe('');
		expect(conTransferencia(214500, 10)).toBe(193050);
	});
	test('descuento de una oferta', () => {
		expect(descuento(168000, 224000)).toBe('−25\u00a0%');
		expect(descuento(100, null)).toBe('');
		expect(descuento(100, 90)).toBe('');
	});
	test('sin viudas: la última palabra va pegada', () => {
		expect(sinViuda('Ropa de autor, hecha para durar.')).toBe(
			'Ropa de autor, hecha para\u00a0durar.'
		);
		expect(sinViuda('Lumbre')).toBe('Lumbre');
	});
	test('fotos de la demo y subidas', () => {
		expect(urlFoto('abc', 400)).toContain('images.unsplash.com/photo-abc');
		expect(urlFoto('https://x.com/a.webp', 400)).toBe('https://x.com/a.webp');
		expect(urlFoto(null, 400)).toBeNull();
	});
});
