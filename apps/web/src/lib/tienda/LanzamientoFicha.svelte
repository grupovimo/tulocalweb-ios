<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { Conteo } from './conteo.svelte.ts';
	import { cuando, proximamente, senaDe } from './lanzamiento.ts';
	import { varsTema, type TemaOpiniones } from './opiniones.ts';
	import type { Seleccion } from './seleccion.svelte.ts';
	import type { ProductoPublico, TiendaPublica } from './tipos.ts';

	/**
	 * Lo del lanzamiento y la seña en la ficha (§52), con los colores de cada plantilla: antes de la
	 * hora, cuándo sale con su cuenta regresiva, «Avisame cuando salga» y, si la tienda lo ofrece,
	 * «Reservar con seña»; ya lanzado, «Encargar con seña» si el producto lo tiene. La seña y el
	 * precio los vuelve a calcular el servidor.
	 */
	let {
		tienda,
		producto,
		sel,
		tema,
		cuenta = true,
		abierto = $bindable(false)
	}: {
		tienda: TiendaPublica;
		producto: ProductoPublico;
		sel: Seleccion;
		tema: TemaOpiniones;
		/** La cuenta regresiva (Pampa la dibuja con su propio bloque del handoff). */
		cuenta?: boolean;
		/** El formulario de la seña abierto (el «Encargar» del configurador de Quebracho lo abre). */
		abierto?: boolean;
	} = $props();

	const uid = $props.id();
	const conteo = new Conteo(() => tienda);
	const antes = $derived(proximamente(producto, conteo.ahora));
	const pct = $derived(antes ? (producto.reserva ?? null) : (producto.encargo ?? null));
	const total = $derived(sel.precio);
	const sena = $derived(pct ? senaDe(total, pct) : 0);
	const fmt = (n: number) => '$ ' + Math.round(n).toLocaleString('es-AR');

	let email = $state('');
	let avisado = $state<string | null>(null);
	let nombre = $state('');
	let correo = $state('');
	let telefono = $state('');
	let enviando = $state(false);
	let error = $state<string | null>(null);

	async function pedir(ruta: string, cuerpo: Record<string, unknown>) {
		try {
			const r = await fetch(`/t/${tienda.slug}/${ruta}`, {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify(cuerpo)
			});
			const b = (await r.json().catch(() => null)) as { error?: string; url?: string } | null;
			return r.ok ? { url: b?.url ?? null } : { error: b?.error ?? 'No se pudo. Probá de nuevo.' };
		} catch {
			return { error: 'No hay conexión. Probá de nuevo en un rato.' };
		}
	}

	async function avisar(e: SubmitEvent) {
		e.preventDefault();
		if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim())) return (error = 'Revisá el email.');
		enviando = true;
		const r = await pedir('aviso', { productoId: producto.id, email: email.trim() });
		enviando = false;
		error = r.error ?? null;
		if (!r.error) avisado = email.trim();
	}

	function abrir() {
		error = null;
		if (sel.falta) {
			sel.aviso = true;
			error = sel.falta;
			return;
		}
		abierto = true;
	}

	async function senar(e: SubmitEvent) {
		e.preventDefault();
		if (nombre.trim().length < 2) return (error = 'Poné tu nombre.');
		if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(correo.trim())) return (error = 'Revisá el email.');
		enviando = true;
		const r = await pedir('sena', {
			productoId: producto.id,
			varianteId: sel.variante?.id ?? null,
			nombre: nombre.trim(),
			email: correo.trim(),
			telefono: telefono.trim()
		});
		if (r.error || !r.url) {
			enviando = false;
			error = r.error ?? 'No se pudo. Probá de nuevo.';
			return;
		}
		// A Mercado Pago (con la cuenta de la tienda) o, en la demo, a la confirmación.
		window.location.assign(r.url);
	}
</script>

