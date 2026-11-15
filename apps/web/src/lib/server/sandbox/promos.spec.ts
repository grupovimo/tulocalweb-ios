import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import { afterEach, describe, expect, it } from 'vitest';
import CASOS from '@mostrador/contrato/casos/ventanas.json';
import { d1Local } from './d1-local.ts';
import { crear, DatosSandbox, dbSandbox } from './sandbox.ts';
import {
	nombrePila,
	premioCompleto,
	promosDemo,
	provinciaDe,
	rango,
	sortear,
	ventanaCompleta,
	type Ventana
} from './promos.ts';

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
async function demo(rubro: string, plantilla: string) {
	d1 = d1Local(migraciones);
	const db = dbSandbox(d1);
	const sb = (await crear(db, rubro, Date.now(), plantilla))!;
	return new DatosSandbox(db, sb);
}

describe('ventanas: completas como en Go (casos/ventanas.json)', () => {
	for (const c of CASOS.ventanas.filter((x) => x.valida))
		it(c.nombre, () => {
			expect(ventanaCompleta(c.ventana as unknown as Ventana)).toBe(c.completa);
		});
	it('premio', () => {
		expect(premioCompleto({ tipo: 'cuponUnico', descuento: 'envioGratis' })).toBe(true);
		expect(premioCompleto({ tipo: 'cuponUnico', descuento: 'porcentaje' })).toBe(false);
		expect(premioCompleto(null)).toBe(false);
	});
});

describe('reglas compartidas con Go', () => {
	it('el sorteo respeta los pesos', () => {
		expect(sortear([0, 0])).toBe(-1);
		const cuenta = [0, 0, 0];
		for (let i = 0; i < 3000; i++) cuenta[sortear([70, 0, 30])]++;
		expect(cuenta[1]).toBe(0);
		expect(cuenta[0]).toBeGreaterThan(1800);
	});
	it('nombre de pila, provincia y rango', () => {
		expect(nombrePila('ana maría gómez')).toBe('Ana');
		expect(nombrePila('Consumidor final')).toBe('');
		expect(provinciaDe('Lanús')).toBe('Buenos Aires');
		expect(provinciaDe('', 'X5000ABC')).toBe('Córdoba');
		expect(provinciaDe('Springfield')).toBe('');
		expect(rango(null, null, Date.UTC(2026, 10, 15, 2))).toEqual({
			desde: '2026-10-16',
			hasta: '2026-11-14'
		});
		expect(rango('2026-11-10', '2026-11-01', Date.now())).toBeNull();
	});
	it('las demos estrenan sus ventanas', () => {
		expect(promosDemo('ceibo', 40000, Date.now()).ventanas[0]).toMatchObject({
			id: 'bienvenida',
			contenido: 'captacion'
		});
		expect(promosDemo('mostaza', 40000, Date.now()).ventanas[0]).toMatchObject({
			contenido: 'ruleta'
		});
		const verde = promosDemo('yerbal', 40000, Date.now());
		expect(verde.conversion?.pruebaSocial?.activa).toBe(true);
		expect((verde.ventanas[0].datos as { titulo: string }).titulo).toContain('40.000');
		for (const v of [
			...promosDemo('ceibo', 1, 0).ventanas,
			...promosDemo('mostaza', 1, 0).ventanas
		])
			expect(ventanaCompleta(v)).toBe(true);
	});
});

describe('promociones en la demo', () => {
	it('Lumbre: captar una vez por email, el cupón sirve una vez y suma a la newsletter y a las métricas', async () => {
		const datos = await demo('moda', 'ceibo');
		const t = await datos.tiendaPublica(false);
		const v = (t.diseno as { ventanas?: Ventana[] }).ventanas!;
		expect(v[0].id).toBe('bienvenida');
		expect(t.topVendidos?.length).toBeGreaterThan(0);
		expect(t.productos[0].alta).toBeTruthy();

		expect(await datos.participarPromo('ruleta', { ventana: 'bienvenida', email: 'a@b.com' })).toBe(
			'no-existe'
		);
		expect(await datos.participarPromo('captacion', { ventana: 'bienvenida', email: 'no' })).toBe(
			'invalido'
		);
		const r = await datos.participarPromo('captacion', {
			ventana: 'bienvenida',
			email: 'Lu@Ejemplo.com',
			nombre: 'Lu'
		});
		if (typeof r === 'string') throw new Error(r);
		expect(r.premio).toMatchObject({ tipo: 'cuponUnico', descripcion: '10 % off' });
		expect(r.premio.codigo).toMatch(/^BIENVE-[A-Z0-9]{8}$/);
		const otra = await datos.participarPromo('captacion', {
			ventana: 'bienvenida',
			email: 'lu@ejemplo.com'
		});
		expect(otra).toMatchObject({ repetido: true, premio: { codigo: r.premio.codigo } });

		const cupon = await datos.cuponTienda(r.premio.codigo!);
		expect(cupon).toMatchObject({ tipo: 'pct', valor: 10, limite: 1, usos: 0 });
		expect((await datos.cupones()).cupones.some((c) => c.codigo === r.premio.codigo)).toBe(false);
		const n = await datos.newsletter();
		expect(n.suscriptores[0]).toMatchObject({
			email: 'lu@ejemplo.com',
			origen: 'ventana:bienvenida',
			nombre: 'Lu'
		});

		await datos.eventosPromo([
			{ ventana: 'bienvenida', tipo: 'vista', n: 4 },
			{ ventana: 'bienvenida', tipo: 'cierre' },
			{ ventana: 'otraventana', tipo: 'vista' }
		]);
		const m = (await datos.metricasPromos())!;
		expect(m.ventanas[0]).toMatchObject({
			id: 'bienvenida',
			vistas: 4,
			cierres: 1,
			envios: 1,
			tasaEnvio: 0.25
		});
		expect(m.ventanas).toHaveLength(1);
		expect(await datos.metricasPromos('2026-02-01', '2026-01-01')).toBeNull();

		// La vitrina muestra el premio y no guarda nada.
		const vit = await datos.participarPromo(
			'captacion',
			{ ventana: 'bienvenida', email: 'x@y.com' },
			true
		);
		expect(vit).toMatchObject({ demo: true, premio: { codigo: 'DEMO-BIEN' } });
	});

	it('Café Bruma: la ruleta cae en una porción y es una por email', async () => {
		const datos = await demo('gastronomia', 'mostaza');
		const t = await datos.tiendaPublica(false);
		const id = (t.diseno as { ventanas?: Ventana[] }).ventanas![0].id;
		const g = await datos.participarPromo('ruleta', { ventana: id, email: 'beto@example.com' });
		if (typeof g === 'string') throw new Error(g);
		expect(g.porcion).toBeGreaterThanOrEqual(0);
		expect(g.texto).toBeTruthy();
		const g2 = await datos.participarPromo('ruleta', { ventana: id, email: 'beto@example.com' });
		expect(g2).toMatchObject({ repetido: true, porcion: g.porcion });
	});

	it('Almacén Verde: prueba social con compras pagadas de la semana', async () => {
		const datos = await demo('almacen', 'yerbal');
		const c = await datos.promosRecientes();
		expect(c.length).toBeGreaterThan(0);
		expect(c[0].nombre).toMatch(/^\p{Lu}/u);
		expect(Date.now() - Date.parse(c[0].cuando)).toBeLessThan(7 * 86_400_000);
		// Sin la prueba social prendida, nada.
		expect(await (await demo('moda', 'ceibo')).promosRecientes()).toEqual([]);
	});
});
