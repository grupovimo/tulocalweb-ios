<script lang="ts">
	import { campo, textoDe, type Bloque } from '../catalogo.ts';
	import { fotosUsadasDe, usarTienda } from '../contexto.ts';
	import { categoriasDe } from '../datos.ts';
	import Foto from '../primitivas/Foto.svelte';
	import Grilla from '../primitivas/Grilla.svelte';
	import Seccion from '../primitivas/Seccion.svelte';

	/**
	 * Categorías, con la foto elegida o la del primer producto (sin foto: la inicial de marca).
	 * - `tarjetas`: foto grande con el nombre abajo (A-moda) o tarjeta con foto al costado (B-moda).
	 * - `circulos`: fotos redondas en fila (en el celular se desliza).
	 * - `lista`: chips en fila (A-celular) o filas con la cantidad (C-moda).
	 */
	let { bloque }: { bloque: Bloque } = $props();
	const ctx = usarTienda();
	const cats = $derived(categoriasDe(ctx.tienda, campo(bloque, 'items'), fotosUsadasDe(bloque.id)));
	const titulo = $derived(textoDe(bloque, 'titulo'));
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
<Seccion
	{titulo}
	link={ctx.dir === 'vidriera' && bloque.variante === 'tarjetas'
		? { t: 'Ver todo', href: `${ctx.base}/catalogo` }
		: null}
	clase="tv-cats tv-cats--{bloque.variante}"
>
	{#if bloque.variante === 'circulos'}
		<ul class="tv-circulos" aria-label={titulo || 'Categorías'}>
			{#each cats as c (c.nombre)}
				<li>
					<a href={c.href}>
						<Foto src={c.foto} alt="" nombre={c.nombre} ancho={320} proporcion="1/1" />
						<span>{c.nombre}</span>
					</a>
				</li>
			{/each}
		</ul>
	{:else if bloque.variante === 'lista'}
		<ul class="tv-cats-lista" aria-label={titulo || 'Categorías'}>
			{#each cats as c (c.nombre)}
				<li>
					<a href={c.href}><span>{c.nombre}</span><span class="tv-suave">{c.cantidad}</span></a>
				</li>
			{/each}
		</ul>
	{:else}
		<Grilla
			items={cats}
			columnas={cats.length === 4 ? 4 : 3}
			movil={ctx.dir === 'mostrador' ? 1 : 2}
			etiqueta={titulo || 'Categorías'}
		>
			{#snippet children(c)}
				<a class="tv-cat" href={c.href}>
					<Foto src={c.foto} alt="" nombre={c.nombre} ancho={700} />
					<span class="tv-cat-textos">
						<span class="tv-cat-nombre">{c.nombre}</span>
						<span class="tv-cat-cant">{c.cantidad}</span>
						{#if ctx.dir === 'mostrador'}<span class="tv-cat-ver">Ver productos →</span>{/if}
					</span>
				</a>
			{/snippet}
		</Grilla>
	{/if}
</Seccion>
