import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });
const pesos = (s: string) => Number(s.replace(/\D/g, ''));

test('tienda: un cupón se valida, descuenta y suma un uso en el panel', async ({ page }) => {
	await page.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
	await page.goto('/prueba');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	const slug = (await page.locator('.t-d').first().textContent())!.split('/t/')[1].trim();

	await page.goto(`/t/${slug}/catalogo`);
	await listo(page);
	await page.goto(`/t/${slug}/producto/bombilla-alpaca`);
	await listo(page);
	await listo(page);
	await page.getByRole('button', { name: 'Agregar al carrito' }).click();
	await expect(page).toHaveURL(/\/carrito$/);
	await page.getByRole('link', { name: 'Finalizar compra' }).click();
	await listo(page);

	await page.getByLabel('Email').fill(`cupon${Date.now()}@ejemplo.com`);
	await page.getByLabel('Nombre y apellido').fill('Lucía Pérez');
	await page.getByLabel('Teléfono').fill('11 5555-0101');
	await page.getByLabel('Código postal').fill('1414');
	await page.getByLabel(/Dirección/).fill('Gorriti 4500');
	await page.getByRole('radio', { name: /Retiro en el local/ }).click();
	await page.getByRole('radio', { name: /Mercado Pago/ }).click();
	const total = page.locator('b', { hasText: /^\$/ }).last();
	const antes = pesos((await total.textContent())!);

	// Uno que no existe: lo dice y no descuenta.
	await page.getByLabel('¿Tenés un cupón?').fill('NOEXISTE');
	await page.getByRole('button', { name: 'Aplicar' }).click();
	await expect(page.getByRole('alert')).toContainText('no existe');

	// FINDE10 (10 %, vale en la web): descuenta y queda aplicado.
	await page.getByLabel('¿Tenés un cupón?').fill('finde10');
	await page.getByRole('button', { name: 'Aplicar' }).click();
	await expect(page.getByRole('status').filter({ hasText: 'FINDE10' })).toBeVisible();
	await expect(page.getByText(/^Cupón FINDE10$/)).toBeVisible();
	await expect.poll(async () => pesos((await total.textContent())!)).toBeLessThan(antes);
	expect(
		await page.evaluate(() => document.documentElement.scrollWidth - innerWidth)
	).toBeLessThanOrEqual(0);

	await page.getByRole('button', { name: /^Pagar/ }).click();
	await expect(page).toHaveURL(/\/pedido\//);

	// En el panel, FINDE10 sumó un uso (tenía 41).
	await page.goto('/panel/cupones');
	await listo(page);
	await expect(page.getByText('42 de 100 usos')).toBeVisible();
});

test('tienda: el link de un cupón lo deja aplicado en el checkout', async ({ page }) => {
	await page.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
	await page.goto('/prueba');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	const slug = (await page.locator('.t-d').first().textContent())!.split('/t/')[1].trim();

	await page.goto(`/t/${slug}/c/finde10`);
	await expect(page).toHaveURL(new RegExp(`/t/${slug}/catalogo$`));
	await listo(page);
	await page.goto(`/t/${slug}/producto/bombilla-alpaca`);
	await listo(page);
	await listo(page);
	await page.getByRole('button', { name: 'Agregar al carrito' }).click();
	await expect(page).toHaveURL(/\/carrito$/);
	// El carrito ya lo trae aplicado (lo dejó el link) y el checkout también.
	await expect(page.getByRole('status').filter({ hasText: 'FINDE10' })).toBeVisible();
	await page.getByRole('link', { name: 'Finalizar compra' }).click();
	await expect(page).toHaveURL(/\/checkout$/);
	await listo(page);
	await expect(page.getByRole('status').filter({ hasText: 'FINDE10' })).toBeVisible();

	// Quitarlo no lo vuelve a traer.
	await page.getByRole('button', { name: 'Quitar', exact: true }).click();
	await page.reload();
	await listo(page);
	await expect(page.getByLabel('¿Tenés un cupón?')).toHaveValue('');
});
