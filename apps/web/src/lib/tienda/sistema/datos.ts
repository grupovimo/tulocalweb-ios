/**
 * Los datos de cada bloque resueltos contra la tienda: productos por fuente, categorías con foto,
 * beneficios automáticos, reseñas con compra, fotos de la galería en filas completas. `seDibuja`
 * decide si un bloque tiene con qué dibujarse (regla 1: nada vacío; el editor dice qué falta).
 */
import { hrefDe as hrefPagina } from '../estilo/menu.ts';
import { proximamente } from '../lanzamiento.ts';
import { tarjetaDe } from '../seleccion.svelte.ts';
import type { ProductoPublico, ResenaPublica, TiendaPublica } from '../tipos.ts';
import {
	campo,
	textoDe,
	type Bloque,
	type BotonDatos,
	type Icono,
	type SeleccionProductos
} from './catalogo.ts';
import { CUOTAS_DESDE, cuantos, precio } from './formato.ts';

export type Tienda = TiendaPublica;

/** El prefijo de los links: '' en el dominio propio; si no, /t/<slug>. */
export const baseDe = (t: Pick<Tienda, 'base' | 'slug'>) => t.base ?? `/t/${t.slug}`;

/** El link de un campo «link» del catálogo (páginas, cat:, prod: o https://). */
export function enlace(t: Tienda, a: string | null | undefined): string | null {
	if (!a) return null;
	const base = baseDe(t);
	if (a.startsWith('prod:')) {
		const p = t.productos.find((x) => x.id === a.slice(5) || x.slug === a.slice(5));
		return p ? productoHref(t, p) : null;
	}
	return hrefPagina(a, base);
}

export const productoHref = (t: Tienda, p: Pick<ProductoPublico, 'id' | 'slug'>) =>
	`${baseDe(t)}/producto/${p.slug ?? p.id}`;
export const categoriaHref = (t: Tienda, nombre: string) =>
	`${baseDe(t)}/catalogo?${new URLSearchParams({ cat: nombre })}`;

/** El número de WhatsApp en dígitos ('' sin número). */
export const numeroWa = (t: Pick<Tienda, 'whatsapp'>) => (t.whatsapp ?? '').replace(/\D/g, '');
/** El link a WhatsApp con un mensaje ('' sin número). */
export function waHref(t: Pick<Tienda, 'whatsapp'>, texto = '¡Hola!'): string {
	const n = numeroWa(t);
	return n ? `https://wa.me/${n}?text=${encodeURIComponent(texto)}` : '';
}

// ---- Productos ---------------------------------------------------------------------------

/** Se puede mostrar en una vidriera: no es un lanzamiento escondido. */
const aLaVista = (p: ProductoPublico) => !!p.nombre;

/**
 * Los productos de una selección («Qué productos» del bloque). Los agotados van al final; nunca
 * más de `cantidad`. Sin ninguno de la fuente elegida (ofertas, una categoría vacía), devuelve [].
 */
export function productosDe(
	t: Tienda,
	sel: Partial<SeleccionProductos> | undefined
): ProductoPublico[] {
	const fuente = sel?.fuente ?? 'vendidos';
	const cantidad = Math.max(1, Math.min(24, sel?.cantidad ?? 8));
	const todos = t.productos.filter(aLaVista);
	let lista: ProductoPublico[];
	switch (fuente) {
		case 'elegidos': {
			const porId = new Map(todos.map((p) => [p.id, p]));
			lista = (sel?.ids ?? []).map((id) => porId.get(id)).filter((p): p is ProductoPublico => !!p);
			break;
		}
		case 'ofertas':
			lista = todos.filter((p) => p.anterior && p.anterior > p.precio);
			break;
		case 'nuevos':
			// Los marcados como nuevos primero; después, del más nuevo al más viejo (el orden de la tienda).
			lista = [
				...todos.filter((p) => /nuevo|novedad/i.test(p.etiqueta ?? '') || proximamente(p)),
				...todos
			];
			break;
		case 'categoria':
			lista = todos.filter((p) => p.categoria && p.categoria === sel?.categoria);
			break;
		default: {
			const pos = new Map(t.masVendidos.map((id, i) => [id, i]));
			lista = [...todos].sort((a, b) => (pos.get(a.id) ?? 1e9) - (pos.get(b.id) ?? 1e9));
		}
	}
	const unicos = [...new Map(lista.map((p) => [p.id, p])).values()];
	if (fuente !== 'elegidos') {
		const hay = unicos.filter((p) => !tarjetaDe(p).agotado);
		const no = unicos.filter((p) => tarjetaDe(p).agotado);
		return [...hay, ...no].slice(0, cantidad);
	}
	return unicos.slice(0, cantidad);
}

