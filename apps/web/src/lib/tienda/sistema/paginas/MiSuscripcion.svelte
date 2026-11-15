<script lang="ts">
	import { enhance } from '$app/forms';
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { accionesDe, textosSuscripcion } from '../../suscripcion.svelte.ts';
	import { CADA, diaCobro } from '../../suscripciones.ts';
	import type { SuscripcionPublica, TiendaPublica } from '../../tipos.ts';
	import { waHref } from '../datos.ts';
	import { precio } from '../formato.ts';
	import Tienda from '../marco/Tienda.svelte';
	import Boton from '../primitivas/Boton.svelte';

	/**
	 * El link privado de una suscripción: en qué estado está, qué incluye, el próximo cobro y dónde
	 * llega; pausar, reanudar y dar de baja (con «¿Seguro?»), o volver a sumarse si se canceló.
	 */
	let {
		tienda,
		suscripcion,
		aviso = null,
		demo = false,
		error = null
	}: {
		tienda: TiendaPublica;
		suscripcion: SuscripcionPublica;
		aviso?: string | null;
		demo?: boolean;
		error?: string | null;
	} = $props();
	const base = $derived(tienda.base ?? `/t/${tienda.slug}`);
	const tx = $derived(textosSuscripcion(suscripcion, aviso === 'alta', demo));
	const puede = $derived(accionesDe(suscripcion));
	const wa = $derived(
		waHref(tienda, `¡Hola! Tengo una consulta sobre mi suscripción a ${suscripcion.plan}.`)
	);
	let cambiando = $state(false);
	let seguro = $state(false);
	const mandar = () => {
		cambiando = true;
		return async ({ update }: { update: () => Promise<void> }) => {
			cambiando = false;
			seguro = false;
			await update();
		};
	};
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
<Tienda {tienda} pagina="suscripciones">
	<section class="tv-ancho tv-pag tv-confirma">
		<span class="tv-confirma-ico"
			><Icon
				icon={suscripcion.estado === 'cancelada'
					? I.phXBold
					: suscripcion.estado === 'pausada'
						? I.phPauseFill
						: I.phCheckBold}
				size={26}
			/></span
		>
		<h1 class="tv-h1 tv-pag-titulo">{tx.titulo}</h1>
		<p class="tv-lead">{tx.texto}</p>
		<div class="tv-confirma-caja tv-tarjeta">
			<p><b>{suscripcion.plan}</b></p>
			<p class="tv-chico tv-suave">
				{precio(suscripcion.precio)}
				{CADA[suscripcion.frecuencia]}{suscripcion.cobros
					? ` · ${suscripcion.cobros} ${suscripcion.cobros === 1 ? 'cobro' : 'cobros'}`
					: ''}
			</p>
			{#if suscripcion.incluye}<p class="tv-chico tv-suave">{suscripcion.incluye}</p>{/if}
			{#if suscripcion.estado === 'activa' && suscripcion.proximoCobro}
				<p class="tv-confirma-donde">
					<Icon icon={I.phCalendarPlusBold} size={17} />Próximo cobro: {diaCobro(
						suscripcion.proximoCobro
					)}
				</p>
			{/if}
			<p class="tv-confirma-donde">
				<Icon icon={I.phMapPinFill} size={17} />{suscripcion.entrega === 'envio'
					? `Te llega a ${suscripcion.direccion}`
					: `Retirás en ${suscripcion.tienda}`}
			</p>
		</div>
		{#if error}<p class="tv-error tv-error--caja" role="alert">{error}</p>{/if}
		<div class="tv-botones tv-confirma-botones">
			{#if puede.pausar || puede.reanudar}
				<form method="POST" action={puede.pausar ? '?/pausar' : '?/reanudar'} use:enhance={mandar}>
					<button class="tv-btn tv-btn--primario" disabled={cambiando}
						>{puede.pausar ? 'Pausar' : 'Reanudar'}</button
					>
				</form>
			{/if}
			{#if puede.cancelar}
				<form method="POST" action="?/cancelar" use:enhance={mandar}>
					{#if seguro}<button class="tv-btn tv-btn--secundario" disabled={cambiando}
							>Sí, darla de baja</button
						>{:else}<button
							type="button"
							class="tv-btn tv-btn--secundario"
							onclick={() => (seguro = true)}>Dar de baja</button
						>{/if}
				</form>
			{:else}
				<Boton href="{base}/suscripciones/{suscripcion.planId}">Volver a sumarme</Boton>
			{/if}
			<Boton href={base || '/'} tipo="secundario">Volver a la tienda</Boton>
		</div>
		{#if wa}<p class="tv-chico tv-suave">
				¿Dudas? Escribinos&nbsp;<a href={wa} target="_blank" rel="noopener">por WhatsApp</a>.
			</p>{/if}
	</section>
</Tienda>
