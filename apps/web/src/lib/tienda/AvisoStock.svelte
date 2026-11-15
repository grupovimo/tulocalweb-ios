<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { varsTema, type TemaOpiniones } from './opiniones.ts';
	import type { Seleccion } from './seleccion.svelte.ts';
	import type { ProductoPublico, TiendaPublica } from './tipos.ts';

	/**
	 * «Avisame cuando vuelva» (§60) en la ficha, con los colores de cada plantilla (las piezas de
	 * LanzamientoFicha): con el producto o la variante elegida sin stock, quien compra deja su email y,
	 * si quiere, su WhatsApp, y marca que acepta que le escriban. Cuando vuelve el stock le llega un
	 * mail de la tienda, una sola vez.
	 */
	let {
		tienda,
		producto,
		sel,
		tema
	}: {
		tienda: TiendaPublica;
		producto: ProductoPublico;
		sel: Seleccion;
		tema: TemaOpiniones;
	} = $props();

	const uid = $props.id();
	const toca = $derived(
		sel.stock === 0 && !sel.proximo && (!sel.conOpciones || !!sel.variante) && !tienda.vitrina
	);
	let email = $state('');
	let whatsapp = $state('');
	let acepta = $state(false);
	let enviando = $state(false);
	let error = $state<string | null>(null);
	/** Lo ya anotado, por producto o variante: cambiar de talle vuelve a mostrar el formulario. */
	let anotados = $state<Record<string, string>>({});
	const clave = $derived(`${producto.id}:${sel.variante?.id ?? ''}`);

	async function avisar(e: SubmitEvent) {
		e.preventDefault();
		error = null;
		if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim())) return (error = 'Revisá el email.');
		if (whatsapp.trim() && !/^\+?[0-9 ()-]{6,24}$/.test(whatsapp.trim()))
			return (error = 'Revisá el WhatsApp.');
		if (!acepta) return (error = 'Marcá que querés que te escribamos.');
		enviando = true;
		try {
			const r = await fetch(`/t/${tienda.slug}/aviso-stock`, {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({
					productoId: producto.id,
					varianteId: sel.variante?.id ?? null,
					email: email.trim(),
					whatsapp: whatsapp.trim(),
					consentimiento: true
				})
			});
			const b = (await r.json().catch(() => null)) as { error?: string } | null;
			if (r.ok) anotados = { ...anotados, [clave]: email.trim() };
			else error = b?.error ?? 'No se pudo. Probá de nuevo.';
		} catch {
			error = 'No hay conexión. Probá de nuevo en un rato.';
		} finally {
			enviando = false;
		}
	}
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- la página de privacidad es de Tu Local Web, no de la tienda -->
{#if toca}
	<section class="av" style={varsTema(tema)} aria-labelledby="{uid}-t">
		<p class="tit" id="{uid}-t">
			<Icon icon={I.phBellRingingBold} size={17} />Avisame cuando vuelva
		</p>
		{#if anotados[clave]}
			<p class="listo" role="status">
				<Icon icon={I.phCheckBold} size={16} />Listo: te escribimos a {anotados[clave]} cuando haya&nbsp;stock.
			</p>
		{:else}
			<form onsubmit={avisar} novalidate>
				<div class="dos">
					<label class="campo"
						><span>Email</span><input
							id="{uid}-email"
							type="email"
							autocomplete="email"
							maxlength="200"
							placeholder="tu@email.com"
							bind:value={email}
						/></label
					>
					<label class="campo"
						><span>WhatsApp <i>(opcional)</i></span><input
							type="tel"
							autocomplete="tel"
							maxlength="24"
							placeholder="11 5555 1234"
							bind:value={whatsapp}
						/></label
					>
				</div>
				<label class="acepto"
					><input type="checkbox" bind:checked={acepta} /><span
						>Acepto que {tienda.nombre} me escriba sólo para avisarme. Ver
						<a href="/legales/privacidad" target="_blank" rel="noopener">privacidad</a>.</span
					></label
				>
				<button type="submit" class="pri" disabled={enviando}
					>{enviando ? 'Un momento…' : 'Avisame'}</button
				>
			</form>
		{/if}
		{#if error}<p class="error" role="alert">{error}</p>{/if}
	</section>
{/if}

<style>
	.av {
		box-sizing: border-box;
		margin-top: 14px;
		display: flex;
		flex-direction: column;
		gap: 10px;
		padding: 14px;
		border: 1px dashed var(--op-ln);
		border-radius: var(--op-r);
		background: var(--op-sf);
		color: var(--op-tx);
		font-family: var(--op-f);
		text-align: left;
	}
	.av * {
		box-sizing: border-box;
	}
	p {
		margin: 0;
		text-wrap: pretty;
	}
	.tit {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 14.5px;
		font-weight: 700;
	}
	.tit :global(svg),
	.listo :global(svg) {
		flex: none;
	}
	form {
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.dos {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 180px), 1fr));
		gap: 10px;
	}
	.campo {
		display: flex;
		flex-direction: column;
		gap: 6px;
		font-size: 13.5px;
		font-weight: 600;
	}
	.campo i {
		font-style: normal;
		font-weight: 400;
		color: var(--op-mu);
	}
	input:not([type='checkbox']) {
		width: 100%;
		min-height: 44px;
		padding: 11px 12px;
		border: 1px solid var(--op-ln);
		border-radius: var(--op-r);
		background: var(--op-campo);
		color: var(--op-tx);
		font: 400 15px var(--op-f);
	}
	input:focus-visible {
		outline: 2px solid var(--op-ac);
		outline-offset: -2px;
	}
	.acepto {
		display: flex;
		gap: 10px;
		align-items: flex-start;
		min-height: 44px;
		font-size: 13px;
		line-height: 1.45;
		color: var(--op-mu);
		text-wrap: pretty;
	}
	.acepto input {
		flex: none;
		width: 20px;
		height: 20px;
		margin: 0;
		accent-color: var(--op-ac);
	}
	.acepto a {
		color: inherit;
	}
	button {
		min-height: 46px;
		padding: 12px 18px;
		border: 0;
		border-radius: var(--op-rb);
		background: var(--op-ac);
		color: var(--op-on);
		font: 600 14.5px var(--op-f);
		cursor: pointer;
	}
	button:disabled {
		opacity: 0.6;
		cursor: default;
	}
	.listo {
		display: flex;
		gap: 8px;
		align-items: baseline;
		font-size: 14.5px;
		font-weight: 600;
	}
	.error {
		font-size: 13.5px;
		color: var(--op-err);
	}
</style>
