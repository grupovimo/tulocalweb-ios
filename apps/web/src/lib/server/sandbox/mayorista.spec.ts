import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import { afterEach, describe, expect, it } from 'vitest';
import { detalleLineas } from '#lib/tienda/lineas.ts';
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

describe('venta mayorista en la demo (como Go, §49)', () => {
	it('arranca prendida con 12 unidades y la mitad del catálogo con precio mayorista', async () => {
		const datos = await almacen();
		expect((await datos.configuracion()).mayorista).toEqual({
			activa: true,
			unidades: 12,
			monto: null
		});
		const t = await datos.tiendaPublica(false);
		expect(t.mayorista).toEqual({ unidades: 12, monto: null });
		const con = t.productos.filter((p) => p.mayorista);
		expect(con.length).toBeGreaterThan(0);
		for (const p of con) expect(p.mayorista!).toBeLessThan(p.precio);
		const a = (await datos.cupones()).automaticas.find((x) => x.id === 'mayorista')!;
		expect(a.detalle).toMatch(/^Desde 12\u00a0unidades, con el precio mayorista/);
	});

	it('el mínimo se valida, se guarda y apagada no publica nada', async () => {
		const datos = await almacen();
		expect(
			await mensaje(datos.guardarMayorista({ activa: true, unidades: null, monto: null }))
		).toEqual([400, 'Poné un mínimo de unidades, de monto o los dos.']);
		await datos.guardarMayorista({ activa: true, unidades: null, monto: 50_000 });
		expect((await datos.tiendaPublica(false)).mayorista).toEqual({ unidades: null, monto: 50_000 });
		await datos.guardarMayorista({ activa: false, unidades: 6, monto: null });
		const t = await datos.tiendaPublica(false);
		expect(t.mayorista).toBe(null);
		expect(t.productos.every((p) => !p.mayorista)).toBe(true);
		// Es el mismo interruptor que la automática de Cupones.
		expect((await datos.cupones()).automaticas.find((x) => x.id === 'mayorista')!.activa).toBe(
			false
		);
	});

	it('la ficha guarda el mayorista (null deja, 0 saca) y el pedido sale marcado con esos precios', async () => {
		const datos = await almacen();
		const t0 = await datos.tiendaPublica(false);
		const p = t0.productos.find((x) => !x.opciones.length && x.stock >= 12)!;
		const ficha = (await datos.producto(p.id))!;
		const precio = ficha.precio;
		const mayor = Math.round((precio * 0.8) / 10) * 10;
		const editable = { nombre: ficha.nombre, precio, categoria: ficha.categoria };
		expect(
			(await datos.guardarProducto(p.id, { ...editable, precioMayorista: mayor }))!.precioMayorista
		).toBe(mayor);
		expect((await datos.guardarProducto(p.id, editable))!.precioMayorista).toBe(mayor);

		const t = await datos.tiendaPublica(false);
		const lineas = detalleLineas(
			[{ id: p.id, varianteId: null, cantidad: 12 }],
			t.productos,
			t.mayorista ?? null
		);
		expect(lineas[0]).toMatchObject({ alPorMayor: true, subtotal: mayor * 12 });
		const r = await datos.crearPedido(
			{
				items: [{ productoId: p.id, varianteId: null, cantidad: 12 }],
				pago: 'transf',
				canal: 'web',
				envio: 'retiro',
				pagado: false
			},
			{
				email: 'lucia@ejemplo.com',
				cupon: null,
				descuentoCupon: 0,
				total: mayor * 12,
				precios: { [p.id]: mayor },
				mayorista: true
			}
		);
		expect(r).toMatchObject({ total: mayor * 12 });
		const lista = await datos.listarPedidos({ mayorista: true });
		expect(lista.filas.some((f) => typeof r === 'object' && f.numero === r.numero)).toBe(true);

		expect(
			(await datos.guardarProducto(p.id, { ...editable, precioMayorista: 0 }))!.precioMayorista
		).toBe(null);
	});
});
