import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });

/**
 * Nada sale a internet: gtag.js y fbevents.js se contestan con un script vacío (así los eventos
 * quedan en dataLayer y en la cola de fbq para revisarlos) y todo lo demás de afuera se corta.
 */
async function sinInternet(page: Page) {
	const pedidos = { gtag: 0, fbevents: 0 };
	await page.context().route(
		(u) => !['localhost', '127.0.0.1'].includes(u.hostname),
		(route) => {
			const u = new URL(route.request().url());
			if (u.hostname === 'www.googletagmanager.com' && u.pathname === '/gtag/js') {
				pedidos.gtag++;
				return route.fulfill({ contentType: 'text/javascript', body: '' });
			}
			if (u.hostname === 'connect.facebook.net' && u.pathname.endsWith('/fbevents.js')) {
				pedidos.fbevents++;
				return route.fulfill({ contentType: 'text/javascript', body: '' });
			}
			return route.abort();
		}
	);
	return pedidos;
}

async function demo(page: Page) {
	await page.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
	await page.goto('/prueba');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	return (await page.locator('.t-d').first().textContent())!.split('/t/')[1].trim();
}

/** Los eventos que la tienda le mandó a GA4 (dataLayer) y a Meta (cola de fbq). */
const eventos = (page: Page) =>
	page.evaluate(() => {
		const w = window as unknown as {
			dataLayer?: ArrayLike<unknown>[];
			fbq?: { queue: unknown[][] };
		};
		return {
			ga: (w.dataLayer ?? [])
				.map((a) => Array.from(a))
				.filter((a) => a[0] === 'event')
				.map((a) => ({ n: a[1] as string, p: a[2] as Record<string, unknown> })),
			meta: (w.fbq?.queue ?? [])
				// trackSingle: [, píxel, nombre, datos, opciones] (sólo al píxel de esta tienda).
				.filter((a) => a[0] === 'trackSingle')
				.map((a) => ({ n: a[2] as string, p: a[3], o: a[4] as { eventID?: string } | undefined }))
		};
	});

