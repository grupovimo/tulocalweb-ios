<script lang="ts">
	import { tick } from 'svelte';
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { usarTienda } from '../contexto.ts';
	import { categoriaHref } from '../datos.ts';
	import { miles } from '../formato.ts';

	/**
	 * «Categorías» del encabezado en la compu: un desplegable con todas las categorías que tienen
	 * productos (y cuántos), para cuando no entran como pestañas. Se abre con el botón (o con la
	 * flecha abajo) y se cierra con Esc, al tocar afuera o al elegir; el foco vuelve al botón.
	 */
	const ctx = usarTienda();
	const cats = $derived(ctx.tienda.categorias.filter((c) => c.cantidad > 0));
	const uid = $props.id();
	let abierto = $state(false);
	let caja = $state<HTMLElement>();
	let boton = $state<HTMLButtonElement>();

	function cerrar(volver = true) {
		abierto = false;
		if (volver) boton?.focus();
	}
	function teclas(e: KeyboardEvent) {
		if (e.key === 'Escape' && abierto) {
			e.preventDefault();
			cerrar();
		} else if (e.key === 'ArrowDown' && !abierto) {
			e.preventDefault();
			abierto = true;
			void tick().then(() => caja?.querySelector<HTMLAnchorElement>('.tv-cats-menu a')?.focus());
		}
	}
	function afuera(e: MouseEvent) {
		if (abierto && caja && !caja.contains(e.target as Node)) cerrar(false);
	}
</script>

<svelte:document onclick={afuera} />

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
{#if cats.length >= 2}
	<div
		class="tv-cats-desplegable"
		bind:this={caja}
		onkeydown={teclas}
		onfocusout={(e) => {
			if (abierto && caja && !caja.contains(e.relatedTarget as Node)) abierto = false;
		}}
		role="presentation"
	>
		<button
			type="button"
			class="tv-cats-boton"
			bind:this={boton}
			aria-expanded={abierto}
			aria-controls="{uid}-menu"
			onclick={() => (abierto = !abierto)}
			>Categorías<Icon icon={I.phCaretDownBold} size={12} /></button
		>
		{#if abierto}
			<ul class="tv-cats-menu" id="{uid}-menu">
				{#each cats as c (c.nombre)}<li>
						<a href={categoriaHref(ctx.tienda, c.nombre)} onclick={() => cerrar(false)}
							><span>{c.nombre}</span><span class="tv-cats-n">{miles(c.cantidad)}</span></a
						>
					</li>{/each}
				<li class="tv-cats-todo">
					<a href="{ctx.base}/catalogo" onclick={() => cerrar(false)}>Ver todos los productos</a>
				</li>
			</ul>
		{/if}
	</div>
{/if}
