/**
 * La vista previa al compartir una tienda (WhatsApp, Instagram, Facebook): título, descripción e
 * imagen. Todo sale de datos reales de la tienda; lo que la tienda escribe en Mi tienda › Al
 * compartir pisa lo sugerido. La imagen la dibuja el servidor (lib/server/compartir) con el estilo de
 * la plantilla, y su URL lleva una versión que cambia cuando cambia lo que se ve: WhatsApp guarda la
 * vista previa por URL.
 */
import { SITE } from '#lib/site.ts';
import type { ProductoPublico, TiendaPublica } from './tipos.ts';

/** Lo que escribió la tienda en Mi tienda › Al compartir (vacío: lo sugerido). */
export type Compartir = { titulo: string; descripcion: string; foto: string };
export const COMPARTIR_VACIO: Compartir = { titulo: '', descripcion: '', foto: '' };
export const LARGO_TITULO = 70;
export const LARGO_DESCRIPCION = 160;

/** Lo que se sabe de la tienda y sirve para escribir la vista previa. */
export type Hechos = {
	nombre: string;
	/** Título de la portada (si es el nombre, no suma nada). */
	titulo: string;
	lema: string;
	ciudad: string;
	/** Categorías con productos, de la que más tiene a la que menos. */
	categorias: string[];
	productos: number;
	/** Servicios con turno online (vacío: no toma turnos). */
	servicios: string[];
	envios: { correo: boolean; moto: boolean; motoZona: string; retiro: boolean };
	mercadoPago: boolean;
};

export function hechosDe(t: TiendaPublica): Hechos {
	return {
		nombre: t.nombre.trim(),
		titulo: t.portada.titulo?.trim() ?? '',
		lema: t.portada.lema?.trim() ?? '',
		ciudad: t.ciudad?.trim() ?? '',
		categorias: [...t.categorias]
			.filter((c) => c.cantidad > 0 && c.nombre.trim())
			.sort((a, b) => b.cantidad - a.cantidad)
			.map((c) => c.nombre.trim()),
		productos: t.productos.length,
		servicios: (t.turnos?.servicios ?? []).map((s) => s.nombre.trim()).filter(Boolean),
		envios: {
			correo: !!t.envios?.correo,
			moto: !!t.envios?.moto,
			motoZona: t.envios?.motoZona?.trim() ?? '',
			retiro: !!t.envios?.retiro
		},
		mercadoPago: !!t.cobros?.mercadoPago
	};
}

/** "a", "a y b", "a, b y c". */
export function lista(cosas: string[]) {
	if (cosas.length < 2) return cosas[0] ?? '';
	return `${cosas.slice(0, -1).join(', ')} y ${cosas.at(-1)}`;
}

/** Corta en la última palabra entera que entra, con «…». */
export function recortar(texto: string, max: number) {
	// Los espacios duros (el de los precios) quedan: un precio no se parte.
	const t = texto.replace(/[^\S\u00a0]+/g, ' ').trim();
	if (t.length <= max) return t;
	const corte = t.slice(0, max - 1);
	const espacio = corte.lastIndexOf(' ');
	return `${(espacio > max * 0.6 ? corte.slice(0, espacio) : corte).replace(/[\s,;:.·-]+$/, '')}…`;
}

const iguales = (a: string, b: string) => a.localeCompare(b, 'es', { sensitivity: 'base' }) === 0;

/** El lema, si dice algo distinto del nombre. */
const lemaDe = (h: Hechos) => (h.lema && !iguales(h.lema, h.nombre) ? h.lema : '');
/** El título de la portada, si no es el nombre (ni el lema). */
const tituloDe = (h: Hechos) =>
	h.titulo && !iguales(h.titulo, h.nombre) && !iguales(h.titulo, h.lema) ? h.titulo : '';

/** La frase corta de la tienda: lema, título de la portada o qué vende. Vacía si no hay nada. */
export function fraseDe(h: Hechos) {
	return (
		lemaDe(h) ||
		tituloDe(h) ||
		(h.categorias.length ? lista(h.categorias.slice(0, 3)) : '') ||
		(h.servicios.length ? `Turnos online: ${lista(h.servicios.slice(0, 3))}` : '')
	);
}

