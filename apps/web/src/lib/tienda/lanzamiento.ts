/**
 * Lanzamientos («drops»), ofertas con fecha de fin y señas (docs/decisiones.md §52), lo que es igual
 * en la tienda, el panel y el sandbox: qué contar en la cuenta regresiva, cómo se parte el tiempo,
 * la hora de Argentina y las reglas para guardar (las mismas que Go, lanzamientos.validar).
 */
import type { ProductoPublico, TiendaPublica } from './tipos.ts';

export const ZONA = 'America/Argentina/Buenos_Aires';

/** El producto todavía no salió a la venta. */
export function proximamente(p: Pick<ProductoPublico, 'lanzamiento'>, ahora = Date.now()) {
	return !!p.lanzamiento && Date.parse(p.lanzamiento) > ahora;
}

/** Lo único que se cuenta: un lanzamiento próximo o una oferta con fin. Nunca una cuenta inventada. */
export type Conteo = {
	tipo: 'lanzamiento' | 'oferta';
	producto: ProductoPublico;
	/** Epoch en ms. */
	hasta: number;
};

/** El evento real más cercano de la tienda (lanzamiento u oferta que termina), o null. */
export function proximoConteo(
	t: Pick<TiendaPublica, 'productos'>,
	ahora = Date.now()
): Conteo | null {
	let mejor: Conteo | null = null;
	for (const p of t.productos) {
		const l = p.lanzamiento ? Date.parse(p.lanzamiento) : NaN;
		const o = p.ofertaHasta && p.anterior ? Date.parse(p.ofertaHasta) : NaN;
		for (const [tipo, hasta] of [
			['lanzamiento', l],
			['oferta', o]
		] as const)
			if (hasta > ahora && (!mejor || hasta < mejor.hasta)) mejor = { tipo, producto: p, hasta };
	}
	return mejor;
}

/** El % de la oferta (redondeado), o 0. */
export const descuento = (p: Pick<ProductoPublico, 'precio' | 'anterior'>) =>
	p.anterior && p.anterior > p.precio ? Math.round((1 - p.precio / p.anterior) * 100) : 0;

export type Parte = { v: string; u: string };

/**
 * El tiempo que falta en partes de dos cifras, como el handoff (`cd`). Con 4: días, horas, minutos
 * y segundos. Con 3 (las plantillas que muestran tres): horas, minutos y segundos, o días, horas y
 * minutos cuando falta más de un día. `unidades` son los rótulos de la plantilla (d h m s / DÍAS…).
 */
export function partes(ms: number, n: 3 | 4 = 4, unidades = ['d', 'h', 'm', 's']): Parte[] {
	const s = Math.max(0, Math.floor(ms / 1000));
	const pad = (x: number) => String(Math.floor(x)).padStart(2, '0');
	const todo = [
		{ v: pad(s / 86400), u: unidades[0] },
		{ v: pad((s % 86400) / 3600), u: unidades[1] },
		{ v: pad((s % 3600) / 60), u: unidades[2] },
		{ v: pad(s % 60), u: unidades[3] }
	];
	if (n === 4) return todo;
	return s >= 86400 ? todo.slice(0, 3) : [{ ...todo[1], v: pad(s / 3600) }, todo[2], todo[3]];
}

const FECHA = new Intl.DateTimeFormat('es-AR', {
	timeZone: ZONA,
	weekday: 'long',
	day: 'numeric',
	month: 'long'
});
const HORA = new Intl.DateTimeFormat('es-AR', {
	timeZone: ZONA,
	hour: '2-digit',
	minute: '2-digit',
	hour12: false
});

/** «viernes 7 de noviembre a las 19:00» (hora de Argentina). */
export function cuando(iso: string | number) {
	const d = new Date(iso);
	return `${FECHA.format(d)} a las ${HORA.format(d)}`;
}

/** «7/11 19:00» corto, hora de Argentina. */
export function cuandoCorto(iso: string | number) {
	const d = new Date(iso);
	const dia = new Intl.DateTimeFormat('es-AR', {
		timeZone: ZONA,
		day: 'numeric',
		month: 'numeric'
	});
	return `${dia.format(d)} ${HORA.format(d)}`;
}

/** Lo que dice la barra junto a la cuenta: qué sale o qué oferta termina. */
export function textoConteo(c: Conteo) {
	if (c.tipo === 'lanzamiento') return `${c.producto.nombre} sale en`;
	const pct = descuento(c.producto);
	return pct
		? `${c.producto.nombre} con ${pct}% off · termina en`
		: `Oferta en ${c.producto.nombre} · termina en`;
}

