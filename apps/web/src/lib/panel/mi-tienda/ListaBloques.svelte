<script lang="ts">
	import { tick } from 'svelte';
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import type { Bloque } from './bloques.ts';
	import { defDe, nombreVariante } from './catalogo.ts';
	import { queContradice, queFalta, type Contexto } from './faltan.ts';
	import { ICONO_BLOQUE } from './iconos.ts';

	/**
	 * La lista de bloques de una página: tocar uno abre su panel; la manija lo arrastra (con el dedo
	 * o el mouse) o lo mueve con las flechas del teclado; entre dos bloques se agrega otro.
	 */
	let {
		lista,
		seleccionado,
		cx,
		onabrir,
		onmover,
		onalternar,
		onagregar
	}: {
		lista: Bloque[];
		seleccionado: string | null;
		cx: Contexto;
		onabrir: (id: string) => void;
		onmover: (de: number, a: number) => void;
		onalternar: (i: number) => void;
		/** Abrir la biblioteca para agregar en esa posición. */
		onagregar: (en: number) => void;
	} = $props();

	let anuncio = $state('');
	let filas: HTMLElement[] = $state([]);
	let arrastre = $state<{ i: number; dy: number; destino: number; alto: number } | null>(null);
	let inicio = { y: 0, centros: [] as number[] };

	const nombre = (b: Bloque) => defDe(b.tipo)?.nombre ?? b.tipo;

	function empezar(e: PointerEvent, i: number) {
		if (e.button !== 0) return;
		const el = e.currentTarget as HTMLElement;
		el.setPointerCapture(e.pointerId);
		const rects = filas.map((f) => f.getBoundingClientRect());
		inicio = { y: e.clientY, centros: rects.map((r) => r.top + r.height / 2) };
		arrastre = { i, dy: 0, destino: i, alto: rects[i].height + 6 };
	}
	function seguir(e: PointerEvent) {
		if (!arrastre) return;
		const dy = e.clientY - inicio.y;
		const centro = inicio.centros[arrastre.i] + dy;
		let destino = 0;
		inicio.centros.forEach((c, k) => {
			if (k !== arrastre!.i && c < centro) destino++;
		});
		arrastre = { ...arrastre, dy, destino };
	}
	function soltar() {
		if (!arrastre) return;
		const { i, destino } = arrastre;
		arrastre = null;
		if (destino !== i) {
			onmover(i, destino);
			anunciar(i, destino);
		}
	}
	/** Cuánto se corre una fila mientras otra pasa por encima. */
	function corrimiento(k: number) {
		if (!arrastre) return 0;
		const { i, destino, dy, alto } = arrastre;
		if (k === i) return dy;
		if (i < destino && k > i && k <= destino) return -alto;
		if (i > destino && k < i && k >= destino) return alto;
		return 0;
	}
	function anunciar(de: number, a: number) {
		anuncio = `${nombre(lista[de])}: puesto ${a + 1} de ${lista.length}.`;
	}
	async function teclado(e: KeyboardEvent, i: number) {
		const a =
			e.key === 'ArrowUp'
				? i - 1
				: e.key === 'ArrowDown'
					? i + 1
					: e.key === 'Home'
						? 0
						: e.key === 'End'
							? lista.length - 1
							: -2;
		if (a === -2) return;
		e.preventDefault();
		if (a < 0 || a >= lista.length || a === i) return;
		const id = lista[i].id;
		anunciar(i, a);
		onmover(i, a);
		await tick();
		document.querySelector<HTMLElement>(`[data-manija="${id}"]`)?.focus();
	}
</script>