test('medición: IDs en Integraciones, sin consentimiento no carga nada y con él mide la compra', async ({
	page
}) => {
	const pedidos = await sinInternet(page);
	const slug = await demo(page);
	await page.goto('/panel/integraciones');
	await listo(page);

	// Google Analytics: valida el ID antes de guardarlo.
	const ga = page.locator('.tarjeta', { hasText: 'Google Analytics' });
	await ga.getByRole('button', { name: 'Conectar' }).click();
	const d = page.getByRole('dialog');
	await d.getByLabel('ID de medición de GA4').fill('UA-1234-1');
	await d.getByRole('button', { name: 'Guardar' }).click();
	await expect(d.getByRole('alert')).toContainText('GA4');
	await d.getByLabel('ID de medición de GA4').fill('g-test123456');
	await d.getByRole('button', { name: 'Guardar' }).click();
	await expect(ga.getByText('● Conectada')).toBeVisible();

	// Meta: el píxel y la URL del feed. La API de conversiones vive en Go: la demo no la ofrece.
	const meta = page.locator('.tarjeta', { hasText: 'Meta Ads' });
	await meta.getByRole('button', { name: 'Conectar' }).click();
	await expect(d).toContainText(`/t/${slug}/feed.xml`);
	await expect(d).not.toContainText('API de conversiones');
	await d.getByLabel('ID del píxel de Meta').fill('123456789012345');
	await d.getByRole('button', { name: 'Guardar' }).click();
	await expect(meta.getByText('● Conectada')).toBeVisible();

	// Sin elegir, la tienda muestra su aviso y no pide nada a Google ni a Meta.
	await page.goto(`/t/${slug}`);
	await listo(page);
	const aviso = page.getByRole('region', { name: 'Cookies de la tienda' });
	await expect(aviso).toContainText('Google Analytics y el píxel de Meta');
	await page.goto(`/t/${slug}/catalogo`);
	await listo(page);
	await aviso.getByRole('button', { name: 'Sólo necesarias' }).click();
	await expect(aviso).toBeHidden();
	await page.reload();
	await listo(page);
	await expect(aviso).toBeHidden();
	expect(pedidos).toEqual({ gtag: 0, fbevents: 0 });

	// Con consentimiento se cargan los dos y se mide la compra de punta a punta.
	await page.evaluate((s) => localStorage.removeItem(`mostrador-cookies:${s}`), slug);
	await page.reload();
	await listo(page);
	await aviso.getByRole('button', { name: 'Aceptar' }).click();
	await expect.poll(() => pedidos.gtag).toBe(1);
	await expect.poll(() => pedidos.fbevents).toBe(1);

	await page.goto(`/t/${slug}/producto/bombilla-alpaca`);
	await listo(page);
	await page.getByRole('button', { name: 'Agregar al carrito' }).click();
	await expect(page).toHaveURL(/\/carrito$/);
	await page.getByRole('link', { name: 'Finalizar compra' }).click();
	await page.getByLabel('Email').fill('lucia@ejemplo.com');
	await page.getByLabel('Nombre y apellido').fill('Lucía Pérez');
	await page.getByLabel('Teléfono').fill('11 5555-0101');
	await page.getByLabel('Código postal').fill('1414');
	await page.getByLabel(/Dirección/).fill('Gorriti 4500');
	await page.getByRole('radio', { name: /Retiro en el local/ }).click();
	await page.getByRole('radio', { name: /Transferencia/ }).click();
	await page.getByRole('button', { name: /^Pagar/ }).click();
	await expect(page.getByRole('heading', { name: '¡Gracias, Lucía!' })).toBeVisible();
	const numero = (await page
		.locator('b', { hasText: /^#\d+$/ })
		.first()
		.textContent())!.slice(1);

	await expect
		.poll(async () => (await eventos(page)).ga.some((e) => e.n === 'purchase'))
		.toBe(true);
	const e = await eventos(page);
	const nombres = e.ga.map((x) => x.n);
	for (const n of ['page_view', 'view_item', 'add_to_cart', 'begin_checkout', 'purchase'])
		expect(nombres, n).toContain(n);
	const compra = e.ga.find((x) => x.n === 'purchase')!.p;
	expect(compra).toMatchObject({ transaction_id: numero, currency: 'ARS' });
	expect(compra.value).toBeGreaterThan(0);
	expect((compra.items as unknown[]).length).toBe(1);
	const metas = e.meta.map((x) => x.n);
	for (const n of ['PageView', 'ViewContent', 'AddToCart', 'InitiateCheckout', 'Purchase'])
		expect(metas, n).toContain(n);
	expect(e.meta.find((x) => x.n === 'Purchase')!.o).toEqual({ eventID: numero });

	// Recargar el pedido no manda la compra otra vez.
	await page.reload();
	await listo(page);
	await expect
		.poll(async () => (await eventos(page)).ga.some((x) => x.n === 'page_view'))
		.toBe(true);
	expect((await eventos(page)).ga.some((x) => x.n === 'purchase')).toBe(false);
});

test('medición: la vitrina de una plantilla nunca mide', async ({ page }) => {
	const pedidos = await sinInternet(page);
	await page.goto('/t/tienda-lumbre');
	await listo(page);
	await expect(page.getByText(/tienda de ejemplo/)).toBeVisible();
	await expect(page.getByRole('region', { name: 'Cookies de la tienda' })).toHaveCount(0);
	expect(pedidos).toEqual({ gtag: 0, fbevents: 0 });
});

test('feed de productos: XML válido con los productos publicados', async ({ page }) => {
	await sinInternet(page);
	const slug = await demo(page);
	const r = await page.request.get(`/t/${slug}/feed.xml`);
	expect(r.status()).toBe(200);
	expect(r.headers()['content-type']).toContain('application/xml');
	const texto = await r.text();
	await page.goto(`/t/${slug}`);
	const f = await page.evaluate((x) => {
		const doc = new DOMParser().parseFromString(x, 'application/xml');
		const G = 'http://base.google.com/ns/1.0';
		const items = [...doc.querySelectorAll('item')];
		const de = (i: Element, n: string) => i.getElementsByTagNameNS(G, n)[0]?.textContent ?? '';
		return {
			error: doc.querySelector('parsererror')?.textContent ?? null,
			raiz: doc.documentElement.nodeName,
			items: items.map((i) => ({
				id: de(i, 'id'),
				link: de(i, 'link'),
				precio: de(i, 'price'),
				disponible: de(i, 'availability')
			}))
		};
	}, texto);
	expect(f.error).toBeNull();
	expect(f.raiz).toBe('rss');
	expect(f.items.length).toBeGreaterThan(0);
	for (const i of f.items) {
		expect(i.precio).toMatch(/^\d+\.\d{2} ARS$/);
		expect(['in_stock', 'out_of_stock']).toContain(i.disponible);
		expect(i.link).toContain(`/t/${slug}/producto/${i.id}`);
	}
	// Los productos del feed son los de la tienda online.
	await page.goto(`/t/${slug}/producto/${f.items[0].id}`);
	await expect(
		page.getByRole('button', { name: /Agregar al carrito|Sin stock|^Elegí/ })
	).toBeVisible();
});
