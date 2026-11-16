<script lang="ts">
	import { campo, textoDe, type Bloque, type SeleccionProductos } from '../catalogo.ts';
	import { usarTienda } from '../contexto.ts';
	import { productosDe } from '../datos.ts';
	import { MIN_PRODUCTOS_TOP } from '../conversion.ts';
	import { cuantos } from '../formato.ts';
	import Carrusel from '../primitivas/Carrusel.svelte';
	import Grilla from '../primitivas/Grilla.svelte';
	import Seccion from '../primitivas/Seccion.svelte';
	import TarjetaProducto from '../primitivas/TarjetaProducto.svelte';
	import UnProducto from './UnProducto.svelte';

	/**
	 * Productos: más vendidos, novedades, ofertas, de una categoría o elegidos a mano.
	 * - `grilla`: la grilla según cuántos hay (1 → destacado, 2 → mitades, 3 → tercios, 4+ → las
	 *   columnas elegidas, en filas completas).
	 * - `carrusel`: una fila que se desliza (scroll-snap), con flechas en la compu y tantas tarjetas a
	 *   la vista como columnas (primitivas/Carrusel).
	 * - `lista`: filas con foto chica, nombre, detalle y precio (para almacén y gastronomía).
	 */
	let { bloque }: { bloque: Bloque } = $props();
	const ctx = usarTienda();
	const sel = $derived(campo<SeleccionProductos>(bloque, 'productos'));
	const lista = $derived(productosDe(ctx.tienda, sel));
	const columnas = $derived(campo<number>(bloque, 'columnas') ?? 4);
	// «Lo más elegido» (el título por defecto) sólo si dice algo: con ventas que lo respalden y con
	// algo entre qué elegir. Sin ventas, los «más vendidos» son el orden del catálogo; con uno a tres
	// productos, «lo más elegido» es la tienda entera (y con uno solo, el antetítulo sobra).
	const porDefecto = $derived(
		(sel?.fuente ?? 'vendidos') === 'vendidos' &&
			!(bloque.datos.titulo as string | undefined)?.trim()
	);
	const noDistingue = $derived(
		porDefecto &&
			(!(ctx.tienda.topVendidos ?? []).length || ctx.tienda.productos.length < MIN_PRODUCTOS_TOP)
	);
	const titulo = $derived(
		noDistingue
			? ctx.tienda.productos.length === 1
				? ''
				: 'Nuestros productos'
			: textoDe(bloque, 'titulo')
	);
	const total = $derived(ctx.tienda.productos.length);
	const link = $derived(
		campo<boolean>(bloque, 'verTodo') !== false && total > lista.length
			? {
					t:
						ctx.dir === 'vidriera'
							? `Ver los ${cuantos(total, 'producto', 'productos')}`
							: 'Ver todo',
					href:
						sel?.fuente === 'categoria' && sel.categoria
							? `${ctx.base}/catalogo?${new URLSearchParams({ cat: sel.categoria })}`
							: `${ctx.base}/catalogo`
				}
			: null
	);
</script>

{#if lista.length === 1}
	<UnProducto producto={lista[0]} ante={titulo} />
{:else}
	<Seccion {titulo} ante={textoDe(bloque, 'etiqueta')} {link}>
		{#if bloque.variante === 'carrusel'}
			<Carrusel items={lista} visibles={columnas} etiqueta={titulo || 'Productos'}>
				{#snippet children(p)}<TarjetaProducto producto={p} ancho={500} />{/snippet}
			</Carrusel>
		{:else if bloque.variante === 'lista'}
			<ul class="tv-lista" aria-label={titulo || 'Productos'}>
				{#each lista as p (p.id)}<li>
						<TarjetaProducto producto={p} forma="lista" ancho={240} />
					</li>{/each}
			</ul>
		{:else}
			<Grilla items={lista} {columnas} etiqueta={titulo || 'Productos'}>
				{#snippet children(p)}<TarjetaProducto
						producto={p}
						ancho={lista.length <= 2 ? 1000 : 600}
					/>{/snippet}
			</Grilla>
		{/if}
	</Seccion>
{/if}
