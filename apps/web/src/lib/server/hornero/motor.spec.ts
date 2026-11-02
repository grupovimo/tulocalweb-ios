import { describe, expect, it } from 'vitest';
import { demoDe } from '../sandbox/demo.ts';
import { contextoDemo, tableroDemo } from '../sandbox/tablero.ts';
import type { DatosHornero } from './contexto.ts';
import { normalizar, productoDe, responderSinIa } from './motor.ts';
import { contextoSinMontos, sinMontos } from '../datos/montos.ts';

const BASE = Date.parse('2026-10-20T15:00:00Z');
const AHORA = BASE + 30 * 60_000;
const d = demoDe('litoral', BASE)!;
const datos = (): DatosHornero => ({
	tienda: { nombre: 'Mate & Co', rubro: 'almacen' },
	hola: 'Valeria',
	hoy: tableroDemo(d, 'hoy', { base: BASE, ahora: AHORA, hola: 'Valeria', cambios: {} }),
	semana: tableroDemo(d, '7d', { base: BASE, ahora: AHORA, hola: 'Valeria', cambios: {} }),
	contexto: contextoDemo(d, AHORA)
});
const pesos = (n: number) => '$ ' + Math.round(n).toLocaleString('es-AR');

describe('Tincho sin IA', () => {
	it('normaliza tildes y signos', () => {
		expect(normalizar('¿Cuánto VENDÍ ayer?')).toBe('cuanto vendi ayer');
	});

	it('ventas de ayer, de la semana y de hoy, con los datos reales', () => {
		const x = datos();
		const ayer = x.contexto.ventasPorDia.at(-2)!;
		expect(responderSinIa('¿Cuánto vendí ayer?', x)).toContain(pesos(ayer.ventas!));
		expect(responderSinIa('cuanto facturé esta semana', x)).toContain(
			pesos(x.semana.ventas!.actual)
		);
		expect(responderSinIa('¿cómo vienen las ventas hoy?', x)).toContain(
			pesos(x.hoy.ventas!.actual)
		);
	});

	it('lo que más sale y lo que está sin stock', () => {
		const x = datos();
		const top = [...x.contexto.productos].sort((a, b) => b.vendidos30d - a.vendidos30d)[0];
		expect(responderSinIa('¿qué es lo que más se vende?', x)).toContain(top.nombre);
		expect(responderSinIa('qué tengo que reponer?', x)).toMatch(
			/sin stock|quedan|Está todo con stock/i
		);
	});

	it('stock y precio de un producto que nombra a medias', () => {
		const x = datos();
		const p = x.contexto.productos.find((y) => y.nombre.split(' ').length > 1)!;
		const palabra = p.nombre.split(' ').find((w) => w.length > 3)!;
		expect(
			productoDe(`cuanto stock queda de ${palabra}`, x.contexto.productos)?.nombre
		).toBeTruthy();
		expect(responderSinIa(`¿cuánto sale el ${p.nombre.toLowerCase()}?`, x)).toContain(
			pesos(p.precio)
		);
	});

	it('los "cómo hago" llevan a la sección del panel', () => {
		expect(responderSinIa('¿Cómo cargo un producto nuevo?', datos())).toContain('Nuevo producto');
		expect(responderSinIa('quiero crear un cupón de descuento', datos())).toContain('Cupones');
	});

	it('lo que no entiende queda para la IA', () => {
		expect(responderSinIa('escribime un poema sobre el otoño', datos())).toBeNull();
	});

	it('a Ventas y Local (sin montos) les cuenta pedidos, nunca pesos', () => {
		const x = datos();
		const sin: DatosHornero = {
			...x,
			hoy: sinMontos(x.hoy),
			semana: sinMontos(x.semana),
			contexto: contextoSinMontos(x.contexto)
		};
		for (const q of [
			'hola',
			'¿Cuánto vendí ayer?',
			'cuanto facturé esta semana',
			'¿cómo vienen las ventas hoy?',
			'ventas de los últimos 14 días',
			'¿cuál fue mi mejor día de ventas?',
			'¿cuál es el ticket promedio?',
			'¿por qué canal vendo más?'
		]) {
			const r = responderSinIa(q, sin);
			expect(r, q).toBeTruthy();
			expect(r, q).not.toContain('$');
		}
		expect(responderSinIa('cuanto facturé esta semana', sin)).toContain(
			String(x.semana.kpis.pedidos.actual)
		);
	});
});