// ---- Categorías --------------------------------------------------------------------------

export type CategoriaVista = {
	nombre: string;
	cantidad: string;
	foto: string | null;
	href: string;
};

/**
 * Las categorías del bloque (vacío: todas), con la foto elegida o la de un producto de la
 * categoría: la primera que la página no muestre ya (`usadas`), así la tarjeta de «Abrigos» no
 * repite la foto del sweater que está en la grilla de abajo.
 */
export function categoriasDe(
	t: Tienda,
	items?: { categoria: string; foto?: string }[],
	usadas?: ReadonlySet<string>
): CategoriaVista[] {
	// La foto de la categoría que manda la API es la de su primer producto (no la eligió nadie):
	// cuenta como una candidata más.
	const deCategoria = (nombre: string, primera: string | null | undefined) => {
		const fotos = [
			...new Set(
				[
					primera,
					...t.productos
						.filter((p) => p.categoria === nombre)
						.flatMap((p) => [p.foto, ...(p.fotos ?? [])])
				].filter((f): f is string => !!f)
			)
		];
		return fotos.find((f) => !usadas?.has(f)) ?? fotos[0] ?? null;
	};
	const de = (nombre: string) => t.categorias.find((c) => c.nombre === nombre);
	const lista = items?.length
		? items.map((i) => ({ c: de(i.categoria), foto: i.foto }))
		: t.categorias.map((c) => ({ c, foto: undefined as string | undefined }));
	return lista
		.filter(
			(x): x is { c: Tienda['categorias'][number]; foto: string | undefined } =>
				!!x.c && x.c.cantidad > 0
		)
		.map(({ c, foto }) => ({
			nombre: c.nombre,
			cantidad: cuantos(c.cantidad, 'producto', 'productos'),
			foto: foto ?? deCategoria(c.nombre, c.foto),
			href: categoriaHref(t, c.nombre)
		}))
		.slice(0, 12);
}

// ---- Beneficios --------------------------------------------------------------------------

export type Beneficio = { icono: Icono; titulo: string; texto: string };

/** Los beneficios que salen solos de la configuración de la tienda (nada inventado). */
export function beneficiosAuto(t: Tienda): Beneficio[] {
	const out: Beneficio[] = [];
	const e = t.envios;
	if (e.gratisDesde)
		out.push({ icono: 'envio', titulo: 'Envío gratis', texto: `Desde ${precio(e.gratisDesde)}` });
	else if (e.correo)
		out.push({ icono: 'envio', titulo: 'Envíos a todo el país', texto: 'Por Correo Argentino' });
	else if (e.moto)
		out.push({
			icono: 'envio',
			titulo: 'Envío en moto',
			texto: e.motoZona ? e.motoZona : e.motoDemora || `Desde ${precio(e.motoCosto)}`
		});
	if (t.cobros.cuotas > 1 && t.cobros.mercadoPago)
		out.push({
			icono: 'cuotas',
			titulo: `${t.cobros.cuotas} cuotas sin interés`,
			texto: `Desde ${precio(CUOTAS_DESDE)}`
		});
	if (t.cobros.transferencia > 0 && t.cobros.datosTransferencia)
		out.push({
			icono: 'transferencia',
			titulo: `${t.cobros.transferencia}\u00a0% off`,
			texto: 'Pagando con transferencia'
		});
	if (e.retiro)
		out.push({
			icono: 'retiro',
			titulo: 'Retiro sin cargo',
			texto: e.retiroDireccion || t.ciudad || 'En el local'
		});
	if (out.length < 4)
		out.push({ icono: 'cambios', titulo: 'Compra protegida', texto: '10 días para arrepentirte' });
	return out.slice(0, 4);
}

export function beneficiosDe(t: Tienda, b: Bloque): Beneficio[] {
	const propios = (campo<Beneficio[]>(b, 'items') ?? []).filter((x) => x.titulo?.trim());
	return propios.length ? propios.map((x) => ({ ...x, texto: x.texto ?? '' })) : beneficiosAuto(t);
}

