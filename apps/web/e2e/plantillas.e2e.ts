import { expect, test, type Page } from '@playwright/test';
import { PLANTILLAS } from '../src/lib/plantillas.ts';

const slugDe = (n: string) =>
	n
		.normalize('NFD')
		.replace(/[̀-ͯ]/g, '')
		.toLowerCase()
		.replace(/&/g, 'y')
		.replace(/[^a-z0-9]+/g, '-')
		.replace(/^-|-$/g, '');

const desborde = (page: Page) =>
	page.evaluate(() => document.documentElement.scrollWidth - innerWidth);

/**
 * Cada punto de partida (las 17 vitrinas) por todas las páginas de la tienda en bloques: abren,
 * no desbordan a lo ancho y no tiran errores. El carrito y el checkout, con dos productos adentro
 * (la vitrina no compra, pero se dibujan igual).
 */
for (const p of PLANTILLAS) {
	const slug = slugDe(p.tienda);
	test(`vitrina ${p.id} por todas sus páginas`, async ({ page }) => {
		const errores: string[] = [];
		page.on('pageerror', (e) => errores.push(e.message));
		await page.addInitScript((s) => {
			localStorage.setItem('tlw-cookies', 'nec');
			localStorage.setItem(`mayor18:${s}`, '1');
		}, slug);
		const r = await page.goto(`/t/${slug}`);
		expect(r?.status()).toBe(200);
		expect(await desborde(page), `${p.id} inicio`).toBeLessThanOrEqual(0);

		await page.goto(`/t/${slug}/catalogo`);
		const fichas = await page
			.locator(`a[href^="/t/${slug}/producto/"]`)
			.evaluateAll((as) => [...new Set(as.map((a) => a.getAttribute('href')!))]);
		expect(fichas.length, `${p.id}: productos en el catálogo`).toBeGreaterThan(0);
		await page.evaluate(
			({ s, ids }) =>
				localStorage.setItem(
					`carrito:${s}`,
					JSON.stringify(ids.map((id) => ({ id, varianteId: null, cantidad: 1 })))
				),
			{ s: slug, ids: fichas.slice(0, 2).map((h) => h.split('/').pop()!) }
		);

		const rutas = [
			'/catalogo?orden=menor&ofertas=1',
			'/catalogo?q=zzzz',
			fichas[0].slice(`/t/${slug}`.length),
			'/nosotros',
			'/carrito',
			'/checkout',
			'/arrepentimiento',
			'/no-existe-esta-pagina'
		];
		for (const ruta of rutas) {
			const res = await page.goto(`/t/${slug}${ruta}`);
			expect(res?.status(), `${p.id}${ruta}`).toBe(ruta.startsWith('/no-existe') ? 404 : 200);
			expect(await desborde(page), `${p.id}${ruta}`).toBeLessThanOrEqual(0);
		}
		expect(errores).toEqual([]);
	});
}

/**
 * Turnos en cada dirección: en una demo se prenden los turnos con un servicio y alguien que
 * atiende, y la reserva y su link en el menú se dibujan sin errores ni desborde.
 */
for (const id of ['ceibo', 'malbec', 'litoral', 'andes', 'mostaza', 'miga']) {
	const p = PLANTILLAS.find((x) => x.id === id)!;
	test(`turnos en ${id}`, async ({ page }) => {
		const errores: string[] = [];
		page.on('pageerror', (e) => errores.push(e.message));
		await page.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
		await page.goto('/prueba');
		await page.locator('html[data-listo]').waitFor({ state: 'attached' });
		// El nombre del botón es su texto visible ("Miga Gastronomía · …"); "Miga" no es "Miga Clásica".
		await page
			.getByRole('button', { name: new RegExp(`^${p.nombre}(?! Clásica)`) })
			.first()
			.click();
		await expect(page).toHaveURL(/\/panel$/);
		const s = await page.request.post('/panel/turnos/servicios', {
			data: {
				nombre: 'Asesoría',
				descripcion: '',
				precio: 9000,
				duracion: 30,
				foto: null,
				etiqueta: null,
				activo: true
			}
		});
		expect(s.status()).toBe(201);
		const pr = await page.request.post('/panel/turnos/profesionales', {
			data: { nombre: 'Marta', rol: 'Atiende', foto: null, servicios: [], activo: true }
		});
		expect(pr.status()).toBe(201);
		const a = await page.request.put('/panel/turnos/ajustes', {
			data: {
				activo: true,
				sena: 30,
				anticipacionMin: 60,
				diasVista: 14,
				horario: [1, 2, 3, 4, 5, 6].map((dia) => ({ dia, desde: '09:00', hasta: '20:00' }))
			}
		});
		expect(a.ok()).toBe(true);
		const slug = slugDe(p.tienda);
		await page.goto(`/t/${slug}`);
		await page.locator('html[data-listo]').waitFor({ state: 'attached' });
		// Malbec pregunta la edad antes de mostrar la tienda (sale al hidratar: se espera).
		if (id === 'malbec') await page.getByRole('button', { name: 'Sí, soy mayor' }).click();
		await expect(page.locator(`a[href="/t/${slug}/turnos"]`).first()).toBeAttached();
		await page.goto(`/t/${slug}/turnos`);
		await page.locator('html[data-listo]').waitFor({ state: 'attached' });
		await page.getByRole('radiogroup', { name: 'Servicio' }).getByRole('radio').first().click();
		// De noche ya no quedan horarios hoy: se pasa al día siguiente abierto.
		const horarios = page.locator('[role=radiogroup][aria-label^="Horarios"]').first();
		const sinHorarios = page.getByText('No quedan horarios este día');
		await expect(horarios.or(sinHorarios)).toBeVisible();
		if (await sinHorarios.isVisible())
			await page
				.getByRole('radiogroup', { name: 'Día' })
				.locator('[role=radio]:not([disabled])')
				.nth(1)
				.click();
		await expect(horarios).toBeVisible();
		expect(await desborde(page), `${id}/turnos`).toBeLessThanOrEqual(0);
		expect(errores).toEqual([]);
	});
}
