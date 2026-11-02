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

/** Mañana a las 19:00 (hora de Argentina), para el <input type="datetime-local">. */
function manana() {
	const d = new Date(Date.now() - 3 * 3600_000 + 86_400_000);
	return `${d.toISOString().slice(0, 10)}T19:00`;
}

test('lanzamiento: «Próximamente» con cuenta, aviso, reserva con seña y saldado', async ({
	page
}) => {
	const slug = await demo(page);
	await page.goto('/panel/productos');
	await listo(page);
	await page.getByRole('button', { name: 'Nuevo producto' }).click();
	// Primero pregunta qué es.
	await page
		.getByRole('dialog')
		.getByRole('button', { name: /^Otra cosa/ })
		.click();
	const ed = page.getByRole('dialog');
	await ed.getByLabel('Nombre', { exact: true }).fill('Drop de prueba');
	await ed.getByLabel('Precio base').fill('40000');
	await ed.getByLabel('Precio base').blur();
	await ed.getByLabel('Stock', { exact: true }).fill('3');
	await ed.getByLabel('Stock', { exact: true }).blur();
	await ed.getByRole('switch', { name: 'Programar el lanzamiento' }).click();
	await ed.getByLabel('Sale a la venta (hora de Argentina)').fill(manana());
	await ed.getByRole('switch', { name: 'Reservar con seña' }).click();
	await ed.getByLabel('Porcentaje de seña de la reserva').fill('25');
	await ed.getByLabel('Porcentaje de seña de la reserva').blur();
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	await ed.getByRole('button', { name: /Guardar/ }).click();
	await expect(ed).toBeHidden();

	// En la tienda: «Próximamente», cuándo sale y su cuenta; no se agrega al carrito.
	await page.goto(`/t/${slug}/catalogo`);
	await listo(page);
	await page
		.getByRole('link', { name: /Drop de prueba/ })
		.first()
		.click();
	await listo(page);
	await expect(page.getByText(/Sale el .* a las 19:00/)).toBeVisible();
	await expect(page.getByRole('button', { name: 'Próximamente' })).toBeVisible();
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);

	await page.getByLabel('Tu email para el aviso').fill('ana@ejemplo.com');
	await page.getByRole('button', { name: 'Avisame cuando salga' }).click();
	await expect(
		page.getByText('Listo: te escribimos a ana@ejemplo.com cuando salga.')
	).toBeVisible();

	// Reservar con seña: en la demo no se cobra y la confirmación lo dice.
	await expect(page.getByText(/\$ 10\.000.*de seña \(25%\)/)).toBeVisible();
	await page.getByRole('button', { name: 'Reservar con seña' }).click();
	await page.getByLabel('Tu nombre').fill('Lucía Pérez');
	await page.getByLabel('Email', { exact: true }).fill('lucia@ejemplo.com');
	await page.getByRole('button', { name: 'Pagar la seña de $ 10.000' }).click();
	await expect(page).toHaveURL(/\/pedido\//);
	await listo(page);
	await expect(page.getByText('Seña pagada · falta $ 30.000')).toBeVisible();
	await expect(page.getByText(/En la demo no se cobra/)).toBeVisible();
	const numero = (page.url().split('/pedido/')[1] ?? '').split('-')[0];

	// En el panel: la hoja dice cuánto falta y se marca como saldado (con Deshacer).
	await page.goto(`/panel/pedidos?pedido=${numero}`);
	await listo(page);
	const hoja = page.getByRole('dialog', { name: `Pedido #${numero}` });
	await expect(hoja.getByText('Seña pagada · falta $ 30.000')).toBeVisible();
	await hoja.getByRole('button', { name: 'Marcar como saldado' }).click();
	await expect(page.getByText(`Pedido #${numero} saldado`)).toBeVisible();
	await expect(hoja.getByText('Seña y saldo cobrados')).toBeVisible();

	// En el producto: cuántos pidieron aviso y cuántos reservaron.
	await page.goto('/panel/productos');
	await listo(page);
	await page.locator('.fila', { hasText: 'Drop de prueba' }).locator('.abrir').click();
	await expect(page.getByRole('dialog').getByText(/1\s+pidió aviso/)).toBeVisible();
	await expect(page.getByRole('dialog').getByText(/1\s+reservó/)).toBeVisible();
});
