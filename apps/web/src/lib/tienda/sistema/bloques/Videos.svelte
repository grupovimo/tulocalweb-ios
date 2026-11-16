<script lang="ts">
	import { browser } from '$app/env';
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { getUi } from '#lib/ui.svelte.ts';
	import { tarjetaDe } from '../../seleccion.svelte.ts';
	import { agregar, sumaDirecto } from '../acciones.ts';
	import { textoDe, type Bloque } from '../catalogo.ts';
	import { usarTienda } from '../contexto.ts';
	import { productoHref, videosDe, type VideoCorto } from '../datos.ts';
	import { eleccion } from '../eleccion.svelte.ts';
	import { precio } from '../formato.ts';
	import Carrusel from '../primitivas/Carrusel.svelte';
	import Foto from '../primitivas/Foto.svelte';
	import Seccion from '../primitivas/Seccion.svelte';

	/**
	 * Videos: una fila de videos verticales (9:16) con su portada; al tocar uno se abre un visor a
	 * pantalla completa (un <dialog>: Esc, la cruz o tocar afuera lo cierran, el foco vuelve al
	 * video tocado) con el video, las flechas para pasar al siguiente y, si tiene, su producto para
	 * agregar. Los archivos mp4/webm se ven en el reproductor del navegador (con «Reducir
	 * movimiento» no arrancan solos); los reels, TikToks y Shorts, en el de su red. Nada se carga
	 * hasta que se abre el visor.
	 */
	let { bloque }: { bloque: Bloque } = $props();
	const ctx = usarTienda();
	const ui = getUi();
	const items = $derived(videosDe(ctx.tienda, bloque));
	const titulo = $derived(textoDe(bloque, 'titulo'));

	let abierto = $state<number | null>(null);
	let desde: HTMLElement | null = null;
	let dialogo = $state<HTMLDialogElement>();
	const actual = $derived<VideoCorto | null>(abierto === null ? null : (items[abierto] ?? null));
	const quieto = $derived(browser && matchMedia('(prefers-reduced-motion: reduce)').matches);

	function abrir(i: number, e: MouseEvent) {
		desde = e.currentTarget as HTMLElement;
		abierto = i;
	}
	function cerrar() {
		abierto = null;
		if (desde?.isConnected) desde.focus();
		desde = null;
	}
	const pasar = (d: number) => {
		if (abierto === null) return;
		abierto = (abierto + d + items.length) % items.length;
	};
	$effect(() => {
		if (!dialogo) return;
		if (actual && !dialogo.open) dialogo.showModal();
		else if (!actual && dialogo.open) dialogo.close();
	});
	function teclas(e: KeyboardEvent) {
		if (e.key === 'ArrowRight') pasar(1);
		else if (e.key === 'ArrowLeft') pasar(-1);
	}

	function sumar() {
		const p = actual?.producto;
		if (!p) return;
		if (sumaDirecto(ctx, p)) {
			cerrar();
			agregar(ctx, p, ui);
		} else if (p.opciones?.length && !tarjetaDe(p).agotado && !ctx.tienda.vitrina) {
			const volver = desde;
			cerrar();
			eleccion.abrir(p, volver);
		}
	}
	const puedeSumar = (v: VideoCorto) =>
		!!v.producto &&
		!ctx.tienda.vitrina &&
		!tarjetaDe(v.producto).agotado &&
		!v.producto.lanzamiento;
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
<Seccion {titulo}>
	<Carrusel {items} visibles={5} movil="42%" etiqueta={titulo || 'Videos'} clase="tv-videos">
		{#snippet children(v, i)}
			<button
				type="button"
				class="tv-video-corto"
				aria-haspopup="dialog"
				aria-label="Ver el video{v.titulo ? `: ${v.titulo}` : ` ${i + 1}`}"
				onclick={(e) => abrir(i, e)}
			>
				<Foto src={v.portada} alt="" nombre={ctx.tienda.nombre} ancho={420} proporcion="9/16" />
				<span class="tv-video-corto-velo" aria-hidden="true"></span>
				<span class="tv-video-corto-play" aria-hidden="true"
					><Icon icon={I.phPlayFill} size={20} /></span
				>
				{#if v.titulo || v.producto}<span class="tv-video-corto-pie" aria-hidden="true"
						>{#if v.titulo}<b>{v.titulo}</b>{/if}{#if v.producto}<span
								>{precio(tarjetaDe(v.producto).precio)}</span
							>{/if}</span
					>{/if}
			</button>
		{/snippet}
	</Carrusel>
</Seccion>

<dialog
	bind:this={dialogo}
	class="tv-visor"
	aria-label={actual?.titulo || 'Video'}
	onclose={() => abierto !== null && cerrar()}
	onclick={(e) => e.target === dialogo && cerrar()}
	onkeydown={teclas}
>
	{#if actual}
		<div class="tv-visor-caja">
			<div class="tv-visor-video">
				{#key abierto}
					{#if actual.fuente.tipo === 'archivo'}
						<video
							src={actual.fuente.src}
							poster={actual.portada}
							controls
							playsinline
							loop
							autoplay={!quieto}
							preload="metadata"
						></video>
					{:else}
						<iframe
							src={quieto
								? actual.fuente.src.replace('autoplay=1', 'autoplay=0')
								: actual.fuente.src}
							title={actual.titulo || 'Video'}
							allow="autoplay; encrypted-media; picture-in-picture; fullscreen"
							referrerpolicy="strict-origin-when-cross-origin"
						></iframe>
					{/if}
				{/key}
			</div>
			{#if actual.producto}
				{@const p = actual.producto}
				<div class="tv-visor-prod">
					<a
						class="tv-visor-prod-foto"
						href={productoHref(ctx.tienda, p)}
						tabindex="-1"
						aria-hidden="true"
						><Foto src={p.foto} alt="" nombre={p.nombre} ancho={160} proporcion="1/1" /></a
					>
					<a class="tv-visor-prod-info" href={productoHref(ctx.tienda, p)}
						><span>{p.nombre}</span><b>{precio(tarjetaDe(p).precio)}</b></a
					>
					{#if puedeSumar(actual)}<button
							type="button"
							class="tv-btn tv-btn--primario tv-btn--chico"
							onclick={sumar}>{sumaDirecto(ctx, p) ? 'Agregar' : 'Elegir'}</button
						>{/if}
				</div>
			{/if}
			<button type="button" class="tv-visor-cerrar" aria-label="Cerrar" onclick={cerrar}
				><Icon icon={I.phXBold} size={20} /></button
			>
			{#if items.length > 1}
				<button
					type="button"
					class="tv-visor-flecha tv-visor-flecha--antes"
					aria-label="Video anterior"
					onclick={() => pasar(-1)}><Icon icon={I.phCaretLeftBold} size={20} /></button
				>
				<button
					type="button"
					class="tv-visor-flecha tv-visor-flecha--despues"
					aria-label="Video siguiente"
					onclick={() => pasar(1)}><Icon icon={I.phCaretRightBold} size={20} /></button
				>
			{/if}
		</div>
	{/if}
</dialog>
