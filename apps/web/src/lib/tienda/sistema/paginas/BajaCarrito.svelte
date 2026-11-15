<script lang="ts">
	import { enhance } from '$app/forms';
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { imagenTienda } from '../../compartir.ts';
	import type { TiendaPublica } from '../../tipos.ts';
	import Tienda from '../marco/Tienda.svelte';
	import { sinViuda } from '../formato.ts';
	import Boton from '../primitivas/Boton.svelte';

	/**
	 * Dejar de recibir los recordatorios del carrito (§62), con el marco de la tienda y las piezas de
	 * la baja del newsletter. Se confirma con un botón: los antivirus de correo abren los links solos.
	 */
	let {
		tienda,
		token,
		form
	}: {
		tienda: TiendaPublica;
		token: string;
		form: { listo?: boolean } | null;
	} = $props();
	const base = $derived(tienda.base ?? `/t/${tienda.slug}`);
	let enviando = $state(false);
</script>

<Seo
	title="No recibir más recordatorios · {tienda.nombre}"
	description="Los recordatorios del carrito de {tienda.nombre}."
	path="/t/{tienda.slug}/carrito/baja/{token}"
	noindex
	imagen={imagenTienda(tienda)}
/>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
<Tienda {tienda} pagina="newsletter">
	<section class="tv-ancho tv-pag tv-confirma">
		<span class="tv-confirma-ico"><Icon icon={I.phArrowCounterClockwiseBold} size={24} /></span>
		<h1 class="tv-h1 tv-pag-titulo">No recibir más recordatorios</h1>
		<p class="tv-lead">
			{sinViuda(
				`Si confirmás, ${tienda.nombre} no te manda más mails por los carritos que dejes sin terminar.`
			)}
		</p>
		{#if form?.listo}
			<div class="tv-confirma-caja tv-tarjeta" role="status">
				<p><b>Listo, no te llegan más</b></p>
				<p class="tv-suave">
					{sinViuda(
						`Tu carrito sigue guardado en este navegador si querés terminar la compra en ${tienda.nombre}.`
					)}
				</p>
			</div>
			<div class="tv-botones tv-confirma-botones">
				<Boton href={base || '/'}>Volver a la tienda</Boton>
			</div>
		{:else}
			<form
				class="tv-botones tv-confirma-botones"
				method="POST"
				use:enhance={() => {
					enviando = true;
					return async ({ update }) => {
						await update({ reset: false });
						enviando = false;
					};
				}}
			>
				<button class="tv-btn tv-btn--primario" type="submit" disabled={enviando}
					>{enviando ? 'Un momento…' : 'No quiero más estos mails'}</button
				>
				<Boton href={base || '/'} tipo="secundario">Volver a la tienda</Boton>
			</form>
		{/if}
	</section>
</Tienda>
