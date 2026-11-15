<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { usarTienda } from '../contexto.ts';
	import { conTransferencia, cuantos, cuotas, precio } from '../formato.ts';
	import { panelCarrito } from '../panel.svelte.ts';
	import Boton from '../primitivas/Boton.svelte';
	import AvisosCarrito from './AvisosCarrito.svelte';
	import LineasCarrito from './LineasCarrito.svelte';

	/**
	 * El carrito como panel al costado (paginas.carrito.variante = 'panel'): se abre al agregar desde
	 * la ficha o al tocar la bolsa del encabezado, sin salir de la página. Las mismas líneas y avisos
	 * que la página; el cupón y el envío quedan para el checkout.
	 */
	let { envioGratis = true }: { envioGratis?: boolean } = $props();
	const ctx = usarTienda();
	const t = $derived(ctx.tienda);
	const lineas = $derived(ctx.carrito.detalle(t.productos, t.mayorista ?? null));
	const total = $derived(lineas.reduce((s, l) => s + l.subtotal, 0));
	const unidades = $derived(lineas.reduce((s, l) => s + l.cantidad, 0));
	const titulo = $derived(
		ctx.estilo.gastronomia ? 'Tu pedido' : ctx.dir === 'mostrador' ? 'Tu carrito' : 'Tu bolsa'
	);
	const extra = $derived(
		[
			t.cobros.mercadoPago ? cuotas(total, t.cobros.cuotas) : '',
			t.cobros.transferencia > 0 && t.cobros.datosTransferencia
				? `${precio(conTransferencia(total, t.cobros.transferencia))} con transferencia`
				: ''
		].filter(Boolean)
	);

	let dialogo = $state<HTMLDialogElement>();
	$effect(() => {
		if (!dialogo) return;
		if (panelCarrito.abierto && !dialogo.open) dialogo.showModal();
		else if (!panelCarrito.abierto && dialogo.open) dialogo.close();
	});
	// La bolsa del encabezado abre el panel en vez de ir a la página.
	function alTocar(e: MouseEvent) {
		const a = (e.target as HTMLElement).closest?.('a.tv-bolsa');
		if (!a || e.metaKey || e.ctrlKey) return;
		e.preventDefault();
		panelCarrito.abierto = true;
	}
</script>

<svelte:document onclick={alTocar} />

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
<dialog
	bind:this={dialogo}
	class="tv-panel"
	aria-labelledby="tv-panel-titulo"
	onclose={() => (panelCarrito.abierto = false)}
	onclick={(e) => e.target === dialogo && (panelCarrito.abierto = false)}
>
	<div class="tv-panel-caja">
		<div class="tv-panel-cab">
			<h2 id="tv-panel-titulo" class="tv-h3">{titulo}</h2>
			{#if unidades}<span class="tv-chico tv-suave"
					>{cuantos(unidades, 'producto', 'productos')}</span
				>{/if}
			<button
				type="button"
				class="tv-icono"
				aria-label="Cerrar"
				onclick={() => (panelCarrito.abierto = false)}><Icon icon={I.phXBold} size={18} /></button
			>
		</div>
		<div class="tv-panel-cuerpo">
			{#if lineas.length}
				<AvisosCarrito {lineas} {envioGratis} />
				<LineasCarrito {lineas} compacta />
			{:else}
				<p class="tv-suave">Todavía no sumaste nada.</p>
			{/if}
		</div>
		{#if lineas.length}
			<div class="tv-panel-pie">
				<div class="tv-cuentas-total"><span>Subtotal</span><b>{precio(total)}</b></div>
				{#if extra.length}<p class="tv-chico tv-suave">o {extra.join(' · ')}</p>{/if}
				<Boton href="{ctx.base}/checkout" bloque onclick={() => (panelCarrito.abierto = false)}
					>Finalizar compra</Boton
				>
				<a class="tv-link" href="{ctx.base}/carrito" onclick={() => (panelCarrito.abierto = false)}
					>Ver el carrito</a
				>
			</div>
		{/if}
	</div>
</dialog>
