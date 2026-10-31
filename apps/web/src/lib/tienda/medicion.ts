/**
 * Medición de la tienda online con las cuentas de la propia tienda: Google Analytics 4 (gtag.js) y
 * el píxel de Meta (fbevents.js). Nada se carga si la tienda no pegó sus IDs, si es la vitrina de
 * una plantilla o si quien compra no aceptó las cookies de medición (AvisoCookies.svelte). La
 * elección se recuerda por tienda en localStorage.
 *
 * Las plantillas no llaman a nada de acá: el layout de /t/[tienda] manda las páginas vistas, las
 * rutas de producto, checkout y pedido mandan lo suyo, y el carrito compartido avisa lo que se suma.
 */
import type { ProductoPublico, TiendaPublica } from './tipos.ts';

export type Medicion = TiendaPublica['medicion'];

// ---------------------------------------------------------------- Validación (servidor y panel)

export const FORMATO_GA4 = /^G-[A-Z0-9]{6,12}$/;
export const FORMATO_PIXEL = /^\d{15,16}$/;

type Validado = { ok: true; valor: string | null } | { ok: false; error: string };

/** El ID de medición de GA4 ("G-XXXXXXXXXX"). Vacío es válido: borra el ID. */
export function validarGa4(v: string | null | undefined): Validado {
	const s = (v ?? '').trim().toUpperCase();
	if (!s) return { ok: true, valor: null };
	if (s.startsWith('UA-'))
		return { ok: false, error: 'Ese es un ID de Universal Analytics: usá el de GA4 (G-…).' };
	if (!FORMATO_GA4.test(s))
		return { ok: false, error: 'El ID de medición empieza con G- (por ejemplo, G-AB12CD34EF).' };
	return { ok: true, valor: s };
}

/** El ID del píxel de Meta: sólo números, 15 o 16. Vacío es válido: borra el ID. */
export function validarPixel(v: string | null | undefined): Validado {
	const s = (v ?? '').trim();
	if (!s) return { ok: true, valor: null };
	if (!FORMATO_PIXEL.test(s))
		return { ok: false, error: 'El ID del píxel son 15 o 16 números, sin letras ni espacios.' };
	return { ok: true, valor: s };
}

// ---------------------------------------------------------------- Consentimiento

const claveCookies = (slug: string) => `mostrador-cookies:${slug}`;
const claveCompra = (slug: string) => `mostrador-compra:${slug}`;
const claveMedidas = (slug: string) => `mostrador-compras-medidas:${slug}`;

/** Si la tienda pide el aviso de cookies: tiene algún ID y no es una vitrina. */
export function pideConsentimiento(t: Pick<TiendaPublica, 'vitrina' | 'medicion'>) {
	return !t.vitrina && !!(t.medicion?.ga4 || t.medicion?.pixel);
}

/**
 * La cookie que lee el checkout del servidor: con 'si', la compra va además a la API de conversiones
 * de Meta (§47). Refleja lo que quedó en localStorage; sin elección, no hay cookie.
 */
export const cookieMedir = (slug: string) => `tlw-medir-${slug}`;

function marcarCookie(slug: string, si: boolean) {
	try {
		document.cookie = `${cookieMedir(slug)}=${si ? 'si' : 'no'}; path=/; max-age=31536000; samesite=lax; secure`;
	} catch {
		// Sin cookies: la compra se mide sólo desde el navegador.
	}
}

/** true: aceptó; false: sólo necesarias; null: todavía no eligió. */
export function leerConsentimiento(slug: string): boolean | null {
	try {
		const v = localStorage.getItem(claveCookies(slug));
		return v === 'si' ? true : v === 'no' ? false : null;
	} catch {
		return null;
	}
}

/** Guarda la elección y, si aceptó, carga la medición en el momento (con la página actual). */
export function guardarConsentimiento(slug: string, si: boolean) {
	try {
		localStorage.setItem(claveCookies(slug), si ? 'si' : 'no');
	} catch {
		// Sin almacenamiento (modo privado): vale para esta visita.
		sinAlmacenamiento.set(slug, si);
	}
	marcarCookie(slug, si);
	if (si && activa?.slug === slug) {
		cargar(activa);
		pagina();
	}
}

