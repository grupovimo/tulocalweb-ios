<script lang="ts">
	import { enhance } from '$app/forms';
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import type { TiendaPublica } from '../../tipos.ts';
	import Tienda from '../marco/Tienda.svelte';
	import Boton from '../primitivas/Boton.svelte';

	/**
	 * El botón de arrepentimiento (Ley 24.240, art. 34; Res. 424/2020) con el marco de la tienda:
	 * revocar una compra hasta 10 días después de recibirla, sin dar motivo, con un código de trámite.
	 */
	let {
		tienda,
		codigo = '',
		form
	}: {
		tienda: TiendaPublica;
		codigo?: string;
		form: { tramite?: string; error?: string; previo?: Record<string, string> } | null;
	} = $props();
	const base = $derived(tienda.base ?? `/t/${tienda.slug}`);
	const previo = $derived(form?.previo ?? {});
	let enviando = $state(false);
</script>

<Tienda {tienda} pagina="arrepentimiento">
	<section class="tv-ancho tv-pag tv-legal">
		<h1 class="tv-h1 tv-pag-titulo">Botón de arrepentimiento</h1>
		<p class="tv-lead">
			Podés revocar tu compra sin costo y sin dar explicaciones, hasta 10 días corridos después de
			recibirla (Ley 24.240, artículo&nbsp;34).
		</p>
		{#if form?.tramite}
			<div class="tv-confirma-caja tv-tarjeta" role="status">
				<p><b>Recibimos tu pedido</b></p>
				<p>
					Tu código de trámite es <b class="tv-confirma-alias">{form.tramite}</b>. Te lo mandamos
					por email.
					{tienda.nombre} se va a comunicar con vos para coordinar la devolución y el&nbsp;reintegro.
				</p>
				<Boton href={base || '/'}>Volver a la tienda</Boton>
			</div>
		{:else}
			<form
				class="tv-legal-form tv-tarjeta"
				method="POST"
				novalidate
				use:enhance={() => {
					enviando = true;
					return async ({ update }) => {
						await update({ reset: false });
						enviando = false;
					};
				}}
			>
				<h2 class="tv-h3">
					<Icon icon={I.phArrowCounterClockwiseBold} size={20} />Revocar mi compra
				</h2>
				{#if form?.error}<p class="tv-error tv-error--caja" role="alert">{form.error}</p>{/if}
				<div class="tv-campos">
					<div class="tv-campo">
						<label for="ar-pedido">Código o número del pedido</label>
						<input
							id="ar-pedido"
							name="pedido"
							autocomplete="off"
							placeholder="Está en el mail de la compra"
							value={previo.pedido ?? codigo}
						/>
					</div>
					<div class="tv-campo">
						<label for="ar-email">Email de la compra</label>
						<input
							id="ar-email"
							name="email"
							type="email"
							autocomplete="email"
							value={previo.email ?? ''}
						/>
					</div>
					<div class="tv-campo">
						<label for="ar-nombre">Nombre y apellido</label>
						<input id="ar-nombre" name="nombre" autocomplete="name" value={previo.nombre ?? ''} />
					</div>
				</div>
				<div class="tv-campo">
					<label for="ar-motivo">Motivo (opcional)</label>
					<textarea
						id="ar-motivo"
						name="motivo"
						rows="3"
						placeholder="Contanos si querés, no es obligatorio">{previo.motivo ?? ''}</textarea
					>
				</div>
				<input class="tv-oculto" name="web" tabindex="-1" autocomplete="off" aria-hidden="true" />
				<button class="tv-btn tv-btn--primario tv-btn--bloque" type="submit" disabled={enviando}
					>{enviando ? 'Enviando…' : 'Enviar pedido de arrepentimiento'}</button
				>
				<p class="tv-chico tv-suave">
					Sin costo y sin dar explicaciones. Te llega el código por&nbsp;email.
				</p>
			</form>
		{/if}
	</section>
</Tienda>
