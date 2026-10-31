import { expect, test, type Page } from '@playwright/test';
import { PLANTILLAS } from '../src/lib/plantillas.ts';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });
const sinDesborde = (page: Page) =>
	page.evaluate(() => document.documentElement.scrollWidth - innerWidth);

/** Una demo de la plantilla (por su nombre en /prueba); devuelve el slug de su tienda. */
async function demo(page: Page, nombre: string) {
	await page.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
	await page.goto('/prueba');
	await listo(page);
	// "Miga" no es "Miga Clásica".
	await page
		.getByRole('button', { name: new RegExp(`^${nombre}(?! Clásica)`) })
		.first()
		.click();
	await expect(page).toHaveURL(/\/panel$/);
	return (await page.locator('.t-d').first().textContent())!.split('/t/')[1].trim();
}

/** Malbec pregunta la edad al hidratar: se espera el cartel (si aparece después, tapa el clic). */
async function mayor(page: Page) {
	await page.getByRole('button', { name: 'Sí, soy mayor' }).click();
}

test('suscripciones: el club de vinos de Malbec, el alta sin cobrar en la demo y el link privado', async ({
	page
}) => {
	const slug = await demo(page, 'Malbec');

	// El panel: los planes del club y quienes ya se suscribieron.
	await page.goto('/panel/suscripciones');
	await listo(page);
	await expect(page.getByRole('heading', { name: 'Suscripciones' })).toBeVisible();
	await expect(page.getByText('Explorador', { exact: true }).first()).toBeVisible();
	await expect(page.getByText('Gustavo Rinaldi')).toBeVisible();
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);

	// La tienda: la sección del club con los planes reales, y la página del plan.
	await page.goto(`/t/${slug}`);
	await listo(page);
	await mayor(page);
	await expect(page.getByText('Club de vinos').first()).toBeVisible();
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	await page.locator(`a[href="/t/${slug}/suscripciones/club-explorador"]`).first().click();
	await expect(page).toHaveURL(/\/suscripciones\/club-explorador$/);
	await listo(page);
	await expect(page.getByText('Es la demo: no se cobra nada')).toBeVisible();
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);

	await page.getByLabel('Email').fill('julia@ejemplo.com');
	await page.getByLabel('Nombre y apellido').fill('Julia Paz');
	await page.getByRole('textbox', { name: 'WhatsApp' }).fill('11 5555-0101');
	await page.getByLabel(/Dirección/).fill('Gorriti 4500');
	await page.getByRole('button', { name: 'Sumarme al club' }).click();

	// El link privado: activa, sin cobro en la demo; se pausa, se reanuda y se da de baja.
	await expect(page).toHaveURL(/\/suscripcion\/[A-Za-z0-9_-]{16,}\?alta=1$/);
	await expect(page.getByRole('heading', { name: '¡Listo, Julia!' })).toBeVisible();
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	await page.getByRole('button', { name: 'Pausar' }).click();
	await expect(page.getByRole('heading', { name: 'Suscripción pausada' })).toBeVisible();
	await page.getByRole('button', { name: 'Reanudar' }).click();
	await expect(page.getByRole('button', { name: 'Pausar' })).toBeVisible();
	await page.getByRole('button', { name: 'Dar de baja' }).click();
	await page.getByRole('button', { name: 'Sí, darla de baja' }).click();
	await expect(page.getByRole('heading', { name: 'Suscripción cancelada' })).toBeVisible();
	await expect(page.getByRole('link', { name: 'Volver a sumarme' })).toBeVisible();

	// En el panel se ve la baja.
	await page.goto('/panel/suscripciones');
	await listo(page);
	await page.getByRole('button', { name: 'Todas' }).click();
	await expect(page.locator('.fila', { hasText: 'Julia Paz' })).toContainText('Cancelada');
});

