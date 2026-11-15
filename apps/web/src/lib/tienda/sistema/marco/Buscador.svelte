<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { usarTienda } from '../contexto.ts';
	import { productoHref } from '../datos.ts';
	import { cuantos, precio } from '../formato.ts';

	/**
	 * El buscador del encabezado (píldora en Mostrador, radio 10 en Pedido, recto en Vidriera): busca
	 * mientras se escribe entre los productos que ya tiene la página (sin tildes ni mayúsculas) y
	 * muestra hasta 6. Enviar lleva al catálogo.
	 */
	let { id = 'tv-buscar' }: { id?: string } = $props();
	const ctx = usarTienda();
	let q = $state('');
	let abierto = $state(false);
	const limpio = (s: string) => s.normalize('NFD').replace(/[̀-ͯ]/g, '').toLowerCase();
	const encontrados = $derived.by(() => {
		const b = limpio(q.trim());
		if (b.length < 2) return [];
		return ctx.tienda.productos
			.filter((p) => limpio(`${p.nombre} ${p.categoria ?? ''}`).includes(b))
			.slice(0, 6);
	});
	const n = $derived(ctx.tienda.productos.length);
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
<form
	class="tv-buscar"
	role="search"
	action="{ctx.base}/catalogo"
	method="get"
	onfocusin={() => (abierto = true)}
	onfocusout={(e) => {
		if (!e.currentTarget.contains(e.relatedTarget as Node)) abierto = false;
	}}
>
	<Icon icon={I.phMagnifyingGlass} size={18} />
	<label for={id} class="tv-oculto">Buscar productos</label>
	<input
		{id}
		name="q"
		type="search"
		autocomplete="off"
		maxlength="60"
		placeholder={n > 1 ? `Buscá entre ${cuantos(n, 'producto', 'productos')}` : 'Buscar'}
		bind:value={q}
	/>
	{#if abierto && q.trim().length >= 2}
		<div class="tv-buscar-lista" role="listbox" aria-label="Productos encontrados">
			{#each encontrados as p (p.id)}
				<a role="option" aria-selected="false" href={productoHref(ctx.tienda, p)}
					><span>{p.nombre}</span><b>{precio(p.precio)}</b></a
				>
			{:else}
				<p>No encontramos productos con «{q.trim()}».</p>
			{/each}
		</div>
	{/if}
</form>