// ---- Reseñas, galería, local -------------------------------------------------------------

/** Sólo las de compra verificada (regla del bloque), las mejores escritas primero. */
export function resenasDe(t: Tienda, cantidad = 3): ResenaPublica[] {
	return t.resenas
		.filter((r) => r.conCompra && r.texto.trim().length >= 8)
		.slice(0, Math.max(1, Math.min(12, cantidad)));
}

/** Cuántas reseñas con compra tiene la tienda. */
export const conCompra = (t: Tienda) => t.resenas.filter((r) => r.conCompra).length;

/** Las fotos de los productos a la vista: la principal y las demás de cada uno, sin repetir. */
const fotosDeProductos = (t: Tienda, conSecundarias = true) => [
	...new Set(
		t.productos
			.filter(aLaVista)
			.flatMap((p) => [p.foto, ...(conSecundarias ? (p.fotos ?? []) : [])])
			.filter((f): f is string => !!f)
	)
];

/**
 * Las fotos de la galería recortadas a filas completas (regla 2): 3, 4 o múltiplos de 6 (6 en la
 * compu, 3 en el celular). Menos de 3, nada. Vacío: las de la tienda y las de los productos, pero
 * sólo las que la página no muestra ya (`usadas`): una galería que repite la grilla de arriba no
 * suma nada, así que si no quedan 3 nuevas no se dibuja.
 */
export function fotosGaleria(t: Tienda, b: Bloque, usadas?: ReadonlySet<string>): string[] {
	const propias = campo<string[]>(b, 'fotos') ?? [];
	const fotos = [...new Set(propias.length ? propias : [...t.fotos, ...fotosDeProductos(t)])]
		.filter((f) => propias.length || !usadas?.has(f))
		.slice(0, 12);
	const n = fotos.length;
	if (n < 3) return [];
	if (n < 6) return fotos.slice(0, n === 5 ? 4 : n);
	return fotos.slice(0, Math.floor(n / 6) * 6);
}

export type Horario = { dias: string; horas: string };

/** El local: dirección, horarios (los del bloque o el del retiro) y el link para llegar. */
export function localDe(t: Tienda, b: Bloque) {
	const direccion = (t.direccion || t.envios.retiroDireccion || '').trim();
	const propios = (campo<Horario[]>(b, 'horarios') ?? []).filter((h) => h.dias && h.horas);
	const horarios: Horario[] = propios.length
		? propios
		: t.envios.retiroHorario
			? [{ dias: 'Horario', horas: t.envios.retiroHorario }]
			: [];
	const lugar = [direccion, t.ciudad].filter(Boolean).join(', ');
	return {
		direccion,
		ciudad: t.ciudad,
		horarios,
		mapa: lugar
			? `https://www.google.com/maps/search/?api=1&query=${encodeURIComponent(lugar)}`
			: '',
		wa: waHref(t, '¡Hola! Quería hacer una consulta.')
	};
}

// ---- Portada -----------------------------------------------------------------------------

/** El título de un bloque con lo sugerido del catálogo (el nombre o el lema de la tienda). */
export function tituloPortada(t: Tienda, b: Bloque): string {
	return textoDe(b, 'titulo') || t.portada?.titulo?.trim() || t.nombre;
}
export function fraseDe(t: Tienda, b: Bloque): string {
	return textoDe(b, 'texto') || t.portada?.lema?.trim() || '';
}
export function fotosPortada(t: Tienda, b: Bloque, usadas?: ReadonlySet<string>): string[] {
	const propias = campo<string[]>(b, 'fotos') ?? [];
	if (propias.length) return propias;
	// La foto de portada de v1 (o la sugerida por la demo) y las de los productos, para collage.
	// Con una sola foto de producto no: el bloque de productos ya la muestra en grande justo abajo
	// y la portada quedaría repitiéndola (Yulz); sin foto propia, la portada va como texto.
	// La API, sin portada propia, manda la primera foto de producto como `portada.foto`: ésa no cuenta.
	// De las de productos, primero las que la página no muestra ya (`usadas`: la grilla de abajo):
	// las segundas fotos y los productos que no entran en la grilla.
	const principales = fotosDeProductos(t, false);
	const deProductos = fotosDeProductos(t);
	const propia = t.portada?.foto && !principales.includes(t.portada.foto) ? [t.portada.foto] : [];
	const candidatas = principales.length >= 2 ? deProductos : [];
	return [
		...new Set([
			...propia,
			...candidatas.filter((f) => !usadas?.has(f)),
			...candidatas.filter((f) => usadas?.has(f))
		])
	].slice(0, 5);
}

