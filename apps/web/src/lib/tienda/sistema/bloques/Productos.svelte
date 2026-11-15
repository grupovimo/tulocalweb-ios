<script lang="ts">
	import { campo, textoDe, type Bloque, type SeleccionProductos } from '../catalogo.ts';
	import { usarTienda } from '../contexto.ts';
	import { productosDe } from '../datos.ts';
	import { cuantos } from '../formato.ts';
	import Grilla from '../primitivas/Grilla.svelte';
	import Seccion from '../primitivas/Seccion.svelte';
	import TarjetaProducto from '../primitivas/TarjetaProducto.svelte';
	import UnProducto from './UnProducto.svelte';

	/**
	 * Productos: más vendidos, novedades, ofertas, de una categoría o elegidos a mano.
	 * - `grilla`: la grilla según cuántos hay (1 → destacado, 2 → mitades, 3 → tercios, 4+ → las
	 *   columnas elegidas, en filas completas).
	 * - `carrusel`: una fila que se desliza (scroll-snap), con las flechas del teclado.
	 * - `lista`: filas con foto chica, nombre, detalle y precio (para almacén y gastronomía).
	 */
	let { bloque }: { bloque: Bloque } = $props();
	const ctx = usarTienda();
	const sel = $derived(campo<SeleccionProductos>(bloque, 'productos'));
	const lista = $derived(productosDe(ctx.tienda, sel));
	const columnas = $derived(campo<number>(bloque, 'columnas') ?? 4);
	// «Lo más elegido» (el título por defecto) sólo si hay ventas que lo respalden: sin ninguna, los
	// «más vendidos» son el orden del catálogo y el título prometía algo que no era.
	const sinVentas = $derived(
		(sel?.fuente ?? 'vendidos') === 'vendidos' &&
			!(bloque.datos.titulo as string | undefined)?.trim() &&
			!(ctx.tienda.topVendidos ?? []).length
	);
	const titulo = $derived(sinVentas ? 'Nuestros productos' : textoDe(bloque, 'titulo'));
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
			<ul class="tv-carrusel" aria-label={titulo || 'Productos'}>
				{#each lista as p (p.id)}<li><TarjetaProducto producto={p} /></li>{/each}
			</ul>
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
