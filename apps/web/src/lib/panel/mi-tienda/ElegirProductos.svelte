<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { unsplash } from '#lib/landing/formato.ts';
	import { pesos } from '#lib/panel/formato.ts';
	import { buscar } from './campos.ts';
	import type { Contexto } from './faltan.ts';

	/**
	 * Elegir productos del catálogo de la tienda con un buscador: uno solo (destacado) o varios en
	 * orden (combo, elegidos a mano), con subir y quitar.
	 */
	let {
		ids,
		productos,
		etiqueta,
		max = 1,
		onids
	}: {
		ids: string[];
		productos: Contexto['productos'];
		etiqueta: string;
		max?: number;
		onids: (ids: string[]) => void;
	} = $props();

	let q = $state('');
	let abierto = $state(false);
	const uid = $props.id();
	const elegidos = $derived(ids.map((id) => productos.find((p) => p.id === id)).filter((p) => !!p));
	const resultados = $derived(buscar(productos, q, ids));
	const lleno = $derived(ids.length >= max);

	function elegir(id: string) {
		onids(max === 1 ? [id] : [...ids, id].slice(0, max));
		q = '';
		abierto = max > 1 && ids.length + 1 < max;
	}
	function mover(i: number) {
		const r = [...ids];
		[r[i - 1], r[i]] = [r[i], r[i - 1]];
		onids(r);
	}
</script>

{#if elegidos.length}
	<ul class="elegidos">
		{#each elegidos as p, i (p.id)}
			<li>
				<span class="ft"
					>{#if p.foto}<img src={unsplash(p.foto, 96, '&h=96')} alt="" />{/if}</span
				>
				<span class="n"><b>{p.nombre}</b><small>{pesos(p.precio)}</small></span>
				{#if max > 1}
					<button
						type="button"
						class="mini-btn icono"
						aria-label="Subir {p.nombre}"
						disabled={i === 0}
						onclick={() => mover(i)}><Icon icon={I.phCaretUpBold} size={12} /></button
					>
				{/if}
				<button
					type="button"
					class="mini-btn icono"
					aria-label="Quitar {p.nombre}"
					onclick={() => onids(ids.filter((x) => x !== p.id))}
					><Icon icon={I.phX} size={13} /></button
				>
			</li>
		{/each}
	</ul>
{/if}
{#if !productos.length}
	<p class="nota">Todavía no cargaste productos.</p>
{:else if !lleno || max === 1}
	<div class="buscador">
		<Icon icon={I.phMagnifyingGlass} size={15} style="color:var(--fa)" />
		<input
			class="input"
			type="search"
			role="combobox"
			aria-expanded={abierto}
			aria-controls="r-{uid}"
			aria-label="Buscar para {etiqueta.toLowerCase()}"
			placeholder={max === 1 && elegidos.length ? 'Cambiar por otro…' : 'Buscá un producto'}
			bind:value={q}
			onfocus={() => (abierto = true)}
			oninput={() => (abierto = true)}
			onkeydown={(e) => {
				if (e.key === 'Escape') abierto = false;
				if (e.key === 'Enter' && resultados[0]) {
					e.preventDefault();
					elegir(resultados[0].id);
				}
			}}
		/>
	</div>
	{#if abierto}
		<ul class="resultados" id="r-{uid}" role="listbox" aria-label="Productos">
			{#each resultados as p (p.id)}
				<li role="option" aria-selected="false">
					<button type="button" onclick={() => elegir(p.id)}>
						<span class="ft"
							>{#if p.foto}<img src={unsplash(p.foto, 96, '&h=96')} alt="" />{/if}</span
						>
						<span class="n"><b>{p.nombre}</b><small>{pesos(p.precio)}</small></span>
						<Icon icon={I.phPlusBold} size={13} />
					</button>
				</li>
			{:else}
				<li class="nota">No hay productos con «{q}».</li>
			{/each}
		</ul>
	{/if}
{/if}
{#if max > 1}<p class="nota">{ids.length} de {max}</p>{/if}

<style>
	.elegidos,
	.resultados {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 4px;
	}
	.elegidos li,
	.resultados button {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 4px 6px 4px 4px;
		border-radius: 12px;
		border: 1px solid var(--ln);
		background: var(--sf);
		min-height: 44px;
	}
	.resultados {
		max-height: 260px;
		overflow: auto;
	}
	.resultados button {
		width: 100%;
		font: inherit;
		color: inherit;
		text-align: left;
		cursor: pointer;
	}
	.resultados button:hover,
	.resultados button:focus-visible {
		border-color: var(--terracota);
	}
	.ft {
		width: 36px;
		height: 36px;
		flex: none;
		border-radius: 8px;
		overflow: hidden;
		background: var(--sf2);
	}
	.ft img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	.n {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		font-size: 13.5px;
		font-weight: 500;
	}
	.n b {
		font-weight: 600;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.n small {
		font-size: 12px;
		color: var(--mu);
	}
	.buscador {
		position: relative;
		display: flex;
		align-items: center;
	}
	.buscador :global(svg) {
		position: absolute;
		left: 12px;
		pointer-events: none;
	}
	.buscador .input {
		padding-left: 34px;
	}
	.nota {
		margin: 0;
		font-size: 12.5px;
		font-weight: 500;
		color: var(--mu);
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
