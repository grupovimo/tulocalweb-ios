<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import type { ProductoPublico } from '../../tipos.ts';
	import { compararDe } from '../comparar.svelte.ts';
	import { usarTienda } from '../contexto.ts';

	/**
	 * «Comparar» en un producto (función propia de Andes): lo suma o lo saca del comparador. Con dos o
	 * más, el panel (marco/Comparador) se abre desde la barra de abajo.
	 */
	let { producto, grande = false }: { producto: ProductoPublico; grande?: boolean } = $props();
	const ctx = usarTienda();
	const cmp = $derived(compararDe(ctx.tienda.slug));
	const on = $derived(cmp.tiene(producto.id));
</script>

<button
	type="button"
	class="tv-comparar"
	class:tv-comparar--grande={grande}
	aria-pressed={on}
	aria-label={on ? `Sacar ${producto.nombre} de la comparación` : `Comparar ${producto.nombre}`}
	onclick={() => cmp.alternar(producto.id)}
	><Icon icon={on ? I.phCheckBold : I.phColumnsBold} size={15} />{on
		? 'En la comparación'
		: grande
			? 'Comparar con otros'
			: 'Comparar'}</button
>
