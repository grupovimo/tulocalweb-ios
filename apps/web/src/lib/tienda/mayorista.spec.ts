import { describe, expect, it } from 'vitest';
import { avisoMayorista, detalleLineas } from './lineas.ts';
import {
	alcanza,
	falta,
	lineaPorMayor,
	mayoristaDeVariante,
	precioMayorista,
	problemaMayorista,
	textoAviso,
	textoFalta,
	textoMinimo
} from './mayorista.ts';
import type { ProductoPublico } from './tipos.ts';

const pesos = (n: number) => '$ ' + n.toLocaleString('es-AR');
const base = { descripcion: '', categoria: null, anterior: null, etiqueta: null, fotos: [] };
const mate: ProductoPublico = {
	...base,
	id: 'mate',
	nombre: 'Mate',
	precio: 10_000,
	mayorista: 8_000,
	foto: null,
	stock: 50,
	opciones: [],
	variantes: []
};
const yerba: ProductoPublico = {
	...mate,
	id: 'yerba',
	nombre: 'Yerba',
	precio: 5_000,
	mayorista: null
};
const buzo: ProductoPublico = {
	...mate,
	id: 'buzo',
	nombre: 'Buzo',
	precio: 24_000,
	mayorista: 20_000,
	opciones: [{ nombre: 'Talle', valores: ['M', 'L'] }],
	variantes: [
		{ id: 'm', valores: { Talle: 'M' }, precio: null, stock: 9, foto: null },
		{ id: 'l', valores: { Talle: 'L' }, precio: 26_000, mayorista: 21_000, stock: 9, foto: null }
	]
};
const productos = [mate, yerba, buzo];
const linea = (id: string, cantidad: number, varianteId: string | null = null) => ({
	id,
	varianteId,
	cantidad
});

describe('venta mayorista (los mismos casos que Go)', () => {
	it('alcanza: unidades, monto y los dos', () => {
		expect(alcanza(null, 10, 0)).toBe(false);
		expect(alcanza({ unidades: null, monto: null }, 10, 100_000)).toBe(false);
		expect(alcanza({ unidades: 6, monto: null }, 6, 0)).toBe(true);
		expect(alcanza({ unidades: 6, monto: null }, 5, 999_999)).toBe(false);
		expect(alcanza({ unidades: null, monto: 30_000 }, 1, 30_000)).toBe(true);
		expect(alcanza({ unidades: null, monto: 30_000 }, 50, 29_999)).toBe(false);
		expect(alcanza({ unidades: 6, monto: 30_000 }, 6, 30_000)).toBe(true);
		expect(alcanza({ unidades: 6, monto: 30_000 }, 6, 29_000)).toBe(false);
		expect(alcanza({ unidades: 6, monto: 30_000 }, 5, 90_000)).toBe(false);
	});

	it('el precio mayorista nunca sube; la variante usa el suyo o el del producto', () => {
		expect(precioMayorista(10_000, null)).toBe(10_000);
		expect(precioMayorista(10_000, 8_000)).toBe(8_000);
		expect(precioMayorista(10_000, 12_000)).toBe(10_000);
		expect(precioMayorista(10_000, 0)).toBe(10_000);
		expect(mayoristaDeVariante(null, 5)).toBe(5);
		expect(mayoristaDeVariante(3, 5)).toBe(3);
		expect(mayoristaDeVariante(undefined, undefined)).toBe(null);
	});

	it('textos sin renglones viudos', () => {
		expect(textoMinimo({ unidades: 12, monto: null }, pesos)).toBe('12\u00a0unidades');
		expect(textoMinimo({ unidades: null, monto: 50_000 }, pesos)).toBe(
			'$\u00a050.000 de\u00a0compra'
		);
		expect(textoMinimo({ unidades: 6, monto: 150_000 }, pesos)).toBe(
			'6\u00a0unidades y $\u00a0150.000'
		);
		expect(textoFalta({ unidades: 1, monto: 0 }, pesos)).toBe('1\u00a0unidad\u00a0más');
		expect(textoFalta({ unidades: 3, monto: 4_000 }, pesos)).toBe(
			'3\u00a0unidades y $\u00a04.000\u00a0más'
		);
		expect(textoFalta({ unidades: 0, monto: 0 }, pesos)).toBe(null);
		expect(falta({ unidades: 6, monto: null }, 8, 0)).toEqual({ unidades: 0, monto: 0 });
		expect(lineaPorMayor(8_000, { unidades: 12, monto: null }, pesos)).toBe(
			'Por mayor: $\u00a08.000 desde 12\u00a0unidades'
		);
		expect(lineaPorMayor(null, { unidades: 12, monto: null }, pesos)).toBe(null);
		expect(lineaPorMayor(8_000, null, pesos)).toBe(null);
		expect(
			textoAviso({ aplicado: false, falta: { unidades: 3, monto: 0 }, ahorro: 9_600 }, pesos)
		).toEqual({
			pre: 'Sumando ',
			b: '3\u00a0unidades\u00a0más',
			post: ' pagás precio mayorista: ahorrás\u00a0$\u00a09.600'
		});
		expect(
			textoAviso({ aplicado: true, falta: { unidades: 0, monto: 0 }, ahorro: 2_000 }, pesos).b
		).toBe('Precio mayorista');
	});

	it('valida el mínimo', () => {
		expect(problemaMayorista({ unidades: null, monto: null })).toMatch(/mínimo/);
		expect(problemaMayorista({ unidades: 1, monto: null })).toMatch(/Revisá/);
		expect(problemaMayorista({ unidades: null, monto: 0 })).toMatch(/Revisá/);
		expect(problemaMayorista({ unidades: 2, monto: 1 })).toBe(null);
	});
});

