<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import type { Seleccion } from '../../seleccion.svelte.ts';
	import type { Direccion } from '../catalogo.ts';
	import Foto from './Foto.svelte';

	/**
	 * Las opciones de un producto (talle, color, presentación) para elegir, en la ficha y en el
	 * «Agregar» rápido de las tarjetas (marco/ElegirRapido):
	 * - Color: cada valor conocido («Negro», «Verde oliva», «Rosa viejo») es una muestra redonda de su
	 *   color; uno con foto propia, la foto; los que no se reconocen, la píldora de texto de siempre.
	 *   En Pedido, la píldora lleva el punto de color.
	 * - Lo agotado va tachado y deshabilitado; si falta elegir, la opción se marca.
	 * - `alGuia`: con la guía de talles cargada, el link «Guía de talles» junto al talle.
	 */
	let {
		sel,
		dir,
		alGuia,
		prefijo = 'tv-op'
	}: {
		sel: Seleccion;
		dir: Direccion;
		alGuia?: (() => void) | null;
		/** Para los ids (la ficha y el «Agregar» rápido pueden estar en la misma página). */
		prefijo?: string;
	} = $props();

	/** El talle (o la medida): ahí va la guía. */
	const esTalle = (nombre: string) => /^(talles?|medidas?|tamaños?|size)\b/i.test(nombre.trim());
	const esColor = (nombre: string) => /^colou?r(es)?\b/i.test(nombre.trim());
	const guiaEn = $derived(
		alGuia ? (sel.opciones.find((o) => esTalle(o.nombre))?.nombre ?? null) : null
	);
</script>

{#each sel.opciones as o, i (o.nombre)}
	{@const color = esColor(o.nombre)}
	<!-- Muestras: en Color, las que se reconocen; en otra opción, sólo si todas tienen foto. -->
	{@const muestras = color
		? o.valores.some((x) => x.color || x.foto)
		: o.valores.every((x) => x.foto)}
	{@const falta = sel.aviso && !o.elegido}
	<fieldset
		class="tv-opcion"
		class:tv-opcion--falta={falta}
		id={i === 0 ? `${prefijo}-opciones` : undefined}
	>
		<legend class:tv-opcion-cab={guiaEn === o.nombre}
			><span
				><b>{o.nombre}{o.elegido ? ':' : ''}</b
				>{#if o.elegido}&nbsp;{o.elegido}{:else if falta}&nbsp;· elegí uno{/if}</span
			>{#if guiaEn === o.nombre}<button
					type="button"
					class="tv-link tv-ficha-guia"
					aria-haspopup="dialog"
					onclick={() => alGuia?.()}><Icon icon={I.phRuler} size={16} />Guía de talles</button
				>{/if}</legend
		>
		<div class="tv-opcion-valores" class:tv-opcion-valores--muestras={muestras}>
			{#each o.valores as x (x.valor)}
				{#if muestras && dir !== 'pedido' && (x.color || x.foto)}
					<button
						type="button"
						class="tv-muestra"
						aria-label="{o.nombre} {x.valor}{x.disponible ? '' : ', sin stock'}"
						title={x.valor}
						aria-pressed={x.elegido}
						disabled={!x.disponible && !x.elegido}
						onclick={() => sel.elegir(o.nombre, x.valor)}
						>{#if x.color}<span style:background={x.color}></span>{:else}<Foto
								src={x.foto}
								alt=""
								nombre={x.valor}
								ancho={120}
								proporcion="1/1"
							/>{/if}</button
					>
				{:else}
					<button
						type="button"
						class="tv-valor"
						aria-pressed={x.elegido}
						disabled={!x.disponible && !x.elegido}
						onclick={() => sel.elegir(o.nombre, x.valor)}
						>{#if color && x.color}<span class="tv-valor-punto" style:background={x.color}
							></span>{/if}{x.valor}{#if !x.disponible}<span class="tv-oculto">, sin stock</span
							>{/if}</button
					>
				{/if}
			{/each}
		</div>
	</fieldset>
{/each}
