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
async function moda() {
	d1 = d1Local(migraciones);
	const db = dbSandbox(d1);
	const sb = (await crear(db, 'moda', Date.now(), 'ceibo'))!;
	return new DatosSandbox(db, sb);
}

describe('carritos de la tienda (demo)', () => {
	it('se guarda uno por email, aparece en el panel, se restaura por token y se cierra al comprar', async () => {
		const datos = await moda();
		const tienda = await datos.tiendaPublica(false);
		const remera = tienda.productos.find((p) => p.id === 'remera-pima')!;
		const v = remera.variantes.find((x) => x.stock > 0)!;
		const antes = await datos.carritos();

		// Una variante que no es del producto se ignora; si no queda nada, no se guarda.
		expect(
			await datos.guardarCarritoTienda('lu@ejemplo.com', 'Lu', [
				{ id: 'remera-pima', varianteId: 'otra', cantidad: 1 }
			])
		).toBe(false);
		expect(
			await datos.guardarCarritoTienda('Lu@Ejemplo.com', 'Lucía', [
				{ id: 'remera-pima', varianteId: v.id, cantidad: 2 },
				{ id: 'no-existe', varianteId: null, cantidad: 1 }
			])
		).toBe(true);
		// El mismo email (sin importar mayúsculas) actualiza el mismo carrito.
		await datos.guardarCarritoTienda('lu@ejemplo.com', 'Lucía Pérez', [
			{ id: 'remera-pima', varianteId: v.id, cantidad: 3 }
		]);

		const c = await datos.carritos();
		expect(c.abiertos).toBe(antes.abiertos + 1);
		expect(c.carritos[0]).toMatchObject({
			nombre: 'Lucía Pérez',
			desde: 'la tienda',
			cantidad: 3,
			total: (v.precio ?? remera.precio) * 3,
			recordado: false
		});
		expect(await datos.recordarCarrito(c.carritos[0].id)).toBe('ok');
		expect((await datos.carritos()).carritos[0].recordado).toBe(true);

		// El token del link vuelve a armar el carrito; uno inventado, no.
		const { results } = await d1!
			.prepare("select datos from cambios where entidad = 'carrito-tienda'")
			.all<{ datos: string }>();
		const fila = JSON.parse(results[0].datos) as { token: string };
		expect(await datos.carritoTienda(fila.token)).toEqual([
			{ id: 'remera-pima', varianteId: v.id, cantidad: 3 }
		]);
		expect(await datos.carritoTienda('x'.repeat(32))).toBeNull();

		// Compró: deja de aparecer, el link ya no sirve y suma a lo recuperado.
		await datos.cerrarCarritoTienda('LU@ejemplo.com', 50_000);
		const despues = await datos.carritos();
		expect(despues.abiertos).toBe(antes.abiertos);
		expect(despues.recuperado).toBe(antes.recuperado + 50_000);
		expect(await datos.carritoTienda(fila.token)).toBeNull();
	});

	it('los tres recordatorios (§62): arrancan prendidos en la demo, se validan como Go y la baja corta', async () => {
		const datos = await moda();
		const k = await datos.carritos();
		expect(k.recordatorios.activo).toBe(true);
		expect(k.recordatorios.pasos.map((p) => p.horas)).toEqual([1, 24, 72]);
		expect(k.recordatorios.pasos[1].cupon).toMatchObject({ descuento: 'porcentaje', valor: 10 });
		expect(k.recuperados.monto).toBe(k.recuperado);

		const r = structuredClone(k.recordatorios);
		r.pasos[2].horas = 12;
		expect(await datos.guardarRecordatorios(r)).toBe(
			'Cada recordatorio tiene que salir después del anterior.'
		);
		r.pasos[2].horas = 96;
		r.pasos[0].cupon = { descuento: 'envioGratis', valor: 0, dias: 3 };
		expect(await datos.guardarRecordatorios(r)).toBe(
			'El cupón va en el segundo o en el tercer recordatorio.'
		);
		r.pasos[0].cupon = null;
		r.activo = false;
		expect(await datos.guardarRecordatorios(r)).toBeNull();
		expect((await datos.carritos()).recordatorios).toMatchObject({ activo: false });
		// Las apps de antes prenden con «email».
		await datos.guardarAutomaticos({ hora: false, dia: false, email: true });
		expect((await datos.carritos()).recordatorios.activo).toBe(true);
		expect((await datos.carritos()).recordatorios.pasos[2].horas).toBe(96);

		const tienda = await datos.tiendaPublica(false);
		const remera = tienda.productos.find((x) => x.id === 'remera-pima')!;
		const v = remera.variantes.find((x) => x.stock > 0)!;
		await datos.guardarCarritoTienda('ana@ejemplo.com', 'Ana', [
			{ id: remera.id, varianteId: v.id, cantidad: 1 }
		]);
		const token = (
			(await d1!
				.prepare("SELECT datos AS valor FROM cambios WHERE entidad = 'carrito-tienda'")
				.first()) as {
				valor: string;
			}
		).valor;
		const t = (JSON.parse(token) as { token: string }).token;
		await datos.bajaCarritoTienda(t);
		const fila = (await d1!
			.prepare("SELECT datos AS valor FROM cambios WHERE entidad = 'carrito-tienda'")
			.first()) as {
			valor: string;
		};
		expect((JSON.parse(fila.valor) as { baja?: number }).baja).toBeTypeOf('number');
	});
});
