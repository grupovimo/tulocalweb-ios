import { describe, expect, test } from 'vitest';
import casos from '@mostrador/contrato/casos/diseno-v2.json';
import { validarBloques } from './catalogo.ts';
import { migrar } from './migrar.ts';
import { PRESETS } from './presets.ts';

describe('migrar: los casos compartidos con Go (casos/diseno-v2.json)', () => {
	for (const c of casos.casos)
		test(c.nombre, () => {
			expect(migrar(c.v1, c.plantilla)).toEqual(c.v2);
		});
});

describe('migrar', () => {
	test('el inicial de cada preset es válido contra el catálogo', () => {
		for (const [id] of Object.entries(PRESETS)) {
			const d = migrar(null, id);
			expect(validarBloques(d.inicio), id).toEqual(d.inicio);
		}
	});
	test('es idempotente', () => {
		const v1 = casos.casos[1].v1;
		const una = migrar(v1, 'ceibo');
		expect(migrar(una, 'ceibo')).toBe(una);
	});
	test('no toca el diseño que recibe', () => {
		const v1 = structuredClone(casos.casos[1].v1);
		migrar(v1, 'ceibo');
		expect(v1).toEqual(casos.casos[1].v1);
	});
});
