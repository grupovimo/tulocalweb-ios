<script lang="ts">
	import { usarTienda } from '../contexto.ts';
	import { precio } from '../formato.ts';

	/**
	 * La barra fija de abajo en el celular (Pedido, C-celular): cuántos productos, «Ver pedido» y el
	 * total. Sólo con algo en el carrito; reserva su alto al final de la página (regla 8).
	 *
	 * Uso: `<BarraPedido />` al final de una página de Pedido. El botón flotante de WhatsApp se saca
	 * mientras se ve (Tienda.svelte lo sabe por `ctx.carrito.unidades`).
	 */
	const ctx = usarTienda();
	const lineas = $derived(ctx.carrito.detalle(ctx.tienda.productos, ctx.tienda.mayorista ?? null));
	const unidades = $derived(lineas.reduce((s, l) => s + l.cantidad, 0));
	const total = $derived(lineas.reduce((s, l) => s + l.subtotal, 0));
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
{#if unidades}
	<div class="tv-reserva-barra tv-solo-movil" aria-hidden="true"></div>
	<a class="tv-barra-pedido tv-solo-movil" href="{ctx.base}/carrito"
		><span class="tv-barra-pedido-n">{unidades}</span><span
			>Ver {ctx.estilo.gastronomia ? 'pedido' : 'bolsa'}</span
		><span>{precio(total)}</span></a
	>
{/if}
