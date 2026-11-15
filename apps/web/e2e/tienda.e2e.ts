import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });
const sinDesborde = (page: Page) =>
	page.evaluate(() => document.documentElement.scrollWidth - innerWidth);

async function demo(page: Page) {
	await page.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
	await page.goto('/prueba');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	return (await page.locator('.t-d').first().textContent())!.split('/t/')[1].trim();
}

test('tienda: se compra de punta a punta y el pedido entra al panel', async ({ page }) => {
	const slug = await demo(page);
	await page.goto(`/t/${slug}`);
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	await page.goto(`/t/${slug}/catalogo`);
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	// Uno sin opciones (los que tienen se compran eligiendo: variantes.e2e.ts).
	await page.goto(`/t/${slug}/producto/bombilla-alpaca`);
	await listo(page);
	await page.getByRole('button', { name: 'Agregar al carrito' }).click();
	await expect(page).toHaveURL(/\/carrito$/);
	await expect(page.getByRole('heading', { name: 'Tu carrito' })).toBeVisible();
	await expect(page.getByText('1 producto', { exact: true })).toBeVisible();
	await page.getByRole('link', { name: 'Finalizar compra' }).click();

	await page.getByLabel('Email').fill('lucia@ejemplo.com');
	await page.getByLabel('Nombre y apellido').fill('Lucía Pérez');
	await page.getByLabel('Teléfono').fill('11 5555-0101');
	await page.getByLabel('Código postal').fill('1414');
	await page.getByLabel(/Dirección/).fill('Gorriti 4500');
	await page.getByRole('radio', { name: /Retiro en el local/ }).click();
	await page.getByRole('radio', { name: /Transferencia/ }).click();
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	await page.getByRole('button', { name: /^Pagar/ }).click();
	await expect(page.getByRole('heading', { name: '¡Gracias, Lucía!' })).toBeVisible();
	await expect(page.getByText(/al alias/)).toBeVisible();
	const numero = (await page
		.locator('b', { hasText: /^#\d+$/ })
		.first()
		.textContent())!.slice(1);

	// El carrito quedó vacío y el pedido está en el panel, sin pagar (espera la transferencia).
	await page.goto(`/t/${slug}/carrito`);
	await expect(page.getByText('Todavía no sumaste nada.')).toBeVisible();
	await page.goto(`/panel/pedidos?q=%23${numero}`);
	await listo(page);
	await expect(page.locator('.fila').first()).toContainText('Lucía Pérez');
});

test('tienda: la vitrina de una plantilla se mira pero no se compra', async ({ page }) => {
	await page.goto('/t/tienda-lumbre');
	await expect(page.getByText(/tienda de ejemplo/)).toBeVisible();
	await page.locator('a[href*="/producto/"]').first().click();
	// Con opciones, el botón pide elegir; igual no se puede comprar.
	await expect(
		page.getByRole('button', { name: /^(Agregar al carrito|Elegí)/ }).first()
	).toBeDisabled();
	expect((await page.goto('/t/no-existe-esta'))?.status()).toBe(404);
});
