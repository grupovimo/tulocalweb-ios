/**
 * Las líneas del carrito de la tienda, sin estado: una por producto y variante. El carrito
 * (carrito.svelte.ts) las guarda en el navegador; el servidor las vuelve a validar al comprar.
 */
import {
	alcanza,
	falta,
	mayoristaDeVariante,
	precioMayorista,
	type MinimoMayorista
} from './mayorista.ts';
import type { ProductoPublico } from './tipos.ts';
import { precioDe, textoVariante } from './variantes.ts';

export type Linea = { id: string; varianteId: string | null; cantidad: number };

export const MAX_LINEAS = 50;

/** Identifica la línea: el producto y, si tiene, la variante. */
export const claveLinea = (l: { id: string; varianteId?: string | null }) =>
	l.varianteId ? `${l.id}~${l.varianteId}` : l.id;

/**
 * Lo guardado en el navegador, revisado. Los carritos de antes de las variantes no tenían
 * `varianteId`: sus líneas siguen valiendo como producto sin variante.
 */
export function migrarLineas(g: unknown): Linea[] {
	if (!Array.isArray(g)) return [];
	const vistas = new Set<string>();
	const lineas: Linea[] = [];
	for (const x of g as Partial<Linea>[]) {
		if (!x || typeof x.id !== 'string' || !x.id || x.id.length > 80) continue;
		const cantidad = Number(x.cantidad);
		if (!Number.isInteger(cantidad) || cantidad < 1) continue;
		const l: Linea = {
			id: x.id,
			varianteId: typeof x.varianteId === 'string' && x.varianteId ? x.varianteId : null,
			cantidad: Math.min(99, cantidad)
		};
		if (vistas.has(claveLinea(l))) continue;
		vistas.add(claveLinea(l));
		lineas.push(l);
		if (lineas.length >= MAX_LINEAS) break;
	}
	return lineas;
}

/** Una línea con su producto tal como se compra: precio, foto y stock de la variante. */
export type LineaDetalle = Linea & {
	clave: string;
	/** "Talle M · Negro" (null sin variante). */
	variante: string | null;
	/**
	 * Con el precio que se cobra: al por mayor, `precio` es el mayorista y `anterior` el minorista
	 * (así la plantilla lo tacha con lo que ya tiene).
	 */
	producto: ProductoPublico;
	subtotal: number;
	/** El precio de lista de la unidad, el que cuenta para el mínimo mayorista. */
	minorista: number;
	/** El precio mayorista de la unidad si la tienda vende por mayor y es menor (null: no tiene). */
	mayorista: number | null;
	/** El carrito llegó al mínimo y esta línea va a precio mayorista. */
	alPorMayor: boolean;
};

/**
 * Las líneas con su producto. Se descartan las de productos que ya no están, las que piden una
 * variante que no existe (o ya no está activa) y las de productos con opciones sin variante. Con el
 * carrito al mínimo mayorista (`minimo`: TiendaPublica.mayorista), las que tienen precio mayorista
 * pasan a ese, como lo cobra el servidor (§49).
 */
export function detalleLineas(
	lineas: Linea[],
	productos: ProductoPublico[],
	minimo: MinimoMayorista | null = null
): LineaDetalle[] {
	const porId = new Map(productos.map((p) => [p.id, p]));
	const base = lineas.flatMap((l): LineaDetalle[] => {
		const p = porId.get(l.id);
		if (!p) return [];
		const opciones = p.opciones ?? [];
		const linea = (producto: ProductoPublico, variante: string | null, mayor: number | null) => {
			const m = precioMayorista(producto.precio, mayor);
			return {
				...l,
				clave: claveLinea(l),
				variante,
				producto,
				subtotal: producto.precio * l.cantidad,
				minorista: producto.precio,
				mayorista: minimo && m < producto.precio ? m : null,
				alPorMayor: false
			};
		};
		if (!opciones.length) return l.varianteId ? [] : [linea(p, null, p.mayorista ?? null)];
		const v = (p.variantes ?? []).find((x) => x.id === l.varianteId);
		if (!v) return [];
		const precio = precioDe(v, p.precio);
		return [
			linea(
				{ ...p, precio, anterior: null, foto: v.foto ?? p.foto, stock: v.stock },
				textoVariante(opciones, v.valores),
				mayoristaDeVariante(v.mayorista, p.mayorista, v.precio)
			)
		];
	});
	const unidades = base.reduce((s, l) => s + l.cantidad, 0);
	const monto = base.reduce((s, l) => s + l.subtotal, 0);
	if (!alcanza(minimo, unidades, monto)) return base;
	return base.map((l) =>
		l.mayorista === null
			? l
			: {
					...l,
					producto: { ...l.producto, precio: l.mayorista, anterior: l.minorista },
					subtotal: l.mayorista * l.cantidad,
					alPorMayor: true
				}
	);
}

/**
 * El aviso del carrito: si falta para el precio mayorista (cuánto) o si ya se aplicó, y lo que se
 * ahorra (o se ahorraría con las mismas líneas). null si la tienda no vende por mayor o ninguna
 * línea tiene precio mayorista.
 */
export function avisoMayorista(lineas: LineaDetalle[], minimo: MinimoMayorista | null) {
	if (!minimo || !lineas.some((l) => l.mayorista !== null)) return null;
	const unidades = lineas.reduce((s, l) => s + l.cantidad, 0);
	const monto = lineas.reduce((s, l) => s + l.minorista * l.cantidad, 0);
	const ahorro = lineas.reduce(
		(s, l) => s + (l.mayorista === null ? 0 : (l.minorista - l.mayorista) * l.cantidad),
		0
	);
	const aplicado = lineas.some((l) => l.alPorMayor);
	// Para la barrita (la misma del envío gratis): lo que más falta de los dos mínimos.
	const parte = (hay: number, min: number | null) => (min ? Math.min(1, hay / min) : 1);
	return {
		aplicado,
		falta: falta(minimo, unidades, monto),
		ahorro,
		progreso: aplicado
			? 100
			: Math.round(Math.min(parte(unidades, minimo.unidades), parte(monto, minimo.monto)) * 100)
	};
}