/**
 * Las fotos que ya muestra cada lista de bloques, para que los automáticos (portada sin fotos
 * propias, categorías y galería) prefieran otras: primero lo que se eligió (grillas de productos,
 * menú, combo, destacado, ofertas, fotos subidas a un bloque) y después, en el orden de la página,
 * lo que va tomando cada automático. Devuelve, por id de bloque, las fotos que ese bloque debería
 * evitar (las de todos los demás).
 */
export function fotosUsadas(t: Tienda, bloques: Bloque[]): Map<string, Set<string>> {
	const fijas = new Set<string>();
	const sumar = (fs: (string | null | undefined)[]) => {
		for (const f of fs) if (f) fijas.add(f);
	};
	for (const b of bloques) {
		switch (b.tipo) {
			case 'productos':
			case 'oferta':
				sumar(productosDe(t, campo(b, 'productos')).map((p) => p.foto));
				break;
			case 'menu':
				sumar(seccionesMenu(t, b).flatMap((s) => s.productos.map((p) => p.foto)));
				break;
			case 'combo':
				sumar(comboDe(t, b).map((p) => p.foto));
				break;
			case 'destacado':
				sumar([destacadoDe(t, b)?.foto]);
				break;
			case 'portada':
			case 'galeria':
				sumar(campo<string[]>(b, 'fotos') ?? []);
				break;
			case 'textoImagen':
			case 'banner':
				sumar([campo<string>(b, 'foto'), campo<{ foto?: string }>(b, 'segundo')?.foto]);
				break;
		}
	}
	const out = new Map<string, Set<string>>();
	const tomadas = new Set(fijas);
	for (const b of bloques) {
		const evitar = new Set(tomadas);
		out.set(b.id, evitar);
		if (b.tipo === 'portada') {
			const v = b.variante;
			const cuantas =
				v === 'texto' ? 0 : v === 'foto' ? 1 : v === 'partida' || v === 'collage' ? 3 : 5;
			for (const f of fotosPortada(t, b, evitar).slice(0, cuantas)) tomadas.add(f);
		} else if (b.tipo === 'categorias') {
			for (const c of categoriasDe(t, campo(b, 'items'), evitar)) if (c.foto) tomadas.add(c.foto);
		} else if (b.tipo === 'galeria') {
			for (const f of fotosGaleria(t, b, evitar)) tomadas.add(f);
		}
	}
	return out;
}

/** Los botones de la portada: los elegidos o, si no, los que tienen sentido con lo que hay. */
export function botonesPortada(t: Tienda, b: Bloque): { t: string; href: string }[] {
	const ok = (x: BotonDatos | undefined) => {
		const href = x?.t?.trim() ? enlace(t, x.a) : null;
		return href ? { t: x!.t.trim(), href } : null;
	};
	const elegidos = [ok(campo<BotonDatos>(b, 'boton')), ok(campo<BotonDatos>(b, 'boton2'))].filter(
		(x): x is { t: string; href: string } => !!x
	);
	if (elegidos.length) return elegidos;
	const base = baseDe(t);
	const n = t.productos.length;
	const out: { t: string; href: string }[] = [];
	if (n === 1) out.push({ t: 'Ver el producto', href: productoHref(t, t.productos[0]) });
	else if (n > 1) out.push({ t: 'Ver los productos', href: `${base}/catalogo` });
	const cats = t.categorias.filter((c) => c.cantidad > 0);
	if (cats.length >= 2) out.push({ t: cats[0].nombre, href: categoriaHref(t, cats[0].nombre) });
	else if (numeroWa(t)) out.push({ t: 'Escribinos', href: waHref(t) });
	return out.slice(0, 2);
}

/** La línea de abajo de la portada partida: dónde retirar (si hay retiro). */
export function retiroDe(t: Tienda): string {
	const e = t.envios;
	if (!e.retiro || !e.retiroDireccion) return '';
	return `Retiro sin cargo en ${e.retiroDireccion}${e.retiroHorario ? ` · ${e.retiroHorario}` : ''}`;
}

