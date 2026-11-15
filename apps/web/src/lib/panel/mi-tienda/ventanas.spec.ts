import { describe, expect, test } from 'vitest';
import { origenDe } from '#lib/panel/newsletter.ts';
import { validarVentana, VENTANAS } from '#lib/tienda/sistema/promos.ts';
import type { DisenoV2 } from './diseno.ts';
import { conContenido, desdePlantilla, duplicar, mover } from './ventanas.ts';

const d: DisenoV2 = { version: 2, paleta: 0, inicio: [] };
const plantilla = (id: string) => VENTANAS.plantillas.find((p) => p.id === id)!;

describe('ventanas en el editor', () => {
	test('las plantillas salen válidas, con el monto del envío gratis y un id nuevo', () => {
		for (const p of VENTANAS.plantillas)
			expect(validarVentana(desdePlantilla(p, d, 60000)), p.id).not.toBeNull();
		const w = desdePlantilla(plantilla('envio-gratis'), d, 60000);
		expect(w.datos.titulo).toBe('Envío gratis desde $\u00a060.000');
		expect(desdePlantilla(plantilla('envio-gratis'), d, null).datos.titulo).toContain('$\u00a0X');
	});
	test('lanzamiento: la cuenta regresiva a los días de la plantilla', () => {
		const ahora = Date.parse('2026-11-15T12:00:00Z');
		const w = desdePlantilla(plantilla('lanzamiento'), d, null, ahora);
		expect(Date.parse(w.datos.hasta as string) - ahora).toBe(7 * 86_400_000);
	});
	test('en blanco: pausada y válida', () => {
		const w = desdePlantilla(null, d, null);
		expect(w.activa).toBe(false);
		expect(validarVentana(w)).not.toBeNull();
	});
	test('duplicar, mover y cambiar el contenido conservando lo común', () => {
		const a = desdePlantilla(plantilla('bienvenida'), d, null);
		const l = duplicar([a], 0, { ...d, ventanas: [a] });
		expect(l).toHaveLength(2);
		expect(l[1].id).not.toBe(a.id);
		expect(l[1].activa).toBe(false);
		expect(mover(l, 1, 0)[0].id).toBe(l[1].id);
		const promo = conContenido(a, 'promo');
		expect(promo.datos.titulo).toBe(a.datos.titulo);
		expect(promo.datos.premio).toBeUndefined();
	});
	test('newsletter: el origen de una ventana con su nombre', () => {
		expect(origenDe('ventana:bienve01', { bienve01: 'Bienvenida' })).toBe('Ventana «Bienvenida»');
		expect(origenDe('ventana:otra')).toBe('Ventana de la tienda');
		expect(origenDe('checkout')).toBe('Checkout');
	});
});
