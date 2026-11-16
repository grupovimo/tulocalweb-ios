import { describe, expect, it } from 'vitest';
import { problemaCobros, responsableInscripto, sinImpuestos } from './cobros.ts';

// Los mismos casos que apps/api/internal/dominio/cobros/cobros_test.go.
describe('cobros de la tienda', () => {
	it('la condición fiscal es texto libre', () => {
		for (const c of ['Responsable inscripta', 'responsable_inscripto', 'RI'])
			expect(responsableInscripto(c)).toBe(true);
		for (const c of ['Monotributo', 'monotributo', 'exento', '', undefined])
			expect(responsableInscripto(c)).toBe(false);
	});
	it('valida cuotas y monto', () => {
		expect(problemaCobros({ cuotas: 0, cuotasDesde: 0 })).toBeNull();
		expect(problemaCobros({ cuotas: 12, cuotasDesde: 30_000 })).toBeNull();
		expect(problemaCobros({ cuotas: 18, cuotasDesde: 0 })).toMatch(/cuotas/);
		expect(problemaCobros({ cuotas: 6, cuotasDesde: -1 })).toMatch(/monto/);
	});
	it('precio sin impuestos nacionales, redondeado a centavos', () => {
		expect(sinImpuestos(12100)).toBe(10000);
		expect(sinImpuestos(1000, 10.5)).toBe(904.98);
		expect(sinImpuestos(1000, 0)).toBe(1000);
	});
});
