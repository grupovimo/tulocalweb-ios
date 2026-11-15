import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });

async function demo(page: Page, ruta: string) {
	await page.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
	await page.goto('/prueba');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	await page.goto(ruta);
	await listo(page);
}

test('bandeja: abrir, mandar la sugerencia, escribir y tomar', async ({ page, isMobile }) => {
	await demo(page, '/panel/whatsapp');
	await expect(page.getByRole('heading', { name: 'Bandeja' })).toBeVisible();
	await page.getByRole('button', { name: /^Te esperan/ }).click();
	await expect(page.locator('.conv')).toHaveCount(2);
	await page.locator('.conv').first().click();
	await expect(page).toHaveURL(/\?c=c2/);
	await expect(page.locator('.sugerencia')).toContainText('dirección');
	await page.locator('.sugerencia').getByRole('button', { name: 'Enviar' }).click();
	await expect(page.locator('.msg.yo').last()).toContainText('dirección');
	await expect(page.locator('.sugerencia')).toHaveCount(0);

	await page.getByRole('textbox', { name: 'Escribí un mensaje' }).fill('Listo, ya quedó cambiada');
	await page.getByRole('button', { name: 'Mandar' }).click();
	await expect(page.locator('.msg.yo').last()).toContainText('Listo, ya quedó cambiada');
	await page.getByRole('button', { name: 'Tomar' }).click();
	await expect(page.getByText('Tomaste la conversación')).toBeVisible();
	await expect(page.getByRole('button', { name: 'Devolver a Tincho' })).toBeVisible();
	if (isMobile) {
		await page.getByRole('button', { name: 'Volver a la lista' }).click();
		await expect(page.locator('.conv').first()).toBeVisible();
	}
});

test('carritos: recordar uno, todos y los automáticos', async ({ page }) => {
	await demo(page, '/panel/carritos-abandonados');
	await expect(page.getByText('En carritos abiertos')).toBeVisible();
	await page.locator('.carrito').first().getByRole('button', { name: 'Recordar' }).click();
	await expect(page.getByText('Recordatorio enviado')).toBeVisible();
	await expect(page.locator('.carrito').first()).toContainText('Recordado');
	await page.getByRole('button', { name: /Recordar a todos/ }).click();
	await expect(page.getByText('12 recordatorios en camino')).toBeVisible();
	await expect(page.getByRole('button', { name: /^Recordar$/ })).toHaveCount(0);

	// Los tres recordatorios por mail (§62): en la demo arrancan prendidos.
	await page.getByRole('switch', { name: 'Tercero recordatorio' }).click();
	await expect(page.getByText('Tercero recordatorio apagado')).toBeVisible();
	await page.reload();
	await listo(page);
	await expect(page.getByRole('switch', { name: 'Tercero recordatorio' })).toHaveAttribute(
		'aria-checked',
		'false'
	);
	// El Resumen también los ve recordados.
	await page.goto('/panel');
	await listo(page);
	await expect(page.locator('.burbuja')).toContainText('Listo');
});
