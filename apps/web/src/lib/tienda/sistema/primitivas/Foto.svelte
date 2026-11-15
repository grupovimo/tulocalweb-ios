<script lang="ts">
	import { inicial } from '../datos.ts';
	import { urlFoto } from '../formato.ts';

	/**
	 * Una foto de producto, categoría o portada. Sin foto (o si no carga), el placeholder de marca:
	 * la inicial y una textura con el acento, nunca un gris (regla 3). `proporcion` es el
	 * aspect-ratio (vacío: llena a su contenedor).
	 */
	let {
		src,
		alt,
		nombre,
		ancho = 800,
		proporcion,
		prioridad = false,
		clase = ''
	}: {
		src: string | null | undefined;
		/** Vacío: decorativa. */
		alt: string;
		/** De qué es (para la inicial del placeholder). */
		nombre: string;
		/** El ancho que se pide al CDN (px). */
		ancho?: number;
		proporcion?: string;
		/** La foto de arriba de todo: no se demora la carga. */
		prioridad?: boolean;
		clase?: string;
	} = $props();

	let fallo = $state(false);
	const url = $derived(fallo ? null : urlFoto(src, ancho));
</script>

{#if url}
	<img
		class="tv-foto {clase}"
		src={url}
		{alt}
		loading={prioridad ? 'eager' : 'lazy'}
		decoding="async"
		fetchpriority={prioridad ? 'high' : undefined}
		style:aspect-ratio={proporcion}
		onerror={() => (fallo = true)}
	/>
{:else}
	<div
		class="tv-foto tv-ph {clase}"
		role={alt ? 'img' : undefined}
		aria-label={alt ? `${alt}, todavía sin foto` : undefined}
		aria-hidden={alt ? undefined : 'true'}
		style:aspect-ratio={proporcion}
	>
		<span aria-hidden="true">{inicial(nombre)}</span>
	</div>
{/if}
