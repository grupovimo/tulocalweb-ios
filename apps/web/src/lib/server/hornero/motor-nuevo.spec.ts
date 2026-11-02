import { describe, expect, it } from 'vitest';
import { demoDe } from '../sandbox/demo.ts';
import { contextoDemo, tableroDemo } from '../sandbox/tablero.ts';
import type { DatosHornero } from './contexto.ts';
import { responder } from './motor.ts';
import { contextoSinMontos, sinMontos } from '../datos/montos.ts';

const BASE = Date.parse('2026-10-20T15:00:00Z');
const AHORA = BASE + 30 * 60_000;
const d = demoDe('litoral', BASE)!;
const op = { base: BASE, ahora: AHORA, hola: 'Valeria', cambios: {} };
const datos = (): DatosHornero => ({
	tienda: { nombre: 'Mate & Co', rubro: 'almacen' },
	hola: 'Valeria',
	hoy: tableroDemo(d, 'hoy', op),
	semana: tableroDemo(d, '7d', op),
	mes: tableroDemo(d, '30d', op),
	contexto: contextoDemo(d, AHORA)
});
const pesos = (n: number) => '$ ' + Math.round(n).toLocaleString('es-AR');
const r = (q: string, x = datos(), antes: string[] = []) => responder(q, x, antes, new Date(AHORA));

describe('Tincho sin IA: entiende', () => {
	it('errores de tipeo y sinónimos', () => {
		const x = datos();
		for (const q of [
			'cuanto bendi hoy',
			'cuánto facturé hoy',
			'cuanta guita hice hoy',
			'cuanto vendimo hoy',
			'como vienen las ventas'
		])
			expect(r(q, x)?.texto, q).toContain(pesos(x.hoy.ventas!.actual));
		for (const q of [
			'que me falta reponer',
			'que productos estan agotados',
			'qué tengo sin stok',
			'se me termino algo?'
		])
			expect(r(q, x)?.tema, q).toBe('stock');
	});

	it('los períodos: semana pasada, mes, últimos N días, un día de la semana', () => {
		const x = datos();
		const dias = x.contexto.ventasPorDia;
		const pasada = dias.slice(-14, -7).reduce((n, y) => n + (y.ventas ?? 0), 0);
		expect(r('cuanto vendi la semana pasada', x)?.texto).toContain(pesos(pasada));
		expect(r('ventas del mes', x)?.texto).toContain(pesos(x.mes!.ventas!.actual));
		const tres = dias.slice(-3).reduce((n, y) => n + (y.ventas ?? 0), 0);
		expect(r('cuanto vendi en los ultimos 3 dias', x)?.texto).toContain(pesos(tres));
		expect(r('cuanto vendi anteayer', x)?.texto).toContain(pesos(dias.at(-3)!.ventas!));
	});

	it('sigue la charla: «¿y ayer?», «¿y la semana pasada?»', () => {
		const x = datos();
		const a = r('¿y ayer?', x, ['¿cuánto vendí hoy?']);
		expect(a?.texto).toContain(pesos(x.contexto.ventasPorDia.at(-2)!.ventas!));
		const p = x.contexto.productos.find((y) => y.nombre.split(' ').length > 1)!;
		const otro = x.contexto.productos.find(
			(y) => y !== p && y.nombre.split(' ')[0] !== p.nombre.split(' ')[0]
		)!;
		expect(
			r(`¿y ${otro.nombre.toLowerCase()}?`, x, [`cuánto stock queda de ${p.nombre}`])?.texto
		).toContain(otro.nombre);
	});

	it('el resumen del día, pendientes, ideas, clientes, canales, visitas', () => {
		const x = datos();
		expect(r('¿cómo vengo hoy?', x)?.tema).toBe('resumen');
		expect(r('hola', x)?.tema).toBe('resumen');
		expect(r('que tengo pendiente', x)?.tema).toBe('pendientes');
		expect(r('ideas para vender más', x)?.tema).toBe('ideas');
		expect(r('qué clientes dejaron de comprar', x)?.tema).toBe('clientes');
		expect(r('por donde vendo mas', x)?.tema).toBe('canales');
		expect(r('cuántas visitas tuvo la tienda', x)?.tema).toBe('visitas');
		expect(r('cual fue mi mejor dia', x)?.tema).toBe('dias');
		expect(r('lo que no se vende', x)?.tema).toBe('quietos');
		expect(r('mi producto mas caro', x)?.tema).toBe('caro');
		expect(r('cuantos productos tengo', x)?.tema).toBe('catalogo');
		expect(r('cual fue el ultimo pedido', x)?.tema).toBe('ultimos');
		expect(r('voy a llegar a la meta?', x)?.tema).toBe('meta');
	});

	it('frases sueltas de todos los días', () => {
		const x = datos();
		expect(r('cómo me fue el martes', x)?.tema).toBe('ventas');
		expect(r('comparame esta semana con la pasada', x)?.texto).toContain('7 días anteriores');
		expect(r('cual es mi producto estrella', x)?.tema).toBe('top');
		expect(r('quienes no volvieron', x)?.tema).toBe('clientes');
		expect(r('cuanta gente entró a la tienda', x)?.tema).toBe('visitas');
		expect(r('como viene la meta del mes', x)?.tema).toBe('meta');
		expect(r('cuanto vale todo mi stock', x)?.tema).toBe('catalogo');
		expect(r('necesito subir los precios un 10%', x)?.tema).toBe('guia');
		expect(r('como cambio los colores de la tienda', x)?.acciones?.[0].ir).toBe('mi-tienda');
		expect(r('precio del mate imperial', x)?.texto).toContain('No encontré «imperial»');
		expect(r('y de bombillas?', x, ['cuanto stock hay de termos'])?.texto).not.toContain(
			'No encontré'
		);
	});

	it('si el nombre coincide con varios productos, los muestra a todos', () => {
		const x = datos();
		const yerbas = x.contexto.productos.filter((p) => /yerba/i.test(p.nombre));
		const y = r('me quedan yerbas?', x)!;
		if (yerbas.length > 1) expect(y.bloques?.[0].tipo).toBe('tabla');
		for (const p of yerbas) expect(y.texto).toContain(p.nombre);
	});

	it('los cómo hago llevan con un botón', () => {
		const x = datos();
		expect(r('¿cómo creo un cupón?', x)?.acciones?.[0].ir).toBe('cupones');
		expect(r('quiero traer mis productos de tiendanube', x)?.acciones?.[0].ir).toBe('importar');
		expect(r('donde cambio el logo', x)?.acciones?.[0].ir).toBe('mi-tienda');
		expect(r('como conecto mercado pago', x)?.acciones?.[0].ir).toBe('integraciones');
		expect(r('como vendo en el local', x)?.acciones?.[0].ir).toBe('venta');
	});
});

