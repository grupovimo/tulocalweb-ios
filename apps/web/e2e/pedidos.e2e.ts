import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });

async function demo(page: Page) {
	await page.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
	await page.goto('/prueba');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	await page.goto('/panel/pedidos');
	await listo(page);
}

/** Carga un pedido propio, como «Crear pedido», con un producto de Litoral sin variantes y con stock. */
async function pedidoPropio(page: Page, extra: { pagado: boolean }) {
	for (const productoId of [
		'mate-de-calabaza',
		'miel-de-monte',
		'dulce-de-mamon',
		'bombilla-alpaca'
	]) {
		const r = await page.request.post('/panel/pedidos/nuevo', {
			data: { items: [{ productoId, cantidad: 1 }], pago: 'transf', canal: 'wa', ...extra }
		});
		if (r.status() === 201) return String(((await r.json()) as { numero: number }).numero);
	}
	throw new Error('No se pudo cargar un pedido propio');
}

test('lista, filtra por la URL y vuelve a todos', async ({ page }) => {
	await demo(page);
	await expect(page.getByRole('heading', { name: 'Pedidos' })).toBeVisible();
	await expect(page.locator('.fila').first()).toContainText('#1048');
	expect(
		await page.evaluate(
			() => document.documentElement.scrollWidth - document.documentElement.clientWidth
		)
	).toBeLessThanOrEqual(0);

	await page
		.getByRole('group', { name: 'Etapas' })
		.getByRole('button', { name: /Para preparar/ })
		.click();
	await expect(page).toHaveURL(/estado=para-preparar/);
	await expect(page.locator('.fila .p-e, .fila .accion').first()).toBeVisible();
	await page.getByRole('button', { name: 'Sin pagar' }).first().click();
	await expect(page).toHaveURL(/estado=sin-pagar$/);
	await page.getByRole('textbox', { name: 'Buscar pedidos' }).fill('zzzz');
	await expect(page.getByText('Ningún pedido con esos filtros')).toBeVisible();
	await page.getByRole('button', { name: 'Limpiar filtros' }).click();
	await expect(page.locator('.fila').first()).toContainText('#1048');

	// Al recargar, los filtros siguen ahí.
	await page.goto('/panel/pedidos?canal=ml&orden=monto');
	await listo(page);
	await expect(page.getByText('Mercado Libre').first()).toBeVisible();
	await expect(page.getByRole('button', { name: /Orden: Mayor monto/ })).toBeVisible();
});

test('preparar un pedido, deshacerlo y verlo en la hoja', async ({ page }) => {
	await demo(page);
	// Uno propio (los de la demo dependen de la hora: el primero puede ser un retiro).
	const numero = await pedidoPropio(page, { pagado: true });
	await page.goto(`/panel/pedidos?estado=para-preparar&q=%23${numero}`);
	await listo(page);
	const fila = page.locator('.fila').filter({ hasText: `#${numero}` });
	// Un retiro en el local queda «listo para retirar» en vez de «preparado».
	await fila
		.getByRole('button', { name: new RegExp(`(Preparar|Listo para retirar) #${numero}`) })
		.click();
	await expect(
		page.getByText(new RegExp(`^#${numero} (preparado|listo para retirar)$`))
	).toBeVisible();
	await page.getByRole('button', { name: 'Deshacer' }).click();
	await expect(page.locator('.fila').filter({ hasText: `#${numero}` })).toBeVisible();

	await page
		.locator('.fila')
		.filter({ hasText: `#${numero}` })
		.locator('.abrir')
		.click();
	const hoja = page.getByRole('dialog', { name: `Pedido #${numero}` });
	await expect(hoja.getByText('Historial')).toBeVisible();
	await hoja.getByRole('button', { name: /^Marcar como (preparado|listo para retirar)$/ }).click();
	await expect(hoja.getByText(/^Listo para (despachar|retirar)$/).first()).toBeVisible();
});

test('seleccionar varios y prepararlos en lote', async ({ page }) => {
	await demo(page);
	await page.goto('/panel/pedidos?estado=para-preparar');
	await listo(page);
	const antes = Number(await page.locator('.etapa').nth(1).locator('.e-n').textContent());
	await page.locator('.fila .tilde').nth(0).click();
	await page.locator('.fila .tilde').nth(1).click();
	await expect(page.locator('.lote')).toContainText('2 seleccionados');
	await expect(page.locator('.lote')).toBeVisible();
	await page.locator('.lote').getByRole('button', { name: 'Preparar' }).click();
	await expect(page.getByText('Preparar · 2 pedidos')).toBeVisible();
	await expect(page.locator('.etapa').nth(1).locator('.e-n')).toHaveText(String(antes - 2));
});

test('el Resumen abre un pedido y exporta CSV', async ({ page }) => {
	await demo(page);
	await page.goto('/panel/pedidos?pedido=1048');
	await listo(page);
	await expect(page.getByRole('dialog', { name: 'Pedido #1048' })).toBeVisible();
	const r = await page.request.get('/panel/pedidos/exportar?canal=wa');
	expect(r.headers()['content-type']).toContain('text/csv');
	const csv = await r.text();
	expect(csv.split('\r\n')[0]).toContain('Número;Fecha;Cliente');
	// Canal en castellano, y sólo los de WhatsApp (el filtro de la URL).
	const filas = csv.trim().split('\r\n').slice(1);
	expect(filas.length).toBeGreaterThan(0);
	for (const f of filas) expect(f.split(';')[4]).toBe('WhatsApp');
});

