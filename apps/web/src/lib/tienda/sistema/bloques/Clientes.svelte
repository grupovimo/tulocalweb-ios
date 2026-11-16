<script lang="ts">
	import { campo, textoDe, type Bloque, type BotonDatos } from '../catalogo.ts';
	import { usarTienda } from '../contexto.ts';
	import { clientesDe, enlace } from '../datos.ts';
	import { sinViuda } from '../formato.ts';
	import Boton from '../primitivas/Boton.svelte';
	import Carrusel from '../primitivas/Carrusel.svelte';
	import Foto from '../primitivas/Foto.svelte';
	import Seccion from '../primitivas/Seccion.svelte';

	/**
	 * Clientes: fotos de quienes compraron, con su nombre o usuario sobre un velo (contraste AA sobre
	 * cualquier foto) y, si tiene, el link a su publicación.
	 * - `grilla`: filas completas en la compu (4, 3, 5 o 6 por fila según cuántas hay; lo que no
	 *   llena una fila no se muestra); en el celular, una fila que se desliza.
	 * - `carrusel`: una fila que se desliza con flechas (primitivas/Carrusel).
	 * El botón (`cta`, «Subí tu foto y etiquetanos») va abajo, centrado.
	 */
	let { bloque }: { bloque: Bloque } = $props();
	const ctx = usarTienda();
	const items = $derived(clientesDe(ctx.tienda, bloque));
	const titulo = $derived(textoDe(bloque, 'titulo'));
	const texto = $derived(textoDe(bloque, 'texto'));
	const cta = $derived.by(() => {
		const b = campo<BotonDatos>(bloque, 'cta');
		const href = b?.t?.trim() ? enlace(ctx.tienda, b.a) : null;
		return href ? { t: b!.t.trim(), href } : null;
	});

	/** Las columnas que llenan filas: la primera que divide a n (si no, se recorta a filas de 4 o 3). */
	function armar(n: number) {
		for (const c of [4, 3, 5, 6]) if (n % c === 0) return { cols: c, n };
		if (n < 3) return { cols: n, n };
		const de4 = Math.floor(n / 4) * 4;
		const de3 = Math.floor(n / 3) * 3;
		return de4 >= de3 ? { cols: 4, n: de4 } : { cols: 3, n: de3 };
	}
	const armado = $derived(armar(items.length));
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- links que eligió la tienda -->
{#snippet tarjeta(c: (typeof items)[number])}
	<figure class="tv-cliente">
		{#if c.href}<a
				class="tv-cliente-link"
				href={c.href}
				target={c.href.startsWith('https://') ? '_blank' : undefined}
				rel={c.href.startsWith('https://') ? 'noopener' : undefined}
				aria-label={c.nombre ? `Ver la publicación de ${c.nombre}` : 'Ver la publicación'}
			></a>{/if}
		<Foto
			src={c.foto}
			alt={c.nombre ? `${c.nombre} con su compra` : ''}
			nombre={ctx.tienda.nombre}
			ancho={520}
			proporcion="4/5"
		/>
		{#if c.nombre}<figcaption>{c.nombre}</figcaption>{/if}
	</figure>
{/snippet}

<Seccion {titulo}>
	{#if texto}<p class="tv-lead tv-clientes-texto">{sinViuda(texto)}</p>{/if}
	{#if bloque.variante === 'carrusel'}
		<Carrusel
			{items}
			visibles={5}
			movil="62%"
			etiqueta={titulo || 'Clientes'}
			clase="tv-clientes-fila"
		>
			{#snippet children(c)}{@render tarjeta(c)}{/snippet}
		</Carrusel>
	{:else}
		<ul class="tv-clientes" style:--cols={armado.cols} aria-label={titulo || 'Clientes'}>
			{#each items.slice(0, armado.n) as c, i (i)}<li>{@render tarjeta(c)}</li>{/each}
		</ul>
	{/if}
	{#if cta}<div class="tv-clientes-cta">
			<Boton href={cta.href} tipo="secundario" externo={cta.href.startsWith('https://')}
				>{cta.t}</Boton
			>
		</div>{/if}
</Seccion>
