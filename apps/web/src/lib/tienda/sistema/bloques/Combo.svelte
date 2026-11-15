<script lang="ts">
	import { goto } from '$app/navigation';
	import { getUi } from '#lib/ui.svelte.ts';
	import { tarjetaDe } from '../../seleccion.svelte.ts';
	import { textoDe, type Bloque } from '../catalogo.ts';
	import { sumaDirecto } from '../acciones.ts';
	import { usarTienda } from '../contexto.ts';
	import { comboDe, productoHref } from '../datos.ts';
	import { precio } from '../formato.ts';
	import Boton from '../primitivas/Boton.svelte';
	import Foto from '../primitivas/Foto.svelte';

	/**
	 * Combo / completá el look: varios productos que se agregan juntos, con la tarjeta del precio
	 * (la suma, el descuento si hay y un solo botón). Si alguno tiene opciones, el botón lleva a
	 * elegir en su ficha. Con descuento, «Ahorrás»: agregados juntos, las líneas quedan marcadas con el
	 * combo y el checkout lo cobra mientras estén todos (combos.ts, igual que el servidor).
	 */
	let { bloque }: { bloque: Bloque } = $props();
	const ctx = usarTienda();
	const ui = getUi();
	const items = $derived(comboDe(ctx.tienda, bloque));
	const total = $derived(items.reduce((s, p) => s + tarjetaDe(p).precio, 0));
	const pct = $derived(Math.max(0, Math.min(50, Math.floor(Number(bloque.datos?.descuento ?? 0)))));
	const ahorro = $derived(Math.floor((total * pct + 50) / 100));
	const todosDirecto = $derived(items.every((p) => sumaDirecto(ctx, p)));
	function sumarTodo() {
		if (ctx.tienda.vitrina) return;
		ctx.carrito.sumarCombo(items, bloque.id);
		ui?.toast({
			k: 'ok',
			title: 'Sumaste el combo',
			action: 'Ver carrito',
			// eslint-disable-next-line svelte/no-navigation-without-resolve -- ruta de la tienda con su slug
			onAction: () => goto(`${ctx.base}/carrito`)
		});
	}
	const falta = $derived(items.find((p) => !sumaDirecto(ctx, p)));
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
<section class="tv-combo tv-ancho">
	<div class="tv-cab"><h2 class="tv-h2">{textoDe(bloque, 'titulo')}</h2></div>
	<div class="tv-combo-fila" style:--n={items.length}>
		<ul class="tv-combo-items">
			{#each items as p (p.id)}
				<li>
					<a href={productoHref(ctx.tienda, p)}>
						<Foto src={p.foto} alt="" nombre={p.nombre} ancho={500} proporcion="var(--t-foto)" />
						<span class="tv-prod-nombre">{p.nombre}</span>
						<span class="tv-suave">{precio(tarjetaDe(p).precio)}</span>
					</a>
				</li>
			{/each}
		</ul>
		<div class="tv-combo-precio tv-tarjeta">
			<p class="tv-chico tv-suave">Los {items.length} juntos</p>
			<p class="tv-combo-total">
				<b>{precio(total - ahorro)}</b>
			</p>
			{#if ahorro}<p class="tv-chico tv-suave">
					Ahorrás {precio(ahorro)} ({pct}&nbsp;%)
				</p>{/if}
			{#if todosDirecto}<Boton bloque onclick={sumarTodo}>Agregar los {items.length}</Boton
				>{:else if falta && !ctx.tienda.vitrina}<Boton bloque href={productoHref(ctx.tienda, falta)}
					>Elegir {falta.opciones[0]?.nombre.toLowerCase() ?? 'opción'}</Boton
				>{/if}
		</div>
	</div>
</section>
