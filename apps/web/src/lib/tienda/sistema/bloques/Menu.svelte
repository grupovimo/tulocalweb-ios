<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { getUi } from '#lib/ui.svelte.ts';
	import { textoDe, type Bloque } from '../catalogo.ts';
	import { agregar, sumaDirecto } from '../acciones.ts';
	import { usarTienda } from '../contexto.ts';
	import { productoHref, seccionesMenu } from '../datos.ts';
	import Foto from '../primitivas/Foto.svelte';
	import Precio from '../primitivas/Precio.svelte';

	/**
	 * Menú para pedir (gastronomía): una sección por categoría con sus pestañas arriba, foto chica y
	 * «+» para agregar sin entrar a la ficha (C-cafe, C-celular, A-cafe). `lista`: filas; `columnas`:
	 * tarjetas de a dos.
	 */
	let { bloque }: { bloque: Bloque } = $props();
	const ctx = usarTienda();
	const ui = getUi();
	const secciones = $derived(seccionesMenu(ctx.tienda, bloque));
	const idDe = (n: string) =>
		`menu-${bloque.id}-${n
			.normalize('NFD')
			.replace(/[^\w]+/g, '-')
			.toLowerCase()}`;
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
<section class="tv-menu tv-ancho" data-variante={bloque.variante}>
	<div class="tv-cab"><h2 class="tv-h2">{textoDe(bloque, 'titulo')}</h2></div>
	{#if secciones.length > 1}
		<nav class="tv-menu-tabs" aria-label="Secciones del menú">
			{#each secciones as s (s.nombre)}<a href="#{idDe(s.nombre)}">{s.nombre}</a>{/each}
		</nav>
	{/if}
	{#each secciones as s (s.nombre)}
		<section class="tv-menu-sec" id={idDe(s.nombre)} aria-labelledby="{idDe(s.nombre)}-t">
			{#if secciones.length > 1}<h3 class="tv-h3" id="{idDe(s.nombre)}-t">{s.nombre}</h3>{/if}
			<ul class="tv-menu-items">
				{#each s.productos as p (p.id)}
					<li class="tv-menu-item tv-tarjeta">
						<div class="tv-menu-textos">
							<a class="tv-prod-nombre" href={productoHref(ctx.tienda, p)}>{p.nombre}</a>
							{#if p.descripcion}<span class="tv-prod-detalle"
									>{p.descripcion.split('\n')[0].slice(0, 110)}</span
								>{/if}
							<Precio producto={p} linea />
						</div>
						<div class="tv-menu-foto">
							<Foto src={p.foto} alt="" nombre={p.nombre} ancho={240} proporcion="1/1" />
							{#if sumaDirecto(ctx, p)}<button
									type="button"
									class="tv-mas"
									aria-label="Agregar {p.nombre}"
									onclick={() => agregar(ctx, p, ui)}><Icon icon={I.phPlusBold} size={18} /></button
								>{/if}
						</div>
					</li>
				{/each}
			</ul>
		</section>
	{/each}
</section>
