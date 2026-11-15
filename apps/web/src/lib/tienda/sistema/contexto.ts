/**
 * Lo que comparten todas las piezas del sistema: la tienda, su estilo final y la dirección, el
 * prefijo de los links, el menú y el carrito. Lo pone marco/Tienda.svelte (con getters: sigue a las
 * props) y lo lee cualquier pieza con `usarTienda()`.
 */
import { getContext, setContext } from 'svelte';
import type { Carrito } from '../carrito.svelte.ts';
import type { Favoritos } from './favoritos.svelte.ts';
import type { Conversion } from './promos.ts';
import type { Direccion } from './catalogo.ts';
import type { Tienda } from './datos.ts';
import type { EstiloFinal } from './presets.ts';

export type ItemNav = { t: string; href: string; on?: boolean };

export type ContextoTienda = {
	readonly tienda: Tienda;
	readonly estilo: EstiloFinal;
	readonly dir: Direccion;
	/** '' en el dominio propio; si no, /t/<slug>. */
	readonly base: string;
	/** Las pestañas del encabezado (las de Mi tienda › Menú o las de la dirección). */
	readonly nav: ItemNav[];
	readonly carrito: Carrito;
	/** La página que se está dibujando (inicio, catalogo, producto…). */
	readonly pagina: string;
	/** Las herramientas para vender más (estilo.conversion con sus valores por defecto). */
	readonly conversion: Conversion;
	/** Los favoritos de quien compra (sólo se muestran con conversion.favoritos). */
	readonly favoritos: Favoritos;
};

const CLAVE = Symbol('tienda-sistema');

export const ponerTienda = (c: ContextoTienda) => setContext(CLAVE, c);

export function usarTienda(): ContextoTienda {
	const c = getContext<ContextoTienda | undefined>(CLAVE);
	if (!c) throw new Error('Las piezas del sistema van dentro de marco/Tienda.svelte');
	return c;
}
