<script lang="ts">
	import { enhance } from '$app/forms';
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { plantilla as buscarPlantilla } from '#lib/plantillas.ts';
	import { imagenTienda } from './compartir.ts';
	import type { TiendaPublica } from './tipos.ts';

	/**
	 * Confirmar la suscripción al newsletter o darse de baja (§50). No hay diseño propio: las piezas
	 * del botón de arrepentimiento (formulario de Legales) con los colores y la tipografía de la
	 * plantilla de la tienda.
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
	const p = $derived(buscarPlantilla(tienda.plantilla));
	let enviando = $state(false);
	const titulo = $derived(modo === 'confirmar' ? 'Confirmá tu suscripción' : 'Darte de baja');
</script>

<Seo
	title="{titulo} · {tienda.nombre}"
	description="Las novedades de {tienda.nombre} por email."
	path="/t/{tienda.slug}/newsletter/{modo}/{token}"
	noindex
	imagen={imagenTienda(tienda)}
/>

<div
	class="pagina"
	style="--bg:{p?.bg ?? '#FBF7F2'};--tx:{p?.tx ?? '#2A140C'};--ac:{p?.ac ??
		'#C24E25'};--ff:{p?.ff ?? 'inherit'};--r:{p?.radio ?? '14px'}"
>
	<div class="ancho">
		<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- ruta de la tienda con su slug -->
		<a class="volver" href="/t/{tienda.slug}"
			><Icon icon={I.phArrowLeftBold} size={14} />{tienda.nombre}</a
		>
		<h1>{titulo}</h1>
		<p class="bajada">
			{modo === 'confirmar'
				? `Confirmá tu email y te llegan las novedades de ${tienda.nombre}: lo nuevo, las ofertas y lo que se viene.`
				: `Si te das de baja, no te llegan más las novedades de ${tienda.nombre} por email.`}
		</p>

		<div class="caja">
			{#if form?.listo}
				<div class="exito" role="status">
					<img src="/hornero/exito.webp" alt="" width="88" height="88" />
					<div>
						{#if modo === 'confirmar'}
							<b>¡Listo, ya estás suscripto!</b>
							<p>
								Te van a llegar las novedades de {tienda.nombre}. Cada mail trae el link para darte
								de&nbsp;baja.
							</p>
						{:else}
							<b>Listo, te diste de baja</b>
							<p>
								No te llegan más las novedades de {tienda.nombre}. Podés volver a sumarte desde
								la&nbsp;tienda.
							</p>
						{/if}
					</div>
				</div>
			{:else if form?.invalido}
				<div class="err-form" role="alert">
					Ese link ya no sirve: puede que ya te hayas dado de baja. Si querés recibir las novedades,
					sumate de nuevo desde la&nbsp;tienda.
				</div>
			{:else}
				<form
					method="POST"
					use:enhance={() => {
						enviando = true;
						return async ({ update }) => {
							await update({ reset: false });
							enviando = false;
						};
					}}
				>
					<div class="cab">
						<span class="ico"
							><Icon
								icon={modo === 'confirmar' ? I.phEnvelopeSimple : I.phArrowCounterClockwiseBold}
								size={19}
							/></span
						>
						<b>{modo === 'confirmar' ? 'Novedades por email' : 'Dejar de recibir novedades'}</b>
					</div>
					<button class="enviar" type="submit" disabled={enviando}
						>{enviando
							? 'Un momento…'
							: modo === 'confirmar'
								? 'Confirmar mi suscripción'
								: 'Darme de baja'}</button
					>
				</form>
			{/if}
		</div>
	</div>
</div>

<style>
	.pagina {
		min-height: 100vh;
		background: var(--bg);
		color: var(--tx);
		font-family: var(--ff);
	}
	.ancho {
		max-width: 680px;
		margin: 0 auto;
		padding: 28px 16px 64px;
	}
	.volver {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		color: inherit;
		font-size: 14px;
		font-weight: 600;
		text-decoration: none;
	}
	h1 {
		margin: 24px 0 0;
		font-size: clamp(30px, 6vw, 44px);
		line-height: 1.05;
		letter-spacing: -0.03em;
		text-wrap: balance;
	}
	.bajada {
		margin: 12px 0 0;
		font-size: 16px;
		line-height: 1.5;
		opacity: 0.8;
		text-wrap: pretty;
	}
	.caja {
		margin-top: 28px;
		padding: clamp(18px, 4vw, 28px);
		border-radius: var(--r);
		background: color-mix(in srgb, var(--bg) 60%, #fff);
		box-shadow: 0 0 0 1px color-mix(in srgb, var(--tx) 10%, transparent);
	}
	.cab {
		display: flex;
		align-items: center;
		gap: 10px;
		font-size: 18px;
	}
	.ico {
		width: 36px;
		height: 36px;
		border-radius: 50%;
		display: grid;
		place-items: center;
		background: var(--ac);
		color: #fff;
	}
	.enviar {
		margin-top: 18px;
		width: 100%;
		min-height: 48px;
		border: 0;
		border-radius: min(var(--r), 999px);
		background: var(--tx);
		color: var(--bg);
		font: inherit;
		font-size: 15.5px;
		font-weight: 600;
		cursor: pointer;
	}
	.enviar:disabled {
		opacity: 0.6;
	}
	.err-form {
		padding: 10px 12px;
		border-radius: 10px;
		background: #fbe9e6;
		color: #8c2a1f;
		font-size: 14px;
		line-height: 1.5;
		text-wrap: pretty;
	}
	.exito {
		display: flex;
		gap: 16px;
		align-items: center;
	}
	.exito p {
		margin: 6px 0 0;
		line-height: 1.5;
		text-wrap: pretty;
	}
	@media (max-width: 480px) {
		.exito {
			flex-direction: column;
			align-items: flex-start;
		}
	}
</style>
