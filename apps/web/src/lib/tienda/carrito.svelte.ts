/**
 * El carrito de la tienda online: vive en el navegador (por tienda) y el servidor lo valida al
 * cerrar la compra (precios y stock salen de la base, nunca de acá). Una línea por producto y
 * variante (lineas.ts).
 */
import { browser } from '$app/env';
import { agregado, recordarCompra } from './medicion.ts';
import { claveLinea, detalleLineas, MAX_LINEAS, migrarLineas, type Linea } from './lineas.ts';
import { proximamente } from './lanzamiento.ts';
import type { MinimoMayorista } from './mayorista.ts';
import type { ProductoPublico } from './tipos.ts';

export type { Linea };

export class Carrito {
	lineas = $state<Linea[]>([]);
	#clave: string;
	#slug: string;

	constructor(slug: string) {
		this.#slug = slug;
		this.#clave = `carrito:${slug}`;
		if (!browser) return;
		try {
			this.lineas = migrarLineas(JSON.parse(localStorage.getItem(this.#clave) ?? '[]'));
		} catch {
			this.lineas = [];
		}
	}

	#guardar() {
		try {
			localStorage.setItem(this.#clave, JSON.stringify(this.lineas));
		} catch {
			// Sin almacenamiento (modo privado): el carrito dura lo que la página.
		}
	}

	get unidades() {
		return this.lineas.reduce((s, l) => s + l.cantidad, 0);
	}

	/** Suma un producto (y su variante, si tiene opciones). */
	sumar(id: string, cantidad = 1, tope = 99, varianteId: string | null = null) {
		const clave = claveLinea({ id, varianteId });
		const l = this.lineas.find((x) => claveLinea(x) === clave);
		const antes = l?.cantidad ?? 0;
		if (l) l.cantidad = Math.min(tope, l.cantidad + cantidad);
		else if (this.lineas.length < MAX_LINEAS)
			this.lineas.push({ id, varianteId, cantidad: Math.min(tope, cantidad) });
		this.#guardar();
		// add_to_cart / AddToCart (sólo con medición y consentimiento: lib/tienda/medicion.ts).
		agregado(
			this.#slug,
			id,
			(this.lineas.find((x) => claveLinea(x) === clave)?.cantidad ?? 0) - antes
		);
	}

	/**
	 * Sumar desde una tarjeta o "te puede gustar": un producto con opciones no se suma sin elegir
	 * (devuelve false y la plantilla lleva a la ficha).
	 */
	sumarProducto(p: ProductoPublico, cantidad = 1) {
		// Con opciones se elige en la ficha; un lanzamiento que no salió, también va a la ficha.
		if (p.opciones?.length || proximamente(p)) return false;
		if (p.stock > 0) this.sumar(p.id, cantidad, p.stock);
		return true;
	}

	/** Cantidad de una línea (por su clave: el id del producto, o producto~variante). */
	fijar(clave: string, cantidad: number) {
		this.lineas =
			cantidad > 0
				? this.lineas.map((l) => (claveLinea(l) === clave ? { ...l, cantidad } : l))
				: this.lineas.filter((l) => claveLinea(l) !== clave);
		this.#guardar();
	}

	/** Se vacía cuando el pedido salió: lo comprado queda para medir la compra en el pedido. */
	vaciar() {
		if (this.lineas.length) recordarCompra(this.#slug, $state.snapshot(this.lineas));
		this.lineas = [];
		this.#guardar();
	}

	/** Vuelve a armar un carrito guardado (el link del recordatorio): reemplaza lo que había. */
	restaurar(lineas: Linea[]) {
		this.lineas = migrarLineas(lineas);
		this.#guardar();
	}

	/**
	 * Las líneas con su producto (precio, foto y stock de la variante). Con el mínimo mayorista de la
	 * tienda (§49), las que tienen precio mayorista pasan a ese cuando el carrito llega.
	 */
	detalle(productos: ProductoPublico[], mayorista: MinimoMayorista | null = null) {
		return detalleLineas(this.lineas, productos, mayorista);
	}
}

// eslint-disable-next-line svelte/prefer-svelte-reactivity -- registro de carritos, no es estado
const carritos = new Map<string, Carrito>();
/** Un carrito por tienda (y por pestaña). */
export function carritoDe(slug: string) {
	let c = carritos.get(slug);
	if (!c) carritos.set(slug, (c = new Carrito(slug)));
	return c;
}
