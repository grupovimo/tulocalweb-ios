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

// Precios mayoristas (§49): el mínimo en Configuración, el precio en la ficha del producto, la
// tienda avisa cuánto falta y al llegar cobra el mayorista; el pedido entra como mayorista.
test('venta mayorista: mínimo, precio del producto, carrito y pedido', async ({ page }) => {
	const slug = await demo(page);

	// Configuración › Venta mayorista: 3 unidades.
	await page.goto('/panel/configuracion');
	await listo(page);
	await page.getByRole('button', { name: 'Venta mayorista' }).click();
	await expect(page.getByRole('switch', { name: 'Venta mayorista' })).toBeChecked();
	await page.getByLabel('Unidades').fill('3');
	await page.getByRole('button', { name: 'Guardar mínimo' }).click();
	await expect(page.getByText('Mínimo guardado')).toBeVisible();
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);

	// El editor del producto: la caja «Mayorista» con el mínimo de verdad.
	await page.goto('/panel/productos?editar=bombilla-alpaca');
	await listo(page);
	await expect(page.getByText('Desde 3 u')).toBeVisible();
	const mayorista = page.getByLabel('Precio mayorista', { exact: true });
	await mayorista.fill('10000');
	await mayorista.blur();
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	await page.getByRole('button', { name: 'Guardar y publicar' }).click();
	await expect(page.getByText('Producto guardado')).toBeVisible();

	// La ficha dice el precio por mayor; con 1 en el carrito falta, con 3 se aplica.
	await page.goto(`/t/${slug}/producto/bombilla-alpaca`);
	await listo(page);
	await expect(page.getByText(/Por mayor: \$\s10\.000 desde 3\sunidades/)).toBeVisible();
	await page.getByRole('button', { name: 'Agregar al carrito' }).click();
	await expect(page).toHaveURL(/\/carrito$/);
	await expect(page.getByText(/Sumando/)).toContainText('2 unidades más');
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	await page.getByRole('button', { name: 'Uno más' }).first().click();
	await page.getByRole('button', { name: 'Uno más' }).first().click();
	await expect(page.getByText(/aplicado: ahorrás/)).toBeVisible();
	await page.getByRole('link', { name: 'Finalizar compra' }).click();

	await page.getByLabel('Email').fill('mayorista@ejemplo.com');
	await page.getByLabel('Nombre y apellido').fill('Almacén Norte');
	await page.getByLabel('Teléfono').fill('11 5555-0102');
	await page.getByLabel('Código postal').fill('1414');
	await page.getByLabel(/Dirección/).fill('Gorriti 4500');
	await page.getByRole('radio', { name: /Retiro en el local/ }).click();
	await page.getByRole('radio', { name: /Transferencia/ }).click();
	await page.getByRole('button', { name: /^Pagar/ }).click();
	await expect(page.getByRole('heading', { name: /¡Gracias, Almacén!/ })).toBeVisible();
	// 3 × $ 10.000, retiro sin costo y 15 % por transferencia.
	await expect(page.getByText(/25\.500/).first()).toBeVisible();
	const numero = (await page
		.locator('b', { hasText: /^#\d+$/ })
		.first()
		.textContent())!.slice(1);

	// En Pedidos › Mayoristas.
	await page.goto(`/panel/pedidos?mayorista=1&q=%23${numero}`);
	await listo(page);
	await expect(page.locator('.fila').first()).toContainText('Almacén Norte');
});
