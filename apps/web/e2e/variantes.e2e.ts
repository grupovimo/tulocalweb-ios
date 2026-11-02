import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });
const sinDesborde = (page: Page) =>
	page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);

async function demo(page: Page) {
	await page.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
	await page.goto('/prueba');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	return (await page.locator('.t-d').first().textContent())!.split('/t/')[1].trim();
}

test('variantes: se cargan en el panel y se compran eligiendo el talle', async ({
	page,
	isMobile
}) => {
	const slug = await demo(page);
	await page.goto('/panel/productos');
	await listo(page);
	await page.getByRole('button', { name: 'Nuevo producto' }).click();
	// Primero pregunta qué es.
	await page.getByRole('dialog').getByRole('button', { name: /^Ropa y accesorios/ }).click();
	const ed = page.getByRole('dialog');
	await ed.getByLabel('Nombre', { exact: true }).fill('Buzo de prueba');
	await ed.getByLabel('Precio base').fill('30000');
	await ed.getByLabel('Precio base').blur();

	// Talle con dos valores (separados por coma) y Color con uno: dos combinaciones.
	await ed.getByRole('button', { name: 'Sumar talle, color u otra opción' }).click();
	await ed.getByLabel('Sumar valores a Talle').fill('S, M');
	await ed.getByLabel('Sumar valores a Talle').press('Enter');
	await ed.getByRole('button', { name: 'Sumar otra opción' }).click();
	await ed.getByLabel('Sumar valores a Color').fill('Negro');
	await ed.getByLabel('Sumar valores a Color').press('Enter');
	await expect(ed.getByText('2 combinaciones · 0 u en total')).toBeVisible();
	await ed.getByLabel('Stock de S · Negro').fill('3');
	await ed.getByLabel('Stock de S · Negro').blur();
	await ed.getByLabel('Stock de M · Negro').fill('5');
	await ed.getByLabel('Stock de M · Negro').blur();
	await ed.getByLabel('Precio de M · Negro').fill('32000');
	await ed.getByLabel('Precio de M · Negro').blur();
	await expect(ed.getByText('2 combinaciones · 8 u en total')).toBeVisible();
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	await ed.getByRole('button', { name: /Guardar/ }).click();
	await expect(ed).toBeHidden();

	// Al reabrirlo, las opciones y el stock por combinación siguen ahí.
	await page.locator('.fila', { hasText: 'Buzo de prueba' }).locator('.abrir').click();
	const otra = page.getByRole('dialog');
	await expect(otra.getByLabel('Stock de M · Negro')).toHaveValue('5');
	// En el celular, editando, cierra la cruz de arriba (no hay «Cancelar» abajo).
	await otra.getByRole('button', { name: isMobile ? 'Cerrar' : 'Cancelar' }).click();
	await expect(otra).toBeHidden();

	// En la tienda (la yerba de la demo viene en 500 g y 1 kg): hay que elegir la presentación y el
	// carrito lleva la variante.
	await page.goto(`/t/${slug}/producto/yerba-barbacua-1-kg`);
	await listo(page);
	await expect(page.getByText(/Desde/).first()).toBeVisible();
	await page.getByRole('button', { name: 'Elegí una presentación' }).click();
	await page.getByRole('button', { name: /^1 kg/ }).click();
	await page
		.getByRole('button', { name: /^(Agregar|Sumar)/ })
		.first()
		.click();
	await page.goto(`/t/${slug}/carrito`);
	await listo(page);
	await expect(page.getByText(/Presentación 1 kg/).first()).toBeVisible();
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
});

test('variantes: el punto de venta vende cada combinación', async ({ page }) => {
	await demo(page);
	await page.goto('/panel/ventas-en-local');
	await listo(page);
	await page
		.getByRole('button', { name: /^Sumar Mate de calabaza · Color/ })
		.first()
		.click();
	await expect(page.locator('.ticket .i-n').first()).toContainText('Mate de calabaza · ');
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	await page
		.getByRole('group', { name: 'Medio de pago' })
		.getByRole('button', { name: 'Efectivo' })
		.click();
	await page.getByRole('button', { name: /^Cobrar \$/ }).click();
	await expect(page.getByText(/^Cobrado \$/)).toBeVisible();
	await page.getByRole('button', { name: 'Ver pedido' }).click();
	await expect(page.getByRole('dialog', { name: /Pedido #\d+/ })).toContainText('Color');
});
