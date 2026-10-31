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
async function demo() {
	d1 = d1Local(migraciones);
	const db = dbSandbox(d1);
	const sb = (await crear(db, 'gastronomia', Date.now()))!;
	return new DatosSandbox(db, sb);
}

const SIN = {
	lanzamiento: null,
	reserva: null,
	encargo: null,
	ofertaPrecio: null,
	ofertaHasta: null
};
const en = (ms: number) => new Date(Date.now() + ms).toISOString();

describe('lanzamientos de la demo (§52)', () => {
	it('«Próximamente» hasta la hora, aviso, reserva con seña que descuenta stock y saldado', async () => {
		const datos = await demo();
		const t0 = await datos.tiendaPublica();
		const p = t0.productos.find((x) => !x.opciones.length && x.stock >= 2)!;
		// Las mismas reglas que Go.
		expect(await datos.guardarLanzamiento(p.id, { ...SIN, lanzamiento: en(-60_000) })).toEqual({
			error: 'La fecha del lanzamiento tiene que ser más adelante que ahora.'
		});
		expect(await datos.guardarLanzamiento(p.id, { ...SIN, reserva: 30 })).toEqual({
			error: 'Para reservar con seña, poné la fecha del lanzamiento.'
		});
		const l = await datos.guardarLanzamiento(p.id, {
			...SIN,
			lanzamiento: en(3_600_000),
			reserva: 30
		});
		expect(l).toMatchObject({ reserva: 30, avisos: 0, reservas: 0 });

		const t = await datos.tiendaPublica();
		const pub = t.productos.find((x) => x.id === p.id)!;
		expect(pub).toMatchObject({ etiqueta: 'Próximamente', reserva: 30 });
		expect(pub.lanzamiento).toBeTruthy();

		expect(await datos.avisarLanzamiento(p.id, 'Ana@Ejemplo.com')).toBe('ok');
		expect(await datos.avisarLanzamiento(p.id, 'ana@ejemplo.com')).toBe('ok');
		expect((await datos.lanzamiento(p.id))!.avisos).toBe(1);

		const r = await datos.senar({
			productoId: p.id,
			varianteId: null,
			cantidad: 1,
			nombre: 'Lucía',
			email: 'lu@ejemplo.com',
			telefono: ''
		});
		if (typeof r === 'string') throw new Error(r);
		expect(r.sena).toBe(Math.round((pub.precio * 30) / 100));
		const d = (await datos.pedido(r.numero))!;
		expect(d.sena).toMatchObject({ tipo: 'reserva', pagada: true, saldado: false });
		expect(d.sena!.falta).toBe(pub.precio - r.sena);
		expect((await datos.tiendaPublica()).productos.find((x) => x.id === p.id)!.stock).toBe(
			pub.stock - 1
		);
		expect((await datos.lanzamiento(p.id))!.reservas).toBe(1);

		expect(await datos.saldarPedido(r.numero, true)).toBe('ok');
		expect((await datos.pedido(r.numero))!.sena).toMatchObject({ saldado: true, falta: 0 });
		expect((await datos.pedido(r.numero))!.estadoPago).toBe('Cobrado');
		expect(await datos.saldarPedido(r.numero, false)).toBe('ok');
		expect((await datos.pedido(r.numero))!.estadoPago).toBe('Seña');
	});

	it('la oferta con fin baja el precio y vuelve al de lista; el encargo pide seña ya lanzado', async () => {
		const datos = await demo();
		const p = (await datos.tiendaPublica()).productos.find(
			(x) => !x.opciones.length && !x.anterior
		)!;
		const oferta = Math.round(p.precio * 0.8);
		expect(
			await datos.guardarLanzamiento(p.id, { ...SIN, ofertaPrecio: p.precio + 1 })
		).toMatchObject({ error: expect.stringContaining('menor') });
		await datos.guardarLanzamiento(p.id, {
			...SIN,
			encargo: 50,
			ofertaPrecio: oferta,
			ofertaHasta: en(86_400_000)
		});
		const pub = (await datos.tiendaPublica()).productos.find((x) => x.id === p.id)!;
		expect(pub).toMatchObject({ precio: oferta, anterior: p.precio, encargo: 50 });
		expect(pub.reserva ?? null).toBeNull();
		expect(await datos.avisarLanzamiento(p.id, 'ana@ejemplo.com')).toBe('ya-salio');
		const r = await datos.senar({
			productoId: p.id,
			varianteId: null,
			cantidad: 2,
			nombre: 'Ana',
			email: 'ana@ejemplo.com',
			telefono: ''
		});
		if (typeof r === 'string') throw new Error(r);
		expect(r.total).toBe(oferta * 2);
		expect((await datos.pedido(r.numero))!.items[0].precio).toBe(oferta);
	});
});
