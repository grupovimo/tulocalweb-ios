<script lang="ts">
	import type { Snippet } from 'svelte';
	import { colorDe, colorValido, esOscuro } from './estilo/css.ts';
	import type { TiendaPublica } from './tipos.ts';

	/**
	 * La marca de la cabecera: el logo subido en Mi tienda o, sin logo, el nombre como lo dibuja cada
	 * plantilla (children) o en letras. Con el encabezado oscuro va el logo para fondos oscuros, si
	 * hay; el alto sale de Mi tienda › Estilo (si no, el de la plantilla).
	 */
	let {
		tienda,
		alto = 36,
		children
	}: {
		tienda: Pick<TiendaPublica, 'logo' | 'nombre'> &
			Partial<Pick<TiendaPublica, 'estilo' | 'plantilla'>>;
		alto?: number;
		children?: Snippet;
	} = $props();

	const OSCURAS = ['pampa', 'malbec'];
	const fondo = $derived(
		colorValido(tienda.estilo?.colores?.encabezado) ??
			(tienda.plantilla ? colorDe(tienda.plantilla, tienda.estilo, 'fondo') : null)
	);
	const oscuro = $derived(fondo ? esOscuro(fondo) : OSCURAS.includes(tienda.plantilla ?? ''));
	const src = $derived((oscuro && tienda.estilo?.logoOscuro) || tienda.logo);
	const h = $derived(tienda.estilo?.logoAlto || alto);
</script>

{#if src}<img
		class="logo-tienda"
		{src}
		alt={tienda.nombre}
		style:height="{h}px"
	/>{:else if children}{@render children()}{:else}{tienda.nombre}{/if}

<style>
	.logo-tienda {
		display: block;
		width: auto;
		max-width: min(220px, 46vw);
		object-fit: contain;
	}
</style>