test('suscripciones: el panel crea un plan, lo pausa y cancela un suscriptor', async ({ page }) => {
	await demo(page, 'Malbec');
	await page.goto('/panel/suscripciones');
	await listo(page);
	await page.getByRole('button', { name: 'Nuevo plan' }).click();
	await page.getByLabel('Nombre', { exact: true }).fill('Caja de espumantes');
	await page.getByLabel('Precio por período').fill('52000');
	await page.getByLabel('Se cobra').selectOption('bimestral');
	await page.getByLabel('Contalo con tus palabras').fill('3 espumantes elegidos');
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	await page.getByRole('button', { name: 'Guardar' }).click();
	await expect(page.locator('.fila', { hasText: 'Caja de espumantes' })).toContainText(
		'cada 2 meses'
	);
	await page.getByRole('button', { name: 'Pausar Caja de espumantes' }).click();
	await expect(page.locator('.fila', { hasText: 'Caja de espumantes' })).toContainText('Pausado');

	await page.getByRole('button', { name: 'Cancelar la suscripción de Sole Ferreyra' }).click();
	await page.getByRole('button', { name: 'Cancelar suscripción' }).click();
	await expect(page.getByText('Suscripción cancelada', { exact: true })).toBeVisible();
	await expect(page.locator('.fila', { hasText: 'Sole Ferreyra' })).toHaveCount(0);
});

/**
 * Suscripciones en cada plantilla: en una demo se crea un plan; la portada lo muestra, la página del
 * plan y el link privado se dibujan sin errores ni desborde. Sin planes, la portada no lo nombra.
 */
const PLANTILLAS_E2E = [
	'andes',
	'quebracho',
	'miga',
	'miga-clasica',
	'tijera',
	'pampa',
	'rosedal',
	'malbec',
	'mostaza',
	'yerbal',
	'ceibo',
	'patio'
];
for (const id of PLANTILLAS_E2E) {
	const p = PLANTILLAS.find((x) => x.id === id)!;
	test(`suscripciones en ${id}`, async ({ page }) => {
		const errores: string[] = [];
		page.on('pageerror', (e) => errores.push(e.message));
		const slug = await demo(page, p.nombre);
		if (id !== 'malbec' && id !== 'mostaza') {
			await page.goto(`/t/${slug}`);
			await listo(page);
			await expect(page.locator(`a[href*="/t/${slug}/suscripciones/"]`)).toHaveCount(0);
		}
		const r = await page.request.post('/panel/suscripciones/planes', {
			data: {
				nombre: 'Caja del mes',
				descripcion: 'Lo mejor de la temporada, elegido por nosotros.',
				foto: null,
				precio: 25000,
				frecuencia: 'mensual',
				productos: [],
				incluye: 'Una selección de la casa',
				envio: true,
				retiro: true,
				cupo: null,
				activo: true
			}
		});
		expect(r.status()).toBe(201);
		const plan = (await r.json()) as { id: string };
		await page.goto(`/t/${slug}`);
		await listo(page);
		if (id === 'malbec') await mayor(page);
		// Mostaza lleva desde la tarjeta «Para tu casa» de su portada al primer plan (desde ahí se elige).
		await expect(
			page
				.locator(
					id === 'mostaza'
						? `a[href*="/t/${slug}/suscripciones/"]`
						: `a[href="/t/${slug}/suscripciones/${plan.id}"]`
				)
				.first()
		).toBeVisible();
		expect(await sinDesborde(page), `${id} portada`).toBeLessThanOrEqual(0);

		await page.goto(`/t/${slug}/suscripciones/${plan.id}`);
		await listo(page);
		await expect(page.getByText('Caja del mes').first()).toBeVisible();
		await expect(page.getByText('Es la demo: no se cobra nada')).toBeVisible();
		expect(await sinDesborde(page), `${id} plan`).toBeLessThanOrEqual(0);
		// El alta (el pie puede tener otro formulario con email: el de la newsletter).
		const alta = page.locator('form:has(input[name=entrega])');
		await alta.locator('input[name=email]').fill('ana@ejemplo.com');
		await alta.locator('input[name=nombre]').fill('Ana Gómez');
		await alta.locator('input[name=telefono]').fill('11 5555-0102');
		await alta.getByRole('radio', { name: /Retir/ }).click();
		await alta.locator('button:not([type=button])').last().click();
		await expect(page).toHaveURL(/\/suscripcion\/[A-Za-z0-9_-]{16,}\?alta=1$/);
		await expect(page.getByRole('heading', { name: '¡Listo, Ana!' })).toBeVisible();
		expect(await sinDesborde(page), `${id} link privado`).toBeLessThanOrEqual(0);
		expect(errores).toEqual([]);
	});
}
