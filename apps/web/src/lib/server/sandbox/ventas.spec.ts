import { describe, expect, it } from 'vitest';
import { RECORDATORIOS_POR_DEFECTO } from '#lib/panel/recordatorios.ts';
import { demoDe } from './demo.ts';
import { bandejaDemo, carritosDemo, conversacionDemo } from './ventas.ts';

const BASE = Date.parse('2026-10-20T15:00:00Z');
const d = demoDe('litoral', BASE)!;

describe('bandeja y carritos de la demo', () => {
	it('siete conversaciones: las tres del Resumen primero', () => {
		const b = bandejaDemo(d, new Map());
		expect(b.conversaciones).toHaveLength(7);
		expect(b.conversaciones[0].ultimo).toContain('envíos');
		expect(b.conversaciones.filter((c) => c.atiende === 'vos')).toHaveLength(2);
	});

	it('la sugerencia responde a lo último que preguntó y desaparece al responder', () => {
		const c = conversacionDemo(d, 'c1', new Map())!;
		expect(c.sugerencia).toMatch(/Andreani/);
		expect(c.mensajes.some((m) => m.producto)).toBe(true);
		const respondida = conversacionDemo(
			d,
			'c1',
			new Map([['c1', { mensajes: [{ texto: 'Sí, llega el jueves', en: BASE }] }]])
		)!;
		expect(respondida.sugerencia).toBeNull();
		expect(respondida.mensajes.at(-1)!.de).toBe('vos');
		expect(respondida.atiende).toBe('resuelto');
		expect(conversacionDemo(d, 'nada', new Map())).toBeNull();
	});

	it('los carritos suman y respetan los recordados', () => {
		const e = {
			recordados: new Set(['k1']),
			todos: false,
			automaticos: { hora: true, dia: true, email: false },
			recordatorios: RECORDATORIOS_POR_DEFECTO
		};
		const k = carritosDemo(d, BASE, e, 120);
		expect(k.abiertos).toBe(12);
		expect(k.monto).toBe(d.carritos.reduce((s, x) => s + x.total, 0));
		expect(k.carritos.find((x) => x.id === 'k1')!.recordado).toBe(true);
		expect(
			carritosDemo(d, BASE, { ...e, todos: true }, 120).carritos.every((x) => x.recordado)
		).toBe(true);
	});
});
