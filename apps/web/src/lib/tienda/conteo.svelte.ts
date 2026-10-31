/**
 * La cuenta regresiva de las plantillas (§52): corre cada segundo en el navegador y cuenta sólo algo
 * real de la tienda (un lanzamiento próximo o una oferta con fin, lanzamiento.ts). Cuando llega a
 * cero recarga la tienda: el producto lanzado queda a la venta (o la oferta vuelve al precio de lista).
 * Cada plantilla la dibuja con sus piezas del handoff (`cd`).
 */
import { browser } from '$app/env';
import { invalidate } from '$app/navigation';
import { partes, proximoConteo, textoConteo, type Parte } from './lanzamiento.ts';
import type { TiendaPublica } from './tipos.ts';

export class Conteo {
	ahora = $state(Date.now());
	#tienda: () => Pick<TiendaPublica, 'productos'>;

	constructor(tienda: () => Pick<TiendaPublica, 'productos'>) {
		this.#tienda = tienda;
		if (!browser) return;
		$effect(() => {
			const t = setInterval(() => {
				const antes = this.actual;
				this.ahora = Date.now();
				// Llegó la hora: la tienda se vuelve a pedir (Go ya lo da a la venta).
				if (antes && antes.hasta <= this.ahora) void invalidate('tienda:vidriera');
			}, 1000);
			return () => clearInterval(t);
		});
	}

	/** Lo que se cuenta ahora, o null (sin nada real, la plantilla no la dibuja). */
	get actual() {
		return proximoConteo(this.#tienda(), this.ahora);
	}
	get texto() {
		const c = this.actual;
		return c ? textoConteo(c) : '';
	}
	/** Las partes para los chips de la plantilla. */
	partes(n: 3 | 4 = 4, unidades?: string[]): Parte[] {
		const c = this.actual;
		return c ? partes(c.hasta - this.ahora, n, unidades) : [];
	}
	/** Lo que falta hasta un momento dado (para la ficha de un lanzamiento). */
	hasta(iso: string | null | undefined, n: 3 | 4 = 4, unidades?: string[]): Parte[] {
		if (!iso) return [];
		return partes(Date.parse(iso) - this.ahora, n, unidades);
	}
}
