import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });
const sinDesborde = (page: Page) =>
	page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);

/** Entra a una demo, guarda y publica ese diseño, y devuelve el slug de la tienda. */
async function conDiseno(page: Page, demo: RegExp, diseno: object) {
	await page.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
	await page.goto('/prueba');
	await listo(page);
	await page.getByRole('button', { name: demo }).click();
	await expect(page).toHaveURL(/\/panel$/);
	const slug = (await page.locator('.t-d').first().textContent())!.split('/t/')[1].trim();
	expect((await page.request.put('/panel/mi-tienda', { data: diseno })).status()).toBe(204);
	expect(
		await page.evaluate(
			async () => (await fetch('/panel/mi-tienda/publicar', { method: 'POST' })).status
		)
	).toBe(204);
	return slug;
}

test('ficha: «Completá el look» con el combo del producto agrega todo junto', async ({ page }) => {
	const slug = await conDiseno(page, /^Litoral/, {
		version: 2,
		paleta: 0,
		inicio: [
			{ id: 'portada', tipo: 'portada', variante: 'foto', datos: {} },
			{
				id: 'combomate',
				tipo: 'combo',
				variante: 'tarjeta',
				datos: { productos: ['miel-de-monte', 'bombilla-alpaca'], descuento: 10 }
			}
		]
	});

	await page.goto(`/t/${slug}/producto/miel-de-monte`);
	await listo(page);
	const look = page.locator('.tv-ficha-look');
	await expect(look.getByRole('heading', { name: 'Llevalos juntos' })).toBeVisible();
	// Muestra los otros del combo (no el que se está viendo), el total y el ahorro de los dos.
	await expect(look.getByRole('link', { name: /Bombilla/ })).toBeVisible();
	await expect(look.getByRole('link', { name: /Miel/ })).toHaveCount(0);
	await expect(look.getByText('Con este, los 2 juntos')).toBeVisible();
	await expect(look.getByText(/Ahorrás \$\s2\.040/)).toBeVisible();
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	await look.getByRole('button', { name: 'Agregar los 2' }).click();

	await page.goto(`/t/${slug}/carrito`);
	await listo(page);
	await expect(page.getByText('En combo')).toHaveCount(2);
	await expect(page.locator('.tv-carrito-resumen .tv-cuentas-total')).toContainText('18.360');
});

test('ficha: la guía de talles se abre en un panel junto al talle', async ({ page }) => {
	const slug = await conDiseno(page, /^Ceibo/, {
		version: 2,
		paleta: 0,
		inicio: [{ id: 'portada', tipo: 'portada', variante: 'foto', datos: {} }],
		paginas: {
			producto: {
				guiaTalles: {
					tabla: [
						['Talle', 'Busto (cm)', 'Largo (cm)'],
						['S', '88', '96'],
						['M', '94', '98'],
						['L', '100', '100']
					]
				}
			}
		}
	});

	await page.goto(`/t/${slug}/producto/tapado-merino-ceibo`);
	await listo(page);
	const abrir = page.getByRole('button', { name: 'Guía de talles' });
	await expect(abrir).toBeVisible();
	await abrir.click();
	const panel = page.getByRole('dialog', { name: 'Guía de talles' });
	await expect(panel).toBeVisible();
	await expect(panel.getByRole('columnheader', { name: 'Busto (cm)' })).toBeVisible();
	await expect(panel.getByRole('rowheader', { name: 'M' })).toBeVisible();
	await expect(panel.getByRole('cell', { name: '94' })).toBeVisible();
	await page.keyboard.press('Escape');
	await expect(panel).toBeHidden();
	await expect(abrir).toBeFocused();
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
});