<p class="sr-only" aria-live="polite">{anuncio}</p>
<ol class="lista" class:arrastrando={!!arrastre}>
	{#each lista as b, i (b.id)}
		{@const falta = b.oculto ? null : queFalta(b, cx)}
		{@const contradice = b.oculto || falta ? null : queContradice(b.datos, cx)}
		<li>
			<button
				type="button"
				class="insertar"
				aria-label="Agregar un bloque antes de {nombre(b)}"
				onclick={() => onagregar(i)}><Icon icon={I.phPlusBold} size={11} /></button
			>
			<div
				class="fila"
				class:sel={seleccionado === b.id}
				class:oculta={b.oculto}
				class:movida={arrastre?.i === i}
				data-fila={b.id}
				bind:this={filas[i]}
				style:transform={arrastre ? `translateY(${corrimiento(i)}px)` : undefined}
			>
				<button
					type="button"
					class="manija"
					data-manija={b.id}
					aria-label="Mover {nombre(b)} (con las flechas)"
					aria-describedby="ayuda-mover"
					onpointerdown={(e) => empezar(e, i)}
					onpointermove={seguir}
					onpointerup={soltar}
					onpointercancel={() => (arrastre = null)}
					onkeydown={(e) => teclado(e, i)}><Icon icon={I.phDotsSixVerticalBold} size={16} /></button
				>
				<button
					type="button"
					class="abrir"
					onclick={() => onabrir(b.id)}
					aria-current={seleccionado === b.id}
				>
					<span class="ic"><Icon icon={ICONO_BLOQUE[b.tipo]} size={16} /></span>
					<span class="txt">
						<b>{nombre(b)}</b>
						{#if falta}
							<span class="falta"><Icon icon={I.phWarningFill} size={12} />No se ve: {falta}</span>
						{:else if contradice}
							<span class="falta"><Icon icon={I.phWarningFill} size={12} />{contradice}</span>
						{:else}
							<span class="var">{b.oculto ? 'Oculto' : nombreVariante(b.tipo, b.variante)}</span>
						{/if}
					</span>
				</button>
				<button
					type="button"
					class="ojo"
					aria-label="{b.oculto ? 'Mostrar' : 'Ocultar'} {nombre(b)}"
					onclick={() => onalternar(i)}
					><Icon icon={b.oculto ? I.phEyeSlash : I.phEye} size={15} /></button
				>
			</div>
		</li>
	{/each}
</ol>
<span id="ayuda-mover" class="sr-only">Usá las flechas arriba y abajo para cambiarlo de lugar.</span
>
<button type="button" class="agregar" onclick={() => onagregar(lista.length)}
	><Icon icon={I.phPlusBold} size={14} />Agregar bloque</button
>

<style>
	.lista {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
	}
	.fila {
		position: relative;
		display: flex;
		align-items: center;
		gap: 2px;
		padding: 3px;
		border-radius: 14px;
		background: var(--sf);
		border: 1px solid var(--ln);
		transition: transform 0.18s var(--ease);
	}
	.arrastrando .fila {
		user-select: none;
	}
	.fila.movida {
		transition: none;
		z-index: 2;
		box-shadow: var(--shd);
		border-color: var(--terracota);
	}
	.fila.sel {
		border-color: var(--terracota);
		box-shadow: 0 0 0 1px var(--terracota);
	}
	.manija {
		width: 30px;
		align-self: stretch;
		display: grid;
		place-items: center;
		border: 0;
		border-radius: 10px;
		background: none;
		color: var(--fa);
		cursor: grab;
		touch-action: none;
	}
	.arrastrando .manija {
		cursor: grabbing;
	}
	.abrir {
		flex: 1;
		min-width: 0;
		display: flex;
		align-items: center;
		gap: 10px;
		min-height: 44px;
		padding: 4px 6px;
		border: 0;
		border-radius: 10px;
		background: none;
		color: inherit;
		font: inherit;
		text-align: left;
		cursor: pointer;
	}
	.ic {
		width: 30px;
		height: 30px;
		flex: none;
		display: grid;
		place-items: center;
		border-radius: 9px;
		background: color-mix(in srgb, var(--terracota) 14%, transparent);
		color: var(--terracota);
	}
	.txt {
		min-width: 0;
		display: flex;
		flex-direction: column;
	}
	.txt b {
		font-size: 13.5px;
		font-weight: 600;
	}
	.var,
	.falta {
		font-size: 12px;
		color: var(--mu);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.falta {
		display: flex;
		align-items: center;
		gap: 4px;
		color: #9a5b12;
		white-space: normal;
		text-wrap: pretty;
	}
	.falta :global(svg) {
		flex: none;
	}
	.oculta .txt,
	.oculta .ic {
		opacity: 0.45;
	}
	.ojo {
		width: 36px;
		height: 36px;
		flex: none;
		display: grid;
		place-items: center;
		border: 0;
		border-radius: 10px;
		background: none;
		color: var(--mu);
		cursor: pointer;
	}
	.manija:focus-visible,
	.abrir:focus-visible,
	.ojo:focus-visible,
	.insertar:focus-visible {
		outline: 2px solid var(--terracota);
		outline-offset: -2px;
	}
	/* Entre dos bloques: una raya con + que aparece al pasar (siempre a la vista en el celular). */
	.insertar {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 100%;
		height: 14px;
		padding: 0;
		border: 0;
		background: none;
		color: var(--terracota);
		opacity: 0;
		cursor: pointer;
		position: relative;
	}
	li:first-child .insertar {
		height: 8px;
	}
	.insertar::before {
		content: '';
		position: absolute;
		left: 12px;
		right: 12px;
		top: 50%;
		border-top: 1.5px dashed currentColor;
	}
	.insertar :global(svg) {
		position: relative;
		background: var(--sf);
		border-radius: 50%;
		padding: 2px;
		width: 18px;
		height: 18px;
		box-shadow: 0 0 0 1.5px currentColor;
	}
	.insertar:hover,
	.insertar:focus-visible {
		opacity: 1;
	}
	.arrastrando .insertar {
		visibility: hidden;
	}
	@media (hover: none) {
		.insertar {
			opacity: 0.55;
			height: 22px;
		}
		.manija {
			width: 40px;
		}
		.ojo {
			width: 44px;
			height: 44px;
		}
	}
	.agregar {
		margin-top: 10px;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		min-height: 44px;
		border: 1.5px dashed var(--ln2);
		border-radius: 14px;
		background: none;
		color: var(--tx);
		font: inherit;
		font-size: 13.5px;
		font-weight: 600;
		cursor: pointer;
	}
	.agregar:hover {
		border-color: var(--terracota);
	}
</style>
