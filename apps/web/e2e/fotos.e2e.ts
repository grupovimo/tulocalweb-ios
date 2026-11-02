import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });

// Un PNG de 64 × 48 de un solo color (el navegador lo pasa a WebP antes de subirlo).
const PNG = Buffer.from(
	'iVBORw0KGgoAAAANSUhEUgAAAEAAAAAwCAIAAAAuKetIAAAAZElEQVR4nO3PUQkAIBTAwNfIOuY2kSH8OITBAtzm7PV1wwUNaEEDWtCAFjSgBQ1oQQNa0IAWNKAFDWhBA1rQgBY0oAUNaEEDWtCAFjSgBQ1oQQNa0IAWNKAFDWhBA1rQgBY8dgEOyyzx7DtpsgAAAABJRU5ErkJggg==',
	'base64'
);

test('fotos: se suben desde el editor, quedan en la galería y se sirven desde /f/', async ({
	page
}) => {
	await page.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
	await page.goto('/prueba');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	await page.goto('/panel/productos');
	await listo(page);
	await page.getByRole('button', { name: 'Nuevo producto' }).click();
	// Primero pregunta qué es.
	await page.getByRole('dialog').getByRole('button', { name: /^Otra cosa/ }).click();
	const ed = page.getByRole('dialog');
	await ed.getByLabel('Nombre', { exact: true }).fill('Mate con foto');

	const subida = page.waitForResponse((r) => r.url().endsWith('/panel/fotos'));
	await ed
		.getByLabel('Elegir fotos')
		.setInputFiles({ name: 'mate.png', mimeType: 'image/png', buffer: PNG });
	const r = await subida;
	expect(r.status()).toBe(201);
	const { ruta } = (await r.json()) as { ruta: string };
	expect(ruta).toMatch(/^\/f\/fotos\/.+\.webp$/);
	const url = new URL(ruta, page.url()).href;
	await expect(ed.locator(`img[src="${url}"]`)).toBeVisible();

	// La foto se sirve con caché larga y es una imagen de verdad.
	const f = await page.request.get(url);
	expect(f.status()).toBe(200);
	expect(f.headers()['content-type']).toBe('image/webp');
	expect(f.headers()['cache-control']).toContain('immutable');

	await ed.getByRole('button', { name: /Guardar/ }).click();
	await expect(ed).toBeHidden();
	await page.locator('.fila', { hasText: 'Mate con foto' }).locator('.abrir').click();
	await expect(page.getByRole('dialog').locator(`img[src="${url}"]`)).toBeVisible();
	await page.getByRole('dialog').getByRole('button', { name: 'Sacar la foto 1' }).click();
	await expect(page.getByRole('dialog').getByText('Sumá la primera foto')).toBeVisible();

	// Lo que no es una foto, no entra.
	const mal = await page.request.post('/panel/fotos', {
		headers: { 'content-type': 'application/octet-stream' },
		data: '<svg xmlns="http://www.w3.org/2000/svg"></svg>'
	});
	expect(mal.status()).toBe(415);
});