// ---- Oferta, video, turnos ---------------------------------------------------------------

export function hastaOferta(b: Bloque): number | null {
	const h = campo<string>(b, 'hasta');
	const ms = h ? Date.parse(h) : NaN;
	return Number.isFinite(ms) ? ms : null;
}

/** El id del video y su reproductor sin cookies (YouTube) o de Vimeo. */
export function videoDe(url: string | undefined): string | null {
	if (!url) return null;
	const yt = url.match(
		/(?:youtube\.com\/(?:watch\?(?:.*&)?v=|embed\/|shorts\/)|youtu\.be\/)([\w-]{6,20})/
	);
	if (yt) return `https://www.youtube-nocookie.com/embed/${yt[1]}?rel=0`;
	const vm = url.match(/vimeo\.com\/(?:video\/)?(\d{5,12})/);
	if (vm) return `https://player.vimeo.com/video/${vm[1]}?dnt=1`;
	return null;
}

// ---- ¿Se dibuja? -------------------------------------------------------------------------

/**
 * ¿El bloque tiene con qué dibujarse en esta tienda? Los que dependen de datos (local, turnos,
 * WhatsApp, reseñas, una oferta vencida) no se dibujan si falta lo necesario.
 */
export function seDibuja(b: Bloque, t: Tienda, ahora = Date.now()): boolean {
	if (b.oculto) return false;
	switch (b.tipo) {
		case 'portada':
			return true;
		case 'banner':
			return !!textoDe(b, 'titulo');
		case 'oferta': {
			const h = hastaOferta(b);
			if (!h || h <= ahora) return false;
			return b.variante === 'banda' || productosDe(t, campo(b, 'productos')).length > 0;
		}
		case 'textoImagen':
			return !!textoDe(b, 'titulo');
		case 'video':
			return !!videoDe(campo<string>(b, 'url'));
		case 'beneficios':
			return beneficiosDe(t, b).length > 0;
		case 'productos':
			return productosDe(t, campo(b, 'productos')).length > 0;
		case 'destacado':
			return !!destacadoDe(t, b);
		case 'combo':
			return comboDe(t, b).length >= 2;
		case 'categorias':
			return categoriasDe(t, campo(b, 'items')).length >= 2;
		case 'menu':
			return seccionesMenu(t, b).length > 0;
		case 'local':
			return !!localDe(t, b).direccion;
		case 'resenas':
			return resenasDe(t, 1).length > 0;
		case 'preguntas':
			return (campo<unknown[]>(b, 'items') ?? []).length > 0;
		case 'logos':
			return (campo<unknown[]>(b, 'items') ?? []).length > 0;
		case 'galeria':
			return fotosGaleria(t, b).length > 0;
		case 'newsletter':
			return true;
		case 'whatsapp':
			return !!numeroWa(t);
		case 'turnos':
			return !!t.turnos?.servicios.length;
	}
}

export function destacadoDe(t: Tienda, b: Bloque): ProductoPublico | null {
	const id = campo<string>(b, 'producto');
	return t.productos.find((p) => p.id === id || p.slug === id) ?? null;
}

export function comboDe(t: Tienda, b: Bloque): ProductoPublico[] {
	return productosDe(t, {
		fuente: 'elegidos',
		ids: campo<string[]>(b, 'productos') ?? [],
		cantidad: 4
	});
}

/** Las secciones del menú para pedir: una por categoría (vacío: todas), con sus productos. */
export function seccionesMenu(
	t: Tienda,
	b: Bloque
): { nombre: string; productos: ProductoPublico[] }[] {
	const elegidas = campo<string[]>(b, 'categorias') ?? [];
	const nombres = elegidas.length
		? elegidas
		: [...new Set(t.productos.map((p) => p.categoria ?? ''))];
	return nombres
		.map((nombre) => ({
			nombre: nombre || 'Más',
			productos: t.productos.filter((p) => (p.categoria ?? '') === nombre && aLaVista(p))
		}))
		.filter((s) => s.productos.length);
}

/** La inicial de un nombre (para los placeholders de marca y el logo sin foto). */
export function inicial(nombre: string): string {
	const l = nombre.trim().match(/[\p{L}\p{N}]/u)?.[0] ?? '·';
	return l.toLocaleUpperCase('es');
}
