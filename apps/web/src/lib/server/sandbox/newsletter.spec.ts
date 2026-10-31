import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import { afterEach, describe, expect, it } from 'vitest';
import { filasCsv, validarNovedad } from '#lib/panel/newsletter.ts';
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

describe('newsletter de la demo (§50)', () => {
	it('trae suscriptos de ejemplo, una novedad en el historial y los productos publicados', async () => {
		const n = await (await moda()).newsletter();
		expect(n.suscriptores.length).toBeGreaterThan(5);
		expect(n.pendientes).toBe(2);
		expect(n.confirmados).toBe(n.suscriptores.length - 2);
		expect(n.envios).toHaveLength(1);
		expect(n.tope).toBe(500);
		expect(n.productos.length).toBeGreaterThan(0);
	});

	it('alta pendiente, confirmar con el token, baja y borrar', async () => {
		const datos = await moda();
		const antes = await datos.newsletter();
		expect(await datos.suscribirNewsletter('no-es-email', 'tienda')).toBeNull();
		const s = (await datos.suscribirNewsletter(' Lu@Ejemplo.com ', 'tienda'))!;
		expect(s.confirmado).toBe(false);
		// Pedirlo de nuevo devuelve el mismo token.
		expect((await datos.suscribirNewsletter('lu@ejemplo.com', 'checkout'))!.token).toBe(s.token);
		let n = await datos.newsletter();
		expect(n.pendientes).toBe(antes.pendientes + 1);
		expect(n.suscriptores[0]).toMatchObject({ email: 'lu@ejemplo.com', confirmado: false });

		expect(await datos.confirmarNewsletter('x'.repeat(24))).toBe(false);
		expect(await datos.confirmarNewsletter(s.token)).toBe(true);
		n = await datos.newsletter();
		expect(n.confirmados).toBe(antes.confirmados + 1);
		expect((await datos.suscribirNewsletter('lu@ejemplo.com', 'tienda'))!.confirmado).toBe(true);

		expect(await datos.bajaNewsletter(s.token)).toBe(true);
		expect(await datos.bajaNewsletter(s.token)).toBe(false);
		const ejemplo = n.suscriptores.find((x) => x.id.startsWith('ej-'))!;
		expect(await datos.borrarSuscriptor(ejemplo.id)).toBe('ok');
		expect(await datos.borrarSuscriptor(ejemplo.id)).toBe('no-existe');
		n = await datos.newsletter();
		expect(n.suscriptores.length).toBe(antes.suscriptores.length - 1);
	});

	it('mandar valida, queda en el historial y no manda mails', async () => {
		const datos = await moda();
		const n = await datos.newsletter();
		expect(await datos.mandarNovedad({ asunto: 'Hola', texto: 'corto', productos: [] })).toEqual({
			error: 'Escribí la novedad (al menos 10 caracteres).'
		});
		expect(
			await datos.mandarNovedad({
				asunto: 'Hola',
				texto: 'Una novedad larga',
				productos: ['no-existe']
			})
		).toMatchObject({ error: expect.stringContaining('publicado') });
		const e = await datos.mandarNovedad({
			asunto: 'Llegó  la  temporada',
			texto: 'Ya están los productos nuevos.',
			productos: [n.productos[0].id]
		});
		expect(e).toMatchObject({
			asunto: 'Llegó la temporada',
			destinatarios: n.confirmados,
			estado: 'enviado'
		});
		const despues = await datos.newsletter();
		expect(despues.envios[0].asunto).toBe('Llegó la temporada');
		expect(despues.hoy).toBe(n.confirmados);
	});
});

describe('newsletter: validación y CSV', () => {
	it('valida como Go y arma el CSV sin fórmulas', () => {
		expect(validarNovedad({ asunto: 'ab', texto: 'x'.repeat(20), productos: [] })).toEqual({
			ok: false,
			error: 'Poné un asunto.'
		});
		expect(
			validarNovedad({
				asunto: 'Asunto',
				texto: 'x'.repeat(20),
				productos: ['a', 'b', 'c', 'd', 'e']
			})
		).toEqual({ ok: false, error: 'Elegí hasta 4 productos.' });
		expect(
			validarNovedad({ asunto: ' Asunto ', texto: 'x'.repeat(20), productos: ['a', 'a'] })
		).toMatchObject({ ok: true, novedad: { asunto: 'Asunto', productos: ['a'] } });
		const filas = filasCsv([
			{ email: '=1+1@x.com', confirmado: true, creado: '2026-10-31T02:00:00Z', origen: 'tienda' }
		]);
		expect(filas[0]).toEqual(['Email', 'Estado', 'Desde', 'Origen']);
		expect(filas[1]).toEqual(["'=1+1@x.com", 'Confirmado', '2026-10-30', 'Tienda online']);
	});
});
