<script lang="ts" module>
	/**
	 * Cuántas columnas y cuántos elementos muestra una grilla (regla 2: grillas llenas). Con 1, 2 o 3
	 * elementos, una columna por elemento (1 → destacado, 2 → mitades, 3 → tercios); con más, las
	 * columnas elegidas, recortado a filas completas. En el celular, `movil` columnas (y si sobra uno,
	 * se esconde: nunca media fila vacía).
	 */
	export function armado(n: number, columnas: number, movil = 2) {
		const cols = Math.max(1, Math.min(6, n <= 3 ? n : columnas));
		const mostrar = n <= cols ? n : Math.floor(n / cols) * cols;
		const m = Math.max(1, Math.min(movil, mostrar));
		return { cols, mostrar, movil: m, sobra: mostrar % m };
	}
</script>

<script lang="ts" generics="T">
	import type { Snippet } from 'svelte';

	let {
		items,
		columnas = 4,
		movil = 2,
		recortar = true,
		clase = '',
		etiqueta,
		children
	}: {
		items: T[];
		/** Las columnas en la compu cuando hay 4 o más (2–6). */
		columnas?: number;
		movil?: number;
		/** false: muestra todos aunque la última fila quede incompleta (listas). */
		recortar?: boolean;
		clase?: string;
		/** Para los lectores de pantalla (una lista de productos, de fotos…). */
		etiqueta?: string;
		children: Snippet<[T, number]>;
	} = $props();

	const a = $derived(armado(items.length, columnas, movil));
	const lista = $derived(recortar ? items.slice(0, a.mostrar) : items);
</script>

<ul
	class="tv-grilla {clase}"
	data-n={Math.min(lista.length, 4)}
	style:--cols={recortar ? a.cols : Math.min(columnas, Math.max(1, lista.length))}
	style:--cols-m={a.movil}
	aria-label={etiqueta}
>
	{#each lista as item, i (i)}
		<li class:tv-sobra={recortar && a.sobra > 0 && i >= lista.length - a.sobra}>
			{@render children(item, i)}
		</li>
	{/each}
</ul>
