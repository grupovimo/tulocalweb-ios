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

const filtros = { q: '', seg: [], facetas: {}, banderas: [], orden: null, pagina: 1 };

describe('«Avisame cuando vuelva» en la demo (§60, como avisostock en Go)', () => {
	it('anota sin stock, rechaza con stock, cuenta los que esperan y los da por avisados al reponer', async () => {
		const datos = await demo();
		const t = await datos.tiendaPublica();
		const p = t.productos.find((x) => !x.opciones.length && x.stock > 0)!;
		const pedido = { productoId: p.id, varianteId: null, email: 'ana@ejemplo.com', whatsapp: '' };
		expect(await datos.avisarStock(pedido)).toBe('hay-stock');
		expect(await datos.avisarStock({ ...pedido, productoId: 'no-existe' })).toBe('no-existe');
		const esperan = async () => {
			const l = await datos.listar('productos', { ...filtros, q: p.nombre } as never);
			return (l.filas as { id: string; esperan?: number }[]).find((f) => f.id === p.id)?.esperan;
		};

		// Se agota desde Inventario: ahora sí se anota (repetido no suma). Los agotados traen unos
		// pocos de ejemplo.
		await datos.accionLista('inventario', [p.id], 'ajustar', -p.stock);
		const antes = (await esperan()) ?? 0;
		expect(antes).toBeGreaterThan(0);
		expect(await datos.avisarStock(pedido)).toBe('ok');
		expect(await datos.avisarStock({ ...pedido, email: 'ANA@ejemplo.com' })).toBe('ok');
		expect(await datos.avisarStock({ ...pedido, email: 'beto@ejemplo.com' })).toBe('ok');
		expect(await esperan()).toBe(antes + 2);

		// Vuelve el stock: la demo no manda mails, pero dejan de esperar.
		await datos.accionLista('inventario', [p.id], 'ingreso', 3);
		expect(await esperan()).toBe(0);
	});
});
