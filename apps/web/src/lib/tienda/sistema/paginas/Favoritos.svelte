<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import type { TiendaPublica } from '../../tipos.ts';
	import { favoritosDe } from '../favoritos.svelte.ts';
	import Tienda from '../marco/Tienda.svelte';
	import Boton from '../primitivas/Boton.svelte';
	import Grilla from '../primitivas/Grilla.svelte';
	import TarjetaProducto from '../primitivas/TarjetaProducto.svelte';

	/**
	 * Favoritos (conversion.favoritos): los productos que quien compra guardó con el corazón, en
	 * este dispositivo. Vacía, dice cómo se guardan y ofrece el catálogo.
	 */
	let { tienda }: { tienda: TiendaPublica } = $props();
	const base = $derived(tienda.base ?? `/t/${tienda.slug}`);
	const favoritos = $derived(favoritosDe(tienda.slug));
	const productos = $derived(
		favoritos.ids.map((id) => tienda.productos.find((p) => p.id === id)).filter((p) => !!p)
	);
</script>

<Tienda {tienda} pagina="favoritos">
	<div class="tv-ancho tv-pag tv-favoritos">
		<h1 class="tv-h1 tv-pag-titulo">Favoritos</h1>
		{#if productos.length}
			<p class="tv-suave">Quedan guardados en este dispositivo.</p>
			<Grilla items={productos} columnas={4} fija etiqueta="Favoritos">
				{#snippet children(p)}<TarjetaProducto producto={p} ancho={500} />{/snippet}
			</Grilla>
		{:else}
			<div class="tv-carrito-vacio tv-tarjeta">
				<Icon icon={I.phHeart} size={28} />
				<p class="tv-lead">Tocá el corazón de un producto para guardarlo&nbsp;acá.</p>
				<Boton href="{base}/catalogo">Ver el catálogo</Boton>
			</div>
		{/if}
	</div>
</Tienda>
