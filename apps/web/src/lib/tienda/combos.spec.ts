import { describe, expect, it } from 'vitest';
import casos from '@mostrador/contrato/casos/combos.json';
import { combosDe, descuentoCombos, type Combo, type LineaCombo } from './combos.ts';

describe('combos: los casos compartidos con Go (casos/combos.json)', () => {
	for (const c of casos.casos)
		it(c.nombre, () => {
			expect(descuentoCombos(c.combos as Combo[], c.lineas as LineaCombo[]).total).toBe(
				c.descuento
			);
		});
});

describe('combos del diseño', () => {
	it('sólo los visibles, con 2 o más productos y descuento', () => {
		const b = (id: string, datos: Record<string, unknown>, oculto = false) => ({
			id,
			tipo: 'combo' as const,
			variante: 'tarjeta',
			oculto,
			datos
		});
		expect(
			combosDe({
				inicio: [
					b('a', { productos: ['x', 'y'], descuento: 10 }),
					b('b', { productos: ['x', 'y'], descuento: 0 }),
					b('c', { productos: ['x'], descuento: 10 }),
					b('d', { productos: ['x', 'y'], descuento: 10 }, true)
				],
				paginas: {
					producto: {
						variante: 'grilla',
						relacionados: true,
						calcularEnvio: true,
						bloques: [b('e', { productos: ['y', 'z'], descuento: 5 })]
					}
				}
			})
		).toEqual([
			{ id: 'a', productos: ['x', 'y'], descuento: 10 },
			{ id: 'e', productos: ['y', 'z'], descuento: 5 }
		]);
	});
});
