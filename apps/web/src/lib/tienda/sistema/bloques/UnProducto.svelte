<script lang="ts">
	import { getUi } from '#lib/ui.svelte.ts';
	import type { Fondo } from '../catalogo.ts';
	import type { ProductoPublico } from '../../tipos.ts';
	import { agregar, sumaDirecto } from '../acciones.ts';
	import { usarTienda } from '../contexto.ts';
	import { productoHref } from '../datos.ts';
	import { sinViuda } from '../formato.ts';
	import Boton from '../primitivas/Boton.svelte';
	import Foto from '../primitivas/Foto.svelte';
	import Precio from '../primitivas/Precio.svelte';

	/**
	 * Un producto solo, grande, con su botón de compra: el bloque Destacado y la grilla de Productos
	 * cuando hay uno (regla 2: 1 → destacado; A/B/C-vacia). `partido`: foto y texto lado a lado;
	 * `foto`: la foto de fondo con el texto encima, con velo.
	 */
	let {
		producto,
		ante = '',
		texto = '',
		variante = 'partido',
		fondo = 'normal'
	}: {
		producto: ProductoPublico;
		ante?: string;
		texto?: string;
		variante?: 'partido' | 'foto';
		fondo?: Fondo;
	} = $props();
	const ctx = usarTienda();
	const ui = getUi();
	const href = $derived(productoHref(ctx.tienda, producto));
	const directo = $derived(sumaDirecto(ctx, producto));
	const detalle = $derived(texto || (producto.descripcion ?? '').split('\n')[0].slice(0, 200));
	const accion = $derived(
		ctx.estilo.gastronomia
			? 'Agregar al pedido'
			: ctx.dir === 'mostrador'
				? 'Agregar al carrito'
				: 'Agregar a la bolsa'
	);
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
<section class="tv-uno tv-ancho" data-variante={variante}>
	<div class="tv-uno-caja tv-z" data-fondo={variante === 'foto' ? 'normal' : fondo}>
		<a class="tv-uno-foto" {href} tabindex="-1" aria-hidden="true">
			<Foto src={producto.foto} alt="" nombre={producto.nombre} ancho={1200} />
		</a>
		{#if variante === 'foto'}<div class="tv-velo"></div>{/if}
		<div class="tv-uno-info">
			{#if ante}<p class="tv-ante">{ante}</p>{/if}
			<h2 class="tv-h2"><a {href}>{sinViuda(producto.nombre)}</a></h2>
			{#if detalle}<p class="tv-lead">{detalle}</p>{/if}
			<Precio {producto} tam="grande" />
			<div class="tv-botones">
				{#if directo}<Boton onclick={() => agregar(ctx, producto, ui)}>{accion}</Boton>{:else}<Boton
						{href}>Ver el producto</Boton
					>{/if}
			</div>
		</div>
	</div>
</section>