describe('el carrito al por mayor', () => {
	const minimo = { unidades: 6, monto: null };

	it('abajo del mínimo: minorista, con el aviso de cuánto falta y lo que ahorraría', () => {
		const d = detalleLineas([linea('mate', 3), linea('yerba', 2)], productos, minimo);
		expect(d.map((l) => l.subtotal)).toEqual([30_000, 10_000]);
		expect(d.every((l) => !l.alPorMayor)).toBe(true);
		expect(d[0].mayorista).toBe(8_000);
		expect(d[1].mayorista).toBe(null);
		expect(avisoMayorista(d, minimo)).toEqual({
			aplicado: false,
			falta: { unidades: 1, monto: 0 },
			ahorro: 6_000
		});
	});

	it('al mínimo: los que tienen mayorista pasan a ese y el minorista queda tachado', () => {
		const d = detalleLineas([linea('mate', 4), linea('yerba', 2)], productos, minimo);
		expect(d.map((l) => l.subtotal)).toEqual([32_000, 10_000]);
		expect(d[0]).toMatchObject({ alPorMayor: true, producto: { precio: 8_000, anterior: 10_000 } });
		expect(d[1].alPorMayor).toBe(false);
		expect(avisoMayorista(d, minimo)).toMatchObject({ aplicado: true, ahorro: 8_000 });
	});

	it('variantes: la suya (L) o la del producto (M)', () => {
		const d = detalleLineas([linea('buzo', 3, 'l'), linea('buzo', 3, 'm')], productos, minimo);
		expect(d.map((l) => l.subtotal)).toEqual([3 * 21_000, 3 * 20_000]);
	});

	it('por monto, contra los precios minoristas', () => {
		const m = { unidades: null, monto: 30_000 };
		expect(detalleLineas([linea('mate', 3)], productos, m)[0].subtotal).toBe(24_000);
		expect(detalleLineas([linea('mate', 2)], productos, m)[0].subtotal).toBe(20_000);
	});

	it('sin venta mayorista (o sin ningún precio mayorista) no hay aviso', () => {
		const d = detalleLineas([linea('mate', 10)], productos, null);
		expect(d[0].subtotal).toBe(100_000);
		expect(avisoMayorista(d, null)).toBe(null);
		expect(avisoMayorista(detalleLineas([linea('yerba', 2)], productos, minimo), minimo)).toBe(
			null
		);
	});
});
