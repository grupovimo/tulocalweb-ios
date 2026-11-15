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
	 * Confirmar la suscripción a las novedades o darse de baja (§50), con el marco de la tienda.
	 * Confirma con un botón y no al abrir: los antivirus de correo abren los links solos.
	 */
	let {
		tienda,
		token,
		modo,
		form
	}: {
		tienda: TiendaPublica;
		token: string;
		modo: 'confirmar' | 'baja';
		form: { listo?: boolean; invalido?: boolean } | null;
	} = $props();
	const base = $derived(tienda.base ?? `/t/${tienda.slug}`);
	const titulo = $derived(modo === 'confirmar' ? 'Confirmá tu suscripción' : 'Darte de baja');
	let enviando = $state(false);
</script>

<Seo
	title="{titulo} · {tienda.nombre}"
	description="Las novedades de {tienda.nombre} por email."
	path="/t/{tienda.slug}/newsletter/{modo}/{token}"
	noindex
	imagen={imagenTienda(tienda)}
/>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
<Tienda {tienda} pagina="newsletter">
	<section class="tv-ancho tv-pag tv-confirma">
		<span class="tv-confirma-ico"
			><Icon
				icon={modo === 'confirmar' ? I.phEnvelopeSimple : I.phArrowCounterClockwiseBold}
				size={24}
			/></span
		>
		<h1 class="tv-h1 tv-pag-titulo">{titulo}</h1>
		<p class="tv-lead">
			{sinViuda(
				modo === 'confirmar'
					? `Confirmá tu email y te llegan las novedades de ${tienda.nombre}: lo nuevo, las ofertas y lo que se viene.`
					: `Si te das de baja, no te llegan más las novedades de ${tienda.nombre} por email.`
			)}
		</p>
		{#if form?.listo}
			<div class="tv-confirma-caja tv-tarjeta" role="status">
				{#if modo === 'confirmar'}
					<p><b>¡Listo, ya estás suscripto!</b></p>
					<p class="tv-suave">
						Te van a llegar las novedades de {tienda.nombre}. Cada mail trae el link para darte
						de&nbsp;baja.
					</p>
				{:else}
					<p><b>Listo, te diste de baja</b></p>
					<p class="tv-suave">
						No te llegan más las novedades de {tienda.nombre}. Podés volver a sumarte desde
						la&nbsp;tienda.
					</p>
				{/if}
			</div>
			<div class="tv-botones tv-confirma-botones">
				<Boton href={base || '/'}>Volver a la tienda</Boton>
			</div>
		{:else if form?.invalido}
			<p class="tv-error tv-error--caja" role="alert">
				Ese link ya no sirve: puede que ya te hayas dado de baja. Si querés recibir las novedades,
				sumate de nuevo desde la&nbsp;tienda.
			</p>
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
					>{enviando
						? 'Un momento…'
						: modo === 'confirmar'
							? 'Confirmar mi suscripción'
							: 'Darme de baja'}</button
				>
				<Boton href={base || '/'} tipo="secundario">Volver a la tienda</Boton>
			</form>
		{/if}
	</section>
</Tienda>
