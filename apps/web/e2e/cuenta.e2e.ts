import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });

async function demo(page: Page, ruta: string) {
	await page.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
	await page.goto('/prueba');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	await page.goto(ruta);
	await listo(page);
}

test('plan: elegir no cobra, el resumen muestra lo que se paga y la baja se programa sin perder lo pagado', async ({
	page
}) => {
	await demo(page, '/panel/plan-y-facturacion');
	// Pasar a anual propone el cambio en el resumen; nada se cobra hasta confirmar.
	const resumen = page.locator('#resumen');
	await expect(page.getByText('Elegí un plan para ver el detalle')).toBeVisible();
	await page.getByRole('group', { name: 'Ciclo' }).getByRole('button', { name: /Anual/ }).click();
	await expect(page.locator('.plan.cur')).toContainText('$ 37.425');
	await expect(resumen).toContainText('Pro anual');
	await expect(resumen).toContainText('$ 449.100 por año');
	await expect(page.locator('.c-s')).toContainText('plan Pro mensual');
	await page.getByRole('group', { name: 'Ciclo' }).getByRole('button', { name: 'Mensual' }).click();
	await expect(page.getByText('Elegí un plan para ver el detalle')).toBeVisible();

	// Elegir Negocio: el resumen dice qué se paga hoy (nada: está en la prueba) y desde cuándo.
	await page.getByRole('radio', { name: /Negocio/ }).click();
	await expect(resumen).toContainText('Total hoy');
	await expect(resumen).toContainText('no pagás nada hasta');
	await expect(resumen).toContainText('$ 149.900 por mes');
	await expect(resumen).toContainText('IVA incluido');
	await resumen.getByRole('button', { name: 'Seguir a Mercado Pago' }).click();
	await expect(page.getByText('Ya tenés el plan Negocio')).toBeVisible();
	await expect(page.locator('.c-s')).toContainText('plan Negocio mensual');

	await page.getByRole('button', { name: /\+2\.000 respuestas/ }).click();
	await expect(page.getByText('2.000 respuestas más para Tincho')).toBeVisible();

	// El descuento para quedarse (§57): «es caro» → 50 % en los próximos 3 cobros, una sola vez.
	await page.getByRole('button', { name: 'Cancelar plan' }).click();
	await page.getByRole('radio', { name: 'Es caro para lo que vendo' }).click();
	await page.getByRole('button', { name: 'Continuar' }).click();
	await expect(page.getByText('Te dejamos 3 cobros a mitad de precio')).toBeVisible();
	await page.getByRole('button', { name: 'Quiero el descuento' }).click();
	await expect(page.getByText('¡Gracias por quedarte!')).toBeVisible();
	await expect(page.getByText('50 % off en tus próximos 3 cobros', { exact: true })).toBeVisible();
	await page.getByRole('button', { name: 'Cancelar plan' }).click();
	await page.getByRole('button', { name: 'Continuar' }).click();
	await expect(page.getByText('Contanos qué te falta')).toBeVisible();
	await page.getByRole('button', { name: 'Cancelar igual' }).click();
	await page.getByRole('button', { name: 'Volver' }).click();

	// Cancelar: el motivo, la oferta y la baja, que rige cuando termina lo pagado.
	await page.getByRole('button', { name: 'Cancelar plan' }).click();
	await page.getByRole('radio', { name: 'Me voy a otra plataforma' }).click();
	await page.getByRole('button', { name: 'Continuar' }).click();
	await expect(page.getByText('Antes de irte, hablemos')).toBeVisible();
	await page.getByRole('button', { name: 'Cancelar igual' }).click();
	await page.getByRole('button', { name: 'Cancelar mi plan' }).click();
	await expect(page.getByText('Plan cancelado')).toBeVisible();
	await expect(page.getByText(/Pasás al plan Gratis el/)).toBeVisible();
	await expect(page.getByText('Tu plan está cancelado')).toBeVisible();

	// Seguir con el plan deshace la baja desde el mismo resumen.
	await page.getByRole('button', { name: 'Seguir con Negocio' }).first().click();
	await resumen.getByRole('button', { name: 'Deshacer el cambio' }).click();
	await expect(page.getByText('Seguís con Negocio', { exact: true })).toBeVisible();
	await expect(page.getByText('¿Necesitás frenar?')).toBeVisible();
});

test('usuarios: invitar, rol fijo y quitar', async ({ page }) => {
	await demo(page, '/panel/usuarios-y-permisos');
	const martin = page.locator('.persona').filter({ hasText: 'Martín Ríos' });
	await martin.getByRole('group').getByRole('button', { name: 'Local' }).click();
	await expect(page.getByText('Martín ahora es Local')).toBeVisible();
	await expect(martin).toContainText('Punto de venta');
	// Roles fijos: la tarjeta explica qué puede cada uno (sin matriz para tocar).
	await expect(page.getByText('Qué puede cada rol')).toBeVisible();

	await page.getByRole('button', { name: 'Invitar persona' }).click();
	await page.getByLabel('Email').fill('juli@mail.com');
	await page
		.getByRole('group', { name: 'Rol', exact: true })
		.getByRole('button', { name: /^Administración/ })
		.click();
	await page.getByRole('button', { name: 'Mandar invitación' }).click();
	const juli = page.locator('.persona').filter({ hasText: 'juli@mail.com' });
	await expect(juli).toContainText('Invitación enviada');
	await juli.getByRole('button', { name: 'Reenviar' }).click();
	await expect(page.getByText('Invitación reenviada')).toBeVisible();
	await juli.getByRole('button', { name: 'Cancelar la invitación a juli@mail.com' }).click();
	await expect(page.locator('.persona').filter({ hasText: 'juli@mail.com' })).toHaveCount(0);

	await page.getByRole('button', { name: 'Quitar a Carla Díaz' }).click();
	await expect(page.getByText('Carla Díaz ya no tiene acceso')).toBeVisible();
	await expect(page.locator('.persona')).toHaveCount(3);
});
