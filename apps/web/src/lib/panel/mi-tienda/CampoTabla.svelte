<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { LIMITES } from './catalogo.ts';

	/**
	 * Un campo `tabla` (bloques.json › tiposCampo): filas × columnas de textos cortos, la primera
	 * fila es el encabezado. Siempre queda rectangular, con el encabezado y al menos una fila; lo
	 * vacío se saca al guardar (guardar.ts › tablaParaGuardar).
	 */
	let {
		nombre,
		valor,
		onvalor
	}: {
		nombre: string;
		valor: unknown;
		onvalor: (v: string[][] | undefined, confirmar: boolean) => void;
	} = $props();

	const filas = $derived(
		Array.isArray(valor) && Array.isArray(valor[0])
			? (valor as unknown[][]).map((f) =>
					Array.isArray(f) ? f.map((x) => (typeof x === 'string' ? x : '')) : []
				)
			: []
	);
	const ancho = $derived(filas[0]?.length ?? 0);
	const MAXF = LIMITES.filasTabla;
	const MAXC = LIMITES.columnasTabla;
	const LETRAS = LIMITES.letrasCelda;

	const celda = (s: string) => s.replace(/[<>\r\n]/g, '').slice(0, LETRAS);
	function con(i: number, j: number, s: string) {
		return filas.map((f, a) => (a === i ? f.map((x, b) => (b === j ? celda(s) : x)) : f));
	}
	const vacia = () => Array.from({ length: ancho }, () => '');
</script>

<fieldset class="tabla-campo">
	<legend class="rotulo">{nombre}</legend>
	{#if !ancho}
		<button
			type="button"
			class="agregar"
			onclick={() =>
				onvalor(
					[
						['Talle', ''],
						['', '']
					],
					true
				)}><Icon icon={I.phPlusBold} size={13} />Armar la tabla</button
		>
	{:else}
		<div class="marco">
			<table>
				<tbody>
					{#each filas as f, i (i)}
						<tr class:cab={i === 0}>
							{#each f as x, j (j)}
								<td>
									<input
										class="input"
										maxlength={LETRAS}
										value={x}
										aria-label={i === 0
											? `Encabezado, columna ${j + 1}`
											: `Fila ${i}, columna ${j + 1}`}
										placeholder={i === 0 ? (j === 0 ? 'Talle' : 'Medida') : j === 0 ? 'S' : ''}
										oninput={(e) => onvalor(con(i, j, e.currentTarget.value), false)}
										onchange={(e) => onvalor(con(i, j, e.currentTarget.value.trim()), true)}
									/>
								</td>
							{/each}
							<td class="quitar">
								{#if i > 0}
									<button
										type="button"
										class="mini-btn"
										aria-label="Quitar la fila {i}"
										disabled={filas.length <= 2}
										onclick={() =>
											onvalor(
												filas.filter((_, k) => k !== i),
												true
											)}><Icon icon={I.phTrash} size={13} /></button
									>
								{/if}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
		<p class="ayuda">La primera fila es el encabezado. Hasta {MAXF} filas y {MAXC} columnas.</p>
		<div class="acciones">
			<button
				type="button"
				class="chip"
				aria-label="Sumar una fila"
				disabled={filas.length >= MAXF}
				onclick={() => onvalor([...filas, vacia()], true)}
				><Icon icon={I.phPlusBold} size={12} />Fila</button
			>
			<button
				type="button"
				class="chip"
				aria-label="Sumar una columna"
				disabled={ancho >= MAXC}
				onclick={() =>
					onvalor(
						filas.map((f) => [...f, '']),
						true
					)}><Icon icon={I.phPlusBold} size={12} />Columna</button
			>
			<button
				type="button"
				class="chip"
				aria-label="Quitar la última columna"
				disabled={ancho <= 1}
				onclick={() =>
					onvalor(
						filas.map((f) => f.slice(0, -1)),
						true
					)}><Icon icon={I.phMinusBold} size={12} />Columna</button
			>
			<button type="button" class="chip" onclick={() => onvalor(undefined, true)}
				><Icon icon={I.phTrash} size={12} />Borrar la tabla</button
			>
		</div>
	{/if}
</fieldset>

<style>
	.tabla-campo {
		display: flex;
		flex-direction: column;
		gap: 8px;
		min-width: 0;
		margin: 0;
		padding: 0;
		border: 0;
	}
	.rotulo {
		font-size: 13px;
		font-weight: 600;
		padding: 0;
	}
	.marco {
		overflow-x: auto;
		margin-inline: -2px;
		padding: 2px;
	}
	table {
		border-collapse: separate;
		border-spacing: 4px;
		margin: -4px;
	}
	td {
		padding: 0;
	}
	.input {
		width: 92px;
		min-width: 92px;
		padding-inline: 10px;
		font-size: 14px;
	}
	.cab .input {
		font-weight: 700;
		background: var(--sf2);
	}
	.quitar {
		width: 44px;
	}
	.ayuda {
		margin: 0;
		font-size: 12px;
		color: var(--mu);
	}
	.acciones {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
	}
	.chip,
	.agregar {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 6px;
		min-height: 40px;
		padding: 0 12px;
		border: 1px solid var(--ln2);
		border-radius: 999px;
		background: var(--sf);
		color: var(--tx);
		font: inherit;
		font-size: 13px;
		font-weight: 600;
		cursor: pointer;
	}
	.agregar {
		border-style: dashed;
		border-radius: 12px;
	}
	.chip:disabled,
	.mini-btn:disabled {
		opacity: 0.35;
		cursor: default;
	}
	.mini-btn {
		display: inline-grid;
		place-items: center;
		width: 40px;
		min-height: 40px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: var(--sf);
		color: var(--tx);
		cursor: pointer;
	}
	@media (max-width: 760px) {
		.chip,
		.agregar,
		.mini-btn {
			min-height: 44px;
		}
		.mini-btn {
			width: 44px;
		}
	}
</style>