describe('Tincho sin IA: muestra', () => {
	it('cifras, gráficos y tablas con los datos', () => {
		const x = datos();
		const semana = r('cuanto vendi esta semana', x)!;
		expect(semana.bloques?.map((b) => b.tipo)).toEqual(['cifras', 'linea']);
		const linea = semana.bloques![1] as Extract<
			NonNullable<typeof semana.bloques>[number],
			{ tipo: 'linea' }
		>;
		expect(linea.puntos).toHaveLength(7);
		const top = r('que es lo que mas se vende', x)!;
		expect(top.bloques?.map((b) => b.tipo)).toEqual(['barras', 'tabla']);
		expect(top.sugerencias?.length).toBeGreaterThan(0);
	});

	it('sin montos: nada en pesos, ni en los bloques', () => {
		const x = datos();
		const sin: DatosHornero = {
			...x,
			hoy: sinMontos(x.hoy),
			semana: sinMontos(x.semana),
			mes: sinMontos(x.mes!),
			contexto: contextoSinMontos(x.contexto)
		};
		for (const q of [
			'hola',
			'cuanto vendi esta semana',
			'ventas del mes',
			'que es lo que mas se vende',
			'por donde vendo mas',
			'cuantos productos tengo',
			'ideas para vender mas',
			'cual fue el ultimo pedido'
		]) {
			const y = r(q, sin);
			expect(y, q).toBeTruthy();
			expect(JSON.stringify(y), q).not.toContain('$');
		}
	});

	it('no inventa: lo que no entiende queda para la IA', () => {
		expect(r('escribime un poema sobre el otoño')).toBeNull();
		expect(r('cuál es la capital de Francia')).toBeNull();
	});
});
