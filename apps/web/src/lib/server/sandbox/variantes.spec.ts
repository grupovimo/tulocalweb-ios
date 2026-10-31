import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import { afterEach, describe, expect, it } from 'vitest';
import { CATALOGOS } from '#lib/catalogos.ts';
import { d1Local } from './d1-local.ts';
import { crear, DatosSandbox, dbSandbox } from './sandbox.ts';
import { variantesDemo } from './variantes.ts';

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
async function moda() {
	d1 = d1Local(migraciones);
	const db = dbSandbox(d1);
	const sb = (await crear(db, 'moda', Date.now(), 'ceibo'))!;
	return new DatosSandbox(db, sb);
}

describe('variantes de la demo', () => {
	it('cada plantilla que vende productos muestra alguno con variantes; la de turnos, no', () => {
		const con = (p: string) => variantesDemo(CATALOGOS.find((c) => c.plantilla === p)!);
		for (const c of CATALOGOS.filter((x) => x.plantilla !== 'tijera'))
			expect(con(c.plantilla).size, c.plantilla).toBeGreaterThan(0);
		expect(con('tijera').size).toBe(0);
		const zapa = con('pampa').get('zapa-runner-90')!;
		expect(zapa.opciones).toEqual([
			{ nombre: 'Número', valores: ['38', '39', '40', '41', '42', '43', '44'] }
		]);
		const cafe = con('mostaza').get('cafe-yirgacheffe')!;
		expect(cafe.variantes).toHaveLength(9);
		expect(
			cafe.variantes.find((v) => v.valores['Presentación'] === '1 kg')!.precio
		).toBeGreaterThan(21_900);
		// Siempre iguales: salen del id del producto.
		expect(con('ceibo')).toEqual(con('ceibo'));
	});

	it('la ficha trae opciones, galería y stock por variante; el total es la suma', async () => {
		const datos = await moda();
		const p = (await datos.producto('remera-pima'))!;
		expect(p.opciones.map((o) => o.nombre)).toEqual(['Talle', 'Color']);
		expect(p.variantes).toHaveLength(12);
		expect(p.fotos.length).toBeGreaterThan(1);
		expect(p.foto).toBe(p.fotos[0]);
		expect(p.variantes.filter((v) => v.valores.Talle === 'S').every((v) => v.stock === 0)).toBe(
			true
		);
		expect(p.stock).toBe(p.variantes.reduce((s, v) => s + v.stock, 0));
		const t = await datos.tiendaPublica();
		const pub = t.productos.find((x) => x.id === 'remera-pima')!;
		expect(pub.variantes).toHaveLength(12);
		expect(Object.keys(pub.variantes[0]).sort()).toEqual([
			'foto',
			'id',
			'mayorista',
			'precio',
			'stock',
			'valores'
		]);
	});

	it('vender descuenta de la variante y exige una variante activa del producto', async () => {
		const datos = await moda();
		const p = (await datos.producto('remera-pima'))!;
		const v = p.variantes.find((x) => x.stock >= 2 && x.valores.Talle === 'XL')!;
		const otro = (await datos.producto('tote-de-cuero'))!;
		const base = {
			pago: 'efectivo' as const,
			canal: 'pos' as const,
			envio: 'retiro' as const,
			pagado: true
		};
		expect(await datos.crearPedido({ ...base, items: [{ productoId: p.id, cantidad: 1 }] })).toBe(
			'invalido'
		);
		expect(
			await datos.crearPedido({
				...base,
				items: [{ productoId: p.id, varianteId: otro.variantes[0].id, cantidad: 1 }]
			})
		).toBe('invalido');
		expect(
			await datos.crearPedido({
				...base,
				items: [{ productoId: p.id, varianteId: v.id, cantidad: v.stock + 1 }]
			})
		).toBe('sin-stock');
		const r = await datos.crearPedido({
			...base,
			items: [{ productoId: p.id, varianteId: v.id, cantidad: 2 }]
		});
		expect(r).toMatchObject({ total: 2 * v.precio! });
		const despues = (await datos.producto('remera-pima'))!;
		expect(despues.variantes.find((x) => x.id === v.id)!.stock).toBe(v.stock - 2);
		expect(despues.stock).toBe(p.stock! - 2);
		const pedido = (await datos.pedido((r as { numero: number }).numero))!;
		expect(pedido.items[0]).toMatchObject({
			varianteId: v.id,
			variante: `Talle XL · Color ${v.valores.Color}`
		});
		// Un producto sin opciones no acepta variante.
		const lentes = (await datos.listar('productos', {})).filas as unknown as {
			id: string;
			variantes: { id: string; texto: string }[];
		}[];
		expect(lentes.find((x) => x.id === 'remera-pima')!.variantes[0].texto).toBe(
			'Talle S · Color Blanco'
		);
	});

	it('editar opciones rearma las combinaciones; inactiva no se vende ni se publica', async () => {
		const datos = await moda();
		const p = (await datos.producto('remera-pima'))!;
		const m = p.variantes.find((v) => v.valores.Talle === 'M' && v.valores.Color === 'Negro')!;
		const g = (await datos.guardarProducto(p.id, {
			nombre: p.nombre,
			precio: p.precio,
			fotos: [p.fotos[1], p.fotos[0]],
			opciones: [
				{ nombre: 'Talle', valores: ['M', 'L'] },
				{ nombre: 'Color', valores: ['Negro', 'Verde'] }
			],
			// Sólo la que se toca, por sus valores (el id no cuenta).
			variantes: [{ valores: m.valores, stock: 7, activa: false }]
		}))!;
		expect(g.foto).toBe(p.fotos[1]);
		expect(g.variantes).toHaveLength(4);
		const mNegro = g.variantes.find((v) => v.id === m.id)!;
		expect(mNegro).toMatchObject({ stock: 7, activa: false });
		const verde = g.variantes.filter((v) => v.valores.Color === 'Verde');
		expect(verde.every((v) => v.stock === 0 && v.activa && v.id)).toBe(true);
		// El total sólo suma las activas.
		expect(g.stock).toBe(g.variantes.filter((v) => v.activa).reduce((s, v) => s + v.stock, 0));
		expect(
			await datos.crearPedido({
				pago: 'efectivo',
				canal: 'pos',
				envio: 'retiro',
				pagado: true,
				items: [{ productoId: p.id, varianteId: m.id, cantidad: 1 }]
			})
		).toBe('invalido');
		const pub = (await datos.tiendaPublica()).productos.find((x) => x.id === p.id)!;
		expect(pub.variantes.some((v) => v.id === m.id)).toBe(false);
		// Sacar las opciones lo deja sin variantes.
		const sin = (await datos.guardarProducto(p.id, {
			nombre: p.nombre,
			precio: p.precio,
			stock: 4,
			opciones: [],
			variantes: []
		}))!;
		expect(sin).toMatchObject({ variantes: [], opciones: [], stock: 4 });
	});

	it('inventario: se ajusta y se cuenta por variante', async () => {
		const datos = await moda();
		const inv = await datos.listar('inventario', { q: 'Remera pima' });
		const fila = inv.filas[0] as {
			id: string;
			variantes: { id: string; stock: number; texto: string }[];
		};
		const v = fila.variantes.find((x) => x.texto === 'Talle M · Color Negro')!;
		expect(await datos.accionLista('inventario', [v.id], 'ajustar', 5)).toBe(1);
		// El producto con opciones no se ajusta entero: hay que decir qué variante.
		expect(await datos.accionLista('inventario', [fila.id], 'ajustar', 5)).toBe(0);
		expect(await datos.accionLista('inventario', [], 'conteo', [{ id: v.id, cantidad: 40 }])).toBe(
			1
		);
		const p = (await datos.producto('remera-pima'))!;
		expect(p.variantes.find((x) => `remera-pima:${x.id}` === v.id)!.stock).toBe(40);
		const movs = await datos.movimientosHoy();
		expect(
			movs.some((m) => m.texto.includes('Remera') && m.texto.includes('Talle M · Color Negro'))
		).toBe(true);
	});
});
