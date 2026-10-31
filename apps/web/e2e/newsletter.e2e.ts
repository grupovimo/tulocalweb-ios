import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });
const sinDesborde = (page: Page) =>
	page.evaluate(() => document.documentElement.scrollWidth - innerWidth);

/** Entra a la demo con la plantilla del botón y devuelve la dirección de su tienda. */
async function demo(page: Page, boton: RegExp) {
	await page.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
	await page.goto('/prueba');
	await listo(page);
	await page.getByRole('button', { name: boton }).first().click();
	await expect(page).toHaveURL(/\/panel$/);
	return (await page.locator('.t-d').first().textContent())!.split('/t/')[1].trim();
}

test('newsletter: sumarse desde el pie, confirmar, aparecer en el panel y darse de baja', async ({
	page
}) => {
	const slug = await demo(page, /^Litoral/);
	await page.goto(`/t/${slug}`);
	await listo(page);
	const pie = page.locator('footer');
	await pie.getByLabel('Tu email').fill('no-es-un-email');
	await pie.getByRole('button', { name: 'Sumarme' }).click();
	await expect(pie.getByRole('status')).toContainText('Revisá el email.');
	await pie.getByLabel('Tu email').fill('nueva@ejemplo.com');
	await pie.getByRole('button', { name: 'Sumarme' }).click();
	// En la demo no sale el mail: el aviso trae el link de confirmar.
	await expect(pie.getByRole('status')).toContainText('En la demo no mandamos el mail');
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	const confirmar = (await pie.getByRole('link', { name: 'Confirmar acá' }).getAttribute('href'))!;
	expect(confirmar).toMatch(new RegExp(`^/t/${slug}/newsletter/confirmar/[\\w-]{20,}$`));

	// Abrir el link no confirma: hace falta el botón.
	await page.goto('/panel/newsletter');
	await listo(page);
	await expect(page.locator('.sus', { hasText: 'nueva@ejemplo.com' })).toContainText(
		'Sin confirmar'
	);
	await page.goto(confirmar);
	await listo(page);
	await expect(page.getByRole('heading', { name: 'Confirmá tu suscripción' })).toBeVisible();
	await page.getByRole('button', { name: 'Confirmar mi suscripción' }).click();
	await expect(page.getByText('¡Listo, ya estás suscripto!')).toBeVisible({ timeout: 30_000 });
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	await page.goto('/panel/newsletter');
	await listo(page);
	await expect(page.locator('.sus', { hasText: 'nueva@ejemplo.com' })).toContainText('Confirmado');

	// La baja de un clic (RFC 8058): el POST del servidor de correo, sin Origin.
	const baja = confirmar.replace('/confirmar/', '/baja/');
	const r = await page.request.post(baja, {
		form: { 'List-Unsubscribe': 'One-Click' },
		headers: { origin: '' }
	});
	expect(r.status()).toBe(200);
	await page.goto('/panel/newsletter');
	await listo(page);
	await expect(page.locator('.sus', { hasText: 'nueva@ejemplo.com' })).toHaveCount(0);
	// El link ya no sirve.
	await page.goto(baja);
	await listo(page);
	await page.getByRole('button', { name: 'Darme de baja' }).click();
	await expect(page.getByText('Ese link ya no sirve')).toBeVisible();
});

