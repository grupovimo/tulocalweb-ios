import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });

test('botón de arrepentimiento de la tienda: código de trámite y devolución en el panel', async ({
	page
}) => {
	await page.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
	await page.goto('/prueba');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	const slug = (await page.locator('.t-d').first().textContent())!.split('/t/')[1].trim();

	// Una compra en la tienda.
	// Un producto sin presentaciones, para que "Agregar al carrito" esté a mano (como en carritos).
	await page.goto(`/t/${slug}/producto/bombilla-alpaca`);
	await listo(page);
	await page.getByRole('button', { name: 'Agregar al carrito' }).click();
	await page.getByRole('link', { name: 'Finalizar compra' }).click();
	await listo(page);
	await page.getByLabel('Email').fill('lucia@ejemplo.com');
	await page.getByLabel('Nombre y apellido').fill('Lucía Pérez');
	await page.getByLabel('Teléfono').fill('11 5555-0101');
	await page.getByLabel('Código postal').fill('1414');
	await page.getByLabel(/Dirección/).fill('Gorriti 4500');
	await page.getByRole('radio', { name: /Retiro en el local/ }).click();
	await page.getByRole('radio', { name: /Transferencia/ }).click();
	await page.getByRole('button', { name: /^Pagar/ }).click();
	await expect(page).toHaveURL(/\/pedido\//);
	const codigo = page.url().split('/pedido/')[1].split('?')[0];

	// El link del footer lleva al botón de la tienda, no al de Tu Local Web.
	await expect(page.locator(`a[href="/t/${slug}/arrepentimiento"]`).first()).toBeAttached();
	await page.goto(`/t/${slug}/arrepentimiento`);
	await listo(page);
	expect(
		await page.evaluate(() => document.documentElement.scrollWidth - innerWidth)
	).toBeLessThanOrEqual(0);
	await page.getByLabel('Código o número del pedido').fill('99999-x');
	await page.getByLabel('Email de la compra').fill('lucia@ejemplo.com');
	await page.getByLabel('Nombre y apellido').fill('Lucía Pérez');
	await page.getByRole('button', { name: 'Enviar pedido de arrepentimiento' }).click();
	await expect(page.getByRole('alert')).toContainText('No encontramos ese pedido');
	await page.getByLabel('Código o número del pedido').fill(codigo);
	await page.getByRole('button', { name: 'Enviar pedido de arrepentimiento' }).click();
	const tramite = await page.locator('.tv-confirma-alias').textContent();
	expect(tramite).toMatch(/^ARR-\d{4}-\d+$/);

	await page.goto('/panel/devoluciones');
	await listo(page);
	await expect(page.getByText('Lucía Pérez').first()).toBeVisible();
});
