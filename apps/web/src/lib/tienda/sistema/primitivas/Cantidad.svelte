<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';

	/**
	 * El «− 1 +» de la tienda (carrito, ficha, catálogo de Pedido): botones reales de 44 px en el
	 * celular, con el nombre de lo que suma para los lectores. `lleno`: con el color del botón
	 * principal (B y C-catalogo, el producto que ya está en el pedido).
	 */
	let {
		valor,
		max = 99,
		nombre = '',
		lleno = false,
		alCambiar
	}: {
		valor: number;
		max?: number;
		/** Lo que se cuenta («Bombilla de alpaca»), para «Uno más de …». */
		nombre?: string;
		lleno?: boolean;
		alCambiar: (n: number) => void;
	} = $props();
</script>

<span class="tv-cant" class:tv-cant--lleno={lleno}>
	<button
		type="button"
		aria-label={nombre ? `Uno menos de ${nombre}` : 'Uno menos'}
		onclick={() => alCambiar(valor - 1)}><Icon icon={I.phMinusBold} size={14} /></button
	><span class="tv-cant-n">{valor}</span><button
		type="button"
		aria-label={nombre ? `Uno más de ${nombre}` : 'Uno más'}
		disabled={valor >= max}
		onclick={() => alCambiar(valor + 1)}><Icon icon={I.phPlusBold} size={14} /></button
	>
</span>
