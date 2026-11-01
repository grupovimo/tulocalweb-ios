import { dev } from '$app/env';
import type { Handle, HandleServerError } from '@sveltejs/kit/hooks';
import { env } from 'cloudflare:workers';
import { csrfForbidden } from '#lib/server/csrf.ts';
import { COOKIE, dbSandbox, leer } from '#lib/server/sandbox/sandbox.ts';
import { COOKIE_SESION, COOKIE_SESION_VIEJA } from '#lib/server/auth/sesion.ts';
import { COOKIE_EQUIPO, COOKIE_MIRADA } from '#lib/server/staff/index.ts';
import { direccionNueva, redireccionTienda } from '#lib/server/tienda/redireccion.ts';
import { olvidarDominio, slugDeDominio } from '#lib/server/tienda/dominio.ts';
import { contarVista } from '#lib/server/tienda/vistas.ts';
import { esDominioPropio, GLOBALES_EN_DOMINIO } from '#lib/tienda/dominio.ts';

/**
 * Rutas que usan el sandbox de la demo: sólo ahí se lee D1 (la landing no gasta lecturas).
 * /partners/panel es el panel de partners de ejemplo. /oauth/autorizar ata la IA al sandbox del visitante; /api/mcp lo saca del token, no de la cookie.
 */
const CON_SANDBOX = /^\/(prueba|panel|integraciones|oauth|partners\/panel)(\/|$)/;

/**
 * CSP de la tienda online (/t/...): los scripts y las conexiones sólo pueden ir al propio sitio y a
 * lo que necesitan GA4 (gtag.js) y el píxel de Meta (fbevents.js), que la tienda carga con el
 * consentimiento de quien compra (lib/tienda/medicion.ts). 'unsafe-inline' es por el arranque de
 * SvelteKit; las fotos pueden venir de cualquier https (Unsplash, el CDN de la tienda).
 */
const CSP_TIENDA = [
	"default-src 'self'",
	"script-src 'self' 'unsafe-inline' https://www.googletagmanager.com https://connect.facebook.net https://static.cloudflareinsights.com",
	// wss://tulocalweb.com.ar: el tiempo real de una tienda en su dominio propio (§44).
	"connect-src 'self' wss://tulocalweb.com.ar https://www.googletagmanager.com https://*.google-analytics.com https://*.analytics.google.com https://connect.facebook.net https://www.facebook.com https://cloudflareinsights.com",
	"img-src 'self' data: https:",
	"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com",
	"font-src 'self' data: https://fonts.gstatic.com",
	'frame-src https://www.facebook.com https://www.googletagmanager.com',
	"object-src 'none'",
	"base-uri 'self'",
	"frame-ancestors 'self'"
].join('; ');

/** La CSP de la tienda con el origen del tiempo real de las pruebas, si lo hay (TIEMPO_REAL_ORIGEN). */
function cspTienda() {
	const o = String(env.TIEMPO_REAL_ORIGEN ?? '');
	return o && /^wss?:\/\/[\w.:-]+$/.test(o)
		? CSP_TIENDA.replace("connect-src 'self'", `connect-src 'self' ${o}`)
		: CSP_TIENDA;
}

/**
 * Encabezados de seguridad en todas las respuestas (también las tempranas, como el rechazo por CSRF).
 * Nadie puede meter el panel en un iframe, el navegador no adivina tipos y siempre se usa HTTPS.
 * El iframe "Panel de Tu Local Web en el celular" de la landing es del mismo origen: SAMEORIGIN.
 */
function secure(res: Response): Response {
	const set = (r: Response) => {
		r.headers.set('X-Content-Type-Options', 'nosniff');
		r.headers.set('Referrer-Policy', 'strict-origin-when-cross-origin');
		r.headers.set('Permissions-Policy', 'camera=(self), microphone=(self), geolocation=()');
		r.headers.set('X-Frame-Options', 'SAMEORIGIN');
		if (!dev) r.headers.set('Strict-Transport-Security', 'max-age=31536000; includeSubDomains');
		return r;
	};
	try {
		return set(res);
	} catch {
		// Encabezados inmutables (respuesta de fetch o de la caché): se copia.
		return set(new Response(res.body, res));
	}
}

