import { describe, expect, test } from 'vitest';
import casos from '@mostrador/contrato/casos/diseno-v2.json';
import { validarBloques } from './catalogo.ts';
import { conVisibles, estiloV1, migrar, seccionesV1 } from './migrar.ts';
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
	test('convivencia: secciones y estilo v1 derivados del v2, como Go', () => {
		const d = migrar(casos.casos[2].v1, 'andes');
		expect(seccionesV1(d)).toEqual(['prods', 'hero', 'cats']);
		const conAnuncio = migrar(casos.casos[1].v1, 'ceibo');
		const e = estiloV1(conAnuncio);
		expect(e.anuncio).toBe('Envío gratis desde $ 80.000');
		expect(e.textos).toEqual({ prods: 'Lo más pedido', ig: 'Usado de verdad' });
	});
	test('el onboarding prende y apaga los bloques de las secciones', () => {
		const d = conVisibles(migrar(null, 'ceibo'), { revs: false, cats: true });
		expect(d.inicio.find((b) => b.tipo === 'resenas')?.oculto).toBe(true);
		expect(d.inicio.find((b) => b.tipo === 'categorias')?.oculto).toBeUndefined();
	});
	test('no toca el diseño que recibe', () => {
		const v1 = structuredClone(casos.casos[1].v1);
		migrar(v1, 'ceibo');
		expect(v1).toEqual(casos.casos[1].v1);
	});
});
