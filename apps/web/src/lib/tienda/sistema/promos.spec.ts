import { describe, expect, test } from 'vitest';
import CASOS from '@mostrador/contrato/casos/ventanas.json';
import { ventanasParaGuardar } from '#lib/panel/mi-tienda/guardar.ts';
import {
	completa,
	conversionDe,
	conversionValida,
	validarVentana,
	ventanasValidas,
	type Ventana
} from './promos.ts';
import { califica, elegir, faltan, textoFaltan, varianteEn, type Ocasion } from './ventanas.ts';

describe('casos/ventanas.json (los mismos que Go)', () => {
	for (const c of CASOS.ventanas) {
		test(c.nombre, () => {
			const w = validarVentana(c.ventana);
			expect(w !== null).toBe(c.valida);
			if (w && c.valida) expect(completa(w)).toBe(c.completa);
		});
	}
	for (const c of CASOS.listas)
		test(`lista: ${c.nombre}`, () => expect(ventanasValidas(c.ventanas)).toBe(c.valida));
	for (const c of CASOS.conversion)
		test(`conversión: ${c.nombre}`, () => expect(conversionValida(c.conversion)).toBe(c.valida));
});

describe('guardar', () => {
	test('lo válido pasa igual y lo de otro premio se saca', () => {
		const validas = CASOS.ventanas.filter((c) => c.valida).map((c) => c.ventana) as Ventana[];
		const r = ventanasParaGuardar(validas.slice(0, 8));
		expect(r.length).toBe(Math.min(8, new Set(validas.slice(0, 8).map((w) => w.id)).size));
		for (const w of r) expect(validarVentana(w)).not.toBeNull();
	});
	test('la conversión vacía toma lo de por defecto', () => {
		const c = conversionDe({});
		expect(c.envioGratis).toBe(true);
		expect(c.stockBajo).toBe(3);
		expect(c.pruebaSocial).toEqual({ activa: false, cada: 30, paginas: [] });
		expect(conversionDe({ conversion: { stockBajo: 99 } }).stockBajo).toBe(3);
	});
});

describe('cuándo sale una ventana', () => {
	const base = CASOS.ventanas.find((c) => c.valida && c.completa)!.ventana as Ventana;
	const w = (x: Partial<Ventana> = {}): Ventana => ({ ...base, ...x });
	const o = (x: Partial<Ocasion> = {}): Ocasion => ({
		pagina: 'inicio',
		movil: false,
		ahora: Date.parse('2026-11-15T12:00:00Z'),
		memoria: { visitas: 1, vistas: {}, hechas: [] },
		sesion: { vistas: [] },
		...x
	});
	test('nunca en el checkout ni en el listo', () => {
		expect(califica(w({ paginas: [] }), o({ pagina: 'checkout' }))).toBe(false);
		expect(califica(w({ paginas: [] }), o({ pagina: 'listo' }))).toBe(false);
		expect(califica(w({ paginas: [], publico: 'todos' }), o())).toBe(true);
	});
	test('páginas, dispositivos, público y fechas', () => {
		const t = w({ publico: 'todos', paginas: ['producto'] });
		expect(califica(t, o())).toBe(false);
		expect(califica(t, o({ pagina: 'producto' }))).toBe(true);
		expect(califica(w({ publico: 'todos', dispositivos: 'celular' }), o())).toBe(false);
		expect(
			califica(w({ publico: 'nuevos' }), o({ memoria: { visitas: 2, vistas: {}, hechas: [] } }))
		).toBe(false);
		expect(califica(w({ publico: 'recurrentes' }), o())).toBe(false);
		expect(califica(w({ publico: 'todos', hasta: '2026-11-01T00:00:00Z' }), o())).toBe(false);
	});
	test('frecuencia y captación hecha', () => {
		const t = w({ publico: 'todos', frecuencia: { cada: 'dias', dias: 7 } });
		const vista = Date.parse('2026-11-12T12:00:00Z');
		expect(califica(t, o({ memoria: { visitas: 1, vistas: { [t.id]: vista }, hechas: [] } }))).toBe(
			false
		);
		expect(
			califica(
				t,
				o({
					ahora: vista + 8 * 86400e3,
					memoria: { visitas: 1, vistas: { [t.id]: vista }, hechas: [] }
				})
			)
		).toBe(true);
		const s = w({ publico: 'todos', frecuencia: { cada: 'sesion' } });
		expect(califica(s, o({ sesion: { vistas: [s.id] } }))).toBe(false);
		expect(califica(s, o({ memoria: { visitas: 1, vistas: {}, hechas: [s.id] } }))).toBe(
			base.contenido !== 'captacion' && base.contenido !== 'ruleta'
		);
	});
	test('una sola: la primera que califica', () => {
		const a = w({ id: 'aaaaaa', publico: 'todos', activa: false });
		const b = w({ id: 'bbbbbb', publico: 'todos' });
		const c = w({ id: 'cccccc', publico: 'todos' });
		expect(elegir([a, b, c], o())?.id).toBe('bbbbbb');
	});
	test('en el celular, completa y centrada al entrar suben desde abajo', () => {
		expect(varianteEn({ variante: 'completa', disparo: { cuando: 'salida' } }, true)).toBe(
			'inferior'
		);
		expect(varianteEn({ variante: 'centrada', disparo: { cuando: 'entrar' } }, true)).toBe(
			'inferior'
		);
		expect(varianteEn({ variante: 'centrada', disparo: { cuando: 'salida' } }, true)).toBe(
			'centrada'
		);
		expect(varianteEn({ variante: 'completa', disparo: { cuando: 'entrar' } }, false)).toBe(
			'completa'
		);
	});
	test('cuenta regresiva', () => {
		const ahora = Date.parse('2026-11-15T12:00:00Z');
		expect(faltan('2026-11-15T11:00:00Z', ahora)).toBeNull();
		expect(textoFaltan(faltan('2026-11-17T13:02:03Z', ahora)!)).toBe('2 d 01:02:03');
		expect(textoFaltan(faltan('2026-11-15T12:00:09Z', ahora)!)).toBe('00:00:09');
	});
});
