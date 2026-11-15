<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { getUi } from '#lib/ui.svelte.ts';
	import { proximamente } from '../../lanzamiento.ts';
	import { tarjetaDe } from '../../seleccion.svelte.ts';
	import type { ProductoPublico } from '../../tipos.ts';
	import { agregar, sumaDirecto } from '../acciones.ts';
	import { usarTienda } from '../contexto.ts';
	import { productoHref } from '../datos.ts';
	import { descuento } from '../formato.ts';
	import Foto from './Foto.svelte';
	import Precio from './Precio.svelte';

	/**
	 * La tarjeta de un producto, con las variantes de cada dirección del handoff:
	 * - Vidriera: foto sin radio, etiqueta clara arriba a la izquierda, nombre 16 y precio con cuotas
	 *   abajo; sin caja (A-moda). En el celular, el «+» cuadrado con borde (A-celular).
	 * - Mostrador: tarjeta blanca con sombra, foto con radio, precio grande con cuotas y precio con
	 *   transferencia en verde, «+» redondo (B-moda, B-cafe).
	 * - Pedido: foto con radio 12, nombre y precio en una línea y el «+» sobre la foto (C-moda, C-cafe).
	 * `forma`: 'grilla' (la de arriba) o 'lista' (fila con foto chica, para listas y el menú).
	 */
	let {
		producto,
		forma = 'grilla',
		ancho = 600,
		prioridad = false
	}: {
		producto: ProductoPublico;
		forma?: 'grilla' | 'lista';
		/** El ancho de la foto que se pide (px). */
		ancho?: number;
		prioridad?: boolean;
	} = $props();

	const ctx = usarTienda();
	const ui = getUi();
	const t = $derived(tarjetaDe(producto));
	const href = $derived(productoHref(ctx.tienda, producto));
	const off = $derived(t.desde ? '' : descuento(t.precio, producto.anterior));
	const etiqueta = $derived(
		proximamente(producto)
			? 'Próximamente'
			: t.agotado
				? 'Sin stock'
				: off || (producto.etiqueta?.trim() ?? '')
	);
	const directo = $derived(sumaDirecto(ctx, producto));
	const detalle = $derived(
		forma === 'lista' ? (producto.descripcion ?? '').split('\n')[0].slice(0, 90) : ''
	);
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
<article class="tv-prod tv-prod--{forma}" class:tv-prod--agotado={t.agotado}>
	<a class="tv-prod-foto" {href} tabindex="-1" aria-hidden="true">
		<Foto
			src={producto.foto}
			alt=""
			nombre={producto.nombre}
			{ancho}
			{prioridad}
			proporcion={forma === 'lista' ? '1/1' : 'var(--t-foto)'}
		/>
		{#if etiqueta && forma === 'grilla'}<span
				class="tv-chip tv-prod-chip"
				class:tv-chip--off={!!off}>{etiqueta}</span
			>{/if}
	</a>
	<div class="tv-prod-info">
		<a class="tv-prod-nombre" {href}>{producto.nombre}</a>
		{#if detalle}<span class="tv-prod-detalle">{detalle}</span>{/if}
		<div class="tv-prod-pie">
			<Precio {producto} linea={ctx.dir === 'pedido' || forma === 'lista'} />
			{#if directo}<button
					type="button"
					class="tv-mas"
					aria-label="Agregar {producto.nombre} al carrito"
					onclick={() => agregar(ctx, producto, ui)}
					><Icon icon={ctx.dir === 'vidriera' ? I.phPlus : I.phPlusBold} size={20} /></button
				>{/if}
		</div>
	</div>
</article>
