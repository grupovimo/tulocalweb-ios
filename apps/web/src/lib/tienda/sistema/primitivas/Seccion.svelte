<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { Fondo } from '../catalogo.ts';
	import { sinViuda } from '../formato.ts';

	/**
	 * Un bloque con su encabezado: antetítulo, título (h2, sin viudas), el link «Ver todo» o lo que
	 * vaya a la derecha (`extra`) y el contenido. Con `fondo` distinto de «normal», el bloque va en una
	 * caja con ese color (y los textos de adentro se recalculan para que se lean: base.css).
	 */
	let {
		titulo = '',
		ante = '',
		link = null,
		fondo = 'normal',
		extra,
		chico = false,
		clase = '',
		children
	}: {
		titulo?: string;
		ante?: string;
		link?: { t: string; href: string } | null;
		fondo?: Fondo;
		extra?: Snippet;
		/** Título un escalón más chico (galería, preguntas). */
		chico?: boolean;
		clase?: string;
		children: Snippet;
	} = $props();
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
<section class="tv-sec tv-ancho {clase}" data-fondo={fondo}>
	<div class="tv-sec-caja">
		{#if titulo || ante || link || extra}
			<div class="tv-cab">
				<div class="tv-cab-textos">
					{#if ante}<p class="tv-ante">{ante}</p>{/if}
					{#if titulo}<h2 class={chico ? 'tv-h2 tv-h2--chico' : 'tv-h2'}>{sinViuda(titulo)}</h2>{/if}
				</div>
				{#if extra}{@render extra()}{:else if link}<a class="tv-link" href={link.href}>{link.t}</a
					>{/if}
			</div>
		{/if}
		{@render children()}
	</div>
</section>