test('marcar pagado una transferencia desde la fila y deshacerlo', async ({ page }) => {
	await demo(page);
	// Los pedidos de la demo dependen de la hora (a la mañana puede no haber ninguno para preparar
	// por transferencia): se carga uno sin pagar, como «Crear pedido», con un producto de Litoral
	// sin variantes (los que tienen piden elegir una).
	let numero = '';
	for (const productoId of [
		'mate-de-calabaza',
		'miel-de-monte',
		'dulce-de-mamon',
		'bombilla-alpaca'
	]) {
		const r = await page.request.post('/panel/pedidos/nuevo', {
			data: { items: [{ productoId, cantidad: 1 }], pago: 'transf', canal: 'wa', pagado: false }
		});
		if (r.status() !== 201) continue;
		numero = String(((await r.json()) as { numero: number }).numero);
		break;
	}
	expect(numero).not.toBe('');
	await page.goto(`/panel/pedidos?estado=sin-pagar&q=%23${numero}`);
	await listo(page);
	const fila = page.locator('.fila').filter({ hasText: `#${numero}` });
	const boton = fila.getByRole('button', { name: `Marcar pagado #${numero}` });
	// En el celular, botón visible de 44 px.
	const caja = (await boton.boundingBox())!;
	expect(caja.height).toBeGreaterThanOrEqual(page.viewportSize()!.width < 760 ? 44 : 30);
	await boton.click();
	await expect(page.getByText(`#${numero} pagado`)).toBeVisible();
	await expect(page.locator('.fila').filter({ hasText: `#${numero}` })).toHaveCount(0);
	await page.getByRole('button', { name: 'Deshacer' }).click();
	await expect(page.locator('.fila').filter({ hasText: `#${numero}` })).toBeVisible();
});

test('WhatsApp con el mensaje del estado, desde la fila y la hoja', async ({ page }) => {
	await demo(page);
	// Uno en camino propio: los de la demo dependen de la hora (a la mañana no hay ninguno).
	let numero = '';
	for (const productoId of [
		'mate-de-calabaza',
		'miel-de-monte',
		'dulce-de-mamon',
		'bombilla-alpaca'
	]) {
		const r = await page.request.post('/panel/pedidos/nuevo', {
			data: {
				items: [{ productoId, cantidad: 1 }],
				pago: 'transf',
				canal: 'wa',
				envio: 'moto',
				cliente: 'Lu Pérez',
				telefono: '1155556666'
			}
		});
		if (r.status() !== 201) continue;
		numero = String(((await r.json()) as { numero: number }).numero);
		break;
	}
	expect(numero).not.toBe('');
	for (const estado of ['Listo para despachar', 'En camino']) {
		const r = await page.request.post(`/panel/pedidos/${numero}`, { data: { estado } });
		expect(r.status(), estado).toBe(204);
	}
	await page.goto(`/panel/pedidos?estado=en-camino&q=%23${numero}`);
	await listo(page);
	const href = (await page
		.locator('.fila')
		.filter({ hasText: `#${numero}` })
		.locator('a.wa')
		.getAttribute('href'))!;
	expect(href).toMatch(/^https:\/\/wa\.me\/549\d{10}\?text=/);
	expect(decodeURIComponent(href.split('text=')[1])).toMatch(
		/^¡Hola \S+! Tu pedido #\d+ ya (salió para tu casa|está en camino)\./
	);

	await page.goto('/panel/pedidos?estado=entregado&canal=web');
	await listo(page);
	// Uno con teléfono (sin teléfono no hay a quién escribirle).
	await page
		.locator('.fila')
		.filter({ has: page.locator('a.wa') })
		.first()
		.locator('.abrir')
		.click();
	const escribir = page.getByRole('dialog').getByRole('link', { name: 'Escribir' });
	const texto = decodeURIComponent((await escribir.getAttribute('href'))!.split('text=')[1]);
	expect(texto).toMatch(/dejanos una reseña: https:\/\/tulocalweb\.com\.ar\/t\/.+\/pedido\/\d+-/);
});

test('volver a vender abre Crear pedido completo para revisar', async ({ page }) => {
	await demo(page);
	// Uno propio con un producto con stock (los productos del 1048 de la demo dependen de la hora).
	const numero = await pedidoPropio(page, { pagado: true });
	await page.goto(`/panel/pedidos?pedido=${numero}`);
	await listo(page);
	const hoja = page.getByRole('dialog', { name: `Pedido #${numero}` });
	await expect(hoja.locator('.item').first()).toBeVisible();
	const productos = await hoja.locator('.item').count();
	await hoja.getByRole('link', { name: 'Volver a vender' }).click();
	await expect(page).toHaveURL(new RegExp(`/panel/pedidos/nuevo\\?desde=${numero}`));
	await expect(page.getByText(`Lo mismo que el pedido #${numero}`)).toBeVisible();
	await expect(page.locator('.resumen .item').first()).toBeVisible();
	expect(await page.locator('.resumen .item').count()).toBeLessThanOrEqual(productos);
	await expect(page.getByRole('button', { name: /^Crear/ })).toBeEnabled();
});
