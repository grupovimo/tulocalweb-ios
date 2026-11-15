<script lang="ts">
	import { enhance } from '$app/forms';
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { cuandoLargo } from '#lib/turnos/formato.ts';
	import type { TiendaPublica, TurnoPublico } from '../../tipos.ts';
	import { waHref } from '../datos.ts';
	import { precio } from '../formato.ts';
	import Tienda from '../marco/Tienda.svelte';
	import Boton from '../primitivas/Boton.svelte';

	/**
	 * El link del turno: confirmado, esperando la seña o cancelado; cuándo, con quién, dónde; agregar
	 * al calendario y cancelar (con «¿Seguro?») hasta la anticipación que pide la tienda.
	 */
	let {
		tienda,
		turno,
		aviso = null,
		error = null
	}: {
		tienda: TiendaPublica;
		turno: TurnoPublico;
		aviso?: string | null;
		error?: string | null;
	} = $props();
	const base = $derived(tienda.base ?? `/t/${tienda.slug}`);
	const cancelado = $derived(turno.estado === 'cancelado');
	const wa = $derived(waHref(tienda, `¡Hola! Tengo un turno el ${cuandoLargo(turno.inicio)}.`));
	let cancelando = $state(false);
	let seguro = $state(false);
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
<Tienda {tienda} pagina="turno">
	<section class="tv-ancho tv-pag tv-confirma">
		<span class="tv-confirma-ico"
			><Icon icon={cancelado ? I.phCalendarX : I.phCheckBold} size={26} /></span
		>
		<h1 class="tv-h1 tv-pag-titulo">
			{cancelado
				? 'Turno cancelado'
				: turno.estado === 'reservado'
					? `Falta la seña, ${turno.nombre}`
					: `¡Listo, ${turno.nombre}!`}
		</h1>
		<p class="tv-lead">
			{#if cancelado}Liberamos el horario{turno.senaPagada
					? '. La seña te la devuelve la tienda por Mercado Pago'
					: ''}. Si querés, reservá&nbsp;otro.{:else if turno.estado === 'reservado'}Te guardamos el
				horario unos minutos mientras pagás la seña con Mercado&nbsp;Pago.{:else}Tu turno está
				confirmado. Te esperamos en
				{turno.tienda}.{/if}
		</p>
		{#if aviso === 'pendiente' || aviso === 'error'}
			<p class="tv-error tv-error--caja" role="status">
				{aviso === 'pendiente'
					? 'Mercado Pago todavía no acreditó la seña: te avisamos cuando entre.'
					: 'El pago de la seña no salió. Podés reservar de nuevo.'}
			</p>
		{/if}
		<div class="tv-confirma-caja tv-tarjeta">
			<p class="tv-confirma-cuando" class:tv-tachado={cancelado}>
				<b>{cuandoLargo(turno.inicio)} con {turno.profesional}</b>
			</p>
			<p class="tv-chico tv-suave">
				{turno.servicio} · {precio(turno.precio)}{turno.sena
					? ` · seña ${precio(turno.sena)}${turno.senaPagada ? ' pagada' : ''}`
					: ' · pagás en el local'}
			</p>
			{#if turno.direccion}<p class="tv-confirma-donde">
					<Icon icon={I.phMapPinFill} size={18} />{turno.direccion}
				</p>{/if}
		</div>
		{#if error}<p class="tv-error tv-error--caja" role="alert">{error}</p>{/if}
		<div class="tv-botones tv-confirma-botones">
			{#if !cancelado}
				<a
					class="tv-btn tv-btn--primario"
					href="{base}/turno/{turno.codigo}/turno.ics"
					download="turno.ics"
					data-sveltekit-reload
					><Icon icon={I.phCalendarPlusBold} size={17} />Agregar al calendario</a
				>
			{/if}
			{#if turno.cancelable}
				<form
					method="POST"
					action="?/cancelar"
					use:enhance={() => {
						cancelando = true;
						return async ({ update }) => {
							cancelando = false;
							seguro = false;
							await update();
						};
					}}
				>
					{#if seguro}<button class="tv-btn tv-btn--secundario" disabled={cancelando}
							>Sí, cancelar el turno</button
						>{:else}<button
							type="button"
							class="tv-btn tv-btn--secundario"
							onclick={() => (seguro = true)}>Cancelar turno</button
						>{/if}
				</form>
			{:else if cancelado}
				<Boton href="{base}/turnos"
					><Icon icon={I.phCalendarPlusBold} size={17} />Reservar otro</Boton
				>
			{/if}
			<Boton href={base || '/'} tipo="secundario">Volver a la tienda</Boton>
		</div>
		{#if !cancelado && !turno.cancelable}
			<p class="tv-chico tv-suave">
				Para cambiarlo o cancelarlo, escribile a la&nbsp;tienda{#if wa}&nbsp;<a
						href={wa}
						target="_blank"
						rel="noopener">por WhatsApp</a
					>{/if}.
			</p>
		{/if}
	</section>
</Tienda>