{#if antes || pct}
	<section class="lz" style={varsTema(tema)} aria-labelledby="{uid}-t">
		{#if antes && producto.lanzamiento}
			<p class="sale" id="{uid}-t">
				<Icon icon={I.phTimerBold} size={17} />Sale el {cuando(producto.lanzamiento)}
			</p>
			{#if cuenta}
				<div class="cd" aria-hidden="true">
					{#each conteo.hasta(producto.lanzamiento, 4, ['días', 'hs', 'min', 'seg']) as c (c.u)}
						<div><b>{c.v}</b><span>{c.u}</span></div>
					{/each}
				</div>
			{/if}
			{#if avisado}
				<p class="listo" role="status">
					<Icon icon={I.phCheckBold} size={16} />Listo: te escribimos a {avisado} cuando salga.
				</p>
			{:else}
				<form class="aviso" onsubmit={avisar} novalidate>
					<label for="{uid}-email" class="oculto">Tu email para el aviso</label>
					<input
						id="{uid}-email"
						type="email"
						autocomplete="email"
						maxlength="200"
						placeholder="Tu email"
						bind:value={email}
						disabled={tienda.vitrina}
					/>
					<button type="submit" class="pri" disabled={enviando || tienda.vitrina}
						><Icon icon={I.phBellRingingBold} size={16} />Avisame cuando salga</button
					>
				</form>
			{/if}
		{:else}
			<p class="sale" id="{uid}-t"><Icon icon={I.phHandCoinsBold} size={17} />Encargalo con seña</p>
		{/if}

		{#if pct}
			<div class="sena">
				<p>
					{antes ? 'Reservalo ahora' : 'Pagás'}
					{#if sel.desde}una seña del {pct}%{:else}<b>{fmt(sena)}</b> de seña ({pct}%){/if} con Mercado
					Pago y el resto{#if !sel.desde}&nbsp;(<b>{fmt(total - sena)}</b>){/if} al&nbsp;retirarlo.
				</p>
				{#if !abierto}
					<button
						type="button"
						class="sec"
						onclick={abrir}
						disabled={tienda.vitrina || sel.stock <= 0}
						>{sel.stock <= 0
							? antes
								? 'Sin cupo para reservar'
								: 'Sin stock'
							: antes
								? 'Reservar con seña'
								: 'Encargar con seña'}</button
					>
				{:else}
					<form onsubmit={senar} novalidate>
						<div class="dos">
							<label class="campo"
								><span>Tu nombre</span><input
									bind:value={nombre}
									autocomplete="name"
									maxlength="80"
								/></label
							>
							<label class="campo"
								><span>Email</span><input
									bind:value={correo}
									type="email"
									autocomplete="email"
									maxlength="200"
								/></label
							>
						</div>
						<label class="campo"
							><span>Teléfono <i>(opcional)</i></span><input
								bind:value={telefono}
								type="tel"
								autocomplete="tel"
								maxlength="30"
							/></label
						>
						<div class="acc">
							<button type="submit" class="pri" disabled={enviando}
								>{enviando ? 'Un momento…' : `Pagar la seña de ${fmt(sena)}`}</button
							><button type="button" class="sec" onclick={() => (abierto = false)}>Cancelar</button>
						</div>
					</form>
				{/if}
			</div>
		{/if}
		{#if tienda.vitrina}<p class="nota">Es una tienda de ejemplo: no se reserva de verdad.</p>{/if}
		{#if error}<p class="error" role="alert">{error}</p>{/if}
	</section>
{/if}

<style>
	.lz {
		box-sizing: border-box;
		margin-top: 18px;
		display: flex;
		flex-direction: column;
		gap: 12px;
		color: var(--op-tx);
		font-family: var(--op-f);
		text-align: left;
	}
	.lz * {
		box-sizing: border-box;
	}
	p {
		margin: 0;
		text-wrap: pretty;
	}
	.sale {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 14.5px;
		font-weight: 700;
	}
	.sale :global(svg),
	.listo :global(svg) {
		flex: none;
	}
	.cd {
		display: grid;
		grid-template-columns: repeat(4, minmax(0, 1fr));
		border: 1px solid var(--op-ln);
		border-radius: var(--op-r);
	}
	.cd div {
		padding: 10px 0;
		text-align: center;
		border-right: 1px solid var(--op-ln);
	}
	.cd div:last-child {
		border-right: 0;
	}
	.cd b {
		display: block;
		font-family: var(--op-tf);
		font-weight: var(--op-tw);
		font-size: 26px;
		line-height: 1;
		font-variant-numeric: tabular-nums;
	}
	.cd span {
		display: block;
		margin-top: 4px;
		font-size: 10.5px;
		font-weight: 700;
		letter-spacing: 0.12em;
		text-transform: uppercase;
		color: var(--op-mu);
	}
	.aviso {
		display: flex;
		border: 1px solid var(--op-ln);
		border-radius: var(--op-r);
		overflow: hidden;
	}
	.aviso input {
		flex: 1;
		min-width: 0;
		border: 0;
		border-radius: 0;
	}
	.aviso button {
		border-radius: 0;
	}
	.sena {
		display: flex;
		flex-direction: column;
		gap: 10px;
		padding: 14px;
		border: 1px dashed var(--op-ln);
		border-radius: var(--op-r);
		background: var(--op-sf);
		font-size: 14.5px;
		line-height: 1.5;
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
	input {
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
	button {
		min-height: 46px;
		padding: 12px 18px;
		border-radius: var(--op-rb);
		font: 600 14.5px var(--op-f);
		cursor: pointer;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		white-space: nowrap;
	}
	button:disabled {
		opacity: 0.6;
		cursor: default;
	}
	.pri {
		border: 0;
		background: var(--op-ac);
		color: var(--op-on);
	}
	.sec {
		border: 1px solid var(--op-tx);
		background: transparent;
		color: var(--op-tx);
	}
	.acc {
		display: flex;
		gap: 8px;
		flex-wrap: wrap;
	}
	.listo {
		display: flex;
		gap: 8px;
		align-items: baseline;
		font-size: 14.5px;
		font-weight: 600;
	}
	.nota {
		font-size: 12.5px;
		color: var(--op-mu);
	}
	.error {
		font-size: 13.5px;
		color: var(--op-err);
	}
	.oculto {
		position: absolute;
		width: 1px;
		height: 1px;
		overflow: hidden;
		clip: rect(0 0 0 0);
	}
	@media (max-width: 420px) {
		.aviso {
			flex-direction: column;
		}
		.aviso input {
			border-bottom: 1px solid var(--op-ln);
		}
		.acc {
			flex-direction: column;
		}
		.acc button {
			width: 100%;
		}
	}
</style>
