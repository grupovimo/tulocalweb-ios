import { existsSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { describe, expect, it } from 'vitest';
import { CATALOGOS } from '#lib/catalogos.ts';
import { PLANTILLAS } from '#lib/plantillas.ts';
import { demoDe } from './demo.ts';
import { resumenDemo } from './resumen.ts';
import { inicioDelDia, tableroDemo } from './tablero.ts';

const BASE = Date.parse('2026-10-20T15:00:00Z');

describe('datos de la demo', () => {
	it('todas las plantillas tienen catálogo con categorías y fotos', () => {
		expect(new Set(CATALOGOS.map((c) => c.plantilla))).toEqual(
			new Set(PLANTILLAS.map((p) => p.id))
		);
		for (const c of CATALOGOS) {
			expect(c.productos.length).toBeGreaterThanOrEqual(8);
			for (const p of c.productos) {
				expect(c.categorias).toContain(p.categoria);
				// Un id de Unsplash o una foto propia de static/demo (que tiene que estar en el repo).
				const propia = p.foto.match(
					/^https:\/\/tulocalweb\.com\.ar\/(demo\/[a-z0-9-]+\/[a-z0-9-]+\.webp)$/
				);
				if (propia)
					expect(
						existsSync(fileURLToPath(new URL(`../../../../static/${propia[1]}`, import.meta.url))),
						p.foto
					).toBe(true);
				else expect(p.foto).toMatch(/^\d{10,}-[0-9a-f]+$/);
				expect(p.precio).toBeGreaterThan(0);
			}
			expect(new Set(c.productos.map((p) => p.id)).size).toBe(c.productos.length);
		}
	});

	it('es siempre la misma para una plantilla y se fecha desde la base', () => {
		const a = demoDe('litoral', BASE)!;
		const b = demoDe('litoral', BASE + 1)!;
		expect(a.pedidos.map((p) => p.total)).toEqual(b.pedidos.map((p) => p.total));
		expect(b.pedidos[0].creado - a.pedidos[0].creado).toBe(1);
		expect(a.pedidos[0].numero).toBe(1048);
		expect(Math.max(...a.pedidos.map((p) => p.creado))).toBeLessThan(BASE);
		expect(Math.min(...a.pedidos.map((p) => p.creado))).toBeGreaterThan(BASE - 61 * 86_400_000);
	});

	it('los pedidos usan productos del catálogo y los totales cierran', () => {
		for (const c of CATALOGOS) {
			const d = demoDe(c.plantilla, BASE)!;
			const ids = new Set(c.productos.map((p) => p.id));
			const clientes = new Set(d.clientes.map((x) => x.id));
			// Un solo expect por tienda: miles de expect sueltos hacen lento el test.
			const malos = d.pedidos.filter(
				(p) =>
					!p.items.length ||
					p.items.some((i) => !ids.has(i.productoId)) ||
					p.total !== p.items.reduce((s, i) => s + i.precio * i.cantidad, 0) ||
					!clientes.has(p.clienteId)
			);
			expect(
				malos.map((p) => p.numero),
				c.plantilla
			).toEqual([]);
		}
	});

	it('el resumen cuenta los pedidos para preparar y respeta las leídas', () => {
		const d = demoDe('andes', BASE)!;
		const r = resumenDemo(d, BASE, 0, BASE);
		expect(r.insignias.pedidos).toBe(d.pedidos.filter((p) => p.estado === 'Para preparar').length);
		expect(r.notificaciones.filter((n) => !n.leida)).toHaveLength(6);
		expect(r.notificaciones[0].detalle).toContain('Tincho ya respondió');
		const leido = resumenDemo(d, BASE, BASE, BASE);
		expect(leido.notificaciones.every((n) => n.leida)).toBe(true);
	});

	it('los clientes nuevos aparecen con su primera compra y hay caras en los últimos', () => {
		const d = demoDe('pampa', BASE)!;
		const primera = new Map<string, number>();
		for (const p of d.pedidos) primera.set(p.clienteId, p.creado);
		for (const c of d.clientes.filter((c) => !c.mayorista && c.alta > BASE - 60 * 86_400_000))
			expect(primera.get(c.id)).toBe(c.alta);
		expect(d.clientes.filter((c) => c.foto)).toHaveLength(8);
		expect(d.clientes.find((c) => c.id === d.pedidos[0].clienteId)!.foto).not.toBeNull();
	});

	it('el local sólo vende en horario', () => {
		const d = demoDe('miga', BASE)!;
		for (const p of d.pedidos.filter((p) => p.canal === 'pos')) {
			const h = (p.creado - inicioDelDia(p.creado)) / 3_600_000;
			expect(h >= 9 && h < 21).toBe(true);
		}
	});
});

describe('tablero de la demo', () => {
	const d = demoDe('litoral', BASE)!;
	const AHORA = BASE + 30 * 60_000;
	const tab = (periodo: 'hoy' | '7d' | '30d', cambios = {}) =>
		tableroDemo(d, periodo, { base: BASE, ahora: AHORA, hola: 'Valeria', cambios });

	it('las ventas, la serie y los pedidos cierran con los pedidos de la ventana', () => {
		const t = tab('7d');
		const ventana = d.pedidos.filter(
			(p) => p.estado !== 'Cancelado' && p.creado >= AHORA - 7 * 86_400_000 && p.creado < AHORA
		);
		expect(t.kpis.pedidos.actual).toBe(ventana.length);
		expect(t.ventas.actual).toBe(ventana.reduce((s, p) => s + p.total, 0));
		expect(t.serie.actual).toHaveLength(14);
		expect(t.serie.actual.at(-1)).toBe(t.ventas.actual);
		expect(t.serie.actual).toEqual([...t.serie.actual].sort((a, b) => a - b));
		expect(t.canales.reduce((s, c) => s + c.parte, 0)).toBeCloseTo(1);
		expect(t.ultimosPedidos[0].numero).toBe(1048);
	});

	it('hoy compara contra ayer y 30 días contra el mes anterior', () => {
		const hoy = tab('hoy');
		expect(Date.parse(hoy.desde)).toBe(inicioDelDia(AHORA));
		const mes = tab('30d');
		expect(mes.ventas.anterior).toBeGreaterThan(0);
		expect(mes.kpis.clientesNuevos.actual).toBeGreaterThan(0);
	});

	it('respeta la meta, el aviso oculto y los carritos recordados', () => {
		const t = tab('7d', { meta: 25_000_000, certificadoOculto: true, ideasHechas: ['carritos'] });
		expect(t.meta.objetivo).toBe(25_000_000);
		expect(t.avisos.certificadoDias).toBeNull();
		expect(t.hornero.recordados).toBe(true);
		expect(tab('7d').avisos.certificadoDias).toBe(5);
		expect(tab('7d').meta.objetivo % 100_000).toBe(0);
	});

	it('liquida un medio por fila, en los próximos 7 días', () => {
		const t = tab('7d');
		expect(new Set(t.liquidaciones.map((l) => l.medio)).size).toBe(t.liquidaciones.length);
		const hoy = new Date(inicioDelDia(AHORA) - 3 * 3_600_000).toISOString().slice(0, 10);
		for (const l of t.liquidaciones) expect(l.fecha >= hoy).toBe(true);
	});

	it('para hacer hoy y la insignia del menú cuentan lo mismo', () => {
		const t = tab('7d');
		const r = resumenDemo(d, BASE, 0, AHORA);
		const conAlgo = [
			t.paraHacer.preparar,
			t.paraHacer.chats,
			t.paraHacer.transferencias,
			t.paraHacer.sinStock
		];
		expect(r.insignias.paraHacer).toBe(conAlgo.filter((n) => n > 0).length);
		expect(t.paraHacer.preparar).toBe(r.insignias.pedidos);
	});
});
