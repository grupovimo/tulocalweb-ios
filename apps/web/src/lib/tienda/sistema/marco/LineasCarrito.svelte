<script lang="ts">
	import { faltantes } from '../../faltantes.ts';
	import type { LineaDetalle } from '../../lineas.ts';
	import { usarTienda } from '../contexto.ts';
	import { productoHref } from '../datos.ts';
	import { precio } from '../formato.ts';
	import Cantidad from '../primitivas/Cantidad.svelte';
	import Foto from '../primitivas/Foto.svelte';

	/**
	 * Las líneas del carrito (página y panel): foto, nombre, variante, «− n +», el subtotal y
	 * «Quitar». Lo que se agotó o bajó de stock mientras tanto se avisa en la misma línea, en el
	 * momento (faltantes.ts; el layout manda además el toast), con «Ajustar».
	 */
	let { lineas, compacta = false }: { lineas: LineaDetalle[]; compacta?: boolean } = $props();
	const ctx = usarTienda();
	const faltan = $derived(
		new Map(faltantes(ctx.carrito.lineas, ctx.tienda.productos).map((f) => [f.clave, f]))
	);
	// Las que ya no se venden no tienen detalle: se avisan aparte.
	const perdidas = $derived(
		[...faltan.values()].filter((f) => f.quedan === 0 && !lineas.some((l) => l.clave === f.clave))
	);
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
<ul class="tv-lineas" class:tv-lineas--compacta={compacta} aria-label="Productos del carrito">
	{#each lineas as l (l.clave)}
		{@const f = faltan.get(l.clave)}
		<li class="tv-linea" class:tv-linea--falta={!!f}>
			<a
				class="tv-linea-foto"
				href={productoHref(ctx.tienda, l.producto)}
				tabindex="-1"
				aria-hidden="true"
				><Foto
					src={l.producto.foto}
					alt=""
					nombre={l.producto.nombre}
					ancho={240}
					proporcion="4/5"
				/></a
			>
			<div class="tv-linea-info">
				<a class="tv-linea-nombre" href={productoHref(ctx.tienda, l.producto)}
					>{l.producto.nombre}</a
				>
				<span class="tv-chico tv-suave"
					>{l.variante ? `${l.variante} · ` : ''}{precio(l.producto.precio)} c/u{#if l.alPorMayor}&nbsp;·
						por mayor{/if}</span
				>
				{#if f}
					<p class="tv-linea-aviso" role="alert">
						{f.quedan ? `Quedan ${f.quedan}: ajustá la cantidad.` : 'Se agotó: sacalo para seguir.'}
						<button type="button" onclick={() => ctx.carrito.fijar(l.clave, f.quedan)}
							>{f.quedan ? 'Ajustar' : 'Sacar'}</button
						>
					</p>
				{/if}
				<div class="tv-linea-pie">
					<Cantidad
						valor={l.cantidad}
						max={l.producto.stock}
						nombre={l.producto.nombre}
						alCambiar={(n) => ctx.carrito.fijar(l.clave, n)}
					/>
					<b class="tv-linea-total">{precio(l.subtotal)}</b>
				</div>
				<button type="button" class="tv-linea-quitar" onclick={() => ctx.carrito.fijar(l.clave, 0)}
					>Quitar<span class="tv-oculto"> {l.producto.nombre}</span></button
				>
			</div>
		</li>
	{/each}
	{#each perdidas as f (f.clave)}
		<li class="tv-linea tv-linea--falta">
			<p class="tv-linea-aviso" role="alert">
				{f.nombre} ya no se vende.
				<button type="button" onclick={() => ctx.carrito.fijar(f.clave, 0)}>Sacar</button>
			</p>
		</li>
	{/each}
</ul>
