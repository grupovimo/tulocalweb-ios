<script lang="ts">
	import { browser } from '$app/env';
	import { partes } from '../../lanzamiento.ts';
	import { campo, textoDe, type Bloque, type Fondo, type SeleccionProductos } from '../catalogo.ts';
	import { usarTienda } from '../contexto.ts';
	import { hastaOferta, productosDe } from '../datos.ts';
	import { cuantos } from '../formato.ts';
	import Boton from '../primitivas/Boton.svelte';
	import Grilla from '../primitivas/Grilla.svelte';
	import Seccion from '../primitivas/Seccion.svelte';
	import TarjetaProducto from '../primitivas/TarjetaProducto.svelte';

	/**
	 * Oferta con cuenta regresiva: los productos en oferta y cuánto falta (B-moda «Ofertas que
	 * terminan hoy», chips oscuros con HH:MM:SS). Al llegar a cero se esconde sola. `banda`: sólo el
	 * título, la cuenta y el botón a las ofertas.
	 */
	let { bloque }: { bloque: Bloque } = $props();
	const ctx = usarTienda();
	const hasta = $derived(hastaOferta(bloque) ?? 0);
	let ahora = $state(Date.now());
	$effect(() => {
		if (!browser) return;
		const i = setInterval(() => (ahora = Date.now()), 1000);
		return () => clearInterval(i);
	});
	const queda = $derived(hasta - ahora);
	const chips = $derived(queda > 0 ? partes(queda, queda >= 86_400_000 ? 4 : 3) : []);
	const lista = $derived(productosDe(ctx.tienda, campo<SeleccionProductos>(bloque, 'productos')));
	const ofertas = $derived(
		ctx.tienda.productos.filter((p) => p.anterior && p.anterior > p.precio).length
	);
	const fondo = $derived(campo<Fondo>(bloque, 'fondo') ?? 'superficie');
	const link = $derived(
		ofertas > lista.length
			? {
					t: `Ver las ${cuantos(ofertas, 'oferta', 'ofertas')}`,
					href: `${ctx.base}/catalogo?ofertas=1`
				}
			: null
	);
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
{#snippet cuenta()}
	<p
		class="tv-cuenta"
		role="timer"
		aria-label="Termina en {chips.map((c) => `${c.v} ${c.u}`).join(' ')}"
	>
		{#each chips as c, i (c.u)}{#if i}<span class="tv-cuenta-dos" aria-hidden="true">:</span
				>{/if}<span class="tv-cuenta-chip" aria-hidden="true">{c.v}</span>{/each}
	</p>
{/snippet}

{#if queda > 0}
	{#if bloque.variante === 'banda'}
		<section class="tv-banner tv-ancho" data-variante="banda" data-n="1">
			<div class="tv-banner-caja tv-z" data-fondo={fondo === 'normal' ? 'superficie' : fondo}>
				<div class="tv-banner-textos tv-oferta-cab">
					<h2 class="tv-h2">{textoDe(bloque, 'titulo')}</h2>
					{@render cuenta()}
				</div>
				<Boton href="{ctx.base}/catalogo">Ver las ofertas</Boton>
			</div>
		</section>
	{:else}
		<Seccion titulo={textoDe(bloque, 'titulo')} {fondo} {link} clase="tv-oferta">
			{#snippet extra()}
				<div class="tv-oferta-der">
					{@render cuenta()}
					{#if link}<a class="tv-link" href={link.href}>{link.t}</a>{/if}
				</div>
			{/snippet}
			<Grilla items={lista} columnas={4} etiqueta="Productos en oferta">
				{#snippet children(p)}<TarjetaProducto producto={p} />{/snippet}
			</Grilla>
		</Seccion>
	{/if}
{/if}
