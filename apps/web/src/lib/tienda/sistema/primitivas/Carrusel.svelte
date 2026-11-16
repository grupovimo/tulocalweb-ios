<script lang="ts" generics="T">
	import type { Snippet } from 'svelte';
	import { browser } from '$app/env';
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';

	/**
	 * Una fila que se desliza (scroll-snap) con flechas en la compu, como los carruseles de producto
	 * de las tiendas grandes: `visibles` elementos a la vista a 1440 (uno más a partir de 1800), la
	 * mitad y un poco en el celular (se ve que sigue). Las flechas sólo aparecen si hay más de lo que
	 * entra y se apagan en los extremos; el desplazamiento respeta «Reducir movimiento».
	 */
	let {
		items,
		visibles = 4,
		movil = '46%',
		clase = '',
		etiqueta,
		children
	}: {
		items: T[];
		/** Cuántos se ven a la vez en la compu (2–6). */
		visibles?: number;
		/** El ancho de cada uno en el celular. */
		movil?: string;
		clase?: string;
		etiqueta?: string;
		children: Snippet<[T, number]>;
	} = $props();

	let pista = $state<HTMLUListElement>();
	let alInicio = $state(true);
	let alFinal = $state(false);
	let sobra = $state(false);

	function medir() {
		if (!pista) return;
		const max = pista.scrollWidth - pista.clientWidth;
		sobra = max > 4;
		alInicio = pista.scrollLeft <= 4;
		alFinal = pista.scrollLeft >= max - 4;
	}
	$effect(() => {
		if (!browser || !pista) return;
		void items.length;
		medir();
		const ro = new ResizeObserver(medir);
		ro.observe(pista);
		return () => ro.disconnect();
	});
	function mover(d: number) {
		if (!pista) return;
		const quieto = matchMedia('(prefers-reduced-motion: reduce)').matches;
		pista.scrollBy({ left: d * pista.clientWidth * 0.9, behavior: quieto ? 'auto' : 'smooth' });
	}
</script>

<div class="tv-carrusel-caja {clase}" class:tv-carrusel-caja--flechas={sobra}>
	<ul
		class="tv-carrusel"
		bind:this={pista}
		onscroll={medir}
		style:--vis={Math.max(2, Math.min(6, visibles))}
		style:--vis-m={movil}
		aria-label={etiqueta}
	>
		{#each items as item, i (i)}<li>{@render children(item, i)}</li>{/each}
	</ul>
	{#if sobra}
		<button
			type="button"
			class="tv-carrusel-flecha tv-carrusel-flecha--antes"
			aria-label="Ver los anteriores"
			disabled={alInicio}
			onclick={() => mover(-1)}><Icon icon={I.phCaretLeftBold} size={18} /></button
		>
		<button
			type="button"
			class="tv-carrusel-flecha tv-carrusel-flecha--despues"
			aria-label="Ver más"
			disabled={alFinal}
			onclick={() => mover(1)}><Icon icon={I.phCaretRightBold} size={18} /></button
		>
	{/if}
</div>
