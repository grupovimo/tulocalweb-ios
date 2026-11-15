import { mkdirSync } from 'node:fs';
import { test, type Page } from '@playwright/test';
import { PLANTILLAS } from '../src/lib/plantillas.ts';

/**
 * Capturas de las 17 vitrinas para revisar el diseño (workflow «Capturas», a mano). No es un test:
 * no afirma nada, sólo saca fotos de página entera a cada ancho y las deja en `capturas/`.
 *
 * - CAPTURAS_PRESETS: ids separados por coma (vacío = los 17).
 * - CAPTURAS_ANCHOS: anchos separados por coma (por defecto 375,1440,2560).
 * - CAPTURAS_VACIA: «1» suma la tienda casi vacía (una demo con un solo producto) en cada dirección.
 */
const slugDe = (n: string) =>
	n
		.normalize('NFD')
		.replace(/[̀-ͯ]/g, '')
		.toLowerCase()
		.replace(/&/g, 'y')
		.replace(/[^a-z0-9]+/g, '-')
		.replace(/^-|-$/g, '');

const lista = (s: string | undefined, def: string[]) =>
	s?.trim()
		? s
				.split(',')
				.map((x) => x.trim())
				.filter(Boolean)
		: def;

const PRESETS = lista(
	process.env.CAPTURAS_PRESETS,
	PLANTILLAS.map((p) => p.id)
);
const ANCHOS = lista(process.env.CAPTURAS_ANCHOS, ['375', '1440', '2560']).map(Number);
const VACIA = process.env.CAPTURAS_VACIA !== '0';
const DIR = 'capturas';
mkdirSync(DIR, { recursive: true });

test.describe.configure({ mode: 'parallel' });

/** Carga entera: baja hasta el fondo para que entren las fotos perezosas y cierra las ventanas. */
async function preparar(page: Page) {
	await page.locator('html[data-listo]').waitFor({ state: 'attached', timeout: 15_000 });
	await page.evaluate(async () => {
		const paso = innerHeight * 0.8;
		for (let y = 0; y < document.documentElement.scrollHeight; y += paso) {
			scrollTo(0, y);
			await new Promise((r) => setTimeout(r, 120));
		}
		scrollTo(0, 0);
	});
	await page.waitForLoadState('networkidle', { timeout: 10_000 }).catch(() => {});
	await page.evaluate(async () => {
		await Promise.all(
			[...document.images].map((i) =>
				i.complete ? null : new Promise((r) => ((i.onload = r), (i.onerror = r)))
			)
		);
		for (const d of document.querySelectorAll('dialog[open]')) (d as HTMLDialogElement).close();
	});
	await page.waitForTimeout(300);
}

async function foto(page: Page, nombre: string) {
	await preparar(page);
	await page.screenshot({
		path: `${DIR}/${nombre}.jpg`,
		fullPage: true,
		type: 'jpeg',
		quality: 72,
		animations: 'disabled'
	});
}

async function recorrer(page: Page, slug: string, id: string) {
	await page.goto(`/t/${slug}/catalogo`);
	await page.locator('html[data-listo]').waitFor({ state: 'attached' });
	const fichas = await page
		.locator(`a[href^="/t/${slug}/producto/"]`)
		.evaluateAll((as) => [...new Set(as.map((a) => a.getAttribute('href')!))]);
	// La primera ficha con opciones (talle, color, presentación) y la primera sin.
	let con: string | null = null;
	let sin: string | null = null;
	for (const f of fichas.slice(0, 16)) {
		if (con && sin) break;
		await page.goto(f);
		await page.locator('html[data-listo]').waitFor({ state: 'attached' });
		const tiene = (await page.locator('.tv-opcion').count()) > 0;
		if (tiene && !con) con = f;
		if (!tiene && !sin) sin = f;
	}
	const ids = fichas.slice(0, 2).map((h) => h.split('/').pop()!);
	for (const ancho of ANCHOS) {
		await page.setViewportSize({ width: ancho, height: ancho < 760 ? 812 : 900 });
		const n = (pag: string) => `${id}/${pag}-${ancho}`;
		await page.goto(`/t/${slug}`);
		await foto(page, n('1-inicio'));
		await page.goto(`/t/${slug}/catalogo`);
		await foto(page, n('2-catalogo'));
		if (sin) {
			await page.goto(sin);
			await foto(page, n('3-ficha-sin-variantes'));
		}
		if (con) {
			await page.goto(con);
			await foto(page, n('4-ficha-con-variantes'));
		}
		await page.goto(`/t/${slug}/nosotros`);
		await foto(page, n('7-nosotros'));
		await page.evaluate(
			({ s, ids }) =>
				localStorage.setItem(
					`carrito:${s}`,
					JSON.stringify(ids.map((x) => ({ id: x, varianteId: null, cantidad: 1 })))
				),
			{ s: slug, ids }
		);
		await page.goto(`/t/${slug}/carrito`);
		await foto(page, n('5-carrito'));
		await page.goto(`/t/${slug}/checkout`);
		await foto(page, n('6-checkout'));
	}
}

async function inicio(page: Page, slug: string) {
	await page.addInitScript((s) => {
		localStorage.setItem('tlw-cookies', 'nec');
		localStorage.setItem(`mayor18:${s}`, '1');
	}, slug);
}

for (const id of PRESETS) {
	const p = PLANTILLAS.find((x) => x.id === id);
	if (!p) continue;
	test(`capturas ${id}`, async ({ page }) => {
		test.setTimeout(600_000);
		mkdirSync(`${DIR}/${id}`, { recursive: true });
		const slug = slugDe(p.tienda);
		await inicio(page, slug);
		await recorrer(page, slug, id);
	});
}

/** Una demo recién abierta: un solo producto, para ver cómo se arma la tienda con casi nada. */
if (VACIA)
	for (const id of ['ceibo', 'litoral', 'mostaza']) {
		const p = PLANTILLAS.find((x) => x.id === id)!;
		test(`capturas casi vacía ${id}`, async ({ page }) => {
			test.setTimeout(600_000);
			mkdirSync(`${DIR}/vacia-${id}`, { recursive: true });
			await page.addInitScript(() => localStorage.setItem('tlw-cookies', 'nec'));
			await page.goto('/prueba');
			await page.locator('html[data-listo]').waitFor({ state: 'attached' });
			await page
				.getByRole('button', { name: new RegExp(`^${p.nombre}(?! Clásica)`) })
				.first()
				.click();
			await page.waitForURL(/\/panel$/);
			const slug = (await page.locator('.t-d').first().textContent())!.split('/t/')[1].trim();
			await inicio(page, slug);
			// Todos los productos menos uno, de a tandas del catálogo (que pagina).
			for (let vuelta = 0; vuelta < 10; vuelta++) {
				await page.goto(`/t/${slug}/catalogo`);
				const ids = await page
					.locator(`a[href^="/t/${slug}/producto/"]`)
					.evaluateAll((as) => [
						...new Set(as.map((a) => a.getAttribute('href')!.split('/').pop()!))
					]);
				if (ids.length <= 1) break;
				for (const x of ids.slice(1)) await page.request.delete(`/panel/productos/${x}`);
			}
			await recorrer(page, slug, `vacia-${id}`);
		});
	}
