import { describe, expect, it } from 'vitest';
import {
	aLocal,
	aplicarLanzamiento,
	deLocal,
	errorLanzamiento,
	partes,
	proximoConteo,
	SIN_LANZAMIENTO,
	senaDe,
	textoConteo
} from './lanzamiento.ts';
import type { ProductoPublico } from './tipos.ts';

const prod = (x: Partial<ProductoPublico>): ProductoPublico => ({
	id: 'p',
	nombre: 'Buzo',
	descripcion: '',
	categoria: null,
	precio: 1000,
	anterior: null,
	foto: null,
	etiqueta: null,
	stock: 3,
	fotos: [],
	opciones: [],
	variantes: [],
	...x
});

describe('cuenta regresiva (§52)', () => {
	const ahora = Date.parse('2026-11-01T12:00:00Z');
	it('cuenta sólo lo real: el lanzamiento o la oferta con fin más cercanos', () => {
		expect(proximoConteo({ productos: [prod({})] }, ahora)).toBeNull();
		// Una oferta sin precio tachado no se cuenta (no hay oferta).
		expect(
			proximoConteo({ productos: [prod({ ofertaHasta: '2026-11-02T12:00:00Z' })] }, ahora)
		).toBeNull();
		const c = proximoConteo(
			{
				productos: [
					prod({ id: 'a', anterior: 1500, ofertaHasta: '2026-11-03T12:00:00Z' }),
					prod({ id: 'b', nombre: 'Zapa', lanzamiento: '2026-11-02T12:00:00Z' }),
					prod({ id: 'c', lanzamiento: '2026-10-30T12:00:00Z' })
				]
			},
			ahora
		)!;
		expect(c.producto.id).toBe('b');
		expect(textoConteo(c)).toBe('Zapa sale en');
	});
	it('parte el tiempo como el handoff, con 4 o 3 casillas', () => {
		const ms = ((2 * 24 + 3) * 3600 + 4 * 60 + 5) * 1000;
		expect(partes(ms).map((p) => p.v + p.u)).toEqual(['02d', '03h', '04m', '05s']);
		expect(partes(ms, 3).map((p) => p.v)).toEqual(['02', '03', '04']);
		expect(partes(3_725_000, 3).map((p) => p.v + p.u)).toEqual(['01h', '02m', '05s']);
		expect(
			partes(-5)
				.map((p) => p.v)
				.join('')
		).toBe('00000000');
	});
	it('hora de Argentina en el editor', () => {
		expect(deLocal('2026-11-07T19:00')).toBe('2026-11-07T22:00:00.000Z');
		expect(aLocal('2026-11-07T22:00:00.000Z')).toBe('2026-11-07T19:00');
		expect(deLocal('7/11')).toBeNull();
	});
	it('seña redondeada al peso y reglas para guardar', () => {
		expect(senaDe(55_120, 30)).toBe(16_536);
		expect(senaDe(1, 1)).toBe(1);
		const c = { ...SIN_LANZAMIENTO, ofertaPrecio: 900, ofertaHasta: '2026-11-01T11:00:00Z' };
		expect(errorLanzamiento(c, SIN_LANZAMIENTO, 1000, ahora)).toMatch(/terminar más adelante/);
		// La que ya estaba se puede dejar aunque haya pasado.
		expect(errorLanzamiento(c, c, 1000, ahora)).toBeNull();
	});
	it('aplica la oferta vigente y las señas sólo con Mercado Pago', () => {
		const c = {
			...SIN_LANZAMIENTO,
			lanzamiento: '2026-11-02T12:00:00Z',
			reserva: 30,
			ofertaPrecio: 800
		};
		expect(aplicarLanzamiento(1000, c, true, ahora)).toMatchObject({
			precio: 800,
			anterior: 1000,
			etiqueta: 'Próximamente',
			reserva: 30
		});
		expect(aplicarLanzamiento(1000, c, false, ahora).reserva).toBeUndefined();
		expect(
			aplicarLanzamiento(1000, { ...c, encargo: 50 }, true, Date.parse('2026-11-03T00:00:00Z'))
		).toMatchObject({ encargo: 50 });
	});
});
