import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import { afterEach, describe, expect, it } from 'vitest';
import { d1Local } from './d1-local.ts';
import { crear, DatosSandbox, dbSandbox } from './sandbox.ts';
import type { RegCliente } from './listas.ts';

const dir = join(import.meta.dirname, '../../../../migraciones-sandbox');
const migraciones = readdirSync(dir)
	.filter((f) => f.endsWith('.sql'))
	.sort()
	.map((f) => readFileSync(join(dir, f), 'utf8'));

let d1: ReturnType<typeof d1Local> | undefined;
afterEach(() => {
	d1?.cerrar();
	d1 = undefined;
});
async function demo() {
	d1 = d1Local(migraciones);
	const db = dbSandbox(d1);
	const sb = (await crear(db, 'moda', Date.now(), 'ceibo'))!;
	return new DatosSandbox(db, sb);
}
const filas = async (d: DatosSandbox) =>
	(await d.listar('clientes', { desde: 0, cuantos: 200 })).filas as unknown as RegCliente[];

describe('corregir y borrar los datos de un cliente (como la API)', () => {
	it('corrige el nombre, el email y el WhatsApp', async () => {
		const d = await demo();
		const [c] = await filas(d);
		expect(
			await d.accionLista('clientes', [c.id], 'editar', {
				nombre: 'Ana María',
				email: 'Ana@Ejemplo.com',
				telefono: '223 555-0000'
			})
		).toBe(1);
		const otro = (await filas(d)).find((x) => x.id === c.id)!;
		expect(otro).toMatchObject({ nombre: 'Ana María', email: 'ana@ejemplo.com', wa: true });
		await expect(d.accionLista('clientes', [c.id], 'editar', { nombre: 'x' })).rejects.toThrow();
	});

	it('suprime a quien no tiene nada en curso y la saca de la lista', async () => {
		const d = await demo();
		const antes = await filas(d);
		let borrado: string | null = null;
		for (const c of antes.slice(0, 10)) {
			const n = await d.accionLista('clientes', [c.id], 'suprimir').catch(() => 0);
			if (n) {
				borrado = c.id;
				break;
			}
		}
		expect(borrado).not.toBeNull();
		expect((await filas(d)).some((x) => x.id === borrado)).toBe(false);
	});
});
