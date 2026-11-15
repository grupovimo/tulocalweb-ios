<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import CampoBloque from './CampoBloque.svelte';
	import Miniatura from './Miniatura.svelte';
	import { conDato, type Bloque } from './bloques.ts';
	import type { Ayuda } from './campos.ts';
	import { defDe } from './catalogo.ts';
	import { queContradice, queFalta } from './faltan.ts';

	/** El panel de un bloque: la variante con miniaturas, sus campos y lo que le falta para verse. */
	let {
		bloque,
		ayuda,
		onbloque,
		onvolver,
		onduplicar,
		onalternar,
		onborrar,
		puedeDuplicar
	}: {
		bloque: Bloque;
		ayuda: Ayuda;
		onbloque: (b: Bloque, confirmar: boolean) => void;
		onvolver: () => void;
		onduplicar: () => void;
		onalternar: () => void;
		onborrar: () => void;
		puedeDuplicar: boolean;
	} = $props();

	const def = $derived(defDe(bloque.tipo)!);
	const falta = $derived(queFalta(bloque, ayuda.cx));
	const contradice = $derived(queContradice(bloque.datos, ayuda.cx));
</script>

<div class="panel-bloque">
	<div class="pb-cab">
		<button type="button" class="volver" onclick={onvolver}
			><Icon icon={I.phCaretLeftBold} size={13} />Bloques</button
		>
		<div class="pb-acciones">
			<button
				type="button"
				class="mini-btn icono"
				aria-label="{bloque.oculto ? 'Mostrar' : 'Ocultar'} {def.nombre}"
				title={bloque.oculto ? 'Mostrar' : 'Ocultar'}
				onclick={onalternar}
				><Icon icon={bloque.oculto ? I.phEyeSlash : I.phEye} size={15} /></button
			>
			<button
				type="button"
				class="mini-btn icono"
				aria-label="Duplicar {def.nombre}"
				title="Duplicar"
				disabled={!puedeDuplicar}
				onclick={onduplicar}><Icon icon={I.phCopy} size={15} /></button
			>
			<button
				type="button"
				class="mini-btn icono"
				aria-label="Borrar {def.nombre}"
				title="Borrar"
				onclick={onborrar}><Icon icon={I.phTrash} size={15} /></button
			>
		</div>
	</div>
	<h2 class="pb-t">{def.nombre}</h2>
	<p class="pb-d">{def.descripcion}</p>
	{#if bloque.oculto}
		<p class="aviso info">
			<Icon icon={I.phEyeSlash} size={15} />Está oculto: no se ve en la tienda.
		</p>
	{:else if falta}
		<p class="aviso" role="status">
			<Icon icon={I.phWarningFill} size={15} />No se ve todavía. {falta}
		</p>
	{/if}
	{#if contradice && !bloque.oculto}
		<p class="aviso" role="status">
			<Icon icon={I.phWarningFill} size={15} />{contradice} Cambiá el texto o tu configuración.
		</p>
	{/if}

	{#if def.variantes.length > 1}
		<div class="sub">Diseño</div>
		<div class="variantes" role="radiogroup" aria-label="Diseño de {def.nombre}">
			{#each def.variantes as v, i (v)}
				<button
					type="button"
					role="radio"
					aria-checked={bloque.variante === v}
					class="variante"
					onclick={() => onbloque({ ...bloque, variante: v }, true)}
				>
					<Miniatura tipo={bloque.tipo} variante={v} alto={64} />
					<span>{def.nombresVariantes[i]}</span>
				</button>
			{/each}
		</div>
	{/if}

	<div class="campos">
		{#each def.campos as c (c.id)}
			<CampoBloque
				campo={c}
				valor={bloque.datos[c.id]}
				{ayuda}
				onvalor={(v, confirmar) => onbloque(conDato(bloque, c.id, v), confirmar)}
			/>
		{/each}
	</div>
</div>

<style>
	.panel-bloque {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.pb-cab {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 8px;
	}
	.volver {
		display: flex;
		align-items: center;
		gap: 6px;
		min-height: 40px;
		padding: 0 10px 0 4px;
		border: 0;
		background: none;
		color: var(--mu);
		font: inherit;
		font-size: 13.5px;
		font-weight: 600;
		cursor: pointer;
	}
	.pb-acciones {
		display: flex;
		gap: 4px;
	}
	.pb-t {
		margin: 4px 4px 0;
		font-size: 19px;
		font-weight: 600;
		letter-spacing: -0.02em;
	}
	.pb-d {
		margin: 0 4px 4px;
		font-size: 13px;
		color: var(--mu);
		text-wrap: pretty;
	}
	.aviso {
		display: flex;
		gap: 8px;
		align-items: flex-start;
		margin: 4px 0;
		padding: 10px 12px;
		border-radius: 12px;
		background: color-mix(in srgb, var(--alerta) 16%, transparent);
		color: var(--tx);
		font-size: 13px;
		font-weight: 500;
		text-wrap: pretty;
	}
	.aviso :global(svg) {
		flex: none;
		margin-top: 1px;
		color: #b7791f;
	}
	.aviso.info {
		background: var(--chip);
	}
	.aviso.info :global(svg) {
		color: var(--mu);
	}
	.sub {
		font-size: 12.5px;
		font-weight: 600;
		color: var(--mu);
		padding: 8px 4px 2px;
	}
	.variantes {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 8px;
	}
	.variante {
		display: flex;
		flex-direction: column;
		gap: 5px;
		padding: 5px;
		border: 2px solid transparent;
		border-radius: 12px;
		background: none;
		color: var(--tx);
		font: inherit;
		font-size: 12.5px;
		font-weight: 600;
		text-align: left;
		cursor: pointer;
	}
	.variante[aria-checked='true'] {
		border-color: var(--terracota);
	}
	.variante span {
		padding: 0 2px;
	}
	.campos {
		display: flex;
		flex-direction: column;
		gap: 14px;
		padding: 10px 4px 4px;
	}
	.mini-btn {
		display: inline-grid;
		place-items: center;
		width: 36px;
		min-height: 36px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: var(--sf);
		color: var(--tx);
		cursor: pointer;
	}
	.mini-btn:disabled {
		opacity: 0.35;
	}
	@media (max-width: 760px) {
		.mini-btn {
			width: 44px;
			min-height: 44px;
		}
	}
</style>
