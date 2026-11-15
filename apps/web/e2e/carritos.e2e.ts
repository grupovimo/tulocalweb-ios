import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });

async function demo(page: Page) {
	await page.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
	await page.goto('/prueba');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	return (await page.locator('.t-d').first().textContent())!.split('/t/')[1].trim();
}

test('carritos: el que deja su email en el checkout aparece en el panel', async ({ page }) => {
	const slug = await demo(page);
	await page.goto(`/t/${slug}/producto/bombilla-alpaca`);
	await listo(page);
	await page.getByRole('button', { name: 'Agregar al carrito' }).click();
	await expect(page).toHaveURL(/\/carrito$/);
	await page.getByRole('link', { name: 'Finalizar compra' }).click();
	await listo(page);

	// Al dejar el email y el nombre, el carrito se guarda (sin tocar "Pagar").
	await page.getByLabel('Nombre y apellido').fill('Marta Abandono');
	const guardado = page.waitForResponse((r) => r.url().endsWith('/carrito/guardar'));
	await page.getByLabel('Email').fill('marta@ejemplo.com');
	await page.getByLabel('Teléfono').focus();
	expect((await guardado).status()).toBe(204);

	await page.goto('/panel/carritos-abandonados');
	await listo(page);
	await expect(page.getByText('Marta Abandono').first()).toBeVisible();

	// Un link de recordatorio que ya no sirve deja el carrito como estaba.
	await page.goto(`/t/${slug}/carrito/recuperar/${'x'.repeat(32)}`);
	await expect(page).toHaveURL(/\/carrito$/);
	await expect(page.getByRole('heading', { name: 'Tu carrito' })).toBeVisible();
	await expect(page.getByText('1 producto', { exact: true })).toBeVisible();
});
