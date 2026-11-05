import { readFileSync } from 'node:fs';
import { describe, expect, it } from 'vitest';
import { cotizar, ErrorCotizar, type EstadoAbono } from './cotizar.ts';
import type { IdPlan } from './planes.ts';

// La misma tabla que el test de Go: el cobro y la pantalla hacen la misma cuenta.
type Caso = {
	nombre: string;
	estado: EstadoAbono;
	ahora: string;
	plan: IdPlan;
	ciclo: 'mes' | 'anio';
	espera?: Record<string, unknown>;
	error?: string;
};
const casos: Caso[] = JSON.parse(
	readFileSync(
		new URL('../../../../api/internal/dominio/abono/testdata/cotizar.json', import.meta.url),
		'utf8'
	)
);

// Las fechas se comparan por instante (Go escribe "…T12:00:00Z" y JS "…T12:00:00.000Z").
const normal = (v: unknown): unknown => {
	if (typeof v === 'string' && /^\d{4}-\d\d-\d\dT/.test(v)) return new Date(v).getTime();
	if (Array.isArray(v)) return v.map(normal);
	if (v && typeof v === 'object')
		return Object.fromEntries(Object.entries(v).map(([k, x]) => [k, normal(x)]));
	return v;
};

describe('cotizar (espejo de abono.Cotizar)', () => {
	for (const c of casos) {
		it(c.nombre, () => {
			if (c.error) {
				expect(() => cotizar(c.estado, c.plan, c.ciclo, new Date(c.ahora))).toThrow(ErrorCotizar);
				return;
			}
			const r = cotizar(c.estado, c.plan, c.ciclo, new Date(c.ahora));
			expect(normal(r)).toEqual(normal({ ...c.espera, plan: c.plan, ciclo: c.ciclo }));
			if (r.lineas.length) expect(r.lineas.reduce((s, l) => s + l.importe, 0)).toBe(r.hoy);
		});
	}
});
