import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });

/**
 * El combo con descuento de punta a punta: la tienda lo arma en Mi tienda, quien compra agrega los
 * dos juntos, el carrito marca las líneas y descuenta, y el checkout cobra lo mismo (lo calcula el
 * servidor con la regla de casos/combos.json).
 */
test('combo: se agrega junto, el carrito lo descuenta y el pedido lo cobra', async ({ page }) => {
	await page.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
	await page.goto('/prueba');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	const slug = (await page.locator('.t-d').first().textContent())!.split('/t/')[1].trim();

	const diseno = {
		version: 2,
		paleta: 0,
		inicio: [
			{ id: 'portada', tipo: 'portada', variante: 'foto', datos: {} },
			{
				id: 'combomate',
				tipo: 'combo',
				variante: 'tarjeta',
				datos: {
					titulo: 'El mate completo',
					productos: ['miel-de-monte', 'bombilla-alpaca'],
					descuento: 10
				}
			}
		]
	};
	expect((await page.request.put('/panel/mi-tienda', { data: diseno })).status()).toBe(204);
	// Publicar va desde la página (el pedido lleva el origen, como el botón de Mi tienda).
	expect(
		await page.evaluate(
			async () => (await fetch('/panel/mi-tienda/publicar', { method: 'POST' })).status
		)
	).toBe(204);

	await page.goto(`/t/${slug}`);
	await listo(page);
	// $ 7.800 + $ 12.600 = $ 20.400, con el 10 %: $ 18.360.
	await expect(page.getByText(/Ahorrás \$\s2\.040/)).toBeVisible();
	await page.getByRole('button', { name: 'Agregar los 2' }).click();

	await page.goto(`/t/${slug}/carrito`);
	await listo(page);
	await expect(page.getByText('En combo')).toHaveCount(2);
	await expect(page.locator('.tv-carrito-resumen')).toContainText('Combo');
	await expect(page.locator('.tv-carrito-resumen .tv-cuentas-total')).toContainText('18.360');
	expect(
		await page.evaluate(() => document.documentElement.scrollWidth - innerWidth)
	).toBeLessThanOrEqual(0);

	await page.getByRole('link', { name: 'Finalizar compra' }).first().click();
	await expect(page).toHaveURL(/\/checkout$/);
	await listo(page);
	await page.getByLabel('Email').fill('combo@ejemplo.com');
	await page.getByLabel('Nombre y apellido').fill('Lucía Pérez');
	await page.getByLabel('Teléfono').fill('11 5555-0101');
	await page.getByRole('radio', { name: /Retiro en el local/ }).click();
	await page.getByRole('radio', { name: /Mercado Pago/ }).click();
	await expect(page.locator('.tv-checkout-resumen .tv-cuentas-total')).toContainText('18.360');
	await page.getByRole('button', { name: /^Pagar/ }).click();
	await expect(page.getByRole('heading', { name: '¡Gracias, Lucía!' })).toBeVisible();
	await expect(page.getByText(/18\.360/).first()).toBeVisible();

	// Sacar uno del combo: el otro queda, sin descuento.
	await page.goto(`/t/${slug}`);
	await listo(page);
	await page.getByRole('button', { name: 'Agregar los 2' }).click();
	await page.goto(`/t/${slug}/carrito`);
	await listo(page);
	await page.getByRole('button', { name: /^Quitar Miel/ }).click();
	await expect(page.getByText('En combo')).toHaveCount(0);
	await expect(page.locator('.tv-carrito-resumen')).not.toContainText('Combo');
});
