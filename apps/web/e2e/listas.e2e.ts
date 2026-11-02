import { expect, test, type Page } from '@playwright/test';

const listo = (page: Page) => page.locator('html[data-listo]').waitFor({ state: 'attached' });
const sinDesborde = (page: Page) =>
	page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);

async function demo(page: Page, ruta: string) {
	await page.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
	await page.goto('/prueba');
	await listo(page);
	await page.getByRole('button', { name: /^Litoral/ }).click();
	await expect(page).toHaveURL(/\/panel$/);
	await page.goto(ruta);
	await listo(page);
}

test('clientes: segmentos en la URL, ficha y cupón', async ({ page }) => {
	await demo(page, '/panel/clientes');
	await expect(page.getByRole('heading', { name: 'Clientes' })).toBeVisible();
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	await page.getByRole('group', { name: 'Segmentos' }).getByRole('button', { name: /VIP/ }).click();
	await expect(page).toHaveURL(/seg=vip/);
	await expect(page.locator('.chip-f, .fila').first()).toBeVisible();

	await page.locator('.fila .abrir').first().click();
	const ficha = page.getByRole('dialog');
	await expect(ficha.getByText('Compras de los últimos 12 meses')).toBeVisible();
	await ficha.getByRole('button', { name: /Mandar cupón/ }).click();
	await expect(page.getByText(/Cupón del 10% para/)).toBeVisible();
	await expect(ficha.getByRole('button', { name: 'Cupón enviado' })).toBeDisabled();
});

test('clientes: alta y etiqueta', async ({ page }) => {
	await demo(page, '/panel/clientes');
	await page.getByRole('button', { name: 'Nuevo cliente' }).click();
	const alta = page.getByRole('dialog');
	await alta.getByLabel('Nombre y apellido').fill('Rosa Pérez Prueba');
	await alta.getByLabel('WhatsApp').fill('11 5555-1234');
	await alta.getByRole('button', { name: 'Guardar cliente' }).click();
	await expect(page.getByText('Rosa Pérez Prueba ya es cliente')).toBeVisible();
	await page.getByRole('textbox', { name: 'Buscar clientes' }).fill('Rosa Pérez Prueba');
	await expect(page.locator('.fila')).toHaveCount(1);
	await page.locator('.fila .tilde').click();
	await page.locator('.lote').getByRole('button', { name: 'Etiquetar' }).click();
	await page.getByRole('dialog').getByLabel('Etiqueta').fill('Feria');
	await page.getByRole('dialog').getByRole('button', { name: 'Etiquetar' }).click();
	await expect(page.getByText('Etiqueta «Feria»')).toBeVisible();
	await page.getByRole('textbox', { name: 'Buscar clientes' }).fill('Feria');
	await expect(page.locator('.fila')).toHaveCount(1);
});

test('productos: pausar, editar, crear y exportar', async ({ page }) => {
	await demo(page, '/panel/productos');
	// La invitación a cargar con foto es sólo para la lista vacía; la demo ya tiene productos.
	await expect(page.locator('.fila').first()).toBeVisible();
	await expect(page.getByText('Cargá productos con sus fotos')).toHaveCount(0);
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);

	const primero = page.locator('.fila').first();
	const nombre = (await primero.locator('.q-t').textContent())!.trim();
	await primero.getByRole('switch').click();
	await expect(page.getByText(`${nombre} pausado`)).toBeVisible();
	await page
		.getByRole('group', { name: 'Segmentos' })
		.getByRole('button', { name: /Pausados/ })
		.click();
	await expect(page.locator('.fila').filter({ hasText: nombre })).toBeVisible();

	await page.locator('.fila .abrir').first().click();
	const editor = page.getByRole('dialog', { name: 'Editar producto' });
	await editor.getByLabel('Nombre').fill(`${nombre} edición especial`);
	await editor.getByRole('button', { name: '✦ Mejorar con Tincho' }).click();
	// En la demo, sin modelo, la descripción arranca con el nombre (sin promesas inventadas).
	await expect(editor.locator('#desc-prod')).toHaveValue(
		new RegExp(`^${nombre} edición especial\\.`)
	);
	await editor.getByRole('button', { name: 'Guardar y publicar' }).click();
	await expect(page.getByText('Producto guardado')).toBeVisible();

	await page.getByRole('button', { name: 'Nuevo producto' }).click();
	const nuevo = page.getByRole('dialog', { name: 'Nuevo producto' });
	// Primero pregunta qué es.
	await nuevo.getByRole('button', { name: /^Otra cosa/ }).click();
	await nuevo.getByLabel('Nombre').fill('Termo de prueba');
	await nuevo.getByLabel('Precio base').fill('45000');
	await nuevo.getByLabel('Precio base').blur();
	await nuevo.getByRole('button', { name: 'Guardar y publicar' }).click();
	await expect(page.getByText('Producto creado')).toBeVisible();
	await page.goto('/panel/productos?q=Termo%20de%20prueba');
	await listo(page);
	await expect(page.locator('.fila')).toHaveCount(1);
	await expect(page.locator('.fila')).toContainText('$ 45.000');

	const csv = await (await page.request.get('/panel/listas/productos/exportar')).text();
	expect(csv.split('\r\n')[0]).toContain('Nombre;SKU;Categoría;Precio');
	expect(csv).toContain('Termo de prueba');
});

