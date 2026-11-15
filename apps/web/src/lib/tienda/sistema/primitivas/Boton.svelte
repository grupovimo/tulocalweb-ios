<script lang="ts">
	import type { Snippet } from 'svelte';

	/**
	 * El botón de la tienda con la forma de la dirección (rectangular, píldora o radio 10) y la forma
	 * elegida en Mi tienda. `primario`: uno por pantalla (regla 5); `secundario`: con borde;
	 * `suave`: tono del fondo; `claro`: claro sobre foto o color; `wa`: el verde de WhatsApp. Con
	 * `href` es un link.
	 */
	let {
		href,
		tipo = 'primario',
		chico = false,
		bloque = false,
		type = 'button',
		disabled = false,
		externo = false,
		label,
		onclick,
		children
	}: {
		href?: string | null;
		tipo?: 'primario' | 'secundario' | 'suave' | 'wa' | 'claro';
		chico?: boolean;
		/** De lado a lado (celular). */
		bloque?: boolean;
		type?: 'button' | 'submit';
		disabled?: boolean;
		/** Abre en otra pestaña (WhatsApp, mapas). */
		externo?: boolean;
		label?: string;
		onclick?: (e: MouseEvent) => void;
		children: Snippet;
	} = $props();
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
{#if href}
	<a
		class="tv-btn tv-btn--{tipo}"
		class:tv-btn--chico={chico}
		class:tv-btn--bloque={bloque}
		{href}
		aria-label={label}
		target={externo ? '_blank' : undefined}
		rel={externo ? 'noopener' : undefined}
		{onclick}>{@render children()}</a
	>
{:else}
	<button
		class="tv-btn tv-btn--{tipo}"
		class:tv-btn--chico={chico}
		class:tv-btn--bloque={bloque}
		{type}
		{disabled}
		aria-label={label}
		{onclick}>{@render children()}</button
	>
{/if}