const sinAlmacenamiento = new Map<string, boolean>();
const acepto = (slug: string) => leerConsentimiento(slug) ?? sinAlmacenamiento.get(slug) ?? false;

// ---------------------------------------------------------------- Eventos (sin navegador)

export type Linea = { id: string; cantidad: number };
type ItemGa = {
	item_id: string;
	item_name: string;
	item_category?: string;
	price: number;
	quantity: number;
};

export function itemsGa(lineas: Linea[], productos: ProductoPublico[]): ItemGa[] {
	const porId = new Map(productos.map((p) => [p.id, p]));
	return lineas.flatMap((l) => {
		const p = porId.get(l.id);
		return p
			? [
					{
						item_id: p.id,
						item_name: p.nombre,
						...(p.categoria ? { item_category: p.categoria } : {}),
						price: p.precio,
						quantity: l.cantidad
					}
				]
			: [];
	});
}

/** Lo mismo para Meta (contents + content_ids). */
export function contenidoMeta(items: ItemGa[], value: number) {
	return {
		content_ids: items.map((i) => i.item_id),
		contents: items.map((i) => ({ id: i.item_id, quantity: i.quantity, item_price: i.price })),
		content_type: 'product',
		num_items: items.reduce((s, i) => s + i.quantity, 0),
		value,
		currency: 'ARS'
	};
}

const valor = (items: ItemGa[]) => items.reduce((s, i) => s + i.price * i.quantity, 0);

// ---------------------------------------------------------------- En el navegador

type Ventana = Window & {
	dataLayer?: unknown[];
	gtag?: (...a: unknown[]) => void;
	fbq?: ((...a: unknown[]) => void) & Record<string, unknown>;
	_fbq?: unknown;
};

type Activa = { slug: string; medicion: Medicion; productos: ProductoPublico[] };
let activa: Activa | null = null;
const cargados = new Set<string>();

/** Fija la tienda que se está mirando. Sin IDs o en una vitrina, no queda nada activo. */
export function preparar(t: TiendaPublica) {
	if (typeof window === 'undefined') return;
	activa = pideConsentimiento(t)
		? { slug: t.slug, medicion: t.medicion, productos: t.productos }
		: null;
	if (!activa) return;
	// Quien eligió antes de que existiera la cookie (o la borró): se vuelve a marcar.
	const eligio = leerConsentimiento(activa.slug);
	if (eligio !== null) marcarCookie(activa.slug, eligio);
	if (acepto(activa.slug)) cargar(activa);
}

function script(src: string) {
	const s = document.createElement('script');
	s.async = true;
	s.src = src;
	document.head.appendChild(s);
}

function cargar(a: Activa) {
	const w = window as Ventana;
	const { ga4, pixel } = a.medicion;
	if (ga4 && !cargados.has(ga4)) {
		cargados.add(ga4);
		w.dataLayer = w.dataLayer || [];
		// gtag.js lee `arguments`, no un array: por eso no es una función flecha.
		w.gtag ??= function () {
			// eslint-disable-next-line prefer-rest-params
			w.dataLayer!.push(arguments);
		};
		w.gtag('js', new Date());
		// Las páginas vistas las manda pagina() en cada navegación (es una SPA).
		w.gtag('config', ga4, { send_page_view: false });
		script(`https://www.googletagmanager.com/gtag/js?id=${encodeURIComponent(ga4)}`);
	}
	if (pixel && !cargados.has(pixel)) {
		cargados.add(pixel);
		if (!w.fbq) {
			// El arranque oficial de fbevents.js: una cola hasta que carga el script.
			const n = function (...a: unknown[]) {
				if (n.callMethod) (n.callMethod as (...x: unknown[]) => void)(...a);
				else (n.queue as unknown[]).push(a);
			} as NonNullable<Ventana['fbq']>;
			n.push = n;
			n.loaded = true;
			n.version = '2.0';
			n.queue = [];
			w.fbq = n;
			w._fbq ??= n;
			script('https://connect.facebook.net/en_US/fbevents.js');
		}
		w.fbq('init', pixel);
	}
}

