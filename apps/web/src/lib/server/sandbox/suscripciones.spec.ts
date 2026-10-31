import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import { afterEach, describe, expect, it } from 'vitest';
import * as v from 'valibot';
import { PlanSuscripcion } from '../suscripciones-esquemas.ts';
import { d1Local } from './d1-local.ts';
import { crear, DatosSandbox, dbSandbox, guardarCambio } from './sandbox.ts';
import { suscripcionesDemo } from './suscripciones.ts';

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

async function demo(plantilla = 'malbec', rubro = 'bebidas') {
	d1 = d1Local(migraciones);
	const db = dbSandbox(d1);
	const sb = (await crear(db, rubro, Date.now(), plantilla))!;
	return { datos: new DatosSandbox(db, sb), db, sb };
}

const alta = {
	planId: 'club-explorador',
	nombre: 'Julia Paz',
	email: 'julia@mail.com',
	telefono: '11 5555 4444',
	entrega: 'envio' as const,
	direccion: 'Av. Corrientes 1234',
	cp: '1043'
};

describe('suscripciones de la demo', () => {
	it('Malbec arranca con el club de vinos del diseño y Mostaza con el café; las demás sin planes', () => {
		const ahora = Date.UTC(2026, 9, 20, 15);
		expect(suscripcionesDemo('malbec', ahora).planes.map((p) => [p.nombre, p.precio])).toEqual([
			['Descubrir', 64000],
			['Explorador', 118000],
			['Coleccionista', 196000]
		]);
		expect(suscripcionesDemo('mostaza', ahora).planes[0].frecuencia).toBe('quincenal');
		expect(suscripcionesDemo('ceibo', ahora).planes).toEqual([]);
	});

	it('la tienda muestra los planes con sus productos y suscribirse no cobra', async () => {
		const { datos } = await demo();
		const t = await datos.tiendaPublica();
		expect(t.suscripciones.map((p) => p.nombre)).toEqual([
			'Descubrir',
			'Explorador',
			'Coleccionista'
		]);
		expect(t.suscripciones[1].productos[0]).toMatchObject({
			nombre: 'Malbec Valle de Uco 2021',
			cantidad: 2
		});
		const r = await datos.suscribirPublico(alta);
		if (typeof r === 'string') throw new Error(r);
		const s = await datos.suscripcionPublica(r.codigo);
		expect(s).toMatchObject({
			plan: 'Explorador',
			precio: 118000,
			estado: 'activa',
			cobros: 0,
			nombre: 'Julia'
		});
		const panel = await datos.suscripcionesPanel();
		expect(panel.suscriptores[0]).toMatchObject({ email: 'julia@mail.com', estado: 'activa' });
		expect(panel.planes.find((p) => p.id === 'club-explorador')!.suscriptores).toBe(3);
	});

	it('pausar, reanudar y cancelar desde el link privado', async () => {
		const { datos } = await demo();
		const r = await datos.suscribirPublico(alta);
		if (typeof r === 'string') throw new Error(r);
		expect(await datos.cambiarSuscripcionPublica(r.codigo, 'reanudar')).toBe('estado');
		expect(await datos.cambiarSuscripcionPublica(r.codigo, 'pausar')).toMatchObject({
			estado: 'pausada',
			proximoCobro: null
		});
		expect(await datos.cambiarSuscripcionPublica(r.codigo, 'reanudar')).toMatchObject({
			estado: 'activa'
		});
		expect(await datos.cambiarSuscripcionPublica(r.codigo, 'cancelar')).toMatchObject({
			estado: 'cancelada'
		});
		// Cancelar una cancelada no es error.
		expect(await datos.cambiarSuscripcionPublica(r.codigo, 'cancelar')).toMatchObject({
			estado: 'cancelada'
		});
		expect(await datos.cambiarSuscripcionPublica('no-existe-este-codigo', 'pausar')).toBe(
			'no-existe'
		);
	});

	it('el panel crea, edita y borra planes; con suscriptores no se borra; el cupo se respeta', async () => {
		const { datos } = await demo();
		const nuevo = v.parse(PlanSuscripcion, {
			nombre: 'Caja de espumantes',
			descripcion: '',
			foto: null,
			precio: 50000,
			frecuencia: 'bimestral',
			productos: [{ productoId: 'espumante-nature', varianteId: '', cantidad: 2 }],
			incluye: '',
			envio: true,
			retiro: false,
			cupo: 1,
			activo: true
		});
		const p = await datos.guardarPlanSuscripcion(null, nuevo);
		if (!p || p === 'invalido') throw new Error('no creó el plan');
		expect(p.orden).toBe(3);
		expect(
			await datos.guardarPlanSuscripcion(null, {
				...nuevo,
				productos: [{ productoId: 'no-existe', varianteId: '', cantidad: 1 }]
			})
		).toBe('invalido');
		// Retiro no se ofrece en este plan.
		expect(await datos.suscribirPublico({ ...alta, planId: p.id, entrega: 'retiro' })).toBe(
			'invalido'
		);
		expect(typeof (await datos.suscribirPublico({ ...alta, planId: p.id }))).toBe('object');
		expect(await datos.suscribirPublico({ ...alta, planId: p.id, email: 'otra@mail.com' })).toBe(
			'cupo'
		);
		expect(await datos.borrarPlanSuscripcion(p.id)).toBe('no-se-puede');
		const pausado = await datos.guardarPlanSuscripcion(p.id, { ...nuevo, activo: false });
		expect(pausado).toMatchObject({ activo: false, suscriptores: 1 });
		expect((await datos.tiendaPublica()).suscripciones.some((x) => x.id === p.id)).toBe(false);
		expect(await datos.borrarPlanSuscripcion('club-coleccionista')).toBe('no-se-puede');
	});

	it('sin Mercado Pago conectado la tienda no ofrece planes (como en Go)', async () => {
		const { datos, db, sb } = await demo();
		await guardarCambio(db, sb.id, 'cuenta', 'mp', { conectada: false, desde: Date.now() });
		expect((await datos.tiendaPublica()).suscripciones).toEqual([]);
		expect((await datos.suscripcionesPanel()).mercadoPago).toBe(false);
	});

	it('la tienda cancela un suscriptor desde el panel', async () => {
		const { datos } = await demo();
		const x = (await datos.suscripcionesPanel()).suscriptores.find((s) => s.estado === 'activa')!;
		expect(await datos.cancelarSuscriptor(x.id)).toMatchObject({
			estado: 'cancelada',
			canceladaPor: 'tienda'
		});
		expect(await datos.cancelarSuscriptor('no-existe')).toBe('no-existe');
	});
});