/** «24 productos · Córdoba», con lo que haya. */
export function detalleDe(h: Hechos) {
	return [
		h.productos ? `${h.productos} ${h.productos === 1 ? 'producto' : 'productos'}` : '',
		!h.productos && h.servicios.length ? 'Turnos online' : '',
		h.ciudad
	]
		.filter(Boolean)
		.join(' · ');
}

/** Cómo entrega, en una frase (vacía si no se sabe). */
function entrega(h: Hechos) {
	if (h.envios.correo) return 'Envíos a todo el país';
	if (h.envios.moto)
		return h.envios.motoZona ? `Envíos en moto a ${h.envios.motoZona}` : 'Envíos en moto';
	if (h.envios.retiro) return h.ciudad ? `Retiro en ${h.ciudad}` : 'Retiro en el local';
	return '';
}

/**
 * Título y descripción sugeridos. El título suma al nombre lo que mejor dice qué es la tienda (el
 * lema, el título de la portada o lo que vende); la descripción cuenta lo demás, sin repetirlo.
 */
export function textosPorDefecto(h: Hechos): { titulo: string; descripcion: string } {
	const lema = lemaDe(h);
	const titulo = tituloDe(h);
	const cats = h.categorias.slice(0, 3);
	const conNombre = (extra: string) =>
		extra && h.nombre.length + extra.length + 3 <= LARGO_TITULO ? `${h.nombre} · ${extra}` : '';
	// Qué va en el título (y entonces no se repite abajo).
	let usado: 'lema' | 'titulo' | 'cats' | 'turnos' | 'tienda' = 'tienda';
	let t = '';
	for (const [k, extra] of [
		['lema', lema],
		['titulo', titulo],
		['cats', lista(cats.slice(0, 2))],
		['turnos', h.servicios.length && !h.productos ? 'Turnos online' : '']
	] as const) {
		t = conNombre(extra);
		if (t) {
			usado = k;
			break;
		}
	}
	if (!t) t = conNombre('Tienda online') || recortar(h.nombre, LARGO_TITULO);

	const frases: string[] = [];
	if (lema && usado !== 'lema') frases.push(lema);
	else if (titulo && usado !== 'titulo' && usado !== 'lema') frases.push(titulo);
	if (h.productos) {
		const n = `${h.productos} ${h.productos === 1 ? 'producto' : 'productos'}`;
		frases.push(
			cats.length && usado !== 'cats'
				? `${lista(cats)}: ${n} para comprar online`
				: `${n} para comprar online`
		);
	}
	if (h.servicios.length)
		frases.push(
			usado === 'turnos' || !h.productos
				? `Reservá tu turno: ${lista(h.servicios.slice(0, 3))}`
				: 'También hay turnos online'
		);
	const e = entrega(h);
	if (e) frases.push(e);
	if (h.mercadoPago && h.productos) frases.push('Pagás con Mercado Pago');
	if (!frases.length)
		frases.push(h.ciudad ? `Tienda online en ${h.ciudad}` : `Conocé ${h.nombre} y escribinos`);
	let descripcion = '';
	for (const f of frases) {
		const sig = `${descripcion}${descripcion ? ' ' : ''}${f.replace(/[.…]$/, '')}.`;
		if (sig.length > LARGO_DESCRIPCION) break;
		descripcion = sig;
	}
	return { titulo: t, descripcion: descripcion || recortar(frases[0], LARGO_DESCRIPCION) };
}

/** Título y descripción de la vista previa: lo escrito por la tienda o lo sugerido. */
export function textosCompartir(t: TiendaPublica) {
	const sugeridos = textosPorDefecto(hechosDe(t));
	const c = t.compartir ?? COMPARTIR_VACIO;
	return {
		titulo: c.titulo.trim() || sugeridos.titulo,
		descripcion: c.descripcion.trim() || sugeridos.descripcion
	};
}

/** Precio como en la tienda: "$ 42.900". */
const precio = (n: number) => '$ ' + Math.round(n).toLocaleString('es-AR');

/** Título y descripción de un producto: la descripción nunca repite el título. */
export function textosProducto(t: TiendaPublica, p: ProductoPublico) {
	const titulo = recortar(`${p.nombre} · ${t.nombre}`, LARGO_TITULO);
	const oferta = p.anterior && p.anterior > p.precio ? ` (antes ${precio(p.anterior)})` : '';
	const desc = p.descripcion?.trim();
	const extra = desc
		? desc
		: [
				p.stock > 0 ? 'Disponible para comprar online.' : 'Sin stock por ahora.',
				entrega(hechosDe(t)) ? `${entrega(hechosDe(t))}.` : ''
			]
				.filter(Boolean)
				.join(' ');
	return {
		titulo,
		descripcion: recortar(`${precio(p.precio)}${oferta}. ${extra}`, LARGO_DESCRIPCION)
	};
}

