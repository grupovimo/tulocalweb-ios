/**
 * Agregar al carrito desde una tarjeta, el menú o un combo (sin entrar a la ficha). Un producto con
 * opciones, o un lanzamiento que todavía no salió, lleva a la ficha para elegir. En la vitrina no se
 * compra. Avisa con un toast con «Deshacer» (CONVENCIONES › Textos).
 */
import { goto } from '$app/navigation';
import { claveLinea } from '../lineas.ts';
import type { ProductoPublico } from '../tipos.ts';
import type { UiState } from '#lib/ui.svelte.ts';
import type { ContextoTienda } from './contexto.ts';
import { productoHref } from './datos.ts';

/** ¿Se puede sumar con un toque? (sin opciones, con stock, ya lanzado y fuera de la vitrina). */
export function sumaDirecto(ctx: ContextoTienda, p: ProductoPublico): boolean {
	return !ctx.tienda.vitrina && !p.opciones?.length && p.stock > 0 && !p.lanzamiento;
}

export function agregar(ctx: ContextoTienda, p: ProductoPublico, ui?: UiState | null) {
	if (ctx.tienda.vitrina) return;
	const antes = ctx.carrito.lineas.find((l) => l.id === p.id && !l.varianteId)?.cantidad ?? 0;
	// eslint-disable-next-line svelte/no-navigation-without-resolve -- ruta de la tienda con su slug
	if (!ctx.carrito.sumarProducto(p)) return void goto(productoHref(ctx.tienda, p));
	// En Pedido la bolsa está siempre a la vista: no hace falta avisar.
	if (ctx.dir === 'pedido' || !ui) return;
	ui.toast({
		k: 'ok',
		title: `Sumaste ${p.nombre}`,
		action: 'Deshacer',
		onAction: () => ctx.carrito.fijar(claveLinea({ id: p.id, varianteId: null }), antes)
	});
}
