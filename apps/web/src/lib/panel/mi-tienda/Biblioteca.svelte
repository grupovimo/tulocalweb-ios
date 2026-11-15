<script lang="ts">
	import Modal from '#lib/components/Modal.svelte';
	import type { TipoBloque } from '#lib/tienda/sistema/catalogo.ts';
	import Miniatura from './Miniatura.svelte';
	import { porQueNo, type Bloque } from './bloques.ts';
	import { DEFS, GRUPOS } from './catalogo.ts';

	/** La biblioteca de bloques por grupo, con su miniatura (docs/handoff/tienda-bloques/Bloques.dc.html). */
	let {
		open = $bindable(false),
		lista,
		donde,
		onelegir
	}: {
		open?: boolean;
		lista: Bloque[];
		/** Dónde va: «al final», «después de Portada». */
		donde: string;
		onelegir: (tipo: TipoBloque) => void;
	} = $props();
</script>

<Modal bind:open titulo="Agregar un bloque" ancho={760}>
	<div class="b-cab">
		<h2>Agregar un bloque</h2>
		<p>Va {donde}. Toma los colores y la tipografía de tu tienda.</p>
	</div>
	<div class="grupos">
		{#each GRUPOS as g (g.id)}
			<section aria-labelledby="g-{g.id}">
				<h3 id="g-{g.id}">{g.nombre}</h3>
				<div class="bloques">
					{#each DEFS.filter((d) => d.grupo === g.id) as d (d.tipo)}
						{@const no = porQueNo(lista, d.tipo)}
						<button
							type="button"
							class="bloque"
							disabled={!!no}
							title={no ?? undefined}
							onclick={() => {
								open = false;
								onelegir(d.tipo);
							}}
						>
							<Miniatura tipo={d.tipo} variante={d.variantes[0]} alto={78} />
							<b>{d.nombre}</b>
							<span>{no ?? d.descripcion}</span>
						</button>
					{/each}
				</div>
			</section>
		{/each}
	</div>
</Modal>

<style>
	.b-cab h2 {
		margin: 0;
		font-size: 22px;
		letter-spacing: -0.02em;
	}
	.b-cab p {
		margin: 4px 0 12px;
		color: var(--mu);
		font-size: 14px;
		text-wrap: pretty;
	}
	.grupos {
		display: flex;
		flex-direction: column;
		gap: 18px;
		max-height: min(68vh, 640px);
		overflow: auto;
		margin: 0 -4px;
		padding: 0 4px 4px;
	}
	h3 {
		margin: 0 0 8px;
		font-size: 14px;
		font-weight: 600;
		padding-bottom: 6px;
		border-bottom: 1px solid var(--ln2);
	}
	.bloques {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
		gap: 10px;
	}
	.bloque {
		display: flex;
		flex-direction: column;
		gap: 4px;
		padding: 6px;
		border: 1px solid var(--ln);
		border-radius: 14px;
		background: var(--sf);
		color: var(--tx);
		font: inherit;
		text-align: left;
		cursor: pointer;
	}
	.bloque:hover:not(:disabled),
	.bloque:focus-visible {
		border-color: var(--terracota);
	}
	.bloque:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
	.bloque b {
		font-size: 13.5px;
		font-weight: 600;
		padding: 2px 2px 0;
	}
	.bloque span {
		font-size: 12px;
		color: var(--mu);
		padding: 0 2px 2px;
		text-wrap: pretty;
	}
	@media (max-width: 560px) {
		.bloques {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
	}
</style>
