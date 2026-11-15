import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });

test('mi tienda en bloques: agregar, reordenar, variante, tocar en la vista, editar, deshacer y publicar', async ({
	page,
	isMobile
}) => {
	await page.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
	await page.goto('/prueba');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	await page.goto('/panel/mi-tienda');
	await listo(page);
	await expect(page.getByRole('button', { name: 'Todo publicado' })).toBeVisible();

	let guardados = 0;
	page.on('response', (r) => {
		if (
			r.url().endsWith('/panel/mi-tienda') &&
			r.request().method() === 'PUT' &&
			r.status() === 204
		)
			guardados++;
	});
	const filas = page.locator('[data-fila]');
	const nombres = () => filas.locator('.txt b').allInnerTexts();
	const inicial = await nombres();
	expect(inicial[0]).toBe('Portada');

	// Agregar un banner al final, desde la biblioteca; se abre su panel.
	await page.getByRole('button', { name: 'Agregar bloque', exact: true }).click();
	await page
		.getByRole('dialog')
		.getByRole('button', { name: /^Banner/ })
		.click();
	await expect(page.getByRole('heading', { name: 'Banner' })).toBeVisible();
	await expect(page.getByText(/No se ve todavía\. Le falta: título/)).toBeVisible();

	// Editar un campo: se guarda al salir del campo y el aviso se va.
	await page
		.getByLabel(/^Título/)
		.first()
		.fill('Hot Sale de mates');
	await page
		.getByLabel(/^Título/)
		.first()
		.press('Tab');
	await expect(page.getByText(/No se ve todavía/)).toBeHidden();
	await expect.poll(() => guardados).toBe(2);

	// Cambiar la variante.
	await page.getByRole('radio', { name: 'Doble' }).click();
	await expect(page.getByRole('radio', { name: 'Doble' })).toHaveAttribute('aria-checked', 'true');
	await expect.poll(() => guardados).toBe(3);

	// Reordenar con el teclado: el banner sube arriba de todo.
	await page.getByRole('button', { name: 'Bloques', exact: true }).click();
	await page.getByRole('button', { name: 'Mover Banner (con las flechas)' }).press('Home');
	await expect.poll(nombres).toEqual(['Banner', ...inicial]);

	// Deshacer (vuelve al final) y rehacer.
	await page.getByRole('button', { name: 'Deshacer' }).click();
	await expect.poll(nombres).toEqual([...inicial, 'Banner']);
	await page.getByRole('button', { name: 'Rehacer' }).click();
	await expect.poll(nombres).toEqual(['Banner', ...inicial]);
	await expect.poll(() => guardados).toBe(6);

	// Tocar un bloque en la vista previa abre su panel en el editor.
	if (isMobile) await page.getByRole('button', { name: 'Vista previa' }).click();
	const vista = page.frameLocator('iframe[title="Vista previa de tu tienda"]');
	await vista.locator('[data-bloque="productos"]').first().click();
	await expect(page.getByRole('heading', { name: 'Productos' })).toBeVisible();

	// Lo guardado sigue al recargar.
	await page.reload();
	await listo(page);
	await expect.poll(nombres).toEqual(['Banner', ...inicial]);

	// Publicar: la tienda pública muestra el banner.
	await page.getByRole('button', { name: 'Publicar cambios' }).click();
	await expect(page.getByText('Cambios publicados')).toBeVisible();
	await expect(page.getByRole('button', { name: 'Todo publicado' })).toBeVisible();
	const tienda = await page.getByRole('link', { name: 'Ver mi tienda' }).getAttribute('href');
	await page.goto(tienda!);
	await expect(page.getByText('Hot Sale de mates').first()).toBeVisible();
});
