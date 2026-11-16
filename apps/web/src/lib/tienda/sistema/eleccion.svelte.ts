/**
 * «Agregar» rápido desde una tarjeta de un producto con opciones (talle, color): en vez de ir a la
 * ficha se abre marco/ElegirRapido.svelte con las opciones. Una sola por pestaña, como el panel del
 * carrito. `desde`: el botón que la abrió, para devolverle el foco al cerrar.
 */
import type { ProductoPublico } from '../tipos.ts';

class Eleccion {
	producto = $state<ProductoPublico | null>(null);
	desde: HTMLElement | null = null;
	abrir(p: ProductoPublico, desde: HTMLElement | null = null) {
		this.desde = desde;
		this.producto = p;
	}
	cerrar() {
		this.producto = null;
		const d = this.desde;
		this.desde = null;
		if (d?.isConnected) d.focus();
	}
}

export const eleccion = new Eleccion();