/** Un pedido de página (no un fetch, una imagen ni un iframe de otro sitio). */
function esDocumento(h: Headers) {
	const dest = h.get('sec-fetch-dest');
	return dest ? dest === 'document' : (h.get('accept') ?? '').includes('text/html');
}

/** Rutas de la web que la app de iPhone llama con `Authorization: Bearer`. */
const RUTAS_APP = new Set(['/panel/fotos', '/panel/hornero']);

export const handle: Handle = async ({ event, resolve }) => {
	// www.tulocalweb.com.ar es el mismo Worker: se manda al dominio principal, con el mismo camino.
	if (event.url.hostname.startsWith('www.')) {
		const u = new URL(event.url);
		u.hostname = u.hostname.slice(4);
		return secure(new Response(null, { status: 301, headers: { location: u.href } }));
	}
	// El dominio propio de una tienda sólo muestra esa tienda (reroute la llevó a /t/<slug>): ni el
	// panel, ni la landing, ni otra tienda.
	event.locals.dominio = null;
	if (esDominioPropio(event.url.hostname)) {
		const slug = await slugDeDominio(event.url.hostname).catch(() => undefined);
		if (slug === undefined)
			return secure(new Response('Probá de nuevo en un momento.', { status: 503 }));
		const suya =
			slug &&
			(GLOBALES_EN_DOMINIO.test(event.url.pathname) ||
				!event.route.id ||
				(event.route.id.startsWith('/t/[tienda]') && event.params.tienda === slug));
		if (!slug || !suya)
			return secure(
				new Response(
					slug ? 'No encontramos esta página.' : 'Este dominio todavía no tiene una tienda.',
					{
						status: 404,
						headers: { 'content-type': 'text/plain; charset=utf-8' }
					}
				)
			);
		event.locals.dominio = slug;
	}
	if (!dev && csrfForbidden(event.request, event.url))
		return secure(
			new Response(`Cross-site ${event.request.method} form submissions are forbidden`, {
				status: 403
			})
		);
	const started = Date.now();
	event.locals.sandbox = null;
	event.locals.token = null;
	event.locals.equipo = null;
	event.locals.mirada = false;
	if (String(env.DATA_SOURCE) === 'api') {
		// La sesión la valida Go en cada pedido del panel; acá sólo se lee la cookie.
		event.locals.token =
			event.cookies.get(COOKIE_SESION) ?? event.cookies.get(COOKIE_SESION_VIEJA) ?? null;
		// La app de iPhone usa dos cosas de la web con su token (el mismo de la API, que Go valida):
		// subir fotos y hablar con Tincho. Sólo en esas dos rutas.
		if (!event.locals.token && RUTAS_APP.has(event.url.pathname)) {
			const auth = event.request.headers.get('authorization') ?? '';
			if (auth.startsWith('Bearer ')) event.locals.token = auth.slice(7).trim() || null;
		}
		if (event.url.pathname.startsWith('/staff'))
			event.locals.equipo = event.cookies.get(COOKIE_EQUIPO) ?? null;
		// Alguien del equipo viendo una tienda: el panel usa ese token de sólo lectura (Go rechaza
		// cualquier escritura con él) y muestra el aviso fijo.
		const mirada = event.cookies.get(COOKIE_MIRADA);
		if (mirada && event.url.pathname.startsWith('/panel')) {
			event.locals.token = mirada;
			event.locals.mirada = true;
		}
	} else if (CON_SANDBOX.test(event.url.pathname)) {
		event.locals.sandbox = await leer(dbSandbox(env.SANDBOX), event.cookies.get(COOKIE));
	}
	let res = secure(await resolve(event));
	// Visitas de la tienda (decisiones §36): la carga completa de una página se cuenta acá, en
	// segundo plano (no suma latencia). La navegación dentro de la tienda la avisa el layout.
	if (
		event.request.method === 'GET' &&
		res.status === 200 &&
		!event.isDataRequest &&
		event.route.id?.startsWith('/t/[tienda]') &&
		esDocumento(event.request.headers)
	)
		contarVista(event, event.route.id, event.params.id);
	// Sesión vencida o inválida al abrir una página del panel: en vez de la página de error, se borra
	// la cookie y se va a ingresar, que después vuelve a donde estaba.
	if (
		res.status === 401 &&
		event.request.method === 'GET' &&
		!event.locals.mirada &&
		/^\/(panel|onboarding)(\/|$)/.test(event.url.pathname) &&
		!event.url.pathname.includes('/__data.json') &&
		(event.request.headers.get('accept') ?? '').includes('text/html')
	) {
		const volver = encodeURIComponent(event.url.pathname + event.url.search);
		res = new Response(null, { status: 303, headers: { location: `/ingresar?volver=${volver}` } });
		for (const c of [COOKIE_SESION, COOKIE_SESION_VIEJA])
			res.headers.append(
				'set-cookie',
				`${c}=; Path=/; Max-Age=0; HttpOnly; SameSite=Lax${dev ? '' : '; Secure'}`
			);
	}
	// Un link de la tienda vieja (importada): si el path de la tienda no existe y había una dirección
	// guardada, 301 a la nueva. Sólo ante un 404, así las páginas que existen no consultan nada.
	// En el dominio propio el path no lleva /t/<slug>: se consulta con él y se responde sin él.
	const dom = event.locals.dominio;
	if (
		res.status === 404 &&
		event.request.method === 'GET' &&
		(dom || event.url.pathname.startsWith('/t/'))
	) {
		const base = dom ? `/t/${dom}` : '';
		// Primero, si es una dirección vieja de la tienda (la cambió, decisiones §38): 301 a la nueva
		// con el mismo camino y la misma consulta (las vueltas de Mercado Pago traen la suya).
		const nueva = await direccionNueva(base + event.url.pathname, event.cookies.get(COOKIE)).catch(
			() => null
		);
		if (nueva && dom) {
			// El dominio propio sigue andando: este isolate tenía guardada la dirección vieja. Se olvida
			// y se vuelve a pedir la misma URL, que ya sale con la nueva.
			olvidarDominio(event.url.hostname);
			res = secure(new Response(null, { status: 307, headers: { location: event.url.href } }));
		} else if (nueva) {
			res = secure(
				new Response(null, { status: 301, headers: { location: nueva + event.url.search } })
			);
		} else {
			let hacia = await redireccionTienda(
				base + event.url.pathname,
				event.cookies.get(COOKIE)
			).catch(() => null);
			if (hacia && dom) hacia = hacia.slice(base.length) || '/';
			if (hacia) res = secure(new Response(null, { status: 301, headers: { location: hacia } }));
		}
	}
	try {
		res.headers.set('Server-Timing', `app;dur=${Date.now() - started}`);
		// En desarrollo, Vite inyecta scripts y un websocket propios: la CSP va sólo en el build.
		if (!dev && (dom || event.url.pathname.startsWith('/t/')))
			res.headers.set('Content-Security-Policy', cspTienda());
	} catch {
		// Respuesta inmutable: el diagnóstico no vale una copia.
	}
	return res;
};

/**
 * Un error inesperado (la API no contesta, un bug): los avisos del panel muestran este texto en lugar
 * del «Internal Error» de SvelteKit. Los esperados (`error(400, …)`) y los 404 quedan como vienen.
 */
export const handleError: HandleServerError = ({ kind, error }) => {
	if (kind !== 'unknown') return;
	console.error(error);
	return {
		message: 'Algo falló de nuestro lado. Probá de nuevo en un rato; lo que escribiste sigue ahí.'
	};
};
