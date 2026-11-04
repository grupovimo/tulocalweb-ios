import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });

test('mi tienda: ocultar, reordenar, acento y combinación de colores quedan sin publicar hasta publicar', async ({
	page
}) => {
	await page.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
	await page.goto('/prueba');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	await page.goto('/panel/mi-tienda');
	await listo(page);
	await expect(page.getByRole('button', { name: 'Todo publicado' })).toBeVisible();
	// La vista previa es la tienda de verdad con el borrador: se recarga en cada guardado.
	const previa = page.locator('iframe[title="Vista previa de tu tienda"]');
	await expect(previa).toHaveAttribute('src', /\/panel\/mi-tienda\/vista\?v=0$/);
	// Los tres cambios se guardan uno detrás del otro: se esperan los tres antes de recargar.
	let guardados = 0;
	page.on('response', (r) => {
		if (
			r.url().endsWith('/panel/mi-tienda') &&
			r.request().method() === 'PUT' &&
			r.status() === 204
		)
			guardados++;
	});
	await page.getByRole('button', { name: 'Ocultar Botón de WhatsApp' }).click();
	await expect(previa).not.toHaveAttribute('src', /v=0$/);
	await page.getByRole('button', { name: 'Subir Reseñas' }).click();
	await page
		.getByRole('group', { name: 'Color de acento' })
		.getByRole('button', { name: 'Color 4' })
		.click();
	await expect(page.getByRole('button', { name: 'Publicar cambios' })).toBeVisible();
	await page.getByRole('button', { name: /Oscura/ }).click();
	await expect.poll(() => guardados, { timeout: 15_000 }).toBe(4);
	await page.reload();
	await listo(page);
	await expect(page.locator('.panelito .s-t').nth(2)).toContainText('Reseñas');
	await expect(
		page.getByRole('group', { name: 'Color de acento' }).getByRole('button', { name: 'Color 4' })
	).toHaveAttribute('aria-pressed', 'true');
	// La combinación oscura quedó guardada en el estilo.
	await expect(page.locator('.color-fila', { hasText: 'Fondo' }).first()).toContainText('#111214');
	await page.getByRole('button', { name: 'Publicar cambios' }).click();
	await expect(page.getByText('Cambios publicados')).toBeVisible();
	await expect(page.getByRole('button', { name: 'Todo publicado' })).toBeVisible();
});
