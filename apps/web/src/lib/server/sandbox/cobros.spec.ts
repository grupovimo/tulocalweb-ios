import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import { afterEach, describe, expect, it } from 'vitest';
import { d1Local } from './d1-local.ts';
import { crear, DatosSandbox, dbSandbox } from './sandbox.ts';

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
async function almacen() {
	d1 = d1Local(migraciones);
	const db = dbSandbox(d1);
	const sb = (await crear(db, 'almacen', Date.now(), 'yerbal'))!;
	return new DatosSandbox(db, sb);
}

const mensaje = (p: Promise<unknown>) =>
	p.then(
		() => null,
		(e: { status?: number; body?: { message?: string } }) => [e.status, e.body?.message]
	);

describe('cuotas sin interés y precio sin impuestos en la demo (como Go)', () => {
	it('arranca con 6 cuotas desde $ 30.000 y, responsable inscripta, con el precio sin impuestos', async () => {
		const datos = await almacen();
		expect((await datos.configuracion()).cobros).toEqual({
			cuotas: 6,
			cuotasDesde: 30_000,
			impuestos: { mostrar: true },
			mercadoPago: true
		});
		const t = await datos.tiendaPublica(false);
		expect(t.cobros).toMatchObject({
			cuotas: 6,
			cuotasDesde: 30_000,
			impuestos: { mostrar: true }
		});
		for (const p of t.productos) {
			expect([0, 10.5, 21, 27]).toContain(p.iva);
			for (const v of p.variantes) expect(v.iva).toBe(p.iva);
		}
		const a = (await datos.cupones()).automaticas.find((x) => x.id === 'cuotas')!;
		expect(a.detalle).toBe('6 cuotas sin interés desde $ 30.000, con Mercado Pago');
	});

	it('se guarda, «Ninguna» apaga la automática y prenderla vuelve a las elegidas', async () => {
		const datos = await almacen();
		expect(
			await mensaje(
				datos.guardarCobros({ cuotas: 4 as 3, cuotasDesde: 0, impuestos: { mostrar: true } })
			)
		).toEqual([400, 'Elegí ninguna, 3, 6, 9 o 12 cuotas.']);
		await datos.guardarCobros({ cuotas: 12, cuotasDesde: 80_000, impuestos: { mostrar: false } });
		let t = await datos.tiendaPublica(false);
		expect(t.cobros).toMatchObject({
			cuotas: 12,
			cuotasDesde: 80_000,
			impuestos: { mostrar: false }
		});
		await datos.guardarCobros({ cuotas: 0, cuotasDesde: 80_000, impuestos: { mostrar: false } });
		t = await datos.tiendaPublica(false);
		expect([t.cobros.cuotas, t.cobros.cuotasDesde]).toEqual([0, 0]);
		await datos.fijarAutomatica('cuotas', true);
		expect((await datos.tiendaPublica(false)).cobros.cuotas).toBe(12);
	});

	it('sin Mercado Pago no se ofrecen cuotas', async () => {
		const datos = await almacen();
		await datos.desconectarMercadoPago();
		const t = await datos.tiendaPublica(false);
		expect([t.cobros.cuotas, t.cobros.cuotasDesde]).toEqual([0, 0]);
	});

	it('el IVA del producto se guarda y llega a la vidriera', async () => {
		const datos = await almacen();
		const p = (await datos.guardarProducto(null, {
			nombre: 'Yerba suelta',
			precio: 5000,
			iva: 10.5
		}))!;
		expect(p.iva).toBe(10.5);
		const otro = (await datos.guardarProducto(null, { nombre: 'Mate', precio: 9000 }))!;
		expect(otro.iva).toBe(21);
		const editado = (await datos.guardarProducto(p.id, { nombre: 'Yerba suelta', precio: 5200 }))!;
		expect(editado.iva).toBe(10.5);
	});
});