test('newsletter: mandar una novedad con vista previa y confirmación, borrar con Deshacer y exportar', async ({
	page
}) => {
	await demo(page, /^Litoral/);
	await page.goto('/panel/newsletter');
	await listo(page);
	await expect(page.getByText('En la demo no sale ningún mail')).toBeVisible();
	const n = Number((await page.locator('.o-v').textContent())!.replace(/\D/g, ''));
	expect(n).toBeGreaterThan(0);
	const mandar = page.getByRole('button', { name: `Mandar a ${n} personas` });
	await expect(mandar).toBeDisabled();
	await page.getByLabel('Asunto').fill('Llegó la yerba nueva');
	await page
		.getByLabel('Texto')
		.fill('Hola.\n\nYa está la cosecha nueva, con envío a todo el país.');
	const productos = page.getByRole('group', { name: 'Productos destacados' });
	const primero = productos.getByRole('button').first();
	const nombre = (await primero.textContent())!.trim();
	await primero.click();
	await expect(productos.getByRole('button', { pressed: true })).toHaveCount(1);
	// La vista previa sigue lo que se escribe.
	const previa = page.locator('.mail');
	await expect(previa.locator('.m-asunto')).toHaveText('Llegó la yerba nueva');
	await expect(previa.locator('p')).toHaveCount(2);
	await expect(previa.locator('.m-prod')).toContainText(nombre);
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);

	await mandar.click();
	const modal = page.getByRole('dialog');
	await expect(modal).toContainText('¿Mandar «Llegó la yerba nueva»?');
	await expect(modal).toContainText('no sale ningún mail');
	await modal.getByRole('button', { name: `Mandar a ${n} personas` }).click();
	await expect(page.getByText('Novedad guardada')).toBeVisible();
	await expect(page.locator('.campana').first()).toContainText('Llegó la yerba nueva');
	await expect(page.locator('.campana').first()).toContainText('Enviada');

	// Borrar uno: se saca de la lista y Deshacer lo devuelve.
	const fila = page.locator('.sus').first();
	const email = (await fila.locator('.n').textContent())!;
	await fila.getByRole('button', { name: `Sacar a ${email} de la lista` }).click();
	await expect(page.locator('.sus', { hasText: email })).toHaveCount(0);
	await page.getByRole('button', { name: 'Deshacer' }).click();
	await expect(page.locator('.sus', { hasText: email })).toHaveCount(1);

	const descarga = page.waitForEvent('download');
	await page.getByRole('link', { name: 'Exportar CSV' }).click();
	expect((await descarga).suggestedFilename()).toMatch(/^newsletter-\d{4}-\d{2}-\d{2}\.csv$/);
});

test('newsletter: la casilla del checkout en una plantilla sin bloque en el diseño', async ({
	page
}) => {
	const slug = await demo(page, /^Andes/);
	await page.goto(`/t/${slug}`);
	await listo(page);
	await expect(page.locator('footer').getByLabel('Tu email')).toHaveCount(0);
	// Un producto que se agrega sin elegir opciones.
	await page.goto(`/t/${slug}/catalogo`);
	await listo(page);
	const links = await page
		.locator(`a[href^="/t/${slug}/producto/"]`)
		.evaluateAll((as) => [...new Set(as.map((a) => a.getAttribute('href')!))]);
	let agregado = false;
	for (const h of links) {
		await page.goto(h);
		await listo(page);
		const b = page.getByRole('button', { name: 'Agregar al carrito' }).first();
		if ((await b.count()) && (await b.isEnabled())) {
			await b.click();
			agregado = true;
			break;
		}
	}
	expect(agregado).toBe(true);
	await page.goto(`/t/${slug}/checkout`);
	await listo(page);
	const casilla = page.getByRole('checkbox', { name: 'Quiero recibir novedades' });
	await expect(casilla).not.toBeChecked();
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	await page.getByLabel('Email').fill('compra@ejemplo.com');
	await page.getByLabel('Nombre y apellido').fill('Lucía Pérez');
	await page.getByLabel('Teléfono').fill('11 5555-0101');
	await page.getByLabel('Código postal').fill('1414');
	await page.getByLabel(/Dirección/).fill('Gorriti 4500');
	await casilla.check();
	await page.getByRole('radio', { name: /Retiro en el local/ }).click();
	await page.getByRole('radio', { name: /Transferencia/ }).click();
	await page.getByRole('button', { name: /^Pagar/ }).click();
	await expect(page.getByRole('heading', { name: /Gracias/ })).toBeVisible();
	await page.goto('/panel/newsletter');
	await listo(page);
	const fila = page.locator('.sus', { hasText: 'compra@ejemplo.com' });
	await expect(fila).toContainText('Sin confirmar');
	await expect(fila).toContainText('Checkout');
});
