import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import { afterEach, describe, expect, it } from 'vitest';
import { cotizar } from '#lib/tienda/envio.ts';
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

/** El error 400 de SvelteKit con su mensaje. */
const mensaje = (p: Promise<unknown>) =>
	p.then(
		() => null,
		(e: { status?: number; body?: { message?: string } }) => [e.status, e.body?.message]
	);

describe('Envíos y retiro (demo, como configuracion.Entregas en Go)', () => {
	it('arranca con retiro (la dirección del negocio) y sin moto; lo guardado llega a la tienda', async () => {
		const datos = await moda();
		const conf = await datos.configuracion();
		const e = await datos.entregas();
		expect(e).toMatchObject({ retiro: true, moto: false, correo: false });
		expect(e.retiroDireccion).toBe(conf.negocio.direccion ?? '');
		let t = await datos.tiendaPublica(false);
		expect(t.envios.moto).toBe(false);
		expect(t.envios.retiro).toBe(true);

		await datos.guardarEntregas({
			retiro: false,
			retiroDireccion: '',
			retiroHorario: '',
			moto: true,
			motoCosto: 3500,
			motoZona: 'Palermo y Belgrano',
			motoDemora: 'En el día'
		});
		t = await datos.tiendaPublica(false);
		expect(t.envios).toMatchObject({
			retiro: false,
			retiroDireccion: '',
			moto: true,
			motoCosto: 3500,
			motoZona: 'Palermo y Belgrano'
		});
		// El checkout ofrece sólo la moto, con su costo (o gratis desde el monto de la promoción).
		expect(cotizar('1414', 1000, t)).toEqual([
			{
				id: 'moto',
				nombre: 'Envío en moto',
				detalle: 'Llega a Palermo y Belgrano · En el día',
				precio: 3500
			}
		]);
		expect(t.envios.gratisDesde).not.toBeNull();
		expect(cotizar('1414', t.envios.gratisDesde!, t)[0].precio).toBe(0);
	});

	it('no deja apagar todo (la demo no tiene Correo) ni la moto sin zona', async () => {
		const datos = await moda();
		const base = {
			retiro: false,
			retiroDireccion: '',
			retiroHorario: '',
			moto: false,
			motoCosto: 0,
			motoZona: '',
			motoDemora: ''
		};
		expect(await mensaje(datos.guardarEntregas(base))).toEqual([
			400,
			'Dejá al menos una forma de entrega: retiro, moto o Correo Argentino.'
		]);
		expect(await mensaje(datos.guardarEntregas({ ...base, moto: true, motoCosto: 100 }))).toEqual([
			400,
			'Poné hasta dónde llega la moto (barrios o localidades).'
		]);
		expect((await datos.entregas()).retiro).toBe(true);
	});
});

describe('Cambios y devoluciones (demo, como configuracion.Negocio.Politica en Go)', () => {
	it('arranca vacía y lo guardado en Negocio llega a la tienda', async () => {
		const datos = await moda();
		expect((await datos.tiendaPublica(false)).politica).toBe('');
		const conf = await datos.configuracion();
		const politica = 'Cambios dentro de los 30 días con el ticket.';
		await datos.guardarNegocio({ ...conf.negocio, politica });
		expect((await datos.configuracion()).negocio.politica).toBe(politica);
		expect((await datos.tiendaPublica(false)).politica).toBe(politica);
	});
});
