<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/** Una foto o varias (con orden): se suben como las de los productos. */
	let {
		fotos,
		max,
		etiqueta,
		subir,
		onfotos
	}: {
		fotos: string[];
		max: number;
		etiqueta: string;
		subir: (f: File) => Promise<string>;
		onfotos: (f: string[]) => void;
	} = $props();
	const ui = getUi();
	let subiendo = $state(0);

	async function elegidas(e: Event) {
		const input = e.currentTarget as HTMLInputElement;
		const archivos = [...(input.files ?? [])].slice(0, Math.max(0, max - fotos.length) || 1);
		input.value = '';
		if (!archivos.length) return;
		subiendo = archivos.length;
		try {
			const urls: string[] = [];
			for (const a of archivos) {
				urls.push(await subir(a));
				subiendo--;
			}
			onfotos(max === 1 ? urls.slice(0, 1) : [...fotos, ...urls].slice(0, max));
		} catch (err) {
			ui.toast({
				k: 'err',
				title: 'No pudimos subir la foto',
				desc: err instanceof Error ? err.message : 'Probá de nuevo en un rato.'
			});
		} finally {
			subiendo = 0;
		}
	}
	function mover(i: number) {
		const r = [...fotos];
		[r[i - 1], r[i]] = [r[i], r[i - 1]];
		onfotos(r);
	}
</script>

<div class="fotos" class:una={max === 1}>
	{#each fotos as f, i (f + i)}
		<div class="foto">
			<img src={f} alt="{etiqueta} {i + 1}" />
			<div class="acciones">
				{#if i > 0}<button
						type="button"
						aria-label="Mover la foto {i + 1} antes"
						onclick={() => mover(i)}><Icon icon={I.phArrowLeftBold} size={12} /></button
					>{/if}
				<button
					type="button"
					aria-label="Quitar la foto {i + 1}"
					onclick={() => onfotos(fotos.filter((_, k) => k !== i))}
					><Icon icon={I.phTrash} size={13} /></button
				>
			</div>
		</div>
	{/each}
	{#if fotos.length < max || max === 1}
		<label class="subir" class:ocupado={subiendo > 0}>
			<Icon icon={subiendo ? I.phSpinnerGap : I.phCameraPlusFill} size={18} />
			<span>{subiendo ? 'Subiendo…' : max === 1 && fotos.length ? 'Cambiar' : 'Subir'}</span>
			<input
				type="file"
				accept="image/jpeg,image/png,image/webp"
				class="sr-only"
				aria-label="Subir {etiqueta.toLowerCase()}"
				multiple={max > 1}
				disabled={subiendo > 0}
				onchange={elegidas}
			/>
		</label>
	{/if}
</div>

<style>
	.fotos {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(76px, 1fr));
		gap: 6px;
	}
	.fotos.una {
		grid-template-columns: 96px 96px;
	}
	.foto,
	.subir {
		position: relative;
		aspect-ratio: 1;
		border-radius: 12px;
		overflow: hidden;
	}
	.foto img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		display: block;
	}
	.acciones {
		position: absolute;
		right: 4px;
		bottom: 4px;
		display: flex;
		gap: 4px;
	}
	.acciones button {
		width: 30px;
		height: 30px;
		border: 0;
		border-radius: 999px;
		background: rgba(20, 12, 8, 0.72);
		color: #fff;
		display: grid;
		place-items: center;
		cursor: pointer;
	}
	.subir {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 4px;
		border: 1.5px dashed var(--ln2);
		color: var(--mu);
		font-size: 12.5px;
		font-weight: 600;
		cursor: pointer;
	}
	.subir:focus-within {
		border-color: var(--terracota);
	}
	.subir.ocupado {
		opacity: 0.6;
	}
	@media (max-width: 760px) {
		.acciones button {
			width: 36px;
			height: 36px;
		}
	}
</style>
