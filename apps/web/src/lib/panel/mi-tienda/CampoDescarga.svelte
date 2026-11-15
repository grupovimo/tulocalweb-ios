<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { getUi } from '#lib/ui.svelte.ts';
	import { ACEPTA_ARCHIVO, nombreDeArchivo, subirArchivo } from '#lib/panel/archivos.ts';
	import { URL_HTTPS } from './campos.ts';

	/**
	 * El premio «descarga» de una captación: se sube el archivo (PDF, EPUB o ZIP) o se pega un link
	 * https, y se le pone el nombre que ve quien lo baja. Completa `archivo` y `nombreArchivo`.
	 */
	let {
		archivo,
		nombreArchivo,
		onvalor,
		subir = subirArchivo
	}: {
		archivo: string;
		nombreArchivo: string;
		onvalor: (v: { archivo: string; nombreArchivo: string }, confirmar: boolean) => void;
		subir?: (f: File) => Promise<string>;
	} = $props();
	const ui = getUi();
	let subiendo = $state(false);
	let error = $state('');
	const propio = $derived(/\/f\/archivos\//.test(archivo));
	const extension = $derived(archivo.match(/\.(pdf|epub|zip)$/i)?.[1]?.toLowerCase() ?? '');

	async function elegido(e: Event) {
		const input = e.currentTarget as HTMLInputElement;
		const f = input.files?.[0];
		input.value = '';
		if (!f) return;
		subiendo = true;
		try {
			const url = await subir(f);
			error = '';
			onvalor({ archivo: url, nombreArchivo: nombreArchivo || nombreDeArchivo(f.name) }, true);
		} catch (err) {
			ui.toast({
				k: 'err',
				title: 'No pudimos subir el archivo',
				desc: err instanceof Error ? err.message : 'Probá de nuevo en un rato.'
			});
		} finally {
			subiendo = false;
		}
	}
</script>

<div class="descarga">
	{#if archivo && propio}
		<div class="subido">
			<Icon icon={extension === 'pdf' ? I.phFilePdfFill : I.phFileZipFill} size={22} />
			<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- el archivo subido, en /f/ -->
			<a href={archivo} target="_blank" rel="noopener">{nombreArchivo || 'Tu archivo'}</a>
			<button
				type="button"
				class="icon-btn"
				aria-label="Quitar el archivo"
				onclick={() => onvalor({ archivo: '', nombreArchivo }, true)}
				><Icon icon={I.phTrash} size={15} /></button
			>
		</div>
	{/if}
	<label class="btn btn-secondary btn-sm subir" class:ocupado={subiendo}>
		<Icon icon={subiendo ? I.phSpinnerGap : I.phUploadSimpleBold} size={15} />
		<span
			>{subiendo ? 'Subiendo…' : archivo && propio ? 'Cambiar el archivo' : 'Subir archivo'}</span
		>
		<input
			type="file"
			accept={ACEPTA_ARCHIVO}
			class="sr-only"
			aria-label="Subir el archivo para descargar"
			disabled={subiendo}
			onchange={elegido}
		/>
	</label>
	<span class="pista">PDF, EPUB o ZIP de hasta 20 MB.</span>
	{#if !propio}
		<label class="field">
			<span class="etiqueta">O pegá un link</span>
			<input
				class="input"
				type="url"
				inputmode="url"
				maxlength="300"
				placeholder="https://"
				aria-invalid={!!error}
				value={archivo}
				onchange={(e) => {
					const u = e.currentTarget.value.trim();
					error = !u || URL_HTTPS.test(u) ? '' : 'Tiene que empezar con https://';
					if (!error) onvalor({ archivo: u, nombreArchivo }, true);
				}}
			/>
			{#if error}<span class="error">{error}</span>{/if}
		</label>
	{/if}
	<label class="field">
		<span class="etiqueta">Cómo se llama</span>
		<input
			class="input"
			maxlength="60"
			placeholder="Guía"
			value={nombreArchivo}
			onchange={(e) =>
				onvalor({ archivo, nombreArchivo: e.currentTarget.value.trim().slice(0, 60) }, true)}
		/>
	</label>
</div>

<style>
	.descarga {
		display: grid;
		gap: 8px;
	}
	.subido {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 8px 8px 8px 12px;
		border: 1px solid var(--ln2);
		border-radius: 12px;
		color: var(--terracota);
	}
	.subido a {
		flex: 1;
		min-width: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		color: inherit;
		font-weight: 600;
	}
	.subir {
		justify-self: start;
		cursor: pointer;
	}
	.subir:focus-within {
		outline: 2px solid var(--terracota);
		outline-offset: 2px;
	}
	.subir.ocupado {
		opacity: 0.6;
	}
	.pista {
		font-size: 12.5px;
		color: var(--mu);
	}
	.error {
		font-size: 12.5px;
		font-weight: 500;
		color: var(--error);
	}
	@media (max-width: 760px) {
		.subir {
			min-height: 44px;
		}
	}
</style>
