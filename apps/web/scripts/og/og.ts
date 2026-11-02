/**
 * Las imágenes para compartir (static/og/*.jpg, 1200×630): una plantilla con la marca (wordmark,
 * Instrument Sans, terracota y Tincho) y el texto de cada página. Se regeneran con
 *   bun scripts/og/og.ts
 * Los números que aparecen (precios, comisión) salen de lib/panel/planes.ts y de /partners: si
 * cambian allá, se cambian acá.
 */
import { chromium } from '@playwright/test';
import { mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';

const WEB = join(import.meta.dirname, '../..');
const STATIC = join(WEB, 'static');
const archivo = (p: string) => 'file://' + join(STATIC, p);
const fuente = (peso: string) =>
	'file://' +
	join(
		WEB,
		'../../node_modules/@fontsource-variable/instrument-sans/files',
		`instrument-sans-latin-${peso}-normal.woff2`
	);

type Og = {
	id: string;
	titulo: string;
	/** La parte del título en terracota. */
	acento: string;
	texto: string;
	tincho: string;
	fondo: 'claro' | 'oscuro' | 'terracota';
	/** Lo de abajo del texto: logos, chips o nada. */
	pie?: string;
	/** Algo flotando junto a Tincho (una burbuja, una tarjeta). */
	extra?: string;
};

const logos = (ns: string[]) =>
	`<div class="logos">${ns.map((n) => `<span class="logo"><img src="${archivo('marcas/' + n)}"></span>`).join('')}</div>`;
const chips = (ts: string[]) =>
	`<div class="chips">${ts.map((t) => `<span class="chip">${t}</span>`).join('')}</div>`;
const burbuja = (t: string, mia = false, y = 0, x = 0) =>
	`<div class="burbuja${mia ? ' mia' : ''}" style="top:${y}px;right:${x}px">${t}</div>`;

const OGS: Og[] = [
	{
		id: 'home',
		titulo: 'Todo tu negocio.',
		acento: 'Sin complicaciones.',
		texto:
			'Tienda online, WhatsApp, Mercado Libre, el local y la factura de ARCA en un solo panel.',
		tincho: 'saludo',
		fondo: 'claro',
		pie: logos([
			'mercadolibre.svg',
			'tiendanube.svg',
			'correo-argentino.svg',
			'arca.webp',
			'payway.svg'
		]),
		extra: burbuja('Venta nueva · $\u00a044.600', false, 64, 300)
	},
	{
		id: 'general',
		titulo: 'Claro, simple y',
		acento: 'sin letra chica.',
		texto: 'Ayuda, legales y todo lo que necesitás saber para vender tranquilo.',
		tincho: 'portapapeles',
		fondo: 'claro'
	},
	{
		id: 'hornero',
		titulo: 'Tincho atiende tu WhatsApp',
		acento: 'mientras dormís.',
		texto: 'Responde con tu stock real, cobra y te avisa lo importante.',
		tincho: 'whatsapp',
		fondo: 'terracota',
		extra:
			burbuja('¿Tenés el termo en negro?', false, 70, 300) +
			burbuja('¡Sí! Quedan 3. Te llega el jueves. ¿Te mando el link de pago?', true, 130, 250)
	},
	{
		id: 'migrar',
		titulo: 'Mudate a',
		acento: 'Tu\u00a0Local\u00a0Web.',
		texto: 'Traemos tus productos con fotos y variantes. Tus links viejos siguen andando.',
		tincho: 'caja',
		fondo: 'claro',
		pie: chips(['Tiendanube', 'Shopify', 'WooCommerce'])
	},
	{
		id: 'partners',
		titulo: 'Armás tiendas.',
		acento: 'Ganás cada mes.',
		texto:
			'Programa de partners para agencias, diseñadores y contadores. Hasta 30 % de comisión recurrente.',
		tincho: 'celebrando',
		fondo: 'oscuro'
	},
	{
		id: 'plantillas',
		titulo: 'Una tienda que se\u00a0ve',
		acento: 'como tu marca.',
		texto: 'Plantillas para cada rubro: moda, cafés, almacenes, vinotecas, belleza y más.',
		tincho: 'foto',
		fondo: 'claro'
	},
	{
		id: 'precios',
		titulo: 'Empezás gratis.',
		acento: 'Crecés sin sorpresas.',
		texto:
			'Un plan gratis para arrancar y planes pagos desde $\u00a019.900 por mes. Probá todo 14 días sin tarjeta.',
		tincho: 'ticket',
		fondo: 'oscuro',
		pie: chips(['Gratis', 'Emprende $ 19.900', 'Pro $ 49.900', 'Negocio $ 149.900'])
	},
	{
		id: 'seguridad',
		titulo: 'Tus datos, cuidados.',
		acento: 'Y siempre tuyos.',
		texto: 'Conexión cifrada, permisos por persona del equipo y todo exportable cuando quieras.',
		tincho: 'seguridad',
		fondo: 'oscuro'
	}
];

function html(o: Og) {
	const wordmark = o.fondo === 'claro' ? 'wordmark-brown.webp' : 'wordmark-cream.webp';
	// El isotipo del tablero de identidad (static/marca/logo/), en la variante de cada fondo.
	const logo = { claro: 'logo', oscuro: 'logo-sobre-oscuro', terracota: 'logo-crema' }[o.fondo];
	return `<!doctype html><html><head><meta charset="utf-8"><style>
@font-face{font-family:IS;font-weight:400 700;src:url(${fuente('wght')})}
*{box-sizing:border-box}
body{margin:0;width:1200px;height:630px;overflow:hidden;font-family:IS,system-ui,sans-serif;position:relative}
.claro{background:#fff;color:#2a140c}
.oscuro{background:#2a140c;color:#fbf7f2}
.terracota{background:#d2582f;color:#fff}
.mancha{position:absolute;right:-120px;bottom:-200px;width:760px;height:760px;border-radius:50%}
.claro .mancha{background:radial-gradient(closest-side,#fcede5 0%,#fcede5 55%,rgba(252,237,229,0) 100%)}
.oscuro .mancha{background:radial-gradient(closest-side,rgba(210,88,47,.55) 0%,rgba(210,88,47,.18) 55%,rgba(210,88,47,0) 100%)}
.terracota .mancha{background:radial-gradient(closest-side,rgba(255,255,255,.22) 0%,rgba(255,255,255,0) 100%)}
.marca{position:absolute;left:64px;top:50px;display:flex;align-items:center;gap:14px}
.marca .iso{width:46px;height:46px}
.marca .wm{height:34px}
.izq{position:absolute;left:64px;top:150px;width:640px}
h1{margin:0;font-size:68px;line-height:1;letter-spacing:-.045em;font-weight:600;text-wrap:balance}
h1 span{color:#d2582f;display:block}
.terracota h1 span{color:#2a140c}
.oscuro h1 span{color:#e8875f}
p{margin:22px 0 0;font-size:25px;line-height:1.38;font-weight:400;max-width:560px;text-wrap:pretty}
.claro p{color:#6b5244}.oscuro p{color:#cbbcae}.terracota p{color:#fff3ec}
.logos{display:flex;gap:12px;margin-top:30px}
.logo{width:78px;height:78px;border-radius:18px;background:#fff;box-shadow:0 0 0 1px rgba(42,20,12,.1);display:flex;align-items:center;justify-content:center}
.logo img{max-width:58px;max-height:46px}
.chips{display:flex;flex-wrap:wrap;gap:10px;margin-top:28px}
.chip{padding:9px 16px;border-radius:999px;font-size:19px;font-weight:600}
.claro .chip{background:#fcede5;color:#b0441d}
.oscuro .chip{background:rgba(251,247,242,.1);color:#fbf7f2;box-shadow:inset 0 0 0 1px rgba(251,247,242,.16)}
.tincho{position:absolute;right:36px;bottom:-6px;width:430px;height:430px;object-fit:contain;filter:drop-shadow(0 24px 40px rgba(42,20,12,.25))}
.burbuja{position:absolute;max-width:330px;padding:14px 18px;border-radius:20px 20px 20px 6px;background:#fff;color:#2a140c;font-size:21px;line-height:1.35;font-weight:500;box-shadow:0 12px 30px -10px rgba(42,20,12,.35)}
.burbuja.mia{border-radius:18px 18px 6px 18px;background:#2a140c;color:#fbf7f2}
</style></head><body class="${o.fondo}">
<div class="mancha"></div>
<div class="marca"><img class="iso" src="${archivo('marca/logo/' + logo + '.svg')}"><img class="wm" src="${archivo(wordmark)}"></div>
<div class="izq"><h1>${o.titulo} <span>${o.acento}</span></h1><p>${o.texto}</p>${o.pie ?? ''}</div>
<img class="tincho" src="${archivo('hornero/' + o.tincho + '.webp')}">
${o.extra ?? ''}
</body></html>`;
}

// Desde un archivo (no setContent): una página about:blank no puede leer las imágenes locales.
const solo = process.argv[2];
const tmp = mkdtempSync(join(tmpdir(), 'og-'));
const b = await chromium.launch();
const p = await b.newPage({ viewport: { width: 1200, height: 630 } });
for (const o of OGS.filter((x) => !solo || x.id === solo)) {
	const f = join(tmp, `${o.id}.html`);
	writeFileSync(f, html(o));
	await p.goto('file://' + f, { waitUntil: 'load' });
	await p.evaluate(() => document.fonts.ready);
	await p.screenshot({ path: join(STATIC, 'og', `${o.id}.jpg`), type: 'jpeg', quality: 86 });
	console.log('og', o.id);
}
await b.close();
rmSync(tmp, { recursive: true });