/** Para un <input type="datetime-local">: «2026-11-07T19:00» en hora de Argentina. */
export function aLocal(iso: string | null | undefined) {
	if (!iso) return '';
	const d = new Date(Date.parse(iso) - 3 * 3600_000);
	return d.toISOString().slice(0, 16);
}

/** De un <input type="datetime-local"> (hora de Argentina, UTC−3 todo el año) a ISO. */
export function deLocal(v: string) {
	if (!/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}$/.test(v)) return null;
	return new Date(`${v}:00-03:00`).toISOString();
}

/** La seña: el % del total redondeado al peso, nunca menos de $ 1 (Go, vidriera.SenaDe). */
export const senaDe = (total: number, pct: number) => Math.max(1, Math.round((total * pct) / 100));

/** Lo que la tienda configura en el editor del producto (LanzamientoEditable). */
export type ConfigLanzamiento = {
	lanzamiento: string | null;
	reserva: number | null;
	encargo: number | null;
	ofertaPrecio: number | null;
	ofertaHasta: string | null;
};

export const SIN_LANZAMIENTO: ConfigLanzamiento = {
	lanzamiento: null,
	reserva: null,
	encargo: null,
	ofertaPrecio: null,
	ofertaHasta: null
};

/** La configuración con lo que pasó: avisos pedidos y enviados, reservas y encargos con seña. */
export type LanzamientoProducto = ConfigLanzamiento & {
	avisos: number;
	avisados: number;
	reservas: number;
	encargos: number;
};

/**
 * Lo que vale ahora en la tienda (Go, vidriera.lanzamiento.aplicar): la oferta vigente, «Próximamente»
 * antes del lanzamiento y las señas que se pueden pagar (con Mercado Pago). Sólo devuelve las claves
 * que cambian el producto público.
 */
export function aplicarLanzamiento(
	precio: number,
	c: ConfigLanzamiento | undefined,
	mp: boolean,
	ahora = Date.now()
): Partial<ProductoPublico> {
	if (!c) return {};
	const out: Partial<ProductoPublico> = {};
	if (
		c.ofertaPrecio !== null &&
		c.ofertaPrecio < precio &&
		(!c.ofertaHasta || Date.parse(c.ofertaHasta) > ahora)
	) {
		out.precio = c.ofertaPrecio;
		out.anterior = precio;
		out.ofertaHasta = c.ofertaHasta;
	}
	if (c.lanzamiento && Date.parse(c.lanzamiento) > ahora) {
		out.lanzamiento = c.lanzamiento;
		out.etiqueta = 'Próximamente';
		if (mp && c.reserva) out.reserva = c.reserva;
	} else if (mp && c.encargo) out.encargo = c.encargo;
	return out;
}

const igual = (a: string | null, b: string | null) =>
	(!a && !b) || (!!a && !!b && Date.parse(a) === Date.parse(b));

/** Por qué no se puede guardar, o null (las mismas reglas que Go). */
export function errorLanzamiento(
	c: ConfigLanzamiento,
	antes: ConfigLanzamiento,
	precio: number,
	ahora = Date.now()
): string | null {
	const pct = (p: number | null) => p === null || (Number.isInteger(p) && p >= 1 && p <= 100);
	if (
		c.lanzamiento &&
		!igual(c.lanzamiento, antes.lanzamiento) &&
		Date.parse(c.lanzamiento) <= ahora
	)
		return 'La fecha del lanzamiento tiene que ser más adelante que ahora.';
	if (c.lanzamiento && Date.parse(c.lanzamiento) > ahora + 366 * 86400_000)
		return 'El lanzamiento puede ser hasta dentro de un año.';
	if (c.reserva !== null && !c.lanzamiento)
		return 'Para reservar con seña, poné la fecha del lanzamiento.';
	if (!pct(c.reserva) || !pct(c.encargo)) return 'La seña va de 1 % a 100 %.';
	if (c.ofertaPrecio !== null && (c.ofertaPrecio <= 0 || c.ofertaPrecio >= precio))
		return 'El precio de oferta tiene que ser menor que el precio del producto.';
	if (c.ofertaHasta && c.ofertaPrecio === null)
		return 'Para ponerle fin a la oferta, cargá el precio de oferta.';
	if (
		c.ofertaHasta &&
		!igual(c.ofertaHasta, antes.ofertaHasta) &&
		Date.parse(c.ofertaHasta) <= ahora
	)
		return 'La oferta tiene que terminar más adelante que ahora.';
	return null;
}
