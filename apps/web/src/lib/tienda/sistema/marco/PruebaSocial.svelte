<script lang="ts">
	import { browser } from '$app/env';
	import { onMount } from 'svelte';
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { usarTienda } from '../contexto.ts';
	import { productoHref } from '../datos.ts';
	import Foto from '../primitivas/Foto.svelte';
	import { hace, recientes, type Reciente } from '../promos-api.ts';
	import { paginaDeVentana } from '../ventanas.ts';
	import { vistaVentanas } from '../vista-ventanas.svelte.ts';

	/**
	 * Compras recientes (conversion.pruebaSocial): un aviso chico abajo a la izquierda, «Ana de
	 * Córdoba compró Mate Imperial · hace 2 h», cada `cada` segundos y por 6 s. Sólo pedidos reales
	 * pagados de los últimos 7 días (GET /promos/recientes): sin compras, no sale nada; nunca se
	 * inventa un nombre ni una provincia. No en el checkout, los legales ni la vista previa; se corta
	 * mientras hay una ventana abierta, con el foco adentro o al cerrarlo.
	 */
	const ctx = usarTienda();
	const op = $derived(ctx.conversion.pruebaSocial);
	const va = $derived.by(() => {
		const p = paginaDeVentana(ctx.pagina);
		if (!p) return false;
		return !op.paginas.length || (p !== 'otra' && op.paginas.includes(p));
	});

	let lista = $state<Reciente[]>([]);
	let actual = $state<Reciente | null>(null);
	let i = 0;
	let cerrado = $state(false);
	let quieto = false;

	onMount(() => {
		if (!browser || vistaVentanas.vista) return;
		let vivo = true;
		void recientes(ctx.tienda.slug).then((l) => vivo && (lista = l.slice(0, 10)));
		return () => (vivo = false);
	});

	$effect(() => {
		if (!va || !lista.length || cerrado) return;
		const cada = Math.max(15, op.cada) * 1000;
		let ocultar: ReturnType<typeof setTimeout> | undefined;
		const mostrar = () => {
			if (quieto || document.querySelector('dialog.tv-vent[open], dialog.tv-panel[open]')) return;
			actual = lista[i++ % lista.length];
			clearTimeout(ocultar);
			ocultar = setTimeout(() => !quieto && (actual = null), 6000);
		};
		const primero = setTimeout(mostrar, 6000);
		const t = setInterval(mostrar, cada);
		return () => {
			clearTimeout(primero);
			clearInterval(t);
			clearTimeout(ocultar);
		};
	});
	const producto = $derived(
		actual?.productoId ? ctx.tienda.productos.find((p) => p.id === actual!.productoId) : undefined
	);
	const href = $derived(producto ? productoHref(ctx.tienda, producto) : null);
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
<div class="tv-social-zona" aria-live="polite">
	{#if actual}
		{@const c = actual}
		<div
			class="tv-social tv-tarjeta"
			role="status"
			onmouseenter={() => (quieto = true)}
			onmouseleave={() => (quieto = false)}
			onfocusin={() => (quieto = true)}
			onfocusout={() => (quieto = false)}
		>
			<span class="tv-social-foto"
				><Foto src={c.foto ?? producto?.foto} alt="" nombre={c.producto} ancho={120} /></span
			>
			<p>
				<b>{c.nombre}{c.provincia ? ` de ${c.provincia}` : ''}</b> compró
				{#if href}<a {href}>{c.producto}</a>{:else}{c.producto}{/if}
				<span class="tv-suave">· {hace(c.cuando)}</span>
			</p>
			<button
				type="button"
				class="tv-icono"
				aria-label="No mostrar más compras recientes"
				onclick={() => {
					cerrado = true;
					actual = null;
				}}><Icon icon={I.phXBold} size={14} /></button
			>
		</div>
	{/if}
</div>
