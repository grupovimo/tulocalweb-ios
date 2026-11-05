import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });

test('configuración: negocio, equipo, seguridad y plan', async ({ page, isMobile }) => {
	await page.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
	await page.goto('/prueba');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	await page.goto('/panel/configuracion');
	await listo(page);
	const nav = page.getByRole('navigation', { name: 'Secciones' });

	await page.getByLabel('Horario').fill('Lun a vie · 9 a 18 h');
	await page.getByRole('button', { name: 'Guardar cambios' }).click();
	await expect(page.getByText('Datos guardados')).toBeVisible();
	await page.reload();
	await listo(page);
	await expect(page.getByLabel('Horario')).toHaveValue('Lun a vie · 9 a 18 h');

	// El equipo se lista acá; se invita desde Usuarios y permisos (cuenta.e2e.ts).
	await nav.getByRole('button', { name: 'Equipo y permisos' }).click();
	await expect(page.getByText('Equipo · 4 personas')).toBeVisible();
	await expect(page.getByRole('link', { name: 'Invitar' })).toHaveAttribute(
		'href',
		'/panel/usuarios-y-permisos'
	);

	await nav.getByRole('button', { name: 'Seguridad' }).click();
	await page
		.locator('.sesion')
		.filter({ hasText: 'App de Tu Local Web en iPhone' })
		.getByRole('button', { name: 'Cerrar' })
		.click();
	await expect(page.locator('.sesion')).toHaveCount(2);
	await page.getByRole('button', { name: 'Cerrar todas las demás' }).click();
	await expect(page.locator('.sesion')).toHaveCount(1);
	await expect(page.locator('.sesion')).toContainText('Esta');

	// El plan se elige en Plan y facturación, con el resumen de lo que se paga (§56).
	await nav.getByRole('button', { name: 'Plan' }).click();
	await page.getByRole('link', { name: 'Elegir Emprende' }).click();
	await expect(page.getByRole('radio', { name: /Emprende/ })).toHaveAttribute(
		'aria-checked',
		'true'
	);
	await page.locator('#resumen').getByRole('button', { name: 'Seguir a Mercado Pago' }).click();
	await expect(page.getByText('Ya tenés el plan Emprende')).toBeVisible();
	if (!isMobile) await expect(page.locator('.prueba')).toContainText('Emprende');
});

test('configuración: la política de cambios llega a la tienda', async ({ page }) => {
	await page.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
	await page.goto('/prueba');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	const slug = (await page.locator('.t-d').first().textContent())!.split('/t/')[1].trim();
	await page.goto('/panel/configuracion');
	await listo(page);

	const politica = 'Cambios dentro de los 30 días con el ticket y la etiqueta puesta.';
	await page.getByLabel('Cambios y devoluciones').fill(politica);
	await page.getByRole('button', { name: 'Guardar cambios' }).click();
	await expect(page.getByText('Datos guardados')).toBeVisible();

	// En la ficha del producto va antes del texto legal del botón de arrepentimiento.
	await page.goto(`/t/${slug}/producto/bombilla-alpaca`);
	await listo(page);
	await page.getByRole('button', { name: 'Cambios y devoluciones' }).click();
	const texto = page.getByText(politica);
	await expect(texto).toBeVisible();
	await expect(texto).toContainText(`${politica} Tenés 10 días desde que lo recibís`);
	expect(
		await page.evaluate(() => document.documentElement.scrollWidth - innerWidth)
	).toBeLessThanOrEqual(0);
});