test('productos: compartir uno publicado; pausado pide publicarlo', async ({ page }, info) => {
	// La hoja del sistema, de mentira: registra lo que se compartió.
	await page.addInitScript(() => {
		const w = window as unknown as { compartido: ShareData[] };
		w.compartido = [];
		Object.defineProperty(navigator, 'share', {
			value: async (d: ShareData) => void w.compartido.push(d)
		});
		Object.defineProperty(navigator, 'canShare', { value: () => true });
	});
	await demo(page, '/panel/productos');
	const fila = page
		.locator('.fila')
		.filter({ has: page.locator('[role=switch][aria-checked=true]') })
		.first();
	const nombre = (await fila.locator('.q-t').textContent())!.trim();
	const esa = page.locator('.fila').filter({ hasText: nombre }).first();
	await esa.getByRole('button', { name: `Compartir: ${nombre}` }).click();
	if (info.project.name === 'mobile') {
		await expect
			.poll(() =>
				page.evaluate(() => (window as unknown as { compartido: ShareData[] }).compartido)
			)
			.toHaveLength(1);
		const [d] = await page.evaluate(
			() => (window as unknown as { compartido: ShareData[] }).compartido
		);
		expect(d.text).toMatch(
			new RegExp(
				`^${nombre.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')} · \\$ [\\d.]+ — https://tulocalweb\\.com\\.ar/t/[^/]+/producto/`
			)
		);
	} else {
		await expect(page.getByText(/link del producto/i)).toBeVisible();
		await expect(page.getByRole('button', { name: 'Enviar por WhatsApp' })).toBeVisible();
	}

	// Pausado: avisa que primero hay que publicarlo, con el botón para hacerlo.
	await esa.getByRole('switch').click();
	await expect(page.getByText(`${nombre} pausado`)).toBeVisible();
	await esa.getByRole('button', { name: `Compartir: ${nombre}` }).click();
	await expect(page.getByText('Primero publicalo')).toBeVisible();
	await page.getByRole('button', { name: 'Publicar', exact: true }).click();
	await expect(page.getByText(`${nombre} publicado`)).toBeVisible();
});

test('productos: importar un CSV', async ({ page }) => {
	await demo(page, '/panel/productos');
	await page.locator('input[type=file]').setInputFiles({
		name: 'productos.csv',
		mimeType: 'text/csv',
		buffer: Buffer.from(
			'nombre;precio;stock;categoria\nBombilla importada;1.700;12;Mates\nMatera importada;36500;3;Mates\n'
		)
	});
	await expect(page.getByText('2 productos importados')).toBeVisible();
	await page.goto('/panel/productos?q=importada');
	await listo(page);
	await expect(page.locator('.fila')).toHaveCount(2);
});

test('productos: importar tienda pide un link válido', async ({ page }) => {
	await demo(page, '/panel/productos');
	await page.getByRole('button', { name: 'Importar tienda' }).click();
	const importar = page.getByRole('button', { name: 'Importar', exact: true });
	await expect(importar).toBeDisabled();
	await page.getByRole('button', { name: 'Empretienda' }).click();
	await expect(page.getByLabel('Link de tu tienda')).toHaveAttribute(
		'placeholder',
		'mitienda.empretienda.com.ar'
	);
	await page.getByLabel('Link de tu tienda').fill('no existe nada');
	await importar.click();
	await expect(page.getByRole('alert')).toContainText('mitienda.empretienda.com.ar');
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
});

test('productos: importar tienda de Shopify va por la cola', async ({ page }) => {
	// Nada sale a la red: la plataforma y la cola se responden acá.
	const pedidos: { accion: string }[] = [];
	await page.route('**/panel/productos/tienda', (r) =>
		r.fulfill({
			json: {
				fuente: 'shopify',
				origen: 'https://mitienda.myshopify.com',
				total: 7,
				paginas: 2,
				omitidas: 0
			}
		})
	);
	const cola = (estado: string, productos: number) => ({
		id: 'imp1',
		plataforma: 'shopify',
		estado,
		etapa: 'productos',
		progreso: { productos, clientes: 0, pedidos: 0, redirecciones: 0, repetidos: 1, recortados: 1 },
		errores: []
	});
	await page.route('**/panel/importaciones', (r) => {
		if (r.request().method() === 'GET')
			return r.fulfill({ json: { importacion: null, tiendanube: true } });
		const b = r.request().postDataJSON() as { accion: string };
		pedidos.push(b);
		const avances = pedidos.filter((x) => x.accion === 'avanzar').length;
		return r.fulfill({
			json:
				b.accion === 'crear'
					? cola('pendiente', 0)
					: avances < 2
						? cola('corriendo', 5)
						: cola('lista', 6)
		});
	});
	await demo(page, '/panel/productos');
	await page.getByRole('button', { name: 'Importar tienda' }).click();
	await page.getByRole('button', { name: 'WooCommerce' }).click();
	await page.getByLabel('Link de tu tienda').fill('mitienda.myshopify.com');
	await page.getByRole('button', { name: 'Importar', exact: true }).click();
	// Si el link era de otra plataforma, se usa la que respondió.
	await expect(page.getByRole('status')).toContainText('Tu tienda de Shopify ya está');
	await expect(page.getByText('Productos 6')).toBeVisible();
	await expect(page.getByText('1 ya estaban · 1 recortados')).toBeVisible();
	await expect(page.getByText('Shopify no publica el stock')).toBeVisible();
	expect(pedidos[0]).toEqual({
		accion: 'crear',
		plataforma: 'shopify',
		url: 'https://mitienda.myshopify.com'
	});
	expect(await sinDesborde(page)).toBeLessThanOrEqual(0);
	await page.getByRole('button', { name: 'Listo' }).click();
});
