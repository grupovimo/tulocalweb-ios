<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { Seleccion } from '../../seleccion.svelte.ts';
	import { ofreceDe, textoCuotas } from '../cobros.ts';
	import { usarTienda } from '../contexto.ts';
	import { productoHref } from '../datos.ts';
	import { eleccion } from '../eleccion.svelte.ts';
	import { precio, sinViuda } from '../formato.ts';
	import { panelCarrito } from '../panel.svelte.ts';
	import Foto from '../primitivas/Foto.svelte';
	import Opciones from '../primitivas/Opciones.svelte';

	/**
	 * «Agregar» rápido de un producto con opciones, desde su tarjeta (eleccion.svelte.ts): un panel al
	 * costado con la foto, el precio, las opciones (con las muestras de color) y el botón. Al sumar
	 * se cierra y abre el carrito al costado (en Pedido, la bolsa ya está a la vista). Esc, tocar
	 * afuera o la cruz lo cierran y el foco vuelve a la tarjeta.
	 */
	const ctx = usarTienda();
	const p = $derived(eleccion.producto);
	const sel = $derived(p ? new Seleccion(p) : null);
	const ofrece = $derived(ofreceDe(ctx.tienda));
	const verbo = $derived(
		ctx.dir === 'mostrador'
			? 'Agregar al carrito'
			: ctx.estilo.gastronomia
				? 'Agregar al pedido'
				: 'Agregar a la bolsa'
	);
	let dialogo = $state<HTMLDialogElement>();
	$effect(() => {
		if (!dialogo) return;
		if (p && !dialogo.open) dialogo.showModal();
		else if (!p && dialogo.open) dialogo.close();
	});

	function agregar() {
		if (!sel || !sel.agregar(ctx.carrito, 1)) return;
		eleccion.cerrar();
		if (ctx.dir !== 'pedido' && panelCarrito.montado) panelCarrito.abierto = true;
	}
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
<dialog
	bind:this={dialogo}
	class="tv-panel tv-elegir"
	aria-labelledby="tv-elegir-titulo"
	onclose={() => eleccion.producto && eleccion.cerrar()}
	onclick={(e) => e.target === dialogo && eleccion.cerrar()}
>
	{#if p && sel}
		<div class="tv-panel-caja">
			<div class="tv-panel-cab">
				<h2 id="tv-elegir-titulo" class="tv-h3">Elegí tus opciones</h2>
				<button type="button" class="tv-icono" aria-label="Cerrar" onclick={() => eleccion.cerrar()}
					><Icon icon={I.phXBold} size={18} /></button
				>
			</div>
			<div class="tv-panel-cuerpo">
				<div class="tv-elegir-prod">
					<div class="tv-elegir-foto">
						<Foto src={sel.foto} alt="" nombre={p.nombre} ancho={300} proporcion="var(--t-foto)" />
					</div>
					<div class="tv-elegir-info">
						<p class="tv-elegir-nombre">{sinViuda(p.nombre)}</p>
						<p class="tv-elegir-precio">
							<b>{sel.desde ? 'Desde ' : ''}{precio(sel.precio)}</b
							>{#if sel.anterior && sel.anterior > sel.precio}<s
									><span class="tv-oculto">Antes </span>{precio(sel.anterior)}</s
								>{/if}
						</p>
						{#if textoCuotas(ofrece, sel.precio, 'corto')}<p class="tv-chico tv-suave">
								{textoCuotas(ofrece, sel.precio, 'corto')}
							</p>{/if}
					</div>
				</div>
				<Opciones {sel} dir={ctx.dir} prefijo="tv-elegir" />
			</div>
			<div class="tv-panel-pie">
				<button
					type="button"
					class="tv-btn tv-btn--primario tv-btn--bloque"
					disabled={sel.proximo || sel.stock <= 0}
					onclick={agregar}>{sel.boton(verbo)}</button
				>
				<a
					class="tv-link"
					href={productoHref(ctx.tienda, p)}
					onclick={() => (eleccion.producto = null)}>Ver el producto</a
				>
			</div>
		</div>
	{/if}
</dialog>
