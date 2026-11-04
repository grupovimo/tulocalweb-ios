import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });

async function demo(page: Page) {
	await page.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
	await page.goto('/prueba');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	await listo(page);
}

test('Tincho saluda, propone ideas y responde con los datos', async ({ page, isMobile }) => {
	await demo(page);
	if (isMobile)
		await page
			.getByRole('navigation', { name: 'Accesos' })
			.getByRole('button', { name: 'Tincho' })
			.click();
	else await page.getByRole('button', { name: 'Preguntale a Tincho' }).first().click();
	const panel = page.getByRole('dialog', { name: 'Tincho' });
	await expect(panel).toBeVisible();
	await expect(
		panel.getByText(/(Buen día|Buenas tardes|Buenas noches), Valeria\. Hoy vendiste/)
	).toBeVisible();
	await expect(panel.getByText('Recuperar 12 carritos')).toBeVisible();
	expect(
		await page.evaluate(
			() => document.documentElement.scrollWidth - document.documentElement.clientWidth
		)
	).toBeLessThanOrEqual(0);

	await panel.getByRole('button', { name: '¿Cuánto vendí ayer?' }).click();
	await expect(panel.locator('.globo.persona')).toHaveText('¿Cuánto vendí ayer?');
	await expect(panel.locator('.globo.hornero').last()).toContainText('Ayer vendiste');
	await expect(panel.locator('.globo.hornero b').last()).toContainText('$');

	// Una pregunta escrita: Tincho contesta con un globo nuevo (el motor sin IA guía, no repite el %).
	const antes = await panel.locator('.globo.hornero').count();
	await panel.getByRole('textbox', { name: 'Escribile a Tincho' }).fill('Subí precios 10%');
	await panel.getByRole('button', { name: 'Mandar' }).click();
	await expect(panel.locator('.globo.hornero')).toHaveCount(antes + 1);

	// La charla queda al cerrar y volver a abrir.
	await page.keyboard.press('Escape');
	await expect(panel).toBeHidden();
	await page.keyboard.press('ControlOrMeta+k');
	await expect(page.getByRole('textbox', { name: 'Buscar' })).toBeFocused();
	await page.keyboard.type('¿Quiénes no vuelven a comprar?');
	await page.keyboard.press('Enter');
	await expect(panel).toBeVisible();
	await expect(panel.locator('.globo.persona')).toHaveCount(3);
	await expect(panel.locator('.globo.hornero').last()).toContainText('clientes');
});

test('aprobar una idea queda hecho y se puede deshacer', async ({ page }) => {
	await demo(page);
	await page
		.locator('.hornero-card')
		.getByRole('button', { name: /Ver las/ })
		.click();
	const panel = page.getByRole('dialog', { name: 'Tincho' });
	const idea = panel.locator('.idea').filter({ hasText: 'Promo de fin de semana' });
	await idea.getByRole('button', { name: 'Aprobar' }).click();
	await expect(idea.getByText('Hecho')).toBeVisible();
	await expect(page.getByText('Campaña programada')).toBeVisible();
	await page.getByRole('button', { name: 'Deshacer' }).click();
	await expect(idea.getByRole('button', { name: 'Aprobar' })).toBeVisible();

	// Aprobar los carritos es lo mismo que "Sí, enviar" del Resumen.
	await panel
		.locator('.idea')
		.filter({ hasText: 'carritos' })
		.getByRole('button', { name: 'Aprobar' })
		.click();
	await page.keyboard.press('Escape');
	await expect(page.locator('.burbuja')).toContainText('Listo');
});

test('la API de Tincho valida lo que recibe', async ({ page }) => {
	await demo(page);
	const r = await page.request.post('/panel/hornero', {
		data: { mensajes: [{ rol: 'hornero', texto: 'hola' }] },
		headers: { origin: new URL(page.url()).origin }
	});
	expect(r.status()).toBe(400);
	const x = await page.request.post('/panel/hornero/ideas/inventada', {
		data: { hecha: true },
		headers: { origin: new URL(page.url()).origin }
	});
	expect(x.status()).toBe(404);
});
