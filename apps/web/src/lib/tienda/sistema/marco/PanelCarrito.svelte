<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { usarTienda } from '../contexto.ts';
	import { cuantos, precio } from '../formato.ts';
	import {
		ivaDe,
		ofreceDe,
		textoCuotas,
		textoTotalSinImpuestos,
		textoTransferencia
	} from '../cobros.ts';
	import { panelCarrito } from '../panel.svelte.ts';
	import Boton from '../primitivas/Boton.svelte';
	import AvisosCarrito from './AvisosCarrito.svelte';
	import LineasCarrito from './LineasCarrito.svelte';

	/**
	 * El carrito como panel al costado (paginas.carrito.variante = 'panel'): se abre al agregar desde
	 * la ficha o al tocar la bolsa del encabezado, sin salir de la página. Las mismas líneas y avisos
	 * que la página; el cupón y el envío quedan para el checkout.
	 */
	let {
		envioGratis = true,
		interceptar = true
	}: {
		envioGratis?: boolean;
		/** La bolsa del encabezado abre el panel (carrito.variante «panel»); si no, va a la página. */
		interceptar?: boolean;
	} = $props();
	$effect(() => {
		panelCarrito.montado = true;
		return () => {
			panelCarrito.montado = false;
		};
	});
	const ctx = usarTienda();
	const t = $derived(ctx.tienda);
	const lineas = $derived(ctx.carrito.detalle(t.productos, t.mayorista ?? null));
	const total = $derived(lineas.reduce((s, l) => s + l.subtotal, 0));
	const unidades = $derived(lineas.reduce((s, l) => s + l.cantidad, 0));
	const titulo = $derived(
		ctx.estilo.gastronomia ? 'Tu pedido' : ctx.dir === 'mostrador' ? 'Tu carrito' : 'Tu bolsa'
	);
	const ofrece = $derived(ofreceDe(t));
	const extra = $derived(
		[textoCuotas(ofrece, total), textoTransferencia(ofrece, total)].filter(Boolean)
	);
	const sinImp = $derived(
		textoTotalSinImpuestos(
			ofrece,
			lineas.map((l) => ({ subtotal: l.subtotal, iva: ivaDe(l.producto) }))
		)
	);

	let dialogo = $state<HTMLDialogElement>();
	$effect(() => {
		if (!dialogo) return;
		if (panelCarrito.abierto && !dialogo.open) dialogo.showModal();
		else if (!panelCarrito.abierto && dialogo.open) dialogo.close();
	});
	// La bolsa del encabezado abre el panel en vez de ir a la página.
	function alTocar(e: MouseEvent) {
		if (!interceptar) return;
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
				{#if sinImp}<p class="tv-chico tv-suave">{sinImp}</p>{/if}
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
