<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { campo, textoDe, type Bloque } from '../catalogo.ts';
	import { usarTienda } from '../contexto.ts';
	import { conCompra, resenasDe } from '../datos.ts';
	import { cuantos, promedio } from '../formato.ts';
	import Estrellas from '../primitivas/Estrellas.svelte';
	import Foto from '../primitivas/Foto.svelte';
	import Grilla from '../primitivas/Grilla.svelte';
	import Seccion from '../primitivas/Seccion.svelte';

	/**
	 * Reseñas: sólo las de compra verificada, con el promedio (un decimal) y cuántas son.
	 * - `tarjetas`: hasta 3 en tercios (A-moda: cita en serif con borde; B: tarjeta con sombra).
	 * - `carrusel`: en fila deslizable.
	 * - `cita`: una sola, grande, centrada.
	 */
	let { bloque }: { bloque: Bloque } = $props();
	const ctx = usarTienda();
	const t = $derived(ctx.tienda);
	const lista = $derived(resenasDe(t, campo<number>(bloque, 'cantidad') ?? 3));
	const total = $derived(conCompra(t));
	const titulo = $derived(textoDe(bloque, 'titulo'));
	const fotoDe = (nombre: string) => t.productos.find((p) => p.nombre === nombre)?.foto ?? null;
</script>

{#snippet resena(r: (typeof lista)[number], grande = false)}
	<figure class="tv-resena tv-tarjeta" class:tv-resena--grande={grande}>
		<Estrellas valor={r.estrellas} tam={grande ? 18 : 15} />
		<blockquote>“{r.texto}”</blockquote>
		{#if r.respuesta}<p class="tv-resena-resp"><b>Respuesta de {t.nombre}:</b> {r.respuesta}</p>{/if}
		<figcaption>
			{#if r.producto && fotoDe(r.producto) && !grande}<Foto
					src={fotoDe(r.producto)}
					alt=""
					nombre={r.producto}
					ancho={120}
					proporcion="1/1"
					clase="tv-resena-foto"
				/>{/if}
			<span
				><b>{r.autor}</b>{#if r.producto}&nbsp;compró {r.producto}{:else}&nbsp;· compra verificada{/if}</span
			>
		</figcaption>
	</figure>
{/snippet}

<Seccion {titulo}>
	{#snippet extra()}
		{#if t.promedio && total}
			<p class="tv-resenas-prom">
				<Icon icon={I.phSealCheckFill} size={16} /><b>{promedio(t.promedio)}</b>&nbsp;de 5 ·
				{cuantos(total, 'reseña', 'reseñas')} con compra verificada
			</p>
		{/if}
	{/snippet}
	{#if bloque.variante === 'cita'}
		{@render resena(lista[0], true)}
	{:else if bloque.variante === 'carrusel'}
		<ul class="tv-carrusel tv-carrusel--resenas" aria-label="Reseñas">
			{#each lista as r, i (i)}<li>{@render resena(r)}</li>{/each}
		</ul>
	{:else}
		<Grilla items={lista} columnas={3} movil={1} etiqueta="Reseñas">
			{#snippet children(r)}{@render resena(r)}{/snippet}
		</Grilla>
	{/if}
</Seccion>