/** Manda el evento a GA4 y a Meta, sólo si hay medición activa y consentimiento. */
function enviar(
	ga: [string, Record<string, unknown>] | null,
	meta: [string, Record<string, unknown>, { eventID: string }?] | null
) {
	if (!activa || !acepto(activa.slug)) return;
	const w = window as Ventana;
	// A la medición de esta tienda, no a todas las que se cargaron en la sesión.
	if (ga && activa.medicion.ga4 && w.gtag)
		w.gtag('event', ga[0], { ...ga[1], send_to: activa.medicion.ga4 });
	if (meta && activa.medicion.pixel && w.fbq) {
		const [nombre, datos, opciones] = meta;
		const pixel = activa.medicion.pixel;
		if (opciones) w.fbq('trackSingle', pixel, nombre, datos, opciones);
		else w.fbq('trackSingle', pixel, nombre, datos);
	}
}

/** page_view / PageView: en cada navegación (afterNavigate del layout). */
export function pagina() {
	if (typeof window === 'undefined') return;
	enviar(
		['page_view', { page_location: location.href, page_title: document.title }],
		['PageView', {}]
	);
}

/** view_item / ViewContent: la ficha de un producto. */
export function verProducto(p: ProductoPublico) {
	if (!activa) return;
	const items = itemsGa([{ id: p.id, cantidad: 1 }], [p]);
	enviar(
		['view_item', { currency: 'ARS', value: p.precio, items }],
		['ViewContent', { ...contenidoMeta(items, p.precio), content_name: p.nombre }]
	);
}

/** add_to_cart / AddToCart: lo llama el carrito compartido al sumar. */
export function agregado(slug: string, id: string, cantidad: number) {
	if (!activa || activa.slug !== slug || cantidad <= 0) return;
	const items = itemsGa([{ id, cantidad }], activa.productos);
	if (!items.length) return;
	const v = valor(items);
	enviar(
		['add_to_cart', { currency: 'ARS', value: v, items }],
		['AddToCart', contenidoMeta(items, v)]
	);
}

/** begin_checkout / InitiateCheckout: al entrar al checkout con algo en el carrito. */
export function iniciarCompra(lineas: Linea[]) {
	if (!activa || !lineas.length) return;
	const items = itemsGa(lineas, activa.productos);
	if (!items.length) return;
	const v = valor(items);
	enviar(
		['begin_checkout', { currency: 'ARS', value: v, items }],
		['InitiateCheckout', contenidoMeta(items, v)]
	);
}

/**
 * El carrito se vacía cuando el pedido salió: se guarda lo comprado para la página del pedido
 * (que puede llegar después de pagar en Mercado Pago, en la misma pestaña).
 */
export function recordarCompra(slug: string, lineas: Linea[]) {
	try {
		sessionStorage.setItem(claveCompra(slug), JSON.stringify(lineas.slice(0, 50)));
	} catch {
		// Sin almacenamiento: la compra se mide igual, sin el detalle de productos.
	}
}

/**
 * purchase / Purchase en la página del pedido listo. El número de pedido es el transaction_id de
 * GA4 y el eventID de Meta (deduplican), y además se anota para no mandarlo dos veces al recargar.
 */
export function compra(pedido: { numero: number; total: number }) {
	if (!activa || !acepto(activa.slug)) return;
	const { slug } = activa;
	let medidas: number[];
	let lineas: Linea[] = [];
	try {
		const g = JSON.parse(localStorage.getItem(claveMedidas(slug)) ?? '[]') as unknown;
		medidas = Array.isArray(g) ? (g as number[]) : [];
	} catch {
		medidas = [];
	}
	if (medidas.includes(pedido.numero)) return;
	try {
		const g = JSON.parse(sessionStorage.getItem(claveCompra(slug)) ?? '[]') as Linea[];
		if (Array.isArray(g)) lineas = g;
		sessionStorage.removeItem(claveCompra(slug));
	} catch {
		lineas = [];
	}
	const items = itemsGa(lineas, activa.productos);
	const id = String(pedido.numero);
	enviar(
		['purchase', { transaction_id: id, currency: 'ARS', value: pedido.total, items }],
		['Purchase', contenidoMeta(items, pedido.total), { eventID: id }]
	);
	try {
		localStorage.setItem(
			claveMedidas(slug),
			JSON.stringify([...medidas, pedido.numero].slice(-30))
		);
	} catch {
		// Sin almacenamiento: GA4 y Meta igual deduplican por el número de pedido.
	}
}
