<script lang="ts">
	import { textoDe, type Bloque } from '../catalogo.ts';
	import { usarTienda } from '../contexto.ts';
	import { localDe } from '../datos.ts';
	import { sinViuda } from '../formato.ts';
	import Boton from '../primitivas/Boton.svelte';

	/**
	 * El local: dirección, horarios, cómo llegar y WhatsApp (A-cafe, B-cafe). `mapa`: los datos a un
	 * lado y un mapa dibujado (sin cargar Google Maps: «Cómo llegar» lo abre); `tarjeta`: todo en una
	 * tarjeta. Sin dirección no se dibuja.
	 */
	let { bloque }: { bloque: Bloque } = $props();
	const ctx = usarTienda();
	const l = $derived(localDe(ctx.tienda, bloque));
	const titulo = $derived(textoDe(bloque, 'titulo') || l.direccion);
	// Debajo del título, lo que el título no dice: la dirección (si el título es otro) y la ciudad,
	// sin repetirla si ya está («Av. San Martín 1240, CABA» y abajo «CABA»).
	const debajo = $derived(
		[
			titulo !== l.direccion ? l.direccion : '',
			l.ciudad && !`${titulo} ${l.direccion}`.includes(l.ciudad) ? l.ciudad : ''
		]
			.filter(Boolean)
			.join(', ')
	);
</script>

<section class="tv-local tv-ancho" data-variante={bloque.variante}>
	<div class="tv-local-datos" class:tv-tarjeta={bloque.variante === 'tarjeta'}>
		<p class="tv-ante">El local</p>
		<h2 class="tv-h2">{sinViuda(titulo)}</h2>
		{#if debajo}<p class="tv-lead">{debajo}</p>{/if}
		{#if l.horarios.length}
			<dl class="tv-local-horarios">
				{#each l.horarios as h, i (i)}<dt>{h.dias}</dt>
					<dd>{h.horas}</dd>{/each}
			</dl>
		{/if}
		<div class="tv-botones">
			{#if l.mapa}<Boton href={l.mapa} externo>Cómo llegar</Boton>{/if}
			{#if l.wa}<Boton href={l.wa} tipo={ctx.dir === 'vidriera' ? 'secundario' : 'wa'} externo
					>WhatsApp</Boton
				>{/if}
		</div>
	</div>
	{#if bloque.variante === 'mapa'}
		<div class="tv-local-mapa" role="img" aria-label="Mapa de {ctx.tienda.nombre}">
			<svg aria-hidden="true" viewBox="0 0 700 420" preserveAspectRatio="xMidYMid slice"
				><g class="tv-mapa-calle" fill="none"
					><path d="M-20 90 L720 140" /><path d="M-20 300 L720 250" /><path
						d="M160 -20 L220 440"
					/><path d="M470 -20 L420 440" /></g
				><g class="tv-mapa-av" fill="none"
					><path d="M-20 200 L720 190" /><path d="M320 -20 L330 440" /></g
				></svg
			>
			<span class="tv-local-pin">{ctx.tienda.nombre}</span>
		</div>
	{/if}
</section>
