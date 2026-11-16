<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { getUi } from '#lib/ui.svelte.ts';
	import { proximamente } from '../../lanzamiento.ts';
	import { tarjetaDe } from '../../seleccion.svelte.ts';
	import type { ProductoPublico } from '../../tipos.ts';
	import { claveLinea } from '../../lineas.ts';
	import { agregar, sumaDirecto } from '../acciones.ts';
	import { usarTienda } from '../contexto.ts';
	import { productoHref } from '../datos.ts';
	import { etiquetasDe } from '../conversion.ts';
	import { conEnvioGratis, ofreceDe } from '../cobros.ts';
	import { eleccion } from '../eleccion.svelte.ts';
	import Cantidad from './Cantidad.svelte';
	import Corazon from './Corazon.svelte';
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
	 * `accion`: 'mas' (el «+») o 'agregar' (el botón «Agregar» del catálogo, A- y B-catalogo); con
	 * 'agregar', si ya está en el carrito se ve el «− n +».
	 *
	 * Agregar rápido: uno sin opciones se suma con un toque y se abre el carrito al costado; uno con
	 * opciones abre la elección (marco/ElegirRapido) sin salir de la página. En Vidriera, en la compu,
	 * el botón aparece sobre la foto al pasar el mouse o con el teclado.
	 * Sobre la foto, abajo: «Envío gratis» si el precio ya alcanza el monto de la tienda, y el
	 * descuento por mayor si la tienda vende por mayor y el producto tiene precio mayorista.
	 */
	let {
		producto,
		forma = 'grilla',
		ancho = 600,
		prioridad = false,
		accion = 'mas'
	}: {
		producto: ProductoPublico;
		forma?: 'grilla' | 'lista';
		/** El ancho de la foto que se pide (px). */
		ancho?: number;
		prioridad?: boolean;
		accion?: 'mas' | 'agregar';
	} = $props();
	const clave = $derived(claveLinea({ id: producto.id, varianteId: null }));

	const ctx = usarTienda();
	const ui = getUi();
	const t = $derived(tarjetaDe(producto));
	const href = $derived(productoHref(ctx.tienda, producto));
	// Las etiquetas automáticas (conversion.etiquetas): la oferta, la propia, Nuevo, Últimas N y Más
	// vendido; en la tarjeta van las dos primeras.
	const etiquetas = $derived(etiquetasDe(producto, ctx.tienda, ctx.conversion).slice(0, 2));
	const directo = $derived(sumaDirecto(ctx, producto));
	const enCarrito = $derived(
		accion === 'agregar'
			? (ctx.carrito.lineas.find((l) => l.id === producto.id && !l.varianteId)?.cantidad ?? 0)
			: 0
	);
	// Con opciones: se elige en el panel rápido (no en la vitrina, ni agotado, ni un lanzamiento).
	const elige = $derived(
		!ctx.tienda.vitrina && !!producto.opciones?.length && !t.agotado && !proximamente(producto)
	);
	const ofrece = $derived(ofreceDe(ctx.tienda));
	const gratis = $derived(!t.agotado && conEnvioGratis(ofrece, t.precio));
	const porMayor = $derived.by(() => {
		const m = ctx.tienda.mayorista;
		if (!m || (m.unidades === null && m.monto === null)) return '';
		const pm = producto.mayorista;
		if (!pm || pm >= t.precio) return '';
		const pct = Math.round((1 - pm / t.precio) * 100);
		return pct >= 1 ? `−${pct}\u00a0% por mayor` : '';
	});
	const sellos = $derived([gratis ? 'Envío gratis' : '', porMayor].filter(Boolean));
	function rapido(e: MouseEvent) {
		if (directo) agregar(ctx, producto, ui);
		else if (elige) eleccion.abrir(producto, e.currentTarget as HTMLElement);
	}
	const detalle = $derived(
		forma === 'lista' ? (producto.descripcion ?? '').split('\n')[0].slice(0, 90) : ''
	);
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
<article class="tv-prod tv-prod--{forma}" class:tv-prod--agotado={t.agotado}>
	<div class="tv-prod-media">
		<a class="tv-prod-foto" {href} tabindex="-1" aria-hidden="true">
			<Foto
				src={producto.foto}
				alt=""
				nombre={producto.nombre}
				{ancho}
				{prioridad}
				proporcion={forma === 'lista' ? '1/1' : 'var(--t-foto)'}
			/>
			{#if etiquetas.length && forma === 'grilla'}<span class="tv-prod-chips"
					>{#each etiquetas as e (e.texto)}<span
							class="tv-chip tv-prod-chip tv-chip--{e.tipo}"
							class:tv-chip--off={e.tipo === 'off'}>{e.texto}</span
						>{/each}</span
				>{/if}
			{#if sellos.length && forma === 'grilla'}<span class="tv-prod-sellos"
					>{#each sellos as x, i (x)}<span class="tv-prod-sello" class:tv-prod-sello--extra={i > 0}
							>{x}</span
						>{/each}</span
				>{/if}
		</a>
		{#if (directo || elige) && forma === 'grilla' && ctx.dir === 'vidriera'}<button
				type="button"
				class="tv-prod-rapido"
				aria-label="{directo ? 'Agregar' : 'Elegir opciones de'} {producto.nombre}"
				onclick={rapido}>{directo ? 'Agregar' : 'Elegir'}</button
			>{/if}
	</div>
	{#if ctx.conversion.favoritos && forma === 'grilla'}<Corazon {producto} />{/if}
	<div class="tv-prod-info">
		<a class="tv-prod-nombre" {href}>{producto.nombre}</a>
		{#if detalle}<span class="tv-prod-detalle">{detalle}</span>{/if}
		{#if sellos.length && forma === 'lista'}<span class="tv-chico tv-ok">{sellos.join(' · ')}</span
			>{/if}
		<div class="tv-prod-pie">
			<Precio {producto} linea={ctx.dir === 'pedido' || forma === 'lista'} />
			{#if directo && accion === 'agregar' && enCarrito}<Cantidad
					valor={enCarrito}
					max={producto.stock}
					nombre={producto.nombre}
					lleno
					alCambiar={(n) => ctx.carrito.fijar(clave, n)}
				/>{:else if directo && accion === 'agregar'}<button
					type="button"
					class="tv-agregar"
					aria-label="Agregar {producto.nombre} al carrito"
					onclick={() => agregar(ctx, producto, ui)}>Agregar</button
				>{:else if accion === 'agregar' && elige}<button
					type="button"
					class="tv-agregar"
					aria-haspopup="dialog"
					aria-label="Elegir opciones de {producto.nombre}"
					onclick={rapido}>Elegir</button
				>{:else if directo || elige}<button
					type="button"
					class="tv-mas"
					aria-haspopup={elige ? 'dialog' : undefined}
					aria-label={directo
						? `Agregar ${producto.nombre} al carrito`
						: `Elegir opciones de ${producto.nombre}`}
					onclick={rapido}
					><Icon icon={ctx.dir === 'vidriera' ? I.phPlus : I.phPlusBold} size={20} /></button
				>{/if}
		</div>
	</div>
</article>