// ---- La imagen ----

/** Lo que dibuja la imagen de una tienda. Sólo datos de la tienda. */
export type TarjetaTienda = {
	tipo: 'tienda';
	plantilla: string;
	paleta: number;
	nombre: string;
	/** URL del logo ('' si no hay: va el nombre en letras). */
	logo: string;
	frase: string;
	detalle: string;
	/** Las categorías con más productos (hasta 5), con cuántos tiene cada una. */
	categorias: { nombre: string; cantidad: number }[];
	/** Dónde está la tienda (dominio propio o <slug>.tulocalweb.com.ar). */
	direccion: string;
	/** La portada y las fotos de productos (hasta 4, sin repetir). */
	fotos: string[];
	/** Una foto propia subida en Al compartir: va sola, recortada a 1200 × 630. */
	propia: string;
};

export type TarjetaProducto = {
	tipo: 'producto';
	plantilla: string;
	paleta: number;
	tienda: string;
	logo: string;
	nombre: string;
	precio: string;
	anterior: string;
	foto: string;
	direccion: string;
};

export type Tarjeta = TarjetaTienda | TarjetaProducto;

/** Dirección visible: el dominio propio o la de Tu Local Web, sin https. */
export function direccionDe(t: Pick<TiendaPublica, 'slug' | 'dominio'>) {
	return t.dominio || `${t.slug}.${SITE.replace(/^https?:\/\//, '')}`;
}

export function tarjetaDe(t: TiendaPublica): TarjetaTienda {
	const h = hechosDe(t);
	const fotos = [
		...new Set([t.portada.foto, ...t.productos.map((p) => p.foto)].filter((f): f is string => !!f))
	].slice(0, 4);
	return {
		tipo: 'tienda',
		plantilla: t.plantilla,
		paleta: t.paleta,
		nombre: h.nombre,
		logo: t.logo || '',
		frase: recortar(fraseDe(h), 110),
		detalle: detalleDe(h),
		categorias: [...t.categorias]
			.filter((c) => c.cantidad > 0 && c.nombre.trim())
			.sort((a, b) => b.cantidad - a.cantidad)
			.slice(0, 5)
			.map((c) => ({ nombre: recortar(c.nombre, 28), cantidad: c.cantidad })),
		direccion: direccionDe(t),
		fotos,
		propia: t.compartir?.foto || ''
	};
}

export function tarjetaProducto(t: TiendaPublica, p: ProductoPublico): TarjetaProducto {
	return {
		tipo: 'producto',
		plantilla: t.plantilla,
		paleta: t.paleta,
		tienda: t.nombre.trim(),
		logo: t.logo || '',
		nombre: recortar(p.nombre, 90),
		precio: precio(p.precio),
		anterior: p.anterior && p.anterior > p.precio ? precio(p.anterior) : '',
		foto: p.foto || '',
		direccion: direccionDe(t)
	};
}

/** Sube cuando cambia el dibujo: todas las URLs viejas pasan a ser otras. */
const DIBUJO = 1;

/** Versión corta (FNV-1a de 32 bits en base 36) de lo que se dibuja. */
export function versionDe(tarjeta: Tarjeta) {
	let h = 0x811c9dc5;
	const s = `${DIBUJO}|${JSON.stringify(tarjeta)}`;
	for (let i = 0; i < s.length; i++) {
		h ^= s.charCodeAt(i);
		h = Math.imul(h, 0x01000193);
	}
	return (h >>> 0).toString(36);
}

/** URL absoluta de la imagen de la tienda (siempre en tulocalweb.com.ar: sirve en el dominio propio). */
export function imagenTienda(t: TiendaPublica) {
	return `${SITE}/t/${t.slug}/compartir/${versionDe(tarjetaDe(t))}.jpg`;
}

export function imagenProducto(t: TiendaPublica, p: ProductoPublico) {
	return `${SITE}/t/${t.slug}/producto/${encodeURIComponent(p.slug ?? p.id)}/compartir/${versionDe(tarjetaProducto(t, p))}.jpg`;
}
